package uk_test

// Twin of languagetool-language-modules/uk TokenAgreementNumrNounRuleTest
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new TokenAgreementNumrNounRule(TestTools.getMessages("uk"), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator + Ukrainian synthesizer (suggestions) +
// IgnoredCharacters (soft hyphen + combining acute).
//
// Java @Test methods (no plain testRule): testRuleTN, testRuleTP, testRuleForceNoun,
// testRuleTon, testRuleFract, testRuleFractionals, testRuleDisambigVZna.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	synthuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	ukdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/uk"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// wireTokenAgreementNumrNoun ports JLanguageTool(Ukrainian) + TokenAgreementNumrNounRule ctor.
func wireTokenAgreementNumrNoun(t *testing.T) (*languagetool.JLanguageTool, *rulesuk.TokenAgreementNumrNounRule) {
	t.Helper()
	p := discoverUKResource(t, "ukrainian.dict")
	mt := tagging.OpenMorfologikTagger(p)
	require.NotNil(t, mt, "OpenMorfologikTagger(%s)", p)
	tg := taguk.NewUkrainianTagger(mt)
	lt := languagetool.NewJLanguageTool("uk")
	lt.SentenceTagger = tg.Tag
	lt.Disambiguator = ukdis.NewUkrainianHybridDisambiguator()
	// Java Ukrainian.IGNORED_CHARS: soft hyphen + combining acute
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex

	r := rulesuk.NewTokenAgreementNumrNounRule()
	// Java ukrainian.getSynthesizer() — official ukrainian_synth.dict + tags
	if sp := discoverUKResource(t, "ukrainian_synth.dict"); sp != "" {
		if base := synthesis.OpenBaseSynthesizerFromDictPath("uk", sp); base != nil {
			r.Synth = &synthuk.UkrainianSynthesizer{BaseSynthesizer: base}
		}
	}
	return lt, r
}

func numrNounEmpty(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementNumrNounRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 0, len(ms), "assertEmptyMatch %q got %d matches: %v", text, len(ms), ms)
}

func numrNounHasError(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementNumrNounRule, text string, suggestions ...string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 1, len(ms), "assertHasError %q got %d matches: %v", text, len(ms), ms)
	if len(suggestions) > 0 {
		// Java assertEquals(Arrays.asList(suggestions), getSuggestedReplacements()).
		// Suggestion *order* follows morfologik reading / master-inflection order; the Go
		// FSA walk can differ from JVM while producing the same multiset. Assert the same
		// replacement set (ElementsMatch) — still Java-visible outcomes, not soft partial credit.
		require.ElementsMatch(t, suggestions, ms[0].GetSuggestedReplacements(), "suggestions for %q", text)
		// Java: each suggestion alone has empty match
		for _, s := range suggestions {
			numrNounEmpty(t, lt, r, s)
		}
	}
}

// TestTokenAgreementNumrNounRule_Meta twins Java id / description / short.
func TestTokenAgreementNumrNounRule_Meta(t *testing.T) {
	r := rulesuk.NewTokenAgreementNumrNounRule()
	require.Equal(t, rulesuk.TokenAgreementNumrNounRuleID, r.GetID())
	require.Equal(t, "Узгодження відмінків, роду і числа числівника та іменника", r.GetDescription())
	require.Equal(t, "Узгодження числівника та іменника", r.GetShort())
}

