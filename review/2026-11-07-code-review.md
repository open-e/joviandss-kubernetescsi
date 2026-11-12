# Code review: joviandss-kubernetescsi (after the NFS merge)

- Date: 2026-11-07
- Revision: `11ffbfd` (Merge branch 'nfs-support' into master). The working tree has only doc changes.
- Previous review: [2026-10-02-code-review.md](2026-10-02-code-review.md) (revision `c466239`, before the merge). This review re-checks every ID from that review on `11ffbfd` and adds new findings. IDs continue the same series (C21+, D10+, N13+, K11+, R7+, P4+, X2+), and the tracker `kub-testing/ISSUES.md` uses the same IDs. R14-R18 come from a second, detailed pass of `pkg/rest` (end of section 2d).
- Scope: all Go code: `pkg/rest`, `pkg/driver`, `pkg/controller`, `pkg/pluginserver`, `pkg/identity`, `pkg/common`, `pkg/node`, `app/`, `cmd/`.
- Method: static, read-only. Four reviewers each read one area in full (rest, driver, controller with server and app, node with CLI). The lead then re-read the lines behind every high-severity finding and most medium ones; those are marked **[verified]**. Go was not installed, so nothing was compiled, vetted or run.
- Confidence: **[verified]** means the code path was read line by line. **[plausible]** means the outcome also depends on external behaviour (iscsiadm, kernel, mount-utils, the appliance) that was not tested.
- Line numbers are exact on `11ffbfd`.

Main conclusion: the iSCSI node path of the merged code probably cannot work. NodeStage returns success when discovery, node-record creation or login fails (N13). Discovery passes the IQN as the discovery type (`-t iqn`), which iscsiadm rejects. A retry with an existing session also returns success with nothing mounted (N14), and NodeUnstage never logs out (N15). NodePublish then bind-mounts an empty staging directory, so a pod would write to the node's local disk. csi-sanity node presets use fake publish contexts and never reach this code, so no recorded run shows it. A real PVC on the lab settles it.

---

## 1. Status of the 2026-10-02 IDs on `11ffbfd`

Status values: FIXED (no longer in the code), GONE (code removed; the defect may survive in dead code), CHANGED (still wrong, but differently), PRESENT (unchanged).

### 1a. Crashes

| ID | Status | Where now | Notes |
|---|---|---|---|
| C1 / F4 | PRESENT | `pkg/driver/iscsi_driver.go:497` | `err.GetCode()` on the nil returned by a successful DeleteSnapshot. Not copied to NFS (`nfs_driver.go:630` uses `jrest.ErrCode`). |
| C2 | PRESENT | `pkg/driver/snaptoken.go:100,106,110` | Writes through the nil named return `t`. Called from `controllerplugin.go:591,756` with any `starting_token`. |
| C3 | PRESENT | `pkg/rest/rproxy.go:138,141,145,148,153,157` | `res.StatusCode` with `res == nil` on every transport error (DNS, dial, timeout, TLS). `:152,156` also print with `fmt.Println`. |
| C4 | PRESENT | `rproxy.go:122-124` | `SetBasicAuth` before the `http.NewRequest` error check. |
| C5 | CHANGED (more sites) | `pkg/rest/errors.go:140-149` | Logs the nil case and goes on to read `err.Errno`. Nil callers: `entrielisters.go:100,148,194,240,290` (the last three are the NAS listers from the merge). |
| C6 | FIXED | `controllerplugin.go:649-677,724-733` | Now switches on `jrest.ErrCode(rErr)` and prints `rErr.Error()`. |
| C7 / F2 | PRESENT | `controllerplugin.go:408-414`, `general_volume.go:35` | Only DNE returns; other errors reach `vdata.GetSize()` on nil. |
| C8 | PRESENT | `controllerplugin.go:793-807,826` | Condition inverted (`rErr != nil`): the entry is built from nil `snap` when the lookup fails, and nothing is appended when it succeeds; `:826` returns `nil, nil`. |
| C9 | PRESENT (two copies) | `pkg/controller/listers.go:196` (iSCSI), `:242` (NAS) | `if sd, err := ...; err != nil` then `sd.CSIID()` on nil. |
| C10 / F3 | PRESENT here | `controllerplugin.go:1009-1011` | No `WithLogger`. Fixed only in the user's uncommitted kub-gate tree. |
| C11 | PRESENT | `controllerplugin.go:961-964` | `m.Mode` with `m` possibly nil. |
| C12 | GONE | (dead) `pkg/node/targetapi.go:287` | `starget` is no longer written or read. The nil-logger bug remains in dead code. |
| C13 | PRESENT | `app/joviandssplugin/main.go:124,126` | `jdss, _ :=`, then `jdss.Run()` on nil. Now also reached when controller setup fails (P1 change). |
| C14 | PRESENT | `main.go:44-48,110-119`, `controllerplugin.go:146` | `cfg == nil` without `--config`. |
| C15 | PRESENT | `cmd/controller/listSnapshots.go:69` | Reads `resp.Entries` on the nil response from `controllerplugin.go:826`. |
| C16 | PRESENT, wider | `pkg/driver/snapid.go:139,147-151` | `Name()` panics on an empty name (`ss_`/`s_`). Callers: `iscsi_driver.go:521,527`; `nfs_driver.go:662,668` and **`:632`, the success path of NFS DeleteSnapshot**; `controllerplugin.go:335,727,729`. |
| C17 | PRESENT (three sites) | `pkg/rest/rest.go:104-107` | `SetupEndpoint` logs the error and returns nil. Ignored at `controllerplugin.go:193`, `iscsi_driver.go:409`, `nfs_driver.go:546`. The panic that follows is at `rproxy.go:91`, `storageapi.go:71`, `rest.go:94`. `cp.re` from `:193` is never used. |
| C18 | PRESENT | `lunid.go:104`, `snapid.go:27` | `[0]` without a length check. |
| C19 | PRESENT | `lunid.go:159-172` | Unreachable today. |
| C20 | PRESENT | `pkg/pluginserver/pluginserver.go:70,121-134` | The only interceptor has no `recover`; no `recover()` anywhere in the repo. |

