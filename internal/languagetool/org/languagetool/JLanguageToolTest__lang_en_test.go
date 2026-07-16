package languagetool

// Twin of JLanguageToolTest homepage demos — smoke analysis only (no full rules engine).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJLanguageTool_DemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	require.Equal(t, "en-US", lt.GetLanguageCode())
	sents := lt.Analyze("A sentence with a error in the Hitchhiker's Guide tot he Galaxy")
	require.NotEmpty(t, sents)
}

func TestJLanguageTool_SpellCheckerDemoCodeForHomepage(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	// Analysis path works without dictionary-backed speller wiring.
	sents := lt.Analyze("A speling error")
	require.NotEmpty(t, sents)
}

func TestJLanguageTool_SpellCheckerDemoCodeForHomepageWithAddedWords(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	sents := lt.Analyze("LanguageTool")
	require.NotEmpty(t, sents)
}
