package dev

// Twin of FrequencyIndexCreatorTest — delegates to bigdata green tests
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/bigdata"
	"github.com/stretchr/testify/require"
)

func TestFrequencyIndexCreator_NoTests(t *testing.T) {
	// smoke aggregation used by FrequencyIndexCreator plain-text path
	got, err := bigdata.AggregateGoogleNgramLines(strings.NewReader("a\t2000\t1\na\t2001\t2\n"), true)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, int64(3), got[0].Count)
}
