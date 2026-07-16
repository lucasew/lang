package tools

import (
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/pmezard/go-difflib/difflib"
)

// DiffsAsMatches ports org.languagetool.tools.DiffsAsMatches.
// Positions are Java String / UTF-16 code unit offsets.
type DiffsAsMatches struct {
	FilterOutApostropheDiffs bool
	JoinContiguousMatches    bool
}

const (
	maxContiguousDistance = 3
	insertLookbackMax     = 2
)

// java-diff-utils DiffRowGenerator.SPLIT_BY_WORD_PATTERN
var splitByWordPattern = regexp.MustCompile(`\s+|[,.\[\](){}/\\*+\-#]`)

type deltaType int

const (
	deltaEqual deltaType = iota
	deltaInsert
	deltaDelete
	deltaChange
)

type tokenDelta struct {
	typ      deltaType
	srcPos   int // token index in original
	srcLines []string
	tgtLines []string
}

func NewDiffsAsMatches() *DiffsAsMatches {
	return &DiffsAsMatches{
		FilterOutApostropheDiffs: true,
		JoinContiguousMatches:    true,
	}
}

// GetPseudoMatches ports getPseudoMatches.
func (d *DiffsAsMatches) GetPseudoMatches(original, revised string) []*PseudoMatch {
	origTokens := splitStringPreserveDelimiter(original, splitByWordPattern)
	revTokens := splitStringPreserveDelimiter(revised, splitByWordPattern)
	deltas := computeTokenDeltas(origTokens, revTokens)
	matches := d.processDeltasIntoMatches(deltas, origTokens, original)
	if d.FilterOutApostropheDiffs {
		matches = filterOutApostropheDiffs(matches, original)
	}
	if d.JoinContiguousMatches {
		matches = joinContiguousMatches(matches, original)
	}
	return matches
}

func splitStringPreserveDelimiter(str string, re *regexp.Regexp) []string {
	if str == "" {
		return nil
	}
	var list []string
	pos := 0
	for _, loc := range re.FindAllStringIndex(str, -1) {
		if pos < loc[0] {
			list = append(list, str[pos:loc[0]])
		}
		list = append(list, str[loc[0]:loc[1]])
		pos = loc[1]
	}
	if pos < len(str) {
		list = append(list, str[pos:])
	}
	return list
}

func computeTokenDeltas(a, b []string) []tokenDelta {
	m := difflib.NewMatcher(a, b)
	var out []tokenDelta
	for _, op := range m.GetOpCodes() {
		switch op.Tag {
		case 'e':
		case 'i':
			out = append(out, tokenDelta{
				typ: deltaInsert, srcPos: op.I1,
				srcLines: nil, tgtLines: append([]string(nil), b[op.J1:op.J2]...),
			})
		case 'd':
			out = append(out, tokenDelta{
				typ: deltaDelete, srcPos: op.I1,
				srcLines: append([]string(nil), a[op.I1:op.I2]...), tgtLines: nil,
			})
		case 'r':
			out = append(out, tokenDelta{
				typ: deltaChange, srcPos: op.I1,
				srcLines: append([]string(nil), a[op.I1:op.I2]...),
				tgtLines: append([]string(nil), b[op.J1:op.J2]...),
			})
		}
	}
	return out
}

func javaStrLen(s string) int { return len(utf16.Encode([]rune(s))) }

