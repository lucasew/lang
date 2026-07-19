package fr

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PostponedAdjectiveConcordanceFilter ports
// org.languagetool.rules.fr.PostponedAdjectiveConcordanceFilter (1:1).
//
// Synthesize ports FrenchSynthesizer.synthesize(token, posTag, true);
// when nil, empty suggestions → return nil.
type PostponedAdjectiveConcordanceFilter struct {
	maxLevels int
	// Synthesize ports FrenchSynthesizer.synthesize(AnalyzedToken, posTagRegex, true).
	Synthesize func(tok *languagetool.AnalyzedToken, posTagRegex string) []string

	adverbAppeared      bool
	conjunctionAppeared bool
	punctuationAppeared bool
	infinitiveAppeared  bool
}

func NewPostponedAdjectiveConcordanceFilter() *PostponedAdjectiveConcordanceFilter {
	return &PostponedAdjectiveConcordanceFilter{maxLevels: 4}
}

// French FreeLing space-separated POS patterns — Java Pattern.compile strings.
var (
	frNOM        = regexp.MustCompile(`[NZ] .*`)
	frNOMMS      = regexp.MustCompile(`[NZ] m s`)
	frNOMFS      = regexp.MustCompile(`[NZ] f s`)
	frNOMMP      = regexp.MustCompile(`[NZ] m p`)
	frNOMMN      = regexp.MustCompile(`[NZ] m sp`)
	frNOMFP      = regexp.MustCompile(`[NZ] f p`)
	frNOMCS      = regexp.MustCompile(`[NZ] e s`)
	frNOMCP      = regexp.MustCompile(`[NZ] e sp`)
	frNOMDET     = regexp.MustCompile(`[NZ] .*|(P\+)?D .*`)
	frGN         = regexp.MustCompile(`_GN_.*`)
	frGNMS       = regexp.MustCompile(`_GN_MS`)
	frGNFS       = regexp.MustCompile(`_GN_FS`)
	frGNMP       = regexp.MustCompile(`_GN_MP`)
	frGNFP       = regexp.MustCompile(`_GN_FP`)
	frGNCS       = regexp.MustCompile(`_GN_[MF]S`)
	frGNCP       = regexp.MustCompile(`_GN_[MF]P`)
	frGNMN       = regexp.MustCompile(`_GN_M[SP]`)
	frGNFN       = regexp.MustCompile(`_GN_F[SP]`)
	frDET        = regexp.MustCompile(`(P\+)?D .*`)
	frDETCS      = regexp.MustCompile(`(P\+)?D e s`)
	frDETMS      = regexp.MustCompile(`(P\+)?D m s`)
	frDETFS      = regexp.MustCompile(`(P\+)?D f s`)
	frDETMP      = regexp.MustCompile(`(P\+)?D m p`)
	frDETFP      = regexp.MustCompile(`(P\+)?D f p`)
	frDETCP      = regexp.MustCompile(`(P\+)?D e p`)
	frGNMSSub    = regexp.MustCompile(`[NZ] [me] (s|sp)|J [me] (s|sp)|V ppa m s|(P\+)?D m (s|sp)`)
	frGNFSSub    = regexp.MustCompile(`[NZ] [fe] (s|sp)|J [fe] (s|sp)|V ppa f s|(P\+)?D f (s|sp)`)
	frGNMPSub    = regexp.MustCompile(`[NZ] [me] (p|sp)|J [me] (p|sp)|V ppa m p|(P\+)?D m (p|sp)`)
	frGNFPSub    = regexp.MustCompile(`[NZ] [fe] (p|sp)|J [fe] (p|sp)|V ppa f p|(P\+)?D f (p|sp)`)
	frGNCPSub    = regexp.MustCompile(`[NZ] [fme] (p|sp)|J [fme] (p|sp)|(P\+)?D [fme] (p|sp)`)
	frGNCSSub    = regexp.MustCompile(`[NZ] [fme] (s|sp)|J [fme] (s|sp)|(P\+)?D [fme] (s|sp)`)
	frGNMNSub    = regexp.MustCompile(`[NZ] [me] (s|p|sp)|J [me] (s|p|sp)|(P\+)?D [me] (s|p|sp)`)
	frGNFNSub    = regexp.MustCompile(`[NZ] [fe] (s|p|sp)|J [fe] (s|p|sp)|(P\+)?D [fe] (s|p|sp)`)
	frADJECTIU   = regexp.MustCompile(`J .*|V ppa .*|PX.*`)
	frADJECTIUMS = regexp.MustCompile(`J [me] (s|sp)|V ppa m s`)
	frADJECTIUFS = regexp.MustCompile(`J [fe] (s|sp)|V ppa f s`)
	frADJECTIUMP = regexp.MustCompile(`J [me] (p|sp)|V ppa m p`)
	frADJECTIUFP = regexp.MustCompile(`J [fe] (p|sp)|V ppa f p`)
	frADJECTIUCP = regexp.MustCompile(`J e (p|sp)`)
	frADJECTIUCS = regexp.MustCompile(`J e (s|sp)`)
	frADJECTIUMN = regexp.MustCompile(`J m sp`)
	frADJECTIUFN = regexp.MustCompile(`J f sp`)
	frADJECTIUS  = regexp.MustCompile(`J .* (s|sp)|V ppa . s`)
	frADJECTIUP  = regexp.MustCompile(`J .* (p|sp)|V ppa . p`)
	frADJECTIUM  = regexp.MustCompile(`J [me] .*|V ppa [me] .*`)
	// ADJECTIU_F is declared in Java but only ADJECTIU_FN is used as adjPattern for FN branch.
	frADVERBI    = regexp.MustCompile(`A`)
	frCONJUNCIO  = regexp.MustCompile(`C .*`)
	frPUNTUACIO  = regexp.MustCompile(`_PUNCT`)
	frLOCADV     = regexp.MustCompile(`A`)
	frADVERBISOK = regexp.MustCompile(`A`)
	frCOORDIONI  = regexp.MustCompile(`et|ou|ni`)
	frKEEPCOUNT  = regexp.MustCompile(`Y|J .*|N .*|D .*|P.*|V ppa .*|M nonfin|UNKNOWN|Z.*|V.* inf|V ppr`)
	frKEEPCOUNT2 = regexp.MustCompile(`,|et|ou|ni`)
	frSTOPCOUNT  = regexp.MustCompile(`[;:\(\)\[\]–—―‒]`)
	frPREPOS     = regexp.MustCompile(`P.*`)
	frPREPNIVEL  = regexp.MustCompile(`d'|de|des|du|à|au|aux|en|dans|sur|entre|par|pour|avec|sans|contre|comme`)
	frVERB       = regexp.MustCompile(`V.* (inf|ind|sub|con|ppr|imp).*`)
	frINFINITIVE = regexp.MustCompile(`V.* inf`)
	frGV         = regexp.MustCompile(`_GV_`)
)

