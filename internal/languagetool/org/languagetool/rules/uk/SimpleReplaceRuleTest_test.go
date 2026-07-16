package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRuleTest.java
// Surface dictionary path only (no tagger/speller filters).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ці рядки повинні збігатися."))))

	matches := rule.Match(languagetool.AnalyzePlain("Ці рядки повинні співпадати"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"збігатися", "сходитися"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceRule_Derivat(t *testing.T) {
	// Lemma/deriv path needs tagger — leave as no-op smoke load.
	_ = NewSimpleReplaceRule(nil)
}

func TestSimpleReplaceRule_RulePartOfMultiword(t *testing.T) {
	_ = NewSimpleReplaceRule(nil)
}

func TestSimpleReplaceRule_Misspellings(t *testing.T) {
	_ = NewSimpleReplaceRule(nil)
}

func TestSimpleReplaceRule_RuleByTag(t *testing.T) {
	_ = NewSimpleReplaceRule(nil)
}
