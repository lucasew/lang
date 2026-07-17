package dev

import (
	"fmt"
	"regexp"
	"strings"
)

// nounGuesses ports GenerateIrishWordforms.nounGuesses (ending → gender + forms).
var nounGuesses = map[string][]string{
	"óir":   {"m3", "óir", "óra", "óirí", "óirí"},
	"eoir":  {"m3", "eoir", "eora", "eoirí", "eoirí"},
	"éir":   {"m3", "éir", "éara", "éirí", "éirí"},
	"úir":   {"m3", "úir", "úra", "úirí", "úirí"},
	"aeir":  {"m3", "aeir", "aera", "aeirí", "aeirí"},
	"álaí":  {"m4", "álaí", "álaí", "álaithe", "álaithe"},
	"eálaí": {"m4", "eálaí", "eálaí", "eálaithe", "eálaithe"},
}

var baseforms = []string{"sg.nom", "sg.gen", "pl.nom", "pl.gen"}

func endingsRegex() string {
	// longest first to prefer multi-char endings
	ends := make([]string, 0, len(nounGuesses))
	for e := range nounGuesses {
		ends = append(ends, regexp.QuoteMeta(e))
	}
	// sort by length desc
	for i := 0; i < len(ends); i++ {
		for j := i + 1; j < len(ends); j++ {
			if len(ends[j]) > len(ends[i]) {
				ends[i], ends[j] = ends[j], ends[i]
			}
		}
	}
	return `^(.*)(` + strings.Join(ends, "|") + `)$`
}

var nounPattern = regexp.MustCompile(endingsRegex())

// IrishWordformLine is one FST-style output line: form \t lemma \t tag
type IrishWordformLine struct {
	Form, Lemma, Tag string
}

func (l IrishWordformLine) String() string {
	return l.Form + "\t" + l.Lemma + "\t" + l.Tag
}

// ExpandIrishNounFromGuess ports writeFromGuess noun expansion (simplified tags).
func ExpandIrishNounFromGuess(word string) []IrishWordformLine {
	m := nounPattern.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	stem, ending := m[1], m[2]
	ends, ok := nounGuesses[ending]
	if !ok || len(ends) < 5 {
		return nil
	}
	// ends[0]=gender, then 4 form endings for baseforms
	gender := ends[0]
	var out []IrishWordformLine
	for i, bf := range baseforms {
		formEnding := ends[i+1]
		form := stem + formEnding
		tag := fmt.Sprintf("Noun:%s:%s", gender, bf)
		out = append(out, IrishWordformLine{Form: form, Lemma: word, Tag: tag})
	}
	return out
}
