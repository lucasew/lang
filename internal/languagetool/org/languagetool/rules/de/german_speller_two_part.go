package de

import (
	"regexp"
	"strings"
)

// processTwoPartCompounds + resource sets for infix-s / verb stems / prefixes.

var (
	reInvalidCompPart1 = regexp.MustCompile(`^(?:adresse|ahmen|kamp|kontrolle|leuchte|norden|osten|perspektive|schule|sprache|stelle|suche|süden|westen)$`)
	reInvalidCompPart2 = regexp.MustCompile(`^kamp$`)
	// INFIX_S_SUFFIXES: heit|(s|[^c]t|x)ion|ität|keit|ling|ung|schaft|tum
	reInfixSSuffixes   = regexp.MustCompile(`(?:heit|(?:s|[^c]t|x)ion|ität|keit|ling|ung|schaft|tum)$`)
	reCitiesExceptions = regexp.MustCompile(`^(?:bahnhof|flughafen|haupt)`)
	reWechselInfix     = regexp.MustCompile(`^(?:arbeit|dienstag|donnerstag|freitag|montag|mittwoch|link|recht|samstag|sonntag|verband)s?$`)
	reConfusedPrefixes = regexp.MustCompile(`^(?:bach|bade?|wi(?:e)?der)$`)
	// SUBINF_SINGULAR_OBJECT: putzen|rauchen|sein|spielen
	reSubInfSingularObject = regexp.MustCompile(`^(?:putzen|rauchen|sein|spielen)$`)
	// NEEDS_TO_BE_PLURAL — Java GermanSpellerRule.NEEDS_TO_BE_PLURAL (lemma must be plural in compounds)
	reNeedsToBePlural = regexp.MustCompile(`^(?:absolvent(in)?|adressat|aktie|antenne|apache|arbeitnehmer(in)?|ärztin|assistent(in)?|astronom(in)?|asylant(in)?|autor(in)?|azteke|bakterie|ballade|bauer|billion|bisexuelle|blume|bonze|börse|bot(e|in)|buche|bürg(e|in)|bürokrat(in)?|chrysantheme|dän(e|in)?|debatte|debitor(in)?|decke|diakon(in)?|diktator(in)?|direktor(in)?|doktorand(in)?|domäne|dozent(in)?|drohne|druid(e|in)?|düne|ehre|eibe|elefant|elektron|ellipse|emittent(in)?|elfe|elle|enge|erbse|eremit|erde|erste|esche|exot(e|in)?|expert(e|in)?|extremist(in)?|fabrikant(in)?|falke|fassade|farbe|fasan|favorit(in)?|felge|ferien|figur|fluor|frage|franz(ose|ösin)|frau|frisur|förde|galle|gatt(e|in)?|gerät|gepard|gezeit|gigant|gilde|göttin|griech(e|in)?|halt|heid(e|in)?|herde|historie|hölle|höhle|hose|hugenott(e|in)?|hund|hündin|immigrant(in)?|investor(in)?|irokes(e|in)|islamist(in)?|jesuit(in)?|jungfer|jungfrau|junggesell(e|in)|juror(in)?|kadett|kante|kaskade|kathode|katholik(in)?|katze|kette|kid|klasse|kirche|klaue|klient(in)?|klinge|knappe|koeffizient|kojote|komet|kommentator(in)?|komödie|kompliz(e|in)|konkurrent(in)?|konfirmand(in)?|konsonant|kontrahent(in)?|krake|kralle|kranke|krähe|kraut|krippe|kurd(e|in)|kuriosität|kurve|kusine|küste|laie|laterne|laute|legende|lehne|leise|lektor(in)?|leopard|lerche|leserin|lieferant(in)?|lippe|loge|lotse|länge|läuse|löwe|lücke|luke|made|mädel|maske|maßnahme|matriarchin|menge|mensch|metapher|methode|metropole|miene|miete|migrant(in)?|million|mitte|maus|moderator(in)?|monarch(in)?|mongol(e|in)|mormone|mücke|mühle|musikant(in)?|mysterium|nerv|niederlage|nixe|nonne|note|obdachlose|ode|organist|panne|papagei|parzelle|pastor(in)?|pate|patient|patriarch(in)?|petze|pfadfinderin|pfanne|pfaffe|pfau|pfeife|platte|polle|pomade|pomeranze|posse|praktikant(in)?|prinz(essin)?|prise|produzent(in)?|prominente|prophet(in)?|prototyp|prälat|psychopath(in)?|puppe|pädophile|pygmäe|rabe|radikale|rakete|rampe|ranke|rassist(in)?|rate|raupe|rendite|repressalie|rest|riese|rinde|rind|robbe|robe|romanist|rose|ross|route|nummer|runde|russ(e|in)?|röhre|rübe|salbe|schabe|schale|scheide|schelle|schenke|schere|sphäre|dicke|kröte|schauspielerin|schimpans(e|in)|schlampe|schlange|schluchte|schmiere|schnake|schnalle|schneide|schnelle|schokolade|schotte|schurke|schwabe|schwalbe|schwede|schwule|seele|seide|seite|senator(in)?|serb(e|in)?|serie|silbe|skulptur|sonne|sorge|sorte|spanne|sparte|spatz|sperre|spitze|sproße|spule|stalaktit|steppe|straße|streife|studie|stunde|stütze|tabelle|therapeut(in)?|tinte|tote|toilette|torte|traube|treffe|treppe|truhe|träne|tunte|tüte|tyrann|urne|utensil|vandal(e|in)|vasall(in)?|vene|versicherte|verwandte|veteran(in)?|virtuose|vorname|waffe|wanne|ware|watte|wehe|welle|welpe|wiese|wirtin|zar(in)?|zentrum|zutat)$`)
)

