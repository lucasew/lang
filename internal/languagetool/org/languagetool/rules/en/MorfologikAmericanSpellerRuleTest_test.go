package en

// Twin of org.languagetool.rules.en.MorfologikAmericanSpellerRuleTest.
// Official resources: /en/hunspell/en_US.dict, spelling_en-US.txt, en/en-US-GB.txt,
// spelling.txt / multiwords / english_synth.dict (irregular forms) — same as Java.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	entag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

// requireAmericanDict skips when official en_US.dict is not discoverable (no invent).
func requireAmericanDict(t *testing.T) *MorfologikVariantSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not discoverable (third_party/english-pos-dict or inspiration)")
	}
	r := NewMorfologikAmericanSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// analyzeEN ports JLanguageTool(en-US).getAnalyzedSentence: EnglishWordTokenizer.
func analyzeEN(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, entok.NewEnglishWordTokenizer())
}

// analyzeENSoft ports getAnalyzedSentence with Language.getIgnoredCharactersRegex [\\u00AD].
func analyzeENSoft(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizerAndIgnore(text, entok.NewEnglishWordTokenizer(), languagetool.GermanIgnoredCharactersRegex)
}

// analyzeENFull ports JLanguageTool EN analysis used for apostrophe/clitic keep (tagger).
func analyzeENFull(text string) *languagetool.AnalyzedSentence {
	return entag.AnalyzeEnglishSentence(text)
}

func TestMorfologikAmericanSpellerRule_Meta(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	require.Equal(t, MorfologikAmericanSpellerRuleID, r.GetID())
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", r.GetID())
	require.Equal(t, AmericanSpellerDict, r.GetFileName())
	require.Equal(t, "/en/hunspell/en_US.dict", r.GetFileName())
	require.Equal(t, AmericanVariantSpellingFile, r.GetLanguageVariantSpellingFileName())
	require.Equal(t, "en/hunspell/spelling_en-US.txt", r.GetLanguageVariantSpellingFileName())
	// Java OtherVariant from en/en-US-GB.txt column 1 (British → American)
	if len(r.OtherVariant) == 0 {
		t.Skip("en/en-US-GB.txt not loaded")
	}
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())
	// Java MorfologikAmericanSpellerRule.getAdditionalTopSuggestions: automize*
	require.NotNil(t, r.GetAdditionalTopSuggestionsFn)
	require.Equal(t, []string{"automate"}, r.GetAdditionalTopSuggestionsFn(nil, "automize"))
	require.Equal(t, []string{"automated"}, r.GetAdditionalTopSuggestionsFn(nil, "automized"))
	require.Equal(t, []string{"automating"}, r.GetAdditionalTopSuggestionsFn(nil, "automizing"))
	require.Equal(t, []string{"automates"}, r.GetAdditionalTopSuggestionsFn(nil, "automizes"))
}

// Twin of testSuggestionForMisspelledHyphenatedWords
func TestMorfologikAmericanSpellerRule_SuggestionForMisspelledHyphenatedWords(t *testing.T) {
	r := requireAmericanDict(t)
	assertSuggestionContains(t, r, "one-diminensional", "one-dimensional")
	assertSuggestionContains(t, r, "parple-people-eater", "purple-people-eater")
}

// Twin of testNamedEntityIgnore — Java constructs rule with FakeLanguageModel; no assertions on matches.
func TestMorfologikAmericanSpellerRule_NamedEntityIgnore(t *testing.T) {
	r := requireAmericanDict(t)
	// Java: LanguageModel lm = new FakeLanguageModel(map); rule.match(... "Petr Smith.")
	// Smoke: Match does not panic without NER pipe (languageModel alone insufficient).
	r.LanguageModel = countMapLM(map[string]int{"Peter": 100, "Petr": 10})
	ms, err := r.Match(analyzeEN("This is a test with Petr Smith."))
	require.NoError(t, err)
	_ = ms
}

// Twin of testSuggestions → AbstractEnglishSpellerRuleTest.testNonVariantSpecificSuggestions
func TestMorfologikAmericanSpellerRule_Suggestions(t *testing.T) {
	r := requireAmericanDict(t)
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
	// hyphenated word suggestions (AbstractEnglishSpellerRuleTest.testHyphenatedWordSuggestions)
	assertFirstSuggestions(t, r, "long-trem", "long-term")
	assertFirstSuggestions(t, r, "self-sefficient", "self-sufficient")
	assertFirstSuggestions(t, r, "parple-people-eater", "purple-people-eater")
}

