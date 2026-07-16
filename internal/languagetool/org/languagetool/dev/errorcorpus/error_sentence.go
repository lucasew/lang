package errorcorpus

// Error is a corpus gold span (start inclusive, end exclusive).
type Error struct {
	StartPos   int
	EndPos     int
	Correction string
}

// ErrorSentence ports org.languagetool.dev.errorcorpus.ErrorSentence.
type ErrorSentence struct {
	MarkupText string
	Errors     []Error
}

func NewErrorSentence(markupText string, errors []Error) *ErrorSentence {
	return &ErrorSentence{MarkupText: markupText, Errors: append([]Error(nil), errors...)}
}

// MatchSpan is a minimal stand-in for RuleMatch from/to positions.
type MatchSpan struct {
	FromPos int
	ToPos   int
}

// HasErrorCoveredByMatch reports whether any gold error is fully covered by the match span.
func (s *ErrorSentence) HasErrorCoveredByMatch(m MatchSpan) bool {
	if s == nil {
		return false
	}
	for _, e := range s.Errors {
		if m.FromPos <= e.StartPos && m.ToPos >= e.EndPos {
			return true
		}
	}
	return false
}

// HasErrorOverlappingWithMatch reports any overlap between match and a gold error.
func (s *ErrorSentence) HasErrorOverlappingWithMatch(m MatchSpan) bool {
	if s == nil {
		return false
	}
	for _, e := range s.Errors {
		if (m.FromPos <= e.StartPos && m.ToPos >= e.EndPos) ||
			(m.FromPos >= e.StartPos && m.FromPos <= e.EndPos) ||
			(m.ToPos >= e.StartPos && m.ToPos <= e.EndPos) {
			return true
		}
	}
	return false
}
