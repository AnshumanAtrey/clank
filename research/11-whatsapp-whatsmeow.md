# 11 — WhatsApp via `whatsmeow` (deep dive for clank integration)

Research dive on `go.mau.fi/whatsmeow` (mirror: `github.com/tulir/whatsmeow`) ahead of building `clank whatsapp <login | lookup | logout>`. All API signatures and types quoted below were fetched directly from raw.githubusercontent.com on 2026-04-29.

## 1. Health audit

- **Repo URL (canonical):** `https://github.com/tulir/whatsmeow`. Module path is `go.mau.fi/whatsmeow` (Tulir's vanity import, served from his own server). They are the *same code* — `go.mau.fi/whatsmeow` redirects to the github.com/tulir/whatsmeow git repo.
- **Stars / forks / last commit / license** (from GitHub API on 2026-04-29):
  - Stars: **5,972**
  - Forks: **955**
  - Open issues: **62**
  - Last commit: **2026-04-27** (`.github: link to general contributing guidelines`, sha `7514259253`)
  - License: **MPL-2.0**
  - `archived: false`, `disabled: false`
- **Maintainer:** [Tulir Asokan](https://github.com/tulir) (`tulir@maunium.net`). Solo maintainer, but funded indirectly via [Beeper](https://www.beeper.com/) (he's a Beeper engineer; whatsmeow powers their WhatsApp bridge).
- **No deprecation / archive / "moving to X" notice.** README is alive; recent commits are bug fixes and protocol refresh PRs.
- **Used by:** `pkg.go.dev` lists **183 importers** for `go.mau.fi/whatsmeow/store/sqlstore` and 228 packages overall. Top production consumers:
  - `go.mau.fi/mautrix-whatsapp` — flagship Matrix↔WhatsApp puppeting bridge (Tulir's own)
  - `go.mau.fi/mautrix-meta` — Matrix↔Messenger/Instagram bridge
  - `element-hq/mautrix-whatsapp` — Element/Vector's hard fork
  - `EvolutionAPI/evolution-go` — commercial WhatsApp REST API
  - `aldinokemal/go-whatsapp-web-multidevice` — popular multi-session REST API
  - `slidge-whatsapp` — XMPP gateway
  - `lharries/whatsapp-mcp` — MCP server for WhatsApp
  - `vicentereig/whatsapp-cli`, `steipete/wacli` — CLI tools (closest analogues to clank)
  - `Okramjimmy/whatsapp_server`, `wsapi-chat/wsapi-app` — custom servers
- **2026 risk rating: GREEN.** Daily-active maintenance by a paid engineer at a real company shipping it in production for thousands of Beeper users. Single-maintainer = bus factor risk, but the codebase is heavily exercised and the API surface for *our* use case (login + IsOnWhatsApp + UserInfo) is rock solid and unchanged for years.

## 2. Auth flow — QR-pair from scratch

**Important:** the `mdtest/main.go` example file *no longer exists*. Tulir deleted it 2024-07-16 with the commit message:

> "mdtest: delete — People keep misusing it for things it was never meant to do, so it's easier to delete the whole thing."

The canonical replacement is the [godoc package example](https://pkg.go.dev/go.mau.fi/whatsmeow#example-package). Verbatim from `pkg.go.dev` (current package version `v0.0.0-20260427122815-7514259253a7`):

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect()
}
```

### Step-by-step

1. **Open container.** From `store/sqlstore/container.go`:
   ```go
   func New(ctx context.Context, dialect, address string, log waLog.Logger) (*Container, error)
   ```
   Pass `"sqlite3"` (or `"sqlite"` for modernc — both work, see §4) and a DSN. Foreign keys *must* be enabled — `Container.Upgrade()` panics with `"foreign keys are not enabled"` otherwise.
2. **Get / create device.** From the same file:
   ```go
   func (c *Container) GetFirstDevice(ctx context.Context) (*store.Device, error)
   ```
   > "GetFirstDevice is a convenience method for getting the first device in the store. If there are no devices, then a new device will be created."

   The returned `*store.Device` has `ID == nil` if it's brand new (no JID yet).
3. **Construct client.** From `client.go`:
   ```go
   func NewClient(deviceStore *store.Device, log waLog.Logger) *Client
   ```
4. **Detect "no session"** with `client.Store.ID == nil` — same check `GetQRChannel` performs internally:
   ```go
   } else if cli.Store.ID != nil {
       return nil, ErrQRStoreContainsID
   }
   ```
5. **Get QR channel BEFORE Connect.** From `qrchan.go`:
   ```go
   func (cli *Client) GetQRChannel(ctx context.Context) (<-chan QRChannelItem, error)
   ```
   Returns `ErrQRAlreadyConnected` if you call after Connect.
6. **Connect.** From `client.go`:
   ```go
   func (cli *Client) Connect() error
   ```
7. **Loop on `QRChannelItem`.** From `qrchan.go`:
   ```go
   type QRChannelItem struct {
       Event   string        // "code", "error", "success", "timeout", ...
       Error   error         // populated when Event == "error"
       Code    string        // populated when Event == "code"
       Timeout time.Duration // before next code rotates
   }
   ```
   Constants and sentinel events:
   ```go
   const QRChannelEventCode = "code"
   const QRChannelEventError = "error"

   var (
       QRChannelSuccess                   = QRChannelItem{Event: "success"}
       QRChannelTimeout                   = QRChannelItem{Event: "timeout"}
       QRChannelErrUnexpectedEvent        = QRChannelItem{Event: "err-unexpected-state"}
       QRChannelClientOutdated            = QRChannelItem{Event: "err-client-outdated"}
       QRChannelScannedWithoutMultidevice = QRChannelItem{Event: "err-scanned-without-multidevice"}
   )
   ```
   Event meanings:
   - `code` — render `evt.Code` as a QR. First code is shown for **60s**, all subsequent for **20s** (from `qrchan.go`: `if len(codes) == 6 { timeout = 60 * time.Second }`).
   - `success` — pairing OK; the channel is then closed. Session is persisted via `Device.Save()`.
   - `timeout` — server closed the websocket before scan / ran out of codes.
   - `err-client-outdated` — WhatsApp says the client version is too old. You need to bump `store.SetWAVersion`.
   - `err-scanned-without-multidevice` — phone scanned the QR but doesn't have multi-device enabled. User can re-enable and re-scan (same code is still valid).
   - `err-unexpected-state` — `events.Connected` / `events.ConnectFailure` / `events.LoggedOut` / `events.TemporaryBan` arrived during pairing. Generally means pairing already happened.
   - `error` — `evt.Error` populated; pairing failed locally after server sent pair-success.
8. **After success, persist.** Pairing automatically calls `Device.Save()` which writes the row to `whatsmeow_device`. Subsequent runs find `Store.ID != nil` and skip QR.

### Imports recap

```go
import (
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
)
```

## 3. Lookup APIs

### `IsOnWhatsApp` (from `user.go`)

```go
// IsOnWhatsApp checks if the given phone numbers are registered on WhatsApp.
// The phone numbers should be in international format, including the `+` prefix.
func (cli *Client) IsOnWhatsApp(ctx context.Context, phones []string) ([]types.IsOnWhatsAppResponse, error)
```

It builds a USync IQ over `c.us` (legacy server). Note the line:
```go
jids[i] = types.NewJID(phones[i], types.LegacyUserServer)
```
where `LegacyUserServer = "c.us"`. The server then responds with the canonical JID on `s.whatsapp.net`.

Response struct (`types/user.go`):
```go
type IsOnWhatsAppResponse struct {
    Query string // The query string used
    JID   JID    // The canonical user ID
    IsIn  bool   // Whether the phone is registered or not.
    VerifiedName *VerifiedName // If the phone is a business, the verified business details.
}
```

- **`Query`** — the phone you submitted, with `@c.us` stripped (e.g. `"+14155552671"`).
- **`JID`** — canonical WA JID, almost always `<digits>@s.whatsapp.net` (e.g. `14155552671@s.whatsapp.net`). The newer `@lid` ("hidden user") JIDs are returned by `GetUserInfo` in the `LID` field, not here.
- **`IsIn`** — `true` if `<contact type="in">` came back from WA. **Critical gotcha:** if WhatsApp considers the number outright invalid, **no `<user>` node is returned**, so the response slice will simply be missing that entry (issue [#1086](https://github.com/tulir/whatsmeow/issues/1086), closed *not planned* 2026-02-21). Always reconcile input → output by `Query`. Do not assume `len(out) == len(phones)`.
- **`VerifiedName`** — non-nil only for WhatsApp Business accounts with verified names (green-check). Confirmed.
- **Privacy:** `IsIn == true` doesn't mean the user is reachable — they may have you blocked or set "Who can message me?" to contacts-only. The IQ doesn't expose those settings.

### `GetUserInfo` (from `user.go`)

```go
// GetUserInfo gets basic user info (avatar, status, verified business name, device list).
func (cli *Client) GetUserInfo(ctx context.Context, jids []types.JID) (map[types.JID]types.UserInfo, error)
```

Response struct (`types/user.go`):
```go
type UserInfo struct {
    VerifiedName *VerifiedName
    Status       string  // "About" text. Empty if user has restricted who sees it.
    PictureID    string
    Devices      []JID
    LID          JID
}
```

- **`Status`** — the user's "About" text. Privacy-gated server-side: if the user restricts who sees their About to "Contacts" or "Nobody" and you aren't a contact, this comes back as empty string — *not* an error.
- **`PictureID`** — opaque ID; pass to `GetProfilePictureInfo` to fetch the URL. Use `params.ExistingID = info.PictureID` to skip download if unchanged.
- **`Devices`** — list of AD-JIDs (`<user>:<agent>:<device>@s.whatsapp.net`) for every device the account has linked (phone + companions). Useful as a "this user has N devices" signal.
- **`LID`** — the user's privacy-preserving "hidden user" JID on the `lid` server (`<digits>@lid`). WhatsApp is gradually migrating from phone-number JIDs to LIDs; for now most APIs accept either.
- **`VerifiedName`** — populated only for verified businesses, same as in `IsOnWhatsApp`.

### `GetProfilePictureInfo` (from `user.go`)

```go
type GetProfilePictureParams struct {
    Preview     bool
    ExistingID  string
    IsCommunity bool
    CommonGID   types.JID
    InviteCode  string
    PersonaID   string
}

// GetProfilePictureInfo gets the URL where you can download a WhatsApp user's profile picture or group's photo.
//
// Optionally, you can pass the last known profile picture ID.
// If the profile picture hasn't changed, this will return nil with no error.
//
// To get a community photo, you should pass `IsCommunity: true`, as otherwise you may get a 401 error.
func (cli *Client) GetProfilePictureInfo(ctx context.Context, jid types.JID, params *GetProfilePictureParams) (*types.ProfilePictureInfo, error)
```

Response (`types/user.go`):
```go
type ProfilePictureInfo struct {
    URL        string `json:"url"`         // Download via plain HTTP GET
    ID         string `json:"id"`          // Same as UserInfo.PictureID
    Type       string `json:"type"`        // "image" (full) or "preview" (thumb)
    DirectPath string `json:"direct_path"` // Internal path; URL is what you want
    Hash       []byte `json:"hash"`
}
```

`Preview: true` returns a low-res thumbnail. Returns `(nil, nil)` if the picture hasn't changed (when `ExistingID` is given), and a `*ElementMissingError` if the user has profile-photo privacy locked down.

### Phone-number format

The function comment says `"international format, including the +"`. In practice the `+` is stripped by `IsOnWhatsApp` before building the JID; passing or omitting it both work, but the docstring is the source of truth — pass with `+` (`"+14155552671"`).

### Rate limiting

Whatsmeow itself does not rate-limit `IsOnWhatsApp`. WhatsApp server-side does — a December 2025 [Equixly disclosure](https://equixly.com/blog/2025/12/14/whats-app-api-vulnerability/) demonstrated **7,000 lookups/sec** on a single session before WA patched. As of 2026, expect IQ throttling (returns `<error code="429">`) if you blast >50/min. Whatsmeow does not surface a typed error for this; it returns a generic IQ error from `sendIQ`.

In discussion [#199](https://github.com/tulir/whatsmeow/discussions/199) and [#567](https://github.com/tulir/whatsmeow/discussions/567), maintainers and community converge on:

> "got recently 2 disconnects stating that those numbers use a unofficial WA Version."

OSINT-style usage *can* trigger `events.TemporaryBan` (`code: 401` → 403/406 connect failures) within hours if the account isn't warmed up. Ban thresholds are not published.

## 4. Storage / persistence

- **Package import:** `go.mau.fi/whatsmeow/store/sqlstore`
- **Container constructor** (`store/sqlstore/container.go`):
  ```go
  func New(ctx context.Context, dialect, address string, log waLog.Logger) (*Container, error)
  ```
  > "Only SQLite and Postgres are currently fully supported."
- **Driver dialect parsing** lives in `go.mau.fi/util/dbutil`:
  ```go
  func ParseDialect(engine string) (Dialect, error) {
      engine = strings.ToLower(engine)
      if strings.HasPrefix(engine, "postgres") || engine == "pgx" {
          return Postgres, nil
      } else if strings.HasPrefix(engine, "sqlite") || strings.HasPrefix(engine, "litestream") {
          return SQLite, nil
      }
      ...
  }
  ```
  So **any string starting with `sqlite`** is treated as the SQLite dialect — meaning both `"sqlite3"` (mattn) and `"sqlite"` (modernc) work as the first arg to `sqlstore.New`. The string is passed straight to `sql.Open`, so it must match a registered driver name.
- **mattn/go-sqlite3** (CGo): registers driver `"sqlite3"`. DSN: `"file:path/to/file.db?_foreign_keys=on"` (mattn-style query string).
- **modernc.org/sqlite** (pure Go): registers driver `"sqlite"`. DSN: `"file:path/to/file.db?_pragma=foreign_keys(1)"` (URI-style with `_pragma=` repeated for each pragma). Useful: `?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(wal)`.
- The godoc canonical example uses `"sqlite3"` (mattn). The example explicitly notes: `"NOTE: You must also import the appropriate DB connector, e.g. github.com/mattn/go-sqlite3 for SQLite"`.
- mautrix-whatsapp's own go.mod ships **`github.com/mattn/go-sqlite3 v1.14.42`** as an indirect dep; it does not import modernc.

### Container API

```go
// container/container.go
func (c *Container) GetFirstDevice(ctx context.Context) (*store.Device, error)
func (c *Container) GetAllDevices(ctx context.Context) ([]*store.Device, error)
func (c *Container) GetDevice(ctx context.Context, jid types.JID) (*store.Device, error)
func (c *Container) NewDevice() *store.Device
func (c *Container) DeleteDevice(ctx context.Context, store *store.Device) error
func (c *Container) Close() error
```

Use `GetFirstDevice` for single-session apps (clank). `NewDevice()` returns an unsaved device with fresh keys; pairing automatically saves it.

### File location & permissions

clank should use `~/.clank/whatsapp.db`. Recommended: `0700` on the directory, `0600` on the db file (plus `-wal` and `-shm` shadow files). The DB stores the device's noise/identity/signed-pre-key private keys and is effectively a session token — anyone who reads it can impersonate the WhatsApp account. WAL mode means three files; back up all three or quiesce first.

## 5. Connection lifecycle / pitfalls

- **`Connect()`** (from `client.go`):
  ```go
  func (cli *Client) Connect() error {
      return cli.ConnectContext(cli.BackgroundEventCtx)
  }
  ```
  It opens the websocket synchronously, sends the noise handshake, and returns once the socket transition completes. **Auth is not done yet** when `Connect()` returns — wait for `events.Connected` (or call `cli.WaitForConnection(timeout)`).
- **Auto-reconnect:** `EnableAutoReconnect: true` by default in `NewClient`. From `client.go`:
  ```go
  EnableAutoReconnect: true,
  ```
  After `events.Disconnected`, the client retries with exponential-ish backoff (`AutoReconnectErrors * 2 * time.Second`) — *unless* it received a permanent-disconnect event (`LoggedOut`, `StreamReplaced`, `ClientOutdated`, `TemporaryBan`, `CATRefreshError`). All of those implement `events.PermanentDisconnect`.
- **14-day phone limit** is enforced *server-side by WhatsApp*, not by the library. When the primary phone has been offline >14 days, the server unlinks all companions; whatsmeow surfaces this as `events.LoggedOut{OnConnect: true, Reason: 401}` (`ConnectFailureLoggedOut`) on the next reconnect. From `connectionevents.go`:
  ```go
  if reason.IsLoggedOut() {
      cli.Log.Infof("Got %s connect failure, sending LoggedOut event and deleting session", reason)
      go cli.dispatchEvent(&events.LoggedOut{OnConnect: true, Reason: reason})
      err := cli.Store.Delete(ctx)
      ...
  }
  ```
  Whatsmeow auto-deletes the device row from SQLite on this event, so the next clank invocation will re-show the QR.
- **`Disconnect()`** is idempotent and silent:
  > "This will not emit any events, the Disconnected event is only used when the connection is closed by the server or a network error."
- **`Logout(ctx)`** sends a `remove-companion-device` IQ, then calls `Disconnect()` *and* `Store.Delete(ctx)`. Source:
  ```go
  func (cli *Client) Logout(ctx context.Context) error {
      ...
      _, err := cli.sendIQ(ctx, infoQuery{
          Namespace: "md", Type: "set", To: types.ServerJID,
          Content: []waBinary.Node{{Tag: "remove-companion-device", ...}},
      })
      if err != nil { return fmt.Errorf("error sending logout request: %w", err) }
      cli.Disconnect()
      err = cli.Store.Delete(ctx)
      ...
  }
  ```
  So the device row is deleted but the **SQLite file remains**. clank can either leave it (it'll be re-populated next pairing) or `os.Remove` it for a clean reset.
- **Goroutine safety:** the `Client` struct uses `sync.RWMutex` (`socketLock`, `eventHandlersLock`, `userDevicesCacheLock`, etc) and `sync.Map`-style locking around every cache. **Multiple `IsOnWhatsApp` / `GetUserInfo` calls can be issued concurrently** from different goroutines — each is a synchronous IQ request/response with its own response-waiter slot.

## 6. Recent (2024–2026) breakages

- **#984** "Error with new WhatsApp Version" — periodic protocol bumps require updating `store.GetWAVersion()`. Library is patched within hours/days each time.
- **#1086** (Feb 2026, *not planned*) — `IsOnWhatsApp` silently drops invalid numbers. Code defensively (see §10).
- **#1085** (Feb 2026) — "Data race in FrameSocket" still open as of 2026-04. Low risk for our use case (single goroutine driving the lookup) but worth knowing.
- **#1117** (Apr 2026) — "scanning a devices takes 30-40s". Scaling concern, not a correctness issue.
- **#1074** (Feb 2026) — error 463 sending to specific contacts. Send-side, irrelevant for OSINT.
- **#818** (closed not-planned, 2025) — connections dropping in <20min on heavy-traffic accounts. Auto-reconnect handles it transparently.
- **#14** (closed dup, Jul 2025) — long-running "Keep getting banned" thread. Tulir's standing position: bans are a WhatsApp-server policy issue, not a library bug.
- **No major protocol-version bumps** in the core USync/IsOnWhatsApp surface over the last 12 months. The `usync` IQ shape used by `IsOnWhatsApp` and `GetUserInfo` has been stable since ~2022.
- **No deprecation notices** on these APIs.

The **biggest 2024 breaking change** was the deletion of `mdtest/` (commit 2024-07-16) — purely a docs/example change, no API impact.

## 7. Logging

`waLog` interface (`util/log/log.go`):

```go
type Logger interface {
    Warnf(msg string, args ...interface{})
    Errorf(msg string, args ...interface{})
    Infof(msg string, args ...interface{})
    Debugf(msg string, args ...interface{})
    Sub(module string) Logger
}

var Noop Logger = &noopLogger{}

func Stdout(module string, minLevel string, color bool) Logger
```

`waLog.Stdout("X", "ERROR", false)` only emits ERROR-level lines, no color. For clank's `lookup` subcommand we want **silent unless something breaks** — pass `waLog.Noop` to both `sqlstore.New` and `whatsmeow.NewClient` (or use `Stdout("clank-wa", "ERROR", false)` if we want surfaced errors). `Stdout` writes to `os.Stdout` (not stderr), so `Noop` is safest if we're piping JSON to stdout.

## 8. Minimum viable integration — full Go code

Drop-in spec for `clank/internal/whatsapp/whatsapp.go`. Compiles against current `go.mau.fi/whatsmeow` (Apr 2026).

```go
// Package whatsapp provides a thin wrapper around go.mau.fi/whatsmeow for clank.
package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/mdp/qrterminal/v3"
	_ "github.com/mattn/go-sqlite3" // registers "sqlite3" driver — see §9 for CGo decision

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// dbPath returns ~/.clank/whatsapp.db, creating the parent dir at 0700.
func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".clank")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "whatsapp.db"), nil
}

