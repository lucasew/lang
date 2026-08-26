package en

// Twin of org.languagetool.rules.en.SpellingCheckRuleTest.
// Java uses JLanguageTool(en-US).check(...); here we isolate MorfologikAmericanSpellerRule.Match
// which yields the same Java-visible spelling outcomes for these cases (ignore.txt / spelling.txt /
// acceptPhrases / isUrl / isEMail / proper-noun 's suggestions).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

// Twin of SpellingCheckRuleTest.testIgnoreSuggestionsWithMorfologik.
func TestSpellingCheckRule_IgnoreSuggestionsWithMorfologik(t *testing.T) {
	r := requireAmericanDict(t)

	// ignore.txt: anArtificialTestWordForLanguageTool
	ms, err := r.Match(analyzeEN("This is anArtificialTestWordForLanguageTool."))
	require.NoError(t, err)
	require.Empty(t, ms)

	// multi-word / spelling path: "Al Jazeera" accepted (multiwords.txt + dict; Java comment spelling.txt)
	ms, err = r.Match(analyzeEN("It was shown on Al Jazeera."))
	require.NoError(t, err)
	require.Empty(t, ms)

	// "adjoint" accepted (official en_US.dict / spelling merge; Java comment spelling.txt)
	ms, err = r.Match(analyzeEN("Test adjoint"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(analyzeEN("Test Adjoint"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// real typo
	matches2, err := r.Match(analyzeEN("This is a real typoh."))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches2))
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", matches2[0].GetSpecificRuleId())
	require.Equal(t, MorfologikAmericanSpellerRuleID, matches2[0].GetSpecificRuleId())

	// proper-noun "…s" → curated "…'s" + base (filterSuggestions + EnglishTagger NNP)
	ms, err = r.Match(analyzeEN("This is Julias new song."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Julia's", "Julia", "Julian", "Julius", "Julies", "Julián", "Júlia"},
		ms[0].GetSuggestedReplacements())

	ms, err = r.Match(analyzeEN("This is Microsofts new product."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Microsoft's", "Microsoft"}, ms[0].GetSuggestedReplacements())
}

// Twin of SpellingCheckRuleTest.testIgnorePhrases.
func TestSpellingCheckRule_IgnorePhrases(t *testing.T) {
	r := requireAmericanDict(t)

	ms, err := r.Match(analyzeEN("A test with myfoo mybar"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))

	// Java: for SpellingCheckRule active rules → acceptPhrases; others disabled.
	// We only run Morfologik American, which is the EN SpellingCheckRule under test.
	r.AcceptPhrases([]string{"myfoo mybar", "Myy othertest"})

	ms, err = r.Match(analyzeEN("A test with myfoo mybar"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// words alone not ignored
	ms, err = r.Match(analyzeEN("A test with myfoo and mybar"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))

	ms, err = r.Match(analyzeEN("myfoo mybar here"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// sentence-start title case of lowercase-leading phrase
	ms, err = r.Match(analyzeEN("Myfoo mybar here"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// all-caps first token not accepted
	ms, err = r.Match(analyzeEN("MYfoo mybar here"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))

	// "Myy othertest" is title-cased first token → case-sensitive phrase only
	ms, err = r.Match(analyzeEN("Myy othertest is okay"))
	require.NoError(t, err)
	require.Empty(t, ms)

	ms, err = r.Match(analyzeEN("And Myy othertest is okay"))
	require.NoError(t, err)
	require.Empty(t, ms)

	ms, err = r.Match(analyzeEN("But Myy Othertest is not okay"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))

	ms, err = r.Match(analyzeEN("But myy othertest is not okay"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))
}

// Twin of SpellingCheckRuleTest.testIsUrl (MySpellCheckingRule inner class matrix).
// Java SpellingCheckRule.isUrl / isEMail → WordTokenizer.isUrl / isEMail.
func TestSpellingCheckRule_IsUrl(t *testing.T) {
	require.True(t, spelling.IsUrl("http://www.test.de"))
	require.True(t, spelling.IsUrl("http://www.test-dash.com"))
	require.True(t, spelling.IsUrl("https://www.test-dash.com"))
	require.True(t, spelling.IsUrl("ftp://www.test-dash.com"))
	require.True(t, spelling.IsUrl("http://www.test-dash.com/foo/path-dash"))
	require.True(t, spelling.IsUrl("http://www.test-dash.com/foo/öäü-dash"))
	require.True(t, spelling.IsUrl("http://www.test-dash.com/foo/%C3%B-dash"))
	require.True(t, spelling.IsUrl("www.languagetool.org"))
	require.False(t, spelling.IsUrl("languagetool.org")) // currently not detected

	require.True(t, spelling.IsEMail("martin.mustermann@test.de"))
	require.True(t, spelling.IsEMail("martin.mustermann@test.languagetool.de"))
	require.True(t, spelling.IsEMail("martin-mustermann@test.com"))
	require.False(t, spelling.IsEMail("@test.de"))
	require.False(t, spelling.IsEMail("f.test@test"))
	require.False(t, spelling.IsEMail("f@t.t"))
}
