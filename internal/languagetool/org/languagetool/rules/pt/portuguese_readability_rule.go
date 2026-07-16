package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// PortugueseReadabilityRule ports org.languagetool.rules.pt.PortugueseReadabilityRule metadata.
type PortugueseReadabilityRule struct {
	*rules.ReadabilityRule
	TooEasyTest bool
}

func NewPortugueseReadabilityRule(tooEasy bool, level int) *PortugueseReadabilityRule {
	return &PortugueseReadabilityRule{
		ReadabilityRule: rules.NewReadabilityRule(tooEasy, level),
		TooEasyTest:     tooEasy,
	}
}

func (r *PortugueseReadabilityRule) GetID() string {
	if r.TooEasyTest {
		return "READABILITY_RULE_SIMPLE_PT"
	}
	return "READABILITY_RULE_DIFFICULT_PT"
}

func (r *PortugueseReadabilityRule) GetDescription() string {
	if r.TooEasyTest {
		return "Legibilidade: texto demasiado simples"
	}
	return "Legibilidade: texto demasiado complexo"
}
