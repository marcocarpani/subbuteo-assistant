import { ref, onUnmounted } from 'vue'

export function useSpeechRecognition() {
  const transcript = ref('')
  const isListening = ref(false)
  const error = ref(null)
  const isSupported = ref('webkitSpeechRecognition' in window || 'SpeechRecognition' in window)

  let recognition = null

  if (isSupported.value) {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition
    recognition = new SpeechRecognition()

    recognition.lang = 'it-IT'         // Italiano!
    recognition.continuous = false
    recognition.interimResults = false
    recognition.maxAlternatives = 1

    recognition.onresult = (event) => {
      transcript.value = event.results[0][0].transcript
      isListening.value = false
    }

    recognition.onerror = (event) => {
      error.value = event.error === 'no-speech'
        ? 'Nessun audio rilevato. Riprova.'
        : `Errore: ${event.error}`
      isListening.value = false
    }

    recognition.onend = () => {
      isListening.value = false
    }
  }

  function startListening() {
    if (!isSupported.value) {
      error.value = 'Il tuo browser non supporta il riconoscimento vocale.'
      return
    }
    error.value = null
    transcript.value = ''
    isListening.value = true
    recognition.start()
  }

  function stopListening() {
    if (recognition && isListening.value) {
      recognition.stop()
    }
  }

  onUnmounted(() => {
    stopListening()
  })

  return {
    transcript,
    isListening,
    error,
    isSupported,
    startListening,
    stopListening
  }
}
