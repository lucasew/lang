package ca

import (
	"regexp"
	"strings"
)

// PronomsFeblesHelper ports org.languagetool.rules.ca.PronomsFeblesHelper.

// PronounPosition selects a column in the weak-pronoun table.
type PronounPosition int

const (
	PronounDavant PronounPosition = iota
	PronounDavantApos
	PronounDarrere
	PronounDarrereApos
	PronounDarrereNogionetNoapos
	PronounDarreAposNogionetNoapos
	PronounNormalized
	pronounPosCount
)

var pronomsFebles = []string{
	"el",
	"l'",
	"-lo",
	"'l",
	"lo",
	"l",
	"el",
	"els el",
	"els l'",
	"-los-el",
	"'ls-el",
	"losel",
	"lsel",
	"els el",
	"els els",
	"els els",
	"-los-els",
	"'ls-els",
	"losels",
	"lsels",
	"els els",
	"els en",
	"els n'",
	"-los-en",
	"'ls-en",
	"losen",
	"lsen",
	"els en",
	"els hi",
	"els hi",
	"-los-hi",
	"'ls-hi",
	"loshi",
	"lshi",
	"els hi",
	"els ho",
	"els ho",
	"-los-ho",
	"'ls-ho",
	"losho",
	"lsho",
	"els ho",
	"els la",
	"els l'",
	"-los-la",
	"'ls-la",
	"losla",
	"lsla",
	"els la",
	"els les",
	"els les",
	"-los-les",
	"'ls-les",
	"losles",
	"lsles",
	"els les",
	"els",
	"els",
	"-los",
	"'ls",
	"los",
	"ls",
	"els",
	"em",
	"m'",
	"-me",
	"'m",
	"me",
	"m",
	"em",
	"en",
	"n'",
	"-ne",
	"'n",
	"ne",
	"n",
	"en",
	"ens el",
	"ens l'",
	"-nos-el",
	"'ns-el",
	"nosel",
	"nsel",
	"ens el",
	"ens els",
	"ens els",
	"-nos-els",
	"'ns-els",
	"nosels",
	"nsels",
	"ens els",
	"ens en",
	"ens n'",
	"-nos-en",
	"'ns-en",
	"nosen",
	"nsen",
	"ens en",
	"ens hi",
	"ens hi",
	"-nos-hi",
	"'ns-hi",
	"noshi",
	"nshi",
	"ens hi",
	"ens ho",
	"ens ho",
	"-nos-ho",
	"'ns-ho",
	"nosho",
	"nsho",
	"ens ho",
	"ens la",
	"ens l'",
	"-nos-la",
	"'ns-la",
	"nosla",
	"nsla",
	"ens la",
	"ens les",
	"ens les",
	"-nos-les",
	"'ns-les",
	"nosles",
	"nsles",
	"ens les",
	"ens li",
	"ens li",
	"-nos-li",
	"'ns-li",
	"nosli",
	"nsli",
	"ens li",
	"ens",
	"ens",
	"-nos",
	"'ns",
	"nos",
	"ns",
	"ens",
	"es",
	"s'",
	"-se",
	"'s",
	"se",
	"s",
	"es",
	"et",
	"t'",
	"-te",
	"'t",
	"te",
	"t",
	"et",
	"hi",
	"hi",
	"-hi",
	"-hi",
	"hi",
	"hi",
	"hi",
	"ho",
	"ho",
	"-ho",
	"-ho",
	"ho",
	"ho",
	"ho",
	"l'en",
	"el n'",
	"-l'en",
	"-l'en",
	"len",
	"len",
	"el en",
	"l'hi",
	"l'hi",
	"-l'hi",
	"-l'hi",
	"lhi",
	"lhi",
	"el hi",
	"la hi",
	"la hi",
	"-la-hi",
	"-la-hi",
	"lahi",
	"lahi",
	"la hi",
	"la",
	"l'",
	"-la",
	"-la",
	"la",
	"la",
	"la",
	"la'n",
	"la n'",
	"-la'n",
	"-la'n",
	"lan",
	"lan",
	"la en",
	"les en",
	"les n'",
	"-les-en",
	"-les-en",
	"lesen",
	"lesen",
	"les en",
	"les hi",
	"les hi",
	"-les-hi",
	"-les-hi",
	"leshi",
	"leshi",
	"les hi",
	"les",
	"les",
	"-les",
	"-les",
	"les",
	"les",
	"les",
	"li hi",
	"li hi",
	"-li-hi",
	"-li-hi",
	"lihi",
	"lihi",
	"li hi",
	"li ho",
	"li ho",
	"-li-ho",
	"-li-ho",
	"liho",
	"liho",
	"li ho",
	"li la",
	"li l'",
	"-li-la",
	"-li-la",
	"lila",
	"lila",
	"li la",
	"li les",
	"li les",
	"-li-les",
	"-li-les",
	"liles",
	"liles",
	"li les",
	"li",
	"li",
	"-li",
	"-li",
	"li",
	"li",
	"li",
	"li'l",
	"li l'",
	"-li'l",
	"-li'l",
	"lil",
	"lil",
	"li el",
	"li'ls",
	"li'ls",
	"-li'ls",
	"-li'ls",
	"lils",
	"lils",
	"li els",
	"li'n",
	"li n'",
	"-li'n",
	"-li'n",
	"lin",
	"lin",
	"li en",
	"m'hi",
	"m'hi",
	"-m'hi",
	"-m'hi",
	"mhi",
	"mhi",
	"em hi",
	"m'ho",
	"m'ho",
	"-m'ho",
	"-m'ho",
	"mho",
	"mho",
	"em ho",
	"me la",
	"me l'",
	"-me-la",
	"-me-la",
	"mela",
	"mela",
	"em la",
	"me les",
	"me les",
	"-me-les",
	"-me-les",
	"meles",
	"meles",
	"em les",
	"me li",
	"me li",
	"-me-li",
	"-me-li",
	"meli",
	"meli",
	"em li",
	"me'l",
	"me l'",
	"-me'l",
	"-me'l",
	"mel",
	"mel",
	"em el",
	"me'ls",
	"me'ls",
	"-me'ls",
	"-me'ls",
	"mels",
	"mels",
	"em els",
	"me'n",
	"me n'",
	"-me'n",
	"-me'n",
	"men",
	"men",
	"em en",
	"n'hi",
	"n'hi",
	"-n'hi",
	"-n'hi",
	"nhi",
	"nhi",
	"en hi",
	"s'hi",
	"s'hi",
	"-s'hi",
	"-s'hi",
	"shi",
	"shi",
	"es hi",
	"s'ho",
	"s'ho",
	"-s'ho",
	"-s'ho",
	"sho",
	"sho",
	"es ho",
	"se la",
	"se l'",
	"-se-la",
	"-se-la",
	"sela",
	"sela",
	"es la",
	"se les",
	"se les",
	"-se-les",
	"-se-les",
	"seles",
	"seles",
	"es les",
	"se li",
	"se li",
	"-se-li",
	"-se-li",
	"seli",
	"seli",
	"es li",
	"se us",
	"se us",
	"-se-us",
	"-se-us",
	"seus",
	"seus",
	"es us",
	"se vos",
	"se vos",
	"-se-vos",
	"-se-vos",
	"sevos",
	"sevos",
	"es vos",
	"se'l",
	"se l'",
	"-se'l",
	"-se'l",
	"sel",
	"sel",
	"es el",
	"se'ls",
	"se'ls",
	"-se'ls",
	"-se'ls",
	"sels",
	"sels",
	"es els",
	"se'm",
	"se m'",
	"-se'm",
	"-se'm",
	"sem",
	"sem",
	"es em",
	"se'n",
	"se n'",
	"-se'n",
	"-se'n",
	"sen",
	"sen",
	"es en",
	"se'ns",
	"se'ns",
	"-se'ns",
	"-se'ns",
	"sens",
	"sens",
	"es ens",
	"se't",
	"se t'",
	"-se't",
	"-se't",
	"set",
	"set",
	"es et",
	"t'hi",
	"t'hi",
	"-t'hi",
	"-t'hi",
	"thi",
	"thi",
	"et hi",
	"t'ho",
	"t'ho",
	"-t'ho",
	"-t'ho",
	"tho",
	"tho",
	"et ho",
	"te la",
	"te l'",
	"-te-la",
	"-te-la",
	"tela",
	"tela",
	"et la",
	"te les",
	"te les",
	"-te-les",
	"-te-les",
	"teles",
	"teles",
	"et les",
	"te li",
	"te li",
	"-te-li",
	"-te-li",
	"teli",
	"teli",
	"et li",
	"te'l",
	"te l'",
	"-te'l",
	"-te'l",
	"tel",
	"tel",
	"et el",
	"te'ls",
	"te'ls",
	"-te'ls",
	"-te'ls",
	"tels",
	"tels",
	"et els",
	"te'm",
	"te m'",
	"-te'm",
	"-te'm",
	"tem",
	"tem",
	"et em",
	"te'n",
	"te n'",
	"-te'n",
	"-te'n",
	"ten",
	"ten",
	"et en",
	"te'ns",
	"te'ns",
	"-te'ns",
	"-te'ns",
	"tens",
	"tens",
	"et ens",
	"us el",
	"us l'",
	"-vos-el",
	"-us-el",
	"vosel",
	"usel",
	"us el",
	"us els",
	"us els",
	"-vos-els",
	"-us-els",
	"vosels",
	"usels",
	"us els",
	"us em",
	"us m'",
	"-vos-em",
	"-us-em",
	"vosem",
	"usem",
	"us em",
	"us en",
	"us n'",
	"-vos-en",
	"-us-en",
	"vosen",
	"usen",
	"us en",
	"us ens",
	"us ens",
	"-vos-ens",
	"-us-ens",
	"vosens",
	"usens",
	"us ens",
	"us hi",
	"us hi",
	"-vos-hi",
	"-us-hi",
	"voshi",
	"ushi",
	"us hi",
	"us ho",
	"us ho",
	"-vos-ho",
	"-us-ho",
	"vosho",
	"usho",
	"us ho",
	"us la",
	"us l'",
	"-vos-la",
	"-us-la",
	"vosla",
	"usla",
	"us la",
	"us les",
	"us les",
	"-vos-les",
	"-us-les",
	"vosles",
	"usles",
	"us les",
	"us li",
	"us li",
	"-vos-li",
	"-us-li",
	"vosli",
	"usli",
	"us li",
	"us",
	"us",
	"-vos",
	"-us",
	"vos",
	"us",
	"us",
	"se me'n",
	"se me n'",
	"-se-me'n",
	"-se-me'n",
	"semen",
	"semen",
	"es em en",
	"se te'n",
	"se te n'",
	"-se-te'n",
	"-se-te'n",
	"seten",
	"seten",
	"es et en",
	"se li'n",
	"se li n'",
	"-se-li'n",
	"-se-li'n",
	"selin",
	"selin",
	"es li en",
	"se'ns en",
	"se'ns n'",
	"-se'ns-en",
	"-se'ns-en",
	"sensen",
	"sensen",
	"es ens en",
	"se us en",
	"se us n'",
	"-se-us-en",
	"-se-us-en",
	"seusen",
	"seusen",
	"es us en",
	"se vos en",
	"se vos n'",
	"-se-vos-en",
	"-se-vos-en",
	"sevosen",
	"sevosen",
	"es vos en",
	"se'ls en",
	"se'ls n'",
	"-se'ls-en",
	"-se'ls-en",
	"selsen",
	"selsen",
	"es els en",
}

