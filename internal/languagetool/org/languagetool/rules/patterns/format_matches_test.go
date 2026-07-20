package patterns

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestProcessRuleMessage_MatchTagAndLegacy(t *testing.T) {
	raw := `Use <suggestion><match no="1" case_conversion="startupper"/></suggestion> not \2.`
	msg, matches := ProcessRuleMessage(raw)
	require.NotContains(t, msg, "<match")
	require.Contains(t, msg, `\1`)
	require.Contains(t, msg, `\2`)
	require.GreaterOrEqual(t, len(matches), 2)
	// first is real match (case conversion), second is legacy bare \2
	require.Equal(t, CaseStartUpper, matches[0].GetCaseConversionType())
	require.True(t, matches[1].IsInMessageOnly())
}

func TestFormatMatches_RegexAndCase(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("hello", 0),
		atr("world", 6),
	}
	// \1 with case startupper
	m := NewMatch("", "", false, "", "", CaseStartUpper, false, false, IncludeNone)
	m.SetInMessageOnly(false)
	msg := FormatMatches(toks, []int{1, 1}, 0, `Say \1`, []*Match{m}, "en")
	require.Equal(t, "Say Hello", msg)

	// regex replace (Java regexp_match on whole token surface)
	m2 := NewMatch("", "", false, `(?i)saas`, "SaaS", CaseNone, false, false, IncludeNone)
	toks2 := []*languagetool.AnalyzedTokenReadings{atr("saas", 0)}
	msg2 := FormatMatches(toks2, []int{1}, 0, `\1`, []*Match{m2}, "en")
	require.Equal(t, "SaaS", msg2)
}

func TestFormatMatches_OptionalZeroPosition(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("foo", 0),
		atr("bar", 4),
	}
	// pattern: optional + foo + bar; optional absent → positions [0,1,1]
	// \2 refers to foo (element index 1)
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	m.SetInMessageOnly(true)
	// two backrefs \2 — need matches for each occurrence; only \2 once
	msg := FormatMatches(toks, []int{0, 1, 1}, 0, `got \2`, []*Match{m}, "en")
	require.Equal(t, "got foo", msg)
}

func TestPatternRuleMatcher_FormatsBackrefs(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("hello", 0),
		atr("world", 6),
	}
	sent := testSentence(toks...)
	raw := `Bad <suggestion><match no="1" case_conversion="allupper"/> \2</suggestion>`
	msg, matches := ProcessRuleMessage(raw)
	display, suggs := extractSuggestions(msg)
	pr := NewPatternRule("T", "en", []*PatternToken{Token("hello"), Token("world")}, "d", display, "")
	pr.SuggestionMatches = matches
	pr.SuggestionTemplates = suggs
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.NotEmpty(t, ms[0].GetSuggestedReplacements())
	require.Equal(t, "HELLO world", ms[0].GetSuggestedReplacements()[0])
}

