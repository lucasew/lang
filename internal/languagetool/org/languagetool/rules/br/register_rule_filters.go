package br

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.br.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
}
