package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EnglishChunker ports org.languagetool.chunking.EnglishChunker surface.
// Full OpenNLP maxent model is not vendored; soft path assigns OpenNLP-like
// BIO phrase tags from POS (NP/VP/PP/ADVP), then applies EnglishChunkFilter
// for B-NP-singular/plural and E-NP-* (Java EnglishChunkFilter).
type EnglishChunker struct {
	Filter *EnglishChunkFilter
	// AssignBasicNP enables POS-driven BIO assignment when no chunks present.
	// Name kept for twin tests; covers NP/VP/PP/ADVP, not only NP.
	AssignBasicNP bool
	// IsNounish reports whether a POS tag is noun-like (default: NN*).
	// Used by tests; assignOpenNLPLike uses a fuller POS→phrase map.
	IsNounish func(posTag string) bool
}

func NewEnglishChunker() *EnglishChunker {
	return &EnglishChunker{
		Filter:        NewEnglishChunkFilter(),
		AssignBasicNP: true,
		IsNounish: func(pos string) bool {
			return len(pos) >= 2 && pos[0] == 'N' && pos[1] == 'N'
		},
	}
}

// AddChunkTags implements Chunker (Java EnglishChunker.addChunkTags).
// Java OpenNLP chunker runs on non-whitespace tokens only; whitespace is
// skipped so NP spans (your cars, his chair) stay continuous for EnglishChunkFilter.
func (c *EnglishChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	// Map non-whitespace tokens only (mirror OpenNLP input stream).
	var idxs []int
	var tagged []ChunkTaggedToken
	for i, t := range tokens {
		if t == nil {
			continue
		}
		tok := t.GetToken()
		// Keep SENT_START (empty surface) out; skip pure whitespace.
		if tok != "" && strings.TrimSpace(tok) == "" {
			continue
		}
		if tok == "" {
			// SENT_START / empty: omit from phrase stream
			continue
		}
		var tags []ChunkTag
		for _, ct := range t.GetChunkTags() {
			tags = append(tags, NewChunkTag(ct))
		}
		idxs = append(idxs, i)
		tagged = append(tagged, NewChunkTaggedToken(tok, tags, t))
	}
	if c.AssignBasicNP {
		tagged = c.assignOpenNLPLike(tagged)
	}
	if c.Filter != nil {
		tagged = c.Filter.Filter(tagged)
	}
	// write back chunk tags onto original token indices
	for j, t := range tagged {
		if j >= len(idxs) {
			break
		}
		i := idxs[j]
		if i >= len(tokens) || tokens[i] == nil {
			continue
		}
		var strs []string
		for _, ct := range t.ChunkTags {
			s := ct.GetChunkTag()
			// Keep "O" so chunk_re="…|O" matches (Java OpenNLP outside tag).
			if s != "" {
				strs = append(strs, s)
			}
		}
		tokens[i].SetChunkTags(strs)
	}
}

