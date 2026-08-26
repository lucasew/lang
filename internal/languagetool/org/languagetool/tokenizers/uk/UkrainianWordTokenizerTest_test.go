package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/tokenizers/uk/UkrainianWordTokenizerTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func assertTok(t *testing.T, input string, want ...string) {
	t.Helper()
	w := NewUkrainianWordTokenizer()
	got := w.Tokenize(input)
	require.Equal(t, want, got, "tokenize(%q) got %q", input, got)
}

func TestUkrainianWordTokenizer_TokenizeUrl(t *testing.T) {
	url := "http://youtube.com:80/herewego?start=11&quality=high%3F"
	assertTok(t, url+" ", url, " ")
	url = "http://example.org"
	assertTok(t, " "+url, " ", url)
	url = "www.example.org"
	assertTok(t, url, url)
	url = "elect@ombudsman.gov.ua"
	assertTok(t, url, url)
	assertTok(t, "https://www.foo.com/foo https://youtube.com Зе",
		"https://www.foo.com/foo", " ", "https://youtube.com", " ", "Зе")
	assertTok(t, `https://www.phpbb.com/downloads/">сторінку`,
		"https://www.phpbb.com/downloads/", `"`, ">", "сторінку")
}

func TestUkrainianWordTokenizer_TokenizeTags(t *testing.T) {
	assertTok(t, "<sup>3</sup>", "<sup>", "3", "</sup>")
}

func TestUkrainianWordTokenizer_Numbers(t *testing.T) {
	assertTok(t, "300 грн на балансі", "300", " ", "грн", " ", "на", " ", "балансі")
	assertTok(t, "надійшло 2,2 мільйона", "надійшло", " ", "2,2", " ", "мільйона")
	assertTok(t, "надійшло 84,46 мільйона", "надійшло", " ", "84,46", " ", "мільйона")
	// Java TODO commented: "в 1996,1997,1998"
	assertTok(t, "2 000 тон з 12 000 відер", "2 000", " ", "тон", " ", "з", " ", "12 000", " ", "відер")
	assertTok(t, "надійшло 12 000 000 тон", "надійшло", " ", "12 000 000", " ", "тон")
	assertTok(t, "надійшло 12\u202F000\u202F000 тон", "надійшло", " ", "12 000 000", " ", "тон")
	assertTok(t, "до 01.01.42 400 000 шт.", "до", " ", "01.01.42", " ", "400 000", " ", "шт.")
	assertTok(t, "2 15 мільярдів", "2", " ", "15", " ", "мільярдів")
	assertTok(t, "у 2004 200 мільярдів", "у", " ", "2004", " ", "200", " ", "мільярдів")
	assertTok(t, "в бюджеті-2004 200 мільярдів", "в", " ", "бюджеті-2004", " ", "200", " ", "мільярдів")
	assertTok(t, "з 12 0001 відер", "з", " ", "12", " ", "0001", " ", "відер")
	assertTok(t, "сталося 14.07.2001 вночі", "сталося", " ", "14.07.2001", " ", "вночі")
	assertTok(t, "вчора о 7.30 ранку", "вчора", " ", "о", " ", "7.30", " ", "ранку")
	assertTok(t, "вчора о 7:30 ранку", "вчора", " ", "о", " ", "7:30", " ", "ранку")
	assertTok(t, "3,5-5,6% 7° 7,4°С", "3,5-5,6", "%", " ", "7", "°", " ", "7,4", "°", "С")
	// Java commented: "+400C"
	assertTok(t, "відбулася 17.8.1245", "відбулася", " ", "17.8.1245")
	assertTok(t, "1814.03.09", "1814.03.09")
	// Java DECIMAL_SPACE uses [\h] between digit groups (EN SPACE U+2002 etc).
	// Java only substitutes ' ', NBSP, NNBSP — so U+2002 still splits via SPLIT_CHARS.
	// Match must still accept full \h (left boundary after EN SPACE protects regular spaces).
	assertTok(t, "2\u2002000 тон", "2", "\u2002", "000", " ", "тон")
	assertTok(t, "надійшло 12\u2002000\u2002000 тон",
		"надійшло", " ", "12", "\u2002", "000", "\u2002", "000", " ", "тон")
	// After EN SPACE left boundary, regular thin groups still protect:
	assertTok(t, "x\u200212 000 y", "x", "\u2002", "12 000", " ", "y")
}

