package tools

import (
	"fmt"
	"net/url"
)

// GetURL ports Tools.getUrl — parse URL or panic (Java RuntimeException wrapping
// MalformedURLException). Returns the validated absolute URL string (callers
// store URL as string on rules; Java returns java.net.URL).
func GetURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		panic(fmt.Sprintf("RuntimeException: %v", err))
	}
	// Java new URL requires a protocol for absolute URLs used by LT.
	if u.Scheme == "" || u.Host == "" {
		// Match MalformedURLException for protocol-less strings like "not a url"
		panic(fmt.Sprintf("RuntimeException: no protocol: %s", raw))
	}
	return u.String()
}
