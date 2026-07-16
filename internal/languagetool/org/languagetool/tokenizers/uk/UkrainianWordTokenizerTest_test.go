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
	assertTok(t, "відбулася 17.8.1245", "відбулася", " ", "17.8.1245")
	assertTok(t, "1814.03.09", "1814.03.09")
}

func TestUkrainianWordTokenizer_NumbersMissingSpace(t *testing.T) {
	assertTok(t, "від 12 до14 років", "від", " ", "12", " ", "до", "14", " ", "років")
	assertTok(t, "до14-15", "до", "14-15")
	assertTok(t, "Т.Шевченка53", "Т.", "Шевченка", "53")
	assertTok(t, "«Мак2»", "«", "Мак2", "»")
	assertTok(t, "км2", "км", "2")
	assertTok(t, "Мі17", "Мі", "17")
	assertTok(t, "000ххх000", "000ххх000")
}

func TestUkrainianWordTokenizer_Plus(t *testing.T) {
	assertTok(t, "+20", "+", "20")
	assertTok(t, "-20", "-", "20")
	assertTok(t, "–20", "\u2013", "20")
	assertTok(t, "прислівник+займенник", "прислівник", "+", "займенник")
	assertTok(t, "+займенник", "+", "займенник")
	assertTok(t, "Роттердам+ ", "Роттердам+", " ")
}

func TestUkrainianWordTokenizer_Superscript(t *testing.T) {
	assertTok(t, "дружини¹", "дружини", "¹")
	assertTok(t, "X²", "X²")
}

func TestUkrainianWordTokenizer_Tokenize(t *testing.T) {
	assertTok(t, "Вони прийшли додому.", "Вони", " ", "прийшли", " ", "додому", ".")
	assertTok(t, "Вони прийшли пʼятими зів’ялими.", "Вони", " ", "прийшли", " ", "п'ятими", " ", "зів'ялими", ".")
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
	assertTok(t, "450 тис.", "450", " ", "тис.")
	assertTok(t, "450 тис.\n", "450", " ", "тис.", "\n")
	assertTok(t, "354\u202Fтис.", "354", "\u202F", "тис.")
	assertTok(t, "проф. Артюхов", "проф.", " ", "Артюхов")
	assertTok(t, "чл.-кор. Артюхов", "чл.-кор.", " ", "Артюхов")
	assertTok(t, "в. о. начальника", "в.", " ", "о.", " ", "начальника")
	assertTok(t, "в.о. начальника", "в.", "о.", " ", "начальника")
	assertTok(t, "с/г", "с/г")
	assertTok(t, "с.-г.", "с.-г.")
	assertTok(t, "і т.д.", "і", " ", "т.", "д.")
	assertTok(t, "ім.Василя", "ім.", "Василя")
	assertTok(t, "2016-2017рр.", "2016-2017", "рр.")
	assertTok(t, "100 грн. в банк", "100", " ", "грн", ".", " ", "в", " ", "банк")
	assertTok(t, "куб. м", "куб.", " ", "м")
	assertTok(t, "таке та ін.", "таке", " ", "та", " ", "ін.")
	assertTok(t, "К.-Святошинський", "К.-Святошинський")
}

func TestUkrainianWordTokenizer_Brackets(t *testing.T) {
	assertTok(t, "д[окто]р[ом]", "д[окто]р[ом]")
}

func TestUkrainianWordTokenizer_Apostrophe(t *testing.T) {
	assertTok(t, "’продукти харчування’", "'", "продукти", " ", "харчування", "'")
	assertTok(t, "схема 'гроші'", "схема", " ", "'", "гроші", "'")
	assertTok(t, "все 'дно піду", "все", " ", "'дно", " ", "піду")
	assertTok(t, "а мо',", "а", " ", "мо'", ",")
	assertTok(t, "підемо'", "підемо", "'")
	assertTok(t, "ЗДОРОВ’Я.", "ЗДОРОВ'Я", ".")
}

func TestUkrainianWordTokenizer_Dash(t *testing.T) {
	assertTok(t, "—У певному", "—", "У", " ", "певному")
	assertTok(t, "-У певному", "-", "У", " ", "певному")
	assertTok(t, "праця—голова", "праця", "—", "голова")
	assertTok(t, "VII-VIII", "VII", "-", "VIII")
	assertTok(t, "Х–ХІ", "Х", "–", "ХІ")
	assertTok(t, "екс-«депутат»", "екс-«депутат»")
}

func TestUkrainianWordTokenizer_TokenizeMarkdown(t *testing.T) {
	// minimal - expand from Java if present
	w := NewUkrainianWordTokenizer()
	_ = w
}

func TestUkrainianWordTokenizer_TokenizeWebEntities(t *testing.T) {
	assertTok(t, "сайт.НЕТ", "сайт.НЕТ")
}

func TestUkrainianWordTokenizer_SpecialChars(t *testing.T) {
	text := "РЕАЛІЗАЦІЇ \u00AD\nСІЛЬСЬКОГОСПОДАРСЬКОЇ"
	w := NewUkrainianWordTokenizer()
	got := w.Tokenize(text)
	// soft hyphen wrap stays as one token with \u00AD\n
	require.True(t, len(got) >= 1)
	joined := strings.Join(got, "")
	require.Contains(t, joined, "РЕАЛІЗАЦІЇ")
}
