package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextToNumberFilter_Spanish(t *testing.T) {
	f := NewTextToNumberFilter()
	require.Equal(t, "0", f.ConvertTokens([]string{"cero"}))
	require.Equal(t, "1", f.ConvertTokens([]string{"uno"}))
	require.Equal(t, "16", f.ConvertTokens([]string{"dieciséis"}))
	require.Equal(t, "21", f.ConvertTokens([]string{"veintiuno"}))
	require.Equal(t, "100", f.ConvertTokens([]string{"cien"}))
	require.Equal(t, "200", f.ConvertTokens([]string{"doscientos"}))
	require.Equal(t, "2000", f.ConvertTokens([]string{"dos", "mil"}))
	require.Equal(t, "1000", f.ConvertTokens([]string{"mil"}))
	require.Equal(t, "1000000", f.ConvertTokens([]string{"un", "millón"}))
	// 30 + 5 = additive before multiplier boundary
	require.Equal(t, "35", f.ConvertTokens([]string{"treinta", "cinco"}))
	// percentage: "por ciento"
	require.Equal(t, "50\u202F%", f.ConvertTokens([]string{"cincuenta", "por", "ciento"}))
	// decimal
	require.Equal(t, "3.5", f.ConvertTokens([]string{"tres", "coma", "cinco"}))
}

func TestTextToNumberFilter_SpanishCompoundHundreds(t *testing.T) {
	f := NewTextToNumberFilter()
	// ciento + dos + mil → 100+2 then *? No: ciento and dos are numbers, mil multiplies current
	// total path: current=100+2=102, mil → total=102000
	require.Equal(t, "102000", f.ConvertTokens([]string{"ciento", "dos", "mil"}))
}
