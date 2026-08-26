package tokenizers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of SrxTools.createSrxDocument + SRXSentenceTokenizer path selection.
func TestSrxTools_CreateSrxDocument(t *testing.T) {
	doc, err := createSrxDocument("/segment.srx")
	require.NoError(t, err)
	require.NotNil(t, doc)

	docSimple, err := createSrxDocument("/org/languagetool/tokenizers/segment-simple.srx")
	require.NoError(t, err)
	require.NotNil(t, docSimple)

	// normalize leading slash
	doc2, err := createSrxDocument("segment.srx")
	require.NoError(t, err)
	require.NotNil(t, doc2)
}

func TestSrxTools_TokenizeWithCreatedDocument(t *testing.T) {
	// Java SrxTools.tokenize(text, doc, languageCode+parCode)
	doc, err := cachedCreateSrxDocument("/org/languagetool/tokenizers/segment-simple.srx")
	require.NoError(t, err)
	segs := doc.Split("Hi! This is a test.", "xx", "_two")
	require.GreaterOrEqual(t, len(segs), 2)
	require.Equal(t, "Hi! ", segs[0])
}
