package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	gasynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ga"
	gadis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ga"
	gatag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// IrishLang ports org.languagetool.language.Irish metadata + default factories.
// Java extends LanguageWithModel; only peer-ported LanguageWithModel surface is used
// (no invent ngram model). Overrides getMaintainedState → ActivelyMaintained.
// Pattern grammar / IrishPatternRuleTest: honest gap (pattern sector), same bar as RO/EO/FA roots.
type IrishLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Irish is the default Irish language surface (Java new Irish() / Irish.getInstance()).
var Irish = IrishLang{
	ShortCode:     "ga",
	Name:          "Irish",
	SpellerRuleID: "MORFOLOGIK_RULE_GA_IE",
	Countries:     []string{"IE"},
}

// NewIrish ports new Irish().
func NewIrish() IrishLang {
	return Irish
}

// IrishShortCode ports Irish.getShortCode ("ga").
func IrishShortCode() string { return "ga" }

func (v IrishLang) GetName() string { return v.Name }

// GetShortCode ports Irish.getShortCode ("ga").
func (v IrishLang) GetShortCode() string { return "ga" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Java countries {"IE"} (len==1) → "ga-IE".
func (v IrishLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("ga", v.Countries, "")
}

func (v IrishLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ga/common_words.txt.
func (v IrishLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule /
// MorfologikIrishSpellerRule.getId → MORFOLOGIK_RULE_GA_IE.
func (v IrishLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_GA_IE"
}

// GetMaintainedState ports Irish.getMaintainedState → ActivelyMaintained.
func (v IrishLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Irish.getMaintainers (four contributors, exact Java names).
func (v IrishLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Jim O'Regan"),
		NewContributor("Emily Barnes"),
		NewContributor("Mícheál J. Ó Meachair"),
		NewContributor("Seanán Ó Coistín"),
	}
}

// GetDefaultLanguageVariantCode ports Irish.getDefaultLanguageVariant → getInstance() (self).
// Package helper uses short code "ga" (Languages registry / getInstance LANGUAGE_SHORT_CODE).
func (v IrishLang) GetDefaultLanguageVariantCode() string {
	return IrishDefaultLanguageVariantCode()
}

// CreateDefaultTagger ports Irish.createDefaultTagger → new IrishTagger().
// Loads official /ga/irish.dict when discoverable (Java BaseTagger path).
// Returns process DefaultIrishTagger (MapWordTagger until dict loads).
func (v IrishLang) CreateDefaultTagger() *gatag.IrishTagger {
	gatag.EnsureDefaultIrishTagger()
	return gatag.DefaultIrishTagger
}

// CreateDefaultSentenceTokenizer ports Irish.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "ga".
func (v IrishLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Irish.createDefaultWordTokenizer
// → new WordTokenizer() (core, not GA-specific).
func (v IrishLang) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

// CreateDefaultDisambiguator ports Irish.createDefaultDisambiguator
// → new IrishHybridDisambiguator() (multiwords then XML).
func (v IrishLang) CreateDefaultDisambiguator() *gadis.IrishHybridDisambiguator {
	return gadis.NewIrishHybridDisambiguator()
}

// CreateDefaultSynthesizer ports Irish.createDefaultSynthesizer
// → IrishSynthesizer.INSTANCE (lazy official load).
// Nil if irish_synth.dict cannot be opened (fail-closed).
func (v IrishLang) CreateDefaultSynthesizer() *gasynth.IrishSynthesizer {
	return gasynth.Instance()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "ga"
// (Java Irish does not add RULE_FILES extras).
func (v IrishLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v IrishLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
