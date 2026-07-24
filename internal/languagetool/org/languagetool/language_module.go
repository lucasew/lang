package languagetool

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LanguageModuleProperties ports META-INF/org/languagetool/language-module.properties.
// Key languageClasses = comma-separated fully-qualified class names.
type LanguageModuleProperties struct {
	LanguageClasses []string
}

// languageClassesCommaSplit ports Languages.java: classNames.split("\\s*,\\s*")
// (Pattern without UNICODE_CHARACTER_CLASS → ASCII whitespace around commas).
var languageClassesCommaSplit = regexp.MustCompile(`[ \t\n\v\f\r]*,[ \t\n\v\f\r]*`)

// ParseLanguageModuleProperties reads a properties stream.
func ParseLanguageModuleProperties(r io.Reader) (LanguageModuleProperties, error) {
	var out LanguageModuleProperties
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		// Java Properties line handling is more complex; trim empty/# via String.trim.
		line := tools.JavaStringTrim(sc.Text())
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
		key := tools.JavaStringTrim(line[:i])
		val := tools.JavaStringTrim(line[i+1:])
		if key != "languageClasses" {
			continue
		}
		// Java: classNames.split("\\s*,\\s*")
		for _, part := range languageClassesCommaSplit.Split(val, -1) {
			part = tools.JavaStringTrim(part)
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
