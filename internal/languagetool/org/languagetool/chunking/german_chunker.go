package chunking

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// GermanChunker ports org.languagetool.chunking.GermanChunker.
// REGEXES1 (basic OpenNLP-like B-NP/I-NP) is implemented from Java GermanChunker.REGEXES1
// with sequential POS/surface matchers — not invent. REGEXES2 (NPS/NPP/PP OpenRegex)
// is a growing faithful subset; remaining OpenRegex-only paths stay incomplete.
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

// applyRegexes1 runs Java REGEXES1 matchers into tags (B-NP/I-NP overwrite spans).
func applyRegexes1(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	if len(content) == 0 {
		return
	}
	// Java doApplyRegex: each REGEXES1 match assigns B-NP/I-NP over the full span
	// (later patterns may overwrite earlier assignments on the same tokens).
	applyNP := func(start, end int) {
		if start < 0 || end <= start || end > len(content) {
			return
		}
		for i := start; i < end; i++ {
			if i == start {
				tags[i] = []string{"B-NP"}
			} else {
				tags[i] = []string{"I-NP"}
			}
		}
	}
	// REGEXES1 patterns from Java GermanChunker (in order).
	matchers := []func([]*languagetool.AnalyzedTokenReadings, int) int{
		matchArtProAdvPa2AdjSub, // (ART|PRO)? ADV* PA2* ADJ* SUB+
		matchSubConjSub,         // SUB und|oder SUB
		matchAdjConjPa2Sub,      // ADJ und|oder PA2 SUB
		matchAdjConjAdjSub,      // ADJ und|oder ADJ SUB
		matchArtAdvAdjCapital,   // ART ADV* ADJ* Capitalized
		matchProZalSub,          // PRO? ZAL SUB
		matchTitleEig,           // Herr|Herrn|Frau EIG+
		matchTitleCapital,       // Herr|Herrn|Frau Capitalized+
		matchDerAlone,           // der
	}
	for _, m := range matchers {
		// Java findAll: scan whole sequence; do not skip already-tagged starts.
		for i := 0; i < len(content); {
			end := m(content, i)
			if end > i {
				applyNP(i, end)
				i = end
				continue
			}
			i++
		}
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

	// REGEXES2 subset: NPS/NPP/PP on top of B-NP/I-NP (Java OpenRegex paths; incomplete full set).
	// No invent POS→BIO fill-in for untagged tokens (Java leaves them as O).
	applyRegexes2(content, tags)

	for i, t := range content {
		if len(tags[i]) > 0 {
			t.SetChunkTags(tags[i])
		}
	}
}

// addChunkTag appends tag if missing; removes "O" like Java.
func addChunkTag(tags [][]string, i int, tag string) {
	if i < 0 || i >= len(tags) || tag == "" {
		return
	}
	for _, t := range tags[i] {
		if t == tag {
			return
		}
	}
	out := make([]string, 0, len(tags[i])+1)
	for _, t := range tags[i] {
		if t != "O" {
			out = append(out, t)
		}
	}
	out = append(out, tag)
	tags[i] = out
}

func addChunkTagSpan(tags [][]string, start, end int, tag string) {
	for i := start; i < end && i < len(tags); i++ {
		addChunkTag(tags, i, tag)
	}
}

func hasChunk(tags [][]string, i int, tag string) bool {
	if i < 0 || i >= len(tags) {
		return false
	}
	for _, t := range tags[i] {
		if t == tag {
			return true
		}
	}
	return false
}

// npSpanFrom returns end of B-NP I-NP* starting at i, or -1.
func npSpanFrom(tags [][]string, i int) int {
	if !hasChunk(tags, i, "B-NP") {
		return -1
	}
	j := i + 1
	for j < len(tags) && hasChunk(tags, j, "I-NP") {
		j++
	}
	return j
}

func posHasPLU(t *languagetool.AnalyzedTokenReadings) bool {
	return posContains(t, "PLU")
}

func posHasSIN(t *languagetool.AnalyzedTokenReadings) bool {
	return posContains(t, "SIN")
}

// applyRegexes2 ports a subset of GermanChunker.REGEXES2 (high-value NPP/NPS/PP).
// Full OpenRegex set remains incomplete; only Java-listed patterns below.
func applyRegexes2(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	if len(content) == 0 {
		return
	}
	// <pos=EIG> <und> <pos=EIG> → NPP
	for i := 0; i+2 < len(content); i++ {
		if posContains(content[i], "EIG") && surfaceEq(content[i+1], "und") && posContains(content[i+2], "EIG") {
			addChunkTagSpan(tags, i, i+3, "NPP")
		}
	}
	// <ich|du|er|sie|es|wir|ihr|sie> <und|oder|sowie> <NP> → NPP
	for i := 0; i+2 < len(content); {
		if surfaceEq(content[i], "ich", "du", "er", "sie", "es", "wir", "ihr") &&
			surfaceEq(content[i+1], "und", "oder", "sowie") {
			if end := npSpanFrom(tags, i+2); end > i+2 {
				addChunkTagSpan(tags, i, end, "NPP")
				i = end
				continue
			}
			// also: er und seine Schwester — PRO after und before NP
			if i+3 < len(content) && posContains(content[i+2], "PRO") {
				if end := npSpanFrom(tags, i+3); end > i+3 {
					addChunkTagSpan(tags, i, end, "NPP")
					i = end
					continue
				}
			}
		}
		i++
	}
	// <weder> <pos=SUB> <noch> <pos=SUB> → NPP
	for i := 0; i+3 < len(content); i++ {
		if surfaceEq(content[i], "weder") && posContains(content[i+1], "SUB") &&
			surfaceEq(content[i+2], "noch") && posContains(content[i+3], "SUB") {
			addChunkTagSpan(tags, i, i+4, "NPP")
		}
	}
	// <sowohl> … <als> <auch> … → NPP (Java REGEXES2; tests often commented but patterns are live)
	// 1) <sowohl> <NP> <als> <auch> <NP>
	// 2) <sowohl> <pos=EIG> <als> <auch> <pos=EIG>
	// 3) <sowohl> <ich|du|er|sie|es|wir|ihr|sie> <als> <auch> <NP>
	for i := 0; i+4 < len(content); i++ {
		if !surfaceEq(content[i], "sowohl") {
			continue
		}
		// pronoun form
		if surfaceEq(content[i+1], "ich", "du", "er", "sie", "es", "wir", "ihr") &&
			surfaceEq(content[i+2], "als") && surfaceEq(content[i+3], "auch") {
			end := npSpanFrom(tags, i+4)
			if end > i+4 {
				addChunkTagSpan(tags, i, end, "NPP")
				continue
			}
		}
		// EIG als auch EIG
		if posContains(content[i+1], "EIG") && surfaceEq(content[i+2], "als") &&
			surfaceEq(content[i+3], "auch") && i+4 < len(content) && posContains(content[i+4], "EIG") {
			addChunkTagSpan(tags, i, i+5, "NPP")
			continue
		}
		// NP als auch NP
		end1 := npSpanFrom(tags, i+1)
		if end1 <= i+1 {
			continue
		}
		if end1+1 >= len(content) || !surfaceEq(content[end1], "als") || !surfaceEq(content[end1+1], "auch") {
			continue
		}
		end2 := npSpanFrom(tags, end1+2)
		if end2 > end1+2 {
			addChunkTagSpan(tags, i, end2, "NPP")
		}
	}
	// <zwei|…|zwölf> <chunk=I-NP> → NPP; also number as B-NP head of plural NP
	numbers := map[string]struct{}{
		"zwei": {}, "drei": {}, "vier": {}, "fünf": {}, "sechs": {}, "sieben": {},
		"acht": {}, "neun": {}, "zehn": {}, "elf": {}, "zwölf": {},
	}
	for i := 0; i+1 < len(content); i++ {
		low := strings.ToLower(content[i].GetToken())
		if _, ok := numbers[low]; !ok {
			continue
		}
		if hasChunk(tags, i+1, "I-NP") {
			addChunkTagSpan(tags, i, i+2, "NPP")
		}
		// "drei Katzen" where drei may be B-NP and Katzen I-NP after REGEXES1 ZAL SUB
		if end := npSpanFrom(tags, i); end > i {
			addChunkTagSpan(tags, i, end, "NPP")
		}
	}
	// <pos=ADJ> <,> <chunk=B-NP> <chunk=I-NP>* <und|sowie> <NP> → NPP
	// "In christlichen, islamischen und jüdischen Traditionen"
	for i := 0; i+4 < len(content); i++ {
		if !posContains(content[i], "ADJ") || content[i+1].GetToken() != "," {
			continue
		}
		if !hasChunk(tags, i+2, "B-NP") {
			continue
		}
		j := i + 3
		for j < len(content) && hasChunk(tags, j, "I-NP") {
			j++
		}
		if j >= len(content) || !surfaceEq(content[j], "und", "sowie") {
			continue
		}
		end := npSpanFrom(tags, j+1)
		if end > j+1 {
			addChunkTagSpan(tags, i, end, "NPP")
			addChunkTag(tags, i+1, "NPP") // comma
		}
	}
	// <chunk=B-NP & !jede[rs]?> <chunk=I-NP>* <und|sowie> <pos=ADV>? <NP> → NPP
	// also <oder> for "Rekonstruktionen oder der Wiederaufbau"
	for i := 0; i < len(content); {
		end1 := npSpanFrom(tags, i)
		if end1 <= i {
			i++
			continue
		}
		// Java: !regex=jede[rs]?
		if isJedeForm(content[i].GetToken()) {
			i++
			continue
		}
		if end1 < len(content) && surfaceEq(content[end1], "und", "sowie", "oder") {
			next := end1 + 1
			// und keine … → NPS overwrite (Java before generic NPP on second NP)
			// "eine Masseeinheit und keine Gewichtseinheit"
			if next < len(content) && surfaceEq(content[next], "keine", "kein", "keinen") {
				j := next + 1
				for j < len(content) && hasChunk(tags, j, "I-NP") {
					j++
				}
				if j > next+1 {
					addChunkTagSpanOverwrite(tags, i, j, "NPS")
					i = j
					continue
				}
				if endK := npSpanFrom(tags, next); endK > next {
					addChunkTagSpanOverwrite(tags, i, endK, "NPS")
					i = endK
					continue
				}
			}
			// optional ADV after conj then NP
			if next < len(content) && posContains(content[next], "ADV") && !hasChunk(tags, next, "B-NP") {
				if endAdv := npSpanFrom(tags, next+1); endAdv > next+1 {
					addChunkTagSpan(tags, i, endAdv, "NPP")
					i = endAdv
					continue
				}
			}
			end2 := npSpanFrom(tags, next)
			if end2 > next {
				addChunkTagSpan(tags, i, end2, "NPP")
				i = end2
				continue
			}
		}
		i++
	}
	// <NP> <und|sowie> <pos=ART> <pos=PA1> <pos=SUB> → NPP (overwrite)
	// "Der See und das anliegende Marschland"
	for i := 0; i < len(content); {
		end1 := npSpanFrom(tags, i)
		if end1 <= i || end1+2 >= len(content) {
			i++
			continue
		}
		if !surfaceEq(content[end1], "und", "sowie") {
			i++
			continue
		}
		if posPrefix(content[end1+1], "ART") && posContains(content[end1+2], "PA1") &&
			end1+3 < len(content) && posContains(content[end1+3], "SUB") {
			addChunkTagSpanOverwrite(tags, i, end1+4, "NPP")
			i = end1 + 4
			continue
		}
		i++
	}
	// <pos=SUB> <und|oder|sowie> <chunk=B-NP & !ihre> … → NPP
	// Java excludes "ihre" so "Isolation und ihre Überwindung" uses the generic NP path instead.
	for i := 0; i+2 < len(content); i++ {
		if !posContains(content[i], "SUB") || !surfaceEq(content[i+1], "und", "oder", "sowie") {
			continue
		}
		if surfaceEq(content[i+2], "ihre") {
			continue
		}
		if end := npSpanFrom(tags, i+2); end > i+2 {
			addChunkTagSpan(tags, i, end, "NPP")
		}
	}
	// <er|sie|es> <und> <NP> <NP> → NPP — "sie und sein Sohn ein Paar"
	for i := 0; i+3 < len(content); i++ {
		if !surfaceEq(content[i], "er", "sie", "es") || !surfaceEq(content[i+1], "und") {
			continue
		}
		end1 := npSpanFrom(tags, i+2)
		if end1 <= i+2 {
			continue
		}
		end2 := npSpanFrom(tags, end1)
		if end2 > end1 {
			addChunkTagSpan(tags, i, end2, "NPP")
		}
	}
	// <Herr|Frau> <und> <Herr|Frau> <EIG>* → NPP
	for i := 0; i+2 < len(content); i++ {
		if !surfaceEq(content[i], "Herr", "Frau") || !surfaceEq(content[i+1], "und") ||
			!surfaceEq(content[i+2], "Herr", "Frau") {
			continue
		}
		j := i + 3
		for j < len(content) && posContains(content[j], "EIG") {
			j++
		}
		addChunkTagSpan(tags, i, j, "NPP")
	}
	// <pos=ART> <pos=ADJ> <und|sowie> (<pos=ADJ>|<pos=PA2>) <chunk=I-NP & !pos=PLU>+ → NPS (overwrite)
	// "die älteste und bekannteste Maßnahme"
	for i := 0; i+4 < len(content); i++ {
		if !posPrefix(content[i], "ART") || !posContains(content[i+1], "ADJ") ||
			!surfaceEq(content[i+2], "und", "sowie") {
			continue
		}
		if !posContains(content[i+3], "ADJ") && !posContains(content[i+3], "PA2") {
			continue
		}
		if !hasChunk(tags, i+4, "I-NP") || posHasPLU(content[i+4]) {
			continue
		}
		j := i + 5
		for j < len(content) && hasChunk(tags, j, "I-NP") && !posHasPLU(content[j]) {
			j++
		}
		addChunkTagSpanOverwrite(tags, i, j, "NPS")
	}
	// <pos=ADJ> <und|sowie> <chunk=B-NP & !pos=PLU> <chunk=I-NP>* → NPS (overwrite)
	// "größte und erfolgreichste Erfindung"
	for i := 0; i+2 < len(content); i++ {
		if !posContains(content[i], "ADJ") || !surfaceEq(content[i+1], "und", "sowie") {
			continue
		}
		if !hasChunk(tags, i+2, "B-NP") || posHasPLU(content[i+2]) {
			continue
		}
		end := npSpanFrom(tags, i+2)
		if end > i+2 {
			addChunkTagSpanOverwrite(tags, i, end, "NPS")
		}
	}
	// <deren> <chunk=B-NP & !pos=PLU> <und|sowie> <chunk=B-NP>* → NPS (overwrite)
	// "deren Bestimmung und Funktion"
	// Java OpenRegex matches surface <und> even when REGEXES1 tagged und as I-NP
	// (SUB und SUB). Do not skip und while advancing past true I-NP modifiers.
	for i := 0; i+2 < len(content); i++ {
		if !surfaceEq(content[i], "deren") || !hasChunk(tags, i+1, "B-NP") || posHasPLU(content[i+1]) {
			continue
		}
		j := i + 2
		// Skip I-NP after first B-NP head, but stop at und|sowie (may itself be I-NP).
		for j < len(content) && hasChunk(tags, j, "I-NP") && !surfaceEq(content[j], "und", "sowie") {
			j++
		}
		if j >= len(content) || !surfaceEq(content[j], "und", "sowie") {
			continue
		}
		j++ // past und|sowie
		end := j
		if end < len(content) && hasChunk(tags, end, "B-NP") {
			end = npSpanFrom(tags, end)
		} else {
			// After SUB und SUB, second noun may be I-NP not B-NP — still include it.
			for end < len(content) && hasChunk(tags, end, "I-NP") {
				end++
			}
		}
		if end > i+1 {
			addChunkTagSpanOverwrite(tags, i, end, "NPS")
		}
	}
	// <eins|eines> <chunk=B-NP> <chunk=I-NP>+ → NPS
	// "eins ihrer drei Autos"
	for i := 0; i+2 < len(content); i++ {
		if !surfaceEq(content[i], "eins", "eines") || !hasChunk(tags, i+1, "B-NP") {
			continue
		}
		if !hasChunk(tags, i+2, "I-NP") {
			continue
		}
		end := npSpanFrom(tags, i+1)
		if end > i+2 {
			addChunkTagSpan(tags, i, end, "NPS")
		}
	}
	// <chunk=B-NP> <pos=PRP> <NP> <pos=PA2> <chunk=B-NP> … → NPS/NPP
	// "der von der Regierung geprüfte Hund"
	// Also accept PA2 as B-NP head (REGEXES1 often tags PA2+SUB as one NP).
	for i := 0; i+4 < len(content); i++ {
		if !hasChunk(tags, i, "B-NP") || !posContains(content[i+1], "PRP") {
			continue
		}
		midEnd := npSpanFrom(tags, i+2)
		if midEnd <= i+2 {
			continue
		}
		if midEnd >= len(content) || !posContains(content[midEnd], "PA2") {
			continue
		}
		head := -1
		if hasChunk(tags, midEnd, "B-NP") {
			head = midEnd
		} else if midEnd+1 < len(content) && hasChunk(tags, midEnd+1, "B-NP") {
			head = midEnd + 1
		} else if midEnd+1 < len(content) && (posContains(content[midEnd+1], "SUB") || hasChunk(tags, midEnd+1, "I-NP")) {
			// PA2 + SUB without separate B-NP on head
			head = midEnd
		}
		if head < 0 {
			continue
		}
		end := npSpanFrom(tags, head)
		if end <= head {
			// bare PA2 + SUB
			end = head + 1
			for end < len(content) && (hasChunk(tags, end, "I-NP") || posContains(content[end], "SUB")) {
				end++
			}
		}
		if end <= head {
			continue
		}
		// number from head noun (last token of head span preferred)
		numTok := content[end-1]
		if posHasPLU(numTok) && !posHasSIN(numTok) {
			addChunkTagSpan(tags, i, end, "NPP")
		} else {
			addChunkTagSpan(tags, i, end, "NPS")
		}
	}
	// <chunk=B-NP> <pos=PRP> <NP> <chunk=B-NP & pos=SIN|PLU> … (no PA2)
	// sibling of PA2 pattern: "der von der Regierung [geprüfte] Hund" without participle
	for i := 0; i+3 < len(content); i++ {
		if !hasChunk(tags, i, "B-NP") || !posContains(content[i+1], "PRP") {
			continue
		}
		midEnd := npSpanFrom(tags, i+2)
		if midEnd <= i+2 || midEnd >= len(content) {
			continue
		}
		// skip if PA2 path already handles (next is PA2)
		if posContains(content[midEnd], "PA2") {
			continue
		}
		if !hasChunk(tags, midEnd, "B-NP") {
			continue
		}
		end := npSpanFrom(tags, midEnd)
		if end <= midEnd {
			continue
		}
		if posHasPLU(content[midEnd]) && !posHasSIN(content[midEnd]) {
			addChunkTagSpan(tags, i, end, "NPP")
		} else if posHasSIN(content[midEnd]) || !posHasPLU(content[midEnd]) {
			addChunkTagSpan(tags, i, end, "NPS")
		}
	}
	// <regex=eine[rs]?> <der> <beiden> <pos=ADJ>* <pos=SUB> → NPS
	// "Einer der beiden Höfe"
	for i := 0; i+3 < len(content); i++ {
		if !isEinerForm(content[i].GetToken()) || !surfaceEq(content[i+1], "der") || !surfaceEq(content[i+2], "beiden") {
			continue
		}
		j := i + 3
		for j < len(content) && posContains(content[j], "ADJ") {
			j++
		}
		if j < len(content) && posContains(content[j], "SUB") {
			addChunkTagSpan(tags, i, j+1, "NPS")
		}
	}
	// <regex=eine[rs]?> <der> <am> <pos=ADJ> <pos=PA2> <NP> → NPS
	// "eine der am meisten verbreiteten Krankheiten"
	for i := 0; i+5 < len(content); i++ {
		if !isEinerForm(content[i].GetToken()) || !surfaceEq(content[i+1], "der") || !surfaceEq(content[i+2], "am") {
			continue
		}
		if !posContains(content[i+3], "ADJ") || !posContains(content[i+4], "PA2") {
			continue
		}
		end := npSpanFrom(tags, i+5)
		if end > i+5 {
			addChunkTagSpan(tags, i, end, "NPS")
		} else if posContains(content[i+5], "SUB") {
			addChunkTagSpan(tags, i, i+6, "NPS")
		}
	}
	// <regex=eine[rs]?> <seiner|ihrer> <pos=PA1> <pos=SUB> → NPS
	for i := 0; i+3 < len(content); i++ {
		if !isEinerForm(content[i].GetToken()) || !surfaceEq(content[i+1], "seiner", "ihrer") {
			continue
		}
		if posContains(content[i+2], "PA1") && posContains(content[i+3], "SUB") {
			addChunkTagSpan(tags, i, i+4, "NPS")
		}
	}
	// <regex=[\d,.]+> <&prozent;> → NPS and NPP (Java SYNTAX_EXPANSION &prozent;)
	// &prozent; = Prozent|Kilo|Kilogramm|Gramm|Euro|Pfund
	for i := 0; i+1 < len(content); i++ {
		if !isNumberLike(content[i].GetToken()) || !isProzentExpansion(content[i+1].GetToken()) {
			continue
		}
		addChunkTagSpan(tags, i, i+2, "NPS")
		addChunkTagSpan(tags, i, i+2, "NPP")
	}
	// <dass> <sie> <wie> <NP> → NPP
	for i := 0; i+3 < len(content); i++ {
		if !surfaceEq(content[i], "dass", "daß") || !surfaceEq(content[i+1], "sie") || !surfaceEq(content[i+2], "wie") {
			continue
		}
		end := npSpanFrom(tags, i+3)
		if end > i+3 {
			addChunkTagSpan(tags, i, end, "NPP")
		}
	}
	// <pos=PLU> <die> <Regel> → NPP
	for i := 0; i+2 < len(content); i++ {
		if posHasPLU(content[i]) && surfaceEq(content[i+1], "die") && surfaceEq(content[i+2], "Regel") {
			addChunkTagSpan(tags, i, i+3, "NPP")
		}
	}
	// <NP> <,> <NP> <,> <NP> → NPP
	for i := 0; i < len(content); {
		end1 := npSpanFrom(tags, i)
		if end1 <= i || end1 >= len(content) || content[end1].GetToken() != "," {
			i++
			continue
		}
		end2 := npSpanFrom(tags, end1+1)
		if end2 <= end1+1 || end2 >= len(content) || content[end2].GetToken() != "," {
			i++
			continue
		}
		end3 := npSpanFrom(tags, end2+1)
		if end3 > end2+1 {
			addChunkTagSpan(tags, i, end3, "NPP")
			// also tag commas
			addChunkTag(tags, end1, "NPP")
			addChunkTag(tags, end2, "NPP")
			i = end3
			continue
		}
		i++
	}
	// Singular/plural NP from B-NP spans — Java REGEXES2 (sequential; NPP sees prior NPS via !chunk=NPS):
	// NPS: <chunk=B-NP & !pos=ZAL & !pos=PLU & !chunk=NPP & !einige & !(regex=&prozent;)>
	//      <chunk=I-NP & !pos=PLU & !und>*
	// NPP: <chunk=B-NP & !pos=SIN & !chunk=NPS & !Ellen> <chunk=I-NP & !pos=SIN>*
	// Must run before genitive patterns that require chunk=NPS/NPP.
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "B-NP") {
			i++
			continue
		}
		// NPS
		if !hasChunk(tags, i, "NPP") && !posContains(content[i], "ZAL") && !posHasPLU(content[i]) &&
			!surfaceEq(content[i], "einige") && !isProzentExpansion(content[i].GetToken()) {
			j := i + 1
			for j < len(content) && hasChunk(tags, j, "I-NP") &&
				!posHasPLU(content[j]) && !surfaceEq(content[j], "und") {
				j++
			}
			addChunkTagSpan(tags, i, j, "NPS")
		}
		// NPP (Java !chunk=NPS — skip if NPS just assigned)
		if !hasChunk(tags, i, "NPS") && !posHasSIN(content[i]) && !surfaceEq(content[i], "Ellen") {
			j := i + 1
			for j < len(content) && hasChunk(tags, j, "I-NP") && !posHasSIN(content[j]) {
				j++
			}
			addChunkTagSpan(tags, i, j, "NPP")
		}
		if end := npSpanFrom(tags, i); end > i {
			i = end
		} else {
			i++
		}
	}

	// Genitive phrases (Java REGEXES2 genitive section) — overwrite mode for NPS/NPP.
	applyGenitiveRegexes2(content, tags)

	// PP patterns (Java prepositional phrases)
	// <pos=PRP> <pos=ART:> <pos=ADV>* <pos=ADJ> <NP>
	for i := 0; i < len(content); {
		if !posContains(content[i], "PRP") {
			i++
			continue
		}
		j := i + 1
		if j < len(content) && posPrefix(content[j], "ART") {
			j++
		}
		for j < len(content) && posContains(content[j], "ADV") {
			j++
		}
		if j < len(content) && posContains(content[j], "ADJ") {
			j++
		}
		// optional PA1/PA2 before NP
		if j < len(content) && (posContains(content[j], "PA1") || posContains(content[j], "PA2")) {
			j++
		}
		end := npSpanFrom(tags, j)
		if end > j {
			addChunkTagSpan(tags, i, end, "PP")
			i = end
			continue
		}
		// <pos=PRP> <NP>
		end = npSpanFrom(tags, i+1)
		if end > i+1 {
			addChunkTagSpan(tags, i, end, "PP")
			i = end
			continue
		}
		// <pos=PRP> <pos=ADV> <pos=ZAL> <chunk=B-NP>
		if i+3 < len(content) && posContains(content[i+1], "ADV") &&
			(posContains(content[i+2], "ZAL") || isDigits(content[i+2].GetToken())) &&
			hasChunk(tags, i+3, "B-NP") {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=ADV> <regex=\d+> <NP>
		if i+3 < len(content) && posContains(content[i+1], "ADV") && isDigits(content[i+2].GetToken()) {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=ADJ> <und|oder|sowie> <NP>
		if i+3 < len(content) && posContains(content[i+1], "ADJ") &&
			surfaceEq(content[i+2], "und", "oder", "sowie") {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=ADV> <pos=ADJ> <NP>
		if i+3 < len(content) && posContains(content[i+1], "ADV") && posContains(content[i+2], "ADJ") {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=PA1> <NP>
		if i+2 < len(content) && posContains(content[i+1], "PA1") {
			end = npSpanFrom(tags, i+2)
			if end > i+2 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=ADJ:PRD:GRU> <pos=ZAL> <NP> — "Von ursprünglich drei Almhütten"
		if i+3 < len(content) && posContains(content[i+1], "ADJ:PRD:GRU") &&
			(posContains(content[i+2], "ZAL") || isDigits(content[i+2].GetToken())) {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=ADJ> <pos=PA1> <NP> — "Aufgrund stark schwankender Absatzmärkte"
		if i+3 < len(content) && posContains(content[i+1], "ADJ") && posContains(content[i+2], "PA1") {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <chunk=B-NP> <pos=ADV> <NP> — "in den darauf folgenden Wochen"
		if i+3 < len(content) && hasChunk(tags, i+1, "B-NP") && posContains(content[i+2], "ADV") {
			end = npSpanFrom(tags, i+3)
			if end > i+3 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// <pos=PRP> <pos=PRO> <NP> — "in deren deutschen Installationen"
		if i+2 < len(content) && posContains(content[i+1], "PRO") {
			end = npSpanFrom(tags, i+2)
			if end > i+2 {
				addChunkTagSpan(tags, i, end, "PP")
				i = end
				continue
			}
		}
		// Multi-NP PP expansions run as separate Java REGEXES2 passes below
		// (short <pos=PRP> <NP> above must not skip them — Java findAll is independent).
		i++
	}
	// Java REGEXES2 independent entry: <pos=PRP> <NP> <pos=ADJ> <und|oder|bzw.> <NP> → PP
	// "einschließlich der biologischen und sozialen Grundlagen"
	for i := 0; i < len(content); {
		if !posContains(content[i], "PRP") {
			i++
			continue
		}
		endNP1 := npSpanFrom(tags, i+1)
		if endNP1 <= i+1 {
			i++
			continue
		}
		if endNP1 < len(content) && posContains(content[endNP1], "ADJ") &&
			endNP1+1 < len(content) && surfaceEq(content[endNP1+1], "und", "oder", "bzw.") {
			end2 := npSpanFrom(tags, endNP1+2)
			if end2 > endNP1+2 {
				addChunkTagSpan(tags, i, end2, "PP")
				i = end2
				continue
			}
		}
		i++
	}
	// Java REGEXES2 independent entry: <pos=PRP> <NP> <NP> <und|oder> <NP> → PP
	// "durch Einsatz größerer Maschinen und bessere Kapazitätsplanung"
	for i := 0; i < len(content); {
		if !posContains(content[i], "PRP") {
			i++
			continue
		}
		endNP1 := npSpanFrom(tags, i+1)
		if endNP1 <= i+1 {
			i++
			continue
		}
		endNP2 := npSpanFrom(tags, endNP1)
		if endNP2 <= endNP1 {
			i++
			continue
		}
		if endNP2 < len(content) && surfaceEq(content[endNP2], "und", "oder") {
			end3 := npSpanFrom(tags, endNP2+1)
			if end3 > endNP2+1 {
				addChunkTagSpan(tags, i, end3, "PP")
				i = end3
				continue
			}
		}
		i++
	}
	// Java REGEXES2 independent entry: <pos=PRP> (<NP>)+ → PP
	// "für Ärzte und Ärztinnen festgestellte Risikoprofil"
	// One or more consecutive B-NP/I-NP* spans after the preposition (no invent gaps).
	for i := 0; i < len(content); {
		if !posContains(content[i], "PRP") {
			i++
			continue
		}
		end := i + 1
		nps := 0
		for end < len(content) {
			npEnd := npSpanFrom(tags, end)
			if npEnd <= end {
				break
			}
			nps++
			end = npEnd
		}
		if nps >= 1 {
			addChunkTagSpan(tags, i, end, "PP")
			i = end
			continue
		}
		i++
	}
	// <regex=(vor)?letzte[sn]?> <Woche|Monat|Jahr|…> → PP
	for i := 0; i+1 < len(content); i++ {
		if isLetzteForm(content[i].GetToken()) && isTimeUnit(content[i+1].GetToken()) {
			addChunkTagSpan(tags, i, i+2, "PP")
		}
	}
	// <die> <pos=ADJ> <Sekunden|…|Jahre> → PP
	for i := 0; i+2 < len(content); i++ {
		if surfaceEq(content[i], "die") && posContains(content[i+1], "ADJ") && isTimeUnit(content[i+2].GetToken()) {
			addChunkTagSpan(tags, i, i+3, "PP")
		}
	}
	// <die> <pos=ADJ> <pos=ZAL> <time unit> → PP — "die letzten zwei Monate"
	for i := 0; i+3 < len(content); i++ {
		if !surfaceEq(content[i], "die") || !posContains(content[i+1], "ADJ") {
			continue
		}
		if !posContains(content[i+2], "ZAL") && !isDigits(content[i+2].GetToken()) {
			continue
		}
		if isTimeUnit(content[i+3].GetToken()) {
			addChunkTagSpan(tags, i, i+4, "PP")
		}
	}
	// <für> <in> <pos=EIG> <pos=PA1> <pos=SUB> <und> <pos=SUB> → PP (overwrite)
	for i := 0; i+6 < len(content); i++ {
		if surfaceEq(content[i], "für") && surfaceEq(content[i+1], "in") &&
			posContains(content[i+2], "EIG") && posContains(content[i+3], "PA1") &&
			posContains(content[i+4], "SUB") && surfaceEq(content[i+5], "und") &&
			posContains(content[i+6], "SUB") {
			addChunkTagSpanOverwrite(tags, i, i+7, "PP")
		}
	}
	// Late REGEXES2 that need NPS/NPP already assigned (Java: after PP section).
	applyLateRegexes2(content, tags)
}

func applyGenitiveRegexes2(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	// <der|die|das> <pos=ADJ> <der> <pos=PA1> <pos=SUB> → NPS
	// "Das letzte der teilnehmenden Länder"
	for i := 0; i+4 < len(content); i++ {
		if !surfaceEq(content[i], "der", "die", "das") || !posContains(content[i+1], "ADJ") ||
			!surfaceEq(content[i+2], "der") || !posContains(content[i+3], "PA1") || !posContains(content[i+4], "SUB") {
			continue
		}
		addChunkTagSpanOverwrite(tags, i, i+5, "NPS")
	}
	// <pos=SUB & PLU> <der> <pos=PA1> <pos=SUB> → NPP
	for i := 0; i+3 < len(content); i++ {
		if !posContains(content[i], "SUB") || !posHasPLU(content[i]) || !surfaceEq(content[i+1], "der") ||
			!posContains(content[i+2], "PA1") || !posContains(content[i+3], "SUB") {
			continue
		}
		addChunkTagSpanOverwrite(tags, i, i+4, "NPP")
	}
	// <der|die|das> <pos=ADJ> <der> <pos=PRO>? <pos=SUB> → NPS
	// "die ältere der beiden Töchter"
	for i := 0; i+3 < len(content); i++ {
		if !surfaceEq(content[i], "der", "die", "das") || !posContains(content[i+1], "ADJ") || !surfaceEq(content[i+2], "der") {
			continue
		}
		j := i + 3
		if j < len(content) && posContains(content[j], "PRO") {
			j++
		}
		if j < len(content) && posContains(content[j], "SUB") {
			addChunkTagSpanOverwrite(tags, i, j+1, "NPS")
		}
	}
	// <der|das> <pos=ADJ> <der> <pos=ZAL> <NP> → NPS
	// "der letzte der vier großen Flüsse"
	for i := 0; i+4 < len(content); i++ {
		if !surfaceEq(content[i], "der", "das") || !posContains(content[i+1], "ADJ") ||
			!surfaceEq(content[i+2], "der") || !posContains(content[i+3], "ZAL") {
			continue
		}
		end := npSpanFrom(tags, i+4)
		if end > i+4 {
			addChunkTagSpanOverwrite(tags, i, end, "NPS")
		}
	}
	// <chunk=NPS & !einige> <chunk=NPP & (pos=GEN | pos=ZAL)>+ → NPS (overwrite)
	// "Synthese organischer Verbindungen", "die Anordnung der vier Achsen"
	// but not "Einige der Inhaltsstoffe"
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "NPS") || surfaceEq(content[i], "einige") {
			i++
			continue
		}
		j := i
		for j < len(content) && hasChunk(tags, j, "NPS") {
			j++
		}
		k := j
		for k < len(content) && hasChunk(tags, k, "NPP") &&
			(posContains(content[k], "GEN") || posContains(content[k], "ZAL")) {
			k++
		}
		if k > j {
			addChunkTagSpanOverwrite(tags, i, k, "NPS")
			i = k
			continue
		}
		// "organischer Verbindungen" may still be B-NP/I-NP with GEN before NPP assignment
		// only if already NPP — Java requires chunk=NPP; no invent from bare GEN ADJ+SUB.
		if j > i {
			i = j
		} else {
			i++
		}
	}
	// Java REGEXES2 genitive (surface token <der> only — not des/dem/den invent):
	//   <chunk=NPS>+ <der> <pos=ADV> <pos=PA2> <chunk=I-NP>
	//   <chunk=NPS>+ <der> (<pos=ADJ>|<pos=ZAL>) <NP>
	//   <chunk=NPS>+ <der> <NP>
	//   <chunk=NPS>+ <der> <pos=ADJ> <pos=ADV> <pos=PA2> <NP>
	// "des/dem/den" genitives use other patterns (NPS+NPP&GEN, PRO:POS, …).
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "NPS") {
			i++
			continue
		}
		j := i
		for j < len(content) && hasChunk(tags, j, "NPS") {
			j++
		}
		if j < len(content) && surfaceEq(content[j], "der") {
			// optional ADJ|ZAL then NP
			k := j + 1
			if k < len(content) && (posContains(content[k], "ADJ") || posContains(content[k], "ZAL")) {
				// may be two modifiers "ersten beiden"
				k2 := k + 1
				if k2 < len(content) && (posContains(content[k2], "ADJ") || posContains(content[k2], "ZAL") || posContains(content[k2], "PRO")) {
					if end := npSpanFrom(tags, k2); end > k2 {
						addChunkTagSpanOverwrite(tags, i, end, "NPS")
						i = end
						continue
					}
				}
				if end := npSpanFrom(tags, k); end > k {
					addChunkTagSpanOverwrite(tags, i, end, "NPS")
					i = end
					continue
				}
			}
			if end := npSpanFrom(tags, k); end > k {
				addChunkTagSpanOverwrite(tags, i, end, "NPS")
				i = end
				continue
			}
			// der ADV PA2 I-NP — "Teil der dort ausgestellten Bestände"
			if k < len(content) && posContains(content[k], "ADV") && k+1 < len(content) && posContains(content[k+1], "PA2") {
				if k+2 < len(content) && (hasChunk(tags, k+2, "I-NP") || hasChunk(tags, k+2, "B-NP")) {
					end := k + 2
					if hasChunk(tags, k+2, "B-NP") {
						end = npSpanFrom(tags, k+2)
					} else {
						for end < len(content) && hasChunk(tags, end, "I-NP") {
							end++
						}
					}
					if end > k+1 {
						addChunkTagSpanOverwrite(tags, i, end, "NPS")
						i = end
						continue
					}
				}
			}
			// der ADJ ADV PA2 NP
			if k+3 < len(content) && posContains(content[k], "ADJ") && posContains(content[k+1], "ADV") &&
				posContains(content[k+2], "PA2") {
				if end := npSpanFrom(tags, k+3); end > k+3 {
					addChunkTagSpanOverwrite(tags, i, end, "NPS")
					i = end
					continue
				}
			}
		}
		// Java: after NPS+NPP(GEN) absorbed "der", extend remaining genitive material.
		// - ADV PA2 I-NP*: "Teil der dort ausgestellten Bestände"
		// - B-NP I-NP* (optional untagged ADJ|PRO before NP): "Autor der (ersten) beiden Bücher"
		if j > i && j < len(content) {
			last := content[j-1]
			if surfaceEq(last, "der") {
				if posContains(content[j], "ADV") && j+1 < len(content) && posContains(content[j+1], "PA2") {
					end := j + 2
					if end < len(content) && (hasChunk(tags, end, "I-NP") || hasChunk(tags, end, "B-NP")) {
						if hasChunk(tags, end, "B-NP") {
							end = npSpanFrom(tags, end)
						} else {
							for end < len(content) && hasChunk(tags, end, "I-NP") {
								end++
							}
						}
						if end > j+1 {
							addChunkTagSpanOverwrite(tags, i, end, "NPS")
							i = end
							continue
						}
					}
				}
				// Optional untagged modifiers then B-NP (ersten may stay O after REGEXES1).
				k := j
				for k < len(content) && !hasChunk(tags, k, "B-NP") && !hasChunk(tags, k, "I-NP") &&
					(posContains(content[k], "ADJ") || posContains(content[k], "ZAL") || posContains(content[k], "PRO")) {
					k++
				}
				if k < len(content) && hasChunk(tags, k, "B-NP") {
					if end := npSpanFrom(tags, k); end > k {
						addChunkTagSpanOverwrite(tags, i, end, "NPS")
						i = end
						continue
					}
				}
				// Trailing I-NP/ADJ/PRO/SUB residual of a split genitive NP.
				end := j
				for end < len(content) {
					if !hasChunk(tags, end, "I-NP") && !hasChunk(tags, end, "B-NP") {
						break
					}
					if !(posContains(content[end], "ADJ") || posContains(content[end], "ADV") ||
						posContains(content[end], "SUB") || posContains(content[end], "PRO") ||
						posContains(content[end], "PA1") || posContains(content[end], "PA2") ||
						posContains(content[end], "ZAL")) {
						break
					}
					end++
				}
				if end > j {
					addChunkTagSpanOverwrite(tags, i, end, "NPS")
					i = end
					continue
				}
			}
		}
		// <chunk=NPS>+ <pos=PRO:POS> <pos=ADJ> <NP>
		if j < len(content) && posContains(content[j], "PRO") && j+1 < len(content) && posContains(content[j+1], "ADJ") {
			if end := npSpanFrom(tags, j+2); end > j+2 {
				addChunkTagSpanOverwrite(tags, i, end, "NPS")
				i = end
				continue
			}
		}
		// Java REGEXES2: <chunk=NPS>+ <und> <chunk=NP[SP] & (pos=GEN | pos=ADV)>+ → NPS (overwrite)
		// "die Pyramide des Friedens und der Eintracht" — second span must carry GEN or ADV.
		// Do not invent merge for bare NPP after und (e.g. "der Sowjetunion und Kuba").
		if j < len(content) && surfaceEq(content[j], "und") {
			k := j + 1
			if k < len(content) && (hasChunk(tags, k, "NPS") || hasChunk(tags, k, "NPP") || hasChunk(tags, k, "B-NP")) {
				end := k + 1
				if hasChunk(tags, k, "B-NP") {
					end = npSpanFrom(tags, k)
				} else {
					for end < len(content) && (hasChunk(tags, end, "NPS") || hasChunk(tags, end, "NPP") || hasChunk(tags, end, "I-NP")) {
						end++
					}
				}
				// Java: pos=GEN | pos=ADV on the post-und NP[SP] span (OpenRegex feature filter).
				hasGenOrAdv := false
				for t := k; t < end; t++ {
					if posContains(content[t], "GEN") || posContains(content[t], "ADV") {
						hasGenOrAdv = true
						break
					}
				}
				if hasGenOrAdv {
					addChunkTagSpanOverwrite(tags, i, end, "NPS")
					i = end
					continue
				}
			}
		}
		i = j
		if i == 0 || i <= len(content) && i == j {
			i++
		}
	}
	// <chunk=NPP> <chunk=NPS & GEN>+ → NPP — "die Kenntnisse der Sprache"
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "NPP") {
			i++
			continue
		}
		j := i
		for j < len(content) && hasChunk(tags, j, "NPP") {
			j++
		}
		k := j
		for k < len(content) && hasChunk(tags, k, "NPS") && posContains(content[k], "GEN") {
			k++
		}
		// also GEN on following NP after NPP (Java: chunk=NPS & pos=GEN — require GEN, no invent on bare dem/den)
		if k == j {
			if end := npSpanFrom(tags, j); end > j {
				hasGEN := false
				for t := j; t < end; t++ {
					if posContains(content[t], "GEN") {
						hasGEN = true
						break
					}
				}
				// "der Sprache" / "des Friedens": ART+SUB carry GEN in Morphy tags
				if hasGEN {
					if end2 := npSpanFrom(tags, j); end2 > j {
						addChunkTagSpanOverwrite(tags, i, end2, "NPP")
						i = end2
						continue
					}
				}
			}
		} else if k > j {
			addChunkTagSpanOverwrite(tags, i, k, "NPP")
			i = k
			continue
		}
		i++
	}
	// <eine> <menge> <NP>+ → NPP (overwrite)
	for i := 0; i+2 < len(content); i++ {
		if !surfaceEq(content[i], "eine", "ein") || !surfaceEq(content[i+1], "Menge", "menge") {
			continue
		}
		end := npSpanFrom(tags, i+2)
		if end > i+2 {
			addChunkTagSpanOverwrite(tags, i, end, "NPP")
		} else if posContains(content[i+2], "ADJ") || posContains(content[i+2], "SUB") {
			// "englischer Wörter" may not yet be one NP — take ADJ* SUB+
			j := i + 2
			for j < len(content) && (posContains(content[j], "ADJ") || posContains(content[j], "SUB")) {
				j++
			}
			if j > i+2 {
				addChunkTagSpanOverwrite(tags, i, j, "NPP")
			}
		}
	}
	// <laut> … <Quellen> → PP
	for i := 0; i < len(content); i++ {
		if !surfaceEq(content[i], "laut") {
			continue
		}
		for j := i + 1; j < len(content) && j <= i+4; j++ {
			if surfaceEq(content[j], "Quellen", "Quelle") {
				addChunkTagSpanOverwrite(tags, i, j+1, "PP")
				break
			}
		}
	}
}

func applyLateRegexes2(content []*languagetool.AnalyzedTokenReadings, tags [][]string) {
	// <chunk=NPS> <pos=PRO> <pos=ADJ> <pos=ADJ> <NP> → NPS
	// "die hohe Zahl dieser relativ kleinen Verwaltungseinheiten"
	// REGEXES1 often makes "dieser … Verwaltungseinheiten" one B-NP/I-NP* span
	// (PRO + ADJ* + SUB), so also accept PRO-headed NP directly after NPS.
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "NPS") {
			i++
			continue
		}
		j := i
		for j < len(content) && hasChunk(tags, j, "NPS") {
			j++
		}
		if j < len(content) && posContains(content[j], "PRO") {
			// Strict Java: PRO ADJ ADJ NP (NP may start at third ADJ or after).
			if j+3 < len(content) && posContains(content[j+1], "ADJ") && posContains(content[j+2], "ADJ") {
				if end := npSpanFrom(tags, j+3); end > j+3 {
					addChunkTagSpan(tags, i, end, "NPS")
					i = end
					continue
				}
				// NP head may be the second ADJ token when REGEXES1 started B-NP there.
				if end := npSpanFrom(tags, j+2); end > j+2 {
					addChunkTagSpan(tags, i, end, "NPS")
					i = end
					continue
				}
			}
			// PRO-headed B-NP span (common after REGEXES1 on "dieser relativ kleinen …").
			if hasChunk(tags, j, "B-NP") {
				if end := npSpanFrom(tags, j); end > j {
					// Prefer spans that look like genitive NP: contain ADJ or SUB after PRO.
					ok := false
					for t := j + 1; t < end; t++ {
						if posContains(content[t], "ADJ") || posContains(content[t], "SUB") {
							ok = true
							break
						}
					}
					if ok {
						addChunkTagSpan(tags, i, end, "NPS")
						i = end
						continue
					}
				}
			}
		}
		// Genitive NPS+NPP(GEN) may absorb only the PRO ("dieser") into NPS and leave
		// trailing I-NP (relativ kleinen Verwaltungseinheiten) as NPP — extend NPS over them.
		if j > i && j < len(content) && posContains(content[j-1], "PRO") {
			end := j
			for end < len(content) {
				if !hasChunk(tags, end, "I-NP") && !hasChunk(tags, end, "B-NP") {
					break
				}
				if !(posContains(content[end], "ADJ") || posContains(content[end], "ADV") ||
					posContains(content[end], "SUB") || posContains(content[end], "PA1") ||
					posContains(content[end], "PA2") || posContains(content[end], "PRO")) {
					break
				}
				end++
			}
			if end > j {
				addChunkTagSpan(tags, i, end, "NPS")
				i = end
				continue
			}
		}
		if j > i {
			i = j
		} else {
			i++
		}
	}
	// <chunk=B-NP & pos=SIN|PLU> <chunk=I-NP>* <,> <die> <pos=ADV>+ <chunk=NPS>+ → NPS/NPP
	// "Veranstaltung, die immer wieder ein kultureller Höhepunkt"
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "B-NP") {
			i++
			continue
		}
		endNP := npSpanFrom(tags, i)
		if endNP <= i || endNP >= len(content) || content[endNP].GetToken() != "," {
			i++
			continue
		}
		if endNP+1 >= len(content) || !surfaceEq(content[endNP+1], "die") {
			i++
			continue
		}
		// at least one ADV
		k := endNP + 2
		if k >= len(content) || !posContains(content[k], "ADV") {
			i++
			continue
		}
		for k < len(content) && posContains(content[k], "ADV") {
			k++
		}
		if k >= len(content) || !hasChunk(tags, k, "NPS") {
			// trailing phrase may be B-NP not yet NPS — accept B-NP singular-ish span
			if k < len(content) && hasChunk(tags, k, "B-NP") {
				end2 := npSpanFrom(tags, k)
				if end2 > k {
					tag := "NPS"
					if posHasPLU(content[i]) && !posHasSIN(content[i]) {
						tag = "NPP"
					}
					addChunkTagSpan(tags, i, end2, tag)
					addChunkTag(tags, endNP, tag)
					i = end2
					continue
				}
			}
			i++
			continue
		}
		end2 := k
		for end2 < len(content) && hasChunk(tags, end2, "NPS") {
			end2++
		}
		tag := "NPS"
		if posHasPLU(content[i]) && !posHasSIN(content[i]) {
			tag = "NPP"
		}
		addChunkTagSpan(tags, i, end2, tag)
		addChunkTag(tags, endNP, tag)
		i = end2
	}
	// <chunk=NPP> <zwischen> <pos=EIG> <und|sowie> <NP> → NPP
	// "die Beziehungen zwischen Kanada und dem Iran"
	for i := 0; i < len(content); {
		if !hasChunk(tags, i, "NPP") {
			i++
			continue
		}
		j := i
		for j < len(content) && hasChunk(tags, j, "NPP") {
			j++
		}
		if j < len(content) && surfaceEq(content[j], "zwischen") &&
			j+1 < len(content) && posContains(content[j+1], "EIG") &&
			j+2 < len(content) && surfaceEq(content[j+2], "und", "sowie") {
			end := npSpanFrom(tags, j+3)
			if end > j+3 {
				addChunkTagSpan(tags, i, end, "NPP")
				i = end
				continue
			}
			// dem Iran may be ART+EIG not yet one NP
			if j+3 < len(content) && (posPrefix(content[j+3], "ART") || posContains(content[j+3], "EIG")) {
				k := j + 4
				if j+3 < len(content) && posContains(content[j+3], "EIG") {
					k = j + 4
				} else if j+4 < len(content) && posContains(content[j+4], "EIG") {
					k = j + 5
				}
				if k > j+3 {
					addChunkTagSpan(tags, i, k, "NPP")
					i = k
					continue
				}
			}
		}
		i = j
		if i <= len(content) && j == i {
			i++
		}
	}
	// <,> <die|welche> <NP> <chunk=NPS & pos=GEN>+ → NPP
	// "Atome, welche der Urstoff aller Körper"
	for i := 0; i+3 < len(content); i++ {
		if content[i].GetToken() != "," || !surfaceEq(content[i+1], "die", "welche") {
			continue
		}
		endNP := npSpanFrom(tags, i+2)
		if endNP <= i+2 {
			continue
		}
		k := endNP
		for k < len(content) && hasChunk(tags, k, "NPS") && posContains(content[k], "GEN") {
			k++
		}
		// Java: after <NP>, require <chunk=NPS & pos=GEN>+ (OpenRegex pos=GEN substring).
		// "der Urstoff" may be NOM NP; GEN sits on "aller Körper" — scan past first NP for GEN material.
		// No invent on bare der/des/dem/den without a GEN feature somewhere in the tail.
		if k == endNP {
			// First NP after welche/die already in [i+2, endNP). Look for GEN-bearing span after it.
			if endNP < len(content) {
				end2 := endNP
				if hasChunk(tags, endNP, "B-NP") {
					end2 = npSpanFrom(tags, endNP)
				} else if hasChunk(tags, endNP, "NPS") || hasChunk(tags, endNP, "NPP") {
					for end2 < len(content) && (hasChunk(tags, end2, "NPS") || hasChunk(tags, end2, "NPP") || hasChunk(tags, end2, "I-NP")) {
						end2++
					}
				} else if end := npSpanFrom(tags, endNP); end > endNP {
					end2 = end
				}
				hasGEN := false
				for t := endNP; t < end2; t++ {
					if posContains(content[t], "GEN") {
						hasGEN = true
						break
					}
				}
				// Trailing genitive NP after an intermediate NOM NP ("der Urstoff" + "aller Körper")
				if !hasGEN && end2 < len(content) {
					if end3 := npSpanFrom(tags, end2); end3 > end2 {
						for t := end2; t < end3; t++ {
							if posContains(content[t], "GEN") {
								hasGEN = true
								end2 = end3
								break
							}
						}
					} else if hasChunk(tags, end2, "NPS") || hasChunk(tags, end2, "NPP") || hasChunk(tags, end2, "B-NP") {
						e := end2 + 1
						if hasChunk(tags, end2, "B-NP") {
							e = npSpanFrom(tags, end2)
						} else {
							for e < len(content) && (hasChunk(tags, e, "NPS") || hasChunk(tags, e, "NPP") || hasChunk(tags, e, "I-NP")) {
								e++
							}
						}
						for t := end2; t < e; t++ {
							if posContains(content[t], "GEN") {
								hasGEN = true
								end2 = e
								break
							}
						}
					}
				}
				if hasGEN && end2 > endNP {
					addChunkTagSpan(tags, i, end2, "NPP")
				}
			}
		} else if k > endNP {
			addChunkTagSpan(tags, i, k, "NPP")
		}
	}
	// <NP> <,> <NP> <,> <wie> <auch> <chunk=NPS>+ → NPP
	for i := 0; i < len(content); {
		end1 := npSpanFrom(tags, i)
		if end1 <= i || end1 >= len(content) || content[end1].GetToken() != "," {
			i++
			continue
		}
		end2 := npSpanFrom(tags, end1+1)
		if end2 <= end1+1 || end2 >= len(content) || content[end2].GetToken() != "," {
			i++
			continue
		}
		if end2+2 < len(content) && surfaceEq(content[end2+1], "wie") && surfaceEq(content[end2+2], "auch") {
			k := end2 + 3
			if k < len(content) && hasChunk(tags, k, "NPS") {
				for k < len(content) && hasChunk(tags, k, "NPS") {
					k++
				}
				addChunkTagSpan(tags, i, k, "NPP")
				addChunkTag(tags, end1, "NPP")
				addChunkTag(tags, end2, "NPP")
				i = k
				continue
			}
		}
		i++
	}
	// <pos=PRP> <chunk=NPP>+ → PP (and optional ", NP")
	// "für die Stadtteile und selbständigen Ortsteile"
	// "in den alten Religionen, Mythen und Sagen"
	for i := 0; i < len(content); {
		if !posContains(content[i], "PRP") {
			i++
			continue
		}
		if !hasChunk(tags, i+1, "NPP") && !(i+1 < len(content) && hasChunk(tags, i+1, "B-NP")) {
			i++
			continue
		}
		j := i + 1
		for j < len(content) && (hasChunk(tags, j, "NPP") || hasChunk(tags, j, "I-NP") || hasChunk(tags, j, "B-NP")) {
			// stop if we left the NPP/NP chain into non-conj material without NPP
			j++
		}
		// include und/sowie inside NPP span already tagged
		if j > i+1 {
			// optional ", NP"
			if j < len(content) && content[j].GetToken() == "," {
				end2 := npSpanFrom(tags, j+1)
				if end2 > j+1 {
					addChunkTagSpan(tags, i, end2, "PP")
					addChunkTag(tags, j, "PP")
					i = end2
					continue
				}
			}
			addChunkTagSpan(tags, i, j, "PP")
			i = j
			continue
		}
		i++
	}
}

