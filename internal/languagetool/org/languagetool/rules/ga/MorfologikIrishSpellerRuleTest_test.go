package ga

// Twin of org.languagetool.rules.ga.MorfologikIrishSpellerRuleTest
// Official resource: /ga/hunspell/ga_IE.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / AddWord / fake Suggestions invent — real binary dict only.
//
// Java setup:
//   rule = new MorfologikIrishSpellerRule(TestTools.getMessages("ga"), Irish.getInstance(), null, emptyList());
//   lt   = new JLanguageTool(Irish.getInstance());
// Match positions are UTF-16 (Java RuleMatch fromPos/toPos).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireIrishSpellerDict skips when official ga_IE.dict is not discoverable (no invent).
func requireIrishSpellerDict(t *testing.T) *MorfologikIrishSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(IrishSpellerDict) == "" {
		t.Skip("ga_IE.dict not discoverable (inspiration/.../ga/hunspell; Maven languagetool-ga-dicts)")
	}
	r := NewMorfologikIrishSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

func matchGA(t *testing.T, r *MorfologikIrishSpellerRule, text string) []*rules.RuleMatch {
	t.Helper()
	ms, err := r.Match(languagetool.AnalyzePlain(text))
	require.NoError(t, err)
	return ms
}

// Port of MorfologikIrishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikIrishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireIrishSpellerDict(t)
	require.Equal(t, MorfologikIrishSpellerRuleID, r.GetID())
	require.Equal(t, IrishSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_GA_IE", r.GetID())
	require.Equal(t, "/ga/hunspell/ga_IE.dict", r.GetFileName())
	// Java: super.ignoreWordsWithLength = 1
	require.Equal(t, 1, r.IgnoreWordsWithLength)
	// Java tokenizingPattern(): Pattern.compile("-")
	require.NotNil(t, r.TokenizingPattern)
	require.True(t, r.TokenizingPattern.MatchString("-"))

	// incorrect sentences:
	matches := matchGA(t, r, "botun")
	// check match positions:
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	sugs := matches[0].GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Equal(t, "botún", sugs[0])

	matches = matchGA(t, r, "amaċ")
	// check match positions:
	require.Len(t, matches, 1)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())
	sugs = matches[0].GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Equal(t, "amach", sugs[0])

	matches = matchGA(t, r, "😂 botun")
	// check match positions (UTF-16: emoji = 2 units + space → botun at 3..8):
	require.Len(t, matches, 1)
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 8, matches[0].GetToPos())
	sugs = matches[0].GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Equal(t, "botún", sugs[0])
}
