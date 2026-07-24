package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/wordiness.txt
var wordinessFS embed.FS

// EnglishPlainEnglishReplaceID ports EnglishPlainEnglishRule.EN_PLAIN_ENGLISH_REPLACE.
const EnglishPlainEnglishReplaceID = "EN_PLAIN_ENGLISH_REPLACE"

// EnglishPlainEnglishFilePath is the official resource path Java loads
// (EnglishPlainEnglishRule.FILE_NAME / getFileNames() → "/en/wordiness.txt").
const EnglishPlainEnglishFilePath = "/en/wordiness.txt"

// EnglishPlainEnglishURL ports EnglishPlainEnglishRule.getUrl()
// → Tools.getUrl("https://en.wikipedia.org/wiki/List_of_plain_English_words_and_phrases").
const EnglishPlainEnglishURL = "https://en.wikipedia.org/wiki/List_of_plain_English_words_and_phrases"

var (
	plainOnce sync.Once
	plainBase *rules.AbstractSimpleReplaceRule2
)

// loadPlainEnglish loads official wordiness.txt once (Java AbstractSimpleReplaceRule2 lazyInit).
func loadPlainEnglish() *rules.AbstractSimpleReplaceRule2 {
	plainOnce.Do(func() {
		f, err := wordinessFS.Open("data/wordiness.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of EnglishPlainEnglishRule + AbstractSimpleReplaceRule2 defaults:
		// - id EN_PLAIN_ENGLISH_REPLACE
		// - description / short / message as Java overrides
		// - suggestions separator " or "
		// - CaseSensitivy.CI (base default)
		// - Locale en via LanguageCode (getLocale → new Locale("en"))
		// - does NOT call useSubRuleSpecificIds()
		// - Categories.PLAIN_ENGLISH, ITSIssueType.Style
		// - setDefaultOff()
		// - URL Wikipedia list of plain English words and phrases
		// - example fatal outcome → death
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   EnglishPlainEnglishReplaceID,
			Description:          "1. Wordiness (General)",
			ShortMsg:             "Wordiness",
			MessageTemplate:      "'$match' is a wordy or complex expression. In some cases, it might be preferable to use $suggestions.",
			SuggestionsSeparator: " or ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "en",
			SubRuleSpecificIDs:   false,
			Category:             rules.CatPlainEnglish.GetCategory(nil),
			IssueType:            rules.ITSStyle,
			DefaultOff:           true,
			URL:                  EnglishPlainEnglishURL,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, EnglishPlainEnglishFilePath); err != nil {
			panic(err)
		}
		base.AddExamplePair(
			rules.Wrong("<marker>fatal outcome</marker>"),
			rules.Fixed("<marker>death</marker>"),
		)
		plainBase = base
	})
	return plainBase
}

// EnglishPlainEnglishRule ports org.languagetool.rules.en.EnglishPlainEnglishRule
// (extends AbstractSimpleReplaceRule2).
type EnglishPlainEnglishRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewEnglishPlainEnglishRule ports EnglishPlainEnglishRule(ResourceBundle messages).
// Official data is go:embed of the same wordiness.txt Java loads from the classpath.
func NewEnglishPlainEnglishRule(messages map[string]string) *EnglishPlainEnglishRule {
	base := loadPlainEnglish()
	r := *base
	r.Messages = messages
	r.Category = rules.CatPlainEnglish.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.DefaultOff = true
	r.URL = EnglishPlainEnglishURL
	return &EnglishPlainEnglishRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports EnglishPlainEnglishRule.getFileNames() → singleton "/en/wordiness.txt".
func (r *EnglishPlainEnglishRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{EnglishPlainEnglishFilePath}
}

// GetShort ports EnglishPlainEnglishRule.getShort().
func (r *EnglishPlainEnglishRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports EnglishPlainEnglishRule.getMessage().
func (r *EnglishPlainEnglishRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetSuggestionsSeparator ports EnglishPlainEnglishRule.getSuggestionsSeparator() → " or ".
func (r *EnglishPlainEnglishRule) GetSuggestionsSeparator() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.SuggestionsSeparator
}

// GetLocaleCode ports EnglishPlainEnglishRule.getLocale() → Locale("en").
func (r *EnglishPlainEnglishRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from wordiness.txt).
func (r *EnglishPlainEnglishRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
