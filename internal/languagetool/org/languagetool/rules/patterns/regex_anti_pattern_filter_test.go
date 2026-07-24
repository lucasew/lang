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

// Java Matcher.start/end are UTF-16; multi-byte prefix must not desync antipattern overlap.
func TestRegexAntiPatternFilter_UTF16(t *testing.T) {
	f := RegexAntiPatternFilter{}
	// "café world" — é is 1 UTF-16 unit / 2 UTF-8 bytes.
	// "world" regex at byte 6; UTF-16 index 5.
	// Match on the space (UTF-16 4..5): anti "world" partially overlaps ToPos==5 in UTF-16.
	// Byte-only compare would miss (anti start 6 > ToPos 5) and wrongly keep the match.
	text := "café world"
	sent := languagetool.AnalyzePlain(text)
	m := rules.NewRuleMatch(rules.NewFakeRule("R"), sent, 4, 5, "msg")
	require.Nil(t, f.AcceptRegexMatch(m, map[string]string{"antipatterns": "world"}, sent),
		"UTF-16 antipattern edge overlap must drop match")
	require.NotNil(t, f.AcceptRegexMatch(m, map[string]string{"antipatterns": "xyz"}, sent))
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
