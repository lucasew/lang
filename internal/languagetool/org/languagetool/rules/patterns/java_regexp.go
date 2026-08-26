package patterns

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// javaRegexp is a minimal full-string matcher for Java Pattern features that
// Go RE2 lacks — primarily (?!...) / (?=...) lookaround — used by LT XML.
// It is an unavoidable Go mapping of java.util.regex for StringMatcher.
// Scope: constructs appearing in LanguageTool grammar/disambiguation patterns.
type javaRegexp struct {
	root          jNode
	caseSensitive bool
}

// jNode is one fragment of a Java-style regex AST.
type jNode interface {
	// match reports whether s[pos:] can start matching this node; returns
	// end positions in s (after consuming). Empty means no match.
	// Lookarounds return pos unchanged on success.
	match(s string, pos int) []int
}

type jLit struct {
	r             rune
	caseSensitive bool
}

func (n *jLit) match(s string, pos int) []int {
	if pos >= len(s) {
		return nil
	}
	r, size := utf8.DecodeRuneInString(s[pos:])
	if n.caseSensitive {
		if r != n.r {
			return nil
		}
	} else if !runesEqualFold(r, n.r) {
		return nil
	}
	return []int{pos + size}
}

type jAny struct{}

func (n *jAny) match(s string, pos int) []int {
	if pos >= len(s) {
		return nil
	}
	_, size := utf8.DecodeRuneInString(s[pos:])
	return []int{pos + size}
}

type jClass struct {
	// allowed runes / ranges; negate inverts
	chars         map[rune]struct{}
	ranges        [][2]rune
	negate        bool
	caseSensitive bool
}

func (n *jClass) match(s string, pos int) []int {
	if pos >= len(s) {
		return nil
	}
	r, size := utf8.DecodeRuneInString(s[pos:])
	ok := n.contains(r)
	if n.negate {
		ok = !ok
	}
	if !ok {
		return nil
	}
	return []int{pos + size}
}

func (n *jClass) contains(r rune) bool {
	if n.caseSensitive {
		if n.chars != nil {
			if _, ok := n.chars[r]; ok {
				return true
			}
		}
		for _, rg := range n.ranges {
			if r >= rg[0] && r <= rg[1] {
				return true
			}
		}
		return false
	}
	// case-insensitive: fold both sides
	rl := unicode.ToLower(r)
	ru := unicode.ToUpper(r)
	if n.chars != nil {
		for c := range n.chars {
			if unicode.ToLower(c) == rl || c == r || c == rl || c == ru {
				return true
			}
		}
	}
	for _, rg := range n.ranges {
		if (r >= rg[0] && r <= rg[1]) ||
			(rl >= unicode.ToLower(rg[0]) && rl <= unicode.ToLower(rg[1])) {
			return true
		}
	}
	return false
}

type jConcat []jNode

func (n jConcat) match(s string, pos int) []int {
	positions := []int{pos}
	for _, child := range n {
		var next []int
		seen := map[int]struct{}{}
		for _, p := range positions {
			for _, e := range child.match(s, p) {
				if _, ok := seen[e]; !ok {
					seen[e] = struct{}{}
					next = append(next, e)
				}
			}
		}
		if len(next) == 0 {
			return nil
		}
		positions = next
	}
	return positions
}

type jAlt []jNode

func (n jAlt) match(s string, pos int) []int {
	var out []int
	seen := map[int]struct{}{}
	for _, child := range n {
		for _, e := range child.match(s, pos) {
			if _, ok := seen[e]; !ok {
				seen[e] = struct{}{}
				out = append(out, e)
			}
		}
	}
	return out
}

// jRepeat is greedy * / + / ? (and {m,n} simplified).
type jRepeat struct {
	node jNode
	min  int
	max  int // -1 = unlimited
}

