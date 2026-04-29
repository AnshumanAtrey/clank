# 05 — Messaging-App Presence Detection

How OSINT tools determine whether a phone number is registered on each major messaging
platform. Goal: port the reliable methods into `clank` (Go CLI for phone OSINT).

Important upfront finding: **PhoneInfoga has no WhatsApp or Telegram scanner.** Its
`lib/remote/` directory contains only `local`, `numverify`, `googlesearch`, `googlecse`,
and `ovh` scanners. Messaging-app lookup is a *planned* feature that was never shipped.
Mr.Holmes' phone module (`Core/Searcher_phone.py`) is just Google/Yandex dorking and
country-specific JSON site-list lookups — no API calls into messengers either.
The actual presence-check capability lives in three categories: WhatsApp Web protocol
clients (whatsmeow / whatsapp-web.js / Baileys), Telegram MTProto clients (gotd/td,
Telethon, Pyrogram), and a small set of platform-specific tricks.

---

## 1. WhatsApp

WhatsApp has four investigative surfaces, in increasing order of usefulness.

### 1.1 `wa.me/{number}` redirect — least useful

`https://wa.me/15551234567` is the universal WhatsApp deep-link. Both registered and
unregistered numbers return HTTP 200 — the JS on the page decides which UI to render
client-side after WhatsApp's CDN resolves the number. With a static fetch you cannot
deterministically distinguish the two states. The page eventually shows either a
"Continue to Chat" button or "The phone number isn't on WhatsApp", but the discriminator
is rendered by `app.js`, not in the initial HTML payload. Headless-browser execution
plus DOM polling can read that text, but is fragile (Meta changes the strings) and slow
(~2-4 s per number). **Not recommended as a primary signal in 2026.**

### 1.2 `web.whatsapp.com/wa/checknumber` — historic endpoint

Killed years ago. Multi-device WhatsApp Web rebuilt around the Noise/Signal-protocol
binary socket; there is no plain HTTPS GET that returns registration status anymore.
**Dead since ~2019-2020.**

### 1.3 `v.whatsapp.net/v2/exist` — encrypted, not enumerable

This endpoint backs the official mobile clients. Body is encrypted with a key derived
from the registered device session, so without a paired device you cannot craft a
request. Mobile-Hacking-Lab confirmed in 2024-2025: *"the request body is not useful,
so easy enumeration via the API is not possible."* Used to leak `wa_old_device_name`
(e.g. `"Samsung Galaxy A32"`) from response on Android; Meta partially mitigated for
iOS in late 2024. **Not viable for an OSINT CLI.**

### 1.4 whatsmeow `IsOnWhatsApp` (the real method)

This is what every working WhatsApp-OSINT tool actually uses under the hood: a paired
WhatsApp Web session sends a USync IQ over the multi-device socket. In Go the canonical
implementation is `go.mau.fi/whatsmeow`. Source from `whatsmeow/user.go`:

```go
func (cli *Client) IsOnWhatsApp(ctx context.Context, phones []string) ([]types.IsOnWhatsAppResponse, error) {
    jids := make([]types.JID, len(phones))
    for i := range jids {
        jids[i] = types.NewJID(phones[i], types.LegacyUserServer)
    }
    list, err := cli.usync(ctx, jids, "query", "interactive", []waBinary.Node{
        {Tag: "business", Content: []waBinary.Node{{Tag: "verified_name"}}},
        {Tag: "contact"},
    })
    if err != nil { return nil, err }
    output := make([]types.IsOnWhatsAppResponse, 0, len(jids))
    querySuffix := "@" + types.LegacyUserServer
    for _, child := range list.GetChildren() {
        jid, jidOK := child.Attrs["jid"].(types.JID)
        if child.Tag != "user" || !jidOK { continue }
        var info types.IsOnWhatsAppResponse
        info.JID = jid
        info.VerifiedName, err = parseVerifiedName(child.GetChildByTag("business"))
        contactNode := child.GetChildByTag("contact")
        info.IsIn = contactNode.AttrGetter().String("type") == "in"
        contactQuery, _ := contactNode.Content.([]byte)
        info.Query = strings.TrimSuffix(string(contactQuery), querySuffix)
        output = append(output, info)
    }
    return output, nil
}
```

Response struct (from `whatsmeow/types/user.go`):

```go
type IsOnWhatsAppResponse struct {
    Query        string
    JID          JID            // the canonical WA user ID, e.g. 15551234567@s.whatsapp.net
    IsIn         bool           // registered on WhatsApp
    VerifiedName *VerifiedName  // populated for WhatsApp Business accounts
}
```