### 1b. Data loss or wrong results

| ID | Status | Where now | Notes |
|---|---|---|---|
| D1 | PRESENT (moved) | `pkg/node/iscsi_stage.go:112` → `host.go:113-116` | NodeStage never rejects Block; `FormatAndMount` with fsType `""` becomes ext4, so raw data gets mkfs. Publish rejects Block only afterwards (`publish.go:48-50`). |
| D2 | CHANGED | `host.go:196-197` (publish, fixed); `host.go:116` (stage) | Publish binds with `ro`. Stage still mounts read-write (N25). |
| D3 / F6 | CHANGED | `iscsi_stage.go:101-106,122` | A retry no longer logs out a live session; it now deletes the node record and returns OK with nothing mounted (N14). |
| D4 | FIXED | `controllerplugin.go:327,336,370` | Assigns the outer `err`. |
| D5 | PRESENT (two copies) | `iscsi_driver.go:211-217`, `nfs_driver.go:315-322` | Deletes every clone-less snapshot whose name starts with `v`. |
| D6 | CHANGED | `listers.go:123,147` | Shadowed counter gone; iSCSI by-volume listing still empty through D7/F12. |
| D7 / F12 | PRESENT | `pkg/rest/entrielisters.go:53-55`, `iscsi_driver.go:392` | Stub returns empty entries. |
| D8 | PRESENT (two copies) | `listers.go:194-213`, `:242` | Descriptor built from the clone's own `vd`; combined with C9. |
| D9 | PRESENT | `lunid.go:123`, `snapid.go:77` | `strings.Join(parts[1:], "")` drops underscores; `\w` allows `_` in names. |

### 1c. Security

| ID | Status | Where now | Notes |
|---|---|---|---|
| S1 | PRESENT | `rproxy.go:251` | `InsecureSkipVerify: true`, no switch. |
| S2 | PRESENT, **worse** | `rproxy.go:113,118,130-131`; `controllerplugin.go:587,626,703,837,914`; `pkg/node/publish.go:43` | At Debug, `httputil.DumpRequestOut(req, true)` logs the whole request **including the `Authorization: Basic` header**, so the appliance admin password is logged on every REST call. Request bodies (CHAP password) and full CSI requests (`Secrets` maps) are also logged. Controller sample config is Info; node manifests ship Debug (MF5). |
| S3 | PRESENT, wider | all `pkg/rest` path builders, now also `nas_*.go`, `shares.go:60` | No `url.PathEscape`; `?` or `#` in a name also cuts the path. |

### 1d. CSI behaviour

