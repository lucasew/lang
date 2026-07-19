package patterns

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// matchMarker is Java XMLRuleHandler's SOH prefix for real <match> elements (\u0001\N).
const matchMarker = "\x01"

// reMatchTag ports PatternRuleHandler <match …/> / <match …></match> in messages.
// Group 1 = attributes, group 2 = optional body (lemma string).
var reMatchTag = regexp.MustCompile(`(?is)<match\b([^>]*)(?:/>|>(.*?)</match>)`)

// ProcessRuleMessage ports message-side Match handling:
//  1. inject <pleasespellme/> into suppress_misspelled suggestions (PatternRuleHandler)
//  2. rewrite <match no="N" …/> → \u0001\N + Match list (XMLRuleHandler.setMatchElement)
//  3. addLegacyMatches for bare \N (inMessageOnly)
//  4. strip SOH markers left in the string
// Returns cleaned message text and ordered SuggestionMatches for formatMatches.
func ProcessRuleMessage(raw string) (string, []*Match) {
	if raw == "" {
		return "", nil
	}
	msg := injectPleaseSpellMe(raw)
	msg, fromTags := rewriteMatchTags(msg)
	// addLegacyMatches: one Match per \digits occurrence, using tag Matches for SOH-prefixed refs
	combined := addLegacyMatches(fromTags, msg)
	// strip remaining SOH
	msg = strings.ReplaceAll(msg, matchMarker, "")
	return msg, combined
}

// reSuggestionOpen matches <suggestion …> including suppress_misspelled.
var reSuggestionOpen = regexp.MustCompile(`(?is)<suggestion(\s[^>]*)?>`)

// injectPleaseSpellMe ports PatternRuleHandler suggestion start with suppress_misspelled="yes".
func injectPleaseSpellMe(msg string) string {
	return reSuggestionOpen.ReplaceAllStringFunc(msg, func(open string) string {
		attrs := parseXMLAttrs(open)
		if strings.EqualFold(attrs["suppress_misspelled"], "yes") {
			// <suggestion…><pleasespellme/>
			return open + PleaseSpellMe
		}
		return open
	})
}

// rewriteMatchTags replaces <match …/> with \u0001\N and builds Match configs.
func rewriteMatchTags(msg string) (string, []*Match) {
	var matches []*Match
	var b strings.Builder
	last := 0
	for _, loc := range reMatchTag.FindAllStringSubmatchIndex(msg, -1) {
		// loc: full start/end, attrs, body
		if len(loc) < 4 {
			continue
		}
		fullStart, fullEnd := loc[0], loc[1]
		attrsStr := msg[loc[2]:loc[3]]
		body := ""
		if len(loc) >= 6 && loc[4] >= 0 {
			body = msg[loc[4]:loc[5]]
		}
		attrs := parseXMLAttrs(attrsStr)
		no := strings.TrimSpace(attrs["no"])
		if no == "" {
			continue // leave unparseable tag as-is by not consuming
		}
		m := matchFromAttrs(attrs)
		if n, err := strconv.Atoi(no); err == nil {
			m.SetTokenRef(n)
		}
		if body = strings.TrimSpace(body); body != "" {
			m.SetLemmaString(body)
		}
		inSug := isInsideSuggestion(msg, fullStart)
		m.SetInMessageOnly(!inSug)
		// Inherit suppress_misspelled from enclosing <suggestion> (Java setMatchElement).
		if inSug && suggestionSuppressMisspelled(msg, fullStart) {
			m.SuppressMisspelled = true
		}
		matches = append(matches, m)
		b.WriteString(msg[last:fullStart])
		b.WriteString(matchMarker)
		b.WriteByte('\\')
		b.WriteString(no)
		last = fullEnd
	}
	b.WriteString(msg[last:])
	return b.String(), matches
}

// suggestionSuppressMisspelled reports whether the enclosing <suggestion> has suppress_misspelled="yes".
func suggestionSuppressMisspelled(msg string, at int) bool {
	if at < 0 || at > len(msg) {
		return false
	}
	lower := strings.ToLower(msg[:at])
	open := strings.LastIndex(lower, "<suggestion")
	if open < 0 {
		return false
	}
	end := strings.Index(msg[open:], ">")
	if end < 0 {
		return false
	}
	tag := msg[open : open+end+1]
	attrs := parseXMLAttrs(tag)
	return strings.EqualFold(attrs["suppress_misspelled"], "yes")
}

func isInsideSuggestion(msg string, at int) bool {
	if at < 0 {
		return false
	}
	// last <suggestion> before at vs last </suggestion>
	open := strings.LastIndex(strings.ToLower(msg[:at]), "<suggestion")
	close := strings.LastIndex(strings.ToLower(msg[:at]), "</suggestion>")
	return open > close
}

