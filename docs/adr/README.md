# Architecture decision records

This folder holds the decision records of the JovianDSS Kubernetes CSI
plugin, one file per problem. This file describes the format of a record and
how one is written and maintained. It does not list the records: the folder
does, and the records themselves are the worked examples.

A record treats one problem: what is wrong, the facts it rests on, what we
care about, the options weighed, the decision and its reasons, what changes,
what follows, and how it is checked. Where a problem needs a full design, the
record is that design, from the problem to the last implementation item. A
design that predates this form stays in [`../design/`](../design/README.md)
until it is reworked into a record.

Writing the record is half its value: it forces the reasoning into the open
where it can be disagreed with. The other half is history: a reader years
later can see why the system is the way it is and how long each decision
governed the work.

The rules below have the same shape: the rule in bold, and under it its
parts, each on its own line: **Why**, its purpose, so that a reader can tell
when it does not apply; **Example**, an illustration that depends on no
record, though what never changes, such as a name, may be real; **Not**, what
the rule rules out; the last two where there is one; and a labelled part for
data of the rule's own, such as **Parts** or **The checks**. Tables hold
definitions; a rule stated inside a definition is part of it and has no Why
of its own.

## One problem, one record, one at a time

- **One problem per record. The problem is the unit, not the decision.**
  - **Why:** a problem may need several decisions, and they make sense only
    together.
  - **Example:** supporting a new API version needs the move to it, the reply
    types that match it and a client that is correct against it; none solves
    the problem alone, so they are one record, whose Options weigh the first
    two and whose questions decide the third, which the evidence dictates.
  - **Not:** a record per decision, each waiting on the others.
- **Records are not piled up: a record is implemented after it is accepted,
  and the next record, about the next problem, is written after that.**
  - **Why:** a record describes the current code, and the next one is written
    against the code the first one produced.
  - **Not:** several proposed records waiting on each other, which is a sign
    that the problem was cut in the wrong place.
- **An editorial choice gets no record; it is one line in a record's
  Decisions table.**
  - **Why:** a record is for a problem, and a choice of tense or of where
    release notes live has no problem behind it.
  - **Example:** where the changelog lives, decided as a numbered question of
    the record that first needed it.
  - **Not:** a record whose problem statement is a preference.
- **What a record does not solve is said in a sentence of its Problem
  statement and nowhere else; the record or place that handles it is named
  in Relationship.**
  - **Why:** a reader looking for a subject the record leaves out finds the
    answer in one place and is sent on from one place.
  - **Example:** a record on one product line says it leaves the other line
    out, and its Relationship names the record that will cover it.
  - **Not:** the exclusion repeated in Context, Consequences and Risks, each
    worded a little differently; every unrelated issue listed with a
    pointer.

## File name and status

- **A record is `NNNN-<problem-slug>.md`, a sequence number for order and
  then the problem it answers, and its title line is `# NNNN — <problem>`.**
  - **Why:** the name says the problem, not the decision, so a proposed
    record may change its decision without changing its name, and a later
    record that supersedes it about the same problem sorts next to it.
  - **Example:** a record that answers "support API vN on product line X" is
    `NNNN-api-vN-support-x.md`; its decision, whatever it becomes, is the bold
    paragraph of its Decision section.
  - **Not:** a name that says the decision, wrong the day the decision
    changes.
- **Numbers are never reused. The sequence of this folder starts at 0002;
  0001 is not used.**
  - **Why:** a number cited anywhere must keep pointing at the same record,
    and a superseded record keeps its text and its number.
- **The status is one word from this table, with a link when Superseded.**
  - **Why:** a reader must know whether the text can still change and what to
    read instead.

  | Status | Meaning |
  |---|---|
  | Proposed | under discussion; the text may still change |
  | Accepted | signed off by the approver; from then on the text is frozen |
  | Superseded | replaced by a later record, which the status line links; the text stays as it was |
  | Retired | the decision no longer applies and nothing replaces it; the text stays |

- **The date a decision was taken and the date the record was accepted are
  different things, and both go in Metadata.**
  - **Why:** the decision governs the work from its own date; acceptance
    freezes the text.

## Sections, in order

