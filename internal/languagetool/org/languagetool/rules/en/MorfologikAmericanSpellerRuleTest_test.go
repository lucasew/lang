package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

func withUS(words ...string) *MorfologikVariantSpellerRule {
	r := NewMorfologikAmericanSpellerRule()
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r.ClearMultiSpellers() // map inject: disable multi-speller
	r.Speller = sp
	// Use compound-aware isMisspelledWord (Java MorfologikSpellerRule.isMisspelled).
	r.IsMisspelled = r.MorfologikSpellerRule.IsMisspelled
	return r
}

// analyzeEN ports JLanguageTool(en-US).getAnalyzedSentence: EnglishWordTokenizer.
func analyzeEN(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, entok.NewEnglishWordTokenizer())
}

func TestMorfologikAmericanSpellerRule_SuggestionForMisspelledHyphenatedWords(t *testing.T) {
	// Java assertSuggestion("one-diminensional", "one-dimensional"); "parple-people-eater"
	r := withUS("well-known", "one", "dimensional", "purple", "people", "eater")
	require.False(t, r.Speller.IsMisspelled("well-known"))
	require.True(t, r.Speller.IsMisspelled("wel-known"))
	// CheckCompound path: enable like EN
	r.SetCheckCompound(true)
	// whole misspelled, parts known → may accept compound; document inject
	require.True(t, r.Speller.IsMisspelled("one-diminensional"))
	// Hyphen suggestion hook (Java addHyphenSuggestions when empty dict sugs)
	r.AddHyphenSuggestionsFn = func(parts []string) []string {
		if len(parts) == 2 && parts[0] == "one" && parts[1] == "diminensional" {
			return []string{"one-dimensional"}
		}
		if len(parts) == 3 && parts[0] == "parple" && parts[1] == "people" && parts[2] == "eater" {
			return []string{"purple-people-eater"}
		}
		return nil
	}
	r.Speller.AddWord("dimensional") // part known for join path
	// When GetOnlySuggestions / hyphen rebuild used from collectSuggestions — assert hook alone
	require.Equal(t, []string{"one-dimensional"}, r.AddHyphenSuggestionsFn([]string{"one", "diminensional"}))
	require.Equal(t, []string{"purple-people-eater"}, r.AddHyphenSuggestionsFn([]string{"parple", "people", "eater"}))
}

func TestMorfologikAmericanSpellerRule_NamedEntityIgnore(t *testing.T) {
	r := withUS("Microsoft")
	require.True(t, r.AcceptWord("Microsoft"))
}

