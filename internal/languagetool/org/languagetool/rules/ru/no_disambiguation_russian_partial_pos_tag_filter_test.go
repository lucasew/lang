package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	tagru "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/stretchr/testify/require"
)

// Java-visible outcome matrix for NoDisambiguationRussianPartialPosTagFilter
// + PartialPosTagFilter (tag path without disambiguator).
// Mirrors EN NoDisambiguationEnglishPartialPosTagFilter_test matrix.

func TestNoDisambiguationRussianPartialPosTagFilter_InjectedTagAccept(t *testing.T) {
	// Java: partial "дом" from домов tags NN:masc → postag NN.* → keep match (same object).
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "Java returns the same RuleMatch instance on accept")
}

func TestNoDisambiguationRussianPartialPosTagFilter_InjectedTagDropPOS(t *testing.T) {
	// POS of partial does not match postag_regexp → drop.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationRussianPartialPosTagFilter_RegexpNoMatchDrops(t *testing.T) {
	// Java matcher.matches() fails → return null (tag never consulted).
	called := false
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		called = true
		return []string{"NN:masc"}
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(кот)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
	require.False(t, called, "tag must not run when surface regexp fails")
}

func TestNoDisambiguationRussianPartialPosTagFilter_UntaggedPartialDrops(t *testing.T) {
	// tag returns no POS → partialTagHasRequiredTag false.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		return nil
	})
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_NegatePos(t *testing.T) {
	// negate_pos present: keep only when no reading matches postag_regexp.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		if p == "идти" {
			return []string{"VB:INF", "NN:other"}
		}
		return nil
	})
	// Has NN:masc → negate rejects against NN.*.
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
	// Has VB/NN:other, none match ^JJ$ → negate accepts.
	ok, err = f.Accept("поидти", "по(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_TwoGroupsAndPrefixSuffix(t *testing.T) {
	// two_groups: group1+group2 retagged; prefix/suffix wrap surface before match.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "домкнига" {
			return []string{"NN"}
		}
		return nil
	})
	ok, err := f.Accept("домкнига", "(дом)(книга)", "NN", false, true, "", "")
	require.NoError(t, err)
	require.True(t, ok)
	// prefix+token+suffix: prefix "дом" + token "книга" → "домкнига"
	ok, err = f.Accept("книга", "(дом)(книга)", "NN", false, true, "дом", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_FailClosedWithoutTagger(t *testing.T) {
	ClearDefaultRussianPartialPosTagger()
	t.Cleanup(ClearDefaultRussianPartialPosTagger)
	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationRussianPartialPosTagFilter_ProcessWideWire(t *testing.T) {
	// Java process-wide Languages.getLanguageForShortCode("ru").getTagger() path.
	ClearDefaultRussianPartialPosTagger()
	t.Cleanup(ClearDefaultRussianPartialPosTagger)

	WireRussianFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		if token == "дом" {
			return []languagetool.TokenTag{{POS: "NN:Inanim:Masc:Sin:Nom", Lemma: "дом"}}
		}
		if token == "идти" {
			return []languagetool.TokenTag{{POS: "VB:INF:", Lemma: "идти"}}
		}
		return nil
	})

	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)
	// Partial "дом" from "домов" → NN → keep.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("RU"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out)

	// "идти" is VB, not NN → drop.
	atr2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("идти", "", "", false, false), 0)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{
		"no": "1", "regexp": "^(идти)$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr2}, nil))
}