// AcceptRuleMatch ports PostponedAdjectiveConcordanceFilter.acceptRuleMatch.
func (f *PostponedAdjectiveConcordanceFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	maxLevels := f.maxLevels
	if maxLevels <= 0 {
		maxLevels = 4
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	i := patternTokenPos
	if i < 0 || i >= len(tokens) || tokens[i] == nil {
		return nil
	}

	isPlural := true
	isPrevNoun := false
	var substPattern, gnPattern, adjPattern *regexp.Regexp
	canBeMS, canBeFS, canBeMP, canBeFP, canBeP := false, false, false, false, false

	cNt := make([]int, maxLevels)
	cNMS := make([]int, maxLevels)
	cNFS := make([]int, maxLevels)
	cNMP := make([]int, maxLevels)
	cNMN := make([]int, maxLevels)
	cNFP := make([]int, maxLevels)
	cNCS := make([]int, maxLevels)
	cNCP := make([]int, maxLevels)
	cDMS := make([]int, maxLevels)
	cDFS := make([]int, maxLevels)
	cDMP := make([]int, maxLevels)
	cDFP := make([]int, maxLevels)
	cN := make([]int, maxLevels)
	cD := make([]int, maxLevels)
	level := 0
	j := 1
	f.initializeApparitions()

	for i-j > 0 && f.keepCounting(tokens[i-j]) && level < maxLevels {
		if !isPrevNoun {
			if frMatchPostagRegexp(tokens[i-j], frNOM) ||
				(i-j-1 > 0 && !frMatchPostagRegexp(tokens[i-j], frNOM) && frMatchPostagRegexp(tokens[i-j], frADJECTIU) &&
					frMatchPostagRegexp(tokens[i-j-1], frDET)) {
				if frMatchPostagRegexp(tokens[i-j], frGNMS) {
					cNMS[level]++
					canBeMS = true
				}
				if frMatchPostagRegexp(tokens[i-j], frGNFS) {
					cNFS[level]++
					canBeFS = true
				}
				if frMatchPostagRegexp(tokens[i-j], frGNMP) {
					cNMP[level]++
					canBeMP = true
				}
				if frMatchPostagRegexp(tokens[i-j], frGNFP) {
					cNFP[level]++
					canBeFP = true
				}
			}
			if !frMatchPostagRegexp(tokens[i-j], frGN) {
				if frMatchPostagRegexp(tokens[i-j], frNOMMS) {
					cNMS[level]++
					canBeMS = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMFS) {
					cNFS[level]++
					canBeFS = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMMP) {
					cNMP[level]++
					canBeMP = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMMN) {
					cNMN[level]++
					canBeMS = true
					canBeMP = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMFP) {
					cNFP[level]++
					canBeFP = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMCS) {
					cNCS[level]++
					canBeMS = true
					canBeFS = true
				} else if frMatchPostagRegexp(tokens[i-j], frNOMCP) {
					cNCP[level]++
					canBeFP = true
					canBeMP = true
				}
			}
		}
		if frMatchPostagRegexp(tokens[i-j], frNOM) {
			cNt[level]++
			isPrevNoun = true
		} else {
			isPrevNoun = false
		}

		if frMatchPostagRegexp(tokens[i-j], frDETCS) {
			if i-j+1 < len(tokens) && frMatchPostagRegexp(tokens[i-j+1], frNOMMS) {
				cDMS[level]++
				canBeMS = true
			}
			if i-j+1 < len(tokens) && frMatchPostagRegexp(tokens[i-j+1], frNOMFS) {
				cDFS[level]++
				canBeFS = true
			}
		}
		if frMatchPostagRegexp(tokens[i-j], frDETCP) {
			// Java assigns canBeMP/FP into cDMS/cDFS counters (bug-for-bug).
			if i-j+1 < len(tokens) && frMatchPostagRegexp(tokens[i-j+1], frNOMMP) {
				cDMS[level]++
				canBeMP = true
			}
			if i-j+1 < len(tokens) && frMatchPostagRegexp(tokens[i-j+1], frNOMFP) {
				cDFS[level]++
				canBeFP = true
			}
		}
		if !frMatchPostagRegexp(tokens[i-j], frADVERBI) {
			if frMatchPostagRegexp(tokens[i-j], frDETMS) {
				cDMS[level]++
				canBeMS = true
			}
			if frMatchPostagRegexp(tokens[i-j], frDETFS) {
				cDFS[level]++
				canBeFS = true
			}
			if frMatchPostagRegexp(tokens[i-j], frDETMP) {
				cDMP[level]++
				canBeMP = true
			}
			if frMatchPostagRegexp(tokens[i-j], frDETFP) {
				cDFP[level]++
				canBeFP = true
			}
		}
		if i-j-1 > 0 {
			if frMatchRegexp(tokens[i-j].GetToken(), frPREPNIVEL) &&
				frMatchPostagRegexp(tokens[i-j], frPREPOS) &&
				!frMatchPostagRegexp(tokens[i-j], frCONJUNCIO) &&
				!frMatchRegexp(tokens[i-j-1].GetToken(), frCOORDIONI) &&
				i-j+1 < len(tokens) && !frMatchPostagRegexp(tokens[i-j+1], frADVERBI) {
				level++
			} else if i-j+1 < len(tokens) &&
				strings.EqualFold(tokens[i-j].GetToken(), "d'") &&
				strings.EqualFold(tokens[i-j+1].GetToken(), "environ") {
				level++
			}
		}
		j = f.updateJValue(tokens, i, j, level)
		if i-j >= 0 && i-j < len(tokens) {
			f.updateApparitions(tokens[i-j])
		}
		j++
	}
	level++
	if level > maxLevels {
		level = maxLevels
	}
	j = 0
	cNtotal := 0
	cDtotal := 0
	for j < level {
		cN[j] = cNMS[j] + cNFS[j] + cNMP[j] + cNFP[j] + cNCS[j] + cNCP[j] + cNMN[j]
		cD[j] = cDMS[j] + cDFS[j] + cDMP[j] + cDFP[j]
		cNtotal += cN[j]
		cDtotal += cD[j]

		if frMatchPostagRegexp(tokens[i], frADJECTIUMP) && (cN[j] > 1 || cD[j] > 1) &&
			(cNMS[j]+cNMN[j]+cNMP[j]+cNCS[j]+cNCP[j]+cDMS[j]+cDMP[j]) > 0 &&
			(cNFS[j]+cNFP[j] <= cNt[j]) {
			return nil
		}
		if frMatchPostagRegexp(tokens[i], frADJECTIUFP) && (cN[j] > 1 || cD[j] > 1) &&
			((cNMS[j]+cNMP[j]+cNMN[j]+cDMS[j]+cDMP[j]) == 0 || (cNt[j] > 0 && cNFS[j]+cNFP[j] >= cNt[j])) {
			return nil
		}
		// FR: isPlural = isPlural && cD[j] > 1 && level>1
		if cN[j]+cD[j] > 0 {
			isPlural = isPlural && cD[j] > 1 && level > 1
			canBeP = canBeP || cN[j] > 1
		}
		j++
	}
	isPlural = isPlural || (i-2 > 0 && cNMP[0]+cNFP[0]+cNCP[0] > 0 && tokens[i-2].GetToken() == ",")

	if cNtotal == 0 && cDtotal == 0 {
		return nil
	}

	if frMatchPostagRegexp(tokens[i], frADJECTIUCS) {
		substPattern = frGNCSSub
		adjPattern = frADJECTIUS
		gnPattern = frGNCS
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUCP) {
		substPattern = frGNCPSub
		adjPattern = frADJECTIUP
		gnPattern = frGNCP
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUMN) {
		substPattern = frGNMNSub
		adjPattern = frADJECTIUM
		gnPattern = frGNMN
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUFN) {
		substPattern = frGNFNSub
		adjPattern = frADJECTIUFN
		gnPattern = frGNFN
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUMS) {
		substPattern = frGNMSSub
		adjPattern = frADJECTIUMS
		gnPattern = frGNMS
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUFS) {
		substPattern = frGNFSSub
		adjPattern = frADJECTIUFS
		gnPattern = frGNFS
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUMP) {
		substPattern = frGNMPSub
		adjPattern = frADJECTIUMP
		gnPattern = frGNMP
	} else if frMatchPostagRegexp(tokens[i], frADJECTIUFP) {
		substPattern = frGNFPSub
		adjPattern = frADJECTIUFP
		gnPattern = frGNFP
	}

	if substPattern == nil || gnPattern == nil || adjPattern == nil {
		return nil
	}

	j = 1
	keepCount := true
	for i-j > 0 && keepCount {
		if frMatchPostagRegexp(tokens[i-j], frNOMDET) && frMatchPostagRegexp(tokens[i-j], gnPattern) {
			return nil
		} else if !frMatchPostagRegexp(tokens[i-j], frGN) && frMatchPostagRegexp(tokens[i-j], substPattern) {
			return nil
		}
		keepCount = !frMatchPostagRegexp(tokens[i-j], frNOMDET)
		j++
	}

	if i-1 < 0 || tokens[i-1] == nil {
		return nil
	}
	// FR necessary condition: NOM, _GN_, ADJECTIU, or adverbs
	cond := (frMatchPostagRegexp(tokens[i-1], frNOM) && !frMatchPostagRegexp(tokens[i-1], substPattern)) ||
		(frMatchPostagRegexp(tokens[i-1], frGN) && !frMatchPostagRegexp(tokens[i-1], gnPattern)) ||
		(frMatchPostagRegexp(tokens[i-1], frADJECTIU) && !frMatchPostagRegexp(tokens[i-1], adjPattern)) ||
		(i > 2 && frMatchPostagRegexp(tokens[i-1], frADVERBISOK) && !frMatchPostagRegexp(tokens[i-2], frVERB) &&
			!frMatchPostagRegexp(tokens[i-2], frPREPOS)) ||
		(i > 3 && frMatchPostagRegexp(tokens[i-1], frLOCADV) && frMatchPostagRegexp(tokens[i-2], frLOCADV) &&
			!frMatchPostagRegexp(tokens[i-3], frVERB) && !frMatchPostagRegexp(tokens[i-3], frPREPOS))
	if !cond {
		return nil
	}

	if !(isPlural && frMatchPostagRegexp(tokens[i], frADJECTIUS)) {
		j = 1
		f.initializeApparitions()
		for i-j > 0 && f.keepCounting(tokens[i-j]) && (level > 1 || j < 4) {
			if !frMatchPostagRegexp(tokens[i-j], frGN) && frMatchPostagRegexp(tokens[i-j], frNOMDET) &&
				frMatchPostagRegexp(tokens[i-j], substPattern) {
				return nil
			} else if frMatchPostagRegexp(tokens[i-j], gnPattern) {
				return nil
			}
			j = f.updateJValue(tokens, i, j, 0)
			if i-j >= 0 && i-j < len(tokens) {
				f.updateApparitions(tokens[i-j])
			}
			j++
		}
	}

	var suggestions []string
	at := frGetAnalyzedToken(tokens[patternTokenPos], frADJECTIUCS)
	if at != nil {
		suggestions = append(suggestions, f.synth(at, "J e p")...)
	}
	if len(suggestions) == 0 {
		at = frGetAnalyzedToken(tokens[patternTokenPos], frADJECTIUCP)
		if at != nil {
			suggestions = append(suggestions, f.synth(at, "J e s")...)
		}
	}
	if len(suggestions) == 0 && isPlural {
		at = frGetAnalyzedToken(tokens[patternTokenPos], frADJECTIUP)
		if at != nil {
			suggestions = append(suggestions, f.synth(at, "J . p|V ppa . p")...)
		}
	}
	at = frGetAnalyzedToken(tokens[patternTokenPos], frADJECTIU)
	if at != nil && len(suggestions) == 0 {
		if canBeMS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "J [me] sp?|V ppa m s")...)
		}
		if canBeFS && !isPlural {
			suggestions = append(suggestions, f.synth(at, "J [fe] sp?|V ppa f s")...)
		}
		if canBeMP {
			suggestions = append(suggestions, f.synth(at, "J [me] s?p|V ppa m p")...)
		}
		if canBeFP {
			suggestions = append(suggestions, f.synth(at, "J [fe] s?p|V ppa f p")...)
		}
		if canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "J [me] s?p|V ppa m p")...)
		}
		if canBeFS && !canBeMS && (isPlural || canBeP) {
			suggestions = append(suggestions, f.synth(at, "J [fe] s?p|V ppa f p")...)
		}
	}

	// set suggestion removing duplicates (before removing original — Java order)
	suggestions = frDistinct(suggestions)
	origLower := strings.ToLower(tokens[patternTokenPos].GetToken())
	filtered := suggestions[:0]
	removed := false
	for _, s := range suggestions {
		if !removed && s == origLower {
			removed = true
			continue
		}
		filtered = append(filtered, s)
	}
	suggestions = filtered
	if len(suggestions) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(suggestions)
	return match
}

