package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPOSAndSpellAndSynthBuilders(t *testing.T) {
	pos := NewPOSDictionaryBuilder(map[string]string{"fsa.dict.encoding": "utf-8", "fsa.dict.separator": "+"})
	require.Equal(t, "utf-8", pos.Encoding())
	require.Equal(t, "+", pos.Separator())
	var buf strings.Builder
	n, err := pos.NormalizeTaggerInput(strings.NewReader("dogs\tdog\tNNS\n"), &buf)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "dogs\tdog\tNNS")

	sp := NewSpellDictionaryBuilder(map[string]string{"fsa.dict.separator": "+"})
	sp.FreqList["hello"] = 10
	buf.Reset()
	n, err = sp.TokenizeInput(strings.NewReader("hello\nworld\n"), &buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Contains(t, buf.String(), "hello+")

	syn := NewSynthDictionaryBuilder(nil)
	require.NoError(t, syn.SetIgnorePOSRegex(PolishIgnorePOSRegex))
	buf.Reset()
	n, err = syn.ReverseLineContent(strings.NewReader("dogs\tdog\tNNS\nbad\tx\t:neg\n"), &buf)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Contains(t, buf.String(), "dog\tdogs\tNNS")

	buf.Reset()
	n, err = WritePOSTags(strings.NewReader("a\tb\tTAG1\nc\td\tTAG1\ne\tf\tTAG2\n"), &buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)

	exp := NewDictionaryExporter(nil)
	exp.SetOutputFilename("out.txt")
	require.Equal(t, ExportModeFSA, ExportModeFor("/uk/hunspell/uk.dict"))
	require.Equal(t, ExportModeDict, ExportModeFor("/uk/ukrainian.dict"))
	require.Contains(t, exp.DescribeExport("/x.dict"), "export")
}

func TestValidateTaggerLine(t *testing.T) {
	require.NoError(t, ValidateTaggerLine("a\tb\tc"))
	require.Error(t, ValidateTaggerLine("a b c"))
}
