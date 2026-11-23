# JovianDSS Kubernetes CSI plugin: project review

Date: 2026-11-08. Reviewed: `joviandss-kubernetescsi` at `master` = `origin/master` =
`11ffbfd` plus the untracked work files, and the parts of `kub-testing` and
`jdss-rest-api-spec` that the plugin work depends on. Report only; nothing was
edited.

What could not be checked here: this machine has no Go toolchain, so no build,
`go vet`, `gofmt` or `govulncheck` ran (kub-testing's notes say builds happen on
kub-gate with Go 1.26). No cluster or appliance was contacted. GitHub was not
reachable (`gh` not logged in), so the visibility of the `kub-testing` repository
is unknown.

## 1. Verdict in short

The project is in a transition. The code is a 2024 code base with known crash
bugs, no automated tests and no CI. The last month built the scaffolding that was
missing: three design documents, an issue tracker with stable IDs, recorded REST
evidence from real appliances, and a csi-sanity tool chain. The plan (0002 REST
compliance, then 0003 controller crashes, then the node fixes) is the right order.

The five things that matter most, in order:

1. **A possible silent data-loss path on the node is unverified.** The merged
   iSCSI node code probably reports a successful stage without mounting anything
   (kub-testing N13-N15), so a pod would write to the node's disk. No test has
   reached that code since the merge (T10). Nothing built from `11ffbfd` should be
   given to users before T10 runs.
2. **Cluster credentials are on GitHub.** `kub-testing` commit `456642f` holds
   `talosconfig` and the two Talos machine configs with 23 secret-shaped values
   (CA keys, tokens, client key). It is on `origin/main` of
   `github.com:open-e/kub-testing`. The files were removed later (`6101eec`), but
   history keeps them. Treat the cluster's credentials as exposed: rotate them,
   then rewrite history.
3. **A month of design knowledge exists only as untracked files.** Design 0002
   (~1500 lines), 0003, the architecture notes, the Nov 7 review, the backlog,
   and in `kub-testing` the issue tracker, status file and csi-sanity tooling are
   all uncommitted. One disk failure loses them.
4. **A user following the public docs today cannot get a working install.** The
   README links to a `doc/` folder that no longer exists; the install steps apply
   manifest paths that moved in the merge; the configuration example puts `pool`
   under `endpoint`, where the code never reads it, so every REST call would go to
   an empty pool; the docs create only the controller Secret while the node
   DaemonSet requires a second one, so node pods never start; and the namespace
   manifest lacks the Pod Security labels the docs promise, so Talos rejects the
   privileged node pods. On top of that the manifests pull sidecars from
   2019-2021 and an unpinned `latest` plugin image.
5. **Stale status text is spreading across documents.** Design 0001 still says
   "merge in progress, nothing pushed"; the architecture notes still call the
   crash-fix design "0002". Several places record status by hand, and they drift.

## 2. Findings by area

Severity: H = fix before the next delivery, M = fix in this development cycle,
L = when convenient.

### 2.1 Code health

