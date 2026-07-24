package ca

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDiacriticsIEC_NPException(t *testing.T) {
	rule := NewSimpleReplaceDiacriticsIEC(nil)
	// Pick a word from the replace list if any — NP-tagged surface is skipped.
	words := loadDiacriticsIEC()
	var sample string
	for k := range words {
		sample = k
		break
	}
	if sample == "" {
		t.Skip("empty diacritics list")
	}
	// Capitalized sample with NP → no match
	cap := strings.ToUpper(sample[:1]) + sample[1:]
	sent := languagetool.AnalyzePlain(cap + ".")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && strings.EqualFold(tok.GetToken(), sample) {
			pos := "NP00SP0"
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		}
	}
	require.Equal(t, 0, len(rule.Match(sent)))
	// Without NP → match if entry exists
	require.GreaterOrEqual(t, len(rule.Match(languagetool.AnalyzePlain(sample+"."))), 1)
}