func (n *jRepeat) match(s string, pos int) []int {
	// Collect positions after k repetitions for k in [min, max].
	// Greedy: try longer first by returning longer ends first when used
	// as concat — we return all valid ends (capped).
	type state struct {
		pos int
		k   int
	}
	var ends []int
	seenEnd := map[int]struct{}{}
	// BFS by count
	frontier := []int{pos}
	for k := 0; ; k++ {
		if k >= n.min {
			for _, p := range frontier {
				if _, ok := seenEnd[p]; !ok {
					seenEnd[p] = struct{}{}
					ends = append(ends, p)
				}
			}
		}
		if n.max >= 0 && k >= n.max {
			break
		}
		if n.max < 0 && k > len(s)+1 {
			break
		}
		var next []int
		seenNext := map[int]struct{}{}
		for _, p := range frontier {
			for _, e := range n.node.match(s, p) {
				// progress required for unlimited
				if e == p && n.max < 0 {
					continue
				}
				if _, ok := seenNext[e]; !ok {
					seenNext[e] = struct{}{}
					next = append(next, e)
				}
			}
		}
		if len(next) == 0 {
			break
		}
		frontier = next
	}
	// Prefer longer matches first (greedy bias for alternation consumers).
	for i, j := 0, len(ends)-1; i < j; i, j = i+1, j-1 {
		ends[i], ends[j] = ends[j], ends[i]
	}
	return ends
}

type jLookahead struct {
	node    jNode
	negate  bool // (?!...) vs (?=...)
}

func (n *jLookahead) match(s string, pos int) []int {
	ends := n.node.match(s, pos)
	// Lookahead succeeds if the inner pattern can match starting at pos
	// (any length). Java: zero-width assertion.
	ok := len(ends) > 0
	if n.negate {
		ok = !ok
	}
	if ok {
		return []int{pos}
	}
	return nil
}

type jAnchor struct {
	start bool // ^ vs $
}

func (n *jAnchor) match(s string, pos int) []int {
	if n.start {
		if pos == 0 {
			return []int{pos}
		}
		return nil
	}
	if pos == len(s) {
		return []int{pos}
	}
	return nil
}

// compileJavaRegexp builds a lookaround-capable full-string matcher.
// pattern is already flag-stripped / \u-normalized when possible.
func compileJavaRegexp(pattern string, caseSensitive bool) (*javaRegexp, error) {
	p := &jParser{s: pattern, i: 0, caseSensitive: caseSensitive}
	root, err := p.parseAlt()
	if err != nil {
		return nil, err
	}
	if p.i != len(p.s) {
		return nil, fmt.Errorf("java regexp: trailing input at %d in %q", p.i, pattern)
	}
	return &javaRegexp{root: root, caseSensitive: caseSensitive}, nil
}

func (r *javaRegexp) fullMatch(s string) bool {
	if r == nil || r.root == nil {
		return false
	}
	for _, end := range r.root.match(s, 0) {
		if end == len(s) {
			return true
		}
	}
	return false
}

type jParser struct {
	s             string
	i             int
	caseSensitive bool
}

func (p *jParser) peek() byte {
	if p.i >= len(p.s) {
		return 0
	}
	return p.s[p.i]
}

func (p *jParser) parseAlt() (jNode, error) {
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}
	if p.peek() != '|' {
		return left, nil
	}
	var alts jAlt
	alts = append(alts, left)
	for p.peek() == '|' {
		p.i++
		next, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		alts = append(alts, next)
	}
	return alts, nil
}

func (p *jParser) parseConcat() (jNode, error) {
	var parts jConcat
	for {
		c := p.peek()
		if c == 0 || c == '|' || c == ')' {
			break
		}
		atom, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		atom, err = p.parseQuantifier(atom)
		if err != nil {
			return nil, err
		}
		parts = append(parts, atom)
	}
	if len(parts) == 0 {
		// empty branch (e.g. a|)
		return jConcat{}, nil
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return parts, nil
}

func (p *jParser) parseQuantifier(node jNode) (jNode, error) {
	c := p.peek()
	switch c {
	case '*':
		p.i++
		if p.peek() == '?' {
			p.i++ // non-greedy — still explore all ends
		}
		return &jRepeat{node: node, min: 0, max: -1}, nil
	case '+':
		p.i++
		if p.peek() == '?' {
			p.i++
		}
		return &jRepeat{node: node, min: 1, max: -1}, nil
	case '?':
		p.i++
		if p.peek() == '?' {
			p.i++
		}
		return &jRepeat{node: node, min: 0, max: 1}, nil
	case '{':
		// {m}, {m,}, {m,n}
		p.i++
		min, ok := p.readInt()
		if !ok {
			return nil, fmt.Errorf("java regexp: bad quantifier at %d", p.i)
		}
		max := min
		if p.peek() == ',' {
			p.i++
			if p.peek() == '}' {
				max = -1
			} else {
				var ok2 bool
				max, ok2 = p.readInt()
				if !ok2 {
					return nil, fmt.Errorf("java regexp: bad quantifier max at %d", p.i)
				}
			}
		}
		if p.peek() != '}' {
			return nil, fmt.Errorf("java regexp: expected }} at %d", p.i)
		}
		p.i++
		if p.peek() == '?' {
			p.i++
		}
		return &jRepeat{node: node, min: min, max: max}, nil
	default:
		return node, nil
	}
}

func (p *jParser) readInt() (int, bool) {
	if p.i >= len(p.s) || p.s[p.i] < '0' || p.s[p.i] > '9' {
		return 0, false
	}
	n := 0
	for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
		n = n*10 + int(p.s[p.i]-'0')
		p.i++
	}
	return n, true
}

