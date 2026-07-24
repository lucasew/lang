package pattern

import (
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	reEntityDecl = regexp.MustCompile(`<!ENTITY\s+([A-Za-z_][\w.-]*)\s+("([^"]*)"|'([^']*)')\s*>`)
	reEntityRef  = regexp.MustCompile(`&([A-Za-z_][\w.-]*);`)
)

// predefined XML entities left for the XML parser
var xmlPredefined = map[string]bool{
	"lt": true, "gt": true, "amp": true, "apos": true, "quot": true,
}

// OpenExpandedXML expands custom DTD entities and strips DOCTYPE (shared by grammar + disambiguation).
func OpenExpandedXML(path string) (io.Reader, error) {
	return readXMLExpandEntities(path)
}

// readXMLExpandEntities loads an LT grammar XML file, expands custom DTD entities, strips DOCTYPE.
func readXMLExpandEntities(path string) (io.Reader, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := string(b)
	if strings.HasPrefix(s, "\ufeff") {
		s = strings.TrimPrefix(s, "\ufeff")
	}

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
		// Unescape XML character references inside entity values for storage.
		entities[name] = val
	}

	// Expand nested custom entity references inside entity values.
	for pass := 0; pass < 30; pass++ {
		changed := false
		for k, v := range entities {
			nv := expandCustomRefs(v, entities)
			if nv != v {
				entities[k] = nv
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Remove DOCTYPE (entities already captured).
	if i := strings.Index(s, "<!DOCTYPE"); i >= 0 {
		if j := strings.Index(s[i:], "]>"); j >= 0 {
			s = s[:i] + s[i+j+2:]
		}
	}

	// Expand custom entities in document; leave &amp; &lt; etc. alone.
	s = expandCustomRefs(s, entities)
	return strings.NewReader(s), nil
}

func expandCustomRefs(s string, entities map[string]string) string {
	return reEntityRef.ReplaceAllStringFunc(s, func(ref string) string {
		name := ref[1 : len(ref)-1]
		if xmlPredefined[name] {
			return ref
		}
		// numeric &#...; leave for parser
		if name == "" {
			return ref
		}
		if v, ok := entities[name]; ok {
			return v
		}
		// Unknown custom entity: keep as empty comment to avoid parse fail
		// (or leave ref which would fail). Prefer empty.
		return ""
	})
}
