package uk_test

// Twin of languagetool-language-modules/uk TokenAgreementVerbNounRuleTest
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new TokenAgreementVerbNounRule(TestTools.getMessages("uk"), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator + Ukrainian synthesizer (suggestions) +
// IgnoredCharacters (soft hyphen + combining acute).

import (
	"strings"
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

// wireTokenAgreementVerbNoun ports JLanguageTool(Ukrainian) + TokenAgreementVerbNounRule ctor.
func wireTokenAgreementVerbNoun(t *testing.T) (*languagetool.JLanguageTool, *rulesuk.TokenAgreementVerbNounRule) {
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

	r := rulesuk.NewTokenAgreementVerbNounRule()
	// Java ukrainian.getSynthesizer() — official ukrainian_synth.dict + tags
	if sp := discoverUKResource(t, "ukrainian_synth.dict"); sp != "" {
		if base := synthesis.OpenBaseSynthesizerFromDictPath("uk", sp); base != nil {
			r.Synth = &synthuk.UkrainianSynthesizer{BaseSynthesizer: base}
		}
	}
	return lt, r
}

func verbNounEmpty(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementVerbNounRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 0, len(ms), "assertEmptyMatch %q got %d matches: %v", text, len(ms), ms)
}

func verbNounHasError(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementVerbNounRule, text string, suggestions ...string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 1, len(ms), "assertHasError %q got %d matches: %v", text, len(ms), ms)
	if len(suggestions) > 0 {
		// Suggestion order can follow synthesizer reading order; assert same multiset (Java-visible).
		require.ElementsMatch(t, suggestions, ms[0].GetSuggestedReplacements(), "suggestions for %q", text)
		for _, s := range suggestions {
			verbNounEmpty(t, lt, r, s)
		}
	}
}

func verbNounMatches(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementVerbNounRule, n int, text string, msgContains string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, n, len(ms), "assertMatches(%d) %q got %d: %v", n, text, len(ms), ms)
	if msgContains != "" {
		for _, m := range ms {
			require.True(t, strings.Contains(m.GetMessage(), msgContains), "msg %q should contain %q", m.GetMessage(), msgContains)
		}
	}
}

// TestTokenAgreementVerbNounRule_Meta twins Java id / description / short.
func TestTokenAgreementVerbNounRule_Meta(t *testing.T) {
	r := rulesuk.NewTokenAgreementVerbNounRule()
	require.Equal(t, rulesuk.TokenAgreementVerbNounRuleID, r.GetID())
	require.Equal(t, "Узгодження дієслова з іменником", r.GetDescription())
	require.Equal(t, "Узгодження дієслова з іменником", r.GetShort())
}

// TestTokenAgreementVerbNounRule_RuleTPSuggestions twins TokenAgreementVerbNounRuleTest.testRuleTPSuggestions.
func TestTokenAgreementVerbNounRule_RuleTPSuggestions(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	// explicit span + suggestions (Java testRuleTPSuggestions)
	// Java start/end are character offsets (same as ATR GetStartPos/GetEndPos here).
	text := "перераховувати причин"
	sent := lt.GetAnalyzedSentence(text)
	ms := r.Match(sent)
	require.Equal(t, 1, len(ms))
	runes := []rune(text)
	require.Equal(t, "перераховувати причин", string(runes[ms[0].GetFromPos():ms[0].GetToPos()]))
	require.ElementsMatch(t, []string{"перераховувати причинам", "перераховувати причинами", "перераховувати причини"}, ms[0].GetSuggestedReplacements())
	verbNounHasError(t, lt, r, "зазнавати глибоке", "зазнавати глибокого")
}