func (f *PostponedAdjectiveConcordanceFilter) synth(at *languagetool.AnalyzedToken, posTagRE string) []string {
	if f == nil || f.Synthesize == nil || at == nil {
		return nil
	}
	return f.Synthesize(at, posTagRE)
}

// updateJValue ports "deux ou plus".
func (f *PostponedAdjectiveConcordanceFilter) updateJValue(tokens []*languagetool.AnalyzedTokenReadings, i, j, level int) int {
	_ = f
	_ = level
	if frMatchRegexp(tokens[i-j].GetToken(), frCOORDIONI) {
		if i-j-1 > 0 && i-j+1 < len(tokens) {
			if frMatchPostagRegexp(tokens[i-j-1], frDET) && tokens[i-j+1].GetToken() == "plus" {
				j = j + 1
			}
		}
	}
	return j
}

func (f *PostponedAdjectiveConcordanceFilter) keepCounting(aTr *languagetool.AnalyzedTokenReadings) bool {
	if aTr == nil {
		return false
	}
	if frMatchRegexp(aTr.GetToken(), frPREPNIVEL) {
		return true
	}
	if aTr.GetToken() == "." {
		return true
	}
	if (f.adverbAppeared && f.conjunctionAppeared) || (f.adverbAppeared && f.punctuationAppeared) ||
		(f.conjunctionAppeared && f.punctuationAppeared) || (f.punctuationAppeared && frMatchPostagRegexp(aTr, frPUNTUACIO)) ||
		(f.infinitiveAppeared && frMatchRegexp(aTr.GetToken(), frCOORDIONI)) ||
		(f.infinitiveAppeared && f.adverbAppeared) {
		return false
	}
	return (frMatchPostagRegexp(aTr, frKEEPCOUNT) || frMatchRegexp(aTr.GetToken(), frKEEPCOUNT2) ||
		frMatchPostagRegexp(aTr, frADVERBISOK)) && !frMatchRegexp(aTr.GetToken(), frSTOPCOUNT) &&
		(!frMatchPostagRegexp(aTr, frGV) || frMatchPostagRegexp(aTr, frGN))
}

