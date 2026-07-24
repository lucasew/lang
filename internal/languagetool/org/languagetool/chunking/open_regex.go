package chunking

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// OpenRegex sequence matcher ports edu.washington.cs.knowitall.regex.RegularExpression
// for GermanChunker patterns: <token>, quantifiers *+?{n,m}, groups (), floating |.
// Matching uses recursive backtracking (same results for LT chunk patterns).

// SeqMatch is one match [Start, End).
type SeqMatch struct {
	Start, End int
}

// SeqRegex is a compiled OpenRegex over ChunkTaggedToken sequences.
type SeqRegex struct {
	nodes []seqNode
}

type seqNode interface {
	lookingAt(tokens []ChunkTaggedToken, pos int) []int // candidate end positions (longest first)
	minLen() int
}

type atomNode struct {
	pred func(ChunkTaggedToken) bool
	src  string
}

func (n *atomNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	if pos < 0 || pos >= len(tokens) {
		return nil
	}
	if n.pred(tokens[pos]) {
		return []int{pos + 1}
	}
	return nil
}
func (n *atomNode) minLen() int { return 1 }

type starNode struct{ inner seqNode }

func (n *starNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	return quantLookingAt(n.inner, tokens, pos, 0, -1)
}
func (n *starNode) minLen() int { return 0 }

type plusNode struct{ inner seqNode }

func (n *plusNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	return quantLookingAt(n.inner, tokens, pos, 1, -1)
}
func (n *plusNode) minLen() int { return 1 }

type optNode struct{ inner seqNode }

func (n *optNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	return quantLookingAt(n.inner, tokens, pos, 0, 1)
}
func (n *optNode) minLen() int { return 0 }

type rangeNode struct {
	inner      seqNode
	minN, maxN int
}

func (n *rangeNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	return quantLookingAt(n.inner, tokens, pos, n.minN, n.maxN)
}
func (n *rangeNode) minLen() int { return n.minN }

type groupNode struct{ nodes []seqNode }

func (n *groupNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	return matchSeq(n.nodes, tokens, pos)
}
func (n *groupNode) minLen() int {
	s := 0
	for _, x := range n.nodes {
		s += x.minLen()
	}
	return s
}

type orNode struct{ a, b seqNode }

func (n *orNode) lookingAt(tokens []ChunkTaggedToken, pos int) []int {
	var out []int
	out = append(out, n.a.lookingAt(tokens, pos)...)
	out = append(out, n.b.lookingAt(tokens, pos)...)
	sortIntsDesc(out)
	return uniqInts(out)
}
func (n *orNode) minLen() int {
	ma, mb := n.a.minLen(), n.b.minLen()
	if ma < mb {
		return ma
	}
	return mb
}

func quantLookingAt(inner seqNode, tokens []ChunkTaggedToken, pos, minN, maxN int) []int {
	var ends []int
	var dfs func(p, n int)
	dfs = func(p, n int) {
		if maxN >= 0 && n > maxN {
			return
		}
		if n >= minN {
			ends = append(ends, p)
		}
		if maxN >= 0 && n == maxN {
			return
		}
		for _, np := range inner.lookingAt(tokens, p) {
			if np <= p {
				continue // avoid zero-width loop
			}
			dfs(np, n+1)
		}
	}
	dfs(pos, 0)
	sortIntsDesc(ends)
	return uniqInts(ends)
}

func matchSeq(nodes []seqNode, tokens []ChunkTaggedToken, pos int) []int {
	if len(nodes) == 0 {
		return []int{pos}
	}
	var ends []int
	var dfs func(i, p int)
	dfs = func(i, p int) {
		if i == len(nodes) {
			ends = append(ends, p)
			return
		}
		for _, np := range nodes[i].lookingAt(tokens, p) {
			dfs(i+1, np)
		}
	}
	dfs(0, pos)
	sortIntsDesc(ends)
	return uniqInts(ends)
}

