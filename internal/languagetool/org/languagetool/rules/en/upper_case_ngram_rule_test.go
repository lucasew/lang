package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_Meta(t *testing.T) {
	r := newUpperCaseNgramRuleForTest()
	require.Equal(t, "EN_UPPER_CASE_NGRAM", r.GetID())
	require.Equal(t, "Checks wrong uppercase spelling of words that are not proper nouns", r.GetDescription())
	require.Equal(t, 50, THRESHOLD)
	require.Equal(t, 50, UpperCaseNgramThreshold)
}
