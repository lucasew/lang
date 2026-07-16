package languagetool

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AnalyzedTokenReadings ports org.languagetool.AnalyzedTokenReadings (subset needed for tests; expand 1:1).
type AnalyzedTokenReadings struct {
	anTokReadings            []*AnalyzedToken
	startPos                 int
	token                    string
	isWhitespace             bool
	isLinebreak              bool
	isSentStart              bool
	isSentEnd                bool
	isParaEnd                bool
	isWhitespaceBefore       bool
	isImmunized              bool
	immunizationSrcLine      int
	historicalAnnotations    string
	hasSameLemmas            bool
	hasTypographicApostrophe bool
	chunkTags                []string
	isIgnoredBySpeller       bool
}

func NewAnalyzedTokenReadings(tok *AnalyzedToken) *AnalyzedTokenReadings {
	return NewAnalyzedTokenReadingsAt(tok, 0)
}

func NewAnalyzedTokenReadingsAt(tok *AnalyzedToken, startPos int) *AnalyzedTokenReadings {
	return NewAnalyzedTokenReadingsList([]*AnalyzedToken{tok}, startPos)
}

func NewAnalyzedTokenReadingsList(tokens []*AnalyzedToken, startPos int) *AnalyzedTokenReadings {
	if len(tokens) == 0 {
		panic("AnalyzedTokenReadings: empty tokens")
	}
	r := &AnalyzedTokenReadings{
		anTokReadings: append([]*AnalyzedToken(nil), tokens...),
		startPos:      startPos,
		token:         tokens[0].GetToken(),
	}
	r.isWhitespace = tools.IsWhitespace(r.token)
	r.isWhitespaceBefore = tokens[0].IsWhitespaceBefore()
	r.isLinebreak = r.token == "\n" || r.token == "\r\n" || r.token == "\r" || r.token == "\n\r"
	if pt := tokens[0].GetPOSTag(); pt != nil {
		r.isSentStart = *pt == SentenceStartTagName
	}
	r.isParaEnd = r.HasPosTag(ParagraphEndTagName)
	r.isSentEnd = r.HasPosTag(SentenceEndTagName)
	r.setNoRealPOStag()
	r.hasSameLemmas = r.areLemmasSame()
	return r
}

// NewAnalyzedTokenReadingsFromOld ports constructor (old, newReadings, ruleApplied).
func NewAnalyzedTokenReadingsFromOld(old *AnalyzedTokenReadings, newReadings []*AnalyzedToken, ruleApplied string) *AnalyzedTokenReadings {
	r := NewAnalyzedTokenReadingsList(newReadings, old.startPos)
	if old.IsSentenceEnd() {
		r.SetSentEnd()
	}
	if old.IsParagraphEnd() {
		r.SetParagraphEnd()
	}
	r.isWhitespaceBefore = old.isWhitespaceBefore
	for _, t := range r.anTokReadings {
		t.SetWhitespaceBefore(r.isWhitespaceBefore)
	}
	if old.isImmunized {
		r.Immunize(old.immunizationSrcLine)
	}
	if old.isIgnoredBySpeller {
		r.IgnoreSpelling()
	}
	r.historicalAnnotations = old.historicalAnnotations
	_ = ruleApplied
	return r
}

func (r *AnalyzedTokenReadings) GetReadings() []*AnalyzedToken {
	return append([]*AnalyzedToken(nil), r.anTokReadings...)
}

func (r *AnalyzedTokenReadings) Readings() []*AnalyzedToken { return r.GetReadings() }

func (r *AnalyzedTokenReadings) GetAnalyzedToken(idx int) *AnalyzedToken {
	return r.anTokReadings[idx]
}

func (r *AnalyzedTokenReadings) HasPosTag(posTag string) bool {
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && *reading.GetPOSTag() == posTag {
			return true
		}
	}
	return false
}

func (r *AnalyzedTokenReadings) HasPartialPosTag(posTag string) bool {
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && strings.Contains(*reading.GetPOSTag(), posTag) {
			return true
		}
	}
	return false
}