func TestUkrainianWordTokenizer_NumbersMissingSpace(t *testing.T) {
	assertTok(t, "від 12 до14 років", "від", " ", "12", " ", "до", "14", " ", "років")
	assertTok(t, "до14-15", "до", "14-15")
	assertTok(t, "Т.Шевченка53", "Т.", "Шевченка", "53")
	// Java commented: "«Тен»103."
	assertTok(t, "«Мак2»", "«", "Мак2", "»")
	assertTok(t, "км2", "км", "2")
	assertTok(t, "Мі17", "Мі", "17")
	assertTok(t, "000ххх000", "000ххх000")
	// Java NUMBER_MISSING_SPACE: [0-9]+(?![letter…]) — lookahead does not consume.
	// Greedy+backtrack: "до14а" matches digits "1" only (next "4" is non-letter), tokens "до","14а".
	// Must not invent ([0-9]+(?:$|[^letter])) which consumes non-letters/digits.
	assertTok(t, "до14а", "до", "14а")
	assertTok(t, "а12б", "а", "12б")
	assertTok(t, "аб2в", "аб2в") // single digit before letter: lookahead fails, no split
	assertTok(t, "тест99тест", "тест", "99тест")
}

func TestUkrainianWordTokenizer_Plus(t *testing.T) {
	assertTok(t, "+20", "+", "20")
	assertTok(t, "-20", "-", "20")
	assertTok(t, "–20", "\u2013", "20")
	assertTok(t, "прислівник+займенник", "прислівник", "+", "займенник")
	assertTok(t, "+займенник", "+", "займенник")
	assertTok(t, "Роттердам+ ", "Роттердам+", " ")
	// Java \\+(?=[а-яіїєґА-ЯІЇЄҐ0-9]) — Latin after + must not split
	assertTok(t, "foo+bar", "foo+bar")
}

func TestUkrainianWordTokenizer_Superscript(t *testing.T) {
	assertTok(t, "дружини¹", "дружини", "¹")
	// Java commented: "км²"
	assertTok(t, "X²", "X²")
}

func TestUkrainianWordTokenizer_Tokenize(t *testing.T) {
	assertTok(t, "Вони прийшли додому.", "Вони", " ", "прийшли", " ", "додому", ".")
	assertTok(t, "Вони прийшли пʼятими зів’ялими.", "Вони", " ", "прийшли", " ", "п'ятими", " ", "зів'ялими", ".")
	// Java commented: combining accents / soft hyphens strip case
	assertTok(t, "я українець(сміється", "я", " ", "українець", "(", "сміється")
	assertTok(t, "ОУН(б) та КП(б)У", "ОУН(б)", " ", "та", " ", "КП(б)У")
	assertTok(t, "Негода є... заступником", "Негода", " ", "є", "...", " ", "заступником")
	assertTok(t, "Запагубили!.. також", "Запагубили", "!..", " ", "також")
	assertTok(t, "Цей графин.", "Цей", " ", "графин", ".")
	assertTok(t, "— Гм.", "—", " ", "Гм", ".")
	assertTok(t, "стін\u00ADку", "стін\u00ADку")
	assertTok(t, "стін\u00AD\nку", "стін\u00AD\nку")
	assertTok(t, `п"яний`, `п"яний`)
	assertTok(t, "▶Трансформація", "▶", "Трансформація")
	assertTok(t, "усмішку😁", "усмішку", "😁")
}

