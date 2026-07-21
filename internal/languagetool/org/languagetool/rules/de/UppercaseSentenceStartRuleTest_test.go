package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UppercaseSentenceStartRuleTest.java
// Multi-sentence analysis uses GermanSRXSentenceTokenizer (Java JLT sentence split), not naive SplitAndAnalyze.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	tokde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
	"github.com/stretchr/testify/require"
)

// analyzeDESentence splits with German SRX then AnalyzePlain each segment (sentence-local positions;
// MatchList adds CorrectedTextLength offsets like Java TextLevelRule).
func analyzeDESentence(text string) []*languagetool.AnalyzedSentence {
	parts := tokde.NewGermanSRXSentenceTokenizer().Tokenize(text)
	if len(parts) == 0 {
		return nil
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, languagetool.AnalyzePlain(p))
	}
	return out
}

func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	r := NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case":          "This sentence does not start with an uppercase letter",
		"category_case":           "Capitalization",
		"desc_uppercase_sentence": "Checks that a sentence starts with an uppercase letter",
	})
	require.Equal(t, "UPPERCASE_SENTENCE_START", r.GetID())
	require.Contains(t, r.GetURL(), "gross-klein")

	// Java good
	require.Empty(t, r.MatchList(analyzeDESentence("Dies ist ein Satz. Und hier kommt noch einer")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dies ist ein Satz. Ätsch, noch einer mit Umlaut.")))
	require.Empty(t, r.MatchList(analyzeDESentence("\"Dies ist ein Satz!\"")))
	require.Empty(t, r.MatchList(analyzeDESentence("mRNA-Impfstoffe sind wichtig.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Willkommen… im Berlin.")))
	require.Empty(t, r.MatchList(analyzeDESentence("schön")))
	require.Empty(t, r.MatchList(analyzeDESentence("Satz")))
	// abbreviations: no false sentence break → no false uppercase hit
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist bspw. okay so.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist z.B. okay so.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist z. B. okay so.")))

	// Java bad
	require.Equal(t, 2, len(r.MatchList(analyzeDESentence("etwas beginnen. und der auch nicht"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("schön!"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. ätsch, noch einer mit Umlaut."))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. \"aber der hier auch!\""))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. „aber der hier auch!“"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("\"dies ist ein Satz!\""))))

	// positions: "Ein Test. was?"
	sents := analyzeDESentence("Ein Test. was?")
	require.GreaterOrEqual(t, len(sents), 2)
	ms := r.MatchList(sents)
	require.Equal(t, 1, len(ms))
	off := sents[0].GetCorrectedTextLength()
	var was *languagetool.AnalyzedTokenReadings
	for _, tok := range sents[1].GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "was" {
			was = tok
			break
		}
	}
	require.NotNil(t, was)
	require.Equal(t, off+was.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, off+was.GetEndPos(), ms[0].GetToPos())
}
