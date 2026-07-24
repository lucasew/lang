package chunking

import (
	"fmt"
	"regexp"
	"strings"
)

// TokenPredicate ports org.languagetool.chunking.TokenPredicate without knowitall.
// Descriptions: "word", "string=word", "regex=...", "regexCS=...", "chunk=...", "pos=...", "posre=...", "posregex=...".
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

// unquote ports TokenPredicate.unquote — single-quoted strings only (Java uses ').
func unquote(s string) string {
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") && len(s) >= 2 {
		return s[1 : len(s)-1]
	}
	return s
}

func (p *TokenPredicate) compile(description string, caseSensitive bool) func(ChunkTaggedToken) bool {
	// Java: description.split("=") — all splits; length 1 or 2 only
	parts := strings.Split(description, "=")
	var exprType, exprValue string
	if len(parts) == 1 {
		exprType = "string"
		exprValue = unquote(parts[0])
	} else if len(parts) == 2 {
		exprType = parts[0]
		exprValue = unquote(parts[1])
	} else {
		panic(fmt.Sprintf("Could not parse expression: %s", description))
	}

	switch exprType {
	case "string", "regex", "regexCS":
		// StringMatcher.create(exprValue, isRegex, caseSensitive || regexCS)
		isRegex := exprType != "string"
		cs := caseSensitive || exprType == "regexCS"
		if !isRegex {
			if cs {
				return func(t ChunkTaggedToken) bool { return t.Token == exprValue }
			}
			return func(t ChunkTaggedToken) bool { return strings.EqualFold(t.Token, exprValue) }
		}
		pat := exprValue
		if !cs {
			pat = "(?i)" + exprValue
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			panic(err)
		}
		return func(t ChunkTaggedToken) bool { return re.MatchString(t.Token) }

	case "chunk":
		// StringMatcher.regexp — full match semantics via anchored pattern when possible
		re, err := regexp.Compile("^(?:" + exprValue + ")$")
		if err != nil {
			re = regexp.MustCompile(exprValue)
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

	case "posre", "posregex":
		re, err := regexp.Compile("^(?:" + exprValue + ")$")
		if err != nil {
			re = regexp.MustCompile(exprValue)
		}
		return func(t ChunkTaggedToken) bool {
			if t.Readings == nil {
				return false
			}
			for _, reading := range t.Readings.GetReadings() {
				if pt := reading.GetPOSTag(); pt != nil && re.MatchString(*pt) {
					return true
				}
			}
			return false
		}

	default:
		panic(fmt.Sprintf("Expression type not supported: '%s'", exprType))
	}
}
