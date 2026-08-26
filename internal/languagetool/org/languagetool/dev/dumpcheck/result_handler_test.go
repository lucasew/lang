package dumpcheck

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type idRule struct{ id string }

func (r idRule) GetID() string { return r.id }

func TestResultHandler_Limits(t *testing.T) {
	// Java: checkMaxSentences(++sentenceCount) throws when count reaches max.
	h := NewResultHandler(2, 0)
	sent := NewSentence("Hello world this is long enough.", "test", "T", "", 1)
	require.NoError(t, h.HandleResult(sent, nil, "en")) // count=1
	err := h.HandleResult(sent, nil, "en")              // count=2 → limit
	require.Error(t, err)
	var dl DocumentLimitReachedError
	require.ErrorAs(t, err, &dl)
	require.Equal(t, 2, dl.Limit)
}

func TestResultHandler_ErrorLimit(t *testing.T) {
	// maxErrors=2: first sentence with 1 error → count=1 OK; second adds → count=2 limit
	h := NewResultHandler(0, 2)
	sent := NewSentence("Hello world this is long enough.", "test", "T", "", 1)
	m := rules.NewRuleMatch(idRule{"R1"}, nil, 0, 1, "msg")
	require.NoError(t, h.HandleResult(sent, []*rules.RuleMatch{m}, "en"))
	err := h.HandleResult(sent, []*rules.RuleMatch{m}, "en")
	require.Error(t, err)
	var el ErrorLimitReachedError
	require.ErrorAs(t, err, &el)
}

func TestStdoutHandler_Prints(t *testing.T) {
	var buf strings.Builder
	h := NewStdoutHandler(&buf, 0, 0, 10)
	sent := NewSentence("Hello wrong word here.", "plain", "Title1", "", 1)
	m := rules.NewRuleMatch(idRule{"DEMO"}, nil, 6, 11, "bad word")
	m.SetSuggestedReplacement("right")
	require.NoError(t, h.HandleResult(sent, []*rules.RuleMatch{m}, "en"))
	out := buf.String()
	require.Contains(t, out, "Title: Title1")
	require.Contains(t, out, "DEMO")
	require.Contains(t, out, "right")
	require.Contains(t, out, MarkerStart)
}

func TestSentenceSourceChecker_Run(t *testing.T) {
	src := NewPlainTextSentenceSource(strings.NewReader(
		"First long enough sentence here.\nSecond long enough sentence here.\n"))
	var seen int
	h := NewResultHandler(0, 0)
	h.Handle = func(s Sentence, matches []*rules.RuleMatch, lang string) error {
		seen++
		return nil
	}
	checker := NewSentenceSourceChecker("en", func(text, lang string) []*rules.RuleMatch {
		return nil
	}, h)
	require.NoError(t, checker.Run(src))
	require.Equal(t, 2, seen)
	require.Equal(t, 2, h.SentenceCount)
}

func TestSentenceSourceChecker_MaxSentences(t *testing.T) {
	src := NewPlainTextSentenceSource(strings.NewReader(
		"First long enough sentence here.\nSecond long enough sentence here.\nThird long enough sentence here.\n"))
	h := NewResultHandler(1, 0)
	checker := NewSentenceSourceChecker("en", nil, h)
	err := checker.Run(src)
	require.Error(t, err)
	var dl DocumentLimitReachedError
	require.ErrorAs(t, err, &dl)
	require.Equal(t, 1, h.SentenceCount)
}
