package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UnnecessaryPhraseRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUnnecessaryPhraseRule_Rule(t *testing.T) {
	rule := NewUnnecessaryPhraseRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Das ist allem Anschein nach eine Phrase."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das ist eine Phrase."))))
}
