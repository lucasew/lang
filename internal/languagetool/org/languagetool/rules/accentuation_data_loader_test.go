package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccentuationDataLoader_Single(t *testing.T) {
	in := `# comment
cafe;café;NCMN000
`
	m, err := NewAccentuationDataLoader(false).LoadWords(strings.NewReader(in), "test")
	require.NoError(t, err)
	require.Contains(t, m, "cafe")
	require.Equal(t, "café", m["cafe"].GetToken())
	require.True(t, m["cafe"].HasPosTag("NCMN000"))
}

func TestAccentuationDataLoader_Multi(t *testing.T) {
	in := `x;a;A
x;b;B
`
	m, err := NewAccentuationDataLoader(true).LoadWords(strings.NewReader(in), "test")
	require.NoError(t, err)
	require.Equal(t, 2, m["x"].GetReadingsLength())
	require.True(t, m["x"].HasPosTag("A"))
	require.True(t, m["x"].HasPosTag("B"))
}

func TestAccentuationDataLoader_Replace(t *testing.T) {
	in := `x;a;A
x;b;B
`
	m, err := NewAccentuationDataLoader(false).LoadWords(strings.NewReader(in), "test")
	require.NoError(t, err)
	require.Equal(t, 1, m["x"].GetReadingsLength())
	require.True(t, m["x"].HasPosTag("B"))
}

func TestAccentuationDataLoader_BadFormat(t *testing.T) {
	_, err := NewAccentuationDataLoader(false).LoadWords(strings.NewReader("only;two\n"), "f")
	require.Error(t, err)
}
