# Marexstore — P1 + T0 + T1 (Studio Design, socle + backend dur)

Date : 2026-11-14. Auteur : Leo. Plan : PLAN-STUDIO-100-2.md (v2 validée) + addendum final (A1–A6, T2 étendu).
Périmètre de cette livraison : **P1 (fix thème) + T0 (socle) + T1 (backend dur)** — la part de Leo.
T2 (interactif, A1–A6) et T3 (fichiers, sur nouveau feu vert) restent à l'agent.

## P1 — Le serveur gagne (fix du conflit localStorage > thème serveur)

Problème : `theme.js` et `app.js:applyTheme` appliquaient le choix couleur du
visiteur (localStorage) sans condition — le thème publié par l'admin n'atteignait
jamais les visiteurs qui avaient déjà choisi une couleur.

Fix :
- `settings.theme_version`, incrémenté à chaque Publier (une seule requête
  atomique `INSERT ... ON CONFLICT ... RETURNING`, dans la transaction de
  publication ; `catalog.Theme.Version` ; lu par `site.ThemeFromDB`).
- Rendu : `<html data-theme-version="N">` (via `Theme.HtmlAttrs()`), émis
  seulement quand le thème est non défaut (parité v3.25 intacte).
- `static/theme.js` (v2) et `static/app.js` (v339) : si la version serveur
  diffère de la version mémorisée, on purge la couleur (et le mode si le
  serveur en impose un), puis on applique le serveur. Le visiteur ne garde
  en local que clair/sombre, tant que le serveur n'impose pas de mode.
- Réinitialiser efface aussi `theme_version` (retour au défaut, parité).
- Versions JS bumpées partout (rendu dynamique + `_foot.html` statique +
  aperçu design) : `theme.js?v=2`, `app.js?v=339`.
- Tests : `TestThemeVersionAttr` (catalog) + `TestThemeServeurGagneApresPublish`
  (handlers, Publier 1 → 2, relecture DB, reset).

Vérifié en live : Publier gazonnier → `data-theme-version="1"`, Publier
liquid → `"2"`. JS vérifié par jsdom (5 cas : purge au changement de version,
conservation à version égale, repli liquid sans version, mode serveur imposé).

## T0 — Socle

- **Migrations 003/004** (`db/migrations/`, via `EnsureSchema`, jamais destructif,
  `IF NOT EXISTS`, rejouables) :
  - `003_design_files.sql` : `design_files(nom, contenu, version, auteur,
    cree_le, actif)` PK (nom, version) + index (nom, actif) — versions
    append-only, une seule active par nom (socle T3).
  - `004_backend_ops.sql` : `ops_jobs(nom, derniere_ok, detail)`,
    `admin_audit(...)`, `design_previews(token, payload, expire_le)`,
    `idempotency_keys(cle, order_id, cree_le)` (T1).
- **Loader** (`site.LoadStaticDB`) : `design_files(actif=1)` → disque →
  défaut Go (chaîne vide). Un fichier manquant ne bloque plus le démarrage
  (avertissement en log). Les pages présentes uniquement en base sont aussi
  énumérées (T3). `reload()` recharge aussi les gabarits (publication T3
  sans redémarrage). `LoadStatic` délègue à `LoadStaticDB(nil, ...)`.
- **Kill-switch** : `site.FileEditAllowed()` — `DISALLOW_FILE_EDIT=1`
  interdit l'éditeur de fichiers (T3). Affiché sur /admin/systeme.
- **tools/backup.sh** : `sqlite3 .backup` (cohérent même si le serveur tourne)
  + `integrity_check`, rotation 7 jours + 4 mensuelles (1re du mois),
  `chmod 600`, cibles `local|s3|ftp|all` (BACKUP_TARGET, s3/ftp best-effort),
  `ops_jobs` mis à jour, `tools/notify.sh` appelé s'il existe (fourni par
  l'agent — jamais d'échec si absent). Testé : backup + rotation + `ops_jobs`.
  Cron à installer par l'agent sur la VM : `0 2 * * * backup.sh`.
- **tools/restore.sh** : restauration JETABLE (copie temporaire) +
  `integrity_check` + comptes (products, products_fts, settings). N'écrase
  jamais la base live ; la promotion est manuelle (étapes affichées). Testé.
