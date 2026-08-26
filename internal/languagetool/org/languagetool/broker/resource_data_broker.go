package broker

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// Default resource/rules directory path prefixes (Java RESOURCE_DIR / RULES_DIR).
const (
	ResourceDir = "/org/languagetool/resource"
	RulesDir    = "/org/languagetool/rules"
)

// ResourceDataBroker ports org.languagetool.broker.ResourceDataBroker for Go.
// Paths are relative under resource/rules roots (e.g. "/en/filename").
type ResourceDataBroker interface {
	ResourceExists(path string) bool
	RuleFileExists(path string) bool
	GetFromResourceDirAsStream(path string) (io.ReadCloser, error)
	GetFromResourceDirAsLines(path string) ([]string, error)
	GetFromRulesDirAsStream(path string) (io.ReadCloser, error)
	GetAsStream(path string) (io.ReadCloser, error)
	GetResourceDir() string
	GetRulesDir() string
}

// FSResourceDataBroker serves resources from an fs.FS (typically embed.FS or os.DirFS).
// ResourceRoot and RulesRoot are prefixes inside the FS (may be empty).
type FSResourceDataBroker struct {
	FS           fs.FS
	ResourceRoot string // e.g. "org/languagetool/resource"
	RulesRoot    string // e.g. "org/languagetool/rules"
}

func NewFSResourceDataBroker(fsys fs.FS, resourceRoot, rulesRoot string) *FSResourceDataBroker {
	return &FSResourceDataBroker{FS: fsys, ResourceRoot: resourceRoot, RulesRoot: rulesRoot}
}

func (b *FSResourceDataBroker) GetResourceDir() string {
	if b.ResourceRoot != "" {
		return "/" + strings.Trim(b.ResourceRoot, "/")
	}
	return ResourceDir
}

func (b *FSResourceDataBroker) GetRulesDir() string {
	if b.RulesRoot != "" {
		return "/" + strings.Trim(b.RulesRoot, "/")
	}
	return RulesDir
}

func (b *FSResourceDataBroker) join(root, path string) string {
	path = strings.TrimPrefix(path, "/")
	if root == "" {
		return path
	}
	return strings.Trim(root, "/") + "/" + path
}

func (b *FSResourceDataBroker) ResourceExists(path string) bool {
	_, err := fs.Stat(b.FS, b.join(b.ResourceRoot, path))
	return err == nil
}

func (b *FSResourceDataBroker) RuleFileExists(path string) bool {
	_, err := fs.Stat(b.FS, b.join(b.RulesRoot, path))
	return err == nil
}

func (b *FSResourceDataBroker) open(full string) (io.ReadCloser, error) {
	f, err := b.FS.Open(full)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s: %w", full, err)
	}
	return f, nil
}

func (b *FSResourceDataBroker) GetFromResourceDirAsStream(path string) (io.ReadCloser, error) {
	return b.open(b.join(b.ResourceRoot, path))
}

func (b *FSResourceDataBroker) GetFromRulesDirAsStream(path string) (io.ReadCloser, error) {
	return b.open(b.join(b.RulesRoot, path))
}

func (b *FSResourceDataBroker) GetAsStream(path string) (io.ReadCloser, error) {
	return b.open(strings.TrimPrefix(path, "/"))
}

func (b *FSResourceDataBroker) GetFromResourceDirAsLines(path string) ([]string, error) {
	rc, err := b.GetFromResourceDirAsStream(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return readLines(rc)
}

func readLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	// large lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// MapResourceDataBroker is an in-memory broker for tests (path → content).
type MapResourceDataBroker struct {
	Resource map[string]string
	Rules    map[string]string
	Other    map[string]string
}

func NewMapResourceDataBroker() *MapResourceDataBroker {
	return &MapResourceDataBroker{
		Resource: map[string]string{},
		Rules:    map[string]string{},
		Other:    map[string]string{},
	}
}

func (m *MapResourceDataBroker) GetResourceDir() string { return ResourceDir }
func (m *MapResourceDataBroker) GetRulesDir() string    { return RulesDir }

func normPath(path string) string { return strings.TrimPrefix(path, "/") }

func (m *MapResourceDataBroker) ResourceExists(path string) bool {
	_, ok := m.Resource[normPath(path)]
	return ok
}

func (m *MapResourceDataBroker) RuleFileExists(path string) bool {
	_, ok := m.Rules[normPath(path)]
	return ok
}

func (m *MapResourceDataBroker) GetFromResourceDirAsStream(path string) (io.ReadCloser, error) {
	s, ok := m.Resource[normPath(path)]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", path)
	}
	return io.NopCloser(strings.NewReader(s)), nil
}

func (m *MapResourceDataBroker) GetFromRulesDirAsStream(path string) (io.ReadCloser, error) {
	s, ok := m.Rules[normPath(path)]
	if !ok {
		return nil, fmt.Errorf("rules resource not found: %s", path)
	}
	return io.NopCloser(strings.NewReader(s)), nil
}

func (m *MapResourceDataBroker) GetAsStream(path string) (io.ReadCloser, error) {
	s, ok := m.Other[normPath(path)]
	if !ok {
		// try resource then rules
		if s, ok = m.Resource[normPath(path)]; ok {
			return io.NopCloser(strings.NewReader(s)), nil
		}
		if s, ok = m.Rules[normPath(path)]; ok {
			return io.NopCloser(strings.NewReader(s)), nil
		}
		return nil, fmt.Errorf("path not found: %s", path)
	}
	return io.NopCloser(strings.NewReader(s)), nil
}

func (m *MapResourceDataBroker) GetFromResourceDirAsLines(path string) ([]string, error) {
	rc, err := m.GetFromResourceDirAsStream(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return readLines(rc)
}
