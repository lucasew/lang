# Soft port anti-cheat policy

Porting LanguageTool to Go must **follow Java LT logic**, not invent parallel behavior that only passes goldens.

## Hard rules

1. **No invention of goldens or rule lists**  
   Official LT data is the source of truth. Prefer `scripts/vendor-lt-testdata.py` extracts under `testdata/upstream/` and `testdata/grammar/*-upstream-soft.xml`. Do not hand-write soft golden sentences or fabricate upstream rule IDs.

2. **Every behavioral change needs an upstream reference**  
   Soft match / tagger / chunker / disambig / tokenizer changes must cite **where Java (or official data) does the same thing**:
   - Prefer: `inspiration/languagetool/...` path + type/method or rule `id`
   - Or: vendored extract path (`testdata/upstream/...`, `*-disambiguation-upstream-soft.xml`) + rule/id
   - Or: official tagset file shipped with LT (e.g. FreeLing `tagset_*.txt` under the language module)
   Put the reference in the code comment next to the change **and** in the commit body.

3. **Logic must be essentially the same**  
   Soft paths may omit unavailable assets (full dict, OpenNLP model) but must not invent a different algorithm. Acceptable: empty dict → soft surface probe **only** where Java would accept the same surface given the same pattern constraints. Unacceptable: new closed-class word lists with no Java/dict/tagset source; POS remaps that Java never applies; multiword tags invented for goldens.

4. **Prefer vendor over re-implement**  
   Missing disambiguation/grammar pieces: extract from upstream XML into soft packs (same rules/ids). Do not rewrite a “simpler” rule that only matches the golden.

5. **Soft approximations must be gated and labeled**  
   When soft behavior is a deliberate fall-back (no Morfologik dict), keep it behind the existing soft path (`StrictPOS == false`, soft helpers), comment the Java condition it approximates, and keep `StrictPOS`/real-tagger paths Java-faithful.

## Allowed soft techniques (with reference)

| Technique | Valid when referenced to |
|-----------|---------------------------|
| Soft closed-class surface match | Java closed POS family + dict/tagset closed-class membership |
| Soft irregular lemma map | Forms that Java tagger/dict would lemmatize to the pattern lemma |
| Soft OpenNLP-like chunk BIO | Java `EnglishChunker` / `EnglishChunkFilter` behavior |
| Soft FreeLing open/closed | Official FreeLing tagset + LT FreeLing postag patterns |
| Soft `_IS_URL` | Java `disambiguation-global.xml` URL rules / `WordTokenizer.isUrl` |

## Disallowed (cheat patterns)

- Inventing multiwords that change POS only to fire goldens (e.g. “for example” → `RB` without Java multiword entry)
- Broadening soft-accept so any letter word matches a closed POS
- Soft goldens not produced by the vendor script
- “Fix the golden” by weakening the rule instead of matching Java analysis

## Verification

- Main soft goldens: `LANG_{LANG}_MISS_SCAN=1 go test -run TestDebug{LANG}MissScan`
- Optional soft goldens: `LANG_{LANG}_OPT_MISS_SCAN=1 go test -run TestDebug{LANG}OptMissScan` (enables `SOFT_OPTIONAL`)
- Prefer diffs that only change code paths with an explicit upstream citation in the comment.

## Note on invent packs

Packs named `*-picky-soft.xml` / invent `SOFT_PICKY_*` / demo soft rules are **demo/UX**, not upstream fidelity. Do not use them as evidence that soft matching is correct. Upstream fidelity is measured only against vendored goldens and Java references.
