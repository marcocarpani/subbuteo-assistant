import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')

// Registrazione esplicita Service Worker
// vite-plugin-pwa genera /sw.js — lo registriamo manualmente
// per garantire compatibilità con PWABuilder
if ('serviceWorker' in navigator) {
  window.addEventListener('load', async () => {
    try {
      const reg = await navigator.serviceWorker.register('/sw.js', {
        scope: '/'
      })
      console.log('[SW] Registered:', reg.scope)

      // Aggiorna automaticamente il SW quando disponibile
      reg.addEventListener('updatefound', () => {
        const newWorker = reg.installing
        newWorker?.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            // Nuovo SW disponibile — ricarica silenzioso
            newWorker.postMessage({ type: 'SKIP_WAITING' })
            window.location.reload()
          }
        })
      })
    } catch (err) {
      console.warn('[SW] Registration failed:', err)
    }
  })
}
