#!/usr/bin/env python3
"""Run the README's "Checks before a review" over a decision record.

Usage: check_record.py [RECORD.md]
Default record: ../0002-api-v4-compliance-jdss1.md, next to this folder.

Checks: every ](#anchor) resolves to a heading; every ## and ### heading has
its self-anchor line; the Table of Contents matches the ## and ### headings;
table captions are numbered in order of appearance and the List of Tables
matches them; risks, priorities, key observations, testing steps and stages
are numbered 1..n without gaps; open questions (bullets with the number in bold)
have no duplicates and no overlap with the Decisions table; no forbidden
phrasing ("this design", "this record", "design NNNN", ...); every linked
name has a #### heading; relative file links exist; no horizontal rules; no
sentence longer than 60 characters appears twice outside tables; no link to
the section it sits in.

Exit status 1 when a check fails. The lines printed before PROBLEMS are
information, not failures: the numbered lists found, the decided questions,
the letter-and-number tokens outside the ID sources table (other documents'
numbering cited as theirs, such as T11, D3, F28, and line references such as
L84 are expected there), the Key observation references, and the counts.
"""
import re
import sys
from collections import Counter
from pathlib import Path

ADR = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).resolve().parent.parent / "0002-api-v4-compliance-jdss1.md"
t = ADR.read_text()
lines = t.split("\n")
problems = []


def anchor(h):
    a = h.strip().lower()
    a = re.sub(r"[^\w\s-]", "", a)
    return re.sub(r"\s", "-", a)


headings = [(i + 1, m.group(1), m.group(2)) for i, l in enumerate(lines)
            if (m := re.match(r"^(#{1,4}) (.+)$", l))]
anchors = Counter(anchor(h) for _, _, h in headings)
dups = [a for a, n in anchors.items() if n > 1]
if dups:
    problems.append(f"duplicate anchors: {dups}")

# every ](#anchor) resolves
for i, l in enumerate(lines, 1):
    if l.startswith("    ") or l.startswith("```"):
        continue
    for m in re.finditer(r"\]\(#([^)]+)\)", l):
        if m.group(1) not in anchors:
            problems.append(f"line {i}: bad anchor #{m.group(1)}")

# self-anchor lines under ## and ###
for n, (i, level, h) in enumerate(headings):
    if level in ("##", "###") and i > 1:
        nxt = lines[i] if i < len(lines) else ""
        if not re.match(r"^\[[^\]]+\]\(#" + re.escape(anchor(h)) + r"\)$", nxt):
            problems.append(f"line {i}: heading '{h}' lacks its self-anchor line (next: {nxt[:60]!r})")

# ToC entries vs headings
toc_start = next(i for i, _, h in headings if h == "Table of Contents")
toc_end = next(i for i, _, h in headings if h == "List of Tables")
toc = [m.group(1) for l in lines[toc_start:toc_end - 1] for m in [re.search(r"\]\(#([^)]+)\)", l)] if m]
h23 = [anchor(h) for i, level, h in headings if level in ("##", "###") and i > toc_start]
missing = [a for a in h23 if a not in toc]
extra = [a for a in toc if a not in h23 and a != "table-of-contents"]
if missing:
    problems.append(f"headings not in ToC: {missing}")
if extra:
    problems.append(f"ToC entries without heading: {extra}")

# tables: captions numbered in order, LoT matches
caps = [(i, int(m.group(1)), m.group(2), m.group(3)) for i, l in enumerate(lines, 1)
        if (m := re.match(r"^\*\*Table (\d+) — (.+?)\*\* · `(tbl_\w+)`", l))]
nums = [n for _, n, _, _ in caps]
if nums != list(range(1, len(nums) + 1)):
    problems.append(f"table numbers: {nums}")
lot_end = next(i for i, _, h in headings if h == "Metadata")
lot = [(int(m.group(1)), m.group(2), m.group(3)) for l in lines[toc_end:lot_end]
       if (m := re.match(r"^(\d+)\. `(tbl_\w+)` — \*\*(.+?)\*\*", l))]
if [(n, tag, title) for _, n, title, tag in caps] != lot:
    problems.append(f"List of Tables differs from captions:\n  caps={[(n, tag) for _, n, _, tag in caps]}\n  lot={[(n, tag) for n, tag, _ in lot]}")
# body references to Table n stay within range
for i, l in enumerate(lines, 1):
    for m in re.finditer(r"\bTable (\d+)\b", l):
        if int(m.group(1)) > len(caps):
            problems.append(f"line {i}: reference to Table {m.group(1)} beyond {len(caps)}")


def numbered(section, stop):
    s = next(i for i, _, h in headings if h == section)
    e = next(i for i, _, h in headings if h == stop)
    # "1. " for lists without gaps; "- **30. " for a list with gaps (README: renderers renumber ordered lists)
    return [int(m.group(1) or m.group(2)) for l in lines[s:e]
            if (m := re.match(r"^(?:(\d+)\. |- \*\*(\d+)\. )", l))]


