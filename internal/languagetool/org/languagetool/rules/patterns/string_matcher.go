package patterns

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// MaxMatchLength ports StringMatcher.MAX_MATCH_LENGTH.
const MaxMatchLength = 250

// StringMatcher ports org.languagetool.rules.patterns.StringMatcher.
// Encapsulates a pattern plus case-sensitivity / regexp matching.
type StringMatcher struct {
	Pattern       string
	CaseSensitive bool
	IsRegExp      bool
	// possibleValues when non-nil is the exhaustive set of accept strings.
	possibleValues map[string]struct{}
	// possibleSorted used for case-insensitive set membership via binary search
	possibleSorted []string
	// re is used when possible values cannot be enumerated
	re *regexp.Regexp
}

// NewStringMatcherRegexp creates a case-sensitive regexp matcher.
func NewStringMatcherRegexp(pattern string) *StringMatcher {
	return NewStringMatcher(pattern, true, true)
}

// NewStringMatcher ports StringMatcher.create.
func NewStringMatcher(pattern string, isRegExp, caseSensitive bool) *StringMatcher {
	if !isRegExp || pattern == "\\0" {
		return stringEqualsMatcher(pattern, isRegExp, caseSensitive)
	}
	// always compile to validate syntax (Java Pattern.compile)
	flags := ""
	if !caseSensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + "^(?:" + pattern + ")$")
	if err != nil {
		panic(err) // match Java PatternSyntaxException
	}

	if vals := getPossibleRegexpValues(pattern); vals != nil {
		if len(vals) == 1 {
			only := ""
			for v := range vals {
				only = v
			}
			return stringEqualsMatcher(only, true, caseSensitive)
		}
		if !caseSensitive {
			sorted := make([]string, 0, len(vals))
			for v := range vals {
				sorted = append(sorted, v)
			}
			sort.Slice(sorted, func(i, j int) bool {
				return strings.ToLower(sorted[i]) < strings.ToLower(sorted[j])
			})
			return &StringMatcher{
				Pattern:        pattern,
				CaseSensitive:  false,
				IsRegExp:       true,
				possibleValues: vals,
				possibleSorted: sorted,
			}
		}
		return &StringMatcher{
			Pattern:        pattern,
			CaseSensitive:  true,
			IsRegExp:       true,
			possibleValues: vals,
		}
	}
	return &StringMatcher{
		Pattern:       pattern,
		CaseSensitive: caseSensitive,
		IsRegExp:      true,
		re:            re,
	}
}

func stringEqualsMatcher(pattern string, isRegExp, caseSensitive bool) *StringMatcher {
	return &StringMatcher{
		Pattern:        pattern,
		CaseSensitive:  caseSensitive,
		IsRegExp:       isRegExp,
		possibleValues: map[string]struct{}{pattern: {}},
	}
}

// GetPossibleValues returns all values this matcher can accept, or nil if unknown.
func (m *StringMatcher) GetPossibleValues() map[string]struct{} {
	if m == nil {
		return nil
	}
	if m.possibleValues == nil {
		return nil
	}
	out := make(map[string]struct{}, len(m.possibleValues))
	for k := range m.possibleValues {
		out[k] = struct{}{}
	}
	return out
}

// Matches reports whether s is accepted.
func (m *StringMatcher) Matches(s string) bool {
	if m == nil {
		return false
	}
	if len(s) > MaxMatchLength {
		return false
	}
	if m.possibleValues != nil && m.re == nil {
		if m.CaseSensitive {
			_, ok := m.possibleValues[s]
			return ok
		}
		if m.possibleSorted != nil {
			// binary search case-insensitive
			i := sort.Search(len(m.possibleSorted), func(i int) bool {
				return strings.ToLower(m.possibleSorted[i]) >= strings.ToLower(s)
			})
			if i < len(m.possibleSorted) && strings.EqualFold(m.possibleSorted[i], s) {
				return true
			}
			return false
		}
		for k := range m.possibleValues {
			if strings.EqualFold(k, s) {
				return true
			}
		}
		return false
	}
	if m.re != nil {
		return m.re.MatchString(s)
	}
	if m.CaseSensitive {
		return s == m.Pattern
	}
	return strings.EqualFold(s, m.Pattern)
}

// GetPossibleRegexpValues ports StringMatcher.getPossibleRegexpValues for common patterns.
// Returns nil when the set cannot be enumerated.
func GetPossibleRegexpValues(regexp string) map[string]struct{} {
	return getPossibleRegexpValues(regexp)
}

func getPossibleRegexpValues(re string) map[string]struct{} {
	// strip common anchors
	if strings.HasPrefix(re, "\\b") {
		re = re[2:]
	}
	if strings.HasPrefix(re, "^") {
		re = re[1:]
	}
	if strings.HasSuffix(re, "\\b") && !strings.HasSuffix(re, "\\\\b") {
		re = re[:len(re)-2]
	}
	if strings.HasSuffix(re, "$") && !strings.HasSuffix(re, "\\$") {
		re = re[:len(re)-1]
	}
	vals, ok := parseRegexpAlternatives(re)
	if !ok {
		return nil
	}
	out := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		out[v] = struct{}{}
	}
	return out
}

