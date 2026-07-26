package uk_test

// Twin of languagetool-language-modules/uk TokenAgreementNounVerbRuleTest
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new TokenAgreementNounVerbRule(TestTools.getMessages("uk"));
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator + IgnoredCharacters (soft hyphen + combining acute).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	ukdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/uk"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// wireTokenAgreementNounVerb ports JLanguageTool(Ukrainian) + TokenAgreementNounVerbRule ctor.
func wireTokenAgreementNounVerb(t *testing.T) (*languagetool.JLanguageTool, *rulesuk.TokenAgreementNounVerbRule) {
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

	r := rulesuk.NewTokenAgreementNounVerbRule()
	return lt, r
}

func nounVerbEmpty(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementNounVerbRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 0, len(ms), "assertEmptyMatch %q got %d matches: %v", text, len(ms), ms)
}

func nounVerbHasError(t *testing.T, lt *languagetool.JLanguageTool, r *rulesuk.TokenAgreementNounVerbRule, text string) {
	t.Helper()
	ms := r.Match(lt.GetAnalyzedSentence(text))
	require.Equal(t, 1, len(ms), "assertHasError %q got %d matches: %v", text, len(ms), ms)
}

// TestTokenAgreementNounVerbRule_Meta twins Java id / description / short.
func TestTokenAgreementNounVerbRule_Meta(t *testing.T) {
	r := rulesuk.NewTokenAgreementNounVerbRule()
	require.Equal(t, rulesuk.TokenAgreementNounVerbRuleID, r.GetID())
	require.Equal(t, "Узгодження іменника та дієслова за родом, числом та особою", r.GetDescription())
	require.Equal(t, "Узгодження іменника з дієсловом", r.GetShort())
}

// TestTokenAgreementNounVerbRule_Rule twins TokenAgreementNounVerbRuleTest.testRule.
func TestTokenAgreementNounVerbRule_Rule(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbHasError(t, lt, r, "Тарас прибігла")
	nounVerbHasError(t, lt, r, "вони прибіг")
	nounVerbHasError(t, lt, r, "я прибіжиш")
	nounVerbHasError(t, lt, r, "вони швидко прибіг")
	nounVerbHasError(t, lt, r, "та з інших питань перевірка проведено не повно")
	nounVerbHasError(t, lt, r, "з часом пара вирішили узаконити")
	nounVerbEmpty(t, lt, r, "На честь Джудіт Резнік названо кратер")
	nounVerbEmpty(t, lt, r, "чи зуміє наша держава забезпечити власні потреби")
	nounVerbEmpty(t, lt, r, "так навчила мене бабуся місити пухке дріжджове тісто")
	nounVerbEmpty(t, lt, r, "чи можуть російськомовні громадяни вважатися українцями")
	nounVerbEmpty(t, lt, r, "не шкода віддати життя")
	nounVerbEmpty(t, lt, r, "не шкода було віддати життя")
	nounVerbEmpty(t, lt, r, "чоловік прибіг")
	nounVerbEmpty(t, lt, r, "я прибіг")
	nounVerbEmpty(t, lt, r, "я прибігла")
	nounVerbEmpty(t, lt, r, "кандидат в президенти поїхав")
	nounVerbEmpty(t, lt, r, "кандидат в народні депутати поїхав")
	nounVerbEmpty(t, lt, r, "40 тисяч чоловік виявили бажання")
	nounVerbEmpty(t, lt, r, "клан Алькапоне було знищено")
	nounVerbEmpty(t, lt, r, "він було вмовк")
	nounVerbEmpty(t, lt, r, "тобто Угорщина було пішла шляхом")
	nounVerbEmpty(t, lt, r, "він було трохи скис")
	nounVerbEmpty(t, lt, r, "більше ніж будь-хто маємо повне право")
	nounVerbEmpty(t, lt, r, "а вона візьми і зроби")
	nounVerbEmpty(t, lt, r, "Збережені Я позбудуться необхідності")
	nounVerbEmpty(t, lt, r, "У місті Біла Церква було сформовано")
	nounVerbHasError(t, lt, r, "із містом юності Ви очікувала")
	nounVerbEmpty(t, lt, r, "Андрій Качала")
	nounVerbEmpty(t, lt, r, "Любов Євтушок зауважила")
	nounVerbEmpty(t, lt, r, "а він давай пити горілку")
	nounVerbEmpty(t, lt, r, "Тарас ЗАКУСИЛО")
	nounVerbEmpty(t, lt, r, "не сила була йти далі")
	nounVerbEmpty(t, lt, r, "Вже давно пора було Мовчану відійти від керма.")
	nounVerbEmpty(t, lt, r, "тому, що як австрієць маєте")
	nounVerbHasError(t, lt, r, "не встиг вона отямитися")
	nounVerbEmpty(t, lt, r, "про припинення їхньої діяльності ми ухвалити, зрозуміло, не могли")
	nounVerbHasError(t, lt, r, "Ми може спробувати знайти")
	nounVerbEmpty(t, lt, r, nounVerbGoodText)
	nounVerbEmpty(t, lt, r, "— це були невільники")
	nounVerbHasError(t, lt, r, "щоб конкуренти підішли до виборів")
}

