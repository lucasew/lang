package languagetool

import (
	"strings"
)

// AnalyzedSentence ports org.languagetool.AnalyzedSentence.
type AnalyzedSentence struct {
	tokens                    []*AnalyzedTokenReadings
	preDisambigTokens         []*AnalyzedTokenReadings
	nonBlankTokens            []*AnalyzedTokenReadings
	nonBlankPreDisambigTokens []*AnalyzedTokenReadings
	whPositions               []int
	// tokenOffsets / lemmaOffsets: lowercase → indices in nonBlankTokens (Java maps).
	tokenOffsets map[string][]int
	lemmaOffsets map[string][]int
	// text caches getText() (Java volatile String text).
	text       string
	textCached bool
}

func NewAnalyzedSentence(words []*AnalyzedTokenReadings) *AnalyzedSentence {
	return NewAnalyzedSentenceFull(words, words)
}

func NewAnalyzedSentenceFull(tokens, preDisambig []*AnalyzedTokenReadings) *AnalyzedSentence {
	// Java primitives are pass-by-value: each getNonBlankReadings starts counters at 0;
	// the shared mapping array is overwritten by the second call.
	mapping := make([]int, len(tokens)+1)
	nonBlank := getNonBlankReadings(tokens, mapping)
	s := &AnalyzedSentence{
		tokens:            tokens,
		preDisambigTokens: preDisambig,
		whPositions:       mapping,
		nonBlankTokens:    nonBlank,
	}
	s.nonBlankPreDisambigTokens = getNonBlankReadings(preDisambig, mapping)
	s.tokenOffsets = indexTokens(nonBlank)
	s.lemmaOffsets = indexLemmas(nonBlank)
	return s
}

// newAnalyzedSentencePrivate ports the package-private copy constructor.
func newAnalyzedSentencePrivate(tokens []*AnalyzedTokenReadings, mapping []int, nonBlank, nonBlankPre []*AnalyzedTokenReadings) *AnalyzedSentence {
	s := &AnalyzedSentence{
		tokens:                    tokens,
		preDisambigTokens:         tokens, // Java sets preDisambigTokens = tokens
		whPositions:               mapping,
		nonBlankTokens:            nonBlank,
		nonBlankPreDisambigTokens: nonBlankPre,
	}
	s.tokenOffsets = indexTokens(nonBlank)
	s.lemmaOffsets = indexLemmas(nonBlank)
	return s
}

func indexTokens(tokens []*AnalyzedTokenReadings) map[string][]int {
	result := make(map[string][]int, len(tokens))
	for i, t := range tokens {
		if t == nil {
			continue
		}
		key := strings.ToLower(t.GetToken())
		result[key] = append(result[key], i)
	}
	return result
}

func indexLemmas(tokens []*AnalyzedTokenReadings) map[string][]int {
	result := make(map[string][]int, len(tokens))
	for i, tr := range tokens {
		if tr == nil {
			continue
		}
		for j := 0; j < tr.GetReadingsLength(); j++ {
			tok := tr.GetAnalyzedToken(j)
			key := tok.GetToken()
			if lem := tok.GetLemma(); lem != nil {
				key = *lem
			}
			key = strings.ToLower(key)
			list := result[key]
			if len(list) == 0 || list[len(list)-1] != i {
				result[key] = append(list, i)
			}
		}
	}
	return result
}

func getNonBlankReadings(tokens []*AnalyzedTokenReadings, mapping []int) []*AnalyzedTokenReadings {
	var l []*AnalyzedTokenReadings
	whCounter, nonWhCounter := 0, 0
	for _, token := range tokens {
		if !token.IsWhitespace() || token.IsSentenceStart() || token.IsSentenceEnd() || token.IsParagraphEnd() {
			l = append(l, token)
			if nonWhCounter < len(mapping) {
				mapping[nonWhCounter] = whCounter
			}
			nonWhCounter++
		}
		whCounter++
	}
	return l
}

func (s *AnalyzedSentence) GetTokens() []*AnalyzedTokenReadings { return s.tokens }

// GetPreDisambigTokens ports getPreDisambigTokens.
func (s *AnalyzedSentence) GetPreDisambigTokens() []*AnalyzedTokenReadings {
	if s == nil {
		return nil
	}
	return s.preDisambigTokens
}

func (s *AnalyzedSentence) GetTokensWithoutWhitespace() []*AnalyzedTokenReadings {
	return s.nonBlankTokens
}

