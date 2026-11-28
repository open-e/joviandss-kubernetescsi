# Improvement backlog

Items to review and decide on later. Collected on 2026-10-17 while fixing the
module import paths on `master` (`596f5ef`). Nothing here is implemented yet.
Line numbers refer to `master` at that time.

## Build and Makefile

1. Remove `go mod tidy` and `go get ./app/joviandssplugin` from the build
   targets (`joviandss`, `cli`). Builds should not edit `go.mod`/`go.sum`,
   download new code or need network.
2. Add explicit dependency targets: `tidy` (`go mod tidy`, run by hand, review
   and commit the diff) and `deps` (`go mod download` + `go mod verify`).
3. Add check targets: `vet` (`go vet ./...`), `fmt-check` (`gofmt -l .`),
   `check-tidy` (`go mod tidy -diff`, Go >= 1.23), grouped as `check`.
4. Keep the module name in one variable (`MODULE`) and use it in the `-X`
   flag; today the path is written out in place.
5. `cli` target builds without the `-X ...Version` flag.
6. Drop `-a` (forces full rebuild; the Go build cache handles it).
7. `VERSION`: consider `git describe --long --tags --always --dirty`.
8. `clean` target is broken (`rm -rf _outpusudo podman push ...`).
9. `.PHONY` lists targets that do not exist (`hostpath-container`, `iscsi`,
   `rest`).
10. Reconcile with the `nfs-support` branch `Makefile` (`dev`/`prod`,
    `PROTOCOL_TYPE`), which `doc/compile.md` already describes.

## Go module

11. `go` directive: `go.mod` says `go 1.19`. Raising it (e.g. to 1.26)
    changes language semantics (per-iteration loop variables since 1.22).
    Decide deliberately, not as a side effect of `go mod tidy`.
12. `go.sum`: check whether `HEAD`'s `go.sum` is complete
    (`go-cmp v0.5.9/go.mod` line was added by tidy on 2026-10-17).
13. Run `gofmt -w .` after the import path change (import order).
14. Ten commented-out imports still use the old short path
    (`joviandss-kubernetescsi/...`); remove or update.
15. Dependencies are from 2024 (`grpc`, `protobuf`, `x/net v0.24.0`):
    check `go list -m -u all` and `govulncheck ./...`.

## Code

16. Nil-dereference crashes listed in `review/2026-10-02-code-review.md`
    section 1. The fix made in an earlier session (`DeleteSnapshot`) was in an
    untracked file and is lost; redo from the review.
17. `app/joviandssplugin/main.go:119` ignores the error of
    `GetPluginServer` (`jdss, _ :=`); a failure leads to a nil dereference.
18. `app/joviandssplugin/main.go:66` opens the log file with
    `O_CREATE|O_WRONLY` (no `O_APPEND`/`O_TRUNC`): a restart overwrites the
    start of the old log and leaves its tail.
19. Config keys: the code reads `logfile`, the docs say `logpath`; `pool` is
    read only at top level, the docs put it under `endpoint`.
20. `RestEndpointCfg`/`ISCSIEndpointCfg` have only `json` tags but are read
    with `yaml.v3`; they work only because yaml.v3 lowercases field names.
    Add `yaml` tags.
    Items 19 and 20 are done on `master` since `4c03c7d` (2026-11-23, ADR
    0002): both guides and `deploy/cfg/cfg.yaml` describe the keys the code
    reads, every key has a `yaml` tag, and a test reads the guides' examples.
21. `PluginName` is hard-coded to `iscsi.csi.joviandss.open-e.com`; NFS
    config section (`nfs:`) is not parsed on `master`.

## Documentation

22. Commit `596f5ef` docs do not match the repo files: `namespace.yaml` has no
    Pod Security labels, `deploy/talos/modulepatch.yaml` is the old format,
    `deploy/talos/workload-isolation-patch.yaml` does not exist,
    `deploy/talos/securitypatch.yaml` still exists. kub-testing depends on the
    namespace labels.
23. Duplicate `docs/` folder (identical copy of `doc/`) committed in
    `596f5ef`; README links `doc/`.
24. Node DaemonSet declares secret volume `jdss-node-cfg` (not `optional`),
    the install docs never create it: node pods stay `ContainerCreating`.
25. `configuration.md`: undocumented `prot`, `user`, `pass`; typo
    `iddletimeout`.
26. `doc/compile.md` matches the `nfs-support` `Makefile`, not `master`'s.
27. README links NFS configuration to `configuration.md` instead of
    `configuration-nfs.md`.