// TestTokenAgreementNounVerbRule_RuleNe twins TokenAgreementNounVerbRuleTest.testRuleNe.
func TestTokenAgreementNounVerbRule_RuleNe(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbHasError(t, lt, r, "Тарас не прибігла")
	nounVerbHasError(t, lt, r, "Тарас би не прибігла")
	nounVerbHasError(t, lt, r, "вони не прибіг")
	nounVerbHasError(t, lt, r, "я не прибіжиш")
	nounVerbEmpty(t, lt, r, "ні він, ні вона не хотіли")
	nounVerbEmpty(t, lt, r, "уряд та поліція не контролюють події")
	nounVerbEmpty(t, lt, r, "— це не була хвороба")
	nounVerbEmpty(t, lt, r, "— це не передбачено")
	nounVerbEmpty(t, lt, r, "Решта не мають жодних дозволів")
	nounVerbEmpty(t, lt, r, "ні лауреат, ні його жінка не розмовляють жодними мовами")
	nounVerbEmpty(t, lt, r, "Чи ж могла я не повернутися назад?")
	nounVerbEmpty(t, lt, r, "кефаль, хамса не затримуються")
	nounVerbHasError(t, lt, r, "інша мова, вона примітивізовано й")
	nounVerbEmpty(t, lt, r, "душ, одеколони, навіть хлорка не допомогли")
	nounVerbEmpty(t, lt, r, "і Виговський, ні навіть Мазепа не розглядали")
	nounVerbEmpty(t, lt, r, "уряд та поліція не контролюють")
	nounVerbEmpty(t, lt, r, "кадрове забезпечення, матеріальна база не дісталися")
	nounVerbEmpty(t, lt, r, "Волосожар/Морозов не припинили")
	nounVerbEmpty(t, lt, r, "ні Європа, ні тим більше Україна не мають")
}

