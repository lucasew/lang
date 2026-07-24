package wikipedia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorMarker(t *testing.T) {
	e := NewErrorMarker("<err>", "</err>")
	require.Equal(t, "<err>", e.GetStartMarker())
	require.Equal(t, "</err>", e.GetEndMarker())
}
