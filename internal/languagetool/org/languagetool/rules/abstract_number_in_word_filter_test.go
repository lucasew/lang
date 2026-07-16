package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumberInWordFilter(t *testing.T) {
	f := NewNumberInWordFilter()
	require.Nil(t, f.Suggestions("hello"))
	require.Equal(t, []string{"word", "wrd"}, f.Suggestions("w0rd"))
	require.Equal(t, []string{"cas"}, f.Suggestions("cas4"))
}