- **/admin/systeme** (lecture seule + 2 actions) : moteur Go, thème publié,
  comptes (produits/clients/commandes/lignes), taille DB/WAL, dernier backup,
  dernière vérification d'intégrité, dernier produit ajouté, FTS (664/664),
  maintenance, kill-switch édition. Boutons : « Sauvegarder (télécharger) »
  (VACUUM INTO → téléchargement, le conteneur est en lecture seule) et
  « Vérifier l'intégrité » (PRAGMA integrity_check). Sidebar : entrée
  « Système » (icône activity). Vérifié en live (GET 200, POST intégrité OK,
  téléchargement 3,5 Mo, integrity ok, 664 produits) + capture 1440 px.

## T1 — Backend dur

- **Commandes** :
  - `ApiCommande` : commande + lignes + événement de création + clé
    d'idempotence en UNE SEULE transaction (jamais de commande partielle).
    `Idempotency-Key` : rejouer la même requête renvoie la commande existante
    (vérifié en live : 2 POST même clé = 1 seule commande). Total calculé
    côté serveur depuis la base. Événement de création (`par='site'`).
  - `AdminCommande` : transition durcie — `Begin()` vérifié, lignes de stock
    collectées puis mises à jour (une erreur annule toute la transition),
    `Commit()` vérifié. Le stock n'est décrémenté qu'au passage à « paye »
    (une seule fois, machine à états stricte — transition interdite rejetée).
    Vérifié en live : stock 3 → 2, timeline complète, pas de double décrément.
  - Test : `TestApiCommandeIdempotente` (2 clés, 2 commandes, totaux serveur).
- **Audit** (`admin_audit`) : helper `auditAdmin(r, action, objet)` (admin
  Basic Auth + IP sans port, best-effort). Appelé sur : design publier /
  réinitialiser, commande statut, système sauvegarder / intégrité, réglages,
  produit enregistrer. Vérifié en live (2 lignes commande.statut).
- **/health/detailed** (admin, Basic Auth, exempté de la maintenance) :
  `{"ok":true, "produits":664, "clients":3, "commandes":6, "go":"go1.27.2",
  "theme_version":0}`. Sans auth : 401. `/health` public inchangé.

## Vérification (tout vert)

- `gofmt` propre, `go build ./...`, `go vet ./...`, `go test ./...` (catalog,
  handlers, render — parité v3.25 incluse).
- Serveur de test (base 664 produits) : P1 (2 publications, versions 1→2),
  /admin/systeme (GET + 2 POST + capture), commandes (idempotence live,
  transitions, stock, audit), /health/detailed (401 sans auth, JSON avec auth).
- `theme.js` v2 vérifié par jsdom (5 cas). `backup.sh`/`restore.sh` testés
  (backup, rotation 7j+4 mensuelles, restore, `ops_jobs`).

## Reste à l'agent (division convenue)

- `tools/notify.sh` (WhatsApp Meta Cloud + mail + Telegram) — déclencheurs
  déjà câblés (backup.sh l'appelle s'il existe ; jamais d'échec si absent).
- Cron `0 2 * * * backup.sh` + dossier `/var/backups/marexstore` sur la VM.
- Déploiement (extraire, `go build ./... && go vet ./...`,
  `docker compose up -d --build --force-recreate`) + vérification live
  (/admin, /admin/systeme, /admin/design, accueil avec thème publié).
- Puis T2 (interactif, A1–A6 de l'addendum final) — T3 seulement sur nouveau
  feu vert explicite.

## Points ouverts / décisions pour Zelef (non modifiés ici)

- **Prix de la commande** : `ApiCommande` totalise sur `prix` (stock) alors
  que l'affichage utilise `COALESCE(prix_boutique, prix)`. Si un produit a un
  prix boutique différent, le total de la commande ne correspond pas au prix
  affiché. Non modifié (question « prix stock ou prix boutique ? » en attente).
- **Version de l'application** sur /admin/systeme : affiche le moteur Go +
  la version du thème (pas de version de code injectée — `.git` exclu du
  contexte Docker, pas de ldflags). Une version git via ldflags est possible
  plus tard si souhaité.
- `notify.sh` : l'agent le fournit (WhatsApp + autres) ; en attendant, les
  échecs de backup sont journalisés dans `ops_jobs` et visibles sur
  /admin/systeme.
