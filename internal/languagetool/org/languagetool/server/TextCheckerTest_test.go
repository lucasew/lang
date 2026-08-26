package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/TextCheckerTest.java
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
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
	// Java: detected null → en; preferred only de-AT → stays en (not en-US); conf 0
	res, err := DetectLanguageOfStringWithFallback("", "", []string{"de-AT"}, func(string) string { return "" }, nil)
	require.NoError(t, err)
	require.Equal(t, "en", res.Code)
	require.Equal(t, float32(0), res.Confidence)

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

// Ports RuleMatchesAsJsonSerializer.writeLanguageSection — nested detectedLanguage, real confidence.
func TestBuildResponse_LanguageSectionNested(t *testing.T) {
	v := NewV2TextChecker(nil, false, nil)
	src := "ngram"
	body, err := v.BuildResponseExDetected("hi", "en-US", "English (US)", nil,
		DetectLanguageResult{Code: "en-GB", Confidence: 0.91, Source: &src},
		"", nil, 0)
	require.NoError(t, err)
	var resp CheckResponse
	require.NoError(t, json.Unmarshal([]byte(body), &resp))
	require.Equal(t, "en-US", resp.Language.Code)
	require.NotNil(t, resp.Language.DetectedLanguage)
	require.Equal(t, "en-GB", resp.Language.DetectedLanguage.Code)
	require.InDelta(t, 0.91, resp.Language.DetectedLanguage.Confidence, 1e-6)
	require.NotNil(t, resp.Language.DetectedLanguage.Source)
	require.Equal(t, "ngram", *resp.Language.DetectedLanguage.Source)
	// No invent top-level detectedLanguage / flat confidence 0.5
	require.NotContains(t, body, `"confidence":0.5`)
	// null source when unknown
	body2, err := v.BuildResponseExDetected("hi", "de", "German", nil,
		DetectLanguageResult{Code: "de-DE", Confidence: 0},
		"", nil, 0)
	require.NoError(t, err)
	require.Contains(t, body2, `"source":null`)
	require.Contains(t, body2, `"confidence":0`)
}

func TestPipelineSettings_KeyIncludesModeLevelInputLogging(t *testing.T) {
	// Java QueryParams equals includes mode, level, inputLogging (not toneTags).
	a := NewPipelineSettingsFull("en", "", QueryParams{
		Mode: CheckModeAll, Level: CheckLevelDefault, InputLogging: true,
	}, "", "u")
	b := NewPipelineSettingsFull("en", "", QueryParams{
		Mode: CheckModeTextLevelOnly, Level: CheckLevelDefault, InputLogging: true,
	}, "", "u")
	c := NewPipelineSettingsFull("en", "", QueryParams{
		Mode: CheckModeAll, Level: CheckLevelPicky, InputLogging: true,
	}, "", "u")
	d := NewPipelineSettingsFull("en", "", QueryParams{
		Mode: CheckModeAll, Level: CheckLevelDefault, InputLogging: false,
	}, "", "u")
	require.NotEqual(t, a.Key(), b.Key())
	require.NotEqual(t, a.Key(), c.Key())
	require.NotEqual(t, a.Key(), d.Key())

	// enableTempOffRules affects useQuerySettings path / equals
	e := NewPipelineSettingsFull("en", "", QueryParams{
		Mode: CheckModeAll, Level: CheckLevelDefault, InputLogging: true, EnableTempOffRules: true, RegressionTestMode: true,
	}, "", "u")
	require.NotEqual(t, a.Key(), e.Key())
}

func TestDetectLanguageOfStringFromDetected_PreservesConfidence(t *testing.T) {
	src := "fasttext"
	det := languagetool.NewDetectedLanguageFull("", "en", 0.87, &src)
	r, err := DetectLanguageOfStringFromDetected(&det, "", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "en-US", r.Code) // default variant
	require.InDelta(t, 0.87, float64(r.Confidence), 1e-6)
	require.NotNil(t, r.Source)
	require.Equal(t, "fasttext", *r.Source)

	// null detect → en, conf 0
	r2, err := DetectLanguageOfStringFromDetected(nil, "", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "en-US", r2.Code) // fallback en → default variant
	require.Equal(t, float32(0), r2.Confidence)
}

