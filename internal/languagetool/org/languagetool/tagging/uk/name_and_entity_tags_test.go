package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestNameSuffixPOS(t *testing.T) {
	require.Equal(t, "noun:anim:m:v_naz:prop:lname", NameSuffixPOS("Петренко"))
	require.Equal(t, "noun:anim:m:v_naz:prop:lname", NameSuffixPOS("Тимошенко"))
	require.Equal(t, "", NameSuffixPOS("петренко")) // not capitalized
	require.Equal(t, "", NameSuffixPOS("дім"))
}

func TestNumberedEntityPOS(t *testing.T) {
	require.NotEmpty(t, NumberedEntityPOS("Т-80"))
	require.NotEmpty(t, NumberedEntityPOS("Ан-225"))
}

func TestUkrainianTagger_ProperNameAllCaps(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"НАТО"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "prop")
}

func TestUkrainianTagger_NumberedEntities(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"Т-80"})
	require.True(t, got[0].IsTagged())
}

func TestUkrainianTagger_NameSuffix(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"Шевченко"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "lname")
}
