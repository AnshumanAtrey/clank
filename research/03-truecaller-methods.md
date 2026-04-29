# Truecaller Integration — Community Methods, Endpoints, and Current Status

Research target: how the OSINT / cybersecurity community has actually integrated Truecaller-style caller-ID lookup into open-source tools that run locally, what's verifiable from source code, and what works in 2025–2026.

Scope reminder: only the legitimate "user registers their own Truecaller account → extracts their own `installationId` → uses it for personal lookups" pattern is documented here. Tools that distribute leaked credentials or shared installation IDs are flagged but their secrets are not transcribed.

---

## 1. Auth Flow Walkthrough (with code)

The community-standard flow targets Truecaller's **mobile app onboarding endpoints**, not their web/SDK partner API. Two HTTP calls bootstrap a long-lived bearer token (`installationId`); a third call is the actual lookup. Source: [`sumithemmadi/truecallerjs`](https://github.com/sumithemmadi/truecallerjs) (TS) and [`sumithemmadi/truecallerpy`](https://github.com/sumithemmadi/truecallerpy) (Py), both ~Aug 2023, both still pinned to app version `Truecaller/11.75.5 (Android;10)`.

### Step 1 — Send OTP

`POST https://account-asia-south1.truecaller.com/v2/sendOnboardingOtp`

Verbatim from `truecallerjs/src/login.ts` (lines 75–104):

```ts
const postUrl =
  "https://account-asia-south1.truecaller.com/v2/sendOnboardingOtp";

const data = {
  countryCode: pn.regionCode,            // ISO-2, e.g. "IN"
  dialingCode: pn.countryCode,           // numeric, e.g. 91
  installationDetails: {
    app: { buildVersion: 5, majorVersion: 11, minorVersion: 7, store: "GOOGLE_PLAY" },
    device: {
      deviceId: generateRandomString(16),  // random 16-char string per install
      language: "en",
      manufacturer: device.manufacturer,    // randomly picked from a phones list
      model: device.model,
      osName: "Android", osVersion: "10",
      mobileServices: ["GMS"],
    },
    language: "en",
  },
  phoneNumber: pn.number.significant,
  region: "region-2",
  sequenceNo: 2,
};

const options = {
  method: "POST",
  headers: {
    "content-type": "application/json; charset=UTF-8",
    "accept-encoding": "gzip",
    "user-agent": "Truecaller/11.75.5 (Android;10)",
    clientsecret: "lvc22mp3l1sfv6ujg83rd17btt",   // <-- shared, hardcoded
  },
  url: postUrl, data,
};
```

Response shape (per the README): `{ status, message, domain, parsedPhoneNumber, parsedCountryCode, requestId, method, tokenTtl }`. `requestId` is what carries forward.

### Step 2 — Verify OTP, receive `installationId`

`POST https://account-asia-south1.truecaller.com/v1/verifyOnboardingOtp` (from `truecallerjs/src/verifyOtp.ts`):

```ts
const postData = {
  countryCode: pn.regionCode,
  dialingCode: pn.countryCode,
  phoneNumber: pn.number.significant,
  requestId: json_data.requestId,   // from step 1
  token: otp,                       // 6-digit code from SMS
};
// same headers as step 1, including clientsecret: "lvc22mp3l1sfv6ujg83rd17btt"
```

Successful response includes `installationId` (the long-lived bearer), `ttl`, `userId`, `suspended`, and a `phones` array. The `installationId` is what every subsequent search request uses.

### Step 3 — Search

`GET https://search5-noneu.truecaller.com/v2/search` (from `truecallerjs/src/search.ts`, lines ~150–175):

