package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// GrammarCategory ports AgreementRule.GrammarCategory bits that AgreementTools
// accepts in its omit set (Java: AgreementRule.GrammarCategory).
// Full enum with displayName lives with AgreementRule; these constants are the
// omit-set members used by AgreementTools.makeString.
type GrammarCategory int

const (
	// CatKasus ports GrammarCategory.KASUS.
	CatKasus GrammarCategory = iota
	// CatNumerus ports GrammarCategory.NUMERUS.
	CatNumerus
	// CatGenus ports GrammarCategory.GENUS.
	CatGenus
)

// GetAgreementCategories ports AgreementTools.getAgreementCategories.
// Return Kasus, Numerus, Genus of those forms with a determiner.
// omit may skip Kasus/Numerus/Genus (Java Set.contains); skipSol skips :SOL readings.
func GetAgreementCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool, skipSol bool) map[string]struct{} {
	// Java: private AgreementTools(); static methods only.
	set := map[string]struct{}{}
	if aToken == nil {
		// Java NPE on null aToken; no twin call site passes nil.
		return set
	}
	if omit == nil {
		// Java emptySet / new HashSet<>() → nothing omitted.
		omit = map[GrammarCategory]bool{}
	}
	// Java: List<AnalyzedToken> readings = aToken.getReadings();
	for _, tmpReading := range aToken.GetReadings() {
		if tmpReading == nil {
			// Java list elements are non-null; skip defensive.
			continue
		}
		// Java: if (skipSol && tmpReading.getPOSTag() != null && tmpReading.getPOSTag().endsWith(":SOL"))
		if skipSol {
			if pos := tmpReading.GetPOSTag(); pos != nil && strings.HasSuffix(*pos, ":SOL") {
				// SOL = alleinstehend - needs to be skipped so we find errors like "An der roter Ampel."
				continue
			}
		}
		// Java: AnalyzedGermanToken reading = new AnalyzedGermanToken(tmpReading);
		reading := detag.NewAnalyzedGermanToken(tmpReading)
		// Java: if (reading.getCasus() == null && reading.getNumerus() == null && reading.getGenus() == null)
		if reading.GetCasus() == "" && reading.GetNumerus() == "" && reading.GetGenus() == "" {
			continue
		}
		pos := tmpReading.GetPOSTag()
		// Java: if (reading.getGenus() == ALLGEMEIN &&
		//   tmpReading.getPOSTag() != null && !tmpReading.getPOSTag().endsWith(":STV") &&
		//   !possessiveSpecialCase(aToken, tmpReading))
		if reading.GetGenus() == detag.GenusAllgemein &&
			pos != nil && !strings.HasSuffix(*pos, ":STV") &&
			!possessiveSpecialCase(aToken, tmpReading) {
			// genus=ALG in the original data. Expand so e.g. "Ich Arbeiter" is not flagged.
			if reading.GetDetermination() == "" {
				// Nouns don't have determination; pretend they fulfill both DEF and IND.
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusMaskulinum, detag.DeterminationDefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusMaskulinum, detag.DeterminationIndefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusFemininum, detag.DeterminationDefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusFemininum, detag.DeterminationIndefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusNeutrum, detag.DeterminationDefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusNeutrum, detag.DeterminationIndefinite, omit)] = struct{}{}
			} else {
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusMaskulinum, reading.GetDetermination(), omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusFemininum, reading.GetDetermination(), omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), detag.GenusNeutrum, reading.GetDetermination(), omit)] = struct{}{}
			}
		} else {
			// Java: if (reading.getDetermination() == null || "jed".equals(lemma) || "manch".equals(lemma))
			lemma := tmpReading.GetLemma()
			jedOrManch := lemma != nil && (*lemma == "jed" || *lemma == "manch")
			if reading.GetDetermination() == "" || jedOrManch {
				// "jeder" etc. needs a special case to avoid false alarm
				set[makeString(reading.GetCasus(), reading.GetNumerus(), reading.GetGenus(), detag.DeterminationDefinite, omit)] = struct{}{}
				set[makeString(reading.GetCasus(), reading.GetNumerus(), reading.GetGenus(), detag.DeterminationIndefinite, omit)] = struct{}{}
			} else {
				set[makeString(reading.GetCasus(), reading.GetNumerus(), reading.GetGenus(), reading.GetDetermination(), omit)] = struct{}{}
			}
		}
	}
	return set
}

// GetAgreementSOLCategories ports AgreementTools.getAgreementSOLCategories.
func GetAgreementSOLCategories(aToken *languagetool.AnalyzedTokenReadings, omit map[GrammarCategory]bool) map[string]struct{} {
	set := map[string]struct{}{}
	if aToken == nil {
		return set
	}
	if omit == nil {
		omit = map[GrammarCategory]bool{}
	}
	for _, tmpReading := range aToken.GetReadings() {
		if tmpReading == nil {
			continue
		}
		// Java: if (tmpReading.getPOSTag() != null && !tmpReading.getPOSTag().endsWith(":SOL")) continue;
		if pos := tmpReading.GetPOSTag(); pos != nil && !strings.HasSuffix(*pos, ":SOL") {
			continue
		}
		reading := detag.NewAnalyzedGermanToken(tmpReading)
		if reading.GetCasus() == "" && reading.GetNumerus() == "" && reading.GetGenus() == "" {
			continue
		}
		set[makeString(reading.GetCasus(), reading.GetNumerus(), reading.GetGenus(), detag.DeterminationDefinite, omit)] = struct{}{}
		set[makeString(reading.GetCasus(), reading.GetNumerus(), reading.GetGenus(), detag.DeterminationIndefinite, omit)] = struct{}{}
	}
	return set
}

// possessiveSpecialCase ports AgreementTools.possessiveSpecialCase (private).
// would cause error misses as it contains 'ALG', e.g. in "Der Zustand meiner Gehirns."
func possessiveSpecialCase(aToken *languagetool.AnalyzedTokenReadings, tmpReading *languagetool.AnalyzedToken) bool {
	// Java: return aToken.hasPosTagStartingWith("PRO:POS") && StringUtils.equalsAny(tmpReading.getLemma(), "ich", "sich");
	if aToken == nil || !aToken.HasPosTagStartingWith("PRO:POS") {
		return false
	}
	if tmpReading == nil {
		return false
	}
	lem := tmpReading.GetLemma()
	if lem == nil {
		// StringUtils.equalsAny(null, ...) → false
		return false
	}
	return *lem == "ich" || *lem == "sich"
}

// makeString ports AgreementTools.makeString (private).
// Java joins non-omitted Kasus/Numerus/Genus + determination via '/';
// toString() values are German names (Nominativ, Singular, Maskulinum, definit, …).
func makeString(casus detag.Kasus, num detag.Numerus, gen detag.Genus, determination detag.Determination, omit map[GrammarCategory]bool) string {
	var l []string
	// Java: if (casus != null && !omit.contains(GrammarCategory.KASUS))
	if casus != "" && !omit[CatKasus] {
		l = append(l, string(casus))
	}
	if num != "" && !omit[CatNumerus] {
		l = append(l, string(num))
	}
	if gen != "" && !omit[CatGenus] {
		l = append(l, string(gen))
	}
	if determination != "" {
		l = append(l, string(determination))
	}
	return strings.Join(l, "/")
}
