package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Java-visible AcceptRuleMatch outcomes: rewrite only when mapped adj exists and differs from adverb.
func TestAdverbFilter_AcceptRuleMatch(t *testing.T) {
	f := NewAdverbFilter()

	// irregular: well → good
	m := rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"adverb": "well", "noun": "advice"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"good advice"}, out.GetSuggestedReplacements())

	// regular: quickly → quick, simply → simple
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "quickly", "noun": "run"}, 0, nil, nil)
	require.Equal(t, []string{"quick run"}, out.GetSuggestedReplacements())

	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "simply", "noun": "task"}, 0, nil, nil)
	require.Equal(t, []string{"simple task"}, out.GetSuggestedReplacements())

	// identity maps (adj.equals(adverb)): leave suggestions unchanged
	for _, adv := range []string{"fast", "hard", "late", "early", "daily", "straight"} {
		m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
		m.SetSuggestedReplacements([]string{"original " + adv})
		out = f.AcceptRuleMatch(m, map[string]string{"adverb": adv, "noun": "x"}, 0, nil, nil)
		require.Equal(t, []string{"original " + adv}, out.GetSuggestedReplacements(), "identity adverb %q must not rewrite", adv)
	}

	// unmapped adverb: no rewrite
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	m.SetSuggestedReplacements([]string{"original"})
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "notanadverb", "noun": "n"}, 0, nil, nil)
	require.Equal(t, []string{"original"}, out.GetSuggestedReplacements())

	// unduly: commented out in Java (// put("unduly", "undue")) — not in map
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	m.SetSuggestedReplacements([]string{"keep"})
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "unduly", "noun": "influence"}, 0, nil, nil)
	require.Equal(t, []string{"keep"}, out.GetSuggestedReplacements())

	// jokily: Java puts jokey then joking — HashMap final value is joking
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "jokily", "noun": "remark"}, 0, nil, nil)
	require.Equal(t, []string{"joking remark"}, out.GetSuggestedReplacements())

	// special cases from Java comments
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "predominately", "noun": "view"}, 0, nil, nil)
	require.Equal(t, []string{"predominant view"}, out.GetSuggestedReplacements())

	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "likewise", "noun": "case"}, 0, nil, nil)
	require.Equal(t, []string{"similar case"}, out.GetSuggestedReplacements())

	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "twice", "noun": "attempt"}, 0, nil, nil)
	require.Equal(t, []string{"second attempt"}, out.GetSuggestedReplacements())

	// always returns the same match object
	m = rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out = f.AcceptRuleMatch(m, map[string]string{"adverb": "xyz", "noun": "n"}, 0, nil, nil)
	require.Same(t, m, out)
}

func TestAdverbFilter_MapParity(t *testing.T) {
	// Java: 505 active put() calls, 503 unique keys (jokily×2, secondly×2).
	require.Equal(t, 503, len(adverb2Adj))

	// irregular
	require.Equal(t, "good", adverb2Adj["well"])
	require.Equal(t, "fast", adverb2Adj["fast"])
	require.Equal(t, "hard", adverb2Adj["hard"])

	// regular / irregular stems
	require.Equal(t, "quick", adverb2Adj["quickly"])
	require.Equal(t, "simple", adverb2Adj["simply"])
	require.Equal(t, "easy", adverb2Adj["easily"])

	// HashMap overwrite: second put("jokily","joking") wins over "jokey"
	require.Equal(t, "joking", adverb2Adj["jokily"])

	// commented in Java — must not exist
	_, hasUnduly := adverb2Adj["unduly"]
	require.False(t, hasUnduly)
	_, hasEmpty := adverb2Adj[""]
	require.False(t, hasEmpty)

	// special cases
	require.Equal(t, "predominant", adverb2Adj["predominately"])
	require.Equal(t, "similar", adverb2Adj["likewise"])
	require.Equal(t, "below", adverb2Adj["beneath"])
	require.Equal(t, "least", adverb2Adj["lest"])
	require.Equal(t, "other", adverb2Adj["either"])
	require.Equal(t, "second", adverb2Adj["twice"])
	require.Equal(t, "intentional", adverb2Adj["purposely"])
	require.Equal(t, "partial", adverb2Adj["partly"])
}

func TestAdverbFilter_Registered(t *testing.T) {
	// Java grammar.xml uses class="org.languagetool.rules.en.AdverbFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.AdverbFilter"))
	rf := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.en.AdverbFilter")
	require.NotNil(t, rf)
	_, ok := rf.(*AdverbFilter)
	require.True(t, ok)
}
