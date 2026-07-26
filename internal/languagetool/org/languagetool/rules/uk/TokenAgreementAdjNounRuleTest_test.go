package uk_test

// Twin of languagetool-language-modules/uk TokenAgreementAdjNounRuleTest
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new TokenAgreementAdjNounRule(TestTools.getMessages("uk"), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator + Ukrainian synthesizer (suggestions) +
// IgnoredCharacters (soft hyphen + combining acute).

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

// wireTokenAgreementAdjNoun ports JLanguageTool(Ukrainian) + TokenAgreementAdjNounRule ctor.
func wireTokenAgreementAdjNoun(t *testing.T) (*languagetool.JLanguageTool, *rulesuk.TokenAgreementAdjNounRule) {
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

	r := rulesuk.NewTokenAgreementAdjNounRule()
	// Java ukrainian.getSynthesizer() — official ukrainian_synth.dict + tags
	if sp := discoverUKResource(t, "ukrainian_synth.dict"); sp != "" {
		if base := synthesis.OpenBaseSynthesizerFromDictPath("uk", sp); base != nil {
			r.Synth = &synthuk.UkrainianSynthesizer{BaseSynthesizer: base}
		}
	}
	return lt, r
}

func adjNounEmpty(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementAdjNounRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 0, len(ms), "assertEmptyMatch %q got %d", text, len(ms))
}

func adjNounHasError(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementAdjNounRule, text string, suggestions ...string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 1, len(ms), "assertHasError %q got %d", text, len(ms))
	if len(suggestions) > 0 {
		require.Equal(t, suggestions, ms[0].GetSuggestedReplacements(), "suggestions for %q", text)
		for _, s := range suggestions {
			adjNounEmpty(t, lt, r, s)
		}
	}
}

// TestTokenAgreementAdjNounRule_Meta twins Java id / description / short.
func TestTokenAgreementAdjNounRule_Meta(t *testing.T) {
	r := rulesuk.NewTokenAgreementAdjNounRule()
	require.Equal(t, rulesuk.TokenAgreementAdjNounRuleID, r.GetID())
	require.Equal(t, "Узгодження відмінків, роду і числа прикметника та іменника", r.GetDescription())
	require.Equal(t, "Узгодження прикметника та іменника", r.GetShort())
}

