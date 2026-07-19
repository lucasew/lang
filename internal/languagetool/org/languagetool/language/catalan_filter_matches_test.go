package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterCatalanRuleMatches_IgnoreProperNounsDropsMorfologik(t *testing.T) {
	// IGNORE_PROPER_NOUNS after MORFOLOGIK at same FromPos removes both
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "MORFOLOGIK_RULE_CA_ES"},
		{FromPos: 0, ToPos: 5, RuleID: "IGNORE_PROPER_NOUNS"},
		{FromPos: 6, ToPos: 8, RuleID: "OTHER"},
	}
	out := FilterCatalanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].RuleID)
}

func TestFilterCatalanRuleMatches_IgnoreProperNounsDropsFollowingMorfologik(t *testing.T) {
	// IGNORE first sets ignore pos; following MORFOLOGIK same pos dropped
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "IGNORE_PROPER_NOUNS"},
		{FromPos: 0, ToPos: 5, RuleID: "MORFOLOGIK_RULE_CA_ES"},
		{FromPos: 6, ToPos: 8, RuleID: "OTHER"},
	}
	out := FilterCatalanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].RuleID)
}

func TestFilterCatalanRuleMatches_EmptySuggestionExpandSpace(t *testing.T) {
	// Java: empty suggestion with spaces both sides of span → extend ToPos by 1.
	// Sentence " a " — match "a" at [1,2); char before and at toSent are spaces.
	in := []languagetool.LocalMatch{
		{
			FromPos: 1, ToPos: 2, RuleID: "X", Suggestions: []string{""},
			SentenceText: " a ", FromPosSentence: 1, ToPosSentence: 2,
		},
	}
	out := FilterCatalanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, 3, out[0].ToPos)
	require.Equal(t, 3, out[0].ToPosSentence)
	require.Equal(t, []string{""}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatchesAfterOverlapping_TrimAndSort(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 10, ToPos: 17, OriginalErrorStr: "foo bar",
			Suggestions: []string{"baz bar"},
		},
		{
			FromPos: 0, ToPos: 3, OriginalErrorStr: "xyz",
			Suggestions: []string{"abc"},
		},
	}
	out := FilterCatalanRuleMatchesAfterOverlapping(in)
	require.Len(t, out, 2)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 10, out[1].FromPos)
	require.Equal(t, 13, out[1].ToPos) // trimmed "foo"
	require.Equal(t, []string{"baz"}, out[1].Suggestions)
}

func TestFilterCatalanRuleMatches_FaltaElementSkip(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "FALTA_ELEMENT_ENTRE_VERBS[3]", FromPosSentence: 0, ToPosSentence: 5},
		{FromPos: 10, ToPos: 12, RuleID: "OTHER", FromPosSentence: 10, ToPosSentence: 12}, // delta 5 < 20
	}
	out := FilterCatalanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "OTHER", out[0].RuleID)
}

func TestFilterCatalanRuleMatches_AposTipografic(t *testing.T) {
	en := map[string]struct{}{"APOSTROF_TIPOGRAFIC": {}}
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "X", Suggestions: []string{"l'home"}, EnabledRules: en},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"l’home"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_ExigeixPossessiusU(t *testing.T) {
	en := map[string]struct{}{"EXIGEIX_POSSESSIUS_U": {}}
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "X", Suggestions: []string{"meva feina"}, EnabledRules: en},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"meua faena"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_ExigeixAccentGeneralSkipsEAcute(t *testing.T) {
	// When both é and è forms present, drop é form under EXIGEIX_ACCENTUACIO_GENERAL.
	en := map[string]struct{}{"EXIGEIX_ACCENTUACIO_GENERAL": {}}
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 4, RuleID: "X",
			Suggestions:  []string{"café", "cafè"},
			EnabledRules: en,
		},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"cafè"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_TraditionalDiacriticsKeptWhenEnabled(t *testing.T) {
	en := map[string]struct{}{"DIACRITICS_TRADITIONAL_RULES": {}}
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "X", Suggestions: []string{"sóc"}, EnabledRules: en},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"sóc"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_OldDiacriticsStrippedByDefault(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "X", Suggestions: []string{"sóc"}},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"soc"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_TypographicApostropheInSentence(t *testing.T) {
	// Java: hasTypographicApostrophe from tokens OR APOSTROF_TIPOGRAFIC
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 4, RuleID: "X", Suggestions: []string{"l'home"},
			HasTypographicApostropheInSentence: true,
		},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"l’home"}, out[0].Suggestions)
}

func TestFilterCatalanRuleMatches_AposRecteBlocksTypographic(t *testing.T) {
	en := map[string]struct{}{"APOSTROF_TIPOGRAFIC": {}, "APOSTROF_RECTE": {}}
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "X", Suggestions: []string{"l'home"}, EnabledRules: en},
	}
	out := FilterCatalanRuleMatches(in)
	require.Equal(t, []string{"l'home"}, out[0].Suggestions)
}
