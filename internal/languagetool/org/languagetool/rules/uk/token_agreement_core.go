package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// FakeFemList ports TokenAgreementAdjNounRule.FAKE_FEM_LIST.
var FakeFemList = []string{
	"ступінь", "степінь", "продаж", "собака", "дріб", "ярмарок",
	"нежить", "рукопис", "накип", "насип", "путь",
}

var (
	adjInflectionPattern  = regexp.MustCompile(`:([mfnp]):(v_...)(:r(in)?anim)?`)
	nounInflectionPattern = regexp.MustCompile(`((?:[iu]n)?anim):([mfnps]):(v_...)`)
	nounVZnaVarIgnore     = regexp.MustCompile(`v_zna:var`)
)

// CollectPOSTags gathers non-nil POS tags from an AnalyzedTokenReadings.
func CollectPOSTags(tok *languagetool.AnalyzedTokenReadings) []string {
	if tok == nil {
		return nil
	}
	var out []string
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		out = append(out, *r.GetPOSTag())
	}
	return out
}

// HasAdjReading reports whether any reading is adj*.
func HasAdjReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSAdj.Match(p) {
			return true
		}
	}
	return false
}

// HasNounReading reports whether any reading is noun*.
func HasNounReading(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNoun.Match(p) {
			return true
		}
	}
	return false
}

// HasNounOrPronSubjectReading treats personal pronouns as subjects for noun–verb agreement.
func HasNounOrPronSubjectReading(tok *languagetool.AnalyzedTokenReadings) bool {
	if HasNounReading(tok) {
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "pron:pers") {
			return true
		}
	}
	return false
}

// AdjNounAgree reports whether adj and noun POS tag sets share an inflection.
func AdjNounAgree(adjTags, nounTags []string) bool {
	master := GetAdjCaseInflections(adjTags)
	slave := GetNounInflectionsFromTags(nounTags, nounVZnaVarIgnore)
	if len(master) == 0 || len(slave) == 0 {
		return true // insufficient data — no flag
	}
	return InflectionsIntersect(master, slave)
}

// NumrNounAgree uses numr inflection pattern against nouns.
func NumrNounAgree(numrTags, nounTags []string) bool {
	master := GetNumrCaseInflections(numrTags)
	slave := GetNounCaseInflections(nounTags)
	if len(master) == 0 || len(slave) == 0 {
		return true
	}
	return InflectionsIntersect(master, slave)
}

// tokenAgreementMatch is shared match infrastructure.
// Java TokenAgreement* rules: setCategory(Categories.MISC).
type tokenAgreementMatch struct {
	ruleID      string
	description string
	shortMsg    string
	// Category ports Rule.category (Java MISC).
	category *rules.Category
	// pairChecker returns false when the pair disagrees
	pairChecker func(left, right *languagetool.AnalyzedTokenReadings) bool
	// isLeftToken identifies the "master" token class
	isLeftToken func(tok *languagetool.AnalyzedTokenReadings) bool
	// isRightToken identifies the "slave" token class
	isRightToken func(tok *languagetool.AnalyzedTokenReadings) bool
	// exception when true skips the flag
	exception func(tokens []*languagetool.AnalyzedTokenReadings, leftIdx, rightIdx int) bool
}

func (r *tokenAgreementMatch) GetID() string          { return r.ruleID }
func (r *tokenAgreementMatch) GetDescription() string { return r.description }
func (r *tokenAgreementMatch) GetShort() string       { return r.shortMsg }

// GetCategory ports Rule.getCategory (Java MISC).
func (r *tokenAgreementMatch) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.category
}

// initTokenAgreementMeta applies Java TokenAgreement* constructor metadata (MISC category).
func initTokenAgreementMeta(r *tokenAgreementMatch, messages map[string]string) {
	if r == nil {
		return
	}
	if r.category == nil {
		r.category = rules.CatMisc.GetCategory(messages)
	}
}

func (r *tokenAgreementMatch) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r.pairChecker == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	leftIdx := -1
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if r.isLeftToken != nil && r.isLeftToken(tok) {
			leftIdx = i
			continue
		}
		if leftIdx < 0 {
			continue
		}
		if r.isRightToken != nil && !r.isRightToken(tok) {
			// skip ignorable intermediates (не, і, commas soft)
			if isIgnorableAgreementIntervening(tok) {
				continue
			}
			// non-matching intermediate — reset
			leftIdx = -1
			continue
		}
		if r.exception != nil && r.exception(tokens, leftIdx, i) {
			leftIdx = -1
			continue
		}
		if !r.pairChecker(tokens[leftIdx], tok) {
			msg := r.shortMsg
			if msg == "" {
				msg = r.description
			}
			m := rules.NewRuleMatch(r, sentence, tokens[leftIdx].GetStartPos(), tok.GetEndPos(), msg)
			out = append(out, m)
		}
		leftIdx = -1
	}
	return out
}

// isIgnorableAgreementIntervening allows particle/conj glue between master and slave.
func isIgnorableAgreementIntervening(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	// surface fast path
	switch strings.ToLower(tok.GetToken()) {
	case "не", "й", "і", "та", "чи", "то", "ж", "би", "б":
		return true
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "part") || strings.HasPrefix(p, "conj") {
			return true
		}
	}
	return false
}

// IsPredicativeAdjException soft-skips predicative adjectives.
func IsPredicativeAdjException(adj *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(adj) {
		if strings.Contains(p, "predic") || strings.HasPrefix(p, "predic") {
			return true
		}
	}
	return false
}

// IsAdjpException soft-skips pure participle adjp without case agreement expectation.
func IsAdjpException(adj *languagetool.AnalyzedTokenReadings) bool {
	tags := CollectPOSTags(adj)
	if len(tags) == 0 {
		return false
	}
	hasAdjp, hasCaseAdj := false, false
	for _, p := range tags {
		if strings.Contains(p, "adjp") {
			hasAdjp = true
		}
		if strings.HasPrefix(p, "adj") && strings.Contains(p, "v_") {
			hasCaseAdj = true
		}
	}
	return hasAdjp && !hasCaseAdj
}

// --- Agreement exception helpers (logic in Is*Exception; thin twins in *_exception_helper.go) ---

