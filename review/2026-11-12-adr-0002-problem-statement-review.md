# ADR 0002: the Problem statement against the rule of 2026-11-12

Date: 2026-11-12. The user corrected the README's description of the Problem statement: the four questions ("What's wrong, or what's missing? What change in system behavior do you want? What does success look like? What problems are you explicitly not solving?") are a quote from ceejbot's blog post on lightweight decision records, given to convey what a problem statement is; the statement "can be answer to one or all of them", it "is not a bullet list but a minimalist coherent sentence or set of sentences describing problem that ADR is going to solve", and "not a comprehensive covering of future tests, unrelated issues and etc". On 2026-11-11 the questions had been copied into `docs/adr/README.md` as a form: a bold sentence per part, "What has to change:", "what success looks like, stated once", "not solved ... where each is handled". The README now carries the rule in its own words, with an example and counter-examples (section "Sections, in order", the "not solved" rule, the table row, the "each fact once" example and the skeleton). This report measures the record's own Problem statement against it. Report only: the section changes on the user's decision.

## Verdict

The section (lines 95 to 181 of the record, 85 lines) was written to the form of 2026-11-11 and fails the rule in every way the rule names: three bold parts, each followed by its evidence and its links into the catalogues; a labelled "What has to change" paragraph; a seven-bullet "What success looks like" list that names compliance rules, implementation items, the csi-sanity baseline cycle and the log checks, which is the Testing section and the What changes section said a second time; and a "Not solved" paragraph of four exclusions with a pointer each. A reader cannot hold it in mind, and the rest of the record is read against the Testing section it repeats rather than against a problem.

The content is not lost by shrinking it: every statement of the success list and of the exclusions already lives where the README places it, with two exceptions, listed below.

## Draft

A first draft of five sentences, which listed the failing calls, the client's faults and the mechanisms of the fix, was rejected by the user on 2026-11-12 as "oversaturated with details" and "generalised poorly". The second draft names the class of the problem and leaves its instances to Context:

The plugin uses JovianDSS REST API v3, which is deprecated, through a client that follows documentation and assumptions rather than what the appliance serves. It has to use API v4 through a client that complies with the API as served. Success is compliance with the JovianDSS 1 REST API v4. JovianDSS 2 will not be supported within this iteration.

Four sentences; the last two are the user's words (2026-11-12), replacing a success sentence that pointed at the compliance rules and an exclusion that said only "not solved here". Every link of the current section belongs to Context and to the Appendix. The exclusions of older JovianDSS 1 releases and of the hand-written request types are decided inside the record (question 29, the request body rule) and are not repeated in the statement.

## Where each dropped statement already lives

| Statement of the current section | Lives in | Record line |
|---|---|---|
| 34 hard-coded `api/v3` strings | What changes (the prefix), Affected systems, Measurements | 562, 1248, 2530 |
| five never-read fields with a type the appliance does not send; the calls that fail | the Response types catalogue, Context's key observations | Appendix |
| 86 fields decoded on the volume alone | Option A2, the generated types item, the types table | 419, 1022, 2554 |
| the oldest supported JovianDSS 1 release stated in the documentation | What changes, "Deprecating v3" | 574 |
| every call exists on v4, is answered as recorded and meets every rule | What "compliant" means, the compliance matrix | Appendix |
| docs and appliance disagree: follow the appliance, list the difference | `rule_rest_docs_vs_appliance_listed` | Appendix |
| jdss2 recordings are evidence only; token login first | Relationship, Future work | 1498, 1587 |
| a missing read field fails the call with the field's name | the generated types item | 612 |
| a changed recording shows as a diff of the generated file | the generated types item | 643 |
| csi-sanity shows no `cannot unmarshal`; iSCSI as good as on v3; baseline cycle; NFS without baseline | Testing steps 4 and 5 | 1347, 1363 |
| every known error maps to one code, unknown stays unknown, logged once in full | the error catalogue item | 478 |
| no REST failure crashes the plugin | the crash fixes item | What changes |
| no password, CHAP secret or CSI `Secrets` in any log | the two logging items | What changes |
| HTTPS, TLS 1.3, certificate verified unless configured otherwise | the TLS item | 1079, 1159 |
| older JovianDSS 1 releases unsupported | question 29, risk 2 | Decisions, Risks |
| request types stay hand-written and are checked as they are | `rule_rest_request_body_valid`, the request body item | Appendix, What changes |

## Two statements that need a home before the section shrinks

- "Appliance states that no recording shows, such as a volume with encryption enabled, whose fields stay tolerant until recorded" (line 180) is said nowhere else. Its place is the generated types item, which already says that only a read field can fail a call, or the Evidence gaps introduction.
- "The migration is done only when the client is compliant with the recorded v4 API in all of this, which is stages 3 and 4" (lines 124 to 126) is the draft's third sentence in short; whether the What changes introduction says which stages make the client compliant is to be checked when the section is replaced, and one sentence added there if not.

Rhetoric that goes without a home: "an API that has no future", "one more way to fail", "two sources of truth exist and disagree", which the draft's "have drifted from what the appliance sends" carries.

## If the user decides to replace the section

- Lines 98 to 180 become the draft; the self-anchor line stays.
- The two statements above move first.
- Context is re-read: the README says Context carries the claims and links to the catalogues; the links the Problem statement carried (response types, error messages, logic bugs, crashes, security findings) must be reachable from Context, and are added there if they are not.
- `tools/check_record.py` is run: the duplicated-sentence check and the self-link check are the ones a rewrite of this section can trip.
- The user commits.

## Outcome

Applied on 2026-11-12 at the user's request. The section is the four sentences of the Draft, with one link from "compliance" to What "compliant" means; 85 lines became 5, and the record went from 2,635 to 2,569 lines. The two statements with no other home were placed first: the What changes introduction now says that stage 1 alone reaches v4, stage 2 decodes its replies, and the client is compliant only when stages 3 and 4 are done as well; the generated types item says that appliance states no recording shows keep their fields tolerant until recorded. Context carried links only to the response types and the evidence gaps, so "The REST layer, measured" gained one sentence that names and links the catalogues the old section pointed at: response types, error messages, logic bugs, crashes, security findings. The one other sentence of the record that cites the Problem statement (Option A, "API v3 is deprecated, as the Problem statement says") still holds. Both checkers pass. Not committed; the user commits.
