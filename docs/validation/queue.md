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
| docs/faithful-port-checklist.md#3.A.1-LithuanianSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/lt/src/test/java/org/languagetool/tokenizers/lt/LithuanianSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/lt/; internal/languagetool/org/languagetool/tokenizers/ | c1e8a1d3f80a7a69ec7ff31aba6cbf4979e08acb | 0 | 0 | | | 2026-07-22T22:14:59Z |
| docs/faithful-port-checklist.md#3.A.1-MalayalamSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ml/src/test/java/org/languagetool/tokenizers/ml/MalayalamSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/ml/; internal/languagetool/org/languagetool/tokenizers/ | 45dddca11347d82499b9e6ea397afe579cc1657f | 0 | 0 | | | 2026-07-22T22:24:48Z |
| docs/faithful-port-checklist.md#3.A.1-TagalogSRXSentenceTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/tl/src/test/java/org/languagetool/tokenizers/tl/TagalogSRXSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/tl/; internal/languagetool/org/languagetool/tokenizers/ | 614f57df50dadbf0822eaf508d69fc53a90807da | 0 | 0 | | | 2026-07-22T22:35:41Z |
| docs/faithful-port-checklist.md#3.A.1-UkrainianWordTokenizer | blocked | inspiration/languagetool/languagetool-language-modules/uk/src/main/java/org/languagetool/tokenizers/uk/UkrainianWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/uk/src/test/java/org/languagetool/tokenizers/uk/UkrainianWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/uk/ | a12f5e61c2096d58cbb631cbd1e4175c634d22b1 | 3 | 1 | r1 FIXED NUMBER/WEB/meter/\h; r1 ABBR_DOT_2 BOS fixed@33 then re-broken; r2 FIXED apostrophe case + soft-hyphen BOS @a12f5e61; r3 REJECT: ABBR_DOT_2_SMALL BOS ^ invent reintroduced (е.е. glues); twins dropped. | reject CAP=3 hit @ a12f5e61 attempt=1; ABBR_DOT_2 BOS invent; revisit after K=5 productive steps | 2026-07-23T02:49:25Z |
| docs/faithful-port-checklist.md#3.A.1-DutchWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/nl/src/main/java/org/languagetool/tokenizers/nl/DutchWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/nl/src/test/java/org/languagetool/tokenizers/nl/DutchWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/nl/ | 67a384077c37b485a78b87ccb06239ac5a241a0a | 0 | 0 | | | 2026-07-22T23:31:00Z |
| docs/faithful-port-checklist.md#3.A.1-SpanishWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/es/src/main/java/org/languagetool/tokenizers/es/SpanishWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/es/src/test/java/org/languagetool/tokenizers/es/SpanishWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/es/ | 32995ed8e9e303ab3efcd406e53eb6abcc0634a1 | 0 | 1 | | | 2026-07-23T02:55:59Z |
| docs/faithful-port-checklist.md#3.A.1-GoogleStyleWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/en/src/main/java/org/languagetool/rules/en/GoogleStyleWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/GoogleStyleWordTokenizerTest.java | internal/languagetool/org/languagetool/rules/en/ | c37e38b0e35ac2b048bc49070049897ac41ec1b1 | 0 | 0 | | | 2026-07-23T03:10:38Z |
| docs/faithful-port-checklist.md#3.A.1-RussianWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ru/src/main/java/org/languagetool/tokenizers/ru/RussianWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/ru/ | ce01636889272ec0d17539e9877b28c7d72b07c3 | 0 | 0 | | | 2026-07-23T03:25:50Z |
| docs/faithful-port-checklist.md#3.A.1-GermanWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/de/src/main/java/org/languagetool/tokenizers/de/GermanWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/de/ | 8daf13f02d497784cb309ef23fc3b7a6aed4eb52 | 0 | 0 | | | 2026-07-23T03:34:59Z |
| docs/faithful-port-checklist.md#3.A.1-ArabicWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ar/src/main/java/org/languagetool/tokenizers/ArabicWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/ | 957a5824be3b22fe1d938f440bff5a9347d398dd | 0 | 0 | | | 2026-07-23T03:45:26Z |
| docs/faithful-port-checklist.md#3.A.1-PersianWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/fa/src/main/java/org/languagetool/tokenizers/PersianWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/ | 8a2714aac2d47383ebc3cd71308771f251c2874c | 0 | 0 | | | 2026-07-23T03:55:29Z |
| docs/faithful-port-checklist.md#3.A.1-SimpleSentenceTokenizer | accepted | inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SimpleSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/tokenizers/SimpleSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/tokenizers/segment-simple.srx | internal/languagetool/org/languagetool/tokenizers/ | 17f7a3f4757361bd7bfca3edf7243db497ee8341 | 0 | 0 | | | 2026-07-23T04:05:29Z |
| docs/faithful-port-checklist.md#3.A.1-TagalogWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/tl/src/main/java/org/languagetool/language/tokenizers/TagalogWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/tl/ | ba9d62fddb8d1c195e5e9bb0c6f259e31625eea4 | 0 | 0 | | | 2026-07-23T04:15:40Z |
| docs/faithful-port-checklist.md#3.A.1-KhmerWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/km/src/main/java/org/languagetool/tokenizers/km/KhmerWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/km/ | 8e70ba8abc9286bc72ead039433e49e09f48b3f6 | 0 | 0 | | | 2026-07-23T04:24:53Z |
| docs/faithful-port-checklist.md#3.A.1-MalayalamWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ml/src/main/java/org/languagetool/tokenizers/ml/MalayalamWordTokenizer.java | internal/languagetool/org/languagetool/tokenizers/ml/ | 2e1a98ff6db2582f49280bdd1ee5a6d1b1cc2ef3 | 0 | 0 | | | 2026-07-23T04:34:41Z |
| docs/faithful-port-checklist.md#3.A.1-FrenchWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/fr/src/main/java/org/languagetool/tokenizers/fr/FrenchWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/fr/src/test/java/org/languagetool/tokenizers/fr/FrenchWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/fr/ | 124558390772cdfab14cfc268861955f3191da1b | 0 | 0 | | | 2026-07-23T00:17:40Z |
| docs/faithful-port-checklist.md#3.A.1-PortugueseWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/pt/src/main/java/org/languagetool/tokenizers/pt/PortugueseWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/pt/src/test/java/org/languagetool/tokenizers/pt/PortugueseWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/pt/ | 63245687540188a0cfa7c02aefa4900d52eb1599 | 0 | 0 | | | 2026-07-23T00:33:00Z |
| docs/faithful-port-checklist.md#3.A.1-CatalanWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ca/src/main/java/org/languagetool/tokenizers/ca/CatalanWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/ca/src/test/java/org/languagetool/tokenizers/ca/CatalanWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/ca/ | d9a58178450237159168934c953b0d4816a874fd | 0 | 0 | | | 2026-07-23T00:46:50Z |
| docs/faithful-port-checklist.md#3.A.1-BelarusianWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/be/src/main/java/org/languagetool/tokenizers/be/BelarusianWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/be/src/test/java/org/languagetool/tokenizers/be/BelarusianWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/be/ | 67a384077c37b485a78b87ccb06239ac5a241a0a | 0 | 0 | | | 2026-07-23T00:56:20Z |
| docs/faithful-port-checklist.md#3.A.1-BretonWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/br/src/main/java/org/languagetool/tokenizers/br/BretonWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/br/src/test/java/org/languagetool/tokenizers/br/BretonWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/br/ | 820a52e0c257c2dbb2d59fe73300c23ce0ff3817 | 0 | 0 | | | 2026-07-23T01:05:50Z |
| docs/faithful-port-checklist.md#3.A.1-RomanianWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ro/src/main/java/org/languagetool/tokenizers/ro/RomanianWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/ro/src/test/java/org/languagetool/tokenizers/ro/RomanianWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/ro/ | 8811b9c8e5b886f9962c99095e46a9748c9a41fb | 0 | 0 | | | 2026-07-23T01:15:20Z |
| docs/faithful-port-checklist.md#3.A.1-EsperantoWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/eo/src/main/java/org/languagetool/tokenizers/eo/EsperantoWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/eo/src/test/java/org/languagetool/tokenizers/eo/EsperantoWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/eo/ | b5dda793ba3176efaa4b00afe1c13117746907c8 | 0 | 0 | | | 2026-07-23T01:30:40Z |
| docs/faithful-port-checklist.md#3.A.1-CrimeanTatarWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/crh/src/main/java/org/languagetool/tokenizers/crh/CrimeanTatarWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/crh/src/test/java/org/languagetool/tokenizers/crh/CrimeanTatarWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/crh/ | b6e81c88f78c3ee90ba28087a8fbba5864456044 | 0 | 0 | | | 2026-07-23T01:40:40Z |
| docs/faithful-port-checklist.md#3.A.1-PolishWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/pl/src/main/java/org/languagetool/tokenizers/pl/PolishWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/pl/src/test/java/org/languagetool/tokenizers/pl/PolishWordTokenizerTest.java; inspiration/languagetool/languagetool-language-modules/pl/src/main/java/org/languagetool/tagging/pl/PolishTagger.java; inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl/polish.dict | internal/languagetool/org/languagetool/tokenizers/pl/; internal/languagetool/org/languagetool/tagging/pl/ | c1772adc5670ff1a6cb9066944e5e7d473fda6c7 | 0 | 0 | | | 2026-07-23T01:58:40Z |
| docs/faithful-port-checklist.md#3.A.1-JapaneseWordTokenizer | accepted | inspiration/languagetool/languagetool-language-modules/ja/src/main/java/org/languagetool/tokenizers/ja/JapaneseWordTokenizer.java; inspiration/languagetool/languagetool-language-modules/ja/src/test/java/org/languagetool/tokenizers/ja/JapaneseWordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/ja/ | 528837ae27b4daee7e0dc41af6a255b613d01283 | 0 | 0 | | | 2026-07-23T02:03:44Z |
| docs/faithful-port-checklist.md#3.A.1-core-WordTokenizer | accepted | inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/WordTokenizer.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/tokenizers/WordTokenizerTest.java | internal/languagetool/org/languagetool/tokenizers/ | 786cbae497100e820762081b9e786763bf3c97af | 0 | 0 | | | 2026-07-23T02:12:25Z |

