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
| `[x]` | Done — proven 1:1 with Java (reviewer/human ACCEPT) |
| **N/A** | Not applicable yet (depends on earlier item) |

**Status column** below is the current known state as of the first deletion wave (`2b1b4ee5`). Update as sectors land.

---

## 0. Process & policy gates

| # | Check | Status |
|---|--------|--------|
| 0.1 | Policy doc is sole law; no competing soft-port policy in repo | [x] `docs/faithful-port-policy.md`, soft-port removed |
| 0.2 | Agents load policy (`AGENTS.md` + nested `internal/languagetool/AGENTS.md`) | [x] |
| 0.3 | Implementer → reviewer ACCEPT loop used on every sector (no self-ACCEPT) | [ ] process discipline |
| 0.4 | Outside-tree freeze holds (no golden/vendor/cmd edits to greenwash) | [ ] freeze declared; enforce on each PR |
| 0.5 | Path law: new files under wall map to `inspiration/languagetool/...` | [ ] CI or reviewer always verifies |
| 0.6 | No percentage / soft-score “parity” gates | [ ] soft miss-scan removed; watch for new knobs |

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
| 1.10 | Soft expand / backref invent removed if not Java-equivalent | [ ] formatMatches + multi-lang `*_synth.dict` + suppress_misspelled + setpos getPosTagCorrection |
| 1.11 | Soft chunker approximations removed; only Java `EnglishChunker` / filter logic | [ ] soft invent lists removed; POS→BIO interim until OpenNLP maxent wired |
| 1.12 | `SoftRuleMeta` invent removed; rule category/ITS come from Java Rule meta | [ ] `RuleMeta` (Soft* aliases); fallback only for known Java families |
| 1.13 | Soft discovery APIs removed (`Discover*Soft*`, soft typos path, soft disambig XML paths) | [x] soft discover deleted; official discover added |
| 1.14 | Soft picky XML load paths removed from server text checker | [ ] picky uses official RegisterPickyEnglishRules; pipeline uses official resources |
| 1.15 | Demo/map tagger & speller not presented as default engine (or moved outside when freeze lifts) | [ ] binary speller default (no invent map); demo only under `LANG_DEMO_SPELLER` |
| 1.16 | Invent multiword lists / embedded soft multiwords gone | [ ] SoftEnglish phrase/token invent packs removed from EN core |
| 1.17 | `SoftForeignIgnoreRanges` / soft user-dict naming cleaned to Java twins or removed | [ ] `ForeignScriptIgnoreRanges` + `UserDictionary` (Soft* aliases kept) |
| 1.18 | ZH tokenizer: real HanLP-equivalent or explicit incomplete (no soft POS invent) | [ ] per-rune + `x`; needs real twin |
| 1.19 | Dead soft comments / CLI help (`--level` soft packs) cleaned | [ ] help text + false-friends resolve use official paths; SoftRuleMeta labels remain |
| 1.20 | No resurrected soft modules (path law + reviewer) | [ ] ongoing |

**Keep (not invent):** Java-named types such as UK `SimpleReplaceSoftRule`, `TestHackHelper` — verify against Java periodically.

---

## 2. Resources (same as Java)

| # | Check | Status |
|---|--------|--------|
| 2.1 | Engine loads **same** grammar/style XML Java loads (not `*-soft.xml` substitutes) | [ ] LANG_USE_UPSTREAM_GRAMMAR=1: getRuleFileNames order (grammar/style/custom + variant); filters/dicts wired |
| 2.2 | Engine loads **same** `disambiguation.xml` (+ global when Java does) | [ ] EN/FR/ES/PT/DE/CA/NL hybrids load official XML+global when present |
| 2.3 | Engine loads **same** `multiwords.txt` / multitoken lists | [ ] EN hybrid multiwords; EnglishMultitokenSpeller loads multiwords+spelling_global for MultitokenSpellerFilter |
| 2.4 | Engine loads **same** Morfologik POS dicts (per language) | [ ] EN english.dict for tagger + FindSuggestions desiredPostag; other langs partial |
| 2.5 | Engine loads **same** speller dicts | [ ] EN en_US.dict for speller rule + grammar filter hooks |
| 2.6 | Engine loads **same** SRX / tokenizer resources | [ ] |
| 2.7 | OpenNLP / chunker models where Java uses them (or documented missing asset = incomplete, not soft invent) | [ ] models in third_party/opennlp-models; runtime still POS→BIO interim (no invent OpenNLP) |
| 2.8 | FreeLing / language-specific tagger resources as Java | [ ] |
| 2.9 | `spelling_global.txt` and other core resource files as Java | [ ] |
| 2.10 | No soft extract as production input | [ ] loaders removed; discover helpers still point at soft paths |

