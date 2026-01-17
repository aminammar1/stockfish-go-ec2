package http

import (
	"github.com/gin-gonic/gin"

	"github.com/aminammar1/stockfish-go-ec2/internal/app"
)

func RegisterRoutes(r *gin.Engine, svc *app.ChessService) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler(svc))
		v1.POST("/analyze", analyzeHandler(svc))
	}
}
