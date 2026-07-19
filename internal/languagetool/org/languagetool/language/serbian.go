package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Serbian ports org.languagetool.language.Serbian (metadata + rule file list).
// Jekavian is a dialect flag (JekavianSerbian and country variants).
type Serbian struct {
	ShortCode string
	Name      string
	Countries []string
	// Jekavian ports JekavianSerbian dialect (vs Ekavian base / SerbianSerbian).
	Jekavian bool
	// ExtraRuleFiles are language-specific grammar XML files (without base grammar.xml).
	ExtraRuleFiles []string
}

// serbianExtraRuleFiles ports Serbian.RULE_FILES (grammar.xml added by getRuleFileNames).
var serbianExtraRuleFiles = []string{
	"grammar-barbarism.xml",
	"grammar-logical.xml",
	"grammar-punctuation.xml",
	"grammar-spelling.xml",
	"grammar-style.xml",
}

// DefaultSerbian is the base Serbian language (Java Serbian: empty countries, Ekavian rules).
var DefaultSerbian = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian",
	Countries:      nil, // Java getCountries → new String[]{}
	Jekavian:       false,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

// SerbianSerbia ports SerbianSerbian (default country variant of Serbian).
var SerbianSerbia = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian (Serbia)",
	Countries:      []string{"RS"},
	Jekavian:       false,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

// JekavianSerbian ports org.languagetool.language.JekavianSerbian (dialect; empty countries).
var JekavianSerbian = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian",
	Countries:      nil,
	Jekavian:       true,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

// BosnianSerbian ports BosnianSerbian extends JekavianSerbian.
var BosnianSerbian = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian (Bosnia and Herzegovina)",
	Countries:      []string{"BA"},
	Jekavian:       true,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

// CroatianSerbian ports CroatianSerbian extends JekavianSerbian.
var CroatianSerbian = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian (Croatia)",
	Countries:      []string{"HR"},
	Jekavian:       true,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

// MontenegrinSerbian ports MontenegrinSerbian extends JekavianSerbian.
var MontenegrinSerbian = Serbian{
	ShortCode:      "sr",
	Name:           "Serbian (Montenegro)",
	Countries:      []string{"ME"},
	Jekavian:       true,
	ExtraRuleFiles: append([]string(nil), serbianExtraRuleFiles...),
}

func NewSerbian() Serbian             { return DefaultSerbian }
func NewSerbianSerbia() Serbian       { return SerbianSerbia }
func NewJekavianSerbian() Serbian     { return JekavianSerbian }
func NewBosnianSerbian() Serbian      { return BosnianSerbian }
func NewCroatianSerbian() Serbian     { return CroatianSerbian }
func NewMontenegrinSerbian() Serbian  { return MontenegrinSerbian }

// AllSerbianVariants lists base + country/dialect descriptors (metadata surface).
func AllSerbianVariants() []Serbian {
	return []Serbian{
		DefaultSerbian, SerbianSerbia, JekavianSerbian,
		BosnianSerbian, CroatianSerbian, MontenegrinSerbian,
	}
}

func (s Serbian) GetShortCode() string { return s.ShortCode }
func (s Serbian) GetName() string      { return s.Name }
func (s Serbian) GetCountries() []string {
	return append([]string(nil), s.Countries...)
}

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Base/Jekavian with empty countries → "sr"; SerbianSerbia → "sr-RS", etc.
func (s Serbian) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant(s.ShortCode, s.Countries, "")
}

// GetCommonWordsPath ports Language.getCommonWordsPath → sr/common_words.txt.
func (s Serbian) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(s.GetShortCode())
}

// GetMaintainedState ports Serbian.getMaintainedState → LookingForNewMaintainer.
func (s Serbian) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Serbian.getMaintainers.
func (s Serbian) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Золтан Чала (Csala, Zoltán)"),
	}
}

// GetRuleFileNames ports Serbian.getRuleFileNames — grammar.xml first, then extras under /org/languagetool/rules/sr/.
func (s Serbian) GetRuleFileNames() []string {
	const dirBase = "/org/languagetool/rules/sr/"
	out := []string{dirBase + "grammar.xml"}
	files := s.ExtraRuleFiles
	if len(files) == 0 {
		files = serbianExtraRuleFiles
	}
	for _, f := range files {
		out = append(out, dirBase+f)
	}
	return out
}

// GetDefaultSpellingRuleID ports dialect Morfologik*SpellerRule getId.
func (s Serbian) GetDefaultSpellingRuleID() string {
	if s.Jekavian {
		return "MORFOLOGIK_RULE_SR_JEKAVIAN"
	}
	return "MORFOLOGIK_RULE_SR_EKAVIAN"
}
