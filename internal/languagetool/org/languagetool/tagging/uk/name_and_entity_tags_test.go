package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestNameSuffixPOS(t *testing.T) {
	// Fail closed: no invent prop:lname from surface suffix without dictionary.
	require.Equal(t, "", NameSuffixPOS("Петренко"))
	require.Equal(t, "", NameSuffixPOS("Тимошенко"))
	require.Equal(t, "", NameSuffixPOS("петренко"))
	require.Equal(t, "", NameSuffixPOS("дім"))
}

func TestNumberedEntityPOS(t *testing.T) {
	// Official entities.txt patterns (not invent generic letter-digit regex)
	require.NotEmpty(t, NumberedEntityPOS("Т-80"))
	require.NotEmpty(t, NumberedEntityPOS("Ан-225")) // А[нН]-[А-Я0-9-]+
	require.Empty(t, NumberedEntityPOS("random"))
	require.Empty(t, NumberedEntityPOS("Ффф-999")) // no matching entities.txt line
}

func TestEntityReadings(t *testing.T) {
	rs := EntityReadings("Т-80")
	require.NotEmpty(t, rs)
	require.NotNil(t, rs[0].GetPOSTag())
	require.Contains(t, *rs[0].GetPOSTag(), "noun")
	require.Empty(t, EntityReadings("xyzzy"))
}

func TestUkrainianTagger_ProperNameAllCaps(t *testing.T) {
	// Java: ALLCAPS → capitalizeProperName + dict prop/noninfl (no invent without dict).
	wt := tagging.MapWordTagger{
		"Нато": {tagging.NewTaggedWord("Нато", "noun:inanim:m:v_naz:prop")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"НАТО"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "prop")
	// without dict fails closed
	tg2 := NewUkrainianTagger(tagging.MapWordTagger{})
	require.False(t, tg2.Tag([]string{"НАТО"})[0].IsTagged())
}

func TestCapitalizeProperName(t *testing.T) {
	require.Equal(t, "Нато", capitalizeProperName("НАТО"))
	require.Equal(t, "Київ", capitalizeProperName("КИЇВ"))
	require.Equal(t, "Івано-Франківськ", capitalizeProperName("ІВАНО-ФРАНКІВСЬК"))
}

func TestUkrainianTagger_NumberedEntities(t *testing.T) {
	tg := NewUkrainianTagger(tagging.MapWordTagger{})
	got := tg.Tag([]string{"Т-80", "Ан-225", "Ффф-999"})
	require.True(t, got[0].IsTagged(), "Т-80 from entities.txt")
	require.True(t, got[1].IsTagged(), "Ан-225 from А[нН]-… pattern")
	require.False(t, got[2].IsTagged(), "no invent outside entities.txt")
}

func TestUkrainianTagger_NameSuffix(t *testing.T) {
	// Java: surnames from dictionary, not surface invent.
	wt := tagging.MapWordTagger{
		"Шевченко": {tagging.NewTaggedWord("Шевченко", "noun:anim:m:v_naz:prop:lname")},
	}
	tg := NewUkrainianTagger(wt)
	got := tg.Tag([]string{"Шевченко"})
	require.True(t, got[0].IsTagged())
	require.Contains(t, *got[0].GetReadings()[0].GetPOSTag(), "lname")
	// without dict fails closed
	require.False(t, NewUkrainianTagger(tagging.MapWordTagger{}).Tag([]string{"Шевченко"})[0].IsTagged())
}
