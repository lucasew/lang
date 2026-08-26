package chunking

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"math"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// POSTaggerME ports opennlp.tools.postag.POSTaggerME for OpenNLP 1.5 en-pos-maxent.bin.
type POSTaggerME struct {
	model    *GISModel
	beamSize int
	// tagDict maps token -> allowed POS tags (from tags.tagdict); nil = unrestricted.
	tagDict map[string][]string
}

// NewPOSTaggerME loads an OpenNLP POS model zip (pos.model + optional tags.tagdict).
func NewPOSTaggerME(modelPath string) (*POSTaggerME, error) {
	m, err := LoadGISModelFromZip(modelPath)
	if err != nil {
		return nil, err
	}
	dict, _ := loadPOSTagDictFromZip(modelPath) // optional
	return &POSTaggerME{model: m, beamSize: 10, tagDict: dict}, nil
}

// Tag assigns a POS tag to each token (beam search + DefaultPOSContextGenerator).
func (p *POSTaggerME) Tag(tokens []string) []string {
	if p == nil || p.model == nil || len(tokens) == 0 {
		return nil
	}
	return beamSearchPOS(p.model, tokens, p.beamSize, p.tagDict)
}

// DefaultPOSContext ports opennlp.tools.postag.DefaultPOSContextGenerator.getContext
// (dict == null path: always add suffix/prefix/shape features).
func DefaultPOSContext(i int, toks, tags []string) []string {
	const se, sb = "*SE*", "*SB*"
	lex := toks[i]
	next, nextnext := se, se
	if i+1 < len(toks) {
		next = toks[i+1]
		if i+2 < len(toks) {
			nextnext = toks[i+2]
		}
	}
	prev, prevprev := sb, sb
	var tagprev, tagprevprev string
	if i-1 >= 0 {
		prev = toks[i-1]
		if i-1 < len(tags) {
			tagprev = tags[i-1]
		}
		if i-2 >= 0 {
			prevprev = toks[i-2]
			if i-2 < len(tags) {
				tagprevprev = tags[i-2]
			}
		}
	}

	e := []string{"default", "w=" + lex}
	// dict is null for our port path when not using ngram dict — always shape features.
	for _, s := range posSuffixes(lex) {
		e = append(e, "suf="+s)
	}
	for _, s := range posPrefixes(lex) {
		e = append(e, "pre="+s)
	}
	if strings.Contains(lex, "-") {
		e = append(e, "h")
	}
	if hasCap.MatchString(lex) {
		e = append(e, "c")
	}
	if hasNum.MatchString(lex) {
		e = append(e, "d")
	}
	// surrounding context (Java: if prev != null always true after init)
	e = append(e, "p="+prev)
	if tagprev != "" {
		e = append(e, "t="+tagprev)
	}
	e = append(e, "pp="+prevprev)
	if tagprevprev != "" && tagprev != "" {
		e = append(e, "t2="+tagprevprev+","+tagprev)
	}
	e = append(e, "n="+next, "nn="+nextnext)
	return e
}

var (
	hasCap = regexp.MustCompile(`[A-Z]`)
	hasNum = regexp.MustCompile(`[0-9]`)
)

func posPrefixes(lex string) []string {
	out := make([]string, 4)
	for li := 0; li < 4; li++ {
		end := li + 1
		if end > len(lex) {
			end = len(lex)
		}
		out[li] = lex[:end]
	}
	return out
}

func posSuffixes(lex string) []string {
	out := make([]string, 4)
	for li := 0; li < 4; li++ {
		start := len(lex) - li - 1
		if start < 0 {
			start = 0
		}
		out[li] = lex[start:]
	}
	return out
}

func beamSearchPOS(model *GISModel, toks []string, beamSize int, tagDict map[string][]string) []string {
	if beamSize < 1 {
		beamSize = 10
	}
	beam := []beamSeq{{outcomes: nil, score: 0}}
	for i := 0; i < len(toks); i++ {
		var next []beamSeq
		allowed := tagDictAllowed(tagDict, toks[i])
		for _, seq := range beam {
			ctx := DefaultPOSContext(i, toks, seq.outcomes)
			probs := model.Eval(ctx)
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
				if allowed != nil && !allowed[out] {
					continue
				}
				ns := beamSeq{
					outcomes: append(append([]string(nil), seq.outcomes...), out),
					score:    seq.score + math.Log(p),
				}
				next = append(next, ns)
			}
		}
		// If tagdict filtered everything, retry without filter for this position.
		if len(next) == 0 {
			for _, seq := range beam {
				ctx := DefaultPOSContext(i, toks, seq.outcomes)
				probs := model.Eval(ctx)
				best := 0
				for oi := 1; oi < len(probs); oi++ {
					if probs[oi] > probs[best] {
						best = oi
					}
				}
				p := probs[best]
				if p <= 0 {
					p = 1e-12
				}
				ns := beamSeq{
					outcomes: append(append([]string(nil), seq.outcomes...), model.Outcome(best)),
					score:    seq.score + math.Log(p),
				}
				next = append(next, ns)
			}
		}
		sort.Slice(next, func(a, b int) bool { return next[a].score > next[b].score })
		if len(next) > beamSize {
			next = next[:beamSize]
		}
		beam = next
		if len(beam) == 0 {
			out := make([]string, len(toks))
			for j := range out {
				out[j] = "NN"
			}
			return out
		}
	}
	return beam[0].outcomes
}

func tagDictAllowed(dict map[string][]string, token string) map[string]bool {
	if dict == nil {
		return nil
	}
	tags, ok := dict[token]
	if !ok || len(tags) == 0 {
		return nil
	}
	m := make(map[string]bool, len(tags))
	for _, t := range tags {
		m[t] = true
	}
	return m
}

type posDictXML struct {
	Entries []posDictEntry `xml:"entry"`
}

type posDictEntry struct {
	Tags  string `xml:"tags,attr"`
	Token string `xml:"token"`
}

func loadPOSTagDictFromZip(path string) (map[string][]string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var f *zip.File
	for _, zf := range zr.File {
		if strings.HasSuffix(zf.Name, "tags.tagdict") || zf.Name == "tags.tagdict" {
			f = zf
			break
		}
	}
	if f == nil {
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	var doc posDictXML
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(doc.Entries))
	for _, e := range doc.Entries {
		// Token surfaces are model-dict XML; avoid Unicode TrimSpace invent.
		tok := tools.JavaStringTrim(e.Token)
		if tok == "" {
			continue
		}
		var tags []string
		// Tags are space-separated POS labels (ASCII tokenizer-like).
		for _, t := range asciiStringTokenizerSplit(e.Tags) {
			if t != "" {
				tags = append(tags, t)
			}
		}
		if len(tags) > 0 {
			out[tok] = tags
		}
	}
	return out, nil
}

// DiscoverOpenNLPPOSModel finds en-pos-maxent.bin under third_party (walk-up).
func DiscoverOpenNLPPOSModel() string {
	return walkUpFindFile(filepath.Join("third_party", "opennlp-models", "en-pos-maxent.bin"))
}

// ensure unicode used for letter checks if needed later
