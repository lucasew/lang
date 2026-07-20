package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atr(token string, start int) *languagetool.AnalyzedTokenReadings {
	r := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, nil, nil), start)
	return r
}

func TestPatternRuleMatcherSimple(t *testing.T) {
	// tokens: This(0) is(5) foo(8) bar(12)
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("This", 0),
		atr("is", 5),
		atr("foo", 8),
		atr("bar", 12),
	}
	// fix end positions roughly
	sent := testSentence(toks...)
	rule := NewPatternRule("DEMO", "en",
		[]*PatternToken{Token("foo"), Token("bar")},
		"demo", "found foo bar", "short")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, 8, matches[0].FromPos)
	// GetEndPos of bar
	require.Equal(t, toks[3].GetEndPos(), matches[0].ToPos)
	require.Equal(t, "found foo bar", matches[0].Message)
}

func TestPatternRuleMatcherOptional(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("hello", 0),
		atr("world", 6),
	}
	sent := testSentence(toks...)
	opt := Token("the")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("OPT", "en",
		[]*PatternToken{opt, Token("hello"), Token("world")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestPatternRuleMatcherNoMatch(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("hello", 0)}
	sent := testSentence(toks...)
	rule := NewPatternRule("X", "en", []*PatternToken{Token("bye")}, "d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)
}

func TestPatternRuleMatcher_CreateRuleMatchCaseAndSuggestions(t *testing.T) {
	// Message with <suggestion> — startsWithUppercase when match starts capital.
	rule := NewPatternRule("CASE", "en",
		[]*PatternToken{Token("Hello")},
		"d",
		`Use <suggestion>hi</suggestion> instead`,
		"")
	sent := testSentence(atr("Hello", 0))
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.NotEmpty(t, ms[0].GetSuggestedReplacements())
	// startsWithUppercase → UppercaseFirstChar("hi") = "Hi"
	require.Equal(t, "Hi", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, ms[0].FromPos, ms[0].PatternFromPos)

	// PLEASE_SPELL_ME without suggestion tags → no match (Java createRuleMatch)
	rule2 := NewPatternRule("SPELL", "en",
		[]*PatternToken{Token("xyz")},
		"d",
		"bad "+PleaseSpellMe+" only",
		"")
	sent2 := testSentence(atr("xyz", 0))
	ms, err = rule2.Match(sent2)
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin of RuleMatch ctor LinkedHashSet<SuggestedReplacement>: ordered unique by string.
func TestCreateRuleMatch_SuggestionDedupLinkedHashSet(t *testing.T) {
	// Duplicate "foo" in message + outMsg; case lower (no startsWithUpper on "hi")
	rule := NewPatternRule("DEDUP", "en",
		[]*PatternToken{Token("hi")},
		"d",
		`Use <suggestion>foo</suggestion> or <suggestion>bar</suggestion> or <suggestion>foo</suggestion>`,
		"")
	// suggestionsOutMsg also has foo again
	rule.SuggestionsOutMsg = `<suggestion>foo</suggestion><suggestion>baz</suggestion>`
	sent := testSentence(atr("hi", 0))
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	// Order preserved; foo only once
	require.Equal(t, []string{"foo", "bar", "baz"}, ms[0].GetSuggestedReplacements())
}

// Twin of RuleMatch ctor: empty suggestion body is kept (no invent skip of "").
func TestCreateRuleMatch_EmptySuggestionBodyKept(t *testing.T) {
	rule := NewPatternRule("EMPTY", "en",
		[]*PatternToken{Token("hi")},
		"d",
		`Empty <suggestion></suggestion> ok`,
		"")
	sent := testSentence(atr("hi", 0))
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, []string{""}, ms[0].GetSuggestedReplacements())
}

// Java AbstractPatternRulePerformer: scope=next on element blocks when immediate next matches
// (prevSkipNext == 0 path).
func TestPatternRuleMatcher_NextExceptionImmediate(t *testing.T) {
	// pattern: can + verb; exception next be|do → "can be" does not match
	can := Token("can")
	can.AddNextException(NewPatternToken("be", false, false, false))
	can.AddNextException(NewPatternToken("do", false, false, false))
	rule := NewPatternRule("NEXT", "en",
		[]*PatternToken{can, Token("run")},
		"d", "m", "")
	// "can be" — next exception fires
	sent1 := testSentence(atr("can", 0), atr("be", 4),)
	ms, err := rule.Match(sent1)
	require.NoError(t, err)
	require.Empty(t, ms)
	// "can run" — ok
	sent2 := testSentence(atr("can", 0), atr("run", 4),)
	ms, err = rule.Match(sent2)
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

// Java prevSkip==0 uses only getAnalyzedToken(0) for scope=next — other readings ignored.
func TestPatternRuleMatcher_NextExceptionFirstReadingOnly(t *testing.T) {
	can := Token("can")
	can.AddNextException(NewPatternToken("be", false, false, false))
	rule := NewPatternRule("NEXT0", "en",
		[]*PatternToken{can, Token("run")},
		"d", "m", "")
	// Next token "x" with readings [run, be]: first reading is run → no block; pattern needs "run"
	// After "can", next must be "run" for pattern. Readings of "run" token: first=run ok.
	// Build next token with first reading not "be", second "be" — exception must not fire.
	nn := "NN"
	runTok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("run", &nn, nil),
		languagetool.NewAnalyzedToken("be", &nn, nil), // second reading would match exception
	}, 4)
	// Force surface token to "run" for GetToken()
	sent := testSentence(atr("can", 0), runTok)
	// Replace second token if testSentence rebuilds — build manually:
	ss := languagetool.SentenceStartTagName
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atr("can", 0),
		runTok,
	}
	sent = languagetool.NewAnalyzedSentence(tokens)
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1, "first reading is run; second-reading 'be' must not block (Java getAnalyzedToken(0))")

	// First reading is "be" → block
	beFirst := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("be", &nn, nil),
		languagetool.NewAnalyzedToken("run", &nn, nil),
	}, 4)
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atr("can", 0),
		beFirst,
	})
	// Pattern second token is Token("run") — surface of beFirst is still "be" from first token...
	// AnalyzedTokenReadings.GetToken is first reading's token typically.
	ms, err = rule.Match(sent2)
	require.NoError(t, err)
	// Either blocked by next-exception on first reading "be", or surface "be" != "run"
	require.Empty(t, ms)
}

