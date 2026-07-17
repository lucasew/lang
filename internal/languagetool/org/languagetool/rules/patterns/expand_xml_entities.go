package patterns

import (
	"os"
	"regexp"
	"strings"
)

// ExpandLTXMLEntities expands LanguageTool custom DTD entities and strips DOCTYPE,
// so encoding/xml can load official grammar/false-friends files.
// Predefined XML entities (&amp; &lt; …) are left for the parser.
func ExpandLTXMLEntities(data []byte) []byte {
	s := string(data)
	if strings.HasPrefix(s, "\ufeff") {
		s = strings.TrimPrefix(s, "\ufeff")
	}
	reEntityDecl := regexp.MustCompile(`<!ENTITY\s+([A-Za-z_][\w.-]*)\s+("([^"]*)"|'([^']*)')\s*>`)
	reEntityRef := regexp.MustCompile(`&([A-Za-z_][\w.-]*);`)
	xmlPredefined := map[string]bool{"lt": true, "gt": true, "amp": true, "apos": true, "quot": true}

	entities := map[string]string{}
	for _, m := range reEntityDecl.FindAllStringSubmatch(s, -1) {
		name := m[1]
		if xmlPredefined[name] {
			continue
		}
		val := m[3]
		if val == "" {
			val = m[4]
		}
		entities[name] = val
	}
	expand := func(in string) string {
		return reEntityRef.ReplaceAllStringFunc(in, func(ref string) string {
			name := ref[1 : len(ref)-1]
			if xmlPredefined[name] {
				return ref
			}
			if v, ok := entities[name]; ok {
				return v
			}
			return ""
		})
	}
	for pass := 0; pass < 30; pass++ {
		changed := false
		for k, v := range entities {
			nv := expand(v)
			if nv != v {
				entities[k] = nv
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	if i := strings.Index(s, "<!DOCTYPE"); i >= 0 {
		if j := strings.Index(s[i:], "]>"); j >= 0 {
			s = s[:i] + s[i+j+2:]
		} else if j := strings.Index(s[i:], ">"); j >= 0 {
			// external SYSTEM DTD without internal subset
			s = s[:i] + s[i+j+1:]
		}
	}
	// strip stylesheets that confuse some parsers
	s = regexp.MustCompile(`<\?xml-stylesheet[^?]*\?>`).ReplaceAllString(s, "")
	s = expand(s)
	return []byte(s)
}

// ReadExpandedGrammarFile reads path and expands LT entities.
func ReadExpandedGrammarFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ExpandLTXMLEntities(b), nil
}
