package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/diacritics.txt
var diacriticsFS embed.FS

// EnglishDiacriticsReplaceID ports EnglishDiacriticsRule.EN_DIACRITICS_REPLACE
// ("EN_DIACRITICS_REPLACE_ORTHOGRAPHY").
const EnglishDiacriticsReplaceID = "EN_DIACRITICS_REPLACE_ORTHOGRAPHY"

// EnglishDiacriticsFilePath is the official resource path Java loads
// (EnglishDiacriticsRule.FILE_NAME / getFileNames() → "/en/diacritics.txt").
const EnglishDiacriticsFilePath = "/en/diacritics.txt"

// EnglishDiacriticsURL ports EnglishDiacriticsRule.getUrl()
// → Tools.getUrl("https://en.wikipedia.org/wiki/English_terms_with_diacritical_marks").
const EnglishDiacriticsURL = "https://en.wikipedia.org/wiki/English_terms_with_diacritical_marks"

var (
	diacriticsOnce sync.Once
	diacriticsBase *rules.AbstractSimpleReplaceRule2
)

// loadDiacritics loads official diacritics.txt once (Java AbstractSimpleReplaceRule2 lazyInit).
func loadDiacritics() *rules.AbstractSimpleReplaceRule2 {
	diacriticsOnce.Do(func() {
		f, err := diacriticsFS.Open("data/diacritics.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of EnglishDiacriticsRule + AbstractSimpleReplaceRule2 defaults:
		// - id EN_DIACRITICS_REPLACE_ORTHOGRAPHY
		// - description / short / message as Java overrides
		// - suggestions separator " or "
		// - CaseSensitivy.CI (base default)
		// - Locale en via LanguageCode (getLocale → new Locale("en"))
		// - useSubRuleSpecificIds()
		// - Categories.TYPOS, ITSIssueType.Misspelling
		// - URL Wikipedia English terms with diacritical marks
		// - example blase → blasé
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   EnglishDiacriticsReplaceID,
			Description:          "Suggest diacritics for '$match'",
			ShortMsg:             "The original word has a diacritic",
			MessageTemplate:      "'$match' is an imported foreign name or expression, which originally has a diacritic.",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   true,
			Category:             rules.CatTypos.GetCategory(nil),
			IssueType:            rules.ITSMisspelling,
			URL:                  EnglishDiacriticsURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, EnglishDiacriticsFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("<marker>blase</marker>"),
			rules.Fixed("<marker>blasé</marker>"),
		)
		diacriticsBase = base
	})
	return diacriticsBase
}

// EnglishDiacriticsRule ports org.languagetool.rules.en.EnglishDiacriticsRule
// (extends AbstractSimpleReplaceRule2).
type EnglishDiacriticsRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewEnglishDiacriticsRule ports EnglishDiacriticsRule(ResourceBundle messages).
// Official data is go:embed of the same diacritics.txt Java loads from the classpath.
func NewEnglishDiacriticsRule(messages map[string]string) *EnglishDiacriticsRule {
	base := loadDiacritics()
	r := *base
	r.Messages = messages
	r.Category = rules.CatTypos.GetCategory(messages)
	r.IssueType = rules.ITSMisspelling
	r.URL = EnglishDiacriticsURL
	return &EnglishDiacriticsRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports EnglishDiacriticsRule.getFileNames() → singleton "/en/diacritics.txt".
func (r *EnglishDiacriticsRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{EnglishDiacriticsFilePath}
}

// GetShort ports EnglishDiacriticsRule.getShort().
func (r *EnglishDiacriticsRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports EnglishDiacriticsRule.getMessage().
func (r *EnglishDiacriticsRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports EnglishDiacriticsRule.getSuggestionsSeparator() → " or ".
func (r *EnglishDiacriticsRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports EnglishDiacriticsRule.getLocale() → Locale("en").
func (r *EnglishDiacriticsRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from diacritics.txt).
func (r *EnglishDiacriticsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
