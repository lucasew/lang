package chunking

import "strings"

// TokenExpression ports knowitall Expression.BaseExpression for one token.
type TokenExpression interface {
	Apply(token ChunkTaggedToken) bool
	Source() string
}

type tokenExpr struct {
	src string
	fn  func(ChunkTaggedToken) bool
}

func (e tokenExpr) Apply(token ChunkTaggedToken) bool { return e.fn(token) }
func (e tokenExpr) Source() string                    { return e.src }

// TokenExpressionFactory ports org.languagetool.chunking.TokenExpressionFactory.
// Supports simple predicates and boolean combinations:
//   string=foo / regex=... / chunk=... / pos=...
//   AND of space-separated atoms:  string=the pos=DT
//   OR with | :                    string=a|string=an  (not inside regex=)
// Full knowitall LogicExpression is not implemented.
type TokenExpressionFactory struct {
	CaseSensitive bool
}

func NewTokenExpressionFactory(caseSensitive bool) *TokenExpressionFactory {
	return &TokenExpressionFactory{CaseSensitive: caseSensitive}
}

// Create compiles expr into a token matcher.
func (f *TokenExpressionFactory) Create(expr string) TokenExpression {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return tokenExpr{src: expr, fn: func(ChunkTaggedToken) bool { return true }}
	}
	// OR alternatives (avoid splitting regex=a|b)
	if strings.Contains(expr, "|") && !strings.HasPrefix(expr, "regex") {
		parts := strings.Split(expr, "|")
		if len(parts) > 1 {
			var alts []TokenExpression
			for _, p := range parts {
				alts = append(alts, f.Create(strings.TrimSpace(p)))
			}
			return tokenExpr{src: expr, fn: func(t ChunkTaggedToken) bool {
				for _, a := range alts {
					if a.Apply(t) {
						return true
					}
				}
				return false
			}}
		}
	}
	// AND of whitespace-separated predicates
	atoms := strings.Fields(expr)
	if len(atoms) > 1 {
		var preds []TokenExpression
		for _, a := range atoms {
			preds = append(preds, f.atom(a))
		}
		return tokenExpr{src: expr, fn: func(t ChunkTaggedToken) bool {
			for _, p := range preds {
				if !p.Apply(t) {
					return false
				}
			}
			return true
		}}
	}
	return f.atom(expr)
}

func (f *TokenExpressionFactory) atom(expr string) TokenExpression {
	cs := false
	if f != nil {
		cs = f.CaseSensitive
	}
	if !strings.Contains(expr, "=") {
		expr = "string=" + expr
	}
	p := NewTokenPredicate(expr, cs)
	return tokenExpr{src: expr, fn: p.Apply}
}
