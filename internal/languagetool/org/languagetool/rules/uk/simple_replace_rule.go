package uk

import (
	"bufio"
	"embed"
	"regexp"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/derivats.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceMap  map[string][]string

	derivatsOnce sync.Once
	derivatsMap  map[string]map[string]struct{}

	// Java: Pattern.compile(".*?adjp:actv.*?:bad.*")
	adjpActvBadRE = regexp.MustCompile(`.*?adjp:actv.*?:bad.*`)
)

func loadReplace() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceMap = m
	})
	return replaceMap
}

// loadDerivats ports CaseGovernmentHelper.DERIVATIVES_MAP from /uk/derivats.txt.
// Format: derivative base_verb (space-separated; multiple verbs via ":" like case_government).
func loadDerivats() map[string]map[string]struct{} {
	derivatsOnce.Do(func() {
		m := map[string]map[string]struct{}{}
		f, err := replaceFS.Open("data/derivats.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				continue
			}
			key := parts[0]
			verbs := strings.Split(parts[1], ":")
			set, ok := m[key]
			if !ok {
				set = map[string]struct{}{}
				m[key] = set
			}
			for _, v := range verbs {
				if v != "" {
					set[v] = struct{}{}
				}
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		derivatsMap = m
	})
	return derivatsMap
}

// SimpleReplaceRule ports org.languagetool.rules.uk.SimpleReplaceRule.
// setIgnoreTaggedWords + custom isTagged; findMatches adds adjp:actv:bad, derivat,
// and :bad speller paths. SpellingSuggestions optional (Java Morfologik getSuggestionsFromDefaultDicts).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
	// SpellingSuggestions ports morfologikSpellerRule.getSpeller1().getSuggestionsFromDefaultDicts.
	// When nil, :bad speller branch fails closed (match with empty suggestions only if we still add — Java always adds match).
	// Java always creates match even with empty suggestions list.
	SpellingSuggestions func(word string) []string
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	// Warm derivats load (same static init as Java CaseGovernmentHelper).
	_ = loadDerivats()
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadReplace(),
		CaseSensitive:     false,
		CheckLemmas:       true, // Java SimpleReplaceRule default checkLemmas true
		IgnoreTaggedWords: true, // Java setIgnoreTaggedWords()
		IsTagged:          ukSimpleReplaceIsTagged,
		ID:                "UK_SIMPLE_REPLACE",
		Description:       "Пошук помилкових слів",
		ShortMsg:          "Помилка?",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "«" + tokenStr + "» - помилкове слово, виправлення: " + joinCommaUK(replacements) + "."
		},
	}
	// Java AbstractSimpleReplaceRule: Categories.MISC
	rules.InitSimpleReplaceMeta(base, messages)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

// Match ports AbstractSimpleReplaceRule.match with UK findMatches override.
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	var ruleMatches []*rules.RuleMatch
	for _, tokenReadings := range sentence.GetTokensWithoutWhitespace() {
		if tokenReadings == nil || tokenReadings.IsSentenceStart() || tokenReadings.IsImmunized() {
			continue
		}
		if r.TokenException != nil && r.TokenException(tokenReadings) {
			continue
		}
		if tokenReadings.IsIgnoredBySpeller() {
			continue
		}
		if r.IgnoreTaggedWords {
			tagged := tokenReadings.IsTagged()
			if r.IsTagged != nil {
				tagged = r.IsTagged(tokenReadings)
			}
			if tagged {
				continue
			}
		}
		matches := r.findMatchesUK(tokenReadings, sentence)
		ruleMatches = append(ruleMatches, matches...)
	}
	return ruleMatches
}

