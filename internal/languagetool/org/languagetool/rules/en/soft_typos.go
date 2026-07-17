package en

import (
	"bufio"
	"os"
	"strings"
)

// LoadSoftTyposFile loads a TSV of wrong→right[,right2…] spell suggestions.
// Lines: typo<TAB>suggestion[,suggestion2…]  (# comments and blanks skipped).
func LoadSoftTyposFile(path string) (map[string][]string, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string][]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		// tab or first whitespace-separated pair
		var wrong, rights string
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			wrong = strings.TrimSpace(line[:i])
			rights = strings.TrimSpace(line[i+1:])
		} else {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			wrong, rights = fields[0], fields[1]
		}
		if wrong == "" || rights == "" {
			continue
		}
		var sugs []string
		for _, s := range strings.Split(rights, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				sugs = append(sugs, s)
			}
		}
		if len(sugs) == 0 {
			continue
		}
		out[wrong] = sugs
		out[strings.ToLower(wrong)] = sugs
	}
	return out, sc.Err()
}

// MergeSpellerSuggestions returns a new map with base then extra (extra wins on conflict).
func MergeSpellerSuggestions(base, extra map[string][]string) map[string][]string {
	out := map[string][]string{}
	for k, v := range base {
		out[k] = append([]string(nil), v...)
	}
	for k, v := range extra {
		out[k] = append([]string(nil), v...)
	}
	return out
}
