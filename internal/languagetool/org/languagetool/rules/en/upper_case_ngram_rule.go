package en

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UpperCaseNgramThreshold ports UpperCaseNgramRule.THRESHOLD.
const UpperCaseNgramThreshold = 50

// THRESHOLD is the Java-name twin of UpperCaseNgramRule.THRESHOLD.
const THRESHOLD = UpperCaseNgramThreshold

var (
	// Java Pattern.matcher(...).matches() = full-string match.
	upperCaseNgramPunctPattern = regexp.MustCompile(`^[.!?:]$`)
	upperCaseNgramTypicalLower = regexp.MustCompile(`^(and|or|the|of|on|with|to|it|in|for|as|at|his|her|its|into|&|/)$`)
	upperCaseNgramExceptions   = map[string]struct{}{
		"Bin": {}, "Spot": {},
		"Go":  {},
		"French": {}, "Roman": {}, "Hawking": {}, "Square": {}, "Japan": {}, "Premier": {}, "Allied": {},
		"Counsel": {},
		"Act": {}, "Acts": {}, "Associate": {}, "Associates": {}, "Bill": {}, "Bills": {}, "Board": {},
		"Center": {}, "Channel": {}, "Fantasy": {}, "Judge": {}, "Mass": {}, "Staff": {}, "Trust": {},
	}

	upperCaseNgramSpellerOnce sync.Once
	upperCaseNgramSpeller     *MorfologikVariantSpellerRule
	upperCaseNgramLinguMu     sync.Mutex
	upperCaseNgramLingu       *languagetool.LinguServices

	upperCaseNgramTrieOnce sync.Once
	upperCaseNgramTrie     *upperCaseExceptionTrie

	upperCaseNgramAntiOnce  sync.Once
	upperCaseNgramAntiRules []*disambigrules.DisambiguationPatternRule
)

