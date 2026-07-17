package commandline

import (
	"fmt"
	"strings"
)

// CommandLineParser ports org.languagetool.commandline.CommandLineParser.
type CommandLineParser struct{}

// ParseOptions parses CLI args into options.
// Unlike Java, empty args yield usage-only options instead of requiring 1–14 args.
func (p *CommandLineParser) ParseOptions(args []string) (*CommandLineOptions, error) {
	opts := NewCommandLineOptions()
	if len(args) == 0 {
		opts.SetPrintUsage(true)
		return opts, nil
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--version":
			opts.SetPrintVersion(true)
		case "--list":
			opts.SetPrintLanguages(true)
		case "-h", "-help", "--help", "--?":
			opts.SetPrintUsage(true)
		case "-adl", "--autoDetect":
			opts.SetAutoDetect(true)
		case "-v", "--verbose":
			opts.SetVerbose(true)
		case "--line-by-line":
			opts.SetLineByLine(true)
		case "--enable-temp-off":
			opts.SetEnableTempOff(true)
		case "--clean-overlapping":
			opts.SetCleanOverlapping(true)
		case "--level":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetLevel(args[i])
		case "-a", "--apply":
			if opts.TaggerOnly {
				return nil, fmt.Errorf("You cannot apply suggestions when tagging only")
			}
			opts.SetApplySuggestions(true)
		case "-t", "--taggeronly":
			if opts.IsListUnknown() {
				return nil, fmt.Errorf("You cannot list unknown words when tagging only")
			}
			if opts.IsApplySuggestions() {
				return nil, fmt.Errorf("You cannot apply suggestions when tagging only")
			}
			opts.SetTaggerOnly(true)
		case "-r", "--recursive":
			opts.SetRecursive(true)
		case "-b2", "--bitext":
			opts.SetBitext(true)
		case "-eo", "--enabledonly":
			if len(opts.GetDisabledRules()) > 0 {
				return nil, fmt.Errorf("You cannot specify both disabled rules and enabledonly")
			}
			opts.SetUseEnabledOnly()
		case "-d", "--disable":
			if opts.IsUseEnabledOnly() {
				return nil, fmt.Errorf("You cannot specify both disabled rules and enabledonly")
			}
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetDisabledRules(splitCSV(args[i]))
		case "-e", "--enable":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetEnabledRules(splitCSV(args[i]))
		case "--enablecategories":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetEnabledCategories(splitCSV(args[i]))
		case "--disablecategories":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetDisabledCategories(splitCSV(args[i]))
		case "-l", "--language":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetLanguage(args[i])
		case "-m", "--mothertongue":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetMotherTongue(args[i])
		case "-c", "--encoding":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetEncoding(args[i])
		case "--xmlfilter":
			opts.SetXMLFiltering(true)
		case "--rulefile":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetRuleFile(args[i])
		case "--falsefriends":
			if err := needArg(a, i, args); err != nil {
				return nil, err
			}
			i++
			opts.SetFalseFriendsFile(args[i])
		case "--json":
			opts.SetOutputFormat(OutputJSON)
		case "--xml":
			opts.SetOutputFormat(OutputXML)
		case "-u", "--list-unknown":
			if opts.TaggerOnly {
				return nil, fmt.Errorf("You cannot list unknown words when tagging only")
			}
			opts.SetListUnknown(true)
		case "-b", "--bitext-false-friend":
			// legacy alias often used with bitext; treat as bitext flag
			opts.SetBitext(true)
		default:
			// "-" means stdin (not an unknown flag)
			if a == "-" {
				opts.SetFilename("-")
				continue
			}
			if strings.HasPrefix(a, "-") {
				return nil, UnknownParameterException{Param: a}
			}
			// positional filename
			if opts.Filename == "" {
				opts.SetFilename(a)
			} else {
				return nil, UnknownParameterException{Param: a}
			}
		}
	}
	return opts, nil
}

func needArg(flag string, i int, args []string) error {
	if i+1 >= len(args) {
		return fmt.Errorf("missing argument for %s", flag)
	}
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
