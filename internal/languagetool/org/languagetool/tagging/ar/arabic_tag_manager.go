package ar

import "strings"

// ArabicTagManager ports org.languagetool.tagging.ar.ArabicTagManager.
type ArabicTagManager struct {
	flagPos map[string]int
}

func NewArabicTagManager() *ArabicTagManager {
	m := &ArabicTagManager{flagPos: map[string]int{}}
	m.loadHashmap()
	return m
}

func (m *ArabicTagManager) loadHashmap() {
	// noun
	m.flagPos["NOUN_WORDTYPE"] = 0
	m.flagPos["NOUN_CATEGORY"] = 1
	m.flagPos["NOUN_GENDER"] = 4
	m.flagPos["NOUN_NUMBER"] = 5
	m.flagPos["NOUN_CASE"] = 6
	m.flagPos["NOUN_INFLECT_MARK"] = 7
	m.flagPos["NOUN_CONJ"] = 9
	m.flagPos["NOUN_JAR"] = 10
	m.flagPos["NOUN_PRONOUN"] = 11
	// verb
	m.flagPos["VERB_WORDTYPE"] = 0
	m.flagPos["VERB_CATEGORY"] = 1
	m.flagPos["VERB_TRANS"] = 2
	m.flagPos["VERB_GENDER"] = 4
	m.flagPos["VERB_NUMBER"] = 5
	m.flagPos["VERB_PERSON"] = 6
	m.flagPos["VERB_INFLECT_MARK"] = 7
	m.flagPos["VERB_TENSE"] = 8
	m.flagPos["VERB_VOICE"] = 9
	m.flagPos["VERB_CASE"] = 10
	m.flagPos["VERB_CONJ"] = 12
	m.flagPos["VERB_ISTIQBAL"] = 13
	m.flagPos["VERB_PRONOUN"] = 14
	// particle
	m.flagPos["PARTICLE_WORDTYPE"] = 0
	m.flagPos["PARTICLE_CATEGORY"] = 1
	m.flagPos["PARTICLE_OPTION"] = 2
	m.flagPos["PARTICLE_CONJ"] = 8
	m.flagPos["PARTICLE_JAR"] = 9
	m.flagPos["PARTICLE_PRONOUN"] = 10
}

func (m *ArabicTagManager) flagKey(postag, flagType string) string {
	if m.IsNoun(postag) {
		return "NOUN_" + flagType
	}
	if m.IsVerb(postag) {
		return "VERB_" + flagType
	}
	if m.IsStopWord(postag) {
		return "PARTICLE_" + flagType
	}
	return ""
}

func (m *ArabicTagManager) GetFlagPos(postag, flagType string) int {
	if m == nil {
		return 0
	}
	key := m.flagKey(postag, flagType)
	if p, ok := m.flagPos[key]; ok {
		return p
	}
	return 0
}

func (m *ArabicTagManager) GetFlag(postag, flagType string) rune {
	pos := m.GetFlagPos(postag, flagType)
	runes := []rune(postag)
	if pos < len(runes) {
		return runes[pos]
	}
	return '-'
}

func (m *ArabicTagManager) SetFlag(postag, flagType string, flag rune) string {
	pos := m.GetFlagPos(postag, flagType)
	runes := []rune(postag)
	if pos < 0 || pos >= len(runes) {
		return postag
	}
	runes[pos] = flag
	return string(runes)
}

func (m *ArabicTagManager) IsStopWord(postag string) bool {
	return len(postag) > 0 && postag[0] == 'P'
}

func (m *ArabicTagManager) IsNoun(postag string) bool {
	return len(postag) > 0 && postag[0] == 'N'
}

func (m *ArabicTagManager) IsVerb(postag string) bool {
	return len(postag) > 0 && postag[0] == 'V'
}

func (m *ArabicTagManager) IsAdj(postag string) bool {
	return len(postag) >= 2 && postag[:2] == "NA"
}

func (m *ArabicTagManager) IsMasdar(postag string) bool {
	return len(postag) >= 2 && postag[:2] == "NM"
}

func (m *ArabicTagManager) IsDual(postag string) bool {
	return postag != "" && m.GetFlag(postag, "NUMBER") == '2'
}

