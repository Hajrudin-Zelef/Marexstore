# Plan complet — Marexstore (store.neva-ci.pro)

Date : 2026-11-12. Source : `catalogue-mediatec-v3.25.tar.gz` (12M, 483 fichiers).

## 0. État actuel (constaté en lecture seule)

- Contenu : `index.html`, `boutique.html` (hub 8 catégories), 7 pages `categories/`,
  171 pages `produits/`, `assets/style.css` (~57Ko) + `app.js` (856 lignes) + `facets.js`,
  ~200 images (majorité `jpg/png`, minorité `webp/avif`), 7 pages groupe, pages devis,
  contact, suivi-commande, compte, parametres, CGV, livraison, confidentialité.
- Source de vérité : `sources/produits-v3.json` (~539Ko, 171 fiches) + `build_site_v3.py`
  + `enrich/*.json`. Tests `test-v3.js` : 286 tests OK, 0 lien mort sur 198 pages.
- Aujourd'hui 100% statique : panier `localStorage[mediatec-panier-v1]`, commande via
  `https://wa.me/2250705087507`, compte en localStorage, recherche JS côté client,
  1 breakpoint à 860px, pas de `srcset`, Google Fonts bloquant.
- Infra : VM Debian 13, Go 1.27.1, Docker 29, nginx host (page défaut sur :80),
  `store.neva-ci.pro` → `159.195.40.112` (HTTP 200 nginx, HTTPS timeout à configurer),
  Postgres Supabase déjà sur `127.0.0.1:5433`.
- Repo `Hajrudin-Zelef/Marexstore` : vide au contrôle du 12/11. Prêt pour premier push.
- Sécurité : token `ghp_` posté en clair = considéré compromis, à révoquer, non utilisé.

Décisions validées : Go + chi + html/template + htmx + SQLite puis Postgres,
compte optionnel, commande WhatsApp/devis + Mobile Money (Wave/Orange/MTN) +
livraison/virement, design réutilisé à 100%, ultra-responsive/rapide/sécurisé sans friction.

## 1. Stack retenue et pourquoi

- `Go + net/http (+ chi)` : langage déjà maîtrisé (NevaLink, Cetas Gateway),
  un seul binaire, même déploiement Docker sur la VM.
- `html/template` : réutilise le HTML/CSS actuel tel quel, découpé en layouts/partials.
  Pas de réécriture React (Next.js écarté pour cette raison).
- `htmx` : recherche, filtres par facettes, tri, pagination, tiroir panier en fragments
  HTML, sans SPA.
- `SQLite (WAL) puis Postgres` : démarrage rapide, abstraction dès le jour 1 pour
  basculer sur le Postgres existant quand le volume l'exige.
- `sessions DB + bcrypt(12)` : comptes comme NevaLink, cookies sécurisés.
- `/admin` maison : CRUD produits/commandes, import JSON, statuts.
- Déploiement : binaire `:8080` en Docker + nginx reverse-proxy + TLS.

## 2. Architecture cible

```
marexstore/
  cmd/store/main.go
  internal/
    config/      // PORT, DB_PATH, SESSION_SECRET, BASE_URL via env
    db/          // open SQLite WAL, migrations, requêtes préparées
    models/      // Product, User, Order, OrderItem, facettes
    handlers/    // public, search, cart, checkout, account, admin
    auth/        // sessions DB + bcrypt + CSRF
    middleware/  // security headers, rate-limit, Request-ID, logs
  templates/
    layout.html
    partials/header.html, footer.html, card.html, panier.html, filtres.html
    pages/index.html, boutique.html, categorie.html, produit.html,
          devis.html, suivi.html, compte.html, admin.html
  static/        // assets actuels minifiés + images/ + thumbs/
  db/migrations/001_init.sql
  import/        // produits-v3.json -> DB + search_index + FTS5
  Dockerfile     // multi-stage, distroless, non-root, FS read-only
  docker-compose.yml
  nginx/store.neva-ci.pro.conf
```

### 2.1 Schéma DB V1 (migrable Postgres)

- `products(id INTEGER PK, slug TEXT UNIQUE, nom, marque, prix INTEGER, devise,
  dispo, categorie, sous_categorie, image, accroche, description, specs JSON,
  search_index TEXT)`
- FTS5 sur `search_index` (nom+marque+catégories+accroche+description normalisés).
- `users(id, email UNIQUE, hash_bcrypt, nom, tel, created_at)`
- `sessions(token PK, user_id, expires_at, ip, ua)`
- `orders(id, user_id NULL, statut, moyen_paiement, total, client JSON, preuve TEXT, created_at)`
  - `moyen_paiement IN (whatsapp, wave, orange, mtn, livraison, virement)`
  - `statut IN (devis, whatsapp_recu, attente_momo, paye, prepare, livre, annule)`
- `order_items(order_id, product_id, qty, prix_unitaire)`
- Index : `products(slug)`, `products(categorie)`, `orders(created_at)`, `sessions(expires_at)`.

### 2.2 Routes

Public :
- `GET /`, `/boutique`, `/categories/{slug}`, `/produits/{slug}`
- `GET /recherche?q=`, `/devis`, `/suivi-commande`, `/compte`, `/static/*`

Fragments htmx :
- `GET /partials/search`, `GET /partials/filtres`
- `POST /panier/ajouter|retirer|vider`, `GET /panier/count`
- `POST /commande` (choix moyen paiement)

Auth (optionnel) :
- `POST /auth/inscription|connexion|logout`

