package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// ArabicConfusionProbabilityRule ports org.languagetool.rules.ar.ArabicConfusionProbabilityRule.
type ArabicConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
	// incorrectExamples / correctExamples port Rule.addExamplePair (not on ngrams package).
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewArabicConfusionProbabilityRule(lm ngrams.LanguageModel) *ArabicConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, 3)
	base.DefaultOff = false
	r := &ArabicConfusionProbabilityRule{ConfusionProbabilityRule: base}
	// Java demo (trailing unclosed <marker> kept as upstream)
	r.AddExamplePair(
		rules.Wrong("إن بعض <marker>الضن</marker> إثم.<marker>"),
		rules.Fixed("إن بعض <marker>الظن</marker> إثم.<marker>"),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *ArabicConfusionProbabilityRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ArabicConfusionProbabilityRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ArabicConfusionProbabilityRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}
