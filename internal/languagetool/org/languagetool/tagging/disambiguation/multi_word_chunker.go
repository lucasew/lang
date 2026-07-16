package disambiguation

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// TagForNotAddingTags ports MultiWordChunker.tagForNotAddingTags.
const TagForNotAddingTags = "_NONE_"

const maxTokensInMultiword = 20

// MultiWordChunkerSettings ports MultiWordChunker.Settings fields used at load time.
type MultiWordChunkerSettings struct {
	DefaultTag            string // if set, lines are phrase-only (no tag column)
	AllowAllUppercase     bool
	AllowFirstCapitalized bool
	AllowTitlecase        bool
}

// MultiWordChunker ports org.languagetool.tagging.disambiguation.MultiWordChunker
// (dictionary-driven multiword POS chunker; German line expander deferred).
type MultiWordChunker struct {
	AbstractDisambiguator
	settings MultiWordChunkerSettings

	mu            sync.Mutex
	initialized   bool
	mStartSpace   map[string]int
	mStartNoSpace map[string]int
	mFullSpace    map[string]*languagetool.AnalyzedToken
	mFullNoSpace  map[string]*languagetool.AnalyzedToken

	// Lines is the dictionary source when Filename broker loading is not used.
	// Format: phrase\ttag (or phrase only when DefaultTag is set).
	Lines []string

	AddIgnoreSpelling  bool
	RemovePreviousTags bool
}

func NewMultiWordChunker(lines []string, settings MultiWordChunkerSettings) *MultiWordChunker {
	return &MultiWordChunker{
		settings: settings,
		Lines:    append([]string(nil), lines...),
	}
}

// NewMultiWordChunkerFromReader loads dictionary lines from r.
func NewMultiWordChunkerFromReader(r io.Reader, settings MultiWordChunkerSettings) (*MultiWordChunker, error) {
	lines, err := loadMultiWordLines(r)
	if err != nil {
		return nil, err
	}
	return NewMultiWordChunker(lines, settings), nil
}

func loadMultiWordLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#separatorRegExp=") {
				// separator handled per-file in fillMaps; keep marker for fillMaps
				lines = append(lines, line)
			}
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, sc.Err()
}

// SetIgnoreSpelling ports MultiWordChunker.setIgnoreSpelling.
func (c *MultiWordChunker) SetIgnoreSpelling(v bool) {
	if c != nil {
		c.AddIgnoreSpelling = v
	}
}

func (c *MultiWordChunker) lazyInit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.initialized {
		return
	}
	c.mStartSpace = map[string]int{}
	c.mStartNoSpace = map[string]int{}
	c.mFullSpace = map[string]*languagetool.AnalyzedToken{}
	c.mFullNoSpace = map[string]*languagetool.AnalyzedToken{}
	c.fillMaps()
	c.initialized = true
}

func (c *MultiWordChunker) fillMaps() {
	separator := "\t"
	for _, line := range c.Lines {
		if strings.HasPrefix(line, "#separatorRegExp=") {
			separator = strings.TrimPrefix(line, "#separatorRegExp=")
			continue
		}
		parts := strings.Split(line, separator)
		var original, tag string
		if c.settings.DefaultTag != "" {
			if len(parts) != 1 {
				panic(fmt.Sprintf("Invalid format: '%s', expected one element with no separator", line))
			}
			original = parts[0]
			tag = c.settings.DefaultTag
		} else {
			if len(parts) != 2 {
				panic(fmt.Sprintf("Invalid format: '%s', expected two tab-separated parts", line))
			}
			original = parts[0]
			tag = parts[1]
		}
		containsSpace := strings.Contains(original, " ")
		variants := []string{original}
		if containsSpace {
			variants = append(variants, c.tokenLettercaseVariants(original, c.mFullSpace)...)
		} else {
			variants = append(variants, c.tokenLettercaseVariants(original, c.mFullNoSpace)...)
		}
		for _, casingVariant := range variants {
			if !containsSpace {
				first, _ := utf8.DecodeRuneInString(casingVariant)
				firstChar := string(first)
				if n, ok := c.mStartNoSpace[firstChar]; !ok || n < len(casingVariant) {
					c.mStartNoSpace[firstChar] = len(casingVariant)
				}
				lemma := original
				pos := tag
				c.mFullNoSpace[casingVariant] = languagetool.NewAnalyzedToken(casingVariant, &pos, &lemma)
			} else {
				tokens := strings.Split(casingVariant, " ")
				firstToken := tokens[0]
				if n, ok := c.mStartSpace[firstToken]; !ok || n < len(tokens) {
					c.mStartSpace[firstToken] = len(tokens)
				}
				lemma := original
				pos := tag
				c.mFullSpace[casingVariant] = languagetool.NewAnalyzedToken(casingVariant, &pos, &lemma)
			}
		}
	}
}

func (c *MultiWordChunker) tokenLettercaseVariants(original string, tokenMap map[string]*languagetool.AnalyzedToken) []string {
	var newTokens []string
	if c.settings.AllowAllUppercase && !isCamelCase(original) {
		allUp := strings.ToUpper(original)
		if _, ok := tokenMap[allUp]; !ok && original != allUp {
			newTokens = append(newTokens, allUp)
		}
	}
	if c.settings.AllowFirstCapitalized {
		firstCap := tools.UppercaseFirstChar(original)
		if _, ok := tokenMap[firstCap]; !ok && original != firstCap {
			newTokens = append(newTokens, firstCap)
		}
		if c.settings.AllowTitlecase && strings.Contains(original, " ") && allStartWithLowercase(original) {
			naive := titleCaseWords(original)
			if naive != firstCap && original != naive {
				if _, ok := tokenMap[naive]; !ok {
					newTokens = append(newTokens, naive)
				}
			}
		}
	}
	return newTokens
}

