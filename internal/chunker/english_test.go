package chunker

import (
	"testing"

	"github.com/lucasew/lang/internal/pipeline"
)

func TestEnglishNPChunks(t *testing.T) {
	// the big house → DT JJ NN
	toks := []pipeline.Token{
		{Text: "SENT_START", Readings: []pipeline.Reading{{POS: "SENT_START"}}},
		{Text: "the", Readings: []pipeline.Reading{{POS: "DT"}}},
		{Text: "big", Readings: []pipeline.Reading{{POS: "JJ"}}},
		{Text: "house", Readings: []pipeline.Reading{{POS: "NN"}}},
		{Text: "runs", Readings: []pipeline.Reading{{POS: "VBZ"}}},
	}
	EnglishWithMultiTags(toks)
	// NP span on the big house
	if !has(toks[1].ChunkTags, "B-NP") && !has(toks[1].ChunkTags, "B-NP-singular") {
		t.Fatalf("the chunks=%v", toks[1].ChunkTags)
	}
	if !has(toks[3].ChunkTags, "E-NP-singular") && !has(toks[3].ChunkTags, "B-NP-singular") {
		t.Fatalf("house chunks=%v", toks[3].ChunkTags)
	}
	if !has(toks[4].ChunkTags, "B-VP") {
		t.Fatalf("runs chunks=%v", toks[4].ChunkTags)
	}
}

func TestPluralNP(t *testing.T) {
	toks := []pipeline.Token{
		{Text: "SENT_START", Readings: []pipeline.Reading{{POS: "SENT_START"}}},
		{Text: "the", Readings: []pipeline.Reading{{POS: "DT"}}},
		{Text: "houses", Readings: []pipeline.Reading{{POS: "NNS", Lemma: "house"}}},
	}
	EnglishWithMultiTags(toks)
	found := false
	for _, c := range toks[1].ChunkTags {
		if c == "B-NP-plural" {
			found = true
		}
	}
	for _, c := range toks[2].ChunkTags {
		if c == "E-NP-plural" || c == "B-NP-plural" {
			found = true
		}
	}
	if !found {
		t.Fatalf("plural tags: %v %v", toks[1].ChunkTags, toks[2].ChunkTags)
	}
}

func has(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}