// TestTokenAgreementNounVerbRule_ProperNames twins TokenAgreementNounVerbRuleTest.testProperNames.
func TestTokenAgreementNounVerbRule_ProperNames(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "Конституційний суд Республіки Молдова визнав румунську державною")
	nounVerbEmpty(t, lt, r, "Мешканці планети Земля споживають щороку")
	nounVerbEmpty(t, lt, r, "жителі селища Новобудова зверталися.")
	nounVerbHasError(t, lt, r, "при своїх дружинах Крішна не роздягалася і одразу...")
	nounVerbEmpty(t, lt, r, "всі українські жінки з ім’ям Марія мають знати")
	nounVerbEmpty(t, lt, r, "а ім’я Франклін згадують не досить часто")
	nounVerbEmpty(t, lt, r, "шимпанзе на прізвисько Чита зіграли 16 «акторів»")
	nounVerbEmpty(t, lt, r, "лижний курорт Криниця розташувався в Бескидах")
	nounVerbEmpty(t, lt, r, "містечко Баришівка потрапило до історії")
	nounVerbEmpty(t, lt, r, "У місті Ліда незабаром мають встановити")
	nounVerbEmpty(t, lt, r, "ми в державі Україна маємо такі підстави")
	nounVerbEmpty(t, lt, r, "У литовський порт Клайпеда прибуло плавуче сховище")
	nounVerbEmpty(t, lt, r, "на австралійський штат Вікторія налетів сильний шторм")
	nounVerbEmpty(t, lt, r, "в селі Червона Слобода було вирішено перейменувати")
	nounVerbEmpty(t, lt, r, "Хоча б межі курорту Східниця визначено?")
	nounVerbEmpty(t, lt, r, "комітет порятунку села Оляниця вирішив")
	nounVerbEmpty(t, lt, r, "колишній кандидат у губернатори штату Аризона їхав до Чернівців")
	nounVerbEmpty(t, lt, r, "У невизнаній республіці Південна Осетія відбулися вибори")
	nounVerbEmpty(t, lt, r, "Рибалки італійського острова Лампедуза заблокували")
	nounVerbEmpty(t, lt, r, "у латвійське курортне містечко Юрмала з’їхався весь бомонд")
	nounVerbEmpty(t, lt, r, "Суд американського штату Каліфорнія присудив акторці")
	nounVerbEmpty(t, lt, r, "У штатах Техас і Луїзіана запроваджено надзвичайний стан")
	nounVerbEmpty(t, lt, r, "на прізвисько Михайло відбулася")
	nounVerbHasError(t, lt, r, "Вистава зроблено чесно, професійно.")
}

// TestTokenAgreementNounVerbRule_NounAsAdv twins TokenAgreementNounVerbRuleTest.testNounAsAdv.
func TestTokenAgreementNounVerbRule_NounAsAdv(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "під три чорти пертися")
}

// TestTokenAgreementNounVerbRule_Pron twins TokenAgreementNounVerbRuleTest.testPron.
func TestTokenAgreementNounVerbRule_Pron(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbHasError(t, lt, r, "яка прибіг")
	nounVerbHasError(t, lt, r, "яка не мають паспортів")
	nounVerbHasError(t, lt, r, "яка не залежать від волі")
	nounVerbHasError(t, lt, r, "яка було нещодавно опубліковано")
	nounVerbEmpty(t, lt, r, "вони як ніхто інший знали")
	nounVerbEmpty(t, lt, r, "який прибила хвиля")
	nounVerbEmpty(t, lt, r, "Ті, хто зрозуміли")
	nounVerbEmpty(t, lt, r, "ті, хто сповідує")
	nounVerbEmpty(t, lt, r, "ті, хто не сповідують")
	nounVerbEmpty(t, lt, r, "Ті, хто просто зрозуміли")
	nounVerbEmpty(t, lt, r, "про те, хто була ця клята Пандора")
	nounVerbEmpty(t, lt, r, "що можна було й інший пошукати")
}

// TestTokenAgreementNounVerbRule_VerbInf twins TokenAgreementNounVerbRuleTest.testVerbInf.
func TestTokenAgreementNounVerbRule_VerbInf(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "не встиг я отямитися")
	nounVerbEmpty(t, lt, r, "що я зробити встиг")
	nounVerbEmpty(t, lt, r, "це я робити швидко вмію")
	nounVerbEmpty(t, lt, r, "Саудівську Аравію ми проходити зобов’язані")
	nounVerbHasError(t, lt, r, "я бігати")
	nounVerbEmpty(t, lt, r, "ми воювати з нашими людьми не збираємося")
	nounVerbEmpty(t, lt, r, "що ми зробити не зможемо")
	nounVerbEmpty(t, lt, r, "Я уявити себе не можу без нашої програми.")
	nounVerbEmpty(t, lt, r, "ми розраховувати не повинні")
	nounVerbEmpty(t, lt, r, "Хотів би я подивитися")
	nounVerbEmpty(t, lt, r, "на останніх ми працювати не згідні")
	nounVerbHasError(t, lt, r, "на останніх ми працювати не питаючи нікого")
	nounVerbEmpty(t, lt, r, "те, чого я слухати не дуже хочу")
	nounVerbEmpty(t, lt, r, "чи гідні ми бути незалежними")
	nounVerbEmpty(t, lt, r, "Чи здатен її автор навести хоча б один факт")
	nounVerbEmpty(t, lt, r, "чи готові ми сидіти без світла")
	nounVerbEmpty(t, lt, r, "Чи повинен я просити")
	nounVerbEmpty(t, lt, r, "ніхто робити цього не буде")
	nounVerbEmpty(t, lt, r, "ніхто знижувати тарифи на газ, зрозуміло, не збирається")
	nounVerbHasError(t, lt, r, "Та припинити ти переживати")
	nounVerbEmpty(t, lt, r, "та я купувати цю куртку не дуже хотіла")
	nounVerbEmpty(t, lt, r, "Правоохоронці зняти провокаційний стяг не змогли, оскільки")
	nounVerbEmpty(t, lt, r, "Однак Банкова виконувати приписи Закону \"Про очищення влади\" не поспішає")
	nounVerbEmpty(t, lt, r, "Хтось намагається її торкнутися, хтось сфотографувати.")
	nounVerbEmpty(t, lt, r, "Ми не проти сплачувати податки")
}

