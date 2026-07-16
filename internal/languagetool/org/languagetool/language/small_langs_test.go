package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmallLangs(t *testing.T) {
	require.GreaterOrEqual(t, len(AllSmallLangs()), 12)
	require.Equal(t, "Ukrainian", Ukrainian.GetName())
	require.Equal(t, "sk", Slovak.GetShortCode())
}
