# Vendored LanguageTool testdata

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
