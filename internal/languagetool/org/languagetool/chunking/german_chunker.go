package chunking

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// GermanChunker ports org.languagetool.chunking.GermanChunker.
// REGEXES1 and REGEXES2 run Java OpenRegex pattern strings via CompileOpenRegex
// (knowitall sequence matcher + TokenExpressionFactory / LogicExpression).
// No POS→BIO fallback: Java only runs REGEXES1 then REGEXES2 (tokens start as O).
type GermanChunker struct{}

func NewGermanChunker() *GermanChunker {
	return &GermanChunker{}
}

// contentTokens filters whitespace / sentence markers like Java getBasicChunks.
func germanContentTokens(tokens []*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	var content []*languagetool.AnalyzedTokenReadings
	for _, t := range tokens {
		if t == nil || t.IsWhitespace() || t.IsSentenceStart() || t.GetToken() == "" {
			continue
		}
		content = append(content, t)
	}
	return content
}

// Java GermanChunker.REGEXES1 OpenRegex patterns (order matters; later overwrite).
var germanRegexes1 = []string{
	`(<posre=^ART.*>|<pos=PRO>)? <pos=ADV>* <pos=PA2>* <pos=ADJ>* <pos=SUB>+`,
	`<pos=SUB> (<und|oder>|(<bzw> <.>)) <pos=SUB>`,
	`<pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=PA2> <pos=SUB>`,
	`<pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=ADJ> <pos=SUB>`,
	`<posre=^ART.*> <pos=ADV>* <pos=ADJ>* <regexCS=[A-ZÖÄÜ][a-zöäü]+>`,
	`<pos=PRO>? <pos=ZAL> <pos=SUB>`,
	`<Herr|Herrn|Frau> <pos=EIG>+`,
	`<Herr|Herrn|Frau> <regexCS=[A-ZÖÄÜ][a-zöäü-]+>+`,
	`<der>`,
}

// applyRegexes1 runs Java REGEXES1 via OpenRegex into tags (B-NP/I-NP).
// Java doApplyRegex *adds* B-NP/I-NP (overwrite=false for all REGEXES1) and drops O —
// does not replace prior tags. Overlapping matches may leave both B-NP and I-NP.
func applyRegexes1(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	if len(content) == 0 {
		return
	}
	factory := NewChunkTokenFactory(false)
	// Side list like Java getBasicChunks (chunk starts as O; REGEXES1 uses POS/surface).
	toks := make([]ChunkTaggedToken, len(content))
	for i, t := range content {
		toks[i] = NewChunkTaggedToken(t.GetToken(), []ChunkTag{NewChunkTag("O")}, t)
	}
	for _, pat := range germanRegexes1 {
		re := CompileOpenRegex(ExpandGermanChunkSyntax(pat), factory)
		for _, m := range re.FindAll(toks) {
			for i := m.Start; i < m.End; i++ {
				newTag := "I-NP"
				if i == m.Start {
					newTag = "B-NP"
				}
				// Java: addAll existing, add newTag if missing, remove O.
				newTags := make([]ChunkTag, 0, len(toks[i].ChunkTags)+1)
				newTags = append(newTags, toks[i].ChunkTags...)
				has := false
				for _, ct := range newTags {
					if ct.GetChunkTag() == newTag {
						has = true
						break
					}
				}
				if !has {
					newTags = append(newTags, NewChunkTag(newTag))
				}
				filtered := make([]ChunkTag, 0, len(newTags))
				for _, ct := range newTags {
					if ct.GetChunkTag() != "O" {
						filtered = append(filtered, ct)
					}
				}
				toks[i] = NewChunkTaggedToken(toks[i].Token, filtered, toks[i].Readings)
			}
		}
	}
	// Sync parallel tags slice from working token list.
	for i := range toks {
		var strs []string
		for _, ct := range toks[i].ChunkTags {
			if s := ct.GetChunkTag(); s != "" && s != "O" {
				strs = append(strs, s)
			}
		}
		tags[i] = strs
	}
}