---

## Human inbox

**You only need to look when the loop is idle** — no `ready`, no `rejected` under CAP, no implementable leaf, only `blocked` and/or `accepted`.

| When idle | Summary |
|-----------|---------|
| Last idle at | _(never)_ |
| Blocked lines | `3.A.5-fr-hybrid-disambig-testChunker` (missing french.dict); `3.A.1-UkrainianWordTokenizer` (CAP-revisit attempt=1 CAP=3: ABBR_DOT_2 BOS invent reintroduced; apostrophe/soft-hyphen/NUMBER/WEB fixed); `3.A.1-SpanishWordTokenizer` (ACCEPTed @ 32995ed8 CAP-revisit attempt=1) |
| Notes | MalayalamWordTokenizer ACCEPT @ 2e1a98ff. UK CAP-revisit still eligible (K≥5). Remaining 3.A.1 leaves: Galician/Greek/Chinese WordTokenizer, ChineseSentence; FR hybrid still missing french.dict. |

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
- 2026-07-22T22:08:46Z implement: 3.A.1-LithuanianSRXSentenceTokenizer (full LithuanianSRXSentenceTokenizerTest twin; replace smoke)
- 2026-07-22T22:09:29Z ready: 3.A.1-LithuanianSRXSentenceTokenizer @ c1e8a1d3f80a (implementer report; 1 exact multi-part testSplit; parent green reconfirm; no attic change)
- 2026-07-22T22:13:40Z validating: 3.A.1-LithuanianSRXSentenceTokenizer @ c1e8a1d3
- 2026-07-22T22:14:59Z ACCEPT: 3.A.1-LithuanianSRXSentenceTokenizer @ c1e8a1d3 (validator + parent green reconfirm; 1 exact multi-part testSplit)
- 2026-07-22T22:18:47Z implement: 3.A.1-MalayalamSRXSentenceTokenizer (full MalayalamSRXSentenceTokenizerTest twin; replace smoke)
- 2026-07-22T22:19:40Z ready: 3.A.1-MalayalamSRXSentenceTokenizer @ 45dddca11347 (implementer report; 1 exact multi-part testSplit; parent green reconfirm; no attic change)
- 2026-07-22T22:23:41Z validating: 3.A.1-MalayalamSRXSentenceTokenizer @ 45dddca1
- 2026-07-22T22:24:48Z ACCEPT: 3.A.1-MalayalamSRXSentenceTokenizer @ 45dddca1 (validator + parent green reconfirm; 1 exact multi-part testSplit)
- 2026-07-22T22:28:50Z implement: 3.A.1-TagalogSRXSentenceTokenizer (full TagalogSRXSentenceTokenizerTest twin; replace smoke)
- 2026-07-22T22:29:36Z ready: 3.A.1-TagalogSRXSentenceTokenizer @ 614f57df50da (implementer report; 1 exact multi-part testSplit; parent green reconfirm; no attic change)
- 2026-07-22T22:33:42Z validating: 3.A.1-TagalogSRXSentenceTokenizer @ 614f57df
- 2026-07-22T22:35:41Z ACCEPT: 3.A.1-TagalogSRXSentenceTokenizer @ 614f57df (validator + parent green reconfirm; 1 exact multi-part testSplit)
- 2026-07-22T22:40:00Z implement: 3.A.1-UkrainianWordTokenizer (full UkrainianWordTokenizerTest outcome twin; fill markdown/special/web/dash gaps; N_DASH_SPACE exact)
- 2026-07-22T22:49:00Z ready: 3.A.1-UkrainianWordTokenizer @ c24a2fc2ed1b (implementer report; 15 @Test full exact twins + prod abbr/web fixes; parent go test uk green)
- 2026-07-22T22:53:40Z validating: 3.A.1-UkrainianWordTokenizer @ c24a2fc2
- 2026-07-22T23:00:00Z REJECT: 3.A.1-UkrainianWordTokenizer @ c24a2fc2 round=1 — LEADING_DASH first-byte bug; breakPlus Latin over-split
- 2026-07-22T23:03:40Z fix/implementing: 3.A.1-UkrainianWordTokenizer round=1 findings (LEADING_DASH rune gate; breakPlus no Latin)
- 2026-07-22T23:05:30Z ready: 3.A.1-UkrainianWordTokenizer @ 94c94c72 (fix round=1: LEADING_DASH first-rune + breakPlus no Latin; regression asserts; parent go test uk green)
- 2026-07-22T23:08:40Z validating: 3.A.1-UkrainianWordTokenizer @ 94c94c72
- 2026-07-22T23:12:30Z REJECT: 3.A.1-UkrainianWordTokenizer @ 94c94c72 round=2 — prior LEADING_DASH/breakPlus fixed; invent bare в. via replaceAbbrDotV
- 2026-07-22T23:13:40Z fix/implementing: 3.A.1-UkrainianWordTokenizer round=2 (remove replaceAbbrDotV invent; keep VO paths)
- 2026-07-22T23:15:40Z ready: 3.A.1-UkrainianWordTokenizer @ cafa2037 (fix round=2: bare в. invent removed; regression asserts; parent go test uk green)
- 2026-07-22T23:18:40Z validating: 3.A.1-UkrainianWordTokenizer @ cafa2037
- 2026-07-22T23:23:00Z REJECT→blocked CAP: 3.A.1-UkrainianWordTokenizer @ cafa2037 round=3 — invent (?i) abbr + mid-word emdash; prior r1/r2 fixed; skip thrash
- 2026-07-22T23:24:30Z implement: 3.A.1-DutchWordTokenizer (full DutchWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-22T23:26:30Z ready: 3.A.1-DutchWordTokenizer @ 67a38407 (implementer audit: 20/20 exact assertTokenize; prod 1:1; no further edit; parent go test nl green)
- 2026-07-22T23:28:40Z validating: 3.A.1-DutchWordTokenizer @ 67a38407
- 2026-07-22T23:31:00Z ACCEPT: 3.A.1-DutchWordTokenizer @ 67a38407 (validator + parent green reconfirm; 20 exact assertTokenize)
- 2026-07-22T23:34:00Z implement: 3.A.1-SpanishWordTokenizer (full SpanishWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-22T23:36:30Z ready: 3.A.1-SpanishWordTokenizer @ 4109e2ce (implementer audit: full exact twins; IsTaggedES hook; no further edit; parent go test es green)
- 2026-07-22T23:38:45Z validating: 3.A.1-SpanishWordTokenizer @ 4109e2ce
- 2026-07-22T23:42:40Z REJECT: 3.A.1-SpanishWordTokenizer @ 4109e2ce round=1 — ORDINAL_POINT missing trailing Unicode \b (1.apple over-merge)
- 2026-07-22T23:43:40Z fix/implementing: 3.A.1-SpanishWordTokenizer round=1 findings (ORDINAL_POINT trailing Unicode \b)
- 2026-07-22T23:46:00Z ready: 3.A.1-SpanishWordTokenizer @ e063713b (fix round=1: ORDINAL_POINT Unicode right-edge gate; regression 1.apple/1.asiento; parent go test es green)
- 2026-07-22T23:48:30Z validating: 3.A.1-SpanishWordTokenizer @ e063713b
- 2026-07-22T23:52:10Z REJECT: 3.A.1-SpanishWordTokenizer @ e063713b round=2 — LEADING ORDINAL_POINT Unicode \b missing (ñ1.o / á1.º over-merge); r1 trailing fixed
- 2026-07-22T23:53:40Z fix/implementing: 3.A.1-SpanishWordTokenizer round=2 findings (LEADING ORDINAL_POINT Unicode \b)
- 2026-07-22T23:55:30Z ready: 3.A.1-SpanishWordTokenizer @ e69a0d01 (fix round=2: ORDINAL_POINT both-edge Unicode gate; ñ1.o/á1.º; parent go test es green)
- 2026-07-22T23:58:35Z validating: 3.A.1-SpanishWordTokenizer @ e69a0d01
- 2026-07-23T00:04:10Z REJECT→blocked CAP: 3.A.1-SpanishWordTokenizer @ e69a0d01 round=3 — invent-incomplete UCC \w/\d ORDINAL_POINT; r1/r2 letter gates fixed; skip thrash
- 2026-07-23T00:09:00Z implement: 3.A.1-FrenchWordTokenizer (full FrenchWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T00:12:20Z ready: 3.A.1-FrenchWordTokenizer @ 12455839 (U+2010/U+2011 normalize + full testTokenize twin; parent go test fr green)
- 2026-07-23T00:13:40Z validating: 3.A.1-FrenchWordTokenizer @ 12455839
- 2026-07-23T00:17:40Z ACCEPT: 3.A.1-FrenchWordTokenizer @ 12455839 (validator + parent green reconfirm; full testTokenize twin; U+2010/U+2011 1:1)
- 2026-07-23T00:19:00Z implement: 3.A.1-PortugueseWordTokenizer (full PortugueseWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T00:24:00Z ready: 3.A.1-PortugueseWordTokenizer @ 63245687 (date/DECIMAL_SPACE/wordChars 1:1; 28/28 twins; parent go test pt green)
- 2026-07-23T00:28:40Z validating: 3.A.1-PortugueseWordTokenizer @ 63245687
- 2026-07-23T00:33:00Z ACCEPT: 3.A.1-PortugueseWordTokenizer @ 63245687 (validator + parent green reconfirm; 28/28 exact twins)
- 2026-07-23T00:33:50Z implement: 3.A.1-CatalanWordTokenizer (full CatalanWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T00:40:20Z ready: 3.A.1-CatalanWordTokenizer @ d9a58178 (ELA_GEMINADA no (?i); full testTokenize twin; parent go test ca green)
- 2026-07-23T00:43:40Z validating: 3.A.1-CatalanWordTokenizer @ d9a58178
- 2026-07-23T00:46:50Z ACCEPT: 3.A.1-CatalanWordTokenizer @ d9a58178 (validator + parent green reconfirm; full testTokenize twin; ELA_GEMINADA 1:1)
- 2026-07-23T00:49:00Z implement: 3.A.1-BelarusianWordTokenizer (full BelarusianWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T00:52:30Z ready: 3.A.1-BelarusianWordTokenizer @ 67a38407 (implementer audit: full testTokenize twin; UTF16Len length>1; no further edit; parent go test be green)
- 2026-07-23T00:53:40Z validating: 3.A.1-BelarusianWordTokenizer @ 67a38407
- 2026-07-23T00:56:20Z ACCEPT: 3.A.1-BelarusianWordTokenizer @ 67a38407 (validator + parent green reconfirm; full testTokenize twin; UTF16Len length>1 1:1)
- 2026-07-23T00:58:40Z implement: 3.A.1-BretonWordTokenizer (full BretonWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T01:01:10Z ready: 3.A.1-BretonWordTokenizer @ 820a52e0 (implementer audit: full testTokenize twin; c'h/n’eo 1:1; no further edit; parent go test br green)
- 2026-07-23T01:03:40Z validating: 3.A.1-BretonWordTokenizer @ 820a52e0
- 2026-07-23T01:05:50Z ACCEPT: 3.A.1-BretonWordTokenizer @ 820a52e0 (validator + parent green reconfirm; full testTokenize twin; c'h/n’eo 1:1)
- 2026-07-23T01:08:40Z implement: 3.A.1-RomanianWordTokenizer (full RomanianWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T01:11:00Z ready: 3.A.1-RomanianWordTokenizer @ 8811b9c8 (implementer audit: full testTokenize twin; delimiter set 1:1; no further edit; parent go test ro WordTokenizer green)
- 2026-07-23T01:13:40Z validating: 3.A.1-RomanianWordTokenizer @ 8811b9c8
- 2026-07-23T01:15:20Z ACCEPT: 3.A.1-RomanianWordTokenizer @ 8811b9c8 (validator + parent green reconfirm; full testTokenize twin; delimiter set 1:1)
- 2026-07-23T01:18:40Z implement: 3.A.1-EsperantoWordTokenizer (full EsperantoWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T01:24:30Z ready: 3.A.1-EsperantoWordTokenizer @ b5dda793 (ASCII \\b lookaround fidelity fix; full testTokenize twin; parent go test eo green)
- 2026-07-23T01:28:40Z validating: 3.A.1-EsperantoWordTokenizer @ b5dda793
- 2026-07-23T01:30:40Z ACCEPT: 3.A.1-EsperantoWordTokenizer @ b5dda793 (validator + parent green reconfirm; full testTokenize twin; ASCII \\b lookaround 1:1)
- 2026-07-23T01:33:40Z implement: 3.A.1-CrimeanTatarWordTokenizer (full CrimeanTatarWordTokenizerTest outcome twin; Java-king prod fidelity)
- 2026-07-23T01:35:30Z ready: 3.A.1-CrimeanTatarWordTokenizer @ b6e81c88 (UTF16Len trailing-dash gate; full testTokenize twin; parent go test crh WordTokenizer green)
- 2026-07-23T01:38:40Z validating: 3.A.1-CrimeanTatarWordTokenizer @ b6e81c88
- 2026-07-23T01:40:40Z ACCEPT: 3.A.1-CrimeanTatarWordTokenizer @ b6e81c88 (validator + parent green reconfirm; full testTokenize twin; UTF16Len trailing-dash 1:1)
- 2026-07-23T01:44:32Z implement: 3.A.1-PolishWordTokenizer (full PolishWordTokenizerTest twin; real PolishTagger/polish.dict not invent mocks)
- 2026-07-23T01:51:34Z ready: 3.A.1-PolishWordTokenizer @ c1772adc (implementer report; full testTokenize twin; real PolishTagger/polish.dict; mocks deleted; parent green reconfirm)
- 2026-07-23T01:52:34Z validating: 3.A.1-PolishWordTokenizer @ c1772adc
- 2026-07-23T01:58:40Z ACCEPT: 3.A.1-PolishWordTokenizer @ c1772adc (validator + parent green reconfirm; full testTokenize twin; real PolishTagger/polish.dict)
- 2026-07-23T01:59:54Z implement: 3.A.1-JapaneseWordTokenizer (full JapaneseWordTokenizerTest twin; Sen→kagome IPA)
- 2026-07-23T02:02:07Z ready: 3.A.1-JapaneseWordTokenizer @ 528837ae (implementer report; full testTokenize twin; kagome IPA; Segment invent removed; parent green reconfirm)
- 2026-07-23T02:02:07Z validating: 3.A.1-JapaneseWordTokenizer @ 528837ae
- 2026-07-23T02:03:44Z ACCEPT: 3.A.1-JapaneseWordTokenizer @ 528837ae (validator + parent green reconfirm; full testTokenize twin; kagome IPA ≡ Sen-visible)
- 2026-07-23T02:04:12Z implement: 3.A.1-core-WordTokenizer (full core WordTokenizerTest twin)
- 2026-07-23T02:09:28Z ready: 3.A.1-core-WordTokenizer @ 786cbae4 (implementer report; full WordTokenizerTest twins; email/emoji/currency; parent green reconfirm)
- 2026-07-23T02:09:28Z validating: 3.A.1-core-WordTokenizer @ 786cbae4
- 2026-07-23T02:12:25Z ACCEPT: 3.A.1-core-WordTokenizer @ 786cbae4 (validator + parent green reconfirm; 9/9 Java @Test twins; email/emoji/currency faithful)
- 2026-07-23T02:12:25Z CAP-revisit implement: 3.A.1-UkrainianWordTokenizer attempt=1 (K≥5 productive steps; fix case-sensitive abbr + mid-word emdash invent)
- 2026-07-23T02:16:48Z ready: 3.A.1-UkrainianWordTokenizer @ 78f80747 attempt=1 (case-sensitive abbr + emdash invent removed; parent green reconfirm)
- 2026-07-23T02:16:48Z validating: 3.A.1-UkrainianWordTokenizer @ 78f80747 attempt=1
- 2026-07-23T02:21:36Z REJECT: 3.A.1-UkrainianWordTokenizer @ 78f80747 attempt=1 round=1 — NUMBER_MISSING_SPACE/WEB_ENTITIES/ABBR_DOT_2_SMALL/\h invent
- 2026-07-23T02:21:36Z fix/implementing: 3.A.1-UkrainianWordTokenizer attempt=1 round=1 findings
- 2026-07-23T02:32:24Z ready: 3.A.1-UkrainianWordTokenizer @ 6e4179cb attempt=1 round=1 fix (parent green reconfirm)
- 2026-07-23T02:32:24Z validating: 3.A.1-UkrainianWordTokenizer @ 6e4179cb attempt=1 round=1
- 2026-07-23T02:37:36Z ready: 3.A.1-UkrainianWordTokenizer @ 33ddaa56 attempt=1 round=1 fix (NUMBER_MISSING_SPACE/WEB_ENTITIES/ABBR_DOT_2_SMALL/DECIMAL_SPACE; parent green reconfirm)
- 2026-07-23T02:38:50Z validating: 3.A.1-UkrainianWordTokenizer @ 33ddaa56 attempt=1 round=1
- 2026-07-23T02:45:36Z REJECT: 3.A.1-UkrainianWordTokenizer @ 33ddaa56 attempt=1 round=2 — APOSTROPHE_BEGIN ToLower + SOFT_HYPHEN BOS invent
- 2026-07-23T02:45:36Z fix/ready: 3.A.1-UkrainianWordTokenizer @ a12f5e61 attempt=1 round=2 (case-sensitive apostrophe + BOS soft-hyphen; parent green reconfirm)
- 2026-07-23T02:45:36Z validating: 3.A.1-UkrainianWordTokenizer @ a12f5e61 attempt=1 round=2
- 2026-07-23T02:48:59Z REJECT→blocked CAP: 3.A.1-UkrainianWordTokenizer @ a12f5e61 attempt=1 round=3 — SPLIT-% uppercase invent + ABBR_DOT_DASH \\b invent; r1/r2/prior CAP invents fixed; skip thrash
- 2026-07-23T02:49:25Z REJECT: 3.A.1-UkrainianWordTokenizer @ a12f5e61 attempt=1 round=3 — ABBR_DOT_2 BOS invent reintroduced
- 2026-07-23T02:49:25Z blocked: 3.A.1-UkrainianWordTokenizer CAP=3 @ a12f5e61 — skip thrash; revisit after K=5
- 2026-07-23T02:49:26Z CAP-revisit implement: 3.A.1-SpanishWordTokenizer attempt=1 (K≥5; full Java-UCC ORDINAL_POINT)
- 2026-07-23T02:52:49Z ready: 3.A.1-SpanishWordTokenizer @ 32995ed8 attempt=1 (full Java-UCC ORDINAL_POINT; parent green reconfirm)
- 2026-07-23T02:52:49Z validating: 3.A.1-SpanishWordTokenizer @ 32995ed8 attempt=1
- 2026-07-23T02:55:59Z ACCEPT: 3.A.1-SpanishWordTokenizer @ 32995ed8 attempt=1 (validator + parent green; full Java-UCC ORDINAL_POINT)
- 2026-07-23T02:59:26Z implement: 3.A.1-GoogleStyleWordTokenizer (full GoogleStyleWordTokenizerTest twin; hyphen + contraction glue)
- 2026-07-23T03:03:10Z ready: 3.A.1-GoogleStyleWordTokenizer @ c37e38b0 (full testTokenize twin; hyphen + contraction glue; parent green reconfirm)
- 2026-07-23T03:08:50Z validating: 3.A.1-GoogleStyleWordTokenizer @ c37e38b0
- 2026-07-23T03:10:38Z ACCEPT: 3.A.1-GoogleStyleWordTokenizer @ c37e38b0 (validator + parent green reconfirm; full testTokenize twin; hyphen + contraction glue)
- 2026-07-23T03:15:13Z implement: 3.A.1-RussianWordTokenizer (full outcome twin; Java-king; б/у б/н SP_DOT; core WordTokenizer accepted)
- 2026-07-23T03:18:09Z ready: 3.A.1-RussianWordTokenizer @ ce016368 (implementer report; full exact behavior matrix; parent go test ru green)
- 2026-07-23T03:23:39Z validating: 3.A.1-RussianWordTokenizer @ ce016368
- 2026-07-23T03:25:50Z ACCEPT: 3.A.1-RussianWordTokenizer @ ce016368 (validator + parent green reconfirm; full behavior-matrix twin)
- 2026-07-23T03:28:44Z implement: 3.A.1-GermanWordTokenizer (leaf: super delims + "_‚"; full exact behavior matrix; core WordTokenizer accepted)
- 2026-07-23T03:31:39Z ready: 3.A.1-GermanWordTokenizer @ 8daf13f0 (implementer report; full exact matrix + core contrast; parent go test de GermanWordTokenizer green)
- 2026-07-23T03:33:46Z validating: 3.A.1-GermanWordTokenizer @ 8daf13f0
- 2026-07-23T03:34:59Z ACCEPT: 3.A.1-GermanWordTokenizer @ 8daf13f0 (validator + parent green reconfirm; delims super+"_‚" U+201A; inherited tokenize)
- 2026-07-23T03:38:48Z implement: 3.A.1-ArabicWordTokenizer (leaf: super delims + "،؟؛-"; full exact behavior matrix; core WordTokenizer accepted)
- 2026-07-23T03:41:12Z ready: 3.A.1-ArabicWordTokenizer @ 957a5824 (implementer report; full exact matrix + core contrast; parent go test ArabicWordTokenizer green)
- 2026-07-23T03:43:46Z validating: 3.A.1-ArabicWordTokenizer @ 957a5824
- 2026-07-23T03:45:26Z ACCEPT: 3.A.1-ArabicWordTokenizer @ 957a5824 (validator + parent green reconfirm; delims super+"،؟؛-"; inherited tokenize; UK K=5)
- 2026-07-23T03:50:03Z implement: 3.A.1-PersianWordTokenizer (leaf: super delims + "،؟؛"; full exact behavior matrix; core WordTokenizer accepted)
- 2026-07-23T03:51:45Z ready: 3.A.1-PersianWordTokenizer @ 8a2714aa (implementer report; FA delims ،؟؛ no hyphen; parent go test PersianWordTokenizer green)
- 2026-07-23T03:53:47Z validating: 3.A.1-PersianWordTokenizer @ 8a2714aa
- 2026-07-23T03:55:29Z ACCEPT: 3.A.1-PersianWordTokenizer @ 8a2714aa (validator + parent green reconfirm; FA delims ،؟؛ no hyphen; inherited tokenize)
- 2026-07-23T03:58:50Z implement: 3.A.1-SimpleSentenceTokenizer (full SimpleSentenceTokenizerTest testTokenize twin; segment-simple.srx; no invent abbrev)
- 2026-07-23T04:02:41Z ready: 3.A.1-SimpleSentenceTokenizer @ 17f7a3f4 (implementer report; testTokenize twin; invent fallbacks removed; parent go test tokenizers+en/de/fr green)
- 2026-07-23T04:03:51Z validating: 3.A.1-SimpleSentenceTokenizer @ 17f7a3f4
- 2026-07-23T04:05:29Z ACCEPT: 3.A.1-SimpleSentenceTokenizer @ 17f7a3f4 (validator + parent green reconfirm; testTokenize twin; segment-simple.srx official; invent fallbacks removed)
- 2026-07-23T04:09:12Z implement: 3.A.1-TagalogWordTokenizer (leaf: super delims + "-"; full exact behavior matrix; core WordTokenizer accepted)
- 2026-07-23T04:11:12Z ready: 3.A.1-TagalogWordTokenizer @ ba9d62fd (implementer report; full exact matrix + core contrast; parent go test tl green)
- 2026-07-23T04:13:39Z validating: 3.A.1-TagalogWordTokenizer @ ba9d62fd
- 2026-07-23T04:15:40Z ACCEPT: 3.A.1-TagalogWordTokenizer @ ba9d62fd (validator + parent green reconfirm; delims super+"-"; inherited tokenize)
- 2026-07-23T04:19:01Z implement: 3.A.1-KhmerWordTokenizer (leaf: custom StringTokenizer delims + joinEMailsAndUrls; full exact behavior matrix; no Java @Test)
- 2026-07-23T04:22:38Z ready: 3.A.1-KhmerWordTokenizer @ 8e70ba8a (implementer report; full exact matrix + core contrast; parent go test km green)
- 2026-07-23T04:23:43Z validating: 3.A.1-KhmerWordTokenizer @ 8e70ba8a
- 2026-07-23T04:24:53Z ACCEPT: 3.A.1-KhmerWordTokenizer @ 8e70ba8a (validator + parent green reconfirm; hardcoded delims U+17D4/U+17D5 + joinEMailsAndUrls)
- 2026-07-23T04:28:47Z implement: 3.A.1-MalayalamWordTokenizer (leaf: implements Tokenizer; hardcoded delims; NO joinEMailsAndUrls; full exact behavior matrix)
- 2026-07-23T04:31:32Z ready: 3.A.1-MalayalamWordTokenizer @ 2e1a98ff (implementer report; full matrix; no joinEMailsAndUrls; parent go test ml green)
- 2026-07-23T04:33:41Z validating: 3.A.1-MalayalamWordTokenizer @ 2e1a98ff
- 2026-07-23T04:34:41Z ACCEPT: 3.A.1-MalayalamWordTokenizer @ 2e1a98ff (validator + parent green reconfirm; Tokenizer-only; no joinEMailsAndUrls)