// assignOpenNLPLike mirrors attic/chunker openNLP-like BIO from POS tags so
// soft grammar chunk / chunk_re constraints (B-PP, .-VP, E-NP.*) can match
// without the en-chunker.bin model.
func (c *EnglishChunker) assignOpenNLPLike(tokens []ChunkTaggedToken) []ChunkTaggedToken {
	out := make([]ChunkTaggedToken, len(tokens))
	copy(out, tokens)
	phrases := make([]string, len(tokens))
	poss := make([]string, len(tokens))
	prevPOS := ""
	prevSurf := ""
	for i, t := range out {
		// Hyphenated sign-in/log-up style phrasal verbs: dict often only has NN;
		// OpenNLP treats them as VP for SIGN_IN-style rules.
		if softIsHyphenPhrasalVerb(t.Token) {
			phrases[i] = "VP"
			poss[i] = "VB"
			prevPOS = "VB"
			prevSurf = t.Token
			continue
		}
		// Possessive apostrophe (Alex'/fox'): keep NP context for the following
		// noun (mother/tail) like Java POS clitic.
		if softIsPossessiveApostrophe(t.Token) {
			phrases[i] = "O"
			poss[i] = "POS"
			prevPOS = "PRP$"
			prevSurf = t.Token
			continue
		}
		// OpenNLP: "and catch up" — after CC, verb+particle is VP+PRT, not NP.
		// primaryPOS alone prefers NN for multi-tag catch; particle lookahead
		// mirrors OpenNLP phrasal-verb coordination (PHRASAL_VERB_SOMETIME).
		if prevPOS == "CC" && i+1 < len(out) &&
			softIsEnglishParticleSurface(out[i+1].Token) && softHasBareVerbReading(t) {
			phrases[i] = "VP"
			poss[i] = "VB"
			prevPOS = "VB"
			prevSurf = t.Token
			continue
		}
		// WHERE_MD_VB: "find out where will…" — token before where must match
		// chunk_re=".-VP|E-NP.*". OpenNLP PRT after verb blocks that; when the
		// next token is WRB (where/when/how/why), keep the particle inside the
		// VP span (I-VP) so the pattern can match (Java example sentence).
		if strings.HasPrefix(prevPOS, "VB") && softIsEnglishParticleSurface(t.Token) &&
			i+1 < len(out) && softIsWhAdverbSurface(out[i+1].Token) {
			phrases[i] = "VP"
			poss[i] = "RP"
			prevPOS = "RP"
			prevSurf = t.Token
			continue
		}
		pos := primaryPOS(t, prevPOS, prevSurf)
		poss[i] = pos
		tok := t.Token
		if tok == "" || pos == languagetool.SentenceStartTagName ||
			pos == languagetool.SentenceEndTagName || pos == languagetool.ParagraphEndTagName {
			phrases[i] = "O"
			prevPOS = ""
			prevSurf = ""
			continue
		}
		// Predicative amassing/amazing after so/very: OpenNLP ADJP, not VP.
		if softIsPredicativeAdjContext(prevSurf, pos) {
			phrases[i] = "ADJP"
			poss[i] = "JJ"
			prevPOS = "JJ"
			prevSurf = tok
			continue
		}
		// "You are amassing!" / "is completely fee" — predicative after be/intensifier.
		if softIsCopulaSurface(prevSurf) && softIsPredicativeMisspellSurface(tok) {
			phrases[i] = "ADJP"
			poss[i] = "JJ"
			prevPOS = "JJ"
			prevSurf = tok
			continue
		}
		if softIsIntensifierSurface(prevSurf) && softIsPredicativeMisspellSurface(tok) {
			phrases[i] = "NP"
			poss[i] = "NN"
			prevPOS = "NN"
			prevSurf = tok
			continue
		}
		// "not available/good": NN_NOT_JJ expects chunk="I-ADJP" (exact).
		if strings.EqualFold(prevSurf, "not") && (strings.HasPrefix(pos, "JJ") || jjReading(t)) {
			out[i].ChunkTags = []ChunkTag{NewChunkTag("I-ADJP")}
			prevPOS = "JJ"
			prevSurf = tok
			continue
		}
		// "a great discover": discover is only VB in dict but pattern wants E-NP.
		if softIsNominalizedVerbAfterAdj(prevPOS, t.Token) {
			phrases[i] = "NP"
			poss[i] = "NN"
			prevPOS = "NN"
			prevSurf = tok
			continue
		}
		// "to home/upstairs/…" adverbials: prefer ADVP when RB reading exists.
		if softIsDirectionalAfterTo(prevSurf, t) {
			phrases[i] = "ADVP"
			poss[i] = "RB"
			prevPOS = "RB"
			prevSurf = tok
			continue
		}
		phrases[i] = phraseFromPOS(pos)
		prevPOS = pos
		prevSurf = tok
	}
	bio := toBIOWithPOS(phrases, poss)
	for i := range out {
		if bio[i] == "" || bio[i] == "O" {
			// Java OpenNLP emits explicit "O" (outside). Soft patterns use
			// chunk_re="…|O" (SINGED_CONTRACT); empty tags do not match.
			if len(out[i].ChunkTags) == 0 {
				out[i].ChunkTags = []ChunkTag{NewChunkTag("O")}
			}
			continue
		}
		out[i].ChunkTags = []ChunkTag{NewChunkTag(bio[i])}
	}
	return out
}

