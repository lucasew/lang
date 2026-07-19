package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// partizip2contains1PluPra / Prt ports GermanTagger arrays for safe PA2 derivation.
var partizip2contains1PluPra = map[string]struct{}{
	"blasen": {}, "fahren": {}, "fallen": {}, "fangen": {}, "fressen": {}, "geben": {},
	"halten": {}, "kommen": {}, "laden": {}, "lassen": {}, "laufen": {}, "lesen": {},
	"messen": {}, "raten": {}, "schlafen": {}, "schlagen": {}, "sehen": {}, "tragen": {}, "treten": {},
}
var partizip2contains1PluPrt = map[string]struct{}{
	"bieten": {}, "bleiben": {}, "fliegen": {}, "fließen": {}, "heben": {}, "leiden": {},
	"meiden": {}, "scheiden": {}, "schließen": {}, "schreiben": {}, "stehen": {}, "steigen": {},
	"streiten": {}, "treiben": {}, "weisen": {}, "ziehen": {},
}

// postagsPartizipEnding* ports GermanTagger postagsPartizipEndingE/Em/En/Er/Es
// (suffix after "PA2:").
var postagsPartizipEndingE = []string{
	"AKK:PLU:FEM:GRU:SOL:VER", "AKK:PLU:MAS:GRU:SOL:VER", "AKK:PLU:NEU:GRU:SOL:VER",
	"AKK:SIN:FEM:GRU:DEF:VER", "AKK:SIN:FEM:GRU:IND:VER", "AKK:SIN:FEM:GRU:SOL:VER",
	"AKK:SIN:NEU:GRU:DEF:VER", "NOM:PLU:FEM:GRU:SOL:VER", "NOM:PLU:MAS:GRU:SOL:VER",
	"NOM:PLU:NEU:GRU:SOL:VER", "NOM:SIN:FEM:GRU:DEF:VER", "NOM:SIN:FEM:GRU:IND:VER",
	"NOM:SIN:FEM:GRU:SOL:VER", "NOM:SIN:MAS:GRU:DEF:VER", "NOM:SIN:NEU:GRU:DEF:VER",
}
var postagsPartizipEndingEm = []string{
	"DAT:SIN:MAS:GRU:SOL:VER", "DAT:SIN:NEU:GRU:SOL:VER",
}
var postagsPartizipEndingEn = []string{
	"AKK:PLU:FEM:GRU:DEF:VER", "AKK:PLU:FEM:GRU:IND:VER", "AKK:PLU:MAS:GRU:DEF:VER",
	"AKK:PLU:MAS:GRU:IND:VER", "AKK:PLU:NEU:GRU:DEF:VER", "AKK:PLU:NEU:GRU:IND:VER",
	"AKK:SIN:MAS:GRU:DEF:VER", "AKK:SIN:MAS:GRU:IND:VER", "AKK:SIN:MAS:GRU:SOL:VER",
	"DAT:PLU:FEM:GRU:DEF:VER", "DAT:PLU:FEM:GRU:IND:VER", "DAT:PLU:FEM:GRU:SOL:VER",
	"DAT:PLU:MAS:GRU:DEF:VER", "DAT:PLU:MAS:GRU:IND:VER", "DAT:PLU:MAS:GRU:SOL:VER",
	"DAT:PLU:NEU:GRU:DEF:VER", "DAT:PLU:NEU:GRU:IND:VER", "DAT:PLU:NEU:GRU:SOL:VER",
	"DAT:SIN:FEM:GRU:DEF:VER", "DAT:SIN:FEM:GRU:IND:VER", "DAT:SIN:MAS:GRU:DEF:VER",
	"DAT:SIN:MAS:GRU:IND:VER", "DAT:SIN:NEU:GRU:DEF:VER", "DAT:SIN:NEU:GRU:IND:VER",
	"GEN:PLU:FEM:GRU:DEF:VER", "GEN:PLU:FEM:GRU:IND:VER", "GEN:PLU:MAS:GRU:DEF:VER",
	"GEN:PLU:MAS:GRU:IND:VER", "GEN:PLU:NEU:GRU:DEF:VER", "GEN:PLU:NEU:GRU:IND:VER",
	"GEN:SIN:FEM:GRU:DEF:VER", "GEN:SIN:FEM:GRU:IND:VER", "GEN:SIN:MAS:GRU:DEF:VER",
	"GEN:SIN:MAS:GRU:IND:VER", "GEN:SIN:MAS:GRU:SOL:VER", "GEN:SIN:NEU:GRU:DEF:VER",
	"GEN:SIN:NEU:GRU:IND:VER", "GEN:SIN:NEU:GRU:SOL:VER", "NOM:PLU:FEM:GRU:DEF:VER",
	"NOM:PLU:FEM:GRU:IND:VER", "NOM:PLU:MAS:GRU:DEF:VER", "NOM:PLU:MAS:GRU:IND:VER",
	"NOM:PLU:NEU:GRU:DEF:VER", "NOM:PLU:NEU:GRU:IND:VER",
}
var postagsPartizipEndingEr = []string{
	"DAT:SIN:FEM:GRU:SOL:VER", "GEN:PLU:FEM:GRU:SOL:VER", "GEN:PLU:MAS:GRU:SOL:VER",
	"GEN:PLU:NEU:GRU:SOL:VER", "GEN:SIN:FEM:GRU:SOL:VER", "NOM:SIN:MAS:GRU:IND:VER",
	"NOM:SIN:MAS:GRU:SOL:VER",
}
var postagsPartizipEndingEs = []string{
	"AKK:SIN:NEU:GRU:IND:VER", "AKK:SIN:NEU:GRU:SOL:VER",
	"NOM:SIN:NEU:GRU:IND:VER", "NOM:SIN:NEU:GRU:SOL:VER",
}

