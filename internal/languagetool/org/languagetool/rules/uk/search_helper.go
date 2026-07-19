package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Java LemmaHelper.QUOTES_PATTERN used by SearchHelper.Match
var searchQuotesRE = regexp.MustCompile(`^[«»„""\x{201C}]$`)

// SearchCondition ports SearchHelper.Condition.
type SearchCondition struct {
	Postag       *regexp.Regexp
	Lemma        *regexp.Regexp
	TokenPattern *regexp.Regexp
	TokenStr     string // exact ignore-case clean token
	Negate       bool
}

// ConditionPostag builds Condition.postag(pattern).
func ConditionPostag(pattern *regexp.Regexp) SearchCondition {
	return SearchCondition{Postag: pattern}
}

// ConditionLemma builds Condition.lemma(pattern).
func ConditionLemma(pattern *regexp.Regexp) SearchCondition {
	return SearchCondition{Lemma: pattern}
}

// ConditionTokenRE builds Condition.token(Pattern).
func ConditionTokenRE(pattern *regexp.Regexp) SearchCondition {
	return SearchCondition{TokenPattern: pattern}
}

// ConditionToken builds Condition.token(String).
func ConditionToken(token string) SearchCondition {
	return SearchCondition{TokenStr: token}
}

// Negate marks the condition as negated (Java Condition.negate).
func (c SearchCondition) WithNegate() SearchCondition {
	c.Negate = true
	return c
}

// Matches ports Condition.matches (POS/lemma Matcher.matches; token Pattern.matches).
func (c SearchCondition) Matches(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return c.Negate
	}
	ok := true
	// Java PosTagHelper.hasPosTag(…, Pattern) = Matcher.matches
	if c.Postag != nil && !HasPosTagMatches(tok, c.Postag) {
		ok = false
	}
	// Java hasLemma(…, Pattern) = Matcher.matches on lemma
	if ok && c.Lemma != nil && !HasLemmaTokenRE(tok, c.Lemma) {
		ok = false
	}
	if ok && c.TokenPattern != nil {
		ct := tok.GetCleanToken()
		if ct == "" {
			ct = tok.GetToken()
		}
		// Java token.matcher(clean).matches()
		loc := c.TokenPattern.FindStringIndex(ct)
		if loc == nil || loc[0] != 0 || loc[1] != len(ct) {
			ok = false
		}
	}
	if ok && c.TokenStr != "" {
		ct := tok.GetCleanToken()
		if ct == "" {
			ct = tok.GetToken()
		}
		if !strings.EqualFold(ct, c.TokenStr) {
			ok = false
		}
	}
	if c.Negate {
		return !ok
	}
	return ok
}

// SearchMatch ports SearchHelper.Match (token-line targets and/or Condition targets).
type SearchMatch struct {
	// Targets is the legacy surface string list (tokenLine).
	Targets []string
	// Conditions is the Java targets list when set via Target().
	Conditions []SearchCondition
	// Skips are optional skip Conditions (Java skip()).
	Skips         []SearchCondition
	Limit         int // max logical distance; -1 = unlimited
	IgnoreQuotes  bool
	IgnoreInserts bool
}

// NewSearchMatch builds a matcher for space-separated target tokens (Java tokenLine).
func NewSearchMatch(tokenLine string) *SearchMatch {
	line := strings.ReplaceAll(tokenLine, ",", " ,")
	parts := strings.Fields(line)
	conds := make([]SearchCondition, 0, len(parts))
	for _, p := range parts {
		conds = append(conds, ConditionToken(p))
	}
	return &SearchMatch{
		Targets:      parts,
		Conditions:   conds,
		Limit:        -1,
		IgnoreQuotes: true,
	}
}

// WithLimit sets the max logical steps (Java limit).
func (m *SearchMatch) WithLimit(n int) *SearchMatch {
	m.Limit = n
	return m
}

// IgnoreInsertsOn enables comma-insert skipping (Java ignoreInserts()).
func (m *SearchMatch) IgnoreInsertsOn() *SearchMatch {
	m.IgnoreInserts = true
	return m
}

// Target sets Condition targets (Java target(...)).
func (m *SearchMatch) Target(conds ...SearchCondition) *SearchMatch {
	m.Conditions = append([]SearchCondition{}, conds...)
	m.Targets = nil
	return m
}

// Skip sets skip Conditions (Java skip(...)).
func (m *SearchMatch) Skip(conds ...SearchCondition) *SearchMatch {
	m.Skips = append([]SearchCondition{}, conds...)
	return m
}

func (m *SearchMatch) targets() []SearchCondition {
	if len(m.Conditions) > 0 {
		return m.Conditions
	}
	out := make([]SearchCondition, 0, len(m.Targets))
	for _, t := range m.Targets {
		out = append(out, ConditionToken(t))
	}
	return out
}

func (m *SearchMatch) canSkip(tok *languagetool.AnalyzedTokenReadings) bool {
	if len(m.Skips) == 0 {
		return true // Java: skips.isEmpty() → true
	}
	for _, s := range m.Skips {
		if s.Matches(tok) {
			return true
		}
	}
	return false
}

func (m *SearchMatch) isQuoteToken(s string) bool {
	return searchQuotesRE.MatchString(s) || QuotesPattern.MatchString(s)
}

