package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	asttag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ast"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// AsturianLang ports org.languagetool.language.Asturian metadata + default factories.
// Java @Deprecated since 3.6 (unmaintained); getMaintainedState not overridden →
// Language default LookingForNewMaintainer.
type AsturianLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Asturian is the default Asturian language surface (Java new Asturian()).
var Asturian = AsturianLang{
	ShortCode:     "ast",
	Name:          "Asturian",
	SpellerRuleID: "MORFOLOGIK_RULE_AST",
	Countries:     []string{"ES"},
}

// NewAsturian ports new Asturian().
func NewAsturian() AsturianLang {
	return Asturian
}

// AsturianShortCode ports Asturian.getShortCode ("ast").
func AsturianShortCode() string { return "ast" }

func (v AsturianLang) GetName() string { return v.Name }

// GetShortCode ports Asturian.getShortCode ("ast").
func (v AsturianLang) GetShortCode() string { return "ast" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v AsturianLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("ast", v.Countries, "")
}

func (v AsturianLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ast/common_words.txt.
func (v AsturianLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikAsturianSpellerRule.getId
// → MORFOLOGIK_RULE_AST.
func (v AsturianLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_AST"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Asturian is @Deprecated (unmaintained since 3.6) but does not override getMaintainedState.
func (v AsturianLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Asturian.getMaintainers → Xesús González Rato.
func (v AsturianLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Xesús González Rato")}
}

// CreateDefaultTagger ports Asturian.createDefaultTagger → new AsturianTagger().
// Java: super("/ast/asturian.dict", new Locale("ast")) via BaseTagger default
// tagLowercaseWithUppercase=true. Dict load is the tagging/ast leaf (MapWordTagger
// until official binary is discoverable — same fail-closed pattern as other langs).
func (v AsturianLang) CreateDefaultTagger() *asttag.AsturianTagger {
	return asttag.NewAsturianTagger(tagging.MapWordTagger{})
}

// CreateDefaultSentenceTokenizer ports Asturian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "ast".
func (v AsturianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Language.createDefaultWordTokenizer default
// (Java Asturian does not override → WORD_TOKENIZER / new WordTokenizer()).
func (v AsturianLang) CreateDefaultWordTokenizer() tokenizers.Tokenizer {
	return tokenizers.NewWordTokenizer()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "ast"
// (Java Asturian does not add RULE_FILES extras).
func (v AsturianLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v AsturianLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