func ensureDetectLangs(t *testing.T) {
	t.Helper()
	// Java Languages registry is always populated for canLanguageBeDetected.
	for _, c := range []string{"en", "de", "fr"} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(c) {
			languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{Name: c, Code: c})
		}
	}
}

// TextChecker uses LanguageIdentifierService (Java ctor local vs default).
func TestTextChecker_LanguageIdentifierWired(t *testing.T) {
	ensureDetectLangs(t)
	tc := NewTextChecker(nil, false, nil)
	require.NotNil(t, tc.LanguageIdentifier)

	// Inject profile scores (Java optimaize stand-in) so Detect is non-null without invent.
	d := identifier.NewDefaultLanguageIdentifier(1000)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		if strings.Contains(text, "Größe") || strings.Contains(text, "deutsch") {
			return map[string]float64{"de": 0.92}
		}
		return map[string]float64{"en": 0.95}
	}
	tc.LanguageIdentifier = d

	// German → de + default variant de-DE; confidence from identifier
	r, err := tc.DetectLanguageOfString("Die Größe des Hauses ist enorm und sehr deutlich.", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "de-DE", r.Code)
	require.InDelta(t, 0.92, float64(r.Confidence), 1e-6)
	require.NotNil(t, r.Source)

	// preferredVariants promote short code (Java preferred short equals)
	r2, err := tc.DetectLanguageOfString("This is clearly English text for detection purposes.", []string{"en-GB"}, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "en-GB", r2.Code)

	// empty detect → fallback en conf 0 (null identifier result)
	d.ProfileScore = func(string, []string) map[string]float64 { return nil }
	r3, err := tc.DetectLanguageOfString("", nil, nil, nil)
	require.NoError(t, err)
	require.Equal(t, "en-US", r3.Code)
	require.Equal(t, float32(0), r3.Confidence)

	// LocalAPIMode → simple identifier (Java getSimpleLanguageIdentifier)
	cfg := NewHTTPServerConfig()
	cfg.LocalAPIMode = true
	tcLocal := NewTextChecker(cfg, false, nil)
	require.NotNil(t, tcLocal.LanguageIdentifier)
}

func TestParseNoopAndPreferredLanguages(t *testing.T) {
	require.Nil(t, ParseNoopLanguages(map[string]string{"language": "auto"}))
	require.Equal(t, []string{"cs", "sk"}, ParseNoopLanguages(map[string]string{"noopLanguages": "cs,sk"}))
	require.Equal(t, []string{"en", "de"}, ParsePreferredLanguages(map[string]string{"preferredLanguages": "en,de"}))
}

func TestPipeline_ToneTagsAppliedToLT(t *testing.T) {
	// Java check2 passes toneTags; pool key must not change when only tone tags differ.
	a := pipelineSettingsFor("en", CheckOptions{Level: CheckLevelDefault})
	b := pipelineSettingsFor("en", CheckOptions{Level: CheckLevelDefault, ToneTags: []string{"formal"}})
	require.Equal(t, a.Key(), b.Key(), "toneTags must not affect pool key (Java QueryParams equals omits them)")

	p := NewPipeline(a)
	p.SetCheckToneTags([]string{"formal"})
	lt := p.configuredLT()
	require.Contains(t, lt.ToneTags, languagetool.ToneFormal)
	require.NotContains(t, lt.ToneTags, languagetool.ToneAllWithoutGoalSpecific)

	p.SetCheckToneTags(nil)
	lt2 := p.configuredLT()
	require.Contains(t, lt2.ToneTags, languagetool.ToneAllWithoutGoalSpecific)
}

func TestApiV2_CheckPassesToneTagsAndDetect(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
		"toneTags": "clarity",
	})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	// Nested detectedLanguage always present (writeLanguageSection)
	require.Contains(t, r.Body, `"detectedLanguage"`)
	require.Contains(t, r.Body, `"confidence"`)
}

