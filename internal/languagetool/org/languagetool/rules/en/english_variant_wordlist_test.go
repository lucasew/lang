package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUSGBVariantMap(t *testing.T) {
	// column 1: British key → American (American speller)
	gbToUS := LoadUSGBVariantMap(1)
	if len(gbToUS) == 0 {
		t.Skip("en/en-US-GB.txt not discoverable")
	}
	// colour is GB; color is US — column 1 keys are GB forms
	require.Equal(t, "color", gbToUS["colour"])
	// column 0: American key → British
	usToGB := LoadUSGBVariantMap(0)
	require.Equal(t, "colour", usToGB["color"])
}

func TestAmericanSpeller_ColourVariant(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	if len(r.OtherVariant) == 0 {
		t.Skip("en-US-GB not loaded")
	}
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())
}
