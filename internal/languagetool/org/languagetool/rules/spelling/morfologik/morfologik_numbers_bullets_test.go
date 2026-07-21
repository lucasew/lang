package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of MorfologikSpellerRule getRuleMatches numbers/bullets arm.
func TestApplyNumbersBulletsPrefix_SpaceInsert(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	// second part is correctly spelled
	sp.AddWord("hello")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)

	// "1.2hello" → suggest "1.2 hello", prevent further
	clean, before, prevent, space := r.applyNumbersBulletsPrefix("1.2hello")
	require.Equal(t, "1.2hello", clean)
	require.Equal(t, "", before)
	require.True(t, prevent)
	require.Equal(t, "1.2 hello", space)

	// "***hello" (non-letters + known word)
	clean, before, prevent, space = r.applyNumbersBulletsPrefix("***hello")
	require.True(t, prevent)
	require.Equal(t, "*** hello", space)
	require.Equal(t, "***hello", clean)
	_ = before
}

func TestApplyNumbersBulletsPrefix_StripPrefixWhenSecondMisspelled(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)

	// "1recieve" → second part misspelled: cleanWord=recieve, before="1 "
	clean, before, prevent, space := r.applyNumbersBulletsPrefix("1recieve")
	require.False(t, prevent)
	require.Equal(t, "", space)
	require.Equal(t, "recieve", clean)
	require.Equal(t, "1 ", before)

	// collectSuggestions on cleanWord with prefix
	sugs := r.collectSuggestions(clean)
	require.Contains(t, sugs, "receive")
}

func TestApplyNumbersBulletsPrefix_ExceptionDollar(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	// Java pStartsWithNumbersBulletsExceptions matches $... → no strip
	clean, before, prevent, space := r.applyNumbersBulletsPrefix("$100")
	require.Equal(t, "$100", clean)
	require.Equal(t, "", before)
	require.False(t, prevent)
	require.Equal(t, "", space)
}

func TestApplyNumbersBulletsPrefix_PlainWordUnchanged(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	clean, before, prevent, space := r.applyNumbersBulletsPrefix("recieve")
	require.Equal(t, "recieve", clean)
	require.Equal(t, "", before)
	require.False(t, prevent)
	require.Equal(t, "", space)
}

func TestMatch_NumbersBullets_SuggestsSpace(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	// Digits make Speller.isMisspelled return false when ignore-numbers is true
	// (morfologik default / many LT dicts). Disable so "1hello" is a misspelling
	// and getRuleMatches numbers/bullets arm runs (Java same gate).
	sp.IgnoreNumbers = false
	sp.AddWord("hello")
	// WordTokenizer keeps "1hello" as one token (not "1.2hello" which splits on '.')
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	ms, err := r.Match(languagetool.AnalyzePlain("1hello"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	found := false
	for _, m := range ms {
		for _, s := range m.GetSuggestedReplacements() {
			if s == "1 hello" {
				found = true
			}
		}
	}
	require.True(t, found, "expected space insert among matches %+v", ms)
}