- **A section with nothing to say is kept, with one sentence saying so.**
  - **Why:** a reader then knows the subject was considered.
  - **Not:** a heading dropped because it was empty, which reads as an
    oversight.
- **Every `##` and `###` heading is followed by its self-anchor line**, a
  [Navigation](#navigation) rule.
- **The Problem statement is the problem the record solves, in a few plain
  sentences that read as one piece: what is wrong or missing, what change
  in the system's behaviour is wanted, what success looks like, what is
  explicitly not solved, in whichever of these the problem is said.**
  - **Why:** every later section is read against it, so a reader must be
    able to hold the whole of it in mind. The four questions are a way to
    understand a problem, not a form: the answer to one of them can be the
    whole statement.
  - **Example:** "The plugin writes every request to the log at Debug,
    credentials included. No log level may ever show a secret. Success is
    a Debug log an outsider could read. Which messages stay at Debug is not
    decided here."
  - **Not:** a bullet or a bold sentence per question; the tests that will
    prove success, or the issues the record leaves out, listed with a
    pointer each; the Context repeated with its evidence; the instances
    of the problem, such as the calls that fail, where the problem itself
    is meant.

| Section | What it holds |
|---|---|
| Table of Contents | the `##` and `###` headings, as underscore tokens; named headings (see [Names, not IDs](#names-not-ids)) are not listed |
| List of Tables | the numbered tables with their `tbl_*` tags, in order of appearance |
| Metadata | `Date:` when the decision of the Decision section was taken and when the text was written, the question decisions being dated in Advice; `Driver:`; `Approver:`; `Code:` the commit the line numbers refer to, called "the current code"; `Implementation:` where the work is described and how it is delivered; `Accepted:` the date, once it is; `Formerly:` the design the record was merged from and when, only for a record that absorbed one |
| Status | one word from the table above, with a link when Superseded |
| Problem statement | the problem the record solves, in a few plain sentences that read as one piece (the rule above): what is wrong or missing, what has to change, what success looks like, what is not solved here, in whichever of these the problem is said |
| Context | the facts the decision rests on, in subsections by subject: what exists today, what the evidence shows, what it does not show, each a claim with its data or a link to the Appendix. Ends with **Key observations**: the few conclusions the decision turns on, numbered |
| Priorities | the project values that weigh on this record, in the order they weigh here. Values are the project's and stable; priorities are this record's order among them, and the options argue against them by number; the number is the rank, so a priority added later takes its place and the ones below it move down, every reference rewritten in the same change |
| Options considered | the architectural and strategic decisions of the record: the ways the problem in the Problem statement could be solved, and which one is taken. The section answers the Problem statement and is read against it, each group of options (A, B, …) answering one part of the problem; a part whose fix follows from the evidence has no group, and its decisions are questions in the Decisions table. Decisions below that level, on how a chosen way is carried out, are questions too, never options. For each choice, what was evaluated: two or three real contenders, their trade-offs stated honestly against the priorities, one subsection per option, named by its group letter and its number within the group (Option A1, Option A2), the adopted one first and marked. Where only one option was considered, the subsection, the group letter alone without a number (Option B), says so and explains why, with the price that option carries; alternatives nobody weighed are not written up to fill the list. An option carries what distinguishes it, not the mechanism, which belongs to the implementation item |
| Decision | its own section, right after Options considered: the decision that has been taken, in a few sentences, bold so that it stands out; then the rationale naming each priority; then, for a choice with contenders, the sentence that tells a reader with another order of priorities which option they would pick; then, where question decisions shape the delivery, a paragraph naming them with their question numbers and a link to the Decisions table |
| What changes | the work: the commits in order, grouped in stages, each implementation item a named heading with what changes, why, how, and what must hold once it is done. The mechanism lives here and nowhere else. The unit tests that show an item works are the implementing agent's, not the record's ([The Testing section](#the-testing-section)) |
| Consequences | what becomes true once the record is applied, positive and negative: what is certain, while a possible failure is a risk; then the confidence level and the condition that should trigger a re-evaluation. It follows What changes because it refers to its stages and items |
| Risks | backward compatibility first, preserved and not preserved; then numbered risks, each with the concrete failure and its guard |
| Affected systems and files | a table by system first, then the files one row each with the stage that changes them; a last row for what is not affected |
| Testing | the tests that check the implementation, from the cheapest to the live one, as numbered steps; each a general description that an agent other than the author can construct and run; no unit tests ([The Testing section](#the-testing-section)) |
| Open questions | the open ones, as bullets with the number in bold, each with how it gets answered; then the subsection **Decisions**, a table of the decided questions with their numbers kept, one line of decision each and a link to the section that argues it |
| Relationship to other designs | the designs, records and repositories this record touches, one line each with the link; then **Future work** subsections for what is deliberately left to a later record |
| Advice | the inputs the decision was formed from, each with its date: findings, reviews, tracker entries, user decisions |
| Appendix | what is looked up rather than read: the evidence sources, the evidence gaps, the compliance rules and matrix, the finding catalogues by kind, the measurements, where the names and IDs come from, and the terms |

## The Testing section

- **The Testing section pins the conditions the implementation must meet
  that only the evidence or the lab know, and finds the edges the design did
  not foresee by running the result against the real thing.**
  - **Why:** a record is written before the code, so this section is where
    the implementation meets what the designer could not see from the design.
  - **Example:** a reply shape, a tool that refuses an image, a log that leaks
    secrets at Debug.
- **Unit tests are not listed: the agent implementing an item writes them as
  part of the item, and the item says only what must hold once it is done.**
  - **Why:** unit tests do neither of the section's jobs, and how many there
    are or how they are built is the implementer's.
  - **Example:** "a reply over the limit fails with an error naming the
    limit".
  - **Not:** a count of tests, a fake server described, a list of cases.
- **Each step is one test, described so that an agent other than the author
  can construct it and run it without asking.**
  - **Why:** the author is not there when it runs.
  - **Parts:**
    - what it proves: the edge or the condition, with a link to the finding,
      risk or question it answers;
    - where it runs and who runs it: the machine, the appliance or the build,
      and whether the user, the implementing agent or a tool drives it. A test
      the agent cannot run itself, because the lab is out of its reach, says
      so;
    - what has to be in place first: builds, settings, fixes committed, tools
      that exist, data to compare with. A prerequisite that does not exist yet
      is named, with where it is tracked and who provides it;
    - what is done: the command or the sequence, concrete enough to repeat;
    - what counts as pass, as something observed: a log line, a status, a
      diff, a comparison with a named baseline; and what is recorded when the
      result is neither pass nor fail, so that an unknown outcome becomes a
      result instead of being passed over.
- **How much a step spells out follows the nature of the test.**
  - **Why:** the parts above are what a reader needs, in as many words as the
    test takes, not a form to fill in.
  - **Example:** a simple test is a sentence: delete a volume, then send the
    same delete again, and check that the REST request that now fails is
    handled as "already gone" and not as an error. A scenario needs its
    sequence, its timing and what to watch: shut the appliance off by force
    while create and delete requests are in flight, and record how the
    cluster behaves under that load and what each request is answered while
    the pool moves to the other node.
- **The steps go from the cheapest to the live one, and the section ends
  with what the tests cannot reach.**
  - **Why:** a reader knows which findings have no cover in this delivery.
- **A test that depends on an implementation detail says so, and that
  detail becomes part of the item it depends on.**
  - **Why:** otherwise the implementer may remove, without knowing it, what
    the test reads.
  - **Example:** a log line keeping its query string, an error text staying
    searchable.
- **Results go to the testing repository with the rest of the lab's data,
  not into the record.**
  - **Why:** the record says what must hold; the lab's data changes with
    every run and carries lab details.

## How the text is organised

- **Claim, explanation, data, in that order, in every block of every
  section.**
  - **Why:** a reader who trusts the claim stops after the first sentence; a
    reader who doubts it finds the evidence in the same place every time.
  - **Example:** the claim as one bold sentence or as the heading; the
    explanation in a few plain sentences; the data and the references last,
    as a short table or a labelled line.
  - **Not:** details dropped to make a record shorter; they are placed, not
    cut.
- **Each fact once, at one place; when two sections start to say the same
  thing, one of them links to the other.**
  - **Why:** a fact stated twice drifts into two facts.
  - **Example:** Options carry the alternatives and their price; What changes
    carries the mechanism; the Appendix carries the catalogue; Context
    carries the claims and links to the catalogue; the Problem statement
    says the problem and leaves the evidence to Context; the Decisions
    table indexes the decisions and does
    not restate their arguments.
- **Plain language; jargon is explained once, in Terms.**
  - **Why:** a reader from outside the project must be able to follow the
    argument.
- **No history and no dates in the body.**
  - **Why:** the body describes the current code and what the record changes;
    dates belong to Metadata and Advice, where a reader looks for them.
  - **Not:** "on the day the evidence came in we decided" in Context.
- **The tense rule: while Proposed the record says what "would" change;
  facts about the current code and the evidence are in the present tense;
  after implementation the tense is converted, not the text rewritten.**
  - **Why:** the same text serves before and after implementation.
- **The record calls itself "this document", never "this design" or "this
  record".**
  - **Why:** the text reads the same whether the reader came for the design
    or for the decision.
- **References are links: `file:line` of the commit named in Metadata, as
  relative links into the code where a link helps; sections by their anchor;
  the spec repository's findings by their number.**
  - **Why:** a link is checked by a script and followed by a click; the spec
    repository is read only for this work, so its findings keep its numbers.
- **No test-environment details and no secrets in a record.**
  - **Why:** machine names, pool names and addresses change with the lab and
    belong to the testing repository, which holds them; a secret in a
    document is a leak.
  - **Example:** "the lab appliance", not its address.
  - **Not:** a pool name in a Testing step.

## Names, not IDs

- **Findings, gaps, compliance rules and implementation items are named
  `<kind>_<area>_<slug>`, not numbered.**
  - **Why:** a human reader has to keep track of what is referenced. A name
    says what the thing is; a letter and a number say nothing, are hard to
    remember and get confused with each other. A name never changes; a number
    is a position in a list that grows.
  - **Example:** `gap_jdss_rest_empty_body_on_get_delete`,
    `bug_rest_v4_pools_pool_get_iostats_type_mismatch`,
    `rule_rest_reply_types_decode`, `impl_plugin_driver_validate_names_and_ids`.
  - **Not:** the letter-and-number IDs of the reviews, F8, C3, G5, A11, which
    are mapped to names (below).
- **The kind comes first and says whether the name is a finding or a
  prescription.**
  - **Why:** a reader knows from the first word whether the entry describes
    what is or prescribes what is done.

  | Kind | Names | Where its heading lives |
  |---|---|---|
  | `bug` | behaviour of the code that is wrong | the Appendix catalogue of its subject |
  | `crash` | a panic the code can reach | Appendix, Crashes |
  | `dataloss` | a wrong result that loses or corrupts data | Appendix |
  | `security` | a finding of the security review | Appendix, Security findings |
  | `gap` | evidence that is missing: a case no recording shows | Appendix, Evidence gaps |
  | `rule` | a compliance rule the code is checked against | Appendix, What "compliant" means |
  | `impl` | an implementation item: a change the record delivers, which implements the decision | What changes, under its stage |

  - **Definitions:** the first five are findings: they describe what is, and
    their field table says where, what happens, the evidence, and which
    implementation item or stage fixes them. A rule says what must hold. An
    implementation item says what is done, whatever its form: a repair, a new
    file, a tool, a setting.
- **An implementation item names the findings and rules it closes, and a
  finding names the item or stage that fixes it.**
  - **Why:** the two directions link each other, so a reader goes from a
    finding to its fix and back.
- **Then the area; lower-case letters, digits and underscores only.**
  - **Why:** the area tells a reader which system the name belongs to before
    the slug is read, and the character set makes the name its own anchor.
  - **The areas:** `jdss_rest` (what the appliance's REST API does),
    `plugin_rest` (`pkg/rest`), `rest` (the API as the contract between the
    two: the compliance rules, and what fails to comply with the spec),
    `plugin_driver`, `plugin_ctrl`, `plugin_node`, `plugin_cli`, `plugin`
    (several packages), `lab`, `spec`, `docs`, `build`. A new concern adds a
    new area word; nothing existing changes.
- **Then the slug, given with care: it catches the essence of the finding or
  the item, object first, precisely enough that no later item can be confused
  with it, and is no longer than that takes.**
  - **Why:** a name is read by a human who must know what it refers to
    without looking it up, and it is never changed, so a vague name blocks
    every later name near it.
  - **Example:** for a REST call, the API version, the path with its
    placeholders as words, the method, then what is wrong:
    `bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec`.
    For an object, the object first: `volume_rollback`.
  - **Not:** `bug_plugin_rest_nas_volume_keys_never_match`, a symptom in the
    code that could belong to several endpoints and versions;
    `rollback_volume`, the object not first.
- **Each named thing is a `####` heading with a small field table under it:
  in the Appendix for findings, gaps and rules, in What changes for
  implementation items.**
  - **Why:** the heading is its own anchor and a single word, so `*` in vim
    and a link on GitHub both reach it; such headings need no self-anchor line
    and are not listed in the Table of Contents.
- **A name is a link, `[name](#name)`, wherever it is mentioned; a name that
  has no heading in the record is written in code font without a link.**
  - **Why:** a link is checked, every linked name has a heading, and code
    font marks a name that is described only in a review or the tracker.
  - **Not:** a bare name in prose.
- **Names have no order: list them, or name the group and link the section
  that holds it.**
  - **Why:** a range of names tells a reader nothing.
  - **Example:** "the stage 4 items", "the first eight gaps under Evidence
    gaps".
  - **Not:** "`bug_a` to `bug_b`".
- **The older letter-and-number IDs live on in the reviews and in
  kub-testing's `ISSUES.md`; the map between them and the names is
  `kub-testing/issues/`, one file per old prefix, which is also where a new
  name is registered.**
  - **Why:** the reviews are not rewritten, and one registry keeps a name
    unique across documents.
- **A name is chosen once and never changed after it is referenced; a
  changed understanding gets a new name, and the old one is marked
  superseded.**
  - **Why:** a name is cited from documents that are not rewritten.
  - **Not:** a rename to a better slug.
  - **Exception:** the kind was `work` until 2026-11-11 and became `impl`
    before any document outside the record used the names; on 2026-11-12,
    when the slug rule above was expanded, 38 names were given new names
    under it, each with its old row marked superseded in the registry.
- **Numbers stay where they identify an entry: risks, questions, testing
  steps, stages, key observations. They are never renumbered once
  referenced; new entries are appended. Numbers that rank or place, tables
  and priorities, follow the position (see [Navigation](#navigation)): a
  priority added later takes its place in the order, the ones below it move
  down, and every "priority n" reference is rewritten in the same change.**
  - **Why:** a reference such as "risk 9" or "question 27" keeps its
    meaning; "priority 2" means the second weight, which is the thing that
    changes.
  - **Not:** closing a gap by renumbering; a priority appended as number
    eight that weighs second.
- **Another document's numbering is cited as that document's.**
  - **Why:** its numbers are its own and change with it.
  - **Example:** the hurl suite's Findings, a design's items, the testing
    repository's rows.
- **A list that has gaps is written as a bullet list with the number inside
  the bold, `- **30. <question>?**`.**
  - **Why:** a renderer takes an ordered list's numbers from its first item
    and renumbers the rest, so the numbers a reader saw would not be the ones
    written.
  - **Example:** the open questions, once some are decided and moved to the
    Decisions table.
  - **Not:** `30.` as an ordered-list item, which renders with the number of
    its position: 24 for the third item of a list that starts at 22.

## Navigation

- **Every `##` and `###` heading is followed by a line that links to itself,
  with the heading's words joined by underscores, and the Table of Contents
  uses the same tokens.**
  - **Why:** `*` in vim jumps between an entry and its section, and the links
    work on GitHub.
  - **Example:** `[Open_questions](#open-questions)`.
- **A table that carries a number has a caption
  `` **Table n — Title** · `tbl_tag` `` and appears in the List of Tables;
  numbering follows order of appearance, so a table inserted or removed
  renumbers those after it, and every "Table n" reference is updated in the
  same change.**
  - **Why:** a reader finds a table by its number in the List of Tables, and
    a script by its tag.
  - **Not:** a catalogue that became headings keeping its table number; it is
    no longer a table.
- **A heading that something links to is not renamed; when a section has to
  move or change its name, every link to it is updated in the same change.**
  - **Why:** a broken anchor fails silently on GitHub.
  - **Not:** a rename left for a later commit.

## Maintaining a record

- **While Proposed, the text changes freely, and every change keeps the
  checks below green.**
  - **Why:** a reviewer reads the whole text, so a change must not leave it
    broken anywhere.
  - **Example:** a new finding gets a name, a heading in the Appendix, a row
    in the mapping file and links from where it matters; a new question gets
    the next number and, when decided, moves to the Decisions table with its
    number kept; a new risk is appended with the next number; a Testing step
    folded into another keeps its number, with one line saying where it went.
- **Once Accepted, the text is frozen; a change of decision is a new record
  that supersedes this one, and the status line links it.**
  - **Why:** the frozen text is the history half of a record's value.
  - **Not:** an edit to an accepted record to improve its wording.
- **Checks before a review, by script or by hand.**
  - **Why:** the reviewer's time goes to the content, not to what a script
    can find.
  - **Scripts:** `tools/check_record.py`, and `tools/check_names.py` for the
    names, which reads `kub-testing/issues/` from a checkout beside this
    repository.
  - **The checks:**
    - every `](#anchor)` resolves to a heading;
    - every entry in the Table of Contents and the List of Tables matches a
      heading or a caption;
    - risks, steps, stages, key observations, priorities and tables are
      numbered without gaps or duplicates;
    - questions have no duplicates, their numbers are kept when they move to
      the Decisions table, and the open ones are written as bullets with the
      number in bold;
    - every "Table n" reference is within range;
    - every Testing step says where it runs and who runs it;
    - every name is unique and every linked name has a heading;
    - no letter-and-number ID remains outside the ID sources table and the
      citations of other documents' numbering;
    - every relative link to a file exists;
    - no horizontal rule;
    - no sentence over sixty characters repeated outside tables;
    - no link to the section it sits in;
    - no "design NNNN" with the record's own number, "this design" or "this
      record" phrasing has crept back in.
    - no name that the registry marks superseded is written in the record.

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
  - [Future_work_<slug>](#future-work-<slug>)
- [Advice](#advice)
- [Appendix](#appendix)
  - [Evidence](#evidence)
  - [Evidence_gaps](#evidence-gaps)
  - [What_compliant_means](#what-compliant-means)
  - [Compliance_matrix](#compliance-matrix)
  - [<catalogue by kind>](#<catalogue>)
  - [Measurements](#measurements)
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

<A few plain sentences that read as one piece: what is wrong or missing,
what has to change, what success looks like, what is not solved here.>

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

<what distinguishes it, in a few sentences>

Advantages: ... (priorities n, m)
Disadvantages: ... (priority k)

### Option A2: <name>
[Option_A2_<slug>](#option-a2-<slug>)

<what distinguishes it, in a few sentences>

Advantages: ... (priority k)
Disadvantages: ... (priorities n, m)

### Option B: <name>, the only option considered
[Option_B_<slug>_the_only_option_considered](#option-b-<slug>-the-only-option-considered)

Only one option was considered. <why; the price it carries>

## Decision
[Decision](#decision)

**<the decision that has been taken, in a few sentences>.**

Rationale: ... (priority 1) ... (priority 2). A reader who puts <other
priority> first would choose option n.

<where question decisions shape the delivery: the paragraph naming them, with
their question numbers and a link to the Decisions table>

## What changes
[What_changes](#what-changes)

### Stage 1: <slug>
[Stage_1_<slug>](#stage-1-<slug>)

#### impl_<area>_<slug>

**<Title>.** <what changes, why, how, what must hold once it is done>

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
   is in place first; what is done; what is observed as pass, and what is
   recorded when the result is neither pass nor fail. One sentence for a
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

### Evidence
[Evidence](#evidence)

### Evidence gaps
[Evidence_gaps](#evidence-gaps)

#### gap_<area>_<slug>

### What "compliant" means
[What_compliant_means](#what-compliant-means)

#### rule_<area>_<slug>

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

### <catalogue by kind>
[<catalogue>](#<catalogue>)

#### <kind>_<area>_<slug>

| | |
|---|---|
| <field> | <value> |

### Measurements
[Measurements](#measurements)

### ID sources
[ID_sources](#id-sources)

### Terms
[Terms](#terms)
```
