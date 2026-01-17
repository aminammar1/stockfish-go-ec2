package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aminammar1/stockfish-go-ec2/internal/app"
	"github.com/aminammar1/stockfish-go-ec2/internal/ports"
)

type analyzeRequest struct {
	FEN string `json:"fen"`
	PGN string `json:"pgn"`
	UCI string `json:"uci"`
	SAN string `json:"san"`
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
// @Description Sends FEN or PGN to Stockfish and returns best move + evaluation
// @Tags Analysis
// @Accept json
// @Produce json
// @Param request body analyzeRequest true "Analyze request"
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

		result, err := svc.Analyze(c.Request.Context(), ports.AnalyzeRequest{FEN: req.FEN, PGN: req.PGN, UCIMoves: req.UCI, SANMoves: req.SAN})
		if err != nil {
			if err.Error() == "fen, pgn, uci or san required" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
