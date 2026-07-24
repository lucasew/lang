package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withGB(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikBritishSpellerRule()
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.ClearMultiSpellers() // map inject: disable multi-speller (Java tests use single inject)
	r.Speller = sp
	r.IsMisspelled = r.MorfologikSpellerRule.IsMisspelled
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	return r
}

// Twin of MorfologikBritishSpellerRuleTest.testSuggestions + AbstractEnglish non-variant tops.
func TestMorfologikBritishSpellerRule_Suggestions(t *testing.T) {
	r := withGB("the", "separate", "definitely", "receive", "speech", "official")
	require.Equal(t, MorfologikBritishSpellerRuleID, r.GetID())
	// Curated EN tops present in Java getAdditionalTopSuggestions maps
	for _, tc := range []struct{ bad, good string }{
		{"speach", "speech"},
		{"alot", "a lot"},
		{"ur", "your"},
		{"slimiar", "similar"},
	} {
		tops := EnglishAdditionalTopSuggestions(tc.bad, r.IsMisspelled)
		require.NotEmpty(t, tops, tc.bad)
		require.Equal(t, tc.good, tops[0], tc.bad)
		ms, err := r.Match(languagetool.AnalyzePlain(tc.bad))
		require.NoError(t, err, tc.bad)
		require.NotEmpty(t, ms, tc.bad)
		require.Equal(t, tc.good, ms[0].GetSuggestedReplacements()[0], tc.bad)
	}
	// Dict-distance typos (Java Morfologik); inject as Speller.Suggestions for Match order
	r.Speller.AddWord("the")
	r.Speller.Suggestions["teh"] = []string{"the"}
	ms, err := r.Match(languagetool.AnalyzePlain("teh"))
	require.NoError(t, err)
	require.Equal(t, "the", ms[0].GetSuggestedReplacements()[0])
}

// Twin of MorfologikBritishSpellerRuleTest.testVariantMessages
func TestMorfologikBritishSpellerRule_VariantMessages(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	if r.OtherVariant == nil {
		r.OtherVariant = map[string]string{}
	}
	if _, ok := r.OtherVariant["color"]; !ok {
		r.OtherVariant["color"] = "colour"
	}
	r.OtherVariantName = "American English"
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant

	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	for _, w := range []string{"This", "is", "a", "nice", "the", "American", "English", "word"} {
		sp.AddWord(w)
	}
	r.ClearMultiSpellers() // map inject: disable multi-speller (Java tests use single inject)
	r.Speller = sp
	r.IsMisspelled = r.MorfologikSpellerRule.IsMisspelled

	ms, err := r.Match(languagetool.AnalyzePlain("This is a nice color."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "is American English")
	require.Equal(t, "colour", ms[0].GetSuggestedReplacements()[0])

	ms2, err := r.Match(languagetool.AnalyzePlain("Color is the American English word."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetMessage(), "is American English")
	require.Equal(t, "Colour", ms2[0].GetSuggestedReplacements()[0])
}

// Twin of MorfologikBritishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikBritishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := withGB(
		"This", "is", "an", "example", "we", "get", "behaviour", "as", "a", "dictionary", "word",
		"Why", "don", "t", "speak", "today", "He", "doesn", "know", "what", "to", "do",
		"The", "entrée", "at", "the", "café", "my", "Ph", "D", "thesis",
		"Ménage", "ménage", "trois", "quid", "pro", "quo",
		"Ma", "am", "O", "Connell", "Connor", "Neill",
		"going", "taught", "Behaviour", "He", "us",
	)
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"He doesn't know what to do.",
		"The entrée at the café.",
		"This is my Ph.D. thesis.",
		",", "123454", "μ",
		"Ménage à trois", "ménage à trois", "The quid pro quo",
		"Ma'am, O'Connell, O’Connell, O'Connor, O’Neill",
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
	ms, err := r.Match(languagetool.AnalyzePlain("Behavior"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 8, ms[0].GetToPos())
	require.Equal(t, "Behaviour", ms[0].GetSuggestedReplacements()[0])

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

	ms, err = r.Match(languagetool.AnalyzePlain("I'm g oing"))
	require.NoError(t, err)
	m := firstENSuggestion(ms, "going")
	require.NotNil(t, m)
	require.Equal(t, 4, m.GetFromPos())
	require.Equal(t, 10, m.GetToPos())
	require.Equal(t, "going", m.GetSuggestedReplacements()[0])

	// Java custom URL for archeological — pattern from enVariantBlogPatterns
	require.Equal(t,
		"https://languagetool.org/insights/post/our-or/#likeable-vs-likable-judgement-vs-judgment-oestrogen-vs-estrogen",
		enVariantBlogURL("archeological"))
}
