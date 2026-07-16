package rules

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SynonymsData ports org.languagetool.rules.SynonymsData (postag/chunk ignored on surface port).
type SynonymsData struct {
	Synonyms []string
	Postag   string
	Chunk    string
}

// AbstractRepeatedWordsRule is a surface-level port of AbstractRepeatedWordsRule.
// Tokens are treated as lemmas (no tagger/synthesizer).
type AbstractRepeatedWordsRule struct {
	Messages     map[string]string
	ID           string
	Description  string
	Message      string
	ShortMsg     string
	WordsToCheck map[string]*SynonymsData
	MaxDistance  int // default 150
	// IsException optional; when nil, empty tokens and sentence-start punct are skipped lightly.
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, i int, sentStart, isCapitalized, isAllUpper bool) bool
}

func (r *AbstractRepeatedWordsRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "REPEATEDWORDS"
}

func (r *AbstractRepeatedWordsRule) maxDist() int {
	if r.MaxDistance > 0 {
		return r.MaxDistance
	}
	return 150
}

// LoadSynonymsWords ports AbstractRepeatedWordsRule.loadWords.
func LoadSynonymsWords(reader io.Reader) (map[string]*SynonymsData, error) {
	hashPattern := regexp.MustCompile(`#.*`)
	out := map[string]*SynonymsData{}
	sc := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(hashPattern.ReplaceAllString(sc.Text(), ""))
		if line == "" {
			continue
		}
		mainParts := strings.SplitN(line, "=", 2)
		var parts []string
		var postag, chunk, word string
		if len(mainParts) == 2 {
			parts = strings.Split(mainParts[1], ";")
			wordPos := strings.Split(mainParts[0], "/")
			word = wordPos[0]
			if len(wordPos) > 1 {
				postag = wordPos[1]
			}
			if len(wordPos) > 2 {
				chunk = wordPos[2]
			}
		} else if len(mainParts) == 1 {
			parts = strings.Split(line, ";")
			word = ""
		} else {
			continue
		}
		// trim parts
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if word != "" {
			// drop empty synonyms
			var syns []string
			for _, p := range parts {
				if p != "" {
					syns = append(syns, p)
				}
			}
			out[strings.ToLower(word)] = &SynonymsData{Synonyms: syns, Postag: postag, Chunk: chunk}
		} else {
			for _, key := range parts {
				if key == "" {
					continue
				}
				var values []string
				for _, v := range parts {
					if v != "" && v != key {
						values = append(values, v)
					}
				}
				out[strings.ToLower(key)] = &SynonymsData{Synonyms: values, Postag: postag, Chunk: chunk}
			}
		}
	}
	return out, sc.Err()
}

func isCapitalizedWord(s string) bool {
	rs := []rune(s)
	if len(rs) < 2 {
		return false
	}
	return unicode.IsUpper(rs[0]) && unicode.IsLower(rs[1])
}

// MatchList ports AbstractRepeatedWordsRule.match.
func (r *AbstractRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var matches []*RuleMatch
	wordNumber := 0
	wordsLastSeen := map[string]int{}
	pos := 0
	prevSentenceLength := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		pos += prevSentenceLength
		prevSentenceLength = len([]byte(sentence.GetText())) // UTF-8; Java uses char length — OK for BMP tests
		if len(tokens) == 0 {
			continue
		}
		lastToken := tokens[len(tokens)-1].GetToken()
		if lastToken != "." && lastToken != "!" && lastToken != "?" {
			continue
		}
		sentStart := true
		lemmasInSentence := map[string]bool{}
		for i, atrs := range tokens {
			if atrs.IsImmunized() {
				continue
			}
			token := atrs.GetToken()
			if token != "" {
				wordNumber++
			}
			isCap := isCapitalizedWord(token)
			isAllUpper := tools.IsAllUppercase(token)
			isException := token == ""
			if r.IsException != nil {
				isException = isException || r.IsException(tokens, i, sentStart, isCap, isAllUpper)
			}
			if sentStart && token != "" && !isPunctOnly(token) {
				sentStart = false
			}
			if isException {
				continue
			}
			// surface lemma = lowercased token
			lemma := strings.ToLower(token)
			if seen, ok := wordsLastSeen[lemma]; ok && !lemmasInSentence[lemma] &&
				(wordNumber-seen) <= r.maxDist() {
				if data, ok := r.WordsToCheck[lemma]; ok {
					msg := r.Message
					if msg == "" {
						msg = "Repeated word"
					}
					rm := NewRuleMatch(r, sentence, pos+atrs.GetStartPos(), pos+atrs.GetEndPos(), msg)
					rm.ShortMessage = r.ShortMsg
					for _, rep := range data.Synonyms {
						sugg := rep
						if isAllUpper {
							sugg = strings.ToUpper(rep)
						} else if isCap {
							sugg = tools.UppercaseFirstChar(rep)
						}
						rm.SuggestedReplacements = append(rm.SuggestedReplacements, sugg)
					}
					matches = append(matches, rm)
				}
			}
			if _, ok := r.WordsToCheck[lemma]; ok {
				wordsLastSeen[lemma] = wordNumber
				lemmasInSentence[lemma] = true
			}
		}
	}
	return matches
}

func isPunctOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) {
			return false
		}
	}
	return s != ""
}
