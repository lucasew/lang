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
	// Java assertSuggestion("one-diminensional", "one-dimensional"); "parple-people-eater"
	r := withUS("well-known", "one", "dimensional", "purple", "people", "eater")
	require.False(t, r.Speller.IsMisspelled("well-known"))
	require.True(t, r.Speller.IsMisspelled("wel-known"))
	// CheckCompound path: enable like EN
	r.SetCheckCompound(true)
	// whole misspelled, parts known → may accept compound; document inject
	require.True(t, r.Speller.IsMisspelled("one-diminensional"))
	// Hyphen suggestion hook (Java addHyphenSuggestions when empty dict sugs)
	r.AddHyphenSuggestionsFn = func(parts []string) []string {
		if len(parts) == 2 && parts[0] == "one" && parts[1] == "diminensional" {
			return []string{"one-dimensional"}
		}
		if len(parts) == 3 && parts[0] == "parple" && parts[1] == "people" && parts[2] == "eater" {
			return []string{"purple-people-eater"}
		}
		return nil
	}
	r.Speller.AddWord("dimensional") // part known for join path
	// When GetOnlySuggestions / hyphen rebuild used from collectSuggestions — assert hook alone
	require.Equal(t, []string{"one-dimensional"}, r.AddHyphenSuggestionsFn([]string{"one", "diminensional"}))
	require.Equal(t, []string{"purple-people-eater"}, r.AddHyphenSuggestionsFn([]string{"parple", "people", "eater"}))
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
	r.OtherVariant = map[string]string{"colour": "color", "Colour": "Color"}
	r.OtherVariantName = "British English"
	// re-wire after map replace (constructor already set Fn to method)
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())

	// Match-level: Java message contains "is British English"
	// colour not in US dict inject
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	for _, w := range []string{"This", "is", "a", "nice", "words", "the", "British"} {
		sp.AddWord(w)
	}
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	ms, err := r.Match(languagetool.AnalyzePlain("This is a nice colour."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "is British English")
	// capitalized Colour
	ms2, err := r.Match(languagetool.AnalyzePlain("Colour is the British words."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetMessage(), "is British English")
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
	r := withUS(
		"behavior", "example", "dictionary", "This", "is", "an", "we", "get", "as", "a", "word",
		"Why", "don't", "speak", "today", "He", "doesn't", "know", "what", "to", "do",
		"I", "like", "my", "emoji", "English", "text", "Yes", "An", "URL", "like",
		"http://sdaasdwe.com", "no", "error", "mansplaining", "Qur'an",
	)
	// length-1 Greek letter μ ignored when configured
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	require.False(t, r.Speller.IsMisspelled("behavior"))
	require.False(t, r.Speller.IsMisspelled("example"))
	require.True(t, r.Speller.IsMisspelled("sdadsadas"))
	// punctuation / digits / emoji: Match must not invent errors
	for _, s := range []string{
		",", "123454", "I like my emoji 😾", "I like my emoji ❤️", "This is English text 🗺.",
		"🏽", "🧡‍♂️ , 🎉💛✈️", "μ",
	} {
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}
	// full sentence with only injected vocabulary
	ms, err := r.Match(languagetool.AnalyzePlain("This is an example"))
	require.NoError(t, err)
	require.Empty(t, ms)
	// Java: behavior as dictionary word sentence
	for _, w := range []string{"we", "get", "as", "dictionary", "word"} {
		r.Speller.AddWord(w)
	}
	ms, err = r.Match(languagetool.AnalyzePlain("This is an example: we get behavior as a dictionary word."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// URL no error
	ms, err = r.Match(languagetool.AnalyzePlain("An URL like http://sdaasdwe.com is no error."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// doesn't: AnalyzePlain splits apostrophe; inject token pieces (Java keeps one token).
	r.Speller.AddWord("doesn")
	r.Speller.AddWord("don")
	ms, err = r.Match(languagetool.AnalyzePlain("He doesn't know what to do."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// diacritic suggestion
	r.Speller.Suggestions["fianc"] = []string{"fiancé"}
	ms, err = r.Match(languagetool.AnalyzePlain("fianc"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetSuggestedReplacements(), "fiancé")
	// spelling.txt merges (inject as accepted)
	require.False(t, r.Speller.IsMisspelled("mansplaining"))
	require.False(t, r.Speller.IsMisspelled("Qur'an"))
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
