# 08 — GSM / SIM / IMEI Ecosystem for clank

Research scope: open-source tools, libraries, and public datasets for the mobile-network OSINT layer beneath "phone number lookup" — IMEI/TAC decoding, IMSI structure, USSD codes, IMSI-catcher detection (defensive), the osmocom ecosystem, country/carrier numbering plans, and adjacent device intelligence (MAC OUI). Each subdomain is evaluated for realistic integration into a Go CLI that runs locally.

**Ethical boundary applied throughout:** detection-only, parser-only, and reference-data tooling. SIM cloning, IMSI-catcher attack frameworks, SS7 exploitation, and unauthorized network access are explicitly excluded from integration plans. Where attack frameworks exist (e.g. SigPloit), they are noted purely for situational awareness and tagged "do not integrate."

---

## 1. IMEI Decoding (TAC -> manufacturer/model)

**Structure recap.** IMEI is 15 digits: 8-digit TAC (Type Allocation Code) + 6-digit serial + 1-digit Luhn checksum. Validating the Luhn digit is trivial; mapping TAC -> manufacturer/model requires a database. The authoritative GSMA Device Database is a paid/partner-only resource — open-source projects work around this with a crowdsourced mirror.

### 1.1 Osmocom TAC Database (`osmo-tacdb`)
- URL: https://github.com/osmocom/osmo-tacdb (mirror of gitea.osmocom.org)
- License: CC-BY-SA 3.0 (data); GPL-3.0 (code)
- Stars: ~4 (the GitHub mirror is sparse; primary distribution is `tacdb.osmocom.org`)
- Status: **Active, gold-standard open mirror.** The README on the mirror is sparse (TBD), but the live database at tacdb.osmocom.org is the de-facto open replacement for the restricted GSMA database.
- Format: SQL/CSV exports downloadable from tacdb.osmocom.org
- **clank integration:** embed the CSV at build time or vendor a periodically refreshed snapshot under `internal/data/tacdb.csv`. No shell-out needed — this is pure data. Lookup is `tac[:8] -> {manufacturer, model}`.

### 1.2 jpanganiban/imei
- URL: https://github.com/jpanganiban/imei
- License: not specified (treat as restrictive)
- Stars: 6, single-commit Python repo
- Contains: `db.csv`, `phone.csv`, `tac-new_cleanup.csv`, `babt00_clean.csv` — historical TAC dumps
- **Maintenance: yellow/abandoned.** Useful as a one-shot historical seed but not a live source.

### 1.3 VTSTech/IMEIDB
- URL: https://github.com/Veritas83/IMEIDB
- A public IMEI validator dataset; less curated than osmo-tacdb. Treat as supplementary.

### 1.4 Go-language IMEI parsers
Two repos exist. Both are tiny:
- `flida-dev/go-imei` — MIT, 2 stars, last update May 2025. Idiomatic API:
  ```go
  import "flida.dev/imei"
  v, err := imei.NewIMEI("35-209900-176148-1")
  // v.TAC, v.Serial, v.Checksum
  ```
- `goloop/is` — broader validation grab-bag including IMEI; 4 stars.

Neither has the TAC -> model mapping; both stop at structural parsing + Luhn.

### 1.5 Luhn checksum in Go
Three solid options, all trivial:
- `github.com/EClaesson/go-luhn`
- `github.com/phedde/luhn-algorithm`
- `github.com/neonxp/checksum` (`luhn.Check()`)

Implementing Luhn inline is ~15 lines of Go and avoids a dependency.

### 1.6 GSMA TAC Database (the restricted source)
- URL: https://www.gsma.com/get-involved/working-groups/terminal-steering-group/imei-database/
- Partner-only. **Not for clank.** Note its existence so users understand why open coverage is incomplete (~80–90% of consumer devices are well-mapped via osmo-tacdb; obscure IoT/regional models often missing).

**Concrete clank flow:**
```go
// pseudo: imei.Parse("353209900176148")
// 1. validate length=15
// 2. luhn check on last digit
// 3. extract TAC = imei[:8]
// 4. lookup TAC in embedded osmo-tacdb CSV
// 5. return {manufacturer: "Apple", model: "iPhone 14", luhnValid: true}
```

---

