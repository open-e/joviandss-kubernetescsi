# Code review: joviandss-kubernetescsi

- Date: 2026-10-02
- Revision: `c466239` (Add NFS installation guide). The only local change is `go.sum`, from `go mod tidy`.
- Method: static, read-only review. Go was not installed, so nothing was compiled, vetted or tested.
- Line numbers: section 1 (crashes) was fully re-verified in a second, dedicated nil-dereference pass, so its line numbers are exact. Elsewhere, findings marked **[verified]** were re-checked line by line, and the other line numbers are approximate (±a few lines). Confirm those before editing.

A pattern behind many of the findings: `RestError` is an interface (`pkg/rest/errors.go:49`), and several REST calls return `nil` on success. Code that calls `err.GetCode()` or `err.Error()` without first checking for nil therefore panics. The gRPC server does not recover from panics, so a panic takes down the whole plugin process.

---

## 1. Crashes (nil dereferences and panics)

Every location in this section was re-checked line by line in a dedicated nil-dereference pass. The pass read all of `pkg/rest`, `pkg/driver`, `pkg/controller`, `pkg/node`, `pkg/pluginserver`, `pkg/identity`, `pkg/common`, `app/` and the call sites in `cmd/`. The line numbers in this section are exact.

### 1a. Nil pointer or nil interface dereferences

| # | Location | Defect | Trigger |
|---|---|---|---|
| C1 | `pkg/driver/driver.go:536` | `err.GetCode()` is called on the result of `re.DeleteSnapshot`, which is `nil` on success (`pkg/rest/storageapi.go:285-287`). | **Every successful DeleteSnapshot** crashes the controller. |
| C2 | `pkg/driver/snaptoken.go:100,106,110` | `NewCSIListingTokenFromTokenString` writes to the named return `t`, which is still nil. It should write to the local `ct`. | ListVolumes or ListSnapshots with any non-empty `starting_token`. |
| C3 | `pkg/rest/rproxy.go:134,137,141,144,149,153` | Every transport-error branch returns `res.StatusCode`, but `res` is nil after a failed `Do()`. | The appliance is unreachable, DNS fails, or a request times out. The plugin crashes on every REST call until the appliance comes back. |
| C4 | `pkg/rest/rproxy.go:119-120` | `req.SetBasicAuth` runs before the error from `http.NewRequest` is checked, so `req` may be nil. | A URL that `http.NewRequest` rejects, for example a name containing a control character. |
| C5 | `pkg/rest/errors.go:129-138` | `ErrorFromErrorT` logs "no error case" when `err == nil` but then reads `err.Errno` (and `err.String()` at :247). It is called with nil from `pkg/rest/entrielisters.go:104` and `:154`. | Volume or snapshot listing gets HTTP 200 with `"data": null` (json sets the `Data` interface to nil, so :91/:141 is skipped), or a non-2xx response whose body has no `"error"` key. |
| C6 | `pkg/controller/controllerplugin.go:622,624,634` (CreateSnapshot) and `:641,643,647`, `:698,704` (DeleteSnapshot) | These branches call `err.Error()`, but `err` holds the descriptor-parse result (`:611`, `:687`), which is nil there. The REST error is in `rErr`. | CreateSnapshot gets busy, not-found, out-of-space or any unknown error. DeleteSnapshot hits a snapshot that has clones (`:698`) or an unknown error (`:704`). |
| C7 | `pkg/controller/controllerplugin.go:373-379` | `VolumeComply` returns only on `RestErrorResourceDNE`. Any other error falls through to `vdata.GetSize()` with `vdata == nil`, and `GetSize` has a pointer receiver with no nil check (`pkg/rest/types.go:133`). | CreateVolume while the appliance returns 500, times out, or sends a body that does not parse (RPM). |
| C8 | `pkg/controller/controllerplugin.go:765-774` | In ListSnapshots by `snapshot_id`, when `rErr != nil` and the code is not DNE, the entry is built from `snap.Creation` with `snap == nil`. | ListSnapshots by ID while the appliance returns any error other than "not found". The success path has the opposite problem: it returns `nil, nil` (`:798`), an empty response. |
| C9 | `pkg/controller/listers.go:135-139` | The condition is inverted (`err != nil`), so `sd.CSIID()` runs on a nil `*SnapshotDesc`. | ListVolumes when any volume's origin snapshot name starts with `s` but is not a CSI name, for example a clone of a manually created snapshot `snap-daily`. |
| C10 | `pkg/controller/controllerplugin.go:983` (GetCapacity) | This is the only RPC that never calls `jcom.WithLogger(ctx, …)`. `cp.d.GetPool` then calls `jcom.LFC(ctx)`, which panics when no logger is attached (`pkg/common/common.go:181-185`). **New in this pass.** | **Any GetCapacity call.** The capability is advertised, so csi-sanity calls it, and external-provisioner does too when `--enable-capacity` is set. |
| C11 | `pkg/controller/controllerplugin.go:933-936` (ValidateVolumeCapabilities) | `m := c.GetAccessMode()` can be nil, and then `m.Mode` is read directly instead of through `m.GetMode()`. **New in this pass.** | A capability with only an access type (block or mount) and no access mode. |
| C12 | `pkg/node/targetapi.go:290` | `DeSerializeTarget` calls `t.l.Warn`, but `GetTargetFromPath` (`:226-228`) creates `Target{}` with `l == nil` and sets `t.l` only afterwards (`:234`). | NodeUnstage with a corrupt or truncated `<staging>/starget`. The write at `:256-272` is not atomic. Every retry crashes the node plugin. |
| C13 | `app/joviandssplugin/main.go:119-121` | The error from `GetPluginServer` is discarded, so `jdss.Run()` runs on a nil `*PluginServer`. | `net.Listen` fails: the default `--soc-type=tcp` with the default unix-path `--csi-address`, the port is busy, or the socket directory is missing. `GetNodePlugin` fails. Unix socket cleanup fails. |
| C14 | `pkg/controller/controllerplugin.go:141` via `main.go:45,106-114` | Without `--config`, `handleArgs` returns `cfg == nil`, and `SetupControllerPlugin(&cp, nil)` reads `cfg.LLevel`. **New in this pass.** | `joviandssplugin --controller` started without `--config`. |
| C15 | `cmd/controller/listSnapshots.go:72` (CLI only) | `resp.Entries` is read on a nil response, because ListSnapshots by `snapshot_id` returns `nil, nil` (`controllerplugin.go:798`). **New in this pass.** | `jdss-csi-cli controller listsnapshots` with a snapshot ID. |