// GetBasicChunks ports GermanChunker.getBasicChunks — REGEXES1 only (OpenNLP-like B-NP/I-NP).
// Tokens without a REGEXES1 hit get chunk tag "O" (Java singleton O before assign).
// Does not mutate the input readings' chunk tags (Java builds a side list).
func (c *GermanChunker) GetBasicChunks(tokens []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	if c == nil || len(tokens) == 0 {
		return nil
	}
	content := germanContentTokens(tokens)
	if len(content) == 0 {
		return nil
	}
	tags := make([][]string, len(content))
	applyRegexes1(content, tags)
	out := make([]ChunkTaggedToken, 0, len(content))
	for i, t := range content {
		chunkTags := []ChunkTag{NewChunkTag("O")}
		if len(tags[i]) > 0 {
			chunkTags = make([]ChunkTag, 0, len(tags[i]))
			for _, s := range tags[i] {
				chunkTags = append(chunkTags, NewChunkTag(s))
			}
		}
		out = append(out, NewChunkTaggedToken(t.GetToken(), chunkTags, t))
	}
	return out
}

func (c *GermanChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	content := germanContentTokens(tokens)
	if len(content) == 0 {
		return
	}

	// Parallel tags; empty means O / untagged.
	tags := make([][]string, len(content))
	applyRegexes1(content, tags)

	// REGEXES2: full Java OpenRegex pattern list (NPS/NPP/PP) on B-NP/I-NP.
	// No invent POS→BIO fill-in for untagged tokens (Java leaves them as O).
	applyRegexes2(content, tags)

	for i, t := range content {
		if len(tags[i]) > 0 {
			t.SetChunkTags(tags[i])
		}
	}
}

// applyRegexes2 ports Java GermanChunker.REGEXES2 via OpenRegex (full pattern list).
// FILTER_TAGS overwrite removes PP/NPP/NPS before adding the new phrase tag (Java).
func applyRegexes2(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	if len(content) == 0 {
		return
	}
	factory := NewChunkTokenFactory(false)
	// Working token list mirrors Java List<ChunkTaggedToken> updated after each pattern.
	toks := make([]ChunkTaggedToken, len(content))
	for i, t := range content {
		cts := []ChunkTag{NewChunkTag("O")}
		if len(tags[i]) > 0 {
			cts = make([]ChunkTag, 0, len(tags[i]))
			for _, s := range tags[i] {
				cts = append(cts, NewChunkTag(s))
			}
		}
		toks[i] = NewChunkTaggedToken(t.GetToken(), cts, t)
	}
	filterTags := map[string]bool{"PP": true, "NPP": true, "NPS": true}
	for _, spec := range germanRegexes2 {
		pat := ExpandGermanChunkSyntax(spec.pattern)
		re := CompileOpenRegex(pat, factory)
		matches := re.FindAll(toks)
		for _, m := range matches {
			tagName := spec.phrase.tagName()
			if tagName == "" {
				continue
			}
			for i := m.Start; i < m.End; i++ {
				newTags := make([]ChunkTag, 0, len(toks[i].ChunkTags)+1)
				for _, ct := range toks[i].ChunkTags {
					s := ct.GetChunkTag()
					if spec.overwrite && filterTags[s] {
						continue
					}
					newTags = append(newTags, ct)
				}
				// add phrase tag if missing; drop O
				has := false
				for _, ct := range newTags {
					if ct.GetChunkTag() == tagName {
						has = true
						break
					}
				}
				if !has {
					newTags = append(newTags, NewChunkTag(tagName))
				}
				filtered := make([]ChunkTag, 0, len(newTags))
				for _, ct := range newTags {
					if ct.GetChunkTag() != "O" {
						filtered = append(filtered, ct)
					}
				}
				toks[i] = NewChunkTaggedToken(toks[i].Token, filtered, toks[i].Readings)
			}
		}
	}
	// Write back to parallel tags slice
	for i := range toks {
		var strs []string
		for _, ct := range toks[i].ChunkTags {
			if s := ct.GetChunkTag(); s != "" && s != "O" {
				strs = append(strs, s)
			}
		}
		tags[i] = strs
	}
}

// firstPOS returns the first non-empty POS tag (shared by RussianChunker etc.).
func firstPOS(t *languagetool.AnalyzedTokenReadings) string {
	if t == nil {
		return ""
	}
	for _, r := range t.GetReadings() {
		if r == nil {
			continue
		}
		if p := r.GetPOSTag(); p != nil && *p != "" {
			return *p
		}
	}
	return ""
}
