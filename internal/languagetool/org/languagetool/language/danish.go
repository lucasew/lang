package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	datag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/da"
	dadis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/da"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// DanishLang ports org.languagetool.language.Danish metadata + default factories.
// Java @Deprecated since 3.6 (unmaintained); getMaintainedState not overridden →
// Language default LookingForNewMaintainer.
// No createDefaultWordTokenizer / createDefaultSynthesizer (Java has none — do not invent).
type DanishLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Danish is the default Danish language surface (Java new Danish()).
var Danish = DanishLang{
	ShortCode:     "da",
	Name:          "Danish",
	SpellerRuleID: "HUNSPELL_RULE",
	Countries:     []string{"DK"},
}

// NewDanish ports new Danish().
func NewDanish() DanishLang {
	return Danish
}

// DanishShortCode ports Danish.getShortCode ("da").
func DanishShortCode() string { return "da" }

func (v DanishLang) GetName() string { return v.Name }

// GetShortCode ports Danish.getShortCode ("da").
func (v DanishLang) GetShortCode() string { return "da" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v DanishLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("da", v.Countries, "")
}

func (v DanishLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → da/common_words.txt.
func (v DanishLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / HunspellRule.RULE_ID
// → HUNSPELL_RULE.
func (v DanishLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "HUNSPELL_RULE"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Danish is @Deprecated (unmaintained since 3.6) but does not override getMaintainedState.
func (v DanishLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Danish.getMaintainers → Esben Aaberg + Henrik Bendt.
func (v DanishLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Esben Aaberg"),
		NewContributor("Henrik Bendt"),
	}
}

// CreateDefaultTagger ports Danish.createDefaultTagger → new DanishTagger().
// Loads official /da/danish.dict when discoverable (Java BaseTagger path).
// Returns process DefaultDanishTagger (MapWordTagger until dict loads).
func (v DanishLang) CreateDefaultTagger() *datag.DanishTagger {
	datag.EnsureDefaultDanishTagger()
	return datag.DefaultDanishTagger
}

// CreateDefaultSentenceTokenizer ports Danish.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "da".
func (v DanishLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Language.createDefaultWordTokenizer default
// (Java Danish does not override → WORD_TOKENIZER / new WordTokenizer()).
func (v DanishLang) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

// CreateDefaultDisambiguator ports Danish.createDefaultDisambiguator
// → new XmlRuleDisambiguator(this) (useGlobalDisambiguation default false).
// Nil if da/disambiguation.xml cannot be loaded (fail-closed).
func (v DanishLang) CreateDefaultDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return dadis.DanishXmlRuleDisambiguator()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "da"
// (Java Danish does not add RULE_FILES extras).
func (v DanishLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v DanishLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
