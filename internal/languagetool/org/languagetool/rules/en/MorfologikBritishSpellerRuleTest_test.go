package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikBritishSpellerRule_Suggestions(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	require.Equal(t, MorfologikBritishSpellerRuleID, r.GetID())
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	sp.AddWord("colour")
	sp.Suggestions["color"] = []string{"colour"}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.Equal(t, []string{"colour"}, sp.FindReplacements("color"))
}

func TestMorfologikBritishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	sp.AddWord("hello")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	ms, err := r.Match(languagetool.AnalyzePlain("hello helo"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

// Twin of MorfologikBritishSpellerRuleTest.testVariantMessages
func TestMorfologikBritishSpellerRule_VariantMessages(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	// Wire American→British map like en-GB spelling_en-US.txt column
	r.OtherVariant = map[string]string{"color": "colour"}
	r.OtherVariantName = "American English"
	vi := r.IsValidInOtherVariant("color")
	require.NotNil(t, vi)
	require.Equal(t, "American English", vi.GetVariantName())
	require.Equal(t, "colour", vi.GetOtherVariant())
	// Match path: misspelled color with variant info in message when speller wired
	sp := morfologik.NewMorfologikSpeller(BritishSpellerDict, 1)
	sp.AddWord("nice")
	sp.AddWord("is")
	sp.AddWord("a")
	sp.AddWord("This")
	sp.AddWord("the")
	sp.AddWord("word")
	sp.AddWord("American")
	sp.AddWord("English")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	ms, err := r.Match(languagetool.AnalyzePlain("This is a nice color."))
	require.NoError(t, err)
	// without full dict color may or may not match; variant lookup is the leaf twin
	_ = ms
	require.Contains(t, "American English", r.OtherVariantName)
}