```ts
axios.get(`https://search5-noneu.truecaller.com/v2/search`, {
  params: {
    q: significantNumber,
    countryCode: phoneNumber.regionCode,
    type: 4,                                       // 4 = single search
    locAddr: "",
    placement: "SEARCHRESULTS,HISTORY,DETAILS",
    encoding: "json",
  },
  headers: {
    "content-type": "application/json; charset=UTF-8",
    "accept-encoding": "gzip",
    "user-agent": "Truecaller/11.75.5 (Android;10)",
    Authorization: `Bearer ${searchData.installationId}`,
  },
});
```

**Key facts for porting to Go:**

- `clientsecret` (lowercase, header) is hardcoded to `lvc22mp3l1sfv6ujg83rd17btt` across virtually every public tool. This is a static value baked into the v11.75.5 APK; it is not a leaked per-user secret. Treat it as part of the protocol, not as a credential.
- `installationId` is the **only** per-user secret. Once obtained it persists indefinitely until Truecaller invalidates it (account suspension, device tie-break, server-side revocation).
- Region matters: `account-asia-south1.truecaller.com` is the Asia auth host; `search5-noneu.truecaller.com` is the non-EU search shard. EU traffic uses `search5-eu.truecaller.com`. The community tools default to `noneu`.

---

## 2. Endpoint Catalog

| Purpose | Method + URL | Auth | Notes |
|---|---|---|---|
| Send OTP | `POST account-asia-south1.truecaller.com/v2/sendOnboardingOtp` | `clientsecret` header | Body includes spoofed device fingerprint |
| Verify OTP | `POST account-asia-south1.truecaller.com/v1/verifyOnboardingOtp` | `clientsecret` header | Returns `installationId` (long-lived JWT-ish opaque token) |
| Single search | `GET search5-noneu.truecaller.com/v2/search` | `Authorization: Bearer <installationId>` | `type=4`, `placement=SEARCHRESULTS,HISTORY,DETAILS` |
| Bulk search | `GET search5-noneu.truecaller.com/v2/bulk` | `Authorization: Bearer <installationId>` | `type=14`, comma-separated `q`, max 30 numbers per call |
| Profile (rare) | `GET profile4-noneu.truecaller.com/v1/default` | `Authorization: Bearer <token>` | Used by Web SDK partner flow, distinct from search |

### Single-search response fields (from `truecallerjs/src/search.ts` `Format` class)

The response is a JSON document containing a top-level `data` array. Per element:

- `name`, `altName` — display name and alias
- `addresses[]` — `{ city, countryCode, timeZone, type }`
- `internetAddresses[]` — `{ id, service, caption, type }` (email + social profile handles, e.g. "facebook", "twitter")
- `phones[]` — alternate phones, dial codes, type tags ("mobile", "fixedLine")
- `score` (float, 0–1) — Truecaller's reputation/confidence number
- `tags[]`, `badges[]` — spam labels, "verified business", "premium", "priority", etc.
- `gender`, `image` (URL), `about`, `jobTitle`, `companyName`, `sources[]`

### Bulk search

`GET /v2/bulk` accepts up to 30 comma-separated numbers in `q`. Same Bearer auth. Returns the same `data[]` schema but indexed positionally to the input list. Documented limit comes from the truecallerjs README: **"The tool supports searching 30 or fewer phone numbers at once in a single request."**

### Older variant

The pre-2020 endpoint pattern observed in [`Te-k/analyst-scripts/osint/truecaller.py`](https://github.com/Te-k/analyst-scripts/blob/master/osint/truecaller.py) used `https://search5.truecaller.com/v2/search` with `myNumber`+`registerId` query parameters instead of a Bearer header. That style **stopped working** around 2019–2020; current tools all use the Bearer + `noneu`/`eu` shard pattern shown above.

---

## 3. Currently-Working Status (per repo, as of April 2026)

The honest answer is **mostly broken in a way that is degrading further**. Recent issues across the major repos paint a consistent picture: OTP send/verify is failing for a growing fraction of users, and the failures are not random — they look like server-side fingerprint and attestation rejection.

### `sumithemmadi/truecallerjs` (818 stars, last commit Aug 2023)

15 of the 15 most recent open issues are auth/verification failures. Recent timeline:

