package en

// Twin of EnglishNumberInWordFilterTest.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishNumberInWordFilter_Filter(t *testing.T) {
	f := NewEnglishNumberInWordFilter()
	require.Contains(t, f.Suggestions("H0use"), "House")
	require.Empty(t, f.Suggestions("House"))
}
