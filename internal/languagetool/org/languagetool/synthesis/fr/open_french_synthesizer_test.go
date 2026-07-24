package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenFrenchSynthesizer_Missing(t *testing.T) {
	require.Nil(t, OpenFrenchSynthesizerFromDir(""))
	require.Nil(t, OpenFrenchSynthesizerFromDictPath(""))
	require.Nil(t, OpenFrenchSynthesizerFromDir(t.TempDir()))
}

func TestFrenchSynthesizer_IsException(t *testing.T) {
	s := NewFrenchSynthesizer(nil)
	require.True(t, s.IsException("qqchose"))
	require.True(t, s.IsException("informè")) // trailing è not in allowlist
	require.False(t, s.IsException("burkinabè"))
	require.False(t, s.IsException("maison"))
}
