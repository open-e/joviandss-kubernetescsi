# ADR 0002: coherence with the Problem statement

Date: 2026-11-12, after the Problem statement was replaced by four sentences (`review/2026-11-12-adr-0002-problem-statement-review.md`). Report only: nothing in the record is changed; each finding says what the user would decide.

The four sentences, numbered here for reference only:

- S1. The plugin uses JovianDSS REST API v3, which is deprecated, through a client that follows documentation and assumptions rather than what the appliance serves.
- S2. It has to use API v4 through a client that complies with the API as served.
- S3. Success is compliance with the JovianDSS 1 REST API v4.
- S4. JovianDSS 2 will not be supported within this iteration.

Method: every section was read against them. A finding is one of three things: a sentence of the record that says the problem, the change, the success or the exclusion differently from the statement; work the statement does not ask for; or a claim of the statement that no section substantiates. Line numbers refer to the record as of this review (2,569 lines).

## Findings

### 1. High: success is compliance, but a third of the delivery is not compliance work, and the record says so itself

Where: Stage 4 (lines 698 to 706) delivers the five crash fixes, no request logging, path escaping, name and ID validation, TLS verification by default, the reply size limit and no secrets in any log. Stage 3 (588) also fixes the two driver bugs of the delete cleanup. The first commit (`go vet`) and the dead-code commit are hygiene. The Relationship section says it in so many words: "REST client behaviour: not a compliance matter" and "the other three entries are fixed here" (1501 to 1505); "Security: every security finding is fixed in this delivery" (1506). The Decision's third paragraph (406 to 416) calls all of it "the migration", whose "decisions were taken as questions"; Advice records the user's decision of 2026-11-10 that stages 3 and 4 belong to the same migration (1616 to 1617).

Against: S3, and the definition the statement links to (1869): the plugin is compliant when every call meets seven rules on calls, request bodies, success statuses, error statuses, reply types, error bodies and the documentation. None of the seven asks for a crash fix, a verified certificate, a size limit or a clean log. Read against the statement, stage 4 is work the problem does not ask for.

One sentence makes the contradiction explicit: the What changes introduction (425 to 427) says that "the client is compliant with the recorded v4 API only when stages 3 and 4 are done as well". It was moved there on 2026-11-12 from the old Problem statement, where "compliant" covered everything the old statement listed; under the definition at 1869 and the Relationship bullet at 1501, stage 4 is not what makes the client compliant, so the sentence is wrong as it stands.

Three ways to coherence:

- (a) Widen the statement: S2 and S3 cover robustness and security as well. This puts back the kind of detail the user removed on 2026-11-12.
- (b) Widen the definition: "What compliant means" gains rules for robustness and security. The audit checks the seven rules against the spec repository; these could not be checked there, so the compliance matrix would carry rows nothing measures.
- (c) Keep the statement and the definition; say once, in the Decision or the What changes introduction, that the series also carries the robustness and security fixes of the same code, found by the same reviews, and that they are not part of the success criterion; correct the moved sentence to "stages 1 to 3 make the client compliant; stage 4 is carried in the same series".

Recommendation: (c). It matches the user's decision of 2026-11-10 and the words the user chose for the statement. User decision.

Update, later on 2026-11-12: the user added security as priority 2 (the list and its references were renumbered). Stage 4's security items are then in the series by priority, which is the record's own way of justifying work; the crash fixes and the size limit are covered if security is read to include robustness against a broken or hostile appliance. Under (c) what remains is the one sentence that says the series carries them, and the corrected sentence in the What changes introduction.

### 2. Medium: the title and the Decision's bold text frame the record as "v4 support", the statement as compliance

Where: the title (line 1), "API v4 support for JovianDSS 1"; the Decision's bold text (382 to 392), which decides the version, the generated types and "JovianDSS 2 is not touched", and leaves compliance beyond the types to the third paragraph (406). Metadata (88) records that the record was formerly `0002-rest-api-compliance-design.md`.

Against: S2 and S3, where the client's compliance is the larger half of the problem and the whole of success.

Recommendation: the bold text gains one sentence: the REST client is brought to compliance with the recorded jdss1 v4 API as the Appendix defines it, with the decisions of that work in the Decisions table. The title would read "API v4 compliance for JovianDSS 1" to match; whether the file name follows is the user's call under the README's file-name rule.

### 3. Medium: "assumptions" in S1 is shown but never said

Where: Context substantiates "documentation" in full: the reply types were transcribed from documentation that is itself wrong in places, and "the code follows the documentation" (132 to 157). "Assumptions" has no sentence of its own. Key observation 8 (253) says that error handling depends on message text; that the statuses the code accepts, the patterns it matches and the bodies it sends were written from assumptions is said only by instances in What changes and the Appendix: "a test string that never matches" (595), "the patterns no recording supports" (645), the create status passed on a delete (Key observation 7), the `{}` body on every GET. The paragraph added to "The REST layer, measured" on 2026-11-12 (204 to 210) lists the catalogues without naming the cause.

Against: S1. The statement's first claim about the client is carried by the Appendix, not by Context, which the README says carries the claims.

