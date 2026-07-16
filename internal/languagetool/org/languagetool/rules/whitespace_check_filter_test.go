package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWhitespaceCheckFilter(t *testing.T) {
	f := NewWhitespaceCheckFilter()
	// keep when whitespace before token != expected
	keep, err := f.Accept([]string{"", " "}, 2, " ")
	require.Empty(t, err)
	require.False(t, keep) // matches expected space → suppress
	keep, err = f.Accept([]string{"", " "}, 2, "\t")
	require.Empty(t, err)
	require.True(t, keep)
	_, err = f.Accept([]string{" "}, 2, " ")
	require.NotEmpty(t, err)
}
