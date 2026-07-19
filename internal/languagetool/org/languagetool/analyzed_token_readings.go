package languagetool

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AnalyzedTokenReadings ports org.languagetool.AnalyzedTokenReadings (1:1 expand).
type AnalyzedTokenReadings struct {
	anTokReadings            []*AnalyzedToken
	startPos                 int
	token                    string
	// cleanToken is optional surface after soft-hyphen / fixup (Java cleanToken); nil ⇒ use token.
	cleanToken               *string
	// fixPos ports posFix for getCorrectedTextLength (soft-hyphen removal offsets).
	fixPos                   int
	isWhitespace             bool
	isLinebreak              bool
	isSentStart              bool
	isSentEnd                bool
	isParaEnd                bool
	isWhitespaceBefore       bool
	// whitespaceBeforeChar is the preceding whitespace string (Java whitespaceBeforeChar).
	// Empty when there is no whitespace before the token.
	whitespaceBeforeChar     string
	isImmunized              bool
	immunizationSrcLine      int
	historicalAnnotations    string
	hasSameLemmas            bool
	hasTypographicApostrophe bool
	chunkTags                []string
	isIgnoredBySpeller       bool
	// isPosTagUnknown ports Java isPosTagUnknown: single reading with null POS at construction.
	isPosTagUnknown          bool
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
		anTokReadings:        append([]*AnalyzedToken(nil), tokens...),
		startPos:             startPos,
		token:                tokens[0].GetToken(),
		whitespaceBeforeChar: "", // Java constructor default
	}
	// Java: isPosTagUnknown = tokens.size() == 1 && tokens.get(0).getPOSTag() == null
	r.isPosTagUnknown = len(tokens) == 1 && tokens[0].GetPOSTag() == nil
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
	// Java: setWhitespaceBefore(oldAtr.getWhitespaceBefore())
	r.SetWhitespaceBeforeToken(old.GetWhitespaceBefore())
	// Java: setChunkTags(oldAtr.getChunkTags()) — required for SubjectVerbAgreement etc.
	if len(old.chunkTags) > 0 {
		r.SetChunkTags(old.chunkTags)
	}
	if old.isImmunized {
		r.Immunize(old.immunizationSrcLine)
	}
	if old.isIgnoredBySpeller {
		r.IgnoreSpelling()
	}
	// Java: if (oldAtr.hasTypographicApostrophe()) setTypographicApostrophe()
	if old.hasTypographicApostrophe {
		r.SetTypographicApostrophe(true)
	}
	// cleanToken / posFix are not copied by Java FromOld constructor — leave defaults.
	// Java: setHistoricalAnnotations + addHistoricalAnnotations (only when GlobalConfig.isVerbose)
	r.setHistoricalAnnotations(old.GetHistoricalAnnotations())
	r.addHistoricalAnnotations(old.String(), ruleApplied)
	return r
}

// GetHistoricalAnnotations ports getHistoricalAnnotations.
func (r *AnalyzedTokenReadings) GetHistoricalAnnotations() string {
	if r == nil {
		return ""
	}
	return r.historicalAnnotations
}

// setHistoricalAnnotations ports private setHistoricalAnnotations (verbose-gated).
func (r *AnalyzedTokenReadings) setHistoricalAnnotations(s string) {
	if r == nil || !IsVerbose() {
		return
	}
	r.historicalAnnotations = s
}

// addHistoricalAnnotations ports private addHistoricalAnnotations (verbose-gated).
func (r *AnalyzedTokenReadings) addHistoricalAnnotations(oldValue, ruleApplied string) {
	if r == nil || !IsVerbose() || ruleApplied == "" {
		return
	}
	r.historicalAnnotations = r.GetHistoricalAnnotations() + "\n" + ruleApplied + ": " + oldValue + " -> " + r.String()
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

// HasPosTagAndLemma ports AnalyzedTokenReadings.hasPosTagAndLemma.
func (r *AnalyzedTokenReadings) HasPosTagAndLemma(posTag, lemma string) bool {
	if r == nil {
		return false
	}
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() == nil || *reading.GetPOSTag() != posTag {
			continue
		}
		lem := ""
		if reading.GetLemma() != nil {
			lem = *reading.GetLemma()
		}
		if lem == lemma {
			return true
		}
	}
	return false
}

// ReadingWithExactTag returns the first reading whose POS equals tag (or nil).
func (r *AnalyzedTokenReadings) ReadingWithExactTag(tag string) *AnalyzedToken {
	if r == nil {
		return nil
	}
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && *reading.GetPOSTag() == tag {
			return reading
		}
	}
	return nil
}

// ReadingWithTagRegex ports readingWithTagRegex — first reading whose POS fully matches the regex.
func (r *AnalyzedTokenReadings) ReadingWithTagRegex(posTagRegex string) *AnalyzedToken {
	if r == nil {
		return nil
	}
	re, err := regexp.Compile("^(?:" + posTagRegex + ")$")
	if err != nil {
		return nil
	}
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && re.MatchString(*reading.GetPOSTag()) {
			return reading
		}
	}
	return nil
}

// ReadingWithLemma ports readingWithLemma — first reading with exact lemma.
func (r *AnalyzedTokenReadings) ReadingWithLemma(lemma string) *AnalyzedToken {
	if r == nil {
		return nil
	}
	for _, reading := range r.anTokReadings {
		if reading.GetLemma() != nil && *reading.GetLemma() == lemma {
			return reading
		}
	}
	return nil
}

func (r *AnalyzedTokenReadings) HasPartialPosTag(posTag string) bool {
	for _, reading := range r.anTokReadings {
		if reading.GetPOSTag() != nil && strings.Contains(*reading.GetPOSTag(), posTag) {
			return true
		}
	}
	return false
}

