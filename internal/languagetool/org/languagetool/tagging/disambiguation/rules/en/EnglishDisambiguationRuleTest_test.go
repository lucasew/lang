package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

func TestEnglishDisambiguationRule_ChunkerInject(t *testing.T) {
	c := disambiguation.NewMultiWordChunker([]string{"New York\tB-NP"}, disambiguation.MultiWordChunkerSettings{AllowFirstCapitalized: true})
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(" ", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", nil, nil)),
	}
	out := c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
	require.NotNil(t, out)
}

// Twin of EnglishDisambiguationRuleTest.testChunker hybrid path for multiwords.
// Java: TestTools.myAssert("The quid pro quo.", ... hybridDisam)
func TestEnglishDisambiguationRule_HybridQuidProQuo(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("The quid pro quo.")
	// quid/pro/quo should share multiword lemma "quid pro quo" and NN
	for _, want := range []string{"quid", "pro", "quo"} {
		tok := findTok(sent, want)
		require.NotNil(t, tok, want)
		require.True(t, hasLemmaPOS(tok, "quid pro quo", "NN"),
			"%s tags=%s", want, dumpTags(tok))
	}
}

// Twin of EnglishDisambiguationRuleTest hybrid uppercase multiword.
func TestEnglishDisambiguationRule_HybridQuidProQuoUpper(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("The QUID PRO QUO.")
	for _, want := range []string{"QUID", "PRO", "QUO"} {
		tok := findTok(sent, want)
		require.NotNil(t, tok, want)
		require.True(t, hasLemmaPOS(tok, "quid pro quo", "NN"),
			"%s tags=%s", want, dumpTags(tok))
	}
}

// Twin of disambiguation.xml QUARAN example: Qur'an → an[NNP]
func TestEnglishDisambiguationRule_QuranAnNNP(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("Qur'an.")
	an := findTok(sent, "an")
	require.NotNil(t, an)
	require.True(t, an.IsIgnoredBySpeller(), "multiword ignore-spelling")
	require.True(t, hasPOS(an, "NNP"), "QUARAN / multiword NNP: %s", dumpTags(an))
}

// Twin of UNKNOWN_PCT via hybrid XML path (also covered in tagging/en).
func TestEnglishDisambiguationRule_UnknownPCT(t *testing.T) {
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree")
	}
	sent := tagen.AnalyzeEnglishSentence("Hello.")
	dot := findTok(sent, ".")
	require.NotNil(t, dot)
	require.True(t, hasPOS(dot, "PCT"), dumpTags(dot))
}

func findTok(sent *languagetool.AnalyzedSentence, surface string) *languagetool.AnalyzedTokenReadings {
	if sent == nil {
		return nil
	}
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == surface {
			return tok
		}
	}
	return nil
}

func hasPOS(tok *languagetool.AnalyzedTokenReadings, pos string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && *r.GetPOSTag() == pos {
			return true
		}
	}
	return false
}

func hasLemmaPOS(tok *languagetool.AnalyzedTokenReadings, lemma, pos string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		p, l := "", ""
		if r.GetPOSTag() != nil {
			p = *r.GetPOSTag()
		}
		if r.GetLemma() != nil {
			l = *r.GetLemma()
		}
		if p == pos && l == lemma {
			return true
		}
	}
	return false
}

func dumpTags(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	var b strings.Builder
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		p, l := "", ""
		if r.GetPOSTag() != nil {
			p = *r.GetPOSTag()
		}
		if r.GetLemma() != nil {
			l = *r.GetLemma()
		}
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString(l)
		b.WriteByte('/')
		b.WriteString(p)
	}
	return b.String()
}
