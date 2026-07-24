package sv

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// SwedishHunspellClasspath ports Java HunspellRule.getDictFilenameInResources for sv_SE.
const SwedishHunspellClasspath = "/sv/hunspell/sv_SE.dic"

// NewSwedishHunspellRule ports Swedish Language createDefaultSpellingRule /
// getRelevantRules registration of org.languagetool.rules.spelling.hunspell.HunspellRule
// (id HUNSPELL_RULE). Opens official sv_SE.dic when present; nil dict fails closed.
func NewSwedishHunspellRule() *hunspell.HunspellRule {
	return hunspell.NewHunspellRule("sv", hunspell.TryOpenFromClasspath(SwedishHunspellClasspath))
}
