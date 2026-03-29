<template>

  <div class="app" :class="{ 'menu-open': menuOpen }">

    <!-- ── OVERLAY ── -->
    <div class="overlay" @click="menuOpen = false"></div>

    <!-- ── MENU LATERALE CAPITOLI ── -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="sidebar-title">
          <span class="sidebar-icon">📖</span>
          <div>
            <p class="sidebar-label">Old Subbuteo</p>
            <p class="sidebar-sub">Rev. 2.5 · 2013</p>
          </div>
        </div>
        <button class="close-btn" @click="menuOpen = false" aria-label="Chiudi menu">✕</button>
      </div>

      <nav class="chapter-list">
        <button
          v-for="ch in chapters"
          :key="ch.id"
          class="chapter-item"
          :class="{ active: activeChapter === ch.id }"
          @click="selectChapter(ch)"
        >
          <span class="chapter-num">{{ ch.id }}</span>
          <span class="chapter-title">{{ ch.title }}</span>
        </button>
      </nav>

      <div class="sidebar-footer">
        <span>⚽ Flick with us</span>
      </div>
    </aside>

    <!-- ── MAIN ── -->
    <div class="main">

      <!-- Header -->
      <header>
        <button class="menu-btn" @click="menuOpen = !menuOpen" aria-label="Apri menu">
          <span class="hamburger" :class="{ open: menuOpen }">
            <span></span><span></span><span></span>
          </span>
        </button>
        <div class="logo-area">
          <span class="logo">⚽</span>
          <div>
            <h1>Subbuteo <span>Assistant</span></h1>
            <p class="tagline">Regolamento Old Subbuteo</p>
          </div>
        </div>
      </header>

      <!-- Content -->
      <main>
        <div class="response-box" :class="{ empty: !answer && !isLoading && !activeChapter }">

          <!-- Loading -->
          <div v-if="isLoading" class="loading">
            <span class="dot"></span><span class="dot"></span><span class="dot"></span>
            <span class="loading-text">Consultando il regolamento...</span>
          </div>

          <!-- Risposta LLM -->
          <div v-else-if="answer" class="answer">
            <div class="answer-meta">
              <span class="answer-icon">🎤</span>
              <span class="answer-query">{{ lastQuery }}</span>
            </div>
            <!-- markdown renderizzato come HTML con stili corretti -->
            <div class="answer-body markdown-content" v-html="renderedAnswer"></div>
            <div class="answer-actions">
              <button class="action-btn" @click="toggleSpeak" :class="{ active: isSpeaking }">
                <span>{{ isSpeaking ? '⏹' : '🔊' }}</span>
                {{ isSpeaking ? 'Ferma audio' : 'Leggi ad alta voce' }}
              </button>
              <button class="action-btn secondary" @click="clearAnswer">
                <span>✕</span> Nuova domanda
              </button>
            </div>
          </div>

          <!-- Placeholder iniziale -->
          <div v-else class="placeholder">
            <div class="placeholder-icon">🏟️</div>
            <p class="placeholder-main">Chiedi una regola del Subbuteo</p>
            <div class="examples">
              <button class="example-chip" @click="askExample('Quanti tocchi può fare una miniatura?')">
                Quanti tocchi può fare una miniatura?
              </button>
              <button class="example-chip" @click="askExample('Come si batte il calcio di rigore?')">
                Come si batte il calcio di rigore?
              </button>
              <button class="example-chip" @click="askExample('Quando scatta il fuorigioco?')">
                Quando scatta il fuorigioco?
              </button>
              <button class="example-chip" @click="askExample('Cos\'è il back al volo?')">
                Cos'è il back al volo?
              </button>
            </div>
            <p class="placeholder-hint">oppure sfoglia i capitoli con ☰</p>
          </div>

        </div>

        <div v-if="errorMsg" class="error-msg">⚠️ {{ errorMsg }}</div>
      </main>

      <!-- Trascrizione live -->
      <div v-if="isListening || liveTranscript" class="transcript-box">
        <span v-if="isListening" class="mic-pulse">🎙️</span>
        <span class="transcript-text">{{ liveTranscript || 'In ascolto...' }}</span>
      </div>

      <!-- Microfono -->
      <footer>
        <button
          class="mic-btn"
          :class="{ listening: isListening, loading: isLoading }"
          @click="handleMicClick"
          :disabled="isLoading"
        >
          <div class="mic-rings" v-if="isListening">
            <span></span><span></span>
          </div>
          <span class="mic-icon">{{ isListening ? '⏹' : '🎤' }}</span>
          <span class="mic-label">{{ micLabel }}</span>
        </button>
      </footer>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { marked } from 'marked'