// TestTokenAgreementVerbNounRule_RuleTP twins TokenAgreementVerbNounRuleTest.testRuleTP.
func TestTokenAgreementVerbNounRule_RuleTP(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounMatches(t, lt, r, 1, "вибирався Києві", "орудний")
	verbNounHasError(t, lt, r, "вповільнятися підлоги")
	verbNounHasError(t, lt, r, "вповільнятися дерев'яні підлоги")
	verbNounHasError(t, lt, r, "встановити електроні датчики")
	verbNounHasError(t, lt, r, "пройтися трьом книгам")
	verbNounHasError(t, lt, r, "Існує Західноєвропейській союз")
	verbNounHasError(t, lt, r, "почався справжнісінькій абстинентний синдром")
	verbNounHasError(t, lt, r, "досягнув піку")
	verbNounHasError(t, lt, r, "сягне піку")
	verbNounHasError(t, lt, r, "сягнувши піку")
	verbNounHasError(t, lt, r, "побачив озброєнні формування")
	verbNounHasError(t, lt, r, "втрапили халепу")
	verbNounHasError(t, lt, r, "доведено світовім досвідом")
	verbNounHasError(t, lt, r, "боятися закордоном")
	verbNounHasError(t, lt, r, "з точки зори антилатинської")
	verbNounHasError(t, lt, r, "не вірить свої очам")
	verbNounHasError(t, lt, r, "зменшити впив країні")
	verbNounHasError(t, lt, r, "займатися модернізацію закладів")
	verbNounHasError(t, lt, r, "вважать засобами масової інформації")
	verbNounHasError(t, lt, r, "вийшли фотку")
	verbNounHasError(t, lt, r, "визнання догорів недійсними")
	verbNounHasError(t, lt, r, "що купуються нову техніку")
	verbNounHasError(t, lt, r, "наживаються дільці")
	verbNounHasError(t, lt, r, "зусилля наблизити перемовин")
	verbNounHasError(t, lt, r, "витіснено протестувальники")
	verbNounHasError(t, lt, r, "фіктивних догорів оренди")
	verbNounHasError(t, lt, r, "Крим залишається територію України")
	verbNounHasError(t, lt, r, "Навожу відомі слова")
	verbNounHasError(t, lt, r, "розвиток рикну житла")
	verbNounHasError(t, lt, r, "покращення епідемічного станув країні")
	verbNounHasError(t, lt, r, "вчасно приходить всі профілактичні огляди")
	verbNounHasError(t, lt, r, "із наймеш благополучної частини")
	verbNounHasError(t, lt, r, "почуватиметься вправі")
	verbNounHasError(t, lt, r, "важко уявити  країн Балтії")
	verbNounHasError(t, lt, r, "користується попи том")
	verbNounHasError(t, lt, r, "ще більше погрішать ситуацію")
	verbNounHasError(t, lt, r, "стався також вибув метану")
	verbNounHasError(t, lt, r, "від спати єдиного податку")
	verbNounHasError(t, lt, r, "прочитати сааме цю книжку")
	verbNounHasError(t, lt, r, "сиплять дуст", "сипати дуст")
	verbNounHasError(t, lt, r, "Охочих навчитися цьому ремеслу")
	verbNounHasError(t, lt, r, "поступилася португальці")
	verbNounHasError(t, lt, r, "прориваються скрізь стару стріху")
	verbNounHasError(t, lt, r, "за будь яку ціну")
	verbNounHasError(t, lt, r, "також було велика непевність")
	verbNounHasError(t, lt, r, "мелять їхні українські папуги")
	verbNounEmpty(t, lt, r, "сидять понад міру")
	verbNounEmpty(t, lt, r, "стало відомо точну дату")
	verbNounEmpty(t, lt, r, "стало видно тільки оранжеві щоки")
}

// TestTokenAgreementVerbNounRule_RuleWithPart twins TokenAgreementVerbNounRuleTest.testRuleWithPart.
func TestTokenAgreementVerbNounRule_RuleWithPart(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounHasError(t, lt, r, "сягнув би піку", "сягнув би піка")
	verbNounHasError(t, lt, r, "це не відбувається не державному рівні")
	verbNounEmpty(t, lt, r, "Вінні прощалася й охоронці виводили їх")
	verbNounEmpty(t, lt, r, "Але працювати не те щоб складно")
	verbNounEmpty(t, lt, r, "не піднявся не те що Київ")
	verbNounEmpty(t, lt, r, "не піднявся не те, що Київ")
}

