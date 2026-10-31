# Merge `nfs-support` into `master` — Design Document (PROPOSED)
[Merge_nfs_support_into_master](#merge-nfs-support-into-master--design-document-proposed)

> **Status: proposed (2026-10-28), revised 2026-10-29 and 2026-10-31; merge
> in progress.**
> The user started the merge on local `master` on 2026-10-29 and resolved
> the conflicts by hand; nothing is committed or pushed. R2 was dropped on
> 2026-10-31: the `jcom` alias is part of the merge commit. This document defines
> what a correct result is and the rules to reach it.
>
> Makes `master` the single line of development again by bringing in the
> unmerged `nfs-support` branch (all plugin work since April 2024), without
> adding new behavior. Comes before the controller fixes
> ([0002](0002-controller-crash-fixes-design.md)), user decision
> 2026-10-28: "this will be longer but it will save time later, because
> diff will be smaller".
>
> Open questions 1-7 resolved 2026-10-28 (user); 8 (registry) in progress;
> 9 new.

## Table of Contents
[Table_of_Contents](#table-of-contents)

- [Overview](#overview) — Overview
- [Problem](#problem) — Problem
- [Background](#background) — Background
- [Key Observation](#key-observation) — Key_Observation
- [Project Values](#project-values) — Project_Values
- [Options Considered](#options-considered) — Options_Considered
- [Design](#design) — Design
- [Verification](#verification) — Verification
- [Consequences](#consequences) — Consequences
- [Risks & Backward Compatibility](#risks--backward-compatibility) — Risks_Backward_Compatibility
- [Relationship to Other Work](#relationship-to-other-work) — Relationship_to_Other_Work
- [Open Questions](#open-questions) — Open_Questions
- [Appendix A. Path-by-path resolution](#appendix-a-path-by-path-resolution) — Appendix_A
- [Appendix B. Procedure](#appendix-b-procedure) — Appendix_B

---

## Overview
[Overview](#overview)

The plugin has had two lines of development since April 2024. `master`
received only maintenance (docs, deploy fixes, an import rename);
`nfs-support` received all feature work (NFS, clones, node rework,
dependency updates) and was never merged. This design merges them into one
`master` by a small set of general rules
([Design](#design)), and defines success as checkable equivalences plus
behavior parity with the published image ([Verification](#verification)).

---

## Problem
[Problem](#problem)

### What is wrong

1. **The main line is not the product.** `master` is where releases, docs
   and fixes are expected, but two and a half years of feature work live
   elsewhere. `master`'s own docs and manifests describe NFS
   (`c466239`) while `master`'s code cannot run it. Anyone who works on the
   plugin first has to pick a line, and there is no rule for which.
2. **Divergence has a cost, and it grows.** Every change on one line must
   be redone or ported on the other. This has already happened: both lines
   fixed the import paths independently (`7dfb5d3` in 2024, `1f1f814` in
   2026), and both changed the same deploy files. The next changes queued,
   the controller fixes in 0002 and the node finding F6, land in code that
   `nfs-support` restructured.
3. **The quality of the unmerged code is unknown.** `nfs-support` was last
   built in July 2024 (Go 1.22) and never verified against the CSI spec on
   a current platform. The published image predates it, so all test
   evidence is for the older code.

### What success looks like

- **S1. One line.** `master` contains every change from both lines since
  the split; each one kept or dropped on purpose, with the reason written
  down.
- **S2. Builds.** Both protocol variants (`iscsi`, `nfs`) build from
  `master` with the current toolchain; every commit on the way builds.
- **S3. No regression on what is deployed today.** On the test cluster the
  merged iSCSI image does at least as well in csi-sanity as the published
  image (same or fewer failures, no new ones).
- **S4. A base for further work.** 0002, F6 and later work target `master`
  only.

### Not solving

- Bugs present on either line (F1–F6): 0002 and later designs. A merge that
  also fixes things cannot be compared with the baseline.
- NFS behavior: built, not tested live ([Open question 6](#open-questions)).
- Cleanup of `nfs-support` code (commented-out code, `lunid.go_tmp`,
  intermediate leftovers), dependency upgrades beyond the branch, the
  Makefile's `PROTOCOL_TYPE` default.
- The other unmerged branch, `helm` (see [Background](#background)).

---

## Background
[Background](#background)

### The lines

| | `master` | `nfs-support` |
|---|---|---|
| Since split `22b7662` (2024-04-15) | 4 commits | 21 commits (2024-04-22 to 2024-07-23) |
| Kind of change | docs, deploy, import rename (`1f1f814`) | features, refactoring, dependencies |
| Size | small | +5504/−1634 lines, 85 files |
| Last activity | 2026-10-21 | 2024-07-23 |

The last release tag, `v1.1.0`, is `25207b1` (2024-04-12), before the
split. The published image `opene/joviandss-csi:latest` matches `master`'s
code layout, so it was built from that era; its exact commit is unknown.

What `nfs-support` brings, reconstructed from the files touched (most
commits are named "Intermediate update"):

1. **Node rework** (Apr–May 2024): new stage/publish code, volume binding,
   a changed error code for failed attach, CentOS images instead of Ubuntu,
   grpc v1.57.1 → v1.63.2, `k8s.io/mount-utils`.
2. **Import paths** (`7dfb5d3`): full module path, like `master`'s later
   `1f1f814`, but without the uniform `jcom` alias.
3. **NAS/NFS REST layer and driver split** (May–Jul): the driver becomes an
   interface with `iscsi_driver.go` and `nfs_driver.go`.
4. **NFS node side** (Jul).
5. **Packaging** (Jul): one binary per protocol, chosen at build time
   (`PROTOCOL_TYPE`, which sets the driver name `iscsi.csi...` or
   `nfs.csi...`); manifests moved to `deploy/joviandss/iscsi/` and
   `deploy/joviandss/nfs/`.
6. **Snapshots and clones** (Jul): NAS snapshots and clones, shared
   volume/snapshot code; 1578 lines of old driver code removed.

Other branches: `node-volume-mounting-rework` is an ancestor of
`nfs-support` (nothing to merge). `helm` (27 commits up to 2024-04-14,
based on `25207b1`) adds a Helm chart and a Talos guide, 14 files, +570
lines; not part of this design, but the rules below apply to it too.

### The merge, measured

Dry run (`git merge-tree --write-tree master nfs-support`, 2026-10-28),
confirmed by the real merge on 2026-10-29: 27 conflicting paths.

| Kind | Paths | Content of the conflicts |
|---|---|---|
| Go code | 19 | import blocks and `common.` vs `jcom.` only; one exception, `pkg/node/host.go`, where `nfs-support` rewrote a function that `master` only renamed |
| Build | `Makefile`, `go.mod`, `go.sum` | `nfs-support` replaced the build; `go` directive 1.26 vs 1.22 |
| Deploy | 2 (NFS manifests, added on both sides) | sidecar image registry |
| Docs | 4 | `doc/` → `docs/` rename on `master` vs new files in `doc/` on `nfs-support`; one text conflict |

Of the Go files that merged without conflict, all 34 came out identical to
`nfs-support` (checked in the in-progress merge, 2026-10-29).

---

## Key Observation
[Key_Observation](#key-observation)

**Changes come in two kinds, and the kind decides how to merge them.**

- A **semantic** change alters behavior or content: a feature, a fix, a new
  fact about the outside world (an image moved to another registry).
- A **mechanical** change is a transformation that can be described by a
  rule and repeated: a rename, an import alias, a file move, formatting.

Merging a semantic change with a mechanical one by text, hunk by hunk, is
the expensive and error-prone way: the result has to be read line by line,
and a clean textual merge can still be wrong. Example from this merge: a
file can take `master`'s `jcom` import and keep `nfs-support`'s `common.`
calls, merging without conflict and failing to build. Cheaper and checkable:
**take the semantic side whole, then re-apply the mechanical change as its
own step.**

In this merge the split is almost total:

- **All of `master`'s Go changes are mechanical** (one commit, `1f1f814`:
  import paths and alias; checked with
  `git log 22b7662..master -- '*.go' go.mod go.sum Makefile`).
- **`nfs-support`'s changes are semantic**, except its own import fix.
- **`master`'s semantic changes are all outside the code**: deploy (sidecar
  registry, Talos patches) and docs. They do not overlap with
  `nfs-support`'s code.

A second consequence: **"no conflict" is not evidence of correctness, so
success has to be shown by equivalence checks** (this tree equals that
tree), not by the absence of conflict markers ([Verification](#verification)).

---

## Project Values
[Project_Values](#project-values)

1. **Nothing lost silently**: every change from either line is kept or
   dropped on purpose (S1).
2. **No new behavior in the merge**: the result behaves like `nfs-support`
   plus `master`'s non-code fixes, so test differences have one cause.
3. **Verifiable**: the result can be checked by commands, not by trust in a
   careful reading.
4. **Every commit builds**, for both protocols, so `git bisect` works.
5. **Readable history**: `git blame` on the NFS code points to the
   original commits.
6. **Speed**: the merge unblocks 0002; no side quests.

---

## Options Considered
[Options_Considered](#options-considered)

| Option | What | Against the values |
|---|---|---|
| **A. Merge `nfs-support` into `master`** (chosen) | one merge commit, rules below | keeps history (5), result verifiable (3); conflict work happens once |
| B. Promote the branch | merge `master` into `nfs-support`, then fast-forward `master` | same tree as A; only the merge commit's parent order differs (`master`'s first-parent history would show the branch line). No gain over A |
| C. Squash merge | one plain commit with the combined result | loses blame and authorship for 5500 lines (5) |
| D. Rebase the branch onto `master` | replay 21 commits | each commit hits the import rename again; intermediate commits likely do not build (4); most work |
| E. Port selected features | cherry-pick or re-implement parts | everything not ported stays stranded (1); the driver split makes partial ports large |
| F. Fix `master` first, merge later | the earlier plan | 0002 hunks written for `driver.go`, redone in `iscsi_driver.go`; rejected by the user |
| G. Keep two lines | NFS stays a separate product | the cost of divergence (Problem 2) stays forever |

---

## Design
[Design](#design)

The merge is done by **classifying every path that changed on either line,
then applying one rule per class**. Appendix A applies the rules path by
path.

### R1. Changed on one side only: keep that side
[R1_Changed_on_one_side_only](#r1-changed-on-one-side-only-keep-that-side)

Git does this already. Examples: `master`'s Talos patches and
`docs/pod-security.md`; `nfs-support`'s new driver files.

R2 (take `nfs-support`'s Go code whole, redo the `jcom` alias as a separate
commit) was dropped on 2026-10-31: the user resolved the Go conflicts by
hand, alias included. The numbers of R3–R6 are kept.

### R3. Semantic vs semantic: combine by hand; facts about the outside world come from the newer side
[R3_Semantic_vs_semantic](#r3-semantic-vs-semantic-combine-by-hand-facts-about-the-outside-world-come-from-the-newer-side)

Where both sides changed the meaning of the same text, keep both changes.
Where they state different facts about something outside the repository,
the newer statement wins, because the world changed in between: the CSI
sidecar images are no longer on quay.io (`master`, 2026); Pod Security is
set by the namespace manifest (`master`, 2026).

### R4. Structure changes on one side apply to the other side's content
[R4_Structure_changes](#r4-structure-changes-on-one-side-apply-to-the-other-sides-content)

A move or rename on one side (`master`: `doc/` → `docs/`; `nfs-support`:
`deploy/joviandss/*.yaml` → `deploy/joviandss/iscsi/`) also applies to
the other side's files and **references**: new files go to the new place,
and links and paths in docs follow. Git handles file moves; references in
text it does not, so they get their own commit after the merge.

### R5. Conflicting build settings: one decision, written down
[R5_Conflicting_build_settings](#r5-conflicting-build-settings-one-decision-written-down)

Settings that cannot be combined are decided once and recorded here:
`go 1.26` without a `toolchain` line (Open question 3); `PROTOCOL_TYPE`
default stays `nfs` (Open question 4).

### R6. The merge resolves; it does not improve
[R6_The_merge_resolves_it_does_not_improve](#r6-the-merge-resolves-it-does-not-improve)

No bug fixes, cleanups or upgrades in these commits, even obvious ones
(value 2). They go into the backlog or 0002.

### Commit shape
[Commit_shape](#commit-shape)

1. **Merge commit** (`--no-ff`, Open question 1): resolution by R1, R3–R5,
   with `pkg/common` imported as `jcom` everywhere.
2. **References commit**: doc links and paths that R4 moved
   (`docs/install*.md` → `deploy/joviandss/iscsi/`, README `doc/` →
   `docs/`, `deploy/examples/nfs` → `deploy/example/nfs`). Some links were
   already broken on `master` since `2d94828`; fixed here because R4 is the
   same work.

Each commit builds both protocols.

---

## Verification
[Verification](#verification)

Success criteria S1–S4 as checks.

| Check | How | Shows |
|---|---|---|
| E1. Code equals the branch up to the alias | after the merge commit: `git diff nfs-support -- '*.go'` shows only import lines, deleted commented-out imports and `common.`/`cli_common.` → `jcom.`; no file imports `pkg/common` under another name | no semantic change lost in the hand resolution (S1) |
| E2. `master`-only content kept | `git diff master -- deploy/talos/ docs/pod-security.md review/` empty | R1 for `master` (S1) |
| E3. Nothing left unresolved | `git grep -n '^<<<<<<<'` empty; `git status` clean apart from untracked work files | S1 |
| B1. Builds | on kub-gate, after each commit: `go mod tidy` (no diff after the first), `make plugin cli PROTOCOL_TYPE=iscsi`, same with `nfs`, `go vet ./...` | S2 |
| T1. Parity | merged `iscsi` image on the test cluster: csi-sanity `identity`, `ctrl-A`, `ctrl-B`, `ctrl-cap`, `node`, compared per test with the baseline (`kub-testing/project-status.md`) | S3 |

T1 pass rule: no test that passed on the published image fails. Known
failures F1–F6 may stay; any change in them (better or worse) is explained
before `master` is pushed. Expected change: CreateSnapshot error handling
(C6) is already fixed on `nfs-support`.

---

## Consequences
[Consequences](#consequences)

**For operators** (anyone deploying the plugin):

- Two images, one per protocol. The iSCSI driver name is unchanged
  (`iscsi.csi.joviandss.open-e.com`); NFS registers as
  `nfs.csi.joviandss.open-e.com`.
- Manifests move to `deploy/joviandss/iscsi/` and `deploy/joviandss/nfs/`.
- Config gains an optional `nfs:` section; the `iscsi:` section is
  unchanged in form (now optional in code). Existing iSCSI configs are
  expected to load; T1 checks it.

**For developers:**

- A build must name the protocol; the default builds NFS.
- The driver is an interface with one implementation per protocol: a fix in
  shared behavior may need to be made in both.
- `deploy/dev/` now targets NFS.

**For the project:**

- `review/2026-10-02-code-review.md` and `review/backlog.md` cite `master`
  before the merge. They stay as a record; new work cites the merged code.
- `kub-testing` docs that name manifest paths (article 05) follow.

---

## Risks & Backward Compatibility
[Risks_Backward_Compatibility](#risks--backward-compatibility)

1. **Old code, new toolchain.** The branch was last built with Go 1.22;
   `go 1.26` and current module resolution may break the build. Guard: B1
   before the merge commit; a build fix would be a separate, recorded
   commit (R6 exception, written down).
2. **Wrong-protocol image.** A default build registers as `nfs.csi...` and
   nothing works on the iSCSI cluster. Guard: build with
   `PROTOCOL_TYPE=iscsi`; the driver name shows in the Identity test.
3. **Untested feature code.** The iSCSI path was reworked (node rework,
   driver split) and not verified since 2024. Guard: T1 against the
   baseline, including the Node baseline of the old code (done 2026-10-28).
4. **Silent semantic mistakes in a clean merge.** Guard: E1 and E2 test
   equivalence instead of relying on conflict markers.
5. **Merge on `master` directly.** The merge is being done on local
   `master`, not on a branch. Harmless while local; guard: do not push
   `master` until B1 and T1 pass ([Open question 9](#open-questions)).

---

## Relationship to Other Work
[Relationship_to_Other_Work](#relationship-to-other-work)

- Next: [0002 controller crash fixes](0002-controller-crash-fixes-design.md),
  on the merged code; then F6 (NodeStageVolume idempotency).
- Backlog: `review/backlog.md` item 35.
- Baselines and results: `kub-testing/project-status.md`.
- Needs the development registry for T1 (Open question 8;
  `kub-testing/docs/joviandss-csi-on-talos/11-dev-registry.md`).
- Later, same rules: the `helm` branch.

---

## Open Questions
[Open_Questions](#open-questions)

1. ~~**Merge commit or squash?**~~ **(Resolved 2026-10-28: merge commit,
   `--no-ff`.)**
2. ~~**Alias rename as a separate commit?**~~ **(Resolved 2026-10-28:
   separate commit; changed 2026-10-31: in the merge commit, R2 dropped.)**
3. ~~**`go` directive.**~~ **(Resolved 2026-10-28: `go 1.26`, no
   `toolchain` line.)**
4. ~~**`PROTOCOL_TYPE` default `nfs`.**~~ **(Resolved 2026-10-28: keep it
   in the merge.)**
5. ~~**Node baseline on the published image?**~~ **(Resolved, done
   2026-10-28:** `kub-testing/tmp/node-20261028-120252`, 14/15, F6.)
6. ~~**NFS live testing.**~~ **(Resolved 2026-10-28: not in this design;
   the `nfs` build must compile.)**
7. ~~**Old branches after the merge.**~~ **(Resolved 2026-10-28: keep
   `nfs-support` and `node-volume-mounting-rework`; user checks them
   later.)**
8. **Build and registry**: the user is setting up a LAN registry
   (article 11). Needed for T1.
9. **Where the merge commit lives until T1 passes.** It is being made on
   local `master`. Keep it there and push only after T1, or move it to a
   branch after committing (`git branch merge-nfs-support`, then reset
   local `master` to `origin/master`)? Recommendation: keep it on local
   `master`, do not push.

---

## Appendix A. Path-by-path resolution
[Appendix_A](#appendix-a-path-by-path-resolution)

| Paths | Class | Rule | Result |
|---|---|---|---|
| 18 conflicted Go files (`app/`, `cmd/`, `pkg/driver/{driver,restwrappers,snapid}.go`, `pkg/node/{host,nodeplugin}.go`, `pkg/pluginserver/pluginserver.go`) | semantic vs mechanical | by hand | `nfs-support` code with the `jcom` alias |
| 34 auto-merged Go files | already equal to `nfs-support` | — | as merged; `jcom` alias where a file imported `pkg/common` under another name |
| `Makefile` | semantic vs mechanical (`-X` path) | by hand | `nfs-support`; its `-X` paths are already full |
| `go.mod` | semantic vs build setting | R5 | `nfs-support`'s requirements; `go 1.26`; no `toolchain` |
| `go.sum` | derived | — | `nfs-support`, then `go mod tidy` |
| `deploy/joviandss/iscsi/*` | move (`nfs-support`) + fact (`master`: sidecar registry) | R4, R3 | as merged: git applied `master`'s image change to the moved files |
| `deploy/joviandss/nfs/*.yaml` | added on both sides, differ in sidecar images | R3 | `nfs-support` content, `master`'s `registry.k8s.io/sig-storage/...` images and registrar `v1.2.0` |
| `deploy/container/`, `deploy/cfg/`, `deploy/dev/` | `nfs-support` only | R1 | `nfs-support` |
| `deploy/talos/` | `master` only | R1 | `master` |
| `docs/compile.md`, `docs/configuration-nfs.md` | move (`master`) + identical content | R4 | either side (identical) |
| `docs/install-nfs.md` | move + text conflict (Pod Security) | R4, R3 | `master`'s text (newer fact) |
| `docs/pod-security.md`, `review/` | `master` only | R1 | `master` |
| README, `docs/install*.md` references | references to moved paths | R4 | references commit |

## Appendix B. Procedure
[Appendix_B](#appendix-b-procedure)

The merge is already started (`git merge nfs-support` on local `master`)
and the conflicts are resolved by hand (2026-10-31). From there:

```bash
# mark resolved, by name (never "git add -A": docs/design/ and
# review/backlog.md are untracked work files)
git add <resolved paths>
# E1-E3 here; B1 on kub-gate; then commit
```

Then the references commit, with B1.