var incorrectOrders = map[string]string{
	"me se": "se'm",
	"me s'": "se m'",
	"te se": "se't",
	"te s'": "se t'",
	"li se": "se li",
	"li s'": "se li",
	"mi":    "m'hi",
	"si":    "s'hi",
	"nosi":  "-nos-hi",
	"losi":  "-los-hi",
	"lis":   "els",
	"m'en":  "me'n",
	"t'en":  "te'n",
	"s'en":  "se'n",
}

var reflexivePronoun = map[string]string{
	"1S": "em",
	"2S": "et",
	"3S": "es",
	"1P": "ens",
	"2P": "us",
	"3P": "es",
}

var dativePronoun = map[string]string{
	"1S": "em",
	"2S": "et",
	"3S": "li",
	"3C": "li",
	"1P": "ens",
	"2P": "us",
	"3P": "els",
}

var (
	pApostropheNeeded            = regexp.MustCompile(`(?i)^h?[aeiouàèéíòóú].*`)
	pApostropheNeededEnd         = regexp.MustCompile(`(?i).*[aei]$`)
	PPronomFeble                 = regexp.MustCompile(`^P0.{6}$|^PP3CN000$|^PP3NN000$|^PP3..A00$|^PP[123]CP000$|^PP3CSD00$`)
	pContainsReflexivePronoun    = regexp.MustCompile(`(?i).*([mts][e']|[e'][mts]|vos|us|ens|-nos|-vos).*`)
	deWrongApostrophation        = regexp.MustCompile(`(?i).*d'[^aeiouh].*`)
	pronounWrongApostrophation   = regexp.MustCompile(`(?i)^([mts])'([^aeiouh].*)$`)
	pronounMissingApostrophation = regexp.MustCompile(`(?i)(.*)\be([stm]) (h?[aeiouh].*)$`)
	pronounWrongHypphen          = regexp.MustCompile(`(?i)(.*)(-[stm])e-(h[oi])$`)
)

