package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestEnglishOnlySuggestions_ReplaceOnce(t *testing.T) {
	require.Equal(t, []string{"cemetery"}, EnglishOnlySuggestions("cemetary"))
	require.Equal(t, []string{"Cemetery"}, EnglishOnlySuggestions("Cemetary"))
	require.Equal(t, []string{"basically"}, EnglishOnlySuggestions("basicly"))
	require.Nil(t, EnglishOnlySuggestions("proliferation")) // correct spelling — no only-suggestion arm
	require.Equal(t, []string{"proliferation"}, EnglishOnlySuggestions("profileration"))
	require.Equal(t, []string{"Proliferation"}, EnglishOnlySuggestions("Profileration"))
}

func TestEnglishOnlySuggestions_Fixed(t *testing.T) {
	require.Equal(t, []string{"swam"}, EnglishOnlySuggestions("swimmed"))
	require.Equal(t, []string{"LanguageTool"}, EnglishOnlySuggestions("languagetool"))
	require.Equal(t, []string{"Microsoft"}, EnglishOnlySuggestions("microsoft"))
	// Java ADHOC = [Ad]hoc → "Ahoc" / "dhoc" (not "Adhoc")
	require.Equal(t, []string{"ad hoc"}, EnglishOnlySuggestions("Ahoc"))
	require.Equal(t, []string{"ad hoc"}, EnglishOnlySuggestions("dhoc"))
	require.Nil(t, EnglishOnlySuggestions("hello"))
}

func TestEnglishOnlySuggestions_WiredOnSpeller(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	require.NotNil(t, r.GetOnlySuggestionsFn)
	// empty map Words → AcceptWord fail-closed not misspelled; inject misspell
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("ok")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	m, err := r.Match(languagetool.AnalyzePlain("cemetary"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"cemetery"}, m[0].GetSuggestedReplacements())
}

func TestEnglishOnlySuggestions_Multi(t *testing.T) {
	require.Equal(t, []string{"QuillBot's", "QuillBot"}, EnglishOnlySuggestions("QuillBots"))
	require.Equal(t, []string{"TV", "to"}, EnglishOnlySuggestions("tv"))
	require.Equal(t, []string{"just", "gist"}, EnglishOnlySuggestions("jist"))
}
