package de

import (
	"bufio"
	"embed"
	"strings"
	"sync"
)

//go:embed data/addedCompound.txt
var addedCompoundFS embed.FS

var (
	addedCompoundOnce sync.Once
	// part1 lower -> set of part2 lower
	addedCompounds map[string]map[string]struct{}
)

func loadAddedCompounds() map[string]map[string]struct{} {
	addedCompoundOnce.Do(func() {
		addedCompounds = map[string]map[string]struct{}{}
		f, err := addedCompoundFS.Open("data/addedCompound.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			if line == "" {
				continue
			}
			parts := strings.Split(line, ";")
			if len(parts) != 2 {
				continue
			}
			p0 := strings.ToLower(strings.TrimSpace(parts[0]))
			p1 := strings.ToLower(strings.TrimSpace(parts[1]))
			if p0 == "" || p1 == "" {
				continue
			}
			if addedCompounds[p0] == nil {
				addedCompounds[p0] = map[string]struct{}{}
			}
			addedCompounds[p0][p1] = struct{}{}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
	})
	return addedCompounds
}

// CompoundCheckFilter ports org.languagetool.rules.de.CompoundCheckFilter.
type CompoundCheckFilter struct{}

func NewCompoundCheckFilter() *CompoundCheckFilter {
	_ = loadAddedCompounds()
	return &CompoundCheckFilter{}
}

// Accept reports whether part1+part2 is a listed compound split (keep match).
func (f *CompoundCheckFilter) Accept(part1, part2 string) bool {
	m := loadAddedCompounds()
	set := m[strings.ToLower(part1)]
	if set == nil {
		return false
	}
	_, ok := set[strings.ToLower(part2)]
	return ok
}
