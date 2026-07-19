package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFilterGermanRuleMatches_OverlapSkip(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "AI_DE_GGEC_A", Suggestions: []string{"a"}},
		{FromPos: 3, ToPos: 8, RuleID: "AI_DE_GGEC_B", Suggestions: []string{"b"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_GGEC_A", out[0].RuleID)
}

func TestFilterGermanRuleMatches_AdjacentMergeSameITS(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 6, out[0].ToPos)
	require.Equal(t, []string{"foobar"}, out[0].Suggestions)
	// Java German.mergeMatches: message + shortMessage on the merged RuleMatch
	require.Equal(t, "Hier scheint es einen Fehler zu geben.", out[0].Message)
	require.Equal(t, "Potenzieller Fehler", out[0].ShortMessage)
}

func TestFilterGermanRuleMatches_AdjacentGapMerge(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"foo bar"}, out[0].Suggestions)
}

func TestFilterGermanRuleMatches_NonGGECPassthrough(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 2, RuleID: "DE_AGREEMENT"},
		{FromPos: 3, ToPos: 5, RuleID: "GERMAN_SPELLER_RULE"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 2)
	require.Equal(t, "DE_AGREEMENT", out[0].RuleID)
	require.Equal(t, "GERMAN_SPELLER_RULE", out[1].RuleID)
}

func TestFilterGermanRuleMatches_DropTrailingPeriodSuggestion(t *testing.T) {
	// Java: sentence "Hallo" + suggestion "Hallo." → drop AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD.
	in := []languagetool.LocalMatch{
		{
			FromPos:      0,
			ToPos:        5,
			RuleID:       "AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD",
			Suggestions:  []string{"Hallo."},
			SentenceText: "Hallo  ",
		},
		{FromPos: 6, ToPos: 10, RuleID: "DE_AGREEMENT"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "DE_AGREEMENT", out[0].RuleID)
}

func TestFilterGermanRuleMatches_KeepPeriodWithoutSentenceText(t *testing.T) {
	// Fail-closed: no SentenceText → do not invent drop.
	in := []languagetool.LocalMatch{
		{
			RuleID:      "AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD",
			Suggestions: []string{"Hallo."},
		},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
}

func TestFilterGermanRuleMatches_KeepPeriodWhenNotSentenceEnd(t *testing.T) {
	in := []languagetool.LocalMatch{
		{
			RuleID:       "AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD",
			Suggestions:  []string{"Welt."},
			SentenceText: "Hallo Welt ist gross",
		},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
}

func TestFilterGermanRuleMatches_MergeOriginalErrorStr(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}, OriginalErrorStr: "aaa"},
		{FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}, OriginalErrorStr: "bbb"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "aaa bbb", out[0].OriginalErrorStr)
}

func TestFilterGermanRuleMatches_MergeOriginalFromSentenceText(t *testing.T) {
	// Java getOriginalErrorStr empty → sentence substring via positions.
	in := []languagetool.LocalMatch{
		{
			FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar",
			Suggestions: []string{"foo"}, SentenceText: "aaa bbb",
			FromPosSentence: 0, ToPosSentence: 3,
		},
		{
			FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar",
			Suggestions: []string{"bar"}, SentenceText: "aaa bbb",
			FromPosSentence: 4, ToPosSentence: 7,
		},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "aaa bbb", out[0].OriginalErrorStr)
	require.Equal(t, 0, out[0].FromPosSentence)
	require.Equal(t, 7, out[0].ToPosSentence)
}

func TestFilterGermanRuleMatches_NoMergeDifferentPicky(t *testing.T) {
	// Java: merge only when both share Tag.picky status.
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}, IsPicky: true},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}, IsPicky: false},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 2)
	require.Equal(t, "AI_DE_GGEC_A", out[0].RuleID)
	require.Equal(t, "AI_DE_GGEC_B", out[1].RuleID)
}

// Java: same ITS Style → merge and keep Style (mergeMatches keeps ITS when equal).
func TestFilterGermanRuleMatches_AdjacentMergeBothStyle(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "style", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "style", Suggestions: []string{"bar"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, "style", out[0].IssueType)
	require.Equal(t, "Potenzieller Fehler", out[0].ShortMessage)
	require.Equal(t, []string{"foobar"}, out[0].Suggestions)
}

// Java: different ITS, neither Style → merge as Grammar.
func TestFilterGermanRuleMatches_AdjacentMergeDifferentNonStyleITS(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "misspelling", Suggestions: []string{"bar"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, "grammar", out[0].IssueType)
	require.Equal(t, "Potenzieller Fehler", out[0].ShortMessage)
}

