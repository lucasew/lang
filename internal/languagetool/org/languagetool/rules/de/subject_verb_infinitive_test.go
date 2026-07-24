package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSubjectVerbAntiPatternsCount(t *testing.T) {
	require.GreaterOrEqual(t, len(SubjectVerbAntiPatterns), 50)
}

func TestSubjectVerbAgreementRule_ContainsOnlyInfinitives(t *testing.T) {
	// Direct unit: containsOnlyInfinitivesToTheLeft requires ≥2 SUB that lookup as VER:INF.
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Kopieren", "SUB:NOM:SIN:NEU", "Kopieren"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("Einfügen", "SUB:NOM:SIN:NEU", "Einfügen"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
	}
	r := NewSubjectVerbAgreementRule(nil).WithLookupInfinitive(func(w string) bool {
		return w == "kopieren" || w == "einfügen"
	})
	require.True(t, r.containsOnlyInfinitivesToTheLeft(toks, 3),
		"Kopieren + Einfügen should count as two infinitives")

	rNone := NewSubjectVerbAgreementRule(nil)
	require.False(t, rNone.containsOnlyInfinitivesToTheLeft(toks, 3),
		"without LookupInfinitive, must not claim infinitives")

	// Match-level: NPP before ist is suppressed when both SUBs are infinitives
	toks[3].SetChunkTags([]string{chunkNPP})
	toks[3].SetStartPos(20)
	toks[4].SetStartPos(30)
	sent := languagetool.NewAnalyzedSentence(toks)
	require.Empty(t, r.Match(sent), "infinitive coordination should suppress NPP+ist")
}

func TestSubjectVerbAgreementRule_AntiPatternProzent(t *testing.T) {
	// "Prozent der Menschen sind" immunized
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Prozent", "SUB:NOM:SIN:NEU", "Prozent"),
		atrWithPOS("der", "ART:DEF:GEN:PLU:ALG", "der"),
		atrWithPOS("Menschen", "SUB:GEN:PLU:MAS", "Mensch"),
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
	}
	pos := 0
	for _, tk := range toks {
		tk.SetStartPos(pos)
		if n := len(tk.GetToken()); n > 0 {
			pos += n + 1
		}
	}
	// mark Menschen as NPS so plural verb might false-alarm without anti-pattern
	toks[3].SetChunkTags([]string{chunkNPS})
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewSubjectVerbAgreementRule(nil)
	imm := r.getSentenceWithImmunization(sent)
	any := false
	for _, tk := range imm.GetTokensWithoutWhitespace() {
		if tk.IsImmunized() {
			any = true
		}
	}
	require.True(t, any, "Prozent der … sind anti-pattern should immunize")
}
