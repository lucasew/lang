package chunking

import (
	"fmt"
	"regexp"
	"strings"
)

// TokenPredicate ports org.languagetool.chunking.TokenPredicate without knowitall.
// Descriptions: "word", "string=word", "regex=...", "regexCS=...", "chunk=...", "pos=...".
type TokenPredicate struct {
	Description   string
	CaseSensitive bool
	match         func(ChunkTaggedToken) bool
}

func NewTokenPredicate(description string, caseSensitive bool) *TokenPredicate {
	p := &TokenPredicate{Description: description, CaseSensitive: caseSensitive}
	p.match = p.compile(description, caseSensitive)
	return p
}

func (p *TokenPredicate) Apply(token ChunkTaggedToken) bool {
	return p.match(token)
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func (p *TokenPredicate) compile(description string, caseSensitive bool) func(ChunkTaggedToken) bool {
	parts := strings.SplitN(description, "=", 2)
	var exprType, exprValue string
	if len(parts) == 1 {
		exprType = "string"
		exprValue = unquote(parts[0])
	} else {
		exprType = parts[0]
		exprValue = unquote(parts[1])
	}
	switch exprType {
	case "string":
		if caseSensitive {
			return func(t ChunkTaggedToken) bool { return t.Token == exprValue }
		}
		return func(t ChunkTaggedToken) bool { return strings.EqualFold(t.Token, exprValue) }
	case "regex":
		re, err := regexp.Compile("(?i)" + exprValue)
		if err != nil {
			panic(err)
		}
		if caseSensitive {
			re, err = regexp.Compile(exprValue)
			if err != nil {
				panic(err)
			}
		}
		return func(t ChunkTaggedToken) bool { return re.MatchString(t.Token) }
	case "regexCS":
		re, err := regexp.Compile(exprValue)
		if err != nil {
			panic(err)
		}
		return func(t ChunkTaggedToken) bool { return re.MatchString(t.Token) }
	case "chunk":
		re, err := regexp.Compile(exprValue)
		if err != nil {
			// treat as full-string regexp like StringMatcher.regexp
			re = regexp.MustCompile("^(?:" + exprValue + ")$")
		}
		return func(t ChunkTaggedToken) bool {
			for _, ct := range t.ChunkTags {
				if re.MatchString(ct.GetChunkTag()) {
					return true
				}
			}
			return false
		}
	case "pos":
		return func(t ChunkTaggedToken) bool {
			if t.Readings == nil {
				return false
			}
			for _, reading := range t.Readings.GetReadings() {
				if pt := reading.GetPOSTag(); pt != nil && strings.Contains(*pt, exprValue) {
					return true
				}
			}
			return false
		}
	default:
		panic(fmt.Sprintf("Could not parse expression type: %s", exprType))
	}
}
