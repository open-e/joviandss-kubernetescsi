# Architecture decision records

This folder holds the decision records of the JovianDSS Kubernetes CSI
plugin. A record treats one problem: what is wrong, the facts it rests on,
what we care about, the options weighed, the decision and its reasons, what
changes, what follows, and how it is checked. Where a problem needs a full
design, the record is that design: 0002 holds the whole migration to API v4,
from the problem to the last implementation item. Older designs stay in
[`../design/`](../design/README.md) until they are reworked into this form.

Writing the record is half its value: it forces the reasoning into the open
where it can be disagreed with. The other half is history: a reader years
later can see why the system is the way it is and how long each decision
governed the work.

## Records

| Record | Problem | Decision | Status |
|---|---|---|---|
| [0002](0002-api-v4-support-jdss1.md) | API v4 support for JovianDSS 1: v3 is deprecated, the plugin talks only v3, its reply types do not match what the appliance sends, and the client misreads replies, crashes on bad input and leaks secrets | v4 only, v3 deprecated, minimum release stated; reply types generated from the recorded jdss1 v4 API under the same names, only read fields can fail; the error, validation, TLS and logging decisions of its stages 3 and 4; delivered as one series | Proposed |

The sequence starts at 0002 on purpose; 0001 is not used in this repository.

## One problem, one record, one at a time

- **One problem per record.** The problem is the unit, not the decision: a
  problem may need several decisions, and they belong together. 0002 answers
  "support API v4 on JovianDSS 1" with the move to v4, the reply types
  and the correctness of the client against v4, because none of them solves
  the problem alone. The first two are architectural choices and are weighed
  in its Options considered, one group each; the third is fixes the evidence
  dictates, decided as questions. For the move itself only one option was
  considered, v4: v3 is deprecated and will not be supported. So its Options
  considered says that, and why, instead of weighing alternatives that were
  never real.
- **Records are not piled up.** A record is implemented after it is
  accepted, and the next record about the next problem is written after
  that. Several proposed records waiting on each other are a sign that the
  problem was cut in the wrong place.
- **What does not get a record:** editorial choices, such as the tense of a
  document or where release notes live. They are one line in the record's
  Decisions table.
- **What is outside a record is said once**, in its Problem statement under
  "not solved", with where it is handled. 0002 does not touch JovianDSS 2;
  a later record will.

## File name and status

`NNNN-<problem-slug>.md`: a sequence number for order, then the problem the
record answers, for example `0002-api-v4-support-jdss1.md`. The name says the
problem, not the decision: a proposed record may change its decision without
changing its name, and a later record that supersedes it about the same
problem is found next to it. The decision is the bold paragraph in the
record and a column in the table above. Numbers are never reused. The title
line of the file is `# NNNN — <problem>`.

| Status | Meaning |
|---|---|
| Proposed | under discussion; the text may still change |
| Accepted | signed off by the approver; from then on the text is frozen |
| Superseded | replaced by a later record, which the status line links; the text stays as it was |
| Retired | the decision no longer applies and nothing replaces it; the text stays |

The date a decision was taken and the date the record was accepted are
different things and both go in Metadata.

## Sections, in order

