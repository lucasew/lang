package ga

import (
	"bufio"
	"io"
	"strings"
	"sync"
)

// DativePluralsData ports org.languagetool.rules.ga.DativePluralsData.
type DativePluralsData struct {
	Entries             []*DativePluralsEntry
	SimpleReplacements  map[string]string
	Modernisations      map[string]string
}

var (
	dativeDataOnce sync.Once
	dativeData     *DativePluralsData
)

// LoadDativePluralsData loads embedded dative-plurals.txt (or returns cached).
func LoadDativePluralsData() *DativePluralsData {
	dativeDataOnce.Do(func() {
		f, err := dativePluralsFS.Open("data/dative-plurals.txt")
		if err != nil {
			dativeData = &DativePluralsData{
				SimpleReplacements: map[string]string{},
				Modernisations:     map[string]string{},
			}
			return
		}
		defer f.Close()
		d, err := ParseDativePluralsData(f)
		if err != nil {
			dativeData = &DativePluralsData{
				SimpleReplacements: map[string]string{},
				Modernisations:     map[string]string{},
			}
			return
		}
		dativeData = d
	})
	return dativeData
}

// ParseDativePluralsData parses semicolon lines form;lemma;gender;replacement
// with optional form:modern colon pairs.
func ParseDativePluralsData(r io.Reader) (*DativePluralsData, error) {
	var entries []*DativePluralsEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 4 {
			continue
		}
		form, formModern := splitColon2(parts[0])
		lemma, lemmaModern := splitColon2(parts[1])
		gender := parts[2]
		repl, replModern := splitColon2(parts[3])
		e := NewDativePluralsEntry(form, lemma, gender, repl)
		if formModern != "" {
			e.SetModernised(formModern)
		}
		if replModern != "" {
			e.SetEquivalent(replModern)
		}
		if lemmaModern != "" {
			e.SetModernLemma(lemmaModern)
		}
		entries = append(entries, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	d := &DativePluralsData{Entries: entries}
	d.SimpleReplacements = buildSimpleReplacements(entries)
	d.Modernisations = buildModernisations(entries)
	return d, nil
}

func (d *DativePluralsData) GetSimpleReplacements() map[string]string {
	if d == nil {
		return nil
	}
	return d.SimpleReplacements
}

func (d *DativePluralsData) GetModernisations() map[string]string {
	if d == nil {
		return nil
	}
	return d.Modernisations
}

func buildSimpleReplacements(entries []*DativePluralsEntry) map[string]string {
	out := map[string]string{}
	for _, e := range entries {
		if e == nil {
			continue
		}
		std := e.GetStandard()
		out[e.Form] = std
		if e.HasModernised() {
			out[e.FormModern] = std
		}
	}
	return out
}

func buildModernisations(entries []*DativePluralsEntry) map[string]string {
	out := map[string]string{}
	for _, e := range entries {
		if e != nil && e.HasModernised() {
			out[e.Form] = e.FormModern
		}
	}
	return out
}
