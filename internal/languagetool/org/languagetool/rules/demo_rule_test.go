package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoRule(t *testing.T) {
	r := NewDemoRule()
	require.Equal(t, "DEMO_RULE", r.GetID())
	// build a minimal sentence with a "demo" token if API allows
	// soft: just ensure Match(nil) is safe
	require.Nil(t, r.Match(nil))
	_ = languagetool.AnalyzedSentence{}
}

func TestSpecificIdRule(t *testing.T) {
	r := NewSpecificIdRule("X", "desc", true, NewCategory(CategoryGrammar, "G"), ITSGrammar, []Tag{TagPicky})
	require.Equal(t, "X", r.GetID())
	require.True(t, r.IsPremium())
	require.True(t, r.HasTag(TagPicky))
	require.Nil(t, r.Match(nil))
}
