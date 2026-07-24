package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPersianWordRepeatRule_Rule(t *testing.T) {
	rule := NewPersianWordRepeatRule(map[string]string{"repetition": "تکرار"})
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("من لی لی را دیدم"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("لک لک یک پرنده است"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("این کار برای برای تو بود."))))
}
