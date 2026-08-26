package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_profanity.txt
var profanityFS embed.FS

// EnglishProfanityRuleID ports SimpleReplaceProfanityRule.getId() → "PROFANITY".
const EnglishProfanityRuleID = "PROFANITY"

// EnglishProfanityFilePath ports SimpleReplaceProfanityRule.getFileNames()[0]
// → "/en/replace_profanity.txt".
const EnglishProfanityFilePath = "/en/replace_profanity.txt"

// EnglishProfanityURL ports SimpleReplaceProfanityRule setUrl
// → Tools.getUrl("https://en.wiktionary.org/wiki/Category:English_offensive_terms").
const EnglishProfanityURL = "https://en.wiktionary.org/wiki/Category:English_offensive_terms"

var (
	profanityOnce sync.Once
	profanityBase *rules.AbstractSimpleReplaceRule2
)

// loadProfanity loads official replace_profanity.txt once
// (Java AbstractSimpleReplaceRule2.lazyInit / fillMaps over getFileNames()).
// File lines are phrase-only (no '='); setRuleHasSuggestions(false) permits that.
func loadProfanity() *rules.AbstractSimpleReplaceRule2 {
	profanityOnce.Do(func() {
		f, err := profanityFS.Open("data/replace_profanity.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Field twin of SimpleReplaceProfanityRule + AbstractSimpleReplaceRule2:
		// - id PROFANITY
		// - description / short "Profanity"
		// - message "This expression can be considered offensive."
		// - CaseSensitivy.CI (base default)
		// - Locale EN via LanguageCode (getLocale → new Locale("EN"))
		// - does NOT call useSubRuleSpecificIds() (commented out in Java)
		// - Categories.STYLE, ITSIssueType.Style
		// - setTags(Collections.singletonList(Tag.picky))
		// - setRuleHasSuggestions(false) — no replacements in file / match
		// - URL Wiktionary English offensive terms
		// - no example pairs in Java ctor
		noSug := false
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              EnglishProfanityRuleID,
			Description:     "Profanity",
			ShortMsg:        "Profanity",
			MessageTemplate: "This expression can be considered offensive.",
			// SuggestionsSeparator unused when ruleHasSuggestions is false
			CaseSens:           rules.CaseInsensitive,
			LanguageCode:       "EN",
			SubRuleSpecificIDs: false, // Java: //useSubRuleSpecificIds();
			Category:           rules.CatStyle.GetCategory(nil),
			IssueType:          rules.ITSStyle,
			Tags:               []rules.Tag{rules.TagPicky},
			URL:                EnglishProfanityURL,
			RuleHasSuggestions: &noSug,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, EnglishProfanityFilePath); err != nil {
			panic(err)
		}
		profanityBase = base
	})
	return profanityBase
}

// SimpleReplaceProfanityRule ports org.languagetool.rules.en.SimpleReplaceProfanityRule
// (extends AbstractSimpleReplaceRule2; suggestion-less dictionary matches).
type SimpleReplaceProfanityRule struct {
	*rules.AbstractSimpleReplaceRule2
}

// NewSimpleReplaceProfanityRule ports
// SimpleReplaceProfanityRule(ResourceBundle messages, Language language).
// Official data is go:embed of the same replace_profanity.txt Java loads.
// language is English; getLocale is Locale("EN") — stored as LanguageCode.
func NewSimpleReplaceProfanityRule(messages map[string]string) *SimpleReplaceProfanityRule {
	base := loadProfanity()
	r := *base
	r.Messages = messages
	r.Category = rules.CatStyle.GetCategory(messages)
	r.IssueType = rules.ITSStyle
	r.URL = EnglishProfanityURL
	// Tags / RuleHasSuggestions / SubRuleSpecificIDs already on base copy.
	return &SimpleReplaceProfanityRule{AbstractSimpleReplaceRule2: &r}
}

// GetFileNames ports SimpleReplaceProfanityRule.getFileNames()
// → singleton "/en/replace_profanity.txt".
func (r *SimpleReplaceProfanityRule) GetFileNames() []string {
	if r == nil {
		return nil
	}
	return []string{EnglishProfanityFilePath}
}

// GetShort ports SimpleReplaceProfanityRule.getShort() → "Profanity".
func (r *SimpleReplaceProfanityRule) GetShort() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.ShortMsg
}

// GetMessage ports SimpleReplaceProfanityRule.getMessage().
func (r *SimpleReplaceProfanityRule) GetMessage() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.MessageTemplate
}

// GetLocaleCode ports SimpleReplaceProfanityRule.getLocale() → Locale("EN").
func (r *SimpleReplaceProfanityRule) GetLocaleCode() string {
	if r == nil || r.AbstractSimpleReplaceRule2 == nil {
		return ""
	}
	return r.LanguageCode
}

// Match ports AbstractSimpleReplaceRule2.match (surface / multiword keys from
// replace_profanity.txt; no suggested replacements when ruleHasSuggestions false).
func (r *SimpleReplaceProfanityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
