# Changelog

All notable changes to the Open-E JovianDSS Kubernetes CSI plugin, newest
first. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/);
versions are the git tags of this repository. Until this file existed, release
notes were published only on the
[GitHub releases page](https://github.com/open-e/joviandss-kubernetescsi/releases),
and only for v1.1.0; those notes are copied below. Changes on `master` that no
tag covers yet are listed under "Unreleased".

**How the tags relate.** The history is not one line. v1.0.2, v1.0.3 and
v1.1.0 all branch from the same commit of 2023-08-20 (the end of the "go
modules" work): v1.0.2 adds one dependency update to it, v1.0.3 is the end of a
development branch of nineteen commits that was never merged as such, and
v1.1.0 is the squashed pull request #5, which contains the work of the v1.0.3
branch, the dependency update of v1.0.2 as a separate commit, and the driver
rework. Only v1.1.0 is an ancestor of `master`. The compare links at the end of
this file show the commits that one tag has and the previous one has not.

## [Unreleased]

Everything on `master` after v1.1.0, including the `nfs-support` branch
(21 commits from April to July 2024), merged on 2026-10-31.

### Added

- NFS driver: NAS volumes, snapshots and clones in the REST layer, a second
  driver implementation next to iSCSI, and NFS stage and publish on the node
  side. Driver name `nfs.csi.joviandss.open-e.com`.
- One binary per protocol, chosen at build time with `PROTOCOL_TYPE=iscsi` or
  `PROTOCOL_TYPE=nfs`. Manifests moved to `deploy/joviandss/iscsi/` and
  `deploy/joviandss/nfs/`.
- Cloning for NFS volumes (NAS clones in the REST layer). Snapshots and clones
  are shared between the two protocols.
- Guides: compilation (`docs/compile.md`), NFS installation
  (`docs/install-nfs.md`), NFS configuration (`docs/configuration-nfs.md`) and
  Pod Security (`docs/pod-security.md`).
- Talos patches in `deploy/talos/`: a `KernelModuleConfig` document for the
  iSCSI modules (the old single-document patch is kept as
  `modulepatch-legacy.yaml`) and `workload-isolation-patch.yaml`.
- Settings of the `endpoint` section: `cafile` and `insecure` for the
  certificate of the storage, `retrydelay`, the pause before a `GET` answered
  without data is sent again, and `maxreplysize`, the largest reply the plugin
  reads (1 MiB when not set). `tries`, which nothing read before, is the
  number of times such a `GET` is sent.
- Unit tests, run by `make test`, on replies recorded from a JovianDSS 1
  appliance (`pkg/rest/testdata`), and `make vet`.

### Changed

- The plugin uses JovianDSS REST API v4 only, with no setting and no fallback:
  API v3 is abandoned. The appliance still serves v3, but the plugin no longer
  calls it. The oldest supported JovianDSS 1 release is `1.0up33.65410`; an
  older release, or one that serves a different API v4, stops working after
  the upgrade.
- **The certificate of the storage is checked by default.** A JovianDSS that
  presents a self-signed certificate, as it does until an administrator
  imports one, is refused after the upgrade until the `endpoint` section of
  the configuration sets `cafile`, to the certificate or its CA, or `insecure:
  true` (`docs/configuration.md`, "Certificate of the storage"). The plugin
  uses HTTPS only, with TLS 1.2 or newer, and follows no redirect; `prot` is
  no longer needed, and a value other than `https` is refused.
- A volume that has a snapshot is not deleted: `DeleteVolume` answers
  `FailedPrecondition` and names what is in the way, for iSCSI and NFS, where
  it answered `Internal` before. Leftover snapshots that the plugin made
  itself while cloning are removed first.
- `ListVolumes` and `ListSnapshots` read the whole list of the storage on
  every call and resume after the last name returned. A `next_token` of an
  earlier build is refused with `Aborted`, as is any token the plugin did not
  issue for that list; a negative `max_entries` is answered `InvalidArgument`.
  A call returns `max_entries` entries whenever that many follow.
- Volume and snapshot IDs of a CSI request are checked against the names the
  storage takes, and refused with `InvalidArgument` before any request is
  sent. The name the plugin gives a new volume, which is also its CSI ID, has
  128 characters at most; a volume that an earlier build made under a longer
  name is no longer found by its CSI name.
- `iscsi.iqn` is checked at start: lower case letters, digits, `.`, `:` and
  `-`, and 155 characters at most.
- Errors of the storage are recognised by status, class, errno and message
  for each request, from the replies recorded on JovianDSS 1 with API v4: a
  missing pool, a failed login, a name that is too long, a full pool on the
  NAS volume create, and a repeated snapshot or share create among them. An
  error the plugin does not know is reported with everything the storage
  sent, and logged once.
- The iSCSI deletes send `force_umount: false`.
- The configuration guides describe the keys the code reads: `pool` is a key
  of the top level, the log file key is `logfile`, only the first address of
  `endpoint.addrs` is used.
- Node plugin reworked: new stage and publish code on `k8s.io/mount-utils`, and
  a different error code when a volume cannot be attached.
- Container images are built from CentOS Stream 9 instead of Ubuntu.
- CSI sidecars are pulled from `registry.k8s.io/sig-storage` (csi-provisioner
  v2.1.2, csi-attacher v3.1.0, csi-node-driver-registrar v1.2.0), because
  `quay.io/k8scsi` no longer serves them.
- Internal packages are imported by the full module path
  `github.com/open-e/joviandss-kubernetescsi`; the `go` directive is 1.26.
- `google.golang.org/grpc` v1.57.1 to v1.63.2.
- Documentation lives in `docs/`; the duplicate `doc/` folder was removed.

### Removed

- The old driver code, replaced by the shared driver interface with one
  implementation per protocol.

### Fixed

- NFS `DeleteVolume` destroyed every snapshot of the volume: the storage
  deletes a NAS volume together with its snapshots, and the plugin did not
  look first.
- The clean-up of the snapshots that the plugin makes while cloning deleted a
  snapshot without asking whether it had clones; on NFS the storage then
  deletes the clones with it. The clones are read first.
- A repeated NFS `CreateVolume` failed for ever on "Share already exists".
  An existing share is taken when its path is the volume's own.
- `ControllerUnpublishVolume` left the target behind when the volume was no
  longer attached to it.
- A volume with a leftover clone snapshot could never be deleted.
- The lists of one volume's snapshots returned nothing on iSCSI and failed on
  NFS; the lists of all snapshots mixed the snapshots of zvols and of NAS
  volumes.
- The controller crashed on any `starting_token` that was not empty, on a
  request that got no reply (an address that does not resolve, a refused
  connection, a timeout), on a failed reply without an error body, and on a
  bad `endpoint` configuration.

### Security

- Request bodies, which can hold a CHAP password, request headers, which hold
  the REST password, and whole CSI requests are no longer logged at any
  level.
- Names put into a REST path are escaped, and an empty name, `.` and `..` are
  refused, so that an ID from a CSI request cannot address another object.
- `golang.org/x/net` (v0.24.0) and `google.golang.org/grpc` updated against
  the HTTP/2 Rapid Reset attack on the gRPC server (#6).

## [1.1.0] - 2024-04-12

The only version with notes on GitHub, published as "v1.1.0 major code rework"
and marked as a pre-release. The published notes, verbatim:

> - Improvements: Reduce volume creation and deletion time
> - Support: Adds support of TalosOS Linux. JovianDSS supported up to 1.0up30.53
> - Breaking Changes: Volume snapshots and clones have to be deleted before
>   volume itself can be deleted
> - Caution: Was not tested on many linux distributions

From the commits (pull request #5, "Major driver rework", squashed):

### Added

- Talos Linux support.
- From the v1.0.3 branch: the `jdss-csi-cli` command-line tool
  (`app/jdss-csi-cli`, `cmd/controller`, `cmd/node`) to create, delete, list,
  publish and unpublish volumes and snapshots against a running controller;
  controller and node plugin code separated; REST reply types in
  `pkg/rest/types.go`.

### Changed

- Volume naming scheme.
- The node plugin uses the `iscsid` provided by the host machine.
- Services moved to a dedicated Kubernetes namespace.
- Code refactored; documentation extended.

### Removed

- Reverting a volume to a snapshot.

### Fixed

- Target name transfer through the publish context.

### Security

- The dependency update of v1.0.2 (`google.golang.org/grpc` v1.57.1,
  `golang.org/x/net` v0.17.0), as a separate commit on this line.

## [1.0.3] - 2024-03-19

No GitHub release. Tagged on a development branch of nineteen commits
(October 2023 to March 2024, most titled "intermediate"); not an ancestor of
`master`. Its work reached `master` squashed into v1.1.0. It does not contain
the dependency update of v1.0.2.

### Added

- `jdss-csi-cli` command-line tool.
- REST reply types in `pkg/rest/types.go`.

### Changed

- Controller and node plugin code separated.
- Ubuntu container updated.

## [1.0.2] - 2023-10-30

No GitHub release. Everything since v1.0 on the main line:

### Changed

- Go modules: `go.mod` and `go.sum` added (#3); module name fixed and
  versioned.
- Code style fixes.

### Fixed

- Configuration file names.
- Deployment files.

### Security

- Dependencies updated for a security vulnerability: `google.golang.org/grpc`
  v1.57.0 to v1.57.1, `golang.org/x/net` v0.14.0 to v0.17.0,
  `github.com/golang/protobuf` replaced by `google.golang.org/protobuf` v1.31.0.

## [1.0] - 2020-05-31

No GitHub release. Initial import of the plugin ("Repository reinit") with an
unimplemented `ControllerGetVolume`. There is no v1.0.1 tag.

[Unreleased]: https://github.com/open-e/joviandss-kubernetescsi/compare/v1.1.0...master
[1.1.0]: https://github.com/open-e/joviandss-kubernetescsi/compare/v1.0.3...v1.1.0
[1.0.3]: https://github.com/open-e/joviandss-kubernetescsi/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/open-e/joviandss-kubernetescsi/compare/v1.0...v1.0.2
[1.0]: https://github.com/open-e/joviandss-kubernetescsi/tree/v1.0