### 1b. Other panics with the same effect (explicit `panic` or index out of range)

| # | Location | Defect | Trigger |
|---|---|---|---|
| C16 | `pkg/driver/snapid.go:151-156` and its callers | `SnapshotDesc.Name()` deliberately panics on an empty name. `parseSDS` leaves the name empty for the `ss_` (SHA-hashed, used for names over 240 chars or whose base32 form exceeds 240) and `s_` formats, and `NewSnapshotDescFromCSIID` ignores the error from `parseSDS` (`:143`). Callers: `controllerplugin.go:301` (restore), `:700`, `:702` (DeleteSnapshot), `:763`, and `driver.go:560,566`. | Restoring from or deleting a snapshot with a long name. |
| C17 | `pkg/rest/rest.go:134-136` → `pkg/rest/rproxy.go:89` | `SetupEndpoint` ignores the error from `SetupRestProxy`, so `rp.addrs` stays empty and `rp.httpRestProxy` stays nil. `Send` then indexes `rp.addrs[0]`. The same panic happens if `endpoint.addrs` is empty in the config. `GetAddress` (`storageapi.go:73`) and `String` (`rest.go:96`) also index `Addrs[0]`. | `idletimeout` is missing or invalid (for example `30` with no unit), or `addrs` is empty. The first REST call crashes the controller. |
| C18 | `pkg/driver/lunid.go:103-105`, `pkg/driver/snapid.go:27-29` | `IsVDS` and `IsSDS` index `[0]` without a length check. | An empty name from the appliance, such as a clone or snapshot name. This is a low risk, but the fix is trivial. |
| C19 | `pkg/driver/lunid.go:159-172` | `VolumeDesc.VDS()` and `CSIID()` panic on an empty `vds`. This cannot be reached today, because every constructor sets `vds`, but it will break under future refactors. | — |

