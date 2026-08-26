package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Fail-closed without resources/hooks: no invent det synthesis.
func TestAdaptSuggestionFilter_FailClosedWithoutHooks(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	// Force empty hooks (do not call Wire)
	f.GenderOf = func(string) string { return "" }
	f.Synthesize = func(string, string) []string { return nil }
	require.Empty(t, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
}

// Possessive first-char filter: do not suggest "dein" for surface "meine" (shared lemma family).
func TestAdaptSuggestionFilter_PossessiveFirstCharFilter(t *testing.T) {
	f := WireAdaptSuggestionFilter()
	if f.Synthesize == nil || f.GenderOf == nil {
		t.Skip("DE synth/tagger unavailable")
	}
	require.Equal(t, []string{"mein"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:NOM:SIN:FEM:BEG", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"dein"}, f.AdaptedDet(DetReading{Token: "deine", POS: "PRO:POS:NOM:SIN:FEM:BEG", Lemma: "dein"}, "Plan"))
}

// Title-case det surfaces upper-case the adapted form (Java Die → Der).
func TestAdaptSuggestionFilter_UppercaseDet(t *testing.T) {
	f := WireAdaptSuggestionFilter()
	if f.Synthesize == nil || f.GenderOf == nil {
		t.Skip("DE synth/tagger unavailable")
	}
	require.Equal(t, []string{"Der"}, f.AdaptedDet(DetReading{Token: "Die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Plan"))
}

// AcceptRuleMatch expands span to previous det and returns original match when no adapt.
func TestAdaptSuggestionFilter_AcceptRuleMatchSpanAndPassthrough(t *testing.T) {
	f, lt := adaptSuggestionFilterTestLT(t)
	sentenceStr := "Hier steht die Roadmap."
	sentence := lt.GetAnalyzedSentence(sentenceStr)
	require.NotNil(t, sentence)
	fromPos := stringsIndex(sentenceStr, "Roadmap")
	toPos := fromPos + len("Roadmap")
	match := rules.NewRuleMatch(rules.NewFakeRule(""), sentence, fromPos, toPos, "fake message")
	match.SetSuggestedReplacement("Plan")
	tokens := sentence.GetTokensWithoutWhitespace()
	tokenPos := -1
	for i, tok := range tokens {
		if tok != nil && tok.GetToken() == "Roadmap" {
			tokenPos = i
			break
		}
	}
	require.Greater(t, tokenPos, 0)
	prev := tokens[tokenPos-1]
	newMatch := f.AcceptRuleMatch(match, map[string]string{"sub": `\1`}, tokenPos, nil, nil)
	require.NotNil(t, newMatch)
	require.Equal(t, prev.GetStartPos(), newMatch.GetFromPos())
	require.Equal(t, toPos, newMatch.GetToPos())
	require.NotEmpty(t, newMatch.GetSuggestedReplacements())

	// patternTokenPos == 0 → original match
	pass := f.AcceptRuleMatch(match, nil, 0, nil, nil)
	require.Equal(t, match, pass)
}

func stringsIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// getAdaptedDetAdj is present (Java method) even though accept path is disabled.
func TestAdaptSuggestionFilter_GetAdaptedDetAdjExists(t *testing.T) {
	f := WireAdaptSuggestionFilter()
	if f.Synthesize == nil || f.GenderOf == nil {
		t.Skip("DE synth/tagger unavailable")
	}
	det := atrWithPOS("eine", "ART:IND:NOM:SIN:FEM", "ein")
	adj := atrWithPOS("neue", "ADJ:NOM:SIN:FEM:GRU:IND", "neu")
	// May be empty or non-empty depending on synth tags; must not panic.
	_ = f.GetAdaptedDetAdj(det, adj, "Plan")
}

// nounGender first-reading SUB: gender (Java getNounGender).
func TestAdaptSuggestionFilter_NounGender(t *testing.T) {
	tagger := DiscoveredGermanTagger()
	if tagger == nil {
		t.Skip("no tagger")
	}
	require.Equal(t, "MAS", nounGenderFromTagger(tagger, "Plan"))
	require.Equal(t, "FEM", nounGenderFromTagger(tagger, "Idee"))
	require.Equal(t, "NEU", nounGenderFromTagger(tagger, "Verfahren"))
	require.Equal(t, "MAS", nounGenderFromTagger(tagger, "Mann"))
	require.Equal(t, "FEM", nounGenderFromTagger(tagger, "Frau"))
	require.Equal(t, "NEU", nounGenderFromTagger(tagger, "Kind"))
}

