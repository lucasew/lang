package ru

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianWordRepeatRule ports org.languagetool.rules.ru.RussianWordRepeatRule.
type RussianWordRepeatRule struct {
	*rules.AdvancedWordRepeatRule
}

func NewRussianWordRepeatRule(messages map[string]string) *RussianWordRepeatRule {
	exc := map[string]bool{}
	for _, w := range []string{
		"не", "ни", "а", "их", "на", "в", "по", "минута", "друг", "час", "секунда",
		"ПАО", "ООО", "табл", "рис",
		// surface prepositions/particles (Java EXC_POS PREP|PARTICLE|…)
		"и", "с", "к", "у", "о", "об", "от", "до", "за", "из", "под", "над", "при", "для", "без",
	} {
		exc[w] = true
	}
	base := &rules.AdvancedWordRepeatRule{
		Messages:           messages,
		ExcludedWords:      exc,
		ExcludedNonWords:   regexp.MustCompile(`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`),
		ExcludedPos:        regexp.MustCompile(`INTERJECTION|PRDC|PREP|CONJ|PARTICLE|ABR|NumC:.*|Num:.*`),
		ID:                 "RU_WORD_REPEAT",
		Message:            "Повтор слов в предложении",
		ShortMessage:       "Повтор слов в предложении",
		AlsoExcludeSurface: true,
	}
	return &RussianWordRepeatRule{AdvancedWordRepeatRule: base}
}