| # | Sev | Finding | Evidence |
|---|---|---|---|
| CH1 | H | No automated tests at all: 0 `_test.go` files for 11,304 lines of Go. Design 0002's REST decoding tests would be the first tests in the repository. | `find . -name '*_test.go'` is empty |
| CH2 | H | 23 crash findings (C1-C23), most still present, and no panic recovery in the gRPC server, so any nil dereference kills the controller or node process. | `kub-testing/ISSUES.md` §3.1; `pkg/pluginserver/pluginserver.go` has no `recover` |
| CH3 | M | 6 `panic` calls in library code, reachable from CSI requests: logger lookup, volume and snapshot name decoding. The other 27 are in the CLI, where a panic is acceptable. | `pkg/common/common.go:164`, `pkg/driver/lunid.go:162,170`, `pkg/driver/snapid.go:149,156,163` |
| CH4 | M | Dead code kept after the node rework: `pkg/node/targetapi.go` (566 lines, 7 methods on `*Target`, no callers outside the file) and `pkg/node/target.go`. The C12 nil-logger bug lives in it. `pkg/controller/lunid.go_tmp` is a tracked stray copy. | `grep` of the 11 functions; `docs/architecture/02-code-map.md:184`; `git ls-files pkg/controller/lunid.go_tmp` |
| CH5 | M | Commented-out code: 443 lines (3.9% of the code). Worst files: `cmd/controller/createSnapshot.go` 21%, `cmd/controller/listSnapshots.go` 19%, `pkg/rest/storageapi.go` 15%, `cmd/controller/listVolumes.go` 12%, `pkg/rest/targets.go` 11%. 37 `TODO`/`FIXME` markers. | heuristic: `//` lines that look like Go statements (calls, assignments, `if`/`for`/`return`, closing braces), counted per file |
| CH6 | M | Duplicated logic between the two drivers: the same bugs exist twice (D5, D8, C9 "two copies" in `iscsi_driver.go` and `nfs_driver.go`), and the REST layer repeats the volume code for NAS volumes. Every fix must be made twice. | `kub-testing/ISSUES.md` D5, D8, C9 |
| CH7 | M | Protocol chosen at compile time through a linker flag (`PluginProtocolCompileString`), giving two binaries, two driver names and two images. A cluster cannot run both (MF7). Worth its own design document later. | `Makefile` `-X ...PluginProtocolCompileString`, `pkg/common/common.go:191-207` |
| CH8 | L | Debug prints to stdout remain in the REST client and logger setup. 0002 stage 4 (A8) removes the REST ones. | `pkg/rest/rproxy.go:152,156`, `pkg/common/common.go:103` |
| CH9 | M | Dependencies are from April 2024 (`grpc v1.63.2`, `x/net v0.24.0`, `mount-utils v0.30.0`, CSI spec `v1.9.0`). Not scanned: `govulncheck` needs network and Go. Run it on kub-gate. | `go.mod` |

### 2.2 Build, CI and release

| # | Sev | Finding | Evidence |
|---|---|---|---|
| BR1 | H | No CI. `.travis.yml` is dead: Go 1.12/1.13, target `make joviandss` (does not exist), import path in the wrong case. No `.github/` workflows. Nothing runs `vet`, `gofmt`, tests or a build on push. | `.travis.yml`; `Makefile` targets |
| BR2 | M | Build targets run `go mod tidy`, so a build can rewrite `go.mod`/`go.sum` and needs network. `-a` forces a full rebuild each time. The `container` target builds the same Dockerfile twice instead of tagging once. | `Makefile` `cli`, `plugin`, `container` |
| BR3 | H | NFS image name mismatch: the NFS manifests pull `opene/joviandss-csi-nfs:latest`, but no Makefile target produces a name with a protocol suffix (`IMAGE_NAME=joviandss-csi` for both protocols). Building NFS would overwrite the iSCSI tag. | `deploy/joviandss/nfs/*.yaml:180,94`; `Makefile` `IMAGE_TAG_*` |
| BR4 | L | Default `PROTOCOL_TYPE` is `nfs` while the README and the only tested path are iSCSI. | `Makefile` line `PROTOCOL_TYPE ?= "nfs"` |
| BR5 | M | No release since `v1.1.0` (2024-04-12); 27 commits since, including the merge. No changelog. Manifests deploy `:latest`, so a deployment cannot be reproduced or rolled back by tag. | `git tag`; `git describe` = `v1.1.0-27-g11ffbfd` |
| BR6 | M | Container images: `centos:stream9` is unpinned (no digest); the old `deploy/container/Dockerfile` uses `centos:latest`, whose repositories went away in 2021, and is unused by the Makefile. The dev image copies an 11.9 MB `csc` binary that is committed to git. | `deploy/container/*.Dockerfile`; `deploy/container/bin/csc` |

