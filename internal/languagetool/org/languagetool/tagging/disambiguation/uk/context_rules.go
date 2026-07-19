package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
)

var (
	// Java ST_ABBR / LATIN_DIGITS / DIGITS / STATION / PATTERN_2
	stAbbr             = "ст."
	stLatinDigitsRE    = regexp.MustCompile(`^[XIVХІ]+(?:[–—-][XIVХІ]+)?$`)
	stArabicDigitsRE   = regexp.MustCompile(`^[0-9]+(?:[–—-][0-9]+)?$`)
	stStationNameRE    = regexp.MustCompile(`^метро|[А-Я][а-яіїєґ'-]+$`)
	stArticlePageNumRE = regexp.MustCompile(`^[0-9]+(?:[.,–—-][0-9]+)?$`)
	stXp3KeepRE        = regexp.MustCompile(`noun.*:xp3.*`)
	stNounInanimFRE    = regexp.MustCompile(`^noun:inanim:f:.*`)
	stNounInanimPRE    = regexp.MustCompile(`^noun:inanim:p:.*`)
	stNounInanimNRE    = regexp.MustCompile(`^noun:inanim:n:.*`)
	stNounInanimNFRE   = regexp.MustCompile(`^noun:inanim:[nf]:.*`)
	stAdjFPRE          = regexp.MustCompile(`^adj:[fp]:.*`)
	stAdjMRE           = regexp.MustCompile(`^adj:m:.*`)
	stRankLemmas       = []string{"лейтенант", "сержант", "солдат", "науковий", "медсестра"}
)

// DisambiguateSt ports UkrainianHybridDisambiguator.disambiguateSt for "ст.".
// Filters existing readings only (never invents tags).
func DisambiguateSt(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil || tok.GetToken() != stAbbr {
			continue
		}

		// 10 мм рт. ст. → keep xp3; otherwise drop xp3 when i > 1
		if i > 1 && tokens[i-1] != nil {
			if tokens[i-1].GetToken() == "рт." {
				removeTokensWithout(tok, stXp3KeepRE)
				continue
			}
			// Java: Pattern.compile("(?!.*:xp3).*") — keep readings without :xp3
			removeTokensWithoutRE2NoXp3(tok)
		}

		// стаття/сторінка: next is number → f (or p for ст. ст.)
		if i < len(tokens)-1 && tokens[i+1] != nil &&
			stArticlePageNumRE.MatchString(tokens[i+1].GetToken()) {
			pat := stNounInanimFRE
			if i > 2 && tokens[i-1] != nil && tokens[i-1].GetToken() == stAbbr {
				pat = stNounInanimPRE
				removeTokensWithout(tokens[i-1], pat)
			}
			removeTokensWithout(tok, pat)
			continue
		}

		if i < len(tokens)-1 && tokens[i+1] != nil {
			next := tokens[i+1]
			// столова: ложка / л.
			if hasLemmaToken(next, "ложка") || next.GetToken() == "л." {
				removeTokensWithout(tok, stAdjFPRE)
				i++
				continue
			}
			// старший: rank lemmas
			if hasAnyLemma(next, stRankLemmas) {
				removeTokensWithout(tok, stAdjMRE)
				i++
				continue
			}
			// станція
			if stStationNameRE.MatchString(next.GetToken()) {
				removeTokensWithout(tok, stNounInanimFRE)
				i++
				continue
			}
		}

		// століття: latin / arabic digits before
		if i > 1 && tokens[i-1] != nil {
			prevTok := tokens[i-1].GetToken()
			if stLatinDigitsRE.MatchString(prevTok) {
				pat := stNounInanimNRE
				if i < len(tokens)-1 && tokens[i+1] != nil && tokens[i+1].GetToken() == stAbbr {
					pat = stNounInanimPRE
					removeTokensWithout(tokens[i+1], pat)
				}
				removeTokensWithout(tok, pat)
				i++
				continue
			}
			if stArabicDigitsRE.MatchString(prevTok) {
				pat := stNounInanimNFRE
				if i < len(tokens)-1 && tokens[i+1] != nil && tokens[i+1].GetToken() == stAbbr {
					pat = stNounInanimPRE
					removeTokensWithout(tokens[i+1], pat)
				}
				removeTokensWithout(tok, pat)
				i++
				continue
			}
		}
	}
}

