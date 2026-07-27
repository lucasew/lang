package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SlovenianLang ports org.languagetool.language.Slovenian metadata + default factories.
// Java @Deprecated since 3.6 (unmaintained); getMaintainedState not overridden →
// Language default LookingForNewMaintainer.
type SlovenianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Slovenian is the default Slovenian language surface (Java new Slovenian()).
var Slovenian = SlovenianLang{
	ShortCode:     "sl",
	Name:          "Slovenian",
	SpellerRuleID: "MORFOLOGIK_RULE_SL_SI",
	Countries:     []string{"SI"},
}

// NewSlovenian ports new Slovenian().
func NewSlovenian() SlovenianLang {
	return Slovenian
}

// SlovenianShortCode ports Slovenian.getShortCode ("sl").
func SlovenianShortCode() string { return "sl" }

func (v SlovenianLang) GetName() string { return v.Name }

// GetShortCode ports Slovenian.getShortCode ("sl").
func (v SlovenianLang) GetShortCode() string { return "sl" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v SlovenianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("sl", v.Countries, "")
}

func (v SlovenianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → sl/common_words.txt.
func (v SlovenianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikSlovenianSpellerRule.getId
// → MORFOLOGIK_RULE_SL_SI.
func (v SlovenianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_SL_SI"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Slovenian is @Deprecated (unmaintained since 3.6) but does not override getMaintainedState.
func (v SlovenianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Slovenian.getMaintainers → Martin Srebotnjak.
func (v SlovenianLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Martin Srebotnjak")}
}

// CreateDefaultSentenceTokenizer ports Slovenian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "sl".
func (v SlovenianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Language.createDefaultWordTokenizer default
// (Java Slovenian does not override → WORD_TOKENIZER / new WordTokenizer()).
func (v SlovenianLang) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "sl"
// (Java Slovenian does not add RULE_FILES extras).
func (v SlovenianLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v SlovenianLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