func matchFromAttrs(attrs map[string]string) *Match {
	postag := attrs["postag"]
	postagReplace := attrs["postag_replace"]
	postagRE := strings.EqualFold(attrs["postag_regexp"], "yes")
	regexMatch := attrs["regexp_match"]
	regexReplace := attrs["regexp_replace"]
	caseConv := CaseNone
	if v := attrs["case_conversion"]; v != "" {
		caseConv = CaseConversion(strings.ToUpper(v))
	}
	include := IncludeNone
	if v := attrs["include_skipped"]; v != "" {
		include = IncludeRange(strings.ToUpper(v))
	}
	setPos := strings.EqualFold(attrs["setpos"], "yes")
	suppress := strings.EqualFold(attrs["suppress_misspelled"], "yes")
	return NewMatch(postag, postagReplace, postagRE, regexMatch, regexReplace, caseConv, setPos, suppress, include)
}

// parseXMLAttrs pulls attr="value" pairs from a start tag fragment.
func parseXMLAttrs(tag string) map[string]string {
	out := map[string]string{}
	// name="value" or name='value'
	re := regexp.MustCompile(`([A-Za-z_][\w-]*)\s*=\s*("([^"]*)"|'([^']*)')`)
	for _, m := range re.FindAllStringSubmatch(tag, -1) {
		name := strings.ToLower(m[1])
		val := m[3]
		if val == "" {
			val = m[4]
		}
		out[name] = val
	}
	return out
}

// addLegacyMatches ports XMLRuleHandler.addLegacyMatches.
// existing are Matches from real <match> tags in document order of SOH markers.
func addLegacyMatches(existing []*Match, messageStr string) []*Match {
	var sugMatch []*Match
	matchCounter := 0
	for i := 0; i < len(messageStr); i++ {
		if messageStr[i] != '\\' || i+1 >= len(messageStr) {
			continue
		}
		if !unicode.IsDigit(rune(messageStr[i+1])) {
			continue
		}
		// preceded by SOH → real <match>
		if i > 0 && messageStr[i-1] == matchMarker[0] {
			if matchCounter < len(existing) {
				sugMatch = append(sugMatch, existing[matchCounter])
				matchCounter++
			} else {
				// incomplete pairing — fall back to bare ref Match
				mw := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
				mw.SetInMessageOnly(true)
				sugMatch = append(sugMatch, mw)
			}
		} else {
			mw := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
			mw.SetInMessageOnly(true)
			sugMatch = append(sugMatch, mw)
		}
	}
	if len(sugMatch) == 0 {
		return existing
	}
	return sugMatch
}

