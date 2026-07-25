package de

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// domains ports Java GermanCommaWhitespaceRule.DOMAINS = Pattern.compile("[a-z]{2,10}-Domains?").
// Anchors mirror Java Matcher.matches() (full-string match).
var domains = regexp.MustCompile(`^[a-z]{2,10}-Domains?$`)

// GermanCommaWhitespaceRule ports org.languagetool.rules.de.GermanCommaWhitespaceRule.
// Extends CommaWhitespaceRule; only override is isException for ". de-Domains" spacing.
type GermanCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
	url               string
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewGermanCommaWhitespaceRule ports
// GermanCommaWhitespaceRule(ResourceBundle, IncorrectExample, CorrectExample, URL).
// Java super(messages, incorrectExample, correctExample, url): category/issue type from
// CommaWhitespaceRule; addExamplePair when both examples non-null; setUrl when url non-null.
func NewGermanCommaWhitespaceRule(
	messages map[string]string,
	incorrect rules.IncorrectExample,
	correct rules.CorrectExample,
	url string,
) *GermanCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	// Java GermanCommaWhitespaceRule.isException:
	// tokenIdx+2 < tokens.length && tokens[tokenIdx] == "." && DOMAINS.matches(next).
	base.IsException = func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool {
		if tokenIdx+2 < len(tokens) &&
			tokens[tokenIdx].GetToken() == "." &&
			domains.MatchString(tokens[tokenIdx+1].GetToken()) {
			return true
		}
		return false
	}
	r := &GermanCommaWhitespaceRule{
		CommaWhitespaceRule: base,
		url:                 url,
	}
	// Java: if (incorrectExample != null && correctExample != null) addExamplePair(...)
	// Go has no null values; empty Example string is still a non-null object — only skip
	// when both sides are the zero IncorrectExample/CorrectExample (no example text).
	if incorrect.GetExample() != "" || correct.GetExample() != "" {
		r.AddExamplePair(incorrect, correct)
	}
	return r
}

// GetURL ports Rule.getUrl (set via constructor from German.getRelevantRules / tests).
func (r *GermanCommaWhitespaceRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.url
}

// AddExamplePair ports Rule.addExamplePair (via CommaWhitespaceRule super path).
func (r *GermanCommaWhitespaceRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *GermanCommaWhitespaceRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *GermanCommaWhitespaceRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports CommaWhitespaceRule.match with RuleMatch.rule = this (Java new RuleMatch(this, ...)).
func (r *GermanCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			m.Rule = r
		}
	}
	return ms
}
