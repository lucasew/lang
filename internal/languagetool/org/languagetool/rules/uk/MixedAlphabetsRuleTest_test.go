package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/MixedAlphabetsRuleTest.java
// POS/lemma inject for prep, fname, and гепатит|група|турнір arms (Java FreeLing tags).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMixedAlphabetsRule_Rule(t *testing.T) {
	rule := NewMixedAlphabetsRule(nil)
	assert0 := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "good %q", s)
	}
	assert1 := func(s, msg string, suggs ...string) {
		t.Helper()
		m := rule.Match(analyzeMixed(s))
		require.Equal(t, 1, len(m), "bad %q got %d", s, len(m))
		if msg != "" {
			require.Equal(t, msg, m[0].GetMessage(), "msg %q", s)
		}
		if len(suggs) > 0 {
			require.Equal(t, suggs, m[0].GetSuggestedReplacements(), "sugg %q", s)
		}
	}

	assert0("сміття")
	assert0("not mixed")
	assert0("123454")
	assert0("x = a якщо")
	assert0("x − a та y − b")
	assert0("записати x та y через параметр t")
	assert0("ЛЮДИНИ І НАЦІЇ")

	assert1("смiття", "Вжито кириличні й латинські літери в одному слові", "сміття")
	assert1("mіхed", "Вжито кириличні літери замість латинських", "mixed")
	assert1("горíти", "Вжито кириличні й латинські літери в одному слові", "горі́ти")
	assert1("двоáктний", "Вжито кириличні й латинські літери в одному слові")
	assert1("Чорного i Азовського", "Вжито латинську «i» замість кириличної", "і")
	assert1("A нема", "Вжито латинську «A» замість кириличної", "А")

	// capitalized name before І (LemmaHelper.isCapitalized)
	assert1("Петро І", "Вжито кириличну літеру замість латинської", "I")
	// І. needs fname POS (not :abbr)
	assert1("Миколая І.", "Вжито кириличну літеру замість латинської", "I.")
	// next surface "квартал" / "ст."
	assert1("У І кварталі", "Вжито кириличну літеру замість латинської", "I")
	assert0("ЗА І ПРОТИ")
	assert0("Ленін В. І.")
	assert0("Тому І.    Вишенський радить ")
	assert1("у І ст.", "Вжито кириличну літеру замість латинської", "I")

	assert1("XІ", "Вжито кириличні літери замість латинських", "XI")
	assert1("ХI", "Вжито кириличні літери замість латинських", "XI")
	assert1("VIIІ-го", "Вжито кириличні літери замість латинських. Також: до римських цифр букви не дописуються.", "VIII")
	assert1("ІІІ-го", "Вжито кириличні літери замість латинських на позначення римської цифри. Також: до римських цифр букви не дописуються.", "III")
	assert1("ХІ", "Вжито кириличні літери замість латинських на позначення римської цифри", "XI")
	assert1("СOVID-19", "Вжито кириличні літери замість латинських", "COVID-19")
	assert1("австрo-турецької", "Вжито кириличні й латинські літери в одному слові", "австро-турецької")

	// lemma гепатит / група on first reading
	assert1("Щеплення від гепатиту В.", "Вжито кириличну літеру замість латинської", "B")
	assert1("група А", "Вжито кириличну літеру замість латинської", "A")
	assert1("На 0,6°С.", "Вжито кириличну літеру замість латинської", "C")
}

func TestMixedAlphabetsRule_FailClosedWithoutLemmaPOS(t *testing.T) {
	rule := NewMixedAlphabetsRule(nil)
	// no lemma on гепатиту → no common-cyr group match
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Щеплення від гепатиту В.")))
	// no fname POS → no І. after name
	// "Миколая І." without tags: І + . may still join; fname arm fails closed
	// (capitalized arm only for bare І, not І.)
	m := rule.Match(languagetool.AnalyzePlain("Миколая І."))
	// may be 0 if only fname arm applied
	for _, rm := range m {
		require.NotContains(t, rm.GetSuggestedReplacements(), "I.", "fname arm without POS")
	}
}

func TestMixedAlphabetsRule_CombiningChars(t *testing.T) {
	rule := NewMixedAlphabetsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Білоруський - українці"))
	require.Equal(t, 2, len(matches))
	require.Equal(t, "Білоруський", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, "українці", matches[1].GetSuggestedReplacements()[0])
}

// analyzeMixed injects FreeLing-style tags for MixedAlphabets POS/lemma arms.
func analyzeMixed(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		low := strings.ToLower(tok)
		switch {
		case low == "миколая" || low == "петро":
			return []languagetool.TokenTag{{POS: "noun:m:v_naz:fname", Lemma: low}}
		case low == "у" || low == "від" || low == "в":
			return []languagetool.TokenTag{{POS: "prep", Lemma: low}}
		case strings.HasPrefix(low, "гепатит"):
			return []languagetool.TokenTag{{POS: "noun:m:v_rod", Lemma: "гепатит"}}
		case low == "група":
			return []languagetool.TokenTag{{POS: "noun:f:v_naz", Lemma: "група"}}
		case low == "турнір":
			return []languagetool.TokenTag{{POS: "noun:m:v_naz", Lemma: "турнір"}}
		default:
			return nil
		}
	})
}
