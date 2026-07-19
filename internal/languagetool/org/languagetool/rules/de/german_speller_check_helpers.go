package de

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Ports checkInfixSForPart1Part2Combination, checkConfusionForPart1Part2Combination,
// checkPluralForPart1Part2Combination, findLemmaForNoun, isOldSpelling + alt_neu.csv.

var (
	reArbeitComp = regexp.MustCompile(`^(?:(gebe|nehme)(r(s|n|innen|in)?|nde[mnr]?))$`)
	reBachComp = regexp.MustCompile(`^(?:aue|bett|biograph(ie|in)?|chor|forelle|gesellschaft|kantate|lauf|niederung|stelze|tal|verein)$`)
	reBadComp = regexp.MustCompile(`^(?:accessoire|architektur|besitzer(in)?|design|eintritt|fenster|größe|grund|kollektion|konzept|lösung|maß|möbel|nutzer(in)?|regal|spezialist(in)?|spiegel|straße|tür|utensil)$`)
	reLinkComp = regexp.MustCompile(`^(?:element|inhalt|liste|portal|text|titel|tracking|verzeichnis)$`)
	reLinksComp = regexp.MustCompile(`^(?:abbieger(in)?|abweichler(in)?|anwalt|anwältin|anwaltschaft|ausfall|auslage|ausleger(in)?|au(ss|ß)en|bündnis|drall|drehung|extremer?|extremis(t|tin|mus)|faschis(t|tin|mus)|fraktion|galopp|gewinde|händ(er|erin|igkeit)|hörnchen|innen|intellektueller?|katholizis(t|tin|mus)|koalition|konter|kurs|kurve|lastigkeit|lenker|nationalis(t|tin|mus)|opposition|orientierung|partei|populis(t|tin|mus)|radikal(e|er|ismus|ist|istin)|regierung|ruck|rutsch|schnitt|schuss|schwenk(ung)?|sektierer(in)?|steuerung|terror(t|tin|ismus)|verbinder(in)?|verkehr|wendung|wichser)$`)
	reRechtComp = regexp.MustCompile(`^(?:bank|eck|fertigung|gläubigkeit|haber|haberei|leitung|losigkeit|mäßigkeit|winkligkeit|zeitigkeit)$`)
	reRechtsComp = regexp.MustCompile(`^(?:abbieger(in)?|abteilung|akt|akte|angelegenheit|ansicht|anspruch|anwalt|anwalts|anwaltschaft|anwendung|anwältin|auffassung|aufsicht|auskunft|ausleger(in)?|ausschuss|au(ss|ß)en|begehren|begriff|behelf|beistand|berater|beratung|bereich|beschwerde|beugung|beziehung|brecher|bruch|dienst|drall|durchsetzung|empfinden|entwicklung|setzung|experte|experten|extremer?|extremis(t|tin|mus)|fall|fehler|folge|form|fortbildung|frage|fähigkeit|gebiet|gebieten|gelehrte|gelehrter|geschichte|geschäft|gewinde|gleichheit|grund|grundlage|grundsatz|gründen|gut|gutachten|gültigkeit|güter|handlung|hilfe|händ(er|erin|igkeit)|hängigkeit|inhaber|institut|katholizis(t|tin|mus)|klick|konformität|kraft|kreis|kurve|lage|lastigkeit|lehre|lenker|medizin|mediziner|meinung|missbrauch|mittel|mitteln|mängel|nachfolge|nachfolger|nachfolgerin|nationalis(t|tin|mus)|natur|norm|ordnung|persönlichkeit|pflege|pfleger|pflicht|philosophie|politik|populis(t|tin|mus)|position|praxis|problem|quelle|radikal(e|er|ismus|ist|istin)|rahmen|rat|ratgeber(in)?|ruck|rutsch|sache|sachen|satz|schutz|sicherheit|sinn|sprache|soziologie|sprechung|staat|staatlichkeit|stand|status|stellung|streit|streitigkeit|system|terroris(t|tin|ismus)|texte|texter|thema|theorie|tipp|titel|träger|unsicherheit|verfolgung|vergleichung|verhältnis|verkehr|verletzung|verletzungen|verordnung|verstoß|verständnis|verteidiger|verteidigung|vertreter|vertretung|vorschrift|wahl|weg|wesen|widrigkeit|wirksamkeit|wirkung|wissenschaft|wissenschaften|wissenschaftler|zug|änderung)$`)
	reVerbandComp = regexp.MustCompile(`^(?:klammer|kasten|kiste|mull|material|päckchen|platz|raum|schere|zeug|zimmer)$`)
	reVerbandsComp = regexp.MustCompile(`^(?:chef(in)?|flug|funktionär(in)?|kasse|klage|leben|leitung|leiter(in)?|ligist(in)|material|päckchen|präsident(in)?|presse|spiel|vertreter(in)|vorsitzender?|vorstand|wechsel|zeichen|zeit(schrift|ung))$`)
	reWiderComp = regexp.MustCompile(`^(?:glanz|hall|haken|klage|klang|lager|rechtliche|ruf|spiel|spruch|sehen|stand|streit|wille)$`)
	reWochentagComp = regexp.MustCompile(`^(?:abend|mittag|morgen|nachmittag|vormittag)$`)
	reWochentage = regexp.MustCompile(`^(?:Montag|Dienstag|Mittwoch|Donnerstag|Freitag|Samstag|Sonntag)$`)
	reWochentageS = regexp.MustCompile(`^(?:(Montag|Dienstag|Mittwoch|Donnerstag|Freitag|Samstag|Sonntag)s)$`)
	reWeltenComp = regexp.MustCompile(`^(?:Brand|Bummler(in)?|Drama|Wende)$`)
	reWoerterComp = regexp.MustCompile(`^(?:Buch|Liste|Verzeichnis)$`)
	reWechselnumerus = regexp.MustCompile(`^(?:wort|welt)$`)
)


