package bitext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.bitext.TabBitextReaderTest.

func TestTabBitextReader_Reader(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	content := "This is not actual.\tTo nie jest aktualne.\n" +
		"Test\tTest\n" +
		"ab\tVery strange data indeed, much longer than input\n"
	require.NoError(t, os.WriteFile(input, []byte(content), 0o644))

	reader, err := NewTabBitextReader(input, "UTF-8")
	require.NoError(t, err)
	i := 1
	for reader.HasNext() {
		srcAndTrg, ok, err := reader.Next()
		require.NoError(t, err)
		require.True(t, ok)
		require.NotEmpty(t, srcAndTrg.GetSource())
		require.NotEmpty(t, srcAndTrg.GetTarget())
		switch i {
		case 1:
			require.Equal(t, "This is not actual.", srcAndTrg.GetSource())
		case 2:
			require.Equal(t, "Test", srcAndTrg.GetSource())
		case 3:
			require.Equal(t, "Very strange data indeed, much longer than input", srcAndTrg.GetTarget())
		}
		i++
	}
	require.Equal(t, 4, i) // 3 pairs + exit
}
