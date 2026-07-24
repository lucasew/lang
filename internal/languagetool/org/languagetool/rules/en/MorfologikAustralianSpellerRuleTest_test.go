package en

// Twin of org.languagetool.rules.en.MorfologikAustralianSpellerRuleTest.
// Official resources: /en/hunspell/en_AU.dict, spelling_en-AU.txt, en/en-US-GB.txt
// (column 0), spelling.txt / multiwords — same as Java. No inject-only soft maps.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireAustralianDict skips when official en_AU.dict is not discoverable (no invent).
func requireAustralianDict(t *testing.T) *MorfologikVariantSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(AustralianSpellerDict) == "" {
		t.Skip("en_AU.dict not discoverable (third_party/english-pos-dict or inspiration)")
	}
	r := NewMorfologikAustralianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

func TestMorfologikAustralianSpellerRule_Meta(t *testing.T) {
	r := NewMorfologikAustralianSpellerRule()
	require.Equal(t, MorfologikAustralianSpellerRuleID, r.GetID())
	require.Equal(t, "MORFOLOGIK_RULE_EN_AU", r.GetID())
	require.Equal(t, AustralianSpellerDict, r.GetFileName())
	require.Equal(t, "/en/hunspell/en_AU.dict", r.GetFileName())
	require.Equal(t, AustralianVariantSpellingFile, r.GetLanguageVariantSpellingFileName())
	require.Equal(t, "en/hunspell/spelling_en-AU.txt", r.GetLanguageVariantSpellingFileName())
	// Java US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0) — American form → Australian/British form
	if len(r.OtherVariant) == 0 {
		t.Skip("en/en-US-GB.txt not loaded")
	}
	vi := r.IsValidInOtherVariant("color")
	require.NotNil(t, vi)
	require.Equal(t, "American English", vi.GetVariantName())
	require.Equal(t, "colour", vi.GetOtherVariant())
	vi2 := r.IsValidInOtherVariant("behavior")
	require.NotNil(t, vi2)
	require.Equal(t, "behaviour", vi2.GetOtherVariant())
	// column 0, not 1: Australian/British form is not a key for AU rule
	require.Nil(t, r.IsValidInOtherVariant("colour"))
	require.Nil(t, r.IsValidInOtherVariant("behaviour"))
}

// Twin of testSuggestions → AbstractEnglishSpellerRuleTest.testNonVariantSpecificSuggestions
func TestMorfologikAustralianSpellerRule_Suggestions(t *testing.T) {
	r := requireAustralianDict(t)
	// assertFirstMatch ordered prefixes (Java assertFirstMatch)
	assertFirstSuggestions(t, r, "teh", "the")
	assertFirstSuggestions(t, r, "seperate", "separate")
	assertFirstSuggestions(t, r, "definately", "definitely")
	assertFirstSuggestions(t, r, "recieve", "receive")
	assertFirstSuggestions(t, r, "offical", "official")
	assertFirstSuggestions(t, r, "managment", "management")
	assertFirstSuggestions(t, r, "Febuary", "February")
	assertFirstSuggestions(t, r, "enviroment", "environment")
	assertFirstSuggestions(t, r, "occurence", "occurrence")
	assertFirstSuggestions(t, r, "commision", "commission")
	assertFirstSuggestions(t, r, "assocation", "association")
	assertFirstSuggestions(t, r, "Cincinatti", "Cincinnati")
	assertFirstSuggestions(t, r, "milennium", "millennium")
	assertFirstSuggestions(t, r, "accomodation", "accommodation")
	assertFirstSuggestions(t, r, "foriegn", "foreign")
	assertFirstSuggestions(t, r, "chemcial", "chemical")
	assertFirstSuggestions(t, r, "developement", "development")
	assertFirstSuggestions(t, r, "maintainance", "maintenance")
	assertFirstSuggestions(t, r, "restaraunt", "restaurant")
	assertFirstSuggestions(t, r, "garentee", "guarantee")
	assertFirstSuggestions(t, r, "greatful", "grateful")
	assertFirstSuggestions(t, r, "hipocrit", "hypocrite")
	assertFirstSuggestions(t, r, "mischevious", "mischievous")
	assertFirstSuggestions(t, r, "hygeine", "hygiene")
	assertFirstSuggestions(t, r, "ur", "your")
	assertFirstSuggestions(t, r, "doublecheck", "double-check")
	assertFirstSuggestions(t, r, "alot", "a lot")
	assertFirstSuggestions(t, r, "speach", "speech")
	assertFirstSuggestions(t, r, "qualifys", "qualifies")
	assertFirstSuggestions(t, r, "nicefys", "nicely")
	// assertAllMatches: expected are a prefix of suggestions in order
	assertFirstSuggestions(t, r, "campaignt", "campaign", "campaigns")
	assertFirstSuggestions(t, r, "campaignd", "campaign", "campaigns", "campaigned")
	assertFirstSuggestions(t, r, "campaignll", "campaign", "campaigns", "campaigned", "campaigner")
	assertFirstSuggestions(t, r, "spreaded", "spread", "spreader")
	assertFirstSuggestions(t, r, "slimiar", "similar", "slimier")
	assertFirstSuggestions(t, r, "vacume", "value", "volume", "acute", "vacuum")
	// contractions sentence
	ms, err := r.Match(analyzeEN("You couldn't; he didn't; it doesn't; they aren't; I hadn't; etc."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// Java MorfologikAustralianSpellerRuleTest: language-specific spelling_en-AU.txt assert is commented out
	// (AUTestWordToBeIgnore → AUTestWordToBeIgnored) — not exercised.
}

// Twin of MorfologikAustralianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikAustralianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireAustralianDict(t)

	// correct sentences — Java JLanguageTool(AustralianEnglish).getAnalyzedSentence
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"My fiancée", // check encoding isn't broken, spelling_merged.txt
		"He doesn't know what to do.",
		",",
		"123454",
		"I like my emoji 😍!",
		"μ",
		// Australian dict:
		"Dinkum",
	} {
		ms, err := r.Match(analyzeENFull(s))
		require.NoError(t, err, s)
		require.Empty(t, ms, "good %q", s)
	}

	// American "behavior" → Australian "behaviour" via variant map + positions
	matches1, err := r.Match(analyzeEN("behavior"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 0, matches1[0].GetFromPos())
	require.Equal(t, 8, matches1[0].GetToPos())
	require.Equal(t, "behaviour", matches1[0].GetSuggestedReplacements()[0])
	require.Contains(t, matches1[0].GetMessage(), "American English")

	require.Equal(t, 1, matchCount(t, r, "aõh"))
	require.Equal(t, 0, matchCount(t, r, "a"))

	// replacement pairs / irregular (teached → taught needs synthesizer)
	matches2, err := r.Match(analyzeEN("He teached us."))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 3, matches2[0].GetFromPos())
	require.Equal(t, 10, matches2[0].GetToPos())
	require.Equal(t, "taught", matches2[0].GetSuggestedReplacements()[0])
}
