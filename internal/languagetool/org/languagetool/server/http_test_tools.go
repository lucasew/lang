package server

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// HTTPTestTools ports org.languagetool.server.HTTPTestTools helpers for integration tests.

// GetDefaultPort returns 8081 or lt.default.port property-style env LT_DEFAULT_PORT.
func GetDefaultPort() int {
	if v := os.Getenv("LT_DEFAULT_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			return p
		}
	}
	return DefaultPort
}

// CheckAtURL GETs a URL and returns body text.
func CheckAtURL(rawURL string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
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

// CheckAtURLByPost POSTs form body to rawURL.
func CheckAtURLByPost(rawURL, postData string, headers map[string]string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(postData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
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

// FormEncode builds application/x-www-form-urlencoded data.
func FormEncode(values map[string]string) string {
	v := url.Values{}
	for k, val := range values {
		v.Set(k, val)
	}
	return v.Encode()
}