// TestTokenAgreementNounVerbRule_Plural twins TokenAgreementNounVerbRuleTest.testPlural.
func TestTokenAgreementNounVerbRule_Plural(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbHasError(t, lt, r, "21 гравець дивилися")
	nounVerbHasError(t, lt, r, "один гравець дивилися")
	nounVerbHasError(t, lt, r, "Серед вбитих і полонених радянських солдат були чоловіки")
	nounVerbHasError(t, lt, r, "Пригадую випадок, коли одна аудіокомпанія звернулися до провідних київських «ефемок»")
	nounVerbHasError(t, lt, r, "підписку про невиїзд двох молодиків, яких міліція затримали першими")
	nounVerbHasError(t, lt, r, "рук та ніг Параски, темниця розчинилася і дівчина опинилися за стінами фортеці")
	nounVerbHasError(t, lt, r, "його арештували і вислали у Воркуту, я залишилися одна з дитиною.")
	nounVerbHasError(t, lt, r, "На проспекті Чорновола, ближче до центру, вона зупинилися на перехресті")
	nounVerbHasError(t, lt, r, "порадилися і громада запропонували мені зайняти його місце")
	nounVerbHasError(t, lt, r, "то й небо вона бачите саме таким")
	nounVerbHasError(t, lt, r, "молочного туману, потім вона заснувалися")
	nounVerbHasError(t, lt, r, "Наташа смикала за волосся, а Софія намагалися бризнути")
	nounVerbEmpty(t, lt, r, "моя мама й сестра мешкали")
	nounVerbEmpty(t, lt, r, "чи то Вальтер, чи я вжили фразу")
	nounVerbEmpty(t, lt, r, "То вона, то Гриць виринають перед її душею")
	nounVerbEmpty(t, lt, r, "Кожен чоловік і кожна жінка мають")
	nounVerbEmpty(t, lt, r, "каналізація і навіть охорона пропонувалися")
	nounVerbEmpty(t, lt, r, "Кавказ загалом і Чечня зокрема лишаться")
	nounVerbEmpty(t, lt, r, "Європейський Союз і моя рідна дочка переживуть це збурення")
	nounVerbEmpty(t, lt, r, "Бразилія, Мексика, Індія збувають")
	nounVerbEmpty(t, lt, r, "Банкова й особисто президент дістали")
	nounVerbEmpty(t, lt, r, "українська мода та взагалі українська культура впливають")
	nounVerbEmpty(t, lt, r, "Тато і Юзь Федорків були прикладом")
	nounVerbEmpty(t, lt, r, "Клочкова ти Лисогор перемагають на своїх дистанціях")
	nounVerbEmpty(t, lt, r, "він особисто й облдержадміністрація винесли")
	nounVerbEmpty(t, lt, r, "тисяча й одна ознака вказують")
	nounVerbEmpty(t, lt, r, "і “більшовики”, і Президент звинуватили опозицію у зриві")
	nounVerbEmpty(t, lt, r, "І “швидка“, і міліція приїхали майже вчасно")
	nounVerbEmpty(t, lt, r, "І уряд, і президент позитивно оцінюють")
	nounVerbEmpty(t, lt, r, "Андрій Ярмоленко, Євген Коноплянка, Ярослав Ракицький допомогли вітчизняній «молодіжці»")
	nounVerbEmpty(t, lt, r, "Мустафа Джемілєв, Рефат Чубаров зможуть")
	nounVerbEmpty(t, lt, r, "Єжи Тур, Александр Рибіцький перебороли")
	nounVerbEmpty(t, lt, r, "27-річний водій та 54-річна пасажирка були травмовані")
	nounVerbEmpty(t, lt, r, "фізична робота та щоденна 30–хвилинна фіззарядка приносять")
	nounVerbEmpty(t, lt, r, "Б. Єльцин і Л. Кучма погодилися вважати Азовське море внутрішнім морем")
	nounVerbEmpty(t, lt, r, "Ґорбачов і його дружина виглядали")
	nounVerbEmpty(t, lt, r, "і Ципкалов, і Кисельов могли одразу")
	nounVerbEmpty(t, lt, r, "«Самопоміч» та Радикальна партія дали більшість голосів")
	nounVerbEmpty(t, lt, r, "Канада, Австралія та й Західна Європа знають")
	nounVerbEmpty(t, lt, r, "процес формування уряду та його відставка залежать")
	nounVerbEmpty(t, lt, r, "Усі розписи, а також архітектура відрізняються")
	nounVerbEmpty(t, lt, r, "Німеччина (ще демократична) та Росія почали «дружити»")
	nounVerbEmpty(t, lt, r, "Як Україна, так і Німеччина відчувають")
	nounVerbEmpty(t, lt, r, "обласної ради, а й уся львівська громада виявилися обманутими")
	nounVerbEmpty(t, lt, r, "Авіація ж і космонавтика справили на розвиток науки")
	nounVerbEmpty(t, lt, r, "І спочатку Білорусь, а тепер і Україна пішли")
	nounVerbEmpty(t, lt, r, "Саме тоді Англія, а невдовзі й уся Європа дізналися")
	nounVerbEmpty(t, lt, r, "але концепція, а потім і програма мають бути")
	nounVerbEmpty(t, lt, r, "Низка трагедій, а потім громадянська війна вигубили чимало люду")
	nounVerbEmpty(t, lt, r, "Наша родина, я особисто і наша політична сила займаємося благодійністю")
	nounVerbEmpty(t, lt, r, "і держава, і, відповідно, посада перестали існувати")
	nounVerbEmpty(t, lt, r, "Швидке заселення земель, вирубування лісів, меліорація призвели")
	nounVerbEmpty(t, lt, r, "узагалі, почуття гумору, іронія були притаманні")
	nounVerbEmpty(t, lt, r, "почуття гумору, іронія були притаманні")
	nounVerbEmpty(t, lt, r, "оскільки назва фільму, прізвище режисера, країна будуть викарбувані на бронзовій плиті")
	nounVerbEmpty(t, lt, r, "Дух Києва, його атмосфера складаються")
	nounVerbEmpty(t, lt, r, "Верховний суд, Рада суддів, Кваліфікаційна комісія могли б перевірити")
	nounVerbEmpty(t, lt, r, "зеленкуваті плями моху, сіро-коричнева бруківка повертають нас на кілька століть назад")
	nounVerbEmpty(t, lt, r, "водій “Мазди” та її пасажир загинули")
	nounVerbEmpty(t, lt, r, "біологічна і ядерна зброя стають товаром")
	nounVerbEmpty(t, lt, r, "І та й інша група вводили в клітини шкіри")
	nounVerbEmpty(t, lt, r, "матч Туреччина — Україна зіграють")
	nounVerbEmpty(t, lt, r, "у смузі Піски—Авдіївка оперують")
	nounVerbEmpty(t, lt, r, "злість плюс іронія можуть вбити")
	nounVerbEmpty(t, lt, r, "із яких 50% плюс одна акція знаходяться")
	nounVerbEmpty(t, lt, r, "Матеріальна заінтересованість плюс гарна вивіска зіграли злий жарт")
	nounVerbEmpty(t, lt, r, "Колесніков/Ахметов посилили")
	nounVerbEmpty(t, lt, r, "решта забороняються")
	nounVerbEmpty(t, lt, r, "все решта відійшло на другий план")
	nounVerbEmpty(t, lt, r, "Більш ніж половина віддали голоси")
	nounVerbEmpty(t, lt, r, "Левова їхня частка працюють через російських туроператорів")
	nounVerbEmpty(t, lt, r, "дві групи з трьох осіб кожна виконували просте завдання")
	nounVerbEmpty(t, lt, r, "Що пачка цигарок, що ковбаса коштують")
	nounVerbEmpty(t, lt, r, "не вулична злочинність, не корупція відлякували")
	nounVerbEmpty(t, lt, r, "з Василем Кричевським Оскар Германович разом брали участь")
	nounVerbEmpty(t, lt, r, "інший нардеп та лідер «Правого сектору» Дмитро Ярош просили")
	nounVerbEmpty(t, lt, r, "пара Катерина Морозова/Дмитро Алексєєв тріумфувала")
}

