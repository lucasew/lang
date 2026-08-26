package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of AnalyzedTokenReadings.addReading: replaces surface when the new
// token.getToken().length() (UTF-16) is strictly greater than this.token.length().
func TestAnalyzedTokenReadings_AddReading_UTF16LongerSurface(t *testing.T) {
	// Base surface "ab" (UTF-16 len 2). Add emoji "😀" (UTF-16 len 2) — not longer.
	base := NewAnalyzedToken("ab", nil, nil)
	r := NewAnalyzedTokenReadings(base)
	require.Equal(t, "ab", r.GetToken())

	emoji := NewAnalyzedToken("😀", nil, nil)
	r.AddReading(emoji, "test")
	require.Equal(t, "ab", r.GetToken(), "emoji length 2 is not > ab length 2 (UTF-16)")

	// "a" (len 1) + add "😀" (len 2) → surface becomes emoji
	r2 := NewAnalyzedTokenReadings(NewAnalyzedToken("a", nil, nil))
	r2.AddReading(NewAnalyzedToken("😀", nil, nil), "test")
	require.Equal(t, "😀", r2.GetToken(), "UTF-16 len 2 > 1 replaces surface")

	// UTF-8 trap: "é" is 2 UTF-8 bytes / 1 UTF-16 unit; "ab" is 2/2.
	// Adding "é" must NOT replace "ab" (byte compare would wrongly think 2>2 is false anyway;
	// reverse: base "é" (utf8 2, utf16 1), add "x" (1,1) — neither longer.
	// base "é", add "xy" (utf16 2) → replaces.
	r3 := NewAnalyzedTokenReadings(NewAnalyzedToken("é", nil, nil))
	r3.AddReading(NewAnalyzedToken("xy", nil, nil), "test")
	require.Equal(t, "xy", r3.GetToken(), "UTF-16: 2 > 1 replaces (byte len equal 2==2 would not)")
}