// findLemmaForNoun ports GermanSpellerRule.findLemmaForNoun.
// Uses LemmaOf when set (tagger lemma for uppercase noun form); empty if unavailable.
func (r *GermanSpellerRule) findLemmaForNoun(word string) string {
	if r == nil || word == "" {
		return ""
	}
	if r.LemmaOf == nil {
		return ""
	}
	return r.LemmaOf(uppercaseFirstChar(word))
}

// part2LemmaForChecks ports the lemma-or-strip-s logic shared by check* helpers.
func (r *GermanSpellerRule) part2LemmaForChecks(part2 string) string {
	lemma := r.findLemmaForNoun(strings.TrimSuffix(part2, "-"))
	if lemma == "" && strings.HasSuffix(strings.TrimSuffix(part2, "-"), "s") {
		lemma = r.findLemmaForNoun(removeTrailingSAndHyphen(part2))
	}
	return lemma
}

// checkInfixSForPart1Part2Combination ports GermanSpellerRule.checkInfixSForPart1Part2Combination.
func (r *GermanSpellerRule) checkInfixSForPart1Part2Combination(part1, part2 string) bool {
	part2Lemma := r.part2LemmaForChecks(part2)
	lcLemma := lowercaseFirstChar(part2Lemma)

	if part1 == "Arbeit" && reArbeitComp.MatchString(part2) {
		return true
	}
	if part1 == "Arbeits" && !reArbeitComp.MatchString(part2) {
		return true
	}
	if part1 == "Link" && reLinkComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Links" && reLinksComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Recht" && reRechtComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Rechts" && reRechtsComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Verband" && reVerbandComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Verbands" && reVerbandsComp.MatchString(lcLemma) {
		return true
	}
	if reWochentage.MatchString(part1) && reWochentagComp.MatchString(lcLemma) {
		return true
	}
	if reWochentageS.MatchString(part1) && !reWochentagComp.MatchString(lcLemma) {
		return true
	}
	return false
}

// checkConfusionForPart1Part2Combination ports checkConfusionForPart1Part2Combination.
func (r *GermanSpellerRule) checkConfusionForPart1Part2Combination(part1, part2 string) bool {
	part2Lemma := r.findLemmaForNoun(strings.TrimSuffix(part2, "-"))
	lcLemma := lowercaseFirstChar(part2Lemma)

	if part1 == "Bad" && reBachComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Bad" && reBadComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Bade" && !reBadComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Wider" && reWiderComp.MatchString(lcLemma) {
		return true
	}
	if part1 == "Wieder" && !reWiderComp.MatchString(lcLemma) {
		return true
	}
	return false
}

// checkPluralForPart1Part2Combination ports checkPluralForPart1Part2Combination.
func (r *GermanSpellerRule) checkPluralForPart1Part2Combination(part1, part2 string) bool {
	part2Lemma := r.part2LemmaForChecks(part2)
	if part1 == "Welt" && !reWeltenComp.MatchString(part2Lemma) {
		return true
	}
	if part1 == "Welten" && reWeltenComp.MatchString(part2Lemma) {
		return true
	}
	if part1 == "Wort" && !reWoerterComp.MatchString(part2Lemma) {
		return true
	}
	if part1 == "Wörter" && reWoerterComp.MatchString(part2Lemma) {
		return true
	}
	return false
}

// LoadAltNeuOldSpelling ports loadFile for /de/alt_neu.csv (column 0 = old form).
func LoadAltNeuOldSpelling(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		old := strings.TrimSpace(strings.SplitN(line, ";", 2)[0])
		if old != "" {
			out[old] = struct{}{}
		}
	}
	return out, sc.Err()
}

// InitOldSpellingFile loads alt_neu.csv old forms into OldSpelling.
func (r *GermanSpellerRule) InitOldSpellingFile(path string) error {
	if r == nil {
		return nil
	}
	m, err := LoadAltNeuOldSpelling(path)
	if err != nil {
		return err
	}
	r.OldSpelling = m
	return nil
}

// isOldSpelling ports GermanSpellerRule.isOldSpelling.
func (r *GermanSpellerRule) isOldSpelling(parts []string) bool {
	if r == nil || len(r.OldSpelling) == 0 {
		return false
	}
	for _, part := range parts {
		cleaned := removeTrailingSAndHyphen(part)
		if strings.HasSuffix(part, "s") {
			if r.setHas(r.OldSpelling, uppercaseFirstChar(part)) ||
				r.setHas(r.OldSpelling, uppercaseFirstChar(cleaned)) ||
				r.setHas(r.OldSpelling, lowercaseFirstChar(part)) ||
				r.setHas(r.OldSpelling, lowercaseFirstChar(cleaned)) {
				return true
			}
		} else {
			if r.setHas(r.OldSpelling, uppercaseFirstChar(cleaned)) ||
				r.setHas(r.OldSpelling, lowercaseFirstChar(cleaned)) {
				return true
			}
		}
	}
	return false
}
