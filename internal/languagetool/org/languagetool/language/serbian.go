package language

// Serbian ports org.languagetool.language.Serbian (metadata + rule file list).
type Serbian struct {
	ShortCode string
	Name      string
	Countries []string
	// ExtraRuleFiles are language-specific grammar XML files (without base grammar.xml).
	ExtraRuleFiles []string
}

// DefaultSerbian is the base Serbian language (Ekavian default variant surface).
var DefaultSerbian = Serbian{
	ShortCode: "sr",
	Name:      "Serbian",
	Countries: []string{"RS"},
	ExtraRuleFiles: []string{
		"grammar-barbarism.xml",
		"grammar-logical.xml",
		"grammar-punctuation.xml",
		"grammar-spelling.xml",
		"grammar-style.xml",
	},
}

func NewSerbian() Serbian { return DefaultSerbian }

func (s Serbian) GetShortCode() string { return s.ShortCode }
func (s Serbian) GetName() string      { return s.Name }
func (s Serbian) GetCountries() []string {
	return append([]string(nil), s.Countries...)
}

// GetRuleFileNames ports Serbian.getRuleFileNames — grammar.xml first, then extras under /org/languagetool/rules/sr/.
func (s Serbian) GetRuleFileNames() []string {
	const dirBase = "/org/languagetool/rules/sr/"
	out := []string{dirBase + "grammar.xml"}
	for _, f := range s.ExtraRuleFiles {
		out = append(out, dirBase+f)
	}
	return out
}
