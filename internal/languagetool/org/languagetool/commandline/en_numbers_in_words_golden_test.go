package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_EN_NumbersInWords(t *testing.T) {
	text := "of America1s real religion."
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents)
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok == nil || tok.GetToken() != "America1s" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				require.NotEqual(t, "JJ", *r.GetPOSTag(), "OF_VBN_JJ must not invent JJ on untagged typo")
			}
		}
	}
	found := false
	for _, m := range lt.Check(text) {
		if m.RuleID == "NUMBERS_IN_WORDS" {
			found = true
		}
	}
	require.True(t, found, "NUMBERS_IN_WORDS should flag America1s")
}
