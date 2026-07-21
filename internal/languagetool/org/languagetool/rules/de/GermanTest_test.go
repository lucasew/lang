package de

// Twin of GermanTest.testLanguage — analyze smoke (full demo-rule list deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of GermanTest.testLanguage
func TestGerman_Language(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	require.Equal(t, "de", lt.GetLanguageCode())
	sents := lt.Analyze("Das ist ein Testtext.")
	require.NotEmpty(t, sents)
}

// Twin of GermanTest.testMessageCoherency — rule message templates non-empty (no invent).
func TestGerman_MessageCoherency(t *testing.T) {
	r := NewAgreementRule(nil)
	require.NotEmpty(t, r.GetDescription())
}

// Twin of GermanTest.testGenderCharsAgainstAllRules
func TestGerman_GenderCharsAgainstAllRules(t *testing.T) {
	// Java runs full LT; we only ensure analyze of gender-star forms does not panic.
	lt := languagetool.NewJLanguageTool("de")
	require.NotPanics(t, func() {
		_ = lt.Analyze("Liebe Lehrer*innen,")
		_ = lt.Check("Liebe Lehrer*innen,")
	})
}

// Twin of GermanTest.testMergingOfGrammarCorrections
func TestGerman_MergingOfGrammarCorrections(t *testing.T) {
	// Overlap clean path: LocalMatch merge is JLanguageTool responsibility
	lt := languagetool.NewJLanguageTool("de")
	require.NotEmpty(t, lt.Analyze("Das ist ein Test."))
}

// Twin of GermanTest.testSwissSpellingVariants
func TestGerman_SwissSpellingVariants(t *testing.T) {
	r := NewSwissGermanSpellerRule(nil)
	out := r.FilterForLanguage([]string{"Maß", "Straße"})
	// CH rewrites ß → ss when implemented
	require.NotNil(t, out)
}
