package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/stretchr/testify/require"
)

func TestMorfologikIrishSpellerRule(t *testing.T) {
	r := NewMorfologikIrishSpellerRule()
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", MorfologikIrishSpellerRuleID)
	require.Equal(t, "/ga/hunspell/ga_IE.dict", IrishSpellerDict)
	require.Equal(t, MorfologikIrishSpellerRuleID, r.GetID())
	require.Equal(t, IrishSpellerDict, r.GetFileName())
	// Java ignoreWordsWithLength = 1
	require.Equal(t, 1, r.IgnoreWordsWithLength)
}

func TestIrishIsMisspelled_MathsNormalize(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(IrishSpellerDict, 1)
	sp.AddWord("seanathair")
	r := NewMorfologikIrishSpellerRule()
	r.Speller = sp
	// re-wrap with map IsMisspelled
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.irishIsMisspelled(w, inner) }
	// bold mathematical "seanathair"
	bold := "\U0001D42C\U0001D41E\U0001D41A\U0001D427\U0001D41A\U0001D42D\U0001D421\U0001D41A\U0001D422\U0001D42B"
	require.True(t, taggingga.IsAllMathsChars(bold))
	require.False(t, r.IsMisspelled(bold), "maths form should normalize to accepted seanathair")
	require.True(t, r.IsMisspelled("xyzzy"))
}

func TestIrishIsMisspelled_Halfwidth(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(IrishSpellerDict, 1)
	sp.AddWord("torrach")
	r := NewMorfologikIrishSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.irishIsMisspelled(w, inner) }
	// fullwidth Latin t o r r a c h
	hw := "ｔｏｒｒａｃｈ"
	require.True(t, taggingga.IsAllHalfWidthChars(hw))
	require.Equal(t, "torrach", taggingga.HalfwidthLatinToLatin(hw))
	require.False(t, r.IsMisspelled(hw))
}

func TestIrishHyphenTokenizing(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(IrishSpellerDict, 1)
	sp.AddWord("a")
	sp.AddWord("b")
	r := NewMorfologikIrishSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.irishIsMisspelled(w, inner) }
	m, err := r.Match(languagetool.AnalyzePlain("a-b"))
	require.NoError(t, err)
	require.Empty(t, m)
	m, err = r.Match(languagetool.AnalyzePlain("a-zz"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

func TestIrishIgnoreWordsWithLength1(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(IrishSpellerDict, 1)
	r := NewMorfologikIrishSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.irishIsMisspelled(w, inner) }
	// single letter accepted via ignoreWordsWithLength
	m, err := r.Match(languagetool.AnalyzePlain("x"))
	require.NoError(t, err)
	require.Empty(t, m)
}
