package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWireAgreementRule_ReturnsRule(t *testing.T) {
	r := WireAgreementRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "DE_AGREEMENT", r.GetID())
}

func TestWireSubjectVerbAgreementRule_ReturnsRule(t *testing.T) {
	r := WireSubjectVerbAgreementRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "DE_SUBJECT_VERB_AGREEMENT", r.GetID())
}

func TestWireVerbAgreementRule_ReturnsRule(t *testing.T) {
	r := WireVerbAgreementRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "DE_VERBAGREEMENT", r.GetID())
}

func TestCompoundPhraseValid_FiltersSuggestions(t *testing.T) {
	// With validator that only accepts closed form, hyphen is dropped.
	r := NewAgreementRule(nil)
	r.CompoundPhraseValid = func(phrase string) bool {
		return !containsHyphen(phrase)
	}
	det := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die")
	noun := atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original")
	next := atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail")
	det.SetStartPos(0)
	noun.SetStartPos(4)
	next.SetStartPos(13)
	orig := []*languagetool.AnalyzedTokenReadings{det, noun, next}
	sent := languagetool.NewAnalyzedSentence(orig)
	rm := getCompoundErrorDetNoun(det, noun, 0, orig, sent, r)
	require.NotNil(t, rm)
	require.Len(t, rm.GetSuggestedReplacements(), 1)
	require.Equal(t, "die Originalmail", rm.GetSuggestedReplacements()[0])
}

func containsHyphen(s string) bool {
	for _, r := range s {
		if r == '-' {
			return true
		}
	}
	return false
}

func TestLookupIsInfinitive_NilTagger(t *testing.T) {
	require.False(t, lookupIsInfinitive(nil, "kopieren"))
}

func TestCompoundLastWordIsNoun_NilTagger(t *testing.T) {
	require.False(t, compoundLastWordIsNoun(nil, "die Originalmail"))
}

func TestCompoundPhraseValid_NilTagger(t *testing.T) {
	require.False(t, compoundPhraseValid(nil, "die Originalmail"))
}

func TestCompoundPhraseAgreementOK_EmptyParts(t *testing.T) {
	require.False(t, compoundPhraseAgreementOK(nil, nil))
	require.False(t, compoundPhraseAgreementOK(nil, []string{"die", "Haus"}))
}
