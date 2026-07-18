# Faithful port checklist (must check)

Master TODO for `internal/languagetool` under [`faithful-port-policy.md`](faithful-port-policy.md).  
**Java under `inspiration/languagetool` is king.** Go ≠ Java → bug in Go. Soft goldens are not proof.

Use this list for implementer sectors, reviewer ACCEPT/REJECT, and human progress tracking.  
Mark items `[x]` only when **reviewer ACCEPT** (or human) confirms for that item — not when implementer self-claims.

**Process reminder:** one Java type per sector · leaves → root · no stubs · freeze outside `internal/languagetool` until human lifts it.

---

## Legend

| Mark | Meaning |
|------|---------|
| `[ ]` | Not done |
| `[~]` | Partial / in progress |
| `[x]` | Done (reviewer/human ACCEPT) |
| **N/A** | Not applicable yet (depends on earlier item) |

**Status column** below is the current known state as of the first deletion wave (`2b1b4ee5`). Update as sectors land.

---

## 0. Process & policy gates

| # | Check | Status |
|---|--------|--------|
| 0.1 | Policy doc is sole law; no competing soft-port policy in repo | [x] `docs/faithful-port-policy.md`, soft-port removed |
| 0.2 | Agents load policy (`AGENTS.md` + nested `internal/languagetool/AGENTS.md`) | [x] |
| 0.3 | Implementer → reviewer ACCEPT loop used on every sector (no self-ACCEPT) | [ ] process discipline |
| 0.4 | Outside-tree freeze holds (no golden/vendor/cmd edits to greenwash) | [~] freeze declared; enforce on each PR |
| 0.5 | Path law: new files under wall map to `inspiration/languagetool/...` | [ ] CI or reviewer always verifies |
| 0.6 | No percentage / soft-score “parity” gates | [~] soft miss-scan removed; watch for new knobs |

---

## 1. Deletion / no invent (fakeness out)

| # | Check | Status |
|---|--------|--------|
| 1.1 | Soft hybrid disambiguator module gone | [x] |
| 1.2 | Soft enable aliases (`SOFT_OPTIONAL` invent expansion) gone | [x] |
| 1.3 | Soft EN typos / soft EN disambiguator invent gone | [x] |
| 1.4 | Soft CJK lexicon invent gone | [x] |
| 1.5 | Soft grammar dir loader (`RegisterSoftGrammarDir`) gone | [x] |
| 1.6 | Soft miss-scan / soft golden harnesses gone | [x] first wave |
| 1.7 | Soft pack loading removed from `configureCoreLT` / server pipeline | [x] first wave |
| 1.8 | **Soft POS surface invent** removed from `pattern_token_matcher.go` (closed-class lists, FreeLing soft, STTS soft, URL soft surface, etc.) | [x] sector: PatternTokenMatcher |
| 1.9 | Soft cover/align hacks removed from `pattern_rule_matcher.go` (CJK surface align, fused prep, hyphen cover, …) | [x] |
| 1.10 | Soft expand / backref invent removed if not Java-equivalent | [ ] |
| 1.11 | Soft chunker approximations removed; only Java `EnglishChunker` / filter logic | [ ] `english_chunker.go` still soft-heavy |
| 1.12 | `SoftRuleMeta` invent removed; rule category/ITS come from Java Rule meta | [~] soft pack invent removed; fallback only for known Java families |
| 1.13 | Soft discovery APIs removed (`Discover*Soft*`, soft typos path, soft disambig XML paths) | [x] soft discover deleted; official discover added |
| 1.14 | Soft picky XML load paths removed from server text checker | [ ] `text_checker_check.go` |
| 1.15 | Demo/map tagger & speller not presented as default engine (or moved outside when freeze lifts) | [ ] `DemoEnglish*`, `LANG_DEMO_SPELLER` |
| 1.16 | Invent multiword lists / embedded soft multiwords gone | [~] SoftEnglish phrase/token invent packs removed from EN core |
| 1.17 | `SoftForeignIgnoreRanges` / soft user-dict naming cleaned to Java twins or removed | [ ] server |
| 1.18 | ZH tokenizer: real HanLP-equivalent or explicit incomplete (no soft POS invent) | [~] per-rune + `x`; needs real twin |
| 1.19 | Dead soft comments / CLI help (`--level` soft packs) cleaned | [ ] |
| 1.20 | No resurrected soft modules (path law + reviewer) | [ ] ongoing |