// LReflexivePronouns is the set of simple reflexive clitics.
var LReflexivePronouns = map[string]struct{}{
	"em": {}, "et": {}, "es": {}, "ens": {}, "us": {}, "vos": {},
}

// GetReflexivePronoun returns em/et/es/ens/us for person+number keys like "1S".
func GetReflexivePronoun(key string) string {
	return reflexivePronoun[key]
}

// GetDativePronoun returns em/et/li/ens/us/els for person+number keys.
func GetDativePronoun(key string) string {
	return dativePronoun[key]
}

// Transform maps a weak pronoun form to another position variant.
func Transform(inputPronom string, pronounPos PronounPosition) string {
	inputPronom = strings.ToLower(strings.TrimSpace(inputPronom))
	if v, ok := incorrectOrders[inputPronom]; ok {
		inputPronom = v
	}
	i := 0
	for i < len(pronomsFebles) && !strings.EqualFold(inputPronom, pronomsFebles[i]) {
		i++
	}
	pfPos := int(pronounPosCount)*(i/int(pronounPosCount)) + int(pronounPos)
	if pfPos > len(pronomsFebles)-1 {
		return ""
	}
	pronom := pronomsFebles[pfPos]
	if pronounPos == PronounDavant || (pronounPos == PronounDavantApos && !strings.HasSuffix(pronom, "'")) {
		pronom = pronom + " "
	}
	return pronom
}

