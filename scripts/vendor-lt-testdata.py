#!/usr/bin/env python3
"""Vendor LanguageTool resources into testdata/upstream and derive soft packs.

Source of truth: inspiration/languagetool (git submodule).
Does NOT invent rules or golden cases — only copies and filters upstream XML
to the soft pattern-loader subset (plain surface <token> sequences).

Usage:
  python3 scripts/vendor-lt-testdata.py
  python3 scripts/vendor-lt-testdata.py --langs en,de,fr
"""
from __future__ import annotations

import argparse
import json
import re
import shutil
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
LT = ROOT / "inspiration" / "languagetool"
OUT = ROOT / "testdata" / "upstream"
SOFT_OUT = ROOT / "testdata" / "grammar"
DIS_OUT = ROOT / "testdata" / "disambiguation"
GOLDEN_OUT = ROOT / "testdata" / "upstream" / "goldens"

RE_ENTITY = re.compile(
    r'<!ENTITY\s+([A-Za-z_][\w.-]*)\s+("([^"]*)"|\'([^\']*)\')\s*>'
)
PREDEF = {"amp", "lt", "gt", "quot", "apos"}


def die(msg: str, code: int = 1) -> None:
    print(f"error: {msg}", file=sys.stderr)
    raise SystemExit(code)


def local(tag: str | None) -> str:
    if not tag:
        return ""
    return tag.split("}")[-1]


def expand_entities(raw: str) -> str:
    entities: dict[str, str] = {}
    for m in RE_ENTITY.finditer(raw):
        name = m.group(1)
        val = m.group(3) if m.group(3) is not None else m.group(4)
        entities[name] = val

    def expand(s: str, depth: int = 0) -> str:
        if depth > 30:
            return s

        def repl(m: re.Match[str]) -> str:
            n = m.group(1)
            if n in PREDEF:
                return m.group(0)
            if n in entities:
                return expand(entities[n], depth + 1)
            # Upstream XML sometimes references entities from external DTD
            # includes; drop unresolved ones so soft extract can continue.
            return ""

        return re.sub(r"&([A-Za-z_][\w.-]*);", repl, s)

    # expand entity values for nested refs
    for _ in range(8):
        changed = False
        for k, v in list(entities.items()):
            nv = expand(v)
            if nv != v:
                entities[k] = nv
                changed = True
        if not changed:
            break

    if "<!DOCTYPE" in raw:
        i = raw.index("<!DOCTYPE")
        j = raw.find("]>", i)
        if j < 0:
            die("unclosed DOCTYPE")
        raw = raw[:i] + raw[j + 2 :]
    raw = re.sub(r"<\?xml-stylesheet[^?]*\?>", "", raw)
    return expand(raw)


def parse_rules_xml(path: Path) -> ET.Element:
    raw = path.read_text(encoding="utf-8", errors="replace")
    raw = expand_entities(raw)
    try:
        return ET.fromstring(raw.encode("utf-8"))
    except ET.ParseError as e:
        raise RuntimeError(f"parse {path}: {e}") from e


# Soft PatternRuleLoader supports these token attributes (see pattern_rule_loader.go).
SOFT_TOKEN_ATTRS = {
    "regexp",
    "case_sensitive",
    "negate",
    "inflected",
    "min",
    "max",
    "skip",
    "postag",
    "postag_regexp",
}


def token_is_soft(tok: ET.Element) -> bool:
    """True if token is loadable by the soft Go pattern loader."""
    if local(tok.tag) != "token":
        return False
    # Allow simple <exception> children only (no and/or/unify/phraseref).
    for child in list(tok):
        if local(child.tag) != "exception":
            return False
        if list(child):
            return False
        for k in child.attrib:
            if k not in ("regexp", "negate", "case_sensitive", "inflected", "postag", "postag_regexp"):
                return False
        if (child.get("negate") or "").lower() == "yes":
            return False  # soft loader skips negate exceptions
        ex = (child.text or "").strip()
        if not ex or "&" in ex:
            return False
    for k in tok.attrib:
        if k not in SOFT_TOKEN_ATTRS:
            return False
    text = (tok.text or "").strip()
    has_postag = bool((tok.get("postag") or "").strip())
    if not text and not has_postag:
        return False
    if "&" in text:
        return False
    return True