// HasPosTagStartingWith ports AnalyzedTokenReadings.hasPosTagStartingWith.
func (r *AnalyzedTokenReadings) HasPosTagStartingWith(posTag string) bool {
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && strings.HasPrefix(*reading.GetPOSTag(), posTag) {
			return true
		}
	}
	return false
}

// HasAnyLemma ports AnalyzedTokenReadings.hasAnyLemma.
func (r *AnalyzedTokenReadings) HasAnyLemma(lemmas ...string) bool {
	for _, reading := range r.anTokReadings {
		lem := reading.GetLemma()
		if lem == nil {
			continue
		}
		for _, want := range lemmas {
			if *lem == want {
				return true
			}
		}
	}
	return false
}

// IsTagged ports AnalyzedTokenReadings.isTagged — true if any reading has a real POS tag.
func (r *AnalyzedTokenReadings) IsTagged() bool {
	for _, element := range r.anTokReadings {
		if !element.HasNoTag() {
			return true
		}
	}
	return false
}

// HasTypographicApostrophe ports AnalyzedTokenReadings.hasTypographicApostrophe.
func (r *AnalyzedTokenReadings) HasTypographicApostrophe() bool {
	return r.hasTypographicApostrophe
}

// SetTypographicApostrophe ports setTypographicApostrophe.
func (r *AnalyzedTokenReadings) SetTypographicApostrophe(v bool) {
	r.hasTypographicApostrophe = v
}

func (r *AnalyzedTokenReadings) MatchesPosTagRegex(posTagRegex string) bool {
	re := regexp.MustCompile("^(?:" + posTagRegex + ")$")
	// Java Pattern.matches is full match; Go Compile + MatchString is full match for whole string
	// Actually Java matches() = entire region. regexp MatchString matches whole string in Go if anchors...
	// Pattern.compile(posTagRegex).matcher(tag).matches() — full string match
	re2, err := regexp.Compile("^(?:" + posTagRegex + ")$")
	if err != nil {
		re2 = re
	}
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && re2.MatchString(*reading.GetPOSTag()) {
			return true
		}
	}
	return false
}

func (r *AnalyzedTokenReadings) AddReading(token *AnalyzedToken, ruleApplied string) {
	_ = ruleApplied
	l := make([]*AnalyzedToken, 0, len(r.anTokReadings)+1)
	// Java: subList(0, length-1) then maybe add last if POS non-null
	if len(r.anTokReadings) > 0 {
		l = append(l, r.anTokReadings[:len(r.anTokReadings)-1]...)
		last := r.anTokReadings[len(r.anTokReadings)-1]
		if last.GetPOSTag() != nil {
			l = append(l, last)
		}
	}
	token.SetWhitespaceBefore(r.isWhitespaceBefore)
	l = append(l, token)
	r.anTokReadings = l
	if len(token.GetToken()) > len(r.token) {
		r.token = token.GetToken()
	}
	r.anTokReadings[len(r.anTokReadings)-1].SetWhitespaceBefore(r.isWhitespaceBefore)
	r.isParaEnd = r.HasPosTag(ParagraphEndTagName)
	r.isSentEnd = r.HasPosTag(SentenceEndTagName)
	r.setNoRealPOStag()
	r.hasSameLemmas = r.areLemmasSame()
}

func (r *AnalyzedTokenReadings) RemoveReading(token *AnalyzedToken, ruleApplied string) {
	_ = ruleApplied
	tmpTok := NewAnalyzedToken(token.GetToken(), token.GetPOSTag(), token.GetLemma())
	tmpTok.SetWhitespaceBefore(r.isWhitespaceBefore)
	var l []*AnalyzedToken
	removedSentEnd, removedParaEnd := false, false
	for _, anTokReading := range r.anTokReadings {
		if !anTokReading.Matches(tmpTok) {
			l = append(l, anTokReading)
		} else if anTokReading.GetPOSTag() != nil && *anTokReading.GetPOSTag() == SentenceEndTagName {
			removedSentEnd = true
		} else if anTokReading.GetPOSTag() != nil && *anTokReading.GetPOSTag() == ParagraphEndTagName {
			removedParaEnd = true
		}
	}
	if len(l) == 0 {
		empty := NewAnalyzedToken(r.token, nil, nil)
		empty.SetWhitespaceBefore(r.isWhitespaceBefore)
		l = append(l, empty)
	}
	r.anTokReadings = l
	r.setNoRealPOStag()
	if removedSentEnd {
		r.isSentEnd = false
		r.SetSentEnd()
	}
	if removedParaEnd {
		r.isParaEnd = false
		r.SetParagraphEnd()
	}
	r.hasSameLemmas = r.areLemmasSame()
}

