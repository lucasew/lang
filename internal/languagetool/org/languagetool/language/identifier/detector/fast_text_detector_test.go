package detector

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func registerDetectLangs(t *testing.T) {
	t.Helper()
	// Java Languages registry is always populated; tests need en/de/fr for canLanguageBeDetected.
	for _, m := range []languagetool.LanguageMeta{
		{Name: "English", Code: "en"},
		{Name: "German", Code: "de"},
		{Name: "French", Code: "fr"},
	} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(m.Code) {
			languagetool.GlobalLanguages.Register(m)
		}
	}
}

func TestFastTextParseBuffer(t *testing.T) {
	registerDetectLangs(t)
	d := NewFastTextDetectorForTest()
	// Java: canLanguageBeDetected — supported langs kept
	m, err := d.ParseBuffer("__label__en 0.9 __label__de 0.1", nil)
	require.NoError(t, err)
	require.InDelta(t, 0.9, m["en"], 1e-9)
	require.InDelta(t, 0.1, m["de"], 1e-9)

	// unsupported code dropped unless in additional
	m, err = d.ParseBuffer("__label__en 0.5 __label__zz 0.4", nil)
	require.NoError(t, err)
	require.Contains(t, m, "en")
	require.NotContains(t, m, "zz")
	m, err = d.ParseBuffer("__label__en 0.5 __label__zz 0.4", []string{"zz"})
	require.NoError(t, err)
	require.Contains(t, m, "zz")

	// invent removed: non-empty additional must NOT drop supported langs
	m, err = d.ParseBuffer("__label__en 0.9 __label__de 0.1", []string{"fr"})
	require.NoError(t, err)
	require.Contains(t, m, "en")
	require.Contains(t, m, "de")

	_, err = d.ParseBuffer("nope", nil)
	require.Error(t, err)

	d.Runner = func(line string) (string, error) {
		return "__label__fr 0.8 __label__en 0.2\n", nil
	}
	m, err = d.RunFasttext("bonjour", nil)
	require.NoError(t, err)
	require.InDelta(t, 0.8, m["fr"], 1e-9)
}

func TestFastTextCanDetect(t *testing.T) {
	d := NewFastTextDetectorForTest()
	d.CanDetect = func(code string, _ []string) bool { return code == "en" }
	m, err := d.ParseBuffer("__label__en 0.9 __label__de 0.1", nil)
	require.NoError(t, err)
	require.Contains(t, m, "en")
	require.NotContains(t, m, "de")
}

func TestFastTextException(t *testing.T) {
	e := NewFastTextException("bad", true)
	require.Equal(t, "bad", e.Error())
	require.True(t, e.IsDisabled())
}

// Twin: WHITESPACE = Pattern.compile("\\s+") + buffer.trim(); startsWith on raw buffer.
func TestFastTextParseBuffer_JavaWhitespaceAndTrim(t *testing.T) {
	d := NewFastTextDetectorForTest()
	// tab between pairs
	m, err := d.ParseBuffer("__label__en\t0.9\t__label__de\t0.1", nil)
	require.NoError(t, err)
	require.InDelta(t, 0.9, m["en"], 1e-9)
	// leading spaces: Java startsWith("__label__") fails on untrimmed buffer
	_, err = d.ParseBuffer("  __label__en 0.9", nil)
	require.Error(t, err)
	// NBSP is not \s without UNICODE_CHARACTER_CLASS → one field with NBSP inside
	// "__label__en\u00a00.9" after trim still has NBSP → not even pairs of space-split tokens
	// (single token if only NBSP between) → odd length error or parse fail
	_, err = d.ParseBuffer("__label__en\u00a00.9", nil)
	require.Error(t, err)
}

func TestJavaFastTextWhitespaceSplit(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, javaFastTextWhitespaceSplit("a  b"))
	require.Equal(t, []string{""}, javaFastTextWhitespaceSplit(""))
	// NBSP not delimiter
	require.Equal(t, []string{"a\u00a0b"}, javaFastTextWhitespaceSplit("a\u00a0b"))
}

// Twin: toLowerCase(Locale.ROOT) — never Turkish dotted/dotless I mapping.
func TestJavaLocaleRootToLower(t *testing.T) {
	require.Equal(t, "i", javaLocaleRootToLower("I"))
	require.Equal(t, "café", javaLocaleRootToLower("CAFÉ"))
	// dotted capital I (U+0130) lowercases to i + combining dot under Unicode default
	require.Equal(t, strings.ToLower("İ"), javaLocaleRootToLower("İ"))
}
