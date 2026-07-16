# lang — Specification

Decisions locked in grilling (2026-07-15). This is the source of truth for product and process goals.

## 1. Purpose

Go reimplementation of [LanguageTool](https://github.com/languagetool-org/languagetool): a **CLI linter** (`lang lint`) that accepts language + text and reports grammar, spelling, and style issues.

**Goals:**

- **1:1 behavioral parity** with upstream LanguageTool when running on the **same official data** (rule IDs, messages, suggestions, spans, severities).
- **Pure Go runtime** — no JVM, no LanguageTool server, no shelling out to Java when linting.
- **Kitchen sink** — multi-language; architecture and data loading are not English-only.

**Not goals (runtime):**

- Shipping or requiring a Java service.
- A simplified “inspired by LT” rule set with different IDs/messages.
- Stubbing pipeline stages in a way that breaks parity (no fake tagger forever, no tokenizer-only cosplay).

**Dev-only Java is allowed** to generate goldens / compare against the oracle; day-to-day `lang lint` and default `go test` should not need a JVM once fixtures exist.

## 2. Product surface

### 2.1 Identity

| Item | Value |
|------|--------|
| Binary | `lang` |
| Go module | `github.com/lucasew/lang` |
| Primary command | `lang lint` |
| CLI stack | **Cobra** + **Viper** |

### 2.2 `lang lint`

**Input**

- Text from **file paths** and/or **stdin** (no paths → stdin; `-` may mean stdin).
- **Language:** `--lang` / `-l`, default **`auto`**. Explicit language code always wins over detection.
- **Data root:** see §5.

**Output (finding contract)**

Each issue exposes at least:

| Field | Notes |
|-------|--------|
| `rule` | Stable upstream rule ID (1:1 with LT) |
| `type` | LanguageTool **ITS issue type** for the match (e.g. `misspelling`, `whitespace`, `grammar`, `style`, `duplication`). 1:1 with LT’s `locQualityIssueType` serialization (lowercase / hyphenated). |
| `severity` | **SARIF 2.1** result level only: `error` \| `warning` \| `note` \| `none`. Derived from `type` (see below). Used for CI exit codes and SARIF `result.level`. |
| `message` | 1:1 with LT message text |
| `location` | `file:line:col` (standard linter) |
| `suggestion` | When LT provides one (or more; text mode may show primary / first) |

**Severity mapping (`type` → SARIF `severity`)**

| `type` (ITS) | `severity` |
|--------------|------------|
| `misspelling`, `grammar` | `error` |
| `style`, `register`, `locale-violation`, `locale-specific-content` | `note` |
| everything else (`whitespace`, `typographical`, `duplication`, `other`, …) | `warning` |

Do **not** put ITS types in `severity`. Do **not** invent parallel severity systems beyond SARIF levels.

**Formats (`--format`)**

| Value | Role |
|-------|------|
| `text` (default) | Tabwriter columns: `location`, `severity`, `type`, `rule`, `message`, `suggestion`. |
| `sarif` | SARIF 2.1.0; `result.level` = `severity`; ITS `type` in `result.properties.type`. |
| `json` | Machine-readable findings (includes `severity` and `type`); useful for goldens/dev. |

Additional formats may be added later; do not block on a large format zoo.

**Exit codes**

| Code | When |
|------|------|
| `0` | Ran successfully; no findings with SARIF severity **`error`** (`warning` / `note` OK). |
| `1` | At least one finding with severity **`error`**. |
| `2` (preferred) | Usage / I/O / engine failure (tool did not complete a normal lint). |

Optional later: `--fail-on=warning` (or similar) to fail on lower severities. Default remains “0 on warning.”

**Flags (v1)**

- Control is **flags-only** for project config (no required `.lang.toml` in v1).
- Viper binds flags and env where natural.
- Expected family (names may be refined at implement time): `--lang`, `--format`, `--data-dir`, enable/disable/only rule filters, and whatever else the port needs.
- **No project config file product** until we deliberately add it later.

### 2.3 Other commands

- **`lang lint` is the product.**
- Extra Cobra commands (`languages`, `rules`, `golden`, `doctor`, `compare`, …) are **allowed for development** and may be cleaned up or hidden later.
- Prefer keeping oracle/golden generators easy to run for agents; they may depend on JDK via `mise` when regenerating fixtures.

## 3. Correctness / “done”

### 3.1 Parity meaning (1:1)

For a given `(language, text, data revision, equivalent options)`:

| Layer | Requirement |
|-------|-------------|
| Rule ID | Same as LT |
| Message | Same as LT |
| Suggestions | Same as LT (order and strings) |
| Span | Same character/token offsets as LT (document any intentional tokenization edge cases) |
| Type (ITS) | Same as LT `locQualityIssueType` |
| Severity (SARIF) | Derived from type via the mapping table in §2.2 (not a second LT field) |

Not “similar category in the same region.” Goldens from upstream are law.

### 3.2 Pipeline (no shortcuts)

Faithful port of LanguageTool’s real analysis chain, not a reduced toy:

**text → sentence split → tokenize → tag (POS/morph) → chunk → disambiguate → pattern/rules → filters → suggestions**

- Every stage exists as a real component in the architecture.
- Stages are implemented by **mirroring upstream behavior and consuming official data**, not by inventing a parallel linguistics stack.
- **Chunking (en):** currently a POS-driven BIO heuristic plus NP singular/plural refinement (LT’s `EnglishChunkFilter` shape). Full OpenNLP maxent models remain a plateau toward 1:1.
- Claiming parity for a language pack requires goldens for that pack to pass — not merely “XML parses.”

### 3.3 Official data

- Engine runs on **upstream LanguageTool resources** (rules XML, messages, dictionaries, disambiguation, etc.).
- We do **not** maintain a hand-written duplicate catalog of kitchen-sink rules as the source of truth.
- Data lives in-tree for now via **`inspiration/languagetool`** (git clone / submodule). See §5.

### 3.4 Tests and goldens

| Mechanism | Role |
|-----------|------|
| LT as oracle | Run official LT (Java) in **dev** to produce/verify expected matches |
| Translated fixtures | Store goldens in **our** harness format under the repo |
| `go test` | Assert 1:1 against committed fixtures (no JVM required for the default suite once fixtures exist) |
| Regen | Documented path (tool command or `go run ./tools/...`) may need JDK |

**Policy:** steal goldens, translate tests; do not reimplement LT’s entire JUnit runner unless it pays rent. Fixture format should record enough to assert full parity fields (§3.1).

### 3.5 Coverage vs design

| Axis | Bar |
|------|-----|
| **Design** | Full LT-class port, all languages the data supports, 1:1 |
| **Calendar** | Progress is incremental; green pack-by-pack / suite-by-suite |
| **Honesty** | Do not advertise parity for packs that fail goldens |

Kitchen-sink **architecture and loading** from day one; kitchen-sink **green test matrix** grows over time.

## 4. Languages

- **All languages** present in official data are in scope.
- Loader and CLI are multi-language; no English-only product design.
- `--lang auto`: detect among **loaded** packs; must not silently fall back to a wrong language without a clear policy (prefer explicit detection confidence / error over lying).
- Explicit `--lang` is authoritative for CI and goldens.

## 5. Data location

**v1 resolution order** (iterate later for convenience):

1. **`--data-dir`** (flag, if set)
2. **`LANG_DATA`** environment variable
3. Default: **`./inspiration/languagetool`** (cwd-relative)

Missing/invalid data → **clear error** (how to set `LANG_DATA` / init submodule), not an empty successful lint with zero rules.

**Distribution (now):** git submodule or full clone under `inspiration/languagetool`. Heavy but acceptable until a better packaging approach (extract, fetch, embed) is chosen deliberately.

**License:** LanguageTool data and upstream code remain under **their** licenses (typically LGPL for LT). Preserve attribution; do not relicense upstream assets. Document this when README exists.

## 6. Engine architecture (intent)

### 6.1 Strategy

- **Interpret / execute official LT resources in Go** — same implementation shape as upstream, different language.
- Not: hand-port thousands of rules into a custom DSL as the primary path.
- Not: codegen-only without a faithful runtime (codegen may appear later as an optimization, not a correctness escape hatch).

### 6.2 Mirroring discipline

When behavior diverges from goldens:

1. Find the corresponding upstream Java path (tokenizer, tagger, disambiguator, rule match, filter, message formatting).
2. Align Go control flow and data use to that path.
3. Prefer structural fixes that generalize across languages/rules over single-fixture hacks.

### 6.3 Out of scope for “shortcuts”

Rejected even if demos look good:

- Empty/stub tagger or disambiguator left in place while claiming 1:1
- Different rule IDs or paraphrased messages
- Matching only “some issue near the same word” without span/ID/message parity
- Runtime dependency on Java LT

## 7. Tooling / repo

| Item | Notes |
|------|--------|
| `mise.toml` | Pins toolchain (e.g. Go, Java for oracle/dev) |
| `inspiration/languagetool` | Upstream tree (submodule/clone); read-only inspiration + data + oracle source |
| Module path | `github.com/lucasew/lang` |
| Layout | Conventional Go (`cmd/lang`, internal engine packages, `testdata/`, optional `tools/`) — exact tree left to implementation as long as SPEC holds |

## 8. Process

| Rule | Detail |
|------|--------|
| Source of truth | This `SPEC.md`; update it when product decisions change |
| Parity metric | Count of goldens/packs passing 1:1; first-class over vibe demos |
| Driver | Implement pipeline + data load → golden → fix divergence against Java source/data |
| Plateau | Read upstream; do not invent parallel linguistics |
| Dev commands | Allowed; clean up later |
| Data UX | Default path + `LANG_DATA` is enough for v1; improve discovery later |

## 9. Milestone sketch (non-binding order)

1. Repo skeleton: module, `cmd/lang`, cobra `lint`, flags, text/json format stubs, exit codes.
2. Data resolution (`--data-dir` / `LANG_DATA` / default path) and pack discovery.
3. Pipeline interfaces + first real stages wired end-to-end on official files.
4. Oracle/golden tool path; first 1:1 fixtures (start wherever upstream is easiest; expand to all langs).
5. Grow stages (tagger, disambiguator, rule XML, filters) until goldens pass; broaden language matrix.
6. SARIF polish, rule enable/disable flags, auto-lang quality, packaging convenience.

Milestones do not weaken §3; they sequence work.

## 10. Open iterations (explicitly deferred)

- Nicer data discovery (walk-up, binary-relative, fetch).
- Project config file (e.g. `.lang.toml`).
- `--fail-on` severity threshold.
- Replacing submodule with vendored extract / download / embed.
- Trimming or stabilizing dev-only commands.
- Full CI matrix for every upstream language pack.

---

**Locked one-liner:** *LanguageTool’s brain and data, Go CLI linter UX, 1:1 goldens, no JVM at lint time.*
