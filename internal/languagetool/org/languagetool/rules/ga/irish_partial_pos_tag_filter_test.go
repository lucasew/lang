package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

// Java-visible outcome matrix for NoDisambiguationIrishPartialPosTagFilter
// + PartialPosTagFilter (tag path without disambiguator).
// Mirrors EN NoDisambiguationEnglishPartialPosTagFilter_test matrix.
// GA surfaces/POS come from official irish.dict via IrishTagger (when real-tagger path).

func TestNoDisambiguationIrishPartialPosTagFilter_InjectedTagAccept(t *testing.T) {
	// Java: partial "fear" from fears tags Noun → postag Noun.* → keep match (same object).
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "Noun:Masc:Com:Sg", "fear", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(fear)s$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "Java returns the same RuleMatch instance on accept")
}

func TestNoDisambiguationIrishPartialPosTagFilter_InjectedTagDropPOS(t *testing.T) {
	// POS of partial does not match postag_regexp → drop.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		return nil
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "Noun:Masc:Com:Sg", "fear", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(fear)s$", "postag_regexp": "Verb.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationIrishPartialPosTagFilter_RegexpNoMatchDrops(t *testing.T) {
	// Java matcher.matches() fails → return null (tag never consulted).
	called := false
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		called = true
		return []string{"Noun:Masc:Com:Sg"}
	})
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "Noun:Masc:Com:Sg", "fear", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(cat)s$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
	require.False(t, called, "tag must not run when surface regexp fails")
}

