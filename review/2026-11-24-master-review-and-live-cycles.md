# Review of `master` after the API v4 series: branch review and live cycles

- Date: 2026-11-24
- Revision: `03fc012` (Merge pull request #11 from `rest-api-v4`; the same tree as the branch head `5961a80`). Line numbers are exact on it.
- Previous reviews: [2026-11-07-code-review.md](2026-11-07-code-review.md) (revision `11ffbfd`); its IDs (C1, D15, N13, ...) are used here where a finding is the same defect. F-numbers and other IDs are those of the tracker, `kub-testing/ISSUES.md`.
- Sources, three:
  1. A code review of the branch `rest-api-v4` against `master` at `14c4b17`, on 2026-11-23, by reading only: nothing was compiled or run for it. It read the branch before its last three commits, which fix its four important findings.
  2. The live csi-sanity cycles on the lab (Talos, iSCSI build, JovianDSS 1 `1.0up33.65410`, csi-sanity v5.6.0, plugin log at Info): `20261123-110200` on `5961a80` and `20261123-231354` on `03fc012`, the first valid cycles of the whole series. Results and logs: `kub-testing/tmp/cycles/<cycle>/`, whole on the lab machine.
  3. The request times of the appliance, taken from the hurl recordings of JovianDSS 1 on API v4 in `jdss-rest-api-spec` (read only).
- Confidence: **[live]** was seen in a cycle. **[code]** was read line by line on `03fc012`. **[recorded]** comes from the recordings. **[open]** is a hypothesis that nothing has confirmed yet.

Main conclusion: the series of ADR 0002 works against the appliance and brought no regression: the controller presets went from 6 to 20 passing specs and stopped crashing in the first of them. What stops the suite now is outside that record: one nil dereference in the iSCSI snapshot delete (C1), which ends the controller in the second preset, and a handful of CSI answers the controller gives wrongly. One new behaviour of the appliance needs a decision of its own: it resets a connection the plugin reuses.

## 1. What the series fixed, seen live

| ID | Before (`14c4b17`) | On `03fc012` |
|---|---|---|
| F1 | every GetVolume and ListVolumes of an existing volume failed to decode | ListVolumes and CreateVolume specs pass **[live]** |
| F2 / C7 | CreateVolume on an existing volume crashed the controller | no crash in the first controller preset; the guard is `controllerplugin.go:437-448` **[live]** |
| F8, F8b | GetCapacity failed to decode the pool | the capacity preset passes **[live]** |
| F9 | ListSnapshots failed once the pool had a snapshot | ListSnapshots without options passes, when it runs before the C1 crash **[live]** |
| C3 | a request without a reply crashed the plugin | `connection reset by peer` is now an error with its cause, logged once with method and path (cycle `073923` on `c272b65` still crashed on it) **[live]** |

First controller preset: 20 pass, 8 fail, no crash, the same 8 in both cycles. Capacity preset: passes. Second controller preset: cut short by C1. Node preset: 2 of 27, unchanged (F7).

## 2. Open defects on `03fc012`

### 2a. Crashes

| ID | Where | What | Confidence |
|---|---|---|---|
| C1 / F4 | `pkg/driver/iscsi_driver.go:524` | `err.GetCode()` on the nil error of a successful `DeleteSnapshot`: every successful iSCSI DeleteSnapshot ends the controller. Every spec after it fails with `Unavailable`, so the second controller preset shows almost nothing else. The NFS driver switches on `jrest.ErrCode(err)` and does not have it. Fix: return nil for a nil error before the first `GetCode()`. | **[live]** in both cycles |
| C8 | `pkg/controller/controllerplugin.go:833`, `:870` | ListSnapshots by `snapshot_id`: the condition is inverted. The entry is built inside `rErr != nil`, from a nil snapshot, and the success path falls through to `return nil, nil`. Seen live as an empty list where one entry is expected; on any failure of the read other than "not found" it is a nil dereference. | **[live]** for the empty list, **[code]** for the crash |
| C9 | `pkg/controller/listers.go:196`, `:242` | ListVolumes: `sd.CSIID()` runs when `NewSnapshotDescFromSDS` failed, so on a nil descriptor, and a snapshot that does parse never becomes the content source. A clone of a snapshot made by hand crashes the list. | **[code]** |
| C20 | `pkg/pluginserver` | No panic recovery in the gRPC server, so each of the above ends the process. Unchanged from the 2026-11-07 review. | **[code]**, that review |

### 2b. Data loss

| ID | Where | What | Confidence |
|---|---|---|---|
| D15 | `pkg/driver/nfs_driver.go:589` | NFS DeleteSnapshot sends the delete without reading the clones, and the appliance deletes a NAS snapshot together with its clones: every volume restored from the snapshot is destroyed. The cleanup paths read the clones first since the series (`deleteSnapshotWithoutClones`); this one does not. ADR 0002 leaves it to another record (its question 34). | **[code]**; the appliance's behaviour **[recorded]** |
| new | `pkg/driver/iscsi_driver.go:529`, `pkg/driver/nfs_driver.go:605` | When a snapshot delete is refused for its clones, DeleteSnapshot deletes every clone whose name starts with `s` (`IsSDS` tests the first letter only). A volume cloned by hand under such a name, `staging-db` for example, is deleted. The cleanup had the same test for `v` and now asks whether the name reads as a plugin name (`aa7d075`); this one was left. | **[code]** |
| D10 | `pkg/rest/types_methods.go:350`, `pkg/controller/controllerplugin.go:453` | Not data loss, but it blocks NFS: a NAS volume without a quota reads as size 0, Kubernetes sends no `limit_bytes`, so a repeated CreateVolume compares `required_bytes` with 0 and ends in `AlreadyExists` for ever. | **[code]** |

### 2c. CSI answers that csi-sanity fails

The eight failing specs of the first controller preset, and what else the cycles showed.

| What | Where | Detail | Confidence |
|---|---|---|---|
| CreateVolume returns no `content_source` | `controllerplugin.go:547`, `:559` | Both return paths set only the ID and the capacity. CSI: a request with a source gets it back. Fails "should create volume from an existing source snapshot"; the clone from a volume passes only because the suite does not check the field there. | **[live]** `231354` |
| An ID that is not in the plugin's format gets `InvalidArgument` | `NewVolumeDescFromVDS` (`pkg/driver/lunid.go`), reached from `controllerplugin.go:373` (source volume), `:584` (DeleteVolume) and through `NewVolumeDescFromCSIID` (ValidateVolumeCapabilities, ControllerPublishVolume) | csi-sanity uses IDs such as `fake-vol-id-ad4c070e-...` and expects `NotFound`, and success for DeleteVolume: a volume the plugin cannot have made does not exist. Four specs. ADR 0002 decided `InvalidArgument` for an ID that breaks the appliance's name rules, so changing this reverses a part of that record. | **[live]** both cycles; a decision |
| ValidateVolumeCapabilities confirms no capability | `controllerplugin.go:1030-1031`, field `vCap` at `:112` | `Confirmed.VolumeCapabilities` is `cp.vCap`, which nothing ever sets, so the answer confirms an empty list (`Expected <[]*csi.VolumeCapability | len:0> not to be empty`). The loop above it computes `supported` and nothing reads it. | **[live]** both cycles |
| ControllerPublishVolume to a node that does not exist succeeds | `controllerplugin.go:884` is the only use of the node ID, in a log line | The node ID is never checked. CSI asks for `NotFound`. | **[live]** both cycles |
| ListSnapshots by `snapshot_id` returns an empty list | C8 above | | **[live]** `110200` |
| The publish error loses its cause | `controllerplugin.go:948` | The default branch answers `Unknown internal error` without `rErr.Error()`; the other handlers pass the message on. Seen with the connection reset below. | **[live]** |
| CreateVolume turns any other status into `Unknown` | `controllerplugin.go:553` | The default branch after `VolumeComply` drops the code it was given, `Internal` for example. | **[code]** |
| The AlreadyExists message gives the requested size twice | `controllerplugin.go:454` | It prints `caprage.GetRequiredBytes()` where the size of the existing volume belongs. | **[live]** `110200`, low |

The two lifecycle specs fail on the connection reset of 2d, not on an answer of their own.

### 2d. Connections and time

**The appliance resets a connection that the plugin reuses.** **[live]**, cause **[open]**

Three resets in `231354`, all `read tcp ...->192.168.28.102:82: read: connection reset by peer`:

| Spec | Publish began | Request that was reset | At |
|---|---|---|---|
| publish to a node that does not exist | 22:16:33.4 | `DELETE .../san/iscsi/targets/<iqn>` in the unpublish after it | 22:17:06 |
| volume lifecycle should work | 22:17:15.7 | `DELETE .../san/iscsi/targets/<iqn>` in the unpublish | 22:17:48 |
| volume lifecycle should be idempotent | 22:17:57.2 | `POST .../targets/<iqn>/luns` of the second publish | 22:18:29 |

What is known: each reset hit the request sent right after the reply to the one before it; each used another local port (41494, 47672, 55410); each came about 32 s after its spec's publish began. The controller was not idle. The same reset was there on `c272b65`, where it still crashed the plugin, so it is not new with the series.

What is not known: why. Two explanations fit, and the Info log cannot tell them apart: the appliance, or something on the path, ends a kept-alive connection about 30 s after it was opened; or it closes the connection after some replies without saying so. Go's client sends a GET again on a new connection in such a case, but not a DELETE or a POST. curl, which the hurl suite uses, most likely sends any request again when a reused connection turns out dead before a byte of the reply came, which would be why no recording shows a reset; this was not checked in curl's source.

To settle it: one focused cycle at Debug, after the Debug line of a request (`pkg/rest/rproxy.go`, `Send`) also gives its duration, the local port and whether the connection was reused.

The remedy either way: no kept-alive connections. The transport is built at `pkg/rest/rproxy.go:222` with keep-alive on. A new connection to the appliance has its TLS handshake done after about 2 ms in the recordings (median of 12,509 connections); a request takes 0.5 to 14 s.

**The slow ControllerPublishVolume is the appliance's own time.** **[recorded]** The plugin does not wait. Request times on JovianDSS 1 with API v4, successful requests of all hurl runs:

| Request | Median | 90th percentile | Longest | Count |
|---|---|---|---|---|
| `POST pools/<p>/san/iscsi/targets` | 10.5 s | 11.7 s | 12.1 s | 323 |
| `POST .../targets/<t>/luns` | 14.3 s | 16.0 s | 18.2 s | 182 |
| `GET .../targets/<t>` | 0.5 s | 0.9 s | 1.2 s | 263 |
| `DELETE .../targets/<t>/luns/<v>` | 4.8 s | 5.3 s | 5.7 s | 131 |
| `DELETE .../targets/<t>` | 4.8 s | 5.3 s | 5.8 s | 318 |
| `DELETE pools/<p>/nas-volumes/<d>` | 12.6 s | 14.6 s | 22.0 s | 549 |
| `POST shares` | 6.7 s | 6.9 s | 7.1 s | 236 |
| `DELETE pools/<p>/volumes/<v>` | 6.1 s | 52.9 s | 111.9 s | 1722 |

One publish is a target create, a LUN attach and a read of the target: about 25 s at the medians and up to 31 s, which is what the cycles show (28 to 40 s).

**The request timeout is too short for a full pool.** **[recorded]** `idletimeout` is the timeout of one request, 30 s in the lab configuration. A volume delete took longer in 346 of 1722 successful recorded deletes, all of them on a pool that held 100 to 500 volumes. On such a pool DeleteVolume would end in `RestErrorRequestTimeout` while the appliance goes on deleting.

Smaller, from the branch review, all **[code]**:

| Where | What |
|---|---|
| `pkg/rest/rproxy.go:94`, `:170` | A cancelled CSI call keeps talking to the appliance: the request is built without the call's context, and the pause before a GET is sent again is a plain sleep. A list call reads every page on every call. |
| `pkg/driver/common.go:56` | A list read never ends on an appliance that returns a full page for every page number: the read stops only on a page shorter than 100 and never looks at `results`. |
| `pkg/rest/rproxy.go:116`, `:195` | A `maxreplysize` of the largest 64-bit number makes every reply read as empty: one is added to it for the reader, which overflows; setup refuses only negative values. |
| `pkg/driver/iscsi_driver.go:665-676` | The wait for a new target in PublishVolume does not wait: it sleeps only on `RestErrorResourceDNE` (`:671`), and a missing target is `RestErrorResourceDNETarget` since the error catalogue, so its three reads follow each other at once. |
| `pkg/rest/types_methods.go:195` | The creation time of a snapshot read by name may be off by the appliance's offset from UTC: the string is parsed as UTC, and the recordings suggest local time. ListSnapshots uses the epoch field, so the two calls would disagree. **[open]**: to check on the lab. |

### 2e. Logs

At Debug the controller logs no request body, no request header and no whole CSI request any more. The node plugin still prints the whole volume capability, with the mount flags of the StorageClass, which can hold a password: `pkg/node/nodeplugin.go:113`. The node manifests still ship `--loglevel=DEBUG` (MF5).

### 2f. Node

Unchanged by the series and not reviewed again: F7 (NodePublishVolume takes a request it should refuse and leaves a leftover that breaks the test pod), and N13 to N15 of the 2026-11-07 review (NodeStage can report success with nothing mounted). The node preset stays at 2 of 27.

## 3. The branch review of 2026-11-23: what became of its findings

| # | Finding | Status on `03fc012` |
|---|---|---|
| 1 | The cleanup deleted any snapshot starting with `v` that had no clones | FIXED, `aa7d075` |
| 2 | "Not found" for one snapshot during the cleanup was read as "the volume is gone" | FIXED, `aa7d075` |
| 3 | ListSnapshots for a missing source volume answered `Internal` | FIXED, `9a7dfab` |
| 4 | `VolumeComply` read a nil volume when the read failed | FIXED, `9a7dfab`; this is F2, confirmed live |
| 5 | iSCSI DeleteSnapshot crashes on success | OPEN: C1 in 2a |
| 6 | NFS DeleteSnapshot deletes the clones with the snapshot | OPEN: D15 in 2b |
| 7 | ListSnapshots by `snapshot_id` inverted | OPEN: C8 in 2a |
| 8 | A NAS volume without a quota reads as size 0 | OPEN: D10 in 2b |
| 9 | `listers.go` inverted check | OPEN: C9 in 2a |
| 10 | Creation time possibly off by the UTC offset | OPEN, to check live: 2d |
| 11 | `maxreplysize` overflow | OPEN: 2d |
| 12 | Requests ignore a cancelled context | OPEN: 2d |
| 13 | A list read that never ends | OPEN: 2d |
| 14 | The PublishVolume wait does not sleep | OPEN: 2d |
| 15 | The publish error loses its cause | OPEN: 2c, seen live |

## 4. Order proposed

Put to the user on 2026-11-24, not confirmed.

1. A record for the controller, the next after ADR 0002: C1 first, since every other check of the second controller preset waits on it; then the connections (after the Debug cycle), the unknown IDs (a decision), `content_source`, the confirmed capabilities, the node check in publish, C8, C9, panic recovery (C20), the two data-loss paths of 2b, and the CSI status of the REST codes that ADR 0002 added.
2. One real iSCSI volume in a pod, to see whether NodeStage mounts anything (N13 to N15); then a record for the node, sized by the result.
3. A live cycle of the NFS build, which the test suite cannot run yet, with D10.
4. Beside these, without a record: the references commit for the docs and manifests.

ADR 0002 stays Proposed until the second controller preset runs through and its request-log checks pass at Debug.
