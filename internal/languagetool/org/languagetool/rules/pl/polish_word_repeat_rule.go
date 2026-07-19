package pl

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PolishWordRepeatRule ports org.languagetool.rules.pl.PolishWordRepeatRule.
// Prepositions/pronouns are excluded via ExcludedPos (prep:.*|ppron.*) only when tagged —
// no surface invent of prep lists (Java).
type PolishWordRepeatRule struct {
	*rules.AdvancedWordRepeatRule
}

func NewPolishWordRepeatRule(messages map[string]string) *PolishWordRepeatRule {
	// Java EXC_WORDS only (lemma match when tagged).
	exc := map[string]bool{}
	for _, w := range []string{
		"nie", "tuż", "aż", "to", "siebie", "być", "ani", "ni", "albo", "lub", "czy",
		"bądź", "jako", "zł", "np", "coraz", "bardzo", "bardziej", "proc", "ten", "jak",
		"mln", "tys", "swój", "mój", "twój", "nasz", "wasz", "i", "zbyt", "się",
	} {
		exc[w] = true
	}
	base := &rules.AdvancedWordRepeatRule{
		ExcludedWords:    exc,
		ExcludedNonWords: regexp.MustCompile(`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`),
		ExcludedPos:      regexp.MustCompile(`prep:.*|ppron.*`),
		ID:               "PL_WORD_REPEAT",
		Message:          "Powtórzony wyraz w zdaniu",
		ShortMessage:     "Powtórzenie wyrazu",
	}
	rules.InitAdvancedWordRepeatMeta(base, messages)
	return &PolishWordRepeatRule{AdvancedWordRepeatRule: base}
}