// IsAdjNounException ports TokenAgreementAdjNounExceptionHelper.isException
// (surface/lemma/conj/case-gov arms; RE2-safe stand-ins for lookarounds).
func IsAdjNounException(tokens []*languagetool.AnalyzedTokenReadings, adjPos, nounPos int) bool {
	if adjPos < 0 || nounPos < 0 || adjPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	// skip if same token
	if adjPos == nounPos {
		return true
	}
	// Java: LemmaHelper.hasLemma(noun, FAKE_FEM_LIST, "noun:inanim:m:")
	if HasLemmaWithPartPos(tokens[nounPos], FakeFemList, "noun:inanim:m:") {
		return true
	}

	adj := tokens[adjPos]
	noun := tokens[nounPos]
	if adj == nil || noun == nil {
		return true
	}

	// схований всередині номера: intervening adv with case government matches noun case
	if nounPos-adjPos > 1 {
		mid := tokens[adjPos+1]
		cases := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(mid, "adv")
		if len(cases) > 0 {
			var list []string
			for c := range cases {
				list = append(list, c)
			}
			if HasVidmPosTag(list, noun) {
				return true
			}
		}
	}

	// Великий + Вітчизняний/Житомирський (capitalized), not війна
	if adjPos > 1 {
		prev := tokens[adjPos-1]
		if prev != nil &&
			IsCapitalized(adj.GetCleanToken()) && IsCapitalized(prev.GetCleanToken()) &&
			(HasLemmaToken(adj, "вітчизняний") || HasLemmaToken(adj, "житомирський")) &&
			HasLemmaToken(prev, "великий") &&
			!HasLemmaToken(noun, "війна") {
			return true
		}
		// Перший Національний (both uppercased first char)
		if HasLemmaToken(adj, "національний") && HasLemmaToken(prev, "перший") {
			at, pt := adj.GetToken(), prev.GetToken()
			if at != "" && pt != "" && isUpperFirst(at) && isUpperFirst(pt) {
				return true
			}
		}
		// (ні)чого доброго
		if CleanTokenLower(adj) == "доброго" {
			if regexp.MustCompile(`^(ні)?чого$`).MatchString(CleanTokenLower(prev)) {
				return true
			}
		}
		// у/в середньому|цілому|основному|подальшому (compare lowercased; RE2 has no (?iu))
		if regexp.MustCompile(`^(середньому|цілому|основному|подальшому)$`).MatchString(CleanTokenLower(adj)) &&
			regexp.MustCompile(`^[ву]$`).MatchString(CleanTokenLower(prev)) {
			return true
		}
		// лава запасних
		if adj.GetToken() == "запасних" && HasLemmaToken(prev, "лава") {
			return true
		}
		// статтю 6-ту / num after стаття
		if HasPosTagPart(adj, "num") && HasLemmaToken(prev, "стаття") {
			return true
		}
	}

	// голому сорочка
	if strings.EqualFold(CleanTokenLower(adj), "голому") && strings.EqualFold(CleanTokenLower(noun), "сорочка") {
		return true
	}
	// бережений бог
	if HasLemmaWithPosRE(adj, []string{"бережений"}, regexp.MustCompile(`^adj:m:v_rod.*$`)) &&
		HasLemmaWithPosRE(noun, []string{"бог"}, regexp.MustCompile(`^noun:anim:m:v_naz.*$`)) {
		return true
	}
	// кожний + mass/quantity noun in instrumental
	if HasLemmaWithPosRE(adj, []string{"кожний"}, regexp.MustCompile(`^adj:f:v_naz.*$`)) &&
		HasLemmaWithPosRE(noun,
			[]string{"вага", "маса", "вартість", "потужність", "тривалість", "чисельність", "номінал", "наклад"},
			regexp.MustCompile(`^noun:inanim:.:v_oru.*$`)) {
		return true
	}
	// Божий / Господній / Христовий capitalized
	if HasLemmaTokenAny(adj, []string{"божий", "господній", "Христовий"}) && isUpperFirst(adj.GetToken()) {
		return true
	}
	// 5-а клас
	if regexp.MustCompile(`^([1-9]|1[0-2])[\x{2018}-][а-д]$`).MatchString(adj.GetToken()) && HasLemmaToken(noun, "клас") {
		return true
	}
	// перший + not FAKE_FEM inanim:m
	if nounPos > 1 && HasLemmaTokenAny(adj, []string{"перший"}) &&
		!HasLemmaWithPartPos(noun, FakeFemList, "noun:inanim:m:") {
		return true
	}
	// старший зміни/групи
	if (CleanTokenLower(noun) == "зміни" || CleanTokenLower(noun) == "групи") && HasLemmaToken(adj, "старший") {
		return true
	}

	// --- further TokenAgreementAdjNounExceptionHelper arms (surface/lemma; no invent) ---

	// на повну людей
	if adjPos > 1 && adj.GetToken() == "повну" && strings.EqualFold(tokens[adjPos-1].GetToken(), "на") {
		return true
	}
	// у Другій світовій (f: + f:)
	if adjPos > 1 &&
		HasLemmaWithPartPos(adj, []string{"світовий"}, ":f:") &&
		HasLemmaWithPartPos(tokens[adjPos-1], []string{"другий", "перший"}, ":f:") {
		return true
	}
	// знайдений увечері понеділка
	if nounPos > 1 &&
		HasLemmaTokenAny(tokens[nounPos-1], []string{"увечері", "уранці", "ввечері", "вранці"}) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_rod.*`)) {
		return true
	}
	// площею 100 кв. м / довжиною до 500
	if nounPos < len(tokens)-1 {
		nt := noun.GetToken()
		for _, w := range []string{"площею", "об'ємом", "довжиною", "висотою", "зростом"} {
			if nt == w && HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`prep.*|.*num.*`)) {
				return true
			}
		}
	}
	// 10 метрів квадратних води
	if adjPos > 1 &&
		HasLemmaTokenRE(tokens[adjPos-1], regexp.MustCompile(`.*метр.*`)) &&
		HasLemmaTokenRE(adj, regexp.MustCompile(`^(квадратний|кубічний)$`)) &&
		HasPosTagPart(noun, "v_rod") {
		return true
	}
	// 200% річних
	if adjPos > 1 && strings.HasSuffix(tokens[adjPos-1].GetToken(), "%") && adj.GetToken() == "річних" {
		return true
	}
	// пасли задніх / не мати рівних
	if adjPos > 1 {
		if HasLemmaToken(tokens[adjPos-1], "пасти") && adj.GetToken() == "задніх" {
			return true
		}
		if HasLemmaToken(tokens[adjPos-1], "мати") && adj.GetToken() == "рівних" {
			return true
		}
		// на манер
		if CleanTokenLower(noun) == "манер" && strings.EqualFold(tokens[adjPos-1].GetToken(), "на") {
			return true
		}
		// усі до єдиного
		if adjPos > 2 && adj.GetToken() == "єдиного" && tokens[adjPos-1].GetToken() == "до" &&
			HasLemmaWithPartPos(tokens[adjPos-2], []string{"весь", "увесь"}, ":p:") {
			return true
		}
		// порядок денний
		if HasLemmaToken(adj, "денний") && HasLemmaToken(tokens[adjPos-1], "порядок") {
			return true
		}
	}
	// сильні світу/миру (цього)
	if nounPos < len(tokens)-1 {
		nc := CleanTokenLower(noun)
		if nc == "миру" || nc == "світу" {
			if HasLemmaTokenAny(adj, []string{"сильний", "могутній", "великий"}) ||
				HasLemmaWithPartPos(tokens[nounPos+1], []string{"цей", "сей"}, ":m:v_rod") {
				return true
			}
		}
	}
	// колишня Маяковського
	if HasLemmaWithPosRE(adj, []string{"колишній", "тодішній", "теперішній", "нинішній"}, regexp.MustCompile(`adj.*:f:.*`)) &&
		isUpperFirst(noun.GetToken()) {
		return true
	}
	// імені / ім. / ордена
	if nounPos < len(tokens)-1 {
		nt := noun.GetToken()
		if nt == "ім." || nt == "імені" || nt == "ордена" {
			return true
		}
	}
	// на дівоче Анна
	if adj.GetToken() == "дівоче" && HasPosTagPart(noun, "name") {
		return true
	}
	// вольному/вільному воля
	al := strings.ToLower(adj.GetToken())
	if (al == "вольному" || al == "вільному") && noun.GetToken() == "воля" {
		return true
	}
	// здатний / змушений / … (Java list)
	if HasLemmaTokenAny(adj, []string{"здатний", "змушений", "винний", "повинний", "готовий", "спроможний"}) {
		return true
	}

	masterInfs := GetAdjCaseInflections(CollectPOSTags(adj))
	slaveInfs := GetNounCaseInflections(CollectPOSTags(noun))

	// молодшого гвардії сержанта
	if nounPos < len(tokens)-1 && noun.GetToken() == "гвардії" &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`^noun`)) &&
		InflectionsIntersect(masterInfs, GetNounCaseInflections(CollectPOSTags(tokens[nounPos+1]))) {
		return true
	}

	// князівством Литовським — prev noun + capitalized adj, inflections overlap
	if adjPos > 1 && HasPosTagPart(tokens[adjPos-1], "noun") && isUpperFirst(adj.GetToken()) &&
		InflectionsIntersect(masterInfs, GetNounCaseInflections(CollectPOSTags(tokens[adjPos-1]))) {
		return true
	}

	// абзац другий частини першої
	if adjPos > 1 && nounPos < len(tokens)-1 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:[mf]:.*numr.*|^number`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:inanim:.:v_rod.*`)) &&
		HasLemmaTokenAny(tokens[adjPos-1], []string{"абзац", "розділ", "пункт", "підпункт", "частина", "стаття"}) {
		return true
	}

	// протягом минулих травня – липня
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:p:`)) &&
		DashesPattern.MatchString(tokens[nounPos+1].GetToken()) &&
		HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`^(?:adj|noun)`)) &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
		return true
	}

	// зв'язаних ченця з черницею / на зарубаних матір з двома синами
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:p:`)) {
		mid := tokens[nounPos+1]
		if mid != nil {
			ml := CleanTokenLower(mid)
			if (ml == "з" || ml == "із" || ml == "зі") &&
				HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`(?:noun|numr).*:v_oru.*`)) &&
				InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
				return true
			}
		}
	}

	// на довгих півстоліття
	if HasPosTagRE(adj, regexp.MustCompile(`adj:p:v_rod.*`)) &&
		strings.HasPrefix(noun.GetToken(), "пів") &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_rod.*`)) {
		return true
	}
	// на довгих чверть століття
	if nounPos < len(tokens)-1 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:p:v_rod.*`)) &&
		noun.GetToken() == "чверть" &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*v_rod.*`)) {
		return true
	}
	// розділеного вже чверть століття / створених близько чверті століття
	if nounPos < len(tokens)-1 &&
		HasPosTagPart(adj, "adjp") &&
		HasLemmaTokenAny(noun, []string{"чверть", "третина"}) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*v_rod.*`)) {
		return true
	}
	// заклопотані чимало людей — adjp, quant before noun
	if HasPosTagPart(adj, "adjp") && nounPos > 0 &&
		HasLemmaTokenAny(tokens[nounPos-1], []string{"чимало", "багато", "небагато", "немало", "обмаль"}) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*:p:v_rod.*`)) {
		return true
	}

	// присудок ж.р. + професія ч.р.
	if containsStr([]string{"переконана", "впевнена", "упевнена", "годна", "ладна", "певна", "причетна", "обрана", "призначена"}, adj.GetToken()) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:m:v_naz.*`)) {
		return true
	}

	// чинних станом на
	if nounPos < len(tokens)-1 && noun.GetToken() == "станом" &&
		tokens[nounPos+1] != nil && tokens[nounPos+1].GetToken() == "на" {
		return true
	}

	// на таку Богом забуту
	if HasPosTagPart(adj, "pron") && strings.EqualFold(noun.GetCleanToken(), "богом") {
		return true
	}
	// той родом з
	if HasLemmaToken(adj, "той") &&
		containsStr([]string{"родом", "кулею", "розміром"}, CleanTokenLower(noun)) {
		return true
	}
	// такого світ ще не бачив
	if containsStr([]string{"таке", "такого"}, CleanTokenLower(adj)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*:v_naz.*`)) &&
		NewSearchMatch("").
			Target(ConditionPostag(regexp.MustCompile(`verb.*`))).
			WithLimit(2).
			Skip(ConditionPostag(regexp.MustCompile(`^(?:part|adv)`))).
			MAfterATR(tokens, nounPos+1) > 0 {
		return true
	}
	// той мантію надів
	if nounPos < len(tokens)-1 &&
		strings.EqualFold(CleanTokenLower(adj), "той") &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*:v_(zna|oru).*`)) &&
		HasPosTagStart(tokens[nounPos+1], "verb") {
		return true
	}
	// що таке звук
	if adjPos > 1 && adj.GetToken() == "таке" &&
		ReverseSearch(tokens, adjPos-1, 3, regexp.MustCompile(`^що$`), nil) {
		return true
	}
	// таких + p:v_naz / меншість|більшість
	if strings.EqualFold(adj.GetToken(), "таких") &&
		(HasPosTagPart(noun, ":p:v_naz") ||
			containsStr([]string{"меншість", "більшість"}, CleanTokenLower(noun))) {
		return true
	}
	// на рівних
	if adjPos > 1 && adj.GetToken() == "рівних" &&
		strings.EqualFold(tokens[adjPos-1].GetToken(), "на") {
		return true
	}
	// польські зразка 1620
	if nounPos < len(tokens)-1 && noun.GetToken() == "зразка" {
		return true
	}
	// три зелених плюс два
	if noun.GetToken() == "мінус" || noun.GetToken() == "плюс" {
		return true
	}

	// важкими пару років / неконституційними низку законів
	if nounPos < len(tokens)-1 &&
		HasLemmaTokenAny(noun, []string{"пара", "низка", "ряд", "купа", "більшість", "десятка", "сотня", "тисяча", "мільйон"}) &&
		(HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*?:p:v_rod.*`)) ||
			(nounPos < len(tokens)-2 &&
				HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`adj:p:v_rod.*`)) &&
				HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`noun.*?:p:v_rod.*`)))) {
		return true
	}

	// разів (у) десять
	if nounPos < len(tokens)-1 &&
		HasLemmaWithPosRE(noun, []string{"раз"}, regexp.MustCompile(`.*p:v_(naz|rod).*`)) &&
		(HasPosTagRE(tokens[nounPos+1], adjNounNumberVNazRE) ||
			HasPosTagPart(tokens[nounPos+1], "prep")) {
		return true
	}

	// років 6, відсотків зо два
	if nounPos < len(tokens)-1 &&
		HasLemmaWithPosRE(noun, TimePlusLemmaList(), regexp.MustCompile(`noun.*?p:v_(naz|rod).*`)) {
		if HasPosTagRE(tokens[nounPos+1], adjNounNumberVNazRE) {
			return true
		}
		if nounPos < len(tokens)-2 &&
			HasLemmaWithPartPos(tokens[nounPos+1], []string{"на", "за", "з", "із", "зо", "через", "під"}, "prep") &&
			HasPosTagRE(tokens[nounPos+2], adjNounNumberVNazRE) {
			return true
		}
	}

	// осіб на 30
	if nounPos < len(tokens)-2 &&
		HasLemmaWithPosRE(noun, []string{"особа"}, regexp.MustCompile(`noun.*?p:v_(naz|rod).*`)) &&
		HasLemmaWithPartPos(tokens[nounPos+1], []string{"на", "з", "із", "зо", "під"}, "prep") &&
		HasPosTagRE(tokens[nounPos+2], adjNounNumberVNazRE) {
		return true
	}

	// хвилини з 55-ї — prep case gov on both time lemma and num adj
	if adjPos > 2 &&
		HasLemmaTokenAny(tokens[adjPos-2], TimeLemmasShort) &&
		HasPosTagStart(tokens[adjPos-1], "prep") &&
		HasPosTagPart(adj, "num") {
		govs := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(tokens[adjPos-1], "prep")
		if len(govs) > 0 {
			var list []string
			for c := range govs {
				list = append(list, c)
			}
			if HasVidmPosTag(list, tokens[adjPos-2]) && HasVidmPosTag(list, adj) {
				return true
			}
		}
	}

	// predic + verb inf/past:n/futr
	if nounPos < len(tokens)-1 && HasPosTagPart(noun, "predic") {
		afterPred := regexp.MustCompile(`.*(?:inf|past:n|futr:s:3).*`)
		if HasPosTagRE(tokens[nounPos+1], afterPred) {
			return true
		}
		if nounPos < len(tokens)-2 &&
			HasPosTagStart(tokens[nounPos+1], "adv") &&
			HasPosTagRE(tokens[nounPos+2], afterPred) {
			return true
		}
	}

	// моїх маму й сестер
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:p:`)) &&
		forwardConjFind(tokens, nounPos+1, 2) &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "p", "") {
		return true
	}

	// навчальної та середньої шкіл
	if adjPos > 2 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*:p:`)) &&
		(reverseConjFind(tokens, adjPos-1, 3) || reverseConjAdvFind(tokens, adjPos-1, 3)) &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "p") &&
		// Java hasPosTag Pattern matches full tag → need .* after prefix
		ReverseSearch(tokens, adjPos-2, 100, nil, regexp.MustCompile(`^(?:adj|numr).*`)) {
		return true
	}

	// Большого та Маріїнського театрів / 3, 4 і 5-ї категорій
	if adjPos > 2 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*:p:`)) &&
		reverseConjFind2(tokens, adjPos-1, 3) &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "p") {
		return true
	}

	// ні у методологічному, ні у практичному аспектах
	if adjPos > 6 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*:p:`)) &&
		HasPosTagRE(adj, regexp.MustCompile(`^adj:`)) &&
		HasPosTagStart(tokens[adjPos-1], "prep") &&
		HasLemmaTokenAny(tokens[adjPos-2], []string{"ні", "ані", "хоч", "що", "як"}) &&
		tokens[adjPos-3] != nil && tokens[adjPos-3].GetToken() == "," &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
		return true
	}

	// коринфський з іонійським ордери
	if adjPos > 2 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*:p:`)) &&
		regexp.MustCompile(`^(?:з|із|зі)$`).MatchString(CleanTokenLower(tokens[adjPos-1])) &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) &&
		InflectionsIntersectIgnoreGender(
			GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-2])), slaveInfs, "", "") {
		return true
	}

	// пофарбований рік тому
	if nounPos < len(tokens)-1 &&
		HasLemmaTokenAny(noun, TimeLemmas) &&
		HasLemmaToken(tokens[nounPos+1], "тому") {
		return true
	}
	// замість звичного десятиліттями
	if nounPos < len(tokens)-1 &&
		HasLemmaWithPosRE(noun, TimePlusLemmaList(), regexp.MustCompile(`noun:inanim:p:v_oru.*`)) {
		return true
	}

	// кількох десятих відсотка
	if HasLemmaTokenAny(adj, []string{"десятий", "сотий", "тисячний", "десятитисячний", "стотитисячний", "мільйонний", "мільярдний"}) &&
		HasPosTagRE(adj, regexp.MustCompile(`.*:[fp]:.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_rod.*`)) {
		return true
	}

	// два нових горнятка / 33 народних обранці
	if adjPos > 1 &&
		HasPosTagRE(adj, regexp.MustCompile(`.*:p:v_(rod|naz).*`)) &&
		ReverseSearch(tokens, adjPos-1, 5, DovyeTroyeRE, nil) &&
		(HasPosTagRE(noun, regexp.MustCompile(`.*(?:p:v_naz|:n:v_rod).*`)) ||
			containsStr([]string{"імені", "ока"}, noun.GetToken())) {
		return true
	}

	// 1-3-й класи / на сьомому–восьмому поверхах
	cleanAdj := adj.GetCleanToken()
	if cleanAdj == "" {
		cleanAdj = adj.GetToken()
	}
	if (regexp.MustCompile(`^[0-9]+[\x{2014}\x{2013}-][0-9]+[\x{2013}-][а-яіїєґ]{1,3}$`).MatchString(cleanAdj) ||
		(regexp.MustCompile(`.*[а-яїієґ][\x{2014}\x{2013}-].*`).MatchString(cleanAdj) && HasPosTagPart(adj, "numr"))) &&
		HasPosTagPart(noun, ":p:") &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
		return true
	}
	// восьмого – дев’ятого класів
	if nounPos > 2 && adjPos > 1 &&
		containsStr([]string{"\u2013", "\u2014"}, tokens[adjPos-1].GetToken()) &&
		HasPosTagPart(adj, "num") && HasPosTagPart(tokens[adjPos-2], "num") &&
		HasPosTagPart(noun, ":p:") &&
		(HasPosTagStart(tokens[adjPos-2], "number") ||
			InflectionsIntersectIgnoreGender(GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-2])), slaveInfs, "", "")) &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
		return true
	}

	// найближчі рік-два
	if HasPosTagRE(adj, regexp.MustCompile(`adj.*:p:`)) &&
		regexp.MustCompile(`.*[\x{2014}\x{2013}-].*`).MatchString(noun.GetToken()) {
		lemma0 := lemmaOf(noun)
		base := strings.Split(lemma0, "\u2014")[0]
		base = strings.Split(base, "\u2013")[0]
		base = strings.Split(base, "-")[0]
		if IsTimePlusLemma(base) || InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
			return true
		}
	}

	// Від наступних пари десятків
	if nounPos < len(tokens)-1 &&
		HasLemmaToken(noun, "пара") &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*:p:`)) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`.*:p:v_rod.*`)) {
		return true
	}

	// п'ять шостих / одній восьмій
	if nounPos > 1 && adjPos > 0 &&
		HasPosTagPart(tokens[adjPos-1], "num") &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*num.*`)) {
		if HasPosTagRE(tokens[adjPos-1], regexp.MustCompile(`^(?:noun|numr)`)) &&
			HasPosTagRE(adj, regexp.MustCompile(`adj:p:v_rod.*`)) {
			if HasLemmaToken(adj, "другий") && !HasLemmaToken(tokens[adjPos-1], "один") {
				// Java: return false (not exception)
			} else {
				return true
			}
		}
		if HasLemmaWithPosRE(tokens[adjPos-1], []string{"один"}, regexp.MustCompile(`numr:f:.*`)) &&
			InflectionsIntersect(
				GetNumrCaseInflections(CollectPOSTags(tokens[adjPos-1])),
				GetAdjCaseInflections(CollectPOSTags(adj))) {
			return true
		}
	}

	// 1/8-ї фіналу
	if nounPos > 3 && adjPos > 1 &&
		tokens[adjPos-1] != nil && tokens[adjPos-1].GetToken() == "/" &&
		HasPosTagPart(tokens[adjPos-2], "numb") &&
		InflectionsIntersectIgnoreGender(masterInfs, slaveInfs, "", "") {
		return true
	}

	// dates with :numr
	if HasPosTagPart(adj, ":numr") {
		at := adj.GetToken()
		if regexp.MustCompile(`^(?:[12][0-9])?[0-9][0-9][\x{2014}\x{2013}-](?:й|го|м|му)$`).MatchString(at) ||
			regexp.MustCompile(`^(?:[12][0-9])?[0-9]0[\x{2014}\x{2013}-](?:ті|тих|их|х)$`).MatchString(at) ||
			regexp.MustCompile(`^(?:[12][0-9])?[0-9][0-9][\x{2014}\x{2013}-](?:[12][0-9])?[0-9][0-9][\x{2014}\x{2013}-](?:й|го|м|му|ті|тих|их|х)$`).MatchString(at) {
			return true
		}
		if adjPos > 1 && HasPosTagPart(adj, ":f:") &&
			HasLemmaTokenAny(tokens[adjPos-1], []string{"на", "в", "у", "за", "о", "до", "після", "близько", "раніше"}) &&
			!HasLemmaTokenAny(noun, []string{"хвилина", "година"}) {
			return true
		}
		if HasPosTagPart(adj, ":f:") &&
			regexp.MustCompile(`^(?:ранку|дня|вечора|ночі|пополудня)$`).MatchString(noun.GetToken()) {
			return true
		}
		// дев'яте травня — Java hasLemma(…, MONTH_LEMMAS, ":v_rod") part-contains
		if HasPosTagPart(adj, ":n:") &&
			HasLemmaWithPartPos(noun, MonthLemmas, ":v_rod") {
			return true
		}
	}

	// обмежуючий власність — adjp:actv:bad
	if HasPosTagRE(adj, regexp.MustCompile(`.*?adjp:actv.*:bad.*`)) {
		return true
	}

	// нічого протизаконного / щось подібне
	if nounPos > 2 && nounPos <= len(tokens)-1 && adjPos > 0 &&
		HasLemmaTokenAny(tokens[adjPos-1], []string{"ніщо", "щось", "ніхто", "хтось"}) &&
		InflectionsIntersect(GetNounCaseInflections(CollectPOSTags(tokens[adjPos-1])), masterInfs) {
		return true
	}

	// визнання неконституційним закону
	if adjPos > 1 &&
		RevSearch(tokens, adjPos-1, regexp.MustCompile(`.*(ння|ття)$`), "") &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*:v_oru.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*:v_rod.*`)) &&
		GenderMatches(masterInfs, slaveInfs, "v_oru", "v_rod") {
		return true
	}

	// бути/стати/… + adjp:pasv / predicative oru
	verbPos := RevSearchIdx(tokens, adjPos-1, regexp.MustCompile(`^(?:бути|ставати|стати|залишатися|залишитися)$`), "")
	if verbPos != -1 {
		if HasPosTagRE(adj, regexp.MustCompile(`adj.*v_naz.*adjp:pasv.*`)) {
			if GenderMatches(masterInfs, slaveInfs, "v_naz", "v_naz") {
				return true
			}
		} else if HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) {
			if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_naz.*`)) {
				if GenderMatches(masterInfs, slaveInfs, "v_oru", "v_naz") {
					if HasPosTagPart(tokens[verbPos], ":inf") ||
						VerbInflectionsOverlap(CollectPOSTags(tokens[verbPos]), CollectPOSTags(noun)) {
						return true
					}
				} else if nounPos < len(tokens)-1 &&
					HasPosTagPart(adj, "adj:p:") &&
					isConjForPluralWithComma(tokens[nounPos+1]) {
					return true
				}
			} else if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_dav.*`)) {
				if GenderMatches(masterInfs, slaveInfs, "v_oru", "v_dav") {
					return true
				}
			}
		}
	}

	// визнали справедливою наставники
	verbPos = RevSearchIdx(tokens, adjPos-1, nil, "verb.*")
	if verbPos != -1 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_naz.*`)) &&
		VerbInflectionsOverlap(CollectPOSTags(tokens[verbPos]), CollectPOSTags(noun)) {
		return true
	}

	// помальована в біле кімната
	colorTokens := []string{"біле", "чорне", "оранжеве", "червоне", "жовте", "синє", "зелене", "фіолетове"}
	if adjPos > 2 &&
		containsStr(colorTokens, adj.GetToken()) &&
		containsStr([]string{"в", "у"}, tokens[adjPos-1].GetToken()) &&
		HasPosTagPart(tokens[adjPos-2], "adjp:pasv") {
		prevAdj := GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-2]))
		if InflectionsIntersect(prevAdj, slaveInfs) {
			return true
		}
	}
	if adjPos > 3 &&
		containsStr([]string{"біле", "чорне"}, adj.GetToken()) &&
		containsStr([]string{"усе", "все"}, tokens[adjPos-1].GetToken()) &&
		containsStr([]string{"в", "у"}, tokens[adjPos-2].GetToken()) &&
		HasPosTagPart(tokens[adjPos-3], "adjp:pasv") {
		prevAdj := GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-3]))
		if InflectionsIntersect(prevAdj, slaveInfs) {
			return true
		}
	}

	// повторена тисячу разів
	if nounPos < len(tokens)-1 &&
		HasPosTagPart(adj, "adjp:pasv") &&
		containsStr([]string{"тисячу", "сотню", "десятки"}, noun.GetToken()) &&
		containsStr([]string{"разів", "раз", "років"}, tokens[nounPos+1].GetToken()) {
		return true
	}
	// покликана ще раз
	if nounPos > 0 &&
		strings.EqualFold(noun.GetCleanToken(), "раз") &&
		strings.EqualFold(tokens[nounPos-1].GetToken(), "ще") {
		return true
	}

	// порівняно з попереднім / аналогічно з …
	if adjPos > 2 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) &&
		HasLemmaTokenAny(tokens[adjPos-2], []string{"порівняно", "аналогічно"}) &&
		HasLemmaTokenRE(tokens[adjPos-1], regexp.MustCompile(`^(?:з|із|зі)$`)) {
		return true
	}

	// наближена до сімейної форма — prep before adj
	if adjPos > 2 && HasPosTagPart(tokens[adjPos-1], "prep") &&
		HasPosTagRE(tokens[adjPos-2], regexp.MustCompile(`^(?:adj|verb|part|noun|adv)`)) {
		govs := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(tokens[adjPos-1], "prep")
		if len(govs) > 0 {
			var list []string
			for c := range govs {
				list = append(list, c)
			}
			if HasVidmPosTag(list, adj) {
				// відрізнялася (б) від нинішньої ситуація / поряд / відміну
				prev2 := tokens[adjPos-2]
				if (HasPosTagStart(prev2, "verb") || HasLemmaTokenAny(prev2, []string{"би", "б"}) ||
					containsStr([]string{"поряд", "відміну", "порівнянні"}, CleanTokenLower(prev2))) &&
					HasPosTagRE(noun, regexp.MustCompile(`noun.*v_(naz|zna|oru).*`)) {
					return true
				}
				prevAdjInfs := GetAdjCaseInflections(CollectPOSTags(prev2))
				if InflectionsIntersect(prevAdjInfs, slaveInfs) {
					return true
				}
				// тотожні із загальносоюзними герб і прапор
				if nounPos < len(tokens)-1 &&
					HasPosTagPart(adj, "adj:p:") &&
					isConjForPluralWithComma(tokens[nounPos+1]) &&
					HasPosTagPart(prev2, "adj:p:") &&
					InflectionsIntersectIgnoreGender(
						GetAdjCaseInflections(CollectPOSTags(prev2)), slaveInfs, "p", "") {
					return true
				}
			}
		}
	}

	// підсвічений синім діамант — adjp:pasv + adj:v_oru
	if adjPos > 1 &&
		HasPosTagPart(tokens[adjPos-1], "adjp:pasv") &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) &&
		InflectionsIntersect(GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-1])), slaveInfs) {
		return true
	}

	// захищені законом / Змучений тягарем — adjp:pasv + noun v_oru
	if HasPosTagPart(adj, "adjp:pasv") && HasPosTagPart(noun, "v_oru") {
		return true
	}

	// Найнижчою частка таких є… — adj v_oru + noun v_zna|naz + nearby verb
	if adjPos > 0 &&
		!HasPosTagRE(tokens[adjPos-1], regexp.MustCompile(`.*adjp:pasv.*|^prep`)) &&
		HasPosTagRE(adj, regexp.MustCompile(`adj.*v_oru.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_(zna|naz).*`)) {
		vPos := TokenSearch(tokens, nounPos+1, "verb", nil, nil, DirForward)
		if vPos > 0 && vPos <= nounPos+5 {
			if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_naz.*`)) ||
				(hasCaseGovFromReadings(tokens[vPos], "v_zna") &&
					GenderMatches(masterInfs, slaveInfs, "v_oru", "v_zna")) {
				return true
			}
		}
	}

	// зроблять неможливою ротацію — verb/advp reverse + dual case gov
	if adjPos > 1 &&
		HasPosTagRE(adj, regexp.MustCompile(`adj:.:v_oru.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`.*v_zna.*`)) &&
		GenderMatches(masterInfs, slaveInfs, "v_oru", "v_zna") {
		vPos := tokenSearchPosRE(tokens, adjPos-1, verbAdvpPattern, DirReverse)
		if vPos > 0 && vPos >= adjPos-3 {
			if hasCaseGovPosRE(tokens[vPos], verbAdvpPattern, "v_oru") &&
				hasCaseGovPosRE(tokens[vPos], verbAdvpPattern, "v_zna") {
				return true
			}
		}
	}

	// робив неймовірно високими шанси — verb + adv + adj:v_oru + noun:v_zna
	if adjPos > 2 &&
		hasAdvNotAdvp(tokens[adjPos-1]) &&
		hasCaseGovPosRE(tokens[adjPos-2], verbAdvpPattern, "v_oru") &&
		HasPosTagPart(adj, "v_oru") &&
		HasPosTagRE(noun, regexp.MustCompile(`.*v_zna.*`)) &&
		GenderMatches(masterInfs, slaveInfs, "v_oru", "v_zna") {
		return true
	}

	// case government of adj on slave noun (вдячний редакторові …)
	if caseGovernmentMatches(adj, slaveInfs) {
		if nounPos < len(tokens)-1 && HasPosTagPart(tokens[nounPos+1], "noun:") {
			if HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*v_(rod|oru|naz|dav).*`)) {
				return true
			}
			slave2 := GetNounCaseInflections(CollectPOSTags(tokens[nounPos+1]))
			if InflectionsIntersect(masterInfs, slave2) {
				return true
			}
		} else {
			return true
		}
	}

	// альтернативну олігархічній модель — prev adj governs current adj cases
	if adjPos > 1 && HasPosTagPart(tokens[adjPos-1], "adj") &&
		caseGovernmentMatches(tokens[adjPos-1], masterInfs) {
		preAdj := GetAdjCaseInflections(CollectPOSTags(tokens[adjPos-1]))
		if InflectionsIntersect(preAdj, slaveInfs) {
			return true
		}
	}

	return false
}

