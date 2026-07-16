package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// WordCoherencyData is the loaded variant map plus surface→base form mapping.
type WordCoherencyData struct {
	// WordMap maps a spelling (and expanded surfaces) to its alternatives.
	WordMap map[string]map[string]struct{}
	// ToBase maps a surface form to the uninflected form from the data file.
	ToBase map[string]string
}

// LoadWordCoherencyData ports WordCoherencyDataLoader.loadWords (+ optional inflection expansion).
func LoadWordCoherencyData(r io.Reader, path string, expandInflections bool) (*WordCoherencyData, error) {
	d := &WordCoherencyData{
		WordMap: make(map[string]map[string]struct{}),
		ToBase:  make(map[string]string),
	}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		line = strings.TrimRight(line, ";")
		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Format error in file %s, line: %s", path, line)
		}
		a, b := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if a == "" || b == "" {
			return nil, fmt.Errorf("Format error in file %s, line: %s", path, line)
		}
		if expandInflections {
			expandCoherencyPair(d, a, b)
		} else {
			addCoherencyPair(d, a, b, a, b)
		}
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

func expandCoherencyPair(d *WordCoherencyData, a, b string) {
	for _, fa := range coherencySurfaceForms(a) {
		for _, fb := range coherencySurfaceForms(b) {
			addCoherencyPair(d, fa, fb, a, b)
		}
	}
}

func coherencySurfaceForms(w string) []string {
	out := []string{w, w + "s"}
	if strings.HasSuffix(w, "e") {
		out = append(out, w+"d", w[:len(w)-1]+"ing")
	} else {
		out = append(out, w+"ed", w+"ing")
	}
	// German adjective/noun full forms (lemma stand-in without a tagger).
	for _, suf := range []string{
		"e", "er", "es", "en", "em",
		"ere", "erer", "eres", "eren", "erem",
		"ste", "ster", "stes", "sten", "stem",
	} {
		out = append(out, w+suf)
	}
	// Polish noun case endings (blef → blefu, bluff → bluffem, …).
	for _, suf := range []string{
		"u", "owi", "em", "ie", "y", "ą", "ę",
		"ów", "om", "ami", "ach", "ami",
	} {
		out = append(out, w+suf)
	}
	// Russian soft-sign nouns: нуль → нулю, ноль → ноля, …
	if strings.HasSuffix(w, "ь") && len([]rune(w)) > 1 {
		runes := []rune(w)
		stem := string(runes[:len(runes)-1])
		for _, suf := range []string{"я", "ю", "ем", "ём", "е", "и", "ей"} {
			out = append(out, stem+suf)
		}
	}
	return out
}
