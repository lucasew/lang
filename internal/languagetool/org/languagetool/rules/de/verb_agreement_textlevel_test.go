package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbAgreementAntiPatternsCount(t *testing.T) {
	require.GreaterOrEqual(t, len(VerbAgreementAntiPatterns), 90)
}

func TestVerbAgreementRule_MatchList_ConjunctionSplit(t *testing.T) {
	// "Ich sind müde, weil du bist hier." — first clause still errors
	// Build with whitespace tokens like full analyzer: word, ",", " ", "weil", ...
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	space := " "
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("sind", "VER:1:PLU:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		// comma + space + weil (Java split on tokens[i-2]="," and tokens[i]="weil")
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(",", nil, nil), 12),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(space, nil, nil), 13),
		atrWithPOS("weil", "KON:UNT", "weil"),
		atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"),
		atrWithPOS("hier", "ADV", "hier"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 30),
	}
	// mark space as whitespace if needed
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewVerbAgreementRule(nil)
	ms := r.MatchList([]*languagetool.AnalyzedSentence{sent})
	// At least the first-clause "Ich sind" should fire
	require.NotEmpty(t, ms)
}

func TestVerbAgreementRule_BinIgnoreArabicName(t *testing.T) {
	// "... Osama bin Laden" — "bin" after Osama must not fire as 1:SIN without ich
	ss := languagetool.SentenceStartTagName
	se := languagetool.SentenceEndTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Osama", "EIG:NOM:SIN:MAS", "Osama"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Laden", "EIG:NOM:SIN:MAS", "Laden"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &se, nil), 20),
	}
	// Only "bin" as unambiguous 1:SIN — but predecessor in BIN_IGNORE
	// Note: hasUnambiguouslyPersonAndNumber for "bin" with only 1:SIN
	// Also need lowercase start - "bin" is lower, good.
	// bin ignore: BIN_IGNORE.contains(tokens[i-1]) when strToken=="bin"
	sent := languagetool.NewAnalyzedSentence(toks)
	// Ensure bin only has 1:SIN reading so it would be posVer1Sin without ignore
	ms := NewVerbAgreementRule(nil).Match(sent)
	// Should not report wrong verb for bin (name pattern)
	for _, m := range ms {
		require.NotContains(t, m.Message, "bin")
	}
}