def serialize_token(tok: ET.Element) -> dict:
    d: dict = {"text": (tok.text or "").strip()}
    for k in SOFT_TOKEN_ATTRS:
        v = tok.get(k)
        if v is not None and str(v).strip() != "":
            d[k] = v
    excs = []
    for child in tok:
        if local(child.tag) != "exception":
            continue
        e = {"text": (child.text or "").strip()}
        for k in ("regexp", "case_sensitive"):
            v = child.get(k)
            if v is not None and str(v).strip() != "":
                e[k] = v
        if e["text"]:
            excs.append(e)
            break  # soft loader keeps first exception only
    if excs:
        d["exceptions"] = excs
    return d


def collapse_or(or_el: ET.Element) -> dict | None:
    """Collapse a simple <or> of plain surface tokens into one soft regexp token."""
    alts: list[str] = []
    for t in or_el:
        if local(t.tag) != "token":
            return None
        # plain surface only inside or (no attrs/exceptions/postag)
        if list(t) or t.attrib:
            return None
        s = (t.text or "").strip()
        if not s or "&" in s:
            return None
        alts.append(s)
    if len(alts) < 2:
        return None
    # Escape for RE; join as non-capturing alternation
    body = "|".join(re.escape(a) for a in alts)
    return {"text": body, "regexp": "yes"}


def pattern_is_simple(pattern: ET.Element) -> list[dict] | None:
    toks: list[dict] = []

    def add_child(child: ET.Element) -> bool:
        tag = local(child.tag)
        if tag == "token":
            if not token_is_soft(child):
                return False
            toks.append(serialize_token(child))
            return True
        if tag == "or":
            collapsed = collapse_or(child)
            if collapsed is None:
                return False
            toks.append(collapsed)
            return True
        return False

    for child in pattern:
        tag = local(child.tag)
        if tag == "marker":
            for t in child:
                if not add_child(t):
                    return None
            continue
        if not add_child(child):
            return None
    if not toks:
        return None
    return toks


def strip_markers(s: str) -> str:
    s = re.sub(r"</?marker>", "", s)
    return " ".join(s.split())


def example_text(ex: ET.Element) -> str:
    # prefer full string with markers stripped
    parts: list[str] = []
    if ex.text:
        parts.append(ex.text)
    for c in ex:
        if c.text:
            parts.append(c.text)
        if c.tail:
            parts.append(c.tail)
    raw = "".join(parts) if parts else ("".join(ex.itertext()) if ex is not None else "")
    return strip_markers(raw)


