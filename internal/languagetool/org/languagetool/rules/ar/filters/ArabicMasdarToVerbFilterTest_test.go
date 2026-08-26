package filters

// Twin of ArabicMasdarToVerbFilterTest (Java king).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/stretchr/testify/require"
)

// TestArabicMasdarToVerbFilter_Filter twins ArabicMasdarToVerbFilterTest.testFilter.
// Java assertSuggestion: verb="يقوم", noun="بال"+word; tag with ArabicTagger; expect 1 suggestion exact.
func TestArabicMasdarToVerbFilter_Filter(t *testing.T) {
	assertMasdarSuggestion(t, "عمل", "يعمل")
	assertMasdarSuggestion(t, "إعمال", "يعمل")
	assertMasdarSuggestion(t, "سؤال", "يسأل")
	assertMasdarSuggestion(t, "أكل", "يأكل")
}

func assertMasdarSuggestion(t *testing.T, word, expectedSuggestion string) {
	t.Helper()
	ar_tag.EnsureDefaultArabicTagger()
	require.NotNil(t, ar_tag.DefaultArabicTagger)
	require.NotEmpty(t, ar_tag.ArabicPOSDictPath(), "arabic.dict required for Java twin")

	f := NewArabicMasdarToVerbFilter()
	require.NotNil(t, f.synthesizer, "ArabicSynthesizer.INSTANCE required for Java twin")

	word1 := "يقوم"
	word2 := "بال" + word
	args := map[string]string{
		"verb": word1,
		"noun": word2,
	}
	patternTokens := ar_tag.DefaultArabicTagger.Tag([]string{word1, word2})
	require.Len(t, patternTokens, 2)

	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	ruleMatch := f.AcceptRuleMatch(match, args, -1, patternTokens, nil)

	require.NotNil(t, ruleMatch)
	sugs := ruleMatch.GetSuggestedReplacements()
	require.Len(t, sugs, 1, "word=%s token=%s", word, word2)
	require.Equal(t, expectedSuggestion, sugs[0], "word=%s", word)
}

// TestArabicMasdarToVerbFilter_Rule twins ArabicMasdarToVerbFilterTest.testRule — loadFromPath not null.
func TestArabicMasdarToVerbFilter_Rule(t *testing.T) {
	m := LoadFromPath(ArabicMasdarVerbFileName)
	require.NotNil(t, m)
	require.NotEmpty(t, m, "Java loadWords(/ar/arabic_masdar_verb.txt) must resolve")
	// Known masdar→verb rows from official file (exact keys, no invent trim)
	require.Equal(t, []string{"عَمِلَ"}, m["عمل"])
	require.Equal(t, []string{"أَعْمَلَ"}, m["إعمال"])
	require.Equal(t, []string{"سَأَلَ"}, m["سؤال"])
	require.Equal(t, []string{"أَكَلَ"}, m["أكل"])
}