| ID | Status | Where now | Notes |
|---|---|---|---|
| N1 | GONE | (dead) `targetapi.go` | No device wait at all now (N17). |
| N2 | GONE | (dead) `stage.go:50-66` | Unstage now never logs out (N15), which is worse. |
| N3 | FIXED | `iscsi_stage.go:83-101` | All calls use `addr:port`. |
| N4 | FIXED | `publish.go:52` | Checks `IsMountPoint`, but ignores its error (N19). |
| N5 | CHANGED | `host.go:217` | `IsMountPoint` error ignored: corrupt mount counts as unmounted (N19). |
| N6 | PRESENT | `nodeplugin.go:37` | `UNKNOWN` capability. |
| N7 | PRESENT | `nodeplugin.go:69-74` | NodeExpandVolume no-op success. |
| N8 | CHANGED | `host.go:220` | No longer `Aborted`; returns a raw error, which gRPC reports as `Unknown` (N22). |
| N9 | PRESENT | `pkg/node` | No locking. |
| N10 | CHANGED | `iscsi_stage.go:81` | Tries each address in turn; one session, no multipath. |
| N11 | GONE | — | ext3 default was in dead code; live path uses ext4. |
| N12 | PRESENT | `nodeplugin.go:127,196`, `host.go:106` | `0o640`. |
| K1 | PRESENT | `controllerplugin.go:359-360,384,495-501,525`; `iscsi_driver.go:74-75` | No capacity range sends size 0 (the `limit == required` branch wins over the 16 MiB minimum); `required > limit` not rejected; clone and restore report `CapacityBytes = 0`. |
| K2 | PRESENT | `pkg/driver/common.go:58,64-65,76`; `snaptoken.go:60-62`; `restwrappers.go:40` | All parts. See also F25 and K12. |
| K3 | PRESENT | `controllerplugin.go:974-984` | Empty `if`, unset `cp.vCap`, `vcap` tested instead of `vCtx`. |
| K4 | PRESENT | `controllerplugin.go:768-769,786-787,794-796` | InvalidArgument where the spec wants an empty list. The sub-point (`err` tested instead of `rErr`) is fixed at `:813`. |
| K5 | PRESENT | `iscsi_driver.go:245,257`; `nfs_driver.go:383,396`; `controllerplugin.go:565` | `VolumeHasSnapshots` reaches `Internal`. See D12. |
| K6 | PRESENT (two copies) | `iscsi_driver.go:160-168`, `nfs_driver.go:249-257` | |
| K7 | PRESENT | `iscsi_driver.go:577,610,612-615` | Rollback deletes `tname`, target was created as `iqn`; existing LUN mode not checked. |
| K8 | PRESENT (two copies) | `iscsi_driver.go:515-537`, `nfs_driver.go:656-678` | |
| K9 | CHANGED | `controllerplugin.go:597-598` | ListVolumes still `Internal`; ListSnapshots now `InvalidArgument` (`:762-763`). |
| K10 | PRESENT | `controllerplugin.go:418,422` | |

### 1e. REST client, server, CLI

| ID | Status | Where now | Notes |
|---|---|---|---|
| R1 | PRESENT | `rproxy.go:91,274,281` | `active_addr` always 0, `tries` stored and never read. |
| R2 / F24 | PRESENT | `pkg/rest/targets.go:89-98` | |
| R3 / F15 | PRESENT | `errors.go:96,247` | |
| R4 | PRESENT, spread | format strings: `storageapi.go:51,53`, `errors.go:173,265`, `targets.go:45,76`; masking as `RequestMalfunction`: `storageapi.go:92`, `targets.go:51`, `clones.go:44`, `nas_volumes.go:44,75,106`, `nas_volumes_snapshots.go:45,76`, `nas_volumes_snapshots_clones.go:44`, `shares.go:45,76` | Masking matters once C3 is fixed. |
| R5 | PRESENT | `rest.go:83-95` | |
| R6 | FIXED | `clones.go:75`, `nas_volumes_snapshots_clones.go:75` | |
| P1 | CHANGED | `pluginserver.go:84-91`; `identity/identityplugin.go:41-58` | Setup failure now returns an error, which `main.go` drops (C13 crash). Probe always ready; CONTROLLER_SERVICE advertised by the node DaemonSet (`--node --identity`). |
| P2 | PRESENT | `main.go:51-54,59,126`; `pluginserver.go:56-61` | No signal handling or `GracefulStop`; exit 0 on errors; socket removed only at start. |
| P3 | PRESENT | `main.go:71`, `pkg/common/common.go:95` | No `O_APPEND`. Worse with P4. |
| B5 | PRESENT in part | `common.go:51-75`; `deploy/cfg/cfg.yaml:2,12` | `logpath` vs `logfile`; `endpoint.pool`/`endpoint.name` ignored (top-level `pool` now in the sample); `json` tags ignored by yaml.v3 (`namelen`/`passlen`); no `loglevel` exits 1 (`main.go:82-86`); empty `pool` not rejected (`controllerplugin.go:195`). |
| CLI1-5 | PRESENT | `publishVolume.go:36-59,85-86`; `listSnapshots.go:49-61,102`; `createSnapshot.go:26,30`; `createVolume.go:98-108`; `SetupControllerPlugin` errors ignored in 8 subcommands | |
| X1 | PRESENT | `publish.go:39,63,64` | iSCSI publish logs `proto: NFS`; Unpublish labelled `PublishVolume`. |
| B9 | PRESENT | see section 3 | |

