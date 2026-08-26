package index

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// PatternRuleQueryBuilder ports dev.index.PatternRuleQueryBuilder surface
// (Lucene query construction deferred — builds a soft query string from tokens).
type PatternRuleQueryBuilder struct {
	// Field is the Lucene field name (Java default often "field").
	Field string
}

func NewPatternRuleQueryBuilder(field string) *PatternRuleQueryBuilder {
	if field == "" {
		field = "field"
	}
	return &PatternRuleQueryBuilder{Field: field}
}

// BuildFromTokens builds a soft AND query of token surfaces.
func (b *PatternRuleQueryBuilder) BuildFromTokens(tokens []*patterns.PatternToken) string {
	if b == nil {
		b = NewPatternRuleQueryBuilder("")
	}
	var parts []string
	for _, tok := range tokens {
		if tok == nil {
			continue
		}
		s := tok.Token
		if s == "" {
			continue
		}
		// escape soft quotes
		s = strings.ReplaceAll(s, `"`, `\"`)
		parts = append(parts, b.Field+`:"`+s+`"`)
	}
	if len(parts) == 0 {
		return "*:*"
	}
	return strings.Join(parts, " AND ")
}

// BuildSimple is a convenience for literal token strings.
func (b *PatternRuleQueryBuilder) BuildSimple(words ...string) string {
	var toks []*patterns.PatternToken
	for _, w := range words {
		toks = append(toks, patterns.NewPatternTokenBuilder().Token(w).Build())
	}
	return b.BuildFromTokens(toks)
}
