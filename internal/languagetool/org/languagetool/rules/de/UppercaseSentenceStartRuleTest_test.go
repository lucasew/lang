package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UppercaseSentenceStartRuleTest.java
// Multi-sentence splits use SplitAndAnalyze (simple .!?) — not full DE SRX (bspw./z.B. need SRX sector).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUppercaseSentenceStartRule_Rule(t *testing.T) {
	r := NewUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "This sentence does not start with an uppercase letter",
		"category_case":  "Capitalization",
		"desc_uppercase_sentence": "Checks that a sentence starts with an uppercase letter",
	})
	require.Equal(t, "UPPERCASE_SENTENCE_START", r.GetID())
	require.Contains(t, r.GetURL(), "gross-klein")

	analyze := func(s string) []*languagetool.AnalyzedSentence {
		if strings.Contains(s, ". ") || strings.Contains(s, "! ") || strings.Contains(s, "? ") ||
			strings.Contains(s, ".\n") || strings.Contains(s, "\n") {
			return languagetool.SplitAndAnalyze(s)
		}
		return []*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}
	}
	// Java good
	require.Empty(t, r.MatchList(analyze("Dies ist ein Satz. Und hier kommt noch einer")))
	require.Empty(t, r.MatchList(analyze("Dies ist ein Satz. Ätsch, noch einer mit Umlaut.")))
	require.Empty(t, r.MatchList(analyze("\"Dies ist ein Satz!\"")))
	require.Empty(t, r.MatchList(analyze("mRNA-Impfstoffe sind wichtig.")))
	require.Empty(t, r.MatchList(analyze("Willkommen… im Berlin.")))
	// single word not a real sentence (Java special case: tokens length == 2)
	require.Empty(t, r.MatchList(analyze("schön")))
	require.Empty(t, r.MatchList(analyze("Satz")))

	// Java bad
	require.Equal(t, 2, len(r.MatchList(analyze("etwas beginnen. und der auch nicht"))))
	require.Equal(t, 1, len(r.MatchList(analyze("schön!"))))
	require.Equal(t, 1, len(r.MatchList(analyze("Dies ist ein Satz. ätsch, noch einer mit Umlaut."))))
	require.Equal(t, 1, len(r.MatchList(analyze("Dies ist ein Satz. \"aber der hier auch!\""))))
	require.Equal(t, 1, len(r.MatchList(analyze("Dies ist ein Satz. „aber der hier auch!“"))))
	require.Equal(t, 1, len(r.MatchList(analyze("\"dies ist ein Satz!\""))))

	// positions: "Ein Test. was?" — Java from=10 to=13 on full text
	// Use SplitAndAnalyze; offset = first sentence CorrectedTextLength
	sents := languagetool.SplitAndAnalyze("Ein Test. was?")
	require.GreaterOrEqual(t, len(sents), 2)
	ms := r.MatchList(sents)
	require.Equal(t, 1, len(ms))
	off := sents[0].GetCorrectedTextLength()
	// second sentence "was?" local start of "was"
	toks := sents[1].GetTokensWithoutWhitespace()
	var was *languagetool.AnalyzedTokenReadings
	for _, t := range toks {
		if t != nil && t.GetToken() == "was" {
			was = t
			break
		}
	}
	require.NotNil(t, was)
	require.Equal(t, off+was.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, off+was.GetEndPos(), ms[0].GetToPos())
}
