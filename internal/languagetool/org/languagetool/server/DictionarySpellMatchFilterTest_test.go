package server

// Twin of DictionarySpellMatchFilterTest — full EN/DE JLT pipeline deferred;
// getPhrases + accepted-phrase filter smokes via rules package.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestDictionarySpellMatchFilter_GetPhrases(t *testing.T) {
	text := "This is aa bb and then xx yyy zzzz"
	rule := &rules.DictFilterRule{ID: "SPELL"}
	matches := []*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, 8, 10, "fake msg"),
		rules.NewRuleMatch(rule, nil, 11, 13, "fake msg"),
		rules.NewRuleMatch(rule, nil, 23, 25, "fake msg"),
		rules.NewRuleMatch(rule, nil, 26, 29, "fake msg"),
		rules.NewRuleMatch(rule, nil, 30, 34, "fake msg"),
	}
	require.Equal(t, "aa", text[8:10])
	require.Equal(t, "bb", text[11:13])
	require.Equal(t, "xx", text[23:25])
	require.Equal(t, "yyy", text[26:29])
	require.Equal(t, "zzzz", text[30:34])

	f := rules.NewDictionarySpellMatchFilter(nil)
	result := f.GetPhrases(matches, text)
	require.Equal(t, 3, len(result))
	require.Contains(t, result, "aa bb")
	require.Contains(t, result, "xx yyy")
	require.Contains(t, result, "xx yyy zzzz")
}

func TestDictionarySpellMatchFilter_FilterMorfologik(t *testing.T) {
	f := rules.NewDictionarySpellMatchFilter([]string{"baad mistak"})
	text := "This is a baad mistak."
	rule := &rules.DictFilterRule{ID: "SPELL"}
	baadStart := strings.Index(text, "baad")
	mistakStart := strings.Index(text, "mistak")
	out := f.Filter([]*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, baadStart, baadStart+4, "miss"),
		rules.NewRuleMatch(rule, nil, mistakStart, mistakStart+6, "miss"),
	}, text)
	require.Empty(t, out)

	iso := "This is a mistak."
	ms := strings.Index(iso, "mistak")
	out = f.Filter([]*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, ms, ms+6, "m"),
	}, iso)
	require.Len(t, out, 1)
}

func TestDictionarySpellMatchFilter_PartialMatches(t *testing.T) {
	f := rules.NewDictionarySpellMatchFilter([]string{"baad mistake"})
	text := "This is a baad mistake."
	rule := &rules.DictFilterRule{ID: "SPELL"}
	baad := strings.Index(text, "baad")
	out := f.Filter([]*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, baad, baad+4, "m"),
	}, text)
	require.Empty(t, out)

	iso := "This is baad."
	b := strings.Index(iso, "baad")
	out = f.Filter([]*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, b, b+4, "m"),
	}, iso)
	require.Len(t, out, 1)
}

func TestDictionarySpellMatchFilter_FilterHunspell(t *testing.T) {
	f := rules.NewDictionarySpellMatchFilter([]string{"schlim Fehlar"})
	text := "Das ist ein schlim Fehlar."
	rule := &rules.DictFilterRule{ID: "SPELL"}
	s1 := strings.Index(text, "schlim")
	s2 := strings.Index(text, "Fehlar")
	out := f.Filter([]*rules.RuleMatch{
		rules.NewRuleMatch(rule, nil, s1, s1+6, "m"),
		rules.NewRuleMatch(rule, nil, s2, s2+6, "m"),
	}, text)
	require.Empty(t, out)
}