// Twin of testVariantMessages
func TestMorfologikAmericanSpellerRule_VariantMessages(t *testing.T) {
	r := requireAmericanDict(t)
	if len(r.OtherVariant) == 0 {
		t.Skip("en/en-US-GB.txt not loaded")
	}
	ms, err := r.Match(analyzeEN("This is a nice colour."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "is British English")
	require.Equal(t, "color", ms[0].GetSuggestedReplacements()[0])

	ms2, err := r.Match(analyzeEN("Colour is the British words."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetMessage(), "is British English")
	require.Equal(t, "Color", ms2[0].GetSuggestedReplacements()[0])
}

// Twin of testUserDict
func TestMorfologikAmericanSpellerRule_UserDict(t *testing.T) {
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not discoverable")
	}
	// Java UserConfig(Arrays.asList("mytestword", "mytesttwo")) — free account ignores accepted words
	uc := languagetool.NewUserConfigWithWords([]string{"mytestword", "mytesttwo"}, nil)
	r := NewMorfologikAmericanSpellerRuleWithUser(uc)
	ms, err := r.Match(analyzeEN("mytestword"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(analyzeEN("mytesttwo"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(analyzeEN("mytestthree"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
}

// Twin of testMorfologikSpeller
func TestMorfologikAmericanSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireAmericanDict(t)

	// spelling.txt merge + multiwords
	require.False(t, r.IsMisspelled("mansplaining"))
	// Qur'an: Java EnglishTagger / multiword path accepts; full EN analysis keeps orthography
	ms, err := r.Match(analyzeENFull("Qur'an"))
	require.NoError(t, err)
	require.Empty(t, ms, "Qur'an accepted with EN tagger/multiword path")

	// diacritic suggestion
	ms, err = r.Match(analyzeEN("fianc"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetSuggestedReplacements(), "fiancé")

	// correct sentences — Java JLanguageTool(en-US).getAnalyzedSentence (tokenizer+tagger)
	for _, s := range []string{
		"This is an example: we get behavior as a dictionary word.",
		"Why don't we speak today.",
		"An URL like http://sdaasdwe.com is no error.",
		"He doesn't know what to do.",
		",",
		"123454",
		"I like my emoji 😾",
		"μ",
		"I like my emoji ❤️",
		"This is English text 🗺.",
		"Yes ma'am.",
		"Yes ma’am.", // EnglishTagger ’→' before speller
		"'twas but a dream of thee",
		"fo'c'sle",
		"O'Connell, O’Connell, O'Connor, O’Neill",
		"viva voce, a fortiori, in vitro",
		`5¼" floppy disk drive`,
		"a visual magnitude of -2½",
		"Water freezes at 0º C. 175ºC",
		"33°5′40″N and 32°59′0″E.",
		"It's up to 5·10-³ meters.",
		"141°00′7.128″W",
		"1031－1095",
		"It is thus written 1″.",
		"a 30½-inch scale length.",
		"symbolically stated as A ∈ ℝ3.",
		"Thus ℵ0 is a regular cardinal.",
		"the classical space B(ℓ2)",
		"🏽",
		"🧡‍♂️ , 🎉💛✈️",
		"компьютерная",
		"中文維基百科 中文维基百科",
		"The statements¹ of⁷ the⁵⁰ government⁹‽",
		"At 3 o'clock.",
		"At 3 o’clock.",
		"fast⇿superfast",
		// multiwords.txt
		"C'est la vie.",
		"c’est la guerre!",
		"Andorra la Vella is the capital and largest city of Andorra.",
		"bona fides.",
		"doctor honoris causa",
	} {
		ms, err = r.Match(analyzeENFull(s))
		require.NoError(t, err, s)
		require.Empty(t, ms, "good %q", s)
	}

	// multiword / place-name suggestions (Java assertSuggestion contains)
	assertSuggestionContains(t, r, "Abu-Bakr", "Abu Bakr")
	assertSuggestionContains(t, r, "Abudhabi", "Abu Dhabi")
	assertSuggestionContains(t, r, "Casagrande", "Casa Grande")
	assertSuggestionContains(t, r, "ELPASO", "El Paso")
	assertSuggestionContains(t, r, "Eldorado", "El Dorado")
	assertSuggestionContains(t, r, "nom-de-plume", "nom de plume")
	assertSuggestionContains(t, r, "sui-generis", "sui generis")
	assertSuggestionContains(t, r, "Wiener-Neustadt", "Wiener Neustadt")
	assertSuggestionContains(t, r, "Zyklon-B", "Zyklon B")

	// British behaviour → American behavior (variant map + positions)
	matches1, err := r.Match(analyzeEN("behaviour"))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches1))
	require.Equal(t, 0, matches1[0].GetFromPos())
	require.Equal(t, 9, matches1[0].GetToPos())
	require.Equal(t, "behavior", matches1[0].GetSuggestedReplacements()[0])
	require.Contains(t, matches1[0].GetMessage(), "British English")

	require.Equal(t, 1, matchCount(t, r, "aõh"))
	require.Equal(t, 0, matchCount(t, r, "a"))

	// replacement pairs / irregular (teached → taught needs synthesizer)
	matches2, err := r.Match(analyzeEN("He teached us."))
	require.NoError(t, err)
	require.Equal(t, 1, len(matches2))
	require.Equal(t, 3, matches2[0].GetFromPos())
	require.Equal(t, 10, matches2[0].GetToPos())
	require.Equal(t, "taught", matches2[0].GetSuggestedReplacements()[0])

	// hyphens: accept when all parts ok
	require.Equal(t, 0, matchCount(t, r, "A web-based software."))
	require.Equal(t, 1, matchCount(t, r, "A wxeb-based software."))
	require.Equal(t, 1, matchCount(t, r, "A web-baxsed software."))
	require.Equal(t, 0, matchCount(t, r, "A web-feature-driven-car software."))
	require.Equal(t, 1, matchCount(t, r, "A web-feature-drivenx-car software."))

	// assertAllMatches robinson
	assertAllSuggestions(t, r, "robinson",
		"Robinson", "robin son", "robins on", "Robson", "Robeson", "robins", "Roberson", "Robins")

	// contractions with apostrophe (need tagger isTagged for foolin' keep)
	for _, s := range []string{
		"You're only foolin' round.",
		"You’re only foolin’ round.",
		"This is freakin' hilarious.",
		"It's the meal that keeps on givin'.",
		"Don't Stop Believin'.",
	} {
		ms, err = r.Match(analyzeENFull(s))
		require.NoError(t, err, s)
		require.Empty(t, ms, "good %q", s)
	}
	require.Equal(t, 1, matchCount(t, r, "wrongwordin'"))
	require.Equal(t, 1, matchCount(t, r, "wrongwordin’"))
}

// Twin of testIgnoredChars
func TestMorfologikAmericanSpellerRule_IgnoredChars(t *testing.T) {
	r := requireAmericanDict(t)
	require.Equal(t, 0, matchCountSoft(t, r, "software"))
	require.Equal(t, 0, matchCountSoft(t, r, "soft\u00ADware"))
	require.Equal(t, 0, matchCountSoft(t, r, "A software"))
	require.Equal(t, 0, matchCountSoft(t, r, "A soft\u00ADware"))

	// Java lt.check("sux\u00AD tainability") → 2 matches; second [5,16]
	// rule.Match with soft-hyphen metadata: first span is token endPos [0,4]
	// (annotated-text remaps first to [0,3] at JLanguageTool level — not rule layer).
	ms, err := r.Match(analyzeENSoft("sux\u00AD tainability"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 4, ms[0].GetToPos()) // surface includes U+00AD
	require.Equal(t, 5, ms[1].GetFromPos())
	require.Equal(t, 16, ms[1].GetToPos())

	// Java issue #1769: "The sus\u00AD tainability" → one wrong-split match [4,20]
	ms, err = r.Match(analyzeENSoft("The sus\u00AD tainability"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 20, ms[0].GetToPos())
	require.Equal(t, "sustainability", ms[0].GetSuggestedReplacements()[0])
}

// Twin of testRuleWithWrongSplit
func TestMorfologikAmericanSpellerRule_RuleWithWrongSplit(t *testing.T) {
	r := requireAmericanDict(t)

	ms, err := r.Match(analyzeEN("But than kyou for the feedback"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "thank you", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 13, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("But thanky ou for the feedback"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "thank you", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 13, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("But thank you for th efeedback"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "the feedback", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 18, ms[0].GetFromPos())
	require.Equal(t, 30, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("But thank you for thef eedback"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "the feedback", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 18, ms[0].GetFromPos())
	require.Equal(t, 30, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("But thnk you fo rthe feedback"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))
	require.Equal(t, "tank", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 1, len(ms[1].GetSuggestedReplacements()))
	require.Equal(t, "for the", ms[1].GetSuggestedReplacements()[0])
	require.Equal(t, 13, ms[1].GetFromPos())
	require.Equal(t, 20, ms[1].GetToPos())

	ms, err = r.Match(analyzeEN("LanguageTol offer sspell checking"))
	require.NoError(t, err)
	require.Equal(t, 2, len(ms))
	require.Equal(t, "LanguageTool", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, "offers spell", ms[1].GetSuggestedReplacements()[0])
	require.Equal(t, 12, ms[1].GetFromPos())
	require.Equal(t, 24, ms[1].GetToPos())

	ms, err = r.Match(analyzeEN("I'm g oing"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "going", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 10, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("I'm go ing"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "going", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 10, ms[0].GetToPos())

	ms, err = r.Match(analyzeEN("to thow"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 7, ms[0].GetToPos())
	require.Equal(t, "show", ms[0].GetSuggestedReplacements()[0])

	ms, err = r.Match(analyzeEN("dditionally, this is it."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Additionally", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 11, ms[0].GetToPos())
}

// Twin of testSuggestionForIrregularWords
func TestMorfologikAmericanSpellerRule_SuggestionForIrregularWords(t *testing.T) {
	r := requireAmericanDict(t)
	if r.Synthesize == nil {
		t.Skip("english_synth.dict not discoverable — irregular forms need synthesizer")
	}
	// verbs
	assertSuggestionContains(t, r, "He teached us.", "taught")
	assertSuggestionContains(t, r, "He buyed the wrong brand", "bought")
	assertSuggestionContains(t, r, "I thinked so.", "thought")
	assertSuggestionContains(t, r, "She becomed", "became")
	assertSuggestionContains(t, r, "It begined", "began")
	assertSuggestionContains(t, r, "It bited", "bit")
	assertSuggestionContains(t, r, "She dealed", "dealt")
	assertSuggestionContains(t, r, "She drived", "drove")
	assertSuggestionContains(t, r, "He drawed", "drew")
	assertSuggestionContains(t, r, "She finded", "found")
	assertSuggestionContains(t, r, "It hurted", "hurt")
	assertSuggestionContains(t, r, "It was keeped", "kept")
	assertSuggestionContains(t, r, "He maked", "made")
	assertSuggestionContains(t, r, "She runed", "ran")
	assertSuggestionContains(t, r, "She selled", "sold")
	assertSuggestionContains(t, r, "He speaked", "spoke")
	// nouns
	assertSuggestionContains(t, r, "auditory stimuluses", "stimuli")
	assertSuggestionContains(t, r, "analysises", "analyses")
	assertSuggestionContains(t, r, "parenthesises", "parentheses")
	assertSuggestionContains(t, r, "childs", "children")
	assertSuggestionContains(t, r, "womans", "women")
	// adjectives comparative / superlative
	assertSuggestionContains(t, r, "gooder", "better")
	assertSuggestionContains(t, r, "bader", "worse")
	assertSuggestionContains(t, r, "farer", "further", "farther")
	assertSuggestionContains(t, r, "goodest", "best")
	assertSuggestionContains(t, r, "badest", "worst")
	assertSuggestionContains(t, r, "farest", "furthest", "farthest")
	// message form for past tense
	ms, err := r.Match(analyzeEN("He teached us."))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetMessage(), "past tense")
	require.Equal(t, "taught", ms[0].GetSuggestedReplacements()[0])
}

// Twin of testIsMisspelled
func TestMorfologikAmericanSpellerRule_IsMisspelled(t *testing.T) {
	r := requireAmericanDict(t)
	require.True(t, r.IsMisspelled("sdadsadas"))
	require.True(t, r.IsMisspelled("bicylce"))
	require.True(t, r.IsMisspelled("tabble"))
	require.True(t, r.IsMisspelled("tabbles"))
	require.False(t, r.IsMisspelled("bicycle"))
	require.False(t, r.IsMisspelled("table"))
	require.False(t, r.IsMisspelled("tables"))
}

// Twin of testGetOnlySuggestions
func TestMorfologikAmericanSpellerRule_GetOnlySuggestions(t *testing.T) {
	r := requireAmericanDict(t)
	require.NotNil(t, r.GetOnlySuggestionsFn)
	only := r.GetOnlySuggestionsFn("cemetary")
	require.Equal(t, []string{"cemetery"}, only)
	only = r.GetOnlySuggestionsFn("Cemetary")
	require.Equal(t, []string{"Cemetery"}, only)
	ms, err := r.Match(analyzeEN("cemetary"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"cemetery"}, ms[0].GetSuggestedReplacements())
}

// --- helpers ---

func matchCount(t *testing.T, r *MorfologikVariantSpellerRule, text string) int {
	t.Helper()
	ms, err := r.Match(analyzeEN(text))
	require.NoError(t, err)
	return len(ms)
}

func matchCountSoft(t *testing.T, r *MorfologikVariantSpellerRule, text string) int {
	t.Helper()
	ms, err := r.Match(analyzeENSoft(text))
	require.NoError(t, err)
	return len(ms)
}

// assertFirstSuggestions ports Java assertFirstMatch: first N suggestions in order.
func assertFirstSuggestions(t *testing.T, r *MorfologikVariantSpellerRule, input string, expected ...string) {
	t.Helper()
	ms, err := r.Match(analyzeEN(input))
	require.NoError(t, err, input)
	require.Equal(t, 1, len(ms), "Expected 1 match for %q, got %d %v", input, len(ms), ms)
	sugs := ms[0].GetSuggestedReplacements()
	require.GreaterOrEqual(t, len(sugs), len(expected), "%q sugs %v", input, sugs)
	for i, exp := range expected {
		require.Equal(t, exp, sugs[i], "%q suggestion[%d] want %q got %v", input, i, exp, sugs)
	}
}

// assertSuggestionContains ports MorfologikAmericanSpellerRuleTest.assertSuggestion (contains).
func assertSuggestionContains(t *testing.T, r *MorfologikVariantSpellerRule, input string, expected ...string) {
	t.Helper()
	ms, err := r.Match(analyzeEN(input))
	require.NoError(t, err, input)
	require.Equal(t, 1, len(ms), "Expected 1 match for %q, got %d", input, len(ms))
	sugs := ms[0].GetSuggestedReplacements()
	require.GreaterOrEqual(t, len(sugs), len(expected), "%q sugs %v", input, sugs)
	for _, exp := range expected {
		require.Contains(t, sugs, exp, "%q missing %q in %v", input, exp, sugs)
	}
}

// assertAllSuggestions ports assertAllMatches: expected appear in order as a prefix.
func assertAllSuggestions(t *testing.T, r *MorfologikVariantSpellerRule, input string, expected ...string) {
	t.Helper()
	assertFirstSuggestions(t, r, input, expected...)
}

// countMapLM is a minimal CountProvider for NamedEntityIgnore smoke (FakeLanguageModel twin).
type countMapLM map[string]int

func (m countMapLM) GetCountToken(token string) int64 {
	if n, ok := m[token]; ok {
		return int64(n)
	}
	return 0
}

func (m countMapLM) GetCount(tokens []string) int64 {
	if len(tokens) == 0 {
		return 0
	}
	if len(tokens) == 1 {
		return m.GetCountToken(tokens[0])
	}
	return 0
}

var _ CountProvider = countMapLM(nil)

func firstENSuggestion(ms []*rules.RuleMatch, want string) *rules.RuleMatch {
	for _, m := range ms {
		if m == nil {
			continue
		}
		for _, s := range m.GetSuggestedReplacements() {
			if s == want {
				return m
			}
		}
	}
	return nil
}
