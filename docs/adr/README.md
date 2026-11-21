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


## Rules

The rule name is in bold, and under it its
parts, each on its own line: **Why**, its purpose, so that a reader can tell
when it does not apply; **Example**, an illustration that depends on no
record, though what never changes, such as a name, may be real; **Not**, what
the rule rules out; the last two where there is one; and a labelled part for
data of the rule's own, such as **Parts** or **The checks**. Tables hold
definitions; a rule stated inside a definition is part of it and has no Why
of its own.

### One problem, one record, one at a time

- **One problem per record. The problem is the unit, not the decision.**
  - **Why:** a problem may need several decisions, and they make sense only
    together.
  - **Example:** supporting a new API version needs the move to it, the reply
    types that match it and a client that is correct against it; none solves
    the problem alone, so they are one record: its Options weigh the move,
    and its questions decide the reply types and, as the evidence dictates,
    the client.
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
- **Problem statement formalizes the scope and what has to be done.
    Scope limitations and explicit mentions of what will not be done or accomplished in given ADR
    are listed in Options considered and accepted in Decision section.**
  - **Why:** Problem statement should be clear and minimalistic in its description.
    Purpose of ADR is to formalize thought process on the matter.
    In its progression decisions being made in the process of development.
    Different solutions / architectural approaches may involve different effort,
      different number of components or system to be changed, created, deleted.
    That may require additional clarification of scope, that is a part of
    considering different options and making appropriate decisions.
  - **Example:**
    ```
    The plugin uses JovianDSS REST API v3, which is deprecated, through a client
    that follows documentation and assumptions rather than what the appliance
    serves, and it pages its CSI lists in a way that crashes the controller or
    returns entries twice or never. It has to use API v4 through a client that
    complies with the API as served, down to the way the appliance pages its
    lists. JovianDSS 2 will not be supported within this iteration.
    ```
    Redundant part: `JovianDSS 2 will not be supported within this iteration.`
  - **Not:** the exclusion defined in Problem statement

### File name and status

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
  0001 is not used, since design 0001 in `docs/design/` is a predecessor
  of these records.**
  - **Why:** a number cited anywhere must keep pointing at the same record,
    and a superseded record keeps its text and its number. Record 0002
    keeps the number of design 0002, which it reworks.
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

## The Testing section

- **The Testing section describes the corner cases that have to be tested,
  such as the conditions the implementation must meet that only the evidence
  or the lab know. It is not a description of the tests themselves but the
  foundation for developing them later.**
  - **Why:** a record is written before the code, so this section is where
    the implementation meets what the designer could not see from the design.
  - **Example:** a reply shape, a tool that refuses an image, a log that leaks
    secrets at Debug.
- **Unit tests are not listed: the agent implementing an item writes them as
  part of the item, and the item says only what must hold once it is done.**
  - **Why:** the record does not have the full, exact picture of what will be
    implemented and how, so it cannot guide how unit tests are written; the
    agent implementing the item has that picture.
  - **Example:** "a reply over the limit fails with an error naming the
    limit".
  - **Not:** a count of tests, a fake server described, a list of cases.
- **Each step names one corner case, what has to hold there, and the
  finding, risk or question it comes from. It says where only when the case
  can be seen in one place alone.**
  - **Why:** a step is what a test is built from later. Who runs the test,
    what has to be in place first and its commands are known when the test
    is written, not when the record is, so they belong to the test.
  - **Parts:**
    - the corner case: the edge or the condition, with a link to the
      finding, risk or question it comes from;
    - what has to hold there, as something someone can observe: a status, a
      reply, a log line, what is left on the appliance; and what counts as
      a result when the case cannot be brought about, so that an unknown
      outcome becomes a result instead of being passed over;
    - where, only when the case can be seen in one place alone, such as the
      lab appliance.
  - **Example:** "**A reply larger than `maxreplysize`**
    (`impl_plugin_rest_reply_size_limit`). The call fails with an error that
    names the limit, and no more of the reply is read."
  - **Not:** "**Reply size test** (the lab machine, run by the user). Build
    the plugin, set `maxreplysize: 1000`, run `dev_cycle.py all` and look for
    the error in `plugin.log`": who runs it, the setting and the commands
    belong to the test written from the step.
