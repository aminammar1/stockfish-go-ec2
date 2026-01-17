package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/aminammar1/stockfish-go-ec2/docs"
	"github.com/aminammar1/stockfish-go-ec2/internal/adapters/stockfish_ssh"
	"github.com/aminammar1/stockfish-go-ec2/internal/app"
	"github.com/aminammar1/stockfish-go-ec2/internal/config"
	httpadapter "github.com/aminammar1/stockfish-go-ec2/internal/http"
)

// @title stockfish-ec2-service API
// @version 1.0
// @description Hexagonal service that proxies Stockfish over SSH.
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	adapter := stockfish_ssh.NewAdapter(cfg)
	service := app.NewChessService(adapter)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpadapter.RegisterRoutes(r, service)

	docs.SwaggerInfo.Title = "stockfish-ec2-service API"
	docs.SwaggerInfo.Description = "Hexagonal service that proxies Stockfish over SSH."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	addr := ":" + cfg.ServerPort
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