// TestTokenAgreementAdjNounRule_RuleTP twins TokenAgreementAdjNounRuleTest.testRuleTP.
func TestTokenAgreementAdjNounRule_RuleTP(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounHasError(t, lt, r, "скрутна справі")
	adjNounHasError(t, lt, r, "район імпозантних віл")
	adjNounHasError(t, lt, r, "офіційний статистика")
	adjNounHasError(t, lt, r, "зелена яблуко.")
	adjNounHasError(t, lt, r, "сп’янілі свободою")
	adjNounHasError(t, lt, r, "кволий депутата")
	adjNounHasError(t, lt, r, "кволого тюльпан")
	adjNounHasError(t, lt, r, "цинічна винахідливості")
	adjNounHasError(t, lt, r, "наступній рік свого життя")
	adjNounHasError(t, lt, r, "жодного кубічного метру в Україні не буде")
	adjNounHasError(t, lt, r, "складний рік на фондовим ринку")
	adjNounHasError(t, lt, r, "є найкращий засобом для очистки")
	adjNounHasError(t, lt, r, "має вчену ступінь з хімії")
	adjNounHasError(t, lt, r, "розкішні чорні коси й, засукавши широкі рукава своєї сорочки,")
	adjNounHasError(t, lt, r, "то коло стола її вишивані рукава мають, то коло печі")
	adjNounHasError(t, lt, r, "валялись одірвані рукава кунтушів та поли жупанів")
	adjNounHasError(t, lt, r, "вчинили страшенний диплома тичний галас")
	adjNounHasError(t, lt, r, "Раймон Бенжамен і керівник європейського філії")
	adjNounHasError(t, lt, r, "обвинувачення у вчинені злочину, передбаченого")

	adjNounHasError(t, lt, r, "перша ступінь")
	adjNounHasError(t, lt, r, "друга ступінь")

	adjNounHasError(t, lt, r, "встановлена Верховної Радою 17 січня 2017 року")
	adjNounHasError(t, lt, r, "фракцію у Верховні Раді й мати")

	matches0 := r.Match(lt.GetAnalyzedSentence("4 російських винищувача"))
	require.Equal(t, 1, len(matches0))
	require.Contains(t, matches0[0].GetMessage(), "[ч.р.: родовий, знахідний]")
	require.Equal(t, []string{"російських винищувачів", "російського винищувача"}, matches0[0].GetSuggestedReplacements())

	// і-и
	adjNounHasError(t, lt, r, "під зеківській нуль")
	adjNounHasError(t, lt, r, "у повітряній простір")
	adjNounHasError(t, lt, r, "в дитячий лікарні")
	adjNounHasError(t, lt, r, "у Київський філармонії")
	adjNounHasError(t, lt, r, "на керівні посаді")
	adjNounHasError(t, lt, r, "незворотній процес")
	adjNounHasError(t, lt, r, "нинішній російські владі")
	adjNounHasError(t, lt, r, "заробітної платі")
	adjNounHasError(t, lt, r, "сталеві панцирі")
	adjNounHasError(t, lt, r, "У львівській та київський Книгарнях")
	// relies on disambiguation
	adjNounHasError(t, lt, r, "— робочій день.")
	adjNounEmpty(t, lt, r, "президентів Леонідів Кравчука та Кучму")

	adjNounHasError(t, lt, r, "угода на 30 років зі щорічного поставкою росіянами ...")

	// missing/extra letter
	adjNounHasError(t, lt, r, "сприймали так власні громадян")
	adjNounHasError(t, lt, r, "прокатні транспорті засоби")
	adjNounHasError(t, lt, r, "Ви не притягнене капіталу")
	adjNounHasError(t, lt, r, "потрібна змін поколінь")
	adjNounHasError(t, lt, r, "Незадовільне забезпеченням паливом")
	adjNounHasError(t, lt, r, "сочиться коштовний камінням")
	adjNounHasError(t, lt, r, "будь-якої демократичної крани")
	adjNounHasError(t, lt, r, "Генеральній прокураторі")
	adjNounHasError(t, lt, r, "була низька передумов")
	adjNounHasError(t, lt, r, "після смерті легендарного Фреді Меркюрі")
	adjNounHasError(t, lt, r, "юна викрадача й не здогадувалася")

	adjNounHasError(t, lt, r, "По повернені кореспондентів")
	adjNounHasError(t, lt, r, "в очікувані експериментатора")

	// wrong letter
	adjNounHasError(t, lt, r, "певної мірою")
	adjNounHasError(t, lt, r, "люмпенізується дедалі більша частини")
	adjNounHasError(t, lt, r, "Державна фіскальну служба")
	adjNounHasError(t, lt, r, "Як правило, це чоловіки, годувальними сімей")
	adjNounHasError(t, lt, r, "з московською боку")

	// wrong gender
	adjNounHasError(t, lt, r, "з насиджених барліг")
	adjNounHasError(t, lt, r, "У польському Лодзі")
	adjNounHasError(t, lt, r, "панування зимових сутінок")

	adjNounHasError(t, lt, r, "за наявною інформацію")
	adjNounHasError(t, lt, r, "асоціюється в нас із сучасною цивілізацію")

	adjNounHasError(t, lt, r, "зловживання монопольних становищем")
	adjNounHasError(t, lt, r, "проживання та дворазове харчуванням")

	// false v_rod with -у
	matches := r.Match(lt.GetAnalyzedSentence("кримського безсмертнику"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Можливо")

	// false кількисний числівник з v_rod
	matches = r.Match(lt.GetAnalyzedSentence("до 7-ми відсотків ВВП"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Можливо")
	require.Contains(t, matches[0].GetSuggestedReplacements(), "7 відсотків")

	matches = r.Match(lt.GetAnalyzedSentence("Та пахових ділянках"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Можливо")

	// false v_rod with -а
	adjNounHasError(t, lt, r, "федерального округа")

	// false :nv
	adjNounHasError(t, lt, r, "затверджений народним віче")
	adjNounHasError(t, lt, r, "На великому родинному віче")
	adjNounHasError(t, lt, r, "в листопаді 2015 року на народні віче до Кривого Рогу")
	adjNounHasError(t, lt, r, "як японські ніндзя")
	adjNounHasError(t, lt, r, "приталені пальто")

	// missing/extra space, dash etc
	adjNounHasError(t, lt, r, "соціальними мережа ми")
	adjNounHasError(t, lt, r, "принциповими країна ми")
	adjNounHasError(t, lt, r, "отримала мандатна ведення")
	adjNounHasError(t, lt, r, "по вуличному Копійчина")

	// lowercase city
	adjNounHasError(t, lt, r, "У мінську влада")

	// barbarism path: empty
	adjNounEmpty(t, lt, r, "на пострадянський манер")

	adjNounHasError(t, lt, r, "вздовж дніпровської вісі")
	adjNounHasError(t, lt, r, "ніщо так не зближає приморських партизан")
	adjNounHasError(t, lt, r, "як боротьбу сунітської більшість")

	adjNounHasError(t, lt, r, "найцікавішій час")

	adjNounHasError(t, lt, r, "відкинутий набагато років назад")

	adjNounHasError(t, lt, r, "символічною акцію стало складання")
	adjNounHasError(t, lt, r, "в будь–яким момент")
	adjNounHasError(t, lt, r, "хронічними порушення торговельних угод")

	// 148-ми
	adjNounHasError(t, lt, r, "вбивство 148-ми селян")
}

// TestTokenAgreementAdjNounRule_Rule twins TokenAgreementAdjNounRuleTest.testRule.
func TestTokenAgreementAdjNounRule_Rule(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	// correct sentences:
	adjNounEmpty(t, lt, r, "холодний яр")
	adjNounEmpty(t, lt, r, "страшне плацебо")
	adjNounEmpty(t, lt, r, "військової продукції")

	adjNounEmpty(t, lt, r, "Ім'я Мандели, дане йому при народженні")
	adjNounEmpty(t, lt, r, "Я не бачив сенсу в тому, щоб виклика́ти свідків і захищатися.")
	adjNounEmpty(t, lt, r, "погоджувальної комісії Інституту")
	adjNounEmpty(t, lt, r, "відштовхнути нового колегу.")
	adjNounEmpty(t, lt, r, "державну зраду.")
	adjNounEmpty(t, lt, r, "(Пізніше Вальтер став першим")
	adjNounEmpty(t, lt, r, "складовою успіху")
	adjNounEmpty(t, lt, r, "про екс-першого віце-спікера.")
	adjNounEmpty(t, lt, r, "Маю лишній квиток і подумав за свого найкращого друга")

	adjNounEmpty(t, lt, r, "Засвідчувана досить часто наукою «гнучкість» — один із коренів\n"+
		"паранаукових явищ на кшталт «нової хронології» Фоменка.")

	// skip adv after adjp
	adjNounEmpty(t, lt, r, "прикрита швидко рука")
	adjNounHasError(t, lt, r, "прикрита швидко руку")
	adjNounEmpty(t, lt, r, "прикрита отруйливо гарячим")
	adjNounEmpty(t, lt, r, "після короткого резюме справи")
	adjNounEmpty(t, lt, r, "білий як полотно")
	adjNounEmpty(t, lt, r, "розділеного вже чверть століття")
	adjNounEmpty(t, lt, r, "розділеного третину століття")
	adjNounEmpty(t, lt, r, "заклопотані чимало людей")
	adjNounEmpty(t, lt, r, "заклопотані дуже обмаль людей")
	adjNounEmpty(t, lt, r, "заданою відносно спостерігача")
	adjNounEmpty(t, lt, r, "опублікований увечері понеділка")

	adjNounEmpty(t, lt, r, "зареєстровані зокрема БЮТівець Микола Булатецький та самовисуванець")

	adjNounEmpty(t, lt, r, "схований всередині номера")

	adjNounEmpty(t, lt, r, "надана як раз")
	adjNounEmpty(t, lt, r, "підірваною ще раз")

	// nv no plural
	adjNounEmpty(t, lt, r, "спортивні ЦРУ")
	adjNounEmpty(t, lt, r, "Сумське НПО")
}

// TestTokenAgreementAdjNounRule_Exceptions twins testExceptions.
func TestTokenAgreementAdjNounRule_Exceptions(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "у могутні Максимові обійми")
	adjNounEmpty(t, lt, r, "надання болгарській статусу")
	adjNounEmpty(t, lt, r, "старший зміни")
	adjNounEmpty(t, lt, r, "Наступні півроку.")

	adjNounEmpty(t, lt, r, "одної шостої світу")
	adjNounEmpty(t, lt, r, "Кожному наглядач кивав")

	adjNounEmpty(t, lt, r, "чогось схожого Європа")
	adjNounEmpty(t, lt, r, "писав про щось подібне Юрій")

	adjNounEmpty(t, lt, r, "з 11-ма годинами")

	adjNounEmpty(t, lt, r, "ще чого доброго ціна впаде")

	adjNounEmpty(t, lt, r, "чотирициліндровий об’ємом 1000 куб. см.")
	adjNounEmpty(t, lt, r, "10 жовтих площею 1,5 ")
	adjNounEmpty(t, lt, r, "безплатні довжиною від 100 до 1000 метрів")

	adjNounEmpty(t, lt, r, "за метр кубічний води")

	adjNounEmpty(t, lt, r, "200% річних прибутку")

	adjNounEmpty(t, lt, r, "муніципальна плюс виробнича")

	adjNounEmpty(t, lt, r, "Головне центр правильно вибити")
	adjNounEmpty(t, lt, r, "вибране сучукрліту")

	adjNounEmpty(t, lt, r, "Менше народу – більше кисню")
	adjNounEmpty(t, lt, r, "Найчастіше випадки прямого підкупу")
	adjNounEmpty(t, lt, r, "– щонайперше олівець, простий, твердий")
	adjNounEmpty(t, lt, r, "Найбільше звинувачень у відьомстві")
	adjNounEmpty(t, lt, r, "— Раніше Україна неодноразово заявляла")

	adjNounEmpty(t, lt, r, "Від наступних пари десятків")
	adjNounEmpty(t, lt, r, "Суд визнав неконституційними низку положень")

	adjNounHasError(t, lt, r, "від наступних пари")

	adjNounEmpty(t, lt, r, "Північний Рейн-Вестфалія")
}

// TestTokenAgreementAdjNounRule_ExceptionsNumbers twins testExceptionsNumbers.
func TestTokenAgreementAdjNounRule_ExceptionsNumbers(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "33 народних обранці")
	adjNounEmpty(t, lt, r, "ще троє автокефальних єпископи")
	adjNounEmpty(t, lt, r, "два-три реальних кандидати")
	adjNounEmpty(t, lt, r, "три жовтих обірваних чоловіки")
	adjNounEmpty(t, lt, r, "обидва вітчизняних наукових ступені")
	adjNounEmpty(t, lt, r, "3-4 реально хворих депутати")
	adjNounEmpty(t, lt, r, "два–три колишніх кандидати")
	adjNounEmpty(t, lt, r, "два (чи навіть три) різних завершення роману")

	adjNounEmpty(t, lt, r, "два нових горнятка")
	adjNounEmpty(t, lt, r, "два жіночих імені")
	adjNounEmpty(t, lt, r, "два різних міста")
	adjNounEmpty(t, lt, r, "два абсолютно різних міста")

	adjNounEmpty(t, lt, r, "три предкові слов’янські племені")

	adjNounEmpty(t, lt, r, "два «круглих столи»")

	adjNounHasError(t, lt, r, "два високих депутат")
	adjNounHasError(t, lt, r, "дві високих дівчині")
	adjNounHasError(t, lt, r, "83,7 квадратних кілометра")

	// дріб
	adjNounEmpty(t, lt, r, "дві мільярдних метра")
	adjNounEmpty(t, lt, r, "п’ять шостих населення")
	adjNounEmpty(t, lt, r, "чотирьох п’ятих прибутку")

	adjNounHasError(t, lt, r, "п'ять шості світу")

	adjNounEmpty(t, lt, r, "1/8-ї фіналу")
	adjNounHasError(t, lt, r, "1/8-ї фіналом")

	adjNounEmpty(t, lt, r, "В одній другій українка здолала")
	adjNounEmpty(t, lt, r, "поступився в одній восьмій французу")

	adjNounEmpty(t, lt, r, "дві других дівчини")
	adjNounHasError(t, lt, r, "дві других дівчині")

	// 1–3-й класи
	adjNounEmpty(t, lt, r, "1–3-й класи поснідали й побігли")
	adjNounEmpty(t, lt, r, "у 5–8-му класах")
	adjNounEmpty(t, lt, r, "на сьомому–восьмому поверхах")
	adjNounEmpty(t, lt, r, "на 14—16-те місця")

	adjNounHasError(t, lt, r, "3-й класи поснідали")
	adjNounHasError(t, lt, r, "одному-два непоганих шанси")

	match := r.Match(lt.GetAnalyzedSentence("залишилося сиротами 22-є дітей"))
	require.Equal(t, 1, len(match))
	require.Contains(t, match[0].GetMessage(), "літерне нарощення після кількісного числівника")

	adjNounEmpty(t, lt, r, "восьмого – дев’ятого класів")
	adjNounEmpty(t, lt, r, "перший — дев’ятий класи")

	// time
	adjNounEmpty(t, lt, r, "і о 7-й ранку нас зустрічає")
	adjNounEmpty(t, lt, r, "Призначений на 11-ту похід")
	adjNounEmpty(t, lt, r, "о шостій ранку")
	adjNounEmpty(t, lt, r, "дванадцята дня")

	// years
	adjNounEmpty(t, lt, r, "Ставши 2003-го прем’єром")
	adjNounEmpty(t, lt, r, "У 1990-х скрута змусила")
	adjNounEmpty(t, lt, r, "за 2009-й відомство зобов’язало")
	adjNounEmpty(t, lt, r, "підвів риску під 2011-м програмою «ТОП-100»")
	adjNounEmpty(t, lt, r, "Лише в 1990-ті частину саду вдруге зробили доступною")
	adjNounEmpty(t, lt, r, "У 2009–2010-му дефіцит бюджету сягав близько 1/3 видатків")
	adjNounEmpty(t, lt, r, "в 1920–1930-х батько митця показав себе як український патріот")
	adjNounEmpty(t, lt, r, "за часів «конфронтації» 2008–2009-го квота на них зросла")
	adjNounEmpty(t, lt, r, "тільки за 1986–1988-й країна втратила близько 40 млрд крб")
	adjNounEmpty(t, lt, r, "На початку двотисячних режисер зустрів двох людей")
	adjNounEmpty(t, lt, r, "корифеї американської поезії 1950-60-х Лоренс Ферлінгетті")
	adjNounEmpty(t, lt, r, "На зламі 80-90-их функціонери ...")

	adjNounEmpty(t, lt, r, "щороку під Дев’яте травня")
	adjNounEmpty(t, lt, r, "з четвертого по одинадцяте липня")

	adjNounEmpty(t, lt, r, "замість звичного десятиліттями «Українського»")
	adjNounEmpty(t, lt, r, "природний тисячею років підтверджений")

	adjNounHasError(t, lt, r, "на 131-му хвилині")
	adjNounHasError(t, lt, r, "О 12–ї годині")

	adjNounEmpty(t, lt, r, "Анонсована тиждень тому домовленість")
	adjNounEmpty(t, lt, r, "забули про популярні пару років тому")

	adjNounEmpty(t, lt, r, "завищена разів у десять")
	adjNounEmpty(t, lt, r, "інвестиція на найближчі років п’ять")
	adjNounEmpty(t, lt, r, "до розташованого кілометрів за шість")
	adjNounEmpty(t, lt, r, "заповнені відсотків на 80")
	adjNounEmpty(t, lt, r, "лячно було перші хвилин 40")
	adjNounEmpty(t, lt, r, "і посаджений років на 10–15")
	adjNounEmpty(t, lt, r, "і піднятий відсотки на 3")
	adjNounEmpty(t, lt, r, "Поховавши році в п'ятдесятому жінку")

	adjNounEmpty(t, lt, r, "хвилини з 55-ї вірмени почали застосовувати пресинг")

	adjNounEmpty(t, lt, r, "один 5-а клас")
	adjNounEmpty(t, lt, r, "маршрутка номер 29-а фірми “Фіакр”")

	adjNounEmpty(t, lt, r, "на вул. Рубчака, 17-а Тарас Стецьків")
	adjNounEmpty(t, lt, r, "вулиці Володимира Великого, 35-а Юрій Борсук")
}

// TestTokenAgreementAdjNounRule_ExceptionsOther twins testExceptionsOther.
func TestTokenAgreementAdjNounRule_ExceptionsOther(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "Завдяки останнім бізнес")

	adjNounEmpty(t, lt, r, "порядок денний парламенту")
	adjNounEmpty(t, lt, r, "зокрема статтю 6-ту закону")
	adjNounEmpty(t, lt, r, "князівством Литовським подоляни")
	adjNounEmpty(t, lt, r, "абзац перший частини другої")
	adjNounEmpty(t, lt, r, "абзац другий частини першої")
	adjNounEmpty(t, lt, r, "частина четверта статті 53")

	adjNounEmpty(t, lt, r, "яких не мала рівних українка")

	adjNounEmpty(t, lt, r, "Київський імені Шевченка")
	adjNounEmpty(t, lt, r, "і колишня Маяковського")
	adjNounEmpty(t, lt, r, "Львівської ім. С. Крушельницької")
	adjNounEmpty(t, lt, r, "4-й Запорізький ім. гетьмана Б. Хмельницького")

	adjNounHasError(t, lt, r, "у Великій Вітчизняній Війн")
	adjNounHasError(t, lt, r, "у Великій Вітчизняній війна")
	adjNounEmpty(t, lt, r, "Після Великої Вітчизняної будівництво істотно розширилося")

	adjNounEmpty(t, lt, r, "польські зразка 1620—1650 років")

	adjNounEmpty(t, lt, r, "чинних станом на 4 червня")

	adjNounEmpty(t, lt, r, "а старший групи")

	adjNounEmpty(t, lt, r, "Не пасли задніх міліціонери")
	adjNounEmpty(t, lt, r, "сильних світу цього")
	adjNounEmpty(t, lt, r, "найвпливовіших світу сього")
	adjNounEmpty(t, lt, r, "усіх до єдиного")
	adjNounEmpty(t, lt, r, "усі до єдиного депутати")
	adjNounEmpty(t, lt, r, "Вольному воля")
	adjNounEmpty(t, lt, r, "порядку денного засідань")
	adjNounEmpty(t, lt, r, "лаву запасних партії")
	adjNounEmpty(t, lt, r, "викладатися на повну артисти")
	adjNounEmpty(t, lt, r, "молодшого гвардії сержанта")
	adjNounEmpty(t, lt, r, "постійно на рівних міністри, президенти")

	adjNounEmpty(t, lt, r, "під час Другої світової командири")

	adjNounEmpty(t, lt, r, "до слова Божого людей")
	adjNounEmpty(t, lt, r, "Різдва Христова вигнанець")
	adjNounEmpty(t, lt, r, "ведуча Першого Національного Марія Орлова")

	adjNounEmpty(t, lt, r, "дівоче Анна")

	adjNounEmpty(t, lt, r, "В середньому тривалість курсів для отримання")
	adjNounEmpty(t, lt, r, "в цілому результатом задоволені")

	adjNounEmpty(t, lt, r, "на червень поточного року $29,3 млрд")

	adjNounEmpty(t, lt, r, "Одним із перших бажання придбати")
	adjNounEmpty(t, lt, r, "Першими голодування оголосили депутати")
	adjNounEmpty(t, lt, r, "Перший людина проходить")
	adjNounEmpty(t, lt, r, "Перший митців принуджував")
	adjNounEmpty(t, lt, r, "вважаючи перших джерелом")

	adjNounEmpty(t, lt, r, "нічого протизаконного жінка не зробила")
	adjNounEmpty(t, lt, r, "Нічого подібного Сергій не казав")
	adjNounEmpty(t, lt, r, "Нічого поганого людям")
	adjNounEmpty(t, lt, r, "що нічим дієвим ініціативи не завершаться")

	adjNounHasError(t, lt, r, "на Західній України")

	adjNounEmpty(t, lt, r, "переконана психолог")

	adjNounEmpty(t, lt, r, "Серед присутніх Микола")
	adjNounEmpty(t, lt, r, "контраргументами рідних засновника структури")
	adjNounEmpty(t, lt, r, "спіймали на гарячому хабарників")

	adjNounEmpty(t, lt, r, "була б зовсім іншою динаміка")
	adjNounEmpty(t, lt, r, "була такою жорстокою політика")
	adjNounEmpty(t, lt, r, "стали архаїчними структури")
	adjNounEmpty(t, lt, r, "назвав винним Юрія")

	adjNounEmpty(t, lt, r, "відмінних від російської моделей")

	adjNounEmpty(t, lt, r, "не перевищував кількох десятих відсотка")

	adjNounEmpty(t, lt, r, "Береженого Бог береже")
}

// TestTokenAgreementAdjNounRule_ExceptionsPredic twins testExceptionsPredic.
func TestTokenAgreementAdjNounRule_ExceptionsPredic(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)
	adjNounEmpty(t, lt, r, "всі сумнівні слід викинути")
	adjNounEmpty(t, lt, r, "все зроблене слід обов'язково перетворити")
}

// TestTokenAgreementAdjNounRule_ExceptionsAdjp twins testExceptionsAdjp.
func TestTokenAgreementAdjNounRule_ExceptionsAdjp(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "обмежуючий власність")
	adjNounEmpty(t, lt, r, "створивший історію")

	adjNounEmpty(t, lt, r, "Помальована в біле кімната")
	adjNounEmpty(t, lt, r, "Помальована в усе біле кімната")

	adjNounHasError(t, lt, r, "помальований в біле кімната")

	adjNounEmpty(t, lt, r, "вкриті плющем будинки")
	adjNounEmpty(t, lt, r, "всі вкриті плющем")
	adjNounEmpty(t, lt, r, "оприлюднений депутатом Луценком")
	adjNounEmpty(t, lt, r, "щойно оголошених спікером як відсутніх")
	adjNounEmpty(t, lt, r, "групи захищені законом від образ")
	adjNounEmpty(t, lt, r, "змучений тягарем життя")
	adjNounEmpty(t, lt, r, "відправлені глядачами протягом 20 хвилин")
	adjNounEmpty(t, lt, r, "здивований запалом, який проступав")
	adjNounEmpty(t, lt, r, "охопленому насильством ваальському трикутнику")
	adjNounEmpty(t, lt, r, "переданих заповідником церкві")
	adjNounEmpty(t, lt, r, "більше занепокоєних захистом власних прав")

	adjNounEmpty(t, lt, r, "підсвічений синім діамант")

	adjNounEmpty(t, lt, r, "повторена тисячу разів")

	adjNounEmpty(t, lt, r, "підсвічений синім діамант")

	adjNounHasError(t, lt, r, "підсвічений синім діамантів")

	adjNounEmpty(t, lt, r, "Нав’язаний Австрії нейтралітет")
	adjNounEmpty(t, lt, r, "Нав’язаний Австрії коаліцією")
	adjNounEmpty(t, lt, r, "Наймилішою українцеві залишається бронза")
	adjNounEmpty(t, lt, r, "на цих загальновідомих американцям зразках")
	adjNounEmpty(t, lt, r, "слід бути свідомими необхідності")
	adjNounEmpty(t, lt, r, "влаштованою Мазепі Петром")
	adjNounEmpty(t, lt, r, "будуть вдячні державі Україна")
	adjNounEmpty(t, lt, r, "мають бути підпорядковані служінню чоловікові")
	adjNounEmpty(t, lt, r, "більше відомої загалу як")
	adjNounEmpty(t, lt, r, "одержимі суверенітетом")

	adjNounHasError(t, lt, r, "Нав’язаний Австрії нейтралітеті")
}

// TestTokenAgreementAdjNounRule_ExceptionsVerb twins testExceptionsVerb.
func TestTokenAgreementAdjNounRule_ExceptionsVerb(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "слід бути обережними туристам у горах")
	adjNounHasError(t, lt, r, "слід бути обережною туристам у горах")

	adjNounEmpty(t, lt, r, "зараз повинне ділом довести")
	adjNounEmpty(t, lt, r, "Вони здатні екскаватором переорювати")
	adjNounEmpty(t, lt, r, "яке готове матір рідну продати")
	adjNounEmpty(t, lt, r, "Через якийсь час був змушений академію покинути")

	adjNounEmpty(t, lt, r, "Досі була чинною заборона")
	adjNounEmpty(t, lt, r, "Досі була б чинною заборона")
	adjNounEmpty(t, lt, r, "Стає очевидною наявність")
	adjNounEmpty(t, lt, r, "було куди зрозумілішим гасло самостійності")
	adjNounEmpty(t, lt, r, "є очевидною війна")
	adjNounEmpty(t, lt, r, "була б такою ж суттєвою явка")
	adjNounEmpty(t, lt, r, "і була б дещо абсурдною ситуація.")
	adjNounEmpty(t, lt, r, "Стали дорожчими хліб чи бензин")

	adjNounEmpty(t, lt, r, "дівчат не залишила байдужими інформація")
	adjNounEmpty(t, lt, r, "визнали справедливою наставники обох команд")
	adjNounEmpty(t, lt, r, "визнало незаконною Міністерство юстиції")
	adjNounEmpty(t, lt, r, "яку роблять знаковою плями на мундирі")
	adjNounEmpty(t, lt, r, "видається цілком стабільною демократія")
	adjNounEmpty(t, lt, r, "може бути не ідеальною форма тістечок")
	adjNounEmpty(t, lt, r, "не можуть бути толерантними ізраїльтяни")

	adjNounEmpty(t, lt, r, "які зроблять неможливою ротацію влади")
	adjNounEmpty(t, lt, r, "зробити відкритим доступ")
	adjNounEmpty(t, lt, r, "визнають регіональними облради")
	adjNounEmpty(t, lt, r, "залишивши незруйнованим Карфаген")
	adjNounEmpty(t, lt, r, "зробить обтяжливим використання")
	adjNounEmpty(t, lt, r, "На сьогодні залишається невідомою доля близько 200 людей")

	adjNounEmpty(t, lt, r, "виявлено побитою Катю")
	adjNounEmpty(t, lt, r, "лишаючи порожньою клітку")
	adjNounEmpty(t, lt, r, "роблячи жорсткішими правила")
	adjNounEmpty(t, lt, r, "вважає повністю вірною постанову")
	adjNounEmpty(t, lt, r, "визнав протиправним і недійсним внесення")

	adjNounEmpty(t, lt, r, "зробило можливою і необхідною появу нового гравця")

	adjNounHasError(t, lt, r, "саме з такою тенденцію стикнулися")

	adjNounHasError(t, lt, r, "був продиктований глибоким розуміння")

	adjNounHasError(t, lt, r, "вважають нелегітимними анексію")
	adjNounHasError(t, lt, r, "не залишили байдужими адміністрацію")

	adjNounEmpty(t, lt, r, "визнання неконституційним закону")
	adjNounEmpty(t, lt, r, "визнання недійсним рішення")
	adjNounHasError(t, lt, r, "через визнання тут шкідливою орієнтацію на народну мову")

	adjNounHasError(t, lt, r, "визнання неконституційними закону")
	adjNounHasError(t, lt, r, "визнання недійсним рішенню")
}

