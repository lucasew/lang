package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.NewLineFilterTest (NewLineMatchFilter).

// TestNewLineFilter_Filter ports NewLineFilterTest.testFilter exactly.
func TestNewLineFilter_Filter(t *testing.T) {
	filter := NewNewLineMatchFilter()
	// Java: var text1 = "I’m+the+Heeadline\n\n\n\n\u2063\n\n\n\nI’m+some+plain+teext.";
	// curly apostrophe U+2019
	text1 := "I’m+the+Heeadline\n\n\n\n\u2063\n\n\n\nI’m+some+plain+teext."

	// case1: from=8,to=21, suggestion "Headline\n\n\n\n\u2063\n\n\n\n"
	// → keep, from=8,to=17, contains "Headline"
	runNewLineCase(t, filter, text1, 8, 21, "Headline\n\n\n\n\u2063\n\n\n\n", 8, 17, true, "Headline")

	// case2: suggestion "Heeadline\n\n\n\n\u2063\n\n\n\n" → drop (empty filtered; equals trimmed match)
	runNewLineCase(t, filter, text1, 8, 21, "Heeadline\n\n\n\n\u2063\n\n\n\n", 8, 17, false, "")
}

// TestNewLineFilter_OutOfBoundsKeep: Java keeps matches when originalText.length() < from|to.
func TestNewLineFilter_OutOfBoundsKeep(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "short"
	// from/to beyond UTF-16 length → keep unchanged (no mutation of positions/suggestions)
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 10, 20, "match1")
	rm.SetSuggestedReplacement("anything")
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 10, filtered[0].GetFromPos())
	require.Equal(t, 20, filtered[0].GetToPos())
	require.Equal(t, []string{"anything"}, filtered[0].GetSuggestedReplacements())
}

// TestNewLineFilter_OnlyNewlinesDrop: match that is only newlines/invis with a
// suggestion that also trims to "" → drop via size==1 && equals(matchText).
// Note: Java's `to < from` branch is unreachable for well-formed from<=to spans
// (after L=to-from trims, to==from and endsWith("") is false). Pure-newline
// no-op drops go through the equals path instead.
func TestNewLineFilter_OnlyNewlinesDrop(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "aa\n\n\u2063\nbb"
	// span covers only \n\n\u2063\n at positions 2..6 → trims to ""
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 2, 6, "match1")
	rm.SetSuggestedReplacement("\n\u2063\n") // also trims to ""
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Empty(t, filtered)
}

// TestNewLineFilter_OnlyNewlinesNonEmptySuggestion: pure-newline span with a
// non-empty suggestion after trim keeps an empty span (from==to); Java does not drop.
func TestNewLineFilter_OnlyNewlinesNonEmptySuggestion(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "aa\n\n\u2063\nbb"
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 2, 6, "match1")
	rm.SetSuggestedReplacement("x")
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 2, filtered[0].GetFromPos())
	require.Equal(t, 2, filtered[0].GetToPos()) // empty span after trim
	require.Equal(t, []string{"x"}, filtered[0].GetSuggestedReplacements())
}

// TestNewLineFilter_MultiSuggestionKeep: Java requires size==1 && equals to drop;
// multiple suggestions → keep even if first equals matchText after trim.
func TestNewLineFilter_MultiSuggestionKeep(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "Heeadline\n\n"
	// match "Heeadline\n\n" → trim to "Heeadline" [0,9)
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 0, 11, "match1")
	// first equals trimmed matchText, second is different — size != 1 so keep
	rm.SetSuggestedReplacements([]string{"Heeadline\n\n", "Headline\n"})
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 0, filtered[0].GetFromPos())
	require.Equal(t, 9, filtered[0].GetToPos())
	require.Equal(t, []string{"Heeadline", "Headline"}, filtered[0].GetSuggestedReplacements())
}

// TestNewLineFilter_LeadingAndTrailingTrim: trims both ends of span and suggestions.
func TestNewLineFilter_LeadingAndTrailingTrim(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "xx\n\u2063Heeadline\n\n\u2063yy"
	// positions: 0:x 1:x 2:\n 3:\u2063 4:H ... 12:e 13:\n 14:\n 15:\u2063 16:y 17:y
	// match from=2 to=16 → "\n\u2063Heeadline\n\n\u2063"
	// trailing trim → to moves to 13 ("\n\u2063Heeadline"); leading trim → from to 4 ("Heeadline")
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 2, 16, "match1")
	rm.SetSuggestedReplacement("\n\u2063Headline\n\u2063")
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 4, filtered[0].GetFromPos())
	require.Equal(t, 13, filtered[0].GetToPos())
	require.Equal(t, []string{"Headline"}, filtered[0].GetSuggestedReplacements())
	// matched text after filter
	matched := utf16Substring(text, filtered[0].GetFromPos(), filtered[0].GetToPos())
	require.Equal(t, "Heeadline", matched)
}

// TestNewLineFilter_AlreadyCleanEqualDrop: no trim needed but single suggestion equals match → drop.
func TestNewLineFilter_AlreadyCleanEqualDrop(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "Heeadline is bad"
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 0, 9, "match1")
	rm.SetSuggestedReplacement("Heeadline") // equals matchText, size==1 → drop
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Empty(t, filtered)
}

// TestNewLineFilter_AlreadyCleanKeep: clean span, different single suggestion → keep as-is.
func TestNewLineFilter_AlreadyCleanKeep(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "Heeadline is bad"
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 0, 9, "match1")
	rm.SetSuggestedReplacement("Headline")
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 0, filtered[0].GetFromPos())
	require.Equal(t, 9, filtered[0].GetToPos())
	require.Equal(t, []string{"Headline"}, filtered[0].GetSuggestedReplacements())
}

// TestNewLineFilter_EmptySuggestionsKeep: no suggestions → size!=1, keep after trim.
func TestNewLineFilter_EmptySuggestionsKeep(t *testing.T) {
	filter := NewNewLineMatchFilter()
	text := "Heeadline\n"
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, 0, 10, "match1")
	// no suggestions
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	require.Len(t, filtered, 1)
	require.Equal(t, 0, filtered[0].GetFromPos())
	require.Equal(t, 9, filtered[0].GetToPos())
	require.Empty(t, filtered[0].GetSuggestedReplacements())
}

func runNewLineCase(t *testing.T, filter *NewLineMatchFilter, text string,
	fromBefore, toBefore int, suggestionBefore string,
	fromAfter, toAfter int, expectMatch bool, suggestionAfter string) {
	t.Helper()
	// Java: var sentence = lt.analyzeText(text).get(0);
	//       var ruleMatch = new RuleMatch(new FakeRule(), sentence, from, to, "match1");
	//       ruleMatch.setSuggestedReplacement(suggestionBeforeFilter);
	//       filter.filter(Collections.singletonList(ruleMatch), new AnnotatedTextBuilder().addText(text).build());
	sentence := languagetool.AnalyzePlain(text)
	rm := NewRuleMatch(NewFakeRule(""), sentence, fromBefore, toBefore, "match1")
	rm.SetSuggestedReplacement(suggestionBefore)
	filtered := filter.Filter([]*RuleMatch{rm}, text)
	if expectMatch {
		require.Len(t, filtered, 1)
		m := filtered[0]
		require.Equal(t, fromAfter, m.GetFromPos())
		require.Equal(t, toAfter, m.GetToPos())
		require.Contains(t, m.GetSuggestedReplacements(), suggestionAfter)
	} else {
		require.Empty(t, filtered)
	}
}