def extract_simple_rules(root: ET.Element, source: str) -> tuple[list[dict], list[dict]]:
    """Return (soft_rules, golden_cases) from upstream rules root."""
    rules_out: list[dict] = []
    goldens: list[dict] = []
    seen_ids: set[str] = set()

    def walk_rule(el: ET.Element, cat_id: str, cat_name: str) -> None:
        rid = el.get("id") or ""
        if not rid or rid in seen_ids:
            return
        pattern = None
        message = ""
        short = ""
        for c in el:
            t = local(c.tag)
            if t == "pattern" and pattern is None:
                pattern = c
            elif t == "message":
                message = "".join(c.itertext()).strip()
                # keep suggestion tags in serialized soft XML separately
            elif t == "short":
                short = "".join(c.itertext()).strip()
        if pattern is None:
            return
        toks = pattern_is_simple(pattern)
        if toks is None:
            return
        # message element for soft XML: keep suggestion children if present
        msg_el = None
        for c in el:
            if local(c.tag) == "message":
                msg_el = c
                break
        msg_xml = ""
        if msg_el is not None:
            # rebuild simple message with suggestion tags
            chunks: list[str] = []
            if msg_el.text:
                chunks.append(msg_el.text)
            for ch in msg_el:
                if local(ch.tag) == "suggestion":
                    chunks.append("<suggestion>" + "".join(ch.itertext()) + "</suggestion>")
                else:
                    chunks.append("".join(ch.itertext()))
                if ch.tail:
                    chunks.append(ch.tail)
            msg_xml = "".join(chunks).strip()
        else:
            msg_xml = message

        examples = []
        for c in el:
            if local(c.tag) != "example":
                continue
            corr = c.get("correction")
            if corr is None:
                continue  # correct example — skip for positive golden
            text = example_text(c)
            if not text:
                continue
            # first correction alternative
            sug = corr.split("|")[0].strip()
            examples.append({"text": text, "suggestion": sug})

        if not examples:
            return

        seen_ids.add(rid)
        soft_id = rid if rid.startswith("EN_") or "_" in rid else rid
        rules_out.append(
            {
                "id": soft_id,
                "name": el.get("name") or soft_id,
                "category_id": cat_id or "GRAMMAR",
                "category_name": cat_name or "Grammar",
                "tokens": toks,
                "message": msg_xml or f"Did you mean a correction for {soft_id}?",
                "short": short or soft_id,
                "source": source,
            }
        )
        for ex in examples:
            goldens.append(
                {
                    "rule": soft_id,
                    "text": ex["text"],
                    "suggestion": ex["suggestion"],
                    "source": source,
                }
            )

    # categories
    for cat in root:
        if local(cat.tag) != "category":
            continue
        cid, cname = cat.get("id") or "GRAMMAR", cat.get("name") or "Grammar"
        for child in cat:
            t = local(child.tag)
            if t == "rule":
                walk_rule(child, cid, cname)
            elif t == "rulegroup":
                for r in child:
                    if local(r.tag) == "rule":
                        # prefer rule id; fall back to group id + index handled by walk
                        if not r.get("id") and child.get("id"):
                            # anonymous rule in group — skip (needs synthetic id)
                            continue
                        walk_rule(r, cid, cname)

    # top-level rules
    for child in root:
        if local(child.tag) == "rule":
            walk_rule(child, "GRAMMAR", "Grammar")

    return rules_out, goldens


def write_soft_xml(path: Path, lang: str, rules: list[dict]) -> None:
    # group by category
    cats: dict[tuple[str, str], list[dict]] = {}
    for r in rules:
        key = (r["category_id"], r["category_name"])
        cats.setdefault(key, []).append(r)

    lines = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        f"<!-- GENERATED by scripts/vendor-lt-testdata.py — do not invent rules. -->",
        f"<!-- Source: upstream LanguageTool simple token patterns only. -->",
        f'<rules lang="{lang}">',
    ]
    for (cid, cname), rs in cats.items():
        lines.append(f'  <category id="{xml_esc(cid)}" name="{xml_esc(cname)}">')
        for r in rs:
            lines.append(f'    <rule id="{xml_esc(r["id"])}" name="{xml_esc(r["name"])}">')
            lines.append("      <pattern>")
            for t in r["tokens"]:
                if isinstance(t, str):
                    lines.append(f"        <token>{xml_esc(t)}</token>")
                    continue
                attrs = []
                for k in (
                    "regexp",
                    "case_sensitive",
                    "negate",
                    "min",
                    "max",
                    "skip",
                    "postag",
                    "postag_regexp",
                ):
                    if k in t and t[k] is not None and str(t[k]) != "":
                        attrs.append(f'{k}="{xml_esc(str(t[k]))}"')
                attr_s = (" " + " ".join(attrs)) if attrs else ""
                body = xml_esc(t.get("text") or "")
                excs = t.get("exceptions") or []
                if not excs:
                    lines.append(f"        <token{attr_s}>{body}</token>")
                else:
                    lines.append(f"        <token{attr_s}>{body}")
                    for e in excs:
                        ea = []
                        for k in ("regexp", "case_sensitive"):
                            if k in e and e[k]:
                                ea.append(f'{k}="{xml_esc(str(e[k]))}"')
                        eas = (" " + " ".join(ea)) if ea else ""
                        lines.append(f"          <exception{eas}>{xml_esc(e.get('text') or '')}</exception>")
                    lines.append("        </token>")
            lines.append("      </pattern>")
            lines.append(f"      <message>{r['message']}</message>")  # may contain <suggestion>
            lines.append(f"      <short>{xml_esc(r['short'])}</short>")
            lines.append("    </rule>")
        lines.append("  </category>")
    lines.append("</rules>")
    lines.append("")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines), encoding="utf-8")