## 2. MEID / ESN (CDMA equivalent)

MEID is 56 bits (14 hex chars); legacy ESN is 32 bits. Used by Verizon/Sprint and a few APAC CDMA holdouts; **largely deprecated** post-3G shutdown.

- No actively maintained open-source MEID/ESN decoding library was found. Closest reference: Bo Bayles' web reference, plus `cwTool` (Windows freeware, closed-source).
- `panigrahip/3GPPDecoder` decodes 3GPP messages but is GSM/UMTS/LTE — not CDMA.

**clank verdict:** skip. Audience overlap with modern targets is near-zero. If needed, ship a `meid validate` command that does pure structural validation (14 hex chars, optional pseudo-ESN derivation via SHA-1) without any database. ~30 lines, no dependency.

---

## 3. USSD / MMI Code Databases

### 3.1 traviszech/Android-Dialer-Codes
- URL: https://github.com/traviszech/Android-Dialer-Codes
- License: CC-BY-4.0
- Stars: 7
- Format: a single curated `ANDROID_SECRET_CODES_MASTER_LIST.md`
- Coverage: Samsung, Pixel, Xiaomi, LG, generic GSM. Includes universal MMI codes (`*#06#` -> IMEI, `*#21#` -> call-forwarding-status, `##002#` -> cancel-all-forwarding, `*#62#` -> when-unreachable forwarding) and per-OEM service menus.

### 3.2 What clank should ship
The standard GSM MMI codes are stable and small (~50 entries). Embedding a curated subset as JSON is the right move:
```json
{"code":"*#06#","desc":"Display IMEI","type":"universal"}
{"code":"*#21#","desc":"Check unconditional call forwarding","type":"universal"}
{"code":"##002#","desc":"Cancel all call forwarding","type":"universal"}
{"code":"*#0*#","desc":"Service menu (Samsung)","type":"oem","vendor":"Samsung"}
```
**clank integration:** `clank ussd lookup '*#21#'` -> returns description. Pure embedded data, no runtime dependency. Source: derive from the Android-Dialer-Codes markdown + the GSM 02.30 / 22.030 standards (which the MMI syntax follows).

---

## 4. IMSI / SIM Toolkit

**Structure.** IMSI = 15 digits: MCC (3) + MNC (2 or 3, ITU-T E.212) + MSIN (the rest). Parsing requires the MCC/MNC list to know whether MNC is 2 or 3 digits per country.

### 4.1 pySim (osmocom) — gold standard
- URL: https://github.com/osmocom/pysim
- License: GPL-2.0
- Stars: 535
- Last activity: actively maintained (most recent commit March 2026); 1,687 commits on master
- Supports: SIM, USIM, ISIM, HPSIM, eUICC
- CLI tools: `pySim-shell.py`, `pySim-read.py`, `pySim-prog.py`, `pySim-trace.py`, `osmo-smdpp.py`
- Requires: programmable SIM cards + PC/SC reader hardware
- **clank integration:** clank is a network-layer OSINT CLI; pySim assumes physical card access. **Do not embed.** Recommended posture: emit a `clank sim` subcommand that *recommends* pySim when the user wants on-card operations and does not pretend to do hardware work clank cannot do. Keep clank in pure parser territory.

### 4.2 IMSI parser libraries
- `atis--/imsi-grok` — JS, breaks IMSI into MCC/MNC/MSIN with country/operator lookup. Algorithm portable.
- `Oros42/IMSI-catcher` — Python; ships an `mcc-mnc/mcc_codes.json` reference file; CC0-1.0; 4.2k stars. Most useful piece is the JSON, not the catcher script.

### 4.3 SIM ATR (Answer-To-Reset) parsers
- `dsicari/ATR-Parser` — C, basic ATR field decoder
- `LudovicRousseau/pyscard-contrib/parseATR` — Python, the reference implementation used by smartcard-atr.apdu.fr
- ATR is hardware-side; only relevant to clank if a future reader-integration ships. **Defer.**

### 4.4 What clank should ship
A pure-Go IMSI parser using an embedded MCC/MNC table (see §7 below). API:
```go
i, err := imsi.Parse("310410123456789")
// i.MCC = "310" -> US
// i.MNC = "410" -> AT&T
// i.MSIN = "123456789"
// i.Country, i.Operator from embedded table
```
~80 lines + a JSON table. The 2-vs-3-digit-MNC lookup is the only nontrivial part — handle it via the same MCC/MNC dataset.

