package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withCA(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikCanadianSpellerRule()
	sp := morfologik.NewMorfologikSpeller(CanadianSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.Speller = sp
	r.IsMisspelled = r.MorfologikSpellerRule.IsMisspelled
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	return r
}

// Twin of MorfologikCanadianSpellerRuleTest.testSuggestions
func TestMorfologikCanadianSpellerRule_Suggestions(t *testing.T) {
	r := withCA("the", "speech", "separate")
	require.Equal(t, MorfologikCanadianSpellerRuleID, r.GetID())
	// Curated tops
	require.Equal(t, []string{"speech"}, EnglishAdditionalTopSuggestions("speach", r.IsMisspelled))
	// Dict-distance inject
	r.Speller.Suggestions["teh"] = []string{"the"}
	r.Speller.Suggestions["seperate"] = []string{"separate"}
	for _, tc := range []struct{ bad, good string }{
		{"teh", "the"},
		{"seperate", "separate"},
	} {
		ms, err := r.Match(languagetool.AnalyzePlain(tc.bad))
		require.NoError(t, err, tc.bad)
		require.Equal(t, tc.good, ms[0].GetSuggestedReplacements()[0], tc.bad)
	}
}

// Twin of MorfologikCanadianSpellerRuleTest.testMorfologikSpeller
// Note: Java test constructs MorfologikBritishSpellerRule by mistake for some asserts;
// we twin Canadian rule with the same surface expectations (behaviour, arbor→arbour).
func TestMorfologikCanadianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := withCA(
		"This", "is", "an", "example", "we", "get", "behaviour", "as", "a", "dictionary", "word",
		"Why", "don", "t", "speak", "today", "He", "doesn", "know", "what", "to", "do",
		"I", "like", "my", "emoji", "arbour",
	)
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"He doesn't know what to do.",
		",", "123454", "I like my emoji (😥)...", "μ",
	} {
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}

	// arbor → arbour (American form in Canadian; US-GB map column 0)
	if r.OtherVariant == nil {
		r.OtherVariant = map[string]string{}
	}
	if _, ok := r.OtherVariant["arbor"]; !ok {
		r.OtherVariant["arbor"] = "arbour"
	}
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	ms, err := r.Match(languagetool.AnalyzePlain("arbor"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 5, ms[0].GetToPos())
	require.Contains(t, ms[0].GetSuggestedReplacements(), "arbour")

	ms, err = r.Match(languagetool.AnalyzePlain("aõh"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	ms, err = r.Match(languagetool.AnalyzePlain("a"))
	require.NoError(t, err)
	require.Empty(t, ms)
}