// LoadWordSetFile loads non-comment lines into a set (Java loadFile for infix/verb lists).
func LoadWordSetFile(path string) (map[string]struct{}, error) {
	lines, err := LoadSpellingWordList(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(lines))
	for _, w := range lines {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		out[w] = struct{}{}
	}
	return out, nil
}

// InitCompoundResourceFiles ports GermanSpellerRule.init() loads of
// words_infix_s.txt, verb_stems.txt, verb_prefixes.txt, other_prefixes.txt.
func (r *GermanSpellerRule) InitCompoundResourceFiles(infixS, verbStems, verbPrefixes, otherPrefixes string) error {
	if r == nil {
		return nil
	}
	load := func(path string, dest *map[string]struct{}) error {
		if path == "" {
			return nil
		}
		m, err := LoadWordSetFile(path)
		if err != nil {
			return err
		}
		*dest = m
		return nil
	}
	if err := load(infixS, &r.WordsNeedingInfixS); err != nil {
		return err
	}
	if err := load(verbStems, &r.VerbStems); err != nil {
		return err
	}
	if err := load(verbPrefixes, &r.VerbPrefixes); err != nil {
		return err
	}
	return load(otherPrefixes, &r.OtherPrefixes)
}

func (r *GermanSpellerRule) setHas(m map[string]struct{}, word string) bool {
	if r == nil || len(m) == 0 || word == "" {
		return false
	}
	_, ok := m[word]
	return ok
}

// isNounNom ports isNounNom (any SUB:NOM…).
func (r *GermanSpellerRule) isNounNom(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "SUB:NOM") {
			return true
		}
	}
	return false
}

// isNounNomSin / isNounNomPlu / isCountryOrRegionNomSin — TagPOS fail-closed.
func (r *GermanSpellerRule) isNounNomSin(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "SUB:NOM:SIN") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isNounNomPlu(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "SUB:NOM:PLU") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isCountryOrRegionNomSin(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	// Java: EIG:NOM:SIN.+(COU|GEB|STD|WAT)
	re := regexp.MustCompile(`^EIG:NOM:SIN.+(?:COU|GEB|STD|WAT)`)
	for _, t := range r.TagPOS(word) {
		if re.MatchString(t) {
			return true
		}
	}
	return false
}

