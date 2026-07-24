package crh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrimeanTatarSynthesizer(t *testing.T) {
	s := NewCrimeanTatarSynthesizer(nil)
	require.Equal(t, "crh", s.LangShortCode)
	require.Equal(t, CrimeanTatarSynthDict, s.ResourceFileName)
}
