package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core ConfusionSetLoaderTest.testLoadWithStrictLimits
// (uses the yy demo confusion_sets fixture, inlined).
func TestConfusionSetLoader_languagetool_core_LoadWithStrictLimits(t *testing.T) {
	sample := `# test confusion set file
their|example 2; there|example 1; 10  # this is a comment
bar; foo; 5
baz; foo; 8
goo; lol; 11
goo; something; 12
one -> two; 13
three -> four; 14
four -> five; 15
im; um; 1  # German
`
	loader := NewConfusionSetLoader(nil)
	m, err := loader.LoadConfusionPairs(strings.NewReader(sample))
	require.NoError(t, err)
	require.Equal(t, 15, len(m))

	require.Len(t, m["there"], 2)
	require.Equal(t, int64(10), m["there"][0].GetFactor())
	require.Len(t, m["their"], 2)
	require.Equal(t, int64(10), m["their"][0].GetFactor())

	require.Len(t, m["foo"], 4)
	require.Equal(t, int64(5), m["foo"][0].GetFactor())
	require.Equal(t, int64(5), m["foo"][1].GetFactor())
	require.Equal(t, int64(8), m["foo"][2].GetFactor())
	require.Equal(t, int64(8), m["foo"][3].GetFactor())

	require.Len(t, m["goo"], 4)
	require.Equal(t, int64(11), m["goo"][0].GetFactor())
	require.Equal(t, int64(12), m["goo"][2].GetFactor())
	require.Len(t, m["lol"], 2)
	require.Len(t, m["something"], 2)

	require.Len(t, m["bar"], 2)
	require.Equal(t, int64(5), m["bar"][0].GetFactor())
	require.Equal(t, "bar", m["bar"][0].GetTerm1().GetString())
	require.Equal(t, "foo", m["bar"][0].GetTerm2().GetString())
	require.Equal(t, "foo", m["bar"][1].GetTerm1().GetString())
	require.Equal(t, "bar", m["bar"][1].GetTerm2().GetString())

	require.Len(t, m["one"], 1)
	require.Equal(t, int64(13), m["one"][0].GetFactor())
	require.Equal(t, "one", m["one"][0].GetTerm1().GetString())
	require.Equal(t, "two", m["one"][0].GetTerm2().GetString())

	require.Len(t, m["three"], 1)
	require.Equal(t, int64(14), m["three"][0].GetFactor())
	require.Len(t, m["four"], 2)
	require.Equal(t, int64(14), m["four"][0].GetFactor())
	require.Equal(t, int64(15), m["four"][1].GetFactor())
	require.Equal(t, "five", m["four"][1].GetTerm2().GetString())

	thereTerms := m["there"][0].GetTerms()
	var descParts []string
	for _, cs := range thereTerms {
		d := ""
		if cs.GetDescription() != nil {
			d = *cs.GetDescription()
		}
		descParts = append(descParts, cs.GetString()+" - "+d)
	}
	joined := strings.Join(descParts, " ")
	require.Contains(t, joined, "there - example 1")
	require.Contains(t, joined, "their - example 2")
	require.NotContains(t, joined, "comment")
}