// findMatchesUK ports SimpleReplaceRule.findMatches.
func (r *SimpleReplaceRule) findMatchesUK(tokenReadings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	matches := r.AbstractSimpleReplaceRule.FindMatches(tokenReadings, sentence)
	if len(matches) > 0 {
		return matches
	}

	// Active participles adjp:actv … :bad
	if hasPosTagRE(tokenReadings, adjpActvBadRE) {
		msg := "Активні дієприкметники не властиві українській мові."
		url := ""
		lemma := firstLemma(tokenReadings)
		if strings.HasSuffix(lemma, "ший") {
			msg += " Їх можна замінити на що + дієслово (випавший сніг - сніг, що випав), або на форму з суфіксом -л- (промокший - промоклий)"
			url = "http://padaread.com/?book=53784&pg=94"
		} else {
			msg += " Їх можна замінити питомими словами в різний спосіб: що + дієслово (роблячий  - що робить), дієслівний корінь+ суфікси -льн-, -лив- тощо (збираючий - збиральний, обтяжуючий - обтяжливий)," +
				" заміна іменником (завідуючий - завідувач), заміна прикметником із відповідним значенням (діюча модель - робоча модель), зміна конструкції (з наступаючим Новим роком - з настанням Нового року) тощо."
			url = "http://nbuv.gov.ua/j-pdf/Nchnpu_8_2013_5_2.pdf"
		}
		m := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetStartPos()+utf16LenUK(tokenReadings.GetToken()), msg)
		m.ShortMessage = r.ShortMsg
		if m.ShortMessage == "" {
			m.ShortMessage = "Помилка?"
		}
		if url != "" {
			m.SetURL(url)
		}
		return []*rules.RuleMatch{m}
	}

	// Derivat path
	clean := tokenReadings.GetCleanToken()
	if clean == "" {
		clean = tokenReadings.GetToken()
	}
	derivatSuggestions := findInDeriv(strings.ToLower(clean))
	if len(derivatSuggestions) > 0 {
		msg := "Неправильне слово."
		m := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetStartPos()+utf16LenUK(tokenReadings.GetToken()), msg)
		m.ShortMessage = r.ShortMsg
		if m.ShortMessage == "" {
			m.ShortMessage = "Помилка?"
		}
		m.SetSuggestedReplacements(derivatSuggestions)
		return []*rules.RuleMatch{m}
	}

	// :bad misspelling path (not number*)
	if hasPosTagPart(tokenReadings, ":bad") && !hasPosTagStart(tokenReadings, "number") {
		msg := "Неправильно написане слово."
		m := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetStartPos()+utf16LenUK(tokenReadings.GetToken()), msg)
		m.ShortMessage = r.ShortMsg
		if m.ShortMessage == "" {
			m.ShortMessage = "Помилка?"
		}
		if r.SpellingSuggestions != nil {
			sugs := r.SpellingSuggestions(tokenReadings.GetToken())
			// Java: removeIf(s -> s.contains(" "))
			filtered := sugs[:0]
			for _, s := range sugs {
				if !strings.Contains(s, " ") {
					filtered = append(filtered, s)
				}
			}
			m.SetSuggestedReplacements(filtered)
		}
		// Without speller: still emit match (Java always does); empty suggestions.
		return []*rules.RuleMatch{m}
	}

	return nil
}

// findInDeriv ports SimpleReplaceRule.findInDeriv.
func findInDeriv(w string) []string {
	derivats := loadDerivats()
	verbSet, ok := derivats[w]
	if !ok || len(verbSet) == 0 {
		return nil
	}
	wrong := loadReplace()
	ending := ""
	if r := []rune(w); len(r) >= 3 {
		ending = string(r[len(r)-3:])
	}
	var suggestions []string
	for verb := range verbSet {
		reps, ok := wrong[verb]
		if !ok || len(reps) == 0 {
			continue
		}
		for _, t := range reps {
			chosen := t
			// Java: first derivative key ending with same 3-char ending whose value contains t
			for der, verbs := range derivats {
				if !strings.HasSuffix(der, ending) {
					continue
				}
				if _, has := verbs[t]; has {
					chosen = der
					break
				}
			}
			suggestions = append(suggestions, chosen)
		}
	}
	return suggestions
}

func hasPosTagRE(tokenReadings *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) bool {
	if tokenReadings == nil || re == nil {
		return false
	}
	for _, at := range tokenReadings.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		if re.MatchString(*at.GetPOSTag()) {
			return true
		}
	}
	return false
}

func hasPosTagPart(tokenReadings *languagetool.AnalyzedTokenReadings, part string) bool {
	if tokenReadings == nil || part == "" {
		return false
	}
	for _, at := range tokenReadings.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		if strings.Contains(*at.GetPOSTag(), part) {
			return true
		}
	}
	return false
}

func hasPosTagStart(tokenReadings *languagetool.AnalyzedTokenReadings, prefix string) bool {
	if tokenReadings == nil {
		return false
	}
	for _, at := range tokenReadings.GetReadings() {
		if at == nil || at.GetPOSTag() == nil {
			continue
		}
		if strings.HasPrefix(*at.GetPOSTag(), prefix) {
			return true
		}
	}
	return false
}

func firstLemma(tokenReadings *languagetool.AnalyzedTokenReadings) string {
	if tokenReadings == nil {
		return ""
	}
	for _, at := range tokenReadings.GetReadings() {
		if at != nil && at.GetLemma() != nil {
			return *at.GetLemma()
		}
	}
	return ""
}

// ukSimpleReplaceIsTagged ports SimpleReplaceRule.isTagged + isGoodPosTag.
func ukSimpleReplaceIsTagged(tokenReadings *languagetool.AnalyzedTokenReadings) bool {
	if tokenReadings == nil {
		return false
	}
	for _, token := range tokenReadings.GetReadings() {
		if token == nil {
			continue
		}
		if token.HasNoTag() {
			return false
		}
		posTag := token.GetPOSTag()
		if posTag != nil && ukIsGoodPosTag(*posTag) {
			return true
		}
	}
	return false
}

// ukIsGoodPosTag ports SimpleReplaceRule.isGoodPosTag.
func ukIsGoodPosTag(posTag string) bool {
	if posTag == "" {
		return false
	}
	if posTag == languagetool.ParagraphEndTagName || posTag == languagetool.SentenceEndTagName {
		return false
	}
	if strings.Contains(posTag, "bad") {
		return false
	}
	if strings.Contains(posTag, "subst") {
		return false
	}
	if strings.HasPrefix(posTag, "<") {
		return false
	}
	return true
}

func joinCommaUK(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func utf16LenUK(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