// TestTokenAgreementAdjNounRule_ExceptionsAdj twins testExceptionsAdj.
func TestTokenAgreementAdjNounRule_ExceptionsAdj(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "жадібна землі")
	adjNounEmpty(t, lt, r, "вдячного батьку")
	adjNounEmpty(t, lt, r, "Я вдячний редакторові Вільяму Філліпсу")
	adjNounEmpty(t, lt, r, "радий присутності генерала")
	adjNounEmpty(t, lt, r, "відомий мешканцям")
	adjNounEmpty(t, lt, r, "менш незрозумілу киянам")
	adjNounEmpty(t, lt, r, "найстарший віком із нас")
	adjNounEmpty(t, lt, r, "таких немилих серцю Булгакова")
	adjNounEmpty(t, lt, r, "експозиція, присвячена Леоніду Іллічу")
	adjNounEmpty(t, lt, r, "печаткою та вручене платнику")

	adjNounEmpty(t, lt, r, "і кожна масою 10 кг")

	adjNounHasError(t, lt, r, "жадібна землею")

	adjNounEmpty(t, lt, r, "протилежний очікуваному результат")
	adjNounEmpty(t, lt, r, "альтернативну олігархічній модель")
	adjNounEmpty(t, lt, r, "альтернативні газовому варіанти")

	adjNounHasError(t, lt, r, "альтернативну олігархічній порядку")

	adjNounEmpty(t, lt, r, "найчисленнішими цеховики були саме в Грузії")
	adjNounEmpty(t, lt, r, "Дефіцитною торгівля США є")
	adjNounEmpty(t, lt, r, "Не менш виснажливою війна є і для ворога")
	adjNounEmpty(t, lt, r, "Найнижчою частка таких є на Півдні")
	adjNounEmpty(t, lt, r, "розвинутою Україну назвати важко")

	adjNounEmpty(t, lt, r, "кількість визнаних недійсними бюлетенів")

	adjNounEmpty(t, lt, r, "Слабшою критики вважають")
	adjNounEmpty(t, lt, r, "найбільш райдужною перспектива членства в ЄС залишається")

	adjNounEmpty(t, lt, r, "Однак безлюдним місто також не назвеш")
	adjNounEmpty(t, lt, r, "Вагомим експерти називають той факт")
	adjNounEmpty(t, lt, r, "такою ситуацію бачить сам")
	adjNounEmpty(t, lt, r, "таким піднесеним президента не бачили давно")

	adjNounEmpty(t, lt, r, "Найближчі півроку-рік")
	adjNounEmpty(t, lt, r, "найближчих тиждень-два")
	adjNounEmpty(t, lt, r, "протягом минулих травня-липня")
	adjNounEmpty(t, lt, r, "Перші рік-два влада відбивалася")
	adjNounEmpty(t, lt, r, "суперкризовими січнем–лютим")

	adjNounHasError(t, lt, r, "найближчі тиждень")
}

