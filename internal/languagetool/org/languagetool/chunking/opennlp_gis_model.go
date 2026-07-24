package chunking

import (
	"archive/zip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// OpenNLP GIS maxent model (opennlp.tools.ml.maxent.GISModel) binary load + eval.
// Used by EnglishChunker (Java ChunkerME / POSTaggerME). Path law: dependency of
// org.languagetool.chunking.EnglishChunker, not a soft invent pipeline.

type gisContext struct {
	outcomes   []int32
	parameters  []float64
}

// GISModel ports opennlp.tools.ml.maxent.GISModel for inference only.
type GISModel struct {
	outcomes   []string
	pmap       map[string]*gisContext
	numOutcomes int
	priorLog    float64 // UniformPrior: log(1/numOutcomes)
}

// LoadGISModelFromZip loads a .bin model package (manifest + *.model) as in
// OpenNLP 1.5 TokenizerModel / POSModel / ChunkerModel archives.
func LoadGISModelFromZip(path string) (*GISModel, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var modelFile *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, ".model") {
			modelFile = f
			break
		}
	}
	if modelFile == nil {
		return nil, fmt.Errorf("opennlp: no .model in %s", path)
	}
	rc, err := modelFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return readGISModel(rc)
}

func readGISModel(r io.Reader) (*GISModel, error) {
	typ, err := readJavaUTF(r)
	if err != nil {
		return nil, err
	}
	if typ != "GIS" {
		return nil, fmt.Errorf("opennlp: expected GIS model, got %q", typ)
	}
	// correction constant (unused in OpenNLP 1.5+ GISModelReader.constructModel)
	if _, err := readJavaInt(r); err != nil {
		return nil, err
	}
	if _, err := readJavaDouble(r); err != nil {
		return nil, err
	}
	nOut, err := readJavaInt(r)
	if err != nil {
		return nil, err
	}
	outcomes := make([]string, nOut)
	for i := 0; i < int(nOut); i++ {
		s, err := readJavaUTF(r)
		if err != nil {
			return nil, err
		}
		outcomes[i] = s
	}
	nPat, err := readJavaInt(r)
	if err != nil {
		return nil, err
	}
	pats := make([][]int, nPat)
	for i := 0; i < int(nPat); i++ {
		s, err := readJavaUTF(r)
		if err != nil {
			return nil, err
		}
		// OpenNLP AbstractModelReader: StringTokenizer(pattern) default delim
		// " \t\n\r\f" — collapse consecutive, not Unicode Fields / NBSP.
		for _, f := range asciiStringTokenizerSplit(s) {
			v, err := strconv.Atoi(f)
			if err != nil {
				return nil, err
			}
			pats[i] = append(pats[i], v)
		}
	}
	nPred, err := readJavaInt(r)
	if err != nil {
		return nil, err
	}
	preds := make([]string, nPred)
	for i := 0; i < int(nPred); i++ {
		s, err := readJavaUTF(r)
		if err != nil {
			return nil, err
		}
		preds[i] = s
	}
	// getParameters: for each outcome pattern, pattern[0] contexts each with (len-1) doubles
	params := make([]*gisContext, nPred)
	idx := 0
	for _, pat := range pats {
		if len(pat) < 1 {
			return nil, fmt.Errorf("opennlp: empty outcome pattern")
		}
		oc := make([]int32, len(pat)-1)
		for j := 1; j < len(pat); j++ {
			oc[j-1] = int32(pat[j])
		}
		for c := 0; c < pat[0]; c++ {
			ps := make([]float64, len(oc))
			for j := range ps {
				ps[j], err = readJavaDouble(r)
				if err != nil {
					return nil, err
				}
			}
			if idx >= len(params) {
				return nil, fmt.Errorf("opennlp: parameter index overflow")
			}
			params[idx] = &gisContext{outcomes: oc, parameters: ps}
			idx++
		}
	}
	m := &GISModel{
		outcomes:    outcomes,
		pmap:        make(map[string]*gisContext, nPred),
		numOutcomes: int(nOut),
		priorLog:    math.Log(1.0 / float64(nOut)),
	}
	for i, name := range preds {
		m.pmap[name] = params[i]
	}
	return m, nil
}

// Eval ports GISModel.eval(String[]): prior + sum features + softmax.
func (m *GISModel) Eval(context []string) []float64 {
	if m == nil {
		return nil
	}
	out := make([]float64, m.numOutcomes)
	for i := range out {
		out[i] = m.priorLog
	}
	for _, c := range context {
		ctx := m.pmap[c]
		if ctx == nil {
			continue
		}
		for j, oi := range ctx.outcomes {
			if int(oi) < len(out) {
				out[oi] += ctx.parameters[j]
			}
		}
	}
	var sum float64
	for i, v := range out {
		e := math.Exp(v)
		out[i] = e
		sum += e
	}
	if sum > 0 {
		for i := range out {
			out[i] /= sum
		}
	}
	return out
}

func (m *GISModel) Outcome(i int) string {
	if m == nil || i < 0 || i >= len(m.outcomes) {
		return ""
	}
	return m.outcomes[i]
}

// BestOutcome returns the outcome name with the highest probability.
func (m *GISModel) BestOutcome(probs []float64) string {
	if m == nil || len(probs) == 0 {
		return ""
	}
	best := 0
	for i := 1; i < len(probs) && i < len(m.outcomes); i++ {
		if probs[i] > probs[best] {
			best = i
		}
	}
	return m.outcomes[best]
}

// OutcomeIndex returns the index of outcome name, or -1.
func (m *GISModel) OutcomeIndex(name string) int {
	if m == nil {
		return -1
	}
	for i, o := range m.outcomes {
		if o == name {
			return i
		}
	}
	return -1
}

func (m *GISModel) NumOutcomes() int {
	if m == nil {
		return 0
	}
	return m.numOutcomes
}

// Outcomes returns a copy of outcome names (for POS/chunk inventories).
func (m *GISModel) Outcomes() []string {
	if m == nil {
		return nil
	}
	out := make([]string, len(m.outcomes))
	copy(out, m.outcomes)
	return out
}

func readJavaUTF(r io.Reader) (string, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return "", err
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	// Model files use ASCII feature labels; modified UTF-8 ≡ UTF-8 for those.
	return string(b), nil
}

func readJavaInt(r io.Reader) (int32, error) {
	var v int32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

func readJavaDouble(r io.Reader) (float64, error) {
	var v float64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

// asciiStringTokenizerSplit ports java.util.StringTokenizer default delimiters
// " \t\n\r\f": consecutive delimiters collapse; no empty tokens; not Unicode Zs.
func asciiStringTokenizerSplit(s string) []string {
	var out []string
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			out = append(out, s[start:end])
		}
		start = -1
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' {
			flush(i)
			continue
		}
		if start < 0 {
			start = i
		}
	}
	flush(len(s))
	return out
}
