package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ConsistentApostrophesRule ports org.languagetool.rules.en.ConsistentApostrophesRule.
// Check for consistent use of apostrophes (typographic (’) vs. typewriter (')).
// @since 5.3
type ConsistentApostrophesRule struct {
	Messages map[string]string
	// URL ports Rule.url (Java setUrl punctuation-guide apostrophe).
	URL string
	// DefaultTempOff ports Rule.setDefaultTempOff (Java TODO still off by default).
	DefaultTempOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewConsistentApostrophesRule ports ConsistentApostrophesRule(ResourceBundle messages).
func NewConsistentApostrophesRule(messages map[string]string) *ConsistentApostrophesRule {
	r := &ConsistentApostrophesRule{
		Messages:       messages,
		URL:            "https://languagetool.org/insights/post/punctuation-guide/#what-is-an-apostrophe",
		DefaultTempOff: true, // Java setDefaultTempOff()
	}
	// Java: addExamplePair(doesn’t → doesn't) with mixed apostrophe styles
	r.AddExamplePair(
		rules.Wrong("It's nice, but it <marker>doesn’t</marker> work."),
		rules.Fixed("It's nice, but it <marker>doesn't</marker> work."),
	)
	return r
}

// GetID ports ConsistentApostrophesRule.getId.
func (r *ConsistentApostrophesRule) GetID() string { return "EN_CONSISTENT_APOS" }

// GetDescription ports ConsistentApostrophesRule.getDescription.
func (r *ConsistentApostrophesRule) GetDescription() string {
	return "Checks if the two types of apostrophes (' and ’) are used consistently in a text."
}

// GetURL ports Rule.getUrl.
func (r *ConsistentApostrophesRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *ConsistentApostrophesRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// IsDefaultTempOff ports Rule.isDefaultTempOff.
func (r *ConsistentApostrophesRule) IsDefaultTempOff() bool {
	return r != nil && r.DefaultTempOff
}

// MinToCheckParagraph ports TextLevelRule.minToCheckParagraph (Java -1 full text).
func (r *ConsistentApostrophesRule) MinToCheckParagraph() int { return -1 }

// AddExamplePair ports Rule.addExamplePair.
func (r *ConsistentApostrophesRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ConsistentApostrophesRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ConsistentApostrophesRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *ConsistentApostrophesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	pos := 0
	hasTwoTypes := hasTwoApostropheTypes(sentences)
	if !hasTwoTypes {
		// Java: return toRuleMatchArray(matches) → empty array
		return matches
	}
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		for _, token := range tokens {
			var message, repl string
			// Java: token != null && token.getToken().contains("'") && !token.hasTypographicApostrophe()
			if token != nil && strings.Contains(token.GetToken(), "'") && !token.HasTypographicApostrophe() {
				message = "You used a typewriter-style apostrophe here, but a typographic apostrophe elsewhere in this text."
				repl = strings.ReplaceAll(token.GetToken(), "'", "’")
			} else if token != nil && strings.Contains(token.GetToken(), "'") && token.HasTypographicApostrophe() {
				message = "You used a typographic apostrophe here, but a typewriter-style apostrophe elsewhere in this text."
				// Java: repl = token.getToken() (surface already normalized to typewriter ')
				repl = token.GetToken()
			}
			if message != "" {
				msg := message + " Both are correct, but consider using the same type everywhere in your text."
				rm := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
				rm.SetSuggestedReplacement(repl)
				matches = append(matches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return matches
}

// hasTwoApostropheTypes ports ConsistentApostrophesRule.hasTwoApostropheTypes.
func hasTwoApostropheTypes(sentences []*languagetool.AnalyzedSentence) bool {
	hasTypewriterStyle := false
	hasTypographicStyle := false
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		for _, token := range tokens {
			if token != nil && strings.Contains(token.GetToken(), "'") && !token.HasTypographicApostrophe() {
				hasTypewriterStyle = true
			} else if token != nil && strings.Contains(token.GetToken(), "'") && token.HasTypographicApostrophe() {
				hasTypographicStyle = true
			}
			if hasTypewriterStyle && hasTypographicStyle {
				return true
			}
		}
	}
	return false
}
