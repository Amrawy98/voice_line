package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"voice_line_task/internal/config"
	"voice_line_task/internal/handler"
	"voice_line_task/internal/validation"
)

type Server struct {
	port    int
	handler *handler.Handler
}

func NewServer(cfg config.Config) *http.Server {
	port, _ := strconv.Atoi(cfg.Port)

	validator := validation.NewAudioValidator(cfg.MaxFileSizeMB)
	h := handler.NewHandler(validator, cfg.MaxFileSizeMB)

	NewServer := &Server{
		port:    port,
		handler: h,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}