func (m *ArabicTagManager) IsFutureTense(postag string) bool {
	return m.IsVerb(postag) && m.GetFlag(postag, "TENSE") == 'f'
}

func (m *ArabicTagManager) IsUnAttachedNoun(postag string) bool {
	return m.IsNoun(postag) && m.GetFlag(postag, "PRONOUN") != 'H' && !endsWithX(postag)
}

func (m *ArabicTagManager) IsAttached(postag string) bool {
	return (m.IsNoun(postag) || m.IsVerb(postag)) && m.GetFlag(postag, "PRONOUN") == 'H'
}

func (m *ArabicTagManager) IsDefinite(postag string) bool {
	return m.IsNoun(postag) && m.GetFlag(postag, "PRONOUN") == 'L'
}

func (m *ArabicTagManager) IsFeminin(postag string) bool {
	return m.IsNoun(postag) && m.GetFlag(postag, "GENDER") == 'F'
}

func (m *ArabicTagManager) IsMajrour(postag string) bool {
	f := m.GetFlag(postag, "CASE")
	return f == 'I' || f == '-'
}

func (m *ArabicTagManager) HasJar(postag string) bool {
	return m.IsNoun(postag) && m.GetFlag(postag, "JAR") != '-'
}

func (m *ArabicTagManager) HasPronoun(postag string) bool {
	return m.GetFlag(postag, "PRONOUN") == 'H'
}

func (m *ArabicTagManager) HasConjunction(postag string) bool {
	flag := m.GetFlag(postag, "CONJ")
	if m.IsNoun(postag) || m.IsVerb(postag) {
		return flag != '-'
	}
	if m.IsStopWord(postag) {
		return flag != 'W'
	}
	return false
}

func (m *ArabicTagManager) IsBreak(postag string) bool {
	return (m.IsStopWord(postag) && !m.HasConjunction(postag)) ||
		(m.IsNoun(postag) && !m.HasJar(postag) && !m.HasConjunction(postag)) ||
		(m.IsVerb(postag) && !m.HasConjunction(postag))
}

func (m *ArabicTagManager) SetJar(postag, jar string) string {
	if !m.IsMajrour(postag) {
		return postag
	}
	var myflag rune
	switch jar {
	case "ب", "B":
		myflag = 'B'
	case "ل", "L":
		myflag = 'L'
	case "ك", "K":
		myflag = 'K'
	case "-", "":
		myflag = '-'
	default:
		return postag
	}
	return m.SetFlag(postag, "JAR", myflag)
}

func (m *ArabicTagManager) SetDefinite(postag, flag string) string {
	if !(m.IsNoun(postag) && m.IsUnAttachedNoun(postag)) {
		return postag
	}
	var myflag rune
	switch flag {
	case "ال", "L", "لل", "D":
		myflag = 'L'
	case "-", "":
		myflag = '-'
	default:
		return postag
	}
	return m.SetFlag(postag, "PRONOUN", myflag)
}

func (m *ArabicTagManager) SetConjunction(postag, flag string) string {
	var myflag rune
	switch flag {
	case "و", "W", "ف", "F":
		myflag = 'W'
	case "-", "":
		myflag = '-'
	default:
		return postag
	}
	if m.IsNoun(postag) || m.IsVerb(postag) {
		return m.SetFlag(postag, "CONJ", myflag)
	}
	return postag
}

func (m *ArabicTagManager) SetPronoun(postag, flag string) string {
	var myflag rune
	switch flag {
	case "ه", "H":
		myflag = 'H'
	default:
		return postag
	}
	if m.IsNoun(postag) || m.IsVerb(postag) {
		return m.SetFlag(postag, "PRONOUN", myflag)
	}
	return postag
}

// GetConjunctionPrefix returns the Arabic conjunction letter for a CONJ flag.
func (m *ArabicTagManager) GetConjunctionPrefix(postag string) string {
	switch m.GetFlag(postag, "CONJ") {
	case 'F':
		return "ف"
	case 'W':
		return "و"
	default:
		return ""
	}
}

// GetJarPrefix ports ArabicTagManager.getJarPrefix.
func (m *ArabicTagManager) GetJarPrefix(postag string) string {
	if postag == "" || !m.IsNoun(postag) {
		return ""
	}
	switch m.GetFlag(postag, "JAR") {
	case 'L':
		return "ل"
	case 'K':
		return "ك"
	case 'B':
		return "ب"
	default:
		return ""
	}
}

