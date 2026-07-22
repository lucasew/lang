# lang — agent instructions

Go 1:1 port of LanguageTool. Pure Go at lint time. **Java under `inspiration/languagetool` is king of correctness.**

## Commands

- Test (scoped): `go test ./internal/languagetool/...` or package path under change
- Full test: `mise exec -- go test ./...` (or `go test ./...`)
- CLI: `go run ./cmd/lang doctor`

## Layout

- `internal/languagetool/` — **faithful transcription only** (mirrors Java `org/languagetool/...`)
- `inspiration/languagetool/` — upstream Java reference + official data (submodule)
- `cmd/` — product CLI (frozen for port work unless human lifts freeze)
- `testdata/`, `scripts/` — **frozen** while faithful-port freeze holds

## Boundaries

**Always**

- Read and follow [`docs/faithful-port-policy.md`](docs/faithful-port-policy.md) before changing `internal/languagetool/`.
- Work **one Java type per sector**, **leaves → root**, no stubs.
- Load the **same** dictionaries / XML / models as Java.
- After implementer finishes a sector: **validator** must ACCEPT (implementer cannot self-ACCEPT and **never** marks the checklist).
- Unattended port loop: read and obey [`docs/loop-protocol.md`](docs/loop-protocol.md); queue source of truth is [`docs/validation/queue.md`](docs/validation/queue.md) (**parent-only** writes).
- Treat Go ≠ Java as a **bug in Go**.

**Ask first**

- Lifting the outside-tree freeze.
- Changing product CLI contract ([`SPEC.md`](SPEC.md)).
- Adding dependencies or new top-level packages.

**Never**

- Soft / invent / approximate logic inside `internal/languagetool/`.
- Edit outside `internal/languagetool/` during faithful-port work (goldens, vendor scripts, demos) unless a human explicitly lifts the freeze — **exception:** parent may update `docs/validation/*`, checklist marks after validator ACCEPT, and this protocol as directed by a human.
- Use soft goldens or miss-scan % as proof of 1:1 fidelity.
- Stub dependencies so a type “compiles” with wrong behavior.
- Hardcoded name blocklists as the anti-cheat strategy (path law + reviewer + Java-king).
- Freestyle “continue checklist” loops without rehydration + decision table in `docs/loop-protocol.md`.
- Implementer writing checklist `[x]` or queue status.

## More context (read when relevant)

- [`docs/loop-protocol.md`](docs/loop-protocol.md) — **required** for unattended implementer/validator parent loop (P2)
- [`docs/validation/queue.md`](docs/validation/queue.md) — durable ready/reject/accepted queue (parent-only)
- [`docs/faithful-port-policy.md`](docs/faithful-port-policy.md) — **required** for any `internal/languagetool` work (implementer + validator)
- [`docs/faithful-port-checklist.md`](docs/faithful-port-checklist.md) — **master TODO** of everything that must be checked
- [`SPEC.md`](SPEC.md) — product CLI contract
- [`README.md`](README.md) — setup and usage

User instructions override this file for the session; they do not rewrite repo policy unless the user is explicitly changing policy.
