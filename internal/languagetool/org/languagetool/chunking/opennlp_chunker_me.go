package chunking

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ChunkerME ports opennlp.tools.chunker.ChunkerME (OpenNLP 1.5 en-chunker.bin).
// Uses DefaultChunkerContextGenerator + BeamSearch + DefaultChunkerSequenceValidator.
type ChunkerME struct {
	model    *GISModel
	beamSize int
}

// NewChunkerME loads an OpenNLP 1.5 chunker model zip (en-chunker.bin).
func NewChunkerME(modelPath string) (*ChunkerME, error) {
	m, err := LoadGISModelFromZip(modelPath)
	if err != nil {
		return nil, err
	}
	return &ChunkerME{model: m, beamSize: 10}, nil
}

// Chunk assigns BIO chunk tags (B-NP, I-VP, O, …) given tokens and POS tags.
func (c *ChunkerME) Chunk(tokens, posTags []string) []string {
	if c == nil || c.model == nil || len(tokens) == 0 || len(tokens) != len(posTags) {
		return nil
	}
	return beamSearchChunk(c.model, tokens, posTags, c.beamSize)
}

// DefaultChunkerContext ports opennlp.tools.chunker.DefaultChunkerContextGenerator.getContext
// (OpenNLP 1.5 / 1.9 — feature order and p_2 quirk without '=' when not bos).
func DefaultChunkerContext(i int, toks, tags, preds []string) []string {
	// Words in a 5-word window
	var w_2, w_1, w0, w1, w2 string
	// Tags in a 5-word window
	var t_2, t_1, t0, t1, t2 string
	// Previous predictions
	var p_2, p_1 string

	if i < 2 {
		w_2 = "w_2=bos"
		t_2 = "t_2=bos"
		p_2 = "p_2=bos"
	} else {
		w_2 = "w_2=" + toks[i-2]
		t_2 = "t_2=" + tags[i-2]
		// Java quirk: no '=' when not bos — "p_2" + preds[i-2]
		p_2 = "p_2" + preds[i-2]
	}

	if i < 1 {
		w_1 = "w_1=bos"
		t_1 = "t_1=bos"
		p_1 = "p_1=bos"
	} else {
		w_1 = "w_1=" + toks[i-1]
		t_1 = "t_1=" + tags[i-1]
		p_1 = "p_1=" + preds[i-1]
	}

	w0 = "w0=" + toks[i]
	t0 = "t0=" + tags[i]

	if i+1 >= len(toks) {
		w1 = "w1=eos"
		t1 = "t1=eos"
	} else {
		w1 = "w1=" + toks[i+1]
		t1 = "t1=" + tags[i+1]
	}

	if i+2 >= len(toks) {
		w2 = "w2=eos"
		t2 = "t2=eos"
	} else {
		w2 = "w2=" + toks[i+2]
		t2 = "t2=" + tags[i+2]
	}

	// Feature order matches Java DefaultChunkerContextGenerator exactly.
	return []string{
		// word features
		w_2, w_1, w0, w1, w2,
		w_1 + w0, w0 + w1,
		// tag features
		t_2, t_1, t0, t1, t2,
		t_2 + t_1, t_1 + t0, t0 + t1, t1 + t2,
		t_2 + t_1 + t0, t_1 + t0 + t1, t0 + t1 + t2,
		// pred tags
		p_2, p_1, p_2 + p_1,
		// pred and tag
		p_1 + t_2, p_1 + t_1, p_1 + t0, p_1 + t1, p_1 + t2,
		p_1 + t_2 + t_1, p_1 + t_1 + t0, p_1 + t0 + t1, p_1 + t1 + t2,
		p_1 + t_2 + t_1 + t0, p_1 + t_1 + t0 + t1, p_1 + t0 + t1 + t2,
		// pred and word
		p_1 + w_2, p_1 + w_1, p_1 + w0, p_1 + w1, p_1 + w2,
		p_1 + w_1 + w0, p_1 + w0 + w1,
	}
}

