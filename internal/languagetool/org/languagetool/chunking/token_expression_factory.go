package chunking

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
// Uses LogicExpression (& | ! ()) over TokenPredicate atoms — same as Java.
type TokenExpressionFactory struct {
	CaseSensitive bool
}

func NewTokenExpressionFactory(caseSensitive bool) *TokenExpressionFactory {
	return &TokenExpressionFactory{CaseSensitive: caseSensitive}
}

// Create compiles expr into a token matcher (Java: LogicExpression.compile + TokenPredicate).
func (f *TokenExpressionFactory) Create(expr string) TokenExpression {
	cs := false
	if f != nil {
		cs = f.CaseSensitive
	}
	logic := CompileLogicExpression(expr, func(atom string) func(ChunkTaggedToken) bool {
		return NewTokenPredicate(atom, cs).Apply
	})
	return tokenExpr{src: expr, fn: logic.Apply}
}
