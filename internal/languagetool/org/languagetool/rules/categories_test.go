package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategories(t *testing.T) {
	require.Equal(t, "GRAMMAR", CatGrammar.GetID().String())
	c := CatGrammar.GetCategory(map[string]string{"category_grammar": "Grammar"})
	require.Equal(t, "Grammar", c.GetName())
	c2 := CatTypos.GetCategory(nil)
	require.Equal(t, "TYPOS", c2.GetName())
	require.NotEmpty(t, AllCategories)
	require.Equal(t, CatStyle.ID, AllCategories[0].ID)
}