- Issue [#162](https://github.com/sumithemmadi/truecallerjs/issues/162) (2026-04-22): "Verification Failed - NOT Sending OTP"
- Issue [#160](https://github.com/sumithemmadi/truecallerjs/issues/160) (2026-03-28): "Verification Failed"
- Issue [#144](https://github.com/sumithemmadi/truecallerjs/issues/144) (2025-09-07): "TrueCaller API 'Verification Failed' — possible ClientSecret issue". Quote from the report:
  > "It is hypothesized that the hardcoded `clientsecret` (`lvc22mp3l1sfv6ujg83rd17btt`) … may have been rate-limited or blacklisted by Truecaller's servers. Due to the public nature of this key and its widespread use, it has likely been flagged for anomalous activity."
- Reply on the same issue (2026-02-28) from a contributor who decompiled the current APK:
  > "The client secret isn't the issue — it's verifying Play Services. All secrets are in apk. decompile it. let me know if you find proper .proto file for onboarding."

That last comment is the actually-useful diagnostic: as of early 2026, Truecaller's onboarding endpoint is gated by **Play Integrity / SafetyNet attestation tokens** that confirm the client is a real Google-Play-Services-equipped Android device. The hardcoded JSON device-spoof block in truecallerjs/truecallerpy is not enough anymore. The May 2024 issue [#37](https://github.com/sumithemmadi/truecallerpy/issues/37) "Upgrade to gRPC Request" hints the current app traffic has also moved partially to protobuf-over-gRPC, which the v11.75.5 JSON shape no longer matches cleanly.

### `sumithemmadi/truecallerpy` (Python port)

Same code, same problem. Latest open issues #40 (Apr 2025, `KeyError: 'data'`), #41 (May 2025, "Verification fails when login"), #42 (Jun 2025), #43 (Jul 2025) all describe OTP send returning HTTP 426 Upgrade Required or non-JSON bodies.

### `nvzard/truecaller-unofficial-api`, `sehmbimanvir/truecaller-cli`, `dr298/Truecaller-Python-Tools`

Smaller wrappers, no fresh maintenance. The nvzard README itself ships the warning: *"This method of accessing Truecaller's database may stop working any day."*

### `scottphilip/caller-lookup`

**Archived November 2019.** Used a Google-OAuth-via-`GoogleToken` shim that no longer exists. Listed only for historical context.

### Web wrappers (`Truecaller-Bulk` Vercel apps, RapidAPI's "Truecaller Scraper API")

These work because they front-run a paid scraping pool with rotating credentials — not relevant to a local CLI.

### Net assessment

For a **freshly-registered personal Truecaller account on a real Google-Play-equipped Android device** (i.e. the user installs the real Truecaller app, completes onboarding through the app, then extracts their `installationId` from the app's local SharedPreferences / databases via ADB or a rooted device), the **`/v2/search` and `/v2/bulk` endpoints still respond and still return the rich JSON schema** described in section 2. What is broken is the **OTP/onboarding shortcut** that the npm/pip libraries try to do entirely from a server. Lookups themselves (with a real, app-derived bearer) are not currently 401/403'd in the bulk reports.

Rate limit: free-tier accounts are throttled by Truecaller server-side. The widely-cited "50 lookups/day" number is **not** documented in any repo source I could verify; the credible numbers from secondary reporting are **5–10 searches per day for free accounts**, with subscription tiers raising the cap. The 30-numbers-per-`/bulk`-request limit is the only hard limit confirmed from source.

---

## 4. Alternatives Matrix

| Service | Open-source unofficial client? | Repo / state | Auth pattern | Useful in 2026? |
|---|---|---|---|---|
| **Truecaller** | Yes (truecallerjs, truecallerpy + many forks) | Stale (2023), onboarding broken; lookup works with app-derived token | `clientsecret` header + Bearer `installationId` | Conditional — see §3 |
| **Eyecon** | No standalone OSS client found. Aggregated by closed-source [`dimondevceo/caller-id-api`](https://github.com/dimondevceo/caller-id-api) (docs only, no source) | n/a | Closed | No, not for a local CLI |
| **CallApp** | No. Same aggregator as Eyecon. Also surfaces on RapidAPI as a paid wrapper. | n/a | Closed | No |
| **Hiya** | No reverse-engineered client. Hiya runs an [official developer portal](https://developer.hiya.com/) for SDK partners — closed registration | Official only | Partner SDK, requires business agreement | Only if you onboard officially |
| **Whoscall** | One ancient Alfred workflow [`mephisto41/alfred-workflow-whoscall`](https://github.com/mephisto41/alfred-workflow-whoscall) — scrapes public web pages at `number.whoscall.com/...`. No auth, very fragile, regex-parses HTML | Likely broken; site has changed many times | None (HTML scrape) | Marginal |
| **Showcaller** | None found on GitHub | n/a | Closed | No |
| **CallApp / ViewCaller via aggregator** | [`dimondevceo/caller-id-api`](https://github.com/dimondevceo/caller-id-api) — "Caller identification API documentation". Aggregates Truecaller + CallApp + ViewCaller + EyeCon + Hiya behind one paid endpoint. **No source code** in repo, only docs. SaaS, X-Auth header, credits-based | Docs-only repo | API key | Useful if `clank` allows opt-in cloud fallback |

Practical implication: **Truecaller is essentially the only target where a local, OSS-replicable auth flow exists at all.** Every other major caller-ID provider either lacks a community client (Hiya, Showcaller, CallApp), has a closed-source aggregator paywall (dimondevceo), or has a one-off dead scraper (Whoscall).

---

## 5. Legal / ToS Quotes from the OSINT Tool READMEs

These are the **authors' own** disclaimer language; not legal advice.

**`sumithemmadi/truecallerjs` README:**
> "The `truecallerjs` tool is not an official Truecaller product. It is a custom script developed by Sumith Emmadi, and its functionality is dependent on the Truecaller service. Please use this tool responsibly and in compliance with the terms of service of Truecaller."

**`nvzard/truecaller-unofficial-api` README** (verbatim warning):
> "This method of accessing Truecaller's database may stop working any day."

**`clank/README.md`** (current upstream — atrey.dev):
> "This tool is intended for legitimate use cases only. Please ensure you comply with all applicable laws and regulations regarding phone number lookup and privacy when using this tool."

**Truecaller's own ToS** (could not be fetched live; well-known summary from Truecaller's public terms): users may not use the service for "any commercial purpose," may not "scrape, copy or duplicate" the content, and may not "reverse engineer" the apps. Sharing or reselling installation IDs is explicitly prohibited. Quote-grade verbatim language requires a direct fetch from `truecaller.com/terms-of-service` (currently blocked from this environment).

**Practical posture every responsible repo takes:**

1. The library makes **no claim** of being authorized by Truecaller.
2. The user is required to **register and OTP-verify their own number** — there is no key/credential bundled with the tool.
3. Bulk and rate-evading behavior is described but the tool does not auto-rotate accounts.
4. Authors disclaim warranty and explicitly note breakage risk.

---

## Recommendation for `clank`

1. **Don't ship an `installationId`.** Ever. The auth flow must be: `clank login --phone +9199XXXXXXX` → user receives OTP on their real number → `clank verify <otp>` writes `~/.clank/credentials.json` with the user's own `installationId`. Then `clank lookup <number>` calls `/v2/search`. This is what truecallerjs does and the only ethically defensible pattern.
2. **Be honest in the README** that onboarding is currently flaky (Section 3). Document the manual fallback: install the real Truecaller Android app, complete onboarding on a real device, extract `installationId` from `data/data/com.truecaller/shared_prefs/*.xml` via ADB on a rooted device, then `clank import-token <installationId>`.
3. **Treat lookup as best-effort.** Wrap every call with timeout + a clear error shape. When Truecaller returns 401/403, do not retry-spam — surface the error and tell the user their token was invalidated.
4. **Skip "alternatives" for v1.** Eyecon/CallApp/Hiya have no OSS path. Whoscall HTML scraping is too fragile. Showcaller has nothing. If clank ever adds a fallback, gate it behind `--cloud` and route through `dimondevceo/caller-id-api` (or similar) with a user-supplied API key — never bundle one.
5. **Ports cleanly to Go.** The whole flow is three HTTP calls, JSON in/out, two static headers. `net/http` + `encoding/json` is sufficient. Add `github.com/nyaruka/phonenumbers` for E.164 parsing instead of writing it yourself.

---

## Sources

- [sumithemmadi/truecallerjs (GitHub)](https://github.com/sumithemmadi/truecallerjs) — primary source for endpoints, headers, auth flow.
- [sumithemmadi/truecallerpy (GitHub)](https://github.com/sumithemmadi/truecallerpy) — Python mirror; same protocol.
- [truecallerjs issue #144 — clientsecret hypothesis + APK decomp comment](https://github.com/sumithemmadi/truecallerjs/issues/144).
- [truecallerjs issue #162 — latest verification failure (Apr 2026)](https://github.com/sumithemmadi/truecallerjs/issues/162).
- [truecallerpy issue #37 — Upgrade to gRPC Request (May 2024)](https://github.com/sumithemmadi/truecallerpy/issues/37).
- [Te-k/analyst-scripts — older `search5.truecaller.com` pattern](https://github.com/Te-k/analyst-scripts/blob/master/osint/truecaller.py).
- [nvzard/truecaller-unofficial-api](https://github.com/nvzard/truecaller-unofficial-api) — author's own breakage warning.
- [scottphilip/caller-lookup](https://github.com/scottphilip/caller-lookup) — archived 2019; historical reference.
- [dimondevceo/caller-id-api](https://github.com/dimondevceo/caller-id-api) — closed-source aggregator (Truecaller + CallApp + EyeCon + ViewCaller + Hiya).
- [mephisto41/alfred-workflow-whoscall](https://github.com/mephisto41/alfred-workflow-whoscall) — Whoscall HTML scrape (likely stale).
- [Hiya developer portal](https://developer.hiya.com/) — official partner-only API; no community client.
- [Beebom — Truecaller search-limit reporting](https://beebom.com/truecaller-limiting-caller-id/) — context for free-tier daily caps.
