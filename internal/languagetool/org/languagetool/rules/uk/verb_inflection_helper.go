package uk

import (
	"regexp"
	"strings"
)

// VerbInflection is VerbInflectionHelper.Inflection (gender/person for agreement).
type VerbInflection struct {
	Gender string // m/f/n/i/o or empty when plural-only
	Plural string // s/p/i
	Person string // 1/2/3 or empty
}

var (
	verbInflRE   = regexp.MustCompile(`:([mfnps])(:([123])?|$)`)
	nounInflRE   = regexp.MustCompile(`(?::((?:[iu]n)?anim))?:([mfnps]):(v_naz)`)
	adjInflRE    = regexp.MustCompile(`(adj|numr):([mfnps]):(v_naz)`)
	nounPersonRE = regexp.MustCompile(`:([123])`)
)

// NewVerbInflection ports VerbInflectionHelper.Inflection constructor.
func NewVerbInflection(gender, person string) VerbInflection {
	inf := VerbInflection{Person: person}
	switch gender {
	case "s", "p":
		inf.Gender = ""
		inf.Plural = gender
	case "i":
		inf.Gender = gender
		inf.Plural = gender
	default:
		inf.Gender = gender
		inf.Plural = "s"
	}
	return inf
}

// GetVerbInflections extracts verb agreement slots from POS tags.
func GetVerbInflections(posTags []string) []VerbInflection {
	var out []VerbInflection
	for _, posTag := range posTags {
		if posTag == "" || !strings.HasPrefix(posTag, "verb") {
			continue
		}
		if strings.Contains(posTag, ":inf") {
			out = append(out, NewVerbInflection("i", ""))
			continue
		}
		if strings.Contains(posTag, ":impers") {
			out = append(out, NewVerbInflection("o", ""))
			continue
		}
		m := verbInflRE.FindStringSubmatch(posTag)
		if m == nil {
			continue
		}
		gen := m[1]
		person := ""
		if len(m) > 3 {
			person = m[3]
		}
		out = append(out, NewVerbInflection(gen, person))
	}
	return out
}

// GetNounInflections extracts nominative noun gender/person from POS tags.
func GetNounInflections(posTags []string) []VerbInflection {
	var out []VerbInflection
	for _, posTag := range posTags {
		if posTag == "" {
			continue
		}
		m := nounInflRE.FindStringSubmatch(posTag)
		if m == nil {
			continue
		}
		gen := m[2]
		person := ""
		if pm := nounPersonRE.FindStringSubmatch(posTag); pm != nil {
			person = pm[1]
		}
		out = append(out, NewVerbInflection(gen, person))
	}
	return out
}

// GetAdjInflections extracts nominative adj/numr gender/person from POS tags.
func GetAdjInflections(posTags []string) []VerbInflection {
	var out []VerbInflection
	for _, posTag := range posTags {
		if posTag == "" {
			continue
		}
		m := adjInflRE.FindStringSubmatch(posTag)
		if m == nil {
			continue
		}
		gen := m[2]
		person := ""
		if pm := nounPersonRE.FindStringSubmatch(posTag); pm != nil {
			person = pm[1]
		}
		out = append(out, NewVerbInflection(gen, person))
	}
	return out
}

// VerbInflectionsOverlap reports non-empty intersection of verb and noun inflections.
func VerbInflectionsOverlap(verbTags, nounTags []string) bool {
	v := GetVerbInflections(verbTags)
	n := GetNounInflections(nounTags)
	for _, a := range v {
		for _, b := range n {
			if a.Equals(b) {
				return true
			}
		}
	}
	return false
}

// Equals compares gender/person/plural loosely (as used for overlap checks).
func (inf VerbInflection) Equals(other VerbInflection) bool {
	if inf.Person != "" && other.Person != "" && inf.Person != other.Person {
		return false
	}
	// match gender when both set
	if inf.Gender != "" && other.Gender != "" {
		if inf.Gender == other.Gender {
			return true
		}
		// i/o specials only match self
		return false
	}
	// plural-only (s/p)
	if inf.Plural != "" && other.Plural != "" {
		return inf.Plural == other.Plural
	}
	return true
}
