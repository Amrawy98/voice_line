package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"voice_line_task/internal/service"
)

type Handler struct {
	validator   *service.AudioValidator
	transcriber *service.TranscriptionService
	analyzer    *service.AnalysisService
	forwarder   *service.ForwardingService
	maxFileSize int64
}

func NewHandler(validator *service.AudioValidator, transcriber *service.TranscriptionService, analyzer *service.AnalysisService, forwarder *service.ForwardingService, maxFileSizeMB int) *Handler {
	return &Handler{
		validator:   validator,
		transcriber: transcriber,
		analyzer:    analyzer,
		forwarder:   forwarder,
		maxFileSize: int64(maxFileSizeMB) * 1024 * 1024,
	}
}

func (h *Handler) CreateVoiceLine(c *gin.Context) {
	c.Request.ParseMultipartForm(h.maxFileSize)

	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Failed to parse audio",
		})
		return
	}
	defer file.Close()

	if header.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": "file too large",
		})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "failed to read file",
		})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if err := h.validator.Validate(header.Filename, contentType, data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	transcript, err := h.transcriber.Transcribe(c.Request.Context(), header.Filename, data)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "transcription_failed",
			"message": "failed to transcribe audio",
		})
		return
	}

	analysis, err := h.analyzer.Analyze(c.Request.Context(), transcript)
	if err != nil {
		log.Printf("Analysis error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "analysis_failed",
			"message": "failed to analyze transcript",
		})
		return
	}

	pageInput := service.PageInput{
		Summary:           analysis.Summary,
		DealOutlook:       analysis.DealOutlook,
		CustomerSentiment: analysis.CustomerSentiment,
		Company:           analysis.DealDetails.Company,
		Contact:           analysis.DealDetails.Contact,
		Product:           analysis.DealDetails.Product,
		PositiveSignals:   analysis.PositiveSignals,
		NegativeSignals:   analysis.NegativeSignals,
		NextSteps:         analysis.NextSteps,
	}

	if err := h.forwarder.CreatePage(c.Request.Context(), pageInput); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "forwarding_failed",
			"message": "failed to forward to Notion",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transcript": transcript,
		"analysis":   analysis,
	})
}