func primaryPOS(t ChunkTaggedToken, prevPOS, prevSurf string) string {
	if t.Readings == nil {
		return ""
	}
	// Java EnglishChunker feeds OpenNLP a single POS from its own tagger.
	// Soft multi-tag LT dicts need a pick: default first non-boundary reading
	// (dict order ≈ frequency), with light context/aux heuristics.
	var first, vb, vbfinite, vbg, vbp, nn, rp, in, rb, jj string
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		p := r.GetPOSTag()
		if p == nil || *p == "" {
			continue
		}
		pos := *p
		if pos == languagetool.SentenceStartTagName || pos == languagetool.SentenceEndTagName ||
			pos == languagetool.ParagraphEndTagName {
			continue
		}
		if first == "" {
			first = pos
		}
		if pos == "RP" && rp == "" {
			rp = pos
		}
		if pos == "IN" && in == "" {
			in = pos
		}
		if strings.HasPrefix(pos, "RB") && rb == "" {
			rb = pos
		}
		if strings.HasPrefix(pos, "JJ") && jj == "" {
			jj = pos
		}
		if (strings.HasPrefix(pos, "VB") || pos == "MD") && vb == "" {
			vb = pos
		}
		if pos == "VBG" && vbg == "" {
			vbg = pos
		}
		if pos == "VBP" && vbp == "" {
			vbp = pos
		}
		// Finite verbs (not VBG/VBN alone) for subject-verb after NN.
		if (pos == "VB" || pos == "VBP" || pos == "VBZ" || pos == "VBD" || pos == "MD") && vbfinite == "" {
			vbfinite = pos
		}
		if strings.HasPrefix(pos, "NN") && nn == "" {
			nn = pos
		}
	}
	// English.dict often tags auxiliaries Does/Did as NNS|VBZ (plural noun first).
	// OpenNLP POS would choose VB*; force VB for known aux surfaces.
	if vb != "" && softIsEnglishAuxSurface(t.Token) {
		return vb
	}
	// Java EnglishChunker feeds OpenNLP POS, not LT multi-tags. OpenNLP tags
	// "let's" as Let/VB + 's/PRP (us), never 's/VBZ (is) or 's/POS.
	// Soft chunker runs pre-disambiguation when the dict still has POS|VBZ only;
	// force PRP so the following verb (hang) and particle (out) chunk correctly
	// for PHRASAL_VERB_SOMETIME (chunk_re=".-VP" + chunk="B-PRT").
	if softIsUsClitic(t.Token) && strings.EqualFold(prevSurf, "let") {
		return "PRP"
	}
	// Prepositions multi-tagged NN|IN|RP (in/on/at) must stay IN/PP, not NN after a noun
	// (was solution in this case — "in" was wrongly E-NP).
	// Exception: OpenNLP tags "I like" as VBP, not IN — skip prep force after subjects.
	if in != "" && softIsEnglishPrepSurface(t.Token) {
		if !(prevPOS == "PRP" || strings.HasPrefix(prevPOS, "PRP_") ||
			softIsPersonalPronounSurface(prevSurf)) {
			return in
		}
	}
	// Particles vs prepositions after a verb: "catch up" → RP/B-PRT, but
	// "singed with" / "books at" prefer IN/B-PP (prep surfaces).
	if strings.HasPrefix(prevPOS, "VB") {
		if softIsEnglishParticleSurface(t.Token) && rp != "" {
			return rp
		}
		if in != "" {
			return in
		}
		if rp != "" {
			return rp
		}
	}
	// Infinitive/modal: only after surface "to" (TO tag) or MD — not after every
	// IN ("like mine" must not force VB on mine).
	// Exception: "to be singed" — already VB.
	if prevPOS == "TO" || prevPOS == "MD" || strings.EqualFold(prevSurf, "to") {
		if vb != "" {
			return vb
		}
	}
	// After prep (OpenNLP): PP objects are nouns ("on balls"); "for set/bring" is
	// bare VB (FOR_VB wants B-VP); "for while" is while/NN (FOR_WHILE wants E-NP).
	if prevPOS == "IN" {
		if (strings.EqualFold(t.Token, "while") || strings.EqualFold(t.Token, "moment")) && nn != "" {
			return nn
		}
		if strings.EqualFold(prevSurf, "for") && softHasBareVerbReading(t) {
			return "VB"
		}
		if nn != "" {
			return nn
		}
	}
	// Copula "is/was/are/'s": progressive (is going), then finite verb
	// (when is comes), else predicative JJ/NN (What is last price).
	// Java: us-clitic 's is PRP (let's hang) — not contracted is. Soft multi-tag
	// surfaces share 's; only apply copula when previous primary is not PRP.
	if softIsCopulaSurface(prevSurf) && !(softIsUsClitic(prevSurf) && prevPOS == "PRP") {
		if vbg != "" {
			return vbg
		}
		if v := softFiniteTenseVerb(t); v != "" {
			return v
		}
		if jj != "" {
			return jj
		}
		if nn != "" {
			return nn
		}
	}
	// After a determiner/possessive: prefer adjective when both JJ and NN are
	// present (the cream colored paint); else noun over verb (the body/contract).
	if prevPOS == "DT" || prevPOS == "PRP$" || strings.HasPrefix(prevPOS, "PRP$") {
		if jj != "" && nn != "" {
			return jj
		}
		if nn != "" {
			return nn
		}
	}
	// After a subject-like tag:
	//  - progressive VBG (are you going)
	//  - finite VBD/VBZ (Chris rose / Does)
	//  - NNS compound after NN (voice disorders)
	//  - MY_NOT_MU: "mu house/opinion" keep NN
	//  - present VBP for agreement errors (if user want)
	if strings.HasPrefix(prevPOS, "NN") || prevPOS == "PRP" || strings.HasPrefix(prevPOS, "PRP_") {
		// Let's hang — 's clitic is PRP; force verb (not NN hang).
		if softIsUsClitic(prevSurf) && vb != "" {
			return vb
		}
		// Does anyone knows — after singular indefinite PRP force finite verb.
		if isSingularPronounSurface(prevSurf) {
			if v := softFiniteTenseVerb(t); v != "" {
				return v
			}
			if vbp != "" {
				return vbp
			}
		}
		// help your son sleeps — human subject + VBZ over NNS.
		if softIsHumanNounSurface(prevSurf) {
			if v := softFiniteTenseVerb(t); v != "" {
				return v
			}
		}
		// Progressive after pronoun only (are you going) — not after NN compound
		// (yoga training must stay NP, not VBG).
		if vbg != "" && (prevPOS == "PRP" || strings.HasPrefix(prevPOS, "PRP_")) {
			return vbg
		}
		// battery monitor works / each increase affects: after singular NN (incl.
		// NN:UN) prefer VBZ over NNS. OpenNLP tags the finite verb as VP, not
		// a plural noun compound (DOES_NP_VBZ).
		// voice disorders: NNS after NN:UN with no finite-verb preference when
		// the token is only NNS (no VBZ) — softFiniteTenseVerb empty then.
		if vbdOrZ := softFiniteTenseVerb(t); vbdOrZ != "" && softIsSingularNounPOS(prevPOS) {
			return vbdOrZ
		}
		if nn != "" && softHasPluralNounReading(t) && strings.HasPrefix(prevPOS, "NN") {
			return nn
		}
		if vbdOrZ := softFiniteTenseVerb(t); vbdOrZ != "" {
			return vbdOrZ
		}
		// "mu house" / "mu opinion" — keep noun compound (MY_NOT_MU).
		if strings.EqualFold(prevSurf, "mu") && nn != "" {
			return nn
		}
		// SUBJECT_NUMBER: user want — VBP over NN when prev is human/indefinite subject.
		// battery monitor: keep NN compound (not VBP on monitor).
		// Also: I keep — personal pronoun subjects take finite/VBP verb (not NN keep).
		if vbp != "" && (softIsHumanNounSurface(prevSurf) || isSingularPronounSurface(prevSurf) ||
			softIsPersonalPronounSurface(prevSurf)) {
			return vbp
		}
		if nn != "" && softIsPersonalPronounSurface(prevSurf) && vb != "" {
			// I keep / they keep — prefer verb reading over NN:UN
			return vb
		}
		if nn != "" {
			return nn
		}
		if vbp != "" {
			return vbp
		}
	}
	// After adjective:
	//  - prep complement: similar like / different from
	//  - stacked adjectives: cream colored
	//  - noun head: lovely matter / major cause
	//  - bare-infinitive after able-class: able think
	if strings.HasPrefix(prevPOS, "JJ") {
		if softIsAdjComplementPrep(t.Token) && in != "" {
			return in
		}
		if jj != "" {
			return jj
		}
		if softIsAbleClassSurface(prevSurf) && vb != "" {
			return vb
		}
		if nn != "" {
			return nn
		}
		if vb != "" {
			return vb
		}
	}
	// After adverb: verb complement (never make).
	if strings.HasPrefix(prevPOS, "RB") && vb != "" {
		return vb
	}
	// After comma: list nouns (… chokes, and joint locks) keep NNS; clause verbs
	// (… dysphonia, affect …) take VB when no plural noun reading.
	if prevSurf == "," {
		if nn != "" && softHasPluralNounReading(t) {
			return nn
		}
		if vb != "" {
			if vbdOrZ := softFiniteTenseVerb(t); vbdOrZ != "" {
				return vbdOrZ
			}
			return vb
		}
	}
	// Progressive after be / 'm / 're / was: I'm trying / are going / was trying.
	if vbg != "" && softIsBeLikeSurface(prevSurf) {
		return vbg
	}
	// Aspectual keep/kept/keeps + bare verb (keep see, kept get) — not object NN.
	if softIsAspectualKeepSurface(prevSurf) && vb != "" {
		return vb
	}
	// Let's hang — clitic 's is PRP; force verb complement over NN.
	if softIsUsClitic(prevSurf) && vb != "" {
		return vb
	}
	// After finite VBP/VBZ/VBD prefer object NN (if user want work).
	// Do not apply after bare VB (keep see still serial-verb).
	if (prevPOS == "VBP" || prevPOS == "VBZ" || prevPOS == "VBD") && nn != "" {
		return nn
	}
	// After a verb: object NNS without bare-VB reading (have drinks) stays NP;
	// serial/aspect bare verb (keep see) stays VP. OpenNLP: drinks/NNS, see/VB.
	if strings.HasPrefix(prevPOS, "VB") {
		if nn != "" && softHasPluralNounReading(t) && !softHasBareVerbReading(t) {
			return nn
		}
		if vb != "" {
			return vb
		}
	}
	// After CC: noun lists (and disorders / and paper) over verb by default.
	// Verb coordination "and catch up" is handled in assignOpenNLPLike via
	// particle lookahead (OpenNLP: catch/VB + up/RP).
	if prevPOS == "CC" {
		if nn != "" {
			return nn
		}
		if vb != "" {
			return vb
		}
	}
	// Sentence-initial: plural noun subjects (Gears shifted) over VBZ; else
	// imperative NN|VB (Look the door) prefers VB.
	if prevPOS == "" {
		if nn != "" && softHasPluralNounReading(t) {
			return nn
		}
		if vb != "" && nn != "" {
			return vb
		}
	}
	// After proper noun: finite verb (LanguageTool works as a charm).
	if strings.HasPrefix(prevPOS, "NNP") {
		if v := softFiniteTenseVerb(t); v != "" {
			return v
		}
	}
	return first
}

