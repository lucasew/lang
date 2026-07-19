package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSubjectVerbAgreementRule_MorphPluralSubjectSingularVerb(t *testing.T) {
	// Die Autos ist — NPP-like SUB:PLU + ist
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die"),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
	}
	// chunk NPP on subject noun (as German chunker would)
	toks[2].SetChunkTags([]string{chunkNPP})
	toks[1].SetChunkTags([]string{chunkNPP})
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(10)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "sind")
	// Java message embeds <suggestion>sind</suggestion>; no invent shortMessage.
	require.Contains(t, ms[0].GetMessage(), "<suggestion>sind</suggestion>")
	require.Empty(t, ms[0].GetShortMessage())
}

func TestSubjectVerbAgreementRule_Meta(t *testing.T) {
	r := NewSubjectVerbAgreementRule(nil)
	require.Equal(t, "Kongruenz von Subjekt und Prädikat (unvollständig)", r.GetDescription())
	require.Greater(t, r.EstimateContextForSureMatch(), 0)
	require.Equal(t,
		"https://dict.leo.org/grammatik/deutsch/Wort/Verb/Kategorien/Numerus-Person/ProblemNum.html",
		r.GetURL())
}

func TestSubjectVerbAgreementRule_MorphSingularSubjectPluralVerb(t *testing.T) {
	// Das Auto sind — NPS + sind
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
	}
	toks[1].SetChunkTags([]string{chunkNPS})
	toks[2].SetChunkTags([]string{chunkNPS})
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(9)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "ist")
}

func TestSubjectVerbAgreementRule_MorphOK(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
	}
	toks[1].SetChunkTags([]string{chunkNPS})
	toks[2].SetChunkTags([]string{chunkNPS})
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.Empty(t, ms)
}

func TestSubjectVerbAgreementRule_NoChunksNoInvent(t *testing.T) {
	// Java: only chunkTags.contains(NPP/NPS) — POS-only SUB:PLU must not invent a hit.
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(6)
	// no SetChunkTags
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.Empty(t, ms, "without NPP/NPS chunks Java does not match")
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominative_NoChunkFalse(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
	}
	// NOM alone without NPS/NPP must not invent nominative chunk span
	require.False(t, prevChunkIsNominative(toks, 1))
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominativeMorph(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
	}
	toks[1].SetChunkTags([]string{chunkNPP})
	require.True(t, prevChunkIsNominative(toks, 1))
}
