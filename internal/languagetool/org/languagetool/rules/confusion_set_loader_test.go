package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfusionSetLoader_AlphabetOrder(t *testing.T) {
	loader := NewConfusionSetLoader(nil)
	_, err := loader.LoadConfusionPairs(strings.NewReader("zebra; apple; 1\n"))
	require.Error(t, err)
}

func TestConfusionSetLoader_WordDefsHook(t *testing.T) {
	loader := NewConfusionSetLoader(func(word string) *string {
		if word == "a" {
			d := "letter a"
			return &d
		}
		return nil
	})
	m, err := loader.LoadConfusionPairs(strings.NewReader("a; b; 3\n"))
	require.NoError(t, err)
	require.NotNil(t, m["a"][0].GetTerm1().GetDescription())
	require.Equal(t, "letter a", *m["a"][0].GetTerm1().GetDescription())
}
