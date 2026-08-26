package uk

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestGetNvPrefixNounMatch(t *testing.T) {
	right := []*languagetool.AnalyzedToken{
		tok("БПЛА", "noun:inanim:n:v_naz:prop", "БПЛА"),
		tok("БПЛА", "noun:inanim:n:v_kly:prop", "БПЛА"), // skipped
	}
	rs := GetNvPrefixNounMatch("міні-БПЛА", right, "міні", ":alt")
	require.Len(t, rs, 1)
	// capitalized lemma → skip :alt
	require.NotContains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "міні-БПЛА", *rs[0].GetLemma())

	right2 := []*languagetool.AnalyzedToken{tok("готель", "noun:inanim:m:v_naz", "готель")}
	rs2 := GetNvPrefixNounMatch("міні-готель", right2, "міні", ":alt")
	require.NotEmpty(t, rs2)
	require.Contains(t, *rs2[0].GetPOSTag(), ":alt")
}

func TestDynamicTwoHyphen_adjChain(t *testing.T) {
	// second adj + third adj equal POS → TagMatch
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "нігерійський", "нігерійсько":
			return nil
		case "нігерійська":
			return []tagging.TaggedWord{{Lemma: "нігерійський", PosTag: "adj:f:v_naz"}}
		case "зімбабвійський":
			return []tagging.TaggedWord{{Lemma: "зімбабвійський", PosTag: "adj:m:v_naz"}}
		case "зімбабвійська":
			return []tagging.TaggedWord{{Lemma: "зімбабвійський", PosTag: "adj:f:v_naz"}}
		case "іранський":
			return []tagging.TaggedWord{{Lemma: "іранський", PosTag: "adj:m:v_naz"}}
		}
		// second part surface as-is
		if w == "нігерійсько" {
			return nil
		}
		return nil
	}
	// Use surfaces that tag as adj with same POS after strip
	tag = func(w string) []tagging.TaggedWord {
		switch w {
		case "а", "б", "в":
			return nil
		case "червоний":
			return []tagging.TaggedWord{{Lemma: "червоний", PosTag: "adj:m:v_naz"}}
		case "синій":
			return []tagging.TaggedWord{{Lemma: "синій", PosTag: "adj:m:v_naz"}}
		case "жовтий":
			return []tagging.TaggedWord{{Lemma: "жовтий", PosTag: "adj:m:v_naz"}}
		}
		return nil
	}
	// parts: червоний-синій-жовтий — second+third adj match; returns second-third TagMatch on full surface
	rs := DynamicTwoHyphenReadings("червоний-синій-жовтий", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "adj")
	require.Equal(t, "синій-жовтий", *rs[0].GetLemma()) // left lemma from second, right from third
}

func TestDynamicTwoHyphen_failClosed(t *testing.T) {
	require.Nil(t, DynamicTwoHyphenReadings("а-б", nil))
	require.Nil(t, DynamicTwoHyphenReadings("один-два", func(string) []tagging.TaggedWord { return nil }))
	require.Nil(t, DynamicTwoHyphenReadings("a-b-c-d", func(string) []tagging.TaggedWord {
		return []tagging.TaggedWord{{Lemma: "x", PosTag: "adj:m:v_naz"}}
	}))
}

func TestDynamicTwoHyphen_dashPrefix(t *testing.T) {
	// If official dash_prefixes has a two-part key this would fire; inject via IsDashPrefix path
	// by using a known prefix if any two-dash keys exist. Otherwise skip when resource missing.
	loadDashPrefixResources()
	var twoKey string
	for k := range dashPrefixes {
		if strings.Count(k, "-") == 1 {
			twoKey = k
			break
		}
	}
	if twoKey == "" {
		t.Skip("no two-part dash prefix in resource")
	}
	parts := strings.SplitN(twoKey, "-", 2)
	token := twoKey + "-готель"
	tag := func(w string) []tagging.TaggedWord {
		if w == "готель" {
			return []tagging.TaggedWord{{Lemma: "готель", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	rs := DynamicTwoHyphenReadings(token, tag)
	require.NotEmpty(t, rs, "prefix %s", twoKey)
	require.Contains(t, *rs[0].GetPOSTag(), "noun")
	_ = parts
}
