# Faithful port policy (anti-cheat)

**Source of truth for `internal/languagetool`.**  
Java LanguageTool under `inspiration/languagetool` is the **king of correctness**.  
This document supersedes any prior “soft port” policy. Soft approximations are **not** valid inside the wall.

## 1. The wall

| Zone | Rule |
|------|------|
| **`internal/languagetool/**`** | Faithful **transcription** only: same structure, same algorithms, **bug-for-bug** behavior as Java LT. |
| **Outside `internal/languagetool`** | **Frozen for now** — do not change (no goldens, vendor scripts, demos, harnesses, or soft glue). |

Anything that is not a 1:1 port of LT logic does **not** belong under `internal/languagetool`.

## 2. Transcription standard

1. **Function-by-function, bug-for-bug** — including Java quirks that affect output. Do not “fix” upstream bugs inside the port.
2. **Same structure** — each Go type maps to one Java type under `inspiration/languagetool/...` (package layout mirrors `org/languagetool/...`). Thin Go load/bootstrap with no Java twin is the only allowlisted exception and must stay minimal.
3. **Same resources** — load the **same** dictionaries, XML packs, and models Java uses. Soft extracts, invent packs, and hand-written substitutes are **not** engine input.
4. **Same results** — same text + same language + same resources → same analysis and matches as JVM LanguageTool. No soft scores, no “≥N%”, no knobs that let divergence look green.
5. **Delete fakeness** — soft/approximate/invent branches in a sector are **removed**, not renamed. Only proven 1:1 code may remain or enter.
6. **No stubs** — a type may be ported only when its real dependencies already exist as faithful twins (or pure stdlib equivalents). No temporary stubs that return empty/wrong “success.”

## 3. Path law (structure fence)

- New or kept production code under `internal/languagetool` **must** correspond to a real path under `inspiration/languagetool` (Java source or official resource layout as Java loads it).
- Do **not** grow parallel soft APIs, helper packages, or “approx” pipelines inside the tree.
- **No hardcoded name blocklists** (e.g. banning the word `soft`) — they are rename-cheatable. Path law + reviewer + Java-king behavior are the fences.

## 4. Work model: sector → implementer → reviewer

### Sector

- **One unit of work = one Java type** (Go twin file(s) + colocated tests for that type).
- **Order: leaves → root** — port low-dependency types first; composites only after dependencies are faithful twins.

### Implementer

In-bounds for a sector (only under `internal/languagetool`):

- Transcribe the Java type (algorithm and control flow).
- Wire **real** LT resource loading where that type needs it.
- **Delete** soft/fake code in that sector.
- Colocated tests for that type when needed for the twin.

**Out of bounds:** anything outside `internal/languagetool`; inventing behavior; soft goldens as proof; stubs; expanding scope to multiple unrelated types.

### Reviewer agent

Mandatory two-phase loop for agent work on the wall:

1. Implementer finishes **one sector**.
2. **Reviewer agent** checks this policy (and path law / Java twin).
3. **REJECT** → implementer fixes the **same** sector and resubmits.
4. Loop until **ACCEPT** or **human abort**.
5. Implementer **must not** self-ACCEPT.

Reviewer **REJECT** if any of:

- No clear Java twin (path + type/method) for the change.
- Algorithm or control flow diverges from Java without being an unavoidable Go mapping of the same logic.
- Soft, approximate, invent, or golden-driven behavior.
- Stubs or fake dependencies.
- Loads non-Java / soft-only resources as the real engine input.
- Sector too large (more than one Java type / unfocused).
- Touches files outside `internal/languagetool` (while freeze holds).
- Claims “parity” from soft metrics or edited expectations instead of Java behavior.

Reviewer **ACCEPT** only when the sector is a credible 1:1 twin and fakeness in that sector is gone.

### Proven to enter (current bar)

A type may land when:

1. It maps to a Java twin under `inspiration/languagetool`, and  
2. The **reviewer ACCEPT**s.

Soft goldens / miss-scan percentages are **not** proof. Product-level claim “same results on any text” still means exact match with JVM LT on the corpus when end-to-end parity is asserted — Java remains king; Go differences are **bugs in Go**.

## 5. Corpus and regressions

- **Upstream corpus:** official LT examples, fixtures, and public regressions, as upstream provides them.
- **Our regressions:** cases where Go does **not** yet match Java. These track **Go bugs**. Expected side is Java’s behavior — never rewrite the expectation to match wrong Go output.
- No metric flexibility that can be used to cheat (no partial-credit gates as “parity”).

## 6. Explicitly forbidden inside the wall

- Invented multiwords, closed-class lists, POS remaps, or rules with no Java/data source.
- Soft surface probes that Java does not perform.
- “Fix the golden” or weaken logic so a non-faithful path passes tests.
- Alternate pipelines injected into the transcription (soft hybrid seams, fake disambiguators) presented as the Java engine.
- Quarantine skips or percentage scores that greenwash divergence from Java.

## 7. Outside the wall (later)

When the freeze lifts (human decision only):

- Glue, demos, and incomplete-asset UX may live **outside** `internal/languagetool`.
- They must not be imported by the faithful engine or inject fake stages into it.
- Until freeze lifts: **do not edit outside** to make the port look better.

## 8. Related

- Product contract: [`SPEC.md`](../SPEC.md)
- Java reference tree: `inspiration/languagetool/`
- Test twin audit (when used): `scripts/check_lt_test_twins.py`, `internal/languagetool/twin_audit_test.go`
