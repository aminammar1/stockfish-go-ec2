package ports

import (
	"context"
)

type AnalyzeRequest struct {
	FEN      string
	PGN      string
	UCIMoves string
	SANMoves string
}

type AnalyzeResult struct {
	BestMoveUCI    string `json:"bestMoveUci"`
	BestMoveSAN    string `json:"bestMoveSan,omitempty"`
	EvaluationCp   *int   `json:"evaluationCp,omitempty"`
	EvaluationMate *int   `json:"evaluationMate,omitempty"`
	EvalBar        *int   `json:"evalBar,omitempty"`
	Depth          int    `json:"depth,omitempty"`
	Nodes          int    `json:"nodes,omitempty"`
	NPS            int    `json:"nps,omitempty"`
	PV             string `json:"pv,omitempty"`
	PositionFEN    string `json:"positionFen,omitempty"`
	Raw            string `json:"raw,omitempty"`
}

type StockfishEnginePort interface {
	Health(ctx context.Context) error
	Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResult, error)
}
