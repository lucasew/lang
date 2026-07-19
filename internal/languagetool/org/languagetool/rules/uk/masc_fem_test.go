package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMascFemSet(t *testing.T) {
	require.True(t, IsInMascFemSet("автор"))
	require.True(t, IsInMascFemSet("адвокат"))
	require.True(t, IsInMascFemSet("екс-автор"))
	require.False(t, IsInMascFemSet("стіл"))
}

func TestHasMascFemLemma(t *testing.T) {
	lem := "лікар"
	require.True(t, HasMascFemLemma(atrLemma("лікар", &lem, "noun:anim:m:v_naz")))
	require.True(t, HasMascFemLemma(atr("біолог", "noun:anim:m:v_naz")))
	require.False(t, HasMascFemLemma(atr("жінка", "noun:anim:f:v_naz")))
}

func TestNounVerbException_MascFem(t *testing.T) {
	// лікар may not be in file — use автор
	lem := "автор"
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atrLemma("автор", &lem, "noun:anim:m:v_naz"),
		atr("написала", "verb:perf:past:f"),
	}, 0, 1))
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("пора", "noun:inanim:f:v_naz"),
		atr("було", "verb:imperf:past:n"),
	}, 0, 1))
	require.True(t, IsNounVerbException([]*languagetool.AnalyzedTokenReadings{
		atr("решта", "noun:inanim:f:v_naz"),
		atr("забороняються", "verb:imperf:pres:p:3"),
	}, 0, 1))
}