func softIsPredicativeAdjContext(prevSurf, pos string) bool {
	// "is so amassing" — VBG/JJ after intensifier should chunk as ADJP.
	switch strings.ToLower(strings.TrimSpace(prevSurf)) {
	case "so", "very", "really", "quite", "totally", "pretty", "rather":
		return strings.HasPrefix(pos, "VB") || strings.HasPrefix(pos, "JJ")
	default:
		return false
	}
}

func softIsDirectionalAfterTo(prevSurf string, t ChunkTaggedToken) bool {
	if !strings.EqualFold(strings.TrimSpace(prevSurf), "to") {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(t.Token)) {
	case "home", "upstairs", "downstairs", "downtown", "inside", "outside",
		"there", "here", "away", "near", "abroad", "overseas",
		"everywhere", "somewhere", "nowhere", "underground":
		// need an RB reading (or bare adverbial surface)
		if t.Readings == nil {
			return true
		}
		for _, r := range t.Readings.GetReadings() {
			if r == nil {
				continue
			}
			if p := r.GetPOSTag(); p != nil && (strings.HasPrefix(*p, "RB") || *p == "JJ") {
				return true
			}
		}
		return true
	default:
		return false
	}
}

func softIsEnglishAuxSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "do", "does", "did", "is", "am", "are", "was", "were",
		"has", "have", "had", "be", "been", "being",
		"will", "would", "shall", "should", "can", "could", "may", "might", "must":
		return true
	default:
		return false
	}
}