func TestUkrainianWordTokenizer_Initials(t *testing.T) {
	assertTok(t, "Засідав І.Єрмолюк.", "Засідав", " ", "І.", "Єрмолюк", ".")
	assertTok(t, "Засідав І.   Єрмолюк.", "Засідав", " ", "І.", " ", " ", " ", "Єрмолюк", ".")
	assertTok(t, "Засідав І. П. Єрмолюк.", "Засідав", " ", "І.", " ", "П.", " ", "Єрмолюк", ".")
	assertTok(t, "Засідав І.П.Єрмолюк.", "Засідав", " ", "І.", "П.", "Єрмолюк", ".")
	assertTok(t, "І.\u00A0Єрмолюк.", "І.", "\u00A0", "Єрмолюк", ".")
	assertTok(t, "Засідав Єрмолюк І.", "Засідав", " ", "Єрмолюк", " ", "І.")
	assertTok(t, "Засідав Єрмолюк І. П.", "Засідав", " ", "Єрмолюк", " ", "І.", " ", "П.")
	assertTok(t, "Засідав Єрмолюк І. та інші", "Засідав", " ", "Єрмолюк", " ", "І.", " ", "та", " ", "інші")
}

func TestUkrainianWordTokenizer_Abbreviations(t *testing.T) {
	assertTok(t, "140 тис. працівників", "140", " ", "тис.", " ", "працівників")
	assertTok(t, "450 тис. 297 грн", "450", " ", "тис.", " ", "297", " ", "грн")
	assertTok(t, "297 грн...", "297", " ", "грн", "...")
	assertTok(t, "297 грн.", "297", " ", "грн", ".")
	// Java commented: "297 грн.!!!", "297 грн.??"
	assertTok(t, "450 тис.", "450", " ", "тис.")
	assertTok(t, "450 тис.\n", "450", " ", "тис.", "\n")
	assertTok(t, "354\u202Fтис.", "354", "\u202F", "тис.")
	assertTok(t, "911 тис.грн. з бюджету", "911", " ", "тис.", "грн", ".", " ", "з", " ", "бюджету")
	assertTok(t, "за $400\n  тис., здавалося б",
		"за", " ", "$", "400", "\n", " ", " ", "тис.", ",", " ", "здавалося", " ", "б")
	assertTok(t, "найважчого жанру— оповідання", "найважчого", " ", "жанру", "—", " ", "оповідання")
	assertTok(t, "\u2015оповідання", "\u2015", "оповідання")
	assertTok(t, "проф. Артюхов", "проф.", " ", "Артюхов")
	assertTok(t, "чл.-кор. Артюхов", "чл.-кор.", " ", "Артюхов")
	assertTok(t, "проф.\u00A0Артюхов", "проф.", "\u00A0", "Артюхов")
	assertTok(t, "Ів. Франко", "Ів.", " ", "Франко")
	assertTok(t, "кутю\u00A0— щедру", "кутю", "\u00A0", "—", " ", "щедру")
	assertTok(t, "також зав. відділом", "також", " ", "зав.", " ", "відділом")
	assertTok(t, "до н. е.", "до", " ", "н.", " ", "е.")
	assertTok(t, "до н.е.", "до", " ", "н.", "е.")
	// CAP: ABBR_DOT_2_SMALL requires a non-letter prefix char (Java group, not BOS ^ invent).
	// BOS "е.е." cannot match ABBR_DOT_2 (no prefix); bare "е" is also not NON_ENDING → splits.
	assertTok(t, "е.е.", "е", ".", "е", ".")
	// With space prefix → dual-abbr glue (т.ч.).
	assertTok(t, " і т.ч.", " ", "і", " ", "т.", "ч.")
	// second token "м" excluded by (?![смкд]?м.) — dual-abbr does not glue; "п." still NON_ENDING.
	assertTok(t, " 1 п.м.", " ", "1", " ", "п.", "м", ".")
	assertTok(t, "в. о. начальника", "в.", " ", "о.", " ", "начальника")
	assertTok(t, "в.о. начальника", "в.", "о.", " ", "начальника")
	// Java ABBR_DOT_2_SMALL meter exclusion is only (?![смкд]?м\.) — freestanding "мк" is allowed.
	assertTok(t, "до к.мк. щось", "до", " ", "к.", "мк.", " ", "щось")
	assertTok(t, " н.мк.", " ", "н.", "мк.")
	// meter second units excluded from dual-abbr glue; "к." still protected via NON_ENDING list
	assertTok(t, " к.м. x", " ", "к.", "м", ".", " ", "x")
	assertTok(t, " к.см. x", " ", "к.", "см", ".", " ", "x")
	// ABBR_DOT_DASH: Java UNICODE \b — after digit does not glue (digit is word char);
	// bare "К." splits; hyphen is not a SPLIT_CHARS delimiter so stays with following word.
	assertTok(t, "1К.-Святошинський", "1К", ".", "-Святошинський")
	// Java ABBR_DOT_NON_ENDING has dead в(?!\.+): bare mid-sentence/BOS "в."/"В." split (not one token)
	assertTok(t, "слово в. слово", "слово", " ", "в", ".", " ", "слово")
	assertTok(t, "в. слово", "в", ".", " ", "слово")
	assertTok(t, "В. слово", "В", ".", " ", "слово")
	assertTok(t, "100 к.с.", "100", " ", "к.", "с.")
	assertTok(t, "1998 р.н.", "1998", " ", "р.", "н.")
	assertTok(t, "22 коп.", "22", " ", "коп.")
	// Java ABBR_DOT_ENDING / NON_ENDING / PROF / LAT are case-sensitive — no (?i).
	// Uppercase КОП./ДИВ./ПРОФ./ЛАТ. must not glue like lowercase arms.
	// (Avoid "ПРОФ. Name": initials pattern matches trailing "Ф. Name" independently.)
	assertTok(t, "22 КОП.", "22", " ", "КОП", ".")
	assertTok(t, "отримав ДИВ. орден", "отримав", " ", "ДИВ", ".", " ", "орден")
	assertTok(t, "ПРОФ. щось", "ПРОФ", ".", " ", "щось")
	assertTok(t, "від ЛАТ. momento", "від", " ", "ЛАТ", ".", " ", "momento")
	// Explicit dual-case arms still match (Java [Пп]роф)
	assertTok(t, "Проф. Артюхов", "Проф.", " ", "Артюхов")
	assertTok(t, "800 гр. м'яса", "800", " ", "гр.", " ", "м'яса")
	assertTok(t, "18-19 ст.ст. були", "18-19", " ", "ст.", "ст.", " ", "були")
	assertTok(t, "І ст. 11", "І", " ", "ст.", " ", "11")
	assertTok(t, "куб. м", "куб.", " ", "м")
	assertTok(t, "куб.м", "куб.", "м")
	assertTok(t, "ам. долл", "ам.", " ", "долл")
	assertTok(t, "4 дол.", "4", " ", "дол.")
	assertTok(t, "св. ап. Петра", "св.", " ", "ап.", " ", "Петра")
	assertTok(t, "У с. Вижва", "У", " ", "с.", " ", "Вижва")
	assertTok(t, "оз. Вижва", "оз.", " ", "Вижва")
	assertTok(t, "Довжиною 30 см. з гаком.",
		"Довжиною", " ", "30", " ", "см", ".", " ", "з", " ", "гаком", ".")
	assertTok(t, "Довжиною 30 см. Поїхали.",
		"Довжиною", " ", "30", " ", "см", ".", " ", "Поїхали", ".")
	assertTok(t, "100 м. дороги.", "100", " ", "м", ".", " ", "дороги", ".")
	assertTok(t, "в м.Київ", "в", " ", "м.", "Київ")
	assertTok(t, "На висоті 4000 м...", "На", " ", "висоті", " ", "4000", " ", "м", "...")
	assertTok(t, "№47 (м. Слов'янськ)", "№", "47", " ", "(", "м.", " ", "Слов'янськ", ")")
	assertTok(t, "с.-г.", "с.-г.")
	assertTok(t, "100 грн. в банк", "100", " ", "грн", ".", " ", "в", " ", "банк")
	assertTok(t, "таке та ін.", "таке", " ", "та", " ", "ін.")
	assertTok(t, "і т. ін.", "і", " ", "т.", " ", "ін.")
	assertTok(t, "і т.д.", "і", " ", "т.", "д.")
	assertTok(t, "в т. ч.", "в", " ", "т.", " ", "ч.")
	assertTok(t, "до т. зв. сальону", "до", " ", "т.", " ", "зв.", " ", "сальону")
	assertTok(t, "(т. зв. сальон)", "(", "т.", " ", "зв.", " ", "сальон", ")")
	assertTok(t, " і под.", " ", "і", " ", "под.")
	assertTok(t, "Інститут ім. акад. Вернадського.",
		"Інститут", " ", "ім.", " ", "акад.", " ", "Вернадського", ".")
	assertTok(t, "Палац ім. гетьмана Скоропадського.",
		"Палац", " ", "ім.", " ", "гетьмана", " ", "Скоропадського", ".")
	assertTok(t, "від лат. momento", "від", " ", "лат.", " ", "momento")
	assertTok(t, "отримав рос. орден", "отримав", " ", "рос.", " ", "орден")
	assertTok(t, "на 1-кімн. кв. в центрі", "на", " ", "1-кімн.", " ", "кв.", " ", "в", " ", "центрі")
	assertTok(t, "1 кв. км.", "1", " ", "кв.", " ", "км", ".")
	assertTok(t, "Валерій (міліціонер-пародист.\n–  Авт.) стане пародистом.",
		"Валерій", " ", "(", "міліціонер-пародист", ".", "\n", "–", " ", " ", "Авт.", ")", " ", "стане", " ", "пародистом", ".")
	assertTok(t, "Сьогодні (у четвер.  — Ред.), вранці.",
		"Сьогодні", " ", "(", "у", " ", "четвер", ".", " ", " ", "—", " ", "Ред.", ")", ",", " ", "вранці", ".")
	// Java uses assertTrue(contains "Авт.") for this case
	{
		w := NewUkrainianWordTokenizer()
		got := w.Tokenize("Fair trade [«Справедлива торгівля». –    Авт.], який стежить за тим, щоб у країнах")
		require.Contains(t, got, "Авт.")
	}
	assertTok(t, "яку авт. устиг", "яку", " ", "авт.", " ", "устиг")
	assertTok(t, "пише ред. Бойків", "пише", " ", "ред.", " ", "Бойків")
	assertTok(t, "диво з див.", "диво", " ", "з", " ", "див", ".")
	assertTok(t, "диво з див...", "диво", " ", "з", " ", "див", "...")
	assertTok(t, "тел.: 044-425-20-63", "тел.", ":", " ", "044-425-20-63")
	assertTok(t, "с/г", "с/г")
	assertTok(t, "ім.Василя", "ім.", "Василя")
	assertTok(t, "ст.231", "ст.", "231")
	assertTok(t, "2016-2017рр.", "2016-2017", "рр.")
	assertTok(t, "30.04.2010р.", "30.04.2010", "р.")
	assertTok(t, "ні могили 6в. ", "ні", " ", "могили", " ", "6в", ".", " ")
	assertTok(t, "в... одягненому", "в", "...", " ", "одягненому")
	assertTok(t, "10 млн. чоловік", "10", " ", "млн.", " ", "чоловік")
	assertTok(t, "від Таврійської губ.5", "від", " ", "Таврійської", " ", "губ.", "5")
	assertTok(t, "від червоних губ.", "від", " ", "червоних", " ", "губ", ".")
	assertTok(t, "К.-Святошинський", "К.-Святошинський")
	assertTok(t, "К.-Г. Руффман", "К.-Г.", " ", "Руффман")
	assertTok(t, "Рис. 10", "Рис.", " ", "10")
	assertTok(t, "худ. фільм", "худ.", " ", "фільм")
	assertTok(t, "рік нар. невідомий", "рік", " ", "нар.", " ", "невідомий")
	assertTok(t, "нар. 1945", "нар.", " ", "1945")
	assertTok(t, "(1995 р. нар.)", "(", "1995", " ", "р.", " ", "нар.", ")")
	assertTok(t, "нар. бл. 1720", "нар.", " ", "бл.", " ", "1720")
	assertTok(t, "(нар. у серпні 1904)", "(", "нар.", " ", "у", " ", "серпні", " ", "1904", ")")
	assertTok(t, "977 — нар. Кріс Мартін", "977", " ", "—", " ", "нар.", " ", "Кріс", " ", "Мартін")
	assertTok(t, "Ради нар. депутатів", "Ради", " ", "нар.", " ", "депутатів")
	assertTok(t, "нар. арт.", "нар.", " ", "арт", ".")
	assertTok(t, "біля нар. Сумно", "біля", " ", "нар", ".", " ", "Сумно")
	assertTok(t, "- Вибори-2019", "-", " ", "Вибори-2019")
	assertTok(t, "порівн. з англ", "порівн.", " ", "з", " ", "англ")
	// Java commented: "30.04.10р."
	assertTok(t, "поч. 1945 - кін. 1946", "поч.", " ", "1945", " ", "-", " ", "кін.", " ", "1946")
	assertTok(t, "Поч. XX ст.", "Поч.", " ", "XX", " ", "ст.")
	assertTok(t, "Чигиринський пов. Такої губернії", "Чигиринський", " ", "пов.", " ", "Такої", " ", "губернії")
	assertTok(t, "Чигиринський пов.", "Чигиринський", " ", "пов.")
	assertTok(t, "З пов. Горобець", "З", " ", "пов.", " ", "Горобець")
	assertTok(t, "пом. 1994", "пом.", " ", "1994")
}