- **How much a step spells out follows the nature of the case.**
  - **Why:** the parts above are what a reader needs, in as many words as the
    case takes, not a form to fill in.
  - **Example:** a simple case is a sentence: a volume deleted twice, whose
    second delete is handled as "already gone" and not as an error. A
    scenario needs its sequence, its timing and what to watch: an appliance
    shut off by force while create and delete requests are in flight, how
    the cluster behaves under that load, and what each request is answered
    while the pool moves to the other node.
- **The steps go from the cheapest to the live one, and the section ends
  with what the tests cannot reach.**
  - **Why:** a reader knows which findings have no cover in this delivery.
- **A step that depends on an implementation detail says so, and that
  detail becomes part of the item it depends on.**
  - **Why:** otherwise the implementer may remove, without knowing it, what
    the test built from the step reads.
  - **Example:** a log line keeping its query string, an error text staying
    searchable.
- **Results go to the testing repository with the rest of the lab's data,
  not into the record.**
  - **Why:** the record says what must hold; the lab's data changes with
    every run and carries lab details.

### How the text is organised

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
- **The mood rule: the record describes the changes it makes in the
  imperative mood, as a commit message of the Linux kernel does; what
  would follow from them, as in Consequences and Risks, is in the
  conditional; facts about the current code and the evidence are in the
  present tense. The wording is not changed once the record is accepted or
  implemented.**
  - **Why:** the kernel's guide to submitting patches asks to "describe
    your changes in imperative mood, as if you are giving orders to the
    codebase to change its behaviour". An order reads the same before and
    after the work is merged, so the text, frozen at acceptance, stays
    true; the present tense of the facts is about the commit named in
    Metadata.
  - **Example:** "make `getResourcesList` read the list once on every
    call"; "send `per_page=100` on every list request".
  - **Not:** "`getResourcesList` would read the list once", "this commit
    makes `getResourcesList` read the list once" or "I changed
    `getResourcesList` to read the list once"; nor an order rewritten as
    "does" once the work is merged.
- **The tense rule: what is, in the present; what happened, in the past;
  nothing in the future. A change, and what would follow from it, take the
  mood rule's imperative and conditional.**
  - **Why:** the wording is frozen at acceptance and read before and after
    the work is merged. The present stays true of the commit named in
    Metadata and of the conditions the work has to meet, and the past of a
    run or a check once it has happened; "will" is true only until the
    work is merged.

  | Text | Tense | Example |
  |---|---|---|
  | the current code, the evidence and what it shows | present | "the driver cuts the list at `max_entries` before the controller skips anything" |
  | a decision as it stands | present | "the plugin takes TLS 1.2 or higher" |
  | a condition that has to hold: What must hold, a Testing step, a rule | present | "a GET answered 204 is sent again after `retrydelay`" |
  | what happened at one time: a run, a recording, a check, a review | past | "129 characters were recorded accepted, 150 refused" |

  - **Not:** "the plugin will send `per_page=100`"; "the driver cut the
    list at `max_entries`" for today's code; "a GET answered 204 was sent
    again" in What must hold.
- **The record calls itself "this document", never "this design" or "this
  record".**
  - **Why:** the text reads the same whether the reader came for the design
    or for the decision.
- **References are links: `file:line` of the commit named in Metadata, as
  relative links into the code where a link helps; sections by their anchor;
  the findings of `jdss-rest-api-spec` by their number.**
  - **Why:** a link is checked by a script and followed by a click;
    `jdss-rest-api-spec` is read only for this work, so its findings keep its
    numbers.
- **No test-environment details and no secrets in a record.**
  - **Why:** machine names, pool names and addresses change with the lab and
    belong to the testing repository, which holds them; a secret in a
    document is a leak.
  - **Example:** "the lab appliance", not its address.
  - **Not:** a pool name in a Testing step.