Every `##` and `###` heading is followed by its self-anchor line (see
[Navigation](#navigation)). A section with nothing to say is kept with one
sentence saying so, so that a reader knows it was considered.

| Section | What it holds |
|---|---|
| Table of Contents | the `##` and `###` headings, as underscore tokens; named headings (see below) are not listed |
| List of Tables | the numbered tables with their `tbl_*` tags, in order of appearance |
| Metadata | `Date:` when the decisions of the Options section were taken and when the text was written, the question decisions being dated in Advice; `Driver:`; `Approver:`; `Code:` the commit the line numbers refer to, called "the current code"; `Implementation:` where the work is described and how it is delivered; `Accepted:` the date, once it is; `Formerly:` the design the record was merged from and when, only for a record that absorbed one |
| Status | one word from the table above, with a link when Superseded |
| Problem statement | what is wrong, one bold sentence per part and a few plain ones; what has to change; what success looks like, stated once; what this record does not solve and where that is handled |
| Context | the facts the decision rests on, in subsections by subject: what exists today, what the evidence shows, what it does not show, each a claim with its data or a link to the Appendix. Ends with **Key observations**: the few conclusions the decision turns on, numbered |
| Priorities | the project values that weigh on this record, in the order they weigh here. Values are the project's and stable; priorities are this record's order among them, and the options argue against them by number |
| Options considered | the architectural and strategic decisions of the record: the ways the problem in the Problem statement could be solved, and which one is taken. The section answers the Problem statement and is read against it, each group of options (A, B, …) answering one part of the problem; a part whose fix follows from the evidence has no group, and its decisions are questions in the Decisions table. Decisions below that level, on how a chosen way is carried out, are questions too, never options. For each choice, what was evaluated: two or three real contenders, their trade-offs stated honestly against the priorities, one subsection per option, the adopted one first and marked. Where only one option was considered, the subsection says so and explains why, with the price that option carries; alternatives nobody weighed are not written up to fill the list. An option carries what distinguishes it, not the mechanism, which belongs to the implementation item. Ends with **Decision**: bold, then the rationale naming each priority, then, for a choice with contenders, the sentence that tells a reader with another order of priorities which option they would pick |
| What changes | the work: the commits in order, grouped in stages, each implementation item a named heading with what changes, why, how, and what must hold once it is done. The mechanism lives here and nowhere else. The unit tests that show an item works are the implementing agent's, not the record's ([The Testing section](#the-testing-section)) |
| Consequences | what becomes true once the record is applied, positive and negative; then the confidence level and the condition that should trigger a re-evaluation |
| Risks | backward compatibility first, preserved and not preserved; then numbered risks, each with the concrete failure and its guard. Numbers are never changed once given; new risks are appended |
| Affected systems and files | a table by system first, then the files one row each with the stage that changes them; a last row for what is not affected |
| Testing | the tests that check the record, from the cheapest to the live one, as numbered steps; each a general description that an agent other than the author can construct and run; no unit tests ([The Testing section](#the-testing-section)) |
| Open questions | the open ones, numbered, each with how it gets answered; then the subsection **Decisions**, a table of the decided questions with their numbers kept, one line of decision each and a link to the section that argues it |
| Relationship to other designs | the designs, records and repositories this record touches, one line each with the link; then **Future work** subsections for what is deliberately left to a later record |
| Advice | the inputs the decision was formed from, each with its date: findings, reviews, tracker entries, user decisions |
| Appendix | what is looked up rather than read: the evidence sources, the evidence gaps, the compliance rules and matrix, the finding catalogues by kind, the measurements, where the names and IDs come from, and the terms |

## Names, not IDs

Findings, gaps, compliance rules and implementation items are named, not numbered:
`<kind>_<area>_<slug>`, for example `gap_jdss_rest_empty_body_on_get_delete`,
`bug_plugin_rest_pool_iostats_type`, `rule_rest_reply_types_decode`,
`impl_plugin_driver_validate_names_and_ids`.

- The kind comes first and says whether the name is a finding or a
  prescription:

  | Kind | Names | Where its heading lives |
  |---|---|---|
  | `bug` | behaviour of the code that is wrong | the Appendix catalogue of its subject |
  | `crash` | a panic the code can reach | Appendix, Crashes |
  | `dataloss` | a wrong result that loses or corrupts data | Appendix |
  | `security` | a finding of the security review | Appendix, Security findings |
  | `gap` | evidence that is missing: a case no recording shows | Appendix, Evidence gaps |
  | `rule` | a compliance rule the code is checked against | Appendix, What "compliant" means |
  | `impl` | an implementation item: a change the record delivers, which implements the decision | What changes, under its stage |

  The first five are findings: they describe what is, and their field table
  says where, what happens, the evidence, and which implementation item or
  stage fixes them. A rule says what must hold. An implementation item says
  what is done, whatever its form: a repair, a new file, a tool, a setting;
  it names the findings and rules it closes, so the two directions link each
  other. The kind was `work` until 2026-11-11 and became `impl` before any
  document outside the record used the names; that is the one exception to
  the rule below that a name never changes.
- Then the
  area: `jdss_rest` (what the appliance's REST API does), `plugin_rest`
  (`pkg/rest`), `rest` (the API as the contract between the two: the
  compliance rules), `plugin_driver`, `plugin_ctrl`, `plugin_node`,
  `plugin_cli`, `plugin` (several packages), `lab`, `spec`, `docs`, `build`.
  Then two to five words, object first. Lower-case letters, digits and
  underscores only.
- Each named thing is a `####` heading, in the Appendix for findings, gaps
  and rules, in What changes for implementation items, with a small field table under
  it holding its data. The heading is its own anchor and a single word, so
  `*` in vim and a link on GitHub both reach it; such headings need no
  self-anchor line and are not listed in the Table of Contents.
- A name is used as a link, `[name](#name)`, wherever it is mentioned. A name
  that has no heading in the record, because it is described only in a review
  or the tracker, is written in code font without a link.
- Names have no order, so a range such as "`bug_a` to `bug_b`" tells a reader
  nothing. List the names, or name the group and link the section that holds
  it: "the stage 4 items", "the first eight gaps under Evidence gaps".
- The older letter-and-number IDs (F8, C3, G5, A11) live on in the reviews
  and in kub-testing's `ISSUES.md`; the map between them and the names is
  `kub-testing/issues/`, one file per old prefix, which is also where a new
  name is registered. A name is chosen once and never changed after it is
  referenced; a changed understanding gets a new name and the old one is
  marked superseded.
- Numbers stay where they mean order or count: risks, questions, testing
  steps, stages, tables, key observations, priorities. They are never
  renumbered once referenced, so a reference such as "risk 9" or "question
  27" keeps its meaning; new entries are appended. Other documents'
  numbering, the hurl suite's Findings, design 0003's items, kub-testing's
  rows, is cited as theirs. A list that has gaps, because entries moved out
  of it, as the open questions do when they are decided, is written as a
  bullet list with the number inside the bold, `- **30. <question>?**`: a
  renderer takes an ordered list's numbers from its first item and renumbers
  the rest, so the numbers a reader saw would not be the ones written.

## How the text is organised

- **Claim, explanation, data.** Every block, in every section, in the same
  order: the claim, one sentence in bold or as the heading; the explanation,
  a few plain sentences; the data and the references last, as a short table
  or a labelled line. A reader who trusts the claim stops after the first
  sentence; a reader who doubts it finds the evidence in the same place every
  time. Details are never dropped to make a record shorter; they are placed.
- **Each fact once, at one place.** Options carry the alternatives and their
  price; What changes carries the mechanism; the Appendix carries the
  catalogue; Context carries the claims and links to the catalogue; the
  Problem statement states success once; the Decisions table indexes the
  decisions and does not restate their arguments. When two sections start to
  say the same thing, one of them links to the other.
- **Plain language.** Jargon is explained once in Terms. No history and no
  dates in the body: dates belong to Metadata and Advice, and the body
  describes the current code and what the record changes. While Proposed, the
  record says what "would" change; facts about the current code and the
  evidence are in the present tense; after implementation the tense is
  converted, not the text rewritten.
- **The record calls itself "this document."** Not "this design" or "this
  record", so that the text reads the same whether the reader came for the
  design or for the decision.
- **References are links.** `file:line` of the commit named in Metadata, as
  relative links into the code where a link helps; sections by their anchor;
  the spec repository's findings by their number, since that repository is
  read only for this work. Test-environment details, such as machine names,
  pool names and addresses, do not belong in a record; the testing
  repository holds them. Secrets never.

## The Testing section

A record is written before the code, so its Testing section is where the
implementation meets what the designer could not see from the design. The
section has two jobs: to pin the conditions the implementation must meet
that only the evidence or the lab know, such as a reply shape, a tool that
refuses an image, or a log that leaks secrets at Debug; and to find the edges
the design did not foresee, by running the result against the real thing.
Unit tests do neither. The agent implementing an item writes them, as part
of the item, and the record does not list them: an item may say what its
tests must prove ("a reply over the limit fails with an error naming the
limit"), never how many tests there are or how they are built.

Each step is one test, described so that an agent other than the author can
construct it and run it without asking:

- **what it proves**: the edge or the condition, with a link to the finding,
  risk or question it answers;
- **where it runs and who runs it**: the machine, the appliance or the build,
  and whether the user, the implementing agent or a tool drives it. A test the
  agent cannot run itself, because the lab is out of its reach, says so;
- **what has to be in place first**: builds, settings, fixes committed, tools
  that exist, data to compare with. A prerequisite that does not exist yet is
  named, with where it is tracked and who provides it;
- **what is done**: the command or the sequence, concrete enough to repeat;
- **what counts as pass**, as something observed: a log line, a status, a
  diff, a comparison with a named baseline; and **what is recorded** when the
  result is neither pass nor fail, so that an unknown outcome becomes a
  result instead of being passed over.

How much of this a step spells out follows the nature of the test. A simple
test is a sentence: delete a volume, then send the same delete again, and
check that the REST request that now fails is handled as "already gone" and
not as an error. A scenario needs its sequence, its timing and what to watch:
shut the appliance off by force while create and delete requests are in
flight, and record how the cluster behaves under that load and what each
request is answered while the pool moves to the other node. The parts above
are what a reader needs, in as many words as the test takes, not a form to
fill in.

The steps go from the cheapest to the live one, and the section ends with
what the tests cannot reach, so that a reader knows which findings have no
cover in this delivery. A test that depends on an implementation detail, such
as a log line keeping its query string or an error text staying searchable,
says so, and that detail becomes part of the item it depends on. The results,
once a step has run, go to the testing repository with the rest of the lab's
data, not into the record.

## Navigation

Every `##` and `###` heading is followed by a line that links to itself, with
the heading's words joined by underscores, `[Open_questions](#open-questions)`,
and the Table of Contents uses the same tokens, so `*` in vim jumps between
an entry and its section and the links work on GitHub. Tables that carry a
number have a caption `**Table n — Title** · `tbl_tag`` and appear in the
List of Tables; numbering follows order of appearance, and a catalogue that
became headings is no longer a table. A heading that something links to is
not renamed; when a section has to move or change its name, every link to it
is updated in the same change.

## Maintaining a record

- **While Proposed**, the text changes freely, and every change keeps the
  checks below green. A new finding gets a name, a heading in the Appendix, a
  row in the mapping file, and links from where it matters. A new question
  gets the next number; when it is decided, it moves to the Decisions table
  with its number kept. A new risk is appended with the next number.
- **Once Accepted**, the text is frozen. A change of decision is a new record
  that supersedes this one, and the status line here links it.
- **Checks before a review**, by script or by hand: every `](#anchor)`
  resolves to a heading; every entry in the Table of Contents and the List of
  Tables matches a heading or a caption; risks, steps and tables are
  numbered without gaps or duplicates, and questions without duplicates,
  their numbers kept when they move to the Decisions table and the open ones
  written as bullets with the number in bold; every Testing step says where it
  runs and who runs it; every name is unique and every
  linked name has a heading; no letter-and-number ID remains outside the ID
  sources table and the citations of other documents' numbering; every
  relative link to a file exists; and no "design 0002", "this design" or
  "this record" phrasing has crept back in.

## Skeleton

```markdown
# NNNN — <problem>

## Table of Contents
[Table_of_Contents](#table-of-contents)

- [List_of_Tables](#list-of-tables)
- [Metadata](#metadata)
- [Status](#status)
- [Problem_statement](#problem-statement)
- [Context](#context)
  - [<subject>](#<subject>)
  - [Key_observations](#key-observations)
- [Priorities](#priorities)
- [Options_considered](#options-considered)
  - [Option_A1_<slug>_adopted](#option-a1-<slug>-adopted)
  - [Option_A2_<slug>](#option-a2-<slug>)
  - [Option_B_<slug>_the_only_option_considered](#option-b-<slug>-the-only-option-considered)
  - [Decision](#decision)
- [What_changes](#what-changes)
  - [Stage_1_<slug>](#stage-1-<slug>)
- [Consequences](#consequences)
- [Risks](#risks)
- [Affected_systems_and_files](#affected-systems-and-files)
- [Testing](#testing)
- [Open_questions](#open-questions)
  - [Decisions](#decisions)
- [Relationship_to_other_designs](#relationship-to-other-designs)
- [Advice](#advice)
- [Appendix](#appendix)
  - [Evidence](#evidence)
  - [<catalogue by kind>](#<catalogue>)
  - [ID_sources](#id-sources)
  - [Terms](#terms)

## List of Tables
[List_of_Tables](#list-of-tables)

1. `tbl_<tag>` — **<Title>**

## Metadata
[Metadata](#metadata)

Date: YYYY-MM-DD (<decision>); written down YYYY-MM-DD
Driver: <name>
Approver: <name>
Code: commit `<hash>`, called "the current code"; line numbers refer to it
Implementation: the [What changes](#what-changes) section, <how delivered>

## Status
[Status](#status)

Proposed

## Problem statement
[Problem_statement](#problem-statement)

**<what is wrong, one sentence>.** <a few plain sentences, with links>

What has to change: ...
What success looks like: ...
Not solved by this document: ..., handled in ...

## Context
[Context](#context)

### <subject>
[<subject>](#<subject>)

**<the fact the decision rests on>.** <explanation> <table or links>

### Key observations
[Key_observations](#key-observations)

1. **<conclusion>.** <why it follows>

## Priorities
[Priorities](#priorities)

The project values that weigh on this document, in the order they weigh here:

1. **<value>.**

## Options considered
[Options_considered](#options-considered)

<the parts of the Problem statement decided here, one group each; what was
decided as questions instead>

### Option A1: <name> (adopted)
[Option_A1_<slug>_adopted](#option-a1-<slug>-adopted)

<what it is, in a few sentences>

Advantages: ... (priorities n, m)
Disadvantages: ...

### Option B: <name>, the only option considered
[Option_B_<slug>_the_only_option_considered](#option-b-<slug>-the-only-option-considered)

Only one option was considered. <why; the price it carries>

### Decision
[Decision](#decision)

**<the decision, a paragraph>.**

Rationale: ... (priority 1) ... (priority 2). A reader who puts <other
priority> first would choose option n.

## What changes
[What_changes](#what-changes)

### Stage 1: <slug>
[Stage_1_<slug>](#stage-1-<slug>)

#### impl_<area>_<slug>

**<Title>.** <what changes, why, how, how it is tested>

## Consequences
[Consequences](#consequences)

- ...

Confidence: ...
Re-evaluate when ...

## Risks
[Risks](#risks)

**Backward compatibility, preserved.** ...
**Backward compatibility, not preserved.** ...

1. **<risk>.** <failure>. Guard: ...

## Affected systems and files
[Affected_systems_and_files](#affected-systems-and-files)

| System | Effect of this document | Where |
|---|---|---|
| Not affected | ... | |

## Testing
[Testing](#testing)

1. **<test>** (<where>, run by <who>). <What it proves, with the link; what
   is in place first; what is done; what is observed. One sentence for a
   simple test, the sequence and the timing for a scenario.>

## Open questions
[Open_questions](#open-questions)

- **1. <question>?** <how it gets answered>

### Decisions
[Decisions](#decisions)

| # | Question | Decision | Where it is argued |
|---|---|---|---|

## Relationship to other designs
[Relationship_to_other_designs](#relationship-to-other-designs)

- ...

### Future work: <slug>
[Future_work_<slug>](#future-work-<slug>)

## Advice
[Advice](#advice)

- <source> (YYYY-MM-DD): ...

## Appendix
[Appendix](#appendix)

### <catalogue by kind>
[<catalogue>](#<catalogue>)

#### <kind>_<area>_<slug>

| | |
|---|---|
| <field> | <value> |
```
