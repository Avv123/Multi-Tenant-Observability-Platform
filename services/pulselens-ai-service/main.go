package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ── Config ────────────────────────────────────────────────────────────────────

var (
	ollamaURL   = getenv("OLLAMA_URL", "http://ollama:11434")
	ollamaModel = getenv("OLLAMA_MODEL", "llama3.2:3b")
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
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
}

// ── System prompt ─────────────────────────────────────────────────────────────

const systemPrompt = `You are the PulseLens AI assistant. PulseLens is a multi-tenant SaaS observability platform.

PLATFORM OVERVIEW:
- Tenants are isolated workspaces. Each has users, API keys, services, and their own telemetry data.
- Telemetry signal types: Logs (structured JSON), Metrics (numeric gauges/counters), Traces (distributed spans).
- API keys authenticate ingestion (X-API-Key header) and queries (Authorization: Bearer <JWT>).

KEY PAGES:
- Overview: high-level KPIs — log volumes, error rate, active incidents.
- Logs: search and filter log events by severity, service, environment, and time.
- Metrics: time-series charts for numeric telemetry (CPU, latency, custom counters).
- Traces: distributed request spans — search by trace_id, service, operation.
- Alerts: configure detection rules (signal + threshold + window) and notification channels (webhook/Slack/email).
- Incidents: created automatically when an alert rule fires; acknowledge or resolve here.
- Archive: cold-storage log archive in MinIO; replay jobs restore data to ClickHouse for historical queries.
- Settings: manage team users, API keys (rotate/revoke), and view the audit log.
- Infrastructure (admin only): view all tenants, platform service health.
- Setup (bootstrap): first-time wizard to create a workspace and generate an API key.

INGESTION:
- Send events via POST to the ingest service with X-API-Key header.
- Event format: { event_type: "log"|"metric"|"trace", payload: { ... } }
- Log payload: { severity, message, service_name, environment, ... }
- Metric payload: { metric_name, value, service_name, ... }
- Trace payload: { trace_id, span_id, parent_span_id, operation, status, start_time, end_time }

ALERT RULES:
- Signal types: log, metric, trace
- Aggregations: count, avg, sum, min, max
- Comparators: >=, >, <=, <, ==
- Window: rolling minutes to evaluate
- Cooldown: minimum minutes between repeated firings (prevents alert storms)

API KEYS:
- Rotate: generates new secret; old key immediately invalidated in Redis cache
- Revoke: permanently deactivates; services get 401 errors immediately
- Scopes: "ingest" for sending data, "query" for reading data

Answer concisely and helpfully. When referencing UI features, use exact page names. If the user asks something you do not know, say so honestly.`

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

	if strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Build system prompt with dynamic context
	sp := systemPrompt
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
	}

	body, _ := json.Marshal(ollamaReq)

	client := &http.Client{Timeout: 120 * time.Second}
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
	if err := json.Unmarshal(raw, &ollamaResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse ollama response", "raw": string(raw)})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{Reply: ollamaResp.Message.Content})
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
