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

// Port of TextCheckerTest.testDetectLanguageOfString — inject heuristic (full FastText deferred).
// Expectations match Java TextCheckerTest (default variants, preferred short-code equals).
func TestTextChecker_DetectLanguageOfString(t *testing.T) {
	// empty text + preferred → fallback "en" then preferred en-GB
	require.Equal(t, "en-GB", DetectLanguageOfString("", []string{"en-GB"}, func(string) string { return "" }))

	// detect en + preferred en-GB → promote variant
	require.Equal(t, "en-GB", DetectLanguageOfString("X", []string{"en-GB"}, func(string) string { return "en" }))

	// English sample + preferred en-ZA among list
	english := "This is a longer English sample for detection."
	require.Equal(t, "en-ZA", DetectLanguageOfString(english, []string{"de-AT", "en-ZA"}, func(string) string { return "en" }))

	// English + empty preferred → default variant en-US (Java AmericanEnglish)
	require.Equal(t, "en-US", DetectLanguageOfString(english, nil, func(string) string { return "en" }))
	require.Equal(t, "en-US", DetectLanguageOfString(english, []string{}, func(string) string { return "en" }))

	// German sample + de-AT preferred
	german := "Das ist ein deutscher Text mit Größe."
	require.Equal(t, "de-AT", DetectLanguageOfString(german, []string{"de-AT", "en-ZA"}, nil))

	// Java: de-at lowercase region → parseLanguage canonical de-AT
	require.Equal(t, "de-AT", DetectLanguageOfString(german, []string{"de-at", "en-ZA"}, nil))

	// no preferred: default language variant (German → de-DE, Ukrainian base → uk)
	require.Equal(t, "de-DE", DetectLanguageOfString(german, nil, nil))
	require.Equal(t, "uk", DetectLanguageOfString("Це українська мова з ї.", nil, nil))

	// preferred non-empty that does not match detected short code: keep fallback/detect, no default variant
	// Java: detected null → en; preferred only de-AT → stays en (not en-US)
	code, err := DetectLanguageOfStringWithFallback("", "", []string{"de-AT"}, func(string) string { return "" }, nil)
	require.NoError(t, err)
	require.Equal(t, "en", code)

	// case-sensitive preferred base: "EN-GB" does not match short "en" (Java String.equals)
	require.Equal(t, "en", DetectLanguageOfString("X", []string{"EN-GB"}, func(string) string { return "en" }))
}

// Port of TextCheckerTest.testInvalidPreferredVariant
func TestTextChecker_InvalidPreferredVariant(t *testing.T) {
	// "en" is not a variant (needs dash) — thrown from detectLanguageOfString itself
	_, err := DetectLanguageOfStringErr("This is English.", []string{"en"}, func(string) string { return "en" })
	require.Error(t, err)
	require.Contains(t, err.Error(), "preferredVariants")

	err = ValidatePreferredVariants([]string{"en"}, nil)
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

	_, err = DetectLanguageOfStringWithFallback("English text.", "", []string{"en-YY"},
		func(string) string { return "en" },
		func(string) bool { return false })
	require.Error(t, err)
	require.Contains(t, err.Error(), "en-YY")
}

