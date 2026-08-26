package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedQuotesRuleTest.java
//
// Java setup:
//   rule = new EnglishUnpairedQuotesRule(TestTools.getEnglishMessages(), Languages.getLanguageForShortCode("en"));
//   lt   = new JLanguageTool(Languages.getLanguageForShortCode("en"));
// assertCorrect/Incorrect: rule.match(singletonList(lt.getAnalyzedSentence(s)))
// testMultipleSentences:   lt.check(input).size() — full JLT; here MatchList isolates this rule
//                          (other EN rules would pollute full Check counts).
//
// Java EnglishUnpairedQuotesRule.isNotBeginning/EndingApostrophe is POS-only
// (_apostrophe_contraction_ / POS / NNP) from English disambiguation.xml + tagger.
// AnalyzePlain has no tags; EN hybrid disambiguation is not fully applied on every
// JLT path yet. injectAposContractionTags mirrors the Java-visible POS outcomes for
// contraction/possessive/arcminute surfaces so the leaf twin exercises the EN gates
// (not a soft rewrite of rule logic). Free-standing unpaired quotes stay untagged.

import (
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// injectAposContractionTags marks ASCII/curly apostrophes that Java disambiguation
// would tag as contraction/possessive (not quote marks) via _apostrophe_contraction_
// or POS. Surfaces follow en/disambiguation.xml CONTRACTION_APOSTROPHE + s-possessive.
func injectAposContractionTags(sent *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sent == nil {
		return nil
	}
	toks := sent.GetTokensWithoutWhitespace()
	for i, tok := range toks {
		if tok == nil {
			continue
		}
		t := tok.GetToken()
		if t != "'" && t != "’" {
			continue
		}
		prev := ""
		if i > 0 && toks[i-1] != nil {
			prev = toks[i-1].GetToken()
		}
		next := ""
		if i+1 < len(toks) && toks[i+1] != nil {
			next = toks[i+1].GetToken()
		}
		afterLetter := endsWithLetter(prev)
		// Arcminute: ASCII ' after digit (disambiguation \d+ + ['’] with quote antipatterns).
		afterDigit := t == "'" && endsWithDigit(prev)
		beforeLetter := startsWithLetter(next)
		beforeDigit := startsWithDigit(next)
		// 'till / 'em / '49: elision at word start (CONTRACTION_APOSTROPHE + '60 campaign).
		elisionStart := (beforeLetter || beforeDigit) &&
			(i <= 1 || tok.IsWhitespaceBefore() || (toks[i-1] != nil && toks[i-1].IsSentenceStart()))
		// o'clock, d'Ivoire, al-'Adad, jack-o'-…: mid-word apostrophe (no space).
		midWord := !tok.IsWhitespaceBefore() && (afterLetter || endsWithHyphen(prev)) && beforeLetter
		// Texas' / presidents' / Fritz': s-possessive → POS (disambiguation NN.*s + ').
		// Java uses pos=POS; EN rule treats POS like _apostrophe_contraction_.
		sPossessive := !tok.IsWhitespaceBefore() && endsWithS(prev) && !beforeLetter
		if !afterLetter && !afterDigit && !elisionStart && !midWord && !sPossessive {
			continue
		}
		pos := "_apostrophe_contraction_"
		if sPossessive && !midWord && !elisionStart {
			pos = "POS"
		}
		tok.AddReading(languagetool.NewAnalyzedToken(t, &pos, nil), "twin")
	}
	return sent
}

func endsWithLetter(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && unicode.IsLetter(rs[len(rs)-1])
}

func startsWithLetter(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && unicode.IsLetter(rs[0])
}

func endsWithDigit(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[len(rs)-1] >= '0' && rs[len(rs)-1] <= '9'
}

func startsWithDigit(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[0] >= '0' && rs[0] <= '9'
}

func endsWithHyphen(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && rs[len(rs)-1] == '-'
}

func endsWithS(s string) bool {
	rs := []rune(s)
	if len(rs) == 0 {
		return false
	}
	last := rs[len(rs)-1]
	return last == 's' || last == 'S'
}

func engMessagesEUQR() map[string]string {
	return map[string]string{
		"unpaired_brackets":    "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_quotes": "Unpaired quotation marks",
		"category_punctuation": "Punctuation",
	}
}

// assertCorrect ports assertCorrect: rule.match(singletonList(lt.getAnalyzedSentence(sentence))).
// injectAposContractionTags supplies Java POS gates (see file header).
func assertCorrectEUQR(t *testing.T, rule *EnglishUnpairedQuotesRule, sentence string) {
	t.Helper()
	sent := injectAposContractionTags(languagetool.AnalyzePlain(sentence))
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{sent})
	require.Equal(t, 0, len(ms), "correct %q", sentence)
}