// dsn builds a mattn/go-sqlite3 DSN with foreign keys + sane busy timeout.
func dsn(path string) string {
	q := url.Values{}
	q.Set("_foreign_keys", "on")
	q.Set("_busy_timeout", "5000")
	q.Set("_journal_mode", "WAL")
	return "file:" + path + "?" + q.Encode()
}

// Open returns a connected, authenticated whatsmeow client. Runs QR pairing if needed.
// The caller MUST eventually call Close on the returned handle.
func Open(ctx context.Context) (*Handle, error) {
	path, err := dbPath()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	logger := waLog.Noop // ERROR-only: waLog.Stdout("clank-wa", "ERROR", false)
	container, err := sqlstore.New(ctx, "sqlite3", dsn(path), logger)
	if err != nil {
		return nil, fmt.Errorf("open sqlstore: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		container.Close()
		return nil, fmt.Errorf("get device: %w", err)
	}
	client := whatsmeow.NewClient(device, logger)

	connected := make(chan struct{}, 1)
	client.AddEventHandler(func(evt any) {
		if _, ok := evt.(*events.Connected); ok {
			select {
			case connected <- struct{}{}:
			default:
			}
		}
	})

	if client.Store.ID == nil {
		// Fresh login — render QR codes to stderr until success.
		qrChan, err := client.GetQRChannel(ctx)
		if err != nil {
			container.Close()
			return nil, fmt.Errorf("qr channel: %w", err)
		}
		if err := client.Connect(); err != nil {
			container.Close()
			return nil, fmt.Errorf("connect: %w", err)
		}
		for evt := range qrChan {
			switch evt.Event {
			case "code":
				fmt.Fprintln(os.Stderr, "Scan this QR with WhatsApp → Linked Devices → Link a Device:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stderr)
			case "success":
				// channel closes after this
			case "timeout":
				container.Close()
				return nil, errors.New("QR pairing timed out before scan")
			case "err-client-outdated":
				container.Close()
				return nil, errors.New("whatsmeow client version is too old; upgrade clank")
			case "err-scanned-without-multidevice":
				fmt.Fprintln(os.Stderr, "Enable multi-device mode in WhatsApp and re-scan.")
			case "error":
				container.Close()
				return nil, fmt.Errorf("pair error: %w", evt.Error)
			default:
				fmt.Fprintln(os.Stderr, "QR event:", evt.Event)
			}
		}
	} else {
		if err := client.Connect(); err != nil {
			container.Close()
			return nil, fmt.Errorf("connect: %w", err)
		}
	}

	// Block until events.Connected (auth done) or timeout.
	select {
	case <-connected:
	case <-time.After(30 * time.Second):
		client.Disconnect()
		container.Close()
		return nil, errors.New("timed out waiting for events.Connected after Connect()")
	case <-ctx.Done():
		client.Disconnect()
		container.Close()
		return nil, ctx.Err()
	}

	return &Handle{Client: client, container: container}, nil
}

// Handle bundles a connected client + container so we can clean both up.
type Handle struct {
	Client    *whatsmeow.Client
	container *sqlstore.Container
}

func (h *Handle) Close() {
	h.Client.Disconnect()
	_ = h.container.Close()
}

// Lookup runs a single IsOnWhatsApp + (if present) UserInfo + ProfilePicture for a phone number.
func (h *Handle) Lookup(ctx context.Context, phone string) (*Result, error) {
	resps, err := h.Client.IsOnWhatsApp(ctx, []string{phone})
	if err != nil {
		return nil, fmt.Errorf("IsOnWhatsApp: %w", err)
	}
	if len(resps) == 0 {
		// Issue #1086: WA dropped this number entirely (likely invalid format).
		return &Result{Query: phone, Registered: false, Reason: "no response (likely invalid number)"}, nil
	}
	r := resps[0]
	out := &Result{Query: r.Query, JID: r.JID.String(), Registered: r.IsIn}
	if r.VerifiedName != nil && r.VerifiedName.Details != nil {
		out.VerifiedBusinessName = r.VerifiedName.Details.GetVerifiedName()
	}
	if !r.IsIn {
		return out, nil
	}
	infoMap, err := h.Client.GetUserInfo(ctx, []types.JID{r.JID})
	if err != nil {
		return out, nil // soft-fail: we still have the IsOn result
	}
	if info, ok := infoMap[r.JID]; ok {
		out.About = info.Status
		out.PictureID = info.PictureID
		out.DeviceCount = len(info.Devices)
		if !info.LID.IsEmpty() {
			out.LID = info.LID.String()
		}
	}
	pic, err := h.Client.GetProfilePictureInfo(ctx, r.JID, &whatsmeow.GetProfilePictureParams{Preview: false})
	if err == nil && pic != nil {
		out.ProfilePictureURL = pic.URL
	}
	return out, nil
}

type Result struct {
	Query                string `json:"query"`
	JID                  string `json:"jid,omitempty"`
	LID                  string `json:"lid,omitempty"`
	Registered           bool   `json:"registered"`
	Reason               string `json:"reason,omitempty"`
	VerifiedBusinessName string `json:"verified_business_name,omitempty"`
	About                string `json:"about,omitempty"`
	PictureID            string `json:"picture_id,omitempty"`
	ProfilePictureURL    string `json:"profile_picture_url,omitempty"`
	DeviceCount          int    `json:"device_count,omitempty"`
}

// Logout unlinks the device server-side AND deletes the local row.
// Returns the path of the still-present sqlite file (caller may os.Remove it).
func (h *Handle) Logout(ctx context.Context) (string, error) {
	err := h.Client.Logout(ctx)
	path, _ := dbPath()
	return path, err
}
```

## 9. CGo decision

| Aspect | `mattn/go-sqlite3` (CGo) | `modernc.org/sqlite` (pure Go) |
|---|---|---|
| Driver name (for `sql.Open` / `sqlstore.New`) | `"sqlite3"` | `"sqlite"` |
| `go install` UX | Requires C compiler (gcc/clang) on the build machine | Just works — no toolchain needed |
| Cross-compile | Painful (need cross-toolchain per arch) | Trivial (`GOOS=linux GOARCH=arm64 go build`) |
| Binary size | ~3–5 MB extra | ~10–15 MB extra (transpiled SQLite) |
| Insert perf | Baseline | ~2× slower |
| Select perf | Baseline | ~10–100% slower |
| dbutil dialect parsing | Accepted (matches `^sqlite`) | Accepted (matches `^sqlite`) |
| whatsmeow's own godoc example | Uses this | — |
| mautrix-whatsapp production go.mod | Uses this (`v1.14.42`) | — |
| DSN pragma syntax | `?_foreign_keys=on&_busy_timeout=5000` | `?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)` |

**Recommendation: `modernc.org/sqlite` (pure Go).** clank ships as a `go install`-able CLI; CGo would force every user to have gcc/Xcode-CLT installed and break `go install` for half our audience. The 2× INSERT slowdown is irrelevant — whatsmeow writes a few KB of session state and rarely after pairing. Whatsmeow's `dbutil.ParseDialect` accepts both `"sqlite"` and `"sqlite3"`, so the swap is one import + one DSN tweak. If a user hits perf issues we can offer a `-tags cgo` build that swaps in mattn.

If you go modernc, the integration in §8 changes to:
```go
_ "modernc.org/sqlite" // registers "sqlite" driver

// ...
container, err := sqlstore.New(ctx, "sqlite", "file:"+path+
    "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(wal)", logger)
```

## 10. Open questions / gotchas to watch in code

1. **`IsOnWhatsApp` may return fewer items than you sent.** Issue #1086 — invalid numbers are silently dropped. Always reconcile by `Query` field, not by index. The §8 code handles `len(resps) == 0`; extend to handle partial returns when batching.
2. **`Connect()` returning `nil` does NOT mean authenticated.** Wait for `events.Connected` (or use `client.WaitForConnection(timeout)`) before issuing IQs — otherwise `IsOnWhatsApp` will fail with `failed to send usync query: websocket not connected` (issue #297, recurring).
3. **Foreign keys must be on.** `Container.Upgrade` returns `"foreign keys are not enabled"` if you forget the DSN pragma. Easy to silently break when copy-pasting DSNs.
4. **`whatsmeow_device` row auto-delete on logout-from-other-side.** `connectionevents.go` calls `cli.Store.Delete(ctx)` on `LoggedOut` reasons (401/403/406). The next clank invocation will see `Store.ID == nil` and re-show the QR — usually fine, but surprising. Detect this by registering an `events.LoggedOut` handler.
5. **No release tags exist** — the package is `v0.0.0-<date>-<sha>` pseudo-version. `go.sum` pinning is the only safety net. Plan for `go get -u go.mau.fi/whatsmeow` to occasionally surface API changes (rare for our APIs but possible).
6. **`waLog.Stdout` writes to `os.Stdout`**, not stderr. If clank pipes JSON to stdout, you must use `waLog.Noop` or write a custom logger that wraps stderr.
7. **`@lid` migration is partial.** `IsOnWhatsApp` returns `@s.whatsapp.net`; `GetUserInfo` populates `info.LID` on `@lid`. Some methods accept either; some accept only one. For lookup-only flows we don't care, but if clank ever stores JIDs persistently, store both.
8. **Profile picture URLs are short-lived.** The `URL` field in `ProfilePictureInfo` is a presigned WhatsApp media URL; download it immediately or expect 403 within minutes.
9. **`TemporaryBan` event is permanent for the session** — `EnableAutoReconnect` will not re-try. Surface the ban code (`evt.Code`) and `Expire` to the user; the SQLite file is *not* auto-deleted, so they can wait it out and re-run.
10. **`PrePairCallback` and `PrePairCallback`-style hooks must be set BEFORE `Connect`.** No-op for our flow but tempting to add later for "is this the right phone?" UX.
11. **Concurrent lookups are safe but not free.** The library uses one websocket; bursts of >50 IQs/min will get throttled by WA server, not by whatsmeow. Throttle in clank if you ever batch.
12. **WhatsApp protocol-version bumps require library updates.** When users see `events.ClientOutdated` or `err-client-outdated`, clank can't fix it — they need a new clank build. Print a clear "upgrade clank" message.

---

## Sources

- [tulir/whatsmeow on GitHub](https://github.com/tulir/whatsmeow)
- [whatsmeow on pkg.go.dev (canonical example)](https://pkg.go.dev/go.mau.fi/whatsmeow)
- [client.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/client.go), [user.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/user.go), [qrchan.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/qrchan.go), [connectionevents.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/connectionevents.go), [types/events/events.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/types/events/events.go), [types/jid.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/types/jid.go), [types/user.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/types/user.go), [util/log/log.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/util/log/log.go), [store/store.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/store/store.go), [store/sqlstore/container.go](https://raw.githubusercontent.com/tulir/whatsmeow/main/store/sqlstore/container.go)
- [go.mau.fi/util/dbutil source (mautrix/go-util)](https://github.com/mautrix/go-util/blob/main/dbutil/database.go)
- [Issue #1086 — IsOnWhatsApp drops invalid numbers](https://github.com/tulir/whatsmeow/issues/1086)
- [Issue #818 — disconnects on heavy-traffic accounts](https://github.com/tulir/whatsmeow/issues/818)
- [Issue #984 — protocol version bumps](https://github.com/tulir/whatsmeow/issues/984)
- [Issue #297 — IsOnWhatsApp before websocket ready](https://github.com/tulir/whatsmeow/issues/297)
- [Discussion #199 — bans](https://github.com/tulir/whatsmeow/discussions/199)
- [Discussion #567 — improved ban rules](https://github.com/tulir/whatsmeow/discussions/567)
- [modernc.org/sqlite docs](https://pkg.go.dev/modernc.org/sqlite)
- [mdp/qrterminal/v3 docs](https://pkg.go.dev/github.com/mdp/qrterminal/v3)
- [WhatsApp linked-devices help (14-day rule)](https://faq.whatsapp.com/378279804439436/?cms_platform=android)
- [Equixly WhatsApp rate-limit disclosure (Dec 2025)](https://equixly.com/blog/2025/12/14/whats-app-api-vulnerability/)
