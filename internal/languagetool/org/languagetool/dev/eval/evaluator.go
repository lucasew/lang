package eval

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// Match is a minimal rule match for evaluation (from/to + first suggestion).
type Match struct {
	FromPos, ToPos        int
	SuggestedReplacements []string
	RuleID                string
}

// FromRuleMatch adapts rules.RuleMatch.
func FromRuleMatch(m *rules.RuleMatch) Match {
	if m == nil {
		return Match{}
	}
	return Match{
		FromPos:               m.FromPos,
		ToPos:                 m.ToPos,
		SuggestedReplacements: append([]string(nil), m.GetSuggestedReplacements()...),
		RuleID:                "",
	}
}

// Evaluator ports org.languagetool.dev.eval.Evaluator with plain text.
type Evaluator interface {
	Check(text string) ([]Match, error)
	Close() error
}

// FuncEvaluator adapts a function to Evaluator.
type FuncEvaluator struct {
	Fn func(text string) ([]Match, error)
}

func (e FuncEvaluator) Check(text string) ([]Match, error) {
	if e.Fn == nil {
		return nil, nil
	}
	return e.Fn(text)
}
func (e FuncEvaluator) Close() error { return nil }
