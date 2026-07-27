package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	tageo "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/eo"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	eotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/eo"
)

// EsperantoLang ports org.languagetool.language.Esperanto metadata + default factories.
// Java overrides getMaintainedState → ActivelyMaintained.
// Java getCountries → new String[]{} (empty).
// No createDefaultSynthesizer (Java has none — do not invent).
// DateCheckFilter exists under rules/eo but is NOT in getRelevantRules — do not invent.
type EsperantoLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Esperanto is the default Esperanto language surface (Java new Esperanto()).
var Esperanto = EsperantoLang{
	ShortCode:     "eo",
	Name:          "Esperanto",
	SpellerRuleID: "HUNSPELL_RULE",
	Countries:     nil, // Java getCountries → new String[]{}
}

// NewEsperanto ports new Esperanto().
func NewEsperanto() EsperantoLang {
	return Esperanto
}

// EsperantoShortCode ports Esperanto.getShortCode ("eo").
func EsperantoShortCode() string { return "eo" }

func (v EsperantoLang) GetName() string { return v.Name }

// GetShortCode ports Esperanto.getShortCode ("eo").
func (v EsperantoLang) GetShortCode() string { return "eo" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Empty countries → "eo".
func (v EsperantoLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("eo", v.Countries, "")
}

func (v EsperantoLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → eo/common_words.txt.
func (v EsperantoLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / HunspellRule.getId
// → HUNSPELL_RULE (Java new HunspellRule(messages, this, null, null)).
func (v EsperantoLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "HUNSPELL_RULE"
}

// GetMaintainedState ports Esperanto.getMaintainedState → ActivelyMaintained.
func (v EsperantoLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Esperanto.getMaintainers → Contributors.DOMINIQUE_PELLE.
func (v EsperantoLang) GetMaintainers() []Contributor {
	return []Contributor{DominiquePelle}
}

// CreateDefaultTagger ports Esperanto.createDefaultTagger → new EsperantoTagger().
func (v EsperantoLang) CreateDefaultTagger() *tageo.EsperantoTagger {
	return tageo.NewEsperantoTagger()
}

// CreateDefaultSentenceTokenizer ports Esperanto.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "eo".
func (v EsperantoLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Esperanto.createDefaultWordTokenizer
// → new EsperantoWordTokenizer().
func (v EsperantoLang) CreateDefaultWordTokenizer() *eotok.EsperantoWordTokenizer {
	return eotok.NewEsperantoWordTokenizer()
}

// CreateDefaultDisambiguator ports Esperanto.createDefaultDisambiguator
// → new XmlRuleDisambiguator(new Esperanto()) (useGlobalDisambiguation default false).
// Nil if eo/disambiguation.xml cannot be loaded (fail-closed).
func (v EsperantoLang) CreateDefaultDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return tageo.EsperantoXmlRuleDisambiguator()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "eo"
// (Java Esperanto does not add RULE_FILES extras).
func (v EsperantoLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v EsperantoLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
