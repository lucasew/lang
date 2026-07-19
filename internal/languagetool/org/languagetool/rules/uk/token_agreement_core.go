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

// --- Exception helper stubs (full tables deferred) ---

// IsAdjNounException ports TokenAgreementAdjNounExceptionHelper early arms
// (full 1300-line table still deferred). FAKE_FEM uses Java lemma+partPos.
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

	return false
}

// HasPosTagRE reports whether any POS matches re (Java PosTagHelper.hasPosTag Pattern).
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

// HasLemmaTokenRE ports LemmaHelper.hasLemma(readings, lemmaRegex).
func HasLemmaTokenRE(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tok == nil || re == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetLemma() != nil && re.MatchString(*r.GetLemma()) {
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

// IsPrepNounException ports TokenAgreementPrepNounExceptionHelper early arms
// (full table deferred). Invalid index layout is treated as exception (no flag).
func IsPrepNounException(tokens []*languagetool.AnalyzedTokenReadings, prepPos, nounPos int) bool {
	if prepPos < 0 || nounPos <= prepPos || prepPos >= len(tokens) || nounPos >= len(tokens) {
		return true
	}
	prep, noun := tokens[prepPos], tokens[nounPos]
	if prep == nil || noun == nil {
		return true
	}
	prepLower := CleanTokenLower(prep)
	nounClean := noun.GetCleanToken()
	if nounClean == "" {
		nounClean = noun.GetToken()
	}
	nounLower := strings.ToLower(nounClean)

	// на дивом уцілілій техніці
	if nounClean == "дивом" {
		return true
	}
	// в тисяча шістсот …
	if nounPos < len(tokens)-1 && nounClean == "тисяча" {
		next := tokens[nounPos+1]
		if HasPosTagPart(next, "numr") || HasLemmaToken(next, "якийсь") {
			return true
		}
	}

	if prepLower == "на" {
		// на (свято) Купала — capitalized + v_rod
		if IsCapitalized(nounClean) && HasPosTagRE(noun, regexp.MustCompile(`noun.*?:.:v_rod.*`)) {
			return true
		}
		// на (ім'я/прізвище) …
		if HasPosTagRE(noun, regexp.MustCompile(`noun:anim:.:v_naz:prop:[fl]name.*`)) {
			if prepPos > 1 && (isNameToken(tokens[prepPos-1]) || (prepPos > 2 && isNameLemma(tokens[prepPos-2]))) {
				return true
			}
		}
		if nounLower == "ти" || nounLower == "ви" {
			return true
		}
		if nounPos < len(tokens)-1 && nounClean == "Піп" && tokens[nounPos+1] != nil &&
			tokens[nounPos+1].GetCleanToken() == "Іван" {
			return true
		}
		if nounLower == "манер" {
			return true
		}
	}
	// справедливості заради
	if prepPos > 0 && prepLower == "заради" {
		prev := CleanTokenLower(tokens[prepPos-1])
		// Java (?iu)справедливості|об.єктивності — CleanTokenLower already lowercases
		if regexp.MustCompile(`^(справедливості|об.єктивності)$`).MatchString(prev) {
			return true
		}
	}
	// при їх …
	if prepLower == "при" && nounClean == "їх" {
		return true
	}
	// з рана
	if prepLower == "з" && nounClean == "рана" {
		return true
	}
	// від а/рана/корки/мала
	if prepLower == "від" {
		if strings.EqualFold(nounClean, "а") || nounClean == "рана" || nounClean == "корки" || nounClean == "мала" {
			return true
		}
	}
	// до я/корки/велика
	if prepLower == "до" {
		if strings.EqualFold(nounClean, "я") || nounClean == "корки" || nounClean == "велика" {
			return true
		}
	}

	if nounPos < len(tokens)-1 {
		next := tokens[nounPos+1]
		// від мінус 1 / плюс 1
		if (HasPosTagStart(next, "num") || (next != nil && next.GetToken() == "$")) &&
			IsPlusMinusLemma(nounLower) {
			return true
		}
		// на мохом стеленому — skip v_oru before adjp:pasv (Java RuleException(1) → treat as exception)
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*?:v_oru.*`)) &&
			next != nil && next.HasPartialPosTag("adjp:pasv") {
			return true
		}
		if nounClean == "святая" && next != nil && next.GetToken() == "святих" {
			return true
		}
		// через/на + TIME_PLUS p:v_rod|v_zna + num
		if prepLower == "через" || prepLower == "на" {
			if HasLemmaWithPosRE(noun, TimePlusLemmaList(), regexp.MustCompile(`noun:inanim:p:v_(rod|zna).*`)) &&
				(next.HasPartialPosTag("num") ||
					(nounPos < len(tokens)-2 &&
						HasLemmaTokenAny(next, []string{"зо", "з", "із"}) &&
						tokens[nounPos+2] != nil && tokens[nounPos+2].HasPartialPosTag("num"))) {
				return true
			}
		}
		// noun v_dav refl/pers + подібн*
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_dav.*:pron:(refl|pers).*`)) &&
			strings.HasPrefix(CleanTokenLower(next), "подібн") {
			return true
		}
		if (nounClean == "усім" || nounClean == "всім") && strings.HasPrefix(CleanTokenLower(next), "відом") {
			return true
		}
		if prepLower == "до" && nounClean == "схід" && next != nil && next.GetCleanToken() == "сонця" {
			return true
		}
	}
	if nounPos < len(tokens)-2 {
		// adj m/f/n v_rod + matching gender noun v_rod → skip (Java RuleException(1))
		if HasPosTagRE(noun, regexp.MustCompile(`adj:[mfn]:v_rod.*`)) {
			genders := gendersFromPos(noun, regexp.MustCompile(`adj:([mfn]):v_rod.*`))
			if genders != "" && HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*?:[`+genders+`]:v_rod.*`)) {
				return true
			}
		}
		// нікому/ніким… + не
		if HasPosTagRE(noun, regexp.MustCompile(`noun.*v_(dav|oru).*:pron:neg.*`)) &&
			tokens[nounPos+1] != nil && tokens[nounPos+1].GetCleanToken() == "не" {
			return true
		}
	}

	return false
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

// IsNumrNounException ports TokenAgreementNumrNounExceptionHelper surface arms
// (inflection-overlap arms deferred). Invalid layout → exception.
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

	// для багатьох/обох/двох/… — Java full-string matches
	if regexp.MustCompile(`^(багать(ох|ом|ма)|обо(х|м|ма)|(дв|трь|чотирь)о[хм]|скільки(сь)?(-небудь)?|стільки)$`).MatchString(numrLower) {
		return true
	}
	// плюс|мінус|ранку|…
	if regexp.MustCompile(`^(плюс|мінус|ранку|вечора|ночі|тепла|морозу|родом|зростом|дивом|станом|вагою|слід|типу|формату|вартістю|році|населення)$`).MatchString(nounLower) {
		return true
	}
	// lemma set on noun
	if HasLemmaTokenRE(noun, regexp.MustCompile(`^(у?весь|який(сь)?|свій|сам|цей|решта|кількість|вартий|кожний|жодний|менший|більший|вищий|нижчий)$`)) {
		return true
	}
	// півтора + adj:p + noun:p:v_naz
	if nounPos < len(tokens)-1 &&
		regexp.MustCompile(`^(один-|одне-)?півтора|(одна-)?півтори$`).MatchString(CleanTokenLower(numr)) &&
		HasPosTagRE(noun, regexp.MustCompile(`adj:p:v_(naz|rod).*`)) &&
		HasPosTagRE(tokens[nounPos+1], regexp.MustCompile(`noun.*?:p:v_naz.*`)) {
		return true
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
		return true
	}

	return false
}

// IsNounVerbException stub.
func IsNounVerbException(tokens []*languagetool.AnalyzedTokenReadings, nounPos, verbPos int) bool {
	return nounPos < 0 || verbPos <= nounPos
}

// IsVerbNounException stub.
func IsVerbNounException(tokens []*languagetool.AnalyzedTokenReadings, verbPos, nounPos int) bool {
	return verbPos < 0 || nounPos <= verbPos
}
