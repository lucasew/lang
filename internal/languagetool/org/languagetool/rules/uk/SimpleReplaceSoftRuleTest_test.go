package uk_test

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceSoftRuleTest.java
//
// Java setup:
//   private JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   SimpleReplaceSoftRule rule = new SimpleReplaceSoftRule(TestTools.getEnglishMessages(), lt.getLanguage());
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator (multiwords + disambiguation.xml + simple maps).
// External package uk_test avoids import cycle with tagging/disambiguation/uk → rules/uk.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	ukdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/uk"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
	"github.com/stretchr/testify/require"
)

// wireUKSoftLT ports JLanguageTool(Ukrainian.DEFAULT_VARIANT):
// tagger + hybrid disambiguator (required for "у Трускавці." geo reading).
func wireUKSoftLT(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	p := discoverUkrainianPOSDictSoft(t)
	mt := tagging.OpenMorfologikTagger(p)
	require.NotNil(t, mt, "OpenMorfologikTagger(%s)", p)
	tg := taguk.NewUkrainianTagger(mt)
	lt := languagetool.NewJLanguageTool("uk")
	// Java: language.getTagger().tag(tokens) via getAnalyzedSentence
	lt.SentenceTagger = tg.Tag
	// Java: language.getDisambiguator() → UkrainianHybridDisambiguator
	// (pre simple → multiwords → XmlRuleDisambiguator(Ukrainian) + context filters)
	lt.Disambiguator = ukdis.NewUkrainianHybridDisambiguator()
	return lt
}

func discoverUkrainianPOSDictSoft(t *testing.T) string {
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

// TestSimpleReplaceSoftRule_Rule twins SimpleReplaceSoftRuleTest.testRule — every active Java case.
func TestSimpleReplaceSoftRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceSoftRule(TestTools.getEnglishMessages(), lt.getLanguage())
	rule := rulesuk.NewSimpleReplaceSoftRule(nil)
	lt := wireUKSoftLT(t)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Ці рядки повинні збігатися."))))

	// Java: geo Трускавець after XML disambig — not soft "трускавка"→полуниця
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("у Трускавці."))))

	// Java isTokenException("завидна") — lemma may be завидний in soft map
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("завидна"))))

	matches := rule.Match(lt.GetAnalyzedSentence("Цей брелок"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"дармовис"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("Не знайде спасіння."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"рятування", "рятунок", "порятунок", "визволення"}, matches[0].GetSuggestedReplacements())
	// Java: assertTrue(matches[0].getMessage().contains(": релігія"))
	require.Contains(t, matches[0].GetMessage(), ": релігія")
}

// TestSimpleReplaceSoftRule_RuleForDerivats twins SimpleReplaceSoftRuleTest.testRuleForDerivats.
// Surface "Підключивши" is a key in replace_soft.txt (not SimpleReplaceRule derivats path).
func TestSimpleReplaceSoftRule_RuleForDerivats(t *testing.T) {
	rule := rulesuk.NewSimpleReplaceSoftRule(nil)
	lt := wireUKSoftLT(t)

	matches := rule.Match(lt.GetAnalyzedSentence("Підключивши"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Увімкнувши", "Під'єднавши", "Приєднавши"}, matches[0].GetSuggestedReplacements())
}

// TestSimpleReplaceSoftRule_Meta twins Java id / description / short / Style ITS.
func TestSimpleReplaceSoftRule_Meta(t *testing.T) {
	rule := rulesuk.NewSimpleReplaceSoftRule(nil)
	require.Equal(t, "UK_SIMPLE_REPLACE_SOFT", rule.GetID())
	require.Equal(t, "Пошук нерекомендованих слів", rule.GetDescription())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
}
