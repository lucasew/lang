package ml

// Twin of org.languagetool.rules.ml.MorfologikMalayalamSpellerRuleTest
// Official resource: /ml/hunspell/ml_IN.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
//
// Java setup:
//   rule = new MorfologikMalayalamSpellerRule(TestTools.getMessages("ml"), new Malayalam(), null, emptyList());
//   lt   = new JLanguageTool(new Malayalam()); // MalayalamWordTokenizer
// Match positions are UTF-16 (Java RuleMatch fromPos/toPos).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	mltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ml"
	"github.com/stretchr/testify/require"
)

// requireMalayalamSpellerDict skips when official ml_IN.dict is not discoverable (no invent).
func requireMalayalamSpellerDict(t *testing.T) *MorfologikMalayalamSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(MorfologikMalayalamSpellerRuleDict) == "" {
		t.Skip("ml_IN.dict not discoverable (inspiration/.../ml/hunspell)")
	}
	r := NewMorfologikMalayalamSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// analyzeML ports JLanguageTool(new Malayalam()).getAnalyzedSentence: MalayalamWordTokenizer.
func analyzeML(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, mltok.NewMalayalamWordTokenizer())
}

func matchML(t *testing.T, r *MorfologikMalayalamSpellerRule, text string) []*rules.RuleMatch {
	t.Helper()
	ms, err := r.Match(analyzeML(text))
	require.NoError(t, err)
	return ms
}

// Port of MorfologikMalayalamSpellerRuleTest.testMorfologikSpeller
func TestMorfologikMalayalamSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireMalayalamSpellerDict(t)
	require.Equal(t, MorfologikMalayalamSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikMalayalamSpellerRuleDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_ML_IN", r.GetID())
	require.Equal(t, "/ml/hunspell/ml_IN.dict", r.GetFileName())
	// Java pure super: isLatinScript default true → NonLatinScript false.
	if r.SpellingCheckRule != nil {
		require.False(t, r.NonLatinScript)
	}

	// correct sentences:
	// Java: "എന്തുകൊണ്ട്‌ അംഗത്വം" (virama + ZWNJ U+200C after ്)
	require.Len(t, matchML(t, r, "എന്തുകൊണ്ട്\u200c അംഗത്വം"), 0)
	require.Len(t, matchML(t, r, "എങ്ങനെ അംഗമാകാം?"), 0)
	// test for "LanguageTool":
	require.Len(t, matchML(t, r, "LanguageTool"), 0)
	require.Len(t, matchML(t, r, ","), 0)
	require.Len(t, matchML(t, r, "123454"), 0)

	// incorrect sentences:

	matches := matchML(t, r, "Aagohw")
	// check match positions:
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	require.Empty(t, matches[0].GetSuggestedReplacements())

	matches = matchML(t, r, "എaങ്ങനെ")
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())
	sugs := matches[0].GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Equal(t, "എങ്ങനെ", sugs[0])

	require.Len(t, matchML(t, r, "aõh"), 1)
	require.Len(t, matchML(t, r, "a"), 1)
}
