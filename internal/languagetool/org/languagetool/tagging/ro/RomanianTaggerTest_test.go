package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/tagging/ro/RomanianTaggerTest.java

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireDefaultRO(t *testing.T) *RomanianTagger {
	t.Helper()
	if DiscoverRomanianPOSDict() == "" {
		t.Skip("romanian.dict not in tree")
	}
	EnsureDefaultRomanianTagger()
	require.NotNil(t, DefaultRomanianTagger)
	require.NotNil(t, DefaultRomanianTagger.GetWordTagger())
	return DefaultRomanianTagger
}

// Twin of RomanianTaggerTest.testTaggerMerge
func TestRomanianTagger_TaggerMerge(t *testing.T) {
	tagger := requireDefaultRO(t)
	// merge - verb indicativ imperfect, persoana întâi, singular
	assertHasLemmaAndPos(t, tagger, "mergeam", "merge", "V0s1000ii0")
	// merge - verb indicativ imperfect, persoana întâi, plural
	assertHasLemmaAndPos(t, tagger, "mergeam", "merge", "V0p1000ii0")
}

// Twin of RomanianTaggerTest.testTaggerMerseseram
func TestRomanianTagger_TaggerMerseseram(t *testing.T) {
	tagger := requireDefaultRO(t)
	// first make sure lemma is correct (ignore POS)
	assertHasLemmaAndPos(t, tagger, "merseserăm", "merge", "")
	// now that lemma is correct, also check POS
	assertHasLemmaAndPos(t, tagger, "merseserăm", "merge", "V0p1000im0")
	got := myAssertTagger(tagger, "merseserăm")
	require.Equal(t, "merseserăm/[merge]V0p1000im0", got)
}

// Twin of RomanianTaggerTest.testTagger_Fi
func TestRomanianTagger_Tagger_Fi(t *testing.T) {
	tagger := requireDefaultRO(t)
	// fi - verb indicativ prezent, persoana întâi, singular
	assertHasLemmaAndPos(t, tagger, "sunt", "fi", "V0s1000izf")
	// fi verb indicativ prezent, persoana a treia, plural
	assertHasLemmaAndPos(t, tagger, "sunt", "fi", "V0p3000izf")
}

// Twin of RomanianTaggerTest.testTaggerUserDict
// configurați comes from resource/ro/added.txt via BaseTagger CombiningTagger.
func TestRomanianTagger_TaggerUserDict(t *testing.T) {
	tagger := requireDefaultRO(t)
	assertHasLemmaAndPos(t, tagger, "configurați", "configura", "V0p2000cz0")
}

// Twin of RomanianTaggerTest.testTagger (Java TestTools.myAssert sentence).
func TestRomanianTagger_Tagger(t *testing.T) {
	tagger := requireDefaultRO(t)
	got := myAssertTagger(tagger, "Cartea este frumoasă.")
	require.Equal(t,
		"Cartea/[carte]Sfs3aac000 -- este/[fi]V0s3000izb -- frumoasă/[frumos]Afs3an0000",
		got)
}
