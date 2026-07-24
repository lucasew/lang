package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestCollapseStretch(t *testing.T) {
	require.Equal(t, "дуже", collapseStretch("ду-у-у-же"))
	require.Equal(t, "Так", collapseStretch("Та-а-ак"))
	// capitalized surface → capitalized collapsed form (Java StringUtils.capitalize)
	require.Equal(t, "Му", collapseStretch("Му-у-у"))
	require.Equal(t, "му", collapseStretch("му-у-у"))
	require.Equal(t, "вареники", collapseStretch("ва-ре-ни-ки"))
}

func TestDynamicMultiHyphen_mergeAlt(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "вареники" || w == "Вареники" {
			return []tagging.TaggedWord{
				{Lemma: "вареник", PosTag: "noun:inanim:p:v_naz"},
				{Lemma: "вареник", PosTag: "noun:inanim:p:v_naz:abbr"},
			}
		}
		return nil
	}
	rs := DynamicMultiHyphenStretchReadings("ва-ре-ни-ки", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.NotContains(t, *rs[0].GetPOSTag(), "abbr")
	require.Equal(t, "вареник", *rs[0].GetLemma())
	require.Nil(t, DynamicMultiHyphenStretchReadings("ва-ре-ни-ки", nil))
	require.Nil(t, DynamicMultiHyphenStretchReadings("ва-ре-ни-ки", func(string) []tagging.TaggedWord { return nil }))
}

func TestDynamicMultiHyphen_stretchAlt(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "так", "Так":
			return []tagging.TaggedWord{{Lemma: "так", PosTag: "adv:pron:dem"}}
		case "му", "Му":
			return []tagging.TaggedWord{{Lemma: "му", PosTag: "noninfl:onomat:predic"}}
		}
		return nil
	}
	rs := DynamicMultiHyphenStretchReadings("Та-а-ак", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "так", *rs[0].GetLemma())

	rs2 := DynamicMultiHyphenStretchReadings("Му-у-у", tag)
	require.NotEmpty(t, rs2)
	require.Contains(t, *rs2[0].GetPOSTag(), "onomat")
}

func TestDynamicMultiHyphen_skipDashPrefix(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		return []tagging.TaggedWord{{Lemma: "x", PosTag: "noun:inanim:m:v_naz"}}
	}
	loadDashPrefixResources()
	var pref string
	for k := range dashPrefixes {
		if !strings.Contains(k, "-") && IsDashPrefix(k) {
			pref = k
			break
		}
	}
	if pref == "" {
		t.Skip("no single-token dash prefix")
	}
	// avoid 3-part EntityReadings: use 4 parts so only merge/stretch arms apply
	token := pref + "-а-б-в"
	require.True(t, IsDashPrefix(pref))
	require.Nil(t, DynamicMultiHyphenStretchReadings(token, tag), "prefix %q should skip merge", pref)
}

func TestTagBothCases(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "слово":
			return []tagging.TaggedWord{{Lemma: "слово", PosTag: "noun:inanim:n:v_naz"}}
		case "Слово":
			return []tagging.TaggedWord{{Lemma: "Слово", PosTag: "noun:inanim:n:v_naz:prop"}}
		}
		return nil
	}
	got := tagBothCases("слово", tag, nil)
	require.GreaterOrEqual(t, len(got), 1)
}
