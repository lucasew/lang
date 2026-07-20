package en

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	enTok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ConsistentApostrophesRule ports org.languagetool.rules.en.ConsistentApostrophesRule.
type ConsistentApostrophesRule struct {
	Messages map[string]string
	// URL ports Rule.url (Java setUrl punctuation-guide apostrophe).
	URL string
	// DefaultTempOff ports Rule.setDefaultTempOff (Java TODO still off by default).
	DefaultTempOff bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewConsistentApostrophesRule(messages map[string]string) *ConsistentApostrophesRule {
	r := &ConsistentApostrophesRule{
		Messages:       messages,
		URL:            "https://languagetool.org/insights/post/punctuation-guide/#what-is-an-apostrophe",
		DefaultTempOff: true, // Java setDefaultTempOff()
	}
	// Java: addExamplePair(doesn’t → doesn't) with mixed apostrophe styles
	r.AddExamplePair(
		rules.Wrong("It's nice, but it <marker>doesn’t</marker> work."),
		rules.Fixed("It's nice, but it <marker>doesn't</marker> work."),
	)
	return r
}

func (r *ConsistentApostrophesRule) GetID() string { return "EN_CONSISTENT_APOS" }

// GetDescription ports ConsistentApostrophesRule.getDescription.
func (r *ConsistentApostrophesRule) GetDescription() string {
	return "Checks if the two types of apostrophes (' and ’) are used consistently in a text."
}

// GetURL ports Rule.getUrl.
func (r *ConsistentApostrophesRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *ConsistentApostrophesRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// IsDefaultTempOff ports Rule.isDefaultTempOff.
func (r *ConsistentApostrophesRule) IsDefaultTempOff() bool {
	return r != nil && r.DefaultTempOff
}

// MinToCheckParagraph ports TextLevelRule.minToCheckParagraph (Java -1 full text).
func (r *ConsistentApostrophesRule) MinToCheckParagraph() int { return -1 }

// AddExamplePair ports Rule.addExamplePair.
func (r *ConsistentApostrophesRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *ConsistentApostrophesRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *ConsistentApostrophesRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *ConsistentApostrophesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if !hasTwoApostropheTypes(sentences) {
		return nil
	}
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		for _, token := range sentence.GetTokens() {
			if token == nil {
				continue
			}
			// Java: contains("'") + hasTypographicApostrophe flag (not U+2019 surface alone).
			t := token.GetToken()
			var message, repl string
			if strings.Contains(t, "'") && !token.HasTypographicApostrophe() {
				message = "You used a typewriter-style apostrophe here, but a typographic apostrophe elsewhere in this text."
				repl = strings.ReplaceAll(t, "'", "’")
			} else if strings.Contains(t, "'") && token.HasTypographicApostrophe() {
				message = "You used a typographic apostrophe here, but a typewriter-style apostrophe elsewhere in this text."
				// Java: repl = token.getToken() (surface unchanged)
				repl = t
			}
			if message != "" {
				msg := message + " Both are correct, but consider using the same type everywhere in your text."
				rm := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
				rm.SetSuggestedReplacement(repl)
				matches = append(matches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return matches
}

func hasTwoApostropheTypes(sentences []*languagetool.AnalyzedSentence) bool {
	hasTypewriter, hasTypographic := false, false
	for _, sentence := range sentences {
		for _, token := range sentence.GetTokens() {
			if token == nil {
				continue
			}
			t := token.GetToken()
			if strings.Contains(t, "'") && !token.HasTypographicApostrophe() {
				hasTypewriter = true
			} else if strings.Contains(t, "'") && token.HasTypographicApostrophe() {
				hasTypographic = true
			}
			if hasTypewriter && hasTypographic {
				return true
			}
		}
	}
	return false
}

// AnalyzeEnglishPlain analyzes text with EnglishWordTokenizer (contraction splits).
// Java EnglishWordTokenizer.wordsToAdd uses EnglishTagger.INSTANCE so "'s"/"n't"
// stay intact; without a tagger they split on the apostrophe and positions diverge.
func AnalyzeEnglishPlain(text string) *languagetool.AnalyzedSentence {
	ensureEnglishTokenizerTagger()
	wt := enTok.NewEnglishWordTokenizer()
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	readings := make([]*languagetool.AnalyzedTokenReadings, 0, len(raw)+1)
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	prev := ""
	for i, tok := range raw {
		// Java pipeline: typographic ’ may be flagged via hasTypographicApostrophe while
		// surface often uses straight ' after tokenizer normalization paths.
		typoApos := strings.Contains(tok, "’")
		if typoApos {
			tok = strings.ReplaceAll(tok, "’", "'")
		}
		at := languagetool.NewAnalyzedToken(tok, nil, nil)
		ar := languagetool.NewAnalyzedTokenReadingsAt(at, positions[i])
		if typoApos {
			ar.SetTypographicApostrophe(true)
		}
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	return languagetool.NewAnalyzedSentence(readings)
}

var enTokTaggerMu sync.Mutex

// ensureEnglishTokenizerTagger wires EnglishTagger for EnglishWordTokenizer.wordsToAdd
// (Java: EnglishTagger.INSTANCE). Idempotent; no-op if already set or dict missing.
// Re-wires after ClearEnglishFilterTagger / tests that nil IsTaggedEN.
func ensureEnglishTokenizerTagger() {
	if enTok.IsTaggedEN != nil {
		return
	}
	enTokTaggerMu.Lock()
	defer enTokTaggerMu.Unlock()
	if enTok.IsTaggedEN != nil {
		return
	}
	p := discoverEnglishPOSDictForTokenizer()
	if p == "" {
		return
	}
	// Reuse RegisterBinaryEnglishTagger wiring (dict + manuals + IsTaggedEN).
	lt := languagetool.NewJLanguageTool("en")
	_ = RegisterBinaryEnglishTagger(lt, p)
}

func discoverEnglishPOSDictForTokenizer() string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "resource", "en", "english.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 12; i++ {
		for _, rel := range relPaths {
			cand := filepath.Join(dir, rel)
			if st, e := os.Stat(cand); e == nil && st.Mode().IsRegular() {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// AnalyzeEnglishText splits into sentences and analyzes each with EnglishWordTokenizer.
func AnalyzeEnglishText(text string) []*languagetool.AnalyzedSentence {
	// Split like SplitAndAnalyze but use English tokenizer per sentence.
	parts := splitSentencesEN(text)
	if len(parts) == 0 {
		return []*languagetool.AnalyzedSentence{AnalyzeEnglishPlain(text)}
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	offset := 0
	for _, p := range parts {
		if p == "" {
			continue
		}
		s := AnalyzeEnglishPlain(p)
		if offset > 0 {
			for _, t := range s.GetTokens() {
				t.SetStartPos(t.GetStartPos() + offset)
			}
		}
		out = append(out, s)
		for _, r := range p {
			offset += len(utf16.Encode([]rune{r}))
		}
	}
	return out
}

func splitSentencesEN(text string) []string {
	// Reuse languagetool.SplitAndAnalyze structure via simple split
	sents := languagetool.SplitAndAnalyze(text)
	if len(sents) == 0 {
		return []string{text}
	}
	var parts []string
	for _, s := range sents {
		parts = append(parts, s.GetText())
	}
	return parts
}
