package da

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// DanishHunspellClasspath ports Java HunspellRule.getDictFilenameInResources for da_DK.
const DanishHunspellClasspath = "/da/hunspell/da_DK.dic"

// NewDanishHunspellRule ports Danish Language createDefaultSpellingRule /
// getRelevantRules registration of org.languagetool.rules.spelling.hunspell.HunspellRule
// (id HUNSPELL_RULE — not an invent Morfologik wrapper).
// Opens official da_DK.dic when present; nil dict fails closed.
func NewDanishHunspellRule() *hunspell.HunspellRule {
	return hunspell.NewHunspellRule("da", hunspell.TryOpenFromClasspath(DanishHunspellClasspath))
}