// FormatMatches ports PatternRuleMatcher.formatMatches.
// positions[i] = tokens consumed by pattern element i (0 = optional absent).
// suggestionMatches ordered per backref occurrence (addLegacyMatches).
// Uses LanguageSynthesizer(langCode) when registered; otherwise surface path.
func FormatMatches(
	tokenReadings []*languagetool.AnalyzedTokenReadings,
	positions []int,
	firstMatchTok int,
	errorMsg string,
	suggestionMatches []*Match,
	langCode string,
) string {
	if errorMsg == "" || !strings.Contains(errorMsg, `\`) {
		return errorMsg
	}
	errorMessage := errorMsg
	errorMessageProcessed := 0
	matchCounter := 0

	for {
		backslashPos := -1
		for i := errorMessageProcessed; i < len(errorMessage); i++ {
			if errorMessage[i] == '\\' && i+1 < len(errorMessage) && unicode.IsDigit(rune(errorMessage[i+1])) {
				backslashPos = i
				break
			}
		}
		if backslashPos < 0 {
			break
		}
		numLen := 1
		for backslashPos+numLen < len(errorMessage) && unicode.IsDigit(rune(errorMessage[backslashPos+numLen])) {
			numLen++
		}
		j, err := strconv.Atoi(errorMessage[backslashPos+1 : backslashPos+numLen])
		if err != nil {
			errorMessageProcessed = backslashPos + 1
			continue
		}
		j-- // 0-based pattern element index

		repTokenPos := 0
		for l := 0; l <= j && l < len(positions); l++ {
			repTokenPos += positions[l]
		}
		nextTokenPos := 0
		if j+1 < len(positions) {
			nextTokenPos = firstMatchTok + repTokenPos + positions[j+1]
		}

		newWay := false
		if len(suggestionMatches) > 0 && matchCounter < len(suggestionMatches) {
			var matches []string
			if j >= len(positions) {
				matches = concatMatches(matchCounter, j, firstMatchTok+repTokenPos, tokenReadings, nextTokenPos, suggestionMatches, langCode)
			} else if j >= 0 && j < len(positions) && positions[j] != 0 {
				matches = concatMatches(matchCounter, j, firstMatchTok+repTokenPos, tokenReadings, nextTokenPos, suggestionMatches, langCode)
			} else {
				matches = []string{""}
			}
			leftSide := errorMessage[:backslashPos]
			rightSide := errorMessage[backslashPos+numLen:]
			if len(matches) == 1 {
				if matches[0] == "" {
					errorMessage = concatWithoutExtraSpace(leftSide, rightSide)
					errorMessageProcessed = len(leftSide)
				} else {
					errorMessage = leftSide + matches[0] + rightSide
					errorMessageProcessed = len(leftSide) + len(matches[0])
				}
			} else {
				// Java formatMultipleSynthesis
				errorMessage = formatMultipleSynthesis(matches, leftSide, rightSide)
				errorMessageProcessed = len(errorMessage)
			}
			matchCounter++
			newWay = true
		}
		if !newWay {
			// bare surface replace (no Match config)
			tokIdx := firstMatchTok + repTokenPos - 1
			surface := ""
			if tokIdx >= 0 && tokIdx < len(tokenReadings) && tokenReadings[tokIdx] != nil {
				surface = tokenReadings[tokIdx].GetToken()
			}
			// replace only this occurrence region
			errorMessage = errorMessage[:backslashPos] + surface + errorMessage[backslashPos+numLen:]
			errorMessageProcessed = backslashPos + len(surface)
		}
	}
	return removeSuppressMisspelled(errorMessage)
}

// removeSuppressMisspelled ports PatternRuleMatcher.removeSuppressMisspelled:
// drop suggestions that contain <pleasespellme/> plus (…form…) or <mistake/>,
// then strip remaining <pleasespellme/> inside suggestions.
func removeSuppressMisspelled(s string) string {
	if s == "" {
		return s
	}
	if !strings.Contains(s, PleaseSpellMe) && !strings.Contains(s, MistakeMarker) {
		return s
	}
	// <suggestion><pleasespellme/>…(…)…</suggestion> or …<mistake/>…
	// allowedChars = [^<>()]*?
	reDrop := regexp.MustCompile(`(?is)<suggestion>` + regexp.QuoteMeta(PleaseSpellMe) +
		`[^<>()]*?(\([^<>()]*\)|` + regexp.QuoteMeta(MistakeMarker) + `)[^<>()]*?</suggestion>`)
	result := reDrop.ReplaceAllString(s, "")
	// remaining <suggestion><pleasespellme/> → <suggestion>
	reStrip := regexp.MustCompile(`(?is)<suggestion>` + regexp.QuoteMeta(PleaseSpellMe))
	result = reStrip.ReplaceAllString(result, "<suggestion>")
	// bare <pleasespellme/> outside suggestion tags (message body)
	result = strings.ReplaceAll(result, PleaseSpellMe, "")
	return result
}

// formatMultipleSynthesis ports PatternRuleMatcher.formatMultipleSynthesis.
// Suggestion tags: <suggestion>…</suggestion> (RuleMatch.SUGGESTION_*_TAG).
func formatMultipleSynthesis(matches []string, leftSide, rightSide string) string {
	const sugStart = "<suggestion>"
	const sugEnd = "</suggestion>"
	suggestionLeft := ""
	suggestionRight := ""
	rightSideNew := rightSide
	errorMessage := leftSide
	if sPos := strings.LastIndex(leftSide, sugStart); sPos >= 0 {
		suggestionLeft = leftSide[sPos+len(sugStart):]
		if suggestionLeft != "" {
			errorMessage = leftSide[:sPos] + sugStart
		}
	}
	if rPos := strings.Index(rightSide, sugEnd); rPos >= 0 {
		suggestionRight = rightSide[:rPos]
		rightSideNew = rightSide[rPos:]
	}
	lastLeftSugEnd := strings.Index(leftSide, sugEnd)
	lastLeftSugStart := strings.LastIndex(leftSide, sugStart)
	var sb strings.Builder
	sb.WriteString(errorMessage)
	for z, m := range matches {
		sb.WriteString(suggestionLeft)
		sb.WriteString(m)
		sb.WriteString(suggestionRight)
		if z < len(matches)-1 && lastLeftSugEnd < lastLeftSugStart {
			sb.WriteString(sugEnd)
			sb.WriteString(", ")
			sb.WriteString(sugStart)
		}
	}
	sb.WriteString(rightSideNew)
	return sb.String()
}

func concatWithoutExtraSpace(leftSide, rightSide string) string {
	if (strings.HasSuffix(leftSide, " ") && strings.HasPrefix(rightSide, "</suggestion>")) ||
		(strings.HasSuffix(leftSide, " ") && rightSide != "" && isWSOrPunct(rightSide[0])) {
		return leftSide[:len(leftSide)-1] + rightSide
	}
	if strings.HasSuffix(leftSide, "suggestion>") && strings.HasPrefix(rightSide, " ") {
		return leftSide + rightSide[1:]
	}
	return leftSide + rightSide
}

func isWSOrPunct(b byte) bool {
	switch b {
	case ' ', '\t', '\n', ',', ':', ';', '.', '!', '?':
		return true
	}
	return false
}

// concatMatches ports PatternRuleMatcher.concatMatches (len=1 phrase path).
func concatMatches(
	start, index, tokenIndex int,
	tokens []*languagetool.AnalyzedTokenReadings,
	nextTokenPos int,
	suggestionMatches []*Match,
	langCode string,
) []string {
	if start < 0 || start >= len(suggestionMatches) || suggestionMatches[start] == nil {
		return []string{""}
	}
	skippedTokens := nextTokenPos - tokenIndex
	if skippedTokens < 0 {
		skippedTokens = 1
	}
	// Java: tokenIndex - 1 is the matched token index into tokens array
	idx := tokenIndex - 1
	ms := NewMatchStateWithSynth(suggestionMatches[start], LanguageSynthesizer(langCode))
	if idx >= 0 && idx < len(tokens) {
		ms.SetTokenRange(tokens, idx, skippedTokens)
	}
	return ms.ToFinalString(langCode)
}

// ExpandSuggestionTemplate formats one suggestion template.
// When the template is a single \N (optional whitespace) and synthesis yields
// multiple forms, returns one string per form (Java multi-suggestion path).
func ExpandSuggestionTemplate(
	tmpl string,
	tokenReadings []*languagetool.AnalyzedTokenReadings,
	positions []int,
	firstMatchTok int,
	suggestionMatches []*Match,
	langCode string,
) []string {
	t := strings.TrimSpace(tmpl)
	// Pure backref: \N only
	if len(t) >= 2 && t[0] == '\\' && unicode.IsDigit(rune(t[1])) {
		only := true
		for i := 1; i < len(t); i++ {
			if !unicode.IsDigit(rune(t[i])) {
				only = false
				break
			}
		}
		if only && len(suggestionMatches) > 0 {
			forms := FormatMatches(tokenReadings, positions, firstMatchTok, t, suggestionMatches, langCode)
			// FormatMatches returns one string; for pure \N with multi forms via
			// formatMultipleSynthesis without suggestion tags, multi forms join wrong.
			// Re-run concat path:
			j, _ := strconv.Atoi(t[1:])
			j--
			repTokenPos := 0
			for l := 0; l <= j && l < len(positions); l++ {
				repTokenPos += positions[l]
			}
			nextTokenPos := 0
			if j+1 < len(positions) {
				nextTokenPos = firstMatchTok + repTokenPos + positions[j+1]
			}
			if j >= 0 && (j >= len(positions) || positions[j] != 0) {
				ms := concatMatches(0, j, firstMatchTok+repTokenPos, tokenReadings, nextTokenPos, suggestionMatches, langCode)
				if len(ms) > 0 {
					return ms
				}
			}
			return []string{forms}
		}
	}
	return []string{FormatMatches(tokenReadings, positions, firstMatchTok, tmpl, suggestionMatches, langCode)}
}

// defaultPositions returns all-1s positions when matchFrom did not track them.
func defaultPositions(n int) []int {
	if n <= 0 {
		return nil
	}
	p := make([]int, n)
	for i := range p {
		p[i] = 1
	}
	return p
}

// FormatMessageAndSuggestions expands \N / <match> in message and suggestion strings.
func FormatMessageAndSuggestions(
	msg string,
	suggs []string,
	matches []*Match,
	tokens []*languagetool.AnalyzedTokenReadings,
	firstMatchTok int,
	positions []int,
	langCode string,
) (string, []string) {
	if len(positions) == 0 {
		// Fallback when caller has no positions: one token per pattern slot unknown —
		// use 1 for each digit index seen is wrong; leave positions empty and FormatMatches
		// will still resolve firstMatchTok+repTokenPos with zero sums for missing slots.
		positions = nil
	}
	outMsg := FormatMatches(tokens, positions, firstMatchTok, msg, matches, langCode)
	outSuggs := make([]string, len(suggs))
	for i, s := range suggs {
		outSuggs[i] = FormatMatches(tokens, positions, firstMatchTok, s, matches, langCode)
	}
	return outMsg, outSuggs
}

func suppressMisspelledIn(matches []*Match) bool {
	for _, m := range matches {
		if m != nil && m.ChecksSpelling() {
			return true
		}
	}
	return false
}

// isParenOnlyForm reports Java empty-synth form "(token)".
func isParenOnlyForm(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')'
}
