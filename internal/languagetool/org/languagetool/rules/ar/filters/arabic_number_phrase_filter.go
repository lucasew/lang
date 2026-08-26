package filters

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicNumberPhraseFilter ports org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter.
// Filter that maps suggestion for numeric phrases.
type ArabicNumberPhraseFilter struct {
	// synthesizer ports private final ArabicSynthesizer synthesizer = new ArabicSynthesizer(new Arabic()).
	synthesizer *ar_synth.ArabicSynthesizer
	// tagmanager ports private static final ArabicTagManager tagmanager.
	tagmanager *ar_tag.ArabicTagManager
}

// NewArabicNumberPhraseFilter ports the Java ctor.
func NewArabicNumberPhraseFilter() *ArabicNumberPhraseFilter {
	return &ArabicNumberPhraseFilter{
		synthesizer: ar_synth.Instance(),
		tagmanager:  ar_tag.NewArabicTagManager(),
	}
}

// AcceptRuleMatch ports ArabicNumberPhraseFilter.acceptRuleMatch.
func (f *ArabicNumberPhraseFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	// get the previous word
	previousWord := arguments["previous"]
	if previousWord == "" {
		// Java getOrDefault("previous", "")
		previousWord = ""
	}
	// previous word index in token list
	previousWordPos := getPreviousPos(arguments)

	// get the inflect mark
	inflectArg := arguments["inflect"]
	// get the next word as units
	nextWord := arguments["next"]

	nextWordPos := getNextPos(arguments, len(patternTokens))

	// get all numeric tokens
	// Java: startPos = (previousWordPos > 0) ? previousWordPos + 1 : 0
	startPos := 0
	if previousWordPos > 0 {
		startPos = previousWordPos + 1
	}
	// Java: endPos = (nextWordPos > 0) ? Integer.min(nextWordPos, length) : length + nextWordPos
	// Note getNextPos already maps negative → size+n (non-negative absolute).
	endPos := len(patternTokens) + nextWordPos
	if nextWordPos > 0 {
		endPos = nextWordPos
		if endPos > len(patternTokens) {
			endPos = len(patternTokens)
		}
	}

	numWordTokens := make([]string, 0, endPos-startPos)
	for i := startPos; i < endPos && i < len(patternTokens); i++ {
		if patternTokens[i] == nil {
			// Java would NPE; skip soft — fail-closed empty token
			numWordTokens = append(numWordTokens, "")
			continue
		}
		// Java: patternTokens[i].getToken().trim()
		numWordTokens = append(numWordTokens, tools.JavaStringTrim(patternTokens[i].GetToken()))
	}

	numPhrase := strings.Join(numWordTokens, " ")
	// extract features from previous
	feminine := false
	attached := false
	inflection := getInflectedCase(patternTokens, previousWordPos, inflectArg)
	var suggestionList []string
	if nextWord == "" {
		suggestionList = PrepareSuggestion(numPhrase, previousWord, nil, feminine, attached, inflection)
	} else {
		var nextWordToken *languagetool.AnalyzedTokenReadings
		if endPos > 0 && endPos < len(patternTokens) {
			nextWordToken = patternTokens[endPos]
		}
		suggestionList = f.PrepareSuggestionWithUnits(numPhrase, previousWord, nextWordToken, feminine, attached, inflection)
	}

	newMatch := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	newMatch.ShortMessage = match.ShortMessage
	if len(suggestionList) > 0 {
		// Java: for (String sug : suggestionList) newMatch.addSuggestedReplacement(sug);
		newMatch.SetSuggestedReplacements(suggestionList)
	}
	return newMatch
}

// getInflectedCase ports getInflectedCase.
func getInflectedCase(patternTokens []*languagetool.AnalyzedTokenReadings, previousPos int, inflect string) string {
	// Java: if (!Objects.equals(inflect, "")) return inflect;
	if inflect != "" {
		return inflect
	}
	// if the previous is Jar
	if previousPos >= 0 && previousPos < len(patternTokens) && patternTokens[previousPos] != nil {
		for _, tk := range patternTokens[previousPos].GetReadings() {
			if tk == nil {
				continue
			}
			pos := tk.GetPOSTag()
			if pos != nil && strings.HasPrefix(*pos, "PR") {
				return "jar"
			}
		}
	}
	// Java: String firstWord = patternTokens[previousPos + 1].getToken();
	// No bounds check in Java — only when previousPos+1 in range in Go to avoid panic.
	if previousPos+1 >= 0 && previousPos+1 < len(patternTokens) && patternTokens[previousPos+1] != nil {
		firstWord := patternTokens[previousPos+1].GetToken()
		if strings.HasPrefix(firstWord, "ب") || strings.HasPrefix(firstWord, "ل") || strings.HasPrefix(firstWord, "ك") {
			return "jar"
		}
	}
	return ""
}

// getFemininCase ports getFemininCase (unused on accept path; kept as twin).
func (f *ArabicNumberPhraseFilter) getFemininCase(patternTokens []*languagetool.AnalyzedTokenReadings, nextPos int) bool {
	if f == nil || f.tagmanager == nil || nextPos < 0 || nextPos >= len(patternTokens) || patternTokens[nextPos] == nil {
		return false
	}
	for _, tk := range patternTokens[nextPos].GetReadings() {
		if tk == nil || tk.GetPOSTag() == nil {
			continue
		}
		if f.tagmanager.IsFeminin(*tk.GetPOSTag()) {
			return true
		}
	}
	return false
}

