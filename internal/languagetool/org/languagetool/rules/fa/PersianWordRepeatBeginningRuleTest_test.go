package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPersianWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewPersianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "adverb start",
		"desc_repetition_beginning_word":      "word start",
		"desc_repetition_beginning_thesaurus": "thesaurus",
	})
	// two successive همچنین
	matches := rule.MatchList(languagetool.SplitAndAnalyze("همچنین، خیابان مسکونی است. همچنین، به افتخار یک شاعر نامگذاری شده‌است."))
	require.Equal(t, 1, len(matches))
	// two non-adverb same start — no error
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("این خوب است. این بهتر است."))))
}
