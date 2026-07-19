package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWhitespaceCheckFilter(t *testing.T) {
	f := NewWhitespaceCheckFilter()
	// keep when whitespace before token != expected
	keep, err := f.Accept([]string{"", " "}, 2, " ")
	require.Empty(t, err)
	require.False(t, keep) // matches expected space → suppress
	keep, err = f.Accept([]string{"", " "}, 2, "\t")
	require.Empty(t, err)
	require.True(t, keep)
	_, err = f.Accept([]string{" "}, 2, " ")
	require.NotEmpty(t, err)
}

func TestWhitespaceCheckFilter_AcceptRuleMatch(t *testing.T) {
	f := NewWhitespaceCheckFilter()
	tok1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("a", nil, nil), 0)
	tok2 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("b", nil, nil), 1)
	tok2.SetWhitespaceBeforeToken("\u00A0") // nbsp
	m := NewRuleMatch(nil, nil, 0, 2, "msg")
	// expected regular space, got nbsp → keep
	out := f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": " ", "position": "2"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok1, tok2}, nil)
	require.NotNil(t, out)
	// expected nbsp → drop
	out = f.AcceptRuleMatch(m, map[string]string{"whitespaceChar": "\u00A0", "position": "2"}, 0,
		[]*languagetool.AnalyzedTokenReadings{tok1, tok2}, nil)
	require.Nil(t, out)
}

func TestWhitespaceBeforeChar_Stored(t *testing.T) {
	tok := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("x", nil, nil), 0)
	require.Equal(t, "", tok.GetWhitespaceBefore())
	tok.SetWhitespaceBeforeToken("\u2006")
	require.True(t, tok.IsWhitespaceBefore())
	require.Equal(t, "\u2006", tok.GetWhitespaceBefore())
}
