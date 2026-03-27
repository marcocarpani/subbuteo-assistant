# ⚽ Subbuteo Assistant

Assistente vocale PWA per il regolamento Old Subbuteo (Rev. 2.5, 2013).

**Repository**: https://github.com/marcocarpani/subbuteo-assistant  
**Frontend live**: https://marcocarpani.github.io/subbuteo-assistant

## Stack
- **Backend**: Go + RAG minimale (chunking per capitolo + TF-IDF scoring)
- **LLM**: Google Gemini 1.5 Flash (free tier: 1500 req/giorno)
- **Frontend**: Vue 3 + Vite + vite-plugin-pwa
- **Hosting**: Railway (backend) + GitHub Pages (frontend)

## Setup locale

### 1. Backend

```bash
cd backend

# Copia il PDF del regolamento
cp /path/to/Regolamento-OldSubbuteo-2013.pdf ./regolamento.pdf

# Installa dipendenze
go mod tidy

# Avvia (serve su :8080)
GEMINI_API_KEY=la_tua_chiave go run .
```

### 2. Frontend

```bash
cd frontend
npm install

# Sviluppo (proxy verso :8080)
npm run dev

# Build produzione
VITE_API_URL=https://tuo-backend.railway.app npm run build
```

## Deploy

### Backend su Railway
1. Push su GitHub
2. Crea nuovo progetto su Railway → "Deploy from GitHub repo"
3. Railway rileva `railway.toml` automaticamente
4. Aggiungi variabile d'ambiente: `GEMINI_API_KEY=...`
5. Copia l'URL pubblico (es. `https://subbuteo-xxx.railway.app`)

### Frontend su GitHub Pages
1. Vai su repository → Settings → Pages → Source: "GitHub Actions"
2. Aggiungi secret: `VITE_API_URL=https://subbuteo-xxx.railway.app`
3. Fai push su `main` → deploy automatico

## Ottenere la Gemini API Key
1. Vai su [aistudio.google.com](https://aistudio.google.com)
2. "Get API Key" → "Create API key"
3. È gratis fino a 1500 richieste/giorno

## Struttura progetto
```
subbuteo-assistant/
├── backend/
│   ├── main.go           # HTTP server + endpoint /api/ask
│   ├── pdf/loader.go     # Estrazione testo dal PDF embedded
│   ├── rag/chunker.go    # Chunking 17 capitoli + ricerca TF-IDF
│   ├── llm/gemini.go     # Client Gemini API
│   ├── regolamento.pdf   # PDF embedded nel binario (go:embed)
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/VoiceAssistant.vue  # UI principale
│   │   ├── composables/
│   │   │   ├── useSpeechRecognition.js    # STT (Web Speech API)
│   │   │   └── useSpeechSynthesis.js      # TTS (Web Speech API)
│   │   └── main.js
│   ├── vite.config.js    # PWA configurata
│   └── package.json
├── .github/workflows/
│   └── deploy-frontend.yml
└── railway.toml
```

## Come funziona

```
[Utente parla] 
  → Web Speech API (STT, browser, gratis)
  → POST /api/ask
  → Go: cerca top-3 chunk rilevanti nel PDF (17 capitoli)
  → Gemini 1.5 Flash con contesto RAG
  → Risposta in italiano
  → Web Speech Synthesis (TTS, browser, gratis)
[Utente sente la risposta]
```
