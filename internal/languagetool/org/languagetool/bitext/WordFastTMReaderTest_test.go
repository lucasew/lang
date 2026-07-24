package bitext

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.bitext.WordFastTMReaderTest.

func TestWordFastTMReader_Reader(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	content := "%20100801~111517\t%UserID,AHLJat,AHLJat\t%TU=00008580\t%EN-US\t%Wordfast TM v.546/00\t%PL-PL\t%\t.\n" +
		"20100727~145333\tAHLJat\t2\tEN-US\tObjection:\tPL-PL\tZarzut: \n" +
		"20100727~051350\tAHLJat\t2\tEN-US\tWhy not?&tA;\tPL-PL\tDlaczego nie?&tA; \n"
	require.NoError(t, os.WriteFile(input, []byte(content), 0o644))

	reader, err := NewWordFastTMReader(input, "UTF-8")
	require.NoError(t, err)
	i := 1
	for reader.HasNext() {
		srcAndTrg, ok, err := reader.Next()
		require.NoError(t, err)
		require.True(t, ok)
		require.NotNil(t, srcAndTrg.GetSource())
		require.NotNil(t, srcAndTrg.GetTarget())
		if i == 1 {
			require.Equal(t, "Objection:", srcAndTrg.GetSource())
		} else if i == 2 {
			require.Equal(t, "Why not?&tA;", srcAndTrg.GetSource())
		}
		i++
	}
	require.Equal(t, 3, i) // 2 pairs then exit
}
