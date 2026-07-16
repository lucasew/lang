package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WiederVsWiderRule ports org.languagetool.rules.de.WiederVsWiderRule.
// Without a German tagger, lemma "spiegeln" is approximated by surface prefixes.
type WiederVsWiderRule struct {
	Messages map[string]string
}

func NewWiederVsWiderRule(messages map[string]string) *WiederVsWiderRule {
	return &WiederVsWiderRule{Messages: messages}
}

func (r *WiederVsWiderRule) GetID() string { return "DE_WIEDER_VS_WIDER" }

func isSpiegelnSurface(tok string) bool {
	return strings.HasPrefix(strings.ToLower(tok), "spiegel")
}

func (r *WiederVsWiderRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	foundSpiegelt, foundWieder, foundWider := false, false, false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if isSpiegelnSurface(token) {
			foundSpiegelt = true
		} else if strings.EqualFold(token, "wieder") && foundSpiegelt {
			foundWieder = true
		} else if strings.EqualFold(token, "wider") && foundSpiegelt {
			foundWider = true
		}
		// Java: skip if wider appears within next two tokens after current "wieder"
		widerSoon := len(tokens) > i+2 &&
			(strings.EqualFold(tokens[i+1].GetToken(), "wider") ||
				strings.EqualFold(tokens[i+2].GetToken(), "wider"))
		// also allow wider immediately next when near end (len == i+2)
		if !widerSoon && i+1 < len(tokens) && strings.EqualFold(tokens[i+1].GetToken(), "wider") {
			widerSoon = true
		}
		if foundSpiegelt && foundWieder && !foundWider && !widerSoon {
			msg := "'wider' in 'widerspiegeln' wird mit 'i' statt mit 'ie' " +
				"geschrieben, z.B. 'Das spiegelt die Situation gut wider.'"
			shortMsg := "'wider' in 'widerspiegeln' wird mit 'i' geschrieben"
			rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			rm.ShortMessage = shortMsg
			rm.SetSuggestedReplacement("wider")
			ruleMatches = append(ruleMatches, rm)
			foundSpiegelt, foundWieder, foundWider = false, false, false
		}
	}
	return ruleMatches
}