Companion calls give the rest of the profile:

```go
func (cli *Client) GetUserInfo(ctx context.Context, jids []types.JID) (map[types.JID]types.UserInfo, error)
func (cli *Client) GetProfilePictureInfo(ctx context.Context, jid types.JID, params *GetProfilePictureParams) (*types.ProfilePictureInfo, error)

type UserInfo struct {
    VerifiedName *VerifiedName
    Status       string   // the "About" text, if privacy allows
    PictureID    string
    Devices      []JID    // multi-device fanout — count of paired devices leaks
    LID          JID
}
type ProfilePictureInfo struct {
    URL        string `json:"url"`        // pps.whatsapp.net/v/t61.../...jpg?ccb=11-4&oh=...&oe=...
    ID         string `json:"id"`
    Type       string `json:"type"`
    DirectPath string `json:"direct_path"`
    Hash       []byte `json:"hash"`
}
```

Auth: pair the bot once via QR code (whatsmeow has `Client.GetQRChannel`), persist the
session in SQLite. Subsequent runs reconnect silently. Equivalent JS API in
`whatsapp-web.js` v1.34.7: `client.isRegisteredUser(id)`, `client.getNumberId(number)`,
`client.getProfilePicUrl(contactId)`, `client.getContactById(contactId)`. Note multiple
open issues (`pedroslopez/whatsapp-web.js#1812`, `#2252`) where `isRegisteredUser`
silently returns false — wwebjs reverse-engineers the Web client's internal store, and
breaks every few WhatsApp-Web releases. whatsmeow is more stable because it reimplements
the binary protocol directly.

**Reliability:** Works in 2026. **Auth needed:** Yes (paired WA session). **Returns:**
boolean registered, business verified-name, status text, profile-pic URL, device count.
**Legal/ToS:** Violates WhatsApp ToS §"Acceptable Use" if you query at scale; bot
accounts get banned. Keep query rate well under 50/min and randomize. One-off lookups
match normal client behavior.

---

## 2. Telegram

Telegram's MTProto API has two relevant methods: `contacts.importContacts` (bulk) and
`contacts.resolvePhone` (single). Both require a real user account with API ID + API
hash from `my.telegram.org/apps`.

### 2.1 `contacts.importContacts`

TL signature:

```
contacts.importContacts#2c800be5 contacts:Vector<InputContact> = contacts.ImportedContacts;
```

Returns a `contacts.ImportedContacts` with `imported`, `popular_invites`,
`retry_contacts`, and `users` vectors. Each `User` carries `id`, `first_name`,
`last_name`, `username`, `photo` (small + large file references), `phone`, plus
`access_hash` you need for downstream calls. Telegram explicitly notes:
*"according to the user's privacy settings, not all contacts which have an associated
Telegram account may be returned."* If the target has set Phone Privacy = Nobody +
"Who can find me by number" = My Contacts, they will not appear in the response — even
though their account exists.

### 2.2 `contacts.resolvePhone` (preferred for single-number lookup)

```
contacts.resolvePhone#8af94344 phone:string = contacts.ResolvedPeer;
```

Returns `peer:Peer chats:Vector<Chat> users:Vector<User>`. Errors:

- `PHONE_NOT_OCCUPIED` — number not registered (or hidden by privacy).
- `PHONE_NUMBER_INVALID` — bad format.
- `FLOOD_WAIT_X` — pause X seconds. Hit aggressively above ~20 lookups/minute on a
  fresh account; trusted accounts can sustain ~200/hour. Triggered earlier when
  resolving rather than importing.

### 2.3 Go code with gotd/td

```go
import (
    "context"
    "github.com/gotd/td/telegram"
    "github.com/gotd/td/tg"
)

func lookup(ctx context.Context, appID int, appHash, phone string) (*tg.User, error) {
    client := telegram.NewClient(appID, appHash, telegram.Options{})
    return client.Run(ctx, func(ctx context.Context) error {
        api := client.API()
        res, err := api.ContactsResolvePhone(ctx, phone) // phone in E.164, no '+'
        if err != nil {
            return err // tgerr.Is(err, "PHONE_NOT_OCCUPIED") -> not registered
        }
        for _, u := range res.Users {
            if user, ok := u.(*tg.User); ok {
                return inspect(user) // user.ID, FirstName, LastName, Username, Photo
            }
        }
        return nil
    })
}
```

For Python the equivalent is Telethon `client(functions.contacts.ResolvePhoneRequest(phone))`
or Pyrogram `app.invoke(ResolvePhone(phone=phone))`.

