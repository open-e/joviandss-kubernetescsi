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
  - [The_observed_API_in_jdss_rest_api_spec](#the-observed-api-in-jdss-rest-api-spec)
  - [The_REST_layer_measured](#the-rest-layer-measured)
  - [How_lists_are_paged](#how-lists-are-paged)
  - [What_we_do_not_know](#what-we-do-not-know)
  - [Key_observations](#key-observations)
- [Priorities](#priorities)
- [Options_considered](#options-considered)
- [Decision](#decision)
- [What_changes](#what-changes)
  - [First_commit_make_go_vet_pass_in_pkg_rest_and_pkg_controller](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller)
  - [Before_stage_1_keep_secrets_out_of_the_logs](#before-stage-1-keep-secrets-out-of-the-logs)
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
  - [Code_segments](#code-segments)
  - [ID_sources](#id-sources)
  - [Terms](#terms)

## List of Tables
[List_of_Tables](#list-of-tables)

1. `tbl_api_versions` — **API versions and auth**
2. `tbl_observed_api` — **The observed API's files and tools**
3. `tbl_vet_reports` — **The `go vet` reports in `pkg/rest`**
4. `tbl_type_generation` — **Reply type rules**
5. `tbl_read_fields` — **Read fields per reply type**
6. `tbl_name_rules` — **Name rules per object**
7. `tbl_systems_affected` — **Affected systems**
8. `tbl_files_changed` — **Files that would change**
9. `tbl_decisions` — **Decided questions**
10. `tbl_evidence_sources` — **Evidence sources**
11. `tbl_spec_forms` — **What `jdss-rest-api-spec` holds**
12. `tbl_spec_appliances` — **Appliances in `jdss-rest-api-spec`**
13. `tbl_compliance_matrix` — **Compliance matrix per call**
14. `tbl_compliance_status` — **Compliance status per rule**
15. `tbl_recorded_comparisons` — **Recorded comparisons**
16. `tbl_list_request_times` — **List request times on jdss1 v4**
17. `tbl_rest_package_files` — **Files of the REST package**
18. `tbl_reply_type_fields` — **Reply types, fields, accessors and decoders**
19. `tbl_error_kind_codes` — **Kinds of error and their codes**
20. `tbl_id_sources` — **Where the names and IDs come from**
21. `tbl_terms` — **Terms**

## Metadata
[Metadata](#metadata)

Date: 2026-11-05 (API v4 only decided), 2026-11-10 (the whole migration, compliance with the recorded v4 API, decided as one record); first written 2026-11-05 as the document named under Formerly, in this form 2026-11-09, merged into one record 2026-11-10
Driver: Andrei Perepiolkin
Approver: andrei.perepiolkin
Code: commit `a3800e4`, called "the current code" below; line numbers refer to it
Implementation: the [What changes](#what-changes) section, one series of commits delivered and tested in one go
Formerly: `docs/design/0002-rest-api-compliance-design.md`, merged into this document on 2026-11-10

## Status
[Status](#status)

Proposed

## Problem statement
[Problem_statement](#problem-statement)

The plugin uses JovianDSS REST API v3, which is deprecated, through a client
that follows documentation and assumptions rather than what the appliance
serves, and it pages its CSI lists in a way that crashes the controller or
returns entries twice or never. It has to use API v4 through a client that
complies with the API as served, down to the way the appliance pages its
lists.

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

**On JovianDSS 1, v3 and v4 answer the plugin's requests the same, except for the lists of one
volume's snapshots.** The hurl suite compared them directly: in 363 delete
cases (hurl Finding 14) the two versions returned the same status, errno,
class and message in every case, the list questions of that finding got the
same answers on both, and in 281 rollback and clone cases (hurl Finding 15)
they did the same, apart from a few requests that got no answer at all and
one count that varied. The shape of list replies differs: the v4 lists that page have one shape, `entries` and `results`, where v3 returns a plain array
without paging and the v4 shape with it (hurl Finding 14), and the plugin always asks for a page; the two clones calls,
`GET .../snapshots/<s>/clones` for a zvol and for a NAS volume, are not
paged and return a plain array on v4 too. The two per-volume snapshot lists,
`GET .../volumes/<v>/snapshots` and `GET .../nas-volumes/<d>/snapshots`,
differ in two more ways: v3 lists full snapshot objects where v4 lists short
entries of `name`, `parent` and `properties` (hurl caveat C13); and a missing
pool or parent is answered 500 with errno 1 on v3 and 404 on v4
([gap_jdss_rest_clones_of_missing_parent](#gap_jdss_rest_clones_of_missing_parent),
[gap_jdss_rest_missing_pool_replies](#gap_jdss_rest_missing_pool_replies)).
The plugin reads neither list today: the iSCSI one is an empty stub and the
NFS one asks the wrong path
([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub),
[bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)),
and stage 3, which repairs them, reads their v4 entries
([impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow)).
So the move from v3 to v4 by itself would not change how the plugin
behaves on JovianDSS 1. The JovianDSS Proxmox
plugin already uses v4 by default
(`JovianDSS-Proxmox/jdssc/jdssc/jovian_common/rest_proxy.py:135`).

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
jdss1 v4 and jdss2 v4.
No other type mismatch was found in pool `available` and `size`, snapshots read
by name, NAS volumes, NAS snapshots, targets or error bodies. The documentation
is right for two of the mismatches and wrong for two: for [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) the documentation
is right and the code wrong; for [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) the documentation says strings,
the appliance sends numbers and an object, and the code follows the
documentation. The decoder reports only the first type error in a reply, so
a live test shows one failure at a time; [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) has never been reported for that
reason. Full detail, with what the documentation says:
[response types](#response-types).

### The observed API in `jdss-rest-api-spec`
[The_observed_API_in_jdss_rest_api_spec](#the-observed-api-in-jdss-rest-api-spec)

**`jdss-rest-api-spec` already holds a validated, machine-readable description
of what each appliance sends.** It is read only for this
work; the user runs its tools.

**Table 2 — The observed API's files and tools** · `tbl_observed_api`

| What | Where in `jdss-rest-api-spec` | State |
|---|---|---|
| Empirical TypeSpec, one file per method, types as observed | `empirical/jdss1/v4/typespec/<path>/<METHOD>.tsp`; jdss2 v4 has its own, under `empirical/jdss2/v4/typespec/` | tracked in git; grows as runs are added |
| Error models, one file per kind of error | `empirical/jdss1/v4/typespec/models/errors/<Kind>.tsp` | tracked in git; grows as runs are added |
| Scenarios, one file per method: which test case gave which error, with its status, `class`, `errno`, a pattern for the message and the kind of error | `empirical/jdss1/v4/typespec/<path>/<METHOD>-scenarios.json` | plain JSON, tracked in git; no case carries a scenario key yet |
| Tracked OpenAPI 3.1, compiled from the TypeSpec | `empirical/jdss1/v4/openapi3/`: `openapi.yaml`, `paths/<path>.yaml` with every method of one path, `components/schemas/<Kind>.yaml` with one kind of error each | YAML, tracked in git, rewritten only after a passing check. **What the reply types and the error kinds are read from** ([question 31](#decisions)) |
| Build folder of the compiler | `empirical/pilot/tools/tsp/b_<appliance>_<api>/` | rewritten by the check on every run and ignored by git: not an input of this work |
| Validation | `empirical/hurl/build_typespec.py --check` | compiles each tree and validates every recorded response against it; fails on an invalid response or on a key the spec does not have |
| Manifest | `b_<appliance>_<api>.manifest.json`, beside the build folder | every recorded response behind each route and status; part of the same ignored build output, and not used here |

How the observed types are written ([empirical
README](../../../jdss-rest-api-spec/empirical/README.md)): a field is
required when it was never seen missing; a field that was `null` in every
recording is typed `null`; the values seen are listed when there are at most four; a zvol
and its clone are separate alternatives split on `is_clone`; other shapes
are merged, so optional fields do not show which go together. For the
plugin's calls, the observed types say exactly what the six mismatches say.
Types differ per endpoint for the same thing: a snapshot's `creation` is
`int64` in the all-snapshots list and a string in the reply by name. `volsize` and the pool's `available` are strings holding numbers, which the
hand-written decoders parse today. jdss1 and jdss2 agree on types
for the endpoints the plugin uses. The build folder holds the compiled OpenAPI
as one JSON file, which nothing here reads; the tracked OpenAPI is split into
YAML files, which `jdss-rest-api-spec`'s check reads back with Python's YAML
parser.

### The REST layer, measured
[The_REST_layer_measured](#the-rest-layer-measured)

**The plugin's REST layer, measured** on the current code (the
[Appendix](#appendix) has the detail): `pkg/rest` is 3,806 lines. The reply
types are `types.go`, 675 lines, 19 `Resource*` structs, with 10 accessor
methods and 8 custom decoders that turn string numbers into `int64` and
strings into times. The 37 call functions (34 send a request, three are stubs), the request types (`reqtypes.go`,
254 lines), the transport and the error mapping are in other files. `Send` takes the status that counts as success as an argument and never reads it; each caller checks the status itself (`rproxy.go:84`). Outside
`pkg/rest`, the drivers and the controller use 7 of the 19 types at about 40
places, almost all `.Name`, plus `GetSize`, `OriginSnapshot`, `ClonesNames`
and `Creation`. Twenty-eight lines read a field whose type or key would change:
`Creation` as a field at six places, the list entry's `Volume` at three (the
key the appliance sends is `parent.name`, [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead)), the pool's `Available` at
one, the two drivers' copies of `Creation`, and `Name` of `ResourceSnapshot`, the type of a zvol snapshot read by name, whose reply does not carry it, at sixteen: once in the iSCSI
`GetSnapshot` and fifteen times on entries of the per-volume snapshot lists
([questions 42 and 43](#decisions)).

What the code reviews and the recordings found in this layer is catalogued
in the Appendix: replies whose [types](#response-types) do not decode,
[error messages](#error-messages) the code does not recognise,
[logic bugs](#logic-bugs) such as a status checked after the body is
decoded, [crashes](#crashes) on a transport error or an empty body, and
[security findings](#security-findings): secrets in the logs, certificate
checks off, unescaped path segments.

### How lists are paged
[How_lists_are_paged](#how-lists-are-paged)

**The appliance serves a list as a slice of the whole list, and nothing
holds the list still between two requests** (hurl Finding 24, the volume
list on jdss1 v4). A request with `page` and `per_page` gets at most
`per_page` entries, from entry `page × per_page` on, of the whole list in
alphabetical order of the names, and `results`, the length of the whole
list. Pages start at 0, any `per_page` is served in full with no cap, a
page past the end is empty, and every complete read returned each volume
once. Two cases are special: a request without `per_page` returns pages of
20, and `page` still works (`page=1` alone gives entries 20 to 39);
`per_page=0` lists every entry in one reply, whatever `page`. A read cannot start at a name: `where` is ignored, and `search` only
filters. When the list changed during a read, the pages shifted: a volume
deleted from a page already read made the next page start one entry later,
so one volume was never seen, and a volume added before the read's
position would make one come twice. Every request makes the appliance
build the whole list, whatever `per_page` ([Measurements](#measurements)).
No paging run has been made of the other five lists the plugin reads, or of v3; they
are taken to page the same way ([question 49](#decisions), [risk 28](#risks)).

**The CSI list calls are built on that paging, and break on it.**
ListVolumes and ListSnapshots take `max_entries` and a `starting_token` and
return a `next_token`. The plugin writes into its token the appliance page
it stopped on, a random number it sends as `_dc`, and the base64 form of
the last name it returned; the next call reads that page again and looks
for that name (`driver/common.go:31-86`, `driver/snaptoken.go`). A token
the plugin issued crashes the controller when it comes back
([crash_plugin_driver_list_token_parse_writes_through_nil_pointer](#crash_plugin_driver_list_token_parse_writes_through_nil_pointer)).
Behind that crash, a resumed call repeats an entry, leaves pages out, and
stops early when the name it looks for has gone, and all snapshots of one
volume share one token ([driver bugs](#driver-bugs)). csi-sanity reaches
none of this: its ListVolumes paging spec is pending in csi-test, its
invalid-token spec sends a token that fails the parse before the crash,
and its ListSnapshots specs wait for [design 0003](../design/0003-controller-crash-fixes-design.md) ([Testing](#testing) limit 9).

### What we do not know
[What_we_do_not_know](#what-we-do-not-know)

What the recordings do not show: how JovianDSS 1 releases older than
`1.0up33.65410` behave on v4, since no such appliance was available; the user
states that they do not serve the same v4, so none is supported
([question 29](#decisions), [risk 2](#risks));
what simultaneous requests do on jdss1 v4, because that test ran there on
v3 only; what a snapshot or a clone create returns on a full pool, because
the full-pool run on v4 covered the volume creates only; how lists of hundreds of objects page, apart from the volume list,
because the paging runs of the other lists have not been made
([How lists are paged](#how-lists-are-paged));
and the type of any field that was `null` in every recording.

The Appendix holds all twelve [evidence gaps](#evidence-gaps), with what was found by the three runs made to close them, by a fourth for out of space and by the paging run of the volume list (hurl Finding 24).

### Key observations
[Key_observations](#key-observations)

1. **v3 ends with JovianDSS 1.** It is deprecated, the newer generation does
   not serve it, and the Proxmox plugin has already moved to v4. Staying on v3 is
   a decision to stop at JovianDSS 1.
2. **The move alone would change no behaviour on JovianDSS 1.** v3 and v4 answer
   the same in every recorded case the plugin reaches today. List replies
   differ in shape, and the plugin always asks for the paged shape; the two
   per-volume snapshot lists differ in their entries and in the reply to a
   missing pool or parent, and the plugin reads neither before stage 3
   repairs them ([API versions](#api-versions)).
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
   and a few accessors are the whole surface; keep those and the type
   change touches the drivers and the controller at thirteen lines ([Stage 2](#stage-2-response-types)), besides the readers of the
   per-volume snapshot lists, which stage 3 rewrites.
7. **The expected-status argument of `Send` is ignored** (`rproxy.go:84`). `Send`
   takes an argument that says which HTTP status counts as success, but never looks
   at it; each caller checks the status itself. So when a delete passes the create
   constant by mistake, `CreateCloneRCode` (hurl Finding 4.4), nothing goes wrong:
   `DeleteClone` checks for 204 on its own.
8. **Error handling depends on message text** (`errors.go`). If the appliance
   rewords a message, the plugin's behaviour changes and nothing reports an error.
   [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool) would catch this when new hurl runs are added, and [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) would keep
   all message patterns in one catalogue, kept per call, so callers decide on codes, not on text.
9. **A CSI list can resume only after a name.** A page number in a token
   names other entries as soon as the list changes, and the appliance
   cannot start a read at a name, so every read starts at page 0 and costs
   one request per page; the last name returned keeps its place in the
   order whatever else is created or deleted
   ([How lists are paged](#how-lists-are-paged)).

## Priorities
[Priorities](#priorities)

The project values that weigh on this document, in the order they weigh here:

1. **Works with the latest JovianDSS 1 release.**
2. **Security.**
3. **Evidence over documentation.** Paths and types rest on recorded replies,
   not on what the company documentation says a version does.
4. **Bugs stay visible.** A field the plugin reads keeps an exact type and
   fails loudly when changed, or when absent if it is required; a field it does not read cannot fail.
5. **Repeatable checks.** After a new recording the types are read
   again from the spec and the fixtures and the audit are re-run; the diff is
   the report.
6. **Smallest change for the callers, one code path, revertable commits.**
   One set of paths and fixtures; the seven type names and the accessors stay;
   the path move and the types are commits that can be reverted alone.
7. **Right first time.** The whole series is delivered and tested in one go; every check that does not need the lab passes before the live tests, so that no follow-up fixes are needed. The REST layer's structure stays; it is reworked where this document says, not rewritten.
8. **Errors mean what they say.** Every known appliance error has one deliberate meaning; an unknown one stays unknown, logged in full once by the REST module, never guessed into another meaning. Misreading an error, such as taking "not found" for "busy", is a bug, not a fallback.

## Options considered
[Options_considered](#options-considered)

One matter is decided here: the API version, where only one option was
considered. How the client is made compliant, the reply types, the error
mapping, the transport and the paging of the lists, is implementation, decided as questions, with the alternatives they were weighed against where there were some, named in the [Decision](#decision)
where they shape the delivery and kept in the [Decisions](#decisions) table.

#### aopt_rest_api_version_v4_only

**v4 only.** Only one option was considered for the API version: replace
the `api/v3` prefix with `api/v4` everywhere and abandon v3 in favour of
v4, with no fallback and no setting: the appliance still serves v3, but the
plugin no longer calls it. The minimum JovianDSS 1 release is stated in the
documentation and the changelog. Why no other: API v3 is deprecated, as the
[Problem statement](#problem-statement) says; it ends with JovianDSS 1,
JovianDSS 2 serves only v4, and the Proxmox plugin has already moved
([Key observation 1](#key-observations)). Keeping v3 would be a decision to
stop at JovianDSS 1. Carrying both versions, behind a setting or a check at
start-up, would keep a second set of paths and fixtures for answers that are
the same on JovianDSS 1 ([Key observation 2](#key-observations)), against
[priority 6](#priorities). The user ruled both out as never real options, so
neither was weighed. The recordings show
the move to be behaviour-neutral on JovianDSS 1 for every call the plugin
reaches today, and the same fields fail on every version ([Key observations 2 and 3](#key-observations)), which is what lets it
be one commit that changes nothing else.

The price of the one option: a JovianDSS 1 release without v4, or with a
different v4, would stop working after the upgrade, with no fallback
([risk 2](#risks)); the minimum release has to be stated and kept true
([Stage 1](#stage-1-move-to-api-v4)).

How far it reaches: the move itself changes the paths of `pkg/rest`, the
documentation and the changelog; the driver, the controller and the node
stay as they are
([Stage 1](#stage-1-move-to-api-v4)). The work that follows from it,
compliance with v4 and the security fixes made with it, reaches further, as
the questions decide: `pkg/rest` throughout; the driver and the controller
at the places Table 8 lists; of the node, only the lines that log whole CSI
requests; the configuration and its documentation; the `Makefile`; and a new audit tool in
kub-testing ([Affected systems and files](#affected-systems-and-files)). It
leaves alone the node's mount and iSCSI code, everything on the appliances,
and `jdss-rest-api-spec`, which is only read. JovianDSS 2 stays
out: it accepts only token login, so v4 alone does not make the plugin work
there ([Future work: JovianDSS 2 support](#future-work-joviandss-2-support)).
What else is left to other designs and to the tracker is listed in
[Relationship to other designs](#relationship-to-other-designs).

## Decision
[Decision](#decision)

**The plugin's REST client is brought to
[compliance](#what-compliant-means) with the JovianDSS 1 REST API v4. The
plugin uses v4 only: v3 is abandoned in favour of v4, with no fallback and
no setting (the appliance still serves v3; the plugin no longer calls it),
and the minimum supported JovianDSS 1 release is `1.0up33.65410`, the latest JovianDSS 1 release and the one the recordings and the test lab run,
or the newer one the live tests ran on ([question 58](#decisions)). The work
changes `pkg/rest`, the driver and the controller, of the node only the lines that log whole CSI requests, the configuration and its documentation and the `Makefile`, and adds an audit tool to kub-testing; it leaves alone the node's mount and iSCSI code, everything on the appliances, and `jdss-rest-api-spec`, which is only read ([aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only)); the code it adds
or changes brings no new vulnerability, and the existing ones within its
scope are removed. JovianDSS 2 is not supported within this iteration.**

Rationale, from [aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only)
and the [Risks](#risks): v3 ends with JovianDSS 1, and the recordings show
that the move would change no behaviour the plugin reaches there today (priority 1); its price, a release
without this v4 stopping with no fallback ([risk 2](#risks)), is accepted
because no such release is supported ([question 29](#decisions)). Leaving
JovianDSS 2 out keeps the work on JovianDSS 1 (priority 1), at the price of
a plugin that still does not run on JovianDSS 2 until a later design adds
token login ([Future work: JovianDSS 2 support](#future-work-joviandss-2-support)).
The security findings are fixed in the same series (priority 2), at the price
of upgrades that would stop until the certificate is dealt
with ([risk 11](#risks)). Compliance is measured against the
recordings, not the documentation (priority 3); what they never show stays
a risk, not an assumption ([risks 15 and 16](#risks)). Only a read field can
fail, and it fails by name (priority 4), the guard of [risk 9](#risks). A
new recording is a new reading of the spec, checked by the fixtures and the
audit, with the diff as the report (priority 5); that is the guard of
[risks 18 and 22](#risks) and, more weakly, of [risk 17](#risks). One API, one code path, the names the other
modules use kept, and commits that revert alone (priority 6). The series is
delivered and tested once, every check that needs no lab first (priority 7,
[risk 7](#risks)). Every known appliance error means one thing and an
unknown one stays unknown (priority 8). The lists are read as the appliance
pages them, as recorded for the volume list ([question 49](#decisions)), and resume after a name, at the price of a read of
the whole list on every call (priority 3, [risks 26, 27 and 28](#risks)).
Without a guard in this delivery: a
hung call past the CSI deadline ([risk 12](#risks)), the accepted edge of a
volume name of 129 to 132 characters ([risk 13](#risks)), a reply behind the request ([risk 14](#risks)), the time a CSI list call takes on a long list ([risk 26](#risks)) and a list that serves fewer than 100 entries to a page ([risk 28](#risks)).

The decisions below the options were taken as questions and are kept in the
[Decisions](#decisions) table: the reply types are transcribed by an
agent from `jdss-rest-api-spec`'s OpenAPI of the recorded jdss1 v4 API, under
the same names, only a read field can fail, and
the drivers change where they read a field whose type or key changes
(question 38, which replaces 28), the iSCSI `GetSnapshot` taking the snapshot
name it asked for (question 42) and the readers of the per-volume snapshot
lists building again only in stage 3 (question 43); every error known for a call, recorded or known from elsewhere, is mapped
for that call, from its status, `class`, `errno` and message, and an unknown one is logged once in full by the REST module
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue));
a 404 means "not found" whatever its body (question 40), and a GET answered
without data is sent again, `tries` times in all at most (question 39);
each driver lists only its own snapshot kind (question 7); the CSI lists read the whole appliance list once on every call, with `per_page=100`, count for `max_entries` only the entries their reply carries, and resume after the last name they returned (questions 60 to 62, 64 and 66); IDs and names are validated where the CSI parsers read them, with two
entry points left uncovered ([Relationship](#relationship-to-other-designs)), generated volume names
are at most 128 bytes, so that a volume's CSI ID is within the CSI spec's limit for a string, names on the appliance at most 131 characters and target names lower case (questions 20, 21 and 71); the plugin speaks HTTPS
only, TLS 1.2 or higher, with certificate checks on by default (questions 23,
26 and 50); no secret reaches any log at any level, no request body or header
is logged, and no Warn line that could print a REST or CHAP credential is
left (questions 19, for tokens, 36 and 45; [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)); release notes live in
`CHANGELOG.md` (question 24).

## What changes
[What_changes](#what-changes)

Deliver everything here in one go: one series of commits, built, checked and
tested together ([Testing](#testing)). Put the commits in this order: a
`go vet` clean-up ([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply), [impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text)), the two
logging commits ([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)), `per_page`, dead-code removal, the move to v4, response
types (stage 2), behaviour and error mapping (stage 3), robustness and security
(stage 4). Stage 1 alone reaches v4 and stage 2 decodes its replies; the client
is compliant with the recorded v4 API only when stages 3 and 4 are done as
well. Make each commit
build on its own, except those from the stage 2 types commit up to the
stage 3 commit that moves the readers of the per-volume snapshot lists to
their short entries ([question 43](#decisions)), so that if the live test fails,
`git bisect` can find the commit that caused it, skipping those.
The implementing agent writes the code and its unit tests on the analysis
machine, which has no Go, and builds, vets and runs the unit tests on the
lab machine, commit by commit ([question 47](#decisions)).
This document stays Proposed while the series is implemented and tested,
and is accepted with the tested series. Until then, write into its text
whatever the implementation changes in what it states: the read fields of
Table 5 ([question 51](#decisions)) and the minimum release
([question 58](#decisions)) are the changes it expects
(user decision, [question 72](#decisions)).
One thing in the order differs in the series as it is written: the fixtures,
their unit tests and the package comment of stage 2 stand after the first
two commits of stage 3, because those tests pass every recorded error
through the handler of a failed reply, which stage 3 makes.

### First commit: make `go vet` pass in `pkg/rest` and `pkg/controller`
[First_commit_make_go_vet_pass_in_pkg_rest_and_pkg_controller](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller)

#### impl_plugin_rest_no_format_string_from_reply

**Its own commit, first in the series.** [Testing](#testing) step 1 asks that `go vet`,
Go's built-in static checker, report nothing for `pkg/rest`, `pkg/driver` and `pkg/controller` on every commit that builds ([question 43](#decisions)), so vet has to pass there from the first commit of the series on. Since Go 1.24, vet's
printf check reports a format string that is not a constant and has no arguments;
the module is Go 1.26, so the check applies. `go vet -a ./...` on the current code
gives 34 reports, all of this kind and nothing else. Eight are in `pkg/rest`;
fix those eight:

**Table 3 — The `go vet` reports in `pkg/rest`** · `tbl_vet_reports`

| Where | Today | Change to |
|---|---|---|
| `storageapi.go:51` (`getError`) | `bs := fmt.Sprintf(string(body[:]))`: the reply body is the format | `bs := string(body)` |
| `storageapi.go:53` (`getError`) | `l.Warnf(msg)`, and `msg` holds the reply body | `l.Warnf("%s", msg)` |
| `errors.go:173`, `:265` (`ErrorFromErrorT`) | `l.Warnf(msg)`, and `msg` holds a name cut from the appliance's message | `l.Warnf("%s", msg)` |
| `targets.go:150` (`AttachVolumeToTarget`), `:184` (`DettachVolumeFromTarget`) | `l.Debugf(msg)`, and `msg` holds the target name | `l.Debugf("%s", msg)` |
| `targets.go:45` (`GetTarget`), `:76` (`CreateTarget`) | `l.Debugf(fmt.Sprintf("... %s ...", name))` | `l.Debugf("... %s ...", name)`: the format and the name passed directly |

No message would change for text without a `%`. This is also the fix for
[bug_plugin_rest_reply_text_as_format_string](#bug_plugin_rest_reply_text_as_format_string)'s format-string
part: a `%` in an appliance message could no longer garble the message or the log.
In the same commit, add `vet` and `test` targets to the `Makefile`, so that the vet and the tests of [Testing](#testing) step 1 are a command each from the first commit on: `make vet`
runs `go vet -a ./pkg/rest/... ./pkg/driver/... ./pkg/controller/...`, and `make test` has nothing to run until
[impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures). Keep the `-a`
(user decision, [question 76](#decisions)): with Go 1.26.0 the go command
takes a vet result from its build cache, and a package that was vetted
before as a dependency of another one then gets no report when it is vetted
itself. `-a` makes the go command analyse the packages again
([risk 30](#risks)). Change nothing else in this commit
but the controller's format strings, which go with it
([impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text)).
The other 26 reports are in the controller and the node
(`bug_plugin_error_text_as_grpc_status_format_string`): that item takes the
23 of the controller, and the three of the node stay
([Relationship](#relationship-to-other-designs)), so `go vet -a ./...` keeps
reporting those three after this delivery. What must hold:
`go vet -a ./pkg/rest/...` gives no report, `make vet` and `make test` exist, and a `%` in an appliance message or
in a name reaches the error text and the log unchanged.

#### impl_plugin_ctrl_no_format_string_from_error_text

**In the first commit, with the reports of `pkg/rest`** (user decision,
[question 69](#decisions)): fix the 23 reports of `pkg/controller`, all in
`controllerplugin.go` and all of the kind that
[impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply)
describes; they are the controller's part of
`bug_plugin_error_text_as_grpc_status_format_string`. `go test` runs vet's
printf check on a package before it runs the package's tests, and runs no
test when the check reports anything. Stage 3 gives `pkg/controller` its
first test files
([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
[question 63](#decisions)), so with these reports left in, `make test` would
end there in a failed build before a test ran. `pkg/driver` has no report.

- Where the message is a variable or the text of an error, pass it as the
  message and not as the format, `status.Error(code, msg)` in place of
  `status.Errorf(code, msg)`: `:127`, `:134`, `:377`, `:381`, `:386`,
  `:410`, `:565`, `:659`, `:663`, `:676`, `:732`, `:884`, `:893`, `:897`.
- Where the message is built by `fmt.Sprintf` inside the call, pass the
  format and its arguments to `status.Errorf` directly: `:418`, `:422`,
  `:431`, `:433`, `:439`, `:442`, `:445`.
- Where `fmt.Sprintf` wraps a joined list of addresses and nothing else,
  drop it and keep `strings.Join`: `:869`, `:872`.

The status codes stay as they are, and so does every message without a `%`.
What must hold: `go vet -a ./pkg/controller/...` and
`go vet -a ./pkg/driver/...` give no report, `make vet` runs them together with that of `pkg/rest`, and a
`%` in the text of an error reaches the gRPC status unchanged.

### Before stage 1: keep secrets out of the logs
[Before_stage_1_keep_secrets_out_of_the_logs](#before-stage-1-keep-secrets-out-of-the-logs)

**Two commits, right after the `go vet` commit and before everything else**
(user decision): the REST password, a CHAP password and the CSI `Secrets`
must never reach a log, at Debug, at Info or at any other level. The
current code breaks that rule at Debug, where the REST password and the CSI
`Secrets` reach the log, and so would every build of the series made before
these two commits. Five Warn lines print a whole CSI request as well, but
only when a capability check fails, which the fixed list of capabilities
(`controllerplugin.go:63-70`) never lets happen; remove them all
the same ([question 45](#decisions)). With the two commits at the front, every
later build would keep secrets out of the log whatever its level, so a live run
of any of those builds, a cycle of [Testing](#testing) or a
`git bisect` round after a failed one ([risk 7](#risks)), may be made at Debug. The log level is set in
the plugin configuration for a run. Info is the decided default (user
decision), and the current code does not apply it: the shipped controller configuration sets
Info, the `--loglevel` argument defaults to Warning, a configuration without
the key stops the controller, and the node manifests pass Debug. Making Info
the default is startup and manifest work outside `pkg/rest`, handled
separately (kub-testing `ISSUES.md`, MF15).

#### impl_plugin_rest_no_request_logging

**No request bodies or headers in the log** ([security_plugin_secrets_in_logs](#security_plugin_secrets_in_logs), [security_plugin_rest_request_dump_every_call](#security_plugin_rest_request_dump_every_call);
user decision): make `Send` stop logging request bodies (`rproxy.go:113`, `:118`)
and stop dumping the request (`:130-131`), so that neither the CHAP password
nor the `Authorization: Basic` header can reach a log. Remove the dump call,
do not just silence it: today it runs on every request even when nothing is
logged, so removing it would also save that work. Method, path (with its query string)
and status stay logged at Debug, in one line for each reply: [Testing](#testing) step 4 reads `per_page` from
that line. Bodies and headers are left out of the log for now, while the
one-line request log stays; whether a
safe, redacted log of bodies comes back later is a separate decision. What must
hold: a request with a password in its body leaves neither the password nor an
`Authorization` header in the log output, at any level, and every request that gets a reply still leaves one Debug line with its method, its path with the query string, and its status. The
commit lands before any unit test exists, so this condition is checked later
(user decision): its test comes with the fixtures in stage 2
([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)), and until then the commit is checked by
reading it.

Follow the code under [Code segments](#code-segments) for the log line of a
request.

#### impl_plugin_no_secrets_in_logs

**No secrets in any log** (user decision; review [security_plugin_secrets_in_logs](#security_plugin_secrets_in_logs) outside `pkg/rest`): let no
log line print a whole CSI request, because CSI requests carry `Secrets`.
Today they are printed at Debug (`controllerplugin.go:587`, `:626`, `:703`,
`:750`, `:837`, `:914`; `node/publish.go:43`, `:68`), and at **Warn** when the
capability check of a create, delete or publish fails (`controllerplugin.go:479`,
`:546`, `:634`, `:710`, `:857`). That check cannot fail today: the list of
capabilities it reads is fixed (`:63-70`) and holds every capability these
calls ask for, so in practice the secrets reach the log at Debug only.
`ControllerGetCapabilities` logs its request at Info (`:1078`; no secrets, but
the rule is the same). Make the Debug and Info lines log only the fields
needed (volume ID, name, capability). Remove the Warn lines that could print a
REST or CHAP credential, the five above, instead of cutting them down to fields
(user decision, [question 45](#decisions)), so that no later change to the
list of capabilities can bring a secret back at Warn. With [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), no secret would
reach a log at any level. The configuration, which holds the REST password, is
not logged today (`controllerplugin.go:152-153` are commented out); keep it
so. What must hold: no log line of the controller or the node prints a whole CSI
request, at any level, and no Warn line that could print a REST or CHAP
credential is left.

### Before stage 1: send `per_page` on every list
[Before_stage_1_send_per_page_on_every_list](#before-stage-1-send-per_page-on-every-list)

**Its own commit, after the logging commits and before the v4 move.** On JovianDSS 1 the list loop ends today: on v4 a volume list request without `per_page` returns pages of 20, `page` still works, and a page past the end is empty ([How lists are paged](#how-lists-are-paged)); the other lists, and v3, are taken to do the same ([question 49](#decisions)).
On jdss2 v4 a list request without `per_page` ignores the page number and
returns the first page again and again, so CSI ListVolumes would run forever when `max_entries` is not set, and return duplicates when it is ([bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page)). The commit is therefore not a
condition of the v4 move on JovianDSS 1; it is there to make the loop safe against a
server that repeats the first page when `per_page` is missing, as jdss2 v4 does, and it comes before the v4 move so that the v4 commit changes
paths only. Paging with `per_page` works on jdss1, v3 and v4, and on jdss2 v4
(gap [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page)).
Send `per_page=100` with every list call, on v3 and on v4, so that lists
are paged the same way on every version ([rule_rest_lists_paged_as_served](#rule_rest_lists_paged_as_served)).
Add one condition to the stop rule. Today the plugin asks for page after
page until a page comes back empty; keep that rule and also stop when
a page holds fewer entries than `per_page`. Like the empty page, the new condition would need no total (`results`), so it would also end a list that resumes from a token at a later page, and a list shorter than one page would cost one request instead of two. The new condition
takes for granted that the appliance serves 100 entries to a page. The
volume list of jdss1 v4 does, with no cap, and the other lists are taken to
do the same ([question 49](#decisions)); a lower cap would end every longer
list early ([risk 28](#risks)). A server that returned
the same full page again and again would still keep the loop running; no
recording shows one once `per_page` is sent. Both conditions count a page as
the appliance sent it, before the filter by kind of stage 3 drops any entry
([question 7](#decisions)): counted after the filter, a full page that held
the other kind's snapshots would look short, and the list would end early.

What must hold: every list request carries `per_page=100`, on v3 and on v4;
the loop stops on an empty page and on a page that holds fewer entries than
`per_page`, both counted as the appliance sent the page.

### Before stage 1: remove dead REST code
[Before_stage_1_remove_dead_REST_code](#before-stage-1-remove-dead-rest-code)

#### impl_plugin_rest_remove_dead_code_before_v4_move

**Its own commit, after the `per_page` commit and before the v4 move,** so that the
v4 commit touches only code that is used. Remove everything listed
under [dead code](#dead-code) ([rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_request_body_valid](#rule_rest_request_body_valid)). With `DeleteShare` gone, no call of the plugin would meet
[bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class) any more.

One of the functions to remove, `DeleteNASVolumeSnapshot`, is a broken copy of
`DeleteNASSnapshot` (`nas_snapshots.go:67`): both address one NAS snapshot at the
same path, but the copy sends `POST` instead of `DELETE`. Its one caller is the
cleanup of an intermediate snapshot at `nfs_driver.go:322`. Point
that call at `DeleteNASSnapshot`, which `deleteIntermediateSnapshot` and the normal snapshot delete already use
(`nfs_driver.go:193`, `:623`), so that deleting NAS snapshots stays supported, through
one working function ([question 10](#decisions)). The
build would keep working, and **[bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) and [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check) would be
fixed here** without new code. Only that call would be correct after this commit; the
cleanup around it would still not run properly until stage 3 fixes [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) and [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) ([driver bugs](#driver-bugs), [risk 23](#risks)).

Dead code outside `pkg/rest`, such as the old node implementation in
`pkg/node/targetapi.go` and `target.go`, is not part of this commit, with one
exception. `StorageCfg` has one user left, itself dead: `ControllerCfg` in `pkg/controller/config.go`, used only by `GetConfing` there, which nothing calls, and by the unused `cfg` field of the controller (`controllerplugin.go:93`). Remove that file and that field too, because with them left in, the commit would not build
(user decision).

What must hold: `go build ./...` passes with everything under
[dead code](#dead-code) removed, nothing that is called is gone, and the
intermediate-snapshot cleanup deletes a NAS snapshot through
`DeleteNASSnapshot`.

### Stage 1: move to API v4
[Stage_1_move_to_API_v4](#stage-1-move-to-api-v4)

Replace the hard-coded `api/v3` strings with one API prefix, set to `api/v4`: 34
in the current code (plus 3 in comments), 30 after the dead-code commit. Add
no setting and no fallback to v3. **Change no other code in this commit**,
so that if the live test fails, `git bisect` can tell an effect of v4 from an effect
of a fix. No change in behaviour is expected on JovianDSS 1 (hurl Findings 14, 15):
of the calls the plugin would make, only the per-volume snapshot lists answer
differently on v3 and v4, and the plugin does not read them before stage 3
([API versions](#api-versions)).
Of the evidence gaps, [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) (the empty `{}` body the plugin sends with every GET) made
no difference on jdss1, v3 and v4, or on jdss2 v4; [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page) and [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse) became findings of their own,
[bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) (paging), fixed by the [`per_page` commit](#before-stage-1-send-per_page-on-every-list), and [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) (thick volumes), left to its own design ([question 9](#decisions)).

**Abandoning v3:** make the plugin docs say that it needs a JovianDSS release with API
v4, and that the plugin no longer calls v3, which the appliance still serves;
name in the plugin docs and in `CHANGELOG.md` the oldest supported JovianDSS 1 release,
`1.0up33.65410` or the newer one the live tests ran on ([question 58](#decisions)); base the reply types, the audit and the fixtures, in stage 2, on v4
replies; and remove the comments that mention v3.

What must hold: no `api/v3` string is left in `pkg/rest`, every path starts
with the one `api/v4` prefix, and the code changes in nothing else but the
comments that name v3; the plugin docs and `CHANGELOG.md` name
the release the live tests ran on as the oldest supported JovianDSS 1
release ([question 58](#decisions)).

### Stage 2: response types
[Stage_2_response_types](#stage-2-response-types)

Replace the hand-written [response types](#response-types) with types
transcribed from `jdss-rest-api-spec`'s description of the recorded API ([question 38](#decisions)), and so close
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch),
[bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec), [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) and the key of [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) in one commit; return parse errors
instead of swallowing them ([impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors)).

#### impl_plugin_rest_reply_types_transcribed_from_observed_typespec

**Reply types transcribed from the observed API description** ([rule_rest_reply_types_decode](#rule_rest_reply_types_decode); user decision,
[question 38](#decisions)). Have an agent read `jdss-rest-api-spec`'s tracked
OpenAPI for jdss1 v4, one YAML file per path under
`empirical/jdss1/v4/openapi3/paths/`. It is compiled from the observed
TypeSpec, which the item's name refers to, and rewritten only after
`build_typespec.py --check` has validated it against every recorded reply.
In it a field listed under `required` was never seen missing, and the types
are those that were seen. Have the agent write the reply types into
`pkg/rest/types.go`, in place of today's file. There is no generator, no
mapping file and no compiler, and nothing is read from `jdss-rest-api-spec`'s build
folders ([question 31](#decisions)). Open the file with a header
naming the `jdss-rest-api-spec` commit and, for each type, the path file it was read from,
so that a later reading knows what to compare with. Keep the 19 `Resource*` type
names, so that the drivers and the controller keep compiling against
them wherever they read a field the transcribed type still has (below), and add a sub-type wherever a read field sits in an object that no type holds today: `ResourceSnapshotParent`, the `parent` of a zvol snapshot list entry, with its `name` and `type` (user decision); move the accessors and the conversions (`GetSize`, `OriginVolume`, `OriginSnapshot`,
`ClonesNames`, the parsing of numbers sent as strings that the custom
decoders do today) to `pkg/rest/types_methods.go`, with a
`checkRequired()` per type. Split the rest of today's `types.go` the
same way: transcribe the reply sub-types `Disk`, `ShareNFS` and `ShareSMB`
under their names, and move the envelopes `GeneralResponse` and
`ResultEntries` and the two regular expressions the accessors use
to `types_methods.go`. Stage 3 removes both envelopes: its handler reads the
error object of a reply itself, and its list calls return a typed page. `jdss-rest-api-spec` stays read only: the user runs its
tools, and reading its tracked files needs none of them. When the `jdss-rest-api-spec` project adds recordings, read the types again from the path files
that changed between the header's commit and the new one, and review the diff of
`types.go`. Two checks guard the types: the audit compares
`types.go` with the OpenAPI field by field ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)), and
the fixtures, recorded replies chosen to exercise every read field of every
call, are decoded with them by the unit tests
([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures), [risk 17](#risks)).

**Table 4 — Reply type rules** · `tbl_type_generation`

| What | Rule |
|---|---|
| a field the plugin reads | the type the recordings show for that endpoint, exact. A comment on the field marks it **required** or **optional**; where the recordings never show it set, the type is stated by hand, with the reason in the comment |
| a required read field absent from a reply | the call fails with an error that names the field: a `checkRequired()` per type, called after decoding. Which read fields are required is read from the OpenAPI: those no recording shows missing or `null` ([question 32](#decisions)), but for `is_clone`, optional by the user's decision ([question 65](#decisions)). Table 5 lists them, type by type. The zvol snapshot reply by name carries no `name` ([bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key)), so the iSCSI `GetSnapshot` takes the name it asked for ([question 42](#decisions)) |
| an optional read field absent from a reply | the accessor returns the value its comment gives for an absent field, such as `""` from `OriginSnapshot()` for a volume that is not a clone. Which read fields are optional is read from the OpenAPI, not listed by hand: those the recordings show missing or `null` at least once ([question 32](#decisions)), and `is_clone` ([question 65](#decisions)). A later recording that shows a required field missing makes it optional at the next reading, and the required-field tests pin the list as it stands. Table 5 lists them: `origin` of a volume and of a NAS volume, `clones` of a zvol and of a NAS snapshot, NAS `quota`, and from stage 3 `is_clone` of a volume; no other read field is optional |
| a field the plugin does not read | kept as `json.RawMessage`: stored as sent, never able to fail. [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) and the keys of [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) become this |
| the key of a read field | the recorded key: [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) reads `parent.name`, and [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key)'s `type` comes under its own key (stage 3 filters by it) |
| a field never seen set in any recording | `json.RawMessage`, unless it is typed by hand |
| the zvol-or-clone alternatives of a volume reply | one struct with the fields of both; a field that one of them lacks or sends as `null` is optional |
| a read field typed as a single-value enum | its base type: `is_clone`, `false` in the zvol alternative and `true` in the clone alternative, is a `bool`, optional, read from stage 3 ([question 65](#decisions)) |
| one thing typed differently per endpoint, such as snapshot `creation` ([bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch)), a number in the list and a string by name | each type follows its endpoint; the accessor in `types_methods.go` converts, as the decoders do today |
| jdss2 | outside this document; the later JovianDSS 2 design decides how its types join ([JovianDSS 2 support](#future-work-joviandss-2-support)) |

The fields the plugin reads once this document is applied, type by type, as
the OpenAPI marks them at `jdss-rest-api-spec` commit `6149fa9`; the read fields are the
same at `94d0c4f`. A field is optional where the OpenAPI lets it be missing
or `null`, and `is_clone` by the user's decision ([question 65](#decisions)). The path files are those of the GET under
`empirical/jdss1/v4/openapi3/paths/`, where `…/` stands for
`pools/+poolname+/`. The sub-types `ResourceSnapshotShortProperties`, `ResourceSnapshotParent` (new), `ResourceNASSnapshotParent` and `ResourceNASSnapshotShortProperties` are read
through the dotted fields of their types; every other field, and every field
of the types not listed (the pool's sub-types, `Disk`,
`ShareNFS`, `ShareSMB`), is not read and becomes `json.RawMessage`.
Which fields the plugin would read and require is hard to tell before the
implementation, so this list may change with it; a field it comes to read
takes its mark by Table 4's rules ([question 51](#decisions)). Such a change
is written into this table, since this document stays Proposed until the
series is tested ([question 72](#decisions)).

**Table 5 — Read fields per reply type** · `tbl_read_fields`

| Type | Path file | Required read fields | Optional read fields | Read by |
|---|---|---|---|---|
| `ResourceVolume` | `…/volumes/+volumename+.yaml`; the entries of `…/volumes.yaml` | `name`; `volsize`, a number sent as a string; from stage 3, `guid` and `default_scsi_id`, strings ([question 65](#decisions)) | `origin`: `null` on a zvol, a string on a clone; from stage 3, `is_clone`, a boolean, in every recording but optional by the user's decision ([question 65](#decisions)) | the iSCSI `GetVolume` and the volume list: `Name`, `GetSize()`, `OriginVolume()`, `OriginSnapshot()`; from stage 3, the list's check of a name that comes twice |
| `ResourceSnapshot` | `…/volumes/+volumename+/snapshots/+snapshotname+.yaml` | `volsize`, a number sent as a string; `creation`, a date sent as a string | `clones`, missing on a snapshot without clones | the iSCSI `GetSnapshot` and `deleteIntermediateSnapshot`: `GetSize()`, `Creation`, `ClonesNames()`. The reply has no `name` ([question 42](#decisions)); the readers of the per-volume list read `Name` until stage 3 and do not build in between ([question 43](#decisions)) |
| `ResourceSnapshotShort` | the entries of `…/volumes/snapshots.yaml`, and from stage 3 of `…/volumes/+volumename+/snapshots.yaml` | `name`; `parent.name`; `parent.type`, read from stage 3 ([question 7](#decisions)); `properties.creation`, a number; `properties.guid`, a string, read from stage 3 ([question 65](#decisions)) | none | the iSCSI snapshot lists |
| `ResourceNASSnapshotShort` | the entries of `…/nas-volumes/snapshots.yaml`, and from stage 3 of `…/nas-volumes/+datasetname+/snapshots.yaml` | the same five as `ResourceSnapshotShort` | none | the NFS snapshot lists |
| `ResourceVolumeSnapshotClones` | the items of `…/volumes/+volumename+/snapshots/+snapshotname+/clones.yaml` and of `…/nas-volumes/+datasetname+/snapshots/+snapshotname+/clones.yaml` | `name` | none | the snapshot deletes of both drivers, and from stage 3 the cleanups and the NFS `deleteIntermediateSnapshot` ([question 34](#decisions)) |
| `ResourcePool` | `pools/+poolname+.yaml` | `available`, a number sent as a string | none | `GetCapacity` |
| `ResourceTarget` | `…/san/iscsi/targets/+targetname+.yaml` | `active` | none | the iSCSI `PublishVolume` (`iscsi_driver.go:629`) |
| `ResourceNASVolume` | `…/nas-volumes/+datasetname+.yaml`; the entries of `…/nas-volumes.yaml` | `name` | `quota`, `null` without a quota ([bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec), review `bug_plugin_driver_nas_quota_null_size_zero`); `origin`, `null` on a volume that is not a clone | the NFS `GetVolume` and the NAS volume list: `Name`, `GetSize()`, `OriginVolume()`, `OriginSnapshot()` |
| `ResourceNASVolumeSnapshot` | `…/nas-volumes/+datasetname+/snapshots/+snapshotname+.yaml` | `name`; `creation`, a date sent as a string | `clones`, missing on a snapshot without clones | the NFS `GetSnapshot`, and until stage 3 `deleteIntermediateSnapshot` and the per-volume NAS list: `Name`, `Creation`, `ClonesNames()`. The reply has no `volsize`, so `GetSize()` stays 0, as today |
| `ResourceShare`, from stage 3 | `shares/+sharename+.yaml` | `path` | none | the NFS creates, when a share create answers "already exists" ([question 57](#decisions)) |

Change the callers at the thirteen lines that read a field whose type
or key changes: read `Creation` through an accessor (`controller/listers.go:52`, `:82`,
`:124`, `:148`; `controllerplugin.go:679`, `:802`; `driver/iscsi_driver.go:456`;
`nfs_driver.go:591`), the list entry's `Parent.Name` instead of a `Volume` key
the appliance never sends (`listers.go:54`, `:56`; `driver/restwrappers.go:40`),
the pool's `GetAvailable()` (`controllerplugin.go:1026`); and take the name of a
zvol snapshot in the iSCSI `GetSnapshot` from its request, the name it asked
for, `sd.SDS()`, since the reply by name has none (`iscsi_driver.go:454`;
user decision, [question 42](#decisions)). Fifteen more lines read `Name` of
`ResourceSnapshot` on entries of the per-volume snapshot lists
(`iscsi_driver.go:211`, `:217`, `:224`, `:279-281`; `nfs_driver.go:315`,
`:322`, `:329`, `:418-420`; `listers.go:126`, `:128`; `restwrappers.go:35`).
The transcribed `ResourceSnapshot` has no `Name`, so they do not build until
stage 3 moves them to the short entries
([impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow)),
and the commits between do not build on their own: the series is tested
once all stages are in (user decision, [question 43](#decisions)).

What must hold: every fixture of a call the plugin makes decodes with the
transcribed types; each read field is required or optional as the OpenAPI
shows it, but for `is_clone`, optional ([question 65](#decisions)); a required read field absent from a reply
fails the call with an error that names it, in a fixed text that [Testing](#testing) step 4 searches the log for; the 19 type names are unchanged,
and the drivers and the controller build with changes at the thirteen lines,
the readers of the per-volume snapshot lists building again with stage 3;
the header of `types.go` names the `jdss-rest-api-spec` commit and the path file
behind each type.

Follow the code under [Code segments](#code-segments) for a reply type and
its checks.

#### impl_lab_rest_audit_tool

**Audit tool.** Turn the two prototypes into one tool, `kub-testing/tools/rest-type-audit.py`.
Make it build the compliance matrix from jdss1 v4's
`empirical/jdss1/v4/INDEX.md` ([rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted), [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled)), with jdss2 v4 beside it as
evidence, and compare the plugin's request
keys with the company `-src` schemas ([rule_rest_request_body_valid](#rule_rest_request_body_valid)); prototypes:
`kub-testing/tmp/rest-audit/compliance_matrix.py`, `compliance_bodies.py`. Make it
also list every recorded error of a plugin call that [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s catalogue does
not cover, read from the `-scenarios.json` files of the plugin's calls ([rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled)), and every one of them that came without an `errno` or a `class`, which those files mark ([rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled)). Have it list apart the catalogue entries marked "not in the spec", for the user to pass on to the `jdss-rest-api-spec` project. Have it read the plugin's list of calls and success statuses
from `pkg/rest` (`compliance_matrix.py` has them typed in by hand). Run it on the
analysis machine, where `jdss-rest-api-spec` is, after each new hurl run the `jdss-rest-api-spec` project adds and after each plugin change in `pkg/rest`. Have it only read `jdss-rest-api-spec`, never write to it. Make it also check the reply types ([rule_rest_reply_types_decode](#rule_rest_reply_types_decode); user decision): it reads the
structs and their `json` tags from the source of `pkg/rest/types.go`, which
needs no Go, and compares each type with the reply schema of the OpenAPI
path file that the header of `types.go` names for it. For every field with a Go type it reports a key
the schema does not have, a type that differs from the one recorded, and a
required or optional mark that differs from the schema's `required` list and
`null` alternatives; it also reports a property of the schema that the type
does not hold. That is the command that shows whether the transcribed types
match the spec; the unit tests on the fixtures show, in Go, that they decode
([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)). What must hold: run on the transcribed types, the type check reports no
difference from the OpenAPI at the commit the header names but Table 4's
deliberate changes, which the implementation decides how to treat
([question 51](#decisions)), and any other type or mark changed by hand makes
it fail; run on the current code, the tool gives the
[compliance matrix](#compliance-matrix) and its status per rule from `jdss-rest-api-spec` and `pkg/rest`, with no list of calls typed in by hand; it reports every
recorded error of a plugin call that the catalogue does not cover for that call; it lists the entries marked "not in the spec" apart, not as faults; and it
writes nothing into `jdss-rest-api-spec`.

The tool as it is written ([question 51](#decisions)): it reads each call's
method, path, success statuses, catalogue and types from the source of
`pkg/rest`, and lists Table 4's deliberate changes apart, as notes that are
no faults: `is_clone`, optional by decision where the OpenAPI has it
required, and a key that is not read and that one of two replies sharing a
type lacks, as `full_name` and `is_clone` of the NAS clones reply. A status
that a call accepts and no recording shows is a note too. It exits with
status 1 when a check finds a fault.

#### impl_plugin_rest_return_parse_errors

**Return parse errors** ([bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors)). Two custom decoders swallow a
parse error today and leave the field at zero, so a pool `available` that does
not parse reads as a full pool. With [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) the conversions move to
`types_methods.go`: make them return the error instead, and replace the snapshot decoder
with the transcribed type and its accessor. What must hold: a pool
`available` or a snapshot `creation` that does not parse fails the call with
an error naming the field, in a fixed text that [Testing](#testing) step 4 searches the log for, and is never read as zero.

Follow the code under [Code segments](#code-segments) for the conversion of
a number sent as a string.

### Stage 3: behaviour and error mapping
[Stage_3_behaviour_and_error_mapping](#stage-3-behaviour-and-error-mapping)

Fix the rest of [response content](#response-content-and-defaults), of the [error messages](#error-messages) and of the [logic bugs](#logic-bugs), and the [driver bugs](#driver-bugs), and make the driver changes below:

- Lists and checks that come back empty or wrong. [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) (the iSCSI
  "snapshots of this volume" request is an empty stub): implement the request.
  [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) (the NFS version asks the wrong path): fix the path.
  Make their callers read what the two lists then return as it
  is sent: [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow).
  [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name) (the "target not found" pattern is a test string that never
  matches): use the real message. What must hold: the recorded 404
  `Target <name> not exists.` maps to the target-not-found code. [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) (the
  snapshot kind is read from the wrong key): read `type`, and filter by it: make the
  iSCSI lister drop entries whose parent is a NAS volume and the NFS lister
  those whose parent is a zvol, and keep other values
  ([question 7](#decisions)).
  ([bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) belongs to stage 2, with [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch);
  [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) to the dead-code commit.)
- Error messages the plugin does not recognise. [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised) (a repeated
  zvol or NAS snapshot create): accept `Snapshot ... already exists.` [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)
  (two different 404 replies on detach are treated the same): tell them apart;
  `not attached` means the LUN is already gone, so go on and delete the target;
  `Target ... not exists` means nothing is left to do. [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised) (NFS
  delete of a volume with clones): accept `dataset` as well as `zvol`.
  [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) (NFS CreateVolume retry): map `Share ... already exists`
  to "exists". Have the three NFS creates that make a share (`nfs_driver.go:140`,
  `:181`, `:285`) then read it with a new call, `GetShare`
  (`GET shares/<name>`, [Table 5](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), and take it as their own only when its
  `path` is the one they send, `<pool>/<name>/<name>`; a share with another
  path fails the create ([question 57](#decisions)). The recorded replies
  give `path` as it was sent, for shares of a dataset's root; a path one
  level down, as the plugin sends it, is not recorded. What must hold: a
  create that finds the share takes it as success only when the share's
  `path` is the one it sends (kub-testing T11.8). [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised)
  (NFS delete of a clone that is already gone): map to "not found".
- Make the NFS `deleteIntermediateSnapshot` (`nfs_driver.go:184-228`) read
  a snapshot's clones before it deletes the snapshot, by the rule of the
  cleanup (user decision, [question 34](#decisions)): `GET .../snapshots/<s>/clones`
  first, the delete only when that reply lists no clone, and nothing
  deleted when the call fails. `CreateVolumeFromVolume` calls the function
  when the create of the intermediate snapshot is answered "already exists"
  (`nfs_driver.go:254-255`), an answer that is not recognised today and
  would be once [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised) is fixed, and when the
  clone create fails (`:264`). The function deletes first and looks at the
  clones only after a refusal (`:193-219`), and the appliance does not
  refuse: it deletes a NAS snapshot that has a clone, and the clone with it
  ([risk 23](#risks)). With a clone listed, let the function go on as it
  does today after a refusal. The iSCSI function keeps deleting before it reads the clones, because
  the appliance refuses the delete of a zvol snapshot that has a clone.
  What must hold: the NFS driver deletes no intermediate snapshot before
  the appliance has answered that it has no clones. Follow the code under
  [Code segments](#code-segments) for the NFS `deleteIntermediateSnapshot`.
- [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised): give each out-of-space reply an exact entry, matched on status, `class`, `errno` and message pattern like every other entry
  (user decision): on the zvol create 400, `BadRequest`, no errno, `New zvol size(...) exceeds available space`;
  on the NAS volume create 500, `ZfsCmdError`, errno 4, `cannot create ...: out of space`; and the entries not in
  the spec that [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) lists. The zvol snapshot and clone creates get none ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [question 59](#decisions)).
- [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised), REST side only: give a missing pool its own code, so that
  a caller can tell it from an unknown error (on the
  snapshot and clone creates its error cannot be told from a missing
  parent's, [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap),
  so take it there as the parent missing, [question 48](#decisions)); map "name too
  long" to `RestErrorArgumentIncorrect`, a code the drivers return
  today but the mapping in `errors.go` never produces; and on v4 map the snapshot list of a missing parent to "not found".
  The caller that turns every zvol `DeleteClone` error into `Busy`
  (`iscsi_driver.go:520`) is a finding of its own, [bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy](#bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy),
  fixed by [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) below.
- [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status): today the code decodes the body first and looks at the HTTP
  status only afterwards, so a reply without a JSON body (a plain-text 401 or 404,
  or a 204 with no body at all) is reported as a parsing problem or, on the calls that pass it to `getError`, as a failure to send (`RequestMalfunction`). The fix: check
  the status first, which the handler of [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) below needs too; when the body is not
  JSON, map the status itself (401 and 403 to a new "authentication failed" code
  that names the endpoint, 404 to "not found") and keep the first bytes of the body
  in the message. A wrong password would then say so. The JovianDSS 2 design would
  get its "token expired" signal from the same code. What must hold: a
  plain-text 401 or 403 gives the authentication-failed code with the endpoint named,
  a plain-text 404 gives "not found", and none is reported as a parsing problem or a failure to send. Follow the code under [Code segments](#code-segments) for a
  failed reply.
- [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check): give a GET answered without data a handler of
  its own (user decision, [question 39](#decisions)). A 204 has no body, so
  it gives the caller nothing to read: no volume, no pool, no list. Five
  GETs name 204 as success today, in a branch the code cannot reach. Make one
  function in `pkg/rest`, which every GET the plugin sends passes
  through, take such a reply and send the same GET again after a
  pause, until a reply carries data or the request has been sent `tries`
  times in all. Only a GET is sent again: a DELETE answered 204 without a
  body has succeeded, as it does today, and is never repeated. `tries` is the existing setting of the REST endpoint
  configuration, 3 when it is not set (`rproxy.go:260-262`), which nothing
  reads today ([bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover)):
  3 means the first request and two more, 1 means no second request. Take the
  pause from a new setting beside it, `retrydelay`, a duration written
  like `idletimeout`, 1 s when the configuration does not set it
  ([question 55](#decisions)). Handle a reply
  other than 204 to a repeated request as if it had come
  first. When a later request gets its data, log one line
  at Warn with the method, the path and how many requests were answered
  204 before it: no recording shows a 204 on a GET, so the first one must
  leave a trace (user decision). When the last reply is a 204 too, fail the call with an
  error naming the status and the number of requests sent. Never turn a 204
  into an empty object or an empty list: a caller would read those
  as a volume of size 0, a full pool, or a snapshot without clones, which
  the cleanup would then delete ([risk 23](#risks)). What must hold: a GET
  answered 204 is sent again after `retrydelay`, `tries` times in all at
  most; a reply with data to any of them is returned as if it were the
  first, and a call that got its data after a 204 leaves one line at Warn;
  and 204 to all of them is an error naming the status, never an
  empty result. Follow the code under [Code segments](#code-segments) for
  the GET sent again.
- [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay): remove the 202 branch, so that a 202 is treated as an error
  like any other unexpected status ([risk 6](#risks)). What must hold: a 202
  on `CreateTarget` is an error naming the status.
- The review finding [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error): `cleanIntermediateSnapshots` checks the type of the
  snapshot list before it checks whether the listing failed, so a real error such
  as "not found" turns into "unknown failure"; and the NFS version expects zvol
  snapshots although its lister returns NAS snapshots
  (`[]jrest.ResourceNASVolumeSnapshot`), so it always fails on the type check.
  Fix both; the type the NFS cleanup takes is the NAS short entry
  of [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow).
  Without the second fix, the NFS cleanup would never reach
  the call that the dead-code commit repairs ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete),
  `nfs_driver.go:322`). The NFS part is checked by the NFS cycle ([Testing](#testing) step 5), and the driver checks are kept in kub-testing (T11).
- **`force_umount: false` on the iSCSI deletes** (user decision,
  [question 56](#decisions)). Make the five deletes that send `true` today, of a
  volume (`iscsi_driver.go:240`), of a snapshot (`:487`), of an intermediate
  snapshot (`:102`, `:214`) and of a snapshot's clone (`:516`), send
  `false`. The plugin detaches a zvol from its target only by the call made for
  that, `DettachVolumeFromTarget`
  (`DELETE pools/<pool>/san/iscsi/targets/<target>/luns/<zvol>`), which
  ControllerUnpublishVolume makes (`iscsi_driver.go:658`); no delete relies
  on the flag to detach it. On idle resources the flag made no difference in
  any recorded delete ([impl_plugin_rest_request_body_keys_checked_against_schema](#impl_plugin_rest_request_body_keys_checked_against_schema));
  with a client still connected it is untested ([risk 8](#risks)). What must
  hold: no request of the plugin sends `force_umount: true` (kub-testing
  T11.9).
- **The paging of the CSI lists** (user decision, [question 60](#decisions)):
  [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) below. It comes
  after the per-volume lists send their requests, and delivers the filter by
  kind with it: a page is counted as the appliance sent it, and the other
  driver's snapshots are dropped after that count and before the cut, a
  place that only the paging has.

#### impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow

**Decode the per-volume snapshot lists into the short entry types, and make
their callers follow.** Close [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name). Once the two listers above
([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) send their
requests, decode their replies into `ResourceSnapshotShort` and
`ResourceNASSnapshotShort`, and change these callers in the same
commits:

- `ListVolumeSnapshots` in both drivers: the entry type of the list and of
  its token function (`RestSnapshotEntryBasedID`,
  `RestNASVolumeSnapshotEntryBasedID`, `driver/restwrappers.go`), which
  [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) then replaces with its key;
- `completeListResponseFromVolumeSnapshots` (`controller/listers.go:111-170`):
  the creation time comes from `Properties.Creation`, as in
  `completeListResponseFromSnapshotShort`; lines `:124` and `:148`, which
  stage 2 moves to an accessor, change again here;
- `cleanIntermediateSnapshots` in both drivers (`iscsi_driver.go:212`; `nfs_driver.go:316`), and `deleteLUN`, which names the clones of the snapshots the cleanup leaves (`iscsi_driver.go:282`; `nfs_driver.go:421`): both call `ClonesNames()` on each entry, and a short entry has no clones to give. Make the cleanup ask the
  appliance instead, with `GET .../snapshots/<s>/clones` for each
  intermediate snapshot (`GetVolumeSnapshotClones`,
  `GetNASVolumeSnapshotClones`: the calls both drivers already make when a
  snapshot delete is refused), and delete the snapshot only when that reply
  lists no clone. When the call fails, nothing is deleted and the cleanup
  returns the error ([question 34](#decisions)). It costs one more request
  per intermediate snapshot. Have the cleanup hand back, with each snapshot it keeps, the clones it read, so that `deleteLUN` names them in its message as it does today.

What must hold: a fixture of each per-volume list decodes into its short
type with `name`, `parent.name`, `parent.type`, `properties.creation` and `properties.guid` set, no code
reads a field of the full object from a list entry, and the cleanup deletes
no snapshot before the appliance has answered that it has no clones.

Follow the code under [Code segments](#code-segments) for the delete of a
snapshot without clones.

#### impl_plugin_csi_list_paging_resumes_after_last_name

**Read the whole appliance list on every call of ListVolumes and of ListSnapshots without a `snapshot_id`, and resume after the last name returned** ([questions 60 to
62](#decisions)). Close
[crash_plugin_driver_list_token_parse_writes_through_nil_pointer](#crash_plugin_driver_list_token_parse_writes_through_nil_pointer),
[bug_plugin_driver_list_resume_repeats_last_entry_returned](#bug_plugin_driver_list_resume_repeats_last_entry_returned),
[bug_plugin_driver_list_resume_compares_base64_forms_not_names](#bug_plugin_driver_list_resume_compares_base64_forms_not_names),
[bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages](#bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages),
[bug_plugin_driver_list_resume_drops_page_missing_token_entry](#bug_plugin_driver_list_resume_drops_page_missing_token_entry),
[bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume](#bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume)
and [bug_plugin_ctrl_list_volumes_negative_max_entries_answered_internal](#bug_plugin_ctrl_list_volumes_negative_max_entries_answered_internal),
and bring the lists to [rule_rest_lists_paged_as_served](#rule_rest_lists_paged_as_served). A page number
names other entries as soon as the list changes, while a name keeps its
place in the order ([Key observation 9](#key-observations)). A page kept in
the token beside the name, as a hint, would read fewer pages, but it rests
on the appliance's order of names, which is recorded for the volume list
only ([question 61](#decisions)).

- **The read.** Make `getResourcesList` (`driver/common.go:31-86`) read
  the list once on every call, whatever the token: from page 0 to the last
  page, with `per_page=100` ([question 62](#decisions)) and the stop rule
  of the [`per_page` commit](#before-stage-1-send-per_page-on-every-list).
  Do not read it again when it changed during the read
  ([question 64](#decisions)); an entry can then come twice or be missed
  ([risk 27](#risks)). A second read when `results` changed during the
  first would notice a single create or delete during the read, at the
  price of another read of the whole list. When a name comes twice in the
  read, compare the two entries on `default_scsi_id`, `guid` and
  `is_clone`, as far as the list has them: all three in the volume list, the snapshot's
  `properties.guid` in the four snapshot lists, and none in the dataset
  list, so that there a name that comes twice is always dropped once
  ([question 65](#decisions)). The three identify the object and do not
  change with writes, as `used`, `written` and `available` do: compared on
  every field, the two entries of a volume in use would both stay.
  `is_clone` may be missing from an entry: count two entries as the same
  in it when both lack it or both hold the same value. When they are the
  same, drop the entry read first, since the later one holds the newer
  state, a new size for instance. When they differ, keep both: two
  snapshots of one name on two volumes, or an object deleted and made
  again under its name during the read.
- **What `max_entries` counts.** Count only the entries the reply carries
  ([question 66](#decisions)). The appliance's lists also hold entries
  the plugin leaves out: objects whose
  names the plugin did not make, which the controller skips today with a
  warning (`controller/listers.go:54-63`, `:84-93`, `:126-130`,
  `:150-154`, `:186-189`, `:232-235`), and the other driver's snapshots,
  which stage 3 drops, since both all-snapshot lists return zvol and NAS
  snapshots together ([question 7](#decisions)). Today the driver cuts the
  list at `max_entries` before the controller skips anything, so a call
  for 10 entries can return 7, or none and a `next_token`; once every call
  reads the whole list, each such call would cost a whole read
  ([risk 26](#risks)). Run both checks in the driver, therefore, before
  the cut, so that a call returns `max_entries` entries whenever that many
  follow and its token names the last entry the caller got. Keep the
  warning of the name check ([question 37](#decisions)).
- **The order and the cut.** Sort the entries by a key, in byte
  order: the volume's name; the snapshot's name in the list of one volume's
  snapshots; the parent's name and then the snapshot's in the list of all
  snapshots, so that two snapshots of one name on two volumes stay apart.
  Return the first `max_entries` entries from the place the token gives (below), all of them when `max_entries` is 0; the entries the token was made from need not exist any more. Keep two entries of one key that both stay in
  the order of the read; when the cut falls between them, do not return
  the second, since the next call starts after that key.
- **The token.** Put in it a version, the list it was issued for (the
  volumes, all snapshots, or the snapshots of one volume, for which a hash of the volume's name stands) and a bound: the shortest string that sorts after the key of the last entry returned, or is equal to it, and before the key of the next one; the next call starts after the bound. Keep the token one ASCII string of at most 128 bytes, the limit CSI sets for a string field, with no page and no `_dc`: a key alone can be longer, since an appliance name can have 131 characters and a key of the list of all snapshots holds two names. Where even the bound does not fit, because the two keys share a long beginning, put in its first bytes and a hash of each of the two keys; the next call then starts at the next entry or after the last one, whichever it finds, and after those first bytes when it finds neither ([risk 29](#risks), [question 68](#decisions)).
  Write a byte of a name that only a name made by hand can hold, anything
  but printable ASCII, and `!`, `%` and `~`, into a key as `%` and two hex
  digits, so that a token is printable ASCII whatever the appliance holds;
  a name the plugin made is written as it is.
  Set `next_token` only when an entry follows the last one returned.
  Answer a `starting_token` that does not parse, has another version or
  belongs to another list with `Aborted`, as CSI prescribes for an invalid
  token, and let no input make the parse panic. Answer a negative
  `max_entries` with `InvalidArgument` in both calls.
- **What goes.** Remove these: `_dc` in the list requests (`pagedcSuffix`,
  `rest/entrielisters.go:34`), which changed nothing in the gap runs
  ([gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page)); the token's page, and its `_dc` value, which `NewCSIListingTokenFromBasedID` loses when that value is not 0
  (`snaptoken.go:60-62`); the base64 ID functions
  (`driver/restwrappers.go:29-57`); and `snapshotToken` with
  `NewSnapshotTokenFromStr` (`snaptoken.go:115-170`), which nothing calls
  and which reads `parts[3]` after checking that there are three parts; and
  the `ResultEntries` envelope with the type assertion each driver made on
  it, since the six list calls of `pkg/rest` return a typed page.

Leave ListSnapshots by `snapshot_id` as it is: it reads that one snapshot
and is not paged. `cleanIntermediateSnapshots` reads one volume's
snapshots through the same read, with no token and no limit.

From this item on, make `make test` also run
`go test ./pkg/driver/... ./pkg/controller/...`: the paging is the one part
of the driver and controller code with unit tests in this delivery, since
csi-sanity resumes no list here ([question 63](#decisions)). `go test` vets
a package before it runs its tests, and both packages pass that vet since
the first commit
([impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text)).
The series tests more of the two packages than the paging: the volume
delete, the cleanup of intermediate snapshots, the clone deletes and the
share check of both drivers run against a fake appliance
(`net/http/httptest`), which records the requests each flow sends, and the
controller's answers to a negative `max_entries` and to a foreign token are
tested with a driver that stands in for the real one. `make test` runs
`pkg/common` as well, from the configuration item of stage 4 on.

What must hold: calls that start with no token and follow `next_token`
return every entry of a list that does not change, each exactly once and in key order, for
any `max_entries`; an entry that exists through the whole sequence is
returned once while other entries, the two the token was made from among them, are created or deleted between the calls, but in the case of [risk 29](#risks); every `next_token` is ASCII of at most 128 bytes; `next_token` is empty exactly when no entry
follows; a token the plugin did not issue for that list gets `Aborted`, a negative `max_entries` gets `InvalidArgument`, and no token
makes the controller panic; a call returns `max_entries` entries whenever
that many of its own follow; every such call reads the list once, from page 0,
with `per_page=100` and no `_dc` on each request; of two entries of one
name in a read that are the same in the fields of
[question 65](#decisions), the reply carries the one read later, and of
two that differ, both, unless the cut falls between them ([questions 63 to 66](#decisions)).

Follow the code under [Code segments](#code-segments) for the read, the
merge of its pages, the order and the cut, and the token.

#### impl_plugin_rest_error_catalogue

**Every known error handled, one path for the rest** (user decision): the
fixes above ([bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised) and [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status)) are not enough on their own: the next
reworded message would start the same trouble again. The plugin needs one way of
handling errors.
- **A catalogue of known errors.** Map on purpose every error recorded for a call
  the plugin makes, from its status, `class`, `errno` and message
  pattern to one `RestError` code. The source is `jdss-rest-api-spec`:
  the `-scenarios.json` file of each call
  (`empirical/jdss1/v4/typespec/<path>/<METHOD>-scenarios.json`), which gives
  for every error seen its status, `class`, `errno`, a pattern for the message
  and the kind of error, whose schema is `openapi3/components/schemas/<Kind>.yaml`, and says which of those fields the appliance did not
  send; `ERROR-FIELDS.md`, `STATUS-CODES.md` and the recorded bodies stand
  beside it. Have the agent transcribe the catalogue from these files, as it
  does the reply types. An entry matches on the HTTP status, the `class` and
  the `errno`, a missing `errno` and an `errno` of `null` being two
  different things, and on the message pattern of the error's schema, not on
  a text of the plugin's own (`jdss-rest-api-spec`'s guidance). Make the type that
  decodes an error body therefore keep three cases of `errno` apart:
  a number, `null`, and no key at all. Go's decoder reads the last two as
  the same, so the field is decoded from its raw JSON with a mark that says
  whether the key was present; a unit test covers the three cases, with
  bodies written by hand for `null`, which no recording of the plugin's
  calls shows yet (user decision). The catalogue is kept per call, as the scenarios
  files are: a failed reply is checked against the errors recorded for that
  call, not against every known error (user decision). The rule: when the
  storage can send an error for a REST resource, the call on that resource
  has an entry for it; when it cannot, the call has none. A snapshot
  create can be answered with "volume not found", so it handles that; it is
  never answered with a missing virtual IP or "target already exists", so it
  does not. What the storage can send for a resource is, first, what its
  scenarios file lists, a reply such as `123 is not of type 'string'`
  included: a later change of the code could cause it (user decision). The
  spec aims to hold every reply and every error, but its recordings are
  incomplete and still growing, so an error known from elsewhere counts too:
  from today's code, from the jdss1 v3 or jdss2 recordings of the same
  resource, from testing, or from another implementation. What the storage
  sent once, it can send. Such an error keeps its entry, marked "not in the
  spec", and is pointed out to the user so that it can be added to the spec
  (user decision). Handled does
  not always mean a code of its own. Take the code that an entry gives, the
  one `GetCode()` returns, from the current code (user decision,
  [question 77](#decisions)): it is the code that the call returns today for
  that reply, Unknown included, and for a call that this document adds, the
  code that today's mapping in `errors.go` gives the reply. Give an entry
  another code only where this document decides one: the fixes listed at
  the head of this stage, and a 404, which is "not found"
  ([question 40](#decisions)). So a caller is told nothing
  new about a reply it got before, but for those. No error of a call's list is left without an entry (user decision): this is what [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) asks for, and, for the errors without an `errno` or a `class`, [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled). Compile each pattern with Go's `regexp` in a unit test, because the `jdss-rest-api-spec` project states them for Python and JavaScript. The entries "not in the spec" known today: the two patterns of today's code that no recording supports
  ([evidence gaps](#evidence-gaps): `Resource ... already exists.` with errno 1 or 100,
  and `Zfs resource: ... not found in this collection` without an errno), most likely from an older JovianDSS release;
  out of space on the NAS snapshot create, errno 1000, which no jdss1 v4 scenarios file lists yet;
  and `cannot create ...: out of space` with errno 4 on the zvol create, recorded in the earlier runs and not on jdss1 v4
  (`POST volumes` lists `New zvol size(...) exceeds available space` and `POST nas-volumes` lists `cannot create ...: out of space`)
  ([bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [risk 15](#risks)).
  [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool) lists these entries as to be reported to the `jdss-rest-api-spec` project, not as faults.
  They have no recorded reply on jdss1 v4, so no fixture, and this delivery does not test them: that needs a test
  server that answers with such errors, which is outside this delivery (kub-testing `ISSUES.md`, T13; user decision).
  The zvol snapshot and clone creates get no out-of-space entry: no run has recorded their reply, so nothing
  exact can be written, and such a reply is Unknown, logged at Error. Their entries are left to another
  record ([question 59](#decisions); kub-testing `ISSUES.md`, F30).
  Until T13's test server exists, give each entry marked "not in the spec" a `TODO` comment in the code that names T13 and says its test is missing.
  On the snapshot and clone creates a missing pool cannot be told from a missing parent
  ([bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap), [risk 25](#risks)): take an error that can be either as the
  parent missing, the volume on the snapshot creates and the snapshot on the clone creates,
  never as the pool missing, and give its entry a `TODO` comment saying that the
  missing parent cannot be told from a missing pool ([question 48](#decisions)).
- **How the catalogue is written.** Write it into one file,
  `pkg/rest/error_catalogue.go`, as the reply types are written into
  `types.go`, with the `jdss-rest-api-spec` commit in its header. In the
  scenarios files a kind of error has one status, one `class`, one `errno`
  and one pattern whatever the call, so write each kind once, as an
  `errorKind`, and for each call the list of its kinds with the code each one
  means for that call (`kind.means(code)`): 95 kinds and 210 entries over
  the 32 calls at commit `434c9cd`. The entries of a call are tried in
  order and the first that matches decides, so put an entry whose pattern
  also matches the message of another after that other. Two kinds have no
  pattern in the spec, `ValidationError_6317af` and `ValidationError_7d2ccd`:
  match them on status, `class` and `errno` alone. Three kinds of the target
  create cannot be told apart by status, `class`, `errno` and pattern,
  `TargetNameConflictErrorTargetWithNameIqnOrgHurltest`,
  `TargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4` and
  `TargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7`, and the
  first holds the IQN prefix of the hurl tests in its pattern: all three
  mean "exists", a reply gets the second code of the first of them that
  matches, and the three are reported to the `jdss-rest-api-spec` project
  as one error. Keep the message of an appliance error on one line in the
  `RestError`: some messages list the datasets in the way on lines of their
  own.
- **The entries that get another code than before**, 63 of the 210; the 147
  others keep the code their call returned, 107 of them Unknown:
  - a missing pool, `RestErrorResourceDNEPool` (new, 17): the kinds
    `ItemNotFoundErrorPoolDoesNotExist`, `ItemNotFoundErrorPoolNotFound`,
    `ItemNotFoundErrorTargetPartitionNotExists`,
    `ZpoolError1GivenZpoolDoesnTExists`,
    `ZfsCmdError0CannotOpenDatasetDoesNotExist_96e556` on the NAS volume
    create, and `ZfsResourceError1ZfsResourceNotFoundInThis_597b56`, whose
    message names `<pool>@<pool>`, on the calls that are not creates;
  - a name too long, `RestErrorArgumentIncorrect`:
    `ZvolValidationError1ZvolNameTooLongZvolName`,
    `DatasetValidationError1DatasetNameTooLongDatasetName`, the four kinds
    `ZfsCmdError1000CannotOpenNameIsTooLong`, the two kinds
    `ZfsCmdError201CannotCreateSnapshotNameIsToo` and
    `OSError36ErrnoFileNameTooLongNfs`. `Target name is too long.` shares
    the kind `ValidationError_7d2ccd`, which has no pattern, with other
    messages, and stays Unknown;
  - `ItemNotFoundErrorTargetNotExists` is `RestErrorResourceDNETarget` on
    the four target calls that get it, and
    `ItemNotFoundErrorVolumeNameIsNotAttachedTo` is
    `RestErrorResourceDNELun` (new, 18);
  - `ItemNotFoundErrorResourceNotFound`, the five `NotFound` kinds of the
    NAS clone delete, and on the snapshot and clone creates the two kinds
    that can be the pool or the parent, are `RestErrorResourceDNE`;
  - `ZfsResourceError5SnapshotAlreadyExists` and
    `Unclassed1ShareAlreadyExists` are `RestErrorResourceExists`,
    `ZfsOeErrorInOrderToDeleteADataset` is `RestErrorResourceBusy`, and
    `ZfsCmdError4CannotCreateOutOfSpace` is `RestErrorOutOfSpace`;
  - a reply without a JSON error body: 401 and 403 are `RestErrorAuthFailed`
    (new, 19), 404 is `RestErrorResourceDNE`.

  Give `Error()` a text for every code, so that no known error is printed
  as "Unknown internal Error".
- **The entries not in the spec, as they are written.** Five: the old
  `Resource ... already exists.` with errno 1 and with errno 100, on the six
  creates of volumes, clones and snapshots; the old `Zfs resource: ... not
  found in this collection` without an errno, on the fifteen calls that list
  a kind of that message; `cannot create ...: out of space` with errno 4 on
  the zvol create; and `cannot create snapshots : out of space` with errno
  1000 on the NAS snapshot create. No recording gives the status or the
  `class` of the two old patterns, so their entries take any. The two
  out-of-space entries are 500 with the class `zfslib.wrap.zfs.ZfsCmdError`,
  as the jdss1 v3 and jdss2 recordings show them. They stand after the
  entries of the spec in each call, and an error they match gets the code
  of the layer of its class as its second code.
- **Codes that say what happened, per kind**, as the Proxmox plugin's `jdssc`
  does (`jdssc/jdssc/jovian_common/exception.py`: not found per volume, snapshot, target and pool; already exists per volume, snapshot and dataset; busy; communication failure): give each kind its own code, for example volume not found, snapshot already exists, volume has clones, pool missing, out of space, invalid argument, authentication failed, and make callers decide on the code alone, never on message text.
- **A code for each kind of error, by layer** (user decision,
  [question 73](#decisions)). The `class` of an error shows which layer of
  the appliance raised it. Give each layer a range of 1000 codes, and each
  kind of error in a call's scenarios file a code inside the range of its
  layer:
  - 1000 to 1999: no `class`: login errors and some replies of 400 and 404;
  - 2000 to 2999: `builtins.*`: Python errors the appliance did not catch;
  - 3000 to 3999: `zfslib.*`: the ZFS wrapper;
  - 4000 to 4999: `opene.*`: the product's own logic;
  - 5000 to 5999: `jsonschema.exceptions.ValidationError`: a request body
    that fails its schema;
  - 6000 to 6999: `werkzeug.exceptions.*`: routing and the API's own
    checks.

  The multiple of 1000 is the code of the layer itself, and the kinds of
  error of a layer take the numbers after it. The codes below 1000 are the
  plugin's own, as today.
- **Two codes on every error** (user decision, [question 74](#decisions)).
  Keep `GetCode()` as it is: it returns the code that says what the error
  means for the call, and today's callers compare with it, unchanged. Add a
  second code beside it, which says which error the appliance sent: the code
  of its kind for an error that an entry of the call holds, the code of its
  layer alone for one that no entry holds, and 0 for a failure that is not
  an error of the appliance, such as a request without a reply. Each entry
  of the catalogue gives both codes. The kinds of error of the plugin's
  calls and their numbers are in Table 19 ([Measurements](#measurements);
  user decision, [question 75](#decisions)). A kind keeps its number, and
  one that a later reading of the spec adds takes the next free number of
  its layer. An entry marked "not in the spec" has no kind there: it gives
  the code of its layer alone until the spec holds it.
- **One handler for every failed reply**, in `pkg/rest`: make it match the
  reply against the catalogue entries of the call that failed, and turn anything that matches no entry into `RestErrorFailureUnknown`, with the status, `class`, `errno`, message and endpoint kept in the error, unless its status gives the meaning: a 404 (below), and a 401 or 403 without a JSON body (stage 3, above). Nothing else is guessed: an error that matches no entry is never turned into "busy" or "exists".
  A 404 is not a guess: a reply with that status means "not found" whatever
  its body (user decision, [question 40](#decisions)). The entry it matches
  says what is missing, the pool, the parent, the target or the object
  itself; give a 404 that matches no entry of its call, with a JSON body or
  without one, the general not-found code
  (`RestErrorResourceDNE` today), and keep its fields in the error
  ([risk 24](#risks)).
  A call that gets no reply at all, a transport error, goes through the same
  handler and gets its log line there ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)). The handler takes over the work of `ErrorFromErrorT` and of the matching in `getError`: remove `ErrorFromErrorT`, and with it `ErrorT.String()`, which only it calls. A reply without an error object then gets the code its status gives, which closes [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body) (user decision).
- **One log line per failed call, in one format**, written by that handler:
  method, path, status, `class`, `errno`, message and the two codes it was given.
  The handler cannot know whether a failure was expected (only the caller knows
  that "not found" on a repeated delete is fine), so have it log every error it
  recognises at Debug and every unknown error at Error. An error whose
  entry says Unknown on purpose is logged at Error too: the plugin cannot
  say what it means, and that has to be seen (user decision). So is a
  404 that matches no entry of its call: its status gives the meaning, but
  the catalogue does not hold the reply, and that has to be seen as well. So, for the same reason, are a reply without a JSON error body, a call that got no reply, and a reply the call cannot use. "One line" and "once"
  mean once in the REST module: no other code in `pkg/rest` logs the same
  failure again. What the driver and the controller log for a failed
  operation is their own record, and this document changes it only where it prints a whole CSI request ([impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs); [question 37](#decisions)). Let it replace today's
  scattered Warn lines ([bug_plugin_rest_every_error_logged_at_warn](#bug_plugin_rest_every_error_logged_at_warn)) and the empty message of
  unmatched errors ([bug_plugin_rest_unknown_error_empty_message](#bug_plugin_rest_unknown_error_empty_message)).
- **Handling a code is the caller's business** (user decision,
  [question 70](#decisions)). This document implements API v4 in the REST
  module and gives the modules that use it one code for each failure, with
  the appliance's message in the error. It does not decide which CSI status
  the driver and the controller return for a code. A code that no caller
  handles today, such as a missing pool or a failed authentication, gets
  the caller's default until the work on those modules decides otherwise
  ([Relationship](#relationship-to-other-designs)). This stage changes a
  caller only where an earlier decision says so. Let no caller map "any error" to one meaning: fix the zvol `DeleteClone` caller that does so today ([bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy](#bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy), `iscsi_driver.go:520`), and, for review [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children), keep `volume has children` on its own code (`RestErrorResourceBusyVolumeHasSnapshots` today) and
  make `DeleteVolume` run the intermediate-snapshot cleanup for it too. Do not delete a volume that still has snapshots after the cleanup, on iSCSI and NFS alike (user decision): answer `FailedPrecondition`, naming them, as CSI asks for a volume with snapshots, where today a volume with a CSI snapshot gets `Internal`. The NFS driver has to look first, since the appliance deletes a NAS volume together with its snapshots ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots)): make its `DeleteVolume` read the volume's snapshots, through the per-volume list this stage repairs, run the cleanup, and send the delete only when no snapshot is left. On iSCSI, where the appliance refuses, send the delete again when the cleanup leaves no snapshot. Count a snapshot that the plugin did not make as a snapshot in both drivers: it keeps its volume as a CSI snapshot does, and the error names it. Fix the NFS caller of `DeleteNASClone` as the zvol one is fixed: it returned every error as `Busy` too (`nfs_driver.go:636`).
- **Kept complete by the audit:** [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool) lists every recorded error of a plugin
  call that the catalogue does not cover for that call, so a new error in a hurl run shows up as
  a failed check.

What must hold: every error a call's list holds maps to the code its entry gives for that call, and that code is the one the same call returned for the reply before stage 3, but where this document decides another; every error `pkg/rest` returns carries two codes: the one `GetCode()` returns, which no caller has to read differently, and beside it the code of its kind or of its layer, inside the layer's range, or 0 for a failure that is not an error of the appliance; an `errno` that is a number, one that is `null` and one that is missing are three different inputs to the matching; a 404 that matches no entry of
its call gets the general not-found code and is logged at Error; an
unknown body keeps all its fields and is logged once by `pkg/rest`; a reply without an error body returns a `RestError` ([crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body)); no caller
maps every error of a call to one meaning, the zvol `DeleteClone` caller
included; and `DeleteVolume` runs the intermediate-snapshot cleanup for
`volume has children` too; and a volume that has a snapshot is not deleted, on iSCSI and NFS: `DeleteVolume` answers `FailedPrecondition`, and the volume and its snapshots stay on the appliance.

Follow the code under [Code segments](#code-segments) for the error body and
the handler of a failed reply.

### Stage 4: robustness and security
[Stage_4_robustness_and_security](#stage-4-robustness-and-security)

Fix the REST-layer [crashes](#crashes) ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)),
escape path segments ([impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments)), and add the security measures [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids), [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit)
([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) is the [first commit](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller), and the two logging items, [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs), are commits [before stage 1](#before-stage-1-keep-secrets-out-of-the-logs)). Most of it is
inside `pkg/rest`; [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) also touches the driver and the controller, and [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)
and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) add settings to the plugin configuration, after [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys) has tagged its structs and corrected its guides.

**Not possible today:** a REST user with fewer rights than admin. JovianDSS does
not offer such a role (confirmed by the user), so the plugin keeps using the admin account.

#### impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes

**Crash fixes**, all inside `pkg/rest` ([question 12](#decisions)):
- [crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error): a transport error is one where no HTTP reply arrives at all
  (the address does not resolve, the connection is refused, the request times
  out). Every one of the six branches that handle it reads `res.StatusCode`
  while `res` is `nil`, so any transport error panics today. The fix: return a
  `RestError` without touching `res`, and tell the kinds apart with `errors.As`
  (`*net.DNSError`, `*net.OpError` for a failed dial, `Timeout()` for any
  timeout, including the client's own). Remove the two `fmt.Println` calls in these
  branches (`rproxy.go:152`, `:156`), which write to stdout outside the logger;
  [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s handler logs the error. This is
  also the transport-error part of [bug_plugin_rest_reply_text_as_format_string](#bug_plugin_rest_reply_text_as_format_string):
  the error says which kind it was, where today's code, before it panics, builds `RequestMalfunction` for the client's own timeout and for any failure that is not a DNS error or a failed dial (`rproxy.go:148`, `:153`, `:157`).
- [crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check): check the `http.NewRequest` error before `SetBasicAuth`; and
  fix the URL scheme to `https` in `SetupRestProxy` ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): the plugin always uses
  HTTPS), and make it refuse a `prot` setting other than `https` with an error
  that names the setting, so that the URL can no longer start with `://`.
- [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body): nothing is left to fix here. Stage 3 removes `ErrorFromErrorT`, and its handler gives a reply without an error object the code its status gives ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue); user decision).
- [crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error): make `SetupEndpoint` log the error from `SetupRestProxy`
  **and** return it, instead of logging it and returning `nil`; and make `Send`
  return a `RestError` ("REST client not configured", followed by the setup
  error's text) when the proxy was never set up, instead of panicking.
  Make `SetupRestProxy` return an error for an empty `addrs`, which it accepts
  today. `SetupEndpoint`'s three
  callers (`controllerplugin.go:193`, `iscsi_driver.go:409`, `nfs_driver.go:546`)
  ignore the returned error today; with this fix a bad configuration would give
  an error on every REST call, naming the cause, instead of a panic. The log
  line stays because the callers ignore the error: without it, the cause of a
  bad `idletimeout`, an empty `addrs`, a `prot` other than `https` or a bad `cafile` ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile))
  would appear only once a REST call fails. Making startup fail early on that error is driver work,
  outside `pkg/rest` (for design 0003).
- [crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs): make `RestEndpoint.String()` (`rest.go:83-95`)
  return the string it builds and not read `Addrs[0]` when the list is
  empty. Go calls `String()` whenever the endpoint is formatted with `%v` or
  `%s`, so today a log line alone can panic.

What must hold, one condition per crash: a refused or timed-out connection returns a
`RestError` that names its kind ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)); a
configuration without `prot` works as HTTPS and one with `prot: http` is refused with
an error naming the setting ([crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check), [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)); an invalid `idletimeout` or an empty `addrs` makes every
REST call return the setup error ([crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error)); an endpoint with
no addresses formats with `%v` without a panic ([crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs)).

Follow the code under [Code segments](#code-segments) for the setup error
and the transport errors.

#### impl_plugin_rest_escape_path_segments

**Escape path segments** ([security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments); user decision): pass every
name or ID put into a REST path through `url.PathEscape`, in all path
builders in `pkg/rest`; one function, `apiPath`, builds the path of all 32
calls from a format and the names that go into it. A name then stays one path segment: `/`, `?`, `#`, `%` and
spaces are encoded and can no longer add segments, cut the path short or start a
query. `url.PathEscape` does not change `.` or `..`, and does not catch an empty
name, so it does not stop a name from addressing the parent or the collection;
that needs name validation, which is not part of [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments). The query string that
`pagedcSuffix` appends (`entrielisters.go:34-51`) is not a segment: escape the segments
first and add the query after, or `?` and `&` would be
encoded and paging would break. What must hold: a name containing `/`, `?`, `#`
or `%` reaches the appliance as one escaped path segment, and the query string of a list request reaches it unescaped.


#### impl_plugin_driver_validate_names_and_ids

**Validate names and IDs** (user decision): check a volume or snapshot ID
that comes in through a CSI request where it is parsed
(`NewVolumeDescFromCSIID`, `driver/lunid.go:147`; `NewSnapshotDescFromCSIID`,
`driver/snapid.go:118`; two entry points do not pass these parsers and are
not covered, [Relationship](#relationship-to-other-designs)), against the rules the appliance enforces
([question 20](#decisions)); target names are built by the plugin and
checked at setup (below):

**Table 6 — Name rules per object** · `tbl_name_rules`

| Name | Allowed characters | Longest | Why |
|---|---|---|---|
| zvol (volume, clone), NAS volume, NAS clone | `a-z A-Z 0-9 _ -` | **131** (user decision) | the appliance refuses anything else, saying so: "Allowed characters in zvol name: a-z, A-Z, 0-9, _, -", "cannot exceed 132 characters"; 131 keeps one character of margin |
| iSCSI target | **lower case only** (user decision): `^[a-z0-9]([a-z0-9.:-]*[a-z0-9])?$` | 220, the longest seen accepted | the documented pattern without upper case; the appliance also accepts upper case, but the plugin would not use it |
| snapshot | no empty name, no `.` or `..` (for the ID parsed from a CSI request; the name built at CreateSnapshot is left out, [question 25](#decisions)) | see [future work](#future-work-snapshot-name-limits) | ZFS's own rules, including a limit on the full name, are future work |

Reject a bad ID with `InvalidArgument` before any REST call; the check closes
none of the review's panics next to these parsers ([question 25](#decisions)). As a second line,
make the `pkg/rest` path builders refuse an empty, `.` or `..` segment, which
[impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments)'s escaping does not catch. Why IDs need checking: the plugin does not control
them. A volume ID comes back from Kubernetes in every later request, and anyone
allowed to create a static PersistentVolume or a `VolumeSnapshotContent` object
can put any string there.

**Keep the plugin's own names within the same limit.** The plugin builds the
appliance name of a volume from the CSI name in one of three forms: `vp_` plus
the CSI name as it is, when its characters are allowed; `vb_` plus its base32
form, when they are not; and `vs_` plus a SHA-256 hash (67 characters), when the
name is too long. `NewVolumeDescFromName` (`driver/lunid.go:65-81`) checks
two lengths against 240: the CSI name's, before it takes the `vp_` form
(`:65`), and its base32 form's, before it takes the `vb_` form (`:69`); so it
can build names of up to 243, which the appliance refuses. Lower **both**
thresholds to 125 (user decision, [question 71](#decisions)). The appliance
name of a volume is also its CSI ID (`driver/lunid.go:167-173`), and CSI
naming follows the CSI spec, which limits a string field to 128 bytes. With
the three characters of the prefix, every generated name is then at most 128
bytes, which is also within the 131 allowed for a name on the appliance
([question 21](#decisions)). A CSI name,
or a base32 form, longer than 125 would fall through to the `vs_` form.
Lowering only the first would leave `vb_` names of up to 243. A share takes
the name of its NAS volume (`nfs_driver.go:175`, `:279`), so the same limits
hold for it.
The appliance's message gives 132 as the limit (129 characters were recorded
accepted, 150 refused), so no volume with a longer name exists. One of 129
to 132 characters may: a volume that the current code made from a CSI name
of 126 to 129 characters (`vp_` form), or from one whose base32 form has 126
to 129 (`vb_`), would be looked up by its CSI name under its `vs_` name
after the change and not be found. Kubernetes volume names have 40
characters, so none is expected ([risk 13](#risks)).

**Target names are built, not received:** `<prefix>:<hash>`
(`iscsi_driver.go:569`). The hash is lower-case SHA-256 hex; the prefix is the
`iscsi.iqn` setting (default `iqn.csi.2019-04`, `controllerplugin.go:168-171`),
so the setting decides whether the name meets the target rule above.
Make `SetupControllerPlugin` (`controllerplugin.go:143`) check the prefix once
against that rule and return an error for upper case, other characters, or a
prefix so long that `prefix + 1 + 64` exceeds 220; today only the hash's length
is checked (`iscsi_driver.go:571`). What the plugin does with that error at
startup is not decided here ([question 41](#decisions)).

What must hold: an ID that breaks the rules of the table fails with
`InvalidArgument` at the two parsers, before any REST call; no generated
volume name, and so no CSI ID of a volume the plugin creates, is longer than
128 bytes; the `pkg/rest` path builders
refuse an empty, `.` or `..` segment; and `SetupControllerPlugin` returns an
error for an `iscsi.iqn` prefix outside the target rule.

#### impl_plugin_config_yaml_tags_and_guide_keys

**The configuration as the code reads it** (user decision; backlog items 19
and 20, which
[`review/2026-11-08-design-0002-review.md`](../../review/2026-11-08-design-0002-review.md)
asked to fix in the same edit of the guide as the TLS settings): its own
commit in stage 4, before [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), so that the configuration guide is right
when that item describes its settings in it. A user who sets up TLS from a
guide that puts `pool` in the wrong place still gets no pool.
- give every key of `RestEndpointCfg` and `ISCSIEndpointCfg`
  (`pkg/common/common.go:50-66`) a `yaml` tag with the name the guides use:
  both structs have only `json` tags today, and yaml.v3 finds their keys
  only through its fallback to the lowercased field name; a key added later
  in this stage comes with its tag. Tag `NFSEndpointCfg` the same way. The
  two keys of `ISCSIEndpointCfg` that nothing reads, `namelen` and
  `passlen`, take the names of their `json` tags, where the fallback read
  `vnamelen` and `vpasslen`;
- correct both configuration guides, `docs/configuration.md` and
  `docs/configuration-nfs.md`: `pool` belongs at the top level, not under
  `endpoint` (`pkg/common/common.go:76`); the key is `logfile`, not
  `logpath` (`:75`); the addresses after the first in `addrs` are not used
  ([bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover)), so mark them as unused; and describe `tries` and
  `retrydelay` as what stage 3 makes of them, the number of requests and the
  pause for a GET answered without data, and say that no other request is
  sent again: not a DELETE answered 204, which is a success, and not a
  request that failed ([question 39](#decisions)).

Correct the example `deploy/cfg/cfg.yaml` with the guides: it has `logpath`
and a `pool` under `endpoint` too. Show `logfile` commented out in the
examples: set, it moves the log from standard output into a file in the
container. A unit test in `pkg/common` reads the example of each guide and
the example file through `SetupConfig` and checks the pool and every key of
the endpoint, so a guide that drifts from the code fails `make test`.

What must hold: every key of `RestEndpointCfg` and `ISCSIEndpointCfg` has a
`yaml` tag with the name the guides give it; both guides put `pool` at the
top level and name the log file key `logfile`, and a configuration copied
from either guide gives the driver its pool.

#### impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile


**TLS and transport** (user decision; [security_plugin_rest_tls_verification_off](#security_plugin_rest_tls_verification_off)):
- verify certificates by default (today `InsecureSkipVerify: true`,
  `rproxy.go:251`);
- add an `insecure` setting to the REST endpoint configuration (`RestEndpointCfg`)
  that turns verification off;
- add a `cafile` setting that names a certificate to trust: the appliance's
  own certificate or the CA that signed it. JovianDSS lets the administrator
  import a certificate, so the appliance can present one that is within its
  dates and lists the address the plugin connects to; verification needs
  both, and `cafile` cannot make up for either (user decision: the
  administrator's certificate and `cafile`, no pinning by fingerprint). An
  appliance that still presents its own self-signed certificate (issuer
  "Open-E JovianDSS") would need `cafile` or `insecure` after upgrading
  ([risk 11](#risks)). The
  file has to reach the container: make the install guides and the controller
  manifests show how to add it to the config Secret
  (`jdss-controller-cfg`) and which path to put in `cafile`;
- say in `CHANGELOG.md` that certificate checks are on by default and that an appliance with a self-signed certificate needs `cafile` or `insecure` ([risk 11](#risks));
- set the minimum TLS version to **1.2** (`tls.Config.MinVersion`), fixed:
  the lab appliance and jdss1 offer no TLS 1.3 ([question 50](#decisions),
  [risk 10](#risks)); not a configuration key, though `pkg/rest` may take it
  as a parameter of its setup ([question 26](#decisions));
- do not follow redirects (`CheckRedirect` returning
  `http.ErrUseLastResponse`): the API has no reason to redirect, and a redirect
  means that something other than the API answered. Go already drops the
  `Authorization` header on a redirect to another host, but keeps it on the
  same host;
- always use HTTPS (user decision,
  [question 23](#decisions)): fix the URL scheme to
  `https`, and make a `prot` setting other than `https` a configuration
  error ([crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check)'s check in [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)); mark `prot` in the configuration guide as no
  longer needed;
- describe the new keys in the configuration guide (`docs/configuration.md`), which [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys) corrects in the commit before this one, and give each a `yaml` tag.

What must hold: with neither setting, a connection to an appliance whose
certificate does not verify fails with an error that names the certificate
problem and the two settings; with `cafile` naming the certificate or its CA,
for a certificate within its dates that lists the address the plugin
connects to, or with `insecure`, it succeeds; a server that offers less than TLS 1.2 is
refused; a redirect is not followed; `CHANGELOG.md` names both settings for a self-signed certificate; and a `prot` other than `https` is a
configuration error.

Follow the code under [Code segments](#code-segments) for the setup of the
client.

#### impl_plugin_rest_reply_size_limit

**Reply size limit** (user decision): cap how much of a reply `Send` reads
with a new setting in the REST endpoint configuration, `maxreplysize`
(`io.LimitReader` before `io.ReadAll`, `rproxy.go:164`). Let the reader
allow one byte more than the limit, and take a read that reaches that byte as a
reply over the limit: `io.LimitReader` alone cuts the reply short without an
error, and the caller would see broken JSON. Fail a reply over the limit
with an error naming the limit, instead of reading it into memory
whatever its size. Default: 1 MiB, 1,048,576 bytes ([question 54](#decisions)).
With `per_page=100` ([question 62](#decisions)) a page is at most about
200 KB: an entry of the volume list is about 2 KB (Table 16), and an entry
of each other list the plugin reads is under 1 KB in every recorded reply of
the hurl runs. A reply of the whole list, which `per_page=0` asks for
([How lists are paged](#how-lists-are-paged)), would pass the limit a little
over 500 volumes (Table 16); the plugin never asks for one. What must hold: a reply larger than `maxreplysize` fails with an error
naming the limit and is read no further than one byte past it, and a reply of exactly the limit is
read whole.

Follow the code under [Code segments](#code-segments) for the read of a
reply.

### Supporting work
[Supporting_work](#supporting-work)

#### impl_plugin_rest_recorded_reply_fixtures

**Fixtures, and what the unit tests must show.** Put replies taken from hurl's `out/`, trimmed, with
passwords and the `Authorization` header removed ([question 4](#decisions)), into
`pkg/rest/testdata/`, including replies to the per-volume snapshot list that
stage 3 adds ([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub); recorded by `list/volume-kinds.hurl`); no jdss2 reply is a
fixture ([question 19](#decisions)). The fixtures of a call are the smallest
set of its recorded replies that together shows every read field present,
every optional read field missing or `null` where a recording shows it so
(none shows `is_clone` missing, [question 65](#decisions)), and every alternative of the
reply, such as a zvol and a clone, and one recorded body for each error its
scenarios file lists; the agent picks them on the analysis
machine, where the recordings are. That every recorded reply fits the
spec is `jdss-rest-api-spec`'s own check, not repeated here. The tests on them are the implementing
agent's. What they must show: every fixture decodes with the transcribed types
([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), which is the check in Go of [rule_rest_reply_types_decode](#rule_rest_reply_types_decode); a required read
field removed from a fixture fails with the error naming it; every error
fixture of a call passed through the handler gets the code its entry gives
for that call (the entries marked "not in the spec" have no fixture and
are not tested here, T13); a request with a
password in its body leaves no password and no `Authorization` header in the
log, the condition of [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), whose commit
comes before these tests; and a fake HTTP server
(`net/http/httptest`) covers [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) and [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies); these tests rely on the fixes for
[crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error), which stage 4 makes ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)), and [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body), which stage 3 makes ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)). These would be
the first unit tests in the repository. Run them with
`go test ./pkg/rest/...` on the lab machine (`make test`; the target comes with the
[first commit](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller)), and keep them to the
standard library: no new module dependency, because the build targets run
`go mod tidy` and would pick any change up. Decoded with the current
hand-written types, **before** any fix, the same fixtures show the failures
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch),
[bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch),
[bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) and [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status) and, with a fake server,
[bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) and [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay): the fixtures are
evidence before they are tests.

The fixtures as they are written: 30 success replies under
`pkg/rest/testdata/replies/`, listed in `replies.json` with the hurl run,
file and case of each and what it shows, and 335 error replies under
`testdata/errors/`, one for each error entry of the scenarios files of the 32
calls, listed in `errors.json` with the two codes each must get and the code
its call returned before the catalogue. Two things in them are replaced: an
address of the lab in the message of one recorded error, by `192.0.2.1`, as
the spec masks it; and in the pool reply, trimmed to one disk of each vdev,
the identifiers of the pool and of the disks, by zeros. The tests pass each
error through the handler with the catalogue of its call, and serve every
fixture to the call itself from a fake appliance.

#### impl_plugin_rest_request_body_keys_checked_against_schema

**Request side (check only)**, against [rule_rest_request_body_valid](#rule_rest_request_body_valid). The keys the plugin puts into each request body
were compared with the company schema and with what hurl sent. The appliances accepted the plugin's bodies in the recordings (`force_umount: true` on zvol, snapshot and clone delete; `name, snapshot`; `snapshot_name`; target `active, name`; LUN `lun, mode, name`; the NAS and share forms), the `{}` body ([gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete)) and a volume create without `sparse` ([gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse)) included, once the gap runs sent them as the plugin does. Hurl Finding 14 shows what `force_umount: true` does: in
every recorded delete it gave the same result as `false` or no flag, a
refusal for a zvol with a snapshot, for the origin of a clone and for a
snapshot with a clone, and a 204 for a zvol attached as a LUN, all on idle
resources
([gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients)).
The flag itself changes in this delivery: stage 3 sends `false` (user
decision, [question 56](#decisions)). Put anything else found in the
backlog, not into this document. What must
hold: every key the plugin sets in a request body is in the company schema, every required key is sent and the body
was accepted in the recordings, and whatever the check finds is in the
backlog.

#### impl_plugin_rest_package_documentation

**Package documentation.** Add `pkg/rest/doc.go` to hold the package
comment, which Go shows with `go doc` and in editors. Make it state the rules of [priorities 4 and 5](#priorities), exact read fields and repeatable checks: the reply types are transcribed from `jdss-rest-api-spec`'s
OpenAPI of the replies recorded from real appliances ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), and the header of
`types.go` names the `jdss-rest-api-spec` commit they were read from; only the fields the
plugin reads carry a type, and the rest are raw JSON; one field with the
wrong JSON type makes the whole reply fail to decode, which is why read fields
are exact and the others are not; after new recordings the types are read
again, and `make test` and the audit ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) are re-run. What must hold: `go doc` on the package shows this rule, and it
agrees with this document.

## Consequences
[Consequences](#consequences)

- One API to implement, record, audit and test; every fixture would be a jdss1 v4
  reply.
- Installations on an older JovianDSS 1 release would stop working after the
  upgrade until the appliance is updated; the documentation and
  `CHANGELOG.md` would name the minimum release
  ([risk 2](#risks)).
- `types.go` would be rewritten from `jdss-rest-api-spec`'s observed OpenAPI, with a
  hand-written `types_methods.go` beside it; neither the build nor the
  transcription would need `jdss-rest-api-spec`'s tools or Node. Every field the
  appliance sends would be stored, and only the read ones would have a
  type; the rule for absent read fields would hold from the first transcription of `types.go`.
  Appliance states that no recording shows, such as a volume with encryption
  enabled, would keep their fields tolerant.
- The audit
  ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) would give the compliance matrix, the request keys, the error coverage and the comparison of the types with the OpenAPI; the fixtures ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) would be the check
  that the types decode; the future work on missing fields would shrink to the
  read fields the recordings never show set
  ([missing fields in replies](#future-work-missing-fields-in-replies)).
- `jdss-rest-api-spec` would become an input of the development process: a new
  recording would mean that the user runs its `--check` and commits the TypeSpec and the OpenAPI, an
  agent reads the changed files into `types.go` and the error catalogue, and
  the diff is reviewed. The audit would show whether the types still match the
  spec, and the fixtures that they decode what was recorded ([risk 17](#risks)).
- A CSI list that does not change while it is read would return each entry once ([risk 27](#risks)), and a walk would resume past an entry deleted between two calls; each list call but ListSnapshots by `snapshot_id` would read the whole appliance list, about
  one request per 100 entries ([risk 26](#risks)), and a token of today's
  form would get `Aborted`.
- The whole of this document would be delivered as one series, with no record of its own for stages 3 and 4: their files, risks and tests are in
  [Affected systems and files](#affected-systems-and-files),
  [risks 4 to 8, 10, 11, 13 to 15 and 23 to 28](#risks) and [Testing](#testing).

Confidence: high for JovianDSS 1 at the recorded release, where v3 and v4 answer the plugin's requests the same apart from the per-volume snapshot lists, which the plugin
reads only in their v4 shape ([Key observation 2](#key-observations)), and every recorded reply of every call is validated against the
spec the types come from; medium for read fields never seen set in the
recordings, whose type is stated by hand with a reason, of which there is
none today ([risk 16](#risks)); none
is claimed for older JovianDSS 1 releases, which are not supported
([question 29](#decisions)).

Re-evaluate when a new recording changes the type of a read field, when
the `jdss-rest-api-spec` project changes the layout of its OpenAPI, when the reading
and review of each new recording costs more than a generator would, or
when a generator has been shown to handle the optional fields: the
tracked OpenAPI makes generating the types possible, and the optional
fields are why it is set aside
([question 38](#decisions)). An installation on a release
older than the minimum is outside what this document supports and does not
reopen it ([question 29](#decisions)).

## Risks
[Risks](#risks)

**Backward compatibility, preserved.** An installation on the recorded
JovianDSS 1 release would keep working as before: the recordings show v3 and v4
answer the same for everything the plugin reads today
([Key observation 2](#key-observations)); the configuration keeps every key it has; the names and IDs of
existing volumes, snapshots and targets do not depend on the API version;
the seven type names and the accessors the drivers use, every call the plugin
makes and the request types stay (the dead-code commit removes only what
nothing calls, and `DeleteNASVolumeSnapshot`, whose one call it points at
`DeleteNASSnapshot`); Basic auth stays, and tokens would come only with
JovianDSS 2 support ([JovianDSS 2 support](#future-work-joviandss-2-support));
nothing on the appliance is touched. Reverting the plugin image restores v3
and the old types. What the calls do on success would stay the same, with
these exceptions:

- the `per_page` commit would change how lists are paged ([Testing](#testing)
  step 4);
- [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) and [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) would make snapshot lists return real entries
  ([risk 4](#risks));
- stage 3 would remove the 202 branch of `CreateTarget` ([risk 6](#risks));
- stage 3 would refuse the delete of a volume that has a snapshot, which on NFS deletes the snapshots today ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots));
- stage 3 would change what a list call returns ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)):
  each entry once, in the byte order of the names, with a `next_token` only
  when entries follow, each call reading the whole list ([risk 26](#risks)); a token of today's form would get `Aborted`, where today it crashes the controller;
- stage 4 could make a call that works today fail: [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) would reject IDs the plugin
  accepts now and refuse, at setup, an `iscsi.iqn` prefix outside the target
  rule; [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) would refuse an unverified certificate
  ([risk 11](#risks)) or a `prot` other than `https`; and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit) would
  refuse a reply over `maxreplysize`.

**Backward compatibility, not preserved.** These installations would stop
working after the upgrade until something changes: appliances on a JovianDSS 1
release without v4, or with a v4 that differs from the recorded one, until the
appliance is updated ([risk 2](#risks));
appliances whose certificate does not verify, such as their own self-signed one, until `cafile` or `insecure` is set
([risk 11](#risks)); configurations whose `iscsi.iqn` prefix breaks the target
rule, until the prefix is changed ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids));
and configurations with `prot: http`, which the plugin no longer accepts
([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)). In
the code, the thirteen lines that read `Creation` as a field, the list entry's
`Volume`, the pool's `Available` as a number or a zvol snapshot's `Name`
change in the same commit as the types, and the readers of the per-volume
snapshot lists in stage 3 ([risk 1](#risks), [question 43](#decisions)).

1. **A caller still reads a field whose type or key changes.** Guard: `go build`
   would fail on any remaining use; the thirteen lines are changed in
   the same commit, and the readers of the per-volume snapshot lists, which
   do not build in between, in stage 3 ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec),
   [question 43](#decisions)).
2. **Older releases.** All recordings come from current releases. **A JovianDSS 1
   release without v4, or with a different v4, would stop working, and there would
   be no fallback.** Guard: the docs and release notes would state the minimum
   release, `1.0up33.65410`, or the newer one the live tests ran on
   ([question 58](#decisions)). `1.0up33.65410` is the latest JovianDSS 1 release, and both the lab appliance and jdss1, the appliance of the hurl recordings, run it (confirmed by the user). Older JovianDSS 1 releases
   do not serve the same v4 and are not supported: there is no backward
   compatibility with them (user decision, [question 29](#decisions)). The v4
   move would be its own commit, so it could be reverted on its own.
3. **A transcription mistake, or the spec's layout moves.** Two
   risks, each described once with its guard: [risks 18 and 19](#risks).
4. **[bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) and [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) would change results, not only errors.** ListSnapshots
   would start returning real entries, which could expose paging bugs (the CSI list
   token) that are hidden today. Guard in this delivery: the fixtures of those list replies ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) and the implementer's tests on them, because
   csi-sanity cannot reach ListSnapshots until design 0003 stops the controller crashes
   ([Testing](#testing) limit 9); csi-sanity's ListSnapshots specs would follow after
   design 0003. The paging bugs known in those lists are fixed in the same stage
   ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)), among them
   [bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume](#bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume),
   which the type change alone would keep as it is. [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) would likewise change which path a repeated request takes, and [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised) the path of an NFS delete of a volume with clones.
5. **Recorded data contains passwords and lab details** (names, IP addresses,
   serial numbers; hurl's report of a run, `report.json`, also holds certificates and host names). Guard:
   fixtures would be trimmed and their passwords (CHAP and REST admin, including the
   Basic `Authorization` header) removed before commit; no jdss2 reply would be a
   fixture, so no token would reach the repository
   ([questions 4 and 19](#decisions)); lab details may stay.
6. **Removing [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay)'s 202 branch.** If an appliance ever answered
   `CreateTarget` with 202, the call would fail instead of polling. 202 was never
   recorded in any run ([how the gap runs went](#how-the-gap-runs-went)). Guard: a 202 on `CreateTarget` is an error naming the status ([Stage 3](#stage-3-behaviour-and-error-mapping)), so a real 202 would be
   noticed.
7. **One live test per driver for the whole delivery.** A regression could not be
   traced by comparing a live cycle per commit. Guard: every check that does not
   need the lab would pass first ([Testing](#testing) steps 1 and 2); each commit would
   build on its own but those [question 43](#decisions) names; and `git bisect`,
   skipping those, would find the commit if a
   live test failed among the builds from the second logging commit on, which all keep secrets out of the log ([Before stage 1](#before-stage-1-keep-secrets-out-of-the-logs)); the commits before it are traced by reading them (kub-testing `testcases/plugin-checks-design-0002.md`). The commit
   found would be fixed before the delivery is finished, not as a follow-up.
8. **Connected clients are untested ([gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients)).** No recorded
   reply had an initiator logged in or a volume in use, and the iSCSI node code may
   not log in at all today, nor log out at unstage (review `bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout`, `ISSUES.md` T10). The calls that act on a
   target or volume a client may still use have been seen only on idle resources,
   and never on v4 on the lab: detach, target delete, and the iSCSI deletes, which
   send `force_umount: true` today (`iscsi_driver.go:102`, `:214`, `:240`, `:487`, `:516`)
   and `false` from stage 3 ([question 56](#decisions)).
   The v4 move and [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)'s change, which would go on to delete the
   target when the LUN is not attached, act exactly on these calls. If the appliance
   refused them while a session is active, unpublish or delete would fail and leave
   targets behind. If it accepted them, a delete could remove a volume that a
   node still uses, for example after a node crash: a data-safety question. Guard:
   [Testing](#testing) steps 6 and 7. Until they have run, the behaviour with a
   client connected is unknown, on v3 as on v4.
9. **A read field absent from a reply.** Go's `encoding/json` fails only on a
   wrong type or broken JSON: a missing key or `null` leaves the field at its zero value without
   an error. Cases already seen: [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead) (no `volume`
   key, so every snapshot entry is skipped and ListSnapshots is empty),
   [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) (zvol snapshot without `name`) and [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) (NAS `quota` is `null` without a quota, so the size is 0;
   review `bug_plugin_driver_nas_quota_null_size_zero`). A value that does not parse ends the same way today ([bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors): a pool `available` becomes 0; [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors) would fix this one). Unguarded, a later release that dropped or
   renamed a read field could make GetCapacity report no free space, a volume report a size of
   0, lists come back empty without an error, an empty name go into the next
   request's path (see also [security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments)), or a snapshot `creation` of
   year 1 reach Kubernetes. Guard in this delivery: the
   required-or-optional rule and `checkRequired()` ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), a required
   field removed from its fixture failing by name ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)), and raw JSON for everything the plugin does not
   read, which cannot fail. What stays: a read field the recordings never show
   set is typed by hand until a recording confirms it
   ([missing fields in replies](#future-work-missing-fields-in-replies)).
10. **TLS 1.2 or higher ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)).** The lab appliance and jdss1
    offer no TLS 1.3 (checked by the user, [question 50](#decisions)), so the
    minimum would be TLS 1.2, the floor the current code already gets from
    Go's default (`rproxy.go:249` sets no minimum): the TLS version would
    refuse no appliance the plugin reaches today. An appliance that offered
    only an older version could not connect, as today. Guard: the minimum is
    set in `tls.Config`, so a change of Go's default would not move it. There
    is no `tlsmin` setting (user decision, [question 26](#decisions)): another
    minimum would need a code change.
11. **Certificate checks on by default ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile)) would stop some installations at the upgrade.** An appliance whose certificate does not verify, such as its own self-signed one, would make an existing installation stop connecting after the upgrade until the installation sets `cafile` or `insecure`. Guard: the
    release notes say so, with both settings, and the error names the certificate
    problem and the two settings. The lab's live test needs one of them set.
12. **No request deadline ([bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied)).** The request is built
    without the CSI call's context, so a hung appliance call is not cancelled when
    the CSI deadline passes; it lasts until `idletimeout`, which is used as the
    timeout of the whole request. Guard: none in this delivery; future work
    ([request deadlines](#future-work-request-deadlines)).
13. **Volume names of 129 to 132 characters ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)).** A volume with such a name, which the current code made
    from a CSI name of 126 to 129 characters or from one whose base32 form has 126 to 129, would be looked up by its CSI name under its `vs_`
    name after the change and not be found: a CreateVolume repeated for it
    would make a second volume. Its ID, the name itself, still reaches the
    volume, but for an ID of 132 characters, which would be refused with
    `InvalidArgument` where the ID check runs.
    Kubernetes volume names have 40
    characters, so none is expected. Guard: none; the edge is accepted with the
    limits of 131 for a name on the appliance and of 128 bytes for a
    generated name (user decisions, [questions 21 and 71](#decisions)), which
    may change later.
14. **Replies may be behind the request.** Two cases are recorded. On jdss2, a
    snapshot list read within half a second of a snapshot create or delete still
    shows the old state, while a read by name is current at once (hurl
    Finding 20; in that run jdss1 showed the lag twice in eight reads, on v3, and in three of four earlier runs its v4 NAS snapshot lists missed a new snapshot up to 0.55 s after the create, as `jdss-rest-api-spec/empirical/hurl/project-status.md` records). And
    a new clone's `san:volume_id` is empty in the first read of the clone, by name
    or in a list, however long after the create; the next read has it (hurl
    Findings 14 and 15). The plugin decodes that field (`types.go:100`, `:211`)
    and nothing reads it today. Other fields and other lists may behave the same
    way: the visibility tests of hurl Finding 20 saw no further case, but they only cover what they
    read.
    This document changes what the client sends and how it decodes and maps
    what comes back, and the driver steps that [Stage 3](#stage-3-behaviour-and-error-mapping)
    lists; none of them changes how the driver takes a reply as the current
    state, the clones checks of the cleanup and of the NFS
    `deleteIntermediateSnapshot` included ([risk 23](#risks)).
    Today that logic takes a reply as the current state: `cleanIntermediateSnapshots`
    deletes the intermediate snapshots one list shows and keeps the rest
    (`iscsi_driver.go:185`, `nfs_driver.go:291`), and ListSnapshots reports what
    one list read returns. A read moments after a change could therefore miss an
    object that exists or show one that is gone, and a step taken on it would be
    a false operation: an intermediate snapshot left behind, a snapshot reported
    to Kubernetes as missing, a field read as empty that the next read would
    fill. The exposure is small on JovianDSS 1; JovianDSS 2 shows the lag most ([JovianDSS 2 support](#future-work-joviandss-2-support)). Guard: none in this
    delivery. Every later piece of driver work, whether a design, an audit or a
    fix, would have to treat a reply as possibly behind the request: read by name
    where the name is known, retry a list that lacks an expected entry before
    concluding that it is absent, and take no destructive step on the strength
    of a single list read straight after a change.

15. **Some situations have never been seen on jdss1 v4.** The test of
   simultaneous requests ran on jdss1 on v3 only, the full-pool tests on v4
   cover the volume creates but not the snapshot and clone creates, and of
   the lists of hundreds of objects only the volume list has been read
   ([How lists are paged](#how-lists-are-paged); the `scale/` tests ran on no
   appliance, [how the gap runs went](#how-the-gap-runs-went)).
   Guard: v3 and v4 answered the same wherever both were recorded, apart from
   the per-volume snapshot lists, which the plugin reads only in their v4
   shape ([API versions](#api-versions)); the live
   tests ([Testing](#testing) steps 4 and 5) run on v4 every call that
   csi-sanity reaches (limit 9 lists the findings it does not reach), and the
   path move being its own commit, `git bisect` can tell its effect from a
   fix's.
16. **A read field was never seen set, so the recordings give it no type.**
   Guard: the type is stated by hand in the source, with the reason in a
   comment; a live test in which the field is set shows whether that type decodes. At `jdss-rest-api-spec` commit
   `6149fa9` no read field is in this state: each has been seen set. The
   risk is for a field the plugin starts to read, or a later recording.
17. **The transcribed types drift from the spec**, because the spec gains
   recordings that nobody reads in, or because someone edits a type. Guard:
   the header of `types.go` names the `jdss-rest-api-spec` commit and the file behind each
   type, so a later reading compares against a known state, the audit compares
   `types.go` with the OpenAPI field by field ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)), and the unit tests
   decode the fixtures. A generator with a check target is set aside (user
   decision, [question 38](#decisions)); the audit's comparison stands in for
   it, and has to be run, on the analysis machine, for the guard to work.
18. **A transcription mistake**, for example a missed `null` alternative or a
   type read from the wrong endpoint. Guard: every fixture decodes with the
   transcribed types, the fixtures of a call showing each read field present
   and each optional one absent where a recording shows it so, and a required read field removed from
   its fixture fails by name
   ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).
19. **The spec's layout moves** between commits of `jdss-rest-api-spec`, as when it moved to one error model per kind. Guard: only files tracked in `jdss-rest-api-spec` are read, at a commit
   the header names ([question 31](#decisions)); a layout change shows in
   that repository's history and changes how the next reading is done, not
   the types.
20. **The procedure depends on one person.** Only the user may run `jdss-rest-api-spec`'s tools, and only a machine with the recordings can. Guard:
   the OpenAPI is committed in `jdss-rest-api-spec` and the types in this
   one, so a reading needs no tool. The written procedure is not part of this
   delivery ([question 33](#decisions)).
21. **Per-endpoint types diverge for one accessor**, such as `creation` as
   `int64` in a list and a string by name. Guard: conversions live in the
   methods file, per type, as the decoders do today, and the fixtures cover
   both endpoints.
22. **A later JovianDSS release changes v4.** Guard: a new reading and the
   audit on the new recordings show the change as a diff or a failed check.
23. **Stage 3 would make an NFS data-loss path reachable
   (`dataloss_plugin_driver_nfs_intermediate_snapshot_delete_destroys_clone_volume`).**
   The appliance does not refuse the delete of a NAS snapshot that has a
   clone: it answers 204 and destroys the clone with it (hurl Finding 7); a
   zvol snapshot with a clone is refused. The NFS `cleanIntermediateSnapshots`
   deletes an intermediate snapshot that its list shows without clones
   (`nfs_driver.go:315-322`), and that list never shows clones
   ([bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name)).
   An NFS volume delete would therefore destroy every volume cloned from
   the deleted one. The cleanup deletes nothing today, because four bugs
   stand in its way, and stage 3 and the dead-code commit would remove all
   four: [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised),
   [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots),
   [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) and
   [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete).
   Stage 3 would open a second way to the same loss. `CreateVolumeFromVolume`
   calls `deleteIntermediateSnapshot` when the create of the intermediate
   snapshot is answered "already exists" (`nfs_driver.go:254-255`), which
   the NFS driver would recognise for the first time
   ([bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised)),
   and that function deletes the snapshot without looking at its clones
   (`:193`). The controller checks first whether the volume exists
   (`controllerplugin.go:506`), so this way needs a volume left half-created,
   or two requests for the same volume at the same time, which nothing
   serialises (`lockVolume`, `controllerplugin.go:200`, has no caller).
   Guard: the cleanup and the NFS `deleteIntermediateSnapshot` ask the
   appliance for a snapshot's clones before they delete it
   ([impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow),
   [Stage 3](#stage-3-behaviour-and-error-mapping),
   [question 34](#decisions)), and [Testing](#testing) step 5 sends DeleteVolume for such a
   volume and checks that its clone is still there; the second way has no
   test in this delivery ([Testing](#testing) limit 12). The clones reply is a
   read like any other: a clone made moments before could be missing from it
   ([risk 14](#risks)). The same loss through
   DeleteSnapshot does not depend on this delivery and is handled outside it
   ([Relationship](#relationship-to-other-designs)).
24. **A 404 is "not found" whatever its body ([question 40](#decisions)).**
   Not every 404 has to mean that an object is missing. The plain-text 404s
   in the recordings answer a name with an escaped slash, a path the
   appliance does not route; a path it does not serve at all gets a JSON 404 of class `werkzeug.exceptions.NotFound`, which no entry of a plugin call holds (the scenarios file of `GET +unknown+`); a 404 could also carry a message no recording
   shows. The call would take any of them as "does not exist", so a delete
   would report success while the object stays: on an appliance that does
   not serve this v4 ([risk 2](#risks)), possibly every delete. In the
   recordings a 404 always names something that does not exist: every one that
   the scenarios files of the plugin's calls hold on jdss1 v4, at
   `jdss-rest-api-spec` commit `94d0c4f`, is a missing
   pool, parent, target, attachment or clone, or a name no object can
   have. Guard: IDs and names are checked before any REST call
   ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)), except at the two entry
   points the [Relationship](#relationship-to-other-designs) names; a 404
   that matches no entry of its call is logged at Error with its fields
   ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)); and the minimum release is stated
   ([risk 2](#risks)). What these do not cover has no guard in this
   delivery; the 404 of a path the appliance does not route is left to
   another record (kub-testing `ISSUES.md`, F31).
25. **A missing pool would be reported as a missing parent on the snapshot
   and clone creates
   ([bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap)).**
   On the zvol snapshot create the two are one kind of error, told apart only
   by the names in the message; on the zvol and NAS clone creates the missing
   snapshot's message also matches the missing pool's pattern. An error that
   can be either would be taken as the parent missing ([question 48](#decisions)):
   the CSI caller would be told that the source is gone when the pool is, and
   the code of a missing pool
   ([Stage 3](#stage-3-behaviour-and-error-mapping)) would not be given on
   these calls. Guard: a `TODO` comment at these entries says that the missing
   parent cannot be told from a missing pool; the handler's log line at Debug
   keeps the message, which names what is missing
   ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)).
26. **Every CSI list call reads the whole list
   ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)).** Each request makes the
   appliance build the whole list, so it costs the same whatever
   `per_page`: at 501 volumes a ListVolumes is six requests of 1.3 to 2.1 s,
   8 to 13 s in all, against about 0.3 s for one request at 100 volumes
   ([Measurements](#measurements)). The number of requests and the time of
   each both grow with the list, so the time of a call grows faster than
   the list does (inferred from the two lengths measured). A caller whose
   deadline is shorter gets no list, and the plugin reads on, since its
   requests do not carry the CSI deadline ([risk 12](#risks)). Guard: none
   in the plugin; the page size, which sets how many requests a read takes,
   stays at 100 ([question 62](#decisions)). No sidecar of the shipped controller
   manifests is known to list volumes or all snapshots: the attacher lists
   volumes only for a plugin that offers `LIST_VOLUMES_PUBLISHED_NODES`,
   which this one does not (`controllerplugin.go:63-70`), and the
   snapshotter asks for one snapshot by its ID, which the plugin reads
   alone. So the cost falls on csi-sanity and on tools that list.
27. **A list that changes while it is read
   ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)).** When an entry before
   the read's position is deleted between two requests of one read, the
   next page starts one entry later and the read misses one entry; when
   one is created there, the next page would start with the last entry
   already read, which then comes twice
   ([How lists are paged](#how-lists-are-paged)). The list is read once
   ([question 64](#decisions)), so an entry missed is missing from that
   call's reply, and a sequence of calls through tokens passes it by when
   its key falls between the token's key and the last key of that reply;
   a later list sees it. Guard: of two entries of one name that are the
   same in `default_scsi_id`, `guid` and `is_clone`, as far as the list
   has them, the one read first is dropped; two that differ both stay
   ([question 65](#decisions)). CSI allows this: a
   caller that pages while volumes or snapshots are created or deleted
   may see some twice, miss some, or both (CSI spec, ListVolumes and
   ListSnapshots). A list of fewer than 100 entries is read in one
   request and is not affected.
28. **A list that serves fewer than 100 entries to a page ([Before stage 1](#before-stage-1-send-per_page-on-every-list)).**
    The stop rule ends a list on a page shorter than `per_page`. The volume
    list of jdss1 v4 serves any `per_page` in full (hurl Finding 24); the
    five other lists, and v3, are taken to do the same without a run of their
    own ([question 49](#decisions)). If one of them served fewer than 100
    entries to a page, a longer list of it would end after its first page,
    and the plugin would not see the entries after it. Guard: none in this delivery ([question 49](#decisions)); a paging run
    of each of the other lists, like the one of hurl Finding 24, would show
    such a cap; `jdss-rest-api-spec` holds those tests for the five other lists (commit `434c9cd`). A list shorter than the cap is not affected.
29. **Two keys that share a long beginning could make a list call return
    entries again.** Where the last entry a call returns and the next one
    share a beginning too long for the 128 bytes of the token, the token
    holds that beginning and a hash of each key
    ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
    [question 68](#decisions)). If both entries are deleted before the next
    call, that call starts after the beginning, and the entries that share
    it and were returned already come again; none is missed. Keys made from
    the names Kubernetes gives, `pvc-` or `snapshot-` and a UUID, are far
    shorter; the snapshots of a volume with a name of over about 100
    characters are the case. Guard: none.
30. **`go vet` passes from its cache.** With Go 1.26.0 the go command takes
    a vet result from its build cache. A package that was vetted before as a
    dependency of another one gets no report when it is vetted itself: on
    the lab machine, a plain `go vet ./pkg/rest/...` reported nothing on the
    current code, which has eight reports there, after `pkg/driver`, which
    imports `pkg/rest`, had been vetted. A commit that vet should stop would
    then pass [Testing](#testing) step 1. Guard: `make vet` runs
    `go vet -a`, which analyses the packages again and reported all eight
    ([question 76](#decisions)); step 1 is made with `make vet`. The vet
    that `go test` runs by itself is not forced, so it can miss a report
    that `make vet` shows.

## Affected systems and files
[Affected_systems_and_files](#affected-systems-and-files)

**Table 7 — Affected systems** · `tbl_systems_affected`

| System | Effect of this document | Where |
|---|---|---|
| Plugin REST client, paths | one API prefix, `api/v4`, replaces the 34 hard-coded `api/v3` strings (30 after the dead-code commit); comments that name v3 go | `pkg/rest/*.go` (Table 8) |
| Plugin REST client, reply types | `types.go` rewritten from `jdss-rest-api-spec`'s observed OpenAPI; the accessors, conversions and required-field checks in a methods file | `pkg/rest/types.go` (transcribed), `pkg/rest/types_methods.go` |
| Fixtures, tests and audit | fixtures chosen from the recorded jdss1 v4 replies, which the implementing agent's tests decode with the transcribed types and check required field by required field; the audit compares the plugin's calls with the v4 index and the transcribed types with the OpenAPI | `pkg/rest/*_test.go`, `pkg/rest/testdata/`; `kub-testing/tools/rest-type-audit.py` ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures), [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)) |
| Drivers and controller | the thirteen lines: `Creation` via accessor, `Parent.Name` for list entries, `GetAvailable()` for the pool, the requested name in the iSCSI `GetSnapshot`; the callers of the per-volume snapshot lists ([impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow)); the paging of the CSI lists ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)) | `pkg/controller/listers.go:52`, `:54`, `:56`, `:82`, `:124`, `:148`; `pkg/controller/controllerplugin.go:679`, `:802`, `:1026`; `pkg/driver/restwrappers.go:40`; `pkg/driver/iscsi_driver.go:454`, `:456`; `pkg/driver/nfs_driver.go:591`; the other driver and controller changes are rows of Table 8 |
| Plugin documentation and release notes | the required JovianDSS release is stated: one that serves API v4, for JovianDSS 1 at least `1.0up33.65410`, or the newer release the live tests ran on ([question 58](#decisions)) | `docs/`, `CHANGELOG.md` |
| JovianDSS appliances | must serve API v4: JovianDSS 1 at `1.0up33.65410`, or the newer release the live tests ran on ([question 58](#decisions)), or later | the lab appliance and the recorded appliances already do |
| Test lab | the live cycles become the plugin's first runs on v4; the suite itself does not change | kub-testing `dev_cycle.py` cycles ([Testing](#testing) steps 4 and 5); `ISSUES.md` T9 |
| `jdss-rest-api-spec` | read-only input: the tracked OpenAPI for jdss1 v4, with its error schemas, and the scenarios files beside the TypeSpec; for the audit and the fixtures also the summaries of jdss1 v4 (`INDEX.md`, `ERROR-FIELDS.md`, `STATUS-CODES.md`), jdss2 v4, the company specs and the hurl runs | `jdss-rest-api-spec/empirical/jdss1/v4/openapi3/`, `.../typespec/`; `empirical/` for the rest |
| Not affected | the node DaemonSet manifest; the node's mount and iSCSI code, apart from the lines of `pkg/node/publish.go` that print whole CSI requests ([impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)) | |

The changes, one row each, with the files they touch and the stage that makes them:

Entries marked *NFS* are NFS-only code: the NFS cycle ([Testing](#testing) step 5) reaches it as far as csi-sanity does (limits 9 and 12), and unit tests cover it where it is `pkg/rest` code.

**Table 8 — Files that would change** · `tbl_files_changed`

| File | Change | Stage |
|---|---|---|
| `pkg/rest/entrielisters.go` (`pagedcSuffix`), `pkg/driver/common.go` (`getResourcesList` stop rule) | [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) `per_page` | before 1 |
| `pkg/rest/storageapi.go`, `rest.go`, `rproxy.go`, `payload-def.go`, `reqtypes.go`, `nas_volumes_snapshots.go`, `shares.go`, `targets.go`; `pkg/driver/nfs_driver.go:322`; `pkg/controller/config.go` (removed), `controllerplugin.go:93` | [impl_plugin_rest_remove_dead_code_before_v4_move](#impl_plugin_rest_remove_dead_code_before_v4_move), [dead code](#dead-code); [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) and [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check) (switch to `DeleteNASSnapshot`) | before 1 |
| every `pkg/rest/*.go` with a path | one API prefix, `api/v4` | 1 |
| plugin docs, `CHANGELOG.md` | v3 abandoned in favour of v4 (still served, no longer called), minimum release `1.0up33.65410` or the newer one the live tests ran on ([question 58](#decisions)) | 1 |
| `pkg/rest/types.go` (rewritten), `types_methods.go` (new) | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec): transcribed reply types; closes [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) *NFS* and [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead); [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) *NFS* key (stage 3 filters by it); [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors) | 2, 3 |
| `pkg/controller/listers.go`, `controllerplugin.go`; `pkg/driver/restwrappers.go`, `iscsi_driver.go`, `nfs_driver.go` | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec): the thirteen lines that read `Creation`, the list entry's `Volume`, the pool's `Available`, a zvol snapshot's `Name` ([question 42](#decisions)) | 2 |
| `pkg/rest/*_test.go`, `pkg/rest/testdata/` (new) | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) | 2, 3, 4 |
| `Makefile` | `vet` (`go vet -a` on `pkg/rest`, `pkg/driver` and `pkg/controller`, [question 76](#decisions)) and `test` targets | first commit |
| `kub-testing/tools/rest-type-audit.py` (new) | [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool): compliance matrix, request keys, error coverage, the types against the OpenAPI | 2 |
| `pkg/rest/doc.go` (new) | [impl_plugin_rest_package_documentation](#impl_plugin_rest_package_documentation), package documentation | 2 |
| `pkg/rest/entrielisters.go` | [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub) (a new call, `GET pools/<x>/volumes/<x>/snapshots`, in the [compliance matrix](#compliance-matrix)), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) *NFS* | 3 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go` (`ListVolumeSnapshots`, `cleanIntermediateSnapshots`), `pkg/driver/restwrappers.go`, `pkg/controller/listers.go` (`completeListResponseFromVolumeSnapshots`) | [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow): the short entry type, `Properties.Creation`; the cleanup's clones by [question 34](#decisions) | 3 |
| `pkg/rest/errors.go`, `targets.go` | [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) and [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) *NFS*, [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised), [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap) ([question 48](#decisions)) | 3 |
| `pkg/rest/errors.go`, `storageapi.go` (`getError`), `payload-def.go` (`ErrorT`); `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/*.go` (callers) | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue): error catalogue, one handler and log line, codes per kind; callers decide on the code only ([bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy](#bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy), `iscsi_driver.go:520`, included); no volume that has a snapshot deleted, the NFS one looked at first ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots)); `ErrorFromErrorT` and `ErrorT.String()` removed, which closes [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body) | 3 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go` (`DeleteVolume`, `cleanIntermediateSnapshots`) | review [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children) (cleanup also on `volume has children`, with [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)), [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error) (listing error first; NFS takes the NAS short entries, *NFS*) | 3 |
| `pkg/rest/storageapi.go` (`getError`), `clones.go`, `pools.go`, `targets.go`, `nas_volumes.go`, `nas_snapshots.go`, `nas_volumes_snapshots.go`, `nas_volumes_snapshots_clones.go`, `shares.go`, `entrielisters.go`; `rproxy.go` and `pkg/common/common.go` (`RestEndpointCfg`: `tries` read, `retrydelay` new) | [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status) and [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check): status checked before decoding, non-JSON bodies mapped by status, a GET answered 204 sent again ([question 39](#decisions)) | 3 |
| `pkg/driver/nfs_driver.go` (`deleteIntermediateSnapshot`) | the clones read before an intermediate snapshot is deleted ([question 34](#decisions), [risk 23](#risks)) *NFS* | 3 |
| `pkg/driver/iscsi_driver.go:102`, `:214`, `:240`, `:487`, `:516` | `force_umount: false` on the iSCSI deletes ([question 56](#decisions)) | 3 |
| `pkg/rest/shares.go` (`GetShare`, new), `types.go` (`ResourceShare`); `pkg/driver/nfs_driver.go:140`, `:181`, `:285` | the existing-share check of [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) ([question 57](#decisions)) *NFS* | 3 |
| `pkg/driver/common.go` (`getResourcesList`), `snaptoken.go`, `restwrappers.go`, the list functions of `iscsi_driver.go` and `nfs_driver.go`; `pkg/controller/controllerplugin.go` (ListVolumes, ListSnapshots), `listers.go` (the name check moves to the driver); `pkg/rest/entrielisters.go` (`pagedcSuffix`, no `_dc`); `Makefile` (`test` also runs `pkg/driver` and `pkg/controller`); `pkg/driver/*_test.go`, `pkg/controller/*_test.go` (new) | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name): the whole list read once on every call, key order, a token of at most 128 bytes that resumes after the last name, `Aborted` for a token the plugin did not issue; [crash_plugin_driver_list_token_parse_writes_through_nil_pointer](#crash_plugin_driver_list_token_parse_writes_through_nil_pointer), [bug_plugin_driver_list_resume_repeats_last_entry_returned](#bug_plugin_driver_list_resume_repeats_last_entry_returned), [bug_plugin_driver_list_resume_compares_base64_forms_not_names](#bug_plugin_driver_list_resume_compares_base64_forms_not_names), [bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages](#bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages), [bug_plugin_driver_list_resume_drops_page_missing_token_entry](#bug_plugin_driver_list_resume_drops_page_missing_token_entry), [bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume](#bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume), [bug_plugin_ctrl_list_volumes_negative_max_entries_answered_internal](#bug_plugin_ctrl_list_volumes_negative_max_entries_answered_internal) | 3 |
| `pkg/rest/rproxy.go`, `errors.go`, `rest.go` | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes): [crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error), [crash_plugin_rest_basic_auth_before_request_check](#crash_plugin_rest_basic_auth_before_request_check), [crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error), [crash_plugin_rest_endpoint_string_empty_addrs](#crash_plugin_rest_endpoint_string_empty_addrs) | 4 |
| `pkg/rest/rproxy.go` | [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging): no request body or header logged ([security_plugin_secrets_in_logs](#security_plugin_secrets_in_logs), [security_plugin_rest_request_dump_every_call](#security_plugin_rest_request_dump_every_call)) | before 1 |
| every `pkg/rest/*.go` with a path builder | [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments): `url.PathEscape` on every path segment ([security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments)) | 4 |
| `pkg/driver/lunid.go` (`NewVolumeDescFromName`: both thresholds, of the CSI name and of its base32 form, 240 → 125), `snapid.go`; `pkg/controller/controllerplugin.go` (`iscsi.iqn` prefix check); `pkg/rest` path builders | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids): ID and name validation, generated volume names at most 128 bytes ([question 71](#decisions)), target prefix checked | 4 |
| `pkg/common/common.go` (`RestEndpointCfg`, `ISCSIEndpointCfg`: `yaml` tags), `docs/configuration.md`, `docs/configuration-nfs.md` (`pool` and `logfile` corrected; `tries` and `retrydelay` described; extra `addrs` marked unused) | [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys): the configuration as the code reads it | 4 |
| `pkg/rest/rproxy.go`, `pkg/common/common.go` (`RestEndpointCfg`: the new keys), `docs/configuration.md` (new keys; `prot` marked no longer needed), `docs/install.md`, `docs/install-nfs.md`, the controller manifests (the CA file in the config Secret) | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): HTTPS only, certificate checks, `insecure`, `cafile`, TLS 1.2 or higher, no redirects; [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit): `maxreplysize` | 4 |
| `pkg/rest/storageapi.go:51`, `:53`, `errors.go:173`, `:265`, `targets.go:45`, `:76`, `:150`, `:184` | [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply): `"%s"` formats and direct arguments, so `go vet` passes for `pkg/rest` | first commit |
| `pkg/controller/controllerplugin.go` (the 23 lines the item lists) | [impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text): the text of an error passed as the message, not as the format, so `go vet` passes for `pkg/controller` and `go test` runs its tests ([question 69](#decisions)) | first commit |
| `pkg/controller/controllerplugin.go`, `pkg/node/publish.go` | [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs): no whole CSI requests in logs | before 1 |
| `CHANGELOG.md` | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile): certificate checks on by default, `cafile`/`insecure` needed for self-signed certificates | 4 |
| `pkg/driver/iscsi_driver.go`, `nfs_driver.go`, `pkg/controller/listers.go` | callers of [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) *NFS*, [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) *NFS*; the filter by kind ([bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key), [question 7](#decisions)) | 2, 3 |
| `review/backlog.md` | item 36 addressed; item 37 closed (zero-padded format checked against recorded snapshots); items 19 and 20 addressed by [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys); [impl_plugin_rest_request_body_keys_checked_against_schema](#impl_plugin_rest_request_body_keys_checked_against_schema) findings added | all |
| `docs/adr/0002-api-v4-compliance-jdss1.md` | this document: status and the decisions as they are taken | all |

## Testing
[Testing](#testing)

**The delivery is tested once, as one series: the steps that need no
appliance come first and have to pass, then the live steps on the lab
appliance, one cycle for each driver, iSCSI and NFS ([risk 7](#risks)).**
Every appliance a test uses runs the latest release of its JovianDSS line, and one that
does not is updated before the tests ([question 11](#decisions)). If that
makes the lab run a newer JovianDSS 1 release than `1.0up33.65410`, the
docs and the changelog state that release as the minimum
([question 58](#decisions)). Unit tests are not listed here: the items say
what theirs must show. Who runs each test, what it needs first and its
commands are kept with the tests, in kub-testing
(`testcases/plugin-checks-design-0002.md`).

**Before the live test**

1. **Each commit of the series on its own** ([risk 7](#risks)). It builds,
   and `make vet`, which runs `go vet -a` ([risk 30](#risks)), reports
   nothing for `pkg/rest`, `pkg/driver` and `pkg/controller`, from the
   [first commit](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller) on, so that
   `git bisect` stays usable; the commits [question 43](#decisions) names
   are the exception and do not build. On the last commit of the series the
   unit tests pass, those of `pkg/rest` and those of the paging of the CSI
   lists ([question 63](#decisions)).
2. **Every call the plugin makes, against the recorded jdss1 v4 API**
   ([What "compliant" means](#what-compliant-means), [risk 17](#risks)).
   [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4),
   [rule_rest_request_body_valid](#rule_rest_request_body_valid),
   [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted),
   [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) and
   [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) pass
   for every call, and no recorded error of a plugin call is outside the
   catalogue. For [rule_rest_reply_types_decode](#rule_rest_reply_types_decode),
   the transcribed types differ from the OpenAPI only by Table 4's
   deliberate changes, treated as the implementation decides
   ([question 51](#decisions)); that they decode is for the fixtures to show
   ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).
   [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed)
   has no pass or fail: where the documentation and the recordings differ, the recordings are the source of truth (user decision), and the differences found so far stand in its row of Table 14, for the `jdss-rest-api-spec` project; the audit does not compare the two.
3. **NFS code.** Folded into step 5.

**Live test**

4. **The iSCSI build on API v4, against the lab appliance**
   ([risk 15](#risks)). Every CSI call that works before the series still
   works after it, and:
   - no reply fails to decode: the controller's log holds none of Go's
     `cannot unmarshal` errors, no required-field error of `checkRequired()`
     and no conversion error of `types_methods.go`, whose texts the items
     set ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors));
   - ValidateVolumeCapabilities, which reads the volume through `GetVolume`, ListVolumes and GetCapacity, which answer `Internal` before
     the series because a reply does not decode
     ([bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch),
     [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch)),
     reach the plugin's logic: their GET is answered 200, and a call that
     still fails does not fail with `Internal` and a decode message; it
     fails for another reason, which is a new finding;
   - every list request carries `per_page=100` and no `_dc`, and each CSI
     list call reads its pages in order from page 0, none twice
     ([bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page),
     [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)).
     The request log shows it, since it keeps the method, the path with its
     query string and the status at Debug
     ([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging));
   - every GET, and every DELETE of a target or a LUN, is answered as the
     [compliance matrix](#compliance-matrix) records, none with 400 or 415,
     so the `{}` body makes no difference on the lab appliance either
     ([question 30](#decisions));
   - a well-formed ID of a missing volume gets `NotFound` from the calls that must report it, such as ValidateVolumeCapabilities and ControllerPublishVolume, and a malformed
     one `InvalidArgument`
     ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids));
      - a volume that has a snapshot is not deleted: DeleteVolume answers
     `FailedPrecondition`, and the volume and its snapshot are still on the
     appliance ([bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children));
   - ListVolumes followed through its tokens, one entry at a time, returns
     every volume once; when the volume that the last token names is
     deleted before the next call, the rest of the walk still returns every
     volume after it once; and a token of today's form, such as `0_1_x`,
     gets `Aborted` while the controller keeps running
     ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)).
5. **The NFS build on API v4, against the lab appliance**
   ([risk 15](#risks)). No NFS build has run live before, so nothing is
   compared with an earlier run. What has to hold:
   - no reply fails to decode, and every list request carries `per_page`,
     as in step 4;
   - no call fails because of a finding this document fixes, in particular
     ListSnapshots by volume ([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)), deleting a volume that has clones
     ([bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised)), a repeated CreateVolume
     ([bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised)), deleting a clone that is already gone
     ([bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised)), and deleting a volume after a clone from it
     ([bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error), [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete));
   - once DeleteVolume has been sent for a volume that has a clone, the clone still
     exists on the appliance and can be read ([risk 23](#risks)). No
     csi-sanity spec reads a clone after deleting its source, so the case is
     made apart, on a volume and a clone of it made for it
     ([question 35](#decisions));
      - a volume that has a snapshot is not deleted, as in step 4 ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots));
   - any other failure is a new finding.

   If the NFS node path does not work in the lab at all, that is a result
   as well, recorded and not passed over.
6. **A volume deleted after a client was connected to it**
   ([gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients)).
   While a client uses the volume, its node holds a session to the volume's
   target and the volume is mounted. Once the client has stopped using it
   and the volume has been deleted, its target and its LUN are gone from
   the appliance
   ([bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone)),
   and no detach, target delete or volume delete has failed. The status of each
   of these calls and, for a failed one, its `class`, `errno` and message
   are the first evidence on v4 with a client connected. csi-sanity alone
   may never log a client in: the iSCSI NodeStage of the current code
   probably reports success without logging in or mounting (review `bug_plugin_node_iscsi_login_failure_ignored`), and its NodeUnstage never logs out (`bug_plugin_node_unstage_no_logout`). If no session appears, the gap
   stays untested, and that is recorded as a result, not passed over. Only
   the lab can show it.
7. **A volume unpublished and deleted while its client is still logged
   in** ([risk 8](#risks)), a recommendation. ControllerUnpublishVolume and
   then DeleteVolume are each sent once while the initiator is still logged
   in, as after a node that did not clean up; this is where unpublish and
   `force_umount: false` ([question 56](#decisions)) matter most. There is
   no pass condition: the replies are recorded, as evidence for risk 8
   either way. Only the lab can show it.
8. **A failed cycle.** Moved to kub-testing
   (`testcases/plugin-checks-design-0002.md`): finding the commit that
   caused it is a procedure, the guard of [risk 7](#risks).

**Limits**

9. csi-sanity reaches [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) only after design 0003
   stops the controller crashes; until then they have no live cover in this
   delivery, and neither has the NFS code that csi-sanity reaches only
   after design 0003.
10. **Not testable here:** [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) beyond steps 6 and 7 (several nodes, a pool moving
    between nodes); [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call); JovianDSS 2 live, which needs
    [JovianDSS 2 support](#future-work-joviandss-2-support) (a later version) and a
    jdss2 pool; `maxreplysize` ([impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit)), which the
    lab's replies, about 200 KB at most for a page of 100 entries against a
    default of 1 MiB, never reach; the handler of a GET answered 204 ([question 39](#decisions)), since no
    appliance has given that reply in any recorded run; a list longer than one
    page of 100 entries, which no step makes on the lab, so that reads across
    pages rest on hurl Finding 24 and on the unit tests of
    [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) ([question 63](#decisions)); and a list that changes while it is read ([risk 27](#risks)).
11. **No test of the certificate cases.** The lab appliance is reached with
    `cafile` or `insecure`, whichever its certificate allows
    ([risk 11](#risks)), and no test here covers the cases apart: a
    certificate imported into the appliance and trusted through `cafile`,
    an expired one, one that does not list the address, an unknown issuer, the two nodes of a cluster, and an appliance that offers only a TLS version below 1.2. Those tests are still to be developed
    (kub-testing `ISSUES.md`, T12).
12. **Driver and controller code without an automated check.** The driver
    and controller code this document changes has no automated check in
    this delivery (user decision), apart from the paging of the CSI lists,
    whose unit tests come with
    [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)
    ([question 63](#decisions)). Without one are [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)'s name and ID rules
    and its `iscsi.iqn` prefix check, [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)'s lines that print CSI requests, the delete cleanup of
    [bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children](#bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children) and [bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error](#bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error), the filter by kind ([question 7](#decisions)), the
    NFS cleanup reaching [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)'s call, the existing-share check
    ([question 57](#decisions)) and the `force_umount` flag ([question 56](#decisions)).
    kub-testing holds these as the checks of T11
    (`testcases/plugin-checks-design-0002.md`) for later; steps 4 and 5
    reach them as far as csi-sanity does, and step 4 also follows a list
    through its tokens. The callers changed by [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow) are not among them and have no automated
    check either. Nor has the clones read of the NFS
    `deleteIntermediateSnapshot` ([risk 23](#risks)), and the live steps
    are not expected to reach it: the state it guards needs a volume left
    half-created or two requests for one volume at the same time.

## Open questions
[Open_questions](#open-questions)

Open: none. The numbers are kept, gaps included, so that "question 27" means the same thing everywhere. The decided questions follow as a table, each with who decided it and the section that argues it.

### Decisions
[Decisions](#decisions)

**Table 9 — Decided questions** · `tbl_decisions`

| # | Question | Decision | Decided by | Where it is argued |
|---|---|---|---|---|
| 1 | Which API? | v4 only; v3 abandoned in favour of v4: the appliance still serves it, the plugin no longer calls it; no other option was considered. | the user | [aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only) |
| 2 | Minimum JovianDSS 1 release? | `1.0up33.65410`, the release of the lab and the recordings; a newer one if the live tests run on it ([question 58](#decisions)). | the user | [aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only), [risk 2](#risks) |
| 3 | When does token login come? | with JovianDSS 2 support, a later design. | the user | [Future work: JovianDSS 2](#future-work-joviandss-2-support) |
| 4 | Fixtures: real replies or hand-written? | real replies from the recordings, trimmed, passwords and the `Authorization` header removed. | the user | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) |
| 5 | Unused fields with the wrong type? | remove, as in 27; both superseded by 28: kept as raw JSON that cannot fail. | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 6 | Group the type fixes, or the two seen live first? | all together, delivered and tested in one go. | the user | [What changes](#what-changes) |
| 7 | Should the list functions filter by kind? | yes: each driver drops the other kind's entries; other values are kept. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 8 | Who closes the first eight evidence gaps under [Evidence gaps](#evidence-gaps), those the hurl suite can record? | the user adds the missing requests to the hurl suite; done for most, see the gaps. | the user | [Evidence gaps](#evidence-gaps) |
| 9 | `sparse` default? | in its own design, not this document: sent on every create, from the StorageClass parameter, else the plugin default, else sparse. | the user | [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) |
| 10 | Which commit carries the NAS snapshot delete switch? | the dead-code commit. | the user | [the dead-code commit](#before-stage-1-remove-dead-rest-code) |
| 11 | Is the lab appliance hurl's `jdss1`? | not a design question; every test appliance runs the latest release. | the user | [Testing](#testing) |
| 12 | Where do the REST-layer crashes go? | every crash fixable inside `pkg/rest` is fixed here, in stage 4, but for the nil body, which the handler of stage 3 closes when it replaces `ErrorFromErrorT`. | the user | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes) |
| 13 | Commit hash in the body? | no; the body says "the current code", Metadata names the commit. | the user | [Metadata](#metadata) |
| 14 | Keep the note that this replaces the original 0002? | not in the body, which carries no history; Metadata names it under Formerly. | the user | [Metadata](#metadata) |
| 15 | Tense? | the imperative for the changes, as in a Linux kernel commit message; the conditional for what would follow from them; the present for what is: today's code, the evidence, the decisions and the conditions that must hold; the past for what happened at one time; no future tense. Weighed against the conditional for the changes as well. | the user | [README, the mood and tense rules](README.md#how-the-text-is-organised) |
| 16 | Where do the TLS and path-escaping findings go? | this delivery, stage 4: in the TLS item and the path-escaping item. | the user | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 17 | Where do the delete-cleanup findings go? | this delivery, stage 3: one inside the error catalogue, one as its own fix there. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 18 | Where does the caller that turns every clone-delete error into Busy go? | this delivery, stage 3, in the error catalogue. | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |
| 19 | jdss2 tokens in fixtures? | tokens are secrets: never logged, never in a fixture; no jdss2 reply is a fixture now. | the user | [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) |
| 20 | Exact rules for the name check? | the rules of [Table 6](#impl_plugin_driver_validate_names_and_ids): the characters the appliance accepts in the strings tests, `a-z A-Z 0-9 _ -`; at most 131 characters, one below its limit of 132 ([question 21](#decisions)); targets in lower case only, which the appliance does not require. | the user | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) |
| 21 | Cap generated names at 131 or 132? | 131 for a name on the appliance; a generated name is shorter still, at most 128 bytes, since it is also the volume's CSI ID ([question 71](#decisions)). The volumes whose names the current code made longer are the accepted edge (risk 13). | the user | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) |
| 22 | Who prepares the NFS live cycle? | not this document: the test suite and the image name are prepared separately, before [Testing](#testing) step 5 can run; who prepares them is still for the user to decide, in the tracker (kub-testing `ISSUES.md`). | the user | [Relationship](#relationship-to-other-designs) |
| 23 | Is `prot: http` still allowed? | no; HTTPS only, `prot` other than `https` is a configuration error. | the user | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 24 | Where do release notes live? | `CHANGELOG.md` at the repository root. | the user | [Stage 1](#stage-1-move-to-api-v4) |
| 25 | Which of the four crashes in the ID and name code does [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) close? | none: the item stays as it is, and the four are handled separately. | the user | [Relationship](#relationship-to-other-designs) |
| 26 | TLS minimum as a setting? | no; TLS 1.2 or higher, fixed, since the lab appliance and jdss1 offer no TLS 1.3 ([question 50](#decisions)); a parameter of `pkg/rest` at most, no `tlsmin` key. | the user | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) |
| 27 | Keep or remove the unread fields that fail? | remove; superseded by 28. | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 28 | Generate the reply types instead of fixing them by hand? | yes, from the recorded jdss1 v4 API, same names, only read fields can fail; superseded by 38, which keeps the names and that rule. Weighed against: the rule per field (27), a transcription that drifts again with the next release, which only the audit would show; every field modelled by hand, which fails on a field nobody reads and drifts the same way; the whole client generated, which replaces about 2,100 lines and every caller and redoes the stage 3 and 4 work on generated code; tolerant decoding, which lets a real change in a read field pass silently. Price: about 650 lines of generator and mapping to keep, and [risks 19 and 20](#risks). | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 29 | Do older JovianDSS 1 releases serve the same v4? | no: they do not serve the same API v4, and there is no backward compatibility with them; nothing older than `1.0up33.65410` is supported. | the user | [Risk 2](#risks) |
| 30 | Does the `{}` body the plugin sends with every GET matter on jdss1 v4? | no: the replies with and without it are the same on jdss1 v4, as on jdss1 v3 and jdss2 v4. | the user | [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) |
| 31 | Which output of `jdss-rest-api-spec` is read? | the tracked OpenAPI, `empirical/jdss1/v4/openapi3/`: the path file of each call for the reply types, the schema of each kind of error, and the `-scenarios.json` beside the TypeSpec for which case gave which error. Not the build folder under `empirical/pilot/tools/tsp/`, which git ignores. Weighed against: the tracked `.tsp` files. | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 32 | Which read fields are optional? | read from `jdss-rest-api-spec`'s OpenAPI, not listed by hand: a read field is optional when the recordings show it missing or `null` at least once, and required when no recording does; so each call has one reply type. `origin`, `clones` and NAS `quota` are the ones the OpenAPI shows; `is_clone`, read from stage 3, is the one exception, optional by the user's decision although every recording has it ([question 65](#decisions)). A required field that an unrecorded appliance state drops fails by name ([risk 9](#risks)). Weighed against: those three listed by hand, with the list left open for testing to extend. | the user | [Table 4](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 33 | Where does the procedure for refreshing the reply types live? | in the plugin's documentation, since the types file is the plugin's; the page is written separately, not in this delivery. | the user | [Relationship](#relationship-to-other-designs) |
| 34 | Where does `cleanIntermediateSnapshots` take a snapshot's clones from, and when? | from `GET .../snapshots/<s>/clones`, asked for each intermediate snapshot before it is deleted, in both drivers and in the same commits as the list change; without waiting for the driver fix of the NFS snapshot deletes that destroy clones, which is outside this document ([Relationship](#relationship-to-other-designs)). The NFS `deleteIntermediateSnapshot` reads the clones the same way before its delete. | the user | [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow), [Stage 3](#stage-3-behaviour-and-error-mapping), [risk 23](#risks) |
| 35 | How does [Testing](#testing) step 5 observe that the clone survived? | by a read made by hand: on a volume and a clone of it made for the check, after DeleteVolume is sent for the source volume through the plugin, `GET pools/<pool>/nas-volumes/<clone>` on the appliance answers 200. No csi-sanity spec reads a clone after deleting its source, so the cycle alone cannot show it, and the check does not depend on the order in which csi-sanity deletes what it made; the check is kept in the tracker with the other driver checks (kub-testing `ISSUES.md`, T11). | the user | [Testing](#testing) step 5, [risk 23](#risks) |
| 36 | Where in the series do the two logging items go? | at the front, right after the `go vet` commit: no password or CSI secret may reach a log at any level, so no build that is run live may predate them; the log level is set in the configuration. Weighed against: leaving them in stage 4 and tracing a failure on builds at Info, which print no secret, since the Warn lines that could are unreachable (question 45), but show no request log. | the user | [Before stage 1](#before-stage-1-keep-secrets-out-of-the-logs), [risk 7](#risks) |
| 37 | What does "logged once" cover? | the REST module: the handler writes the one line of `pkg/rest` for a failed call, with the appliance's fields in full. The driver and the controller make their own log records for a failed operation; this document changes them only where they print a whole CSI request ([impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs)) and moves the warning of the name check into the driver with that check ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)). | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |
| 38 | Generate the reply types, or have an agent transcribe them? | transcribe: an agent reads `jdss-rest-api-spec`'s tracked OpenAPI and writes the Go types and the error catalogue; no generator, mapping file or compiler. Weighed against: generating the types from the tracked OpenAPI, set aside because of the optional fields: how well an automated tool would handle them is not known; compiling the TypeSpec with Node, a tool chain to keep; a `.tsp` parser in the plugin. Price: each new recording is a new reading and a review, and the match with the spec is shown by the audit's comparison, not by regenerating ([risk 17](#risks)). | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 39 | What does the plugin do with a GET answered without data, a 204? | a handler of its own in `pkg/rest` sends the GET again after a configurable pause, `tries` times in all at most; a 204 that a later request makes good is logged at Warn; 204 to the last one is an error naming the status, never an empty result. A DELETE answered 204 without data is a success and is not sent again. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 40 | What does a 404 mean when its body is not JSON, or matches no entry of its call? | "not found": the status decides, whatever the body; the entry a 404 matches says only what is missing. | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [risk 24](#risks) |
| 41 | What does the plugin do at startup when controller setup returns an error, as it would for a refused `iscsi.iqn` prefix? | not decided here: today the error ends in a panic, and the question is kept in the issue tracker. | the user | [Relationship](#relationship-to-other-designs) |
| 42 | Where does the iSCSI `GetSnapshot` take a snapshot's name from, when the zvol snapshot reply by name has none? | from its own request: the name it asked for, `sd.SDS()`, changed on the spot in stage 2; every function that calls `GetVolumeSnapshot` already holds the name. | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 43 | Do the readers of the per-volume snapshot lists build between stage 2 and stage 3? | no: they read `Name` of `ResourceSnapshot` on list entries, which the transcribed type does not have, until stage 3 moves them to the short entries; the commits between do not build on their own, and the series is tested once all stages are in. | the user | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), [Testing](#testing) step 1, [risk 7](#risks) |
| 44 | Does the plain-text 404 that a name with an escaped slash gets have an entry of its own, instead of the general not-found code? | no: `pkg/rest` takes no logic upon itself and answers a 404 with the error its status gives; question 40 stands, with risk 24 as its price. | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [risk 24](#risks) |
| 45 | What happens to a Warn line that could print a REST or CHAP credential? | it is removed, not cut down to fields: the five controller lines that print a whole CSI request when a capability check fails go, though that check cannot fail today. | the user | [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) |
| 46 | Should the ID check sit in `NewVolumeDescFromVDS`, where it would also cover the two uncovered entry points and the volume part of a CSI snapshot ID? | outside this document: kept in the issue tracker with `bug_plugin_ctrl_delete_volume_and_clone_source_ids_skip_csi_id_parser`, for later. | the user | [Relationship](#relationship-to-other-designs) |
| 47 | Where is the series compiled, as the analysis machine has no Go? | on the lab machine: the implementing agent writes the code and its unit tests on the analysis machine and builds, vets and runs the unit tests on the lab machine, commit by commit. Weighed against: the agent writing the whole series without a compiler, and the user building it afterwards. | the user | [What changes](#what-changes) |
| 48 | On the snapshot and clone creates, how does the catalogue tell a missing pool from a missing parent? | it cannot for now: `pkg/rest` takes an error that can be either as the parent missing, the volume on the snapshot creates and the snapshot on the clone creates, never as the pool missing, and its entry carries a `TODO` comment saying that the missing parent cannot be told from a missing pool. The two ways were an entry that also compares the names in the message with the request's path, and one code for "pool or parent missing". | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [risk 25](#risks) |
| 49 | Does the appliance serve 100 entries to a page? | yes for the volume list of jdss1 v4, with no cap (hurl Finding 24); the five other lists, and v3, are taken to do the same and are read with `per_page=100` without a run of their own. | the user | [Before stage 1](#before-stage-1-send-per_page-on-every-list), [risk 28](#risks) |
| 50 | Which TLS versions do the lab appliance and jdss1 offer? | not TLS 1.3, as the user checked: the plugin takes TLS 1.2 or higher. | the user | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), [risk 10](#risks) |
| 51 | How does the audit's type check treat what Table 4 changes on purpose? | left to the implementation: which fields the plugin would read and require is hard to tell before it; whether a read field is required or optional is derived from `jdss-rest-api-spec`'s data, as [question 32](#decisions) has it, `is_clone` excepted. The two ways were: the audit applies Table 4's rules before it compares, or it lists those differences apart, as expected ones. | the user | [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool), [Table 4](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 52 | Should every figure this document takes from `jdss-rest-api-spec` be read at one commit of it, or each at a commit of its own? | neither: the document gives no counts of the files and errors in `jdss-rest-api-spec`, which change as runs are added, and names the files and their paths instead; a fact that rests on one commit keeps that commit. The two ways were one commit for every figure, read again at it, and each figure with its own commit. | the user | [The observed API in `jdss-rest-api-spec`](#the-observed-api-in-jdss-rest-api-spec) |
| 54 | Which default does `maxreplysize` take? | 1 MiB, 1,048,576 bytes: lists are read in pages of 100 entries ([question 62](#decisions)), at most about 200 KB each; `per_page=0`, which returns the whole list in one reply, is not used. Weighed against 16 MiB. | the user | [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit), [How lists are paged](#how-lists-are-paged) |
| 55 | What are the name and the default of the setting for the pause between repeated GETs? | `retrydelay`, a duration written like `idletimeout`, and 1 s when the configuration does not set it. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 56 | Which `force_umount` do the iSCSI deletes send? | `false`; the plugin detaches a zvol from its target only by the call made for that, `DettachVolumeFromTarget`, which ControllerUnpublishVolume makes. DeleteVolume does not detach a zvol that is still attached; it only sends `false`. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping), [risk 8](#risks) |
| 57 | What does an NFS create do with a share that already exists? | it reads the share and takes it as its own only when its `path` is the one it sends; a share with another path fails the create. | the user | [Stage 3](#stage-3-behaviour-and-error-mapping) |
| 58 | Which minimum release do the docs state if the lab runs a newer JovianDSS 1 release when the live tests run? | the release the tests ran on; `1.0up33.65410` until then. | the user | [Testing](#testing), [Stage 1](#stage-1-move-to-api-v4) |
| 59 | Do the zvol snapshot and clone creates get out-of-space entries? | not in this document: no run has recorded their reply, so another record takes them up (kub-testing `ISSUES.md`, F30); until then such a reply is Unknown, logged at Error. | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |
| 60 | Does this document fix the paging of the CSI lists? | yes: the CSI lists are built on the appliance's paging, and how they must page follows from how it pages (hurl Finding 24); [rule_rest_lists_paged_as_served](#rule_rest_lists_paged_as_served) states it. | the user | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), [How lists are paged](#how-lists-are-paged) |
| 61 | Where does a CSI list call resume? | after the last name it returned; the token names no appliance page, and one the plugin did not issue gets `Aborted`. Weighed against: an appliance page in the token, today's form; a page kept beside the name as a hint. | the user | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) |
| 62 | Which page size do the list requests use? | 100. Weighed against 1000, which reads up to 1000 entries in one request at the same cost, since every request makes the appliance build the whole list. | the user | [Before stage 1](#before-stage-1-send-per_page-on-every-list), [risk 26](#risks) |
| 63 | Do the paging functions get a check in this delivery? | yes: unit tests, written by the implementing agent with [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) and run by `make test` in [Testing](#testing) step 1, the one exception to the rule that the driver and controller code this document changes has no automated check in this delivery. Weighed against: the kub-testing check T11.10 for later, with the walk of step 4 as the only cover until then. | the user | [Testing](#testing) step 1 and limit 12, [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) |
| 64 | How often does a CSI list call read the appliance list, and what becomes of a name that comes twice? | once, from page 0 to the last page, and not again when the list changed during the read; of two entries with one name, one is dropped when they are the same in the fields of [question 65](#decisions). Weighed against: a second read when `results` changes during the first, and `Unavailable` when it changes again. | the user | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), [risk 27](#risks) |
| 65 | When a name comes twice in one read, which fields are compared, and what does a call do when they differ? | `default_scsi_id`, `guid` and `is_clone`, as far as the list has them: all three in the volume list, `properties.guid` in the four snapshot lists, none in the dataset list; `is_clone` is optional, `guid` and `default_scsi_id` are required. When they are the same, the entry read later stays; when they differ, both stay. Weighed against: every field; the fields of the CSI entry, with the later entry kept when they differ. | the user | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), [Table 5](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) |
| 66 | What does `max_entries` count? | only the entries the controller sends back in its reply: names the plugin did not make, and the other driver's snapshots, are dropped in the driver before the list is cut. Weighed against: the cut before those checks, today's order. | the user | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name) |
| 67 | What do the certificates of the lab appliance and jdss1 name? | this document does not need them: the CA file that verifies an appliance's certificate is given in the configuration file (`cafile`); kub-testing T12 reads the current certificates for its connection tests. | the user | [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), [risk 11](#risks) |
| 68 | How does the list token stay within CSI's limit of 128 bytes for a string? | it holds a bound instead of the last key: the shortest string that sorts after the key of the last entry returned, or equals it, and before the next one; a hash stands for the volume of a per-volume list; where the bound does not fit, its first bytes and a hash of each of the two keys ([risk 29](#risks)). Weighed against: the last key in full, which a name of 131 characters already makes too long. | the user; the form proposed by the agent | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), [Code segments](#code-segments) |
| 69 | Do the unit tests of `pkg/controller` run under the vet of `go test`, which the 23 format strings of that package fail? | yes: the 23 are fixed in the first commit, with those of `pkg/rest`, so that `make test` runs `go test` on the package as it is; the node's three stay for later. Weighed against: `-vet=off` for that package in `make test`, which leaves the controller's format strings for later and its tests without vet. | the user; proposed by the agent | [impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text) |
| 70 | Which `RestError` code does each kind of error get in the catalogue, and which CSI status does each code get from the controller? | the CSI status is out of scope. This document implements API v4 in the REST module and gives the modules that use it one code for each failure. How those modules handle a code, including what they reply to Kubernetes, is not decided here. The caller changes that stage 3 makes by earlier decisions stay. Which code each kind of error gets is [question 73](#decisions). | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [Relationship](#relationship-to-other-designs) |
| 71 | Does the CSI ID of a volume have to stay within CSI's limit of 128 bytes for a string, as the list token does? | yes. CSI naming follows the CSI spec, so a generated volume name, which is also the volume's CSI ID, is at most 128 bytes: both thresholds of `NewVolumeDescFromName` become 125. Volume and share names on the appliance stay within 131 ([question 21](#decisions)). Weighed against: thresholds of 128, which allow IDs of up to 131 bytes. | the user | [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids), [risk 13](#risks) |
| 72 | Can this document still be edited after it is accepted? | it stays Proposed while the series is implemented and tested, and is accepted with the tested series; until then, what the implementation changes in it is written into its text. Weighed against: amendments to an accepted record, listed in Advice; the accepted text left alone, with the changes kept in the tracker and the commit messages. | the user | [What changes](#what-changes) |
| 73 | Which `RestError` code does each kind of error get? | a code inside the range of its layer, which the `class` of the error shows. Each layer has 1000 codes: 1000 to 1999 without a class, 2000 to 2999 for `builtins.*`, 3000 to 3999 for `zfslib.*`, 4000 to 4999 for `opene.*`, 5000 to 5999 for `jsonschema.*` and 6000 to 6999 for `werkzeug.exceptions.*`. The multiple of 1000 is the code of the layer itself, and the codes below 1000 stay the plugin's own. What today's callers compare with is [question 74](#decisions). | the user; the four ranges after the first two assigned by the agent, in the order of the study's table | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |
| 74 | What do today's callers compare with, once an appliance error has a code of its layer? | with the code of today. An error carries two codes: the one that says what it means for the call, which the callers keep reading unchanged, and beside it the code of its kind in the range of its layer, or of the layer alone when no entry of the call holds the error. Weighed against: the code of the layer only, with every caller changed to compare with it. | the user; proposed by the agent | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [Code segments](#code-segments) |
| 75 | Who numbers the kinds of error inside the range of a layer, and when? | they are numbered in this document, before the series is started: Table 19 gives each kind of error of the plugin's calls its code, in the order of the names inside each layer. A kind keeps its number, and one that the spec gains later takes the next free number of its layer. Weighed against: the implementing agent numbering the kinds as it transcribes the catalogue. | the user | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), [Measurements](#measurements) |
| 76 | How does `make vet` give a true result when the build cache holds an older one? | it runs `go vet -a`, which analyses the packages again: with Go 1.26.0 a package vetted before as a dependency of another one gets no report from a plain `go vet`. Weighed against: an empty build cache for every vet run; a Go release that does not have the fault, which has not been looked for. | the user; proposed by the agent | [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply), [risk 30](#risks) |
| 77 | Which code does a catalogue entry give for `GetCode()`, the code today's callers read? | the code that the call returns today for that reply, Unknown included; another code only where this document decides one: the fixes of stage 3, and a 404, which is "not found" ([question 40](#decisions)). Weighed against: the implementing agent choosing a code wherever it takes the meaning of an error as known. | the user; proposed by the agent | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) |

## Relationship to other designs
[Relationship_to_other_designs](#relationship-to-other-designs)

- Findings are tracked in `kub-testing/ISSUES.md` and `kub-testing/project-status.md`;
  the review files and ID prefixes are in [ID sources](#id-sources).
- `jdss-rest-api-spec`: the reference for compliance
  ([The `jdss-rest-api-spec` repository](#the-jdss-rest-api-spec-repository)).
  Only read; its tracked OpenAPI of the observed types is the
  source of the reply types, and its scenarios files of the error catalogue ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)), refreshed by the user with
  `build_typespec.py --check`. Also to report: a name with non-ASCII characters gets **500
  `UnicodeDecodeError`** from the appliance (zvol, NAS and snapshot strings tests)
  instead of a 400 validation error, which is a server bug. The [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed) differences are
  a list for the user to pass on to the `jdss-rest-api-spec` project; this work does not change `jdss-rest-api-spec`.
- JovianDSS-Proxmox plugin: its `jdssc` already has typed errors per kind
  (`jdssc/jdssc/jovian_common/exception.py`) and one general error path
  (`jovian_common/rest.py`, `_general_error`); [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue) follows the same idea.
- hurl suite (`jdss-rest-api-spec/empirical/hurl/`): it plans to report the bugs it finds
  in API clients such as this plugin; the findings its recordings revealed (every entry of [error messages](#error-messages), and the entries of [response types](#response-types), [response content and defaults](#response-content-and-defaults) and [logic bugs](#logic-bugs) that `kub-testing/issues/test-findings.md` registers), [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check),
  [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status), [bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) are candidates for that report. Its Finding 4.4 (a delete that
  passes the create status constant) can be downgraded, because the constant is
  ignored ([Key observation 7](#key-observations)). Its open question on `ListAllSnapshots` (does the
  plugin drop the other kind of snapshot from the combined list?) is answered by
  [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key): today no plugin code filters by kind,
  only the name-format check in the list functions can drop an entry; stage 3 filters by kind ([question 7](#decisions)).
- Backlog (`review/backlog.md`): item 36 is [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch); item 37 (snapshot
  time format) has been verified correct and can be closed; items 19 and 20 (the
  configuration keys the configuration guides get wrong, and the missing `yaml` tags) are addressed by [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys).
- [design 0003, controller crash fixes](../design/0003-controller-crash-fixes-design.md): deferred. Its
  design item D2 (`crash_plugin_ctrl_create_volume_on_rest_error`) would remove the crash that [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) causes in
  CreateVolume. With this document alone, [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch)'s error would go away, so [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) would no longer
  trigger `crash_plugin_ctrl_create_volume_on_rest_error`. The REST-layer [crashes](#crashes) are fixed here ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)); design 0003 keeps the
  controller's own crashes; making startup fail on the error that `SetupEndpoint` returns once [crash_plugin_rest_setup_endpoint_ignores_error](#crash_plugin_rest_setup_endpoint_ignores_error) is fixed is driver work this document leaves for it.
  Design 0003 leaves the paging of the lists out; the crash in the parse of the
  list token, [crash_plugin_driver_list_token_parse_writes_through_nil_pointer](#crash_plugin_driver_list_token_parse_writes_through_nil_pointer),
  is fixed here with the paging ([question 60](#decisions)).
  Design 0003's D3 (the GetCapacity logger fix, `crash_plugin_ctrl_get_capacity_no_logger`) is in the current code: the series is built on it, as was the cycle before the series,
  which [Testing](#testing) step 4 compares with.
- Four crashes in the ID and name code next to the parsers that
  [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) changes: `crash_plugin_ctrl_create_snapshot_empty_name`,
  `crash_plugin_driver_snapshot_name_empty`,
  `crash_plugin_driver_volume_desc_empty_value` and
  `crash_plugin_driver_id_index_without_length_check`. None is fixed here;
  the item does not close any of them (kub-testing `issues/crashes.md`, for
  design 0003). The item's ID check also misses two CSI entry points, the volume ID
  of DeleteVolume and the source volume ID of CreateVolume, which reach
  `NewVolumeDescFromVDS` without passing the CSI parser
  (`controllerplugin.go:550`, `:345`): `bug_plugin_ctrl_delete_volume_and_clone_source_ids_skip_csi_id_parser`
  (kub-testing `issues/security.md`), not covered here and kept in the tracker for later. Whether
  the check belongs in `NewVolumeDescFromVDS`, where it would also run on
  the volume part of a CSI snapshot ID (`snapid.go:132`), is kept with it in
  the tracker, outside this document ([question 46](#decisions)).
- What the plugin does at startup when controller setup returns an error
  ([question 41](#decisions)): `main.go` drops the error and calls `Run()`
  on a server that was never made, so the controller ends in a nil-pointer
  panic (`crash_plugin_ctrl_main_ignores_setup_error`, kub-testing
  `issues/crashes.md`; `app/joviandssplugin/main.go:124`, `:126`). Not
  fixed and not decided here. The `iscsi.iqn` check of
  [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) adds one more setup error
  that ends there; design 0003 holds the crash as its open question 3.
- `sparse` provisioning ([bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse)): sparse by default,
  configurable in the plugin configuration and per StorageClass; its own design
  ([question 9](#decisions)).
- The NFS live cycle ([Testing](#testing) step 5): making the test suite and
  the build able to run one is handled separately (kub-testing `ISSUES.md`,
  TOOL10 and MK11; who prepares them is still for the user to decide). Step 5 stays a requirement of this
  delivery.
- The procedure for reading the reply types again after new recordings
  ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)): a page in the plugin's documentation, written
  separately and not in this delivery (kub-testing `ISSUES.md`, DOC27).
- Format strings outside `pkg/rest`: the controller and the node pass error
  text to `status.Errorf` as the format, 26 of the 34 `go vet` reports
  (`bug_plugin_error_text_as_grpc_status_format_string`, kub-testing
  `issues/data-loss.md`). The controller's 23 are fixed here, in the first
  commit, because `go test` does not run the tests of a package that vet
  reports on
  ([impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text),
  [question 69](#decisions)); the node's three are not, and are fixed later
  with that module.
- NFS snapshot deletes that destroy clones
  (`dataloss_plugin_driver_nfs_delete_snapshot_destroys_volumes_cloned_from_it`,
  `dataloss_plugin_driver_nfs_intermediate_snapshot_delete_destroys_clone_volume`):
  driver bugs, handled outside this document; described in kub-testing,
  `issues/notes/dataloss-nfs-snapshot-delete-destroys-clones.md`. This
  document only keeps its own commits from opening the second one
  ([risk 23](#risks), [question 34](#decisions)); the first exists today
  and stays.
- The CSI status of a REST error: out of scope (user decision,
  [question 70](#decisions)). This document implements API v4 in the REST
  module and gives the modules that use it one code for each failure. Which
  CSI status the driver and the controller return for a code, including the
  new codes, is left to the work on those modules (kub-testing `ISSUES.md`,
  F32). The caller changes made here are the ones
  [Stage 3](#stage-3-behaviour-and-error-mapping) lists.
- REST client behaviour: not a compliance matter. Of its catalogue
  ([REST client behaviour](#rest-client-behaviour)) this delivery leaves
  [bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover) and
  [bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied) to a later REST client
  design; the other three entries are fixed here. Of the first, stage 3
  uses `tries` for one case only, a GET answered without data
  ([question 39](#decisions)); a request that fails in transport is still
  not sent again, and the other addresses are still not used.
- Security: every [security finding](#security-findings) is assigned to this delivery, to [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments) and [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile) in [stage 4](#stage-4-robustness-and-security) and to [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs) [before stage 1](#before-stage-1-keep-secrets-out-of-the-logs); stage 4 also adds [impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) and [impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit), and [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) is the
  [first commit](#first-commit-make-go-vet-pass-in-pkgrest-and-pkgcontroller).
- A build of the series is made on the lab machine from the tree the series is written in: the plugin checkout there has to be in step with that tree before each build (kub-testing `ISSUES.md`, INF8).
- JovianDSS 2 is outside this document and left to a later design ([Future work: JovianDSS 2 support](#future-work-joviandss-2-support)).
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
at the same time. It would touch `rproxy.go`, the configuration and secrets. This document
already stops the logging of request bodies and headers ([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging)). No token, refresh or access, may ever be logged (user decision, [question 19](#decisions)): any
logging the token design adds must keep both out of the logs
([security_plugin_secrets_in_logs](#security_plugin_secrets_in_logs), [security_plugin_rest_request_dump_every_call](#security_plugin_rest_request_dump_every_call)).
Testing would need a jdss2 pool. This document does not touch JovianDSS 2; it
prepares for it only by sending `per_page` on every list request, without which jdss2 v4 returns the first page again and again ([bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page)), by giving 401 and 403 an error code of their own ([bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status)) and
by keeping the jdss2 recordings beside the jdss1 ones as evidence. That design would also have to allow for replies that are
behind the request, which jdss2 shows most ([risk 14](#risks)).

### Future work: request deadlines
[Future_work_request_deadlines](#future-work-request-deadlines)

Not part of this delivery; for the later REST client design. That design would address [bug_plugin_rest_request_built_without_context_csi_deadline_not_applied](#bug_plugin_rest_request_built_without_context_csi_deadline_not_applied) and [risk 12](#risks). Every CSI call arrives with a
deadline, and Kubernetes may cancel it; today neither reaches the HTTP request,
which runs until `idletimeout` expires. The request would be built with the CSI
call's context (`http.NewRequestWithContext`), so the deadline and the cancellation
would stop the HTTP call too; `idletimeout` would then be only the idle timeout of the
connection pool, as its name says, not the timeout of every request.

### Future work: snapshot name limits
[Future_work_snapshot_name_limits](#future-work-snapshot-name-limits)

Not part of this delivery; to consider later. Snapshot names follow ZFS's rules, not
the appliance's own check, so a bad one is refused late, with a 500 instead of a 400 validation error: by `zfs snapshot`, with `ZfsCmdError`, or, for a non-ASCII character, with `UnicodeDecodeError`. What the hurl strings tests
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
which today allows up to 243 characters) would have to stay within it. The
reply to a name that is too long is already in this delivery:
[impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)'s
catalogue takes it from the scenarios files of the two snapshot creates,
which list it as 500, `ZfsCmdError`, errno 201, `cannot create snapshot '...':
name is too long`.

The CSI ID of a snapshot is built from the snapshot's name and its volume's
(`driver/snapid.go:57-59`), so it can be longer than the 128 bytes the CSI
spec allows for a string. CSI naming follows the CSI spec
([question 71](#decisions)); that work would bring the snapshot's ID within
the limit too.

### Future work: missing fields in replies
[Future_work_missing_fields_in_replies](#future-work-missing-fields-in-replies)

Mostly covered by [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec) ([question 38](#decisions)), which addresses
[risk 9](#risks) for the fields the plugin reads: each is marked required or
optional in the source, a required one absent from a reply fails the call with
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
hand in the source until a recording confirms it. The required-or-optional
list already rests on the recordings, through the OpenAPI, but for `is_clone`, optional by the user's decision ([questions 32 and 65](#decisions)).

## Advice
[Advice](#advice)

- hurl suite, Finding 9 (2026-11-04): JovianDSS 2 serves v4 only and accepts
  only tokens.
- hurl suite, jdss1 run `20261107-010938` (2026-11-07, read 2026-11-13):
  the gap tests on jdss1 v4; the `{}` body, paging without `per_page`, the
  default provisioning, the name limit and the missing-object replies as on
  v3, the snapshot list of a missing volume answering 404.
- hurl suite, Findings 14 and 15 (2026-11-04): v3 and v4 identical on
  JovianDSS 1 in 363 delete cases and 281 rollback and clone cases.
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
- User decision (2026-11-05): v4 only.
- User decision (2026-11-09, question 27, superseded): remove the unread
  mismatching fields.
- User decisions (2026-11-10): generate the types, names kept, jdss1 v4
  only, only read fields can fail with optional where a value may be
  absent, drivers updated; one record for the whole problem.
- User decisions (2026-11-10): JovianDSS 2 is not touched by this document;
  stages 3 and 4 belong to the same migration, so one record
  covers the whole migration and is implemented as one series; its
  questions are kept here with their numbers.
- User decision (2026-11-11): older JovianDSS 1 releases do not serve the
  same API v4; there is no backward compatibility with them (question 29).
- `go vet ./...`, run by the user on commit `ef4ea8b` (2026-11-14; `pkg` and
  `app` as in `11ffbfd`): 34 reports, all "non-constant format string", in
  `pkg/rest` (8), `pkg/controller` (23) and `pkg/node` (3).
- User decision (2026-11-15, question 34): the cleanup takes a snapshot's
  clones from `GET .../snapshots/<s>/clones`, the safest source.
- jdss-rest-api-spec, commit `a34abc0` and a note from its session (read
  2026-11-16): one error model per kind of error and a scenarios file per
  method, tracked beside the TypeSpec; the compiled OpenAPI is a build folder
  that git ignores; a tracked `openapi.json` is planned.
- User decision (2026-11-16, questions 31 and 38): the reply types and the
  error catalogue are transcribed by an agent from the tracked TypeSpec; no
  generator.
- User decision (2026-11-16, question 32): the optional read fields are
  worked out from the recorded replies, through the TypeSpec, so that each
  call has one reply template; the fixtures of a call are the few recorded
  replies that exercise it, not every recorded reply.
- jdss-rest-api-spec, commit `94d0c4f` and a note from its session (read
  2026-11-17): the `nospace/` tests on jdss1 v4, run `20261116-223132`; what
  a zvol and a NAS volume create return when the pool has no room; snapshot
  and clone creates on a full pool, v3 and jdss2 still to come.
- User decision (2026-11-17): the error type tells three cases of `errno`
  apart, a number, `null` and a missing key, with the extra code that needs.
- User decision (2026-11-17): out of space gets exact catalogue entries;
  it is no longer recognised by its message whatever the errno.
- Review of this document (2026-11-17): the NFS
  `deleteIntermediateSnapshot` deletes an intermediate snapshot without
  reading its clones, a second way to the loss of risk 23; a GET answered
  204 has no data to return as "the empty result"; a controller setup
  error ends in a panic at startup.
- User decision (2026-11-17, question 34): the clones read also covers the
  NFS `deleteIntermediateSnapshot`, in this delivery.
- User decisions (2026-11-17, question 39): a GET answered without data
  gets a handler of its own and is sent again; the number of requests is
  the existing `tries` setting; the pause between them is configurable;
  when the last one is answered 204 too, the call fails with an error
  naming the status; a 204 that a later request makes good is logged at
  Warn.
- User decision (2026-11-17, question 40): a reply with status 404 is
  treated as "not found".
- User decision (2026-11-17, question 41): what the plugin does at startup
  when controller setup returns an error is outside this document and is
  kept in the issue tracker.
- User decision (2026-11-17, question 38): transcription was chosen because
  the replies have optional fields, and it is not known how well an
  automated tool would handle them.
- jdss-rest-api-spec, the scenarios files of the plugin's calls on jdss1 v4
  at commit `94d0c4f` (read 2026-11-17): 34 of the 333 recorded errors are
  404s, each for something that does not exist.
- Review of this document (2026-11-17, the second that day): the Warn lines
  that print CSI requests cannot run; v3 and v4 differ on the per-volume
  snapshot lists; the zvol snapshot reply by name has no `name`, which the
  iSCSI `GetSnapshot` and the readers of the per-volume lists read; on the
  snapshot and clone creates a missing pool and a missing parent can share
  one error pattern; the stop rule takes a page of 100 entries for granted;
  the audit's comparison and Table 4's deliberate changes disagree; two spec
  commits are quoted.
- User decisions (2026-11-17, questions 42 to 47): `GetSnapshot` takes the
  name it asked for; the readers of the per-volume lists build again only in
  stage 3, the series being tested once all stages are in; question 40
  stands, `pkg/rest` taking no logic upon itself; a Warn line that could
  print a REST or CHAP credential is removed; where the ID check sits goes to
  the issue tracker; the user compiles on another machine.
- User decisions (2026-11-17): the overlap of the missing-pool and
  missing-parent errors is a finding and risk 25, with open question 48; the
  page size the stop rule relies on is open question 49, to be answered by the
  user's scaling tests; the read fields are listed per type (Table 5); the
  questions of the review that were not answered are open, 48 to 53.
- jdss-rest-api-spec, commit `94d0c4f` (read 2026-11-17): the scenarios
  files and error schemas of the four snapshot and clone creates; the
  per-volume snapshot lists on jdss1 v3 and v4; the read fields of the reply
  types, the same at `6149fa9`.
- User decisions (2026-11-17, the third round that day, questions 56 to
  59): the iSCSI deletes send `force_umount: false`, and a zvol is detached
  from its target only by the call made for that; an NFS create takes an
  existing share as its own only when its `path` is the one it sends; the
  minimum release stated is the one the live tests ran on; the out-of-space
  entries of the zvol snapshot and clone creates are left to another
  record. Also: the Approver is named; success includes no new
  vulnerability, and the removal of the existing ones within the scope; v3
  is abandoned in favour of v4, still served by the appliance and no longer
  called by the plugin; the lab machine is named by its role; the two
  findings that covered more than their names said are split.
- jdss-rest-api-spec, commit `434c9cd` (read 2026-11-17): hurl Finding 24,
  the paging of the volume list on jdss1 v4: 20 entries to a page by
  default, pages from 0, any `per_page` served in full with no cap,
  complete walks with no miss and no repeat, and pages that shift when the
  list changes during a walk.
- User decisions (2026-11-17, the fourth round that day, questions 60 to
  62): the paging of the CSI lists is part of this document, because it
  rests closely on how the API pages; the paging proposed with the review
  of the third round is taken: every list call reads the whole list and
  resumes after the last name it returned; the page size stays 100; and
  DeleteVolume does not detach a zvol that is still attached, it only
  sends `force_umount: false` (question 56).
- User decisions (2026-11-18, question 64): a CSI list call reads the
  appliance list once, from page 0 to the last page, and not again when
  the list changed during the read; of two entries with one name, one is
  dropped when their other fields are the same. Which fields are compared,
  and what happens when they differ, is open (question 65). The bullet of
  the paging item on what `max_entries` counts was reworded, since the
  user found it unclear.
- User decisions (2026-11-19, questions 63, 65 and 66): the paging
  functions get unit tests in this delivery, the one exception for the
  driver and controller code; a name that comes twice in one read is
  compared on `default_scsi_id`, `guid` and `is_clone`; and `max_entries`
  counts only the entries the controller sends back in its reply, as the
  reworded bullet proposed.
- User decisions (2026-11-19, the second round that day, question 65): of
  two entries of one name, the one read later stays when the compared
  fields are the same, and both stay when they differ, as proposed; and
  `is_clone` is optional, since it may be missing, although every
  recording has it, while `guid` and `default_scsi_id`, always present,
  stay required.
- User decisions (2026-11-19, the third round that day, questions 49 and
  50): the five other lists, and v3, are read with `per_page=100` as the
  volume list is, without a run of their own (risk 28); the lab appliance
  and jdss1 offer no TLS 1.3, as the user checked, so the plugin takes
  TLS 1.2 or higher, which changes question 26's minimum; what their
  certificates name stays open, as question 67. Question 53 was withdrawn:
  it concerned the README, not this document.
- User decision (2026-11-19, the fourth round that day, question 52): no
  counts of the files and errors in `jdss-rest-api-spec`, which change as
  runs are added; the document names the files and their paths instead.
- User decision (2026-11-19, the fifth round that day, question 67): the CA
  file that verifies an appliance's certificate is given in the
  configuration file (`cafile`), so this document does not depend on what
  the current certificates name.
- User decision (2026-11-19, the sixth round that day, question 54):
  `maxreplysize` defaults to 1 MiB instead of the proposed 16 MiB: lists
  are read in pages of 100 entries, at most about 200 KB each, and
  `per_page=0`, which returns the whole list in one reply, is not used.
  Also: a list request without `per_page` gets 20 entries to a page on v4
  (hurl Finding 24), not the whole list, which the `per_page` commit and its finding had said; and the Consequences still had stage 4 refuse a TLS
  version below 1.3, where question 50 set the minimum at TLS 1.2.
- User decisions (2026-11-19, the seventh round that day, questions 54 and
  55): the default of `maxreplysize` is 1 MiB, as written; the pause
  between repeated GETs is `retrydelay`, 1 s when the configuration does
  not set it; and the plugin takes TLS from version 1.2, as the user
  confirmed. On paging, in the user's words: a list request without
  `per_page` returns pages of 20 and `page` still works; everything on
  page 0 only looks true of a list of 20 objects or fewer; `per_page=0`
  lists every element in one reply.
- User decisions (2026-11-20, questions 48 and 51): on the snapshot and
  clone creates a missing pool cannot be told from a missing parent for
  now, so `pkg/rest` acts as if the volume were missing, taking such an
  error as the parent missing, and a `TODO` in the code says that the two
  cannot be told apart; which fields the plugin would read and require is
  hard to tell before the implementation, so how the audit treats Table 4's
  deliberate changes is left to it, and whether a field is optional or
  present is derived from `jdss-rest-api-spec`'s data. No question is open.
- User decisions (2026-11-20, the second round that day): the questions
  decided before the Decisions table had a "Decided by" column, 3, 4, 6,
  7, 9 to 18, 24, 30 and 35, were decided by the user. Also: the paging
  item describes its changes in the imperative mood, as the README now
  asks; rows 31, 32 and 61 to 66 of the Decisions table give the decision
  and its choices without dates or history, and the arguments of rows 61,
  63, 64 and 65 are in the paging item.
- User decision (2026-11-20, the second round that day), applied in the
  third: a Testing step names its corner case, what has to hold there
  and what it comes from, and leaves who runs the test, what it needs
  first and its commands to the test. Steps 1 to 7 are written so; what
  they held of that kind, and step 8, finding the commit after a failed
  cycle, are in kub-testing (`testcases/plugin-checks-design-0002.md`).
  The driver and controller code without an automated check moved from
  step 1 to limit 12, and question 47 is argued in What changes.
- User decision (2026-11-20, the second round that day), applied in the
  fourth: What changes gives its changes in the imperative mood, as a
  Linux kernel commit message does, and question 15 is answered so; what
  would follow from a change, and the ways not taken, stay in the
  conditional, and the meaning is unchanged. Following question 65, the
  What must hold of the reply types and the definition of the fixtures no
  longer take the mark of `is_clone` from the OpenAPI.
- User decision (2026-11-20, the fifth round that day): the README gains a
  tense rule: the present for what is, the past for what happened, and no
  future tense. This document follows it where it used "will", and
  question 15's answer names it.
- Review of this document (2026-11-20): four readers, one per part, checked it
  against itself and the README. Corrected: facts stated two ways (the
  minimum release, `tries`, the TLS version as an upgrade price, the Table 8
  entry for `prot`, the request-body rule, the code of an unknown error);
  What must hold of the items that missed a change they order; links to
  a Testing limit called a step, and to the wrong item; Consequences and
  the effects in Risks put in the conditional; dates, history and counts
  of `jdss-rest-api-spec` errors left in the body and the Decisions rows;
  terms used in senses Terms did not give; and design 0003 cited by its
  bare number.
- User decision (2026-11-21): the code of a critical algorithm is written
  out in the Appendix, under Code segments; the first segments were the
  paging of the lists: the read, the merge of its pages, and the order and
  the cut.
- User decision (2026-11-21), in the README: "Problem statement formalizes
  the scope and what has to be done. Scope limitations and explicit
  mentions of what will not be done or accomplished in given ADR are listed
  in Options considered and accepted in Decision section." The Problem
  statement is now the README's example of one ("api v4 compliance jdss1"); that JovianDSS 2 is not
  supported within this iteration, and that the code this document adds or
  changes brings no new vulnerability while the existing ones within its
  scope are removed, move to the Decision;
  [aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only) says how far
  the work reaches. The paging item links its code segment last, as the README's
  skeleton shows.
- User decision (2026-11-21): the most critical points of this document are
  drafted in code segments, for the implementations that follow them.
  Chosen for it, besides the paging of the lists: the setup of the client,
  one request and its reply, a GET answered without data, the error body, a
  failed reply, a reply type, the token of a CSI list and the delete of a
  snapshot without clones. Each segment names the files it goes into, and
  each item it implements links it.
- Review of this document (2026-11-21, the fourth round that day): seven
  readers, one per part, checked it against itself, the README, the code
  and `jdss-rest-api-spec`. Corrected: facts stated two ways (the paging
  of v3 and of the lists other than the volume list, who reads a
  snapshot's clones once stage 3 is in, the Debug line of a request, the
  read past `maxreplysize`, the owner of the NFS live cycle); code
  segments that did not do what their items say (two catalogue entries
  whose patterns overlap, in the wrong order; the Warn line after a 204;
  the files the code goes into); claims stronger than the evidence or the
  code; wrong line numbers, names and links; the limits of the scope in
  the Decision; decisions put in the conditional and consequences in the
  present; history and lab details in the body; and jargon missing from
  Terms. The two patterns of the LUN detach that overlap come from
    `jdss-rest-api-spec` (`ItemNotFoundErrorTargetNotExists` matches the
  message of `ItemNotFoundErrorTargetPartitionNotExists`), to be passed on
  to its project.
- User decisions (2026-11-22), on the questions of that review: the list
  token is shortened to fit CSI's 128 bytes ("design a shorter token";
  question 68, risk 29); the nil-body crash is closed in stage 3, by the
  handler that replaces `ErrorFromErrorT` ("make it stage 3"); question 48
  stays as decided ("keep 48 as decided"); Testing step 5's check of
  ListSnapshots by volume, which risk 4 and limit 9 say csi-sanity reaches
  only after design 0003, is left as it is ("lets ignore for now"); the
  reply types get the sub-types their read fields need, `ResourceSnapshotParent`
    for the `parent` of a zvol snapshot list entry ("add all needed
  sub-types").
- User decisions (2026-11-22, the second round that day): no volume that
  has snapshots is deleted, on iSCSI and NFS alike ("both iscsi and nfs
  volume should not be deleted if they have snapshots"); the NFS driver,
  whose appliance deletes a NAS volume with its snapshots (hurl Finding 7),
  looks for them first ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots)). For the REST API the
  empirical findings are the source of truth over the documentation ("in
  terms of rest api trust empirical findings over documentation as source
  of truth"): the audit does not compare the two, and the differences
    found stay listed for the `jdss-rest-api-spec` project.
- User decision (2026-11-22, the third round that day): the fixes of
  backlog items 19 and 20, the `yaml` tags of the endpoint configuration
  and the keys the configuration guides get wrong, leave the TLS item
  for an item of their own ("a, give them their own item"),
  [impl_plugin_config_yaml_tags_and_guide_keys](#impl_plugin_config_yaml_tags_and_guide_keys), which also covers
  `docs/configuration-nfs.md` and `ISCSIEndpointCfg`.
- User decision (2026-11-22), in the README: an option is named
  `aopt_<area>_<slug>`, like every other named entry, and not by a letter
  and a number; Options considered does not say which option is adopted,
  the Decision names it. The one option here, until then "Option A: v4
  only, the only option considered", is
  [aopt_rest_api_version_v4_only](#aopt_rest_api_version_v4_only).
- User decisions (2026-11-23): only a GET answered 204 is sent again, and
  a DELETE answered 204 without data is a success, as it is today ("for
  DELETE 204 and no data is a success"; question 39); both thresholds of
  `NewVolumeDescFromName`, the CSI name's and its base32 form's, go down to
  128 ("please mark that length should be reduced"); a 404 for a path the
  appliance does not route, which risk 24 describes, is left to another
  record (kub-testing `ISSUES.md`, F31).
- User decisions (2026-11-23, the second round that day): the GetCapacity
  logger fix, design 0003's D3, is committed on the branch the series is
  built on, as commit `a3800e4`, which becomes the current code in place of
  `11ffbfd` and moves two line numbers of `controllerplugin.go`; the
  implementing agent writes the code on the analysis machine and builds on
  the lab machine ("you can use local repo for editing and build
  remotely"), which changes question 47; the lab machine's checkout is kept
  in step with the tree the code is written in, in place of the clean clone
  of `master` asked for before.
- Review of this document (2026-11-23): one reader, the whole text against
  the README, the current code and `jdss-rest-api-spec` at commit
  `434c9cd`. Found: `go test` runs vet's printf check before the tests of a
  package, so the unit tests that stage 3 gives `pkg/controller` would not
  run with its 23 format strings left in (tried on the lab machine with Go
  1.26.0: a failed build, and a pass with `-vet=off`); the catalogue says
  neither which code a kind of error gets nor which CSI status a new code
  gets; a volume's CSI ID can have 131 bytes, over the limit that question
  68 keeps the token within; the README freezes an accepted record, which
  this document expects to change with the implementation.
- User decision (2026-11-23, the third round that day): the changes
  proposed with that review are applied ("please apply changes"). The 23
  format strings of `pkg/controller` are fixed in the first commit, with
  those of `pkg/rest`, which changes the decision of 2026-11-14 for the
  controller (question 69,
  [impl_plugin_ctrl_no_format_string_from_error_text](#impl_plugin_ctrl_no_format_string_from_error_text));
  the three other findings are open questions 70 to 72.
- User decision (2026-11-23, the fourth round that day, question 70):
  "providing mapping of rest error codes to replies to kubernetes cluster
  is not within scope of current adr". This document implements the
  JovianDSS REST API v4 in the REST module and provides its errors to the
  modules that use it; how those modules handle them is out of scope. The
  rest of the question, which code each kind of error gets, stays open as
  question 73; the caller changes decided before that day are kept.
- User decision (2026-11-23, the fourth round that day, question 71):
  "volume and share names should be less then 131 bytes, csi naming should
  be decided on the basis of csi spec". Names on the appliance keep the
  limit of 131 that question 21 set; a generated volume name, which is also
  the volume's CSI ID, is at most 128 bytes, the CSI spec's limit for a
  string, so both thresholds of `NewVolumeDescFromName` go down from 128 to
  125. The CSI ID of a snapshot is left to the work on snapshot names.
- User decision (2026-11-23, the fifth round that day, question 72): "keep
  proposed". This document stays Proposed while the series is implemented
  and tested, and is accepted with the tested series; what the
  implementation changes in it is written into its text until then.
- User decision (2026-11-23, the sixth round that day, question 73): each
  layer of the appliance's errors gets a range of 1000 codes ("please give
  each layer code multiple of 1000 so that no class errors will be 1000 to
  1999, python errors 2000 to 2999 and so forth"). The layers are those of a
  study of the REST errors that the user brought: routing and the API's own
  checks (`werkzeug.exceptions.*`), the request body checked against a
  schema (`jsonschema`), product logic (`opene.*`), the ZFS wrapper
  (`zfslib.*`), uncaught Python errors (`builtins.*`) and errors without a
  class. The four layers the user did not number take 3000 to 6000 in the
  order of the study's table, upwards. Every class in the scenarios files of
  the plugin's calls on jdss1 v4, at `jdss-rest-api-spec` commit `434c9cd`,
  falls into one of the six. What today's callers compare with is open, as
  question 74.
- User decisions (2026-11-23, the seventh round that day, questions 73 and
  74): the order of the four layers and the codes below 1000 staying the
  plugin's own are confirmed ("your intuition looks good to me"), and an
  error carries two codes, so that today's callers keep working ("lets have
  2 codes to be compatible"): the code of today, and the code of its kind
  in the range of its layer. Who numbers the kinds inside a range is open,
  as question 75.
- User decision (2026-11-23, the eighth round that day, question 75):
  "generate the table for the 96 kinds now". Table 19 numbers the kinds of
  error of the plugin's calls, read from their scenarios files on jdss1 v4
  at `jdss-rest-api-spec` commit `434c9cd`. It holds 95 kinds, not 96: the
  count given with the review took the plain-text replies of those files,
  401 and 404 without an error object, as one more kind, and they have none.
  No question is open.
- User decision (2026-11-23, the ninth round that day, question 70): "keep
  the caller changes in stage 3". With the callers' handling out of scope,
  the changes that earlier decisions made in the driver and the controller
  stay in this document: the zvol `DeleteClone` caller (question 18), the
  cleanup on `volume has children` (question 17), `FailedPrecondition` for a
  volume that has snapshots, the target delete after a LUN is reported as
  not attached, and the check of an existing share (question 57).
- Check on the lab machine (2026-11-23): a plain `go vet ./pkg/rest/...`
  reported nothing on the current code, and `go vet ./...` 26 reports where
  the run of 2026-11-14 had 34. With an empty build cache, and with `-a`,
  the eight reports of `pkg/rest` came back. On an empty cache,
  `go vet ./pkg/driver/` followed by `go vet ./pkg/rest/` reported nothing:
  Go 1.26.0 takes over the result of a package that was vetted as a
  dependency. Whether a later Go release has the fault was not looked for.
- User decisions (2026-11-23, the tenth round that day, questions 76 and
  77): `make vet` runs `go vet -a`; and a catalogue entry gives, as the code
  today's callers read, the code that the call returns today for the reply,
  another one only where this document decides it ("yes, change make vet to
  go vet -a and add the rule").




- User decision (2026-11-16): the audit checks the reply types too, by
  comparing `types.go` with the OpenAPI.
- User decisions (2026-11-16): an error handled as Unknown on purpose is
  logged at Error; the catalogue entries that are not in the spec are not
  tested in this delivery, their test needs a test server built later
  (tracker T13), and each carries a `TODO` in the code until then.
- jdss-rest-api-spec, commit `6149fa9` and a note from its session (read
  2026-11-16): the OpenAPI is tracked beside the TypeSpec, one file per path
  and one per kind of error, rewritten only after a passing check.
- User decision (2026-11-16, questions 31 and 38): transcription by an agent
  stays; the agent reads the tracked OpenAPI instead of the `.tsp` files.
- User decision (2026-11-16): the error catalogue is per call; a failed
  reply is checked only against the errors known for that call, which the
  scenarios files and the TypeSpec list.
- User decision (2026-11-16): an error has a handler in a call when the
  storage can send it for that REST resource, and none when it cannot; the
  replies to input the plugin does not send today are included, because a
  later change of the code could cause them.
- User decision (2026-11-16): every error must be handled; one whose
  meaning is known gets its code, one whose meaning is not known is handled
  as Unknown.
- User decision (2026-11-16): an error known from old code, from testing
  or from another implementation, and missing from the spec, is kept and
  handled, and pointed out to the user to be added to the spec; the two
  patterns no recording supports most likely come from an older JovianDSS
  release.
- jdss-rest-api-spec, the scenarios files of the plugin's calls on jdss1 v4
  (read 2026-11-16): 332 recorded errors, 190 without an `errno`; 333 and
  191 at commit `94d0c4f` (read 2026-11-17).
- User decision (2026-11-16): the condition of the first logging commit is
  checked later, by a unit test that comes with the fixtures in stage 2.
- User decision (2026-11-16): the default log level should be Info. Today
  it is not a default but a value in the shipped controller configuration;
  the change is recorded in the tracker, not made here.
- User decision (2026-11-15, question 37): this document is about the REST
  module, so "logged once" means once in that module; the driver and the
  controller make their own log records, which are left alone at this stage.
- User decision (2026-11-15, question 36): the REST password and a CHAP
  password never reach a log, at Debug, at Info or at any other level; the
  log level is set in the configuration for a run; the
  two logging items move to the front of the series.
- Review of this document (2026-11-15, question 35): csi-sanity does not
  check that a clone survives the delete of its source; the check is a read
  made by hand, noted in the tracker under T11.
- User decision (2026-11-15, question 25): the four crashes in the ID and
  name code go to the issue tracker; the validation item is not extended.
- User decision (2026-11-15, question 22): the preparation of the NFS live
  cycle goes to the issue tracker.
- User decision (2026-11-14, question 33): the regeneration procedure is
  handled separately, as an issue of its own.
- User decision (2026-11-14, question 32): `origin`, `clones` and NAS `quota`
  are optional, and more optional fields may come up in testing, because the
  JovianDSS API is not settled.
- User decisions (2026-11-14): the dead-code commit removes `StorageCfg`
  together with its dead user in `pkg/controller`; certificates are verified
  through a certificate the administrator imports into JovianDSS and `cafile`
  names; the connection tests with a certificate are developed separately.
- User decision (2026-11-14): this document fixes the format strings of
  `pkg/rest` only; the controller's and the node's are fixed later.
- User decision (2026-11-14): the list loop stops on an empty page and on a
  page shorter than `per_page`, and `per_page` is sent on every version. The
  check for a repeated page was set aside: it would have to compare whole
  pages, because a volume removed or added between two requests shifts the
  entries.
- User decision (2026-11-13): the NFS snapshot deletes that destroy clones
  are driver bugs, fixed separately and not in this document; it carries
  them as risk 23.
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

**Table 10 — Evidence sources** · `tbl_evidence_sources`

| Source | What it is | Use |
|---|---|---|
| `jdss-rest-api-spec/empirical/hurl/out/` | every request and reply of the hurl runs on real appliances, jdss1 and jdss2. Not in git | the replies the fixtures are taken from ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) and the values the generated specs mask, because it has the raw values; the reply types are read from the OpenAPI ([question 31](#decisions)). Only jdss1 v4 replies decide pass or fail; jdss2 v4 replies are evidence beside them, and v3 replies background |
| `jdss-rest-api-spec/empirical/{jdss1,jdss2}/` | specs generated from those runs (`INDEX.md`, `-doc`, `-src`), and the TypeSpec of the observed types, validated by `jdss-rest-api-spec`'s `--check` against every recorded reply (`typespec/`, with one error model per kind of error and a scenarios file per method), and the OpenAPI compiled from it (`openapi3/`) | reference for status codes, errors and behaviour; **for jdss1 v4, the source of the reply types and of the error catalogue** ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)). Sizes, guids and creation times in the generated `-doc` replies, and every value of `GET pools/<pool>` but booleans and nulls, are masked, so types are not read from those |
| `jdss-rest-api-spec/empirical/hurl/project-status.md` | the hurl suite's numbered findings (cited as "hurl Finding n") | behaviour, data safety |
| `jdss-rest-api-spec/{v4,v3}/d/` | the official (company) API docs | can differ from the appliance; the source of the request schemas and the documented statuses ([rule_rest_request_body_valid](#rule_rest_request_body_valid), [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed)), otherwise used only where there are no recordings |
| `kub-testing/tmp/cycles/` | csi-sanity runs against the lab appliance | confirms an effect on a live system |

When sources disagree, recorded replies win, then the v4 docs, then the v3 docs.

**How the reply types were checked.** The analysis machine has no Go, so a Python
script imitates what Go's `encoding/json` does with each recorded reply (prototype
`kub-testing/tmp/rest-audit/audit.py`, output `audit.txt`). That is how the
mismatches were found. The types themselves are transcribed from `jdss-rest-api-spec`'s validated description of the replies ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), and [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)'s unit tests check
them in Go. The [error messages](#error-messages) catalogue and the
[evidence gaps](#evidence-gaps) were produced the same way (`coverage.py`,
`bodies.py`, same folder).

**What this check cannot show.** A field that was null in every recording, or a
list that was always empty, could not be checked (for example vdev spares, or a
pool `scan` in progress). Each hurl appliance is one installation, so a reply
shape that only appears with another pool layout or another feature switched on is not in the recordings. The lab appliance and hurl's `jdss1` run the same
JovianDSS 1 release, `1.0up33.65410` (confirmed by the user), and
[bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch) and [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) fail the same way on both. Those
two are the only reply-type findings confirmed on the live system so far (the
"Seen live?" row of each entry under [response types](#response-types)).

### Evidence gaps
[Evidence_gaps](#evidence-gaps)

hurl recorded replies for every request the plugin sends, on jdss1 v4 and on jdss2
v4 (see the [compliance matrix](#compliance-matrix)). Two things are missing: hurl
did not always send a request the exact way the plugin does (an extra body, no
`per_page`, longer names), and some situations were never brought about on purpose (a full pool, a missing pool, an appliance that does not answer). So there are
two kinds of gap: a request the plugin sends differently from the hurl tests, and
a situation that was never recorded. Each gap
below says which kind it is, what the evidence showed before the gap runs, and
what was found by the three hurl runs made for the gaps, by a fourth for out of space and by the paging run of the volume list (hurl Finding 24); [how the gap runs went](#how-the-gap-runs-went)
says what the runs themselves managed to cover, and lists the statuses and error
patterns the code expects that no recording has ever shown.

#### gap_jdss_rest_empty_body_on_get_delete

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | a `{}` body on **every GET**, and on DELETE of target, LUN and share, where hurl sent none (`rproxy.go:108`; the NAS deletes send `{}` as hurl did) |
| hurl sent | GET: never a body. DELETE of those three: never a body |
| Why it matters | works with v3 on the lab appliance today; unknown on v4 and on jdss2 before the gap runs |
| Result of the gap runs | No difference with and without the `{}` body, in 80 (jdss1 v3) and 82 (jdss2 v4) pairs of requests, including DELETE of target, LUN and share. On jdss1 v4, in run `20261107-010938`: no difference either, in 82 pairs of GET requests and 134 pairs of DELETE requests. Closed on all three |

#### gap_jdss_rest_list_paging_without_per_page

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | lists as `?page=N&_dc=<random>`, without `per_page` |
| hurl sent | never `_dc`. `page` without `per_page` only on snapshot lists, with at most 10 items |
| Why it matters | the default page size was unknown before the gap runs. The plugin keeps asking for the next page until it gets an empty one (`driver/common.go:52`). That stop rule was seen working, but never on a long list without `per_page`. |
| Result of the gap runs | On jdss2 v4, the `volumes` and `nas-volumes` lists ignore `page` when `per_page` is missing; on jdss1, v3 and v4, without `per_page`, page 0 held all the test objects and page 1 and page 99 were empty. Page 0 holds the whole list only when the list has 20 entries or fewer, and these lists held at most 10: on jdss1 v4 a volume list request without `per_page` returns pages of 20 (hurl Finding 24, run `20261117-032241`), and the other lists, and v3, are taken to do the same ([question 49](#decisions)). With `per_page`, paging works on all three; `results` is the total, not the page's count. `_dc` changes nothing. Finding: [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page), which concerns jdss2; on JovianDSS 1 the plugin's loop ends |

#### gap_jdss_rest_volume_create_without_sparse

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | `POST volumes` with only `name` and `size` |
| hurl sent | always with `sparse` |
| Why it matters | every CSI volume gets the appliance's default provisioning, and defaults can change between releases: block size did (hurl Finding 9) |
| Result of the gap runs | The default is **thick** (the full size is reserved) on both: the same `refreservation` as with `sparse: false`, about the volume size. Finding: [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse). The same on jdss1 v4 |

#### gap_jdss_rest_name_length_limit

| | |
|---|---|
| Kind | a request the plugin sends differently from the hurl tests |
| The plugin sends | names up to 243 characters (`vp_`/`vb_` + 240) |
| hurl sent | longest name about 59 characters |
| Why it matters | the length limit and its error were unknown before the gap runs |
| Result of the gap runs | Zvol and clone names: 129 characters accepted, 150 refused; the appliance's own message gives the limit, 132. Snapshots: the limit is on the full name `pool/volume@snapshot`, at most 255; with the tested pool and volume names a 216-character snapshot name was accepted and 217 refused ([snapshot name limits](#future-work-snapshot-name-limits)). Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); the plugin caps names at 131 and keeps its generated names within 128 bytes ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)). Zvol and clone names give the same on jdss1 v4 |


#### gap_jdss_rest_out_of_space_messages

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **Out of space:** a volume, thick volume, snapshot or NAS reservation larger than the pool's free space (`available`). The error message was never seen |
| Plugin code that depends on it | `errors.go:270`; `ResourceExhausted` at `controllerplugin.go:378`, `:656` |
| Result of the gap runs | Three different out-of-space messages; only one is recognised. In those first runs the `nospace/` cases of zvol snapshots, clones and thick volumes were not run. Finding: [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised). On jdss1 v4, run `20261116-223132` (the `nospace/` tests): a zvol create that is thick, or sent without `sparse`, and larger than the pool's free space (`available`) is refused by the API's own check with 400, class `BadRequest`, no errno, `New zvol size(...) exceeds available space on pool ...`, the one message today's code recognises; a thin zvol is accepted at any size; a NAS volume with a reservation larger than the pool's free space gets 500, errno 4, `cannot create ...: out of space`. A refused create leaves nothing behind, and a GET or a DELETE of the name then answers 500, errno 1, `Zfs resource: ... not found in this collection`. On jdss1 v4, snapshot and clone creates on a full pool are not recorded |

#### gap_jdss_rest_delete_missing_zvol_clone

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Zvol `DeleteClone` of a clone that does not exist (the NAS case is recorded) |
| Plugin code that depends on it | `iscsi_driver.go:520` |
| Result of the gap runs | `500`, errno 1, `Zfs resource: ... not found`: recognised. Finding: [bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy](#bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy) (the problem is in the caller); low. The same on jdss1 v4 |

#### gap_jdss_rest_clones_of_missing_parent

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | `GET .../clones` on a missing snapshot or volume, zvol and NAS: only 200 recorded |
| Plugin code that depends on it | `iscsi_driver.go:498`, `nfs_driver.go:635` |
| Result of the gap runs | Clones of a missing parent: `500`, errno 1, recognised. A snapshot with no clones: `200` with an empty list, never `204`. Snapshot list of a missing parent: `500` on jdss1 v3, `404` on jdss1 v4 and jdss2 v4. Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); NFS only |

#### gap_jdss_rest_missing_pool_replies

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | The pool is missing or unavailable, on any call except `GET pools/<pool>` |
| Plugin code that depends on it | every call |
| Result of the gap runs | `POST` and list of volumes on a missing pool: `500`, errno 1. Targets: `404`. Finding: [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised); low. The same on jdss1 v4, where the snapshot lists of a missing pool answer `404`, `Pool <p> not found.` An unavailable pool was not run: `pool-unavailable/` needs one made by hand |

#### gap_jdss_rest_detach_and_delete_with_connected_clients

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **Clients connected:** in the hurl tests no initiator logs in and nothing is mounted (empirical README). Detach, target delete and deletes with `force_umount: true` were recorded only on idle resources ([risk 8](#risks)) |
| Plugin code that depends on it | unpublish and every delete |
| Result of the gap runs | not run: the hurl tests log no initiator in. [Testing](#testing) steps 6 and 7 would give the first evidence |

#### gap_plugin_rest_appliance_unreachable

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | **The appliance cannot be reached:** connection refused, DNS, timeout, TLS, reply cut off. No test brings it about, but it happens: jdss1 stopped answering for 15-18 minutes three times (hurl C24), and once for over 8 hours. On this path the plugin crashes ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)) |
| Plugin code that depends on it | `Send` |
| Result of the gap runs | not run: it cannot be recorded reliably; left to unit tests with a fake HTTP server ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) |

#### gap_plugin_rest_unexpected_reply_bodies

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Unexpected reply bodies: HTML or empty 5xx replies, or a list answered with `200` and `data: null`; on the last the plugin crashes ([crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body)) |
| Plugin code that depends on it | `getError`, list functions |
| Result of the gap runs | No `200` with `data: null` on any GET. Plain-text `404 not found` and `401 unauthorized` replies exist; they cause a parse error, not a crash. The crash on `null` is still in the code, but no recorded reply reaches it. Finding: [crash_plugin_rest_error_from_nil_body](#crash_plugin_rest_error_from_nil_body), [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status); unit test only |

#### gap_plugin_rest_failover_during_call

| | |
|---|---|
| Kind | a situation that was never recorded |
| Situation | Failover: the pool moves to the other node during a call; several addresses in `addrs` ([bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover](#bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover)) |
| Plugin code that depends on it | `Send` |
| Result of the gap runs | not run: no test moves the pool to the other node during a call; not testable in this delivery ([Testing](#testing) limit 10) |

#### What the evidence covers but the plugin ignores
[What_the_evidence_covers_but_the_plugin_ignores](#what-the-evidence-covers-but-the-plugin-ignores)

On jdss2, a snapshot list is briefly out of date after a snapshot is created or
deleted: for under 0.5 s (hurl Finding 20; for jdss1, see [risk 14](#risks)). So a ListSnapshots straight after a
CreateSnapshot may not show the new snapshot. Reading a snapshot by its name works
at once. This document proposes no change for it; what later driver
work would have to allow for is in [risk 14](#risks).

#### How the gap runs went
[How_the_gap_runs_went](#how-the-gap-runs-went)

Three runs were made to fill the gaps: jdss1 `20261105-221223` on v3, jdss2
`20261105-220744` on v4, and jdss1 `20261107-010938` on v4. The third replaces
the v4 part of the first, which was lost: jdss1 stopped answering 2 minutes
after it started (the 8-hour outage in [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable)).
In the first two runs one test, which
puts a newline into a share comment, broke share handling and NAS deletes on both
appliances for the rest of the run (hurl Finding 21 for jdss2, 22 for jdss1),
and on jdss1 the pool also filled up. So their results come only from the
part of each run before that test, or from zvol and iSCSI tests, which the breakage did
not touch. The third run had no such breakage: about 14,100 requests, one
without an answer. It left out the `race/`, `scale/` and `nospace/` tests.
The `scale/` tests ran on no appliance. The `nospace/` tests ran
later, on jdss1 v4, in run `20261116-223132`.

Never seen in any of the three runs (about 21,000 requests in each of the
first two): `202`, `204` on a GET, `200`
on a DELETE. Also never seen, not even in the `race/` tests:
`Resource ... already exists.` with errno 1 or 100, and
`Zfs resource: ... not found in this collection` without an errno, the one
"not found" pattern the code checks on the message alone when the errno is missing (`errors.go:261`) (it was
recorded with errno 1 only; other not-found replies without an errno, such as
404 `Resource <pool>/<volume> not found.`, are recorded). Creating the same object twice at the same time gives errno 100
`dataset already exists` (recognised) or errno 5.

#### How to close the gaps
[How_to_close_the_gaps](#how-to-close-the-gaps)

`jdss-rest-api-spec` is its own project, so this document does not change it. The user adds
tests for the first eight gaps above ([gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete), [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page), [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse), [gap_jdss_rest_name_length_limit](#gap_jdss_rest_name_length_limit), [gap_jdss_rest_out_of_space_messages](#gap_jdss_rest_out_of_space_messages), [gap_jdss_rest_delete_missing_zvol_clone](#gap_jdss_rest_delete_missing_zvol_clone), [gap_jdss_rest_clones_of_missing_parent](#gap_jdss_rest_clones_of_missing_parent) and [gap_jdss_rest_missing_pool_replies](#gap_jdss_rest_missing_pool_replies)) to the hurl suite; the results so far are
in their entries. [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) and [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call) need a connected client or a failover during a call. csi-sanity's node preset
would cover part of [gap_jdss_rest_detach_and_delete_with_connected_clients](#gap_jdss_rest_detach_and_delete_with_connected_clients) on the lab appliance, but only if the iSCSI NodeStage really
logs in, which review findings `bug_plugin_node_iscsi_login_failure_ignored` and `bug_plugin_node_unstage_no_logout` put in doubt; [Testing](#testing) steps 6 and
7 check this. [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) and [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies) cannot be recorded
reliably, so they are left to unit tests with a fake HTTP server ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).

### The `jdss-rest-api-spec` repository
[The_jdss_rest_api_spec_repository](#the-jdss-rest-api-spec-repository)

`jdss-rest-api-spec` (a checkout beside the plugin repository, `../jdss-rest-api-spec`;
not on the lab machine) is what this document measures the plugin against. It describes the
JovianDSS REST API twice, with the same paths in both: once as Open-E documents it
(the company spec), and once as real appliances behave (the empirical spec). It was
written for the JovianDSS-Proxmox and CSI plugin teams.

**Table 11 — What `jdss-rest-api-spec` holds** · `tbl_spec_forms`

| Part | Where | What it says | Role here |
|---|---|---|---|
| **Company spec** | `v4/d/<path>/<METHOD>-doc` (description), `-src` (JSON Schema of the request body), `-desc`; also `v4/typespec/`, `v4/openapi3/`. `v3/` is laid out the same way, without an OpenAPI | what Open-E documents, taken from the Swagger HTML shipped with the appliance (`v4/origin/`) | which request keys exist, documented status codes; fills gaps |
| **Empirical spec** | `empirical/<appliance>/v4/`: `INDEX.md` (every tested endpoint, with documented and observed status codes), `STATUS-CODES.md`, `SRC-COMPARISON.md`, `ERROR-FIELDS.md`, `d/<path>/<METHOD>-doc`/`-src` (as observed), `typespec/` (namespace `Observed`: the observed types per method, `models/errors/` with one model per kind of error, and a `-scenarios.json` per method, the source of the error catalogue, [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)), `openapi3/` (the OpenAPI compiled from it, one file per path and one per kind of error; the source of the reply types and of each kind of error's schema, [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)); `empirical/<appliance>/status-codes.json` | what real appliances did, generated from the hurl runs | **decides where the two specs disagree** |
| Raw responses | `empirical/hurl/out/` (local only, not in git) | every request and reply of every run | the fixtures, and the values the generated specs mask: sizes, guids and creation times in every reply, and every value but booleans and nulls of `GET pools/<pool>` |
| Findings | `empirical/hurl/project-status.md`, `bug-reports/` | numbered findings, runs, repeatability | behaviour, data safety |

The appliances, as `jdss-rest-api-spec`'s README describes them:

**Table 12 — Appliances in `jdss-rest-api-spec`** · `tbl_spec_appliances`

| Name | Release | Auth | APIs |
|---|---|---|---|
| `jdss1` | JovianDSS 1 **`1.0up33.65410`**, two-node high-availability cluster | Basic | v3, v4 |
| `jdss2` | JovianDSS 2 `2.0up01.67856` | JWT Bearer token, from `POST api/v4/auth/login`, valid 900 s | v4 only |

**Rules for this work.** `jdss-rest-api-spec` belongs to its own project. This document and
the plugin work only read it: they never change its code or data, and never run its
scripts or tests (the tests create and delete objects on real appliances). When new
recordings arrive, the user runs its `build_typespec.py --check`, which rewrites, once it passes, the tracked OpenAPI that the
reply types ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)) are read from. Problems
found in it are reported to the user. `jdss-rest-api-spec` covers the requests the Proxmox and CSI plugins make, the test types of its `test-types.md` and endpoints found along the way: an endpoint missing from it was simply not tested, which does
not mean it is broken. It aims to hold every reply and every error of those
requests, but its recordings are incomplete and still growing: an error found in
testing, in old code or in another implementation and missing from it is not dropped: it is reported to the user so that it can be added
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)).

### What "compliant" means
[What_compliant_means](#what-compliant-means)

The plugin is compliant when every REST call it makes meets the rules below on
jdss1 v4. The jdss2 v4 recordings are evidence beside those of jdss1 v4: a difference there is
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
| Checked against | `INDEX.md`, `STATUS-CODES.md`, the `-scenarios.json` file of each call, recorded error bodies |

#### rule_rest_reply_types_decode

| | |
|---|---|
| Requirement | Reply types: every Go type the plugin decodes accepts every recorded v4 reply of that call |
| Checked against | `jdss-rest-api-spec`'s OpenAPI of the observed types, validated by its `--check` against the raw replies in `hurl/out/`; the Go types are transcribed from it ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), and [impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool) compares them with it field by field, and [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures) decodes the fixtures with them |

#### rule_rest_error_bodies_handled

| | |
|---|---|
| Requirement | Error bodies: for each call, every error recorded for that call that comes without an `errno` or a `class` is mapped like the others, from the fields it has |
| Checked against | `ERROR-FIELDS.md`, and the fields each `-scenarios.json` file marks as not sent |

#### rule_rest_docs_vs_appliance_listed

| | |
|---|---|
| Requirement | Docs and reality: where the company spec and the appliance differ, the plugin follows the appliance, and the difference is listed here so it can be reported to the `jdss-rest-api-spec` project; the empirical findings, not the documentation, are the source of truth (user decision) |
| Checked against | company `-doc` against empirical `-doc`, by reading |

#### rule_rest_lists_paged_as_served

| | |
|---|---|
| Requirement | Lists are read as the appliance pages them: every list request carries `per_page`, a read starts at page 0 and ends on an empty page or one shorter than `per_page`, and nothing the plugin keeps between two CSI calls names an appliance page, which holds other entries once the list changes |
| Checked against | hurl Finding 24 for the volume list on jdss1 v4, and the paging runs of the other lists if they come ([question 49](#decisions), [risk 28](#risks)); the request log of the live cycle ([Testing](#testing) step 4) |

The rules but [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed), which has no pass or fail, are checked again after each new hurl run that the `jdss-rest-api-spec` project
adds to its generated specs: [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4), [rule_rest_request_body_valid](#rule_rest_request_body_valid), [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted), [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) and [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) by the audit ([impl_lab_rest_audit_tool](#impl_lab_rest_audit_tool)), [rule_rest_reply_types_decode](#rule_rest_reply_types_decode) by
the audit's comparison of the types with the OpenAPI, by reading the types again ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)) and running the unit tests ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)).
[rule_rest_lists_paged_as_served](#rule_rest_lists_paged_as_served) is checked against each paging run of the hurl suite, which
its generated specs leave out, and against the request log of the live
cycle.

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

One row per REST call. "v4 doc" is the statuses the company spec lists; "jdss1 v4"
and "jdss2 v4" are the statuses the appliances returned (from both appliances'
`INDEX.md`; jdss2 v4 as evidence only; prototype `kub-testing/tmp/rest-audit/compliance_matrix.py`); "Plugin
success" is the statuses the code treats as success. All 34 call functions of the current code that send a request, three of them dead, appear in both `INDEX.md` files, and so do the three paths new in stage 3,
the iSCSI per-volume snapshot list
([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub)), the
NFS one at its corrected path
([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) and the share read of
[question 57](#decisions), each with its own row. "Note" links the findings that affect
the call; findings that affect every call, such as [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status), are not
repeated.

**Table 13 — Compliance matrix per call** · `tbl_compliance_matrix`

| Plugin call | v4 method and path | v4 doc | jdss1 v4 | jdss2 v4 | Plugin success | Note |
|---|---|---|---|---|---|---|
| `GetPool` | `GET pools/<x>` | 200 | 200, 401, 404 | 200, 401, 404 | 200, 204 | [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `GetVolume` | `GET pools/<x>/volumes/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 204 | [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateVolume` | `POST pools/<x>/volumes` | 201 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised) (on jdss1 v4 the recorded out-of-space reply of this call is the one the code recognises), [bug_plugin_rest_create_volume_never_sets_sparse](#bug_plugin_rest_create_volume_never_sets_sparse) |
| `DeleteVolume` | `DELETE pools/<x>/volumes/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 200, 204 |  |
| `GetVolumesEntries` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page) |
| `ListVolumes` | `GET pools/<x>/volumes` | 200 | 200, 500 | 200, 500 | 200 | dead, for the dead-code commit |
| `GetVolumeSnapshot` | `GET pools/<x>/volumes/<x>/snapshots/<x>` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 202 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key) |
| `CreateSnapshot` | `POST pools/<x>/volumes/<x>/snapshots` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap) |
| `DeleteSnapshot` | `DELETE pools/<x>/volumes/<x>/snapshots/<x>` | 204 | 204, 400, 404, 500 | 204, 400, 404, 500 | 204 |  |
| `GetSnapshotsEntries` | `GET pools/<x>/volumes/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead](#bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead), [bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors](#bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors) |
| `GetVolumeSnapshotsEntries` (new call in stage 3) | `GET pools/<x>/volumes/<x>/snapshots` | 200 | 200, 404, 500 | 200, 404, 500 | none today (the function is a stub); 200 once stage 3 implements it | [bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name); fixtures from `list/volume-kinds.hurl` ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) |
| `GetVolumeSnapshotClones` | `GET pools/<x>/volumes/<x>/snapshots/<x>/clones` | 200 | 200, 500 | 200, 500 | 200, 204 | [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch), [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateClone` | `POST pools/<x>/volumes/<x>/clone` | 200 | 200, 400, 500 | 200, 400, 500 | 200, 201 | [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap) |
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
| `GetNASVolumeSnapshotsEntries` (path corrected in stage 3) | `GET pools/<x>/nas-volumes/<x>/snapshots` | 200, 500 | 200, 404, 500 | 200, 404, 500 | none today; 200, 201 once stage 3 corrects the path ([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) | [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name); fixtures from `list/volume-kinds.hurl` ([impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures)) |
| `GetNASSnapshotsEntries` | `GET pools/<x>/nas-volumes/snapshots` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 201 | [bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key) |
| `CreateNASSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap) |
| `DeleteNASSnapshot` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>` | 204, 400, 500 | 204, 404, 500 | 204, 404, 500 | 204 |  |
| `DeleteNASVolumeSnapshot` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>` | **method missing** | 405 | 405 | 204 | **[bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete): wrong method**, [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check); for the dead-code commit |
| `CreateNASVolumeSnapshot` | `POST pools/<x>/nas-volumes` | 201, 400, 500 | 201, 400, 500 | 201, 400, 500 | 200, 201 | dead (`bug_plugin_rest_nas_snapshot_create_copy_of_volume_create`), for the dead-code commit |
| `GetNASVolumeSnapshot` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>` | 200, 500 | 200, 404, 500 | 200, 404, 500 | 200, 202 |  |
| `GetNASVolumeSnapshotClones` | `GET pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 500 | 200, 500 | 200, 500 | 200, 204 | [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) |
| `CreateNASClone` | `POST pools/<x>/nas-volumes/<x>/snapshots/<x>/clones` | 200, 400, 500 | 200, 400, 500 | 200, 400, 500 | 200, 201 | [bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap](#bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap) |
| `DeleteNASClone` | `DELETE pools/<x>/nas-volumes/<x>/snapshots/<x>/clones/<x>` | 204, 400, 404, 500 | 204, 404, 500 | 204, 404, 500 | 204 | [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised) |
| `CreateShare` | `POST shares` | 201 | 201, 400, 409, 500 | 201, 400, 409, 500 | 200, 201 | [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) |
| `GetShare` (new in stage 3) | `GET shares/<x>` | 200 | 200, 404 | 200, 404, 500 | none today; 200 once stage 3 adds it | the existing-share check of [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised) ([question 57](#decisions)) |
| `DeleteShare` | `DELETE shares/<x>` | 204 | 204, 404 | 204, 404, 500 | 204 | dead (`bug_plugin_rest_delete_share_nil_descriptor`), for the dead-code commit; [bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class) |

Where the plugin stands on each rule today:

**Table 14 — Compliance status per rule** · `tbl_compliance_status`

| Rule | Status in the current code | Assigned to |
|---|---|---|
| [rule_rest_call_exists_on_v4](#rule_rest_call_exists_on_v4) | **33 of 34 pass.** Fails: [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete). [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots) passes on paper, but calls the wrong path | dead-code commit ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)), stage 3 ([bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) |
| [rule_rest_request_body_valid](#rule_rest_request_body_valid) | **Passes for every key the driver really sets.** Three optional keys in the Go request structs are not in the company schema, but they are never set ([dead code](#dead-code)). `properties` is an open object in the schema, so any sub-key is allowed. The `{}` body sent on GET and DELETE is outside every schema; [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) found it has no effect. Prototype: `kub-testing/tmp/rest-audit/compliance_bodies.py` | dead-code commit |
| [rule_rest_success_statuses_accepted](#rule_rest_success_statuses_accepted) | **Every 2xx status returned is accepted.** The code also accepts statuses that were never returned: [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check), [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay), 202 on both snapshot GETs, 200 on `DeleteVolume`, 201 on list GETs, 200 on creates that answer 201, and 201 on the creates that answer 200 (`CreateSnapshot`, `CreateClone`, `CreateNASClone`) | stage 3 ([bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check), [bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay](#bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay)); the rest are harmless and would stay |
| [rule_rest_error_statuses_handled](#rule_rest_error_statuses_handled) | **Fails.** [bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status); `401` on `GetPool` is not mapped; [bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name](#bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name), [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised), [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone), [bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised), [bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised), [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised), [bug_rest_v4_create_calls_out_of_space_messages_unrecognised](#bug_rest_v4_create_calls_out_of_space_messages_unrecognised), [bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised](#bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised) | stage 3 (with [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) |
| [rule_rest_reply_types_decode](#rule_rest_reply_types_decode) | **Fails.** [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch), [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch), [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch), [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch), [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch). ([bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec](#bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec) decodes without an error, so it does not fail [rule_rest_reply_types_decode](#rule_rest_reply_types_decode)) | stage 2 ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), transcribed types) |
| [rule_rest_error_bodies_handled](#rule_rest_error_bodies_handled) | **Fails.** The scenarios files of the plugin's calls on jdss1 v4 hold many errors that come without an `errno` and, for `CreateShare`, many that come without a `class`. The code checks such an error against a few patterns only, so most become Unknown: among them 404 `Pool <p> not found.` on both all-snapshots lists, and `Share <n> already exists.` ([bug_rest_v4_shares_post_409_already_exists_unrecognised](#bug_rest_v4_shares_post_409_already_exists_unrecognised)). Many are replies to input the plugin never sends, such as a wrong JSON type. On jdss2, outside this document, [bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class) is such an error too, noted, not a failure; only `DeleteShare` runs into it, and `DeleteShare` is dead code | stage 3 (with [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)); the jdss2 case by the dead-code commit |
| [rule_rest_docs_vs_appliance_listed](#rule_rest_docs_vs_appliance_listed) | Differences found so far: `AttachVolumeToTarget` is documented as 200, the appliance returns 201; the [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) fields are documented as strings; [bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_get_encryption_type_mismatch)'s v4 example shows `"off"`; for most zvol, target and share calls, and for `GetPool`, only the success status is documented, while the appliance also returns 400, 401, 404, 409 or 500; on six NAS calls a returned 404 is not documented; the `POST` used by [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) is not documented and returns 405; target names are documented as up to 255 characters, but 225 is refused (220 accepted); target names are documented as lower case, but upper case is accepted; the name rules for zvols, NAS datasets and snapshots are not documented at all (the company schemas say only `"type": "string"`). The plugin already follows the appliance, except where a finding says otherwise | a list for the `jdss-rest-api-spec` project ([Relationship](#relationship-to-other-designs)) |
| [rule_rest_lists_paged_as_served](#rule_rest_lists_paged_as_served) | **Fails.** No list request carries `per_page` ([bug_plugin_rest_list_calls_send_page_without_per_page](#bug_plugin_rest_list_calls_send_page_without_per_page)), and the CSI tokens name an appliance page, which a resumed call reads again ([bug_plugin_driver_list_resume_drops_page_missing_token_entry](#bug_plugin_driver_list_resume_drops_page_missing_token_entry)) | the [`per_page` commit](#before-stage-1-send-per_page-on-every-list); stage 3 ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)) |

### Response types
[Response_types](#response-types)

Fields whose JSON type or key in the appliance's reply differs from the Go
struct. Any one of them but the last makes the whole reply fail to decode. This catalogue and the two that
follow, [response content and defaults](#response-content-and-defaults) and
[error messages](#error-messages), hold most of the places where the code expects something other than what the appliance does; stages 2 and 3 take most of their entries, and each entry's Assigned to row names where it goes. Unless an
entry says otherwise, a mismatch looks the same on jdss1 v3, jdss1 v4 and jdss2 v4.

No **type** mismatch was found in pool `available`/`size`, snapshots read by name,
NAS volumes, NAS snapshots, targets, or error bodies. The other `Encryption` fields
(`types.go:200`, `:560`) are fine in the recordings. `encoding/json` reports only the
first type error in a reply, so a live test shows only the first failure and hides
the rest; that is why all of them are closed together, by the transcribed
types ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)), and checked with [impl_plugin_rest_recorded_reply_fixtures](#impl_plugin_rest_recorded_reply_fixtures), not one at a time. For [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) and [bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch](#bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_clones_get_is_clone_type_mismatch) **the docs are right and the code is
wrong**; for [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) and [bug_rest_v4_pools_pool_get_disk_led_type_mismatch](#bug_rest_v4_pools_pool_get_disk_led_type_mismatch) the docs are wrong, and the code follows them.

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
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

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
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

#### bug_rest_v4_pools_pool_get_disk_led_type_mismatch

| | |
|---|---|
| Type and field | `Disk.Led` ([`types.go:376`](../../pkg/rest/types.go#L376)) |
| Struct says | `string` |
| Appliance sends | `{"locate": false, "fault": false}` |
| v3 docs | `"off"` |
| v4 docs | `"off"` |
| Plugin reads it? | no |
| Effect | GetCapacity → `Internal`; hidden behind [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) |
| Seen live? | present in the replies; never reported, because [bug_rest_v4_pools_pool_get_iostats_type_mismatch](#bug_rest_v4_pools_pool_get_iostats_type_mismatch) fails first |
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

#### bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch

| | |
|---|---|
| Type and field | `ResourceSnapshotShortProperties.Creation` ([`types.go:273`](../../pkg/rest/types.go#L273)) |
| Struct says | `string` in its decoder, parsed to a number and stored as a time |
| Appliance sends | number |
| v3 docs | number |
| v4 docs | number |
| Plugin reads it? | **yes** |
| Effect | iSCSI ListSnapshots → `Internal` as soon as the pool has a snapshot |
| Seen live? | not yet: the ctrl-B ListSnapshots specs fail earlier, because the controller has already crashed (`connection refused`) |
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

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
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec

| | |
|---|---|
| Type and field | `ResourceNASVolume`, most of its tags (`types.go:447-476`) |
| Struct says | keys of its own, mostly camelCase: `fullName`, `logBias`, `primaryCache`, `secondaryCache`, `accessTime`, `deduplication` |
| Appliance sends | lower-case keys in every recorded NAS volume GET: `full_name`, `logbias`, `primarycache`, `secondarycache`, `atime`, `dedup`. When set, `quota`, `reservation`, `refquota`, `refreservation`, `recordsize` and `copies` are strings, as the custom decoder expects |
| Plugin reads it? | only `name`, `quota` and `origin`, whose keys match; the custom decoder also parses `copies` and `recordsize`, which nothing reads |
| Effect | none today: a type check only sees keys that are present, so the other fields stay empty without an error, and nothing reads them. So the plugin can read only the fields whose keys match. With no quota set, `quota` is `null`, which leaves the size at 0 (review `bug_plugin_driver_nas_quota_null_size_zero`, a driver issue) |
| Seen live? | nothing to see: no failure |
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

### Response content and defaults
[Response_content_and_defaults](#response-content-and-defaults)

Replies that decode, but whose content the code reads wrongly, and two request parameters the plugin never sends, `per_page` and `sparse`, so that the appliance's default applies.

#### bug_rest_v4_pools_pool_volumes_snapshots_get_no_volume_key_parent_name_instead

| | |
|---|---|
| Where | `ResourceSnapshotShort.Volume` ([`types.go:294`](../../pkg/rest/types.go#L294)) |
| What happens | there is no such key: the appliance sends `parent.name`. `Volume` stays empty, and the lister skips the entry (`controller/listers.go:54`). **Once [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch) is fixed, iSCSI ListSnapshots would return an empty list, without an error** |
| Evidence | recorded replies, code reading |
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2 |

#### bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key

| | |
|---|---|
| Where | `ResourceNASSnapshotParent.UnmarshalJSON` ([`types.go:658`](../../pkg/rest/types.go#L658)) |
| What happens | reads `type` under the tag `copies`, so `Parent.Type` is always 0. Unused today, but both all-snapshot lists return zvol **and** NAS snapshots, and `parent.type` (4 zvol, 1 NAS) is the only way to tell them apart (hurl Finding 14); the filter by kind of stage 3 reads it and drops the other kind's entries ([question 7](#decisions)) |
| Evidence | code reading |
| Assigned to | the key: [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2; the filter by kind: [Stage 3](#stage-3-behaviour-and-error-mapping) ([question 7](#decisions)) |

#### bug_plugin_rest_list_calls_send_page_without_per_page

| | |
|---|---|
| Where | `pagedcSuffix` ([`entrielisters.go:34`](../../pkg/rest/entrielisters.go#L34)), `getResourcesList` (`driver/common.go:52`) |
| What happens | lists are requested with `page` and `_dc`, but without `per_page`. On jdss2 v4, the `volumes` and `nas-volumes` lists then ignore `page` and return the first page again (tried with `page=1` and `page=99`). The loop stops only on an empty page, so on jdss2 CSI ListVolumes would never end when `max_entries` is not set, and would return duplicates when it is. On jdss1, v3 and v4, `page` still works and a page past the end is empty, so there the loop ends; a volume list request on jdss1 v4 without `per_page` returns pages of 20 (hurl Finding 24), and the other lists, and v3, are taken to do the same ([question 49](#decisions)) |
| Evidence | recorded replies (results [gap_jdss_rest_list_paging_without_per_page](#gap_jdss_rest_list_paging_without_per_page)); the loop from code reading |
| Assigned to | the [`per_page` commit](#before-stage-1-send-per_page-on-every-list) before stage 1 |

#### bug_plugin_rest_create_volume_never_sets_sparse

| | |
|---|---|
| Where | `CreateVolumeDescriptor.Sparse` ([`reqtypes.go:115`](../../pkg/rest/reqtypes.go#L115)) |
| What happens | never set, so every CSI volume gets the appliance default, which is **thick** on both appliances: the full size is reserved |
| Evidence | recorded replies (results [gap_jdss_rest_volume_create_without_sparse](#gap_jdss_rest_volume_create_without_sparse)) |
| Assigned to | not this document: its own design ([question 9](#decisions), [Relationship](#relationship-to-other-designs)) |

#### bug_rest_v4_pools_pool_volumes_volume_snapshots_snapshot_get_no_name_key

| | |
|---|---|
| Where | `GetVolumeSnapshot` ([`storageapi.go:191`](../../pkg/rest/storageapi.go#L191)) |
| What happens | a zvol snapshot GET has **no `name` key** (NAS snapshot GETs do), so `ResourceSnapshot.Name` and the iSCSI `GeneralSnapshot.Name` (`iscsi_driver.go:454`) are always empty. No caller reads `GeneralSnapshot.Name`; callers take the name from the snapshot descriptor. Low |
| Evidence | recorded replies |
| Assigned to | [impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec), stage 2: the transcribed type has no `name`, and `GetSnapshot` takes the name it asked for ([question 42](#decisions)) |

#### bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name

| | |
|---|---|
| Where | the per-volume snapshot lists: `ListVolumeSnapshots` (`iscsi_driver.go:385-393`, `nfs_driver.go:523-530`), `GetNASVolumeSnapshotsEntries` (`entrielisters.go:217`), and what reads their entries: `cleanIntermediateSnapshots` (`iscsi_driver.go:212`, `nfs_driver.go:316`), `deleteLUN`, on the entries the cleanup returns (`iscsi_driver.go:282`, `nfs_driver.go:421`), and `completeListResponseFromVolumeSnapshots` (`controller/listers.go:124`, `:148`) |
| What happens | the appliance describes a snapshot in two shapes. Read by name it is the full object: `creation` as a date string, `volsize` for a zvol snapshot, and `clones` when it has any. In a list it is a short entry: `name`, `parent` with `name` and `type`, `properties` with `creation` as a number. On v4 the lists of one volume's snapshots, zvol and NAS, send the short entry, as the two all-snapshots lists do; on v3 they send the full object (hurl caveat C13), so the fault comes with the move to v4. The code decodes them as the full object, which gives no error, a `creation` of year 1 and no `clones`. Hidden today, because both lists return nothing ([bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub](#bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub), [bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots](#bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots)) |
| Evidence | recorded replies of both per-volume lists on jdss1 v4 (`list/volume-kinds.hurl`); code reading |
| Assigned to | [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow), stage 3 |

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
| Plugin expects | no `errno`, the class `werkzeug.exceptions.NotFound` and `Share .* not exists.` (`errors.go:251-253`) |
| Effect | not recognised on jdss2 (recognised on jdss1) |
| Assigned to | no fix of its own: [impl_plugin_rest_remove_dead_code_before_v4_move](#impl_plugin_rest_remove_dead_code_before_v4_move), the dead-code commit before stage 1, removes the only call that meets it |

#### bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised

| | |
|---|---|
| Call | CreateSnapshot |
| Recorded reply | 500, `ZfsResourceError`, errno 5, `Snapshot <p>/<v>@<s> already exists.`, in the cases `duplicate` of `crude/volume-snapshot.hurl` and `crude/volume-snapshot-thick.hurl` and `snapshot-as-snapshot-create` of `crude/resource-name-taken.hurl`. The NAS snapshot create gets the same kind of error: [bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised) |
| Plugin expects | `Resource .*/.* already exists.` |
| Effect | a repeated create is not seen as "already exists", so the path for repeated requests (`controllerplugin.go:654`, `iscsi_driver.go:55`, `:165`) is not taken |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_post_already_exists_unrecognised

| | |
|---|---|
| Call | CreateNASSnapshot |
| Recorded reply | 500, `ZfsResourceError`, errno 5, `Snapshot <p>/<ds>@<s> already exists.`, in the cases `duplicate` of `crude/nas-snapshot.hurl` and `dataset-snapshot-as-snapshot-create` of `crude/resource-name-taken.hurl`: the kind of error of the zvol snapshot create, [bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised](#bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised) |
| Plugin expects | `Resource .*/.* already exists.` |
| Effect | a repeated NAS snapshot create is not seen as "already exists", so the path for repeated requests (`controllerplugin.go:654`, `nfs_driver.go:66`, `:254`) is not taken; once it is recognised, `CreateVolumeFromVolume` reaches `deleteIntermediateSnapshot` on it (`nfs_driver.go:254-255`, [risk 23](#risks)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone

| | |
|---|---|
| Call | DettachVolumeFromTarget |
| Recorded reply | 404, `Volume name <v> is not attached to target <t>` |
| Plugin expects | any 404 means "target gone" (`targets.go:182`) |
| Effect | ControllerUnpublish on a target that exists but lacks this LUN returns early (`iscsi_driver.go:667`), so the **target is left behind** |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised

| | |
|---|---|
| Call | DeleteNASVolume |
| Recorded reply | 500, `In order to delete a dataset, you must delete all of its clones first.` |
| Plugin expects | the pattern says `zvol` |
| Effect | an NFS delete of a volume with clones is not seen as `Busy` (`nfs_driver.go:383`, `:396`), so the clone handling is skipped |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_shares_post_409_already_exists_unrecognised

| | |
|---|---|
| Call | CreateShare |
| Recorded reply | 409, errno 1, `Share <n> already exists.` |
| Plugin expects | nothing |
| Effect | on an NFS CreateVolume retry, the existing NAS volume is accepted (`nfs_driver.go:113`), but then the share create fails. No retry can succeed |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3, with the existing-share check ([question 57](#decisions)) |

#### bug_rest_v4_pools_pool_nas_volumes_dataset_snapshots_snapshot_clones_clone_delete_404_not_exists_unrecognised

| | |
|---|---|
| Call | DeleteNASClone |
| Recorded reply | 404, `Clone <n> not exists` |
| Plugin expects | nothing |
| Effect | deleting a clone that is already gone is an error, not success |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_create_calls_out_of_space_messages_unrecognised

| | |
|---|---|
| Call | create on a full pool |
| Recorded reply | 500, errno 4, `cannot create '...': out of space` (zvol, NAS); 500, errno 1000, `cannot create snapshots : out of space` (NAS). Only `400 New zvol size(...) exceeds available space` is recognised; on jdss1 v4 that is what a zvol create without `sparse` gets on a pool that is too small, so a CSI CreateVolume of a zvol is already answered `ResourceExhausted` there |
| Plugin expects | the out-of-space pattern, checked only for errors **without** an errno (`errors.go:270`) |
| Effect | Unknown, so `Internal` instead of `ResourceExhausted`, for every out-of-space reply but the recognised one. On jdss1 v4 that is the NAS volume create; the zvol create gets the recognised message there (results [gap_jdss_rest_out_of_space_messages](#gap_jdss_rest_out_of_space_messages)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3; the zvol snapshot and clone creates, whose reply no run has recorded, are left to another record ([question 59](#decisions); kub-testing `ISSUES.md`, F30) |

#### bug_rest_v4_missing_pool_missing_parent_and_name_too_long_errors_unrecognised

| | |
|---|---|
| Call | several |
| Recorded reply | missing pool: 500, errno 1, `ZpoolError Given zpool '<p>' doesn't exists.`, and for targets 404 `Target partition: <p> not exists.`; name too long: 400, errno 1, `ZvolValidationError Zvol name too long ... 132` ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids) keeps generated names within 128 bytes); v4 snapshot list of a missing parent: 404 `ItemNotFoundError Resource <p>/<v> not found.` |
| Plugin expects | nothing |
| Effect | Unknown (results [gap_jdss_rest_name_length_limit](#gap_jdss_rest_name_length_limit), [gap_jdss_rest_clones_of_missing_parent](#gap_jdss_rest_clones_of_missing_parent), [gap_jdss_rest_missing_pool_replies](#gap_jdss_rest_missing_pool_replies)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_rest_v4_snapshot_and_clone_creates_missing_pool_and_missing_parent_errors_overlap

| | |
|---|---|
| Call | CreateSnapshot, CreateClone and CreateNASClone; for CreateNASSnapshot only the missing pool is recorded |
| Recorded reply | all 500, `ZfsResourceError`, errno 1, `Zfs resource: ... not found in this collection.` On the zvol snapshot create a missing pool gives `<p>/<p>` and a missing volume `<p>/<v>`, one kind of error with one pattern, `^Zfs resource: [\s\S]*/[\s\S]* not found in this collection\.$`. On the clone creates a missing pool gives `<p>/<p>` (zvol) or `<p>@<p>` (NAS) and a missing snapshot `<p>/<v>@<s>`, two kinds whose patterns overlap: the missing snapshot's message matches the missing pool's pattern too. On the NAS snapshot create a missing pool gives `<p>/<p>`. The message templates of the scenarios files keep the cases apart, `{{path.poolname}}/{{path.poolname}}` against `{{path.poolname}}/{{path.volumename}}` (the jdss1 v4 scenarios files of the four calls, `jdss-rest-api-spec` commit `94d0c4f`) |
| Plugin expects | a message with a `/` in the name matches `Zfs resource: (.+\/.+) not found` and is "not found" (`errors.go:81`, `:167-177`), whatever is missing; stage 3 gives a missing pool its own code on the other calls, not on these ([question 48](#decisions)) |
| Effect | on the zvol snapshot create no entry matched on status, `class`, `errno` and pattern could give a missing pool its own code, so a missing pool would be reported as a missing volume; on the clone creates the code would depend on which entry is tried first: with the missing pool's entry first, a missing snapshot would be reported as a missing pool ([risk 25](#risks)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3: not told apart for now; an error that can be either is taken as the parent missing, and its entry carries a `TODO` saying so ([question 48](#decisions)) |

### Logic bugs
[Logic_bugs](#logic-bugs)

Bugs in the REST code itself, whatever the appliance does.

#### bug_plugin_rest_volume_snapshots_lister_is_an_empty_stub

| | |
|---|---|
| Where | `GetVolumeSnapshotsEntries` ([`entrielisters.go:53`](../../pkg/rest/entrielisters.go#L53)) |
| What happens | an empty stub that sends no request. ListSnapshots by source volume on iSCSI always returns nothing. What the request would return, once sent, is misread too: [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name) |
| Evidence | code reading |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping): the request is implemented; its callers by [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow) |

#### bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots

| | |
|---|---|
| Where | `GetNASVolumeSnapshotsEntries` ([`entrielisters.go:198`](../../pkg/rest/entrielisters.go#L198)) |
| What happens | the path lacks `/snapshots`, so it reads the NAS volume itself; the reply has no `entries` key, so the list is empty, without an error. What the right path returns is misread too: [bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name](#bug_plugin_volume_snapshot_list_entries_read_as_snapshots_by_name) |
| Evidence | code reading |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping): the path is corrected; its callers by [impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow) |

#### bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete

| | |
|---|---|
| Where | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:72`](../../pkg/rest/nas_volumes_snapshots.go#L72)), called at `nfs_driver.go:322` |
| What happens | sends `POST` instead of `DELETE`; the appliance answers 405 |
| Evidence | hurl Finding 7 |
| Assigned to | [impl_plugin_rest_remove_dead_code_before_v4_move](#impl_plugin_rest_remove_dead_code_before_v4_move), the dead-code commit before stage 1 |

#### bug_plugin_rest_target_not_exists_pattern_is_a_hard_coded_test_name

| | |
|---|---|
| Where | `targetDneMsgPattern` ([`errors.go:96`](../../pkg/rest/errors.go#L96)) |
| What happens | a hard-coded test name, `Target iqn.20215:test2asd not exists.`, so it can never match a real target. The real 404 is `ItemNotFoundError`, `Target <name> not exists.`, so the general error mapping never returns `RestErrorResourceDNETarget`. `DettachVolumeFromTarget` sets that code itself on any 404 (`targets.go:185`), which leads to [bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone](#bug_rest_v4_pools_pool_san_iscsi_targets_target_luns_lun_delete_404_not_attached_taken_as_target_gone) |
| Evidence | hurl Finding 4 |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_plugin_rest_get_calls_decode_body_before_204_check

| | |
|---|---|
| Where | every `Get*` that accepts 204 (`GetVolume`, `GetPool`, `GetNASVolume`, `GetVolumeSnapshotClones`, `GetNASVolumeSnapshotClones`; e.g. [`storageapi.go:95-99`](../../pkg/rest/storageapi.go#L95)) |
| What happens | decodes the body **before** checking the status. A real 204 has no body, so decoding fails and returns `RestErrorRPM`: the 204 success branch can never be reached. No GET answered 204 in any recorded run |
| Evidence | code reading |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping): the status is checked before the body is decoded, and a GET answered 204 is sent again ([question 39](#decisions)) |

#### bug_plugin_rest_create_target_202_polls_get_target_ten_times_without_delay

| | |
|---|---|
| Where | `CreateTarget` on 202 ([`targets.go:89-98`](../../pkg/rest/targets.go#L89)) |
| What happens | calls `GetTarget` 10 times in a row, with no delay and no stop on success, and returns only the last result. 202 was never recorded, so this may never run |
| Evidence | code reading |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping): the 202 branch is removed |

#### bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check

| | |
|---|---|
| Where | `DeleteNASVolumeSnapshot` ([`nas_volumes_snapshots.go:79-85`](../../pkg/rest/nas_volumes_snapshots.go#L79)) |
| What happens | decodes the body **before** checking for 204. A 204 has no body, so decoding fails and the call returns `RestErrorRPM`. Hidden behind [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete): fixing only [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete)'s method would make every successful delete report an error. Its siblings `DeleteNASVolume` and `DeleteShare` check for 204 first. Not [bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check) (that is about GETs) |
| Evidence | code reading |
| Assigned to | [impl_plugin_rest_remove_dead_code_before_v4_move](#impl_plugin_rest_remove_dead_code_before_v4_move), the dead-code commit before stage 1 |

#### bug_plugin_rest_pool_and_snapshot_unmarshal_swallow_parse_errors

| | |
|---|---|
| Where | `ResourcePool.UnmarshalJSON` ([`types.go:337-340`](../../pkg/rest/types.go#L337)), `ResourceSnapshotShortProperties.UnmarshalJSON` (`types.go:283-286`) |
| What happens | a value that does not parse returns no error and leaves a zero value: pool `available` becomes 0 (GetCapacity would report a full pool), snapshot `creation` becomes the zero time |
| Evidence | code reading |
| Assigned to | [impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors), stage 2 |

#### bug_plugin_rest_get_error_maps_body_only_ignores_http_status

| | |
|---|---|
| Where | `getError` ([`storageapi.go:42-58`](../../pkg/rest/storageapi.go#L42)); every `Get*` decodes before looking at the status, except the two snapshot reads by name (`GetVolumeSnapshot`, [`storageapi.go:211`](../../pkg/rest/storageapi.go#L211); `GetNASVolumeSnapshot`, `nas_volumes_snapshots.go:111`) |
| What happens | only the JSON body is mapped; `getError` never sees the HTTP status. Error bodies that are not JSON (plain text `401 unauthorized` on jdss1, `404 not found`, both sent with the content type `text/html`, which is how the OpenAPI lists them) give `RestErrorRPM` ("Failure during processing response from storage") on the calls that decode the body before they look at the status (the GETs but the two snapshot reads by name, `CreateNASVolume`, `CreateShare` and `DeleteNASVolume`), and `RequestMalfunction` ("Failure in sending data to storage", `errors.go:287`) on the other calls, which wrongly points at the network. A wrong password gives no hint about credentials, and a plain 404 is not seen as "not found" |
| Evidence | recorded bodies, code reading |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping): the status is checked first and mapped when the body is not JSON |

### Driver bugs
[Driver_bugs](#driver-bugs)

Bugs outside `pkg/rest`, in the driver and controller code that calls it. This
document fixes them because they misread REST errors ([priority 8](#priorities)) or belong to the paging of the CSI lists ([question 60](#decisions)), and the NFS volume delete, because no volume that has a snapshot is deleted (user decision).

#### bug_plugin_driver_delete_volume_snapshot_cleanup_only_on_busy_not_on_has_children

| | |
|---|---|
| Where | the iSCSI `DeleteVolume`, in `deleteLUN` (`iscsi_driver.go:245`, `:257`). The NFS `deleteLUN` has the same check (`nfs_driver.go:383`, `:396`), but a NAS volume delete never gets `volume has children` ([dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots](#dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots)) |
| What happens | the intermediate-snapshot cleanup runs only for `Busy`. A leftover snapshot without clones gives `volume has children` (errno 1000), mapped to `RestErrorResourceBusyVolumeHasSnapshots`, so the cleanup never runs, `DeleteVolume` returns `Internal` on every retry, and **the volume can never be deleted**; a volume that has a CSI snapshot, and must not be deleted (user decision), gets `Internal` too, where CSI asks for `FailedPrecondition` |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### dataloss_plugin_driver_nfs_delete_volume_destroys_its_snapshots

| | |
|---|---|
| Where | the NFS `DeleteVolume`, in `deleteLUN` (`nfs_driver.go:380`) and `deleteNASVolAndShare` (`:358`) |
| What happens | the driver sends the delete of the NAS volume without looking for snapshots, and the appliance deletes a NAS volume that has a snapshot together with the snapshot, with 204 |
| Effect | DeleteVolume destroys every CSI snapshot of the volume, which may still be needed to restore it; a volume whose snapshot has a clone is refused, with a message the plugin does not recognise ([bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised](#bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised)) |
| Evidence | hurl Finding 7: the delete of a NAS volume that has a snapshot answered 204 and removed the snapshot with it; code reading |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 (user decision) |

#### bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error


| | |
|---|---|
| Where | `cleanIntermediateSnapshots` (`iscsi_driver.go:197-208`, `nfs_driver.go:302-312`) |
| What happens | the listing error is checked after the type check, so a real error such as "not found" becomes Unknown; and the NFS version expects the zvol snapshot type although its lister returns NAS snapshots, so it always fails with "Unable to typecast snapshot data" and never reaches the [bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete) call at `nfs_driver.go:322` |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [Stage 3](#stage-3-behaviour-and-error-mapping), its own fix ([question 17](#decisions)) |

#### bug_plugin_driver_zvol_snapshot_delete_every_clone_delete_error_taken_as_busy

| | |
|---|---|
| Where | `DeleteSnapshot` in the iSCSI driver (`iscsi_driver.go:515-523`) |
| What happens | when the appliance refuses a zvol snapshot's delete for its clones, the driver deletes each clone whose name has the plugin's snapshot form (`IsSDS`) with `DeleteClone`, and returns any error of that call as `Busy`, which the controller answers with `FailedPrecondition` (`controllerplugin.go:725-726`): a clone that is already gone and a lost connection included. The reply of `DeleteClone` to a missing clone is recognised, 500, errno 1, `Zfs resource: ... not found` ([gap_jdss_rest_delete_missing_zvol_clone](#gap_jdss_rest_delete_missing_zvol_clone)), so the fault is the caller's. |
| Evidence | code reading, and the reply of the gap run |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 ([question 18](#decisions)) |

#### bug_plugin_driver_list_resume_repeats_last_entry_returned

| | |
|---|---|
| Where | `getResourcesList` (`driver/common.go:62-66`, `:76`) |
| What happens | the next token names the last entry a call returns (`:76`), and the call that resumes from it takes the token's page from that entry on, the entry included (`:64`), so every resumed call starts with the entry the previous one ended with |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

#### bug_plugin_driver_list_resume_compares_base64_forms_not_names

| | |
|---|---|
| Where | `getResourcesList` (`driver/common.go:58`); the ID functions of `driver/restwrappers.go:29-57` |
| What happens | a resumed call decides whether a page lies after the token by comparing the base64 forms of names with `<=`, and base64 does not keep the order of the names: `vp_pvc-Z` sorts before `vp_pvc-_`, their base64 forms the other way round. So a page can be taken whole, entries already returned included, or searched for the token's entry and left out |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)); the example computed |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

#### bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages

| | |
|---|---|
| Where | `getResourcesList` (`driver/common.go:65`) |
| What happens | once the resumed call has found the token's entry, it calls `token.BasedID()`, which reads the name, where `DropBasedID()`, which clears it, was meant. Each later page of the call is compared with the token again and, unless its first entry's base64 form sorts at or after the token's, searched for the token's entry, not found, and left out whole |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

#### bug_plugin_driver_list_resume_drops_page_missing_token_entry

| | |
|---|---|
| Where | `getResourcesList` (`driver/common.go:61-68`) |
| What happens | a resumed call starts at the page the token names and looks for the token's entry there and on the later pages, never on an earlier one. When the entry is not there, because it was deleted, or because a create or a delete before it moved it to another page (hurl Finding 24), nothing of that page is taken and the token is kept, so the later pages fare as in [bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages](#bug_plugin_driver_list_resume_keeps_token_and_drops_later_pages): the list leaves pages out or ends early, without an error |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)); the shifting pages, hurl Finding 24 |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

#### bug_plugin_driver_snapshot_list_token_same_for_every_snapshot_of_a_volume

| | |
|---|---|
| Where | `RestSnapshotShortEntryBasedID` (`driver/restwrappers.go:39-42`), the token function of the iSCSI list of all snapshots |
| What happens | the token's name is built from the volume's name twice and never from the snapshot's, so every snapshot of one volume gets the same token; a resumed call finds the volume's first snapshot on the page instead of the last one returned, and returns the snapshots between them again. Stage 2's change of `Volume` to `Parent.Name` on this line keeps the fault |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

#### bug_plugin_ctrl_list_volumes_negative_max_entries_answered_internal

| | |
|---|---|
| Where | ListVolumes (`controllerplugin.go:597-599`) |
| What happens | a negative `max_entries`, which CSI forbids, is answered `Internal`, which says the plugin failed; ListSnapshots answers it `InvalidArgument` (`:762-763`) |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

### Security findings
[Security_findings](#security-findings)

All are assigned to this delivery, to [stage 4](#stage-4-robustness-and-security) unless said otherwise: [security_plugin_rest_tls_verification_off](#security_plugin_rest_tls_verification_off) to [impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile); [security_plugin_secrets_in_logs](#security_plugin_secrets_in_logs) and [security_plugin_rest_request_dump_every_call](#security_plugin_rest_request_dump_every_call) to [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) and [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs), the two commits [before stage 1](#before-stage-1-keep-secrets-out-of-the-logs); [security_plugin_rest_unescaped_path_segments](#security_plugin_rest_unescaped_path_segments) to [impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments).

#### security_plugin_rest_tls_verification_off

| | |
|---|---|
| Where | `rproxy.go:251` |
| Finding | TLS certificate checks are always off (`InsecureSkipVerify: true`, no setting), while the admin's Basic credentials go with every request |
| Evidence | code reading (the security review, [ID sources](#id-sources)) |
| Assigned to | this delivery: verification on by default, `insecure` and `cafile` settings ([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile), stage 4) |

#### security_plugin_rest_unescaped_path_segments

| | |
|---|---|
| Where | every path builder in `pkg/rest` |
| Finding | paths are built with `fmt.Sprintf` and no `url.PathEscape`; a `?` or `#` in a name cuts the path short, and a crafted volume handle can reach other resources |
| Evidence | code reading (the security review, [ID sources](#id-sources)) |
| Assigned to | this delivery: `url.PathEscape` on every path segment ([impl_plugin_rest_escape_path_segments](#impl_plugin_rest_escape_path_segments), stage 4) |

#### security_plugin_secrets_in_logs

| | |
|---|---|
| Where | `Send` (`rproxy.go:113`, `:118`, `:130-131`); the controller (`controllerplugin.go:479`, `:546`, `:634`, `:710`, `:857` at Warn; `:587`, `:626`, `:703`, `:750`, `:837`, `:914` at Debug) and the node (`node/publish.go:43`, `:68`) |
| Finding | secrets reach the logs. `Send` logs every request body at Debug and dumps the whole request with its `Authorization: Basic` header, which carries the REST password. The controller and the node print whole CSI requests, `Secrets` included, at Debug; the controller also at Warn when the capability check of a create, delete or publish fails, which the fixed list of capabilities (`controllerplugin.go:63-70`) never lets happen, so in practice at Debug only |
| Evidence | code reading (the security review, [ID sources](#id-sources)) |
| Assigned to | the REST part: [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging); the controller and node part: [impl_plugin_no_secrets_in_logs](#impl_plugin_no_secrets_in_logs); the two commits [before stage 1](#before-stage-1-keep-secrets-out-of-the-logs) |

#### security_plugin_rest_request_dump_every_call

| | |
|---|---|
| Where | `Send` (`rproxy.go:130-131`) |
| Finding | `httputil.DumpRequestOut` runs on every request, whatever the log level: the dump, with the body and the `Authorization` header, is built even when nothing is logged, and is written at Debug |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging), the first logging commit [before stage 1](#before-stage-1-keep-secrets-out-of-the-logs): the dump call is removed |

### Crashes
[Crashes](#crashes)

Places where the plugin can crash the controller (a Go panic): five in the REST
layer, which `pkg/rest` fixes, four of them in stage 4 ([impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)) and the nil body in stage 3 ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)), and
one in the parse of the CSI list token, which stage 3 fixes ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)).

#### crash_plugin_rest_nil_response_on_transport_error

| | |
|---|---|
| Where | `rproxy.go:138-157` |
| Crash | every branch that handles a transport error reads `res.StatusCode` while `res` is `nil` |
| Trigger | a refused connection (`:141`), or the client's own timeout, which is not a `*net.OpError` (`:153`). During jdss1's 8-hour outage a controller would have crashed again and again (inferred). Gap [gap_plugin_rest_appliance_unreachable](#gap_plugin_rest_appliance_unreachable) |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)); no recording, since the hurl suite records replies and this path has none |
| Assigned to | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes), stage 4 |

#### crash_plugin_rest_basic_auth_before_request_check

| | |
|---|---|
| Where | `rproxy.go:122-123` |
| Crash | `SetBasicAuth` runs before the error from `http.NewRequest` is checked |
| Trigger | a config without `prot` (there is no default) gives the URL `://host`, and `NewRequest` fails |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes), stage 4 |

#### crash_plugin_rest_error_from_nil_body

| | |
|---|---|
| Where | `errors.go:140-149` |
| Crash | `ErrorFromErrorT(nil)` logs the `nil` case, then uses it anyway |
| Trigger | a list reply with `"data": null`, or a list's error reply without `"error"`. Gap [gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies) |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)); no recorded reply reaches it ([gap_plugin_rest_unexpected_reply_bodies](#gap_plugin_rest_unexpected_reply_bodies)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### crash_plugin_rest_setup_endpoint_ignores_error

| | |
|---|---|
| Where | `rest.go:104-107` |
| Crash | `SetupEndpoint` logs the error from `SetupRestProxy` and returns `nil` (`rest.go:105-107`) |
| Trigger | a missing or invalid `idletimeout`: `SetupRestProxy` returns an error and the proxy stays unset. An empty `addrs`: `SetupRestProxy` returns no error, and `Send` indexes the empty list (`rproxy.go:91`). In both cases the first REST call panics |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes), stage 4 |

#### crash_plugin_rest_endpoint_string_empty_addrs

| | |
|---|---|
| Where | `rest.go:83-95` |
| Crash | `RestEndpoint.String()` throws away the string it builds and returns `Addrs[0]`, so it panics on an empty list |
| Trigger | an endpoint with no addresses formatted with `%v` or `%s`: a log line alone can panic |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes), stage 4 |

#### crash_plugin_driver_list_token_parse_writes_through_nil_pointer

| | |
|---|---|
| Where | `NewCSIListingTokenFromTokenString` (`driver/snaptoken.go:78-113`), called by ListVolumes and ListSnapshots (`controllerplugin.go:591`, `:756`) |
| Crash | the function writes the token's page, `_dc` value and name through its named return `t`, which is still `nil` (`:100`, `:106`, `:110`), where the local `ct` was meant |
| Trigger | a `starting_token` whose first part, up to the first `_`, is a number: every `next_token` the plugin returns. The gRPC server recovers no panic (`pluginserver.go:121-133`), so the controller exits. A token without a `_`, such as csi-sanity's `invalid-token`, or whose first part is not a number, fails the parse before the write (`snaptoken.go:93`, `:97`) and gets `Aborted` |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name), stage 3 |

### REST client behaviour
[REST_client_behaviour](#rest-client-behaviour)

Not wrong replies, but how the client itself behaves: retries, failover, timeouts,
error text and logging. For a later REST client design, except [bug_plugin_rest_unknown_error_empty_message](#bug_plugin_rest_unknown_error_empty_message) and [bug_plugin_rest_every_error_logged_at_warn](#bug_plugin_rest_every_error_logged_at_warn) (both assigned to stage 3, [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) and [bug_plugin_rest_reply_text_as_format_string](#bug_plugin_rest_reply_text_as_format_string) ([impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply) in the first commit, [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes) in stage 4).

#### bug_plugin_rest_proxy_tries_and_extra_addrs_unused_no_retry_no_failover

| | |
|---|---|
| Where | `rproxy.go:91`, `:274`, `:281` |
| Behaviour | `tries` and every address after the first are never used: no retry and no failover (gap [gap_plugin_rest_failover_during_call](#gap_plugin_rest_failover_during_call)) |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | not this document: a later REST client design ([Relationship](#relationship-to-other-designs)). Only `tries` gets a use here, for a GET answered without data ([Stage 3](#stage-3-behaviour-and-error-mapping), [question 39](#decisions)) |

#### bug_plugin_rest_reply_text_as_format_string

| | |
|---|---|
| Where | `storageapi.go:51`, `:53`, `errors.go:173`, `:265`, `targets.go:45`, `:76`, `:150`, `:184`; transport errors: `rproxy.go:148`, `:153`, `:157` |
| Behaviour | the reply body, or a name, is used as a format string; transport errors are reported as `RequestMalfunction`, which hides what kind of error it was. |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)); the eight `go vet` reports in `pkg/rest` ([Advice](#advice)) |
| Assigned to | the format-string part: [impl_plugin_rest_no_format_string_from_reply](#impl_plugin_rest_no_format_string_from_reply), the first commit; the transport-error part: [impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes), stage 4 |

#### bug_plugin_rest_request_built_without_context_csi_deadline_not_applied

| | |
|---|---|
| Where | `rproxy.go:122`, `:257` |
| Behaviour | the request is built without the context, so the CSI deadline never reaches it; `idletimeout` is used as the timeout for the whole request ([risk 12](#risks), [future work](#future-work-request-deadlines)) |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | not this document: a later REST client design ([request deadlines](#future-work-request-deadlines)) |

#### bug_plugin_rest_unknown_error_empty_message

| | |
|---|---|
| Where | `errors.go:275-277`, `:297` |
| Behaviour | errors that match no pattern get an empty message; the appliance's text appears only in a Warn log. |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

#### bug_plugin_rest_every_error_logged_at_warn

| | |
|---|---|
| Where | `errors.go:150`, `:173`, `:265`, `:291` |
| Behaviour | every appliance error that carries an `errno`, and every one that matches no pattern, is logged at Warn, including the expected "not found" on repeated deletes; the not-found message repeats a word ("Resource Resource ... do not exists"). |
| Evidence | code reading (the code reviews, [ID sources](#id-sources)) |
| Assigned to | [impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue), stage 3 |

### Dead code
[Dead_code](#dead-code)

Code with no callers in the current code, keys the driver never sets, and `DeleteNASVolumeSnapshot`, whose one call the [dead-code commit](#before-stage-1-remove-dead-rest-code) points at `DeleteNASSnapshot`. That commit removes all of it.

- **Functions:** `ListVolumes` (with `GetVolumesData`), `getResultsEntries`, `GetAddress` (named only by the dead `StorageInterface`),
  `GetTimeStamp`, `CreateNASVolumeSnapshot` (`bug_plugin_rest_nas_snapshot_create_copy_of_volume_create`: a copy of `CreateNASVolume` that
  would create a NAS volume), `DeleteNASVolumeSnapshot` ([bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete](#bug_plugin_rest_delete_nas_volume_snapshot_sends_post_not_delete), [bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check](#bug_plugin_rest_delete_nas_volume_snapshot_decodes_body_before_204_check); its one call is pointed at `DeleteNASSnapshot` in the same commit), `PromoteClone` and `AddUserToTarget` (stubs
  that return nil), `DeleteShare` (`bug_plugin_rest_delete_share_nil_descriptor`: sends nil instead of its descriptor; its
  only caller is commented out at `nfs_driver.go:344`; the recordings show that
  deleting a NAS volume removes its shares, so it is not needed).
- **Types:** `StorageInterface` (old signatures that `RestEndpoint` does not
  implement), `RestProxyInterface`, the duplicate `restProxy` struct, `RestProxyCfg`, `GetResourceEntries`,
  `StorageCfg` (with its one user, the dead `ControllerCfg` of
  `pkg/controller/config.go`), `SnapshotDescriptor`, and the unused `*Data` reply types in
  `payload-def.go` (about 16).
- **Request keys** the driver never sets and the company schema does not have ([rule_rest_request_body_valid](#rule_rest_request_body_valid)):
  `TargetLunDescriptor.EUI`, `CreateNASVolumeDescriptor.ResWithDescendents` and
  `.RefquotaWithDescendents`.

### Measurements
[Measurements](#measurements)

**Comparisons recorded** by the hurl suite
(`jdss-rest-api-spec/empirical/hurl/`):

**Table 15 — Recorded comparisons** · `tbl_recorded_comparisons`

| Comparison | Recorded | Cases | Differences | Source |
|---|---|---|---|---|
| jdss1 v3 against jdss1 v4, lists and delete flags | jdss1 run `20261104-200231` | 363 delete cases per version, and the list questions | none: status, errno, class and message identical | hurl Finding 14 |
| jdss1 v3 against jdss1 v4, rollback and clone id | jdss1, 2026-11-04 | 281 common | 7: one count of the rollback check on snapshots made with no pause between them (case `rapid`, 1 against 0) and six NAS rollback cases (`effect`) that got no answer during an appliance outage on the v3 side | hurl Finding 15 |
| jdss2 v4 against jdss1 v4, lists and delete flags | jdss2, 2026-11-04 | 751 common | 12: eleven in the share 404 that jdss2 sends with `errno` 1 and no class ([bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class](#bug_jdss_rest_jdss2_v4_shares_share_delete_404_without_class)), one a virtual IP address inside a message | hurl Finding 15 |
| with and without the `{}` body on GET and on DELETE of target, LUN and share | jdss1 v3, jdss2 v4 and jdss1 v4 | 80 and 82 pairs; on jdss1 v4, 82 GET pairs and 134 DELETE pairs | none | [how the gap runs went](#how-the-gap-runs-went), [gap_jdss_rest_empty_body_on_get_delete](#gap_jdss_rest_empty_body_on_get_delete) |

**Hard-coded API paths in the current code:** 34 `api/v3` strings in
`pkg/rest`, 22 distinct path patterns; 30 strings after the dead-code commit.

**List requests, jdss1 v4**, the volume list (hurl Finding 24, run
`20261117-032241`; the reply size measured on its recorded reply that
holds every entry):

**Table 16 — List request times on jdss1 v4** · `tbl_list_request_times`

| Volumes on the pool | Time of one request | Reply |
|---|---|---|
| about 100 | about 0.3 s | not measured |
| 501 | 1.3 to 2.1 s, whatever `per_page` | 1,011,556 bytes for all 501 entries, about 2 KB an entry |

**The REST package, measured on commit `11ffbfd`:**

**Table 17 — Files of the REST package** · `tbl_rest_package_files`

| File | Lines | Role |
|---|---|---|
| `types.go` | 675 | 19 `Resource*` reply types, 10 accessors, 8 custom decoders |
| `payload-def.go` | 508 | mostly unused types, removed by the dead-code commit |
| `storageapi.go` | 401 | volume and snapshot calls, `getError` |
| `errors.go` | 306 | error mapping |
| `entrielisters.go` | 291 | list calls |
| `rproxy.go` | 284 | transport |
| `targets.go` | 260 | target and LUN calls |
| `reqtypes.go` | 254 | request types |
| `nas_volumes.go`, `nas_volumes_snapshots.go`, `nas_volumes_snapshots_clones.go`, `clones.go` | 462 | NAS and clone calls |
| `rest.go` | 108 | setup |
| `pools.go`, `nas_snapshots.go`, `shares.go`, `general_resources.go` | 257 | pool, NAS snapshot and share calls; the last holds only the licence header |
| total | 3,806 | |

**Reply types with their fields, accessors and decoders, on commit `11ffbfd`:**

**Table 18 — Reply types, fields, accessors and decoders** · `tbl_reply_type_fields`

| Type | JSON fields | Accessors | Custom decoder converts |
|---|---|---|---|
| `ResourceVolume` | 86 | `GetSize`, `OriginVolume`, `OriginSnapshot` | `volsize` string to `int64` |
| `ResourceSnapshot` | 25 | `ClonesNames`, `GetSize` | `creation` string to a time, `volsize` string to `int64` |
| `ResourceSnapshotShortProperties` | 2 | | `creation` string to number (the list sends `int64`: [bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch](#bug_rest_v4_pools_pool_volumes_snapshots_get_creation_type_mismatch)) |
| `ResourcePool` | 15 | | `available` string to `int64` |
| `ResourceNASVolume` | 28 | `GetSize`, `OriginVolume`, `OriginSnapshot` | `copies`, `recordsize`, `quota` strings to numbers |
| `ResourceNASVolumeSnapshot` | 33 | `ClonesNames`, `GetSize` | `creation` |
| `ResourceNASSnapshotShortProperties` | 3 | | `creation` |
| `ResourceNASSnapshotParent` | 3 | | `copies`; reads `type` under the wrong key ([bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key](#bug_plugin_rest_nas_snapshot_parent_type_read_from_copies_key)) |
| `ResourceTarget`, `ResourceShare`, `ResourceVolumeSnapshotClones`, pool sub-types | 1 to 8 each | | none |

**Observed types, jdss1 v4.** Fields the plugin reads: volume by name:
`name` and `volsize` strings; `origin` `null` in the zvol alternative and a
string in the clone alternative. From stage 3, in the volume list as by
name: `guid` and `default_scsi_id` strings, and `is_clone`, `false` in the
zvol alternative and `true` in the clone alternative (`jdss-rest-api-spec` commit
`434c9cd`, [question 65](#decisions)); `is_clone` is read as optional
even so.
All-snapshots list: `name` string, `parent.name` string, `properties.creation` `int64`,
and from stage 3 `parent.type`, a number, and `properties.guid` string.
Pool: `available` string.
Fields of the same replies that nothing reads: volume `available`, `used`,
`reservation` and `refreservation` strings; `clones` and `quota` `null` in
every recording. Pool `size`, `name` and `health` strings; `status` `int64`.

**The kinds of error of the plugin's calls, and their codes**, from the
scenarios files of the calls on jdss1 v4 at `jdss-rest-api-spec` commit
`434c9cd`
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue),
[question 75](#decisions)). A kind is an error model of the spec, under the
model's name, and each has one class, which gives its layer. Inside the
range of a layer the kinds are numbered here in the order of their names. A
kind keeps its number: one that a later reading of the spec adds takes the
next free number of its layer, and no number is used again. The plain-text
replies that the scenarios files list, 401 and 404 without an error object,
have no kind; they carry the code of the layer without a class, 1000. The
Message column gives one message of the kind as the spec writes it, cut
where it is long; the patterns that the entries match on are in the
scenarios files.

**Table 19 — Kinds of error and their codes** · `tbl_error_kind_codes`

| Code | Kind of error | Class | Status | Message |
|---|---|---|---|---|
| 1001 | `Unclassed1PathPoolsDoesnTPointTo` | none | 400 | `Path /Pools/{{pool}}/{{name}} doesn't point to NAS volume.` |
| 1002 | `Unclassed1ShareAlreadyExists` | none | 409 | `Share {{body.name}} already exists.` |
| 1003 | `Unclassed2PathPoolsIsAlreadyUsedIn` | none | 409 | `Path /Pools/{{pool}}/{{name}} is already used in a different share.` |
| 1004 | `Unclassed4ValueContainsInvalidCharactersOnlyA` | none | 400 | `Value {{body.name}} contains invalid characters. Only a-z, A-Z, 0-9 . : $ _ and - are allo …` |
| 1005 | `Unclassed4ValueHasInvalidLengthMaximumOf` | none | 400 | `Value has invalid length. Maximum of 255 characters are allowed.` |
| 1006 | `Unclassed4ValueHasInvalidLengthMinimumCharacter` | none | 400 | `Value has invalid length. Minimum 1 character required.` |
| 1007 | `Unclassed6NotValidCharactersInARecycle` | none | 400 | `Not valid characters in a recycle bin name. A name cannot contain any of the following cha …` |
| 2001 | `KeyError_004371` | `builtins.KeyError` | 500 | `'{{query.search_by}}'` |
| 2002 | `KeyError_40bcba` | `builtins.KeyError` | 500 | `'{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}'` |
| 2003 | `OSError36ErrnoFileNameTooLongNfs` | `builtins.OSError` | 500 | `[Errno 36] File name too long: '/NFS/{{body.name}}'` |
| 2004 | `TypeErrorNotSupportedBetweenInstancesOfNonetype_a38cd7` | `builtins.TypeError` | 500 | `'<' not supported between instances of 'NoneType' and 'NoneType'` |
| 2005 | `TypeErrorNotSupportedBetweenInstancesOfNonetype_b1b8e2` | `builtins.TypeError` | 500 | `'<' not supported between instances of 'NoneType' and 'str'` |
| 2006 | `UnicodeDecodeErrorUtfCodecCanTDecodeByte` | `builtins.UnicodeDecodeError` | 500 | `'utf-8' codec can't decode byte {{echo.byte}} in position {{echo.position}}: invalid conti …` |
| 2007 | `ValueErrorDoesNotAppearToBeAn` | `builtins.ValueError` | 500 | `'{{body.allow_ip}}' does not appear to be an IPv4 or IPv6 network` |
| 2008 | `ValueErrorInvalidLiteralForIntWithBase` | `builtins.ValueError` | 500 | `invalid literal for int() with base 10: '{{query.page}}'` |
| 2009 | `ValueErrorSizeSuffixGbIsIncorrect` | `builtins.ValueError` | 500 | `Size suffix GB is incorrect.` |
| 2010 | `ValueErrorSizeUnitHasIncorrectValue` | `builtins.ValueError` | 500 | `Size unit {{body.blocksize}} has incorrect value.` |
| 3001 | `PropertyError2CannotSetPropertyNosuchpropertyItNot` | `zfslib.zfsapi.properties.PropertyError` | 500 | `Cannot set property nosuchproperty it not exists.` |
| 3002 | `ZfsCmdError0CannotOpenDatasetDoesNotExist_96e556` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}': dataset does not exist` |
| 3003 | `ZfsCmdError0CannotOpenDatasetDoesNotExist_a33d60` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.poolname}}@{{path.poolname}}': dataset does not exis …` |
| 3004 | `ZfsCmdError1000CannotCreateInvalidNumericSuffix` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create '{{path.poolname}}/{{body.name}}': invalid numeric suffix '{{echo.suffix}}'` |
| 3005 | `ZfsCmdError1000CannotCreateSnapshotBookmarkDelimiterIs` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': bookmark de …` |
| 3006 | `ZfsCmdError1000CannotCreateSnapshotEmptyComponentOr` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@': empty component or mispl …` |
| 3007 | `ZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_8fa47d` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': invalid cha …` |
| 3008 | `ZfsCmdError1000CannotCreateSnapshotInvalidCharacterIn_cee0b6` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': invalid cha …` |
| 3009 | `ZfsCmdError1000CannotCreateSnapshotMultipleAndOr` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': multiple '@ …` |
| 3010 | `ZfsCmdError1000CannotCreateSnapshotTrailingSlashIn` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': trailing sl …` |
| 3011 | `ZfsCmdError1000CannotCreateSnapshotsThisPropertyCan` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshots : this property can not be modified for snapshots` |
| 3012 | `ZfsCmdError1000CannotCreateUseNoneToDisable` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create '{{path.poolname}}/{{body.name}}': use 'none' to disable quota/refquota` |
| 3013 | `ZfsCmdError1000CannotCreateVolumeSizeCannotBe` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create '{{path.poolname}}/{{body.name}}': volume size cannot be zero` |
| 3014 | `ZfsCmdError1000CannotDestroySnapshotHasDependentClones` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot destroy '{{path.poolname}}/{{path.volumename}}@{{path.snapshotname}}': snapshot has …` |
| 3015 | `ZfsCmdError1000CannotDestroyVolumeHasChildren` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot destroy '{{path.poolname}}/{{path.clonename}}': volume has children\nuse '-r' to des …` |
| 3016 | `ZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_214e41` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}': bookmark delimiter '#' is not expect …` |
| 3017 | `ZfsCmdError1000CannotOpenBookmarkDelimiterIsNot_c9f5d6` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}@{{path.snapshotname}}': bookmark delim …` |
| 3018 | `ZfsCmdError1000CannotOpenInvalidCharacterInName_61dfda` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}@{{path.snapshotname}}': invalid charac …` |
| 3019 | `ZfsCmdError1000CannotOpenInvalidCharacterInName_6d342b` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}': invalid character '{{echo.char}}' in …` |
| 3020 | `ZfsCmdError1000CannotOpenMultipleAndOrDelimiters` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}@{{path.snapshotname}}': multiple '@' a …` |
| 3021 | `ZfsCmdError1000CannotOpenNameIsTooLong_37e45d` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{name}}: name is too long` |
| 3022 | `ZfsCmdError1000CannotOpenNameIsTooLong_671b8f` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}@{{name}}: name is too long` |
| 3023 | `ZfsCmdError1000CannotOpenNameIsTooLong_a68ac4` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}@{{path.snapshotname}}': name is too lo …` |
| 3024 | `ZfsCmdError1000CannotOpenNameIsTooLong_d50fea` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot open '{{path.poolname}}/{{path.datasetname}}': name is too long` |
| 3025 | `ZfsCmdError1000WarningVolblocksizeIsLessThanTheVolumeSizeCannotBe` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `Warning: volblocksize ({{echo.blocksize}}) is less than the default minimum block size (16 …` |
| 3026 | `ZfsCmdError100CannotCreateDatasetAlreadyExists` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create '{{path.poolname}}/{{body.name}}': dataset already exists` |
| 3027 | `ZfsCmdError201CannotCreateSnapshotNameIsToo_abcab7` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{name}}: name is too long` |
| 3028 | `ZfsCmdError201CannotCreateSnapshotNameIsToo_d5a47e` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create snapshot '{{path.poolname}}/{{path.datasetname}}@{{body.name}}': name is too …` |
| 3029 | `ZfsCmdError4CannotCreateOutOfSpace` | `zfslib.wrap.zfs.ZfsCmdError` | 500 | `cannot create '{{path.poolname}}/{{body.name}}': out of space` |
| 3030 | `ZfsResourceError1ZfsResourceNotFoundInThis_3b952d` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Zfs resource: {{path.poolname}}/{{path.clonename}} not found in this collection.` |
| 3031 | `ZfsResourceError1ZfsResourceNotFoundInThis_597b56` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Zfs resource: {{path.poolname}}@{{path.poolname}} not found in this collection.` |
| 3032 | `ZfsResourceError1ZfsResourceNotFoundInThis_8f51ec` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Zfs resource: {{path.poolname}}/{{path.datasetname}}@{{path.datasetname}} not found in thi …` |
| 3033 | `ZfsResourceError1ZfsResourceNotFoundInThis_a720ee` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Zfs resource: {{path.poolname}}/{{name}} not found in this collection.` |
| 3034 | `ZfsResourceError4DatasetCanTContainCharacterIn` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Dataset can't contain '@' character in int's full name.` |
| 3035 | `ZfsResourceError4ZvolCanTContainCharacterIn` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Zvol can't contain '@' character in int's full name.` |
| 3036 | `ZfsResourceError5ResourceAlreadyExists` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Resource {{path.poolname}}/{{body.name}} already exists.` |
| 3037 | `ZfsResourceError5SnapshotAlreadyExists` | `zfslib.zfsapi.resources.ZfsResourceError` | 500 | `Snapshot {{path.poolname}}/{{path.datasetname}}@{{body.name}} already exists.` |
| 3038 | `ZpoolError1GivenZpoolDoesnTExists` | `zfslib.zfsapi.zpool.ZpoolError` | 500 | `Given zpool '{{path.poolname}}' doesn't exists.` |
| 4001 | `AddressValueError` | `opene.lib.ipaddr.AddressValueError` | 500 | `{{body.nfs.allow_access_ip}}` |
| 4002 | `CalculationZvolSizeErrorVolumeSizeItsBlockSizeAnd` | `opene.tools.zreservation.CalculationZvolSizeError` | 500 | `volume size, its block size and amount of copies must be greater than 0` |
| 4003 | `DatasetValidationError1DatasetNameTooLongDatasetName` | `opene.storage.zfs.exceptions.DatasetValidationError` | 400 | `Dataset name too long. Dataset name cannot exceed 132 characters.` |
| 4004 | `DatasetValidationError2DatasetNameIsInvalidAllowedCharacters` | `opene.storage.zfs.exceptions.DatasetValidationError` | 400 | `Dataset name is invalid. Allowed characters in dataset name: a-z, A-Z, 0-9, _, - ` |
| 4005 | `ItemConflictErrorLunIsAlreadyUsedIn` | `opene.exceptions.ItemConflictError` | 409 | `LUN 0 is already used in {{path.targetname}}.` |
| 4006 | `ItemConflictErrorVolumeDevIsAlreadyUsed` | `opene.exceptions.ItemConflictError` | 409 | `Volume /dev/{{path.poolname}}/{{body.name}} is already used.` |
| 4007 | `ItemNotFoundErrorPoolDoesNotExist` | `opene.exceptions.ItemNotFoundError` | 404 | `Pool {{path.poolname}} does not exist.` |
| 4008 | `ItemNotFoundErrorPoolNotFound` | `opene.exceptions.ItemNotFoundError` | 404 | `Pool {{path.poolname}} not found.` |
| 4009 | `ItemNotFoundErrorResourceNotFound` | `opene.exceptions.ItemNotFoundError` | 404 | `Resource {{path.poolname}}/{{path.datasetname}} not found.` |
| 4010 | `ItemNotFoundErrorTargetNotExists` | `opene.exceptions.ItemNotFoundError` | 404 | `Target {{path.targetname}} not exists.` |
| 4011 | `ItemNotFoundErrorTargetPartitionNotExists` | `opene.exceptions.ItemNotFoundError` | 404 | `Target partition: {{path.poolname}} not exists.` |
| 4012 | `ItemNotFoundErrorVolumeNameIsNotAttachedTo` | `opene.exceptions.ItemNotFoundError` | 404 | `Volume name {{path.zvol}} is not attached to target {{path.targetname}}` |
| 4013 | `ItemNotFoundErrorVolumeNotFoundInPool` | `opene.exceptions.ItemNotFoundError` | 404 | `Volume {{body.name}} not found in pool {{path.poolname}}.` |
| 4014 | `TargetNameConflictErrorTargetAlreadyExistsInScst` | `opene.san.target.base.iscsi.target.TargetNameConflictError` | 409 | `Target already exists in SCST.` |
| 4015 | `TargetNameConflictErrorTargetWithNameIqnOrgHurltest` | `opene.san.target.base.iscsi.target.TargetNameConflictError` | 409 | `Target with name iqn.2026-11.org.hurltest:{{name}} is already present on {{path.poolname}} …` |
| 4016 | `TargetNameConflictErrorTargetWithNameIsAlreadyPresent_7a89c4` | `opene.san.target.base.iscsi.target.TargetNameConflictError` | 409 | `Target with name {{body.name}} is already present on {{path.poolname}}.` |
| 4017 | `TargetNameConflictErrorTargetWithNameIsAlreadyPresent_cfbcb7` | `opene.san.target.base.iscsi.target.TargetNameConflictError` | 409 | `Target with name {{name}} is already present on {{path.poolname}}.` |
| 4018 | `ValidationError_7d2ccd` | `opene.exceptions.ValidationError` | 400 | `Block size have to be greater than or equal 512.` |
| 4019 | `ZfsOeErrorInOrderToDeleteADataset` | `opene.storage.zfs.zfs.ZfsOeError` | 500 | `In order to delete a dataset, you must delete all of its clones first.` |
| 4020 | `ZfsOeErrorInOrderToDeleteAZvol` | `opene.storage.zfs.zfs.ZfsOeError` | 500 | `In order to delete a zvol, you must delete all of its clones first.` |
| 4021 | `ZfsOeErrorResourceIsNotADatasetType` | `opene.storage.zfs.zfs.ZfsOeError` | 500 | `Resource {{path.datasetname}} is not a dataset type.` |
| 4022 | `ZfsOeErrorResourceIsNotAZvolType` | `opene.storage.zfs.zfs.ZfsOeError` | 500 | `Resource {{path.volumename}} is not a zvol type.` |
| 4023 | `ZvolValidationError1ZvolNameTooLongZvolName` | `opene.storage.zfs.exceptions.ZvolValidationError` | 400 | `Zvol name too long. Zvol name cannot exceed 132 characters.` |
| 4024 | `ZvolValidationError2ZvolNameIsInvalidAllowedCharacters` | `opene.storage.zfs.exceptions.ZvolValidationError` | 400 | `Zvol name is invalid. Allowed characters in zvol name: a-z, A-Z, 0-9, _, - ` |
| 5001 | `ValidationError_6317af` | `jsonschema.exceptions.ValidationError` | 400 | `'' does not match '^(?P<size_value>\\d+([.,]\\d+)?(e\\+\\d+)?)\\s*(?P<size_suffix>[BKMGT]i …` |
| 6001 | `BadRequestAllowedValuesForBlocksizeAre` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for blocksize are: 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, …` |
| 6002 | `BadRequestAllowedValuesForCompressionAreOff` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for compression are: off, on, lzjb, gzip, gzip-1, gzip-2, gzip-3, gzip-4, g …` |
| 6003 | `BadRequestAllowedValuesForCopiesAre` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for copies are: 1, 2, 3` |
| 6004 | `BadRequestAllowedValuesForDedupAreOff` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for dedup are: off, on, verify, sha256, sha256,verify` |
| 6005 | `BadRequestAllowedValuesForLogbiasAreLatency` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for logbias are: latency, throughput` |
| 6006 | `BadRequestAllowedValuesForPrimarycacheAreAll` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for primarycache are: all, none, metadata` |
| 6007 | `BadRequestAllowedValuesForSecondarycacheAreAll` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for secondarycache are: all, none, metadata` |
| 6008 | `BadRequestAllowedValuesForSyncAreAlways` | `werkzeug.exceptions.BadRequest` | 400 | `Allowed values for sync are: always, standard, disabled` |
| 6009 | `BadRequestNewZvolSizeExceedsAvailableSpace` | `werkzeug.exceptions.BadRequest` | 400 | `New zvol size({{echo.size}}) exceeds available space on pool {{path.poolname}}({{echo.avai …` |
| 6010 | `NotFound` | `werkzeug.exceptions.NotFound` | 404 | `"{{path.clonename}}"` |
| 6011 | `NotFoundB_2e360f` | `werkzeug.exceptions.NotFound` | 404 | `'{{name}}\\b'` |
| 6012 | `NotFoundB_699120` | `werkzeug.exceptions.NotFound` | 404 | `'{{name}}\nb'` |
| 6013 | `NotFoundCloneNotExists_0d481d` | `werkzeug.exceptions.NotFound` | 404 | `Clone {{path.clonename}} not exists` |
| 6014 | `NotFoundShareNotExists` | `werkzeug.exceptions.NotFound` | 404 | `Share {{path.sharename}} not exists.` |
| 6015 | `NotFoundTb` | `werkzeug.exceptions.NotFound` | 404 | `'{{name}}\tb'` |

### Code segments
[Code_segments](#code-segments)

The code of the parts of the work where the order of the steps, or an exact
rule, decides what a caller gets: in `pkg/rest`, the setup of the client,
one request and its reply, a GET answered without data, the error body, a
failed reply and a reply type; in `pkg/driver`, the paging of the CSI lists,
its token, and the delete of a snapshot that may have clones. The segments
follow a call from the client to the driver; each is written in Go and
names the files it goes into. Names that the current code does not have are proposals; the
order of the steps and the rules are what an implementation keeps.

**The setup of the client**
([impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile](#impl_plugin_rest_tls_verification_on_by_default_with_insecure_and_cafile),
[impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)).
`SetupRestProxy` (`pkg/rest/rproxy.go`) refuses a setting it cannot use,
naming it, and builds a client that speaks HTTPS only, verifies
certificates unless `insecure` is set, takes TLS 1.2 or higher and follows
no redirect. `SetupEndpoint` (`rest.go`) returns the setup error, logs it, and
keeps it for `Send`.

```go
// The settings a configuration may leave out.
const (
	defaultTries        = 3 // the first request and two more
	defaultRetryDelay   = time.Second
	defaultMaxReplySize = 1 << 20 // 1 MiB
)

// SetupRestProxy checks the endpoint configuration and builds the client.
// Its error names the setting it refuses.
func SetupRestProxy(rp *RestProxy, cfg *jcom.RestEndpointCfg, l *logrus.Entry) error {
	rp.l = l.WithFields(logrus.Fields{"section": "restproxy", "addrs": cfg.Addrs, "port": cfg.Port})
	if len(cfg.Addrs) == 0 {
		return fmt.Errorf("addrs: no address of the appliance is set")
	}
	if cfg.Prot != "" && cfg.Prot != "https" {
		return fmt.Errorf("prot: %q is refused, the plugin uses https only", cfg.Prot)
	}
	timeout, err := time.ParseDuration(cfg.IdleTimeOut)
	if err != nil {
		return fmt.Errorf("idletimeout: %w", err)
	}
	retryDelay := defaultRetryDelay
	if cfg.RetryDelay != "" {
		if retryDelay, err = time.ParseDuration(cfg.RetryDelay); err != nil {
			return fmt.Errorf("retrydelay: %w", err)
		}
	}
	if cfg.MaxReplySize < 0 {
		return fmt.Errorf("maxreplysize: %d is refused, it is a number of bytes", cfg.MaxReplySize)
	}

	tc := &tls.Config{MinVersion: tls.VersionTLS12}
	switch {
	case cfg.Insecure:
		tc.InsecureSkipVerify = true // no check at all; cafile is not read
	case cfg.CAFile != "":
		pem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return fmt.Errorf("cafile: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(pem) {
			return fmt.Errorf("cafile: %s holds no PEM certificate", cfg.CAFile)
		}
		// the appliance's own certificate, or the CA that signed it; the
		// certificate must still be within its dates and list the address
		tc.RootCAs = roots
	}
	rp.httpRestProxy = &http.Client{
		Transport: &http.Transport{IdleConnTimeout: sessionTimeout, TLSClientConfig: tc},
		Timeout:   timeout,
		// The API has no reason to redirect: a redirect means that something
		// else answered, and its reply is returned as it came.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	rp.prot = "https"
	rp.addrs = append(rp.addrs, cfg.Addrs...)
	rp.active_addr = 0
	rp.port = cfg.Port
	rp.user = cfg.User
	rp.pass = cfg.Pass
	rp.tries = cfg.Tries
	if rp.tries <= 0 {
		rp.tries = defaultTries
	}
	rp.retryDelay = retryDelay
	rp.maxReplySize = cfg.MaxReplySize
	if rp.maxReplySize == 0 {
		rp.maxReplySize = defaultMaxReplySize
	}
	return nil
}

// SetupEndpoint returns the setup error. It logs it as well, since its
// callers drop it, and keeps it for Send, which returns it on every call.
func SetupEndpoint(rn *RestEndpoint, cfg *jcom.RestEndpointCfg, logger *logrus.Entry) error {
	rn.rec = *cfg
	rn.l = logger.WithFields(logrus.Fields{"section": "rest"})
	if err := SetupRestProxy(&rn.rp, cfg, rn.l); err != nil {
		rn.rp.setupErr = err
		rn.l.Errorf("cannot set up the REST client for %v: %s", cfg.Addrs, err)
		return err
	}
	return nil
}
```

**One request and its reply**
([impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging),
[impl_plugin_rest_reply_size_limit](#impl_plugin_rest_reply_size_limit),
[impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes](#impl_plugin_rest_fix_transport_auth_nil_body_setup_and_string_crashes)).
`Send` (`pkg/rest/rproxy.go`) logs one Debug line for each reply, with no
body and no header, and reads no more of it than one byte past `maxreplysize`. A request
that gets no reply returns the kind of its failure from `transportError`,
which never reads the reply that did not come.

```go
// Send sends one request and reads its reply. It returns a RestError only
// when it has no reply to give back: no client, no reply, or one over
// maxreplysize. A reply of any status comes back with its body.
func (rp *RestProxy) Send(ctx context.Context, method string, path string, data interface{}, ok int) (int, []byte, RestError) {
	if rp.httpRestProxy == nil {
		msg := "REST client not configured"
		if rp.setupErr != nil {
			msg += ": " + rp.setupErr.Error()
		}
		return 0, nil, GetError(RestErrorRequestMalfunction, msg)
	}
	body := []byte("{}") // a GET carries it too (question 30)
	if data != nil {
		var err error
		if body, err = json.Marshal(data); err != nil {
			return 0, nil, GetError(RestErrorRequestMalfunction, "request body: "+err.Error())
		}
	}
	url := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addrs[rp.active_addr], rp.port, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, GetError(RestErrorRequestMalfunction, "request: "+err.Error())
	}
	req.SetBasicAuth(rp.user, rp.pass) // only once the request exists
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kubernetes CSI Plugin")

	res, err := rp.httpRestProxy.Do(req)
	if err != nil {
		return 0, nil, transportError(err) // res is nil: no reply came
	}
	defer res.Body.Close()
	// One line per reply: the method, the path with its query string and
	// the status; never a body or a header, which can hold a password.
	jcom.LFC(ctx).Debugf("%s %s: %d", method, path, res.StatusCode)

	// One byte past the limit tells a reply over it from one of exactly it.
	reply, err := io.ReadAll(io.LimitReader(res.Body, rp.maxReplySize+1))
	if err != nil {
		return res.StatusCode, nil, GetError(RestErrorRequestMalfunction, "reading the reply: "+err.Error())
	}
	if int64(len(reply)) > rp.maxReplySize {
		return res.StatusCode, nil, GetError(RestErrorRPM,
			fmt.Sprintf("reply larger than maxreplysize, %d bytes", rp.maxReplySize))
	}
	return res.StatusCode, reply, nil
}

// transportError names the kind of failure of a request that got no reply.
func transportError(err error) RestError {
	var certErr *tls.CertificateVerificationError
	var dnsErr *net.DNSError
	var netErr net.Error
	var opErr *net.OpError
	switch {
	case errors.As(err, &certErr):
		return GetError(RestErrorUnableToConnect, fmt.Sprintf(
			"certificate of the appliance not verified: %s; set cafile to the certificate or its CA, or set insecure",
			certErr.Err))
	case errors.As(err, &dnsErr):
		return GetError(RestErrorUnableToConnect, "address does not resolve: "+dnsErr.Error())
	case errors.As(err, &netErr) && netErr.Timeout(): // the client's own timeout too
		return GetError(RestErrorRequestTimeout, "request timed out: "+err.Error())
	case errors.As(err, &opErr) && opErr.Op == "dial":
		return GetError(RestErrorUnableToConnect, "connection refused or failed: "+opErr.Error())
	default:
		return GetError(RestErrorRequestMalfunction, "request failed: "+err.Error())
	}
}
```

**A GET answered without data**
([bug_plugin_rest_get_calls_decode_body_before_204_check](#bug_plugin_rest_get_calls_decode_body_before_204_check),
[stage 3](#stage-3-behaviour-and-error-mapping), [question 39](#decisions)).
`getWithData` (`pkg/rest/rproxy.go`) is the one way a GET is sent: it sends
a GET answered 204 again after `retrydelay`, `tries` times in all, and
returns the first other reply as if it had come first.

```go
// getWithData sends a GET of the plugin. A reply of 204 carries no data,
// so the same GET is sent again after retrydelay, tries times in all. A
// 204 is never turned into an empty object or an empty list: a caller
// would read those as a volume of size 0, a full pool or a snapshot
// without clones.
func (rp *RestProxy) getWithData(ctx context.Context, path string) (int, []byte, RestError) {
	for sent := 1; ; sent++ {
		status, body, err := rp.Send(ctx, "GET", path, nil, CodeOK)
		if err != nil || status != CodeNoContent {
			if err == nil && status == CodeOK && sent > 1 {
				// no recording shows a 204 on a GET: one made good leaves a trace
				jcom.LFC(ctx).Warnf("GET %s: status %d after %d replies of 204", path, status, sent-1)
			}
			return status, body, err
		}
		if sent >= rp.tries {
			return status, nil, GetError(RestErrorRPM,
				fmt.Sprintf("GET %s: status 204, no data, to all %d requests", path, sent))
		}
		time.Sleep(rp.retryDelay)
	}
}
```

**The error body**
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)).
`ErrorT` (`pkg/rest/payload-def.go`) decodes the error object of a failed
reply. Its `errno` keeps apart a number, `null` and a missing key, the three
cases a catalogue entry matches on; `ErrorFromErrorT` and `ErrorT.String()`, which read it as a pointer, go, since `failedReply` does their work.

```go
// errnoField is the errno of an error body: a number, null, or no key at
// all. Go's decoder reads a missing key as it reads null, so the field is
// decoded from its raw JSON, and present says whether the key came.
type errnoField struct {
	present bool
	null    bool
	value   int
}

// UnmarshalJSON runs only for a key that is in the body, null included.
func (e *errnoField) UnmarshalJSON(raw []byte) error {
	*e = errnoField{present: true}
	if string(raw) == "null" {
		e.null = true
		return nil
	}
	return json.Unmarshal(raw, &e.value)
}

func (e errnoField) String() string {
	switch {
	case !e.present:
		return "missing"
	case e.null:
		return "null"
	default:
		return strconv.Itoa(e.value)
	}
}

// The three cases, as a catalogue entry writes them.
var (
	errnoMissing = errnoField{}
	errnoNull    = errnoField{present: true, null: true}
)

func errnoIs(n int) errnoField { return errnoField{present: true, value: n} }

// ErrorT is the error object of a failed reply.
type ErrorT struct {
	Class   *string    `json:"class"`
	Errno   errnoField `json:"errno"`
	Message *string    `json:"message"`
	Url     *string    `json:"url"`
}

func (e *ErrorT) class() string {
	if e.Class == nil {
		return ""
	}
	return *e.Class
}

func (e *ErrorT) message() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
```

**A failed reply**
([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue),
[bug_plugin_rest_get_error_maps_body_only_ignores_http_status](#bug_plugin_rest_get_error_maps_body_only_ignores_http_status)).
`failedReply` (`pkg/rest/errors.go`) is the one place where a failure of a
call is mapped and logged: a request without a reply, a reply the call
cannot use, and a reply that is not the call's success. Without an error
object in the body, the status alone gives the code; with one, the first
entry of the call's catalogue that it matches gives it, and a reply that
matches none is Unknown, or "not found" for a 404. The catalogue shown is
that of the LUN detach as its scenarios file holds it at `jdss-rest-api-spec`
commit `434c9cd`: replies of one status and one class, without `errno`, that
only the message tells apart; the pool's entry comes first, since its message
matches the target's pattern too. A kind is written once, with the status,
class, errno and pattern it has in every scenarios file, and a call's
catalogue lists its kinds with the code each means for that call. Each entry gives two codes
([question 74](#decisions)): what the error means for the call, and the code
of its kind in the range of its layer, with its number from Table 19
([Measurements](#measurements)).

```go
// In errors.go.

// A RestError carries two codes. GetCode says what the error means for the
// call, and is what the callers compare with, as today. GetKindCode says
// which error the appliance sent: the code of its kind inside the range of
// its layer, the code of the layer alone for an error that no entry of the
// call holds, and 0 for a failure that is not an error of the appliance.
type RestError interface {
	Error() (out string)
	GetCode() int
	GetKindCode() int
}

type restError struct {
	code int
	kind int
	msg  string
}

func (err *restError) GetKindCode() int { return err.kind }

// GetError makes the error of a failure that is not an error of the
// appliance: its second code is 0. Every `&restError{code, msg}` of today
// becomes a call of it.
func GetError(c int, m string) RestError { return &restError{code: c, msg: m} }

// applianceError makes the error of a reply of the appliance.
func applianceError(c int, kind int, m string) RestError {
	return &restError{code: c, kind: kind, msg: m}
}

// The layers of the appliance's errors, each with a range of 1000 codes.
// The code of a layer is the first of its range; its kinds of error take
// the numbers after it.
const (
	LayerNoClass = 1000 // no class: login errors and some 400 and 404
	LayerPython  = 2000 // builtins.*: errors the appliance did not catch
	LayerZfs     = 3000 // zfslib.*: the ZFS wrapper
	LayerProduct = 4000 // opene.*: the product's own logic
	LayerSchema  = 5000 // jsonschema.*: a request body that fails its schema
	LayerRouting = 6000 // werkzeug.exceptions.*: routing and the API's own checks
)

// layerOf gives the code of the layer that the class of an error shows, and
// 0 for a class of no layer known.
func layerOf(class string) int {
	switch {
	case class == "":
		return LayerNoClass
	case strings.HasPrefix(class, "builtins."):
		return LayerPython
	case strings.HasPrefix(class, "zfslib."):
		return LayerZfs
	case strings.HasPrefix(class, "opene."):
		return LayerProduct
	case strings.HasPrefix(class, "jsonschema."):
		return LayerSchema
	case strings.HasPrefix(class, "werkzeug.exceptions."):
		return LayerRouting
	}
	return 0
}

// In error_catalogue.go, transcribed from the scenarios files.

// The kinds of error: one constant for each error model of the spec, under
// the model's name, numbered once inside the range of its layer and never
// renumbered. Table 19 lists them all; these three are those of this call.
const (
	KindItemNotFoundErrorTargetNotExists           = LayerProduct + 10 // 4010
	KindItemNotFoundErrorTargetPartitionNotExists  = LayerProduct + 11 // 4011
	KindItemNotFoundErrorVolumeNameIsNotAttachedTo = LayerProduct + 12 // 4012
)

// What a reply of each kind looks like. In the scenarios files a kind has
// one status, one class, one errno and one pattern whatever the call.
var (
	kindItemNotFoundErrorTargetNotExists = errorKind{KindItemNotFoundErrorTargetNotExists, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Target [\s\S]* not exists\.$`)}
	kindItemNotFoundErrorTargetPartitionNotExists = errorKind{KindItemNotFoundErrorTargetPartitionNotExists, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Target partition: [\s\S]* not exists\.$`)}
	kindItemNotFoundErrorVolumeNameIsNotAttachedTo = errorKind{KindItemNotFoundErrorVolumeNameIsNotAttachedTo, 404, "opene.exceptions.ItemNotFoundError", errnoMissing,
		regexp.MustCompile(`^Volume name [\s\S]* is not attached to target [\s\S]*$`)}
)

// The catalogue of DettachVolumeFromTarget,
// DELETE pools/<pool>/san/iscsi/targets/<target>/luns/<zvol>: its kinds,
// each with the code it means for this call.
var dettachVolumeFromTargetErrors = []errorEntry{
	// the pool's entry first: its message matches the target's pattern too
	kindItemNotFoundErrorTargetPartitionNotExists.means(RestErrorResourceDNEPool), // new code: the pool is missing
	kindItemNotFoundErrorTargetNotExists.means(RestErrorResourceDNETarget),        // nothing is left to do
	kindItemNotFoundErrorVolumeNameIsNotAttachedTo.means(RestErrorResourceDNELun), // new code: the LUN is gone; the caller deletes the target next
}

// In errors.go.

// errorKind is one error model of the spec: how a reply of that kind looks.
type errorKind struct {
	kind    int
	status  int
	class   string // "" where the spec lists the class as absent
	errno   errnoField
	message *regexp.Regexp // nil where the spec gives the kind no pattern
}

// means makes the entry of a call's catalogue for this kind of error: code
// is what the error means for that call.
func (k errorKind) means(code int) errorEntry {
	return errorEntry{status: k.status, class: k.class, errno: k.errno, message: k.message, code: code, kind: k.kind}
}

// What an entry that is not in the spec puts where no recording gives the
// status or the class of its reply.
const (
	anyStatus = 0
	anyClass  = "*"
)

// errorEntry is one error a call can get. A reply matches it on the status,
// the class, the errno in its three cases and the message pattern of the
// error's schema.
type errorEntry struct {
	status  int
	class   string
	errno   errnoField
	message *regexp.Regexp // nil: any message
	code    int            // what the error means for this call
	kind    int            // which error the appliance sent; 0: not in the spec
}

func (en errorEntry) matches(status int, e *ErrorT) bool {
	if en.status != anyStatus && en.status != status {
		return false
	}
	if en.class != anyClass && en.class != e.class() {
		return false
	}
	if en.errno != e.Errno {
		return false
	}
	if en.message == nil {
		return true
	}
	return e.Message != nil && en.message.MatchString(*e.Message)
}

// In targets.go.

// DettachVolumeFromTarget ends as every call ends: in its success, or in
// failedReply with the call's catalogue. apiPath escapes each name and
// refuses an empty one, "." and ".."; request sends and checks the status.
func (s *RestEndpoint) DettachVolumeFromTarget(ctx context.Context, pool string, tname string, vname string) RestError {
	path := apiPath("/pools/%s/san/iscsi/targets/%s/luns/%s", pool, tname, vname)
	return s.request(ctx, "DELETE", path, nil, dettachVolumeFromTargetErrors, CodeNoContent)
}

// In storageapi.go: how a request whose reply holds nothing to read ends.
func (s *RestEndpoint) request(ctx context.Context, method string, p requestPath, data interface{},
	entries []errorEntry, ok ...int,
) RestError {
	if p.err != nil { // a name that cannot be a part of a path
		return s.failedReply(ctx, method, p.path, nil, 0, nil, p.err)
	}
	status, body, err := s.rp.Send(ctx, method, p.path, data, ok[0])
	if err == nil && isOneOf(status, ok) {
		return nil
	}
	return s.failedReply(ctx, method, p.path, entries, status, body, err)
}

// In errors.go.

// failedReply maps a failure of a call to one RestError and writes its one
// log line in pkg/rest: at Debug for an error the call's catalogue knows, at
// Error for an entry that says Unknown on purpose and for anything else.
// err is an error found before the reply was read as a failure: no reply,
// a reply over maxreplysize, no data to tries GETs, a reply that does not
// decode, or a name that cannot go into a path.
func (s *RestEndpoint) failedReply(ctx context.Context, method, path string, entries []errorEntry,
	status int, body []byte, err RestError) RestError {
	l := jcom.LFC(ctx).WithFields(logrus.Fields{"section": "rest", "method": method, "path": path, "status": status})
	// the log line of a failure holds both of its codes
	withCodes := func(re RestError) *logrus.Entry {
		return l.WithFields(logrus.Fields{"code": re.GetCode(), "kind": re.GetKindCode()})
	}
	if err != nil {
		withCodes(err).Error(err.Error())
		return err
	}

	var reply struct {
		Error *ErrorT `json:"error"`
	}
	if json.Unmarshal(body, &reply) != nil || reply.Error == nil {
		// not a JSON error body: the status alone gives the code
		err = statusError(method, path, status, body)
		withCodes(err).Error(err.Error())
		return err
	}

	e := reply.Error
	l = l.WithFields(logrus.Fields{"class": e.class(), "errno": e.Errno.String(), "message": e.message()})
	for _, en := range entries { // the first entry that matches decides
		if en.matches(status, e) {
			kind := en.kind
			if kind == 0 {
				// not in the spec: the code of the layer alone
				kind = layerOf(e.class())
			}
			// the message on one line: some list datasets on lines of their own
			err = applianceError(en.code, kind, strings.Join(strings.Fields(e.message()), " "))
			if en.code == RestErrorFailureUnknown {
				withCodes(err).Error("error of the catalogue, unknown on purpose")
			} else {
				withCodes(err).Debug("error of the catalogue")
			}
			return err
		}
	}
	code := RestErrorFailureUnknown
	if status == 404 {
		code = RestErrorResourceDNE // a 404 means not found, whatever its body
	}
	// no entry of the call holds it: its second code is that of its layer alone
	err = applianceError(code, layerOf(e.class()), fmt.Sprintf("%s %s: status %d, class %q, errno %s, message %q",
		method, path, status, e.class(), e.Errno, e.message()))
	withCodes(err).Error("error not in the catalogue of the call")
	return err
}

// statusError maps a reply without a JSON error body by its status alone,
// and keeps the first bytes of the body in the message. Such a reply has no
// class, so its second code is that of the layer without a class.
func statusError(method, path string, status int, body []byte) RestError {
	if len(body) > 200 {
		body = body[:200]
	}
	switch status {
	case 401, 403:
		return applianceError(RestErrorAuthFailed, LayerNoClass, // new code
			fmt.Sprintf("%s %s: authentication failed, status %d: %q", method, path, status, body))
	case 404:
		return applianceError(RestErrorResourceDNE, LayerNoClass,
			fmt.Sprintf("%s %s: not found, status 404: %q", method, path, body))
	default:
		return applianceError(RestErrorFailureUnknown, LayerNoClass,
			fmt.Sprintf("%s %s: status %d without an error body: %q", method, path, status, body))
	}
}
```

**A reply type**
([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec),
[impl_plugin_rest_return_parse_errors](#impl_plugin_rest_return_parse_errors)).
`ResourceVolume` is shown as transcribed into `pkg/rest/types.go`; its checks, its accessors and the decoders, `decodeData` for a reply and `decodeObject` for an entry of a list, go into `types_methods.go`, and `GetVolume` stays in `storageapi.go`. A required read field stays a plain field, since
the drivers and the controller read it so, and `checkRequired` finds it
missing among the keys of the reply; an optional one is a pointer, and a
field that nothing reads is raw JSON.

```go
// In types.go: the reply of GET pools/<pool>/volumes/<volume>, from the
// path file pools/+poolname+/volumes/+volumename+.yaml at the commit of
// the header. The comment on each read field gives its mark.
type ResourceVolume struct {
	Name          string  `json:"name"`            // required
	VolSize       string  `json:"volsize"`         // required: a number sent as a string
	Origin        *string `json:"origin"`          // optional: null on a zvol
	Guid          string  `json:"guid"`            // required, read from stage 3
	DefaultScsiId string  `json:"default_scsi_id"` // required, read from stage 3
	IsClone       *bool   `json:"is_clone"`        // optional (question 65), read from stage 3

	// Not read: kept as sent, so that no type of theirs can fail.
	Encryption json.RawMessage `json:"encryption"`
	Used       json.RawMessage `json:"used"`
	// ... every other property of the reply schema, the same way

	size int64 // VolSize as a number, set by checkRequired
}

// In types_methods.go.

// replyType checks the required read fields of a reply type against the
// keys of the JSON object it was decoded from.
type replyType interface {
	checkRequired(keys map[string]json.RawMessage) error
}

// decodeData decodes the data object of a reply and checks it. An entry of
// a list, and an item of a clones reply, go through decodeObject the same
// way.
func decodeData(body []byte, v replyType) RestError {
	var reply struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	return decodeObject(reply.Data, v)
}

func decodeObject(raw json.RawMessage, v replyType) RestError {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		return GetError(RestErrorRPM, "reply data is not an object: "+err.Error())
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return GetError(RestErrorRPM, "reply does not decode: "+err.Error())
	}
	if err := v.checkRequired(keys); err != nil {
		return GetError(RestErrorRPM, err.Error())
	}
	return nil
}

// requireKeys fails on the first of names that the object lacks or holds
// as null, in the fixed text that the live test searches the log for.
func requireKeys(keys map[string]json.RawMessage, names ...string) error {
	for _, n := range names {
		if raw, ok := keys[n]; !ok || string(raw) == "null" {
			return fmt.Errorf("required reply field %s is missing", n)
		}
	}
	return nil
}

// parseNumber converts a number sent as a string. Its error names the
// field, in a fixed text too, and the value is never read as zero.
func parseNumber(field, s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("reply field %s does not convert: %q", field, s)
	}
	return n, nil
}

func (v *ResourceVolume) checkRequired(keys map[string]json.RawMessage) (err error) {
	if err = requireKeys(keys, "name", "volsize", "guid", "default_scsi_id"); err != nil {
		return err
	}
	v.size, err = parseNumber("volsize", v.VolSize)
	return err
}

func (v *ResourceVolume) GetSize() int64 { return v.size }

// OriginSnapshot returns "" for a volume that is not a clone, as today.
func (v *ResourceVolume) OriginSnapshot() string {
	if v.Origin == nil || !originNameRegexp.MatchString(*v.Origin) {
		return ""
	}
	m := originNameRegexp.FindStringSubmatch(*v.Origin)
	return m[originNameRegexp.SubexpIndex("snapshot")]
}

// In storageapi.go.

// readObject is how every GET of one object runs: getWithData, decodeData
// for the reply, failedReply for a failure, a reply that does not decode
// among them.
func readObject[T any, PT interface {
	*T
	replyType
}](ctx context.Context, s *RestEndpoint, p requestPath, entries []errorEntry, ok ...int) (*T, RestError) {
	if p.err != nil { // a name that cannot be a part of a path
		return nil, s.failedReply(ctx, "GET", p.path, nil, 0, nil, p.err)
	}
	status, body, err := s.rp.getWithData(ctx, p.path)
	if err == nil && isOneOf(status, ok) {
		v := new(T)
		if err = decodeData(body, PT(v)); err == nil {
			return v, nil
		}
	}
	return nil, s.failedReply(ctx, "GET", p.path, entries, status, body, err)
}

func (s *RestEndpoint) GetVolume(ctx context.Context, pool string, vname string) (*ResourceVolume, RestError) {
	path := apiPath("/pools/%s/volumes/%s", pool, vname) // each name escaped and checked
	return readObject[ResourceVolume](ctx, s, path, getVolumeErrors, CodeOK)
}
```

**The read of one list**
([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
[the `per_page` commit](#before-stage-1-send-per_page-on-every-list)).
`readList` goes into `pkg/driver/common.go`, in place of
`getResourcesList`, and `pageQuery` into `pkg/rest/entrielisters.go`, in
place of `pagedcSuffix`. `readList` asks for page after page, from page 0,
and ends on the first page that is not full, an empty one included;
`pageQuery` writes the query of each request. The entry type, the request
for a page and the fields compared differ between the lists, so they are
parameters of the paging functions.

```go
// perPage is the page size of every list request.
const perPage = 100

// pageQuery adds the paging to the path of a list: no `_dc`.
func pageQuery(path string, page, perPage int) string {
	return fmt.Sprintf("%s?page=%d&per_page=%d", path, page, perPage)
}

// readList reads one appliance list from page 0 to its last page.
// getPage returns one page as the appliance sent it: the stop rule
// counts a page before any entry is dropped from it.
func readList[T any](ctx context.Context,
	getPage func(ctx context.Context, page, perPage int) ([]T, jrest.RestError),
) ([]T, jrest.RestError) {
	var read []T
	for page := 0; ; page++ {
		ents, err := getPage(ctx, page, perPage)
		if err != nil {
			return nil, err
		}
		read = append(read, ents...)
		if len(ents) < perPage {
			// an empty page, or the last one, which is not full
			return read, nil
		}
	}
}
```

**The merge of the pages**
([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
[question 65](#decisions)). `mergePages` (`pkg/driver/common.go`) settles a
name that one read returned twice. `same` compares the fields question 65 names for the list
at hand, `is_clone` through `sameOptional`; for the dataset list it always
answers true.

```go
// sameOptional reports whether two optional values agree: both are
// missing, or both hold the same value.
func sameOptional[V comparable](a, b *V) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// mergePages takes the entries of one read, in the order of the read.
// Of two entries of one name that same counts as one object, the one
// read first is dropped; two that differ both stay.
func mergePages[T any](read []T, name func(T) string, same func(a, b T) bool) []T {
	dropped := make([]bool, len(read))
	staying := make(map[string][]int) // name -> its entries that stay
	for i, e := range read {
		var differ []int
		for _, j := range staying[name(e)] {
			if same(read[j], e) {
				dropped[j] = true // the later entry holds the newer state
			} else {
				differ = append(differ, j)
			}
		}
		staying[name(e)] = append(differ, i)
	}
	merged := make([]T, 0, len(read))
	for i, e := range read {
		if !dropped[i] {
			merged = append(merged, e)
		}
	}
	return merged
}
```

**The order and the cut**
([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
[question 66](#decisions)). `listEntries` (`pkg/driver/common.go`) runs the
steps of one CSI list call in their order, and `cutAfter` picks the entries
the call returns and, when entries follow them, the token of the next call,
whose `Token()` the caller writes into `next_token`.

```go
// The key of an entry is its place in the order of its list, one string
// compared byte by byte: the name for volumes and for the snapshots of one
// volume; for a list of all snapshots, the parent's name, "!" and the
// snapshot's name. "!" sorts before every character a name of the plugin
// holds, so the order is by parent, then by name.
func snapshotsListKey(parent, name string) string { return keyPart(parent) + "!" + keyPart(name) }

// keyPart writes a name into a key. A name the plugin made holds letters,
// digits, "-", "_" and "=" and is written as it is. Any other byte, which
// only a name made by hand can hold, is written as "%" and two hex digits,
// so that a key, and the token made from it, is printable ASCII whatever
// the appliance holds.
func keyPart(name string) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		// printable ASCII but "!", which parts the two names of a key, "%",
		// which starts an escaped byte, and "~", the last printable character
		if c := name[i]; c > '!' && c < '~' && c != '%' {
			out = append(out, c)
		} else {
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}

// cutAfter sorts the entries by key and returns the first limit of them
// from the place the token gives, all of them when limit is 0; limit is
// never negative here. When entries follow those it returns, next is the
// token of the call after it, else nil.
func cutAfter[T any](ents []T, key func(T) string, from *CSIListingToken, limit int) (out []T, next *CSIListingToken) {
	// stable: two entries of one key keep the order of the read
	sort.SliceStable(ents, func(i, j int) bool { return key(ents[i]) < key(ents[j]) })
	keys := make([]string, len(ents))
	for i, e := range ents {
		keys[i] = key(e)
	}
	rest := ents[from.start(keys):]
	if limit == 0 || len(rest) <= limit {
		return rest, nil
	}
	out = rest[:limit]
	last := key(out[limit-1])
	// When the cut falls between two entries of one key, the second is not
	// returned: the next call starts after that key.
	for _, e := range rest[limit:] {
		if k := key(e); k > last {
			return out, from.nextCall(last, k)
		}
	}
	return out, nil
}

// listEntries answers one CSI list call. It reads the whole list once,
// whatever the token. own drops what the reply does not carry, a name
// the plugin did not make (with its warning) and the other driver's
// snapshots, before the cut, so that limit counts entries of the reply.
func listEntries[T any](ctx context.Context,
	getPage func(ctx context.Context, page, perPage int) ([]T, jrest.RestError),
	name func(T) string, same func(a, b T) bool, own func(T) bool,
	key func(T) string, from *CSIListingToken, limit int,
) (out []T, next *CSIListingToken, err jrest.RestError) {
	read, err := readList(ctx, getPage)
	if err != nil {
		return nil, nil, err
	}
	var ents []T
	for _, e := range mergePages(read, name, same) {
		if own(e) {
			ents = append(ents, e)
		}
	}
	out, next = cutAfter(ents, key, from, limit)
	return out, next, nil
}
```

**The token**
([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name),
[question 68](#decisions)). `CSIListingToken` (`pkg/driver/snaptoken.go`)
holds the list a token was issued for and a bound: the shortest string that
sorts after the key of the last entry returned, or is equal to it, and
before the key of the next one; the next call starts after the bound. Where
the bound does not fit in CSI's 128 bytes, the token holds its first bytes
and a hash of each of the two keys. A token that does not parse, has
another version or belongs to another list is refused, and the controller
answers it with `Aborted`, as it does today.

```go
const (
	tokenVersion = "1" // a token of another version, today's among them, is refused
	tokenMax     = 128 // bytes: CSI's limit for a string field
	hashLen      = 11  // base64url of the first 64 bits of a SHA-256
)

// shortHash names a string in hashLen ASCII characters.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(sum[:8])
}

// ListID names the list a token is issued for: the volumes, all
// snapshots, or the snapshots of one volume, for which a hash of the
// volume's name stands, so that the name does not take room in the token.
type ListID struct{ tag string }

// Tag returns the name of the list as a token carries it.
func (l ListID) Tag() string { return l.tag }

func VolumesList() ListID                      { return ListID{"v"} }
func SnapshotsList() ListID                    { return ListID{"s"} }
func VolumeSnapshotsList(volume string) ListID { return ListID{"w" + shortHash(volume)[:8]} }

// CSIListingToken is where a CSI list call resumes: after bound, or, when
// the bound did not fit, by the hashes of the last key returned and of the
// next, with prefix, the bound's first bytes.
type CSIListingToken struct {
	list               ListID
	bound              string
	prefix             string
	hashLast, hashNext string // with prefix only
}

// bound returns the shortest string b with last <= b < next, for
// last < next: the next call starts after it.
func bound(last, next string) string {
	i := 0
	for i < len(last) && i < len(next) && last[i] == next[i] {
		i++
	}
	switch {
	case i == len(last): // last is the beginning of next
		return last
	case last[i]+1 < next[i]: // a byte between the two fits
		return last[:i] + string([]byte{last[i] + 1})
	case i+1 < len(next): // next[i] follows last[i]: the beginning of next
		return next[:i+1]
	}
	// next ends on the byte that follows last[i]: the bound keeps last[i],
	// and is made greater than last at the first later byte that has room
	for j := i + 1; j < len(last); j++ {
		if last[j] < 0xFF {
			return last[:j] + string([]byte{last[j] + 1})
		}
	}
	return last
}

// nextCall is the token of the call after one that returned up to the key
// last, when next is the key of the entry that follows it.
func (t *CSIListingToken) nextCall(last, next string) *CSIListingToken {
	n := &CSIListingToken{list: t.list, bound: bound(last, next)}
	room := tokenMax - len(tokenVersion) - len(t.list.tag) - 1
	if len(n.bound) > room {
		// the two keys share a beginning too long for the token
		n.prefix = n.bound[:room-2*hashLen]
		n.hashLast, n.hashNext = shortHash(last), shortHash(next)
		n.bound = ""
	}
	return n
}

// start is the place, in keys sorted in byte order, of the first entry
// the call returns.
func (t *CSIListingToken) start(keys []string) int {
	after := func(b string) int {
		return sort.Search(len(keys), func(i int) bool { return keys[i] > b })
	}
	switch {
	case t == nil:
		return 0 // no token: the whole list, as the cleanup reads it
	case t.bound != "":
		return after(t.bound)
	case t.prefix == "":
		return 0 // the first call of a walk
	}
	for i, k := range keys {
		if shortHash(k) == t.hashNext {
			return i
		}
	}
	for i := len(keys) - 1; i >= 0; i-- {
		if shortHash(keys[i]) == t.hashLast {
			return i + 1
		}
	}
	// both entries are gone: start after the beginning they shared, so that
	// no entry is missed; the entries that share it come again (risk 29)
	return after(t.prefix)
}

// Token writes the next_token, ASCII of at most tokenMax bytes: the
// version, the list, then "." and the bound, or "~", the two hashes and
// the prefix.
func (t *CSIListingToken) Token() string {
	head := tokenVersion + t.list.tag
	if t.prefix != "" {
		return head + "~" + t.hashLast + t.hashNext + t.prefix
	}
	return head + "." + t.bound
}

// NewCSIListingTokenFromTokenString reads the starting_token of a call on
// list, "" for the first call of a walk. No input makes it panic.
func NewCSIListingTokenFromTokenString(ts string, list ListID) (*CSIListingToken, jrest.RestError) {
	t := &CSIListingToken{list: list}
	if ts == "" {
		return t, nil
	}
	refused := jrest.GetError(jrest.RestErrorArgumentIncorrect,
		fmt.Sprintf("token %q was not issued for this list", ts))
	head := tokenVersion + list.tag
	if len(ts) > tokenMax || len(ts) < len(head)+2 || ts[:len(head)] != head {
		return nil, refused
	}
	for i := 0; i < len(ts); i++ {
		if ts[i] < '!' || ts[i] > '~' { // ASCII, no space, no control
			return nil, refused
		}
	}
	rest := ts[len(head)+1:]
	switch ts[len(head)] {
	case '.':
		t.bound = rest
	case '~':
		if len(rest) <= 2*hashLen {
			return nil, refused
		}
		t.hashLast, t.hashNext, t.prefix = rest[:hashLen], rest[hashLen:2*hashLen], rest[2*hashLen:]
	default:
		return nil, refused
	}
	return t, nil
}
```

**A snapshot deleted only without clones**
([impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow](#impl_plugin_volume_snapshot_lists_decode_short_entries_and_callers_follow),
[stage 3](#stage-3-behaviour-and-error-mapping), [question 34](#decisions),
[risk 23](#risks)). `deleteSnapshotWithoutClones`
(`pkg/driver/restwrappers.go`) reads the clones of a snapshot first and
deletes it only when there is none. The cleanups of both drivers delete an
intermediate snapshot through it, and so does the NFS
`deleteIntermediateSnapshot` (`nfs_driver.go`), shown here.

```go
// In restwrappers.go.

// deleteSnapshotWithoutClones deletes a snapshot only after the appliance
// has answered that it has no clones, since the appliance deletes a NAS
// snapshot that has a clone, and the clone with it. When the clones cannot
// be read, nothing is deleted and the error is returned.
func deleteSnapshotWithoutClones(ctx context.Context,
	clonesOf func(context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError),
	deleteSnapshot func(context.Context) jrest.RestError,
) (deleted bool, clones []jrest.ResourceVolumeSnapshotClones, err jrest.RestError) {
	if clones, err = clonesOf(ctx); err != nil {
		return false, nil, err
	}
	if len(clones) > 0 {
		return false, clones, nil
	}
	if err = deleteSnapshot(ctx); err != nil {
		return false, nil, err
	}
	return true, nil, nil
}

// In nfs_driver.go. With a clone listed, the function goes on as it does
// today after a refused delete.
func (d *CSINFSDriver) deleteIntermediateSnapshot(ctx context.Context, pool string, vds string, sds string) jrest.RestError {
	deleted, clones, err := deleteSnapshotWithoutClones(ctx,
		func(ctx context.Context) ([]jrest.ResourceVolumeSnapshotClones, jrest.RestError) {
			return d.re.GetNASVolumeSnapshotClones(ctx, pool, vds, sds)
		},
		func(ctx context.Context) jrest.RestError {
			return d.re.DeleteNASSnapshot(ctx, pool, vds, sds)
		})
	switch {
	case err != nil || deleted:
		return err
	case len(clones) == 1 && clones[0].Name == sds:
		// the volume this snapshot was made for exists already
		return jrest.GetError(jrest.RestErrorResourceExists,
			fmt.Sprintf("NAS volume %s made from volume %s already exists", sds, vds))
	case len(clones) == 1:
		return jrest.GetError(jrest.RestErrorResourceBusySnapshotHasClones,
			fmt.Sprintf("intermediate snapshot %s of volume %s has the clone %s", sds, vds, clones[0].Name))
	default:
		return jrest.GetError(jrest.RestErrorStorageFailureUnknown,
			fmt.Sprintf("intermediate snapshot %s of volume %s has %d clones", sds, vds, len(clones)))
	}
}
```

### ID sources
[ID_sources](#id-sources)

**Table 20 — Where the names and IDs come from** · `tbl_id_sources`

| Names and IDs | Meaning | Source |
|---|---|---|
| `gap_…`, `bug_…`, `crash_…`, `security_…`, `rule_…`, `aopt_…`, `impl_…` | the findings, gaps, compliance rules, architectural options and implementation items of this document, each a heading in the [Options considered](#options-considered) section, the [What changes](#what-changes) section or this Appendix; the form is `kind_area_slug` | this document; the old letter-and-number IDs are mapped to them in `kub-testing/issues/`. A name in code font without a link, the `dataloss_…` names among them, has no heading here and is described in `kub-testing/issues/` |
| C, D, K, N, R, S, X + number | the code reviews' own IDs, kept in the reviews and the tracker | [`review/2026-10-02-code-review.md`](../../review/2026-10-02-code-review.md) and [`review/2026-11-07-code-review.md`](../../review/2026-11-07-code-review.md); mapped to names in `kub-testing/issues/` |
| F + number | the test findings' IDs, kept in the tracker | `kub-testing/ISSUES.md`; mapped to names in `kub-testing/issues/test-findings.md` |
| D1-D7 in 0003 | design 0003's own items, not review IDs | [0003](../design/0003-controller-crash-fixes-design.md) |
| hurl Finding n, C13, C24 | the hurl suite's findings and caveats | `jdss-rest-api-spec`'s own notes |
| backlog item n | the items of the plugin's review backlog | [`review/backlog.md`](../../review/backlog.md) |
| DEC10, DOC27, INF8, MF15, MK11, T9 to T13, TOOL10 | kub-testing items: decisions for the user, plugin documentation, lab infrastructure, manifests, Makefile, testing, tooling | `kub-testing/ISSUES.md` |

### Terms
[Terms](#terms)

**Table 21 — Terms** · `tbl_terms`

| Term | Meaning |
|---|---|
| jdss1 | the hurl suite's JovianDSS 1 appliance ([Table 12](#the-jdss-rest-api-spec-repository)) |
| jdss2 | the hurl suite's JovianDSS 2 appliance ([Table 12](#the-jdss-rest-api-spec-repository)) |
| lab appliance | the JovianDSS 1 appliance the Kubernetes lab and csi-sanity use ([Testing](#testing)) |
| hurl, hurl suite | the HTTP tests in `jdss-rest-api-spec/empirical/hurl/` that send requests to real appliances and record the replies |
| run | one execution of the hurl suite on one appliance; its folder in `hurl/out/` is named by date and time |
| csi-sanity | the standard CSI conformance test suite, run against the plugin in the lab |
| csi-test | the repository of csi-sanity, where its specs live |
| TypeSpec | a language for describing an API, compiled to OpenAPI; `jdss-rest-api-spec` writes the observed API in it, one file per method |
| preset | a group of csi-sanity tests run together, such as ctrl-A, ctrl-B, ctrl-cap or node |
| cycle | one `dev_cycle.py` run: build the plugin, deploy it to the lab, run the csi-sanity presets. Named by date and time, e.g. `20261102-140341` |
| lab machine | the machine where the plugin is built, deployed and unit-tested (`dev_cycle.py` runs there) |
| analysis machine | the machine with `jdss-rest-api-spec` and the hurl recordings, where the audit scripts run and the spec is read; it has no Go |
| company spec, empirical spec | the API as Open-E documents it, and the API as the appliances behaved in the hurl runs ([Table 11](#the-jdss-rest-api-spec-repository)) |
| decode | turn a JSON reply into a Go struct (Go's `encoding/json`) |
| reply type | the Go struct in `pkg/rest/types.go` that one JSON reply of the REST API is decoded into, such as `ResourceVolume` for the reply about one volume; the plugin reads its fields, such as the volume's name and size |
| zvol | a ZFS block volume; the plugin uses it for iSCSI volumes |
| NAS volume | a ZFS file system (dataset); the plugin uses it for NFS volumes |
| errno, class, message | three of the four fields of a JovianDSS error reply, `url` being the fourth: `{"error": {"errno": …, "class": …, "message": …, "url": …}}` |
| layer | the part of the appliance that raised an error, which the beginning of the error's `class` shows: no class, `builtins`, `zfslib`, `opene`, `jsonschema` or `werkzeug`; each layer has a range of 1000 `RestError` codes ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) |
| `RestError` code | the plugin's own code for a REST failure (`errors.go`), e.g. `RestErrorResourceDNE` ("does not exist"); the controller turns it into a CSI status. It says what a failure means for the call; an error of the appliance carries a second code beside it, of its kind in the range of its layer. Unknown, written without code font, is the code of an error the plugin does not recognise (`RestErrorFailureUnknown` today); `Unknown` in code font is the CSI status |
| `Internal`, `Unknown`, `AlreadyExists`, `ResourceExhausted`, … | CSI (gRPC) status codes the plugin returns to Kubernetes |
| repeated request | CSI may send the same create or delete again; the plugin must treat "already exists" on a create, when what exists is what the request asks for, and "not found" on a delete, as success |
| page, `per_page`, `results` | the appliance's paging of a list: a request with `page` and `per_page` gets at most `per_page` entries, from entry `page × per_page` on, of the whole list; `results` is the length of the whole list ([How lists are paged](#how-lists-are-paged)) |
| `max_entries`, `starting_token`, `next_token` | the paging of the CSI list calls: a call returns at most `max_entries` entries and, when more follow, a `next_token`, which the caller sends back as `starting_token` for the next ones ([impl_plugin_csi_list_paging_resumes_after_last_name](#impl_plugin_csi_list_paging_resumes_after_last_name)) |
| walk | the CSI list calls of one listing, from the call without a token through each `next_token` to the last |
| request log | the line the plugin writes at Debug for each reply to a REST request once [impl_plugin_rest_no_request_logging](#impl_plugin_rest_no_request_logging) is in the build: method, path with its query string, and status; no body and no header. A failed call also gets the error handler's own line, with `class`, `errno` and message ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)) |
| scenarios file | the `<METHOD>-scenarios.json` file of a call in `jdss-rest-api-spec`, which lists every error recorded for the call: its status, `class`, `errno` and a pattern for its message |
| fixture | a recorded reply saved in the plugin repository as test data for unit tests |
| transcribed types | the reply types in `pkg/rest/types.go`, written by an agent from `jdss-rest-api-spec`'s tracked OpenAPI of the recorded API; the file's header names the `jdss-rest-api-spec` commit ([impl_plugin_rest_reply_types_transcribed_from_observed_typespec](#impl_plugin_rest_reply_types_transcribed_from_observed_typespec)) |
| stage 1-4 | the four stages of [What changes](#what-changes); the commits before stage 1 are outside them, and the supporting work is not a stage of its own: its items land in the stages that Table 8 gives them; not the CSI NodeStage call |
| `vp_`, `vb_`, `vs_`; `sp_`, `sb_`, `ss_` | the prefixes of the names the plugin gives volumes (`v`) and snapshots (`s`) on the appliance: `p` for the CSI name as it is, `b` for its base32 form, `s` for a SHA-256 hash of it ([impl_plugin_driver_validate_names_and_ids](#impl_plugin_driver_validate_names_and_ids)) |
| short entry, full snapshot | the two shapes in which the appliance describes a snapshot: the short entry of a list (`name`, `parent`, `properties`), and the full object of a read by name, with `clones` when it has any and, for a zvol snapshot, `volsize` |
| intermediate snapshot | a snapshot the plugin creates on its own when it makes a volume from another volume; the new volume is a clone of that snapshot, and `cleanIntermediateSnapshots` removes those without clones when the volume they were taken of is deleted |
| handler | the one function in `pkg/rest` that every failed call passes through, whether the appliance answered with an error or no reply arrived (a transport error): it matches the reply against the error catalogue's entries for that call and returns the code of the entry it matches, or, for a reply that matches none, an unknown error, or the general not-found code for a 404, with all the reply's fields kept; and it writes the one log line of the REST module for that failure ([impl_plugin_rest_error_catalogue](#impl_plugin_rest_error_catalogue)). Today the work is spread over `getError`, `ErrorFromErrorT` and several call functions. The handler of a GET answered 204 is another function: the one every GET passes through, which sends such a GET again ([question 39](#decisions)) |
| transport error | a request that gets no HTTP reply at all: the address does not resolve, the connection is refused, the appliance's certificate does not verify, or the request times out ([crash_plugin_rest_nil_response_on_transport_error](#crash_plugin_rest_nil_response_on_transport_error)) |
| Basic auth, Bearer token | two ways to log in to the REST API: Basic sends the user name and password with every request; a Bearer token is obtained once by logging in and sent instead of the password until it expires |
| CHAP | the password login of iSCSI between a node and a target; the plugin does not use it today |
| initiator, target, LUN | iSCSI terms: the initiator is the client (here the Kubernetes node), the target is what the appliance exposes, and a LUN is one volume attached to a target |
| static PersistentVolume | a PersistentVolume an administrator writes by hand, naming an existing volume, instead of one the plugin creates on request |
| `VolumeSnapshotContent` | the Kubernetes object that names a snapshot on the storage; like a static PersistentVolume, it can be written by hand |
| sidecar, attacher, snapshotter | the helper containers of Kubernetes CSI that run beside the controller in its pod and call it: the attacher for attaching volumes to nodes, the snapshotter for snapshots |

