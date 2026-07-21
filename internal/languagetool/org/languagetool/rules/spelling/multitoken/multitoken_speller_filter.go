package multitoken

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultitokenSpellerFilter ports
// org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter.
// Speller and misspelled hooks are pluggable (no full Language stack).
type MultitokenSpellerFilter struct {
	Speller *MultitokenSpeller
	// IsMisspelled optional token-level speller (Java SpellingCheckRule.isMisspelled).
	// When set, isMisspelled tokenizes the error with WordTokenizer and ORs token results
	// (Java MultitokenSpellerFilter.isMisspelled). Nil → null SpellingCheckRule path.
	IsMisspelled func(token string) bool
	// Tokenize ports language.getWordTokenizer().tokenize; nil → WordTokenizer.
	Tokenize func(s string) []string
	// AtSentenceStart when true forces capitalization of lower-case suggestions
	// (overrides auto patternTokenPos detection).
	AtSentenceStart bool
	// PatternTokenPos ports patternTokenPos (index in tokensWithoutWhitespace for the
	// first matched token). 0 = auto-detect from match.FromPos when possible.
	PatternTokenPos int
	// CheckSpelling enables Java en/de/pt/nl areTokensAcceptedBySpeller path.
	// When false (default), areTokensAcceptedBySpeller stays false (fr/es/ca/…).
	// When true and IsMisspelled is nil, acceptedBySpeller is true (null speller → !false).
	// When true and IsMisspelled is set, acceptedBySpeller = !isMisspelled(error).
	CheckSpelling bool
}

// AcceptRuleMatch is a convenience surface: originalError when non-empty, else
// match.OriginalErrorStr / sentence span. Pattern tokens optional via AcceptRuleMatchFull.
func (f *MultitokenSpellerFilter) AcceptRuleMatch(match *rules.RuleMatch, originalError string) *rules.RuleMatch {
	return f.AcceptRuleMatchFull(match, nil, f.PatternTokenPos, nil, originalError)
}