Recommendation: one sentence in Context, in Key observation 8 or in the paragraph at 204: the statuses the client accepts, the messages it matches and the bodies it sends were written from assumptions, and the recordings show where each is wrong; the instances stay in the Appendix.

### 4. Medium: the exclusion of JovianDSS 2 is said in five places in four wordings, and "iteration" is defined nowhere

Where: S4 "will not be supported within this iteration"; the Decision's bold text (392) "is not touched"; the What changes introduction (429) "not part of this delivery: it would come in a later plugin version"; Relationship (1513) "outside this document; a later design covers it"; Future work (1519) "Not part of this document (user decision)"; What "compliant" means (1871) "this document does not touch JovianDSS 2". Advice (1616) is a dated log of the decision and stays. The record says "iteration", "delivery", "series" and "this document" for the same thing; Terms (2532 to 2568) defines none of the four.

Against: the README's rule that what a record does not solve is said in a sentence of the Problem statement and nowhere else, with the place that handles it named in Relationship; and "each fact once".

Recommendation: Terms gains "iteration: the one series of commits this record describes, built and tested in one go; also called the delivery". The Decision keeps its sentence, since it is the decision; Relationship keeps the pointer; Future work keeps its opening. The What changes introduction and What "compliant" means link to the Problem statement instead of restating it. Which word wins, iteration or delivery, is the user's choice.

### 5. Low: the Options introduction ties group B to the reply types, not to the client's compliance

Where: 279 to 283: "Two matters are decided here: the API version (A) ... and how the reply types are made to match it (B) ... The migration's other decisions were taken as questions".

Against: S2 has two parts, the version and a compliant client; group B answers the types only, and the README allows the rest to be questions when the fix follows from the evidence. The reader is not told that this is why.

Recommendation: one clause: the rest of compliance (statuses, errors, request bodies) follows from the recordings and was decided as questions.

### 6. Low: S3 says "the JovianDSS 1 REST API v4"; the record proves compliance with one release

Where: Terms defines jdss1 as release `1.0up33.65410`; question 29 says older releases are not supported; Consequences' confidence paragraph (975 to 979) claims nothing for them.

Against: S3 reads broader than what is shown. It is coherent through Terms and question 29; "the recorded JovianDSS 1 release" would be exact. The sentence is the user's, so this is noted, not recommended.

## Coherent as they are

Context's API versions (the plugin uses only v3; v3 served on jdss1, gone on jdss2: S1), the reply types subsection (documentation: S1), Priorities 1 and 3 (JovianDSS 1; evidence over documentation: S1 and S3; priority 2, security, was added after this review), Key observation 1 (S1), Option A (S2; it cites the statement at 284 and the citation still holds), Consequences and its confidence paragraph (S3), Testing step 2, which proves S3 in its own words ("every call the plugin makes meets the rules on jdss1 v4"), What "compliant" means (1869, the target of the statement's link), the Relationship bullet on JovianDSS 2 and the Future work subsection (the place S4 points to).

Earlier backlog items on these sections (40, 46) are unchanged by the new statement.

## Summary

One High, which needs the user's decision between widening the statement, widening the definition of compliant, or keeping both and saying once that the series carries more than the success criterion measures, with the moved sentence in the What changes introduction corrected either way. Three Medium: the Decision's bold text and the title, the unsaid "assumptions", the five restatements of the exclusion and the undefined "iteration". Two Low. Nothing is changed. Backlog item 55 holds the decision.

## Outcome so far

Later on 2026-11-12, at the user's request: security became priority 2, as the bare bold name (the list and its 15 references renumbered; the README moved priorities to the position-number regime), and the Decision's bold text now opens with one sentence, the REST client brought to compliance with the JovianDSS 1 REST API v4, followed by the original text. A first rewrite had added a restatement of the compliance rules, a scope sentence about the security and crash fixes and a priority-2 clause in the rationale; the user rejected those as details placed outside their sections, and they were removed. That closes the Decision half of finding 2; finding 1 is open in full, including the moved sentence in the What changes introduction; the title and findings 3 to 6 are open. Both checkers pass.

Later still on 2026-11-12, by the user's decisions: the sentence of finding 1 in the What changes introduction stays as it is; the record was renamed `0002-api-v4-compliance-jdss1.md` with the title "API v4 compliance for JovianDSS 1", which closes the title half of finding 2 (the dated reviews keep the old file name); and Options considered was reduced to Option A, the user having ruled that options are strategies and B1 to B5 were implementation nuances, now the alternatives of question 28. The new Options introduction says that the rest of compliance is implementation decided as questions, which closes finding 5. Open: findings 3, 4 and 6, and how the Decision names priority 2 now that the options are in order. Both checkers pass at 2,499 lines.

Last on 2026-11-12, at the user's request, the Decision was rewritten from Option A and the Risks: the bold text is compliance, v4 only with the minimum release, and the exclusion; the rationale names all eight priorities, each with the risk it guards or accepts (risk 2 accepted by question 29; risks 10 and 11 the price of priority 2; risks 15 and 16, 9, 17 to 19 and 22, 7; risks 12 and 14 without a guard); the reply types moved from the bold text to the question paragraph as question 28. That closes the priority-2 item; findings 3, 4 and 6 remain.
