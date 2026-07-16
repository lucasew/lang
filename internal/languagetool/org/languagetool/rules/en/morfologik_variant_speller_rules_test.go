package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikVariantSpellers(t *testing.T) {
	us := NewMorfologikAmericanSpellerRule()
	require.Equal(t, MorfologikAmericanSpellerRuleID, us.GetID())
	require.Equal(t, AmericanSpellerDict, us.GetFileName())
	require.Equal(t, AmericanVariantSpellingFile, us.GetLanguageVariantSpellingFileName())

	us.OtherVariant = map[string]string{"colour": "color"}
	us.OtherVariantName = "British English"
	// American map: British form → American? Java uses BRITISH_ENGLISH map word→british form
	// For US rule, word "color" might map to British "colour" for "valid in other variant"
	us.OtherVariant = map[string]string{"color": "colour"}
	vi := us.IsValidInOtherVariant("Color")
	require.NotNil(t, vi)
	require.Equal(t, "colour", vi.GetOtherVariant())
	require.Equal(t, "British English", vi.GetVariantName())

	require.Equal(t, MorfologikBritishSpellerRuleID, NewMorfologikBritishSpellerRule().GetID())
	require.Equal(t, MorfologikCanadianSpellerRuleID, NewMorfologikCanadianSpellerRule().GetID())
	require.Equal(t, MorfologikAustralianSpellerRuleID, NewMorfologikAustralianSpellerRule().GetID())
	require.Equal(t, MorfologikNewZealandSpellerRuleID, NewMorfologikNewZealandSpellerRule().GetID())
	require.Equal(t, MorfologikSouthAfricanSpellerRuleID, NewMorfologikSouthAfricanSpellerRule().GetID())

	m := LoadOtherVariantMap([]string{"colour\tcolor", "#c", "foo=bar"}, 0)
	require.Equal(t, "color", m["colour"])
	require.Equal(t, "bar", m["foo"])
}
