# ADR 0002: coherence with the Decision

Date: 2026-11-12, after the Decision was rewritten from Option A and the Risks (`review/2026-11-12-adr-0002-coherence-review.md`, Outcome). Report only: nothing in the record is changed; each finding says what the user would decide. Line numbers refer to `docs/adr/0002-api-v4-compliance-jdss1.md` as of this review (2,504 lines).

The Decision, as read here: the bold text (lines 304 to 310) decides compliance with the JovianDSS 1 REST API v4, v4 only with no fallback, no setting and the minimum release `1.0up33.65410`, and "JovianDSS 2 is not touched"; the rationale (312 to 329) names the eight priorities, each with the risk it guards or accepts, and ends with the risks accepted without a guard, 12 and 14; the question paragraph (331 to 343) names the question decisions that shape the delivery: 28, the error catalogue, 7, 20 and 21, 23 and 26, 19, 24.

Method: every section was read against those three parts. A finding is a sentence of the record that says the decision, its reasons or its price differently from the Decision, a claim of the Decision that a section contradicts or leaves unsupported, or work the record delivers that the Decision does not name where it names its peers.

## Findings

### 1. Medium: Metadata dates a question decision as the Decision's date

Where: Metadata (78): "Date: 2026-11-05 (API v4 only decided), 2026-11-10 (generated reply types decided, replacing the hand-written rule of 2026-11-09)".

Against: the README's Metadata row: `Date:` is when the decision of the Decision section was taken; the question decisions are dated in Advice. Since today the reply types are question 28, not part of the bold decision. The bold decision's own dates are 2026-11-05 for v4 only and 2026-11-10 for the whole migration in one record, which is what makes compliance the decision (Advice, 1545 to 1548); the types' date already stands in Advice.

Recommendation: "Date: 2026-11-05 (v4 only decided), 2026-11-10 (the whole migration, compliance with the recorded v4 API, decided as one record)", the rest of the line unchanged.

### 2. Medium: the Decision names two risks accepted without a guard; Risks has three

Where: the rationale's last sentence (328 to 329): "Without a guard in this delivery: a hung call past the CSI deadline (risk 12) and a reply behind the request (risk 14)". Risks: "Guard: none" at 1039 (risk 12), 1044 (risk 13, "the edge is accepted with the limit of 131 (user decision, question 21)") and 1067 (risk 14); the Decisions table says the same of 21 (1379).

Against: a sentence that lists the unguarded risks and omits one is wrong, not incomplete.

Recommendation: "Without a guard in this delivery: a hung call past the CSI deadline (risk 12), a reply behind the request (risk 14), and the accepted edge of a name of exactly 132 characters (risk 13, question 21)"; or the sentence names 12 and 14 as future work and 13 as accepted, in the same words the Risks use.

### 3. Medium: "JovianDSS 2 is not touched" is contradicted by the record's own account of what it does for JovianDSS 2

Where: the Decision (310) and Advice (1545): "not touched". Future work (1460 to 1461): "This document does not touch JovianDSS 2; it prepares for it only by giving 401 its own error code and by keeping the jdss2 recordings beside the jdss1 ones as evidence." What "compliant" means (1805 to 1806): "The jdss2 v4 recordings are evidence beside it ... because this document does not touch JovianDSS 2."

Against: the Problem statement's words, "JovianDSS 2 will not be supported within this iteration", are exact; "not touched" is contradicted in the next clause wherever it appears. This is the previous review's finding 4 seen from the Decision: the Decision is now the place where the wrong wording is decided.

Recommendation: the Decision's last sentence takes the Problem statement's wording; the other restatements as in the previous review's finding 4 (user decision pending). Advice keeps its dated wording.

### 4. Low: stage 4's crash fixes and the reply size limit are not named in the Decision where their peers are

Where: the question paragraph names the error catalogue, question 7, questions 20 and 21, 23 and 26, 19 and 24: every stage 3 and stage 4 item except the crash fixes (question 12, 1370; the item at 638; the Crashes introduction at 2279, "Stage 4 would fix all five") and the reply size limit, which has no question of its own.

Against: a reader of the Decision would not learn that the series fixes the REST-layer crashes, while it learns that release notes live in `CHANGELOG.md`. Whether every question decision that shapes the delivery is named is the README's "where question decisions shape the delivery"; the crash fixes shape it more than question 24 does.

Recommendation: one clause, "every crash fixable inside `pkg/rest` is fixed (question 12)", or the paragraph says that it names the decisions that change behaviour and leaves the rest to the table; user's call.

