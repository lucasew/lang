package pt

import (
	"embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official confusion_pairs resources Java loads via JLanguageTool.getDataBroker().getFromRulesDirAsStream:
//
//	/pt/confusion_pairs.txt
//	/pt/pt-BR/confusion_pairs.txt
//	/pt/pt-PT/confusion_pairs.txt
//
//go:embed data/confusion_pairs.txt data/pt-BR/confusion_pairs.txt data/pt-PT/confusion_pairs.txt
var confusionFS embed.FS

// Gender/number POS patterns — ports ConfusionCheckFilter MS/FS/MP/FP/CP/CS.
// Java Pattern.compile(...).matcher(tag).matches() ⇒ full-string match.
var (
	confusionPatternMS = regexp.MustCompile(`^(?:NC[MC][SN]000|A..[MC][SN].|V.P..SM)$`)
	confusionPatternFS = regexp.MustCompile(`^(?:NC[FC][SN]000|A..[FC][SN].|V.P..SF)$`)
	confusionPatternMP = regexp.MustCompile(`^(?:NC[MC][PN]000|A..[MC][PN].|V.P..PM)$`)
	confusionPatternFP = regexp.MustCompile(`^(?:NC[FC][PN]000|A..[FC][PN].|V.P..PF)$`)
	confusionPatternCP = regexp.MustCompile(`^(?:NC[MFC][PN]000|A..[MFC][PN].|V.P..P.)$`)
	confusionPatternCS = regexp.MustCompile(`^(?:NC[MFC][SN]000|A..[MFC][SN].|V.P..S.)$`)
)

// Classpath path → embed path under this package.
var confusionEmbedByClasspath = map[string]string{
	"/pt/confusion_pairs.txt":       "data/confusion_pairs.txt",
	"/pt/pt-BR/confusion_pairs.txt": "data/pt-BR/confusion_pairs.txt",
	"/pt/pt-PT/confusion_pairs.txt": "data/pt-PT/confusion_pairs.txt",
}

var (
	relevantWordsPTOnce sync.Once
	relevantWordsPT     map[string]*languagetool.AnalyzedTokenReadings
	relevantWordsBROnce sync.Once
	relevantWordsBR     map[string]*languagetool.AnalyzedTokenReadings
)

// ConfusionCheckFilter ports org.languagetool.rules.pt.ConfusionCheckFilter.
type ConfusionCheckFilter struct {
	// relevantWordsPT / relevantWordsBR twin Java final instance fields
	// (loaded at construction via ConfusionPairsDataLoader.loadWords).
	relevantWordsPT map[string]*languagetool.AnalyzedTokenReadings
	relevantWordsBR map[string]*languagetool.AnalyzedTokenReadings
}

// NewConfusionCheckFilter ports the Java no-arg constructor field initializers.
func NewConfusionCheckFilter() *ConfusionCheckFilter {
	return &ConfusionCheckFilter{
		relevantWordsPT: loadRelevantWordsPT(),
		relevantWordsBR: loadRelevantWordsBR(),
	}
}

func loadRelevantWordsPT() map[string]*languagetool.AnalyzedTokenReadings {
	relevantWordsPTOnce.Do(func() {
		relevantWordsPT = loadConfusionWordsMerged(getFilepathsPT())
	})
	return relevantWordsPT
}

func loadRelevantWordsBR() map[string]*languagetool.AnalyzedTokenReadings {
	relevantWordsBROnce.Do(func() {
		relevantWordsBR = loadConfusionWordsMerged(getFilepathsBR())
	})
	return relevantWordsBR
}

// getFilepathsBR ports ConfusionCheckFilter.getFilepathsBR.
func getFilepathsBR() []string {
	return []string{
		"/pt/confusion_pairs.txt",
		"/pt/pt-BR/confusion_pairs.txt",
	}
}

// getFilepathsPT ports ConfusionCheckFilter.getFilepathsPT.
func getFilepathsPT() []string {
	return []string{
		"/pt/confusion_pairs.txt",
		"/pt/pt-PT/confusion_pairs.txt",
	}
}

// loadConfusionWordsMerged ports ConfusionPairsDataLoader.loadWords(List<String>):
// multi-file HashMap merge with addReading for keys already present.
func loadConfusionWordsMerged(classpaths []string) map[string]*languagetool.AnalyzedTokenReadings {
	loader := ConfusionPairsDataLoader{}
	out := map[string]*languagetool.AnalyzedTokenReadings{}
	for _, cp := range classpaths {
		embedPath, ok := confusionEmbedByClasspath[cp]
		if !ok {
			panic("ConfusionCheckFilter: unknown confusion_pairs classpath: " + cp)
		}
		f, err := confusionFS.Open(embedPath)
		if err != nil {
			panic(err)
		}
		part, err := loader.LoadWords(f, cp)
		_ = f.Close()
		if err != nil {
			panic(err)
		}
		for key, atr := range part {
			existing, ok := out[key]
			if !ok {
				out[key] = atr
				continue
			}
			// Java: map.containsKey → atrs.addReading(analyzedToken, "") per new line.
			// Go loader already folded multi-line keys inside one file; fold file→map here.
			for _, reading := range atr.GetReadings() {
				if reading == nil {
					continue
				}
				existing.AddReading(reading, "")
			}
		}
	}
	return out
}

// getLanguageVariantCode ports ConfusionCheckFilter.getLanguageVariantCode:
// ((PatternRule) match.getRule()).getLanguage().getShortCodeWithCountryAndVariant().
func getLanguageVariantCode(match *rules.RuleMatch) string {
	rule := match.GetRule()
	if pr, ok := rule.(*patterns.PatternRule); ok {
		return pr.GetLanguageCode()
	}
	if r, ok := rule.(interface{ GetLanguageCode() string }); ok {
		return r.GetLanguageCode()
	}
	// Java ClassCastException when rule is not a PatternRule.
	panic(fmt.Sprintf("ConfusionCheckFilter: rule is not PatternRule: %T", rule))
}

func confusionRuleFullID(rule any) string {
	if r, ok := rule.(interface{ GetFullId() string }); ok {
		return r.GetFullId()
	}
	if pr, ok := rule.(*patterns.PatternRule); ok {
		if pr.SubID != "" {
			return pr.ID + "[" + pr.SubID + "]"
		}
		return pr.ID
	}
	if r, ok := rule.(interface{ GetID() string }); ok {
		return r.GetID()
	}
	return ""
}

// AcceptRuleMatch ports ConfusionCheckFilter.acceptRuleMatch.
func (f *ConfusionCheckFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}

	var desiredGenderNumberPattern *regexp.Regexp
	var replacement *string // Java String replacement = null
	postag := patterns.GetRequired("postag", arguments)
	form := strings.ToLower(patterns.GetRequired("form", arguments))
	gendernumberFrom := patterns.GetOptional("gendernumberFrom", arguments)
	// Java getOptional returns null when absent; Go returns "". Treat "" as absent.
	if gendernumberFrom != "" {
		i, err := strconv.Atoi(gendernumberFrom)
		if err != nil {
			// Java Integer.parseInt throws NumberFormatException.
			panic(err)
		}
		if i < 1 || i > len(patternTokens) {
			panic(fmt.Sprintf(
				"ConfusionCheckFilter: Index out of bounds in %s, value: %d",
				confusionRuleFullID(match.GetRule()), i))
		}
		atr := patternTokens[i-1]
		// Java if/else if order for gender/number probes — bug-for-bug.
		if atr.MatchesPosTagRegex(`[NAPD].+MS.*|V.P..SM`) {
			desiredGenderNumberPattern = confusionPatternMS
		} else if atr.MatchesPosTagRegex(`[NAPD].+MP.*|V.P..PM`) {
			desiredGenderNumberPattern = confusionPatternMP
		} else if atr.MatchesPosTagRegex(`[NAPD].+FS.*|V.P..SF`) {
			desiredGenderNumberPattern = confusionPatternFS
		} else if atr.MatchesPosTagRegex(`[NAPD].+FP.*|V.P..PF`) {
			desiredGenderNumberPattern = confusionPatternFP
		} else if atr.MatchesPosTagRegex(`[NAPD].+CP.*|V.P..P.`) {
			desiredGenderNumberPattern = confusionPatternCP
		} else if atr.MatchesPosTagRegex(`[NAPD].+CS.*|V.P..S.`) {
			desiredGenderNumberPattern = confusionPatternCS
		}
	}

	var relevantWords map[string]*languagetool.AnalyzedTokenReadings
	variant := getLanguageVariantCode(match)
	if variant == "pt-BR" {
		relevantWords = f.relevantWordsBR
	} else if variant == "pt-PT" {
		relevantWords = f.relevantWordsPT
	} else {
		// Africa and elsewhere: PT spelling dominant (Java comment).
		relevantWords = f.relevantWordsPT
	}

	if atr, ok := relevantWords[form]; ok {
		if atr.MatchesPosTagRegex(postag) {
			if desiredGenderNumberPattern != nil {
				// Java: matcher on first reading POSTag only.
				readings := atr.GetReadings()
				firstPOS := readings[0].GetPOSTag()
				if firstPOS == nil || !desiredGenderNumberPattern.MatchString(*firstPOS) {
					return nil
				}
				tok := atr.GetToken()
				replacement = &tok
			} else if gendernumberFrom == "" {
				// no desired gender number defined
				tok := atr.GetToken()
				replacement = &tok
			}
			// Java quirk: gendernumberFrom set but no POS branch matched →
			// desiredGenderNumberPattern stays null and replacement stays null.
		}
	}

	if replacement != nil {
		repl := *replacement
		message := match.GetMessage()
		// Change the message if the replacement has no diacritic gain over form.
		if !(tools.HasDiacritics(repl) && !tools.HasDiacritics(form)) {
			message = strings.ReplaceAll(message, "se escribe con tilde", "se escribe de otra manera")
		}
		ruleMatch := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), message)
		ruleMatch.SetShortMessage(match.GetShortMessage())
		ruleMatch.SetType(match.GetType())
		// Java: getSuggestedReplacements().get(0) — empty list is IndexOutOfBoundsException.
		suggestion := match.GetSuggestedReplacements()[0]
		suggestion = strings.ReplaceAll(suggestion, "{suggestion}", repl)
		suggestion = strings.ReplaceAll(suggestion, "{Suggestion}", tools.UppercaseFirstChar(repl))
		suggestion = strings.ReplaceAll(suggestion, "{SUGGESTION}", strings.ToUpper(repl))
		ruleMatch.SetSuggestedReplacement(suggestion)
		return ruleMatch
	}
	return nil
}
