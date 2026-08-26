package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin helpers for WhitespaceCheckFilter.acceptRuleMatch (Java-visible outcomes only).

func atrWithWS(token, whitespaceBefore string) *languagetool.AnalyzedTokenReadings {
	tok := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, nil, nil), 0)
	if whitespaceBefore != "" {
		tok.SetWhitespaceBeforeToken(whitespaceBefore)
	}
	return tok
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_keepWhenWhitespaceDiffers(t *testing.T) {
	// Java: !getWhitespaceBefore().equals(wsChar) → return match
	f := NewWhitespaceCheckFilter()
	tok1 := atrWithWS("a", "")
	tok2 := atrWithWS("b", "\t")
	m := NewRuleMatch(nil, nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "2"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok1, tok2}, nil)
	require.Same(t, m, out)
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_nullWhenEqual(t *testing.T) {
	// Java: equals → return null
	f := NewWhitespaceCheckFilter()
	tok1 := atrWithWS("a", "")
	tok2 := atrWithWS("b", " ")
	m := NewRuleMatch(nil, nil, 0, 2, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "2"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok1, tok2}, nil)
	require.Nil(t, out)
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_emptyVsSpaceTabNBSP(t *testing.T) {
	f := NewWhitespaceCheckFilter()
	m := NewRuleMatch(nil, nil, 0, 1, "msg")

	// default whitespaceBefore is "" (Java constructor); expected " " → keep
	empty := atrWithWS("x", "")
	require.Equal(t, "", empty.GetWhitespaceBefore())
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{empty}, nil))

	// expected empty, got empty → suppress
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": "", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{empty}, nil))

	// space equal → suppress; unequal tab → keep
	sp := atrWithWS("x", " ")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{sp}, nil))
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": "\t", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{sp}, nil))

	// tab equal → suppress
	tab := atrWithWS("x", "\t")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": "\t", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tab}, nil))
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tab}, nil))

	// NBSP (U+00A0) — used in DE grammar.xml WhitespaceCheckFilter args
	nbsp := atrWithWS("x", "\u00A0")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": "\u00A0", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{nbsp}, nil))
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
		[]*languagetool.AnalyzedTokenReadings{nbsp}, nil))
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_badPosition(t *testing.T) {
	// Java: IllegalArgumentException "Wrong position in WhitespaceCheckFilter: {pos}, must be 1 to {len}"
	f := NewWhitespaceCheckFilter()
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	tokens := []*languagetool.AnalyzedTokenReadings{atrWithWS("a", " "), atrWithWS("b", " ")}

	require.PanicsWithValue(t, "Wrong position in WhitespaceCheckFilter: 0, must be 1 to 2", func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "0"}, 0, tokens, nil)
	})
	require.PanicsWithValue(t, "Wrong position in WhitespaceCheckFilter: 3, must be 1 to 2", func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "3"}, 0, tokens, nil)
	})
	// empty patternTokens: pos 1 is out of range (must be 1 to 0)
	require.PanicsWithValue(t, "Wrong position in WhitespaceCheckFilter: 1, must be 1 to 0", func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
			[]*languagetool.AnalyzedTokenReadings{}, nil)
	})
	// boundary: pos == length is valid (1-based last)
	out := f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "2"}, 0, tokens, nil)
	require.Nil(t, out)
	out = f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0, tokens, nil)
	require.Nil(t, out)
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_missingRequiredArgs(t *testing.T) {
	// Java RuleFilter.getRequired → IllegalArgumentException "Missing key '…'"
	f := NewWhitespaceCheckFilter()
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	tokens := []*languagetool.AnalyzedTokenReadings{atrWithWS("a", " ")}

	require.PanicsWithValue(t, "Missing key 'whitespaceChar'", func() {
		f.AcceptRuleMatch(m, map[string]string{"position": "1"}, 0, tokens, nil)
	})
	require.PanicsWithValue(t, "Missing key 'position'", func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " "}, 0, tokens, nil)
	})
	require.PanicsWithValue(t, "Missing key 'whitespaceChar'", func() {
		f.AcceptRuleMatch(m, map[string]string{}, 0, tokens, nil)
	})
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_nullPatternTokenNPE(t *testing.T) {
	// Java: patternTokens[pos-1].getWhitespaceBefore() NPE when element is null
	f := NewWhitespaceCheckFilter()
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "1"}, 0,
			[]*languagetool.AnalyzedTokenReadings{nil}, nil)
	})
}

func TestWhitespaceCheckFilter_AcceptRuleMatch_invalidPositionParse(t *testing.T) {
	// Java Integer.parseInt → NumberFormatException
	f := NewWhitespaceCheckFilter()
	m := NewRuleMatch(nil, nil, 0, 1, "msg")
	tokens := []*languagetool.AnalyzedTokenReadings{atrWithWS("a", " ")}
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "x"}, 0, tokens, nil)
	})
}