// TestTokenAgreementVerbNounRule_RuleTN twins TokenAgreementVerbNounRuleTest.testRuleTN.
func TestTokenAgreementVerbNounRule_RuleTN(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "вкрадено державою")
	verbNounEmpty(t, lt, r, "я буду каву")
	verbNounEmpty(t, lt, r, "спроєктувати проект")
	verbNounEmpty(t, lt, r, "купив книгу")
	verbNounEmpty(t, lt, r, "позбавляло людину")
	verbNounEmpty(t, lt, r, "залишати межі")
	verbNounEmpty(t, lt, r, "залишаються джерелом")
	verbNounEmpty(t, lt, r, "відкидати атрибути")
	verbNounEmpty(t, lt, r, "належати людині")
	verbNounEmpty(t, lt, r, "належати руку")
	verbNounEmpty(t, lt, r, "забракло вмінь")
	verbNounEmpty(t, lt, r, "Може поменшати ентузіазму")
	verbNounEmpty(t, lt, r, "Може видатися парадоксальним твердження")
	verbNounEmpty(t, lt, r, "розподілятиметься пропорційно вкладеній праці")
	verbNounEmpty(t, lt, r, "зростатимуть прямо пропорційно україноненависництву")
	verbNounEmpty(t, lt, r, "рухатися далі власної історії")
	verbNounEmpty(t, lt, r, "запропоновано урядом")
	verbNounEmpty(t, lt, r, "мало часу")
	verbNounEmpty(t, lt, r, "з’явитися перед ним")
	verbNounEmpty(t, lt, r, "не було меблів")
	verbNounEmpty(t, lt, r, "оплачуватися повинна відповідно")
	verbNounEmpty(t, lt, r, "відрізнятись один від одного")
	verbNounEmpty(t, lt, r, "співпрацювати один із одним")
	verbNounEmpty(t, lt, r, "допомагати одне одному")
	verbNounEmpty(t, lt, r, "залучити інвестицій на 20—30 мільйонів")
	verbNounEmpty(t, lt, r, "збереться людей зо 200")
	verbNounEmpty(t, lt, r, "як боротися підприємцям")
	verbNounEmpty(t, lt, r, "захиститися неспроможні")
	verbNounEmpty(t, lt, r, "висміювати такого роду забобони")
	verbNounEmpty(t, lt, r, "вважається свого роду психологічним")
	verbNounEmpty(t, lt, r, "хто мітингуватиме таким чином")
	verbNounEmpty(t, lt, r, "уникнувши тим самим")
	verbNounEmpty(t, lt, r, "міг хитрістю змусити")
	verbNounEmpty(t, lt, r, "могли займатися структури")
	verbNounEmpty(t, lt, r, "могли б займатися структури")
	verbNounEmpty(t, lt, r, "мусить чимось перекрити")
	verbNounEmpty(t, lt, r, "довелося її розбирати")
	verbNounEmpty(t, lt, r, "повинні існувати такі")
	verbNounEmpty(t, lt, r, "робити здатна")
	verbNounEmpty(t, lt, r, "боротиметься кілька однопартійців")
	verbNounEmpty(t, lt, r, "вийшла 1987-го")
	verbNounEmpty(t, lt, r, "зігріває його серце")
	verbNounEmpty(t, lt, r, "дай Боже")
	verbNounEmpty(t, lt, r, "спробуймо йому відповісти")
	verbNounEmpty(t, lt, r, "прокинься Тарасе")
	verbNounEmpty(t, lt, r, "див. новинні")
	verbNounEmpty(t, lt, r, "повинен відбудватися процес")
	verbNounEmpty(t, lt, r, "виявилося нікому не потрібним")
	verbNounEmpty(t, lt, r, "покататися самому")
	verbNounEmpty(t, lt, r, "віддати-відрізати Донбас")
	verbNounEmpty(t, lt, r, "вона називалася Оперативний злам")
	verbNounEmpty(t, lt, r, "мене звати Ігор")
	verbNounEmpty(t, lt, r, "підклали дров")
	verbNounEmpty(t, lt, r, "наголосила політик")
	verbNounEmpty(t, lt, r, "тривав довгих десять раундів")
	verbNounEmpty(t, lt, r, "лежали всю дорогу")
	verbNounEmpty(t, lt, r, "мріяли все життя")
	verbNounEmpty(t, lt, r, "сподобалося гуляти")
	verbNounEmpty(t, lt, r, "належить пройтися")
	verbNounEmpty(t, lt, r, "стали каламутного кольору")
	verbNounEmpty(t, lt, r, "ні сіло ні впало Комітет держбезпеки")
	verbNounEmpty(t, lt, r, "звичайна, якщо не сказати слабка, людина")
	verbNounEmpty(t, lt, r, "займаючись кожен своїми справами")
	verbNounEmpty(t, lt, r, "підстрахуватися не зайве")
	verbNounEmpty(t, lt, r, "поназбиравши купу обіцянок")
	verbNounEmpty(t, lt, r, "пішов світ за очі")
	verbNounEmpty(t, lt, r, "тікати куди очі бачать")
	verbNounEmpty(t, lt, r, "Видно було також великий")
	verbNounEmpty(t, lt, r, "буде видно тільки супутники")
	verbNounEmpty(t, lt, r, "чути було лише стукіт")
	verbNounEmpty(t, lt, r, "було дуже чутно стукіт")
	verbNounEmpty(t, lt, r, "зеленій частині")
	verbNounEmpty(t, lt, r, "будь то буряки чи малина")
	verbNounEmpty(t, lt, r, "поліклініці не було де амбулаторні")
	verbNounEmpty(t, lt, r, "чесніше було б державний фонд")
	verbNounEmpty(t, lt, r, "Процвітати буде лише бізнес")
	verbNounEmpty(t, lt, r, "були б іншої думки")
	verbNounEmpty(t, lt, r, "має своїм неодмінним наслідком")
	verbNounEmpty(t, lt, r, "що є сил")
	verbNounEmpty(t, lt, r, "Конкурс був десь шість")
	verbNounEmpty(t, lt, r, "повна була образів")
	verbNounEmpty(t, lt, r, "відкрито було журнал")
	verbNounEmpty(t, lt, r, "треба буде ще склянку")
}

