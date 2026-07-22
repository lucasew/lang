# Validation queue

**Source of truth** for unattended port loop status.  
**Only the parent** edits this file (see [`docs/loop-protocol.md`](../loop-protocol.md)).

Implementer and validator subagents **report** in their final message; they do **not** write this file.

---

## Constants

| Name | Value |
|------|--------|
| Reject CAP per attempt | 3 |
| Revisit after | 5 productive steps elsewhere, or next calendar day |
| Max parallel validators | 2 (disjoint `go_paths`) |
| Parent lock | `docs/validation/parent.lock` |
| Parent lock stale | 45 minutes |

---

## Status machine

```text
open → implementing → ready → validating → accepted
                               ↘ rejected → implementing → ready → …
blocked  (skip; revisit later)
```

---

## Schema (one row per checklist line in flight)

| Column | Description |
|--------|-------------|
| `checklist_id` | Stable id (e.g. `CHECKLIST.md#Languages` or path+heading) |
| `status` | `open` \| `implementing` \| `ready` \| `validating` \| `rejected` \| `accepted` \| `blocked` |
| `java_paths` | Paths under `inspiration/languagetool` |
| `go_paths` | Paths under `internal/languagetool` |
| `ready_commit` | SHA claimed ready / last audited |
| `round` | Reject count this attempt (0..CAP) |
| `attempt` | Revisit generation |
| `findings` | Validator REJECT notes (or empty) |
| `blocked_reason` | Required if `blocked` |
| `updated_at` | ISO-8601 UTC |

At most **one** `implementing` and at most **one** batch of `validating` owned by the current parent turn.

---

## Live queue

<!-- Parent maintains rows below. Empty table = nothing in flight. -->