Edge case: even a "successful" resolve can give `peer = peerUser` with a user object
that has no photo and a placeholder name if the target hides everything. Workaround
mentioned in MadelineProto issue #1317 — adding the contact via `contacts.addContact`
first sometimes upgrades visibility — but that creates a real contact link on the
target's side and is poor OPSEC for OSINT.

**Reliability:** Works in 2026, very widely used. **Auth needed:** Yes (Telegram user
session, not bot). **Returns:** user_id, first/last name, username, profile photo file
ref, premium flag, possibly bio (extra `users.getFullUser` call). **Legal/ToS:**
Telegram ToS forbids automated mass scraping; per-account rate is the practical limit.

---

## 3. Signal

**No public registration-check exists, by design.** Signal's contact discovery is
implemented via CDSI (Contact Discovery Service — Icelake) running inside Intel SGX
enclaves. Clients perform a private-set-intersection-style protocol: phone numbers are
hashed and queried inside an attested enclave, so even Signal's own server operators
cannot see which numbers a client looked up. Researchers reproducing the protocol
report that after roughly 4,000 queries the server returns HTTP 413 and rate-limits the
account aggressively. The system is the canonical example of "privacy by design"
preventing OSINT.

The only OSINT signal you can extract from a number on Signal is the side-channel
"this number does/doesn't get a registration-lock prompt when you try to register" —
which requires actually attempting account creation with that number, is destructive
(it locks them out for 7 days), and is unambiguously abuse. **Do not implement.**

**Reliability:** N/A — not feasible. **Auth needed:** N/A. **Returns:** Nothing.
**Legal/ToS:** Any workaround violates Signal ToS and may be CFAA-actionable in the US.

---

## 4. Viber, LINE, KakaoTalk, WeChat

**Viber.** No clean HTTP endpoint. The `viber.click/{number}` deep-link returns the
same template for registered/unregistered. The `sgxgsx/ViberOSINT` repo confirms the
practical method is to import the number as a contact in Viber Desktop and watch
whether the contact name auto-resolves to a profile name and avatar — a manual UI
inspection, not an API. Possible to script by automating Viber Desktop, but brittle
and slow. **Auth needed:** Yes (paired Viber Desktop). **Returns:** display name and
avatar if registered. **2026:** still works manually, no public API.

**LINE.** No documented OSINT-friendly endpoint. LINE's `line.me/R/ti/p/...` deep-link
takes a LINE ID, not a phone number. The phone-to-LINE-ID step is internal to the
mobile app's contact-sync (which uploads the entire phonebook to LINE servers — so it
exists, just not as a public method). Third-party tools all require a real LINE
account, automation of the mobile client, or unofficial API wrappers like `linepy`.
**Reliability:** poor. **2026:** account-bans common, no clean answer.

**KakaoTalk.** Same shape as LINE — phone→KakaoID resolution happens inside the
contact-sync flow only, no public API. The "Friend Suggestions" feature reveals
matches when you upload a contact, so a paired Kakao account can *batch* check
phone numbers, but doing so at any scale will get the account flagged. **Reliability:**
poor without a Korean phone for verification. **2026:** Kakao ID search by username
is the practical OSINT path; phone-to-account is essentially closed.

**WeChat.** WeChat technically supports adding contacts by phone number from inside
the app (Add Friend → Phone), but: (a) a real WeChat account is required, (b) Chinese
WeChat accounts increasingly require verification by an existing user, and (c) WeChat's
risk-control system flags accounts that bulk-search phone numbers within hours.
Limited to single-target interactive lookups. **Reliability:** functionally poor for a
CLI; great for hand-investigation if you already have a WeChat account.

---

## 5. Discord and iMessage

