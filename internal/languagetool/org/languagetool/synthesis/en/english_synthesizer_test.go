package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME) without dict.
func TestEnglishSynthesizer_ResourcePaths(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	require.Equal(t, EnglishSynthResource, s.ResourceFileName)
	require.Equal(t, EnglishTagsFile, s.TagFileName)
	require.Equal(t, EnglishSorFile, s.SorFileName)
	require.Equal(t, "en", s.LangShortCode)
}

// IsException ports EnglishSynthesizer.isException (leading apostrophe / exceptions list).
func TestEnglishSynthesizer_IsException(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	require.True(t, s.IsException("'ve"))
	require.True(t, s.IsException("n't"))
	require.True(t, s.IsException("ne'er"))
	require.False(t, s.IsException("was"))
}

func TestEnglishSynthesizer_RemoveExceptions(t *testing.T) {
	s := NewEnglishSynthesizer(nil)
	require.Equal(t, []string{"was"}, s.removeExceptions([]string{"was", "'ve", "n't"}))
}
