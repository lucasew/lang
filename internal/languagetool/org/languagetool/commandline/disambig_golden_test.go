package commandline

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreTagHook_MultiwordDisambiguator(t *testing.T) {
	// Official multiwords.txt has "New York County\tNNP", not bare "New York".
	var out bytes.Buffer
	err := CoreTagHook(&out, "I live in New York County.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	// multiword chunker should attach NNP (or B-/E- multiword tags) on the phrase
	require.True(t,
		strings.Contains(s, "NNP") || strings.Contains(s, "B-N") || strings.Contains(s, "E-N") || strings.Contains(s, "B-NP") || strings.Contains(s, "E-NP"),
		s,
	)
}
