package gl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// GalicianHunspellClasspath ports Java HunspellRule.getDictFilenameInResources for gl_ES.
const GalicianHunspellClasspath = "/gl/hunspell/gl_ES.dic"

// NewGalicianHunspellRule ports Galician Language createDefaultSpellingRule /
// getRelevantRules registration of org.languagetool.rules.spelling.hunspell.HunspellRule.
// Opens official gl_ES.dic when present; nil dict fails closed.
func NewGalicianHunspellRule() *hunspell.HunspellRule {
	return hunspell.NewHunspellRule("gl", hunspell.TryOpenFromClasspath(GalicianHunspellClasspath))
}
