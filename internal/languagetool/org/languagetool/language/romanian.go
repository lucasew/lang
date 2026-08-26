package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rosynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ro"
	rodis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ro"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	rotag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ro"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	rotok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ro"
)

// RomanianLang ports org.languagetool.language.Romanian metadata + default factories.
type RomanianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Romanian is the default Romanian language surface (Java new Romanian()).
var Romanian = RomanianLang{
	ShortCode:     "ro",
	Name:          "Romanian",
	SpellerRuleID: "MORFOLOGIK_RULE_RO_RO",
	Countries:     []string{"RO"},
}

// NewRomanian ports new Romanian().
func NewRomanian() RomanianLang {
	return Romanian
}

// RomanianShortCode ports Romanian.getShortCode ("ro").
func RomanianShortCode() string { return "ro" }

func (v RomanianLang) GetName() string { return v.Name }

// GetShortCode ports Romanian.getShortCode ("ro").
func (v RomanianLang) GetShortCode() string { return "ro" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v RomanianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("ro", v.Countries, "")
}

func (v RomanianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ro/common_words.txt.
func (v RomanianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikRomanianSpellerRule.getId
// → MORFOLOGIK_RULE_RO_RO.
func (v RomanianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_RO_RO"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Romanian does not override getMaintainedState.
func (v RomanianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Romanian.getMaintainers → Ionuț Păduraru + archeus URL.
func (v RomanianLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributorWithURL("Ionuț Păduraru", "http://www.archeus.ro")}
}

// CreateDefaultTagger ports Romanian.createDefaultTagger → new RomanianTagger().
// Loads official /ro/romanian.dict when discoverable (Java BaseTagger path).
// Returns process DefaultRomanianTagger (MapWordTagger until dict loads).
func (v RomanianLang) CreateDefaultTagger() *rotag.RomanianTagger {
	rotag.EnsureDefaultRomanianTagger()
	return rotag.DefaultRomanianTagger
}

// CreateDefaultSentenceTokenizer ports Romanian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "ro".
func (v RomanianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Romanian.createDefaultWordTokenizer
// → new RomanianWordTokenizer().
func (v RomanianLang) CreateDefaultWordTokenizer() *rotok.RomanianWordTokenizer {
	return rotok.NewRomanianWordTokenizer()
}

// CreateDefaultDisambiguator ports Romanian.createDefaultDisambiguator
// → new XmlRuleDisambiguator(this) (useGlobalDisambiguation default false).
// Nil if ro/disambiguation.xml cannot be loaded (fail-closed).
func (v RomanianLang) CreateDefaultDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return rodis.RomanianXmlRuleDisambiguator()
}

// CreateDefaultSynthesizer ports Romanian.createDefaultSynthesizer
// → RomanianSynthesizer.INSTANCE (lazy official load).
// Nil if romanian_synth.dict cannot be opened (fail-closed).
func (v RomanianLang) CreateDefaultSynthesizer() *rosynth.RomanianSynthesizer {
	return rosynth.Instance()
}