// TestTokenAgreementAdjNounRule_ExceptionsPrepAdj twins testExceptionsPrepAdj.
func TestTokenAgreementAdjNounRule_ExceptionsPrepAdj(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "діє подібний до попереднього закон")
	adjNounEmpty(t, lt, r, "з відмінним від їхнього набором цінностей")
	adjNounEmpty(t, lt, r, "Про далеку від взірцевої поведінку")
	adjNounEmpty(t, lt, r, "нижчими від ринкових цінами")
	adjNounEmpty(t, lt, r, "протилежний до загальнодержавного процес")
	adjNounEmpty(t, lt, r, "Схожої з тамтешньою концепції")
	adjNounEmpty(t, lt, r, "відрізнялася від нинішньої ситуація")
	adjNounEmpty(t, lt, r, "відрізнялася б від нинішньої ситуація")
	adjNounEmpty(t, lt, r, "відрізнялося від російського способом")

	adjNounHasError(t, lt, r, "асоціюється в нас із сучасною цивілізацію")

	adjNounEmpty(t, lt, r, "На відміну від європейських санкції США")
	adjNounEmpty(t, lt, r, "поряд з енергетичними Москва висувала")
	adjNounEmpty(t, lt, r, "тотожні із загальносоюзними герб і прапор")

	adjNounEmpty(t, lt, r, "чотири подібних до естонських звіти.")

	adjNounEmpty(t, lt, r, "порівняно з попереднім результат")
	adjNounEmpty(t, lt, r, "порівняно з російським рівень")

	adjNounHasError(t, lt, r, "він є одним із найстаріший амфітеатрів")
	adjNounHasError(t, lt, r, "подібний до попереднього закони")
	adjNounHasError(t, lt, r, "порівняний з попереднім результатів")
	adjNounHasError(t, lt, r, "Схожої з тамтешньою концепція")
	adjNounHasError(t, lt, r, "протилежний до загальнодержавному процес")
	adjNounHasError(t, lt, r, "вдалися до збройної боротьбі")
}

