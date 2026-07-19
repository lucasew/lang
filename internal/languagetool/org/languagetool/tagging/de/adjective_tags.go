package de

// AdjectiveTags ports org.languagetool.tagging.de.AdjectiveTags.

// TagsForAdj base forms like "chemisch".
var TagsForAdj = []string{"ADJ:PRD:GRU"}

// TagsForAdjE forms ending in -e like "chemische".
var TagsForAdjE = []string{
	"ADJ:AKK:PLU:FEM:GRU:SOL",
	"ADJ:AKK:PLU:MAS:GRU:SOL",
	"ADJ:AKK:PLU:NEU:GRU:SOL",
	"ADJ:AKK:SIN:FEM:GRU:DEF",
	"ADJ:AKK:SIN:FEM:GRU:IND",
	"ADJ:AKK:SIN:FEM:GRU:SOL",
	"ADJ:AKK:SIN:NEU:GRU:DEF",
	"ADJ:NOM:PLU:FEM:GRU:SOL",
	"ADJ:NOM:PLU:MAS:GRU:SOL",
	"ADJ:NOM:PLU:NEU:GRU:SOL",
	"ADJ:NOM:SIN:FEM:GRU:DEF",
	"ADJ:NOM:SIN:FEM:GRU:IND",
	"ADJ:NOM:SIN:FEM:GRU:SOL",
	"ADJ:NOM:SIN:MAS:GRU:DEF",
	"ADJ:NOM:SIN:NEU:GRU:DEF",
}

// TagsForAdjEn forms ending in -en (full Java list).
var TagsForAdjEn = []string{
	"ADJ:AKK:PLU:FEM:GRU:DEF",
	"ADJ:AKK:PLU:FEM:GRU:IND",
	"ADJ:AKK:PLU:MAS:GRU:DEF",
	"ADJ:AKK:PLU:MAS:GRU:IND",
	"ADJ:AKK:PLU:NEU:GRU:DEF",
	"ADJ:AKK:PLU:NEU:GRU:IND",
	"ADJ:AKK:SIN:MAS:GRU:DEF",
	"ADJ:AKK:SIN:MAS:GRU:IND",
	"ADJ:AKK:SIN:MAS:GRU:SOL",
	"ADJ:DAT:PLU:FEM:GRU:DEF",
	"ADJ:DAT:PLU:FEM:GRU:IND",
	"ADJ:DAT:PLU:FEM:GRU:SOL",
	"ADJ:DAT:PLU:MAS:GRU:DEF",
	"ADJ:DAT:PLU:MAS:GRU:IND",
	"ADJ:DAT:PLU:MAS:GRU:SOL",
	"ADJ:DAT:PLU:NEU:GRU:DEF",
	"ADJ:DAT:PLU:NEU:GRU:IND",
	"ADJ:DAT:PLU:NEU:GRU:SOL",
	"ADJ:DAT:SIN:FEM:GRU:DEF",
	"ADJ:DAT:SIN:FEM:GRU:IND",
	"ADJ:DAT:SIN:MAS:GRU:DEF",
	"ADJ:DAT:SIN:MAS:GRU:IND",
	"ADJ:DAT:SIN:NEU:GRU:DEF",
	"ADJ:DAT:SIN:NEU:GRU:IND",
	"ADJ:GEN:PLU:FEM:GRU:DEF",
	"ADJ:GEN:PLU:FEM:GRU:IND",
	"ADJ:GEN:PLU:MAS:GRU:DEF",
	"ADJ:GEN:PLU:MAS:GRU:IND",
	"ADJ:GEN:PLU:NEU:GRU:DEF",
	"ADJ:GEN:PLU:NEU:GRU:IND",
	"ADJ:GEN:SIN:FEM:GRU:DEF",
	"ADJ:GEN:SIN:FEM:GRU:IND",
	"ADJ:GEN:SIN:MAS:GRU:DEF",
	"ADJ:GEN:SIN:MAS:GRU:IND",
	"ADJ:GEN:SIN:MAS:GRU:SOL",
	"ADJ:GEN:SIN:NEU:GRU:DEF",
	"ADJ:GEN:SIN:NEU:GRU:IND",
	"ADJ:GEN:SIN:NEU:GRU:SOL",
	"ADJ:NOM:PLU:FEM:GRU:DEF",
	"ADJ:NOM:PLU:FEM:GRU:IND",
	"ADJ:NOM:PLU:MAS:GRU:DEF",
	"ADJ:NOM:PLU:MAS:GRU:IND",
	"ADJ:NOM:PLU:NEU:GRU:DEF",
	"ADJ:NOM:PLU:NEU:GRU:IND",
}

// TagsForAdjEr forms ending in -er.
var TagsForAdjEr = []string{
	"ADJ:DAT:SIN:FEM:GRU:SOL",
	"ADJ:GEN:PLU:FEM:GRU:SOL",
	"ADJ:GEN:PLU:MAS:GRU:SOL",
	"ADJ:GEN:PLU:NEU:GRU:SOL",
	"ADJ:GEN:SIN:FEM:GRU:SOL",
	"ADJ:NOM:SIN:MAS:GRU:IND",
	"ADJ:NOM:SIN:MAS:GRU:SOL",
}

// TagsForAdjEm forms ending in -em.
var TagsForAdjEm = []string{
	"ADJ:DAT:SIN:MAS:GRU:SOL",
	"ADJ:DAT:SIN:NEU:GRU:SOL",
}

// TagsForAdjEs forms ending in -es.
var TagsForAdjEs = []string{
	"ADJ:AKK:SIN:NEU:GRU:IND",
	"ADJ:AKK:SIN:NEU:GRU:SOL",
	"ADJ:NOM:SIN:NEU:GRU:IND",
	"ADJ:NOM:SIN:NEU:GRU:SOL",
}

// HasAdjectiveTag reports whether tag looks like a German adjective POS.
func HasAdjectiveTag(tag string) bool {
	return len(tag) >= 3 && (tag[:3] == "ADJ" || tag[:3] == "PA2" || tag[:3] == "PA1")
}

// ToPA2 ports GermanTagger.toPA2: ADJ: → PA2: and append ":VER" for /P expansions.
func ToPA2(tags []string) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		if len(t) >= 4 && t[:4] == "ADJ:" {
			out[i] = "PA2:" + t[4:] + ":VER"
		} else {
			out[i] = t + ":VER"
		}
	}
	return out
}

// AdjectiveTags is the Java-name twin for ending helpers.
type AdjectiveTags struct{}

// ForEnding returns tag lists for common adjective endings.
func (AdjectiveTags) ForEnding(ending string) []string {
	switch ending {
	case "", "base":
		return append([]string(nil), TagsForAdj...)
	case "e":
		return append([]string(nil), TagsForAdjE...)
	case "en":
		return append([]string(nil), TagsForAdjEn...)
	case "er":
		return append([]string(nil), TagsForAdjEr...)
	case "em":
		return append([]string(nil), TagsForAdjEm...)
	case "es":
		return append([]string(nil), TagsForAdjEs...)
	default:
		return nil
	}
}
