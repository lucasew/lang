package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestRegexAntiPatternFilter(t *testing.T) {
	f := RegexAntiPatternFilter{}
	sent := languagetool.AnalyzePlain("hello world")
	// match covering "world" at some positions - use from/to from token
	tokens := sent.GetTokensWithoutWhitespace()
	var from, to int
	for _, tok := range tokens {
		if tok.GetToken() == "world" {
			from, to = tok.GetStartPos(), tok.GetEndPos()
			break
		}
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), sent, from, to, "msg")
	// antipattern matching whole sentence start
	require.NotNil(t, f.AcceptRegexMatch(m, map[string]string{"antipatterns": "xyz"}, sent))
	// antipattern that overlaps "world"
	require.Nil(t, f.AcceptRegexMatch(m, map[string]string{"antipatterns": "world"}, sent))
}

func TestApostropheTypeFilter(t *testing.T) {
	f := ApostropheTypeFilter{}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("l'eau", nil, nil))
	tok.SetTypographicApostrophe(true)
	tok.SetStartPos(0)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), nil, 0, 5, "msg")
	args := map[string]string{"wordFrom": "1", "hasTypographicalApostrophe": "true"}
	require.NotNil(t, f.AcceptRuleMatch(m, args, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
	args["hasTypographicalApostrophe"] = "false"
	require.Nil(t, f.AcceptRuleMatch(m, args, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil))
}
