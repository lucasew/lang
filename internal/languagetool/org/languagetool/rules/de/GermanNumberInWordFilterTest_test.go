package de

// Twin of GermanNumberInWordFilterTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Twin of GermanNumberInWordFilterTest.testFilter / runFilter cases.
// Java: AbstractNumberInWordFilter filter = new GermanNumberInWordFilter();
// always wires GermanyGerman.getDefaultSpellingRule() — no inject / no FilterDict wire.
func TestGermanNumberInWordFilter_Filter(t *testing.T) {
	f := NewGermanNumberInWordFilter()
	// Java runFilter cases (assertNotNull + suggestedReplacements.contains):
	//   "Schöne Grüß0e aus Potsdam." arg Grüß0e → Grüße
	//   "Schöne Grü0e aus Potsdam."  arg Grü0e  → Grüße
	//   "Der Aut0r schreibt ein Buch." arg Aut0r → Autor
	for _, tc := range []struct {
		arg  string
		want string
	}{
		{"Grüß0e", "Grüße"},
		{"Grü0e", "Grüße"},
		{"Aut0r", "Autor"},
	} {
		// Java: RuleMatch match = new RuleMatch(new FakeRule(), sentence, fromPos, toPos, "fake msg");
		// fromPos=0 toPos=5 patternTokenPos=2 for all three Java cases.
		m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "fake msg")
		out := f.AcceptRuleMatch(m, map[string]string{"word": tc.arg}, 2, nil, nil)
		require.NotNil(t, out, tc.arg)
		require.Contains(t, out.GetSuggestedReplacements(), tc.want, tc.arg)
	}
}

func TestGermanNumberInWordFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.de.GermanNumberInWordFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	f := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.NotNil(t, f)
}
