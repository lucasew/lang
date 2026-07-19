package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestAdverbFilter_Suggest(t *testing.T) {
	f := NewAdverbFilter()
	require.Equal(t, "quick noun", f.Suggest("quickly", "noun"))
	require.Equal(t, "simple noun", f.Suggest("simply", "noun"))
	require.Equal(t, "good noun", f.Suggest("well", "noun"))
	// same form (fast→fast): leave empty
	require.Equal(t, "", f.Suggest("fast", "noun"))
	// unknown adverb
	require.Equal(t, "", f.Suggest("notanadverb", "noun"))
}

func TestAdverbFilter_AcceptRuleMatch(t *testing.T) {
	f := NewAdverbFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"adverb": "quickly", "noun": "run"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"quick run"}, out.GetSuggestedReplacements())

	// unmapped: return match without suggestion rewrite
	m2 := rules.NewRuleMatch(rules.NewFakeRule("A"), nil, 0, 10, "msg")
	m2.SetSuggestedReplacements([]string{"original"})
	out = f.AcceptRuleMatch(m2, map[string]string{"adverb": "xyz", "noun": "n"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"original"}, out.GetSuggestedReplacements())
}

func TestAdverbFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.en.AdverbFilter"))
}