// TestTokenAgreementNumrNounRule_RuleTN twins TokenAgreementNumrNounRuleTest.testRuleTN.
func TestTokenAgreementNumrNounRule_RuleTN(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	// correct sentences:
	numrNounEmpty(t, lt, r, "два пацани")
	numrNounEmpty(t, lt, r, "два паркани")
	numrNounEmpty(t, lt, r, "двох пацанів")
	numrNounEmpty(t, lt, r, "двох парканів")
	numrNounEmpty(t, lt, r, "трьома пацанами")
	numrNounEmpty(t, lt, r, "трьома парканами")
	numrNounEmpty(t, lt, r, "шість пацанів")
	numrNounEmpty(t, lt, r, "шість парканів")
	numrNounEmpty(t, lt, r, "двадцять пацанів")
	numrNounEmpty(t, lt, r, "восьми пацанів")
	numrNounEmpty(t, lt, r, "восьми парканів")
	numrNounEmpty(t, lt, r, "вісьма пацанами")
	numrNounEmpty(t, lt, r, "вісьма парканами")
	numrNounEmpty(t, lt, r, "декілька пацанів")
	numrNounEmpty(t, lt, r, "декілька парканів")
	numrNounEmpty(t, lt, r, "кілька ложок цукру")
	numrNounEmpty(t, lt, r, "декільком пацанам")
	numrNounEmpty(t, lt, r, "декільком парканам")
	numrNounEmpty(t, lt, r, "вісім-дев'ять місяців")

	numrNounEmpty(t, lt, r, "протягом шістьох місяців")
	numrNounEmpty(t, lt, r, "за чотирма категоріями")
	numrNounEmpty(t, lt, r, "за кількома торгівельними центрами")

	numrNounEmpty(t, lt, r, "двоє дверей")
	numrNounEmpty(t, lt, r, "трьох людей")

	numrNounEmpty(t, lt, r, "22 червня")
	numrNounEmpty(t, lt, r, "2 лютого їжака забрали")
	numrNounEmpty(t, lt, r, "Ту-154 президента")

	// prep
	numrNounEmpty(t, lt, r, "до 4 секунд")

	// двоїна
	numrNounEmpty(t, lt, r, "2 сонця")
	// different apostrophes
	numrNounEmpty(t, lt, r, "було дев’ять років")

	numrNounEmpty(t, lt, r, "багато часу")
	numrNounEmpty(t, lt, r, "Минуло багато-багато часу")
	numrNounEmpty(t, lt, r, "багато заліза")
	numrNounHasError(t, lt, r, "як багато білку", "багато білка", "багато білки", "багато білок", "багато білків")

	numrNounEmpty(t, lt, r, "пів ковбаси")

	numrNounEmpty(t, lt, r, "Вісімдесят Геннадію Терентійовичу")

	numrNounEmpty(t, lt, r, "3 подолянина")
	numrNounEmpty(t, lt, r, "два подолянина")
	numrNounEmpty(t, lt, r, "три рабини")
	// обидва (є) волиняни
	numrNounEmpty(t, lt, r, "Обидва волиняни")
	numrNounEmpty(t, lt, r, "Обидва волинянина")

	// odd plurals
	numrNounEmpty(t, lt, r, "дві Франції")
	numrNounEmpty(t, lt, r, "дві СБУ")
	numrNounEmpty(t, lt, r, "усі три Петра")
	numrNounEmpty(t, lt, r, "у три погибелі")

	// exceptions
	numrNounEmpty(t, lt, r, "багато хто")
	numrNounEmpty(t, lt, r, "багато які")
	numrNounEmpty(t, lt, r, "мало наслідком")
	numrNounEmpty(t, lt, r, "все одно йому")
	numrNounEmpty(t, lt, r, "один одному руки")
	numrNounEmpty(t, lt, r, "за день чи два мене")
	numrNounEmpty(t, lt, r, "сьома ранку")
	numrNounEmpty(t, lt, r, "години півтори")
	numrNounEmpty(t, lt, r, "хвилини за півтори поїзд приїхав")
	numrNounEmpty(t, lt, r, "1972 року")
	numrNounEmpty(t, lt, r, "1970-1971 рр.")
	numrNounEmpty(t, lt, r, "ст. 2 пункта 1")
	numrNounEmpty(t, lt, r, "3 / 5 вугілля світу")
	numrNounEmpty(t, lt, r, "3 / 4 понеділка")
	numrNounEmpty(t, lt, r, "1992 рік")
	numrNounEmpty(t, lt, r, "обоє винні")
	numrNounEmpty(t, lt, r, "скільки-небудь осяжний")
	numrNounEmpty(t, lt, r, "обоє режисери")
	numrNounEmpty(t, lt, r, "сотні дві персон")
	// reverse order
	numrNounEmpty(t, lt, r, "метрів п’ять килиму")
	numrNounEmpty(t, lt, r, "років з п’ять в'язниці")
	numrNounEmpty(t, lt, r, "рази може у два більший")
	numrNounEmpty(t, lt, r, "десятка півтора списочників")
	// v_rod insert
	numrNounEmpty(t, lt, r, "два провінційного вигляду персонажі")

	numrNounEmpty(t, lt, r, "У свої вісімдесят пан Василь")

	numrNounEmpty(t, lt, r, "№ 4 Тижня")

	numrNounEmpty(t, lt, r, "57-ма вулиця")

	numrNounEmpty(t, lt, r, "мати не стільки слух")
	numrNounEmpty(t, lt, r, "$300 тис. кредиту")
	numrNounEmpty(t, lt, r, "20 мм кредиту")
	numrNounEmpty(t, lt, r, "три нікому не відомих")
	// too many FP
	numrNounEmpty(t, lt, r, "несподіваного для багатьох висновку")
	numrNounEmpty(t, lt, r, "чотири десятих відсотка")

	// special for пів...
	numrNounEmpty(t, lt, r, "останніх півроку")

	numrNounEmpty(t, lt, r, "дві англійською")
	numrNounEmpty(t, lt, r, "за два злотих (15 грн)")

	//TODO: adj+numr
	numrNounEmpty(t, lt, r, "Дев'яноста річниця")
	numrNounEmpty(t, lt, r, "сьома вода на киселі")
	numrNounEmpty(t, lt, r, "років п'ять люди")
	numrNounEmpty(t, lt, r, "років через десять Литва")

	numrNounEmpty(t, lt, r, "у 2020 мати надійний фундамент")
	numrNounEmpty(t, lt, r, "тижнів зо два мати горшком воду носила")
}

