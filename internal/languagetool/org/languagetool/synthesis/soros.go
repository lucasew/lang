package synthesis

import (
	"regexp"
	"strconv"
	"strings"
)

// Soros ports org.languagetool.synthesis.Soros (numbertext interpreter).
// Supports a practical subset of the Java interpreter sufficient for
// simple rewrite rules and left-zero stripping from __numbertext__.
type Soros struct {
	patterns []*regexp.Regexp
	values   []string
	begins   []bool
	ends     []bool
}

const (
	pu0 = '\uE000'
	pu1 = '\uE001'
	pu2 = '\uE002'
	pu3 = '\uE003' // pipe marker
)

// NewSoros compiles a Soros program for lang (e.g. "en").
func NewSoros(source, lang string) *Soros {
	// strip comments and normalize separators
	source = stripSorosComments(source, lang)
	if !strings.Contains(source, "__numbertext__") {
		source = "__numbertext__;" + source
	}
	source = strings.ReplaceAll(source, "__numbertext__",
		`"([a-z][-a-z]* )?0+(0|[1-9]\d*)" $1$2;`+
			// empty rule for failed separator (noop pattern not added if empty replacement-only)
			`"__noop_unused__" x`)

	s := &Soros{}
	lineRE := regexp.MustCompile(`^\s*("(?:\\.|[^"])*"|[^\s]*)\s*(.*)?\s*$`)
	for _, line := range strings.Split(source, ";") {
		line = strings.TrimSpace(line)
		if line == "" || line == `"__noop_unused__" x` {
			continue
		}
		sp := lineRE.FindStringSubmatch(line)
		if sp == nil {
			continue
		}
		pat := unquoteSoros(sp[1])
		repl := ""
		if len(sp) > 2 {
			repl = strings.TrimSpace(sp[2])
			repl = unquoteSoros(repl)
		}
		begin := strings.HasPrefix(pat, "^")
		end := strings.HasSuffix(pat, "$")
		core := strings.TrimPrefix(pat, "^")
		core = strings.TrimSuffix(core, "$")
		re, err := regexp.Compile("^" + core + "$")
		if err != nil {
			continue
		}
		s.patterns = append(s.patterns, re)
		s.values = append(s.values, repl)
		s.begins = append(s.begins, begin)
		s.ends = append(s.ends, end)
	}
	return s
}

func stripSorosComments(source, lang string) string {
	// remove # comments to end of line; keep ; as rule separator
	var b strings.Builder
	for _, line := range strings.Split(source, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			// keep country selectors roughly: #[:en:] lines handled by enabling
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(';')
		}
		b.WriteString(line)
	}
	_ = lang
	return b.String()
}

func unquoteSoros(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Run applies the Soros program to input.
func (s *Soros) Run(input string) string {
	return s.run(input, true, true)
}

func (s *Soros) run(input string, begin, end bool) string {
	// iterate until stable (zero-strip then rewrite), max few steps
	cur := input
	for step := 0; step < 8; step++ {
		matched := false
		for i, p := range s.patterns {
			if (!begin && s.begins[i]) || (!end && s.ends[i]) {
				continue
			}
			sub := p.FindStringSubmatchIndex(cur)
			if sub == nil {
				continue
			}
			cur = expandBackrefs(s.values[i], cur, sub)
			// expand $() calls
			cur = s.expandCalls(cur, begin, end)
			matched = true
			break
		}
		if !matched {
			break
		}
		// if we produced empty and had a match that is terminal
		if cur == "" {
			return ""
		}
	}
	// if still original after no match rules for digit rewrite, return ""
	// Java returns "" when no rule matches at all
	if cur == input {
		// check if any rule matched would have changed - if input never matched any, return ""
		for i, p := range s.patterns {
			if (!begin && s.begins[i]) || (!end && s.ends[i]) {
				continue
			}
			if p.MatchString(input) {
				return cur
			}
		}
		return ""
	}
	return cur
}

var callRE = regexp.MustCompile(`\$\(([^()]*)\)`)

func (s *Soros) expandCalls(str string, begin, end bool) string {
	for {
		loc := callRE.FindStringSubmatchIndex(str)
		if loc == nil {
			break
		}
		inner := str[loc[2]:loc[3]]
		// recursive run on inner
		repl := s.run(inner, begin, end)
		str = str[:loc[0]] + repl + str[loc[1]:]
	}
	return str
}

func expandBackrefs(tmpl, input string, sub []int) string {
	// $0, $1, \1 style
	var b strings.Builder
	for i := 0; i < len(tmpl); i++ {
		c := tmpl[i]
		if (c == '$' || c == '\\') && i+1 < len(tmpl) {
			j := i + 1
			if tmpl[j] == '{' {
				k := strings.IndexByte(tmpl[j+1:], '}')
				if k >= 0 {
					n, err := strconv.Atoi(tmpl[j+1 : j+1+k])
					if err == nil {
						b.WriteString(groupAt(input, sub, n))
						i = j + 1 + k
						continue
					}
				}
			}
			if tmpl[j] >= '0' && tmpl[j] <= '9' {
				n := int(tmpl[j] - '0')
				b.WriteString(groupAt(input, sub, n))
				i = j
				continue
			}
		}
		b.WriteByte(c)
	}
	return b.String()
}

func groupAt(input string, sub []int, n int) string {
	if 2*n+1 >= len(sub) || sub[2*n] < 0 {
		return ""
	}
	return input[sub[2*n]:sub[2*n+1]]
}