for sec, stop, contiguous in (("Risks", "Affected systems and files", True),
                              ("Priorities", "Options considered", True),
                              ("Key observations", "Priorities", True),
                              ("Testing", "Open questions", True),
                              ("Open questions", "Decisions", False)):
    ns = numbered(sec, stop)
    if contiguous and ns != list(range(1, len(ns) + 1)):
        problems.append(f"{sec}: numbering {ns}")
    if len(ns) != len(set(ns)):
        problems.append(f"{sec}: duplicate numbers {ns}")
    print(f"{sec}: {ns if not contiguous else f'1..{len(ns)}'}")

# stages: "### Stage n: ..." headings numbered 1..n in order
stages = [int(m.group(1)) for _, lvl, h in headings if lvl == "###" and (m := re.match(r"Stage (\d+)\b", h))]
if stages != list(range(1, len(stages) + 1)):
    problems.append(f"Stages: numbering {stages}")
print(f"Stages: 1..{len(stages)}")

# decisions table numbers
dec_s = next(i for i, _, h in headings if h == "Decisions")
dec_e = next(i for i, _, h in headings if h == "Relationship to other designs")
dec = [int(m.group(1)) for l in lines[dec_s:dec_e] if (m := re.match(r"^\| (\d+) \|", l))]
print(f"Decisions: {dec}")
openq = numbered("Open questions", "Decisions")
if set(dec) & set(openq):
    problems.append(f"questions both open and decided: {set(dec) & set(openq)}")

# bare letter-and-number IDs outside the ID sources table (information)
id_s = next(i for i, _, h in headings if h == "ID sources")
id_e = next(i for i, _, h in headings if h == "Terms")
bare = Counter()
for i, l in enumerate(lines, 1):
    if id_s <= i < id_e:
        continue
    for m in re.finditer(r"(?<![\w/`.-])([A-Z]{1,4}\d{1,3}(?:\.\d+)?)(?![\w-])", l):
        tok = m.group(1)
        if re.match(r"^(A|B)\d$", tok) and "Option" in l or re.match(r"^(A|B)\d$", tok) and re.search(r"\b(choose|option|options)\b", l):
            continue  # option labels
        bare[tok] += 1
print("bare IDs:", dict(bare))

# wording
# the record must not call itself by its number; another record or design is cited by its number
own = (m.group(1) if (m := re.match(r"(\d{4})-", ADR.name)) else r"\d{4}")
for w in (r"this design", r"this record", r"The design", r"the design's", rf"design {own}", rf"ADR {own}", r"this ADR", r"earlier text", r"earlier list", r"as listed above"):
    k = len(re.findall(w + r"\b", t, flags=re.I))
    if k:
        problems.append(f"wording {w!r}: {k}")
print("this document:", len(re.findall(r"this document\b", t, flags=re.I)))

# linked names have headings; code-font names without heading are allowed
names = set(h for _, lvl, h in headings if lvl == "####")
for i, l in enumerate(lines, 1):
    for m in re.finditer(r"\[([a-z]+_[a-z0-9_]+)\]\(#([a-z0-9_]+)\)", l):
        if m.group(1) != m.group(2) or m.group(1) not in names:
            problems.append(f"line {i}: linked name without heading {m.group(1)}")

# key observation references (information)
for i, l in enumerate(lines, 1):
    for m in re.finditer(r"Key observations? (\d+)", l):
        print(f"line {i}: key observation {m.group(1)}: {l.strip()[:90]}")

# horizontal rules
hr = [i for i, l in enumerate(lines, 1) if l == "---"]
if hr:
    problems.append(f"horizontal rules at {hr}")

# relative file links exist
for i, l in enumerate(lines, 1):
    for m in re.finditer(r"\]\((\.\.?/[^)#]+)(#[^)]*)?\)", l):
        p = (ADR.parent / m.group(1)).resolve()
        if not p.exists():
            problems.append(f"line {i}: missing file {m.group(1)}")

# duplicated sentences (longer than 60 chars), outside tables: two catalogue
# entries may legitimately share a cell such as "Finding: [name](#name);"
body = " ".join(l for l in lines if not l.startswith("|"))
sent = Counter()
for s in re.split(r"(?<=[.;])\s+", re.sub(r"\s+", " ", body)):
    s = s.strip()
    if len(s) > 60 and not s.startswith("["):
        sent[s] += 1
dupsent = [s for s, n in sent.items() if n > 1]
if dupsent:
    problems.append("duplicated sentences:\n  " + "\n  ".join(d[:120] for d in dupsent))

# self-links: a link to the section it sits in
cur = None
for i, l in enumerate(lines, 1):
    if (m := re.match(r"^(#{2,3}) (.+)$", l)):
        cur = anchor(m.group(2))
        continue
    if re.match(r"^\[[^\]]+\]\(#" + (re.escape(cur) if cur else "x") + r"\)$", l):
        continue  # the self-anchor line itself
    for m in re.finditer(r"\]\(#([^)]+)\)", l):
        if m.group(1) == cur and cur not in ("risks", "open-questions", "testing", "appendix", "what-changes", "decisions", "evidence-gaps", "key-observations"):
            problems.append(f"line {i}: links to its own section #{cur}")

print(f"lines: {len(lines)}, headings: {len(headings)}, tables: {len(caps)}")
if problems:
    print("PROBLEMS:")
    print("\n".join(problems))
    sys.exit(1)
print("all checks pass")