// PrepareSuggestion ports static prepareSuggestion.
// Java quirk: only adds previousWord+" "+sug when previousWord non-empty;
// if previous empty returns empty list even when tmp has items.
func PrepareSuggestion(numPhrase, previousWord string, nextWord *languagetool.AnalyzedTokenReadings, feminin, attached bool, inflection string) []string {
	_ = nextWord // Java param unused in body
	tmpsuggestionList := tools.GetSuggestionsNumericPhrase(numPhrase, feminin, attached, inflection)
	var suggestionList []string
	if len(tmpsuggestionList) > 0 {
		for _, sug := range tmpsuggestionList {
			if previousWord != "" {
				suggestionList = append(suggestionList, previousWord+" "+sug)
			}
		}
	}
	return suggestionList
}

// PrepareSuggestionWithUnits ports prepareSuggestionWithUnits.
func (f *ArabicNumberPhraseFilter) PrepareSuggestionWithUnits(numPhrase, previousWord string, nextWord *languagetool.AnalyzedTokenReadings, feminin, attached bool, inflection string) []string {
	defaultUnit := "دينار"

	tmpsuggestionList := tools.GetSuggestionsNumericPhraseWithUnits(numPhrase, defaultUnit, feminin, attached, inflection)
	var suggestionList []string
	if len(tmpsuggestionList) > 0 {
		for _, sugMap := range tmpsuggestionList {
			sug := sugMap["phrase"]
			inflectedUnitList := f.inflectUnit(nextWord, sugMap)
			// Java: for (String unit : inflectedUnitList) — NPE if null; treat nil as empty.
			for _, unit := range inflectedUnitList {
				var tmp strings.Builder
				if previousWord != "" {
					tmp.WriteString(previousWord + " ")
				}
				tmp.WriteString(sug)
				if unit != "" {
					tmp.WriteString(" " + unit)
				}
				suggestionList = append(suggestionList, tmp.String())
			}
		}
	}
	return suggestionList
}

// inflectUnit ports inflectUnit — synthesizer-backed unit inflection.
func (f *ArabicNumberPhraseFilter) inflectUnit(unit *languagetool.AnalyzedTokenReadings, sugMap map[string]string) []string {
	if unit == nil {
		return nil
	}
	if f == nil || f.tagmanager == nil {
		return nil
	}
	inflection := sugMap["unitInflection"]
	number := sugMap["unitNumber"]
	_ = unit.GetToken() + "{" + inflection + "+" + number + "}" // Java dead local `inflected`

	var tmpList []string
	var inflectedList []string
	for _, tk := range unit.GetReadings() {
		if tk == nil || tk.GetPOSTag() == nil {
			continue
		}
		postag := *tk.GetPOSTag()
		if f.tagmanager.IsNoun(postag) && !f.tagmanager.IsDefinite(postag) && !f.tagmanager.HasPronoun(postag) {
			// add inflection flag
			switch inflection {
			case "jar":
				postag = f.tagmanager.SetMajrour(postag)
			case "raf3":
				postag = f.tagmanager.SetMarfou3(postag)
			case "nasb":
				postag = f.tagmanager.SetMansoub(postag)
			default:
				postag = f.tagmanager.SetMarfou3(postag)
			}
			// add number flag
			switch number {
			case "one":
				postag = f.tagmanager.SetSingle(postag)
			case "two":
				postag = f.tagmanager.SetDual(postag)
			case "plural":
				postag = f.tagmanager.SetPlural(postag)
			default:
				postag = f.tagmanager.SetSingle(postag)
			}
			// add Tanwin
			if number == "one" && inflection == "nasb" {
				postag = f.tagmanager.SetTanwin(postag)
			}
			// for each postag generate a new token
			if !containsString(tmpList, postag) {
				tmpList = append(tmpList, postag)
				if f.synthesizer != nil {
					syhthesizedList, _ := f.synthesizer.Synthesize(tk, postag)
					if len(syhthesizedList) > 0 {
						inflectedList = append(inflectedList, syhthesizedList...)
					}
				}
			}
		}
	}
	return inflectedList
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// getPreviousPos ports getPreviousPos (1-based previousPos → 0-based).
func getPreviousPos(args map[string]string) int {
	previousWordPos := 0
	if args["previousPos"] != "" {
		// Java: try { previousWordPos = Integer.valueOf(args.get("previousPos")) - 1; } catch → -1
		n, err := strconv.Atoi(args["previousPos"])
		if err != nil {
			return -1
		}
		previousWordPos = n - 1
	}
	return previousWordPos
}

// getNextPos ports getNextPos (negative = size+n).
func getNextPos(args map[string]string, size int) int {
	nextPos := 0
	s := args["nextPos"]
	if s == "" {
		s = "0"
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	nextPos = n
	// the next token is index with a negative offset
	if nextPos < 0 {
		nextPos = size + nextPos
	}
	return nextPos
}
