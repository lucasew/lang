package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestMatch_SkipsURLToken(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/test.dict", 1)
	sp.AddWord("hello")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_XX", "en", "/xx/test.dict", sp)
	// single URL token
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("https://example.com/foo", nil, nil))
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		atr,
	})
	require.True(t, spelling.CanBeIgnoredToken(atr))
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_IgnorePotentiallyMisspelledWord(t *testing.T) {
	sp := NewMorfologikSpeller("/xx/test.dict", 1)
	sp.AddWord("hello")
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_XX", "en", "/xx/test.dict", sp)
	// "xyzzy" not in dict → would be misspelled unless potential-ignore accepts it
	r.IgnorePotentiallyMisspelledWordFn = func(word string) bool {
		return word == "xyzzy"
	}
	sent := languagetool.AnalyzePlain("hello xyzzy")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
	// still flags other misspellings
	r.IgnorePotentiallyMisspelledWordFn = nil
	m, err = r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func strPtr(s string) *string { return &s }
