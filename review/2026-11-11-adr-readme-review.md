# ADR README: coherence, order and overlap

Date: 2026-11-11, night, after the Options scope, the Problem statement questions and the Decision row were added. Reviewed: `docs/adr/README.md` (455 lines) as a whole, for whether its rules agree with each other, whether its own section order and the record order it prescribes make sense, and where it says one thing in several places. Checked against the record where a rule claims a practice: the ten table captions are in reading order; no `dataloss_` name exists; step 3 is a numbered pointer to step 5. Report only; nothing edited.

## Verdict

The README is coherent: its rules agree with each other and with the record on everything a reader would act on, and the three-level split added today (architectural choices in Options, how-it-is-done and editorial choices as questions) is stated the same way in the "One problem" bullet, the Options row, the Open questions row and "What does not get a record". Two contradictions remain, one old and one left by today's rule: the numbering rule for tables is stated twice with opposite meanings, and the skeleton's implementation item still ends "how it is tested". The README's own order mixes levels after the Sections table, and four rules are stated three or four times where one statement and links would do.

## Contradictions

1. **Tables are "never renumbered" and "numbered by order of appearance" at once.** The numbers bullet under Names, not IDs (L146–149) lists tables among the things that "are never renumbered once referenced ... new entries are appended"; Navigation (L240–243) says table "numbering follows order of appearance"; the checks (L258–259) demand tables "numbered without gaps". The record's Tables 1 to 10 are in reading order, so order of appearance is the practised rule, and a table inserted in the middle would be renumbered and move every "Table n" reference after it. One decision: take tables out of the never-renumbered list and say in Navigation that a renumbered table has every reference updated in the same change, the rule the same paragraph already gives for renamed headings; or keep tables fixed and drop "order of appearance" and the no-gaps check for them. Practice says the first.
2. **The skeleton's implementation item says "how it is tested"** (L388: `<what changes, why, how, how it is tested>`), while the What changes row (L84) says "what must hold once it is done" and that the unit tests are the implementing agent's. The skeleton predates today's rule. One phrase: "what must hold once it is done".

## The README's own order

Intro, Records, One problem, File name and status, Sections in order, Names not IDs, How the text is organised, The Testing section, Navigation, Maintaining, Skeleton. The first five read from the outside in: what a record is, which exist, how the problem is cut, how a file is named, what its sections hold. After the table the levels mix: Names, a mechanical rule, comes before How the text is organised, the general one; The Testing section, the only deep-dive on one row, sits between two cross-cutting rules; Navigation, mechanical like Names, is separated from it by two sections. An order that goes general to specific and keeps like with like:

Sections in order → The Testing section (next to the table that links it twice) → How the text is organised → Names, not IDs (with the numbers) → Navigation → Maintaining (whose checks test the two mechanical sections just before it) → Skeleton.

Moving sections changes no sentence; the two links to The Testing section and the one to Navigation keep working.

## The record order the README prescribes

Problem, Context, Priorities, Options with Decision, What changes, Consequences, Risks, Affected systems, Testing, Open questions, Relationship, Advice, Appendix: facts, values, choice, work, what follows, how it is checked. It makes sense, and the intro's one-sentence summary (L4–6) follows it. Two things to weigh:

- **Consequences come after What changes.** Consequences, the confidence and the re-evaluation trigger belong to the decision; in 0002 about nine hundred lines of implementation separate them from it. The classic record puts Consequences right after Decision. Moving Consequences, and perhaps Risks, before What changes keeps the decision-level sections together at the price of forward references to stages and items, which the risk guards make. Either order is defensible; the README should say why it chose this one, in the Consequences row or the intro.
- **The Decision row.** Decision is a `###` subsection with its own row, while Key observations, Decisions, Future work and the stages are described inside their parents' rows. The asymmetry is justified, the decision is what the Records table and File name and status quote, and the row says it is a subsection; it only needs to be deliberate.

## Overlap