### 1c. Checked and found safe

- **Nil errors stored in interfaces:** no function returns a nil `*restError` as a `RestError`. `ErrorFromErrorT` and `GetError` always return non-nil, so `!= nil` checks on `RestError` are reliable.
- **Snapshot, volume and pool lookups:** `GetVolume`, `GetVolumeSnapshot`, `GetTarget` and `GetPool` always return a non-nil pointer together with a nil error. The dereferences at `controllerplugin.go:645,650,841,989` and `driver.go:667` are therefore safe on the success path.
- **Logger in context:** every other gRPC handler (controller, node, identity) calls `WithLogger` before any `LFC`.
- **Controller registration:** the controller is registered only when `SetupControllerPlugin` succeeds (`pluginserver.go:84-91`). However, `NewJovianDSSCSIDriver` never fails, so a broken REST configuration still registers the controller, which leads to C17.
- **Unreachable dead code:** `lockVolume` (nil-map write), `getVolume` (`:267` `rErr.GetCode()` on nil), `parseOrigin` (`:127-133`) and `NewSnapshotTokenFromStr` (`snaptoken.go:168` `parts[3]` with `len == 3`) are never called. Delete them before someone wires them up.
- **`entrielisters.go:97,147`:** `*data` when the type assertion fails would panic. It cannot be reached today, because json reuses the pre-set `*ResultEntries`.

## 2. Data loss or silently wrong results

| # | Location | Defect | Failure scenario |
|---|---|---|---|
| D1 | `pkg/node/targetapi.go:380-383` **[verified]** | Block access type is not handled. `FormatAndMount` runs with an empty fsType, and the controller does not reject block capabilities either. | A `volumeMode: Block` PVC holding raw data with no filesystem signature is run through mkfs.ext4 and its data is destroyed. Kubelet expected a device file and gets a mounted directory instead. |
| D2 | `pkg/node/targetapi.go:380-383` **[verified]** | `req.GetReadonly()` is never read, and `ro` is never added to the mount options, even though `SINGLE_NODE_READER_ONLY` is advertised. | Read-only mounts are mounted read-write. |
| D3 | `pkg/node/targetapi.go:466-490` **[verified]** | NodeStage is not idempotent. It always runs `-o new` and `--login`, and if login fails it runs `--logout` and `-o delete`. | When kubelet retries NodeStage, login fails because the session already exists. The error path then logs out the live session under a mounted filesystem, and the pod gets I/O errors. |
| D4 | `pkg/controller/controllerplugin.go:299-302, 336` **[verified]** | `sd, err := ...` shadows the outer `err`, so the switch at :336 never sees the error from `CreateVolumeFromSnapshot`. | A failed restore from snapshot (out of space, missing snapshot) is reported as success, and the PV points at a volume that does not exist. |
| D5 | `pkg/driver/driver.go:~200` | `cleanIntermediateSnapshots` deletes every clone-less snapshot whose name starts with `v`. | Snapshots that users created outside CSI (for example "vacation") are deleted. |
| D6 | `pkg/controller/listers.go:89,108` **[verified]** | `for i, s := range snaps` shadows the counter `i`, so `entries[:i]` is always `entries[:0]`. | ListSnapshots filtered by `source_volume_id` always returns an empty list. |
| D7 | `pkg/rest/entrielisters.go:54-56` **[verified]** | `GetVolumeSnapshotsEntries` is a stub that returns no entries. | This also causes the empty result in D6, and intermediate snapshots are never cleaned up, so they leak. |
| D8 | `pkg/controller/listers.go:~135-145` | The snapshot ID is built from the clone's own descriptor instead of the origin volume. For volume-to-volume clones the intermediate snapshot is named after the new volume. | ListVolumes reports wrong content sources, and a clone shows itself as its own source. |
| D9 | `pkg/driver/lunid.go:~123`, `snapid.go:~253` | `strings.Join(parts[1:], "")` drops the underscores. | `vp_my_vol` decodes to `myvol`. |

