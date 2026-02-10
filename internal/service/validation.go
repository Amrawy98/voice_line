package service

import (
	"fmt"
	"strings"
)

// Groq Whisper supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm
// Max file size: 19.5MB (we default to 19MB to stay safe)
var allowedMIMETypes = map[string]bool{
	"audio/wav":    true,
	"audio/x-wav":  true,
	"audio/wave":   true,
	"audio/mpeg":   true, // mp3, mpeg, mpga
	"audio/mp4":    true, // mp4, m4a
	"audio/x-m4a":  true, // m4a
	"audio/ogg":    true,
	"audio/webm":   true,
	"audio/flac":   true,
	"audio/x-flac": true,
}

type AudioValidator struct {
	maxFileSizeBytes int64
}

func NewAudioValidator(maxFileSizeMB int) *AudioValidator {
	return &AudioValidator{
		maxFileSizeBytes: int64(maxFileSizeMB) * 1024 * 1024,
	}
}

func (v *AudioValidator) Validate(fileName string, contentType string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("file is empty")
	}

	if int64(len(data)) > v.maxFileSizeBytes {
		return fmt.Errorf("file size exceeds maximum of %d MB", v.maxFileSizeBytes/(1024*1024))
	}

	if !allowedMIMETypes[strings.ToLower(contentType)] {
		return fmt.Errorf("unsupported audio format: %s", contentType)
	}

	return nil
}