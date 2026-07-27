package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/tl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	tltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/tl"
)

// TagalogLang ports org.languagetool.language.Tagalog metadata + default factories.
// Java @Deprecated since 3.6 (unmaintained); getMaintainedState not overridden →
// Language default LookingForNewMaintainer.
type TagalogLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// Tagalog is the default Tagalog language surface (Java new Tagalog()).
var Tagalog = TagalogLang{
	ShortCode:     "tl",
	Name:          "Tagalog",
	SpellerRuleID: "MORFOLOGIK_RULE_TL",
	Countries:     []string{"PH"},
}

// NewTagalog ports new Tagalog().
func NewTagalog() TagalogLang {
	return Tagalog
}

// TagalogShortCode ports Tagalog.getShortCode ("tl").
func TagalogShortCode() string { return "tl" }

func (v TagalogLang) GetName() string { return v.Name }

// GetShortCode ports Tagalog.getShortCode ("tl").
func (v TagalogLang) GetShortCode() string { return "tl" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v TagalogLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("tl", v.Countries, "")
}

func (v TagalogLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → tl/common_words.txt.
func (v TagalogLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule / MorfologikTagalogSpellerRule.getId
// → MORFOLOGIK_RULE_TL.
func (v TagalogLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_TL"
}

// GetMaintainedState ports Language.getMaintainedState default (LookingForNewMaintainer).
// Java Tagalog is @Deprecated (unmaintained since 3.6) but does not override getMaintainedState.
func (v TagalogLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.LookingForNewMaintainer
}

// GetMaintainers ports Tagalog.getMaintainers → Nathaniel Oco, Allan Borra.
func (v TagalogLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Nathaniel Oco"),
		NewContributor("Allan Borra"),
	}
}

// CreateDefaultTagger ports Tagalog.createDefaultTagger → new TagalogTagger().
// Loads official /tl/tagalog.dict when discoverable (Java BaseTagger path).
// Returns process DefaultTagalogTagger (MapWordTagger until dict loads).
func (v TagalogLang) CreateDefaultTagger() *tltag.TagalogTagger {
	tltag.EnsureDefaultTagalogTagger()
	return tltag.DefaultTagalogTagger
}

// CreateDefaultSentenceTokenizer ports Tagalog.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "tl".
func (v TagalogLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Tagalog.createDefaultWordTokenizer
// → new TagalogWordTokenizer().
func (v TagalogLang) CreateDefaultWordTokenizer() *tltok.TagalogWordTokenizer {
	return tltok.NewTagalogWordTokenizer()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "tl"
// (Java Tagalog does not add RULE_FILES extras).
func (v TagalogLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v TagalogLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