The layering rule → Sections row → skeleton is by design and reads well. Beyond it, the same rule is stated in three or four places; the README's own line applies, "when two sections start to say the same thing, one of them links to the other":

- A decided question moves to the Decisions table with its number kept: Open questions row (L89), numbers bullet (L151–152), Maintaining (L252–253), checks (L260).
- A new risk is appended with the next number: Risks row (L86), numbers bullet (L149), Maintaining (L253).
- The mechanism lives in What changes only: Options row (L82), What changes row (L84), Each fact once (L166).
- Unit tests are the implementing agent's: What changes row (L84), Testing row (L88), The Testing section (L195–198); two of the three link the third, which is the acceptable form.
- Self-anchor lines: Sections intro (L69–70, with a link), Navigation (L237), Names bullet (L132–133); named headings absent from the Table of Contents: ToC row (L75) and Names bullet (L133).
- Once Accepted the text is frozen: Status table (L60) and Maintaining (L254).
- "From the cheapest to the live one": Testing row (L88) and The Testing section (L227), verbatim.
- What is outside the record is said under "not solved": One problem bullet (L43–45), Problem statement row (L79), skeleton (L331).

Where to cut: the Sections rows are the definitions and stay; Maintaining keeps only what happens over time (a new entry, a move, a freeze) and links the row for the form; the numbers bullet keeps the never-renumber rule and the bullet form; Each fact once keeps its list, which is a summary by nature. The Status table's "frozen" and the Testing row's "cheapest to the live one" can each lose one copy.

## Low

- **L53–54 "The decision is the bold paragraph in the record and a column in the table above."** The Records column for 0002 also lists the question decisions of stages 3 and 4 and the one-series delivery, which are not in the bold paragraph. "Summarised in a column".
- **L88 "the tests that check the record".** They check the implementation; The Testing section says so (L189–190).
- **L241 the caption example's backticks are malformed.** `` `**Table n — Title** · `tbl_tag`` `` closes the code span at the second backtick and leaves "tbl_tag``" outside it. Wrap the whole example in double backticks with a space inside.
- **L75 "(see below)"** could link Names, not IDs.
- **Names, not IDs holds the numbering rules and the bullet form for lists with gaps** (L146–155); its title says names. "Names and numbers", or a short section of its own.
- **Two misfiled sentences.** "References are links" (L180–185) carries the rules on test-environment details and secrets; "Plain language" (L171–176) carries the tense rule. Each is a rule of its own.
- **The skeleton's Table of Contents lists Evidence and ID_sources under Appendix** (L300, L302) and the skeleton body shows neither, so the skeleton would fail its own check (L257–258). Both or neither.
- **The skeleton's Testing step** (L416–418) omits "what is recorded when the result is neither pass nor fail", the fifth part of The Testing section.
- **Maintaining says nothing about a retired Testing step.** The record's practice is a numbered one-line pointer ("3. NFS code. Folded into step 5"), which is how "never renumbered" and "no gaps" hold at once for steps. One sentence.
- **Consequences and Risks have no boundary.** A negative consequence is certain once the record is applied; a risk is a possible failure with a guard. 0002 keeps them apart and cites risks from Consequences by number; one clause in the Consequences row would keep the next record doing the same.
- **The `dataloss` row names no catalogue** ("Appendix"), and the record has no `dataloss_` entry yet. Fine until one exists; then the row names where it lives.
- **L119–121 keeps the history of `work` becoming `impl`.** The README forbids history in a record, not in itself, but the sentence ages; the registry README in `kub-testing/issues/` is where old names are mapped.

## Verified and found coherent