// removeTokensWithout ports Java removeTokensWithout (keep SENT_END + matching POS).
func removeTokensWithout(tok *languagetool.AnalyzedTokenReadings, keep *regexp.Regexp) {
	if tok == nil || keep == nil {
		return
	}
	readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
	for _, r := range readings {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if pos == languagetool.SentenceEndTagName {
			continue
		}
		// Java Matcher.matches() = full string
		loc := keep.FindStringIndex(pos)
		if loc != nil && loc[0] == 0 && loc[1] == len(pos) {
			continue
		}
		tok.RemoveReading(r, "UkranianHybridDisambiguator")
	}
}

// removeTokensWithoutRE2NoXp3 keeps readings that do not contain :xp3 (Java (?!.*:xp3).*).
func removeTokensWithoutRE2NoXp3(tok *languagetool.AnalyzedTokenReadings) {
	if tok == nil {
		return
	}
	readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
	for _, r := range readings {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if pos == languagetool.SentenceEndTagName {
			continue
		}
		if strings.Contains(pos, ":xp3") {
			tok.RemoveReading(r, "UkranianHybridDisambiguator")
		}
	}
}

func hasLemmaToken(tok *languagetool.AnalyzedTokenReadings, lemma string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetLemma() != nil && *r.GetLemma() == lemma {
			return true
		}
	}
	return false
}

func hasAnyLemma(tok *languagetool.AnalyzedTokenReadings, lemmas []string) bool {
	for _, l := range lemmas {
		if hasLemmaToken(tok, l) {
			return true
		}
	}
	return false
}

// isNumberish reports digit-only surface or number POS (tests / soft helpers).
func isNumberish(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	if tok.HasPosTag("number") || tok.HasPartialPosTag("number") {
		return true
	}
	s := tok.GetToken()
	return stArabicDigitsRE.MatchString(s)
}

// ignoreInPronPos ports IGNORE_IN_PRON_POS (substring match on POS tags).
var ignoreInPronPosRE = regexp.MustCompile(`pron|noun:anim:p:v_zna.*:rare.*`)

// DisambiguatePronPos ports UkrainianHybridDisambiguator.disambiguatePronPos:
// for його/її/їх with adj.*pron:pos, drop adj readings that do not share gender/case
// with neighboring noun inflections (prev and/or next).
func DisambiguatePronPos(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		low := strings.ToLower(cleanOrToken(tok))
		if low != "його" && low != "її" && low != "їх" {
			continue
		}
		// Java: only if token has adj.*pron:pos
		if !hasPosTagREMatch(tok, `adj.*pron:pos.*`) {
			continue
		}
		var nounInfs []rulesuk.Inflection
		if i > 1 && tokens[i-1] != nil {
			nounInfs = append(nounInfs, nounInfsFromTok(tokens[i-1], ignoreInPronPosRE)...)
		}
		if i < len(tokens)-1 && tokens[i+1] != nil {
			nounInfs = append(nounInfs, nounInfsFromTok(tokens[i+1], ignoreInPronPosRE)...)
		}
		if len(nounInfs) == 0 {
			continue
		}
		readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
		for _, r := range readings {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			if !strings.HasPrefix(*r.GetPOSTag(), "adj") {
				continue
			}
			adjInfs := rulesuk.GetAdjCaseInflections([]string{*r.GetPOSTag()})
			if !rulesuk.InflectionsIntersect(nounInfs, adjInfs) {
				tok.RemoveReading(r, "dis_pron_pos")
			}
		}
	}
}

