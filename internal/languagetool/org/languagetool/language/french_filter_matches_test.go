package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterFrenchRuleMatches_AdjacentMergeSameITS(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_FR_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_FR_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"foobar"}, out[0].Suggestions)
	require.Equal(t, "Il pourrait y avoir un problème ici.", out[0].Message)
	require.Equal(t, "Erreur potentielle", out[0].ShortMessage)
}

func TestFilterFrenchRuleMatches_ChainMergeThree(t *testing.T) {
	// Java keeps rule id AI_FR_GGEC* after merge (specificRuleId MERGED); chain continues.
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_FR_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_FR_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}},
		{FromPos: 6, ToPos: 9, RuleID: "AI_FR_GGEC_C", IssueType: "grammar", Suggestions: []string{"c"}},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"abc"}, out[0].Suggestions)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 9, out[0].ToPos)
}

func TestFilterFrenchRuleMatches_MergeBothStylePicky(t *testing.T) {
	// Java: AI_FR_MERGED_MATCH_STYLE_PICKY when both Style and both picky.
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_FR_GGEC_A", IssueType: "style", Suggestions: []string{"a"}, IsPicky: true},
		{FromPos: 3, ToPos: 6, RuleID: "AI_FR_GGEC_B", IssueType: "style", Suggestions: []string{"b"}, IsPicky: true},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_MERGED_MATCH_STYLE_PICKY", out[0].RuleID)
	require.Equal(t, "style", out[0].IssueType)
	require.True(t, out[0].IsPicky)
}

func TestFilterFrenchRuleMatches_DropTrailingPeriod(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5, RuleID: "AI_FR_GGEC_MISSING_PUNCTUATION_PERIOD",
			Suggestions: []string{"Hallo."}, SentenceText: "Hallo  ",
		},
		{FromPos: 6, ToPos: 10, RuleID: "FR_OTHER"},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "FR_OTHER", out[0].RuleID)
}

func TestFilterFrenchRuleMatches_AdjustCasingOnly(t *testing.T) {
	// Java adjustFrenchRuleMatch: ignore-case equal suggestion → CASING rewrite.
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 5, RuleID: "AI_FR_GGEC_REPLACEMENT_ORTHOGRAPHY_X",
			Suggestions: []string{"Paris"}, OriginalErrorStr: "paris",
		},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_GGEC_REPLACEMENT_CASING_X", out[0].RuleID)
	require.Equal(t, "typographical", out[0].IssueType)
	require.Equal(t, "CASING", out[0].CategoryID)
	require.Equal(t, "Majuscules et minuscules", out[0].ShortMessage)
}

func TestFilterFrenchRuleMatches_AdjustQuotesPicky(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 1, RuleID: "AI_FR_GGEC_REPLACEMENT_PUNCTUATION_QUOTE_X", Suggestions: []string{"«"}},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_GGEC_QUOTES", out[0].RuleID)
	require.True(t, out[0].IsPicky)
	require.Equal(t, "typographical", out[0].IssueType)
}

func TestFilterFrenchRuleMatches_SiLonPicky(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			FromPos: 3, ToPos: 5, RuleID: "AI_FR_GGEC_MISSING_PRONOUN_LAPOSTROPHE",
			Suggestions: []string{"l'on"}, OriginalErrorStr: "on",
			SentenceText: "si on veut",
		},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_GGEC_SI_LON", out[0].RuleID)
	require.True(t, out[0].IsPicky)
}

func TestFilterFrenchRuleMatches_EmptySuggestionsMerge(t *testing.T) {
	// Java NPE on get(0); Go fail-closed empty suggestions, still merge.
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_FR_GGEC_A", IssueType: "grammar"},
		{FromPos: 3, ToPos: 6, RuleID: "AI_FR_GGEC_B", IssueType: "grammar"},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_MERGED_MATCH", out[0].RuleID)
	require.Empty(t, out[0].Suggestions)
	require.Equal(t, "GRAMMAR", out[0].CategoryID)
}

func TestFilterFrenchRuleMatches_NoMergeDifferentPicky(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_FR_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}, IsPicky: true},
		{FromPos: 3, ToPos: 6, RuleID: "AI_FR_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}, IsPicky: false},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 2)
}

func TestFilterFrenchRuleMatches_AposTyp(t *testing.T) {
	// Java: enabledRules.contains("APOS_TYP") → replace ' with ’ in suggestions len>1
	en := map[string]struct{}{"APOS_TYP": {}}
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 4, RuleID: "SOME_FR",
			Suggestions:  []string{"l'eau", "a"},
			EnabledRules: en,
		},
	}
	out := FilterFrenchRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, []string{"l’eau", "a"}, out[0].Suggestions)
}

func TestFilterFrenchRuleMatches_AposTypOff(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 4, RuleID: "SOME_FR", Suggestions: []string{"l'eau"}},
	}
	out := FilterFrenchRuleMatches(in)
	require.Equal(t, []string{"l'eau"}, out[0].Suggestions)
}
