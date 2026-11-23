# Controller Crash Fixes — Design Document (PROPOSED, deferred)
[Controller_Crash_Fixes](#controller-crash-fixes--design-document-proposed-deferred)

> **Status: proposed (2026-10-28), deferred (2026-11-02) — to be reviewed
> after [0002](0002-rest-api-compliance-design.md).** Split from the original
> 0002 on 2026-11-02 (user decision: 0002 focuses on REST API compliance).
> What moved: D1 (F1, the volume `encryption` field) and the REST unit tests
> and fixtures now belong to 0002; this document keeps the controller/driver
> error-handling fixes D2-D7. D1 is kept below only as a pointer.
> Since the split: D3 (F3) is implemented by the user on kub-gate
> (uncommitted, tested 2026-11-02: GetCapacity no longer crashes, which
> exposed F8, a 0002 item); F4 re-confirmed on the merged build.
> Line references are still pre-merge (ISSUES DOC20); re-anchor on review.
>
> **Status of the original text (2026-10-28):** draft for review — nothing implemented.
> Recommends a small batch of fixes in the controller of
> `joviandss-kubernetescsi` so that csi-sanity Controller tests can run to the
> end instead of stopping at the first crash. Covers five findings from the
> csi-sanity runs (F1–F5 in `kub-testing/project-status.md`) and one review
> defect that the fixes would otherwise expose (C6). Out of scope: Node
> service, pagination, capacity semantics, the rest of the code review, the
> `nfs-support` merge itself (design 0001, see
> [Relationship to other work](#relationship-to-other-work)).
>
> **Order (user decision 2026-10-28): after
> [0001](0001-nfs-support-merge-design.md), the `nfs-support` merge.**
> Renumbered from 0001. File and line references below are `master` before
> the merge (the deployed image); they are re-anchored to the merged code
> once 0001 lands, and this design is re-checked then: D5 is already done on
> `nfs-support`, and D4 moves to `pkg/driver/iscsi_driver.go`.
>
> Not committed. Open questions 2-5 and 7 still need decisions before any
> code change.

## Table of Contents
[Table_of_Contents](#table-of-contents)

- [Overview](#overview) — Overview
- [Problem](#problem) — Problem
- [Key Observation](#key-observation) — Key_Observation
- [Project Values](#project-values) — Project_Values
- [Design](#design) — Design
- [Alternatives Considered](#alternatives-considered) — Alternatives_Considered
- [Consequences](#consequences) — Consequences
- [Risks & Backward Compatibility](#risks--backward-compatibility) — Risks_Backward_Compatibility
- [Files That Would Change](#files-that-would-change) — Files_That_Would_Change
- [Testing](#testing) — Testing
- [Relationship to Other Work](#relationship-to-other-work) — Relationship_to_Other_Work
- [Open Questions](#open-questions) — Open_Questions

---

## Overview
[Overview](#overview)

The published controller image (`opene/joviandss-csi:latest`, digest
`sha256:1ab2aa21…`) was tested with csi-sanity v5.6.0 against a JovianDSS
appliance on 2026-10-27/28. Every Controller test run ended in a controller
crash; most reported failures are csi-sanity getting `connection refused` on
the controller's gRPC socket (`/csi-controller/csi.sock`) while the crashed
controller restarted. The connection to the appliance itself works:
volumes were listed, created and snapshotted with the configured
credentials. The same crashes hit real Kubernetes use: a CreateVolume retry, a
VolumeSnapshot deletion, and (with capacity tracking enabled) GetCapacity each
take the controller down.

The change spans two layers of the plugin: the **REST client** (`pkg/rest`,
how appliance JSON is decoded) and the **controller/driver** (`pkg/controller`,
`pkg/driver`, how errors are handled), plus optionally the **gRPC server**
(`pkg/pluginserver`). The plugin keeps talking to the appliance through the
`api/v3` paths; nothing moves to v4.

Success: on the deployed appliance, csi-sanity `ctrl-A`, `ctrl-B` and
`ctrl-cap` runs complete **without any controller restart**. Tests may still
fail, but each failure is then a real, individual finding.

---

## Problem
[Problem](#problem)

Evidence: `kub-testing/tmp/ctrl-A-20261027-155122`,
`ctrl-B-20261028-085605`, `ctrl-B-20261028-090159` (`plugin.log`,
`plugin-previous.log`). Line numbers are on current `master`; the image's
stack traces match them.

| ID | Site | Defect | Trigger | Effect | Review |
|---|---|---|---|---|---|
| F1 | [`types.go:81`](../../pkg/rest/types.go#L81), [`:173`](../../pkg/rest/types.go#L173) | `Encryption string`; the appliance sends `{"enabled": false, "method": null, "keystatus": null}` | any response that describes an existing volume (GetVolume, ListVolumes) | JSON decode error → `Internal` | — (new) |
| F2 | [`controllerplugin.go:373-379`](../../pkg/controller/controllerplugin.go#L373-L379) | `VolumeComply` returns only on `ResourceDNE`; any other error falls through to `vdata.GetSize()` on nil | CreateVolume for a name that exists (every retry), combined with F1 | **crash** | C7 |
| F3 | [`controllerplugin.go:980-983`](../../pkg/controller/controllerplugin.go#L980-L983) | `GetCapacity` never calls `jcom.WithLogger`; [`driver.go:582`](../../pkg/driver/driver.go#L582) → `jcom.LFC` panics | any GetCapacity | **crash** | C10 |
| F4 | [`driver.go:536`](../../pkg/driver/driver.go#L536) | `err.GetCode()` on the result of `re.DeleteSnapshot`, nil on success | every successful DeleteSnapshot | **crash** | C1 |
| F5 | [`controllerplugin.go:687-690`](../../pkg/controller/controllerplugin.go#L687-L690); DeleteVolume [`:519`](../../pkg/controller/controllerplugin.go#L519) | an ID the plugin cannot parse returns `InvalidArgument`; CSI requires OK for a snapshot/volume that does not exist | DeleteSnapshot/DeleteVolume with a foreign or malformed ID | spec violation | — (new) |
| C6 | CreateSnapshot [`:622-647`](../../pkg/controller/controllerplugin.go#L622-L647), DeleteSnapshot [`:698-704`](../../pkg/controller/controllerplugin.go#L698-L704) | error branches call `err.Error()` where `err` is the (nil) ID-parse result; the REST error is in `rErr` | any REST error in CreateSnapshot / its GetSnapshot read-back / DeleteSnapshot | **crash** | C6 |

F1 is documented: REST spec v4 `pools/+poolname+/volumes/GET-doc` describes the
object (its own example still shows `"encryption": "off"`). The v3 docs in
`jdss-rest-api-spec` (scraped 2023) do not list the field and are older than
the appliance: on `api/v3` it already returns the v4-style list shape. A
Python check of all 86 `ResourceVolume` fields against 20 real volume entries
found `encryption` to be the only type mismatch.

---

## Key Observation
[Key_Observation](#key-observation)

**Each crash hides the next one.** csi-sanity reaches a code path only after
every earlier path survived, and the gRPC server does not recover from panics
([`pluginserver.go:122-133`](../../pkg/pluginserver/pluginserver.go#L122-L133)),
so one panic ends the run. Two chains are visible already:

- F1 turns GetVolume into an error → F2 turns that error into a crash.
- F4 fixed → DeleteSnapshot of a snapshot with clones returns `ResourceBusy` →
  the controller formats it with `err.Error()` on nil → C6 crash. Likewise,
  CreateSnapshot reads the snapshot back with GetSnapshot
  ([`:637`](../../pkg/controller/controllerplugin.go#L637)); if snapshot JSON
  has F1's problem, that read lands in C6.

Fixing the crashes one by one, with a build–deploy–test round each, would
take as many rounds as there are hidden crashes. Hence one batch, plus an
optional [safety net](#panic-recovery-safety-net) that makes any further
hidden crash cost one failed test instead of the rest of the run.

---

## Project Values
[Project_Values](#project-values)

Proposed, in priority order; please correct:

1. **Controller availability** — a single bad request or response must not
   take down provisioning for the whole cluster.
2. **Minimal, surgical diff** — every changed line traces to a finding; no
   refactoring, no unrelated cleanups (agent guide §3). Keeps review cheap and
   later porting to `nfs-support` easy.
3. **Compatibility with older JovianDSS releases** — do not break appliances
   that still send the old formats.
4. **Bugs stay visible** — hiding a failure is not a fix; every caught
   problem is logged with enough context.
5. **Speed to a re-test** — this batch exists to unblock testing.

---

## Design
[Design](#design)

### D1. (moved) Volume `encryption` field (F1)
[D1_moved](#d1-moved-volume-encryption-field-f1)

Moved to [0002 REST API compliance](0002-rest-api-compliance-design.md). F2
below still depends on it: F1 turns GetVolume into an error, F2 turns that
error into a crash. D2 removes the crash whatever 0002 decides.

### D2. `VolumeComply` returns on every error (F2)
[D2_VolumeComply_returns_on_every_error](#d2-volumecomply-returns-on-every-error-f2)

```go
if jerr != nil {
	if jerr.GetCode() == jrest.RestErrorResourceDNE {
		return nil, status.Errorf(codes.NotFound, jerr.Error())
	}
+	return nil, status.Errorf(codes.Internal, jerr.Error())
}
```

CreateVolume's switch ([`:475-487`](../../pkg/controller/controllerplugin.go#L475-L487))
already maps non-OK/NotFound/AlreadyExists codes to `Unknown` with the
message, so a REST failure becomes a clean error the provisioner retries.

### D3. `GetCapacity` attaches a logger (F3)
[D3_GetCapacity_attaches_a_logger](#d3-getcapacity-attaches-a-logger-f3)

The same three lines every other handler starts with
([`:427-432`](../../pkg/controller/controllerplugin.go#L427-L432)):

```go
l := cp.l.WithFields(log.Fields{"request": "GetCapacity", "func": "GetCapacity", "section": "controller"})
ctx = jcom.WithLogger(ctx, l)
```

### D4. `driver.DeleteSnapshot` returns early on success (F4)
[D4_driver_DeleteSnapshot_returns_early_on_success](#d4-driverdeletesnapshot-returns-early-on-success-f4)

```go
err := d.re.DeleteSnapshot(ctx, pool, ld.VDS(), sd.SDS(), deldata)
+if err == nil {
+	return nil
+}
```

The busy/clones handling below stays as it is.

### D5. Use `rErr` in snapshot error branches (C6)
[D5_Use_rErr_in_snapshot_error_branches](#d5-use-rerr-in-snapshot-error-branches-c6)

In CreateSnapshot and DeleteSnapshot, every `err.Error()` that formats a REST
error becomes `rErr.Error()`. Pure substitution, no control-flow change.
The read-back GetSnapshot errors are covered by the same substitution.

After 0001: `nfs-support` already has this for CreateSnapshot (`switch
jrest.ErrCode(rErr)`, `controllerplugin.go:648-676` on that branch); only
DeleteSnapshot remains to check. Use the same `ErrCode` switch style.

### D6. Unknown IDs in Delete* return OK (F5)
[D6_Unknown_IDs_in_Delete_return_OK](#d6-unknown-ids-in-delete-return-ok-f5)

When `NewSnapshotDescFromCSIID` / `NewVolumeDescFromVDS` cannot parse the ID,
DeleteSnapshot / DeleteVolume log a WARN with the ID and return an empty OK
response. An ID this plugin cannot parse was not created by it, so there is
nothing of ours to delete; OK is what the CSI spec requires for a resource
that does not exist, and it lets external-snapshotter / external-provisioner
finish their cleanup instead of retrying forever. An **empty** ID keeps
returning `InvalidArgument` (spec: required field; csi-sanity checks it and
it passes today).

### D7. Panic recovery safety net (optional)
[D7_Panic_recovery_safety_net](#d7-panic-recovery-safety-net-optional)
<a id="panic-recovery-safety-net"></a>

`grpcErrorHandler` already wraps every RPC. A `defer`/`recover` there turns a
panic into `codes.Internal` and logs it at ERROR **with the stack trace**:

```go
defer func() {
	if r := recover(); r != nil {
		s.l.WithFields(...).Errorf("panic in %s: %v\n%s", info.FullMethod, r, debug.Stack())
		resp, err = nil, status.Errorf(codes.Internal, "internal error in %s", info.FullMethod)
	}
}()
```

Trade-off against value 4: a panic no longer restarts the container, so it
is easier to miss. Guard: ERROR level with the stack, and the csi-sanity
runner flags any `panic in` line (one-line change to `run-sanity.sh`). Gain
for value 1: the next hidden crash costs one failed request, not the
controller. This is **implementation-separable**: D1–D6 do not depend on it.
Recommendation: include it ([Open question 2](#open-questions)).

---

## Alternatives Considered
[Alternatives_Considered](#alternatives-considered)

- F1 alternatives (custom type, `json.RawMessage`, lenient decoding): moved
  to 0002 with D1.
- **One fix per build/test round.** Most conservative per change, but needs a
  round per hidden crash (see [Key observation](#key-observation)); against
  value 5.
- **Move to the v4 API now.** Not needed to fix any finding; large change;
  against value 2. Separate decision.

---

## Consequences
[Consequences](#consequences)

- CreateVolume of an existing volume works (idempotent retries), ListVolumes
  works on the current appliance.
- No remaining known path from a REST error or a successful DeleteSnapshot to
  a panic in the controller's snapshot and volume handlers.
- GetCapacity returns the pool's available capacity (correctness of the
  value is not reviewed here).
- Deleting foreign/malformed IDs succeeds with a WARN.
- With D7: any other panic becomes an `Internal` error with a logged stack;
  the container no longer restarts on panics.
- csi-sanity runs show individual failures, which feed the next batch.

---

## Risks & Backward Compatibility
[Risks_Backward_Compatibility](#risks--backward-compatibility)

**Preserved (low risk):** request formats sent to the appliance, volume and
snapshot naming, the `api/v3` paths, success-path behavior of all RPCs, and
the `InvalidArgument` answer for empty IDs.

1. **Untested code reaches the cluster.** The agent machine has no Go; the
   change is first compiled on kub-gate. Guard: `go build`, `go vet` and the
   new unit tests on kub-gate before any image build; csi-sanity before and
   after on the same appliance.
2. **D6 hides an operator mistake.** A mistyped volume handle in a static PV
   would "delete" successfully and do nothing. Guard: nothing is deleted, the
   WARN names the ID. Acceptable.
3. **D7 hides crashes.** See [D7](#panic-recovery-safety-net); guard is the
   ERROR log with stack plus the runner check.
4. **F1 might not be the only mismatch for snapshots.** Snapshot JSON has not
   been captured yet (the C1 crash came first). Snapshot `creation` is parsed
   with a zero-padded layout; the doc example `"2015-5-27 16:8:35"` would fail
   (backlog 37). With D5 (and D7) that is an error, not a crash; the next
   ctrl-B run captures the real JSON. Not fixed blind.
5. **Merged code differs from the tested image.** After 0001 the fixes go
   onto `nfs-support` code: F1, F2, F3, F5 are unchanged there; F4 is fixed
   for NFS but not iSCSI (`iscsi_driver.go:497`); C6/D5 is fixed in
   CreateSnapshot. Guard: re-run the baseline on the merged image first
   (0001 testing), then re-check each D item against that code.

---

## Files That Would Change
[Files_That_Would_Change](#files-that-would-change)

| File | Change |
|---|---|
| `pkg/controller/controllerplugin.go` | D2 (`VolumeComply`), D3 (`GetCapacity`), D5 (CreateSnapshot/DeleteSnapshot), D6 (DeleteSnapshot/DeleteVolume) |
| `pkg/driver/driver.go` (`iscsi_driver.go` after 0001) | D4 (`DeleteSnapshot`) |
| `pkg/pluginserver/pluginserver.go` | D7 (optional) |
| `kub-testing/tests/csi-sanity/run-sanity.sh` | flag `panic in` log lines (with D7) |
| `review/backlog.md` | mark items 16 (C1, C7 parts), 36, 38 as addressed |
| `docs/design/README.md` (new, optional) | short conventions file, adapted from JovianDSS-Proxmox |
| `docs/design/0003-controller-crash-fixes-design.md` | this document |

---

## Testing
[Testing](#testing)

1. **Unit (kub-gate):** REST decoding tests are in 0002. Here: `go vet ./...`
   and `go build ./...` clean.
2. **Baseline (done):** published image, runs listed in
   [Problem](#problem).
3. **Live (after build and deploy of the fix image):** `run-sanity.sh`
   `identity`, `ctrl-A`, `ctrl-B`, `ctrl-cap`. Pass criterion: controller
   restart count unchanged across each run (`run-info.txt`), no `panic` in
   `plugin.log`. Expected newly visible failures: listing/pagination (D6,
   K2), ValidateVolumeCapabilities (K3), capacity (K1) — recorded, not fixed
   in this batch.
4. **Regression check:** compare per-test results against the baseline;
   every test that passed before still passes.

Not covered: unit tests for D2–D6. `CSIDriver.re` is a concrete
`jrest.RestEndpoint` struct, so faking the appliance means an `httptest`
server; worth doing, but as its own step ([Open question 5](#open-questions)).

---

## Relationship to Other Work
[Relationship_to_Other_Work](#relationship-to-other-work)

- Findings and run logs: `kub-testing/project-status.md` (F1–F5).
- [0002 REST API compliance](0002-rest-api-compliance-design.md): F1 and F8
  (response decoding); comes first (user decision 2026-11-02).
- Static review: `review/2026-10-02-code-review.md` (C1, C6, C7, C10).
- Backlog: `review/backlog.md` items 16, 36, 37, 38.
- Depends on, not part of this document: a development image registry and a
  build that does not edit `go.mod` (roadmap items 4–5, backlog 1–9).
- Prerequisite: [0001 `nfs-support` merge](0001-nfs-support-merge-design.md)
  (backlog 35). This batch is applied on the merged code.
- Sits beside, deferred: C13 (`main.go` ignores `GetPluginServer` error) and
  the log file open mode (backlog 17, 18) — roadmap item 6 lists them, but no
  test reaches them ([Open question 3](#open-questions)).

---

## Open Questions
[Open_Questions](#open-questions)

1. ~~**Where do design documents live for this plugin?**~~ **(Resolved
   2026-10-28.)** In the plugin repository, `docs/design/` (user decision).
   A conventions README there is still optional.
2. **Include D7 (panic recovery)?** Recommended yes; it changes failure
   behavior and is a judgment call on value 1 vs 4.
3. **Include C13 and the log file mode (backlog 17, 18)?** Each is a
   few-line change, but outside "makes csi-sanity run through". Proposal: next
   batch.
4. ~~**Fixture content.**~~ Moved to 0002 with the REST tests.
5. **Unit tests for the controller paths** via an `httptest` fake appliance:
   in this batch or the next?
6. ~~**Branch base, and whether to merge `nfs-support` first.**~~
   **(Resolved 2026-10-28: merge first, design 0001; this batch follows on
   the merged code. User: longer now, but a smaller diff later.)** Original
   proposal: a branch from `master` (the deployed image and all test evidence are
   `master`-layout code), ported to `nfs-support` later. Data for the
   alternative: `nfs-support` is 21 commits (2024-04 to 2024-07, mostly
   "Intermediate update") and +5504/-1634 lines over 85 files since the
   merge base; a trial merge into `master` gives 27 conflicting files
   (import-path rewrite `1f1f814`, `doc/` → `docs/` rename, `Makefile`,
   `go.mod`, `pkg/driver`, `pkg/node`, `pkg/pluginserver`). It has F1, F3,
   F4 (see [Risk 5](#risks--backward-compatibility)); an earlier csi-sanity
   run on it hit C1 and C10.
7. **Build and registry** (roadmap item 4): needed before the
   [live test](#testing). Options: LAN registry + Talos mirror patch, a private
   Docker Hub repository, or a short-lived anonymous registry (e.g. `ttl.sh`)
   for test images only.