## 3. Security

| # | Location | Defect |
|---|---|---|
| S1 | `pkg/rest/rproxy.go:249` **[verified]** | `InsecureSkipVerify: true` is hard-coded with no way to turn verification on, and admin Basic-auth credentials are sent on every request. Anyone able to intercept the connection can impersonate the appliance and capture them. |
| S2 | `pkg/node/nodeplugin.go:~89,196`, `pkg/rest/rproxy.go:~110,115` | At DEBUG level, full CSI requests are logged, including the `Secrets` map, and so are REST request bodies, including the CHAP `OutgoingUser.Password`. The node DaemonSet ships with `--loglevel=DEBUG`. |
| S3 | `pkg/rest/storageapi.go`, `targets.go`, `clones.go`, `pools.go` | URL paths are built with `fmt.Sprintf` and no `url.PathEscape`, and the `vp_` payload is not validated (`lunid.go:~107`). A crafted `volumeHandle` such as `vp_x/snapshots/y` can reach other appliance resources. This requires permission to create PVs. |
| S4 | `deploy/joviandss/joviandss-csi-controller.yaml:181,194` **[verified]** | The container has two `securityContext` keys, and the second (`privileged: true`) overrides the hardened first one. The controller only talks REST, yet it runs privileged with `hostNetwork: true`. |
| S5 | Controller and node ClusterRoles | The roles are broader than the plugin needs:<br>• Both can get and list Secrets cluster-wide.<br>• The controller can create and delete CRDs.<br>• The node role can update nodes.<br>• The controller role has a duplicate `persistentvolumes` rule and lists the obsolete `csinodeinfos`. |
| S6 | Node manifest | It mounts the host root `/` with both `privileged` and `SYS_ADMIN`. It also mounts `/etc/iscsi` from the host directory `/system/iscsi`, which is created empty, so the host's initiator name is not visible. |

## 4. CSI spec and behavior defects

### Node

- **N1** `targetapi.go:~492-497`: when the device does not appear in time, the code runs `-o delete` but never logs out, and then `starget` is removed. The iSCSI session is leaked for good, because NodeUnstage finds no `starget` and returns OK.
- **N2** `targetapi.go:~519-535`: UnStage ignores errors from logout and delete and always returns nil, so failed logouts are leaked.
- **N3** `targetapi.go:466` vs `:460` **[verified]**: Stage logs in with `-p t.Portal` (IP only), but the device path is built from `IP:port`. With any non-default port, staging always fails.
- **N4** `targetapi.go:383`: NodePublish does not check whether the target is already mounted, so retries stack mounts on top of each other.
- **N5** `targetapi.go:~415-419`: `PathExists` returns `(true, err)` for a corrupted mount (ENOTCONN or EIO), and Unpublish then returns without unmounting.
- **N6** `nodeplugin.go:~17-21`: the node advertises the `UNKNOWN` RPC capability.
- **N7** `nodeplugin.go:~48-52`: NodeExpandVolume does nothing and returns success, even though EXPAND_VOLUME is not advertised.
- **N8** `nodeplugin.go:~248-257`: an unmount failure is returned as `Aborted`. It should be `Internal`.
- **N9**: there is no per-volume locking anywhere in `pkg/node`, so concurrent Stage and Unstage of the same volume can interleave.
- **N10** `targetapi.go:~188-190`: only `addrs[0]` is used, with no multipath or portal failover.
- **N11**: the default fsType is inconsistent. Stage uses ext3 (stored but not used), while Publish uses ext4.
- **N12** `targetapi.go:~107,374`: directories are created with mode `0640`, which has no execute bit. This only works because the plugin runs as root.

