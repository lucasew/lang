package en

// Twin of EnglishRemoteRuleSuppressMisspelledTest — inject IsMisspelled (full EN speller deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

const testRemoteRuleID = "TEST_REMOTE_RULE"

func withOpts(pairs ...string) *rules.RemoteRuleConfig {
	c := rules.NewRemoteRuleConfig()
	c.RuleID = testRemoteRuleID
	for i := 0; i+1 < len(pairs); i += 2 {
		c.Options[pairs[i]] = pairs[i+1]
	}
	return c
}

// fake remote: one match spanning whole sentence with two suggestions.
func testRemoteMatches(cfg *rules.RemoteRuleConfig, sent *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	r := rules.NewRemoteRule("en-US", cfg)
	r.Execute = func(ss []*languagetool.AnalyzedSentence) *rules.RemoteRuleResult {
		var ms []*rules.RuleMatch
		for _, s := range ss {
			m := rules.NewRuleMatch(r, s, 0, len(s.GetText()), "Test match")
			m.SetSuggestedReplacements([]string{"mistake", "mistak"})
			ms = append(ms, m)
		}
		return rules.NewRemoteRuleResult(true, true, true, ms, ss)
	}
	raw := r.MatchRemote([]*languagetool.AnalyzedSentence{sent})
	// inject dict: "mistake" OK, "mistak" misspelled
	isMiss := func(w string) bool { return w != "mistake" }
	return r.SuppressMisspelled(raw, isMiss)
}

// Port of EnglishRemoteRuleSuppressMisspelledTest.test
func TestEnglishRemoteRuleSuppressMisspelled_Test(t *testing.T) {
	sent := languagetool.AnalyzePlain("This is a test sentence.")

	// no suppression
	m := testRemoteMatches(withOpts(), sent)
	require.Len(t, m, 1)
	require.Len(t, m[0].GetSuggestedReplacements(), 2)

	// suppressMisspelledMatch: any misspelled suggestion → drop match
	m = testRemoteMatches(withOpts("suppressMisspelledMatch", testRemoteRuleID), sent)
	require.Empty(t, m)

	// suppressMisspelledSuggestions: keep only correctly spelled
	m = testRemoteMatches(withOpts("suppressMisspelledSuggestions", testRemoteRuleID), sent)
	require.Len(t, m, 1)
	require.Equal(t, []string{"mistake"}, m[0].GetSuggestedReplacements())

	// regex full match
	m = testRemoteMatches(withOpts("suppressMisspelledMatch", ".*REMOTE.*"), sent)
	require.Empty(t, m)

	// regex partial (Go MatchString is full match like Java Matcher.matches)
	m = testRemoteMatches(withOpts("suppressMisspelledMatch", ".*REMOTE"), sent)
	require.Len(t, m, 1)
}