---

## 3. Pipeline twins (leaves → root)

Port/review **one Java type per sector**. Order is dependency order, not completeness of every language.

### 3.A Core analysis stack

| # | Java area (examples) | Check | Status |
|---|----------------------|--------|--------|
| 3.A.1 | Tokenizers (word + sentence / SRX) | 1:1 with Java for supported langs | [ ] partial; EN EnglishWordTokenizer twin ACCEPT @ f7c09ec6 (sub-sector 3.A.1-EnglishWordTokenizer); EN EnglishSRXSentenceTokenizerTest twin ACCEPT @ e6904f79 (sub-sector 3.A.1-EnglishSRXSentenceTokenizer; attic/srx RE2); ES SpanishSentenceTokenizerTest twin ACCEPT @ de99e3c8 (sub-sector 3.A.1-SpanishSentenceTokenizer; attic/srx \b); PT PortugueseSRXSentenceTokenizerTest twin ACCEPT @ 54736eb6 (sub-sector 3.A.1-PortugueseSRXSentenceTokenizer); NL DutchSRXSentenceTokenizerTest twin ACCEPT @ 88163de5 (sub-sector 3.A.1-DutchSRXSentenceTokenizer); IT ItalianSRXSentenceTokenizerTest twin ACCEPT @ 03e03ed4 (sub-sector 3.A.1-ItalianSRXSentenceTokenizer); DE GermanSRXSentenceTokenizerTest twin ACCEPT @ d1ca77ae (sub-sector 3.A.1-GermanSRXSentenceTokenizer); PL PolishSentenceTokenizerTest twin ACCEPT @ 7c671115 (sub-sector 3.A.1-PolishSentenceTokenizer); UK UkrainianSRXSentenceTokenizerTest twin ACCEPT @ b34d8d2e (sub-sector 3.A.1-UkrainianSRXSentenceTokenizer; attic/srx \h\v lookaround); RO RomanianSentenceTokenizerTest twin ACCEPT @ 47db63ca (sub-sector 3.A.1-RomanianSentenceTokenizer); FR FrenchSentenceTokenizerTest twin ACCEPT @ fbac8d4b (sub-sector 3.A.1-FrenchSentenceTokenizer); CA CatalanSentenceTokenizerTest twin ACCEPT @ a7d83527 (sub-sector 3.A.1-CatalanSentenceTokenizer; attic/srx exception lookbehind); SK SlovakSentenceTokenizerTest twin ACCEPT @ 106b04f0 (sub-sector 3.A.1-SlovakSentenceTokenizer; dual paragraph modes); DA DanishSRXSentenceTokenizerTest twin ACCEPT @ 9181de3d (sub-sector 3.A.1-DanishSRXSentenceTokenizer); RU RussianSRXSentenceTokenizerTest twin ACCEPT @ e30f75f7 (sub-sector 3.A.1-RussianSRXSentenceTokenizer); SV SwedishSRXSentenceTokenizerTest twin ACCEPT @ fb900fab (sub-sector 3.A.1-SwedishSRXSentenceTokenizer); SR SerbianSRXSentenceTokenizerTest twin ACCEPT @ c3041861 (sub-sector 3.A.1-SerbianSRXSentenceTokenizer); JA JapaneseSRXSentenceTokenizerTest twin ACCEPT @ 2a230eb6 (sub-sector 3.A.1-JapaneseSRXSentenceTokenizer); FA PersianSRXSentenceTokenizerTest twin ACCEPT @ e08f2318 (sub-sector 3.A.1-PersianSRXSentenceTokenizer); AST AsturianSRXSentenceTokenizerTest twin ACCEPT @ 2e475bda (sub-sector 3.A.1-AsturianSRXSentenceTokenizer; dual paragraph modes); AR ArabicSRXSentenceTokenizerTest twin ACCEPT @ f77e3aa1 (sub-sector 3.A.1-ArabicSRXSentenceTokenizer); CRH CrimeanTatarSRXSentenceTokenizerTest twin ACCEPT @ 2f01e490 (sub-sector 3.A.1-CrimeanTatarSRXSentenceTokenizer; singleLineBreaks=true); LT LithuanianSRXSentenceTokenizerTest twin ACCEPT @ c1e8a1d3 (sub-sector 3.A.1-LithuanianSRXSentenceTokenizer); ML MalayalamSRXSentenceTokenizerTest twin ACCEPT @ 45dddca1 (sub-sector 3.A.1-MalayalamSRXSentenceTokenizer); TL TagalogSRXSentenceTokenizerTest twin ACCEPT @ 614f57df (sub-sector 3.A.1-TagalogSRXSentenceTokenizer); NL DutchWordTokenizer twin ACCEPT @ 67a38407 (sub-sector 3.A.1-DutchWordTokenizer); FR FrenchWordTokenizer twin ACCEPT @ 12455839 (sub-sector 3.A.1-FrenchWordTokenizer); PT PortugueseWordTokenizer twin ACCEPT @ 63245687 (sub-sector 3.A.1-PortugueseWordTokenizer; 28/28); CA CatalanWordTokenizer twin ACCEPT @ d9a58178 (sub-sector 3.A.1-CatalanWordTokenizer); BE BelarusianWordTokenizer twin ACCEPT @ 67a38407 (sub-sector 3.A.1-BelarusianWordTokenizer); BR BretonWordTokenizer twin ACCEPT @ 820a52e0 (sub-sector 3.A.1-BretonWordTokenizer); RO RomanianWordTokenizer twin ACCEPT @ 8811b9c8 (sub-sector 3.A.1-RomanianWordTokenizer); EO EsperantoWordTokenizer twin ACCEPT @ b5dda793 (sub-sector 3.A.1-EsperantoWordTokenizer); CRH CrimeanTatarWordTokenizer twin ACCEPT @ b6e81c88 (sub-sector 3.A.1-CrimeanTatarWordTokenizer); PL PolishWordTokenizer twin ACCEPT @ c1772adc (sub-sector 3.A.1-PolishWordTokenizer; real polish.dict); JA JapaneseWordTokenizer twin ACCEPT @ 528837ae (sub-sector 3.A.1-JapaneseWordTokenizer; kagome IPA); core WordTokenizer twin ACCEPT @ 786cbae4 (sub-sector 3.A.1-core-WordTokenizer; 9/9); UK CAP-revisit attempt=1 CAP-blocked @ a12f5e61 (ABBR_DOT_2 BOS); ES SpanishWordTokenizer CAP-revisit ACCEPT @ 32995ed8 (UCC ORDINAL_POINT); EN GoogleStyleWordTokenizer twin ACCEPT @ c37e38b0 (sub-sector 3.A.1-GoogleStyleWordTokenizer); RU RussianWordTokenizer twin ACCEPT @ ce016368 (sub-sector 3.A.1-RussianWordTokenizer; behavior matrix); DE GermanWordTokenizer twin ACCEPT @ 8daf13f0 (sub-sector 3.A.1-GermanWordTokenizer; delims _‚); AR ArabicWordTokenizer twin ACCEPT @ 957a5824 (sub-sector 3.A.1-ArabicWordTokenizer; delims ،؟؛-); FA PersianWordTokenizer twin ACCEPT @ 8a2714aa (sub-sector 3.A.1-PersianWordTokenizer; delims ،؟؛ no hyphen); core SimpleSentenceTokenizer twin ACCEPT @ 17f7a3f4 (sub-sector 3.A.1-SimpleSentenceTokenizer; segment-simple.srx); TL TagalogWordTokenizer twin ACCEPT @ ba9d62fd (sub-sector 3.A.1-TagalogWordTokenizer; delims -); KM KhmerWordTokenizer twin ACCEPT @ 8e70ba8a (sub-sector 3.A.1-KhmerWordTokenizer; U+17D4/U+17D5 delims); ML MalayalamWordTokenizer twin ACCEPT @ 2e1a98ff (sub-sector 3.A.1-MalayalamWordTokenizer; no joinEMailsAndUrls); GL GalicianWordTokenizer twin ACCEPT @ 55a75a7a (sub-sector 3.A.1-GalicianWordTokenizer; SPLIT_CHARS+date/space); EL GreekWordTokenizer twin ACCEPT @ d3e4cae3 (sub-sector 3.A.1-GreekWordTokenizer; JFlex Delim+ό,τι); ZH ChineseSentenceTokenizer twin ACCEPT @ ac47695a (sub-sector 3.A.1-ChineseSentenceTokenizer; HanLP SentencesUtil 1:1) (UK case-fold abbr invent; ES ORDINAL_POINT full UCC \w/\d); DE-compound length UTF-16 |
| 3.A.2 | Taggers (Morfologik / language-specific) | 1:1 readings | [ ] startPos + many word.length() gates UTF-16; EN EnglishTaggerTest myAssert twin ACCEPT @ aa8917ed (sub-sector 3.A.2-EnglishTagger-testTagger); RU RussianTaggerTest myAssert twin ACCEPT @ e92c2e55 (sub-sector 3.A.2-RussianTagger-testTagger; real russian.dict); PL PolishTaggerTest myAssert twin ACCEPT @ 0ce69488 (sub-sector 3.A.2-PolishTagger-testTagger; real polish.dict); IT ItalianTaggerTest myAssert twin ACCEPT @ c0d51e35 (sub-sector 3.A.2-ItalianTagger-testTagger; real italian.dict); SV SwedishTaggerTest myAssert twin ACCEPT @ 6faa3c33 (sub-sector 3.A.2-SwedishTagger-testTagger; real swedish.dict); SK SlovakTaggerTest myAssert twin ACCEPT @ 19d5f806 (sub-sector 3.A.2-SlovakTagger-testTagger; real slovak.dict); RO RomanianTaggerTest myAssert+lemma twin ACCEPT @ b8269ee1 (sub-sector 3.A.2-RomanianTagger-testTagger; real romanian.dict + diacritics); SR Ekavian+JekavianTaggerTest twins ACCEPT @ ba97687d (sub-sector 3.A.2-EkavianTagger-testTagger; real serbian.dict; Morfologik freq-byte strip); GL GalicianTaggerTest myAssert twin ACCEPT @ bc01c476 (sub-sector 3.A.2-GalicianTagger-testTagger; real galician.dict; mente/prefix Tag); AR ArabicTaggerTest myAssert twin ACCEPT @ bf13025f (sub-sector 3.A.2-ArabicTagger-testTagger; real arabic.dict); EO EsperantoTaggerTest myAssert twin ACCEPT @ bb8a48cc (sub-sector 3.A.2-EsperantoTagger-testTagger; rule-based + official manual-tagger/verb lists); JA JapaneseTaggerTest myAssert twin ACCEPT @ b8349ddf (sub-sector 3.A.2-JapaneseTagger-testTagger; Sen-encoded token parse + accepted JA tokenizer); core MorfologikTaggerTest twin ACCEPT @ 2e370532 (sub-sector 3.A.2-MorfologikTagger-testTag; real test.dict); core ManualTaggerTest twin ACCEPT @ c0562949 (sub-sector 3.A.2-ManualTagger-testTag; official de/added.txt); core CombiningTaggerTest twin ACCEPT @ 062d22b6 (sub-sector 3.A.2-CombiningTagger-core-test; official xx added/removed); DA DanishTagger BaseTagger twin ACCEPT @ 622e8e1a (sub-sector 3.A.2-DanishTagger-testTagger; real danish.dict); TL TagalogTagger BaseTagger twin ACCEPT @ 0937a0f2 (sub-sector 3.A.2-TagalogTagger-testTagger; real tagalog.dict); KM KhmerTagger BaseTagger twin ACCEPT @ 18e3da2a (sub-sector 3.A.2-KhmerTagger-testTagger; real khmer.dict); ML MalayalamTagger BaseTagger twin ACCEPT @ 7a852e4a (sub-sector 3.A.2-MalayalamTagger-testTagger; real malayalam.dict); TA TamilTagger BaseTagger twin ACCEPT @ d447b149 (sub-sector 3.A.2-TamilTagger-testTagger; real tamil.dict); BR BretonTagger custom-tag twin ACCEPT @ a41de305 (sub-sector 3.A.2-BretonTagger-testTagger; real breton.dict; -mañ/-se/-hont) |
| 3.A.3 | `MultiWordChunker` | Official multiwords only | [ ] EN hybrid multiword stage; core MultiWordChunkerTest twin ACCEPT @ 3b4841a4 (sub-sector 3.A.3-MultiWordChunker-core-test); PL PolishDisambiguationRuleTest.testChunker MultiWordChunker twin ACCEPT @ 09002c2f (sub-sector 3.A.3-PolishMultiWordChunker-testChunker; official pl/multiwords.txt); SV SwedishDisambiguationRuleTest.testChunker MultiWordChunker twin ACCEPT @ 0bf23c5c (sub-sector 3.A.3-SwedishMultiWordChunker-testChunker; official sv/multiwords.txt); GL MultiWordChunker official multiwords outcome twin ACCEPT @ f7e45200 (sub-sector 3.A.3-GalicianMultiWordChunker-testChunker; official gl/multiwords.txt) |
| 3.A.4 | `XmlRuleDisambiguator` | Full lang XML + global when Java enables | [ ] EN hybrid wires official XML+global; loader has rulegroup/`and`/`marker`; RO RomanianRuleDisambiguatorTest myAssert twin ACCEPT @ 5cccbe9d (sub-sector 3.A.4-RomanianRuleDisambiguator-tests; official ro/disambiguation.xml; useGlobal=false) |
| 3.A.5 | Hybrid disambiguators (EN/FR/DE/NL/PT/CA/ES/…) | Same order as Java hybrids | [ ] EN/FR/ES/PT/DE/CA/NL wired; EN testChunker myAssert twin ACCEPT @ 251a9820 (sub-sector 3.A.5-en-hybrid-disambig-testChunker) |
| 3.A.6 | Chunker (`EnglishChunker` + filters) | Same BIO/filter as Java | [ ] partial; EN EnglishChunkFilter twin ACCEPT @ 859b2393; EN EnglishChunker OpenNLP twin ACCEPT @ a59fd76b (sub-sector 3.A.6-EnglishChunker); other-lang chunkers remain |
| 3.A.7 | `JLanguageTool` analyze/check wiring | Same stages, mode flags | [ ] |
| 3.A.8 | Rule match pipeline (enable/disable, categories, text-level) | Java semantics | [ ] |

