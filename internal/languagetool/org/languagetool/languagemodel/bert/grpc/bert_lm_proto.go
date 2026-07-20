// Package grpc ports org.languagetool.languagemodel.bert.grpc (BertLmProto messages).
// Message types mirror bert-lm.proto / BertLmProto.java without requiring a gRPC stack.
// Wire transport (BertLmBlockingStub) is injected by callers when a channel exists.
package grpc

// Mask ports BertLmProto.Mask.
type Mask struct {
	Start      uint32
	End        uint32
	Candidates []string
}

// ScoreRequest ports BertLmProto.ScoreRequest.
type ScoreRequest struct {
	Text string
	Mask []Mask
}

// Prediction ports BertLmProto.Prediction (scores for one mask).
type Prediction struct {
	Score []float64
}

// BertLmResponse ports BertLmProto.BertLmResponse.
type BertLmResponse struct {
	Scores []Prediction
}

// FirstMaskScores ports Java getScoresList().get(0).getScoreList() with bounds checks.
// Returns nil if the response has no mask scores (fail closed — no invent ranks).
func (r *BertLmResponse) FirstMaskScores() []float64 {
	if r == nil || len(r.Scores) == 0 {
		return nil
	}
	return append([]float64(nil), r.Scores[0].Score...)
}

// BatchScoreRequest ports BertLmProto.BatchScoreRequest.
type BatchScoreRequest struct {
	Requests []ScoreRequest
}

// BatchBertLmResponse ports BertLmProto.BatchBertLmResponse.
type BatchBertLmResponse struct {
	Responses []BertLmResponse
}

// BertLmClient ports BertLmGrpc.BertLmBlockingStub surface used by RemoteLanguageModel
// (score + batchScore). Real gRPC wiring implements this; tests inject fakes.
type BertLmClient interface {
	Score(req *ScoreRequest) (*BertLmResponse, error)
	BatchScore(req *BatchScoreRequest) (*BatchBertLmResponse, error)
}

// ServiceName ports BertLmGrpc.SERVICE_NAME.
const ServiceName = "bert.BertLm"