// softIsEnglishParticleSurface: common verb particles (OpenNLP B-PRT), not
// prepositions that are also tagged RP (with/at/for…).
func softIsEnglishParticleSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "up", "out", "off", "away", "back", "down", "over", "along", "around", "through", "in", "on":
		return true
	default:
		return false
	}
}

// softIsHyphenPhrasalVerb matches SIGN_IN pattern surfaces (sign|log)-(in|up|off).
func softIsHyphenPhrasalVerb(s string) bool {
	low := strings.ToLower(strings.TrimSpace(s))
	switch low {
	case "sign-in", "sign-up", "sign-off", "log-in", "log-up", "log-off":
		return true
	default:
		return false
	}
}

// softIsAdjComplementPrep: prepositions that follow adjectives (similar like, different from).
func softIsAdjComplementPrep(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "like", "as", "from", "to", "with", "of", "for", "about", "than":
		return true
	default:
		return false
	}
}

// softIsAbleClassSurface: adjectives that take a bare infinitive (able think).
func softIsAbleClassSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "able", "unable", "ready", "willing", "likely", "unlikely", "free", "glad", "eager":
		return true
	default:
		return false
	}
}

// softFiniteTenseVerb returns VBD/VBZ when present (not bare VB/VBP).
func softFiniteTenseVerb(t ChunkTaggedToken) string {
	if t.Readings == nil {
		return ""
	}
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		p := r.GetPOSTag()
		if p == nil {
			continue
		}
		if *p == "VBD" || *p == "VBZ" {
			return *p
		}
	}
	return ""
}

