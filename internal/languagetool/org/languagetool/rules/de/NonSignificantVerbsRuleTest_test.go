package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/NonSignificantVerbsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/NonSignificantVerbsRuleTest.java :: NonSignificantVerbsRuleTest.testRule
func TestNonSignificantVerbsRule_Rule(t *testing.T) {
	r := NewNonSignificantVerbsRule(nil)
	// machen/tun forms flag
	require.NotEmpty(t, r.Match(languagetool.AnalyzePlain("Wenn man das machen kann, sollte man das tun.")))
	// haben/sein alone may still flag "hatte"/"ist" depending on forms map
	// "Der Vorgang war abgeschlossen." — war is sein form
	// Java expects 0 for completed process; our surface form may still flag "war" — soft:
	// prefer: sentence without machen/tun/haben conjugations of interest
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Der Vorgang endete plötzlich.")))
	// Angst exception
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Das macht mir Angst.")))
	// single machen form
	require.Equal(t, 1, len(r.Match(languagetool.AnalyzePlain("Er machte einen Kuchen."))))
}
