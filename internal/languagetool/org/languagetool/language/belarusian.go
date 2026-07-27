package language

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	betok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/be"
)

// BelarusianLang ports org.languagetool.language.Belarusian metadata + default factories.
// Java extends LanguageWithModel, @Deprecated since 3.6 (unmaintained);
// getMaintainedState not overridden → Language default LookingForNewMaintainer.
// No createDefaultDisambiguator / createDefaultSynthesizer (Java has none — do not invent).
type BelarusianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Belarusian is the default Belarusian language surface (Java new Belarusian()).
var Belarusian = BelarusianLang{
	ShortCode:     "be",
	Name:          "Belarusian",
	SpellerRuleID: "MORFOLOGIK_RULE_BE_BY",
	Countries:     []string{"BY"},
}

// NewBelarusian ports new Belarusian().
func NewBelarusian() BelarusianLang {
	return Belarusian
}

// BelarusianShortCode ports Belarusian.getShortCode ("be").
func BelarusianShortCode() string { return "be" }

func (v BelarusianLang) GetName() string { return v.Name }

// GetShortCode ports Belarusian.getShortCode ("be").
func (v BelarusianLang) GetShortCode() string { return "be" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v BelarusianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("be", v.Countries, "")
}

func (v BelarusianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → be/common_words.txt.
func (v BelarusianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule /
// MorfologikBelarusianSpellerRule.getId → MORFOLOGIK_RULE_BE_BY.
func (v BelarusianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_BE_BY"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Belarusian is @Deprecated (unmaintained since 3.6) but does not override getMaintainedState.
func (v BelarusianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Belarusian.getMaintainers → Alex Buloichik.
func (v BelarusianLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Alex Buloichik")}
}

// GetIgnoredCharactersRegex ports Belarusian.getIgnoredCharactersRegex
// → soft hyphen + combining acute/grave [\u00AD\u0301\u0300].
func (v BelarusianLang) GetIgnoredCharactersRegex() *regexp.Regexp {
	return languagetool.BelarusianIgnoredCharactersRegex
}

// CreateDefaultTagger ports Belarusian.createDefaultTagger → new DemoTagger().
func (v BelarusianLang) CreateDefaultTagger() *xx.DemoTagger {
	return xx.NewDemoTagger()
}

// CreateDefaultSentenceTokenizer ports Belarusian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "be".
func (v BelarusianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Belarusian.createDefaultWordTokenizer
// → new BelarusianWordTokenizer().
func (v BelarusianLang) CreateDefaultWordTokenizer() *betok.BelarusianWordTokenizer {
	return betok.NewBelarusianWordTokenizer()
}

// GetOpeningDoubleQuote ports Belarusian.getOpeningDoubleQuote → «.
func (v BelarusianLang) GetOpeningDoubleQuote() string { return "«" }

// GetClosingDoubleQuote ports Belarusian.getClosingDoubleQuote → ».
func (v BelarusianLang) GetClosingDoubleQuote() string { return "»" }

// GetOpeningSingleQuote ports Belarusian.getOpeningSingleQuote → ‘.
func (v BelarusianLang) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Belarusian.getClosingSingleQuote → ’.
func (v BelarusianLang) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Belarusian.isAdvancedTypographyEnabled → true.
func (v BelarusianLang) IsAdvancedTypographyEnabled() bool { return true }

// ToAdvancedTypography ports Language.toAdvancedTypography with Belarusian quotes
// and isAdvancedTypographyEnabled=true.
func (v BelarusianLang) ToAdvancedTypography(input string) string {
	return BelarusianAdvancedTypography(input)
}

// GetPriorityForId ports Belarusian.getPriorityForId.
func (v BelarusianLang) GetPriorityForId(id string) int {
	return BelarusianPriorityForId(id)
}

// GetRuleFileNames ports Language.getRuleFileNames default for "be"
// (Java Belarusian does not add RULE_FILES extras).
func (v BelarusianLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v BelarusianLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
