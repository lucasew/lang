package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func sentStartATR() *languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &tag, nil), 0)
}

// isMorphTestPunct is true for tokens that Java typically attaches without a
// preceding space (comma, period, …). Used only by withPositions for morph tests.
func isMorphTestPunct(token string) bool {
	if token == "" {
		return false
	}
	// Common DE punctuation / marks seen in rule morph fixtures.
	switch token {
	case ",", ".", ";", ":", "!", "?", "…", ")", "]", "}", "»", "«", "“", "”", "„", "'", "\"", "’", "‘":
		return true
	}
	return false
}

func withPositions(toks ...*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	// Java AnalyzedTokenReadings positions use String.length() (UTF-16).
	// Spacing: no space before punctuation ("Auto,"), space after punct before
	// words (", das"), spaces between words ("Auto das").
	pos := 0
	for i, t := range toks {
		if t == nil {
			continue
		}
		t.SetStartPos(pos)
		n := utf16LenDE(t.GetToken())
		if n == 0 {
			continue
		}
		pos += n
		// Decide whitespace after this token before the next non-empty token.
		var next *languagetool.AnalyzedTokenReadings
		for j := i + 1; j < len(toks); j++ {
			if toks[j] != nil && utf16LenDE(toks[j].GetToken()) > 0 {
				next = toks[j]
				break
			}
		}
		if next == nil {
			continue
		}
		thisPunct := isMorphTestPunct(t.GetToken())
		nextPunct := isMorphTestPunct(next.GetToken())
		if !thisPunct && nextPunct {
			// "Auto," — no space before punctuation
			continue
		}
		// word→word, punct→word (", das"): one space
		pos++
	}
	return toks
}

func TestAgreementRule_AntiPatternImmunizesJedesGrad(t *testing.T) {
	// Java: token("jedes"), token("Grad")
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("jedes", "PRO:IND:NOM:SIN:NEU", "jeder"),
		atrWithPOS("Grad", "SUB:NOM:SIN:NEU", "Grad"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	imm := r.getSentenceWithImmunization(sent)
	anyImm := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			anyImm = true
			break
		}
	}
	require.True(t, anyImm, "expected anti-pattern jedes Grad to immunize")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		require.False(t, tok.IsImmunized(), "original must stay clean")
	}
}

func TestAgreementRule_ImmunizedEinBisschenPLU(t *testing.T) {
	// Java: ein + bisschen + SUB:.*PLU.*
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("bisschen", "ADV", "bisschen"),
		atrWithPOS("Einsichten", "SUB:AKK:PLU:FEM", "Einsicht"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	imm := r.getSentenceWithImmunization(sent)
	anyImm := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			anyImm = true
		}
	}
	require.True(t, anyImm)
	// immunized span must not produce DET-NOUN mismatch on ein+Einsichten path
	ms := r.Match(sent)
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage)
	}
}

func TestAgreementRule_SuggestorAttachedWhenSynthSet(t *testing.T) {
	// Alternate det forms so edits > 0 (Java drops identical phrases).
	// Tags look like ART:DEF:NOM:SIN:NEU (no trailing colon after gender).
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil {
				return nil, nil
			}
			if strings.Contains(posTag, "ART:") && strings.Contains(posTag, "NEU") {
				return []string{"das"}, nil
			}
			if strings.Contains(posTag, "ART:") {
				return []string{"die"}, nil
			}
			if strings.Contains(posTag, "SUB:") {
				return []string{"Haus"}, nil
			}
			return []string{token.GetToken()}, nil
		},
	}
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil).WithSynth(synth)
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
	require.NotEmpty(t, ms[0].GetSuggestedReplacements(), "synth should yield suggestions")
}

func TestAgreementRule_NoSuggestionsWithoutSynth(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
	require.Empty(t, ms[0].GetSuggestedReplacements())
}

func TestAllAgreementAntiPatternsNonEmpty(t *testing.T) {
	aps := AllAgreementAntiPatterns()
	require.Equal(t, 165+135+127, len(aps), "Java AgreementRuleAntiPatterns1+2+3")
	require.Equal(t, len(AgreementRuleAntiPatterns1)+len(AgreementRuleAntiPatterns2)+len(AgreementRuleAntiPatterns3), len(aps))
}

func TestAgreementRule_AntiPatternVieleWenigerBekannte(t *testing.T) {
	// Java AP1: posRegexWithStringException(PRO:(IND|POS).*, eine[nm]) + PA/ADV + ADJ + SUB
	// "Viele weniger bekannte Vorschläge"
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Viele", "PRO:IND:NOM:PLU:FEM", "viel"),
		atrWithPOS("weniger", "ADV:TMP", "weniger"),
		atrWithPOS("bekannte", "ADJ:NOM:PLU:FEM:GRU:IND", "bekannt"),
		atrWithPOS("Vorschläge", "SUB:NOM:PLU:MAS", "Vorschlag"),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	imm := r.getSentenceWithImmunization(sent)
	anyImm := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			anyImm = true
			break
		}
	}
	require.True(t, anyImm, "expected Viele weniger bekannte Vorschläge to immunize")
}

func TestAgreementRule_AntiPatternCADPdf(t *testing.T) {
	// Java AP2: CAD + "." (no ws before) + pdf (no ws before)
	dot := atrWithPOS(".", "PKT", ".")
	dot.SetWhitespaceBefore(false)
	pdf := atrWithPOS("pdf", "UNKNOWN", "pdf")
	pdf.SetWhitespaceBefore(false)
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("CAD", "UNKNOWN", "CAD"),
		dot,
		pdf,
	)
	// withPositions adds +1 space between tokens in start pos only; whitespace-before flags stay as set
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	imm := r.getSentenceWithImmunization(sent)
	anyImm := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			anyImm = true
			break
		}
	}
	require.True(t, anyImm, "expected CAD.pdf anti-pattern to immunize")
}

func TestNounCasesForSuggestor_PrepositionMit(t *testing.T) {
	prep := atrWithPOS("mit", "PRP:DAT:SIN", "mit")
	s := NewAgreementSuggestor2(nil, nil, atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus")).WithPreposition(prep)
	cases := s.nounCasesForSuggestor()
	require.Equal(t, []string{"DAT"}, cases)
}

func TestNounCasesForSuggestor_ReplInsAllCases(t *testing.T) {
	// Java: Ins sets prep "in" but PrepositionToCases omits "in" → all cases
	s := NewAgreementSuggestor2(nil, nil, atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus")).WithReplacementType(ReplIns)
	require.Equal(t, []string{"NOM", "AKK", "DAT", "GEN"}, s.nounCasesForSuggestor())
}
