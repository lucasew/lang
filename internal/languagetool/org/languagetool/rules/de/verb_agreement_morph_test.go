package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbAgreementRule_MorphIchWrongVerb(t *testing.T) {
	// ich + VER:3:SIN (ist) without near 1:SIN
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 12),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(8)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
}

func TestVerbAgreementRule_MorphIchOK(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 11),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewVerbAgreementRule(nil).Match(sent)
	require.Empty(t, ms)
}

func TestVerbAgreementRule_MorphDuWrong(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("hier", "ADV", "hier"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 10),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(3)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
}

func TestVerbAgreementRule_HasUnambiguouslyPersonAndNumber(t *testing.T) {
	// only 1:SIN reading
	v := atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein")
	require.True(t, hasUnambiguouslyPersonAndNumber(v, "1", "SIN"))
	// ambiguous: also 3:SIN reading
	pos2 := "VER:3:SIN:PRÄ:NON"
	lem := "sein"
	v.AddReading(languagetool.NewAnalyzedToken("bin", &pos2, &lem), "")
	require.False(t, hasUnambiguouslyPersonAndNumber(v, "1", "SIN"))
}

func TestIchLooksLikeSubject_JavaGates(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	// "lyrisches Ich" mid-sentence: capital Ich without colon, startPos > 1 → not subject
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("lyrisches", "ADJ:NOM:SIN:NEU", "lyrisch"),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("geht", "VER:3:SIN:PRÄ:SFT", "gehen"),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(10)
	toks[3].SetStartPos(14)
	require.False(t, ichLooksLikeSubject(toks, 2))
	// lowercase ich always
	toks[2] = atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich")
	toks[2].SetStartPos(10)
	require.True(t, ichLooksLikeSubject(toks, 2))
	// Ich after colon
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Hinweis", "SUB:NOM:SIN:MAS", "Hinweis"),
		atrWithPOS(":", "PKT", ":"),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
	}
	toks2[3].SetStartPos(10)
	require.True(t, ichLooksLikeSubject(toks2, 3))
}

// Mid-sentence "Ich" without colon must not fire subject/verb agreement invent
func TestVerbAgreementRule_LyrischesIch_NoMatch(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("lyrische", "ADJ:NOM:SIN:NEU", "lyrisch"),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("geht", "VER:3:SIN:PRÄ:SFT", "gehen"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 20),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(13)
	toks[4].SetStartPos(17)
	// No unambiguous VER:1:SIN so wrong-verb-subject would need Ich gate; mid-sentence Ich fails gate
	ms := NewVerbAgreementRule(nil).Match(languagetool.NewAnalyzedSentence(toks))
	for _, m := range ms {
		// must not flag the mid-sentence Ich as wrong subject for 1:SIN
		require.NotContains(t, m.Message, "1:SIN")
	}
}
