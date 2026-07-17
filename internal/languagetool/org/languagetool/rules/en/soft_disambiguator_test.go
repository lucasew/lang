package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterSoftEnglishDisambiguator_Multiword(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	// tagger optional
	if p := findEnglishPOSDict(t); p != "" {
		require.True(t, RegisterBinaryEnglishTagger(lt, p))
	}
	RegisterSoftEnglishDisambiguator(lt, "")
	require.NotNil(t, lt.Disambiguator)
	sents := lt.Analyze("I live in New York.")
	require.NotEmpty(t, sents)
	// find New or York readings with NNP multiword tag
	var sawNNP bool
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil {
				continue
			}
			w := tok.GetToken()
			if w != "New" && w != "York" {
				continue
			}
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at == nil || at.GetPOSTag() == nil {
					continue
				}
				if *at.GetPOSTag() == "NNP" || *at.GetPOSTag() == "B-NNP" || *at.GetPOSTag() == "E-NNP" {
					sawNNP = true
				}
			}
		}
	}
	// MultiWordChunker may tag with B-/E- or full NNP depending on path
	require.True(t, sawNNP || multiwordTagged(sents), "expected multiword tags on New York, sents=%s", dumpSents(sents))
}

func multiwordTagged(sents []*languagetool.AnalyzedSentence) bool {
	for _, s := range sents {
		for _, tok := range s.GetTokens() {
			if tok == nil {
				continue
			}
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at == nil || at.GetPOSTag() == nil {
					continue
				}
				p := *at.GetPOSTag()
				if p == "NNP" || len(p) > 2 && (p[0] == 'B' || p[0] == 'E') && (containsNNP(p)) {
					return true
				}
			}
		}
	}
	return false
}

func containsNNP(p string) bool {
	return len(p) >= 3 && (p == "NNP" || p[len(p)-3:] == "NNP" || p == "B-NP" || p == "E-NP" || p == "B-NNP" || p == "E-NNP")
}

func dumpSents(sents []*languagetool.AnalyzedSentence) string {
	var b string
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil {
				continue
			}
			b += tok.GetToken() + ":"
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at != nil && at.GetPOSTag() != nil {
					b += *at.GetPOSTag() + ","
				}
			}
			b += " "
		}
	}
	return b
}
