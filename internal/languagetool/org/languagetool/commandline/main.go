package commandline

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// UsageText is the CLI help banner (ports Main usage surface).
const UsageText = `Usage: languagetool [OPTION]... [FILE]
  -l, --language CODE      language code (e.g. en-US)
  -m, --mothertongue CODE  mother tongue for false friends
  -d, --disable RULES      comma-separated disabled rule ids
  -e, --enable RULES       comma-separated enabled rule ids
  -t, --taggeronly         only tag the text
  -a, --apply              apply first suggestions to text, print result
  -v, --verbose            verbose output
  --json                   JSON output
  --xml                    XML output
  --xmlfilter              remove XML/HTML tags from input before check
  --rulefile FILE          additional grammar/rule file
  --falsefriends FILE      external false-friends XML
  --autoDetect, -adl       detect language from text
  --list                   list languages
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

// RunWithIO is Run with explicit streams (tests).
func RunWithIO(args []string, hooks RunHooks, stdout, stderr io.Writer) int {
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

	text, err := loadInput(opts.Filename, hooks)
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
		return 0
	}
	if hooks.Check == nil {
		_, _ = fmt.Fprintln(stderr, "check hook not configured")
		return 1
	}
	n, err := hooks.Check(stdout, text, opts)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		return 1
	}
	if n > 0 {
		return 2 // matches found — common CLI convention
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
