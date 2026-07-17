package suggestions_ordering

// Twin of DetailedDamerauLevenstheinDistanceTest — Java class had no @Test methods;
// green unit coverage lives in detailed_damerau_levensthein_distance_test.go.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of DetailedDamerauLevenstheinDistanceTest (surface smoke)
func TestDetailedDamerauLevenstheinDistance_NoTests(t *testing.T) {
	require.Equal(t, 0, Compare("same", "same").Value())
	require.Equal(t, 1, Compare("ab", "ba").Value())
	require.Equal(t, 1, Compare("cat", "cats").Value())
	require.Equal(t, 1, Compare("cats", "cat").Value())
	require.Equal(t, 1, Compare("cat", "bat").Value())
}
