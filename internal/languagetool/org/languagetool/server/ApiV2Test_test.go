package server

// Twin of ApiV2Test
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_Languages(t *testing.T) {
	api := NewApiV2(nil, []LanguageInfo{{Name: "English", Code: "en"}})
	r, err := api.Handle("languages", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "English")

	// nil languages → default corepack list
	api2 := NewApiV2(nil, nil)
	r2, err := api2.Handle("languages", nil)
	require.NoError(t, err)
	require.Contains(t, r2.Body, "German")
	require.Contains(t, r2.Body, `"code":"de"`)
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

func TestApiV2_CheckJSONP(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "ok",
		"callback": "myCb",
	})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.True(t, strings.HasPrefix(r.Body, "myCb("))
	require.True(t, strings.HasSuffix(r.Body, ");"))
	require.Contains(t, r.ContentType, "javascript")
}

func TestApiV2_CheckJSONP_InvalidCallback(t *testing.T) {
	api := NewApiV2(nil, nil)
	_, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "ok",
		"callback": "bad-1",
	})
	require.Error(t, err)
}

func TestApiV2_AutoPreferredVariants(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language":           "auto",
		"text":               "This is an English sample for detection.",
		"preferredVariants":  "en-GB",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"code":"en-GB"`)
}

func TestApiV2_CheckDataAnnotation(t *testing.T) {
	api := NewApiV2(nil, nil)
	data := `{"annotation":[{"text":"See "},{"markup":"<b>"},{"text":"a error"},{"markup":"</b>"},{"text":" here."}]}`
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"data":     data,
	})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "EN_A_VS_AN")
}
