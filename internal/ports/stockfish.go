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
	BestMove   string `json:"bestMove"`
	Evaluation string `json:"evaluation"`
	Raw        string `json:"raw"`
}

type StockfishEnginePort interface {
	Health(ctx context.Context) error
	Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResult, error)
}
