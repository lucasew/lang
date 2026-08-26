package dev

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDuplicateLines(t *testing.T) {
	in := "# c\na\nb\na\n# c\nc\nb\n"
	var out strings.Builder
	require.NoError(t, RemoveDuplicateLines(strings.NewReader(in), &out))
	// comments always printed; unique a,b,c
	got := out.String()
	lines := strings.Split(strings.TrimSpace(got), "\n")
	// comments always printed; non-comments unique in first-seen order
	require.Equal(t, []string{"# c", "a", "b", "# c", "c"}, lines)
}
