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
| English speller (`MORFOLOGIK_RULE_EN_US`) | done (suggestions later) |
| Pattern XML + POS/inflected match | done (filters/unify/AI incomplete skipped) |
| `WHITESPACE_RULE` / `WORD_REPEAT_RULE` | done |
| Layout (sentence/punct/paragraph whitespace, unpaired, uppercase) | done |
| Soft EN patterns + style (long sentence/paragraph) | growing goldens |
| Disambiguator | not yet |
| Full 1:1 goldens | growing |

## License

This project’s own code: see repository license when published.

**LanguageTool** under `inspiration/languagetool` remains under its upstream licenses (LGPL). Official rule/data files are not re-licensed by this port.
