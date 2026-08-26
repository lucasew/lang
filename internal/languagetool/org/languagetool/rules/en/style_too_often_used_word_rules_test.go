package en

import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Behavior matrix twins for EN StyleTooOftenUsed{Noun,Verb,Adjective}Rule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedWordRules_IDs(t *testing.T) {
	require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", NewStyleTooOftenUsedNounRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_VERB_EN", NewStyleTooOftenUsedVerbRule().GetID())
	require.Equal(t, "TOO_OFTEN_USED_ADJECTIVE_EN", NewStyleTooOftenUsedAdjectiveRule().GetID())
}

// --- StyleTooOftenUsedNounRule (EN) ---
// Behavior matrix twin for org.languagetool.rules.en.StyleTooOftenUsedNounRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedNounRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedNounRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedNounRule()
	require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", rule.GetID())
	require.Equal(t, "Statistical Style Analysis: Overused Noun", rule.GetDescription())
	// Java: "Show when a noun is used more often than ...% of all nouns:"
	require.Equal(t, "Show when a noun is used more often than ...% of all nouns:", rule.GetConfigureText())
	require.Equal(t, 5, rule.MinPercent)        // DEFAULT_MIN_PERCENT
	require.Equal(t, 5, rule.DefaultMinPercent) // Java defaultMinPercent
	require.Equal(t, 100, rule.MinWordCount)    // Abstract MIN_WORD_COUNT
	require.True(t, rule.IsDefaultOff())        // DEFAULT_ACTIVATION = false
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	cat := rule.GetCategory()
	require.NotNil(t, cat)
	require.Equal(t, "CREATIVE_WRITING", cat.GetID().String())
	require.True(t, cat.IsDefaultOff())

	// getLimitMessage(limit)
	require.Equal(t,
		"The noun is used more than 5% times of all nouns. It may be better to replace it with a synonym.",
		rule.LimitMessage(5))
	require.Equal(t,
		"The noun is used more than 12% times of all nouns. It may be better to replace it with a synonym.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedNounRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule()

	// isToCountedWord = hasPosTagStartingWith("NN")
	require.True(t, rule.IsToCountedWord(styleATR("cat", "NN", "cat")))
	require.True(t, rule.IsToCountedWord(styleATR("cats", "NNS", "cat")))
	// NNP also starts with "NN" (Java isToCountedWord true; isException also true)
	require.True(t, rule.IsToCountedWord(styleATR("London", "NNP", "London")))
	require.False(t, rule.IsToCountedWord(styleATR("good", "JJ", "good")))
	require.False(t, rule.IsToCountedWord(styleATR("run", "VB", "run")))

	// isException = NNP|IN|JJ|RB|VB starting tags (token at index n)
	tokens := []*languagetool.AnalyzedTokenReadings{
		styleSentStart(),
		styleATR("London", "NNP", "London"),
		styleATR("in", "IN", "in"),
		styleATR("good", "JJ", "good"),
		styleATR("quickly", "RB", "quickly"),
		styleATR("run", "VB", "run"),
		styleATR("cat", "NN", "cat"),
		styleATR("cats", "NNS", "cat"),
	}
	require.True(t, rule.IsException(tokens, 1))  // NNP
	require.True(t, rule.IsException(tokens, 2))  // IN
	require.True(t, rule.IsException(tokens, 3))  // JJ
	require.True(t, rule.IsException(tokens, 4))  // RB
	require.True(t, rule.IsException(tokens, 5))  // VB
	require.False(t, rule.IsException(tokens, 6)) // pure NN
	require.False(t, rule.IsException(tokens, 7)) // pure NNS

	// Multi-tag token that is NN and also VB → exception (Java ORs all tags)
	nnAndVB := styleATRMulti("run", "run", "NN", "VB")
	require.True(t, rule.IsToCountedWord(nnAndVB))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{styleSentStart(), nnAndVB}, 1))
}

func TestStyleTooOftenUsedNounRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule()
	// Java: getLemmaForPosTagStartsWith("NN", token)
	require.Equal(t, "cat", rule.ToAddedLemma(styleATR("cats", "NNS", "cat")))
	require.Equal(t, "house", rule.ToAddedLemma(styleATR("house", "NN", "house")))
	// No NN reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(styleATR("good", "JJ", "good")))
	// NN reading with nil lemma → empty string path (Java returns null lemma)
	pos := "NN"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
}

func TestStyleTooOftenUsedNounRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedNounRule()
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("cat", "NN", "cat"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("bad", "JJ", "bad"),
		styleATR("cat", "NN", "cat"),
		styleATR(".", ".", "."),
	))
	// Short text: only 2 counted NN lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedNounRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedNounRule()
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("good", "JJ", "good"),
		styleATR("cat", "NN", "cat"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("bad", "JJ", "bad"),
		styleATR("cat", "NN", "cat"),
		styleATR(".", ".", "."),
	))
	// "cat" is 100% of nouns ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "The noun is used more than 5% times of all nouns. It may be better to replace it with a synonym."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_NOUN_EN", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("nice", "JJ", "nice"),
		styleATR("house", "NN", "house"),
		styleATR(".", ".", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedNounRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// Tokens that have NN but also VB/JJ/... are exceptions and must not enter the map.
	// NNP is isToCountedWord (NN*) but isException (NNP*) → not counted.
	rule := NewStyleTooOftenUsedNounRule()
	rule.MinWordCount = 0

	// Only pure NN "cat"; "run" has NN+VB → isException; "London" is NNP → isException
	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("A", "DT", "a"),
		styleATR("cat", "NN", "cat"),
		styleATRMulti("run", "run", "NN", "VB"),
		styleATR("London", "NNP", "London"),
		styleATR(".", ".", "."),
	))
	// only "cat" counted → 100% ≥ 5 → one match on "cat" only
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "cat", extractSurface(s, ms[0]))
}

func TestStyleTooOftenUsedNounRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	rule := NewStyleTooOftenUsedNounRule()
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	// Sentence: He said "cat cat cat" and dog runs.
	// Inside quotes: three "cat" NN; outside: "dog" NN.
	// With exclusion: only "dog" counted → 100% of nouns → match on "dog" only.
	openQ := styleATR("\"", "``", "\"")
	cat1 := styleATR("cat", "NN", "cat")
	cat1.SetWhitespaceBefore(false)
	cat2 := styleATR("cat", "NN", "cat")
	cat2.SetWhitespaceBefore(true)
	cat3 := styleATR("cat", "NN", "cat")
	cat3.SetWhitespaceBefore(true)
	endQ := styleATR("\"", "''", "\"")
	endQ.SetWhitespaceBefore(false)
	dog := styleATR("dog", "NN", "dog")
	dog.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("He", "PRP", "he"),
		styleATR("said", "VBD", "say"),
		openQ,
		cat1,
		cat2,
		cat3,
		endQ,
		styleATR("and", "CC", "and"),
		dog,
		styleATR("runs", "VBZ", "run"),
		styleATR(".", ".", "."),
	))
	toks := s.GetTokensWithoutWhitespace()
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
	require.Equal(t, 1, len(ms), "only outside-quote overused noun should match")
	require.Equal(t, "dog", extractSurface(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "cat" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "cat" at least
}

func TestStyleTooOftenUsedNounRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedNounRule()
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("A cat ran. A cat slept. A cat meowed.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}

// --- StyleTooOftenUsedVerbRule (EN) ---
// Behavior matrix twin for org.languagetool.rules.en.StyleTooOftenUsedVerbRule
// (+ AbstractStyleTooOftenUsedWordRule match algorithm). Java has no dedicated *Test class.

func TestStyleTooOftenUsedVerbRule_JavaMeta(t *testing.T) {
	// Java StyleTooOftenUsedVerbRule + AbstractStyleTooOftenUsedWordRule ctor.
	rule := NewStyleTooOftenUsedVerbRule()
	require.Equal(t, "TOO_OFTEN_USED_VERB_EN", rule.GetID())
	require.Equal(t, "Statistical Style Analysis: Overused Verb", rule.GetDescription())
	// Java: "Show when a verb is used more often than ...% of all verbs:"
	require.Equal(t, "Show when a verb is used more often than ...% of all verbs:", rule.GetConfigureText())
	require.Equal(t, 5, rule.MinPercent)        // DEFAULT_MIN_PERCENT
	require.Equal(t, 5, rule.DefaultMinPercent) // Java defaultMinPercent
	require.Equal(t, 100, rule.MinWordCount)    // Abstract MIN_WORD_COUNT
	require.True(t, rule.IsDefaultOff())        // DEFAULT_ACTIVATION = false
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	cat := rule.GetCategory()
	require.NotNil(t, cat)
	require.Equal(t, "CREATIVE_WRITING", cat.GetID().String())
	require.True(t, cat.IsDefaultOff())

	// getLimitMessage(limit)
	require.Equal(t,
		"The verb is used more than 5% times of all verbs. It may be better to replace it with a synonym.",
		rule.LimitMessage(5))
	require.Equal(t,
		"The verb is used more than 12% times of all verbs. It may be better to replace it with a synonym.",
		rule.LimitMessage(12))

	// getRuleOptions: RuleOption(defaultMinPercent, getConfigureText(), 1, 100)
	opts := rule.GetRuleOptions()
	require.Len(t, opts, 1)
	require.Equal(t, 5, opts[0].DefaultValue)
	require.Equal(t, rule.GetConfigureText(), opts[0].ConfigureText)
	require.Equal(t, 1, opts[0].MinConfigurableValue)
	require.Equal(t, 100, opts[0].MaxConfigurableValue)
}

func TestStyleTooOftenUsedVerbRule_IsToCountedWordAndException(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule()

	// isToCountedWord = hasPosTagStartingWith("VB")
	require.True(t, rule.IsToCountedWord(styleATR("run", "VB", "run")))
	require.True(t, rule.IsToCountedWord(styleATR("runs", "VBZ", "run")))
	require.True(t, rule.IsToCountedWord(styleATR("ran", "VBD", "run")))
	require.True(t, rule.IsToCountedWord(styleATR("running", "VBG", "run")))
	require.True(t, rule.IsToCountedWord(styleATR("been", "VBN", "be")))
	require.False(t, rule.IsToCountedWord(styleATR("good", "JJ", "good")))
	require.False(t, rule.IsToCountedWord(styleATR("cat", "NN", "cat")))

	// isException = hasAnyLemma("be","have","do") || IN* || NN*
	tokens := []*languagetool.AnalyzedTokenReadings{
		styleSentStart(),
		styleATR("is", "VBZ", "be"),
		styleATR("have", "VB", "have"),
		styleATR("did", "VBD", "do"),
		styleATR("in", "IN", "in"),
		styleATR("cat", "NN", "cat"),
		styleATR("run", "VB", "run"),
		styleATR("runs", "VBZ", "run"),
	}
	require.True(t, rule.IsException(tokens, 1))  // lemma be
	require.True(t, rule.IsException(tokens, 2))  // lemma have
	require.True(t, rule.IsException(tokens, 3))  // lemma do
	require.True(t, rule.IsException(tokens, 4))  // IN
	require.True(t, rule.IsException(tokens, 5))  // NN
	require.False(t, rule.IsException(tokens, 6)) // pure VB run
	require.False(t, rule.IsException(tokens, 7)) // pure VBZ run

	// Multi-tag token that is VB and also NN → exception (Java ORs tags / lemmas)
	vbAndNN := styleATRMulti("run", "run", "VB", "NN")
	require.True(t, rule.IsToCountedWord(vbAndNN))
	require.True(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{styleSentStart(), vbAndNN}, 1))
}

func TestStyleTooOftenUsedVerbRule_ToAddedLemma(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule()
	// Java: getLemmaForPosTagStartsWith("VB", token)
	require.Equal(t, "run", rule.ToAddedLemma(styleATR("runs", "VBZ", "run")))
	require.Equal(t, "run", rule.ToAddedLemma(styleATR("running", "VBG", "run")))
	// No VB reading → null/empty (no surface invent)
	require.Equal(t, "", rule.ToAddedLemma(styleATR("good", "JJ", "good")))
	// VB reading with nil lemma → empty string path (Java returns null lemma)
	pos := "VB"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, nil), 0)
	require.Equal(t, "", rule.ToAddedLemma(tok))
}

