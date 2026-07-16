package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractFillerWordsRule is a simplified port of org.languagetool.rules.AbstractFillerWordsRule.
// When MinPercent is 0 (default for this port), every filler token is reported.
// Percentage-based paragraph filtering is not modeled yet.
type AbstractFillerWordsRule struct {
	Messages    map[string]string
	ID          string
	Description string
	ShortMsg    string
	Message     string
	FillerWords map[string]struct{}
	// MinPercent: 0 reports all fillers (Java default-off rules often use 0 for "show all").
	MinPercent int
	// IsException optional skip.
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool
}

func (r *AbstractFillerWordsRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "FILLER_WORDS"
}

func (r *AbstractFillerWordsRule) isFiller(tok string) bool {
	_, ok := r.FillerWords[strings.ToLower(tok)]
	return ok
}

// Match flags filler tokens (minPercent==0 mode).
func (r *AbstractFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if r.MinPercent != 0 {
		// Percentage mode not implemented — no matches rather than false confidence.
		return nil
	}
	var out []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	msg := r.Message
	if msg == "" {
		msg = "Filler word"
	}
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i].GetToken()
		if !r.isFiller(tok) {
			continue
		}
		if r.IsException != nil && r.IsException(tokens, i) {
			continue
		}
		rm := NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.ShortMessage = r.ShortMsg
		out = append(out, rm)
	}
	return out
}
