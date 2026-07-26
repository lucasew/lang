package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestINNNumberFilter_IsInvalid(t *testing.T) {
	f := NewINNNumberFilter()
	// Known valid 10-digit INN (checksum with Java weights)
	require.False(t, f.IsInvalid("7830002293"))
	// Flip last digit → invalid
	require.True(t, f.IsInvalid("7830002294"))
	// Valid 12-digit (body 5001007322 + computed kz1=5, kz2=9)
	require.False(t, f.IsInvalid("500100732259"))
	// Flip last check digit → invalid
	require.True(t, f.IsInvalid("500100732258"))
	// Flip first check digit → invalid
	require.True(t, f.IsInvalid("500100732249"))
	// wrong length → suppress
	require.False(t, f.IsInvalid("123"))
	require.False(t, f.IsInvalid(""))
	// non-digit → suppress (Java pattern non-match → null)
	require.False(t, f.IsInvalid("abc"))
	require.False(t, f.IsInvalid("7676gj8778"))
	require.False(t, f.IsInvalid("783000229a"))
}

func TestINNNumberFilter_AcceptRuleMatch(t *testing.T) {
	f := NewINNNumberFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("INN"), nil, 0, 10, "bad inn")

	// valid 10-digit → nil (suppress)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"inn": "7830002293"}, 0, nil, nil))
	// invalid 10-digit → keep match
	out := f.AcceptRuleMatch(m, map[string]string{"inn": "7830002294"}, 0, nil, nil)
	require.Equal(t, m, out)

	// valid 12-digit → nil
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"inn": "500100732259"}, 0, nil, nil))
	// invalid 12-digit → keep match
	require.Equal(t, m, f.AcceptRuleMatch(m, map[string]string{"inn": "500100732258"}, 0, nil, nil))

	// non-digit / wrong length → nil
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"inn": "7676gj8778"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"inn": "123"}, 0, nil, nil))
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"inn": ""}, 0, nil, nil))
}

func TestINNNumberFilter_GetRequiredPanics(t *testing.T) {
	f := NewINNNumberFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("INN"), nil, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
}

func TestINNNumberFilter_Registered(t *testing.T) {
	class := "org.languagetool.rules.ru.INNNumberFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	rf := patterns.GlobalRuleFilterCreator.GetFilter(class)
	require.IsType(t, &INNNumberFilter{}, rf)
}
