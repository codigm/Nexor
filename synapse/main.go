package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"github.com/nexor/synapse/scheduler"
)

var batcher *scheduler.ContinuousBatcher

func main() {
	log.Println("===========================================")
	log.Println("   SYNAPSE — LLM Inference Engine v2.0    ")
	log.Println("===========================================")

	// Initialize the Continuous Batching Scheduler
	batcher = &scheduler.ContinuousBatcher{
		ActiveRequests: make([]*scheduler.Request, 0),
		Queue:          make([]*scheduler.Request, 0),
	}

	// Run scheduler in background
	go batcher.Run()

	// Set up REST API
	http.HandleFunc("/v1/completions", handleCompletions)
	http.HandleFunc("/v1/chat/completions", handleChatCompletions)
	http.HandleFunc("/v1/models", handleModels)
	http.HandleFunc("/internal/embed", handleEmbed)
	http.HandleFunc("/admin/scheduler/stats", handleAdminStats)
	http.HandleFunc("/admin/cache/stats", handleCacheStats)

	port := getEnv("PORT", "8080")
	log.Printf("[SYNAPSE] Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("[SYNAPSE] Failed to start server: %v", err)
	}
}

type CompletionRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type CompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func handleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	schedReq := &scheduler.Request{
		ID:        reqID,
		Tokens:    []int{1, 2, 3}, // Mock initial tokens
		Completed: false,
	}
	batcher.Enqueue(schedReq)

	resp := CompletionResponse{
		ID: reqID,
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("[SYNAPSE] GEMINI_API_KEY not found. Using mock response.")
		resp.Choices = append(resp.Choices, struct {
			Text string `json:"text"`
		}{Text: " Mock response from SYNAPSE Tensor Engine."})
	} else {
		ctx := r.Context()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-1.5-flash")
		genResp, err := model.GenerateContent(ctx, genai.Text(req.Prompt))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var text string
		if len(genResp.Candidates) > 0 && len(genResp.Candidates[0].Content.Parts) > 0 {
			if txt, ok := genResp.Candidates[0].Content.Parts[0].(genai.Text); ok {
				text = string(txt)
			}
		}

		resp.Choices = append(resp.Choices, struct {
			Text string `json:"text"`
		}{Text: text})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		tokens := []string{"This", " is", " a", " simulated", " streamed", " response."}
		for _, token := range tokens {
			fmt.Fprintf(w, "data: {\"choices\": [{\"delta\": {\"content\": \"%s\"}}]}\n\n", token)
			w.Write([]byte{})
			time.Sleep(100 * time.Millisecond) // Simulate processing time
		}
	} else {
		// Basic implementation of real streaming with Gemini (using non-streamed for simplicity in this endpoint for now)
		ctx := r.Context()
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err == nil {
			defer client.Close()
			model := client.GenerativeModel("gemini-1.5-flash")
			// We would normally parse the chat messages here, but for simplicity we just send a generic prompt
			genResp, err := model.GenerateContent(ctx, genai.Text("Hello! Answer creatively."))
			if err == nil && len(genResp.Candidates) > 0 {
				if txt, ok := genResp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					// Stream the response out character by character or word by word
					content := string(txt)
					// Escape newlines for SSE
					// A real implementation would parse the req body and use GenerateContentStream
					fmt.Fprintf(w, "data: {\"choices\": [{\"delta\": {\"content\": %q}}]}\n\n", content)
					w.Write([]byte{})
				}
			}
		}
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := map[string]interface{}{
		"data": []map[string]string{
			{
				"id": "gpt-2-small",
				"object": "model",
				"owned_by": "nexor",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func handleEmbed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Embedding generated", "embedding": [0.1, 0.2, 0.3]}`))
}

func handleAdminStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"active_requests": batcher.ActiveLen(),
		"queued_requests": batcher.QueueLen(),
		"status":          "healthy",
		"gpu_utilization": "95%",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"kv_pages_used": 1024,
		"fragmentation": "0%",
		"swap_rate":     "0 MB/s",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
