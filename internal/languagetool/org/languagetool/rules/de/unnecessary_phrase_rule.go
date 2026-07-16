package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnnecessaryPhraseRule ports phrase list from org.languagetool.rules.de.UnnecessaryPhraseRule
// in "show all" mode (no per-mill percentage filtering).
type UnnecessaryPhraseRule struct {
	Messages map[string]string
	phrases  [][]string
}

func NewUnnecessaryPhraseRule(messages map[string]string) *UnnecessaryPhraseRule {
	phrases := [][]string{
		{"dann", "und", "wann"},
		{"des", "Ungeachtet"},
		{"ganz", "und", "gar"},
		{"hie", "und", "da"},
		{"im", "Allgemeinen"},
		{"in", "der", "Tat"},
		{"in", "diesem", "Zusammenhang"},
		{"mehr", "oder", "weniger"},
		{"meines", "Erachtens"},
		{"ohne", "weiteres"},
		{"ohne", "Zweifel"},
		{"samt", "und", "sonders"},
		{"sowohl", "als", "auch"},
		{"voll", "und", "ganz"},
		{"von", "Neuem"},
		{"allem", "Anschein", "nach"},
		{"aufs", "Neue"},
		{"ein", "bisschen"},
		{"ein", "wenig"},
		{"des", "Öfteren"},
		{"bei", "weitem"},
		{"an", "sich"},
	}
	return &UnnecessaryPhraseRule{Messages: messages, phrases: phrases}
}

func (r *UnnecessaryPhraseRule) GetID() string { return "UNNECESSARY_PHRASE_DE" }

func (r *UnnecessaryPhraseRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		for _, phrase := range r.phrases {
			if i+len(phrase) > len(tokens) {
				continue
			}
			ok := true
			for j, w := range phrase {
				if !strings.EqualFold(tokens[i+j].GetToken(), w) {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			from := tokens[i].GetStartPos()
			to := tokens[i+len(phrase)-1].GetEndPos()
			msg := "Diese Wendung kann als Phrase wirken und oft weggelassen oder verkürzt werden."
			rm := rules.NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = "Phrase"
			matches = append(matches, rm)
		}
	}
	return matches
}
