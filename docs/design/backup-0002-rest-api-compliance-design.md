# REST API Compliance (API v4) — Design Document (PROPOSED)

> **Status: proposed (2026-11-08) — nothing implemented.** Recommends moving the
> plugin's REST client (`pkg/rest`) from API v3 to API v4, fixing what it gets wrong
> against the `jdss-rest-api-spec` reference, and fixing the REST code's own bugs,
> crashes and security problems, together with the driver, controller and node code
> that misreads REST errors or logs secrets. It would be delivered and tested
> in one go, as one ordered series of commits; JovianDSS 2 support (token auth)
> would come in a later version, with its own design. How the REST client behaves
> (retries, failover, timeouts, logging) is out of scope and kept in the
> [Appendix](#appendix) for a later design ([REST client
> behaviour](#rest-client-behaviour)), except four of its items, which this design
> fixes because they are plain bugs: the empty error message and the noisy Warn
> logs (R9, R18, in A16), reply text used as a format string (R4, in A14 and A8),
> and a panic in `String()` (R5, in A8). Line numbers refer
> to commit
> [`11ffbfd`](https://github.com/open-e/joviandss-kubernetescsi/commit/11ffbfd376c47e0a95956b0cad8da79f399d0325),
> called "the current code" in the body.

## Table of Contents

- [List_of_Tables](#list-of-tables)
- [Problem_Statement](#problem-statement)
  - [API_versions](#api-versions)
- [Background_and_Research](#background-and-research)
  - [Evidence](#evidence)
  - [Evidence_Gaps](#evidence-gaps)
    - [Requests_the_plugin_sends_differently](#requests-the-plugin-sends-differently)
    - [Situations_not_recorded](#situations-not-recorded)
    - [Code_with_no_evidence](#code-with-no-evidence)
    - [What_the_evidence_covers_but_the_plugin_ignores](#what-the-evidence-covers-but-the-plugin-ignores)
    - [Results_of_the_gap_runs](#results-of-the-gap-runs)
    - [How_to_close_the_gaps](#how-to-close-the-gaps)
  - [Key_Observations](#key-observations)
- [Reference_Specification_and_Compliance](#reference-specification-and-compliance)
  - [The_jdss_rest_api_spec_repository](#the-jdss-rest-api-spec-repository)
  - [What_compliant_means](#what-compliant-means)
  - [Compliance_matrix](#compliance-matrix)
- [Project_Values](#project-values)
- [Options_Considered](#options-considered)
- [Recommended_Solution](#recommended-solution)
  - [First_commit_make_go_vet_pass_A14](#first-commit-make-go-vet-pass-a14)
  - [Before_stage_1_send_per_page_on_every_list_F25](#before-stage-1-send-per_page-on-every-list-f25)
  - [Before_stage_1_remove_dead_REST_code_A6_with_the_F14_switch](#before-stage-1-remove-dead-rest-code-a6-with-the-f14-switch)
  - [Stage_1_move_to_API_v4](#stage-1-move-to-api-v4)
  - [Stage_2_response_types](#stage-2-response-types)
  - [Stage_3_behaviour_and_error_mapping](#stage-3-behaviour-and-error-mapping)
  - [Stage_4_robustness_and_security](#stage-4-robustness-and-security)
  - [Supporting_work](#supporting-work)
- [Open_Questions](#open-questions)
- [Relationship_to_Other_Designs](#relationship-to-other-designs)
  - [Future_work_JovianDSS_2_support](#future-work-joviandss-2-support)
  - [Future_work_request_deadlines](#future-work-request-deadlines)
  - [Future_work_snapshot_name_limits](#future-work-snapshot-name-limits)
  - [Future_work_missing_fields_in_replies](#future-work-missing-fields-in-replies)
- [Risks_and_Backward_Compatibility](#risks--backward-compatibility)
  - [Preserved_low_risk](#preserved-low-risk)
  - [Risks](#risks)
- [Files_That_Would_Change_when_implemented](#files-that-would-change-when-implemented)
- [Testing](#testing)
- [Appendix](#appendix)
  - [API_mismatches](#api-mismatches)
    - [Response_types](#response-types)
    - [Response_content_and_defaults](#response-content-and-defaults)
    - [Error_messages](#error-messages)
  - [Logic_bugs](#logic-bugs)
  - [Driver_bugs](#driver-bugs)
  - [Security_findings](#security-findings)
  - [Crashes](#crashes)
  - [REST_client_behaviour](#rest-client-behaviour)
  - [Dead_code](#dead-code)
  - [ID_sources](#id-sources)
  - [Terms](#terms)

## List of Tables
[List_of_Tables](#list-of-tables)

1. `tbl_api_versions` — **API versions and auth**
2. `tbl_evidence_sources` — **Evidence sources**
3. `tbl_gaps_requests` — **Gaps: requests the plugin sends differently**
4. `tbl_gaps_situations` — **Gaps: situations not recorded**
5. `tbl_no_evidence` — **Code with no evidence**
6. `tbl_gap_results` — **Results of the gap runs**
7. `tbl_spec_forms` — **What the spec repo holds**
8. `tbl_spec_appliances` — **Appliances in the spec repo**
9. `tbl_compliance_rules` — **Compliance rules**
10. `tbl_compliance_matrix` — **Compliance matrix per call**
11. `tbl_compliance_status` — **Compliance status per rule**
12. `tbl_field_rules` — **Rule per field (A1)**
13. `tbl_files_changed` — **Files that would change**
14. `tbl_response_types` — **Response types that do not decode**
15. `tbl_response_content` — **Response content and defaults**
16. `tbl_error_messages` — **Error messages the plugin does not recognise**
17. `tbl_logic_bugs` — **Logic bugs in the REST code**
18. `tbl_driver_bugs` — **Driver bugs this design fixes**
19. `tbl_security` — **Security findings in the REST layer**
20. `tbl_crashes` — **Crashes in the REST layer**
21. `tbl_client_behaviour` — **REST client behaviour**
22. `tbl_id_sources` — **Where IDs come from**
23. `tbl_terms` — **Terms**

---

## Problem Statement
[Problem_Statement](#problem-statement)

**What's wrong.** There are three problems:

1. **The plugin was written for REST API v3 and has not kept up with it.** It talks
   to JovianDSS only through `api/v3/...` paths (22 different path patterns in
   `pkg/rest`), with HTTP Basic auth (`rproxy.go:123`). It turns every JSON reply
   into Go structs (`pkg/rest/types.go`, `payload-def.go`), and these structs are
   strict: if one field has a different JSON type than the struct expects, the
   whole reply is rejected and the CSI request fails with `Internal`. Over the years
   the appliance's replies changed. For example, volume `encryption` used to be a
   string and is now an object ([F1](#response-types)). The
   [API mismatches](#api-mismatches) part of the Appendix lists everything found:
   [replies that no longer decode](#response-types),
   [reply content the code reads wrongly](#response-content-and-defaults), and
   [error messages it no longer recognises](#error-messages). The last kind matters
   because the plugin uses the error message to decide whether a repeated request
   counts as success and whether a cleanup should go on. A message it does not
   recognise sends it down the wrong path.
2. **JovianDSS has released API v4, and the plugin has to support it.** JovianDSS 2
   offers only v4, and only with token login instead of Basic auth
   ([API versions](#api-versions)). Decided by the user: the plugin moves to v4 only,
   v3 is deprecated, and the oldest supported JovianDSS 1 release is
   `1.0up33.65410` ([Open questions 1 and 2](#open-questions)). This design moves
   the plugin to v4; JovianDSS 2 support, with token login, would come in a later
   version with its own design
   ([JovianDSS 2 support](#future-work-joviandss-2-support)).
3. **The REST code has bugs of its own**, whatever API version it talks to. They
   need to be found, reviewed and fixed: [logic bugs](#logic-bugs),
   [crashes](#crashes) in the REST layer, [security findings](#security-findings)
   and [dead code](#dead-code). The review also found problems in how the
   [REST client behaves](#rest-client-behaviour) (no retries, no failover, no
   request deadline); those wait for a later design, except the four that are
   plain bugs: the empty error message and the noisy Warn logs (R9, R18, fixed by
   A16), reply text used as a format string (R4, fixed by A14 and A8) and a panic
   in `String()` (R5, fixed by A8). Some of the bugs sit outside `pkg/rest`, in the
   code that calls it: [driver bugs](#driver-bugs) that misread REST errors, and
   log lines that print secrets (A15). This design fixes those too.

**Desired behaviour.** When this design is done:

- the plugin would use only `api/v4`;
- every REST call it makes would meet compliance rules SC1-SC6
  ([What "compliant" means](#what-compliant-means)). In plain words: the appliance
  answers the call, accepts what the plugin sends, the plugin decodes every v4
  reply the appliance has been seen to give, and every error it has been seen to
  give is handled on purpose: mapped to a known code, or left as Unknown by an
  explicit decision. Where the official docs and the appliances disagree (SC7),
  the design lists the difference and follows the appliance. For jdss2 the check
  uses the recordings only (A2, A4), because the plugin cannot log in to
  JovianDSS 2 until token login comes
  ([JovianDSS 2 support](#future-work-joviandss-2-support));
- csi-sanity would show no decode error (`cannot unmarshal`) in any preset;
- csi-sanity results on v4 would be the same as, or better than, before the move.
  The reference is cycle `20261102-140341`, whose build also carried the F3 fix
  ([Testing](#testing) step 4);
- every known appliance error would map to one deliberate code, and an unknown
  error would stay unknown, logged once in full (A16);
- no failure on the REST path would crash the plugin (A8);
- no password, CHAP secret or CSI `Secrets` would reach a log at any log level (A9,
  A15);
- the plugin would always use HTTPS, with TLS 1.3, and verify the appliance
  certificate unless the configuration says otherwise (A12).

### API versions
[API_versions](#api-versions)

**Table 1 — API versions and auth** · `tbl_api_versions`

| What | jdss1 | jdss2 | Plugin | Evidence |
|---|---|---|---|---|
| API v3 | served | **gone:** 401 without a token, 404 with one | uses only `api/v3` | hurl Finding 9 |
| API v4 | served | served | not used | hurl Finding 9 |
| Auth | Basic | **Bearer token only** (log in with `POST auth/login`; the access token lasts 900 s; each refresh token works once) | Basic only (`rproxy.go:123`) | hurl Findings 8, 9 |

On jdss1, the hurl suite compared v3 and v4 directly. In 363 list and delete cases
(hurl Finding 14) the two versions returned the same status, errno, class and
message in every case. In 281 rollback and clone cases (hurl Finding 15) they
did the same, apart from a few requests that got no answer at all and one count
that varied. Only the shape of list replies differs (Key observation 3). So the
move from v3 to v4 by itself should not change how JovianDSS 1 behaves.

---

## Background and Research
[Background_and_Research](#background-and-research)

What the design is based on: the evidence it uses, what that evidence does not
cover, and what follows from it.

### Evidence
[Evidence](#evidence)

**Table 2 — Evidence sources** · `tbl_evidence_sources`

| Source | What it is | Use |
|---|---|---|
| `jdss-rest-api-spec/empirical/hurl/out/` | every request and reply of the hurl runs on real appliances: 15 jdss1 and 20 jdss2 runs, over 15,900 replies to the calls the plugin decodes. Not in git | **main reference for reply types**, because it has the raw values. Only v4 replies decide pass or fail; v3 replies are background |
| `jdss-rest-api-spec/empirical/{jdss1,jdss2}/` | specs generated from those runs (`INDEX.md`, `-doc`, `-src`, TypeSpec) | reference for status codes, errors and behaviour. **Not for value types:** the values in the generated `GET pools/<pool>` replies are masked |
| `empirical/hurl/project-status.md` | the hurl suite's numbered findings (cited as "hurl Finding n") | behaviour, data safety |
| `jdss-rest-api-spec/{v4,v3}/d/` | the official (company) API docs | can differ from the appliance; used only where there are no recordings |
| `kub-testing/tmp/cycles/` | csi-sanity runs against the lab appliance | confirms an effect on a live system |

When sources disagree, recorded replies win, then the v4 docs, then the v3 docs.

**How the reply types were checked.** The analysis machine has no Go, so a Python
script imitates what Go's `encoding/json` does with each recorded reply (prototype
`kub-testing/tmp/rest-audit/audit.py`, output `audit.txt`). The real check in Go
comes with A4's unit tests. The [error messages](#error-messages) table and the
[evidence gaps](#evidence-gaps) were produced the same way (`coverage.py`,
`bodies.py`, same folder).

**What this check cannot show.** A field that was null in every recording, or a
list that was always empty, could not be checked (for example vdev spares, or a
pool `scan` in progress). Each hurl appliance is one installation, so a reply
shape that only appears with another pool layout or another feature switched on
would not be in the recordings. The lab appliance and hurl's `jdss1` run the same
JovianDSS 1 release, `1.0up33.65410` (confirmed by the user), and
[F1](#response-types) and [F8](#response-types) fail the same way on both. Those
two are the only reply-type findings confirmed on the live system so far (the
"Seen live?" column of Table 14).

### Evidence Gaps
[Evidence_Gaps](#evidence-gaps)

hurl recorded replies for every request the plugin sends, on jdss1 v4 and on jdss2
v4 (see the [compliance matrix](#compliance-matrix)). Two things are missing. hurl
did not always send a request the exact way the plugin does (an extra body, no
`per_page`, longer names; [Table 3](#requests-the-plugin-sends-differently)), and
some situations never came up in the tests at all (a full pool, a missing pool, an
appliance that does not answer; [Table 4](#situations-not-recorded)).

The two tables describe the state **before** the gap runs, two extra hurl runs made
to fill these gaps. What those runs found is in
[Results of the gap runs](#results-of-the-gap-runs): where that table has a row for
a gap (G1-G8, G11), it supersedes the row here.

#### Requests the plugin sends differently
[Requests_the_plugin_sends_differently](#requests-the-plugin-sends-differently)

**Table 3 — Gaps: requests the plugin sends differently** · `tbl_gaps_requests`

| ID | The plugin sends | hurl sent | Why it matters |
|---|---|---|---|
| G1 | a `{}` body on **every GET**, and on DELETE of target, LUN and share (`rproxy.go:108`) | GET: never a body. DELETE of those three: never a body | works on jdss1 v3 (the lab today); unknown on v4 and on jdss2 |
| G2 | lists as `?page=N&_dc=<random>`, without `per_page` | never `_dc`. `page` without `per_page` only on snapshot lists, with at most 10 items | the default page size is unknown. The plugin keeps asking for the next page until it gets an empty one (`driver/common.go:52`). That stop rule was seen working, but never on a long list without `per_page`. |
| G3 | `POST volumes` with only `name` and `size` | always with `sparse` | every CSI volume gets the appliance's default provisioning, and defaults can change between releases: block size did (hurl Finding 9) |
| G4 | names up to 243 characters (`vp_`/`vb_` + 240) | longest name about 59 characters | the length limit and its error are unknown |

#### Situations not recorded
[Situations_not_recorded](#situations-not-recorded)

**Table 4 — Gaps: situations not recorded** · `tbl_gaps_situations`

| ID | Situation | Plugin code that depends on it |
|---|---|---|
| G5 | **Out of space:** a volume, thick volume, snapshot or NAS reservation larger than the pool. The error message was never seen | `errors.go:270`; `ResourceExhausted` at `controllerplugin.go:378`, `:656` |
| G6 | Zvol `DeleteClone` of a clone that does not exist (the NAS case is recorded) | `iscsi_driver.go:520` |
| G7 | `GET .../clones` on a missing snapshot or volume, zvol and NAS: only 200 recorded | `iscsi_driver.go:498`, `nfs_driver.go:635` |
| G8 | The pool is missing or unavailable, on any call except `GET pools/<pool>` | every call |
| G9 | **Clients connected:** in the hurl tests no initiator logs in and nothing is mounted (hurl README). Detach, target delete and deletes with `force_umount: true` were recorded only on idle resources ([risk 8](#risks)) | unpublish and every delete |
| G10 | **The appliance cannot be reached:** connection refused, DNS, timeout, TLS, reply cut off. Not recorded on purpose, but it happens: jdss1 stopped answering for 15-18 minutes three times (hurl C24), and once for over 8 hours. On this path the plugin crashes ([C3](#crashes)) | `Send` |
| G11 | Unexpected reply bodies: HTML or empty 5xx replies, or a list answered with `200` and `data: null`, on which the plugin crashes ([C5](#crashes)) | `getError`, list functions |
| G12 | Failover: the pool moves to the other node during a call; several addresses in `addrs` ([R1](#rest-client-behaviour)) | `Send` |

#### Code with no evidence
[Code_with_no_evidence](#code-with-no-evidence)

Code that handles replies no recording has ever shown.

**Table 5 — Code with no evidence** · `tbl_no_evidence`

| Code | Note |
|---|---|
| "already exists" with errno 1 or errno 100; "not found" without an errno | the recordings show "already exists" with errno 5 and "not found" with errno 1. These branches may be for older releases; keep them in A16's catalogue as legacy entries without a recording, and cover them with unit tests |
| ~~`RestErrorOutOfSpace`~~ | **now seen:** `400 New zvol size(...) exceeds available space` on jdss1 v3 and jdss2 v4, and recognised. The other out-of-space messages are [F26](#error-messages) |
| `targetDneMsgPattern` | can never match: [F15](#logic-bugs) |
| status 202 in `CreateTarget`, `GetVolumeSnapshot`, `GetNASVolumeSnapshot`; 204 on GET calls; 200 on `DeleteVolume` | never seen, in about 21,000 requests per run. The 202 branch in `CreateTarget` is broken ([F24](#logic-bugs)), and so is the 204 branch on GETs ([F23](#logic-bugs)). The 202 on the snapshot GETs and the 200 on `DeleteVolume` are harmless and would stay (SC3) |

#### What the evidence covers but the plugin ignores
[What_the_evidence_covers_but_the_plugin_ignores](#what-the-evidence-covers-but-the-plugin-ignores)

On jdss2, a snapshot list is briefly out of date after a snapshot is created or
deleted: for under 0.5 s (hurl Finding 20). So a ListSnapshots straight after a
CreateSnapshot may not show the new snapshot. Reading a snapshot by its name works
at once. This design proposes no change for it (see stage 3); what later driver
work would have to allow for is in [risk 14](#risks).

#### Results of the gap runs
[Results_of_the_gap_runs](#results-of-the-gap-runs)

Two runs were made to fill the gaps: jdss1 `20261105-221223` on v3, and jdss2
`20261105-220744` on v4. **There are no jdss1 v4 results:** jdss1 stopped answering
2 minutes after its v4 part started (the 8-hour outage in G10). One test, which
puts a newline into a share comment, broke share handling and NAS deletes on both
appliances for the rest of the run (hurl Finding 21 for jdss2, 22 for jdss1),
and on jdss1 the pool also filled up. So the results below come only from the
part of each run before that test, or from zvol tests, which the breakage did
not touch. The `scale/` tests and `nospace/volume-no-space.hurl` ran on neither
appliance.

**Table 6 — Results of the gap runs** · `tbl_gap_results`

| ID | Result | Finding |
|---|---|---|
| G1 | No difference with and without the `{}` body, in 80 (jdss1 v3) and 82 (jdss2 v4) pairs of requests, including DELETE of target, LUN and share | closed, except on jdss1 v4 |
| G2 | On jdss2 v4, the `volumes` and `nas-volumes` lists ignore `page` when `per_page` is missing; on jdss1 v3 they return an empty page. With `per_page`, paging works on both. `_dc` changes nothing. The default page size is still unknown | [F25](#response-content-and-defaults): **must be fixed before the v4 move** |
| G3 | The default is **thick** (the full size is reserved) on both: the same `refreservation` as with `sparse: false`, about the volume size | [F28](#response-content-and-defaults) |
| G4 | Zvol and clone names: 129 characters accepted, 150 refused; the appliance's own message gives the limit, 132. Snapshots: the limit is on the full name `pool/volume@snapshot`, at most 255; with the tested pool and volume names it was reached at 217 ([snapshot name limits](#future-work-snapshot-name-limits)) | [F27](#error-messages); A11 would cap names at 131 and keep the plugin's generated names within it |
| G5 | Three different out-of-space messages; only one is recognised. The zvol snapshot, clone and thick-volume cases were not run | [F26](#error-messages) |
| G6 | `500`, errno 1, `Zfs resource: ... not found`: recognised | [F27](#error-messages) (the problem is in the caller); low |
| G7 | Clones of a missing parent: `500`, errno 1, recognised. A snapshot with no clones: `200` with an empty list, never `204`. Snapshot list of a missing parent: `500` on jdss1 v3, `404` on jdss2 v4 | [F27](#error-messages); NFS only |
| G8 | `POST` and list of volumes on a missing pool: `500`, errno 1. Targets: `404` | [F27](#error-messages); low |
| G11 | No `200` with `data: null` on any GET. Plain-text `404 not found` and `401 unauthorized` replies exist; they cause a parse error, not a crash. The crash on `null` is still in the code, but no recorded reply reaches it | [C5](#crashes), [R14](#logic-bugs); unit test only |

Never seen in either run (about 21,000 requests each): `202`, `204` on a GET, `200`
on a DELETE. Also never seen, not even in the `race/` tests:
`Resource ... already exists.` with errno 1 or 100, and "not found" without an
errno. Creating the same object twice at the same time gives errno 100
`dataset already exists` (recognised) or errno 5.

#### How to close the gaps
[How_to_close_the_gaps](#how-to-close-the-gaps)

The spec repo is its own project, so this design does not change it. The user adds
tests for G1-G8 to the hurl suite; the results so far are
above. G9 and G12 need a connected client or a second node. csi-sanity's node preset
would cover part of G9 on the lab appliance, but only if the iSCSI NodeStage really
logs in, which review findings N13-N15 put in doubt; [Testing](#testing) steps 6 and
7 check this. G10 and G11 cannot be recorded
reliably, so they would be covered by unit tests with a fake HTTP server (A4).

### Key Observations
[Key_Observations](#key-observations)

1. **v3 ends with JovianDSS 1.** JovianDSS 2 has only v4, and the Proxmox plugin
   already uses v4 by default. That is why v4 is the only target.
2. **One set of type fixes would cover every combination tested.** The same fields fail
   on jdss1 v3, jdss1 v4 and jdss2 v4, so the v4 move and the type fixes do not
   depend on each other.
3. **v3 and v4 return lists in different shapes.** On v3, the
   `volumes` and `nas-volumes` lists return a plain array when no paging is asked
   for, and `data.results`/`entries` when it is. On v4 they always return
   `results`/`entries`, which is simpler. The plugin always asks for paging
   (`pagedcSuffix`), so it always gets `entries`. The one function that expects the
   plain array, `ListVolumes` (`storageapi.go:165`), has no callers and would break
   on v4; it would be removed with the other [dead code](#dead-code).
4. **Most fields that fail are never used.** Of the fields in
   [response types](#response-types), the plugin reads only [F9](#response-types).
   It decodes far more than it uses, and every unused field is one more way to fail.
5. **The expected-status argument of `Send` is ignored** (`rproxy.go:84`). `Send`
   takes an argument that says which HTTP status counts as success, but never looks
   at it; each caller checks the status itself. So when a delete passes the create
   constant by mistake, `CreateCloneRCode` (hurl Finding 4.4), nothing goes wrong:
   `DeleteClone` checks for 204 on its own.
6. **Error handling depends on message text** (`errors.go`). If the appliance
   rewords a message, the plugin's behaviour changes and nothing reports an error.
   A4's fixtures would catch this when new hurl runs are added, and A16 would keep
   all message patterns in one catalogue, so callers decide on codes, not on text.

---

## Reference Specification and Compliance
[Reference_Specification_and_Compliance](#reference-specification-and-compliance)

### The `jdss-rest-api-spec` repository
[The_jdss_rest_api_spec_repository](#the-jdss-rest-api-spec-repository)

`jdss-rest-api-spec` (local copy at `/home/user/repos/jdss-rest-api-spec`; not on
kub-gate) is what this design measures the plugin against. It describes the
JovianDSS REST API twice, with the same paths in both: once as Open-E documents it
(the company spec), and once as real appliances behave (the empirical spec). It was
written for the JovianDSS-Proxmox and CSI plugin teams.

**Table 7 — What the spec repo holds** · `tbl_spec_forms`

| Part | Where | What it says | Role here |
|---|---|---|---|
| **Company spec** | `v4/d/<path>/<METHOD>-doc` (description), `-src` (JSON Schema of the request body), `-desc`; also `v4/typespec/`, `v4/openapi3/`. `v3/` is laid out the same way | what Open-E documents, taken from the Swagger HTML shipped with the appliance (`v4/origin/`) | which request keys exist, documented status codes; fills gaps |
| **Empirical spec** | `empirical/<appliance>/v4/`: `INDEX.md` (every tested endpoint, with documented and observed status codes), `STATUS-CODES.md`, `SRC-COMPARISON.md`, `ERROR-FIELDS.md`, `d/<path>/<METHOD>-doc`/`-src` (as observed), `typespec/` (namespace `Observed`); `empirical/<appliance>/status-codes.json` | what real appliances did, generated from the hurl runs | **decides where the two specs disagree** |
| Raw responses | `empirical/hurl/out/` (local only, not in git) | every request and reply of every run | value types (the generated `GET pools/<pool>` values are masked) |
| Findings | `empirical/hurl/project-status.md`, `bug-reports/` | numbered findings, runs, repeatability | behaviour, data safety |

The appliances, as the repo's README describes them:

**Table 8 — Appliances in the spec repo** · `tbl_spec_appliances`

| Name | Release | Auth | APIs |
|---|---|---|---|
| `jdss1` | JovianDSS 1 **`1.0up33.65410`** (the plugin's minimum release), two-node HA cluster | Basic | v3, v4 |
| `jdss2` | JovianDSS 2 `2.0up01.67856` | JWT Bearer token, from `POST api/v4/auth/login`, valid 900 s | v4 only |

**Rules for this work.** The spec repo belongs to its own project. This design and
the plugin work only read it: they never change its code or data, and never run its
scripts or tests (the tests create and delete objects on real appliances). Problems
found in it are reported to the user. The repo covers only the requests the Proxmox
and CSI plugins make: an endpoint missing from it was simply not tested, which does
not mean it is broken.

### What "compliant" means
[What_compliant_means](#what-compliant-means)

The plugin is compliant when every REST call it makes meets the rules below on
**both** jdss1 v4 and jdss2 v4. Where the company spec and the empirical spec
disagree, the empirical spec wins (Project value 3).

**Table 9 — Compliance rules** · `tbl_compliance_rules`

| Rule | Requirement | Checked against |
|---|---|---|
| SC1 | Path and method exist: the appliance has answered this call with a success (2xx) status | `empirical/<app>/v4/INDEX.md` |
| SC2 | The request body is valid: every key the plugin can send is in the company schema (`<METHOD>-src`), every required key is sent, and the appliance accepted the body in the recorded runs | `v4/d/<path>/<METHOD>-src`, `empirical/<app>/v4/SRC-COMPARISON.md` |
| SC3 | Success statuses: the plugin accepts every 2xx status the appliance returned for the call, and a status never returned cannot cause wrong behaviour | `INDEX.md` observed codes, `status-codes.json` |
| SC4 | Error statuses: every non-2xx status returned for the call is handled on purpose, either mapped to a named `RestError` code or left as Unknown by an explicit decision | `INDEX.md`, `STATUS-CODES.md`, recorded error bodies |
| SC5 | Reply types: every Go type the plugin decodes accepts every recorded v4 reply of that call | the raw replies in `hurl/out/` (A2). Not the generated TypeSpec: checking against it needs the spec repo's tools, which this work does not run |
| SC6 | Error bodies: errors with missing or varying fields (no `errno`, no `class`) are handled | `ERROR-FIELDS.md` |
| SC7 | Docs and reality: where the company spec and the appliance differ, the plugin follows the appliance, and the difference is listed here so it can be reported to the spec project | company `-doc` against empirical `-doc` |

The rules would be checked again after each new hurl run that the spec project
adds to its generated specs (A2).

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

One row per REST call. "v4 doc" is the statuses the company spec lists; "jdss1 v4"
and "jdss2 v4" are the statuses the appliances returned (from both appliances'
`INDEX.md`; prototype `kub-testing/tmp/rest-audit/compliance_matrix.py`); "Plugin
success" is the statuses the code treats as success. All 34 calls the plugin makes
today appear in both `INDEX.md` files, and so does the one call stage 3 would add
([F12](#logic-bugs)), which has its own row. "Note" links the findings that affect
the call; findings that affect every call, such as [R14](#logic-bugs), are not
repeated.

**Table 10 — Compliance matrix per call** · `tbl_compliance_matrix`

| Plugin call | v4 method and path | v4 doc | jdss1 v4 | jdss2 v4 | Plugin success | Note |
|---|---|---|---|---|---|---|
| `GetPool` | `GET pools/<x>` | 200 | 200, 401, 404 | 200, 401, 404 | 200, 204 | [F8](#response-types), [F8b](#response-types), [R10](#logic-bugs), [F23](#logic-bugs) |
| `GetVolume` | `GET pools/<x>/volumes/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 204 | [F1](#response-types), [F23](#logic-bugs) |
| `CreateVolume` | `POST pools/<x>/volumes` | 201 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [F26](#error-messages), [F28](#response-content-and-defaults) |
| `DeleteVolume` | `DELETE pools/<x>/volumes/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 200, 204 |  |
| `GetVolumesEntries` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200, 201 | [F1](#response-types), [F25](#response-content-and-defaults) |
| `ListVolumes` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200 | dead, removed |
| `GetVolumeSnapshot` | `GET pools/<x>/volumes/<x>/snapshots/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 202 | [R16](#response-content-and-defaults) |
| `CreateSnapshot` | `POST pools/<x>/volumes/<x>/snapshots` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 | [F18](#error-messages) |
| `DeleteSnapshot` | `DELETE pools/<x>/volumes/<x>/snapshots/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 204 |  |
| `GetSnapshotsEntries` | `GET pools/<x>/volumes/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [F9](#response-types), [F11](#response-content-and-defaults), [R10](#logic-bugs) |
| `GetVolumeSnapshotsEntries` (new call in stage 3) | `GET pools/<x>/volumes/<x>/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | none today (the function is a stub); 200 once [F12](#logic-bugs) implements it | [F12](#logic-bugs); fixtures from `list/volume-kinds.hurl` (A4) |
| `GetVolumeSnapshotClones` | `GET pools/<x>/volumes/<x>/snapshots/<x>/clones` | 200 | 200, 500 | 200, 500 | 200, 204 | [F10](#response-types), [F23](#logic-bugs) |
| `CreateClone` | `POST pools/<x>/volumes/<x>/clone` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 |  |
| `DeleteClone` | `DELETE pools/<x>/volumes/<x>/snapshots/<x>/clones/<x>` | 204 | 204, 500 | 204, 500 | 204 |  |
| `GetTarget` | `GET pools/<x>/san/iscsi/targets/<x>` | 200 | 200, 404 | 200, 404 | 200 | [F15](#logic-bugs) |
| `CreateTarget` | `POST pools/<x>/san/iscsi/targets` | 201 | 201, 400, 404, 409, 500 | 201, 400, 404, 409, 500 | 200, 201, 202 | [F24](#logic-bugs) |
| `DeleteTarget` | `DELETE pools/<x>/san/iscsi/targets/<x>` | 204 | 204, 404 | 204, 404 | 204; 404 = gone |  |
| `AttachVolumeToTarget` | `POST pools/<x>/san/iscsi/targets/<x>/luns` | 200 | 201, 400, 404, 409 | 201, 400, 404, 409 | 200, 201 | SC7: doc 200, appliance 201 |
| `DettachVolumeFromTarget` | `DELETE pools/<x>/san/iscsi/targets/<x>/luns/<x>` | 204 | 204, 404 | 204, 404 | 204; 404 = target gone | [F19](#error-messages) |
| `GetNASVolume` | `GET pools/<x>/nas-volumes/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 204 | [F23](#logic-bugs), [R15](#response-types) |
| `CreateNASVolume` | `POST pools/<x>/nas-volumes` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [F26](#error-messages) |
| `DeleteNASVolume` | `DELETE pools/<x>/nas-volumes/<x>` | 204, 400, 500 | 204, 400, 404, 500 | 204, 400, 404, 500 | 204 | [F20](#error-messages) |
| `GetNASVolumesEntries` | `GET pools/<x>/nas-volumes` | 200, 500 | 200, 500 | 200, 500 | 200, 201 | [F25](#response-content-and-defaults) |
| `GetNASVolumeSnapshotsEntries` | `GET pools/<x>/nas-volumes/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [F13](#logic-bugs): wrong path, reads the volume |
| `GetNASSnapshotsEntries` | `GET pools/<x>/nas-volumes/snapshots` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [F16](#response-content-and-defaults) |
| `CreateNASSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [F18](#error-messages), [F26](#error-messages) |
| `DeleteNASSnapshot` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>` | 204, 400, 500 | 204, 404, 500 | 204, 404, 500 | 204 |  |
| `DeleteNASVolumeSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>` | **method missing** | 405 | 405 | 204 | **[F14](#logic-bugs): wrong method**, [R7](#logic-bugs); removed |
| `CreateNASVolumeSnapshot` | `POST pools/<x>/nas-volumes` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | dead ([R11](#dead-code)), removed |
| `GetNASVolumeSnapshot` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 202 |  |
| `GetNASVolumeSnapshotClones` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 500 | 200, 500 | 200, 500 | 200, 204 | [F23](#logic-bugs) |
| `CreateNASClone` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 400, 500 | 200, 400, 500 | 200, 400, 500 | 200, 201 |  |
| `DeleteNASClone` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>/clones/<x>` | 204, 400, 404, 500 | 204, 404, 500 | 204, 404, 500 | 204 | [F22](#error-messages) |
| `CreateShare` | `POST shares` | 201 | 201, 400, 409, 500 | 201, 400, 409, 500 | 200, 201 | [F21](#error-messages) |
| `DeleteShare` | `DELETE shares/<x>` | 204 | 204, 404 | 204, 404, 500 | 204 | dead ([R12](#dead-code)), removed; [F17](#error-messages) |

Where the plugin stands on each rule today:

**Table 11 — Compliance status per rule** · `tbl_compliance_status`

| Rule | Status in the current code | Fixed by |
|---|---|---|
| SC1 | **33 of 34 pass.** Fails: [F14](#logic-bugs). [F13](#logic-bugs) passes on paper, but calls the wrong path | dead-code commit ([F14](#logic-bugs)), stage 3 ([F13](#logic-bugs)) |
| SC2 | **Passes for every key the driver really sets.** Three optional keys in the Go request structs are not in the company schema, but they are never set ([dead code](#dead-code)). `properties` is an open object in the schema, so any sub-key is allowed. The `{}` body sent on GET and DELETE is outside every schema; G1 found it has no effect. Prototype: `kub-testing/tmp/rest-audit/compliance_bodies.py` | dead-code commit |
| SC3 | **Every 2xx status returned is accepted.** The code also accepts statuses that were never returned: [F23](#logic-bugs), [F24](#logic-bugs), 202 on both snapshot GETs, 200 on `DeleteVolume`, 201 on list GETs, and 200 on creates that answer 201 | stage 3 ([F23](#logic-bugs), [F24](#logic-bugs)); the rest are harmless and would stay |
| SC4 | **Fails.** [R14](#logic-bugs); `401` on `GetPool` is not mapped; [F15](#logic-bugs), [F18](#error-messages), [F19](#error-messages), [F20](#error-messages), [F21](#error-messages), [F22](#error-messages), [F26](#error-messages), [F27](#error-messages) | stage 3 (with A16) |
| SC5 | **Fails.** [F1](#response-types), [F8](#response-types), [F8b](#response-types), [F9](#response-types), [F10](#response-types). ([R15](#response-types) decodes without an error, so it does not fail SC5) | stage 2 |
| SC6 | **Fails on jdss2:** [F17](#error-messages). Only `DeleteShare` runs into it, and the dead-code commit would remove `DeleteShare`, so SC6 would pass from that commit on, unless a share delete is added back | dead-code commit |
| SC7 | Differences found so far: `AttachVolumeToTarget` is documented as 200, the appliance returns 201; the [F8](#response-types) and [F8b](#response-types) fields are documented as strings; [F1](#response-types)'s v4 example shows `"off"`; for most zvol calls only the success status is documented, while the appliance also returns 400, 404 and 500; the `POST` used by [F14](#logic-bugs) is not documented and returns 405; target names are documented as up to 255 characters, but 225 is refused (220 accepted); target names are documented as lower case, but upper case is accepted; the name rules for zvols, NAS datasets and snapshots are not documented at all (the company schemas say only `"type": "string"`). The plugin already follows the appliance, except where a finding says otherwise | a list for the spec project ([Relationship](#relationship-to-other-designs)) |

---

## Project Values
[Project_Values](#project-values)

Proposed, most important first; please correct:

1. **API v4 only:** no fallback to v3 and no setting to choose (user decision).
2. **Works with the latest JovianDSS:** the plugin targets the latest release; the
   minimum JovianDSS 1 release is in [Open question 2](#open-questions).
3. **Evidence over docs:** types would follow recorded replies; the docs would only
   fill gaps.
4. **Bugs stay visible:** a different format would be tolerated only where it is
   known to differ. An unknown change would still fail, with a clear error that names
   the field.
5. **Right first time:** the whole design would be delivered and tested in one go.
   Every check that does not need the lab (build, unit tests, the audit) would pass
   before the live tests, so that no follow-up fixes are needed. The REST
   layer's structure would stay; its error handling would be reworked (A16) and its
   transport tightened (stage 4), but it would not be rewritten.
6. **Repeatable checks:** the plugin's audit (A2) could be re-run whenever the spec
   project adds new hurl recordings, so a change in the appliances' replies would
   show up as a failed check. The rule would also be written into the package
   documentation (A7), so whoever changes `pkg/rest` later knows it.
7. **Errors mean what they say:** every known appliance error would have one
   deliberate meaning; an unknown error would stay unknown, be logged in full in one
   place, and never be guessed into another meaning. Misreading an error, such as
   taking "not found" for "busy", is treated as a bug, not as a fallback.

---

## Options Considered
[Options_Considered](#options-considered)

- **Fix only [F1](#response-types) and [F8](#response-types)**, the two failures seen
  live: would leave [F8b](#response-types), [F9](#response-types) and
  [F10](#response-types), which would fail the same way as soon as they are reached;
  against value 5.
- **Make every mismatching type tolerant** instead of removing unused fields: more
  code for data nobody uses; against value 4.
- **Lenient decoding** (ignore all type errors): would hide real changes in fields
  the plugin does use; against value 4.
- **`json.RawMessage` for the odd fields:** tolerant, but would keep fields nobody
  reads.
- **Generate the client from hurl's TypeSpec or OpenAPI:** attractive later, but the
  format is not agreed yet (it is the hurl suite's own next step), it describes
  only what was seen, and
  it would mean rewriting the REST layer; against value 5.
- **Keep v3, or support both with a setting:** rejected (user decision).
  v3 is deprecated, and two APIs would double the testing for no gain.
- **Our own capture script:** not needed; hurl already records the replies, covers
  more, and needs no plugin credentials.

---

## Recommended Solution
[Recommended_Solution](#recommended-solution)

The design would be delivered in one go: one series of commits, built, checked and
tested together ([Testing](#testing)). The commits would come in this order: a
`go vet` clean-up (A14), `per_page`, dead-code removal, the move to v4, response
types (stage 2), behaviour and error mapping (stage 3), robustness and security
(stage 4). Each commit would
build on its own, so if the live test failed, `git bisect` could find the commit
that caused it. JovianDSS 2 support (token auth) is not part of this delivery: it
would come in a later plugin version, with its own design ([JovianDSS 2
support](#future-work-joviandss-2-support)).

### First commit: make `go vet` pass (A14)
[First_commit_make_go_vet_pass_A14](#first-commit-make-go-vet-pass-a14)

**Its own commit, first in the series.** [Testing](#testing) step 1 runs `go vet`,
Go's built-in static checker, on every commit, so vet has to pass before the series
starts. Since Go 1.24, vet's
printf check reports a format string that is not a constant and has no arguments;
the module is Go 1.26, so the check applies. The current code has six such calls,
all with reply text as the format: `fmt.Sprintf(string(body[:]))`
(`storageapi.go:51`), `l.Warnf(msg)` (`storageapi.go:53`, `errors.go:173`, `:265`)
and `l.Debugf(msg)` (`targets.go:150`, `:184`). **A14** turns them into `"%s"`
formats, which is also the fix for [R4](#rest-client-behaviour)'s format-string
part: a `%` in an appliance message can no longer garble the message or the log.
The same commit would add `vet` and `test` targets to the `Makefile`, so that
[Testing](#testing) step 1 is one command from the first commit on (`make test`
has nothing to run until A4). Nothing else would change in this commit. No vet
result has been recorded for the current code; the first step of the work is to
run it and record what else it reports, if anything.

### Before stage 1: send `per_page` on every list (F25)
[Before_stage_1_send_per_page_on_every_list_F25](#before-stage-1-send-per_page-on-every-list-f25)

**Its own commit, before the v4 move.** Without it, the v4 move could make CSI
ListVolumes run forever ([F25](#response-content-and-defaults)): on jdss2 v4 a
list request without `per_page` ignores the page number and returns the first
page again and again. On jdss1 v4 this was never recorded (gap G2), so it may
behave the same way. Paging with `per_page` works on both jdss1 v3 and jdss2 v4.
The stop rule would change too. Today the plugin asks for page after page until
a page comes back empty; it would also stop once it has received as many entries
as the reply's `results` count announces, so a server that repeats a page could
not keep it looping.

### Before stage 1: remove dead REST code (A6), with the F14 switch
[Before_stage_1_remove_dead_REST_code_A6_with_the_F14_switch](#before-stage-1-remove-dead-rest-code-a6-with-the-f14-switch)

**Its own commit, after the `per_page` commit and before the v4 move,** so that the
v4 commit would only touch code that is used. It would remove everything listed
under [dead code](#dead-code).

One of the functions to remove, `DeleteNASVolumeSnapshot`, is a broken copy of
`DeleteNASSnapshot` (`nas_snapshots.go:67`): both delete one NAS snapshot at the
same path, but the copy sends `POST` instead of `DELETE`. Its one caller is the
cleanup of an intermediate snapshot at `nfs_driver.go:322`. This commit would point
that call at `DeleteNASSnapshot`, which the normal snapshot delete already uses
(`nfs_driver.go:193`, `:623`), so deleting NAS snapshots stays supported, through
one working function ([Open question 10](#open-questions)). The
build would keep working, and **[F14](#logic-bugs) and [R7](#logic-bugs) would be
fixed here** without new code. This commit only makes that call correct; the
cleanup around it would still not run properly until stage 3 fixes review findings
[D12 and D13](#driver-bugs).

Dead code outside `pkg/rest`, such as the old node implementation in
`pkg/node/targetapi.go` and `target.go`, is not part of this commit.

### Stage 1: move to API v4
[Stage_1_move_to_API_v4](#stage-1-move-to-api-v4)

One API prefix, set to `api/v4`, would replace the hard-coded `api/v3` strings: 34
in the current code (plus 3 in comments), 30 after the dead-code commit. There would
be no setting and no fallback to v3. **Nothing else would change in this commit**,
so if the live test failed, `git bisect` could tell an effect of v4 from an effect
of a fix. No change in behaviour is expected on JovianDSS 1 (hurl Findings 14, 15).
Of the evidence gaps, G1 (the empty `{}` body the plugin sends with every GET) made
no difference on jdss1 v3 and jdss2 v4, and the live test would be its first check
on jdss1 v4; G2 and G3 became findings of their own,
[F25](#response-content-and-defaults) (paging) and
[F28](#response-content-and-defaults) (thick volumes), handled elsewhere.

**Deprecating v3:** the plugin docs would say it needs a JovianDSS release with API
v4; `CHANGELOG.md` would name the oldest supported JovianDSS 1 release,
`1.0up33.65410`; the audit and the fixtures would use v4 replies; comments and
constants that mention v3 would be removed.

### Stage 2: response types
[Stage_2_response_types](#stage-2-response-types)

Would fix the [response types](#response-types) that do not decode, and
[F11](#response-content-and-defaults), which has to land together with
[F9](#response-types) (see the table).

**A1. Rule per field**

**Table 12 — Rule per field (A1)** · `tbl_field_rules`

| Case | Rule |
|---|---|
| field not read, appliance sends another type | **remove it from the struct**: `encoding/json` then ignores the key, whatever its type. Applies to [F1](#response-types), [F8](#response-types), [F8b](#response-types), [F10](#response-types) (checked: no code reads `Encryption`, `IOStats`, `Led` or `IsClone`), and to [R15](#response-types)'s fields (removed, not renamed) |
| field read, appliance sends another type | keep it; fix the type ([F9](#response-types): accept a number; also a numeric string, but only if a real appliance sends one) |
| field read, under a key the appliance does not send | fix the key: [F11](#response-content-and-defaults) would read `parent.name`. In the same commit as [F9](#response-types): once F9 is fixed the snapshot list decodes, and without F11 every entry would be skipped, so ListSnapshots would turn from an error into an empty list with no error |
| field read, docs and appliance agree | unchanged |
| field read but missing in the reply | check how the code handles the zero value; fix in 0003 if it can crash. A general rule is future work ([missing fields in replies](#future-work-missing-fields-in-replies), [risk 9](#risks)) |
| field not read and missing in the reply | no change. [R16](#response-content-and-defaults) |

Removing unused fields would be the smallest change and would work with any format,
past or future. A field could come back, with the right type, when a feature needs
it.

**A2. Audit tool.** The prototype would become `kub-testing/tools/rest-type-audit.py`.
It would read the Go struct definitions, go through the raw hurl replies for each
endpoint, and report type mismatches (SC5). It would also build the compliance
matrix from both appliances' `empirical/<app>/v4/INDEX.md` (SC1, SC3, SC4), and
compare the plugin's request keys with the company `-src` schemas (SC2); prototypes:
`kub-testing/tmp/rest-audit/compliance_matrix.py`, `compliance_bodies.py`. It would
also list every recorded error of a plugin call that A16's catalogue does not cover
(SC4). It would
read the plugin's list of calls and success statuses from `pkg/rest` (the prototype
has them typed in by hand). It would run on the analysis machine, where the spec
repo is, after each new hurl run the spec project adds and after each plugin change
in `pkg/rest`. It would only read the spec repo.

**A3. Return parse errors** ([R10](#logic-bugs)). Two custom decoders swallow a
parse error today and leave the field at zero, so a pool `available` that does not
parse reads as a full pool. They would return the error instead. One of the two is
the snapshot decoder, which the [F9](#response-types) fix rewrites anyway.

### Stage 3: behaviour and error mapping
[Stage_3_behaviour_and_error_mapping](#stage-3-behaviour-and-error-mapping)

Would fix the rest of [response content](#response-content-and-defaults), the
[error messages](#error-messages), the [logic bugs](#logic-bugs) and the
[driver bugs](#driver-bugs):

- Lists and checks that come back empty or wrong. [F12](#logic-bugs) (the iSCSI
  "snapshots of this volume" request is an empty stub): implement the request.
  [F13](#logic-bugs) (the NFS version asks the wrong path): fix the path.
  [F15](#logic-bugs) (the "target not found" pattern is a test string that never
  matches): use the real message. [F16](#response-content-and-defaults) (the
  snapshot kind is read from the wrong key): read `type`, and filter by it: the
  iSCSI lister would drop entries whose parent is a NAS volume, the NFS lister
  those whose parent is a zvol, and other values would be kept
  ([Open question 7](#open-questions)).
  ([F11](#response-content-and-defaults) would be done in stage 2 with F9;
  [F14](#logic-bugs) would already be fixed in the dead-code commit.)
- Error messages the plugin does not recognise. [F18](#error-messages) (a repeated
  snapshot create): accept `Snapshot ... already exists.` [F19](#error-messages)
  (two different 404 replies on detach are treated the same): tell them apart;
  `not attached` means the LUN is already gone, so go on and delete the target;
  `Target ... not exists` means nothing is left to do. [F20](#error-messages) (NFS
  delete of a volume with clones): accept `dataset` as well as `zvol`.
  [F21](#error-messages) (NFS CreateVolume retry): map `Share ... already exists`
  to "exists" and treat an existing share as success. [F22](#error-messages)
  (NFS delete of a clone that is already gone): map to "not found".
- [F26](#error-messages): recognise out of space by its message, whatever the errno,
  instead of only adding the two new messages.
- [F27](#error-messages), REST side only: a missing pool would get its own code, so
  the controller can report a configuration error instead of Unknown; "name too
  long" would map to `RestErrorArgumentIncorrect`, a code that exists but is unused
  today; and on v4 the snapshot list of a missing parent would map to "not found".
  The caller that turns every zvol `DeleteClone` error into `Busy` is a
  [driver bug](#driver-bugs), fixed by A16 below.
- [R14](#logic-bugs): today the code decodes the body first and looks at the HTTP
  status only afterwards, so a reply without a JSON body (a plain-text 401 or 404,
  or a 204 with no body at all) is reported as a parsing problem. The fix: check
  the status first, which also fixes [F23](#logic-bugs); when the body is not
  JSON, map the status itself (401 and 403 to a new "authentication failed" code
  that names the endpoint, 404 to "not found") and keep the first bytes of the body
  in the message. A wrong password would then say so. The JovianDSS 2 design would
  get its "token expired" signal from the same code.
- [F24](#logic-bugs): remove the 202 branch, so a 202 would be treated as an error
  like any other unexpected status ([risk 6](#risks)). [F23](#logic-bugs) and
  [F24](#logic-bugs) would get unit tests with a fake server (A4).
- **A16. Every known error handled, one path for the rest** (user decision): the
  fixes above (F15, F18-F22, F26, F27, R14) are not enough on their own: the next
  reworded message would start the same trouble again. The plugin needs one way of
  handling errors.
  - **A catalogue of known errors.** Every error recorded for a call the plugin
    makes would be mapped on purpose, from its status, `class`, `errno` and message
    pattern to one `RestError` code. The source is `jdss-rest-api-spec`
    (`empirical/<app>/v4/ERROR-FIELDS.md`, `STATUS-CODES.md` and the recorded
    bodies), which lists every error seen, including those with no `class` or no
    `errno`. This is what SC4 asks for. The patterns no recording supports
    ([Table 5](#code-with-no-evidence): "already exists" with errno 1 or 100,
    "not found" without an errno) would stay in the catalogue as legacy entries
    marked "no recording", so that A2's completeness check does not report them
    and the rule below does not remove them.
  - **Codes that say what happened, per kind**, as the Proxmox plugin's `jdssc`
    does (`jdssc/jdssc/jovian_common/exception.py`: not found and already exists
    per volume, snapshot, target and pool; busy; communication failure): for
    example volume not found, snapshot already exists, volume has clones, pool
    missing, out of space, invalid argument, authentication failed. Callers decide
    on the code alone and never parse message text.
  - **One handler for every failed reply**, in `pkg/rest`: it matches the
    catalogue, and anything that does not match becomes `RestErrorUnknown` with the
    status, `class`, `errno`, message and endpoint kept in the error. Nothing is
    guessed: an unknown error is never turned into "busy", "exists" or "not found".
  - **One log line per failed call, in one format**, written by that handler:
    method, path, status, `class`, `errno`, message and the code it was mapped to.
    The handler cannot know whether a failure was expected (only the caller knows
    that "not found" on a repeated delete is fine), so it logs every error it
    recognises at Debug and every unknown error at Error; the caller logs the
    outcome at the level the situation deserves (out of space or a missing pool
    at Error, an expected "not found" not at all). This replaces today's
    scattered Warn lines ([R18](#rest-client-behaviour)) and the empty message of
    unmatched errors ([R9](#rest-client-behaviour)).
  - **Callers pass the meaning on.** The driver and controller turn each code into
    one CSI status, and pass the appliance message on with an unknown error, so the
    CSI caller sees what really happened. No caller maps "any error" to one meaning;
    the zvol `DeleteClone` caller that does so today is
    fixed here ([driver bugs](#driver-bugs), F27's driver part). So is review
    [D12](#driver-bugs): `volume has children` would get its own code, and
    `DeleteVolume` would run the intermediate-snapshot cleanup for it too.
  - **Kept complete by the audit:** A2 would list every recorded error of a plugin
    call that the catalogue does not cover, so a new error in a hurl run shows up as
    a failed check.

  A unit test (A4) would run every recorded error body through the handler and
  check its code, and check that an unknown body keeps all its fields and is logged
  once.
- Review [D13](#driver-bugs): `cleanIntermediateSnapshots` checks the type of the
  snapshot list before it checks whether the listing failed, so a real error such
  as "not found" turns into "unknown failure"; and the NFS version expects zvol
  snapshots although its lister returns NAS snapshots
  (`[]jrest.ResourceNASVolumeSnapshot`), so it always fails on the type check.
  Both would be fixed. Without the second fix, the NFS cleanup would never reach
  the call that the dead-code commit repairs ([F14](#logic-bugs),
  `nfs_driver.go:322`). NFS part: the NFS cycle ([Testing](#testing) step 5);
  the driver checks are recorded in kub-testing (T11).
- Not addressed: snapshot lists lag briefly on jdss2 after a create or delete
  ([What the evidence covers but the plugin
  ignores](#what-the-evidence-covers-but-the-plugin-ignores)). No change is
  proposed; [risk 14](#risks) records what later driver work would have to allow
  for.

### Stage 4: robustness and security
[Stage_4_robustness_and_security](#stage-4-robustness-and-security)

Would fix the REST-layer [crashes](#crashes) (A8), stop request logging (A9),
escape path segments (A10), and add the security measures A11, A12, A13 and A15
(A14 is the [first commit](#first-commit-make-go-vet-pass-a14)). Most of it is
inside `pkg/rest`; A11 and A15 also touch the driver, controller and node, and A12
and A13 add settings to the plugin configuration:

- **A8. Crash fixes**, all inside `pkg/rest` ([Open question 12](#open-questions)):
  - [C3](#crashes): a transport error is one where no HTTP reply arrives at all
    (the address does not resolve, the connection is refused, the request times
    out). Every one of the six branches that handle it reads `res.StatusCode`
    while `res` is `nil`, so any transport error panics today. The fix: return a
    `RestError` without touching `res`, and tell the kinds apart with `errors.As`
    (`*net.DNSError`, `*net.OpError` for a failed dial, `Timeout()` for any
    timeout, including the client's own). The two `fmt.Println` calls in these
    branches (`rproxy.go:152`, `:156`), which write to stdout outside the logger,
    would go; A16's handler logs the error.
  - [C4](#crashes): check the `http.NewRequest` error before `SetBasicAuth`; and
    fix the URL scheme to `https` in `SetupRestProxy` (A12: the plugin always uses
    HTTPS), which would refuse a `prot` setting other than `https` with an error
    that names the setting, so the URL can no longer start with `://`.
  - [C5](#crashes): `ErrorFromErrorT(nil)` would return a `RestError` ("reply has no
    error body") instead of using the `nil`.
  - [C17](#crashes): `SetupEndpoint` would log the error from `SetupRestProxy`
    **and** return it, instead of logging it and returning `nil`; and `Send` would
    return a `RestError` ("REST client not configured", followed by the setup
    error's text) when the proxy was never set up, instead of panicking. Its three
    callers (`controllerplugin.go:193`, `iscsi_driver.go:409`, `nfs_driver.go:546`)
    ignore the returned error today; with this fix a bad configuration would give
    an error on every REST call, naming the cause, instead of a panic. The log
    line stays because the callers ignore the error: without it, the cause of a
    bad `idletimeout`, an empty `addrs`, a missing `prot` or a bad `cafile` (A12)
    would appear nowhere. Making startup fail early on that error is driver work,
    outside `pkg/rest` (for 0003).
  - [R5](#rest-client-behaviour): `RestEndpoint.String()` (`rest.go:83-95`) would
    return the string it builds and would not read `Addrs[0]` when the list is
    empty. Go calls `String()` whenever the endpoint is formatted with `%v` or
    `%s`, so today a log line alone can panic.

  Each gets a unit test (A4): a fake server that refuses or times out (C3), a
  configuration without `prot` and one with `prot: http` (C4, A12), a reply
  without an error body (C5), an invalid `idletimeout` (C17), and an endpoint with
  no addresses formatted with `%v` (R5).
- **A9. No request logging** ([S2](#security-findings), [R17](#security-findings);
  user decision): `Send` would stop logging request bodies (`rproxy.go:113`, `:118`)
  and would no longer dump the request (`:130-131`), so neither the CHAP password
  nor the `Authorization: Basic` header can reach a log. The dump call would be
  removed, not just silenced: today it runs on every request even when nothing is
  logged, so removing it also saves that work. Method, path (with its query string)
  and status may still be logged. This disables request logging for now; whether a
  safe, redacted request log comes back later is a separate decision. A unit test
  (A4) would send a request with a password in its body and check that the
  captured log output contains neither the password nor an `Authorization` header.

- **A10. Escape path segments** ([S3](#security-findings); user decision): every
  name or ID put into a REST path would go through `url.PathEscape`, in all path
  builders in `pkg/rest`. A name then stays one path segment: `/`, `?`, `#`, `%` and
  spaces are encoded and can no longer add segments, cut the path short or start a
  query. `url.PathEscape` does not change `.` or `..`, and does not catch an empty
  name, so it does not stop a name from addressing the parent or the collection;
  that needs name validation, which is not part of A10. The query string that
  `pagedcSuffix` appends (`entrielisters.go:34-51`) is not a segment: segments
  would be escaped first and the query added after, or `?` and `&` would be
  encoded and paging would break. A unit test (A4) would build each path with
  names containing `/`, `?`, `#` and `%` and check that the request goes to the
  escaped path.

- **A11. Validate names and IDs** (user decision): every volume and snapshot ID
  that comes in through a CSI request would be checked where it is parsed
  (`NewVolumeDescFromCSIID`, `driver/lunid.go:147`; `NewSnapshotDescFromCSIID`,
  `driver/snapid.go:118`), against the rules the appliance enforces
  ([Open question 20](#open-questions)); target names are built by the plugin and
  checked at setup (below):

  | Name | Allowed characters | Longest | Why |
  |---|---|---|---|
  | zvol (volume, clone), NAS volume, NAS clone | `a-z A-Z 0-9 _ -` | **131** (user decision) | the appliance refuses anything else, saying so: "Allowed characters in zvol name: a-z, A-Z, 0-9, _, -", "cannot exceed 132 characters"; 131 keeps one character of margin |
  | iSCSI target | **lower case only** (user decision): `^[a-z0-9]([a-z0-9.:-]*[a-z0-9])?$` | 220, the longest seen accepted | the documented pattern without upper case; the appliance also accepts upper case, but the plugin would not use it |
  | snapshot | no empty name, no `.` or `..` | see [future work](#future-work-snapshot-name-limits) | ZFS's own rules, including a limit on the full name, are future work |

  A bad ID would fail with `InvalidArgument` before any REST call ([Open question
  25](#open-questions) says which of the review's panics next to these parsers
  this closes). As a second line,
  the `pkg/rest` path builders would refuse an empty, `.` or `..` segment, which
  A10's escaping does not catch. Why IDs need checking: the plugin does not control
  them. A volume ID comes back from Kubernetes in every later request, and anyone
  allowed to create a static PersistentVolume or a `VolumeSnapshotContent` object
  can put any string there.

  **The plugin's own names would follow the same limit.** The plugin builds the
  appliance name of a volume from the CSI name in one of three forms: `vp_` plus
  the CSI name as it is, when its characters are allowed; `vb_` plus its base32
  form, when they are not; and `vs_` plus a SHA-256 hash (67 characters), when the
  name is too long. `NewVolumeDescFromName` (`driver/lunid.go:65-81`) uses the
  first two forms up to 240 characters, so it can build names of up to 243, which
  the appliance refuses. The threshold would become 128, so every generated name is
  at most 131; longer CSI names would fall through to the `vs_` form.
  Names over 132 characters were never accepted, so no such volume exists. Names
  of exactly 132 were: a volume made from a CSI name of exactly 129 characters
  (`vp_` form), or one whose base32 form has 129 characters (`vb_`), would be
  looked up under its `vs_` name after the change and not be found. Kubernetes
  volume names have 40 characters, so none is expected ([risk 13](#risks);
  [Open question 21](#open-questions)).

  **Target names are built, not received:** `<prefix>:<hash>`
  (`iscsi_driver.go:569`). The hash is lower-case SHA-256 hex; the prefix is the
  `iscsi.iqn` setting (default `iqn.csi.2019-04`, `controllerplugin.go:168-171`),
  so the setting decides whether the name meets the target rule above.
  `SetupControllerPlugin` (`controllerplugin.go:143`) would check the prefix once
  against that rule and return an error for upper case, other characters, or a
  prefix so long that `prefix + 1 + 64` exceeds 220; today only the hash's length
  is checked (`iscsi_driver.go:571`).
- **A12. TLS and transport** (user decision; [S1](#security-findings)):
  - certificates would be verified by default (today `InsecureSkipVerify: true`,
    `rproxy.go:251`);
  - a new `insecure` setting in the REST endpoint configuration (`RestEndpointCfg`)
    would turn verification off;
  - a new `cafile` setting would name a CA certificate to trust, so an appliance
    with a self-signed certificate can still be verified. The recorded appliance
    certificate is self-signed (issuer "Open-E JovianDSS"), so most installations
    would need `cafile` or `insecure` after upgrading ([risk 11](#risks)). The
    file has to reach the container: the install guides and the controller
    manifests would show how to add it to the config Secret
    (`jdss-controller-cfg`) and which path to put in `cafile`;
  - the minimum TLS version would be **1.3** (`tls.Config.MinVersion`;
    [risk 10](#risks)), fixed: not a configuration key, though `pkg/rest` may
    take it as a parameter of its setup ([Open question 26](#open-questions));
  - redirects would not be followed (`CheckRedirect` returning
    `http.ErrUseLastResponse`): the API has no reason to redirect, and a redirect
    means that something other than the API answered. Go already drops the
    `Authorization` header on a redirect to another host, but keeps it on the
    same host;
  - the plugin would always use HTTPS (user decision,
    [Open question 23](#open-questions)): the URL scheme would be fixed to
    `https`, and a `prot` setting other than `https` would be a configuration
    error (C4's check in A8); the configuration guide would mark `prot` as no
    longer needed;
  - the new keys, and the existing keys of `RestEndpointCfg`, would get `yaml`
    tags: the struct has only `json` tags today and is read by yaml.v3 through
    its lowercased-field fallback (backlog item 20);
  - the configuration guide (`docs/configuration.md`) would describe the new
    keys, and the same edit would correct what the project review found wrong in
    it: `pool` belongs at the top level, not under `endpoint`
    (`pkg/common/common.go:76`); the key is `logfile`, not `logpath` (`:75`);
    `tries` and the addresses after the first in `addrs` are not used
    ([R1](#rest-client-behaviour)) and would be marked so.
- **A13. Reply size limit** (user decision): a new setting in the REST endpoint
  configuration, `maxreplysize`, would cap how much of a reply `Send` reads
  (`io.LimitReader` before `io.ReadAll`, `rproxy.go:164`). A reply over the limit
  would fail with an error naming the limit, instead of being read into memory
  whatever its size. Proposed default: 16 MiB; the largest recorded reply is about
  66 KB.
- **A14. Reply text is never a format string:** moved to the
  [first commit](#first-commit-make-go-vet-pass-a14), because `go vet` has to
  pass from the first commit on.
- **A15. No secrets in any log** (user decision; review S2 outside `pkg/rest`): no
  log line would print a whole CSI request, because CSI requests carry `Secrets`.
  Today they are printed at Debug (`controllerplugin.go:587`, `:626`, `:703`,
  `:750`, `:837`, `:914`; `node/publish.go:43`, `:68`) and, worse, at **Warn**
  whenever a create, delete or publish fails (`controllerplugin.go:479`, `:546`,
  `:634`, `:710`, `:857`), so the secrets reach the log at any log level;
  `ControllerGetCapabilities` logs its request at Info (`:1069`; no secrets, but
  the rule is the same). These lines would log only the fields needed (volume ID,
  name, capability). With A9, no secret would
  reach a log at any level. The configuration, which holds the REST password, is
  not logged today (`controllerplugin.go:152-153` are commented out) and would stay
  so.
- **Not possible today:** a REST user with fewer rights than admin. JovianDSS does
  not offer such a role (user), so the plugin keeps using the admin account.

### Supporting work
[Supporting_work](#supporting-work)

**A4. Fixtures and unit tests.** Replies taken from hurl's `out/`, trimmed, with
passwords removed ([Open question 4](#open-questions)), would go into
`pkg/rest/testdata/`, including replies to the per-volume snapshot list that
[F12](#logic-bugs) adds (recorded by `list/volume-kinds.hurl`). The tests would
decode each one with the real Go types, and pass the recorded errors through
`ErrorFromErrorT`; this would confirm the Python simulation. A fake HTTP server
(`net/http/httptest`) would cover G10 and G11; those tests rely on the fixes for
[C3](#crashes) and [C5](#crashes), which stage 4 would make (A8). These would be
the first unit tests in the repository; they would run with
`go test ./pkg/rest/...` on kub-gate (`make test`; the target comes with the
[first commit](#first-commit-make-go-vet-pass-a14)). The tests would use the
standard library only: no new module dependency, because the build targets run
`go mod tidy` and would pick any change up. The first test could run **before**
any fix,
to show the failures [F1](#response-types), [F8](#response-types),
[F8b](#response-types), [F9](#response-types), [F10](#response-types),
[F18-F22](#error-messages), [R14](#logic-bugs) and, with a fake server,
[F23](#logic-bugs) and [F24](#logic-bugs).

**A5. Request side (check only).** The keys the plugin puts into each request body
were compared with the company schema and with what hurl sent. Apart from G1 (the
`{}` body) and G3 (no `sparse`), the appliances accepted the plugin's bodies in the
recordings (`force_umount: true` on zvol, snapshot and clone delete; `name,
snapshot`; `snapshot_name`; target `active, name`; LUN `lun, mode, name`; the NAS
and share forms). Still to do: read what hurl Finding 14 says `force_umount` really
does. Anything found would go to the backlog, not into this design.

**A7. Package documentation.** A new `pkg/rest/doc.go` would hold the package
comment, which Go shows with `go doc` and in editors. It would state the rule of
value 6: the Go types follow the replies recorded from real appliances, not only the
API docs; one field with the wrong JSON type makes the whole reply fail to decode;
and the audit (A2) is re-run after every change to the package and whenever new hurl
recordings are added.

---

## Open Questions
[Open_Questions](#open-questions)

1. ~~**Which API?**~~ — **resolved**: API v4; v3 deprecated.
2. ~~**Minimum JovianDSS 1 release?**~~ — **resolved**: `1.0up33.65410`.
   The spec repo's README names `jdss1` as exactly that release (and `jdss2` as
   `2.0up01.67856`). **The user confirmed that the lab appliance runs
   `1.0up33.65410`, the latest JovianDSS 1 release, and the hurl tests ran on the
   latest version.** So the lab, the recordings and the minimum release are all the
   same release.
3. ~~**When does token auth come?**~~ — **resolved** (user decision): JovianDSS 2
   support, token auth included, would come in a later plugin version, discussed in
   its own design ([JovianDSS 2 support](#future-work-joviandss-2-support)).
4. ~~**Fixtures: real replies or hand-written ones?**~~ — **resolved** (user
   decision): real replies from hurl's `out/`, trimmed to what a test needs, with
   passwords removed: the CHAP passwords and the REST admin user's password,
   including the Basic `Authorization` header that carries it. Other values, such as
   names and IP addresses, may stay.
5. ~~**Unused fields with the wrong type**
   ([F1](#response-types), [F8](#response-types), [F8b](#response-types),
   [F10](#response-types))~~ — **resolved** (user decision): removed from the
   structs (A1), not kept with tolerant types.
6. ~~**Group the reply-type fixes, or fix F1 + F8 alone first?**~~ — **resolved** by
   the decision to deliver and test in one go: all [response types](#response-types)
   and [F11](#response-content-and-defaults) would land together, and
   [F8](#response-types) and [F8b](#response-types), which fail in the same pool
   reply, would be in the same commit.
7. ~~**Should the list functions filter by kind ([F16](#response-content-and-defaults))?**~~
   — **resolved** (user decision): filter by kind, so each driver lists only its
   own snapshots. Both all-snapshot lists return zvol **and** NAS snapshots, and
   stage 3 would read `parent.type`; the iSCSI lister would then drop entries whose
   parent is a NAS volume (type 1) and the NFS lister those whose parent is a zvol
   (type 4). Entries with any other value are kept, so a reply without the field
   would not empty the list ([risk 9](#risks)). The name-format check stays. Its
   check, the recorded all-snapshots reply with both kinds through each lister,
   is driver code and so recorded in kub-testing (T11), not a unit test of this
   delivery ([Testing](#testing) step 1). The NFS fixes themselves were decided
   earlier: in this delivery, with the NFS live cycle ([Testing](#testing)
   step 5).
8. ~~**Evidence gaps G1-G8?**~~ — **resolved**: the user adds them to the
   hurl suite. The most important are G2, G3 and G5.
9. ~~**`sparse` default ([F28](#response-content-and-defaults))?**~~ — **resolved**
   (user decision): `sparse` would be sent explicitly on every volume create, with
   three levels: a StorageClass parameter, if set; otherwise the general default
   from the plugin configuration; otherwise **sparse**. This changes today's
   behaviour for new volumes, which are thick now (existing volumes are not
   touched). It gets a design of its own; this delivery does not change
   provisioning.
10. ~~**Which commit carries the [F14](#logic-bugs) switch?**~~ — **resolved** (user
    decision): the dead-code commit. It removes the broken copy,
    `DeleteNASVolumeSnapshot`, and points its one caller at `DeleteNASSnapshot`.
    Deleting NAS snapshots stays supported.
11. ~~**Is the lab appliance the same machine as hurl's `jdss1`?**~~ — **resolved**
    (user decision): not a design question. The plugin targets the latest JovianDSS
    release, and every appliance used for testing has to run it; one that does not
    is updated first ([Testing](#testing)). Which machines and pools the test
    environment uses is not part of the design.
12. ~~**Where do the REST-layer [crashes](#crashes) go?**~~ — **resolved** (user
    decision): every crash that can be fixed inside `pkg/rest` is fixed in this
    design. All five (C3, C4, C5, C17 and R5's panic) can, so they are in stage 4
    (A8). Only the
    startup handling of C17's error in the three callers is outside `pkg/rest`, for
    0003.
13. ~~**Commit hash in the body?**~~ — **resolved** (user decision): the body says
    "the current code"; only the status line names the commit, linked to GitHub.
14. ~~**Keep the note that this document replaces the original 0002?**~~ —
    **resolved** (user decision): no; it is history, and the note is removed.
15. ~~**Tense?**~~ — **resolved**: the conditional, as the Proxmox design README asks
    of a PROPOSED document. Facts about today's code and the evidence stay in the
    present tense; what the design would change uses "would".
16. ~~**Where do security findings [S1](#security-findings) and [S3](#security-findings)
    go?**~~ — **resolved** (user decision): both in this delivery, in
    [stage 4](#stage-4-robustness-and-security): S1 as part of A12 (verification on
    by default, `insecure` and `cafile` settings), S3 as A10 (`url.PathEscape`).
17. ~~**Where do review findings D12 and D13 go?**~~ — **resolved** (user
    decision): both in this delivery, in stage 3: [D12](#driver-bugs) as part of
    A16, [D13](#driver-bugs) as its own item.
18. ~~**Where does the caller that turns every zvol `DeleteClone` error into `Busy` go?**~~ — **resolved**
    (user decision): in this delivery, as part of A16 in stage 3
    ([driver bugs](#driver-bugs)): it is the kind of misreading A16 rules out.
19. ~~**jdss2 tokens in fixtures?**~~ — **resolved** (user decision): refresh tokens
    are secrets. They would never be logged, and they would be removed from fixtures,
    like the passwords of [Open question 4](#open-questions). Access tokens are not
    named in the decision; this document treats them the same way, because an access
    token is a credential while it is valid (900 s) and no test needs its value.
20. ~~**Exact rules for A11's name check?**~~ — **resolved** (user decision, from the
    spec repo's evidence): the hurl strings tests
    (`empirical/hurl/v4/crude/*-strings.hurl`) give the rules the appliance enforces; A11 lists them. Zvol and NAS names:
    `a-z A-Z 0-9 _ -`, at most **131** characters (the appliance allows 132). Target
    names: **lower case only**. Snapshot limits: future work
    ([snapshot name limits](#future-work-snapshot-name-limits)).
21. ~~**Cap generated names at 131, or at 132?**~~ — **resolved** (user decision):
    131, one character under the appliance's 132; the limit may change later. The
    one cost is accepted: a volume whose name has exactly 132 characters today
    would be looked up under a different name after the change (A11,
    [risk 13](#risks)); Kubernetes never produces such a name.
22. **Who prepares the NFS live cycle?** [Testing](#testing) step 5 needs five
    things that do not exist: `cycle.py` and `lab.py` accepting an `-nfs` image;
    an image name that the `Makefile` produces and the NFS manifests pull (today
    `opene/joviandss-csi-nfs` is pulled and never built); the NFS manifests
    applied for the first time; a way to switch the lab between the two builds;
    and a result comparison that uses only NFS cycles (the first one has no
    baseline). The tooling needs are recorded in kub-testing as TOOL10 and the
    image name as MK11, both in its `ISSUES.md`. Recommendation: a small
    `Makefile` and manifest commit on `master` before this series for the image
    name; none of it in this design's commits. The owner is still to be named.
23. ~~**Is `prot: http` still allowed once certificate checks are on?**~~ —
    **resolved** (user decision): the plugin always uses HTTPS. The URL scheme is
    fixed to `https`; the `prot` setting is no longer needed: absent or `https` is
    accepted, anything else is a configuration error that names the setting (A12,
    through C4's check in A8).
24. ~~**Where do release notes live?**~~ — **resolved** (user decision): in
    `CHANGELOG.md` at the repository root, which now holds the history back to
    v1.0 (before it, the only notes were on the GitHub release page, for
    v1.1.0). Stage 1 and A12 would add their notes (minimum JovianDSS release;
    certificate checks on by default) under its "Unreleased" heading.
25. **Which of C16, C18, C19 and C23 does A11 close?** A11 changes the ID parsers
    in `lunid.go` and `snapid.go`, next to four review findings that no design
    claims: C16 and C19 (`SnapshotDesc.Name()`, `SDS()`, `CSIID()` and
    `VolumeDesc.VDS()`, `CSIID()` panic on an empty value, `snapid.go:147-166`,
    `lunid.go:159-173`), C18 (`IsVDS`/`IsSDS` index `[0]` without a length check,
    `lunid.go:104`, `snapid.go:27`) and C23 (CreateSnapshot with an empty name
    creates `sb_` on the appliance, then `Name()` panics). Reading the code: C18
    is not about IDs: all sixteen callers pass names that came back from the
    appliance (`controller/listers.go:195-251`, the snapshot and clone cleanups in
    both drivers), so A11 does not touch it. C16 and C19 are panics inside
    accessors that A11 does not change. C23 would be closed only if A11 also
    rejected an empty name in `NewSnapshotDescFromName` (`snapid.go:30`), which
    has no such check today, while `NewVolumeDescFromName` has one
    (`lunid.go:61-63`). Recommendation: add that one check to A11 and record C23
    as closed by it; record C16, C18 and C19 as 0003's in `ISSUES.md`.
26. ~~**TLS minimum as a setting?**~~ — **resolved** (user decision): 1.3 is a
    fixed minimum, not a plugin setting. `pkg/rest` may take the minimum as a
    parameter of its own setup, so a test or a later change can pass another
    value, but nothing in the configuration file sets it, and no `tlsmin` key is
    added. A 1.2-only appliance stays a code change ([risk 10](#risks)).
27. ~~**Keep or remove the unread fields that fail?**~~ — **resolved** (user
    decision): remove them, as A1 says, so the types follow the latest v4 spec.
    They are volume `encryption`, the pool, vdev and disk I/O statistics, disk
    `led`, the snapshot clone `is_clone` ([F1](#response-types),
    [F8](#response-types), [F8b](#response-types), [F10](#response-types)) and
    [R15](#response-types)'s NAS fields: the appliance still sends them, with
    other types than the structs declare, and no code reads them. Fields the
    plugin reads keep their real type (A1, second row); fields the appliance
    sends that the structs never had stay ignored. Keeping every field with its
    real type was not chosen ([Options Considered](#options-considered)); the
    rule for fields absent from a reply stays future work
    ([missing fields in replies](#future-work-missing-fields-in-replies)).

---

## Relationship to Other Designs
[Relationship_to_Other_Designs](#relationship-to-other-designs)

- Findings are tracked in `kub-testing/ISSUES.md` and `kub-testing/project-status.md`;
  the review files and ID prefixes are in [ID sources](#id-sources).
- Spec repo `jdss-rest-api-spec`: the reference for compliance
  ([Reference Specification and Compliance](#reference-specification-and-compliance)).
  Only read. Also to report: a name with non-ASCII characters gets **500
  `UnicodeDecodeError`** from the appliance (zvol, NAS and snapshot strings tests)
  instead of a 400 validation error, which is a server bug. The SC7 differences are
  a list for the user to pass on to the spec
  project; this work would not change it.
- JovianDSS-Proxmox plugin: its `jdssc` already has typed errors per kind
  (`jdssc/jdssc/jovian_common/exception.py`) and one general error path
  (`jovian_common/rest.py`, `_general_error`); A16 follows the same idea.
- hurl suite (`jdss-rest-api-spec/empirical/`): it plans to report the bugs it finds
  in API clients such as this plugin; [F9-F28](#appendix), [R7](#logic-bugs) and
  [R14-R16](#appendix) are candidates for that list. Its Finding 4.4 (a delete that
  passes the create status constant) can be downgraded, because the constant is
  ignored (Key observation 5). Its open question on `ListAllSnapshots` (does the
  plugin drop the other kind of snapshot from the combined list?) is answered by
  [F16](#response-content-and-defaults): today no plugin code filters by kind,
  only the name-format check in the list functions can drop an entry; stage 3
  would filter by kind ([Open question 7](#open-questions)).
- Backlog (`review/backlog.md`): item 36 is [F1](#response-types); item 37 (snapshot
  time format) has been verified correct and can be closed; items 19 and 20 (the
  configuration keys the guide gets wrong, and the missing `yaml` tags) are
  addressed by A12.
- [0003 controller crash fixes](0003-controller-crash-fixes-design.md): deferred. Its
  design item D2 (F2) would remove the crash that [F1](#response-types) causes in
  CreateVolume. With 0002 alone, F1's error would go away, so F1 would no longer
  trigger F2. The REST-layer [crashes](#crashes) are fixed here (A8); 0003 keeps the
  controller's own crashes and the startup check of `SetupEndpoint`'s error (C17).
  0003's D3 (the GetCapacity logger fix, F3) exists on kub-gate uncommitted and
  has to be committed on `master` before this series, because the baseline cycle
  was built with it ([Testing](#testing) step 4).
- `sparse` provisioning ([F28](#response-content-and-defaults)): sparse by default,
  configurable in the plugin configuration and per StorageClass; its own design
  ([Open question 9](#open-questions)).
- REST client behaviour: not a compliance matter; a later design
  ([REST client behaviour](#rest-client-behaviour)).
- Security: every [security finding](#security-findings) is fixed in this delivery:
  A9 to A13 and A15 in [stage 4](#stage-4-robustness-and-security), A14 in the
  [first commit](#first-commit-make-go-vet-pass-a14).
- Before any of this can be built, the plugin checkout on kub-gate has to become a
  real clone of `master`: today it is an old commit with the merged code copied on
  top, so builds carry the wrong commit ID and the analysis machine cannot see what
  kub-gate builds (`ISSUES.md` INF8).

### Future work: JovianDSS 2 support
[Future_work_JovianDSS_2_support](#future-work-joviandss-2-support)

**Not part of this design** (user decision): JovianDSS 2 support would come in a
later plugin version and be discussed in its own design. What that design would
have to cover, as known today: JovianDSS 2 has no v3 and accepts only token login
(Auth row of [API versions](#api-versions)), so v4 alone would not make the plugin
work on it. A config setting would choose Basic or token. With tokens the plugin
would have to log in, refresh the token before it expires (900 s) or on a 401, cope
with refresh tokens that work only once, and stay correct when several requests run
at the same time. It touches `rproxy.go`, the configuration and secrets. This design
already stops request logging (A9). Refresh tokens must never be logged (user
decision, [Open question 19](#open-questions)), and neither must access tokens: any
logging the token design adds must keep both out of the logs
([S2](#security-findings), [R17](#security-findings)).
Testing would need a jdss2 pool. This design prepares for it only by keeping the
plugin compliant with jdss2 v4 recordings and by giving 401 its own error code
([R14](#logic-bugs)). That design would also have to allow for replies that are
behind the request, which jdss2 shows most ([risk 14](#risks)).

### Future work: request deadlines
[Future_work_request_deadlines](#future-work-request-deadlines)

Not part of this delivery; for the later REST client design. It addresses
[R8](#rest-client-behaviour) and [risk 12](#risks). Every CSI call arrives with a
deadline, and Kubernetes may cancel it; today neither reaches the HTTP request,
which runs until `idletimeout` expires. The request would be built with the CSI
call's context (`http.NewRequestWithContext`), so the deadline and the cancellation
stop the HTTP call too; `idletimeout` would then be only the idle timeout of the
connection pool, as its name says, not the timeout of every request.

### Future work: snapshot name limits
[Future_work_snapshot_name_limits](#future-work-snapshot-name-limits)

Not part of this delivery; to consider later. Snapshot names follow ZFS's rules, not
the appliance's own check, so a bad one is refused late, by `zfs snapshot`, with a
500 `ZfsCmdError` instead of a 400 validation error. What the hurl strings tests
recorded (zvol and NAS snapshots alike):

- **characters:** `a-z A-Z 0-9 _ - . :` and space are accepted; tab, newline,
  `/ @ # % + , ; = ? & * $ ! ' " \ ~ ( ) <` and any non-ASCII character are refused;
  an empty name is refused;
- **length:** the limit is on the **full ZFS name**, `pool/volume@snapshot`, which
  can be at most 255 characters. With a 31-character volume on a 6-character pool,
  a 216-character snapshot name was accepted and 217 refused: 6 + 1 + 31 + 1 + 216
  = 255. So the longest snapshot name depends on the pool and volume name lengths,
  and no single number is right.

A later check would compute `len(pool) + 1 + len(volume) + 1 + len(snapshot) ≤ 255`
before creating a snapshot, and the snapshot name generator (`driver/snapid.go:37-41`,
which today allows up to 243 characters) would have to stay within it. A16's
catalogue would need the `ZfsCmdError` message for this case, `name is too long`.

### Future work: missing fields in replies
[Future_work_missing_fields_in_replies](#future-work-missing-fields-in-replies)

Not part of this delivery; a starting point for a later design. It addresses
[risk 9](#risks).

**The problem.** Go's `encoding/json` fails only on a wrong type or broken JSON. A
missing key, or `null`, leaves the field at its zero value (`""`, `0`, `false`,
`nil`) without an error, so the code cannot tell "the appliance sent 0" from "the
appliance sent nothing". Once this design removes the type failures, a reply that
lacks a field the plugin reads would pass silently.

**Strategies Go offers:**

- **Pointer fields** (`*int64`, `*string`): a missing key or `null` gives `nil`, so
  absence can be told from zero.
- **A check after decoding:** a `validate()` on each reply type the plugin reads,
  returning an error that names the missing field (for example
  `GetPool: reply has no "available"`) instead of carrying on with 0.
- **An explicit default:** where absence has a known meaning, such as "no quota",
  map it to that meaning on purpose, and write it down per field.

**Proposed rule**, for every field the plugin **reads** (value 4, "bugs stay
visible"):

- **required** fields (for example pool `available` and `size`, snapshot `creation`,
  volume `name`): if absent or `null`, the request fails with an error naming the
  field;
- **optional** fields with a meaning when absent (for example NAS `quota`): a pointer
  field or an explicit default, listed per field;
- fields the plugin does **not** read: removed or ignored, so they cannot fail
  (as A1 already does).

**Supporting checks:** the audit (A2) would also report every read field that is
missing or `null` in any recording, so the required/optional list rests on evidence;
and A4 would get one unit test per required field, with that field removed from the
recorded reply.

---

## Risks & Backward Compatibility
[Risks_and_Backward_Compatibility](#risks--backward-compatibility)

### Preserved (low risk)
[Preserved_low_risk](#preserved-low-risk)

- **Basic auth** would stay; tokens would come only with JovianDSS 2 support, in a
  later version with its own design.
- **What the calls do on success** would stay the same, with these exceptions:
  - the `per_page` commit would change how lists are paged ([Testing](#testing)
    step 4);
  - [F11-F13](#appendix) would make snapshot lists return real entries
    ([risk 4](#risks));
  - stage 3 would remove the 202 branch of `CreateTarget` ([risk 6](#risks));
  - stage 4 could make a call that works today fail: A11 would reject IDs it
    accepts now and refuse, at setup, an `iscsi.iqn` prefix outside the target
    rule; A12 would refuse an unverified certificate, a TLS version below 1.3
    ([risks 10 and 11](#risks)) or a `prot` other than `https`; and A13 would
    refuse a reply over `maxreplysize`.

### Risks
[Risks](#risks)

**Not preserved:** these installations would stop working after the upgrade
until something changes: appliances without API v4 (risk 2), appliances that offer
no TLS 1.3 (risk 10), appliances with a self-signed certificate until `cafile` or
`insecure` is set (risk 11), configurations whose `iscsi.iqn` prefix breaks the
target rule, until the prefix is changed (A11), and configurations with
`prot: http`, which the plugin no longer accepts (A12).

1. **A removed field turns out to be used.** Guard: `go build` would fail on any
   remaining use.
2. **Older releases.** All recordings come from current releases. **A JovianDSS 1
   release without v4, or with a different v4, would stop working, and there would
   be no fallback.** Guard: the docs and release notes would state the minimum
   release, `1.0up33.65410`. It is the latest JovianDSS 1 release, and both the lab
   and the hurl recordings run it (confirmed by the user). Older JovianDSS 1 releases
   would not be supported. The v4 move would be its own commit, so it could be
   reverted on its own.
3. **The Python simulation differs from Go.** Guard: A4 would first decode with the
   real Go types.
4. **[F11-F13](#appendix) would change results, not only errors.** ListSnapshots
   would start returning real entries, which could expose paging bugs (the CSI list
   token) that are hidden today. Guard in this delivery: A4's unit tests, because
   csi-sanity cannot reach ListSnapshots until 0003 stops the controller crashes
   ([Testing](#testing) step 9); csi-sanity's ListSnapshots specs would follow after
   0003. [F18-F22](#error-messages) would likewise change which path a repeated
   request takes.
5. **Recorded data contains passwords and lab details** (names, IP addresses,
   serial numbers; `report.json` also holds certificates and host names). Guard:
   fixtures would be trimmed and their passwords (CHAP and REST admin, including the
   Basic `Authorization` header) and the jdss2 access and refresh tokens removed
   before commit ([Open questions 4 and 19](#open-questions)); lab details may stay.
6. **Removing [F24](#logic-bugs)'s 202 branch.** If an appliance ever answered
   `CreateTarget` with 202, the call would fail instead of polling. 202 was never
   recorded, in about 21,000 requests per run. Guard: the fake-server test (A4) would
   pin the new behaviour, and the error would name the status, so a real 202 would be
   noticed.
7. **One live test per driver for the whole delivery.** A regression could not be
   traced by comparing a live cycle per commit. Guard: every check that does not
   need the lab would pass first ([Testing](#testing) steps 1-3); each commit would
   build on its own; and `git bisect` with `dev_cycle.py` would find the commit if a
   live test failed.
8. **Connected clients are untested ([G9](#situations-not-recorded)).** No recorded
   reply had an initiator logged in or a volume in use, and the iSCSI node code may
   not log in at all today (review N13-N15, `ISSUES.md` T10). The calls that act on a
   target or volume a client may still use have been seen only on idle resources,
   and never on v4 on the lab: detach, target delete, and the iSCSI deletes, which
   always send `force_umount: true` (`iscsi_driver.go:102`, `:240`, `:487`, `:516`).
   The v4 move and [F19](#error-messages)'s change, which would go on to delete the
   target when the LUN is not attached, act exactly on these calls. If the appliance
   refused them while a session is active, unpublish or delete would fail and leave
   targets behind. If it accepted them, a forced delete could remove a volume that a
   node still uses, for example after a node crash: a data-safety question. Guard:
   [Testing](#testing) steps 6 and 7. Until they have run, the behaviour with a
   client connected is unknown, on v3 as on v4.
9. **A reply without an expected field would not be noticed.** Go's
   `encoding/json` fails only on a wrong type: a missing key or `null` leaves the
   field at its zero value without an error. Once this design removes the type
   failures, a reply that lacks a field the plugin reads would go through silently.
   Cases already seen: [F11](#response-content-and-defaults) (no `volume` key, so every
   snapshot entry is skipped and ListSnapshots is empty),
   [R16](#response-content-and-defaults) (zvol snapshot without `name`),
   [R15](#response-types) (NAS `quota` is `null` without a quota, so the size is 0;
   review D10) and [R10](#logic-bugs) (a pool `available` that does not parse
   becomes 0; A3 would fix this one). If a later JovianDSS release dropped or
   renamed a field the plugin reads, the possible failures would include:
   - GetCapacity reporting no free space, so Kubernetes stops placing volumes on the
     pool, or a size of 0 that makes a volume look empty or full;
   - lists that come back empty without an error, so a snapshot or volume seems not
     to exist and a backup tool or the CSI sidecar acts on that;
   - an empty name or ID passed on to the next request, which would then address a
     different path than intended (see also [S3](#security-findings));
   - a zero time, such as a snapshot `creation` of year 1, reported to Kubernetes.

   Guard in this delivery: only the per-field check in A1 and unit tests built on
   recorded replies, which contain the fields. A systematic guard is future work
   ([missing fields in replies](#future-work-missing-fields-in-replies)).
10. **TLS 1.3 only (A12).** Which TLS versions the appliances offer is not
    recorded: no recorded run holds a TLS version string (checked). If an
    appliance, or a supported JovianDSS 1 release, offered only TLS 1.2, the
    plugin could not connect at all. Guard: a TLS handshake check
    (`openssl s_client -tls1_3`) on the lab appliance, jdss1 and jdss2 **before
    the series starts**, so a 1.2-only appliance is known before code is written
    (on the hurl appliances only when no hurl tests are running). There is no
    `tlsmin` setting (user decision, [Open question 26](#open-questions)): a
    1.2-only appliance would need a code change.
11. **Certificate checks on by default (A12) change upgrades.** The recorded
    appliance certificate is self-signed, so an existing installation would stop
    connecting after the upgrade until it sets `cafile` or `insecure`. Guard: the
    release notes say so, with both settings, and the error names the certificate
    problem and the two settings. The lab's live test needs one of them set.
12. **No request deadline ([R8](#rest-client-behaviour)).** The request is built
    without the CSI call's context, so a hung appliance call is not cancelled when
    the CSI deadline passes; it lasts until `idletimeout`, which is used as the
    timeout of the whole request. Guard: none in this delivery; future work
    ([request deadlines](#future-work-request-deadlines)).
13. **Names of exactly 132 characters (A11).** A volume with such a name, made
    from a CSI name of exactly 129 characters, would be looked up under its `vs_`
    name after the change and not be found. Kubernetes volume names have 40
    characters, so none is expected. Guard: none; the edge is accepted with the
    limit of 131 (user decision, [Open question 21](#open-questions)), which may
    change later.
14. **Replies may be behind the request.** Two cases are recorded. On jdss2, a
    snapshot list read within half a second of a snapshot create or delete still
    shows the old state, while a read by name is current at once (hurl
    Finding 20; on jdss1 the lag was seen twice in eight reads, on v3 only). And
    a new clone's `san:volume_id` is empty in the first read of the clone, by name
    or in a list, however long after the create; the next read has it (hurl
    Findings 14 and 15). The plugin decodes that field (`types.go:100`, `:211`)
    and nothing reads it today. Other fields and other lists may behave the same
    way: the visibility tests saw no further case, but they only cover what they
    read.
    This design leaves the logic by which the driver acts on a reply as it is: it
    changes what the client sends and how it decodes and maps what comes back.
    Today that logic takes a reply as the current state: `cleanIntermediateSnapshots`
    deletes the intermediate snapshots one list shows and keeps the rest
    (`iscsi_driver.go:185`, `nfs_driver.go:291`), and ListSnapshots reports what
    one list read returns. A read moments after a change could therefore miss an
    object that exists or show one that is gone, and a step taken on it would be
    a false operation: an intermediate snapshot left behind, a snapshot reported
    to Kubernetes as missing, a field read as empty that the next read would
    fill. The exposure is small on JovianDSS 1 and grows with
    [JovianDSS 2 support](#future-work-joviandss-2-support). Guard: none in this
    delivery. Every later piece of driver work, whether a design, an audit or a
    fix, would have to treat a reply as possibly behind the request: read by name
    where the name is known, retry a list that lacks an expected entry before
    concluding that it is absent, and take no destructive step on the strength
    of a single list read straight after a change.

---

## Files That Would Change (when implemented)
[Files_That_Would_Change_when_implemented](#files-that-would-change-when-implemented)

Rows marked *NFS* are NFS-only code, covered by the NFS cycle ([Testing](#testing)
step 5) and, where it is `pkg/rest` code, by unit tests.

**Table 13 — Files that would change** · `tbl_files_changed`

| File | Change | Stage |
|---|---|---|
| `pkg/rest/entrielisters.go` (`pagedcSuffix`), `pkg/driver/common.go` (`getResourcesList` stop rule) | [F25](#response-content-and-defaults) `per_page` | before 1 |
| `pkg/rest/storageapi.go`, `rest.go`, `rproxy.go`, `payload-def.go`, `reqtypes.go`, `nas_volumes_snapshots.go`, `shares.go`, `targets.go`; `pkg/driver/nfs_driver.go:322` | A6, [dead code](#dead-code); [F14](#logic-bugs) and [R7](#logic-bugs) (switch to `DeleteNASSnapshot`) | before 1 |
| every `pkg/rest/*.go` with a path | one API prefix, `api/v4` | 1 |
| plugin docs, `CHANGELOG.md` | v3 deprecated, minimum release `1.0up33.65410` | 1 |
| `pkg/rest/types.go`, `payload-def.go` | remove unused mismatching fields ([F1](#response-types), [F8](#response-types), [F8b](#response-types), [F10](#response-types), [R15](#response-types) *NFS*); [F9](#response-types), [F11](#response-content-and-defaults) (stage 2), [F16](#response-content-and-defaults) *NFS* (stage 3); A3 | 2, 3 |
| `pkg/rest/*_test.go`, `pkg/rest/testdata/` (new) | A4 | 2, 3, 4 |
| `Makefile` | `vet` and `test` targets | first commit |
| `kub-testing/tools/rest-type-audit.py` (new) | A2 | 2 |
| `pkg/rest/doc.go` (new) | A7, package documentation | 2 |
| `pkg/rest/entrielisters.go` | [F12](#logic-bugs) (a new call, `GET pools/<x>/volumes/<x>/snapshots`, in the [compliance matrix](#compliance-matrix)), [F13](#logic-bugs) *NFS* | 3 |
| `pkg/rest/errors.go`, `targets.go` | [F15](#logic-bugs), [F18](#error-messages), [F19](#error-messages), [F20-F22](#error-messages) *NFS*, [F24](#logic-bugs), [F26](#error-messages), [F27](#error-messages) | 3 |
| `pkg/rest/errors.go`, `storageapi.go` (`getError`); `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/*.go` (callers) | A16: error catalogue, one handler and log line, codes per kind; callers decide on the code only (`iscsi_driver.go:520` included) | 3 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go` (`DeleteVolume`, `cleanIntermediateSnapshots`) | review D12 (cleanup also on `volume has children`, with A16), D13 (listing error first; NFS takes NAS snapshots, *NFS*) | 3 |
| `pkg/rest/storageapi.go` (`getError`), `clones.go`, `pools.go`, `targets.go`, `nas_volumes.go`, `nas_volumes_snapshots.go`, `nas_volumes_snapshots_clones.go`, `shares.go`, `entrielisters.go` | [R14](#logic-bugs) and [F23](#logic-bugs): status checked before decoding, non-JSON bodies mapped by status | 3 |
| `pkg/rest/rproxy.go`, `errors.go`, `rest.go` | A8: [C3](#crashes), [C4](#crashes), [C5](#crashes), [C17](#crashes), [R5](#rest-client-behaviour); A9 in `rproxy.go`: no request logging ([S2](#security-findings), [R17](#security-findings)) | 4 |
| every `pkg/rest/*.go` with a path builder | A10: `url.PathEscape` on every path segment ([S3](#security-findings)) | 4 |
| `pkg/driver/lunid.go` (`NewVolumeDescFromName` threshold 240 → 128), `snapid.go`; `pkg/controller/controllerplugin.go` (`iscsi.iqn` prefix check); `pkg/rest` path builders | A11: ID and name validation, generated names at most 131 characters, target prefix checked | 4 |
| `pkg/rest/rproxy.go`, `pkg/common/common.go` (`RestEndpointCfg`, with `yaml` tags), `docs/configuration.md` (new keys; `pool` and `logfile` corrected; `tries`, extra `addrs` and `prot` marked unused), `docs/install.md`, `docs/install-nfs.md`, the controller manifests (the CA file in the config Secret) | A12: HTTPS only, certificate checks, `insecure`, `cafile`, TLS 1.3, no redirects; A13: `maxreplysize` | 4 |
| `pkg/rest/storageapi.go:51`, `:53`, `errors.go:173`, `:265`, `targets.go:150`, `:184` | A14: `"%s"` formats, so `go vet` passes | first commit |
| `pkg/controller/controllerplugin.go`, `pkg/node/publish.go` | A15: no whole CSI requests in logs | 4 |
| `CHANGELOG.md` | A12: certificate checks on by default, `cafile`/`insecure` needed for self-signed certificates | 4 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/listers.go` | callers of [F11](#response-content-and-defaults), [F12](#logic-bugs), [F13](#logic-bugs) *NFS*, [F19](#error-messages), [F21](#error-messages) *NFS*; the filter by kind ([F16](#response-content-and-defaults), [Open question 7](#open-questions)) | 2, 3 |
| `review/backlog.md` | item 36 addressed; item 37 closed (zero-padded format checked against recorded snapshots); items 19 and 20 addressed by A12; A5 findings added | all |
| `docs/design/0002-rest-api-compliance-design.md` | this document | all |

JovianDSS 2 support: a later version, with its own design.

---

## Testing
[Testing](#testing)

The whole delivery would be tested once. Everything that does not need the lab would
run first and would have to pass; then one live cycle per driver, iSCSI and NFS.
Every appliance used for
testing has to run the latest JovianDSS release; one that does not is updated
before the tests ([Open question 11](#open-questions)).

**Before the live test**

1. **Build and unit tests (kub-gate):** first, `go vet ./...` on the current code,
   with the result recorded: the [first commit](#first-commit-make-go-vet-pass-a14)
   removes the six format-string reports that are known; anything else vet reports
   is decided then. Then `go build ./...` and `go vet ./...` (`make vet`) on every
   commit of the series, so that each commit builds, vet stays clean and
   `git bisect` stays usable; `go test ./pkg/rest/...` (`make test`) on the last
   one. The unit tests (A4) would cover:
   - **decoding:** every recorded v4 reply decoded with the real Go types;
   - **errors:** every recorded error body mapped to its catalogue code, and an
     unknown one kept whole and logged once (A16);
   - **fake server:** [F23](#logic-bugs), [F24](#logic-bugs), G10, G11, and a
     server that ignores `page` (as jdss2 v4 does), which must not cause an endless
     loop ([F25](#response-content-and-defaults));
   - **crash fixes (A8):** a server that refuses or times out ([C3](#crashes)), a
     failing `http.NewRequest` returning an error instead of panicking, a
     configuration without `prot` accepted as HTTPS and one with `prot: http`
     refused ([C4](#crashes), A12), a reply without an error body
     ([C5](#crashes)), an invalid `idletimeout` ([C17](#crashes)), an endpoint with
     no addresses formatted with `%v` ([R5](#rest-client-behaviour));
   - **security (stage 4):**
     - no password and no `Authorization` header in the log output (A9);
     - names with `/`, `?`, `#` and `%` sent to the escaped path (A10);
     - an empty, `.` or `..` path segment refused by the path builders (A11's
       second line);
     - a self-signed certificate refused by default and accepted with `cafile` or
       `insecure`, a TLS 1.2-only server refused, a redirect not followed (A12);
     - a reply over `maxreplysize` refused (A13);
     - a reply message with `%` kept intact (A14, the first commit; its test
       lands with the others in A4).

   All would have to pass. The driver and controller code this design changes
   is not covered by these unit tests (user decision): A11's name and ID rules
   and its `iscsi.iqn` prefix check, A15's request logs, the delete cleanup of
   D12 and D13, the filter by kind ([Open question 7](#open-questions)) and the
   NFS cleanup reaching [F14](#logic-bugs)'s call. Those checks are recorded in
   kub-testing as T11 (`testcases/plugin-checks-design-0002.md`) for later; the
   live cycles (steps 4 and 5) exercise them as far as csi-sanity reaches them.
2. **Audit and compliance (analysis machine):** A2 would find no type mismatch on
   jdss1 v4 or jdss2 v4, and the compliance matrix would show SC1-SC6 passing for
   every call. SC7 has no pass or fail; it is a list of differences. Re-run after
   each new hurl run.
3. **NFS code:** [F13](#logic-bugs), [F14](#logic-bugs), [R7](#logic-bugs),
   [F16](#response-content-and-defaults), [F20-F22](#error-messages),
   [R15](#response-types) and [F27](#error-messages)'s snapshot-list case would
   have unit tests before the live test, like everything else; step 5 then runs
   them live. The NFS part of review [D13](#driver-bugs) is driver code: a
   check recorded in kub-testing (T11), exercised live by step 5.

**Live test**

4. **iSCSI cycle, `dev_cycle.py all`,** compared spec by spec with the baseline
   cycle `20261102-140341`, with `cafile` or `insecure` set for the lab appliance
   ([risk 11](#risks)). Two prerequisites. The baseline build carried the user's
   GetCapacity fix (F3, 0003's D3), which is not on `master` yet, so **F3 has to
   be committed before the series** (`ISSUES.md`, next action 3); without it the
   ctrl-cap preset would crash again and the comparison would be void. And the
   plugin runs at **Debug** level for the cycle, which the `per_page` check below
   needs; that is safe only because A9 and A15 are in the build (until then Debug
   logs secrets, which is why kub-testing forbids it today). It would pass when:
   - no `plugin.log` contains `cannot unmarshal`;
   - the ctrl-A ListVolumes/GetVolume specs and the ctrl-cap GetCapacity spec reach
     the plugin's real logic (they may still fail for other reasons; those would
     become new findings);
   - the request log shows `per_page` on every list call, and no list call asks for
     the same page twice;
   - no spec that passed in the baseline fails.
5. **NFS cycle, requirement.** The NFS build would replace the iSCSI build in the
   lab for one `dev_cycle.py all` cycle, then the iSCSI build would come back. The
   two cannot be installed side by side today, because their manifests use the
   same names and socket path (`ISSUES.md` MF7). The tooling cannot do this yet
   either: `sanity_runner/cycle.py:110-111` refuses any `-nfs` image, `lab.py:32`
   lists iSCSI as the only tested protocol, the `Makefile` never produces the
   image name the NFS manifests pull (`opene/joviandss-csi-nfs`), and the NFS
   manifests have never been applied anywhere. Those are prerequisites of this
   step, outside this design's commits ([Open question 22](#open-questions);
   recorded in kub-testing as TOOL10 and MK11).
   There is no NFS baseline: no NFS build has ever been run live, so this cycle
   becomes the first one. It would pass when:
   - no `plugin.log` contains `cannot unmarshal`, and the request log shows
     `per_page` on every list call;
   - no spec fails because of a finding this design fixes, in particular
     ListSnapshots by volume ([F13](#logic-bugs)), deleting a volume that has clones
     ([F20](#error-messages)), a repeated CreateVolume
     ([F21](#error-messages)), deleting a clone that is already gone
     ([F22](#error-messages)), and deleting a volume after a clone from it
     ([D13](#driver-bugs), [F14](#logic-bugs));
   - every other failure is recorded as a new finding.

   Specs that reach NFS code only after 0003 stops the controller crashes keep
   unit tests as their only cover until then (step 9). If the NFS node path does
   not work in the lab at all, that is recorded as a result, not passed over.
6. **Connected clients ([G9](#situations-not-recorded)), requirement.** csi-sanity's
   node preset alone may not log an initiator in: the iSCSI NodeStage on the current
   code probably reports success without logging in or mounting (review N13-N15).
   So the live test would also include one real iSCSI PVC used by a pod on the lab
   (`ISSUES.md` T10):
   - while the pod runs, `iscsiadm -m session` on its node shows a session to the
     volume's target, and the volume is mounted;
   - after the pod and then the PVC are deleted, the target and the LUN are gone on
     the appliance ([F19](#error-messages)), and the request log shows no failed
     detach, target delete or volume delete.

   The replies to these calls would be kept as the first v4 evidence with a client
   connected. If no session appears (N13-N15 confirmed), G9 stays untested; that is
   recorded as a result, not passed over.
7. **Connected clients, recommendation:** on a throwaway volume, also send
   ControllerUnpublish and then DeleteVolume once while the initiator is still logged
   in, as after a node that did not clean up, and record the replies. This is the
   case where unpublish and `force_umount: true` matter most ([risk 8](#risks)).

Steps 4 to 7 run on the lab appliance, which runs the latest release like every
test appliance.

8. **If a cycle failed:** find the commit with `git bisect` and `dev_cycle.py`, and
   fix it before the delivery is finished, not as a follow-up.

**Limits**

9. csi-sanity reaches [F9-F12](#appendix) and [F18](#error-messages) only after 0003
   stops the controller crashes; until then, unit tests would be their only cover.
10. **Not testable here:** G9 beyond steps 6 and 7 (several nodes, a pool moving
    between nodes), G12, and JovianDSS 2 live, which needs
    [JovianDSS 2 support](#future-work-joviandss-2-support) (a later version) and a
    jdss2 pool.

---

## Appendix
[Appendix](#appendix)

Every finding this design refers to, each described once; the rest of the document
uses only the ID. Where the IDs come from: [ID sources](#id-sources). Jargon:
[Terms](#terms).

### API mismatches
[API_mismatches](#api-mismatches)

Places where the code expects something other than what the appliance does. They
are why stages 2 and 3 exist. Unless a row says otherwise, a mismatch looks the same
on jdss1 v3, jdss1 v4 and jdss2 v4.

#### Response types
[Response_types](#response-types)

Fields whose JSON type in the appliance's reply differs from the Go struct. Any one
of them makes the whole reply fail to decode.

**Table 14 — Response types that do not decode** · `tbl_response_types`

| ID | Type and field | Struct says | Appliance sends | v3 docs | v4 docs | Plugin reads it? | Effect | Seen live? |
|---|---|---|---|---|---|---|---|---|
| F1 | `ResourceVolume.Encryption` ([`types.go:84`](../../pkg/rest/types.go#L84)) | `string` | `{"enabled": false, "method": null, "keystatus": null}` | not listed | object; the example shows `"off"` | no | GetVolume and volume lists → `Internal` | yes |
| F8 | `ResourcePoolIOStats.Read/Write/Chksum` ([`types.go:395`](../../pkg/rest/types.go#L395)), for pool, vdev and disk | `string` | numbers | strings | strings | no | GetCapacity → `Internal` | yes |
| F8b | `Disk.Led` ([`types.go:376`](../../pkg/rest/types.go#L376)) | `string` | `{"locate": false, "fault": false}` | `"off"` | `"off"` | no | the same; hidden behind F8 | present in the replies; never reported, because F8 fails first |
| F9 | `ResourceSnapshotShortProperties.Creation` ([`types.go:273`](../../pkg/rest/types.go#L273)) | `string`, parsed to a number | number | number | number | **yes** | iSCSI ListSnapshots → `Internal` as soon as the pool has a snapshot | not yet: the ctrl-B ListSnapshots specs fail earlier, because the controller has already crashed (`connection refused`) |
| F10 | `ResourceVolumeSnapshotClones.IsClone` ([`types.go:300`](../../pkg/rest/types.go#L300)) | `string` | `true` | `true` | `true` | no | deleting a snapshot that has clones fails at `iscsi_driver.go:498` | not yet |

No **type** mismatch was found in pool `available`/`size`, snapshots read by name,
NAS volumes, NAS snapshots, targets, or error bodies. The other `Encryption` fields
(`types.go:200`, `:560`) are fine in the recordings. `encoding/json` reports only the
first type error in a reply, so a live test shows only the first failure and hides
the rest; that is why all of them would be fixed together and checked with A2 and
A4, not one at a time. For F9 and F10 **the docs are right and the code is
wrong**; for F8 and F8b the docs are wrong, and the code follows them.

**R15: keys that never match.** Not a failure today, but it explains why the NAS
volume struct reads so little. A type check only sees keys that are present. Most
`ResourceNASVolume` tags (`types.go:447-476`) are camelCase (`fullName`, `logBias`,
`primaryCache`, `secondaryCache`, `accessTime`, `deduplication`), while every
recorded NAS volume GET uses snake_case (`full_name`, `logbias`, `primarycache`,
`secondarycache`, `atime`, `dedup`). Only `name`, `quota`, `origin`, `copies` and
`recordsize` are decoded; the other fields stay empty without any error, and nothing
reads them. When NAS volume values are set, `quota`, `reservation`, `refquota`,
`refreservation`, `recordsize` and `copies` are all strings, as the custom decoder
expects. With no quota set, `quota` is `null`, which leaves the size at 0 (review
D10, a driver issue).

#### Response content and defaults
[Response_content_and_defaults](#response-content-and-defaults)

Replies that decode, but whose content the code reads wrongly, and a request default
the plugin never sets.

**Table 15 — Response content and defaults** · `tbl_response_content`

| ID | Where | What happens | Evidence |
|---|---|---|---|
| F11 | `ResourceSnapshotShort.Volume` ([`types.go:294`](../../pkg/rest/types.go#L294)) | there is no such key: the appliance sends `parent.name`. `Volume` stays empty, and the lister skips the entry (`controller/listers.go:54`). **Once F9 is fixed, iSCSI ListSnapshots would return an empty list, without an error** | recorded replies, code reading |
| F16 | `ResourceNASSnapshotParent.UnmarshalJSON` ([`types.go:658`](../../pkg/rest/types.go#L658)) | reads `type` under the tag `copies`, so `Parent.Type` is always 0. Unused today, but both all-snapshot lists return zvol **and** NAS snapshots, and `parent.type` (4 zvol, 1 NAS) is the only way to tell them apart (hurl Finding 14); stage 3 would read it and drop the other kind's entries ([Open question 7](#open-questions)) | code reading |
| F25 | `pagedcSuffix` ([`entrielisters.go:34`](../../pkg/rest/entrielisters.go#L34)), `getResourcesList` (`driver/common.go:52`) | lists are requested with `page` and `_dc`, but without `per_page`. On jdss2 v4, the `volumes` and `nas-volumes` lists then ignore `page` and return the first page again (tried with `page=1` and `page=99`). The loop stops only on an empty page, so CSI ListVolumes never ends when `max_entries` is not set, and returns duplicates when it is | recorded replies (results G2); the loop from code reading |
| F28 | `CreateVolumeDescriptor.Sparse` ([`reqtypes.go:115`](../../pkg/rest/reqtypes.go#L115)) | never set, so every CSI volume gets the appliance default, which is **thick** on both appliances: the full size is reserved | recorded replies (results G3); decision: sparse by default, configurable in the plugin configuration and per StorageClass, in its own design ([Open question 9](#open-questions)) |
| R16 | `GetVolumeSnapshot` ([`storageapi.go:191`](../../pkg/rest/storageapi.go#L191)) | a zvol snapshot GET has **no `name` key** (NAS snapshot GETs do), so `ResourceSnapshot.Name` and the iSCSI `GeneralSnapshot.Name` (`iscsi_driver.go:455`) are always empty. No caller reads it; callers take the name from the snapshot descriptor. Low | recorded replies |

#### Error messages
[Error_messages](#error-messages)

Error replies the plugin does not recognise. Each becomes `RestErrorFailureUnknown`
today, or the wrong code. The plugin uses these codes to decide whether a repeated
request counts as success and whether a cleanup should go on, so an unrecognised
message takes the wrong path.

**Table 16 — Error messages the plugin does not recognise** · `tbl_error_messages`

| ID | Call | Recorded reply | Plugin expects | Effect |
|---|---|---|---|---|
| F17 | DeleteShare, on jdss2 | 404, errno 1, no class, `Share <n> not exists.` | a class and a message | not recognised on jdss2 (recognised on jdss1). Would be moot once the dead-code commit removed `DeleteShare` (it has no other caller) |
| F18 | CreateSnapshot, CreateNASSnapshot | 500, errno 5, `Snapshot <p>/<v>@<s> already exists.` | `Resource .*/.* already exists.` | a repeated create is not seen as "already exists", so the path for repeated requests (`controllerplugin.go:654`, `iscsi_driver.go:55`, `:165`, `nfs_driver.go:66`, `:254`) is not taken |
| F19 | DettachVolumeFromTarget | 404, `Volume name <v> is not attached to target <t>` | any 404 means "target gone" (`targets.go:182`) | ControllerUnpublish on a target that exists but lacks this LUN returns early (`iscsi_driver.go:667`), so the **target is left behind** |
| F20 | DeleteNASVolume | 500, `In order to delete a dataset, you must delete all of its clones first.` | the pattern says `zvol` | an NFS delete of a volume with clones is not seen as `Busy` (`nfs_driver.go:383`, `:396`), so the clone handling is skipped |
| F21 | CreateShare | 409, errno 1, `Share <n> already exists.` | nothing | on an NFS CreateVolume retry, the existing NAS volume is accepted (`nfs_driver.go:113`), but then the share create fails. No retry can succeed |
| F22 | DeleteNASClone | 404, `Clone <n> not exists` | nothing | deleting a clone that is already gone is an error, not success |
| F26 | create on a full pool | 500, errno 4, `cannot create '...': out of space` (zvol, NAS); 500, errno 1000, `cannot create snapshots : out of space` (NAS). Only `400 New zvol size(...) exceeds available space` is recognised | the out-of-space pattern, checked only for errors **without** an errno (`errors.go:270`) | Unknown instead of `ResourceExhausted` (results G5) |
| F27 | several | missing pool: 500, errno 1, `ZpoolError Given zpool '<p>' doesn't exists.`, and for targets 404 `Target partition: <p> not exists.`; name too long: 400, errno 1, `ZvolValidationError Zvol name too long ... 132` (A11 would keep generated names within 131); v4 snapshot list of a missing parent: 404 `ItemNotFoundError Resource <p>/<v> not found.` | nothing | Unknown (results G4, G7, G8). Its driver part, the `DeleteClone` caller, is in [driver bugs](#driver-bugs) |
| - | any call | 400 `ValidationError`/`BadRequest` | nothing; `RestErrorArgumentIncorrect` exists but is unused | reported as an unknown failure today. A16 would map it to `RestErrorArgumentIncorrect`; the plugin's own requests should not cause these |

### Logic bugs
[Logic_bugs](#logic-bugs)

Bugs in the REST code itself, whatever the appliance does.

**Table 17 — Logic bugs in the REST code** · `tbl_logic_bugs`

| ID | Where | What happens | Evidence |
|---|---|---|---|
| F12 | `GetVolumeSnapshotsEntries` ([`entrielisters.go:53`](../../pkg/rest/entrielisters.go#L53)) | an empty stub that sends no request. ListSnapshots by source volume on iSCSI always returns nothing | code reading |
| F13 | `GetNASVolumeSnapshotsEntries` ([`entrielisters.go:198`](../../pkg/rest/entrielisters.go#L198)) | the path lacks `/snapshots`, so it reads the NAS volume itself; the reply has no `entries` key, so the list is empty, without an error | code reading |
| F14 | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:60`](../../pkg/rest/nas_volumes_snapshots.go#L60)), called at `nfs_driver.go:322` | sends `POST` instead of `DELETE`; the appliance answers 405 | hurl Finding 7 |
| F15 | `targetDneMsgPattern` ([`errors.go:96`](../../pkg/rest/errors.go#L96)) | a hard-coded test name, `Target iqn.20215:test2asd not exists.`, so it can never match a real target. The real 404 is `ItemNotFoundError`, `Target <name> not exists.`, so the general error mapping never returns `RestErrorResourceDNETarget`. `DettachVolumeFromTarget` sets that code itself on any 404 (`targets.go:185`), which leads to F19. = review R3 | hurl Finding 4 |
| F23 | every `Get*` that accepts 204 (`GetVolume`, `GetPool`, `GetNASVolume`, `GetVolumeSnapshotClones`, `GetNASVolumeSnapshotClones`; e.g. [`storageapi.go:95-99`](../../pkg/rest/storageapi.go#L95)) | decodes the body **before** checking the status. A real 204 has no body, so decoding fails and returns `RestErrorRPM`: the 204 success branch can never be reached. No GET answered 204 in about 21,000 recorded requests per run | code reading |
| F24 | `CreateTarget` on 202 ([`targets.go:89-98`](../../pkg/rest/targets.go#L89)) | calls `GetTarget` 10 times in a row, with no delay and no stop on success, and returns only the last result. 202 was never recorded, so this may never run. = review R2 | code reading |
| R7 | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:79-85`](../../pkg/rest/nas_volumes_snapshots.go#L79)) | decodes the body **before** checking for 204. A 204 has no body, so decoding fails and the call returns `RestErrorRPM`. Hidden behind F14: fixing only F14's method would make every successful delete report an error. Its siblings `DeleteNASVolume` and `DeleteShare` check for 204 first. Not F23 (that is about GETs) | code reading |
| R10 | `ResourcePool.UnmarshalJSON` ([`types.go:337-340`](../../pkg/rest/types.go#L337)), `ResourceSnapshotShortProperties.UnmarshalJSON` (`types.go:283-286`) | a value that does not parse returns no error and leaves a zero value: pool `available` becomes 0 (GetCapacity would report a full pool), snapshot `creation` becomes the zero time | code reading |
| R14 | `getError` ([`storageapi.go:42-58`](../../pkg/rest/storageapi.go#L42)); every `Get*` decodes before looking at the status | only the JSON body is mapped; the HTTP status is never used. Error bodies that are not JSON (plain text `401 unauthorized` on jdss1, `404 not found`) give `RestErrorRPM` ("Failure during processing response from storage") on GETs, and `RequestMalfunction` ("Failure in sending data to storage", `errors.go:287`) on create and delete, which wrongly points at the network. A wrong password gives no hint about credentials, and a plain 404 is not seen as "not found" | recorded bodies, code reading |

### Driver bugs
[Driver_bugs](#driver-bugs)

Bugs outside `pkg/rest`, in the driver code that calls it. This design fixes them
because they misread REST errors (value 7).

**Table 18 — Driver bugs this design fixes** · `tbl_driver_bugs`

| ID | Where | What happens | Fixed by |
|---|---|---|---|
| D12 | `DeleteVolume` (`iscsi_driver.go:245`, `:257`; `nfs_driver.go:383`, `:396`) | the intermediate-snapshot cleanup runs only for `Busy`. A leftover snapshot without clones gives `volume has children` (errno 1000), mapped to `BusyVolumeHasSnapshots`, so the cleanup never runs, `DeleteVolume` returns `Internal` on every retry, and **the volume can never be deleted** | A16, stage 3 |
| D13 | `cleanIntermediateSnapshots` (`iscsi_driver.go:197-208`, `nfs_driver.go:302-312`) | the listing error is checked after the type check, so a real error such as "not found" becomes `FailureUnknown`; and the NFS version expects the zvol snapshot type although its lister returns NAS snapshots, so it always fails with "Unable to typecast snapshot data" and never reaches the [F14](#logic-bugs) call at `nfs_driver.go:322` | stage 3 |
| F27, driver part | zvol `DeleteClone` caller (`iscsi_driver.go:520`) | turns any `DeleteClone` error, "not found" included, into `Busy` (results G6) | A16, stage 3 |

### Security findings
[Security_findings](#security-findings)

All are fixed in this delivery, in [stage 4](#stage-4-robustness-and-security): S1 by
A12, S2 and R17 by A9 and A15, S3 by A10. Stage 4 also adds A11 and A13; A14 is
the [first commit](#first-commit-make-go-vet-pass-a14).

**Table 19 — Security findings in the REST layer** · `tbl_security`

| ID | Where | Finding | Goes to |
|---|---|---|---|
| S1 | `rproxy.go:251` | TLS certificate checks are always off (`InsecureSkipVerify: true`, no setting), while the admin's Basic credentials go with every request | this delivery: verification on by default, `insecure` and `cafile` settings (A12, stage 4) |
| S2, R17 | `rproxy.go:113`, `:118`, `:130-131` | at Debug level, request bodies (including the CHAP password) are logged, and `httputil.DumpRequestOut` logs the whole request, including the `Authorization: Basic` header. The dump also runs on every request at any log level, and its error is ignored | this delivery: request logging disabled (A9, stage 4) |
| S3 | every path builder in `pkg/rest` | paths are built with `fmt.Sprintf` and no `url.PathEscape`; a `?` or `#` in a name cuts the path short, and a crafted volume handle can reach other resources | this delivery: `url.PathEscape` on every path segment (A10, stage 4) |

The controller and node also log whole CSI requests, `Secrets` included, at Debug
and at Warn (review S2, outside `pkg/rest`); A15 removes that.

### Crashes
[Crashes](#crashes)

Places where the REST layer can crash the controller (a Go panic). Stage 4 would fix
all four inside `pkg/rest` (A8), together with the panic in
[R5](#rest-client-behaviour).

**Table 20 — Crashes in the REST layer** · `tbl_crashes`

| ID | Where | Crash | Trigger |
|---|---|---|---|
| C3 | `rproxy.go:138-157` | every branch that handles a transport error reads `res.StatusCode` while `res` is `nil` | a refused connection (`:141`), or the client's own timeout, which is not a `*net.OpError` (`:153`). During jdss1's 8-hour outage a controller would have crashed again and again (inferred). Gap G10 |
| C4 | `rproxy.go:122-123` | `SetBasicAuth` runs before the error from `http.NewRequest` is checked | a config without `prot` (there is no default) gives the URL `://host`, and `NewRequest` fails |
| C5 | `errors.go:140-149` | `ErrorFromErrorT(nil)` logs the `nil` case, then uses it anyway | a list with `"data": null`, or an error body without `"error"`. Gap G11 |
| C17 | `rest.go:104-107` | `SetupEndpoint` ignores the error from `SetupRestProxy` | a missing or invalid `idletimeout`, or an empty `addrs`: the first REST call panics |

### REST client behaviour
[REST_client_behaviour](#rest-client-behaviour)

Not wrong replies, but how the client itself behaves: retries, failover, timeouts,
error text and logging. For a later REST client design, except R9 and R18 (fixed in
stage 3 by A16), R4 (A14 in the first commit, A8 in stage 4) and R5 (stage 4, A8).

**Table 21 — REST client behaviour** · `tbl_client_behaviour`

| ID | Where | Behaviour |
|---|---|---|
| R1 | `rproxy.go:91`, `:274`, `:281` | `tries` and every address after the first are never used: no retry and no failover (gap G12) |
| R4 | `storageapi.go:51`, `:53` and others | the reply body is used as a format string; transport errors are reported as `RequestMalfunction`, which hides what kind of error it was. Fixed in this delivery: the format-string part by A14 (first commit), the transport-error part by A8 (C3, stage 4) |
| R5 | `rest.go:83-95` | `String()` throws away the string it builds and returns `Addrs[0]`; it panics on an empty list. Fixed in stage 4 by A8 |
| R8 | `rproxy.go:122`, `:257` | the request is built without the context, so the CSI deadline never reaches it; `idletimeout` is used as the timeout for the whole request ([risk 12](#risks), [future work](#future-work-request-deadlines)) |
| R9 | `errors.go:275-277`, `:297` | errors that match no pattern get an empty message; the appliance's text appears only in a Warn log. Fixed in stage 3 by A16 |
| R18 | `errors.go:150`, `:173`, `:265`, `:291` | every appliance error is logged at Warn, including the expected "not found" on repeated deletes; the not-found message repeats a word ("Resource Resource ... do not exists"). Fixed in stage 3 by A16 |

### Dead code
[Dead_code](#dead-code)

Code with no callers in the current code. The
[dead-code commit](#before-stage-1-remove-dead-rest-code-a6-with-the-f14-switch)
would remove it.

- **Functions:** `ListVolumes` (with `GetVolumesData`), `getResultsEntries`,
  `GetTimeStamp`, `CreateNASVolumeSnapshot` (R11: a copy of `CreateNASVolume` that
  would create a NAS volume), `DeleteNASVolumeSnapshot` (F14, R7; its one caller
  would be switched in the same commit), `PromoteClone` and `AddUserToTarget` (stubs
  that return nil), `DeleteShare` (R12: sends nil instead of its descriptor; its
  only caller is commented out at `nfs_driver.go:344`; the recordings show that
  deleting a NAS volume removes its shares, so it is not needed).
- **Types:** `StorageInterface` (old signatures that `RestEndpoint` does not
  implement), `RestProxyInterface`, the duplicate `restProxy` struct, `RestProxyCfg`,
  `StorageCfg`, `SnapshotDescriptor`, and the unused `*Data` reply types in
  `payload-def.go` (about 16).
- **Request keys** the driver never sets and the company schema does not have (SC2):
  `TargetLunDescriptor.EUI`, `CreateNASVolumeDescriptor.ResWithDescendents` and
  `.RefquotaWithDescendents`.

### ID sources
[ID_sources](#id-sources)

**Table 22 — Where IDs come from** · `tbl_id_sources`

| IDs | Meaning | Source |
|---|---|---|
| F1-F28 | findings from tests and recorded replies | `kub-testing/ISSUES.md` (status per row); those this design acts on are described here in the Appendix |
| C, D, K, N, R, S, X + number | code review findings | [`review/2026-10-02-code-review.md`](../../review/2026-10-02-code-review.md) (before the NFS merge) and [`review/2026-11-07-code-review.md`](../../review/2026-11-07-code-review.md) (the current code, same IDs). R14-R18 come from its detailed `pkg/rest` pass; there is no R13 (it turned out to be F8b). D10, D12 and D13 are driver findings there ([driver bugs](#driver-bugs) lists D12 and D13) |
| D1-D7 in 0003 | 0003's own design items, not review IDs | [0003](0003-controller-crash-fixes-design.md) |
| G1-G12 | evidence gaps | [Evidence Gaps](#evidence-gaps) |
| SC1-SC7 | compliance rules | [What "compliant" means](#what-compliant-means) |
| A1-A16 | design items | [Recommended Solution](#recommended-solution) |
| hurl Finding n, C24 | hurl suite findings and caveats | `jdss-rest-api-spec/empirical/hurl/project-status.md` |
| INF8, MF7, MK11, T11, TOOL10 | kub-testing items: lab infrastructure, manifests, Makefile, testing, tooling | `kub-testing/ISSUES.md` |

### Terms
[Terms](#terms)

**Table 23 — Terms** · `tbl_terms`

| Term | Meaning |
|---|---|
| jdss1 | the hurl suite's JovianDSS 1 appliance: release `1.0up33.65410`, API v3 and v4, Basic auth |
| jdss2 | the hurl suite's JovianDSS 2 appliance: release `2.0up01.67856`, API v4 only, token auth |
| lab appliance | the JovianDSS 1 appliance the Kubernetes lab and csi-sanity use; it runs the latest JovianDSS release, like every test appliance |
| hurl, hurl suite | the HTTP tests in `jdss-rest-api-spec/empirical/hurl/` that send requests to real appliances and record the replies |
| run | one execution of the hurl suite on one appliance; its folder in `hurl/out/` is named by date and time |
| csi-sanity | the standard CSI conformance test suite, run against the plugin in the lab |
| preset | a group of csi-sanity tests run together, such as ctrl-A, ctrl-B, ctrl-cap or node |
| cycle | one `dev_cycle.py` run: build the plugin, deploy it to the lab, run the csi-sanity presets. Named by date and time, e.g. `20261102-140341` |
| kub-gate | the lab machine where the plugin is built, deployed and unit-tested (`dev_cycle.py` runs there) |
| analysis machine | the machine with the spec repo and the hurl recordings, where the audit scripts run; it has no Go |
| company spec, empirical spec | the API as Open-E documents it, and the API as the appliances behaved in the hurl runs ([Table 7](#the-jdss-rest-api-spec-repository)) |
| decode | turn a JSON reply into a Go struct (Go's `encoding/json`) |
| zvol | a ZFS block volume; the plugin uses it for iSCSI volumes |
| NAS volume | a ZFS file system (dataset); the plugin uses it for NFS volumes |
| errno, class, message | the three fields of a JovianDSS error reply: `{"error": {"errno": …, "class": …, "message": …}}` |
| `RestError` code | the plugin's own code for a REST failure (`errors.go`), e.g. `RestErrorResourceDNE` ("does not exist"); the driver turns it into a CSI status |
| `Internal`, `Unknown`, `AlreadyExists`, `ResourceExhausted`, … | CSI (gRPC) status codes the plugin returns to Kubernetes |
| repeated request | CSI may send the same create or delete again; the plugin must treat "already exists" on a create, and "not found" on a delete, as success |
| fixture | a recorded reply saved in the repo as test data for unit tests |
| stage 1-4 | the four groups of commits in this design ([Recommended Solution](#recommended-solution)); not the CSI NodeStage call |
| `vp_`, `vb_`, `vs_`; `sp_`, `sb_`, `ss_` | the prefixes of the names the plugin gives volumes (`v`) and snapshots (`s`) on the appliance: `p` for the CSI name as it is, `b` for its base32 form, `s` for a SHA-256 hash of it (A11) |
| intermediate snapshot | a snapshot the plugin creates on its own when it makes a volume from another volume; the new volume is a clone of that snapshot, and `cleanIntermediateSnapshots` removes such snapshots when the volume is deleted |
| transport error | a request that gets no HTTP reply at all: the address does not resolve, the connection is refused, or the request times out ([C3](#crashes)) |
| Basic auth, Bearer token | two ways to log in to the REST API: Basic sends the user name and password with every request; a Bearer token is obtained once by logging in and sent instead of the password until it expires |
| CHAP | the password login of iSCSI between a node and a target; the plugin does not use it today |
| initiator, target, LUN | iSCSI terms: the initiator is the client (here the Kubernetes node), the target is what the appliance exposes, and a LUN is one volume attached to a target |
| static PV | a PersistentVolume an administrator writes by hand, naming an existing volume, instead of one the plugin creates on request |
