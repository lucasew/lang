package en

// Twin of EnglishSynthesizerTest — manual map + determiner specials (full en.dict deferred)
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func dummyToken(tokenStr, lemma string) *languagetool.AnalyzedToken {
	if lemma == "" {
		lemma = tokenStr
	}
	return languagetool.NewAnalyzedToken(tokenStr, strp(tokenStr), strp(lemma))
}

func TestEnglishSynthesizer_SynthesizeStringString(t *testing.T) {
	// Manual synth format: form\tlemma\tpos
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(strings.Join([]string{
		"was\tbe\tVBD",
		"were\tbe\tVBD",
		"presidents\tpresident\tNNS",
		"tested\ttest\tVBD",
		"testing\ttest\tVBG",
		"absolutized\tabsolutize\tVBD",
		"mixed\tmix\tVBD",
		"mixed\tmix\tVBN",
		"is\tbe\tVBZ",
		"my\tI\tPRP$_A1S",
		"I\tI\tPRP_S1S",
	}, "\n") + "\n"))
	require.NoError(t, err)
	// removal: Christmas VBZ
	removal, err := synthesis.NewManualSynthesizer(strings.NewReader("is\tChristmas\tVBZ\n"))
	require.NoError(t, err)

	synth := NewEnglishSynthesizer(manual)
	synth.Removal = removal

	got, err := synth.Synthesize(dummyToken("blablabla", ""), "blablabla")
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = synth.Synthesize(dummyToken("be", "be"), "VBD")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"was", "were"}, got)

	got, err = synth.Synthesize(dummyToken("president", "president"), "NNS")
	require.NoError(t, err)
	require.Equal(t, []string{"presidents"}, got)

	got, err = synth.Synthesize(dummyToken("test", "test"), "VBD")
	require.NoError(t, err)
	require.Equal(t, []string{"tested"}, got)

	got, err = synth.SynthesizeRE(dummyToken("test", "test"), "VBD|VBG", true)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"tested", "testing"}, got)

	// determiners
	got, err = synth.Synthesize(dummyToken("university", "university"), AddDeterminer)
	require.NoError(t, err)
	require.Equal(t, []string{"a university", "the university"}, got)

	got, err = synth.Synthesize(dummyToken("hour", "hour"), AddDeterminer)
	require.NoError(t, err)
	require.Equal(t, []string{"an hour", "the hour"}, got)

	got, err = synth.Synthesize(dummyToken("hour", "hour"), AddIndDeterminer)
	require.NoError(t, err)
	require.Equal(t, []string{"an hour"}, got)

	// removed Christmas VBZ
	got, err = synth.Synthesize(dummyToken("Christmas", "Christmas"), "VBZ")
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = synth.Synthesize(dummyToken("mix", "mix"), "VBD")
	require.NoError(t, err)
	require.Equal(t, []string{"mixed"}, got)
}
