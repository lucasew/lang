package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestAltTagAdjust_CapsInside(t *testing.T) {
	// lower + Upper + lower interior → tag lower form + :alt
	tag := func(w string) []tagging.TaggedWord {
		if w == "україна" {
			return []tagging.TaggedWord{{Lemma: "україна", PosTag: "noun:inanim:f:v_naz:prop:geo"}}
		}
		return nil
	}
	// "укрАїна" has lower-upper-lower
	rs := AltTagAdjustReadings("укрАїна", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "укрАїна", rs[0].GetToken())
	// fail closed without dict
	require.Empty(t, AltTagAdjustReadings("укрАїна", func(string) []tagging.TaggedWord { return nil }))
	// no caps inside
	require.Empty(t, AltTagAdjustReadings("україна", tag))
}

func TestAltTagAdjust_ZLabial(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "спати" || w == "Спати" {
			return []tagging.TaggedWord{{Lemma: "спати", PosTag: "verb:imperf:inf"}}
		}
		return nil
	}
	// зпати → спати (з before п)
	rs := AltTagAdjustReadings("зпатись", func(w string) []tagging.TaggedWord {
		if w == "спатись" {
			return []tagging.TaggedWord{{Lemma: "спатися", PosTag: "verb:imperf:inf"}}
		}
		return nil
	})
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	// lemma maps с → з
	require.NotNil(t, rs[0].GetLemma())
	require.Equal(t, "зпатися", *rs[0].GetLemma())
	_ = tag
}

func TestAltTagAdjust_YiToI(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "дівчина" {
			return []tagging.TaggedWord{{Lemma: "дівчина", PosTag: "noun:anim:f:v_naz"}}
		}
		return nil
	}
	rs := AltTagAdjustReadings("дївчина", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Empty(t, AltTagAdjustReadings("дївчина", func(string) []tagging.TaggedWord { return nil }))
}

func TestAltTagAdjust_ConvertGhe(t *testing.T) {
	// ґ → г + :alt; lemma maps back
	tag := func(w string) []tagging.TaggedWord {
		if w == "грунт" {
			return []tagging.TaggedWord{{Lemma: "грунт", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	rs := AltTagAdjustReadings("ґрунт", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "ґрунт", *rs[0].GetLemma())
}

func TestAltTagAdjust_ConvertTер(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "компютер" {
			return nil
		}
		if w == "компютр" {
			return []tagging.TaggedWord{{Lemma: "компютр", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	// ends with тер → тр
	rs := convertTokenAltReadings("компютер", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "компютер", *rs[0].GetLemma())
}

func TestAnalyzeAllCapitalizedAdj(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "івано-франківська" {
			return []tagging.TaggedWord{
				{Lemma: "івано-франківський", PosTag: "adj:f:v_naz"},
				{Lemma: "івано-франківський", PosTag: "noun:inanim:f:v_naz:prop"},
			}
		}
		return nil
	}
	rs := AnalyzeAllCapitalizedAdj("Івано-Франківська", tag)
	require.NotEmpty(t, rs)
	for _, r := range rs {
		require.Contains(t, *r.GetPOSTag(), "adj")
	}
	// not all capitalized parts
	require.Empty(t, AnalyzeAllCapitalizedAdj("Івано-франківська", tag))
	// no adj in dict
	require.Empty(t, AnalyzeAllCapitalizedAdj("Київ-Прага", func(string) []tagging.TaggedWord {
		return []tagging.TaggedWord{{Lemma: "x", PosTag: "noun:inanim:m:v_naz:prop"}}
	}))
}

func TestMergeUniqueAnalyzedTokens_capitalizedAdj(t *testing.T) {
	// Java: when surface already has tags, analyzeAllCapitamizedAdj still adds adj readings.
	pNoun, lNoun := "noun:inanim:f:v_naz:prop", "івано-франківська"
	existing := []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Івано-Франківська", &pNoun, &lNoun),
	}
	tag := func(w string) []tagging.TaggedWord {
		if w == "івано-франківська" {
			return []tagging.TaggedWord{
				{Lemma: "івано-франківський", PosTag: "adj:f:v_naz"},
			}
		}
		return nil
	}
	adj := AnalyzeAllCapitalizedAdj("Івано-Франківська", tag)
	require.NotEmpty(t, adj)
	merged := mergeUniqueAnalyzedTokens(existing, adj)
	require.Len(t, merged, 2)
	// re-merge is idempotent
	require.Len(t, mergeUniqueAnalyzedTokens(merged, adj), 2)
}

func TestFilterMissingApoTags(t *testing.T) {
	in := []tagging.TaggedWord{
		{Lemma: "a", PosTag: "noun:m:v_naz"},
		{Lemma: "b", PosTag: "noun:m:v_naz:bad"},
		{Lemma: "c", PosTag: "noun:m:v_naz:arch"},
	}
	out := filterMissingApoTags(in)
	require.Len(t, out, 1)
	require.Equal(t, "a", out[0].Lemma)
}
