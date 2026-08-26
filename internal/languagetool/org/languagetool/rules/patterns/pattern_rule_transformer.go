package patterns

// TransformedRules ports PatternRuleTransformer.TransformedRules.
type TransformedRules struct {
	RemainingRules   []*AbstractPatternRule
	TransformedRules []any // typically []rules.Rule implementations
}

func NewTransformedRules(remaining []*AbstractPatternRule, transformed []any) TransformedRules {
	return TransformedRules{
		RemainingRules:   append([]*AbstractPatternRule(nil), remaining...),
		TransformedRules: append([]any(nil), transformed...),
	}
}

func (t TransformedRules) GetRemainingRules() []*AbstractPatternRule { return t.RemainingRules }
func (t TransformedRules) GetTransformedRules() []any                { return t.TransformedRules }

// PatternRuleTransformer ports org.languagetool.rules.patterns.PatternRuleTransformer.
type PatternRuleTransformer interface {
	Transform(rules []*AbstractPatternRule) TransformedRules
}

// PatternRuleTransformerFunc adapts a function.
type PatternRuleTransformerFunc func(rules []*AbstractPatternRule) TransformedRules

func (f PatternRuleTransformerFunc) Transform(rules []*AbstractPatternRule) TransformedRules {
	return f(rules)
}

// TransformPatternRules applies transformers sequentially (ports JLanguageTool.transformPatternRules shape).
func TransformPatternRules(rules []*AbstractPatternRule, transformers ...PatternRuleTransformer) ([]*AbstractPatternRule, []any) {
	remaining := append([]*AbstractPatternRule(nil), rules...)
	var allTransformed []any
	for _, tr := range transformers {
		if tr == nil {
			continue
		}
		out := tr.Transform(remaining)
		remaining = out.RemainingRules
		allTransformed = append(allTransformed, out.TransformedRules...)
	}
	return remaining, allTransformed
}
