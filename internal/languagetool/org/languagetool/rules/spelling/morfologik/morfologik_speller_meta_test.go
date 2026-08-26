package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of morfologik Speller.isMisspelled DictionaryMetadata gates (defaults + en_US.info).
func TestMorfologikSpeller_IgnoreNumbersDefault(t *testing.T) {
	// Non-EN path keeps default ignore-numbers
	sp := NewMorfologikSpeller("/de/hunspell/de_DE.dict", 1)
	sp.AddWord("ok")
	require.True(t, sp.IgnoreNumbers)
	// words with digits are never misspelled when ignoring numbers
	require.False(t, sp.IsMisspelled("175ºC"))
	require.False(t, sp.IsMisspelled("0º"))
	require.False(t, sp.IsMisspelled("5¼"))
	require.False(t, sp.IsMisspelled("123454"))
	// unknown without digits still misspelled
	require.True(t, sp.IsMisspelled("sdadsadas"))
	require.False(t, sp.IsMisspelled("ok"))
}

func TestMorfologikSpeller_EnglishInfoOverrides(t *testing.T) {
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	// en_US.info: ignore-camel-case=false, ignore-all-uppercase=false; numbers still ignored
	require.True(t, sp.IgnoreNumbers)
	require.False(t, sp.IgnoreCamelCase)
	require.False(t, sp.IgnoreAllUppercase)
	require.False(t, sp.IsMisspelled("175ºC"))
	require.False(t, sp.IsMisspelled("0º"))
	// unknown alphabetic still misspelled
	require.True(t, sp.IsMisspelled("sdadsadas"))
	// convertCase: WATER accepted if water known
	sp.AddWord("water")
	require.False(t, sp.IsMisspelled("Water"))
	require.False(t, sp.IsMisspelled("WATER"))
}

func TestMorfologikSpeller_IgnoreNumbersOff(t *testing.T) {
	sp := NewMorfologikSpeller("/x.dict", 1)
	sp.IgnoreNumbers = false
	sp.AddWord("ok")
	// Java: ignore-all-uppercase defaults true, so "175ºC" (no lowercase letter) is accepted.
	// Probe a mixed-case form with digits so only the numbers gate is under test.
	require.True(t, sp.IsMisspelled("word1"))
	require.False(t, sp.IsMisspelled("175ºC"), "all-upper+digits still ignored by ignore-all-uppercase")
}
