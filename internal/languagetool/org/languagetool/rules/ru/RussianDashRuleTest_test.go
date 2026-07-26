package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianDashRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java: JLanguageTool(ru), rule.match(lt.getAnalyzedSentence(text)).
// AbstractDashRule.match uses sentence text only — AnalyzePlain + rule.Match is the
// established twin pattern (same outcomes as JVM).
func TestRussianDashRule_Rule(t *testing.T) {
	rule := NewRussianDashRule(nil)
	check := func(expectedErrors int, text string, expectedSuggestion string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if expectedSuggestion != "" && len(matches) > 0 {
			require.Equal(t, "Использовано тире вместо дефиса.", matches[0].GetMessage())
			require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
			require.Equal(t, expectedSuggestion, matches[0].GetSuggestedReplacements()[0])
		}
	}

	// correct sentences (Java RussianDashRuleTest):
	check(0, "Он вышел из-за забора.", "")
	check(0, "Ростов-на-Дону.", "")
	check(0, "ведром — работай", "")
	// incorrect sentences:
	check(1, "из—за", "из-за")
	check(1, "Ростов — на — Дону", "Ростов-на-Дону")
}

// Java-visible metadata: id, description, message, empty tags (clears picky), TYPOGRAPHY;
// compounds from /ru/compounds.txt (no URL / no example pair in Java).
func TestRussianDashRule_Metadata(t *testing.T) {
	rule := NewRussianDashRule(nil)
	require.Equal(t, "RU_DASH_RULE", rule.GetID())
	require.Equal(t, "Тире вместо дефиса («из — за» вместо «из-за»).", rule.GetDescription())
	require.Equal(t, "Использовано тире вместо дефиса.", rule.GetMessage())
	require.Equal(t, 2, rule.EstimateContextForSureMatch())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOGRAPHY", rule.GetCategory().GetID().String())
	// Java: setTags(Collections.emptyList()) — not picky (unlike AbstractDashRule default).
	require.Empty(t, rule.GetTags())
	require.False(t, rule.HasTag(rules.TagPicky))
	require.Empty(t, rule.GetURL())
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
	// Official resource path Java loads (embed twin of compounds.txt).
	require.Equal(t, "/ru/compounds.txt", compoundsResourcePath)
	require.NotEmpty(t, loadDashPatterns())
}

// Java isBoundary PATTERN = [\u0400-\u04FF] exactly (Cyrillic block; no invent Latin/Unicode letter).
func TestRussianDashRule_BoundaryPattern(t *testing.T) {
	// Representative letters in U+0400..U+04FF (isCyrillicLetter true → not boundary).
	inPattern := []rune{
		'А', 'я', 'Ё', 'ё', // common Russian
		0x0400, 0x04FF, // block endpoints
		'і', 'ї', 'ў', // other Cyrillic (still in block)
	}
	for _, r := range inPattern {
		require.True(t, isCyrillicLetter(r), "rune %U %q should be letter per Java PATTERN", r, r)
	}
	// Explicit non-matches that a soft IsLetter invent would wrongly accept:
	outPattern := []rune{'a', 'Z', 'Á', 'é', ' ', '—', '–', '-', '1', '.', 0x03FF, 0x0500}
	for _, r := range outPattern {
		require.False(t, isCyrillicLetter(r), "rune %U %q must not be letter (Java PATTERN)", r, r)
	}
}
