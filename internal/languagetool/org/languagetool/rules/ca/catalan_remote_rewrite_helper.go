package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"os"
	"strconv"
)

// CatalanRemoteRewriteHelper ports env/config surface of CatalanRemoteRewriteHelper
// (HTTP client deferred).
type CatalanRemoteRewriteConfig struct {
	ServerURL string
	APIKey    string
	Provider  string
	Model     string
	TimeoutMS int
	MaxChars  int
}

// DefaultRemoteRewriteConfig reads CA_REMOTE_REWRITE_* environment variables.
func DefaultRemoteRewriteConfig() CatalanRemoteRewriteConfig {
	timeout := 5000
	if v := os.Getenv("CA_REMOTE_REWRITE_TIME_OUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			timeout = n
		}
	}
	return CatalanRemoteRewriteConfig{
		ServerURL: tools.JavaStringTrim(os.Getenv("CA_REMOTE_REWRITE_SERVER")),
		APIKey:    os.Getenv("CA_REMOTE_REWRITE_API_KEY"),
		Provider:  os.Getenv("CA_REMOTE_REWRITE_PROVIDER"),
		Model:     os.Getenv("CA_REMOTE_REWRITE_MODEL"),
		TimeoutMS: timeout,
		MaxChars:  1200,
	}
}

// IsRemoteServiceAvailable reports whether a rewrite server URL is configured.
func (c CatalanRemoteRewriteConfig) IsRemoteServiceAvailable() bool {
	return c.ServerURL != ""
}

// AcceptsSentence reports whether the sentence is short enough to send remotely.
func (c CatalanRemoteRewriteConfig) AcceptsSentence(sentence string) bool {
	max := c.MaxChars
	if max <= 0 {
		max = 1200
	}
	return tokenizers.UTF16Len(tools.JavaStringTrim(sentence)) <= max && tools.JavaStringTrim(sentence) != ""
}
