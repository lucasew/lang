package translation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranslation(t *testing.T) {
	src := NewDataSource("http://license", "dict", "http://src")
	tr := NewMapTranslator(src)
	tr.Add("house", "en", "de", NewTranslationEntry([]string{"house"}, []string{"Haus"}, 1))
	got, err := tr.Translate("house", "en", "de")
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, []string{"Haus"}, got[0].GetL2())
	require.Equal(t, "foo", tr.CleanTranslationForReplace("foo (note)", ""))
	require.Equal(t, "(note)", tr.GetTranslationSuffix("foo (note)"))
}