---

## 2. New findings

### 2a. Node (`pkg/node`)

Live code: `nodeplugin.go`, `stage.go` (StageVolume only), `publish.go` (both protocols), `iscsi_stage.go:37-123`, `nfs_stage.go:36-78`, and in `host.go` `GetNodeId`, `FormatMountVolume`, `MountNFSVolume`, `BindVolume`, `UmountVolume`. Everything else is dead (section 3).

| ID | Sev | Location | Defect | Trigger | Confidence |
|---|---|---|---|---|---|
| N13 | H | `iscsi_stage.go:81-122` | Every discovery, node-record or login failure ends in `continue`; when all addresses fail the loop ends and `:122` returns nil. Discovery at `:85` passes the IQN as the discovery type (`-m discovery -t <iqn>`); iscsiadm accepts only `sendtargets`/`st`, `slp`, `isns`, `fw` there. | Any unreachable portal, wrong IQN or login error, and probably every call because of `-t`. NodePublish then bind-mounts the empty staging directory: the pod writes to the node's local disk. | Return path [verified]; `-t` failure [plausible], not tested |
| N14 | H | `iscsi_stage.go:101-106,122` | Login failure (exit 15, session exists) runs `-o delete` and returns OK with nothing mounted. | kubelet retry while the first call still runs mkfs; re-stage on a node with a leaked session (N15). | [verified] |
| N15 | H | `nodeplugin.go:153-161` | NodeUnstage only unmounts; no iSCSI logout or node-record delete. The logout code is only in dead functions. | Every unstage: one session and one `/dev/sdX` leak per volume; after the target is removed the session sits in recovery; feeds N14. | [verified] |
| C21 | H | `iscsi_stage.go:112`, `nfs_stage.go:69` | `*req.GetVolumeCapability()` dereferences nil; no panic recovery (C20). On iSCSI this runs after login, so the session leaks too. | NodeStage without a capability (csi-sanity negative test). | [verified] |
| N16 | M | `nfs_stage.go:67-75`, `host.go:154` | NFS NodeStage has no mount check: each retry stacks another NFS mount; Unstage (`host.go:217-221`) removes one layer and returns OK. | kubelet retry after a timeout; MountDevice again after a kubelet restart. | [verified] |
| N17 | M | `iscsi_stage.go:111-112` | No wait for the device after login (`waitForPathToExist` no longer called). | The udev symlink appears late: blkid/mkfs/mount fail, staging fails now and then. | [plausible] |
| N18 | M | `nodeplugin.go:99-137,140-164,210-231` | NodeStage, NodeUnstage, NodeUnpublish validate nothing (volume ID, paths, capability). | Unstage/Unpublish with an empty path returns OK (ignored `IsMountPoint("")` error) instead of InvalidArgument; empty staging path gives Internal from MkdirAll. Same family as F7. | [verified] |
| N19 | M | `host.go:217`, `publish.go:52` | `IsMountPoint` errors are ignored: a corrupt mount (ESTALE, ENOTCONN, EIO) counts as "not a mountpoint". Unpublish/Unstage return OK and leave it; Publish binds over it. | Stale NFS server, dead iSCSI session. Fix: `mount.CleanupMountWithForce`, `IsCorruptedMnt`. | [plausible] |
| N20 | M | `host.go:151-154` | NFS mount options are hard-coded to `nolock`; `MountFlags` from the capability (StorageClass `mountOptions`, `vers=`, `hard`, `nconnect`) are dropped. | Any StorageClass with `mountOptions`. | [verified] |
| N21 | M | `publish.go:57-71`, `host.go:208-227` | NodeUnpublish never removes the target path the plugin created (the spec requires it; csi-sanity checks it). | Every unpublish; related to the F7 leftover. | [verified] |
| N22 | L | `host.go:220`, `nfs_stage.go:77` | Raw errors, not gRPC status errors: the CO sees `Unknown`. | Unmount or NFS mount failure. | [verified] |
| N23 | L | `iscsi_stage.go:85-113` | `exec.Command` ignores `ctx`, no timeout. | Hung iscsiadm or mkfs outlives the deadline and overlaps the retry (makes N14 and N9 worse). | [verified] |
| N24 | L | `iscsi_stage.go:83`, `host.go:154` | IPv6 addresses are not bracketed in the portal or NFS source. | IPv6 endpoint. | [verified] |
| N25 | L | `host.go:116` | SINGLE_NODE_READER_ONLY is staged read-write and may get mkfs/fsck. | Read-only volume. | [verified] |

