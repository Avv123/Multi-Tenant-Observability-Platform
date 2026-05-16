package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ── Config ────────────────────────────────────────────────────────────────────

var (
	ollamaURL   = getenv("OLLAMA_URL", "http://ollama:11434")
	ollamaModel = getenv("OLLAMA_MODEL", "llama3.2:1b")
	port        = getenv("PORT", "8085")
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ── Ollama types ──────────────────────────────────────────────────────────────

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
}

// ── System prompt ─────────────────────────────────────────────────────────────

const systemPrompt = `You are the PulseLens AI assistant. Answer concisely using markdown.
PAGES: Overview, Logs, Metrics, Traces, Alerts, Incidents, Archive, Settings.
INGESTION: POST to ingest-service with X-API-Key. { event_type: "log"|"metric"|"trace", payload: { ... } }
METRICS: requires "metric_name" and "value".
TRACES: requires "span_id" and "operation".
LOGS: requires "message".`

var fullSystemPrompt = systemPrompt

// ── Handlers ──────────────────────────────────────────────────────────────────

type ChatRequest struct {
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func chatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	fmt.Printf("[ai-service] chat request: %s\n", req.Message)

	if strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Build system prompt with dynamic context
	sp := fullSystemPrompt
	if tenantID, ok := req.Context["tenant_id"].(string); ok && tenantID != "" {
		sp += fmt.Sprintf("\n\nCurrent user tenant_id: %s", tenantID)
	}

	messages := []OllamaMessage{
		{Role: "system", Content: sp},
		{Role: "user", Content: req.Message},
	}

	ollamaReq := OllamaChatRequest{
		Model:    ollamaModel,
		Messages: messages,
		Stream:   false,
		Options: map[string]interface{}{
			"num_ctx":     2048,
			"temperature": 0.7,
		},
	}

	body, _ := json.Marshal(ollamaReq)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Post(ollamaURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ollama unreachable: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read ollama response"})
		return
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(raw, &ollamaResp); err != nil || ollamaResp.Message.Content == "" {
		// Log the raw response for debugging
		fmt.Printf("[ollama raw response] %s\n", string(raw)[:min(len(raw), 500)])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse ollama response", "raw": string(raw)[:min(len(raw), 300)]})
			return
		}
	}

	reply := ollamaResp.Message.Content
	if reply == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "empty response from model", "raw": string(raw)[:min(len(raw), 300)]})
		return
	}
	c.JSON(http.StatusOK, ChatResponse{Reply: reply})
}

func min(a, b int) int {
	if a < b { return a }
	return b
}

func statusHandler(c *gin.Context) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ollamaURL + "/api/tags")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ready": false, "reason": "ollama unreachable"})
		return
	}
	defer resp.Body.Close()

	var tags struct {
		Models []struct{ Name string `json:"name"` } `json:"models"`
	}
	json.NewDecoder(resp.Body).Decode(&tags)

	modelReady := false
	for _, m := range tags.Models {
		if strings.HasPrefix(m.Name, ollamaModel) || m.Name == ollamaModel {
			modelReady = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": modelReady,
		"model": ollamaModel,
	})
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	// Load docs context if available — truncate to ~2000 chars to stay safely inside 4096 ctx
	const maxDocsChars = 3000
	var docsContent strings.Builder
	docsContent.WriteString("\n\n=== PLATFORM DOCUMENTATION (KEY EXCERPTS) ===\n")
	total := 0
	filepath.Walk("/app/docs", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		if total >= maxDocsChars {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		excerpt := string(content)
		if total+len(excerpt) > maxDocsChars {
			excerpt = excerpt[:maxDocsChars-total] + "..."
		}
		docsContent.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", filepath.Base(path), excerpt))
		total += len(excerpt)
		return nil
	})
	fullSystemPrompt += docsContent.String()
	fmt.Printf("[ai-service] system prompt chars: %d\n", len(fullSystemPrompt))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// CORS — allow localhost UI
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api/v1/chat")
	{
		api.GET("/status", statusHandler)
		api.POST("", chatHandler)
	}

	fmt.Printf("pulselens-ai-service listening on :%s (model: %s)\n", port, ollamaModel)
	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