func (r *AnalyzedTokenReadings) LeaveReading(token *AnalyzedToken) {
	tmpTok := NewAnalyzedToken(token.GetToken(), token.GetPOSTag(), token.GetLemma())
	tmpTok.SetWhitespaceBefore(r.isWhitespaceBefore)
	var l []*AnalyzedToken
	for _, anTokReading := range r.anTokReadings {
		if anTokReading.Matches(tmpTok) {
			l = append(l, anTokReading)
		}
	}
	if len(l) == 0 {
		empty := NewAnalyzedToken(r.token, nil, nil)
		empty.SetWhitespaceBefore(r.isWhitespaceBefore)
		l = append(l, empty)
	}
	r.anTokReadings = l
	r.setNoRealPOStag()
	r.hasSameLemmas = r.areLemmasSame()
}

func (r *AnalyzedTokenReadings) GetReadingsLength() int { return len(r.anTokReadings) }
func (r *AnalyzedTokenReadings) IsWhitespace() bool     { return r.isWhitespace }
func (r *AnalyzedTokenReadings) IsLinebreak() bool      { return r.isLinebreak }
func (r *AnalyzedTokenReadings) IsSentenceStart() bool  { return r.isSentStart }
func (r *AnalyzedTokenReadings) IsParagraphEnd() bool   { return r.isParaEnd }
func (r *AnalyzedTokenReadings) IsSentenceEnd() bool    { return r.isSentEnd }
func (r *AnalyzedTokenReadings) GetToken() string       { return r.token }
func (r *AnalyzedTokenReadings) GetStartPos() int       { return r.startPos }

func (r *AnalyzedTokenReadings) SetParagraphEnd() {
	if !r.IsParagraphEnd() {
		var lemma *string
		if r.GetAnalyzedToken(0).GetLemma() != nil {
			l := *r.GetAnalyzedToken(0).GetLemma()
			lemma = &l
		}
		tok := r.GetToken()
		tag := ParagraphEndTagName
		paragraphEnd := NewAnalyzedToken(tok, &tag, lemma)
		r.AddReading(paragraphEnd, "add_paragaph_end")
	}
}

func (r *AnalyzedTokenReadings) SetSentEnd() {
	if !r.IsSentenceEnd() {
		var lemma *string
		if r.GetAnalyzedToken(0).GetLemma() != nil {
			l := *r.GetAnalyzedToken(0).GetLemma()
			lemma = &l
		}
		tok := r.GetToken()
		tag := SentenceEndTagName
		sentenceEnd := NewAnalyzedToken(tok, &tag, lemma)
		r.AddReading(sentenceEnd, "")
	}
}

func (r *AnalyzedTokenReadings) Immunize(sourceLine int) {
	r.isImmunized = true
	r.immunizationSrcLine = sourceLine
}

func (r *AnalyzedTokenReadings) IsImmunized() bool { return r.isImmunized }

// IgnoreSpelling ports AnalyzedTokenReadings.ignoreSpelling.
func (r *AnalyzedTokenReadings) IgnoreSpelling() {
	if r != nil {
		r.isIgnoredBySpeller = true
	}
}

// IsIgnoredBySpeller ports AnalyzedTokenReadings.isIgnoredBySpeller.
func (r *AnalyzedTokenReadings) IsIgnoredBySpeller() bool {
	return r != nil && r.isIgnoredBySpeller
}

