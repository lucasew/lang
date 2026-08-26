package de

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// DashRule ports org.languagetool.rules.de.DashRule.
// Flags hyphen compounds with a space after the hyphen (e.g. "Diäten- Erhöhung").
// Java: COMPOUNDING category + setUrl (not AbstractDashRule; DE-specific).
type DashRule struct {
	Messages map[string]string
	Category *rules.Category
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewDashRule(messages map[string]string) *DashRule {
	r := &DashRule{
		Messages: messages,
		Category: rules.CatCompounding.GetCategory(messages),
	}
	// Java: Diäten- Erhöhung → Diäten-Erhöhung
	r.AddExamplePair(
		rules.Wrong("Bundestag beschließt <marker>Diäten- Erhöhung</marker>"),
		rules.Fixed("Bundestag beschließt <marker>Diäten-Erhöhung</marker>"),
	)
	return r
}

func (r *DashRule) GetID() string { return "DE_DASH" }

// GetDescription ports DashRule.getDescription.
func (r *DashRule) GetDescription() string {
	return "Keine Leerzeichen in Bindestrich-Komposita (wie z.B. in 'Diäten- Erhöhung')"
}

// GetURL ports DashRule constructor setUrl (Tools.getUrl).
func (r *DashRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/grammatik-leerzeichen/#fehler-1-leerzeichen-vor-und-nach-satzzeichen"
}

func (r *DashRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// AddExamplePair ports Rule.addExamplePair.
func (r *DashRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *DashRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *DashRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// dashRuleIsUpperFirst ports Java Character.isUpperCase(token.charAt(0)).
// Empty token would throw StringIndexOutOfBoundsException in Java; return false.
func dashRuleIsUpperFirst(token string) bool {
	u := utf16.Encode([]rune(token))
	if len(u) == 0 {
		return false
	}
	return unicode.IsUpper(rune(u[0]))
}

// dashRuleEqualsAny ports Apache StringUtils.equalsAny(token, values...).
func dashRuleEqualsAny(token string, values ...string) bool {
	for _, v := range values {
		if token == v {
			return true
		}
	}
	return false
}

// Match ports org.languagetool.rules.de.DashRule.match.
func (r *DashRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	// Java: String prevToken = null;
	var prevToken *string
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		token := tokens[i].GetToken()
		if prevToken != nil &&
			strings.HasSuffix(*prevToken, "-") &&
			*prevToken != "-" &&
			!strings.Contains(*prevToken, "--") &&
			!strings.Contains(*prevToken, "–-") { // first char is some special kind of dash, found in Wikipedia
			// Java: char firstChar = token.charAt(0); Character.isUpperCase(firstChar)
			if dashRuleIsUpperFirst(token) {
				// Java: StringUtils.equalsAny(token, "UND", "ODER", "BZW") — exact surface
				if !dashRuleEqualsAny(token, "UND", "ODER", "BZW") {
					msg := "Möglicherweise fehlt ein 'und' oder ein Komma, oder es wurde nach dem Wort " +
						"ein überflüssiges Leerzeichen eingefügt. Eventuell haben Sie auch versehentlich einen Bindestrich statt eines Punktes eingefügt."
					shortMsg := "Fehlendes 'und' oder Komma oder überflüssiges Leerzeichen?"
					fromPos := tokens[i-1].GetStartPos()
					toPos := tokens[i].GetEndPos()
					rm := rules.NewRuleMatch(r, sentence, fromPos, toPos, msg)
					rm.ShortMessage = shortMsg
					// Java: addSuggestedReplacement(prev + curr)
					joined := tokens[i-1].GetToken() + tokens[i].GetToken()
					// Java: StringUtils.countMatches(..., "-") + ... <= 1 → also comma form
					hyphens := strings.Count(tokens[i-1].GetToken(), "-") + strings.Count(tokens[i].GetToken(), "-")
					if hyphens <= 1 {
						rm.SetSuggestedReplacements([]string{
							joined,
							tokens[i-1].GetToken() + ", " + tokens[i].GetToken(),
						})
					} else {
						rm.SetSuggestedReplacements([]string{joined})
					}
					ruleMatches = append(ruleMatches, rm)
				}
			}
		}
		// Java: prevToken = token; (including empty SENT_START surface)
		cp := token
		prevToken = &cp
	}
	return ruleMatches
}
