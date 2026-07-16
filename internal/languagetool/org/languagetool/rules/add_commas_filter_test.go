package rules

import (
	"testing"

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
