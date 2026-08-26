package rules

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.DictionaryMatchFilterTest.
// Uses ForbiddenWordsRule (test double from Java) instead of full JLanguageTool
// for matchesWithoutFilter/filter; spelling cases exercise DictionaryMatchFilter
// directly (speller not required for filter semantics).

type forbiddenWordsRule struct {
	words map[string]struct{}
}

func newForbiddenWordsRule(words ...string) *forbiddenWordsRule {
	m := map[string]struct{}{}
	for _, w := range words {
		m[w] = struct{}{}
	}
	return &forbiddenWordsRule{words: m}
}

func (r *forbiddenWordsRule) GetID() string { return "DictionaryMatchFilterTestRule" }

func (r *forbiddenWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var matches []*RuleMatch
	for _, token := range sentence.GetTokensWithoutWhitespace() {
		word := token.GetToken()
		if _, ok := r.words[word]; ok {
			matches = append(matches, NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), "Forbidden word: "+word))
		}
	}
	return matches
}

func checkForbidden(text string, forbidden []string) []*RuleMatch {
	rule := newForbiddenWordsRule(forbidden...)
	return rule.Match(languagetool.AnalyzePlain(text))
}

func isForbiddenWordMatch(word string, match *RuleMatch) bool {
	return match.Message == "Forbidden word: "+word
}

func TestDictionaryMatchFilter_MatchesWithoutFilter(t *testing.T) {
	matches := checkForbidden("This is fooxxx. Very bar of you! Even foobar, one might say.",
		[]string{"fooxxx", "bar", "foobar"})
	require.Len(t, matches, 3)
	require.True(t, isForbiddenWordMatch("fooxxx", matches[0]))
	require.True(t, isForbiddenWordMatch("bar", matches[1]))
	require.True(t, isForbiddenWordMatch("foobar", matches[2]))
}

func TestDictionaryMatchFilter_SpellingRuleMatches(t *testing.T) {
	// Without dictionary: a spelling-like match on "mistak" is kept.
	text := "This is a mistak"
	s := languagetool.AnalyzePlain(text)
	// find "mistak" span
	var from, to int
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "mistak" {
			from, to = tok.GetStartPos(), tok.GetEndPos()
			break
		}
	}
	spellMsg := "Possible spelling mistake found"
	rm := NewRuleMatch(NewFakeRule("MORFOLOGIK_RULE_EN_US"), s, from, to, spellMsg)
	require.True(t, strings.Contains(rm.Message, "Possible spelling mistake found"))
	require.Len(t, NewDictionaryMatchFilter(nil).Filter([]*RuleMatch{rm}, text), 1)

	// With "mistak" accepted: filtered out
	require.Len(t, NewDictionaryMatchFilter([]string{"mistak"}).Filter([]*RuleMatch{rm}, text), 0)

	// Different word still kept
	text2 := "This is another mistke."
	s2 := languagetool.AnalyzePlain(text2)
	for _, tok := range s2.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "mistke" {
			from, to = tok.GetStartPos(), tok.GetEndPos()
			break
		}
	}
	rm2 := NewRuleMatch(NewFakeRule("MORFOLOGIK_RULE_EN_US"), s2, from, to, spellMsg)
	require.Len(t, NewDictionaryMatchFilter([]string{"mistak"}).Filter([]*RuleMatch{rm2}, text2), 1)
}

func TestDictionaryMatchFilter_Filter(t *testing.T) {
	text := "This is foo. This is bar."
	matches := checkForbidden(text, []string{"foo", "bar"})
	filtered := NewDictionaryMatchFilter([]string{"foo"}).Filter(matches, text)
	require.Len(t, filtered, 1)
	require.True(t, isForbiddenWordMatch("bar", filtered[0]))
}
