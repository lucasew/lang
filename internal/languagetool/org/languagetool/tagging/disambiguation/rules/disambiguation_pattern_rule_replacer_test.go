package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDisambiguationPatternRuleReplacer(t *testing.T) {
	// empty replacer is identity
	r := NewDisambiguationPatternRuleReplacer(nil)
	sent := languagetool.AnalyzePlain("hello")
	require.Equal(t, sent.GetText(), r.Replace(sent).GetText())

	// construct a rule that won't match (safe no-op)
	tok := patterns.Token("zzz")
	rule := NewDisambiguationPatternRule("X", "d", "en", []*patterns.PatternToken{tok}, "NN", nil, ActionFilter)
	r2 := NewDisambiguationPatternRuleReplacer([]*DisambiguationPatternRule{rule})
	out := r2.Replace(sent)
	require.NotNil(t, out)
}
