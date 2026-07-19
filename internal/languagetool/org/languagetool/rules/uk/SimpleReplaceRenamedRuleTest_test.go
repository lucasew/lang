package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRenamedRuleTest.java
// Lemma + GEO POS only (Java); AnalyzePlain injects lemma/POS (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRenamedRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceRenamedRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Київ."))))

	// Дніпродзержинська → noun + adj lemmas with prop/adj tags (Java Morphy)
	matches := rule.Match(withUKLemmas("Дніпродзержинська", []lemmaPOS{
		{"Дніпродзержинськ", "noun:inanim:f:v_naz:prop:geo"},
		{"дніпродзержинський", "adj:f:v_naz"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Кам'янське", "кам'янський"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "2016")

	matches = rule.Match(withUKLemmas("дніпродзержинського.", []lemmaPOS{
		{"дніпродзержинський", "adj:m:v_rod"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"кам'янський"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(withUKLemmas("Червонознам'янка.", []lemmaPOS{
		{"Червонознам'янка", "noun:inanim:f:v_naz:prop:geo"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Знам'янка", "Знаменка"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].URL, "wikipedia")

	matches = rule.Match(withUKLemmas("Переяслав-Хмельницький.", []lemmaPOS{
		{"Переяслав-Хмельницький", "noun:inanim:m:v_naz:prop:geo"},
		{"переяслав-хмельницький", "adj:m:v_naz"},
	}))
	// Only lemmas present in file fire; adj key may differ — check file keys
	// File has Переяслав-Хмельницький=Переяслав only (one entry). Second adj may clear set if not in map.
	// Java: if adj lemma not in list → clear all. So inject only noun lemma for twin.
	matches = rule.Match(withUKLemmas("Переяслав-Хмельницький.", []lemmaPOS{
		{"Переяслав-Хмельницький", "noun:inanim:m:v_naz:prop:geo"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Переяслав"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceRenamedRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewSimpleReplaceRenamedRule(nil)
	// Surface-only (no lemma/POS) never matches — no invent.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Дніпродзержинська"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Червонознам'янка."))))
}

type lemmaPOS struct {
	lemma, pos string
}

// withUKLemmas builds a one-token sentence and injects lemma/POS readings (Java Morphy stand-in).
func withUKLemmas(text string, readings []lemmaPOS) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzePlain(text)
	nws := sent.GetTokensWithoutWhitespace()
	// first content token
	for _, tok := range nws {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// skip pure punct
		if tok.GetToken() == "." || tok.GetToken() == "!" || tok.GetToken() == "?" {
			continue
		}
		for _, rp := range readings {
			lem := rp.lemma
			pos := rp.pos
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, &lem), "test")
		}
		break
	}
	return sent
}
