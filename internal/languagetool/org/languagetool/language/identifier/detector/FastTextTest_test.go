package detector

// Twin of FastTextTest.testParsing
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFastText_Parsing(t *testing.T) {
	ft := NewFastTextDetectorForTest()
	l := []string{"en", "fy", "de", "es", "nl"}

	res1, err := ft.ParseBuffer("__label__nl 0.423696 __label__fy 0.207109\n", l)
	require.NoError(t, err)
	require.Equal(t, 2, len(res1))
	require.Equal(t, 0.423696, res1["nl"])
	require.Equal(t, 0.207109, res1["fy"])

	res2, err := ft.ParseBuffer("__label__de 0.999985 __label__es 2.02195e-05", l)
	require.NoError(t, err)
	require.Equal(t, 2, len(res2))
	require.Equal(t, 0.999985, res2["de"])
	require.InDelta(t, 2.02195e-05, res2["es"], 1e-12)

	res3, err := ft.ParseBuffer("__label__en 1", l)
	require.NoError(t, err)
	require.Equal(t, 1, len(res3))
	require.Equal(t, 1.0, res3["en"])

	res4, err := ft.ParseBuffer("__label__de 1.00003", l)
	require.NoError(t, err)
	require.Equal(t, 1, len(res4))
	require.Equal(t, 1.00003, res4["de"])

	_, err = ft.ParseBuffer("xxx", l)
	require.Error(t, err)
	_, err = ft.ParseBuffer("xxx foo", l)
	require.Error(t, err)

	// multi-line: Java canLanguageBeDetected keeps supported langs even if not in additional list.
	// fr is a built-in supported short code → both de and fr kept (not invent only-additional).
	res5, err := ft.ParseBuffer("__label__de 0.9\n__label__fr 0.1", l)
	require.NoError(t, err)
	require.Equal(t, 2, len(res5))
	require.Equal(t, 0.9, res5["de"])
	require.Equal(t, 0.1, res5["fr"])

	// unsupported code only via additional
	res6, err := ft.ParseBuffer("__label__de 0.9\n__label__zz 0.1", []string{"de"})
	require.NoError(t, err)
	require.Contains(t, res6, "de")
	require.NotContains(t, res6, "zz")
	res7, err := ft.ParseBuffer("__label__de 0.9\n__label__zz 0.1", []string{"zz"})
	require.NoError(t, err)
	require.Contains(t, res7, "de") // still supported
	require.Contains(t, res7, "zz") // additional noop
}
