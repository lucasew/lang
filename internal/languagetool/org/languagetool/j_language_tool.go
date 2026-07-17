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
	RuleID         string
	Suggestions    []string
	// Priority used by CleanOverlappingLocalMatches (higher wins).
	Priority int
}

// SentenceChecker returns matches for one analyzed sentence (offsets relative to sentence text).
type SentenceChecker func(sentence *AnalyzedSentence) []LocalMatch

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
	// Cancelled optional early exit for Check.
	Cancelled CheckCancelledCallback
	// ListUnknownWords enables GetUnknownWords population during Check/AnalyzeUnknown.
	ListUnknownWords bool
	// IsKnownWord optional dictionary probe for unknown-word listing.
	IsKnownWord func(token string) bool
	unknown     map[string]struct{}
}

func NewJLanguageTool(languageCode string) *JLanguageTool {
	return &JLanguageTool{
		LanguageCode: languageCode,
		Mode:         ModeAll,
		Level:        LevelDefault,
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
	for _, p := range parts {
		out = append(out, AnalyzePlain(p))
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
	// Map sentence-local offsets to document by searching each sentence text in remaining source.
	// AnalyzePlain token positions are relative to the sentence string.
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
	return out
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
					FromPos: prevTok.GetStartPos(),
					ToPos:   tok.GetEndPos(),
					Message: "Word repetition",
					RuleID:  ruleID,
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
