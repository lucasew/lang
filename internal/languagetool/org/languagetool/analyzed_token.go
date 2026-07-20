package languagetool

// AnalyzedToken ports org.languagetool.AnalyzedToken (1:1 behavior).
type AnalyzedToken struct {
	token            string
	posTag           *string // nil == Java null
	lemma            *string
	lemmaOrToken     string
	whitespaceBefore bool
	hasNoPOSTag      bool
}

// NewAnalyzedToken ports AnalyzedToken(String token, String posTag, String lemma).
// Empty posTag with useNilPosTag/lemma: pass nil pointers via NewAnalyzedTokenPtr.
func NewAnalyzedToken(token string, posTag, lemma *string) *AnalyzedToken {
	// Java: Objects.requireNonNull(token) — empty string is allowed; null is not
	// (Go string cannot be null).
	t := &AnalyzedToken{token: token}
	// hasNoPOSTag uses the *original* posTag parameter (before trim), matching Java:
	//   hasNoPOSTag = (posTag == null
	//       || SENTENCE_END_TAGNAME.equals(posTag)
	//       || PARAGRAPH_END_TAGNAME.equals(posTag));
	// Note: equals is on the constant, so null-safe; whitespace-padded special tags
	// do not count as "no tag" in Java even after this.posTag is trimmed.
	rawPos := posTag
	if posTag != nil {
		// Java: posTag != null ? intern(posTag.trim()) : null
		p := trimSpace(*posTag)
		t.posTag = &p
	}
	if lemma != nil {
		l := *lemma
		t.lemma = &l
	}
	if t.lemma == nil {
		t.lemmaOrToken = t.token
	} else {
		t.lemmaOrToken = *t.lemma
	}
	t.hasNoPOSTag = rawPos == nil ||
		(rawPos != nil && (*rawPos == SentenceEndTagName || *rawPos == ParagraphEndTagName))
	return t
}

// NewAnalyzedTokenStr is a convenience matching common Java calls with nullable strings
// represented as: use optional - for null lemma/pos pass the NullSentinel or use Ptr helpers.
func NewAnalyzedTokenStr(token, posTag, lemma string, posNull, lemmaNull bool) *AnalyzedToken {
	var p, l *string
	if !posNull {
		p = &posTag
	}
	if !lemmaNull {
		l = &lemma
	}
	return NewAnalyzedToken(token, p, l)
}

func trimSpace(s string) string {
	// Java String.trim()
	start, end := 0, len(s)
	for start < end && (s[start] <= ' ') {
		start++
	}
	for end > start && (s[end-1] <= ' ') {
		end--
	}
	return s[start:end]
}

func (t *AnalyzedToken) GetToken() string { return t.token }

func (t *AnalyzedToken) GetPOSTag() *string { return t.posTag }

func (t *AnalyzedToken) GetLemma() *string { return t.lemma }

func (t *AnalyzedToken) SetWhitespaceBefore(v bool) { t.whitespaceBefore = v }

func (t *AnalyzedToken) IsWhitespaceBefore() bool { return t.whitespaceBefore }

// Matches ports AnalyzedToken.matches.
func (t *AnalyzedToken) Matches(an *AnalyzedToken) bool {
	if t.Equals(an) {
		return true
	}
	if an.GetToken() == "" && an.GetLemma() == nil && an.GetPOSTag() == nil {
		return false
	}
	found := true
	if an.GetToken() != "" {
		found = an.GetToken() == t.token
	}
	if an.GetLemma() != nil {
		found = found && t.lemma != nil && *an.GetLemma() == *t.lemma
	}
	if an.GetPOSTag() != nil {
		found = found && t.posTag != nil && *an.GetPOSTag() == *t.posTag
	}
	return found
}

func (t *AnalyzedToken) HasNoTag() bool { return t.hasNoPOSTag }

func (t *AnalyzedToken) SetNoPOSTag(noTag bool) { t.hasNoPOSTag = noTag }

// String ports toString: lemmaOrToken + '/' + posTag (null prints as "null" like Java).
func (t *AnalyzedToken) String() string {
	pos := "null"
	if t.posTag != nil {
		pos = *t.posTag
	}
	return t.lemmaOrToken + "/" + pos
}

// Equals ports equals (not == on pointers).
func (t *AnalyzedToken) Equals(o *AnalyzedToken) bool {
	if o == nil {
		return false
	}
	if t == o {
		return true
	}
	if t.token != o.token || t.whitespaceBefore != o.whitespaceBefore {
		return false
	}
	if !strPtrEq(t.posTag, o.posTag) || !strPtrEq(t.lemma, o.lemma) {
		return false
	}
	return true
}

func strPtrEq(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
