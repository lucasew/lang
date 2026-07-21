package rules

import (
	"bufio"
	"fmt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"io"
	"strings"
)

// WordCoherencyData is the loaded variant map plus surface→base form mapping.
type WordCoherencyData struct {
	// WordMap maps a spelling to its alternatives (exact file forms only).
	WordMap map[string]map[string]struct{}
	// ToBase maps a form to the uninflected form from the data file.
	ToBase map[string]string
}

// LoadWordCoherencyData ports WordCoherencyDataLoader.loadWords.
// expandInflections is ignored: Java never invents inflection suffixes (exact pairs only).
// The parameter remains for call-site compatibility; always load exact semicolon pairs.
func LoadWordCoherencyData(r io.Reader, path string, expandInflections bool) (*WordCoherencyData, error) {
	_ = expandInflections
	d := &WordCoherencyData{
		WordMap: make(map[string]map[string]struct{}),
		ToBase:  make(map[string]string),
	}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		if line == "" {
			continue
		}
		line = strings.TrimRight(line, ";")
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Format error in file %s, line: %s", path, line)
		}
		a, b := tools.JavaStringTrim(parts[0]), tools.JavaStringTrim(parts[1])
		if a == "" || b == "" {
			return nil, fmt.Errorf("Format error in file %s, line: %s", path, line)
		}
		addCoherencyPair(d, a, b, a, b)
	}
	return d, sc.Err()
}

// LoadWordCoherencyMap is a convenience wrapper returning only the word map.
func LoadWordCoherencyMap(r io.Reader, path string, expandInflections bool) (map[string]map[string]struct{}, error) {
	d, err := LoadWordCoherencyData(r, path, expandInflections)
	if err != nil {
		return nil, err
	}
	return d.WordMap, nil
}

func addCoherencyPair(d *WordCoherencyData, a, b, baseA, baseB string) {
	// Keys are lowercased for case-insensitive lookup (DE data uses capital nouns).
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	if a == b {
		return
	}
	if d.WordMap[a] == nil {
		d.WordMap[a] = make(map[string]struct{})
	}
	d.WordMap[a][b] = struct{}{}
	if d.WordMap[b] == nil {
		d.WordMap[b] = make(map[string]struct{})
	}
	d.WordMap[b][a] = struct{}{}
	if _, ok := d.ToBase[a]; !ok {
		d.ToBase[a] = baseA
	}
	if _, ok := d.ToBase[b]; !ok {
		d.ToBase[b] = baseB
	}
}

// WordCoherencyDataLoader ports org.languagetool.rules.WordCoherencyDataLoader.
// ExpandInflections is deprecated and ignored (Java has no invent expansion).
type WordCoherencyDataLoader struct {
	ExpandInflections bool
}

func (l WordCoherencyDataLoader) LoadWords(r io.Reader, path string) (*WordCoherencyData, error) {
	return LoadWordCoherencyData(r, path, false)
}