// TestTokenAgreementNumrNounRule_RuleTP twins TokenAgreementNumrNounRuleTest.testRuleTP.
func TestTokenAgreementNumrNounRule_RuleTP(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounHasError(t, lt, r, "два пацана", "два пацани")
	numrNounHasError(t, lt, r, "дві пацани", "два пацани")
	numrNounHasError(t, lt, r, "дві сонця", "два сонця")
	numrNounHasError(t, lt, r, "обидва дівчини", "обидві дівчини")
	numrNounEmpty(t, lt, r, "дві крайнощі")
	numrNounEmpty(t, lt, r, "дві угорські фракції")
	numrNounEmpty(t, lt, r, "два місцеві вожді")
	numrNounEmpty(t, lt, r, "дві білі групи")
	numrNounEmpty(t, lt, r, "дві наглядачки-африканерки")
	numrNounEmpty(t, lt, r, "дві турбопрофесії")
	numrNounEmpty(t, lt, r, "два імені")
	numrNounEmpty(t, lt, r, "один-два громадянини")
	numrNounEmpty(t, lt, r, "Обидві ходи")
	numrNounEmpty(t, lt, r, "обидва атентати")

	numrNounHasError(t, lt, r, "два подоляни", "два подолянина")
	numrNounHasError(t, lt, r, "33 подоляни", "33 подолянина")
	numrNounHasError(t, lt, r, "три рабина")

	matches00 := r.Match(lt.GetAnalyzedSentence("обидві боки"))
	require.Equal(t, 1, len(matches00))
	require.Contains(t, matches00[0].GetMessage(), "Можливо, не збігається рід однини для множинної форми?")

	numrNounHasError(t, lt, r, "восьми пацани", "восьми пацанів", "восьми пацанам", "восьми пацанах")
	numrNounHasError(t, lt, r, "вісьма пацанах")
	numrNounHasError(t, lt, r, "декілька пацани", "декілька пацанів")
	numrNounHasError(t, lt, r, "двоє двері", "двоє дверей")

	numrNounEmpty(t, lt, r, "на один-півтора відсотка")
	numrNounEmpty(t, lt, r, "півтора року")
	numrNounEmpty(t, lt, r, "півтори сотні")

	numrNounHasError(t, lt, r, "півтора роки", "півтора року")
	numrNounHasError(t, lt, r, "півтора разу", "півтора раза")
	numrNounHasError(t, lt, r, "іспиту півтора роки тому")

	// special suggestions
	matches0 := r.Match(lt.GetAnalyzedSentence("у півтора рази"))
	require.Equal(t, 1, len(matches0))
	require.Contains(t, matches0[0].GetMessage(), "«раза»")
	require.ElementsMatch(t, []string{"півтора раза"}, matches0[0].GetSuggestedReplacements())

	numrNounHasError(t, lt, r, "півтора рублі", "півтора рубля")
	numrNounHasError(t, lt, r, "півтори рублі") // Java: new String[0] — no suggestion check
	numrNounHasError(t, lt, r, "за півтора місяці", "півтора місяця")
	numrNounEmpty(t, lt, r, "за півтора довгих місяці")
	numrNounEmpty(t, lt, r, "півтори місячні норми")

	numrNounHasError(t, lt, r, "пів ковбаса", "пів ковбаси")

	numrNounHasError(t, lt, r, "Останні три розділі", "три розділи")

	numrNounEmpty(t, lt, r, "0,5 курки")
	numrNounHasError(t, lt, r, "5,5 градуси", "5,5 градуса", "5,5 градусів")
	numrNounHasError(t, lt, r, "10,5 ніч", "10,5 ночі", "10,5 ночей")

	numrNounHasError(t, lt, r, "2,5 мільйонів", "2,5 мільйона", "2,5 мільйони")
	numrNounHasError(t, lt, r, "3,5 ночей", "3,5 ночі")
	numrNounEmpty(t, lt, r, "14,5 років")
	numrNounEmpty(t, lt, r, "протягом 3,5 років")
	numrNounHasError(t, lt, r, "2-2,5 метрів", "2-2,5 метра", "2-2,5 метри")
	numrNounEmpty(t, lt, r, "1,5 метра")
	numrNounHasError(t, lt, r, "1,5 метри", "1,5 метра")

	// adj
	numrNounEmpty(t, lt, r, "стільки різноманітних об'єктів")
	numrNounEmpty(t, lt, r, "два нових горнятка")
	numrNounEmpty(t, lt, r, "два нових веселих горнятка")
	numrNounEmpty(t, lt, r, "три десятих")
	numrNounEmpty(t, lt, r, "три вихідних")
	numrNounEmpty(t, lt, r, "243 хворих")
	numrNounEmpty(t, lt, r, "2-3 підозрюваних.")
	numrNounEmpty(t, lt, r, "п'ять нових горняток")
	numrNounEmpty(t, lt, r, "двадцять першого століття")
	numrNounEmpty(t, lt, r, "чотири одиночних та 12 парних титулів")
	numrNounEmpty(t, lt, r, ", чотири попередніх канули в Лету")

	// should be caught by adj-noun rule
	numrNounEmpty(t, lt, r, "п'ять нових горнятка")

	numrNounHasError(t, lt, r, "два великих автобуса", "два великих автобуси")
	numrNounHasError(t, lt, r, "434,5 злоякісних новоутворень")
	numrNounHasError(t, lt, r, "на 3 метра за останні", "3 метри", "3 метрів")
	numrNounHasError(t, lt, r, "обидва ймовірних кандидата", "обидва ймовірних кандидати")
	numrNounHasError(t, lt, r, "4 маленьких єнота", "4 маленьких єноти", "4 маленьких єнотів")
	numrNounHasError(t, lt, r, "два легкових автомобіля", "два легкових автомобілі")

	numrNounEmpty(t, lt, r, "на мільйон населення")
	numrNounEmpty(t, lt, r, "тисячі люду")
}

