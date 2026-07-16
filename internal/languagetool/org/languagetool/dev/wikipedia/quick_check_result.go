package wikipedia

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// WikipediaQuickCheckResult ports org.languagetool.dev.wikipedia.WikipediaQuickCheckResult.
type WikipediaQuickCheckResult struct {
	Text         string
	LanguageCode string
	RuleMatches  []*rules.RuleMatch
}

func NewWikipediaQuickCheckResult(text string, matches []*rules.RuleMatch, languageCode string) *WikipediaQuickCheckResult {
	return &WikipediaQuickCheckResult{
		Text:         text,
		LanguageCode: languageCode,
		RuleMatches:  matches,
	}
}

func (r *WikipediaQuickCheckResult) GetText() string                    { return r.Text }
func (r *WikipediaQuickCheckResult) GetLanguageCode() string            { return r.LanguageCode }
func (r *WikipediaQuickCheckResult) GetRuleMatches() []*rules.RuleMatch { return r.RuleMatches }
