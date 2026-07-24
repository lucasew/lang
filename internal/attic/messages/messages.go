package messages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/attic/data"
)

// Bundle is a LanguageTool MessagesBundle (.properties) key→value map.
type Bundle map[string]string

// Load loads MessagesBundle for a language family, with core English fallback.
// Mirrors LT ResourceBundle fallback: language module bundle → core MessagesBundle.
func Load(dataRoot, family string) (Bundle, error) {
	core := data.CoreResources(dataRoot)
	b := Bundle{}

	// Base (default) then English core, then language-specific overrides.
	paths := []string{
		filepath.Join(core, "org", "languagetool", "MessagesBundle.properties"),
		filepath.Join(core, "org", "languagetool", "MessagesBundle_en.properties"),
	}
	if family != "" && family != "en" {
		// Language modules often ship MessagesBundle_<lang>.properties
		mod := filepath.Join(data.LanguageModules(dataRoot), family, "src", "main", "resources", "org", "languagetool")
		// Prefer family-specific if present.
		candidates, _ := filepath.Glob(filepath.Join(mod, "MessagesBundle*.properties"))
		paths = append(paths, candidates...)
	}

	var loaded int
	for _, p := range paths {
		if err := loadPropertiesFile(p, b); err == nil {
			loaded++
		}
	}
	if loaded == 0 {
		return nil, fmt.Errorf("no MessagesBundle found under %s", core)
	}
	return b, nil
}

func (b Bundle) Get(key string) string {
	if b == nil {
		return key
	}
	if v, ok := b[key]; ok {
		return v
	}
	return key
}

func loadPropertiesFile(path string, into Bundle) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 2*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		// continuation lines are rare in these bundles; skip full Java properties complexity for v1 keys we need.
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:eq])
		// Java Properties discards whitespace after the separator.
		v := strings.TrimLeft(line[eq+1:], " \t\f")
		v = unescapeProperties(v)
		into[k] = v
	}
	return sc.Err()
}

func unescapeProperties(s string) string {
	// Handle \uXXXX and simple escapes used in LT bundles.
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		n := s[i+1]
		switch n {
		case 'u', 'U':
			if i+5 < len(s) {
				var r rune
				for _, c := range s[i+2 : i+6] {
					r <<= 4
					switch {
					case c >= '0' && c <= '9':
						r |= rune(c - '0')
					case c >= 'a' && c <= 'f':
						r |= rune(c - 'a' + 10)
					case c >= 'A' && c <= 'F':
						r |= rune(c - 'A' + 10)
					}
				}
				b.WriteRune(r)
				i += 5
				continue
			}
		case 'n':
			b.WriteByte('\n')
			i++
		case 't':
			b.WriteByte('\t')
			i++
		case 'r':
			b.WriteByte('\r')
			i++
		default:
			b.WriteByte(n)
			i++
		}
	}
	return b.String()
}