**Keep (not invent):** Java-named types such as UK `SimpleReplaceSoftRule`, `TestHackHelper` — verify against Java periodically.

---

## 2. Resources (same as Java)

| # | Check | Status |
|---|--------|--------|
| 2.1 | Engine loads **same** grammar/style XML Java loads (not `*-soft.xml` substitutes) | [~] official grammar via LANG_USE_UPSTREAM_GRAMMAR=1 (~5k rules); default core until matcher complete |
| 2.2 | Engine loads **same** `disambiguation.xml` (+ global when Java does) | [ ] |
| 2.3 | Engine loads **same** `multiwords.txt` / multitoken lists | [~] EN multiwords + spelling_global wired |
| 2.4 | Engine loads **same** Morfologik POS dicts (per language) | [~] path wiring exists; coverage incomplete |
| 2.5 | Engine loads **same** speller dicts | [~] EN partial |
| 2.6 | Engine loads **same** SRX / tokenizer resources | [~] |
| 2.7 | OpenNLP / chunker models where Java uses them (or documented missing asset = incomplete, not soft invent) | [ ] |
| 2.8 | FreeLing / language-specific tagger resources as Java | [ ] |
| 2.9 | `spelling_global.txt` and other core resource files as Java | [ ] |
| 2.10 | No soft extract as production input | [~] loaders removed; discover helpers still point at soft paths |

---

## 3. Pipeline twins (leaves → root)

Port/review **one Java type per sector**. Order is dependency order, not completeness of every language.

### 3.A Core analysis stack

| # | Java area (examples) | Check | Status |
|---|----------------------|--------|--------|
| 3.A.1 | Tokenizers (word + sentence / SRX) | 1:1 with Java for supported langs | [~] partial |
| 3.A.2 | Taggers (Morfologik / language-specific) | 1:1 readings | [~] |
| 3.A.3 | `MultiWordChunker` | Official multiwords only | [~] EN hybrid multiword stage |
| 3.A.4 | `XmlRuleDisambiguator` | Full lang XML + global when Java enables | [~] EN hybrid wires official XML+global; loader has rulegroup/`and`/`marker` |
| 3.A.5 | Hybrid disambiguators (EN/FR/DE/NL/PT/CA/ES/…) | Same order as Java hybrids | [~] EN/FR/ES/PT/DE/CA/NL wired to official resources |
| 3.A.6 | Chunker (`EnglishChunker` + filters) | Same BIO/filter as Java | [ ] |
| 3.A.7 | `JLanguageTool` analyze/check wiring | Same stages, mode flags | [~] |
| 3.A.8 | Rule match pipeline (enable/disable, categories, text-level) | Java semantics | [~] |

### 3.B Pattern rules

| # | Check | Status |
|---|--------|--------|
| 3.B.1 | Pattern token matching = Java (no soft POS accept) | [ ] blocked on 1.8 |
| 3.B.2 | Exceptions, skip, regex, inflected, negation | [~] |
| 3.B.3 | Unification | [ ] |
| 3.B.4 | Filters / rule filters as Java | [~] |
| 3.B.5 | Full grammar/style load (entities, includes) | [~] grammar.xml load; style/includes deferred |

### 3.C Rule families (per language, after stack)

| # | Check | Status |
|---|--------|--------|
| 3.C.1 | Core layout rules (whitespace, unpaired, …) | [~] |
| 3.C.2 | Speller rules (Morfologik/Hunspell) | [~] |
| 3.C.3 | Language-specific Java rules (EN AvsAn, DE, …) | [~] by language |
| 3.C.4 | False friends | [~] |
| 3.C.5 | Picky / default-off rules as Java (not soft picky XML) | [ ] |

### 3.D Languages

For **each** language claimed supported, check:

