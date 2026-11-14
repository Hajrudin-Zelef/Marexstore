# PLAN COMPLET RÉVISÉ — Marexstore : Studio front 100% + backend fiable

Date : 2026-11-14. Statut : EN ATTENTE DU OK (ne pas exécuter sans validation).
Révision : intègre l'avis de l'agent vérificateur (références contrôlées à 46859a7).
Feu vert demandé : **T0+T1 maintenant, T2 après fix P1, T3 en décision après T2** (pas de OK global).

## 0. Objectif et périmètre acté

Construire dans `/admin` un Studio universel utilisable par des juniors (No-code → Low-code → Pro-code) qui personnalise 100% du front (HTML, CSS, JS front, images, textes, menus libres, sections), sans jamais éditer le backend (Go, SQL métier, commandes, stock, auth).

* Backups : local V1 `/var/backups/marexstore`, abstraction `BACKUP_TARGET=local|s3|ftp|all`.
* Alertes : abstraction `ALERT_CHANNELS=whatsapp,mail,telegram`, best-effort par canal.
* Backend : fiabilité invisible + 1 seul écran `/admin/systeme` lecture seule.
* Choix acté avec le porteur : à la publication, **le serveur gagne** (le thème publié s'impose à tous, y compris aux visiteurs ayant un choix localStorage mémorisé).

## 1. État actuel vérifié (références corrigées)

* Front dynamique Go : `cmd/store/main.go:103-143` routes, `internal/handlers/pages.go:27-37 renderDoc`, `internal/render/render2.go:53-135 Header/HeaderTheme`, `render3.go:70-211 IndexBodyTheme`, `render2.go:394-431 HubHTML`, `222-259 CatalogWrap`, `443-522 FicheBody`, `276-311 Footer` (dont `276-295 WidgetsHTML` — slugs en dur `287-289`, **pas** `render3.go`), `315-390 Head/Doc/ThemeHead/CSS`.
* Thème : `internal/catalog/theme.go:12-21 Theme`, `DefaultSections:10`, `IsDefault:28-44`, `HtmlAttrs:47-58`, global `SiteTheme:62`, `site.ThemeFromDB database main.go:41`.
* Studio existant : `internal/handlers/admin_design.go:169-275` formulaire + iframe `/admin/design/apercu:259-275`, `themeFromForm:84-130`, `themeToSettings:142-159`, `absolutizePreview:53-69`, reset `173-184`.
* Garde-fous : `BasicAuth+bcrypt(12) admin.go:20-29`, `CSRF admin.go:33-39` sur tous POST, `RateLimit(40,80) main.go:70`, CSP `middleware/security.go:30-61` (`script-src 'self'`, aperçu `SAMEORIGIN:40-46`), `EnsureSchema db.go:58-110`, `EnsureFTS/SyncFTS db.go:24-44`, upload 5Mo `admin_extra.go:540-567`.
* Points durs confirmés : HTML concaténé en Go + parité v3.25, `prefix ""/..//../../ pages.go:40,44,53,60,73,86`, CSP anti-inline/CDN/`mailto:`, `SiteTheme` global + cache immutable `?v=`.

## 2. Correctifs imposés par la vérification (avant T2)

### P1 — Conflit localStorage > serveur (bloquant, à fixer en premier)
Constat vérifié : `static/theme.js:10-13` applique `t.color` localStorage sans condition ; `static/app.js:956-965 applyTheme` fait pareil (`if (t.color) setAttribute`). Un visiteur ayant déjà choisi une couleur ne verra donc jamais les publications du studio.
Fix retenu (serveur gagne) :
* `settings` : `theme_version` incrémenté à chaque `Publier` (`themeToSettings`).
* Rendu : `<meta name="mx-theme-global">` + `<html data-theme-version="N">` (`ThemeHead`, `HtmlAttrs` étendu).
* `theme.js` + `applyTheme` : si `localStorage.theme_version !== serveur`, on purge le choix couleur mémorisé et on applique le serveur ; le visiteur ne garde en local que `clair/sombre` (réinitialisé si le serveur impose un mode).
* Test exigé : `TestThemeServeurGagneApresPublish` + parité verte (défaut = aucune sortie).

### P2 — `/admin/fichiers` : pièce la plus risquée, sortie du feu vert
Un éditeur qui sert du HTML/CSS/JS aux visiteurs avec validation par denylist est fragile (contournement toujours possible). Donc : **T2 d'abord, décision T3 après 1-2 semaines de T2 prouvé en prod**. Si T3 un jour : allowlist stricte + parse réel + sandbox + `DISALLOW_FILE_EDIT`, jamais de denylist seule. Le CSS additionnel + interactif couvre déjà ~90% des besoins juniors.

### P3 — Menus : périmètre verrouillé
Le méga-menu catégories reste piloté par la taxonomie (`LiveCats`, `MegaGroups`) — **non éditable**. Les tables `menus/menu_items` ne servent qu'aux **liens libres** (topbar, footer, bandeau), profondeur ≤2, URL http(s)/interne validée. Jamais de remplacement du méga-menu.

### Détails corrigés
* Widgets : référence exacte `render2.go:276-295` (slugs `287-289`), pas `render3.go`.
* WhatsApp : un token seul ne suffit pas — `tools/notify.sh` utilisera l'API Meta Cloud (`WA_ALERT_TOKEN` + `WA_PHONE_ID`) avec repli log + lien `wa.me` manuel si absent. Mail via `msmtp/sendmail`, Telegram via Bot API, timeouts 10s.
* Aperçu : table `design_previews(token TEXT PK, payload JSON, expire_le)` — token `crypto/rand` 32o, TTL 30 min, usage session admin, jamais loggué en clair (remplace l'actuel `hasThemeQuery` sans token).

## 3. Architecture cible

### 3.1 Volet Interactif — étendre `/admin/design` (T2, après P1)
Même écran, nouvelles sections (tout en `settings`, upsert existant) :
* Identité : `site_title, tagline, logo_media_id, favicon_media_id` (120 car, png/svg/jpg).
* Couleurs/Polices/Fond : garder `theme_couleur/accent/mode/police` + `bg_color, bg_image, font_heading, font_body` (regex `^#[0-9a-f]{6}$`), + `theme_version` auto.
* Header/Footer libres : `topbar_liens JSON`, `footer_cols JSON`, `newsletter_texte`, `annonce_on/texte` existants.
* Accueil : `hero_slides JSON` (remplace `banners render3.go:19-23`), `widgets_slugs` (remplace `render2.go:287-289`), `home_sections(type,config,ordre,visible)`, `homepage_mode`.
* Boutique/Fiche (cosmétique seule) : `grille_defaut`, `fiche_onglets`, jamais prix/stock.
* Aperçu : iframe + toggle mobile/desktop + `postMessage` + `htmx`, `?preview_token=` brouillon jamais public, boutons Brouillon/Publier/Réinitialiser.

### 3.2 Volet Code — `/admin/fichiers` (T3, EN SUSPENS — décision après T2)
Table `003_design_files.sql` (via `EnsureSchema`, jamais destructif) : `design_files(nom,contenu,version,auteur,cree_le,actif)` + index `(nom,actif)`.
* Allowlist : `_top.html, _foot.html, _drawer.html, bodies/*.html, custom.css, custom.js, custom-head.html`. Interdit : `*.go`, `../`, >50Ko.
* Cycle : Brouillon → Valider (lint + parse, allowlist) → Aperçu token → Publier (transaction `actif=0` + `INSERT actif=1 version+1`) → Historique/Diff/Rollback (jamais `UPDATE/DELETE` contenu).
* Loader : `LoadStatic` lit `design_files(actif=1)` sinon disque sinon défaut Go. `?v=N`. Kill-switch `DISALLOW_FILE_EDIT=1`.

### 3.3 Backend fiabilité (T0+T1, feu vert demandé maintenant)
* `004_backend_ops.sql` : `ops_jobs(derniere_ok,detail)`, `admin_audit(admin,action,objet,ip,date)`, `design_previews(token,payload,expire_le)`.
* `tools/backup.sh` : `sqlite3 .backup` + `integrity_check`, rotation 7j+4 mensuelles `chmod 600`, cibles `local|s3|ftp|all`.
* `tools/restore.sh` : restore jetable + counts `products/products_fts/settings`.
* `tools/notify.sh` : whatsapp/mail/telegram best-effort. Déclencheurs : backup KO, integrity KO, `produits=0` au boot, disque >85%, `429` massifs, publish/reset/maintenance.
* `/admin/systeme` : version, counts, taille DB/WAL, dernier backup/import, FTS, maintenance, boutons `Sauvegarder` + `Vérifier intégrité`.
* Commandes : transitions strictes + transaction stock + totaux serveur + `Idempotency-Key` + `order_events` conservé.
* `/health/detailed` : ping + counts + version.

## 4. UX 3 niveaux (même studio)
1. **No-code** : palettes, selects, toggles, flèches d'ordre, upload logo.
2. **Low-code** : CSS libre filtré, blocs HTML avec `{{.Prefix}} {{.CSRF}}` documentés, JSON slides/menus schématisé + `?aide=` par champ.
3. **Pro-code (T3 seulement)** : fichiers + lint + diff + revisions + export zip.

## 5. Sécurité / parité / tests
* Garder `adminAuth/CSRF/Esc/CSP/XFO/HSTS`, `render.ThemeHead vide` si défaut, `absolutizePreview`, `withCache ?v=`.
* Nouveaux tests : `TestThemeServeurGagneApresPublish` (P1), `TestDesignFilesFallback`, `TestDesignValidationRefuse`, `TestDesignRollback`, `TestThemeHeadVideParDefaut`, `TestBackupRestore`, `TestNotifyBestEffort`.
* Staging : parité verte, 0 erreur console, captures desktop/mobile, reset → pixel-identique, preview token isolé et expiré.

## 6. Phasage demandé (feu vert partiel, pas global)
* **OK maintenant — T0 socle** : P1 fix thème d'abord, puis `003/004 SQL`, loader fallback, kill-switch, backup/notify/systeme, tests parité.
* **OK maintenant — T1 backend dur** : transactions commandes, idempotence, audit, `/health/detailed`, cron `0 2 * * * backup.sh`.
* **Après P1+T0+T1 — T2 interactif** : identité, header/footer libres, hero JSON, widgets, CSS additionnel, aperçu mobile/postMessage.
* **En décision après 1-2 semaines de T2 prouvé — T3 code** : seulement sur nouveau feu vert explicite.

## 7. Env (jamais commité)
```
BACKUP_TARGET=local
BACKUP_DIR=/var/backups/marexstore
AWS_S3_BUCKET= AWS_REGION= FTP_URL= FTP_USER= FTP_PASS=
ALERT_CHANNELS=whatsapp,mail
WA_ALERT_TOKEN= WA_PHONE_ID= ALERT_MAIL_TO=sales@mediatec-ci.com TG_BOT_TOKEN= TG_CHAT_ID=
DISALLOW_FILE_EDIT=0
```