func TestUkrainianWordTokenizer_Brackets(t *testing.T) {
	assertTok(t, "д[окто]р[ом]", "д[окто]р[ом]")
}

func TestUkrainianWordTokenizer_Apostrophe(t *testing.T) {
	assertTok(t, "’продукти харчування’", "'", "продукти", " ", "харчування", "'")
	assertTok(t, "схема 'гроші'", "схема", " ", "'", "гроші", "'")
	assertTok(t, "('дзеркало')", "(", "'", "дзеркало", "'", ")")
	assertTok(t, "все 'дно піду", "все", " ", "'дно", " ", "піду")
	assertTok(t, "трохи 'дно 'дному сказано", "трохи", " ", "'дно", " ", "'дному", " ", "сказано")
	// APOSTROPHE_BEGIN: Java '(?!дно) is case-sensitive — only lowercase дно stays attached.
	assertTok(t, "'дно", "'дно")
	assertTok(t, "'Дно", "'", "Дно")
	assertTok(t, "'ДНО", "'", "ДНО")
	assertTok(t, "а мо',", "а", " ", "мо'", ",")
	assertTok(t, "підемо'", "підемо", "'")
	assertTok(t, "ЗДОРОВ’Я.", "ЗДОРОВ'Я", ".")
	assertTok(t, "''український''", "''", "український", "''")
	assertTok(t, "'є", "'", "є")
	assertTok(t, "'(є)", "'", "(", "є", ")")
}

