package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UppercaseSentenceStartRuleTest.java
// Multi-sentence analysis uses GermanSRXSentenceTokenizer (Java JLT sentence split), not naive SplitAndAnalyze.
// Soft hyphens: Java lt.check(String) goes through AnnotatedTextBuilder.addText which treats U+00AD as
// markup, so SRX sees plain text without soft hyphens; positions map back to original (Java AnnotatedText).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

// matchUppercaseDE ports Java lt.check path for UPPERCASE_SENTENCE_START on a raw string:
// AnnotatedTextBuilder.addText (soft hyphen → markup) → SRX on plain → MatchList → original positions.
func matchUppercaseDE(r *UppercaseSentenceStartRule, text string) []*rules.RuleMatch {
	at := markup.NewAnnotatedTextBuilder().AddText(text).Build()
	plain := at.GetPlainText()
	parts := tokde.NewGermanSRXSentenceTokenizer().Tokenize(plain)
	sents := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		sents = append(sents, languagetool.AnalyzePlain(p))
	}
	ms := r.MatchList(sents)
	// Map plain-text match positions to original (with soft hyphens / markup) like Java.
	out := make([]*rules.RuleMatch, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			continue
		}
		from := at.GetOriginalTextPositionFor(m.GetFromPos(), false)
		to := at.GetOriginalTextPositionFor(m.GetToPos(), true)
		nm := rules.NewRuleMatch(m.GetRule(), m.Sentence, from, to, m.GetMessage())
		if reps := m.GetSuggestedReplacements(); len(reps) > 0 {
			nm.SetSuggestedReplacements(reps)
		}
		out = append(out, nm)
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
	require.Empty(t, r.MatchList(analyzeDESentence("'Dies ist ein Satz!'")))
	require.Empty(t, r.MatchList(analyzeDESentence("mRNA-Impfstoffe sind wichtig.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Willkommen… im Berlin.")))
	require.Empty(t, r.MatchList(analyzeDESentence("schön")))
	require.Empty(t, r.MatchList(analyzeDESentence("Satz")))
	// abbreviations: no false sentence break → no false uppercase hit
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist bspw. okay so.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist z.B. okay so.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dieser Satz ist z. B. okay so.")))
	require.Empty(t, r.MatchList(analyzeDESentence("Dies ist ein Satz. \"Aber der hier auch!\".")))
	require.Empty(t, r.MatchList(analyzeDESentence("Sehr geehrte Frau Merkel,\nwie wir Ihnen schon früher mitgeteilt haben...")))
	require.Empty(t, r.MatchList(analyzeDESentence("Die neue Kollektion von NAU! ist jetzt online.")))

	// Java bad
	require.Equal(t, 2, len(r.MatchList(analyzeDESentence("etwas beginnen. und der auch nicht"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("schön!"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. ätsch, noch einer mit Umlaut."))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. \"aber der hier auch!\""))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("Dies ist ein Satz. „aber der hier auch!“"))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("\"dies ist ein Satz!\""))))
	require.Equal(t, 1, len(r.MatchList(analyzeDESentence("'dies ist ein Satz!'"))))

	// positions: "Ein Test. was?" — Java fromPos=10 toPos=13
	sents := analyzeDESentence("Ein Test. was?")
	require.GreaterOrEqual(t, len(sents), 2)
	ms := r.MatchList(sents)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 10, ms[0].GetFromPos())
	require.Equal(t, 13, ms[0].GetToPos())

	// Java soft hyphen removal / position fixing via AnnotatedTextBuilder.addText
	ms0 := matchUppercaseDE(r, "Ein Test. was?")
	require.Equal(t, 1, len(ms0))
	require.Equal(t, 10, ms0[0].GetFromPos())
	require.Equal(t, 13, ms0[0].GetToPos())

	ms1 := matchUppercaseDE(r, "Ein \u00ADTest. was?")
	require.Equal(t, 1, len(ms1))
	require.Equal(t, 11, ms1[0].GetFromPos())
	require.Equal(t, 14, ms1[0].GetToPos())

	ms2 := matchUppercaseDE(r, "Ein \u00ADTe\u00ADst. was?")
	require.Equal(t, 1, len(ms2))
	require.Equal(t, 12, ms2[0].GetFromPos())
	require.Equal(t, 15, ms2[0].GetToPos())

	ms3 := matchUppercaseDE(r, "Ein \u00ADTe\u00ADst. Te\u00ADst. was?")
	require.Equal(t, 1, len(ms3))
	require.Equal(t, 19, ms3[0].GetFromPos())
	require.Equal(t, 22, ms3[0].GetToPos())
}
