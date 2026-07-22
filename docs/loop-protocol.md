# Unattended port loop protocol

**Durable law for implementer ↔ validator work toward a 1:1 Java→Go LanguageTool port.**  
Survives chat compaction: agents must re-read this file (and the queue) every parent fire.

Java under `inspiration/languagetool` is **king of correctness**.  
Faithful-port rules in [`faithful-port-policy.md`](faithful-port-policy.md) still apply inside `internal/languagetool/`.

---

## 1. Goal

- **Product of trust:** checklist line marked done only after **behavior equivalence** to Java is proven.
- **Unit of mark:** one line in `CHECKLIST.md` and/or [`faithful-port-checklist.md`](faithful-port-checklist.md) (as applicable to the sector).
- **Quality bar:** twin tests assert **Java-visible outcomes** (matches, tags, spans, suggestions, errors, control-flow results). Not name-only twins, not “no panic,” not empty fail-closed smoke as sole proof.

---

## 2. Roles (three agents)

Subagents **cannot** message each other. The **parent is the only bus**.

| Role | Does | Does not |
|------|------|----------|
| **Parent** | Owns queue writes; owns checklist `[x]` application after validator ACCEPT; spawns kids; decision table; single-flight lock; skip/revisit/blocked policy | Port logic “just this once”; self-grade invent; freestyle without rehydration |
| **Implementer** | Transcribe leaf sector; twins; delete invent; report Ready or blocked reason | Edit checklist marks; write queue status; talk to validator |
| **Validator** | Audit ready line vs Java (checks 1–7); ACCEPT or REJECT with findings | Implement production fixes to paper over FAIL; mark without audit |

**Implementer never writes the checklist.**  
**Only the parent** mutates `docs/validation/queue.md` and applies checklist marks (after validator ACCEPT).

---

## 3. Concurrency

| Resource | Rule |
|----------|------|
| Implementer | **Exactly one** active at a time |
| Validators | **Up to N** in parallel (default **N = 2**) on **disjoint** ready lines (no overlapping `go_paths`) |
| Parent | One orchestration turn at a time (**parent lock**) |

---

## 4. P2 fire model (unattended)

Cadence: **every 5 minutes** (scheduler), with **parent lock** so overlapping fires no-op.

Each parent wake does **exactly one step** from the decision table (not a full multi-hour crusade):

- **implement** — one leaf for one checklist line, or  
- **validate** — one or more ready lines (≤ N validators), or  
- **fix** — only lines under reject cap, or  
- **idle** — nothing productive left (see human return)

Do **not** use a vague prompt like “continue the checklist” without rehydration.

---

## 5. Mandatory rehydration (every parent fire)

**Before any spawn or code change**, parent must:

1. Read **this file** (`docs/loop-protocol.md`).
2. Read **[`docs/validation/queue.md`](validation/queue.md)** (source of truth for status).
3. Read sector brief fields on the chosen line (paths, commit, findings).
4. Apply the **decision table** below.
5. If lock held by a live parent → **exit** (no work).

**Hard stop:** if nothing is `rejected` (under cap), nothing is `ready`, and no implementable unchecked leaf remains (only `blocked` / `accepted` / empty) → **idle**. Do **not** invent horizontal work (random UTF-16/trim waves, unrelated packages).

---

## 6. Decision table

```text
if parent lock held (and not stale)     → exit
if any rejected with round < CAP        → step = fix (those lines only; one implementer)
else if any ready                       → step = validate (FIFO among ready; ≤ N disjoint)
else if implementable unchecked leaf    → step = implement (pick per §7)
else if blocked due for revisit         → step = implement (revisit attempt; see §9)
else                                    → idle (no progress anywhere; human may look)
```

**CAP** (validate→reject cycles per attempt): **3**.  
After CAP → status `blocked` (reason required); **skip**, do not thrash; **revisit later** (§9).

---

## 7. Work selection (implement)

When free to implement:

1. Start from an **unchecked** checklist line (not `accepted` / not already `ready` / `validating`).
2. **Follow references / dependencies** until a **leaf**.
3. Parent sets queue row to `implementing` **before** spawn (prevents double-pick).
4. Implementer works that leaf (and only claims **ready** for the checklist line when that line’s required scope is complete enough to audit).

Eager work on other files is allowed in the tree, but **validator only sees lines the implementer declared ready** (via report → parent sets `ready`).

---

## 8. Queue statuses and transitions

Statuses (only these):

