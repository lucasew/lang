package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DonarseliBeFilter ports
// org.languagetool.rules.ca.DonarseliBeFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports language synthesizer for VerbSynthesizer.
// VariantCode is language shortCodeWithCountryAndVariant (e.g. ca-ES-valencia → eixir).
type DonarseliBeFilter struct {
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// VariantCode ports lang.getShortCodeWithCountryAndVariant().
	VariantCode string
	// AdaptSuggestion ports language.adaptSuggestion; nil → identity.
	AdaptSuggestion func(s, originalErrorStr string) string
}

func NewDonarseliBeFilter() *DonarseliBeFilter {
	return &DonarseliBeFilter{}
}

// AdverbiFinal are terminal adverbs accepted after the verb cluster.
var AdverbiFinal = map[string]struct{}{
	"bé": {}, "malament": {}, "mal": {}, "millor": {}, "pitjor": {}, "fatal": {},
}

// PronomsPersonals are strong personal pronouns for "a mi/tu/…" spans.
var PronomsPersonals = map[string]struct{}{
	"mi": {}, "tu": {}, "ell": {}, "ella": {},
	"nosaltres": {}, "vosaltres": {}, "ells": {}, "elles": {},
}

// ExceptionsQue words that block a preceding "que".
var ExceptionsQue = map[string]struct{}{
	"ja": {}, "ara": {}, "per": {}, "de": {}, "a": {}, "en": {},
}

// DespresDarrerAdverbiPOS matches tokens allowed after the final adverb.
// Java: Pattern.compile("V.N.*|D.*|PD.*") with Matcher.matches.
var DespresDarrerAdverbiPOS = regexp.MustCompile(`V.N.*|D.*|PD.*`)

var reQuePrefix = regexp.MustCompile(`(?i)que `)
var reQueAccentPrefix = regexp.MustCompile(`(?i)què `)
var reNoPrefix = regexp.MustCompile(`(?i)no `)
var reMaiPrefix = regexp.MustCompile(`(?i)mai `)

// NormalizeAdverbi maps mal/fatal → malament.
func NormalizeAdverbi(token string) string {
	switch strings.ToLower(token) {
	case "mal", "fatal":
		return "malament"
	default:
		return token
	}
}

// IsAdverbiFinal reports terminal adverbs.
func IsAdverbiFinal(token string) bool {
	_, ok := AdverbiFinal[strings.ToLower(token)]
	return ok
}

// IsPronomPersonal reports strong personal pronouns.
func IsPronomPersonal(token string) bool {
	_, ok := PronomsPersonals[strings.ToLower(token)]
	return ok
}

// BuildDonarSuggestion attaches "en" to a weak pronoun cluster (unit helper).
func (f *DonarseliBeFilter) BuildDonarSuggestion(pronom, verb string, pronounsBefore bool, casingModel string) string {
	norm := Transform(pronom, PronounNormalized) + " en"
	var s string
	if pronounsBefore {
		s = TransformDavant(norm, verb) + verb
	} else {
		s = verb + TransformDarrere(norm, verb)
	}
	if casingModel != "" {
		s = tools.PreserveCase(s, casingModel)
	}
	return s
}

// IsExceptionQue reports whether a token before "que" blocks the rewrite.
func IsExceptionQue(token string) bool {
	_, ok := ExceptionsQue[strings.ToLower(token)]
	return ok
}