// TestTokenAgreementNounVerbRule_Num twins TokenAgreementNounVerbRuleTest.testNum.
func TestTokenAgreementNounVerbRule_Num(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "понад тисяча отримали поранення")
	nounVerbEmpty(t, lt, r, "Решта 121 депутат висловилися проти")
	nounVerbEmpty(t, lt, r, "понад сотня отримали поранення")
	nounVerbHasError(t, lt, r, "22 льотчики удостоєно")
	nounVerbEmpty(t, lt, r, "два сини народилося там")
}

// TestTokenAgreementNounVerbRule_MascFem twins TokenAgreementNounVerbRuleTest.testMascFem.
func TestTokenAgreementNounVerbRule_MascFem(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "німецький канцлер зателефонувала російському президенту")
	nounVerbEmpty(t, lt, r, "екс-міністр повторила у телезверненні")
	nounVerbEmpty(t, lt, r, "Прем’єр-міністр повторила у телезверненні")
	nounVerbEmpty(t, lt, r, "Прем’єр—міністр повторила у телезверненні")
	nounVerbEmpty(t, lt, r, "єврокомісар зазначила, що")
	nounVerbEmpty(t, lt, r, "кінолог пояснила")
	nounVerbEmpty(t, lt, r, "автор-упорядник назвала збірник")
	nounVerbHasError(t, lt, r, "Прем’єр-міністр повторило у телезверненні")
	nounVerbHasError(t, lt, r, "приятель повторила у телезверненні")
}

