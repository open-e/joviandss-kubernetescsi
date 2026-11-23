# ADR 0002 against the repository and the spec recordings

Date: 2026-11-13. Record: `docs/adr/0002-api-v4-compliance-jdss1.md`, working
tree after `566fd71` (2,505 lines). Code: `11ffbfd`; `pkg/`, `cmd/`, `app/`
and the `Makefile` are unchanged since. Spec repository:
`../jdss-rest-api-spec` at `851176b`, read only.

Earlier reviews covered the record's wording and internal coherence. This one
checks what the record says against the plugin code and against the
recordings, and asks whether the planned changes are safe and buildable.
Nothing was edited: not the record, not the backlog, not the spec repository.
Both checkers pass (`check_record.py`, `check_names.py`).

Limits: this machine has no Go, so nothing was compiled and `go vet` was not
run. Statements about the recordings come from the spec repository's
generated TypeSpec and from reading `hurl/out/` directly; where I compared
replies myself, the finding says so.

## Summary

| # | Severity | Finding |
|---|---|---|
| 1 | High | The NFS fixes of stage 3 arm a path that destroys a live clone |
| 2 | High | The per-volume snapshot lists return a shape the callers are not built for; "implement the request" and "fix the path" are not enough |
| 3 | High | A full jdss1 v4 run of 2026-11-07 answers what the record calls unknown |
| 4 | Medium | The first commit will not make `go vet` pass: at least 19 more reports |
| 5 | Medium | The dead-code commit does not build: `StorageCfg` is used |
| 6 | Medium | `cafile` cannot verify the recorded jdss1 appliance |
| 7 | Medium | `per_page`: no value stated, and the new stop rule fails on a resumed list |
| 8 | Medium | ID validation names two parsers; two CSI entry points bypass them |
| 9 | Medium | The generator's own input, the jdss1 v4 YAML, does not parse |
| 10 | Low | Seven smaller inaccuracies |

## High

### 1. The NFS fixes of stage 3 arm a path that destroys a live clone

The spec project's Finding 7 (`empirical/hurl/project-status.md`, "Data
safety") records: `DELETE` of a NAS snapshot that has a clone answers 204 and
**the clone is destroyed with it**, confirmed in isolation. A zvol snapshot
with a clone is refused (errno 1000). The record cites Finding 7 only for the
`POST` method bug and never mentions this row.

Today four separate bugs keep the NFS cleanup from deleting anything. The
record fixes all four:

| Bug that blocks the path today | Fixed by |
|---|---|
| `DeleteNASVolume` "delete a dataset ... clones first" is not seen as `Busy`, so `nfs_driver.go:396` is never reached | stage 3, `bug_rest_v4_pools_pool_nas_volumes_dataset_delete_has_clones_message_unrecognised` |
| the lister asks the wrong path and returns nothing | stage 3, `bug_plugin_rest_nas_volume_snapshots_lister_path_lacks_snapshots` |
| `cleanIntermediateSnapshots` fails on its type check | stage 3, `bug_plugin_driver_clean_intermediate_snapshots_type_check_before_list_error` |
| the delete call sends `POST` | dead-code commit, switch to `DeleteNASSnapshot` |

With all four fixed, deleting an NFS volume that still has a clone runs
like this: `DeleteNASVolume` is refused and mapped to `Busy`;
`cleanIntermediateSnapshots` (`nfs_driver.go:291-335`) lists the volume's
snapshots; for each intermediate snapshot it asks `snap.ClonesNames()` and
deletes the snapshot when the answer is empty (`:316-322`). The list entries
carry no `clones` key (finding 2), so the answer is always empty. The
appliance answers 204 and destroys the clone, which is another CSI volume,
possibly mounted by a pod.

The record's own rule makes this silent: Table 2 lists `clones` as an
optional read field whose accessor "returns the documented absent value".
Absent then reads as "no clones".

A second path with the same end: `CreateVolumeFromVolume`
(`nfs_driver.go:249-257`) calls `deleteIntermediateSnapshot` when the
snapshot create says "already exists". Today that message is not recognised
for NAS, so the branch is dead. Stage 3 recognises it
(`bug_rest_v4_pools_pool_volumes_volume_snapshots_post_already_exists_unrecognised`
covers `CreateNASSnapshot`). The branch was written for zvols, where the
appliance refuses and the code then inspects the clones (`:198` onwards). On
NAS the delete succeeds and takes the clone of the earlier attempt with it.
The controller's `VolumeComply` check runs first, so this mostly hits
half-created volumes, but the code no longer does what its comments intend.

What the record needs:

- a statement that no NAS snapshot is deleted before its clones were read
  by name (`GET .../snapshots/<s>` has `clones`, or `GET .../clones`), and
  that a list entry without `clones` means "unknown", not "none";
