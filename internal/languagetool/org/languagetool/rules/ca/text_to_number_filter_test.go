package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextToNumberFilter_Catalan(t *testing.T) {
	f := NewTextToNumberFilter()
	require.Equal(t, "0", f.ConvertTokens([]string{"zero"}))
	require.Equal(t, "2", f.ConvertTokens([]string{"dues"}))
	require.Equal(t, "17", f.ConvertTokens([]string{"disset"}))
	require.Equal(t, "20", f.ConvertTokens([]string{"vint"}))
	// hyphen tokenize: vint-i-un style not fully modeled; single tokens work
	require.Equal(t, "21", f.ConvertTokens([]string{"vint-i-un"})) // un=1, vint not in numbers as "vint-i-un" split: vint, i, un → 20+1 if "i" ignored
	require.Equal(t, "1000", f.ConvertTokens([]string{"mil"}))
	require.Equal(t, "2000", f.ConvertTokens([]string{"dos", "mil"}))
	// decimal uses comma formatResult: 0.5 → "0,5"
	require.Equal(t, "0,5", f.ConvertTokens([]string{"mig"}))
	// percentage "per cent"
	require.Equal(t, "10\u202F%", f.ConvertTokens([]string{"deu", "per", "cent"}))
}