// addChunkTagSpanOverwrite removes FILTER_TAGS (PP/NPP/NPS) then adds tag (Java overwrite mode).
func addChunkTagSpanOverwrite(tags [][]string, start, end int, tag string) {
	filter := map[string]struct{}{"PP": {}, "NPP": {}, "NPS": {}}
	for i := start; i < end && i < len(tags); i++ {
		out := make([]string, 0, len(tags[i])+1)
		for _, t := range tags[i] {
			if t == "O" {
				continue
			}
			if _, drop := filter[t]; drop {
				continue
			}
			if t == tag {
				continue
			}
			out = append(out, t)
		}
		out = append(out, tag)
		tags[i] = out
	}
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isNumberLike ports Java regex=[\d,.]+ for "37" / "1,4" Prozent patterns.
func isNumberLike(s string) bool {
	if s == "" {
		return false
	}
	hasDigit := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasDigit = true
			continue
		}
		if r == ',' || r == '.' {
			continue
		}
		return false
	}
	return hasDigit
}

// isEinerForm ports Java regex=eine[rs]? (eine|einer|eines).
func isEinerForm(s string) bool {
	switch strings.ToLower(s) {
	case "eine", "einer", "eines":
		return true
	}
	return false
}

// isJedeForm ports Java !regex=jede[rs]? guard on B-NP und NP → NPP.
func isJedeForm(s string) bool {
	switch strings.ToLower(s) {
	case "jede", "jeder", "jedes":
		return true
	}
	return false
}

