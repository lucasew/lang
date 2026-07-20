package chunking

import (
	"math"
	"os"
	"path/filepath"
	"sort"
)

// ChunkerME ports opennlp.tools.chunker.ChunkerME (OpenNLP 1.5 en-chunker.bin).
// Uses DefaultChunkerContextGenerator + BeamSearch over GIS maxent.
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
// (tokens, postags, previous chunk decisions). Feature names match en-chunker.bin predicates.
func DefaultChunkerContext(i int, toks, tags, preds []string) []string {
	// Previous decisions may be shorter during beam expansion.
	predAt := func(j int) string {
		if j >= 0 && j < len(preds) {
			return preds[j]
		}
		return "other"
	}
	w2, t2, p2 := "w_2=bos", "t_2=bos", "p_2=bos"
	if i-2 >= 0 {
		w2 = "w_2=" + toks[i-2]
		t2 = "t_2=" + tags[i-2]
		p2 = "p_2=" + predAt(i-2)
	}
	w1, t1, p1 := "w_1=bos", "t_1=bos", "p_1=bos"
	if i-1 >= 0 {
		w1 = "w_1=" + toks[i-1]
		t1 = "t_1=" + tags[i-1]
		p1 = "p_1=" + predAt(i-1)
	}
	w0 := "w0=" + toks[i]
	t0 := "t0=" + tags[i]
	w1n, t1n := "w1=eos", "t1=eos"
	if i+1 < len(toks) {
		w1n = "w1=" + toks[i+1]
		t1n = "t1=" + tags[i+1]
	}
	w2n, t2n := "w2=eos", "t2=eos"
	if i+2 < len(toks) {
		w2n = "w2=" + toks[i+2]
		t2n = "t2=" + tags[i+2]
	}
	// Combos are concatenated feature strings (OpenNLP predicate style).
	return []string{
		w2, w1, w0, w1n, w2n,
		t2, t1, t0, t1n, t2n,
		p2, p1,
		p1 + t1, p1 + t0, p1 + t1n,
		t1 + t0, t0 + t1n,
		w1 + w0, w0 + w1n,
		p1 + t1 + t0, p1 + t0 + t1n, t1 + t0 + t1n,
		p1 + w0,
		p2 + p1, p2 + t1, p2 + t0,
		p2 + p1 + t1, p2 + p1 + t0,
		p1 + t1 + t0 + t1n,
		p2 + p1 + t1 + t0,
	}
}

type beamSeq struct {
	outcomes []string
	score    float64
}

func beamSearchChunk(model *GISModel, toks, tags []string, beamSize int) []string {
	if beamSize < 1 {
		beamSize = 10
	}
	beam := []beamSeq{{outcomes: nil, score: 0}}
	for i := 0; i < len(toks); i++ {
		var next []beamSeq
		for _, seq := range beam {
			ctx := DefaultChunkerContext(i, toks, tags, seq.outcomes)
			probs := model.Eval(ctx)
			// threshold: min among top beamSize probabilities
			sorted := append([]float64(nil), probs...)
			sort.Float64s(sorted)
			threshIdx := len(sorted) - beamSize
			if threshIdx < 0 {
				threshIdx = 0
			}
			thresh := sorted[threshIdx]
			for oi, p := range probs {
				if p < thresh || p <= 0 {
					continue
				}
				out := model.Outcome(oi)
				ns := beamSeq{
					outcomes: append(append([]string(nil), seq.outcomes...), out),
					score:    seq.score + math.Log(p),
				}
				next = append(next, ns)
			}
		}
		// keep top beamSize by score
		sort.Slice(next, func(a, b int) bool { return next[a].score > next[b].score })
		if len(next) > beamSize {
			next = next[:beamSize]
		}
		beam = next
		if len(beam) == 0 {
			// fail-closed: O tags
			out := make([]string, len(toks))
			for j := range out {
				out[j] = "O"
			}
			return out
		}
	}
	return beam[0].outcomes
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