// TestTokenAgreementNounVerbRule_IgnoreByIntent twins TokenAgreementNounVerbRuleTest.testIgnoreByIntent.
func TestTokenAgreementNounVerbRule_IgnoreByIntent(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "тому що воно привнесено ззовні")
	nounVerbEmpty(t, lt, r, "Воно просочено історією")
	nounVerbEmpty(t, lt, r, "Все решта зафіксовано")
	nounVerbEmpty(t, lt, r, "решта зафіксовано")
	nounVerbEmpty(t, lt, r, "З охопленого війною Сектора Газа вивезли «більш як 80 громадян України».")
	nounVerbEmpty(t, lt, r, "подружжя Обама запросило 350 гостей")
}

// TestTokenAgreementNounVerbRule_OverTheWord twins TokenAgreementNounVerbRuleTest.testOverTheWord.
func TestTokenAgreementNounVerbRule_OverTheWord(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "діагноз дизентерія підтвердився")
}

// TestTokenAgreementNounVerbRule_CaseGovernment twins TokenAgreementNounVerbRuleTest.testCaseGovernment.
func TestTokenAgreementNounVerbRule_CaseGovernment(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbEmpty(t, lt, r, "коли українцям пора показати")
	nounVerbEmpty(t, lt, r, "або пропозиція збільшити частку")
}

