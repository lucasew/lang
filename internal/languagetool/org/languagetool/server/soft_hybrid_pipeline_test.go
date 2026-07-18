package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Soft hybrid on the server path should match commandline configureCoreLT
// (multiwords + soft disambig for non-EN).
func TestPipeline_SoftHybrid_FR_Multiword(t *testing.T) {
	p := NewPipeline(NewPipelineSettings("fr", "u"))
	lt := p.newConfiguredLT()
	require.NotNil(t, lt)
	require.NotNil(t, lt.Disambiguator, "non-EN soft hybrid should be registered")
	// home page multiword from FR multiwords.txt
	sents := lt.Analyze("La home page est belle.")
	require.NotEmpty(t, sents)
	found := false
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil && (
				*r.GetPOSTag() == "N f s" || *r.GetPOSTag() == "J f s" ||
					*r.GetPOSTag() == "<N f s>" || *r.GetPOSTag() == "</N f s>") {
				found = true
			}
		}
	}
	require.True(t, found, "FR multiword home page should tag after hybrid")
}