### 2b. Driver (`pkg/driver`)

| ID | Sev | Location | Defect | Trigger | Confidence |
|---|---|---|---|---|---|
| D10 | H | `nfs_driver.go:103-106`; `rest/types.go:478-480`; `controllerplugin.go:417` | NFS sets `Quota` only when `limit_bytes >= 16 MiB`. Kubernetes normally sends `required_bytes` only, so the NAS volume has no quota (no size limit) and `GetSize()` returns 0. A retry then fails `VolumeComply` (`required > 0`) with AlreadyExists, which is terminal. ListVolumes reports capacity 0; NAS clones (`:153,:261`) never get a quota. | CreateVolume retry after a lost response; csi-sanity "same name" test. | [verified] code; `"quota": null` from recorded responses |
| D11 | M | `nfs_driver.go:108-140,153-181,249-285` | NAS volume or clone is created first, the share last, with no rollback. A retry finds the volume and stops (OK or AlreadyExists, D10), so the share is never created and every mount fails. | CreateShare fails (timeout, transient 5xx). | [verified] code; NFS untested |
| D12 | M | `iscsi_driver.go:245,257`; `nfs_driver.go:383,396`; `rest/errors.go:87,222` | Intermediate-snapshot cleanup runs only for `ResourceBusy` (clones). A leftover clone-less snapshot gives `volume has children` (errno 1000), mapped to `BusyVolumeHasSnapshots`, so cleanup never runs and DeleteVolume returns Internal (K5) on every retry: the volume can never be deleted. Fixing D7 alone does not help. | A clone that failed after its snapshot was created, then DeleteVolume on the source. | [verified] |
| D13 | M | `nfs_driver.go:302` vs `:523,537` | NFS `cleanIntermediateSnapshots` asserts `[]jrest.ResourceSnapshot`; NFS `ListVolumeSnapshots` returns `[]jrest.ResourceNASVolumeSnapshot`, so it always returns "Unable to typecast". Also (both protocols, `iscsi_driver.go:197-208`, `nfs_driver.go:302-312`) the assertion comes before the `gserr` check, so a listing error (for example DNE) becomes FailureUnknown and the DNE branch (`iscsi :262`, `nfs :401`) is unreachable. | Any NFS DeleteVolume that reaches cleanup, even after F13 and D12. | [verified] |
| K11 | M | `nfs_driver.go:695-742`; `controllerplugin.go:863` | NFS ControllerPublishVolume makes no REST call: OK and a `share_path` for a volume that does not exist (the spec wants NotFound, csi-sanity checks it); `node_id` ignored. | Publish of a deleted or made-up `vp_x`. | [verified] |
| D14 | M | `nfs_driver.go:737`, `:121-138`; `node/host.go:150-153` | NFS readonly not enforced: `read_only` goes into the publish context, nothing reads it, the share is writable and the node mounts with `nolock` only. | ControllerPublish with `readonly=true`. | [verified] |
| K12 | L | `pkg/driver/common.go:62-68` | If the token's based ID is gone (item deleted between pages), nothing is appended and the based ID is never dropped: later pages are skipped and the listing ends early and silently. | Deletion during a paged ListVolumes/ListSnapshots. | [verified] |
| K13 | L | `iscsi_driver.go:571,652-655`; `targetid.go:30` | Publish checks the length of the 64-char hash, `TargetIQN` checks the full IQN: with an IQN prefix over about 190 chars, Publish passes and every Unpublish fails with ArgumentIncorrect, leaving the target. | Long `iqn` in config (unless the appliance rejects it first). | [verified] |
| X2 | L | `iscsi_driver.go:285`, `nfs_driver.go:424` | "not CSI" message built from `dsnaps`, `ncsi` never filled; branch unreachable. | — | [verified] |

### 2c. Controller, server, app