// verbAdvpPattern ports VERB_ADVP_PATTERN = (verb|advp).*
var verbAdvpPattern = regexp.MustCompile(`^(?:verb|advp).*`)

// caseGovernmentMatches ports TokenAgreementAdjNounExceptionHelper.caseGovernmentMatches.
func caseGovernmentMatches(adj *languagetool.AnalyzedTokenReadings, slave []Inflection) bool {
	if adj == nil || len(slave) == 0 {
		return false
	}
	cg := LoadCaseGovernmentHelper()
	seen := map[string]struct{}{}
	for _, r := range adj.GetReadings() {
		if r == nil || r.GetLemma() == nil {
			continue
		}
		lemma := *r.GetLemma()
		if _, ok := seen[lemma]; ok {
			continue
		}
		seen[lemma] = struct{}{}
		for _, s := range slave {
			if cg.HasCaseGovernment(lemma, s.Case) {
				return true
			}
		}
	}
	return false
}

// tokenSearchIgnoreLatinPOS ports the common Java posTagsToIgnore Pattern.compile("[a-z].*"):
// skip-over tags starting with latin letters; break on other (e.g. untagged) tokens.
var tokenSearchIgnoreLatinPOS = regexp.MustCompile(`^[a-z].*`)

// tokenSearchPosRE is TokenSearch with a POS regex (Java tokenSearch Pattern posTag + ignore).
// Uses TokenSearchPosRE so ignore semantics match LemmaHelper (skip-over vs break).
func tokenSearchPosRE(tokens []*languagetool.AnalyzedTokenReadings, pos int, posRE *regexp.Regexp, dir Dir) int {
	if tokens == nil || pos < 0 || posRE == nil {
		return -1
	}
	// Java VerbNoun paths: tokenSearch(..., VERB_PATTERN, null, [a-z].*, dir)
	return TokenSearchPosRE(tokens, pos, posRE, nil, tokenSearchIgnoreLatinPOS, dir)
}

// hasCaseGovPosRE ports hasCaseGovernment(readings, posPattern, case).
func hasCaseGovPosRE(tok *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp, rvCase string) bool {
	if tok == nil || posRE == nil || rvCase == "" {
		return false
	}
	cg := LoadCaseGovernmentHelper()
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil || r.GetLemma() == nil {
			continue
		}
		if !posRE.MatchString(*r.GetPOSTag()) {
			continue
		}
		if cg.HasCaseGovernment(*r.GetLemma(), rvCase) {
			return true
		}
		// adjp:pasv always adds v_oru
		if rvCase == "v_oru" && strings.Contains(*r.GetPOSTag(), "adjp:pasv") {
			return true
		}
	}
	return false
}

// adjNounNumberVNazRE ports TokenAgreementAdjNounExceptionHelper.NUMBER_V_NAZ.
var adjNounNumberVNazRE = regexp.MustCompile(`^(?:number|numr:p:v_naz|noun.*?:p:v_naz.*:numr.*)$`)

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// HasPosTagRE reports whether any POS matches re (Java PosTagHelper.hasPosTag Pattern).
// HasPosTagRE reports whether any POS tag matches re (MatchString / find).
// For Java Matcher.matches() (full tag), call HasPosTagMatches or write re with ^…$ / .* ends.
func HasPosTagRE(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

// HasPosTagMatches ports PosTagHelper.hasPosTag(…, Pattern) — Matcher.matches() full tag.
func HasPosTagMatches(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		loc := re.FindStringIndex(p)
		if loc != nil && loc[0] == 0 && loc[1] == len(p) {
			return true
		}
	}
	return false
}

// HasLemmaTokenRE ports LemmaHelper.hasLemma(readings, lemmaRegex).
// HasLemmaTokenRE ports LemmaHelper.hasLemma(readings, Pattern) — Matcher.matches() on lemma.
func HasLemmaTokenRE(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetLemma() == nil {
			continue
		}
		lem := *r.GetLemma()
		loc := re.FindStringIndex(lem)
		if loc != nil && loc[0] == 0 && loc[1] == len(lem) {
			return true
		}
	}
	return false
}

func isUpperFirst(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsUpper(r)
}

// HasPosTagPart reports whether any POS contains substr (Java PosTagHelper.hasPosTagPart).
func HasPosTagPart(tok *languagetool.AnalyzedTokenReadings, substr string) bool {
	if tok == nil || substr == "" {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}

// prepNounTimeAdvRE* port getExceptionStrong time-adverb full-string matches.
var (
	prepNounTimeDoPoRE = regexp.MustCompile(
		`^(?:сьогодні|[ву]чора|позавчора|(?:після)?завтра|тепер|зараз|нині|опівдня|опівночі|досі|навпаки)$`)
	prepNounTimeNaVidProRE = regexp.MustCompile(
		`^(?:сьогодні|[ву]чора|позавчора|(?:після)?завтра|тепер|зараз|нині|тоді|потім|щодень|повсякдень)$`)
	prepNounTimeZaZRE = regexp.MustCompile(
		`^(?:сьогодні|[ву]чора|позавчора|(?:після)?завтра)$`)
	prepNounLishRE = regexp.MustCompile(`^лиш(?:е(?:нь)?)?$`)
)

// hasAdvNotAdvp is RE2-friendly stand-in for Java adv(?!p).*.
func hasAdvNotAdvp(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, "adv") && !strings.HasPrefix(p, "advp") {
			return true
		}
	}
	return false
}

// hasPosTagAllAdvNotAdvp ports hasPosTagAll(readings, adv(?!p).*).
func hasPosTagAllAdvNotAdvp(tok *languagetool.AnalyzedTokenReadings) bool {
	tags := CollectPOSTags(tok)
	if len(tags) == 0 {
		return false
	}
	for _, p := range tags {
		if !strings.HasPrefix(p, "adv") || strings.HasPrefix(p, "advp") {
			return false
		}
	}
	return true
}

// IsPrepNounException ports TokenAgreementPrepNounExceptionHelper as a boolean
// (any Strong/NonInfl/Infl hit). Invalid index layout is treated as exception (no flag).
func IsPrepNounException(tokens []*languagetool.AnalyzedTokenReadings, prepPos, nounPos int) bool {
	if prepPos < 0 || nounPos <= prepPos || prepPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	if GetPrepNounExceptionStrong(tokens, nounPos, tokens[prepPos]).Type != RuleExceptionNone {
		return true
	}
	if GetPrepNounExceptionNonInfl(tokens, nounPos).Type != RuleExceptionNone {
		return true
	}
	if GetPrepNounExceptionInfl(tokens, prepPos, nounPos).Type != RuleExceptionNone {
		return true
	}
	return false
}

// GetPrepNounExceptionStrong ports getExceptionStrong (time/inserts; Skip keeps prep state).
func GetPrepNounExceptionStrong(tokens []*languagetool.AnalyzedTokenReadings, i int, prepTok *languagetool.AnalyzedTokenReadings) RuleException {
	if tokens == nil || i < 0 || i >= len(tokens) || tokens[i] == nil || prepTok == nil {
		return NewRuleException(RuleExceptionNone)
	}
	noun := tokens[i]
	prepLower := CleanTokenLower(prepTok)
	nounClean := noun.GetCleanToken()
	if nounClean == "" {
		nounClean = noun.GetToken()
	}
	nounLower := strings.ToLower(nounClean)

	if prepLower == "до" || prepLower == "по" {
		if prepNounTimeDoPoRE.MatchString(nounLower) {
			return NewRuleException(RuleExceptionException)
		}
	}
	if prepLower == "на" || prepLower == "від" || prepLower == "про" {
		if prepNounTimeNaVidProRE.MatchString(nounLower) {
			return NewRuleException(RuleExceptionException)
		}
	}
	if prepLower == "за" || prepLower == "з" || prepLower == "зі" || prepLower == "із" {
		if prepNounTimeZaZRE.MatchString(nounLower) {
			return NewRuleException(RuleExceptionException)
		}
	}
	if (prepLower == "в" || prepLower == "у") && nounLower == "нікуди" {
		return NewRuleException(RuleExceptionException)
	}
	// до не властиву — Java RuleException(0) skip
	if nounClean == "не" && i < len(tokens)-1 && HasPosTagStart(tokens[i+1], "ad") {
		return NewRuleExceptionSkip(0)
	}
	// про чимало
	if HasLemmaTokenRE(noun, AdvQuantPattern) && HasPosTagRE(noun, regexp.MustCompile(`^adv`)) {
		return NewRuleException(RuleExceptionException)
	}
	// лежить із сотня
	if (prepLower == "з" || prepLower == "зі" || prepLower == "із") && HasLemmaTokenAny(noun, PseudoNumLemmas) {
		return NewRuleException(RuleExceptionException)
	}
	// adv all (not advp)
	if hasPosTagAllAdvNotAdvp(noun) {
		if i < len(tokens)-1 && tokens[i+1] != nil && tokens[i+1].GetCleanToken() == "собі" {
			return NewRuleExceptionSkip(1)
		}
		return NewRuleExceptionSkip(0)
	}
	// замість … inf within 4
	if prepLower == "замість" {
		if NewSearchMatch("").
			Target(ConditionPostag(verbInfPattern)).
			WithLimit(4).
			Skip(ConditionToken("можна").WithNegate()).
			MAfterATR(tokens, i+1) > 0 {
			return NewRuleException(RuleExceptionException)
		}
	}
	if NewSearchMatch("не те").MBeforeATR(tokens, i) > 0 {
		return NewRuleException(RuleExceptionException)
	}
	return NewRuleException(RuleExceptionNone)
}

