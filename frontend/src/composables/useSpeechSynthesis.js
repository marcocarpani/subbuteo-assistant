import { ref } from 'vue'

export function useSpeechSynthesis() {
  const isSpeaking = ref(false)
  const isSupported = ref('speechSynthesis' in window)

  let utterance = null

  function speak(text) {
    if (!isSupported.value || !text) return

    // Interrompi eventuale parlato precedente
    window.speechSynthesis.cancel()

    utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'it-IT'
    utterance.rate = 0.95    // Leggermente più lento per chiarezza
    utterance.pitch = 1.0
    utterance.volume = 1.0

    // Preferisci voce italiana se disponibile
    const voices = window.speechSynthesis.getVoices()
    const italianVoice = voices.find(v => v.lang.startsWith('it'))
    if (italianVoice) {
      utterance.voice = italianVoice
    }

    utterance.onstart = () => { isSpeaking.value = true }
    utterance.onend = () => { isSpeaking.value = false }
    utterance.onerror = () => { isSpeaking.value = false }

    window.speechSynthesis.speak(utterance)
  }

  function stop() {
    window.speechSynthesis.cancel()
    isSpeaking.value = false
  }

  return { speak, stop, isSpeaking, isSupported }
}