func nounInfsFromTok(tok *languagetool.AnalyzedTokenReadings, ignore *regexp.Regexp) []rulesuk.Inflection {
	if tok == nil {
		return nil
	}
	var tags []string
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		tags = append(tags, *r.GetPOSTag())
	}
	return rulesuk.GetNounInflectionsFromTags(tags, ignore)
}

// DisambiguateYih ports removeYih (їх/його/її → drop adj.*pron when object-like).
func DisambiguateYih(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	adjPronRE := regexp.MustCompile(`adj.*pron.*`)
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		main := tokens[i]
		if main == nil {
			continue
		}
		clean := strings.ToLower(cleanOrToken(main))
		if clean != "їх" && clean != "його" && clean != "її" {
			continue
		}
		if i < len(tokens)-1 && tokens[i+1] != nil {
			nextClean := strings.ToLower(cleanOrToken(tokens[i+1]))
			// їх кількість|розгляд|… or predic (Java hasLemma list)
			if yihHasObjectLemma(tokens[i+1]) || hasPosTagREMatch(tokens[i+1], `noninfl:predic.*`) {
				removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
				continue
			}
			// їх було — next verb only (not also adj/noun)
			if hasPOSPrefix(tokens[i+1], "verb") &&
				!hasPOSPrefix(tokens[i+1], "adj") && !hasPOSPrefix(tokens[i+1], "noun") {
				removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
				continue
			}
			// їх обох|ніхто|ніщо
			if nextClean == "обох" || nextClean == "ніхто" || nextClean == "ніщо" {
				removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
				continue
			}
			// їх я / їх на… but not "їх з"
			if (hasPosTagREMatch(tokens[i+1], `.*pron:pers.*`) || hasPOSPrefix(tokens[i+1], "prep")) &&
				nextClean != "із" && nextClean != "з" {
				removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
				continue
			}
			// їх не було
			if i < len(tokens)-2 && (nextClean == "не" || nextClean == "ні") &&
				hasPOSPrefix(tokens[i+2], "verb") {
				removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
				continue
			}
			// exclude на його душу: next not adj|noun, verb governs v_rod|v_zna
			if !hasPOSPrefix(tokens[i+1], "adj") && !hasPOSPrefix(tokens[i+1], "noun") {
				govs := caseGovForPosRE(tokens[i+1], regexp.MustCompile(`^verb`))
				if setHasAny(govs, "v_rod", "v_zna") {
					removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
					continue
				}
			}
		}
		// посунув їх — prev verb/advp governs v_rod|v_zna
		if i > 1 && tokens[i-1] != nil {
			prevGovs := caseGovForPosRE(tokens[i-1], regexp.MustCompile(`^(?:verb|advp)`))
			if setHasAny(prevGovs, "v_rod", "v_zna") {
				// end of sentence / adv / prep / punct after
				if i == len(tokens)-1 ||
					(tokens[i+1] != nil && (hasPOSPrefix(tokens[i+1], "adv") || hasPOSPrefix(tokens[i+1], "prep") ||
						regexp.MustCompile(`^[,.;\x{2013}\x{2014}-]$`).MatchString(tokens[i+1].GetToken()))) {
					removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
					continue
				}
				// примусили їх сказати — next inf + prev also v_inf
				if i < len(tokens)-1 && tokens[i+1] != nil &&
					(hasPOSPrefix(tokens[i-1], "verb") || hasPOSPrefix(tokens[i-1], "advp")) &&
					hasPosTagREMatch(tokens[i+1], `(?:verb).*:inf.*`) && setHasAny(prevGovs, "v_inf") {
					removeReadingsMatching(main, adjPronRE, "dis_yih_pron_pos")
					continue
				}
			}
		}
	}
}

func yihHasObjectLemma(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	// surface lower
	if yihObjectLemmas[strings.ToLower(cleanOrToken(tok))] {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetLemma() != nil {
			if yihObjectLemmas[strings.ToLower(*r.GetLemma())] {
				return true
			}
		}
	}
	return false
}

