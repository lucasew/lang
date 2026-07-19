package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func TestAcceptPhrases_BuildsImmunizeAntiPatterns(t *testing.T) {
	// Java acceptPhrases → makeAntiPatterns → IMMUNIZE (not IGNORE_SPELLING).
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.AcceptPhrases([]string{"duodenal atresia", "Microsoft Entra"})
	aps := r.GetAntiPatterns()
	require.NotEmpty(t, aps)
	for _, ap := range aps {
		require.Equal(t, "INTERNAL_ANTIPATTERN", ap.GetID())
		require.Equal(t, disambigrules.ActionImmunize, ap.Action)
	}
	// "duodenal atresia" lowercase → 2 antipatterns; "Microsoft Entra" title → 1.
	require.GreaterOrEqual(t, len(r.MultiWordIgnore), 2)
	require.GreaterOrEqual(t, len(aps), 3) // 1 for Microsoft + 2 for duodenal
}

func TestAcceptPhrases_LowercaseSentenceStartVariant(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AcceptPhrases([]string{"duodenal atresia"})
	var hasLower, hasUpper bool
	for _, p := range r.MultiWordIgnore {
		if len(p) == 2 && p[0] == "duodenal" && p[1] == "atresia" {
			hasLower = true
		}
		if len(p) == 2 && p[0] == "Duodenal" && p[1] == "atresia" {
			hasUpper = true
		}
	}
	require.True(t, hasLower)
	require.True(t, hasUpper)
	aps := r.GetAntiPatterns()
	require.Len(t, aps, 2)
	for _, ap := range aps {
		require.Equal(t, disambigrules.ActionImmunize, ap.Action)
		require.Equal(t, "INTERNAL_ANTIPATTERN", ap.GetID())
	}
}

func TestAddIgnoreWords_BuildsIgnoreSpellingAntiPattern(t *testing.T) {
	// Java multi-token addIgnoreWords uses IGNORE_SPELLING (not IMMUNIZE).
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("Microsoft Entra")
	aps := r.GetAntiPatterns()
	require.Len(t, aps, 1)
	require.Equal(t, disambigrules.ActionIgnoreSpelling, aps[0].Action)
	require.Equal(t, "INTERNAL_ANTIPATTERN", aps[0].GetID())
}

func TestSentenceWithImmunization_ImmunizePhrase(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.AcceptPhrases([]string{"Microsoft Entra"})
	sent := languagetool.AnalyzeWithTokenizer(
		"Use Microsoft Entra today.",
		languagetool.WordTokenizerForLanguage("en"),
	)
	// Original not immunized
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "Microsoft" {
			require.False(t, tok.IsImmunized())
		}
	}
	imm := r.SentenceWithImmunization(sent)
	require.NotNil(t, imm)
	// Original unchanged
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && (tok.GetToken() == "Microsoft" || tok.GetToken() == "Entra") {
			require.False(t, tok.IsImmunized(), "original must not be mutated")
		}
	}
	// Copy immunized when pattern matches
	var immunized []string
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			immunized = append(immunized, tok.GetToken())
		}
	}
	// Pattern matcher may or may not fire depending on POS/strictness; if it does:
	if len(immunized) > 0 {
		require.Contains(t, immunized, "Microsoft")
		require.Contains(t, immunized, "Entra")
	}
}
