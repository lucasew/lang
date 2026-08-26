package en

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

// Java-visible outcome matrix for EnglishPartialPosTagFilter
// (tagger + single-token disambiguator). Contrast NoDisambiguation which is tagger-only.

func TestEnglishPartialPosTagFilter_InjectedTagAccept(t *testing.T) {
	// Injected tag path bypasses process-wide wires (tests PartialPosTagFilter + factory).
	// Java: partial "happy" from un(happy) tags JJ → keep same RuleMatch instance.
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
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

func TestEnglishPartialPosTagFilter_InjectedTagDropPOS(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
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

func TestEnglishPartialPosTagFilter_RegexpNoMatchDrops(t *testing.T) {
	// Java matcher.matches() fails → return null (tag never consulted).
	called := false
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
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

func TestEnglishPartialPosTagFilter_UntaggedPartialDrops(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
		return nil
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestEnglishPartialPosTagFilter_NegatePos(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
		if p == "happy" {
			return []string{"JJ"}
		}
		if p == "run" {
			return []string{"VB", "NN"}
		}
		return nil
	})
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = f.Accept("rerun", "re(.*)", "JJ", true, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEnglishPartialPosTagFilter_TwoGroupsAndPrefixSuffix(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string {
		if p == "houseboat" {
			return []string{"NN"}
		}
		return nil
	})
	ok, err := f.Accept("houseboat", "(house)(boat)", "NN", false, true, "", "")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = f.Accept("boat", "(house)(boat)", "NN", false, true, "house", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEnglishPartialPosTagFilter_FailClosedWithoutBothWires(t *testing.T) {
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	f := NewEnglishPartialPosTagFilter(nil)
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok, "no tagger and no disambig → fail closed")

	// Tagger only — still fail closed (Java always runs both).
	filterTagMu.Lock()
	filterTagWord = func(token string) []languagetool.TokenTag {
		if token == "happy" {
			return []languagetool.TokenTag{{POS: "JJ", Lemma: "happy"}}
		}
		return nil
	}
	filterTagMu.Unlock()
	ok, err = f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok, "tagger without disambiguator must not invent")

	// Disambiguator only — fail closed.
	ClearEnglishFilterTagger()
	WireEnglishFilterDisambiguator(passThroughDisambig{})
	ok, err = f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok, "disambiguator without tagger must not invent")

	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestEnglishPartialPosTagFilter_ProcessWideWireTagAndDisambig(t *testing.T) {
	// Java process-wide Languages.getLanguageForShortCode("en").getTagger() + getDisambiguator().
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	filterTagMu.Lock()
	filterTagWord = func(token string) []languagetool.TokenTag {
		if token == "run" {
			return []languagetool.TokenTag{
				{POS: "VB", Lemma: "run"},
				{POS: "NN", Lemma: "run"},
			}
		}
		if token == "happy" {
			return []languagetool.TokenTag{{POS: "JJ", Lemma: "happy"}}
		}
		return nil
	}
	filterTagMu.Unlock()
	WireEnglishFilterDisambiguator(passThroughDisambig{})

	f := NewEnglishPartialPosTagFilter(nil)

	// Accept path: keep same match when any disambiguated POS matches.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("running", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(run).*", "postag_regexp": "VB.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out)

	// Drop when postag does not match any reading.
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "^(run).*", "postag_regexp": "JJ.*",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))

	// happy is JJ → accept JJ, drop NN.
	atr2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "", "", false, false), 0)
	m2 := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Equal(t, m2, f.AcceptRuleMatch(m2, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr2}, nil))
	require.Nil(t, f.AcceptRuleMatch(m2, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "NN",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr2}, nil))
}

func TestEnglishPartialPosTagFilter_DisambiguatorStripsReadings(t *testing.T) {
	// Disambiguator must affect accept/drop: keep only NN from {VB,NN} → VB.* drops.
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	filterTagMu.Lock()
	filterTagWord = func(token string) []languagetool.TokenTag {
		if token == "run" {
			return []languagetool.TokenTag{
				{POS: "VB", Lemma: "run"},
				{POS: "NN", Lemma: "run"},
			}
		}
		return nil
	}
	filterTagMu.Unlock()
	WireEnglishFilterDisambiguator(keepOnlyPOSDisambig{"NN"})

	f := NewEnglishPartialPosTagFilter(nil)
	// After disambig: only NN → VB.* fails, NN succeeds.
	ok, err := f.Accept("running", "^(run).*", "VB.*", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok, "disambiguator stripped VB; VB.* must drop")
	ok, err = f.Accept("running", "^(run).*", "NN", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok, "disambiguator kept NN")

	// Contrast: pass-through would keep VB.
	WireEnglishFilterDisambiguator(passThroughDisambig{})
	ok, err = f.Accept("running", "^(run).*", "VB.*", false, false, "", "")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestEnglishPartialPosTagFilter_NilDisambiguateFailsClosed(t *testing.T) {
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)

	filterTagMu.Lock()
	filterTagWord = func(token string) []languagetool.TokenTag {
		return []languagetool.TokenTag{{POS: "JJ", Lemma: "happy"}}
	}
	filterTagMu.Unlock()
	WireEnglishFilterDisambiguator(nilDisambig{})

	f := NewEnglishPartialPosTagFilter(nil)
	ok, err := f.Accept("unhappy", "un(.*)", "JJ", false, false, "", "")
	require.NoError(t, err)
	require.False(t, ok, "nil Disambiguate result must not fall back to pre-disambig POS")
}