func (p *jParser) parseAtom() (jNode, error) {
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("java regexp: unexpected end")
	}
	c := p.s[p.i]
	switch c {
	case '^':
		p.i++
		return &jAnchor{start: true}, nil
	case '$':
		p.i++
		return &jAnchor{start: false}, nil
	case '.':
		p.i++
		return &jAny{}, nil
	case '[':
		return p.parseClass()
	case '(':
		return p.parseGroup()
	case '\\':
		return p.parseEscape()
	case '*', '+', '?', ')', '|', '{', '}':
		return nil, fmt.Errorf("java regexp: unexpected %q at %d", c, p.i)
	default:
		p.i++
		r := rune(c)
		if c >= 0x80 {
			// multi-byte: re-decode
			p.i--
			r, size := utf8.DecodeRuneInString(p.s[p.i:])
			p.i += size
			return &jLit{r: r, caseSensitive: p.caseSensitive}, nil
		}
		return &jLit{r: r, caseSensitive: p.caseSensitive}, nil
	}
}

func (p *jParser) parseGroup() (jNode, error) {
	// (
	p.i++
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("java regexp: unclosed group")
	}
	if p.s[p.i] == '?' {
		p.i++
		if p.i >= len(p.s) {
			return nil, fmt.Errorf("java regexp: bad group")
		}
		switch p.s[p.i] {
		case ':': // (?:...)
			p.i++
			inner, err := p.parseAlt()
			if err != nil {
				return nil, err
			}
			if p.peek() != ')' {
				return nil, fmt.Errorf("java regexp: expected ) at %d", p.i)
			}
			p.i++
			return inner, nil
		case '!': // (?!...)
			p.i++
			inner, err := p.parseAlt()
			if err != nil {
				return nil, err
			}
			if p.peek() != ')' {
				return nil, fmt.Errorf("java regexp: expected ) at %d", p.i)
			}
			p.i++
			return &jLookahead{node: inner, negate: true}, nil
		case '=': // (?=...)
			p.i++
			inner, err := p.parseAlt()
			if err != nil {
				return nil, err
			}
			if p.peek() != ')' {
				return nil, fmt.Errorf("java regexp: expected ) at %d", p.i)
			}
			p.i++
			return &jLookahead{node: inner, negate: false}, nil
		case '<':
			// lookbehind (?<!...) / (?<=...) — rare in LT; reject clearly
			return nil, fmt.Errorf("java regexp: lookbehind not supported at %d", p.i)
		default:
			return nil, fmt.Errorf("java regexp: unsupported group (?%c at %d", p.s[p.i], p.i)
		}
	}
	// capturing (...)
	inner, err := p.parseAlt()
	if err != nil {
		return nil, err
	}
	if p.peek() != ')' {
		return nil, fmt.Errorf("java regexp: expected ) at %d", p.i)
	}
	p.i++
	return inner, nil
}

