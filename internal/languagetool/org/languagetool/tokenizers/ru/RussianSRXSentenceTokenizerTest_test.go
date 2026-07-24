package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/tokenizers/ru/RussianSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(Russian.getInstance())  // short code "ru"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Russian.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitRU mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitRU(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewRussianSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of RussianSRXSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestRussianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	// From the Russian abbreviation list:
	testSplitRU(t, "Отток капитала из России составил 7 млрд. долларов, сообщил министр финансов Алексей Кудрин.")
	testSplitRU(t, "Журнал издаётся с 1967 г., пользуется большой популярностью в мире.")
	testSplitRU(t, "С 2007 г. периодичность выхода газеты – 120 раз в год.")
	testSplitRU(t, "Редакция журнала находится в здании по адресу: г. Москва, 110000, улица Мира, д. 1.")
	testSplitRU(t, "Все эти вопросы заставляют нас искать ответы в нашей истории 60-80-х гг. прошлого столетия.")
	testSplitRU(t, "Более 300 тыс. документов и справочников.")
	testSplitRU(t, "Скидки до 50000 руб. на автомобили.")
	testSplitRU(t, "Изготовление визиток любыми тиражами (от 20 шт. до 10 тысяч) в минимальные сроки (от 20 минут).")
	testSplitRU(t, "Временно не работает, т.к. не поддерживается.")
}
