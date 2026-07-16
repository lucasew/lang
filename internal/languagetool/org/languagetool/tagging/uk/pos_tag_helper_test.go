package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPosTagHelper(t *testing.T) {
	require.True(t, IsNoun("noun:m:v_naz"))
	require.True(t, IsVerb("verb:imperf:inf"))
	require.Equal(t, "m", Gender("noun:m:v_naz"))
	require.Equal(t, "v_naz", Case("noun:m:v_naz"))
}
