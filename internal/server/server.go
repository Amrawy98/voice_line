package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"voice_line_task/internal/config"
	"voice_line_task/internal/handler"
	"voice_line_task/internal/service"
)

type Server struct {
	port    int
	handler *handler.Handler
}

func NewServer(cfg config.Config) *http.Server {
	port, _ := strconv.Atoi(cfg.Port)

	validator := service.NewAudioValidator(cfg.MaxFileSizeMB)
	transcriber := service.NewTranscriptionService(cfg.GroqAPIKey)
	analyzer := service.NewAnalysisService(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	h := handler.NewHandler(validator, transcriber, analyzer, cfg.MaxFileSizeMB)

	NewServer := &Server{
		port:    port,
		handler: h,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 90 * time.Second,
	}

	return server
}