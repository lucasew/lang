package ga

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"strings"
	"sync"

	taggingga "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official rules-dir resource Java loads:
// JLanguageTool.getDataBroker().getFromRulesDirAsStream("/ga/dative-plurals.txt")
//
//go:embed data/dative-plurals.txt
var dativePluralsFS embed.FS

// DativePluralsFilePath is the official rules-dir path (Java loadWords argument).
const DativePluralsFilePath = "/ga/dative-plurals.txt"

// DativePluralsData ports org.languagetool.rules.ga.DativePluralsData.
type DativePluralsData struct {
	Entries            []*DativePluralsEntry
	SimpleReplacements map[string]string
	Modernisations     map[string]string
}

var (
	dativeDataOnce sync.Once
	dativeData     *DativePluralsData
)

// LoadDativePluralsData ports static initialization of DativePluralsData
// (loadWords + getModernisations + buildSimpleReplacements). Panics on load/parse
// failure to match Java static-init throw semantics.
func LoadDativePluralsData() *DativePluralsData {
	dativeDataOnce.Do(func() {
		f, err := dativePluralsFS.Open("data/dative-plurals.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		d, err := ParseDativePluralsData(f)
		if err != nil {
			panic(err)
		}
		dativeData = d
	})
	return dativeData
}

// GetSimpleReplacements ports DativePluralsData.getSimpleReplacements() (static).
func GetSimpleReplacements() map[string]string {
	return LoadDativePluralsData().GetSimpleReplacements()
}

// GetModernisations ports DativePluralsData.getModernisations() (static).
func GetModernisations() map[string]string {
	return LoadDativePluralsData().GetModernisations()
}

// ParseDativePluralsData ports DativePluralsData.loadWords (Scanner over rules-dir stream).
// Wrong field counts and multi-colon form/lemma/repl throw like Java RuntimeException.
func ParseDativePluralsData(r io.Reader) (*DativePluralsData, error) {
	var entries []*DativePluralsEntry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 4 {
			// Java: throw new RuntimeException("Incorrect number of fields: " + line);
			return nil, fmt.Errorf("Incorrect number of fields: %s", line)
		}
		form, formModern, err := splitColonModern(parts[0], "Form has more than 1 modern form:", line)
		if err != nil {
			return nil, err
		}
		lemma, lemmaModern, err := splitColonModern(parts[1], "Lemma has more than 1 modern form:", line)
		if err != nil {
			return nil, err
		}
		gender := parts[2]
		repl, replModern, err := splitColonModern(parts[3], "Replacement has more than 1 modern form:", line)
		if err != nil {
			return nil, err
		}
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

// splitColonModern ports optional "base:modern" field parsing.
// When the field contains ':', Java requires exactly two segments (split length == 2).
func splitColonModern(field, errPrefix, line string) (base, modern string, err error) {
	if !strings.Contains(field, ":") {
		return field, "", nil
	}
	parts := strings.Split(field, ":")
	if len(parts) != 2 {
		// Java message has no space before the line for form/replacement; lemma message same style.
		return "", "", fmt.Errorf("%s%s", errPrefix, line)
	}
	return parts[0], parts[1], nil
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

// buildSimpleReplacements ports DativePluralsData.buildSimpleReplacements.
// For each entry: form→standard, lenite/eclipse/h-prothesis variants, and the same
// block for modern form when present (values always mutated from getStandard()).
func buildSimpleReplacements(entries []*DativePluralsEntry) map[string]string {
	out := map[string]string{}
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		out[entry.GetForm()] = entry.GetStandard()
		lenitedForm := taggingga.Lenite(entry.GetForm())
		lenitedRepl := taggingga.Lenite(entry.GetStandard())
		if lenitedForm != entry.GetForm() {
			out[lenitedForm] = lenitedRepl
		}
		eclipsedForm := taggingga.Eclipse(entry.GetForm())
		eclipsedRepl := taggingga.Eclipse(entry.GetStandard())
		if eclipsedForm != entry.GetForm() {
			out[eclipsedForm] = eclipsedRepl
		}
		// h-prothesis
		if formRunes := []rune(entry.GetForm()); len(formRunes) > 0 && taggingga.IsVowel(formRunes[0]) {
			out["h"+entry.GetForm()] = "h" + entry.GetStandard()
			out["h-"+entry.GetForm()] = "h" + entry.GetStandard()
		}
		if entry.HasModernised() {
			out[entry.GetModern()] = entry.GetStandard()
			lenitedForm = taggingga.Lenite(entry.GetModern())
			lenitedRepl = taggingga.Lenite(entry.GetStandard())
			if lenitedForm != entry.GetModern() {
				out[lenitedForm] = lenitedRepl
			}
			eclipsedForm = taggingga.Eclipse(entry.GetModern())
			eclipsedRepl = taggingga.Eclipse(entry.GetStandard())
			if eclipsedForm != entry.GetModern() {
				out[eclipsedForm] = eclipsedRepl
			}
			// h-prothesis
			if modRunes := []rune(entry.GetModern()); len(modRunes) > 0 && taggingga.IsVowel(modRunes[0]) {
				out["h"+entry.GetModern()] = "h" + entry.GetStandard()
				out["h-"+entry.GetModern()] = "h" + entry.GetStandard()
			}
		}
	}
	return out
}

// buildModernisations ports DativePluralsData.getModernisations.
// Only entries with a modernised dative form; lenite/eclipse use getReplacement()
// (not getStandard / getModern); h-prothesis maps to modern.
func buildModernisations(entries []*DativePluralsEntry) map[string]string {
	out := map[string]string{}
	for _, entry := range entries {
		if entry == nil || !entry.HasModernised() {
			continue
		}
		out[entry.GetForm()] = entry.GetModern()
		lenitedForm := taggingga.Lenite(entry.GetForm())
		lenitedRepl := taggingga.Lenite(entry.GetReplacement())
		if lenitedForm != entry.GetForm() {
			out[lenitedForm] = lenitedRepl
		}
		eclipsedForm := taggingga.Eclipse(entry.GetForm())
		eclipsedRepl := taggingga.Eclipse(entry.GetReplacement())
		if eclipsedForm != entry.GetForm() {
			out[eclipsedForm] = eclipsedRepl
		}
		// h-prothesis
		if formRunes := []rune(entry.GetForm()); len(formRunes) > 0 && taggingga.IsVowel(formRunes[0]) {
			out["h"+entry.GetForm()] = "h" + entry.GetModern()
			out["h-"+entry.GetForm()] = "h" + entry.GetModern()
		}
	}
	return out
}

// listifyDative ports DativePluralStandardReplaceRule.listify:
// Map form → single-element list of the replacement string.
func listifyDative(singles map[string]string) map[string][]string {
	out := make(map[string][]string, len(singles))
	for k, v := range singles {
		out[k] = []string{v}
	}
	return out
}
