#!/usr/bin/env bash
# Surveillance anti-scraping Marexstore (à lancer avec sudo sur la VM).
# Affiche : requêtes rejetées (444 = user-agent de pompage, 429 = rafale),
# top IP rejetées, et top IP par volume (candidats pompeurs).
# Usage : sudo ./tools/monitor-scraping.sh [/var/log/nginx/access.log]
set -u
LOG="${1:-/var/log/nginx/access.log}"
if [ ! -r "$LOG" ]; then echo "log illisible : $LOG (sudo requis)"; exit 1; fi
echo "== Rejets (444/429) : $(grep -c -E '" 444 |" 429 ' "$LOG" || true)"
echo "== Top IP rejetées (444/429) :"
grep -E '" 444 |" 429 ' "$LOG" | awk '{print $1}' | sort | uniq -c | sort -rn | head -n 10
echo "== Top IP par volume (toutes requêtes) :"
awk '{print $1}' "$LOG" | sort | uniq -c | sort -rn | head -n 10
echo "== Derniers rejets :"
grep -E '" 444 |" 429 ' "$LOG" | tail -n 5 | awk '{print $1, $4, $5, $6, $7, $9, $12}'
