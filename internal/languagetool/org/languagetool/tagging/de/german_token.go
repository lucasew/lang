package de

// POSType ports GermanToken.POSType name strings (toString values).
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

// Kasus ports Morphy case codes used by AnalyzedGermanToken / agreement tools.
// (GermanToken.Kasus toString is Nominativ/…; Morphy tags use NOM/….)
type Kasus string

const (
	KasusNom Kasus = "NOM"
	KasusAkk Kasus = "AKK"
	KasusDat Kasus = "DAT"
	KasusGen Kasus = "GEN"
)

// Numerus ports Morphy number codes.
type Numerus string

const (
	NumerusSin Numerus = "SIN"
	NumerusPlu Numerus = "PLU"
)

// Genus ports Morphy gender codes (ALG = Allgemein).
// NOG is not a stored genus in Java AnalyzedGermanToken — it maps to FEMININUM.
type Genus string

const (
	GenusMas Genus = "MAS"
	GenusFem Genus = "FEM"
	GenusNeu Genus = "NEU"
	// GenusALG ports Genus.ALLGEMEIN (Morphy ALG expands to all genders).
	GenusALG Genus = "ALG"
	// GenusNoG is retained only for legacy callers; ParseGermanPOS maps NOG → GenusFem.
	GenusNoG Genus = "NOG"
)

// Determination ports GermanToken.Determination object identity for agreement strings.
// Java toString is "definit"/"indefinit"; Go agreement strings use DEFINITE/INDEFINITE
// (existing twin of AgreementTools.makeString).
type Determination string

const (
	DetDefinite   Determination = "DEFINITE"
	DetIndefinite Determination = "INDEFINITE"
)

// AnalyzedGermanToken ports tagging.de.AnalyzedGermanToken.
type AnalyzedGermanToken struct {
	Reading       string
	Type          POSType // empty when Java leaves type null
	Kasus         Kasus
	Numerus       Numerus
	Genus         Genus
	Determination Determination
}

// ParseGermanPOS ports AnalyzedGermanToken(AnalyzedToken) field extraction.
// Java: null/short tags (<3 colon parts) leave all fields null; otherwise scan parts
// with EIG/PA always assigning type and SUB/VER/ADJ/PRO/ART only if type still null.
// NOG → FEMININUM (Java comment: no genus because only used as plural).
func ParseGermanPOS(pos string) AnalyzedGermanToken {
	a := AnalyzedGermanToken{Reading: pos}
	if pos == "" {
		return a
	}
	parts := splitColon(pos)
	// Java: StringUtils.split(posTag, ':').length < 3 → all null
	if len(parts) < 3 {
		return a
	}
	var (
		tempType POSType
		hasType  bool
	)
	for _, part := range parts {
		switch part {
		case "EIG":
			tempType = POSProperNoun
			hasType = true
		case "SUB":
			if !hasType {
				tempType = POSNomen
				hasType = true
			}
		case "PA1", "PA2":
			tempType = POSPartizip
			hasType = true
		case "VER":
			if !hasType {
				tempType = POSVerb
				hasType = true
			}
		case "ADJ":
			if !hasType {
				tempType = POSAdjektiv
				hasType = true
			}
		case "PRO":
			if !hasType {
				tempType = POSPronomen
				hasType = true
			}
		case "ART":
			if !hasType {
				tempType = POSDeterminer
				hasType = true
			}
		case "AKK":
			a.Kasus = KasusAkk
		case "GEN":
			a.Kasus = KasusGen
		case "NOM":
			a.Kasus = KasusNom
		case "DAT":
			a.Kasus = KasusDat
		case "PLU":
			a.Numerus = NumerusPlu
		case "SIN":
			a.Numerus = NumerusSin
		case "MAS":
			a.Genus = GenusMas
		case "FEM":
			a.Genus = GenusFem
		case "NEU":
			a.Genus = GenusNeu
		case "NOG":
			// Java: tempGenus = Genus.FEMININUM
			a.Genus = GenusFem
		case "ALG":
			a.Genus = GenusALG
		case "IND":
			a.Determination = DetIndefinite
		case "DEF":
			a.Determination = DetDefinite
		}
	}
	if hasType {
		a.Type = tempType
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

// GermanToken is the Java-name twin namespace for POS/Kasus/Numerus constants
// (org.languagetool.tagging.de.GermanToken).
type GermanToken struct{}
