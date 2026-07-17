package remote

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestGetLanguages(t *testing.T) {
	body := `[{"name":"English (US)","code":"en","longCode":"en-US"},{"name":"German","code":"de","longCode":"de-DE"}]`
	cli := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "/v2/languages", req.URL.Path)
		require.Equal(t, http.MethodGet, req.Method)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})
	rlt := NewRemoteLanguageTool("http://example.invalid")
	rlt.Client = cli
	langs, err := rlt.GetLanguages()
	require.NoError(t, err)
	require.Len(t, langs, 2)
	require.Equal(t, "en-US", langs[0].LongCode)
	require.Equal(t, "en", langs[0].Code)
}
