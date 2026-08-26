package wikipedia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextMapFilter_Simple(t *testing.T) {
	var f TextMapFilter = NewSimpleWikipediaTextFilter()
	m := f.FilterMapped("foo [[Bar|baz]]")
	require.Equal(t, "foo baz", m.GetPlainText())
	require.Equal(t, "foo [[Bar|baz]]", m.GetOriginal())
}