func lemmaIn(set map[string]struct{}, lemma string) bool {
	_, ok := set[lemma]
	return ok
}

// addPartizip2FromLastPart ports non-separable PA2 derivation + declined PA2 endings
// for prefix+stem forms (e.g. erstickt, erstickter).
func (t *GermanTagger) addPartizip2FromLastPart(
	wordOrig, firstPart, lastPart string,
	idxPos int,
	readings []*languagetool.AnalyzedToken,
) []*languagetool.AnalyzedToken {
	if t == nil || lastPart == "" {
		return readings
	}
	atStartOrLower := idxPos == 0 || wordOrig == strings.ToLower(wordOrig)
	if !atStartOrLower {
		return readings
	}
	fpLow := strings.ToLower(firstPart)
	isNonSep := false
	for _, p := range prefixesNonSeparableVerbs {
		if p == fpLow {
			isNonSep = true
			break
		}
	}
	// 1) lastPart is VER:3:SIN:PRÄ:SFT (or special PLU) → VER:PA2 + PA2:PRD
	if isNonSep {
		for _, taggedWord := range t.TagWordExact(lastPart) {
			pos := taggedWord.PosTag
			ok := strings.HasPrefix(pos, "VER:3:SIN:PRÄ:SFT")
			if !ok && strings.HasPrefix(pos, "VER:1:PLU:PRÄ:NON") && lemmaIn(partizip2contains1PluPra, taggedWord.Lemma) {
				ok = true
			}
			if !ok && strings.HasPrefix(pos, "VER:1:PLU:PRT:NON") && lemmaIn(partizip2contains1PluPrt, taggedWord.Lemma) {
				ok = true
			}
			if !ok {
				continue
			}
			if fpLow != "un" {
				fl := pos
				if len(fl) >= 3 {
					fl = fl[len(fl)-3:]
				}
				lemma := firstPart + taggedWord.Lemma
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(lemma, "VER:PA2:"+fl)))
			}
			readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "PA2:PRD:GRU:VER")))
		}
	}

	// 2) declined: lastPart ends with e/em/en/er/es; middle has VER:3:SIN:PRÄ:SFT
	suffixes := []string{"em", "en", "er", "es", "e"} // longer first for em/en/er/es before e
	middlePart, suffix := "", ""
	for _, sffx := range suffixes {
		if strings.HasSuffix(lastPart, sffx) {
			middlePart = lastPart[:len(lastPart)-len(sffx)]
			suffix = sffx
			break
		}
	}
	if middlePart == "" {
		return readings
	}
	for _, taggedM := range t.TagWordExact(middlePart) {
		if !strings.HasPrefix(taggedM.PosTag, "VER:3:SIN:PRÄ:SFT") {
			continue
		}
		lemma := wordOrig
		if len(wordOrig) >= len(suffix) {
			lemma = wordOrig[:len(wordOrig)-len(suffix)]
		}
		var ends []string
		switch suffix {
		case "e":
			ends = postagsPartizipEndingE
		case "em":
			ends = postagsPartizipEndingEm
		case "en":
			ends = postagsPartizipEndingEn
		case "er":
			ends = postagsPartizipEndingEr
		case "es":
			ends = postagsPartizipEndingEs
		}
		for _, pe := range ends {
			readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(lemma, "PA2:"+pe)))
		}
	}
	return readings
}