// parseRegexpAlternatives enumerates finite languages for a subset of RE syntax.
// Returns ok=false when too complex.
func parseRegexpAlternatives(re string) ([]string, bool) {
	p := &reParser{s: re}
	vals, ok := p.disjunction()
	if !ok || p.pos != len(p.s) {
		return nil, false
	}
	return vals, true
}

type reParser struct {
	s   string
	pos int
}

func (p *reParser) disjunction() ([]string, bool) {
	left, ok := p.concatenation()
	if !ok {
		return nil, false
	}
	for p.pos < len(p.s) && p.s[p.pos] == '|' {
		p.pos++
		right, ok := p.concatenation()
		if !ok {
			return nil, false
		}
		left = append(left, right...)
	}
	return left, true
}

func (p *reParser) concatenation() ([]string, bool) {
	left, ok := p.postfix()
	if !ok {
		return nil, false
	}
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		if c == ')' || c == '|' {
			break
		}
		if strings.ContainsRune("?$^{}*+", rune(c)) {
			return nil, false
		}
		right, ok := p.postfix()
		if !ok {
			return nil, false
		}
		left = concatSets(left, right)
	}
	return left, true
}

func (p *reParser) postfix() ([]string, bool) {
	base, ok := p.atom()
	if !ok {
		return nil, false
	}
	if p.pos < len(p.s) {
		next := p.s[p.pos]
		if next == '{' {
			// quantifier → too complex
			return nil, false
		}
		if next == '?' {
			p.pos++
			// optional: base ∪ {""}
			out := append([]string{""}, base...)
			return out, true
		}
		if next == '*' || next == '+' {
			return nil, false
		}
	}
	return base, true
}

func (p *reParser) atom() ([]string, bool) {
	if p.pos >= len(p.s) {
		return []string{""}, true
	}
	switch p.s[p.pos] {
	case '(':
		p.pos++
		if p.pos < len(p.s) && p.s[p.pos] == '?' {
			p.pos++
			if p.pos >= len(p.s) || p.s[p.pos] != ':' {
				return nil, false
			}
			p.pos++
		}
		inner, ok := p.disjunction()
		if !ok {
			return nil, false
		}
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return nil, false
		}
		p.pos++
		return inner, true
	case '[':
		return p.charClass()
	case '\\':
		p.pos++
		ch, ok := p.escape()
		if !ok {
			return nil, false
		}
		if ch == nil {
			return nil, false // \d etc.
		}
		return []string{string(*ch)}, true
	case '.':
		return nil, false
	default:
		start := p.pos
		for p.pos < len(p.s) {
			c := p.s[p.pos]
			if strings.ContainsRune(")|?$^{}*+([\\.", rune(c)) {
				break
			}
			p.pos++
		}
		// Java: if next is ? and multi-char literal, shrink so ? applies to last char only
		if start+1 < p.pos && p.pos < len(p.s) && p.s[p.pos] == '?' {
			p.pos--
		}
		return []string{p.s[start:p.pos]}, true
	}
}

func (p *reParser) charClass() ([]string, bool) {
	p.pos++ // skip [
	start := p.pos
	var options []rune
	negated := false
	for p.pos < len(p.s) {
		c := rune(p.s[p.pos])
		p.pos++
		if c == ']' {
			break
		}
		if c == '^' && p.pos == start+1 {
			// just after [
			negated = true
			options = nil
			continue
		}
		if c == '-' && len(options) > 0 && p.pos < len(p.s) && p.s[p.pos] != ']' {
			last := options[len(options)-1]
			next := rune(p.s[p.pos])
			p.pos++
			if next == '\\' || int(next)-int(last) > 10 {
				return nil, false
			}
			for r := last + 1; r <= next; r++ {
				options = append(options, r)
			}
			continue
		}
		if c == '[' {
			return nil, false
		}
		if c == '\\' {
			ch, ok := p.escape()
			if !ok || ch == nil {
				return nil, false
			}
			if !negated {
				options = append(options, *ch)
			}
			continue
		}
		if !negated {
			options = append(options, c)
		}
	}
	if negated || len(options) == 0 {
		return nil, false
	}
	out := make([]string, 0, len(options))
	for _, r := range options {
		out = append(out, string(r))
	}
	return out, true
}

func (p *reParser) escape() (*rune, bool) {
	if p.pos >= len(p.s) {
		return nil, false
	}
	next := rune(p.s[p.pos])
	p.pos++
	if strings.ContainsRune("0xucpP", next) {
		return nil, false
	}
	if unicode.IsLetter(next) || unicode.IsDigit(next) {
		return nil, true // unknown escape class
	}
	return &next, true
}

func concatSets(a, b []string) []string {
	out := make([]string, 0, len(a)*len(b))
	for _, x := range a {
		for _, y := range b {
			out = append(out, x+y)
		}
	}
	return out
}
