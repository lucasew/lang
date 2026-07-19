package is

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// IcelandicHunspellClasspath ports Java HunspellNoSuggestionRule path for is_IS.
const IcelandicHunspellClasspath = "/is/hunspell/is_IS.dic"

// NewIcelandicHunspellNoSuggestionRule ports Icelandic getRelevantRules registration of
// org.languagetool.rules.spelling.hunspell.HunspellNoSuggestionRule.
// Opens official is_IS.dic when present; nil dict fails closed.
func NewIcelandicHunspellNoSuggestionRule() *hunspell.HunspellNoSuggestionRule {
	return hunspell.NewHunspellNoSuggestionRule("is", hunspell.TryOpenFromClasspath(IcelandicHunspellClasspath))
}
