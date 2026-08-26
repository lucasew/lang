package tools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetJnaBugWorkaroundProperty(t *testing.T) {
	t.Setenv("jna.nosys", "")
	SetJnaBugWorkaroundProperty()
	require.Equal(t, "true", os.Getenv("jna.nosys"))
}