// isProzentExpansion ports Java SYNTAX_EXPANSION &prozent; =
// Prozent|Kilo|Kilogramm|Gramm|Euro|Pfund
func isProzentExpansion(s string) bool {
	switch strings.ToLower(s) {
	case "prozent", "kilo", "kilogramm", "gramm", "euro", "pfund":
		return true
	}
	return false
}

func isLetzteForm(s string) bool {
	low := strings.ToLower(s)
	switch low {
	case "letzte", "letztes", "letzten", "vorletzte", "vorletztes", "vorletzten":
		return true
	}
	return false
}

func isTimeUnit(s string) bool {
	low := strings.ToLower(s)
	switch low {
	case "woche", "wochen", "monat", "monate", "jahr", "jahre", "jahrzehnt", "jahrzehnte",
		"jahrhundert", "jahrhunderte", "sekunde", "sekunden", "minute", "minuten",
		"stunde", "stunden", "tag", "tage":
		return true
	}
	return false
}

func firstPOS(t *languagetool.AnalyzedTokenReadings) string {
	if t == nil {
		return ""
	}
	for _, r := range t.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			return *r.GetPOSTag()
		}
	}
	return ""
}

// Java pos=X is substring match on any reading's POS tag.
func posContains(t *languagetool.AnalyzedTokenReadings, sub string) bool {
	if t == nil {
		return false
	}
	for _, r := range t.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if strings.Contains(*r.GetPOSTag(), sub) {
			return true
		}
	}
	return false
}

