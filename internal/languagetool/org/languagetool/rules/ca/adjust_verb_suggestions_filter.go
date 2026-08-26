package ca

import (
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdjustVerbSuggestionsFilter ports
// org.languagetool.rules.ca.AdjustVerbSuggestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize + GetTargetPosTag port Catalan synthesizer.
// AnalyzeText ports lt.analyzeText for numberFromNextWords (optional).
type AdjustVerbSuggestionsFilter struct {
	// Synthesize ports language.getSynthesizer().synthesize for VerbSynthesizer.
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// GetTargetPosTag ports CatalanSynthesizer.getTargetPosTag; nil → Catalan default.
	GetTargetPosTag func(posTags []string, targetPosTag string) string
	// AnalyzeText ports JLanguageTool.analyzeText; used when numberFromNextWords=true.
	// Returns non-blank tokens of the first sentence (index 0 is often SENT_START).
	AnalyzeText func(text string) []*languagetool.AnalyzedTokenReadings
	// AdaptSuggestion ports language.adaptSuggestion; used via AdaptSuggestionsList.
	// When nil, uses package AdaptSuggestion.
}

func NewAdjustVerbSuggestionsFilter() *AdjustVerbSuggestionsFilter {
	return &AdjustVerbSuggestionsFilter{}
}

var (
	needsApostropheChange  = []string{"de", "d'", "l", "l'", "el"}
	needsContractionChange = []string{"a", "de", "per", "pe"}
)

// VerbSuggestionContext is the surface verb/pronoun input (unit helper).
type VerbSuggestionContext struct {
	PronounsStr            string
	VerbStr                string
	FirstVerbPersonaNumber string
	PronounsAfter          bool
	WholeOriginal          string
	CasingModel            string
}

// Suggest applies actions (unit helper).
func (f *AdjustVerbSuggestionsFilter) Suggest(ctx VerbSuggestionContext, actionsCSV string) []string {
	actions := strings.Split(actionsCSV, ",")
	if actionsCSV == "" {
		actions = []string{"removePronounReflexive"}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, action := range actions {
		action = tools.JavaStringTrim(action)
		var replacement string
		switch action {
		case "addPronounEn":
			np := DoAddPronounEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
			if np != "" {
				if ctx.PronounsAfter {
					replacement = ctx.VerbStr + np
				} else {
					replacement = np + ctx.VerbStr
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
		case "addPronounReflexiveEn":
			replacement = DoAddPronounReflexiveEn(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, "hi "+ctx.VerbStr, ctx.FirstVerbPersonaNumber, false)
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber)
		case "addPronounHi":
			if !strings.Contains(ctx.PronounsStr, "hi") {
				replacement = TransformDavant("hi", ctx.VerbStr) + ctx.VerbStr
			}
		case "None", "none", "":
			continue
		default:
			continue
		}
		if replacement == "" || strings.EqualFold(replacement, ctx.WholeOriginal) {
			continue
		}
		if ctx.CasingModel != "" {
			replacement = tools.PreserveCase(replacement, ctx.CasingModel)
		}
		if _, ok := seen[replacement]; ok {
			continue
		}
		seen[replacement] = struct{}{}
		out = append(out, replacement)
	}
	return out
}

// AcceptRuleMatch ports AdjustVerbSuggestionsFilter.acceptRuleMatch.
// Requires match.SuggestedReplacements as lemma seeds (Java iterates those).
func (f *AdjustVerbSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	if f.Synthesize == nil {
		return nil
	}
	if arguments == nil {
		arguments = map[string]string{}
	}

	numberFromNextWords := strings.EqualFold(patterns.GetOptionalDefault("numberFromNextWords", arguments, "false"), "true")
	actionsCSV := patterns.GetOptionalDefault("actions", arguments, "removePronounReflexive")
	actions := strings.Split(actionsCSV, ",")
	if len(actions) == 0 {
		actions = []string{"removePronounReflexive"}
	}
	forceNumber := patterns.GetOptionalDefault("forceNumber", arguments, "")

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	verbSynthesizer := synthca.NewVerbSynthesizerAt(tokens, posWord, false)
	verbSynthesizer.Synthesize = f.Synthesize
	if verbSynthesizer.IsUndefined() {
		return nil
	}
	if tokens[verbSynthesizer.GetLastVerbIndex()].GetEndPos() > match.GetToPos() {
		return nil
	}

	getTarget := f.GetTargetPosTag
	if getTarget == nil {
		getTarget = catalanGetTargetPosTag
	}

	var replacements []string
	for _, originalSuggestion := range match.GetSuggestedReplacements() {
		originalSuggestion = strings.ToLower(originalSuggestion)
		makeIntransitive := false
		desiredNumber := ""
		desiredPersona := ""
		action := tools.JavaStringTrim(actions[0])
		if strings.HasSuffix(originalSuggestion, " [intr]") {
			originalSuggestion = originalSuggestion[:len(originalSuggestion)-7]
			makeIntransitive = true
		}
		if strings.HasSuffix(originalSuggestion, " [3s]") {
			originalSuggestion = originalSuggestion[:len(originalSuggestion)-5]
			desiredNumber = "S"
			desiredPersona = "3"
		}
		if strings.HasPrefix(originalSuggestion, "[datiu] ") {
			originalSuggestion = originalSuggestion[8:]
			action = "addPronounDative"
		}
		firstSpaceIndex := strings.Index(originalSuggestion, " ")
		newLemma := originalSuggestion
		afterLemma := ""
		if firstSpaceIndex != -1 {
			newLemma = originalSuggestion[:firstSpaceIndex]
			afterLemma = originalSuggestion[firstSpaceIndex+1:]
			if numberFromNextWords && f.AnalyzeText != nil {
				analyzed := f.AnalyzeText(afterLemma)
				// Java: tokensWithoutWhitespace[1] — after SENT_START
				if len(analyzed) > 1 && analyzed[1] != nil {
					if analyzed[1].HasPartialPosTag("S") {
						desiredNumber = "S"
					} else {
						desiredNumber = "P"
					}
				}
			}
		}
		if strings.Contains(newLemma, "haver-hi") {
			desiredNumber = "S"
		}
		if forceNumber != "" {
			desiredNumber = forceNumber
		}
		if strings.HasSuffix(newLemma, "-se'n") {
			newLemma = newLemma[:len(newLemma)-5]
			action = "addPronounReflexiveEn"
		} else if strings.HasSuffix(newLemma, "-se") {
			newLemma = newLemma[:len(newLemma)-3]
			action = "addPronounReflexive"
		} else if strings.HasSuffix(newLemma, "'s") {
			newLemma = newLemma[:len(newLemma)-2]
			action = "addPronounReflexive"
		} else if strings.HasSuffix(newLemma, "-hi") {
			newLemma = newLemma[:len(newLemma)-3]
			action = "addPronounHi"
		} else if strings.HasSuffix(newLemma, "-s'ho") {
			newLemma = newLemma[:len(newLemma)-5]
			action = "addPronounReflexiveHo"
		} else if strings.HasSuffix(newLemma, "-se-les") {
			newLemma = newLemma[:len(newLemma)-7]
			action = "addPronounReflexiveLes"
		} else if strings.HasSuffix(newLemma, "-s'hi") {
			newLemma = newLemma[:len(newLemma)-5]
			action = "addPronounReflexiveHi"
		}

		// synthesize with new lemma
		var postags []string
		firstVerbTok := tokens[verbSynthesizer.GetFirstVerbIndex()]
		for _, reading := range firstVerbTok.GetReadings() {
			if reading == nil || reading.GetPOSTag() == nil {
				continue
			}
			postag := *reading.GetPOSTag()
			if !strings.HasPrefix(postag, "V") {
				continue
			}
			if len(postag) >= 6 {
				if desiredNumber != "" {
					if postag[2:3] != "P" && (postag[5:6] == "S" || postag[5:6] == "P") {
						postag = postag[:5] + desiredNumber + postag[6:]
					}
				}
				if desiredPersona != "" {
					if postag[2:3] != "P" && len(postag) > 4 &&
						(postag[4] == '1' || postag[4] == '2' || postag[4] == '3') {
						postag = postag[:4] + desiredPersona + postag[5:]
					}
				}
			}
			postags = append(postags, postag)
		}
		targetPostag := getTarget(postags, "")
		verbStr := ""
		if targetPostag != "" {
			verbSynthesizer.SetLemmaAndPostag(newLemma, targetPostag)
			verbStr = verbSynthesizer.SynthesizeForm()
		}
		pronounsStr := ""
		isPronounsAfter := verbSynthesizer.GetNumPronounsAfter() > 0 || !verbSynthesizer.IsFirstVerbIS()
		if verbSynthesizer.GetNumPronounsBefore() > 0 {
			pronounsStr = verbSynthesizer.GetPronounsStrBefore()
		} else if verbSynthesizer.GetNumPronounsAfter() > 0 {
			pronounsStr = verbSynthesizer.GetPronounsStrAfter()
		}
		pronounsStr = strings.ToLower(pronounsStr)
		firstVerbPersonaNumber := ""
		if action == "addPronounDative" {
			firstVerbPersonaNumber = verbSynthesizer.GetFirstVerbPersonaNumber()
		} else if len(targetPostag) >= 6 {
			firstVerbPersonaNumber = targetPostag[4:6]
		}

		replacement := ""
		switch action {
		case "addPronounEn":
			newPronoun := DoAddPronounEn(pronounsStr, verbStr, !verbSynthesizer.IsFirstVerbIS())
			if newPronoun != "" {
				if verbSynthesizer.IsFirstVerbIS() {
					replacement = newPronoun + verbStr
				} else {
					replacement = verbStr + newPronoun
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(pronounsStr, verbStr, isPronounsAfter)
		case "addPronounReflexiveEn":
			replacement = DoAddPronounReflexiveEn(pronounsStr, verbStr, firstVerbPersonaNumber, isPronounsAfter)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(pronounsStr, verbStr, isPronounsAfter)
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(pronounsStr, verbStr, firstVerbPersonaNumber, isPronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive("", "hi "+verbStr, firstVerbPersonaNumber, isPronounsAfter)
		case "addPronounReflexiveLes":
			replacement = DoAddPronounReflexive(
				Transform(strings.ToLower(pronounsStr), PronounNormalized)+" les",
				verbStr, firstVerbPersonaNumber, isPronounsAfter)
		case "addPronounDative":
			dativePronoun := GetDativePronoun(firstVerbPersonaNumber)
			if isPronounsAfter {
				replacement = verbStr + TransformDarrere(dativePronoun, verbStr)
			} else {
				replacement = TransformDavant(dativePronoun, verbStr) + verbStr
			}
		case "addPronounReflexiveHo":
			reflexivePronoun := GetReflexivePronoun(firstVerbPersonaNumber)
			if reflexivePronoun == "" {
				if pronounsStr != "" {
					rp := Transform(pronounsStr, PronounNormalized)
					if _, ok := LReflexivePronouns[rp]; ok {
						reflexivePronoun = rp
					}
				}
			}
			if reflexivePronoun == "" {
				reflexivePronoun = "es"
			}
			pronounsNormalized := reflexivePronoun + " ho"
			if isPronounsAfter {
				replacement = verbStr + TransformDarrere(pronounsNormalized, verbStr)
			} else {
				replacement = TransformDavant(pronounsNormalized, verbStr) + verbStr
			}
		case "addPronounHi":
			replacement = "hi " + verbStr
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(pronounsStr, verbStr, firstVerbPersonaNumber)
		case "None":
			if isPronounsAfter {
				replacement = verbStr + TransformDarrere(pronounsStr, verbStr)
			} else {
				replacement = TransformDavant(pronounsStr, verbStr) + verbStr
			}
		}
		if replacement != "" {
			if makeIntransitive {
				replacement = ConvertPronounsForIntransitiveVerb(replacement)
			}
			replacement = FixApostrophes(replacement)
			replacement = tools.JavaStringTrim(replacement + " " + afterLemma)
			replacements = append(replacements, tools.PreserveCase(replacement, verbSynthesizer.GetCasingModel()))
		}
	}
	if len(replacements) == 0 {
		return nil
	}

	posStartUnderline := verbSynthesizer.GetFirstVerbIndex() - verbSynthesizer.GetNumPronounsBefore()
	if verbSynthesizer.GetNumPronounsBefore() == 0 && posStartUnderline > 1 &&
		containsStrFold(needsApostropheChange, tokens[posStartUnderline-1].GetToken()) &&
		anyChangeVowelConsonant(verbSynthesizer.GetVerbStr(), replacements) {
		var prefix strings.Builder
		if posStartUnderline > 2 && containsStrFold(needsContractionChange, tokens[posStartUnderline-2].GetToken()) {
			prefix.WriteString(strings.ToLower(verbSynthesizer.GetStringFromTo(posStartUnderline-2, posStartUnderline-1)))
			if tokens[posStartUnderline].IsWhitespaceBefore() {
				prefix.WriteByte(' ')
			}
			posStartUnderline = posStartUnderline - 2
		} else {
			prefix.WriteString(strings.ToLower(tokens[posStartUnderline-1].GetToken()))
			if tokens[posStartUnderline].IsWhitespaceBefore() {
				prefix.WriteByte(' ')
			}
			posStartUnderline = posStartUnderline - 1
		}
		p := prefix.String()
		for i := range replacements {
			replacements[i] = p + replacements[i]
		}
	}

	endingPos := match.GetToPos()
	if lastEnd := tokens[verbSynthesizer.GetLastIndex()].GetEndPos(); lastEnd > endingPos {
		endingPos = lastEnd
	}
	if posStartUnderline < 0 || posStartUnderline >= len(tokens) {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[posStartUnderline].GetStartPos(), endingPos,
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()

	// Java substring uses UTF-16 positions (token start/end, match ToPos).
	text := match.Sentence.GetText()
	from := tokens[posStartUnderline].GetStartPos()
	to := match.GetToPos()
	originalStr := rules.UTF16Substring(text, from, to)
	out.SetSuggestedReplacements(adaptSuggestionsListCA(replacements, originalStr))
	return out
}

func adaptSuggestionsListCA(suggestions []string, originalErrorStr string) []string {
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		out = append(out, AdaptSuggestion(s, originalErrorStr))
	}
	return out
}

func anyChangeVowelConsonant(originalVerb string, replacements []string) bool {
	originalNeeds := pApostropheNeeded.MatchString(originalVerb)
	for _, replacement := range replacements {
		if originalNeeds != pApostropheNeeded.MatchString(replacement) {
			return true
		}
	}
	return false
}

func containsStrFold(list []string, tok string) bool {
	t := strings.ToLower(tok)
	for _, s := range list {
		if t == s {
			return true
		}
	}
	return false
}

// catalanGetTargetPosTag ports CatalanSynthesizer.getTargetPosTag (PostagComparator).
func catalanGetTargetPosTag(posTags []string, targetPosTag string) string {
	if len(posTags) == 0 {
		return targetPosTag
	}
	tags := append([]string(nil), posTags...)
	sort.SliceStable(tags, func(i, j int) bool {
		return catalanPostagLess(tags[i], tags[j])
	})
	// return the last one to keep the previous results
	return tags[len(tags)-1]
}

// catalanPostagLess is true if a < b under Catalan PostagComparator (sort ascending).
func catalanPostagLess(arg0, arg1 string) bool {
	// Comparator returns negative if arg0 < arg1
	cmp := catalanPostagCompare(arg0, arg1)
	return cmp < 0
}

func catalanPostagCompare(arg0, arg1 string) int {
	len0, len1 := len(arg0), len(arg1)
	if len0 > 4 && len1 > 4 {
		if strings.Contains(arg0, "3S") && arg1 == "1S" {
			return 150
		}
		if strings.Contains(arg0, "1S") && strings.Contains(arg1, "3S") {
			return -150
		}
		if arg0 == "VMIP2P00" && arg1 == "VMIS3S00" {
			return 150
		}
		if arg1 == "VMIP2P00" && arg0 == "VMIS3S00" {
			return -150
		}
		if arg0[2] == 'I' && arg1[2] != 'I' {
			return 100
		}
		if arg1[2] == 'I' && arg0[2] != 'I' {
			return -100
		}
		if arg0[4] == '3' && arg1[4] == '1' {
			return 50
		}
		if arg1[4] == '1' && arg0[4] == '3' {
			return -50
		}
	}
	return 0
}
