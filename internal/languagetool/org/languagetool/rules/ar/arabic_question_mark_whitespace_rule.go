package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ArabicQuestionMarkWhitespaceRule ports org.languagetool.rules.ar.ArabicQuestionMarkWhitespaceRule.
//
// Java:
//
//	public class ArabicQuestionMarkWhitespaceRule extends CommaWhitespaceRule {
//	  public ArabicQuestionMarkWhitespaceRule(ResourceBundle messages) {
//	    super(messages,
//	      Example.wrong("نعم<marker> ؟</marker>"),
//	      Example.fixed("نعم<marker>؟</marker>"));
//	  }
//	  @Override public final String getId() { return "ARABIC_QM_WHITESPACE"; }
//	  @Override public String getCommaCharacter() { return "؟"; }
//	}
//
// Match body is inherited from CommaWhitespaceRule (final in Java). Java does not
// call setDefaultOff(). Examples come from super(messages, wrong, fixed).
type ArabicQuestionMarkWhitespaceRule struct {
	*rules.CommaWhitespaceRule
	// incorrectExamples / correctExamples port Rule.addExamplePair via
	// CommaWhitespaceRule(messages, IncorrectExample, CorrectExample).
	// Base CommaWhitespaceRule has no example storage; store on the leaf like AR semicolon.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewArabicQuestionMarkWhitespaceRule ports ArabicQuestionMarkWhitespaceRule(ResourceBundle).
func NewArabicQuestionMarkWhitespaceRule(messages map[string]string) *ArabicQuestionMarkWhitespaceRule {
	base := rules.NewCommaWhitespaceRule(messages)
	// Java @Override getId
	base.RuleID = "ARABIC_QM_WHITESPACE"
	// Java @Override getCommaCharacter → "؟" (U+061F)
	base.CommaCharacter = "؟"
	// Java does not call setDefaultOff().
	base.DefaultOff = false
	r := &ArabicQuestionMarkWhitespaceRule{CommaWhitespaceRule: base}
	// Java: super(messages, Example.wrong(...), Example.fixed(...)) → addExamplePair
	r.AddExamplePair(
		rules.Wrong("نعم<marker> ؟</marker>"),
		rules.Fixed("نعم<marker>؟</marker>"),
	)
	return r
}

// GetID ports ArabicQuestionMarkWhitespaceRule.getId.
func (r *ArabicQuestionMarkWhitespaceRule) GetID() string {
	return "ARABIC_QM_WHITESPACE"
}

// GetCommaCharacter ports ArabicQuestionMarkWhitespaceRule.getCommaCharacter.
func (r *ArabicQuestionMarkWhitespaceRule) GetCommaCharacter() string {
	return "؟"
}

// IsDefaultOff ports Rule.isDefaultOff — Java never setDefaultOff on this type.
func (r *ArabicQuestionMarkWhitespaceRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// AddExamplePair ports Rule.addExamplePair (via CommaWhitespaceRule super path).
func (r *ArabicQuestionMarkWhitespaceRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ArabicQuestionMarkWhitespaceRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ArabicQuestionMarkWhitespaceRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// Match ports inherited CommaWhitespaceRule.match (final in Java) with RuleMatch.rule = this.
func (r *ArabicQuestionMarkWhitespaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms := r.CommaWhitespaceRule.Match(sentence)
	for _, m := range ms {
		if m != nil {
			// Java new RuleMatch(this, ...) — this is the Arabic subclass instance.
			m.Rule = r
		}
	}
	return ms
}