// TestTokenAgreementVerbNounRule_RuleTnVdav twins TokenAgreementVerbNounRuleTest.testRuleTnVdav.
func TestTokenAgreementVerbNounRule_RuleTnVdav(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "Не бачити вам цирку")
	verbNounEmpty(t, lt, r, "розсміявся йому в обличчя")
	verbNounEmpty(t, lt, r, "закружляли мені десь у тьмі")
	verbNounEmpty(t, lt, r, "ірже вам у вічі")
	verbNounEmpty(t, lt, r, "умоститися господареві на рамена")
	verbNounEmpty(t, lt, r, "прилетів йому від посла")
	verbNounEmpty(t, lt, r, "пробирається людині під шкіру")
	verbNounEmpty(t, lt, r, "їхав їй назустріч")
	verbNounEmpty(t, lt, r, "біжать йому навперейми")
	verbNounHasError(t, lt, r, "Фабрика Миколая вперше запрацювала Львові у 2001 році")
	verbNounEmpty(t, lt, r, "Квапитися їй нікуди")
	verbNounEmpty(t, lt, r, "хворіти їй ніколи")
	verbNounEmpty(t, lt, r, "Жити родині нема де.")
	verbNounEmpty(t, lt, r, "Евакуюватися нам не було куди")
	verbNounEmpty(t, lt, r, "нічим пишатися жителям")
	verbNounEmpty(t, lt, r, "куди подітися селянам")
	verbNounEmpty(t, lt, r, "у таких будиночках жити мешканцям")
	verbNounEmpty(t, lt, r, "як тепер жити нам")
	verbNounEmpty(t, lt, r, "сидіти б панові")
	verbNounEmpty(t, lt, r, "не було де амбулаторні карточки ставити")
}

// TestTokenAgreementVerbNounRule_RuleTn_V_N_Vinf twins TokenAgreementVerbNounRuleTest.testRuleTn_V_N_Vinf.
func TestTokenAgreementVerbNounRule_RuleTn_V_N_Vinf(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "маю тобі щось підказати")
	verbNounEmpty(t, lt, r, "вміємо цим зазвичай користуватися")
	verbNounHasError(t, lt, r, "вміємо цьому зазвичай користуватися")
	verbNounEmpty(t, lt, r, "вони воліли мені якнайбільш ефективно допомогти")
	verbNounEmpty(t, lt, r, "воліли заворушень не допускати")
	verbNounEmpty(t, lt, r, "не втомлюються десятиріччями боротися Берлін і Венеція")
	verbNounEmpty(t, lt, r, "постарається таку двозначність усунути")
	verbNounEmpty(t, lt, r, "розпорядився частину зарплати примусово видавати")
	verbNounEmpty(t, lt, r, "не схотіли нам про це казати")
	verbNounEmpty(t, lt, r, "Самі респонденти пояснити причин не можуть")
	verbNounEmpty(t, lt, r, "довелося план «Б» застосовувати")
	verbNounEmpty(t, lt, r, "сподіваючися щось розтлумачити")
	verbNounEmpty(t, lt, r, "Резюмуючи політик наголосив")
	verbNounEmpty(t, lt, r, "пригадує посміхаючись Аскольд")
	verbNounHasError(t, lt, r, "неопізнаний літаючи об’єкт")
	verbNounHasError(t, lt, r, "знищила існуючи бази даних")
	verbNounEmpty(t, lt, r, "сидячи ціле життя")
	verbNounEmpty(t, lt, r, "Не претендуючи жодною мірою")
	verbNounEmpty(t, lt, r, "є кого згадувати")
}

