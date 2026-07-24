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

// SynonymsData ports org.languagetool.rules.SynonymsData.
type SynonymsData struct {
	Synonyms []string
	Postag   string
	Chunk    string
}

// AbstractRepeatedWordsRule ports org.languagetool.rules.AbstractRepeatedWordsRule.
// Matching uses AnalyzedToken lemmas (not surface invent). Suggestions go through
// SynthesizeRE when set; empty result falls back to the synonym lemma (Java).
// Synthesizer/tagger wiring is language-owned; incomplete when SynthesizeRE is nil
// (lemma suggestions only — same as empty synthesize()).
//
// SentenceWithImmunization ports Rule.getSentenceWithImmunization (language anti-patterns).
// Kept as a callback so this package does not import rules/patterns (import cycle).
//
// Java ctor: setCategory(REPETITIONS_STYLE), setLocQualityIssueType(Style).
type AbstractRepeatedWordsRule struct {
	Messages     map[string]string
	ID           string
	Description  string
	Message      string
	ShortMsg     string
	WordsToCheck map[string]*SynonymsData
	MaxDistance  int // default 150
	// LanguageCode is the short code for StringTools.toId (e.g. "de").
	LanguageCode string
	// Category ports Rule.category (Java REPETITIONS_STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// Tags ports Rule.tags (Java EN/ES/CA set Tag.picky).
	Tags []Tag
	// SentenceWithImmunization ports getSentenceWithImmunization. Nil = identity.
	// Language packages wire ANTI_PATTERNS → IMMUNIZE DisambiguationPatternRules here.
	SentenceWithImmunization func(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	// IsException optional; when nil, empty tokens are skipped only.
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, i int, sentStart, isCapitalized, isAllUpper bool) bool
	// AdjustPostag ports adjustPostag; nil keeps postag unchanged.
	AdjustPostag func(postag string) string
	// SynthesizeRE ports getSynthesizer().synthesize(token, postag, true).
	// Nil → treat as empty forms → use synonym lemma as surface (Java empty array path).
	SynthesizeRE func(token *languagetool.AnalyzedToken, posTag string) []string
}

// InitRepeatedWordsMeta applies Java AbstractRepeatedWordsRule constructor metadata.
func InitRepeatedWordsMeta(r *AbstractRepeatedWordsRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatRepetitionsStyle.GetCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSStyle
	}
}

func (r *AbstractRepeatedWordsRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "REPEATEDWORDS"
}

func (r *AbstractRepeatedWordsRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

// GetCategory ports Rule.getCategory.
func (r *AbstractRepeatedWordsRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AbstractRepeatedWordsRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

// GetTags ports Rule.getTags.
func (r *AbstractRepeatedWordsRule) GetTags() []Tag {
	if r == nil {
		return nil
	}
	return r.Tags
}

// HasTag ports Rule.hasTag.
func (r *AbstractRepeatedWordsRule) HasTag(tag Tag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (r *AbstractRepeatedWordsRule) maxDist() int {
	if r.MaxDistance > 0 {
		return r.MaxDistance
	}
	return 150
}

func (r *AbstractRepeatedWordsRule) adjustPostag(postag string) string {
	if r != nil && r.AdjustPostag != nil {
		return r.AdjustPostag(postag)
	}
	return postag
}

// LoadSynonymsWords ports AbstractRepeatedWordsRule.loadWords.
// Keys stay as in the file (Java does not lower-case).
func LoadSynonymsWords(reader io.Reader) (map[string]*SynonymsData, error) {
	hashPattern := regexp.MustCompile(`#.*`)
	out := map[string]*SynonymsData{}
	sc := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := tools.JavaStringTrim(hashPattern.ReplaceAllString(sc.Text(), ""))
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
		for i := range parts {
			parts[i] = tools.JavaStringTrim(parts[i])
		}
		if word != "" {
			var syns []string
			for _, p := range parts {
				if p != "" {
					syns = append(syns, p)
				}
			}
			// Java: map.put(word, …) — no ToLower
			if _, exists := out[word]; !exists {
				out[word] = &SynonymsData{Synonyms: syns, Postag: postag, Chunk: chunk}
			}
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
				if _, exists := out[key]; !exists {
					out[key] = &SynonymsData{Synonyms: values, Postag: postag, Chunk: chunk}
				}
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
		// Java: getSentenceWithImmunization(sentence).getTokensWithoutWhitespace()
		imm := sentence
		if r != nil && r.SentenceWithImmunization != nil {
			imm = r.SentenceWithImmunization(sentence)
		}
		if imm == nil {
			imm = sentence
		}
		tokens := imm.GetTokensWithoutWhitespace()
		pos += prevSentenceLength
		// Java: sentence.getText().length() is UTF-16 code units (String.length).
		prevSentenceLength = utf16Len(sentence.GetText())
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
			// Java: for each AnalyzedToken reading — real lemmas, not surface invent.
			var lemmas []string
			for _, atr := range atrs.GetReadings() {
				if atr == nil {
					continue
				}
				lemPtr := atr.GetLemma()
				if lemPtr == nil || *lemPtr == "" {
					// Java still adds null lemmas; null keys are useless for lookup — skip empty.
					continue
				}
				lemma := *lemPtr
				lemmas = append(lemmas, lemma)
				seen, ok := wordsLastSeen[lemma]
				if !ok || lemmasInSentence[lemma] || (wordNumber-seen) > r.maxDist() {
					continue
				}
				data, ok := r.WordsToCheck[lemma]
				if !ok || data == nil {
					continue
				}
				createMatch := true
				// Java: postag != null && !atr.getPOSTag().matches(postag)
				if data.Postag != "" {
					posPtr := atr.GetPOSTag()
					if posPtr == nil || !javaStringMatches(data.Postag, *posPtr) {
						createMatch = false
					}
				}
				// Java: chunk != null && !atrs.matchesChunkRegex(chunk)
				if data.Chunk != "" && !atrs.MatchesChunkRegex(data.Chunk) {
					createMatch = false
				}
				if !createMatch {
					continue
				}
				msg := r.Message
				if msg == "" {
					msg = "Repeated word"
				}
				rm := NewRuleMatch(r, sentence, pos+atrs.GetStartPos(), pos+atrs.GetEndPos(), msg)
				rm.ShortMessage = r.ShortMsg
				// Java: setSpecificRuleId(ruleId + "_" + StringTools.toId(lemma, language))
				lang := r.LanguageCode
				rm.SetSpecificRuleId(r.GetID() + "_" + tools.ToId(lemma, lang))
				tokenPos := atr.GetPOSTag()
				var tokenPosStr string
				if tokenPos != nil {
					tokenPosStr = *tokenPos
				}
				for _, replacementLemma := range data.Synonyms {
					replacements := r.synthesizeForms(token, tokenPosStr, replacementLemma)
					if len(replacements) == 0 {
						replacements = []string{replacementLemma}
					}
					for _, sug := range replacements {
						if isAllUpper {
							sug = strings.ToUpper(sug)
						} else if isCap {
							sug = tools.UppercaseFirstChar(sug)
						}
						rm.SuggestedReplacements = append(rm.SuggestedReplacements, sug)
					}
				}
				matches = append(matches, rm)
				// Java: break after first matching reading (remaining lemmas not counted)
				break
			}
			// count even if postag/chunk don't match
			for _, lemma := range lemmas {
				if _, ok := r.WordsToCheck[lemma]; ok {
					wordsLastSeen[lemma] = wordNumber
					lemmasInSentence[lemma] = true
				}
			}
		}
	}
	return matches
}

func (r *AbstractRepeatedWordsRule) synthesizeForms(token, posTag, replacementLemma string) []string {
	if r == nil || r.SynthesizeRE == nil {
		return nil
	}
	var posPtr *string
	if posTag != "" {
		p := posTag
		posPtr = &p
	}
	lem := replacementLemma
	tok := languagetool.NewAnalyzedToken(token, posPtr, &lem)
	adjusted := r.adjustPostag(posTag)
	return r.SynthesizeRE(tok, adjusted)
}

// javaStringMatches ports Java String.matches(regex): full-region Pattern match.
func javaStringMatches(regex, s string) bool {
	re, err := regexp.Compile("^(?:" + regex + ")$")
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func isPunctOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) {
			return false
		}
	}
	return s != ""
}