| checklist_id | status | java_paths | go_paths | ready_commit | round | attempt | findings | blocked_reason | updated_at |
|--------------|--------|------------|----------|--------------|-------|---------|----------|----------------|------------|
| docs/faithful-port-checklist.md#3.A.5-en-hybrid-disambig-testChunker | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/tagging/en/EnglishHybridDisambiguator.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/tagging/disambiguation/rules/en/EnglishDisambiguationRuleTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/rules/XmlRuleDisambiguator.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/xx/DemoDisambiguator.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tagging/disambiguation/rules/en/; internal/languagetool/org/languagetool/tagging/en/; internal/languagetool/org/languagetool/tagging/disambiguation/ | 251a9820c304c212b19957f4b77b7abc8968682a | 0 | 0 | | | 2026-07-22T16:44:55Z |
| docs/faithful-port-checklist.md#3.A.5-fr-hybrid-disambig-testChunker | blocked | inspiration/languagetool/languagetool-language-modules/fr/src/test/java/org/languagetool/tagging/disambiguation/rules/fr/FrenchRuleDisambiguatorTest.java; inspiration/languagetool/languagetool-language-modules/fr/src/main/java/org/languagetool/tagging/disambiguation/fr/FrenchHybridDisambiguator.java; inspiration/languagetool/languagetool-language-modules/fr/src/main/java/org/languagetool/tagging/fr/FrenchTagger.java; inspiration/languagetool/languagetool-language-modules/fr/src/main/java/org/languagetool/tokenizers/fr/FrenchWordTokenizer.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/xx/DemoDisambiguator.java | internal/languagetool/org/languagetool/tagging/disambiguation/rules/fr/; internal/languagetool/org/languagetool/tagging/disambiguation/fr/; internal/languagetool/org/languagetool/tagging/fr/; internal/languagetool/org/languagetool/tokenizers/fr/ | | 0 | 0 | | missing official french.dict (not in inspiration resources or third_party; required for FrenchTagger / testChunker) | 2026-07-22T16:50:00Z |
| docs/faithful-port-checklist.md#3.A.3-MultiWordChunker-core-test | accepted | inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/MultiWordChunker.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/MultiWordChunker2.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/tagging/disambiguation/MultiWordChunkerTest.java; inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/resource/yy/multiwords.txt | internal/languagetool/org/languagetool/tagging/disambiguation/ | 3b4841a422994d3fbf96a5911960cdcd15c6a4e3 | 0 | 0 | | | 2026-07-22T17:01:20Z |
| docs/faithful-port-checklist.md#3.A.6-EnglishChunkFilter | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/chunking/EnglishChunkFilter.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkFilterTest.java | internal/languagetool/org/languagetool/chunking/ | 859b2393d3036268a114d213331661ced356f226 | 0 | 0 | | | 2026-07-22T17:10:02Z |
| docs/faithful-port-checklist.md#3.A.2-EnglishTagger-testTagger | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/tagging/en/EnglishTagger.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/tagging/en/EnglishTaggerTest.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tagging/en/; internal/languagetool/org/languagetool/tokenizers/ | aa8917ed451aad7fc3b4f5ce6458d32ca8988a43 | 0 | 0 | | | 2026-07-22T17:20:21Z |
| docs/faithful-port-checklist.md#3.A.1-EnglishWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/tokenizers/en/EnglishWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/tokenizers/en/EnglishWordTokenizerTest.java; inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/tagging/en/EnglishTagger.java | internal/languagetool/org/languagetool/tokenizers/en/; internal/languagetool/org/languagetool/tagging/en/ | f7c09ec6166551d84be43df852eeb8e22e70f8d5 | 0 | 0 | | | 2026-07-22T17:37:42Z |
| docs/faithful-port-checklist.md#3.A.1-EnglishSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/tokenizers/EnglishSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/; internal/attic/srx/ | e6904f79ce1631b8bbbc064d3f6ab835cc884244 | 0 | 0 | | | 2026-07-22T18:00:00Z |
| docs/faithful-port-checklist.md#3.A.1-SpanishSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/es/src/test/java/org/languagetool/tokenizers/es/SpanishSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/es/; internal/languagetool/org/languagetool/tokenizers/; internal/attic/srx/ | de99e3c8cbf3a5dcaff7682e0a8fa20d920f51fd | 0 | 0 | | | 2026-07-22T18:16:00Z |
| docs/faithful-port-checklist.md#3.A.1-PortugueseSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/pt/src/test/java/org/languagetool/tokenizers/pt/PortugueseSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/pt/; internal/languagetool/org/languagetool/tokenizers/ | 54736eb635a85b00bf4f87273ab56907ae21b8ae | 0 | 0 | | | 2026-07-22T18:25:30Z |
| docs/faithful-port-checklist.md#3.A.6-EnglishChunker | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/chunking/EnglishChunker.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/chunking/EnglishChunkerTest.java | internal/languagetool/org/languagetool/chunking/ | a59fd76b058ceb0a080c2dd8cf0ec8a584b69e6e | 0 | 0 | | | 2026-07-22T18:46:40Z |
| docs/faithful-port-checklist.md#3.A.1-DutchSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/nl/src/test/java/org/languagetool/tokenizers/nl/DutchSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/nl/; internal/languagetool/org/languagetool/tokenizers/ | 88163de559c555b3d1df9c02359c8333b91902ff | 0 | 0 | | | 2026-07-22T18:56:30Z |
| docs/faithful-port-checklist.md#3.A.1-ItalianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/it/src/test/java/org/languagetool/tokenizers/it/ItalianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/it/; internal/languagetool/org/languagetool/tokenizers/ | 03e03ed41bb68c1be70c59a5f299dde1d751ec3a | 0 | 0 | | | 2026-07-22T19:05:00Z |
| docs/faithful-port-checklist.md#3.A.1-GermanSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/de/src/test/java/org/languagetool/tokenizers/de/GermanSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/de/; internal/languagetool/org/languagetool/tokenizers/ | d1ca77aee5179710a64e94a8b7f4506ef0e7f974 | 0 | 0 | | | 2026-07-22T19:16:00Z |
| docs/faithful-port-checklist.md#3.A.1-PolishSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/pl/src/test/java/org/languagetool/tokenizers/pl/PolishSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/pl/; internal/languagetool/org/languagetool/tokenizers/ | 7c671115979b1fecfaabc31ff3903464933b5c8a | 0 | 0 | | | 2026-07-22T19:25:20Z |
| docs/faithful-port-checklist.md#3.A.1-UkrainianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/uk/src/test/java/org/languagetool/tokenizers/uk/UkrainianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/uk/; internal/languagetool/org/languagetool/tokenizers/; internal/attic/srx/ | b34d8d2e3ad1360a610dd030add12fbfb9be4289 | 0 | 0 | | | 2026-07-22T19:47:50Z |
| docs/faithful-port-checklist.md#3.A.1-RomanianSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ro/src/test/java/org/languagetool/tokenizers/ro/RomanianSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ro/; internal/languagetool/org/languagetool/tokenizers/ | 47db63ca9bb73db49be1e38fae5eaf701929b640 | 0 | 0 | | | 2026-07-22T19:55:35Z |
| docs/faithful-port-checklist.md#3.A.1-FrenchSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/fr/src/test/java/org/languagetool/tokenizers/fr/FrenchSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/fr/; internal/languagetool/org/languagetool/tokenizers/ | fbac8d4b9539bbe88179259887cfc2de905be66e | 0 | 0 | | | 2026-07-22T20:06:35Z |
| docs/faithful-port-checklist.md#3.A.1-CatalanSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ca/src/test/java/org/languagetool/tokenizers/ca/CatalanSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ca/; internal/languagetool/org/languagetool/tokenizers/; internal/attic/srx/ | a7d83527d53e219b4d63927eb70bc2d4cfe897d7 | 0 | 0 | | | 2026-07-22T20:26:50Z |
| docs/faithful-port-checklist.md#3.A.1-SlovakSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/sk/src/test/java/org/languagetool/tokenizers/sk/SlovakSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/sk/; internal/languagetool/org/languagetool/tokenizers/ | 106b04f0fd8a1c076a95737bb627d734ce061248 | 0 | 0 | | | 2026-07-22T20:35:20Z |
| docs/faithful-port-checklist.md#3.A.1-DanishSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/da/src/test/java/org/languagetool/tokenizers/da/DanishSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/da/; internal/languagetool/org/languagetool/tokenizers/ | 9181de3d6b9d6f046cfc26e1bc30bf2833c18cb4 | 0 | 0 | | | 2026-07-22T20:45:10Z |
| docs/faithful-port-checklist.md#3.A.1-RussianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ru/src/test/java/org/languagetool/tokenizers/ru/RussianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ru/; internal/languagetool/org/languagetool/tokenizers/ | e30f75f73201560ab1c7048caebb76642c99c262 | 0 | 0 | | | 2026-07-22T20:54:50Z |
| docs/faithful-port-checklist.md#3.A.1-SwedishSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/sv/src/test/java/org/languagetool/tokenizers/sv/SwedishSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/sv/; internal/languagetool/org/languagetool/tokenizers/ | fb900fab0841e79fc968d7f6f5853e8b975419e5 | 0 | 0 | | | 2026-07-22T21:06:30Z |
| docs/faithful-port-checklist.md#3.A.1-SerbianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/sr/src/test/java/org/languagetool/tokenizers/sr/SerbianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/sr/; internal/languagetool/org/languagetool/tokenizers/ | c30418613a01f3441088ec1cf6307774dbde1ec5 | 0 | 0 | | | 2026-07-22T21:15:40Z |
| docs/faithful-port-checklist.md#3.A.1-JapaneseSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ja/src/test/java/org/languagetool/tokenizers/ja/JapaneseSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ja/; internal/languagetool/org/languagetool/tokenizers/ | 2a230eb6c04ded7f8c8a3b9441d76ba172ed6eeb | 0 | 0 | | | 2026-07-22T21:25:00Z |
| docs/faithful-port-checklist.md#3.A.1-PersianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/fa/src/test/java/org/languagetool/tokenizers/PersianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/fa/; internal/languagetool/org/languagetool/tokenizers/ | e08f2318013a9be089bd0ee802369e5db7207400 | 0 | 0 | | | 2026-07-22T21:35:40Z |
| docs/faithful-port-checklist.md#3.A.1-AsturianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ast/src/test/java/org/languagetool/tokenizers/ast/AsturianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ast/; internal/languagetool/org/languagetool/tokenizers/ | 2e475bda4b67de2445491a648ad052a600e7f799 | 0 | 0 | | | 2026-07-22T21:45:10Z |
| docs/faithful-port-checklist.md#3.A.1-ArabicSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/rules/ar/; internal/languagetool/org/languagetool/tokenizers/ | f77e3aa1a5c70649f1ffb2b234fdcef4af61f8f2 | 0 | 0 | | | 2026-07-22T21:54:41Z |
| docs/faithful-port-checklist.md#3.A.1-CrimeanTatarSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/crh/src/test/java/org/languagetool/tokenizers/crh/CrimeanTatarSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/crh/; internal/languagetool/org/languagetool/tokenizers/ | 2f01e490881ace7fb1670c91fa5adc665fa9d92d | 0 | 0 | | | 2026-07-22T22:04:43Z |

