package bitext

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// SelectBitextRules ports Tools.selectBitextRules for []BitextRule.
// See tools.SelectBitextRules for Java control flow (including multi-enable quirk).
func SelectBitextRules(bRules []BitextRule, disabledRules, enabledRules []string, useEnabledOnly bool) []BitextRule {
	return tools.SelectBitextRules(bRules, disabledRules, enabledRules, useEnabledOnly)
}