func TestUkrainianWordTokenizer_Dash(t *testing.T) {
	assertTok(t, "Кан’-Ка Но Рей", "Кан'-Ка", " ", "Но", " ", "Рей")
	assertTok(t, "і екс-«депутат» вибув", "і", " ", "екс-«депутат»", " ", "вибув")
	assertTok(t, "тих \"200\"-х багато", "тих", " ", "\"200\"-х", " ", "багато")
	assertTok(t, "«діди»-українці", "«діди»-українці")
	// Java commented: "«краб»-переросток"
	assertTok(t, "вересні--жовтні", "вересні", "--", "жовтні")
	assertTok(t, "—У певному", "—", "У", " ", "певному")
	assertTok(t, "-У певному", "-", "У", " ", "певному")
	// Mid-word emdash: Java only pre-splits \u2014 before [\h\v]; letter—letter relies on SPLIT_CHARS.
	assertTok(t, "праця—голова", "праця", "—", "голова")
	assertTok(t, "слово—слово", "слово", "—", "слово")
	assertTok(t, "Людина—", "Людина", "—")
	assertTok(t, "Х–ХІ", "Х", "–", "ХІ")
	assertTok(t, "VII-VIII", "VII", "-", "VIII")
	assertTok(t, "Стрий– ", "Стрий", "–", " ")
	assertTok(t, "фіто– та термотерапії", "фіто–", " ", "та", " ", "термотерапії")
	assertTok(t, " –Виділено", " ", "–", "Виділено")
	// BOS en-dash (no leading space): first rune must gate LEADING_DASH, not text[0] byte
	assertTok(t, "–Виділено", "–", "Виділено")
	assertTok(t, "так,\u2013так", "так", ",", "\u2013", "так")
	// compound with quotes from original Go subset
	assertTok(t, "екс-«депутат»", "екс-«депутат»")
}

