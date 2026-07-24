package en

// Twin of org.languagetool.rules.en.MorfologikBritishSpellerRuleTest.
// Official resources: /en/hunspell/en_GB.dict, spelling_en-GB.txt, en/en-US-GB.txt
// (column 0), spelling.txt / multiwords — same as Java. No inject-only soft maps.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireBritishDict skips when official en_GB.dict is not discoverable (no invent).
func requireBritishDict(t *testing.T) *MorfologikVariantSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(BritishSpellerDict) == "" {
		t.Skip("en_GB.dict not discoverable (third_party/english-pos-dict or inspiration)")
	}
	r := NewMorfologikBritishSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

func TestMorfologikBritishSpellerRule_Meta(t *testing.T) {
	r := NewMorfologikBritishSpellerRule()
	require.Equal(t, MorfologikBritishSpellerRuleID, r.GetID())
	require.Equal(t, "MORFOLOGIK_RULE_EN_GB", r.GetID())
	require.Equal(t, BritishSpellerDict, r.GetFileName())
	require.Equal(t, "/en/hunspell/en_GB.dict", r.GetFileName())
	require.Equal(t, BritishVariantSpellingFile, r.GetLanguageVariantSpellingFileName())
	require.Equal(t, "en/hunspell/spelling_en-GB.txt", r.GetLanguageVariantSpellingFileName())
	// Java US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0) — American form → British form
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
	// column 0, not 1: British form is not a key for GB rule
	require.Nil(t, r.IsValidInOtherVariant("colour"))
}

// Twin of testSuggestions → AbstractEnglishSpellerRuleTest.testNonVariantSpecificSuggestions
func TestMorfologikBritishSpellerRule_Suggestions(t *testing.T) {
	r := requireBritishDict(t)
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

// Twin of MorfologikBritishSpellerRuleTest.testVariantMessages
func TestMorfologikBritishSpellerRule_VariantMessages(t *testing.T) {
	r := requireBritishDict(t)
	if len(r.OtherVariant) == 0 {
		t.Skip("en/en-US-GB.txt not loaded")
	}
	ms, err := r.Match(analyzeEN("This is a nice color."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "is American English")
	require.Equal(t, "colour", ms[0].GetSuggestedReplacements()[0])

	ms2, err := r.Match(analyzeEN("Color is the American English word."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetMessage(), "is American English")
	require.Equal(t, "Colour", ms2[0].GetSuggestedReplacements()[0])
}

// Twin of MorfologikBritishSpellerRuleTest.testMorfologikSpeller
func TestMorfologikBritishSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireBritishDict(t)

	// correct sentences — Java JLanguageTool(en-GB).getAnalyzedSentence
	for _, s := range []string{
		"This is an example: we get behaviour as a dictionary word.",
		"Why don't we speak today.",
		"He doesn't know what to do.",
		"The entrée at the café.",
		"This is my Ph.D. thesis.",
		",",
		"123454",
		"μ",
		// multiwords
		"Ménage à trois",
		"ménage à trois",
		"The quid pro quo",
		// apostrophes (EnglishTagger ’→' before speller)
		"Ma'am, O'Connell, O’Connell, O'Connor, O’Neill",
	} {
		ms, err := r.Match(analyzeENFull(s))
		require.NoError(t, err, s)
		require.Empty(t, ms, "good %q", s)
	}

	// Behavior (American) → Behaviour (British) via variant map + positions
	matches1, err := r.Match(analyzeEN("Behavior"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 0, matches1[0].GetFromPos())
	require.Equal(t, 8, matches1[0].GetToPos())
	require.Equal(t, "Behaviour", matches1[0].GetSuggestedReplacements()[0])
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

	// wrong split: I'm g oing → going, g going
	matches3, err := r.Match(analyzeEN("I'm g oing"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches3))
	require.Equal(t, "going", matches3[0].GetSuggestedReplacements()[0])
	require.Equal(t, "g going", matches3[0].GetSuggestedReplacements()[1])
	require.Equal(t, 4, matches3[0].GetFromPos())
	require.Equal(t, 10, matches3[0].GetToPos())

	// custom URLs via isValidInOtherVariant + enVariantBlogURL
	matches4, err := r.Match(analyzeEN("archeological"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches4))
	require.Equal(t,
		"https://languagetool.org/insights/post/our-or/#likeable-vs-likable-judgement-vs-judgment-oestrogen-vs-estrogen",
		matches4[0].GetURL())
}
