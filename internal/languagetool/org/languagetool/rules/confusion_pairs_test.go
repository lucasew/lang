package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfusionPairsAndSuggest(t *testing.T) {
	data := `
# comment
como;cómo;NCMS000
titulo;título;NCMS000
`
	pairs, err := LoadConfusionPairs(strings.NewReader(data))
	require.NoError(t, err)
	require.Len(t, pairs, 2)

	f := &ConfusionCheckFilter{
		Pairs:              pairs,
		MessageDiacritic:   "se escribe con tilde",
		MessageNoDiacritic: "se escribe de otra manera",
	}
	res := f.Suggest("como", "NCMS000", "", "x se escribe con tilde y", "{suggestion}")
	require.True(t, res.OK)
	require.Equal(t, "cómo", res.Replacement)

	res = f.Suggest("Como", "NCMS000", "", "msg", "{suggestion}")
	require.True(t, res.OK)
	require.Equal(t, "Cómo", res.Replacement)

	res = f.Suggest("xyz", "NCMS000", "", "m", "{suggestion}")
	require.False(t, res.OK)
}
