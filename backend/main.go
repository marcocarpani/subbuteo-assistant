package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/marcocarpani/subbuteo-assistant/llm"
	"github.com/marcocarpani/subbuteo-assistant/pdf"
	"github.com/marcocarpani/subbuteo-assistant/rag"

	"github.com/rs/cors"
)

//go:embed regolamento.pdf
var pdfFile embed.FS

var chunks []rag.Chunk

func main() {
	// 1. Estrai testo dal PDF embedded
	log.Println("📄 Caricamento regolamento...")
	pdfBytes, err := pdfFile.ReadFile("regolamento.pdf")
	if err != nil {
		log.Fatalf("Errore lettura PDF: %v", err)
	}

	text, err := pdf.ExtractText(pdfBytes)
	if err != nil {
		log.Fatalf("Errore estrazione testo: %v", err)
	}

	// 2. Chunking per capitolo
	chunks = rag.ChunkByChapter(text)
	log.Printf("✅ %d chunk caricati in memoria", len(chunks))

	// 3. HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ask", handleAsk)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// CORS per il frontend Vue
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server avviato su :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

type AskRequest struct {
	Query string `json:"query"`
}

type AskResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}

func handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AskResponse{Error: "Query mancante"})
		return
	}

	// RAG: trova i chunk più rilevanti
	topChunks := rag.Search(chunks, req.Query, 3)

	// Chiama LLM con il contesto
	apiKey := os.Getenv("GEMINI_API_KEY")
	answer, err := llm.Ask(apiKey, req.Query, topChunks)
	if err != nil {
		log.Printf("Errore LLM: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AskResponse{Error: "Errore nel servizio AI"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AskResponse{Answer: answer})
}