// TestTokenAgreementVerbNounRule_RuleTn_V_Vinf_N twins TokenAgreementVerbNounRuleTest.testRuleTn_V_Vinf_N.
func TestTokenAgreementVerbNounRule_RuleTn_V_Vinf_N(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "має відбуватися ротація")
	verbNounEmpty(t, lt, r, "має також народитися власна ідея")
	verbNounEmpty(t, lt, r, "мали змогу оцінити відвідувачі")
	verbNounEmpty(t, lt, r, "має ж десь поміститися двигун")
	verbNounEmpty(t, lt, r, "дав трохи передихнути бізнесу")
	verbNounEmpty(t, lt, r, "дають змогу з комфортом мандрувати чотирьом пасажирам")
	verbNounEmpty(t, lt, r, "дали б змогу розвиватися національному")
	verbNounEmpty(t, lt, r, "дозволила на початку 1990-х узагалі виникнути такому інституту")
	verbNounEmpty(t, lt, r, "заважає і далі нестримно поширюватися багатьом міфам")
	verbNounEmpty(t, lt, r, "люблять у нас кричати панікери")
	verbNounEmpty(t, lt, r, "Почав різко зростати курс долара")
	verbNounEmpty(t, lt, r, "пропонує «об’єднатися патріотам»")
	verbNounEmpty(t, lt, r, "став формуватися прошарок")
	verbNounEmpty(t, lt, r, "ставимо розварюватися легко відтиснену")
	verbNounEmpty(t, lt, r, "не даючи виїхати ванатжівці")
	verbNounEmpty(t, lt, r, "даючи можливість висловлюватися радикалам")
	verbNounEmpty(t, lt, r, "дозволяючи рухатися російському")
	verbNounEmpty(t, lt, r, "готувала їсти хлопцям")
	verbNounEmpty(t, lt, r, "вкласти спати Маринку")
	verbNounEmpty(t, lt, r, "поклавши спати старого Якима")
	verbNounEmpty(t, lt, r, "заважають розвиватися погане управління, війна")
	verbNounEmpty(t, lt, r, "став все частіше згадуватися незвичайний наслідок")
	verbNounEmpty(t, lt, r, " починає швидко жовтіти й опадати листя")
	verbNounEmpty(t, lt, r, "перестають діяти й розвиватися демократичні")
}

// TestTokenAgreementVerbNounRule_RuleTn_ADV_Vinf_N twins TokenAgreementVerbNounRuleTest.testRuleTn_ADV_Vinf_N.
func TestTokenAgreementVerbNounRule_RuleTn_ADV_Vinf_N(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "важко розібратися багатьом людям.")
	verbNounEmpty(t, lt, r, "незручно займатися президентові")
	verbNounEmpty(t, lt, r, "пізно готуватися нам")
	verbNounEmpty(t, lt, r, "найлегше ігнорувати людям із категорій")
	verbNounEmpty(t, lt, r, "треба всіма силами берегти кожному")
	verbNounEmpty(t, lt, r, "треба дуже уважно прислухатися владі")
	verbNounEmpty(t, lt, r, "слід реагувати Америці")
	verbNounEmpty(t, lt, r, "слід з обережністю їсти людям")
	verbNounEmpty(t, lt, r, "варто сюди прилетіти людині")
	verbNounEmpty(t, lt, r, "неможливо засвоїти одній людині")
	verbNounEmpty(t, lt, r, "тяжче стало жити селянам")
	verbNounEmpty(t, lt, r, "приємно слухати вчителям")
	verbNounEmpty(t, lt, r, "повинно боротися суспільство")
	verbNounEmpty(t, lt, r, "слід реально готуватися суспільству")
}

