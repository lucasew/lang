package tagging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapWordTagger(t *testing.T) {
	wt := MapWordTagger{"cats": {NewTaggedWord("cat", "NNS")}}
	require.Equal(t, "cat", wt.Tag("cats")[0].GetLemma())
	require.Empty(t, wt.Tag("dogs"))
}
