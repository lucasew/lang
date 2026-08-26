package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilderOptionsAndDictBuilder(t *testing.T) {
	o, err := ParseBuilderArgs([]string{"-i", "in.txt", "-o", "out.dict", "-info", "x.info", "-freq", "f.xml"})
	require.NoError(t, err)
	require.Equal(t, "in.txt", o.InputFile)
	require.Equal(t, "out.dict", o.OutputFile)
	require.Equal(t, "x.info", o.InfoFile)

	b := NewDictionaryBuilder(map[string]string{"fsa.dict.separator": "+"})
	b.SetOutputFilename("out.dict")
	require.Equal(t, "out.dict", b.GetOutputFilename())
	require.NoError(t, b.LoadFrequencyList(strings.NewReader(`<w f="10">hello</w>`+"\n"+`<w f="200">world</w>`)))
	require.Equal(t, 10, b.FreqList["hello"])
	require.Equal(t, byte('A'), FreqToRange(0))
	require.True(t, FreqToRange(255) >= 'A' && FreqToRange(255) <= 'Z')

	entries, err := ReadTaggerEntries(strings.NewReader("dogs\tdog\tNNS\n#c\ncat\tcat\tNN\n"))
	require.NoError(t, err)
	require.Len(t, entries, 2)
	var buf strings.Builder
	require.NoError(t, WriteSpellingList(&buf, entries))
	require.Contains(t, buf.String(), "dogs")
	require.Contains(t, buf.String(), "cat")
}
