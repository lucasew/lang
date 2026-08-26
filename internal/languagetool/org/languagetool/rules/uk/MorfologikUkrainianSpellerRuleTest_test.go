package uk_test

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/MorfologikUkrainianSpellerRuleTest.java
//
// Java setup:
//   rule = new MorfologikUkrainianSpellerRule(TestTools.getMessages("uk"), Ukrainian.DEFAULT_VARIANT, null, emptyList());
//   lt = AbstractRuleTest → new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
// Official resources: /uk/hunspell/uk_UA.dict + ukrainian.dict (POS) + hybrid disambiguator.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	ukdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/uk"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// requireUkrainianSpellerDict skips when official uk_UA.dict is not discoverable (no invent).
func requireUkrainianSpellerDict(t *testing.T) *rulesuk.MorfologikUkrainianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(rulesuk.UkrainianSpellerDict) == "" {
		t.Skip("uk_UA.dict not discoverable (set LANG_UK_SPELLER_DICT or /tmp/uk-dict-fetch/.../uk_UA.dict)")
	}
	r := rulesuk.NewMorfologikUkrainianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	require.Equal(t, rulesuk.MorfologikUkrainianSpellerRuleID, r.GetID())
	require.Equal(t, rulesuk.UkrainianSpellerDict, r.GetFileName())
	return r
}

func discoverUkrainianPOSDictSpeller(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("LANG_UKRAINIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
			"src", "main", "resources", "org", "languagetool", "resource", "uk", "ukrainian.dict"),
		"/tmp/uk-dict-fetch/org/languagetool/resource/uk/ukrainian.dict",
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("ukrainian.dict not available")
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
	t.Skip("ukrainian.dict not available (set LANG_UKRAINIAN_DICT or place morfologik-ukrainian-lt resource)")
	return ""
}

// wireUKSpellerLT ports JLanguageTool(Ukrainian.DEFAULT_VARIANT) for this twin:
// UkrainianWordTokenizer + UkrainianTagger + UkrainianHybridDisambiguator + IGNORED_CHARS.
func wireUKSpellerLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	p := discoverUkrainianPOSDictSpeller(t)
	mt := tagging.OpenMorfologikTagger(p)
	require.NotNil(t, mt, "OpenMorfologikTagger(%s)", p)
	tg := taguk.NewUkrainianTagger(mt)
	lt := languagetool.NewJLanguageTool("uk")
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex
	lt.SentenceTagger = tg.Tag
	lt.Disambiguator = ukdis.NewUkrainianHybridDisambiguator()
	return lt
}

func matchUK(t *testing.T, rule *rulesuk.MorfologikUkrainianSpellerRule, lt *languagetool.JLanguageTool, text string) []*rules.RuleMatch {
	t.Helper()
	ms, err := rule.Match(lt.GetAnalyzedSentence(text))
	require.NoError(t, err)
	return ms
}

func utf16LenJava(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func indexOfUTF16(s, sub string) int {
	i := strings.Index(s, sub)
	if i < 0 {
		return -1
	}
	return utf16LenJava(s[:i])
}

// Port of MorfologikUkrainianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikUkrainianSpellerRule_MorfologikSpeller(t *testing.T) {
	rule := requireUkrainianSpellerDict(t)
	lt := wireUKSpellerLT(t)

	matchN := func(s string) int {
		t.Helper()
		return len(matchUK(t, rule, lt, s))
	}
	assertEmpty := func(s string) {
		t.Helper()
		require.Equal(t, 0, matchN(s), "assertEmptyMatch %q", s)
	}

	// correct sentences:
	assertEmpty("До вас прийде заввідділу!")
	assertEmpty(",")
	assertEmpty("123454")
	assertEmpty("До нас приїде The Beatles!")

	// soft hyphen
	assertEmpty("піс\u00ADні")
	assertEmpty("піс\u00ADні піс\u00ADні")

	// non-breaking hyphen
	assertEmpty("ось\u2011ось")

	// frequent infix notation
	assertEmpty("-ськ-")

	require.Equal(t, 1, matchN("Халгін-Гол"))

	// accent
	require.Equal(t, 0, matchN("Іва́н Петро́вич Котляре́вський"))
	require.Equal(t, 0, matchN("1998 ро́ку"))

	assertEmpty("а")

	// mix alphabets — latin 'i'
	ms := matchUK(t, rule, lt, "прийдешнiй")
	require.Equal(t, 1, len(ms))
	require.NotEmpty(t, ms[0].GetSuggestedReplacements())
	require.Equal(t, "прийдешній", ms[0].GetSuggestedReplacements()[0])

	// кличний для неістот
	assertEmpty("душе")

	// incorrect v_kly could be removed by disambig with no tags left
	require.Equal(t, 1, matchN("польовою дружино"))

	// abbreviations
	require.Equal(t, 0, matchN("Читання віршів Т.Г.Шевченко і Г.Тютюнника"))
	require.Equal(t, 0, matchN("Читання віршів Т. Г. Шевченко і Г. Тютюнника"))
	require.Equal(t, 0, matchN("Англі́йська мова (англ. English language, English) належить до германської групи"))
	require.Equal(t, 1, matchN("Англі́йська мова (англ English language, English) належить до германської групи"))

	require.Equal(t, 0, matchN("100 тис. гривень"))
	require.Equal(t, 0, matchN("100 кв. м"))
	require.Equal(t, 0, matchN("100 км²"))
	require.Equal(t, 1, matchN("100 кв м"))

	require.Equal(t, 0, matchN("півтора раза"))

	require.Equal(t, 1, matchN("УКРА"))

	sent := "Іва\u0301н Петро\u0301ввич"
	ms = matchUK(t, rule, lt, sent)
	require.Equal(t, 1, len(ms))
	require.Equal(t, indexOfUTF16(sent, "Петро"), ms[0].GetFromPos())
	require.Equal(t, utf16LenJava(sent), ms[0].GetToPos())

	sent = "голага́нівська"
	ms = matchUK(t, rule, lt, sent)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, utf16LenJava(sent), ms[0].GetToPos())

	assertEmpty("Добрий ранок, Україно!")
}

