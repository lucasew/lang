package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMostlySingularMultiMap(t *testing.T) {
	m := NewMostlySingularMultiMap(map[string][]string{
		"one": {"a"},
		"two": {"b", "c"},
	})
	require.Equal(t, []string{"a"}, m.GetList("one"))
	require.Equal(t, []string{"b", "c"}, m.GetList("two"))
	require.Nil(t, m.GetList("missing"))
}
