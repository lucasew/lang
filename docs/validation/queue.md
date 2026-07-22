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
| docs/faithful-port-checklist.md#3.A.1-PolishSentenceTokenizer | ready | inspiration/languagetool/languagetool-language-modules/pl/src/test/java/org/languagetool/tokenizers/pl/PolishSentenceTokenizerTest.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tokenizers/SRXSentenceTokenizer.java; inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/segment.srx; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/TestTools.java | internal/languagetool/org/languagetool/tokenizers/pl/; internal/languagetool/org/languagetool/tokenizers/ | 7c671115979b1fecfaabc31ff3903464933b5c8a | 0 | 0 | | | 2026-07-22T19:20:15Z |

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

