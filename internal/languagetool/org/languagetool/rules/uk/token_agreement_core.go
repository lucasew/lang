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

	// 22 червня — number + MONTH :m:v_rod
	if isNumberToken(numr) &&
		HasLemmaWithPosRE(noun, MonthLemmas, regexp.MustCompile(`:m:v_rod`)) {
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

// IsNounVerbException ports TokenAgreementNounVerbExceptionHelper early arms
// (full table deferred). Invalid layout → exception.
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
		HasLemmaWithPosRE(tokens[nounPos-1], []string{"два", "три", "чотири"}, regexp.MustCompile(`numr:p:v_zna`)) {
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
				// if not adj, or adj/noun inflections overlap (simplified: always true when non-adj)
				if !HasPosTagStart(tokens[foundIdx], "adj") {
					return true
				}
				// adj: exception if genders/cases can overlap — use soft non-empty POS agree
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
				// soft: if prev is finite verb, exception (full VerbInflection overlap deferred)
				if !HasPosTagRE(tokens[prevVerbIdx], verbInfPattern) {
					return true
				}
			}
		}
	}

	return false
}

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

// adjNounInflectionOverlap is a simplified stand-in for Collections.disjoint on
// noun/adj inflections (Java InflectionHelper) — true when case gender tags share a letter.
func adjNounInflectionOverlap(adj, noun *languagetool.AnalyzedTokenReadings) bool {
	if adj == nil || noun == nil {
		return false
	}
	aInf := GetAdjCaseInflections(CollectPOSTags(adj))
	nInf := GetNounCaseInflections(CollectPOSTags(noun))
	if len(aInf) == 0 || len(nInf) == 0 {
		// insufficient morph: Java may still exception for non-adj path only
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

// IsVerbNounException ports TokenAgreementVerbNounExceptionHelper early arms.
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
	// мати + v_oru (має своїм наслідком) — partial Java arm
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

	return false
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
