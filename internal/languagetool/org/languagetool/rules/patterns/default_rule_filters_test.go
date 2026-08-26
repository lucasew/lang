package patterns

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/stretchr/testify/require"
)

func TestDateRangeCheckerFilterRegistered(t *testing.T) {
	require.True(t, GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.DateRangeChecker"))
	f := GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.DateRangeChecker")
	m := rules.NewRuleMatch(rules.NewFakeRule("DR"), nil, 0, 5, "range")
	// x >= y → keep
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"x": "2020", "y": "2010"}, 0, nil, nil))
	// x < y → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"x": "2010", "y": "2020"}, 0, nil, nil))
	// non-numeric → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"x": "a", "y": "1"}, 0, nil, nil))
}

func TestSetDefaultMultitokenSpeller_WiresIsMisspelledToken(t *testing.T) {
	sp := multitoken.NewMultitokenSpeller()
	miss := func(s string) bool { return s == "x" }
	SetDefaultMultitokenSpellerWithOptions(sp, miss, false)
	t.Cleanup(func() { SetDefaultMultitokenSpellerWithOptions(nil, nil, false) })
	// Java discardRunOnWords uses spelling rule on every language
	require.NotNil(t, sp.IsMisspelledToken)
	require.True(t, sp.IsMisspelledToken("x"))
	require.False(t, sp.IsMisspelledToken("ok"))
	// checkSpelling=false: IsMisspelledToken still set; filter short-code gate stays off
}

func TestMultitokenSpellerFilter_CapitalizesAtSentenceStart(t *testing.T) {
	sp := multitoken.NewMultitokenSpeller()
	require.NoError(t, sp.LoadWords(strings.NewReader("new york\n")))
	// fr: gate off; still capitalizes at sentence start
	SetDefaultMultitokenSpellerWithOptions(sp, nil, false)
	t.Cleanup(func() { SetDefaultMultitokenSpellerWithOptions(nil, nil, false) })

	sent := languagetool.AnalyzePlain("new yrok is big")
	text := sent.GetText()
	from := strings.Index(text, "new yrok")
	require.GreaterOrEqual(t, from, 0)
	to := from + len("new yrok")
	m := rules.NewRuleMatch(rules.NewFakeRule("MT"), sent, from, to, "multi")
	// Java pattern matcher sets OriginalErrorStr before the filter runs.
	m.OriginalErrorStr = "new yrok"
	f := GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter")
	// nil patternTokens: Go convenience (skip allIgnored gate); patternTokenPos=1 → sentence start
	out := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, "New york", out.GetSuggestedReplacements()[0])
}

func TestShortenedYearAndWhitespaceFiltersRegistered(t *testing.T) {
	require.True(t, GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.ShortenedYearRangeChecker"))
	require.True(t, GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.WhitespaceCheckFilter"))
	f := GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.ShortenedYearRangeChecker")
	m := rules.NewRuleMatch(rules.NewFakeRule("SY"), nil, 0, 5, "y")
	// 1998-92 → 1992; invalid range (x>=y) → keep
	require.NotNil(t, f.AcceptRuleMatch(m, map[string]string{"x": "1998", "y": "92"}, 0, nil, nil))
	// 1990-99 → 1999; valid range → drop
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"x": "1990", "y": "99"}, 0, nil, nil))
}


func TestDemoPartialPosTagFilterRegistered(t *testing.T) {
	require.True(t, GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.DemoPartialPosTagFilter"))
	f := GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.DemoPartialPosTagFilter")
	tok := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("inaccurate", nil, nil), 0)
	m := rules.NewRuleMatch(rules.NewFakeRule("D"), nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"no": "1", "regexp": "(?:in|un)(.*)", "postag_regexp": "JJ",
	}, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	require.NotNil(t, out)
}
