package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestArabicTagManagerFlags(t *testing.T) {
	m := NewArabicTagManager()
	// noun: N? positions — craft a 12-char noun tag
	// indices: 0=N wordtype, 4=gender, 5=number, 6=case, 9=conj, 10=jar, 11=pronoun
	tag := []rune("Nxx-F1I-x---") // length 12
	require.Equal(t, 12, len(tag))
	s := string(tag)
	require.True(t, m.IsNoun(s))
	require.False(t, m.IsVerb(s))
	require.True(t, m.IsFeminin(s))
	require.True(t, m.IsMajrour(s))

	s2 := m.SetJar(s, "ب")
	require.True(t, m.HasJar(s2))
	require.Equal(t, 'B', m.GetFlag(s2, "JAR"))

	s3 := m.SetConjunction(s2, "و")
	require.True(t, m.HasConjunction(s3))

	v := "Vxxxxxxxxxxxxxx" // 15 chars
	require.True(t, m.IsVerb(v[:15]))
	require.True(t, m.IsStopWord("Pxxxxxxxxxx"))
}

// Unit: strip tashkeel before WordTagger lookup (Java ArabicStringTools.removeTashkeel).
// MapWordTagger is only for isolating the strip path — real outcomes use arabic.dict.
func TestArabicTagger_RemoveTashkeelLookup(t *testing.T) {
	wt := tagging.MapWordTagger{
		"كتب": {tagging.NewTaggedWord("كتب", "Nxx-M1I-x---")},
	}
	tagger := NewArabicTagger(wt)
	withTashkeel := "كَتب"
	require.Equal(t, "كتب", tools.RemoveTashkeel(withTashkeel))
	got := tagger.Tag([]string{withTashkeel, "xyz"})
	require.Len(t, got, 2)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

func TestArabicHybridDisambiguator(t *testing.T) {
	d := NewDefaultArabicHybridDisambiguator()
	require.Nil(t, d.Disambiguate(nil))
	empty := languagetool.NewAnalyzedSentence(nil)
	require.NotNil(t, d.Disambiguate(empty))
}
