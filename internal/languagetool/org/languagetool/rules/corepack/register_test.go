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