// AcceptRuleMatch ports DonarseliBeFilter.acceptRuleMatch.
func (f *DonarseliBeFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil || f.Synthesize == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	verbSynth := synthca.NewVerbSynthesizerAt(tokens, posWord, false)
	verbSynth.Synthesize = f.Synthesize
	if f.AdaptSuggestion != nil {
		verbSynth.AdaptSuggestion = f.AdaptSuggestion
	}
	if verbSynth.IsUndefined() {
		return nil
	}
	if tokens[verbSynth.GetLastVerbIndex()].GetEndPos() > match.GetToPos() {
		return nil
	}

	posDonar := verbSynth.GetLastVerbIndex()
	posPrimerVerb := verbSynth.GetFirstVerbIndex()
	posInitUnderline := posPrimerVerb - verbSynth.GetNumPronounsBefore()
	isPronomFebleDavant := verbSynth.GetNumPronounsBefore() > 0
	posPronomFebleRelevant := -1
	if verbSynth.GetNumPronounsAfter() == 2 {
		posPronomFebleRelevant = posDonar + 2
	} else if verbSynth.GetNumPronounsBefore() >= 2 {
		// Si n'hi ha tres, suposem que és un "hi" que ignorem
		posPronomFebleRelevant = posPrimerVerb - (verbSynth.GetNumPronounsBefore() - 1)
	}
	if posPronomFebleRelevant < 1 || posPronomFebleRelevant >= len(tokens) {
		return nil
	}
	pronomFebleRelevant := readingWithPPronomFeble(tokens[posPronomFebleRelevant])
	if pronomFebleRelevant == nil || pronomFebleRelevant.GetPOSTag() == nil {
		return nil
	}

	// mira darrere: molt bé
	posWord = verbSynth.GetLastVerbIndex() + verbSynth.GetNumPronounsAfter() + 1
	primerAdverbi := posWord
	for posWord < len(tokens) && !IsAdverbiFinal(tokens[posWord].GetToken()) {
		posWord++
	}
	if posWord == len(tokens) || !IsAdverbiFinal(tokens[posWord].GetToken()) {
		return nil
	}
	darrerAdverbi := posWord
	darrerAdverbiStr := tokens[darrerAdverbi].GetToken()
	if strings.EqualFold(darrerAdverbiStr, "mal") || strings.EqualFold(darrerAdverbiStr, "fatal") {
		darrerAdverbiStr = "malament"
	}

	var despresDarrerAdverbi *languagetool.AnalyzedToken
	if darrerAdverbi+1 < len(tokens) {
		despresDarrerAdverbi = readingWithTagRegex(tokens[darrerAdverbi+1], `V.N.*|D.*|PD.*`)
	}

	addTokensToRight := 0
	addStringToRight := ""
	addTokensToLeft := 0
	addStringToLeft := ""

	// analitza paraules prèvies
	isNo := posInitUnderline-1 > 0 &&
		(strings.EqualFold(tokens[posInitUnderline-1].GetToken(), "no") ||
			strings.EqualFold(tokens[posInitUnderline-1].GetToken(), "mai"))
	isMaiNo := isNo && posInitUnderline-2 > 0 &&
		strings.EqualFold(tokens[posInitUnderline-2].GetToken(), "mai")
	isMalament := strings.EqualFold(darrerAdverbiStr, "malament") || strings.EqualFold(darrerAdverbiStr, "pitjor")
	// No ... malament
	isNoMalament := isNo && isMalament
	// ... malament
	isMalament = isMalament && !isNoMalament
	if isMaiNo {
		addTokensToLeft++
	}
	if isNo {
		addTokensToLeft++
	}
	aMiString := ""
	// Java uses & (bitwise) between boolean expressions — both sides always evaluated.
	if posInitUnderline-addTokensToLeft-2 > 0 &&
		strings.EqualFold(tokens[posInitUnderline-addTokensToLeft-2].GetToken(), "a") &&
		IsPronomPersonal(tokens[posInitUnderline-addTokensToLeft-1].GetToken()) {
		aMiString = tokens[posInitUnderline-addTokensToLeft-2].GetToken() + " " +
			tokens[posInitUnderline-addTokensToLeft-1].GetToken() + " "
		addTokensToLeft += 2
	}
	isQue := posInitUnderline-addTokensToLeft-1 > 0 &&
		strings.EqualFold(tokens[posInitUnderline-addTokensToLeft-1].GetToken(), "que") &&
		!IsVerbDicendiBeforeTokens(tokens, posInitUnderline-addTokensToLeft-2)
	isQueAccent := posInitUnderline-addTokensToLeft-1 > 0 &&
		strings.EqualFold(tokens[posInitUnderline-addTokensToLeft-1].GetToken(), "què")
	if posInitUnderline-addTokensToLeft-2 > 0 &&
		IsExceptionQue(tokens[posInitUnderline-addTokensToLeft-2].GetToken()) {
		isQueAccent = false
		isQue = false
	}
	isElQue := false
	isAQui := false
	if posInitUnderline-addTokensToLeft-1 > 0 &&
		strings.EqualFold(tokens[posInitUnderline-addTokensToLeft-1].GetToken(), "qui") {
		isElQue = true
		if posInitUnderline-addTokensToLeft-2 > 0 &&
			strings.EqualFold(tokens[posInitUnderline-addTokensToLeft-2].GetToken(), "a") {
			isAQui = true
		}
	}
	if isQue && posInitUnderline-addTokensToLeft-2 > 0 &&
		(tokens[posInitUnderline-addTokensToLeft-2].HasPosTagStartingWith("DA") ||
			tokens[posInitUnderline-addTokensToLeft-2].HasAnyLemma("alumne", "persona", "estudiant", "professor")) {
		isElQue = true
		isQue = false
	}
	if isQue {
		addTokensToLeft++
	}
	if isQueAccent {
		addTokensToLeft++
	}
	for j := posInitUnderline - addTokensToLeft; j < posInitUnderline; j++ {
		if j >= 0 && j < len(tokens) {
			addStringToLeft += tokens[j].GetToken() + " "
		}
	}

	// Crea suggeriments
	var replacements []string
	pronomTag := *pronomFebleRelevant.GetPOSTag()
	if len(pronomTag) < 5 {
		return nil
	}
	persona := pronomTag[2:3]
	nombre := pronomTag[4:5]
	primerVerb := readingWithTagRegex(tokens[posPrimerVerb], `V.*`)
	if primerVerb == nil || primerVerb.GetPOSTag() == nil || len(*primerVerb.GetPOSTag()) < 8 {
		return nil
	}
	verbPostag := *primerVerb.GetPOSTag()
	newVerbPostag := verbPostag[:4] + persona + nombre + verbPostag[6:8]

	// tinc traça (per a)
	addStringToLeftTincTraca := reQueAccentPrefix.ReplaceAllString(addStringToLeft, "en què ")
	addStringToLeftTincTraca = reQuePrefix.ReplaceAllString(addStringToLeftTincTraca, "en què ")
	if aMiString != "" {
		addStringToLeftTincTraca = strings.Replace(addStringToLeftTincTraca, aMiString, "", 1)
	}
	if isNoMalament {
		addStringToLeftTincTraca = reNoPrefix.ReplaceAllString(addStringToLeftTincTraca, "")
		addStringToLeftTincTraca = reMaiPrefix.ReplaceAllString(addStringToLeftTincTraca, "")
	}
	var suggestion strings.Builder
	suggestion.WriteString(addStringToLeftTincTraca)
	if isMalament {
		suggestion.WriteString("no ")
	}
	if !strings.HasPrefix(strings.ToLower(addStringToLeft), "qu") && despresDarrerAdverbi == nil {
		suggestion.WriteString("hi ")
	}
	verbSynth.SetLemmaAndPostag("tenir", newVerbPostag)
	suggestion.WriteString(verbSynth.SynthesizeForm())
	if !isNoMalament && !isMalament {
		suggestion.WriteString(getAdverbsFor(tokens, primerAdverbi, darrerAdverbi, "traça"))
	}
	suggestion.WriteString(" traça")
	if despresDarrerAdverbi != nil {
		switch strings.ToLower(despresDarrerAdverbi.GetToken()) {
		case "el":
			suggestion.WriteString(" per al")
			addTokensToRight = 1
			addStringToRight = " el"
		case "els":
			suggestion.WriteString(" per als")
			addTokensToRight = 1
			addStringToRight = " els"
		default:
			suggestion.WriteString(" per a")
		}
	}
	casingTok := tokens[posInitUnderline-addTokensToLeft].GetToken()
	if !isElQue {
		replacements = append(replacements, tools.PreserveCase(suggestion.String(), casingTok))
	}

	// faig bé
	suggestion.Reset()
	leftNoAMi := addStringToLeft
	if aMiString != "" {
		leftNoAMi = strings.Replace(addStringToLeft, aMiString, "", 1)
	}
	suggestion.WriteString(leftNoAMi)
	if !strings.HasPrefix(strings.ToLower(addStringToLeft), "qu") && despresDarrerAdverbi == nil && !isElQue {
		suggestion.WriteString("ho ")
	}
	verbSynth.SetLemmaAndPostag("fer", newVerbPostag)
	suggestion.WriteString(verbSynth.SynthesizeForm())
	suggestion.WriteString(getAdverbsFor(tokens, primerAdverbi, darrerAdverbi, "bé"))
	suggestion.WriteString(" " + darrerAdverbiStr)
	suggestion.WriteString(addStringToRight)
	if !isAQui {
		replacements = append(replacements, tools.PreserveCase(suggestion.String(), casingTok))
	}

	// me'n surto (en)
	suggestion.Reset()
	suggestion.WriteString(addStringToLeftTincTraca)
	if isMalament {
		suggestion.WriteString("no ")
	}
	if isPronomFebleDavant {
		pronom := pronomFebleRelevant.GetToken()
		if strings.EqualFold(pronom, "'ls") || strings.EqualFold(pronom, "li") {
			pronom = "es"
		}
		pronomsNormalitzats := Transform(pronom, PronounNormalized) + " en"
		suggestion.WriteString(TransformDavant(pronomsNormalitzats, primerVerb.GetToken()))
	}
	verbSynth.SetLemmaAndPostag("sortir", newVerbPostag)
	suggestion.WriteString(verbSynth.SynthesizeForm())
	if !isPronomFebleDavant {
		pronom := pronomFebleRelevant.GetToken()
		if strings.EqualFold(pronom, "'ls") || strings.EqualFold(pronom, "-li") {
			pronom = "es"
		}
		pronomsNormalitzats := Transform(pronom, PronounNormalized) + " en"
		suggestion.WriteString(TransformDarrere(pronomsNormalitzats, primerVerb.GetToken()))
	}
	if despresDarrerAdverbi != nil {
		if despresDarrerAdverbi.GetPOSTag() != nil && strings.HasPrefix(*despresDarrerAdverbi.GetPOSTag(), "V") {
			suggestion.WriteString(" a")
		} else {
			suggestion.WriteString(" en")
		}
	}
	suggestion.WriteString(addStringToRight)
	if !isElQue {
		replacements = append(replacements, tools.PreserveCase(suggestion.String(), casingTok))
	}

	// em van bé
	suggestion.Reset()
	suggestion.WriteString(addStringToLeft)
	verbSynth.SetLemmaAndPostag("anar", verbPostag)
	verb := verbSynth.SynthesizeForm()
	if isPronomFebleDavant {
		suggestion.WriteString(TransformDavant(pronomFebleRelevant.GetToken(), verb))
	}
	suggestion.WriteString(verb)
	if !isPronomFebleDavant {
		suggestion.WriteString(TransformDarrere(pronomFebleRelevant.GetToken(), verb))
	}
	suggestion.WriteString(getAdverbsFor(tokens, primerAdverbi, darrerAdverbi, "bé"))
	suggestion.WriteString(" " + darrerAdverbiStr)
	suggestion.WriteString(addStringToRight)
	replacements = append(replacements, tools.PreserveCase(suggestion.String(), casingTok))

	// m'ixen bé / em surten bé
	suggestion.Reset()
	suggestion.WriteString(addStringToLeft)
	newLemmaSortir := "sortir"
	if f.VariantCode == "ca-ES-valencia" {
		newLemmaSortir = "eixir"
	}
	verbSynth.SetLemmaAndPostag(newLemmaSortir, verbPostag)
	verb = verbSynth.SynthesizeForm()
	if isPronomFebleDavant {
		suggestion.WriteString(TransformDavant(pronomFebleRelevant.GetToken(), verb))
	}
	suggestion.WriteString(verb)
	if !isPronomFebleDavant {
		suggestion.WriteString(TransformDarrere(pronomFebleRelevant.GetToken(), verb))
	}
	suggestion.WriteString(getAdverbsFor(tokens, primerAdverbi, darrerAdverbi, "bé"))
	suggestion.WriteString(" " + darrerAdverbiStr)
	suggestion.WriteString(addStringToRight)
	replacements = append(replacements, tools.PreserveCase(suggestion.String(), casingTok))

	if len(replacements) == 0 {
		return nil
	}
	endIdx := darrerAdverbi + addTokensToRight
	if endIdx < 0 || endIdx >= len(tokens) {
		return nil
	}
	startIdx := posInitUnderline - addTokensToLeft
	if startIdx < 0 || startIdx >= len(tokens) {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startIdx].GetStartPos(), tokens[endIdx].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(replacements)
	return out
}

func getAdverbsFor(tokens []*languagetool.AnalyzedTokenReadings, primerAdverbi, darrerAdverbi int, target string) string {
	var result strings.Builder
	for i := primerAdverbi; i < darrerAdverbi; i++ {
		if i < 0 || i >= len(tokens) {
			continue
		}
		if tokens[i].IsWhitespaceBefore() {
			result.WriteByte(' ')
		}
		result.WriteString(tokens[i].GetToken())
	}
	resultStr := result.String()
	if target == "traça" {
		switch strings.ToLower(resultStr) {
		case " molt":
			resultStr = " molta"
		case " gens":
			resultStr = " gens de"
		case " tan":
			resultStr = " tanta"
		}
	}
	return resultStr
}
