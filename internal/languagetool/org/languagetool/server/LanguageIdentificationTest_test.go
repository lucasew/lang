package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/LanguageIdentificationTest.java
// Full FastText deferred; green inject via DetectLanguageOfString.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of LanguageIdentificationTest — heuristic + preferred variants surface.
func TestLanguageIdentification_NoTests(t *testing.T) {
	require.Equal(t, "de", DetectLanguageOfString("Die Größe des Hauses.", nil, nil))
	require.Equal(t, "uk", DetectLanguageOfString("Це українська мова з ї.", nil, nil))
	require.Equal(t, "en-GB", DetectLanguageOfString("Hello world sample.", []string{"en-GB"}, func(string) string { return "en" }))
	require.Equal(t, "en-US", DetectLanguageOfString("", []string{"en-US"}, func(string) string { return "" }))
}