// Java: Style + non-Style adjacent → do not merge (only same ITS, or neither Style).
func TestFilterGermanRuleMatches_NoMergeStyleWithGrammar(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "style", Suggestions: []string{"foo"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 2)
	require.Equal(t, "AI_DE_GGEC_A", out[0].RuleID)
	require.Equal(t, "AI_DE_GGEC_B", out[1].RuleID)
}

// Java mergeMatches calls getSuggestedReplacements().get(0) (NPE if empty).
// Go fail-closed: still merge span/message; omit suggestion when both sides empty.
func TestFilterGermanRuleMatches_MergeBothEmptySuggestions(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar"},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 6, out[0].ToPos)
	require.Empty(t, out[0].Suggestions)
	require.Equal(t, "Hier scheint es einen Fehler zu geben.", out[0].Message)
	require.Equal(t, "Potenzieller Fehler", out[0].ShortMessage)
}

// One side empty: keep the non-empty first suggestion (defensive vs Java NPE).
func TestFilterGermanRuleMatches_MergeOneEmptySuggestion(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"foo"}},
		{FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"foo"}, out[0].Suggestions)

	in2 := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar"},
		{FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"bar"}},
	}
	out2 := FilterGermanRuleMatches(in2)
	require.Len(t, out2, 1)
	require.Equal(t, []string{"bar"}, out2[0].Suggestions)
}

// Java filterRuleMatches loop: three adjacent same-ITS GGEC → one merged span.
// Relies on merge keeping GGEC-merge eligibility (Java: rule id stays AI_DE_GGEC*,
// only specificRuleId becomes AI_DE_MERGED_MATCH).
func TestFilterGermanRuleMatches_ChainMergeThree(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}},
		{FromPos: 6, ToPos: 9, RuleID: "AI_DE_GGEC_C", IssueType: "grammar", Suggestions: []string{"c"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 9, out[0].ToPos)
	require.Equal(t, []string{"abc"}, out[0].Suggestions)
}

// Gap separators accumulate: ToPos+1 between each pair → spaces between suggestions.
func TestFilterGermanRuleMatches_ChainMergeThreeWithGaps(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}, OriginalErrorStr: "xxx"},
		{FromPos: 4, ToPos: 7, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}, OriginalErrorStr: "yyy"},
		{FromPos: 8, ToPos: 11, RuleID: "AI_DE_GGEC_C", IssueType: "grammar", Suggestions: []string{"c"}, OriginalErrorStr: "zzz"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "AI_DE_MERGED_MATCH", out[0].RuleID)
	require.Equal(t, []string{"a b c"}, out[0].Suggestions)
	require.Equal(t, "xxx yyy zzz", out[0].OriginalErrorStr)
	require.Equal(t, 0, out[0].FromPos)
	require.Equal(t, 11, out[0].ToPos)
}

// Merged match keeps picky flag from the pair (Java Tag.picky equality is a merge gate).
func TestFilterGermanRuleMatches_MergeKeepsPicky(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}, IsPicky: true},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}, IsPicky: true},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.True(t, out[0].IsPicky)
}

// RuleMeta for AI_DE_MERGED_MATCH fills empty category (not invent for unknown ids).
func TestFilterGermanRuleMatches_MergeSoftRuleCategoryWhenEmpty(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"}},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"}},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "GRAMMAR", out[0].CategoryID)
	require.Equal(t, "Grammatik", out[0].CategoryName)
	// RuleDescription for AI_DE_MERGED_MATCH matches Java merge message.
	require.Equal(t, "Hier scheint es einen Fehler zu geben.", out[0].Description)
}

// When match1 already has CategoryID, keep it (Java keeps match1 rule category).
func TestFilterGermanRuleMatches_MergeKeepsCategoryFromMatch1(t *testing.T) {
	in := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 3, RuleID: "AI_DE_GGEC_A", IssueType: "grammar", Suggestions: []string{"a"},
			CategoryID: "STYLE", CategoryName: "Stil"},
		{FromPos: 3, ToPos: 6, RuleID: "AI_DE_GGEC_B", IssueType: "grammar", Suggestions: []string{"b"},
			CategoryID: "GRAMMAR", CategoryName: "Grammatik"},
	}
	out := FilterGermanRuleMatches(in)
	require.Len(t, out, 1)
	require.Equal(t, "STYLE", out[0].CategoryID)
	require.Equal(t, "Stil", out[0].CategoryName)
}
