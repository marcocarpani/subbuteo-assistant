import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  base: '/',
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'script',        // inietta script di registrazione nel HTML
      includeAssets: [
        'favicon.ico', 'favicon-32.png',
        'apple-touch-icon.png',
        'icon-*.png',
        'screenshot-mobile.png'
      ],
      manifest: {
        id:               '/',
        name:             'Subbuteo Assistant',
        short_name:       'Subbuteo',
        description:      'Assistente vocale per il regolamento Old Subbuteo Rev. 2.5',
        theme_color:      '#1a6b2a',
        background_color: '#0f3d1a',
        display:          'standalone',
        display_override: ['standalone', 'minimal-ui'],
        orientation:      'portrait',
        lang:             'it',
        start_url:        '/?source=pwa',
        scope:            '/',
        categories:       ['sports', 'reference', 'utilities'],
        prefer_related_applications: false,

        icons: [
          { src: '/icon-72.png',   sizes: '72x72',   type: 'image/png', purpose: 'any' },
          { src: '/icon-96.png',   sizes: '96x96',   type: 'image/png', purpose: 'any' },
          { src: '/icon-128.png',  sizes: '128x128', type: 'image/png', purpose: 'any' },
          { src: '/icon-144.png',  sizes: '144x144', type: 'image/png', purpose: 'any' },
          { src: '/icon-152.png',  sizes: '152x152', type: 'image/png', purpose: 'any' },
          { src: '/icon-192.png',  sizes: '192x192', type: 'image/png', purpose: 'any' },
          { src: '/icon-384.png',  sizes: '384x384', type: 'image/png', purpose: 'any' },
          { src: '/icon-512.png',  sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
        ],

        screenshots: [
          {
            src:         '/screenshot-mobile.png',
            sizes:       '390x844',
            type:        'image/png',
            form_factor: 'narrow',
            label:       'Subbuteo Assistant — Chiedi una regola vocalmente'
          }
        ],

        // Shortcuts — aumentano il punteggio App Capabilities
        shortcuts: [
          {
            name:       'Chiedi una regola',
            short_name: 'Chiedi',
            description:'Apri l\'assistente vocale',
            url:        '/?action=voice',
            icons: [{ src: '/icon-192.png', sizes: '192x192' }]
          },
          {
            name:       'Sfoglia regolamento',
            short_name: 'Regolamento',
            description:'Consulta i capitoli del regolamento',
            url:        '/?action=menu',
            icons: [{ src: '/icon-192.png', sizes: '192x192' }]
          }
        ]
      },

      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        // Strategia cache-first per asset statici
        runtimeCaching: [
          {
            // API backend — network first con fallback cache
            urlPattern: ({ url }) => url.pathname.startsWith('/api/'),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              networkTimeoutSeconds: 10,
              expiration: {
                maxEntries: 30,
                maxAgeSeconds: 60 * 60  // 1 ora
              },
              cacheableResponse: { statuses: [0, 200] }
            }
          },
          {
            // Font e asset esterni
            urlPattern: /^https:\/\/fonts\.(googleapis|gstatic)\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'font-cache',
              expiration: { maxEntries: 10, maxAgeSeconds: 60 * 60 * 24 * 365 }
            }
          }
        ],
        // Permetti al SW di prendere il controllo immediatamente
        skipWaiting: true,
        clientsClaim: true
      }
    })
  ],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
