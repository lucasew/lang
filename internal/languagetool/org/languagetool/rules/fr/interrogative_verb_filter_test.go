package fr

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestInterrogativeVerbFilter_DesiredPostag(t *testing.T) {
	f := NewInterrogativeVerbFilter()
	require.Contains(t, f.DesiredPostagForPronoun("je"), "1 s")
	require.Contains(t, f.DesiredPostagForPronoun("tu"), "imp")
	require.Contains(t, f.DesiredPostagForPronoun("ils"), "3 p")
	require.Empty(t, f.DesiredPostagForPronoun("xyz"))
}

func TestInterrogativeVerbFilter_DesiredPostagFromPOS(t *testing.T) {
	tag := "R pers suj 1 s"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("je", &tag, nil), 0)
	require.Contains(t, DesiredPostagFromPronounPOS(tok), "1 s")
	tag2 := "R pers obj 2 p"
	tok2 := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("vous", &tag2, nil), 0)
	require.Contains(t, DesiredPostagFromPronounPOS(tok2), "imp")
}

func TestInterrogativeVerbFilter_AcceptWithSpeller(t *testing.T) {
	f := NewInterrogativeVerbFilter()
	f.SpellingSuggestions = func(wrong string) []string {
		return []string{"manges", "mange"}
	}
	f.MatchesDesiredPostag = func(cand, re string) bool {
		return cand == "manges"
	}
	ptag := "R pers suj 2 s"
	vtag := "V ind pres 3 s"
	pron := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("tu", &ptag, nil), 5)
	verb := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("mange", &vtag, nil), 0)
	m := rules.NewRuleMatch(nil, nil, 0, 8, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"PronounFrom": "2", "VerbFrom": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{verb, pron}, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "manges-tu")
}

func TestInterrogativeVerbFilter_MakeWrong(t *testing.T) {
	require.Equal(t, "mänge", MakeWrong("mange")) // first a → ä
	require.True(t, strings.HasSuffix(MakeWrong("xyz"), "-"))
}

func TestInterrogativeVerbFilterRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.InterrogativeVerbFilter"))
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.WordWithDeterminerFilter"))
}

func TestInterrogativeVerbFilter_Filter(t *testing.T) {
	f := NewInterrogativeVerbFilter()
	got := f.FilterByDesiredPOS([]string{"mange", "manges", "mangeons"}, "2", func(form, re string) bool {
		return strings.HasSuffix(form, "es")
	})
	require.Equal(t, []string{"manges"}, got)
}
