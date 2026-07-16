package remote

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckConfigurationBuilder(t *testing.T) {
	cfg := NewCheckConfigurationBuilder("en-US").
		SetMotherTongueLangCode("de").
		EnabledRuleIDs("A", "B").
		DisabledRuleIDs("C").
		Mode("all").
		Level("default").
		Build()
	require.Equal(t, "en-US", cfg.LangCode)
	require.Equal(t, "de", cfg.MotherTongueLangCode)
	require.Equal(t, []string{"A", "B"}, cfg.EnabledRuleIDs)
	require.False(t, cfg.GuessLanguage)

	auto := NewAutoDetectCheckConfigurationBuilder().Build()
	require.True(t, auto.GuessLanguage)

	require.Panics(t, func() {
		NewCheckConfigurationBuilder("en").EnabledOnly().Build()
	})
}

func TestGetURLParams(t *testing.T) {
	cfg := NewCheckConfigurationBuilder("en").EnabledRuleIDs("X").Build()
	p := GetURLParams("hi", cfg, map[string]string{"foo": "bar"})
	require.Equal(t, "hi", p.Get("text"))
	require.Equal(t, "en", p.Get("language"))
	require.Equal(t, "X", p.Get("enabledRules"))
	require.Equal(t, "bar", p.Get("foo"))
}

func TestParseCheckJSON(t *testing.T) {
	raw := `{
	  "software": {"name":"LT","version":"6.0"},
	  "language": {"name":"English","code":"en-US"},
	  "matches": [{
	    "message":"spelling",
	    "offset":0,
	    "length":4,
	    "context":{"text":"helo","offset":0,"length":4},
	    "replacements":[{"value":"hello"}],
	    "rule":{"id":"MORFOLOGIK","description":"spell","category":{"id":"TYPOS","name":"Typos"}}
	  }]
	}`
	res, err := ParseCheckJSON([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, "en-US", res.GetLanguageCode())
	require.Len(t, res.GetMatches(), 1)
	require.Equal(t, "MORFOLOGIK", res.GetMatches()[0].GetRuleID())
	require.Equal(t, []string{"hello"}, res.GetMatches()[0].GetReplacements())
	require.Equal(t, "LT", res.GetRemoteServer().GetSoftware())
}

type stubClient struct {
	body string
	code int
}

func (s stubClient) Do(req *http.Request) (*http.Response, error) {
	code := s.code
	if code == 0 {
		code = 200
	}
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestRemoteLanguageToolCheck(t *testing.T) {
	require.Panics(t, func() { NewRemoteLanguageTool("http://x/") })
	body := `{"software":{"name":"LT","version":"1"},"language":{"name":"English","code":"en"},"matches":[]}`
	lt := NewRemoteLanguageTool("http://example.invalid")
	lt.Client = stubClient{body: body}
	res, err := lt.Check("hello", "en")
	require.NoError(t, err)
	require.Empty(t, res.GetMatches())
}
