package morfologik

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpellCheckWord_PrefersCleanToken(t *testing.T) {
	// Soft-hyphen surface + clean metadata (Java replaceSoftHyphens).
	at := languagetool.NewAnalyzedToken("Vertriebsniederlassung", nil, nil)
	tok := languagetool.NewAnalyzedTokenReadings(at)
	tok.SetTokenSurface("Vertriebsniederlassu\u00ADng")
	tok.SetCleanToken("Vertriebsniederlassung")
	require.Equal(t, "Vertriebsniederlassung", spellCheckWord(tok))
	require.Contains(t, tok.GetToken(), "\u00AD")
}

func TestSpellCheckWord_PlainToken(t *testing.T) {
	at := languagetool.NewAnalyzedToken("hello", nil, nil)
	tok := languagetool.NewAnalyzedTokenReadings(at)
	require.Equal(t, "hello", spellCheckWord(tok))
}

// Soft-hyphen compound must be spell-checked without U+00AD so dict accepts it.
func TestMatch_SoftHyphenNotMisspelledWhenCleanInDict(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("Vertriebsniederlassung")
	r := NewMorfologikSpellerRule("TEST", "de", "/xx.dict", sp)

	// Build ATR like applySoftHyphenMetadata: clean readings, dirty surface + cleanToken.
	at := languagetool.NewAnalyzedToken("Vertriebsniederlassung", nil, nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(at, 0)
	atr.SetTokenSurface("Vertriebsniederlassu\u00ADng")
	atr.SetCleanToken("Vertriebsniederlassung")
	// Sentence with SENT_START + soft token + end-ish
	sentStart := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil))
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart, atr})

	ms, err := r.Match(sent)
	require.NoError(t, err)
	// Clean form is in dict → no match on soft-hyphen surface
	for _, m := range ms {
		if m == nil {
			continue
		}
		// Should not flag the soft-hyphen token as misspelled
		require.False(t, m.GetFromPos() == 0 && m.GetToPos() > 0 &&
			strings.Contains(atr.GetToken(), "\u00AD"),
			"soft-hyphen clean form in dict should not produce match, got %v-%v sugs=%v",
			m.GetFromPos(), m.GetToPos(), m.GetSuggestedReplacements())
	}
	// Stronger: zero matches expected for this one-content-token sentence
	require.Empty(t, ms, "expected no spelling match for soft-hyphen form of known word")
}

func TestMatch_SoftHyphenMisspelledUsesCleanForSuggest(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("receive")
	r := NewMorfologikSpellerRule("TEST", "en", "/xx.dict", sp)

	// Soft-hyphen inside misspelling: recie\u00ADve → clean "recieve"
	at := languagetool.NewAnalyzedToken("recieve", nil, nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(at, 0)
	atr.SetTokenSurface("recie\u00ADve")
	atr.SetCleanToken("recieve")
	sentStart := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil))
	// second token so first-word capitalize may apply; use trailing word
	the := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("the", nil, nil), 10)
	sp.AddWord("the")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart, atr, the})

	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	found := false
	for _, m := range ms {
		for _, s := range m.GetSuggestedReplacements() {
			if s == "receive" || s == "Receive" {
				found = true
			}
		}
	}
	require.True(t, found, "expected receive suggestion from clean misspelling, got %+v", ms)
}
