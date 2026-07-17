package commandline

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreTagHook_MultiwordDisambiguator(t *testing.T) {
	var out bytes.Buffer
	err := CoreTagHook(&out, "I live in New York.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "New York")
	// multiword chunker should attach NNP-family tags to New and/or York
	require.True(t,
		strings.Contains(s, "NNP") || strings.Contains(s, "B-N") || strings.Contains(s, "E-N") || strings.Contains(s, "B-NP") || strings.Contains(s, "E-NP"),
		s,
	)
}
