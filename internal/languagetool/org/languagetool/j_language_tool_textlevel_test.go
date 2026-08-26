package languagetool_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/stretchr/testify/require"
)

func TestEnglishCore_WordRepeatBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	en.RegisterCoreEnglishLanguageRules(lt)
	// Inject PRP on "I" (Java EnglishTagger); getSuggestions is PRP-gated (no surface invent).
	lt.TagWord = func(token string) []languagetool.TokenTag {
		if token == "I" {
			return []languagetool.TokenTag{{POS: "PRP", Lemma: "I"}}
		}
		return nil
	}
	// three successive "I" starts → text-level match
	m := lt.Check("I think so. I have seen that before. I don't like it.")
	found := false
	for _, x := range m {
		if x.RuleID == "ENGLISH_WORD_REPEAT_BEGINNING_RULE" {
			found = true
			require.NotEmpty(t, x.Suggestions)
		}
	}
	require.True(t, found, "matches: %+v", m)

	// mode excludes text-level
	lt.SetMode(languagetool.ModeAllButTextLevel)
	m2 := lt.Check("I think so. I have seen that before. I don't like it.")
	for _, x := range m2 {
		require.NotEqual(t, "ENGLISH_WORD_REPEAT_BEGINNING_RULE", x.RuleID)
	}
}

func TestCorepack_AR_SL(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ar")
	corepack.Register(lt, "ar")
	require.NotEmpty(t, lt.Check("a  b"))
	// Arabic word double
	m := lt.Check("كلمة كلمة")
	found := false
	for _, x := range m {
		if x.RuleID == "ARABIC_WORD_REPEAT_RULE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)

	lt2 := languagetool.NewJLanguageTool("sl")
	corepack.Register(lt2, "sl")
	m2 := lt2.Check("test test")
	found2 := false
	for _, x := range m2 {
		// Java Slovenian.getRelevantRules uses WordRepeatRule → id WORD_REPEAT_RULE
		// (not invent SL_WORD_REPEAT_RULE).
		if x.RuleID == "WORD_REPEAT_RULE" {
			found2 = true
		}
	}
	require.True(t, found2, "%+v", m2)
}

func TestCachedCheck_MultiSentence(t *testing.T) {
	cache := languagetool.NewResultCache(32)
	lt := languagetool.NewJLanguageTool("en")
	corepack.Register(lt, "en")
	text := "This is an test. Hello  world."
	a := languagetool.CachedCheck(cache, lt, text)
	b := languagetool.CachedCheck(cache, lt, text)
	require.Equal(t, a, b)
	require.NotEmpty(t, a)
}