func javaSubstr(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

func (d *DiffsAsMatches) processDeltasIntoMatches(deltas []tokenDelta, originalTokens []string, originalText string) []*PseudoMatch {
	var matches []*PseudoMatch
	var lastMatch *PseudoMatch
	var lastDelta *tokenDelta

	for i := range deltas {
		delta := &deltas[i]
		replacement := strings.Join(delta.tgtLines, "")

		errorIndex := delta.srcPos
		indexCorrection := getIndexCorrectionForInserts(delta, errorIndex)
		fromPos := getPositionFromTokenIndex(originalTokens, errorIndex-indexCorrection)

		wasPreviousTokenWhitespace := isPreviousTokenWhitespace(originalTokens, errorIndex)
		lastPunctuationStr := getPreviousPunctuationIfAny(originalTokens, errorIndex)

		underlinedError := strings.Join(delta.srcLines, "")
		toPos := fromPos + javaStrLen(underlinedError)

		prefixReplacement := buildPrefixForInsert(originalTokens, errorIndex, indexCorrection)
		toPos += javaStrLen(prefixReplacement)
		replacement = prefixReplacement + replacement

		textLen := javaStrLen(originalText)
		if fromPos < 0 {
			fromPos = 0
		}
		if toPos > textLen {
			toPos = textLen
		}
		if fromPos > toPos {
			toPos = fromPos
		}
		underlinedError = javaSubstr(originalText, fromPos, toPos)

		// Remove leading white spaces
		for javaStrLen(underlinedError) > 0 && javaStrLen(replacement) > 0 &&
			IsWhitespace(javaSubstr(underlinedError, 0, 1)) && IsWhitespace(javaSubstr(replacement, 0, 1)) {
			fromPos++
			underlinedError = javaSubstr(underlinedError, 1, javaStrLen(underlinedError))
			replacement = javaSubstr(replacement, 1, javaStrLen(replacement))
		}

		// Special case: INSERT at the sentence start
		if fromPos == 0 && toPos == 0 && len(originalTokens) > 0 {
			toPos = javaStrLen(originalTokens[0])
			replacement = replacement + originalTokens[0]
		}

		// Remove trailing whitespace
		for javaStrLen(underlinedError) > 0 && javaStrLen(replacement) > 0 &&
			IsWhitespace(javaSubstr(underlinedError, javaStrLen(underlinedError)-1, javaStrLen(underlinedError))) &&
			IsWhitespace(javaSubstr(replacement, javaStrLen(replacement)-1, javaStrLen(replacement))) {
			toPos--
			underlinedError = javaSubstr(underlinedError, 0, javaStrLen(underlinedError)-1)
			replacement = javaSubstr(replacement, 0, javaStrLen(replacement)-1)
		}

		var match *PseudoMatch
		if shouldMergeChangeWithInsert(delta, lastDelta, lastMatch, wasPreviousTokenWhitespace, lastPunctuationStr) {
			// Merge CHANGE + INSERT
			suffixStart := toPos - fromPos
			var suffix string
			if suffixStart < javaStrLen(replacement) {
				suffix = javaSubstr(replacement, suffixStart, javaStrLen(replacement))
			}
			newReplacement := lastMatch.GetReplacement() + lastPunctuationStr + suffix
			match = NewPseudoMatch(newReplacement, lastMatch.GetFromPos(), toPos)
			matches = matches[:len(matches)-1]
		} else if shouldMergeWithDelete(delta, lastMatch, fromPos, wasPreviousTokenWhitespace) {
			match = NewPseudoMatch(lastMatch.GetReplacement(), lastMatch.GetFromPos(), toPos-1)
			matches = matches[:len(matches)-1]
		} else {
			match = NewPseudoMatch(replacement, fromPos, toPos)
		}
		matches = append(matches, match)
		lastMatch = match
		lastDelta = delta
	}
	return matches
}

func getIndexCorrectionForInserts(delta *tokenDelta, errorIndex int) int {
	correction := 0
	if delta.typ == deltaInsert {
		correction = insertLookbackMax
		for errorIndex-correction < 0 && correction > 0 {
			correction--
		}
	}
	return correction
}

func isPreviousTokenWhitespace(tokens []string, errorIndex int) bool {
	if errorIndex-1 >= 0 && errorIndex-1 < len(tokens) {
		return IsWhitespace(tokens[errorIndex-1])
	}
	return false
}

func getPreviousPunctuationIfAny(tokens []string, errorIndex int) string {
	if errorIndex-1 >= 0 && errorIndex-1 < len(tokens) {
		token := tokens[errorIndex-1]
		if IsPunctuationMark(token) {
			return token
		}
	}
	return ""
}

func buildPrefixForInsert(tokens []string, errorIndex, indexCorrection int) string {
	var b strings.Builder
	for i := errorIndex - indexCorrection; i < errorIndex; i++ {
		if i >= 0 && i < len(tokens) {
			b.WriteString(tokens[i])
		}
	}
	return b.String()
}

func getPositionFromTokenIndex(tokens []string, tokenIndex int) int {
	position := 0
	for i := 0; i < tokenIndex && i < len(tokens); i++ {
		position += javaStrLen(tokens[i])
	}
	return position
}

func shouldMergeChangeWithInsert(current, previous *tokenDelta, previousMatch *PseudoMatch,
	wasPreviousWhitespace bool, lastPunctuationStr string) bool {
	if previousMatch == nil || previous == nil {
		return false
	}
	if previous.typ != deltaChange || current.typ != deltaInsert {
		return false
	}
	if !wasPreviousWhitespace && lastPunctuationStr == "" {
		return false
	}
	currentPosition := current.srcPos
	previousEndPosition := previous.srcPos + len(previous.srcLines)
	return currentPosition-1 == previousEndPosition
}

func shouldMergeWithDelete(current *tokenDelta, previousMatch *PseudoMatch, currentFromPos int, wasPreviousWhitespace bool) bool {
	if previousMatch == nil {
		return false
	}
	if current.typ != deltaDelete {
		return false
	}
	if !wasPreviousWhitespace {
		return false
	}
	return previousMatch.GetToPos()+1 == currentFromPos
}

func filterOutApostropheDiffs(pseudoMatches []*PseudoMatch, original string) []*PseudoMatch {
	var results []*PseudoMatch
	origLen := javaStrLen(original)
	for _, match := range pseudoMatches {
		if match.GetFromPos() < 0 || match.GetToPos() > origLen || match.GetFromPos() > match.GetToPos() {
			results = append(results, match)
			continue
		}
		originalPart := javaSubstr(original, match.GetFromPos(), match.GetToPos())
		normalizedOriginal := normalizeApostrophes(originalPart)
		normalizedReplacement := normalizeApostrophes(match.GetReplacement())
		if normalizedOriginal != normalizedReplacement {
			results = append(results, match)
		}
	}
	return results
}

func normalizeApostrophes(text string) string {
	return strings.ReplaceAll(text, "’", "'")
}

func joinContiguousMatches(pseudoMatches []*PseudoMatch, original string) []*PseudoMatch {
	var results []*PseudoMatch
	previousEndPosition := -1
	for _, match := range pseudoMatches {
		if previousEndPosition > -1 && match.GetFromPos()-previousEndPosition < maxContiguousDistance {
			joined := joinWithPreviousMatch(match, results[len(results)-1], original)
			results[len(results)-1] = joined
		} else {
			results = append(results, match)
		}
		previousEndPosition = match.GetToPos()
	}
	return results
}

func joinWithPreviousMatch(current, previous *PseudoMatch, original string) *PseudoMatch {
	var b strings.Builder
	b.WriteString(previous.GetReplacement())
	if previous.GetToPos() < current.GetFromPos() {
		b.WriteString(javaSubstr(original, previous.GetToPos(), current.GetFromPos()))
	}
	b.WriteString(current.GetReplacement())
	return NewPseudoMatch(b.String(), previous.GetFromPos(), current.GetToPos())
}