### 3.B Pattern rules

| # | Check | Status |
|---|--------|--------|
| 3.B.1 | Pattern token matching = Java (no soft POS accept) | [ ] blocked on 1.8 |
| 3.B.2 | Exceptions, skip, regex, inflected, negation, `<or>`/`phraseref`/setpos/and/raw_pos | [ ] + exception negate/negate_pos XOR; multi-exception lists; sticky `prevMatched` for scope=next+skip (Java field) |
| 3.B.3 | Unification | [ ] Loader parses `<unification>`+`<unify>`; matcher ports testUnification; UniFeatures/Last/Neg/Neutral wired |
| 3.B.4 | Filters / rule filters as Java | [ ] AdvancedSynthesizer+postagReplace (fail-closed); CA/NL dates; DE NumberInWord; CA Adjust* remaining |
| 3.B.5 | Full grammar/style load (entities, includes) | [ ] SYSTEM `.ent`; phrases/phraseref/includephrases expand; style/variant/L2 |

### 3.C Rule families (per language, after stack)

| # | Check | Status |
|---|--------|--------|
| 3.C.1 | Core layout rules (whitespace, unpaired, …) | [ ] Tools.isParagraphEnd shared; ConvertToSentenceCaseFilter registered |
| 3.C.2 | Speller rules (Morfologik/Hunspell) | [ ] |
| 3.C.3 | Language-specific Java rules (EN AvsAn, DE, …) | [ ] by language |
| 3.C.4 | False friends | [ ] |
| 3.C.5 | Picky / default-off rules as Java (not soft picky XML) | [ ] |

### 3.D Languages

For **each** language claimed supported, check:

| # | Check | Status |
|---|--------|--------|
| 3.D.0 | Language module structure mirrors Java package | [ ] many packages exist |
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
