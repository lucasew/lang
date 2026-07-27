package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	sksynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/sk"
	sktag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/sk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SlovakLang ports org.languagetool.language.Slovak metadata + default factories.
type SlovakLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Slovak is the default Slovak language surface (Java new Slovak()).
var Slovak = SlovakLang{
	ShortCode:     "sk",
	Name:          "Slovak",
	SpellerRuleID: "MORFOLOGIK_RULE_SK_SK",
	Countries:     []string{"SK"},
}

// NewSlovak ports new Slovak().
func NewSlovak() SlovakLang {
	return Slovak
}

// SlovakShortCode ports Slovak.getShortCode ("sk").
func SlovakShortCode() string { return "sk" }

func (v SlovakLang) GetName() string { return v.Name }

// GetShortCode ports Slovak.getShortCode ("sk").
func (v SlovakLang) GetShortCode() string { return "sk" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v SlovakLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("sk", v.Countries, "")
}

func (v SlovakLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → sk/common_words.txt.
func (v SlovakLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikSlovakSpellerRule.getId
// → MORFOLOGIK_RULE_SK_SK.
func (v SlovakLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_SK_SK"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Slovak does not override getMaintainedState.
func (v SlovakLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Slovak.getMaintainers → Zdenko Podobný + sk-spell URL.
func (v SlovakLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributorWithURL("Zdenko Podobný", "http://sk-spell.sk.cx")}
}

// CreateDefaultTagger ports Slovak.createDefaultTagger → new SlovakTagger().
// Loads official /sk/slovak.dict when discoverable (Java BaseTagger path).
// Returns process DefaultSlovakTagger (MapWordTagger until dict loads).
func (v SlovakLang) CreateDefaultTagger() *sktag.SlovakTagger {
	sktag.EnsureDefaultSlovakTagger()
	return sktag.DefaultSlovakTagger
}

// CreateDefaultSentenceTokenizer ports Slovak.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "sk".
func (v SlovakLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Language.createDefaultWordTokenizer default
// (Java Slovak does not override → WORD_TOKENIZER / new WordTokenizer()).
func (v SlovakLang) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

// CreateDefaultSynthesizer ports Slovak.createDefaultSynthesizer
// → SlovakSynthesizer.INSTANCE (lazy official load).
// Nil if slovak_synth.dict cannot be opened (fail-closed).
func (v SlovakLang) CreateDefaultSynthesizer() *sksynth.SlovakSynthesizer {
	return sksynth.Instance()
}

// GetRuleFileNames ports Slovak.getRuleFileNames — super base + RULE_FILES
// (grammar-typography.xml).
func (v SlovakLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v SlovakLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	out := languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
	// Slovak.RULE_FILES → grammar-typography.xml
	dirBase := "/org/languagetool/rules/" + code + "/"
	out = append(out, dirBase+"grammar-typography.xml")
	return out
}
