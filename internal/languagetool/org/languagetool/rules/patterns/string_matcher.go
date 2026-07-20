package patterns

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MaxMatchLength ports StringMatcher.MAX_MATCH_LENGTH (Java s.length() = UTF-16 units).
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
	// re is used when possible values cannot be enumerated and substrings are not sufficient
	re *regexp.Regexp
	// javaRE is used when the pattern needs Java lookaround (RE2 cannot compile it).
	// Unavoidable Go mapping of java.util.regex for LT XML postag/surface regexps.
	javaRE *javaRegexp
	// required ports getRequiredSubstrings filter (and checkCanReplaceRegex when sufficient)
	required *Substrings
	// substringsSufficient when true, required alone decides matches (no regex)
	substringsSufficient bool
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
	// Map Java/PCRE surface that RE2 rejects (\u, lookaround after flags strip).
	pattern = normalizeJavaRegexp(pattern)

	// always compile to validate syntax (Java Pattern.compile)
	flags := ""
	if !caseSensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + "^(?:" + pattern + ")$")
	if err != nil {
		// Java supports (?!...) / (?=...); RE2 does not. Use lookaround engine.
		if needsJavaRegexp(pattern) {
			jr, jerr := compileJavaRegexp(pattern, caseSensitive)
			if jerr != nil {
				panic(fmt.Errorf("StringMatcher: java lookaround compile failed for %q: %v (re2: %v)", pattern, jerr, err))
			}
			return &StringMatcher{
				Pattern:       pattern,
				CaseSensitive: caseSensitive,
				IsRegExp:      true,
				javaRE:        jr,
			}
		}
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

	// Java: required substrings + checkCanReplaceRegex for exhaustive path
	required := getRequiredSubstrings(pattern)
	var substrings *Substrings
	sufficient := false
	if required != nil {
		if exhaustive := required.CheckCanReplaceRegex(pattern); exhaustive != nil {
			substrings = exhaustive
			sufficient = true
		} else {
			substrings = required
		}
	}
	m := &StringMatcher{
		Pattern:              pattern,
		CaseSensitive:        caseSensitive,
		IsRegExp:             true,
		required:             substrings,
		substringsSufficient: sufficient,
	}
	if !sufficient {
		m.re = re
	}
	return m
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

// Matches ports StringMatcher.matches.
func (m *StringMatcher) Matches(s string) bool {
	if m == nil {
		return false
	}
	// Java: s.length() > MAX_MATCH_LENGTH (UTF-16 code units)
	if tokenizers.UTF16Len(s) > MaxMatchLength {
		return false
	}
	if m.possibleValues != nil {
		if m.CaseSensitive {
			_, ok := m.possibleValues[s]
			return ok
		}
		if m.possibleSorted != nil {
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
	// required-substring prefilter (and full decision when checkCanReplaceRegex succeeded)
	if m.required != nil && !m.required.Matches(s, m.CaseSensitive) {
		return false
	}
	if m.substringsSufficient {
		return true
	}
	if m.javaRE != nil {
		return m.javaRE.fullMatch(s)
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

// GetRequiredSubstrings ports StringMatcher.getRequiredSubstrings.
// Returns nil when no necessary substrings can be proven.
func GetRequiredSubstrings(regexp string) *Substrings {
	return getRequiredSubstrings(regexp)
}

func stripRegexpAnchors(re string) string {
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
	return re
}

func getPossibleRegexpValues(re string) map[string]struct{} {
	re = stripRegexpAnchors(re)
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

// unknownSubstrings ports Java Substrings UNKNOWN (no fixed fragments).
func unknownSubstrings() Substrings {
	return NewSubstrings(false, false, nil)
}

func getRequiredSubstrings(re string) *Substrings {
	re = stripRegexpAnchors(re)
	p := &subParser{s: re}
	result, ok := p.disjunction()
	if !ok || p.pos != len(p.s) {
		return nil
	}
	if len(result.Substrings) == 0 {
		return nil
	}
	cp := result
	return &cp
}

// subParser ports RegexpParser<Substrings> for getRequiredSubstrings.
type subParser struct {
	s   string
	pos int
}

func (p *subParser) disjunction() (Substrings, bool) {
	left, ok := p.concatenation()
	if !ok {
		return Substrings{}, false
	}
	if p.pos >= len(p.s) || p.s[p.pos] != '|' {
		return left, true
	}
	// any top-level | → UNKNOWN (Java handleOr)
	for p.pos < len(p.s) && p.s[p.pos] == '|' {
		p.pos++
		if _, ok := p.concatenation(); !ok {
			return Substrings{}, false
		}
	}
	return unknownSubstrings(), true
}

func (p *subParser) concatenation() (Substrings, bool) {
	left, ok := p.postfix()
	if !ok {
		return Substrings{}, false
	}
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		if c == ')' || c == '|' {
			break
		}
		if strings.ContainsRune("?$^{}*+", rune(c)) {
			return Substrings{}, false
		}
		right, ok := p.postfix()
		if !ok {
			return Substrings{}, false
		}
		left = left.Concat(right)
	}
	return left, true
}

func (p *subParser) postfix() (Substrings, bool) {
	base, ok := p.atom()
	if !ok {
		return Substrings{}, false
	}
	if p.pos < len(p.s) {
		next := p.s[p.pos]
		if next == '{' {
			closing := strings.IndexByte(p.s[p.pos+1:], '}')
			if closing < 0 {
				return Substrings{}, false
			}
			p.pos += closing + 2
			base = unknownSubstrings()
			if p.pos >= len(p.s) {
				return base, true
			}
			next = p.s[p.pos]
		}
		if next == '?' || next == '*' || next == '+' {
			p.pos++
			// Java optional → UNKNOWN for required-substrings mode
			return unknownSubstrings(), true
		}
	}
	return base, true
}

func (p *subParser) atom() (Substrings, bool) {
	if p.pos >= len(p.s) {
		return NewSubstrings(true, true, []string{""}), true
	}
	switch p.s[p.pos] {
	case '(':
		p.pos++
		if p.pos < len(p.s) && p.s[p.pos] == '?' {
			p.pos++
			if p.pos >= len(p.s) || p.s[p.pos] != ':' {
				return Substrings{}, false
			}
			p.pos++
		}
		inner, ok := p.disjunction()
		if !ok {
			return Substrings{}, false
		}
		if p.pos >= len(p.s) || p.s[p.pos] != ')' {
			return Substrings{}, false
		}
		p.pos++
		return inner, true
	case '[':
		return p.charClass()
	case '\\':
		p.pos++
		ch, ok := p.escape()
		if !ok {
			return Substrings{}, false
		}
		if ch == nil {
			return unknownSubstrings(), true
		}
		return NewSubstrings(true, true, []string{string(*ch)}), true
	case '.':
		p.pos++
		return unknownSubstrings(), true
	default:
		start := p.pos
		for p.pos < len(p.s) {
			c := p.s[p.pos]
			if strings.ContainsRune(")|?$^{}*+([\\.", rune(c)) {
				break
			}
			p.pos++
		}
		if start+1 < p.pos && p.pos < len(p.s) && p.s[p.pos] == '?' {
			p.pos--
		}
		return NewSubstrings(true, true, []string{p.s[start:p.pos]}), true
	}
}

func (p *subParser) charClass() (Substrings, bool) {
	p.pos++ // skip [
	start := p.pos
	var options []rune
	negated := false
	okOptions := true
	for p.pos < len(p.s) {
		c, size := utf8.DecodeRuneInString(p.s[p.pos:])
		p.pos += size
		if c == ']' {
			break
		}
		if c == '^' && p.pos == start+size {
			negated = true
			okOptions = false
			options = nil
			continue
		}
		if c == '-' && len(options) > 0 && p.pos < len(p.s) && p.s[p.pos] != ']' {
			last := options[len(options)-1]
			next, nsize := utf8.DecodeRuneInString(p.s[p.pos:])
			p.pos += nsize
			if next == '\\' || int(next)-int(last) > 10 {
				okOptions = false
				options = nil
			}
			if okOptions {
				for r := last + 1; r <= next; r++ {
					options = append(options, r)
				}
			}
			continue
		}
		if c == '[' {
			return Substrings{}, false
		}
		if c == '\\' {
			ch, eok := p.escape()
			if !eok {
				return Substrings{}, false
			}
			if ch == nil {
				okOptions = false
				options = nil
				continue
			}
			if okOptions {
				options = append(options, *ch)
			}
			continue
		}
		if okOptions {
			options = append(options, c)
		}
	}
	if negated || !okOptions || len(options) == 0 {
		return unknownSubstrings(), true
	}
	if len(options) == 1 {
		return NewSubstrings(true, true, []string{string(options[0])}), true
	}
	// multi-char class → handleOr → UNKNOWN
	return unknownSubstrings(), true
}

func (p *subParser) escape() (*rune, bool) {
	if p.pos >= len(p.s) {
		return nil, false
	}
	next, size := utf8.DecodeRuneInString(p.s[p.pos:])
	p.pos += size
	if strings.ContainsRune("0xucpP", next) {
		return nil, false
	}
	if unicode.IsLetter(next) || unicode.IsDigit(next) {
		return nil, true // class escape → unknown
	}
	return &next, true
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
		c, size := utf8.DecodeRuneInString(p.s[p.pos:])
		p.pos += size
		if c == ']' {
			break
		}
		if c == '^' && p.pos == start+size {
			// just after [
			negated = true
			options = nil
			continue
		}
		if c == '-' && len(options) > 0 && p.pos < len(p.s) && p.s[p.pos] != ']' {
			last := options[len(options)-1]
			next, nsize := utf8.DecodeRuneInString(p.s[p.pos:])
			p.pos += nsize
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
	next, size := utf8.DecodeRuneInString(p.s[p.pos:])
	p.pos += size
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