func (r *AnalyzedTokenReadings) String() string {
	var sb strings.Builder
	sb.WriteString(r.token)
	sb.WriteByte('[')
	for i, element := range r.anTokReadings {
		if i > 0 {
			// Java joins with comma after each including trailing then deletes last comma
		}
		sb.WriteString(element.String())
		if !element.IsWhitespaceBefore() {
			sb.WriteByte('*')
		}
		sb.WriteByte(',')
	}
	s := sb.String()
	if len(s) > 0 && s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	s += "]"
	if r.IsImmunized() {
		s += "{!},"
	}
	return s
}

func (r *AnalyzedTokenReadings) setNoRealPOStag() {
	hasNoPOStag := !r.IsLinebreak()
	for _, an := range r.anTokReadings {
		posTag := an.GetPOSTag()
		if posTag != nil && (*posTag == ParagraphEndTagName || *posTag == SentenceEndTagName) {
			continue
		}
		if posTag != nil {
			hasNoPOStag = false
			break
		}
	}
	for _, an := range r.anTokReadings {
		an.SetNoPOSTag(hasNoPOStag)
	}
}

func (r *AnalyzedTokenReadings) areLemmasSame() bool {
	if len(r.anTokReadings) == 0 {
		return true
	}
	var first *string
	for i, t := range r.anTokReadings {
		if i == 0 {
			first = t.GetLemma()
			continue
		}
		if !strPtrEq(first, t.GetLemma()) {
			return false
		}
	}
	return true
}

func (r *AnalyzedTokenReadings) GetEndPos() int {
	// Java: startPos + token.length() UTF-16
	n := 0
	for _, rr := range r.token {
		if rr >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	// simpler with utf16
	return r.startPos + utf16Len(r.token)
}

func (r *AnalyzedTokenReadings) SetStartPos(p int) { r.startPos = p }

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func (r *AnalyzedTokenReadings) IsFieldCode() bool {
	t := r.token
	return t == "\u0001" || t == "\u0002"
}
func (r *AnalyzedTokenReadings) IsWhitespaceBefore() bool { return r.isWhitespaceBefore }

// SetWhitespaceBefore ports AnalyzedTokenReadings.setWhitespaceBefore(boolean) +
// the prevToken overload used by JLanguageTool (whitespace if prev is whitespace).
func (r *AnalyzedTokenReadings) SetWhitespaceBefore(v bool) {
	r.isWhitespaceBefore = v
	for _, t := range r.anTokReadings {
		t.SetWhitespaceBefore(v)
	}
}

// SetWhitespaceBeforeToken ports setWhitespaceBefore(String prevToken).
func (r *AnalyzedTokenReadings) SetWhitespaceBeforeToken(prevToken string) {
	r.SetWhitespaceBefore(prevToken != "" && tools.IsWhitespace(prevToken))
}

// IsNonWord ports AnalyzedTokenReadings.isNonWord — punctuation/bracket-only tokens.
func (r *AnalyzedTokenReadings) IsNonWord() bool {
	return nonWordRE.MatchString(r.token)
}

var nonWordRE = regexp.MustCompile(`^[.?!…:;,~’'"„“”»«‚‘›‹()\[\]\-–—*×∗·+÷/=]$`)

// SetChunkTags ports AnalyzedTokenReadings.setChunkTags (tags as plain strings).
func (r *AnalyzedTokenReadings) SetChunkTags(tags []string) {
	if r == nil {
		return
	}
	r.chunkTags = append([]string(nil), tags...)
}

// GetChunkTags returns assigned chunk tags (may be nil).
func (r *AnalyzedTokenReadings) GetChunkTags() []string {
	if r == nil {
		return nil
	}
	return r.chunkTags
}

// MatchesChunkRegex reports whether any chunk tag matches the regex.
func (r *AnalyzedTokenReadings) MatchesChunkRegex(chunkRegex string) bool {
	if r == nil || chunkRegex == "" {
		return false
	}
	re, err := regexp.Compile(chunkRegex)
	if err != nil {
		return false
	}
	for _, c := range r.chunkTags {
		if re.MatchString(c) {
			return true
		}
	}
	return false
}