| # | Check | Status |
|---|--------|--------|
| 3.D.0 | Language module structure mirrors Java package | [~] many packages exist |
| 3.D.1 | createDefaultTagger / Tokenizer / Disambiguator / Chunker twins | [ ] per lang |
| 3.D.2 | Default rule registration matches Java | [ ] per lang |
| 3.D.3 | Resources resolve like Java (same paths/names) | [ ] per lang |
| 3.D.4 | End-to-end corpus vs JVM for that lang | [ ] per lang |

Track languages explicitly (add rows as claimed):

- [ ] en (+ variants)
- [ ] de (+ variants)
- [ ] fr (+ variants)
- [ ] es, pt, ca, nl, it, pl, ru, uk, …
- [ ] zh, ja (tokenizers/taggers)
- [ ] others in `corepack.Supported` / Java modules

---

## 4. Parity proof (product bar)

| # | Check | Status |
|---|--------|--------|
| 4.1 | JVM LanguageTool runnable as oracle (same resource root) | [ ] |
| 4.2 | Structured diff: tokens | [ ] |
| 4.3 | Structured diff: POS / lemmas / chunks | [ ] |
| 4.4 | Structured diff: rule matches (id, span) | [ ] |
| 4.5 | Structured diff: suggestions | [ ] |
| 4.6 | Corpus = upstream fixtures + official examples (vendored as-is) | [ ] |
| 4.7 | Regression set for Go≠Java cases (expected side = Java output) | [ ] |
| 4.8 | No editing expected side to match wrong Go | [ ] discipline |
| 4.9 | Twin test audit (`check_lt_test_twins.py` / `twin_audit_test.go`) green or scoped | [ ] |
| 4.10 | Claimed “supported” ⇒ full corpus green (no soft %) | [ ] |

---

## 5. Reviewer checklist (every sector)

Reviewer **must** verify before ACCEPT:

1. [ ] Sector = **one Java type** (or tightly coupled pair with one story).
2. [ ] Java path + type/method identified under `inspiration/languagetool`.
3. [ ] Go path maps under path law.
4. [ ] Control flow / algorithm essentially same (bug-for-bug).
5. [ ] **No** soft/invent/approx in this sector.
6. [ ] Resources used are real LT resources (or fail-closed if missing).
7. [ ] **No stubs** for missing deps.
8. [ ] Dependencies of this type already faithful (leaves → root).
9. [ ] Tests are twin/Java-based, not soft goldens.
10. [ ] Diff does not touch outside wall (while freeze holds).
11. [ ] No metric cheat (skip, %, renamed soft).

**REJECT** if any item fails.

---

## 6. Outside freeze (later — human only)

Do **not** start until freeze lifted:

| # | Check | Status |
|---|--------|--------|
| 6.1 | Soft/demo UX lives **outside** wall only | [ ] |
| 6.2 | Outside must not inject soft stages into faithful engine | [ ] |
| 6.3 | Product CLI (`cmd/`) wired only to faithful API | [ ] |
| 6.4 | Vendor scripts produce real resources / goldens, not soft substitutes for engine | [ ] |
| 6.5 | README/status no longer advertise soft packs as the port | [ ] |

---

## 7. Suggested work order (leaves → root)

1. **Strip remaining invent** (1.8–1.20) — especially pattern soft POS + soft discovery dead code.  
2. **Tokenizer / tagger leaves** per language (3.A.1–3.A.2) with real dicts (2.x).  
3. **MultiWordChunker + XmlRuleDisambiguator + Hybrid** (3.A.3–3.A.5).  
4. **Chunker** (3.A.6).  
5. **Pattern matcher strict Java** (3.B) after invent gone.  
6. **Full rule XML load** (2.1, 3.B.5).  
7. **JVM oracle + corpus** (4.x).  
8. **Language-by-language** 3.D + 3.C until claimed support is green.  
9. **Lift freeze** (6.x) only when wall is trustworthy.

---

## 8. Done definition (whole port)

The whole thing is done when:

- [ ] No invent/soft paths remain inside `internal/languagetool`
- [ ] Pipeline and claimed languages are Java twins (structure + behavior)
- [ ] Same resources as Java for those languages
- [ ] Same corpus → same results as JVM LT (tokens, tags, matches, suggestions)
- [ ] Reviewer process enforced; path law holds
- [ ] Outside freeze can be lifted without reintroducing cheat into the wall

Until then: incomplete is fine; **fake complete is not**.