// TransformDavant chooses davant / davant_apos based on the following word.
func TransformDavant(inputPronom, nextWord string) string {
	nextWord = strings.ToLower(nextWord)
	if pApostropheNeeded.MatchString(nextWord) {
		return Transform(inputPronom, PronounDavantApos)
	}
	pronom := Transform(inputPronom, PronounDavant)
	if pronom == "es " && (strings.HasPrefix(nextWord, "s") || strings.HasPrefix(nextWord, "ce") || strings.HasPrefix(nextWord, "ci")) {
		return "se "
	}
	return pronom
}

// TransformDarrere chooses darrere / darrere_apos based on the preceding word.
func TransformDarrere(inputPronom, previousWord string) string {
	if pApostropheNeededEnd.MatchString(previousWord) {
		return Transform(inputPronom, PronounDarrereApos)
	}
	return Transform(inputPronom, PronounDarrere)
}

// DoAddPronounEn adds "en" (or converts hi→en hi) to a pronoun cluster.
func DoAddPronounEn(pronounsStr, verbStr string, pronounsAfter bool) string {
	pronounNormalized := Transform(pronounsStr, PronounNormalized)
	if strings.HasSuffix(pronounNormalized, "hi") {
		pronounNormalized = strings.Replace(pronounNormalized, "hi", "en hi", 1)
	} else {
		pronounNormalized += " en"
	}
	if pronounsAfter {
		return TransformDarrere(pronounNormalized, verbStr)
	}
	return TransformDavant(pronounNormalized, verbStr)
}

// DoRemovePronounReflexive strips em/et/es/ens/us/vos from a cluster.
func DoRemovePronounReflexive(pronounsStr, verbStr string, pronounsAfter bool) string {
	re := regexp.MustCompile(`(?i)(em|et|es|ens|us|vos)`)
	pronounsReplacement := strings.TrimSpace(re.ReplaceAllString(Transform(strings.ToLower(pronounsStr), PronounNormalized), ""))
	if pronounsAfter {
		replacement := verbStr
		pronounsReplacement = TransformDarrere(pronounsReplacement, verbStr)
		if pronounsReplacement != "" {
			replacement = verbStr + pronounsReplacement
		}
		return replacement
	}
	replacement := verbStr
	pronounsReplacement = TransformDavant(pronounsReplacement, verbStr)
	if pronounsReplacement != "" {
		replacement = pronounsReplacement + verbStr
	}
	return replacement
}

// ConvertPronounsForIntransitiveVerb rewrites object clitics toward dative/hi.
func ConvertPronounsForIntransitiveVerb(s string) string {
	s = strings.ReplaceAll(s, "-se'l", "-se-li")
	s = strings.ReplaceAll(s, "se'l ", "se li ")
	s = strings.ReplaceAll(s, "l'", "li ")
	s = strings.ReplaceAll(s, "-lo", "-li")
	s = strings.ReplaceAll(s, "-la", "-li")
	s = strings.ReplaceAll(s, "la ", "li ")
	s = strings.ReplaceAll(s, "el ", "li ")
	s = strings.ReplaceAll(s, "ho", "hi")
	return s
}