// Java: when prev element has skip>0, its scope=next exception rejects skip-window
// candidates that match the exception (prevMatched path with prevSkipNext > 0).
func TestPatternRuleMatcher_NextExceptionSkipWindow(t *testing.T) {
	// Pattern: see (skip=3, next-exception "bad") + good
	// Immediate next of "see" must NOT be "bad" (path 2 when matching see with prevSkip==0).
	// Within the skip window, "bad" cannot be the match position for "good" (path 1).
	see := Token("see")
	see.SetSkipNext(3)
	see.AddNextException(NewPatternToken("bad", false, false, false))
	rule := NewPatternRule("SKIPNEXT", "en",
		[]*PatternToken{see, Token("good")},
		"d", "m", "")

	// see foo bad good → match; "bad" skipped over as non-candidate for "good"
	sent := testSentence(atr("see", 0), atr("foo", 4), atr("bad", 8), atr("good", 12),)
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
	// [SENT_START, see, foo, bad, good] → good is index 4
	require.Equal(t, sent.GetTokensWithoutWhitespace()[4].GetEndPos(), ms[0].ToPos)

	// see bad good → path 2 blocks matching "see" (immediate next is exception)
	sent2 := testSentence(atr("see", 0), atr("bad", 4), atr("good", 8),)
	ms, err = rule.Match(sent2)
	require.NoError(t, err)
	require.Empty(t, ms)

	// see good — adjacent; next of see is good (not bad) → match
	sent3 := testSentence(atr("see", 0), atr("good", 4),)
	ms, err = rule.Match(sent3)
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

// testPOS builds AnalyzedTokenReadings with a POS tag at start.
func atrPOS(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, strPtr(pos), strPtr(token)), start)
}

