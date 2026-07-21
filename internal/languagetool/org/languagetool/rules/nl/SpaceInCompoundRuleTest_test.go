package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SpaceInCompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

// Java AhoCorasick hit.begin/end are UTF-16; multi-byte prefix must not shift FromPos.
func TestSpaceInCompoundRule_UTF16Positions(t *testing.T) {
	rule := NewSpaceInCompoundRule(nil)
	// "é: lange afstand loper" — é is 1 UTF-16 unit, 2 UTF-8 bytes.
	// "lange afstand loper" starts at UTF-16 3 (é : sp), byte index 4.
	text := "é: lange afstand loper"
	ms := rule.Match(languagetool.AnalyzePlain(text))
	require.NotEmpty(t, ms)
	from, to := ms[0].GetFromPos(), ms[0].GetToPos()
	require.Equal(t, 3, from, "FromPos must be UTF-16, not byte offset")
	require.Equal(t, "lange afstand loper", rules.UTF16Substring(text, from, to))
}
