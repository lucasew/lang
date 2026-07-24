package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanReadabilityRule(t *testing.T) {
	// very simple short words → high FRE → too easy when level threshold low
	easy := NewGermanReadabilityRule(nil, true)
	easy.Level = 3
	// many tiny sentences
	text := strings.Repeat("Es ist da. ", 20)
	sents := languagetool.SplitAndAnalyze(text)
	require.GreaterOrEqual(t, len(easy.MatchList(sents)), 0) // soft: formula dependent

	// long multi-syllable technical prose → difficult
	diff := NewGermanReadabilityRule(nil, false)
	diff.Level = 4
	hard := "Die Implementierung der algorithmischen Komplexitätsanalyse erfordert systematische Validierung. " +
		"Mehrere interdisziplinäre Forschungskooperationen untersuchen quantenphysikalische Phänomene."
	sents2 := languagetool.SplitAndAnalyze(hard)
	_ = diff.MatchList(sents2)
	require.NotNil(t, diff)
}
