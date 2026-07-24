package wikipedia

// Twin of LocationHelperTest (Java class is @Ignore but logic is portable)
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocationHelper_AbsolutePositionFor(t *testing.T) {
	pos, err := AbsolutePositionFor(1, 1, "hallo")
	require.NoError(t, err)
	require.Equal(t, 0, pos)
	pos, err = AbsolutePositionFor(1, 2, "hallo")
	require.NoError(t, err)
	require.Equal(t, 1, pos)
	pos, err = AbsolutePositionFor(2, 1, "hallo\nx")
	require.NoError(t, err)
	require.Equal(t, 6, pos)
	pos, err = AbsolutePositionFor(3, 3, "\n\nxyz")
	require.NoError(t, err)
	require.Equal(t, 4, pos)
}

func TestLocationHelper_InvalidPosition(t *testing.T) {
	pos, err := AbsolutePositionFor(1, 1, "hallo")
	require.NoError(t, err)
	require.Equal(t, 0, pos)
	_, err = AbsolutePositionFor(2, 2, "hallo")
	require.Error(t, err)
}