// CompileOpenRegex compiles pattern; factory builds predicates for text inside <...>.
func CompileOpenRegex(pattern string, factory func(string) func(ChunkTaggedToken) bool) *SeqRegex {
	return &SeqRegex{nodes: tokenizeOpenRegex(pattern, factory)}
}

// LookingAt returns end index of longest match at start, or -1.
func (r *SeqRegex) LookingAt(tokens []ChunkTaggedToken, start int) int {
	if r == nil {
		return -1
	}
	ends := matchSeq(r.nodes, tokens, start)
	if len(ends) == 0 {
		return -1
	}
	return ends[0]
}

// Find returns first match at or after start, or nil.
func (r *SeqRegex) Find(tokens []ChunkTaggedToken, start int) *SeqMatch {
	if r == nil {
		return nil
	}
	minL := r.minMatchingLength()
	for i := start; i <= len(tokens)-minL; i++ {
		end := r.LookingAt(tokens, i)
		if end >= i {
			if end == i {
				// empty match allowed by optional-only patterns
				return &SeqMatch{Start: i, End: end}
			}
			return &SeqMatch{Start: i, End: end}
		}
	}
	return nil
}

// FindAll ports RegularExpression.findAll — non-overlapping, non-empty matches.
func (r *SeqRegex) FindAll(tokens []ChunkTaggedToken) []SeqMatch {
	if r == nil {
		return nil
	}
	var results []SeqMatch
	start := 0
	for {
		m := r.Find(tokens, start)
		if m == nil {
			break
		}
		start = m.End
		if m.End > m.Start {
			results = append(results, *m)
		} else {
			// empty: advance to avoid infinite loop (Java still sets start=end)
			start = m.Start + 1
			if start > len(tokens) {
				break
			}
		}
	}
	return results
}

func (r *SeqRegex) minMatchingLength() int {
	if r == nil {
		return 0
	}
	s := 0
	for _, n := range r.nodes {
		s += n.minLen()
	}
	return s
}

func tokenizeOpenRegex(s string, factory func(string) func(ChunkTaggedToken) bool) []seqNode {
	var expressions []seqNode
	// Java OpenNLP / OpenRegex whitespace is Pattern \\s (ASCII without UNICODE_CHARACTER_CLASS).
	ws := regexp.MustCompile(`[ \t\n\v\f\r]+`)
	unary := regexp.MustCompile(`[*?+]`)
	minMax := regexp.MustCompile(`\{(\d+),(\d+)\}`)
	start := 0
	pendingOR := false

	for start < len(s) {
		if m := ws.FindStringIndex(s[start:]); m != nil && m[0] == 0 {
			start += m[1]
			continue
		}
		if start >= len(s) {
			break
		}
		c := s[start]
		if c == '(' || c == '<' || c == '[' {
			if c == '(' {
				end := indexOfClose(s, start, '(', ')')
				if end < 0 {
					panic(fmt.Sprintf("openregex: unclosed parenthesis at %d", start))
				}
				group := s[start+1 : end]
				start = end + 1
				named := regexp.MustCompile(`^<(\w*)>:(.*)$`)
				unnamed := regexp.MustCompile(`^\?:(.*)$`)
				var groupNodes []seqNode
				if m := named.FindStringSubmatch(group); m != nil {
					groupNodes = tokenizeOpenRegex(m[2], factory)
				} else if m := unnamed.FindStringSubmatch(group); m != nil {
					groupNodes = tokenizeOpenRegex(m[1], factory)
				} else {
					groupNodes = tokenizeOpenRegex(group, factory)
				}
				expressions = append(expressions, &groupNode{nodes: groupNodes})
			} else {
				token := readAngleToken(s[start:])
				inside := token[1 : len(token)-1]
				expressions = append(expressions, &atomNode{pred: factory(inside), src: inside})
				start += len(token)
			}
			if pendingOR {
				pendingOR = false
				if len(expressions) < 2 {
					panic("openregex: OR needs 2 operands")
				}
				b := expressions[len(expressions)-1]
				a := expressions[len(expressions)-2]
				expressions = expressions[:len(expressions)-2]
				expressions = append(expressions, &orNode{a: a, b: b})
			}
			continue
		}
		if m := unary.FindStringIndex(s[start:]); m != nil && m[0] == 0 {
			op := s[start]
			start++
			if len(expressions) == 0 {
				panic("openregex: unary without expression")
			}
			base := expressions[len(expressions)-1]
			expressions = expressions[:len(expressions)-1]
			switch op {
			case '*':
				expressions = append(expressions, &starNode{inner: base})
			case '+':
				expressions = append(expressions, &plusNode{inner: base})
			case '?':
				expressions = append(expressions, &optNode{inner: base})
			}
			continue
		}
		if m := minMax.FindStringSubmatchIndex(s[start:]); m != nil && m[0] == 0 {
			minN, _ := strconv.Atoi(s[start+m[2] : start+m[3]])
			maxN, _ := strconv.Atoi(s[start+m[4] : start+m[5]])
			start += m[1]
			if len(expressions) == 0 {
				panic("openregex: {n,m} without expression")
			}
			base := expressions[len(expressions)-1]
			expressions = expressions[:len(expressions)-1]
			expressions = append(expressions, &rangeNode{inner: base, minN: minN, maxN: maxN})
			continue
		}
		if c == '|' {
			pendingOR = true
			start++
			continue
		}
		panic(fmt.Sprintf("openregex: unexpected %q at %d in %q", c, start, s))
	}
	return expressions
}

