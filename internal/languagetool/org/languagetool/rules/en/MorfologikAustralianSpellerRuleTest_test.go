package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withAU(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikAustralianSpellerRule()
	sp := morfologik.NewMorfologikSpeller(AustralianSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.Multi = nil // map inject: disable multi-speller (Java tests use single inject)
	r.Speller = sp
	r.IsMisspelled = r.MorfologikSpellerRule.IsMisspelled
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	return r
}

// Twin of MorfologikAustralianSpellerRuleTest.testSuggestions
func TestMorfologikAustralianSpellerRule_Suggestions(t *testing.T) {
	r := withAU("the", "speech")
	require.Equal(t, MorfologikAustralianSpellerRuleID, r.GetID())
	tops := EnglishAdditionalTopSuggestions("speach", r.IsMisspelled)
	require.Equal(t, []string{"speech"}, tops)
}

// Twin of MorfologikAustralianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikAustralianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := withAU(
		"This", "is", "an", "example", "we", "get", "behaviour", "as", "a", "dictionary", "word",
		"Why", "don", "t", "speak", "today", "He", "doesn", "know", "what", "to", "do",
		"My", "fiancée", "I", "like", "my", "emoji", "Dinkum", "taught", "He", "us",
	)
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"My fiancée",
		"He doesn't know what to do.",
		",", "123454", "I like my emoji 😍!", "μ",
		"Dinkum",
	} {
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}

	// behavior → behaviour
	if r.OtherVariant == nil {
		r.OtherVariant = map[string]string{}
	}
	if _, ok := r.OtherVariant["behavior"]; !ok {
		r.OtherVariant["behavior"] = "behaviour"
	}
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	ms, err := r.Match(languagetool.AnalyzePlain("behavior"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 8, ms[0].GetToPos())
	require.Equal(t, "behaviour", ms[0].GetSuggestedReplacements()[0])

	ms, err = r.Match(languagetool.AnalyzePlain("aõh"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	ms, err = r.Match(languagetool.AnalyzePlain("a"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// teached → taught
	r.Synthesize = func(surface, lemma, pos string) []string {
		if lemma == "teach" && pos == "VBD" {
			return []string{"taught"}
		}
		return nil
	}
	ms, err = r.Match(languagetool.AnalyzePlain("He teached us."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 10, ms[0].GetToPos())
	require.Equal(t, "taught", ms[0].GetSuggestedReplacements()[0])
}
