package languagetool

import (
	"sort"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Constants and enums from org.languagetool.JLanguageTool.

const (
	SentenceStartTagName = "SENT_START"
	SentenceEndTagName   = "SENT_END"
	ParagraphEndTagName  = "PARA_END"

	PatternFile                 = "grammar.xml"
	StyleFile                   = "style.xml"
	CustomPatternFile           = "grammar_custom.xml"
	FalseFriendFile             = "false-friends.xml"
	MessageBundleName           = "org.languagetool.MessagesBundle"
	DictionaryFilenameExtension = ".dict"
)

// Mode ports JLanguageTool.Mode.
type Mode string

const (
	ModeAll             Mode = "ALL"
	ModeTextLevelOnly   Mode = "TEXTLEVEL_ONLY"
	ModeAllButTextLevel Mode = "ALL_BUT_TEXTLEVEL_ONLY"
)

// ParagraphHandling ports JLanguageTool.ParagraphHandling.
type ParagraphHandling string

const (
	ParagraphNormal      ParagraphHandling = "NORMAL"
	ParagraphOnlyPara    ParagraphHandling = "ONLYPARA"
	ParagraphOnlyNonPara ParagraphHandling = "ONLYNONPARA"
)

// CheckCancelledCallback ports JLanguageTool.CheckCancelledCallback.
type CheckCancelledCallback func() bool

// LocalMatch is a cycle-free rule-match surface for JLanguageTool.Check
// (avoids importing rules package into languagetool).
type LocalMatch struct {
	FromPos, ToPos int
	Message        string
	ShortMessage   string
	RuleID         string
	Suggestions    []string
	// Optional rule metadata (from soft grammar XML or SoftRuleMeta).
	Description  string
	CategoryID   string
	CategoryName string
	IssueType    string
	// Priority used by CleanOverlappingLocalMatches (higher wins).
	Priority int
}

// SentenceChecker returns matches for one analyzed sentence (offsets relative to sentence text).
type SentenceChecker func(sentence *AnalyzedSentence) []LocalMatch

// TextLevelChecker returns matches across all sentences (document-relative offsets).
type TextLevelChecker func(sentences []*AnalyzedSentence) []LocalMatch

// JLanguageTool is a minimal façade for pure-Go check orchestration (growing).
// Full Java parity is not attempted here.
type JLanguageTool struct {
	LanguageCode string
	Mode         Mode
	Level        Level
	// sentenceMatchers reserved for MultiThreaded error surface.
	sentenceMatchers []func(sentence *AnalyzedSentence) error
	// checkers are pluggable sentence rules for Check.
	checkers []SentenceChecker
	// textLevelCheckers are multi-sentence rules (e.g. word-repeat-beginning).
	textLevelCheckers []struct {
		id string
		fn TextLevelChecker
	}
	// activeRuleIDs tracks rule IDs registered via AddRuleChecker (order preserved).
	activeRuleIDs []string
	// DisabledRuleIDs soft-disable matches / registration filtering.
	DisabledRuleIDs map[string]struct{}
	// DefaultOffRuleIDs are rules that registered with XML default="off" (optional packs).
	// SOFT_OPTIONAL re-enables these in addition to SOFT_OPT_* inventeds.
	DefaultOffRuleIDs map[string]struct{}
	// Cancelled optional early exit for Check.
	Cancelled CheckCancelledCallback
	// ListUnknownWords enables GetUnknownWords population during Check/AnalyzeUnknown.
	ListUnknownWords bool
	// IsKnownWord optional dictionary probe for unknown-word listing.
	IsKnownWord func(token string) bool
	// TagWord optional POS/lemma inject used by Analyze (MapWordTagger-friendly).
	TagWord func(token string) []TokenTag
	// Disambiguator optional post-tag sentence filter (multiword chunker / XML rules).
	Disambiguator SentenceDisambiguator
	// IgnoreWords soft user-dictionary / spell-ignore set (surface forms).
	IgnoreWords map[string]struct{}
	// UserConfig optional user preferences (accepted phrases, speller words).
	UserConfig *UserConfig
	unknown    map[string]struct{}
}

// SentenceDisambiguator filters/augments POS on an analyzed sentence (soft LT disambiguator hook).
type SentenceDisambiguator interface {
	Disambiguate(input *AnalyzedSentence) *AnalyzedSentence
}

func NewJLanguageTool(languageCode string) *JLanguageTool {
	return &JLanguageTool{
		LanguageCode:    languageCode,
		Mode:            ModeAll,
		Level:           LevelDefault,
		DisabledRuleIDs: map[string]struct{}{},
	}
}

func (lt *JLanguageTool) GetLanguageCode() string { return lt.LanguageCode }
func (lt *JLanguageTool) GetMode() Mode           { return lt.Mode }
func (lt *JLanguageTool) SetMode(m Mode)          { lt.Mode = m }
func (lt *JLanguageTool) GetLevel() Level         { return lt.Level }
func (lt *JLanguageTool) SetLevel(l Level)        { lt.Level = l }

// AddChecker registers a sentence-level rule for Check.
func (lt *JLanguageTool) AddChecker(c SentenceChecker) {
	if lt == nil || c == nil {
		return
	}
	lt.checkers = append(lt.checkers, c)
}

// AddRuleChecker registers a checker and records its rule ID for enable/disable.
func (lt *JLanguageTool) AddRuleChecker(ruleID string, c SentenceChecker) {
	if lt == nil || c == nil {
		return
	}
	if ruleID != "" {
		lt.activeRuleIDs = append(lt.activeRuleIDs, ruleID)
	}
	id := ruleID
	lt.checkers = append(lt.checkers, func(s *AnalyzedSentence) []LocalMatch {
		if id != "" && lt.isRuleDisabled(id) {
			return nil
		}
		return c(s)
	})
}

// AddTextLevelRuleChecker registers a multi-sentence rule (document-relative offsets).
func (lt *JLanguageTool) AddTextLevelRuleChecker(ruleID string, c TextLevelChecker) {
	if lt == nil || c == nil {
		return
	}
	if ruleID != "" {
		lt.activeRuleIDs = append(lt.activeRuleIDs, ruleID)
	}
	lt.textLevelCheckers = append(lt.textLevelCheckers, struct {
		id string
		fn TextLevelChecker
	}{id: ruleID, fn: c})
}

// DisableRule ports disableRule.
func (lt *JLanguageTool) DisableRule(ruleID string) {
	if lt == nil || ruleID == "" {
		return
	}
	if lt.DisabledRuleIDs == nil {
		lt.DisabledRuleIDs = map[string]struct{}{}
	}
	lt.DisabledRuleIDs[ruleID] = struct{}{}
}

// MarkDefaultOff records that ruleID was registered with XML default="off".
func (lt *JLanguageTool) MarkDefaultOff(ruleID string) {
	if lt == nil || ruleID == "" {
		return
	}
	if lt.DefaultOffRuleIDs == nil {
		lt.DefaultOffRuleIDs = map[string]struct{}{}
	}
	lt.DefaultOffRuleIDs[ruleID] = struct{}{}
}

// GetDefaultOffRuleIDs returns rule IDs registered with default="off".
func (lt *JLanguageTool) GetDefaultOffRuleIDs() []string {
	if lt == nil || len(lt.DefaultOffRuleIDs) == 0 {
		return nil
	}
	out := make([]string, 0, len(lt.DefaultOffRuleIDs))
	for id := range lt.DefaultOffRuleIDs {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// EnableRule ports enableRule (re-enables a previously disabled rule).
func (lt *JLanguageTool) EnableRule(ruleID string) {
	if lt == nil || lt.DisabledRuleIDs == nil {
		return
	}
	delete(lt.DisabledRuleIDs, ruleID)
}

// GetAllRegisteredRuleIDs returns every rule ID registered via AddRuleChecker / AddTextLevelRuleChecker.
func (lt *JLanguageTool) GetAllRegisteredRuleIDs() []string {
	if lt == nil {
		return nil
	}
	return append([]string(nil), lt.activeRuleIDs...)
}

// GetAllActiveRuleIDs returns registered rule IDs that are not disabled.
func (lt *JLanguageTool) GetAllActiveRuleIDs() []string {
	if lt == nil {
		return nil
	}
	var out []string
	for _, id := range lt.activeRuleIDs {
		if !lt.isRuleDisabled(id) {
			out = append(out, id)
		}
	}
	return out
}

func (lt *JLanguageTool) isRuleDisabled(id string) bool {
	if lt == nil || lt.DisabledRuleIDs == nil {
		return false
	}
	_, ok := lt.DisabledRuleIDs[id]
	return ok
}

// SetListUnknownWords ports setListUnknownWords.
func (lt *JLanguageTool) SetListUnknownWords(v bool) {
	if lt != nil {
		lt.ListUnknownWords = v
	}
}

// GetUnknownWords ports getUnknownWords (sorted unique).
func (lt *JLanguageTool) GetUnknownWords() []string {
	if lt == nil || len(lt.unknown) == 0 {
		return nil
	}
	out := make([]string, 0, len(lt.unknown))
	for w := range lt.unknown {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

// Analyze splits text into sentences and builds plain analyzed sentences.
func (lt *JLanguageTool) Analyze(text string) []*AnalyzedSentence {
	st := tokenizers.NewSRXSentenceTokenizer(lt.LanguageCode)
	parts := st.Tokenize(text)
	if len(parts) == 0 {
		if text == "" {
			return nil
		}
		parts = []string{text}
	}
	out := make([]*AnalyzedSentence, 0, len(parts))
	wt := WordTokenizerForLanguage(lt.LanguageCode)
	for _, p := range parts {
		var s *AnalyzedSentence
		if lt.TagWord != nil {
			s = AnalyzeWithTaggerAndTokenizer(p, lt.TagWord, wt)
		} else {
			s = AnalyzeWithTokenizer(p, wt)
		}
		if lt.Disambiguator != nil && s != nil {
			if d := lt.Disambiguator.Disambiguate(s); d != nil {
				s = d
			}
		}
		out = append(out, s)
	}
	return out
}

// Check runs registered checkers over analyzed sentences and returns document-offset matches.
func (lt *JLanguageTool) Check(text string) []LocalMatch {
	if lt == nil {
		return nil
	}
	if lt.Cancelled != nil && lt.Cancelled() {
		return nil
	}
	lt.unknown = map[string]struct{}{}
	sents := lt.Analyze(text)
	var out []LocalMatch
	runSentence := lt.Mode != ModeTextLevelOnly
	runTextLevel := lt.Mode != ModeAllButTextLevel

	// Map sentence-local offsets to document by searching each sentence text in remaining source.
	// AnalyzePlain token positions are relative to the sentence string.
	if runSentence {
		srcRunes := []rune(text)
		searchFrom := 0
		for _, s := range sents {
			if lt.Cancelled != nil && lt.Cancelled() {
				break
			}
			if s == nil {
				continue
			}
			stext := s.GetText()
			// find sentence start in document
			docBase := indexRunesFrom(srcRunes, []rune(stext), searchFrom)
			if docBase < 0 {
				docBase = searchFrom
			}
			if lt.ListUnknownWords {
				lt.collectUnknown(s)
			}
			for _, c := range lt.checkers {
				for _, m := range c(s) {
					m.FromPos += docBase
					m.ToPos += docBase
					out = append(out, m)
				}
			}
			searchFrom = docBase + len([]rune(stext))
		}
	} else if lt.ListUnknownWords {
		for _, s := range sents {
			lt.collectUnknown(s)
		}
	}

	if runTextLevel && len(lt.textLevelCheckers) > 0 {
		if lt.Cancelled == nil || !lt.Cancelled() {
			for _, tc := range lt.textLevelCheckers {
				if tc.id != "" && lt.isRuleDisabled(tc.id) {
					continue
				}
				out = append(out, tc.fn(sents)...)
			}
		}
	}
	return lt.filterMatchesByIgnore(text, out)
}

func (lt *JLanguageTool) collectUnknown(s *AnalyzedSentence) {
	known := lt.IsKnownWord
	if known == nil {
		// without dictionary, nothing is listed as unknown
		return
	}
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetterLocal(w) {
			continue
		}
		if !known(w) {
			lt.unknown[w] = struct{}{}
		}
	}
}

func hasLetterLocal(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func indexRunesFrom(haystack, needle []rune, from int) int {
	if len(needle) == 0 {
		return from
	}
	if from < 0 {
		from = 0
	}
	for i := from; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// CleanOverlappingLocalMatches drops lower-priority matches that overlap higher ones.
// When priorities equal, the earlier (smaller FromPos) match wins.
func CleanOverlappingLocalMatches(matches []LocalMatch) []LocalMatch {
	if len(matches) <= 1 {
		return matches
	}
	sorted := append([]LocalMatch(nil), matches...)
	// process higher priority first, then earlier positions
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		return sorted[i].FromPos < sorted[j].FromPos
	})
	var out []LocalMatch
	for _, m := range sorted {
		overlap := false
		for _, k := range out {
			if spansOverlap(m.FromPos, m.ToPos, k.FromPos, k.ToPos) {
				overlap = true
				break
			}
		}
		if !overlap {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].FromPos < out[j].FromPos })
	return out
}

func spansOverlap(a0, a1, b0, b1 int) bool {
	return a0 < b1 && b0 < a1
}

// SimpleWordRepeatChecker flags consecutive equal word tokens (case-sensitive surface).
// Soft stand-in for GermanWordRepeatRule / WordRepeatRule inject.
func SimpleWordRepeatChecker(ruleID string) SentenceChecker {
	if ruleID == "" {
		ruleID = "WORD_REPEAT_RULE"
	}
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		toks := sentence.GetTokensWithoutWhitespace()
		var out []LocalMatch
		var prevTok *AnalyzedTokenReadings
		for _, tok := range toks {
			if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
				continue
			}
			w := tok.GetToken()
			if w == "" || !hasLetterLocal(w) {
				prevTok = nil
				continue
			}
			if prevTok != nil && prevTok.GetToken() == w {
				out = append(out, LocalMatch{
					FromPos:      prevTok.GetStartPos(),
					ToPos:        tok.GetEndPos(),
					Message:      "Word repetition",
					ShortMessage: "Word repetition",
					RuleID:       ruleID,
				})
			}
			prevTok = tok
		}
		return out
	}
}

// KnownWordSet builds an IsKnownWord from a set of dictionary forms (case-sensitive).
func KnownWordSet(words ...string) func(string) bool {
	m := map[string]struct{}{}
	for _, w := range words {
		m[w] = struct{}{}
	}
	return func(tok string) bool {
		if _, ok := m[tok]; ok {
			return true
		}
		// soft: lowercase probe
		_, ok := m[strings.ToLower(tok)]
		return ok
	}
}

// SimpleMapSpellerChecker flags letter tokens not in known; optional suggestion map.
// When no map entry exists, soft edit-distance suggestions are taken from known
// (capped dictionary size so demo packs stay cheap).
func SimpleMapSpellerChecker(ruleID string, known map[string]struct{}, suggestions map[string][]string) SentenceChecker {
	isKnown := func(w string) bool {
		if _, ok := known[w]; ok {
			return true
		}
		_, ok := known[strings.ToLower(w)]
		return ok
	}
	return SimplePredicateSpellerChecker(ruleID, isKnown, suggestions, known, nil)
}

// SimplePredicateSpellerChecker flags letter tokens rejected by isKnown.
// nearestKnown is optional (edit-distance peers when non-nil and small).
// suggestFn is optional (e.g. CFSA2 edit-candidate Contains); tried after the map.
func SimplePredicateSpellerChecker(ruleID string, isKnown func(string) bool, suggestions map[string][]string, nearestKnown map[string]struct{}, suggestFn func(string) []string) SentenceChecker {
	if ruleID == "" {
		ruleID = "MORFOLOGIK_RULE"
	}
	if isKnown == nil {
		isKnown = func(string) bool { return true }
	}
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		var out []LocalMatch
		for _, tok := range sentence.GetTokensWithoutWhitespace() {
			if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
				continue
			}
			// multiword chunker / disambiguator IGNORE_SPELLING / IMMUNIZE
			if tok.IsIgnoredBySpeller() || tok.IsImmunized() {
				continue
			}
			w := tok.GetToken()
			if w == "" || !hasLetterLocal(w) {
				continue
			}
			if isKnown(w) {
				continue
			}
			m := LocalMatch{
				FromPos:      tok.GetStartPos(),
				ToPos:        tok.GetEndPos(),
				Message:      "Possible spelling mistake",
				ShortMessage: "Spelling mistake",
				RuleID:       ruleID,
			}
			if suggestions != nil {
				if s, ok := suggestions[w]; ok {
					m.Suggestions = append([]string(nil), s...)
				} else if s, ok := suggestions[strings.ToLower(w)]; ok {
					m.Suggestions = append([]string(nil), s...)
				}
			}
			if len(m.Suggestions) == 0 && suggestFn != nil {
				m.Suggestions = suggestFn(w)
			}
			if len(m.Suggestions) == 0 && nearestKnown != nil {
				m.Suggestions = nearestKnownWords(w, nearestKnown, 2, 5)
			}
			out = append(out, m)
		}
		return out
	}
}