// MAfterATR ports Match.mAfter on AnalyzedTokenReadings; returns end index of match (Java pos-1)
// or -1. For compatibility with string MAfter callers that expect start index, see MAfter.
func (m *SearchMatch) MAfterATR(tokens []*languagetool.AnalyzedTokenReadings, pos int) int {
	conds := m.targets()
	if len(conds) == 0 {
		return -1
	}
	foundFirst := false
	logical := 0
	iCond := 0
	for iCond < len(conds) {
		if pos+len(conds)-iCond > len(tokens) {
			return -1
		}
		if m.Limit > 0 && logical > m.Limit {
			return -1
		}
		logical++
		if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
			return -1
		}
		cur := tokens[pos]
		tokStr := cur.GetToken()
		if m.IgnoreQuotes && m.isQuoteToken(tokStr) {
			pos++
			continue
		}
		// Java: ignoreInserts && "(" → jump pos to matching ")", then fall through to matches
		if m.IgnoreInserts && tokStr == "(" {
			for i := pos + 1; i < len(tokens); i++ {
				if tokens[i] != nil && tokens[i].GetToken() == ")" {
					pos = i
					// Java continue is only on the inner for (keeps scanning for more ")")
					// after loop, outer falls through with pos at last ")"
				}
			}
			// re-bind current token after jump
			if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
				return -1
			}
			cur = tokens[pos]
			tokStr = cur.GetToken()
		}
		// Java ignoreInserts(tokens, pos, +1): pos+=2 then for pos++ → skip 3 tokens
		if m.ignoreCommaInsert(tokens, pos, +1) {
			pos += 2
			// next loop iteration advances one more (manual pos++ below only on match/skip)
			// mirror Java: after pos+=2, for does pos++ → net +3 from original start of insert
			pos++
			continue
		}
		if !conds[iCond].Matches(cur) {
			if foundFirst {
				return -1
			}
			if !m.canSkip(cur) {
				return -1
			}
			pos++
			continue
		}
		foundFirst = true
		iCond++
		pos++
	}
	return pos - 1 // Java mAfter returns pos-1 after matching last target
}

// MBeforeATR ports Match.mBefore; returns start index of match or -1.
func (m *SearchMatch) MBeforeATR(tokens []*languagetool.AnalyzedTokenReadings, pos int) int {
	conds := m.targets()
	if len(conds) == 0 {
		return -1
	}
	foundFirst := false
	logical := 0
	iCond := len(conds) - 1
	for iCond >= 0 {
		if pos-1 < iCond {
			return -1
		}
		if m.Limit > 0 && logical > m.Limit {
			return -1
		}
		logical++
		if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
			return -1
		}
		cur := tokens[pos]
		tokStr := cur.GetToken()
		if m.IgnoreQuotes && m.isQuoteToken(tokStr) {
			pos--
			continue
		}
		// Java: ignoreInserts && ")" → jump pos to matching "(", then fall through
		if m.IgnoreInserts && tokStr == ")" {
			for i := pos - 1; i >= 1; i-- {
				if tokens[i] != nil && tokens[i].GetToken() == "(" {
					pos = i
					// fall through like mAfter paren arm
				}
			}
			if pos < 0 || pos >= len(tokens) || tokens[pos] == nil {
				return -1
			}
			cur = tokens[pos]
			tokStr = cur.GetToken()
		}
		// Java ignoreInserts reverse: pos += 2*dir with dir=-1 → pos-=2 then for pos-- → net -3
		if m.ignoreCommaInsert(tokens, pos, -1) {
			pos -= 2
			pos--
			continue
		}
		if !conds[iCond].Matches(cur) {
			if foundFirst {
				return -1
			}
			if !m.canSkip(cur) {
				return -1
			}
			pos--
			continue
		}
		foundFirst = true
		iCond--
		pos--
	}
	return pos + 1 // Java return pos+1 after last match
}

// MNowATR ports Match.mNow (limit 0 mAfter).
func (m *SearchMatch) MNowATR(tokens []*languagetool.AnalyzedTokenReadings, pos int) int {
	lim := m.Limit
	m.Limit = 0
	defer func() { m.Limit = lim }()
	return m.MAfterATR(tokens, pos)
}

func (m *SearchMatch) ignoreCommaInsert(tokens []*languagetool.AnalyzedTokenReadings, pos, dir int) bool {
	if !m.IgnoreInserts {
		return false
	}
	mid := pos + dir
	far := pos + 2*dir
	if dir > 0 {
		if pos+3 >= len(tokens) {
			return false
		}
	} else {
		if pos-3 <= 0 {
			return false
		}
	}
	if tokens[pos] == nil || tokens[mid] == nil || tokens[far] == nil {
		return false
	}
	if tokens[pos].GetToken() != "," || tokens[far].GetToken() != "," {
		return false
	}
	if HasPosTagPart(tokens[mid], "insert") {
		return true
	}
	return HasLemmaTokenAny(tokens[mid], []string{"зокрема", "відповідно"})
}

// MAfter finds targets in order on plain strings; returns start index (legacy API).
func (m *SearchMatch) MAfter(tokens []string, pos int) int {
	atrs := make([]*languagetool.AnalyzedTokenReadings, len(tokens))
	for i, t := range tokens {
		atrs[i] = surfaceATR(t)
	}
	end := m.MAfterATR(atrs, pos)
	if end < 0 {
		return -1
	}
	// convert end index to start for legacy callers
	n := len(m.targets())
	start := end - n + 1
	if start < 0 {
		return -1
	}
	return start
}

// MBefore finds targets ending at pos on plain strings; returns start index.
func (m *SearchMatch) MBefore(tokens []string, pos int) int {
	atrs := make([]*languagetool.AnalyzedTokenReadings, len(tokens))
	for i, t := range tokens {
		atrs[i] = surfaceATR(t)
	}
	return m.MBeforeATR(atrs, pos)
}

func surfaceATR(tok string) *languagetool.AnalyzedTokenReadings {
	t := tok
	return languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(tok, nil, &t)},
		0,
	)
}
