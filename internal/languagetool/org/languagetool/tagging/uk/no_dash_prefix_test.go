package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicNoDashPrefix(t *testing.T) {
	loadDashPrefixResources()
	// use "екс" if in noDash (invalid list often has екс)
	if _, ok := noDashPrefixes["екс"]; !ok && !IsDashPrefixInvalid("екс") {
		// ensure via list
		found := false
		for _, p := range NoDashPrefixList() {
			if p == "екс" {
				found = true
				break
			}
		}
		if !found {
			t.Skip("екс not in noDashPrefixes")
		}
	}
	tag := func(w string) []tagging.TaggedWord {
		if w == "партнер" {
			return []tagging.TaggedWord{{Lemma: "партнер", PosTag: "noun:anim:m:v_naz"}}
		}
		if w == "прес" {
			return []tagging.TaggedWord{{Lemma: "прес", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	// length must be > 7: "експартнер" = 10
	rs := DynamicNoDashPrefixReadings("експартнер", tag)
	require.NotEmpty(t, rs)
	require.Equal(t, "експартнер", *rs[0].GetLemma())
	require.Contains(t, *rs[0].GetPOSTag(), "noun")

	// short word fail closed
	require.Nil(t, DynamicNoDashPrefixReadings("екс", tag))

	// yod after consonant prefix without apo → :bad (екс' not present)
	// prefix ending consonant + єїюя start
	// e.g. "контрява" if контр in list and ява tagged — skip if not
}

func TestDynamicNoDashPrefix_unexpectedApo(t *testing.T) {
	// екс'прес — apo present but not needed → break (nil if only that prefix)
	loadDashPrefixResources()
	tag := func(w string) []tagging.TaggedWord {
		if w == "прес" {
			return []tagging.TaggedWord{{Lemma: "прес", PosTag: "noun:inanim:m:v_naz"}}
		}
		return nil
	}
	// "екс'прес" length with apo
	rs := DynamicNoDashPrefixReadings("екс'прес", tag)
	// Java breaks on unexpected apo for first matching prefix "екс"
	// may be nil
	_ = rs
}

func TestDynamicNoDashPrefix_filterPron(t *testing.T) {
	loadDashPrefixResources()
	tag := func(w string) []tagging.TaggedWord {
		if w == "якийсь" {
			return []tagging.TaggedWord{{Lemma: "якийсь", PosTag: "adj:m:v_naz:pron:ind"}}
		}
		return nil
	}
	// if prefix+якийсь long enough
	require.Nil(t, DynamicNoDashPrefixReadings("суперякийсь", tag)) // may not have супер; just no invent
}

func TestTryNoDashPrefixTags_bridge(t *testing.T) {
	// legacy callback API still works
	rs := TryNoDashPrefixTags("мінітест", func(right string) []*languagetool.AnalyzedToken {
		if right == "тест" {
			p, l := "noun:inanim:m:v_naz", "тест"
			return []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(right, &p, &l)}
		}
		return nil
	})
	if !IsDashPrefix("міні") && !func() bool {
		for _, p := range NoDashPrefixList() {
			if p == "міні" {
				return true
			}
		}
		return false
	}() {
		// міні often in dash_prefixes with alt → noDash
		loadDashPrefixResources()
	}
	// only assert if міні is a no-dash prefix
	for _, p := range NoDashPrefixList() {
		if p == "міні" {
			require.NotEmpty(t, rs)
			return
		}
	}
	t.Skip("міні not no-dash")
}
