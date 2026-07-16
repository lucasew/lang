package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanNumberSpellerFilter(t *testing.T) {
	f := NewCatalanNumberSpellerFilter(func(s string) string {
		if s == "feminine 2" {
			return "dues"
		}
		if s == "21" {
			return "vint-i-un"
		}
		if s == "1234" {
			return "mil dos-cents trenta-quatre extras"
		}
		return "u"
	})
	require.Equal(t, "dues", f.Suggest("2", "feminine", false))
	require.Equal(t, "Vint-i-un", f.Suggest("21", "", true))
	require.Equal(t, "", f.Suggest("1234", "", false)) // too many words
}