// GetPrepNounExceptionNonInfl ports getExceptionNonInfl.
func GetPrepNounExceptionNonInfl(tokens []*languagetool.AnalyzedTokenReadings, i int) RuleException {
	if tokens == nil || i < 0 || i >= len(tokens) || tokens[i] == nil {
		return NewRuleException(RuleExceptionNone)
	}
	noun := tokens[i]
	nounClean := noun.GetCleanToken()
	if nounClean == "" {
		nounClean = noun.GetToken()
	}
	nounLower := strings.ToLower(nounClean)

	if HasPosTagStart(noun, "part") && PartInsertPattern.MatchString(nounLower) {
		return NewRuleExceptionSkip(0)
	}
	if prepNounLishRE.MatchString(nounLower) || nounLower == "наприклад" {
		return NewRuleExceptionSkip(0)
	}
	if hasAdvNotAdvp(noun) {
		if i < len(tokens)-1 && HasPosTagStart(tokens[i+1], "adj") && hasPosTagPartAll(noun, "adv") {
			return NewRuleExceptionSkip(0)
		}
		return NewRuleException(RuleExceptionException)
	}
	if i < len(tokens)-1 {
		if HasPosTagRE(noun, regexp.MustCompile(`noun:(?:un)?anim:.:v_dav.*:pron.*`)) {
			next := tokens[i+1]
			if HasPosTagStart(next, "adj") && hasCaseGovFromReadings(next, "v_dav") {
				return NewRuleExceptionSkip(1)
			}
			if i < len(tokens)-2 &&
				HasPosTagStart(next, "adv") &&
				HasPosTagStart(tokens[i+2], "adj") &&
				hasCaseGovFromReadings(tokens[i+2], "v_dav") {
				return NewRuleExceptionSkip(2)
			}
		}
	}
	if i < len(tokens)-2 &&
		nounClean == "нічого" &&
		tokens[i+1] != nil && tokens[i+1].GetToken() == "не" &&
		HasPosTagStart(tokens[i+2], "adj") {
		return NewRuleExceptionSkip(1)
	}
	return NewRuleException(RuleExceptionNone)
}