- a risk for it, next to risk 8, and the row of Finding 7 in the Appendix;
- Testing step 5: the condition "deleting a volume after a clone from it"
  has to say that the clone still exists and is readable afterwards. As
  written, a run that destroys the clone passes.

### 2. The per-volume snapshot lists return a shape the callers are not built for

Recorded on jdss1 v4 (`empirical/jdss1/v4/typespec/pools/+poolname+/volumes/+volumename+/snapshots/GET.tsp`
and the `nas-volumes` twin): both per-volume lists return the same short
entry as the all-snapshots lists:

```
name, parent { name, type, properties { guid } }, properties { creation: int64, guid }
```

No `clones`, no top-level `creation`, no `volsize`. The code assumes the
full by-name object:

- iSCSI `ListVolumeSnapshots` asserts `*[]jrest.ResourceSnapshot`
  (`iscsi_driver.go:385-393`); the NFS lister decodes into
  `[]ResourceNASVolumeSnapshot` (`entrielisters.go:217`);
- both `cleanIntermediateSnapshots` call `snap.ClonesNames()`
  (`iscsi_driver.go:212`, `:282`; `nfs_driver.go:316`, `:421`);
- `completeListResponseFromVolumeSnapshots` reads `s.Creation`
  (`controller/listers.go:124`, `:148`);
- the list token uses `RestSnapshotEntryBasedID` on the by-name type.

The record says "implement the request" for the zvol stub and "fix the path"
for NAS, and counts twelve caller lines. It treats `listers.go:124` and
`:148` as a type change of `Creation`; on these lists the field is not there
at all, it is `properties.creation`. Decoded into the by-name types the
replies give no error and empty values, which is the silent failure
priority 4 is meant to exclude.

On iSCSI the effect is not data loss: the cleanup would try to delete every
intermediate snapshot, the appliance refuses those with clones, and
`DeleteVolume` returns the plain `Busy` error without the list of dependants.
On NFS it is finding 1.

What the record needs: which Go type each per-volume list decodes into (the
short types fit), where the cleanup gets `clones` from, and the callers
added to the stage 2 and 3 file rows. The fixtures of `list/volume-kinds.hurl`
that the record already plans would show the shape at once, if a test
asserted the fields the callers read.

### 3. A full jdss1 v4 run of 2026-11-07 answers what the record calls unknown

`empirical/hurl/out/jdss1/20261107-010938` is a full v4 run on jdss1: 76
files, 14,128 requests, one without an answer (the spec project's run table,
`project-status.md:716`). It is in the generated jdss1 v4 specs. The gap
tests that were lost on 2026-11-05 all ran and passed as files:
`request-empty-body` (772 requests), `request-list-query` (480),
`pool-missing`, `resource-delete-missing`, `resource-list-missing-parent`,
`resource-name-taken`, `volume-sparse-default`, the `*-strings` files.
`nospace`, `scale` and `race` did not run.

The record counts this run (Table 4 says 15 jdss1 runs, and there are 15
only with it) but its text predates it:

| Record says | Where | The run shows |
|---|---|---|
| "There are no jdss1 v4 results" | How the gap runs went | they exist |
| the `{}` body "is recorded on jdss1 v3 and jdss2 v4 only"; question 30 open; the live test "would be its first check" | What we do not know, Stage 1, risk 15, question 30, Testing step 4, the gap entry, Measurements | my comparison of the 362 nobody/empty pairs of that file: same status, class, errno and message in every pair (two differ only in a generated user name) |
| paging without `per_page`: "On jdss1 v4 this was never recorded, so it may behave the same way" | the `per_page` section, the gap entry | on `volumes`, `nas-volumes`, a volume's snapshots and both all-snapshots lists, page 0 returns everything, page 1 and page 99 are empty; `results` is the total; `_dc` changes nothing. The plugin's stop rule ends |
| snapshot list of a missing parent: "500 on jdss1 v3, 404 on jdss2 v4" | gap entry | jdss1 v4: 404 `ItemNotFoundError`, `Resource <p>/<v> not found.` |

Consequences for the argument, not only for the wording:

- Question 30 can be decided from evidence now. Risk 15 shrinks to what
  `nospace`, `scale` and `race` would have shown.
- The `per_page` commit is justified as "without it, the v4 move could make
  CSI ListVolumes run forever" and the gap entry says "must be fixed before
  the v4 move". On JovianDSS 1, the only product in scope, the recordings
  now show the loop ends on v4 as it does on v3. The commit is still a
  sound hardening and a preparation for JovianDSS 2, but it is not a
  precondition of stage 1, and the record should say which it is.
- The default page size is still unknown: the test lists hold three to
  eight objects.

The Evidence gaps, the gap-run section, Stage 1, risk 15, question 30 and
Testing step 4 need one pass against this run.

## Medium

