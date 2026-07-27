package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicSemiColonWhitespaceRule ports org.languagetool.rules.ar.ArabicSemiColonWhitespaceRule.
//
// Java:
//
//	public class ArabicSemiColonWhitespaceRule extends CommaWhitespaceRule {
//	  public ArabicSemiColonWhitespaceRule(ResourceBundle messages) {
//	    super(messages,
//	      Example.wrong("نعم<marker> ؟</marker>"),
//	      Example.fixed("نعم<marker>؟</marker>"));
//	  }
//	  @Override public final String getId() { return "ARABIC_SC_WHITESPACE"; }
//	  @Override public String getCommaCharacter() { return "؛"; }
//	}
//
// Match body is inherited from CommaWhitespaceRule (final in Java). Java does not
// call setDefaultOff(). Examples come from super(messages, wrong, fixed).
// NOTE: Java example strings use Arabic QUESTION MARK "؟" not semicolon "؛" —
// port that quirk; do not "fix".
type ArabicSemiColonWhitespaceRule struct {
	*rules.CommaWhitespaceRule
	// incorrectExamples / correctExamples port Rule.addExamplePair via
	// CommaWhitespaceRule(messages, IncorrectExample, CorrectExample).
	// Base CommaWhitespaceRule has no example storage; store on the leaf like AR comma.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewArabicSemiColonWhitespaceRule ports ArabicSemiColonWhitespaceRule(ResourceBundle).
func NewArabicSemiColonWhitespaceRule(messages map[string]string) *ArabicSemiColonWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	// Java @Override getId
	base.RuleID = "ARABIC_SC_WHITESPACE"
	// Java @Override getCommaCharacter → "؛" (U+061B)
	base.CommaCharacter = "؛"
	// Java does not call setDefaultOff().
	base.DefaultOff = false
	r := &ArabicSemiColonWhitespaceRule{CommaWhitespaceRule: base}
	// Java: super(messages, Example.wrong(...), Example.fixed(...)) → addExamplePair
	// Quirk: markers use Arabic question mark "؟", not the semicolon character.
	r.AddExamplePair(
		rules.Wrong("نعم<marker> ؟</marker>"),
		rules.Fixed("نعم<marker>؟</marker>"),
	)
	return r
}

// GetID ports ArabicSemiColonWhitespaceRule.getId.
func (r *ArabicSemiColonWhitespaceRule) GetID() string {
	return "ARABIC_SC_WHITESPACE"
}

// GetCommaCharacter ports ArabicSemiColonWhitespaceRule.getCommaCharacter.
func (r *ArabicSemiColonWhitespaceRule) GetCommaCharacter() string {
	return "؛"
}

// IsDefaultOff ports Rule.isDefaultOff — Java never setDefaultOff on this type.
func (r *ArabicSemiColonWhitespaceRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// AddExamplePair ports Rule.addExamplePair (via CommaWhitespaceRule super path).
func (r *ArabicSemiColonWhitespaceRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ArabicSemiColonWhitespaceRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ArabicSemiColonWhitespaceRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports inherited CommaWhitespaceRule.match (final in Java) with RuleMatch.rule = this.
func (r *ArabicSemiColonWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Arabic subclass instance.
			m.Rule = r
		}
	}
	return ms
}
