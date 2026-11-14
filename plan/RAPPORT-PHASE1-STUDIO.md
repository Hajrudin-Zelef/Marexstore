# Phase 1 — Intégration du Studio v16-2 dans /admin/design

**Date :** 2026-11-14 — **Base :** origin/main 0008bc1 (T0/T1 déployés) — **Auteur :** Leo

## Ce qui est livré

Le prototype v16-2 de Zelef (262 Ko, un seul fichier HTML) est découpé en
**MVC** et servi par Go sur la même URL **`/admin/design`** (page dédiée
plein écran, sans sidebar admin, bouton « ← Retour à l'admin » fonctionnel).
L'ancien formulaire « plat » est supprimé (récupérable via le commit 46859a7).

### Nouveaux fichiers — `web/admin/studio/` (268 Ko)

| Fichier | Contenu |
|---|---|
| `studio.html` (44 Ko) | Coquille : `<head>` + corps verbatim du prototype, marqueur `<!--STUDIO_INIT-->`, 4 `<script src="/admin/studio/js/…">` |
| `studio.css` (40 Ko) | Tout le CSS du prototype (37 Ko) |
| `js/model.js` (40 Ko) | État : front/header/footer/shop/fiche/panier/menu/seo/reglages/acc/Tx/MODULE_TEXTS/FILES/studioV16 + historique (snap/undo/redo) + `CATEGORIES` |
| `js/view.js` (52 Ko) | Rendu des panneaux, catégories, éditeurs, compteurs |
| `js/preview.js` (40 Ko) | Tous les `pv*` + en-tête/pied de page partagés + cadre navigateur/iframe |
| `js/controller.js` (52 Ko) | Événements, onglets, undo/redo, Publier serveur, Ctrl+K, export/import, init |

Ordre de chargement préservé : model → view → preview → controller.
Les globaux partagés ne sont pas renommés ni encapsulés (même comportement
que le prototype).

### Backend (Go)

- **`internal/handlers/admin_design.go`** — `AdminDesign` réécrit :
  - GET : lit `web/admin/studio/studio.html`, injecte le thème courant +
    le CSRF au marqueur `<!--STUDIO_INIT-->` (`window.__STUDIO_THEME__`,
    `window.__STUDIO_CSRF__`).
  - POST : publie le thème comme avant (themeToSettings + reload + audit),
    mais répond en **JSON `{"ok":true,"version":N}`** quand la requête a
    `Accept: application/json` (fetch du Studio), sinon redirection.
  - Reset inchangé (JSON aussi pour le fetch).
  - **Nouveau** `AdminStudioFile` : sert les fichiers découpés sous
    `/admin/studio/` (auth admin, chemin nettoyé, pas de `../`, MIME css/js/html).
  - **Nouveau** `studioThemeJSON()` : thème courant au format Studio
    (accent, font, showAnnonce, bandeau, css, couleur, **sections**).
  - `AdminDesignPreview` (`/admin/design/apercu`) conservé tel quel.
- **`cmd/store/main.go`** — route `r.Get("/admin/studio/*", app.AdminStudioFile)`.
- **`Dockerfile`** — `COPY web ./web` ajouté (l'image finale ne copiait que
  templates/static/db : sans cette ligne le Studio serait absent en prod).
- **`internal/handlers/admin_design_test.go`** — 3 nouveaux tests :
  `TestAdminDesignServeStudio` (page + thème + CSRF injectés, 401 sans auth),
  `TestAdminDesignPublieJSON` (POST JSON → version 1, thème persisté),
  `TestAdminStudioFile` (model.js servi, traversée `../` → 404),
  `TestAdminDesignPubliePreserveSections` (les sections sont préservées).

## Publication (phase 1 = thème seulement)

« Publier » envoie **le thème** au serveur (POST /admin/design + CSRF) :
couleur (gazonnier/liquid/perso selon l'accent), accent, mode auto, police,
bandeau (annonce_on + annonce_texte), CSS personnalisé, et les **sections
d'accueil du serveur renvoyées telles quelles**. Le reste du design
(header, footer, boutique, textes…) n'est PAS encore persisté côté serveur :
il le sera en phase 2 (API JSON `design.json` versionné). « Exporter »
(réglages) reste un export JSON local, séparé de Publier.

## Bugs trouvés pendant la découpe (tous corrigés + testés)

1. **`const $` non reconnu par le découpeur** : la regex `DECL_RE` n'acceptait
   que `\w` dans les noms (`$` exclu) → `$` parti en « code exécutable »
   dans controller.js, alors que view.js/preview.js l'utilisent → 3
   `ReferenceError` en cascade, navigation par catégories jamais construite.
   Corrigé : regex `[\w$]+` + `$` déplacé en tête de model.js.
2. **`let studioV16=front._studioV16` lu trop tôt** : la déclaration est
   partie dans model.js mais l'initialisation `if(!front._studioV16)…`
   restait en code exécutable (controller.js) → `studioV16` undefined →
   `applyV16Preview` plantait. Corrigé : l'initialisation par défaut est
   déplacée dans model.js juste avant l'alias.
3. **Sections d'accueil écrasées** : `themeFromFront()` envoyait
   `rows.map(r=>r.id)` (3 lignes de l'éditeur visuel) comme sections du
   thème → la publication aurait masqué 6 sections du site (filtre,
   onglets, vendus, banniere, recents, marques). Corrigé : le serveur
   injecte ses sections, le Studio les renvoie telles quelles.
4. **fetch bloqué par les identifiants dans l'URL** : `fetch('/admin/design')`
   échoue quand la page a été chargée avec `user:pass@` dans l'URL.
   Corrigé : `fetch(window.location.origin + '/admin/design')`.
5. **Cache navigateur** : `AdminStudioFile` envoie `max-age=3600` ; les
   fichiers découpés sont versionnés `?v=1` — penser à bumper à chaque
   modification JS/CSS du Studio.

## Vérification (test live, serveur local port 18099, DB 664 produits)

- `go build` + `go vet` + `go test ./...` : **verts** (4 tests design).
- **CDP navigateur (Chrome headless 1440 px) : 17/17 vérifications OK** :
  page Studio, thème + CSRF injectés, 12 modules, 4 catégories
  (Apparence/Pages/Contenu/Réglages), menu Apparence, onglet En-tête,
  fil « Studio / Apparence / En-tête », statut Publié à l'arrivée,
  Brouillon après modification, aperçu du bandeau mis à jour, Publié
  après Ctrl+Z, palette Ctrl+K, **Publier serveur → toast « Studio publié !
  Le site est à jour (thème v1) »**, thème v1 visible sur la boutique
  (`data-theme-version="1"`, `--accent:#2563eb`), sections 9/9 préservées.
- Non-régression : `/`, `/boutique.html?all=1`, `/admin`, `/admin/design`,
  `/admin/systeme` → 200. Le dialogue de contraste v13 fonctionne
  (un accent clair comme le lime bloque la publication tant qu'il n'est
  pas confirmé — comportement voulu du prototype).
- Captures 1440 px : `shot-studio-front.png`, `shot-studio-header.png`
  (dans `~/workspace/marexstore-test/`, non livrées dans le tar).

## Déploiement (agent sur place)

1. Extraire le tar à la racine du dépôt (`/home/sam/MarexStore`).
2. `chmod -R a+rX web` (les fichiers du tar arrivent en 640, illisibles
   par l'utilisateur non-root du conteneur — même piège que pour db/templates/static).
3. `go build ./... && go vet ./... && go test ./...`
4. `docker compose up -d --build --force-recreate`
5. Vérifier : `/admin/design` (auth admin) affiche le Studio avec les
   4 catégories ; « Publier » incrémente `theme_version` (visible sur la
   boutique : `data-theme-version`).

## Ce qui n'est PAS dans cette phase (phase 2)

- Persistance du design complet (header/footer/boutique/textes…) : API
  `GET/POST /admin/api/design` + `design.json` versionné dans `design_files`.
- Rendu de la boutique à partir du nouvel état (header/footer/shop…) :
  phase 3, à planifier séparément.
- `/admin/fichiers` (T3) : toujours en attente du feu vert explicite.
- Le Studio ne gère pas le mobile autrement qu'en aperçu.
