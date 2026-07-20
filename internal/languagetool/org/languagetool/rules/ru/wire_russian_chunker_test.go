package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWireRussianChunker_PostDisambiguationOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	WireRussianChunker(lt)
	require.Nil(t, lt.Chunker)
	require.NotNil(t, lt.PostDisambiguationChunker)
}

func TestWireRussianChunker_NilSafe(t *testing.T) {
	require.NotPanics(t, func() { WireRussianChunker(nil) })
}

func TestRegisterCore_WiresPostDisambiguationChunker(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	require.Nil(t, lt.Chunker)
	require.NotNil(t, lt.PostDisambiguationChunker)
}