func TestToFinalString_WithManualSynthesizer(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"cats\tcat\tNNS\n" +
			"cat\tcat\tNN\n",
	))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("en", manual)
	RegisterLanguageSynthesizer("en-test-synth", synth)
	t.Cleanup(func() {
		// leave registered; other tests use different codes
	})

	// token "cat" lemma cat POS NN → synthesize NNS
	posNN := "NN"
	lemma := "cat"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("cat", &posNN, &lemma), 0)

	m := NewMatch("NNS", "", false, "", "", CaseNone, false, false, IncludeNone)
	ms := NewMatchStateWithSynth(m, LanguageSynthesizer("en-test-synth"))
	ms.SetToken(tok)
	forms := ms.ToFinalString("en-test-synth")
	require.Equal(t, []string{"cats"}, forms)

	// exact postag empty synth → empty array (Java TreeSet.toArray), not invent "(token)"
	m2 := NewMatch("VBZ", "", false, "", "", CaseNone, false, false, IncludeNone)
	ms2 := NewMatchStateWithSynth(m2, LanguageSynthesizer("en-test-synth"))
	ms2.SetToken(tok)
	forms2 := ms2.ToFinalString("en-test-synth")
	require.Empty(t, forms2)

	// postag_regexp empty synth → "(token)" only on regexp branch
	mEmptyRE := NewMatch("ZZZ.*", "", true, "", "", CaseNone, false, false, IncludeNone)
	msEmptyRE := NewMatchStateWithSynth(mEmptyRE, LanguageSynthesizer("en-test-synth"))
	msEmptyRE.SetToken(tok)
	require.Equal(t, []string{"(cat)"}, msEmptyRE.ToFinalString("en-test-synth"))

	// postag_regexp with replace via GetTargetPosTag path
	// match postag="NN.*" postag_regexp="yes" postag_replace="NNS" → target NNS
	m3 := NewMatch("NN.*", "NNS", true, "", "", CaseNone, false, false, IncludeNone)
	ms3 := NewMatchStateWithSynth(m3, LanguageSynthesizer("en-test-synth"))
	ms3.SetToken(tok)
	forms3 := ms3.ToFinalString("en-test-synth")
	require.Equal(t, []string{"cats"}, forms3)
}

func TestFormatMatches_PostagSynthesize(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader("dogs\tdog\tNNS\n"))
	require.NoError(t, err)
	RegisterLanguageSynthesizer("en-fmt", synthesis.NewBaseSynthesizer("en", manual))

	pos := "NN"
	lemma := "dog"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("dog", &pos, &lemma), 0),
	}
	m := NewMatch("NNS", "", false, "", "", CaseNone, false, false, IncludeNone)
	m.SetInMessageOnly(false)
	msg := FormatMatches(toks, []int{1}, 0, `plural: \1`, []*Match{m}, "en-fmt")
	require.Equal(t, "plural: dogs", msg)
}

func TestRemoveSuppressMisspelled(t *testing.T) {
	// drop suggestion with mistake marker
	in := `msg <suggestion>` + PleaseSpellMe + `foo` + MistakeMarker + `bar</suggestion> end`
	out := removeSuppressMisspelled(in)
	require.NotContains(t, out, "mistake")
	require.NotContains(t, out, PleaseSpellMe)
	require.Contains(t, out, "msg")
	require.Contains(t, out, "end")

	// drop suggestion with parenthesized non-synth form
	in2 := `<suggestion>` + PleaseSpellMe + `(cats)</suggestion>`
	require.Empty(t, strings.TrimSpace(removeSuppressMisspelled(in2)))

	// strip pleasespellme only after <suggestion> (TAG_AND_PLEASE_SPELL_ME)
	in3 := `<suggestion>` + PleaseSpellMe + `dogs</suggestion>`
	out3 := removeSuppressMisspelled(in3)
	require.Equal(t, `<suggestion>dogs</suggestion>`, out3)

	// bare <pleasespellme/> in message body is NOT stripped here (Java removeSuppressMisspelled);
	// createRuleMatch clears it later when building clearMsg.
	in4 := `msg ` + PleaseSpellMe + ` only`
	require.Equal(t, in4, removeSuppressMisspelled(in4))
}

func TestToFinalString_SuppressMisspelledUsesTagger(t *testing.T) {
	// known word → keep; unknown → MISTAKE
	RegisterLanguageTagger("en-sup", func(token string) []languagetool.TokenTag {
		if token == "dogs" {
			return []languagetool.TokenTag{{POS: "NNS", Lemma: "dog"}}
		}
		return nil // empty = unknown
	})
	m := NewMatch("", "", false, "", "", CaseNone, false, true, IncludeNone) // suppress_misspelled
	require.True(t, m.ChecksSpelling())

	tok := atr("dog", 0)
	ms := NewMatchState(m)
	ms.SetToken(tok)
	// no postag synth — surface after regex none
	forms := ms.ToFinalString("en-sup")
	// surface "dog" — if tagger returns nil for dog → MISTAKE
	require.Equal(t, []string{MistakeMarker}, forms)

	// force surface dogs via regex replace
	m2 := NewMatch("", "", false, "dog", "dogs", CaseNone, false, true, IncludeNone)
	ms2 := NewMatchState(m2)
	ms2.SetToken(tok)
	forms2 := ms2.ToFinalString("en-sup")
	require.Equal(t, []string{"dogs"}, forms2)
}