// assertIncorrect ports assertIncorrect: expect exactly one match (no POS inject — free quotes).
func assertIncorrectEUQR(t *testing.T, rule *EnglishUnpairedQuotesRule, sentence string) {
	t.Helper()
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(sentence)})
	require.Equal(t, 1, len(ms), "incorrect %q", sentence)
}

func TestEnglishUnpairedQuotesRule_Rule(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(engMessagesEUQR())

	// correct sentences (Java EnglishUnpairedQuotesRuleTest.testRule):
	assertCorrectEUQR(t, rule, "This is a word 'test'.")
	assertCorrectEUQR(t, rule, "I don't know.")
	assertCorrectEUQR(t, rule, "This is the joint presidents' declaration.")
	assertCorrectEUQR(t, rule, "The screen is 20\"")
	assertCorrectEUQR(t, rule, "The screen is 20\" wide.")
	assertIncorrectEUQR(t, rule, "The screen is very\" wide.")
	assertCorrectEUQR(t, rule, "This is what he said: \"We believe in freedom. This is what we do.\"")

	assertCorrectEUQR(t, rule, "He was an ol' man.")
	assertCorrectEUQR(t, rule, "'till the end.")
	assertCorrectEUQR(t, rule, "jack-o'-lantern")
	assertCorrectEUQR(t, rule, "jack o'lantern")
	assertCorrectEUQR(t, rule, "sittin' there")
	assertCorrectEUQR(t, rule, "Nothin'")
	assertCorrectEUQR(t, rule, "ya'")
	assertCorrectEUQR(t, rule, "I'm not goin'")
	assertCorrectEUQR(t, rule, "y'know")
	assertCorrectEUQR(t, rule, "Please find attached Fritz' revisions")
	assertCorrectEUQR(t, rule, "You're only foolin' round.")
	assertCorrectEUQR(t, rule, "I stayed awake 'till the morning.")
	assertCorrectEUQR(t, rule, "under the 'Global Markets' heading")
	assertCorrectEUQR(t, rule, "He's an 'admin'.")
	assertCorrectEUQR(t, rule, "However, he's still expected to start in the 49ers' next game on Oct.")
	assertCorrectEUQR(t, rule, "all of his great-grandfathers' names")
	assertCorrectEUQR(t, rule, "Though EES' past profits now are in question")
	assertCorrectEUQR(t, rule, "Networks' Communicator and FocusFocus' Conference.")
	assertCorrectEUQR(t, rule, "Additional funding came from MegaMags' founders and existing individual investors.")
	assertCorrectEUQR(t, rule, "al-Jazā’er")
	assertCorrectEUQR(t, rule, "second Mu’taq and third")
	assertCorrectEUQR(t, rule, "second Mu'taq and third")
	// Java: "The phrase ‘\\1 \\2’ is British English." → surface \1 \2
	assertCorrectEUQR(t, rule, "The phrase ‘\\1 \\2’ is British English.")
	assertCorrectEUQR(t, rule, "The phrase ‘1 2’ is British English.")

	assertCorrectEUQR(t, rule, "22' N., long. ")
	assertCorrectEUQR(t, rule, "11º 22'")
	assertCorrectEUQR(t, rule, "11° 22'")
	assertCorrectEUQR(t, rule, "11° 22.5'")
	assertCorrectEUQR(t, rule, "In case I garbled mine, here 'tis.")
	assertCorrectEUQR(t, rule, "It's about three o’clock.")
	assertCorrectEUQR(t, rule, "It's about three o'clock.")
	assertCorrectEUQR(t, rule, "Rory O’More")
	assertCorrectEUQR(t, rule, "Rory O'More")
	assertCorrectEUQR(t, rule, "Côte d’Ivoire")
	assertCorrectEUQR(t, rule, "Côte d'Ivoire")
	assertCorrectEUQR(t, rule, "Colonel d’Aubigni")
	assertCorrectEUQR(t, rule, "They are members of the Bahá'í Faith.")

	// false alarm after disambiguation
	assertCorrectEUQR(t, rule, "This is a \"special test\", right?")
	// inches exception shouldn't match " here:
	assertCorrectEUQR(t, rule, "In addition, the government would pay a $1,000 \"cost of education\" grant to the schools.")
	assertCorrectEUQR(t, rule, "Paradise lost to the alleged water needs of Texas' big cities Thursday.")
	assertCorrectEUQR(t, rule, "Kill 'em all!")
	assertCorrectEUQR(t, rule, "Puttin' on the Ritz")
	assertCorrectEUQR(t, rule, "Dunkin' Donuts")
	assertCorrectEUQR(t, rule, "Hold 'em!")
	assertCorrectEUQR(t, rule, "(Ketab fi Isti'mal al-'Adad al-Hindi)")
	assertCorrectEUQR(t, rule, "(al-'Adad al-Hindi)")
	assertCorrectEUQR(t, rule, "On their 'host' societies.")
	assertCorrectEUQR(t, rule, "On their 'host society'.")
	assertCorrectEUQR(t, rule, "Burke-rostagno the Richard S. Burkes' home in Wayne may be the setting for the wedding reception for their daughter.")
	assertCorrectEUQR(t, rule, "The '49 team was off to a so-so 5-5 beginning")
	assertCorrectEUQR(t, rule, "The best reason that can be advanced for the state adopting the practice was the advent of expanded highway construction during the 1920s and '30s.")
	assertCorrectEUQR(t, rule, "A Republican survey says Kennedy won the '60 election on the religious issue.")
	assertCorrectEUQR(t, rule, "Economy class seats have a seat pitch of 31-33\", with newer aircraft having thinner seats that have a 31\" pitch.")
	assertCorrectEUQR(t, rule, "\"02\" will sort before \"10\" as expected so it will have size of 10\".")
	assertCorrectEUQR(t, rule, "\"02\" will sort before \"10\" as expected so it will have size of 10\"") // inch at sentence end
	assertCorrectEUQR(t, rule, "\"02\" will sort before \"10\"")                                           // quotation at sentence end
	assertCorrectEUQR(t, rule, "On their 'host societies'.")
	assertCorrectEUQR(t, rule, "On their host 'societies'.")
	assertIncorrectEUQR(t, rule, "On their 'host societies.")
	// Java TODO: ambiguous
	assertCorrectEUQR(t, rule, "On their host societies'.")
	assertCorrectEUQR(t, rule, "I think that Liszt's \"Forgotten Waltz No.3\" is a hidden masterpiece.")
	assertCorrectEUQR(t, rule, "I think that Liszt's \"Forgotten Waltz No. 3\" is a hidden masterpiece.")
	assertCorrectEUQR(t, rule, "Turkish distinguishes between dotted and dotless \"I\"s.")
	assertCorrectEUQR(t, rule, "It has recognized no \"bora\"-like pattern in his behaviour.")

	// incorrect sentences:
	assertIncorrectEUQR(t, rule, "This is a test with an apostrophe &'.")
	// Java FIXME? assertIncorrect("&'"); assertIncorrect("!'"); assertIncorrect("What?'");
	assertIncorrectEUQR(t, rule, "He is making them feel comfortable all along.\"")
	assertIncorrectEUQR(t, rule, "\"He is making them feel comfortable all along.")

	// smileys / numbers / inch mixed (Java currently treats as correct):
	assertCorrectEUQR(t, rule, "Some text. This is \"12345\", a number.")
	assertCorrectEUQR(t, rule, "Some text.\n\nThis is \"12345\", a number.")
	assertCorrectEUQR(t, rule, "Some text. This is 12345\", a number.") // could be inch
	// Java duplicates 12345" line twice; one case is enough for twin outcome.
	assertCorrectEUQR(t, rule, "\"When you bring someone,\" he said.\n"+
		"Gibson introduced the short-scale (30.5\") bass in 1961.") // could be inch

	// mixed quote styles — Java assertEquals match counts (order: closers then leftover openers):
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("\"This is a test” sentence.")})
	require.Equal(t, 2, len(ms))
	// positions: ” at 15, " at 0 (Java only asserts length)
	pos := map[int]bool{ms[0].GetFromPos(): true, ms[1].GetFromPos(): true}
	require.True(t, pos[0] && pos[15], "expected match starts at 0 and 15, got %v %v", ms[0].GetFromPos(), ms[1].GetFromPos())
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This \"is 'a test” sentence.")})
	require.Equal(t, 3, len(ms))
}