- The three decision levels: architectural in Options (L82), how-it-is-done as questions (L82), editorial as one line in the Decisions table (L40–42), with 0002 as the example (L30–32).
- Dates: Options decisions in Metadata, question decisions in Advice (L77), acceptance separate (L64–65), no dates in the body (L171–173).
- The Problem statement row's four questions against the skeleton's labels (L327–331) and the record's ("What has to change", "What success looks like, when this document is done", "Not solved by this document").
- The Decision row against the skeleton's Table of Contents (L288), its block (L372–378) and the record's `### Decision` under Options considered.
- The kinds table's "Where its heading lives" column against the headings bullet (L129–130) and the Appendix row (L92).
- The checks (L256–266) against the rules they test: anchors, ToC, numbering, the bullet form, the Testing runner, names, old IDs, file links, forbidden phrasing; each check has its rule.
- The Testing rule and its two rows: the What changes row and the Testing row both link The Testing section, and the section's five parts match the skeleton's step except for the omission above.

## Addendum: the Options example against its rows

Asked the same night: is the skeleton's Options block (L353–378, with its Table of Contents entries L284–288) coherent with the Options row (L82) and the Decision row (L83)? Mostly. What matches: the intro placeholder names the parts of the Problem statement and what became questions; A1 is first and marked adopted; the one-option form says so in the heading and the first sentence and asks for the why and the price; Decision is last, bold, with the rationale by priority number and the sentence for a reader with another order of priorities.

- **Group A shows one contender.** The row asks for two or three real contenders per choice; the body has A1 (adopted) only, while the Table of Contents lists `Option_A2_<slug>` (L286) with no block in the body. The example therefore shows the shape the row forbids, a contenders group of one, and the skeleton fails its own Table of Contents check. Add the A2 block, with advantages and disadvantages.
- **Disadvantages carry no priority numbers** (L365), while "Advantages: ... (priorities n, m)" does and the Priorities row says the options argue against them by number. The record numbers both sides.
- **"<what it is, in a few sentences>"** (L362) is looser than the row's "what distinguishes it, not the mechanism".
- **The naming is shown, not described:** A1, A2 within a group, a bare "Option B" for a one-option group. Today's edit dropped the old clause "grouped A, B, … when the record holds several choices", and the digit convention never had one. One clause in the row.
- **The record's second Decision paragraph** names the question decisions that shape the delivery; neither the Decision row nor the skeleton provides for it, so the only record departs from both. Either the row admits it ("then, where question decisions shape the delivery, the paragraph that names them") or the record's paragraph moves under the Decisions table.
- Optional: the skeleton's letters are the reverse of 0002, where A is the one-option group and B holds the contenders.

**Outcome (same night, on the user's question):** Decision is now its own section right after Options considered in the README: the Decision row says so, the Options row no longer ends with it, the skeleton's Table of Contents lists `[Decision](#decision)` at the top level and the skeleton body has `## Decision`. The skeleton's implementation item says "what must hold once it is done" (contradiction 2). Not changed: the record, whose `### Decision` would become `## Decision` with its Table of Contents line moved to the top level; the anchor `#decision` stays, so no link changes. The other findings stand.

**Outcome (later the same night, on the user's request to fix what is clear in backlog items 48 and 49):** applied to the README: tables are numbered by position (the numbers bullet says so and links Navigation, which now says a table inserted or removed renumbers those after it with every reference updated in the same change); the sections reordered to Sections, The Testing section, How the text is organised, Names, Navigation, Maintaining, Skeleton; the Consequences row says why it follows What changes and that a possible failure is a risk; the Risks row lost its restatement of the append rule; "summarised" in File name and status; "the tests that check the implementation"; the caption example's backticks; the link in "(see Names, not IDs)"; the tense rule and the lab-details rule as bullets of their own; the retired-step rule in Maintaining; the skeleton's Evidence, ID sources, Terms blocks and the Future work entry of its Table of Contents; the skeleton's Testing step with "what is recorded"; the Options row describes the A1, A2 and Option B naming; the Decision row and the skeleton admit the paragraph on question decisions; the skeleton shows A2, "what distinguishes it", and priority numbers on both sides. Not applied, for the user: the title of Names, not IDs; the `dataloss` row; the `work` to `impl` history; the remaining overlap, which on a second look is one rule in four roles (definition, general rule, process, check) rather than four copies, so only the Risks row's sentence was cut.