// needsInfixS ports GermanSpellerRule.needsInfixS.
func (r *GermanSpellerRule) needsInfixS(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if reInfixSSuffixes.MatchString(word) {
		// POS must include SUB:NOM:SIN:FEM when tagger present
		if r.TagPOS != nil {
			for _, t := range r.TagPOS(word) {
				if strings.HasPrefix(t, "SUB:NOM:SIN:FEM") {
					return true
				}
			}
		}
	}
	return r.setHas(r.WordsNeedingInfixS, word)
}

func (r *GermanSpellerRule) isVerbStem(word string) bool {
	return r != nil && r.setHas(r.VerbStems, lowercaseFirstChar(word))
}

func (r *GermanSpellerRule) isOtherPrefix(word string) bool {
	return r != nil && r.setHas(r.OtherPrefixes, lowercaseFirstChar(word))
}

func (r *GermanSpellerRule) isVerbPrefix(word string) bool {
	return r != nil && r.setHas(r.VerbPrefixes, lowercaseFirstChar(word))
}

// needsToBePlural ports GermanSpellerRule.needsToBePlural (lemma in NEEDS_TO_BE_PLURAL).
func needsToBePlural(lemmaLower string) bool {
	return lemmaLower != "" && reNeedsToBePlural.MatchString(lemmaLower)
}

// ProcessTwoPartCompounds ports processTwoPartCompounds (plural gates +
// WECHSELINFIX/CONFUSED/Wechselnumerus + main accept arms).
func (r *GermanSpellerRule) ProcessTwoPartCompounds(part1, part2 string) bool {
	if r == nil || part1 == "" || part2 == "" {
		return false
	}
	part1upcased := uppercaseFirstChar(part1)
	part2upcased := uppercaseFirstChar(part2)
	part1WithoutHyphen := strings.TrimSuffix(part1, "-")
	part2upcasedIsNoun := r.isNoun(part2upcased)
	// Java: isMisspelled(uppercaseFirstChar(part2upcased)) — part2upcased already upper
	part2upcasedIsMispelled := r.IsMisspelled(part2upcased)

	if reInvalidCompPart1.MatchString(lowercaseFirstChar(part1WithoutHyphen)) {
		return false
	}
	if reInvalidCompPart2.MatchString(lowercaseFirstChar(part2)) {
		return false
	}

	part1WithoutInfixS := part1upcased
	// Sometimes part1 requires singular or plural
	part1Lemma := r.findLemmaForNoun(strings.TrimSuffix(part1, "-"))
	if part1Lemma == "" && strings.HasSuffix(strings.TrimSuffix(part1, "-"), "s") {
		part1Lemma = r.findLemmaForNoun(removeTrailingSAndHyphen(part1))
		part1WithoutInfixS = strings.TrimSuffix(part1upcased, "s")
	}

	// Allow part1 to be a plural noun only under conditions — else reject pure plurals
	// Java: if (isNounNomPlu && !isNounNomSin && !country && (!subVerInf || subVerInf&&singularObject)
	//   && !needsToBePlural && !WECHSELNUMERUS && !WECHSELINFIX(lemma-s)) return false
	if r.isNounNomPlu(part1WithoutInfixS) && !r.isNounNomSin(part1WithoutInfixS) &&
		!r.isCountryOrRegionNomSin(part1WithoutInfixS) &&
		(!r.isSubVerInf(part2upcased) ||
			(r.isSubVerInf(part2upcased) && reSubInfSingularObject.MatchString(lowercaseFirstChar(part2)))) &&
		!needsToBePlural(lowercaseFirstChar(part1Lemma)) &&
		!reWechselnumerus.MatchString(lowercaseFirstChar(part1Lemma)) &&
		!reWechselInfix.MatchString(lowercaseFirstChar(strings.TrimSuffix(part1Lemma, "s"))) {
		return false
	}
	// part1 always needs to be plural
	if needsToBePlural(lowercaseFirstChar(part1Lemma)) && r.isNounNomSin(part1WithoutInfixS) {
		return false
	}
	// WECHSELNUMERUS on part1 lemma
	if reWechselnumerus.MatchString(lowercaseFirstChar(part1Lemma)) {
		if !r.checkPluralForPart1Part2Combination(part1, part2) {
			return false
		}
	}

	// WECHSELINFIX → checkInfixS (Java passes part1 surface as used in equals "Arbeit")
	if reWechselInfix.MatchString(lowercaseFirstChar(part1)) {
		return r.checkInfixSForPart1Part2Combination(part1upcased, part2)
	}
	// CONFUSED_PREFIXES → checkConfusion
	if reConfusedPrefixes.MatchString(lowercaseFirstChar(part1)) && !part2upcasedIsMispelled && part2upcasedIsNoun {
		return r.checkConfusionForPart1Part2Combination(part1upcased, part2)
	}

	// Main accept arms (require noun second part that spells OK)
	if !part2upcasedIsNoun || part2upcasedIsMispelled {
		return false
	}

	if strings.HasSuffix(part1WithoutHyphen, "s") &&
		(r.isNounNom(part1upcased) || r.isVerbStem(part1)) &&
		!r.needsInfixS(strings.TrimSuffix(part1upcased, "s")) {
		return true
	}
	if strings.HasSuffix(part1WithoutHyphen, "s") &&
		r.isNounNom(removeTrailingSAndHyphen(part1upcased)) &&
		r.needsInfixS(removeTrailingSAndHyphen(part1upcased)) {
		return true
	}
	if !strings.HasSuffix(part1WithoutHyphen, "s") &&
		(r.isNounNom(part1upcased) || r.isVerbStem(part1)) &&
		!r.needsInfixS(part1upcased) {
		return true
	}
	if (isAllUpperCase(removeTrailingSAndHyphen(part1)) && !r.IsMisspelled(removeTrailingSAndHyphen(part1))) ||
		r.isOtherPrefix(part1) {
		return true
	}
	if r.isCountryOrRegionNomSin(part1) && !reCitiesExceptions.MatchString(lowercaseFirstChar(part2)) {
		return true
	}
	return false
}

