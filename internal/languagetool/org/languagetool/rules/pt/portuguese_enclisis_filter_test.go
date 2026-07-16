package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseEnclisisFilter_PronounTags(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	tags := f.PronounTags([]PronounTagReading{{Token: "nos", POS: "PP1CPO00"}}, "dizem", false)
	require.Equal(t, []string{"PP1CPO00", "PP3MPA00"}, tags)

	tags = f.PronounTags([]PronounTagReading{{Token: "eles", POS: "PP3MPN00"}}, "ver", true)
	require.Equal(t, []string{"PP3MPA00"}, tags)
}

func TestPortugueseEnclisisFilter_Suggest(t *testing.T) {
	f := NewPortugueseEnclisisFilter()
	f.SynthesizeEnclisis = func(verb, pos, ptag string) []string {
		return []string{verb + "-" + ptag}
	}
	got := f.Suggest(VerbReading{Token: "ver", POS: "VMN0000"}, []string{"PP3MSA00"})
	require.Equal(t, []string{"ver-PP3MSA00"}, got)
}
