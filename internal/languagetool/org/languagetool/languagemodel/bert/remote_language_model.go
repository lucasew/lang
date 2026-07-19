package bert

import (
	"fmt"
	"sync"
)

// Request ports RemoteLanguageModel.Request — mask a span with candidate fill-ins.
type Request struct {
	Text       string
	Start      int
	End        int
	Candidates []string
}

func NewRequest(text string, start, end int, candidates []string) Request {
	return Request{
		Text:       text,
		Start:      start,
		End:        end,
		Candidates: append([]string(nil), candidates...),
	}
}

func (r Request) cacheKey() string {
	return fmt.Sprintf("%s\x00%d\x00%d\x00%v", r.Text, r.Start, r.End, r.Candidates)
}

// Scorer scores candidates for a masked span (higher is better).
// Remote gRPC is deferred; inject HTTP/local models here.
type Scorer func(req Request) ([]float64, error)

// RemoteLanguageModel ports org.languagetool.languagemodel.bert.RemoteLanguageModel
// with a pluggable Scorer and in-memory cache (no gRPC dependency).
type RemoteLanguageModel struct {
	Scorer Scorer
	mu     sync.Mutex
	cache  map[string][]float64
	// MaxCache bounds entries (default 1000).
	MaxCache int
}

func NewRemoteLanguageModel(scorer Scorer) *RemoteLanguageModel {
	return &RemoteLanguageModel{
		Scorer:   scorer,
		cache:    map[string][]float64{},
		MaxCache: 1000,
	}
}

// Score returns one score per candidate.
func (m *RemoteLanguageModel) Score(req Request) ([]float64, error) {
	if m == nil {
		return nil, fmt.Errorf("nil RemoteLanguageModel")
	}
	key := req.cacheKey()
	m.mu.Lock()
	if v, ok := m.cache[key]; ok {
		m.mu.Unlock()
		return append([]float64(nil), v...), nil
	}
	m.mu.Unlock()

	if m.Scorer == nil {
		// Fail closed: Java needs a remote BERT service — do not invent edit-distance ranks.
		// Tests/callers may pass EditDistanceScorer explicitly when that stand-in is intended.
		return nil, fmt.Errorf("RemoteLanguageModel: no Scorer configured")
	}
	scores, err := m.Scorer(req)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	if m.MaxCache > 0 && len(m.cache) >= m.MaxCache {
		// drop one
		for k := range m.cache {
			delete(m.cache, k)
			break
		}
	}
	m.cache[key] = append([]float64(nil), scores...)
	m.mu.Unlock()
	return scores, nil
}

// BatchScore scores many requests, using the cache when possible.
func (m *RemoteLanguageModel) BatchScore(reqs []Request) ([][]float64, error) {
	out := make([][]float64, len(reqs))
	for i, r := range reqs {
		s, err := m.Score(r)
		if err != nil {
			return out, err
		}
		out[i] = s
	}
	return out, nil
}

// Shutdown clears resources (cache).
func (m *RemoteLanguageModel) Shutdown() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.cache = map[string][]float64{}
	m.mu.Unlock()
}

// EditDistanceScorer ranks candidates closer to the masked original higher.
func EditDistanceScorer(req Request) ([]float64, error) {
	return editDistanceScores(req), nil
}

func editDistanceScores(req Request) []float64 {
	orig := ""
	if req.Start >= 0 && req.End <= len(req.Text) && req.Start < req.End {
		orig = req.Text[req.Start:req.End]
	}
	out := make([]float64, len(req.Candidates))
	for i, c := range req.Candidates {
		d := lev(orig, c)
		out[i] = 1.0 / float64(1+d)
	}
	return out
}

func lev(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}
