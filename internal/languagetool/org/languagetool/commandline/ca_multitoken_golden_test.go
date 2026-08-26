package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Core CA path still analyzes without soft hybrid multitoken invent.
func TestGolden_CA_AnalyzeWithoutSoftHybrid(t *testing.T) {
	lt, err := configureCoreLT("ca", &CommandLineOptions{Language: "ca"})
	require.NoError(t, err)
	sents := lt.Analyze("Això és un test.")
	require.NotEmpty(t, sents)
}