28. `install.md` says iSCSI tools are easy on most distributions, but the node
    manifest is Talos-specific (`/system/iscsi`, `hostPath` of type `File`
    for `/usr/local/etc/iscsi/iscsid.conf`).
29. Typos, unfinished `[TODO]`, near-duplicate `install.md`/`install-nfs.md`
    and `configuration.md`/`configuration-nfs.md`.
30. README "tested on Talos OS 1.6": update once tested on 1.14.

## Manifests and platform

31. Sidecar images are old and from the frozen `quay.io/k8scsi` registry
    (`csi-provisioner:v2.1.2`, `csi-attacher:v3.1.0`,
    `csi-node-driver-registrar:v1.0.2`); move to `registry.k8s.io/sig-storage`
    and check compatibility with current Kubernetes.
32. Manifest layout: iSCSI at `deploy/joviandss/`, NFS in
    `deploy/joviandss/nfs/`; `classes.yaml` duplicates the classes in the
    controller manifest and is not documented.
33. Test the plugin with Talos workload isolation enabled (Talos >= 1.14).

## Repository

34. Decide where `review/` lives (commit, or keep outside the public repo).
35. Plan the merge of `nfs-support` (21 commits ahead of `master`, already
    uses the full import path).
    Planned: `docs/design/0001-nfs-support-merge-design.md` (PROPOSED
    2026-10-28); first priority, before the controller fixes (now 0003, after REST compliance 0002).

## Found in testing (kub-testing `project-status.md`)

36. `pkg/rest/types.go:81`: `ResourceVolume.Encryption` is `string`, current
    JovianDSS sends an object (`{"enabled": false, "method": null,
    "keystatus": null}`). Every GetVolume/ListVolumes of an existing volume
    fails to parse (csi-sanity 2026-10-27). Same field at `types.go:173`
    (snapshot struct), unverified. Field is unused: make it a struct or
    `json.RawMessage`. Together with C7 (item 16) this crashes the controller
    on every CreateVolume retry. REST spec v4 (`jdss-rest-api-spec/v4/d/
    pools/+poolname+/volumes/GET-doc`) documents the object; older releases
    send `"off"`: accept both (custom `UnmarshalJSON`).
37. `ResourceSnapshot.UnmarshalJSON` (`pkg/rest/types.go`) parses `creation`
    with the zero-padded layout `2006-01-02 15:04:05`; the v3 and v4 snapshot doc
    example is `"2015-5-27 16:8:35"`, which it rejects. Verify against real
    snapshot JSON (csi-sanity ctrl-B) before changing.
38. `pkg/controller/controllerplugin.go:687-690` (DeleteSnapshot): an ID
    that does not parse returns `InvalidArgument`; the CSI spec wants OK for
    an unknown snapshot (csi-sanity 2026-10-28). Check DeleteVolume for the
    same pattern.

## ADR 0002 and the ADR README (review findings of 2026-11-11, pending decision)

Collected from the consistency, testing, options and README reviews in
`review/2026-11-11-adr-*.md`; each item names the review that argues it.
Line numbers refer to the record and the README as of the working tree on
2026-11-11, after commit `4470b19`.

39. Whether a record may cite the hurl suite's `project-status.md` and its
    Finding numbers (Tables 4 and 5, the Relationship bullet, 26 citations);
    the 2026-11-08 rule says that file is internal. (consistency-review-2,
    item 4)
40. Define "36 call functions": 34 compliance-matrix calls plus the two `nil`
    stubs, or the 37 call methods of `RestEndpoint` minus the empty lister;
    "the 36 calls … stay" in Risks is wrong under both; Option B4 uses the
    number too. (consistency-review-2 item 20; options-review item 6)
41. Catalogue entries for `security_plugin_secrets_in_logs` and
    `security_plugin_rest_request_dump_every_call`, both registered, neither
    in the Security findings catalogue. (consistency-review-2, item 21)
42. A README rule for the three prose `####` headings under Evidence gaps.
    (consistency-review-2, item 23)
43. Names for the `DeleteClone` caller and for the cleanup fix, if they are
    to be entries; whether every fix in What changes must be a named item
    (the `per_page` commit and the stage 1 move have none).
44. Testing: the certificate error check of risk 11 and the TLS handshake of
    risk 10 as steps of the record or as kub-testing rows; step 7's mechanism
    (a CSI command-line client against the controller socket, or a session
    kept open by hand), or leave it to the tester. (testing-review)
