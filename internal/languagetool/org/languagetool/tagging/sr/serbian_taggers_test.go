package sr

// Path / singleton wiring for SerbianTagger + Ekavian + Jekavian (real dicts).

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbianTaggers_DictionaryPaths(t *testing.T) {
	require.Equal(t, EkavianDictionaryPath, NewSerbianTagger(nil).GetDictionaryPath())
	require.Equal(t, EkavianDictionaryPath, NewEkavianTagger(nil).GetDictionaryPath())
	require.Equal(t, JekavianDictionaryPath, NewJekavianTagger(nil).GetDictionaryPath())
	require.True(t, NewSerbianTagger(nil).TagLowercaseWithUppercase)
	require.True(t, NewEkavianTagger(nil).TagLowercaseWithUppercase)
	require.True(t, NewJekavianTagger(nil).TagLowercaseWithUppercase)
}

func TestSerbianTagger_DefaultLoadsEkavian(t *testing.T) {
	if DiscoverEkavianPOSDict() == "" {
		t.Skip("ekavian/serbian.dict not in tree")
	}
	EnsureDefaultSerbianTagger()
	require.NotNil(t, DefaultSerbianTagger)
	require.NotNil(t, DefaultSerbianTagger.GetWordTagger())
	require.Equal(t, EkavianDictionaryPath, DefaultSerbianTagger.GetDictionaryPath())
	// Sample lemma/POS from binary dict (same as Ekavian test surface).
	assertHasLemmaAndPos(t, DefaultSerbianTagger, "радим", "радити", "GL:GV:PZ:1L:0J")
}