**Discord.** Phone numbers on Discord are not a public lookup surface — they are
hashed before sending to the server (so Discord itself can match without storing
plaintext). There is no `/api/v9/users/by-phone` endpoint. The only OSINT angle is
secondary: leaked-DB lookup (Discord's 2023 phone-number leak via support tickets),
or username/Discriminator → cross-platform footprint, neither of which is "phone
in → registration boolean out." **Skip.**

**iMessage.** Apple's `query.ess.apple.com` Identity Services endpoint accepts
`com.apple.madrid` queries that return whether a handle (email or `tel:+15551234567`)
is iMessage-capable. This is what blue/green-bubble detection in macOS uses. To call
it directly you need: (a) a valid Apple-ID push token, (b) an Activation Record
generated by a real macOS or iOS device. Beeper's open-source `phone-registration-provider`
showed how to script this via a jailbroken iPhone, and Apple aggressively shut down
Beeper Mini in late 2023 by invalidating their tokens; the `beeper/imessage` Matrix
bridge repo was archived April 2025. A motivated investigator with a Mac running 24/7
can run the queries from a real device, but a packaged CLI cannot ship this without a
device dependency. **Reliability:** technically possible but operationally painful;
**not recommended for clank v1.**

---

## Reliability Matrix

| App        | Method                                        | Auth needed?              | Works 2026? | Returns                                                |
|------------|-----------------------------------------------|---------------------------|-------------|--------------------------------------------------------|
| WhatsApp   | whatsmeow `IsOnWhatsApp` (USync over WA-Web)  | Yes — paired QR session   | Yes         | bool, JID, biz verified-name, about-text, pic URL, devices |
| WhatsApp   | `wa.me/{n}` HTML scrape (headless)            | No                        | Marginal    | bool only, brittle, slow                               |
| WhatsApp   | `v.whatsapp.net/v2/exist`                     | Encrypted session         | Not viable  | —                                                      |
| WhatsApp   | `web.whatsapp.com/wa/checknumber`             | —                         | Dead ~2020  | —                                                      |
| Telegram   | `contacts.resolvePhone` via gotd/td           | Yes — user API ID/hash    | Yes         | user_id, first/last/username, photo ref, premium       |
| Telegram   | `contacts.importContacts` (bulk)              | Yes — user API ID/hash    | Yes         | same, batched; subject to FLOOD_WAIT                   |
| Signal     | —                                             | —                         | No (CDSI)   | nothing                                                |
| Viber      | Viber Desktop contact-import UI scrape        | Yes — paired account      | Manual only | display name, avatar                                   |
| LINE       | None public                                   | —                         | No          | —                                                      |
| KakaoTalk  | Friend-Sync via paired KR account             | Yes                       | Risky       | name, KakaoID                                          |
| WeChat     | In-app Add-by-phone                           | Yes — paired account      | Manual only | nickname, avatar                                       |
| Discord    | None                                          | —                         | No          | — (hashed server-side)                                 |
| iMessage   | `query.ess.apple.com` (Madrid IDS)            | Yes — real Apple device   | Marginal    | bool capable                                           |

---

## Sources

- [PhoneInfoga scanners doc](https://sundowndev.github.io/phoneinfoga/getting-started/scanners/) — confirms no WhatsApp/Telegram scanner.
- [PhoneInfoga `lib/remote/` source](https://github.com/sundowndev/phoneinfoga/tree/master/lib/remote)
- [Mr.Holmes phone module](https://github.com/Lucksi/Mr.Holmes/blob/master/Core/Searcher_phone.py) — only Google/Yandex dorks.
- [whatsmeow Go library](https://github.com/tulir/whatsmeow) — `IsOnWhatsApp` source.
- [whatsmeow types/user.go](https://pkg.go.dev/go.mau.fi/whatsmeow/types) — IsOnWhatsAppResponse, UserInfo, ProfilePictureInfo.
- [whatsapp-web.js Client docs v1.34.7](https://docs.wwebjs.dev/Client.html) — isRegisteredUser, getNumberId, getProfilePicUrl.
- [Mobile Hacking Lab — Silent OSINT WhatsApp](https://www.mobilehackinglab.com/blog/silent-osint-whatsapp) — `/v2/exist` endpoint analysis.
- [kinghacker0 WhatsApp-OSINT (RapidAPI wrapper)](https://github.com/kinghacker0/WhatsApp-OSINT)
- [gotd/td Telegram Go client](https://github.com/gotd/td)
- [contacts.importContacts spec](https://core.telegram.org/method/contacts.importContacts)
- [contacts.resolvePhone spec](https://core.telegram.org/method/contacts.resolvePhone)
- [Telegram contacts API privacy note](https://core.telegram.org/api/contacts)
- [MadelineProto issue #1317 — PHONE_NOT_OCCUPIED workaround](https://github.com/danog/MadelineProto/issues/1317)
- [Signal CDSI Icelake](https://github.com/signalapp/ContactDiscoveryService-Icelake)
- [Signal — Private contact discovery](https://signal.org/blog/private-contact-discovery/)
- [sgxgsx/ViberOSINT](https://github.com/sgxgsx/ViberOSINT)
- [Beeper iMessage bridge (archived)](https://github.com/beeper/imessage)
- [Beeper phone-registration-provider](https://github.com/beeper/phone-registration-provider)
