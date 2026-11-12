# 0002 — API v4 compliance for JovianDSS 1

## Table of Contents
[Table_of_Contents](#table-of-contents)

- [List_of_Tables](#list-of-tables)
- [Metadata](#metadata)
- [Status](#status)
- [Problem_statement](#problem-statement)
- [Context](#context)
  - [API_versions](#api-versions)
  - [The_reply_types](#the-reply-types)
  - [The_observed_API_in_the_spec_repository](#the-observed-api-in-the-spec-repository)
  - [The_REST_layer_measured](#the-rest-layer-measured)
  - [What_we_do_not_know](#what-we-do-not-know)
  - [Key_observations](#key-observations)
- [Priorities](#priorities)
- [Options_considered](#options-considered)
  - [Option_A_v4_only_the_only_option_considered](#option-a-v4-only-the-only-option-considered)
- [Decision](#decision)
- [What_changes](#what-changes)
  - [First_commit_make_go_vet_pass](#first-commit-make-go-vet-pass)
  - [Before_stage_1_send_per_page_on_every_list](#before-stage-1-send-per_page-on-every-list)
  - [Before_stage_1_remove_dead_REST_code](#before-stage-1-remove-dead-rest-code)
  - [Stage_1_move_to_API_v4](#stage-1-move-to-api-v4)
  - [Stage_2_response_types](#stage-2-response-types)
  - [Stage_3_behaviour_and_error_mapping](#stage-3-behaviour-and-error-mapping)
  - [Stage_4_robustness_and_security](#stage-4-robustness-and-security)
  - [Supporting_work](#supporting-work)
- [Consequences](#consequences)
- [Risks](#risks)
- [Affected_systems_and_files](#affected-systems-and-files)
- [Testing](#testing)
- [Open_questions](#open-questions)
  - [Decisions](#decisions)
- [Relationship_to_other_designs](#relationship-to-other-designs)
  - [Future_work_JovianDSS_2_support](#future-work-joviandss-2-support)
  - [Future_work_request_deadlines](#future-work-request-deadlines)
  - [Future_work_snapshot_name_limits](#future-work-snapshot-name-limits)
  - [Future_work_missing_fields_in_replies](#future-work-missing-fields-in-replies)
- [Advice](#advice)
- [Appendix](#appendix)
  - [Evidence](#evidence)
  - [Evidence_gaps](#evidence-gaps)
  - [The_jdss_rest_api_spec_repository](#the-jdss-rest-api-spec-repository)
  - [What_compliant_means](#what-compliant-means)
  - [Compliance_matrix](#compliance-matrix)
  - [Response_types](#response-types)
  - [Response_content_and_defaults](#response-content-and-defaults)
  - [Error_messages](#error-messages)
  - [Logic_bugs](#logic-bugs)
  - [Driver_bugs](#driver-bugs)
  - [Security_findings](#security-findings)
  - [Crashes](#crashes)
  - [REST_client_behaviour](#rest-client-behaviour)
  - [Dead_code](#dead-code)
  - [Measurements](#measurements)
  - [ID_sources](#id-sources)
  - [Terms](#terms)

## List of Tables
[List_of_Tables](#list-of-tables)

1. `tbl_api_versions` — **API versions and auth**
2. `tbl_type_generation` — **Generated reply types**
3. `tbl_files_changed` — **Files that would change**
4. `tbl_evidence_sources` — **Evidence sources**
5. `tbl_spec_forms` — **What the spec repo holds**
6. `tbl_spec_appliances` — **Appliances in the spec repo**
7. `tbl_compliance_matrix` — **Compliance matrix per call**
8. `tbl_compliance_status` — **Compliance status per rule**
9. `tbl_id_sources` — **Where the names and IDs come from**
10. `tbl_terms` — **Terms**

## Metadata
[Metadata](#metadata)

Date: 2026-11-05 (API v4 only decided), 2026-11-10 (the whole migration, compliance with the recorded v4 API, decided as one record); first written 2026-11-05 as the document named under Formerly, in this form 2026-11-09, merged into one record 2026-11-10
Driver: Andrei Perepiolkin
Approver: <name>
Code: commit `11ffbfd`, called "the current code" below; line numbers refer to it
Implementation: the [What changes](#what-changes) section, one series of commits delivered and tested in one go
Formerly: `docs/design/0002-rest-api-compliance-design.md`, merged into this document on 2026-11-10

## Status
[Status](#status)

Proposed

## Problem statement
[Problem_statement](#problem-statement)

The plugin uses JovianDSS REST API v3, which is deprecated, through a client
that follows documentation and assumptions rather than what the appliance
serves. It has to use API v4 through a client that complies with the API as
served. Success is [compliance](#what-compliant-means) with the JovianDSS 1
REST API v4. JovianDSS 2 will not be supported within this iteration.

## Context
[Context](#context)

### API versions
[API_versions](#api-versions)

**Table 1 — API versions and auth** · `tbl_api_versions`

| What | jdss1 | jdss2 | Plugin | Evidence |
|---|---|---|---|---|
| API v3 | served | **gone:** 401 without a token, 404 with one | uses only `api/v3` | hurl Finding 9 |
| API v4 | served | served | not used | hurl Finding 9 |
| Auth | Basic | **Bearer token only** (log in with `POST auth/login`; the access token lasts 900 s; each refresh token works once) | Basic only (`rproxy.go:123`) | hurl Findings 8, 9 |

**On JovianDSS 1, v3 and v4 answer the same.** The hurl suite compared them
directly: in 363 list and delete cases (hurl Finding 14) the two versions
returned the same status, errno, class and message in every case, and in 281
rollback and clone cases (hurl Finding 15) they did the same, apart from a few
requests that got no answer at all and one count that varied. Only the shape
of list replies differs: v4 list replies have one shape, `entries` and
`results`, where v3 returns a plain array without paging and the v4 shape
with it ([Key observation 2](#key-observations)). So the move from v3 to v4
by itself should not change how JovianDSS 1 behaves. The JovianDSS Proxmox
plugin already uses v4 by default ([Key observation 1](#key-observations)).

### The reply types
[The_reply_types](#the-reply-types)

**The hand-written reply types fail on every version in the same way.** Six
mismatches exist on the current code. Five are fields the plugin never reads:
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch)
breaks GetVolume and ListVolumes and
[bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) breaks
GetCapacity, both seen live;
[bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), in every
recorded pool reply, hides behind the second;
[bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch)
breaks the delete of a snapshot that has clones; and
[bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec)
breaks nothing, because its camelCase keys never match and the fields stay
empty. One is a field the plugin reads:
[bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch)
breaks iSCSI ListSnapshots once the pool has a snapshot, not yet seen live
because the controller crashes earlier. They fail the same way on jdss1 v3,
jdss1 v4 and jdss2 v4 ([Key observation 3](#key-observations)).
No other type mismatch exists in pool `available` and `size`, snapshots read
by name, NAS volumes, NAS snapshots, targets or error bodies. Documentation
and appliance disagree in both directions: for [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) the documentation
is right and the code wrong; for [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) the documentation says strings,
the appliance sends numbers and an object, and the code follows the
documentation. The decoder reports only the first type error in a reply, so
a live test shows one failure at a time; [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) has never been reported for that
reason. Full detail, with the documentation columns:
[response types](#response-types).

### The observed API in the spec repository
[The_observed_API_in_the_spec_repository](#the-observed-api-in-the-spec-repository)

**The spec repository already holds a validated, machine-readable description
of what each appliance sends.** `jdss-rest-api-spec` is read only for this
work; the user runs its tools.

| What | Where | Size and state |
|---|---|---|
| Empirical TypeSpec, one file per method, types as observed | `empirical/jdss1/v4/typespec/` | 104 files; 108 for jdss2 v4 |
| Compiled OpenAPI 3.1 | `empirical/pilot/tools/tsp/b_jdss1_v4/out/openapi/openapi.yaml` | 24,163 lines; 24,788 for jdss2 v4; present for all four trees |
| Validation | `hurl/build_typespec.py --check` | compiles each tree and validates every recorded response against it; fails on an invalid response or on a key the spec does not have |
| Manifest | `b_<appliance>_<api>.manifest.json` | every recorded response behind each route and status |

How the observed types are written ([empirical
README](../../../jdss-rest-api-spec/empirical/README.md)): a field is
required when it was never seen missing; a field that was `null` in every
recording is typed `null`; values seen at most four times are listed; a zvol
and its clone are separate alternatives split on `is_clone`; other shapes
are merged, so optional fields do not show which go together. For the
plugin's calls, the observed types say exactly what the six mismatches say.
Types differ per endpoint for the same thing: a snapshot's `creation` is
`int64` in the all-snapshots list and a string in the reply by name;
`volsize`, `available` and the pool's `size` are strings holding numbers,
which the hand-written decoders parse today. jdss1 and jdss2 agree on types
for the endpoints the plugin uses. The compiled YAML has a defect: line 15389
of the jdss2 v4 file holds a literal tab inside a `pattern:` value, which
Python's YAML parser refuses; the emitter can write JSON instead.

### The REST layer, measured
[The_REST_layer_measured](#the-rest-layer-measured)

**The plugin's REST layer, measured** on commit `11ffbfd` (the
[Appendix](#appendix) has the detail): `pkg/rest` is 3,806 lines. The reply
types are `types.go`, 675 lines, 19 `Resource*` structs, with 10 accessor
methods and 8 custom decoders that turn string numbers into `int64` and
strings into times. The 36 call functions, the request types (`reqtypes.go`,
254 lines), the transport and the error mapping are separate files. Outside
`pkg/rest`, the drivers and the controller use 7 of the 19 types at about 40
places, almost all `.Name`, plus `GetSize`, `OriginSnapshot`, `ClonesNames`
and `Creation`. Twelve lines read a field whose type or key would change:
`Creation` as a field at six places, the list entry's `Volume` at three (the
key the appliance sends is `parent.name`, [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead)), the pool's `Available` at
one, and the two drivers' copies of `Creation`.

What the code reviews and the recordings found in this layer is catalogued
in the Appendix: replies whose [types](#response-types) do not decode,
[error messages](#error-messages) the code does not recognise,
[logic bugs](#logic-bugs) such as a status checked after the body is
decoded, [crashes](#crashes) on a transport error or an empty body, and
[security findings](#security-findings): secrets in the logs, certificate
checks off, unescaped path segments.

### What we do not know
[What_we_do_not_know](#what-we-do-not-know)

What the recordings do not show: how JovianDSS 1 releases older than
`1.0up33.65410` behave on v4, since no such appliance was available; the user
states that they do not serve the same v4, so none is supported
([question 29](#decisions), [risk 2](#risks));
some requests on jdss1 v4, because the run meant to close the remaining gaps
there was lost to an appliance outage ([results of the gap
runs](#how-the-gap-runs-went));
and the type of any field that was `null` in every recording.

The [evidence gaps](#evidence-gaps) in the Appendix hold all twelve, with what the two runs made to close them found.

### Key observations
[Key_observations](#key-observations)

1. **v3 ends with JovianDSS 1.** It is deprecated, the newer generation does
   not serve it, and the sister plugin has already left it. Staying on v3 is
   a decision to stop at JovianDSS 1.
2. **The move alone changes no behaviour on JovianDSS 1.** v3 and v4 answer
   the same in every recorded case; only the shape of list replies differs,
   and the plugin always asks for the paged shape.
3. **The move and the types are separable in commits.** The same fields fail
   on every version, so the path change can be one commit that changes
   nothing else, the types another, and a live failure can be bisected to
   one of them.
4. **The types can be derived instead of maintained.** A validated
   description of what the appliance sends exists, per endpoint; the only
   thing it lacks is the plugin's knowledge of which fields matter.
5. **Exactness is needed only where the plugin reads.** A handful of fields
   per type carry the plugin's logic; the rest can be stored without a type
   and can then never fail.
6. **The other modules depend on names, not on shapes.** Seven type names
   and a few accessors are the whole surface; keep those and the drivers
   change at twelve lines.
7. **The expected-status argument of `Send` is ignored** (`rproxy.go:84`). `Send`
   takes an argument that says which HTTP status counts as success, but never looks
   at it; each caller checks the status itself. So when a delete passes the create
   constant by mistake, `CreateCloneRCode` (hurl Finding 4.4), nothing goes wrong:
   `DeleteClone` checks for 204 on its own.
8. **Error handling depends on message text** (`errors.go`). If the appliance
   rewords a message, the plugin's behaviour changes and nothing reports an error.
   [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)'s fixtures would catch this when new hurl runs are added, and [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) would keep
   all message patterns in one catalogue, so callers decide on codes, not on text.

## Priorities
[Priorities](#priorities)

The project values that weigh on this document, in the order they weigh here:

1. **Works with the latest JovianDSS 1 release.**
2. **Security.**
3. **Evidence over documentation.** Paths and types rest on recorded replies,
   not on what the company documentation says a version does.
4. **Bugs stay visible.** A field the plugin reads keeps an exact type and
   fails loudly when absent or changed; a field it does not read cannot fail.
5. **Repeatable checks.** A new recording regenerates the types and re-runs
   the audit; the diff is the report.
6. **Smallest change for the callers, one code path, revertable commits.**
   One set of paths and fixtures; the seven type names and the accessors stay;
   the path move and the types are commits that can be reverted alone.
7. **Right first time.** The whole series is delivered and tested in one go; every check that does not need the lab passes before the live tests, so that no follow-up fixes are needed. The REST layer's structure stays; it is reworked where this document says, not rewritten.
8. **Errors mean what they say.** Every known appliance error has one deliberate meaning; an unknown one stays unknown, logged in full in one place, never guessed into another meaning. Misreading an error, such as taking "not found" for "busy", is a bug, not a fallback.

## Options considered
[Options_considered](#options-considered)

One matter is decided here: the API version (A), where only one option was
considered. How the client is made compliant, the reply types, the error
mapping and the transport, is implementation, decided as questions, each
with its alternatives in its resolution, named in the [Decision](#decision)
where they shape the delivery and kept in the [Decisions](#decisions) table.

### Option A: v4 only, the only option considered
[Option_A_v4_only_the_only_option_considered](#option-a-v4-only-the-only-option-considered)

Only one option was considered for the API version: replace the `api/v3`
prefix with `api/v4` everywhere, deprecate v3 in the plugin, with no fallback
and no setting, and state the minimum JovianDSS 1 release in the
documentation and the changelog. Why no other: API v3 is deprecated, as the
[Problem statement](#problem-statement) says; it ends with JovianDSS 1,
JovianDSS 2 serves only v4, and the Proxmox plugin has already moved
([Key observation 1](#key-observations)). Keeping v3, or carrying both
versions behind a setting or a check at start-up, would be a decision to stop
at JovianDSS 1, so neither was weighed (user decision). The recordings show
the move to be behaviour-neutral on JovianDSS 1
([Key observation 2](#key-observations)), which is what lets it be one commit
that changes nothing else.

The price of the one option: a JovianDSS 1 release without v4, or with a
different v4, stops working after the upgrade, with no fallback
([risk 2](#risks)); the minimum release has to be stated and kept true
([Stage 1](#stage-1-move-to-api-v4)).

## Decision
[Decision](#decision)

**The plugin's REST client is brought to
[compliance](#what-compliant-means) with the JovianDSS 1 REST API v4. The
plugin uses v4 only: v3 is deprecated, with no fallback and no setting, and
the minimum supported JovianDSS 1 release is `1.0up33.65410`, the latest at
the time of the decision and the one the recordings and the test lab run.
JovianDSS 2 is not supported within this iteration.**

Rationale, from [Option A](#option-a-v4-only-the-only-option-considered)
and the [Risks](#risks): v3 ends with JovianDSS 1, and the recordings show
that the move changes no behaviour there (priority 1); its price, a release
without this v4 stopping with no fallback ([risk 2](#risks)), is accepted
because no such release is supported ([question 29](#decisions)). The
security findings are fixed in the same series (priority 2), at the price
of upgrades that stop until the certificate and the TLS version are dealt
with ([risks 10 and 11](#risks)). Compliance is measured against the
recordings, not the documentation (priority 3); what they never show stays
a risk, not an assumption ([risks 15 and 16](#risks)). Only a read field can
fail, and it fails by name (priority 4), the guard of [risk 9](#risks). A
new recording is a diff and a check (priority 5), the guard of
[risks 17, 18 and 22](#risks). One API, one code path, the names the other
modules use kept, and commits that revert alone (priority 6). The series is
delivered and tested once, every check that needs no lab first (priority 7,
[risk 7](#risks)). Every known appliance error means one thing and an
unknown one stays unknown (priority 8). Without a guard in this delivery: a
hung call past the CSI deadline ([risk 12](#risks)), the accepted edge of a
name of exactly 132 characters ([risk 13](#risks)) and a reply behind the
request ([risk 14](#risks)).

The decisions below the options were taken as questions and are kept in the
[Decisions](#decisions) table: the reply types are generated from the
recorded jdss1 v4 API under the same names, only a read field can fail, and
the drivers change where they read a field whose type or key changes
(question 28); every appliance error is mapped from its recorded message
only, and an unknown one is logged once in full
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue));
each driver lists only its own snapshot kind (question 7); every ID and name
is validated at the CSI boundary, generated names are at most 131 characters
and target names lower case (questions 20 and 21); the plugin speaks HTTPS
only, TLS 1.3 fixed, with certificate checks on by default (questions 23,
26); no secret reaches any log at any level and no request is logged
(question 19, [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)); release notes live in
`CHANGELOG.md` (question 24).

## What changes
[What_changes](#what-changes)

Everything here would be delivered in one go: one series of commits, built, checked and
tested together ([Testing](#testing)). The commits would come in this order: a
`go vet` clean-up ([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply)), `per_page`, dead-code removal, the move to v4, response
types (stage 2), behaviour and error mapping (stage 3), robustness and security
(stage 4). Stage 1 alone reaches v4 and stage 2 decodes its replies; the client
is compliant with the recorded v4 API only when stages 3 and 4 are done as
well. Each commit would
build on its own, so if the live test failed, `git bisect` could find the commit
that caused it. JovianDSS 2 support (token auth) is not part of this delivery: it
would come in a later plugin version, with its own design ([JovianDSS 2
support](#future-work-joviandss-2-support)).

### First commit: make `go vet` pass
[First_commit_make_go_vet_pass](#first-commit-make-go-vet-pass)

#### impl_plugin_rest_no_format_string_from_reply

**Its own commit, first in the series.** [Testing](#testing) step 1 runs `go vet`,
Go's built-in static checker, on every commit, so vet has to pass before the series
starts. Since Go 1.24, vet's
printf check reports a format string that is not a constant and has no arguments;
the module is Go 1.26, so the check applies. The current code has six such calls,
all with reply text as the format: `fmt.Sprintf(string(body[:]))`
(`storageapi.go:51`), `l.Warnf(msg)` (`storageapi.go:53`, `errors.go:173`, `:265`)
and `l.Debugf(msg)` (`targets.go:150`, `:184`). **[impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply)** turns them into `"%s"`
formats, which is also the fix for [bug_plugin_rest_reply_text_as_format_string](#bug_plugin_rest_reply_text_as_format_string)'s format-string
part: a `%` in an appliance message can no longer garble the message or the log.
The same commit would add `vet` and `test` targets to the `Makefile`, so that
[Testing](#testing) step 1 is one command from the first commit on (`make test`
has nothing to run until [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)). Nothing else would change in this commit. No vet
result has been recorded for the current code; the first step of the work is to
run it and record what else it reports, if anything.

### Before stage 1: send `per_page` on every list
[Before_stage_1_send_per_page_on_every_list](#before-stage-1-send-per_page-on-every-list)

**Its own commit, before the v4 move.** Without it, the v4 move could make CSI
ListVolumes run forever ([bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page)): on jdss2 v4 a
list request without `per_page` ignores the page number and returns the first
page again and again. On jdss1 v4 this was never recorded (gap [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page)), so it may
behave the same way. Paging with `per_page` works on both jdss1 v3 and jdss2 v4.
The stop rule would change too. Today the plugin asks for page after page until
a page comes back empty; it would also stop once it has received as many entries
as the reply's `results` count announces, so a server that repeats a page could
not keep it looping.

### Before stage 1: remove dead REST code
[Before_stage_1_remove_dead_REST_code](#before-stage-1-remove-dead-rest-code)

#### impl_plugin_rest_remove_dead_code_before_v4_move

**Its own commit, after the `per_page` commit and before the v4 move,** so that the
v4 commit would only touch code that is used. It would remove everything listed
under [dead code](#dead-code).

One of the functions to remove, `DeleteNASVolumeSnapshot`, is a broken copy of
`DeleteNASSnapshot` (`nas_snapshots.go:67`): both delete one NAS snapshot at the
same path, but the copy sends `POST` instead of `DELETE`. Its one caller is the
cleanup of an intermediate snapshot at `nfs_driver.go:322`. This commit would point
that call at `DeleteNASSnapshot`, which the normal snapshot delete already uses
(`nfs_driver.go:193`, `:623`), so deleting NAS snapshots stays supported, through
one working function ([question 10](#decisions)). The
build would keep working, and **[bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) and [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check) would be
fixed here** without new code. This commit only makes that call correct; the
cleanup around it would still not run properly until stage 3 fixes review findings
[bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children) and [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) ([driver bugs](#driver-bugs)).

Dead code outside `pkg/rest`, such as the old node implementation in
`pkg/node/targetapi.go` and `target.go`, is not part of this commit.

### Stage 1: move to API v4
[Stage_1_move_to_API_v4](#stage-1-move-to-api-v4)

One API prefix, set to `api/v4`, would replace the hard-coded `api/v3` strings: 34
in the current code (plus 3 in comments), 30 after the dead-code commit. There would
be no setting and no fallback to v3. **Nothing else would change in this commit**,
so if the live test failed, `git bisect` could tell an effect of v4 from an effect
of a fix. No change in behaviour is expected on JovianDSS 1 (hurl Findings 14, 15).
Of the evidence gaps, [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) (the empty `{}` body the plugin sends with every GET) made
no difference on jdss1 v3 and jdss2 v4, and the live test would be its first check
on jdss1 v4; [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page) and [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse) became findings of their own,
[bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) (paging) and
[bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) (thick volumes), handled elsewhere.

**Deprecating v3:** the plugin docs would say it needs a JovianDSS release with API
v4; `CHANGELOG.md` would name the oldest supported JovianDSS 1 release,
`1.0up33.65410`; the generated types, the audit and the fixtures would use v4
replies; comments and
constants that mention v3 would be removed.

### Stage 2: response types
[Stage_2_response_types](#stage-2-response-types)

Would replace the hand-written [response types](#response-types) with types
generated from the recorded API ([question 28](#decisions)). That closes
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch),
[bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) and
[bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) in one commit, and parse errors would be
returned instead of swallowed ([impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors)).

#### impl_plugin_rest_generated_reply_types

**Generated reply types.** A generator in the plugin repository,
`tools/rest-types-gen.py`, would read the spec repo's compiled OpenAPI for jdss1
v4 (`empirical/pilot/tools/tsp/b_jdss1_v4/out/openapi/`, which its
`build_typespec.py --check` validates against every recorded reply) and a
mapping file, `pkg/rest/types_gen.yaml`, and write `pkg/rest/types_gen.go`,
committed, with a header naming the spec commit, the evidence runs and the
generator version. The 19 `Resource*` type names would stay, so the drivers and
the controller keep compiling against them; the accessors and the conversions
(`GetSize`, `OriginSnapshot`, `ClonesNames`, the parsing of numbers sent as
strings that the custom decoders do today) would move to a hand-written
`pkg/rest/types_methods.go`. The spec repo stays read only: the user runs its
tools; the build never needs them, because the generated file is committed.
`make gen-types` would run the generator; `make check-types` would regenerate
and fail on a difference, the guard against the committed file drifting from
the spec.

**Table 2 — Generated reply types** · `tbl_type_generation`

| What | Rule |
|---|---|
| a field the plugin reads | the type the recordings show for that endpoint, exact. The mapping names it and marks it **required** or **optional**; where the recordings never show it set, the mapping may state the type by hand, with the reason |
| a required read field absent from a reply | the call fails with an error that names the field: a generated `checkRequired()` per type, called after decoding. First list: pool `available` and `size`, snapshot `creation`, volume `name`, and snapshot `name` in list entries and in the NAS snapshot reply by name; the zvol snapshot reply by name carries no `name` ([bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key)) and nothing reads it there |
| an optional read field absent from a reply | the accessor returns the documented absent value. First list: `origin` and `clones` on a volume without them, NAS `quota` without a quota ([bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec), review `bug_plugin_driver_nas_quota_null_size_zero`) |
| a field the plugin does not read | kept as `json.RawMessage`: stored as sent, never able to fail. [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) and the keys of [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) become this |
| the key of a read field | the recorded key: [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) reads `parent.name`, and [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key)'s `type` comes under its own key (stage 3 filters by it) |
| a field never seen set in any recording | `json.RawMessage`, unless the mapping types it by hand |
| the zvol-or-clone alternatives of a volume reply | one struct with the fields of both, optional |
| a single-value enum, such as `is_clone` | its base type |
| one thing typed differently per endpoint, such as snapshot `creation`, a number in the list and a string by name | each type follows its endpoint; the accessor in `types_methods.go` converts, as the decoders do today |
| jdss2 | outside this document; the later JovianDSS 2 design decides how its types join ([JovianDSS 2 support](#future-work-joviandss-2-support)) |

The callers would change at the twelve lines that read a field whose type or key
changes: `Creation` through an accessor (`controller/listers.go:52`, `:82`,
`:124`, `:148`; `controllerplugin.go:679`, `:802`; `driver/iscsi_driver.go:456`;
`nfs_driver.go:591`), the list entry's `Parent.Name` instead of a `Volume` key
the appliance never sends (`listers.go:54`, `:56`; `driver/restwrappers.go:40`)
and the pool's `GetAvailable()` (`controllerplugin.go:1017`).

#### impl_lab_rest_audit_tool

**Audit tool.** The prototype would become `kub-testing/tools/rest-type-audit.py`.
It would build the compliance matrix from jdss1 v4's
`empirical/jdss1/v4/INDEX.md` ([rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted), [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled)), with jdss2 v4 beside it as
evidence, and compare the plugin's request
keys with the company `-src` schemas ([rule_rest_request_body_valid](#rule_rest_request_body_valid)); prototypes:
`kub-testing/tmp/rest-audit/compliance_matrix.py`, `compliance_bodies.py`. It
would also list every recorded error of a plugin call that [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s catalogue does
not cover ([rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled)). It would read the plugin's list of calls and success statuses
from `pkg/rest` (the prototype has them typed in by hand). It would run on the
analysis machine, where the spec repo is, after each new hurl run the spec
project adds and after each plugin change in `pkg/rest`. It would only read the
spec repo. The type check it was to make ([rule_rest_reply_types_decode](#rule_rest_reply_types_decode)) is done by generation instead: a
new recording is a `make gen-types` and a diff of the generated file.

#### impl_plugin_rest_return_parse_errors

**Return parse errors** ([bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors)). Two custom decoders swallow a
parse error today and leave the field at zero, so a pool `available` that does
not parse reads as a full pool. With [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types) the conversions move to
`types_methods.go` and would return the error instead; the snapshot decoder is
replaced by the generated type and its accessor.

### Stage 3: behaviour and error mapping
[Stage_3_behaviour_and_error_mapping](#stage-3-behaviour-and-error-mapping)

Would fix the rest of [response content](#response-content-and-defaults), the
[error messages](#error-messages), the [logic bugs](#logic-bugs) and the
[driver bugs](#driver-bugs):

- Lists and checks that come back empty or wrong. [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) (the iSCSI
  "snapshots of this volume" request is an empty stub): implement the request.
  [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) (the NFS version asks the wrong path): fix the path.
  [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name) (the "target not found" pattern is a test string that never
  matches): use the real message. [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) (the
  snapshot kind is read from the wrong key): read `type`, and filter by it: the
  iSCSI lister would drop entries whose parent is a NAS volume, the NFS lister
  those whose parent is a zvol, and other values would be kept
  ([question 7](#decisions)).
  ([bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) would be done in stage 2 with [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch);
  [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) would already be fixed in the dead-code commit.)
- Error messages the plugin does not recognise. [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) (a repeated
  snapshot create): accept `Snapshot ... already exists.` [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)
  (two different 404 replies on detach are treated the same): tell them apart;
  `not attached` means the LUN is already gone, so go on and delete the target;
  `Target ... not exists` means nothing is left to do. [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised) (NFS
  delete of a volume with clones): accept `dataset` as well as `zvol`.
  [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) (NFS CreateVolume retry): map `Share ... already exists`
  to "exists" and treat an existing share as success. [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised)
  (NFS delete of a clone that is already gone): map to "not found".
- [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised): recognise out of space by its message, whatever the errno,
  instead of only adding the two new messages.
- [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised), REST side only: a missing pool would get its own code, so
  the controller can report a configuration error instead of Unknown; "name too
  long" would map to `RestErrorArgumentIncorrect`, a code that exists but is unused
  today; and on v4 the snapshot list of a missing parent would map to "not found".
  The caller that turns every zvol `DeleteClone` error into `Busy`
  (`iscsi_driver.go:520`) is the driver part of this finding, fixed by
  [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) below.
- [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status): today the code decodes the body first and looks at the HTTP
  status only afterwards, so a reply without a JSON body (a plain-text 401 or 404,
  or a 204 with no body at all) is reported as a parsing problem. The fix: check
  the status first, which also fixes [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check); when the body is not
  JSON, map the status itself (401 and 403 to a new "authentication failed" code
  that names the endpoint, 404 to "not found") and keep the first bytes of the body
  in the message. A wrong password would then say so. The JovianDSS 2 design would
  get its "token expired" signal from the same code.
- [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay): remove the 202 branch, so a 202 would be treated as an error
  like any other unexpected status ([risk 6](#risks)). What must hold: a GET
  answered 204 is the empty result, not a parse error ([bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check)),
  and a 202 on `CreateTarget` is an error naming the status.

#### impl_plugin_rest_error_catalogue

**Every known error handled, one path for the rest** (user decision): the
fixes above ([bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised), [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status)) are not enough on their own: the next
reworded message would start the same trouble again. The plugin needs one way of
handling errors.
- **A catalogue of known errors.** Every error recorded for a call the plugin
  makes would be mapped on purpose, from its status, `class`, `errno` and message
  pattern to one `RestError` code. The source is `jdss-rest-api-spec`
  (`empirical/<app>/v4/ERROR-FIELDS.md`, `STATUS-CODES.md` and the recorded
  bodies), which lists every error seen, including those with no `class` or no
  `errno`. This is what [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) asks for. The patterns no recording supports
  ([evidence gaps](#evidence-gaps): "already exists" with errno 1 or 100,
  "not found" without an errno) would stay in the catalogue as legacy entries
  marked "no recording", so that [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)'s completeness check does not report them
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
  scattered Warn lines ([bug_plugin_rest_every_error_logged_at_warn](#bug_plugin_rest_every_error_logged_at_warn)) and the empty message of
  unmatched errors ([bug_plugin_rest_unknown_error_empty_message](#bug_plugin_rest_unknown_error_empty_message)).
- **Callers pass the meaning on.** The driver and controller turn each code into
  one CSI status, and pass the appliance message on with an unknown error, so the
  CSI caller sees what really happened. No caller maps "any error" to one meaning;
  the zvol `DeleteClone` caller that does so today is
  fixed here ([bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised)'s driver part, `iscsi_driver.go:520`). So is review
  [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children): `volume has children` would get its own code, and
  `DeleteVolume` would run the intermediate-snapshot cleanup for it too.
- **Kept complete by the audit:** [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool) would list every recorded error of a plugin
  call that the catalogue does not cover, so a new error in a hurl run shows up as
  a failed check.

What must hold: every recorded error body maps to its catalogue code, and an
unknown body keeps all its fields and is logged once.

- Review [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error): `cleanIntermediateSnapshots` checks the type of the
  snapshot list before it checks whether the listing failed, so a real error such
  as "not found" turns into "unknown failure"; and the NFS version expects zvol
  snapshots although its lister returns NAS snapshots
  (`[]jrest.ResourceNASVolumeSnapshot`), so it always fails on the type check.
  Both would be fixed. Without the second fix, the NFS cleanup would never reach
  the call that the dead-code commit repairs ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete),
  `nfs_driver.go:322`). NFS part: the NFS cycle ([Testing](#testing) step 5);
  the driver checks are recorded in kub-testing (T11).
- Not addressed: snapshot lists lag briefly on jdss2 after a create or delete
  ([What the evidence covers but the plugin
  ignores](#what-the-evidence-covers-but-the-plugin-ignores)). No change is
  proposed; [risk 14](#risks) records what later driver work would have to allow
  for.

### Stage 4: robustness and security
[Stage_4_robustness_and_security](#stage-4-robustness-and-security)

Would fix the REST-layer [crashes](#crashes) ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)), stop request logging ([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging)),
escape path segments ([impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments)), and add the security measures [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids), [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)
([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) is the [first commit](#first-commit-make-go-vet-pass)). Most of it is
inside `pkg/rest`; [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) also touch the driver, controller and node, and [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)
and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) add settings to the plugin configuration:

#### impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes

**Crash fixes**, all inside `pkg/rest` ([question 12](#decisions)):
- [crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error): a transport error is one where no HTTP reply arrives at all
  (the address does not resolve, the connection is refused, the request times
  out). Every one of the six branches that handle it reads `res.StatusCode`
  while `res` is `nil`, so any transport error panics today. The fix: return a
  `RestError` without touching `res`, and tell the kinds apart with `errors.As`
  (`*net.DNSError`, `*net.OpError` for a failed dial, `Timeout()` for any
  timeout, including the client's own). The two `fmt.Println` calls in these
  branches (`rproxy.go:152`, `:156`), which write to stdout outside the logger,
  would go; [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s handler logs the error.
- [crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check): check the `http.NewRequest` error before `SetBasicAuth`; and
  fix the URL scheme to `https` in `SetupRestProxy` ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): the plugin always uses
  HTTPS), which would refuse a `prot` setting other than `https` with an error
  that names the setting, so the URL can no longer start with `://`.
- [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body): `ErrorFromErrorT(nil)` would return a `RestError` ("reply has no
  error body") instead of using the `nil`.
- [crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error): `SetupEndpoint` would log the error from `SetupRestProxy`
  **and** return it, instead of logging it and returning `nil`; and `Send` would
  return a `RestError` ("REST client not configured", followed by the setup
  error's text) when the proxy was never set up, instead of panicking. Its three
  callers (`controllerplugin.go:193`, `iscsi_driver.go:409`, `nfs_driver.go:546`)
  ignore the returned error today; with this fix a bad configuration would give
  an error on every REST call, naming the cause, instead of a panic. The log
  line stays because the callers ignore the error: without it, the cause of a
  bad `idletimeout`, an empty `addrs`, a missing `prot` or a bad `cafile` ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile))
  would appear nowhere. Making startup fail early on that error is driver work,
  outside `pkg/rest` (for 0003).
- [crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs): `RestEndpoint.String()` (`rest.go:83-95`) would
  return the string it builds and would not read `Addrs[0]` when the list is
  empty. Go calls `String()` whenever the endpoint is formatted with `%v` or
  `%s`, so today a log line alone can panic.

What must hold, one condition per crash: a refused or timed-out connection returns a
`RestError` that names its kind ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)); a
configuration without `prot` works as HTTPS and one with `prot: http` is refused with
an error naming the setting ([crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check), [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)); a reply
without an error body returns a `RestError` ([crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body)); an invalid `idletimeout` makes every
REST call return the setup error ([crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error)); an endpoint with
no addresses formats with `%v` without a panic ([crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs)).

#### impl_plugin_rest_no_request_logging

**No request logging** (`security_plugin_secrets_in_logs`, `security_plugin_rest_request_dump_every_call`;
user decision): `Send` would stop logging request bodies (`rproxy.go:113`, `:118`)
and would no longer dump the request (`:130-131`), so neither the CHAP password
nor the `Authorization: Basic` header can reach a log. The dump call would be
removed, not just silenced: today it runs on every request even when nothing is
logged, so removing it also saves that work. Method, path (with its query string)
and status stay logged at Debug: [Testing](#testing) step 4 reads `per_page` from
that line. This disables request logging for now; whether a
safe, redacted request log comes back later is a separate decision. What must
hold: a request with a password in its body leaves neither the password nor an
`Authorization` header in the log output, at any level.


#### impl_plugin_rest_escape_path_segments

**Escape path segments** ([security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments); user decision): every
name or ID put into a REST path would go through `url.PathEscape`, in all path
builders in `pkg/rest`. A name then stays one path segment: `/`, `?`, `#`, `%` and
spaces are encoded and can no longer add segments, cut the path short or start a
query. `url.PathEscape` does not change `.` or `..`, and does not catch an empty
name, so it does not stop a name from addressing the parent or the collection;
that needs name validation, which is not part of [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments). The query string that
`pagedcSuffix` appends (`entrielisters.go:34-51`) is not a segment: segments
would be escaped first and the query added after, or `?` and `&` would be
encoded and paging would break. What must hold: a name containing `/`, `?`, `#`
or `%` reaches the appliance as one escaped path segment.


#### impl_plugin_driver_validate_names_and_ids

**Validate names and IDs** (user decision): every volume and snapshot ID
that comes in through a CSI request would be checked where it is parsed
(`NewVolumeDescFromCSIID`, `driver/lunid.go:147`; `NewSnapshotDescFromCSIID`,
`driver/snapid.go:118`), against the rules the appliance enforces
([question 20](#decisions)); target names are built by the plugin and
checked at setup (below):

| Name | Allowed characters | Longest | Why |
|---|---|---|---|
| zvol (volume, clone), NAS volume, NAS clone | `a-z A-Z 0-9 _ -` | **131** (user decision) | the appliance refuses anything else, saying so: "Allowed characters in zvol name: a-z, A-Z, 0-9, _, -", "cannot exceed 132 characters"; 131 keeps one character of margin |
| iSCSI target | **lower case only** (user decision): `^[a-z0-9]([a-z0-9.:-]*[a-z0-9])?$` | 220, the longest seen accepted | the documented pattern without upper case; the appliance also accepts upper case, but the plugin would not use it |
| snapshot | no empty name, no `.` or `..` (for the ID parsed from a CSI request; the name built at CreateSnapshot is [question 25](#open-questions)) | see [future work](#future-work-snapshot-name-limits) | ZFS's own rules, including a limit on the full name, are future work |

A bad ID would fail with `InvalidArgument` before any REST call ([Open question
25](#open-questions) says which of the review's panics next to these parsers
this closes). As a second line,
the `pkg/rest` path builders would refuse an empty, `.` or `..` segment, which
[impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments)'s escaping does not catch. Why IDs need checking: the plugin does not control
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
[question 21](#decisions)).

**Target names are built, not received:** `<prefix>:<hash>`
(`iscsi_driver.go:569`). The hash is lower-case SHA-256 hex; the prefix is the
`iscsi.iqn` setting (default `iqn.csi.2019-04`, `controllerplugin.go:168-171`),
so the setting decides whether the name meets the target rule above.
`SetupControllerPlugin` (`controllerplugin.go:143`) would check the prefix once
against that rule and return an error for upper case, other characters, or a
prefix so long that `prefix + 1 + 64` exceeds 220; today only the hash's length
is checked (`iscsi_driver.go:571`).

#### impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile

**TLS and transport** (user decision; [security_plugin_rest_tls_verification_off](#security_plugin_rest_tls_verification_off)):
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
  take it as a parameter of its setup ([question 26](#decisions));
- redirects would not be followed (`CheckRedirect` returning
  `http.ErrUseLastResponse`): the API has no reason to redirect, and a redirect
  means that something other than the API answered. Go already drops the
  `Authorization` header on a redirect to another host, but keeps it on the
  same host;
- the plugin would always use HTTPS (user decision,
  [question 23](#decisions)): the URL scheme would be fixed to
  `https`, and a `prot` setting other than `https` would be a configuration
  error ([crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check)'s check in [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)); the configuration guide would mark `prot` as no
  longer needed;
- the new keys, and the existing keys of `RestEndpointCfg`, would get `yaml`
  tags: the struct has only `json` tags today and is read by yaml.v3 through
  its lowercased-field fallback (backlog item 20);
- the configuration guide (`docs/configuration.md`) would describe the new
  keys, and the same edit would correct what the project review found wrong in
  it: `pool` belongs at the top level, not under `endpoint`
  (`pkg/common/common.go:76`); the key is `logfile`, not `logpath` (`:75`);
  `tries` and the addresses after the first in `addrs` are not used
  ([bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover)) and would be marked so.

#### impl_plugin_rest_reply_size_limit

**Reply size limit** (user decision): a new setting in the REST endpoint
configuration, `maxreplysize`, would cap how much of a reply `Send` reads
(`io.LimitReader` before `io.ReadAll`, `rproxy.go:164`). A reply over the limit
would fail with an error naming the limit, instead of being read into memory
whatever its size. Proposed default: 16 MiB; the largest recorded reply is about
66 KB.

#### impl_plugin_no_secrets_in_logs

**No secrets in any log** (user decision; review `security_plugin_secrets_in_logs` outside `pkg/rest`): no
log line would print a whole CSI request, because CSI requests carry `Secrets`.
Today they are printed at Debug (`controllerplugin.go:587`, `:626`, `:703`,
`:750`, `:837`, `:914`; `node/publish.go:43`, `:68`) and, worse, at **Warn**
whenever a create, delete or publish fails (`controllerplugin.go:479`, `:546`,
`:634`, `:710`, `:857`), so the secrets reach the log at any log level;
`ControllerGetCapabilities` logs its request at Info (`:1069`; no secrets, but
the rule is the same). These lines would log only the fields needed (volume ID,
name, capability). With [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), no secret would
reach a log at any level. The configuration, which holds the REST password, is
not logged today (`controllerplugin.go:152-153` are commented out) and would stay
so.

- **Not possible today:** a REST user with fewer rights than admin. JovianDSS does
  not offer such a role (user), so the plugin keeps using the admin account.

### Supporting work
[Supporting_work](#supporting-work)

#### impl_plugin_rest_recorded_reply_fixtures

**Fixtures, and what the unit tests must show.** Replies taken from hurl's `out/`, trimmed, with
passwords removed ([question 4](#decisions)), would go into
`pkg/rest/testdata/`, including replies to the per-volume snapshot list that
[bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) adds (recorded by `list/volume-kinds.hurl`); no jdss2 reply is a
fixture ([question 19](#decisions)). The tests on them are the implementing
agent's. What they must show: every fixture decodes with the generated types
([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), which holds by construction and
stays the regression check of the generator and its rules; a required read
field removed from a fixture fails with the error naming it; every recorded
error body passed through `ErrorFromErrorT` gets its catalogue code; and a fake HTTP server
(`net/http/httptest`) covers [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) and [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies), which relies on the fixes for
[crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error) and [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body), which stage 4 would make ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)). These would be
the first unit tests in the repository; they would run with
`go test ./pkg/rest/...` on kub-gate (`make test`; the target comes with the
[first commit](#first-commit-make-go-vet-pass)) and use the
standard library only: no new module dependency, because the build targets run
`go mod tidy` and would pick any change up. Decoded with the current
hand-written types, **before** any fix, the same fixtures show the failures
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch),
[bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch),
[bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised), [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status) and, with a fake server,
[bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) and [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay): the fixtures are
evidence before they are tests.

#### impl_plugin_rest_request_body_keys_checked_against_schema

**Request side (check only).** The keys the plugin puts into each request body
were compared with the company schema and with what hurl sent. Apart from [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) (the
`{}` body) and [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse) (no `sparse`), the appliances accepted the plugin's bodies in the
recordings (`force_umount: true` on zvol, snapshot and clone delete; `name,
snapshot`; `snapshot_name`; target `active, name`; LUN `lun, mode, name`; the NAS
and share forms). Still to do: read what hurl Finding 14 says `force_umount` really
does. Anything found would go to the backlog, not into this document.

#### impl_plugin_rest_package_documentation

**Package documentation.** A new `pkg/rest/doc.go` would hold the package
comment, which Go shows with `go doc` and in editors. It would state the rule of
[priority 5](#priorities), repeatable checks: the reply types are generated from the replies recorded from real
appliances ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)) and never edited by hand; only the fields the plugin reads carry a
type, named in `types_gen.yaml`, and the rest are raw JSON; one field with the
wrong JSON type makes the whole reply fail to decode, which is why read fields
are exact and the others are not; `make check-types` runs after every change to
the package and `make gen-types` whenever new recordings are added, with the audit
([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)).

## Consequences
[Consequences](#consequences)

- One API to implement, record, audit and test; every fixture is a jdss1 v4
  reply.
- Installations on an older JovianDSS 1 release stop working after the
  upgrade until the appliance is updated; the documentation and
  `CHANGELOG.md` name the minimum release
  ([risk 2](#risks)).
- `types.go` is replaced by a generated `types_gen.go` and a hand-written
  `types_methods.go`; the generated file is committed, so the build never
  needs the spec repository or Node. Every field the appliance sends is
  stored, 86 on the volume, and only the read ones have a type; the rule for
  absent read fields exists from the first generation. Appliance states that
  no recording shows, such as a volume with encryption enabled, keep their
  fields tolerant until recorded.
- Stage 1 is the path move and the deprecation; stage 2 is the mapping, the
  generator, the generated types, the methods file, the twelve driver lines
  and the fixtures. The audit ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) keeps only the compliance matrix; the
  fixtures ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) become the generator's regression test; the future work on
  missing fields shrinks to the fields the plugin does not read
  ([missing fields in replies](#future-work-missing-fields-in-replies)).
- The spec repository becomes an input of the development process: a new
  recording means the user runs its `--check`, then `make gen-types`, and
  reviews the diff.
- The whole of this document is implemented as one series; stages 3 and 4
  need no further record: their files, risks and tests are in
  [Affected systems and files](#affected-systems-and-files),
  [risks 4 to 8, 10, 11, 13 and 14](#risks) and [Testing](#testing).

Confidence: high for JovianDSS 1 at the recorded release, where v3 and v4 are
recorded identical and every reply of every call is validated against the
spec the types come from; medium for read fields never seen set in the
recordings, whose type the mapping has to supply by hand with a reason; none
is claimed for older JovianDSS 1 releases, which are not supported
([question 29](#decisions)).

Re-evaluate when a new recording changes the type of a read field, or when
the spec project changes its output format. An installation on a release
older than the minimum is outside what this document supports and does not
reopen it ([question 29](#decisions)).

## Risks
[Risks](#risks)

**Backward compatibility, preserved.** An installation on the recorded
JovianDSS 1 release would keep working as before: the recordings show v3 and v4
answer the same; the configuration keeps every key it has; the names and IDs of
existing volumes, snapshots and targets do not depend on the API version;
the seven type names and the accessors the drivers use, the 36 calls and the
request types stay; Basic auth stays, and tokens would come only with
JovianDSS 2 support ([JovianDSS 2 support](#future-work-joviandss-2-support));
nothing on the appliance is touched. Reverting the plugin image restores v3
and the old types. What the calls do on success would stay the same, with
these exceptions:

- the `per_page` commit would change how lists are paged ([Testing](#testing)
  step 4);
- [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) and [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) would make snapshot lists return real entries
  ([risk 4](#risks));
- stage 3 would remove the 202 branch of `CreateTarget` ([risk 6](#risks));
- stage 4 could make a call that works today fail: [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) would reject IDs it
  accepts now and refuse, at setup, an `iscsi.iqn` prefix outside the target
  rule; [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) would refuse an unverified certificate, a TLS version below 1.3
  ([risks 10 and 11](#risks)) or a `prot` other than `https`; and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) would
  refuse a reply over `maxreplysize`.

**Backward compatibility, not preserved.** These installations would stop
working after the upgrade until something changes: appliances on a JovianDSS 1
release without v4, or with a v4 that differs from the recorded one, until the
appliance is updated ([risk 2](#risks));
appliances that offer no TLS 1.3 ([risk 10](#risks)); appliances with a
self-signed certificate until `cafile` or `insecure` is set
([risk 11](#risks)); configurations whose `iscsi.iqn` prefix breaks the target
rule, until the prefix is changed ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids));
and configurations with `prot: http`, which the plugin no longer accepts
([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)). In
the code, the twelve lines that read `Creation` as a field, the list entry's
`Volume`, or the pool's `Available` as a number change in the same commit as
the types ([risk 1](#risks)).

1. **A caller still reads a field whose type or key changes.** Guard: `go build`
   and `go vet` would fail on any remaining use; the twelve lines are changed in
   the same commit ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)).
2. **Older releases.** All recordings come from current releases. **A JovianDSS 1
   release without v4, or with a different v4, would stop working, and there would
   be no fallback.** Guard: the docs and release notes would state the minimum
   release, `1.0up33.65410`. It is the latest JovianDSS 1 release, and both the lab
   and the hurl recordings run it (confirmed by the user). Older JovianDSS 1 releases
   do not serve the same v4 and are not supported: there is no backward
   compatibility with them (user decision, [question 29](#decisions)). The v4
   move would be its own commit, so it could be reverted on its own.
3. **The generator's rules are wrong, or the spec's output moves.** A widening
   rule could drop a field, for example in the zvol-or-clone merge; the spec
   repo's compiled YAML already holds a defect, a tab inside a `pattern:` value.
   Guard: [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) decodes every recorded jdss1 v4 reply with the generated types, and
   the generator is pinned to a spec commit and reads the spec output
   [question 31](#open-questions) picks.
4. **[bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) and [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) would change results, not only errors.** ListSnapshots
   would start returning real entries, which could expose paging bugs (the CSI list
   token) that are hidden today. Guard in this delivery: the fixtures of those list replies ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) and the implementer's tests on them, because
   csi-sanity cannot reach ListSnapshots until 0003 stops the controller crashes
   ([Testing](#testing) step 9); csi-sanity's ListSnapshots specs would follow after
   0003. [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) would likewise change which path a repeated
   request takes.
5. **Recorded data contains passwords and lab details** (names, IP addresses,
   serial numbers; `report.json` also holds certificates and host names). Guard:
   fixtures would be trimmed and their passwords (CHAP and REST admin, including the
   Basic `Authorization` header) removed before commit; no jdss2 reply would be a
   fixture, so no token would reach the repository
   ([questions 4 and 19](#decisions)); lab details may stay.
6. **Removing [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay)'s 202 branch.** If an appliance ever answered
   `CreateTarget` with 202, the call would fail instead of polling. 202 was never
   recorded, in about 21,000 requests per run. Guard: the fake-server test ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) would
   pin the new behaviour, and the error would name the status, so a real 202 would be
   noticed.
7. **One live test per driver for the whole delivery.** A regression could not be
   traced by comparing a live cycle per commit. Guard: every check that does not
   need the lab would pass first ([Testing](#testing) steps 1-3); each commit would
   build on its own; and `git bisect` with `dev_cycle.py` would find the commit if a
   live test failed.
8. **Connected clients are untested ([gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients)).** No recorded
   reply had an initiator logged in or a volume in use, and the iSCSI node code may
   not log in at all today (review `bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout`, `ISSUES.md` T10). The calls that act on a
   target or volume a client may still use have been seen only on idle resources,
   and never on v4 on the lab: detach, target delete, and the iSCSI deletes, which
   always send `force_umount: true` (`iscsi_driver.go:102`, `:240`, `:487`, `:516`).
   The v4 move and [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)'s change, which would go on to delete the
   target when the LUN is not attached, act exactly on these calls. If the appliance
   refused them while a session is active, unpublish or delete would fail and leave
   targets behind. If it accepted them, a forced delete could remove a volume that a
   node still uses, for example after a node crash: a data-safety question. Guard:
   [Testing](#testing) steps 6 and 7. Until they have run, the behaviour with a
   client connected is unknown, on v3 as on v4.
9. **A read field absent from a reply.** Go's `encoding/json` fails only on a
   wrong type: a missing key or `null` leaves the field at its zero value without
   an error. Cases already seen: [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) (no `volume`
   key, so every snapshot entry is skipped and ListSnapshots is empty),
   [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) (zvol snapshot without `name`),
   [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) (NAS `quota` is `null` without a quota, so the size is 0;
   review `bug_plugin_driver_nas_quota_null_size_zero`) and [bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors) (a pool `available` that does not parse
   becomes 0; [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors) would fix this one). Unguarded, a later release that dropped or
   renamed a read field could make GetCapacity report no free space or a size of
   0, lists come back empty without an error, an empty name go into the next
   request's path (see also [security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments)), or a snapshot `creation` of
   year 1 reach Kubernetes. Guard in this delivery: the mapping's
   required-or-optional rule and the generated `checkRequired()` ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), a required
   field removed from its fixture failing by name ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)), and raw JSON for everything the plugin does not
   read, which cannot fail. What stays: a read field the recordings never show
   set is typed by hand until a recording confirms it
   ([missing fields in replies](#future-work-missing-fields-in-replies)).
10. **TLS 1.3 only ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)).** Which TLS versions the appliances offer is not
    recorded: no recorded run holds a TLS version string (checked). If an
    appliance, or a supported JovianDSS 1 release, offered only TLS 1.2, the
    plugin could not connect at all. Guard: a TLS handshake check
    (`openssl s_client -tls1_3`) on the lab appliance and jdss1 **before
    the series starts**, so a 1.2-only appliance is known before code is written
    (on the hurl appliances only when no hurl tests are running). There is no
    `tlsmin` setting (user decision, [question 26](#decisions)): a
    1.2-only appliance would need a code change.
11. **Certificate checks on by default ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)) change upgrades.** The recorded
    appliance certificate is self-signed, so an existing installation would stop
    connecting after the upgrade until it sets `cafile` or `insecure`. Guard: the
    release notes say so, with both settings, and the error names the certificate
    problem and the two settings. The lab's live test needs one of them set.
12. **No request deadline ([bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied)).** The request is built
    without the CSI call's context, so a hung appliance call is not cancelled when
    the CSI deadline passes; it lasts until `idletimeout`, which is used as the
    timeout of the whole request. Guard: none in this delivery; future work
    ([request deadlines](#future-work-request-deadlines)).
13. **Names of exactly 132 characters ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)).** A volume with such a name, made
    from a CSI name of exactly 129 characters, would be looked up under its `vs_`
    name after the change and not be found. Kubernetes volume names have 40
    characters, so none is expected. Guard: none; the edge is accepted with the
    limit of 131 (user decision, [question 21](#decisions)), which may
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
    This document leaves the logic by which the driver acts on a reply as it is: it
    changes what the client sends and how it decodes and maps what comes back.
    Today that logic takes a reply as the current state: `cleanIntermediateSnapshots`
    deletes the intermediate snapshots one list shows and keeps the rest
    (`iscsi_driver.go:185`, `nfs_driver.go:291`), and ListSnapshots reports what
    one list read returns. A read moments after a change could therefore miss an
    object that exists or show one that is gone, and a step taken on it would be
    a false operation: an intermediate snapshot left behind, a snapshot reported
    to Kubernetes as missing, a field read as empty that the next read would
    fill. The exposure is small on JovianDSS 1; JovianDSS 2, outside this document,
    shows the lag most ([JovianDSS 2 support](#future-work-joviandss-2-support)). Guard: none in this
    delivery. Every later piece of driver work, whether a design, an audit or a
    fix, would have to treat a reply as possibly behind the request: read by name
    where the name is known, retry a list that lacks an expected entry before
    concluding that it is absent, and take no destructive step on the strength
    of a single list read straight after a change.

15. **Some requests have never been seen on jdss1 v4.** The gap run on jdss1
   v4 was lost to an outage, so the `{}` body the plugin sends with every GET,
   among others, is recorded on jdss1 v3 and jdss2 v4 only. Guard: the live
   tests ([Testing](#testing)
   steps 4 and 5) run every call on v4, and the path move being its own
   commit, `git bisect` can tell its effect from a fix's.
16. **A read field was never seen set, so the recordings give it no type.**
   Guard: the mapping may state the type by hand, with the reason, and the
   generator refuses a read field it cannot type otherwise; the live test
   covers the state when it occurs.
17. **The generated file drifts from the spec** because someone edits it or
   forgets to regenerate. Guard: the header names the spec commit;
   `make check-types` regenerates and diffs, and fails on a difference.
18. **A widening rule is wrong**, for example the zvol-or-clone merge drops a
   field. Guard: every recorded jdss1 v4 reply decodes with the generated
   types, and a required read field removed from its fixture fails by name
   ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).
19. **The spec output format moves.** The YAML defect shows it. Guard: the
   generator is pinned to a spec commit and reads one spec output
   ([question 31](#open-questions)); a format change is a generator change,
   not a types change.
20. **The procedure depends on one person.** Only the user may run the spec
   repository's tools, and only a machine with the recordings can. Guard:
   the committed generated file; the procedure written down in the plugin's
   documentation ([question 33](#open-questions)).
21. **Per-endpoint types diverge for one accessor**, such as `creation` as
   `int64` in a list and a string by name. Guard: conversions live in the
   methods file, per type, as the decoders do today, and the fixtures cover
   both endpoints.
22. **A later JovianDSS release changes v4.** Guard: regeneration and the
   audit on the new recordings show the change as a diff or a failed check.

## Affected systems and files
[Affected_systems_and_files](#affected-systems-and-files)

| System | Effect of this document | Where |
|---|---|---|
| Plugin REST client, paths | one API prefix, `api/v4`, replaces the 34 hard-coded `api/v3` strings (30 after the dead-code commit); comments and constants that name v3 go | `pkg/rest/*.go` (Table 3) |
| Plugin REST client, reply types | `types.go` replaced by the generated types and the methods file; the mapping added | `pkg/rest/types_gen.go` (generated), `pkg/rest/types_methods.go`, `pkg/rest/types_gen.yaml` |
| Generator and build | the generator script; `gen-types` and `check-types` targets | `tools/rest-types-gen.py`, `Makefile` |
| Fixtures, tests and audit | fixtures of every recorded jdss1 v4 reply, which the implementing agent's tests decode with the generated types and check field by required field; the audit compares the plugin's calls with the v4 index and keeps no type check | `pkg/rest/*_test.go`, `pkg/rest/testdata/`; `kub-testing/tools/rest-type-audit.py` ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures), [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) |
| Drivers and controller | the twelve lines: `Creation` via accessor, `Parent.Name` for list entries, `GetAvailable()` for the pool | `pkg/controller/listers.go:52`, `:54`, `:56`, `:82`, `:124`, `:148`; `pkg/controller/controllerplugin.go:679`, `:802`, `:1017`; `pkg/driver/restwrappers.go:40`; `pkg/driver/iscsi_driver.go:456`; `pkg/driver/nfs_driver.go:591` |
| Plugin documentation and release notes | the required JovianDSS release is stated: one that serves API v4, for JovianDSS 1 at least `1.0up33.65410`; the regeneration procedure ([question 33](#open-questions)) | `docs/`, `CHANGELOG.md`, `pkg/rest/doc.go` |
| JovianDSS appliances | must serve API v4: JovianDSS 1 at `1.0up33.65410` or later | the lab appliance and the recorded appliances already do |
| Test lab | the live cycles become the plugin's first runs on v4; the suite itself does not change | kub-testing `dev_cycle.py` cycles ([Testing](#testing) steps 4 and 5); `ISSUES.md` T9 |
| Spec repository | read-only input: the compiled OpenAPI for jdss1 v4 and its manifest; the YAML defect reported | `jdss-rest-api-spec/empirical/pilot/tools/tsp/b_jdss1_v4/out/openapi/` |
| Not affected | the node DaemonSet manifest; the node's mount and iSCSI code | |

The files, one row each, with the stage that changes them:

Rows marked *NFS* are NFS-only code, covered by the NFS cycle ([Testing](#testing)
step 5) and, where it is `pkg/rest` code, by unit tests.

**Table 3 — Files that would change** · `tbl_files_changed`

| File | Change | Stage |
|---|---|---|
| `pkg/rest/entrielisters.go` (`pagedcSuffix`), `pkg/driver/common.go` (`getResourcesList` stop rule) | [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) `per_page` | before 1 |
| `pkg/rest/storageapi.go`, `rest.go`, `rproxy.go`, `payload-def.go`, `reqtypes.go`, `nas_volumes_snapshots.go`, `shares.go`, `targets.go`; `pkg/driver/nfs_driver.go:322` | [impl_plugin_rest_remove_dead_code_before_v4_move](#impl_plugin_rest_remove_dead_code_before_v4_move), [dead code](#dead-code); [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) and [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check) (switch to `DeleteNASSnapshot`) | before 1 |
| every `pkg/rest/*.go` with a path | one API prefix, `api/v4` | 1 |
| plugin docs, `CHANGELOG.md` | v3 deprecated, minimum release `1.0up33.65410` | 1 |
| `pkg/rest/types.go`, replaced by `types_gen.go` (generated) and `types_methods.go`; `types_gen.yaml` (new) | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types): generated reply types; closes [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) *NFS* and [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead); [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) *NFS* key (stage 3 filters by it); [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors) | 2, 3 |
| `tools/rest-types-gen.py` (new), `Makefile` (`gen-types`, `check-types`) | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types): the generator and its targets | 2 |
| `pkg/controller/listers.go`, `controllerplugin.go`; `pkg/driver/restwrappers.go`, `iscsi_driver.go`, `nfs_driver.go` | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types): the twelve lines that read `Creation`, the list entry's `Volume`, the pool's `Available` | 2 |
| `pkg/rest/*_test.go`, `pkg/rest/testdata/` (new) | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) | 2, 3, 4 |
| `Makefile` | `vet` and `test` targets | first commit |
| `kub-testing/tools/rest-type-audit.py` (new) | [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool): compliance matrix, request keys, error coverage; no type check | 2 |
| `pkg/rest/doc.go` (new) | [impl_plugin_rest_package_documentation](#impl_plugin_rest_package_documentation), package documentation | 2 |
| `pkg/rest/entrielisters.go` | [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) (a new call, `GET pools/<x>/volumes/<x>/snapshots`, in the [compliance matrix](#compliance-matrix)), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) *NFS* | 3 |
| `pkg/rest/errors.go`, `targets.go` | [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) *NFS*, [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised) | 3 |
| `pkg/rest/errors.go`, `storageapi.go` (`getError`); `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/*.go` (callers) | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue): error catalogue, one handler and log line, codes per kind; callers decide on the code only (`iscsi_driver.go:520` included) | 3 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go` (`DeleteVolume`, `cleanIntermediateSnapshots`) | review [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children) (cleanup also on `volume has children`, with [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)), [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) (listing error first; NFS takes NAS snapshots, *NFS*) | 3 |
| `pkg/rest/storageapi.go` (`getError`), `clones.go`, `pools.go`, `targets.go`, `nas_volumes.go`, `nas_volumes_snapshots.go`, `nas_volumes_snapshots_clones.go`, `shares.go`, `entrielisters.go` | [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status) and [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check): status checked before decoding, non-JSON bodies mapped by status | 3 |
| `pkg/rest/rproxy.go`, `errors.go`, `rest.go` | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes): [crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error), [crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check), [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body), [crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error), [crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs); [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) in `rproxy.go`: no request logging (`security_plugin_secrets_in_logs`, `security_plugin_rest_request_dump_every_call`) | 4 |
| every `pkg/rest/*.go` with a path builder | [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments): `url.PathEscape` on every path segment ([security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments)) | 4 |
| `pkg/driver/lunid.go` (`NewVolumeDescFromName` threshold 240 → 128), `snapid.go`; `pkg/controller/controllerplugin.go` (`iscsi.iqn` prefix check); `pkg/rest` path builders | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids): ID and name validation, generated names at most 131 characters, target prefix checked | 4 |
| `pkg/rest/rproxy.go`, `pkg/common/common.go` (`RestEndpointCfg`, with `yaml` tags), `docs/configuration.md` (new keys; `pool` and `logfile` corrected; `tries`, extra `addrs` and `prot` marked unused), `docs/install.md`, `docs/install-nfs.md`, the controller manifests (the CA file in the config Secret) | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): HTTPS only, certificate checks, `insecure`, `cafile`, TLS 1.3, no redirects; [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit): `maxreplysize` | 4 |
| `pkg/rest/storageapi.go:51`, `:53`, `errors.go:173`, `:265`, `targets.go:150`, `:184` | [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply): `"%s"` formats, so `go vet` passes | first commit |
| `pkg/controller/controllerplugin.go`, `pkg/node/publish.go` | [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs): no whole CSI requests in logs | 4 |
| `CHANGELOG.md` | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): certificate checks on by default, `cafile`/`insecure` needed for self-signed certificates | 4 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/listers.go` | callers of [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) *NFS*, [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) *NFS*; the filter by kind ([bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key), [question 7](#decisions)) | 2, 3 |
| `review/backlog.md` | item 36 addressed; item 37 closed (zero-padded format checked against recorded snapshots); items 19 and 20 addressed by [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile); [impl_plugin_rest_request_body_keys_checked_against_schema](#impl_plugin_rest_request_body_keys_checked_against_schema) findings added | all |
| `docs/adr/0002-api-v4-compliance-jdss1.md` | this document: status, tense and the decisions as they are taken | all |

JovianDSS 2 support: a later version, with its own design.

## Testing
[Testing](#testing)

The whole delivery would be tested once. Everything that does not need the lab would
run first and would have to pass; then one live cycle per driver, iSCSI and NFS.
Every appliance used for
testing has to run the latest JovianDSS release; one that does not is updated
before the tests ([question 11](#decisions)). Steps 1 and 4 to 8 run on kub-gate,
which reaches the lab, and are the user's to run; step 2 runs on the analysis
machine and is the agent's; the lab's results come back as the cycle folder under
`kub-testing/tmp/cycles/`. The unit tests of `pkg/rest` are the implementing
agent's: the items say what they must show, and this section only runs them.

**Before the live test**

1. **Build, vet and unit tests (kub-gate, the user).** Proves that every commit of
   the series builds on its own, so that `git bisect` stays usable, and that vet
   and the package tests pass. First, `go vet ./...` on the current code, with the
   result recorded: no vet result exists for it; the
   [first commit](#first-commit-make-go-vet-pass) removes the six format-string
   reports that are known, and anything else vet reports is decided then. Then
   `go build ./...` and `go vet ./...` (`make vet`) on every commit of the series,
   and `go test ./pkg/rest/...` (`make test`) on the last one; all would have to
   pass. The driver and controller code this document changes has no automated
   check in this delivery (user decision): [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)'s name and ID rules
   and its `iscsi.iqn` prefix check, [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)'s request logs, the delete cleanup of
   [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children) and [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error), the filter by kind ([question 7](#decisions)) and the
   NFS cleanup reaching [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)'s call. kub-testing holds them as the
   seven checks of T11 (`testcases/plugin-checks-design-0002.md`) for later; the
   live cycles (steps 4 and 5) exercise them as far as csi-sanity reaches them.
2. **Audit and compliance (analysis machine, the agent).** Proves that every call
   the plugin makes meets the rules on jdss1 v4 and that the committed types match
   the spec. Needs the spec repository refreshed by the user with
   `build_typespec.py --check`, and the audit tool of [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool),
   which does not exist yet and takes `empirical/jdss1/v4/INDEX.md`, the company
   `-src` schemas and the plugin's list of calls. The audit would
   show [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_request_body_valid](#rule_rest_request_body_valid), [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted), [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) and [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) passing for every call on jdss1 v4, and no recorded error outside the catalogue; [rule_rest_reply_types_decode](#rule_rest_reply_types_decode) holds by construction of
   the generated types: `make check-types` produces no diff, run here because
   the spec repo is on the analysis machine only, which has no Go, so the target
   has to regenerate and diff with Python alone; step 1 decodes every
   fixture. [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed) has no pass or fail; it is a
   list of differences. Re-run after each new hurl run, with `make gen-types`.
3. **NFS code.** Folded into step 5, which runs the NFS findings live; their unit
   tests are the implementing agent's, and the NFS part of review
   [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) is driver code, a T11 check
   exercised live by step 5. The number stays, so that "step 5" keeps its meaning.

**Live test**

4. **iSCSI cycle (kub-gate, the user): `dev_cycle.py all --compare` with the
   baseline cycle `20261102-140341`,** spec by spec, the changes in the cycle's
   `summary.txt`; with `cafile` or `insecure` set for the lab appliance in the
   config Secret `jdss-controller-cfg`, because its certificate is self-signed
   ([risk 11](#risks)). Proves that the plugin works on v4 against the lab
   appliance as well as it did on v3, and answers [question 30](#open-questions)
   and [risk 15](#risks), the calls never recorded on jdss1 v4. Three
   prerequisites. The baseline build carried the user's
   GetCapacity fix (`crash_plugin_ctrl_get_capacity_no_logger`, 0003's D3), which is not on `master` yet, so **`crash_plugin_ctrl_get_capacity_no_logger` has to
   be committed before the series** (`ISSUES.md`, next action 3); without it the
   ctrl-cap preset would crash again and the comparison would be void. The
   plugin runs at **Debug** level for the cycle, which the `per_page` check below
   needs; that is safe only because [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) are in the build (until then Debug
   logs secrets, which is why kub-testing forbids it today). And the TLS
   handshake check of [risk 10](#risks), `openssl s_client -tls1_3` against the
   lab appliance, has been made before the series started, so that a 1.2-only
   appliance is known before code is written. It would pass when:
   - no `plugin.log` contains a decode error: Go's `cannot unmarshal`, the
     required-field error of `checkRequired()` or a conversion error of
     `types_methods.go`, whose texts the items fix;
   - the ctrl-A ListVolumes/GetVolume specs and the ctrl-cap GetCapacity spec reach
     the plugin's real logic: the request log shows their `GET` answered 200, and
     a spec that still fails does not fail with `Internal` and a decode message
     (those failures would become new findings);
   - the request log shows `per_page` on every list call, and no list call asks for
     the same page twice;
   - every GET and every DELETE in the request log is answered as the
     [compliance matrix](#compliance-matrix) records, none with 400 or 415, so the
     `{}` body made no difference on jdss1 v4 either ([question 30](#open-questions));
   - a well-formed ID of a missing volume still gets `NotFound` and a malformed one
     `InvalidArgument`, as the baseline cycle shows for csi-sanity's own IDs
     ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids));
   - no spec that passed in the baseline fails.
5. **NFS cycle (kub-gate, the user), requirement.** The NFS build would replace the
   iSCSI build in the lab for one `dev_cycle.py all` cycle, then the iSCSI build
   would come back the same way. The
   two cannot be installed side by side today, because their manifests use the
   same names and socket path (`ISSUES.md` MF7). The tooling cannot do this yet
   either: `sanity_runner/cycle.py:110-111` refuses any `-nfs` image, `lab.py:32`
   (in `tests/csi-sanity/`) lists iSCSI as the only tested protocol, the
   `Makefile` never produces the image name the NFS manifests pull
   (`opene/joviandss-csi-nfs`), the NFS manifests have never been applied
   anywhere, and the result comparison takes the latest cycle whatever its
   protocol, so it would compare the NFS cycle with an iSCSI one. Those are
   prerequisites of this
   step, outside this document's commits ([Open question 22](#open-questions);
   recorded in kub-testing as TOOL10 and MK11).
   There is no NFS baseline: no NFS build has ever been run live, so this cycle
   becomes the first one. It would pass when:
   - no `plugin.log` contains a decode error (as in step 4), and the request log
     shows `per_page` on every list call;
   - no spec fails because of a finding this document fixes, in particular
     ListSnapshots by volume ([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)), deleting a volume that has clones
     ([bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised)), a repeated CreateVolume
     ([bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised)), deleting a clone that is already gone
     ([bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised)), and deleting a volume after a clone from it
     ([bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error), [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete));
   - every other failure is recorded as a new finding in kub-testing's
     `ISSUES.md`, with a name registered in `issues/test-findings.md`.

   Specs that reach NFS code only after 0003 stops the controller crashes have no
   live cover until then (step 9). If the NFS node path does
   not work in the lab at all, that is recorded as a result, not passed over.
6. **Connected clients (the lab, the user; [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients)), requirement.** csi-sanity's
   node preset alone may not log an initiator in: the iSCSI NodeStage on the current
   code probably reports success without logging in or mounting (review `bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout`).
   So the live test would also include one real iSCSI PVC used by a pod on the lab
   (`ISSUES.md` T10):
   - while the pod runs, `iscsiadm -m session` on its node shows a session to the
     volume's target, and the volume is mounted;
   - after the pod and then the PVC are deleted, the target and the LUN are gone on
     the appliance: `GET pools/<pool>/san/iscsi/targets/<name>` answers 404
     ([bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)), and the controller's `plugin.log` at Debug shows no failed
     detach, target delete or volume delete.

   The replies to these calls would be kept with the cycle folder as the first v4
   evidence with a client connected; adding such a case to the hurl suite is the
   user's. If no session appears (`bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout` confirmed), [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) stays untested; that is
   recorded as a result, not passed over.
7. **Connected clients, recommendation (the lab, the user):** on a throwaway volume, also send
   ControllerUnpublish and then DeleteVolume once while the initiator is still logged
   in, as after a node that did not clean up, and record the replies. This is the
   case where unpublish and `force_umount: true` matter most ([risk 8](#risks)).
   Kubernetes does not delete a bound PVC under a running pod, so the two calls are
   made directly: with a CSI command-line client against the controller socket, or
   by freeing the volume from the pod and keeping the iSCSI session open by hand on
   the node; the tester records which. No pass condition: the replies are evidence
   for risk 8 either way.

Steps 4 to 7 run on the lab appliance, which runs the latest release like every
test appliance.

8. **If a cycle failed:** find the commit with `git bisect` and `dev_cycle.py`, and
   fix it before the delivery is finished, not as a follow-up.

**Limits**

9. csi-sanity reaches [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) only after 0003
   stops the controller crashes; until then they have no live cover in this delivery.
10. **Not testable here:** [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) beyond steps 6 and 7 (several nodes, a pool moving
    between nodes), [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call), JovianDSS 2 live, which needs
    [JovianDSS 2 support](#future-work-joviandss-2-support) (a later version) and a
    jdss2 pool, and `maxreplysize` ([impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit)), which the
    lab's replies, about 66 KB at most against a default of 16 MiB, never reach.

## Open questions
[Open_questions](#open-questions)

Open: 22, 25 and 30 to 33. The numbers are kept, gaps included, so that "question 27" means the same thing everywhere; the decided questions follow as a table, each with the section that argues it.

- **22. Who prepares the NFS live cycle?** [Testing](#testing) step 5 lists the
  five things it needs that do not exist, from the tooling that refuses an
  `-nfs` image to the image name nothing builds; kub-testing records them as
  TOOL10 and MK11 in its `ISSUES.md`. Recommendation: a small `Makefile` and
  manifest commit on `master` before this series for the image name; none of
  it in this document's commits. The owner is still to be named.
- **25. Which of `crash_plugin_driver_snapshot_name_empty`, `crash_plugin_driver_id_index_without_length_check`, `crash_plugin_driver_volume_desc_empty_value` and `crash_plugin_ctrl_create_snapshot_empty_name` does [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) close?** [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) changes the ID parsers
  in `lunid.go` and `snapid.go`, next to four review findings that no design
  claims: `crash_plugin_driver_snapshot_name_empty` and `crash_plugin_driver_volume_desc_empty_value` (`SnapshotDesc.Name()`, `SDS()`, `CSIID()` and
  `VolumeDesc.VDS()`, `CSIID()` panic on an empty value, `snapid.go:147-166`,
  `lunid.go:159-173`), `crash_plugin_driver_id_index_without_length_check` (`IsVDS`/`IsSDS` index `[0]` without a length check,
  `lunid.go:104`, `snapid.go:27`) and `crash_plugin_ctrl_create_snapshot_empty_name` (CreateSnapshot with an empty name
  creates `sb_` on the appliance, then `Name()` panics). Reading the code: `crash_plugin_driver_id_index_without_length_check`
  is not about IDs: all sixteen callers pass names that came back from the
  appliance (`controller/listers.go:195-251`, the snapshot and clone cleanups in
  both drivers), so [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) does not touch it. `crash_plugin_driver_snapshot_name_empty` and `crash_plugin_driver_volume_desc_empty_value` are panics inside
  accessors that [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) does not change. `crash_plugin_ctrl_create_snapshot_empty_name` would be closed only if [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) also
  rejected an empty name in `NewSnapshotDescFromName` (`snapid.go:30`), which
  has no such check today, while `NewVolumeDescFromName` has one
  (`lunid.go:61-63`). Recommendation: add that one check to [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) and record `crash_plugin_ctrl_create_snapshot_empty_name`
  as closed by it; record `crash_plugin_driver_snapshot_name_empty`, `crash_plugin_driver_id_index_without_length_check` and `crash_plugin_driver_volume_desc_empty_value` as 0003's in `ISSUES.md`.
- **30. Does the `{}` body the plugin sends with every GET matter on jdss1 v4?**
  It made no difference on jdss1 v3 and jdss2 v4; jdss1 v4 was never
  recorded. The live test answers it.
- **31. Which spec output does the generator read?** The compiled YAML has the
  tab defect; the TypeSpec emitter can write JSON, or the generator can read
  the `.tsp` files, which are regular but need a parser. Recommendation: the
  JSON emit, agreed with the spec project.
- **32. Which read fields are optional?** First list from the current decoders
  and accessors: `origin`, `clones`, NAS `quota`; the mapping decides per
  field, and the required-field tests pin it.
- **33. Where does the regeneration procedure live?** In the plugin's
  documentation, since the generated file is the plugin's, with the
  kub-testing notes pointing at it.

### Decisions
[Decisions](#decisions)

| # | Question | Decision | Where it is argued |
|---|---|---|---|
| 1 | Which API? | v4 only; v3 deprecated; no other option was considered. | [Option A](#option-a-v4-only-the-only-option-considered) |
| 2 | Minimum JovianDSS 1 release? | `1.0up33.65410`, the release of the lab and the recordings. | [Decision](#decision) |
| 3 | When does token login come? | with JovianDSS 2 support, a later design. | [Future work: JovianDSS 2](#future-work-joviandss-2-support) |
| 4 | Fixtures: real replies or hand-written? | real replies from the recordings, trimmed, passwords and the `Authorization` header removed. | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) |
| 5 | Unused fields with the wrong type? | removed at first; the same question as 27, both superseded by 28: kept as raw JSON that cannot fail. | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types) |
| 6 | Group the type fixes, or the two seen live first? | all together, delivered and tested in one go. | [What changes](#what-changes) |
| 7 | Should the list functions filter by kind? | yes: each driver drops the other kind's entries; other values are kept. | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 8 | Who closes the evidence gaps on the appliance side, the first eight under [Evidence gaps](#evidence-gaps)? | the user adds the missing requests to the hurl suite; done for most, see the gaps. | [Evidence gaps](#evidence-gaps) |
| 9 | `sparse` default? | sent on every create: StorageClass parameter, else the plugin default, else sparse. | [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) |
| 10 | Which commit carries the NAS snapshot delete switch? | the dead-code commit. | [the dead-code commit](#before-stage-1-remove-dead-rest-code) |
| 11 | Is the lab appliance hurl's `jdss1`? | not a design question; every test appliance runs the latest release. | [Testing](#testing) |
| 12 | Where do the REST-layer crashes go? | every crash fixable inside `pkg/rest` is fixed here, in stage 4. | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes) |
| 13 | Commit hash in the body? | no; the body says "the current code", Metadata names the commit. | [Metadata](#metadata) |
| 14 | Keep the note that this replaces the original 0002? | no; history. | [Metadata](#metadata) |
| 15 | Tense? | the conditional for what would change; the present for today's code and the evidence. | [Metadata](#metadata) |
| 16 | Where do the TLS and path-escaping findings go? | this delivery, stage 4: in the TLS item and the path-escaping item. | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 17 | Where do the delete-cleanup findings go? | this delivery, stage 3: one inside the error catalogue, one as its own fix there. | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 18 | Where does the caller that turns every clone-delete error into Busy go? | this delivery, stage 3, in the error catalogue. | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |
| 19 | jdss2 tokens in fixtures? | tokens are secrets: never logged, never in a fixture; no jdss2 reply is a fixture now. | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) |
| 20 | Exact rules for the name check? | the appliance's rules from the strings tests: `a-z A-Z 0-9 _ -`, 131 characters, lower-case targets. | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) |
| 21 | Cap generated names at 131 or 132? | 131; a volume of exactly 132 characters is the accepted edge (risk 13). | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) |
| 23 | Is `prot: http` still allowed? | no; HTTPS only, `prot` other than `https` is a configuration error. | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 24 | Where do release notes live? | `CHANGELOG.md` at the repository root. | [Stage 1](#stage-1-move-to-api-v4) |
| 26 | TLS minimum as a setting? | no; 1.3 fixed, a parameter of `pkg/rest` at most, no `tlsmin` key. | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 27 | Keep or remove the unread fields that fail? | remove; superseded by 28. | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types) |
| 28 | Generate the reply types instead of fixing them by hand? | yes, from the recorded jdss1 v4 API, same names, only read fields can fail. Weighed against: the rule per field (27), a transcription that drifts again with the next release and only the audit tells; every field modelled by hand, which fails on a field nobody reads and drifts the same way; the whole client generated, which replaces about 2,100 lines and every caller and redoes the stage 3 and 4 work on generated code; tolerant decoding, which lets a real change in a read field pass silently. Price: about 650 lines of generator and mapping to keep, a procedure that depends on the user running the spec repository's tools, a spec output format still moving. | [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types) |
| 29 | Do older JovianDSS 1 releases serve the same v4? | no: they do not serve the same API v4, and there is no backward compatibility with them; nothing older than `1.0up33.65410` is supported. | [Risk 2](#risks) |

## Relationship to other designs
[Relationship_to_other_designs](#relationship-to-other-designs)

- Findings are tracked in `kub-testing/ISSUES.md` and `kub-testing/project-status.md`;
  the review files and ID prefixes are in [ID sources](#id-sources).
- Spec repo `jdss-rest-api-spec`: the reference for compliance
  ([The `jdss-rest-api-spec` repository](#the-jdss-rest-api-spec-repository)).
  Only read; its TypeSpec of the observed types, compiled to OpenAPI, is the
  source of the generated reply types ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), refreshed by the user with
  `build_typespec.py --check`. Also to report: the compiled YAML for jdss2 v4
  holds a literal tab inside a `pattern:` value (line 15389), which Python's YAML
  parser refuses; and a name with non-ASCII characters gets **500
  `UnicodeDecodeError`** from the appliance (zvol, NAS and snapshot strings tests)
  instead of a 400 validation error, which is a server bug. The [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed) differences are
  a list for the user to pass on to the spec
  project; this work would not change it.
- JovianDSS-Proxmox plugin: its `jdssc` already has typed errors per kind
  (`jdssc/jdssc/jovian_common/exception.py`) and one general error path
  (`jovian_common/rest.py`, `_general_error`); [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) follows the same idea.
- hurl suite (`jdss-rest-api-spec/empirical/`): it plans to report the bugs it finds
  in API clients such as this plugin; the findings its recordings revealed (F9 to F28 in `kub-testing/issues/test-findings.md`, from [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) to [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) in that file's order), [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check) and
  [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) are candidates for that list. Its Finding 4.4 (a delete that
  passes the create status constant) can be downgraded, because the constant is
  ignored ([Key observation 7](#key-observations)). Its open question on `ListAllSnapshots` (does the
  plugin drop the other kind of snapshot from the combined list?) is answered by
  [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key): today no plugin code filters by kind,
  only the name-format check in the list functions can drop an entry; stage 3
  would filter by kind ([question 7](#decisions)).
- Backlog (`review/backlog.md`): item 36 is [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch); item 37 (snapshot
  time format) has been verified correct and can be closed; items 19 and 20 (the
  configuration keys the guide gets wrong, and the missing `yaml` tags) are
  addressed by [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile).
- [0003 controller crash fixes](../design/0003-controller-crash-fixes-design.md): deferred. Its
  design item D2 (`crash_plugin_ctrl_create_volume_on_rest_error`) would remove the crash that [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) causes in
  CreateVolume. With 0002 alone, [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch)'s error would go away, so [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) would no longer
  trigger `crash_plugin_ctrl_create_volume_on_rest_error`. The REST-layer [crashes](#crashes) are fixed here ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)); 0003 keeps the
  controller's own crashes and the startup check of `SetupEndpoint`'s error ([crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error)).
  0003's D3 (the GetCapacity logger fix, `crash_plugin_ctrl_get_capacity_no_logger`) exists on kub-gate uncommitted and
  has to be committed on `master` before this series, because the baseline cycle
  was built with it ([Testing](#testing) step 4).
- `sparse` provisioning ([bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse)): sparse by default,
  configurable in the plugin configuration and per StorageClass; its own design
  ([question 9](#decisions)).
- REST client behaviour: not a compliance matter. Of its catalogue
  ([REST client behaviour](#rest-client-behaviour)) this delivery leaves
  [bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover) and
  [bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied) to a later REST client
  design; the other three entries are fixed here.
- Security: every [security finding](#security-findings) is fixed in this delivery:
  [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments), [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids), [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) in [stage 4](#stage-4-robustness-and-security), [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) in the
  [first commit](#first-commit-make-go-vet-pass).
- Before any of this can be built, the plugin checkout on kub-gate has to become a
  real clone of `master`: today it is an old commit with the merged code copied on
  top, so builds carry the wrong commit ID and the analysis machine cannot see what
  kub-gate builds (`ISSUES.md` INF8).
- JovianDSS 2 is outside this document; a later design covers it.
- The old letter-and-number IDs of the reviews and the tracker are mapped to the names used here in `kub-testing/issues/`.

### Future work: JovianDSS 2 support
[Future_work_JovianDSS_2_support](#future-work-joviandss-2-support)

**Not part of this document** (user decision): JovianDSS 2 support would come in a
later plugin version and be discussed in its own design. What that design would
have to cover, as known today: JovianDSS 2 has no v3 and accepts only token login
(Auth row of [API versions](#api-versions)), so v4 alone would not make the plugin
work on it. A config setting would choose Basic or token. With tokens the plugin
would have to log in, refresh the token before it expires (900 s) or on a 401, cope
with refresh tokens that work only once, and stay correct when several requests run
at the same time. It touches `rproxy.go`, the configuration and secrets. This document
already stops request logging ([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging)). Refresh tokens must never be logged (user
decision, [question 19](#decisions)), and neither must access tokens: any
logging the token design adds must keep both out of the logs
(`security_plugin_secrets_in_logs`, `security_plugin_rest_request_dump_every_call`).
Testing would need a jdss2 pool. This document does not touch JovianDSS 2; it
prepares for it only by giving 401 its own error code ([bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status)) and
by keeping the jdss2 recordings beside the jdss1 ones as evidence. That design would also have to allow for replies that are
behind the request, which jdss2 shows most ([risk 14](#risks)).

### Future work: request deadlines
[Future_work_request_deadlines](#future-work-request-deadlines)

Not part of this delivery; for the later REST client design. It addresses
[bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied) and [risk 12](#risks). Every CSI call arrives with a
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
which today allows up to 243 characters) would have to stay within it. [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s
catalogue would need the `ZfsCmdError` message for this case, `name is too long`.

### Future work: missing fields in replies
[Future_work_missing_fields_in_replies](#future-work-missing-fields-in-replies)

Mostly done by [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types) ([question 28](#decisions)), which addresses
[risk 9](#risks) for the fields the plugin reads: each is marked required or
optional in the mapping, a required one absent from a reply fails the call with
an error naming the field (`checkRequired()`), an optional one has a documented
absent value in its accessor, and a required field removed from its fixture
fails by name ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)). Fields the plugin does not read are raw JSON and
cannot fail.

**The problem it solves.** Go's `encoding/json` fails only on a wrong type or
broken JSON. A missing key, or `null`, leaves the field at its zero value (`""`,
`0`, `false`, `nil`) without an error, so the code cannot tell "the appliance
sent 0" from "the appliance sent nothing". Without the rule, a reply that lacks
a field the plugin reads would pass silently.

**What stays for later:** a read field the recordings never show set is typed by
hand in the mapping until a recording confirms it; the audit ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) could report
every read field that is missing or `null` in any recording, so that the
required-or-optional list rests on evidence rather than on reading the code.

## Advice
[Advice](#advice)

- hurl suite, Finding 9 (2026-11-04): JovianDSS 2 serves v4 only and accepts
  only tokens.
- hurl suite, Findings 14 and 15 (2026-11-04): v3 and v4 identical on
  JovianDSS 1 in 363 and 281 cases.
- kub-testing `ISSUES.md` [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) (confirmed 2026-10-27 and 2026-11-01) and [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch)
  (confirmed 2026-11-02): the two type failures seen live.
- Simulation over the hurl recordings (2026-11-05): [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) predicted; no
  other mismatch in any call the plugin makes.
- The first version of this document (2026-11-05): it listed "keep v3" and
  "support both with a setting" as options, which the user ruled out as never
  real, v4 being the only option; the rule per field; generation set aside
  because the spec format "is not agreed yet".
- kub-testing `ISSUES.md`, DEC10 (2026-11-05): the v4 decision as recorded
  for the test lab, with the minimum release.
- jdss-rest-api-spec `empirical/README.md` and `hurl/build_typespec.py`
  (read 2026-11-10): the empirical TypeSpec, its rules and `--check`; the
  compiled OpenAPI present for all four trees.
- User decisions: v4 only (2026-11-05); remove the unread mismatching fields
  (2026-11-09, question 27, superseded); generate the types, names kept,
  jdss1 v4 only, only read fields can fail with optional where a value may be
  absent, drivers updated (2026-11-10); one record for the whole problem
  (2026-11-10).
- User decisions (2026-11-10): JovianDSS 2 is not touched by this document;
  stages 3 and 4 belong to the same migration, so one record
  covers the whole migration and is implemented as one series; its
  questions are kept here with their numbers.
- User decision (2026-11-11): older JovianDSS 1 releases do not serve the
  same API v4; there is no backward compatibility with them (question 29).
- User decisions (2026-11-12): the Problem statement is four sentences,
  success is compliance with the JovianDSS 1 REST API v4 and JovianDSS 2 is
  not supported within this iteration; security is the second priority;
  the Decision opens with compliance; options are strategies, so the
  reply-type alternatives, implementation nuances, moved from Options to
  question 28.

## Appendix
[Appendix](#appendix)

Every finding this document fixes or rests on, each described once; the rest of
the text uses only the name, as a link. A name in code font without a link is a
review or tracker finding that has no entry here; where it is described, and
where the names and the older IDs come from: [ID sources](#id-sources). Jargon:
[Terms](#terms).

### Evidence
[Evidence](#evidence)

**Table 4 — Evidence sources** · `tbl_evidence_sources`

| Source | What it is | Use |
|---|---|---|
| `jdss-rest-api-spec/empirical/hurl/out/` | every request and reply of the hurl runs on real appliances: 15 jdss1 and 20 jdss2 runs, over 15,900 replies to the calls the plugin decodes. Not in git | **main reference for reply types**, because it has the raw values. Only v4 replies decide pass or fail; v3 replies are background |
| `jdss-rest-api-spec/empirical/{jdss1,jdss2}/` | specs generated from those runs (`INDEX.md`, `-doc`, `-src`), and the TypeSpec of the observed types, compiled to OpenAPI and validated by the spec repo's `--check` against every recorded reply (`typespec/`; `empirical/pilot/tools/tsp/b_<app>_v4/out/openapi/`) | reference for status codes, errors and behaviour; **the source of the generated reply types** ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)). The values in the generated `-doc` replies are masked, so types are not read from those |
| `empirical/hurl/project-status.md` | the hurl suite's numbered findings (cited as "hurl Finding n") | behaviour, data safety |
| `jdss-rest-api-spec/{v4,v3}/d/` | the official (company) API docs | can differ from the appliance; used only where there are no recordings |
| `kub-testing/tmp/cycles/` | csi-sanity runs against the lab appliance | confirms an effect on a live system |

When sources disagree, recorded replies win, then the v4 docs, then the v3 docs.

**How the reply types were checked.** The analysis machine has no Go, so a Python
script imitates what Go's `encoding/json` does with each recorded reply (prototype
`kub-testing/tmp/rest-audit/audit.py`, output `audit.txt`). That is how the
mismatches were found. The types themselves would be generated from the spec
repo's validated description of the replies ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), and [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)'s unit tests would check
them in Go. The [error messages](#error-messages) table and the
[evidence gaps](#evidence-gaps) were produced the same way (`coverage.py`,
`bodies.py`, same folder).

**What this check cannot show.** A field that was null in every recording, or a
list that was always empty, could not be checked (for example vdev spares, or a
pool `scan` in progress). Each hurl appliance is one installation, so a reply
shape that only appears with another pool layout or another feature switched on
would not be in the recordings. The lab appliance and hurl's `jdss1` run the same
JovianDSS 1 release, `1.0up33.65410` (confirmed by the user), and
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) and [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) fail the same way on both. Those
two are the only reply-type findings confirmed on the live system so far (the
"Seen live?" column of [response types](#response-types)).

### Evidence gaps
[Evidence_gaps](#evidence-gaps)

hurl recorded replies for every request the plugin sends, on jdss1 v4 and on jdss2
v4 (see the [compliance matrix](#compliance-matrix)). Two things are missing: hurl
did not always send a request the exact way the plugin does (an extra body, no
`per_page`, longer names), and some situations never came up in the tests at all
(a full pool, a missing pool, an appliance that does not answer). So there are
two kinds of gap: a request the plugin sends differently from the hurl tests, and
a situation that was never recorded. Each gap
below says which kind it is, what the evidence showed before the gap runs, and
what those two extra hurl runs found; [how the gap runs went](#how-the-gap-runs-went)
says what the runs themselves managed to cover, and lists the statuses and error
patterns the code expects that no recording has ever shown.

#### gap_jdss_rest_empty_body_on_get_delete

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | a `{}` body on **every GET**, and on DELETE of target, LUN and share (`rproxy.go:108`) |
| hurl sent | GET: never a body. DELETE of those three: never a body |
| Why it matters | works on jdss1 v3 (the lab today); unknown on v4 and on jdss2 |
| Result of the gap runs | No difference with and without the `{}` body, in 80 (jdss1 v3) and 82 (jdss2 v4) pairs of requests, including DELETE of target, LUN and share. Finding: closed, except on jdss1 v4 |

#### gap_jdss_rest_list_paging_without_per_page

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | lists as `?page=N&_dc=<random>`, without `per_page` |
| hurl sent | never `_dc`. `page` without `per_page` only on snapshot lists, with at most 10 items |
| Why it matters | the default page size is unknown. The plugin keeps asking for the next page until it gets an empty one (`driver/common.go:52`). That stop rule was seen working, but never on a long list without `per_page`. |
| Result of the gap runs | On jdss2 v4, the `volumes` and `nas-volumes` lists ignore `page` when `per_page` is missing; on jdss1 v3, without `per_page`, page 0 held all three test volumes and page 1 and page 99 were empty. With `per_page`, paging works on both. `_dc` changes nothing. The default page size is still unknown. Finding: [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page): **must be fixed before the v4 move** |

#### gap_jdss_rest_volume_create_without_sparse

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | `POST volumes` with only `name` and `size` |
| hurl sent | always with `sparse` |
| Why it matters | every CSI volume gets the appliance's default provisioning, and defaults can change between releases: block size did (hurl Finding 9) |
| Result of the gap runs | The default is **thick** (the full size is reserved) on both: the same `refreservation` as with `sparse: false`, about the volume size. Finding: [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) |

#### gap_jdss_rest_name_length_limit

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | names up to 243 characters (`vp_`/`vb_` + 240) |
| hurl sent | longest name about 59 characters |
| Why it matters | the length limit and its error are unknown |
| Result of the gap runs | Zvol and clone names: 129 characters accepted, 150 refused; the appliance's own message gives the limit, 132. Snapshots: the limit is on the full name `pool/volume@snapshot`, at most 255; with the tested pool and volume names it was reached at 217 ([snapshot name limits](#future-work-snapshot-name-limits)). Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) would cap names at 131 and keep the plugin's generated names within it |


#### gap_jdss_rest_out_of_space_messages

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **Out of space:** a volume, thick volume, snapshot or NAS reservation larger than the pool. The error message was never seen |
| Plugin code that depends on it | `errors.go:270`; `ResourceExhausted` at `controllerplugin.go:378`, `:656` |
| Result of the gap runs | Three different out-of-space messages; only one is recognised. The zvol snapshot, clone and thick-volume cases were not run. Finding: [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised) |

#### gap_jdss_rest_delete_missing_zvol_clone

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Zvol `DeleteClone` of a clone that does not exist (the NAS case is recorded) |
| Plugin code that depends on it | `iscsi_driver.go:520` |
| Result of the gap runs | `500`, errno 1, `Zfs resource: ... not found`: recognised. Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised) (the problem is in the caller); low |

#### gap_jdss_rest_clones_of_missing_parent

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | `GET .../clones` on a missing snapshot or volume, zvol and NAS: only 200 recorded |
| Plugin code that depends on it | `iscsi_driver.go:498`, `nfs_driver.go:635` |
| Result of the gap runs | Clones of a missing parent: `500`, errno 1, recognised. A snapshot with no clones: `200` with an empty list, never `204`. Snapshot list of a missing parent: `500` on jdss1 v3, `404` on jdss2 v4. Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); NFS only |

#### gap_jdss_rest_missing_pool_replies

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | The pool is missing or unavailable, on any call except `GET pools/<pool>` |
| Plugin code that depends on it | every call |
| Result of the gap runs | `POST` and list of volumes on a missing pool: `500`, errno 1. Targets: `404`. Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); low |

#### gap_jdss_rest_detach_and_delete_with_connected_clients

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **Clients connected:** in the hurl tests no initiator logs in and nothing is mounted (hurl README). Detach, target delete and deletes with `force_umount: true` were recorded only on idle resources ([risk 8](#risks)) |
| Plugin code that depends on it | unpublish and every delete |

#### gap_plugin_rest_appliance_unreachable

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **The appliance cannot be reached:** connection refused, DNS, timeout, TLS, reply cut off. Not recorded on purpose, but it happens: jdss1 stopped answering for 15-18 minutes three times (hurl C24), and once for over 8 hours. On this path the plugin crashes ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)) |
| Plugin code that depends on it | `Send` |

#### gap_plugin_rest_unexpected_reply_bodies

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Unexpected reply bodies: HTML or empty 5xx replies, or a list answered with `200` and `data: null`, on which the plugin crashes ([crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body)) |
| Plugin code that depends on it | `getError`, list functions |
| Result of the gap runs | No `200` with `data: null` on any GET. Plain-text `404 not found` and `401 unauthorized` replies exist; they cause a parse error, not a crash. The crash on `null` is still in the code, but no recorded reply reaches it. Finding: [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body), [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status); unit test only |

#### gap_plugin_rest_failover_during_call

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Failover: the pool moves to the other node during a call; several addresses in `addrs` ([bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover)) |
| Plugin code that depends on it | `Send` |

#### What the evidence covers but the plugin ignores
[What_the_evidence_covers_but_the_plugin_ignores](#what-the-evidence-covers-but-the-plugin-ignores)

On jdss2, a snapshot list is briefly out of date after a snapshot is created or
deleted: for under 0.5 s (hurl Finding 20). So a ListSnapshots straight after a
CreateSnapshot may not show the new snapshot. Reading a snapshot by its name works
at once. This document proposes no change for it (see stage 3); what later driver
work would have to allow for is in [risk 14](#risks).

#### How the gap runs went
[How_the_gap_runs_went](#how-the-gap-runs-went)

Two runs were made to fill the gaps: jdss1 `20261105-221223` on v3, and jdss2
`20261105-220744` on v4. **There are no jdss1 v4 results:** jdss1 stopped answering
2 minutes after its v4 part started (the 8-hour outage in [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable)). One test, which
puts a newline into a share comment, broke share handling and NAS deletes on both
appliances for the rest of the run (hurl Finding 21 for jdss2, 22 for jdss1),
and on jdss1 the pool also filled up. So the results below come only from the
part of each run before that test, or from zvol tests, which the breakage did
not touch. The `scale/` tests and `nospace/volume-no-space.hurl` ran on neither
appliance.

Never seen in either run (about 21,000 requests each): `202`, `204` on a GET, `200`
on a DELETE. Also never seen, not even in the `race/` tests:
`Resource ... already exists.` with errno 1 or 100, and "not found" without an
errno. Creating the same object twice at the same time gives errno 100
`dataset already exists` (recognised) or errno 5.

#### How to close the gaps
[How_to_close_the_gaps](#how-to-close-the-gaps)

The spec repo is its own project, so this document does not change it. The user adds
tests for the first eight gaps above ([gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete), [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page), [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse), [gap_jdss_rest_name_length_limit](#gap_jdss_rest_name_length_limit), [gap_jdss_rest_out_of_space_messages](#gap_jdss_rest_out_of_space_messages), [gap_jdss_rest_delete_missing_zvol_clone](#gap_jdss_rest_delete_missing_zvol_clone), [gap_jdss_rest_clones_of_missing_parent](#gap_jdss_rest_clones_of_missing_parent) and [gap_jdss_rest_missing_pool_replies](#gap_jdss_rest_missing_pool_replies)) to the hurl suite; the results so far are
in their entries. [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) and [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call) need a connected client or a second node. csi-sanity's node preset
would cover part of [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) on the lab appliance, but only if the iSCSI NodeStage really
logs in, which review findings `bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout` put in doubt; [Testing](#testing) steps 6 and
7 check this. [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) and [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies) cannot be recorded
reliably, so they would be covered by unit tests with a fake HTTP server ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).

### The `jdss-rest-api-spec` repository
[The_jdss_rest_api_spec_repository](#the-jdss-rest-api-spec-repository)

`jdss-rest-api-spec` (a checkout beside the plugin repository, `../jdss-rest-api-spec`;
not on kub-gate) is what this document measures the plugin against. It describes the
JovianDSS REST API twice, with the same paths in both: once as Open-E documents it
(the company spec), and once as real appliances behave (the empirical spec). It was
written for the JovianDSS-Proxmox and CSI plugin teams.

**Table 5 — What the spec repo holds** · `tbl_spec_forms`

| Part | Where | What it says | Role here |
|---|---|---|---|
| **Company spec** | `v4/d/<path>/<METHOD>-doc` (description), `-src` (JSON Schema of the request body), `-desc`; also `v4/typespec/`, `v4/openapi3/`. `v3/` is laid out the same way | what Open-E documents, taken from the Swagger HTML shipped with the appliance (`v4/origin/`) | which request keys exist, documented status codes; fills gaps |
| **Empirical spec** | `empirical/<appliance>/v4/`: `INDEX.md` (every tested endpoint, with documented and observed status codes), `STATUS-CODES.md`, `SRC-COMPARISON.md`, `ERROR-FIELDS.md`, `d/<path>/<METHOD>-doc`/`-src` (as observed), `typespec/` (namespace `Observed`: the observed types, compiled to OpenAPI under `empirical/pilot/tools/tsp/`, the source of the generated reply types, [impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)); `empirical/<appliance>/status-codes.json` | what real appliances did, generated from the hurl runs | **decides where the two specs disagree** |
| Raw responses | `empirical/hurl/out/` (local only, not in git) | every request and reply of every run | value types (the generated `GET pools/<pool>` values are masked) |
| Findings | `empirical/hurl/project-status.md`, `bug-reports/` | numbered findings, runs, repeatability | behaviour, data safety |

The appliances, as the repo's README describes them:

**Table 6 — Appliances in the spec repo** · `tbl_spec_appliances`

| Name | Release | Auth | APIs |
|---|---|---|---|
| `jdss1` | JovianDSS 1 **`1.0up33.65410`** (the plugin's minimum release), two-node HA cluster | Basic | v3, v4 |
| `jdss2` | JovianDSS 2 `2.0up01.67856` | JWT Bearer token, from `POST api/v4/auth/login`, valid 900 s | v4 only |

**Rules for this work.** The spec repo belongs to its own project. This document and
the plugin work only read it: they never change its code or data, and never run its
scripts or tests (the tests create and delete objects on real appliances). When new
recordings arrive, the user runs its `build_typespec.py --check`, whose output the
generated reply types ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)) are built from. Problems
found in it are reported to the user. The repo covers only the requests the Proxmox
and CSI plugins make: an endpoint missing from it was simply not tested, which does
not mean it is broken.

### What "compliant" means
[What_compliant_means](#what-compliant-means)

The plugin is compliant when every REST call it makes meets the rules below on
jdss1 v4. The jdss2 v4 recordings are evidence beside it: a difference there is
noted, never a failure, because this document does not touch JovianDSS 2. Where the company spec and the empirical spec
disagree, the empirical spec wins ([priority 3](#priorities)).

#### rule_rest_call_exists_on_v4

| | |
|---|---|
| Requirement | Path and method exist: the appliance has answered this call with a success (2xx) status |
| Checked against | `empirical/<app>/v4/INDEX.md` |

#### rule_rest_request_body_valid

| | |
|---|---|
| Requirement | The request body is valid: every key the plugin can send is in the company schema (`<METHOD>-src`), every required key is sent, and the appliance accepted the body in the recorded runs |
| Checked against | `v4/d/<path>/<METHOD>-src`, `empirical/<app>/v4/SRC-COMPARISON.md` |

#### rule_rest_success_statuses_accepted

| | |
|---|---|
| Requirement | Success statuses: the plugin accepts every 2xx status the appliance returned for the call, and a status never returned cannot cause wrong behaviour |
| Checked against | `INDEX.md` observed codes, `status-codes.json` |

#### rule_rest_error_statuses_handled

| | |
|---|---|
| Requirement | Error statuses: every non-2xx status returned for the call is handled on purpose, either mapped to a named `RestError` code or left as Unknown by an explicit decision |
| Checked against | `INDEX.md`, `STATUS-CODES.md`, recorded error bodies |

#### rule_rest_reply_types_decode

| | |
|---|---|
| Requirement | Reply types: every Go type the plugin decodes accepts every recorded v4 reply of that call |
| Checked against | the spec repo's TypeSpec of the observed types, validated by its `--check` against the raw replies in `hurl/out/`; the Go types are generated from it ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), and [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) decodes every recorded reply with them |

#### rule_rest_error_bodies_handled

| | |
|---|---|
| Requirement | Error bodies: errors with missing or varying fields (no `errno`, no `class`) are handled |
| Checked against | `ERROR-FIELDS.md` |

#### rule_rest_docs_vs_appliance_listed

| | |
|---|---|
| Requirement | Docs and reality: where the company spec and the appliance differ, the plugin follows the appliance, and the difference is listed here so it can be reported to the spec project |
| Checked against | company `-doc` against empirical `-doc` |


The rules would be checked again after each new hurl run that the spec project
adds to its generated specs: [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_request_body_valid](#rule_rest_request_body_valid), [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted), [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) and [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) by the audit ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)), [rule_rest_reply_types_decode](#rule_rest_reply_types_decode) by
regenerating the types ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)) and running the unit tests ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

One row per REST call. "v4 doc" is the statuses the company spec lists; "jdss1 v4"
and "jdss2 v4" are the statuses the appliances returned (from both appliances'
`INDEX.md`; jdss2 v4 as evidence only; prototype `kub-testing/tmp/rest-audit/compliance_matrix.py`); "Plugin
success" is the statuses the code treats as success. All 34 calls the plugin makes
today appear in both `INDEX.md` files, and so does the one call stage 3 would add
([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub)), which has its own row. "Note" links the findings that affect
the call; findings that affect every call, such as [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status), are not
repeated.

**Table 7 — Compliance matrix per call** · `tbl_compliance_matrix`

| Plugin call | v4 method and path | v4 doc | jdss1 v4 | jdss2 v4 | Plugin success | Note |
|---|---|---|---|---|---|---|
| `GetPool` | `GET pools/<x>` | 200 | 200, 401, 404 | 200, 401, 404 | 200, 204 | [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `GetVolume` | `GET pools/<x>/volumes/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 204 | [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateVolume` | `POST pools/<x>/volumes` | 201 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) |
| `DeleteVolume` | `DELETE pools/<x>/volumes/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 200, 204 |  |
| `GetVolumesEntries` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) |
| `ListVolumes` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200 | dead, removed |
| `GetVolumeSnapshot` | `GET pools/<x>/volumes/<x>/snapshots/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 202 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) |
| `CreateSnapshot` | `POST pools/<x>/volumes/<x>/snapshots` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) |
| `DeleteSnapshot` | `DELETE pools/<x>/volumes/<x>/snapshots/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 204 |  |
| `GetSnapshotsEntries` | `GET pools/<x>/volumes/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors) |
| `GetVolumeSnapshotsEntries` (new call in stage 3) | `GET pools/<x>/volumes/<x>/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | none today (the function is a stub); 200 once [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) implements it | [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub); fixtures from `list/volume-kinds.hurl` ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) |
| `GetVolumeSnapshotClones` | `GET pools/<x>/volumes/<x>/snapshots/<x>/clones` | 200 | 200, 500 | 200, 500 | 200, 204 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateClone` | `POST pools/<x>/volumes/<x>/clone` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 |  |
| `DeleteClone` | `DELETE pools/<x>/volumes/<x>/snapshots/<x>/clones/<x>` | 204 | 204, 500 | 204, 500 | 204 |  |
| `GetTarget` | `GET pools/<x>/san/iscsi/targets/<x>` | 200 | 200, 404 | 200, 404 | 200 | [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name) |
| `CreateTarget` | `POST pools/<x>/san/iscsi/targets` | 201 | 201, 400, 404, 409, 500 | 201, 400, 404, 409, 500 | 200, 201, 202 | [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay) |
| `DeleteTarget` | `DELETE pools/<x>/san/iscsi/targets/<x>` | 204 | 204, 404 | 204, 404 | 204; 404 = gone |  |
| `AttachVolumeToTarget` | `POST pools/<x>/san/iscsi/targets/<x>/luns` | 200 | 201, 400, 404, 409 | 201, 400, 404, 409 | 200, 201 | [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed): doc 200, appliance 201 |
| `DettachVolumeFromTarget` | `DELETE pools/<x>/san/iscsi/targets/<x>/luns/<x>` | 204 | 204, 404 | 204, 404 | 204; 404 = target gone | [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone) |
| `GetNASVolume` | `GET pools/<x>/nas-volumes/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 204 | [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) |
| `CreateNASVolume` | `POST pools/<x>/nas-volumes` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised) |
| `DeleteNASVolume` | `DELETE pools/<x>/nas-volumes/<x>` | 204, 400, 500 | 204, 400, 404, 500 | 204, 400, 404, 500 | 204 | [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised) |
| `GetNASVolumesEntries` | `GET pools/<x>/nas-volumes` | 200, 500 | 200, 500 | 200, 500 | 200, 201 | [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) |
| `GetNASVolumeSnapshotsEntries` | `GET pools/<x>/nas-volumes/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots): wrong path, reads the volume |
| `GetNASSnapshotsEntries` | `GET pools/<x>/nas-volumes/snapshots` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) |
| `CreateNASSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised) |
| `DeleteNASSnapshot` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>` | 204, 400, 500 | 204, 404, 500 | 204, 404, 500 | 204 |  |
| `DeleteNASVolumeSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>` | **method missing** | 405 | 405 | 204 | **[bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete): wrong method**, [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check); removed |
| `CreateNASVolumeSnapshot` | `POST pools/<x>/nas-volumes` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | dead (`bug_plugin_rest_nas_snapshot_create_copy_of_volume_create`), removed |
| `GetNASVolumeSnapshot` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 202 |  |
| `GetNASVolumeSnapshotClones` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 500 | 200, 500 | 200, 500 | 200, 204 | [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateNASClone` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 400, 500 | 200, 400, 500 | 200, 400, 500 | 200, 201 |  |
| `DeleteNASClone` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>/clones/<x>` | 204, 400, 404, 500 | 204, 404, 500 | 204, 404, 500 | 204 | [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) |
| `CreateShare` | `POST shares` | 201 | 201, 400, 409, 500 | 201, 400, 409, 500 | 200, 201 | [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) |
| `DeleteShare` | `DELETE shares/<x>` | 204 | 204, 404 | 204, 404, 500 | 204 | dead (`bug_plugin_rest_delete_share_nil_descriptor`), removed; [bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class) |

Where the plugin stands on each rule today:

**Table 8 — Compliance status per rule** · `tbl_compliance_status`

| Rule | Status in the current code | Fixed by |
|---|---|---|
| [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4) | **33 of 34 pass.** Fails: [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete). [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) passes on paper, but calls the wrong path | dead-code commit ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)), stage 3 ([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) |
| [rule_rest_request_body_valid](#rule_rest_request_body_valid) | **Passes for every key the driver really sets.** Three optional keys in the Go request structs are not in the company schema, but they are never set ([dead code](#dead-code)). `properties` is an open object in the schema, so any sub-key is allowed. The `{}` body sent on GET and DELETE is outside every schema; [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) found it has no effect. Prototype: `kub-testing/tmp/rest-audit/compliance_bodies.py` | dead-code commit |
| [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted) | **Every 2xx status returned is accepted.** The code also accepts statuses that were never returned: [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check), [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay), 202 on both snapshot GETs, 200 on `DeleteVolume`, 201 on list GETs, and 200 on creates that answer 201 | stage 3 ([bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check), [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay)); the rest are harmless and would stay |
| [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) | **Fails.** [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status); `401` on `GetPool` is not mapped; [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised) | stage 3 (with [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) |
| [rule_rest_reply_types_decode](#rule_rest_reply_types_decode) | **Fails.** [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch). ([bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) decodes without an error, so it does not fail [rule_rest_reply_types_decode](#rule_rest_reply_types_decode)) | stage 2 ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types), generated types) |
| [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) | **Passes on jdss1 v4.** On jdss2, outside this document, [bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class) fails it; only `DeleteShare` runs into it, and the dead-code commit would remove `DeleteShare` | dead-code commit |
| [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed) | Differences found so far: `AttachVolumeToTarget` is documented as 200, the appliance returns 201; the [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) fields are documented as strings; [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch)'s v4 example shows `"off"`; for most zvol calls only the success status is documented, while the appliance also returns 400, 404 and 500; the `POST` used by [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) is not documented and returns 405; target names are documented as up to 255 characters, but 225 is refused (220 accepted); target names are documented as lower case, but upper case is accepted; the name rules for zvols, NAS datasets and snapshots are not documented at all (the company schemas say only `"type": "string"`). The plugin already follows the appliance, except where a finding says otherwise | a list for the spec project ([Relationship](#relationship-to-other-designs)) |

### Response types
[Response_types](#response-types)

Fields whose JSON type or key in the appliance's reply differs from the Go
struct. Any one of them but the last makes the whole reply fail to decode. This catalogue and the two that
follow, [response content and defaults](#response-content-and-defaults) and
[error messages](#error-messages), are the places where the code expects something
other than what the appliance does; they are why stages 2 and 3 exist. Unless an
entry says otherwise, a mismatch looks the same on jdss1 v3, jdss1 v4 and jdss2 v4.

#### bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch

| | |
|---|---|
| Type and field | `ResourceVolume.Encryption` ([`types.go:84`](../../pkg/rest/types.go#L84)) |
| Struct says | `string` |
| Appliance sends | `{"enabled": false, "method": null, "keystatus": null}` |
| v3 docs | not listed |
| v4 docs | object; the example shows `"off"` |
| Plugin reads it? | no |
| Effect | GetVolume and volume lists → `Internal` |
| Seen live? | yes |

#### bug_rest_v4_pools_pool_get_iostats_type_mismatch

| | |
|---|---|
| Type and field | `ResourcePoolIOStats.Read/Write/Chksum` ([`types.go:395`](../../pkg/rest/types.go#L395)), for pool, vdev and disk |
| Struct says | `string` |
| Appliance sends | numbers |
| v3 docs | strings |
| v4 docs | strings |
| Plugin reads it? | no |
| Effect | GetCapacity → `Internal` |
| Seen live? | yes |

#### bug_rest_v4_pools_pool_get_disk_led_type_mismatch

| | |
|---|---|
| Type and field | `Disk.Led` ([`types.go:376`](../../pkg/rest/types.go#L376)) |
| Struct says | `string` |
| Appliance sends | `{"locate": false, "fault": false}` |
| v3 docs | `"off"` |
| v4 docs | `"off"` |
| Plugin reads it? | no |
| Effect | the same; hidden behind [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) |
| Seen live? | present in the replies; never reported, because [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) fails first |

#### bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch

| | |
|---|---|
| Type and field | `ResourceSnapshotShortProperties.Creation` ([`types.go:273`](../../pkg/rest/types.go#L273)) |
| Struct says | `string`, parsed to a number |
| Appliance sends | number |
| v3 docs | number |
| v4 docs | number |
| Plugin reads it? | **yes** |
| Effect | iSCSI ListSnapshots → `Internal` as soon as the pool has a snapshot |
| Seen live? | not yet: the ctrl-B ListSnapshots specs fail earlier, because the controller has already crashed (`connection refused`) |

#### bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch

| | |
|---|---|
| Type and field | `ResourceVolumeSnapshotClones.IsClone` ([`types.go:300`](../../pkg/rest/types.go#L300)) |
| Struct says | `string` |
| Appliance sends | `true` |
| v3 docs | `true` |
| v4 docs | `true` |
| Plugin reads it? | no |
| Effect | deleting a snapshot that has clones fails at `iscsi_driver.go:498` |
| Seen live? | not yet |


No **type** mismatch was found in pool `available`/`size`, snapshots read by name,
NAS volumes, NAS snapshots, targets, or error bodies. The other `Encryption` fields
(`types.go:200`, `:560`) are fine in the recordings. `encoding/json` reports only the
first type error in a reply, so a live test shows only the first failure and hides
the rest; that is why all of them would be closed together, by the generated
types ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)), and checked with [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures), not one at a time. For [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) **the docs are right and the code is
wrong**; for [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) the docs are wrong, and the code follows them.

#### bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec

| | |
|---|---|
| Type and field | `ResourceNASVolume`, most of its tags (`types.go:447-476`) |
| Struct says | camelCase keys: `fullName`, `logBias`, `primaryCache`, `secondaryCache`, `accessTime`, `deduplication` |
| Appliance sends | snake_case keys in every recorded NAS volume GET: `full_name`, `logbias`, `primarycache`, `secondarycache`, `atime`, `dedup`. When set, `quota`, `reservation`, `refquota`, `refreservation`, `recordsize` and `copies` are strings, as the custom decoder expects |
| Plugin reads it? | only `name`, `quota`, `origin`, `copies` and `recordsize`, whose keys match |
| Effect | none today: a type check only sees keys that are present, so the other fields stay empty without an error, and nothing reads them. This is why the NAS volume struct reads so little. With no quota set, `quota` is `null`, which leaves the size at 0 (review `bug_plugin_driver_nas_quota_null_size_zero`, a driver issue) |
| Seen live? | nothing to see: no failure |

### Response content and defaults
[Response_content_and_defaults](#response-content-and-defaults)

Replies that decode, but whose content the code reads wrongly, and a request default
the plugin never sets.

#### bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead

| | |
|---|---|
| Where | `ResourceSnapshotShort.Volume` ([`types.go:294`](../../pkg/rest/types.go#L294)) |
| What happens | there is no such key: the appliance sends `parent.name`. `Volume` stays empty, and the lister skips the entry (`controller/listers.go:54`). **Once [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) is fixed, iSCSI ListSnapshots would return an empty list, without an error** |
| Evidence | recorded replies, code reading |

#### bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key

| | |
|---|---|
| Where | `ResourceNASSnapshotParent.UnmarshalJSON` ([`types.go:658`](../../pkg/rest/types.go#L658)) |
| What happens | reads `type` under the tag `copies`, so `Parent.Type` is always 0. Unused today, but both all-snapshot lists return zvol **and** NAS snapshots, and `parent.type` (4 zvol, 1 NAS) is the only way to tell them apart (hurl Finding 14); stage 3 would read it and drop the other kind's entries ([question 7](#decisions)) |
| Evidence | code reading |

#### bug_plugin_rest_list_calls_send_page_without_per_page

| | |
|---|---|
| Where | `pagedcSuffix` ([`entrielisters.go:34`](../../pkg/rest/entrielisters.go#L34)), `getResourcesList` (`driver/common.go:52`) |
| What happens | lists are requested with `page` and `_dc`, but without `per_page`. On jdss2 v4, the `volumes` and `nas-volumes` lists then ignore `page` and return the first page again (tried with `page=1` and `page=99`). The loop stops only on an empty page, so CSI ListVolumes never ends when `max_entries` is not set, and returns duplicates when it is |
| Evidence | recorded replies (results [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page)); the loop from code reading |

#### bug_plugin_rest_create_volume_never_sets_sparse

| | |
|---|---|
| Where | `CreateVolumeDescriptor.Sparse` ([`reqtypes.go:115`](../../pkg/rest/reqtypes.go#L115)) |
| What happens | never set, so every CSI volume gets the appliance default, which is **thick** on both appliances: the full size is reserved |
| Evidence | recorded replies (results [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse)); decision: sparse by default, configurable in the plugin configuration and per StorageClass, in its own design ([question 9](#decisions)) |

#### bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key

| | |
|---|---|
| Where | `GetVolumeSnapshot` ([`storageapi.go:191`](../../pkg/rest/storageapi.go#L191)) |
| What happens | a zvol snapshot GET has **no `name` key** (NAS snapshot GETs do), so `ResourceSnapshot.Name` and the iSCSI `GeneralSnapshot.Name` (`iscsi_driver.go:455`) are always empty. No caller reads it; callers take the name from the snapshot descriptor. Low |
| Evidence | recorded replies |

### Error messages
[Error_messages](#error-messages)

Error replies the plugin does not recognise. Each becomes `RestErrorFailureUnknown`
today, or the wrong code. The plugin uses these codes to decide whether a repeated
request counts as success and whether a cleanup should go on, so an unrecognised
message takes the wrong path.

#### bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class

| | |
|---|---|
| Call | DeleteShare, on jdss2 |
| Recorded reply | 404, errno 1, no class, `Share <n> not exists.` |
| Plugin expects | a class and a message |
| Effect | jdss2 only, outside this document: not recognised on jdss2 (recognised on jdss1). Would be moot once the dead-code commit removed `DeleteShare` (it has no other caller) |

#### bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised

| | |
|---|---|
| Call | CreateSnapshot, CreateNASSnapshot |
| Recorded reply | 500, errno 5, `Snapshot <p>/<v>@<s> already exists.` |
| Plugin expects | `Resource .*/.* already exists.` |
| Effect | a repeated create is not seen as "already exists", so the path for repeated requests (`controllerplugin.go:654`, `iscsi_driver.go:55`, `:165`, `nfs_driver.go:66`, `:254`) is not taken |

#### bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone

| | |
|---|---|
| Call | DettachVolumeFromTarget |
| Recorded reply | 404, `Volume name <v> is not attached to target <t>` |
| Plugin expects | any 404 means "target gone" (`targets.go:182`) |
| Effect | ControllerUnpublish on a target that exists but lacks this LUN returns early (`iscsi_driver.go:667`), so the **target is left behind** |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised

| | |
|---|---|
| Call | DeleteNASVolume |
| Recorded reply | 500, `In order to delete a dataset, you must delete all of its clones first.` |
| Plugin expects | the pattern says `zvol` |
| Effect | an NFS delete of a volume with clones is not seen as `Busy` (`nfs_driver.go:383`, `:396`), so the clone handling is skipped |

#### bug_rest_v4_shares_post_409_already_exists_unrecognised

| | |
|---|---|
| Call | CreateShare |
| Recorded reply | 409, errno 1, `Share <n> already exists.` |
| Plugin expects | nothing |
| Effect | on an NFS CreateVolume retry, the existing NAS volume is accepted (`nfs_driver.go:113`), but then the share create fails. No retry can succeed |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised

| | |
|---|---|
| Call | DeleteNASClone |
| Recorded reply | 404, `Clone <n> not exists` |
| Plugin expects | nothing |
| Effect | deleting a clone that is already gone is an error, not success |

#### bug_rest_v4_create_calls_out_of_space_messages_unrecognised

| | |
|---|---|
| Call | create on a full pool |
| Recorded reply | 500, errno 4, `cannot create '...': out of space` (zvol, NAS); 500, errno 1000, `cannot create snapshots : out of space` (NAS). Only `400 New zvol size(...) exceeds available space` is recognised |
| Plugin expects | the out-of-space pattern, checked only for errors **without** an errno (`errors.go:270`) |
| Effect | Unknown instead of `ResourceExhausted` (results [gap_jdss_rest_out_of_space_messages](#gap_jdss_rest_out_of_space_messages)) |

#### bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised

| | |
|---|---|
| Call | several |
| Recorded reply | missing pool: 500, errno 1, `ZpoolError Given zpool '<p>' doesn't exists.`, and for targets 404 `Target partition: <p> not exists.`; name too long: 400, errno 1, `ZvolValidationError Zvol name too long ... 132` ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) would keep generated names within 131); v4 snapshot list of a missing parent: 404 `ItemNotFoundError Resource <p>/<v> not found.` |
| Plugin expects | nothing |
| Effect | Unknown (results [gap_jdss_rest_name_length_limit](#gap_jdss_rest_name_length_limit), [gap_jdss_rest_clones_of_missing_parent](#gap_jdss_rest_clones_of_missing_parent), [gap_jdss_rest_missing_pool_replies](#gap_jdss_rest_missing_pool_replies)). Its driver part is the zvol `DeleteClone` caller (`iscsi_driver.go:520`), which turns every error into `Busy`; fixed by [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

### Logic bugs
[Logic_bugs](#logic-bugs)

Bugs in the REST code itself, whatever the appliance does.

#### bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub

| | |
|---|---|
| Where | `GetVolumeSnapshotsEntries` ([`entrielisters.go:53`](../../pkg/rest/entrielisters.go#L53)) |
| What happens | an empty stub that sends no request. ListSnapshots by source volume on iSCSI always returns nothing |
| Evidence | code reading |

#### bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots

| | |
|---|---|
| Where | `GetNASVolumeSnapshotsEntries` ([`entrielisters.go:198`](../../pkg/rest/entrielisters.go#L198)) |
| What happens | the path lacks `/snapshots`, so it reads the NAS volume itself; the reply has no `entries` key, so the list is empty, without an error |
| Evidence | code reading |

#### bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete

| | |
|---|---|
| Where | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:60`](../../pkg/rest/nas_volumes_snapshots.go#L60)), called at `nfs_driver.go:322` |
| What happens | sends `POST` instead of `DELETE`; the appliance answers 405 |
| Evidence | hurl Finding 7 |

#### bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name

| | |
|---|---|
| Where | `targetDneMsgPattern` ([`errors.go:96`](../../pkg/rest/errors.go#L96)) |
| What happens | a hard-coded test name, `Target iqn.20215:test2asd not exists.`, so it can never match a real target. The real 404 is `ItemNotFoundError`, `Target <name> not exists.`, so the general error mapping never returns `RestErrorResourceDNETarget`. `DettachVolumeFromTarget` sets that code itself on any 404 (`targets.go:185`), which leads to [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone) |
| Evidence | hurl Finding 4 |

#### bug_plugin_rest_get_calls_decode_body_before_204_check

| | |
|---|---|
| Where | every `Get*` that accepts 204 (`GetVolume`, `GetPool`, `GetNASVolume`, `GetVolumeSnapshotClones`, `GetNASVolumeSnapshotClones`; e.g. [`storageapi.go:95-99`](../../pkg/rest/storageapi.go#L95)) |
| What happens | decodes the body **before** checking the status. A real 204 has no body, so decoding fails and returns `RestErrorRPM`: the 204 success branch can never be reached. No GET answered 204 in about 21,000 recorded requests per run |
| Evidence | code reading |

#### bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay

| | |
|---|---|
| Where | `CreateTarget` on 202 ([`targets.go:89-98`](../../pkg/rest/targets.go#L89)) |
| What happens | calls `GetTarget` 10 times in a row, with no delay and no stop on success, and returns only the last result. 202 was never recorded, so this may never run |
| Evidence | code reading |

#### bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check

| | |
|---|---|
| Where | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:79-85`](../../pkg/rest/nas_volumes_snapshots.go#L79)) |
| What happens | decodes the body **before** checking for 204. A 204 has no body, so decoding fails and the call returns `RestErrorRPM`. Hidden behind [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete): fixing only [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)'s method would make every successful delete report an error. Its siblings `DeleteNASVolume` and `DeleteShare` check for 204 first. Not [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) (that is about GETs) |
| Evidence | code reading |

#### bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors

| | |
|---|---|
| Where | `ResourcePool.UnmarshalJSON` ([`types.go:337-340`](../../pkg/rest/types.go#L337)), `ResourceSnapshotShortProperties.UnmarshalJSON` (`types.go:283-286`) |
| What happens | a value that does not parse returns no error and leaves a zero value: pool `available` becomes 0 (GetCapacity would report a full pool), snapshot `creation` becomes the zero time |
| Evidence | code reading |

#### bug_plugin_rest_get_error_maps_body_only_ignores_http_status

| | |
|---|---|
| Where | `getError` ([`storageapi.go:42-58`](../../pkg/rest/storageapi.go#L42)); every `Get*` decodes before looking at the status |
| What happens | only the JSON body is mapped; the HTTP status is never used. Error bodies that are not JSON (plain text `401 unauthorized` on jdss1, `404 not found`) give `RestErrorRPM` ("Failure during processing response from storage") on GETs, and `RequestMalfunction` ("Failure in sending data to storage", `errors.go:287`) on create and delete, which wrongly points at the network. A wrong password gives no hint about credentials, and a plain 404 is not seen as "not found" |
| Evidence | recorded bodies, code reading |

### Driver bugs
[Driver_bugs](#driver-bugs)

Bugs outside `pkg/rest`, in the driver code that calls it. This document fixes them
because they misread REST errors ([priority 8](#priorities)).

#### bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children

| | |
|---|---|
| Where | `DeleteVolume` (`iscsi_driver.go:245`, `:257`; `nfs_driver.go:383`, `:396`) |
| What happens | the intermediate-snapshot cleanup runs only for `Busy`. A leftover snapshot without clones gives `volume has children` (errno 1000), mapped to `BusyVolumeHasSnapshots`, so the cleanup never runs, `DeleteVolume` returns `Internal` on every retry, and **the volume can never be deleted** |
| Fixed by | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error

| | |
|---|---|
| Where | `cleanIntermediateSnapshots` (`iscsi_driver.go:197-208`, `nfs_driver.go:302-312`) |
| What happens | the listing error is checked after the type check, so a real error such as "not found" becomes `FailureUnknown`; and the NFS version expects the zvol snapshot type although its lister returns NAS snapshots, so it always fails with "Unable to typecast snapshot data" and never reaches the [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) call at `nfs_driver.go:322` |
| Fixed by | [Stage 3](#stage-3-behaviour-and-error-mapping), its own fix after the error catalogue |

### Security findings
[Security_findings](#security-findings)

All are fixed in this delivery, in [stage 4](#stage-4-robustness-and-security): [security_plugin_rest_tls_verification_off](#security_plugin_rest_tls_verification_off) by
[impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), `security_plugin_secrets_in_logs` and `security_plugin_rest_request_dump_every_call` by [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs), [security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments) by [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments). Stage 4 also adds [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit); [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) is
the [first commit](#first-commit-make-go-vet-pass).

#### security_plugin_rest_tls_verification_off

| | |
|---|---|
| Where | `rproxy.go:251` |
| Finding | TLS certificate checks are always off (`InsecureSkipVerify: true`, no setting), while the admin's Basic credentials go with every request |
| Goes to | this delivery: verification on by default, `insecure` and `cafile` settings ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), stage 4) |

#### security_plugin_rest_unescaped_path_segments

| | |
|---|---|
| Where | every path builder in `pkg/rest` |
| Finding | paths are built with `fmt.Sprintf` and no `url.PathEscape`; a `?` or `#` in a name cuts the path short, and a crafted volume handle can reach other resources |
| Goes to | this delivery: `url.PathEscape` on every path segment ([impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments), stage 4) |


The controller and node also log whole CSI requests, `Secrets` included, at Debug
and at Warn (review `security_plugin_secrets_in_logs`, outside `pkg/rest`); [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) removes that.

### Crashes
[Crashes](#crashes)

Places where the REST layer can crash the controller (a Go panic). Stage 4 would fix
all five inside `pkg/rest` ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)).

#### crash_plugin_rest_nil_response_on_transport_error

| | |
|---|---|
| Where | `rproxy.go:138-157` |
| Crash | every branch that handles a transport error reads `res.StatusCode` while `res` is `nil` |
| Trigger | a refused connection (`:141`), or the client's own timeout, which is not a `*net.OpError` (`:153`). During jdss1's 8-hour outage a controller would have crashed again and again (inferred). Gap [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) |

#### crash_plugin_rest_basic_auth_before_request_check

| | |
|---|---|
| Where | `rproxy.go:122-123` |
| Crash | `SetBasicAuth` runs before the error from `http.NewRequest` is checked |
| Trigger | a config without `prot` (there is no default) gives the URL `://host`, and `NewRequest` fails |

#### crash_plugin_rest_error_from_nil_body

| | |
|---|---|
| Where | `errors.go:140-149` |
| Crash | `ErrorFromErrorT(nil)` logs the `nil` case, then uses it anyway |
| Trigger | a list with `"data": null`, or an error body without `"error"`. Gap [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies) |

#### crash_plugin_rest_setup_endpoint_ignores_error

| | |
|---|---|
| Where | `rest.go:104-107` |
| Crash | `SetupEndpoint` ignores the error from `SetupRestProxy` |
| Trigger | a missing or invalid `idletimeout`, or an empty `addrs`: the first REST call panics |

#### crash_plugin_rest_endpoint_string_empty_addrs

| | |
|---|---|
| Where | `rest.go:83-95` |
| Crash | `RestEndpoint.String()` throws away the string it builds and returns `Addrs[0]`, so it panics on an empty list |
| Trigger | an endpoint with no addresses formatted with `%v` or `%s`: a log line alone can panic |

### REST client behaviour
[REST_client_behaviour](#rest-client-behaviour)

Not wrong replies, but how the client itself behaves: retries, failover, timeouts,
error text and logging. For a later REST client design, except [bug_plugin_rest_unknown_error_empty_message](#bug_plugin_rest_unknown_error_empty_message) and [bug_plugin_rest_every_error_logged_at_warn](#bug_plugin_rest_every_error_logged_at_warn) (fixed in
stage 3 by [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) and [bug_plugin_rest_reply_text_as_format_string](#bug_plugin_rest_reply_text_as_format_string) ([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) in the first commit, [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes) in stage 4).

#### bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover

| | |
|---|---|
| Where | `rproxy.go:91`, `:274`, `:281` |
| Behaviour | `tries` and every address after the first are never used: no retry and no failover (gap [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call)) |

#### bug_plugin_rest_reply_text_as_format_string

| | |
|---|---|
| Where | `storageapi.go:51`, `:53` and others |
| Behaviour | the reply body is used as a format string; transport errors are reported as `RequestMalfunction`, which hides what kind of error it was. Fixed in this delivery: the format-string part by [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) (first commit), the transport-error part by [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes) ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error), stage 4) |

#### bug_plugin_rest_request_built_without_context_csi_deadline_not_applied

| | |
|---|---|
| Where | `rproxy.go:122`, `:257` |
| Behaviour | the request is built without the context, so the CSI deadline never reaches it; `idletimeout` is used as the timeout for the whole request ([risk 12](#risks), [future work](#future-work-request-deadlines)) |

#### bug_plugin_rest_unknown_error_empty_message

| | |
|---|---|
| Where | `errors.go:275-277`, `:297` |
| Behaviour | errors that match no pattern get an empty message; the appliance's text appears only in a Warn log. Fixed in stage 3 by [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |

#### bug_plugin_rest_every_error_logged_at_warn

| | |
|---|---|
| Where | `errors.go:150`, `:173`, `:265`, `:291` |
| Behaviour | every appliance error is logged at Warn, including the expected "not found" on repeated deletes; the not-found message repeats a word ("Resource Resource ... do not exists"). Fixed in stage 3 by [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |

### Dead code
[Dead_code](#dead-code)

Code with no callers in the current code. The
[dead-code commit](#before-stage-1-remove-dead-rest-code)
would remove it.

- **Functions:** `ListVolumes` (with `GetVolumesData`), `getResultsEntries`,
  `GetTimeStamp`, `CreateNASVolumeSnapshot` (`bug_plugin_rest_nas_snapshot_create_copy_of_volume_create`: a copy of `CreateNASVolume` that
  would create a NAS volume), `DeleteNASVolumeSnapshot` ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete), [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check); its one caller
  would be switched in the same commit), `PromoteClone` and `AddUserToTarget` (stubs
  that return nil), `DeleteShare` (`bug_plugin_rest_delete_share_nil_descriptor`: sends nil instead of its descriptor; its
  only caller is commented out at `nfs_driver.go:344`; the recordings show that
  deleting a NAS volume removes its shares, so it is not needed).
- **Types:** `StorageInterface` (old signatures that `RestEndpoint` does not
  implement), `RestProxyInterface`, the duplicate `restProxy` struct, `RestProxyCfg`,
  `StorageCfg`, `SnapshotDescriptor`, and the unused `*Data` reply types in
  `payload-def.go` (about 16).
- **Request keys** the driver never sets and the company schema does not have ([rule_rest_request_body_valid](#rule_rest_request_body_valid)):
  `TargetLunDescriptor.EUI`, `CreateNASVolumeDescriptor.ResWithDescendents` and
  `.RefquotaWithDescendents`.

### Measurements
[Measurements](#measurements)

**v3 against v4, as recorded** by the hurl suite
(`jdss-rest-api-spec/empirical/hurl/`):

| Comparison | Recorded | Cases | Differences | Source |
|---|---|---|---|---|
| jdss1 v3 against jdss1 v4, lists and delete flags | jdss1 run `20261104-200231` | 363 per version | none: status, errno, class and message identical | hurl Finding 14 |
| jdss1 v3 against jdss1 v4, rollback and clone id | jdss1, 2026-11-04 | 281 common | 7: one NAS `rapid` count (1 against 0) and six NAS `effect` cases lost to an appliance outage on the v3 side | hurl Finding 15 |
| jdss2 v4 against jdss1 v4, lists and delete flags | jdss2, 2026-11-04 | 751 common | 12: eleven the known share 404 style, one a VIP address inside a message | hurl Finding 15 |
| with and without the `{}` body on GET and on DELETE of target, LUN and share | jdss1 v3 and jdss2 v4 | 80 and 82 pairs | none | [how the gap runs went](#how-the-gap-runs-went), [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) |

**Hard-coded API paths in the current code:** 34 `api/v3` strings in
`pkg/rest`, 22 distinct path patterns; 30 strings after the dead-code commit.

**The REST package, measured on commit `11ffbfd`:**

| File | Lines | Role |
|---|---|---|
| `types.go` | 675 | 19 `Resource*` reply types, 10 accessors, 8 custom decoders |
| `payload-def.go` | 508 | mostly unused types, removed by the dead-code commit |
| `storageapi.go` | 401 | volume, snapshot and clone calls |
| `errors.go` | 306 | error mapping |
| `entrielisters.go` | 291 | list calls |
| `rproxy.go` | 284 | transport |
| `targets.go` | 260 | target and LUN calls |
| `reqtypes.go` | 254 | request types |
| `nas_volumes.go`, `nas_volumes_snapshots.go`, `nas_volumes_snapshots_clones.go`, `clones.go` | 462 | NAS and clone calls |
| `rest.go` | 108 | setup |
| `pools.go`, `nas_snapshots.go`, `shares.go`, `general_resources.go` | 257 | pool, NAS snapshot and share calls; the last holds only the licence header |
| total | 3,806 | |

**Reply types with their fields, accessors and decoders:**

| Type | JSON fields | Accessors | Custom decoder converts |
|---|---|---|---|
| `ResourceVolume` | 86 | `GetSize`, `OriginVolume`, `OriginSnapshot` | `volsize` string to `int64` |
| `ResourceSnapshot` | 25 | `ClonesNames`, `GetSize` | `creation`, `volsize` strings to numbers |
| `ResourceSnapshotShortProperties` | 2 | | `creation` string to number (the list sends `int64`: [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch)) |
| `ResourcePool` | 15 | | `available` string to `int64` |
| `ResourceNASVolume` | 28 | `GetSize`, `OriginVolume`, `OriginSnapshot` | `copies`, `recordsize`, `quota` strings to numbers |
| `ResourceNASVolumeSnapshot` | 33 | `ClonesNames`, `GetSize` | `creation` |
| `ResourceNASSnapshotShortProperties` | 3 | | `creation` |
| `ResourceNASSnapshotParent` | 3 | | `copies`; reads `type` under the wrong key ([bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key)) |
| `ResourceTarget`, `ResourceShare`, `ResourceVolumeSnapshotClones`, pool sub-types | 4 to 8 each | | none |

**Observed types of the read fields, jdss1 v4:** volume by name: `name`,
`volsize`, `available`, `used`, `reservation`, `refreservation` strings;
`clones` and `quota` `null` in every recording; in the zvol alternative
`is_clone` `false` and `origin` `null`, in the clone alternative `is_clone`
`true` and `origin` a string.
All-snapshots list: `name` string, `parent.name` string, `creation` `int64`.
Pool: `available`, `size`, `name`, `health` strings; `status` `int64`.

**What the company documentation says for each mismatched field**, against
the recordings: [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) `encryption`: v3 not listed, v4 an object whose example
shows `"off"`, appliance an object; [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) I/O statistics: documented as strings,
appliance numbers; [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) `led`: documented `"off"`, appliance an object; [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch)
`creation`: documented number, appliance number; [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) `is_clone`: documented
`true`, appliance boolean.

**The compiled OpenAPI, jdss2 v4 file:** 663 `type: 'null'` fields, 81
`anyOf`, no `additionalProperties: false` (closedness is enforced by
`--check`, not by the schema); `is_clone` as a boolean with a one-value
enum; the tab defect at line 15389.

### ID sources
[ID_sources](#id-sources)

**Table 9 — Where the names and IDs come from** · `tbl_id_sources`

| Names and IDs | Meaning | Source |
|---|---|---|
| `gap_…`, `bug_…`, `crash_…`, `security_…`, `rule_…`, `impl_…` | the findings, gaps, compliance rules and implementation items of this document, each a heading in the [What changes](#what-changes) section or in this Appendix; the form is `kind_area_slug` | this document; the old letter-and-number IDs are mapped to them in `kub-testing/issues/` |
| C, D, K, N, R, S, X + number | the code reviews' own IDs, kept in the reviews and the tracker | [`review/2026-10-02-code-review.md`](../../review/2026-10-02-code-review.md) and [`review/2026-11-07-code-review.md`](../../review/2026-11-07-code-review.md); mapped to names in `kub-testing/issues/` |
| F + number | the test findings' IDs, kept in the tracker | `kub-testing/ISSUES.md`; mapped to names in `kub-testing/issues/test-findings.md` |
| D1-D7 in 0003 | design 0003's own items, not review IDs | [0003](../design/0003-controller-crash-fixes-design.md) |
| hurl Finding n, C24 | the hurl suite's findings and caveats | the spec repo's own notes |
| INF8, MF7, MK11, T11, TOOL10 | kub-testing items: lab infrastructure, manifests, Makefile, testing, tooling | `kub-testing/ISSUES.md` |

### Terms
[Terms](#terms)

**Table 10 — Terms** · `tbl_terms`

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
| analysis machine | the machine with the spec repo and the hurl recordings, where the audit scripts and the type generator run; it has no Go |
| company spec, empirical spec | the API as Open-E documents it, and the API as the appliances behaved in the hurl runs ([Table 5](#the-jdss-rest-api-spec-repository)) |
| decode | turn a JSON reply into a Go struct (Go's `encoding/json`) |
| zvol | a ZFS block volume; the plugin uses it for iSCSI volumes |
| NAS volume | a ZFS file system (dataset); the plugin uses it for NFS volumes |
| errno, class, message | the three fields of a JovianDSS error reply: `{"error": {"errno": …, "class": …, "message": …}}` |
| `RestError` code | the plugin's own code for a REST failure (`errors.go`), e.g. `RestErrorResourceDNE` ("does not exist"); the driver turns it into a CSI status |
| `Internal`, `Unknown`, `AlreadyExists`, `ResourceExhausted`, … | CSI (gRPC) status codes the plugin returns to Kubernetes |
| repeated request | CSI may send the same create or delete again; the plugin must treat "already exists" on a create, and "not found" on a delete, as success |
| fixture | a recorded reply saved in the repo as test data for unit tests |
| generated types | the reply types in `pkg/rest/types_gen.go`, written by `tools/rest-types-gen.py` from the spec repo's recorded API and never edited by hand ([impl_plugin_rest_generated_reply_types](#impl_plugin_rest_generated_reply_types)) |
| mapping | `pkg/rest/types_gen.yaml`: for each generated type, the endpoint it comes from and the fields the plugin reads, each required or optional |
| stage 1-4 | the four stages of [What changes](#what-changes); the three commits before stage 1 and the supporting work are outside them; not the CSI NodeStage call |
| `vp_`, `vb_`, `vs_`; `sp_`, `sb_`, `ss_` | the prefixes of the names the plugin gives volumes (`v`) and snapshots (`s`) on the appliance: `p` for the CSI name as it is, `b` for its base32 form, `s` for a SHA-256 hash of it ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)) |
| intermediate snapshot | a snapshot the plugin creates on its own when it makes a volume from another volume; the new volume is a clone of that snapshot, and `cleanIntermediateSnapshots` removes such snapshots when the volume is deleted |
| transport error | a request that gets no HTTP reply at all: the address does not resolve, the connection is refused, or the request times out ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)) |
| Basic auth, Bearer token | two ways to log in to the REST API: Basic sends the user name and password with every request; a Bearer token is obtained once by logging in and sent instead of the password until it expires |
| CHAP | the password login of iSCSI between a node and a target; the plugin does not use it today |
| initiator, target, LUN | iSCSI terms: the initiator is the client (here the Kubernetes node), the target is what the appliance exposes, and a LUN is one volume attached to a target |
| static PV | a PersistentVolume an administrator writes by hand, naming an existing volume, instead of one the plugin creates on request |

