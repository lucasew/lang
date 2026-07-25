package de

// GermanToken ports org.languagetool.tagging.de.GermanToken.
// Java: private constructor — only static nested type constants.
type GermanToken struct{}

// POSType ports GermanToken.POSType. String value == Java POSType.toString().
type POSType string

const (
	// POSNomen ports POSType.NOMEN.
	POSNomen POSType = "Nomen"
	// POSVerb ports POSType.VERB.
	POSVerb POSType = "Verb"
	// POSAdjektiv ports POSType.ADJEKTIV.
	POSAdjektiv POSType = "Adjektiv"
	// POSDeterminer ports POSType.DETERMINER.
	POSDeterminer POSType = "Determiner"
	// POSPronomen ports POSType.PRONOMEN.
	POSPronomen POSType = "Pronomen"
	// POSPartizip ports POSType.PARTIZIP.
	POSPartizip POSType = "Partizip"
	// POSProperNoun ports POSType.PROPER_NOUN.
	POSProperNoun POSType = "Eigenname"
	// POSOther ports POSType.OTHER (e.g. sentence start).
	POSOther POSType = "Other"
)

// Kasus ports GermanToken.Kasus. String value == Java Kasus.toString().
// AnalyzedGermanToken maps Morphy tag parts NOM/AKK/DAT/GEN onto these objects.
type Kasus string

const (
	// KasusNominativ ports Kasus.NOMINATIV.
	KasusNominativ Kasus = "Nominativ"
	// KasusAkkusativ ports Kasus.AKKUSATIV.
	KasusAkkusativ Kasus = "Akkusativ"
	// KasusDativ ports Kasus.DATIV.
	KasusDativ Kasus = "Dativ"
	// KasusGenitiv ports Kasus.GENITIV.
	KasusGenitiv Kasus = "Genitiv"
	// KasusOther ports Kasus.OTHER.
	KasusOther Kasus = "Other"
)

// Short aliases (same values) for Morphy-oriented call sites / historical names.
const (
	KasusNom = KasusNominativ
	KasusAkk = KasusAkkusativ
	KasusDat = KasusDativ
	KasusGen = KasusGenitiv
)

// Numerus ports GermanToken.Numerus. String value == Java Numerus.toString().
// AnalyzedGermanToken maps Morphy SIN/PLU onto these objects.
type Numerus string

const (
	// NumerusSingular ports Numerus.SINGULAR.
	NumerusSingular Numerus = "Singular"
	// NumerusPlural ports Numerus.PLURAL.
	NumerusPlural Numerus = "Plural"
	// NumerusOther ports Numerus.OTHER.
	NumerusOther Numerus = "Other"
)

// Short aliases for Morphy-oriented call sites.
const (
	NumerusSin = NumerusSingular
	NumerusPlu = NumerusPlural
)

// Genus ports GermanToken.Genus. String value == Java Genus.toString().
// AnalyzedGermanToken maps Morphy MAS/FEM/NEU/ALG (and NOG→FEMININUM) onto these objects.
type Genus string

const (
	// GenusNeutrum ports Genus.NEUTRUM.
	GenusNeutrum Genus = "Neutrum"
	// GenusMaskulinum ports Genus.MASKULINUM.
	GenusMaskulinum Genus = "Maskulinum"
	// GenusFemininum ports Genus.FEMININUM.
	GenusFemininum Genus = "Femininum"
	// GenusOther ports Genus.OTHER.
	GenusOther Genus = "Other"
	// GenusAllgemein ports Genus.ALLGEMEIN (Morphy ALG = Mas, Fem, or Neu).
	GenusAllgemein Genus = "Allgemein"
)

// Short aliases for Morphy-oriented call sites.
const (
	GenusMas = GenusMaskulinum
	GenusFem = GenusFemininum
	GenusNeu = GenusNeutrum
	// GenusALG ports Genus.ALLGEMEIN.
	GenusALG = GenusAllgemein
)

// Determination ports GermanToken.Determination. String value == Java Determination.toString().
// AnalyzedGermanToken maps Morphy DEF/IND onto these objects.
type Determination string

const (
	// DetDefinite ports Determination.DEFINITE.
	DetDefinite Determination = "definit"
	// DetIndefinite ports Determination.INDEFINITE.
	DetIndefinite Determination = "indefinit"
)

// DeterminationDefinite / DeterminationIndefinite are Java-name twins of DEFINITE / INDEFINITE.
const (
	DeterminationDefinite   = DetDefinite
	DeterminationIndefinite = DetIndefinite
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
// Morphy codes map onto GermanToken.* objects (whose toString is the German name above).
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
			a.Kasus = KasusAkkusativ
		case "GEN":
			a.Kasus = KasusGenitiv
		case "NOM":
			a.Kasus = KasusNominativ
		case "DAT":
			a.Kasus = KasusDativ
		case "PLU":
			a.Numerus = NumerusPlural
		case "SIN":
			a.Numerus = NumerusSingular
		case "MAS":
			a.Genus = GenusMaskulinum
		case "FEM":
			a.Genus = GenusFemininum
		case "NEU":
			a.Genus = GenusNeutrum
		case "NOG":
			// Java: tempGenus = Genus.FEMININUM
			a.Genus = GenusFemininum
		case "ALG":
			a.Genus = GenusAllgemein
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
