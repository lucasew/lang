package en

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Java-visible outcome matrix for NoDisambiguationEnglishPartialPosTagFilter
// + PartialPosTagFilter (tag path without disambiguator).

func TestNoDisambiguationEnglishPartialPosTagFilter_InjectedTagAccept(t *testing.T) {
	// Java: partial "happy" from un(happy) tags JJ → postag JJ → keep match (same object).
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "Java returns the same RuleMatch instance on accept")
}

func TestNoDisambiguationEnglishPartialPosTagFilter_InjectedTagDropPOS(t *testing.T) {
	// POS of partial does not match postag_regexp → drop.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationEnglishPartialPosTagFilter_RegexpNoMatchDrops(t *testing.T) {
	// Java matcher.matches() fails → return null (tag never consulted).
	// Single capturing group required; pattern must not match surface "unhappy".
	called := false
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		called = true
		return []string{"JJ"}
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^re(.*)$", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
	require.False(t, called, "tag must not run when surface regexp fails")
}

func TestNoDisambiguationEnglishPartialPosTagFilter_UntaggedPartialDrops(t *testing.T) {
	// tag returns no POS → partialTagHasRequiredTag false.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		return nil
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestNoDisambiguationEnglishPartialPosTagFilter_NegatePos(t *testing.T) {
	// negate_pos present: keep only when no reading matches postag_regexp.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		if p == "run" {
			return []string{"VB", "NN"}
		}
		return nil
	})
	// Has JJ → negate rejects.
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
	// Has VB/NN, none match ^JJ$ → negate accepts.
	ok, err = f.Accept("rerun", "re(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationEnglishPartialPosTagFilter_TwoGroupsAndPrefixSuffix(t *testing.T) {
	// two_groups: group1+group2 retagged; prefix/suffix wrap surface before match.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string {
		if p == "houseboat" {
			return []string{"NN"}
		}
		return nil
	})
	// token "houseboat" with two groups → partial "house"+"boat"
	ok, err := f.Accept("houseboat", "(house)(boat)", "NN", false, true, "", "")
	require.NoError(t, err)
	require.True(t, ok)
	// prefix+token+suffix: prefix "house" + token "boat" → "houseboat", groups on full string
	ok, err = f.Accept("boat", "(house)(boat)", "NN", false, true, "house", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationEnglishPartialPosTagFilter_FailClosedWithoutTagger(t *testing.T) {
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)
	f := NewNoDisambiguationEnglishPartialPosTagFilter(nil)
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationEnglishPartialPosTagFilter_ProcessWideWire(t *testing.T) {
	// Java process-wide Languages.getLanguageForShortCode("en").getTagger() path.
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	filterTagMu.Lock()
	filterTagWord = func(token string) []languagetool.TokenTag {
		if token == "hour" {
			return []languagetool.TokenTag{{POS: "NN", Lemma: "hour"}}
		}
		if token == "happy" {
			return []languagetool.TokenTag{{POS: "JJ", Lemma: "happy"}}
		}
		return nil
	}
	filterTagMu.Unlock()

	f := NewNoDisambiguationEnglishPartialPosTagFilter(nil)
	// disambiguation.xml DIGIT_HYPHENATED_ADJECTIVES-style:
	// 4-hour → extract hour → NN → postag NN(:UN?)? → keep.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("4-hour", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("HYPH"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out)

	// happy is JJ, not NN → drop (filter reject → disambig not applied in Java).
	atr2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "", "", false, false), 0)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atr2}, nil))
}

func TestNoDisambiguationEnglishPartialPosTagFilter_MissingArgsPanic(t *testing.T) {
	// Java RuntimeException when no/regexp/postag_regexp absent.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string { return []string{"JJ"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "un(.*)",
			// postag_regexp missing
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationEnglishPartialPosTagFilter_WrongGroupCountPanic(t *testing.T) {
	// Java: expected 1 group, got 0 → RuntimeException.
	f := NewNoDisambiguationEnglishPartialPosTagFilter(func(p string) []string { return []string{"JJ"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "unhappy", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	// two_groups_regexp present → expect 2 groups.
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
			"two_groups_regexp": "yes",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationEnglishPartialPosTagFilter_Registered(t *testing.T) {
	class := "org.languagetool.rules.en.NoDisambiguationEnglishPartialPosTagFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class))
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	// Factory builds fail-closed instance when tagger unwired.
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationEnglishPartialPosTagFilter_OfficialEnglishDict(t *testing.T) {
	// Prefer real english.dict (same official CFSA2 resource Java EnglishTagger loads).
	root := noDisambigRepoRoot(t)
	pos := filepath.Join(root, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict")
	if st, err := os.Stat(pos); err != nil || st.IsDir() {
		t.Skipf("english.dict missing: %s", pos)
	}
	ClearEnglishFilterTagger()
	require.True(t, WireEnglishFilterTagger(pos))
	t.Cleanup(ClearEnglishFilterTagger)

	f := NewNoDisambiguationEnglishPartialPosTagFilter(nil)

	// DIGIT_HYPHENATED: "4-hour" → partial "hour" is NN* → accept (disambig would add JJ).
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("4-hour", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("HYPH"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "hour from english.dict must match NN(:UN?)?")

	// "500-pages" → partial "pages" is NNS — does not full-match NN(:UN?)? → drop
	// (Java example type=untouched for 500-pages).
	atrPages := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("500-pages", "", "", false, false), 0)
	mPages := rules.NewRuleMatch(rules.NewFakeRule("HYPH"), nil, 0, 9, "msg")
	require.Nil(t, f.AcceptRuleMatch(mPages, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atrPages}, nil))

	// Trademark-style: partial known proper? "Microsoft" + ® — NNP filter.
	// "Microsoft" should tag as NNP in english.dict.
	atrTM := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("Microsoft®", "", "", false, false), 0)
	mTM := rules.NewRuleMatch(rules.NewFakeRule("TM"), nil, 0, 10, "msg")
	outTM := f.AcceptRuleMatch(mTM, map[string]string{
		"no": "1", "regexp": `([a-zA-Z]+)[™®]`, "postag_regexp": "NNP",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrTM}, nil)
	require.Equal(t, mTM, outTM, "Microsoft from english.dict should match NNP")

	// Invented brand not in dict → no NNP → drop (Java rationale for Yarhoo®).
	atrBad := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("Yarhoo®", "", "", false, false), 0)
	mBad := rules.NewRuleMatch(rules.NewFakeRule("TM"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(mBad, map[string]string{
		"no": "1", "regexp": `([a-zA-Z]+)[™®]`, "postag_regexp": "NNP",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrBad}, nil))
}

func noDisambigRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
}
