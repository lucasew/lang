package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/UppercaseSentenceStartRuleTest.java
//
// Java setup:
//   Ukrainian ukrainian = Ukrainian.DEFAULT_VARIANT;
//   UkrainianUppercaseSentenceStartRule rule = new UkrainianUppercaseSentenceStartRule(
//       TestTools.getEnglishMessages(), ukrainian,
//       Example.wrong("<marker>речення</marker> має починатися з великої."),
//       Example.fixed("<marker>Речення</marker> має починатися з великої"));
//   JLanguageTool lt = new JLanguageTool(ukrainian);
//   cases 1–4: rule.match(lt.analyzeText(...))
//   cases 5–8: lt.check(...) → empty overall; twin asserts this TextLevelRule’s MatchList(Analyze) empty.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// engMessagesUppercaseUK ports TestTools.getEnglishMessages keys used by UppercaseSentenceStartRule.
func engMessagesUppercaseUK() map[string]string {
	return map[string]string{
		"incorrect_case":          "This sentence does not start with an uppercase letter",
		"desc_uppercase_sentence": "Checks that a sentence starts with an uppercase letter",
		"category_case":           "Capitalization",
	}
}

// wireUKUppercaseLT ports JLanguageTool(Ukrainian.DEFAULT_VARIANT) for analyzeText/check:
// Ukrainian SRX (LanguageCode "uk"), UkrainianWordTokenizer, IGNORED_CHARS, tagger when dict present.
func wireUKUppercaseLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	lt := languagetool.NewJLanguageTool("uk")
	// Java Ukrainian.IGNORED_CHARS: soft hyphen + combining acute (водопі́й).
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex
	if p := discoverUkrainianPOSDictUpper(); p != "" {
		mt := tagging.OpenMorfologikTagger(p)
		require.NotNil(t, mt, "OpenMorfologikTagger(%s)", p)
		lt.SentenceTagger = taguk.NewUkrainianTagger(mt).Tag
	}
	return lt
}

func discoverUkrainianPOSDictUpper() string {
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
	// Uppercase rule itself does not require POS; SRX+tokenizer is enough for this twin.
	// Prefer real tagger when available (Java JLanguageTool path).
	return ""
}

// Port of UppercaseSentenceStartRuleTest.testUkrainian — full outcome matrix.
func TestUppercaseSentenceStartRule_Ukrainian(t *testing.T) {
	// Java: new UkrainianUppercaseSentenceStartRule(TestTools.getEnglishMessages(), ukrainian, examples…)
	rule := NewUkrainianUppercaseSentenceStartRule(engMessagesUppercaseUK())
	lt := wireUKUppercaseLT(t)

	// 1. capitalized start → 0
	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("Автор написав це речення з великої літери."))))

	// 2. abbreviation с. (село) — no false split / no match
	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("Він приїхав в с. Вижівка."))))

	// 3. URL with leading spaces → 0
	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("         http://narodna.pravda.com.ua"))))

	// 4. lowercase start → 1 match, suggestion "Автор"
	matches := rule.MatchList(lt.Analyze("автор написав це речення з маленької літери."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "Автор", matches[0].GetSuggestedReplacements()[0])

	// 5–8: Java lt.check → empty. Twin asserts this text-level rule contributes no matches
	// (MatchList on full Analyze = same path as TextLevelRule.match in Check).
	require.Empty(t, rule.MatchList(lt.Analyze("Цей список з декількох рядків:\n\nрядок 1,\n\nрядок 2,\n\nрядок 3.")))
	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("Цей список з декількох рядків:\n\nрядок 1;\n\nрядок 2;\n\nрядок 3."))))
	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("Цей список з декількох рядків:\n\n 1) рядок 1;\n\n2) рядок 2;\n\n3)рядок 3."))))
	require.Empty(t, rule.MatchList(lt.Analyze("а) водопі́й")))
}
