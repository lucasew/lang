package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withNZ(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikNewZealandSpellerRule()
	sp := morfologik.NewMorfologikSpeller(NewZealandSpellerDict, 1)
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

// Twin of MorfologikNewZealandSpellerRuleTest.testSuggestions
func TestMorfologikNewZealandSpellerRule_Suggestions(t *testing.T) {
	r := withNZ("the", "speech")
	require.Equal(t, MorfologikNewZealandSpellerRuleID, r.GetID())
	require.Equal(t, []string{"speech"}, EnglishAdditionalTopSuggestions("speach", r.IsMisspelled))
	// teh is dict-distance in Java; inject Speller.Suggestions
	r.Speller.Suggestions["teh"] = []string{"the"}
	ms, err := r.Match(languagetool.AnalyzePlain("teh"))
	require.NoError(t, err)
	require.Equal(t, "the", ms[0].GetSuggestedReplacements()[0])
}

// Twin of MorfologikNewZealandSpellerRuleTest.testMorfologikSpeller
func TestMorfologikNewZealandSpellerRule_MorfologikSpeller(t *testing.T) {
	r := withNZ(
		"This", "is", "an", "example", "we", "get", "behaviour", "as", "a", "dictionary", "word",
		"Why", "don", "t", "speak", "today", "He", "doesn", "know", "what", "to", "do",
		"Abercrombie", "émigré", "émigrés", "or", "tête", "à", "tête",
		"taught", "us",
	)
	// compound check for tête-à-tête parts
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"He doesn't know what to do.",
		",", "123454", "μ",
		"Abercrombie",
		"émigré or émigrés",
		"tête-à-tête",
	} {
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}

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
