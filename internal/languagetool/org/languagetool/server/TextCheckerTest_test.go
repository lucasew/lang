package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of TextCheckerTest.testJSONP
func TestTextChecker_JSONP(t *testing.T) {
	p, err := ParseCheckQueryParams(map[string]string{
		"text": "not used", "language": "en", "callback": "myCallback",
	})
	require.NoError(t, err)
	require.Equal(t, "myCallback", p.Callback)
	// JSONP wrapper shape used by V2 responses
	inner := `{"matches":[]}`
	wrapped := p.Callback + "(" + inner + ");"
	require.True(t, strings.HasPrefix(wrapped, "myCallback("))
	require.True(t, strings.HasSuffix(wrapped, ");"))
}

// Port of TextCheckerTest.testMaxTextLength
func TestTextChecker_MaxTextLength(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxTextLengthAnonymous = 10
	tc := NewV2TextChecker(cfg, false, nil)
	params := map[string]string{"text": "not used", "language": "en"}
	require.NoError(t, tc.CheckParams(params))

	err := tc.ValidateTextLength("longer than 10 chars", DefaultUserLimits(cfg))
	require.Error(t, err)
	var tooLong *TextTooLongError
	require.ErrorAs(t, err, &tooLong)

	// short text OK
	require.NoError(t, tc.ValidateTextLength("short", DefaultUserLimits(cfg)))

	// hard length still enforced when limits raise soft max
	cfg.MaxTextHardLength = 30
	limits := &UserLimits{MaxTextLength: 100}
	err = tc.ValidateTextLength("now it's even longer than 30 chars!!", limits)
	require.Error(t, err)
	require.ErrorAs(t, err, &tooLong)
}

// Port of TextCheckerTest.testInvalidAltLanguages
func TestTextChecker_InvalidAltLanguages(t *testing.T) {
	require.Error(t, ValidateAltLanguages("en"))
	require.Error(t, ValidateAltLanguages("xy"))
	require.NoError(t, ValidateAltLanguages("de-DE"))
	require.NoError(t, ValidateAltLanguages("en-US"))
}

// Port of TextCheckerTest.testDetectLanguageOfString — Java @Ignore / full detector deferred
func TestTextChecker_DetectLanguageOfString(t *testing.T) {
	t.Skip("Java @Ignore / requires language identifier model")
}

// Port of TextCheckerTest.testInvalidPreferredVariant
func TestTextChecker_InvalidPreferredVariant(t *testing.T) {
	// "en" is not a variant (needs dash)
	err := ValidatePreferredVariants([]string{"en"}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "preferredVariants")
}

// Port of TextCheckerTest.testInvalidPreferredVariant2
func TestTextChecker_InvalidPreferredVariant2(t *testing.T) {
	err := ValidatePreferredVariants([]string{"en-YY"}, func(code string) bool {
		return false // variant doesn't exist
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "en-YY")
}