// validChunkOutcome ports DefaultChunkerSequenceValidator.validOutcome.
func validChunkOutcome(outcome string, prevSequence []string) bool {
	if !strings.HasPrefix(outcome, "I-") {
		return true
	}
	var prevOutcome string
	if len(prevSequence) > 0 {
		prevOutcome = prevSequence[len(prevSequence)-1]
	} else {
		return false
	}
	if prevOutcome == "O" {
		return false
	}
	// prevOutcome.substring(2).equals(outcome.substring(2))
	// B-NP / I-NP share type after first 2 chars; also I-NP continues I-NP.
	if len(prevOutcome) < 2 || len(outcome) < 2 {
		return false
	}
	return prevOutcome[2:] == outcome[2:]
}

type beamSeq struct {
	outcomes []string
	score    float64
}

// beamSearchChunk ports opennlp.tools.ml.BeamSearch.bestSequences for the chunker
// (cacheSize=0, DefaultChunkerSequenceValidator, minSequenceScore=zeroLog=-100000).
func beamSearchChunk(model *GISModel, toks, tags []string, beamSize int) []string {
	if beamSize < 1 {
		beamSize = 10
	}
	const zeroLog = -100000.0
	prev := []beamSeq{{outcomes: nil, score: 0}}
	for i := 0; i < len(toks); i++ {
		// Sort prev by score descending so we expand top beamSize (Java PriorityQueue / ListHeap).
		sort.Slice(prev, func(a, b int) bool { return prev[a].score > prev[b].score })
		sz := beamSize
		if len(prev) < sz {
			sz = len(prev)
		}
		var next []beamSeq
		for sc := 0; sc < sz; sc++ {
			seq := prev[sc]
			ctx := DefaultChunkerContext(i, toks, tags, seq.outcomes)
			probs := model.Eval(ctx)
			if len(probs) == 0 {
				continue
			}
			temp := append([]float64(nil), probs...)
			sort.Float64s(temp)
			minIdx := len(temp) - beamSize
			if minIdx < 0 {
				minIdx = 0
			}
			min := temp[minIdx]
			advanced := 0
			for p, score := range probs {
				if score < min {
					continue
				}
				out := model.Outcome(p)
				if !validChunkOutcome(out, seq.outcomes) {
					continue
				}
				// Java Sequence: score += log(p); skip if log(p) would be -Inf for p<=0
				if score <= 0 {
					continue
				}
				ns := beamSeq{
					outcomes: append(append([]string(nil), seq.outcomes...), out),
					score:    seq.score + math.Log(score),
				}
				if ns.score > zeroLog {
					next = append(next, ns)
					advanced++
				}
			}
			// if no advanced sequences, advance all valid (Java BeamSearch fallback)
			if advanced == 0 && len(next) == 0 {
				for p, score := range probs {
					out := model.Outcome(p)
					if !validChunkOutcome(out, seq.outcomes) {
						continue
					}
					if score <= 0 {
						continue
					}
					ns := beamSeq{
						outcomes: append(append([]string(nil), seq.outcomes...), out),
						score:    seq.score + math.Log(score),
					}
					if ns.score > zeroLog {
						next = append(next, ns)
					}
				}
			}
		}
		// Keep top beamSize (ListHeap size limit / PriorityQueue expand window).
		sort.Slice(next, func(a, b int) bool { return next[a].score > next[b].score })
		if len(next) > beamSize {
			next = next[:beamSize]
		}
		prev = next
		if len(prev) == 0 {
			out := make([]string, len(toks))
			for j := range out {
				out[j] = "O"
			}
			return out
		}
	}
	sort.Slice(prev, func(a, b int) bool { return prev[a].score > prev[b].score })
	return prev[0].outcomes
}

// DiscoverOpenNLPChunkerModel finds en-chunker.bin under third_party (walk-up).
func DiscoverOpenNLPChunkerModel() string {
	rel := filepath.Join("third_party", "opennlp-models", "en-chunker.bin")
	return walkUpFindFile(rel)
}

func walkUpFindFile(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
