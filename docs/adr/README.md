# Architecture decision records

This document describes architectural decisio record file format, it purpouse and the way it have to e treated and used, how one is written and maintained.
Document does not list the records, recrds themselves are the worked examples.

A record treats one problem, it might be what is wrong, the facts it rests on, what we care about, the options weighed, the decision and its reasons, what changes, what follows, and how it is checked.
Where a problem needs a full design, the record is that design, from the problem to the last implementation item.

Writing the record is half its value: it forces the reasoning into the open where it can be disagreed with.
The other half is history: a reader years later can see why the system is the way it is and how long each decision governed the work.

## Wording

### Bad habits

There is set of poor habbits that manifested themselves during writing this document and ADR's derived from it.

##### Long sentences held together by commas and semicolons
```
Every CSI call passes one shared step, the gRPC server's interceptor, and it only logs; the checks of the request, the reading of
    its IDs and the choice of the answer are written out again inside each call.
```

##### Nouns where verbs would do
"the reading of its IDs"
"the choice of the answer"
"the price that is accepted".
##### Abstract phrases standing for concrete facts
"the finding meets no step of the operational stage" means Kubernetes never sends this request
"a price that falls on operation" means every snapshot create gets slower

##### Definitions squeezed into the sentence
"the repair this document decides, each call of the iSCSI build fixed where it fails, …".
##### Terms of the record used as if the reader knew them
"the repair", "later work", "fault of operation", "the plugin's form".
##### Flat, formal phrasing
"a request after a slow publish can meet a connection reset"
"an ID of the call is empty in two ways".

### Examples

