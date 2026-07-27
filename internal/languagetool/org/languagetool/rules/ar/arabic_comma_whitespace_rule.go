package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicCommaWhitespaceRule ports org.languagetool.rules.ar.ArabicCommaWhitespaceRule.
//
// Java:
//
//	public class ArabicCommaWhitespaceRule extends CommaWhitespaceRule {
//	  public ArabicCommaWhitespaceRule(ResourceBundle messages) {
//	    super(messages,
//	      Example.wrong("نعم<marker> ،</marker> لقد نجحنا."),
//	      Example.fixed("نعم<marker>،</marker> لقد نجحنا."));
//	  }
//	  @Override public final String getId() { return "ARABIC_COMMA_PARENTHESIS_WHITESPACE"; }
//	  @Override public String getCommaCharacter() { return "،"; }
//	}
//
// Match body is inherited from CommaWhitespaceRule (final in Java). Unlike FA,
// Java does not call setDefaultOff(). Examples come from super(messages, wrong, fixed).
type ArabicCommaWhitespaceRule struct {
	*rules.CommaWhitespaceRule
	// incorrectExamples / correctExamples port Rule.addExamplePair via
	// CommaWhitespaceRule(messages, IncorrectExample, CorrectExample).
	// Base CommaWhitespaceRule has no example storage; store on the leaf like DE.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewArabicCommaWhitespaceRule ports ArabicCommaWhitespaceRule(ResourceBundle).
func NewArabicCommaWhitespaceRule(messages map[string]string) *ArabicCommaWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	// Java @Override getId
	base.RuleID = "ARABIC_COMMA_PARENTHESIS_WHITESPACE"
	// Java @Override getCommaCharacter → "،" (U+060C)
	base.CommaCharacter = "،"
	// Java does not call setDefaultOff() (unlike PersianCommaWhitespaceRule).
	base.DefaultOff = false
	r := &ArabicCommaWhitespaceRule{CommaWhitespaceRule: base}
	// Java: super(messages, Example.wrong(...), Example.fixed(...)) → addExamplePair
	r.AddExamplePair(
		rules.Wrong("نعم<marker> ،</marker> لقد نجحنا."),
		rules.Fixed("نعم<marker>،</marker> لقد نجحنا."),
	)
	return r
}

// GetID ports ArabicCommaWhitespaceRule.getId.
func (r *ArabicCommaWhitespaceRule) GetID() string {
	return "ARABIC_COMMA_PARENTHESIS_WHITESPACE"
}

// GetCommaCharacter ports ArabicCommaWhitespaceRule.getCommaCharacter.
func (r *ArabicCommaWhitespaceRule) GetCommaCharacter() string {
	return "،"
}

// IsDefaultOff ports Rule.isDefaultOff — Java never setDefaultOff on this type.
func (r *ArabicCommaWhitespaceRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// AddExamplePair ports Rule.addExamplePair (via CommaWhitespaceRule super path).
func (r *ArabicCommaWhitespaceRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ArabicCommaWhitespaceRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ArabicCommaWhitespaceRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports inherited CommaWhitespaceRule.match (final in Java) with RuleMatch.rule = this.
func (r *ArabicCommaWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Arabic subclass instance.
			m.Rule = r
		}
	}
	return ms
}
