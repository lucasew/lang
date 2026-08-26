package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianWordRepeatRule ports org.languagetool.rules.ru.RussianWordRepeatRule
// (extends AdvancedWordRepeatRule).
//
// Java static EXC_WORDS / EXC_POS / EXC_NONWORDS are applied via
// AdvancedWordRepeatRule.ExcludedWords / ExcludedPos / ExcludedNonWords.
type RussianWordRepeatRule struct {
	*rules.AdvancedWordRepeatRule
}

// Java RussianWordRepeatRule static EXC_WORDS (lemma set).
var russianWordRepeatExcWords = func() map[string]bool {
	// Java:
	// tempSet.add("не"); tempSet.add("ни"); tempSet.add("а"); tempSet.add("их");
	// tempSet.add("на"); tempSet.add("в"); tempSet.add("по"); tempSet.add("минута");
	// tempSet.add("друг"); tempSet.add("час"); tempSet.add("секунда"); tempSet.add("ПАО");
	// tempSet.add("ООО"); tempSet.add("табл"); tempSet.add("рис");
	words := []string{
		"не", "ни", "а", "их", "на", "в", "по", "минута", "друг", "час", "секунда",
		"ПАО", "ООО", "табл", "рис",
	}
	m := make(map[string]bool, len(words))
	for _, w := range words {
		m[w] = true
	}
	return m
}()

// Java: Pattern.compile("INTERJECTION|PRDC|PREP|CONJ|PARTICLE|ABR|NumC:.*|Num:.*")
var russianWordRepeatExcPos = regexp.MustCompile(`INTERJECTION|PRDC|PREP|CONJ|PARTICLE|ABR|NumC:.*|Num:.*`)

// Java:
// Pattern.compile("&quot|&gt|&lt|&amp|[0-9].*|" +
//   "M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$")
var russianWordRepeatExcNonWords = regexp.MustCompile(
	`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`,
)

// NewRussianWordRepeatRule ports RussianWordRepeatRule(ResourceBundle messages).
// Java: super(messages); addExamplePair(...); // setDefaultTempOff() commented out.
func NewRussianWordRepeatRule(messages map[string]string) *RussianWordRepeatRule {
	base := &rules.AdvancedWordRepeatRule{
		ExcludedWords:    russianWordRepeatExcWords,
		ExcludedNonWords: russianWordRepeatExcNonWords,
		ExcludedPos:      russianWordRepeatExcPos,
		ID:               "RU_WORD_REPEAT",
		// Java getMessage / getShortMessage
		Message:      "Повтор слов в предложении",
		ShortMessage: "Повтор слов в предложении",
	}
	// Java AdvancedWordRepeatRule ctor: MISC, defaultOff, Style
	rules.InitAdvancedWordRepeatMeta(base, messages)
	// Java:
	// addExamplePair(Example.wrong("Всё смешалось в <marker>доме доме</marker> Облонских."),
	//                Example.fixed("Всё смешалось в <marker>доме</marker> Облонских."));
	base.AddExamplePair(
		rules.Wrong("Всё смешалось в <marker>доме доме</marker> Облонских."),
		rules.Fixed("Всё смешалось в <marker>доме</marker> Облонских."),
	)
	return &RussianWordRepeatRule{AdvancedWordRepeatRule: base}
}

// GetID ports getId() → "RU_WORD_REPEAT" (also set on AdvancedWordRepeatRule.ID).
func (r *RussianWordRepeatRule) GetID() string {
	return "RU_WORD_REPEAT"
}

// GetDescription ports getDescription().
func (r *RussianWordRepeatRule) GetDescription() string {
	return "Повтор слов в предложении"
}