### 5. Low: the rationale credits a guard that risk 19 does not have

Where: the rationale (324): "A new recording is a diff and a check (priority 5), the guard of risks 17 to 19 and 22". Risk 19 (the spec output format moves): "Guard: the generator is pinned to a spec commit and reads one spec output; a format change is a generator change, not a types change."

Against: risk 19's guard is the pin, not a diff or a check.

Recommendation: "the guard of risks 17, 18 and 22; risk 19 is held by the pin to a spec commit", or simply "risks 17, 18 and 22".

### 6. Low: the price of priority 2 names two of the four upgrade-stoppers

Where: the rationale (317 to 319): "at the price of upgrades that stop until the certificate and the TLS version are dealt with (risks 10 and 11)". Risks, "Backward compatibility, not preserved" (940 to 947): also "configurations whose `iscsi.iqn` prefix breaks the target rule" (questions 20, 21) and "configurations with `prot: http`" (question 23).

Against: the Decision names a price and the Risks list a larger one; the two missing stoppers are prices of question decisions the Decision does name.

Recommendation: either the rationale says "and until an `iscsi.iqn` prefix or `prot: http` is changed", or the question paragraph's clauses for questions 20, 21 and 23 carry their price in three words each; or nothing, if the price of a question decision belongs to the Risks alone.

### 7. Low: "an unknown error stays unknown" is said twice inside the Decision

Where: the rationale (328, priority 8) and the question paragraph (337, "an unknown one stays Unknown, logged once in full"). Priority 8 itself (270) says it a third time.

Against: each fact once; inside one section twice is the clearest case.

Recommendation: the question paragraph's clause becomes "every appliance error is mapped from one catalogue ([impl_plugin_rest_error_catalogue])", the rationale keeping the why.

## Known and decided, not findings

- The What changes introduction's sentence that the client is compliant only once stages 3 and 4 are done (user decision of 2026-11-12: it stays).
- "Iteration" undefined and the six wordings of the exclusion (previous review, finding 4, pending); finding 3 above is its Decision-side effect.
- Risk 10's TLS handshake check is not a Testing step (backlog 44).

## Coherent as they are

The Problem statement and the bold text say the same three things, the exclusion's wording apart. Option A and the Decisions table's question 1 say v4 only, no other option, as the bold text does. Priorities 1 to 8 are each named once in the rationale, in order. Context's API versions and Key observation 2 carry the rationale's "the move changes no behaviour". Consequences (one API; older releases stop; generated types; one series; the spec repository as an input) matches the bold text, question 28 and priorities 5 and 7. Risk 2 and question 29 carry priority 1's accepted price. Relationship's security bullet and the Security findings introduction say what priority 2's clause says; its REST client behaviour bullet leaves exactly the two entries the Decision accepts as risk 12 and future work. What "compliant" means is the target of the bold text's link and defines it as the rationale uses it. Testing step 2 proves the bold text's first sentence; the introduction's "tested once" is priority 7. Stage 1 delivers the bold text's second sentence and question 24; Stage 2 question 28; Stage 3 the catalogue and question 7; Stage 4 questions 19, 20, 21, 23 and 26. Advice dates every decision the Decision names.

## Summary

Three Medium: the Metadata date, the unguarded-risk sentence that misses risk 13, and the "not touched" wording the record itself contradicts. Four Low: the crash fixes unnamed, risk 19's guard overclaimed, two upgrade-stoppers unpriced, and one fact said twice. Nothing is changed. Backlog item 56 holds the decision.

## Outcome

Applied on 2026-11-12 at the user's request ("fix inconsistencies that are clear and does not require human involvement"): finding 1, the Metadata date now reads "2026-11-05 (API v4 only decided), 2026-11-10 (the whole migration, compliance with the recorded v4 API, decided as one record)"; finding 2, the unguarded-risk sentence names risks 12, 13 and 14; finding 3, the Decision's last sentence is "JovianDSS 2 is not supported within this iteration", the Problem statement's words, Advice keeping its dated wording and the other restatements untouched; finding 5, the rationale credits priority 5 with the guard of risks 17, 18 and 22; finding 7, the question paragraph's clause is "an unknown one is logged once in full", the rationale keeping "stays unknown". Not applied, the user's call: finding 4 (whether the Decision names the crash fixes and the reply size limit) and finding 6 (whether a question decision's price belongs in the rationale). Both checkers pass.
