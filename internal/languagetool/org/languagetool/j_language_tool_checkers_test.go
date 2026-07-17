package languagetool

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/stretchr/testify/require"
)

func TestSimpleMultipleWhitespaceChecker(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("WHITESPACE_RULE", SimpleMultipleWhitespaceChecker())
	m := lt.Check("hello  world")
	require.NotEmpty(t, m)
	require.Equal(t, "WHITESPACE_RULE", m[0].RuleID)
	require.Equal(t, "hello world", CorrectTextFromLocalMatches("hello  world", m))
	require.Empty(t, lt.Check("hello world"))
}

func TestSimpleUnpairedBracketsChecker(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("UNPAIRED_BRACKETS", SimpleUnpairedBracketsChecker())
	require.Empty(t, lt.Check("ok (yes) [x] {y}"))
	require.NotEmpty(t, lt.Check("broken (yes"))
	require.NotEmpty(t, lt.Check("broken yes)"))
}

func TestSimplePhraseReplaceChecker(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("PHRASE_REPLACE", SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"tot he": "to the",
	}))
	src := "Guide tot he Galaxy"
	m := lt.Check(src)
	require.NotEmpty(t, m)
	require.Equal(t, "Guide to the Galaxy", CorrectTextFromLocalMatches(src, m))
}

func TestSimplePhraseReplaceChecker_CaseInsensitive(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("PHRASE_REPLACE", SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"on accident":   "by accident",
		"in regards to": "with regard to",
	}))
	m := lt.Check("I did it On Accident yesterday.")
	require.NotEmpty(t, m)
	require.Equal(t, "by accident", m[0].Suggestions[0])
	fixed := CorrectTextFromLocalMatches("I did it On Accident yesterday.", m)
	require.Contains(t, fixed, "by accident")

	m = lt.Check("In regards to your note.")
	require.NotEmpty(t, m)
	require.Equal(t, "with regard to", m[0].Suggestions[0])
}

func TestCheckAnnotatedAndProject(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())
	at := markup.NewAnnotatedTextBuilder().
		AddText("This is an ").
		AddMarkupInterpretAs("&eacute;", "e").
		AddText(" test.").
		Build()
	// plain: "This is an e test." → an before e → maybe flag an→a? e is vowel → ok for "an e"
	// use "a error" in text parts
	at2 := markup.NewAnnotatedTextBuilder().
		AddText("See a ").
		AddMarkupInterpretAs("<b>", "").
		AddText("error").
		AddMarkupInterpretAs("</b>", "").
		AddText(" here.").
		Build()
	plain := at2.GetPlainText()
	require.Equal(t, "See a error here.", plain)
	matches := lt.CheckAnnotated(at2)
	require.NotEmpty(t, matches)
	// project to original (with markup)
	proj := ProjectMatchesToOriginal(at2, matches)
	require.Len(t, proj, len(matches))
	require.GreaterOrEqual(t, proj[0].FromPos, 0)
	_ = at
}

func TestRegisterDemoEnglishCheckers(t *testing.T) {
	lt := NewJLanguageTool("en-US")
	known := map[string]struct{}{
		"A": {}, "sentence": {}, "with": {}, "error": {}, "in": {}, "the": {},
		"Hitchhiker": {}, "s": {}, "Guide": {}, "to": {}, "Galaxy": {},
	}
	// allow apostrophe tokens soft
	lt.RegisterDemoEnglishCheckers(known, map[string][]string{"speling": {"spelling"}})
	// a error + tot he
	src := "A sentence with a error in the Hitchhiker's Guide tot he Galaxy"
	m := lt.Check(src)
	require.NotEmpty(t, m)
	ids := map[string]bool{}
	for _, x := range m {
		ids[x.RuleID] = true
	}
	require.True(t, ids["EN_A_VS_AN"] || ids["PHRASE_REPLACE"])
}

