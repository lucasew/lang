package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTag_Constants(t *testing.T) {
	require.Equal(t, Tag("picky"), TagPicky)
	require.Len(t, AllTags(), 9)
}
