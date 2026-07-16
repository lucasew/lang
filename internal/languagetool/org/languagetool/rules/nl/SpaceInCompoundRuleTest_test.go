package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SpaceInCompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpaceInCompoundRule_Rule(t *testing.T) {
	rule := NewSpaceInCompoundRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("langeafstandloper"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("lange afstand loper"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("langeafstand loper"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("lange afstandloper"))))
}

func TestSpaceInCompoundRule_Variants(t *testing.T) {
	// Port of generateVariants unit checks from Java twin
	result := map[string]struct{}{}
	GenerateVariants("", []string{"a", "b"}, result)
	// from empty soFar: only joined first path → "ab" not with leading space branch from empty
	// actually GenerateVariants("", [a,b]): only soFar+"a" path → soFar "a" then size1: "a b" only if soFar has space for partial
	// for 2 words [a,b]: soFar="", generateVariants("a",[b]); from "a": size1 → "a b" only (no space in soFar for partial join)
	_, ok := result["a b"]
	require.True(t, ok)
}