### 4. The first commit will not make `go vet` pass

The record says the current code has six non-constant format strings and
that "nothing else would change in this commit". By reading, there are at
least 19 more that the same vet check reports:

- `pkg/rest/targets.go:45` and `:76`: `l.Debugf(fmt.Sprintf(...))`. The code
  review the record draws on lists both (R4: "`targets.go:45,76`"); the
  record lists `:150` and `:184` instead of, not beside, them.
- `status.Errorf(codes.X, <variable>)` with no further argument:
  `pkg/controller/controllerplugin.go:127`, `:134`, `:377`, `:381`, `:386`,
  `:410`, `:565`, `:659`, `:663`, `:676`, `:732`, `:884`, `:893`, `:897`;
  `pkg/node/targetapi.go:516`, `:524`; `pkg/node/iscsi_stage.go:117`. vet
  treats `status.Errorf` as a printf wrapper, and this is the usual report
  after a move to Go 1.24 or later.

Most of the controller sites pass `err.Error()`, which is appliance text, so
they are the same bug the item is named after, one layer up: a `%` in an
appliance message garbles the gRPC status. Multi-line calls were not
searched, so the list is a floor.

The record does say that no vet result exists and that "anything else vet
reports is decided then". But the count, the file row of Table 3 (six
`pkg/rest` lines), the item's `rest` area and Testing step 1 all assume six.
The commit touches the controller and the node, and vet is the gate for
every later commit, so this is better known before the series starts.

### 5. The dead-code commit does not build

`StorageCfg` is on the list of dead types. It is used:
`pkg/controller/config.go:14` (`RestEndpoint rest.StorageCfg` in
`ControllerCfg`, itself the type of a field at `controllerplugin.go:93`).
The commit is limited to `pkg/rest` and `nfs_driver.go:322`, and the record
says dead code outside `pkg/rest` is not part of it. Removing the type
breaks `pkg/controller`, against "each commit would build on its own".
Either `StorageCfg` stays, or `config.go` joins the commit. The code review
of 2026-11-07 has the same wrong entry. The other names on the list have no
remaining users (checked across `pkg`, `cmd` and `app`).

### 6. `cafile` cannot verify the recorded jdss1 appliance

The record says a `cafile` lets a self-signed appliance be verified, and
risk 11 offers "`cafile` or `insecure`". The certificates in the recordings
(`report.json` of the jdss1 runs) say otherwise:

| Property | Recorded | Effect on Go's verification |
|---|---|---|
| validity | 2025-07-21 to 2026-07-21 | expired before every run; fails whatever the trust store |
| subject alternative names | three `172.x` management addresses | the runs connect to `192.168.28.103`; the address check fails |
| two nodes | two different self-signed certificates behind the same address | one `cafile` entry is not enough; the file must hold both |

So on this appliance only `insecure: true` works, unless the administrator
installs a certificate of their own. jdss2's certificate is valid and lists
its addresses, so this is not universal. The lab appliance was not checked.

As planned, "verification on by default" would in practice move every
default installation to `insecure: true`, which buys an upgrade break
(risk 11) and no security. Two ways out, for the user to weigh:

- keep the design and say plainly that verification needs a proper
  certificate on the appliance; the release note should not promise that
  `cafile` is enough;
- pin the appliance certificate by fingerprint instead of, or beside,
  `cafile`. That works with self-signed, expired and address-less
  certificates, and takes one fingerprint per cluster node.

Either way the pre-series TLS check of risk 10 should also read the lab
appliance's certificate: dates, names, and whether both nodes show the same
one.

### 7. `per_page`: no value, and a stop rule that fails on a resumed list

- The record never says which `per_page` the plugin sends. The Proxmox
  plugin sends 100. Without `per_page` jdss1 returned everything in one
  page, so the value decides how many requests a list costs.
- The new stop rule is "stop once it has received as many entries as the
  reply's `results` count announces". `results` is the total (confirmed on
  jdss1 v4). A CSI list that resumes from a token starts at a later page
  (`driver/common.go:31-86`, `token.Page()`), so the entries received in
  this call never reach the total, and the rule never fires: exactly the
  case of a server that repeats a page. A rule that holds: stop when a page
  is shorter than `per_page`, or when `(page + 1) * per_page >= results`.
- `getResourcesList` only sees `[]RestResource`; `results` is dropped in the
  six `grf` closures in `iscsi_driver.go` and `nfs_driver.go`. Table 3 lists
  `entrielisters.go` and `driver/common.go` for this commit, not the
  drivers.
- `restwrappers.go:40`, one of the twelve lines, builds the list token from
  `entry.Volume` twice, so every snapshot of one volume gets the same token
  (tracker K2). A like-for-like edit to `Parent.Name` carries that over into
  lists that will now return real entries (risk 4).

