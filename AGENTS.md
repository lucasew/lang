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
- After implementer finishes a sector: **reviewer agent** must ACCEPT (implementer cannot self-ACCEPT).
- Treat Go ≠ Java as a **bug in Go**.

**Ask first**

- Lifting the outside-tree freeze.
- Changing product CLI contract ([`SPEC.md`](SPEC.md)).
- Adding dependencies or new top-level packages.

**Never**

- Soft / invent / approximate logic inside `internal/languagetool/`.
- Edit outside `internal/languagetool/` during faithful-port work (goldens, vendor scripts, demos) unless a human explicitly lifts the freeze.
- Use soft goldens or miss-scan % as proof of 1:1 fidelity.
- Stub dependencies so a type “compiles” with wrong behavior.
- Hardcoded name blocklists as the anti-cheat strategy (path law + reviewer + Java-king).

## More context (read when relevant)

- [`docs/faithful-port-policy.md`](docs/faithful-port-policy.md) — **required** for any `internal/languagetool` work (implementer + reviewer checklist)
- [`SPEC.md`](SPEC.md) — product CLI contract
- [`README.md`](README.md) — setup and usage

User instructions override this file for the session; they do not rewrite repo policy unless the user is explicitly changing policy.
