package patterns

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Substrings ports org.languagetool.rules.patterns.Substrings.
// Ordered fragments that must appear in text, optionally at start/end.
type Substrings struct {
	Substrings []string
	MustStart  bool
	MustEnd    bool
	minLength  int
}

func NewSubstrings(mustStart, mustEnd bool, parts []string) Substrings {
	minLen := 0
	for _, s := range parts {
		minLen += len(s)
	}
	return Substrings{Substrings: append([]string(nil), parts...), MustStart: mustStart, MustEnd: mustEnd, minLength: minLen}
}

func (s Substrings) withMinLength(min int) Substrings {
	out := s
	out.minLength = min
	return out
}

// CheckCanReplaceRegex returns a refined Substrings if it can fully replace regexp, else nil.
func (s Substrings) CheckCanReplaceRegex(regexp string) *Substrings {
	if s.MustStart || s.MustEnd {
		prefix := ""
		if s.MustStart && len(s.Substrings) > 0 {
			prefix = s.Substrings[0]
		}
		suffix := ""
		if s.MustEnd && len(s.Substrings) > 0 {
			suffix = s.Substrings[len(s.Substrings)-1]
		}
		if strings.HasPrefix(regexp, prefix) && strings.HasSuffix(regexp, suffix) &&
			len(regexp) == len(prefix)+len(suffix)+2 &&
			len(regexp) > len(prefix) && regexp[len(prefix)] == '.' {
			switch regexp[len(prefix)+1] {
			case '*':
				cp := s
				return &cp
			case '+':
				cp := s.withMinLength(s.minLength + 1)
				return &cp
			}
		}
	}
	joined := strings.Join(s.Substrings, ".*")
	expected := ""
	if !s.MustStart {
		expected += ".*"
	}
	expected += joined
	if !s.MustEnd {
		expected += ".*"
	}
	if regexp == expected {
		cp := s
		return &cp
	}
	return nil
}

func (s Substrings) String() string {
	open, close := "(", ")"
	if s.MustStart {
		open = "["
	}
	if s.MustEnd {
		close = "]"
	}
	return open + strings.Join(s.Substrings, ", ") + close
}

// Concat ports Substrings.concat.
func (s Substrings) Concat(another Substrings) Substrings {
	var parts []string
	switch {
	case len(another.Substrings) == 0:
		parts = s.Substrings
	case len(s.Substrings) == 0:
		parts = another.Substrings
	case s.MustEnd && another.MustStart:
		parts = make([]string, 0, len(s.Substrings)+len(another.Substrings)-1)
		parts = append(parts, s.Substrings[:len(s.Substrings)-1]...)
		parts = append(parts, s.Substrings[len(s.Substrings)-1]+another.Substrings[0])
		if len(another.Substrings) > 1 {
			parts = append(parts, another.Substrings[1:]...)
		}
	default:
		parts = append(append([]string{}, s.Substrings...), another.Substrings...)
	}
	return NewSubstrings(s.MustStart, another.MustEnd, parts)
}

// Find returns index of first substring or -1.
func (s Substrings) Find(text string, caseSensitive bool) int {
	if len(s.Substrings) == 0 || len(text) < s.minLength {
		return -1
	}
	start := indexOf(text, s.Substrings[0], caseSensitive, 0)
	if start < 0 {
		return -1
	}
	if len(s.Substrings) > 1 && !s.containsSubstrings(text, caseSensitive, start+len(s.Substrings[0]), 1) {
		return -1
	}
	return start
}

// Matches reports whether text contains all required substrings with start/end constraints.
func (s Substrings) Matches(text string, caseSensitive bool) bool {
	if len(s.Substrings) == 0 {
		return true
	}
	if len(text) < s.minLength {
		return false
	}
	if s.MustStart {
		first := s.Substrings[0]
		if !regionMatches(text, 0, first, caseSensitive) {
			return false
		}
	}
	if s.MustEnd {
		last := s.Substrings[len(s.Substrings)-1]
		if len(text) < len(last) || !regionMatches(text, len(text)-len(last), last, caseSensitive) {
			return false
		}
	}
	if len(s.Substrings) == 1 && (s.MustStart || s.MustEnd) {
		return true
	}
	from, firstIdx := 0, 0
	if s.MustStart {
		from = len(s.Substrings[0])
		firstIdx = 1
	}
	return s.containsSubstrings(text, caseSensitive, from, firstIdx)
}

func (s Substrings) containsSubstrings(text string, caseSensitive bool, textPos, firstSubstringIndex int) bool {
	for i := firstSubstringIndex; i < len(s.Substrings); i++ {
		textPos = indexOf(text, s.Substrings[i], caseSensitive, textPos)
		if textPos < 0 {
			return false
		}
		textPos += len(s.Substrings[i])
	}
	return true
}

func indexOf(text, substring string, caseSensitive bool, from int) int {
	if from > len(text) {
		return -1
	}
	if caseSensitive {
		i := strings.Index(text[from:], substring)
		if i < 0 {
			return -1
		}
		return from + i
	}
	return indexOfIgnoreCase(text, substring, from)
}

func regionMatches(text string, at int, sub string, caseSensitive bool) bool {
	if at < 0 || at+len(sub) > len(text) {
		return false
	}
	if caseSensitive {
		return text[at:at+len(sub)] == sub
	}
	return strings.EqualFold(text[at:at+len(sub)], sub)
}

func indexOfIgnoreCase(text, substring string, from int) int {
	if substring == "" {
		return from
	}
	first, _ := utf8.DecodeRuneInString(substring)
	up := unicode.ToUpper(first)
	low := unicode.ToLower(first)
	for i := from; i < len(text); {
		// scan for first char
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == up || r == low || unicode.ToUpper(r) == up || unicode.ToLower(r) == low {
			if regionMatches(text, i, substring, false) {
				return i
			}
		}
		i += size
		if size == 0 {
			i++
		}
	}
	return -1
}
