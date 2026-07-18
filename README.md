# lang

Go reimplementation of [LanguageTool](https://github.com/languagetool-org/languagetool) as a **CLI linter**.

**Goal:** 1:1 parity with official LanguageTool data and behavior. Pure Go at lint time (no JVM).

See [SPEC.md](./SPEC.md) for the full product contract.

## Requirements

- [mise](https://mise.jdx.dev/) (or Go 1.26+ and optionally JDK for oracle work)
- LanguageTool submodule (official data + reference)
- English binary dicts (morfologik POS + speller) via Maven artifact `english-pos-dict`

```bash
git clone --recurse-submodules https://github.com/lucasew/lang.git
cd lang
# or after clone:
git submodule update --init --depth 1
./scripts/fetch-english-dicts.sh   # english.dict + en_US.dict etc.

mise install
mise exec -- go test ./...
mise exec -- go run ./cmd/lang doctor
```

## Usage

```bash
# product CLI (Cobra)
go run ./cmd/lang doctor
go run ./cmd/lang languages
go run ./cmd/lang rules --lang en
go run ./cmd/lang lint --lang en path/to/file.txt other.txt
echo 'This is an test.' | go run ./cmd/lang lint --lang en --format text
go run ./cmd/lang lint --format sarif --lang en file.txt
go run ./cmd/lang golden --lang en - > findings.json
go run ./cmd/lang compare findings.json --lang en -

# legacy LT-style flags still work
go run ./cmd/lang -l en --lint -

# HTTP API
go run ./cmd/lang-server -port 8081 -public
```

### Data path

1. `--data-dir`
2. `LANG_DATA`
3. `./inspiration/languagetool`

## Status

| Area | State |
|------|--------|
| CLI (`lang lint`, formats, exit codes) | done |
| Data resolve + language discovery | done |
| SRX sentence split (`segment.srx`) | done |
| Word tokenizer | done |
| Morfologik FSA (CFSA2) + dictionary lookup | done |
| English tagger (`english.dict`) | done |
| English speller (`MORFOLOGIK_RULE_EN_US`) | done (CFSA2 edit-1 + soft typos TSV) |
| Pattern XML + POS/inflected match | done (filters/unify/AI incomplete skipped) |
| `WHITESPACE_RULE` / `WORD_REPEAT_RULE` | done |
| Layout (sentence/punct/paragraph whitespace, unpaired, uppercase) | done |
| Soft grammar packs (`testdata/grammar/*-soft.xml`) | 35+ packs; en-US / en-GB soft spelling variants; CoreGoldenHook matrix |
| Soft false friends (`-m` + false-friends-soft.xml) | CoreGoldenHook |
| EN speller (`en_US.dict` CFSA2 when present) | MORFOLOGIK_RULE_EN_US |
| EN POS tagger (`english.dict` CFSA2 when present) | TagWord / `--taggeronly` |
| EN soft multiword disambiguator | MultiWordChunker on Analyze |
| Soft EN XML disambiguation | `testdata/disambiguation/en-soft.xml` (filter/replace/immunize) |
| Soft EN ignore-spelling list | `testdata/disambiguation/en-ignore-spelling.txt` |
| Soft EN multiwords | `testdata/disambiguation/en-multiwords-soft.txt` |
| CLI disambig data | `--ignore-spelling-file`, `--disambiguation-file` |
| Demo EN speller (`LANG_DEMO_SPELLER=1` fallback) | map + edit-distance suggestions |
| Soft EN typos TSV | `testdata/spelling/en-typos.tsv` suggestions |
| `--apply` suggestion rewrite | a/an, false friends, soft patterns |
| `--ignore-words` CSV | suppress spelling matches |
| Disambiguator | soft hybrid (multiwords + XML filter/replace/immunize); full pipeline later |
| Full 1:1 goldens | growing |

## License

This project’s own code: see repository license when published.

**LanguageTool** under `inspiration/languagetool` remains under its upstream licenses (LGPL). Official rule/data files are not re-licensed by this port.

## Vendoring upstream testdata

Do **not** invent soft rules or golden strings. Official LT data is the source of truth.
See **[docs/soft-port-policy.md](docs/soft-port-policy.md)** for the anti-cheat rules:
every soft behavioral change must cite original Java/LT code (or vendored extracts);
logic must be essentially the same as upstream.

```bash
python3 scripts/vendor-lt-testdata.py --langs en
```

This copies grammar/style/disambiguation/multiwords into `testdata/upstream/`,
extracts soft-loader-compatible surface patterns to `testdata/grammar/*-upstream-soft.xml`,
and writes example goldens to `testdata/upstream/goldens/`. Run
`TestGolden_UpstreamENExamples` (set `LANG_UPSTREAM_GOLDEN_ALL=1` for the full matrix).

Main soft miss scan: `LANG_EN_MISS_SCAN=1 go test ./internal/languagetool/org/languagetool/commandline/ -run TestDebugENMissScan -v`  
Optional soft miss scan: `LANG_EN_OPT_MISS_SCAN=1 go test ... -run TestDebugENOptMissScan -v`
