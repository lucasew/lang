package synthesis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.synthesis.ManualSynthesizerTest.
// Fixture matches Java @Before data (tab-separated fullform/baseform/postags).

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

// testLookupNonExisting — Java ManualSynthesizerTest.testLookupNonExisting
func TestManualSynthesizer_LookupNonExisting(t *testing.T) {
	s := newSynth(t)
	require.Nil(t, s.Lookup("", ""))
	require.Nil(t, s.LookupPtr(strPtr(""), nil))
	require.Nil(t, s.LookupPtr(nil, strPtr("")))
	require.Nil(t, s.LookupPtr(nil, nil))
	require.Nil(t, s.Lookup("NONE", "UNKNOWN"))
}

// testInvalidLookup — Java ManualSynthesizerTest.testInvalidLookup
func TestManualSynthesizer_InvalidLookup(t *testing.T) {
	s := newSynth(t)
	require.Nil(t, s.Lookup("NONE", "POS1"))
	require.Nil(t, s.Lookup("Lemma1", "UNKNOWN"))
	// no reg exp on POS: "POS." must not match "POS1"/"POS2"
	require.Nil(t, s.Lookup("Lemma1", "POS."))
	require.Nil(t, s.Lookup("Lemma2", "POS2"))
}

// testValidLookup — Java ManualSynthesizerTest.testValidLookup
func TestManualSynthesizer_ValidLookup(t *testing.T) {
	s := newSynth(t)
	require.Equal(t, "[InflectedForm11]", listStr(s.Lookup("Lemma1", "POS1")))
	require.Equal(t, "[InflectedForm121, InflectedForm122]", listStr(s.Lookup("Lemma1", "POS2")))
	require.Equal(t, "[InflectedForm2]", listStr(s.Lookup("Lemma2", "POS1")))
}

// testCaseSensitive — Java ManualSynthesizerTest.testCaseSensitive
func TestManualSynthesizer_CaseSensitive(t *testing.T) {
	s := newSynth(t)
	// lookup is case sensitive:
	require.Nil(t, s.Lookup("LEmma1", "POS1"))
}

// listStr formats like Java List.toString() for twin asserts.
func listStr(ss []string) string {
	if ss == nil {
		return "null"
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

func strPtr(s string) *string { return &s }

// Extra outcomes grounded in Java encodeForm/decodeForm and getPossibleTags
// (official ManualSynthesizerTest does not cover these paths).

func TestManualSynthesizer_EncodeDecodeSuffix(t *testing.T) {
	// word extends lemma → encode "+" + suffix; lookup returns full form
	data := "dogs\tdog\tNNS\n" +
		"ran\trun\tVBD\n" + // no shared prefix enough for +; full form stored
		"ac\tab\tX\n" // ++ stem-change style: "ab"→"ac"
	s, err := NewManualSynthesizer(strings.NewReader(data))
	require.NoError(t, err)

	require.Equal(t, []string{"dogs"}, s.Lookup("dog", "NNS"))
	require.Equal(t, []string{"ran"}, s.Lookup("run", "VBD"))
	require.Equal(t, []string{"ac"}, s.Lookup("ab", "X"))

	// unit-level encode/decode identity for the paths above
	require.Equal(t, "+s", encodeManualSynthForm("dog", "dogs"))
	require.Equal(t, "dogs", decodeManualSynthForm("dog", "+s"))
	require.Equal(t, "ran", encodeManualSynthForm("run", "ran"))
	require.Equal(t, "++c", encodeManualSynthForm("ab", "ac"))
	require.Equal(t, "ac", decodeManualSynthForm("ab", "++c"))
	// form equals lemma → "++" + last char
	require.Equal(t, "++x", encodeManualSynthForm("x", "x"))
	require.Equal(t, "x", decodeManualSynthForm("x", "++x"))
}

func TestManualSynthesizer_RejectFormStartingWithPlus(t *testing.T) {
	// Java groupByHash: words can't start with SUFFIX_MARKER
	_, err := NewManualSynthesizer(strings.NewReader("+foo\tbar\tN\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "Words can't start with")
}

func TestManualSynthesizer_GetPossibleTags(t *testing.T) {
	s := newSynth(t)
	tags := s.GetPossibleTags()
	require.Contains(t, tags, "POS1")
	require.Contains(t, tags, "POS2")
	require.NotContains(t, tags, "UNKNOWN")
	require.Len(t, tags, 2)
}

func TestManualSynthesizer_SeparatorRegExp(t *testing.T) {
	// Java loadMapping honors #separatorRegExp=
	data := "#separatorRegExp=;\n" +
		"forms;lemma;TAG\n"
	s, err := NewManualSynthesizer(strings.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, []string{"forms"}, s.Lookup("lemma", "TAG"))
}

func TestManualSynthesizer_CommentAndBlankLines(t *testing.T) {
	data := "# header\n" +
		"\n" +
		"  \n" +
		"a\tb\tC # trailing comment\n"
	s, err := NewManualSynthesizer(strings.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, []string{"a"}, s.Lookup("b", "C"))
}

func TestManualSynthesizer_MultiFormOrder(t *testing.T) {
	// file order preserved for multi-form same lemma+POS (Java ArrayList order)
	data := "f1\tl\tP\n" +
		"f2\tl\tP\n" +
		"f3\tl\tP\n"
	s, err := NewManualSynthesizer(strings.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, []string{"f1", "f2", "f3"}, s.Lookup("l", "P"))
}
