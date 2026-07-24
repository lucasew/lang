package en

// Twin of EnglishNumberInWordFilterTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Twin of EnglishNumberInWordFilterTest.testFilter:
// input "Go0d morning, Potsdam.", arg "Go0d" → suggested replacements contain "Good"
// (acceptRuleMatch with FakeRule match fromPos=0 toPos=4, patternTokenPos=2).
func TestEnglishNumberInWordFilter_Filter(t *testing.T) {
	// Java: AbstractNumberInWordFilter filter = new EnglishNumberInWordFilter();
	// always wires MorfologikAmericanSpellerRule (AmericanEnglish) — no inject / no FilterDict wire.
	f := NewEnglishNumberInWordFilter()
	// Java: RuleMatch match = new RuleMatch(new FakeRule(), sentence, fromPos, toPos, "fake msg");
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 4, "fake msg")
	// Java: args.put("word", arg); filter.acceptRuleMatch(match, args, patternTokenPos, tokens, tokenPositions)
	out := f.AcceptRuleMatch(m, map[string]string{"word": "Go0d"}, 2, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "Good")
}

// Sanity: 0→o path uses real speller (House known in en_US).
func TestEnglishNumberInWordFilter_House(t *testing.T) {
	f := NewEnglishNumberInWordFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"word": "H0use"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "House")
}
