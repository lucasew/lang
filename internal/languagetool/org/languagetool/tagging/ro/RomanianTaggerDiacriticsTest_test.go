package ro

// Twin of RomanianTaggerDiacriticsTest.java — UTF-8 /ro/test_diacritics.dict.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireDiacriticsTagger(t *testing.T) *RomanianTagger {
	t.Helper()
	p := DiscoverRomanianDiacriticsDict()
	if p == "" {
		t.Skip("test_diacritics.dict not in tree")
	}
	tagger := OpenRomanianTaggerFromFilesystem(p, RomanianTestDiacriticsDictPath)
	require.NotNil(t, tagger, "failed to open test_diacritics.dict at %s", p)
	require.Equal(t, RomanianTestDiacriticsDictPath, tagger.GetDictionaryPath())
	require.NotNil(t, tagger.GetWordTagger())
	return tagger
}

// Twin of RomanianTaggerDiacriticsTest.testTaggerMerseseram
func TestRomanianTaggerDiacritics_TaggerMerseseram(t *testing.T) {
	tagger := requireDiacriticsTagger(t)
	assertHasLemmaAndPos(t, tagger, "făcusem", "face", "004")
	assertHasLemmaAndPos(t, tagger, "cuțitul", "cuțit", "002")
	// make sure lemma is correct (POS is hard-coded, not important)
	assertHasLemmaAndPos(t, tagger, "merseserăm", "merge", "002")
}

// Twin of RomanianTaggerDiacriticsTest.testTaggerCuscaCutit
func TestRomanianTaggerDiacritics_TaggerCuscaCutit(t *testing.T) {
	tagger := requireDiacriticsTagger(t)
	assertHasLemmaAndPos(t, tagger, "cușcă", "cușcă", "001")
	assertHasLemmaAndPos(t, tagger, "cuțit", "cuțit", "001")
	assertHasLemmaAndPos(t, tagger, "cuțitul", "cuțit", "002")
}