// caseGovForPosRE collects case governments for readings whose POS matches re
// (Java CaseGovernmentHelper.getCaseGovernments(readings, Pattern)).
func caseGovForPosRE(tok *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp) map[string]struct{} {
	out := map[string]struct{}{}
	if tok == nil || posRE == nil {
		return out
	}
	cg := rulesuk.LoadCaseGovernmentHelper()
	if cg == nil {
		return out
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil || r.GetLemma() == nil {
			continue
		}
		if !posRE.MatchString(*r.GetPOSTag()) {
			continue
		}
		for _, c := range cg.GetCaseGovernments(*r.GetLemma()) {
			out[c] = struct{}{}
		}
		// adjp:pasv adds v_oru (same as rules helper)
		if strings.Contains(*r.GetPOSTag(), "adjp:pasv") {
			out["v_oru"] = struct{}{}
		}
	}
	return out
}

func setHasAny(set map[string]struct{}, keys ...string) bool {
	for _, k := range keys {
		if _, ok := set[k]; ok {
			return true
		}
	}
	return false
}

// lemmas after їх that force personal (object) reading — Java hasLemma list (lower surface/lemma).
var verbOnlyRE = regexp.MustCompile(`^verb`)

var yihObjectLemmas = map[string]bool{
	"кількість": true, "розгляд": true, "обговорення": true, "використання": true,
	"реалізація": true, "виконання": true, "звільнення": true, "виробництво": true,
	"застосування": true, "проведення": true, "утримання": true, "вирішення": true,
	"загибель": true, "аналоги": true, "однолітки": true, "перелік": true,
	"затримання": true, "створення": true, "розміщення": true, "лікування": true,
	"втілення": true, "арешт": true, "формування": true, "наявність": true, "збереження": true,
}

func cleanOrToken(tok *languagetool.AnalyzedTokenReadings) string {
	if tok == nil {
		return ""
	}
	c := tok.GetCleanToken()
	if c == "" {
		c = tok.GetToken()
	}
	return c
}

func removeReadingsMatching(main *languagetool.AnalyzedTokenReadings, posRE *regexp.Regexp, label string) {
	if main == nil || posRE == nil {
		return
	}
	for _, r := range append([]*languagetool.AnalyzedToken(nil), main.GetReadings()...) {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if posRE.MatchString(*r.GetPOSTag()) {
			main.RemoveReading(r, label)
		}
	}
}

// RetagInitials ports checkForInitialRetag / getInitialReadings:
// when next token has :prop:lname readings, retag the initial from those tags
// (replace :prop:lname → :nv:abbr:prop:fname). Fail closed without lname POS.
func RetagInitials(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-1; i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		s := tok.GetToken()
		if !isInitialSurface(s) {
			continue
		}
		next := tokens[i+1]
		if next == nil {
			continue
		}
		// Java LAST_NAME_TAG = :prop:lname on following surname
		newReadings := initialReadingsFromLname(s, next, "fname")
		if len(newReadings) == 0 {
			continue
		}
		// replace readings (Java assigns new AnalyzedTokenReadings)
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r != nil {
				tok.RemoveReading(r, "dis_initials")
			}
		}
		for _, r := range newReadings {
			tok.AddReading(r, "dis_initials")
		}
	}
}

