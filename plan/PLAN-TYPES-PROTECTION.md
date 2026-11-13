# Plan — Hub des types + filtre global + protection anti-copie

Date : 2026-11-13. Statut : en cours d'exécution.

## Chantier 1 — Hub des types + filtre global (archive vérifiée, SHA256 OK)

Archive : `plan/marexstore-types-filtre-2026-11-13.tar.gz`
(3923a4c92bb212e84e871e619d67656fa88bdb971e5b22f29da444bd2f0d529b)

### 1.1 Copie des fichiers (depuis /tmp/opencode/types/)
- internal/render/render3.go : homeFilterHTML, TypeHubHTML, CategoryPage (hub + liste masquée), boutique data-sub = slug sous-catégorie
- internal/render/render.go : brandOptions extrait
- internal/render/render2.go : versions v329 (app.js + style.css)
- static/app.js : présélections ?type=&marque=&prixmin=&prixmax=&dispo= + type dépendant de la catégorie (accueil)
- static/style.css : styles .home-filter
- templates/static/_foot.html : app.js v329
- internal/render/parity_test.go : normPage retire les nouveautés
- internal/render/typehub_test.go : nouveau (hub, filtre, data-sub)

### 1.2 Qualité
gofmt + go vet + go test ./... → attendu 9/9.

### 1.3 Vérif staging (base 579)
- Réseaux : 5 tuiles (139/71/8/2/1) ; ?type=cisco-reseau → 71, pastille active
- boutique ?cat=reseaux&type=cisco-reseau&marque=cisco&dispo=en-stock → 69
- boutique ?cat=ordinateurs&marque=hp&prixmin=100000&prixmax=1000000 → 6
- Accueil « Trouver un produit » : 38 types → 6 après choix Réseaux
- Parité externe 199/199

### 1.4 Livraison
Commit + push main, rebuild, redémarrage, vérif HTTPS (4 points du rapport).

Rollback : git revert + rebuild. Risque faible (pas de migration, base inchangée).

## Chantier 2 — Protection anti-copie (3 volets)

Rappel honnête : copie à 100 % impossible (le navigateur télécharge tout).
Objectif = copie chère, lente, inexploitable, punissable.

### Volet A — nginx + robots + canonical (faible risque)
- robots.txt : Disallow /admin /api/ + Crawl-delay (route Go dans cmd/store/main.go)
- nginx : deny UA scrapers (HTTrack, wget, curl, python-requests, Scrapy, MJ12bot, AhrefsBot, SemrushBot, DotBot, Bytespider…) → 444 ; exception bons bots (Googlebot, Bingbot, DuckDuckBot)
- limit_req séparées : pages 20r/s burst 40, /api/ 5r/s burst 10
- canonical sur chaque page (render2.go Head) + X-Robots-Tag noindex sur /admin et /api/*
- Vérifs : curl -A HTTrack → 444 ; curl -A Googlebot → 200 ; crawl humain OK (boutique, fiche, recherche)

### Volet B — Watermark photos (risque moyen, backup obligatoire)
- Inventaire formats (jpg/webp/png) ; Go stdlib ne ré-encode pas webp → JPG d'abord, WebP ensuite si propre (cwebp ou conversion, sinon exclusion assumée)
- Backup intégral /tmp + watermark bandeau « store.neva-ci.pro » + vérif visuelle échantillon AVANT masse
- Recopie volume ./images (pas de rebuild/import, URLs inchangées)
- Vérifs : 573 fichiers, poids, affichage desktop+mobile, revert testé

### Volet C — Juridique + monitoring (risque nul)
- CGV + mentions légales : article propriété intellectuelle (lire pages existantes, ajout minimal)
- Monitoring : compteur 429/444 logs nginx + alerte (IP > N fiches/min) ; Google Alerts sur phrases uniques
- Preuves d'antériorité : captures datées + commits (déjà acquis)

### Interdit
Blocage clic droit / sélection / « chiffrement » HTML : contournable, nuit UX/SEO.

## Ordre et jalons
1. Chantier 1 (1.1→1.4) — commit+push+prod avant suite
2. Volet A — commit+push+prod avant suite
3. Volet B (échantillon puis masse) — commit+push+prod avant suite
4. Volet C — commit+push final

Décision Volet B (avec utilisateur) : JPG seuls vs conversion totale en JPG (réimport DB).
Recommandation : JPG d'abord, WebP ensuite si propre.