func TestStyleTooOftenUsedVerbRule_Match_MinWordCountGate(t *testing.T) {
	// Java getTooOftenUsedWords: if numWords < MIN_WORD_COUNT (100) → empty.
	rule := NewStyleTooOftenUsedVerbRule()
	require.Equal(t, 100, rule.MinWordCount)

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Cats", "NNS", "cat"),
		styleATR("run", "VB", "run"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Dogs", "NNS", "dog"),
		styleATR("run", "VB", "run"),
		styleATR(".", ".", "."),
	))
	// Short text: only 2 counted VB lemmas → under 100 → no matches
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedVerbRule_Match_OverusedLemma(t *testing.T) {
	// Bypass MIN_WORD_COUNT only for percent-threshold unit test (Java production uses 100).
	rule := NewStyleTooOftenUsedVerbRule()
	rule.MinWordCount = 0

	s1 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Cats", "NNS", "cat"),
		styleATR("run", "VB", "run"),
		styleATR(".", ".", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Dogs", "NNS", "dog"),
		styleATR("run", "VB", "run"),
		styleATR(".", ".", "."),
	))
	// "run" is 100% of verbs ≥ DEFAULT_MIN_PERCENT 5 → both hits
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(ms), 2)
	wantMsg := "The verb is used more than 5% times of all verbs. It may be better to replace it with a synonym."
	for _, m := range ms {
		require.Equal(t, wantMsg, m.GetMessage())
		r, ok := m.GetRule().(*rules.AbstractStyleTooOftenUsedWordRule)
		require.True(t, ok)
		require.Equal(t, "TOO_OFTEN_USED_VERB_EN", r.GetID())
	}

	// Two lemmas each 50% ≥ 5 → both flag once
	s3 := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Birds", "NNS", "bird"),
		styleATR("fly", "VB", "fly"),
		styleATR(".", ".", "."),
	))
	ms2 := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})
	require.Equal(t, 2, len(ms2))

	// Raise threshold above 50%: neither flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))
}

func TestStyleTooOftenUsedVerbRule_Match_ExceptionTagsNotCounted(t *testing.T) {
	// be|have|do lemmas and IN*|NN* are exceptions and must not enter the map.
	// VB+NN multi-tag is isToCountedWord (VB*) but isException (NN*) → not counted.
	rule := NewStyleTooOftenUsedVerbRule()
	rule.MinWordCount = 0

	// Only pure VB "run"; "is" lemma be → isException; "walk" has VB+NN → isException
	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Cats", "NNS", "cat"),
		styleATR("run", "VB", "run"),
		styleATR("is", "VBZ", "be"),
		styleATRMulti("walk", "walk", "VB", "NN"),
		styleATR(".", ".", "."),
	))
	// only "run" counted → 100% ≥ 5 → one match on "run" only
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.Equal(t, 1, len(ms))
	require.Equal(t, "run", extractSurface(s, ms[0]))
}

func TestStyleTooOftenUsedVerbRule_Match_DirectSpeechExclusion(t *testing.T) {
	// Java: withoutDirectSpeech only when setWithoutDirectSpeech(true).
	// Opening quote with next token !isWhitespaceBefore → enter DS; content not counted.
	rule := NewStyleTooOftenUsedVerbRule()
	rule.MinWordCount = 0
	rule.WithoutDirectSpeech = true

	// Sentence: Then "run run run" and walk now.
	// (no outside VB reporting verb — "said"/VBD would also be counted as a verb)
	// Inside quotes: three "run" VB; outside: "walk" VB.
	// With exclusion: only "walk" counted → 100% of verbs → match on "walk" only.
	openQ := styleATR("\"", "``", "\"")
	run1 := styleATR("run", "VB", "run")
	run1.SetWhitespaceBefore(false)
	run2 := styleATR("run", "VB", "run")
	run2.SetWhitespaceBefore(true)
	run3 := styleATR("run", "VB", "run")
	run3.SetWhitespaceBefore(true)
	endQ := styleATR("\"", "''", "\"")
	endQ.SetWhitespaceBefore(false)
	walk := styleATR("walk", "VB", "walk")
	walk.SetWhitespaceBefore(true)

	s := languagetool.NewAnalyzedSentence(styleWithPositions(
		styleSentStart(),
		styleATR("Then", "RB", "then"),
		openQ,
		run1,
		run2,
		run3,
		endQ,
		styleATR("and", "CC", "and"),
		walk,
		styleATR("now", "RB", "now"),
		styleATR(".", ".", "."),
	))
	toks := s.GetTokensWithoutWhitespace()
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
	require.Equal(t, 1, len(ms), "only outside-quote overused verb should match")
	require.Equal(t, "walk", extractSurface(s, ms[0]))

	// Default WithoutDirectSpeech=false: quotes do not exclude; "run" is majority
	rule.WithoutDirectSpeech = false
	msAll := rule.MatchList([]*languagetool.AnalyzedSentence{s})
	require.GreaterOrEqual(t, len(msAll), 3) // three "run" at least
}

func TestStyleTooOftenUsedVerbRule_Match_UntaggedNoInvent(t *testing.T) {
	// Soft/untagged SplitAndAnalyze must not invent POS → no matches.
	rule := NewStyleTooOftenUsedVerbRule()
	rule.MinWordCount = 0
	plain := languagetool.SplitAndAnalyze("Cats run. Dogs run. Birds run.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
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
