package patterns

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PasswordAuthenticator ports org.languagetool.rules.patterns.PasswordAuthenticator.
// Extracts user:password from a URL's userinfo.
type PasswordCredentials struct {
	Username string
	Password string
}

// GetPasswordAuthenticationFromURL ports getPasswordAuthentication for a given URL.
func GetPasswordAuthenticationFromURL(u *url.URL) (*PasswordCredentials, error) {
	if u == nil {
		return nil, nil
	}
	userInfo := u.User
	if userInfo == nil {
		return nil, nil
	}
	username := userInfo.Username()
	password, ok := userInfo.Password()
	if !ok {
		// bare username without password — treat like empty userInfo for password part
		info := u.String()
		// fallback: parse raw userinfo if present
		if u.User != nil && username != "" && !ok {
			return nil, fmt.Errorf("Invalid userInfo format, expected 'user:password': %s", username)
		}
		_ = info
		return nil, nil
	}
	if tools.IsEmptyStr(username) && tools.IsEmptyStr(password) {
		return nil, nil
	}
	return &PasswordCredentials{Username: username, Password: password}, nil
}

// ParseUserInfo parses "user:password" (same validation as Java).
func ParseUserInfo(userInfo string) (*PasswordCredentials, error) {
	if tools.IsEmptyStr(userInfo) {
		return nil, nil
	}
	parts := strings.Split(userInfo, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid userInfo format, expected 'user:password': %s", userInfo)
	}
	return &PasswordCredentials{Username: parts[0], Password: parts[1]}, nil
}
