# lang

Go reimplementation of [LanguageTool](https://github.com/languagetool-org/languagetool) as a **CLI linter**.

**Goal:** 1:1 parity with official LanguageTool data and behavior. Pure Go at lint time (no JVM).

See [SPEC.md](./SPEC.md) for the full product contract.

## Requirements

- [mise](https://mise.jdx.dev/) (or Go 1.26+ and optionally JDK for oracle work)
- LanguageTool submodule (official data + reference)

```bash
git clone --recurse-submodules https://github.com/lucasew/lang.git
cd lang
# or after clone:
git submodule update --init --depth 1

mise install
mise exec -- go test ./...
mise exec -- go run ./cmd/lang doctor
```

## Usage

```bash
# lint a file
mise exec -- go run ./cmd/lang lint --lang en-US path/to/file.txt

# stdin
echo 'This  is a test.' | mise exec -- go run ./cmd/lang lint --lang en

# formats
mise exec -- go run ./cmd/lang lint --format json --lang en file.txt
mise exec -- go run ./cmd/lang lint --format sarif --lang en file.txt

# list languages from official tree
mise exec -- go run ./cmd/lang languages
```

### Data path

1. `--data-dir`
2. `LANG_DATA`
3. `./inspiration/languagetool`

## Status

Initial scaffold:

| Area | State |
|------|--------|
| CLI (`lang lint`, formats, exit codes) | done |
| Data resolve + language discovery | done |
| Pipeline architecture | staged; partial implement |
| `WHITESPACE_RULE` (MultipleWhitespaceRule) | done (LT test positions) |
| Tagger / disambiguator / pattern XML | not yet |
| Full language matrix goldens | not yet |

## License

This project’s own code: see repository license when published.

**LanguageTool** under `inspiration/languagetool` remains under its upstream licenses (LGPL). Official rule/data files are not re-licensed by this port.