| ID | Sev | Location | Defect | Trigger | Confidence |
|---|---|---|---|---|---|
| C22 | H | `controllerplugin.go:158-172` vs `:183-185` | iSCSI build with no `iscsi:` config section leaves `cp.d == nil` and returns nil (the NFS branch returns an error). Every controller RPC then panics (`:406,555,601`, ...). | iSCSI image with a config that has only `endpoint`/`nfs`. | [verified] |
| C23 | H | `controllerplugin.go:645,661`; `snapid.go:30-61,147-151` | CreateSnapshot does not reject an empty name. `^[-\w]+$` does not match `""`, `JBase32FromStr("")` is `""`, so a snapshot `sb_` is created on the appliance; then `sd.Name()` panics (log arguments are evaluated even at Info). | CreateSnapshot without a name (csi-sanity test). | [verified] |
| K14 | M | `controllerplugin.go:489-492` | CreateVolume checks only that capabilities are non-nil, not access mode or type. A MULTI_NODE_MULTI_WRITER iSCSI volume is accepted and ControllerPublish never enforces single-node use, so one LUN can be attached RW on two nodes. | RWX PVC on iSCSI (the example PVCs, MF6). | [verified] |
| K15 | M | `controllerplugin.go:468-473,512-514,524-525` | CreateVolumeResponse never sets `Volume.ContentSource`; the spec says it MUST when the request had one. | Any clone or restore. Some external-provisioner versions reject it. | code [verified], impact [plausible] |
| K16 | L | `controllerplugin.go:790-792` | ListSnapshots with `snapshot_id` and a non-matching `source_volume_id` returns FailedPrecondition, not an empty list; the message calls `sd.Name()` (C16 panic for `ss_`/`s_`). | — | [verified] |
| K17 | L | `controllerplugin.go:900-902` | ControllerPublishVolume default branch drops the REST error text ("Unknown internal error"), e.g. the `RequestTimeout` from `iscsi_driver.go:641`. | Any unexpected publish error. | [verified] |
| K18 | L | `controllerplugin.go:568-569` | DeleteVolume with an unparseable ID returns InvalidArgument; the spec treats an unknown volume as OK. Same pattern as F5. | csi-sanity delete of a made-up ID. | [verified] |
| P4 | L | `main.go:44-45,71`; `controllerplugin.go:146`; `common.go:95` | With `logfile` set, the controller process opens the same file twice, two offsets, no `O_APPEND`: lines from the two loggers overwrite each other. | Any config with `logfile:`. | [verified] |
| P5 | L | `controllerplugin.go:146-149`; `common.go:92` | `SetupControllerPlugin` calls `os.Exit(1)` on a logger error; the formatter is set on the global logrus logger, not on `log`, so `FullTimestamp` never applies. | — | [verified] |
| P6 | L | `pluginserver.go:73-76` | Identity plugin error is logged and the possibly nil plugin still registered. Harmless today. | — | [verified] |

### 2d. REST client (`pkg/rest`)

| ID | Sev | Location | Defect | Trigger | Confidence |
|---|---|---|---|---|---|
| R7 | M | `nas_volumes_snapshots.go:79-85` | `DeleteNASVolumeSnapshot` decodes the body before the 204 check; an empty body fails to decode, so it returns `RestErrorRPM`. Siblings `DeleteNASVolume` (`nas_volumes.go:109`) and `DeleteShare` (`shares.go:79`) check 204 first. | Fix F14 (POST→DELETE) alone: every successful delete reports an error, NFS DeleteVolume (`nfs_driver.go:322`) fails. Not F23 (that is GET). | [verified] |
| R8 | M | `rproxy.go:122,257` | `http.NewRequest` without the context, so the CSI deadline never reaches the REST call; `idletimeout` is used as `http.Client.Timeout`, the whole-request limit. | CO deadline or cancel: the call runs on and overlaps the sidecar retry. A short `idletimeout` cuts off slow operations that the appliance still completes. | [verified] code; effect from net/http behaviour |
| R9 | L | `errors.go:275-277,297` | Unmatched errors return FailureUnknown with an empty message; the appliance text goes only to a Warn log. `Error()` prints "Unknown internal Error." for codes 6 and 9-16 too. | Any F26/F27 case: the CSI message tells the user nothing. | [verified] |
| R10 | L | `types.go:283-286,337-340` | Parse errors of snapshot `creation` and pool `available` are ignored (zero time, 0 bytes, no error). | `available` not an integer string: GetCapacity reports 0. | [verified] |
| R11 | L | `nas_volumes_snapshots.go:28-57` | `CreateNASVolumeSnapshot` is a copy of `CreateNASVolume`: POSTs a NAS volume descriptor to `nas-volumes`. No callers (the real path is `nas_snapshots.go:43`). | Wiring it in. Remove it. | [verified] |
| R12 | L | `shares.go:59-72` | `DeleteShare` sends nil instead of its descriptor; log says `CreateShares`. Only caller commented out (`nfs_driver.go:344`); recorded evidence shows a NAS volume delete removes its shares, so no leak today. | Wiring it in. | [verified] |
| X3 | L | `targets.go:160`; `storageapi.go:212`; `nas_volumes_snapshots.go:66,112` | Wrong log text ("Target created" with a volume name after attach; "Obtained volume listing" for one snapshot; `func: DeleteNASVolume`). | Misleads log readers. | [verified] |