### Names, not IDs

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
- **The kind comes first and says what the name is: a finding, a gap, a
  rule or an implementation item.**
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
  | `impl` | an implementation item: a change the record delivers, which implements the decision | What changes, under its stage or its commit |

  - **Format:** `tools/check_names.py` reads the kinds from the rows of this
    table that start with a kind in backquotes, and the areas from the
    paragraph that begins "The areas:" and ends with `build`; both keep
    that form.

  - **Definitions:** the first four are findings: they describe what is, and
    their field table says where, what happens, the evidence, and which
    implementation item, stage or commit they are assigned to
    ([Appendix](#appendix)). A gap says what no source shows, and its
    field table why that matters and what the attempts to close it showed. A
    rule says what must hold. An implementation item says what is done,
    whatever its form: a repair, a new file, a tool, a setting.
- **An implementation item names the findings and rules it closes, and a
  finding names the item, stage or commit it is assigned to.**
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
- **Each named thing is a `####` heading, the reference point that the rest
  of the record and other documents link to: in the Appendix for findings,
  gaps and rules, each with a field table under it; in What changes for
  implementation items, which are prose.**
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


### Navigation

- **Every `##` and `###` heading is followed by a line that links to itself,
  with the heading's words joined by underscores, and the Table of Contents
  uses the same tokens.**
  - **Why:** `*` in vim jumps between an entry and its section, and the links
    work on GitHub.
  - **Example:** `[Open_questions](#open-questions)`.
- **Every table but a field table carries a number, whether the text refers
  to it or not.**
  - **Why:** the List of Tables then shows every table, and citing a table
    later renumbers nothing; a field table is found by its heading.
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
    folded into another, or moved to the testing repository, keeps its
    number, with one line saying where it went.
- **Once Accepted, the text is frozen; a change of decision is a new record
  that supersedes this one, and the status line links it.**
  - **Why:** the frozen text is the history half of a record's value.
  - **Not:** an edit to an accepted record to improve its wording.
- **Checks before a review, by script or by hand.**
  - **Why:** the reviewer's time goes to the content, not to what a script
    can find.
  - **By hand:** the checks marked so below; the scripts run the others,
    and a record they fail is not ready for review.
  - **Scripts:** `tools/check_record.py`, and `tools/check_names.py` for the
    names, which reads `kub-testing/issues/` from a checkout beside this
    repository.
  - **The checks:**
    - every `](#anchor)` resolves to a heading;
    - every `##` and `###` heading has its anchor line under it, such as
      `[Decisions](#decisions)`;
    - every entry in the Table of Contents and the List of Tables matches a
      heading or a caption;
    - risks, steps, stages, key observations, priorities and tables are
      numbered without gaps or duplicates;
    - questions have no duplicates, in the open list or in the Decisions
      table, their numbers are kept when they move to the Decisions table, and the open ones are written as bullets with the
      number in bold;
    - every "Table n" reference is within range;
    - every Testing step names the finding, risk or question it comes from
      (by hand);
    - every name is unique and every linked name has a heading;
    - no letter-and-number ID remains outside the ID sources table and the
      citations of other documents' numbering (by hand: the script lists the
      IDs it finds);
    - every relative link to a file exists;
    - no horizontal rule;
    - no sentence over sixty characters repeated outside tables;
    - no link to the section it sits in, except from a section that numbers
      or catalogues its entries to those entries: Key observations, What
      changes, Risks, Testing, Open questions, Decisions, the Appendix and its
      Evidence gaps;
    - none of these phrases: "design NNNN" or "ADR NNNN" with the record's
      own number, "this design", "this record", "this ADR", "the design"
      and "the design's" in any case (another design is cited as "design
      NNNN"), "earlier text", "earlier list" and "as listed above";
    - no name that the registry marks superseded is written in the record.

## Sections

A record has the sections below, in this order; the [Skeleton](#skeleton)
shows them empty. Each is described the same way: the sentence in bold says
what the section is; **Answers**, the question of the reader it answers;
**Holds**, what goes into it, in order; then **Why**, **Example** and
**Not**, as in every rule.

- **A record is one argument: each section answers one question, in the
  order a reader asks them, and rests on the sections before it.**
  - **Why:** a reader who disagrees can find the step they disagree with, a
    fact, the order of the priorities or the price of an option, and see
    how the conclusion would shift. The discussion is then about that step
    and not about the whole record.
  - **Not:** a section filled in because the skeleton has it, which answers
    no question of the reader.
- **A section with nothing to say is kept, with one sentence saying so.**
  - **Why:** a reader then knows the subject was considered.
  - **Not:** a heading dropped because it was empty, which reads as an
    oversight.
- **Every `##` and `###` heading is followed by its self-anchor line**, a
  [Navigation](#navigation) rule.

### Table of Contents

- **The Table of Contents lists the `##` and `###` headings of the record,
  as underscore tokens.**
  - **Answers:** "how is this laid out, and where is the part I need?"
  - **Holds:** one entry per heading, in order; named headings (see
    [Names, not IDs](#names-not-ids)) are not listed.
  - **Why:** a record is consulted as often as it is read through, and a
    token is one word, so `*` in vim reaches its section
    ([Navigation](#navigation)).

### List of Tables

- **The List of Tables lists the numbered tables with their `tbl_*` tags,
  in order of appearance.**
  - **Answers:** "which tables are there, and where?"
  - **Holds:** one line per table: its number, its tag and its title.
  - **Why:** the tables carry much of the data, and a reader looking for
    one finds it here without scrolling ([Navigation](#navigation)).

### Metadata

- **Metadata says which decision the record is, who stands behind it and
  which commit its facts are true of.**
  - **Answers:** "what am I holding, and as of when?"
  - **Holds:** `Date:` when the decision of the Decision section was taken
    and when the text was written, the question decisions being dated in
    Advice; `Driver:`; `Approver:`; `Code:` the commit the line numbers
    refer to, called "the current code"; `Implementation:` where the work
    is described and how it is delivered; `Accepted:` the date, once it
    is; `Formerly:` the design the record was merged from and when, only
    for a record that absorbed one.
  - **Why:** every "today" and every line number in the body is true of one
    commit, and the decision of one date; a reader has to know both before
    relying on either.

### Status

- **Status is one word from the table in
  [File name and status](#file-name-and-status), with a link when
  Superseded.**
  - **Answers:** "can this text still change, and does its decision still
    govern?"
  - **Why:** it has a section of its own so that it is found at the same
    place in every record and changes in one line.

### Problem statement

**Problem statement** defines
  - the problem the record solves
  - an issue that record addresses
  - goal that needs to be reached by the record
  - new feature or improvement that is introduced by the record

**Answers** 
  - what is wrong
  - what is missing
  - what change in the system's behaviour is wanted
  - what desired operation / capability looks like

**Why**
Core section that defines all following sections.
Must be laconic, succinct and clear on topic.
Every later section is made to answer challenge defined in it. 
AI agent and human reader must be able to hold the whole of it in mind and have clear understanding of the
matter that particular ADR addresses.

**Examples**
  - cluster prefix
    ```
    A JovianDSS pool is a storage resource that can be accessed by multiple
    Proxmox clusters simultaneously. Without isolation, every cluster's plugin
    instance queries the same `volumes list` endpoint and receives all volumes in
    the pool — including volumes owned by other clusters.
    
    The consequences:
    
    - **VM ID collision.** Both clusters may have a VM 100. Both create
      `v_vm-100-disk-0`. The `list_images` output on each cluster contains the
      other cluster's disk, making it appear as an unrecognised orphan.
    - **Accidental deletion.** `pvesm free` or `qm destroy --purge` on cluster A
      may find and delete a volume that belongs to cluster B.
    - **Naming collisions.** `find_free_diskname` scans existing volume names to
      pick the next free index. If cluster B's volumes are included in the scan,
      cluster A skips indices it does not own, wasting names and causing
      unnecessary retries.
    ```

  - api v4 compliance jdss1
    ```
    The plugin uses JovianDSS REST API v3, which is deprecated, through a client
    that follows documentation and assumptions rather than what the appliance
    serves, and it pages its CSI lists in a way that crashes the controller or
    returns entries twice or never. It has to use API v4 through a client that
    complies with the API as served, down to the way the appliance pages its
    lists.
    ```

**Wrong**
  - list of answers to all questions provided above in form of bullet or a bold sentence per question
    That is wrong because it blurs definitions and context becomes unclear
  - list of the tests that will prove success
    ADR is built around features and problems, not passing specific tests
  - list of the issues the record leaves out
    ADR's problem statement describes the things that ADR is focused on, bringing things that are not touched
    in the given ADR will only confuse user or agent
    if there is a need to narrow down the scope, it should be done in `options considered` and `decisions` sections
  - the Context repeated with its evidence
    repeating same things over is a bad practice and should be avoided
  - the instances of the problem, such as the calls that fail, where the problem itself is meant

### Context

- **Context is the facts the decision rests on: what exists today, what the
  evidence shows and what it does not show.**
  - **Answers:** "what do we know?"
  - **Holds:** subsections by subject, each a claim with its data or a link
    to the Appendix; then **Key observations**: the few conclusions the
    decision turns on, numbered.
  - **Why:** a decision is as good as the facts under it, so they are told
    in a way that a reader who does not live in this code can follow the
    rest of the record from there, and what is not known is said as
    plainly as what is: a gap passed over in silence comes back as a
    surprise in the lab. Key observations is where the facts become the
    conclusions the options are argued from; a reader who accepts them
    should find the decision unsurprising.
  - **Not:** a tour of the code; a fact no later section uses; a key
    observation that no subsection supports; a finding described here and
    again in the Appendix.

### Priorities

- **Priorities is the project values that weigh on this record, in the
  order they weigh here.**
  - **Answers:** "what do we care about here, and in which order?"
  - **Holds:** a numbered list, one value each, with a sentence where the
    value needs one. Values are the project's and stable; priorities are
    this record's order among them, and the options argue against them by
    number; the number is the rank, so a priority added later takes its
    place and the ones below it move down, every reference rewritten in
    the same change.
  - **Why:** an option is not good or bad in itself, only against what is
    valued. With the order stated before any option is named, the choice
    follows from it, and a reader who orders the values differently sees
    which choice would change.
  - **Not:** priorities written after the decision, to fit it; a priority
    that no option, decision or risk is argued against.

### Options considered

- **Options considered is the architectural and strategic decisions of the
  record: the ways the problem in the Problem statement could be solved,
  and which one is taken.**
  - **Answers:** "which ways were there, and what does each cost?"
  - **Holds:** the section answers the Problem statement and is read
    against it, each group of options (A, B, …) answering one part of the
    problem; a part whose fix follows from the evidence has no group, and
    its decisions are questions in the Decisions table. Decisions below
    that level, on how a chosen way is carried out, are questions too,
    never options. For each choice, what was evaluated: two or three real
    contenders, their trade-offs stated honestly against the priorities,
    one subsection per option, named by its group letter and its number
    within the group (Option A1, Option A2), the adopted one first and
    marked. Where only one option was considered, the subsection, the
    group letter alone without a number (Option B), says so and explains
    why, with the price that option carries. An option carries what
    distinguishes it, not the mechanism, which belongs to the
    implementation item.
  - **Why:** a decision is understood by what was given up for it. An
    option told fairly enough that someone who prefers it would recognise
    it lets a later reader, whose priorities have changed, reopen the
    choice without doing the work again.
  - **Not:** an option set up to lose; an alternative nobody weighed,
    written up to fill the list; the adopted option without a
    disadvantage.

### Decision

- **Decision is the decision that has been taken, in a few sentences, bold
  so that it stands out, and the reasons for it.**
  - **Answers:** "what was decided, and why this?"
  - **Holds:** its own section, right after Options considered: the
    decision, in sentences that can be quoted alone; then the rationale
    naming each priority, with the price that is accepted and what stays
    without a guard; then, for a choice with contenders, the sentence that
    tells a reader with another order of priorities which option they
    would pick; then, where question decisions shape the delivery, a
    paragraph naming them with their question numbers and a link to the
    Decisions table.
  - **Why:** this is what other documents cite and what a reader years
    later looks for. A reason that names its priority can be checked
    against the options, and a price that is stated was accepted, not
    overlooked.
  - **Not:** the mechanism, which is What changes; a rationale of benefits
    only.

### What changes

- **What changes is the work: the commits in order, grouped in stages, and
  each implementation item under its name.**
  - **Answers:** "what exactly is done, and in which order?"
  - **Holds:** the stages, a commit that has to come before the first
    stage under a heading of its own; each implementation item a named
    heading with what changes, why, how, and what must hold once it is
    done, the changes in the imperative mood
    ([How the text is organised](#how-the-text-is-organised)). The
    mechanism lives here and nowhere else. The unit tests that show an
    item works are the implementing agent's, not the record's
    ([The Testing section](#the-testing-section)).
  - **Why:** a decision governs only through the work, so the work is
    written precisely enough that someone who took no part in the
    discussion can do it from here, and a reviewer can tell when an item
    is done. An item without what must hold cannot be checked, and an
    order of commits whose reason is not given gets changed.
  - **Not:** the argument for the decision told again; code written out
    where its shape is enough.

### Consequences

- **Consequences is what becomes true once the record is applied, positive
  and negative.**
  - **Answers:** "what is true once this is applied, and how sure are we?"
  - **Holds:** what is certain, while a possible failure is a risk; then
    the confidence level and the condition that should trigger a
    re-evaluation. It follows What changes because it refers to its stages
    and items.
  - **Why:** a decision is accepted for what follows from it, so that is
    said in one place and not left to be worked out from the items. The
    confidence and the condition for re-evaluation tell a later reader
    when the record stops being a safe guide.
  - **Not:** advantages only; a failure that may happen, which is a risk.

### Risks

- **Risks is what can go wrong, for whom, and what stops it.**
  - **Answers:** "what can go wrong, and what do I lose on upgrade?"
  - **Holds:** backward compatibility first, preserved and not preserved;
    then numbered risks, each with the concrete failure, as something that
    would be seen, and its guard, or the plain statement that it has none.
  - **Why:** reviewers and operators read this section first. A failure
    described concretely can be tested and watched for; a risk without a
    guard, said so, is a price accepted knowingly.
  - **Not:** "may cause problems"; a risk left out or softened so that the
    record looks safe.

### Affected systems and files

- **Affected systems and files is the reach of the work.**
  - **Answers:** "what does this touch?"
  - **Holds:** a table by system first, then the files one row each with
    the stage that changes them; a last row for what is not affected.
  - **Why:** a reviewer sees the reach at a glance; whoever keeps another
    system, the lab, the documentation or another repository, learns that
    it concerns them; and the row for what is not affected ends a search
    before it starts.

### Testing

- **Testing is the corner cases that have to be tested, as numbered steps
  from the cheapest to the live one, and at the end what the tests cannot
  reach.**
  - **Answers:** "how will we know it works, and what will we still not
    know?"
  - **Holds:** mostly behaviour, integration and what has to be tested in
    general. Not the tests themselves but the foundation for developing
    them later, each step a general description that an agent other than
    the author can build a test from; unit tests are not listed. The rules
    for a step are in [The Testing section](#the-testing-section).
  - **Why:** what is delivered unchecked is part of the decision, and a
    reader has to be able to see it.

### Open questions

- **Open questions is what the author, often an agent, cannot settle from
  the code and the evidence, and a person has to explain or decide.**
  - **Answers:** "what is still undecided, and what is it between?"
  - **Holds:** the open ones, as bullets with the number in bold; then the
    subsection **Decisions** (below).
  - **Why:** an author who guesses writes an assumption into the design,
    where nobody looks for it. A question is the visible place where the
    record waits for a person.
  - **Not:** a question the code or the evidence answers, which is the
    author's work; an assumption made in silence in place of a question.
- **An open question holds what is asked, why it is still open and, where
  there are some, the choices it is between, and nothing of its answer.
  Once answered, it moves to the Decisions table with its answer, the
  choices it was made between and who decided it.**
  - **Why:** the reader who has to answer must see where the question ends;
    an answer, or the way to one, written into it reads as part of what is
    asked, or as if it were answered already.
  - **Example:** "Which code does a missing parent get? Its error reads like
    a missing pool's. Two ways are open: tell the two apart by the names in
    the message, or give both one code."
  - **Not:** the way to the answer run into the explanation, with nothing to
    mark where the question ends: "Which TLS versions does the appliance
    offer? No recording shows one, and its certificate has not been read.
    Answered by a handshake check and a read of the certificate, before the
    record is accepted."
- **Decisions is a table of the decided questions, with their numbers
  kept.**
  - **Answers:** "what was decided along the way, by whom, and where is it
    argued?"
  - **Holds:** one line of decision each, with the choices it was made
    between where there were some, who decided it, and a link to the
    section that argues it.
  - **Why:** the table keeps who decided what, so that a decision is
    neither taken for a fact nor reopened by accident. It indexes the
    decisions and does not restate their arguments.

### Relationship to other designs

- **Relationship to other designs is what stands next to the record: what
  it depends on, replaces or prepares, and what it leaves to a later
  record.**
  - **Answers:** "what else do I have to read, and where is what this
    leaves out?"
  - **Holds:** the designs, records and repositories this record touches,
    one line each with the link; then **Future work** subsections for what
    is deliberately left to a later record, each with what is already
    known about it.
  - **Why:** a record draws a border around one problem. Here a reader
    looking for something beyond the border is sent on, and what was
    learned on the way but belongs to the next record is kept.
  - **Not:** the design of the future work.

### Advice

- **Advice is the inputs the decision was formed from, each with its
  date.**
  - **Answers:** "what was this formed from, and when?"
  - **Holds:** findings, reviews, tracker entries and user decisions, one
    bullet each, the source and the date first.
  - **Why:** the body carries no history, so the history is here: a later
    reader can tell what was known when the decision was taken and what
    came after.

### Appendix

- **The Appendix holds what is looked up rather than read, and each thing in
  it is described there once; the rest of the record refers to it by name or
  by link.**
  - **Why:** a reader follows the argument without the catalogues in the
    way, and one who doubts a claim finds its entry in one known place.
  - **Not:** a finding described in Context and again in the Appendix; an
    argument for the decision that exists only in the Appendix.
- **Its parts come in this order, and a part the record does not need is
  left out.**
  - **Why:** the order goes from what the record rests on, through what was
    found, to what helps to read it, and is the same in every record.

  | Part | Holds | Form |
  |---|---|---|
  | Evidence | the sources the facts come from, how they were checked, and what the check cannot show | a numbered table, one row per source: Source, What it is, Use; then one sentence on which source wins when two disagree |
  | Evidence gaps | what no source shows | one `gap_` heading per gap with a field table; then what the attempts to close them covered, and how the rest would be closed |
  | Reference repositories | a repository the record measures against: what it holds, and the rules for using it | prose and numbered tables; only where the record depends on one |
  | What "compliant" means | the rules the code is checked against | one `rule_` heading per rule with a field table |
  | Compliance matrix | where the code stands on each rule | a numbered table with one row per checked unit, such as a call, and a second with one row per rule |
  | Finding catalogues | the findings, grouped by subject: one catalogue per kind of fault | one `###` section per catalogue, one finding heading per entry with a field table |
  | Dead code | what nothing uses and a commit would remove | a bullet list by kind: functions, types, keys |
  | Measurements | the counts and sizes the record quotes | numbered tables, each under a bold line saying what was measured and on which commit |
  | ID sources | where the names and every foreign numbering in the record come from | a numbered table: Names and IDs, Meaning, Source |
  | Terms | the jargon | a numbered table: Term, Meaning |

- **A catalogue opens with one or two sentences: what its entries have in
  common, and anything that holds for all of them.**
  - **Why:** what is true of the whole group is said once, not in every
    entry.
  - **Example:** "Error replies the code does not recognise. Each becomes an
    unknown error today."
  - **Not:** a statement in the opening that some entry contradicts.
- **Every named entry in the Appendix has a field table: two columns without
  a header, the field in the first and its value in the second, the fields in
  the order below.**
  - **Why:** a reader who has seen one entry knows where to look in the
    next, and a script can read the fields.
  - **Parts:**

  | Entry | Fields, in order |
  |---|---|
  | finding (`bug`, `crash`, `dataloss`, `security`) | **Where**: the file, line and function, or the call. **What happens**: the fault, in the present tense. **Effect**: what a user or a caller sees, where that is not plain from the fault. **Evidence**: a recording, a live run, a code reading, another document's finding. **Assigned to**: see below |
  | gap | **Kind**: a request sent differently from the evidence, or a situation never recorded. **What differs** or **Situation**. **Why it matters** or the code that depends on it. **Result**: what the attempts to close it showed, and the finding it became, if any |
  | rule | **Requirement**: what must hold. **Checked against**: the source or the tool that shows it |

  - **Example:** a catalogue may name the fault's parts in its own words,
    the same in every entry of that catalogue: "Struct says" and "Appliance
    sends" for a type mismatch, "Recorded reply" and "Plugin expects" for an
    unrecognised error, "Crash" and "Trigger" for a crash.
  - **Not:** a field left empty; a field present in some entries of a
    catalogue and absent in others without a reason.
- **Every finding has an "Assigned to" row: the implementation item, stage
  or commit of this record that takes it, as a link; or, when the record leaves it,
  the record, design or tracker entry that takes it; or "not assigned" with
  the reason.**
  - **Why:** a reader who lands on a finding sees who has it without
    searching, and the implementation item names the finding in return, so
    the two link each other. The row says who has the finding, not that it
    is fixed: a Proposed record has fixed nothing, and the row stays true
    when the work is done.
  - **Example:** "`impl_plugin_rest_error_catalogue`, stage 3"; "the
    `per_page` commit before stage 1"; "design 0003"; "not assigned: no caller reads the field".
  - **Not:** "Fixed by", which claims work that has not happened; a finding
    whose only pointer to its item is in the catalogue's opening sentence.

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
  - [Before_stage_1_<slug>](#before-stage-1-<slug>)
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

### Before stage 1: <slug>
[Before_stage_1_<slug>](#before-stage-1-<slug>)

<only for a commit that has to come before the stages: what it changes and
why it comes first>

### Stage 1: <slug>
[Stage_1_<slug>](#stage-1-<slug>)

#### impl_<area>_<slug>

**<Title>.** <the change, in the imperative mood; why; how; what must hold
once it is done>

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

**Table n — <Title>** · `tbl_<tag>`

| System | Effect of this document | Where |
|---|---|---|
| Not affected | ... | |

## Testing
[Testing](#testing)

1. **<corner case>** ([<finding, risk or question>](#<anchor>)). <What has
   to hold there, as something observed, and what counts as a result when
   the case cannot be brought about; where, only when the case can be seen
   in one place alone. One sentence for a simple case, the sequence and the
   timing for a scenario.>

## Open questions
[Open_questions](#open-questions)
  List of questions that are not clear to agent and require explicit explanation from human.
  Once question gets clear answer, or decision being made by human, the question with answer moves to `Decisions`

- **1. <question>?**

### Decisions
[Decisions](#decisions)

**Table n — <Title>** · `tbl_<tag>`

| # | Question | Decision | Decided by | Where it is argued |
|---|---|---|---|---|

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

| | |
|---|---|
| Kind | <a request sent differently, or a situation never recorded> |
| Situation | ... |
| Why it matters | ... |
| Result | ... |

### What "compliant" means
[What_compliant_means](#what-compliant-means)

#### rule_<area>_<slug>

| | |
|---|---|
| Requirement | ... |
| Checked against | ... |

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

### <catalogue by kind>
[<catalogue>](#<catalogue>)

<what the entries have in common>

#### <kind>_<area>_<slug>

| | |
|---|---|
| Where | <file:line, function or call> |
| What happens | ... |
| Effect | ... |
| Evidence | ... |
| Assigned to | <impl item, stage or commit; another record or tracker entry; "not assigned: <reason>"> |

### Measurements
[Measurements](#measurements)

### ID sources
[ID_sources](#id-sources)

### Terms
[Terms](#terms)
```