def xml_esc(s: str) -> str:
    return (
        s.replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
        .replace('"', "&quot;")
    )


def copy_file(src: Path, dst: Path) -> None:
    dst.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dst)
    print(f"  copy {src.relative_to(ROOT)} -> {dst.relative_to(ROOT)}")


def vendor_lang(lang: str) -> dict:
    stats = {"lang": lang, "rules": 0, "goldens": 0, "copied": 0}
    mod = LT / "languagetool-language-modules" / lang
    if not mod.is_dir():
        print(f"skip {lang}: no module")
        return stats

    rules_base = mod / "src/main/resources/org/languagetool/rules" / lang
    res_base = mod / "src/main/resources/org/languagetool/resource" / lang

    # raw copies
    for rel in [
        "grammar.xml",
        "style.xml",
    ]:
        src = rules_base / rel
        if src.is_file():
            copy_file(src, OUT / lang / "rules" / rel)
            stats["copied"] += 1

    # regional grammar packs
    if rules_base.is_dir():
        for p in sorted(rules_base.glob("*/grammar.xml")):
            copy_file(p, OUT / lang / "rules" / p.parent.name / "grammar.xml")
            stats["copied"] += 1

    for rel in ["disambiguation.xml", "multiwords.txt"]:
        src = res_base / rel
        if src.is_file():
            copy_file(src, OUT / lang / "resource" / rel)
            stats["copied"] += 1

    # derive soft pack + goldens from main grammar (+ style) and regional packs
    all_rules: list[dict] = []
    all_goldens: list[dict] = []
    extract_paths: list[Path] = []
    for name in ("grammar.xml", "style.xml"):
        src = rules_base / name
        if src.is_file():
            extract_paths.append(src)
    if rules_base.is_dir():
        for p in sorted(rules_base.glob("*/grammar.xml")):
            extract_paths.append(p)
    for src in extract_paths:
        print(f"  extract simple patterns from {src.relative_to(ROOT)}")
        try:
            root = parse_rules_xml(src)
        except RuntimeError as e:
            print(f"  WARN skip {src.relative_to(ROOT)}: {e}")
            continue
        rules, goldens = extract_simple_rules(root, str(src.relative_to(LT)))
        all_rules.extend(rules)
        all_goldens.extend(goldens)

    # de-dupe rules by id (first wins)
    seen: set[str] = set()
    deduped = []
    for r in all_rules:
        if r["id"] in seen:
            continue
        seen.add(r["id"])
        deduped.append(r)
    all_rules = deduped

    def is_plain(r: dict) -> bool:
        for t in r["tokens"]:
            if isinstance(t, str):
                continue
            if any(k in t for k in ("regexp", "postag", "postag_regexp", "negate", "inflected", "min", "max", "skip", "exceptions")):
                return False
        return True

    # Prefer plain surface rules first (soft engine is most reliable here).
    all_rules.sort(key=lambda r: (0 if is_plain(r) else 1, r["id"]))
    plain_ids = {r["id"] for r in all_rules if is_plain(r)}

    # goldens only for kept rules; plain examples first for sample suites
    keep = {r["id"] for r in all_rules}
    all_goldens = [g for g in all_goldens if g["rule"] in keep]
    all_goldens.sort(key=lambda g: (0 if g["rule"] in plain_ids else 1, g["rule"], g["text"]))

    if all_rules:
        soft_path = OUT / lang / f"{lang}-from-upstream-soft.xml"
        write_soft_xml(soft_path, lang, all_rules)
        # also install as optional soft pack used by engine when named *-soft
        # Do NOT overwrite hand soft packs: write parallel file
        install = SOFT_OUT / f"{lang}-upstream-soft.xml"
        write_soft_xml(install, lang, all_rules)
        print(f"  soft rules: {len(all_rules)} -> {install.relative_to(ROOT)}")
        stats["rules"] = len(all_rules)

    if all_goldens:
        GOLDEN_OUT.mkdir(parents=True, exist_ok=True)
        gpath = GOLDEN_OUT / f"{lang}-examples.json"
        gpath.write_text(
            json.dumps(
                {
                    "source": "inspiration/languagetool",
                    "note": "Generated from upstream <example correction> only; do not invent.",
                    "language": lang,
                    "cases": all_goldens,
                },
                indent=2,
                ensure_ascii=False,
            )
            + "\n",
            encoding="utf-8",
        )
        print(f"  goldens: {len(all_goldens)} -> {gpath.relative_to(ROOT)}")
        stats["goldens"] = len(all_goldens)

    # multiwords into disambiguation as upstream copy (not invented soft mesa names)
    mw = res_base / "multiwords.txt"
    if mw.is_file() and lang == "en":
        copy_file(mw, DIS_OUT / "en-multiwords-upstream.txt")

    return stats


