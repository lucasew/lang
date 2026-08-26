package language

import "strings"

// GermanPrepareLineForSpeller ports German.prepareLineForSpeller.
// Twin of rules/de.PrepareLineForSpeller (kept here for Language API surface).
func GermanPrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	formTag := strings.Split(parts[0], "/")
	if len(formTag) == 0 {
		return []string{""}
	}
	form := formTag[0]
	results := []string{form}
	tag := ""
	if len(formTag) == 2 {
		tag = formTag[1]
	}
	if strings.Contains(tag, "E") {
		results = append(results, form+"e")
	}
	if strings.Contains(tag, "S") {
		results = append(results, form+"s")
	}
	if strings.Contains(tag, "N") {
		results = append(results, form+"n")
	}
	return results
}

// HasMinMatchesRules ports German.hasMinMatchesRules (true).
func (v GermanVariant) HasMinMatchesRules() bool { return true }