// UpperCaseNgramRule ports org.languagetool.rules.en.UpperCaseNgramRule.
type UpperCaseNgramRule struct {
	messages  map[string]string
	lm        ngrams.LanguageModel
	langCode  string
	Category  *rules.Category
	IssueType rules.ITSIssueType

	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewUpperCaseNgramRule ports UpperCaseNgramRule(messages, lm, lang, userConfig).
// langCode mirrors Language for isMisspelled / LinguServices (short or short-with-variant).
func NewUpperCaseNgramRule(messages map[string]string, lm ngrams.LanguageModel, langCode string, userConfig *languagetool.UserConfig) *UpperCaseNgramRule {
	if langCode == "" {
		langCode = "en"
	}
	r := &UpperCaseNgramRule{
		messages:  messages,
		lm:        lm,
		langCode:  langCode,
		Category:  rules.CatCasing.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
	}
	r.AddExamplePair(
		rules.Wrong("This <marker>Prototype</marker> was developed by Miller et al."),
		rules.Fixed("This <marker>prototype</marker> was developed by Miller et al."),
	)
	_ = upperCaseNgramAntiPatterns()
	initUpperCaseNgramTrie()
	initUpperCaseNgramSpeller()
	if userConfig != nil {
		upperCaseNgramLinguMu.Lock()
		if upperCaseNgramLingu == nil {
			upperCaseNgramLingu = userConfig.GetLinguServices()
		}
		upperCaseNgramLinguMu.Unlock()
	}
	return r
}

func initUpperCaseNgramSpeller() {
	upperCaseNgramSpellerOnce.Do(func() {
		upperCaseNgramSpeller = NewMorfologikAmericanSpellerRule()
	})
}

func initUpperCaseNgramTrie() {
	upperCaseNgramTrieOnce.Do(func() {
		words := loadUpperCaseExceptionWords()
		upperCaseNgramTrie = buildUpperCaseExceptionTrie(words)
	})
}

func loadUpperCaseExceptionWords() []string {
	var words []string
	if f, err := specificCaseFS.Open("data/specific_case.txt"); err == nil {
		words = append(words, loadCachingWordListReader(f)...)
		_ = f.Close()
	}
	if p := discoverSpellingGlobalForUpperCase(); p != "" {
		if f, err := os.Open(p); err == nil {
			words = append(words, loadCachingWordListReader(f)...)
			_ = f.Close()
		}
	}
	return words
}

// loadCachingWordListReader ports CachingWordListLoader.loadWords line rules.
func loadCachingWordListReader(r io.Reader) []string {
	var out []string
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = tools.JavaStringTrim(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func discoverSpellingGlobalForUpperCase() string {
	if p := os.Getenv("LANG_SPELLING_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "spelling_global.txt")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// --- exception trie (Aho-Corasick multi-pattern; hit offsets UTF-16 like Java) ---

type upperCaseExceptionHit struct {
	begin, end int
}

type upperCaseExceptionTrie struct {
	root *ucACNode
}

type ucACNode struct {
	next map[byte]*ucACNode
	fail *ucACNode
	out  []int // pattern byte lengths ending here
}

func buildUpperCaseExceptionTrie(words []string) *upperCaseExceptionTrie {
	root := &ucACNode{next: map[byte]*ucACNode{}}
	for _, w := range words {
		if w == "" {
			continue
		}
		n := root
		for i := 0; i < len(w); i++ {
			b := w[i]
			ch, ok := n.next[b]
			if !ok {
				ch = &ucACNode{next: map[byte]*ucACNode{}}
				n.next[b] = ch
			}
			n = ch
		}
		n.out = append(n.out, len(w))
	}
	// BFS fail links
	q := make([]*ucACNode, 0, 256)
	for _, ch := range root.next {
		ch.fail = root
		q = append(q, ch)
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		for b, ch := range n.next {
			q = append(q, ch)
			f := n.fail
			for f != nil && f.next[b] == nil {
				f = f.fail
			}
			if f == nil {
				ch.fail = root
			} else {
				ch.fail = f.next[b]
				if ch.fail == nil {
					ch.fail = root
				}
			}
			if ch.fail != nil && len(ch.fail.out) > 0 {
				ch.out = append(ch.out, ch.fail.out...)
			}
		}
	}
	return &upperCaseExceptionTrie{root: root}
}

func (t *upperCaseExceptionTrie) parseText(text string) []upperCaseExceptionHit {
	if t == nil || t.root == nil || text == "" {
		return nil
	}
	// byte index → UTF-16 offset
	byteToU16 := make([]int, len(text)+1)
	u16 := 0
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		for k := 0; k < size; k++ {
			byteToU16[i+k] = u16
		}
		if r >= 0x10000 {
			u16 += 2
		} else {
			u16++
		}
		i += size
	}
	byteToU16[len(text)] = u16

	var hits []upperCaseExceptionHit
	n := t.root
	for i := 0; i < len(text); i++ {
		b := text[i]
		for n != t.root && n.next[b] == nil {
			n = n.fail
			if n == nil {
				n = t.root
				break
			}
		}
		if ch := n.next[b]; ch != nil {
			n = ch
		} else {
			n = t.root
		}
		for _, plen := range n.out {
			beginB := i + 1 - plen
			if beginB < 0 {
				continue
			}
			hits = append(hits, upperCaseExceptionHit{
				begin: byteToU16[beginB],
				end:   byteToU16[i+1],
			})
		}
	}
	return hits
}

func upperCaseNgramAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	upperCaseNgramAntiOnce.Do(func() {
		aps := UpperCaseNgramAntiPatterns
		upperCaseNgramAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			// Java makeAntiPatterns / cacheAntiPatterns: INTERNAL_ANTIPATTERN + IMMUNIZE
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "en",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			upperCaseNgramAntiRules = append(upperCaseNgramAntiRules, rule)
		}
	})
	return upperCaseNgramAntiRules
}

// GetAntiPatterns ports UpperCaseNgramRule.getAntiPatterns.
func (r *UpperCaseNgramRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return upperCaseNgramAntiPatterns()
}

func (r *UpperCaseNgramRule) GetID() string { return "EN_UPPER_CASE_NGRAM" }

func (r *UpperCaseNgramRule) GetDescription() string {
	return "Checks wrong uppercase spelling of words that are not proper nouns"
}

func (r *UpperCaseNgramRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *UpperCaseNgramRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

func (r *UpperCaseNgramRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	r.incorrectExamples = append(r.incorrectExamples, incorrect)
	r.correctExamples = append(r.correctExamples, correct)
}

func getSentenceWithUpperCaseNgramImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := upperCaseNgramAntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	immunized := sentence.Copy(sentence)
	for _, ap := range aps {
		if ap == nil {
			continue
		}
		if next := ap.Replace(immunized); next != nil {
			immunized = next
		}
	}
	return immunized
}

// Match ports UpperCaseNgramRule.match.
func (r *UpperCaseNgramRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || sentence == nil {
		return nil, nil
	}
	var matches []*rules.RuleMatch
	tokens := getSentenceWithUpperCaseNgramImmunization(sentence).GetTokensWithoutWhitespace()
	atSentStart := true

	if !r.isSentence(tokens) {
		// might be a headline, so skip it
		return matches, nil
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token == nil {
			continue
		}
		tokenStr := token.GetToken()
		if len(tokenStr) > 0 &&
			!token.IsImmunized() &&
			tools.StartsWithUppercase(tokenStr) &&
			!tools.IsAllUppercase(tokenStr) &&
			!atSentStart &&
			token.HasPosTagStartingWith("VB") &&
			!token.HasPosTagStartingWith("NNP") &&
			token.IsTagged() &&
			(!r.prevIsUpperCase(tokens, i) || (r.prevIsUpperCase(tokens, i) && i == 2)) &&
			!r.nextIsUpperCase(tokens, i) &&
			!r.prevIsOneOf(tokens, i, []string{":", "née", "of", "\"", "'"}) &&
			!r.nextIsOneOfThenUppercase(tokens, i, []string{"of"}) &&
			tokenStr != "I" &&
			!upperCaseNgramException(tokenStr) &&
			!r.trieMatches(sentence.GetText(), token) &&
			!r.maybeTitle(tokens, i) &&
			!r.IsMisspelled(tools.LowercaseFirstChar(tokenStr)) {
			if i+1 < len(tokens) {
				ucList := []string{tokens[i-1].GetToken(), tokenStr, tokens[i+1].GetToken()}
				lcList := []string{tokens[i-1].GetToken(), tools.LowercaseFirstChar(tokenStr), tokens[i+1].GetToken()}
				if r.lm != nil {
					ucProb := r.lm.GetPseudoProbability(ucList)
					lcProb := r.lm.GetPseudoProbability(lcList)
					ratio := lcProb.GetProb() / ucProb.GetProb()
					if ratio > UpperCaseNgramThreshold {
						msg := "Only proper nouns start with an uppercase character (there are exceptions for headlines)."
						match := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
						match.SetSuggestedReplacement(tools.LowercaseFirstChar(tokenStr))
						matches = append(matches, match)
					}
				}
			}
		}
		if !token.IsSentenceStart() && tokenStr != "" && !token.IsNonWord() {
			atSentStart = false
		}
	}
	return matches, nil
}

func upperCaseNgramException(tokenStr string) bool {
	_, ok := upperCaseNgramExceptions[tokenStr]
	return ok
}

// IsMisspelled ports UpperCaseNgramRule.isMisspelled.
func (r *UpperCaseNgramRule) IsMisspelled(word string) bool {
	initUpperCaseNgramSpeller()
	upperCaseNgramLinguMu.Lock()
	ls := upperCaseNgramLingu
	upperCaseNgramLinguMu.Unlock()
	if ls != nil {
		return !ls.IsCorrectSpell(word, r.langCode)
	}
	if upperCaseNgramSpeller == nil || upperCaseNgramSpeller.IsMisspelled == nil {
		// fail-closed: treat as misspelled → condition !isMisspelled skips match
		// Java would have constructed speller; if missing dict, Go IsMisspelled is false for empty dict
		return false
	}
	// Java synchronized (spellerRule)
	return upperCaseNgramSpeller.IsMisspelled(word)
}

// maybeTitle ports UpperCaseNgramRule.maybeTitle.
func (r *UpperCaseNgramRule) maybeTitle(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return r.FirstLongWordToLeftIsUppercase(tokens, i) || r.FirstLongWordToRightIsUppercase(tokens, i)
}

// FirstLongWordToLeftIsUppercase ports firstLongWordToLeftIsUppercase.
func (r *UpperCaseNgramRule) FirstLongWordToLeftIsUppercase(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	for i := pos - 1; i > 1; i-- {
		if r.isShortWord(tokens[i]) {
			continue
		}
		return tools.StartsWithUppercase(tokens[i].GetToken())
	}
	return false
}

// FirstLongWordToRightIsUppercase ports firstLongWordToRightIsUppercase.
func (r *UpperCaseNgramRule) FirstLongWordToRightIsUppercase(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	for i := pos + 1; i < len(tokens); i++ {
		if r.isShortWord(tokens[i]) {
			continue
		}
		return tools.StartsWithUppercase(tokens[i].GetToken())
	}
	return false
}

func (r *UpperCaseNgramRule) isShortWord(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return true
	}
	// ignore words typically spelled lowercase even in titles
	t := token.GetToken()
	if strings.TrimSpace(t) == "" || upperCaseNgramTypicalLower.MatchString(t) {
		return true
	}
	return false
}

func (r *UpperCaseNgramRule) trieMatches(text string, token *languagetool.AnalyzedTokenReadings) bool {
	initUpperCaseNgramTrie()
	if upperCaseNgramTrie == nil || token == nil {
		return false
	}
	hits := upperCaseNgramTrie.parseText(text)
	start := token.GetStartPos()
	end := token.GetEndPos()
	for _, hit := range hits {
		if hit.begin <= start && hit.end >= end {
			return true
		}
	}
	return false
}

func (r *UpperCaseNgramRule) prevIsOneOf(tokens []*languagetool.AnalyzedTokenReadings, i int, stringsList []string) bool {
	if i <= 0 || tokens[i-1] == nil {
		return false
	}
	tok := tokens[i-1].GetToken()
	for _, s := range stringsList {
		if tok == s {
			return true
		}
	}
	return false
}

func (r *UpperCaseNgramRule) nextIsOneOfThenUppercase(tokens []*languagetool.AnalyzedTokenReadings, i int, stringsList []string) bool {
	if i+2 >= len(tokens) || tokens[i+1] == nil || tokens[i+2] == nil {
		return false
	}
	mid := tokens[i+1].GetToken()
	found := false
	for _, s := range stringsList {
		if mid == s {
			found = true
			break
		}
	}
	return found && tools.StartsWithUppercase(tokens[i+2].GetToken())
}

func (r *UpperCaseNgramRule) prevIsUpperCase(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i > 0 && tokens[i-1] != nil && tools.StartsWithUppercase(tokens[i-1].GetToken())
}

func (r *UpperCaseNgramRule) nextIsUpperCase(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	return i+1 < len(tokens) && tokens[i+1] != nil && tools.StartsWithUppercase(tokens[i+1].GetToken())
}

func (r *UpperCaseNgramRule) isSentence(tokens []*languagetool.AnalyzedTokenReadings) bool {
	isSentence := false
	for i := len(tokens) - 1; i > 0; i-- {
		if tokens[i] == nil {
			continue
		}
		if upperCaseNgramPunctPattern.MatchString(tokens[i].GetToken()) {
			isSentence = true
			break
		}
		if !tokens[i].IsParagraphEnd() && !tokens[i].IsNonWord() {
			break
		}
	}
	return isSentence
}
