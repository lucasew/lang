package ltgolden

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Extract Java unit tests into ground-truth cases.
// We pull string literals from common LT test helpers so every *Test.java contributes
// executable expectations (TDD ground truth). Patterns mirror LT core/language tests.

var (
	reAssertGood = regexp.MustCompile(`assertGood\s*\(\s*"((?:\\.|[^"\\])*)"`)
	reAssertBad  = regexp.MustCompile(`assertBad\s*\(\s*"((?:\\.|[^"\\])*)"`)
	// assertEquals(0, rule.match...) or matches.size() == 0 style via check strings
	reCheckString = regexp.MustCompile(`(?:lt|langTool|languageTool|tool)\.check\s*\(\s*"((?:\\.|[^"\\])*)"`)
	reTestMethod  = regexp.MustCompile(`(?m)^\s*(?:public\s+|private\s+|protected\s+)?void\s+(test\w*)\s*\(`)
	reClassName   = regexp.MustCompile(`(?m)public\s+class\s+(\w+)`)
	rePackage     = regexp.MustCompile(`(?m)^package\s+([\w.]+)\s*;`)
)

// ExtractJavaCases ports ALL *Test.java files under dataRoot into cases.
// Every extracted assertGood/assertBad/check("...") becomes a Case (no file skipped).
func ExtractJavaCases(javaPaths []string) ([]Case, error) {
	var out []Case
	for _, p := range javaPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		out = append(out, extractJavaFile(p, string(b))...)
	}
	return out, nil
}

func extractJavaFile(path, src string) []Case {
	class := "Unknown"
	if m := reClassName.FindStringSubmatch(src); len(m) == 2 {
		class = m[1]
	}
	lang := langFromJavaPath(path)
	// Split by test methods for better attribution
	methods := reTestMethod.FindAllStringSubmatchIndex(src, -1)
	if len(methods) == 0 {
		return extractJavaRegion(path, class, "wholeFile", lang, src)
	}
	var out []Case
	for i, mi := range methods {
		name := src[mi[2]:mi[3]]
		start := mi[0]
		end := len(src)
		if i+1 < len(methods) {
			end = methods[i+1][0]
		}
		region := src[start:end]
		out = append(out, extractJavaRegion(path, class, name, lang, region)...)
	}
	return out
}

func extractJavaRegion(path, class, method, lang, region string) []Case {
	var out []Case
	add := func(text string, incorrect bool, kind string) {
		text = unescapeJava(text)
		if strings.TrimSpace(text) == "" {
			return
		}
		out = append(out, Case{
			Kind:        KindJavaUnit,
			Lang:        lang,
			RuleID:      class, // class-level; refined later if possible
			Text:        text,
			Incorrect:   incorrect,
			SourceFile:  path,
			ExampleType: kind,
			JavaClass:   class,
			JavaMethod:  method,
		})
	}
	for _, m := range reAssertGood.FindAllStringSubmatch(region, -1) {
		add(m[1], false, "assertGood")
	}
	for _, m := range reAssertBad.FindAllStringSubmatch(region, -1) {
		add(m[1], true, "assertBad")
	}
	// check() alone is ambiguous; keep as incorrect-unknown — expect *some* analysis stability
	// For ground truth TDD we treat bare check("x") as documentation cases that at least must not crash.
	// Prefer not double-count if already assertGood/Bad.
	// Still port them with Incorrect=false (smoke) when not already captured.
	seen := map[string]bool{}
	for _, c := range out {
		seen[c.Text] = true
	}
	for _, m := range reCheckString.FindAllStringSubmatch(region, -1) {
		t := unescapeJava(m[1])
		if seen[t] {
			continue
		}
		out = append(out, Case{
			Kind:        KindJavaUnit,
			Lang:        lang,
			RuleID:      class,
			Text:        t,
			Incorrect:   false, // smoke: engine must accept text
			SourceFile:  path,
			ExampleType: "check_smoke",
			JavaClass:   class,
			JavaMethod:  method,
		})
	}
	return out
}

func unescapeJava(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '"', '\\', '\'':
			b.WriteByte(s[i])
		case 'u':
			if i+4 < len(s) {
				var r rune
				ok := true
				for _, c := range s[i+1 : i+5] {
					r <<= 4
					switch {
					case c >= '0' && c <= '9':
						r |= rune(c - '0')
					case c >= 'a' && c <= 'f':
						r |= rune(c - 'a' + 10)
					case c >= 'A' && c <= 'F':
						r |= rune(c - 'A' + 10)
					default:
						ok = false
					}
				}
				if ok {
					b.WriteRune(r)
					i += 4
					continue
				}
			}
			b.WriteByte('u')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func langFromJavaPath(path string) string {
	// languagetool-language-modules/<lang>/src/test/...
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i, p := range parts {
		if p == "languagetool-language-modules" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if strings.Contains(path, "languagetool-core") {
		return "en" // core demos often English-ish; still run
	}
	return "en"
}
