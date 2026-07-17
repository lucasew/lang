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
		{"sv", "hej hej", "SV_WORD_REPEAT_RULE"},
		{"da", "hej hej", "DA_WORD_REPEAT_RULE"},
		{"gl", "ola ola", "GL_WORD_REPEAT_RULE"},
		{"sk", "test test", "SK_WORD_REPEAT_RULE"},
		{"el", "γεια γεια", "EL_WORD_REPEAT_RULE"},
		{"ro", "test test", "RO_WORD_REPEAT_RULE"},
		{"pt-BR", "teste teste", "PORTUGUESE_WORD_REPEAT_RULE"},
		{"ar", "كلمة كلمة", "ARABIC_WORD_REPEAT_RULE"},
		{"sl", "test test", "SL_WORD_REPEAT_RULE"},
		{"br", "test test", "BR_WORD_REPEAT_RULE"},
		{"fa", "test test", "FA_WORD_REPEAT_RULE"},
		{"ga", "test test", "GA_WORD_REPEAT_RULE"},
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

func TestRegister_WordRepeatBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	corepack.Register(lt, "fr")
	m := lt.Check("Bonjour le monde. Bonjour la terre. Bonjour le ciel.")
	found := false
	for _, x := range m {
		if x.RuleID == "FR_WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}

func TestRegister_GenericPacks(t *testing.T) {
	for _, code := range []string{"be", "zh", "ja", "eo", "sr"} {
		t.Run(code, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(code)
			corepack.Register(lt, code)
			require.NotEmpty(t, lt.Check("a  b"))
			m := lt.Check("test test")
			require.NotEmpty(t, m)
		})
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

func TestRegister_GenericBeginning(t *testing.T) {
	lt := languagetool.NewJLanguageTool("be")
	corepack.Register(lt, "be")
	ids := lt.GetAllRegisteredRuleIDs()
	found := false
	for _, id := range ids {
		if id == "BE_WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	require.True(t, found, "%v", ids)
	// three successive same starts
	m := lt.Check("Test one. Test two. Test three.")
	foundMatch := false
	for _, x := range m {
		if x.RuleID == "BE_WORD_REPEAT_BEGINNING_RULE" {
			foundMatch = true
		}
	}
	require.True(t, foundMatch, "%+v", m)
}