---

## Human inbox

**You only need to look when the loop is idle** — no `ready`, no `rejected` under CAP, no implementable leaf, only `blocked` and/or `accepted`.

| When idle | Summary |
|-----------|---------|
| Last idle at | _(never)_ |
| Blocked lines | `3.A.5-fr-hybrid-disambig-testChunker` (missing french.dict) |
| Notes | |

---

## Changelog (parent appends short lines)

- Protocol bootstrapped; queue empty; no loop scheduled.
- 2026-07-22T16:32:35Z implement: 3.A.5-en-hybrid-disambig-testChunker (EnglishDisambiguationRuleTest.testChunker full twin)
- 2026-07-22T16:37:49Z ready: 3.A.5-en-hybrid-disambig-testChunker @ 251a9820c304 (implementer report)
- 2026-07-22T16:38:40Z validating: 3.A.5-en-hybrid-disambig-testChunker @ 251a9820
- 2026-07-22T16:44:55Z ACCEPT: 3.A.5-en-hybrid-disambig-testChunker @ 251a9820 (validator + parent green reconfirm)
- 2026-07-22T16:49:20Z implement: 3.A.5-fr-hybrid-disambig-testChunker (FrenchRuleDisambiguatorTest.testChunker full twin)
- 2026-07-22T16:50:00Z blocked: 3.A.5-fr-hybrid-disambig-testChunker — missing french.dict official asset
- 2026-07-22T16:50:00Z implement: 3.A.3-MultiWordChunker-core-test (Java MultiWordChunkerTest outcome twin)
- 2026-07-22T16:53:53Z ready: 3.A.3-MultiWordChunker-core-test @ 3b4841a42299 (implementer report)
- 2026-07-22T16:58:45Z validating: 3.A.3-MultiWordChunker-core-test @ 3b4841a4
- 2026-07-22T17:01:20Z ACCEPT: 3.A.3-MultiWordChunker-core-test @ 3b4841a4 (validator green)
- 2026-07-22T17:04:41Z implement: 3.A.6-EnglishChunkFilter (full EnglishChunkFilterTest outcome twin)
- 2026-07-22T17:07:10Z ready: 3.A.6-EnglishChunkFilter @ 859b2393d303 (implementer report)
- 2026-07-22T17:08:45Z validating: 3.A.6-EnglishChunkFilter @ 859b2393
- 2026-07-22T17:10:02Z ACCEPT: 3.A.6-EnglishChunkFilter @ 859b2393 (validator green)
- 2026-07-22T17:14:10Z implement: 3.A.2-EnglishTagger-testTagger (EnglishTaggerTest myAssert + real dict)
- 2026-07-22T17:17:51Z ready: 3.A.2-EnglishTagger-testTagger @ aa8917ed451a (implementer report)
- 2026-07-22T17:18:39Z validating: 3.A.2-EnglishTagger-testTagger @ aa8917ed
- 2026-07-22T17:20:21Z ACCEPT: 3.A.2-EnglishTagger-testTagger @ aa8917ed (validator green)
- 2026-07-22T17:24:06Z implement: 3.A.1-EnglishWordTokenizer (real EnglishTagger isTagged, not invent IsTaggedEN)
- 2026-07-22T17:31:57Z ready: 3.A.1-EnglishWordTokenizer @ f7c09ec61665 (implementer report)
- 2026-07-22T17:33:51Z validating: 3.A.1-EnglishWordTokenizer @ f7c09ec6
- 2026-07-22T17:37:42Z ACCEPT: 3.A.1-EnglishWordTokenizer @ f7c09ec6 (validator green)
- 2026-07-22T17:40:00Z implement: 3.A.1-EnglishSRXSentenceTokenizer (full EnglishSRXSentenceTokenizerTest testSplit twin; replace smoke)
- 2026-07-22T17:50:00Z ready: 3.A.1-EnglishSRXSentenceTokenizer @ e6904f79ce16 (implementer report; twin + attic/srx RE2/trailing-space fixes)
- 2026-07-22T17:53:30Z validating: 3.A.1-EnglishSRXSentenceTokenizer @ e6904f79
- 2026-07-22T18:00:00Z ACCEPT: 3.A.1-EnglishSRXSentenceTokenizer @ e6904f79 (validator + parent green reconfirm; attic/srx RE2 as official segment.srx runtime)
- 2026-07-22T18:05:00Z implement: 3.A.1-SpanishSentenceTokenizer (full SpanishSentenceTokenizerTest testSplit twin; replace smoke)
- 2026-07-22T18:11:30Z ready: 3.A.1-SpanishSentenceTokenizer @ de99e3c8cbf3 (implementer report; twin + attic/srx Java \b word-boundary fix)
- 2026-07-22T18:13:40Z validating: 3.A.1-SpanishSentenceTokenizer @ de99e3c8
- 2026-07-22T18:16:00Z ACCEPT: 3.A.1-SpanishSentenceTokenizer @ de99e3c8 (validator + parent green reconfirm; attic/srx Java \b as official segment.srx runtime)
- 2026-07-22T18:19:00Z implement: 3.A.1-PortugueseSRXSentenceTokenizer (full PortugueseSRXSentenceTokenizerTest testSplit twin)
- 2026-07-22T18:20:15Z ready: 3.A.1-PortugueseSRXSentenceTokenizer @ 54736eb635a8 (implementer report; full exact testSplit twin, no attic change)
- 2026-07-22T18:23:40Z validating: 3.A.1-PortugueseSRXSentenceTokenizer @ 54736eb6
- 2026-07-22T18:25:30Z ACCEPT: 3.A.1-PortugueseSRXSentenceTokenizer @ 54736eb6 (validator + parent green reconfirm)
- 2026-07-22T18:29:00Z implement: 3.A.6-EnglishChunker (full EnglishChunkerTest outcome twin; OpenNLP path; exact tags not Contains/NotEmpty)
- 2026-07-22T18:40:30Z ready: 3.A.6-EnglishChunker @ a59fd76b058c (implementer report; OpenNLP DefaultChunkerContext p_2 fix + exact twins)
- 2026-07-22T18:43:45Z validating: 3.A.6-EnglishChunker @ a59fd76b
- 2026-07-22T18:46:40Z ACCEPT: 3.A.6-EnglishChunker @ a59fd76b (validator + parent green reconfirm; OpenNLP chunker context parity)
- 2026-07-22T18:49:00Z implement: 3.A.1-DutchSRXSentenceTokenizer (full DutchSRXSentenceTokenizerTest testSplit twin; replace smoke)
- 2026-07-22T18:50:50Z ready: 3.A.1-DutchSRXSentenceTokenizer @ 88163de559c5 (implementer report; full exact testSplit twin, no attic change)
- 2026-07-22T18:53:45Z validating: 3.A.1-DutchSRXSentenceTokenizer @ 88163de5
- 2026-07-22T18:56:30Z ACCEPT: 3.A.1-DutchSRXSentenceTokenizer @ 88163de5 (validator + parent green reconfirm)
- 2026-07-22T19:00:05Z implement: 3.A.1-ItalianSRXSentenceTokenizer (full ItalianSRXSentenceTokenizerTest testSplit twin; replace smoke)
- 2026-07-22T19:01:30Z ready: 3.A.1-ItalianSRXSentenceTokenizer @ 03e03ed41bb6 (implementer report; full exact testSplit twin, no attic change)
- 2026-07-22T19:03:30Z validating: 3.A.1-ItalianSRXSentenceTokenizer @ 03e03ed4
- 2026-07-22T19:05:00Z ACCEPT: 3.A.1-ItalianSRXSentenceTokenizer @ 03e03ed4 (validator + parent green reconfirm)
- 2026-07-22T19:09:00Z implement: 3.A.1-GermanSRXSentenceTokenizer (full GermanSRXSentenceTokenizerTest testSplit twin; formalize/verify vs Java)
- 2026-07-22T19:10:30Z ready: 3.A.1-GermanSRXSentenceTokenizer @ d1ca77aee517 (implementer report; 65/65 testSplit + NBSP size asserts; no attic change)
- 2026-07-22T19:14:00Z validating: 3.A.1-GermanSRXSentenceTokenizer @ d1ca77ae
- 2026-07-22T19:16:00Z ACCEPT: 3.A.1-GermanSRXSentenceTokenizer @ d1ca77ae (validator + parent green reconfirm; 65/65 testSplit + 4 NBSP size)
- 2026-07-22T19:18:40Z implement: 3.A.1-PolishSentenceTokenizer (full PolishSentenceTokenizerTest testSplit twin; replace smoke)
- 2026-07-22T19:20:15Z ready: 3.A.1-PolishSentenceTokenizer @ 7c671115979b (implementer report; 24 exact testSplit; smoke deleted; no attic change)
- 2026-07-22T19:23:35Z validating: 3.A.1-PolishSentenceTokenizer @ 7c671115
- 2026-07-22T19:25:20Z ACCEPT: 3.A.1-PolishSentenceTokenizer @ 7c671115 (validator + parent green reconfirm; 24 exact testSplit)
- 2026-07-22T19:29:00Z implement: 3.A.1-UkrainianSRXSentenceTokenizer (full UkrainianSRXSentenceTokenizerTest twins; replace smoke)
- 2026-07-22T19:41:15Z ready: 3.A.1-UkrainianSRXSentenceTokenizer @ b34d8d2e3ad1 (implementer report; 124 exact testSplit; attic/srx \h\v lookaround empty-beforebreak; parent green + DE/PL/IT/NL/PT/ES regression ok)
- 2026-07-22T19:43:40Z validating: 3.A.1-UkrainianSRXSentenceTokenizer @ b34d8d2e
- 2026-07-22T19:47:50Z ACCEPT: 3.A.1-UkrainianSRXSentenceTokenizer @ b34d8d2e (validator + parent green reconfirm; 124 testSplit; attic/srx general ICU→RE2)
- 2026-07-22T19:49:00Z implement: 3.A.1-RomanianSentenceTokenizer (full RomanianSentenceTokenizerTest twin; stokenizer vs stokenizer2 paragraph modes)
- 2026-07-22T19:50:35Z ready: 3.A.1-RomanianSentenceTokenizer @ 47db63ca9bb7 (implementer report; 81 exact cases + dual paragraph modes; no attic change)
- 2026-07-22T19:53:40Z validating: 3.A.1-RomanianSentenceTokenizer @ 47db63ca
- 2026-07-22T19:55:35Z ACCEPT: 3.A.1-RomanianSentenceTokenizer @ 47db63ca (validator + parent green reconfirm; 81 cases dual paragraph modes)
- 2026-07-22T19:59:04Z implement: 3.A.1-FrenchSentenceTokenizer (full FrenchSentenceTokenizerTest twin; replace false "no @Test" smoke)
- 2026-07-22T20:00:50Z ready: 3.A.1-FrenchSentenceTokenizer @ fbac8d4b9539 (implementer report; full exact twin; parent green reconfirm; no SRX change)
- 2026-07-22T20:03:30Z validating: 3.A.1-FrenchSentenceTokenizer @ fbac8d4b
- 2026-07-22T20:06:35Z ACCEPT: 3.A.1-FrenchSentenceTokenizer @ fbac8d4b (validator + parent green reconfirm; 51 testSplit + 9 size asserts)

