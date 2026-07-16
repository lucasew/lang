package de

// AdjectiveTags ports tagging.de.AdjectiveTags base forms (subset for synthesis helpers).
var (
	// TagsForAdj: base form like "chemisch"
	TagsForAdj = []string{"ADJ:PRD:GRU"}
	// TagsForAdjE: forms ending in -e like "chemische"
	TagsForAdjE = []string{
		"ADJ:AKK:PLU:FEM:GRU:SOL",
		"ADJ:AKK:PLU:MAS:GRU:SOL",
		"ADJ:AKK:PLU:NEU:GRU:SOL",
		"ADJ:AKK:SIN:FEM:GRU:DEF",
		"ADJ:AKK:SIN:FEM:GRU:IND",
		"ADJ:AKK:SIN:FEM:GRU:SOL",
		"ADJ:NOM:PLU:FEM:GRU:SOL",
		"ADJ:NOM:PLU:MAS:GRU:SOL",
		"ADJ:NOM:PLU:NEU:GRU:SOL",
		"ADJ:NOM:SIN:FEM:GRU:DEF",
		"ADJ:NOM:SIN:FEM:GRU:IND",
		"ADJ:NOM:SIN:FEM:GRU:SOL",
		"ADJ:NOM:SIN:NEU:GRU:DEF",
	}
)