// TestPatternRuleMatcher_UnificationAgreement ports Java testUnification semantics:
// <unify> requires shared feature type across tokens; negate="yes" fires on mismatch.
func TestPatternRuleMatcher_UnificationAgreement(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "sg", func() *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: "NN", Regexp: false})
		return pt
	}())
	cfg.SetEquivalence("number", "pl", func() *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: "NNS", Regexp: false})
		return pt
	}())

	mkTok := func(pos string, last, neg bool) *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: pos, Regexp: false})
		pt.SetUnification(map[string][]string{"number": {}})
		if last {
			pt.SetLastInUnification()
		}
		if neg {
			pt.SetUniNegation()
		}
		return pt
	}

	// Positive unify: match only when both tokens share number (both sg).
	t1, t2 := mkTok("NN", false, false), mkTok("NN", true, false)
	agree := NewPatternRule("AGREE", "en", []*PatternToken{t1, t2}, "d", "agree", "")
	agree.UnifierConfig = cfg

	// cat(NN) + sits(VBZ) — second token pattern is NN so won't surface-match; use two NN.
	sg1 := atrPOS("cat", "NN", 0)
	sg2 := atrPOS("dog", "NN", 4)
	sentAgree := testSentence(sg1, sg2)
	ms, err := agree.Match(sentAgree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "same number should unify and match")

	// cat(NN) + dogs(NNS) — second pattern token is NN, won't match surface.
	// Change second pattern to accept both via empty postag / separate patterns.
	// Use two POS-open tokens with only unify constraining number via equivalence
	// on readings: pattern postag matches both NN and NNS via regexp.
	p1 := NewPatternToken("", false, false, false)
	p1.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	p1.SetUnification(map[string][]string{"number": {}})
	p2 := NewPatternToken("", false, false, false)
	p2.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	p2.SetUnification(map[string][]string{"number": {}})
	p2.SetLastInUnification()
	agree2 := NewPatternRule("AGREE2", "en", []*PatternToken{p1, p2}, "d", "agree2", "")
	agree2.UnifierConfig = cfg

	pl := atrPOS("dogs", "NNS", 4)
	sentDisagree := testSentence(sg1, pl)
	ms, err = agree2.Match(sentDisagree)
	require.NoError(t, err)
	require.Empty(t, ms, "different number must not unify without negate")

	ms, err = agree2.Match(sentAgree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "same number still matches")

	// Negated unify: fire when tokens do NOT share number.
	n1 := NewPatternToken("", false, false, false)
	n1.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	n1.SetUnification(map[string][]string{"number": {}})
	n2 := NewPatternToken("", false, false, false)
	n2.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	n2.SetUnification(map[string][]string{"number": {}})
	n2.SetLastInUnification()
	n2.SetUniNegation()
	negRule := NewPatternRule("NEG", "en", []*PatternToken{n1, n2}, "d", "neg", "")
	negRule.UnifierConfig = cfg

	ms, err = negRule.Match(sentDisagree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "negate unify should fire on number mismatch")

	ms, err = negRule.Match(sentAgree)
	require.NoError(t, err)
	require.Empty(t, ms, "negate unify must not fire when numbers agree")

	// Fail-closed: unify without config never matches.
	noCfg := NewPatternRule("NOCFG", "en", []*PatternToken{p1, p2}, "d", "x", "")
	ms, err = noCfg.Match(sentAgree)
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestRepeatedAndConsistencyTransformers(t *testing.T) {
	// Java: only ids starting with getConsistencyRulePrefix() are transformed.
	// Convention PREFIX_GROUPOFRULES_FEATURE → main = parts[0]+"_"+parts[1].
	a1 := NewAbstractPatternRule("PREFIXFORCONSISTENCYRULES_STYLE_feat1", "d", "en", nil, false)
	a2 := NewAbstractPatternRule("PREFIXFORCONSISTENCYRULES_STYLE_feat2", "d", "en", nil, false)
	b := NewAbstractPatternRule("OTHER", "d", "en", nil, false)
	noPrefix := NewAbstractPatternRule("STYLE_A_feat1", "d", "en", nil, false)
	ct := NewConsistencyPatternRuleTransformer("en")
	rem, tr := TransformPatternRules([]*AbstractPatternRule{a1, a2, b, noPrefix}, ct)
	require.Len(t, rem, 2)
	require.Equal(t, "OTHER", rem[0].ID)
	require.Equal(t, "STYLE_A_feat1", rem[1].ID)
	require.Len(t, tr, 1)
	cr, ok := tr[0].(*ConsistencyPatternRule)
	require.True(t, ok)
	require.Equal(t, "PREFIXFORCONSISTENCYRULES_STYLE", cr.GetID())

	require.Equal(t, "PREFIXFORCONSISTENCYRULES_STYLE", GetMainRuleId("PREFIXFORCONSISTENCYRULES_STYLE_feat1"))
	require.Equal(t, "feat1", GetFeature("PREFIXFORCONSISTENCYRULES_STYLE_feat1"))
	require.Equal(t, "STYLE_A", GetMainRuleId("STYLE_A_feat1"))
	require.Equal(t, "feat1", GetFeature("STYLE_A_feat1"))

	// Java: only getMinPrevMatches() > 0 enters RepeatedPatternRuleTransformer.
	r1 := NewAbstractPatternRule("REP", "d", "en", nil, false)
	r1.MinPrevMatches = 2
	r1.DistanceTokens = 10
	r2 := NewAbstractPatternRule("REP", "d", "en", nil, false)
	r2.MinPrevMatches = 2
	rt := NewRepeatedPatternRuleTransformer("en")
	rem2, tr2 := TransformPatternRules([]*AbstractPatternRule{r1, r2}, rt)
	require.Empty(t, rem2)
	require.Len(t, tr2, 1)
	// Without min_prev_matches, rules stay remaining (even with shared id).
	r3 := NewAbstractPatternRule("PLAIN", "d", "en", nil, false)
	r4 := NewAbstractPatternRule("PLAIN", "d", "en", nil, false)
	rem3, tr3 := TransformPatternRules([]*AbstractPatternRule{r3, r4}, rt)
	require.Len(t, rem3, 2)
	require.Empty(t, tr3)
}