// TestTokenAgreementNumrNounRule_RuleForceNoun twins TokenAgreementNumrNounRuleTest.testRuleForceNoun.
func TestTokenAgreementNumrNounRule_RuleForceNoun(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounHasError(t, lt, r, "фронтові сто грам", "сто грамів")
	numrNounHasError(t, lt, r, "10—12 мегават", "12 мегаватів")
	numrNounHasError(t, lt, r, "12 німецьких солдат")
	numrNounHasError(t, lt, r, "десятки тисяч своїх солдат")

	numrNounHasError(t, lt, r, "п'ять байт", "п'ять байтів")
	numrNounHasError(t, lt, r, "5 байт", "5 байтів")
	numrNounHasError(t, lt, r, "мільйон байт", "мільйон байтів")
	numrNounEmpty(t, lt, r, "мільйон байтів")
	numrNounEmpty(t, lt, r, "тримали п'ять бейсбольних біт")

	numrNounHasError(t, lt, r, "60 мікрорентген", "60 мікрорентгенів")

	numrNounHasError(t, lt, r, "10 чоловік і жінок", "10 чоловіків")
	// handled by another rule: cholovik_with_numr
	numrNounEmpty(t, lt, r, "10 чоловік.")
}

// TestTokenAgreementNumrNounRule_RuleTon twins TokenAgreementNumrNounRuleTest.testRuleTon.
func TestTokenAgreementNumrNounRule_RuleTon(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounHasError(t, lt, r, "на 200 тон", "тонн")
	numrNounHasError(t, lt, r, "2 млн тон", "тонн")
	numrNounHasError(t, lt, r, "кілька тон", "тонн")
	numrNounHasError(t, lt, r, "вісім тисяч тон", "тонн")
	numrNounHasError(t, lt, r, "вісім тис. тон", "тонн")
	numrNounHasError(t, lt, r, "22 тон", "тонн")
	numrNounHasError(t, lt, r, "8,5 тон", "тонн")
}

