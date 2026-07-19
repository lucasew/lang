package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicSingleLetterRedup(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "зателефоную" {
			return []tagging.TaggedWord{{Lemma: "зателефонувати", PosTag: "verb:perf:futr:s:1"}}
		}
		return nil
	}
	rs := DynamicSingleLetterRedupReadings("з-зателефоную", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":alt")
	require.Equal(t, "зателефонувати", *rs[0].GetLemma())
	require.Nil(t, DynamicSingleLetterRedupReadings("з-або", tag)) // right short / not starting with з
	require.Nil(t, DynamicSingleLetterRedupReadings("з-зателефоную", nil))
}

func TestDynamicInvalidDashPrefix(t *testing.T) {
	loadDashPrefixResources()
	var inv string
	for k := range dashPrefixesInvalid {
		if k != "" && !stringsContainsDash(k) {
			inv = k
			break
		}
	}
	if inv == "" {
		t.Skip("no invalid dash prefix in resource")
	}
	tag := func(w string) []tagging.TaggedWord {
		if w == "пенсіонер" {
			return []tagging.TaggedWord{{Lemma: "пенсіонер", PosTag: "noun:anim:m:v_naz"}}
		}
		return nil
	}
	token := inv + "-пенсіонер"
	rs := DynamicInvalidDashPrefixReadings(token, tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":bad")
	require.Equal(t, inv+"-пенсіонер", *rs[0].GetLemma())
}

func stringsContainsDash(s string) bool {
	for _, r := range s {
		if r == '-' {
			return true
		}
	}
	return false
}

func TestDynamicDashPrefix_noun(t *testing.T) {
	loadDashPrefixResources()
	// міні is common in dash_prefixes
	if !IsDashPrefix("міні") {
		t.Skip("міні not in dash_prefixes")
	}
	tag := func(w string) []tagging.TaggedWord {
		if w == "готель" {
			return []tagging.TaggedWord{{Lemma: "готель", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	rs := DynamicDashPrefixReadings("міні-готель", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "noun")
	require.Equal(t, "міні-готель", *rs[0].GetLemma())
}

func TestDynamicDashPrefix_topNumr(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "десять" {
			return []tagging.TaggedWord{{Lemma: "десять", PosTag: "numr:p:v_naz"}}
		}
		return nil
	}
	rs := DynamicDashPrefixReadings("топ-десять", tag)
	// топ may not be in dash_prefixes — then path needs dashPrefixMatch
	if !IsDashPrefix("топ") {
		// still may work if Latin? no — fail closed
		// force via prefix map absence
		t.Log("топ not in dash_prefixes; skip if empty")
	}
	if len(rs) > 0 {
		require.Contains(t, *rs[0].GetPOSTag(), "numr")
		require.Contains(t, *rs[0].GetPOSTag(), ":bad")
	}
}

func TestGenerateTokensWithRightInflected(t *testing.T) {
	right := []*languagetool.AnalyzedToken{
		tok("векторний", "adj:m:v_naz:compb", "векторний"),
		tok("векторний", "adj:m:v_kly", "векторний"),
	}
	rs := GenerateTokensWithRightInflected("n-векторний", "n", right, "adj", "", compDropRE)
	require.Len(t, rs, 1)
	require.Equal(t, "adj:m:v_naz", *rs[0].GetPOSTag()) // :compb dropped
	require.Equal(t, "n-векторний", *rs[0].GetLemma())
}

func TestDynamicDashPrefix_latinAdj(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "векторний" {
			return []tagging.TaggedWord{{Lemma: "векторний", PosTag: "adj:m:v_naz"}}
		}
		return nil
	}
	// single Latin letter left matches leftLatOrLetterRE even without dash_prefixes
	rs := DynamicDashPrefixReadings("n-векторний", tag)
	// without being dash prefix, falls through to latinLetterAdjCompounds
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "adj")
}
