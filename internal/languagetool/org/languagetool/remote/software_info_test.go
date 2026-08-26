package remote

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSoftwareInfo(t *testing.T) {
	cli := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "/v2/info", req.URL.Path)
		body := `{"software":{"name":"LanguageTool-Go","version":"dev","apiVersion":1}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	rlt := NewRemoteLanguageTool("http://example.invalid")
	rlt.Client = cli
	sw, err := rlt.GetSoftwareInfo()
	require.NoError(t, err)
	require.Equal(t, "LanguageTool-Go", sw["name"])
	require.Equal(t, "dev", sw["version"])
}