---

## 5. IMSI-Catcher Detection (Stingray detection, defensive only)

### 5.1 SnoopSnitch (srlabs)
- URL: https://github.com/srlabs/snoopsnitch
- License: GPL-3.0
- 993 commits; the canonical research-grade detector
- Detects: fake base stations, silent SMS, SS7-style anomalies, OTA security
- **Hard requirement:** rooted Android phone with a Qualcomm chipset (uses Qualcomm DIAG)
- Status: alive but specialized; SRLabs treats it as a research artifact more than a consumer app.

### 5.2 AIMSICD (CellularPrivacy)
- URL: https://github.com/CellularPrivacy/Android-IMSI-Catcher-Detector
- License: GPL-3.0
- 2,707 commits, alpha; explicit "revival underway, light version planned" note
- **Maintenance: yellow.** Long-running stagnation, occasional bursts. The official issue thread (#926) acknowledges multi-year inactivity. The `5GSD/AIMSICDL` fork is a more recent reload attempt.

### 5.3 darshakframework/darshak
- URL: https://github.com/darshakframework/darshak
- **Maintenance: dead** since ~2015. Historical interest only.

### 5.4 Public base-station fingerprint datasets
- OpenCellID and Mozilla Location Service (MLS) are the realistic open data sources for cell-tower fingerprints. Both are HTTPS APIs with bulk downloads.
- For clank: see research file 02 (lookup APIs) for the network-side; here it suffices to note that fingerprint comparison logic is not feasible without per-device radio access.

**clank verdict:** mention SnoopSnitch as the recommended Android-side companion tool; clank itself cannot detect IMSI catchers from a desktop CLI without SDR hardware. Do not pretend otherwise.

---

## 6. GSM / Mobile-Network Analysis Tooling (osmocom + SDR ecosystem)

### 6.1 osmocom suite
All osmocom GitHub repos are mirrors; primary dev happens at gitea.osmocom.org / gerrit.osmocom.org.

| Project | What it is | Stars | Status |
|---|---|---|---|
| `osmocom/osmocom-bb` | Mobile Station GSM baseband (open phone OS) | 323 | Active mirror, updated Jan 2026 |
| `osmocom/osmo-bts` | GSM Base Transceiver Station | 108 | Active |
| `osmocom/osmo-msc` | 3GPP Mobile Switching Centre (2G/3G) | 36 | Active, updated Jan 2026 |
| `osmocom/osmo-bsc` | GSM Base Station Controller | — | Active |
| `osmocom/osmo-tacdb` | TAC database (covered §1.1) | 4 | Active mirror |
| `osmocom/pysim` | SIM/USIM/ISIM tooling | 535 | Active (covered §4.1) |

These are operator-side network elements for running your own GSM/UMTS lab. **Out of scope for a CLI** — clank cannot meaningfully integrate. Reference them for the power-user reader; do not shell out.

### 6.2 SDR-based RAN stacks
- `srsran/srsRAN_Project` — **archived Feb 17, 2026.** Development moved to the new `OCUDU` repo. The `srsRAN_4G` repo also exists as a separate 4G project.
- `simula/openairinterface5g` — actively maintained by the OpenAirInterface Software Alliance (OSA).

Same verdict: full-stack RAN tooling is for SDR labs. Out of clank's scope.

### 6.3 Wireshark dissectors
The official `wireshark/wireshark` repo includes mature dissectors for `gsmtap`, `gsm_a` (3GPP 04.08), `gsm_map`, `gsm_sim`, `gsm_sms`, `gsm_gsup`, plus full LTE RRC/NAS chains. Use Wireshark for `.pcap` analysis; clank does not need to embed any of this. If we ever want to parse a GSMTAP capture, calling `tshark -r capture.pcap -Y gsmtap` is the right shell-out.

### 6.4 SS7 attack tooling (do-not-integrate watchlist)
- `SigPloiter/SigPloit` — Python framework targeting SS7/GTP/Diameter/SIP. ~786 commits; not archived. **For authorized red-team / operator-pentest use only.** clank must not integrate this; even shelling out into it is the wrong message for a phone-OSINT CLI.
- `panigrahip/3GPPDecoder` — opensource 3GPP message decoder (LTE/UMTS/GSM); **passive decode, fine to reference** for educational links, not for integration.

---

## 7. Country / Carrier Prefix Datasets

### 7.1 MCC/MNC tables (E.212)
| Repo | License | Stars | Status | Notes |
|---|---|---|---|---|
| `pbakondy/mcc-mnc-list` | MIT | 128 | Active | ~2189 records as JSON, schema includes brand/operator/status/bands |
| `P1sec/MCC_MNC` | AGPL-3.0 (code); CC-BY-SA (data) | 86 | Active (Feb 2026) | Aggregates Wikipedia + ITU-T bulletins + CIA Factbook + txtNation |
| `musalbas/mcc-mnc-table` | MIT | 363 | **Archived Nov 2024** | Was daily-updated; now stale |
| `Oros42/IMSI-catcher mcc-mnc/` | CC0-1.0 | 4.2k (parent) | Active | JSON snapshot from the catcher repo |

**Pick:** `pbakondy/mcc-mnc-list` (clean schema, MIT, active) as primary; `P1sec/MCC_MNC` as enrichment when you want network-type/MSISDN-prefix detail (AGPL is fine if we use it as data, not embed code).

Sample record (pbakondy):
```json
{"type":"National","countryName":"Hungary","countryCode":"HU","mcc":"216","mnc":"30","brand":"T-Mobile","operator":"Magyar Telekom Plc","status":"Operational","bands":"GSM 900 / GSM 1800 / UMTS 2100 / LTE 800 / LTE 1800 / LTE 2600","notes":null}
```

### 7.2 ITU-T E.164 country code lists
- `biter777/countries` — Go library covering ISO-639/3166/4217 + E.164 IDD codes; very comprehensive. **Direct fit for clank.**
- `datasets/country-codes` — pure data, includes ITU dialing codes
- `geneh/e164-phones-countries` — narrow E.164-only mapping

### 7.3 Per-country numbering plans
- **India:** `hstsethi/in-mob-prefix` — CC-BY-4.0, CSV by 4-digit prefix range mapping to operator + state/circle. Last updated Oct 2025. Source: TRAI + DoT + Wikipedia. **Embed.**
- **NANPA (US/CA):** the official NANPA dataset is downloadable; no curated GitHub mirror dominates. NANP area-code -> region is small enough to embed as JSON.
- **UK Ofcom, EU national plans:** scattered; no single canonical open repo. Build per-country JSON files on-demand from official regulator releases.

**clank integration:** `internal/data/numbering/{us,uk,in,...}.json` with prefix -> {region, carrier, type}. Lookup happens after E.164 parsing (already covered by libphonenumber in research file 04).

---

## 8. Mobile-Device Intelligence Beyond IMEI: MAC OUI

### 8.1 Datasets
| Repo | Format | Source | Notes |
|---|---|---|---|
| `Ringmast4r/OUI-Master-Database` | TXT/CSV/TSV/JSON/XML/SQLite/SQL/Kismet | IEEE + Wireshark + Nmap + HDM | 88,212+ vendors. MIT. Updated April 2026. |
| `uxmansarwar/mac-address-vendor-database` | JSON/CSV/SQL | IEEE | Always-current via auto-sync |
| `ttafsir/mac-oui-lookup` | JSON | IEEE | Daily GitHub Action refreshes |

### 8.2 Go libraries
- `klauspost/oui` — MIT, 31 stars, **archived March 2020.** Still works (in-memory lookups, claimed millions/sec) but unmaintained. Forking-and-adopting is the realistic path if performance matters. For clank's volume, a simple in-memory map from a CSV is faster to ship than a vendored library.
- `silverwind/oui` (Node) and `ouilookup` (Python) exist as references.

### 8.3 clank usage
Most clank users won't have a MAC address attached to a phone-number lookup. The realistic intersection is when a workflow already produces network captures (Wi-Fi probes, Bluetooth scans) and the user wants vendor identification. Ship `clank oui lookup AA:BB:CC` as a small bonus utility wired to an embedded JSON. This is one of the cheapest wins in the whole research file.

---

## 9. Realistic clank Integration Ranking

Ranked by ratio of (user value) / (build cost), under the constraint that clank is a single Go binary that runs locally without SDR or smartcard hardware.

### Tier 1 — Ship in the next milestone
1. **MCC/MNC table + IMSI parser** (`pbakondy/mcc-mnc-list`). Embed JSON; expose `clank imsi parse <imsi>` and use the same table to enrich any phone-number lookup with operator/country detail. Pure Go, ~150 LOC + ~200KB embedded data. **Highest leverage.**
2. **IMEI parser + osmo-tacdb integration** (`flida-dev/go-imei` style + embedded osmo-tacdb CSV snapshot). Expose `clank imei parse <imei>`. Validates Luhn, decodes TAC -> manufacturer/model. ~100 LOC + ~5MB embedded CSV (or refresh-on-first-use).
3. **MMI/USSD code reference** (curated subset of Android-Dialer-Codes + GSM 02.30). Embed ~100 entries as JSON. `clank ussd lookup '*#21#'`. Trivial to build, instantly differentiating in the OSINT-CLI niche.

### Tier 2 — Ship later if there's appetite
4. **MAC OUI lookup** (`Ringmast4r/OUI-Master-Database` JSON dump, refreshed). `clank oui lookup AA:BB:CC`. Cheap utility; only useful when a workflow has captured MACs alongside phone data.
5. **Per-country numbering-plan tables.** Start with India (`hstsethi/in-mob-prefix`) and NANPA. Wire into existing E.164 normalization.

### Tier 3 — Document, do not integrate
6. **pySim** — recommend as the on-card companion when users ask about SIM file system / IMSI-write-back / eSIM. clank stays network-side.
7. **SnoopSnitch / AIMSICD** — recommend as Android companion for IMSI-catcher detection; clank cannot replicate this from a desktop without SDR hardware.
8. **osmocom suite, srsRAN, OpenAirInterface5G, Wireshark dissectors** — link in docs as advanced-user tooling. No integration.

### Tier 0 — Explicitly excluded
- **SigPloit, ss7-attack forks, AIMSICD-attack-mode patches, MEID/ESN cloning, SIM cloning utilities.** Even mentioning them in clank's UI is the wrong signal. Acknowledged here only so future contributors do not accidentally re-research and pull them in.

---

## Sources

- Osmocom TAC DB: https://github.com/osmocom/osmo-tacdb
- jpanganiban IMEI: https://github.com/jpanganiban/imei
- pySim: https://github.com/osmocom/pysim
- SnoopSnitch: https://github.com/srlabs/snoopsnitch
- AIMSICD: https://github.com/CellularPrivacy/Android-IMSI-Catcher-Detector
- Android-Dialer-Codes: https://github.com/traviszech/Android-Dialer-Codes
- pbakondy/mcc-mnc-list: https://github.com/pbakondy/mcc-mnc-list
- P1sec/MCC_MNC: https://github.com/P1sec/MCC_MNC
- musalbas/mcc-mnc-table: https://github.com/musalbas/mcc-mnc-table
- Oros42/IMSI-catcher: https://github.com/Oros42/IMSI-catcher
- flida-dev/go-imei: https://github.com/flida-dev/go-imei
- Ringmast4r OUI DB: https://github.com/Ringmast4r/OUI-Master-Database
- klauspost/oui: https://github.com/klauspost/oui
- hstsethi/in-mob-prefix: https://github.com/hstsethi/in-mob-prefix
- biter777/countries: https://github.com/biter777/countries
- osmocom/osmocom-bb: https://github.com/osmocom/osmocom-bb
- osmocom/osmo-bts: https://github.com/osmocom/osmo-bts
- osmocom/osmo-msc: https://github.com/osmocom/osmo-msc
- srsRAN (archived): https://github.com/srsran/srsRAN_Project
- OpenAirInterface5G: https://github.com/simula/openairinterface5g
- Wireshark: https://github.com/wireshark/wireshark
- SigPloit (do-not-integrate): https://github.com/SigPloiter/SigPloit
- GSMA TAC DB (restricted, reference only): https://www.gsma.com/get-involved/working-groups/terminal-steering-group/imei-database/
