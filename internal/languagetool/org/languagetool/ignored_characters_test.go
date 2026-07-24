package languagetool

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyIgnoredCharactersRegex_SoftHyphen(t *testing.T) {
	in := []string{"Vertriebsniederlassu\u00ADng", "ok", "\u00AD"}
	out := ApplyIgnoredCharactersRegex(in, GermanIgnoredCharactersRegex)
	require.Equal(t, []string{"Vertriebsniederlassung", "ok", ""}, out)
	// nil re unchanged
	require.Equal(t, in, ApplyIgnoredCharactersRegex(in, nil))
}

func TestReplaceSoftHyphens_KeepsOrig(t *testing.T) {
	in := []string{"Vertriebsniederlassu\u00ADng", "ok"}
	cleaned, soft := ReplaceSoftHyphens(in, GermanIgnoredCharactersRegex)
	require.Equal(t, []string{"Vertriebsniederlassung", "ok"}, cleaned)
	require.Contains(t, soft, 0)
	require.Equal(t, "Vertriebsniederlassu\u00ADng", soft[0].Orig)
	require.Equal(t, "Vertriebsniederlassung", soft[0].Clean)
	require.NotContains(t, soft, 1)
}

func TestAnalyze_GermanSoftHyphenCleanTokenAndPosFix(t *testing.T) {
	// Java getRawAnalyzedSentence: getToken() becomes orig (with U+00AD) via addReading;
	// getCleanToken() is without soft hyphen; posFix accumulates for later tokens.
	lt := NewJLanguageTool("de")
	lt.IgnoredCharacters = GermanIgnoredCharactersRegex
	sents := lt.Analyze("Die Vertriebsniederlassu\u00ADng.")
	require.NotEmpty(t, sents)
	var softTok *AnalyzedTokenReadings
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		if strings.Contains(tok.GetToken(), "\u00AD") || tok.GetCleanToken() == "Vertriebsniederlassung" {
			softTok = tok
			break
		}
	}
	require.NotNil(t, softTok, "expected soft-hyphen compound token")
	require.Equal(t, "Vertriebsniederlassung", softTok.GetCleanToken())
	// Surface after soft-hyphen metadata: original with U+00AD (Java addReading)
	require.Contains(t, softTok.GetToken(), "\u00AD")
	// addReading appends a null-POS reading for the orig surface (Java createToken(orig, null)).
	require.GreaterOrEqual(t, len(softTok.GetReadings()), 1)
}

func TestApplyIgnoredCharactersRegex_RussianCombining(t *testing.T) {
	// soft hyphen + combining acute/grave stripped
	in := []string{"мо\u00adло\u0301ко", "да\u0300"}
	out := ApplyIgnoredCharactersRegex(in, RussianIgnoredCharactersRegex)
	require.Equal(t, []string{"молоко", "да"}, out)
}

func TestApplyIgnoredCharactersRegex_UkrainianAcute(t *testing.T) {
	in := []string{"мо\u00adло\u0301ко"}
	out := ApplyIgnoredCharactersRegex(in, UkrainianIgnoredCharactersRegex)
	require.Equal(t, []string{"молоко"}, out)
}

func TestApplyIgnoredCharactersRegex_BelarusianCombining(t *testing.T) {
	// Java Belarusian: same as Russian — soft hyphen + combining acute/grave
	in := []string{"мо\u00adло\u0301ко", "да\u0300"}
	out := ApplyIgnoredCharactersRegex(in, BelarusianIgnoredCharactersRegex)
	require.Equal(t, []string{"молоко", "да"}, out)
}
