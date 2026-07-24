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
	// PrintDoctor soft-runs environment/self-check diagnostics.
	PrintDoctor                   bool
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
	// Filenames soft multi-file lint (SPEC paths). Filename is Filenames[0] when set.
	Filenames                     []string
	DisabledRules                 []string
	EnabledRules                  []string
	EnabledCategories             []string
	DisabledCategories            []string
	UseEnabledOnlyFlag            bool
	RuleValues                    []string
	RuleFile                      string
	FalseFriendsFile              string
	// RemoteRulesFile ports remoteRulesFile (since 4.9).
	RemoteRulesFile string
	// BitextRuleFile ports bitextRuleFile (since 2.9).
	BitextRuleFile string
	// IgnoreWords user-dictionary surfaces (suppress spelling matches).
	IgnoreWords []string
	// IgnoreSpellingFile path to ignore-spelling word list (one form per line).
	IgnoreSpellingFile string
	// DisambiguationFile override for official disambiguation.xml.
	DisambiguationFile string
	// DataDir root for official resources (grammar/style/false-friends when present).
	DataDir string
	// FailOn severity threshold for lint/sarif exit: error|warning|note.
	FailOn string
	// GoldenMode / CompareMode for golden dump / compare subcommands.
	GoldenMode        bool
	CompareMode       bool
	CompareGoldenPath string
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
func (o *CommandLineOptions) SetPrintDoctor(v bool)    { o.PrintDoctor = v }
func (o *CommandLineOptions) SetDataDir(p string)      { o.DataDir = p }
func (o *CommandLineOptions) SetFailOn(s string)       { o.FailOn = s }
func (o *CommandLineOptions) SetGoldenMode(v bool)     { o.GoldenMode = v }
func (o *CommandLineOptions) SetCompareMode(v bool)    { o.CompareMode = v }
func (o *CommandLineOptions) SetCompareGoldenPath(p string) { o.CompareGoldenPath = p }
func (o *CommandLineOptions) GetDataDir() string {
	if o == nil {
		return ""
	}
	return o.DataDir
}
func (o *CommandLineOptions) GetFailOn() string {
	if o == nil || o.FailOn == "" {
		return "error"
	}
	return o.FailOn
}
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
func (o *CommandLineOptions) SetFilename(f string) {
	o.Filename = f
	if f == "" {
		return
	}
	// keep Filenames in sync for multi-file
	if len(o.Filenames) == 0 {
		o.Filenames = []string{f}
	} else if o.Filenames[0] != f && f != "" {
		// first file already set; AddFilename handles extras
	}
}
func (o *CommandLineOptions) AddFilename(f string) {
	if f == "" {
		return
	}
	o.Filenames = append(o.Filenames, f)
	if o.Filename == "" {
		o.Filename = f
	}
}
func (o *CommandLineOptions) GetFilenames() []string {
	if o == nil {
		return nil
	}
	if len(o.Filenames) > 0 {
		return append([]string(nil), o.Filenames...)
	}
	if o.Filename != "" {
		return []string{o.Filename}
	}
	return nil
}
func (o *CommandLineOptions) SetRuleFile(f string)           { o.RuleFile = f }
func (o *CommandLineOptions) SetFalseFriendsFile(f string)   { o.FalseFriendsFile = f }
func (o *CommandLineOptions) SetIgnoreWords(words []string) {
	if o == nil {
		return
	}
	o.IgnoreWords = append([]string(nil), words...)
}
func (o *CommandLineOptions) GetIgnoreWords() []string {
	if o == nil {
		return nil
	}
	return o.IgnoreWords
}
func (o *CommandLineOptions) SetIgnoreSpellingFile(p string) {
	if o != nil {
		o.IgnoreSpellingFile = p
	}
}
func (o *CommandLineOptions) GetIgnoreSpellingFile() string {
	if o == nil {
		return ""
	}
	return o.IgnoreSpellingFile
}
func (o *CommandLineOptions) SetDisambiguationFile(p string) {
	if o != nil {
		o.DisambiguationFile = p
	}
}
func (o *CommandLineOptions) GetDisambiguationFile() string {
	if o == nil {
		return ""
	}
	return o.DisambiguationFile
}
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

// IsJsonFormat ports isJsonFormat.
func (o *CommandLineOptions) IsJsonFormat() bool {
	return o != nil && o.OutputFormat == OutputJSON
}

