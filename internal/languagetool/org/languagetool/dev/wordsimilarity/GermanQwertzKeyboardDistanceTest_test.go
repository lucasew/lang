package wordsimilarity

// Twin of GermanQwertzKeyboardDistanceTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of GermanQwertzKeyboardDistanceTest.testDistance
func TestGermanQwertzKeyboardDistance_Distance(t *testing.T) {
	d := NewGermanQwertzKeyboardDistance()
	require.Equal(t, float32(0), d.GetDistance('q', 'q'))
	require.Equal(t, float32(1), d.GetDistance('q', 'w'))
	require.Equal(t, float32(9), d.GetDistance('q', 'p'))
	require.Equal(t, float32(1), d.GetDistance('q', 'a'))
	require.Equal(t, float32(1), d.GetDistance('t', 'g'))
	require.Equal(t, float32(1), d.GetDistance('a', 's'))
	require.Equal(t, float32(4), d.GetDistance('a', 'g'))
	require.Equal(t, float32(1), d.GetDistance('y', 'x'))
	require.Equal(t, float32(3), d.GetDistance('c', 'n'))
	require.Equal(t, float32(2), d.GetDistance('q', 'y'))
	require.Equal(t, float32(8), d.GetDistance('q', 'm'))
	require.Equal(t, float32(2), d.GetDistance('p', 'ß'))
	require.Equal(t, float32(3), d.GetDistance('o', 'ß'))
	// uppercase
	require.Equal(t, float32(3), d.GetDistance('C', 'n'))
	require.Equal(t, float32(3), d.GetDistance('c', 'N'))
	require.Equal(t, float32(3), d.GetDistance('C', 'N'))
}