// TestTokenAgreementVerbNounRule_RuleTn_ADJ_Vinf_N twins TokenAgreementVerbNounRuleTest.testRuleTn_ADJ_Vinf_N.
func TestTokenAgreementVerbNounRule_RuleTn_ADJ_Vinf_N(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "змушені ночувати пасажири")
	verbNounEmpty(t, lt, r, "повинен усього добитися сам")
	verbNounEmpty(t, lt, r, "схильна лякати така пропаганда")
	verbNounEmpty(t, lt, r, "зацікавлена перейняти угорська сторона")
}

// TestTokenAgreementVerbNounRule_RuleTn_NOUN_Vinf_N twins TokenAgreementVerbNounRuleTest.testRuleTn_NOUN_Vinf_N.
func TestTokenAgreementVerbNounRule_RuleTn_NOUN_Vinf_N(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "Пора дорослішати всім")
	verbNounEmpty(t, lt, r, "можливість висловитися серійному вбивці?")
	verbNounEmpty(t, lt, r, "рішення про можливість балотуватися Кучмі")
	verbNounEmpty(t, lt, r, "гріх уже зараз зайнятися Генеральній прокуратурі...")
	verbNounEmpty(t, lt, r, "готовність спілкуватися людини")
	verbNounEmpty(t, lt, r, "небажання вибачатися пов’язане з національною гордістю")
	verbNounHasError(t, lt, r, "бажання постійно вчитися новому")
	verbNounHasError(t, lt, r, "Кіно вчить вмінню простими словами")
	verbNounHasError(t, lt, r, "Черга вчитися мистецтву мовчання")
	verbNounHasError(t, lt, r, "у своєму виступі на конференції порадив української владі звернуть увагу")
}

// TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_V twins TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_V.
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_V(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "дозволивши розвалитись імперії")
	verbNounEmpty(t, lt, r, "зніматися йому доводилося рідко")
	verbNounEmpty(t, lt, r, "Гуляти мешканки гуртожитку тепер можуть")
	verbNounEmpty(t, lt, r, "працювати ці люди не вміють")
	verbNounEmpty(t, lt, r, "реагувати Майдан має")
	verbNounEmpty(t, lt, r, "працювати українці будуть")
	verbNounEmpty(t, lt, r, "влаштуватися їй не вдається")
	verbNounEmpty(t, lt, r, "працювати цьому політикові доводиться")
	verbNounEmpty(t, lt, r, "Робити прогнозів не буду")
	verbNounEmpty(t, lt, r, "панькатися наміру не має")
	verbNounEmpty(t, lt, r, "було ввезено тракторів")
}

// TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADV twins TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADV.
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADV(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "літати тобі не можна")
	verbNounEmpty(t, lt, r, "Порозумітися обом сторонам було важко")
	verbNounEmpty(t, lt, r, "втихомирюватися нам зарано")
	verbNounEmpty(t, lt, r, "Зупинятися мені вже не можна")
	verbNounEmpty(t, lt, r, "працювати правоохоронцям складно")
	verbNounEmpty(t, lt, r, "працювати правоохоронцям досить складно")
	verbNounEmpty(t, lt, r, "звертатися пенсіонерам не потрібно")
	verbNounEmpty(t, lt, r, "Їхати мені було не страшно")
	verbNounHasError(t, lt, r, "навчитися цьому досить легко")
}

// TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADJ twins TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADJ.
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADJ(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "замислитися нащадки повинні")
	verbNounEmpty(t, lt, r, "працювати студенти готові")
	verbNounEmpty(t, lt, r, "платити Рига згодна")
	verbNounEmpty(t, lt, r, "Владі не потрібен мир, хоча і війну вести вона не здатна.")
}

// TestTokenAgreementVerbNounRule_RuleTn_Vinf_V_N twins TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_V_N.
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_V_N(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "Заспокоювати стали найагресивнішого")
	verbNounEmpty(t, lt, r, "платити доведеться повну вартість")
	verbNounEmpty(t, lt, r, "закінчити цей огляд хочеться словами")
	verbNounHasError(t, lt, r, "все досліджуйте і постійно навчайтеся новому.")
}

