#!/usr/bin/env python3
"""Check the names of a decision record against the record and the registry.

Usage: check_names.py [RECORD.md] [ISSUES_DIR]
Defaults: the record ../0002-api-v4-compliance-jdss1.md next to this folder, and
the registry kub-testing/issues/ in a checkout beside this repository. The
README of the record's folder (its kinds table and area list) is read too.

Read only; prints findings, nothing is edited. Sections:
  A. names with a heading that are mentioned in code font or bare (should be links)
  B. names without a heading, and how they appear (code font expected)
  C. names not registered in the registry (a registry row whose meaning starts
     with "superseded" registers an old name; the row above it holds the current one)
  D. current registered names the record never mentions (information)
  E. impl headings of the record against the current rows of design-items.md
  F. areas used by the names against the README's list
  G. kinds in the README's table against the kinds used
  H. superseded names still written in the record
Exit 1 when C, E or H finds something.
"""
import re
import sys
from collections import Counter, defaultdict
from pathlib import Path

ADR = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).resolve().parent.parent / "0002-api-v4-compliance-jdss1.md"
ADR = ADR.resolve()
README = ADR.parent / "README.md"
PLUGIN = ADR.parent.parent.parent
ISSUES = Path(sys.argv[2]) if len(sys.argv) > 2 else PLUGIN.parent / "kub-testing" / "issues"
t = ADR.read_text()
lines = t.split("\n")

NAME = re.compile(r"\b((?:bug|crash|dataloss|security|gap|rule|impl)_[a-z0-9_]+)\b")

headings = {l[5:].strip(): i for i, l in enumerate(lines, 1) if l.startswith("#### ")}
named_headings = {h: i for h, i in headings.items() if NAME.fullmatch(h)}
prose_headings = {h: i for h, i in headings.items() if not NAME.fullmatch(h)}
print(f"#### headings: {len(headings)}; named: {len(named_headings)}; prose: {list(prose_headings)}")

kinds = Counter(h.split("_", 1)[0] for h in named_headings)
print("named headings per kind:", dict(kinds))

# every mention of a name: linked, code font, or bare
mentions = defaultdict(lambda: Counter())
where = defaultdict(list)
for i, l in enumerate(lines, 1):
    if l.startswith("#### "):
        continue
    for m in NAME.finditer(l):
        name = m.group(1)
        s, e = m.start(), m.end()
        before = l[max(0, s - 1):s]
        after = l[e:e + 1]
        linked = l[max(0, s - 1):s] == "[" and l[e:e + 3] == "](#"
        in_anchor = l[max(0, s - 2):s] == "(#" and l[e:e + 1] == ")"
        code = before == "`" and after == "`"
        if in_anchor:
            continue
        kind = "link" if linked else "code" if code else "bare"
        mentions[name][kind] += 1
        if kind != "link":
            where[(name, kind)].append(i)

print("\nA. names with a heading but mentioned in code font or bare somewhere (should be links):")
for name, c in sorted(mentions.items()):
    if name in named_headings and (c["code"] or c["bare"]):
        print(f"  {name}: {dict(c)} at lines {where[(name, 'code')] + where[(name, 'bare')]}")

print("\nB. names without a heading (code font expected), with how they appear:")
no_heading = sorted(n for n in mentions if n not in named_headings)
for name in no_heading:
    c = mentions[name]
    flag = "  <-- linked without a heading" if c["link"] else ("  <-- bare" if c["bare"] else "")
    print(f"  {name}: {dict(c)}{flag}")

print(f"\nC. registration in {ISSUES}/*.md:")
if not ISSUES.is_dir():
    print(f"  registry not found: {ISSUES}")
    sys.exit(2)
current, superseded, by_file = set(), {}, defaultdict(set)
for f in sorted(ISSUES.glob("*.md")):
    if f.name == "README.md":
        continue
    for l in f.read_text().split("\n"):
        cells = [c.strip() for c in l.strip().strip("|").split("|")] if l.startswith("| ") else []
        if len(cells) != 5 or not NAME.fullmatch(cells[1].strip("`")):
            continue
        name, meaning = cells[1].strip("`"), cells[2]
        if meaning.lower().startswith("superseded"):
            superseded[name] = meaning
        else:
            current.add(name)
            by_file[f.name].add(name)
registered = current | set(superseded)
all_names = set(mentions) | set(named_headings)
unreg = sorted(n for n in all_names if n not in registered)
print(f"  names in the record: {len(all_names)}; registered: {len(all_names) - len(unreg)};"
      f" registry rows: {len(current)} current, {len(superseded)} superseded")
for n in unreg:
    print(f"  NOT REGISTERED: {n}  (heading: {n in named_headings})")

print("\nD. current names registered that the record never mentions (information):")
for n in sorted(current - all_names):
    print(f"  {n}")

print("\nE. impl names: record headings vs the current rows of design-items.md:")
di_names = {n for n in by_file["design-items.md"] if n.startswith("impl_")}
rec_impl = {h for h in named_headings if h.startswith("impl_")}
print(f"  record impl headings: {len(rec_impl)}; design-items.md impl names: {len(di_names)}")
print("  in record only:", sorted(rec_impl - di_names))
print("  in design-items only:", sorted(di_names - rec_impl))

print("\nF. areas used by names vs the README's list:")
readme = README.read_text()
m = re.search(r"The areas:\**\s*(.*?`build`)", readme, re.S)
readme_areas = set(a for a in re.findall(r"`([a-z_]+)`", m.group(0))) if m else set()
used_areas = Counter()
for n in all_names:
    rest = n.split("_", 1)[1]
    area = None
    for a in sorted(readme_areas, key=len, reverse=True):
        if rest.startswith(a + "_"):
            area = a
            break
    used_areas[area or "?" + rest.split("_")[0]] += 1
print("  README areas:", sorted(readme_areas))
print("  used:", dict(used_areas))

print("\nG. kinds in the README table vs kinds used:")
readme_kinds = set(re.findall(r"^\s*\| `([a-z]+)` \|", readme, re.M))
print("  README kinds:", sorted(readme_kinds), " used:", sorted({n.split('_', 1)[0] for n in all_names}))

print("\nH. superseded names still written in the record (should be none):")
stale = sorted(n for n in all_names if n in superseded and n not in current)
for n in stale:
    print(f"  STALE: {n}: {superseded[n]}")
if not stale:
    print("  none")

sys.exit(1 if unreg or stale or rec_impl != di_names else 0)
