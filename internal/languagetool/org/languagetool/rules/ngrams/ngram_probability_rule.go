package ngrams

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// NgramRuleID ports NgramProbabilityRule.RULE_ID.
const NgramRuleID = "NGRAM_RULE"

// NgramProbabilityRule ports org.languagetool.rules.ngrams.NgramProbabilityRule surface:
// flags rare 3-grams via LanguageModel.GetPseudoProbability.
type NgramProbabilityRule struct {
	LM             LanguageModel
	MinProbability float64
	// Tokenize defaults to WordTokenizer if nil.
	Tokenize func(string) []string
}

func NewNgramProbabilityRule(lm LanguageModel) *NgramProbabilityRule {
	return &NgramProbabilityRule{
		LM:             lm,
		MinProbability: 1e-14,
	}
}

func (r *NgramProbabilityRule) GetID() string { return NgramRuleID }

func (r *NgramProbabilityRule) SetMinProbability(p float64) { r.MinProbability = p }

func (r *NgramProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.LM == nil || sentence == nil {
		return nil
	}
	tok := r.Tokenize
	if tok == nil {
		wt := tokenizers.NewWordTokenizer()
		tok = wt.Tokenize
	}
	gTokens := GetGoogleTokensFromSentence(sentence, true, tok)
	var matches []*rules.RuleMatch
	for i := 1; i < len(gTokens)-1; i++ {
		prev := gTokens[i-1]
		cur := gTokens[i]
		next := gTokens[i+1]
		p := r.LM.GetPseudoProbability([]string{prev.Token, cur.Token, next.Token})
		if p.GetProb() < r.MinProbability && p.GetCoverage() > 0 {
			m := rules.NewRuleMatch(r, sentence, cur.StartPos, cur.EndPos,
				"This phrase is rare according to ngram statistics.")
			matches = append(matches, m)
		}
	}
	return matches
}