// GetDefinitePrefix ports ArabicTagManager.getDefinitePrefix (ال / ل with jar ل).
func (m *ArabicTagManager) GetDefinitePrefix(postag string) string {
	if postag == "" {
		return ""
	}
	if m.IsNoun(postag) && m.GetFlag(postag, "PRONOUN") == 'L' {
		if m.HasJar(postag) && m.GetJarPrefix(postag) == "ل" {
			return "ل"
		}
		return "ال"
	}
	return ""
}

func (m *ArabicTagManager) UnifyPronounTag(postag string) string {
	if m.IsAttached(postag) {
		return m.SetFlag(postag, "PRONOUN", 'H')
	}
	return postag
}

func (m *ArabicTagManager) SetProcleticFlags(postag string) string {
	if postag == "" {
		return ""
	}
	if m.IsVerb(postag) {
		p := m.SetFlag(postag, "CONJ", '-')
		return m.SetFlag(p, "ISTIQBAL", '-')
	}
	if m.IsNoun(postag) {
		p := m.SetFlag(postag, "CONJ", '-')
		p = m.SetFlag(p, "JAR", '-')
		if m.IsDefinite(postag) {
			p = m.SetFlag(p, "PRONOUN", '-')
		}
		return p
	}
	if m.IsStopWord(postag) {
		p := m.SetFlag(postag, "CONJ", '-')
		return m.SetFlag(p, "JAR", '-')
	}
	return postag
}

// ModifyPosTag ports ArabicTagManager.modifyPosTag — applies "TYPE;FLAG" tags
// (e.g. CONJ;W, JAR;B, PRONOUN;D). Returns "" if any tag is incompatible (Java null).
func (m *ArabicTagManager) ModifyPosTag(postag string, tags []string) string {
	if postag == "" {
		return ""
	}
	for _, tg := range tags {
		postag = m.AddTag(postag, tg)
		if postag == "" {
			return ""
		}
	}
	return postag
}

// AddTag ports ArabicTagManager.addTag(postag, flagString) with "TYPE;FLAG".
func (m *ArabicTagManager) AddTag(postag, flagString string) string {
	if m == nil || postag == "" {
		return postag
	}
	parts := strings.Split(flagString, ";")
	flag, flagType := "", ""
	if len(parts) == 1 {
		flag = parts[0]
	} else if len(parts) >= 2 {
		flagType = parts[0]
		flag = parts[1]
	}
	return m.addTagTyped(postag, flagType, flag)
}

// addTagTyped ports ArabicTagManager.addTag(postag, flagType, flag).
func (m *ArabicTagManager) addTagTyped(postag, flagType, flag string) string {
	switch flag {
	case "W":
		return m.SetFlag(postag, "CONJ", 'W')
	case "K":
		if !m.IsNoun(postag) {
			return ""
		}
		if !m.IsMajrour(postag) {
			return ""
		}
		return m.SetFlag(postag, "JAR", 'K')
	case "B":
		if !m.IsNoun(postag) {
			return ""
		}
		if !m.IsMajrour(postag) {
			return ""
		}
		return m.SetFlag(postag, "JAR", 'B')
	case "L":
		if m.IsNoun(postag) {
			if !m.IsMajrour(postag) {
				return ""
			}
			return m.SetFlag(postag, "JAR", 'L')
		}
		// verb: istiqbal
		return m.SetFlag(postag, "ISTIQBAL", 'L')
	case "D":
		if !m.IsUnAttachedNoun(postag) {
			return ""
		}
		return m.SetFlag(postag, "PRONOUN", 'L')
	case "S":
		if !m.IsFutureTense(postag) {
			return ""
		}
		return m.SetFlag(postag, "ISTIQBAL", 'S')
	}
	// Pronoun letters (including new-style)
	if flagType == "PRONOUN" && flag != "" && flag != "D" {
		return m.SetFlag(postag, "PRONOUN", []rune(flag)[0])
	}
	return postag
}

func endsWithX(s string) bool {
	return len(s) > 0 && s[len(s)-1] == 'X'
}
