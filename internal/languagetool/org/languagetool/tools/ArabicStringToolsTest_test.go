package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tools.ArabicStringToolsTest.

func TestArabicStringTools_RemoveTashkeel(t *testing.T) {
	require.Equal(t, "", RemoveTashkeel(""))
	require.Equal(t, "a", RemoveTashkeel("a"))
	require.Equal(t, "öäü", RemoveTashkeel("öäü"))
	require.Equal(t, "كتب", RemoveTashkeel("كَتَب"))
}
