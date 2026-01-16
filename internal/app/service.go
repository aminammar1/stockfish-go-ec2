package app

import (
	"context"
	"errors"

	"stockfish-ec2-service/internal/ports"
)

type ChessService struct {
	engine ports.StockfishEnginePort
}

func NewChessService(engine ports.StockfishEnginePort) *ChessService {
	return &ChessService{engine: engine}
}

func (s *ChessService) Health(ctx context.Context) error {
	return s.engine.Health(ctx)
}

func (s *ChessService) Analyze(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResult, error) {
	if req.FEN == "" && req.PGN == "" && req.UCIMoves == "" && req.SANMoves == "" {
		return ports.AnalyzeResult{}, errors.New("fen, pgn, uci or san required")
	}
	return s.engine.Analyze(ctx, req)
}