45. Options: whether B3, B4 and B5 were real contenders; the README asks for
    two or three; if any was listed to complete the design space it goes,
    with one sentence of why not under B1 or B2. (options-review, item 1)
    Closed 2026-11-12: the user ruled that options are strategies and B1 to
    B5 were implementation nuances; all five moved into question 28's
    resolution, Option A is the one strategic matter left.
46. Options, Low: Option A's price and the bold Decision predate question 29
    (nothing older is supported); "mapped from its recorded message only"
    understates the catalogue (status, class, errno, message, two marked
    legacy patterns); the intro's "the Decision lists them" (it names six of
    twenty-seven); B1 restates three Table 2 rules. (options-review, 2–5)
47. Sixth consistency review, Low: step 4 "whose texts the items fix" while
    `impl_plugin_rest_return_parse_errors` states no error shape; "the
    request log" and `plugin.log` not in Terms, and step 6 should say "the
    controller's log"; risk 7 "steps 1-3" should be "steps 1 and 2".
    (consistency-review-4)
48. README: applied on 2026-11-11 (see the review's Outcome): tables numbered
    by position, the section order, the Consequences row, the overlap cut
    in the Risks row, and nine of the twelve Low items. Still open: whether
    "Names, not IDs" is renamed now that it holds the numbering rules; the
    `dataloss` row once a catalogue exists; whether the `work` to `impl`
    history stays in the README. (adr-readme-review)
49. README skeleton: applied on 2026-11-11: the A2 block, priority numbers
    on both sides, "what distinguishes it", the A1/A2 and Option B naming in
    the Options row, the question-decisions paragraph admitted by the
    Decision row and the skeleton. Still open: whether the skeleton's
    letters should match 0002 (A the one-option group, B the contenders).
    (adr-readme-review, addendum)
50. The record's open questions 22, 25, 30 to 33; the approver; acceptance.
51. The stray `docs/design/backup-0002-rest-api-compliance-design.md`,
    untracked: delete or keep.
52. Report to the spec project: a tab character at line 15389 of the jdss2
    v4 OpenAPI in `jdss-rest-api-spec` (YAML forbids tabs); the repository
    is read only for this work.
53. Names against the naming rule of 2026-11-11 (a human must keep track of
    what is referenced; precise enough never to collide with a later item;
    for a REST call: version, path, method, essence): 38 of the record's 90
    names failed it, with a proposed name each in
    `review/2026-11-11-adr-0002-names-review.md`. Decided on 2026-11-12
    (long names accepted; renamed as proposed, the three "split or rename"
    entries under their single proposed name) and applied the same day: the
    record, the registry (each old row kept under the new one, marked
    superseded), the examples of both READMEs and the notes; the dated
    reviews keep the old names, which the superseded rows map; `ISSUES.md`
    and the test cases held none. Still open: the entry
    `bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised`
    also covers the NAS snapshot create (`CreateNASSnapshot`), whose path
    the name does not carry; split it into two entries, or leave the NAS
    case inside this one.
54. The record's Problem statement against the rule of 2026-11-12 (a few
    plain sentences that read as one piece; no bullet per question, no list
    of the tests that prove success or of the issues left out): it was 85
    lines, three bold parts with their evidence, a seven-bullet success list
    and an exclusions paragraph with pointers. Replaced on 2026-11-12 at the
    user's request by four sentences, the last two the user's; the two
    statements with no other home moved to the What changes introduction
    (compliance is stages 3 and 4) and to the generated types item (fields
    of unrecorded states stay tolerant); Context gained the links to the
    catalogues the old section carried. See the Outcome of
    `review/2026-11-12-adr-0002-problem-statement-review.md`. Done.
