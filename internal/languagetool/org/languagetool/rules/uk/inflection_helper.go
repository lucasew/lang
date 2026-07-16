package uk

import "regexp"

// Inflection ports InflectionHelper.Inflection (Ukrainian gender/case/animacy).
type Inflection struct {
	Gender  string
	Case    string
	AnimTag string
}

var mfn = regexp.MustCompile(`^[mfn]$`)

var genOrder = map[string]int{
	"m": 0, "f": 1, "n": 2, "p": 3, "s": 4,
}

var vidmOrder = map[string]int{
	"v_naz": 0, "v_rod": 1, "v_dav": 2, "v_zna": 3,
	"v_oru": 4, "v_mis": 5, "v_kly": 6,
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