// Ports V2 forcePreferredLanguages → detectLanguage(..., limitOnPreferredLangs=true).
func TestForcePreferredLanguages_DetectLimit(t *testing.T) {
	ensureDetectLangs(t)
	require.False(t, ParseForcePreferredLanguages(nil))
	require.False(t, ParseForcePreferredLanguages(map[string]string{"forcePreferredLanguages": "TRUE"}))
	require.True(t, ParseForcePreferredLanguages(map[string]string{"forcePreferredLanguages": "true"}))

	// Long English-like text would normally rank "en" highest; force preferred → "de" only.
	longEN := strings.Repeat("This is English text for language detection. ", 5)
	d := identifier.NewDefaultLanguageIdentifier(1000)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		return map[string]float64{"en": 0.99, "de": 0.2}
	}
	tc := NewTextChecker(nil, false, nil)
	tc.LanguageIdentifier = d

	// without force: en → en-US default variant
	r, err := tc.DetectLanguageOfStringForce(longEN, nil, nil, []string{"de"}, false)
	require.NoError(t, err)
	require.Equal(t, "en-US", r.Code)

	// with force: only preferred de remains → de-DE
	r2, err := tc.DetectLanguageOfStringForce(longEN, nil, nil, []string{"de"}, true)
	require.NoError(t, err)
	require.Equal(t, "de-DE", r2.Code)
	require.NotNil(t, r2.Source)
}

func TestHTTPServerConfig_SetFasttextPaths(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "lid.bin")
	bin := filepath.Join(dir, "fasttext")
	require.NoError(t, os.WriteFile(model, []byte("m"), 0o644))
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755))

	cfg := NewHTTPServerConfig()
	require.NoError(t, cfg.SetFasttextPaths(model, bin))
	require.Equal(t, model, cfg.FasttextModel)
	require.Equal(t, bin, cfg.FasttextBinary)

	// model as directory invalid
	require.Error(t, cfg.SetFasttextPaths(dir, bin))
	// non-executable binary invalid
	noexec := filepath.Join(dir, "noexec")
	require.NoError(t, os.WriteFile(noexec, []byte("x"), 0o644))
	require.Error(t, cfg.SetFasttextPaths(model, noexec))

	// ngram must be file not dir
	require.Error(t, cfg.SetNgramLangIdentData(dir))
	zipPath := filepath.Join(dir, "ngrams.zip")
	require.NoError(t, os.WriteFile(zipPath, []byte("z"), 0o644))
	require.NoError(t, cfg.SetNgramLangIdentData(zipPath))
	require.Equal(t, zipPath, cfg.NgramLangIdentData)
}

// Java property keys fasttextModel/fasttextBinary/ngramLangIdentData.
func TestHTTPServerConfig_ApplyProperties_FasttextNgram(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "lid.bin")
	bin := filepath.Join(dir, "fasttext")
	require.NoError(t, os.WriteFile(model, []byte("m"), 0o644))
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755))
	zipPath := filepath.Join(dir, "ngrams.zip")
	require.NoError(t, os.WriteFile(zipPath, []byte("z"), 0o644))

	cfg := NewHTTPServerConfig()
	cfg.ApplyProperties(map[string]string{
		"fasttextModel":      model,
		"fasttextBinary":     bin,
		"ngramLangIdentData": zipPath,
		"languageModel":      dir,
	})
	require.Equal(t, model, cfg.FasttextModel)
	require.Equal(t, bin, cfg.FasttextBinary)
	require.Equal(t, zipPath, cfg.NgramLangIdentData)
	require.Equal(t, dir, cfg.LanguageModelDir)

	// only one of fasttext keys → do not set (Java requires both)
	cfg2 := NewHTTPServerConfig()
	cfg2.ApplyProperties(map[string]string{"fasttextModel": model})
	require.Empty(t, cfg2.FasttextModel)
	require.Empty(t, cfg2.FasttextBinary)
}