func readAngleToken(remaining string) string {
	if remaining == "" {
		panic("openregex: empty token")
	}
	var end int
	switch remaining[0] {
	case '<':
		end = indexOfClose(remaining, 0, '<', '>')
	case '[':
		end = indexOfClose(remaining, 0, '[', ']')
	default:
		panic("openregex: token must start with < or [")
	}
	if end < 0 {
		panic("openregex: non-matching brackets")
	}
	return remaining[:end+1]
}

func indexOfClose(s string, start int, open, close byte) int {
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '\'' {
			i++
			for i < len(s) && s[i] != '\'' {
				i++
			}
			continue
		}
		if s[i] == open {
			depth++
		} else if s[i] == close {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func uniqInts(a []int) []int {
	if len(a) == 0 {
		return a
	}
	seen := map[int]bool{}
	var out []int
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func sortIntsDesc(a []int) {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[j] > a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

// NewChunkTokenFactory returns the OpenRegex factory used by GermanChunker.
func NewChunkTokenFactory(caseSensitive bool) func(string) func(ChunkTaggedToken) bool {
	f := NewTokenExpressionFactory(caseSensitive)
	return func(expr string) func(ChunkTaggedToken) bool {
		return f.Create(expr).Apply
	}
}

// ExpandGermanChunkSyntax ports GermanChunker SYNTAX_EXPANSION before compile.
func ExpandGermanChunkSyntax(expr string) string {
	expr = strings.ReplaceAll(expr, "<NP>", "<chunk=B-NP> <chunk=I-NP>*")
	expr = strings.ReplaceAll(expr, "&prozent;", "Prozent|Kilo|Kilogramm|Gramm|Euro|Pfund")
	return expr
}

// ExpandRussianChunkSyntax ports RussianChunker SYNTAX_EXPANSION before compile.
func ExpandRussianChunkSyntax(expr string) string {
	expr = strings.ReplaceAll(expr, "<NP>", "<chunk=B-NP> <chunk=I-NP>*")
	expr = strings.ReplaceAll(expr, "<VP>", "<chunk=B-VP> <chunk=I-VP>*")
	expr = strings.ReplaceAll(expr, "<ADJP>", "<chunk=B-ADJP> <chunk=I-ADJP>*")
	expr = strings.ReplaceAll(expr, "<DPT>", "<chunk=B-DPT> <chunk=I-DPT>*")
	return expr
}
