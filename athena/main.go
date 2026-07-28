package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/olympus/athena/scheduler"
)

var batcher *scheduler.ContinuousBatcher

func main() {
	log.Println("===========================================")
	log.Println("   ATHENA — LLM Inference Engine v2.0     ")
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
	log.Printf("[ATHENA] Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("[ATHENA] Failed to start server: %v", err)
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

	batcher.Queue = append(batcher.Queue, schedReq)

	resp := CompletionResponse{
		ID: reqID,
	}
	resp.Choices = append(resp.Choices, struct {
		Text string `json:"text"`
	}{Text: " Mock response from ATHENA Tensor Engine."})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Simulate SSE streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	tokens := []string{"This", " is", " a", " simulated", " streamed", " response."}
	for _, token := range tokens {
		fmt.Fprintf(w, "data: {\"choices\": [{\"delta\": {\"content\": \"%s\"}}]}\n\n", token)
		w.Write([]byte{})
		time.Sleep(100 * time.Millisecond) // Simulate processing time
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
				"owned_by": "olympus",
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
		"active_requests": len(batcher.ActiveRequests),
		"queued_requests": len(batcher.Queue),
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
