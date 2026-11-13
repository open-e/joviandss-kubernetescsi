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

### Changed

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

### Security

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