// GetNonWhitespaceTokenCount ports getNonWhitespaceTokenCount.
func (s *AnalyzedSentence) GetNonWhitespaceTokenCount() int {
	if s == nil {
		return 0
	}
	return len(s.nonBlankTokens)
}

// GetOriginalPosition ports getOriginalPosition(nonWhPosition) via whPositions mapping.
func (s *AnalyzedSentence) GetOriginalPosition(nonWhPosition int) int {
	if s == nil || nonWhPosition < 0 || nonWhPosition >= len(s.whPositions) {
		return -1
	}
	return s.whPositions[nonWhPosition]
}

func (s *AnalyzedSentence) GetPreDisambigTokensWithoutWhitespace() []*AnalyzedTokenReadings {
	return s.nonBlankPreDisambigTokens
}

// cloneAnalyzedTokenSlice deep-copies readings for pre-disambiguation snapshots
// (Java keeps pre-disambig tokens separate from disambiguated tokens).
func cloneAnalyzedTokenSlice(in []*AnalyzedTokenReadings) []*AnalyzedTokenReadings {
	if len(in) == 0 {
		return nil
	}
	out := make([]*AnalyzedTokenReadings, len(in))
	for i, t := range in {
		if t == nil {
			continue
		}
		// Copy readings slice so disambiguator mutations do not alias pre-disambig.
		rds := append([]*AnalyzedToken(nil), t.GetReadings()...)
		out[i] = NewAnalyzedTokenReadingsFromOld(t, rds, "")
	}
	return out
}

// Copy ports AnalyzedSentence.copy.
// Token array is deep-copied; nonBlank slices keep the Java references from the source
// (private ctor reuses sentence.getTokensWithoutWhitespace() arrays).
func (s *AnalyzedSentence) Copy(sentence *AnalyzedSentence) *AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	origTokens := sentence.GetTokens()
	copyTokens := make([]*AnalyzedTokenReadings, len(origTokens))
	for i, analyzedTokens := range origTokens {
		if analyzedTokens == nil {
			continue
		}
		copyTokens[i] = NewAnalyzedTokenReadingsFromOld(analyzedTokens, analyzedTokens.GetReadings(), "")
	}
	return newAnalyzedSentencePrivate(
		copyTokens,
		append([]int(nil), sentence.whPositions...),
		sentence.GetTokensWithoutWhitespace(),
		sentence.GetPreDisambigTokensWithoutWhitespace(),
	)
}

func (s *AnalyzedSentence) String() string {
	return s.ToStringDelim(",")
}

// ToStringDelim ports toString(readingDelimiter) with chunk tags included.
func (s *AnalyzedSentence) ToStringDelim(readingDelimiter string) string {
	return s.toStringJava(readingDelimiter, true)
}

// toStringJava ports private toString(readingDelimiter, includeChunks).
func (s *AnalyzedSentence) toStringJava(readingDelimiter string, includeChunks bool) string {
	if s == nil {
		return ""
	}
	var sb strings.Builder
	for _, element := range s.tokens {
		if !element.IsWhitespace() {
			sb.WriteString(element.GetToken())
			sb.WriteByte('[')
		}
		readings := element.GetReadings()
		for i := 0; i < len(readings); i++ {
			token := readings[i]
			posTag := token.GetPOSTag()
			if element.IsSentenceStart() {
				sb.WriteString("<S>")
			} else if posTag != nil && *posTag == SentenceEndTagName {
				sb.WriteString("</S>")
			} else if posTag != nil && *posTag == ParagraphEndTagName {
				sb.WriteString("<P/>")
			} else if posTag == nil && !includeChunks {
				sb.WriteString(token.GetToken())
			} else {
				if !element.IsWhitespace() {
					sb.WriteString(token.String())
					if i+1 < len(readings) {
						sb.WriteString(readingDelimiter)
					}
				}
			}
		}
		if !element.IsWhitespace() {
			// Java: if (includeChunks && element.getChunkTags().size() > 0)
			if includeChunks {
				if tags := element.GetChunkTags(); len(tags) > 0 {
					sb.WriteByte(',')
					sb.WriteString(strings.Join(tags, "|"))
				}
			}
			if element.IsImmunized() {
				sb.WriteString("{!}")
			}
			sb.WriteByte(']')
		} else {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

// Equals ports AnalyzedSentence.equals:
// Arrays.equals(nonBlankTokens) && Arrays.equals(tokens) && Arrays.equals(whPositions).
func (s *AnalyzedSentence) Equals(o *AnalyzedSentence) bool {
	if s == o {
		return true
	}
	if s == nil || o == nil {
		return false
	}
	if !equalIntSlice(s.whPositions, o.whPositions) {
		return false
	}
	return equalATRSlice(s.nonBlankTokens, o.nonBlankTokens) && equalATRSlice(s.tokens, o.tokens)
}

func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalATRSlice(a, b []*AnalyzedTokenReadings) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] == b[i] {
			continue
		}
		if a[i] == nil || b[i] == nil || !a[i].Equals(b[i]) {
			return false
		}
	}
	return true
}

