package srx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Ukrainian segment.srx needs RE2 adaptations that English does not:
// - Java \h / \v whitespace escapes
// - negative lookbehind/lookahead (р., initials, ст., …)
// - empty beforebreak ("— Ред.")
// - \b inside unused alternation branches (Є.Бакуліна)
func TestUkrainian_HVLookaroundAndInitials(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.LangRules["Ukrainian"]), 45,
		"most Ukrainian rules must compile (\\h/\\v + lookarounds)")

	cases := []struct {
		text string
		want []string
	}{
		{"Вони приїхали в Париж. Але там їм геть не сподобалося.",
			[]string{"Вони приїхали в Париж. ", "Але там їм геть не сподобалося."}},
		{"Є.Бакуліна", []string{"Є.Бакуліна"}},
		{"Засідав І. П. Єрмолюк.", []string{"Засідав І. П. Єрмолюк."}},
		{"В 1941 р. Конрад Цузе побудував.", []string{"В 1941 р. Конрад Цузе побудував."}},
		{"15 вересня 1995 р. Україною було підписано",
			[]string{"15 вересня 1995 р. Україною було підписано"}},
		{"Але закінчилося аж у січні 2013 р. Як бачимо",
			[]string{"Але закінчилося аж у січні 2013 р. ", "Як бачимо"}},
		{"інкримінують ч. 1 ст. 11", []string{"інкримінують ч. 1 ст. 11"}},
		{"(вони самі це визнали. - Ред.)", []string{"(вони самі це визнали. - Ред.)"}},
		{"товариш С.\u202fОхримович.", []string{"товариш С.\u202fОхримович."}},
		{"відбув у тюрмах.\u202fНещодавно письменник",
			[]string{"відбув у тюрмах.\u202f", "Нещодавно письменник"}},
		{"в м.Києві", []string{"в м.Києві"}},
		{"для др.  Харченко.", []string{"для др.  Харченко."}},
	}
	for _, tc := range cases {
		got := doc.Split(tc.text, "uk", "_two")
		require.Equal(t, tc.want, got, "text=%q", tc.text)
	}
}
