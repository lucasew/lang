package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SentenceWithModalVerbRule is a surface stand-in for SentenceWithModalVerbRule.
// Without VER:MOD/VER:INF tags, common modal forms are matched; show-all when MinPercent is 0.
type SentenceWithModalVerbRule struct {
	Messages   map[string]string
	MinPercent int
	modals     map[string]struct{}
}

func NewSentenceWithModalVerbRule(messages map[string]string) *SentenceWithModalVerbRule {
	// common German modal conjugations
	list := []string{
		"kann", "kannst", "können", "könnt", "konnte", "konntest", "konnten", "konntet", "könnte", "könntest", "könnten", "könntet",
		"muss", "muß", "musst", "mußt", "müssen", "müsst", "müßt", "musste", "mußte", "musstest", "mussten", "müsste", "müßten",
		"soll", "sollst", "sollen", "sollt", "sollte", "solltest", "sollten", "solltet",
		"will", "willst", "wollen", "wollt", "wollte", "wolltest", "wollten", "wolltet",
		"darf", "darfst", "dürfen", "dürft", "durfte", "durftest", "durften", "dürfte", "dürftest", "dürften",
		"mag", "magst", "mögen", "mögt", "mochte", "mochtest", "mochten", "möchte", "möchtest", "möchten", "möchtet",
	}
	m := map[string]struct{}{}
	for _, w := range list {
		m[w] = struct{}{}
	}
	return &SentenceWithModalVerbRule{Messages: messages, MinPercent: 0, modals: m}
}

func (r *SentenceWithModalVerbRule) GetID() string { return "SENTENCE_WITH_MODAL_VERB_DE" }

func (r *SentenceWithModalVerbRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.MinPercent != 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		lc := strings.ToLower(tokens[i].GetToken())
		if _, ok := r.modals[lc]; !ok {
			continue
		}
		msg := "Modalverb: Modalverben blähen den Text häufig auf und sollten vermieden werden."
		rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.ShortMessage = "Modalverb"
		return []*rules.RuleMatch{rm}
	}
	return nil
}
