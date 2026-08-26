package remote

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordsClientCRUD(t *testing.T) {
	cli := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/v2/words"):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"words":["foo"]}`)), Header: make(http.Header)}, nil
		case strings.HasSuffix(req.URL.Path, "/v2/words/add"):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"added":true}`)), Header: make(http.Header)}, nil
		case strings.HasSuffix(req.URL.Path, "/v2/words/delete"):
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"deleted":true}`)), Header: make(http.Header)}, nil
		default:
			t.Fatalf("unexpected %s %s", req.Method, req.URL.Path)
			return nil, nil
		}
	})
	rlt := NewRemoteLanguageTool("http://example.invalid")
	rlt.Client = cli
	ok, err := rlt.AddWord("u", "foo")
	require.NoError(t, err)
	require.True(t, ok)
	words, err := rlt.GetWords("u")
	require.NoError(t, err)
	require.Equal(t, []string{"foo"}, words)
	ok, err = rlt.DeleteWord("u", "foo")
	require.NoError(t, err)
	require.True(t, ok)
}
