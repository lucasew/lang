package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	elsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/el"
	eldis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/el"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	eltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/el"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	eltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/el"
)

// GreekLang ports org.languagetool.language.Greek metadata + default factories.
type GreekLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Greek is the default Greek language surface (Java new Greek()).
var Greek = GreekLang{
	ShortCode:     "el",
	Name:          "Greek",
	SpellerRuleID: "MORFOLOGIK_RULE_EL_GR",
	Countries:     []string{"GR"},
}

// NewGreek ports new Greek().
func NewGreek() GreekLang {
	return Greek
}

// GreekShortCode ports Greek.getShortCode ("el").
func GreekShortCode() string { return "el" }

func (v GreekLang) GetName() string { return v.Name }

// GetShortCode ports Greek.getShortCode ("el").
func (v GreekLang) GetShortCode() string { return "el" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v GreekLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("el", v.Countries, "")
}

func (v GreekLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → el/common_words.txt.
func (v GreekLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikGreekSpellerRule.getId
// → MORFOLOGIK_RULE_EL_GR.
func (v GreekLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_EL_GR"
}

// GetMaintainedState ports Greek.getMaintainedState → ActivelyMaintained.
func (v GreekLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Greek.getMaintainers → Panagiotis Minos.
func (v GreekLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Panagiotis Minos")}
}

// CreateDefaultTagger ports Greek.createDefaultTagger → new GreekTagger().
// Loads official /el/greek.dict when discoverable (Java BaseTagger path).
// Returns process DefaultGreekTagger (MapWordTagger until dict loads).
// Note: Java GreekTagger also uses GreekAnalyzer.additionalTags — leaf twin gap.
func (v GreekLang) CreateDefaultTagger() *eltag.GreekTagger {
	eltag.EnsureDefaultGreekTagger()
	return eltag.DefaultGreekTagger
}

// CreateDefaultSentenceTokenizer ports Greek.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "el".
// (Java Greek does not use GreekSRXSentenceTokenizer.)
func (v GreekLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Greek.createDefaultWordTokenizer
// → new GreekWordTokenizer().
func (v GreekLang) CreateDefaultWordTokenizer() *eltok.GreekWordTokenizer {
	return eltok.NewGreekWordTokenizer()
}

// CreateDefaultDisambiguator ports Greek.createDefaultDisambiguator
// → new XmlRuleDisambiguator(this) (useGlobalDisambiguation default false).
// Nil if el/disambiguation.xml cannot be loaded (fail-closed).
func (v GreekLang) CreateDefaultDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return eldis.GreekXmlRuleDisambiguator()
}

// CreateDefaultSynthesizer ports Greek.createDefaultSynthesizer
// → GreekSynthesizer.INSTANCE (lazy official load).
// Nil if greek_synth.dict cannot be opened (fail-closed).
func (v GreekLang) CreateDefaultSynthesizer() *elsynth.GreekSynthesizer {
	return elsynth.Instance()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "el"
// (Java Greek does not add RULE_FILES extras).
func (v GreekLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v GreekLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
