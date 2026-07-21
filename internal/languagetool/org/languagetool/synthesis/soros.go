package synthesis

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Soros ports org.languagetool.synthesis.Soros (numbertext interpreter).
// Compile once; Run applies the first matching rule and expands $() recursively
// (Java single-match return — not multi-step rewrite loops).
type Soros struct {
	patterns []*regexp.Regexp
	values   []string
	begins   []bool
	ends     []bool
}

// NewSoros ports the Java Soros(String source, String lang) constructor.
func NewSoros(raw, lang string) *Soros {
	if lang == "" {
		lang = "en"
	}
	// 1) Escape \\ \" \; \# → \uE000..\uE003
	// Java m = "\\\";#" → chars \, ", ;, #
	source := translateDelim(raw, string([]byte{'\\', '"', ';', '#'}), "\uE000\uE001\uE002\uE003", `\`)

	// 2) Country-dependent lines
	lang = strings.ReplaceAll(lang, "_", "-")
	// switch off country lines, then enable requested lang
	reOff := regexp.MustCompile(`(^|[\n;])([^\n;#]*#[^\n]*\[:[^\n:\]]*:][^\n]*)`)
	source = reOff.ReplaceAllString(source, "${1}#${2}")
	reOn := regexp.MustCompile(`(^|[\n;])#([^\n;#]*#[^\n]*\[:` + regexp.QuoteMeta(lang) + `:][^\n]*)`)
	source = reOn.ReplaceAllString(source, "${1}${2}")
	// comments → ;
	reCmt := regexp.MustCompile(`(#[^\n]*)?(\n|$)`)
	source = reCmt.ReplaceAllString(source, ";")

	if !strings.Contains(source, "__numbertext__") {
		source = "__numbertext__;" + source
	}
	source = strings.ReplaceAll(source, "__numbertext__",
		`"([a-z][-a-z]* )?0+(0|[1-9]\d*)" $(\1\2);`+
			`"`+"\uE00A"+`(.*)`+"\uE00A"+`(.+)`+"\uE00A"+`(.*)" \1\2\3;`+
			`"`+"\uE00A"+`.*`+"\uE00A\uE00A"+`.*"`)

	// Java Soros: Pattern.compile("^\\s*(\"[^\"]*\"|[^\\s]*)\\s*(.*[^\\s])?\\s*$")
	// without UNICODE_CHARACTER_CLASS — ASCII \\s only.
	lineRE := regexp.MustCompile(`^[ \t\n\v\f\r]*("[^"]*"|[^ \t\n\v\f\r]*)[ \t\n\v\f\r]*(.*[^ \t\n\v\f\r])?[ \t\n\v\f\r]*$`)
	macroRE := regexp.MustCompile(`== *(.*[^ ]?) ==`)
	prefix := ""
	s := &Soros{}
	for _, part := range strings.Split(source, ";") {
		// Java: part.trim()
		part = tools.JavaStringTrim(part)
		if part == "" {
			continue
		}
		if mm := macroRE.FindStringSubmatch(part); mm != nil {
			prefix = mm[1]
			continue
		}
		sp := lineRE.FindStringSubmatch(part)
		if sp == nil {
			continue
		}
		if prefix != "" {
			pat0 := unquoteRule(sp[1])
			repl0 := ""
			if len(sp) > 2 {
				repl0 = sp[2]
			}
			caret := ""
			body := pat0
			if strings.HasPrefix(body, "^") {
				caret = "^"
				body = body[1:]
			}
			space := ""
			if body != "" {
				space = " "
			}
			part = `"` + caret + prefix + space + body + `" ` + repl0
			sp = lineRE.FindStringSubmatch(part)
			if sp == nil {
				continue
			}
		}

		pat := unquoteRule(sp[1])
		// translate(s, c.substring(1), m.substring(1), "") — restore ", ;, #
		pat = strings.NewReplacer("\uE001", `"`, "\uE002", `;`, "\uE003", `#`).Replace(pat)
		pat = strings.ReplaceAll(pat, "\uE000", `\\`)

		repl := ""
		// Java: if (sp.group(2) != null) s2 = group(2) with quote strip (pattern already ate \\s*).
		if len(sp) > 2 && sp[2] != "" {
			repl = sp[2]
			if len(repl) >= 2 && repl[0] == '"' && repl[len(repl)-1] == '"' {
				repl = repl[1 : len(repl)-1]
			}
		}
		// translate(s2, m2, c2, "\\") m2="$()|[]" — only escaped forms
		repl = translateDelim(repl, `$()|[]`, "\uE004\uE005\uE006\uE007\uE008\uE009", `\`)
		// Java optional-bracket rewrite on bare $ / [ (not c2 markers)
		repl = convertOptionalBrackets(repl)
		// translate(s2, c, m, "") — restore \, ", ;, # from pattern-side escapes in repl
		repl = strings.NewReplacer(
			"\uE000", `\`, "\uE001", `"`, "\uE002", `;`, "\uE003", `#`,
		).Replace(repl)
		// translate(s2, m2[0:4], c, "") — bare $, (, ), | → E000..E003
		repl = strings.NewReplacer("$", "\uE000", "(", "\uE001", ")", "\uE002", "|", "\uE003").Replace(repl)
		// translate(s2, c2, m2, "") — restore any escaped $()|[] from c2
		repl = strings.NewReplacer(
			"\uE004", "$", "\uE005", "(", "\uE006", ")", "\uE007", "|", "\uE008", "[", "\uE009", "]",
		).Replace(repl)
		repl = transformReplacement(repl)

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

func transformReplacement(repl string) string {
	// \uE000(\d) → \uE000\uE001$\d\uE002  ( $n → $($n) with markers )
	var out []rune
	r := []rune(repl)
	for i := 0; i < len(r); i++ {
		if r[i] == '\uE000' && i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9' {
			out = append(out, '\uE000', '\uE001', '$', r[i+1], '\uE002')
			i++
			continue
		}
		out = append(out, r[i])
	}
	repl = string(out)
	// \\(\d) → $digit
	reBD := regexp.MustCompile(`\\(\d)`)
	repl = reBD.ReplaceAllStringFunc(repl, func(m string) string {
		return "$" + m[1:]
	})
	repl = strings.ReplaceAll(repl, `\n`, "\n")
	return repl
}

func convertOptionalBrackets(s2 string) string {
	// Java (on bare $ / [ / ]):
	// ^[$](\d\d?|\([^)]+\)) with leading \[  → $(\uE00A\uE00A|$…\uE00A
	// \[([^$\[\\]*)[$](\d…) → $(\uE00A$1\uE00A$…\uE00A
	// \uE00A\]$ → |\uE00A)
	// ] → )
	// ($d|)|\ $ → $1||$
	const ea = "\uE00A"
	re1 := regexp.MustCompile(`^\[\$(\d\d?|\([^)]+\))`)
	s2 = re1.ReplaceAllString(s2, `$(`+ea+ea+`|$$${1}`+ea)
	re2 := regexp.MustCompile(`\[([^$\[\\]*)\$(\d\d?|\([^)]+\))`)
	s2 = re2.ReplaceAllString(s2, `$(`+ea+`${1}`+ea+`$$${2}`+ea)
	re3 := regexp.MustCompile(ea + `\]$`)
	s2 = re3.ReplaceAllString(s2, `|`+ea+`)`)
	s2 = strings.ReplaceAll(s2, "]", ")")
	re4 := regexp.MustCompile(`(\$\d|\))\|\$`)
	s2 = re4.ReplaceAllString(s2, `${1}||$`)
	return s2
}

func translateDelim(s, chars, chars2, delim string) string {
	cr := []rune(chars)
	c2 := []rune(chars2)
	n := len(cr)
	if len(c2) < n {
		n = len(c2)
	}
	for i := 0; i < n; i++ {
		s = strings.ReplaceAll(s, delim+string(cr[i]), string(c2[i]))
	}
	return s
}

func unquoteRule(s string) string {
	// Java: replaceFirst("^\"", "").replaceFirst("\"$","") after pattern capture (no Unicode TrimSpace).
	s = tools.JavaStringTrim(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Run ports Soros.run(input).
func (s *Soros) Run(input string) string {
	if s == nil {
		return input
	}
	return s.run(input, true, true)
}

func (s *Soros) run(input string, begin, end bool) string {
	for i, p := range s.patterns {
		if (!begin && s.begins[i]) || (!end && s.ends[i]) {
			continue
		}
		m := p.FindStringSubmatch(input)
		if m == nil {
			continue
		}
		repl := expandJavaGroups(s.values[i], m)
		repl = s.expandFuncs(repl, begin, end)
		return repl
	}
	return ""
}

// func: (?:\|?(?:\$\()+)?(\|?\$\(([^()]*)\)\|?)(?:\)+\|?)?
// with $ ( ) | as \uE000 \uE001 \uE002 \uE003
var funcRE = regexp.MustCompile(
	`(?:` + "\uE003" + `?(?:` + "\uE000\uE001" + `)+)?` +
		`(` + "\uE003" + `?` + "\uE000\uE001" + `([^` + "\uE001\uE002" + `]*)` + "\uE002" + "\uE003" + `?)` +
		`(?:` + "\uE002" + `+` + "\uE003" + `?)?`,
)

func (s *Soros) expandFuncs(str string, begin, end bool) string {
	for {
		loc := funcRE.FindStringSubmatchIndex(str)
		if loc == nil {
			break
		}
		fullStart, fullEnd := loc[2], loc[3]
		inner := str[loc[4]:loc[5]]
		g1 := str[fullStart:fullEnd]
		whole := str[loc[0]:loc[1]]
		b, e := false, false
		if strings.HasPrefix(g1, "\uE003") || strings.HasPrefix(whole, "\uE003") {
			b = true
		} else if loc[0] == 0 {
			b = begin
		}
		if strings.HasSuffix(g1, "\uE003") || strings.HasSuffix(whole, "\uE003") {
			e = true
		} else if loc[1] == len(str) {
			e = end
		}
		repl := s.run(inner, b, e)
		str = str[:fullStart] + repl + str[fullEnd:]
	}
	return str
}

func expandJavaGroups(tmpl string, groups []string) string {
	var b strings.Builder
	r := []rune(tmpl)
	for i := 0; i < len(r); i++ {
		// leave \uE000… sequences for expandFuncs (function call markers)
		if r[i] == '\uE000' || r[i] == '\uE001' || r[i] == '\uE002' || r[i] == '\uE003' {
			b.WriteRune(r[i])
			continue
		}
		if r[i] == '$' && i+1 < len(r) {
			if r[i+1] == '{' {
				j := i + 2
				for j < len(r) && r[j] >= '0' && r[j] <= '9' {
					j++
				}
				if j < len(r) && r[j] == '}' {
					n := 0
					for _, c := range r[i+2 : j] {
						n = n*10 + int(c-'0')
					}
					if n >= 0 && n < len(groups) {
						b.WriteString(groups[n])
					}
					i = j
					continue
				}
			}
			if r[i+1] >= '0' && r[i+1] <= '9' {
				n := int(r[i+1] - '0')
				if n >= 0 && n < len(groups) {
					b.WriteString(groups[n])
				}
				i++
				continue
			}
		}
		if r[i] == '\\' && i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9' {
			n := int(r[i+1] - '0')
			if n >= 0 && n < len(groups) {
				b.WriteString(groups[n])
			}
			i++
			continue
		}
		b.WriteRune(r[i])
	}
	return b.String()
}
