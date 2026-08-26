package synthesis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWords(t *testing.T) {
	words, err := LoadWords(strings.NewReader("#c\nfoo\n\nbar\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, words)
}
