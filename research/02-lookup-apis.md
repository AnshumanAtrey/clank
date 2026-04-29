# Phone-Number Lookup APIs for `clank`

Catalog of phone-number lookup providers that an individual dev can sign up for in 5 minutes (no business KYC, no enterprise sales call). Sorted by `free-tier generosity x data richness` — the most useful free options first. Each entry has enough info to wire a Go HTTP client straight from this doc.

All free tiers verified accessible as of April 2026. Skipped: VeriCall (Pindrop enterprise-only, no public signup) and `phone.email` (it's an OTP-send API, not a lookup API — no use for OSINT).

---

## 1. IPQualityScore Phone Validation

- **Provider:** https://www.ipqualityscore.com
- **Signup:** https://www.ipqualityscore.com/create-account
- **Auth:** API key in URL path
- **Free tier:** 1,000 lookups/month, 35/day, no credit card
- **Paid:** Startup $99/mo for 5,000 lookups
- **Endpoint:** `GET https://www.ipqualityscore.com/api/json/phone/{API_KEY}/{phone}`
- **Returns:** `valid`, `fraud_score` (0-100), `carrier`, `line_type` (Mobile/Landline/VOIP/Toll Free), `prepaid`, `risky`, `recent_abuse`, `VOIP`, `active`, `country`, `region`, `city`, `timezone`, `do_not_call`

```bash
curl "https://www.ipqualityscore.com/api/json/phone/YOUR_KEY/14158586273"
```
```json
{"success":true,"valid":true,"fraud_score":44,"recent_abuse":false,
 "VOIP":false,"prepaid":false,"risky":true,"active":true,
 "carrier":"Verizon Wireless","line_type":"Wireless",
 "country":"US","region":"California","city":"San Francisco",
 "timezone":"America/Los_Angeles","do_not_call":false}
```

- **Why top pick:** Only free tier that ships a real fraud score; ideal for OSINT triage.
- **Reliability:** Active, docs updated regularly, well-known provider.

---

## 2. Veriphone

- **Provider:** https://veriphone.io
- **Signup:** https://veriphone.io/signup (no card)
- **Auth:** Bearer token in `Authorization` header
- **Free tier:** 1,000 lookups/month
- **Paid:** Starter $6.99/mo for 5,000
- **Endpoint:** `GET https://api.veriphone.io/v2/verify?phone={E164}`
- **Returns:** `phone_valid`, `phone_type` (mobile/fixed_line/voip/toll_free/premium_rate/...), `carrier`, `country`, `country_code`, `phone_region`, `e164`, `international_number`, `local_number`

```bash
curl "https://api.veriphone.io/v2/verify?phone=%2B14158586273" \
     -H "Authorization: Bearer YOUR_KEY"
```
```json
{"status":"success","phone":"+14158586273","phone_valid":true,
 "phone_type":"mobile","phone_region":"California",
 "country":"United States","country_code":"US","country_prefix":"1",
 "international_number":"+1 415-858-6273","local_number":"(415) 858-6273",
 "e164":"+14158586273","carrier":"AT&T Mobility"}
```

- **Why high pick:** Cheapest cleanup path ($6.99 → 5k); 249 countries; carrier+line_type+region in one shot.
- **Rate limit:** 429 on burst; documented but exact RPS not published.

---

## 3. NumVerify (apilayer)

- **Provider:** https://numverify.com / https://apilayer.com
- **Signup:** https://numverify.com/signup/free
- **Auth:** `access_key` query param (legacy `apilayer.net`) or `apikey` header (new `api.apilayer.com`)
- **Free tier:** 100 lookups/month
- **Paid:** Starter $9.99/mo for 2,000; Basic $14.99/mo for 5,000
- **Endpoint (legacy):** `GET http://apilayer.net/api/validate?access_key={KEY}&number={MSISDN}`
- **Endpoint (new):** `GET https://api.apilayer.com/number_verification/validate?number={MSISDN}` with header `apikey: YOUR_KEY`
- **Returns:** `valid`, `number`, `local_format`, `international_format`, `country_prefix`, `country_code`, `country_name`, `location`, `carrier`, `line_type`

```bash
curl "https://api.apilayer.com/number_verification/validate?number=14158586273" \
     -H "apikey: YOUR_KEY"
```
```json
{"valid":true,"number":"14158586273","local_format":"4158586273",
 "international_format":"+14158586273","country_prefix":"+1",
 "country_code":"US","country_name":"United States of America",
 "location":"Novato","carrier":"AT&T Mobility LLC","line_type":"mobile"}
```

- **Why high pick:** OG of free phone APIs. Still working in 2026 per status page; well-documented.
- **Caveat:** Free tier is HTTP-only on the legacy host; HTTPS requires a paid plan unless you go through `api.apilayer.com`.

---

## 4. numlookupapi.com

- **Provider:** https://numlookupapi.com
- **Signup:** https://app.numlookupapi.com/sign-up
- **Auth:** `apikey` header
- **Free tier:** 100 lookups/month, 10 req/min
- **Paid:** Small $9/mo for 7,000
- **Endpoint:** `GET https://api.numlookupapi.com/v1/validate/{phone}` (or query param `?phone_number=`)
- **Returns:** `valid`, `number`, `local_format`, `international_format`, `country_prefix`, `country_code`, `country_name`, `location`, `carrier`, `line_type` (landline/mobile/satellite/paging/special_services/premium_rate/toll_free)

```bash
curl "https://api.numlookupapi.com/v1/validate/+14158586273" \
     -H "apikey: YOUR_KEY"
```
```json
{"valid":true,"number":"+14158586273","local_format":"4158586273",
 "international_format":"+14158586273","country_prefix":"+1",
 "country_code":"US","country_name":"United States",
 "location":"California","carrier":"AT&T Mobility","line_type":"mobile"}
```

- **Why included:** Drop-in NumVerify clone; useful as a fallback when NumVerify quota is burned.

---

## 5. AbstractAPI Phone Validation

- **Provider:** https://www.abstractapi.com/api/phone-validation-api
- **Signup:** https://app.abstractapi.com/users/signup
- **Auth:** `api_key` query param
- **Free tier:** 100 requests/month, 3 req/sec, non-commercial only
- **Paid:** $19/mo (volume tiers up to 150k)
- **Endpoint:** `GET https://phonevalidation.abstractapi.com/v1/?api_key={KEY}&phone={MSISDN}`
- **Returns:** `valid`, `format` (`local`+`international`), `country` (`code`+`name`+`prefix`), `location`, `type` (line_type), `carrier`

```bash
curl "https://phonevalidation.abstractapi.com/v1/?api_key=YOUR_KEY&phone=14158586273"
```
```json
{"phone":"14158586273","valid":true,
 "format":{"international":"+14158586273","local":"(415) 858-6273"},
 "country":{"code":"US","name":"United States","prefix":"+1"},
 "location":"California","type":"mobile","carrier":"AT&T Mobility LLC"}
```

- **Note:** Each Abstract product is a separate signup — Phone Validation is its own dashboard.

---

## 6. Twilio Lookup v2

- **Provider:** https://www.twilio.com
- **Signup:** https://www.twilio.com/try-twilio
- **Auth:** HTTP Basic (Account SID + Auth Token, or API Key + Secret)
- **Free trial:** $15.50 credit on signup, no credit card required
- **Pricing:** Validation/format = free. `line_type_intelligence` = $0.008. `caller_name` (CNAM, US only) = $0.01. `identity_match` = $0.10. Many more packages (`sim_swap`, `call_forwarding`, `line_status`, `sms_pumping_risk`, `reassigned_number`).
- **Endpoint:** `GET https://lookups.twilio.com/v2/PhoneNumbers/{E164}?Fields=line_type_intelligence,caller_name`

```bash
curl "https://lookups.twilio.com/v2/PhoneNumbers/+14158586273?Fields=line_type_intelligence,caller_name" \
     -u "$TWILIO_SID:$TWILIO_TOKEN"
```
```json
{"calling_country_code":"1","country_code":"US","phone_number":"+14158586273",
 "national_format":"(415) 858-6273","valid":true,"validation_errors":null,
 "caller_name":{"caller_name":"JOHN DOE","caller_type":"CONSUMER","error_code":null},
 "line_type_intelligence":{"error_code":null,"mobile_country_code":"310",
   "mobile_network_code":"410","carrier_name":"AT&T Wireless",
   "type":"mobile"},
 "url":"https://lookups.twilio.com/v2/PhoneNumbers/+14158586273"}
```

- **Why top pick:** Highest-quality data on the list; CNAM is real-time and accurate. $15.50 trial = ~1,500 line_type calls before paying.
- **Reliability:** Industry standard, daily-updated docs.

---

## 7. Telnyx Number Lookup

- **Provider:** https://telnyx.com
- **Signup:** https://telnyx.com/sign-up
- **Auth:** Bearer token
- **Free trial:** $5 credit on signup
- **Pricing:** LRN $0.0015, MCC/MNC $0.0025, CNAM $0.003 per query (cheapest paid carrier+CNAM combo on this list)
- **Endpoint:** `GET https://api.telnyx.com/v2/number_lookup/{E164}?type=carrier,caller-name`
- **Returns:** `phone_number`, `national_format`, `country_code`, `carrier` (`name`, `type`, `mobile_country_code`, `mobile_network_code`, `normalized_carrier`), `portability` (`ported_status` Y/N, `ported_date`, `lrn`, `line_type`, `ocn`, `spid`), `caller_name` (`caller_name`, `caller_type`)

```bash
curl "https://api.telnyx.com/v2/number_lookup/+14158586273?type=carrier,caller-name" \
     -H "Authorization: Bearer $TELNYX_KEY"
```
```json
{"data":{"country_code":"US","fraud":null,
 "national_format":"(415) 858-6273","phone_number":"+14158586273",
 "portability":{"city":"NOVATO","line_type":"FIXED","lrn":"4158586000",
   "ocn":"9740","ported_date":null,"ported_status":"N","spid":"9740",
   "spid_carrier_name":"AT&T LOCAL","spid_carrier_type":"wireline",
   "state":"CA"},
 "carrier":{"error_code":"","mobile_country_code":"310",
   "mobile_network_code":"410","name":"AT&T Wireless","type":"mobile"},
 "caller_name":{"caller_name":"JOHN DOE","error_code":""}}}
```

- **Why top pick:** Cheapest serious provider — $5 trial buys you ~1,600 carrier+CNAM lookups.
- **Reliability:** Active, modern docs, growing competitor to Twilio.

---

## 8. SignalWire LookUp

- **Provider:** https://signalwire.com
- **Signup:** https://id.signalwire.com/onboarding (no card to start trial)
- **Auth:** HTTP Basic (`project_id:api_token`)
- **Trial:** Trial mode active until you fund the account with $5; you can hit the API in trial mode but rate-limited
- **Pricing:** $0.005 per carrier dip, $0.008 per CNAM dip
- **Endpoint:** `GET https://{SPACE}.signalwire.com/api/relay/rest/lookup/phone_number/{E164}?include=carrier,cnam`
- **Returns:** `valid_number`, `e164`, `country_code`, `country_code_iso3`, `national_number_formatted`, `location`, `timezones`, `carrier{lrn,spid,ocn,lata,city,state,linetype,name,type,mobile_country_code,mobile_network_code}`, `cnam{caller_id}`

```bash
curl --user "$PROJECT_ID:$API_TOKEN" \
  "https://my-space.signalwire.com/api/relay/rest/lookup/phone_number/+14158586273?include=carrier,cnam"
```
```json
{"valid_number":true,"e164":"+14158586273","country_code":"US",
 "national_number_formatted":"(415) 858-6273","location":"NOVATO, CA",
 "timezones":["America/Los_Angeles"],
 "carrier":{"name":"AT&T Wireless","type":"mobile","mobile_country_code":"310",
   "mobile_network_code":"410","lrn":"4158586000","spid":"9740",
   "ocn":"9740","lata":"722","city":"NOVATO","state":"CA","linetype":"mobile"},
 "cnam":{"caller_id":"JOHN DOE"}}
```

- **Why included:** Cheapest carrier dip ($0.005). Twilio-compatible HTTP basic auth makes Go integration trivial.

---

## 9. Vonage Number Insight (Basic / Standard / Advanced)

- **Provider:** https://www.vonage.com
- **Signup:** https://dashboard.nexmo.com/sign-up
- **Auth:** API Key + Secret as query params (or Basic header)
- **Free trial:** EUR 2 credit on signup (no card)
- **Pricing:** Basic free, Standard ~EUR 0.005, Advanced ~EUR 0.04 per lookup
- **SUNSET WARNING:** "Effective February 4, 2027, Vonage will sunset Number Insight." Migration is to Identity Insights API. Wire it up but don't make it a default.
- **Endpoints:** 
  - `GET https://api.nexmo.com/ni/basic/json?number={E164}&api_key=...&api_secret=...`
  - `GET https://api.nexmo.com/ni/standard/json?...` (adds carrier)
  - `GET https://api.nexmo.com/ni/advanced/json?...` (adds ported, roaming, reachable)

```bash
curl "https://api.nexmo.com/ni/advanced/json?api_key=$KEY&api_secret=$SECRET&number=447700900000"
```
```json
{"status":0,"status_message":"Success","international_format_number":"447700900000",
 "national_format_number":"07700 900000","country_code":"GB","country_name":"United Kingdom",
 "country_prefix":"44",
 "current_carrier":{"network_code":"23410","name":"O2","country":"GB","network_type":"mobile"},
 "original_carrier":{"network_code":"23415","name":"Vodafone","country":"GB","network_type":"mobile"},
 "ported":"ported","roaming":{"status":"not_roaming"},"valid_number":"valid","reachable":"reachable"}
```

- **Why include:** Only consumer-grade API on the list returning `roaming`, `reachable`, AND `ported` (with original carrier).

---

## 10. HLR-Lookups.com (vmgltd)

- **Provider:** https://www.hlr-lookups.com
- **Signup:** https://www.hlr-lookups.com/en/register
- **Auth:** Digest-Auth (`X-Digest-Key`, `X-Digest-Signature`, `X-Digest-Timestamp` HMAC-SHA256) or Basic-Auth
- **Free trial:** 100 free lookups on signup
- **Pricing:** ~EUR 0.01 per HLR lookup (varies by destination network)
- **Endpoint:** `POST https://www.hlr-lookups.com/api/v2/hlr-lookup` with body `{"msisdn":"+E164"}`
- **Returns:** `connectivity_status` (CONNECTED/ABSENT/INVALID_MSISDN/UNDETERMINED), `mccmnc`, `mcc`, `mnc`, `imsi`, `is_ported`, `ported_network_name`, `original_network_name`, `ported_network_prefix`, `roaming_network_code`, `cost` (EUR)

```bash
curl -X POST "https://www.hlr-lookups.com/api/v2/hlr-lookup" \
  -H "X-Digest-Key: $KEY" -H "X-Digest-Signature: $SIG" -H "X-Digest-Timestamp: $TS" \
  -d '{"msisdn":"+14156226819"}'
```
```json
{"id":"abc123","msisdnCountry":"US","msisdn":"+14156226819",
 "connectivityStatus":"CONNECTED","mccmnc":"310410","mcc":"310","mnc":"410",
 "originalNetworkName":"AT&T","originalCountryName":"United States",
 "isPorted":false,"isRoaming":false,"cost":0.01,"currency":"EUR"}
```

- **Why include:** Real HLR — actually pings the network. Tells you if the SIM is **alive** and roaming. None of the validation APIs above can do that.

---

## 11. HLRLookup.com (separate provider)

- **Provider:** https://www.hlrlookup.com
- **Signup:** https://www.hlrlookup.com/portal
- **Auth:** API key (header or query)
- **Free trial:** No public free credits; pay-as-you-go from GBP 0.0050 per lookup at 2,500 volume tier (down to GBP 0.0025 at 1M)
- **Endpoint:** Documented at https://www.hlrlookup.com/knowledge — same shape as HLR-Lookups (MSISDN in, MCC/MNC/ported out)
- **Returns:** Status, MCC/MNC, original/ported network, cost

- **Why include only as backup:** Slightly cheaper at scale than HLR-Lookups but no free credits — recommend HLR-Lookups.com first.

---

## 12. PhoneValidator.com

- **Provider:** https://www.phonevalidator.com
- **Signup:** https://api.phonevalidator.com/Join.aspx
- **Auth:** `apikey` query param
- **Free trial:** Test credits on signup (small bundle, exact count not published — historically ~25)
- **Pricing:** ~$0.062 per request (basic+detail+deactivation combined); per-feature pricing on detail page
- **Endpoint:** `GET https://api.phonevalidator.com/api/v2/phonesearch?apikey={KEY}&phone={MSISDN}&type=basic,detail,deactivation`
- **Returns:** `LineType` (CELL PHONE/LANDLINE/VOIP/TOLL-FREE/UNKNOWN), `PhoneCompany`, `PhoneLocation`, `ReportDate`, `Ported`, `LastDeactivation`, `Cost`, `ErrorCode`

```bash
curl "https://api.phonevalidator.com/api/v2/phonesearch?apikey=YOUR_KEY&phone=4158586273&type=basic,detail"
```
```json
{"PhoneBasic":{"PhoneNumber":"4158586273","LineType":"CELL PHONE",
   "PhoneCompany":"AT&T Mobility LLC","PhoneLocation":"NOVATO, CA"},
 "PhoneDetail":{"Ported":"true","LastPortDate":"2018-04-12","ReportDate":"2026-04-15"},
 "Cost":0.062}
```

- **Why include:** US-focused; one of the few APIs that returns `LastDeactivation` (when a number went out of service). Useful for skip-tracing.
- **Rate limit:** 25 req/sec.

---

## 13. OpenCNAM (Hobbyist)

- **Provider:** https://www.opencnam.com (now operated by Telo; the older neustar.com redirect is the dead Neustar B2B page)
- **Signup:** Not required for Hobbyist tier
- **Auth:** None (Hobbyist) / `account_sid`+`auth_token` Basic (Professional)
- **Free tier:** 60 *cached* CNAM lookups per hour, no signup, no API key
- **Pricing (Pro):** $0.004 per real-time successful lookup
- **Endpoint:** `GET https://api.opencnam.com/v3/phone/{E164}?format=pbx` (or `format=json`)
- **Returns (json):** `name`, `number`, `price`, `uri`

```bash
curl "https://api.opencnam.com/v3/phone/+14158586273?format=json"
```
```json
{"name":"JOHN DOE","number":"+14158586273","price":"0.0000","uri":"/v3/phone/+14158586273"}
```

- **Why include:** Zero-config, zero-signup CNAM. Clank can wire this in as the default "no API key needed" CNAM source.
- **Caveat:** Hobbyist returns *cached* data only — coverage thin for non-US numbers.

---

## 14. Free / Honourable Mention: FreeCnam

- **Provider:** https://freecnam.org
- **Signup:** Not required
- **Auth:** None
- **Endpoint:** `GET https://freecnam.org/dip?q={MSISDN}`
- **Returns:** Plain text caller name (NOT JSON)

```bash
curl "https://freecnam.org/dip?q=4158586273"
# JOHN DOE
```

- **Why include:** Truly free, US-only CNAM as a last-resort fallback when OpenCNAM is rate-limited. Ship it as menu option `4` paired with OpenCNAM.

---

## Skipped (don't fit clank's "5-min signup, individual dev" rule)

- **Pindrop VeriCall** — Enterprise fraud-tech sold via sales team. No public API key. Drop.
- **phone.email** — It's a *send-OTP* product, not a *lookup* product. Wrong shape entirely.
- **Truecaller** — Public API discontinued years ago; only available via partner SDK + business agreement.

---

## Comparison Table

| API | Free / mo | Carrier? | Type? | CallerName? | Country? | Fraud? | Ported? |
|---|---|---|---|---|---|---|---|
| IPQualityScore | 1,000 | Y | Y | N | Y | **Y (score)** | N |
| Veriphone | 1,000 | Y | Y | N | Y | N | N |
| NumVerify | 100 | Y | Y | N | Y | N | N |
| numlookupapi | 100 | Y | Y | N | Y | N | N |
| AbstractAPI | 100 | Y | Y | N | Y | risk_score | N |
| Twilio Lookup v2 | $15.50 trial (~1.5k) | Y (paid) | Y (paid) | **Y (US, paid)** | Y | sms_pumping_risk | reassigned_number |
| Telnyx Number Lookup | $5 trial (~1.6k) | Y | Y | **Y** | Y | N | **Y (date+spid)** |
| SignalWire LookUp | $0 trial (rate-limited) | Y | Y | Y | Y | N | N |
| Vonage NI Advanced | EUR 2 trial (~50) | Y | Y | N | Y | N | **Y + roaming** |
| HLR-Lookups.com | 100 lookups | Y | Y | N | Y | N | **Y + live status** |
| HLRLookup.com | 0 (paid only) | Y | Y | N | Y | N | Y |
| PhoneValidator.com | small trial | Y | Y | N | Y | N | Y + last-deactivation |
| OpenCNAM Hobbyist | 60/hr (no signup) | N | N | **Y (cached)** | N | N | N |
| FreeCnam | unlimited (no signup) | N | N | **Y (US)** | N | N | N |

**Default integrations recommended for clank:** IPQualityScore (rich + fraud score, 1k free), Veriphone (cheapest paid path + 1k free), Twilio Lookup v2 (gold-standard data, $15.50 trial). Layer OpenCNAM Hobbyist + FreeCnam as zero-signup CNAM fallbacks for the no-account-yet first-run experience.
