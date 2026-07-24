package data

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Language describes a discovered LanguageTool language / variant.
type Language struct {
	// Code is short code with country/variant when known (e.g. en-US, pt-BR, de).
	Code string
	// Family is the module directory name (e.g. en, pt).
	Family string
	// Name is a human label when known.
	Name string
	// ModuleDir absolute path to languagetool-language-modules/<family>
	ModuleDir string
	// JavaClass is the LT language class simple or FQCN from language-module.properties.
	JavaClass string
}

var (
	reShortCodeReturn = regexp.MustCompile(`return\s+"([a-z]{2,3}(?:-[A-Za-z0-9]+)*)"\s*;`)
	reLangShortConst  = regexp.MustCompile(`LANGUAGE_SHORT_CODE\s*=\s*"([a-z]{2,3}(?:-[A-Za-z0-9]+)*)"`)
	reGetName         = regexp.MustCompile(`public\s+String\s+getName\s*\(\s*\)[\s\S]*?return\s+"([^"]+)"`)
	reClassName       = regexp.MustCompile(`public\s+class\s+(\w+)`)
	reCountries       = regexp.MustCompile(`getCountries\s*\(\s*\)\s*\{[\s\S]*?return\s+new\s+String\s*\[\s*\]\s*\{([^}]*)\}`)
)

// DiscoverLanguages walks official language modules under dataRoot.
func DiscoverLanguages(dataRoot string) ([]Language, error) {
	root := LanguageModules(dataRoot)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("list language modules: %w", err)
	}

	var out []Language
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "all" {
			continue
		}
		family := e.Name()
		modDir := filepath.Join(root, family)
		langs, err := discoverModule(family, modDir)
		if err != nil {
			out = append(out, Language{
				Code:      family,
				Family:    family,
				Name:      family,
				ModuleDir: modDir,
			})
			continue
		}
		out = append(out, langs...)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func discoverModule(family, modDir string) ([]Language, error) {
	props := filepath.Join(modDir, "src", "main", "resources", "META-INF", "org", "languagetool", "language-module.properties")
	b, err := os.ReadFile(props)
	if err != nil {
		return nil, err
	}
	classes := parseLanguageClasses(string(b))
	if len(classes) == 0 {
		return nil, fmt.Errorf("no languageClasses in %s", props)
	}

	javaRoot := filepath.Join(modDir, "src", "main", "java")
	bySimple := indexJavaLanguageFiles(javaRoot)

	var out []Language
	for _, fqcn := range classes {
		simple := fqcn
		if i := strings.LastIndex(fqcn, "."); i >= 0 {
			simple = fqcn[i+1:]
		}
		path, ok := bySimple[simple]
		if !ok {
			out = append(out, Language{
				Code:      family,
				Family:    family,
				Name:      simple,
				ModuleDir: modDir,
				JavaClass: fqcn,
			})
			continue
		}
		code, name := inspectLanguageJava(path, family)
		out = append(out, Language{
			Code:      code,
			Family:    family,
			Name:      name,
			ModuleDir: modDir,
			JavaClass: fqcn,
		})
	}
	return out, nil
}

func parseLanguageClasses(content string) []string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "languageClasses=") {
			raw := strings.TrimPrefix(line, "languageClasses=")
			parts := strings.Split(raw, ",")
			var out []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out
		}
	}
	return nil
}

func indexJavaLanguageFiles(javaRoot string) map[string]string {
	out := map[string]string{}
	_ = filepath.WalkDir(javaRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".java") {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".java")
		out[base] = path
		return nil
	})
	return out
}

func inspectLanguageJava(path, family string) (code, name string) {
	code = family
	name = family
	f, err := os.Open(path)
	if err != nil {
		return code, name
	}
	defer f.Close()

	var content strings.Builder
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		content.WriteString(sc.Text())
		content.WriteByte('\n')
	}
	s := content.String()

	if m := reLangShortConst.FindStringSubmatch(s); len(m) == 2 {
		code = m[1]
	} else {
		base := reShortCodeInGetShortCode(s)
		if base == "" {
			base = family
		}
		if countries := parseCountries(s); len(countries) == 1 && countries[0] != "" {
			code = base + "-" + countries[0]
		} else if strings.Contains(s, "LANGUAGE_SHORT_CODE") {
			code = base
		} else {
			code = base
		}
	}

	if m := reGetName.FindStringSubmatch(s); len(m) == 2 {
		name = m[1]
	} else if m := reClassName.FindStringSubmatch(s); len(m) == 2 {
		name = m[1]
	}
	return code, name
}

func parseCountries(s string) []string {
	m := reCountries.FindStringSubmatch(s)
	if len(m) < 2 {
		return nil
	}
	raw := m[1]
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func reShortCodeInGetShortCode(s string) string {
	idx := strings.Index(s, "getShortCode()")
	if idx < 0 {
		idx = strings.Index(s, "getShortCode ()")
	}
	if idx < 0 {
		return ""
	}
	window := s[idx:]
	if len(window) > 400 {
		window = window[:400]
	}
	if m := reShortCodeReturn.FindStringSubmatch(window); len(m) == 2 {
		return m[1]
	}
	return ""
}

// Lookup finds a language by code (exact, then family fallback).
func Lookup(langs []Language, code string) (Language, bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Language{}, false
	}
	for _, l := range langs {
		if strings.EqualFold(l.Code, code) {
			return l, true
		}
	}
	for _, l := range langs {
		if strings.EqualFold(l.Family, code) {
			return l, true
		}
	}
	want := strings.ToLower(strings.ReplaceAll(code, "_", "-"))
	for _, l := range langs {
		if strings.ToLower(strings.ReplaceAll(l.Code, "_", "-")) == want {
			return l, true
		}
	}
	return Language{}, false
}
