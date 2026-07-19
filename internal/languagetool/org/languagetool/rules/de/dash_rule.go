package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DashRule ports org.languagetool.rules.de.DashRule.
// Flags hyphen compounds with a space after the hyphen (e.g. "Diäten- Erhöhung").
// Java: COMPOUNDING category + setUrl (not AbstractDashRule; DE-specific).
type DashRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewDashRule(messages map[string]string) *DashRule {
	return &DashRule{
		Messages: messages,
		Category: rules.CatCompounding.GetCategory(messages),
	}
}

func (r *DashRule) GetID() string { return "DE_DASH" }

// GetDescription ports DashRule.getDescription.
func (r *DashRule) GetDescription() string {
	return "Keine Leerzeichen in Bindestrich-Komposita (wie z.B. in 'Diäten- Erhöhung')"
}

// GetURL ports DashRule constructor setUrl.
func (r *DashRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/grammatik-leerzeichen/#fehler-1-leerzeichen-vor-und-nach-satzzeichen"
}

func (r *DashRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *DashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	var prevToken string
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
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
				// Java: StringUtils.equalsAny(token, "UND", "ODER", "BZW") — exact surface, not case-fold invent.
				if token != "UND" && token != "ODER" && token != "BZW" {
					msg := "Möglicherweise fehlt ein 'und' oder ein Komma, oder es wurde nach dem Wort " +
						"ein überflüssiges Leerzeichen eingefügt. Eventuell haben Sie auch versehentlich einen Bindestrich statt eines Punktes eingefügt."
					shortMsg := "Fehlendes 'und' oder Komma oder überflüssiges Leerzeichen?"
					fromPos := tokens[i-1].GetStartPos()
					toPos := tokens[i].GetEndPos()
					rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
					rm.ShortMessage = shortMsg
					joined := tokens[i-1].GetToken() + tokens[i].GetToken()
					// Java: first add joined, then optionally comma form when hyphen count ≤ 1.
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