- 2026-07-22T20:08:50Z implement: 3.A.1-CatalanSentenceTokenizer (full CatalanSentenceTokenizerTest twin; replace false "no @Test" smoke)
- 2026-07-22T20:22:20Z ready: 3.A.1-CatalanSentenceTokenizer @ a7d83527d53e (implementer report; full exact twin; attic/srx loomchild exception lookbehind; parent green + FR/EN/ES/PT/NL/IT/DE/PL/UK/RO regression ok)
- 2026-07-22T20:23:40Z validating: 3.A.1-CatalanSentenceTokenizer @ a7d83527
- 2026-07-22T20:26:50Z ACCEPT: 3.A.1-CatalanSentenceTokenizer @ a7d83527 (validator + parent green reconfirm; 99 testSplit; attic/srx loomchild exception lookbehind)
- 2026-07-22T20:28:50Z implement: 3.A.1-SlovakSentenceTokenizer (full SlovakSentenceTokenizerTest twin; dual paragraph modes; replace false "no @Test" smoke)
- 2026-07-22T20:30:10Z ready: 3.A.1-SlovakSentenceTokenizer @ 106b04f0fd8a (implementer report; dual paragraph modes; parent green reconfirm; no attic change)
- 2026-07-22T20:33:40Z validating: 3.A.1-SlovakSentenceTokenizer @ 106b04f0
- 2026-07-22T20:35:20Z ACCEPT: 3.A.1-SlovakSentenceTokenizer @ 106b04f0 (validator + parent green reconfirm; 58 active cases dual paragraph modes)
- 2026-07-22T20:38:50Z implement: 3.A.1-DanishSRXSentenceTokenizer (full DanishSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T20:39:55Z ready: 3.A.1-DanishSRXSentenceTokenizer @ 9181de3d6b9d (implementer report; 31 exact testSplit; parent green reconfirm; no attic change)
- 2026-07-22T20:43:40Z validating: 3.A.1-DanishSRXSentenceTokenizer @ 9181de3d
- 2026-07-22T20:45:10Z ACCEPT: 3.A.1-DanishSRXSentenceTokenizer @ 9181de3d (validator + parent green reconfirm; 30 exact testSplit)
- 2026-07-22T20:48:50Z implement: 3.A.1-RussianSRXSentenceTokenizer (full RussianSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T20:50:00Z ready: 3.A.1-RussianSRXSentenceTokenizer @ e30f75f73201 (implementer report; 9 exact testSplit; parent green reconfirm; no attic change)
- 2026-07-22T20:53:35Z validating: 3.A.1-RussianSRXSentenceTokenizer @ e30f75f7
- 2026-07-22T20:54:50Z ACCEPT: 3.A.1-RussianSRXSentenceTokenizer @ e30f75f7 (validator + parent green reconfirm; 9 exact abbrev testSplit)
- 2026-07-22T20:58:50Z implement: 3.A.1-SwedishSRXSentenceTokenizer (full SwedishSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T21:01:00Z ready: 3.A.1-SwedishSRXSentenceTokenizer @ fb900fab0841 (implementer report; 2 testSplit / 26 parts exact twin; parent green reconfirm; no attic change)
- 2026-07-22T21:03:40Z validating: 3.A.1-SwedishSRXSentenceTokenizer @ fb900fab
- 2026-07-22T21:06:30Z ACCEPT: 3.A.1-SwedishSRXSentenceTokenizer @ fb900fab (validator + parent green reconfirm; 26 exact testSplit parts / 2 calls)
- 2026-07-22T21:09:00Z implement: 3.A.1-SerbianSRXSentenceTokenizer (full SerbianSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T21:10:40Z ready: 3.A.1-SerbianSRXSentenceTokenizer @ c30418613a01 (implementer report; 25 active exact testSplit; parent green reconfirm; no attic change)
- 2026-07-22T21:13:40Z validating: 3.A.1-SerbianSRXSentenceTokenizer @ c3041861
- 2026-07-22T21:15:40Z ACCEPT: 3.A.1-SerbianSRXSentenceTokenizer @ c3041861 (validator + parent green reconfirm; 25 active exact testSplit)
- 2026-07-22T21:18:50Z implement: 3.A.1-JapaneseSRXSentenceTokenizer (full JapaneseSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T21:19:50Z ready: 3.A.1-JapaneseSRXSentenceTokenizer @ 2a230eb6c04d (implementer report; 10 exact testSplit; parent green reconfirm; no attic change)
- 2026-07-22T21:23:40Z validating: 3.A.1-JapaneseSRXSentenceTokenizer @ 2a230eb6
- 2026-07-22T21:25:00Z ACCEPT: 3.A.1-JapaneseSRXSentenceTokenizer @ 2a230eb6 (validator + parent green reconfirm; 10 exact testSplit)
- 2026-07-22T21:29:00Z implement: 3.A.1-PersianSRXSentenceTokenizer (full PersianSRXSentenceTokenizerTest twin; replace incomplete smoke)
- 2026-07-22T21:30:00Z ready: 3.A.1-PersianSRXSentenceTokenizer @ e08f2318013a (implementer report; 6 exact testSplit; deleted invent smoke; parent green reconfirm; no attic change)
- 2026-07-22T21:33:40Z validating: 3.A.1-PersianSRXSentenceTokenizer @ e08f2318
- 2026-07-22T21:35:40Z ACCEPT: 3.A.1-PersianSRXSentenceTokenizer @ e08f2318 (validator + parent green reconfirm; 6 exact testSplit; invent smoke deleted)
- 2026-07-22T21:38:50Z implement: 3.A.1-AsturianSRXSentenceTokenizer (full AsturianSRXSentenceTokenizerTest twin; dual paragraph modes; replace smoke)
- 2026-07-22T21:40:00Z ready: 3.A.1-AsturianSRXSentenceTokenizer @ 2e475bda4b67 (implementer report; 5 exact testSplit dual paragraph modes; parent green reconfirm; no attic change)
- 2026-07-22T21:43:40Z validating: 3.A.1-AsturianSRXSentenceTokenizer @ 2e475bda
- 2026-07-22T21:45:10Z ACCEPT: 3.A.1-AsturianSRXSentenceTokenizer @ 2e475bda (validator + parent green reconfirm; 5 exact testSplit dual paragraph modes)
- 2026-07-22T21:49:14Z implement: 3.A.1-ArabicSRXSentenceTokenizer (full ArabicSRXSentenceTokenizerTest twin; replace smoke)
- 2026-07-22T21:50:28Z ready: 3.A.1-ArabicSRXSentenceTokenizer @ f77e3aa1a5c7 (implementer report; 4 exact testSplit; parent green reconfirm; no attic change)
- 2026-07-22T21:53:44Z validating: 3.A.1-ArabicSRXSentenceTokenizer @ f77e3aa1
- 2026-07-22T21:54:41Z ACCEPT: 3.A.1-ArabicSRXSentenceTokenizer @ f77e3aa1 (validator + parent green reconfirm; 4 exact testSplit)
- 2026-07-22T21:58:56Z implement: 3.A.1-CrimeanTatarSRXSentenceTokenizer (full CrimeanTatarSRXSentenceTokenizerTest twin; replace smoke)
- 2026-07-22T21:59:44Z ready: 3.A.1-CrimeanTatarSRXSentenceTokenizer @ 2f01e490881a (implementer report; 2 exact testSplit; singleLineBreaks=true; parent green reconfirm; no attic change)
- 2026-07-22T22:03:43Z validating: 3.A.1-CrimeanTatarSRXSentenceTokenizer @ 2f01e490
- 2026-07-22T22:04:43Z ACCEPT: 3.A.1-CrimeanTatarSRXSentenceTokenizer @ 2f01e490 (validator + parent green reconfirm; 2 exact testSplit; singleLineBreaks=true)
