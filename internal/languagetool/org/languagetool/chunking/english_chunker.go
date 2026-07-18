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
			if s != "" && s != "O" {
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
	for i, t := range out {
		// Hyphenated sign-in/log-up style phrasal verbs: dict often only has NN;
		// OpenNLP treats them as VP for SIGN_IN-style rules.
		if softIsHyphenPhrasalVerb(t.Token) {
			phrases[i] = "VP"
			poss[i] = "VB"
			prevPOS = "VB"
			continue
		}
		pos := primaryPOS(t, prevPOS)
		poss[i] = pos
		tok := t.Token
		if tok == "" || pos == languagetool.SentenceStartTagName ||
			pos == languagetool.SentenceEndTagName || pos == languagetool.ParagraphEndTagName {
			phrases[i] = "O"
			prevPOS = ""
			continue
		}
		phrases[i] = phraseFromPOS(pos)
		prevPOS = pos
	}
	bio := toBIOWithPOS(phrases, poss)
	for i := range out {
		if bio[i] == "" || bio[i] == "O" {
			// keep any pre-existing tags
			continue
		}
		out[i].ChunkTags = []ChunkTag{NewChunkTag(bio[i])}
	}
	return out
}

func primaryPOS(t ChunkTaggedToken, prevPOS string) string {
	if t.Readings == nil {
		return ""
	}
	// Java EnglishChunker feeds OpenNLP a single POS from its own tagger.
	// Soft multi-tag LT dicts need a pick: default first non-boundary reading
	// (dict order ≈ frequency), with light context/aux heuristics.
	var first, vb, vbfinite, nn, rp, in string
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
		if (strings.HasPrefix(pos, "VB") || pos == "MD") && vb == "" {
			vb = pos
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
	// Infinitive/modal: "to tell", "can run". Soft dict may list IN before TO.
	if prevPOS == "TO" || prevPOS == "IN" || prevPOS == "MD" {
		if vb != "" {
			return vb
		}
	}
	// After a determiner/possessive, prefer noun over verb (the body/contract).
	if prevPOS == "DT" || prevPOS == "PRP$" || strings.HasPrefix(prevPOS, "PRP$") {
		if nn != "" {
			return nn
		}
	}
	// After a subject-like tag, prefer a *finite* verb (Chris rose / they run).
	// Do not prefer VBG over NN ("yoga training" must stay NP).
	if strings.HasPrefix(prevPOS, "NN") || prevPOS == "PRP" || strings.HasPrefix(prevPOS, "PRP_") {
		if vbfinite != "" {
			return vbfinite
		}
	}
	// After a verb, prefer another verb for serial/aspect complements (keep see,
	// going tell) when both NN and VB exist.
	if strings.HasPrefix(prevPOS, "VB") && vb != "" {
		return vb
	}
	// After CC prefer verb (and catch up).
	if prevPOS == "CC" && vb != "" {
		return vb
	}
	// Sentence-initial (or after O) imperative/content verbs: "Look the door"
	// is often NN|VB in the dict; OpenNLP prefers VB in imperative context.
	if prevPOS == "" && vb != "" && nn != "" {
		return vb
	}
	return first
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
