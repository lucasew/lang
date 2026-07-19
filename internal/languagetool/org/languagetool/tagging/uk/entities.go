package uk

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Official /uk/entities.txt (Java CompoundTagger.numberedEntities).
var (
	entitiesOnce sync.Once
	entityRules  []entityRule
)

type entityRule struct {
	re   *regexp.Regexp
	tags []string
}

func loadEntities() {
	entitiesOnce.Do(func() {
		path := discoverUKEntities()
		if path == "" {
			return
		}
		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.Index(line, "#"); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			if line == "" {
				continue
			}
			// Java: split on space or |  → first token is regex, rest are tags
			// line.split(" |\\|")
			parts := splitEntityLine(line)
			if len(parts) < 2 {
				continue
			}
			pat := parts[0]
			// Java word.matches(key) = full match
			re, err := regexp.Compile("^(?:" + pat + ")$")
			if err != nil {
				continue
			}
			entityRules = append(entityRules, entityRule{re: re, tags: parts[1:]})
		}
	})
}

func splitEntityLine(line string) []string {
	// Split on spaces or | (Java " |\\|")
	var out []string
	var b strings.Builder
	for _, r := range line {
		if r == ' ' || r == '|' {
			if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteRune(r)
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func discoverUKEntities() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// tagging/uk → repo root (six levels)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
	p := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk/entities.txt")
	if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
		return p
	}
	// walk from cwd
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk/entities.txt")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// EntityReadings ports CompoundTagger.generateEntities from official entities.txt.
// Fail closed when file missing or no pattern matches (no invent regex for tanks).
func EntityReadings(token string) []*languagetool.AnalyzedToken {
	if token == "" {
		return nil
	}
	loadEntities()
	var out []*languagetool.AnalyzedToken
	for _, rule := range entityRules {
		if !rule.re.MatchString(token) {
			continue
		}
		for _, tag := range rule.tags {
			if strings.Contains(tag, ":nv") {
				// tag like noun:m:nv or noun:f:nv:…
				// Java: tagParts = tag.split(":"); gender = tagParts[1]
				parts := strings.Split(tag, ":")
				genders := "m"
				if len(parts) > 1 {
					genders = parts[1]
				}
				extra := ""
				if i := strings.Index(tag, ":nv"); i >= 0 {
					extra = tag[i+len(":nv"):]
					extra = strings.ReplaceAll(extra, ":np", "")
				}
				out = append(out, generateTokensForNv(token, genders, extra)...)
				if !strings.Contains(tag, ":np") && !strings.Contains(tag, ":p") {
					out = append(out, generateTokensForNv(token, "p", extra)...)
				}
			} else {
				// e.g. noninfl
				pos := tag
				if !strings.Contains(pos, ":") && pos != "" {
					// bare "noninfl" is full POS
				}
				p, l := pos, token
				out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
			}
		}
		// first matching pattern only (Java loops all matching keys; may multi-match)
		// keep all matches like Java generateEntities LinkedHashSet
	}
	return out
}
