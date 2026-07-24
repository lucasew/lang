package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryId(t *testing.T) {
	a := NewCategoryId("GRAMMAR")
	b := NewCategoryId("GRAMMAR")
	require.True(t, a.Equals(b))
	require.Equal(t, "GRAMMAR", a.String())
	require.Panics(t, func() { NewCategoryId("") })
	require.Panics(t, func() { NewCategoryId("   ") })
	// Java String.trim: control chars <= ' ' count as empty
	require.Panics(t, func() { NewCategoryId("\x01\x02") })
	// NBSP is not stripped by String.trim — valid CategoryId
	require.NotPanics(t, func() { NewCategoryId("\u00a0") })
}

func TestCategory(t *testing.T) {
	c := NewCategory(CategoryGrammar, "Grammar")
	require.Equal(t, "Grammar", c.GetName())
	require.False(t, c.IsDefaultOff())
	require.Equal(t, CategoryInternal, c.GetLocation())
	off := NewCategoryFull(CategoryStyle, "Style", CategoryExternal, false, "StyleTab")
	require.True(t, off.IsDefaultOff())
	require.Equal(t, "StyleTab", off.GetTabName())
	require.True(t, CategoryTypos.Equals(NewCategoryId("TYPOS")))
}
