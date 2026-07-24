package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.Tag — enum values are all-lowercase as in XML.

func TestTag_Constants(t *testing.T) {
	// Exact order and values from Tag.java
	want := []Tag{
		"picky", "academic", "clarity", "professional", "creative",
		"customer", "jobapp", "objective", "elegant",
	}
	require.Equal(t, want, AllTags())
	require.Equal(t, Tag("picky"), TagPicky)
	require.Equal(t, Tag("jobapp"), TagJobApp)
	require.Len(t, AllTags(), 9)
}
