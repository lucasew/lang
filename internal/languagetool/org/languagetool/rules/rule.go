package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Rule is the minimal interface for sentence-level language rules
// (subset of org.languagetool.rules.Rule).
type Rule interface {
	GetID() string
	GetDescription() string
	Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch
}

// RuleWithError is used when Match can fail (I/O spellers).
type RuleWithError interface {
	GetID() string
	Match(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error)
}

// BaseRule holds common metadata for concrete rules.
type BaseRule struct {
	ID          string
	Description string
	Category    *Category
	DefaultOff  bool
	// incorrectExamples / correctExamples port Rule lists (Java addExamplePair).
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

func (r *BaseRule) GetID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

func (r *BaseRule) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

func (r *BaseRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

func (r *BaseRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *BaseRule) SetCategory(c *Category) {
	if r != nil {
		r.Category = c
	}
}

// AddExamplePair ports Rule.addExamplePair: stores wrong/fixed demo sentences.
// When the fixed example has <marker>…</marker>, that span is recorded as the
// incorrect example's correction (Java IncorrectExample constructor).
func (r *BaseRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// SetExamplePair ports Rule.setExamplePair (clears then adds one pair).
func (r *BaseRule) SetExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	r.incorrectExamples = nil
	r.correctExamples = nil
	r.AddExamplePair(incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *BaseRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *BaseRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// appendExamplePair is the shared Java Rule.addExamplePair body.
func appendExamplePair(incorrects *[]IncorrectExample, corrects *[]CorrectExample, incorrect IncorrectExample, correct CorrectExample) {
	ex := correct.GetExample()
	start := strings.Index(ex, "<marker>")
	end := strings.Index(ex, "</marker>")
	if start != -1 && end != -1 && end > start {
		correction := ex[start+len("<marker>") : end]
		incorrect = NewIncorrectExample(incorrect.GetExample(), correction)
	}
	*incorrects = append(*incorrects, incorrect)
	*corrects = append(*corrects, correct)
}