55. Coherence of the record with the Problem statement of 2026-11-12
    (`review/2026-11-12-adr-0002-coherence-review.md`). High: success is
    compliance, while stage 4, the driver fixes and the hygiene commits are
    not compliance work, as the Relationship section itself says; the
    sentence moved into the What changes introduction (line 425) claims
    stage 4 is needed for compliance and is wrong under the definition;
    three ways out, (c) keep the statement and say once that the series
    carries more than the success criterion measures, recommended. Medium:
    the Decision's bold text and the title say "v4 support" where the
    statement says compliance; Context shows but never says the statement's
    "assumptions"; the JovianDSS 2 exclusion is restated in five places in
    four wordings and "iteration" is undefined. Low: the Options
    introduction's tie to the statement; "the JovianDSS 1 REST API v4"
    against one recorded release. User decision. Later the same day the
    user added security as priority 2 (bare name; list and references
    renumbered) and had one sentence on compliance put in front of the
    Decision; added detail (restated rules, a scope sentence on the security
    and crash fixes) was rejected and removed. User decisions of 2026-11-12:
    the sentence in the What changes introduction stays as it is; the title
    becomes "API v4 compliance for JovianDSS 1" with the file renamed to
    match. Still open: the rest of the High finding; the "assumptions"
    sentence for Context; the six wordings of the JovianDSS 2 exclusion and
    the undefined "iteration"; the Low item on "the JovianDSS 1 REST API
    v4". Done 2026-11-12: the Decision rewritten from Option A and the
    Risks at the user's request; its rationale names all eight priorities
    with the risks each one guards or accepts, the types sit in the
    question paragraph as question 28. The record is
    `docs/adr/0002-api-v4-compliance-jdss1.md` since 2026-11-12.
