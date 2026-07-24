package en

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Behavior matrix twin for org.languagetool.rules.en.StyleTooOftenUsedAdjectiveRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedWordRules_IDs(t *testing.T) {
	require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", NewStyleTooOftenUsedNounRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_VERB_EN", NewStyleTooOftenUsedVerbRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_EN", NewStyleTooOftenUsedAdjectiveRule().GetID())
}

func TestStyleTooOftenUsedAdjectiveRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedAdjectiveRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedAdjectiveRule()
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_EN", rule.GetID())
	require.Equal(t, "Statistical Style Analysis: Overused Adjective", rule.GetDescription())
	// Java: "Show when a adjective is used more often than ...% of all adjectives:"
	require.Equal(t, "Show when a adjective is used more often than ...% of all adjectives:", rule.GetConfigureText())
	require.Equal(t, 5, rule.MinPercent)                // DEFAULT_MIN_PERCENT
	require.Equal(t, 5, rule.DefaultMinPercent)         // Java defaultMinPercent
	require.Equal(t, 100, rule.MinWordCount)            // Abstract MIN_WORD_COUNT
	require.True(t, rule.IsDefaultOff())                // DEFAULT_ACTIVATION = false
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	cat := rule.GetCategory()
	require.NotNil(t, cat)
	require.Equal(t, "CREATIVE_WRITING", cat.GetID().String())
	require.True(t, cat.IsDefaultOff())

	// getLimitMessage(limit)
	require.Equal(t,
		"The adjective is used more than 5% times of all adjectives. It may be better to replace it with a synonym.",
		rule.LimitMessage(5))
	require.Equal(t,
		"The adjective is used more than 12% times of all adjectives. It may be better to replace it with a synonym.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedAdjectiveRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule()

	// isToCountedWord = hasPosTagStartingWith("JJ")
	require.True(t, rule.IsToCountedWord(styleATR("good", "JJ", "good")))
	require.True(t, rule.IsToCountedWord(styleATR("better", "JJR", "good")))
	require.True(t, rule.IsToCountedWord(styleATR("best", "JJS", "good")))
	require.False(t, rule.IsToCountedWord(styleATR("quickly", "RB", "quickly")))
	require.False(t, rule.IsToCountedWord(styleATR("cat", "NN", "cat")))

	// isException = RB|IN|CD|DT|NN starting tags (token at index n)
	tokens := []*languagetool.AnalyzedTokenReadings{
		styleSentStart(),
		styleATR("quickly", "RB", "quickly"),
		styleATR("in", "IN", "in"),
		styleATR("two", "CD", "two"),
		styleATR("the", "DT", "the"),
		styleATR("cat", "NN", "cat"),
		styleATR("good", "JJ", "good"),
	}
	require.True(t, rule.IsException(tokens, 1))  // RB
	require.True(t, rule.IsException(tokens, 2))  // IN
	require.True(t, rule.IsException(tokens, 3))  // CD
	require.True(t, rule.IsException(tokens, 4))  // DT
	require.True(t, rule.IsException(tokens, 5))  // NN
	require.False(t, rule.IsException(tokens, 6)) // pure JJ

	// Multi-tag token that is JJ and also NN → exception (Java ORs all tags)
	jjAndNN := styleATRMulti("flat", "flat", "JJ", "NN")
	require.True(t, rule.IsToCountedWord(jjAndNN))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{styleSentStart(), jjAndNN}, 1))
}

func TestStyleTooOftenUsedAdjectiveRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule()
	// Java: getLemmaForPosTagStartsWith("JJ", token)
	require.Equal(t, "good", rule.ToAddedLemma(styleATR("better", "JJR", "good")))
	require.Equal(t, "good", rule.ToAddedLemma(styleATR("best", "JJS", "good")))
	// No JJ reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(styleATR("quickly", "RB", "quickly")))
	// JJ reading with nil lemma → empty string path (Java returns null lemma)
	pos := "JJ"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedAdjectiveRule()
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("car", "NN", "car"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("house", "NN", "house"),
		styleATR(".", ".", "."),
	))
	// Short text: only 2 counted JJ lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedAdjectiveRule()
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("car", "NN", "car"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("house", "NN", "house"),
		styleATR(".", ".", "."),
	))
	// "good" is 100% of adjectives ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "The adjective is used more than 5% times of all adjectives. It may be better to replace it with a synonym."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_EN", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("bad", "JJ", "bad"),
		styleATR("idea", "NN", "idea"),
		styleATR(".", ".", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// Tokens that have JJ but also NN/RB/... are exceptions and must not enter the map.
	rule := NewStyleTooOftenUsedAdjectiveRule()
	rule.MinWordCount = 0

	// Only pure JJ "good"; "flat" has JJ+NN → isException
	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATRMulti("flat", "flat", "JJ", "NN"),
		styleATR("car", "NN", "car"),
		styleATR(".", ".", "."),
	))
	// only "good" counted → 100% ≥ 5 → one match on "good" only
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	// span should cover "good" not "flat"
	require.Equal(t, "good", extractSurface(s, ms[0]))
}

func TestStyleTooOftenUsedAdjectiveRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	rule := NewStyleTooOftenUsedAdjectiveRule()
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	// Sentence: He said "good good good" and nice cars.
	// Inside quotes: three "good" JJ; outside: "nice" JJ.
	// With exclusion: only "nice" counted → 100% of adjectives → match on "nice" only.
	openQ := styleATR("\"", "``", "\"")
	// next after open quote has no whitespace before (Java condition for entering DS)
	good1 := styleATR("good", "JJ", "good")
	good1.SetWhitespaceBefore(false)
	good2 := styleATR("good", "JJ", "good")
	good2.SetWhitespaceBefore(true)
	good3 := styleATR("good", "JJ", "good")
	good3.SetWhitespaceBefore(true)
	endQ := styleATR("\"", "''", "\"")
	endQ.SetWhitespaceBefore(false) // Java ENDING_QUOTES: !tokens[n].isWhitespaceBefore()
	nice := styleATR("nice", "JJ", "nice")
	nice.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("He", "PRP", "he"),
		styleATR("said", "VBD", "say"),
		openQ,
		good1,
		good2,
		good3,
		endQ,
		styleATR("and", "CC", "and"),
		nice,
		styleATR("cars", "NNS", "car"),
		styleATR(".", ".", "."),
	))
	// Force openQ next (good1) whitespaceBefore=false after withPositions may have set it.
	toks := s.GetTokensWithoutWhitespace()
	// find quote and next
	for i := 1; i < len(toks)-1; i++ {
		if toks[i].GetToken() == "\"" && toks[i+1].GetToken() == "good" {
			toks[i+1].SetWhitespaceBefore(false)
			break
		}
	}
	// ending quote whitespaceBefore=false
	for i := 1; i < len(toks); i++ {
		if toks[i].GetToken() == "\"" {
			// last quote that ends DS: second quote token
		}
	}
	// Set ending quote: last "\" that is not the open one
	var quoteIdxs []int
	for i := 1; i < len(toks); i++ {
		if toks[i].GetToken() == "\"" {
			quoteIdxs = append(quoteIdxs, i)
		}
	}
	require.GreaterOrEqual(t, len(quoteIdxs), 2)
	toks[quoteIdxs[0]+1].SetWhitespaceBefore(false) // enter DS
	toks[quoteIdxs[1]].SetWhitespaceBefore(false)   // exit DS

	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms), "only outside-quote overused adjective should match")
	require.Equal(t, "nice", extractSurface(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "good" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "good" at least
}

func TestStyleTooOftenUsedAdjectiveRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedAdjectiveRule()
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("A good car. A good house. A good idea.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

// --- hand-built AnalyzedTokenReadings helpers (DE style_too_often_used pattern) ---

func styleSentStart() *languagetool.AnalyzedTokenReadings {
	tag := languagetool.SentenceStartTagName
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &tag, nil), 0)
}

func styleATR(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
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

func styleATRMulti(token, lemma string, tags ...string) *languagetool.AnalyzedTokenReadings {
	if len(tags) == 0 {
		return styleATR(token, "", lemma)
	}
	var readings []*languagetool.AnalyzedToken
	for _, tag := range tags {
		tt := tag
		var l *string
		if lemma != "" {
			ll := lemma
			l = &ll
		}
		readings = append(readings, languagetool.NewAnalyzedToken(token, &tt, l))
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, 0)
}

func styleUTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func styleIsPunct(token string) bool {
	switch token {
	case ",", ".", ";", ":", "!", "?", "…", ")", "]", "}", "»", "«", "“", "”", "„", "'", "\"", "’", "‘":
		return true
	}
	return false
}

func styleWithPositions(toks ...*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	// Java AnalyzedTokenReadings positions use String.length() (UTF-16).
	pos := 0
	for i, t := range toks {
		if t == nil {
			continue
		}
		t.SetStartPos(pos)
		n := styleUTF16Len(t.GetToken())
		if n == 0 {
			continue
		}
		pos += n
		var next *languagetool.AnalyzedTokenReadings
		for j := i + 1; j < len(toks); j++ {
			if toks[j] != nil && styleUTF16Len(toks[j].GetToken()) > 0 {
				next = toks[j]
				break
			}
		}
		if next == nil {
			continue
		}
		// Space between words; no space before punctuation.
		if !styleIsPunct(next.GetToken()) && !styleIsPunct(t.GetToken()) {
			pos++ // inter-word space
			next.SetWhitespaceBefore(true)
		} else if styleIsPunct(t.GetToken()) && !styleIsPunct(next.GetToken()) {
			pos++ // space after punct before word
			next.SetWhitespaceBefore(true)
		} else {
			next.SetWhitespaceBefore(false)
		}
	}
	return toks
}

func extractSurface(sent *languagetool.AnalyzedSentence, m *rules.RuleMatch) string {
	// Match positions are absolute within the single-sentence document offset 0.
	from, to := m.GetFromPos(), m.GetToPos()
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok.GetStartPos() == from && tok.GetEndPos() == to {
			return tok.GetToken()
		}
	}
	// Fallback: raw sentence text slice if available
	_ = to
	return ""
}