import { useSpeechRecognition } from '../composables/useSpeechRecognition.js'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis.js'

const {
  transcript,
  isListening,
  error: speechError,
  startListening,
  stopListening
} = useSpeechRecognition()

const { speak, stop, isSpeaking } = useSpeechSynthesis()

// ── Stato ──────────────────────────────────────────────────────
const answer         = ref('')
const lastQuery      = ref('')
const isLoading      = ref(false)
const errorMsg       = ref('')
const menuOpen       = ref(false)
const activeChapter  = ref(null)
const liveTranscript = ref('')

// ── Marked: configurazione renderer ───────────────────────────
marked.setOptions({
  breaks: true,   // \n → <br>
  gfm: true,      // GitHub Flavored Markdown
})

// HTML renderizzato per la UI
const renderedAnswer = computed(() => {
  if (!answer.value) return ''
  return marked.parse(answer.value)
})

// ── Strip markdown per TTS ─────────────────────────────────────
// Produce testo plain parlabile, senza simboli markdown
function toSpeakable(text) {
  return text
    .replace(/#{1,6}\s?(.+)/g, '$1.')         // ## Titolo → "Titolo."
    .replace(/\*\*(.+?)\*\*/g, '$1')          // **bold** → bold
    .replace(/\*(.+?)\*/g, '$1')              // *italic* → italic
    .replace(/^\s*[-*•]\s+(.+)/gm, '$1. ')   // - item → "item. "
    .replace(/^\s*\d+\.\s+(.+)/gm, '$1. ')   // 1. item → "item. "
    .replace(/`(.+?)`/g, '$1')               // `code` → code
    .replace(/_{1,2}(.+?)_{1,2}/g, '$1')     // _italic_ → italic
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')      // [link](url) → link
    .replace(/>\s?(.+)/g, '$1')              // > quote → quote
    .replace(/---+/g, '')                    // --- separatori
    .replace(/\n{2,}/g, '. ')               // paragrafi → pausa
    .replace(/\n/g, ' ')                    // a capo → spazio
    .replace(/\s{2,}/g, ' ')               // spazi multipli → uno
    .trim()
}

// ── 17 Capitoli Old Subbuteo ───────────────────────────────────
const chapters = [
  { id: 1,  title: 'Colpi a punta di dito',                  icon: '👆' },
  { id: 2,  title: 'Arbitro',                                icon: '🟨' },
  { id: 3,  title: 'Durata degli incontri',                  icon: '⏱️' },
  { id: 4,  title: "Calcio d'inizio e possesso palla",       icon: '⚽' },
  { id: 5,  title: 'Movimenti difensivi',                    icon: '🛡️' },
  { id: 6,  title: 'Intercettazione della palla',            icon: '🖐️' },
  { id: 7,  title: 'Segnatura delle reti',                   icon: '🥅' },
  { id: 8,  title: 'Falli di gioco e calci di punizione',    icon: '🚨' },
  { id: 9,  title: 'Back e back al volo',                    icon: '↩️' },
  { id: 10, title: 'Portiere',                               icon: '🧤' },
  { id: 11, title: 'Calcio di rigore e shoot-out',           icon: '🎯' },
  { id: 12, title: 'Rimessa laterale',                       icon: '↗️' },
  { id: 13, title: 'Corner',                                 icon: '🚩' },
  { id: 14, title: 'Rinvio da fondo campo',                  icon: '🦵' },
  { id: 15, title: 'Fuorigioco',                             icon: '🚫' },
  { id: 16, title: 'Infortuni e rottura materiali',          icon: '🩹' },
  { id: 17, title: 'Regolarità delle miniature',             icon: '📏' },
]

// ── Seleziona capitolo dal menu ────────────────────────────────
async function selectChapter(ch) {
  menuOpen.value     = false
  activeChapter.value = ch.id
  await askBackend(`Spiega in dettaglio le regole del capitolo "${ch.title}" del regolamento Old Subbuteo. Includi tutte le regole numerate, le note e le casistiche principali.`)
  lastQuery.value = `${ch.icon} Capitolo ${ch.id}: ${ch.title}`
}

// ── Chip di esempio ────────────────────────────────────────────
async function askExample(query) {
  activeChapter.value = null
  await askBackend(query)
}

// ── Chiamata backend ───────────────────────────────────────────
async function askBackend(query) {
  isLoading.value = true
  answer.value    = ''
  errorMsg.value  = ''
  lastQuery.value = query

  try {
    const base = import.meta.env.VITE_API_URL || ''
    const res  = await fetch(`${base}/api/ask`, {
      method:  'POST',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ query })
    })
    const data = await res.json()
    if (data.error) {
      errorMsg.value = data.error
    } else {
      answer.value = data.answer
      speak(toSpeakable(data.answer)) // TTS: solo plain text
    }
  } catch {
    errorMsg.value = 'Impossibile contattare il server. Controlla la connessione.'
  } finally {
    isLoading.value = false
  }
}