// initialReadingsFromLname ports getInitialReadings(initials, lname, initialType).
func initialReadingsFromLname(initialSurface string, lname *languagetool.AnalyzedTokenReadings, initialType string) []*languagetool.AnalyzedToken {
	if lname == nil {
		return nil
	}
	const lastNameTag = ":prop:lname"
	// Java PATTERN_4 strips :alt|:nv|:up\d{2}|:xp\d before replace
	stripExtra := func(pos string) string {
		for _, frag := range []string{":alt", ":nv"} {
			pos = strings.ReplaceAll(pos, frag, "")
		}
		// strip :upNN :xpN
		for {
			i := strings.Index(pos, ":up")
			if i < 0 {
				break
			}
			j := i + 3
			for j < len(pos) && pos[j] >= '0' && pos[j] <= '9' {
				j++
			}
			pos = pos[:i] + pos[j:]
		}
		for {
			i := strings.Index(pos, ":xp")
			if i < 0 {
				break
			}
			j := i + 3
			for j < len(pos) && pos[j] >= '0' && pos[j] <= '9' {
				j++
			}
			pos = pos[:i] + pos[j:]
		}
		return pos
	}
	var out []*languagetool.AnalyzedToken
	for _, lr := range lname.GetReadings() {
		if lr == nil || lr.GetPOSTag() == nil {
			continue
		}
		pos := *lr.GetPOSTag()
		if !strings.Contains(pos, lastNameTag) {
			continue
		}
		pos = stripExtra(pos)
		pos = strings.Replace(pos, lastNameTag, ":nv:abbr:prop:"+initialType, 1)
		p, l := pos, initialSurface
		out = append(out, languagetool.NewAnalyzedToken(initialSurface, &p, &l))
	}
	return out
}

func isInitialSurface(s string) bool {
	if !strings.HasSuffix(s, ".") {
		return false
	}
	core := strings.TrimSuffix(s, ".")
	rs := []rune(core)
	if len(rs) != 1 {
		return false
	}
	return unicode.Is(unicode.Cyrillic, rs[0]) || unicode.IsLetter(rs[0])
}

func hasPOSPrefix(tok *languagetool.AnalyzedTokenReadings, prefix string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), prefix) {
			return true
		}
	}
	return false
}

// fem/masc title lemmas for proper-name gender override
var femTitles = map[string]struct{}{
	"пані": {}, "місіс": {}, "місис": {}, "міс": {}, "леді": {}, "княгиня": {}, "німкеня": {},
}
var mascTitles = map[string]struct{}{
	"пан": {}, "містер": {}, "м-р": {}, "сер": {}, "князь": {}, "німець": {}, "поляк": {},
}

var likelyVklySurfaces = map[string]struct{}{
	"суде": {}, "роде": {}, "заходе": {}, "місяченьку": {}, "редакціє": {},
}

// RetagFemNames ports retagFemNames soft: title + name + past verb gender forces name gender.
func RetagFemNames(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-2; i++ {
		title := tokens[i]
		name := tokens[i+1]
		verb := tokens[i+2]
		if title == nil || name == nil || verb == nil {
			continue
		}
		// title lemma or surface
		gen := ""
		if titleHas(title, femTitles, "f") || hasPOSStart(title, "noun:anim:f:v_naz:prop:fname") {
			gen = "f"
		} else if titleHas(title, mascTitles, "m") || hasPOSStart(title, "noun:anim:m:v_naz:prop:fname") {
			gen = "m"
		} else {
			continue
		}
		// past verb of same gender
		if !hasPastGender(verb, gen) {
			continue
		}
		prefix := "noun:anim:" + gen + ":v_naz:prop"
		if hasPOSStart(name, prefix) {
			// drop non-matching gender prop readings
			for _, r := range append([]*languagetool.AnalyzedToken(nil), name.GetReadings()...) {
				if r == nil || r.GetPOSTag() == nil {
					continue
				}
				if !strings.HasPrefix(*r.GetPOSTag(), prefix) {
					name.RemoveReading(r, "proper_name_gender_override")
				}
			}
		} else if gen == "f" && hasPOSStart(name, "noun:anim:m:v_naz:prop") {
			// леді Черчилль → retag as fem lname
			for _, r := range append([]*languagetool.AnalyzedToken(nil), name.GetReadings()...) {
				name.RemoveReading(r, "proper_name_gender_override")
			}
			p := "noun:anim:f:v_naz:prop:lname"
			l := name.GetToken()
			name.AddReading(languagetool.NewAnalyzedToken(name.GetToken(), &p, &l), "proper_name_gender_override")
		}
		i++ // skip name
	}
}