// SetJsonFormat ports setJsonFormat.
func (o *CommandLineOptions) SetJsonFormat() {
	if o != nil {
		o.OutputFormat = OutputJSON
	}
}

func (o *CommandLineOptions) IsPrintUsage() bool     { return o != nil && o.PrintUsage }
func (o *CommandLineOptions) IsPrintVersion() bool   { return o != nil && o.PrintVersion }
func (o *CommandLineOptions) IsPrintLanguages() bool { return o != nil && o.PrintLanguages }
func (o *CommandLineOptions) IsVerbose() bool        { return o != nil && o.Verbose }
func (o *CommandLineOptions) IsRecursive() bool      { return o != nil && o.Recursive }
func (o *CommandLineOptions) IsTaggerOnly() bool     { return o != nil && o.TaggerOnly }
func (o *CommandLineOptions) IsSingleLineBreakMarksParagraph() bool {
	return o != nil && o.SingleLineBreakMarksParagraph
}
func (o *CommandLineOptions) IsProfile() bool        { return o != nil && o.Profile }
func (o *CommandLineOptions) IsBitext() bool         { return o != nil && o.Bitext }
func (o *CommandLineOptions) IsXmlFiltering() bool   { return o != nil && o.XMLFiltering }
func (o *CommandLineOptions) IsLineByLine() bool     { return o != nil && o.LineByLine }
func (o *CommandLineOptions) IsEnableTempOff() bool  { return o != nil && o.EnableTempOff }
func (o *CommandLineOptions) IsCleanOverlapping() bool {
	return o != nil && o.CleanOverlapping
}

func (o *CommandLineOptions) GetOutputFormat() OutputFormat {
	if o == nil {
		return OutputPlaintext
	}
	return o.OutputFormat
}
func (o *CommandLineOptions) GetLanguage() string {
	if o == nil {
		return ""
	}
	return o.Language
}
func (o *CommandLineOptions) GetMotherTongue() string {
	if o == nil {
		return ""
	}
	return o.MotherTongue
}
func (o *CommandLineOptions) GetLanguageModelPath() string {
	if o == nil {
		return ""
	}
	return o.LanguageModelPath
}
func (o *CommandLineOptions) GetFasttextModelPath() string {
	if o == nil {
		return ""
	}
	return o.FasttextModelPath
}
func (o *CommandLineOptions) GetFasttextBinaryPath() string {
	if o == nil {
		return ""
	}
	return o.FasttextBinaryPath
}
func (o *CommandLineOptions) SetFasttextModelPath(p string) {
	if o != nil {
		o.FasttextModelPath = p
	}
}
func (o *CommandLineOptions) SetFasttextBinaryPath(p string) {
	if o != nil {
		o.FasttextBinaryPath = p
	}
}
func (o *CommandLineOptions) GetEncoding() string {
	if o == nil {
		return ""
	}
	return o.Encoding
}
func (o *CommandLineOptions) GetFilename() string {
	if o == nil {
		return ""
	}
	return o.Filename
}
func (o *CommandLineOptions) GetLevel() string {
	if o == nil || o.Level == "" {
		return "DEFAULT"
	}
	return o.Level
}
func (o *CommandLineOptions) GetFalseFriendFile() string {
	if o == nil {
		return ""
	}
	return o.FalseFriendsFile
}
func (o *CommandLineOptions) GetRemoteRulesFile() string {
	if o == nil {
		return ""
	}
	return o.RemoteRulesFile
}
func (o *CommandLineOptions) SetRemoteRulesFile(f string) {
	if o != nil {
		o.RemoteRulesFile = f
	}
}
func (o *CommandLineOptions) GetBitextRuleFile() string {
	if o == nil {
		return ""
	}
	return o.BitextRuleFile
}
func (o *CommandLineOptions) SetBitextRuleFile(f string) {
	if o != nil {
		o.BitextRuleFile = f
	}
}
func (o *CommandLineOptions) GetEnabledCategories() []string {
	if o == nil {
		return nil
	}
	return append([]string(nil), o.EnabledCategories...)
}
func (o *CommandLineOptions) GetDisabledCategories() []string {
	if o == nil {
		return nil
	}
	return append([]string(nil), o.DisabledCategories...)
}