### 2.3 Deployment manifests

| # | Sev | Finding | Evidence |
|---|---|---|---|
| DM1 | M | Sidecars moved to `registry.k8s.io` (backlog 31, first half) but the versions are old: `csi-node-driver-registrar:v1.2.0` (2019), `csi-attacher:v3.1.0` and `csi-provisioner:v2.1.2` (2021), `csi-snapshotter:v6.2.3`, `snapshot-controller:v7.0.1`. | `deploy/joviandss/**/*.yaml` `image:` lines |
| DM2 | M | Snapshot CRDs come from external-snapshotter v4.0 (`controller-gen v0.4.0`, approval PR 419) and still serve `v1beta1`, while the snapshot-controller is v7.0.1. Regenerate the CRDs from the same release as the controller. | `deploy/joviandss/crdt/volumesnapshots.yaml:7-8,156` |
| DM3 | L | RBAC still grants `csi.storage.k8s.io/csinodeinfos` (API removed in Kubernetes 1.17) and cluster-wide `secrets get,list` that the plugin does not use. | `deploy/joviandss/iscsi/joviandss-csi-controller.yaml:44-70` |
| DM4 | L | `deploy/dev/` duplicates `deploy/joviandss/` with older images (`k8s.gcr.io/.../snapshot-controller:v4.0.0`, a dead registry) and `opene/joviandss-csi-dev:latest`. | `deploy/dev/snapshot-controller.yaml:105` |
| DM5 | L | The node DaemonSet is `privileged`, `hostNetwork`, and mounts the host root at `/host` read-write for the `iscsiadm` chroot, plus Talos-specific paths. The wrapper calls the host binary at the fixed Talos path `/usr/local/sbin/iscsiadm` unless `ISCSIADM_HOST_PATH` is set, and no manifest sets it, so the node image only works on Talos as shipped. Normal for an iSCSI driver on Talos, but it is a trust boundary and a platform limit that the docs should state. | `deploy/joviandss/iscsi/joviandss-csi-node.yaml:66,90,137-138,171-180`; `deploy/container/scripts/iscsiadm:3-11` |
| DM6 | L | The config Secret carries the REST admin password; the appliance has no lesser REST role (0002 "Not possible today"). The tracked example configs use placeholder passwords, which is fine; `deploy/dev/cfg.yaml` carries a lab address. | `deploy/cfg/cfg.yaml`, `deploy/dev/cfg.yaml` |
| DM7 | H | Both node DaemonSets mount a Secret `jdss-node-cfg` that is not marked `optional`, and the install docs create only `jdss-controller-cfg`. A user following the docs gets node pods stuck in `ContainerCreating`. The node does not talk REST, so the Secret is probably unused: mark it optional or drop it. Backlog 24, still open, tracked nowhere. | `deploy/joviandss/iscsi/joviandss-csi-node.yaml:153-154`, `nfs/...:142-143`; `docs/install.md:59` |
| DM8 | M | The `csi-controller` container declares `securityContext` twice: a hardened block (no privilege escalation, read-only root) and, 13 lines later, `privileged: true`. YAML parsers keep the last one, so the hardened block is dead and the controller runs privileged although it only talks REST. Same in the NFS and dev manifests. Known as S4 in the 2026-10-02 review, not re-checked since. | `deploy/joviandss/iscsi/joviandss-csi-controller.yaml:181-183,194-195` |
| DM9 | M | The iSCSI example PVC asks for `ReadWriteMany`, but the controller advertises only single-node access modes, and a block device with ext4 must never be mounted by two nodes. The NFS build advertises the same single-node modes, so the one protocol built for shared access cannot offer `ReadWriteMany` at all. Fix the example to `ReadWriteOnce`; add multi-node modes to the NFS driver by design. | `deploy/example/nginx-pvc.yaml:7`, `deploy/example/nfs/nginx-pvc.yaml:7`; `pkg/controller/controllerplugin.go:81-85` |