func titleHas(tok *languagetool.AnalyzedTokenReadings, set map[string]struct{}, gen string) bool {
	low := strings.ToLower(tok.GetToken())
	if _, ok := set[low]; ok {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		if r.GetLemma() != nil {
			if _, ok := set[strings.ToLower(*r.GetLemma())]; ok {
				return true
			}
		}
	}
	// also require anim gender hint soft
	_ = gen
	return false
}

func hasPastGender(tok *languagetool.AnalyzedTokenReadings, gen string) bool {
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if strings.HasPrefix(pos, "verb") && strings.Contains(pos, "past") && strings.Contains(pos, ":"+gen) {
			return true
		}
	}
	return false
}

func hasPOSStart(tok *languagetool.AnalyzedTokenReadings, prefix string) bool {
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), prefix) {
			return true
		}
	}
	return false
}

// RemoveInanimVKly drops inanim vocative when other cases remain and context is not vocative.
func RemoveInanimVKly(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		if likelyVklyContext(tokens, i) {
			continue
		}
		readings := tok.GetReadings()
		var vkly []*languagetool.AnalyzedToken
		other := false
		for _, r := range readings {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") || pos == "SENT_END" {
				continue
			}
			// inanim v_kly not geo
			if strings.Contains(pos, "noun:inanim:") && strings.Contains(pos, "v_kly") && !strings.Contains(pos, ":geo") {
				vkly = append(vkly, r)
			} else {
				other = true
			}
		}
		if len(vkly) == 0 || !other {
			continue
		}
		for _, r := range vkly {
			if r.GetLemma() != nil && *r.GetLemma() == "зоря" {
				continue
			}
			tok.RemoveReading(r, "inanim_v_kly")
		}
	}
}

func likelyVklyContext(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if tokens[i] == nil {
		return false
	}
	if _, ok := likelyVklySurfaces[strings.ToLower(tokens[i].GetToken())]; ok {
		return true
	}
	if i >= len(tokens)-1 {
		return false
	}
	next := tokens[i+1].GetToken()
	if !isPunctAfterKly(next) {
		return false
	}
	if i > 0 && tokens[i-1] != nil {
		prev := strings.ToLower(tokens[i-1].GetToken())
		if prev == "о" {
			return true
		}
		if hasPOSPrefix(tokens[i-1], "adj") && tokens[i-1].HasPartialPosTag("v_kly") {
			return true
		}
	}
	return false
}

func isPunctAfterKly(s string) bool {
	if s == "!" || s == "?" || s == "," || s == "»" || s == "\"" || s == "…" {
		return true
	}
	if strings.HasPrefix(s, "..") || strings.HasPrefix(s, "...") {
		return true
	}
	return false
}

// RemoveLowerCaseHomonymsForAbbreviations drops non-abbr readings on ALL-CAPS abbr tokens.
func RemoveLowerCaseHomonymsForAbbreviations(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	for _, tok := range input.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		s := tok.GetToken()
		if s == "" || !isAllUpperLetters(s) {
			continue
		}
		if !tok.HasPartialPosTag("abbr") {
			continue
		}
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") {
				continue
			}
			if !strings.Contains(pos, ":abbr") {
				tok.RemoveReading(r, "lowercase_vs_abbr")
			}
		}
	}
}

func isAllUpperLetters(s string) bool {
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

// RemovePluralForNames drops plural proper-name readings unless plural context.
func RemovePluralForNames(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		// plural adj/numr/багато before → keep plural names
		if i > 1 && tokens[i-1] != nil {
			prev := tokens[i-1]
			if prev.HasPartialPosTag("adj:p") || prev.HasPartialPosTag("numr") || prev.HasPartialPosTag("number") ||
				prev.HasPartialPosTag(":num") {
				continue
			}
			switch strings.ToLower(prev.GetToken()) {
			case "багато", "мало", "сотня", "півсотня", "два", "дві", "три", "чотири":
				continue
			}
			// prep з/із before plural name (наймолодшого з Моцартів)
			if isPrepZ(prev) {
				continue
			}
		}
		// next is plural lname → keep
		if i+1 < len(tokens) && tokens[i+1] != nil && tokens[i+1].HasPartialPosTag(":lname") && tokens[i+1].HasPartialPosTag(":p:") {
			continue
		}
		var plurals []*languagetool.AnalyzedToken
		other := false
		for _, r := range tok.GetReadings() {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") {
				continue
			}
			// plural prop fname/lname
			if strings.Contains(pos, ":prop") && (strings.Contains(pos, ":p:") || strings.Contains(pos, ":p:v_")) &&
				(strings.Contains(pos, "fname") || strings.Contains(pos, "lname") || strings.Contains(pos, "geo")) {
				plurals = append(plurals, r)
			} else {
				other = true
			}
		}
		if len(plurals) > 0 && other {
			for _, r := range plurals {
				tok.RemoveReading(r, "plural_for_names")
			}
		}
	}
}

