package languagetool

import (
	"bufio"
	"io"
	"strings"
)

// LanguageModuleProperties ports META-INF/org/languagetool/language-module.properties.
// Key languageClasses = comma-separated fully-qualified class names.
type LanguageModuleProperties struct {
	LanguageClasses []string
}

// ParseLanguageModuleProperties reads a properties stream.
func ParseLanguageModuleProperties(r io.Reader) (LanguageModuleProperties, error) {
	var out LanguageModuleProperties
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		// key=value
		i := strings.IndexByte(line, '=')
		if i < 0 {
			i = strings.IndexByte(line, ':')
		}
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		if key != "languageClasses" {
			continue
		}
		for _, part := range strings.Split(val, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out.LanguageClasses = append(out.LanguageClasses, part)
			}
		}
	}
	return out, sc.Err()
}

// ShortClassName returns the simple class name from a FQN.
func ShortClassName(fqn string) string {
	if i := strings.LastIndexByte(fqn, '.'); i >= 0 {
		return fqn[i+1:]
	}
	return fqn
}