function clearAnswer() {
  answer.value        = ''
  lastQuery.value     = ''
  activeChapter.value = null
  stop()
}

function toggleSpeak() {
  isSpeaking.value ? stop() : speak(toSpeakable(answer.value))
}

const micLabel = computed(() => {
  if (isLoading.value)   return 'Elaborazione...'
  if (isListening.value) return 'Tocca per fermare'
  return 'Tocca per parlare'
})

function handleMicClick() {
  if (isListening.value) { stopListening(); return }
  stop()
  errorMsg.value      = ''
  activeChapter.value = null
  liveTranscript.value = ''
  startListening()
}

// Aggiorna trascrizione live mentre parla
watch(transcript, (val) => { liveTranscript.value = val })

// Quando finisce di parlare → invia al backend
watch(isListening, (listening) => {
  if (!listening && liveTranscript.value) {
    askBackend(liveTranscript.value)
    lastQuery.value = liveTranscript.value
  }
})

watch(speechError, (err) => { if (err) errorMsg.value = err })
</script>

<style scoped>
/* ── Base ──────────────────────────────────────────── */
.app {
  display: flex;
  min-height: 100svh;
  overflow: hidden;
  font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
}

/* ── Overlay ───────────────────────────────────────── */
.overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(2px);
  z-index: 10;
  transition: opacity 0.3s;
}
.app.menu-open .overlay { display: block; }

/* ── Sidebar ───────────────────────────────────────── */
.sidebar {
  position: fixed;
  inset: 0 auto 0 0;
  width: 300px;
  background: linear-gradient(180deg, #071f0b 0%, #0d3314 100%);
  z-index: 20;
  display: flex;
  flex-direction: column;
  transform: translateX(-100%);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 6px 0 30px rgba(0,0,0,0.5);
}
.app.menu-open .sidebar { transform: translateX(0); }

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 16px 16px;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}
.sidebar-title { display: flex; align-items: center; gap: 10px; }
.sidebar-icon  { font-size: 1.6rem; }
.sidebar-label { font-weight: 700; color: #a8e6b3; font-size: 0.95rem; }
.sidebar-sub   { font-size: 0.72rem; color: rgba(255,255,255,0.4); margin-top: 2px; }

.close-btn {
  background: rgba(255,255,255,0.08);
  border: none; color: rgba(255,255,255,0.6);
  width: 32px; height: 32px; border-radius: 8px;
  cursor: pointer; font-size: 1rem;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s;
}
.close-btn:hover { background: rgba(255,255,255,0.15); }

.chapter-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 8px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255,255,255,0.1) transparent;
}

.chapter-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 10px;
  background: none;
  border: none;
  border-radius: 10px;
  color: rgba(255,255,255,0.75);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s;
  margin-bottom: 2px;
}
.chapter-item:hover {
  background: rgba(255,255,255,0.07);
  color: white;
}
.chapter-item.active {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
}

