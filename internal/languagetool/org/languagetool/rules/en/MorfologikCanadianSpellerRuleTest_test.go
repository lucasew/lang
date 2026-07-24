package en

// Twin of org.languagetool.rules.en.MorfologikCanadianSpellerRuleTest.
// Official resources: /en/hunspell/en_CA.dict, spelling_en-CA.txt, en/en-US-GB.txt
// (column 0), spelling.txt / multiwords — same as Java. No inject-only soft maps.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireCanadianDict skips when official en_CA.dict is not discoverable (no invent).
func requireCanadianDict(t *testing.T) *MorfologikVariantSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(CanadianSpellerDict) == "" {
		t.Skip("en_CA.dict not discoverable (third_party/english-pos-dict or inspiration)")
	}
	r := NewMorfologikCanadianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

func TestMorfologikCanadianSpellerRule_Meta(t *testing.T) {
	r := NewMorfologikCanadianSpellerRule()
	require.Equal(t, MorfologikCanadianSpellerRuleID, r.GetID())
	require.Equal(t, "MORFOLOGIK_RULE_EN_CA", r.GetID())
	require.Equal(t, CanadianSpellerDict, r.GetFileName())
	require.Equal(t, "/en/hunspell/en_CA.dict", r.GetFileName())
	require.Equal(t, CanadianVariantSpellingFile, r.GetLanguageVariantSpellingFileName())
	require.Equal(t, "en/hunspell/spelling_en-CA.txt", r.GetLanguageVariantSpellingFileName())
	// Java US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0) — American form → Canadian/British form
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
	// column 0, not 1: Canadian/British form is not a key for CA rule
	require.Nil(t, r.IsValidInOtherVariant("colour"))
	require.Nil(t, r.IsValidInOtherVariant("behaviour"))
}

// Twin of testSuggestions → AbstractEnglishSpellerRuleTest.testNonVariantSpecificSuggestions
func TestMorfologikCanadianSpellerRule_Suggestions(t *testing.T) {
	r := requireCanadianDict(t)
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
}

// Twin of MorfologikCanadianSpellerRuleTest.testMorfologikSpeller.
//
// Java king note: the Java test constructs MorfologikBritishSpellerRule by mistake
// (with CanadianEnglish language) for these asserts. We do not "fix" that typo
// in production code; for this sector we exercise MorfologikCanadianSpellerRule +
// official en_CA.dict with the same surface expectations Java asserts
// (behaviour accepted; arbor suggests arbour; aõh / a).
// Note: "arbor" is NOT in en-US-GB.txt (only arborize* etc.); Java only asserts
// contains("arbour") as a dictionary suggestion, not a variant-message path.
func TestMorfologikCanadianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireCanadianDict(t)

	// correct sentences — Java JLanguageTool(CanadianEnglish).getAnalyzedSentence
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"He doesn't know what to do.",
		",",
		"123454",
		"I like my emoji (😥)...",
		"μ",
	} {
		ms, err := r.Match(analyzeENFull(s))
		require.NoError(t, err, s)
		require.Empty(t, ms, "good %q", s)
	}

	// incorrect: arbor → arbour (dict suggestion; Java assertTrue contains "arbour")
	matches1, err := r.Match(analyzeEN("arbor"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 0, matches1[0].GetFromPos())
	require.Equal(t, 5, matches1[0].GetToPos())
	require.Contains(t, matches1[0].GetSuggestedReplacements(), "arbour")

	require.Equal(t, 1, matchCount(t, r, "aõh"))
	require.Equal(t, 0, matchCount(t, r, "a"))
}
