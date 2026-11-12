# ADR 0002: sixth consistency review

Date: 2026-11-11, night, after the Testing section was rewritten to the README's rule and six items were changed to say what must hold. Reviewed: `docs/adr/0002-api-v4-support-jdss1.md` (2,635 lines) against `docs/adr/README.md`. Method: the 18 hunks of tonight's change re-read in their surrounding text; every "step n" reference outside the Testing section checked against the new section; the new claims about the lab tooling checked against `kub-testing/tests/csi-sanity/dev_cycle.py`, the deploy manifests and kub-testing's own rules; the three checkers re-run. Report only; nothing in the record edited.

## Verdict

The rewrite is consistent with the record and with the lab. Every step reference still points at a step that means what the reference says; `dev_cycle.py all --compare` exists and takes a cycle folder; the config Secret is `jdss-controller-cfg` in all three controller manifests; the cycle folder holds `summary.txt` and a `plugin.log` per preset, as the steps say; the user runs every cluster command on kub-gate, as the intro says. The six items read as conditions and no longer prescribe tests, and the checkers are green. What is left is three wording points, none of which changes a decision.

## Low

1. **"whose texts the items fix"** (step 4, first pass condition). The required-field error's shape is fixed by Table 2 ("an error that names the field"); the conversion error of `impl_plugin_rest_return_parse_errors` is not: the item says only that the error "would be returned instead" of swallowed. Either the item says the error names the field and the value it could not parse, or step 4 says "whose shape the items fix".
2. **"the request log" and `plugin.log` are not in Terms.** Steps 4 to 6 search "the request log" and `plugin.log`; neither is defined. The request log is the Debug line per REST request (method, path with its query string, status) that `impl_plugin_rest_no_request_logging` keeps; `plugin.log` is the controller log a cycle saves per preset. Step 6 is a manual test outside a cycle, so there "the controller's log" is the right phrase. Two rows in Terms and one phrase.
3. **Risk 7's "steps 1-3"**: step 3 is now a pointer to step 5, so the checks that do not need the lab are steps 1 and 2. One phrase.

## Carried over, still for the user

- From the testing review: the certificate error check, and whether the lab checks become kub-testing rows.
- From the fourth and fifth reviews: the `project-status.md` citation question; the definition behind "36 call functions"; entries for the two secrets findings; the README rule for prose `####` headings; names for the `DeleteClone` caller and the cleanup fix; whether every fix in What changes must be a named item; open questions 22, 25, 30 to 33; the approver; acceptance; the stray backup design file; the YAML tab defect to report.

## Verified and found consistent

- The README's checks: anchors, self-anchor lines, Table of Contents and List of Tables, risks 1–22, priorities 1–7, key observations 1–8, steps 1–10, tables 1–10, open questions 22, 25, 30–33 as bullets, decisions 1–21, 23, 24, 26–29; no forbidden phrasing; no bare old ID outside its places; 90 names registered, 16 `impl_` headings matching `design-items.md`, every name with a heading linked at every mention.
- Step references outside Testing: step 1 (the first commit), step 4 (the Problem statement's reference cycle, the `per_page` paging change, the Relationship bullet on 0003's D3), step 5 (the success criterion, the cleanup bullet, Table 3's note), steps 4 and 5 (risk 15, Affected systems), steps 6 and 7 (risk 8, How to close the gaps), step 9 (risk 4): each names what the new step says.
- Tonight's claims: `dev_cycle.py all` takes `--compare DIR`, "cycle folder to compare with (default: previous cycle)", and `summary.txt` holds the changes; the TLS handshake prerequisite repeats risk 10's guard with a link; `GET pools/<pool>/san/iscsi/targets/<name>` answers 404 on jdss1 v4 in the matrix; the seven T11 checks exist; `issues/test-findings.md` is where a new F name is registered; the decisions and risks the "what must hold" sentences restate (204 as the empty result, 202 as an error naming the status, the setup error on every call, `https` fixed) say the same in their items and in Table 8.
- The fixtures item keeps questions 4 and 19 and risk 5's rules, and its list of failures the fixtures show on the current types is the list the Response types and Error messages catalogues hold.
