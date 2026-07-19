package uk

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Official /uk/dash_prefixes.txt and dash_prefixes_invalid.txt (Java CompoundTagger).
var (
	dashPrefOnce         sync.Once
	dashPrefixes         map[string]string // prefix → extra tag (may be empty or "alt")
	dashPrefixesInvalid  map[string]struct{}
	noDashPrefixes       map[string]struct{}
)

func loadDashPrefixResources() {
	dashPrefOnce.Do(func() {
		dashPrefixes = map[string]string{}
		dashPrefixesInvalid = map[string]struct{}{}
		noDashPrefixes = map[string]struct{}{}
		if p := discoverUKResource("dash_prefixes.txt"); p != "" {
			loadDashPrefixesMap(p)
		}
		if p := discoverUKResource("dash_prefixes_invalid.txt"); p != "" {
			loadSetInto(p, dashPrefixesInvalid)
			for k := range dashPrefixesInvalid {
				noDashPrefixes[k] = struct{}{}
			}
		}
		// Java: noDashPrefixes2019 = dash prefix keys whose value contains "alt"
		for k, v := range dashPrefixes {
			if strings.Contains(v, "alt") {
				noDashPrefixes[k] = struct{}{}
			}
		}
		// Java: too many false positives
		delete(noDashPrefixes, "мілі")
		delete(noDashPrefixes, "поп")
		delete(noDashPrefixes, "прес")
	})
}

// IsDashPrefixInvalid reports membership in official dash_prefixes_invalid.txt.
func IsDashPrefixInvalid(left string) bool {
	loadDashPrefixResources()
	if _, ok := dashPrefixesInvalid[left]; ok {
		return true
	}
	low := strings.ToLower(left)
	if _, ok := dashPrefixesInvalid[low]; ok {
		return true
	}
	for k := range dashPrefixesInvalid {
		if strings.EqualFold(k, left) {
			return true
		}
	}
	return false
}

func discoverUKResource(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
		p := filepath.Join(root, "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk", name)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk", name)
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

// loadDashPrefixesMap ports ExtraDictionaryLoader.loadMap (key [value]).
func loadDashPrefixesMap(path string) {
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
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		key := parts[0]
		val := ""
		if len(parts) > 1 {
			val = parts[1]
		}
		dashPrefixes[key] = val
	}
}

func loadSetInto(path string, set map[string]struct{}) {
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
		// first field only
		fields := strings.Fields(line)
		if len(fields) > 0 {
			set[fields[0]] = struct{}{}
		}
	}
}

// IsDashPrefix reports membership in official dash_prefixes.txt (case-insensitive).
func IsDashPrefix(left string) bool {
	loadDashPrefixResources()
	if _, ok := dashPrefixes[left]; ok {
		return true
	}
	low := strings.ToLower(left)
	if _, ok := dashPrefixes[low]; ok {
		return true
	}
	for k := range dashPrefixes {
		if strings.EqualFold(k, left) {
			return true
		}
	}
	return false
}

// DashPrefixExtraTag returns the map value for a dash prefix (may be empty).
func DashPrefixExtraTag(left string) (string, bool) {
	loadDashPrefixResources()
	if v, ok := dashPrefixes[left]; ok {
		return v, true
	}
	if v, ok := dashPrefixes[strings.ToLower(left)]; ok {
		return v, true
	}
	for k, v := range dashPrefixes {
		if strings.EqualFold(k, left) {
			return v, true
		}
	}
	return "", false
}

// NoDashPrefixList returns sorted-stable iteration order for no-dash prefix tries.
// Java noDashPrefixes = invalid file ∪ dash keys with "alt", minus false-positive removals.
func NoDashPrefixList() []string {
	loadDashPrefixResources()
	// longest-first so "напів" beats "пів"
	out := make([]string, 0, len(noDashPrefixes))
	for k := range noDashPrefixes {
		out = append(out, k)
	}
	// simple length-desc sort
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if len([]rune(out[j])) > len([]rune(out[i])) ||
				(len([]rune(out[j])) == len([]rune(out[i])) && out[j] < out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
