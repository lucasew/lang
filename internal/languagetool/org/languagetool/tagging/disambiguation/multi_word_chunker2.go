package disambiguation

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultiWordChunker2 ports org.languagetool.tagging.disambiguation.MultiWordChunker2.
// First matching multiword entry wins; no overlapping.
type MultiWordChunker2 struct {
	AbstractDisambiguator
	AllowFirstCapitalized bool
	RemoveOtherReadings   bool
	WrapTag               bool

	mu             sync.Mutex
	tokenToEntries map[string][]multiWordEntry
	lines          []string
	initialized    bool
}

type multiWordEntry struct {
	tokens []string
	tag    string
}

func (e multiWordEntry) lemma() string { return strings.Join(e.tokens, " ") }

// NewMultiWordChunker2 builds from phrase\ttag lines.
func NewMultiWordChunker2(lines []string, allowFirstCapitalized bool) *MultiWordChunker2 {
	return &MultiWordChunker2{
		AllowFirstCapitalized: allowFirstCapitalized,
		WrapTag:               true,
		lines:                 append([]string(nil), lines...),
	}
}

func (c *MultiWordChunker2) SetRemoveOtherReadings(v bool) { c.RemoveOtherReadings = v }
func (c *MultiWordChunker2) SetWrapTag(v bool)             { c.WrapTag = v }

func (c *MultiWordChunker2) lazyInit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.initialized {
		return
	}
	m := map[string][]multiWordEntry{}
	for _, line := range c.lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			panic(fmt.Sprintf("Invalid multiword format: '%s', expected two tab-separated parts", line))
		}
		tokens := strings.Split(parts[0], " ")
		entry := multiWordEntry{tokens: tokens, tag: parts[1]}
		m[tokens[0]] = append(m[tokens[0]], entry)
	}
	c.tokenToEntries = m
	c.initialized = true
}

func (c *MultiWordChunker2) formatPosTag(posTag string, position, multiwordLength int) string {
	if c.WrapTag {
		return "<" + posTag + ">"
	}
	return posTag
}

// Disambiguate ports MultiWordChunker2.disambiguate.
func (c *MultiWordChunker2) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	c.lazyInit()
	inputTokens := input.GetTokens()
	output := make([]*languagetool.AnalyzedTokenReadings, len(inputTokens))
	copy(output, inputTokens)

	for i := 1; i < len(inputTokens); i++ {
		firstToken := inputTokens[i].GetToken()
		items := c.tokenToEntries[firstToken]
		if items == nil && c.AllowFirstCapitalized && tools.IsCapitalizedWord(firstToken) {
			items = c.tokenToEntries[tools.LowercaseFirstChar(firstToken)]
		}
		if items == nil {
			continue
		}
		entry, ok := c.findMultiwordEntry(inputTokens, i, items)
		if !ok {
			continue
		}
		multiwordPos := 0
		for inputTokenPos := i; multiwordPos < len(entry.tokens) && inputTokenPos < len(inputTokens); inputTokenPos++ {
			current := inputTokens[inputTokenPos]
			if current.IsWhitespace() {
				continue
			}
			tag := c.formatPosTag(entry.tag, multiwordPos, len(entry.tokens))
			output[inputTokenPos] = c.prepareNewReading(entry.lemma(), current.GetToken(), current, tag)
			multiwordPos++
		}
	}
	return languagetool.NewAnalyzedSentence(output)
}

func (c *MultiWordChunker2) findMultiwordEntry(inputTokens []*languagetool.AnalyzedTokenReadings, starting int, items []multiWordEntry) (multiWordEntry, bool) {
	for _, e := range items {
		if c.isMatching(inputTokens, starting, e) {
			return e, true
		}
	}
	return multiWordEntry{}, false
}

func (c *MultiWordChunker2) isMatching(inputTokens []*languagetool.AnalyzedTokenReadings, starting int, e multiWordEntry) bool {
	j := 1 // first token already matched
	for i := 1; j < len(e.tokens) && starting+i < len(inputTokens); i++ {
		if inputTokens[starting+i].IsWhitespace() {
			continue
		}
		if !c.matches(e.tokens[j], inputTokens[starting+i]) {
			return false
		}
		j++
	}
	return j == len(e.tokens)
}

func (c *MultiWordChunker2) matches(matchText string, tok *languagetool.AnalyzedTokenReadings) bool {
	return matchText == tok.GetToken()
}

func (c *MultiWordChunker2) prepareNewReading(lemma, tok string, token *languagetool.AnalyzedTokenReadings, tag string) *languagetool.AnalyzedTokenReadings {
	pos := tag
	newTok := languagetool.NewAnalyzedToken(tok, &pos, &lemma)
	if c.RemoveOtherReadings {
		return languagetool.NewAnalyzedTokenReadings(newTok)
	}
	token.AddReading(newTok, "MULTIWORD_CHUNKER")
	return token
}