**Second, detailed pass of `pkg/rest` (same day).** The lead read all 17 files line by line and checked field types and response shapes against the recorded hurl responses (jdss1 v3/v4, jdss2 v4). New items R14-R18 are below. There is no R13: it was `Disk.Led` (string in the struct, object on the wire), which design 0002 already tracks as F8b, so fixing F8 alone does not make GetCapacity work. Design 0002 describes R7, R10, R11 and R14-R16 and says where the rest go.

| ID | Sev | Location | Defect | Trigger | Confidence |
|---|---|---|---|---|---|
| R14 | M | `storageapi.go:42-58`; every `Get*` (decode before status) | The HTTP status is never used to map errors, only the JSON body. A non-JSON error body becomes `RestErrorRPM` ("Failure during processing response") or `RequestMalfunction`. | Wrong password: plain-text `401 unauthorized` (recorded on jdss1) gives no hint about credentials. Plain-text `404 not found` (recorded) gives `RPM`, not not-found. | [verified] code; bodies recorded |
| R15 | L | `types.go:447-476` | Most `ResourceNASVolume` tags are camelCase (`fullName`, `logBias`, `primaryCache`, `secondaryCache`, `accessTime`, `deduplication`); the appliance sends snake_case (`full_name`, `logbias`, `primarycache`, `secondarycache`, `atime`, `dedup`). Only `name`, `quota`, `origin`, `copies`, `recordsize` decode; the rest stay empty, without an error. Nothing reads them: remove. | Every NAS volume GET or list. | [verified] against recorded NAS volume GETs |
| R16 | L | `storageapi.go:191-220`; `iscsi_driver.go:455` | A zvol snapshot GET has no `name` key (NAS snapshot GETs do), so `ResourceSnapshot.Name` and the iSCSI `GeneralSnapshot.Name` are always empty. No caller reads it today; callers use the descriptor's name. | Any future use of `GeneralSnapshot.Name` on iSCSI. | [verified] against recorded responses, all 3 runs |
| R17 | L | `rproxy.go:130-131` | `httputil.DumpRequestOut(req, true)` runs on every request at any log level (it copies the request and body); its error is ignored. At Debug the dump includes the `Authorization: Basic` header (S2). | Every REST call. | [verified] |
| R18 | L | `errors.go:150,173,265,291` | Every appliance error is logged at Warn, including the expected not-found on idempotent deletes, so normal retries look like failures. The not-found text doubles: "Resource Resource Pool-2/x not found do not exists". | Any DeleteVolume/DeleteSnapshot retry. | [verified] |

The same pass also refined known items:
- **C3:** every transport branch crashes. A refused connection crashes at `rproxy.go:141`, and the client's own timeout (`Client.Timeout exceeded`, not a `*net.OpError`) at `:153`. So the controller would crash-loop for as long as an appliance is down (inferred, not run).
- **C4:** concrete trigger: a config without `prot`. There is no default, so the URL becomes `://host:port/...` and `http.NewRequest` fails.
- **F14 with R7:** the NFS driver already uses the correct `DeleteNASSnapshot` (`nas_snapshots.go:67`: DELETE, 204 first) everywhere except `nfs_driver.go:322`. Switching that call and deleting `DeleteNASVolumeSnapshot` removes both.
- **F26:** the out-of-space pattern is checked only for errors without an errno (`errors.go:270`). The unrecognised out-of-space messages carry errno 4 and 1000.
- **Checked OK against recorded data:** NAS `quota`, `reservation`, `refquota`, `refreservation`, `recordsize`, `copies` are strings, as decoded (`quota` is `null` when unset: D10). Snapshot `creation` is zero-padded, so the layout at `types.go:230` works (backlog item [37]). The `clones` string `Pool-1/a,Pool-1/b` parses correctly. Base32 names replace `=` with `-`, so every name regex holds. Target `outgoing_user: null` decodes fine.

