package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/MixedAlphabetsRuleTest.java
//
// Java setup:
//   MixedAlphabetsRule rule = new MixedAlphabetsRule(TestTools.getMessages("uk"));
//   JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real UkrainianWordTokenizer + UkrainianTagger (official ukrainian.dict)
// via wireUKWordRepeatLT — same path as Typography / UkrainianWordRepeat twins.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMixedAlphabetsRule_Rule twins MixedAlphabetsRuleTest.testRule.
func TestMixedAlphabetsRule_Rule(t *testing.T) {
	// Java: new MixedAlphabetsRule(TestTools.getMessages("uk"))
	rule := NewMixedAlphabetsRule(nil)
	// Java: new JLanguageTool(Ukrainian.DEFAULT_VARIANT)
	lt := wireUKWordRepeatLT(t)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("сміття"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("not mixed"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("123454"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("x = a якщо"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("x − a та y − b"))))
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("записати x та y через параметр t"))))

	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("ЛЮДИНИ І НАЦІЇ"))))

	// incorrect sentences:
	matches := rule.Match(lt.GetAnalyzedSentence("смiття")) // latin i
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні й латинські літери в одному слові", matches[0].GetMessage())
	require.Equal(t, []string{"сміття"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("mіхed")) // cyrillic i and x
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських", matches[0].GetMessage())
	require.Equal(t, []string{"mixed"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("горíти")) // umlaut instead of accented і
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні й латинські літери в одному слові", matches[0].GetMessage())
	require.Equal(t, []string{"горі́ти"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("двоáктний")) // latin á
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні й латинські літери в одному слові", matches[0].GetMessage())

	matches = rule.Match(lt.GetAnalyzedSentence("Чорного i Азовського")) // latin i
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито латинську «i» замість кириличної", matches[0].GetMessage())
	require.Equal(t, []string{"і"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("A нема")) // latin A
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито латинську «A» замість кириличної", matches[0].GetMessage())
	require.Equal(t, []string{"А"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Петро І")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "I", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("сибірську колекцію Петра І.\n\n Всім")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "I.", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("Миколая І.")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "I.", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("У І кварталі")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "I", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("ЗА І ПРОТИ")) // cyrillic І
	require.Equal(t, 0, len(matches))

	// ambiguous without semantics:
	// російський хемік Александр І. Опарін (1894–1980)

	matches = rule.Match(lt.GetAnalyzedSentence("Ленін В. І."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("Тому І.    Вишенський радить "))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("у І ст.")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "I", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("XІ")) // cyrillic І and latin X
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських", matches[0].GetMessage())
	require.Equal(t, []string{"XI"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("ХI")) // cyrillic X and latin I
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських", matches[0].GetMessage())
	require.Equal(t, []string{"XI"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("VIIІ-го")) // latin VII and cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських. Також: до римських цифр букви не дописуються.", matches[0].GetMessage())
	require.Equal(t, []string{"VIII"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("ІІІ-го")) // cyrillic І
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських на позначення римської цифри. Також: до римських цифр букви не дописуються.", matches[0].GetMessage())
	require.Equal(t, []string{"III"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("ХІ")) // cyrillic both X and I used for latin number
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських на позначення римської цифри", matches[0].GetMessage())
	require.Equal(t, []string{"XI"}, matches[0].GetSuggestedReplacements())

	// we split tokens for latin numbers with \u2013 inside now
	// matches = rule.match(lt.getAnalyzedSentence("ХI\u2013XX")); // commented in Java

	matches = rule.Match(lt.GetAnalyzedSentence("СOVID-19")) // cyrillic С
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні літери замість латинських", matches[0].GetMessage())
	require.Equal(t, []string{"COVID-19"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("австрo-турецької")) // latin o mixed
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличні й латинські літери в одному слові", matches[0].GetMessage())
	require.Equal(t, []string{"австро-турецької"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Щеплення від гепатиту В.")) // cyrillic B
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "B", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("група А")) // cyrillic А
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "A", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("На 0,6°С.")) // cyrillic С
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Вжито кириличну літеру замість латинської", matches[0].GetMessage())
	require.Equal(t, "C", matches[0].GetSuggestedReplacements()[0])
}

// TestMixedAlphabetsRule_CombiningChars twins MixedAlphabetsRuleTest.testCombiningChars.
func TestMixedAlphabetsRule_CombiningChars(t *testing.T) {
	rule := NewMixedAlphabetsRule(nil)
	lt := wireUKWordRepeatLT(t)

	// й and ї are done via combining characters: и + U+0306, і + U+0308
	matches := rule.Match(lt.GetAnalyzedSentence("Білоруський - українці"))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "Білоруський", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "українці", matches[1].GetSuggestedReplacements()[0])
}

// TestMixedAlphabetsRule_Metadata checks id / description / short / category.
func TestMixedAlphabetsRule_Metadata(t *testing.T) {
	r := NewMixedAlphabetsRule(nil)
	require.Equal(t, "UK_MIXED_ALPHABETS", r.GetID())
	require.Equal(t, "Змішування кирилиці й латиниці", r.GetDescription())
	require.Equal(t, "Мішанина розкладок", r.getShort())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "MISC", r.GetCategory().GetID().String())
}

// TestMixedAlphabets_REFullMatch guards Java Matcher.matches() full-string semantics.
func TestMixedAlphabets_REFullMatch(t *testing.T) {
	require.True(t, likelyLatinNumber.MatchString("XXI"))
	require.False(t, likelyLatinNumber.MatchString("preXXIpost"))
	require.True(t, cyrillicFirstLetter.MatchString("Київ"))
	require.False(t, cyrillicFirstLetter.MatchString("xКиїв"))
	require.True(t, latinNumberWithCyr.MatchString("ХІ"))
	require.False(t, latinNumberWithCyr.MatchString("аХІ"))
}