// TestTokenAgreementVerbNounRule_RuleTn_N_Vinf_ADJ twins TokenAgreementVerbNounRuleTest.testRuleTn_N_Vinf_ADJ (Java @Ignore).
func TestTokenAgreementVerbNounRule_RuleTn_N_Vinf_ADJ(t *testing.T) {
	t.Skip("Java @Ignore")
}

// TestTokenAgreementVerbNounRule_RuleTnNumr twins TokenAgreementVerbNounRuleTest.testRuleTnNumr.
func TestTokenAgreementVerbNounRule_RuleTnNumr(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "купив три книги")
	verbNounEmpty(t, lt, r, "входило двоє студентів")
	verbNounEmpty(t, lt, r, "виповнилося шістнадцять")
	verbNounEmpty(t, lt, r, "зобов'язав обох порушників")
	verbNounMatches(t, lt, r, 1, "зобов'язав обом порушникам", "давальний")
	verbNounEmpty(t, lt, r, "передав решту шкіл")
	verbNounEmpty(t, lt, r, "одержав хабарів на суму 10")
	verbNounEmpty(t, lt, r, "одержав хабарів на загальну суму")
	verbNounEmpty(t, lt, r, "уклали договорів страхування на загальну суму")
	verbNounEmpty(t, lt, r, "и втратили капіталовкладень нерезидентів замалим не на $13,6 млрд.")
	verbNounEmpty(t, lt, r, "приготували сортів десять")
	verbNounEmpty(t, lt, r, "зібрали сортів 10")
	verbNounEmpty(t, lt, r, "зібрали сортів — 10")
	verbNounEmpty(t, lt, r, "вчинено порушень обсягом на 27")
	verbNounEmpty(t, lt, r, "завдав кілька ударів руками")
	verbNounEmpty(t, lt, r, "завдав кількох ударів руками")
	verbNounEmpty(t, lt, r, "потребувала мільйон")
	verbNounEmpty(t, lt, r, "нарубав неймовірну кількість вугілля")
	verbNounEmpty(t, lt, r, "відбувся чверть століття тому")
	verbNounEmpty(t, lt, r, "слідкувало майже мільйон людей")
	verbNounEmpty(t, lt, r, "не виїхало більшість автобусів")
	verbNounEmpty(t, lt, r, "буде лежати двоє хворих")
}

// TestTokenAgreementVerbNounRule_RuleTNvNaz twins TokenAgreementVerbNounRuleTest.testRuleTNvNaz.
func TestTokenAgreementVerbNounRule_RuleTNvNaz(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "прийшов Тарас")
	verbNounEmpty(t, lt, r, "було пасмо")
	verbNounEmpty(t, lt, r, "сміялися смішні гієни")
	verbNounEmpty(t, lt, r, "в мені наростали впевненість і сила")
	verbNounEmpty(t, lt, r, "де є якість")
	verbNounEmpty(t, lt, r, "в усіх країнах є міфи")
	verbNounEmpty(t, lt, r, "звалося Подєбради")
}

// TestTokenAgreementVerbNounRule_RuleTNTime twins TokenAgreementVerbNounRuleTest.testRuleTNTime.
func TestTokenAgreementVerbNounRule_RuleTNTime(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "тренувалися годину")
	verbNounEmpty(t, lt, r, "відбудуться наступного дня")
	verbNounEmpty(t, lt, r, "їхав цілу ніч")
	verbNounEmpty(t, lt, r, "сколихнула минулого року")
	verbNounEmpty(t, lt, r, "збираються цього вечора")
	verbNounEmpty(t, lt, r, "чекати годинами")
	verbNounEmpty(t, lt, r, "спостерігається останнім часом")
	verbNounEmpty(t, lt, r, "відійшли метрів п'ять")
	verbNounEmpty(t, lt, r, "публікується кожні два роки")
	verbNounEmpty(t, lt, r, "зірватися всі ці 27 років")
	verbNounEmpty(t, lt, r, "відбувся минулої неділі")
	verbNounEmpty(t, lt, r, "закінчилося 18-го ввечері")
	verbNounEmpty(t, lt, r, "не знімався останні 10 років")
	verbNounEmpty(t, lt, r, "розпочнеться того ж дня")
	verbNounEmpty(t, lt, r, "помер цього вересня")
	verbNounEmpty(t, lt, r, "з’являться наступного ранку на суд")
	verbNounEmpty(t, lt, r, "мучитися три з половиною роки")
	verbNounEmpty(t, lt, r, "буде стояти двадцять і три роки")
	verbNounEmpty(t, lt, r, "попрацювала місяць-півтора")
	verbNounEmpty(t, lt, r, "сніг не тане всю зиму")
	verbNounEmpty(t, lt, r, "вщухнуть протягом кількох днів")
	verbNounEmpty(t, lt, r, "побутувала одночасно двома чи трьома мовами")
	verbNounEmpty(t, lt, r, "працює більшу частину часу")
	verbNounEmpty(t, lt, r, "які стартували цього та минулого року")
}

