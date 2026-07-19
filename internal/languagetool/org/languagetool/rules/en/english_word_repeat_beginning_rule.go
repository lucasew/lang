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
)

func NewEnglishWordRepeatBeginningRule(messages map[string]string) *EnglishWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "ENGLISH_WORD_REPEAT_BEGINNING_RULE"
	// Java: addExamplePair(Moreover… Moreover → Moreover… It)
	base.AddExamplePair(
		rules.Wrong("Moreover, the street is almost entirely residential. <marker>Moreover</marker>, it was named after a poet."),
		rules.Fixed("Moreover, the street is almost entirely residential. <marker>It</marker> was named after a poet."),
	)
	r := &EnglishWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsExceptionFn = r.isException
	base.IsAdverbFn = r.isAdverb
	base.GetSuggestionsFn = r.getSuggestions
	return r
}

func (r *EnglishWordRepeatBeginningRule) isException(token string) bool {
	return token == "The" || token == "A" || token == "An"
}

// isAdverb ports Java EnglishWordRepeatBeginningRule.isAdverb (fixed adverb sets only).
func (r *EnglishWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	tok := token.GetToken()
	return addAdverbs[tok] || contrastAdverbs[tok] || emphasisAdverbs[tok] || explainAdverbs[tok]
}

// getSuggestions ports Java getSuggestions — personal pronouns only via PRP POS (no surface invent).
func (r *EnglishWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	if token == nil {
		return nil
	}
	tok := token.GetToken()
	// Java: if (token.hasPosTag("PRP"))
	if token.HasPosTag("PRP") {
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
