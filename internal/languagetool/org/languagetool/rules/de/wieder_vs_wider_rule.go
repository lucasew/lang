package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WiederVsWiderRule ports org.languagetool.rules.de.WiederVsWiderRule.
// Java: tokens[i].hasLemma("spiegeln") only (no surface invent).
// Java: TYPOS category.
type WiederVsWiderRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewWiederVsWiderRule(messages map[string]string) *WiederVsWiderRule {
	return &WiederVsWiderRule{
		Messages: messages,
		Category: rules.CatTypos.GetCategory(messages),
	}
}

func (r *WiederVsWiderRule) GetID() string { return "DE_WIEDER_VS_WIDER" }

func (r *WiederVsWiderRule) GetDescription() string {
	return "Möglicher Tippfehler 'spiegeln ... wieder (wider)'"
}

func (r *WiederVsWiderRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// EstimateContextForSureMatch ports WiederVsWiderRule.estimateContextForSureMatch → 0.
func (r *WiederVsWiderRule) EstimateContextForSureMatch() int { return 0 }

func isSpiegelnToken(t *languagetool.AnalyzedTokenReadings) bool {
	return t != nil && t.HasAnyLemma("spiegeln")
}

func (r *WiederVsWiderRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	foundSpiegelt, foundWieder, foundWider := false, false, false
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		token := tokens[i].GetToken()
		// Java: if / else-if chain (lemma spiegeln, then wieder, then wider)
		if isSpiegelnToken(tokens[i]) {
			foundSpiegelt = true
		} else if strings.EqualFold(token, "wieder") && foundSpiegelt {
			foundWieder = true
		} else if strings.EqualFold(token, "wider") && foundSpiegelt {
			foundWider = true
		}
		// Java: !(tokens.length > i + 2 && (i+1 wider || i+2 wider))
		// Note: Java only checks when length > i+2, so wider at i+1 alone is not excluded
		// when there is no token at i+2 — keep that exact gate (no invent).
		widerSoon := false
		if len(tokens) > i+2 {
			if tokens[i+1] != nil && tokens[i+1].GetToken() == "wider" {
				widerSoon = true
			}
			if tokens[i+2] != nil && tokens[i+2].GetToken() == "wider" {
				widerSoon = true
			}
		}
		if foundSpiegelt && foundWieder && !foundWider && !widerSoon {
			msg := "'wider' in 'widerspiegeln' wird mit 'i' statt mit 'ie' " +
				"geschrieben, z.B. 'Das spiegelt die Situation gut wider.'"
			shortMsg := "'wider' in 'widerspiegeln' wird mit 'i' geschrieben"
			// Java marks current token span (token.length), not always "wieder"
			rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			rm.ShortMessage = shortMsg
			rm.SetSuggestedReplacement("wider")
			ruleMatches = append(ruleMatches, rm)
			foundSpiegelt, foundWieder, foundWider = false, false, false
		}
	}
	return ruleMatches
}