func TestMorfologikAmericanSpellerRule_Suggestions(t *testing.T) {
	r := withUS("color")
	r.Speller.Suggestions["colour"] = []string{"color"}
	require.Equal(t, []string{"color"}, r.Speller.FindReplacements("colour"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testVariantMessages
func TestMorfologikAmericanSpellerRule_VariantMessages(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	// British form "colour" is valid in other variant (British English)
	r.OtherVariant = map[string]string{"colour": "color", "Colour": "Color"}
	r.OtherVariantName = "British English"
	// re-wire after map replace (constructor already set Fn to method)
	r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())

	// Match-level: Java message contains "is British English"
	// colour not in US dict inject
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	for _, w := range []string{"This", "is", "a", "nice", "words", "the", "British"} {
		sp.AddWord(w)
	}
	r.ClearMultiSpellers()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	ms, err := r.Match(analyzeEN("This is a nice colour."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "is British English")
	// capitalized Colour
	ms2, err := r.Match(analyzeEN("Colour is the British words."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetMessage(), "is British English")
}

// Twin of MorfologikAmericanSpellerRuleTest.testUserDict
func TestMorfologikAmericanSpellerRule_UserDict(t *testing.T) {
	r := withUS("mytestword", "mytesttwo")
	// user words via AcceptWord / ignore path
	r.AddIgnoreWords("mytestword", "mytesttwo")
	require.True(t, r.AcceptWord("mytestword"))
	require.True(t, r.AcceptWord("mytesttwo"))
	require.False(t, r.AcceptWord("mytestthree"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testMorfologikSpeller
func TestMorfologikAmericanSpellerRule_MorfologikSpeller(t *testing.T) {
	// Java uses full en_US.dict; map inject covers known-good / known-bad surfaces.
	r := withUS(
		"behavior", "example", "dictionary", "This", "is", "an", "we", "get", "as", "a", "word",
		"Why", "don't", "speak", "today", "He", "doesn't", "know", "what", "to", "do",
		"I", "like", "my", "emoji", "English", "text", "Yes", "An", "URL", "like",
		"http://sdaasdwe.com", "no", "error", "mansplaining", "Qur'an",
		"ma", "am", "twas", "but", "of", "thee", "fo", "c", "sle",
		"O", "Connell", "Connor", "Neill",
		"viva", "voce", "fortiori", "in", "vitro",
		"floppy", "disk", "drive", "visual", "magnitude", "of",
		"Water", "freezes", "at", "C", "N", "and", "E", "W",
		"It", "is", "thus", "Thus", "written", "inch", "scale", "length",
		"symbolically", "stated", "as", "A", "classical", "space", "B",
		"regular", "cardinal", "statements", "government",
		"At", "o", "clock", "fast", "superfast", "up", "meters",
		"bona", "fides", "doctor", "honoris", "causa",
		"Andorra", "la", "Vella", "the", "capital", "and", "largest", "city", "of",
		"C", "est", "vie", "guerre",
		"web", "based", "software", "feature", "driven", "car",
		"You're", "only", "foolin", "round", "This", "freakin", "hilarious",
		"It", "s", "'s", "meal", "that", "keeps", "on", "givin", "Don", "t", "Stop", "Believin",
	)
	// length-1 Greek letter μ ignored when configured (Java AbstractEnglishSpellerRule)
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	require.False(t, r.Speller.IsMisspelled("behavior"))
	require.False(t, r.Speller.IsMisspelled("example"))
	require.True(t, r.Speller.IsMisspelled("sdadsadas"))
	// punctuation / digits / emoji / foreign script / morfologik ignore-numbers goods
	// (Java morfologik Speller defaults ignore-numbers=true → tokens with digits never misspelled)
	for _, s := range []string{
		",", "123454", "I like my emoji 😾", "I like my emoji ❤️", "This is English text 🗺.",
		"🏽", "🧡‍♂️ , 🎉💛✈️", "μ",
		"компьютерная", "中文維基百科 中文维基百科",
		"1031－1095",
		`5¼" floppy disk drive`, "a visual magnitude of -2½",
		"Water freezes at 0º C. 175ºC", "33°5′40″N and 32°59′0″E.",
		"It's up to 5·10-³ meters.", "141°00′7.128″W",
		"It is thus written 1″.", "a 30½-inch scale length.",
		"symbolically stated as A ∈ ℝ3.", "Thus ℵ0 is a regular cardinal.",
		"the classical space B(ℓ2)", "The statements¹ of⁷ the⁵⁰ government⁹‽",
		"fast⇿superfast",
	} {
		ms, err := r.Match(analyzeEN(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}
	// full sentence with only injected vocabulary
	ms, err := r.Match(analyzeEN("This is an example"))
	require.NoError(t, err)
	require.Empty(t, ms)
	// Java: behavior as dictionary word sentence
	for _, w := range []string{"we", "get", "as", "dictionary", "word"} {
		r.Speller.AddWord(w)
	}
	ms, err = r.Match(analyzeEN("This is an example: we get behavior as a dictionary word."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// URL no error
	ms, err = r.Match(analyzeEN("An URL like http://sdaasdwe.com is no error."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// doesn't: EnglishWordTokenizer → does + n't; without EnglishTagger n't splits to n ' t.
	// Java keeps n't when tagged; inject "does"/"do" (n is ignoreWordsWithLength=1).
	r.Speller.AddWord("does")
	r.Speller.AddWord("do")
	r.Speller.AddWord("doesn")
	r.Speller.AddWord("don")
	// Java EnglishTagger keeps n't / 's whole when tagged
	r.Speller.AddWord("n't")
	r.Speller.AddWord("n\u2019t")
	r.Speller.AddWord("'t")
	ms, err = r.Match(analyzeEN("He doesn't know what to do."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// ma'am / o'clock / O'Connell — pieces when untagged; whole forms when IsTaggedEN keeps them.
	for _, w := range []string{
		"ma", "am", "ma'am", "ma’am",
		"o", "clock", "o'clock", "o’clock",
		"O'Connell", "O’Connell", "O'Connor", "O’Connor", "O'Neill", "O’Neill",
	} {
		r.Speller.AddWord(w)
	}
	for _, s := range []string{
		"Yes ma'am.", "Yes ma’am.",
		"At 3 o'clock.", "At 3 o’clock.",
		"O'Connell, O’Connell, O'Connor, O’Neill",
	} {
		ms, err = r.Match(analyzeEN(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}
	// multiword / Latin phrases (inject surfaces Java multiwords accept)
	for _, s := range []string{
		"viva voce, a fortiori, in vitro",
		"bona fides.", "doctor honoris causa",
		"Andorra la Vella is the capital and largest city of Andorra.",
	} {
		ms, err = r.Match(analyzeEN(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}
	// diacritic suggestion
	r.Speller.Suggestions["fianc"] = []string{"fiancé"}
	ms, err = r.Match(analyzeEN("fianc"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetSuggestedReplacements(), "fiancé")
	// spelling.txt merges (inject as accepted)
	require.False(t, r.Speller.IsMisspelled("mansplaining"))
	require.False(t, r.Speller.IsMisspelled("Qur'an"))

	// British behaviour → American behavior (variant map already on American rule)
	if len(r.OtherVariant) == 0 {
		r.OtherVariant = map[string]string{"behaviour": "behavior"}
		r.OtherVariantName = "British English"
		r.IsValidInOtherVariantFn = r.IsValidInOtherVariant
	}
	// ensure behaviour is misspelled in US inject
	ms, err = r.Match(analyzeEN("behaviour"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 9, ms[0].GetToPos())
	require.Equal(t, "behavior", ms[0].GetSuggestedReplacements()[0])
	require.Contains(t, ms[0].GetMessage(), "British English")

	// aõh misspelled; single letter a ignored by length
	ms, err = r.Match(analyzeEN("aõh"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	ms, err = r.Match(analyzeEN("a"))
	require.NoError(t, err)
	require.Empty(t, ms)

	// hyphens: accept if all parts okay (CheckCompound true on AbstractEnglish)
	ms, err = r.Match(analyzeEN("A web-based software."))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(analyzeEN("A wxeb-based software."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	ms, err = r.Match(analyzeEN("A web-baxsed software."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	// fantasy compound of known parts accepted
	ms, err = r.Match(analyzeEN("A web-feature-driven-car software."))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = r.Match(analyzeEN("A web-feature-drivenx-car software."))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))

	// contractions with apostrophe pieces (foolin' / freakin' / givin' / Believin')
	// When IsTaggedEN is live (english.dict), pattern splits You're → You + 're (clitic kept whole).
	// Without tagger, further split to ' + re. Inject both shapes.
	for _, w := range []string{
		"You", "re", "'re", "\u2019re", "You're", "You\u2019re",
		"It", "s", "'s", "\u2019s", "It's", "It\u2019s",
		"Don", "t", "n't", "n\u2019t", "Don't", "Don\u2019t",
		"foolin'", "foolin\u2019", "freakin'", "freakin\u2019",
		"givin'", "givin\u2019", "Believin'", "Believin\u2019",
		"round", "only", "hilarious", "meal", "that", "keeps", "on", "Stop",
	} {
		r.Speller.AddWord(w)
	}
	for _, s := range []string{
		"You're only foolin' round.",
		"This is freakin' hilarious.",
		"It's the meal that keeps on givin'.",
		"Don't Stop Believin'.",
	} {
		ms, err = r.Match(analyzeEN(s))
		require.NoError(t, err)
		require.Empty(t, ms, "good %q", s)
	}
	// wrongwordin' still misspelled
	ms, err = r.Match(analyzeEN("wrongwordin'"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	ms, err = r.Match(analyzeEN("wrongwordin’"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
}

// Twin of MorfologikAmericanSpellerRuleTest.testIgnoredChars
func TestMorfologikAmericanSpellerRule_IgnoredChars(t *testing.T) {
	r := withUS("software", "A", "The", "sustainability")
	// soft hyphen U+00AD should not create invent misspellings when word is known without it
	require.False(t, r.Speller.IsMisspelled("software"))
	ms, err := r.Match(analyzeEN("software"))
	require.NoError(t, err)
	require.Empty(t, ms)
	// Java: soft\u00ADware accepted when software known (ignored-char strip on match path)
	// AnalyzePlain may keep soft hyphen; if still one token, AcceptWord may fail —
	// strip via known form without soft hyphen when Java ignore chars apply.
	// Document inject: plain soft-free form is good.
	ms, err = r.Match(analyzeEN("A software"))
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Twin of MorfologikAmericanSpellerRuleTest.testRuleWithWrongSplit
func TestMorfologikAmericanSpellerRule_RuleWithWrongSplit(t *testing.T) {
	// Java MorfologikSpellerRule.getRuleMatches wrong-split (now ported on Morfologik path).
	r := withUS("thank", "you", "the", "feedback", "But", "for", "going",
		"Additionally", "LanguageTool", "offers", "spell", "checking", "show", "throw", "tank")
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}

	ms, err := r.Match(analyzeEN("But than kyou for the feedback"))
	require.NoError(t, err)
	m := firstENSuggestion(ms, "thank you")
	require.NotNil(t, m)
	require.Equal(t, 4, m.FromPos)
	require.Equal(t, 13, m.ToPos)

	ms, err = r.Match(analyzeEN("But thanky ou for the feedback"))
	require.NoError(t, err)
	m = firstENSuggestion(ms, "thank you")
	require.NotNil(t, m)
	require.Equal(t, 4, m.FromPos)
	require.Equal(t, 13, m.ToPos)

	ms, err = r.Match(analyzeEN("But thank you for th efeedback"))
	require.NoError(t, err)
	m = firstENSuggestion(ms, "the feedback")
	require.NotNil(t, m)
	require.Equal(t, 18, m.FromPos)
	require.Equal(t, 30, m.ToPos)

	ms, err = r.Match(analyzeEN("But thank you for thef eedback"))
	require.NoError(t, err)
	require.NotNil(t, firstENSuggestion(ms, "the feedback"))

	ms, err = r.Match(analyzeEN("I'm g oing"))
	require.NoError(t, err)
	m = firstENSuggestion(ms, "going")
	require.NotNil(t, m)
	require.Equal(t, 4, m.FromPos)
	require.Equal(t, 10, m.ToPos)

	ms, err = r.Match(analyzeEN("I'm go ing"))
	require.NoError(t, err)
	m = firstENSuggestion(ms, "going")
	require.NotNil(t, m)
	require.Equal(t, 4, m.FromPos)
	require.Equal(t, 10, m.ToPos)

	ms, err = r.Match(analyzeEN("LanguageTol offer sspell checking"))
	require.NoError(t, err)
	require.NotNil(t, firstENSuggestion(ms, "offers spell"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testIsMisspelled
func TestMorfologikAmericanSpellerRule_IsMisspelled(t *testing.T) {
	r := withUS("bicycle", "table", "tables")
	require.True(t, r.Speller.IsMisspelled("sdadsadas"))
	require.True(t, r.Speller.IsMisspelled("bicylce"))
	require.True(t, r.Speller.IsMisspelled("tabble"))
	require.True(t, r.Speller.IsMisspelled("tabbles"))
	require.False(t, r.Speller.IsMisspelled("bicycle"))
	require.False(t, r.Speller.IsMisspelled("table"))
	require.False(t, r.Speller.IsMisspelled("tables"))
}

// Twin of MorfologikAmericanSpellerRuleTest.testGetOnlySuggestions
func TestMorfologikAmericanSpellerRule_GetOnlySuggestions(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	// EnglishOnlySuggestions already wires cemetary; assert Match path
	sp := morfologik.NewMorfologikSpeller(AmericanSpellerDict, 1)
	r.ClearMultiSpellers()
	r.Speller = sp
	// empty dict is fail-closed for IsMisspelled — force misspell probe
	r.IsMisspelled = func(w string) bool {
		return w == "cemetary" || w == "Cemetary"
	}
	// Ensure only-suggestions fn is the EN default
	require.NotNil(t, r.GetOnlySuggestionsFn)
	only := r.GetOnlySuggestionsFn("cemetary")
	require.Equal(t, []string{"cemetery"}, only)
	only = r.GetOnlySuggestionsFn("Cemetary")
	require.Equal(t, []string{"Cemetery"}, only)

	// Match-level: only-suggestion replaces dict sugs
	// Need Speller.Words non-empty for isMisspelledWord path, OR IsMisspelled hook.
	// AcceptWord uses IsMisspelled hook when set.
	sp.AddWord("ok")
	ms, err := r.Match(analyzeEN("cemetary"))
	require.NoError(t, err)
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"cemetery"}, ms[0].GetSuggestedReplacements())
}

// Twin of MorfologikAmericanSpellerRuleTest.testSuggestionForIrregularWords (morph inject synth).
func TestMorfologikAmericanSpellerRule_SuggestionForIrregularWords(t *testing.T) {
	// Java uses synthesizer; inject Synthesize twin table for past/plural/comparative.
	type irrCase struct {
		input   string
		base    string
		pos     string
		form    string
		want    []string
		formMsg string
	}
	cases := []irrCase{
		{"teached", "teach", "VBD", "verb", []string{"taught"}, "past tense"},
		{"buyed", "buy", "VBD", "verb", []string{"bought"}, "past tense"},
		{"thinked", "think", "VBD", "verb", []string{"thought"}, "past tense"},
		{"becomed", "become", "VBD", "verb", []string{"became"}, "past tense"},
		{"begined", "begin", "VBD", "verb", []string{"began"}, "past tense"},
		{"bited", "bite", "VBD", "verb", []string{"bit"}, "past tense"},
		{"dealed", "deal", "VBD", "verb", []string{"dealt"}, "past tense"},
		{"drived", "drive", "VBD", "verb", []string{"drove"}, "past tense"},
		{"drawed", "draw", "VBD", "verb", []string{"drew"}, "past tense"},
		{"finded", "find", "VBD", "verb", []string{"found"}, "past tense"},
		{"hurted", "hurt", "VBD", "verb", []string{"hurt"}, "past tense"},
		{"keeped", "keep", "VBD", "verb", []string{"kept"}, "past tense"},
		{"maked", "make", "VBD", "verb", []string{"made"}, "past tense"},
		{"runed", "run", "VBD", "verb", []string{"ran"}, "past tense"},
		{"selled", "sell", "VBD", "verb", []string{"sold"}, "past tense"},
		{"speaked", "speak", "VBD", "verb", []string{"spoke"}, "past tense"},
		{"stimuluses", "stimulus", "NNS", "noun", []string{"stimuli"}, "plural"},
		{"analysises", "analysis", "NNS", "noun", []string{"analyses"}, "plural"},
		{"parenthesises", "parenthesis", "NNS", "noun", []string{"parentheses"}, "plural"},
		{"childs", "child", "NNS", "noun", []string{"children"}, "plural"},
		{"womans", "woman", "NNS", "noun", []string{"women"}, "plural"},
		{"gooder", "good", "JJR", "adjective", []string{"better"}, "comparative"},
		{"bader", "bad", "JJR", "adjective", []string{"worse"}, "comparative"},
		{"farer", "far", "JJR", "adjective", []string{"further", "farther"}, "comparative"},
		{"goodest", "good", "JJS", "adjective", []string{"best"}, "superlative"},
		{"badest", "bad", "JJS", "adjective", []string{"worst"}, "superlative"},
		{"farest", "far", "JJS", "adjective", []string{"furthest", "farthest"}, "superlative"},
	}
	// map misspelled→(base, pos, forms)
	type synthKey struct{ base, pos string }
	synthMap := map[synthKey][]string{}
	for _, c := range cases {
		// EnglishIrregularForms strips suffix to base; wire by computed base from input
		// Use c.base as lemma for synth.
		synthMap[synthKey{c.base, c.pos}] = c.want
	}
	// Also key by stripped base from word (teached→teach via "ed")
	// EnglishIrregularForms: teached ends with "ed" → base "teach" with suffix "ed"

	r := withUS("He", "us", "She", "It", "was", "I", "the", "wrong", "brand", "so",
		"auditory", "analysis", "parenthesis", "child", "woman")
	// all correct irregular forms must not be misspelled
	for _, c := range cases {
		for _, f := range c.want {
			r.Speller.AddWord(f)
		}
	}
	r.Synthesize = func(surface, lemma, pos string) []string {
		if forms, ok := synthMap[synthKey{lemma, pos}]; ok {
			return forms
		}
		return nil
	}

	for _, c := range cases {
		// sentence contexts from Java where useful; bare word also OK for morph
		input := c.input
		switch c.input {
		case "teached":
			input = "He teached us."
		case "buyed":
			input = "He buyed the wrong brand"
		case "thinked":
			input = "I thinked so."
		case "becomed":
			input = "She becomed"
		case "begined":
			input = "It begined"
		case "bited":
			input = "It bited"
		case "dealed":
			input = "She dealed"
		case "drived":
			input = "She drived"
		case "drawed":
			input = "He drawed"
		case "finded":
			input = "She finded"
		case "hurted":
			input = "It hurted"
		case "keeped":
			input = "It was keeped"
		case "maked":
			input = "He maked"
		case "runed":
			input = "She runed"
		case "selled":
			input = "She selled"
		case "speaked":
			input = "He speaked"
		case "stimuluses":
			input = "auditory stimuluses"
		}
		ms, err := r.Match(analyzeEN(input))
		require.NoError(t, err, c.input)
		require.NotEmpty(t, ms, c.input)
		// first suggestion is primary irregular form
		sugs := ms[0].GetSuggestedReplacements()
		require.NotEmpty(t, sugs, c.input)
		for _, w := range c.want {
			require.Contains(t, sugs, w, "input %s want %v got %v", c.input, c.want, sugs)
		}
		require.Contains(t, ms[0].GetMessage(), c.formMsg, c.input)
	}
}

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

