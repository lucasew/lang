package ga

// Metadata + isMisspelled control-flow twins for MorfologikIrishSpellerRule.
// No map-inject / AddWord invent — real binary dict only for Match outcomes
// (see MorfologikIrishSpellerRuleTest_test.go).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

// Metadata twin: getId / getFileName / ignoreWordsWithLength / tokenizingPattern / examples.
func TestMorfologikIrishSpellerRule(t *testing.T) {
	r := NewMorfologikIrishSpellerRule()
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", MorfologikIrishSpellerRuleID)
	require.Equal(t, "/ga/hunspell/ga_IE.dict", IrishSpellerDict)
	require.Equal(t, MorfologikIrishSpellerRuleID, r.GetID())
	require.Equal(t, IrishSpellerDict, r.GetFileName())
	// Java MorfologikIrishSpellerRule: super.ignoreWordsWithLength = 1
	require.Equal(t, 1, r.IgnoreWordsWithLength)
	// Java tokenizingPattern(): Pattern.compile("-")
	require.NotNil(t, r.TokenizingPattern)
	require.True(t, r.TokenizingPattern.MatchString("-"))
	// Java addExamplePair: botun → botún
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	// When official binary is present, InitSpellersFromGetters attaches multi.
	if morfologik.DiscoverLanguageDict(IrishSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

// Ports isMisspelled override control flow (Utils maths normalize then super).
// Pure predicate stands in for super — not a map-inject Match golden.
func TestIrishIsMisspelled_MathsNormalize(t *testing.T) {
	r := NewMorfologikIrishSpellerRule()
	// bold mathematical "seanathair"
	bold := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.True(t, taggingga.IsAllMathsChars(bold))
	require.Equal(t, "seanathair", taggingga.SimplifyMathematical(bold))
	// super accepts only the normalized surface
	inner := func(w string) bool { return w != "seanathair" }
	require.False(t, r.irishIsMisspelled(bold, inner), "maths form should normalize before super")
	require.True(t, r.irishIsMisspelled("xyzzy", inner))
}

// Ports isMisspelled override control flow (Utils halfwidth Latin → Latin then super).
func TestIrishIsMisspelled_Halfwidth(t *testing.T) {
	r := NewMorfologikIrishSpellerRule()
	hw := "ｔｏｒｒａｃｈ"
	require.True(t, taggingga.IsAllHalfWidthChars(hw))
	require.Equal(t, "torrach", taggingga.HalfwidthLatinToLatin(hw))
	inner := func(w string) bool { return w != "torrach" }
	require.False(t, r.irishIsMisspelled(hw, inner))
	require.True(t, r.irishIsMisspelled("xyzzy", inner))
}