// TestTokenAgreementVerbNounRule_RuleTnVrod twins TokenAgreementVerbNounRuleTest.testRuleTnVrod.
func TestTokenAgreementVerbNounRule_RuleTnVrod(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "не мав меблів")
	verbNounEmpty(t, lt, r, "не повинен пропускати жодного звуку")
	verbNounEmpty(t, lt, r, "не став витрачати грошей")
	verbNounEmpty(t, lt, r, "казала цього не робити")
	verbNounEmpty(t, lt, r, "не повинно перевищувати граничної величини")
	verbNounEmpty(t, lt, r, "здаватися коаліціанти не збираються")
	verbNounEmpty(t, lt, r, "не існувало конкуренції")
	verbNounEmpty(t, lt, r, "не можна зберігати ілюзій")
	verbNounEmpty(t, lt, r, "скільки отримує грошей")
	verbNounEmpty(t, lt, r, "скільки буде людей")
	verbNounEmpty(t, lt, r, "трохи маю контактів")
	verbNounEmpty(t, lt, r, "скільки загалом здійснили постановок")
	verbNounEmpty(t, lt, r, "стільки заплановано постановок")
	verbNounEmpty(t, lt, r, "Багато в пресі з’явилося публікацій")
	verbNounEmpty(t, lt, r, "небагато надходить книжок")
	verbNounEmpty(t, lt, r, "ніж поставили стільців")
	verbNounEmpty(t, lt, r, "Росія будує трубопроводів більше")
	verbNounEmpty(t, lt, r, "не стане сили")
	verbNounEmpty(t, lt, r, "нарубав лісу")
}

// TestTokenAgreementVerbNounRule_RuleTnInsertPhrase twins TokenAgreementVerbNounRuleTest.testRuleTnInsertPhrase.
func TestTokenAgreementVerbNounRule_RuleTnInsertPhrase(t *testing.T) {
	lt, r := wireTokenAgreementVerbNoun(t)

	verbNounEmpty(t, lt, r, "змалював дивовижної краси церкву")
	verbNounEmpty(t, lt, r, "Піднявся страшенної сили крижаний шторм")
	verbNounEmpty(t, lt, r, "поводиться дивним чином")
	verbNounEmpty(t, lt, r, "триває повним ходом.")
}

// TestTokenAgreementVerbNounRule_RuleDisambigNazInf twins TokenAgreementVerbNounRuleTest.testRuleDisambigNazInf.
func TestTokenAgreementVerbNounRule_RuleDisambigNazInf(t *testing.T) {
	_, r := wireTokenAgreementVerbNoun(t)

	// Java builds synthetic ATRs: SENT_START + повинен + був + випадок + передбачити
	sentStart := languagetool.SentenceStartTagName
	empty := ""
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &sentStart, &empty), 0),
	}
	surfaces := []string{"повинен", "був", "випадок", "передбачити"}
	lemmas := []string{"повинний", "бути", "випадок", "передбачити"}
	pos := []string{"adj:m:v_naz:short:predic", "verb:imperf:past:m", "noun:inanim:m:v_zna", "verb:perf:inf"}
	for i, s := range surfaces {
		lem := lemmas[i]
		p := pos[i]
		at := languagetool.NewAnalyzedToken(s, &p, &lem)
		toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(at, 0))
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	require.Equal(t, 0, len(r.Match(sent)))
}
