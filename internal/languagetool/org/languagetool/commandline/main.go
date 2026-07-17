package commandline

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// UsageText is the CLI help banner (ports Main usage surface).
const UsageText = `Usage: lang [lint|languages|version|help] [OPTION]... [FILE]
       languagetool [OPTION]... [FILE]

Soft product subcommands (SPEC §2):
  lint                     check text with linter columns (same as --lint)
  languages                list languages (same as --list)
  rules                    list registered rule IDs for -l language
  golden                   dump SPEC findings JSON (goldens)
  compare GOLDEN.json      compare live findings to a golden file
  doctor                   environment / self-check diagnostics
  version                  print version
  help                     this help

Options:
  -l, --language CODE      language code (e.g. en-US)
  -m, --mothertongue CODE  mother tongue for false friends
  -d, --disable RULES      comma-separated disabled rule ids
  -e, --enable RULES       comma-separated enabled rule ids
  -t, --taggeronly         only tag the text
  -a, --apply              apply first suggestions to text, print result
  -v, --verbose            verbose output
  --json                   JSON output
  --xml                    XML output
  --sarif                  SARIF 2.1 output
  --lint                   linter columns (location severity type rule message suggestion)
  --format FMT             output format: text|lint|json|sarif|xml (text≡lint per SPEC)
  --xmlfilter              remove XML/HTML tags from input before check
  --rulefile FILE          additional grammar/rule file
  --falsefriends FILE      external false-friends XML
  --ruleValues LIST        soft RULE_ID:value pairs (e.g. TOO_LONG_SENTENCE:10)
  --autoDetect, -adl       detect language from text
  --list                   list languages
  --list-rules             list registered rule IDs for -l language
  --data-dir DIR           soft data root (grammar + false-friends soft files)
  --fail-on LEVEL          lint/sarif fail threshold: error|warning|note (default error)
  --doctor                 environment / self-check diagnostics
  --version                print version
  -h, --help               this help
`

// VersionString is printed by --version (override in main package).
var VersionString = "languagetool-go (dev)"

// RunCLI parses args and executes the corresponding action.
// check/tag hooks are pluggable so the binary can wire real LT later.
type RunHooks struct {
	// Check runs a plain-text check; required unless only meta flags are used.
	Check func(w io.Writer, text string, opts *CommandLineOptions) (int, error)
	// Tag runs tagger-only mode.
	Tag func(w io.Writer, text string, opts *CommandLineOptions) error
	// ListLanguages prints known languages.
	ListLanguages func(w io.Writer) error
	// ReadFile loads file contents; nil uses os.ReadFile. Empty filename → stdin via ReadStdin.
	ReadFile  func(path string) (string, error)
	ReadStdin func() (string, error)
}

// Run is the CLI entry used by main packages.
func Run(args []string, hooks RunHooks) int {
	return RunWithIO(args, hooks, os.Stdout, os.Stderr)
}

// NormalizeProductArgs maps soft product subcommands onto LT-style flags (SPEC §2).
// Examples: "lint -l en -" → "--lint -l en -"; "languages" → "--list".
func NormalizeProductArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	switch args[0] {
	case "lint":
		return append([]string{"--lint"}, args[1:]...)
	case "languages", "list":
		return []string{"--list"}
	case "rules":
		// keep -l / --language and other flags; force --list-rules
		return append([]string{"--list-rules"}, args[1:]...)
	case "doctor":
		return []string{"--doctor"}
	case "golden":
		return append([]string{"--golden"}, args[1:]...)
	case "compare":
		// compare GOLDEN [opts...] FILE  →  --compare GOLDEN [opts...] FILE
		return append([]string{"--compare"}, args[1:]...)
	case "version":
		return []string{"--version"}
	case "help":
		return []string{"--help"}
	default:
		return args
	}
}

// RunWithIO is Run with explicit streams (tests).
func RunWithIO(args []string, hooks RunHooks, stdout, stderr io.Writer) int {
	args = NormalizeProductArgs(args)
	p := &CommandLineParser{}
	opts, err := p.ParseOptions(args)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		_, _ = fmt.Fprint(stderr, UsageText)
		return 1
	}
	if opts.PrintUsage {
		_, _ = fmt.Fprint(stdout, UsageText)
		return 0
	}
	if opts.PrintVersion {
		_, _ = fmt.Fprintln(stdout, VersionString)
		return 0
	}
	if opts.PrintLanguages {
		if hooks.ListLanguages != nil {
			if err := hooks.ListLanguages(stdout); err != nil {
				_, _ = fmt.Fprintln(stderr, err.Error())
				return 1
			}
			return 0
		}
		_, _ = fmt.Fprintln(stdout, "en")
		return 0
	}
	if opts.PrintRules {
		lang := opts.Language
		if lang == "" {
			lang = "en"
		}
		if err := CoreListRules(stdout, lang); err != nil {
			_, _ = fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	}
	if opts.PrintDoctor {
		if err := CoreDoctor(stdout, opts); err != nil {
			_, _ = fmt.Fprintln(stderr, err.Error())
			return 1
		}
		return 0
	}

	// SPEC §2.2: product lint defaults language to auto when unset.
	if opts.OutputFormat == OutputLint && opts.Language == "" && !opts.AutoDetect {
		opts.SetAutoDetect(true)
	}

	files := opts.GetFilenames()
	if len(files) == 0 {
		files = []string{""} // stdin
	}

	// Multi-file: process each path; aggregate exit severity.
	totalN := 0
	for _, fn := range files {
		opts.Filename = fn
		text, err := loadInput(fn, hooks)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err.Error())
			return 1
		}
		if opts.TaggerOnly {
			if hooks.Tag == nil {
				_, _ = fmt.Fprintln(stderr, "tagger hook not configured")
				return 1
			}
			if err := hooks.Tag(stdout, text, opts); err != nil {
				_, _ = fmt.Fprintln(stderr, err.Error())
				return 1
			}
			continue
		}
		if hooks.Check == nil {
			_, _ = fmt.Fprintln(stderr, "check hook not configured")
			return 1
		}
		// Soft golden / compare product modes use dedicated hooks (single-file).
		if opts.GoldenMode {
			n, err := CoreGoldenHook(stdout, text, opts)
			if err != nil {
				_, _ = fmt.Fprintln(stderr, err.Error())
				return 1
			}
			totalN += n
			continue
		}
		if opts.CompareMode {
			n, err := CoreCompareHook(stdout, text, opts)
			if err != nil {
				_, _ = fmt.Fprintln(stderr, err.Error())
				return 1
			}
			totalN += n
			continue
		}

		n, err := hooks.Check(stdout, text, opts)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err.Error())
			return 1
		}
		totalN += n
	}
	if opts.TaggerOnly {
		return 0
	}
	if totalN > 0 {
		// SPEC §2.2: --lint / --sarif use exit 1 for error-severity findings.
		if opts != nil && (opts.OutputFormat == OutputLint || opts.OutputFormat == OutputSARIF || opts.GoldenMode || opts.CompareMode) {
			return 1
		}
		return 2 // matches found — LT-style CLI convention
	}
	return 0
}

func loadInput(filename string, hooks RunHooks) (string, error) {
	if filename == "" || filename == "-" {
		if hooks.ReadStdin != nil {
			return hooks.ReadStdin()
		}
		b, err := io.ReadAll(os.Stdin)
		return string(b), err
	}
	if hooks.ReadFile != nil {
		return hooks.ReadFile(filename)
	}
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	// strip UTF-8 BOM
	s := string(b)
	return strings.TrimPrefix(s, "\ufeff"), nil
}