// nearestKnownWords returns up to maxN dictionary words within maxDist edit distance.
func nearestKnownWords(word string, known map[string]struct{}, maxDist, maxN int) []string {
	if word == "" || known == nil || len(known) == 0 || len(known) > 10000 || maxN <= 0 {
		return nil
	}
	type cand struct {
		w    string
		dist int
	}
	var cands []cand
	low := strings.ToLower(word)
	seen := map[string]struct{}{}
	for k := range known {
		kl := strings.ToLower(k)
		if _, ok := seen[kl]; ok {
			continue
		}
		d := runeLevenshtein(low, kl)
		if d > 0 && d <= maxDist {
			seen[kl] = struct{}{}
			cands = append(cands, cand{w: kl, dist: d})
		}
	}
	// sort by distance then alphabetically (stable, no import sort for tiny N)
	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			if cands[j].dist < cands[i].dist || (cands[j].dist == cands[i].dist && cands[j].w < cands[i].w) {
				cands[i], cands[j] = cands[j], cands[i]
			}
		}
	}
	if len(cands) > maxN {
		cands = cands[:maxN]
	}
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.w)
	}
	return out
}

func runeLevenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	// band optimization for short words
	if absInt(len(ar)-len(br)) > 4 {
		return absInt(len(ar) - len(br))
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// SimpleAvsAnChecker flags "a" before vowel-sound words and "an" before consonant-sound words.
// Soft stand-in for EN_A_VS_AN with a small exception lexicon (not full phonetics).
func SimpleAvsAnChecker() SentenceChecker {
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		toks := sentence.GetTokensWithoutWhitespace()
		var out []LocalMatch
		for i := 0; i < len(toks)-1; i++ {
			cur, next := toks[i], toks[i+1]
			if cur == nil || next == nil {
				continue
			}
			a := strings.ToLower(cur.GetToken())
			n := next.GetToken()
			if n == "" || !hasLetterLocal(n) {
				continue
			}
			vowel := startsWithVowelSound(n)
			switch a {
			case "a":
				if vowel {
					out = append(out, LocalMatch{
						FromPos:      cur.GetStartPos(),
						ToPos:        cur.GetEndPos(),
						Message:      "Use \"an\" before a vowel sound",
						ShortMessage: "Wrong article",
						RuleID:       "EN_A_VS_AN",
						Suggestions:  []string{"an"},
					})
				}
			case "an":
				if !vowel {
					out = append(out, LocalMatch{
						FromPos:      cur.GetStartPos(),
						ToPos:        cur.GetEndPos(),
						Message:      "Use \"a\" before a consonant sound",
						ShortMessage: "Wrong article",
						RuleID:       "EN_A_VS_AN",
						Suggestions:  []string{"a"},
					})
				}
			}
		}
		return out
	}
}

