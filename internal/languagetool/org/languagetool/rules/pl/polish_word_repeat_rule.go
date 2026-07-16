package pl

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PolishWordRepeatRule ports org.languagetool.rules.pl.PolishWordRepeatRule.
type PolishWordRepeatRule struct {
	*rules.AdvancedWordRepeatRule
}

func NewPolishWordRepeatRule(messages map[string]string) *PolishWordRepeatRule {
	exc := map[string]bool{}
	for _, w := range []string{
		"nie", "tuż", "aż", "to", "siebie", "być", "ani", "ni", "albo", "lub", "czy",
		"bądź", "jako", "zł", "np", "coraz", "bardzo", "bardziej", "proc", "ten", "jak",
		"mln", "tys", "swój", "mój", "twój", "nasz", "wasz", "i", "zbyt", "się",
		// surface prepositions (Java uses prep:.* POS)
		"na", "w", "z", "do", "od", "po", "o", "u", "dla", "bez", "za", "przy", "nad", "pod",
	} {
		exc[w] = true
	}
	base := &rules.AdvancedWordRepeatRule{
		Messages:           messages,
		ExcludedWords:      exc,
		ExcludedNonWords:   regexp.MustCompile(`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`),
		ExcludedPos:        regexp.MustCompile(`prep:.*|ppron.*`),
		ID:                 "PL_WORD_REPEAT",
		Message:            "Powtórzony wyraz w zdaniu",
		ShortMessage:       "Powtórzenie wyrazu",
		AlsoExcludeSurface: true,
	}
	return &PolishWordRepeatRule{AdvancedWordRepeatRule: base}
}