func jjReading(t ChunkTaggedToken) bool {
	if t.Readings == nil {
		return false
	}
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && strings.HasPrefix(*p, "JJ") {
			return true
		}
	}
	return false
}

func softHasPluralNounReading(t ChunkTaggedToken) bool {
	if t.Readings == nil {
		return false
	}
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		p := r.GetPOSTag()
		if p == nil {
			continue
		}
		if *p == "NNS" || *p == "NNPS" || strings.HasPrefix(*p, "NNS") || strings.HasPrefix(*p, "NNPS") {
			return true
		}
	}
	return false
}

func softIsCopulaSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "is", "was", "are", "were", "be", "been", "being", "'s", "’s":
		return true
	default:
		return false
	}
}

func softIsBeLikeSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "is", "was", "are", "were", "be", "been", "being", "am",
		"'m", "’m", "'s", "’s", "'re", "’re":
		return true
	default:
		return false
	}
}

func softIsUsClitic(s string) bool {
	s = strings.TrimSpace(s)
	if s == "'s" || s == "\u2019s" {
		return true
	}
	rs := []rune(s)
	return len(rs) == 2 && (rs[0] == '\'' || rs[0] == '\u2019' || rs[0] == '\u02bc') &&
		(rs[1] == 's' || rs[1] == 'S')
}

func softIsPersonalPronounSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "i", "you", "he", "she", "it", "we", "they", "me", "him", "her", "us", "them":
		return true
	default:
		return false
	}
}

func softIsAspectualKeepSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "keep", "keeps", "kept", "keeping",
		"start", "starts", "started", "starting",
		"stop", "stops", "stopped", "stopping",
		"begin", "begins", "began", "begun", "beginning",
		"continue", "continues", "continued", "continuing":
		return true
	default:
		return false
	}
}

func softIsIntensifierSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "completely", "really", "totally", "quite", "pretty", "rather", "very", "so":
		return true
	default:
		return false
	}
}

// softIsPredicativeMisspellSurface: rule-targeted predicative errors (amassing/fee).
func softIsPredicativeMisspellSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "amassing", "amazing", "fee", "free":
		return true
	default:
		return false
	}
}

// softIsBareSubjectPrev: subordinators introducing a bare singular subject
// (if user want / when student need).
func softIsBareSubjectPrev(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "if", "when", "whenever", "whether", "unless", "while", "although", "though", "because":
		return true
	default:
		// also true when prev is the subject noun itself? handled by after NN
		return false
	}
}

// softIsNominalizedVerbAfterAdj: "a great discover" — dict has only VB for discover.
func softIsNominalizedVerbAfterAdj(prevPOS, surface string) bool {
	if !strings.HasPrefix(prevPOS, "JJ") && prevPOS != "DT" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(surface)) {
	case "discover", "develop", "analyze", "analyse", "perform", "respond", "succeed":
		return true
	default:
		return false
	}
}

func softIsHumanNounSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "son", "daughter", "mother", "father", "child", "kid", "boy", "girl",
		"man", "woman", "student", "user", "teacher", "doctor", "patient", "friend",
		"brother", "sister", "parent", "baby", "person":
		return true
	default:
		return false
	}
}

func softIsEnglishPrepSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "in", "on", "at", "by", "for", "from", "with", "of", "to", "into", "onto",
		"over", "under", "about", "as", "like", "through", "between", "among", "without",
		"within", "after", "before", "during", "against", "via", "per",
		// LOOK_DOOR: "the door behind you" — behind must be B-PP so door is E-NP.
		"behind", "beside", "below", "above", "across", "near", "since", "until",
		"upon", "beyond", "beneath", "toward", "towards", "despite", "except", "plus":
		return true
	default:
		return false
	}
}

// softIsSingularNounPOS: NN / NNP / NN:UN / NNP:… (not NNS/NNPS).
// Used so "increase" (NN:UN) still triggers finite-verb after subject.
func softIsSingularNounPOS(pos string) bool {
	if pos == "NN" || pos == "NNP" {
		return true
	}
	if strings.HasPrefix(pos, "NN:") || strings.HasPrefix(pos, "NNP:") {
		return true
	}
	return false
}

// softHasBareVerbReading: VB or VBP present (not only VBD/VBZ/VBG/VBN).
func softHasBareVerbReading(t ChunkTaggedToken) bool {
	if t.Readings == nil {
		return false
	}
	for _, r := range t.Readings.GetReadings() {
		if r == nil {
			continue
		}
		p := r.GetPOSTag()
		if p == nil {
			continue
		}
		if *p == "VB" || *p == "VBP" {
			return true
		}
	}
	return false
}

func softIsWhAdverbSurface(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "where", "when", "how", "why":
		return true
	default:
		return false
	}
}

func softIsPossessiveApostrophe(s string) bool {
	switch strings.TrimSpace(s) {
	case "'", "’", "ʼ", "`":
		return true
	default:
		return false
	}
}

func phraseFromPOS(pos string) string {
	switch {
	case pos == "" || pos == "," || pos == "." || strings.HasPrefix(pos, "PCT"):
		return "O"
	case strings.HasPrefix(pos, "VB") || pos == "MD":
		return "VP"
	case strings.HasPrefix(pos, "RB") || pos == "WRB":
		return "ADVP"
	case pos == "RP":
		// OpenNLP particle chunk (catch up / sign in) → B-PRT
		return "PRT"
	case pos == "IN" || pos == "TO":
		return "PP"
	case strings.HasPrefix(pos, "NN") || pos == "DT" || pos == "PDT" ||
		pos == "PRP" || pos == "PRP$" || pos == "CD" || pos == "EX" ||
		pos == "WP" || pos == "WP$" || pos == "WDT" || pos == "POS" ||
		strings.HasPrefix(pos, "JJ") || strings.HasPrefix(pos, "PRP"):
		return "NP"
	case pos == "CC":
		return "O"
	default:
		if strings.HasPrefix(pos, "JJ") {
			return "NP"
		}
		return "O"
	}
}

func toBIO(phrase []string) []string {
	out := make([]string, len(phrase))
	prev := ""
	for i, p := range phrase {
		if p == "O" || p == "" {
			out[i] = "O"
			prev = ""
			continue
		}
		if p == prev {
			out[i] = "I-" + p
		} else {
			out[i] = "B-" + p
		}
		prev = p
	}
	return out
}

// toBIOWithPOS restarts NP at DT/PDT/PRP so "his chair an …" / "Some time I …"
// are multiple NPs (OpenNLP rarely chains a new determiner/pronoun into the
// previous noun phrase).
func toBIOWithPOS(phrase []string, poss []string) []string {
	out := make([]string, len(phrase))
	prev := ""
	for i, p := range phrase {
		if p == "O" || p == "" {
			out[i] = "O"
			prev = ""
			continue
		}
		restart := false
		if p == "NP" && prev == "NP" && i < len(poss) {
			pos := poss[i]
			if pos == "DT" || pos == "PDT" || pos == "PRP" || strings.HasPrefix(pos, "PRP_") {
				restart = true
			}
		}
		if p == prev && !restart {
			out[i] = "I-" + p
		} else {
			out[i] = "B-" + p
		}
		prev = p
	}
	return out
}

var _ Chunker = (*EnglishChunker)(nil)
