# Design documents

This folder holds the design documents of the JovianDSS Kubernetes CSI plugin.
A design document says what a change would do and why before the code is
written, and keeps the decisions taken on the way. One file per design:

- [0001](0001-nfs-support-merge-design.md) — merge of the `nfs-support` branch into `master`
- 0002 — REST API compliance (API v4): merged on 2026-11-10 into [`../adr/0002-api-v4-compliance-jdss1.md`](../adr/0002-api-v4-compliance-jdss1.md); its old path is a pointer
- [0003](0003-controller-crash-fixes-design.md) — controller crash fixes

The status callout at the top of each document says where it stands and which
documents come before it. Decisions a design leads to are recorded one per
file in [`../adr/`](../adr/README.md); a design links the record instead of
repeating the argument. The documents themselves carry no reading
instructions; this file holds them. It describes the conventions as 0002
follows them; 0001 and 0003 are older and differ in places.

## Format

The format is the one of the JovianDSS Proxmox plugin, described in its
[design README](https://github.com/open-e/JovianDSS-Proxmox/blob/master/docs/design/README.md).
In short:

- **File name** `NNNN-<topic>-design.md`, where `NNNN` is a sequence number.
- **Title and status.** The title ends with the status in parentheses:
  `PROPOSED`, `ACCEPTED` or `IMPLEMENTED`. A blockquote under the title, the
  status callout, holds the detail: what is decided, what is deferred, the
  dates, and the commit or branch the line numbers refer to. The body has no
  dates and no history; it describes the current code and what the design
  would change.
- **Tense.** A `PROPOSED` document says what "would" change. Facts about the
  current code and the evidence are in the present tense. When the design is
  implemented, the tense is converted rather than the text rewritten.
- **Sections.** Problem, background and evidence, key observations, project
  values, options considered, the recommended solution, open questions,
  relationship to other designs, risks and backward compatibility, the files
  that would change (a table), testing. A long document adds a table of
  contents, a list of tables and an appendix.
- **Code references** are `file:line`, as relative links where a link helps.
  The status callout says which commit the lines refer to; 0002 calls it "the
  current code" in the body.
- **Plain language.** Jargon is explained in a Terms table at the end of the
  document (0002: [Table 10](../adr/0002-api-v4-compliance-jdss1.md#terms)). Details of the test environment, such as machine
  names, pool names and addresses, do not belong in a design; the testing
  repository holds them.

## Navigation

- **Table of contents and anchors.** Every heading is followed by a line that
  links to itself, with the heading's words joined by underscores:
  `[Open_Questions](#open-questions)`. The table of contents uses the same
  tokens. In vim, `*` on a token jumps between the table of contents and the
  section; in a browser the links work as usual.
- **List of tables.** Tables are numbered through the document in the order
  they appear, and the body refers to them as "Table n". Each table has a
  unique `tbl_*` tag after its caption, for example
  `` **Table 9 — Where the names and IDs come from** · `tbl_id_sources` ``, and the list of
  tables repeats the tag. In vim, `*` on a tag jumps between the list and the
  table.
- **Open questions** are numbered. A decided one is kept and struck through,
  with the decision appended:
  `15. ~~**Tense?**~~ — **resolved**: the conditional, …`. The document is
  thereby its own decision record.
- **Risks** are numbered; the body says "risk n".

## Findings and IDs

A problem found is called a finding. In the merged 0002 it carries a name of the form `kind_area_slug`, mapped from the older letter-and-number IDs in `kub-testing/issues/`; the older documents still use those IDs. The reviews, the issue
tracker of the testing repository and the designs share the IDs, so F8 is the
same finding everywhere. The letter says where a finding comes from:

| IDs | What they are | Defined in |
|---|---|---|
| F1, F2, … | findings from tests and recorded replies: csi-sanity runs, recorded REST replies | the issue tracker, `kub-testing/ISSUES.md` |
| C, D, K, N, P, R, S, X + number | findings of the code reviews, one letter per area: C crashes, D data loss or wrong results, S security, N node, K controller, R REST client, P plugin server, X command-line tool | the reviews in [`review/`](../../review/) |
| G1, G2, … | gaps in the evidence | the design that found them (0002: [Evidence Gaps](../adr/0002-api-v4-compliance-jdss1.md#evidence-gaps)) |
| SC1, SC2, … | compliance rules | 0002, [What "compliant" means](../adr/0002-api-v4-compliance-jdss1.md#what-compliant-means) |
| A1, A2, … | the items a design would deliver | the design's Recommended Solution |
| hurl Finding n | numbered findings of the hurl test suite in `jdss-rest-api-spec` | 0002, [Table 9](../adr/0002-api-v4-compliance-jdss1.md#id-sources) |

A letter can mean something else inside one document: 0001's R1 to R6 are its
merge rules, and 0003's D1 to D7 are its design items. 0002 has a table that
says where each kind of ID comes from
([Table 9](../adr/0002-api-v4-compliance-jdss1.md#id-sources)); the two older
documents say what their letters mean where they first use them.

How a design uses the IDs:

- Each finding the design acts on is described once, in the design's Appendix.
  The body uses only the ID, as a link to that description.
- A finding outside the design's scope is named by ID with a few words where
  it comes up, and nothing more.
- The design does not track the status of a finding; `kub-testing/ISSUES.md`
  does.
- The items a design would deliver (A1, A2, …) are grouped into stages, which
  are groups of commits in order. The Recommended Solution defines both.
