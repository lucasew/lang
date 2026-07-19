package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGetPosAfterModifier_Sehr(t *testing.T) {
	// ein + sehr + hohes + Haus → after modifier at "hohes"
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("sehr", "ADV", "sehr"),
		atrWithPOS("hohes", "ADJ:NOM:SIN:NEU:GRU:DEF", "hoch"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	got := getPosAfterModifier(1, toks)
	require.Equal(t, 2, got, "should skip 'sehr'")
}

func TestGetPosAfterModifier_Meter(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("500", "ZAL", "500"),
		atrWithPOS("Meter", "SUB:NOM:SIN:MAS", "Meter"),
		atrWithPOS("hoher", "ADJ:NOM:SIN:MAS:GRU:IND", "hoch"),
		atrWithPOS("Turm", "SUB:NOM:SIN:MAS", "Turm"),
	}
	got := getPosAfterModifier(1, toks)
	require.Equal(t, 3, got, "should skip '500 Meter'")
}

func TestAgreementRule_ModifierBetweenDetAdjNoun(t *testing.T) {
	// der + sehr + große + Haus (NEU mismatch) must still fire after skipping modifier
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("sehr", "ADV", "sehr"),
		atrWithPOS("große", "ADJ:NOM:SIN:FEM:GRU:DEF", "groß"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms, "der sehr große Haus should mismatch")
}

func TestAgreementRule_IgnoredPronoun(t *testing.T) {
	// "alles" is PRO but in PRONOUNS_TO_BE_IGNORED
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("alles", "PRO:IND:NOM:SIN:NEU", "all"),
		atrWithPOS("Gute", "SUB:NOM:SIN:NEU", "Gute"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage)
	}
}

func TestAgreementRule_IgnoredNoun(t *testing.T) {
	// "mehrere Prozent" — Prozent is NOUNS_TO_BE_IGNORED
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("mehrere", "PRO:IND:NOM:PLU:ALG", "mehrere"),
		atrWithPOS("Prozent", "SUB:NOM:SIN:NEU", "Prozent"),
	}
	// mehrere is ART-like PRO; if categories empty or mismatch, still skip via noun list
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage)
	}
}

func TestAgreementRule_RelativeClauseSkip(t *testing.T) {
	// ", das" with lemma der → relative clause, skip
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Wahlrecht", "SUB:NOM:SIN:NEU", "Wahlrecht"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("das", "PRO:REL:NOM:SIN:NEU", "der"),
		atrWithPOS("Frauen", "SUB:DAT:PLU:FEM", "Frau"),
		atrWithPOS("zugesprochen", "PA2:PRD:GRU:VER", "zusprechen"),
		atrWithPOS("bekamen", "VER:3:PLU:PRT:SFT", "bekommen"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewAgreementRule(nil).Match(sent)
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage, "relative clause must not fire")
	}
}

func TestIsDetNounException(t *testing.T) {
	require.True(t, isDetNounException(
		atrWithPOS("allen", "PRO:IND:DAT:PLU:ALG", "all"),
		atrWithPOS("Grund", "SUB:AKK:SIN:MAS", "Grund"),
	))
}

func TestGetPosAfterModifier_NoModifier(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	require.Equal(t, 1, getPosAfterModifier(1, toks))
}

// Java: single :STV reading → empty set1 → "Meiner Chef" mismatch
func TestAgreementRule_STV_MeinerChef(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("Meiner", "PRO:POS:NOM:SIN:MAS:STV", "mein"),
		atrWithPOS("Chef", "SUB:NOM:SIN:MAS", "Chef"),
	}
	// Note: STV forces mismatch even if categories might otherwise agree; Java empties set1.
	ms := NewAgreementRule(nil).Match(languagetool.NewAnalyzedSentence(toks))
	require.NotEmpty(t, ms, "Meiner Chef with :STV must fire")
}

func TestIsSingleSTVReading(t *testing.T) {
	require.True(t, isSingleSTVReading(atrWithPOS("Meiner", "PRO:POS:NOM:SIN:MAS:STV", "mein")))
	require.False(t, isSingleSTVReading(atrWithPOS("Mein", "PRO:POS:NOM:SIN:MAS:BEG", "mein")))
}

// Java HERR_FRAU: "das ignorierte Herr Grey" — skip when next is EIG
func TestAgreementRule_HerrGrey_Skip(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ignorierte", "ADJ:NOM:SIN:NEU:GRU:DEF", "ignorieren"),
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"),
		atrWithPOS("Grey", "EIG:NOM:SIN:MAS", "Grey"),
	}
	ms := NewAgreementRule(nil).Match(languagetool.NewAnalyzedSentence(toks))
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage, "Herr Grey must not fire agreement")
	}
}
