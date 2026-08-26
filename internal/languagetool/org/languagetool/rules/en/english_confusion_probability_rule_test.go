package en

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestEnglishConfusionProbabilityRule(t *testing.T) {
	r := NewEnglishConfusionProbabilityRule(nil)
	// Java: English does not override getId → ConfusionProbabilityRule.RULE_ID
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
	require.Equal(t, "CONFUSION_RULE", r.GetID())
	// Java example pair: breaks/brakes
	inex := r.GetIncorrectExamples()
	require.NotEmpty(t, inex)
	require.Contains(t, inex[0].GetExample(), "breaks")
	cex := r.GetCorrectExamples()
	require.NotEmpty(t, cex)
	require.Contains(t, cex[0].GetExample(), "brakes")
	// Without LM, Match must not invent hits even with official pairs loaded
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Don't forget to put on the breaks.")))
	// getFilenames → confusion_sets.txt
	require.Equal(t, []string{"confusion_sets.txt"}, r.GetFilenames())
}

func TestEnglishConfusionExceptions_Count(t *testing.T) {
	// Java EnglishConfusionProbabilityRule.EXCEPTIONS length (comment-stripped string literals)
	require.Len(t, EnglishConfusionExceptions, 483)
	// Spot-check first/last + curly apostrophe entries
	require.Equal(t, "he messages", EnglishConfusionExceptions[0])
	require.Equal(t, "where, when, and who", EnglishConfusionExceptions[len(EnglishConfusionExceptions)-1])
	require.Contains(t, EnglishConfusionExceptions, "he’s")
	require.Contains(t, EnglishConfusionExceptions, "they’re")
	require.Contains(t, EnglishConfusionExceptions, "it’s now better")
	// Invent from Java // comments must not appear as exceptions
	require.NotContains(t, EnglishConfusionExceptions, "know")
	require.NotContains(t, EnglishConfusionExceptions, "want do")
	require.NotContains(t, EnglishConfusionExceptions, "... is host to five classical music orchestras")
	// Wired into constructor
	r := NewEnglishConfusionProbabilityRule(nil)
	require.Len(t, r.Exceptions, 483)
	require.True(t, r.IsLocalException("please let us know in the comments"))
}

func TestEnglishConfusionAntiPatterns_Count(t *testing.T) {
	// Java EnglishConfusionProbabilityRule.ANTI_PATTERNS length
	require.Len(t, EnglishConfusionAntiPatterns, 32)
	// Wired into match immunization path
	require.NotNil(t, NewEnglishConfusionProbabilityRule(nil).IsCoveredByAntiPattern)
}

func TestEnglishConfusionIsException_Contraction(t *testing.T) {
	// Java isException uses UTF-16 indices (GoogleToken.startPos / String.substring).
	// "I don't know it" — "know" at UTF-16 8..12; covered = "'t know"
	s := "I don't know it"
	start := strings.Index(s, "know") // ASCII → UTF-8 == UTF-16
	require.Equal(t, "know", s[start:start+4])
	require.True(t, enConfusionIsException(s, start, start+4))
	// curly apostrophe n't: "I don't know" with U+2019 — UTF-16 "know" at 8..12
	// (not UTF-8 byte index; ' is one UTF-16 unit).
	s2 := "I don’t know it"
	// Java: startPos of "know" after "I don’t " = 2 + 5 + 1 = 8
	require.True(t, enConfusionIsException(s2, 8, 12))
	// wrong (UTF-8) index must not be treated as a free pass for invent
	require.False(t, enConfusionIsException(s2, strings.Index(s2, "know"), strings.Index(s2, "know")+4))
	// non-contraction
	require.False(t, enConfusionIsException("I do know it well", 5, 9))
	// startPos <= 3 → false
	require.False(t, enConfusionIsException("ab", 1, 2))
}

func TestEnglishConfusionAntiPattern_WayToo(t *testing.T) {
	// token-only anti-pattern: way + too
	toks := enWithPositions(
		enSentStart(),
		enAtrPOS("This", "DT", "this"),
		enAtrPOS("way", "NN", "way"),
		enAtrPOS("too", "RB", "too"),
		enAtrPOS("hard", "JJ", "hard"),
		enAtrPOS(".", "PCT", "."),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	way := toks[2]
	require.True(t, enConfusionIsCoveredByAntiPattern(sent, way.GetStartPos(), way.GetEndPos()))
}

func TestEnglishConfusion_WithLMAndPair(t *testing.T) {
	// Fake LM: always prefer "brakes" over "breaks"
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		for _, tok := range tokens {
			if tok == "brakes" {
				return ngrams.NewProbabilitySimple(0.9, 1.0)
			}
			if tok == "breaks" {
				return ngrams.NewProbabilitySimple(0.001, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewEnglishConfusionProbabilityRule(lm)
	pair := rules.NewConfusionPairTokens("breaks", "brakes", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("Don't forget to put on the breaks."))
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "brakes")
}

func TestEnglishConfusion_ExceptionPhraseSkips(t *testing.T) {
	// "let us know" is in EXCEPTIONS — should not fire when phrase covers span
	lm := ngrams.FuncLanguageModel(func(tokens []string) ngrams.Probability {
		for _, tok := range tokens {
			if tok == "now" {
				return ngrams.NewProbabilitySimple(0.9, 1.0)
			}
			if tok == "know" {
				return ngrams.NewProbabilitySimple(0.001, 1.0)
			}
		}
		return ngrams.NewProbabilitySimple(0.1, 1.0)
	})
	r := NewEnglishConfusionProbabilityRule(lm)
	pair := rules.NewConfusionPairTokens("know", "now", 10, true)
	r.SetConfusionPair(pair)
	matches := r.Match(languagetool.AnalyzePlain("Please let us know in the comments."))
	require.Empty(t, matches)
}

func TestEnglishConfusion_LoadsOfficialConfusionSets(t *testing.T) {
	// Java CPR loads /en/confusion_sets.txt in constructor via getFilenames().
	r := NewEnglishConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotEmpty(t, r.WordToPairs,
		"WordToPairs must load from official en/confusion_sets.txt (fail-closed empty only if resource missing)")
	// Spot-check a well-known pair from official file (their/there is classic)
	// Keys exist for at least one side of a known pair.
	hasTheir := len(r.WordToPairs["their"]) > 0 || len(r.WordToPairs["there"]) > 0
	hasBreaks := len(r.WordToPairs["breaks"]) > 0 || len(r.WordToPairs["brakes"]) > 0
	require.True(t, hasTheir || hasBreaks, "expected at least one official confusion pair key")
}

func enSentStart() *languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &tag, nil), 0)
}

func enAtrPOS(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	if lemma != "" {
		ll := lemma
		l = &ll
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, p, l), 0)
}

func enWithPositions(toks ...*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	pos := 0
	for _, t := range toks {
		if t == nil {
			continue
		}
		t.SetStartPos(pos)
		if n := len(t.GetToken()); n > 0 {
			pos += n + 1
		}
	}
	return toks
}
