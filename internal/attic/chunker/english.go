// Package chunker implements English NP/VP-style chunk tags for LanguageTool rules.
//
// This is a POS-driven heuristic approximating OpenNLP + EnglishChunkFilter output
// (B-NP / I-NP / B-VP / … refined to B-NP-singular etc.). It is not a full OpenNLP
// maxent port; rules depending on exact OpenNLP boundaries may still diverge.
package chunker

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/attic/pipeline"
)

// English assigns chunk tags to non-whitespace tokens (in place).
// Tokens should already be POS-tagged (Readings filled).
func English(tokens []pipeline.Token) {
	if len(tokens) == 0 {
		return
	}
	// OpenNLP-like BIO tags from POS
	raw := make([]string, len(tokens))
	for i, t := range tokens {
		if t.Text == "SENT_START" || t.Whitespace {
			raw[i] = "O"
			continue
		}
		raw[i] = openNLPLike(t)
	}
	// Convert consecutive same phrase types to B/I
	bio := toBIO(raw)
	// EnglishChunkFilter: refine NP with singular/plural and E- tags
	refined := filterNP(tokens, bio)
	for i := range tokens {
		if refined[i] != "" && refined[i] != "O" {
			tokens[i].ChunkTags = []string{refined[i]}
		} else if bio[i] != "" && bio[i] != "O" {
			tokens[i].ChunkTags = []string{bio[i]}
		}
	}
}

func openNLPLike(t pipeline.Token) string {
	pos := primaryPOS(t)
	switch {
	case pos == "" || pos == "SENT_START" || pos == "PCT" || pos == ",":
		return "O"
	case isVerb(pos):
		return "VP"
	case isAdv(pos):
		return "ADVP"
	case isPrep(pos):
		return "PP"
	case isNounPhrasePOS(pos):
		return "NP"
	case pos == "CC":
		return "O"
	default:
		// adjectives often inside NP — mark as NP for continuation
		if strings.HasPrefix(pos, "JJ") {
			return "NP"
		}
		return "O"
	}
}

func primaryPOS(t pipeline.Token) string {
	if len(t.Readings) == 0 {
		return ""
	}
	// Prefer noun/verb over other if multiple
	for _, r := range t.Readings {
		if strings.HasPrefix(r.POS, "NN") || strings.HasPrefix(r.POS, "VB") {
			return r.POS
		}
	}
	return t.Readings[0].POS
}

func isVerb(pos string) bool {
	return strings.HasPrefix(pos, "VB") || pos == "MD"
}

func isAdv(pos string) bool {
	return strings.HasPrefix(pos, "RB")
}

func isPrep(pos string) bool {
	return pos == "IN" || pos == "TO"
}