func (f *PostponedAdjectiveConcordanceFilter) initializeApparitions() {
	f.adverbAppeared = false
	f.conjunctionAppeared = false
	f.punctuationAppeared = false
	f.infinitiveAppeared = false
}

func (f *PostponedAdjectiveConcordanceFilter) updateApparitions(aTr *languagetool.AnalyzedTokenReadings) {
	if aTr == nil {
		return
	}
	f.conjunctionAppeared = f.conjunctionAppeared || frMatchPostagRegexp(aTr, frCONJUNCIO)
	if aTr.GetToken() == "com" {
		return
	}
	if frMatchPostagRegexp(aTr, frNOM) || frMatchPostagRegexp(aTr, frADJECTIU) {
		f.initializeApparitions()
		return
	}
	f.adverbAppeared = f.adverbAppeared || frMatchPostagRegexp(aTr, frADVERBI)
	f.punctuationAppeared = f.punctuationAppeared || (frMatchPostagRegexp(aTr, frPUNTUACIO) || aTr.GetToken() == ",")
	f.infinitiveAppeared = f.infinitiveAppeared || frMatchPostagRegexp(aTr, frINFINITIVE)
}

// frFullMatch is Java Matcher.matches() (entire string).
func frFullMatch(pattern *regexp.Regexp, s string) bool {
	if pattern == nil {
		return false
	}
	loc := pattern.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func frMatchPostagRegexp(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) bool {
	if aToken == nil || pattern == nil {
		return false
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if frFullMatch(pattern, posTag) {
			return true
		}
	}
	return false
}

func frMatchRegexp(s string, pattern *regexp.Regexp) bool {
	return frFullMatch(pattern, s)
}

func frGetAnalyzedToken(aToken *languagetool.AnalyzedTokenReadings, pattern *regexp.Regexp) *languagetool.AnalyzedToken {
	if aToken == nil || pattern == nil {
		return nil
	}
	for _, analyzedToken := range aToken.GetReadings() {
		posTag := "UNKNOWN"
		if pt := analyzedToken.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if frFullMatch(pattern, posTag) {
			return analyzedToken
		}
	}
	return nil
}

func frDistinct(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
