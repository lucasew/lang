package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnNoInfinitiuSuggestionFilter(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	f.Synth = func(lemma, postag string) string {
		if postag == "VMIP3S00" {
			return "veu"
		}
		if postag == "VMIP1S00" {
			return "veig"
		}
		return ""
	}
	got := f.Suggest(EnNoInfinitiuInput{
		TempsVerbal: "VMIP1S00",
		Lemma:       "veure",
		VerbBefore:  false,
	})
	require.Contains(t, got, "com que no veu")
	require.Contains(t, got, "com que no veig")
}
