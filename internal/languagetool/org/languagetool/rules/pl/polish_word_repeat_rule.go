package pl

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PolishWordRepeatRule ports org.languagetool.rules.pl.PolishWordRepeatRule
// (extends AdvancedWordRepeatRule).
//
// Java static EXC_WORDS / EXC_POS / EXC_NONWORDS are applied via
// AdvancedWordRepeatRule.ExcludedWords / ExcludedPos / ExcludedNonWords.
// Prepositions/pronouns are excluded via ExcludedPos (prep:.*|ppron.*) only when
// tagged — no surface invent of prep lists (Java).
type PolishWordRepeatRule struct {
	*rules.AdvancedWordRepeatRule
}

// Java PolishWordRepeatRule static EXC_WORDS (lemma set).
var polishWordRepeatExcWords = func() map[string]bool {
	// Java static block order (HashSet → unmodifiableSet):
	words := []string{
		"nie", "tuż", "aż", "to", "siebie", "być", "ani", "ni", "albo", "lub", "czy",
		"bądź", "jako", "zł", "np", "coraz", "bardzo", "bardziej", "proc", "ten", "jak",
		"mln", "tys", "swój", "mój", "twój", "nasz", "wasz", "i", "zbyt", "się",
	}
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}()

// Java: Pattern.compile("prep:.*|ppron.*")
var polishWordRepeatExcPos = regexp.MustCompile(`prep:.*|ppron.*`)

// Java:
// Pattern.compile("&quot|&gt|&lt|&amp|[0-9].*|" +
//   "M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$")
var polishWordRepeatExcNonWords = regexp.MustCompile(
	`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`,
)

// NewPolishWordRepeatRule ports PolishWordRepeatRule(ResourceBundle messages).
// Java: super(messages); addExamplePair(...).
func NewPolishWordRepeatRule(messages map[string]string) *PolishWordRepeatRule {
	base := &rules.AdvancedWordRepeatRule{
		ExcludedWords:    polishWordRepeatExcWords,
		ExcludedNonWords: polishWordRepeatExcNonWords,
		ExcludedPos:      polishWordRepeatExcPos,
		ID:               "PL_WORD_REPEAT",
		// Java getMessage / getShortMessage
		Message:      "Powtórzony wyraz w zdaniu",
		ShortMessage: "Powtórzenie wyrazu",
	}
	// Java AdvancedWordRepeatRule ctor: MISC, defaultOff, Style
	rules.InitAdvancedWordRepeatMeta(base, messages)
	// Java:
	// addExamplePair(Example.wrong("Mówiła długo, bo lubiła robić wszystko <marker>długo</marker>."),
	//                Example.fixed("Mówiła długo, bo lubiła robić wszystko <marker>powoli</marker>."));
	base.AddExamplePair(
		rules.Wrong("Mówiła długo, bo lubiła robić wszystko <marker>długo</marker>."),
		rules.Fixed("Mówiła długo, bo lubiła robić wszystko <marker>powoli</marker>."),
	)
	return &PolishWordRepeatRule{AdvancedWordRepeatRule: base}
}

// GetID ports getId() → "PL_WORD_REPEAT".
func (r *PolishWordRepeatRule) GetID() string {
	return "PL_WORD_REPEAT"
}

// GetDescription ports getDescription().
func (r *PolishWordRepeatRule) GetDescription() string {
	return "Powtórzenia wyrazów w zdaniu (monotonia stylistyczna)"
}
