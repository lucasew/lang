package chunking

import (
	"fmt"
	"strings"

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

// FILTER_TAGS ports Java GermanChunker.FILTER_TAGS (overwrite mode removes these).
var germanFilterTags = map[string]bool{"PP": true, "NPP": true, "NPS": true}

// debug mirrors Java GermanChunker.debug (SetDebug / isDebug).
var germanChunkerDebug bool

// SetDebug ports GermanChunker.setDebug (deprecated for internal use only).
func SetGermanChunkerDebug(debugMode bool) { germanChunkerDebug = debugMode }

// IsGermanChunkerDebug ports GermanChunker.isDebug.
func IsGermanChunkerDebug() bool { return germanChunkerDebug }

// germanContentTokens filters like Java getBasicChunks: non-whitespace tokens.
// Also skips pure SENT_START markers (empty/sentence-start), which Java leaves
// as O without matching content patterns in practice.
func germanContentTokens(tokens []*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	var content []*languagetool.AnalyzedTokenReadings
	for _, t := range tokens {
		if t == nil || t.IsWhitespace() {
			continue
		}
		// Pure sentence-start marker (no surface): skip like empty non-content.
		if t.IsSentenceStart() && t.GetToken() == "" {
			continue
		}
		if t.GetToken() == "" {
			continue
		}
		content = append(content, t)
	}
	return content
}

// Java GermanChunker.REGEXES1 OpenRegex patterns (order matters; overwrite=false, NP → B-NP/I-NP).
var germanRegexes1 = []germanRegex2{
	{`(<posre=^ART.*>|<pos=PRO>)? <pos=ADV>* <pos=PA2>* <pos=ADJ>* <pos=SUB>+`, phraseNP, false},
	{`<pos=SUB> (<und|oder>|(<bzw> <.>)) <pos=SUB>`, phraseNP, false},
	{`<pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=PA2> <pos=SUB>`, phraseNP, false},
	{`<pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=ADJ> <pos=SUB>`, phraseNP, false},
	{`<posre=^ART.*> <pos=ADV>* <pos=ADJ>* <regexCS=[A-ZÖÄÜ][a-zöäü]+>`, phraseNP, false},
	{`<pos=PRO>? <pos=ZAL> <pos=SUB>`, phraseNP, false},
	{`<Herr|Herrn|Frau> <pos=EIG>+`, phraseNP, false},
	{`<Herr|Herrn|Frau> <regexCS=[A-ZÖÄÜ][a-zöäü-]+>+`, phraseNP, false},
	{`<der>`, phraseNP, false},
}

// GetBasicChunks ports GermanChunker.getBasicChunks — REGEXES1 only (OpenNLP-like B-NP/I-NP).
// Tokens without a REGEXES1 hit keep chunk tag "O". Does not mutate input readings.
func (c *GermanChunker) GetBasicChunks(tokens []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	if c == nil || len(tokens) == 0 {
		return nil
	}
	content := germanContentTokens(tokens)
	if len(content) == 0 {
		return nil
	}
	toks := initialChunkTaggedTokens(content)
	if germanChunkerDebug {
		fmt.Println("=============== CHUNKER INPUT ===============")
		fmt.Print(germanChunkDebugString(toks))
	}
	for _, regex := range germanRegexes1 {
		applyGermanRegex(regex, toks)
	}
	return toks
}

// AddChunkTags ports GermanChunker.addChunkTags: REGEXES1 then REGEXES2, assign to readings.
func (c *GermanChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	content := germanContentTokens(tokens)
	if len(content) == 0 {
		return
	}
	toks := initialChunkTaggedTokens(content)
	if germanChunkerDebug {
		fmt.Println("=============== CHUNKER INPUT ===============")
		fmt.Print(germanChunkDebugString(toks))
	}
	for _, regex := range germanRegexes1 {
		applyGermanRegex(regex, toks)
	}
	for _, regex := range germanRegexes2 {
		applyGermanRegex(regex, toks)
	}
	assignGermanChunksToReadings(toks)
}

func initialChunkTaggedTokens(content []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	toks := make([]ChunkTaggedToken, len(content))
	for i, t := range content {
		toks[i] = NewChunkTaggedToken(t.GetToken(), []ChunkTag{NewChunkTag("O")}, t)
	}
	return toks
}

// applyGermanRegex ports GermanChunker.apply / doApplyRegex for one REGEXES1/2 entry.
func applyGermanRegex(regex germanRegex2, toks []ChunkTaggedToken) {
	pat := ExpandGermanChunkSyntax(regex.pattern)
	re := CompileOpenRegex(pat, NewChunkTokenFactory(false))
	prevDebug := ""
	if germanChunkerDebug {
		prevDebug = germanChunkDebugString(toks)
	}
	matches := re.FindAll(toks)
	for _, m := range matches {
		for i := m.Start; i < m.End; i++ {
			token := toks[i]
			newChunkTags := make([]ChunkTag, 0, len(token.ChunkTags)+1)
			newChunkTags = append(newChunkTags, token.ChunkTags...)
			if regex.overwrite {
				filtered := make([]ChunkTag, 0, len(newChunkTags))
				for _, ct := range newChunkTags {
					if !germanFilterTags[ct.GetChunkTag()] {
						filtered = append(filtered, ct)
					}
				}
				newChunkTags = filtered
			}
			newTag := germanChunkTagForMatch(regex, m.Start, i)
			if newTag == "" {
				continue
			}
			has := false
			for _, ct := range newChunkTags {
				if ct.GetChunkTag() == newTag {
					has = true
					break
				}
			}
			if !has {
				newChunkTags = append(newChunkTags, NewChunkTag(newTag))
			}
			// remove O (Java: newChunkTags.remove(new ChunkTag("O")))
			final := make([]ChunkTag, 0, len(newChunkTags))
			for _, ct := range newChunkTags {
				if ct.GetChunkTag() != "O" {
					final = append(final, ct)
				}
			}
			toks[i] = NewChunkTaggedToken(token.Token, final, token.Readings)
		}
	}
	if germanChunkerDebug {
		debug := germanChunkDebugString(toks)
		if debug != prevDebug {
			fmt.Printf("=== Applied %s <= %s (overwrite: %v) ===\n", regex.phrase.tagName(), regex.pattern, regex.overwrite)
			if regex.overwrite {
				fmt.Println("Note: overwrite mode, replacing old [PP NPP NPS] tags")
			}
			fmt.Print(debug)
			fmt.Println()
		}
	}
}

// germanChunkTagForMatch ports GermanChunker.getChunkTag (NP → B-NP/I-NP).
func germanChunkTagForMatch(regex germanRegex2, matchStart, i int) string {
	if regex.phrase == phraseNP {
		if i == matchStart {
			return "B-NP"
		}
		return "I-NP"
	}
	return regex.phrase.tagName()
}

// assignGermanChunksToReadings ports GermanChunker.assignChunksToReadings.
// Always writes tags (including singleton O) onto the linked AnalyzedTokenReadings.
func assignGermanChunksToReadings(toks []ChunkTaggedToken) {
	for _, tagged := range toks {
		if tagged.Readings == nil {
			continue
		}
		strs := make([]string, 0, len(tagged.ChunkTags))
		for _, ct := range tagged.ChunkTags {
			if s := ct.GetChunkTag(); s != "" {
				strs = append(strs, s)
			}
		}
		if len(strs) == 0 {
			strs = []string{"O"}
		}
		tagged.Readings.SetChunkTags(strs)
	}
}

func germanChunkDebugString(tokens []ChunkTaggedToken) string {
	if !germanChunkerDebug {
		return ""
	}
	var b strings.Builder
	for _, token := range tokens {
		b.WriteString("  ")
		b.WriteString(token.String())
		b.WriteString(" -- ")
		if token.Readings != nil {
			b.WriteString(token.Readings.String())
		}
		b.WriteByte('\n')
	}
	return b.String()
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
