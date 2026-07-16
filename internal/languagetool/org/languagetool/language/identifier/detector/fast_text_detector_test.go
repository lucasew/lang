package detector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFastTextParseBuffer(t *testing.T) {
	d := NewFastTextDetectorForTest()
	m, err := d.ParseBuffer("__label__en 0.9 __label__de 0.1", nil)
	require.NoError(t, err)
	require.InDelta(t, 0.9, m["en"], 1e-9)
	require.InDelta(t, 0.1, m["de"], 1e-9)

	_, err = d.ParseBuffer("nope", nil)
	require.Error(t, err)

	d.Runner = func(line string) (string, error) {
		return "__label__fr 0.8 __label__en 0.2\n", nil
	}
	m, err = d.RunFasttext("bonjour", nil)
	require.NoError(t, err)
	require.InDelta(t, 0.8, m["fr"], 1e-9)
}

func TestFastTextCanDetect(t *testing.T) {
	d := NewFastTextDetectorForTest()
	d.CanDetect = func(code string, _ []string) bool { return code == "en" }
	m, err := d.ParseBuffer("__label__en 0.9 __label__de 0.1", nil)
	require.NoError(t, err)
	require.Contains(t, m, "en")
	require.NotContains(t, m, "de")
}