func isPrepZ(tok *languagetool.AnalyzedTokenReadings) bool {
	low := strings.ToLower(tok.GetToken())
	if low == "з" || low == "із" || low == "зі" {
		return true
	}
	return tok.HasPartialPosTag("prep") && (low == "з" || low == "із" || low == "зі")
}

// RemoveLowerCaseBadForUpperCaseGood strips :bad readings when surface is capitalized prop.
func RemoveLowerCaseBadForUpperCaseGood(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	for _, tok := range input.GetTokensWithoutWhitespace() {
		if tok == nil || len(tok.GetReadings()) < 2 {
			continue
		}
		s := tok.GetToken()
		if s == "" {
			continue
		}
		rs := []rune(s)
		if !unicode.IsUpper(rs[0]) {
			continue
		}
		if !tok.HasPartialPosTag("prop") {
			continue
		}
		// drop readings with :bad whose lemma equals lowercased form of another lemma
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			if strings.Contains(*r.GetPOSTag(), ":bad") {
				tok.RemoveReading(r, "lowercase_bad_vs_uppercase_good")
			}
		}
	}
}

// RemoveVerbImpr drops verb:impr when token is also noun and previous adj agrees in case/gender soft.
func RemoveVerbImpr(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 2; i < len(tokens); i++ {
		tok, prev := tokens[i], tokens[i-1]
		if tok == nil || prev == nil {
			continue
		}
		if !hasPOSPrefix(tok, "verb") || !hasPOSPrefix(tok, "noun") || !hasPOSPrefix(prev, "adj") {
			continue
		}
		hasImpr := false
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "impr") {
				hasImpr = true
				break
			}
		}
		if !hasImpr {
			continue
		}
		// Drop impr verb reading when adj+noun share gender/number (POS-gated, no surface invent).
		if adjNounAgree(prev, tok) {
			for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
				if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "verb") && strings.Contains(*r.GetPOSTag(), "impr") {
					tok.RemoveReading(r, "not_an_imperative_2")
				}
			}
		}
	}
}

func adjNounAgree(adj, noun *languagetool.AnalyzedTokenReadings) bool {
	// share :p: or same :m/f/n: on adj and noun readings
	for _, a := range adj.GetReadings() {
		if a == nil || a.GetPOSTag() == nil {
			continue
		}
		ap := *a.GetPOSTag()
		for _, n := range noun.GetReadings() {
			if n == nil || n.GetPOSTag() == nil || !strings.HasPrefix(*n.GetPOSTag(), "noun") {
				continue
			}
			np := *n.GetPOSTag()
			for _, g := range []string{":p:", ":m:", ":f:", ":n:"} {
				if strings.Contains(ap, g) && strings.Contains(np, g) {
					return true
				}
			}
		}
	}
	return false
}