func TestUkrainianWordTokenizer_SpecialChars(t *testing.T) {
	// Java maps tokens for display: \n → "\\n", \u00AD → "\\xAD"
	text := "РЕАЛІЗАЦІЇ \u00AD\nСІЛЬСЬКОГОСПОДАРСЬКОЇ"
	w := NewUkrainianWordTokenizer()
	got := w.Tokenize(text)
	mapped := make([]string, len(got))
	for i, s := range got {
		mapped[i] = strings.ReplaceAll(strings.ReplaceAll(s, "\n", "\\n"), "\u00AD", "\\xAD")
	}
	require.Equal(t, []string{"РЕАЛІЗАЦІЇ", " ", "\\xAD", "\\n", "СІЛЬСЬКОГОСПОДАРСЬКОЇ"}, mapped)

	// SOFT_HYPHEN_WRAP: Java (?<!\s)\u00AD\n — succeeds at BOS; fails after whitespace.
	assertTok(t, "\u00AD\nку", "\u00AD\nку")
	assertTok(t, " \u00AD\nку", " ", "\u00AD", "\n", "ку")
	assertTok(t, "а\u00AD\nку", "а\u00AD\nку")
	assertTok(t, "стін\u00AD\nку", "стін\u00AD\nку")

	assertTok(t, "а%його", "а", "%", "його")
	assertTok(t, "5%-го", "5%-го")
	// Java %(?![-\u2013][а-яіїєґ]) — uppercase after - does not suppress % split
	assertTok(t, "5%-Й", "5", "%", "-Й")
	assertTok(t, "5′", "5", "′") // U+2032
	assertTok(t, "'⚪'", "'", "⚪", "'")
}

