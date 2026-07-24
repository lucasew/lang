package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseAccentuationCheckRule_DetNoun(t *testing.T) {
	r := NewPortugueseAccentuationCheckRule()
	require.Equal(t, "ACCENTUATION_CHECK_PT", r.GetID())
	pos := "NCFS000"
	accented := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("cópia", &pos, nil))
	r.VerbToNoun["copia"] = accented

	det := "DA0FS0"
	ss := languagetool.SentenceStartTagName
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("a", &det, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("copia", nil, nil), 2),
	})
	// set end positions roughly
	sent.GetTokensWithoutWhitespace()[2].SetStartPos(2)

	matches := r.Match(sent)
	require.Len(t, matches, 1)
	require.Equal(t, []string{"cópia"}, matches[0].GetSuggestedReplacements())
}

func TestPortugueseAccentuationCheckRule_NoMap(t *testing.T) {
	r := NewPortugueseAccentuationCheckRule()
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a casa")))
}
