package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractAdvancedSynthesizerFilter(t *testing.T) {
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
		"lemmaFrom": "1", "postagFrom": "2", "lemmaSelect": "go", "postagSelect": "VBG",
	}, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"going"}, out.GetSuggestedReplacements())
}

func TestAbstractAdvancedSynthesizerFilter_PostagReplace(t *testing.T) {
	// \a1 from lemmaSelect on originalPostag, \b1 from postagSelect on desiredPostag
	got := getCompositePostag(`(V).`, `(N)N`, "VB", "NN", `\a1X\b1`)
	require.Equal(t, "VXN", got)
}

func TestAbstractAdvancedSynthesizerFilter_NoSynth(t *testing.T) {
	f := &AbstractAdvancedSynthesizerFilter{}
	m := NewRuleMatch(NewFakeRule("R"), nil, 0, 1, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom": "1", "postagFrom": "1", "lemmaSelect": "x", "postagSelect": "N",
	}, nil, nil))
}