// HashCode ports AnalyzedSentence.hashCode (Objects.hash(nonBlankTokens, tokens, whPositions)).
func (s *AnalyzedSentence) HashCode() int {
	if s == nil {
		return 0
	}
	h := 1
	h = 31*h + atrArrayHash(s.nonBlankTokens)
	h = 31*h + atrArrayHash(s.tokens)
	h = 31*h + intArrayHash(s.whPositions)
	return h
}

func atrArrayHash(a []*AnalyzedTokenReadings) int {
	// Arrays.hashCode
	h := 1
	for _, t := range a {
		th := 0
		if t != nil {
			th = t.HashCode()
		}
		h = 31*h + th
	}
	return h
}

func intArrayHash(a []int) int {
	h := 1
	for _, v := range a {
		h = 31*h + v
	}
	return h
}

// GetCorrectedTextLength ports AnalyzedSentence.getCorrectedTextLength:
// sum of getCleanToken().length() (UTF-16) + getPosFix() only on the last token.
func (s *AnalyzedSentence) GetCorrectedTextLength() int {
	if s == nil {
		return 0
	}
	lenSum := 0
	nTok := len(s.tokens)
	for i, element := range s.tokens {
		if element == nil {
			continue
		}
		t := element.GetCleanToken()
		n := 0
		for _, r := range t {
			if r >= 0x10000 {
				n += 2
			} else {
				n++
			}
		}
		lenSum += n
		// Java: only apply posFix at end so per-token fixes do not accumulate
		if i == nTok-1 {
			lenSum += element.GetPosFix()
		}
	}
	return lenSum
}

// GetText ports AnalyzedSentence.getText — original text by concatenating tokens.
func (s *AnalyzedSentence) GetText() string {
	if s == nil {
		return ""
	}
	if s.textCached {
		return s.text
	}
	var b strings.Builder
	for _, element := range s.tokens {
		if element != nil {
			b.WriteString(element.GetToken())
		}
	}
	s.text = b.String()
	s.textCached = true
	return s.text
}

// GetTokenSet ports getTokenSet — keySet of tokenOffsets (lowercased non-blank tokens).
func (s *AnalyzedSentence) GetTokenSet() map[string]struct{} {
	out := map[string]struct{}{}
	if s == nil {
		return out
	}
	for k := range s.tokenOffsets {
		out[k] = struct{}{}
	}
	return out
}

// GetLemmaSet ports getLemmaSet — keySet of lemmaOffsets
// (lowercase lemma, or surface token when lemma is null).
func (s *AnalyzedSentence) GetLemmaSet() map[string]struct{} {
	out := map[string]struct{}{}
	if s == nil {
		return out
	}
	for k := range s.lemmaOffsets {
		out[k] = struct{}{}
	}
	return out
}

// GetTokenOffsets ports getTokenOffsets — non-blank indices for a lowercased token, or nil.
func (s *AnalyzedSentence) GetTokenOffsets(token string) []int {
	if s == nil {
		return nil
	}
	return s.tokenOffsets[strings.ToLower(token)]
}

// GetLemmaOffsets ports getLemmaOffsets — non-blank indices for a lowercased lemma key, or nil.
func (s *AnalyzedSentence) GetLemmaOffsets(lemma string) []int {
	if s == nil {
		return nil
	}
	return s.lemmaOffsets[strings.ToLower(lemma)]
}

// ToShortString ports toShortString(readingDelimiter) — includeChunks=false.
func (s *AnalyzedSentence) ToShortString(readingDelimiter string) string {
	return s.toStringJava(readingDelimiter, false)
}

// GetAnnotations ports getAnnotations — disambiguator actions log.
func (s *AnalyzedSentence) GetAnnotations() string {
	if s == nil {
		return "Disambiguator log: \n"
	}
	var b strings.Builder
	b.WriteString("Disambiguator log: \n")
	for _, element := range s.tokens {
		if element == nil || element.IsWhitespace() {
			continue
		}
		if a := element.GetHistoricalAnnotations(); a != "" {
			b.WriteString(a)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