### 8. ID validation names two parsers; two CSI entry points bypass them

`impl_plugin_driver_validate_names_and_ids` checks IDs "where it is parsed
(`NewVolumeDescFromCSIID`, `NewSnapshotDescFromCSIID`)". In the code
`NewVolumeDescFromCSIID` is a one-line wrapper around `NewVolumeDescFromVDS`
(`lunid.go:147-149`), and two CSI IDs go to `NewVolumeDescFromVDS` directly:
the volume ID of `DeleteVolume` (`controllerplugin.go:550`) and the source
volume ID of `CreateVolume` (`:345`). A check in the wrapper misses both. A
check in `NewVolumeDescFromVDS` also runs on names the appliance returns in
lists (`controller/listers.go:54`, `:84`, `:186`, `:232`), where a foreign
name must be skipped, not treated as a bad request. The record should say
which: the simplest is to make those two callers use the CSI parser.

Also for risk 13: with the 131 limit applied to IDs, the ID of an existing
132-character volume is refused with `InvalidArgument` on delete and
publish, not only "looked up under its `vs_` name and not found".

### 9. The generator's own input does not parse

The record reports the tab defect in the jdss2 v4 file (line 15389) and
plans to read `b_jdss1_v4/out/openapi/openapi.yaml`. That file has the same
defect: five `pattern:` values with a literal tab, the first at line 14954,
and Python's YAML parser refuses it there (tried with `yaml.safe_load`). So
question 31 is not a preference; the YAML route is closed until the spec
project fixes the emit or switches to JSON. To report to the spec project
with backlog item 52, which names only the jdss2 line.

## Low

10. Smaller inaccuracies, each one clause in the record:

    - `RestErrorArgumentIncorrect` is called "a code that exists but is
      unused today". It is used nine times in `pkg/driver` (`iscsi_driver.go:572`,
      `targetid.go:31`, `snaptoken.go`). It is unused by the mapping in
      `errors.go`.
    - Table 2 names pool `size` as a required read field. Nothing reads it;
      the only pool field read is `Available` (`controllerplugin.go:1017`).
      It also names "`origin` and `clones` on a volume": `ClonesNames()` is
      only called on snapshot replies. Measurements lists volume
      `available`, `used`, `reservation` and `refreservation` under "read
      fields"; none is read.
    - `bug_plugin_rest_get_error_maps_body_only_ignores_http_status` says
      "every `Get*` decodes before looking at the status".
      `GetVolumeSnapshot` checks the status first (`storageapi.go:211`).
    - `crash_plugin_rest_setup_endpoint_ignores_error`: an empty `addrs`
      with a valid `idletimeout` produces no setup error at all;
      `SetupRestProxy` returns nil and `Send` panics at `rproxy.go:91`. The
      fix needs an explicit check for it, which the item implies but does
      not state.
    - `impl_plugin_rest_reply_size_limit`: `io.LimitReader` alone cuts the
      reply short without an error, and the caller then sees broken JSON.
      To "fail with an error naming the limit" the reader has to allow one
      byte more than the limit and test for it.
    - The 19 `Resource*` names stay in the generated file, but `types.go`
      also holds `GeneralResponse`, `ResultEntries`, `Disk`, `ShareNFS`,
      `ShareSMB` and the two regular expressions; the record does not say
      where they go.
    - Seen while reading, not in the record or the reviews:
      `iscsi_driver.go:610` deletes the target by its bare hash (`tname`)
      where the target was created under `iqn`, so that cleanup never
      matches. For the tracker.

## Verified and found correct

- Counts: 34 `api/v3` strings plus 3 in comments; 19 `Resource*` types, 10
  accessors, 8 custom decoders; the file sizes of the Measurements table;
  104 and 108 TypeSpec files; 24,163 and 24,788 lines; 15 and 20 runs.
- Line references in `rproxy.go`, `rest.go`, `errors.go`, `types.go`,
  `storageapi.go`, `entrielisters.go`, `targets.go`, `lunid.go`,
  `iscsi_driver.go`, `nfs_driver.go` and the controller, as far as cited
  above; the five crashes and their triggers; the twelve caller lines as a
  count; the `Makefile`'s `go mod tidy`; the configuration struct and the
  guide's wrong keys.
- Recorded messages on jdss1 v4 for the error findings: snapshot "already
  exists" (errno 5), the three 404 replies of LUN delete, `Clone <n> not
  exists`, the `dataset` clones message, share 409, `parent.type` 1 and 4.
- The name-length arithmetic (243 today, 131 after), the `vb_` alphabet
  against the allowed characters, the default `iscsi.iqn` against the target
  rule.

## Still open from earlier reviews

Backlog items 55 and 56 (the "assumptions" sentence, the wordings of the
JovianDSS 2 exclusion, what the Decision names) are untouched by this
review.
