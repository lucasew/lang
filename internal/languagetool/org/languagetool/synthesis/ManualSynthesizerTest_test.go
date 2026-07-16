package synthesis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.synthesis.ManualSynthesizerTest.

const synthTestData = `# some test data
InflectedForm11	Lemma1	POS1
InflectedForm121	Lemma1	POS2
InflectedForm122	Lemma1	POS2
InflectedForm2	Lemma2	POS1
`

func newSynth(t *testing.T) *ManualSynthesizer {
	t.Helper()
	s, err := NewManualSynthesizer(strings.NewReader(synthTestData))
	require.NoError(t, err)
	return s
}

func TestManualSynthesizer_LookupNonExisting(t *testing.T) {
	s := newSynth(t)
	require.Nil(t, s.Lookup("", ""))
	require.Nil(t, s.LookupPtr(strPtr(""), nil))
	require.Nil(t, s.LookupPtr(nil, strPtr("")))
	require.Nil(t, s.LookupPtr(nil, nil))
	require.Nil(t, s.Lookup("NONE", "UNKNOWN"))
}

func TestManualSynthesizer_InvalidLookup(t *testing.T) {
	s := newSynth(t)
	require.Nil(t, s.Lookup("NONE", "POS1"))
	require.Nil(t, s.Lookup("Lemma1", "UNKNOWN"))
	require.Nil(t, s.Lookup("Lemma1", "POS."))
	require.Nil(t, s.Lookup("Lemma2", "POS2"))
}

func TestManualSynthesizer_ValidLookup(t *testing.T) {
	s := newSynth(t)
	require.Equal(t, "[InflectedForm11]", listStr(s.Lookup("Lemma1", "POS1")))
	require.Equal(t, "[InflectedForm121, InflectedForm122]", listStr(s.Lookup("Lemma1", "POS2")))
	require.Equal(t, "[InflectedForm2]", listStr(s.Lookup("Lemma2", "POS1")))
}

func listStr(ss []string) string {
	if ss == nil {
		return "null"
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

func TestManualSynthesizer_CaseSensitive(t *testing.T) {
	s := newSynth(t)
	require.Nil(t, s.Lookup("LEmma1", "POS1"))
}

func strPtr(s string) *string { return &s }
