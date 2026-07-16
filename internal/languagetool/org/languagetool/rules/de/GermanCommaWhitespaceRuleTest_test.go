package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewGermanCommaWhitespaceRule(nil)
	// Java: "Es gibt 5 Millionen .de-Domains." — space before . is OK for domain label
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Es gibt 5 Millionen .de-Domains."))))
	// normal space-before-comma still flagged
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("Die Partei , die die letzte Wahl gewann."))))
}