// TestTokenAgreementAdjNounRule_ExceptionsPlural twins testExceptionsPlural.
func TestTokenAgreementAdjNounRule_ExceptionsPlural(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "на довгих півстоліття")
	adjNounEmpty(t, lt, r, "цілих півмісяця")
	adjNounEmpty(t, lt, r, "на довгих чверть століття")

	adjNounHasError(t, lt, r, "на довгих місяця")

	adjNounHasError(t, lt, r, "продукту, а просунути український фільми")

	adjNounEmpty(t, lt, r, "щоб моїх маму й сестер")
	adjNounEmpty(t, lt, r, "власними потом i кров’ю")
	adjNounEmpty(t, lt, r, "Перші тиждень чи два")
	adjNounEmpty(t, lt, r, "зазначені ім'я, прізвище та місто")
	adjNounEmpty(t, lt, r, "Житомирська, Кіровоградська області")
	adjNounEmpty(t, lt, r, "ані судова, ані правоохоронна системи")
	adjNounEmpty(t, lt, r, "шиїтську та сунітську, а також курдську частини")
	adjNounEmpty(t, lt, r, "Чорного і Азовського морів")
	adjNounEmpty(t, lt, r, "називає й традиційні корупцію, «відкати», хабарі")
	adjNounEmpty(t, lt, r, "державні Ощадбанк, «Укргазбанк»")
	adjNounEmpty(t, lt, r, "коринфський з іонійським ордери")
	adjNounEmpty(t, lt, r, "від однієї й другої сторін")
	adjNounEmpty(t, lt, r, "під’їздить один, другий автобуси")
	adjNounEmpty(t, lt, r, "зв'язаних ченця з черницею")
	adjNounEmpty(t, lt, r, "на зарубаних матір з двома синами")
	adjNounEmpty(t, lt, r, "повоєнні Австрія з Фінляндією")
	adjNounEmpty(t, lt, r, "Опозиційні Андрієвський і Черников")
	adjNounEmpty(t, lt, r, "директори навчальної та середньої шкіл")
	adjNounEmpty(t, lt, r, "протягом минулих травня – липня")
	adjNounEmpty(t, lt, r, "практично відсутні транспорт, гомінкі базари")
	adjNounEmpty(t, lt, r, "ВАЖЛИВІ МОТОРИКА І ВІДЧУТТЯ РІВНОВАГИ")
	adjNounEmpty(t, lt, r, "канонізовані Іоанн XXIII і Іван Павло II")
	adjNounEmpty(t, lt, r, "у дво- й тривимірному форматах")
	adjNounEmpty(t, lt, r, "Однак ні паровий, ні електричний двигуни не могли")
	adjNounEmpty(t, lt, r, "сміттєпереробного і/або сміттєспалювального заводів")
	adjNounEmpty(t, lt, r, "130-те (мінус вісім позицій порівняно з 2009-м) та 145-те місця")
	adjNounEmpty(t, lt, r, "ні у методологічному, ні у практичному аспектах.")
	adjNounEmpty(t, lt, r, "Хоч в англомовній, хоч в україномовній версіях")

	adjNounEmpty(t, lt, r, "Большого та Маріїнського театрів")
	adjNounEmpty(t, lt, r, "Пляжі 3, 4 і 5-ї категорій.")

	adjNounHasError(t, lt, r, "У львівській та київський Книгарнях")
	adjNounHasError(t, lt, r, "налякані Австрія з Фінляндію")

	adjNounHasError(t, lt, r, "Судячи з січневих продаж, 2009-й може стати")
	adjNounHasError(t, lt, r, "які наполягали на введені санкцій, будуть продовжуватися.")

	adjNounEmpty(t, lt, r, "символізують творчий, оберігальний та руйнівний аспекти Вищої Сили")
	adjNounEmpty(t, lt, r, "на місцевому, так і на центральному рівнях")
	adjNounEmpty(t, lt, r, "передався повоєнним Відню та Парижу")

	adjNounEmpty(t, lt, r, "найхарактерніші лояльність до влади й відданість місцевим лідерам.")
	adjNounEmpty(t, lt, r, "і з першим, і з другим чоловіками")
	adjNounEmpty(t, lt, r, "молодші Олександр Ірванець, Оксана Луцишина, Євгенія Кононенко")
	adjNounEmpty(t, lt, r, "230 вчилися за старшинською і 120 за підстаршинською програмами")
}

