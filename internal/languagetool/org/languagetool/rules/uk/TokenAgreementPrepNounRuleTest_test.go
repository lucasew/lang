package uk_test

// Twin of languagetool-language-modules/uk TokenAgreementPrepNounRuleTest
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new TokenAgreementPrepNounRule(TestTools.getMessages("uk"), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator + Ukrainian synthesizer (suggestions) +
// IgnoredCharacters (soft hyphen + combining acute).

import (
	"os"
	"path/filepath"
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

func discoverUKResource(t *testing.T, name string) string {
	t.Helper()
	if p := os.Getenv("LANG_UKRAINIAN_DICT"); p != "" && name == "ukrainian.dict" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	candidates := []string{
		filepath.Join("/tmp/uk-dict-fetch/org/languagetool/resource/uk", name),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
			"src", "main", "resources", "org", "languagetool", "resource", "uk", name),
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Skip(name + " not available")
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range candidates {
			p := rel
			if !filepath.IsAbs(rel) {
				p = filepath.Join(dir, rel)
			}
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip(name + " not available (set LANG_UKRAINIAN_DICT or place morfologik-ukrainian-lt resource)")
	return ""
}

// wireTokenAgreementPrepNoun ports JLanguageTool(Ukrainian) + TokenAgreementPrepNounRule ctor.
func wireTokenAgreementPrepNoun(t *testing.T) (*languagetool.JLanguageTool, *rulesuk.TokenAgreementPrepNounRule) {
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

	r := rulesuk.NewTokenAgreementPrepNounRule()
	// Java ukrainian.getSynthesizer() — official ukrainian_synth.dict + tags
	if sp := discoverUKResource(t, "ukrainian_synth.dict"); sp != "" {
		if base := synthesis.OpenBaseSynthesizerFromDictPath("uk", sp); base != nil {
			r.Synth = &synthuk.UkrainianSynthesizer{BaseSynthesizer: base}
		}
	}
	// Java ukrainian.getTagger().tag for prep+adv merge message
	r.Tag = func(words []string) []*languagetool.AnalyzedTokenReadings {
		return tg.Tag(words)
	}
	return lt, r
}

func prepNounEmpty(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementPrepNounRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 0, len(ms), "assertEmptyMatch %q got %d", text, len(ms))
}

func prepNounHasError(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementPrepNounRule, text string, suggestions ...string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 1, len(ms), "assertHasError %q got %d", text, len(ms))
	if len(suggestions) > 0 {
		require.Equal(t, suggestions, ms[0].GetSuggestedReplacements(), "suggestions for %q", text)
		// Java: each suggestion alone has empty match
		for _, s := range suggestions {
			prepNounEmpty(t, lt, r, s)
		}
	}
}


// TestTokenAgreementPrepNounRule_Rule twins TokenAgreementPrepNounRuleTest.testRule.
func TestTokenAgreementPrepNounRule_Rule(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	require.Equal(t, rulesuk.TokenAgreementPrepNounRuleID, r.GetID())

	// correct sentences:
	prepNounEmpty(t, lt, r, "без повного")
	prepNounEmpty(t, lt, r, "без неба")
	prepNounEmpty(t, lt, r, "по авеню")
	prepNounEmpty(t, lt, r, "що за ганебна непослідовність?")
	prepNounEmpty(t, lt, r, "щодо власне людини")
	prepNounEmpty(t, lt, r, "у загалом симпатичній повістині")
	prepNounEmpty(t, lt, r, "понад половина людей")
	prepNounEmpty(t, lt, r, "з понад ста людей")
	prepNounEmpty(t, lt, r, "по нервах")
	prepNounEmpty(t, lt, r, "з особливою увагою")
	prepNounEmpty(t, lt, r, "щодо бодай гіпотетичної здатності")
	prepNounEmpty(t, lt, r, "хто їде на заробітки за кордон")
	prepNounEmpty(t, lt, r, "піти в президенти")
	prepNounEmpty(t, lt, r, "піти межі люди")
	prepNounEmpty(t, lt, r, "вивів в люде")
	prepNounEmpty(t, lt, r, "що то була за людина")
	prepNounEmpty(t, lt, r, "що за людина")
	prepNounHasError(t, lt, r, "що балотувався за цім округом")
	prepNounEmpty(t, lt, r, "на дому")
	prepNounEmpty(t, lt, r, "на біс")
	prepNounEmpty(t, lt, r, "згідно з документом")
	prepNounEmpty(t, lt, r, "Серед святкових товарів")
	prepNounEmpty(t, lt, r, "зовсім не святкові товари Серед святкових товарів")
	prepNounEmpty(t, lt, r, "при кому знайдено вогнепальну")
	prepNounEmpty(t, lt, r, "За його словами Україна – це країна...")
	prepNounEmpty(t, lt, r, "славетних од цареві")
	prepNounEmpty(t, lt, r, "А шляхом тим була")
	prepNounHasError(t, lt, r, "що, незважаючи стислі терміни візиту")
	// numr
	prepNounHasError(t, lt, r, "до понад трьома")
	prepNounHasError(t, lt, r, "до тисяча трьома")
	prepNounEmpty(t, lt, r, "лежить із сотня срібних")
	prepNounEmpty(t, lt, r, "залежно що вважати перемогою")
	prepNounEmpty(t, lt, r, "окрім як українці")
	prepNounEmpty(t, lt, r, "за двісті метрів")
	prepNounEmpty(t, lt, r, "від мінус 1 до плюс 1")
	prepNounEmpty(t, lt, r, "від мінус $1 до плюс 1")
	prepNounEmpty(t, lt, r, "до мінус сорока град")
	prepNounEmpty(t, lt, r, "до мінус шістдесяти")
	prepNounEmpty(t, lt, r, "через років 10")
	prepNounEmpty(t, lt, r, "через років зо два")
	prepNounEmpty(t, lt, r, "на хвилин 9-10")
	prepNounEmpty(t, lt, r, "співпрацювати із собі подібними")
	prepNounEmpty(t, lt, r, "через усім відомі причини")
	prepNounEmpty(t, lt, r, "через нікому не відомі причини")
	prepNounEmpty(t, lt, r, "до нічим не виправданого нарцисизму")
	prepNounEmpty(t, lt, r, "прийшли до ВАТ «Кривий Ріг цемент»")
	prepNounEmpty(t, lt, r, "від А до Я")
	prepNounEmpty(t, lt, r, "до та після")
	prepNounEmpty(t, lt, r, "до схід сонця")
	prepNounEmpty(t, lt, r, "з рана до вечора, від рана до ночі")
	prepNounEmpty(t, lt, r, "до НАК «Надра України»")
	prepNounEmpty(t, lt, r, "Вони замість Андрій вибрали Юрій")
	prepNounEmpty(t, lt, r, "час від часу нам доводилось")
	prepNounEmpty(t, lt, r, "ні до чого доброго силові дії не призведуть")
	prepNounEmpty(t, lt, r, "у святая святих")
	prepNounEmpty(t, lt, r, "станом на зараз виконавча влада")
	prepNounEmpty(t, lt, r, "в тисяча шістсот якомусь році")
	prepNounEmpty(t, lt, r, "у ти́сяча сімсо́т вісімдеся́т дев’я́тому")
	prepNounEmpty(t, lt, r, "на Піп Іван")

	// incorrect sentences:
	matches := r.Match(lt.GetAnalyzedSentence("без небу"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"неба"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "родовий")

	matches = r.Match(lt.GetAnalyzedSentence("не в останню чергу через корупцією, міжрелігійну ворожнечу"))
	require.Equal(t, 1, len(matches))

	matches = r.Match(lt.GetAnalyzedSentence("по нервам"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 9, matches[0].GetToPos())
	require.Equal(t, []string{"нервах", "нерви"}, matches[0].GetSuggestedReplacements())

	matches = r.Match(lt.GetAnalyzedSentence("по швидко напруженим рукам"))
	require.Equal(t, 1, len(matches))

	matches = r.Match(lt.GetAnalyzedSentence("по «нервам»"))
	require.Equal(t, 1, len(matches))

	prepNounEmpty(t, lt, r, "про «справжній футбол»")
	prepNounEmpty(t, lt, r, "на «ти»")
	prepNounEmpty(t, lt, r, "писав про «Сновиди»")

	matches = r.Match(lt.GetAnalyzedSentence("до не властиву"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("до не дуже властиву"))
	require.Equal(t, 1, len(matches))

	prepNounEmpty(t, lt, r, "На сьогодні рослинна їжа")

	// numr
	prepNounHasError(t, lt, r, "в п'ятьом людям")
	prepNounHasError(t, lt, r, "в понад п'ятьом людям")
	prepNounEmpty(t, lt, r, "у тисяча якомусь")

	analyzed := lt.GetAnalyzedSentence("О дівчина!")
	match := r.Match(analyzed)
	require.Equal(t, 1, len(match))
	require.Contains(t, match[0].GetSuggestedReplacements(), "дівчино")

	matches = r.Match(lt.GetAnalyzedSentence("по церковним канонам"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "церковних")

	// свята / вулиці / ім'я
	prepNounEmpty(t, lt, r, "на Купала")
	prepNounEmpty(t, lt, r, "на Явдохи")
	prepNounEmpty(t, lt, r, "на Мазепи")
	prepNounEmpty(t, lt, r, "на Кульчицької")
	prepNounEmpty(t, lt, r, "на Правди")
	prepNounEmpty(t, lt, r, "на Ломоносова")
	prepNounEmpty(t, lt, r, "як на Кучми іменини")
	prepNounEmpty(t, lt, r, "змінив ім'я на Фріц Ланг")
	prepNounEmpty(t, lt, r, "Бо заміна прізвища Горбатий на Щербань передбачає i зміну ситуації.")
	prepNounEmpty(t, lt, r, "З одного боку на щастя сім’я Ющенків нарешті з’їжджає з державної дачі.")
	prepNounEmpty(t, lt, r, "До сьогодні українська влада намагалася")
	prepNounEmpty(t, lt, r, "Так висловлюються про екс-першого віце-спікера.")

	matches = r.Match(lt.GetAnalyzedSentence("спиралося на місячної давнини рішенням"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("Від стягу Ататюрка до піратського прапору"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("згідно з документа"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("згідно зі змінам"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("зацікавлених у ви користанні"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("без правда"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("В йому заграла кров."))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("  В йому заграла кров."))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("І от «В йому заграла кров»."))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("похвалила при йому вкраїнську мову"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("думає на тим, як"))
	require.Equal(t, 1, len(matches))
	prepNounEmpty(t, lt, r, "гепатитів В та С")
	matches = r.Match(lt.GetAnalyzedSentence("— О пан Єзус, захисти їх!"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("до Ленінграду"))
	require.Equal(t, 1, len(matches))
	matches = r.Match(lt.GetAnalyzedSentence("для логотипу"))
	require.Equal(t, 1, len(matches))
}

// TestTokenAgreementPrepNounRule_ZandZnaAsRare twins testZandZnaAsRare.
func TestTokenAgreementPrepNounRule_ZandZnaAsRare(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	prepNounEmpty(t, lt, r, "кілометрів зо три")
	prepNounEmpty(t, lt, r, "години з 30")
	prepNounEmpty(t, lt, r, "людей з десяток")
	prepNounEmpty(t, lt, r, "насіння зі жменьку")
	prepNounEmpty(t, lt, r, "розміром з долоню")
	prepNounEmpty(t, lt, r, "шматок з кулак завбільшки")
	prepNounEmpty(t, lt, r, "винайняло з тисячу модераторів")
	prepNounEmpty(t, lt, r, "причеплено ще з сотню дерев'яних крамниць")
	prepNounEmpty(t, lt, r, "з 9-12-поверховий будинок")
	prepNounEmpty(t, lt, r, "виниклу з нізвідки тему")
	prepNounEmpty(t, lt, r, "з Ван Гогом")
	prepNounHasError(t, lt, r, "зі землю")
	prepNounHasError(t, lt, r, "туди ж з своє шкапою")
	prepNounHasError(t, lt, r, "вискочив із барлоги")
	prepNounHasError(t, lt, r, "від співпраці з компанію", "компанією", "компанії")
	prepNounEmpty(t, lt, r, "заввишки з її невеликий зріст")
	prepNounEmpty(t, lt, r, "із добру годину")
}

// TestTokenAgreementPrepNounRule_RulePronPos twins testRulePronPos (active).
// testRulePronPosNew is @Ignore in Java — not twinned as active.
func TestTokenAgreementPrepNounRule_RulePronPos(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	prepNounEmpty(t, lt, r, "повернесть до них-таки.")
	prepNounEmpty(t, lt, r, "колега поруч сонним виглядає")
	prepNounEmpty(t, lt, r, "при їх опублікуванні")
	prepNounEmpty(t, lt, r, "всупереч їх рекомендаціям")
	prepNounEmpty(t, lt, r, "на їх користь стягнуто")
	prepNounEmpty(t, lt, r, "не всупереч, а тому, що він має")
	prepNounEmpty(t, lt, r, "у його (лікаря) присутності")
	prepNounHasError(t, lt, r, "до їх")
	prepNounEmpty(t, lt, r, "Під його, без перебільшення, мудрим")
	prepNounEmpty(t, lt, r, "до її, так би мовити, санітарного стану")
	prepNounEmpty(t, lt, r, "в його, судячи з інтер’єру, службовому кабінеті")
	prepNounEmpty(t, lt, r, "вплив на її або його здоров'я")
	prepNounEmpty(t, lt, r, "щодо його \"лікування\":")
	prepNounEmpty(t, lt, r, "За її словами,")
	prepNounHasError(t, lt, r, "вище за їх?")
	prepNounHasError(t, lt, r, "про їх говорилося")

	analyzed := lt.GetAnalyzedSentence("завдяки їх.")
	match := r.Match(analyzed)
	require.Equal(t, 1, len(match))
	require.Contains(t, match[0].GetSuggestedReplacements(), "їхнім")

	prepNounEmpty(t, lt, r, "при її, м’яко кажучи, невеликій популярності")
	prepNounEmpty(t, lt, r, "через її, м’яко кажучи, невелику популярність")
	prepNounHasError(t, lt, r, "при його ж заняттів")
	prepNounEmpty(t, lt, r, "освітою й без, адже останні, особливо робітники")
	prepNounEmpty(t, lt, r, "сили від одне одного, як сусіди, партнери")
	prepNounEmpty(t, lt, r, "вважали за краще брати")
}

// TestTokenAgreementPrepNounRule_RuleFlexibleOrder twins testRuleFlexibleOrder.
func TestTokenAgreementPrepNounRule_RuleFlexibleOrder(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	prepNounHasError(t, lt, r, "по бодай маленьким справам")
	prepNounHasError(t, lt, r, "по смішно маленьким справам")
	prepNounEmpty(t, lt, r, "за цілком собі реалістичною соціальною")
	prepNounEmpty(t, lt, r, "спиралося на місячної давнини рішення")
	prepNounEmpty(t, lt, r, "На середньої довжини шубу")
	prepNounEmpty(t, lt, r, "При різного роду процесах")
	prepNounHasError(t, lt, r, "завдяки його прийомі")
	prepNounEmpty(t, lt, r, "по лише їм цікавих місцях")
	prepNounHasError(t, lt, r, "по лише їм цікавим місцям")
	prepNounEmpty(t, lt, r, "від дуже близьких людей")
	prepNounEmpty(t, lt, r, "завдяки саме цим сімом голосам")
	prepNounEmpty(t, lt, r, "на мохом стеленому дні")
	prepNounEmpty(t, lt, r, "який до речі вони присягалися")
	prepNounEmpty(t, lt, r, "на нічого не вартий папірець")
	prepNounHasError(t, lt, r, "призвів до значною мірою демократичному середньому класу")
	prepNounHasError(t, lt, r, "це нас для дуже велика сума")
	prepNounEmpty(t, lt, r, "для якої з мов воно первинне")
	prepNounEmpty(t, lt, r, "об'єктивності заради слід зазначити")
	prepNounHasError(t, lt, r, "у дуже обмеженим рамках")
	prepNounEmpty(t, lt, r, "чи не проти я тієї церковної стройки")
	prepNounEmpty(t, lt, r, "З точністю до навпаки ви все це побачите")
	prepNounEmpty(t, lt, r, "Усупереч не те що лихим")
	prepNounEmpty(t, lt, r, "весь світ замість спершу самому засвоїти")
	prepNounEmpty(t, lt, r, "Йдеться про вже всім добре відому")
	prepNounHasError(t, lt, r, "кинулися до мені перші з них")
	prepNounHasError(t, lt, r, "Замість лимону можна брати")
	prepNounEmpty(t, lt, r, "до дев’яносто дев’ятого")
	prepNounEmpty(t, lt, r, "стосовно одне одного")
	prepNounEmpty(t, lt, r, "ніч із тридцять першого серпня")
}

// TestTokenAgreementPrepNounRule_SpecialChars twins testSpecialChars.
func TestTokenAgreementPrepNounRule_SpecialChars(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	prepNounEmpty(t, lt, r, "до їм поді\u00ADбних")

	matches := r.Match(lt.GetAnalyzedSentence("о справедли\u00ADвости."))
	require.Equal(t, 1, len(matches))

	matches = r.Match(lt.GetAnalyzedSentence("по не́рвам, по мо\u00ADстам, по воротам"))
	require.Equal(t, 3, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())
	require.Equal(t, []string{"нервах", "нерви"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, 15, matches[1].GetFromPos())
	require.Equal(t, []string{"мостах", "мости"}, matches[1].GetSuggestedReplacements())
	require.Equal(t, 27, matches[2].GetFromPos())
	require.Equal(t, []string{"воротах", "воротях", "ворота"}, matches[2].GetSuggestedReplacements())
}

// TestTokenAgreementPrepNounRule_UnusualCharacters twins testUnusualCharacters.
func TestTokenAgreementPrepNounRule_UnusualCharacters(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	matches := r.Match(lt.GetAnalyzedSentence("о стін\u00AD\nку"))
	require.Equal(t, 0, len(matches))
}

// TestTokenAgreementPrepNounRule_WithAdv twins testWithAdv.
func TestTokenAgreementPrepNounRule_WithAdv(t *testing.T) {
	lt, r := wireTokenAgreementPrepNoun(t)
	match := r.Match(lt.GetAnalyzedSentence("гречка у двічі дешевша ніж"))
	require.Equal(t, 1, len(match))
	require.Contains(t, match[0].GetMessage(), "Можливо, прийменник і прислівник")
}

// TestTokenAgreementPrepNounRule_IsCapitalized twins testIsCapitalized (LemmaHelper).
func TestTokenAgreementPrepNounRule_IsCapitalized(t *testing.T) {
	require.False(t, rulesuk.IsCapitalized("боснія"))
	require.True(t, rulesuk.IsCapitalized("Боснія"))
	require.True(t, rulesuk.IsCapitalized("Боснія-Герцеговина"))
	require.False(t, rulesuk.IsCapitalized("По-перше"))
	require.False(t, rulesuk.IsCapitalized("ПаП"))
	require.True(t, rulesuk.IsCapitalized("П'ятниця"))
	require.False(t, rulesuk.IsCapitalized("П'ЯТНИЦЯ"))
	require.True(t, rulesuk.IsCapitalized("EuroGas"))
	require.True(t, rulesuk.IsCapitalized("Рясна-2"))
	require.False(t, rulesuk.IsCapitalized("ДБЗПТЛ"))
}

// TestTokenAgreementPrepNounRule_Meta twins Java id / description / short.
func TestTokenAgreementPrepNounRule_Meta(t *testing.T) {
	r := rulesuk.NewTokenAgreementPrepNounRule()
	require.Equal(t, "UK_PREP_NOUN_INFLECTION_AGREEMENT", r.GetID())
	require.Equal(t, "Узгодження прийменника та іменника у реченні", r.GetDescription())
}