### 2.4 Documentation

| # | Sev | Finding | Evidence |
|---|---|---|---|
| DO1 | H | README links to `doc/install.md`, `doc/install-nfs.md`, `doc/configuration.md`; the folder is now `docs/`. A first-time user fails at the first link. Also `deploy/examples/nfs` should be `deploy/example/nfs`. Planned as 0001's "references commit" (R4) and backlog 23/27, not done. | `README.md`; `ls doc` fails |
| DO2 | M | Design 0001's status callout says "merge in progress ... nothing is committed or pushed". The merge is committed and pushed (`11ffbfd` = `origin/master`) and T1 ran on 2026-11-02. Status should move to implemented, with the references commit listed as the remaining part. | `docs/design/0001-...md:4-6`; `git rev-list origin/master...master` = 0 0 |
| DO3 | M | Architecture notes (8 files, 2026-11-02, untracked) are good but already stale: the README and chapter 07 call the crash-fix design "0002" (it is 0003 since the split on 2026-11-02), and 07 does not know the Nov 5-7 findings (F9-F28, N13-N15). | `docs/architecture/README.md` "design 0002 plans the fixes"; `07-...md` "0002 needs re-anchoring" |
| DO4 | L | Design 0003 names its design items D1-D7, while the issue tracker uses D1-D14 for data-loss defects. 0002 already warns about this (Table 23). When 0003 is re-anchored, rename its items to A-numbers like 0002. | `docs/design/0003-...md:148-217`; `0002:1555` |
| DO5 | L | Design 0002 is close to final: one open question left (7, snapshot filtering by kind); one dangling reference (stage 3's F16 bullet does not link to the question). | 17th consistency review, 2026-11-08 |
| DO6 | H | The configuration guide's example does not work: it puts `pool` under `endpoint`, but the code reads `pool` only at the top level, so the plugin runs with an empty pool name. The example also uses `logpath`, while the code reads `logfile` (logs silently go to stdout). The shipped example `deploy/cfg/cfg.yaml` has a top-level `pool` and so works, but also carries the ignored keys. Backlog 19, still open. | `docs/configuration.md:17,26`, `configuration-nfs.md:17,26`; `pkg/common/common.go:73-81` |
| DO7 | M | Two documented settings do nothing: `tries` is stored but never read (one HTTP attempt per call, no retry), and only the first entry of `endpoint.addrs` is ever used (`active_addr` is set to 0 and never changes), so there is no failover between appliance addresses. Either implement them or document them as reserved. | `pkg/rest/rproxy.go:91,274,281`; `docs/configuration.md:48,51` |
| DO8 | H | Install guides apply `deploy/joviandss/joviandss-csi-controller.yaml` and `-node.yaml`, which moved to `iscsi/` and `nfs/` in the merge; and they claim the namespace manifest carries Pod Security labels that it does not have (`docs/pod-security.md:174-176` admits it). On Talos, which enforces `baseline`, the node pods are rejected unless the user also applies the admission exemption that `docs/talos.md` no longer mentions. | `docs/install.md:25-28,71,73`, `docs/install-nfs.md:22-25`, `docs/talos.md:118-121`; `deploy/joviandss/namespace.yaml` |
| DO9 | — | Full list of documentation mismatches: §3.2 (delegated audit, spot-checked). | |

### 2.5 Repository hygiene

| # | Sev | Finding | Evidence |
|---|---|---|---|
| RH1 | H | Uncommitted knowledge: `CLAUDE.md`, `docs/architecture/` (8 files), `docs/design/0002`, `0003`, `review/2026-11-07-code-review.md`, `review/backlog.md`, and the `0001` edit. In `kub-testing`: `ISSUES.md`, `project-status.md`, `docs/csi-sanity/`, `docs/rest-api/`, `tests/csi-sanity/` tooling, five modified docs. | `git status` in both repos |
| RH2 | M | `tmp/` (313 MB, an image tarball) is untracked and not ignored. Design 0001 already warns "never `git add -A`"; one slip commits 313 MB. Add `tmp/` to `.gitignore`. | `du -sh tmp`; `.gitignore` |
| RH3 | M | Tracked files that do not belong: `.gitconfig` (personal pager config, arrived with `11ffbfd`), `pkg/controller/lunid.go_tmp`, `deploy/container/bin/csc` (11.9 MB ELF), `test/env/` and `test/scripts/` (Vagrant and Fabric scripts for Fedora 29 and CentOS, last touched 2020-06-08, superseded by `kub-testing`), `deploy/container/Dockerfile` (2020). | `git ls-files`; `git log -1 -- test/` |
| RH4 | L | Branches: `origin/helm` unmerged while the README promises Helm charts; `origin/node-volume-mounting-rework` stale; local `nfs-support` is fully merged and can go. | `git branch -a`; `git rev-list master...nfs-support` = 5 0 |

### 2.6 Security

| # | Sev | Finding | Evidence |
|---|---|---|---|
| SE1 | H | `kub-testing` commit `456642f` ("talos: 1cntr: init") adds `talosconfig` (3 secret-shaped values: CA, client cert, client key) and `controlplane.yaml` (14) and `worker.yaml` (6) with cluster CA keys and tokens. It is on `origin/main` at `github.com:open-e/kub-testing`; local `main` equals `origin/main`. `6101eec` removed the files from the tree, not from history. The `.gitignore` rule for `secrets.yaml` was added one commit earlier (`f7f8046`), so the risk was known. Whoever can read that repository can administer the lab cluster. | `git branch -r --contains 456642f` → `origin/main`; counts from `git show 456642f:<file>` (values not printed) |
| SE2 | M | Plugin security findings S1-S6 are known; 0002 stage 4 covers TLS verification on by default, no request dumps, `url.PathEscape`, reply size limit, removal of secret-leaking debug prints, access and refresh tokens as secrets. Remaining after 0002: Debug-level node logs may contain secrets (kub-testing rule), and there is no CHAP. | `0002` A8-A15; `kub-testing/ISSUES.md` §3.3 |

### 2.7 Process and testing

Strengths worth keeping:

- Design documents in a fixed format, with resolved questions struck through
  and decisions recorded with the user's words.
- One issue tracker (`kub-testing/ISSUES.md`) with stable IDs: C 23, D 14, N 25,
  R 17, S 6, MF 14, F 29, T 10, DEC 10, DOC 26. Every design item points at an ID.
- REST behaviour taken from recorded appliance responses (`jdss-rest-api-spec/empirical/`)
  instead of guesses; 0002's compliance rules SC1-SC7 are checkable.
- A repeatable csi-sanity cycle (`dev_cycle.py`) with a parity baseline (T1).

Weaknesses:

- Status is written by hand in five places (design callouts, architecture 07,
  `review/backlog.md`, `project-status.md`, `ISSUES.md`), and they drift (DO2, DO3).
  Keep `ISSUES.md` as the only live tracker; make everything else a dated
  snapshot that points to it, and retire `backlog.md` into `ISSUES.md`.
- The node path has not been exercised since the merge (T10), NFS has never run
  live (T6; 0002 now requires an NFS cycle), and unit tests start only with 0002.
- Line numbers are the main anchor in reviews and designs, and every commit moves
  them. 0002 already links to GitHub at `11ffbfd`; do the same elsewhere.

## 3. Delegated audits

Two read-only audits ran as separate agents with only the repository and a
brief; their tables are reproduced here with light editing. I re-checked the
rows that were new to me (DM7, DM8, DM9, DO6, DO7, DO8 above) against the files
and they hold. Neither agent could run Go tools either.

### 3.1 Backlog items 1-38: current status

Totals: DONE 10, PARTIAL 10, OPEN 16, OBSOLETE 1, UNVERIFIED 1. "Tracked" means
a design document or the 2026-11-07 review assigns the item.

| # | Item | Status | Evidence |
|---|---|---|---|
| 1 | Drop `go mod tidy`/`go get` from build | PARTIAL | `go get` gone; `go mod tidy` still in `cli` (`Makefile:64`) and `plugin` (`:68`). |
| 2 | `tidy`/`deps` targets | OPEN | No such targets. |
| 3 | `vet`/`fmt-check`/`check-tidy` targets | OPEN | None; 0002 and 0003 run `go vet` by hand on kub-gate. |
| 4 | `MODULE` variable in `-X` flags | OPEN | Module path written out at `Makefile:13-14,17-18`. |
| 5 | `cli` built without `-X Version` | DONE | `cli` uses `$(FLAGS)` (`Makefile:63-65`). |
| 6 | Drop `-a` | OPEN | `Makefile:10,16`. |
| 7 | `git describe --always --dirty` | PARTIAL | `Makefile:4` has `--long --tags` only. |
| 8 | Broken `clean` target | DONE | `Makefile:76-77` `go clean -r -x`. |
| 9 | `.PHONY` lists missing targets | DONE | Only existing targets; nit: `check-protocol-type` missing from `.PHONY`. |
| 10 | Reconcile with nfs-support Makefile | DONE | `11ffbfd` took that Makefile. |
| 11 | `go` directive | DONE | `go.mod:3` `go 1.26`, decided in 0001 R5 / question 3. |
| 12 | go.sum completeness | UNVERIFIED | No Go here; `go.sum` has entries for all 21 modules. |
| 13 | `gofmt` after import change | OPEN (partly verified) | Two import runs gofmt would reorder: `cmd/controller/controller.go:26-28`, `pkg/node/targetapi.go:36-37`. |
| 14 | Commented-out old-path imports | PARTIAL | 8 of 10 remain (`cmd/node/node.go:11`, `cmd/controller/listVolumes.go:13`, `listSnapshots.go:13`, `deleteVolume.go:27`, `app/joviandssplugin/main.go:12`, `pkg/common/common.go:32`, `pkg/controller/controllerplugin.go:44,47`). |
| 15 | 2024 dependencies / `govulncheck` | OPEN | `go.mod:10-12,20` unchanged; not assigned anywhere. |
| 16 | Nil-dereference crashes | PARTIAL, tracked | C6 fixed by the merge, C12 moved to dead code; the rest present. 0003 D2-D7 covers F2-F5, C6, C20; 0002 A8 covers C3, C4, C5, C17, R5; C2, C8, C9, C11, C13-C16, C18 are only in the 2026-11-07 review. |
| 17 | `main.go` ignores `GetPluginServer` error | OPEN, tracked | `app/joviandssplugin/main.go:124`; 0003 question 3 (deferred). |
| 18 | Log file opened without `O_APPEND` | OPEN, tracked | `main.go:71`, `pkg/common/common.go:95`; 0003 question 3. |
| 19 | `logfile` vs `logpath`; `pool` location | OPEN, tracked | `common.go:75-76`; `docs/configuration.md:17,26,50`; `deploy/cfg/cfg.yaml:2,12`. Review 2026-11-07 B5. See DO6. |
| 20 | yaml tags on endpoint structs | OPEN, tracked | `common.go:50-70` json tags only. Review B5. |
| 21 | `PluginName` hard-coded; `nfs:` not parsed | DONE | `common.go:80,190-200`. |
| 22 | Docs vs Talos patches / namespace labels | PARTIAL | Patches done in `2d94828`; `namespace.yaml` still has no labels while three docs say it does. Not tracked. See DO8. |
| 23 | Duplicate `doc/`; README links | PARTIAL, tracked | `doc/` deleted; README links dead. 0001 "references commit", not done. |
| 24 | Node DaemonSet Secret `jdss-node-cfg` | OPEN | Not `optional`; docs never create it. Not tracked. See DM7. |
| 25 | `configuration.md` gaps and typo | OPEN | `prot`, `user`, `pass` undocumented (`:23-25`); `iddletimeout` at `:52`. |
| 26 | `compile.md` vs Makefile | DONE | Matches since `11ffbfd`. |
| 27 | README NFS config link | OPEN | `README.md:27` points at `doc/configuration.md`. |
| 28 | `install.md` vs Talos-only node manifest | OPEN | `docs/install.md:10`; node manifest `:155-177`. |
| 29 | Typos, `[TODO]`, near-duplicate docs | OPEN | `[TODO]` at `docs/install.md:64`, `install-nfs.md:61`. |
| 30 | README "tested on Talos OS 1.6" | OPEN | No 1.14 result recorded in this repo. |
| 31 | Sidecar images | PARTIAL | Registry moved; versions old; compatibility unchecked. Not tracked. See DM1. |
| 32 | Manifest layout; `classes.yaml` duplicate | PARTIAL | Layout done; `iscsi/classes.yaml` still duplicates the controller manifest; docs name the old paths. |
| 33 | Test with Talos workload isolation | OPEN | `docs/talos.md:86-88`. Not assigned. |
| 34 | Where `review/` lives | OPEN | One review tracked, two untracked. |
| 35 | Plan the nfs-support merge | DONE | `11ffbfd`; the references commit is pending. |
| 36 | `ResourceVolume.Encryption` is `string` | OPEN, tracked | 0002 F1 / A1 (remove the field). |
| 37 | Snapshot `creation` layout | OBSOLETE | Verified correct against recorded replies (review 2026-11-07 §2d; 0002 Relationship). The real defect is F9. |
| 38 | Delete of unknown ID returns `InvalidArgument` | OPEN, tracked | 0003 D6 (F5); review K18 for DeleteVolume. |

### 3.2 User-facing documentation versus the repository

Eight documents were checked statement by statement (README, `install*.md`,
`configuration*.md`, `compile.md`, `talos.md`, `pod-security.md`). Severity:
BLOCKS-USER = a user following the text fails; WRONG = incorrect, recoverable.

| Doc | Says | Repository has | Severity |
|---|---|---|---|
| `docs/install.md:25-28`, `install-nfs.md:22-25`, `talos.md:118-121` | namespace manifest sets Pod Security `privileged` labels | `deploy/joviandss/namespace.yaml` has no labels; Talos enforces `baseline` | BLOCKS-USER |
| `docs/install.md:71,73` | apply `deploy/joviandss/joviandss-csi-{controller,node}.yaml` | files are under `deploy/joviandss/iscsi/` | BLOCKS-USER |
| `docs/install.md:42,62,66`; `install-nfs.md:39,59,63` | same old controller path for the binding, config name and Secret | moved to `iscsi/` and `nfs/`; `install-nfs.md` step 5 uses the right path, so it contradicts itself | WRONG |
| `docs/configuration.md:26,50`; `configuration-nfs.md:26` | `pool` under `endpoint` | top-level field only (`pkg/common/common.go:76`); `endpoint.pool` is dropped | BLOCKS-USER |
| `docs/configuration.md:17,44`; `configuration-nfs.md:17` | `logpath` key | code reads `logfile` (`common.go:75`); `logpath` is only a CLI flag | WRONG |
| `docs/configuration.md:51` | `tries` = number of attempts per request | stored, never read; exactly one HTTP call (`rproxy.go:132`) | WRONG |
| `docs/configuration.md:48` | `addrs` = list of addresses to use | only the first is used (`rproxy.go:91,274`) | WRONG |
| `README.md:15,18,25,26,27` | links to `doc/...` | folder is `docs/` (dropped in `2d94828`) | WRONG |
| `README.md:27` | NFS configuration document = `doc/configuration.md` | `docs/configuration-nfs.md` | WRONG |
| `README.md:83` | `deploy/examples/nfs` | `deploy/example/nfs` | WRONG |
| `docs/compile.md:13`; `install-nfs.md:68-70` | `make dev PROTOCOL_TYPE=nfs` builds the NFS image | Makefile never adds a protocol suffix; NFS manifests pull `opene/joviandss-csi-nfs:latest`, which nothing builds | WRONG |
| `docs/pod-security.md:71,112` | labels live in `namespace.yaml` | the same document says at `:174-176` that they are not there yet | WRONG |
| `docs/install.md:30`, `install-nfs.md:27` | external-snapshotter CRD link | stray backtick inside the URL | COSMETIC |
| `docs/install.md:8` | `iscsi\_tcp` | backslash renders literally | COSMETIC |
| `docs/configuration.md:52` | `iddletimeout` | `idletimeout` | COSMETIC |

Share of checkable statements that were correct: README about 8 of 15;
`install.md` 17 of 22; `install-nfs.md` 18 of 22; `configuration.md` 11 of 16;
`configuration-nfs.md` 10 of 12; `compile.md` 7 of 7; `talos.md` 12 of 13;
`pod-security.md` 12 of 13. Driver names, Secret and file names, pod counts,
snapshot-controller setup, Talos patch contents, the `vp_` naming rule and all
example paths under `deploy/example/` are correct.

## 4. Numbers

| Measure | Value |
|---|---|
| Go lines (`app`, `cmd`, `pkg`) | 11,304 |
| Test files | 0 |
| Crash findings still listed (C) | 23 |
| Data-loss findings (D) | 14 |
| Node findings (N) | 25 |
| Defects found by tests (F) | 29 |
| `panic` calls in library code / CLI | 6 / 27 |
| Commented-out code lines | 443 (3.9%) |
| Dead node code | 590 lines |
| Last release tag | `v1.1.0`, 2024-04-12 |
| Commits since the tag | 27 |
| Untracked image tarball | 313 MB |
| Backlog items done / partial / open / obsolete / unverified | 10 / 10 / 16 / 1 / 1 |
| Documentation mismatches (blocking / wrong / cosmetic) | 4 / 9 / 3 |
| Secret-shaped values in the pushed `kub-testing` commit | 23 |

## 5. Recommended order

1. **Safety:** run T10 (one real iSCSI PVC and pod) to confirm or clear N13-N15
   before anyone relies on the merged node code.
2. **Credentials:** rotate the lab cluster's Talos secrets, then rewrite
   `kub-testing` history to drop `456642f`'s files and force-push. Check the
   repository's visibility first; if it is public, assume the secrets were copied.
3. **Commit the knowledge** in both repositories (your call where `review/` lives,
   backlog 34), and the hygiene fixes: ignore `tmp/`, drop `.gitconfig`,
   `lunid.go_tmp`, `csc`, `test/env`, `test/scripts`, the old Dockerfile.
4. **Make the documented install work** (small, independent of 0002, and it is
   what a visitor to GitHub sees today): the 0001 references commit (README and
   install paths, 0001 status), `pool` at top level in the configuration guides
   and the example config, `logfile` instead of `logpath`, the node Secret made
   optional or removed, Pod Security labels in `namespace.yaml`, the duplicate
   `securityContext` resolved, the example PVCs set to `ReadWriteOnce`, and the
   NFS image name settled between Makefile and manifests.
5. **0002 as planned:** `per_page` fix, dead-code commit, v4, response types,
   error mapping, stage 4 security; with the REST unit tests. Then 0003.
6. **Platform refresh:** sidecars and CRDs from one current external-snapshotter
   release, pinned plugin image tags, a GitHub Actions workflow for `vet`,
   `gofmt`, tests and the build, and `govulncheck` on kub-gate.
7. **Later, by design document:** one image for both protocols (MF7), removing the
   driver duplication (CH6, CH7), multi-node access modes for NFS (DM9), and
   real use of `tries` and `addrs` or their removal (DO7).
