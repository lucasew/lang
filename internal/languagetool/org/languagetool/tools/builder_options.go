package tools

// CLI option names for Morfologik dictionary builders
// (org.languagetool.tools.BuilderOptions).
const (
	BuilderInfoOption   = "info"
	BuilderOutputOption = "o"
	BuilderInputOption  = "i"
	BuilderFreqOption   = "freq"

	BuilderFreqHelp = "optional .xml file with a frequency wordlist, " +
		"see https://dev.languagetool.org/developing-a-tagger-dictionary"
	BuilderInfoHelp = "*.info properties file, " +
		"see https://dev.languagetool.org/developing-a-tagger-dictionary"
	BuilderTabInputHelp = "tab-separated plain-text dictionary file " +
		"with format: wordform<tab>lemma<tab>postag"
)

// BuilderOptions holds parsed dictionary builder CLI options.
type BuilderOptions struct {
	InfoFile   string
	InputFile  string
	OutputFile string
	FreqFile   string
}

// ParseBuilderArgs parses a minimal flag set (-i, -o, -info, -freq).
func ParseBuilderArgs(args []string) (BuilderOptions, error) {
	var o BuilderOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-"+BuilderInputOption, "--input":
			if i+1 >= len(args) {
				return o, errf("missing value for input")
			}
			o.InputFile = args[i+1]
			i++
		case "-"+BuilderOutputOption, "--output":
			if i+1 >= len(args) {
				return o, errf("missing value for output")
			}
			o.OutputFile = args[i+1]
			i++
		case "-"+BuilderInfoOption, "--info":
			if i+1 >= len(args) {
				return o, errf("missing value for info")
			}
			o.InfoFile = args[i+1]
			i++
		case "-"+BuilderFreqOption, "--freq":
			if i+1 >= len(args) {
				return o, errf("missing value for freq")
			}
			o.FreqFile = args[i+1]
			i++
		}
	}
	return o, nil
}

type builderErr string

func (e builderErr) Error() string { return string(e) }
func errf(s string) error          { return builderErr(s) }