// Java AbstractPatternRulePerformer: min=0 with foundNext prefers skipping optional
// when the next required element also matches the same token.
func TestPatternRuleMatcher_OptionalFoundNextPrefersSkip(t *testing.T) {
	// Pattern: optional "the" then "the" required — sentence "the".
	// foundNext: next "the" matches at same position → optional positions=0, required takes token.
	opt := Token("the")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("OPT_SKIP", "en",
		[]*PatternToken{opt, Token("the")},
		"d", "m", "")
	sent := testSentence(atr("the", 0))
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)

	// Pattern: optional "a" then "the" — sentence "the" still matches (optional absent via foundNext).
	opt2 := Token("a")
	opt2.SetMinOccurrence(0)
	rule2 := NewPatternRule("OPT_ABSENT", "en",
		[]*PatternToken{opt2, Token("the")},
		"d", "m", "")
	ms2, err := rule2.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms2, 1)
}

func TestPatternRuleMatcher_SkipMaxTokens(t *testing.T) {
	// max=2: match "x" "x" as one pattern element when consecutive.
	pt := Token("x")
	pt.SetMaxOccurrence(2)
	rule := NewPatternRule("MAX2", "en",
		[]*PatternToken{pt, Token("y")},
		"d", "m", "")
	// x x y — start=0 consumes two x via skipMaxTokens; start=1 may also match one x.
	sent := testSentence(atr("x", 0), atr("x", 2), atr("y", 4))
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// Longest/first span from start 0 covers both x tokens.
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, atr("y", 4).GetEndPos(), ms[0].ToPos)
}

func TestMatchPreservesCase_PleaseSpellMeJava(t *testing.T) {
	// Java: if msg contains PLEASE_SPELL_ME anywhere, sugStart advances by its length
	// after first <suggestion> open (even when not immediately after the tag).
	m := NewMatch("", "", false, "", "", CaseStartUpper, false, false, IncludeNone)
	require.True(t, m.ConvertsCase())

	// <suggestion>\1 → backslash at sugStart after open → does not preserve
	msg := suggestionStartTag + `\1foo</suggestion>`
	require.False(t, matchPreservesCase([]*Match{m}, msg))

	// PLEASE_SPELL_ME right after open: Java sugStart lands on '\\' → false
	msg2 := suggestionStartTag + PleaseSpellMe + `\1`
	require.False(t, matchPreservesCase([]*Match{m}, msg2))

	// PLEASE_SPELL_ME later still adds its length (Java bug-for-bug) — no panic.
	msg3 := suggestionStartTag + `XXXX` + PleaseSpellMe
	_ = matchPreservesCase([]*Match{m}, msg3)

	// No convertsCase match → true
	mNone := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	require.True(t, matchPreservesCase([]*Match{mNone}, msg))
}
