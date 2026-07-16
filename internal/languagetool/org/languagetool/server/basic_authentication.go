package server

import (
	"encoding/base64"
	"strings"
)

// BasicAuthentication ports org.languagetool.server.BasicAuthentication.
type BasicAuthentication struct {
	User     string
	Password string
}

func ParseBasicAuthentication(authHeader string) (*BasicAuthentication, error) {
	if !strings.HasPrefix(authHeader, "Basic ") {
		return nil, NewAuthError("Expected Basic Authentication")
	}
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, NewAuthError("Expected Basic Authentication")
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, NewAuthError("Expected Basic Authentication")
	}
	return &BasicAuthentication{User: parts[0], Password: parts[1]}, nil
}
