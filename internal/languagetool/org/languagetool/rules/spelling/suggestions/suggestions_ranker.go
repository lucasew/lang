package suggestions

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SuggestionsRanker ports org.languagetool.rules.spelling.suggestions.SuggestionsRanker.
// Implementing types must attach confidence via SuggestedReplacement.
type SuggestionsRanker interface {
	SuggestionsOrderer
	ShouldAutoCorrect(ranked []*rules.SuggestedReplacement) bool
}

// ThresholdSuggestionsRanker auto-corrects when the top suggestion's
// confidence is set and at least MinConfidence, and strictly higher than the second.
type ThresholdSuggestionsRanker struct {
	Orderer       SuggestionsOrderer
	MinConfidence float32
}

func NewThresholdSuggestionsRanker(min float32) *ThresholdSuggestionsRanker {
	return &ThresholdSuggestionsRanker{
		Orderer:       EditDistanceSuggestionsOrderer{},
		MinConfidence: min,
	}
}

func (r *ThresholdSuggestionsRanker) IsMlAvailable() bool {
	if r == nil || r.Orderer == nil {
		return false
	}
	return r.Orderer.IsMlAvailable()
}

func (r *ThresholdSuggestionsRanker) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	if r == nil || r.Orderer == nil {
		return IdentitySuggestionsOrderer{}.OrderSuggestions(suggestions, word, sentence, startPos)
	}
	return r.Orderer.OrderSuggestions(suggestions, word, sentence, startPos)
}

func (r *ThresholdSuggestionsRanker) ShouldAutoCorrect(ranked []*rules.SuggestedReplacement) bool {
	if r == nil || len(ranked) == 0 {
		return false
	}
	top := ranked[0]
	if top == nil || top.Confidence == nil {
		return false
	}
	if *top.Confidence < r.MinConfidence {
		return false
	}
	if len(ranked) == 1 {
		return true
	}
	second := ranked[1]
	if second == nil || second.Confidence == nil {
		return true
	}
	return *top.Confidence > *second.Confidence
}

var _ SuggestionsRanker = (*ThresholdSuggestionsRanker)(nil)
