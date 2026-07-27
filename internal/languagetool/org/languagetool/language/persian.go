package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/xx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PersianLang ports org.languagetool.language.Persian metadata + default factories.
// Java extends Language (not LanguageWithModel); no createDefaultTagger /
// createDefaultSpellingRule / createDefaultDisambiguator / createDefaultSynthesizer
// overrides — do not invent PersianTagger, HUNSPELL, or MORFOLOGIK for FA.
// getMaintainedState not overridden → Language default LookingForNewMaintainer.
type PersianLang struct {
	ShortCode, Name string
	Countries       []string
}

// Persian is the default Persian language surface (Java new Persian()).
var Persian = PersianLang{
	ShortCode: "fa",
	Name:      "Persian",
	// Java getCountries → {"IR", "AF"}
	Countries: []string{"IR", "AF"},
}

// NewPersian ports new Persian().
func NewPersian() PersianLang {
	return Persian
}

// PersianShortCode ports Persian.getShortCode ("fa").
func PersianShortCode() string { return "fa" }

func (v PersianLang) GetName() string { return v.Name }

// GetShortCode ports Persian.getShortCode ("fa").
func (v PersianLang) GetShortCode() string { return "fa" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Java lists two countries (IR, AF) → len != 1 → bare "fa".
// Locale with country uses first country IR (getLocaleWithCountryAndVariant) — not this string.
func (v PersianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("fa", v.Countries, "")
}

func (v PersianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → fa/common_words.txt.
func (v PersianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID: Java Persian has no createDefaultSpellingRule and no
// speller in getRelevantRules — empty (do not invent HUNSPELL/MORFOLOGIK for FA).
func (v PersianLang) GetDefaultSpellingRuleID() string {
	return ""
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Persian does not override getMaintainedState.
func (v PersianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Persian.getMaintainers.
func (v PersianLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Reza1615"),
		NewContributor("Alireza Eskandarpour Shoferi"),
		NewContributor("Ebrahim Byagowi"),
	}
}

// CreateDefaultTagger ports Language.createDefaultTagger default → DemoTagger.
// Java Persian does not override createDefaultTagger — do not invent PersianTagger.
func (v PersianLang) CreateDefaultTagger() *xx.DemoTagger {
	return xx.NewDemoTagger()
}

// CreateDefaultSentenceTokenizer ports Persian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "fa".
func (v PersianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Persian.createDefaultWordTokenizer
// → new PersianWordTokenizer().
func (v PersianLang) CreateDefaultWordTokenizer() *tokenizers.PersianWordTokenizer {
	return tokenizers.NewPersianWordTokenizer()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "fa"
// (Java Persian does not add RULE_FILES extras).
func (v PersianLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v PersianLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
