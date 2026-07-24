package uk

import (
	"regexp"
	"strings"
)

// Inflection ports InflectionHelper.Inflection (Ukrainian gender/case/animacy).
type Inflection struct {
	Gender  string
	Case    string
	AnimTag string
}

var mfn = regexp.MustCompile(`^[mfn]$`)

// genOrder ports InflectionHelper.GEN_ORDER (Java sort / formatInflections).
var genOrder = map[string]int{
	"m": 0, "f": 1, "n": 3, "s": 4, "p": 5, "i": 6, "o": 7,
}

// vidmOrder ports InflectionHelper.VIDM_ORDER.
var vidmOrder = map[string]int{
	"v_naz": 10, "v_rod": 20, "v_dav": 30, "v_zna": 40,
	"v_oru": 50, "v_mis": 60, "v_kly": 70,
}

// Equals ports Inflection.equals with soft gender matching (s ≈ m/f/n).
func (inf Inflection) Equals(other Inflection) bool {
	if !genderEquals(inf.Gender, other.Gender) {
		return false
	}
	if inf.Case != other.Case {
		return false
	}
	if inf.AnimTag == "" || other.AnimTag == "" || !inf.animMatters() || !other.isAnimalSensitive() {
		return true
	}
	return inf.AnimTag == other.AnimTag
}

// EqualsIgnoreGender ports equalsIgnoreGender.
func (inf Inflection) EqualsIgnoreGender(other Inflection) bool {
	if inf.Case != other.Case {
		return false
	}
	if inf.AnimTag == "" || other.AnimTag == "" || !inf.animMatters() {
		return true
	}
	return inf.AnimTag == other.AnimTag
}

func genderEquals(g1, g2 string) bool {
	if g1 == g2 {
		return true
	}
	if g1 == "s" && mfn.MatchString(g2) {
		return true
	}
	if g2 == "s" && mfn.MatchString(g1) {
		return true
	}
	return false
}

func (inf Inflection) animMatters() bool {
	return inf.AnimTag != "" && inf.AnimTag != "unanim" && inf.Case == "v_zna" && inf.isAnimalSensitive()
}

func (inf Inflection) isAnimalSensitive() bool {
	return inf.Gender == "m" || inf.Gender == "p"
}

func (inf Inflection) String() string {
	s := ":" + inf.Gender + ":" + inf.Case
	if inf.animMatters() {
		s += "_" + inf.AnimTag
	}
	return s
}

// CompareTo orders by gender then case (unknown genders sort last).
func (inf Inflection) CompareTo(o Inflection) int {
	g1, ok1 := genOrder[inf.Gender]
	g2, ok2 := genOrder[o.Gender]
	if !ok1 {
		g1 = 99
	}
	if !ok2 {
		g2 = 99
	}
	if g1 != g2 {
		return g1 - g2
	}
	c1, ok1 := vidmOrder[inf.Case]
	c2, ok2 := vidmOrder[o.Case]
	if !ok1 {
		c1 = 99
	}
	if !ok2 {
		c2 = 99
	}
	return c1 - c2
}

// --- POS tag extraction (TokenAgreementAdjNounRule patterns) ---

var (
	adjInflectionRE  = regexp.MustCompile(`:([mfnp]):(v_...)(:r(in)?anim)?`)
	nounInflectionRE = regexp.MustCompile(`((?:[iu]n)?anim):([mfnps]):(v_...)`)
)

// GetAdjInflectionsFromTags extracts adj/numr gender/case/anim from POS tags.
func GetAdjInflectionsFromTags(posTags []string, postagStart string) []Inflection {
	if postagStart == "" {
		postagStart = "adj"
	}
	var out []Inflection
	seen := map[string]struct{}{}
	for _, posTag := range posTags {
		if posTag == "" || !strings.HasPrefix(posTag, postagStart) {
			continue
		}
		m := adjInflectionRE.FindStringSubmatch(posTag)
		if m == nil {
			continue
		}
		gen, vidm := m[1], m[2]
		anim := ""
		if m[3] != "" {
			anim = m[3][2:] // strip :r
		}
		inf := Inflection{Gender: gen, Case: vidm, AnimTag: anim}
		key := inf.String() + "|" + anim
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, inf)
	}
	return out
}

// GetAdjCaseInflections is GetAdjInflectionsFromTags with "adj" prefix
// (case/gender agreement; distinct from verb_inflection_helper.GetAdjInflections).
func GetAdjCaseInflections(posTags []string) []Inflection {
	return GetAdjInflectionsFromTags(posTags, "adj")
}

// GetNumrCaseInflections extracts numr case/gender inflections.
func GetNumrCaseInflections(posTags []string) []Inflection {
	return GetAdjInflectionsFromTags(posTags, "numr")
}

// GetNounCaseInflections is GetNounInflectionsFromTags with no ignore filter.
func GetNounCaseInflections(posTags []string) []Inflection {
	return GetNounInflectionsFromTags(posTags, nil)
}

// GetNounInflectionsFromTags extracts noun gender/case/anim from POS tags.
// ignoreRE, if non-nil, skips tags that match (Java ignoreTag.matcher.find()).
func GetNounInflectionsFromTags(posTags []string, ignoreRE *regexp.Regexp) []Inflection {
	var out []Inflection
	seen := map[string]struct{}{}
	for _, posTag := range posTags {
		if posTag == "" {
			continue
		}
		// Java: ignoreTag.matcher(posTag2).find() — substring find, not matches()
		if ignoreRE != nil && ignoreRE.MatchString(posTag) {
			continue
		}
		m := nounInflectionRE.FindStringSubmatch(posTag)
		if m == nil {
			continue
		}
		anim, gen, vidm := m[1], m[2], m[3]
		inf := Inflection{Gender: gen, Case: vidm, AnimTag: anim}
		key := inf.String() + "|" + anim
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, inf)
	}
	return out
}

// InflectionsIntersect reports whether master and slave share an agreeing inflection.
func InflectionsIntersect(master, slave []Inflection) bool {
	for _, m := range master {
		for _, s := range slave {
			if m.Equals(s) {
				return true
			}
		}
	}
	return false
}

// InflectionsIntersectIgnoreGender ports TokenAgreementAdjNounExceptionHelper.hasOverlapIgnoreGender
// (optional gender filters on master/slave; empty string = no filter).
func InflectionsIntersectIgnoreGender(master, slave []Inflection, masterGenderFilter, slaveGenderFilter string) bool {
	for _, m := range master {
		if masterGenderFilter != "" && !strings.EqualFold(m.Gender, masterGenderFilter) {
			continue
		}
		for _, s := range slave {
			if slaveGenderFilter != "" && !strings.EqualFold(s.Gender, slaveGenderFilter) {
				continue
			}
			if m.EqualsIgnoreGender(s) {
				return true
			}
		}
	}
	return false
}

// GenderMatches ports TokenAgreementAdjNounExceptionHelper.genderMatches
// (same gender; optional case filters on master/slave).
func GenderMatches(master, slave []Inflection, masterCaseFilter, slaveCaseFilter string) bool {
	for _, m := range master {
		if masterCaseFilter != "" && m.Case != masterCaseFilter {
			continue
		}
		for _, s := range slave {
			if slaveCaseFilter != "" && s.Case != slaveCaseFilter {
				continue
			}
			if s.Gender == m.Gender {
				return true
			}
		}
	}
	return false
}