Admin :
- `GET|POST /admin/produits`, `GET|POST /admin/commandes/{id}/statut`
- `POST /admin/import` (JSON produits-v3)

Futur V2 :
- `POST /pay/wave/callback`, `POST /pay/orange/callback` (vérif signature HMAC + idempotence).

## 3. Responsive (garder le design à 100%)

- Conteneur fluide `min(1200px,100%-32px)` au lieu de 1200px fixe.
- Grille cartes `auto-fill/minmax(160px,1fr)`, sidebar → accordéon sous 860px.
- Header : search pleine largeur ligne 2 sur mobile, `nav.main` → hamburger + méga-menu accordéon.
- Tiroir panier `100vw` mobile / `380px` desktop, filtres `sticky` mobile en tiroir.
- Cibles tactiles >=44px, `inputmode=tel/email`, erreurs inline en français.
- Images : `srcset/sizes + loading=lazy + width/height` anti-CLS, hero `fetchpriority=high`.
- `prefers-reduced-motion` conservé, focus visibles clavier.
- Fonctionne sans JS (formulaires GET/POST classiques), htmx = amélioration progressive.

## 4. Performance (budget LCP <2.5s sur 4G)

- Go : templates cachés en prod, pagination 24 produits, fragments htmx au lieu de
  `boutique.html` 453Ko, requêtes préparées + `LIMIT`, FTS5.
- HTTP : gzip/br via nginx, `Cache-Control: public, immutable, max-age=31536000`
  sur `/static/*?v=`, `ETag` + `If-None-Match`.
- Assets : minify `style.css/app.js` au `docker build`, `defer` JS, `preconnect`
  fonts + `display=swap` + fallback système, critical CSS header/hero inline.
- Images : générer `thumbs/400px` + `full/`, convertir vers `webp/avif` au build
  en gardant `jpg` en fallback, `expires` nginx long.
- Budget : HTML <80Ko, CSS critique <40Ko, JS <30Ko hors htmx local.

## 5. Sécurité robuste sans bloquer l'UX

Principe : secure by default, friction uniquement en cas d'abus.

- Headers (nginx + middleware Go) :
  `CSP default-src 'self'; img-src 'self' data:; script-src 'self'; style-src 'self'
  fonts.googleapis.com; connect-src 'self'` d'abord `Report-Only` 1 semaine puis enforce,
  `HSTS`, `X-Frame-Options DENY`, `X-Content-Type-Options nosniff`,
  `Referrer-Policy strict-origin-when-cross-origin`,
  `Permissions-Policy camera=(), microphone=()`.
- Sessions : `bcrypt(12)`, cookies `HttpOnly; Secure; SameSite=Lax; Path=/`,
  token 32 octets crypto/rand en DB, rotation à la connexion, expiry 30j si opt-in,
  message login générique.
- CSRF : token synchronizer pour `POST /panier /commande /admin`, renvoyé auto via
  meta htmx, refresh silencieux si expiré (pas de perte panier).
- Anti-abus invisible : rate-limit IP `login 10/min`, `commande 20/min`, `admin 30/min`
  → `429 + Retry-After` + toast htmx, Turnstile invisible seulement en cas de pic.
- XSS/SQLi : `html/template` auto-escape partout, `slug ^[a-z0-9-]+$`, prix/qty
  int bornés serveur, 100% SQL préparé, abandon confiance localStorage.
- Admin : session séparée + rôle, lockout 5 échecs → 15min, audit log, upload
  allowlist `jpg/png/webp/avif` + magic bytes + resize serveur + nom aléatoire,
  stocké hors chemin exécutable. Durcissement ultérieur : restriction IP nginx.
- Infra : image distroless, user non-root, FS read-only + tmpfs, `no-new-privileges`,
  limits pids/memory, secrets en env uniquement, TLS 1.2+, backup SQLite journalier,
  logs sans PII, `X-Request-ID`.
- Paiements V1 : on stocke `moyen + statut`, pas d'API tiền. V2 : callbacks
  Wave/Orange vérifiés HMAC + idempotence `order_id`, jamais confiance montant client.

## 6. Repo + déploiement

1. Hygiène (à faire côté utilisateur, token révoqué) :
   `gh auth login`, `git init`, `.gitignore` (binaire, `*.db`, `.env`, `/tmp`),
   branches `main` + `dev`, protection branche + secret scanning.
2. Commits :
   1. extraction site sans `sources/` sauf `produits-v3.json` (ne pas commiter `.tar.gz`)
   2. scaffolding Go + migrations + import 171 vérifié
   3. templates découpés depuis HTML actuel
   4. recherche/panier htmx
   5. commande multi-paiement + compte optionnel
   6. /admin durci
   7. Dockerfile + compose + vhost + TLS
3. Prod :
   `docker compose up -d` (app `:8080`), nginx `proxy_pass` + cache static +
   rate-limit, `certbot --nginx -d store.neva-ci.pro`, vérifs `curl -I` CSP/HSTS,
   `Lighthouse + zap/nuclei` d'abord en Report-Only.

## 7. Critères d'acceptation V1

- 171 produits importés, 0 lien mort, recherche `hp 650a`, `imprimante` OK.
- Panier serveur + WhatsApp + choix Wave/Orange/MTN/livraison/virement enregistrés.
- Compte optionnel fonctionnel, admin CRUD + statuts + import JSON.
- Mobile 360px utilisable, LCP <2.5s, scan sécu propre, backup OK.
