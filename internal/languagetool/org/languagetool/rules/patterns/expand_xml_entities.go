package patterns

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExpandLTXMLEntities expands LanguageTool custom DTD entities and strips DOCTYPE,
// so encoding/xml can load official grammar/false-friends files.
// Predefined XML entities (&amp; &lt; …) are left for the parser.
// Without a base directory, SYSTEM includes are not resolved (use ExpandLTXMLEntitiesWithBase
// / ReadExpandedGrammarFile).
func ExpandLTXMLEntities(data []byte) []byte {
	return ExpandLTXMLEntitiesWithBase("", data)
}

// ExpandLTXMLEntitiesWithBase is ExpandLTXMLEntities plus SYSTEM .ent includes
// (Java RuleEntityResolver for systemId ending in .ent).
// baseDir is the directory of the grammar XML file (for relative SYSTEM paths).
func ExpandLTXMLEntitiesWithBase(baseDir string, data []byte) []byte {
	s := string(data)
	if strings.HasPrefix(s, "\ufeff") {
		s = strings.TrimPrefix(s, "\ufeff")
	}
	if baseDir != "" {
		s = resolveSystemEntityIncludes(s, baseDir)
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
			// Unknown entity: empty (do not invent content). Same as missing SYSTEM file.
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

// reSystemEntity matches <!ENTITY [% ]name SYSTEM "path"> (general or parameter).
var reSystemEntity = regexp.MustCompile(
	`<!ENTITY\s+(?:%\s*)?([A-Za-z_][\w.-]*)\s+SYSTEM\s+("([^"]+)"|'([^']+)')\s*>`,
)

// reParamEntityRef matches %name; in the DTD.
var reParamEntityRef = regexp.MustCompile(`%([A-Za-z_][\w.-]*);`)

// resolveSystemEntityIncludes ports RuleEntityResolver for .ent SYSTEM entities:
// load relative files and expand parameter-entity references (%name;).
func resolveSystemEntityIncludes(s, baseDir string) string {
	// Collect parameter entities loaded from SYSTEM files.
	peContent := map[string]string{}
	// Replace SYSTEM entity decls with nothing (content applied via %name; or inline for general).
	s = reSystemEntity.ReplaceAllStringFunc(s, func(decl string) string {
		m := reSystemEntity.FindStringSubmatch(decl)
		if m == nil {
			return decl
		}
		name := m[1]
		sysPath := m[3]
		if sysPath == "" {
			sysPath = m[4]
		}
		if !strings.HasSuffix(strings.ToLower(sysPath), ".ent") {
			// Non-.ent SYSTEM (e.g. file:// user rules) — leave for now (Java returns null).
			return decl
		}
		// Skip file:// absolute user paths we cannot resolve faithfully.
		if strings.HasPrefix(sysPath, "file:") {
			return "" // drop unresolved user include
		}
		content, err := readEntRelative(baseDir, sysPath)
		if err != nil || content == "" {
			// Fail-closed: remove decl; entity refs become empty later.
			return ""
		}
		// Parameter entity (% name SYSTEM) vs general entity.
		isParam := strings.Contains(decl, "%")
		if isParam {
			peContent[name] = content
			return "" // PE applied via %name;
		}
		// General entity SYSTEM: inline the file's entity decls into DOCTYPE.
		return content
	})
	// Expand %name; parameter entity references (typically once in DOCTYPE).
	if len(peContent) > 0 {
		s = reParamEntityRef.ReplaceAllStringFunc(s, func(ref string) string {
			name := ref[1 : len(ref)-1]
			if v, ok := peContent[name]; ok {
				return v
			}
			return ""
		})
		// Nested SYSTEM in included content (rare): one more pass.
		if reSystemEntity.MatchString(s) {
			s = resolveSystemEntityIncludes(s, baseDir)
		}
	}
	return s
}

// readEntRelative resolves systemId relative to baseDir (grammar file directory).
// Also tries RuleEntityResolver-style path under .../resource/ when relative fails.
func readEntRelative(baseDir, systemId string) (string, error) {
	systemId = strings.TrimSpace(systemId)
	if systemId == "" {
		return "", os.ErrNotExist
	}
	// Primary: relative to grammar directory (Java SAX resolves against document base).
	cand := filepath.Clean(filepath.Join(baseDir, systemId))
	if b, err := os.ReadFile(cand); err == nil {
		return string(b), nil
	}
	// Basename next to XML (…/resource/pt/entities/foo.ent when XML is …/resource/pt/).
	baseName := filepath.Base(systemId)
	if strings.HasSuffix(strings.ToLower(baseName), ".ent") {
		for _, try := range []string{
			filepath.Join(baseDir, "entities", baseName),
			filepath.Join(baseDir, baseName),
		} {
			if b, err := os.ReadFile(try); err == nil {
				return string(b), nil
			}
		}
	}
	// Fallbacks for vendored layouts that strip intermediate path segments.
	// Java getPathFromLTResourceFolder: after */resource/ strip ../
	if i := strings.Index(systemId, "resource/"); i >= 0 {
		rel := systemId[i+len("resource/"):]
		rel = strings.ReplaceAll(rel, "../", "")
		// Walk up from baseDir looking for resource/<rel>
		dir := baseDir
		for n := 0; n < 10; n++ {
			try := filepath.Join(dir, "resource", rel)
			if b, err := os.ReadFile(try); err == nil {
				return string(b), nil
			}
			// also inspiration layout: .../org/languagetool/resource/...
			try = filepath.Join(dir, "org", "languagetool", "resource", rel)
			if b, err := os.ReadFile(try); err == nil {
				return string(b), nil
			}
			// inspiration languagetool-language-modules/{lang}/.../resource/{lang}/entities
			if lang := firstPathSegment(rel); lang != "" {
				try = filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
					"src", "main", "resources", "org", "languagetool", "resource", rel)
				if b, err := os.ReadFile(try); err == nil {
					return string(b), nil
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", os.ErrNotExist
}

// firstPathSegment returns the first non-empty path component (lang code in resource/pt/…).
func firstPathSegment(rel string) string {
	rel = strings.Trim(filepath.ToSlash(rel), "/")
	if rel == "" {
		return ""
	}
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		return rel[:i]
	}
	return rel
}

// ReadExpandedGrammarFile reads path and expands LT entities (including .ent SYSTEM).
func ReadExpandedGrammarFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ExpandLTXMLEntitiesWithBase(filepath.Dir(path), b), nil
}
