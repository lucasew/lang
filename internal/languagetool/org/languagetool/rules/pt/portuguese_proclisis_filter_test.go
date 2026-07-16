package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseProclisisFilter(t *testing.T) {
	f := NewPortugueseProclisisFilter()
	got := f.Suggest([]struct{ Token, POS string }{
		{Token: "fazê-lo", POS: "VMN0000:PP3MSA00"},
	})
	require.Contains(t, got, "o fazê")
	// nos with plural verb ending
	got = f.Suggest([]struct{ Token, POS string }{
		{Token: "dizem-nos", POS: "VMIP3P0:PP1CPO00"},
	})
	require.Contains(t, got, "nos dizem")
	require.Contains(t, got, "os dizem")
}
