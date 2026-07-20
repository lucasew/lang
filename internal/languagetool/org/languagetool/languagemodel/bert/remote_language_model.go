package bert

import (
	"fmt"
	"sync"
	"time"

	bertgrpc "github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel/bert/grpc"
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

// Convert ports Request.convert() → ScoreRequest with a single Mask.
func (r Request) Convert() *bertgrpc.ScoreRequest {
	return &bertgrpc.ScoreRequest{
		Text: r.Text,
		Mask: []bertgrpc.Mask{{
			Start:      uint32(r.Start),
			End:        uint32(r.End),
			Candidates: append([]string(nil), r.Candidates...),
		}},
	}
}

// Equal ports Request.equals.
func (r Request) Equal(o Request) bool {
	if r.Start != o.Start || r.End != o.End || r.Text != o.Text {
		return false
	}
	if len(r.Candidates) != len(o.Candidates) {
		return false
	}
	for i := range r.Candidates {
		if r.Candidates[i] != o.Candidates[i] {
			return false
		}
	}
	return true
}

// HashCode ports Request.hashCode (Objects.hash(text, start, end, candidates)).
func (r Request) HashCode() int {
	h := 1
	h = 31*h + stringHashBERT(r.Text)
	h = 31*h + r.Start
	h = 31*h + r.End
	ch := 1
	for _, c := range r.Candidates {
		ch = 31*ch + stringHashBERT(c)
	}
	h = 31*h + ch
	return h
}

func (r Request) cacheKey() string {
	return fmt.Sprintf("%s\x00%d\x00%d\x00%v", r.Text, r.Start, r.End, r.Candidates)
}

func stringHashBERT(s string) int {
	h := 0
	for _, r := range s {
		if r >= 0x10000 {
			v := r - 0x10000
			h = 31*h + int(0xD800+(v>>10))
			h = 31*h + int(0xDC00+(v&0x3FF))
		} else {
			h = 31*h + int(r)
		}
	}
	return h
}

// Scorer scores candidates for a masked span (higher is better).
// Used when Model (BertLmClient) is nil — tests / local inject.
type Scorer func(req Request) ([]float64, error)

// BatchScorer scores a batch of uncached requests (Java stub.batchScore path without Client).
// When nil, BatchScore falls back to per-request Scorer or Model.
type BatchScorer func(reqs []Request) ([][]float64, error)

// RemoteLanguageModel ports org.languagetool.languagemodel.bert.RemoteLanguageModel.
// Cache max size 1000 matches Guava CacheBuilder.maximumSize(1000).
// Transport: Model (BertLmBlockingStub twin) preferred; Scorer/BatchScorer for inject.
type RemoteLanguageModel struct {
	// Model ports BertLmGrpc.BertLmBlockingStub (score / batchScore RPCs).
	Model bertgrpc.BertLmClient
	Scorer      Scorer
	BatchScorer BatchScorer
	mu          sync.Mutex
	cache       map[string][]float64
	// MaxCache bounds entries (default 1000).
	MaxCache int
	// Host/Port/UseSSL record Java constructor channel config (for diagnostics).
	Host   string
	Port   int
	UseSSL bool
}

// NewRemoteLanguageModel builds a model with a pluggable Scorer (tests / local inject).
func NewRemoteLanguageModel(scorer Scorer) *RemoteLanguageModel {
	return &RemoteLanguageModel{
		Scorer:   scorer,
		cache:    map[string][]float64{},
		MaxCache: 1000,
	}
}

// NewRemoteLanguageModelWithClient ports RemoteLanguageModel with a BertLm stub.
func NewRemoteLanguageModelWithClient(client bertgrpc.BertLmClient) *RemoteLanguageModel {
	return &RemoteLanguageModel{
		Model:    client,
		cache:    map[string][]float64{},
		MaxCache: 1000,
	}
}

// NewRemoteLanguageModelEndpoint ports RemoteLanguageModel(host, port, useSSL, …).
// Certificate paths are accepted for API parity; without Model/Scorer, Score fails closed.
func NewRemoteLanguageModelEndpoint(host string, port int, useSSL bool, _clientKey, _clientCert, _rootCert string) *RemoteLanguageModel {
	return &RemoteLanguageModel{
		Host:     host,
		Port:     port,
		UseSSL:   useSSL,
		cache:    map[string][]float64{},
		MaxCache: 1000,
	}
}

// Score ports RemoteLanguageModel.score(Request).
// Java: model.score(req.convert()).getScoresList().get(0).getScoreList() — no cache.
func (m *RemoteLanguageModel) Score(req Request) ([]float64, error) {
	if m == nil {
		return nil, fmt.Errorf("nil RemoteLanguageModel")
	}
	if m.Model != nil {
		resp, err := m.Model.Score(req.Convert())
		if err != nil {
			return nil, err
		}
		scores := resp.FirstMaskScores()
		if scores == nil {
			return nil, fmt.Errorf("RemoteLanguageModel: empty BertLmResponse scores")
		}
		return scores, nil
	}
	if m.Scorer == nil {
		// Fail closed: Java needs a remote BERT service — do not invent edit-distance ranks.
		return nil, fmt.Errorf("RemoteLanguageModel: no Scorer or BertLmClient configured")
	}
	scores, err := m.Scorer(req)
	if err != nil {
		return nil, err
	}
	return append([]float64(nil), scores...), nil
}

func (m *RemoteLanguageModel) putCache(req Request, scores []float64) {
	if m == nil {
		return
	}
	key := req.cacheKey()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.MaxCache > 0 && len(m.cache) >= m.MaxCache {
		for k := range m.cache {
			delete(m.cache, k)
			break
		}
	}
	m.cache[key] = append([]float64(nil), scores...)
}

// BatchScore ports batchScore(requests, 0) — no deadline.
func (m *RemoteLanguageModel) BatchScore(reqs []Request) ([][]float64, error) {
	return m.BatchScoreTimeout(reqs, 0)
}

// BatchScoreTimeout ports batchScore(List<Request>, long timeoutMilliseconds).
// Separates cached vs uncached, scores uncached as a batch, merges in order, fills cache.
func (m *RemoteLanguageModel) BatchScoreTimeout(reqs []Request, timeoutMilliseconds int64) ([][]float64, error) {
	if m == nil {
		return nil, fmt.Errorf("nil RemoteLanguageModel")
	}
	cached := make(map[int][]float64)
	var uncached []Request
	for i, req := range reqs {
		key := req.cacheKey()
		m.mu.Lock()
		v, ok := m.cache[key]
		m.mu.Unlock()
		if ok {
			cached[i] = append([]float64(nil), v...)
		} else {
			uncached = append(uncached, req)
		}
	}

	var nonCache [][]float64
	if len(uncached) > 0 {
		var err error
		if timeoutMilliseconds > 0 {
			type result struct {
				scores [][]float64
				err    error
			}
			ch := make(chan result, 1)
			go func() {
				s, e := m.scoreUncached(uncached)
				ch <- result{s, e}
			}()
			select {
			case r := <-ch:
				nonCache, err = r.scores, r.err
			case <-time.After(time.Duration(timeoutMilliseconds) * time.Millisecond):
				return nil, fmt.Errorf("deadline exceeded")
			}
		} else {
			nonCache, err = m.scoreUncached(uncached)
		}
		if err != nil {
			return nil, err
		}
		for j, re := range nonCache {
			m.putCache(uncached[j], re)
		}
	}

	all := make([][]float64, len(reqs))
	j := 0
	for i := range reqs {
		if v, ok := cached[i]; ok {
			all[i] = v
		} else {
			all[i] = nonCache[j]
			j++
		}
	}
	return all, nil
}

func (m *RemoteLanguageModel) scoreUncached(uncached []Request) ([][]float64, error) {
	// Java: BatchScoreRequest from Request.convert list → stub.batchScore.
	if m.Model != nil {
		batch := &bertgrpc.BatchScoreRequest{Requests: make([]bertgrpc.ScoreRequest, 0, len(uncached))}
		for _, r := range uncached {
			if conv := r.Convert(); conv != nil {
				batch.Requests = append(batch.Requests, *conv)
			}
		}
		resp, err := m.Model.BatchScore(batch)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Responses) != len(uncached) {
			return nil, fmt.Errorf("RemoteLanguageModel: batch response size mismatch")
		}
		out := make([][]float64, len(uncached))
		for i := range uncached {
			s := resp.Responses[i].FirstMaskScores()
			if s == nil {
				return nil, fmt.Errorf("RemoteLanguageModel: empty scores at batch index %d", i)
			}
			out[i] = s
		}
		return out, nil
	}
	if m.BatchScorer != nil {
		return m.BatchScorer(uncached)
	}
	if m.Scorer == nil {
		return nil, fmt.Errorf("RemoteLanguageModel: no Scorer or BertLmClient configured")
	}
	out := make([][]float64, len(uncached))
	for i, r := range uncached {
		s, err := m.Scorer(r)
		if err != nil {
			return out, err
		}
		out[i] = s
	}
	return out, nil
}

// Shutdown ports shutdown() — clears cache (Java shuts down the gRPC channel).
func (m *RemoteLanguageModel) Shutdown() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.cache = map[string][]float64{}
	m.mu.Unlock()
}

// EditDistanceScorer ranks candidates closer to the masked original higher.
// Not used by Java RemoteLanguageModel — test/local inject only; never the default.
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