func posPrefix(t *languagetool.AnalyzedTokenReadings, pre string) bool {
	if t == nil {
		return false
	}
	for _, r := range t.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if strings.HasPrefix(*r.GetPOSTag(), pre) {
			return true
		}
	}
	return false
}

func surfaceEq(t *languagetool.AnalyzedTokenReadings, forms ...string) bool {
	if t == nil {
		return false
	}
	s := t.GetToken()
	for _, f := range forms {
		if strings.EqualFold(s, f) {
			return true
		}
	}
	return false
}

// isCapitalizedGerman ports regexCS=[A-ZÖÄÜ][a-zöäü-]+ for unknown nouns / surnames.
func isCapitalizedGerman(s string) bool {
	if s == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(s)
	if !unicode.IsUpper(r) {
		return false
	}
	rest := s[size:]
	if rest == "" {
		return false
	}
	for _, c := range rest {
		if c == '-' {
			continue
		}
		if !unicode.IsLetter(c) || !unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// --- REGEXES1 matchers: return end exclusive or -1 ---

// (<posre=^ART.*>|<pos=PRO>)? <pos=ADV>* <pos=PA2>* <pos=ADJ>* <pos=SUB>+
func matchArtProAdvPa2AdjSub(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	j := i
	if j < len(toks) && (posPrefix(toks[j], "ART") || posContains(toks[j], "PRO")) {
		j++
	}
	for j < len(toks) && posContains(toks[j], "ADV") {
		j++
	}
	for j < len(toks) && posContains(toks[j], "PA2") {
		j++
	}
	for j < len(toks) && posContains(toks[j], "ADJ") {
		j++
	}
	subStart := j
	for j < len(toks) && posContains(toks[j], "SUB") {
		j++
	}
	if j > subStart {
		return j
	}
	return -1
}

// <pos=SUB> (<und|oder>|(<bzw> <.>)) <pos=SUB> — Java buildExpanded + undOderBzw form hints.
func matchSubConjSub(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !posContains(toks[i], "SUB") {
		return -1
	}
	// und|oder
	if i+2 < len(toks) && surfaceEq(toks[i+1], "und", "oder") && posContains(toks[i+2], "SUB") {
		return i + 3
	}
	// bzw .
	if i+3 < len(toks) && surfaceEq(toks[i+1], "bzw") && toks[i+2].GetToken() == "." && posContains(toks[i+3], "SUB") {
		return i + 4
	}
	return -1
}

// <pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=PA2> <pos=SUB>
func matchAdjConjPa2Sub(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !posContains(toks[i], "ADJ") {
		return -1
	}
	if i+3 < len(toks) && surfaceEq(toks[i+1], "und", "oder") &&
		posContains(toks[i+2], "PA2") && posContains(toks[i+3], "SUB") {
		return i + 4
	}
	if i+4 < len(toks) && surfaceEq(toks[i+1], "bzw") && toks[i+2].GetToken() == "." &&
		posContains(toks[i+3], "PA2") && posContains(toks[i+4], "SUB") {
		return i + 5
	}
	return -1
}

// <pos=ADJ> (<und|oder>|(<bzw> <.>)) <pos=ADJ> <pos=SUB>
func matchAdjConjAdjSub(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !posContains(toks[i], "ADJ") {
		return -1
	}
	if i+3 < len(toks) && surfaceEq(toks[i+1], "und", "oder") &&
		posContains(toks[i+2], "ADJ") && posContains(toks[i+3], "SUB") {
		return i + 4
	}
	if i+4 < len(toks) && surfaceEq(toks[i+1], "bzw") && toks[i+2].GetToken() == "." &&
		posContains(toks[i+3], "ADJ") && posContains(toks[i+4], "SUB") {
		return i + 5
	}
	return -1
}

// <posre=^ART.*> <pos=ADV>* <pos=ADJ>* <regexCS=[A-ZÖÄÜ][a-zöäü]+>
func matchArtAdvAdjCapital(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !posPrefix(toks[i], "ART") {
		return -1
	}
	j := i + 1
	for j < len(toks) && posContains(toks[j], "ADV") {
		j++
	}
	for j < len(toks) && posContains(toks[j], "ADJ") {
		j++
	}
	if j < len(toks) && isCapitalizedGerman(toks[j].GetToken()) {
		return j + 1
	}
	return -1
}

// <pos=PRO>? <pos=ZAL> <pos=SUB>
func matchProZalSub(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	j := i
	if j < len(toks) && posContains(toks[j], "PRO") {
		j++
	}
	if j >= len(toks) || !posContains(toks[j], "ZAL") {
		return -1
	}
	j++
	if j >= len(toks) || !posContains(toks[j], "SUB") {
		return -1
	}
	return j + 1
}

// <Herr|Herrn|Frau> <pos=EIG>+
func matchTitleEig(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !surfaceEq(toks[i], "Herr", "Herrn", "Frau") {
		return -1
	}
	j := i + 1
	start := j
	for j < len(toks) && posContains(toks[j], "EIG") {
		j++
	}
	if j > start {
		return j
	}
	return -1
}

// <Herr|Herrn|Frau> <regexCS capitalized surname>+
func matchTitleCapital(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i >= len(toks) || !surfaceEq(toks[i], "Herr", "Herrn", "Frau") {
		return -1
	}
	j := i + 1
	start := j
	for j < len(toks) && isCapitalizedGerman(toks[j].GetToken()) {
		j++
	}
	if j > start {
		return j
	}
	return -1
}

// <der>
func matchDerAlone(toks []*languagetool.AnalyzedTokenReadings, i int) int {
	if i < len(toks) && surfaceEq(toks[i], "der") {
		return i + 1
	}
	return -1
}

var _ Chunker = (*GermanChunker)(nil)