func TestEnglishPartialPosTagFilter_MissingArgsPanic(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string { return []string{"JJ"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "un(.*)",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestEnglishPartialPosTagFilter_WrongGroupCountPanic(t *testing.T) {
	f := NewEnglishPartialPosTagFilter(func(p string) []string { return []string{"JJ"} })
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "unhappy", "postag_regexp": "JJ",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
			"two_groups_regexp": "yes",
		}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	})
}

func TestEnglishPartialPosTagFilter_Registered(t *testing.T) {
	class := "org.languagetool.rules.en.EnglishPartialPosTagFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class))
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
	// Factory builds fail-closed instance when wires absent.
	ClearEnglishFilterTagger()
	t.Cleanup(ClearEnglishFilterTagger)
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("unhappy", "JJ", "happy", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("P"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "un(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil))
}

func TestEnglishPartialPosTagFilter_OfficialEnglishDictAndHybrid(t *testing.T) {
	// Prefer real english.dict + DefaultEnglishHybridDisambiguator (official resources).
	root := enPartialRepoRoot(t)
	pos := filepath.Join(root, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict")
	if st, err := os.Stat(pos); err != nil || st.IsDir() {
		t.Skipf("english.dict missing: %s", pos)
	}
	hybrid := tagen.DefaultEnglishHybridDisambiguator()
	if hybrid == nil {
		t.Skip("DefaultEnglishHybridDisambiguator unavailable")
	}
	// Need at least one stage so we exercise real EN disambiguator (not empty shell).
	if hybrid.GlobalChunker == nil && hybrid.Chunker == nil && hybrid.RulesDisambiguator == nil {
		t.Skip("EN hybrid has no official stages loaded")
	}

	ClearEnglishFilterTagger()
	require.True(t, WireEnglishFilterTagger(pos))
	WireEnglishFilterDisambiguator(hybrid)
	t.Cleanup(ClearEnglishFilterTagger)

	f := NewEnglishPartialPosTagFilter(nil)

	// Same surface outcomes as NoDisambiguation on single-token path when
	// disambig does not strip: "4-hour" → partial "hour" is NN* → accept.
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("4-hour", "", "", false, false), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("HYPH"), nil, 0, 6, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil)
	require.Equal(t, m, out, "hour from english.dict+hybrid must match NN(:UN?)?")

	// "500-pages" → partial "pages" is NNS — does not full-match NN(:UN?)? → drop.
	atrPages := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("500-pages", "", "", false, false), 0)
	mPages := rules.NewRuleMatch(rules.NewFakeRule("HYPH"), nil, 0, 9, "msg")
	require.Nil(t, f.AcceptRuleMatch(mPages, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atrPages}, nil))

	// Trademark-style: Microsoft® → NNP.
	atrTM := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("Microsoft®", "", "", false, false), 0)
	mTM := rules.NewRuleMatch(rules.NewFakeRule("TM"), nil, 0, 10, "msg")
	outTM := f.AcceptRuleMatch(mTM, map[string]string{
		"no": "1", "regexp": `([a-zA-Z]+)[™®]`, "postag_regexp": "NNP",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrTM}, nil)
	require.Equal(t, mTM, outTM, "Microsoft from english.dict should match NNP")

	// Invented brand not in dict → drop.
	atrBad := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedTokenStr("Yarhoo®", "", "", false, false), 0)
	mBad := rules.NewRuleMatch(rules.NewFakeRule("TM"), nil, 0, 7, "msg")
	require.Nil(t, f.AcceptRuleMatch(mBad, map[string]string{
		"no": "1", "regexp": `([a-zA-Z]+)[™®]`, "postag_regexp": "NNP",
	}, 0, []*languagetool.AnalyzedTokenReadings{atrBad}, nil))

	// Without disambiguator wire (tagger still present via re-clear of disambig only):
	ClearEnglishFilterDisambiguator()
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": `[0-9]+-([a-z]+|[A-Z]+)`, "postag_regexp": `NN(:UN?)?`,
	}, 0, []*languagetool.AnalyzedTokenReadings{atr}, nil),
		"official dict alone must not invent disambiguated accept")
}

// --- disambiguator stubs for process-wide wire tests ---

type passThroughDisambig struct{}

func (passThroughDisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return s
}

type nilDisambig struct{}

func (nilDisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return nil
}

// keepOnlyPOSDisambig keeps readings whose POS equals keep (Java-style strip other tags).
type keepOnlyPOSDisambig struct{ keep string }

func (d keepOnlyPOSDisambig) Disambiguate(s *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if s == nil {
		return nil
	}
	toks := s.GetTokens()
	out := make([]*languagetool.AnalyzedTokenReadings, len(toks))
	for i, t := range toks {
		if t == nil {
			continue
		}
		var kept []*languagetool.AnalyzedToken
		for _, r := range t.GetReadings() {
			if r == nil {
				continue
			}
			if p := r.GetPOSTag(); p != nil && *p == d.keep {
				kept = append(kept, r)
			}
		}
		if len(kept) == 0 {
			// Untagged residual (surface preserved).
			out[i] = languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(t.GetToken(), nil, nil))
		} else {
			out[i] = languagetool.NewAnalyzedTokenReadingsList(kept, t.GetStartPos())
		}
	}
	return languagetool.NewAnalyzedSentence(out)
}

func enPartialRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", ".."))
}
