package remote

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMetrics(t *testing.T) {
	cli := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "/v2/metrics", req.URL.Path)
		body := `{"checks":3,"matches":1,"characters":10,"httpRequests":5}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	rlt := NewRemoteLanguageTool("http://example.invalid")
	rlt.Client = cli
	m, err := rlt.GetMetrics()
	require.NoError(t, err)
	require.Equal(t, float64(3), m["checks"])
	require.Equal(t, float64(5), m["httpRequests"])
}