.chapter-num {
  min-width: 28px; height: 28px;
  background: rgba(255,255,255,0.08);
  border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700;
  flex-shrink: 0;
  transition: all 0.15s;
}
.chapter-item.active .chapter-num {
  background: #4ade80;
  color: #071f0b;
}

.chapter-title { font-size: 0.84rem; line-height: 1.3; }

.sidebar-footer {
  padding: 14px 16px;
  border-top: 1px solid rgba(255,255,255,0.06);
  font-size: 0.75rem;
  color: rgba(255,255,255,0.25);
  text-align: center;
}

/* ── Main ──────────────────────────────────────────── */
.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 100svh;
  background: linear-gradient(160deg, #0f3d1a 0%, #1a6b2a 55%, #0f3d1a 100%);
  color: white;
}

header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 16px 10px;
}

/* Hamburger animato */
.menu-btn {
  background: rgba(255,255,255,0.12);
  border: none; color: white;
  width: 42px; height: 42px;
  border-radius: 12px;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}
.menu-btn:hover { background: rgba(255,255,255,0.2); }

.hamburger {
  display: flex; flex-direction: column;
  gap: 5px; width: 20px;
}
.hamburger span {
  display: block; height: 2px;
  background: white; border-radius: 2px;
  transition: all 0.3s;
}

.logo-area { display: flex; align-items: center; gap: 10px; }
.logo { font-size: 1.8rem; }
h1 { font-size: 1.15rem; font-weight: 700; line-height: 1.1; }
h1 span { font-weight: 300; opacity: 0.8; }
.tagline { font-size: 0.7rem; opacity: 0.45; margin-top: 2px; }

/* ── Response Box ──────────────────────────────────── */
main {
  flex: 1;
  padding: 8px 14px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.response-box {
  flex: 1;
  background: rgba(255,255,255,0.08);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 20px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  min-height: 200px;
}
.response-box.empty {
  align-items: center;
  justify-content: center;
}

/* Loading */
.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin: auto;
}
.loading-text {
  font-size: 0.82rem;
  opacity: 0.6;
  font-style: italic;
}
.dot {
  display: inline-block;
  width: 9px; height: 9px;
  background: rgba(255,255,255,0.7);
  border-radius: 50%;
  animation: bounce 1.2s infinite ease-in-out;
}
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.5); opacity: 0.4; }
  40%           { transform: scale(1);   opacity: 1; }
}

/* Answer */
.answer { display: flex; flex-direction: column; gap: 12px; width: 100%; }

.answer-meta {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.1);
}
.answer-icon { font-size: 1rem; flex-shrink: 0; margin-top: 1px; }
.answer-query {
  font-size: 0.82rem;
  opacity: 0.65;
  font-style: italic;
  line-height: 1.4;
}

/* ── Markdown Content ──────────────────────────────── */
.markdown-content {
  flex: 1;
  font-size: 0.93rem;
  line-height: 1.75;
  overflow-y: auto;
  max-height: 45vh;
  padding-right: 4px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255,255,255,0.15) transparent;
}

/* Titoli */
.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3) {
  color: #a8e6b3;
  font-size: 0.95rem;
  font-weight: 700;
  margin: 14px 0 6px;
  padding-bottom: 4px;
  border-bottom: 1px solid rgba(168, 230, 179, 0.2);
}
.markdown-content :deep(h1) { font-size: 1rem; }

