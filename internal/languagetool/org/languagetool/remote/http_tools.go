package remote

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPTools ports lightweight HTTP helpers used by remote LT clients.
type HTTPTools struct {
	Client *http.Client
}

func NewHTTPTools(timeout time.Duration) *HTTPTools {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPTools{Client: &http.Client{Timeout: timeout}}
}

// GetString performs GET and returns body as string.
func (h *HTTPTools) GetString(url string) (string, error) {
	if h == nil || h.Client == nil {
		h = NewHTTPTools(0)
	}
	resp, err := h.Client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// PostForm posts application/x-www-form-urlencoded data.
func (h *HTTPTools) PostForm(url string, form map[string]string) (string, int, error) {
	if h == nil || h.Client == nil {
		h = NewHTTPTools(0)
	}
	vals := make([]string, 0, len(form))
	for k, v := range form {
		vals = append(vals, k+"="+v) // callers should url.QueryEscape if needed
	}
	body := strings.Join(vals, "&")
	resp, err := h.Client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(b), resp.StatusCode, nil
}

// JoinURL joins base and path without double slashes.
func JoinURL(base, path string) string {
	base = strings.TrimRight(base, "/")
	if path == "" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}