def vendor_core_fixtures() -> int:
    n = 0
    xx = LT / "languagetool-core/src/test/resources/org/languagetool/rules/xx"
    if xx.is_dir():
        for p in sorted(xx.glob("*.xml")):
            copy_file(p, OUT / "xx" / p.name)
            n += 1
    ff = LT / "languagetool-core/src/main/resources/org/languagetool/rules/false-friends.xml"
    if ff.is_file():
        copy_file(ff, OUT / "false-friends.xml")
        n += 1
    return n


def write_readme() -> None:
    text = """# Vendored LanguageTool testdata

**Source of truth:** `inspiration/languagetool` (upstream submodule).

Regenerate:

```bash
python3 scripts/vendor-lt-testdata.py
```

| Path | Meaning |
|------|---------|
| `testdata/upstream/<lang>/rules/` | Copies of upstream `grammar.xml` / `style.xml` / regional packs |
| `testdata/upstream/<lang>/resource/` | Copies of `disambiguation.xml`, `multiwords.txt` |
| `testdata/upstream/goldens/<lang>-examples.json` | Official `<example correction>` cases only |
| `testdata/grammar/<lang>-upstream-soft.xml` | Soft-loader subset: plain surface token patterns extracted from upstream |

**Policy (SPEC §3.3–3.4):** do not invent kitchen-sink rules or golden strings.
Goldens come from upstream examples. Soft packs are filters of upstream XML,
not original content.

Hand-written `*-soft.xml` packs elsewhere under `testdata/grammar/` are legacy
scaffolding; prefer `*-upstream-soft.xml` and full upstream files going forward.
"""
    (OUT / "README.md").write_text(text, encoding="utf-8")


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--langs",
        default="en",
        help="comma-separated language module codes (default: en)",
    )
    args = ap.parse_args()
    if not LT.is_dir():
        die(f"missing {LT}; git submodule update --init")

    langs = [x.strip() for x in args.langs.split(",") if x.strip()]
    print(f"vendoring from {LT.relative_to(ROOT)}")
    OUT.mkdir(parents=True, exist_ok=True)
    write_readme()
    n_core = vendor_core_fixtures()
    print(f"core fixtures: {n_core} files")
    totals = {"rules": 0, "goldens": 0, "copied": 0}
    for lang in langs:
        print(f"lang {lang}:")
        st = vendor_lang(lang)
        for k in totals:
            totals[k] += st.get(k, 0)
    print(
        f"done: copied={totals['copied']} simple_rules={totals['rules']} "
        f"golden_cases={totals['goldens']}"
    )


if __name__ == "__main__":
    main()
