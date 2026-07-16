package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomSpelling_SpellingCustomTxt(t *testing.T) {
	words, err := LoadCustomSpellingWords(strings.NewReader(`
# custom spelling
LanguageTool
foo-bar
# comment
baz # inline
`))
	require.NoError(t, err)
	require.Equal(t, []string{"LanguageTool", "foo-bar", "baz"}, words)
}