func TestSimpleMapSpellerChecker_EditDistanceSuggestions(t *testing.T) {
	known := map[string]struct{}{
		"test": {}, "the": {}, "book": {}, "hello": {},
	}
	// no explicit suggestion map → edit-distance fallback
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", SimpleMapSpellerChecker("MORFOLOGIK_RULE_EN_US", known, nil))
	m := lt.Check("tset the bok")
	require.NotEmpty(t, m)
	// tset → test, bok → book
	got := map[string][]string{}
	for _, x := range m {
		got[x.RuleID] = x.Suggestions // last wins; collect by covered word via suggestions
		if len(x.Suggestions) > 0 {
			// keep first suggestion for each match
			_ = x
		}
	}
	var hasTest, hasBook bool
	for _, x := range m {
		for _, s := range x.Suggestions {
			if s == "test" {
				hasTest = true
			}
			if s == "book" {
				hasBook = true
			}
		}
	}
	require.True(t, hasTest, "matches=%+v", m)
	require.True(t, hasBook, "matches=%+v", m)
	_ = got
}

func TestNearestKnownWords(t *testing.T) {
	known := map[string]struct{}{"receive": {}, "the": {}, "separate": {}, "xyzzy": {}}
	sugs := nearestKnownWords("recieve", known, 2, 5)
	require.Contains(t, sugs, "receive")
	require.Empty(t, nearestKnownWords("receive", known, 2, 5)) // exact known not returned (d>0)
}

func TestSimpleAvsAnChecker_PhoneticExceptions(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("EN_A_VS_AN", SimpleAvsAnChecker())

	// silent h → an
	require.Empty(t, lt.Check("This is an hour."))
	m := lt.Check("This is a hour.")
	require.NotEmpty(t, m)
	require.Equal(t, "an", m[0].Suggestions[0])

	// university /juː/ → a
	require.Empty(t, lt.Check("This is a university."))
	m = lt.Check("This is an university.")
	require.NotEmpty(t, m)
	require.Equal(t, "a", m[0].Suggestions[0])

	// european
	require.Empty(t, lt.Check("This is a European car."))
	m = lt.Check("This is an European car.")
	require.NotEmpty(t, m)

	// one-time
	require.Empty(t, lt.Check("This is a one-time offer."))
	m = lt.Check("This is an one-time offer.")
	require.NotEmpty(t, m)

	// honest
	require.Empty(t, lt.Check("He is an honest man."))
	m = lt.Check("He is a honest man.")
	require.NotEmpty(t, m)

	// still catch classic errors
	m = lt.Check("This is an test.")
	require.NotEmpty(t, m)
	require.Equal(t, "a", m[0].Suggestions[0])
}

func TestSimplePredicateSpellerChecker_IgnoresSpellerFlag(t *testing.T) {
	lt := NewJLanguageTool("en")
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US",
		func(w string) bool {
			// only Xyzzy is "unknown"; rest known so we isolate the ignore flag
			return w != "Xyzzy" && w != "xyzzy"
		},
		nil,
		nil,
		nil,
	))
	// without ignore: flags Xyzzy
	m := lt.Check("Xyzzy is here.")
	require.NotEmpty(t, m)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", m[0].RuleID)

	// inject disambiguator that ignores Xyzzy
	lt.Disambiguator = sentenceDisambiguatorFunc(func(s *AnalyzedSentence) *AnalyzedSentence {
		if s == nil {
			return nil
		}
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok != nil && tok.GetToken() == "Xyzzy" {
				tok.IgnoreSpelling()
			}
		}
		return s
	})
	m = lt.Check("Xyzzy is here.")
	for _, x := range m {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID, "%+v", m)
	}
}

type sentenceDisambiguatorFunc func(*AnalyzedSentence) *AnalyzedSentence

func (f sentenceDisambiguatorFunc) Disambiguate(s *AnalyzedSentence) *AnalyzedSentence { return f(s) }