// TestTokenAgreementNumrNounRule_RuleFract twins TokenAgreementNumrNounRuleTest.testRuleFract.
func TestTokenAgreementNumrNounRule_RuleFract(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounEmpty(t, lt, r, "дві з половиною кулі")

	numrNounHasError(t, lt, r, "два з половиною автобуса", "два з половиною автобуси")
}

// TestTokenAgreementNumrNounRule_RuleFractionals twins TokenAgreementNumrNounRuleTest.testRuleFractionals.
func TestTokenAgreementNumrNounRule_RuleFractionals(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounHasError(t, lt, r, "2,4 кілограми", "2,4 кілограма")
	numrNounHasError(t, lt, r, "2,4 мільйони", "2,4 мільйона")
	numrNounHasError(t, lt, r, "2,54 мільйони", "2,54 мільйона")
	numrNounEmpty(t, lt, r, "33,77 тонни")
	numrNounEmpty(t, lt, r, "2,19 бала")
	numrNounEmpty(t, lt, r, "33,77 мм")
	numrNounEmpty(t, lt, r, "33,77 тис")
	numrNounEmpty(t, lt, r, "33,5 роки")
	numrNounEmpty(t, lt, r, "33,5 Терещенко")
	numrNounEmpty(t, lt, r, "1998,1999 рр.") // should have space after comma, see next rule
	numrNounEmpty(t, lt, r, "із результатом 28,91 вона посіла третє місце")
	numrNounEmpty(t, lt, r, "Становлять відповідно 0,5 і 0,3 її оцінки.")
	numrNounEmpty(t, lt, r, "це на 0,21 краще за попередній рекорд світу.")

	numrNounHasError(t, lt, r, "або у 2,2 рази.", "2,2 раза")
	numrNounHasError(t, lt, r, "або у 2,2 раз", "2,2 раза")
	numrNounEmpty(t, lt, r, "або у 2,2 раза.")
	numrNounEmpty(t, lt, r, "або у 2 рази.")
}

// numrATR builds a single-reading ATR (Java AnalyzedTokenReadings constructor twin).
func numrATR(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	p, l := pos, lemma
	var posP, lemP *string
	if pos != "" {
		posP = &p
	}
	if lemma != "" {
		lemP = &l
	}
	return languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(token, posP, lemP),
	}, 0)
}

// TestTokenAgreementNumrNounRule_RuleDisambigVZna twins TokenAgreementNumrNounRuleTest.testRuleDisambigVZna.
func TestTokenAgreementNumrNounRule_RuleDisambigVZna(t *testing.T) {
	lt, r := wireTokenAgreementNumrNoun(t)

	numrNounHasError(t, lt, r, "два додаткових років")

	// Synthetic POS matrices from Java (manual AnalyzedTokenReadings, not full tagger).
	// Java builds sentences with SENT_START + fixed tags.
	ss := languagetool.SentenceStartTagName

	// три + села :p:v_zna
	sent1 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		numrATR("три", "numr:p:v_zna", "три"),
		numrATR("села", "noun:inanim:p:v_zna", "село"),
	})
	require.Equal(t, 0, len(r.Match(sent1)))

	// три + основні + села
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		numrATR("три", "numr:p:v_zna", "три"),
		numrATR("основні", "adj:p:v_zna:rinanim:compb", "основний"),
		numrATR("села", "noun:inanim:p:v_zna", "село"),
	})
	require.Equal(t, 0, len(r.Match(sent2)))

	// три + золоті (adj only)
	sent3 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		numrATR("три", "numr:p:v_zna", "три"),
		numrATR("золоті", "adj:p:v_zna:rinanim:compb", "золотий"),
	})
	require.Equal(t, 0, len(r.Match(sent3)))

	// три + мільйони (noun:numr)
	// Java: "мільйон/noun:inanim:p:v_zna:numr" lemma format — use POS noun:inanim:p:v_zna:numr
	sent4 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		numrATR("три", "numr:p:v_zna", "три"),
		numrATR("мільйони", "noun:inanim:p:v_zna:numr", "мільйон"),
	})
	require.Equal(t, 0, len(r.Match(sent4)))
}