func TestNoDisambiguationRussianPartialPosTagFilter_MissingArgsPanic(t *testing.T) {
	// Java RuntimeException when no/regexp/postag_regexp absent.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string { return []string{"NN"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "^(дом)ов$",
			// postag_regexp missing
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "postag_regexp": "NN.*",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationRussianPartialPosTagFilter_WrongGroupCountPanic(t *testing.T) {
	// Java: expected 1 group, got 0 → RuntimeException.
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string { return []string{"NN"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "домов", "postag_regexp": "NN.*",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	// two_groups_regexp present → expect 2 groups.
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
			"two_groups_regexp": "yes",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationRussianPartialPosTagFilter_Registered(t *testing.T) {
	class := "org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class))
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	// Factory builds fail-closed instance when tagger unwired.
	ClearDefaultRussianPartialPosTagger()
	t.Cleanup(ClearDefaultRussianPartialPosTagger)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationRussianPartialPosTagFilter_OfficialRussianDict(t *testing.T) {
	// Prefer real russian.dict (same official CFSA2 resource Java RussianTagger loads).
	if tagru.DiscoverRussianPOSDict() == "" {
		t.Skip("russian.dict not in tree")
	}
	tagru.EnsureDefaultRussianTagger()
	require.NotNil(t, tagru.DefaultRussianTagger)
	require.NotNil(t, tagru.DefaultRussianTagger.GetWordTagger())
	require.NotEmpty(t, tagru.RussianPOSDictPath(), "real russian.dict must load")

	ClearDefaultRussianPartialPosTagger()
	t.Cleanup(ClearDefaultRussianPartialPosTagger)

	// Wire Java-equivalent tagger: Languages.getLanguageForShortCode("ru").getTagger().
	WireRussianFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		trs := tagru.DefaultRussianTagger.Tag([]string{token})
		if len(trs) == 0 || trs[0] == nil {
			return nil
		}
		var out []languagetool.TokenTag
		for _, r := range trs[0].GetReadings() {
			if r == nil {
				continue
			}
			pos, lemma := "", ""
			if p := r.GetPOSTag(); p != nil {
				pos = *p
			}
			if l := r.GetLemma(); l != nil {
				lemma = *l
			}
			if pos != "" {
				out = append(out, languagetool.TokenTag{POS: pos, Lemma: lemma})
			}
		}
		return out
	})

	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)

	// "дом" is in russian.dict as NN:Inanim:Masc:… → postag NN.* → accept.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("дом", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("RU"), nil, 0, 3, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "дом from russian.dict must match NN.*")

	// Same token must not match pure verb postag when only NN readings exist for "дом".
	// (дом is a noun; VB.* should fail unless a verbal reading is present.)
	// Use invented surface not in dict → untagged → drop.
	atrBad := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("xyzzyнеслово", "", "", false, false), 0)
	mBad := rules.NewRuleMatch(rules.NewFakeRule("RU"), nil, 0, 12, "msg")
	require.Nil(t, f.AcceptRuleMatch(mBad, map[string]string{
		"no": "1", "regexp": "^(xyzzyнеслово)$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrBad}, nil))

	// Full-token identity: "блукать" is VB:INF: in russian.dict (RussianTaggerTest).
	atrVB := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("блукать", "", "", false, false), 0)
	mVB := rules.NewRuleMatch(rules.NewFakeRule("VB"), nil, 0, 7, "msg")
	outVB := f.AcceptRuleMatch(mVB, map[string]string{
		"no": "1", "regexp": "^(блукать)$", "postag_regexp": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrVB}, nil)
	require.Equal(t, mVB, outVB, "блукать from russian.dict must match VB.*")

	// Verb must not keep match under NN.* alone (блукать is VB:INF: only in testTagger).
	require.Nil(t, f.AcceptRuleMatch(mVB, map[string]string{
		"no": "1", "regexp": "^(блукать)$", "postag_regexp": "NN:Inanim:.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrVB}, nil))
}

// Legacy Accept-path smoke (kept for matrix continuity with earlier sector tests).
func TestNoDisambiguationRussianPartialPosTagFilter_InjectedTag(t *testing.T) {
	f := NewNoDisambiguationRussianPartialPosTagFilter(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	ok, err := f.Accept("домов", "^(дом)ов$", "NN.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationRussianPartialPosTagFilter_DefaultTaggerHook(t *testing.T) {
	SetDefaultRussianPartialPosTagger(func(p string) []string {
		if p == "дом" {
			return []string{"NN:masc"}
		}
		return nil
	})
	t.Cleanup(ClearDefaultRussianPartialPosTagger)

	f := NewNoDisambiguationRussianPartialPosTagFilter(nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("домов", "NN:masc", "дом", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(дом)ов$", "postag_regexp": "NN.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.NotNil(t, out)
}
