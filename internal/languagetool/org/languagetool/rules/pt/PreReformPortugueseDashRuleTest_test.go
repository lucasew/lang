package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PreReformPortugueseDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java: JLanguageTool(pt-AO), only PT_PREAO_DASH_RULE, PICKY; AnnotatedText checks.
// AbstractDashRule.match uses sentence text only — AnalyzePlain + rule.Match is the
// established twin pattern (same outcomes as JVM).
func TestPreReformPortugueseDashRule_Test(t *testing.T) {
	rule := NewPreReformPortugueseDashRule(nil)
	// correct (hyphen): no match
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("abaixa-língua"))))
	// incorrect (em dash): one match
	matches := rule.Match(languagetool.AnalyzePlain("abaixa—língua"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Um travessão foi utilizado em vez de um hífen.", matches[0].GetMessage())
	require.Equal(t, []string{"abaixa-língua"}, matches[0].GetSuggestedReplacements())
}

// Java-visible metadata: id, description, message, Typographical, TYPOGRAPHY, picky;
// compounds from /pt/pre-reform-compounds.txt (no URL / no example pair in Java).
func TestPreReformPortugueseDashRule_Metadata(t *testing.T) {
	rule := NewPreReformPortugueseDashRule(nil)
	require.Equal(t, "PT_PREAO_DASH_RULE", rule.GetID())
	require.Equal(t, "Travessões no lugar de hífens", rule.GetDescription())
	require.Equal(t, "Um travessão foi utilizado em vez de um hífen.", rule.GetMessage())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	require.Equal(t, 2, rule.EstimateContextForSureMatch())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOGRAPHY", rule.GetCategory().GetID().String())
	require.True(t, rule.HasTag(rules.TagPicky))
	require.Empty(t, rule.GetURL())
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
	// Official resource path Java loads (embed twin of pre-reform-compounds.txt).
	require.Equal(t, "/pt/pre-reform-compounds.txt", preReformCompoundsResourcePath)
	require.NotEmpty(t, loadPreReformDashPatterns())
}

// Java isBoundary PATTERN = [a-zA-ZÂâÃãÇçÊêÓóÔôÕõü] exactly (shared isPortugueseLetter).
func TestPreReformPortugueseDashRule_BoundaryPattern(t *testing.T) {
	inPattern := []rune{
		'a', 'z', 'A', 'Z',
		'Â', 'â', 'Ã', 'ã', 'Ç', 'ç', 'Ê', 'ê', 'Ó', 'ó', 'Ô', 'ô', 'Õ', 'õ', 'ü',
	}
	for _, r := range inPattern {
		require.True(t, isPortugueseLetter(r), "rune %q should be letter per Java PATTERN", r)
	}
	// Explicit non-matches that a soft Latin/IsLetter invent would wrongly accept:
	outPattern := []rune{'Á', 'á', 'À', 'à', 'É', 'é', 'Í', 'í', 'Ú', 'ú', 'È', 'è', 'Ü', ' ', '—', '1'}
	for _, r := range outPattern {
		require.False(t, isPortugueseLetter(r), "rune %q must not be letter (Java PATTERN)", r)
	}
}
