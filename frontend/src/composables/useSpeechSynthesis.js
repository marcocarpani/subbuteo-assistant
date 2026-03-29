import { ref } from 'vue'

export function useSpeechSynthesis() {
  const isSpeaking  = ref(false)
  const isSupported = ref('speechSynthesis' in window)

  // Precarica le voci (alcune piattaforme le caricano in async)
  let voices = []
  function loadVoices() {
    voices = window.speechSynthesis.getVoices()
  }
  if (isSupported.value) {
    loadVoices()
    window.speechSynthesis.onvoiceschanged = loadVoices
  }

  // Trova la migliore voce italiana disponibile
  function getItalianVoice() {
    if (!voices.length) loadVoices()
    // Priorità: voce italiana nativa, poi qualsiasi it-IT, poi it
    return (
      voices.find(v => v.lang === 'it-IT' && !v.localService === false) ||
      voices.find(v => v.lang === 'it-IT') ||
      voices.find(v => v.lang.startsWith('it')) ||
      null
    )
  }

  /**
   * Legge il testo ad alta voce.
   * @param {string} text  - Testo plain (già stripped da markdown)
   * @param {number} rate  - Velocità: 1.0 normale, 1.5 veloce, 2.0 molto veloce
   */
  function speak(text, rate = 1.0) {
    if (!isSupported.value || !text) return

    // Ferma eventuale parlato precedente
    window.speechSynthesis.cancel()

    // Chunk il testo in frasi per evitare il bug di Chrome
    // che tronca utterance lunghe dopo ~200 parole
    const sentences = splitIntoChunks(text, 200)
    let index = 0

    function speakNext() {
      if (index >= sentences.length) {
        isSpeaking.value = false
        return
      }

      const utterance = new SpeechSynthesisUtterance(sentences[index])
      utterance.lang   = 'it-IT'
      utterance.rate   = Math.max(0.5, Math.min(2.0, rate)) // clamp 0.5–2.0
      utterance.pitch  = 1.0
      utterance.volume = 1.0

      const italianVoice = getItalianVoice()
      if (italianVoice) utterance.voice = italianVoice

      utterance.onstart = () => { isSpeaking.value = true }
      utterance.onend   = () => { index++; speakNext() }
      utterance.onerror = (e) => {
        // Ignora errori "interrupted" (normali quando si ferma manualmente)
        if (e.error !== 'interrupted') {
          console.warn('TTS error:', e.error)
        }
        isSpeaking.value = false
      }

      window.speechSynthesis.speak(utterance)
    }

    speakNext()
  }

  function stop() {
    window.speechSynthesis.cancel()
    isSpeaking.value = false
  }

  return { speak, stop, isSpeaking, isSupported }
}

// Divide il testo in chunk da N parole rispettando la punteggiatura
function splitIntoChunks(text, maxWords) {
  // Prima dividi per frasi naturali (punto, punto esclamativo, punto interrogativo)
  const sentences = text.match(/[^.!?]+[.!?]*/g) || [text]
  const chunks = []
  let current = ''
  let wordCount = 0

  for (const sentence of sentences) {
    const words = sentence.trim().split(/\s+/).length
    if (wordCount + words > maxWords && current) {
      chunks.push(current.trim())
      current = sentence
      wordCount = words
    } else {
      current += ' ' + sentence
      wordCount += words
    }
  }
  if (current.trim()) chunks.push(current.trim())
  return chunks.filter(Boolean)
}