56. Coherence of the record with the Decision of 2026-11-12
    (`review/2026-11-12-adr-0002-decision-coherence-review.md`). Medium:
    Metadata's Date dates the types, now question 28, instead of the
    Decision's own two dates; the rationale's "without a guard" sentence
    names risks 12 and 14 and misses risk 13; "JovianDSS 2 is not touched"
    is contradicted by Future work's "prepares for it" and should take the
    Problem statement's words. Low: the crash fixes (question 12) and the
    reply size limit unnamed in the Decision; risk 19's guard credited to
    "a diff and a check"; the `iscsi.iqn` and `prot: http` upgrade-stoppers
    not in priority 2's price; "an unknown error stays unknown" twice in
    the Decision. Applied 2026-11-12 (see the review's Outcome): the three
    Medium, risk 19 and the duplicated clause. Still open: whether the
    Decision names the crash fixes (question 12) and the reply size limit;
    whether the price of questions 20, 21 and 23 is stated in the Decision.
57. Coherence of the whole record, end of 2026-11-12
    (`review/2026-11-12-adr-0002-consistency-review.md`): no High, no new
    Medium. Low: the Decision's "no request is logged" says more than the
    item (method, path and status stay logged at Debug); question 19 cited
    for the no-secrets decisions it does not hold; question 28's price
    restates risks 19 and 20; row 2's "where it is argued" should be Option
    A and risk 2. Carried over: 55 (the "assumptions" sentence; the
    exclusion restated in nine places in five wordings, "iteration"
    undefined) and 56. Item 46's first point may close: question 29 says
    what the hedge "or with a different v4" describes. The four Low items
    were applied on 2026-11-12 (the review's Outcome); the carried-over
    items stay with 55 and 56.

## Found by the branch review and the live cycles on `03fc012` (2026-11-24)

All in `review/2026-11-24-master-review-and-live-cycles.md`, with lines on
`03fc012` and the evidence for each; IDs in brackets are those of the
earlier reviews and of the tracker. Nothing here is implemented yet. The
four important findings of the same branch review are fixed (`aa7d075`,
`9a7dfab`).

58. `pkg/driver/iscsi_driver.go:524`: `err.GetCode()` on the nil error of a
    successful DeleteSnapshot ends the controller (C1, F4). Seen in every
    live cycle; it cuts the second controller preset short. This is the
    crash item 16 says was fixed once and lost. First thing to fix.
59. `pkg/controller/controllerplugin.go:833`, `:870`: ListSnapshots by
    `snapshot_id` has its condition inverted: an empty list for a snapshot
    that exists, a nil dereference on a failed read (C8).
60. `pkg/controller/listers.go:196`, `:242`: ListVolumes reads a nil
    snapshot descriptor for a clone of a snapshot made by hand (C9).
61. `pkg/driver/nfs_driver.go:589`: NFS DeleteSnapshot deletes the snapshot
    without reading its clones, and the appliance deletes the clones with
    it: restored volumes are destroyed (D15).
62. `pkg/driver/iscsi_driver.go:529`, `pkg/driver/nfs_driver.go:605`: when
    a snapshot delete is refused for its clones, DeleteSnapshot deletes
    every clone whose name starts with `s`, by the first letter alone. New.
63. `pkg/controller/controllerplugin.go:547`, `:559`: CreateVolume returns
    no `content_source` for a volume made from a snapshot or a volume. New,
    from cycle `20261123-231354`.
64. An ID that is not in the plugin's format is answered `InvalidArgument`;
    csi-sanity expects `NotFound`, and success for DeleteVolume
    (`controllerplugin.go:373`, `:584`, and `NewVolumeDescFromCSIID`). Four
    specs. Changing it reverses a part of ADR 0002: a decision.
65. `pkg/controller/controllerplugin.go:1030`: ValidateVolumeCapabilities
    confirms an empty list of capabilities; `vCap` (`:112`) is never set.
66. ControllerPublishVolume never checks the node ID
    (`controllerplugin.go:884` only logs it): publishing to a node that
    does not exist succeeds.
67. `pkg/controller/controllerplugin.go:948`: the publish error answers
    `Unknown internal error` and drops the cause; `:553`: CreateVolume
    turns any other status of `VolumeComply` into `Unknown`; `:454`: the
    AlreadyExists message prints the requested size where the size of the
    existing volume belongs.
68. The appliance resets a connection the plugin reuses
    (`read: connection reset by peer` on a DELETE of a target and on a POST
    of a LUN, three times in cycle `20261123-231354`, about 32 s after each
    publish began). Cause open. Go does not send a DELETE or a POST again.
    To settle: a cycle at Debug with the duration, the local port and the
    reuse of the connection in the request line. Remedy either way: no
    kept-alive connections (`pkg/rest/rproxy.go:222`).
69. The request timeout (`idletimeout`, 30 s in the lab) is shorter than a
    volume delete on a pool with 100 volumes or more (346 of 1722 recorded
    deletes); a publish takes about 25 s by the appliance's own time
    (target create 10.5 s, LUN attach 14.3 s at the medians).
70. `pkg/rest/types_methods.go:350`, `pkg/controller/controllerplugin.go:453`:
    a NAS volume without a quota reads as size 0, so a repeated NFS
    CreateVolume without `limit_bytes` ends in AlreadyExists (D10).
71. Smaller, in `pkg/rest` and `pkg/driver`: requests and the pause before
    a GET is sent again ignore a cancelled context (`rproxy.go:94`, `:170`);
    a list read never ends on an appliance that returns a full page for
    every page number (`common.go:56`); a `maxreplysize` of the largest
    64-bit number overflows (`rproxy.go:116`, `:195`); the wait for a new
    target in PublishVolume does not sleep on the target-not-found code
    (`iscsi_driver.go:665-676`); the creation time of a snapshot read by
    name may be off by the appliance's offset from UTC
    (`types_methods.go:195`, to check on the lab).
72. `pkg/node/nodeplugin.go:113`: the node plugin prints the whole volume
    capability at Debug, with the mount flags, which can hold a password.
    The controller's lines of that kind are gone since ADR 0002.

## Found by the review of ADR 0003 against csi-test v5.6.0 (2026-11-25)

Code reading on `03fc012` and csi-test v5.6.0's `pkg/sanity/controller.go`.
Both lie behind the crash of item 58: their csi-sanity specs run after it
and were answered Unavailable in every cycle. ADR 0003 takes both
(`docs/adr/0003-iscsi-csi-operational-compliance.md`).

73. `pkg/controller/controllerplugin.go:675-712`,
    `pkg/driver/iscsi_driver.go:485-498`: CreateSnapshot with the name of
    another volume's snapshot makes a second snapshot and answers OK; CSI
    gives `AlreadyExists`. The name on the appliance comes from the CSI name
    alone, and ZFS keeps a snapshot's name within its volume, so nothing
    stops it. One spec: "should fail when requesting to create a snapshot
    with already existing name and different source volume ID"
    (`controller.go:1388`). To find the name the plugin would have to read
    the snapshots of the whole pool on every create (ADR 0003 risk 12).
    Kubernetes names each snapshot `snapshot-<UUID>` and never sends it.
74. The snapshot side of item 64: `pkg/driver/snapid.go:124-125` answers
    `InvalidArgument` to a snapshot ID of fewer than three parts, as every
    one csi-sanity makes up is, and `controllerplugin.go:366` answers
    `InvalidArgument` to any snapshot source of CreateVolume its parser
    refuses. Four more specs, where csi-sanity expects `NotFound`, an empty
    list or success: CreateVolume from `non-existing-snapshot-id`,
    ListSnapshots by `none-exist-id` and by a made-up source volume (the
    volume side, as 64), DeleteSnapshot of `reallyfakesnapshotid`. ADR 0003
    question 21 decides the answers.