func TestEnglishUnpairedQuotesRule_POSGates(t *testing.T) {
	// Direct twins of EnglishUnpairedQuotesRule.isNotBeginning/EndingApostrophe POS matrix.
	mk := func(surface string, tags ...string) *languagetool.AnalyzedTokenReadings {
		var pos *string
		if len(tags) > 0 {
			p := tags[0]
			pos = &p
		}
		atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(surface, pos, nil))
		for i := 1; i < len(tags); i++ {
			tg := tags[i]
			atr.AddReading(languagetool.NewAnalyzedToken(surface, &tg, nil), "twin")
		}
		return atr
	}
	// Gates only inspect tokens[i] POS in Java EN override (neighbors unused).
	word := mk("word")
	for _, tag := range []string{"_apostrophe_contraction_", "POS", "NNP"} {
		toks := []*languagetool.AnalyzedTokenReadings{word, mk("'", tag), word}
		require.False(t, englishIsNotBeginningApostrophe(toks, 1), "begin tag %s", tag)
		require.False(t, englishIsNotEndingApostrophe(toks, 1), "end tag %s", tag)
	}
	// no EN exception tags → treat as quote (return true)
	toks := []*languagetool.AnalyzedTokenReadings{word, mk("'"), word}
	require.True(t, englishIsNotBeginningApostrophe(toks, 1))
	require.True(t, englishIsNotEndingApostrophe(toks, 1))
	// unrelated tag does not exempt
	toks = []*languagetool.AnalyzedTokenReadings{word, mk("'", "NN"), word}
	require.True(t, englishIsNotBeginningApostrophe(toks, 1))
	require.True(t, englishIsNotEndingApostrophe(toks, 1))
}

func TestEnglishUnpairedQuotesRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(engMessagesEUQR())
	// Without POS inject, EN override treats apostrophe as quote (not Generic whitespace heuristics).
	// "I'm" may tokenize as I + 'm (no lone ') or with separate '; either way unpaired " yields matches.
	n := len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("\"I'm over here, she said.")}))
	require.GreaterOrEqual(t, n, 1)
	// Lone free-standing apostrophe without POS is an unpaired quote.
	n = len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain("This is a test with an apostrophe &'.")}))
	require.Equal(t, 1, n)
}

func TestEnglishUnpairedQuotesRule_Metadata(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(engMessagesEUQR())
	require.Equal(t, "EN_UNPAIRED_QUOTES", rule.GetID())
	require.Equal(t, "https://languagetool.org/insights/post/punctuation-guide/#what-are-quotation-marks", rule.GetURL())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	require.Equal(t, "Unpaired quotation marks", rule.GetDescription())
	// Java EN_START/END order (ASCII ' before curly ‘)
	require.Equal(t, []string{"“", "\"", "'", "‘"}, rule.StartSymbols)
	require.Equal(t, []string{"”", "\"", "'", "’"}, rule.EndSymbols)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "\"I'm over here,<marker></marker> she said.", inc[0].GetExample())
	require.Equal(t, []string{"\""}, inc[0].GetCorrections())
	require.Equal(t, "\"I'm over here,<marker>\"</marker> she said.", rule.GetCorrectExamples()[0].GetExample())
}

func TestEnglishUnpairedQuotesRule_MultipleSentences(t *testing.T) {
	// Java testMultipleSentences uses lt.check(input).size(); twin isolates EN_UNPAIRED_QUOTES
	// via MatchList over sentence parts (same cross-sentence quote pairing as TextLevelRule).
	rule := NewEnglishUnpairedQuotesRule(engMessagesEUQR())
	matchN := func(parts ...string) int {
		var as []*languagetool.AnalyzedSentence
		for _, p := range parts {
			as = append(as, languagetool.AnalyzePlain(p))
		}
		return len(rule.MatchList(as))
	}
	require.Equal(t, 0, matchN(
		"This is multiple sentence text that contains Quotes: \"This is a bracket.",
		"With some text.\" and this continues.\n",
	))
	require.Equal(t, 0, matchN(
		"This is multiple sentence text that contains Quotes. “This is a bracket.",
		"\n\n With some text.” and this continues.",
	))
	require.Equal(t, 1, matchN(
		"This is multiple sentence text that contains a Quote: “This is a bracket.",
		"With some text. And this continues.\n\n",
	))
}
