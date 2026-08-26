package eval

// Span ports org.languagetool.dev.eval.Span.
type Span struct {
	StartPos int
	EndPos   int
}

func NewSpan(start, end int) Span { return Span{StartPos: start, EndPos: end} }

// Covers reports whether s fully covers other.
func (s Span) Covers(other Span) bool {
	return s.StartPos <= other.StartPos && s.EndPos >= other.EndPos
}

// Overlaps reports any position overlap.
func (s Span) Overlaps(other Span) bool {
	return s.StartPos < other.EndPos && other.StartPos < s.EndPos
}