// HasAnyPartialPosTag ports hasAnyPartialPosTag.
func (r *AnalyzedTokenReadings) HasAnyPartialPosTag(posTags ...string) bool {
	for _, p := range posTags {
		if r.HasPartialPosTag(p) {
			return true
		}
	}
	return false
}

// HasLemma ports hasLemma — true if any reading has the given lemma.
func (r *AnalyzedTokenReadings) HasLemma(lemma string) bool {
	return r.ReadingWithLemma(lemma) != nil
}

// HasReading ports hasReading — true if there is at least one reading slot.
func (r *AnalyzedTokenReadings) HasReading() bool {
	return r != nil && len(r.anTokReadings) > 0
}

// HasSameLemmas ports hasSameLemmas (all readings share one lemma).
func (r *AnalyzedTokenReadings) HasSameLemmas() bool {
	if r == nil {
		return true
	}
	return r.hasSameLemmas
}

// IsPosTagUnknown ports isPosTagUnknown (single untagged reading at construction).
func (r *AnalyzedTokenReadings) IsPosTagUnknown() bool {
	return r != nil && r.isPosTagUnknown
}

// GetImmunizationSourceLine ports getImmunizationSourceLine.
func (r *AnalyzedTokenReadings) GetImmunizationSourceLine() int {
	if r == nil {
		return 0
	}
	return r.immunizationSrcLine
}

// SetPosFix / GetPosFix port posFix (soft-hyphen position fixes).
func (r *AnalyzedTokenReadings) SetPosFix(fix int) {
	if r != nil {
		r.fixPos = fix
	}
}

func (r *AnalyzedTokenReadings) GetPosFix() int {
	if r == nil {
		return 0
	}
	return r.fixPos
}

// SetCleanToken / GetCleanToken port cleanToken (Experimental in Java 5.1).
func (r *AnalyzedTokenReadings) SetCleanToken(clean string) {
	if r == nil {
		return
	}
	c := clean
	r.cleanToken = &c
}

// SetTokenSurface ports AnalyzedTokenReadings.addReading when a longer surface
// replaces this.token (Java soft-hyphen: orig with U+00AD becomes getToken()).
func (r *AnalyzedTokenReadings) SetTokenSurface(surface string) {
	if r == nil {
		return
	}
	r.token = surface
	r.isWhitespace = tools.IsWhitespace(surface)
	r.isLinebreak = surface == "\n" || surface == "\r\n" || surface == "\r" || surface == "\n\r"
}

func (r *AnalyzedTokenReadings) GetCleanToken() string {
	if r == nil {
		return ""
	}
	if r.cleanToken != nil {
		return *r.cleanToken
	}
	return r.token
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
	oldValue := r.String()
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
	r.addHistoricalAnnotations(oldValue, ruleApplied)
}

func (r *AnalyzedTokenReadings) RemoveReading(token *AnalyzedToken, ruleApplied string) {
	oldValue := r.String()
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
	r.addHistoricalAnnotations(oldValue, ruleApplied)
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

// ReplaceReadings ports the Java disambiguator REPLACE path that builds
// new AnalyzedTokenReadings(old, newReadings, ruleId): keep metadata, set readings.
func (r *AnalyzedTokenReadings) ReplaceReadings(newReadings []*AnalyzedToken, ruleApplied string) {
	if r == nil || len(newReadings) == 0 {
		return
	}
	_ = ruleApplied
	sentEnd, paraEnd := r.isSentEnd, r.isParaEnd
	for _, t := range newReadings {
		if t != nil {
			t.SetWhitespaceBefore(r.isWhitespaceBefore)
		}
	}
	r.anTokReadings = append([]*AnalyzedToken(nil), newReadings...)
	r.token = newReadings[0].GetToken()
	r.setNoRealPOStag()
	r.hasSameLemmas = r.areLemmasSame()
	r.isSentEnd, r.isParaEnd = false, false
	if sentEnd {
		r.SetSentEnd()
	}
	if paraEnd {
		r.SetParagraphEnd()
	}
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
func (r *AnalyzedTokenReadings) IsWhitespaceBefore() bool {
	if r == nil {
		return false
	}
	return r.isWhitespaceBefore
}

// GetWhitespaceBefore ports AnalyzedTokenReadings.getWhitespaceBefore (preceding ws string).
func (r *AnalyzedTokenReadings) GetWhitespaceBefore() string {
	if r == nil {
		return ""
	}
	return r.whitespaceBeforeChar
}

// SetWhitespaceBefore ports boolean whitespace-before flag (does not change whitespaceBeforeChar).
func (r *AnalyzedTokenReadings) SetWhitespaceBefore(v bool) {
	if r == nil {
		return
	}
	r.isWhitespaceBefore = v
	for _, t := range r.anTokReadings {
		t.SetWhitespaceBefore(v)
	}
}

// SetWhitespaceBeforeToken ports setWhitespaceBefore(String prevToken).
// Stores prevToken as whitespaceBeforeChar when it is whitespace.
func (r *AnalyzedTokenReadings) SetWhitespaceBeforeToken(prevToken string) {
	if r == nil {
		return
	}
	isWS := prevToken != "" && tools.IsWhitespace(prevToken)
	r.isWhitespaceBefore = isWS
	for _, t := range r.anTokReadings {
		t.SetWhitespaceBefore(isWS)
	}
	// Java only assigns whitespaceBeforeChar when isWhitespaceBefore is true.
	if isWS {
		r.whitespaceBeforeChar = prevToken
	}
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