/* Bold e italic */
.markdown-content :deep(strong) { color: #c8f5d3; font-weight: 700; }
.markdown-content :deep(em)     { color: rgba(255,255,255,0.8); font-style: italic; }

/* Liste */
.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  padding-left: 20px;
  margin: 8px 0;
}
.markdown-content :deep(li) {
  margin: 5px 0;
  line-height: 1.6;
}
.markdown-content :deep(ul li::marker) { color: #4ade80; }
.markdown-content :deep(ol li::marker) { color: #4ade80; font-weight: 600; }

/* Paragrafi */
.markdown-content :deep(p) { margin: 8px 0; }
.markdown-content :deep(p:first-child) { margin-top: 0; }

/* Blockquote (note regolamento) */
.markdown-content :deep(blockquote) {
  border-left: 3px solid #4ade80;
  margin: 10px 0;
  padding: 8px 12px;
  background: rgba(74, 222, 128, 0.08);
  border-radius: 0 8px 8px 0;
  font-size: 0.88rem;
  opacity: 0.9;
}

/* Codice inline */
.markdown-content :deep(code) {
  background: rgba(255,255,255,0.12);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.85em;
}

/* HR separatore */
.markdown-content :deep(hr) {
  border: none;
  border-top: 1px solid rgba(255,255,255,0.1);
  margin: 12px 0;
}

/* Azioni */
.answer-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 12px;
  border-top: 1px solid rgba(255,255,255,0.08);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255,255,255,0.12);
  border: 1px solid rgba(255,255,255,0.2);
  color: white;
  padding: 8px 14px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.82rem;
  transition: all 0.2s;
}
.action-btn:hover { background: rgba(255,255,255,0.2); }
.action-btn.active { background: rgba(255, 100, 100, 0.3); border-color: rgba(255,100,100,0.5); }
.action-btn.secondary { opacity: 0.7; }
.action-btn.secondary:hover { opacity: 1; }

/* Placeholder */
.placeholder {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.placeholder-icon { font-size: 2.5rem; opacity: 0.7; }
.placeholder-main { font-size: 1rem; font-weight: 500; opacity: 0.85; }

.examples {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
.example-chip {
  background: rgba(255,255,255,0.07);
  border: 1px solid rgba(255,255,255,0.12);
  color: rgba(255,255,255,0.75);
  padding: 10px 14px;
  border-radius: 12px;
  cursor: pointer;
  font-size: 0.83rem;
  text-align: left;
  transition: all 0.2s;
  line-height: 1.3;
}
.example-chip:hover {
  background: rgba(255,255,255,0.13);
  color: white;
  border-color: rgba(74, 222, 128, 0.4);
}

.placeholder-hint {
  font-size: 0.75rem;
  opacity: 0.4;
}

/* Error */
.error-msg {
  background: rgba(220,50,50,0.15);
  border: 1px solid rgba(220,50,50,0.35);
  border-radius: 12px;
  padding: 10px 14px;
  font-size: 0.88rem;
}

/* ── Transcript ─────────────────────────────────────── */
.transcript-box {
  margin: 0 14px 6px;
  background: rgba(255,255,255,0.07);
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 14px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
}
.mic-pulse  { animation: pulse 1s infinite; }
.transcript-text { font-size: 0.9rem; opacity: 0.85; }
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.3; }
}

/* ── Microfono ──────────────────────────────────────── */
footer {
  padding: 10px 14px 36px;
  display: flex;
  justify-content: center;
}

.mic-btn {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: white;
  color: #1a6b2a;
  border: none;
  border-radius: 28px;
  padding: 15px 44px;
  cursor: pointer;
  font-weight: 700;
  box-shadow: 0 8px 28px rgba(0,0,0,0.35);
  transition: all 0.25s ease;
  width: 100%;
  max-width: 280px;
  overflow: hidden;
}
.mic-btn:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 10px 32px rgba(0,0,0,0.4); }
.mic-btn:disabled { opacity: 0.55; cursor: not-allowed; }

.mic-btn.listening {
  background: #ef4444;
  color: white;
}
.mic-btn.loading {
  background: rgba(255,255,255,0.7);
}

/* Anelli pulsanti quando in ascolto */
.mic-rings {
  position: absolute;
  inset: 0;
  pointer-events: none;
}
.mic-rings span {
  position: absolute;
  inset: 0;
  border-radius: 28px;
  border: 2px solid rgba(239, 68, 68, 0.5);
  animation: ring 1.8s infinite ease-out;
}
.mic-rings span:nth-child(2) { animation-delay: 0.9s; }
@keyframes ring {
  0%   { transform: scale(1);    opacity: 0.8; }
  100% { transform: scale(1.12); opacity: 0; }
}

.mic-icon  { font-size: 1.5rem; position: relative; z-index: 1; }
.mic-label { font-size: 0.78rem; font-weight: 500; position: relative; z-index: 1; }
</style>