// FixApostrophes repairs common weak-pronoun apostrophe/hyphen mistakes.
func FixApostrophes(s string) string {
	if deWrongApostrophation.MatchString(s) {
		s = strings.ReplaceAll(s, "d'", "de ")
	}
	if m := pronounMissingApostrophation.FindStringSubmatch(s); m != nil {
		s = m[1] + m[2] + "'" + m[3]
	}
	if m := pronounWrongApostrophation.FindStringSubmatch(s); m != nil {
		s = "e" + m[1] + " " + m[2]
	}
	if m := pronounWrongHypphen.FindStringSubmatch(s); m != nil {
		s = m[1] + m[2] + "'" + m[3]
	}
	return s
}

// DoReplaceEmEn rewrites em/m'/m'hi → en/n'/n'hi before a verb.
func DoReplaceEmEn(pronounsStr, verbStr string, pronounsAfter bool) string {
	_ = pronounsAfter
	switch strings.ToLower(pronounsStr) {
	case "em":
		return "en " + verbStr
	case "m'":
		return "n'" + verbStr
	case "m'hi":
		return "n'hi " + verbStr
	}
	return ""
}

// DoReplaceHiEn rewrites hi → en (same special cases as DoReplaceEmEn for m'/m'hi).
func DoReplaceHiEn(pronounsStr, verbStr string, pronounsAfter bool) string {
	_ = pronounsAfter
	switch strings.ToLower(pronounsStr) {
	case "hi":
		return "en " + verbStr
	case "m'":
		return "n'" + verbStr
	case "m'hi":
		return "n'hi " + verbStr
	}
	return ""
}

// DoAddPronounReflexive adds a reflexive clitic before/after the verb.
func DoAddPronounReflexive(pronounsStr, verbStr, firstVerbPersonaNumber string, pronounsAfter bool) string {
	if pronounsAfter {
		if pContainsReflexivePronoun.MatchString(strings.ToLower(pronounsStr)) {
			return verbStr + TransformDarrere(pronounsStr, verbStr)
		}
		if strings.HasSuffix(verbStr, "r") || strings.HasSuffix(verbStr, "re") {
			return verbStr + TransformDarrere("-se", verbStr)
		}
		return verbStr
	}
	pronounToAdd := Transform(pronounsStr, PronounNormalized)
	if !containsAnyReflexivePronoun(pronounsStr) {
		pronounToAdd = GetReflexivePronoun(firstVerbPersonaNumber) + " " + pronounToAdd
	}
	return TransformDavant(pronounToAdd, verbStr) + verbStr
}

// DoAddPronounReflexiveImperative adds a reflexive after an imperative verb form.
func DoAddPronounReflexiveImperative(pronounsStr, verbStr, firstVerbPersonaNumber string) string {
	if pronounsStr != "" {
		return ""
	}
	pronounToAdd := TransformDarrere(GetReflexivePronoun(firstVerbPersonaNumber), verbStr)
	if pronounToAdd == "" {
		return ""
	}
	return verbStr + pronounToAdd
}

// DoAddPronounReflexiveEn adds reflexive + en.
func DoAddPronounReflexiveEn(pronounsStr, verbStr, firstVerbPersonaNumber string, pronounsAfter bool) string {
	if pronounsAfter {
		if pContainsReflexivePronoun.MatchString(strings.ToLower(pronounsStr)) {
			return verbStr + TransformDarrere(pronounsStr+"'n", verbStr)
		}
		return verbStr + TransformDarrere("-se'n", verbStr)
	}
	if pronounsStr == "" {
		return TransformDavant(GetReflexivePronoun(firstVerbPersonaNumber)+" en", verbStr) + verbStr
	}
	pronounToAdd := TransformDavant("es "+Transform(pronounsStr, PronounNormalized)+" en", verbStr)
	if pronounToAdd != "" {
		return pronounToAdd + verbStr
	}
	return TransformDavant(pronounsStr, verbStr) + verbStr
}

func containsAnyReflexivePronoun(pronounsStr string) bool {
	normalized := strings.Fields(Transform(strings.ToLower(pronounsStr), PronounNormalized))
	for _, p := range normalized {
		if _, ok := LReflexivePronouns[p]; ok {
			return true
		}
	}
	return false
}
