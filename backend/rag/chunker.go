package rag

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
)

// Chunk rappresenta una sezione del regolamento
type Chunk struct {
	Title   string
	Content string
}

// I 17 capitoli del regolamento Old Subbuteo come anchor per il chunking
var chapterAnchors = []string{
	"COLPI A PUNTA DI DITO",
	"ARBITRO",
	"DURATA DEGLI INCONTRI",
	"CALCIO D'INIZIO",
	"MOVIMENTI DIFENSIVI",
	"INTERCETTAZIONE DELLA PALLA",
	"SEGNATURA DELLE RETI",
	"FALLI DI GIOCO",
	"BACK E BACK AL VOLO",
	"PORTIERE",
	"CALCIO DI RIGORE",
	"RIMESSA LATERALE",
	"CORNER",
	"RINVIO DA FONDO CAMPO",
	"FUORIGIOCO",
	"INFORTUNI ALLE MINIATURE",
	"REGOLARITÀ DELLE MINIATURE",
}

// ChunkByChapter divide il testo in chunk per capitolo
func ChunkByChapter(fullText string) []Chunk {
	// Aggiungi sempre un chunk con l'intero glossario (termini generali)
	chunks := []Chunk{}

	upperText := strings.ToUpper(fullText)
	lines := strings.Split(fullText, "\n")

	// Trova posizioni dei capitoli nel testo
	type position struct {
		idx   int // indice nel testo
		title string
	}
	var positions []position

	for _, anchor := range chapterAnchors {
		idx := strings.Index(upperText, anchor)
		if idx >= 0 {
			positions = append(positions, position{idx: idx, title: anchor})
		}
	}

	// Ordina per posizione
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].idx < positions[j].idx
	})

	// Estrai i chunk tra posizioni successive
	for i, pos := range positions {
		start := pos.idx
		end := len(fullText)
		if i+1 < len(positions) {
			end = positions[i+1].idx
		}
		content := strings.TrimSpace(fullText[start:end])
		if len(content) > 50 {
			chunks = append(chunks, Chunk{
				Title:   pos.title,
				Content: content,
			})
		}
	}

	// Fallback: se il parsing non ha trovato capitoli, usa chunk per N righe
	if len(chunks) == 0 {
		const linesPerChunk = 40
		for i := 0; i < len(lines); i += linesPerChunk {
			end := i + linesPerChunk
			if end > len(lines) {
				end = len(lines)
			}
			content := strings.Join(lines[i:end], "\n")
			if strings.TrimSpace(content) != "" {
				chunks = append(chunks, Chunk{
					Title:   fmt.Sprintf("Sezione %d", i/linesPerChunk+1),
					Content: content,
				})
			}
		}
	}

	return chunks
}

// Search restituisce i top-k chunk più rilevanti per la query
func Search(chunks []Chunk, query string, topK int) []Chunk {
	if len(chunks) == 0 {
		return nil
	}

	queryTerms := tokenize(query)

	type scored struct {
		chunk Chunk
		score float64
	}

	var results []scored
	for _, chunk := range chunks {
		score := score(chunk, queryTerms)
		results = append(results, scored{chunk, score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	top := results
	if topK < len(top) {
		top = top[:topK]
	}

	var out []Chunk
	for _, r := range top {
		out = append(out, r.chunk)
	}
	return out
}

// score calcola la rilevanza TF-IDF semplificata di un chunk per i termini della query
func score(chunk Chunk, queryTerms []string) float64 {
	text := strings.ToLower(chunk.Title + " " + chunk.Content)
	words := tokenize(text)

	// Frequenza dei termini nel chunk
	tf := make(map[string]int)
	for _, w := range words {
		tf[w]++
	}

	totalWords := float64(len(words))
	if totalWords == 0 {
		return 0
	}

	var score float64
	for _, term := range queryTerms {
		term = strings.ToLower(term)
		count := float64(tf[term])
		if count > 0 {
			// TF normalizzato + boost per match nel titolo
			termTF := count / totalWords
			titleBoost := 1.0
			if strings.Contains(strings.ToLower(chunk.Title), term) {
				titleBoost = 3.0
			}
			score += termTF * titleBoost * math.Log(1+count)
		}
	}

	return score
}

// tokenize converte il testo in token normalizzati (rimuove punteggiatura, stopwords italiane)
func tokenize(text string) []string {
	stopwords := map[string]bool{
		"il": true, "lo": true, "la": true, "i": true, "gli": true, "le": true,
		"un": true, "uno": true, "una": true, "di": true, "a": true, "da": true,
		"in": true, "con": true, "su": true, "per": true, "tra": true, "fra": true,
		"che": true, "e": true, "o": true, "ma": true, "se": true, "non": true,
		"è": true, "si": true, "del": true, "della": true, "dei": true, "delle": true,
		"al": true, "alla": true, "ai": true, "alle": true, "nel": true, "nella": true,
		"nei": true, "nelle": true, "dal": true, "dalla": true, "dai": true, "dalle": true,
	}

	var tokens []string
	text = strings.ToLower(text)
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 1 {
				word := current.String()
				if !stopwords[word] {
					tokens = append(tokens, word)
				}
			}
			current.Reset()
		}
	}
	if current.Len() > 1 {
		word := current.String()
		if !stopwords[word] {
			tokens = append(tokens, word)
		}
	}

	return tokens
}
