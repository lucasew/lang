package server

// Twin of HTTPServerDefaultLanguageIdentifierTest (Java @Ignore load test).
// Soft: DetectLanguageOfString inject (full FastText/Tatoeba deferred).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPServerDefaultLanguageIdentifierTest (no @Test)
func TestHTTPServerDefaultLanguageIdentifier_NoTests(t *testing.T) {
	require.Equal(t, "de", DetectLanguageOfString("Die Größe des Hauses.", nil, nil))
	require.Equal(t, "uk", DetectLanguageOfString("Це українська мова з ї.", nil, nil))
	require.Equal(t, "en-GB", DetectLanguageOfString("Hello world sample.", []string{"en-GB"}, func(string) string { return "en" }))
	require.Equal(t, "en-US", DetectLanguageOfString("", []string{"en-US"}, func(string) string { return "" }))
}