func isCamelCase(s string) bool {
	// crude: lower then upper mid-string (iPad)
	hasLower, hasUpperMid := false, false
	for i, r := range s {
		if unicode.IsLower(r) {
			hasLower = true
		}
		if i > 0 && unicode.IsUpper(r) && hasLower {
			hasUpperMid = true
		}
	}
	return hasLower && hasUpperMid
}

func allStartWithLowercase(s string) bool {
	for _, w := range strings.Split(s, " ") {
		if w == "" {
			continue
		}
		r, _ := utf8.DecodeRuneInString(w)
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func titleCaseWords(s string) string {
	parts := strings.Split(s, " ")
	for i, p := range parts {
		parts[i] = tools.UppercaseFirstChar(p)
	}
	return strings.Join(parts, " ")
}

// Disambiguate ports MultiWordChunker.disambiguate.
func (c *MultiWordChunker) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	c.lazyInit()
	anTokens := input.GetTokens()
	output := make([]*languagetool.AnalyzedTokenReadings, len(anTokens))
	copy(output, anTokens)

	for i := 0; i < len(anTokens); i++ {
		tok := output[i].GetToken()
		if len(tok) < 1 {
			continue
		}
		// concatenate non-whitespace following for no-space lookup key start
		tokBuilder := tok
		k := i + 1
		for k < len(anTokens) && !anTokens[k].IsWhitespace() {
			tokBuilder += output[k].GetToken()
			k++
		}
		tok = tokBuilder

		if _, ok := c.mStartSpace[tok]; ok {
			finalLen := 0
			var keyBuilder strings.Builder
			lenLimit := c.mStartSpace[output[i].GetToken()]
			// prefer first token of original slot
			if n, ok := c.mStartSpace[output[i].GetToken()]; ok {
				lenLimit = n
			}
			j := i
			lenCounter := 0
			for j < len(anTokens) && j-i < maxTokensInMultiword {
				if !anTokens[j].IsWhitespace() {
					keyBuilder.WriteString(anTokens[j].GetToken())
					keyStr := keyBuilder.String()
					if at, ok := c.mFullSpace[keyStr]; ok {
						pos := ""
						if at.GetPOSTag() != nil {
							pos = *at.GetPOSTag()
						}
						if pos != TagForNotAddingTags {
							if finalLen == 0 {
								output[i] = setAndAnnotate(output[i], languagetool.NewAnalyzedToken(anTokens[j].GetToken(), at.GetPOSTag(), at.GetLemma()))
							} else {
								output[i] = prepareNewReading(at, output[i].GetToken(), output[i], false)
								output[finalLen] = prepareNewReading(at, anTokens[finalLen].GetToken(), output[finalLen], true)
							}
						}
						if c.AddIgnoreSpelling {
							if finalLen == 0 {
								output[i].IgnoreSpelling()
							} else {
								for m := i; m <= finalLen && m < len(output); m++ {
									if output[m] != nil {
										output[m].IgnoreSpelling()
									}
								}
							}
						}
					}
				} else {
					if j > 1 && !anTokens[j-1].IsWhitespace() {
						keyBuilder.WriteByte(' ')
						lenCounter++
					}
					if lenCounter == lenLimit {
						break
					}
				}
				j++
				finalLen = j
			}
		}

		r, _ := utf8.DecodeRuneInString(tok)
		first := string(r)
		if _, ok := c.mStartNoSpace[first]; ok {
			j := i
			var keyBuilder strings.Builder
			for j < len(anTokens) && !anTokens[j].IsWhitespace() && j-i < maxTokensInMultiword {
				keyBuilder.WriteString(anTokens[j].GetToken())
				keyStr := keyBuilder.String()
				if at, ok := c.mFullNoSpace[keyStr]; ok {
					pos := ""
					if at.GetPOSTag() != nil {
						pos = *at.GetPOSTag()
					}
					if pos != TagForNotAddingTags {
						if i == j {
							output[i] = setAndAnnotate(output[i], languagetool.NewAnalyzedToken(anTokens[j].GetToken(), at.GetPOSTag(), at.GetLemma()))
						} else {
							output[i] = prepareNewReading(at, anTokens[i].GetToken(), output[i], false)
							output[j] = prepareNewReading(at, anTokens[j].GetToken(), output[j], true)
						}
					}
					if c.AddIgnoreSpelling {
						for m := i; m <= j && m < len(output); m++ {
							if output[m] != nil {
								output[m].IgnoreSpelling()
							}
						}
					}
				}
				j++
			}
		}
	}
	return languagetool.NewAnalyzedSentence(output)
}

func prepareNewReading(at *languagetool.AnalyzedToken, token string, atrs *languagetool.AnalyzedTokenReadings, isLast bool) *languagetool.AnalyzedTokenReadings {
	var b strings.Builder
	b.WriteByte('<')
	if isLast {
		b.WriteByte('/')
	}
	if at.GetPOSTag() != nil {
		b.WriteString(*at.GetPOSTag())
	}
	b.WriteByte('>')
	pos := b.String()
	return setAndAnnotate(atrs, languagetool.NewAnalyzedToken(token, &pos, at.GetLemma()))
}

func setAndAnnotate(oldReading *languagetool.AnalyzedTokenReadings, newReading *languagetool.AnalyzedToken) *languagetool.AnalyzedTokenReadings {
	oldReading.AddReading(newReading, "MULTIWORD_CHUNKER")
	return oldReading
}
