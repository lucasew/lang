package pt

import (
	"regexp"
	"strings"
)

// enclitic endings (simplified set)
var encliticRE = regexp.MustCompile(`(?i)^(.*?)-(me|te|se|nos|vos|lhe|lhes|o|a|os|as|mo|ma|mos|mas|to|ta|tos|tas|lho|lha|lhos|lhas)$`)

// Ordinal abbreviation surfaces → POS (green)
var ordinalAbbrevs = map[string]string{
	"1.º": "AO0MS0", "2.º": "AO0MS0", "3.º": "AO0MS0",
	"1.ª": "AO0FS0", "2.ª": "AO0FS0", "3.ª": "AO0FS0",
	"1º": "AO0MS0", "2º": "AO0MS0", "1ª": "AO0FS0", "2ª": "AO0FS0",
}

// EncliticSplit splits "verb-me" style forms; returns verb, clitic, ok.
func EncliticSplit(surface string) (verb, clitic string, ok bool) {
	m := encliticRE.FindStringSubmatch(surface)
	if m == nil {
		return "", "", false
	}
	return m[1], strings.ToLower(m[2]), true
}

// OrdinalAbbrevPOS returns POS for known ordinal abbreviations.
func OrdinalAbbrevPOS(surface string) string {
	return ordinalAbbrevs[surface]
}
