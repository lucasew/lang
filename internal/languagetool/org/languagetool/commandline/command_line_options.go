package commandline

// OutputFormat ports CommandLineOptions.OutputFormat.
type OutputFormat string

const (
	OutputPlaintext OutputFormat = "PLAINTEXT"
	OutputJSON      OutputFormat = "JSON"
	OutputXML       OutputFormat = "XML"
	// OutputSARIF is a soft SARIF 2.1 report (SPEC §2.2).
	OutputSARIF OutputFormat = "SARIF"
	// OutputLint is tab-separated location/severity/type/rule/message/suggestion (SPEC §2.2 text).
	OutputLint OutputFormat = "LINT"
)

// CommandLineOptions ports org.languagetool.commandline.CommandLineOptions.
type CommandLineOptions struct {
	PrintUsage                    bool
	PrintVersion                  bool
	PrintLanguages                bool
	// PrintRules soft-lists registered rule IDs for the selected language.
	PrintRules                    bool
	Verbose                       bool
	Recursive                     bool
	TaggerOnly                    bool
	SingleLineBreakMarksParagraph bool
	OutputFormat                  OutputFormat
	ListUnknown                   bool
	ApplySuggestions              bool
	Profile                       bool
	Bitext                        bool
	AutoDetect                    bool
	XMLFiltering                  bool
	LineByLine                    bool
	EnableTempOff                 bool
	CleanOverlapping              bool
	Level                         string // DEFAULT, PICKY, ...
	Language                      string
	MotherTongue                  string
	LanguageModelPath             string
	FasttextModelPath             string
	FasttextBinaryPath            string
	Encoding                      string
	Filename                      string
	DisabledRules                 []string
	EnabledRules                  []string
	EnabledCategories             []string
	DisabledCategories            []string
	UseEnabledOnlyFlag            bool
	RuleValues                    []string
	RuleFile                      string
	FalseFriendsFile              string
}

func NewCommandLineOptions() *CommandLineOptions {
	return &CommandLineOptions{
		OutputFormat: OutputPlaintext,
		Level:        "DEFAULT",
	}
}

func (o *CommandLineOptions) SetPrintUsage(v bool)     { o.PrintUsage = v }
func (o *CommandLineOptions) SetPrintVersion(v bool)   { o.PrintVersion = v }
func (o *CommandLineOptions) SetPrintLanguages(v bool) { o.PrintLanguages = v }
func (o *CommandLineOptions) SetPrintRules(v bool)     { o.PrintRules = v }
func (o *CommandLineOptions) SetVerbose(v bool)        { o.Verbose = v }
func (o *CommandLineOptions) SetRecursive(v bool)      { o.Recursive = v }
func (o *CommandLineOptions) SetTaggerOnly(v bool)     { o.TaggerOnly = v }
func (o *CommandLineOptions) SetSingleLineBreakMarksParagraph(v bool) {
	o.SingleLineBreakMarksParagraph = v
}
func (o *CommandLineOptions) SetOutputFormat(f OutputFormat) { o.OutputFormat = f }
func (o *CommandLineOptions) SetListUnknown(v bool)          { o.ListUnknown = v }
func (o *CommandLineOptions) SetApplySuggestions(v bool)     { o.ApplySuggestions = v }
func (o *CommandLineOptions) SetProfile(v bool)              { o.Profile = v }
func (o *CommandLineOptions) SetBitext(v bool)               { o.Bitext = v }
func (o *CommandLineOptions) SetAutoDetect(v bool)           { o.AutoDetect = v }
func (o *CommandLineOptions) SetXMLFiltering(v bool)         { o.XMLFiltering = v }
func (o *CommandLineOptions) SetLineByLine(v bool)           { o.LineByLine = v }
func (o *CommandLineOptions) SetEnableTempOff(v bool)        { o.EnableTempOff = v }
func (o *CommandLineOptions) SetCleanOverlapping(v bool)     { o.CleanOverlapping = v }
func (o *CommandLineOptions) SetLevel(level string)          { o.Level = level }
func (o *CommandLineOptions) SetLanguage(code string)        { o.Language = code }
func (o *CommandLineOptions) SetMotherTongue(code string)    { o.MotherTongue = code }
func (o *CommandLineOptions) SetLanguageModelPath(p string)  { o.LanguageModelPath = p }
func (o *CommandLineOptions) SetEncoding(e string)           { o.Encoding = e }
func (o *CommandLineOptions) SetFilename(f string)           { o.Filename = f }
func (o *CommandLineOptions) SetRuleFile(f string)           { o.RuleFile = f }
func (o *CommandLineOptions) SetFalseFriendsFile(f string)   { o.FalseFriendsFile = f }
func (o *CommandLineOptions) GetRuleFile() string {
	if o == nil {
		return ""
	}
	return o.RuleFile
}
func (o *CommandLineOptions) IsAutoDetect() bool {
	return o != nil && o.AutoDetect
}
func (o *CommandLineOptions) SetDisabledRules(ids []string) {
	o.DisabledRules = append([]string(nil), ids...)
}
func (o *CommandLineOptions) SetEnabledRules(ids []string) {
	o.EnabledRules = append([]string(nil), ids...)
}
func (o *CommandLineOptions) SetEnabledCategories(ids []string) {
	o.EnabledCategories = append([]string(nil), ids...)
}
func (o *CommandLineOptions) SetDisabledCategories(ids []string) {
	o.DisabledCategories = append([]string(nil), ids...)
}
func (o *CommandLineOptions) SetUseEnabledOnly() { o.UseEnabledOnlyFlag = true }
func (o *CommandLineOptions) IsUseEnabledOnly() bool {
	return o != nil && o.UseEnabledOnlyFlag
}
func (o *CommandLineOptions) IsListUnknown() bool {
	return o != nil && o.ListUnknown
}
func (o *CommandLineOptions) IsApplySuggestions() bool {
	return o != nil && o.ApplySuggestions
}
func (o *CommandLineOptions) GetDisabledRules() []string {
	if o == nil {
		return nil
	}
	return append([]string(nil), o.DisabledRules...)
}
func (o *CommandLineOptions) GetEnabledRules() []string {
	if o == nil {
		return nil
	}
	return append([]string(nil), o.EnabledRules...)
}
func (o *CommandLineOptions) GetRuleValues() []string {
	if o == nil {
		return nil
	}
	return append([]string(nil), o.RuleValues...)
}