func isNounPhrasePOS(pos string) bool {
	return strings.HasPrefix(pos, "NN") || pos == "DT" || pos == "PDT" ||
		pos == "PRP" || pos == "PRP$" || pos == "CD" || pos == "EX" ||
		pos == "WP" || pos == "WP$" || pos == "WDT" ||
		strings.HasPrefix(pos, "JJ") || pos == "POS"
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

// filterNP mirrors EnglishChunkFilter: B-NP-singular/plural and E- tags.
func filterNP(tokens []pipeline.Token, bio []string) []string {
	out := make([]string, len(bio))
	copy(out, bio)
	i := 0
	for i < len(bio) {
		if bio[i] != "B-NP" {
			i++
			continue
		}
		// find end of NP
		end := i
		for end+1 < len(bio) && (bio[end+1] == "I-NP" || bio[end+1] == "B-NP") {
			// treat consecutive NP as one; OpenNLP uses I-NP for continuation
			if bio[end+1] == "B-NP" {
				break
			}
			end++
		}
		// for I-NP chain
		j := i + 1
		for j < len(bio) && bio[j] == "I-NP" {
			j++
		}
		end = j - 1

		plural := false
		for k := i; k <= end; k++ {
			if hasPluralNoun(tokens[k]) {
				plural = true
				break
			}
		}
		kind := "singular"
		if plural {
			kind = "plural"
		}
		out[i] = "B-NP-" + kind
		for k := i + 1; k <= end; k++ {
			out[k] = "I-NP-" + kind
		}
		// mark end
		out[end] = "E-NP-" + kind
		if i == end {
			// single token NP: both B and E — LT may have only one tag list with both?
			// EnglishChunkFilter adds E- as additional tag on end; start gets B-
			// For single token both beginning and end:
			out[i] = "B-NP-" + kind
			// store both by joining later — Token has []string; set both
			// We'll handle multi-tag in assign below
		}
		i = end + 1
	}
	// second pass for single-token: add E- as second chunk tag
	// handled when assigning to tokens in English() — refine single NP tokens
	for i := range out {
		if strings.HasPrefix(out[i], "B-NP-") {
			// if next is not I-NP of same, it's also end
			kind := strings.TrimPrefix(out[i], "B-NP-")
			if i+1 >= len(out) || !strings.HasPrefix(out[i+1], "I-NP-") && !strings.HasPrefix(out[i+1], "E-NP-") {
				// single-token NP — EnglishChunkFilter adds both B and E tags
				// represent as B- only here; pattern match will check either
				_ = kind
			}
		}
	}
	// Fix: for single token NP, English() should set both B and E tags
	// Do that in English() after filterNP
	return out
}

func hasPluralNoun(t pipeline.Token) bool {
	for _, r := range t.Readings {
		if r.POS == "NNS" || r.POS == "NNPS" {
			return true
		}
	}
	return false
}

// EnglishWithMultiTags assigns possibly multiple chunk tags (B+E on one token).
func EnglishWithMultiTags(tokens []pipeline.Token) {
	if len(tokens) == 0 {
		return
	}
	raw := make([]string, len(tokens))
	for i, t := range tokens {
		if t.Text == "SENT_START" || t.Whitespace {
			raw[i] = "O"
			continue
		}
		raw[i] = openNLPLike(t)
	}
	bio := toBIO(raw)

	// Identify NP spans
	type span struct {
		start, end int
		plural     bool
	}
	var spans []span
	for i := 0; i < len(bio); {
		if bio[i] != "B-NP" {
			i++
			continue
		}
		end := i
		for end+1 < len(bio) && bio[end+1] == "I-NP" {
			end++
		}
		pl := false
		for k := i; k <= end; k++ {
			if hasPluralNoun(tokens[k]) {
				pl = true
				break
			}
		}
		spans = append(spans, span{i, end, pl})
		i = end + 1
	}

	// default: copy bio for non-NP
	for i := range tokens {
		tokens[i].ChunkTags = nil
		if bio[i] != "O" && !strings.HasSuffix(strings.TrimPrefix(bio[i], "B-"), "NP") &&
			!strings.HasSuffix(strings.TrimPrefix(bio[i], "I-"), "NP") &&
			bio[i] != "B-NP" && bio[i] != "I-NP" {
			tokens[i].ChunkTags = []string{bio[i]}
		}
	}
	for _, sp := range spans {
		kind := "singular"
		if sp.plural {
			kind = "plural"
		}
		for k := sp.start; k <= sp.end; k++ {
			var tags []string
			if k == sp.start {
				tags = append(tags, "B-NP-"+kind)
				// also keep generic B-NP for rules that want exact B-NP
				tags = append(tags, "B-NP")
			}
			if k > sp.start && k < sp.end {
				tags = append(tags, "I-NP-"+kind, "I-NP")
			}
			if k == sp.end {
				tags = append(tags, "E-NP-"+kind)
				if k != sp.start {
					tags = append(tags, "I-NP-"+kind, "I-NP")
				} else {
					// single token is also end
				}
			}
			// OpenNLP original B-NP/I-NP already added for start/mid
			tokens[k].ChunkTags = unique(tags)
		}
	}
}

func unique(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// MatchChunk reports whether token has chunk tag equal to want (or any if want empty).
func MatchChunk(tags []string, want string) bool {
	if want == "" {
		return true
	}
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

// MatchChunkRe reports whether any chunk tag matches the regex.
func MatchChunkRe(tags []string, re *regexp.Regexp) bool {
	if re == nil {
		return true
	}
	for _, t := range tags {
		if re.MatchString(t) {
			return true
		}
	}
	return false
}