---

## 3. Dead code (B9 / MK16, updated)

None of these has a caller on `11ffbfd`. Several would crash or do the wrong thing if wired in. Remove them.

- `pkg/node/targetapi.go` except the const `deviceIPPath` (`:43`): `GetTarget*`, `Serialize`/`DeSerialize` (C12 nil logger), `Target.StageVolume`/`UnStageVolume`/`UnMountVolume`/`FormatMountVolume`, `GetStageStatus`.
- `pkg/node/target.go`; `pkg/node/nfs_publish.go` (also ignores the `BindVolume` error at `:57`).
- `ISCSiUnStageVolume` (`iscsi_stage.go:126`), `NFSUnStageVolume` (`nfs_stage.go:81`): they take a `*NodeStageVolumeRequest`, so they cannot be wired into NodeUnstage as written. `UnStageVolume` (`stage.go:50`) is a stub.
- `host.go`: `UMountDevice`, `waitForPathToExist`.
- `pkg/driver`: `cloneLUN` (`iscsi_driver.go:39`, `nfs_driver.go:52`; the NFS copy calls the zvol `CreateClone`), `nameToID` (`lunid.go:38`), `NewSnapshotTokenFromStr` (`snaptoken.go:144,168`, always panics), `snapshotToken`, `NewSnapshotToken`.
- `pkg/controller`: `lunid.go_tmp`, `lockVolume`/`volumesInProcess`, `parseOrigin`, `getRandom*`, `getStandardID`, `config.go` `GetConfing`/`ControllerCfg`.
- `pkg/rest`: `CreateNASVolumeSnapshot` (R11), `DeleteShare` (R12, caller commented out), `PromoteClone` and `AddUserToTarget` (stubs returning nil). From the second pass: `ListVolumes` (the REST one, with `GetVolumesData`), `getResultsEntries`, `GetTimeStamp`, `StorageInterface` (old signatures, not implemented by `RestEndpoint`), `RestProxyInterface`, the duplicate `restProxy` struct, `RestProxyCfg`, `StorageCfg`, `SnapshotDescriptor`, and about 16 unused `*Data` types in `payload-def.go`. Design 0002 A6 plans their removal.

## 4. Checked and OK

- Every `exec.Command` uses an argument vector, no shell. IQN, port and LUN on the node come from the controller-built publish context; `volume_id` is never used in a path.
- iSCSI device path `ip-<addr:port>-iscsi-<iqn>-lun-<n>` is correct; a `FormatMountVolume` failure logs out and deletes the record.
- Publish bind with `ro` works through mount-utils. NodeGetInfo uses `--nodeid=$(KUBE_NODE_NAME)`. NodeStage logs no secrets.
- Response bodies are closed (`rproxy.go:161`). `RestProxy` fields are written only at setup; no data race (`mu`, `requestID` unused).
- `getError` cannot dereference nil (`ErrorData.Error` is a value). Class-based mapping with no `errno` is reachable (recorded evidence). Regex submatches are indexed only after `MatchString`.
- HTTP methods are correct except F14. NAS request tags match the recorded schemas checked.
- Driver selection uses the compile-time protocol and rejects unknown values; the Makefile default `PROTOCOL_TYPE` is `nfs`.
- `CreateVolumeFromVolume` shadowing (`iscsi :160`, `nfs :249`) still returns the inner error. iSCSI PublishVolume IQN matches `TargetIQN`. Snapshot CSI ID (`sds_base64(VDS)`) splits correctly, since base64 has no `_`.
- `deleteNASVolAndShare` does not delete the share, but a NAS volume delete removes its shares (recorded evidence), so no leak.
- `WithLogger` is set by every RPC except GetCapacity (C10). `status.Errorf(codes.OK, "")` gives a nil error.

## 5. Suggested order

1. **Check N13-N15 on the lab** before more csi-sanity runs: one real iSCSI PVC and pod, then on that node `iscsiadm -m session`, `mount | grep <staging path>` and the node plugin log for "Unable to discover". csi-sanity node presets do not reach this code.
2. Node fixes N13-N15 and C21 with F6/F7 (DEC6): return errors from staging, make staging idempotent by checking the mount and session, log out in Unstage, validate requests.
3. Crashes C1-C23 (design 0003), with C20 panic recovery first, since it turns every remaining crash into an error.
4. NFS: D10-D14, K11, R7 together with F13, F14, F20-F22 (design 0002 covers the REST side).
5. S2: keep Debug off (`CLAUDE.md` rule); remove the request dump and secret logging before anyone enables Debug.