func TestParsePreferredVariants_RequiresAuto(t *testing.T) {
	_, err := ParsePreferredVariants(map[string]string{
		"language":          "en",
		"preferredVariants": "en-GB",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "language=auto")

	// COMMA_WHITESPACE like altLanguages
	p, err := ParsePreferredVariants(map[string]string{
		"language":          "auto",
		"preferredVariants": "en-GB, de-AT",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"en-GB", "de-AT"}, p)

	// multilingual allows preferred without auto
	p, err = ParsePreferredVariants(map[string]string{
		"language":          "en",
		"multilingual":      "true",
		"preferredVariants": "en-GB",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"en-GB"}, p)
}

func TestValidateNoopLanguages(t *testing.T) {
	require.Error(t, ValidateNoopLanguages(map[string]string{
		"language":      "en",
		"noopLanguages": "cs,sk",
	}))
	require.NoError(t, ValidateNoopLanguages(map[string]string{
		"language":      "auto",
		"noopLanguages": "cs,sk",
	}))
}

func TestV2TextChecker_CheckParamsRenamed(t *testing.T) {
	tc := NewV2TextChecker(nil, false, nil)
	require.Error(t, tc.CheckParams(map[string]string{
		"language": "en", "text": "x", "preferredvariants": "en-GB",
	}))
	require.Error(t, tc.CheckParams(map[string]string{
		"language": "en", "text": "x", "autodetect": "true",
	}))
	require.Error(t, tc.CheckParams(map[string]string{
		"language": "en", "text": "x", "enabled": "FOO",
	}))
	require.NoError(t, tc.CheckParams(map[string]string{
		"language": "en", "text": "x",
	}))
}

// Ports TextChecker QueryParams construction + ServerTools.getMode/getLevel/toneTags.
func TestParseCheckQueryParams_JavaQueryParams(t *testing.T) {
	// useQuerySettings includes enableTempOffRules, not useEnabledOnly alone
	p, err := ParseCheckQueryParams(map[string]string{"enableTempOffRules": "true"})
	require.NoError(t, err)
	require.True(t, p.EnableTempOffRules)
	require.True(t, p.UseQuerySettings)
	require.True(t, p.RegressionTestMode)
	require.True(t, p.InputLogging)

	// enabledOnly=yes (Java accepts "yes" and "true" only, case-sensitive)
	p, err = ParseCheckQueryParams(map[string]string{
		"enabledOnly":  "yes",
		"enabledRules": "RULE_A",
	})
	require.NoError(t, err)
	require.True(t, p.UseEnabledOnly)

	// EqualFold invent: TRUE / YES / True must not enable
	p, err = ParseCheckQueryParams(map[string]string{
		"enabledOnly":  "TRUE",
		"enabledRules": "RULE_A",
	})
	require.NoError(t, err)
	require.False(t, p.UseEnabledOnly)

	// enabledOnly without rules/categories
	_, err = ParseCheckQueryParams(map[string]string{"enabledOnly": "true"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "enabled rules or categories")

	// mode/level case-sensitive API values
	p, err = ParseCheckQueryParams(map[string]string{"mode": "textLevelOnly", "level": "academic"})
	require.NoError(t, err)
	require.Equal(t, CheckModeTextLevelOnly, p.Mode)
	require.Equal(t, CheckLevelAcademic, p.Level)

	_, err = ParseCheckQueryParams(map[string]string{"mode": "ALL"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Mode must be")

	_, err = ParseCheckQueryParams(map[string]string{"level": "PICKY"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "level")

	// inputLogging=no (exact)
	p, err = ParseCheckQueryParams(map[string]string{"inputLogging": "no"})
	require.NoError(t, err)
	require.False(t, p.InputLogging)
	p, err = ParseCheckQueryParams(map[string]string{"inputLogging": "NO"})
	require.NoError(t, err)
	require.True(t, p.InputLogging) // Java equals("no") only

	// toneTags default / empty / known
	p, err = ParseCheckQueryParams(map[string]string{})
	require.NoError(t, err)
	require.Equal(t, []string{"ALL_WITHOUT_GOAL_SPECIFIC"}, p.ToneTags)

	p, err = ParseCheckQueryParams(map[string]string{"toneTags": ""})
	require.NoError(t, err)
	require.Equal(t, []string{"ALL_WITHOUT_GOAL_SPECIFIC"}, p.ToneTags)

	p, err = ParseCheckQueryParams(map[string]string{"toneTags": "clarity,formal"})
	require.NoError(t, err)
	require.Equal(t, []string{"clarity", "formal"}, p.ToneTags)

	// multi: meta tags ignored when length>1
	p, err = ParseCheckQueryParams(map[string]string{"toneTags": "clarity,NO_TONE_RULE"})
	require.NoError(t, err)
	require.Equal(t, []string{"clarity"}, p.ToneTags)
}

func TestServerTools_GetModeGetLevel(t *testing.T) {
	m, err := GetMode(nil)
	require.NoError(t, err)
	require.Equal(t, CheckModeAll, m)
	m, err = GetMode(map[string]string{"mode": "allButTextLevelOnly"})
	require.NoError(t, err)
	require.Equal(t, CheckModeAllButTextLevelOnly, m)
	require.Equal(t, "!tlo", GetModeForLog(m))
	m, err = GetMode(map[string]string{"mode": "batch"})
	require.NoError(t, err)
	require.Equal(t, CheckModeAll, m)

	lv, err := GetLevel(map[string]string{"level": "elegant"})
	require.NoError(t, err)
	require.Equal(t, CheckLevelElegant, lv)
}

func TestGetLanguageAutoDetect_CaseSensitive(t *testing.T) {
	tc := NewV2TextChecker(nil, false, nil)
	require.True(t, tc.GetLanguageAutoDetect(map[string]string{"language": "auto"}))
	require.False(t, tc.GetLanguageAutoDetect(map[string]string{"language": "AUTO"}))
	require.False(t, tc.GetLanguageAutoDetect(map[string]string{"language": "Auto"}))
}
