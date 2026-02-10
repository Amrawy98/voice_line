# Voice Line API

Transcribes sales call audio → extracts structured insights → forwards to Notion.

**Stack:** Go + Gin, Groq Whisper, OpenRouter, Notion API

---

## Setup

**Working API keys are included in `.env.example`** (expire in 7 days).

```bash
cp .env.example .env
```

---

## Running Locally

```bash
make run
```

Server starts on http://localhost:8080

---

## Running with Docker

```bash
# Build and run
make docker-run

# Or run in background
docker-compose up -d

# Stop
make docker-down
# Or
docker-compose down
```

---

## Testing

### Health Check
```bash
curl http://localhost:8080/health
```

Expected: `{"status":"ok"}`

### Upload Audio
```bash
curl -X POST \
  -F "audio=@your-file.ogg;type=audio/ogg" \
  http://localhost:8080/api/voice-lines
```

**Supported formats:** MP3, OGG, WAV, M4A, FLAC, WEBM, MP4

**Expected response (201 Created):**
```json
{
  "transcript": "...",
  "analysis": {
    "deal_outlook": "moving_forward",
    "customer_sentiment": "positive",
    "summary": "...",
    "positive_signals": ["..."],
    "negative_signals": ["..."],
    "next_steps": ["..."],
    "deal_details": {
      "company": "...",
      "contact": "...",
      "product": "..."
    }
  }
}
```

A new page will be created in Notion: **"Sales Call - Company - Feb 10, 2026"**

**View results:** https://www.notion.so/Voice-Lines-303868eb1c9380cab58df46353609fa9