func TestProcessRuleMessage_PleaseSpellMe(t *testing.T) {
	raw := `<suggestion suppress_misspelled="yes">\1</suggestion>`
	msg, _ := ProcessRuleMessage(raw)
	// after process, suggestion still there; pleasespellme injected before extract
	require.Contains(t, msg, PleaseSpellMe)
}

// Twin of PatternRuleMatcher.formatMatches isPositiveNumber gate:
// backslash + '0' is not a match placeholder (Java StringTools.isPositiveNumber).
func TestFormatMatches_SkipsBackrefZero(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("hello", 0)}
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	msg := FormatMatches(toks, []int{1}, 0, `keep \0 and \1`, []*Match{m}, "en")
	require.Contains(t, msg, `\0`)
	require.Contains(t, msg, "hello")
	require.NotContains(t, msg, `\1`)
}

// Twin of numbersToMatches reuse (Java FIXME branch) with sticky newWay.
// newWay is declared outside the while and stays true after the first Match path —
// bare surface replace never runs for remaining \N; overrun only appends a reused
// Match so the next iteration applies case conversion again.
func TestFormatMatches_NumbersToMatchesReuse(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("word", 0)}
	m := NewMatch("", "", false, "", "", CaseAllUpper, false, false, IncludeNone)
	msg := FormatMatches(toks, []int{1}, 0, `\1 \1 \1`, []*Match{m}, "en")
	// Java bug-for-bug: Match → overrun append → Match → overrun append → Match
	require.Equal(t, "WORD WORD WORD", msg)
}

// Bare path only while newWay is still false (no suggestionMatches).
func TestFormatMatches_BarePathWhenNoSuggestionMatches(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("word", 0)}
	msg := FormatMatches(toks, []int{1}, 0, `\1 \1 \1`, nil, "en")
	require.Equal(t, "word word word", msg)
}

// Twin of bare-path String.replace: all remaining "\\N" in unprocessed suffix replaced.
func TestFormatMatches_BareReplaceAllInSuffix(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("aa", 0),
		atr("bb", 3),
	}
	// No suggestionMatches → bare path; both \1 become "aa" in one replace.
	msg := FormatMatches(toks, []int{1, 1}, 0, `\1 and \1`, nil, "en")
	require.Equal(t, "aa and aa", msg)
}

// Twin of Java bare-path errorMessageProcessed assignment:
// newProcessed = lastIndexOf("\\N") + token.length() on the pre-replace string,
// always assigned (no invent clamp to keep processed cursor).
func TestFormatMatches_BarePathProcessedFromLastIndexOf(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("X", 0)}
	// Two bare \1; single replaceAll in one bare-path step consumes both.
	msg := FormatMatches(toks, []int{1}, 0, `A \1 B \1 C`, nil, "en")
	require.Equal(t, "A X B X C", msg)
	// After bare replace, no leftover backrefs
	require.NotContains(t, msg, `\`)
}

func TestConcatWithoutExtraSpace_WhitespaceOrPunct(t *testing.T) {
	// Java WHITESPACE_OR_PUNCT = [\s,:;.!?].* — Matcher.matches entire rightSide
	require.Equal(t, "x,", concatWithoutExtraSpace("x ", ","))
	require.Equal(t, "x!y", concatWithoutExtraSpace("x ", "!y")) // strip left space
	require.Equal(t, "x y", concatWithoutExtraSpace("x ", "y"))  // no leading punct → keep space
	require.Equal(t, "a</suggestion>", concatWithoutExtraSpace("a ", "</suggestion>"))
}
