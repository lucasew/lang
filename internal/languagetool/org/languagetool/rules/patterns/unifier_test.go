package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnifierCase(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("case-sensitivity", "lowercase", NewPatternToken(`\p{Ll}+`, true, true, false))
	cfg.SetEquivalence("case-sensitivity", "uppercase", NewPatternToken(`\p{Lu}\p{Ll}+`, true, true, false))
	cfg.SetEquivalence("case-sensitivity", "alluppercase", NewPatternToken(`\p{Lu}+$`, true, true, false))

	lower1 := languagetool.NewAnalyzedToken("lower", strPtr("JJR"), strPtr("lower"))
	lower2 := languagetool.NewAnalyzedToken("lowercase", strPtr("JJ"), strPtr("lowercase"))
	upper1 := languagetool.NewAnalyzedToken("Uppercase", strPtr("JJ"), strPtr("Uppercase"))
	upper2 := languagetool.NewAnalyzedToken("John", strPtr("NNP"), strPtr("John"))

	uni := cfg.CreateUnifier()
	equiv := map[string][]string{"case-sensitivity": {"lowercase"}}

	satisfied := uni.IsSatisfied(lower1, equiv)
	satisfied = satisfied && uni.IsSatisfied(lower2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(lower2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	equiv = map[string][]string{"case-sensitivity": {"uppercase"}}
	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(upper1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
}

func strPtr(s string) *string { return &s }