// startsWithVowelSound reports whether article "an" is preferred before word.
func startsWithVowelSound(word string) bool {
	w := articleWordKey(word)
	if w == "" {
		return false
	}
	// silent-h / vowel sound despite consonant letter → "an"
	if _, ok := anDespiteConsonantLetter[w]; ok {
		return true
	}
	// consonant sound despite vowel letter → "a"
	if _, ok := aDespiteVowelLetter[w]; ok {
		return false
	}
	// prefix families (university, unique, european, one-…)
	for _, p := range aDespiteVowelPrefixes {
		if strings.HasPrefix(w, p) {
			return false
		}
	}
	first, _ := utf8DecodeFirst(w)
	return isVowelLetter(first)
}

// articleWordKey lowercases and keeps the alphabetic prefix (one-time → one).
func articleWordKey(word string) string {
	low := strings.ToLower(strings.TrimSpace(word))
	if low == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range low {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
			continue
		}
		// stop at hyphen/digit/punct after some letters
		if b.Len() > 0 {
			break
		}
	}
	return b.String()
}

// soft EN_A_VS_AN lexicon: silent h / historic vowel sound → take "an"
var anDespiteConsonantLetter = map[string]struct{}{
	"hour": {}, "hours": {}, "hourly": {},
	"honest": {}, "honestly": {}, "honesty": {},
	"honor": {}, "honors": {}, "honour": {}, "honours": {}, "honorable": {}, "honourable": {},
	"heir": {}, "heirs": {}, "heiress": {},
	"herb": {}, "herbs": {}, // US pronunciation
}