// AcceptRuleMatchFull ports RuleFilter.acceptRuleMatch control flow for MultitokenSpellerFilter.
// patternTokenPos is the index in tokensWithoutWhitespace of the first pattern token (Java).
// patternTokens when all isIgnoredBySpeller → drop match (return nil).
func (f *MultitokenSpellerFilter) AcceptRuleMatchFull(
	match *rules.RuleMatch,
	_ map[string]string,
	patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings,
	originalError string,
) *rules.RuleMatch {
	if f == nil || match == nil || f.Speller == nil {
		return match
	}
	// Java: if (Arrays.stream(patternTokens).allMatch(x -> x.isIgnoredBySpeller())) return null;
	if len(patternTokens) > 0 && allIgnoredBySpeller(patternTokens) {
		return nil
	}

	underlinedError := originalError
	if underlinedError == "" {
		underlinedError = match.GetOriginalErrorStr()
	}
	if underlinedError == "" && match.Sentence != nil {
		// UTF-16 span from FromPos/ToPos when OriginalErrorStr unset
		text := match.Sentence.GetText()
		underlinedError = sliceUTF16(text, match.GetFromPos(), match.GetToPos())
	}
	if underlinedError == "" {
		return nil
	}

	// Java: areTokensAcceptedBySpeller for en/de/pt/nl only (CheckSpelling flag)
	acceptedBySpeller := false
	if f.CheckSpelling {
		// null SpellingCheckRule → isMisspelled false → accepted true
		acceptedBySpeller = !f.isMisspelled(underlinedError)
	}
	replacements := f.Speller.GetSuggestionsOpts(underlinedError, acceptedBySpeller)
	if len(replacements) == 0 {
		return nil
	}

	// Java: underlinedError.length()>4 && isAllUppercase — UTF-16 length
	// Java: underlinedError.length()>4 && isAllUppercase — UTF-16 length
	if utf16Len(underlinedError) > 4 && tools.IsAllUppercase(underlinedError) {
		up := make([]string, 0, len(replacements))
		seen := map[string]struct{}{}
		for _, r := range replacements {
			n := strings.ToUpper(r)
			if n == underlinedError {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			up = append(up, n)
		}
		replacements = up
	} else {
		// Capitalize suggestion at sentence start
		atStart := f.AtSentenceStart
		if !atStart && match.Sentence != nil {
			pos := patternTokenPos
			if pos == 0 {
				pos = f.PatternTokenPos
			}
			if pos == 0 {
				pos = patternTokenPosFromMatch(match)
			}
			atStart = isPatternAtSentenceStart(match.Sentence, pos)
		}
		if atStart {
			cap := make([]string, 0, len(replacements))
			seen := map[string]struct{}{}
			for _, r := range replacements {
				n := r
				// do not capitalize iPad (mixed case) — only all-lower
				if r == strings.ToLower(r) {
					n = tools.UppercaseFirstChar(r)
				}
				if n == underlinedError {
					continue
				}
				if _, ok := seen[n]; ok {
					continue
				}
				seen[n] = struct{}{}
				cap = append(cap, n)
			}
			replacements = cap
		}
	}
	if len(replacements) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(replacements)
	return match
}

func allIgnoredBySpeller(tokens []*languagetool.AnalyzedTokenReadings) bool {
	if len(tokens) == 0 {
		return false
	}
	for _, t := range tokens {
		if t == nil || !t.IsIgnoredBySpeller() {
			return false
		}
	}
	return true
}

// isPatternAtSentenceStart ports MultitokenSpellerFilter sentence-start walk:
// wordsStartPos skips leading punct/not-word after SENT_START; true when patternTokenPos == wordsStartPos.
func isPatternAtSentenceStart(sentence *languagetool.AnalyzedSentence, patternTokenPos int) bool {
	if sentence == nil || patternTokenPos <= 0 {
		return false
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	wordsStartPos := 1
	for wordsStartPos < len(tokens) {
		tok := tokens[wordsStartPos]
		if tok == nil {
			wordsStartPos++
			continue
		}
		t := tok.GetToken()
		if tools.IsPunctuationMark(t) || tools.IsNotWordString(t) {
			wordsStartPos++
			continue
		}
		break
	}
	return patternTokenPos == wordsStartPos
}

// patternTokenPosFromMatch finds tokensWithoutWhitespace index of the first token
// whose start equals match.FromPos (UTF-16). 0 if not found.
func patternTokenPosFromMatch(match *rules.RuleMatch) int {
	if match == nil || match.Sentence == nil {
		return 0
	}
	from := match.GetFromPos()
	tokens := match.Sentence.GetTokensWithoutWhitespace()
	for i, tok := range tokens {
		if tok != nil && tok.GetStartPos() == from {
			return i
		}
	}
	return 0
}

// sliceUTF16 returns the substring of s covering UTF-16 units [from,to).
func sliceUTF16(s string, from, to int) string {
	if from < 0 || to <= from {
		return ""
	}
	u := utf16.Encode([]rune(s))
	if from >= len(u) {
		return ""
	}
	if to > len(u) {
		to = len(u)
	}
	return string(utf16.Decode(u[from:to]))
}

// isMisspelled ports MultitokenSpellerFilter.isMisspelled(String, Language):
//
//	if spellerRule == null → false
//	tokens = wordTokenizer.tokenize(s)
//	any token misspelled → true
func (f *MultitokenSpellerFilter) isMisspelled(s string) bool {
	if f == nil || f.IsMisspelled == nil {
		// Java: null SpellingCheckRule → false (not misspelled)
		return false
	}
	tokens := f.tokenizeError(s)
	if len(tokens) == 0 {
		return false
	}
	for _, tok := range tokens {
		if f.IsMisspelled(tok) {
			return true
		}
	}
	return false
}

func (f *MultitokenSpellerFilter) tokenizeError(s string) []string {
	if f != nil && f.Tokenize != nil {
		return f.Tokenize(s)
	}
	return tokenizers.NewWordTokenizer().Tokenize(s)
}
