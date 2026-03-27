#!/bin/bash
# Script di inizializzazione progetto Subbuteo Assistant
# Eseguire dalla cartella subbuteo-assistant/

set -e

echo "📁 Inizializzazione repository Git..."
git init
git add .
git commit -m "feat: initial scaffold - Go backend RAG + Vue 3 PWA"

echo ""
echo "🔗 Collega il remote GitHub:"
echo "   git remote add origin https://github.com/marcocarpani/subbuteo-assistant.git"
echo ""
echo "📤 Poi fai push:"
echo "   git branch -M main"
echo "   git push -u origin main"
echo ""
echo "⚙️  Prima del push ricorda di:"
echo "   1. Copiare il PDF: cp /path/to/Regolamento-OldSubbuteo-2013.pdf backend/regolamento.pdf"
echo "   2. Aggiungere il PDF al .gitignore se non vuoi commitarlo (o includilo per Railway)"
echo ""
echo "🔑 Su GitHub → Settings → Secrets → Actions, aggiungi:"
echo "   VITE_API_URL = https://tuo-progetto.railway.app"
echo ""
echo "🚀 Su Railway:"
echo "   1. New Project → Deploy from GitHub repo → marcocarpani/subbuteo-assistant"
echo "   2. Variables → GEMINI_API_KEY = la_tua_chiave"
echo ""
echo "✅ GitHub Pages si attiva automaticamente al primo push!"