// GetPrepNounExceptionInfl ports getExceptionInfl (after hasVidm fails).
func GetPrepNounExceptionInfl(tokens []*languagetool.AnalyzedTokenReadings, prepPos, i int) RuleException {
	if tokens == nil || prepPos < 0 || prepPos >= len(tokens) || i <= prepPos || i >= len(tokens) {
		return NewRuleException(RuleExceptionNone)
	}
	prep, noun := tokens[prepPos], tokens[i]
	if prep == nil || noun == nil {
		return NewRuleException(RuleExceptionNone)
	}
	prepLower := CleanTokenLower(prep)
	nounClean := noun.GetCleanToken()
	if nounClean == "" {
		nounClean = noun.GetToken()
	}
	nounLower := strings.ToLower(nounClean)

	if nounClean == "дивом" {
		return NewRuleExceptionSkip(0)
	}
	if i < len(tokens)-1 && nounClean == "тисяча" {
		next := tokens[i+1]
		if HasPosTagPart(next, "numr") || HasLemmaToken(next, "якийсь") {
			return NewRuleExceptionSkip(0)
		}
	}
	if i < len(tokens)-1 &&
		HasPosTagPart(noun, "numr") && HasPosTagPart(noun, "v_naz") &&
		HasPosTagPart(tokens[i+1], "numr") &&
		HasPosTagRE(noun, regexp.MustCompile(`.*v_(rod|dav|zna|oru|mis).*`)) {
		return NewRuleExceptionSkip(1)
	}

	if prepLower == "на" {
		if IsCapitalized(nounClean) && HasPosTagRE(noun, regexp.MustCompile(`noun.*?:.:v_rod.*`)) {
			return NewRuleException(RuleExceptionException)
		}
		if HasPosTagRE(noun, regexp.MustCompile(`noun:anim:.:v_naz:prop:[fl]name.*`)) {
			if prepPos > 1 && (isNameToken(tokens[prepPos-1]) || (prepPos > 2 && isNameLemma(tokens[prepPos-2]))) {
				return NewRuleException(RuleExceptionException)
			}
		}
		if nounLower == "ти" || nounLower == "ви" {
			return NewRuleException(RuleExceptionException)
		}
		if i < len(tokens)-1 && nounClean == "Піп" && tokens[i+1] != nil &&
			tokens[i+1].GetCleanToken() == "Іван" {
			return NewRuleException(RuleExceptionException)
		}
		if nounLower == "манер" {
			return NewRuleException(RuleExceptionException)
		}
	}
	if prepPos > 0 && prepLower == "заради" {
		prev := CleanTokenLower(tokens[prepPos-1])
		if regexp.MustCompile(`^(справедливості|об.єктивності)$`).MatchString(prev) {
			return NewRuleException(RuleExceptionException)
		}
	}
	if prepLower == "при" && nounClean == "їх" {
		return NewRuleException(RuleExceptionSkip) // Type.skip, Skip 0
	}
	if prepLower == "з" && nounClean == "рана" {
		return NewRuleException(RuleExceptionException)
	}
	if prepLower == "від" {
		if strings.EqualFold(nounClean, "а") || nounClean == "рана" || nounClean == "корки" || nounClean == "мала" {
			return NewRuleException(RuleExceptionException)
		}
	}
	if prepLower == "до" {
		if strings.EqualFold(nounClean, "я") || nounClean == "корки" || nounClean == "велика" {
			return NewRuleException(RuleExceptionException)
		}
	}

	if i < len(tokens)-1 {
		next := tokens[i+1]
		if (HasPosTagStart(next, "num") || (next != nil && next.GetToken() == "$")) &&
			IsPlusMinusLemma(nounLower) {
			return NewRuleException(RuleExceptionException)
		}
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*?:v_oru.*`)) &&
			next != nil && next.HasPartialPosTag("adjp:pasv") {
			return NewRuleExceptionSkip(1)
		}
		if nounClean == "святая" && next != nil && next.GetToken() == "святих" {
			return NewRuleException(RuleExceptionException)
		}
		if prepLower == "через" || prepLower == "на" {
			if HasLemmaWithPosRE(noun, TimePlusLemmaList(), regexp.MustCompile(`noun:inanim:p:v_(rod|zna).*`)) &&
				(next.HasPartialPosTag("num") ||
					(i < len(tokens)-2 &&
						HasLemmaTokenAny(next, []string{"зо", "з", "із"}) &&
						tokens[i+2] != nil && tokens[i+2].HasPartialPosTag("num"))) {
				return NewRuleException(RuleExceptionException)
			}
		}
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_dav.*:pron:(refl|pers).*`)) &&
			strings.HasPrefix(CleanTokenLower(next), "подібн") {
			return NewRuleExceptionSkip(0)
		}
		if (nounClean == "усім" || nounClean == "всім") && strings.HasPrefix(CleanTokenLower(next), "відом") {
			return NewRuleExceptionSkip(0)
		}
		if prepLower == "до" && nounClean == "схід" && next != nil && next.GetCleanToken() == "сонця" {
			return NewRuleException(RuleExceptionException)
		}
	}
	if i < len(tokens)-2 {
		if HasPosTagRE(noun, regexp.MustCompile(`adj:[mfn]:v_rod.*`)) {
			genders := gendersFromPos(noun, regexp.MustCompile(`adj:([mfn]):v_rod.*`))
			if genders != "" && HasPosTagRE(tokens[i+1], regexp.MustCompile(`noun.*?:[`+genders+`]:v_rod.*`)) {
				return NewRuleExceptionSkip(1)
			}
		}
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_(dav|oru).*:pron:neg.*`)) &&
			tokens[i+1] != nil && tokens[i+1].GetCleanToken() == "не" {
			return NewRuleException(RuleExceptionSkip)
		}
	}
	return NewRuleException(RuleExceptionNone)
}

// hasPosTagPartAll reports every POS tag contains substr (Java hasPosTagPartAll).
func hasPosTagPartAll(tok *languagetool.AnalyzedTokenReadings, substr string) bool {
	tags := CollectPOSTags(tok)
	if len(tags) == 0 || substr == "" {
		return false
	}
	for _, p := range tags {
		if !strings.Contains(p, substr) {
			return false
		}
	}
	return true
}

// IsPlusMinusLemma ports LemmaHelper.PLUS_MINUS membership on surface lower.
func IsPlusMinusLemma(tokenLower string) bool {
	switch tokenLower {
	case "плюс", "мінус", "максимум", "мінімум":
		return true
	}
	return false
}

// TimePlusLemmaList returns TIME_PLUS_LEMMAS as a slice for HasLemmaWithPosRE.
func TimePlusLemmaList() []string {
	out := make([]string, 0, len(TimePlusLemmas))
	for s := range TimePlusLemmas {
		out = append(out, s)
	}
	return out
}

// HasPosTagStart ports PosTagHelper.hasPosTagStart (any reading starts with prefix).
func HasPosTagStart(tok *languagetool.AnalyzedTokenReadings, prefix string) bool {
	if tok == nil || prefix == "" {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// hasPosWithoutPron is RE2-friendly stand-in for Java (?!.*pron) POS patterns.
func hasPosWithoutPron(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "pron") {
			continue
		}
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

// gendersFromPos collects gender letters matching re with one capture group [mfn].
func gendersFromPos(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) string {
	if tok == nil || re == nil {
		return ""
	}
	seen := map[byte]bool{}
	var b strings.Builder
	for _, p := range CollectPOSTags(tok) {
		m := re.FindStringSubmatch(p)
		if len(m) < 2 {
			continue
		}
		g := m[1]
		if g == "" {
			continue
		}
		c := g[0]
		if !seen[c] {
			seen[c] = true
			b.WriteByte(c)
		}
	}
	return b.String()
}

func isNameToken(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	t := tok.GetToken()
	return t == "ім'я" || t == "прізвище"
}

func isNameLemma(tok *languagetool.AnalyzedTokenReadings) bool {
	return HasLemmaTokenAny(tok, []string{"ім'я", "прізвище"})
}

// numrDva34Pattern ports TokenAgreementNumrNounRule.DVA_3_4_PATTERN (full-string).
var numrDva34Pattern = regexp.MustCompile(`^(?:оби(?:два|дві)|(?:.+-)?(?:(?:два|дві)|три|чотири))$`)

// numrManyObPattern / numrNounSoftSurface / numrLemmaSoftRE — Java String.matches surfaces.
var (
	numrManyObPattern = regexp.MustCompile(
		`^(?:багать(?:ох|ом|ма)|обо(?:х|м|ма)|(?:дв|трь|чотирь)о[хм]|скільки(?:сь)?(?:-небудь)?|стільки)$`)
	numrNounSoftSurface = regexp.MustCompile(
		`^(?:плюс|мінус|ранку|вечора|ночі|тепла|морозу|родом|зростом|дивом|станом|вагою|слід|типу|формату|вартістю|році|населення)$`)
	numrLemmaSoftRE = regexp.MustCompile(
		`^(?:у?весь|який(?:сь)?|свій|сам|цей|решта|кількість|вартий|кожний|жодний|менший|більший|вищий|нижчий)$`)
	numrPivtoraFullRE = regexp.MustCompile(`^(?:(?:один-|одне-)?півтора|(?:одна-)?півтори)$`)
	numrArticlePrevRE = regexp.MustCompile(
		`^(?:ч\.|ст\.|п\.|частина|стаття|пункт|підпункт|абзац|№|номер)$`)
	numrFractHalfPrepRE = regexp.MustCompile(
		`^(?:від|до|протягом|[ув]продовж|близько|після|для|більше|менше)$`)
	numrObyeLikeRE = regexp.MustCompile(`^(?:обоє|двоє|троє|.+еро)$`)
	numrObyeAnimRE = regexp.MustCompile(`^(?:обоє|обидвоє|троє)$`)
	numrSyomaRE    = regexp.MustCompile(`^(?:сьома|дев.яноста)$`)
	numrNextSoftRE = regexp.MustCompile(`^(?:[.,:;()«»—–-]|і|й|та)$`)
)

// hasAdjPRodNotNumr is RE2-friendly adj(?!.*numr).*:p:v_rod.*
func hasAdjPRodNotNumr(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "numr") {
			continue
		}
		if strings.HasPrefix(p, "adj") && strings.Contains(p, ":p:v_rod") {
			return true
		}
	}
	return false
}

// isNumberToken reports Java state.number (number POS on numeral).
func isNumberToken(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, p := range CollectPOSTags(tok) {
		if taguk.IPOSNumber.Match(p) {
			return true
		}
	}
	return false
}

// IsNumrNounException ports TokenAgreementNumrNounExceptionHelper.
// Invalid layout → exception (no flag). Incomplete only where Java needs full tagger.
func IsNumrNounException(tokens []*languagetool.AnalyzedTokenReadings, numrPos, nounPos int) bool {
	if numrPos < 0 || nounPos <= numrPos || numrPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	numr, noun := tokens[numrPos], tokens[nounPos]
	if numr == nil || noun == nil {
		return true
	}
	numrLower := CleanTokenLower(numr)
	nounLower := CleanTokenLower(noun)
	numrInfs := GetNumrCaseInflections(CollectPOSTags(numr))

	// для багатьох/обох/двох/… — Java full-string matches
	if numrManyObPattern.MatchString(numrLower) {
		return true
	}
	// плюс|мінус|ранку|…
	if numrNounSoftSurface.MatchString(nounLower) {
		return true
	}
	// lemma set on noun
	if HasLemmaTokenRE(noun, numrLemmaSoftRE) {
		return true
	}

	// хвилин п'ять люди / сотні дві персон — TIME_PLUS before numr, inflections overlap
	if numrPos > 1 &&
		HasLemmaWithPosRE(tokens[numrPos-1], TimePlusLemmaList(), regexp.MustCompile(`noun.*?.:v_(naz|rod).*`)) {
		prevInfs := GetNounCaseInflections(CollectPOSTags(tokens[numrPos-1]))
		if InflectionsIntersect(numrInfs, prevInfs) {
			return true
		}
	}

	// півтора + adj:p + noun:p:v_naz
	if nounPos < len(tokens)-1 &&
		numrPivtoraFullRE.MatchString(CleanTokenLower(numr)) &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:p:v_(naz|rod).*`)) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*?:p:v_naz.*`)) {
		return true
	}

	// хвилин зо п'ять люди — TIME_PLUS + prep + numr
	if numrPos > 2 &&
		HasPosTagStart(tokens[numrPos-1], "prep") &&
		HasLemmaWithPosRE(tokens[numrPos-2], TimePlusLemmaList(), regexp.MustCompile(`noun.*?p:v_(naz|rod).*`)) {
		prevInfs := GetNounCaseInflections(CollectPOSTags(tokens[numrPos-2]))
		if InflectionsIntersect(numrInfs, prevInfs) {
			return true
		}
	}

	// У свої вісімдесят пан Василь
	if numrPos > 2 &&
		HasPosTagStart(tokens[numrPos-2], "prep") &&
		CleanTokenLower(tokens[numrPos-1]) == "свої" &&
		HasPosTagRE(numr, regexp.MustCompile(`numr:p:v_zna.*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:.:v_naz.*`)) {
		return true
	}

	// два провінційного вигляду персонажі
	// Java: noun:inanim:.:v_rod(?!.*pron) / noun(?!.*pron) — RE2 has no lookahead; filter :pron in Go.
	if nounPos <= len(tokens)-3 &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:.:v_rod.*`)) &&
		hasPosWithoutPron(tokens[nounPos+1], regexp.MustCompile(`noun:inanim:.:v_rod`)) &&
		hasPosWithoutPron(tokens[nounPos+2], regexp.MustCompile(`^noun`)) {
		adjG := gendersFromPos(noun, regexp.MustCompile(`adj:([mfnp]):v_rod`))
		nounG := gendersFromPos(tokens[nounPos+1], regexp.MustCompile(`noun:inanim:([mfnp]):v_rod`))
		if adjG != "" && nounG != "" && gendersOverlap(adjG, nounG) {
			realInfs := GetNounCaseInflections(CollectPOSTags(tokens[nounPos+2]))
			if InflectionsIntersect(numrInfs, realInfs) {
				return true
			}
		}
	}

	// ,5 + тон|тис|коп or prep-ish left
	if strings.HasSuffix(numrLower, ",5") {
		if regexp.MustCompile(`^(?:тон|тис|коп)$`).MatchString(nounLower) {
			return true
		}
		if numrPos > 1 && numrFractHalfPrepRE.MatchString(CleanTokenLower(tokens[numrPos-1])) {
			return true
		}
	}

	// обоє горбаті
	if numrObyeLikeRE.MatchString(numrLower) &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:p:v_naz.*`)) &&
		strings.HasSuffix(noun.GetToken(), "і") {
		return true
	}
	// обоє режисери
	if numrObyeAnimRE.MatchString(numrLower) &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:p:v_naz.*`)) {
		return true
	}

	// 22 червня — number + MONTH :m:v_rod (Java hasLemma part-contains, not Pattern)
	if isNumberToken(numr) &&
		HasLemmaWithPartPos(noun, MonthLemmas, ":m:v_rod") {
		return true
	}

	// 3 / 4 понеділка
	if numrPos > 2 && tokens[numrPos-1] != nil && tokens[numrPos-1].GetCleanToken() == "/" {
		return true
	}

	// ч.|ст.|…|№ before numr
	if numrPos > 1 {
		prev := tokens[numrPos-1]
		if prev != nil &&
			(numrArticlePrevRE.MatchString(CleanTokenLower(prev)) || prev.GetCleanToken() == "№" ||
				HasLemmaTokenAny(prev, []string{"частина", "стаття", "пункт", "підпункт", "абзац", "номер"})) {
			return true
		}
	}

	// двадцять перший; дві соті
	if HasPosTagRE(noun, regexp.MustCompile(`adj.*numr.*`)) {
		return true
	}

	// два нових горнятка / 2 хворих
	if numrDva34Pattern.MatchString(numrLower) || isNumberToken(numr) {
		if hasAdjPRodNotNumr(noun) {
			if nounPos == len(tokens)-1 {
				return true
			}
			next := tokens[nounPos+1]
			if next != nil {
				if hasAdjPRodNotNumr(next) ||
					HasPosTagRE(next, regexp.MustCompile(`noun.*:p:v_naz.*`)) ||
					HasPosTagStart(next, "prep") ||
					!HasPosTagRE(next, regexp.MustCompile(`^(?:adj|noun)`)) ||
					numrNextSoftRE.MatchString(next.GetCleanToken()) {
					return true
				}
			}
		}
		if strings.HasSuffix(nounLower, "их") &&
			HasPosTagRE(noun, regexp.MustCompile(`noun.*:p:v_rod.*`)) {
			return true
		}
	}

	// сьома вода
	if numrSyomaRE.MatchString(numrLower) &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun:.*?|adj):[fp]:v_naz.*`)) {
		return true
	}

	return false
}

// gendersOverlap reports whether gender letter sets share a character.
func gendersOverlap(a, b string) bool {
	for i := 0; i < len(a); i++ {
		if strings.ContainsRune(b, rune(a[i])) {
			return true
		}
	}
	return false
}

// IsNounVerbException ports TokenAgreementNounVerbExceptionHelper.isException
// (surface/lemma/conj/geo/impers arms; RE2-safe stand-ins for lookarounds).
func IsNounVerbException(tokens []*languagetool.AnalyzedTokenReadings, nounPos, verbPos int) bool {
	// Invalid subject/verb order → exception (no flag). Missing tokens: only order check.
	if nounPos < 0 || verbPos <= nounPos {
		return true
	}
	if tokens == nil || nounPos >= len(tokens) || verbPos >= len(tokens) {
		return false
	}
	noun, verb := tokens[nounPos], tokens[verbPos]
	if noun == nil || verb == nil {
		return false
	}

	// Любителі фотографувати їжу — inf verb after noun governing v_inf
	if HasPosTagRE(verb, verbInfPattern) {
		if LoadCaseGovernmentHelper().HasCaseGovernment(lemmaOf(noun), "v_inf") ||
			hasCaseGovFromReadings(noun, "v_inf") {
			return true
		}
		if tokenLineBefore(tokens, nounPos, "не", "сила") ||
			tokenLineBefore(tokens, nounPos, "не", "проти") {
			return true
		}
		if nl := CleanTokenLower(noun); nl == "хтось" || nl == "дехто" {
			return true
		}
		if verbPos > 0 && CleanTokenLower(tokens[verbPos-1]) == "намагаючись" {
			return true
		}
	}

	// шкода було / годі буде
	if HasPosTagPart(noun, "predic") {
		vl := CleanTokenLower(verb)
		if vl == "було" || vl == "буде" {
			return true
		}
	}
	if CleanTokenLower(noun) == "правда" {
		return true
	}
	if tokenLineBefore(tokens, nounPos, "під", "три", "чорти") ||
		tokenLineBefore(tokens, nounPos, "не", "штука") ||
		tokenLineBefore(tokens, nounPos, "бісики") {
		return true
	}
	// будь якого after verb
	if tokenLineAfter(tokens, verbPos, "будь", "якого") {
		return true
	}
	// не сказати б after verb-1
	if verbPos > 0 && tokenLineAfter(tokens, verbPos-1, "не", "сказати", "б") {
		return true
	}
	if verbPos > 0 && tokenLineBefore(tokens, verbPos-1, "не", "проти") {
		return true
	}
	// воно/решта + :impers
	if HasLemmaTokenAny(noun, []string{"воно", "решта"}) && HasPosTagPart(verb, ":impers") {
		return true
	}
	if verbPos > 0 && HasLemmaToken(tokens[verbPos-1], "Газа") {
		return true
	}
	// чотири дні був
	if nounPos > 1 &&
		hasPosWithoutPron(noun, regexp.MustCompile(`noun:.*:p:v_naz`)) &&
		// Java Pattern.compile("numr:p:v_zna") Matcher.matches() — full tag exactly
		HasLemmaWithPosRE(tokens[nounPos-1], []string{"два", "три", "чотири"}, regexp.MustCompile(`^numr:p:v_zna$`)) {
		return true
	}

	// кандидат в президенти поїхав
	vPrezPrep := []string{"в", "у", "між", "межи", "поміж", "на"}
	if nounPos > 1 && HasPosTagStart(noun, "noun:anim:p:v_naz") &&
		HasLemmaTokenAny(tokens[nounPos-1], vPrezPrep) {
		return true
	}
	// кандидат в народні депутати
	if nounPos > 2 && HasPosTagStart(noun, "noun:anim:p:v_naz") &&
		HasPosTagStart(tokens[nounPos-1], "adj:p:v_zna:rinanim") &&
		HasLemmaTokenAny(tokens[nounPos-2], vPrezPrep) {
		return true
	}
	// both capitalized (unknown surname as verb)
	if IsCapitalized(verb.GetToken()) && IsCapitalized(noun.GetToken()) {
		return true
	}
	// на прізвисько Михайло
	if nounPos > 1 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:anim:.:v_naz:prop:[fl]name.*`)) {
		pl := CleanTokenLower(tokens[nounPos-1])
		if pl == "ім'я" || pl == "прізвище" || pl == "прізвисько" {
			return true
		}
	}
	// матч Туреччина — Україна
	if nounPos > 2 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*:v_naz.*prop.*`)) &&
		tokens[nounPos-1] != nil &&
		regexp.MustCompile(`^[-\x{2013}\x{2014}]$`).MatchString(CleanTokenLower(tokens[nounPos-1])) &&
		HasPosTagRE(tokens[nounPos-2], regexp.MustCompile(`noun.*:v_naz.*prop.*`)) {
		return true
	}
	// Тарас ЗАКУСИЛО (all-upper verb)
	if isAllUpper(verb.GetToken()) {
		return true
	}
	// Збережені Я позбудуться
	if nounPos > 1 && noun.GetToken() == "Я" {
		return true
	}
	// а він давай пити
	if verbPos > 2 && verbPos < len(tokens)-1 && verb.GetToken() == "давай" {
		return true
	}
	// Ви може образились (може not before inf)
	if verbPos > 1 && verbPos < len(tokens)-1 && verb.GetToken() == "може" &&
		tokens[verbPos-1].GetToken() != "не" &&
		!HasPosTagRE(tokens[verbPos+1], verbInfPattern) {
		return true
	}

	// Прем'єр-міністр повторила — masc profession + fem verb (Java hasMascFemLemma)
	if HasPosTagPart(noun, "noun:anim:m:v_naz") &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*:f(:.*|$)`)) &&
		HasMascFemLemma(noun) {
		return true
	}
	// пора було
	if CleanTokenLower(noun) == "пора" && CleanTokenLower(verb) == "було" {
		return true
	}
	// решта/частина/… + plural/neuter verb
	pseudoPlural := map[string]bool{
		"решта": true, "частина": true, "частка": true, "половина": true, "третина": true, "чверть": true,
	}
	if pseudoPlural[CleanTokenLower(noun)] && HasPosTagRE(verb, regexp.MustCompile(`.*:[pn](:.*|$)`)) {
		return true
	}
	// з Василем … разом брали
	if nounPos+1 < len(tokens) && strings.EqualFold(CleanTokenLower(tokens[nounPos+1]), "разом") &&
		HasPosTagRE(verb, regexp.MustCompile(`.*:p(:.*|$)`)) {
		return true
	}
	// більше ніж будь-хто маємо
	if nounPos > 2 && HasLemmaToken(tokens[nounPos-1], "ніж") {
		return true
	}
	// моя ти зоре — ти + v_kly on "verb" slot (mis-tagged)
	if nounPos > 1 && strings.EqualFold(noun.GetToken(), "ти") &&
		HasPosTagRE(verb, regexp.MustCompile(`noun.*?v_kly.*`)) {
		return true
	}
	// вона візьми та й скажи
	if verbPos < len(tokens)-2 && verb.GetToken() == "візьми" &&
		HasLemmaTokenAny(tokens[verbPos+1], []string{"і", "й", "та"}) {
		return true
	}
	// GEO_QUALIFIERS + proper (в державі Україна / місті Біла Церква)
	if nounPos > 1 && IsPossiblyProperNoun(noun) && HasLemmaTokenAny(tokens[nounPos-1], geoQualifiers) {
		return true
	}
	if nounPos > 2 && IsPossiblyProperNoun(noun) && IsPossiblyProperNoun(tokens[nounPos-1]) &&
		HasLemmaTokenAny(tokens[nounPos-2], geoQualifiers) {
		return true
	}
	// У невизнаній республіці Південна Осетія
	if nounPos > 3 &&
		HasPosTagPart(noun, "v_naz:prop") &&
		HasPosTagRE(tokens[nounPos-1], regexp.MustCompile(`adj:.:v_naz.*`)) &&
		HasPosTagRE(tokens[nounPos-2], regexp.MustCompile(`noun.*:v_(rod|zna|mis).*`)) {
		return true
	}
	// ми в державі Україна маємо — prep+noun:inanim before prop
	if verbPos > 3 && HasPosTagRE(tokens[verbPos-1], regexp.MustCompile(`noun:inanim:.:v_naz:prop.*`)) {
		vPos := verbPos
		if IsCapitalized(tokens[nounPos-1].GetToken()) && HasPosTagStart(tokens[nounPos-1], "adj") {
			vPos--
		}
		if vPos > 3 && HasPosTagStart(tokens[vPos-2], "noun:inanim") && HasPosTagPart(tokens[vPos-3], "prep") {
			cases := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(tokens[vPos-3], "prep")
			if len(cases) > 0 {
				var list []string
				for c := range cases {
					list = append(list, c)
				}
				if HasVidmPosTag(list, tokens[vPos-2]) {
					return true
				}
			}
		}
	}
	// чи готові ми сидіти — adj governing v_inf + agreement with noun
	if nounPos > 1 && HasPosTagPart(tokens[nounPos-1], "adj") && HasPosTagRE(verb, verbInfPattern) {
		if hasCaseGovFromReadings(tokens[nounPos-1], "v_inf") &&
			adjNounInflectionOverlap(tokens[nounPos-1], noun) {
			return true
		}
	}
	// тому що, як австрієць маєте — reverse tokenSearch for Як/як
	if HasPosTagRE(noun, regexp.MustCompile(`noun.*:v_naz.*`)) {
		if TokenSearch(tokens, nounPos-1, "", regexp.MustCompile(`^[Яя]к$`),
			regexp.MustCompile(`adj:.:v_naz.*`), DirReverse) != -1 {
			return true
		}
	}

	// можуть російськомовні громадяни вважатися — INF_ARGREEMENT before/after inf
	// Java: reverseSearchIdx / forwardLemmaSearchIdx with INF_ARGREEMENT_PATTERN
	if HasPosTagRE(verb, verbInfPattern) {
		if nounPos > 1 {
			foundIdx := ReverseSearchIdx(tokens, nounPos-1, 6, infAgreementPattern, nil)
			if foundIdx >= 0 {
				// Java: non-adj → exception; adj → !disjoint(adjInflections, nounInflections)
				if !HasPosTagStart(tokens[foundIdx], "adj") {
					return true
				}
				if adjNounInflectionOverlap(tokens[foundIdx], noun) {
					return true
				}
			}
		}
		if verbPos < len(tokens)-1 {
			foundIdx := ForwardLemmaSearchIdx(tokens, verbPos+1, 7, infAgreementPattern, nil)
			if foundIdx >= 0 {
				if !HasPosTagStart(tokens[foundIdx], "adj") {
					return true
				}
				if adjNounInflectionOverlap(tokens[foundIdx], noun) {
					return true
				}
			}
		}
		// як навчила мене бабуся місити тісто — prev finite verb agrees with noun
		if nounPos > 1 {
			prevVerbIdx := ReverseSearchIdx(tokens, nounPos-1, 7, nil, regexp.MustCompile(`verb.*`))
			if prevVerbIdx >= 0 && prevVerbIdx != verbPos {
				if VerbInflectionsOverlap(CollectPOSTags(tokens[prevVerbIdx]), CollectPOSTags(noun)) {
					return true
				}
			}
		}
		// громадяни проголосувати зможуть — next finite verb agrees
		if verbPos < len(tokens)-1 {
			nextVerbPos := NewSearchMatch("").
				IgnoreInsertsOn().
				WithLimit(8).
				Target(ConditionPostag(regexp.MustCompile(`verb.*`))).
				MAfterATR(tokens, verbPos+1)
			if nextVerbPos >= 0 &&
				VerbInflectionsOverlap(CollectPOSTags(tokens[nextVerbPos]), CollectPOSTags(noun)) {
				return true
			}
		}
	}

	// — це були невільники
	if nounPos > 1 && verbPos < len(tokens)-1 &&
		noun.GetToken() == "це" &&
		DashesPattern.MatchString(tokens[nounPos-1].GetToken()) {
		return true
	}
	// це не передбачено
	if noun.GetToken() == "це" && HasPosTagPart(verb, "impers") {
		return true
	}

	// 22 льотчики загинуло / два сини народилося
	if nounPos > 1 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*:p:v_naz.*`)) &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*?past:n.*`)) {
		prev := tokens[nounPos-1].GetCleanToken()
		if regexp.MustCompile(`^\d+[234]$`).MatchString(prev) ||
			containsStr([]string{"два", "три", "чотири"}, prev) {
			return true
		}
	}

	// зіркова пара … вирішили
	if HasPosTagRE(verb, regexp.MustCompile(`verb.*:[fp](?:$|:.*)`)) {
		if NewSearchMatch("").
			Target(ConditionToken("пара")).
			Skip(ConditionTokenRE(conjForPluralTokenRE).WithNegate()).
			WithLimit(10).
			MBeforeATR(tokens, nounPos-1) > 0 {
			return true
		}
	}

	// plural verb + coordination before subject (Java verb.*:p block)
	if HasPosTagRE(verb, regexp.MustCompile(`verb.*:p(?:$|:.*)`)) {
		// Колесніков/Ахметов посилили
		if nounPos > 2 &&
			(tokens[nounPos-1].GetToken() == "/" || tokens[nounPos-2].GetToken() == "/") {
			return true
		}
		// кефаль, барабуля, хамса
		if nounPos > 2 &&
			isConjForPluralWithComma(tokens[nounPos-1]) &&
			HasPosTagRE(tokens[nounPos-2], nounVNazPattern) {
			return true
		}
		// його побут, життєва поведінка
		if nounPos > 3 &&
			isConjForPluralWithComma(tokens[nounPos-2]) &&
			HasPosTagRE(tokens[nounPos-3], nounVNazPattern) &&
			adjNounInflectionOverlap(tokens[nounPos-1], noun) {
			return true
		}
		// моя мама й сестра мешкали — conj before subject
		m := &SearchMatch{IgnoreQuotes: true, Limit: 7}
		m.Target(ConditionTokenRE(conjForPluralTokenRE))
		m.Skip(ConditionPostag(regexp.MustCompile(`(?:noun.*?v_naz|(?:adj|numr):.:v_naz|adv|part).*`)))
		pos0left := m.MBeforeATR(tokens, nounPos-1)
		if pos0left > 0 && !IsNonPluralA(tokens, pos0left) {
			pos0right := pos0left
			if pos0left > 1 && tokens[pos0left-1] != nil && tokens[pos0left-1].GetToken() == "," {
				pos0left--
			}
			if pos0left > 1 {
				if pos0right > 2 {
					// і та й інша
					if pos0left < len(tokens)-1 &&
						HasLemmaToken(tokens[pos0right+1], "інший") &&
						HasLemmaToken(tokens[pos0left-1], "той") {
						return true
					}
					// як Німеччина, так і Україна — conj before left
					if HasPosTagPart(tokens[pos0left-1], "conj") {
						pos0left--
					}
					// він особисто й …
					if containsStr([]string{"особисто", "зокрема", "загалом"}, CleanTokenLower(tokens[pos0left-1])) {
						pos0left--
					}
					if containsStr([]string{"особисто", "зокрема", "загалом"}, CleanTokenLower(tokens[verbPos-1])) {
						return true
					}
					// ) before conj
					if tokens[pos0left-1].GetToken() == ")" {
						return true
					}
					// і уряд … і президент
					if HasPosTagRE(tokens[pos0left-1], nounVNazPattern) {
						return true
					}
					// І спочатку Білорусь, а тепер і Україна — adv + conj
					if verbPos > 6 && pos0left > 2 &&
						HasPosTagPart(tokens[pos0left-1], "adv") &&
						HasPosTagPart(tokens[pos0left-2], "conj") {
						pos0left -= 2
					}
					// strip trailing quotes/commas left of conj
					for pos0left > 2 && tokens[pos0left-1] != nil &&
						regexp.MustCompile(`^[,»“”"]$`).MatchString(tokens[pos0left-1].GetToken()) {
						pos0left--
					}
				}
				// моя мама й сестра / proper / number:latin
				if pos0left > 1 {
					left := tokens[pos0left-1]
					if HasPosTagStart(left, "noun") ||
						HasPosTagStart(left, "number:latin") ||
						IsPossiblyProperNoun(left) {
						return true
					}
					// біологічна і ядерна зброя стають
					if HasPosTagRE(left, adjVNazPattern) {
						return true
					}
				}
			}
		}

		// Усі розписи, а також архітектура відрізняються
		if pos3 := TokenSearch(tokens, verbPos-2, "", regexp.MustCompile(`(?i)^також$`),
			regexp.MustCompile(`(?:noun|adj:.:v_naz|adv|part).*`), DirReverse); pos3 > 1 {
			return true
		}

		// що пачка, що ковбаса коштують
		if nounPos > 5 {
			prev := strings.ToLower(tokens[nounPos-1].GetToken())
			if prev == "що" || prev == "не" {
				if TokenSearch(tokens, nounPos-3, "", regexp.MustCompile("(?i)^"+regexp.QuoteMeta(prev)+"$"),
					regexp.MustCompile(`(?:noun|adj).*`), DirReverse) > nounPos-7 {
					return true
				}
			}
		}

		// Бразилія, Мексика, Індія збувають
		if pos1 := TokenSearch(tokens, nounPos-1, "", regexp.MustCompile(`^,$`),
			regexp.MustCompile(`^adj.*`), DirReverse); pos1 > 1 {
			if HasPosTagRE(tokens[pos1-1], nounVNazPattern) ||
				(pos1 > 2 &&
					HasPosTagRE(tokens[pos1-1], regexp.MustCompile(`noun.*:v_rod.*`)) &&
					HasPosTagRE(tokens[pos1-2], nounVNazPattern)) {
				return true
			}
		}

		// Мустафа Джемілєв, Рефат Чубаров
		if nounPos > 4 &&
			IsCapitalized(noun.GetToken()) &&
			(HasPosTagStart(tokens[nounPos-1], "noun:anim") || IsInitial(tokens[nounPos-1])) &&
			isConjForPluralWithComma(tokens[nounPos-2]) &&
			IsCapitalized(tokens[nounPos-3].GetToken()) &&
			(HasPosTagStart(tokens[nounPos-4], "noun:anim") || IsInitial(tokens[nounPos-1])) {
			return true
		}

		// закордонний депутат і прем'єр … / а також голова
		idxM := &SearchMatch{IgnoreQuotes: true}
		idxM.Target(ConditionTokenRE(conjForPluralTokenRE))
		idxM.IgnoreInsertsOn()
		idxM.Skip(
			ConditionPostag(regexp.MustCompile(`(?:noun|adj).*?v_(?:naz|rod).*`)),
			ConditionTokenRE(regexp.MustCompile(`^(?:і?з|зі|від|на|навіть|також|потім|згодом)$`)),
		)
		if idx := idxM.MBeforeATR(tokens, nounPos-1); idx > 0 {
			if IsNonPluralA(tokens, idx) {
				idx = -1
			}
			if idx > 1 &&
				(HasPosTagRE(tokens[idx-1], nounVNazPattern) ||
					IsCapitalized(tokens[idx-1].GetCleanToken()) ||
					HasLemmaTokenAny(tokens[idx+1], []string{"навіть", "також", "потім", "згодом"}) ||
					HasLemmaTokenAny(tokens[idx-1], []string{"потім", "згодом"})) {
				return true
			}
		}

		// понад сотня / тисяча
		if (HasPosTagPart(noun, "numr") && !HasLemmaToken(noun, "один")) ||
			HasLemmaTokenAny(noun, []string{"сотня", "тисяча", "десяток"}) {
			return true
		}
		// 121 депутат (not ending in 1 unless 11)
		if nounPos > 1 && HasPosTagPart(tokens[nounPos-1], "number") {
			nt := tokens[nounPos-1].GetToken()
			if !strings.HasSuffix(nt, "1") || strings.HasSuffix(nt, "11") {
				return true
			}
		}
		// 100 чоловік without жінка in sentence
		if nounPos > 0 && HasPosTagPart(tokens[nounPos-1], "num") &&
			noun.GetToken() == "чоловік" &&
			TokenSearch(tokens, 1, "noun:anim:f:", regexp.MustCompile(`жінк[аи]`), regexp.MustCompile(`.*`), DirForward) == -1 {
			return true
		}
		// 50%+1 / плюс
		if nounPos > 1 &&
			(strings.HasSuffix(tokens[nounPos-1].GetToken(), "+1") ||
				TokenSearch(tokens, verbPos-2, "", regexp.MustCompile(`(?i)^плюс$`),
					regexp.MustCompile(`(?:numr|adj).*.:v_naz.*`), DirReverse) > 0) {
			return true
		}
		// Решта 121 депутат
		if nounPos > 2 &&
			HasLemmaToken(tokens[nounPos-2], "решта") &&
			tokens[nounPos-1].GetToken() != "" &&
			regexp.MustCompile(`.+1$`).MatchString(tokens[nounPos-1].GetToken()) {
			return true
		}
		// дві групи, кожна виконували
		if nounPos > 2 &&
			HasLemmaToken(noun, "кожний") &&
			HasPosTagRE(verb, regexp.MustCompile(`verb.*(?:past:p|:p:3).*`)) {
			return true
		}
		// навіть / ані / жоден before subject
		if nounPos > 2 &&
			regexp.MustCompile(`^(?:а?ні|жодн.*|навіть)$`).MatchString(tokens[nounPos-1].GetToken()) {
			return true
		}
		// ані … не + plural verb
		if nounPos > 2 && verbPos > 0 &&
			tokens[verbPos-1].GetToken() == "не" &&
			ReverseSearch(tokens, nounPos-1, 5, regexp.MustCompile(`^а?ні$`), nil) {
			return true
		}
		if nounPos > 3 && verbPos > 0 &&
			tokens[verbPos-1].GetToken() == "не" &&
			regexp.MustCompile(`^а?ні$`).MatchString(tokens[nounPos-2].GetToken()) &&
			adjNounInflectionOverlap(tokens[nounPos-1], noun) {
			return true
		}
	}

	// Сейм Республіки Польща проігнорував — prop + v_rod + noun agrees with verb
	if nounPos > 3 &&
		HasPosTagPart(noun, ":prop") &&
		HasPosTagRE(tokens[nounPos-1], regexp.MustCompile(`noun.*:v_rod.*`)) &&
		VerbInflectionsOverlap(CollectPOSTags(verb), CollectPOSTags(tokens[nounPos-2])) {
		return true
	}

	// комітет … села Оляниця — geo prop after non-naz inanim
	if nounPos > 1 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:inanim:[mnf]:v_naz:prop:geo.*`)) &&
		hasPosWithoutPron(tokens[nounPos-1], regexp.MustCompile(`noun:inanim:[mnf]:v_`)) &&
		!HasPosTagPart(tokens[nounPos-1], "v_naz") {
		return true
	}

	// У штатах Техас … запроваджено — prop + impers
	if nounPos > 1 &&
		HasPosTagPart(noun, ":prop") &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*:impers.*`)) {
		return true
	}

	// на австралійський штат Вікторія налетів — prep+adj+noun + prop, prep gov both
	if nounPos > 3 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:inanim:.:v_naz:prop.*`)) &&
		HasPosTagRE(tokens[nounPos-1], regexp.MustCompile(`noun:inanim:.*`)) &&
		HasPosTagRE(tokens[nounPos-2], regexp.MustCompile(`adj:.*`)) &&
		HasPosTagPart(tokens[nounPos-3], "prep") {
		govs := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(tokens[nounPos-3], "prep")
		if len(govs) > 0 {
			list := make([]string, 0, len(govs))
			for c := range govs {
				list = append(list, c)
			}
			if HasVidmPosTag(list, tokens[nounPos-1]) && HasVidmPosTag(list, tokens[nounPos-2]) {
				return true
			}
		}
	}

	// Угорщина було пішла — було + later finite verb agreeing with subject
	if verbPos < len(tokens)-1 && CleanTokenLower(verb) == "було" {
		pos := TokenSearch(tokens, verbPos+1, "verb:", nil, regexp.MustCompile(`^adv.*`), DirForward)
		if pos >= 0 &&
			VerbInflectionsOverlap(CollectPOSTags(tokens[pos]), CollectPOSTags(noun)) {
			return true
		}
	}

	// клан Рана було знищено — prop + було + impers
	if verbPos < len(tokens)-1 &&
		HasPosTagPart(noun, ":prop") &&
		CleanTokenLower(verb) == "було" &&
		HasPosTagRE(tokens[verbPos+1], regexp.MustCompile(`verb.*:impers.*`)) {
		return true
	}

	// діагноз дизентерія підтвердився — prev inanim v_naz (not pron) agrees with verb
	if nounPos > 1 &&
		HasPosTagRE(tokens[nounPos-1], regexp.MustCompile(`noun:inanim:.:v_naz.*`)) &&
		!HasPosTagPart(tokens[nounPos-1], ":pron") &&
		!HasPosTagRE(noun, regexp.MustCompile(`noun.*pron.*`)) &&
		VerbInflectionsOverlap(CollectPOSTags(verb), CollectPOSTags(tokens[nounPos-1])) {
		return true
	}

	return false
}

// conjForPluralTokenRE matches CONJ_FOR_PLURAL surfaces (full-string).
var conjForPluralTokenRE = regexp.MustCompile(`^(?i:і|а|й|та|чи|або|ані|також|то|a|i)$`)

// Java TokenAgreementNounVerbExceptionHelper.INF_ARGREEMENT_PATTERN
var infAgreementPattern = regexp.MustCompile(
	`^(не)?(здатний|змушений|з?г[іо]дний|зобов'язаний|повинний|готовий|достойний|покликаний|спроможний|радий|налаштований|зацікавлений|повинно|змога|стан|можна)$`,
)