### Controller

- **K1** `controllerplugin.go:~325`: when no capacity range is given, a 0-byte volume is requested. Clone and restore return `CapacityBytes = 0`, and a requested size larger than the source is ignored. A request where required > limit is never rejected.
- **K2** Pagination (`driver.go:~345-400`, `snaptoken.go`, `restwrappers.go`) is wrong in several ways even after C2 is fixed:
  - The next page starts from `lres[maxret-1]`, the last item already returned, so that item appears twice.
  - Positions are compared as base64 strings, which does not follow the real order.
  - `token.BasedID()` is called where `DropBasedID` was intended.
  - The token records the wrong page, so items can be skipped.
  - A non-zero `dc` is never copied.
  - `RestSnapshotShortEntryBasedID` encodes the volume name twice, so the IDs of a volume's snapshots collide.
- **K3** `controllerplugin.go:~946-957`: ValidateVolumeCapabilities has three problems:
  - It sets `Confirmed` even for unsupported modes, because the `if` body is empty.
  - It returns `cp.vCap`, which is never assigned.
  - It checks `vcap` where it should check `vCtx`.
- **K4** `controllerplugin.go:~765-785`: when ListSnapshots by `snapshot_id` finds no snapshot, it returns InvalidArgument. The spec requires an empty list. The code also tests `err` instead of `rErr`, so ListAllSnapshots failures are ignored.
- **K5** `driver.go:~234,246`: `RestErrorResourceBusyVolumeHasSnapshots` reaches the default case, so the CO gets `Internal` instead of `FailedPrecondition`.
- **K6** `driver.go:~160-164`: when an intermediate snapshot is left over from an earlier failed clone, it is deleted (the result is ignored) and `ResourceExists` is returned. The CO receives AlreadyExists, a terminal error, for what was just a retry.
- **K7** `driver.go:~648`: rollback calls `DeleteTarget(tname)`, but the target was created as `iqn`, so the target leaks. An already-attached LUN is accepted whatever its RO/RW mode, and `node_id` is ignored, so there is no initiator restriction.
- **K8** `driver.go:~554-576`: after the dependent clones are deleted, the snapshot delete is never retried, so the call always returns ResourceBusy.
- **K9** `controllerplugin.go:~567`: a negative `max_entries` returns `Internal`. It should be `InvalidArgument`.
- **K10** `controllerplugin.go:~383,387`: the AlreadyExists messages print the requested size where they should print the actual size.

### REST client

- **R1** `rproxy.go:~82-157`: the `tries` setting and every address after the first are never used, because `active_addr` is always 0. There is no retry or failover, even though `doc/configuration.md` promises both.
- **R2** `targets.go:~88-97`: after an HTTP 202, the code polls `GetTarget` 10 times with no sleep and no `break`. The polls finish before the asynchronous target exists, so ControllerPublish fails.
- **R3** `errors.go:94` **[verified]**: `targetDneMsgPattern` is the hard-coded test string `Target iqn.20215:test2asd not exists.`, so real "target not found" errors are never recognized.
- **R4** `storageapi.go:~52-54`: the response body is used as a format string (`fmt.Sprintf(string(body))`, then `Warnf(msg)`). Also, GetVolume, GetTarget and GetVolumeSnapshotClones convert transport errors into `RequestMalfunction`, which hides the error class.
- **R5** `rest.go:~85-96`: `String()` builds a string, throws it away and returns `Addrs[0]`. It panics if the list is empty.
- **R6** `clones.go:~79`: the arguments of the CreateClone debug log are in the wrong order.

### Plugin server and identity

