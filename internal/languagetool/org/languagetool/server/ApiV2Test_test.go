package server

// Twin of ApiV2Test
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_Languages(t *testing.T) {
	api := NewApiV2(nil, []LanguageInfo{{Name: "English", Code: "en"}})
	r, err := api.Handle("languages", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "English")
}

func TestApiV2_InvalidRequest(t *testing.T) {
	api := NewApiV2(nil, nil)
	_, err := api.Handle("unknown-path", nil)
	require.Error(t, err)
}

func TestApiV2_InvalidJsonRequest(t *testing.T) {
	api := NewApiV2(nil, nil)
	// check without language
	_, err := api.Handle("check", map[string]string{"text": "hi"})
	require.Error(t, err)
}

func TestApiV2_MissingLanguageParameter(t *testing.T) {
	api := NewApiV2(nil, nil)
	_, err := api.Handle("check", map[string]string{"text": "Hello world"})
	require.Error(t, err)
}

func TestApiV2_CheckEngine(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "This is an test.",
	})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "EN_A_VS_AN")
	require.Contains(t, r.Body, `"matches"`)

	r2, err := api.Handle("check", map[string]string{
		"language": "fr",
		"text":     "bonjour bonjour",
	})
	require.NoError(t, err)
	require.Contains(t, r2.Body, "FR_WORD_REPEAT_RULE")
}