// Java GEO_QUALIFIERS
var geoQualifiers = []string{
	"село", "селище", "місто", "містечко", "хутір", "республіка", "держава", "гора", "планета",
	"мікрорайон", "райцентр", "заповідник", "мис", "м.", "с.", "п.", "штат", "округ", "графство",
	"вірус", "ураган",
}

// adjNounInflectionOverlap ports !Collections.disjoint(getAdjInflections, getNounInflections).
func adjNounInflectionOverlap(adj, noun *languagetool.AnalyzedTokenReadings) bool {
	if adj == nil || noun == nil {
		return false
	}
	aInf := GetAdjCaseInflections(CollectPOSTags(adj))
	nInf := GetNounCaseInflections(CollectPOSTags(noun))
	if len(aInf) == 0 || len(nInf) == 0 {
		return false
	}
	return InflectionsIntersect(aInf, nInf)
}

func isAllUpper(s string) bool {
	if s == "" {
		return false
	}
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// IsVerbNounException ports TokenAgreementVerbNounExceptionHelper.isException
// plus hard-adj / skip / verb-side soft paths usable from the pair matcher.
func IsVerbNounException(tokens []*languagetool.AnalyzedTokenReadings, verbPos, nounPos int) bool {
	if verbPos < 0 || nounPos <= verbPos {
		return true
	}
	if tokens == nil || verbPos >= len(tokens) || nounPos >= len(tokens) {
		return false
	}
	verb, noun := tokens[verbPos], tokens[nounPos]
	if verb == nil || noun == nil {
		return false
	}

	// verb-side: мусити / може / як є / будь то / pluperfect / спати
	if IsExceptionVerb(tokens, verbPos) || IsExceptionVerbSkip(tokens, verbPos) {
		return true
	}
	// hard adj/noun and insert skips at object position
	if IsVerbNounHardAdjNoun(tokens, nounPos, verbPos) >= 0 {
		return true
	}
	if IsVerbNounExceptionSkip(tokens, nounPos) >= 0 {
		return true
	}

	// numr v_naz / quant + s/n verb (боротиметься кілька / входило двоє)
	quantish := HasPosTagRE(noun, regexp.MustCompile(`numr.*v_naz.*`)) ||
		HasLemmaTokenRE(noun, AdvQuantPattern) ||
		(AdvQuantPattern.MatchString(CleanTokenLower(noun)) &&
			HasPosTagRE(noun, regexp.MustCompile(`noun.*v_naz.*|adv.*|part.*`)))
	if quantish {
		if HasPosTagRE(verb, regexp.MustCompile(`.*:[sn](:.*|$)`)) {
			return true
		}
		if verbPos > 1 && HasPosTagRE(verb, regexp.MustCompile(`verb.*inf.*`)) &&
			HasLemmaWithPosRE(tokens[verbPos-1], []string{"бути", "мусити"}, regexp.MustCompile(`verb.*(past:n|:s:3).*`)) {
			return true
		}
	}

	// здатна була
	if verbPos > 1 && HasLemmaToken(verb, "бути") {
		modals := []string{"змушений", "вимушений", "повинний", "здатний", "готовий", "ладний", "радий"}
		if HasLemmaWithPosRE(tokens[verbPos-1], modals, regexp.MustCompile(`adj:.:v_naz.*`)) {
			return true
		}
	}
	// зможе + v_oru / чим могла
	if HasLemmaTokenRE(verb, regexp.MustCompile(`^з?могти$`)) {
		if HasPosTagPart(noun, "v_oru") {
			return true
		}
		if verbPos > 1 && CleanTokenLower(tokens[verbPos-1]) == "чим" {
			return true
		}
	}
	// стало відомо
	if verbPos < len(tokens)-1 && strings.EqualFold(CleanTokenLower(verb), "стало") {
		next := CleanTokenLower(tokens[verbPos+1])
		if next == "відомо" || next == "видно" || next == "зрозуміло" {
			return true
		}
	}
	// я буду каву
	if verbPos > 1 && CleanTokenLower(tokens[verbPos-1]) == "я" && CleanTokenLower(verb) == "буду" {
		if HasPosTagRE(noun, regexp.MustCompile(`noun:inanim:.:v_zna.*`)) ||
			hasPosWithoutRanim(noun, regexp.MustCompile(`adj:.:v_zna`)) {
			return true
		}
	}
	// хоче маляром
	if HasLemmaToken(verb, "хотіти") && HasPosTagPart(noun, "v_oru") {
		return true
	}
	// були б іншої думки
	if HasLemmaTokenRE(verb, regexp.MustCompile(`^бути$`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`(adj|numr).*v_rod.*`)) {
		return true
	}
	// що є сил
	if verbPos > 1 && CleanTokenLower(tokens[verbPos-1]) == "що" &&
		HasLemmaWithPosRE(verb, []string{"бути"}, regexp.MustCompile(`verb.*(:s:3|past:n).*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`(adj|noun).*v_rod.*`)) {
		return true
	}
	// навіщо було …
	if verbPos > 1 && CleanTokenLower(verb) == "було" && CleanTokenLower(tokens[verbPos-1]) == "навіщо" {
		return true
	}
	// чесніше було б / predic + було
	if CleanTokenLower(verb) == "було" && verbPos > 1 {
		if HasPosTagRE(tokens[verbPos-1], regexp.MustCompile(`(adv:comp[cs].*|.*predic.*)`)) {
			return true
		}
		if verbPos > 2 && regexp.MustCompile(`^би?$`).MatchString(CleanTokenLower(tokens[verbPos-1])) &&
			HasPosTagRE(tokens[verbPos-2], regexp.MustCompile(`(adv:comp[cs].*|.*predic.*)`)) {
			return true
		}
		// квітне притухлий було пафос
		if HasPosTagRE(noun, regexp.MustCompile(`.*v_naz.*`)) &&
			HasPosTagRE(tokens[verbPos-1], regexp.MustCompile(`adj:.:v_naz.*:adjp:.*:perf.*`)) {
			return true
		}
	}
	// підстрахуватися не зайве
	if nounLower := CleanTokenLower(noun); nounLower == "зайве" || nounLower == "резон" {
		return true
	}
	// далі + v_rod
	if nounPos > 0 && CleanTokenLower(tokens[nounPos-1]) == "далі" && HasPosTagPart(noun, "v_rod") {
		return true
	}
	// було всі 90-ті
	if regexp.MustCompile(`^(було|буде)$`).MatchString(CleanTokenLower(verb)) &&
		HasLemmaWithPosRE(noun, []string{"весь"}, regexp.MustCompile(`.*v_zna.*`)) {
		return true
	}
	// він був талановита людина
	if CleanTokenLower(verb) == "був" {
		nl := CleanTokenLower(noun)
		if nl == "людина" || nl == "знаменитість" {
			return true
		}
		if nounPos < len(tokens)-1 && CleanTokenLower(tokens[nounPos+1]) == "людина" {
			return true
		}
	}
	// мати + v_oru (має своїм наслідком) — Java SearchHelper arm is commented; active path is lemma+v_oru only
	if HasLemmaTokenRE(verb, regexp.MustCompile(`^(мати|маючи|мавши)$`)) && HasPosTagPart(noun, "v_oru") {
		return true
	}

	// --- SearchHelper.Match Condition arms (Java TokenAgreementVerbNounExceptionHelper) ---

	vl := CleanTokenLower(verb)
	// буде видно тільки супутники — predic between verb and noun
	if (vl == "було" || vl == "буде") && verbPos+1 < nounPos {
		lim := nounPos - verbPos
		if lim < 1 {
			lim = 1
		}
		if (&SearchMatch{IgnoreQuotes: true}).
			Target(ConditionPostag(regexp.MustCompile(`.*predic.*`))).
			WithLimit(lim).
			MAfterATR(tokens, verbPos+1) >= 0 {
			return true
		}
	}
	// потрібно буде … — lemma треба|потрібно immediately before було/буде
	if (vl == "було" || vl == "буде") && verbPos > 0 {
		if (&SearchMatch{IgnoreQuotes: true}).
			Target(ConditionLemma(regexp.MustCompile(`^(треба|потрібно)$`))).
			MNowATR(tokens, verbPos-1) >= 0 {
			return true
		}
	}
	// Конкурс був … num
	if verbPos > 1 && CleanTokenLower(tokens[verbPos-1]) == "конкурс" &&
		HasLemmaWithPosRE(verb, []string{"бути"}, regexp.MustCompile(`verb.*(:s:3|past:m).*`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`num.*`)) {
		return true
	}
	// розподілятиметься пропорційно вкладеній праці — adv case gov between verb and noun
	if nounPos-verbPos > 1 {
		mid := tokens[nounPos-1]
		// Java Pattern.compile("adv(?!p).*") — RE2: adv but not advp
		cases := LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadings(mid, "adv")
		// drop if mid is advp-only without bare adv prefix case map
		if len(cases) > 0 && !HasPosTagStart(mid, "advp") {
			var list []string
			for c := range cases {
				list = append(list, c)
			}
			for ii := verbPos + 1; ii < nounPos; ii++ {
				if HasVidmPosTag(list, tokens[ii]) {
					return true
				}
			}
		}
	}
	// TIME_PLUS after noun span (відбувається кожні два роки)
	if TimePlusLemmasPattern != nil {
		if (&SearchMatch{IgnoreQuotes: true}).
			Skip(ConditionPostag(regexp.MustCompile(`.*v_(rod|zna|oru).*|part.*|number`))).
			Target(ConditionLemma(TimePlusLemmasPattern)).
			WithLimit(4).
			MAfterATR(tokens, nounPos) > 0 {
			return true
		}
		// йде три з половиною години
		if nounPos < len(tokens)-3 && HasPosTagRE(noun, regexp.MustCompile(`numr.*v_zna.*`)) {
			if (&SearchMatch{IgnoreQuotes: true}).
				Target(ConditionLemma(TimePlusLemmasPattern)).
				WithLimit(4).
				MAfterATR(tokens, nounPos+1) > 0 {
				return true
			}
		}
	}
	// мова instrumental after noun
	if (&SearchMatch{IgnoreQuotes: true}).
		Skip(ConditionPostag(regexp.MustCompile(`.*v_oru.*|part.*|adv.*`))).
		Target(SearchCondition{
			Lemma:  regexp.MustCompile(`^мова$`),
			Postag: regexp.MustCompile(`noun:inanim:.:v_oru.*`),
		}).
		WithLimit(4).
		MAfterATR(tokens, nounPos) > 0 {
		return true
	}
	// став жовтого кольору
	if nounPos < len(tokens)-1 &&
		CleanTokenLower(tokens[nounPos+1]) == "кольору" &&
		HasPosTagStart(noun, "adj:m:v_rod") {
		return true
	}
	// fixed phrases (Java tokenLine mNow)
	for _, phrase := range []string{
		"не те щоб", "не те що", "не останньою чергою",
		"світ за очі", "ні світ ні", "станом на", "страх як", "жах як",
	} {
		// mNow at nounPos: exact phrase starting at noun
		if NewSearchMatch(phrase).MNowATR(tokens, nounPos) >= 0 {
			return true
		}
		if NewSearchMatch(phrase).MNowATR(tokens, verbPos) >= 0 {
			return true
		}
	}
	// куди очі
	if NewSearchMatch("куди очі").MNowATR(tokens, nounPos) >= 0 ||
		NewSearchMatch("куди очі").MNowATR(tokens, verbPos) >= 0 {
		return true
	}
	// не те, що — comma as separate token
	if NewSearchMatch("не те , що").MNowATR(tokens, nounPos) >= 0 ||
		NewSearchMatch("не те , що").MNowATR(tokens, verbPos) >= 0 {
		return true
	}

	nounLower := CleanTokenLower(noun)

	// плюс|мінус|…
	if IsPlusMinusLemma(nounLower) {
		return true
	}
	// закінчилося 18-го / дорогою / толком / …
	if regexp.MustCompile(`^(?:[0-9]+-.+|дорогою|толком|дивом|чверть|третину|половину|святая)$`).MatchString(nounLower) {
		return true
	}
	// сміялася всю дорогу / цілою дорогою
	if nounPos < len(tokens)-1 &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:[fn]:v_(zna|oru).*`)) &&
		HasLemmaWithPosRE(tokens[nounPos+1], []string{"дорога", "життя", "міра"}, regexp.MustCompile(`noun:inanim:[fn]:v_(zna|oru).*`)) {
		return true
	}
	// запропоновано відділом
	if HasPosTagPart(verb, "impers") && HasPosTagRE(noun, regexp.MustCompile(`.*v_oru.*`)) {
		return true
	}
	// займаючись кожен
	if HasLemmaWithPosRE(noun, []string{"кожний"}, regexp.MustCompile(`.*v_naz.*`)) {
		return true
	}
	// звалося Подєбради
	if HasLemmaTokenAny(verb, []string{"звати", "називати", "зватися", "називатися"}) &&
		isUpperFirst(noun.GetCleanToken()) {
		return true
	}
	// тривав довгих десять раундів
	if HasLemmaTokenAny(verb, []string{"тривати", "протривати", "йти", "іти", "ходити", "їхати"}) &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:adj|numr|noun:inanim).*v_zna.*`)) {
		return true
	}
	// ні сіло ні впало
	if verbPos > 3 &&
		strings.EqualFold(verb.GetCleanToken(), "впало") &&
		tokens[verbPos-1] != nil && tokens[verbPos-1].GetCleanToken() == "ні" {
		return true
	}
	// якщо не сказати слабка
	if verbPos > 2 &&
		strings.EqualFold(verb.GetCleanToken(), "сказати") &&
		tokens[verbPos-1] != nil && tokens[verbPos-1].GetCleanToken() == "не" &&
		HasPosTagPart(noun, "v_naz") {
		return true
	}
	// потребувала мільйон — Java state.cases contains v_rod + numr/noun:numr v_zna
	if hasCaseGovFromReadings(verb, "v_rod") &&
		HasPosTagRE(noun, regexp.MustCompile(`numr.*?v_zna.*|noun.*v_zna.*numr.*`)) {
		return true
	}
	// виростили сортів 10
	if nounPos < len(tokens)-1 &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun|adj):.*:v_rod.*`)) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`num.*`)) {
		return true
	}
	// виростили сортів — 10
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun|adj):.*:v_rod.*`)) &&
		IsDash(tokens[nounPos+1]) &&
		HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`num.*`)) {
		return true
	}
	// одержав хабарів на суму
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun:inanim|adj):.:v_rod.*`)) {
		v2 := TokenSearch(tokens, nounPos+1, "", regexp.MustCompile(`^на$`), regexp.MustCompile(`^[a-z].*`), DirForward)
		if v2 >= 0 && v2 <= nounPos+5 && v2 < len(tokens)-1 {
			return true
		}
	}
	// залучити інвестицій на 20
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun.*v_(rod|zna).*`)) &&
		regexp.MustCompile(`^(?:на|з|із|зо|під)$`).MatchString(CleanTokenLower(tokens[nounPos+1])) &&
		HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`number|numr.*v_zna.*`)) {
		return true
	}

	// V_DAV block
	if HasPosTagPart(noun, "v_dav") {
		if HasPosTagPart(verb, ":inf") {
			// як боротися підприємцям
			if verbPos > 1 &&
				HasLemmaTokenAny(tokens[verbPos-1], []string{"як", "куди", "де", "що", "чого", "чи"}) {
				return true
			}
			// Квапитися їй нікуди
			if nounPos < len(tokens)-1 &&
				regexp.MustCompile(`^(?:ніколи|нікуди|нічого|нічим|ніде|немає?|не)$`).MatchString(CleanTokenLower(tokens[nounPos+1])) {
				return true
			}
			// тут жити мешканцям
			if HasLemmaTokenAny(verb, []string{"жити", "сидіти", "судити"}) {
				return true
			}
			// нічим пишатися селянам
			if verbPos > 1 &&
				regexp.MustCompile(`^(?:ніколи|нікуди|нічого|нічим|ніде|де|немає?|не)$`).MatchString(CleanTokenLower(tokens[verbPos-1])) {
				return true
			}
			// не бачити вам цирку
			if verbPos > 1 && nounPos < len(tokens)-1 &&
				regexp.MustCompile(`^(?:не|а?ні)$`).MatchString(CleanTokenLower(tokens[verbPos-1])) &&
				HasPosTagPart(tokens[nounPos+1], "v_rod") {
				return true
			}
			// слід проходити людям
			if verbPos > 1 &&
				regexp.MustCompile(`^(?:слід|снаги|силу)$`).MatchString(CleanTokenLower(tokens[verbPos-1])) {
				return true
			}
		}
		// розсміявся брату в обличчя
		if nounPos < len(tokens)-2 &&
			regexp.MustCompile(`^(?:в|у|на|від|під|по|до|і?з|з[іо]|над|з-під|перед|попід|поза|напереріз)$`).
				MatchString(CleanTokenLower(tokens[nounPos+1])) &&
			HasPosTagRE(tokens[nounPos+2], regexp.MustCompile(`(?:noun|adj).*`)) {
			return true
		}
		if nounPos < len(tokens)-1 &&
			HasPosTagRE(noun, regexp.MustCompile(`.*v_dav.*`)) &&
			regexp.MustCompile(`^(?:назустріч|навперейми|навздогін|услід)$`).MatchString(CleanTokenLower(tokens[nounPos+1])) {
			return true
		}
		// закружляли мені десь
		if nounPos < len(tokens)-2 &&
			HasPosTagRE(noun, regexp.MustCompile(`noun.*?v_dav.*:pron:(?:pers|refl).*`)) {
			return true
		}
	}

	// сміятися гріх
	if HasPosTagPart(verb, ":inf") && strings.EqualFold(noun.GetCleanToken(), "гріх") {
		return true
	}

	// дай Боже
	if HasPosTagPart(noun, "v_kly") && HasPosTagPart(verb, "impr") {
		return true
	}

	// повторила прем'єр-міністр — masc profession + fem verb
	if HasPosTagPart(noun, "noun:anim:m:v_naz") &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*:f(:.*|$)`)) &&
		HasMascFemLemma(noun) {
		return true
	}

	// не існувало конкуренції / не було мізків / стане сили
	if HasPosTagPart(noun, "v_rod") &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*?(?:futr|past):(?:s:3.*|n(?:$|:.*))`)) {
		return true
	}

	// меншає людей — Java (по)?меншати|(по)?більшати|стати + :[sn]
	if HasLemmaTokenRE(verb, regexp.MustCompile(`^(?:(?:по)?меншати|(?:по)?більшати|стати)$`)) &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*:[sn](?:$|:.*)`)) &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:noun|adj).*v_rod.*`)) {
		return true
	}

	// споживає газу менше
	if nounPos < len(tokens)-1 &&
		HasPosTagRE(noun, regexp.MustCompile(`noun:.*v_rod.*`)) &&
		regexp.MustCompile(`^(?:менше|більше)$`).MatchString(CleanTokenLower(tokens[nounPos+1])) {
		return true
	}

	// небагато надходить книжок — V_ROD_DRIVER reverse of verb
	if verbPos > 1 && HasPosTagPart(noun, "v_rod") {
		xpos := TokenSearch(tokens, verbPos-1, "", verbNounVRodDriverRE, regexp.MustCompile(`^[a-z].*`), DirReverse)
		if xpos >= 0 && xpos >= verbPos-4 {
			return true
		}
	}

	// V + N + V:INF — verb governs v_inf, second verb agrees (Java agrees on naz/indir of noun)
	if nounPos < len(tokens)-1 && hasCaseGovPosRE(verb, verbAdvpPattern, "v_inf") {
		v2 := tokenSearchPosRE(tokens, nounPos+1, verbPattern, DirForward)
		// Java tokenSearch ignores [a-z].* POS; our tokenSearchPosRE is verb-only
		if v2 >= 0 && v2 <= nounPos+5 &&
			verbNounAgrees(tokens[v2], noun) {
			return true
		}
	}

	// V:INF + N + V — робити прогнозів не буду
	if nounPos < len(tokens)-1 && HasPosTagPart(verb, ":inf") {
		v2 := tokenSearchPosRE(tokens, nounPos+1, verbPattern, DirForward)
		if v2 >= 0 && v2 <= nounPos+4 && hasCaseGovPosRE(tokens[v2], verbPattern, "v_inf") {
			if verbNounAgrees(tokens[v2], noun) {
				return true
			}
			if v2 > 0 && tokens[v2-1] != nil && tokens[v2-1].GetCleanToken() == "не" {
				return true
			}
		}
	}

	// ADVP + N + V — резюмуючи політик наголосив
	if nounPos < len(tokens)-1 && HasPosTagStart(verb, "advp") {
		v2 := tokenSearchPosRE(tokens, nounPos+1, verbPattern, DirForward)
		if v2 >= 0 && v2 <= nounPos+3 && verbNounAgrees(tokens[v2], noun) {
			return true
		}
	}

	// V + ADVP + N — пригадує посміхаючись Аскольд
	if verbPos > 1 && HasPosTagStart(verb, "advp") &&
		containsStr([]string{"посміхаючись", "сміючись"}, verb.GetCleanToken()) &&
		HasPosTagStart(tokens[verbPos-1], "verb") &&
		verbNounAgrees(tokens[verbPos-1], noun) {
		return true
	}

	// V:INF + N + ADV/predic — розібратися людям важко
	if nounPos < len(tokens)-1 && HasPosTagPart(verb, ":inf") &&
		!HasLemmaTokenRE(verb, verbNounVchytyRE) {
		for v2 := tokenSearchPosRE(tokens, nounPos+1, advPredictPattern, DirForward); v2 >= 0 && v2 <= nounPos+4; v2 = tokenSearchPosRE(tokens, v2+1, advPredictPattern, DirForward) {
			cases := caseGovPosRESet(tokens[v2], advPredictPattern)
			if len(cases) > 0 {
				list := make([]string, 0, len(cases))
				for c := range cases {
					list = append(list, c)
				}
				if HasVidmPosTag(list, noun) {
					return true
				}
			}
		}
	}

	// V:INF + N + ADJ — працювати студенти готові
	if nounPos < len(tokens)-1 && HasPosTagPart(verb, ":inf") {
		v2 := tokenSearchPosRE(tokens, nounPos+1, adjVNazPattern, DirForward)
		if v2 >= 0 && v2 <= nounPos+3 && hasCaseGovPosRE(tokens[v2], adjVNazPattern, "v_inf") {
			if gendersOverlap(gendersFromNaz(noun), gendersFromNaz(tokens[v2])) {
				return true
			}
		}
	}

	// V:INF + ADJ — працювати неспроможні
	if HasPosTagPart(verb, ":inf") &&
		HasPosTagRE(noun, adjVNazPattern) &&
		hasCaseGovPosRE(noun, adjVNazPattern, "v_inf") {
		return true
	}

	// V + V:INF + N — дають … мандрувати чотирьом
	if verbPos > 1 && HasPosTagPart(verb, ":inf") {
		lookupPos := verbPos - 1
		if verbPos > 3 &&
			HasLemmaTokenAny(tokens[verbPos-1], []string{"і", "й", "та"}) &&
			HasPosTagPart(tokens[verbPos-2], ":inf") {
			lookupPos = verbPos - 3
		}
		v2 := tokenSearchPosRE(tokens, lookupPos, verbAdvpPattern, DirReverse)
		if v2 >= 0 && v2 >= verbPos-5 {
			if hasCaseGovPosRE(tokens[v2], verbAdvpPattern, "v_inf") ||
				regexp.MustCompile(`^(?:по)?їсти$`).MatchString(verb.GetCleanToken()) {
				if verbNounAgrees(tokens[v2], noun) {
					return true
				}
				if HasPosTagRE(tokens[v2], regexp.MustCompile(`verb.*:p(?:$|:.*)`)) &&
					HasPosTagRE(noun, regexp.MustCompile(`.*v_naz.*`)) {
					return true
				}
			}
		}
	}

	// ADV + V:INF + N — важко розібратися багатьом
	if verbPos > 1 && HasPosTagPart(verb, ":inf") {
		for v2 := tokenSearchPosRE(tokens, verbPos-1, advPredictPattern, DirReverse); v2 >= 0 && v2 >= verbPos-3; v2 = tokenSearchPosRE(tokens, v2-1, advPredictPattern, DirReverse) {
			if HasPosTagRE(tokens[v2], regexp.MustCompile(`noninfl:predic.*`)) && HasPosTagPart(noun, "v_naz") {
				return true
			}
			cases := caseGovPosRESet(tokens[v2], advPredictPattern)
			if len(cases) > 0 {
				list := make([]string, 0, len(cases))
				for c := range cases {
					list = append(list, c)
				}
				if HasVidmPosTag(list, noun) {
					return true
				}
			}
		}
	}

	// ADJ + V:INF + N — зацікавлена перейняти угорська сторона
	if verbPos > 1 && HasPosTagPart(verb, ":inf") && HasPosTagPart(noun, "v_naz") {
		if regexp.MustCompile(`^(?:змозі|змогу|силі|силах)$`).MatchString(CleanTokenLower(tokens[verbPos-1])) {
			return true
		}
		v2 := tokenSearchPosRE(tokens, verbPos-1, adjVNazPattern, DirReverse)
		if v2 >= 0 && v2 >= verbPos-3 && hasCaseGovPosRE(tokens[v2], adjVNazPattern, "v_inf") {
			if gendersOverlap(gendersFromNaz(noun), gendersFromNaz(tokens[v2])) {
				return true
			}
		}
	}

	// ADJ + бути + N — adj v_naz governs v_rod
	if verbPos > 1 &&
		HasLemmaToken(verb, "бути") &&
		HasPosTagRE(tokens[verbPos-1], regexp.MustCompile(`adj:.:v_naz.*`)) &&
		hasCaseGovFromReadings(tokens[verbPos-1], "v_rod") &&
		HasPosTagRE(noun, regexp.MustCompile(`(?:adj|noun).*v_rod.*`)) {
		return true
	}

	// V:IMPERS + бути + N
	if verbPos > 1 &&
		HasLemmaToken(verb, "бути") &&
		HasPosTagRE(tokens[verbPos-1], regexp.MustCompile(`verb.*impers.*`)) &&
		verbNounAgrees(tokens[verbPos-1], noun) {
		return true
	}

	// NOUN + V:INF + N — гріх зайнятися Генеральній прокуратурі
	if verbPos > 1 && HasPosTagPart(verb, ":inf") &&
		(HasPosTagPart(noun, "v_dav") || HasPosTagPart(noun, "v_rod") ||
			HasPosTagRE(noun, regexp.MustCompile(`adj:.:v_naz.*`))) {
		v2 := tokenSearchPosRE(tokens, verbPos-1, nounVNazPattern, DirReverse)
		if v2 >= 0 && v2 >= verbPos-3 && hasCaseGovPosRE(tokens[v2], nounVNazPattern, "v_inf") {
			// exc: бажання вчитися новому
			if HasPosTagPart(noun, "v_dav") && HasLemmaTokenRE(verb, verbNounVchytyRE) {
				// Java: return false (not exception)
			} else {
				return true
			}
		}
	}

	// V:INF + V + N — платити доведеться повну вартість
	// (verb at verbPos governs v_inf; reverse search finds :inf)
	if verbPos > 1 && hasCaseGovPosRE(verb, verbPattern, "v_inf") {
		v2 := tokenSearchPosRE(tokens, verbPos-1, verbPattern, DirReverse)
		if v2 >= 0 && v2 >= verbPos-3 &&
			HasPosTagPart(tokens[v2], ":inf") &&
			verbNounAgrees(tokens[v2], noun) {
			return true
		}
	}

	// в мені наростали впевненість і …
	if nounPos < len(tokens)-2 &&
		HasPosTagRE(verb, regexp.MustCompile(`verb.*:p(?:$|:.*)`)) &&
		HasPosTagPart(noun, ":v_naz") {
		return true
	}

	// змалював дивовижної краси церкву — adj:v_rod + noun:v_rod + noun/adj
	// Java: adj:.:v_rod(?!.*pron) / noun:.*v_rod(?!.*pron) / (noun|adj)(?!.*pron)
	// then agrees(verb, naz of n2, indir of n2)
	if nounPos < len(tokens)-2 &&
		hasPosWithoutPron(noun, regexp.MustCompile(`adj:.:v_rod`)) &&
		hasPosWithoutPron(tokens[nounPos+1], regexp.MustCompile(`noun:.*v_rod`)) &&
		hasPosWithoutPron(tokens[nounPos+2], regexp.MustCompile(`^(?:noun|adj)`)) {
		if verbNounAgrees(verb, tokens[nounPos+2]) {
			return true
		}
	}

	// могли б займатися структури / має також народитися / мати + inf
	if verbPos > 2 && HasPosTagPart(verb, ":inf") &&
		HasPosTagStart(tokens[verbPos-2], "verb") &&
		(HasLemmaTokenAny(tokens[verbPos-1], []string{"б", "би"}) ||
			hasAdvNotAdvp(tokens[verbPos-1]) ||
			HasLemmaWithPartPos(tokens[verbPos-2], []string{"мати"}, "verb")) {
		return true
	}

	return false
}

// Patterns for TokenAgreementVerbNounExceptionHelper.
var (
	verbNounVRodDriverRE = regexp.MustCompile(
		`(?i)^(?:не|(?:на)?с[кт]ільки|(?:най)?більше|(?:най)?менше|(?:не|за)?багато|(?:не|чи|за)?мало|трохи|годі|неможливо|а?ніж|вдосталь|купу)$`)
	verbNounVchytyRE  = regexp.MustCompile(`.*вч[аи]ти(?:ся)?$`)
	verbPattern       = regexp.MustCompile(`^verb.*`)
	adjVNazPattern    = regexp.MustCompile(`^adj:.:v_naz.*`)
	nounVNazPattern   = regexp.MustCompile(`noun.*:v_naz.*`)
	advPredictPattern = regexp.MustCompile(`^(?:adv|noninfl:predic).*`)
)

// verbNounAgrees ports TokenAgreementVerbNounExceptionHelper.agrees without State:
// 1) v_naz readings: VerbInflectionHelper overlap with verb inflections
// 2) non-v_naz (indir) readings: case government via VERB_ADVP_PATTERN
func verbNounAgrees(verb, noun *languagetool.AnalyzedTokenReadings) bool {
	if verb == nil || noun == nil {
		return false
	}
	var nazTags, indirTags []string
	for _, p := range CollectPOSTags(noun) {
		if p == "" {
			continue
		}
		if strings.Contains(p, "v_naz") {
			nazTags = append(nazTags, p)
		} else if strings.Contains(p, ":v_") {
			indirTags = append(indirTags, p)
		}
	}
	// Java: if nounAdjNazInflections non-empty → check verb∩noun/adj naz
	if len(nazTags) > 0 {
		vInf := GetVerbInflections(CollectPOSTags(verb))
		nInf := GetNounInflections(nazTags)
		nInf = append(nInf, GetAdjInflections(nazTags)...)
		if verbInflectionsOverlapLists(vInf, nInf) {
			return true
		}
	}
	// Java: if indir non-empty → case gov on VERB_ADVP_PATTERN
	if len(indirTags) > 0 {
		cases := caseGovPosRESet(verb, verbAdvpPattern)
		if len(cases) > 0 && hasVidmInTags(cases, indirTags) {
			return true
		}
	}
	return false
}

func caseGovPosRESet(tok *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp) map[string]struct{} {
	// Java getCaseGovernments(readings, Pattern) — custom govs + advp lemma map + adjp:pasv
	return LoadCaseGovernmentHelper().GetCaseGovernmentsFromReadingsRE(tok, posRE)
}

// gendersFromNaz collects gender letters from noun/adj/numr v_naz readings.
func gendersFromNaz(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	seen := map[byte]bool{}
	var b strings.Builder
	re := regexp.MustCompile(`(?:noun|adj|numr).*?:([mfnp]):v_naz|adj:([mfnp]):v_naz`)
	for _, p := range CollectPOSTags(tok) {
		m := re.FindStringSubmatch(p)
		if m == nil {
			continue
		}
		g := m[1]
		if g == "" {
			g = m[2]
		}
		if g == "" {
			continue
		}
		c := g[0]
		if !seen[c] {
			seen[c] = true
			b.WriteByte(c)
		}
	}
	return b.String()
}

// hasPosWithoutRanim is RE2-friendly stand-in for adj:.:v_zna(?!:ranim).
func hasPosWithoutRanim(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, p := range CollectPOSTags(tok) {
		if strings.Contains(p, "ranim") {
			continue
		}
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

// Java PosTagHelper.VERB_INF_PATTERN = verb.*:inf.*
var verbInfPattern = regexp.MustCompile(`verb.*:inf.*`)

// lemmaOf returns first non-empty lemma or clean token lower.
func lemmaOf(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetLemma() != nil && *r.GetLemma() != "" {
			return *r.GetLemma()
		}
	}
	return CleanTokenLower(tok)
}

func hasCaseGovFromReadings(tok *languagetool.AnalyzedTokenReadings, rvCase string) bool {
	if tok == nil {
		return false
	}
	// try noun/adj/adv prefixes used in Java hasCaseGovernment without startPos
	cg := LoadCaseGovernmentHelper()
	for _, prefix := range []string{"noun", "adj", "adv", "verb"} {
		cases := cg.GetCaseGovernmentsFromReadings(tok, prefix)
		if _, ok := cases[rvCase]; ok {
			return true
		}
	}
	// also map lookup by lemma alone
	return cg.HasCaseGovernment(lemmaOf(tok), rvCase)
}

// tokenLineBefore ports new Match().tokenLine(line).mBefore(tokens, pos) >= 0
// (Java starts matching the last line token at pos and walks backward).
func tokenLineBefore(tokens []*languagetool.AnalyzedTokenReadings, pos int, words ...string) bool {
	if pos < 0 || len(words) == 0 {
		return false
	}
	line := strings.Join(words, " ")
	return NewSearchMatch(line).MBeforeATR(tokens, pos) >= 0
}

// tokenLineAfter reports whether the surface line appears at or after pos
// (Java mAfter(tokens, pos) >= 0 / mAfter >= 1).
func tokenLineAfter(tokens []*languagetool.AnalyzedTokenReadings, pos int, words ...string) bool {
	if pos < 0 || len(words) == 0 {
		return false
	}
	line := strings.Join(words, " ")
	return NewSearchMatch(line).MAfterATR(tokens, pos) >= 0
}
