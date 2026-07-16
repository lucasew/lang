package de

// POSType ports GermanToken.POSType constants.
type POSType string

const (
	POSNomen      POSType = "Nomen"
	POSVerb       POSType = "Verb"
	POSAdjektiv   POSType = "Adjektiv"
	POSDeterminer POSType = "Determiner"
	POSPronomen   POSType = "Pronomen"
	POSPartizip   POSType = "Partizip"
	POSProperNoun POSType = "Eigenname"
	POSOther      POSType = "Other"
)

// Kasus ports German case tags.
type Kasus string

const (
	KasusNom Kasus = "NOM"
	KasusAkk Kasus = "AKK"
	KasusDat Kasus = "DAT"
	KasusGen Kasus = "GEN"
)

// Numerus ports number.
type Numerus string

const (
	NumerusSin Numerus = "SIN"
	NumerusPlu Numerus = "PLU"
)

// Genus ports gender.
type Genus string

const (
	GenusMas Genus = "MAS"
	GenusFem Genus = "FEM"
	GenusNeu Genus = "NEU"
	GenusNoG Genus = "NOG"
)

// AnalyzedGermanToken ports tagging.de.AnalyzedGermanToken (POS parse helpers).
type AnalyzedGermanToken struct {
	Reading string
	Type    POSType
	Kasus   Kasus
	Numerus Numerus
	Genus   Genus
}

// ParseGermanPOS extracts coarse fields from a German POS string like "SUB:NOM:SIN:MAS".
func ParseGermanPOS(pos string) AnalyzedGermanToken {
	a := AnalyzedGermanToken{Reading: pos, Type: POSOther}
	if pos == "" {
		return a
	}
	parts := splitColon(pos)
	if len(parts) == 0 {
		return a
	}
	switch parts[0] {
	case "SUB", "EIG":
		if parts[0] == "EIG" {
			a.Type = POSProperNoun
		} else {
			a.Type = POSNomen
		}
	case "VER":
		a.Type = POSVerb
	case "ADJ", "PA1", "PA2":
		if parts[0] == "ADJ" {
			a.Type = POSAdjektiv
		} else {
			a.Type = POSPartizip
		}
	case "ART", "PRO":
		if parts[0] == "ART" {
			a.Type = POSDeterminer
		} else {
			a.Type = POSPronomen
		}
	}
	for _, p := range parts[1:] {
		switch p {
		case "NOM":
			a.Kasus = KasusNom
		case "AKK":
			a.Kasus = KasusAkk
		case "DAT":
			a.Kasus = KasusDat
		case "GEN":
			a.Kasus = KasusGen
		case "SIN":
			a.Numerus = NumerusSin
		case "PLU":
			a.Numerus = NumerusPlu
		case "MAS":
			a.Genus = GenusMas
		case "FEM":
			a.Genus = GenusFem
		case "NEU":
			a.Genus = GenusNeu
		case "NOG":
			a.Genus = GenusNoG
		}
	}
	return a
}

func splitColon(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
