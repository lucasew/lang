package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCaseRuleAntiPatterns_CountAndImmunizeSmoke(t *testing.T) {
	// Java has 316 inner lists; we port 317/317 after string-aware conversion (or 316 if outer miscount).
	n := CaseRuleAntiPatternsCount()
	require.GreaterOrEqual(t, n, 300, "anti-patterns should be mostly ported")
	// Smoke: immunization path must not panic on RE2-unsupported patterns.
	rule := NewCaseRule(nil)
	_ = rule.Match(languagetool.AnalyzePlain("Das Laufen fällt mir leicht."))
	_ = rule.Match(languagetool.AnalyzePlain("Heute spricht Frau Stieg."))
}

func TestLanguageNamesGetAsRegex(t *testing.T) {
	re := LanguageNamesGetAsRegex()
	require.Contains(t, re, "Deutsch")
	require.Contains(t, re, "Englisch")
	require.Contains(t, re, "Weißrussisch")        // restored from Java
	require.NotContains(t, re, "Schweizerdeutsch") // not in Java list
	require.True(t, IsLanguageName("Plattdeutsch"))
}
