<template>
  <div class="assistant">
    <!-- Header -->
    <header>
      <div class="logo">⚽</div>
      <h1>Subbuteo<br><span>Assistant</span></h1>
    </header>

    <!-- Area risposta -->
    <main>
      <div class="response-box" :class="{ empty: !answer && !isLoading }">
        <div v-if="isLoading" class="loading">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
        </div>
        <div v-else-if="answer" class="answer">
          <p class="query-echo">🎤 {{ lastQuery }}</p>
          <p class="answer-text">{{ answer }}</p>
          <button class="speak-btn" @click="toggleSpeak" :class="{ active: isSpeaking }">
            {{ isSpeaking ? '⏹ Stop audio' : '🔊 Leggi risposta' }}
          </button>
        </div>
        <div v-else class="placeholder">
          <p>Premi il microfono e chiedi una regola del Subbuteo</p>
          <p class="example">Es: "Quanti tocchi può fare una miniatura?"</p>
          <p class="example">Es: "Come si batte il calcio di rigore?"</p>
          <p class="example">Es: "Quando scatta il fuorigioco?"</p>
        </div>
      </div>

      <!-- Errori -->
      <div v-if="errorMsg" class="error-msg">
        ⚠️ {{ errorMsg }}
      </div>
    </main>

    <!-- Trascrizione live -->
    <div v-if="isListening || transcript" class="transcript-box">
      <span v-if="isListening" class="pulse">🎙️</span>
      <span>{{ transcript || 'In ascolto...' }}</span>
    </div>

    <!-- Pulsante microfono -->
    <footer>
      <button
        class="mic-btn"
        :class="{ listening: isListening, disabled: isLoading }"
        @click="handleMicClick"
        :disabled="isLoading"
      >
        <span class="mic-icon">{{ isListening ? '⏹' : '🎤' }}</span>
        <span class="mic-label">{{ micLabel }}</span>
      </button>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useSpeechRecognition } from '../composables/useSpeechRecognition.js'
import { useSpeechSynthesis } from '../composables/useSpeechSynthesis.js'

// Composables
const { transcript, isListening, error: speechError, startListening, stopListening } = useSpeechRecognition()
const { speak, stop, isSpeaking } = useSpeechSynthesis()

// Stato locale
const answer = ref('')
const lastQuery = ref('')
const isLoading = ref(false)
const errorMsg = ref('')

// Leggi la risposta
function toggleSpeak() {
  if (isSpeaking.value) {
    stop()
  } else {
    speak(answer.value)
  }
}

// Label pulsante microfono
const micLabel = computed(() => {
  if (isLoading.value) return 'Elaborazione...'
  if (isListening.value) return 'Tocca per fermare'
  return 'Tocca per parlare'
})

// Gestisci click microfono
function handleMicClick() {
  if (isListening.value) {
    stopListening()
    return
  }
  stop() // Ferma TTS se attivo
  errorMsg.value = ''
  startListening()
}

// Quando la trascrizione è pronta → chiama backend
watch(transcript, async (newTranscript) => {
  if (!newTranscript || isLoading.value) return

  lastQuery.value = newTranscript
  isLoading.value = true
  answer.value = ''
  errorMsg.value = ''

  try {
    const backendUrl = import.meta.env.VITE_API_URL || ''
    const res = await fetch(`${backendUrl}/api/ask`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query: newTranscript })
    })

    const data = await res.json()

    if (data.error) {
      errorMsg.value = data.error
    } else {
      answer.value = data.answer
      // Auto-speak la risposta
      speak(data.answer)
    }
  } catch (e) {
    errorMsg.value = 'Impossibile contattare il server. Controlla la connessione.'
  } finally {
    isLoading.value = false
  }
})

// Errori dal riconoscimento vocale
watch(speechError, (err) => {
  if (err) errorMsg.value = err
})
</script>

<style scoped>
.assistant {
  min-height: 100svh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(160deg, #0f3d1a 0%, #1a6b2a 60%, #0f3d1a 100%);
  color: white;
  font-family: 'Segoe UI', system-ui, sans-serif;
  padding: 0;
}

header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 24px 12px;
}

.logo {
  font-size: 2.5rem;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
}

h1 {
  font-size: 1.4rem;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.5px;
}

h1 span {
  font-weight: 300;
  opacity: 0.85;
}

main {
  flex: 1;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.response-box {
  background: rgba(255,255,255,0.1);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 24px;
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255,255,255,0.15);
}

.placeholder {
  text-align: center;
  opacity: 0.7;
}

.placeholder p {
  margin: 6px 0;
  font-size: 0.95rem;
}

.example {
  font-size: 0.8rem !important;
  opacity: 0.7;
  font-style: italic;
}

.answer {
  width: 100%;
}

.query-echo {
  font-size: 0.8rem;
  opacity: 0.7;
  margin-bottom: 12px;
  font-style: italic;
}

.answer-text {
  font-size: 1rem;
  line-height: 1.65;
  margin-bottom: 16px;
}

.speak-btn {
  background: rgba(255,255,255,0.15);
  border: 1px solid rgba(255,255,255,0.3);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: background 0.2s;
}

.speak-btn:hover, .speak-btn.active {
  background: rgba(255,255,255,0.25);
}

/* Loading dots */
.loading {
  display: flex;
  gap: 8px;
}

.dot {
  width: 10px;
  height: 10px;
  background: rgba(255,255,255,0.7);
  border-radius: 50%;
  animation: bounce 1.2s infinite ease-in-out;
}
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.5; }
  40% { transform: scale(1); opacity: 1; }
}

.error-msg {
  background: rgba(220, 50, 50, 0.2);
  border: 1px solid rgba(220, 50, 50, 0.4);
  border-radius: 12px;
  padding: 12px 16px;
  font-size: 0.9rem;
}

.transcript-box {
  margin: 0 20px 8px;
  background: rgba(255,255,255,0.08);
  border-radius: 12px;
  padding: 12px 16px;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
}

.pulse {
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

footer {
  padding: 16px 20px 36px;
  display: flex;
  justify-content: center;
}

.mic-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  background: white;
  color: #1a6b2a;
  border: none;
  border-radius: 32px;
  padding: 18px 48px;
  cursor: pointer;
  font-weight: 700;
  font-size: 1rem;
  box-shadow: 0 8px 24px rgba(0,0,0,0.3);
  transition: all 0.2s ease;
  width: 100%;
  max-width: 320px;
}

.mic-btn.listening {
  background: #ff4444;
  color: white;
  animation: glow 1.5s infinite alternate;
}

.mic-btn.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mic-icon {
  font-size: 1.8rem;
}

.mic-label {
  font-size: 0.85rem;
  font-weight: 500;
}

@keyframes glow {
  from { box-shadow: 0 8px 24px rgba(255, 68, 68, 0.4); }
  to { box-shadow: 0 8px 32px rgba(255, 68, 68, 0.8); }
}
</style>