// ProcessThreePartCompound ports processThreePartCompound.
func (r *GermanSpellerRule) ProcessThreePartCompound(parts []string) bool {
	if r == nil || len(parts) != 3 {
		return false
	}
	part1, part2, part3 := parts[0], parts[1], parts[2]
	compound1 := part1 + part2
	compound2 := uppercaseFirstChar(part2) + part3

	if r.isNoun(compound1) && r.isNoun(compound2) {
		return r.ProcessTwoPartCompounds(part1, strings.TrimSuffix(part2, "-")) &&
			r.ProcessTwoPartCompounds(part2, part3)
	}
	if strings.HasSuffix(compound1, "s") || strings.HasSuffix(compound1, "s-") {
		p2 := removeTrailingSAndHyphen(part2)
		return r.ProcessTwoPartCompounds(part1, p2) && r.ProcessTwoPartCompounds(part2, part3)
	}
	if r.isVerbPrefix(part1) && r.isVerbStem(part2) && r.isNoun(compound2) {
		return true
	}
	if r.isNounNomSin(part1) && r.isVerbStem(part2) && r.isNoun(compound2) {
		return true
	}
	if r.isNounNom(part1) && r.isOtherPrefix(part2) && r.isNoun(compound2) {
		return true
	}
	if r.isOtherPrefix(part1) && r.isVerbStem(part2) && r.isNoun(compound2) {
		return true
	}
	return false
}

// isValidPartLength ports isValidPartLength for 2/3-part splits.
// Java: parts.get(i).length() (UTF-16).
func isValidPartLength(parts []string) bool {
	switch len(parts) {
	case 2:
		return utf16LenDE(parts[0]) >= 3 && utf16LenDE(parts[1]) >= 4
	case 3:
		return utf16LenDE(parts[0]) >= 3 && utf16LenDE(parts[1]) >= 4 && utf16LenDE(parts[2]) >= 4
	default:
		return false
	}
}

// avoidInfixSAsSingleToken ports avoidInfixSAsSingleToken ("s" glued to predecessor).
func avoidInfixSAsSingleToken(parts []string) []string {
	if len(parts) == 0 {
		return parts
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "s" && len(out) > 0 {
			out[len(out)-1] = out[len(out)-1] + "s"
			continue
		}
		out = append(out, p)
	}
	return out
}
