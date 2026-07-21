package gui

import (
	"bufio"
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Configuration ports org.languagetool.gui.Configuration surface:
// disabled/enabled rule IDs per language, properties file I/O.
// Full GUI prefs (n-gram paths, server, etc.) deferred.
type Configuration struct {
	// Dir and FileName form the config path (Java: parentFile + name).
	Dir      string
	FileName string
	// LangCode optional language short code for per-language sections ("" = global).
	LangCode string

	DisabledRuleIDs map[string]struct{}
	EnabledRuleIDs  map[string]struct{}
}

// NewConfiguration loads or creates a config for langCode (may be empty).
func NewConfiguration(dir, fileName, langCode string) (*Configuration, error) {
	c := &Configuration{
		Dir:             dir,
		FileName:        fileName,
		LangCode:        langCode,
		DisabledRuleIDs: map[string]struct{}{},
		EnabledRuleIDs:  map[string]struct{}{},
	}
	path := filepath.Join(dir, fileName)
	if _, err := os.Stat(path); err == nil {
		if err := c.load(path); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Configuration) GetDisabledRuleIDs() map[string]struct{} {
	return cloneSet(c.DisabledRuleIDs)
}

func (c *Configuration) GetEnabledRuleIDs() map[string]struct{} {
	return cloneSet(c.EnabledRuleIDs)
}

func (c *Configuration) SetDisabledRuleIDs(ids []string) {
	c.DisabledRuleIDs = map[string]struct{}{}
	for _, id := range ids {
		c.DisabledRuleIDs[id] = struct{}{}
	}
}

func (c *Configuration) SetEnabledRuleIDs(ids []string) {
	c.EnabledRuleIDs = map[string]struct{}{}
	for _, id := range ids {
		c.EnabledRuleIDs[id] = struct{}{}
	}
}

// SaveConfiguration writes the current language section to the config file,
// preserving other language sections.
func (c *Configuration) SaveConfiguration() error {
	if c == nil {
		return fmt.Errorf("nil configuration")
	}
	path := filepath.Join(c.Dir, c.FileName)
	// load all sections if file exists
	sections := map[string]*sectionData{}
	if _, err := os.Stat(path); err == nil {
		all, err := loadAllSections(path)
		if err != nil {
			return err
		}
		sections = all
	}
	key := c.LangCode
	if key == "" {
		key = "_"
	}
	sections[key] = &sectionData{
		Disabled: keysOf(c.DisabledRuleIDs),
		Enabled:  keysOf(c.EnabledRuleIDs),
	}
	return writeAllSections(path, sections)
}

func (c *Configuration) load(path string) error {
	all, err := loadAllSections(path)
	if err != nil {
		return err
	}
	key := c.LangCode
	if key == "" {
		key = "_"
	}
	sec, ok := all[key]
	if !ok {
		return nil
	}
	c.SetDisabledRuleIDs(sec.Disabled)
	c.SetEnabledRuleIDs(sec.Enabled)
	return nil
}

type sectionData struct {
	Disabled []string
	Enabled  []string
}

func loadAllSections(path string) (map[string]*sectionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]*sectionData{}
	sc := bufio.NewScanner(f)
	var cur string
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			cur = line[1 : len(line)-1]
			if _, ok := out[cur]; !ok {
				out[cur] = &sectionData{}
			}
			continue
		}
		if cur == "" {
			cur = "_"
			if _, ok := out[cur]; !ok {
				out[cur] = &sectionData{}
			}
		}
		sec := out[cur]
		if strings.HasPrefix(line, "disabled=") {
			sec.Disabled = splitCSV(strings.TrimPrefix(line, "disabled="))
		} else if strings.HasPrefix(line, "enabled=") {
			sec.Enabled = splitCSV(strings.TrimPrefix(line, "enabled="))
		}
	}
	return out, sc.Err()
}

func writeAllSections(path string, sections map[string]*sectionData) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	keys := make([]string, 0, len(sections))
	for k := range sections {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sec := sections[k]
		if _, err := fmt.Fprintf(f, "[%s]\n", k); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "disabled=%s\n", strings.Join(sec.Disabled, ",")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "enabled=%s\n", strings.Join(sec.Enabled, ",")); err != nil {
			return err
		}
	}
	return nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = tools.JavaStringTrim(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func keysOf(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func cloneSet(m map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k := range m {
		out[k] = struct{}{}
	}
	return out
}