// TestTokenAgreementNounVerbRule_RuleWithAdjOrKly twins TokenAgreementNounVerbRuleTest.testRuleWithAdjOrKly.
func TestTokenAgreementNounVerbRule_RuleWithAdjOrKly(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)
	nounVerbHasError(t, lt, r, "Ви постане перед вибором")
	nounVerbHasError(t, lt, r, "військовий прибігла")
	nounVerbHasError(t, lt, r, "Фехтувальна збірна було до цього готова")
	nounVerbHasError(t, lt, r, "наші дівчата виграти загальний залік")
	nounVerbHasError(t, lt, r, "окремі ентузіасти переймалася")
	nounVerbHasError(t, lt, r, "угорська влада пообіцяли переглянути")
	nounVerbHasError(t, lt, r, "точно також ви буде платити")
	nounVerbEmpty(t, lt, r, "багато хто в драматурги прийшов")
	nounVerbEmpty(t, lt, r, "з кандидатом у президенти не визначився")
	nounVerbEmpty(t, lt, r, "Мої співвітчизники терпіти це далі не повинні")
	nounVerbEmpty(t, lt, r, "Ви може образились")
	nounVerbEmpty(t, lt, r, "Моя ти зоре в тумані")
	nounVerbEmpty(t, lt, r, "На кожен покладіть по кільцю")
	nounVerbEmpty(t, lt, r, "Любителі фотографувати їжу")
}

// TestTokenAgreementNounVerbRule_RuleDisambigNazInf twins TokenAgreementNounVerbRuleTest.testRuleDisambigNazInf.
func TestTokenAgreementNounVerbRule_RuleDisambigNazInf(t *testing.T) {
	r := rulesuk.NewTokenAgreementNounVerbRule()
	// Java manually builds AnalyzedSentence with SENT_START + noun:p:v_naz + verb:inf
	ss := "SENT_START"
	nPos, nLem := "noun:inanim:p:v_naz", "намагання"
	vPos, vLem := "verb:perf:inf", "спам'ятати"
	readings := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("намагання", &nPos, &nLem), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("спам'ятати", &vPos, &vLem), 0),
	}
	sent := languagetool.NewAnalyzedSentence(readings)
	require.Equal(t, 0, len(r.Match(sent)))
}

// TestTokenAgreementNounVerbRule_SpecialChars twins TokenAgreementNounVerbRuleTest.testSpecialChars.
func TestTokenAgreementNounVerbRule_SpecialChars(t *testing.T) {
	lt, r := wireTokenAgreementNounVerb(t)

	nounVerbEmpty(t, lt, r, "Тарас при\u00ADбіг.")

	ms := r.Match(lt.GetAnalyzedSentence("Тарас при\u00ADбігла."))
	require.Equal(t, 1, len(ms))

	ms = r.Match(lt.GetAnalyzedSentence("Та\u00ADрас прибігла."))
	require.Equal(t, 1, len(ms))
}

// Java TokenAgreementNounVerbRuleTest.GOOD_TEXT
const nounVerbGoodText = "Хоча упродовж десятиліть ширилися численні історії про те, що я був у ряду наступників трону Тембу, щойно наведений простий генеалогічний екскурс викриває міфічність таких тверджень. Я був членом королівської родини, проте не належав до небагатьох привілейованих, що їх виховували на правителів. Натомість мене як нащадка Лівого дому навчали — так само, як і раніше мого батька — бути радником правителів племені. Мій батько був високим темношкірим чоловіком із прямою й величною поставою, яку я, хочеться думати, успадкував. У батька було пасмо білого волосся якраз над чолом, і хлопчиком я бувало брав сірий попіл і втирав його у своє волосся, щоб воно було таке саме, як у тата. Батько мій мав сувору вдачу й не шкодував різки, виховуючи дітей. Він міг бути дивовижно впертим, і це ще одна риса, яка, на жаль, теж могла перейти від батька до сина. Мого батька інколи називали прем’єр-міністром Тембуленду за врядування Далінд’єбо, батька Сабати, який правив на початку 1900-х років, та його сина й наступника Джонгінтаба. Насправді ж такого титулу не існувало, але мій батько справді відіграв роль, яка не надто відрізнялася від функції прем’єра. Як шанований і високо цінований радник обох королів, він супроводжував їх у подорожах і зазвичай перебував поруч із ними на важливих зустрічах із урядовими чиновниками. Він був також визнаним хоронителем історії коса, і частково через це його порадами так дорожили. Моє власне зацікавлення історією прокинулося рано, і батько його підживлював. Він не вмів ні читати, ні писати, але мав репутацію чудового оратора, який захоплював слухачів, розважаючи й водночас повчаючи їх."
