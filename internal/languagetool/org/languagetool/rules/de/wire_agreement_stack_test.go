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

func TestCompoundPhraseValid_NilTagger(t *testing.T) {
	require.False(t, compoundPhraseValid(nil, "die Originalmail"))
	require.False(t, compoundPhraseValid(nil, ""))
}

// Twin of Java getRuleMatch lt.check gate for open-compound phrases
// (DE_AGREEMENT + GERMAN_SPELLER only). Requires discovered tagger + hunspell.
func TestCompoundPhraseValid_JavaTable(t *testing.T) {
	r := WireAgreementRule(nil)
	require.NotNil(t, r)
	if r.CompoundPhraseValid == nil {
		t.Skip("CompoundPhraseValid unavailable (no GermanTagger resources)")
	}
	v := r.CompoundPhraseValid
	// Accepted by Java lt.check (productive compounds / hyphen forms)
	require.True(t, v("die Originalmail"), "die Originalmail")
	require.True(t, v("die Original-Mail"), "die Original-Mail")
	require.True(t, v("dieser kleine Magnesiumanteil"), "Magnesiumanteil closed")
	require.True(t, v("dieser kleine Magnesium-Anteil"), "Magnesium-Anteil hyphen")
	require.True(t, v("Ein neuer LanguageTool-Account"), "LanguageTool-Account")
	require.True(t, v("Die Standardpriorität"), "Standardpriorität")
	require.True(t, v("Die Standard-Priorität"), "Standard-Priorität")
	// Rejected by Java GERMAN_SPELLER on closed non-word form
	require.False(t, v("Ein neuer LanguageToolaccount"), "LanguageToolaccount closed invent")
}

func TestCompoundPhraseAgreementOK_EmptyParts(t *testing.T) {
	require.False(t, compoundPhraseAgreementOK(nil, nil))
	require.False(t, compoundPhraseAgreementOK(nil, []string{"die", "Haus"}))
}