func (p *jParser) parseEscape() (jNode, error) {
	// \
	p.i++
	if p.i >= len(p.s) {
		return nil, fmt.Errorf("java regexp: trailing backslash")
	}
	c := p.s[p.i]
	p.i++
	switch c {
	case 'd':
		return &jClass{chars: nil, ranges: [][2]rune{{'0', '9'}}, caseSensitive: true}, nil
	case 'D':
		return &jClass{ranges: [][2]rune{{'0', '9'}}, negate: true, caseSensitive: true}, nil
	case 'w':
		return &jClass{
			chars:         map[rune]struct{}{'_': {}},
			ranges:        [][2]rune{{'0', '9'}, {'A', 'Z'}, {'a', 'z'}},
			caseSensitive: true,
		}, nil
	case 'W':
		return &jClass{
			chars:         map[rune]struct{}{'_': {}},
			ranges:        [][2]rune{{'0', '9'}, {'A', 'Z'}, {'a', 'z'}},
			negate:        true,
			caseSensitive: true,
		}, nil
	case 's':
		return &jClass{chars: map[rune]struct{}{
			' ': {}, '\t': {}, '\n': {}, '\r': {}, '\f': {},
		}, caseSensitive: true}, nil
	case 'S':
		return &jClass{chars: map[rune]struct{}{
			' ': {}, '\t': {}, '\n': {}, '\r': {}, '\f': {},
		}, negate: true, caseSensitive: true}, nil
	case 'n':
		return &jLit{r: '\n', caseSensitive: true}, nil
	case 't':
		return &jLit{r: '\t', caseSensitive: true}, nil
	case 'r':
		return &jLit{r: '\r', caseSensitive: true}, nil
	case 'x':
		// \x{hhhh} (Go form after normalize) or \xHH
		if p.peek() == '{' {
			p.i++
			start := p.i
			for p.i < len(p.s) && p.s[p.i] != '}' {
				p.i++
			}
			if p.i >= len(p.s) {
				return nil, fmt.Errorf("java regexp: bad \\x{}")
			}
			hex := p.s[start:p.i]
			p.i++ // }
			var r rune
			if _, err := fmt.Sscanf(hex, "%x", &r); err != nil {
				return nil, fmt.Errorf("java regexp: bad \\x{%s}", hex)
			}
			return &jLit{r: r, caseSensitive: p.caseSensitive}, nil
		}
		if p.i+1 < len(p.s) {
			hex := p.s[p.i : p.i+2]
			p.i += 2
			var r rune
			if _, err := fmt.Sscanf(hex, "%x", &r); err != nil {
				return nil, fmt.Errorf("java regexp: bad \\xHH")
			}
			return &jLit{r: r, caseSensitive: p.caseSensitive}, nil
		}
		return nil, fmt.Errorf("java regexp: bad \\x")
	default:
		// literal escaped char
		return &jLit{r: rune(c), caseSensitive: p.caseSensitive}, nil
	}
}

func (p *jParser) parseClass() (jNode, error) {
	// [
	p.i++
	negate := false
	if p.peek() == '^' {
		negate = true
		p.i++
	}
	chars := map[rune]struct{}{}
	var ranges [][2]rune
	first := true
	for p.i < len(p.s) && (p.peek() != ']' || first) {
		first = false
		r, err := p.readClassRune()
		if err != nil {
			return nil, err
		}
		if p.peek() == '-' && p.i+1 < len(p.s) && p.s[p.i+1] != ']' {
			p.i++ // -
			r2, err := p.readClassRune()
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, [2]rune{r, r2})
			continue
		}
		chars[r] = struct{}{}
	}
	if p.peek() != ']' {
		return nil, fmt.Errorf("java regexp: unclosed class at %d", p.i)
	}
	p.i++
	return &jClass{chars: chars, ranges: ranges, negate: negate, caseSensitive: p.caseSensitive}, nil
}

func (p *jParser) readClassRune() (rune, error) {
	if p.i >= len(p.s) {
		return 0, fmt.Errorf("java regexp: bad class")
	}
	if p.s[p.i] == '\\' {
		p.i++
		if p.i >= len(p.s) {
			return 0, fmt.Errorf("java regexp: trailing backslash in class")
		}
		c := p.s[p.i]
		p.i++
		switch c {
		case 'n':
			return '\n', nil
		case 't':
			return '\t', nil
		case 'r':
			return '\r', nil
		case 'x':
			// reuse escape reader by rewinding into parseEscape path
			p.i -= 2 // back to \
			node, err := p.parseEscape()
			if err != nil {
				return 0, err
			}
			if lit, ok := node.(*jLit); ok {
				return lit.r, nil
			}
			return 0, fmt.Errorf("java regexp: non-literal escape in class")
		default:
			return rune(c), nil
		}
	}
	r, size := utf8.DecodeRuneInString(p.s[p.i:])
	p.i += size
	return r, nil
}

func runesEqualFold(a, b rune) bool {
	if a == b {
		return true
	}
	return unicode.ToLower(a) == unicode.ToLower(b)
}

// needsJavaRegexp reports patterns that require the lookaround engine.
func needsJavaRegexp(pattern string) bool {
	// Lookahead / lookbehind markers not supported by RE2.
	return strings.Contains(pattern, "(?!") ||
		strings.Contains(pattern, "(?=") ||
		strings.Contains(pattern, "(?<!") ||
		strings.Contains(pattern, "(?<=")
}
