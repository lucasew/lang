package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseRuleExceptions(t *testing.T) {
	ex := CaseRuleExceptions()
	require.NotEmpty(t, ex)
	// case_rule_exceptions.txt
	require.True(t, IsCaseRuleException("Absolut Vodka"))
	// eigennamen_gross.txt (Java loadExceptions first path)
	require.True(t, IsCaseRuleException("Europäische Zentralbank"))
	require.True(t, IsCaseRuleException("Heilige Vater"))
	require.False(t, IsCaseRuleException("xyzzy not an exception"))
}

// TestCaseRuleExceptionsWordsActiveOnly guards CaseRule.exceptions active literals:
// Java comments must not become set members (validator invent finding).
func TestCaseRuleExceptionsWordsActiveOnly(t *testing.T) {
	// Unique active tokens from Java CaseRule.exceptions (// and /* */ stripped).
	require.Equal(t, 629, len(caseRuleExceptionsWords))

	// Active tokens present.
	for _, w := range []string{"Milk", "De", "Le", "Letzt", "Nachts", "Diesseits", "Übrigen", "Unvorhergesehenes", "Wider", "Erster", "Erste", "Ähnliches", "Familienangehörige", "Mo", "Di", "Fr", "Gr", "Mag"} {
		_, ok := caseRuleExceptionsWords[w]
		require.True(t, ok, "missing active exception %q", w)
	}

	// Comment-only / illustration invents must stay out.
	for _, w := range []string{
		"Leichter", // fully commented in Java
		"melken",   // only in comment on "Milk"
		"Di. und Do. um 18 Uhr",
		"Fr. Dr. Müller",
		"Gr. 12",
		"Mag. Helke Müller",
		"De Morgan",
		"Le Monde",
		"zu guter Letzt",
		"des Nachts",
		"eines Nachts",
		"im Diesseits",
		"im Übrigen",
		"etwas Unvorhergesehenes",
		"das Für und Wider",
		"er wurde Erster im Langlauf",
		"sie wurde Erste im Langlauf",
		"Er hat Ähnliches erlebt",
		"Brüder und solche Familienangehörige, die...",
	} {
		_, ok := caseRuleExceptionsWords[w]
		require.False(t, ok, "invent/comment token must not be loaded: %q", w)
	}
}

// TestCaseNumeralsENCaseSensitive ports CaseRule.NUMERALS_EN (no CASE_INSENSITIVE).
func TestCaseNumeralsENCaseSensitive(t *testing.T) {
	require.True(t, caseNumeralsEN.MatchString("a"))
	require.True(t, caseNumeralsEN.MatchString("12"))
	require.True(t, caseNumeralsEN.MatchString("xiv"))
	require.False(t, caseNumeralsEN.MatchString("A"), "Java NUMERALS_EN is case-sensitive")
	require.False(t, caseNumeralsEN.MatchString("XIV"))
}
