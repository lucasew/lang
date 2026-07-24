package en

// Twin of AbstractEnglishSpellerRuleTest (Java helper; non-variant suggestion tables).
// Prefer official en_US.dict via MorfologikAmericanSpellerRule (same as Java callers).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of AbstractEnglishSpellerRuleTest surface meta + additional spelling file names.
func TestAbstractEnglishSpellerRule_NoTests(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("colour")
	sp.AddWord("color")
	sp.Suggestions["collor"] = []string{"color", "colour"}
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", sp)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", r.GetID())
	require.Equal(t, "en", r.LanguageShortCode)
	require.Equal(t, "en-US", r.VariantCode)
	// Java getAdditionalSpellingFileNames: custom + global + multiwords (not spelling.txt)
	require.Contains(t, r.GetAdditionalSpellingFileNames(), "en/hunspell/spelling_custom.txt")
	require.Contains(t, r.GetAdditionalSpellingFileNames(), "spelling_global.txt")
	require.Contains(t, r.GetAdditionalSpellingFileNames(), "/en/multiwords.txt")
	require.True(t, IsDoNotSuggest("bullshit"))
	require.False(t, IsDoNotSuggest("hello"))
	require.Equal(t, []string{"color"}, FilterEnglishSuggestions([]string{"color", "bullshit"}))

	m, err := r.Match(languagetool.AnalyzePlain("color collor"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Contains(t, m[0].GetSuggestedReplacements(), "color")

	// Java language.getSynthesizer() wired when english_synth.dict present
	if openDiscoveredEnglishSynthesizer() != nil {
		require.NotNil(t, r.Synthesize, "AbstractEnglishSpellerRule must wire synthesizer when dict present")
	}
}

// Twin of AbstractEnglishSpellerRuleTest.testNonVariantSpecificSuggestions (official en_US.dict).
func TestAbstractEnglishSpellerRule_NonVariantSpecificSuggestions(t *testing.T) {
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not discoverable")
	}
	// Java callers pass MorfologikAmericanSpellerRule
	r := NewMorfologikAmericanSpellerRule()
	require.NotNil(t, r.Speller1)

	// assertFirstMatch ordered prefixes
	for _, tc := range []struct {
		bad  string
		want []string
	}{
		{"teh", []string{"the"}},
		{"seperate", []string{"separate"}},
		{"definately", []string{"definitely"}},
		{"recieve", []string{"receive"}},
		{"offical", []string{"official"}},
		{"managment", []string{"management"}},
		{"Febuary", []string{"February"}},
		{"enviroment", []string{"environment"}},
		{"occurence", []string{"occurrence"}},
		{"commision", []string{"commission"}},
		{"assocation", []string{"association"}},
		{"Cincinatti", []string{"Cincinnati"}},
		{"milennium", []string{"millennium"}},
		{"accomodation", []string{"accommodation"}},
		{"foriegn", []string{"foreign"}},
		{"chemcial", []string{"chemical"}},
		{"developement", []string{"development"}},
		{"maintainance", []string{"maintenance"}},
		{"restaraunt", []string{"restaurant"}},
		{"garentee", []string{"guarantee"}},
		{"greatful", []string{"grateful"}},
		{"hipocrit", []string{"hypocrite"}},
		{"mischevious", []string{"mischievous"}},
		{"hygeine", []string{"hygiene"}},
		{"ur", []string{"your"}},
		{"doublecheck", []string{"double-check"}},
		{"alot", []string{"a lot"}},
		{"speach", []string{"speech"}},
		{"qualifys", []string{"qualifies"}},
		{"nicefys", []string{"nicely"}},
		{"campaignt", []string{"campaign", "campaigns"}},
		{"campaignd", []string{"campaign", "campaigns", "campaigned"}},
		{"campaignll", []string{"campaign", "campaigns", "campaigned", "campaigner"}},
		{"spreaded", []string{"spread", "spreader"}},
		{"slimiar", []string{"similar", "slimier"}},
		{"vacume", []string{"value", "volume", "acute", "vacuum"}},
	} {
		assertFirstSuggestions(t, r, tc.bad, tc.want...)
	}

	ms, err := r.Match(analyzeEN("You couldn't; he didn't; it doesn't; they aren't; I hadn't; etc."))
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin of AbstractEnglishSpellerRuleTest.testHyphenatedWordSuggestions
func TestAbstractEnglishSpellerRule_HyphenatedWordSuggestions(t *testing.T) {
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not discoverable")
	}
	r := NewMorfologikAmericanSpellerRule()
	assertFirstSuggestions(t, r, "long-trem", "long-term")
	assertFirstSuggestions(t, r, "self-sefficient", "self-sufficient")
	assertFirstSuggestions(t, r, "parple-people-eater", "purple-people-eater")
}
