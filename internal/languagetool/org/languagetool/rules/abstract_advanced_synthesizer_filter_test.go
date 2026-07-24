package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractAdvancedSynthesizerFilter(t *testing.T) {
	// Java lemmaSelect/postagSelect are POS-tag regexes (not lemma surfaces).
	f := &AbstractAdvancedSynthesizerFilter{
		Synthesize: func(lemma, postag string) []string {
			if lemma == "go" && postag == "VBG" {
				return []string{"going"}
			}
			return nil
		},
	}
	lemma := "go"
	pos1 := "VB"
	pos2 := "VBG"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("go", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("running", &pos2, nil))
	t2.SetStartPos(3)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "VBG",
	}, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	// Empty existing suggestions → Java addAll(raw replacements), no case invent.
	require.Equal(t, []string{"going"}, out.GetSuggestedReplacements())
}

func TestAbstractAdvancedSynthesizerFilter_EmptySuggestionsNoCaseInvent(t *testing.T) {
	// Java: empty getSuggestedReplacements → !suggestionUsed → addAll(raw) without uppercase.
	f := &AbstractAdvancedSynthesizerFilter{
		Synthesize: func(lemma, postag string) []string {
			return []string{"going"}
		},
	}
	lemma := "go"
	pos1 := "VB"
	pos2 := "VBG"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Go", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("running", &pos2, nil))
	t2.SetStartPos(3)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB.*", "postagSelect": "VBG",
	}, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"going"}, out.GetSuggestedReplacements(),
		"must not invent capitalized Going when suggestions list is empty")
}

func TestAbstractAdvancedSynthesizerFilter_PlaceholderAppliesCase(t *testing.T) {
	f := &AbstractAdvancedSynthesizerFilter{
		Synthesize: func(lemma, postag string) []string {
			return []string{"going"}
		},
	}
	lemma := "go"
	pos1 := "VB"
	pos2 := "VBG"
	t1 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("Go", &pos1, &lemma))
	t1.SetStartPos(0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("running", &pos2, nil))
	t2.SetStartPos(3)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 2, "msg")
	m.SetSuggestedReplacements([]string{"try {suggestion}"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "VBG",
	}, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"try Going"}, out.GetSuggestedReplacements())
}

func TestAbstractAdvancedSynthesizerFilter_SelectPOSNotLemma(t *testing.T) {
	// Multi-reading: lemmaSelect matches POS only; invent lemma-path would pick wrong reading.
	f := &AbstractAdvancedSynthesizerFilter{
		Synthesize: func(lemma, postag string) []string {
			return []string{lemma + ":" + postag}
		},
	}
	lemA, lemB := "wrong", "right"
	posA, posB, posSel := "NN", "VB", "JJ"
	// readings: first has lemma "wrong"/NN, second has lemma "right"/VB — select VB
	t1 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("x", &posA, &lemA),
		languagetool.NewAnalyzedToken("x", &posB, &lemB),
	}, 0)
	t2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("y", &posSel, nil))
	t2.SetStartPos(2)
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 1, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "VB", "postagSelect": "JJ",
	}, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"right:JJ"}, out.GetSuggestedReplacements())
}

func TestAbstractAdvancedSynthesizerFilter_PostagReplace(t *testing.T) {
	// \a1 from lemmaSelect on originalPostag, \b1 from postagSelect on desiredPostag
	got := GetCompositePostag(`(V).`, `(N)N`, "VB", "NN", `\a1X\b1`)
	require.Equal(t, "VXN", got)
}

func TestAbstractAdvancedSynthesizerFilter_PostagReplaceCaseSensitive(t *testing.T) {
	// Java UNICODE_CASE without CASE_INSENSITIVE → case-sensitive; do not invent (?i).
	got := GetCompositePostag(`(v).`, `(N)N`, "VB", "NN", `\a1X\b1`)
	require.Equal(t, `\a1X\b1`, got, "lowercase pattern must not match VB")
}

func TestAbstractAdvancedSynthesizerFilter_NoSynth(t *testing.T) {
	f := &AbstractAdvancedSynthesizerFilter{}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x", "postagSelect": "N",
	}, nil, nil))
}
