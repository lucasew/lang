package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrdinalSuffixFilter(t *testing.T) {
	f := NewOrdinalSuffixFilter()
	require.Equal(t, "1st", f.Fix("1nd"))
	require.Equal(t, "2nd", f.Fix("2th"))
	require.Equal(t, "3rd", f.Fix("3st"))
	require.Equal(t, "4th", f.Fix("4nd"))
	require.Equal(t, "11th", f.Fix("11st"))
	require.Equal(t, "12th", f.Fix("12nd"))
	require.Equal(t, "13th", f.Fix("13rd"))
	require.Equal(t, "21st", f.Fix("21nd"))
	require.Equal(t, "22nd", f.Fix("22th"))
	require.Equal(t, "23rd", f.Fix("23th"))
}
