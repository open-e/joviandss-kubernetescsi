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