- **P1** `pluginserver.go:~139-149`, `identityplugin.go:~41-57`: if the controller fails to set up, the server only logs a warning and keeps serving. Probe always reports ready, and CONTROLLER_SERVICE is advertised even by the node-only process.
- **P2** `main.go`, `pluginserver.go:~169-177`: SIGTERM is not handled. There is no `GracefulStop`, the socket is not removed on exit, and errors from `Run` are ignored, so the process exits with code 0.
- **P3** `main.go:~431`, `common.go:~273`: log files are opened without `O_APPEND` or `O_TRUNC`, so a restart overwrites them from offset 0 and leaves stale lines after the new content.

## 5. Deploy, docs, build

- **B1** `deploy/container/scripts/iscsiadm:4`: the wrapper defaults to `/usr/local/sbin/iscsiadm`, the Talos path. The node manifest does not set `ISCSIADM_HOST_PATH`, so iSCSI login fails on Ubuntu and RHEL hosts.
- **B2** `joviandss-csi-node.yaml:~73,77,142`: the socket path is wrong in three ways:
  - The plugin socket sits under `plugins_registry/`. It belongs under `plugins/<driver>`.
  - The directory name `com.open-e.joviandss.csi` does not match the driver name `iscsi.csi.joviandss.open-e.com`.
  - The preStop hook deletes the wrong directory.
- **B3** Images:
  - The driver image is `opene/joviandss-csi:latest` with `imagePullPolicy: Always`.
  - The sidecars come from the frozen `quay.io/k8scsi` registry and are old: registrar v1.0.2, provisioner v2.1.2, attacher v3.1.0.
  - `deploy/dev` uses `k8s.gcr.io`.
- **B4** `deploy/joviandss/crdt/*.yaml`: the CRDs still serve the deprecated `snapshot.storage.k8s.io/v1beta1`, while the snapshot-controller is v7.0.1.
- **B5** Config keys disagree between the docs and the code:
  - `doc/configuration.md` puts `pool` under `endpoint`, but the code reads only the top-level `pool`, so requests go to `pools//volumes`.
  - The docs and `cfg.yaml` use `logpath`, but the code reads `logfile`.
  - The docs misspell `idletimeout` as `iddletimeout`.
  - yaml.v3 ignores the `json` tags, so `namelen` and `passlen` are never read.
  - A config with no `loglevel` makes the plugin exit.
- **B6** `Makefile`: the `clean` target is corrupted (`-rm -rf _outpusudo podman push $(IMAGE_TAG_UBUNTU)`), and `IMAGE_TAG_UBUNTU` is undefined.
- **B7** `deploy/container/Dockerfile` uses the EOL `centos:latest` image.
- **B8** `.travis.yml` targets Go 1.12/1.13 and has the wrong `go_import_path`, so CI cannot build this module.
- **B9** Dead code that would crash if it were ever called:
  - `pkg/controller/lunid.go_tmp`
  - `parseOrigin`
  - `lockVolume`, which writes to a nil map
  - `getVolume`
  - `getRandomName` and `getRandomPassword`
  - `NewSnapshotTokenFromStr`

## 6. CLI (`cmd/`)

- `publishVolume.go` never sets `req.Readonly` or `req.NodeId`.
- `listSnapshots.go` does not pass the `--token` flag on to the request.
- `createSnapshot.go:~30-34` marks its required flags on `createVolumeCmd`, so they are never enforced.
- In `createVolume.go:~87-98`, `--srq` overwrites `--slm`.
- Every subcommand ignores the error from `SetupControllerPlugin`.

## Checked and found OK

- Every `exec.Command` call passes an argument vector with no shell, so there is no shell injection.
- `-T t.Iqn` is correct, because the publish context carries the full per-volume IQN.
- HTTP response bodies are closed, and a client timeout is set.

## Suggested order of work

1. Fix the crashes in section 1. Most need only a nil check before calling a method on `RestError`, or using the right error variable.
2. Fix D1–D4: raw block, read-only mounts, NodeStage idempotency and the restore error shadowing.
3. Fix S1 and S2: make TLS verification the default (with an opt-out setting) and keep secrets out of the logs.
4. Add a CI job that runs `go vet` and `csi-sanity`. `vet` would have caught the shadowed variables, and csi-sanity covers most of the issues in section 4.
