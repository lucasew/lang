package de

// AdjectiveTags ports org.languagetool.tagging.de.AdjectiveTags.
type AdjectiveTags struct{}

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

// TagsForAdjEn forms ending in -en.
var TagsForAdjEn = []string{
	"ADJ:AKK:PLU:FEM:GRU:DEF",
	"ADJ:AKK:PLU:FEM:GRU:IND",
	"ADJ:AKK:PLU:MAS:GRU:DEF",
	"ADJ:AKK:PLU:MAS:GRU:IND",
	"ADJ:AKK:PLU:NEU:GRU:DEF",
	"ADJ:AKK:PLU:NEU:GRU:IND",
	"ADJ:AKK:SIN:MAS:GRU:DEF",
}

// HasAdjectiveTag reports whether tag looks like a German adjective POS.
func HasAdjectiveTag(tag string) bool {
	return len(tag) >= 3 && (tag[:3] == "ADJ" || tag[:3] == "PA2" || tag[:3] == "PA1")
}

// ForEnding returns tag lists for common adjective endings.
func (AdjectiveTags) ForEnding(ending string) []string {
	switch ending {
	case "", "base":
		return append([]string(nil), TagsForAdj...)
	case "e":
		return append([]string(nil), TagsForAdjE...)
	case "en":
		return append([]string(nil), TagsForAdjEn...)
	default:
		return nil
	}
}
