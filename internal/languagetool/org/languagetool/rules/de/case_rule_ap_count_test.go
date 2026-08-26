package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseRuleAntiPatterns_CountAndImmunizeSmoke(t *testing.T) {
	// CaseRuleAntiPatterns.ANTI_PATTERNS inner lists (Go twin of Java static field).
	n := CaseRuleAntiPatternsCount()
	require.Equal(t, 317, n, "ANTI_PATTERNS list count must match ported CaseRuleAntiPatterns")
	// Smoke: immunization path must not panic.
	rule, lt := newCaseRuleFixture(t)
	_ = rule.Match(lt.GetAnalyzedSentence("Das Laufen fällt mir leicht."))
	_ = rule.Match(lt.GetAnalyzedSentence("Heute spricht Frau Stieg."))
}

func TestLanguageNamesGetAsRegex(t *testing.T) {
	re := LanguageNamesGetAsRegex()
	require.Contains(t, re, "Deutsch")
	require.Contains(t, re, "Englisch")
	require.Contains(t, re, "Weißrussisch")        // restored from Java
	require.NotContains(t, re, "Schweizerdeutsch") // not in Java list
	require.True(t, IsLanguageName("Plattdeutsch"))
}
