package language

// Twin of DefaultLanguageIdentifierTest — exercises identifier package.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

func TestDefaultLanguageIdentifier_Detection(t *testing.T) {
	id := identifier.NewDefaultLanguageIdentifier(1000)
	id.ProfileScore = func(clean string, preferred []string) map[string]float64 {
		return map[string]float64{"en": 0.95, "de": 0.1}
	}
	got := id.Detect("This is clearly an English sentence about language detection.", nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.GetDetectedLanguageCode())
}

func TestDefaultLanguageIdentifier_KnownLimitations(t *testing.T) {
	id := identifier.NewDefaultLanguageIdentifier(1000)
	require.Nil(t, id.Detect("   ", nil, nil))
	require.Nil(t, id.Detect("", nil, nil))
}

func TestDefaultLanguageIdentifier_IgnoreSignature(t *testing.T) {
	id := identifier.NewDefaultLanguageIdentifier(1000)
	id.ProfileScore = func(clean string, preferred []string) map[string]float64 {
		// ignore short signature-like texts
		if len([]rune(clean)) < 20 {
			return nil
		}
		return map[string]float64{"en": 0.9}
	}
	require.Nil(t, id.Detect("Best,", nil, nil))
	got := id.Detect("Best regards, this email body is long enough for detection tests.", nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.GetDetectedLanguageCode())
}

func TestDefaultLanguageIdentifier_AdditionalLanguagesBuiltIn(t *testing.T) {
	id := identifier.NewDefaultLanguageIdentifier(1000)
	require.Contains(t, id.IgnoreLangCodes, "ast")
	require.Contains(t, id.IgnoreLangCodes, "gl")
	require.False(t, id.IsFastTextEnabled())
	id.EnableFastText(func(s string) map[string]float64 { return map[string]float64{"fr": 0.99} })
	require.True(t, id.IsFastTextEnabled())
}