// TestTokenAgreementAdjNounRule_ExceptionsPluralConjAdv twins testExceptionsPluralConjAdv.
func TestTokenAgreementAdjNounRule_ExceptionsPluralConjAdv(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "уражені штаб ІДІЛ, а також збройний завод.")

	adjNounEmpty(t, lt, r, "в соціальному, а згодом і в економічному аспектах")
	adjNounEmpty(t, lt, r, "до апеляційного, а відтак і до конституційного судів")
	adjNounEmpty(t, lt, r, "У переносному та навіть у прямому сенсах слова")
	adjNounEmpty(t, lt, r, "в Чернівецькій і частково у Закарпатській областях")
	adjNounEmpty(t, lt, r, "парламентської, а згодом і президентської кампаній")
	adjNounEmpty(t, lt, r, "на західноєвропейському, а потім і на американському ринках")
	adjNounEmpty(t, lt, r, "навчався в реальному, потім у землемірному училищах")

	adjNounHasError(t, lt, r, "для того, щоб пожвавити культурне середовища села")
}

// TestTokenAgreementAdjNounRule_ExceptionsInsertPhrase twins testExceptionsInsertPhrase.
func TestTokenAgreementAdjNounRule_ExceptionsInsertPhrase(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "трагедію російського й, особливо, українського народів")
	adjNounEmpty(t, lt, r, "Китай і, певною мірою, Росія зуміли поставити")

	adjNounHasError(t, lt, r, "що, однак, не змінюють загальної картин")

	adjNounHasError(t, lt, r, "про те, що в різних міста")
	adjNounHasError(t, lt, r, "Він додав, що у комунальних підприємства")
	adjNounHasError(t, lt, r, "наставник сказав, що на світовий першості")
	adjNounHasError(t, lt, r, "Аль-Каїда, і чи не найстрашніше терористичний об’єднання")
	adjNounHasError(t, lt, r, "буде очікувати, коли нова редакції")
	adjNounHasError(t, lt, r, "вважає Порошенко, одночасно закликаючи європейську коаліції")
	adjNounHasError(t, lt, r, "пішли вперед, хай і міліметровим кроками")
	adjNounHasError(t, lt, r, "Словом, у проблематиці подвійного громадянств")
	adjNounHasError(t, lt, r, "Правоохоронцями, зокрема, проведена масштабна операції")
	adjNounHasError(t, lt, r, "Він так і не прийняв односторонню капітуляції")
	adjNounHasError(t, lt, r, "Так, наприклад, косівським аматорами")
	adjNounHasError(t, lt, r, "Якщо третина чи навіть половинна населення")
	adjNounHasError(t, lt, r, "Думаю, це фейковий вкидання")
	adjNounHasError(t, lt, r, "Сьогодні, наприклад, часта машинобудування")

	adjNounEmpty(t, lt, r, "латиську і, здається, молдавську поезії")
	adjNounEmpty(t, lt, r, "в Житомирській чи, скажімо, Миколаївській областях")
}