// soft: initial /juː/ or /w/ despite orthographic vowel → take "a"
var aDespiteVowelLetter = map[string]struct{}{
	"one": {}, "once": {}, "ones": {},
	"european": {}, "europeans": {},
	"ewe": {}, "ewes": {},
	"u": {}, // "a U-turn" handled by prefix "u" + more letters via prefixes
}

// prefixes for university/unique/euro/user/… families
// (avoid short prefixes like "one" that hit onerous, onerous-like words)
var aDespiteVowelPrefixes = []string{
	"uni",  // university, unique, united, uniform, union…
	"euro", // european, eurozone
	"user", "used", "useful", "usual", "usurp", "usurper",
}

func isVowelLetter(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

func utf8DecodeFirst(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}

// CorrectTextFromLocalMatches applies first suggestion of each match (byte offsets; ASCII-safe).
// Ports Tools.correctTextFromMatches without importing tools package.
func CorrectTextFromLocalMatches(contents string, matches []LocalMatch) string {
	if len(matches) == 0 {
		return contents
	}
	sb := []byte(contents)
	// sort by FromPos ascending so offset adjustments work left-to-right
	sorted := append([]LocalMatch(nil), matches...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].FromPos < sorted[j].FromPos })
	offset := 0
	for _, rm := range sorted {
		if len(rm.Suggestions) == 0 {
			continue
		}
		repl := rm.Suggestions[0]
		from := rm.FromPos - offset
		to := rm.ToPos - offset
		if from < 0 || to < from || to > len(sb) {
			continue
		}
		sb = append(sb[:from], append([]byte(repl), sb[to:]...)...)
		offset += (to - from) - len(repl)
	}
	return string(sb)
}
