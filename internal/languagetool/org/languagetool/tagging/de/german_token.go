package de

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

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

// AnalyzedGermanToken ports org.languagetool.tagging.de.AnalyzedGermanToken.
// Java fields are nullable; empty zero values are the Go twin of null.
type AnalyzedGermanToken struct {
	Type          POSType // empty when Java leaves type null
	Kasus         Kasus
	Numerus       Numerus
	Genus         Genus
	Determination Determination
}

// NewAnalyzedGermanToken ports AnalyzedGermanToken(AnalyzedToken).
// Java: String posTag = token.getPOSTag(); then parse Morphy colon parts.
func NewAnalyzedGermanToken(token *languagetool.AnalyzedToken) AnalyzedGermanToken {
	if token == nil {
		// Java would NPE on null token; no POS → all null fields.
		return AnalyzedGermanToken{}
	}
	posTag := token.GetPOSTag()
	if posTag == nil {
		return AnalyzedGermanToken{}
	}
	return ParseGermanPOS(*posTag)
}

// GetType ports AnalyzedGermanToken.getType (nullable POSType).
func (a AnalyzedGermanToken) GetType() POSType { return a.Type }

// GetCasus ports AnalyzedGermanToken.getCasus (nullable Kasus).
func (a AnalyzedGermanToken) GetCasus() Kasus { return a.Kasus }

// GetNumerus ports AnalyzedGermanToken.getNumerus (nullable Numerus).
func (a AnalyzedGermanToken) GetNumerus() Numerus { return a.Numerus }

// GetGenus ports AnalyzedGermanToken.getGenus (nullable Genus).
func (a AnalyzedGermanToken) GetGenus() Genus { return a.Genus }

// GetDetermination ports AnalyzedGermanToken.getDetermination (nullable Determination).
func (a AnalyzedGermanToken) GetDetermination() Determination { return a.Determination }

// ParseGermanPOS ports the POS-tag parse body of AnalyzedGermanToken(AnalyzedToken).
// Java: null/short tags (StringUtils.split(posTag, ':').length < 3) leave all fields null;
// otherwise scan parts with EIG/PA always assigning type and SUB/VER/ADJ/PRO/ART only if
// tempType still null. Kasus/Numerus/Genus/Determination: last matching part wins.
// Morphy codes map onto GermanToken.* objects (toString = German name above).
// NOG → FEMININUM (Java comment: no genus because only used as plural).
func ParseGermanPOS(pos string) AnalyzedGermanToken {
	// Java: if (posTag == null || (parts = StringUtils.split(posTag, ':')).length < 3)
	// Empty string is non-null; StringUtils.split("", ':') → length 0 < 3 → all null.
	if pos == "" {
		return AnalyzedGermanToken{}
	}
	parts := stringUtilsSplitColon(pos)
	if len(parts) < 3 {
		return AnalyzedGermanToken{}
	}
	// Java temps start null; empty string is Go twin of null for typed string enums.
	var (
		tempType          POSType
		tempCasus         Kasus
		tempNumerus       Numerus
		tempGenus         Genus
		tempDetermination Determination
	)
	for _, part := range parts {
		// Java if/else-if chain (order matters only for mutual exclusivity of one match per part).
		if part == "EIG" {
			tempType = POSProperNoun
		} else if part == "SUB" && tempType == "" {
			tempType = POSNomen
		} else if part == "PA1" || part == "PA2" {
			tempType = POSPartizip
		} else if part == "VER" && tempType == "" {
			tempType = POSVerb
		} else if part == "ADJ" && tempType == "" {
			tempType = POSAdjektiv
		} else if part == "PRO" && tempType == "" {
			tempType = POSPronomen
		} else if part == "ART" && tempType == "" {
			tempType = POSDeterminer
		} else if part == "AKK" {
			tempCasus = KasusAkkusativ
		} else if part == "GEN" {
			tempCasus = KasusGenitiv
		} else if part == "NOM" {
			tempCasus = KasusNominativ
		} else if part == "DAT" {
			tempCasus = KasusDativ
		} else if part == "PLU" {
			tempNumerus = NumerusPlural
		} else if part == "SIN" {
			tempNumerus = NumerusSingular
		} else if part == "MAS" {
			tempGenus = GenusMaskulinum
		} else if part == "FEM" {
			tempGenus = GenusFemininum
		} else if part == "NEU" {
			tempGenus = GenusNeutrum
		} else if part == "NOG" {
			// Java: tempGenus = Genus.FEMININUM; // NOG = no genus because only used as plural
			tempGenus = GenusFemininum
		} else if part == "ALG" {
			tempGenus = GenusAllgemein
		} else if part == "IND" {
			tempDetermination = DetIndefinite
		} else if part == "DEF" {
			tempDetermination = DetDefinite
		}
	}
	return AnalyzedGermanToken{
		Type:          tempType,
		Kasus:         tempCasus,
		Numerus:       tempNumerus,
		Genus:         tempGenus,
		Determination: tempDetermination,
	}
}

// stringUtilsSplitColon ports Apache Commons Lang3 StringUtils.split(str, ':')
// (preserveAllTokens=false): empty tokens discarded; consecutive separators collapse.
// See StringUtils.splitWorker(str, separatorChar, false).
func stringUtilsSplitColon(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	match := false
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			if match {
				out = append(out, s[start:i])
				match = false
			}
			start = i + 1
			continue
		}
		match = true
	}
	if match {
		out = append(out, s[start:])
	}
	return out
}
