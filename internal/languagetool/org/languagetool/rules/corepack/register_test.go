package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_MultiLang(t *testing.T) {
	cases := []struct {
		lang string
		text string
		id   string
	}{
		{"en", "This is an test.", "EN_A_VS_AN"},
		{"de", "Ein Test Test.", "GERMAN_WORD_REPEAT_RULE"},
		{"sv", "hej hej", "WORD_REPEAT_RULE"},
		// da: Java WORD_REPEAT only via grammar.xml — no class WordRepeat in getRelevantRules
		// gl/nl: Java has no WordRepeat — covered in pack tests / not MultiLang
		{"sk", "test test", "WORD_REPEAT_RULE"},
		{"el", "γεια γεια", "WORD_REPEAT_RULE"},
		{"ro", "test test", "WORD_REPEAT_RULE"},
		{"pt-BR", "teste teste", "PORTUGUESE_WORD_REPEAT_RULE"},
		{"ar", "كلمة كلمة", "ARABIC_WORD_REPEAT_RULE"},
		{"sl", "test test", "WORD_REPEAT_RULE"},
		// br: Java Breton.getRelevantRules has no WordRepeat — covered in br pack tests
		{"fa", "test test", "PERSIAN_WORD_REPEAT_RULE"},
		{"ga", "test test", "WORD_REPEAT_RULE"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(tc.lang)
			corepack.Register(lt, tc.lang)
			require.NotEmpty(t, lt.Check("a  b"))
			m := lt.Check(tc.text)
			require.NotEmpty(t, m, tc.lang)
			found := false
			for _, x := range m {
				if x.RuleID == tc.id {
					found = true
					break
				}
			}
			require.True(t, found, "want %s in %+v", tc.id, m)
		})
	}
}

// Java French has no WordRepeatBeginning — only FrenchRepeatedWordsRule (style).
func TestRegister_French_NoInventWordRepeat(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	corepack.Register(lt, "fr")
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "FR_WORD_REPEAT_RULE")
	require.NotContains(t, ids, "FR_WORD_REPEAT_BEGINNING_RULE")
	require.NotContains(t, ids, "WORD_REPEAT_RULE")
	for _, m := range lt.Check("test test") {
		require.NotContains(t, m.RuleID, "WORD_REPEAT")
	}
}

func TestRegister_GenericPacks(t *testing.T) {
	// eo/sr still have word-repeat; be has no Java WordRepeat (replace/speller only)
	for _, code := range []string{"eo", "sr"} {
		t.Run(code, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(code)
			corepack.Register(lt, code)
			require.NotEmpty(t, lt.Check("a  b"))
			m := lt.Check("test test")
			require.NotEmpty(t, m)
		})
	}
	t.Run("be", func(t *testing.T) {
		lt := languagetool.NewJLanguageTool("be")
		corepack.Register(lt, "be")
		require.NotEmpty(t, lt.Check("a  b"))
		// No invent word-repeat
		for _, m := range lt.Check("test test") {
			require.NotContains(t, m.RuleID, "WORD_REPEAT")
		}
	})
}

// Java Chinese/Japanese.getRelevantRules: DOUBLE_PUNCTUATION + WHITESPACE_RULE only.
func TestRegister_JapaneseChinese_NoInventWordRepeat(t *testing.T) {
	for _, code := range []string{"ja", "zh"} {
		t.Run(code, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(code)
			corepack.Register(lt, code)
			ids := lt.GetAllRegisteredRuleIDs()
			require.ElementsMatch(t, []string{"DOUBLE_PUNCTUATION", "WHITESPACE_RULE"}, ids)
			// Bare word repeat must not invent a match (Java has no word-repeat for ja/zh)
			for _, m := range lt.Check("test test") {
				require.NotContains(t, m.RuleID, "WORD_REPEAT")
			}
		})
	}
}

// Java Tamil.getRelevantRules IDs.
func TestRegister_Tamil_JavaRelevantOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ta")
	corepack.Register(lt, "ta")
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "COMMA_PARENTHESIS_WHITESPACE")
	require.Contains(t, ids, "DOUBLE_PUNCTUATION")
	require.Contains(t, ids, "WHITESPACE_RULE")
	require.Contains(t, ids, "TOO_LONG_SENTENCE")
	require.Contains(t, ids, "SENTENCE_WHITESPACE")
	for _, id := range ids {
		require.NotContains(t, id, "WORD_REPEAT", "Tamil Java has no word-repeat")
	}
}

func TestRegister_SupportedList(t *testing.T) {
	require.GreaterOrEqual(t, len(corepack.Supported), 30)
	codes := map[string]bool{}
	for _, s := range corepack.Supported {
		codes[s.Code] = true
	}
	require.True(t, codes["en"] && codes["zh"] && codes["be"])
}

// Java Belarusian has ParagraphRepeatBeginningRule (layout), not WordRepeatBeginning.
func TestRegister_Belarusian_NoInventWordRepeatBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	corepack.Register(lt, "be")
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "BE_WORD_REPEAT_BEGINNING_RULE")
	require.NotContains(t, ids, "BE_WORD_REPEAT_RULE")
	// Shared layout still has paragraph-level beginning rule when registered
	require.Contains(t, ids, "PARAGRAPH_REPEAT_BEGINNING_RULE")
}
