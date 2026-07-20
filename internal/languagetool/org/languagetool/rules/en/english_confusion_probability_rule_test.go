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
	require.Equal(t, EnglishConfusionRuleID, r.GetID())
	// Java example pair
	inex := r.GetIncorrectExamples()
	require.NotEmpty(t, inex)
	require.Contains(t, inex[0].GetExample(), "breaks")
	// Without LM, Match must not invent hits
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Don't forget to put on the breaks.")))
}

func TestEnglishConfusionExceptions_Count(t *testing.T) {
	// Java EnglishConfusionProbabilityRule.EXCEPTIONS length
	require.Len(t, EnglishConfusionExceptions, 493)
	// Spot-check first/last + curly apostrophe entries
	require.Equal(t, "he messages", EnglishConfusionExceptions[0])
	require.Equal(t, "where, when, and who", EnglishConfusionExceptions[len(EnglishConfusionExceptions)-1])
	require.Contains(t, EnglishConfusionExceptions, "he’s")
	require.Contains(t, EnglishConfusionExceptions, "they’re")
	require.Contains(t, EnglishConfusionExceptions, "it’s now better")
	// Wired into constructor
	r := NewEnglishConfusionProbabilityRule(nil)
	require.Len(t, r.Exceptions, 493)
	require.True(t, r.IsLocalException("please let us know in the comments"))
}

func TestEnglishConfusionAntiPatterns_Count(t *testing.T) {
	// Java EnglishConfusionProbabilityRule.ANTI_PATTERNS length
	require.Len(t, EnglishConfusionAntiPatterns, 32)
}

func TestEnglishConfusionIsException_Contraction(t *testing.T) {
	// "…n't know…" — covered from startPos-3 looks like "'t know"
	s := "I don't know it"
	start := strings.Index(s, "know")
	require.Equal(t, "know", s[start:start+4])
	// covered = s[start-3:end] = "'t know"
	require.True(t, enConfusionIsException(s, start, start+4))
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