func TestUkrainianWordTokenizer_TokenizeMarkdown(t *testing.T) {
	assertTok(t, "_60-річний_", "_", "60-річний", "_")
	assertTok(t, "**25 жінок України:**", "**", "25", " ", "жінок", " ", "України", ":", "**")
	assertTok(t, "Веретениця**", "Веретениця", "**")
	assertTok(t, "мові***,", "мові", "***", ",")
	assertTok(t, "*Оренбург", "*", "Оренбург")
	assertTok(t, "з*ясував", "з*ясував")
	assertTok(t, "#робота_редактора", "#робота_редактора")
	assertTok(t, "https://uk.wikipedia.org/wiki/Список_аеропортів_України",
		"https://uk.wikipedia.org/wiki/Список_аеропортів_України")
	assertTok(t, "ОСОБА_5", "ОСОБА_5")
}

func TestUkrainianWordTokenizer_TokenizeWebEntities(t *testing.T) {
	// Java entities list (commented "Київ.proUA.com" skipped)
	entities := []string{
		"Паляниця.Інфо",
		"Житомир.info",
		"Жмеринка.City",
		"Ліга.Life",
		"ЛІГА.net",
		"Точка.net",
		"Цензор.НЕТ",
		"Гайдамака.UA",
		"Тиждень.ua",
		"Срана.юа",
		"Рагу.лі",
		"МК.ru",
		"Лента.Ру",
		"Слух.media",
		"Олігарх.com",
		"блогер.фм",
		"ЗМІ.ck.ua",
		"Закарпаття.depo.ua",
	}
	for _, e := range entities {
		assertTok(t, e, e)
	}
	// Java WEB_ENTITIES: CASE_INSENSITIVE|UNICODE_CHARACTER_CLASS — full Cyrillic/Latin case fold
	assertTok(t, "Паляниця.ІНФО", "Паляниця.ІНФО")
	assertTok(t, "Цензор.НеТ", "Цензор.НеТ")
	assertTok(t, "сайт.ОРГ", "сайт.ОРГ")
	assertTok(t, "тест.нет", "тест.нет")
}
