package bigdata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAggregateGoogleNgramLines_Years(t *testing.T) {
	in := strings.Join([]string{
		"the\t1900\t5", // too old
		"the\t1920\t10",
		"the\t1930\t20",
		"cat\t1950\t3",
		"Italian_ADJ\t2000\t9", // POS skip when ignorePOS
		"there\t2000\t1",
	}, "\n") + "\n"
	got, err := AggregateGoogleNgramLines(strings.NewReader(in), true)
	require.NoError(t, err)
	m := map[string]int64{}
	for _, n := range got {
		m[n.Text] = n.Count
	}
	require.Equal(t, int64(30), m["the"])
	require.Equal(t, int64(3), m["cat"])
	require.Equal(t, int64(1), m["there"])
	_, hasPOS := m["Italian_ADJ"]
	require.False(t, hasPOS)
}

func TestAggregateHiveNgramLines(t *testing.T) {
	in := "foo bar\t42\n"
	got, err := AggregateHiveNgramLines(strings.NewReader(in), false)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "foo bar", got[0].Text)
	require.Equal(t, int64(42), got[0].Count)
}

func TestIsRealPOSTag(t *testing.T) {
	require.True(t, IsRealPOSTag("Italian_ADJ"))
	require.False(t, IsRealPOSTag("_START_"))
	require.False(t, IsRealPOSTag("plain"))
}

func TestFilenameModes(t *testing.T) {
	require.True(t, IsCorpusModeFilename("googlebooks-eng-all-1gram-20120701-a.gz"))
	require.True(t, ShouldSkipPOSFilename("something_VERB_x"))
}
