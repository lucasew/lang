package dev

import (
	"fmt"
	"regexp"
	"sort"
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

// GetEndingsRegex ports GenerateIrishWordforms.getEndingsRegex:
// longest-first join of map keys → "(.+)(end1|end2|...)$"
func GetEndingsRegex(m map[string][]string) string {
	ends := make([]string, 0, len(m))
	for e := range m {
		ends = append(ends, e)
	}
	sort.Slice(ends, func(i, j int) bool {
		if len(ends[i]) != len(ends[j]) {
			return len(ends[i]) > len(ends[j])
		}
		return ends[i] < ends[j]
	})
	return "(.+)(" + strings.Join(ends, "|") + ")$"
}

func endingsRegex() string {
	return GetEndingsRegex(nounGuesses)
}

var nounPattern = regexp.MustCompile(endingsRegex())

// IrishWordformLine is one FST-style output line: form \t lemma \t tag
type IrishWordformLine struct {
	Form, Lemma, Tag string
}

func (l IrishWordformLine) String() string {
	return l.Form + "\t" + l.Lemma + "\t" + l.Tag
}

// GetIrishFSTNounClass ports GenerateIrishWordforms.getIrishFSTNounClass.
func GetIrishFSTNounClass(ending string) string {
	switch ending {
	case "óir", "eoir", "éir", "úir", "aeir":
		return "Nm3-1"
	case "álaí", "eálaí":
		return "Nm4-4"
	default:
		return ""
	}
}

// GuessIrishFSTNounClassSimple ports GenerateIrishWordforms.guessIrishFSTNounClassSimple.
func GuessIrishFSTNounClassSimple(word string) string {
	m := nounPattern.FindStringSubmatch(word)
	if m == nil {
		return ""
	}
	return GetIrishFSTNounClass(m[2])
}

// ExtractEnWiktionaryNounTemplate ports GenerateIrishWordforms.extractEnWiktionaryNounTemplate.
func ExtractEnWiktionaryNounTemplate(tpl string) map[string]string {
	out := map[string]string{}
	if !strings.Contains(tpl, "{{") || !strings.Contains(tpl, "}}") {
		return out
	}
	start := strings.Index(tpl, "{{") + 2
	end := strings.Index(tpl[start:], "}}")
	if end < 0 {
		return out
	}
	inner := tpl[start : start+end]
	parts := strings.Split(inner, "|")
	if len(parts) == 0 || parts[0] != "ga-decl-m3" || len(parts) < 4 {
		return out
	}
	out["class"] = "m3"
	out["stem"] = parts[1]
	out["sg.nom"] = parts[2]
	out["sg.gen"] = parts[3]
	switch len(parts) {
	case 4:
		out["pl.nom"] = parts[3]
		out["pl.gen"] = parts[3]
	case 5:
		out["pl.nom"] = parts[3]
		out["pl.gen"] = parts[4]
	default: // 6+
		if len(parts) >= 6 {
			out["pl.nom"] = parts[4]
			out["pl.gen"] = parts[5]
		}
	}
	return out
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
