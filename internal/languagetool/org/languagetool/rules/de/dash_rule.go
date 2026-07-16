package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DashRule ports org.languagetool.rules.de.DashRule.
// Flags hyphen compounds with a space after the hyphen (e.g. "Diäten- Erhöhung").
type DashRule struct {
	Messages map[string]string
}

func NewDashRule(messages map[string]string) *DashRule {
	return &DashRule{Messages: messages}
}

func (r *DashRule) GetID() string { return "DE_DASH" }

func (r *DashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	var prevToken string
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if prevToken != "" &&
			strings.HasSuffix(prevToken, "-") &&
			prevToken != "-" &&
			!strings.Contains(prevToken, "--") &&
			!strings.Contains(prevToken, "–-") {
			if token == "" {
				prevToken = token
				continue
			}
			first, _ := firstRune(token)
			if unicode.IsUpper(first) {
				up := strings.ToUpper(token)
				if up != "UND" && up != "ODER" && up != "BZW" {
					msg := "Möglicherweise fehlt ein 'und' oder ein Komma, oder es wurde nach dem Wort " +
						"ein überflüssiges Leerzeichen eingefügt. Eventuell haben Sie auch versehentlich einen Bindestrich statt eines Punktes eingefügt."
					fromPos := tokens[i-1].GetStartPos()
					toPos := tokens[i].GetEndPos()
					rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
					rm.ShortMessage = "Fehlendes 'und' oder Komma oder überflüssiges Leerzeichen?"
					joined := tokens[i-1].GetToken() + tokens[i].GetToken()
					hyphens := strings.Count(tokens[i-1].GetToken(), "-") + strings.Count(tokens[i].GetToken(), "-")
					if hyphens <= 1 {
						rm.SetSuggestedReplacements([]string{
							joined,
							tokens[i-1].GetToken() + ", " + tokens[i].GetToken(),
						})
					} else {
						rm.SetSuggestedReplacement(joined)
					}
					ruleMatches = append(ruleMatches, rm)
				}
			}
		}
		prevToken = token
	}
	return ruleMatches
}

func firstRune(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}
