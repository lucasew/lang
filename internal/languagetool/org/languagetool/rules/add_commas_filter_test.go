package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAddCommasFilter_AfterOnly(t *testing.T) {
	f := NewAddCommasFilter()
	ctx := CommaContext{
		MatchedText:          "aun así",
		FirstToken:           "aun",
		LastToken:            "así",
		TokenBefore:          ",",
		TokenAfter:           "sigue",
		MatchAtSentenceStart: false,
	}
	require.True(t, f.Accept(ctx))
	require.Equal(t, []string{"así,"}, f.Suggest(ctx))
}

func TestAddCommasFilter_BeforeOnly(t *testing.T) {
	f := NewAddCommasFilter()
	ctx := CommaContext{
		MatchedText: "aun así",
		FirstToken:  "aun",
		LastToken:   "así",
		TokenBefore: "y",
		TokenAfter:  ".",
	}
	require.Equal(t, []string{", aun"}, f.Suggest(ctx))
}

func TestAddCommasFilter_Both(t *testing.T) {
	f := NewAddCommasFilter()
	ctx := CommaContext{
		MatchedText: "aun así",
		FirstToken:  "aun",
		LastToken:   "así",
		TokenBefore: "y",
		TokenAfter:  "sigue",
	}
	require.Equal(t, []string{", aun así,"}, f.Suggest(ctx))
}

func TestAddCommasFilter_Semicolon(t *testing.T) {
	f := NewAddCommasFilter()
	ctx := CommaContext{
		MatchedText:      "aun así",
		FirstToken:       "aun",
		LastToken:        "así",
		TokenBefore:      ",",
		TokenAfter:       "sigue",
		SuggestSemicolon: true,
	}
	require.Equal(t, []string{"; aun así,", ", aun así,"}, f.Suggest(ctx))
}

func TestAddCommasFilter_SuppressWhenOK(t *testing.T) {
	f := NewAddCommasFilter()
	ctx := CommaContext{
		MatchedText:          "Aun así",
		FirstToken:           "Aun",
		LastToken:            "así",
		TokenBefore:          "",
		TokenAfter:           ".",
		MatchAtSentenceStart: true,
	}
	require.False(t, f.Accept(ctx))
	require.Nil(t, f.Suggest(ctx))
}

func TestAddCommasFilter_AcceptRuleMatch(t *testing.T) {
	// "y aun así sigue" — need commas around "aun así"
	sent := languagetool.AnalyzePlain("y aun así sigue.")
	// find "aun" span
	text := sent.GetText()
	from := strings.Index(text, "aun así")
	require.GreaterOrEqual(t, from, 0)
	to := from + len("aun así")
	m := NewRuleMatch(NewFakeRule("COMMA"), sent, from, to, "add commas")
	out := NewAddCommasFilter().AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.NotEmpty(t, out.GetSuggestedReplacements())
}
