package broker

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestDefaultResourceDataBroker(t *testing.T) {
	fsys := fstest.MapFS{
		"org/languagetool/resource/en/words.txt": &fstest.MapFile{Data: []byte("hello\nworld\n")},
		"org/languagetool/rules/en/grammar.xml":  &fstest.MapFile{Data: []byte("<rules/>")},
	}
	b := NewDefaultResourceDataBrokerFS(fsys, "", "")
	require.True(t, b.ResourceExists("/en/words.txt"))
	require.True(t, b.RuleFileExists("/en/grammar.xml"))
	lines, err := b.GetFromResourceDirAsLines("/en/words.txt")
	require.NoError(t, err)
	require.Equal(t, []string{"hello", "world"}, lines)
}