func TestNoDisambiguationIrishPartialPosTagFilter_UntaggedPartialDrops(t *testing.T) {
	// tag returns no POS → partialTagHasRequiredTag false.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		return nil
	})
	ok, err := f.Accept("fears", "^(fear)s$", "Noun.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestNoDisambiguationIrishPartialPosTagFilter_NegatePos(t *testing.T) {
	// negate_pos present: keep only when no reading matches postag_regexp.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		if p == "rith" {
			return []string{"Verb:VI:Imper:2P:Sg", "Noun:Masc:Com:Sg"}
		}
		return nil
	})
	// Has Noun → negate rejects against Noun.*.
	ok, err := f.Accept("fears", "^(fear)s$", "Noun.*", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
	// Has Verb/Noun, none match ^Adv.*$ → negate accepts.
	ok, err = f.Accept("arith", "a(.*)", "Adv.*", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationIrishPartialPosTagFilter_TwoGroupsAndPrefixSuffix(t *testing.T) {
	// two_groups: group1+group2 retagged; prefix/suffix wrap surface before match.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string {
		if p == "seomradór" {
			return []string{"Noun"}
		}
		return nil
	})
	ok, err := f.Accept("seomradór", "(seomra)(dór)", "Noun", false, true, "", "")
	require.NoError(t, err)
	require.True(t, ok)
	// prefix+token: prefix "seomra" + token "dór" → "seomradór"
	ok, err = f.Accept("dór", "(seomra)(dór)", "Noun", false, true, "seomra", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNoDisambiguationIrishPartialPosTagFilter_FailClosedWithoutDict(t *testing.T) {
	// Injected empty tag path: no POS → drop (simulates missing dict / untagged).
	// Process-wide clear still falls back to IrishTagger when dict is discoverable;
	// invent surface not in dict is the fail-closed outcome under real tagger.
	ClearDefaultIrishPartialPosTagger()
	t.Cleanup(ClearDefaultIrishPartialPosTagger)

	// Explicit inject of empty tags (no invent).
	fInject := NewNoDisambiguationIrishPartialPosTagFilter(func(string) []string { return nil })
	ok, err := fInject.Accept("anseo", "^(anseo)$", "Adv.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)

	// Real path without invent: unknown surface → untagged → drop.
	f := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("xyzzy_not_irish_word", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 20, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(xyzzy_not_irish_word)$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestNoDisambiguationIrishPartialPosTagFilter_ProcessWideWire(t *testing.T) {
	// Java process-wide Irish.getInstance().getTagger() path via WireIrishFilterTaggerFromTagWord.
	ClearDefaultIrishPartialPosTagger()
	t.Cleanup(ClearDefaultIrishPartialPosTagger)

	WireIrishFilterTaggerFromTagWord(func(token string) []languagetool.TokenTag {
		if token == "fear" {
			return []languagetool.TokenTag{{POS: "Noun:Masc:Com:Sg", Lemma: "fear"}}
		}
		if token == "anseo" {
			return []languagetool.TokenTag{{POS: "Adv:Loc", Lemma: "anseo"}}
		}
		return nil
	})

	f := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	// Partial "fear" from "fears" → Noun → keep.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(fear)s$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out)

	// "anseo" is Adv, not Noun → drop under Noun.*.
	atr2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("anseo", "", "", false, false), 0)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{
		"no": "1", "regexp": "^(anseo)$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr2}, nil))
}

func TestNoDisambiguationIrishPartialPosTagFilter_MissingArgsPanic(t *testing.T) {
	// Java RuntimeException when no/regexp/postag_regexp absent.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string { return []string{"Noun"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "Noun", "fear", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "^(fear)s$",
			// postag_regexp missing
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "postag_regexp": "Noun.*",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationIrishPartialPosTagFilter_WrongGroupCountPanic(t *testing.T) {
	// Java: expected 1 group, got 0 → RuntimeException.
	f := NewNoDisambiguationIrishPartialPosTagFilter(func(p string) []string { return []string{"Noun"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("fears", "Noun", "fear", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 5, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "fears", "postag_regexp": "Noun.*",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	// two_groups_regexp present → expect 2 groups.
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "^(fear)s$", "postag_regexp": "Noun.*",
			"two_groups_regexp": "yes",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestNoDisambiguationIrishPartialPosTagFilter_Registered(t *testing.T) {
	class := "org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class))
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	// Factory builds instance; invent surface → untagged → drop.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("xyzzy_not_irish_word", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 20, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(xyzzy_not_irish_word)$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func requireIrishDictForFilter(t *testing.T) {
	t.Helper()
	if taggingga.DiscoverIrishPOSDict() == "" {
		t.Skip("irish.dict not found (LANG_IRISH_DICT / LANG_GA_DICT / /tmp/ga-dict-fetch / third_party)")
	}
	taggingga.EnsureDefaultIrishTagger()
	require.NotNil(t, taggingga.DefaultIrishTagger)
	require.NotEmpty(t, taggingga.IrishPOSDictPath(), "real irish.dict must load")
}

func TestNoDisambiguationIrishPartialPosTagFilter_OfficialIrishDict(t *testing.T) {
	// Real irish.dict (same official CFSA2 resource Java IrishTagger loads).
	requireIrishDictForFilter(t)
	ClearDefaultIrishPartialPosTagger()
	t.Cleanup(ClearDefaultIrishPartialPosTagger)

	// Default nil-tag path uses IrishTagger.Tag (Java Irish.getInstance().getTagger()).
	f := NewNoDisambiguationIrishPartialPosTagFilter(nil)

	// IrishTaggerTest: "anseo" → Adv:Loc only.
	realPOS := tagPOSFromDefaultIrishTagger("anseo")
	require.Contains(t, realPOS, "Adv:Loc", "IrishTagger must tag anseo as Adv:Loc")
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("anseo", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(anseo)$", "postag_regexp": "Adv.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "anseo from irish.dict must match Adv.*")

	// Adv must not keep match under pure Verb postag.
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(anseo)$", "postag_regexp": "Verb:.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))

	// IrishTaggerTest: "t-athair" → Noun:Masc:Com:Sg:DefArt
	athairPOS := tagPOSFromDefaultIrishTagger("t-athair")
	require.Contains(t, athairPOS, "Noun:Masc:Com:Sg:DefArt")
	atrAthair := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("t-athair", "", "", false, false), 0)
	mAthair := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 8, "msg")
	outAthair := f.AcceptRuleMatch(mAthair, map[string]string{
		"no": "1", "regexp": "^(t-athair)$", "postag_regexp": "Noun:Masc:.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrAthair}, nil)
	require.Equal(t, mAthair, outAthair, "t-athair from irish.dict must match Noun:Masc:.*")

	// "bhí" has Verb:VI:PastInd:Len among readings → Verb.* accept.
	bhiPOS := tagPOSFromDefaultIrishTagger("bhí")
	require.Contains(t, bhiPOS, "Verb:VI:PastInd:Len")
	atrBhi := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("bhí", "", "", false, false), 0)
	mBhi := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 3, "msg")
	outBhi := f.AcceptRuleMatch(mBhi, map[string]string{
		"no": "1", "regexp": "^(bhí)$", "postag_regexp": "Verb:.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrBhi}, nil)
	require.Equal(t, mBhi, outBhi, "bhí from irish.dict must match Verb:.*")

	// "sé" has Pron:Pers:… among readings.
	sePOS := tagPOSFromDefaultIrishTagger("sé")
	require.NotEmpty(t, sePOS)
	var hasPron bool
	for _, p := range sePOS {
		if len(p) >= 4 && p[:4] == "Pron" {
			hasPron = true
			break
		}
	}
	require.True(t, hasPron, "sé should have Pron* reading from IrishTagger: %v", sePOS)
	atrSe := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("sé", "", "", false, false), 0)
	mSe := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 2, "msg")
	outSe := f.AcceptRuleMatch(mSe, map[string]string{
		"no": "1", "regexp": "^(sé)$", "postag_regexp": "Pron:.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrSe}, nil)
	require.Equal(t, mSe, outSe, "sé from irish.dict must match Pron:.*")

	// Partial-token extract: surface "x-anseo" → group "anseo" → Adv.* keep.
	atrPref := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("x-anseo", "", "", false, false), 0)
	mPref := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 7, "msg")
	outPref := f.AcceptRuleMatch(mPref, map[string]string{
		"no": "1", "regexp": `^x-(anseo)$`, "postag_regexp": `Adv:.*`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atrPref}, nil)
	require.Equal(t, mPref, outPref, "partial anseo from irish.dict must match Adv:.*")

	// Invented surface not in dict → no POS → drop.
	atrBad := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("xyzzy_not_irish", "", "", false, false), 0)
	mBad := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 15, "msg")
	require.Nil(t, f.AcceptRuleMatch(mBad, map[string]string{
		"no": "1", "regexp": "^(xyzzy_not_irish)$", "postag_regexp": "Noun.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrBad}, nil))
}

func TestNoDisambiguationIrishPartialPosTagFilter_WireIrishFilterTagger(t *testing.T) {
	requireIrishDictForFilter(t)
	ClearDefaultIrishPartialPosTagger()
	t.Cleanup(ClearDefaultIrishPartialPosTagger)

	require.True(t, WireIrishFilterTagger(taggingga.IrishPOSDictPath()))
	// Bad path fails closed.
	require.False(t, WireIrishFilterTagger("/no/such/irish.dict"))

	// Re-wire good path after failed call cleared? Wire on bad path returns false without clearing.
	// Process-wide still set from first Wire.
	f := NewNoDisambiguationIrishPartialPosTagFilter(nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("anseo", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("GA"), nil, 0, 5, "msg")
	require.Equal(t, m, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(anseo)$", "postag_regexp": "Adv:Loc",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))

	// TryWire also succeeds when dict discoverable.
	ClearDefaultIrishPartialPosTagger()
	require.True(t, TryWireIrishFilterTagger())
}

func TestNoDisambiguationIrishPartialPosTagFilter_DefaultFallbackMatchesIrishTagger(t *testing.T) {
	// With process-wide cleared, nil-tag filter POS must equal IrishTagger.Tag POS list.
	requireIrishDictForFilter(t)
	ClearDefaultIrishPartialPosTagger()
	t.Cleanup(ClearDefaultIrishPartialPosTagger)

	for _, surface := range []string{"bhí", "sé", "anseo", "t-athair", "fear", "seomra"} {
		want := tagPOSFromDefaultIrishTagger(surface)
		got := irishNoDisambigTagPOS(surface)
		require.Equal(t, want, got, "surface=%q", surface)
		require.NotEmpty(t, got, "known lexicon surface %q must have POS from irish.dict", surface)
	}
}

// IrishPartialPosTagFilter (disambig path) — only smoke; not this sector's ready claim.
func TestIrishPartialPosTagFilter_Injected(t *testing.T) {
	f := NewIrishPartialPosTagFilter(func(p string) []string {
		if p == "fear" {
			return []string{"Noun:Masc:Com:Sg"}
		}
		return nil
	})
	ok, err := f.Accept("fears", "^(fear)s$", "Noun.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestIrishPartialFiltersRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ga.IrishPartialPosTagFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter"))
}
