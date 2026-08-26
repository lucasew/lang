package patterns

// Twin of languagetool-core/src/test/java/org/languagetool/rules/patterns/MatchTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of MatchTest.testStartUpper
func TestMatch_StartUpper(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseStartUpper, false, false, IncludeNone)
	require.True(t, m.ConvertsCase())
	require.Equal(t, CaseStartUpper, m.GetCaseConversionType())
	require.Equal(t, "Hello", ConvertCase(m.GetCaseConversionType(), "hello", "X"))
}

// Port of MatchTest.testStartLower
func TestMatch_StartLower(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseStartLower, false, false, IncludeNone)
	require.Equal(t, "hELLO", ConvertCase(m.GetCaseConversionType(), "HELLO", "X"))
}

// Port of MatchTest.testAllUpper
func TestMatch_AllUpper(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseAllUpper, false, false, IncludeNone)
	require.Equal(t, "HELLO", ConvertCase(m.GetCaseConversionType(), "hello", "X"))
}

// Port of MatchTest.testAllLower
func TestMatch_AllLower(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseAllLower, false, false, IncludeNone)
	require.Equal(t, "hello", ConvertCase(m.GetCaseConversionType(), "HELLO", "X"))
}

// Port of MatchTest.testPreserveStartUpper
func TestMatch_PreserveStartUpper(t *testing.T) {
	m := NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone)
	require.Equal(t, "Hello", ConvertCase(m.GetCaseConversionType(), "hello", "World"))
	require.Equal(t, "HELLO", ConvertCase(m.GetCaseConversionType(), "hello", "WORLD"))
}

// Port of MatchTest.testStaticLemmaPreserveStartLower
func TestMatch_StaticLemmaPreserveStartLower(t *testing.T) {
	m := NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone)
	m.SetLemmaString("lemma")
	require.True(t, m.IsStaticLemma())
	require.Equal(t, "lemma", m.GetLemma())
	require.Equal(t, "hello", ConvertCase(m.GetCaseConversionType(), "hello", "world"))
}

// Port of MatchTest.testStaticLemmaPreserveStartUpper
func TestMatch_StaticLemmaPreserveStartUpper(t *testing.T) {
	m := NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone)
	m.SetLemmaString("lemma")
	require.Equal(t, "Lemma", ConvertCase(m.GetCaseConversionType(), "lemma", "World"))
}

// Port of MatchTest.testStaticLemmaPreserveAllUpper
func TestMatch_StaticLemmaPreserveAllUpper(t *testing.T) {
	m := NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone)
	m.SetLemmaString("lemma")
	require.Equal(t, "LEMMA", ConvertCase(m.GetCaseConversionType(), "lemma", "WORLD"))
}

// Port of MatchTest.testStaticLemmaPreserveMixed
func TestMatch_StaticLemmaPreserveMixed(t *testing.T) {
	m := NewMatch("", "", false, "", "", CasePreserve, false, false, IncludeNone)
	m.SetLemmaString("lemma")
	require.Equal(t, "Lemma", ConvertCase(m.GetCaseConversionType(), "lemma", "World"))
}

// Port of MatchTest.testPreserveStartLower
func TestMatch_PreserveStartLower(t *testing.T) {
	require.Equal(t, "hello", ConvertCase(CasePreserve, "hello", "world"))
}

// Port of MatchTest.testPreserveAllUpper
func TestMatch_PreserveAllUpper(t *testing.T) {
	require.Equal(t, "HELLO", ConvertCase(CasePreserve, "hello", "WORLD"))
}

// Port of MatchTest.testPreserveMixed
func TestMatch_PreserveMixed(t *testing.T) {
	require.Equal(t, "Hello", ConvertCase(CasePreserve, "hello", "World"))
}

// Port of MatchTest.testPreserveNoneUpper
func TestMatch_PreserveNoneUpper(t *testing.T) {
	require.Equal(t, "hello", ConvertCase(CaseNone, "hello", "WORLD"))
}

// Port of MatchTest.testPreserveNoneLower
func TestMatch_PreserveNoneLower(t *testing.T) {
	require.Equal(t, "HELLO", ConvertCase(CaseNone, "HELLO", "world"))
}

// Port of MatchTest.testPreserveNoneMixed
func TestMatch_PreserveNoneMixed(t *testing.T) {
	require.Equal(t, "HeLLo", ConvertCase(CaseNone, "HeLLo", "World"))
}

// Twin of MatchTest.testSimpleIncludeFollowing (IncludeRange.FOLLOWING + skipped tokens).
// Java uses SENT_START at index 0; index 1 is first word.
func TestMatch_SimpleIncludeFollowing(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeFollowing)
	require.Equal(t, IncludeFollowing, m.GetIncludeSkipped())
	// tokens: SENT_START, w1, w2, w3, w4 — createState(..., 1, 3) → following w2 + w3
	toks := matchTestTokens(
		"SENT_START", "inflectedform11", "inflectedform2", "inflectedform122", "inflectedform122",
	)
	st := m.CreateStateRange(nil, toks, 1, 3)
	require.Equal(t, []string{"inflectedform2 inflectedform122"}, st.ToFinalString(""))
}

// Twin of MatchTest.testPOSIncludeFollowing — POS is ignored when FOLLOWING.
func TestMatch_POSIncludeFollowing(t *testing.T) {
	m := NewMatch("POS2", "POS33", true, "", "", CaseNone, false, false, IncludeFollowing)
	require.Equal(t, IncludeFollowing, m.GetIncludeSkipped())
	toks := matchTestTokens(
		"SENT_START", "inflectedform11", "inflectedform2", "inflectedform122", "inflectedform122",
	)
	st := m.CreateStateRange(nil, toks, 1, 3)
	require.Equal(t, []string{"inflectedform2 inflectedform122"}, st.ToFinalString(""))
}

// Twin of MatchTest.testIncludeAll
func TestMatch_IncludeAll(t *testing.T) {
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeAll)
	require.Equal(t, IncludeAll, m.GetIncludeSkipped())
	toks := matchTestTokens(
		"SENT_START", "inflectedform11", "inflectedform2", "inflectedform122", "inflectedform122",
	)
	st := m.CreateStateRange(nil, toks, 1, 3)
	require.Equal(t, []string{"inflectedform11 inflectedform2 inflectedform122"}, st.ToFinalString(""))
}

// Port of MatchTest.testPOSIncludeAll metadata (full synth form deferred without ManualSynthesizer).
func TestMatch_POSIncludeAll(t *testing.T) {
	m := NewMatch("VB.*", "", true, "", "", CaseNone, true, false, IncludeAll)
	require.True(t, m.IsPostagRegexp())
	require.NotNil(t, m.GetPosRegexMatch())
}

func matchTestTokens(surfaces ...string) []*languagetool.AnalyzedTokenReadings {
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(surfaces))
	pos := 0
	for _, s := range surfaces {
		if s == "SENT_START" {
			ss := languagetool.SentenceStartTagName
			out = append(out, languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken("", &ss, nil), pos))
			continue
		}
		// Mark non-first content tokens as whitespaceBefore for space joining.
		atr := languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(s, nil, nil), pos)
		if len(out) > 1 {
			atr.SetWhitespaceBefore(true)
		}
		out = append(out, atr)
		pos += len(s) + 1
	}
	return out
}
