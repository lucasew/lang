package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectLanguageHeuristic(t *testing.T) {
	require.Equal(t, "de", DetectLanguageHeuristic("Die Größe des Hauses."))
	require.Equal(t, "uk", DetectLanguageHeuristic("Це українська мова з ї."))
	require.Equal(t, "en", DetectLanguageHeuristic("Hello world"))
}

func TestResolveLanguage(t *testing.T) {
	opts := NewCommandLineOptions()
	opts.SetLanguage("pl")
	require.Equal(t, "pl", ResolveLanguage("x", opts, nil))
	opts.SetAutoDetect(true)
	require.Equal(t, "de", ResolveLanguage("Größe", opts, nil))
}

func TestInferLanguageFromRuleFileName(t *testing.T) {
	require.Equal(t, "en", InferLanguageFromRuleFileName("/tmp/grammar-en-US.xml"))
	require.Equal(t, "de", InferLanguageFromRuleFileName("rules-de.xml"))
}
