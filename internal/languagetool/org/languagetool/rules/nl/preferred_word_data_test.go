package nl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPreferredWordData(t *testing.T) {
	in := `# comment
fietsen;rijden
foo;bar
`
	d, err := LoadPreferredWordData(strings.NewReader(in), "test.csv")
	require.NoError(t, err)
	require.Len(t, d.Get(), 2)
	require.Equal(t, "fietsen", d.Get()[0].OldWord)
	require.Equal(t, "rijden", d.Get()[0].NewWord)

	_, err = LoadPreferredWordData(strings.NewReader("badline\n"), "x")
	require.Error(t, err)
}
