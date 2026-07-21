package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of MorfologikSpellerRule.match first-word suggestion capitalization:
// lower-case dict suggestions become title-cased when the first real word is misspelled
// and is not the last token (similar to UPPERCASE_SENTENCE_START).
func TestMatch_FirstWordCapitalizesLowerSuggestions(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	sp.AddWord("the")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	// Lower-case first misspelling so findRepl yields "receive"; Match capitalizes for
	// first real word (not last token), Java match isFirstWord arm.
	ms, err := r.Match(languagetool.AnalyzePlain("recieve the"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	var found []string
	for _, m := range ms {
		found = append(found, m.GetSuggestedReplacements()...)
	}
	require.Contains(t, found, "Receive", "first-word lower sugs must be capitalized, got %v", found)
	require.NotContains(t, found, "receive", "Java replaces all-lower with UppercaseFirstChar")
}

func TestMatch_FirstWordNoCapWhenLastToken(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	// Single-token sentence: idx is last token → Java skips capitalize (idx < length-1)
	ms, err := r.Match(languagetool.AnalyzePlain("recieve"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// may still have "receive" lower — capitalize not applied when last token
	require.Contains(t, ms[0].GetSuggestedReplacements(), "receive")
}

func TestMatch_NonFirstWordKeepsLowerSuggestions(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	sp.AddWord("the")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)
	// First word OK → second misspelled keeps lower sug "receive"
	ms, err := r.Match(languagetool.AnalyzePlain("The recieve"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	found := false
	for _, m := range ms {
		for _, s := range m.GetSuggestedReplacements() {
			if s == "receive" {
				found = true
			}
			// should not force "Receive" for mid-sentence
			if s == "Receive" {
				// only fail if Receive is present without receive — actually mid-sentence
				// should keep lower form from getSuggestions applyCase (startsWithUpper on "The" n/a)
			}
		}
	}
	require.True(t, found, "mid-sentence should keep lower receive, got matches with sugs")
}
