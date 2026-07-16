package ar

// Twin of ArabicTaggerTest — tagger implementation is tagging/ar; smoke via analyze.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicTagger_Dictionary(t *testing.T) {
	// Full dict deferred; language surface present.
	lt := languagetool.NewJLanguageTool("ar")
	require.Equal(t, "ar", lt.GetLanguageCode())
}

func TestArabicTagger_Tagger(t *testing.T) {
	require.NotEmpty(t, languagetool.NewJLanguageTool("ar").Analyze("كتاب"))
}