#### Open question on snapshot creation bug
Old version:
````
  - **28. Does the repair this document decides, each call of the iSCSI build
    fixed where it fails
    ([aopt_plugin_iscsi_repair_each_call_in_place](#aopt_plugin_iscsi_repair_each_call_in_place)), take
    [bug_plugin_ctrl_create_snapshot_name_taken_on_other_volume_answers_ok](#bug_plugin_ctrl_create_snapshot_name_taken_on_other_volume_answers_ok),
    or is that finding a third exception to compatibility with CSI v1.9.0?**
    [Question 2](#2-td_plugin_scope_every_finding_but_later_work) takes every finding, this one among them, but the
    finding meets no step of the operational stage: csi-snapshotter names each
    snapshot `snapshot-<UID>` of its VolumeSnapshot, whose source cannot
    change, so Kubernetes never sends one name with two source volumes; only
    csi-sanity's test of it (csi-test v5.6.0, `pkg/sanity/controller.go:1388`)
    or another CO does. Its fix has a price that falls on operation:
    JovianDSS keeps a snapshot's name within its volume, so CreateSnapshot
    would read every snapshot of the pool before each create
    ([risk 12](#risks)). The choices:
    - the repair takes it: CreateSnapshot answers AlreadyExists to a name
      taken on another volume, csi-sanity's test passes, and every snapshot
      create reads the snapshots of the pool, with risk 12, Testing step 22
      and the commit and about 50 lines of Table 6 that come with it;
    - a third exception, a fault left to later work as those of
      [questions 10](#10-td_plugin_rest_connections_kept_alive_reset_fix_later) and [20](#20-td_plugin_ctrl_publish_node_not_checked) are: CreateSnapshot goes on
      making a second snapshot and answering OK, csi-sanity's test fails, the
      finding goes to later work, and risk 12, Testing step 22 and that cost
      go.

````

Reworked version:
```
- **28. Should CreateSnapshot refuse a snapshot name that another volume
  already uses?** The finding is
  [bug_plugin_ctrl_create_snapshot_name_taken_on_other_volume_answers_ok](#bug_plugin_ctrl_create_snapshot_name_taken_on_other_volume_answers_ok).
  Today CreateSnapshot makes a second snapshot under the new volume and
  answers OK; the specification and csi-sanity (csi-test v5.6.0,
  `pkg/sanity/controller.go:1388`) expect AlreadyExists.
  [Question 2](#2-td_plugin_scope_every_finding_but_later_work) puts every finding in the repair, this one too. The
  question is open for two reasons:
  - Kubernetes never sends such a request. csi-snapshotter names each
    snapshot `snapshot-<UID>` of its VolumeSnapshot, and a VolumeSnapshot's
    source cannot change. Only csi-sanity or another CO sends one name with
    two volumes.
  - The fix costs every snapshot create. JovianDSS keeps a snapshot's name
    within its volume, so CreateSnapshot would have to read every snapshot
    of the pool before it makes one ([risk 12](#risks)).

  The choices:
  - fix it: CreateSnapshot answers AlreadyExists, csi-sanity's test
    passes, and every create pays for the read of the pool (risk 12,
    Testing step 22, one commit and about 50 lines in Table 6);
  - leave it as a third exception to compatibility with CSI v1.9.0, a
    fault left to later work like those of
    [questions 10](#10-td_plugin_rest_connections_kept_alive_reset_fix_later) and [20](#20-td_plugin_ctrl_publish_node_not_checked): csi-sanity's test
    ↪ keeps failing,
    and risk 12, Testing step 22 and that cost are dropped.
```

## Rules

Each rule is an entry in a `yaml` block. Its keys:

- `name`: the name by which the rule is referenced:
  `rule_adr_<chapter>_<slug>`.
- `rule`: the rule itself.
- `mapping`: a small table, right after `rule`: one key per row, and what
  it means. A row with more than one column is itself a mapping, with the
  column names as keys. Optional.
- `desc`: the description of the rule: why it exists, and when it applies
  and when it does not.
- `example`: example of correct usage of the rule. Optional.
- `wrong`: a bad example: what the rule rules out. Optional.
- any other key, such as `parts` or `areas`, for data that only that rule
  has.

A short value stands on the line of its key; a longer one, or one that YAML would not take as plain text, such as one that begins with a quotation mark, is a `|` block.
A mapping holds definitions; a rule stated inside a definition has no `desc` of its own.

### One problem, one record, one at a time

```yaml
- name: rule_adr_scope_one_problem_per_record
  rule: One problem per record. The problem is the unit, not the decision.
  desc: |
    a problem may need several decisions, and they make sense only
    together.
  example: |
    supporting a new API version needs the move to it, the reply types
    that match it and a client that is correct against it; none solves
    the problem alone, so they are one record: its Options weigh the
    move, and its questions decide the reply types and, as the evidence
    dictates, the client.
  wrong: a record per decision, each waiting on the others.

- name: rule_adr_scope_next_record_after_previous_is_implemented
  rule: |
    Records are not piled up: a record is implemented after it is
    accepted, and the next record, about the next problem, is written
    after that.
  desc: |
    a record describes the current code, so one written before the
    previous record is implemented describes code that is about to change.
  wrong: |
    several proposed records waiting on each other, which is a sign that
    the problem was cut in the wrong place.

- name: rule_adr_scope_editorial_choice_gets_no_record
  rule: |
    An editorial choice gets no record; it is one line in a record's
    Decisions table.
  desc: |
    a record is for a problem, and a choice of tense or of where release
    notes live has no problem behind it.
  example: |
    where the changelog lives, decided as a numbered question of the
    record that first needed it.
  wrong: a record whose problem statement is a preference.

- name: rule_adr_scope_limits_stated_in_options_and_decision
  rule: |
    Problem statement defines the challange.
    Problem statement is concise and minimalistic, it might include scope, time limitation complexity limitation if that is critical aspect of problem.
    Options considered section weights strategies that supposed to address given Problem.
    Depending on complexity of a problem, it might and will limit scope.
    It might list directions or aproches that will not be taken and motivation behind it.
    Decision section formalize and confirms one of considered options.
  desc: |
    Problem statement aims to be as minimalistic as possible.
    Problem that is defined by problem statement might be addressed in multiple ways.
    This aproaches may vary on technology, time, affected components and etc...
    Overcrouding problem statement will confuse reader and agent.
    To make ADR clear and streightforward the scope aspects have to be layed down in Options Considered unless author explicetly states that feature or bug have to be addressed with specific time limit or resources.
    Some problem might touch multiple components and implementation might stretch to significant amount of time.
    Considering such cases is making decision on the best aproach to pick is a part of Option considered and Decision sections.
  example: |
        ## Problem statement
        [Problem_statement](#problem-statement)
        
        The plugin uses JovianDSS REST API v3, which is deprecated, through a client
        that follows documentation and assumptions rather than what the appliance
        serves, and it pages its CSI lists in a way that crashes the controller or
        returns entries twice or never. It has to use API v4 through a client that
        complies with the API as served, down to the way the appliance pages its
        lists.
        
        ...
        
        ## Options considered
        [Options_considered](#options-considered)
        ...                
        How far it reaches: the move itself changes the paths of `pkg/rest`, the
        documentation and the changelog; the driver, the controller and the node
        stay as they are ([Stage 1](#stage-1-move-to-api-v4)).
        ...
         
  wrong: |
        ## Problem statement
        [Problem_statement](#problem-statement)
        The plugin uses JovianDSS REST API v3, which is deprecated, through a client
        that follows documentation and assumptions rather than what the appliance
        serves, and it pages its CSI lists in a way that crashes the controller or
        returns entries twice or never. It has to use API v4 through a client that
        complies with the API as served, down to the way the appliance pages its
        lists. JovianDSS 2 will not be supported within this iteration.

    Redundant part: `JovianDSS 2 will not be supported within this iteration.`
```

### File name and status

```yaml
- name: rule_adr_record_file_named_by_number_and_problem
  rule: |
    A record is named `NNNN-<problem-slug>.md`: a sequence number for
    order, then the problem it answers. Its title line is
    `# NNNN — <problem>`.
  desc: |
    the name says the problem, not the decision, so a proposed record may
    change its decision without changing its name, and a later record
    that supersedes it about the same problem sorts next to it.
  example: |
    a record that answers "support API vN on product line X" is
    `NNNN-api-vN-support-x.md`; its decision, whatever it becomes, is the
    bold paragraph of its Decision section.
  wrong: a name that says the decision, wrong the day the decision changes.

- name: rule_adr_record_number_never_reused
  rule: |
    Numbers are never reused. The sequence of this folder starts at 0002;
    0001 is not used, since design 0001 in `docs/design/` is a
    predecessor of these records.
  desc: |
    a number cited anywhere must keep pointing at the same record, and a
    superseded record keeps its text and its number. Record 0002 keeps
    the number of design 0002, which it reworks.

- name: rule_adr_record_status_one_word_from_mapping
  rule: The status is one word from this mapping, with a link when Superseded.
  mapping:
    Proposed: under discussion; the text may still change
    Accepted: signed off by the approver; from then on the text is frozen
    Superseded: |
      replaced by a later record, which the status line links; the text
      stays as it was
    Retired: |
      the decision no longer applies and nothing replaces it; the text stays
  desc: |
    a reader must know whether the text can still change and what to read
    instead.

- name: rule_adr_record_decision_and_acceptance_dates_in_metadata
  rule: |
    The date a decision was taken and the date the record was accepted
    are different things, and both go in Metadata.
  desc: |
    the decision governs the work from its own date; acceptance freezes
    the text.
```

### How the text is organised

```yaml
- name: rule_adr_text_claim_first_then_explanation_then_data
  rule: |
    A block that makes a claim starts with it, as one bold sentence or as
    the heading; then the explanation, in a few plain sentences; then the
    data and the references, as a short table or a labelled line.
  desc: |
    a reader who trusts the claim stops after the first sentence; a
    reader who doubts it finds the evidence in the same place every time.
  example: |
    "**The hand-written reply types fail on every version in the same
    way.**", then the sentences that explain it, then a last line that
    links the catalogue with the full detail.
  wrong: a table first, and the claim it supports after it.

- name: rule_adr_text_each_fact_once
  rule: |
    Each fact once, at one place; when two sections start to say the same
    thing, one of them links to the other.
  desc: a fact stated twice drifts into two facts.
  example: |
    Options carry the alternatives and their price; What changes carries
    the mechanism; the Appendix carries the catalogue; Context carries
    the claims and links to the catalogue; the Problem statement says the
    problem and leaves the evidence to Context; the Decisions table
    indexes the decisions and does not restate their arguments.

- name: rule_adr_text_plain_language_jargon_in_terms
  rule: Plain language; jargon is explained once, in Terms.
  desc: a reader from outside the project must be able to follow the argument.

- name: rule_adr_text_no_history_no_dates_in_body
  rule: No history and no dates in the body.
  desc: |
    the body describes the current code and what the record changes;
    dates belong to Metadata and Advice, where a reader looks for them.
  wrong: |
    "on the day the evidence came in we decided" in Context.

- name: rule_adr_text_mood_imperative_conditional_present
  rule: |
    The mood rule: the record describes the changes it makes in the
    imperative mood, as a commit message of the Linux kernel does; what
    would follow from them, as in Consequences and Risks, is in the
    conditional; facts about the current code and the evidence are in the
    present tense.
  desc: |
    the kernel's guide to submitting patches asks to "describe your
    changes in imperative mood, as if you are giving orders to the
    codebase to change its behaviour". An order reads the same before and
    after the work is merged, so the text, frozen at acceptance, stays
    true; the present tense of the facts is about the commit named in
    Metadata.
  example: |
    "make `getResourcesList` read the list once on every call"; "send
    `per_page=100` on every list request".
  wrong: |
    "`getResourcesList` would read the list once", "this commit makes
    `getResourcesList` read the list once" or "I changed
    `getResourcesList` to read the list once"; nor an order rewritten as
    "does" once the work is merged.

- name: rule_adr_text_tense_present_past_no_future
  rule: |
    The tense rule: what is, in the present; what happened, in the past;
    nothing in the future. A change, and what would follow from it, take
    the mood rule's imperative and conditional.
  mapping:
    "the current code, the evidence and what it shows":
      tense: present
      example: |
        "the driver cuts the list at `max_entries` before the controller
        skips anything"
    "a decision as it stands":
      tense: present
      example: |
        "the plugin takes TLS 1.2 or higher"
    "a condition that has to hold: What must hold, a Testing step, a rule":
      tense: present
      example: |
        "a GET answered 204 is sent again after `retrydelay`"
    "what happened at one time: a run, a recording, a check, a review":
      tense: past
      example: |
        "129 characters were recorded accepted, 150 refused"
  desc: |
    the text is read before and after the work is merged. The present
    stays true of the commit named in Metadata and of the conditions the
    work has to meet, and the past of a run or a check once it has
    happened; "will" is true only until the work is merged.
  wrong: |
    "the plugin will send `per_page=100`"; "the driver cut the list at
    `max_entries`" for today's code; "a GET answered 204 was sent again"
    in What must hold.

- name: rule_adr_text_record_calls_itself_this_document
  rule: |
    The record calls itself "this document", never "this design" or "this
    record".
  desc: |
    the text reads the same whether the reader came for the design or for
    the decision.

- name: rule_adr_text_reference_forms_for_code_sections_and_findings
  rule: |
    References: code as `file:line` of the commit named in Metadata, as a
    relative link where a link helps; sections as links to their anchor;
    the findings of another repository by their number there.
  desc: |
    a link can be checked and is followed by a click; another
    repository's findings keep the numbers that repository gave them.

- name: rule_adr_text_no_test_environment_details_no_secrets
  rule: No test-environment details and no secrets in a record.
  desc: |
    machine names, pool names and addresses change with the lab and
    belong to the testing repository, which holds them; a secret in a
    document is a leak.
  example: |
    "the lab appliance", not its address.
  wrong: a pool name in a Testing step.
```

### Names, not IDs

```yaml
- name: rule_adr_names_entries_named_not_numbered
  rule: |
    Findings, gaps, compliance rules, architectural options and
    implementation items are named `<kind>_<area>_<slug>`, not numbered.
  desc: |
    a human reader has to keep track of what is referenced. A name says
    what the thing is; a letter and a number say nothing, are hard to
    remember and get confused with each other. A name never changes; a
    number is a position in a list that grows.
  example: |
    `gap_jdss_rest_empty_body_on_get_delete`,
    `bug_rest_v4_pools_pool_get_iostats_type_mismatch`,
    `rule_rest_reply_types_decode`,
    `impl_plugin_driver_validate_names_and_ids`.
  wrong: |
    the letter-and-number IDs of the reviews, F8, C3, G5, A11, which are
    mapped to names (below); an option called `Option A1` or `Option B`.

- name: rule_adr_names_kind_comes_first
  rule: |
    The kind comes first and says what is named: a finding, a gap, a
    rule, an architectural option or an implementation item.
  mapping:
    bug:
      names: behaviour of the code that is wrong
      heading: the Appendix catalogue of its subject
    crash:
      names: a panic the code can reach
      heading: Appendix, Crashes
    dataloss:
      names: a wrong result that loses or corrupts data
      heading: Appendix
    security:
      names: a finding of the security review
      heading: Appendix, Security findings
    gap:
      names: |
        evidence that is missing: a case no recording shows
      heading: Appendix, Evidence gaps
    rule:
      names: a compliance rule the code is checked against
      heading: Appendix, What "compliant" means
    aopt:
      names: |
        an architectural option: a way the problem could be solved
      heading: Options considered
    impl:
      names: |
        an implementation item: a change the record delivers, which
        implements the decision
      heading: What changes, under its stage or its commit
  desc: |
    a reader knows from the first word whether the entry describes what
    is or prescribes what is done.
  definitions: |
    `bug`, `crash`, `dataloss` and `security` are findings: they describe
    what is, and their field table says where, what happens, the
    evidence, and which implementation item, stage or commit they are
    assigned to ([Appendix](#appendix)). A gap says what no source shows,
    and its field table why that matters and what the attempts to close
    it showed. A rule says what must hold. An architectural option says
    how the problem could be solved. An implementation item says what is
    done, whatever its form: a repair, a new file, a tool, a setting.

- name: rule_adr_names_area_second_and_character_set
  rule: |
    Then the area. The whole name is lower-case letters, digits and
    underscores only.
  desc: |
    the area tells a reader which system the name belongs to before the
    slug is read, and the character set makes the name its own anchor.
  areas: |
    `jdss_rest` (what the appliance's REST API does), `plugin_rest`
    (`pkg/rest`), `rest` (the API as the contract between the two: the
    compliance rules, and what fails to comply with the spec),
    `plugin_driver`, `plugin_ctrl`, `plugin_node`, `plugin_cli`, `plugin`
    (several packages), `lab`, `spec`, `docs`, `adr` (the rules of this
    file), `build`. A new concern adds a new area word; nothing existing
    changes.

- name: rule_adr_names_slug_object_first_and_precise
  rule: |
    Then the slug, given with care: it catches the essence of the finding
    or the item, object first, precisely enough that no later item can be
    confused with it, and is no longer than that takes.
  desc: |
    a name is read by a human who must know what it refers to without
    looking it up, and it is never changed, so a vague name blocks every
    later name near it.
  example: |
    for a REST call, the API version, the path with its placeholders as
    words, the method, then what is wrong:
    `bug_rest_v4_pools_pool_nas_volumes_dataset_get_fail_to_comply_with_api_spec`.
    For an object, the object first: `volume_rollback`.
  wrong: |
    `bug_plugin_rest_nas_volume_keys_never_match`, a symptom in the code
    that could belong to several endpoints and versions;
    `rollback_volume`, the object not first.

- name: rule_adr_names_item_and_finding_name_each_other
  rule: |
    An implementation item names the findings it closes and the rules it
    meets, and a finding names the item, stage or commit it is assigned
    to.
  desc: a reader goes from a finding to its fix and back.

- name: rule_adr_names_named_entry_has_heading
  rule: |
    Each thing the record describes under a name has a `####` heading,
    the reference point that the rest of the record and other documents
    link to: in the Appendix for findings, gaps and rules, each with a
    field table under it; in Options considered for architectural
    options and in What changes for implementation items, both prose.
  desc: |
    the heading is its own anchor and a single word, so `*` in vim and a
    link on GitHub both reach it; such headings need no self-anchor line
    and are not listed in the Table of Contents.

- name: rule_adr_names_name_is_link_or_code_font
  rule: |
    A name is a link, `[name](#name)`, wherever it is mentioned; a name
    that has no heading in the record is written in code font without a
    link.
  desc: |
    a link is checked, every linked name has a heading, and code font
    marks a name that is described only in a review or the tracker.
  wrong: a bare name in prose.

- name: rule_adr_names_no_ranges_of_names
  rule: |
    Names have no order: list them, or name the group and link the
    section that holds it.
  desc: a range of names tells a reader nothing.
  example: |
    "the stage 4 items", "the first eight gaps under Evidence gaps".
  wrong: |
    "`bug_a` to `bug_b`".

- name: rule_adr_names_registry_maps_old_ids_and_registers_names
  rule: |
    The older letter-and-number IDs live on in the reviews and in the
    tracker; the map between them and the names is the registry of names,
    which is also where a new name is registered.
  desc: |
    the reviews are not rewritten, and one registry keeps a name unique
    across documents.

- name: rule_adr_names_name_never_changed_once_referenced
  rule: |
    A name is chosen once and never changed after it is referenced; a
    changed understanding gets a new name, and the old one is marked
    superseded.
  desc: a name is cited from documents that are not rewritten.
  wrong: a rename to a better slug.
  exception: |
    the kind was `work` until 2026-11-11 and became `impl` before any
    document outside the record used the names; on 2026-11-12, when the
    slug rule above was expanded, 38 names were given new names under it,
    each with its old row marked superseded in the registry.

- name: rule_adr_names_numbers_that_identify_are_never_renumbered
  rule: |
    Numbers stay where they identify an entry: risks, questions, testing
    steps, stages, key observations. They are never renumbered once
    referenced; new entries are appended. Numbers that rank or place,
    tables and priorities, follow the position (see
    [Navigation](#navigation)): a priority added later takes its place in
    the order, the ones below it move down, and every "priority n"
    reference is rewritten in the same change.
  desc: |
    a reference such as "risk 9" or "question 27" keeps its meaning;
    "priority 2" means the second weight, which is the thing that
    changes.
  wrong: |
    renumbering to fill a missing number; a priority appended as number
    eight that weighs second.

- name: rule_adr_names_other_documents_numbering_cited_as_theirs
  rule: Another document's numbering is cited as that document's.
  desc: its numbers are its own and change with it.
  example: |
    a test suite's findings, a design's items, the testing repository's
    rows.

- name: rule_adr_names_list_with_missing_numbers_as_bullets
  rule: |
    A numbered list with missing numbers is written as a bullet list with
    the number inside the bold, `- **30. <question>?**`.
  desc: |
    a renderer takes an ordered list's numbers from its first item and
    renumbers the rest, so the numbers a reader saw would not be the ones
    written.
  example: |
    the open questions, once some are decided and moved to the Decisions
    table.
  wrong: |
    `30.` as an ordered-list item, which renders with the number of its
    position: 24 for the third item of a list that starts at 22.
```

### Navigation

```yaml
- name: rule_adr_navigation_heading_followed_by_anchor_line
  rule: |
    Every `##` and `###` heading is followed by a line that links to that
    heading, with the heading's words joined by underscores, and the
    Table of Contents uses the same tokens.
  desc: |
    `*` in vim on the token jumps between the Table of Contents entry and
    its section; on GitHub the same token is a link.
  example: |
    `[Open_questions](#open-questions)`.

- name: rule_adr_navigation_every_table_numbered_except_field_tables
  rule: |
    Every table but a field table carries a number, whether the text
    refers to it or not.
  desc: |
    the List of Tables then shows every table, and citing a table later
    renumbers nothing; a field table is found by its heading.

- name: rule_adr_navigation_table_caption_and_numbering_by_appearance
  rule: |
    A table that carries a number has a caption
    `` **Table n — Title** · `tbl_tag` `` and appears in the List of
    Tables; numbering follows order of appearance, so a table inserted or
    removed renumbers those after it, and every "Table n" reference is
    updated in the same change.
  desc: |
    a reader finds a table by its number in the List of Tables.
  wrong: |
    a catalogue that became headings keeping its table number; it is no
    longer a table.

- name: rule_adr_navigation_heading_renamed_only_with_its_links
  rule: |
    A heading that something links to is renamed or moved only together
    with every link to it, in the same change.
  desc: a broken anchor fails silently on GitHub.
  wrong: a rename whose links are fixed in a later commit.
```

## Maintaining a record

The rules below are entries of the form described under [Rules](#rules).
They follow the life of a record: it is changed while Proposed, checked
before a review, and frozen once Accepted.

```yaml
- name: rule_adr_maintenance_atomic_change_across_entire_document
  rule: |
    While a record is Proposed, its text changes freely. A change is made
    in every place of the record it touches, in the same change. The
    mapping lists what the common changes touch.
  mapping:
    a new finding: |
      a name, a heading in the Appendix, a row in the registry, and links
      from where it matters
    a new question: |
      the next number; when decided, it moves to the Decisions table with
      its number kept
    a new risk: the next number, at the end of the list
    a Testing step folded into another, or moved to the testing repository: |
      its number stays, with one line saying where it went
  desc: |
    a reviewer reads the whole text, so a change must not leave it broken
    anywhere.

- name: rule_adr_maintenance_checks_pass_before_review
  rule: |
    Before a record is given to a reviewer, its author checks it against
    the list below and corrects everything that fails.
  desc: |
    the reviewer's time goes to the content, not to what the checks find.
  not_checked: |
    the lines of a fenced code block, which are code and not text of the
    record.
  checks:
    links: |
      - every `](#anchor)` resolves to a heading;
      - every `##` and `###` heading has its anchor line under it, such as
        `[Decisions](#decisions)`;
      - every entry in the Table of Contents and the List of Tables
        matches a heading or a caption;
      - every "Table n" reference is within range;
      - every relative link points to a file that exists;
      - every Testing step names the finding, risk or question it comes
        from;
      - no link points to the section it sits in. Exception: a section
        that numbers or catalogues its entries, where an entry is cited
        through the section's own anchor: Key observations, What changes,
        Risks, Testing, Open questions, Decisions, the Appendix and its
        Evidence gaps.
    numbering: |
      - risks, steps, stages, key observations, priorities and tables are
        numbered without gaps or duplicates;
      - no question number appears twice, within or across the open list
        and the Decisions table;
      - the open questions are bullets with the number in bold;
      - a question keeps its number when it moves to the Decisions table.
    names: |
      - every name is unique, and every linked name has a heading;
      - every name is registered in the registry;
      - the `impl_` headings of the record are the `impl_` names in the
        registry's list of implementation items;
      - no name that the registry marks superseded is written in the
        record;
      - no letter-and-number ID remains outside the ID sources table and
        the citations of other documents' numbering.
    wording: |
      - none of these phrases: "design NNNN" or "ADR NNNN" with the
        record's own number, "this design", "this record", "this ADR",
        "the design" and "the design's" in upper or lower case (another
        design is cited as "design NNNN"), "earlier text", "earlier list"
        and "as listed above";
      - no sentence longer than sixty characters appears twice outside
        tables;
      - no horizontal rule.

- name: rule_adr_maintenance_accepted_record_is_frozen
  rule: |
    Once a record is Accepted, its text is not edited. A change of
    decision is a new record that supersedes this one, and the status
    line of this one links it.
  desc: the frozen text is the history half of a record's value.
  wrong: an edit to an accepted record to improve its wording.
```

## Sections

A record has the sections below, in this order; the [Skeleton](#skeleton)
shows them empty. Each is described the same way: the sentence in bold says
what the section is; **Answers**, the question of the reader it answers;
then **Why**, **Example** and **Not**, and for Metadata **Parts**, the
lines it is made of. The rules about what goes into a section follow its
description, each as an entry of its own.

- **A record is one argument: the sections answer the reader's questions in
  the order they are asked, each resting on those before it and on the
  Appendix.**
  - **Why:** a reader who disagrees can find the step they disagree with, a
    fact, the order of the priorities or the price of an option, and see
    how the conclusion would shift. The discussion is then about that step
    and not about the whole record.
  - **Not:** a section filled in because the skeleton has it, which answers
    no question of the reader.
- **A `##` section with nothing to say is kept, with one sentence saying
  so.**
  - **Why:** a reader then knows the subject was considered.
  - **Not:** a `##` heading dropped because it was empty, which reads as an
    oversight.

### Table of Contents

- **The Table of Contents lists the `##` and `###` headings of the record
  in order, as underscore tokens.**
  - **Answers:** "how is this laid out, and where is the part I need?"
  - **Why:** a record is consulted as often as it is read through, and a
    token is one word, so `*` in vim reaches its section
    ([Navigation](#navigation)).

### List of Tables

- **The List of Tables lists the numbered tables with their `tbl_*` tags,
  in order of appearance.**
  - **Answers:** "which tables are there, and where?"
  - **Why:** the tables carry much of the data, and a reader looking for
    one finds it here without scrolling ([Navigation](#navigation)).

### Metadata

- **Metadata says when the decision was taken, who stands behind it and
  which commit its facts are true of.**
  - **Answers:** "what am I holding, and as of when?"
  - **Why:** every "today" and every line number in the body refers to one
    commit, and the decision was taken on one date; a reader has to know
    both before relying on either.
  - **Parts:**
    - `Date:` when the decision of the Decision section was taken and when
      the text was written, the question decisions being dated in Advice;
    - `Driver:`;
    - `Approver:`;
    - `Code:` the commit the line numbers refer to, called "the current
      code";
    - `Implementation:` where the work is described and how it is
      delivered;
    - `Accepted:` the date of acceptance, once accepted;
    - `Formerly:` the design the record absorbed, and when; only where it
      absorbed one.

### Status

- **Status is one word from the mapping in
  [File name and status](#file-name-and-status), with a link when
  Superseded.**
  - **Answers:** "can this text still change, and does its decision still
    govern?"
  - **Why:** it has a section of its own so that it is found at the same
    place in every record and changes in one line.

### Problem statement

**Problem statement** defines
  - the problem or issue that the record addresses
  - goal that needs to be reached by the record
  - new feature or improvement that is introduced by the record

It must be succinct and clear on topic.

**Answers**
  - what is wrong
  - what is missing
  - what change in the system's behaviour is wanted
  - what desired operation / capability looks like

**Why**

Core section that defines all following sections.
Every later section is made to answer challenge defined in it.
AI agent and human reader must be able to hold the whole of it in mind and
have clear understanding of the matter that particular ADR addresses.

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
  - list of answers to all questions provided above in form of bullet or a
    bold sentence per question.
    That is wrong because it blurs definitions and context becomes unclear.
  - list of the tests that will prove success.
    ADR is built around features and problems, not passing specific tests.
  - list of the issues the record leaves out.
    ADR's problem statement describes the things that ADR is focused on,
    bringing things that are not touched in the given ADR will only confuse
    user or agent.
    If there is a need to narrow down the scope, it should be done in
    `Options considered` and `Decision` sections.
  - the Context repeated with its evidence.
    Repeating same things over is a bad practice and should be avoided.
  - the instances of the problem, such as the calls that fail, where the
    problem itself is meant.

### Context

- **The Context section holds the facts the decision rests on, in
  subsections by subject: what exists today, what the evidence shows and
  what it does not show.**
  - **Answers:** "what do we know?"
  - **Why:** a decision is as good as the facts under it, so they are told
    in a way that a reader who does not live in this code can follow the
    rest of the record from there, and what is not known is said as
    plainly as what is: a gap passed over in silence comes back as a
    surprise in the lab.
  - **Not:** a tour of the code; a fact no later section uses; a finding
    described here and again in the Appendix.
- **The Context section ends with the subsection Key observations: the few
  conclusions the decision turns on, numbered.**
  - **Why:** Key observations is where the facts become the conclusions the
    options are argued from; a reader who accepts them should find the
    decision unsurprising.
  - **Not:** a key observation that no subsection supports.

### Priorities

- **The Priorities section lists the project values that weigh on this
  record, in the order they weigh here, each with a sentence where it needs
  one. The values are the project's and stable; the priorities are this
  record's order among them, and the options cite them by number; the
  number is the rank ([Names, not IDs](#names-not-ids)).**
  - **Answers:** "what do we care about here, and in which order?"
  - **Why:** an option is not good or bad in itself, only against what is
    valued. With the order stated before any option is named, the choice
    follows from it, and a reader who orders the values differently sees
    which choice would change.
  - **Not:** priorities written after the decision, to fit it; a priority
    that no option, decision or risk cites.

### Options considered

The Options considered section list aproaches to address problem statement on the architectural and strategic levels.
Options expected to be mutualy exclusive.
If they are not, it most likely the scope question.
It should be clear what aspects are mutualy exclusive.
For scope related issues assesment of time and number of changes that will be required to acomplish specific part or component.

The options are mutualy exclusive.
Within single option assesment have to be given along side with scoped components and systems.

Each option shows its context and strategy to address problem statement.

Trade-offs stated honestly against the priorities.
Options under its name, `aopt_<area>_<slug>`, as a heading ([Names, not IDs](#names-not-ids)).
Where only one option was considered, its text says so and explains why, with the price that option carries.
Why: an option told fairly enough that someone who prefers it would recognise it lets a later reader, whose priorities have changed, reopen the choice without doing the work again.
Bad practive is to make an option set up to lose.
The one an alternative nobody weighed, written up to fill the list.
Having option without a disadvantage is a bad sing as well

Each option says what distinguishes it and the limits of its scope: which modules and systems it changes and which it leaves alone.
It does not go in detail on what exactly being changed or implemented, rather brushed general direction and highlites what differs it from the other options.
([One problem, one record, one at a time](#one-problem-one-record-one-at-a-time)).

### Decision

- **The Decision section holds the decision, in a few bold sentences, and
  the reasons for it.**
  - **Answers:** "what was decided, and why this?"
  - **Why:** this is what other documents cite and what a reader years
    later looks for.
  - **Not:** the mechanism, which belongs in What changes.
- **The decision is in sentences that can be quoted alone. It names the
  option adopted in each group and states the limits of its scope as
  finalized: what it changes and what it leaves alone.**
  - **Why:** a sentence cited in another document is read without the text
    around it; which option is adopted is said here and not in Options
    considered; the limits of the scope are known once the options are
    weighed
    ([One problem, one record, one at a time](#one-problem-one-record-one-at-a-time)),
    so they are finalized here.
- **The rationale names each priority, with the price that is accepted and
  what stays without a guard. For a choice with contenders, one sentence
  then tells a reader with another order of priorities which option they
  would pick.**
  - **Why:** a reason that names its priority can be checked against the
    options, and a price that is stated was accepted, not overlooked.
  - **Not:** a rationale of benefits only.
- **Where question decisions shape the delivery, a paragraph names them
  with their question numbers and links the Decisions table.**
  - **Why:** a reader of this section alone would otherwise miss decisions
    that shape the delivery.

### What changes

- **The What changes section describes the work: the commits in order,
  grouped in stages, and each implementation item under its name.**
  - **Answers:** "what exactly is done, and in which order?"
  - **Why:** a decision governs only through the work, so the work is
    written precisely enough that someone who took no part in the
    discussion can do it from here, and a reviewer can tell when an item
    is done.
  - **Not:** the argument for the decision told again.
- **A commit that has to come before the first stage stands under a
  heading of its own, before the stages, with the reason it comes first.**
  - **Why:** an order of commits whose reason is not given gets changed.
- **Each implementation item is a named heading with what changes, why,
  how, and what must hold once it is done, the changes in the imperative
  mood ([How the text is organised](#how-the-text-is-organised)). The unit
  tests that show an item works are the implementing agent's, not the
  record's ([Testing](#testing)).**
  - **Why:** an item without what must hold cannot be checked.
- **The mechanism lives here and nowhere else, except the code of a
  critical algorithm, which is written out in the Appendix under Code
  segments and linked from its item.**
  - **Why:** each fact has one place
    ([How the text is organised](#how-the-text-is-organised)), and code is
    looked up rather than read, so its place is the
    [Appendix](#appendix).
  - **Not:** code written out where its shape is enough.

### Consequences

- **The Consequences section says what becomes true once the record is
  applied, positive and negative.**
  - **Answers:** "what is true once this is applied, and how sure are we?"
  - **Why:** a decision is accepted for what follows from it, so that is
    said in one place and not left to be worked out from the items. The
    section comes after What changes because it refers to the stages and
    items there.
  - **Not:** advantages only; a failure that may happen, which is a risk.
- **The section ends with the confidence in what follows from the work and
  the condition that should trigger a re-evaluation.**
  - **Why:** the confidence and the condition for re-evaluation tell a
    later reader when the record stops being a safe guide.

### Risks

- **The Risks section says what can go wrong, for whom, and what stops it:
  backward compatibility first, preserved and not preserved; then the
  numbered risks.**
  - **Answers:** "what can go wrong, and what do I lose on upgrade?"
  - **Not:** a risk left out or softened so that the record looks safe.
- **Each risk has the concrete failure, as something that would be seen,
  and its guard, or the plain statement that it has none.**
  - **Why:** a failure described concretely can be tested and watched for;
    a risk without a guard, said so, is a price accepted knowingly.
  - **Not:** "may cause problems".

### Affected systems and files

- **The Affected systems and files section shows the reach of the work: a
  table by system first, then a table of the files, one row each with the
  stage that changes it.**
  - **Answers:** "what does this touch?"
  - **Why:** a reviewer sees the reach at a glance; whoever keeps another
    system, the lab, the documentation or another repository, learns that
    it concerns them.
- **The table by system ends with a row for what is not affected.**
  - **Why:** that row ends a search before it starts.

### Testing

```yaml
- section: |
    The Testing section lists the corner cases that have to be tested, as
    numbered steps from the cheapest to the live one, and at the end what
    the tests cannot reach.
  answers: |
    "which corner cases have to be tested, and what potential negative scenario can follow?"
  desc: |
    what is delivered unchecked is part of the decision, and a reader has
    to be able to see it.

- name: rule_adr_testing_corner_cases_not_tests
  rule: |
    The Testing section describes the corner cases that have to be
    tested, such as the conditions the implementation must meet that only
    the evidence or the lab know. It is not a description of the tests
    themselves but the foundation for developing them later.
  desc: |
    a record is written before the code, so this section is where the
    implementation meets what the designer could not see from the design.
  example: |
    a reply shape, a tool that refuses an image, a log that leaks secrets
    at Debug.

- name: rule_adr_testing_unit_tests_not_listed
  rule: |
    Unit tests are not listed: the agent implementing an item writes them
    as part of the item, and the item says only what must hold once it is
    done.
  desc: |
    the record does not have the full, exact picture of what will be
    implemented and how, so it cannot guide how unit tests are written;
    the agent implementing the item has that picture.
  example: |
    "a reply over the limit fails with an error naming the limit".
  wrong: |
    a count of unit tests, a fake server described, a list of unit-test
    cases.

- name: rule_adr_testing_step_names_case_condition_and_source
  rule: |
    Each step names one corner case, what has to hold there, and the
    finding, risk or question it comes from. It says where only when the
    case can be seen in one place alone. An agent other than the author
    has to be able to build a test from it.
  desc: |
    a step is what a test is built from later. Who runs the test, what
    has to be in place first and its commands are known when the test is
    written, not when the record is, so they belong to the test.
  parts: |
    - the corner case: the edge or the condition, with a link to the
      finding, risk or question it comes from;
    - what has to hold there, as something someone can observe: a status, a
      reply, a log line, what is left on the appliance; and what counts as
      a result when the case cannot be brought about, so that an unknown
      outcome becomes a result instead of being passed over;
    - where, only when the case can be seen in one place alone, such as the
      lab appliance.
  example: |
    "**A reply larger than `maxreplysize`**
    (`impl_plugin_rest_reply_size_limit`). The call fails with an error
    that names the limit, and no more of the reply is read."
  wrong: |
    "**Reply size test** (the lab machine, run by the user). Build the
    plugin, set `maxreplysize: 1000`, run `dev_cycle.py all` and look for
    the error in `plugin.log`": who runs it, the setting and the commands
    belong to the test written from the step.

- name: rule_adr_testing_step_detail_follows_the_case
  rule: How much a step spells out follows the nature of the case.
  desc: |
    the parts above are what a reader needs, in as many words as the case
    takes, not a form to fill in.
  example: |
    a simple case takes one sentence: a volume deleted twice, whose second
    delete is handled as "already gone" and not as an error. A scenario
    needs its sequence, its timing and what to watch: an appliance shut
    off by force while create and delete requests are in flight, how the
    cluster behaves under that load, and what each request is answered
    while the pool moves to the other node.

- name: rule_adr_testing_step_order_and_what_tests_cannot_reach
  rule: |
    The steps go from the cheapest to the live one, and the section ends
    with what the tests cannot reach.
  desc: a reader knows which findings have no cover in this delivery.

- name: rule_adr_testing_step_states_implementation_detail_it_depends_on
  rule: |
    A step that depends on an implementation detail says so, and that
    detail is written into the item that produces it.
  desc: |
    otherwise the implementer may remove, without knowing it, what the
    test built from the step reads.
  example: |
    a log line keeping its query string, an error text staying
    searchable.

- name: rule_adr_testing_results_kept_out_of_record
  rule: |
    Results go to the testing repository with the rest of the lab's data,
    not into the record.
  desc: |
    the record says what must hold; the lab's data changes with every run
    and carries lab details.
```

### Tactical concerns

The Tactical concernts section holds what the author, often an agent, cannot settle from the code and the evidence,
And require explain or decide from the human.

It stores open question, ambiguities that have to be address in sub section `Open questions`.
And questions with answers in section `Tactical decisions`.

It provides common place for agents and humans to record various question regarding ADR.
So that they do not overlap, do not get scatered around code base and it is easy to find and trac decisions regarding each individual problem.
This aproach is expected to help agents avoid question duplication and improve comprehention of ADR.
If question is trivial and can be clearly derived from options considered and problem statement, it should not get into a tactical concerns, unless it is added by user.
Agents are alowed to write answers to questions if user asks them to do it.

An open question holds what is asked, why it is still open and, where there are some, the choices it is between, and nothing of its answer.
  Once answered, it moves to the `Tactical decisions` section with its answer, the choices the answer was made between and who decided it.

The one answering question must clearly see what questions are undecided and require attention.
The one reviewing ADR and analising internal logic of decision making, have to see what was asked, and what was the decision.

The `Tactical decisions` subsection contains records of questions and answers to them with their numbers kept.

Each answered question gets heading starting with number and the referensable name starting with `td` in a same format as `impl` or `bug`.

One sub-sub-section for each question, with the choices it was made between where there were some, who decided it, and a link to the section that argues it.
It answers what was decided along the way, by whom, where is was discussed, what code blocks or other asses are related to it.
Section keeps who decided what, so that a decision is neither taken for a fact nor reopened by accident.
It indexes the decisions and does not restate their arguments.

### Relationship to other designs

- **The Relationship to other designs section says what stands next to the
  record: what it depends on, replaces or prepares, and what it leaves to a
  later record. The designs, records and repositories the record touches
  are listed one line each, with the link.**
  - **Answers:** "what else do I have to read, and where is what this
    leaves out?"
  - **Why:** a record draws a border around one problem. Here a reader
    looking for something beyond the border is sent on.
- **Each thing deliberately left to a later record has a Future work
  subsection, with what is already known about it.**
  - **Why:** what was learned on the way but belongs to the next record is
    kept.
  - **Not:** the design of the future work.

### Advice

- **The Advice section lists the findings, reviews, tracker entries and
  user decisions that went into the record and its decisions, each with
  its date.**
  - **Answers:** "what was this formed from, and when?"
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
    found and the code the work turns on, to what helps to read it, and is
    the same in every record.

  | Part | Holds | Form |
  |---|---|---|
  | Evidence | the sources the facts come from, how they were checked, and what the check cannot show | a numbered table, one row per source: Source, What it is, Use; then one sentence on which source wins when two disagree |
  | Evidence gaps | what no source shows | one `gap_` heading per gap with a field table; then what the attempts to close them covered, and how the rest would be closed |
  | Reference repositories | a repository the record measures against: what it holds, and the rules for using it | prose and numbered tables; only where the record depends on one |
  | What "compliant" means | the rules the code is checked against | one `rule_` heading per rule with a field table |
  | Compliance matrix | where the code stands on each rule | a numbered table with one row per checked unit, such as a call, and a second with one row per rule |
  | Finding catalogues | the findings, one catalogue per subject | one `###` section per catalogue, one finding heading per entry with a field table |
  | Dead code | what nothing uses and a commit would remove | a bullet list by kind: functions, types, keys |
  | Measurements | the counts and sizes the record quotes | numbered tables, each under a bold line saying what was measured and on which commit |
  | Code segments | the implementation of the important parts of the code, such as a critical algorithm; for a REST client, the paging: how the pages are loaded and how several pages are merged | one fenced code block per segment, each under a bold line that says what it implements and links to its implementation item, with a sentence or two on what it does |
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
    next.
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
  or commit of this record that takes it, as a link; or, when the record
  leaves it, the record, design or tracker entry that takes it; or "not
  assigned" with the reason.**
  - **Why:** a reader who lands on a finding sees who has it without
    searching, and the implementation item names the finding in return, so
    the two link each other. The row says who has the finding, not that it
    is fixed: a Proposed record has fixed nothing, and the row stays true
    when the work is done.
  - **Example:** "`impl_plugin_rest_error_catalogue`, stage 3"; "the
    `per_page` commit before stage 1"; "design 0003"; "not assigned: no
    caller reads the field".
  - **Not:** "Fixed by", which claims work that has not happened; a finding
    whose only pointer to its item is in the catalogue's opening sentence.

## Skeleton

````markdown
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
- [Decision](#decision)
- [What_changes](#what-changes)
  - [Before_stage_1_<slug>](#before-stage-1-<slug>)
  - [Stage_1_<slug>](#stage-1-<slug>)
- [Consequences](#consequences)
- [Risks](#risks)
- [Affected_systems_and_files](#affected-systems-and-files)
- [Testing](#testing)
- [Tactical_concerns](#tactical-concerns)
  - [Open_questions](#open-questions)
  - [Tactical_decisions](#tactical-decisions)
- [Relationship_to_other_designs](#relationship-to-other-designs)
  - [Future_work_<slug>](#future-work-<slug>)
- [Advice](#advice)
- [Appendix](#appendix)
  - [Evidence](#evidence)
  - [Evidence_gaps](#evidence-gaps)
  - [<repository>](#<repository>)
  - [What_compliant_means](#what-compliant-means)
  - [Compliance_matrix](#compliance-matrix)
  - [<catalogue>](#<catalogue>)
  - [Dead_code](#dead-code)
  - [Measurements](#measurements)
  - [Code_segments](#code-segments)
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
what has to change, what the desired operation or capability looks like.>

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

#### aopt_<area>_<slug>

**<Title>.** <what distinguishes it, in a few sentences; how far it
reaches: what it changes and what it leaves alone>

Advantages: ... (priorities n, m)
Disadvantages: ... (priority k)

#### aopt_<area>_<slug>

**<Title>.** <what distinguishes it, in a few sentences; how far it
reaches: what it changes and what it leaves alone>

Advantages: ... (priority k)
Disadvantages: ... (priorities n, m)

#### aopt_<area>_<slug>

**<Title>.** Only one option was considered. <why; the price it carries;
how far it reaches>

## Decision
[Decision](#decision)

**<the decision that has been taken, in a few sentences: the option
adopted, [aopt_<area>_<slug>](#aopt_<area>_<slug>), and the limits of its
scope>.**

Rationale: ... (priority 1) ... (priority 2). A reader who puts <other
priority> first would choose [aopt_<area>_<slug>](#aopt_<area>_<slug>).

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
once it is done; a link to its code segment, where it has one>

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

**Table n — <Title>** · `tbl_<tag>`

| File | Change | Stage |
|---|---|---|

## Testing
[Testing](#testing)

1. **<corner case>** ([<finding, risk or question>](#<anchor>)). <What has
   to hold there, as something observed, and what counts as a result when
   the case cannot be brought about; where, only when the case can be seen
   in one place alone. One sentence for a simple case, the sequence and the
   timing for a scenario.>

## Tactical concerns
[Tactical_concerns](#Tactical-concerns)

### Open questions
[Open_questions](#open-questions)

<List of questions that are not clear to agent and require explicit
explanation from human. Once question gets clear answer, or decision being
made by human, the question with answer moves to `Decisions`.>

- **1. <question>?** <why it is still open; the choices it is between,
  where there are some>

### Tactical decisions
[Tactical_decisions](#tactical-decisions)

#### 1 `td_<slug>`

**Question**
<actual question text>

**Decision**
<actual decision text>

**Original decision**
<If decision was made through agent, place here actula users text>

**Decided by**
<Who made decision>

**Referecies**
<Where it is argued>

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

### <reference repository>
[<repository>](#<repository>)

<only where the document measures against a repository: what it holds and
the rules for using it>

### What "compliant" means
[What_compliant_means](#what-compliant-means)

#### rule_<area>_<slug>

| | |
|---|---|
| Requirement | ... |
| Checked against | ... |

### Compliance matrix
[Compliance_matrix](#compliance-matrix)

### <catalogue by subject>
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

### Dead code
[Dead_code](#dead-code)

<what nothing uses, and the commit that removes it>

- **Functions:** ...
- **Types:** ...
- **Keys:** ...

### Measurements
[Measurements](#measurements)

### Code segments
[Code_segments](#code-segments)

**<what the segment implements>**
([impl_<area>_<slug>](#impl_<area>_<slug>)). <what it does, in a sentence
or two>

```<language>
<the code of the critical algorithm>
```

### ID sources
[ID_sources](#id-sources)

### Terms
[Terms](#terms)
````