```text
open → implementing → ready → validating → accepted
                               ↘ rejected → implementing → ready → …
```

Also:

- **`blocked`** — cap hit or hard external miss (resource freeze, missing official asset). Skipped until revisit.
- **`deferred`** — optional alias while waiting revisit window (or use `blocked` + `revisit_after`).

Parent-only transitions. Kids **report**; parent writes.

| Field | Purpose |
|--------|---------|
| `checklist_id` | Stable id of the checklist line |
| `status` | See above |
| `java_paths` | Java / resource paths under `inspiration/languagetool` |
| `go_paths` | Go paths under `internal/languagetool` |
| `ready_commit` | SHA implementer claims ready |
| `findings` | Validator REJECT list (Java ref + Go symptom + required fix) |
| `round` | Reject count this attempt |
| `attempt` | Revisit generation (increments when leaving blocked for a new try) |
| `updated_at` | ISO time of last parent update |
| `blocked_reason` | Required when `blocked` |

Schema and live rows: [`docs/validation/queue.md`](validation/queue.md).

---

## 9. Stuck, skip, revisit, human

| Situation | Action |
|-----------|--------|
| REJECT, `round < CAP` | Fix that line (implementer) |
| REJECT, `round ≥ CAP` or missing official resource / freeze wall | `blocked` + reason; **skip** to other work |
| Revisit | After **K = 5** productive steps elsewhere (accept or new ready) **or** next calendar day of loop runtime, allow one new attempt (`attempt++`, `round` reset) |
| Soft pass | **Forbidden** |
| Human looks | When **no progress remains anywhere** (idle): only `blocked` / `accepted` / empty productive queue — see queue “Human inbox” section |

Do **not** page the human on every blocked line.

---

## 10. Validator audit (all mandatory for ACCEPT)

1. **Scope** — `java_paths` / `go_paths` match the checklist claim.  
2. **Twin existence** — every Java `@Test` (or agreed behavior matrix) for scope has a Go twin that **runs**.  
3. **Outcome fidelity** — twins assert Java-visible results (not smoke-only).  
4. **Invent scan** — no soft/invent path where Java has real logic/resources.  
5. **Resources** — same official dict/XML/model paths as Java (else REJECT / blocked-class reason).  
6. **Green tests** — scoped `go test` for touched packages green.  
7. **Mark** — only full pass: parent sets queue `accepted` and checklist `[x]`.

**REJECT** must cite: Java location + Go location + deviation.  
Validator does **not** implement production code to force green.

---

## 11. Handoff reports (kids → parent)

### Implementer report (required fields)

- `checklist_id`
- Result: `ready` | `blocked` | `partial` (eager only; not queue-ready)
- `java_paths` / `go_paths`
- `ready_commit` (if ready)
- Notes / blockers

### Validator report (required fields)

- `checklist_id`
- `ready_commit` audited
- Result: `ACCEPT` | `REJECT`
- Findings list (empty on ACCEPT)
- Tests run

Parent applies reports to the queue; never trust a kid to edit the queue file.

---

## 12. Parent lock

- Path: `docs/validation/parent.lock` (pid + ISO timestamp + step).
- Acquire at start of fire; release on exit.
- Stale lock: if timestamp older than **45 minutes**, parent may steal (previous fire died).
- Overlapping scheduler fire: if lock fresh → **exit 0, no work**.

---

## 13. Scheduler prompt (canonical)

Use this (or equivalent) for unattended loop — **not** “continue checklist”:

```text
You are the PARENT of the faithful port loop. Obey docs/loop-protocol.md.
Rehydrate: read docs/loop-protocol.md and docs/validation/queue.md first.
Acquire docs/validation/parent.lock or exit if held.
Run the decision table for exactly ONE step (implement XOR validate XOR fix XOR idle).
Spawn at most one implementer, or up to N=2 validators on disjoint ready lines.
Parent-only queue and checklist writes. Implementer never marks checklist.
Commit on real progress. No invent. Java is king. No freestyle horizontal waves.
If idle (no progress anywhere), update Human inbox in the queue and stop.
```

---

## 14. Relation to faithful-port-policy

| Policy term | Loop term |
|-------------|-----------|
| Reviewer ACCEPT | Validator ACCEPT (checks §10) |
| One Java type sector | Leaf under a checklist line; line is mark unit |
| Implementer cannot self-ACCEPT | Implementer never checklist; parent marks only after validator |

Unattended automation **must** use this loop protocol. Ad-hoc “commit and continue” without queue/validator is **out of process**.
