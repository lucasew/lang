package languagetool

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tokenizers.ManualTaggerAdapterTest.

const manualTaggerTestData = `# some test data
inflectedform11	lemma1	POS1
inflectedform121	lemma1	POS2
inflectedform122	lemma1	POS2
inflectedform123	lemma1	POS3
inflectedform2	lemma2	POS1a
inflectedform2	lemma2	POS1b
inflectedform2	lemma2	POS1c
inflectedform3	lemma3a	POS3a
inflectedform3	lemma3b	POS3b
inflectedform3	lemma3c	POS3c
inflectedform3	lemma3d	POS3d
`

func newManualTaggerAdapter(t *testing.T) *ManualTaggerAdapter {
	t.Helper()
	mt, err := tagging.NewManualTagger(strings.NewReader(manualTaggerTestData))
	require.NoError(t, err)
	return NewManualTaggerAdapter(mt)
}

func TestManualTaggerAdapter_MultipleLemma(t *testing.T) {
	tagger := newManualTaggerAdapter(t)
	analyzed := tagger.Tag([]string{"inflectedform3"})
	require.NotNil(t, analyzed)
	require.Len(t, analyzed, 1)
	atr := analyzed[0]
	require.Equal(t, "inflectedform3", atr.GetToken())
	require.Equal(t, 4, atr.GetReadingsLength())
	r := atr.GetReadings()
	require.Equal(t, "lemma3a", *r[0].GetLemma())
	require.Equal(t, "POS3a", *r[0].GetPOSTag())
	require.Equal(t, "lemma3b", *r[1].GetLemma())
	require.Equal(t, "POS3b", *r[1].GetPOSTag())
	require.Equal(t, "lemma3c", *r[2].GetLemma())
	require.Equal(t, "POS3c", *r[2].GetPOSTag())
	require.Equal(t, "lemma3d", *r[3].GetLemma())
	require.Equal(t, "POS3d", *r[3].GetPOSTag())
}

func TestManualTaggerAdapter_MultiplePOS(t *testing.T) {
	tagger := newManualTaggerAdapter(t)
	analyzed := tagger.Tag([]string{"inflectedform2"})
	require.Len(t, analyzed, 1)
	require.Equal(t, 3, analyzed[0].GetReadingsLength())
	r := analyzed[0].GetReadings()
	require.Equal(t, "POS1a", *r[0].GetPOSTag())
	require.Equal(t, "lemma2", *r[0].GetLemma())
	require.Equal(t, "POS1b", *r[1].GetPOSTag())
	require.Equal(t, "POS1c", *r[2].GetPOSTag())
}

func TestManualTaggerAdapter_MultipleWords(t *testing.T) {
	tagger := newManualTaggerAdapter(t)
	analyzed := tagger.Tag([]string{"inflectedform2", "inflectedform3"})
	require.Len(t, analyzed, 2)
	require.Equal(t, 3, analyzed[0].GetReadingsLength())
	require.Equal(t, 4, analyzed[1].GetReadingsLength())
}