// Port of MorfologikUkrainianSpellerRuleTest.testSuggestionOrder
func TestMorfologikUkrainianSpellerRule_SuggestionOrder(t *testing.T) {
	rule := requireUkrainianSpellerDict(t)
	lt := wireUKSpellerLT(t)
	ms := matchUK(t, rule, lt, "захворіва-є")
	require.Equal(t, 1, len(ms))
	sugs := ms[0].GetSuggestedReplacements()
	require.GreaterOrEqual(t, len(sugs), 5)
	require.Equal(t, []string{"захворювань", "захворів", "захворіла", "захворіє", "захворював"}, sugs[:5])
}

// Port of MorfologikUkrainianSpellerRuleTest.testCompounds
func TestMorfologikUkrainianSpellerRule_Compounds(t *testing.T) {
	rule := requireUkrainianSpellerDict(t)
	lt := wireUKSpellerLT(t)
	matchN := func(s string) int {
		t.Helper()
		return len(matchUK(t, rule, lt, s))
	}
	require.Equal(t, 0, matchN("Жакет був синьо-жовтого кольору"))
	require.Equal(t, 0, matchN("Він багато сидів на інтернет-форумах"))
	require.Equal(t, 1, matchN("Він багато сидів на інтермет-форумах"))
	// dynamic tagging
	require.Equal(t, 0, matchN("екс-креветка"))
	require.Equal(t, 1, matchN("банд-формування."))
	require.Equal(t, 0, matchN("екоблогер"))
}

// Port of MorfologikUkrainianSpellerRuleTest.testDashedSuggestions
func TestMorfologikUkrainianSpellerRule_DashedSuggestions(t *testing.T) {
	// Java uses Ukrainian.DEFAULT_VARIANT.getDefaultSpellingRule() — same New.
	rule := requireUkrainianSpellerDict(t)
	lt := wireUKSpellerLT(t)
	ms := matchUK(t, rule, lt, "блоксистема")
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"блок система", "блок-система"}, ms[0].GetSuggestedReplacements())
	// Java TODO: show-men commented out — not asserted
	_ = matchUK(t, rule, lt, "шоу-мен")
}

// Port of MorfologikUkrainianSpellerRuleTest.testProhibitedSuggestions
func TestMorfologikUkrainianSpellerRule_ProhibitedSuggestions(t *testing.T) {
	rule := requireUkrainianSpellerDict(t)
	lt := wireUKSpellerLT(t)

	ms := matchUK(t, rule, lt, "онлайннавчання")
	require.Equal(t, 1, len(ms))

	ms = matchUK(t, rule, lt, "авіабегемат")
	require.Equal(t, 1, len(ms))
	require.Empty(t, ms[0].GetSuggestedReplacements(), "Should be empty: %v", ms[0].GetSuggestedReplacements())

	// Java asserts empty twice on same avia match, then continues:
	require.Empty(t, ms[0].GetSuggestedReplacements(), "Unexpected suggestions: %v", ms[0].GetSuggestedReplacements())

	require.Equal(t, 0, len(matchUK(t, rule, lt, "радіо- та відеоспостереження")))
	require.Equal(t, 1, len(matchUK(t, rule, lt, "радіо- засоби")))
}
