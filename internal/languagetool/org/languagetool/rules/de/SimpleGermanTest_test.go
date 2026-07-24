package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleGerman_Language(t *testing.T) {
	// Simple German uses de short code surface in this port
	require.NotEmpty(t, languagetool.NewJLanguageTool("de").Analyze("Das ist ein Test."))
}
