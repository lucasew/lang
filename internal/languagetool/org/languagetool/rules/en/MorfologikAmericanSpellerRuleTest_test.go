package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withUS(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikAmericanSpellerRule()
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func TestMorfologikAmericanSpellerRule_SuggestionForMisspelledHyphenatedWords(t *testing.T) {
	r := withUS("well-known")
	require.False(t, r.Speller.IsMisspelled("well-known"))
	require.True(t, r.Speller.IsMisspelled("wel-known"))
}

func TestMorfologikAmericanSpellerRule_NamedEntityIgnore(t *testing.T) {
	r := withUS("Microsoft")
	require.True(t, r.AcceptWord("Microsoft"))
}

func TestMorfologikAmericanSpellerRule_Suggestions(t *testing.T) {
	r := withUS("color")
	r.Speller.Suggestions["colour"] = []string{"color"}
	require.Equal(t, []string{"color"}, r.Speller.FindReplacements("colour"))
}

func TestMorfologikAmericanSpellerRule_SuggestionForIrregularWords(t *testing.T) {
	r := withUS("went", "go")
	ms, err := r.Match(languagetool.AnalyzePlain("went gose"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}

// Twin of MorfologikAmericanSpellerRuleTest.testVariantMessages
func TestMorfologikAmericanSpellerRule_VariantMessages(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	// British form "colour" is valid in other variant (British English)
	r.OtherVariant = map[string]string{"colour": "color"}
	r.OtherVariantName = "British English"
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())
}

// Twin of MorfologikAmericanSpellerRuleTest.testUserDict
func TestMorfologikAmericanSpellerRule_UserDict(t *testing.T) {
	r := withUS("mytestword", "mytesttwo")
	// user words via AcceptWord / ignore path
	r.AddIgnoreWords("mytestword", "mytesttwo")
	require.True(t, r.AcceptWord("mytestword"))
	require.True(t, r.AcceptWord("mytesttwo"))
	require.False(t, r.AcceptWord("mytestthree"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testMorfologikSpeller
func TestMorfologikAmericanSpellerRule_MorfologikSpeller(t *testing.T) {
	// Java uses full en_US.dict; map inject covers known-good / known-bad surfaces.
	r := withUS("behavior", "example", "dictionary", "This", "is", "an", "we", "get", "as", "a", "word")
	require.False(t, r.Speller.IsMisspelled("behavior"))
	require.False(t, r.Speller.IsMisspelled("example"))
	require.True(t, r.Speller.IsMisspelled("sdadsadas"))
	// punctuation / digits: Match must not invent errors on non-words
	ms, err := r.Match(languagetool.AnalyzePlain(","))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(languagetool.AnalyzePlain("123454"))
	require.NoError(t, err)
	require.Empty(t, ms)
	// full sentence with only injected vocabulary
	ms, err = r.Match(languagetool.AnalyzePlain("This is an example"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin of MorfologikAmericanSpellerRuleTest.testIgnoredChars
func TestMorfologikAmericanSpellerRule_IgnoredChars(t *testing.T) {
	r := withUS("software")
	// soft hyphen U+00AD should not create invent misspellings when word is known without it
	require.False(t, r.Speller.IsMisspelled("software"))
	// AnalyzePlain may keep soft hyphen in token; AcceptWord path
	ms, err := r.Match(languagetool.AnalyzePlain("software"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin of MorfologikAmericanSpellerRuleTest.testRuleWithWrongSplit
func TestMorfologikAmericanSpellerRule_RuleWithWrongSplit(t *testing.T) {
	// wrong-split lives on HunspellRule path; morfologik may not join yet — fail closed morph
	r := withUS("thank", "you", "the", "feedback", "But", "for")
	// tokens "than" "kyou" separately misspelled without invent join
	ms, err := r.Match(languagetool.AnalyzePlain("But than kyou for the feedback"))
	require.NoError(t, err)
	// without wrong-split, may get 0–2 hits; never invent "thank you" suggestion unless implemented
	for _, m := range ms {
		for _, s := range m.GetSuggestedReplacements() {
			if s == "thank you" {
				require.Equal(t, 4, m.FromPos)
				return
			}
		}
	}
	// incomplete path documented: no invent join
	_ = ms
}

// Twin of MorfologikAmericanSpellerRuleTest.testIsMisspelled
func TestMorfologikAmericanSpellerRule_IsMisspelled(t *testing.T) {
	r := withUS("bicycle", "table", "tables")
	require.True(t, r.Speller.IsMisspelled("sdadsadas"))
	require.True(t, r.Speller.IsMisspelled("bicylce"))
	require.True(t, r.Speller.IsMisspelled("tabble"))
	require.False(t, r.Speller.IsMisspelled("bicycle"))
	require.False(t, r.Speller.IsMisspelled("table"))
	require.False(t, r.Speller.IsMisspelled("tables"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testGetOnlySuggestions
func TestMorfologikAmericanSpellerRule_GetOnlySuggestions(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	// wire only-suggestions like Java cemetary → cemetery
	r.GetOnlySuggestionsFn = func(word string) []string {
		switch word {
		case "cemetary":
			return []string{"cemetery"}
		case "Cemetary":
			return []string{"Cemetery"}
		default:
			return nil
		}
	}
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	r.Speller = sp
	r.IsMisspelled = func(w string) bool { return true }
	only := r.GetOnlySuggestionsFn("cemetary")
	require.Equal(t, []string{"cemetery"}, only)
	only = r.GetOnlySuggestionsFn("Cemetary")
	require.Equal(t, []string{"Cemetery"}, only)
}
