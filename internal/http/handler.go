package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aminammar1/stockfish-go-ec2/internal/app"
	"github.com/aminammar1/stockfish-go-ec2/internal/ports"
)

type analyzeRequest struct {
	FEN string `json:"fen" example:"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"`
	PGN string `json:"pgn" example:"1. e4 e5 2. Nf3 Nc6 3. Bb5 a6"`
	UCI string `json:"uci" example:"e2e4 e7e5 g1f3 b8c6"`
	SAN string `json:"san" example:"e4 e5 Nf3 Nc6"`
}

// @Summary Health check
// @Description Checks SSH connectivity to the Stockfish EC2 instance
// @Tags Health
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /health [get]
func healthHandler(svc *app.ChessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := svc.Health(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// @Summary Analyze position
// @Description Analyze a position by providing exactly ONE of: fen, pgn, uci, san.
// @Tags Analysis
// @Accept json
// @Produce json
// @Param request body analyzeRequest true "Analyze request (exactly one of fen|pgn|uci|san)"
// @Success 200 {object} ports.AnalyzeResult
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /analyze [post]
func analyzeHandler(svc *app.ChessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req analyzeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}

		fen := strings.TrimSpace(req.FEN)
		pgn := strings.TrimSpace(req.PGN)
		uci := strings.TrimSpace(req.UCI)
		san := strings.TrimSpace(req.SAN)

		provided := 0
		if fen != "" {
			provided++
		}
		if pgn != "" {
			provided++
		}
		if uci != "" {
			provided++
		}
		if san != "" {
			provided++
		}

		if provided != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide exactly one of: fen, pgn, uci, san"})
			return
		}

		result, err := svc.Analyze(c.Request.Context(), ports.AnalyzeRequest{FEN: fen, PGN: pgn, UCIMoves: uci, SANMoves: san})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
