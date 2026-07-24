package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_ConfigInfo(t *testing.T) {
	api := NewApiV2(nil, nil)
	_, err := api.Handle("configinfo", nil)
	require.Error(t, err)

	r, err := api.Handle("configinfo", map[string]string{"language": "en"})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"rules"`)
	require.Contains(t, r.Body, "EN_A_VS_AN")
	require.Contains(t, r.Body, `"software"`)
	require.Contains(t, r.Body, "maxTextLength")
	var raw map[string]any
	require.NoError(t, json.Unmarshal([]byte(r.Body), &raw))
	rules, ok := raw["rules"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, rules)
}

func TestApiV2_WordsCRUD(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("words/add", map[string]string{"word": "xyzzy", "username": "u1"})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"added":true`)

	r, err = api.Handle("words", map[string]string{"username": "u1", "limit": "20"})
	require.NoError(t, err)
	require.Contains(t, r.Body, "xyzzy")

	r, err = api.Handle("words/delete", map[string]string{"word": "xyzzy", "username": "u1"})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"deleted":true`)

	r, err = api.Handle("words", map[string]string{"username": "u1"})
	require.NoError(t, err)
	require.NotContains(t, r.Body, "xyzzy")
}

func TestApiV2_WordsBatch(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("words/add", map[string]string{
		"mode": "batch", "words": "alpha beta", "username": "batch",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, `"count":2`)
	r, err = api.Handle("words", map[string]string{"username": "batch", "limit": "50"})
	require.NoError(t, err)
	require.Contains(t, r.Body, "alpha")
	require.Contains(t, r.Body, "beta")
}

func TestUserDictionary(t *testing.T) {
	d := NewUserDictionary()
	require.True(t, d.Add("", "foo"))
	require.False(t, d.Add("", "foo"))
	require.Equal(t, []string{"foo"}, d.List("", 0, 10))
	require.True(t, d.Delete("", "foo"))
	require.Empty(t, d.List("", 0, 10))
}
