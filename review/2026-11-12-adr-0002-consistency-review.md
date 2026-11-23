# ADR 0002: coherence of the whole record

Date: 2026-11-12, at the end of a day that changed most sections: the Problem statement replaced by four sentences, 38 names renamed, security added as priority 2 with the list renumbered, the Decision rewritten twice, Options considered reduced to Option A, the file renamed `0002-api-v4-compliance-jdss1.md`, a paragraph added to Context, a sentence to the What changes introduction and to a Consequences bullet, and five fixes from the decision-coherence review. Report only: nothing in the record is changed. Line numbers refer to the record as of this review (2,505 lines). Both checkers pass.

Method: every section was read, the Appendix included, and read against the others: the Problem statement, Context, Priorities, Option A, the Decision, What changes, Consequences, Risks, Affected systems, Testing, the Decisions table, Relationship, Advice, the Evidence sections, the compliance definition and matrix, the catalogues' introductions, Measurements, ID sources and Terms. A finding is two places that say one thing differently, a claim with no section behind it, or a section that the rest no longer refers to as it is.

## Verdict

The record is coherent in its structure and its claims. No High, no new Medium. Four Low findings, three of them wording that today's rewrites made visible. Two Medium items are carried over from the earlier coherence review (backlog 55), and this pass measures one of them: the exclusion of JovianDSS 2 is restated in nine places beyond the Problem statement. The pre-existing backlog items are unchanged by today's work, except one that today's decisions may close.

## Findings

### 1. Low: "no request is logged" says more than the item it cites

Where: the Decision's question paragraph (344): "no secret reaches any log at any level and no request is logged". The item (686 to 688): "Method, path (with its query string) and status stay logged at Debug: Testing step 4 reads `per_page` from that line. This disables request logging for now"; the item's "request logging" means the body and the dump.

Recommendation: "no request body or header is logged".

### 2. Low: question 19 is cited for a decision it does not hold

Where: the same clause (344) cites question 19 for the no-secrets and no-request-logging decisions. Row 19 (1378): "jdss2 tokens in fixtures? tokens are secrets: never logged, never in a fixture". The two decisions the clause states are the items' own ("user decision" in both items), and the clause already links the items.

Recommendation: the clause cites the two items, with question 19 for the tokens only, or not at all.

### 3. Low: question 28's price restates two risks

Where: row 28 (1385): "Price: about 650 lines of generator and mapping to keep, a procedure that depends on the user running the spec repository's tools, a spec output format still moving." Risk 19 (1092) is the moving format, risk 20 (1096) the one-person procedure. The clause was written on 2026-11-12 when Option B1 was removed.

Recommendation: "Price: about 650 lines of generator and mapping to keep, and risks 19 and 20."

### 4. Low: row 2's "where it is argued" points at a section that states, not argues

Where: row 2 (1361): "Minimum JovianDSS 1 release? `1.0up33.65410`, the release of the lab and the recordings. Where it is argued: Decision". The Decision's bold text states the release; the argument, the latest release, the one recorded and in the lab, no older one supported, is Option A's price and risk 2 with question 29.

Recommendation: "Option A, risk 2".

## Carried over, pending (backlog 55)

- Context shows but never says the Problem statement's "assumptions" (the earlier coherence review, finding 3).
- The exclusion of JovianDSS 2, counted on the whole record: beyond the Problem statement (96 to 97) it is restated in the Decision (310, now in the Problem statement's words), the What changes introduction (359), Table 2's jdss2 row (482), the closing line of Affected systems (1159), Testing step 10 (1310), Relationship (1443), Future work (1449), What "compliant" means (1806) and Table 8's error-bodies row (1927), in five wordings; Advice (1545) is a dated log and stays. "Iteration" is defined nowhere, while "delivery", "series" and "this document" name the same thing (the earlier coherence review, finding 4).

## Known and decided, not findings

- The What changes introduction's sentence that the client is compliant only once stages 3 and 4 are done (355): the user's decision of 2026-11-12 is that it stays.
- The decision-coherence review's open items (backlog 56): whether the Decision names the crash fixes and the reply size limit; whether the price of questions 20, 21 and 23 is stated in the rationale.

## Pre-existing, unchanged by today

Backlog 39 (the hurl suite's `project-status.md` cited as evidence, Table 4 and the "hurl Finding n" citations); 40 ("the 36 call functions" at 190 and "the 36 calls" in the Risks' preserved paragraph against "all 34 calls" at 1870); 41 (the two security names without a catalogue entry, still in code font only); 42 (the three prose `####` headings under Evidence gaps); 43; 44 (Testing step 4 now names the TLS handshake check as a prerequisite, which answers half of it); 47 (step 4's "whose texts the items fix", "the request log", risk 7's "steps 1-3" for steps 1 and 2); 50; 51; 52. Backlog 46's first point, that Option A's price "or with a different v4" predates question 29, may be closed: question 29 says that older releases do not serve the same v4, which is what the hedge describes.

## Coherent as they are

The Problem statement's four sentences against the Decision's three, Option A, question 1, the compliance definition and Future work. Priorities 1 to 8 against the rationale, the five tags in Option A's neighbourhood and the Appendix, all renumbered. Option A against the Decision's bold text, risk 2, question 29, Consequences' confidence paragraph and Context's "What we do not know". The Decision's question paragraph against rows 7, 20, 21, 23, 24, 26 and 28 and against Stages 1 to 4. What changes against Table 3's stages and Terms' "stage 1-4". The error catalogue item against priority 8, the Decision's rationale, rule 4 and Table 8. The generated types item against question 28, Table 2, risks 1, 3, 9, 16 to 22, Testing step 2 and the Consequences bullet on tolerant fields. Stage 4 against the Security findings and Crashes introductions, the Relationship's security bullet and risks 10, 11 and 13. Testing steps 4 to 7 against risks 8, 10, 11 and 15 and question 30. The Evidence sections, the spec repository section and Measurements against Context's reply types and REST layer subsections. ID sources and Terms against the names rule and the registry.

## Summary

Four Low findings, each one clause; two Medium carried over, one now counted at nine restatements; the What changes sentence by decision; the pre-existing backlog unchanged, with one item that may close. Nothing is changed. Backlog item 57 holds the decision.

## Outcome

Applied on 2026-11-12 at the user's request ("fix inconsistencies that are clear and does not require human involvement"): the Decision's clause reads "no request body or header is logged (question 19 for tokens; the two items)"; question 28's price is "about 650 lines of generator and mapping to keep, and risks 19 and 20"; row 2's "where it is argued" is Option A and risk 2. The carried-over items and the pre-existing backlog are untouched. Both checkers pass.