// RetagPluralProp ports retagPulralProp: дві Франції → invent p:v_naz prop from f/m/n v_rod prop.
func RetagPluralProp(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	// Java PATTERN_3
	numrRE := regexp.MustCompile(`^(?:два|дві|три|чотири)$`)
	// PATTERN_5 = :[mfn]:v_rod → :p:v_naz
	rodGenderRE := regexp.MustCompile(`:[mfn]:v_rod`)
	tokens := input.GetTokensWithoutWhitespace()
	for i := 2; i < len(tokens); i++ {
		prop := tokens[i]
		prev := tokens[i-1]
		if prop == nil || prev == nil {
			continue
		}
		if !numrRE.MatchString(strings.ToLower(prev.GetCleanToken())) {
			// also try GetToken
			if !numrRE.MatchString(strings.ToLower(prev.GetToken())) {
				continue
			}
		}
		// skip if already has plural or singular naz prop
		if hasPosTagREMatch(prop, `noun.*:p:v_naz.*:prop.*`) ||
			hasPosTagREMatch(prop, `noun.*:[mfn]:v_naz.*:prop.*`) {
			continue
		}
		// filter noun:.*:[fmn]:v_rod.*prop.* with m: only if lemma ends with а/о
		var propOnly []*languagetool.AnalyzedToken
		for _, r := range prop.GetReadings() {
			if r == nil || r.GetPOSTag() == nil || r.GetLemma() == nil {
				continue
			}
			pos, lem := *r.GetPOSTag(), *r.GetLemma()
			if !strings.HasPrefix(pos, "noun:") || !strings.Contains(pos, "prop") {
				continue
			}
			if !regexp.MustCompile(`noun:.*:[fmn]:v_rod`).MatchString(pos) {
				continue
			}
			if strings.Contains(pos, ":m:") && !strings.HasSuffix(lem, "а") && !strings.HasSuffix(lem, "о") {
				continue
			}
			propOnly = append(propOnly, r)
		}
		if len(propOnly) == 0 {
			continue
		}
		src := propOnly[0]
		postag := rodGenderRE.ReplaceAllString(*src.GetPOSTag(), ":p:v_naz")
		lemma := *src.GetLemma()
		// clear readings
		for _, r := range append([]*languagetool.AnalyzedToken(nil), prop.GetReadings()...) {
			if r != nil {
				prop.RemoveReading(r, "dis_plural_prop")
			}
		}
		p, l := postag, lemma
		prop.AddReading(languagetool.NewAnalyzedToken(prop.GetToken(), &p, &l), "dis_plural_prop")
		i++ // Java i++ after retag
	}
}

// RetagUnknownInitials ports retagUnknownInitials: А. without name → noninfl:abbr.
func RetagUnknownInitials(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	// Java INITIAL_REGEX = [А-ЯІЇЄҐ]\.
	initRE := regexp.MustCompile(`^[А-ЯІЇЄҐ]\.$`)
	tokens := input.GetTokensWithoutWhitespace()
	// Java uses getTokens() including whitespace tokens — we use without whitespace.
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		if !initRE.MatchString(tok.GetToken()) {
			continue
		}
		if tok.HasPartialPosTag("name") {
			continue
		}
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r != nil {
				tok.RemoveReading(r, "dis_unknown_initials")
			}
		}
		p := "noninfl:abbr"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &p, nil), "dis_unknown_initials")
	}
}

func hasPosTagREMatch(tok *languagetool.AnalyzedTokenReadings, pattern string) bool {
	if tok == nil {
		return false
	}
	re := regexp.MustCompile(pattern)
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && re.MatchString(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}

// PreferVocativeWhenBang keeps only v_kly on adj+noun immediately before "!" (звертання).
func PreferVocativeWhenBang(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil || tokens[i].GetToken() != "!" {
			continue
		}
		// look back at noun then adj
		for j := i - 1; j >= 1 && j >= i-3; j-- {
			tok := tokens[j]
			if tok == nil {
				continue
			}
			if !tok.HasPartialPosTag("v_kly") {
				continue
			}
			// has other cases too → keep only v_kly
			for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
				if r == nil || r.GetPOSTag() == nil {
					continue
				}
				pos := *r.GetPOSTag()
				if strings.HasSuffix(pos, "_END") {
					continue
				}
				if !strings.Contains(pos, "v_kly") {
					tok.RemoveReading(r, "vkly_zvert")
				}
			}
		}
	}
}
