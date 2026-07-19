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
	sent := languagetool.NewAnalyzedSentence(toks)
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

	// empty synthesis → (token)
	m2 := NewMatch("VBZ", "", false, "", "", CaseNone, false, false, IncludeNone)
	ms2 := NewMatchStateWithSynth(m2, LanguageSynthesizer("en-test-synth"))
	ms2.SetToken(tok)
	forms2 := ms2.ToFinalString("en-test-synth")
	require.Equal(t, []string{"(cat)"}, forms2)

	// postag_regexp with replace via GetTargetPosTag path
	m3 := NewMatch("NN.*", "NNS", true, "", "", CaseNone, false, false, IncludeNone)
	// Wait: postag is regexp pattern, postag_replace transforms matched tag
	// match postag="NN.*" postag_regexp="yes" postag_replace="NNS" → target NNS
	m3 = NewMatch("NN.*", "NNS", true, "", "", CaseNone, false, false, IncludeNone)
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

	// strip pleasespellme but keep good suggestion
	in3 := `<suggestion>` + PleaseSpellMe + `dogs</suggestion>`
	out3 := removeSuppressMisspelled(in3)
	require.Equal(t, `<suggestion>dogs</suggestion>`, out3)
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
