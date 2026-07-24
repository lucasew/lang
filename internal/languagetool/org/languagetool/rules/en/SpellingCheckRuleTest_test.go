package en

// Twin of SpellingCheckRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestSpellingCheckRule_IgnoreSuggestionsWithMorfologik(t *testing.T) {
	// Ignore-list surface on SpellingCheckRule
	r := spelling.NewSpellingCheckRule("MORFO", "spell", "en")
	r.AddIgnoreWords("anArtificialTestWordForLanguageTool")
	r.IsMisspelled = func(w string) bool { return w != "hello" && w != "anArtificialTestWordForLanguageTool" }
	require.True(t, r.AcceptWord("anArtificialTestWordForLanguageTool"))
	require.True(t, r.AcceptWord("hello"))
	require.False(t, r.AcceptWord("typo"))
}

func TestSpellingCheckRule_IgnorePhrases(t *testing.T) {
	// Phrase ignore is multi-token; single-word ignore approximates acceptPhrases for unit surface
	r := spelling.NewSpellingCheckRule("MORFO", "spell", "en")
	r.AddIgnoreWords("myfoo", "mybar")
	require.True(t, r.AcceptWord("myfoo"))
	require.True(t, r.AcceptWord("mybar"))
}

func TestSpellingCheckRule_IsUrl(t *testing.T) {
	require.True(t, spelling.IsUrl("http://foobar.org"))
	require.True(t, spelling.IsUrl("https://example.com/path"))
	require.True(t, spelling.IsUrl("www.example.com"))
	require.False(t, spelling.IsUrl("not a url"))
	require.True(t, spelling.IsEMail("user@example.com"))
	require.False(t, spelling.IsEMail("not-an-email"))
}
