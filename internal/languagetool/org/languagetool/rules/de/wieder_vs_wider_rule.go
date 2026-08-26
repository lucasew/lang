package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WiederVsWiderRule ports org.languagetool.rules.de.WiederVsWiderRule.
// Java match gate: tokens[i].hasLemma("spiegeln") only (no surface invent).
// Java category: Categories.TYPOS.
type WiederVsWiderRule struct {
	Messages map[string]string
	Category *rules.Category
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewWiederVsWiderRule ports WiederVsWiderRule(ResourceBundle messages).
func NewWiederVsWiderRule(messages map[string]string) *WiederVsWiderRule {
	r := &WiederVsWiderRule{
		Messages: messages,
		Category: rules.CatTypos.GetCategory(messages),
	}
	// Java: addExamplePair(Example.wrong(...wieder...), Example.fixed(...wider...))
	r.AddExamplePair(
		rules.Wrong("Das spiegelt die Situation in Deutschland <marker>wieder</marker>."),
		rules.Fixed("Das spiegelt die Situation in Deutschland <marker>wider</marker>."),
	)
	return r
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

// AddExamplePair ports Rule.addExamplePair.
func (r *WiederVsWiderRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *WiederVsWiderRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *WiederVsWiderRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// EstimateContextForSureMatch ports WiederVsWiderRule.estimateContextForSureMatch → 0.
func (r *WiederVsWiderRule) EstimateContextForSureMatch() int { return 0 }

// Match ports WiederVsWiderRule.match — bug-for-bug with Java look-ahead gate:
// !(tokens.length > i+2 && (tokens[i+1]=="wider" || tokens[i+2]=="wider")).
// When length ≤ i+2 the look-ahead is not applied (Java quirk kept).
func (r *WiederVsWiderRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	foundSpiegelt := false
	foundWieder := false
	foundWider := false
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		token := tokens[i].GetToken()
		// Java: if / else-if chain (hasLemma("spiegeln"), then wieder, then wider)
		if tokens[i].HasLemma("spiegeln") {
			foundSpiegelt = true
		} else if strings.EqualFold(token, "wieder") && foundSpiegelt {
			foundWieder = true
		} else if strings.EqualFold(token, "wider") && foundSpiegelt {
			foundWider = true
		}
		// Java: !(tokens.length > i + 2 && (i+1 wider || i+2 wider))
		// Case-sensitive "wider" on look-ahead (Java .equals, not equalsIgnoreCase).
		widerSoon := false
		if len(tokens) > i+2 {
			t1 := tokens[i+1]
			t2 := tokens[i+2]
			if t1 != nil && t1.GetToken() == "wider" {
				widerSoon = true
			}
			if t2 != nil && t2.GetToken() == "wider" {
				widerSoon = true
			}
		}
		if foundSpiegelt && foundWieder && !foundWider && !widerSoon {
			msg := "'wider' in 'widerspiegeln' wird mit 'i' statt mit 'ie' " +
				"geschrieben, z.B. 'Das spiegelt die Situation gut wider.'"
			shortMsg := "'wider' in 'widerspiegeln' wird mit 'i' geschrieben"
			// Java: pos = startPos; end = pos + token.length() (UTF-16) ≡ GetEndPos()
			pos := tokens[i].GetStartPos()
			rm := rules.NewRuleMatch(r, sentence, pos, tokens[i].GetEndPos(), msg)
			rm.ShortMessage = shortMsg
			rm.SetSuggestedReplacement("wider")
			ruleMatches = append(ruleMatches, rm)
			foundSpiegelt = false
			foundWieder = false
			foundWider = false
		}
	}
	return ruleMatches
}
