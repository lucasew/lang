package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishWordRepeatBeginningRule ports org.languagetool.rules.en.EnglishWordRepeatBeginningRule.
type EnglishWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var (
	addAdverbs = map[string]bool{
		"Additionally": true, "Besides": true, "Furthermore": true, "Moreover": true, "Also": true,
	}
	contrastAdverbs = map[string]bool{
		"Nevertheless": true, "Nonetheless": true, "Alternatively": true,
	}
	emphasisAdverbs = map[string]bool{
		"Undoubtedly": true, "Indeed": true, "Obviously": true, "Clearly": true,
		"Importantly": true, "Absolutely": true, "Definitely": true,
	}
	explainAdverbs = map[string]bool{
		"Particularly": true, "Especially": true, "Specifically": true,
	}
	addExpressions      = []string{"In addition", "As well as"}
	contrastExpressions = []string{"Even so", "On the other hand"}
	// personal pronouns (surface fallback for PRP tag)
	personalPronouns = map[string]bool{
		"I": true, "You": true, "He": true, "She": true, "It": true, "We": true, "They": true,
		"Me": true, "Him": true, "Her": true, "Us": true, "Them": true,
	}
)

func NewEnglishWordRepeatBeginningRule(messages map[string]string) *EnglishWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "ENGLISH_WORD_REPEAT_BEGINNING_RULE"
	r := &EnglishWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsExceptionFn = r.isException
	base.IsAdverbFn = r.isAdverb
	base.GetSuggestionsFn = r.getSuggestions
	return r
}

func (r *EnglishWordRepeatBeginningRule) isException(token string) bool {
	return token == "The" || token == "A" || token == "An"
}

func (r *EnglishWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	tok := token.GetToken()
	return addAdverbs[tok] || contrastAdverbs[tok] || emphasisAdverbs[tok] || explainAdverbs[tok]
}

func (r *EnglishWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	tok := token.GetToken()
	if token.HasPosTag("PRP") || personalPronouns[tok] {
		adapted := tok
		if tok != "I" {
			adapted = strings.ToLower(tok)
		}
		return []string{
			"Furthermore, " + adapted,
			"Likewise, " + adapted,
			"Not only that, but " + adapted,
		}
	}
	if addAdverbs[tok] {
		s := differentAdverbs(tok, addAdverbs)
		s = append(s, addExpressions...)
		return s
	}
	if contrastAdverbs[tok] {
		s := differentAdverbs(tok, contrastAdverbs)
		s = append(s, contrastExpressions...)
		return s
	}
	if emphasisAdverbs[tok] {
		return differentAdverbs(tok, emphasisAdverbs)
	}
	if explainAdverbs[tok] {
		return differentAdverbs(tok, explainAdverbs)
	}
	return nil
}

func differentAdverbs(adverb string, category map[string]bool) []string {
	var out []string
	for adv := range category {
		if adv != adverb {
			out = append(out, adv)
		}
	}
	return out
}
