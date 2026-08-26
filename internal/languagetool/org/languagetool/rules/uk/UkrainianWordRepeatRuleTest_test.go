package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/UkrainianWordRepeatRuleTest.java
//
// Java setup:
//   lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   rule = new UkrainianWordRepeatRule(TestTools.getMessages("uk"), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian tagger + official ukrainian.dict (Maven morfologik-ukrainian-lt).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// discoverUkrainianPOSDict finds Java resource /uk/ukrainian.dict
// (morfologik-ukrainian-lt artifact, not always checked into the LT submodule).
func discoverUkrainianPOSDict() string {
	if p := os.Getenv("LANG_UKRAINIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	// Common extract path from morfologik-ukrainian-lt jar (local probe / CI cache).
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
			"src", "main", "resources", "org", "languagetool", "resource", "uk", "ukrainian.dict"),
		"/tmp/uk-dict-fetch/org/languagetool/resource/uk/ukrainian.dict",
	}
	wd, err := os.Getwd()
	if err != nil {
		// still try absolute candidates
		for _, p := range candidates {
			if filepath.IsAbs(p) {
				if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
					return p
				}
			}
		}
		return ""
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
	return ""
}

// wireUKWordRepeatLT ports JLanguageTool(Ukrainian.DEFAULT_VARIANT) analysis for this twin.
func wireUKWordRepeatLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	p := discoverUkrainianPOSDict()
	if p == "" {
		t.Skip("ukrainian.dict not available (set LANG_UKRAINIAN_DICT or place morfologik-ukrainian-lt resource)")
	}
	mt := tagging.OpenMorfologikTagger(p)
	require.NotNil(t, mt, "OpenMorfologikTagger(%s)", p)
	tg := taguk.NewUkrainianTagger(mt)
	lt := languagetool.NewJLanguageTool("uk")
	// Java: language.getTagger().tag(tokens) via getAnalyzedSentence
	lt.SentenceTagger = tg.Tag
	return lt
}

// ukWordRepeatMessages ports TestTools.getMessages("uk") keys used by WordRepeatRule.
func ukWordRepeatMessages() map[string]string {
	return map[string]string{
		"repetition":            "Можлива механічна помилка: повторення слова",
		"desc_repetition":       "Повторення слів (напр., 'буде буде')",
		"desc_repetition_short": "Повторення слів",
	}
}

// TestUkrainianWordRepeatRule_Rule twins UkrainianWordRepeatRuleTest.testRule.
func TestUkrainianWordRepeatRule_Rule(t *testing.T) {
	// Java: new UkrainianWordRepeatRule(TestTools.getMessages("uk"), lt.getLanguage())
	rule := NewUkrainianWordRepeatRule(ukWordRepeatMessages())
	lt := wireUKWordRepeatLT(t)

	assertEmptyMatch := func(text string) {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(text))
		require.Empty(t, matches, "assertEmptyMatch %q", text)
	}

	assertEmptyMatch("без повного розрахунку")
	assertEmptyMatch("без бугіма бугіма")
	assertEmptyMatch("без 100 100")
	assertEmptyMatch("1.30 3.20 3.20")
	assertEmptyMatch("ще в В.Кандинського")
	assertEmptyMatch("Від добра добра не шукають.")
	assertEmptyMatch("Що, що, а кіно в Україні...")
	assertEmptyMatch("Відповідно до ст. ст. 3, 7, 18.")
	assertEmptyMatch("Не можу сказати ні так, ні ні.")

	require.Equal(t, 1, len(rule.Match(lt.GetAnalyzedSentence("без без повного розрахунку"))))

	match := rule.Match(lt.GetAnalyzedSentence("Верховної Ради І і ІІ скликань"))
	require.Equal(t, 1, len(match))
	require.Equal(t, 2, len(match[0].GetSuggestedReplacements()))
	require.Contains(t, match[0].GetSuggestedReplacements(), "І")
	require.Contains(t, match[0].GetSuggestedReplacements(), "I і")
}

// TestUkrainianWordRepeatRule_Metadata checks getId override.
func TestUkrainianWordRepeatRule_Metadata(t *testing.T) {
	r := NewUkrainianWordRepeatRule(ukWordRepeatMessages())
	require.Equal(t, "UKRAINIAN_WORD_REPEAT_RULE", r.GetID())
	require.NotNil(t, r.GetCategory())
}
