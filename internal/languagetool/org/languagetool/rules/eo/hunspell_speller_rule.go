package eo

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// EsperantoHunspellClasspath ports Java HunspellRule path for eo (eo.dic).
const EsperantoHunspellClasspath = "/eo/hunspell/eo.dic"

// NewEsperantoHunspellRule ports Esperanto Language createDefaultSpellingRule /
// getRelevantRules registration of org.languagetool.rules.spelling.hunspell.HunspellRule.
// Opens official eo.dic when present; nil dict fails closed.
func NewEsperantoHunspellRule() *hunspell.HunspellRule {
	return hunspell.NewHunspellRule("eo", hunspell.TryOpenFromClasspath(EsperantoHunspellClasspath))
}
