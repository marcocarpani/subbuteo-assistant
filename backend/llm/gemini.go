package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/marcocarpani/subbuteo-assistant/rag"
)

const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite-preview:generateContent"

const systemPrompt = `Sei un esperto arbitro e consulente del regolamento Old Subbuteo (versione 2.5, luglio 2013).
Rispondi SEMPRE in italiano, in modo chiaro, preciso e conciso.
Basa le tue risposte ESCLUSIVAMENTE sul contesto del regolamento fornito.
Se la risposta non è nel regolamento, dillo chiaramente.
Cita il numero di regola quando possibile (es. "Secondo la regola 8.1...").
Non inventare regole che non esistono nel documento.`

type geminiRequest struct {
	Contents          []content      `json:"contents"`
	SystemInstruction systemInstruct `json:"system_instruction"`
	GenerationConfig  generationCfg  `json:"generation_config"`
}

type systemInstruct struct {
	Parts []part `json:"parts"`
}

type content struct {
	Parts []part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type part struct {
	Text string `json:"text"`
}

type generationCfg struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Ask invia una domanda a Gemini con il contesto RAG e restituisce la risposta
func Ask(apiKey, query string, chunks []rag.Chunk) (string, error) {
	// Costruisci il contesto dai chunk
	var contextBuilder strings.Builder
	contextBuilder.WriteString("=== REGOLAMENTO OLD SUBBUTEO ===\n\n")
	for _, chunk := range chunks {
		contextBuilder.WriteString("--- ")
		contextBuilder.WriteString(chunk.Title)
		contextBuilder.WriteString(" ---\n")
		contextBuilder.WriteString(chunk.Content)
		contextBuilder.WriteString("\n\n")
	}

	userMessage := fmt.Sprintf(
		"Contesto dal regolamento:\n%s\n\nDomanda: %s",
		contextBuilder.String(),
		query,
	)

	reqBody := geminiRequest{
		SystemInstruction: systemInstruct{
			Parts: []part{{Text: systemPrompt}},
		},
		Contents: []content{
			{
				Role:  "user",
				Parts: []part{{Text: userMessage}},
			},
		},
		GenerationConfig: generationCfg{
			Temperature:     0.2, // Bassa per risposte precise sul regolamento
			MaxOutputTokens: 1024,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiURL, apiKey)
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("chiamata Gemini: %w", err)
	}
	defer resp.Body.Close()

	var gemResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if gemResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", gemResp.Error.Message)
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("risposta vuota da Gemini")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}
