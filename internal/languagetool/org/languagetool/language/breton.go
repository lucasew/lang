package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	brtag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/br"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
)

// BretonLang ports org.languagetool.language.Breton metadata + default factories.
// Java overrides getMaintainedState → ActivelyMaintained.
// No createDefaultSynthesizer (Java has none — do not invent).
type BretonLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Breton is the default Breton language surface (Java new Breton()).
var Breton = BretonLang{
	ShortCode:     "br",
	Name:          "Breton",
	SpellerRuleID: "MORFOLOGIK_RULE_BR_FR",
	Countries:     []string{"FR"},
}

// NewBreton ports new Breton().
func NewBreton() BretonLang {
	return Breton
}

// BretonShortCode ports Breton.getShortCode ("br").
func BretonShortCode() string { return "br" }

func (v BretonLang) GetName() string { return v.Name }

// GetShortCode ports Breton.getShortCode ("br").
func (v BretonLang) GetShortCode() string { return "br" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v BretonLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("br", v.Countries, "")
}

func (v BretonLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → br/common_words.txt.
func (v BretonLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikBretonSpellerRule.getId
// → MORFOLOGIK_RULE_BR_FR.
func (v BretonLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_BR_FR"
}

// GetMaintainedState ports Breton.getMaintainedState → ActivelyMaintained.
func (v BretonLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Breton.getMaintainers → Contributors.DOMINIQUE_PELLE, Fulup Jakez.
func (v BretonLang) GetMaintainers() []Contributor {
	return []Contributor{
		DominiquePelle,
		NewContributor("Fulup Jakez"),
	}
}

// CreateDefaultTagger ports Breton.createDefaultTagger → new BretonTagger().
// Loads official /br/breton.dict when discoverable (Java BaseTagger path).
// Returns process DefaultBretonTagger (MapWordTagger until dict loads).
func (v BretonLang) CreateDefaultTagger() *brtag.BretonTagger {
	brtag.EnsureDefaultBretonTagger()
	return brtag.DefaultBretonTagger
}

// CreateDefaultSentenceTokenizer ports Breton.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "br".
func (v BretonLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Breton.createDefaultWordTokenizer
// → new BretonWordTokenizer().
func (v BretonLang) CreateDefaultWordTokenizer() *brtok.BretonWordTokenizer {
	return brtok.NewBretonWordTokenizer()
}

// CreateDefaultDisambiguator ports Breton.createDefaultDisambiguator
// → new XmlRuleDisambiguator(this) (useGlobalDisambiguation default false).
// Nil if br/disambiguation.xml cannot be loaded (fail-closed).
// Uses existing BR XML path (tagging/br) — no invent.
func (v BretonLang) CreateDefaultDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return brtag.BretonXmlRuleDisambiguator()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "br"
// (Java Breton does not add RULE_FILES extras).
func (v BretonLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v BretonLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
