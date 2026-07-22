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
| docs/faithful-port-checklist.md#3.A.3-MultiWordChunker-core-test | ready | inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/MultiWordChunker.java; inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/tagging/disambiguation/MultiWordChunker2.java; inspiration/languagetool/languagetool-core/src/test/java/org/languagetool/tagging/disambiguation/MultiWordChunkerTest.java; inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/resource/yy/multiwords.txt | internal/languagetool/org/languagetool/tagging/disambiguation/ | 3b4841a422994d3fbf96a5911960cdcd15c6a4e3 | 0 | 0 | | | 2026-07-22T16:53:53Z |

---

## Human inbox

**You only need to look when the loop is idle** — no `ready`, no `rejected` under CAP, no implementable leaf, only `blocked` and/or `accepted`.

| When idle | Summary |
|-----------|---------|
| Last idle at | _(never)_ |
| Blocked lines | _(none)_ |
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