// TestTokenAgreementAdjNounRule_Pronouns twins testPronouns (@Test, not @Ignore).
func TestTokenAgreementAdjNounRule_Pronouns(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	// pron
	adjNounEmpty(t, lt, r, "усі решта")
	adjNounEmpty(t, lt, r, "єдину для всіх схему")

	adjNounEmpty(t, lt, r, "без таких документів")
	adjNounEmpty(t, lt, r, "згідно з якими африканцям")
	adjNounEmpty(t, lt, r, "чиновників, яким доступ")

	adjNounEmpty(t, lt, r, "так само")
	adjNounEmpty(t, lt, r, "перед тим гарант")
	adjNounEmpty(t, lt, r, "усього місяць тому")

	// this, that...
	adjNounEmpty(t, lt, r, "це мова сото")
	adjNounEmpty(t, lt, r, "без якої сім’я не проживе")

	adjNounEmpty(t, lt, r, "ВО «Свобода», лідер котрої Олег Тягрибок")

	adjNounEmpty(t, lt, r, "стільки само свідків")

	adjNounEmpty(t, lt, r, "що таке звук?")

	adjNounEmpty(t, lt, r, "обстрілює один за одним охоронців.")

	adjNounEmpty(t, lt, r, "повів сам військо")
	adjNounEmpty(t, lt, r, "що ж таке геноцид")
	adjNounEmpty(t, lt, r, "що воно таке еліта")
	adjNounEmpty(t, lt, r, "Таких меншість.")

	adjNounEmpty(t, lt, r, "той родом з Білорусі")
	adjNounEmpty(t, lt, r, "в нашу Богом забуту Данину")

	adjNounEmpty(t, lt, r, "той кібернетикою займається")

	adjNounEmpty(t, lt, r, "такого світ ще не бачив")
	// Java: // assertEmptyMatch("на таке Євросоюз напевне не наважиться");
	adjNounHasError(t, lt, r, "позицію у такій спосіб, — ділиться думками")

	// Java commented TODOs omitted:
	// // assertEmptyMatch("до такого студенти не готові");
	// // assertEquals(1, rule.match(...("із такою самого зневагою"))...);
	// // assertHasError("на вибори само висуванцем");
}

// TestTokenAgreementAdjNounRule_SpecialChars twins testSpecialChars.
func TestTokenAgreementAdjNounRule_SpecialChars(t *testing.T) {
	lt, r := wireTokenAgreementAdjNoun(t)

	adjNounEmpty(t, lt, r, "зелений поді\u00ADум")
	adjNounHasError(t, lt, r, "зелений по\u00ADділка.")
	adjNounHasError(t, lt, r, "зе\u00ADлений поділка.")
}
