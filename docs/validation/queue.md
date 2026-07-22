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
| _(none)_ | | | | | | | | | |

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
