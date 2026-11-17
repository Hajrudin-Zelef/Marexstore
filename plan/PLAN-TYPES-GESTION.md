# Gestion des catégories (Types d'équipements) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: native execution in this session, task-by-task (TDD, red-green-refactor, commit per task). Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** La page `/admin/types` devient une vraie gestion des catégories : `+` par carte, création/renommage inline, menu `⋯` par ligne (renommer, ajouter, fusionner, voir produits, supprimer si vide), alertes avant destruction, rollback transactionnel, effet prod immédiat.

**Architecture:** Nouvelle table `declared_types` pour mémoriser les catégories déclarées (sinon une catégorie vide, 0 produit, reste invisible car les listes sont calculées depuis `products`). Actions POST sur `/admin/types` existant (`action=creer|renommer|fusionner|supprimer`), `reload()`+FTS+audit à chaque succès, redirects honnêtes `msg=`/`err=`.

**Tech Stack:** Go stdlib + chi, SQLite, `static/admin.js` + `static/admin.css` (bump version à chaque changement).

**Spec (décisions validées en brainstorming) :**
- `+` en haut à droite de chaque carte store → ligne « nouvelle catégorie » en bas de sa carte, texte présélectionné, Entrée crée / Échap annule.
- Nouvelles catégories affichées en bas de leur carte, compteur 0.
- Menu `⋯` au survol de chaque ligne : Renommer (inline), Ajouter (sœur juste en dessous, prête à nommer), Fusionner (choix cible), Voir les produits (lien liste filtrée), Supprimer (0 produit uniquement).
- `+ Nouveau store` en haut de page (niveau 1).
- Alertes : confirmation obligatoire avant fusion/suppression (même pattern `data-mx-confirm` que le reste de l'admin).
- Rollback : chaque action est atomique (transaction ou Exec unique) — en cas d'erreur rien n'est à moitié appliqué ; toutes les actions sont tracées dans `admin_audit` (qui/quoi/quand) pour rejouer manuellement.
- Refus honnêtes : nom vide, doublon, suppression d'un type non vide, source de fusion vide.

## Global Constraints
- Parité boutique v3.25 : aucun changement du rendu public sauf les données (types renommés/fusionnés).
- Plafond et règles existants inchangés ; pas de 3e niveau (store > type uniquement).
- `script-src 'self'` : aucun JS inline (tout dans `static/admin.js`).
- `go test ./...` vert + parité avant chaque commit.

## Review Focus
- Nom avec `|` ou vide → refusé avec `err=` explicite, rien en DB.
- Création d'un doublon (existant avec produits ou déclaré) → `err=`, pas de doublon.
- Renommage vers un nom déjà pris dans le même store → refusé (pas de fusion silencieuse).
- Suppression d'un type avec produits → refusée même si cochée (garde serveur, pas seulement le menu).
- Fusion dont la source se vide pendant l'opération → `err=Rien à fusionner`, store non rechargé pour rien.

---

### Task 1: Table `declared_types`

**Files:**
- Create: `db/migrations/005_types_gestion.sql`
- Modify: `internal/db/db.go` (EnsureSchema : appliquer 005)
- Test: `internal/db/types_gestion_test.go` (ou étendre test existant s'il y a lieu)

**Interfaces:**
- Consumes: rien (table indépendante).
- Produces: table `declared_types(top TEXT NOT NULL, sub TEXT NOT NULL, cree_le TEXT NOT NULL DEFAULT (datetime('now','localtime')), PRIMARY KEY(top, sub))` ; listing cartes = UNION types des produits + déclarés, déclarés vides en bas par `cree_le`.

- [ ] **Step 1: Write the failing test** — `TestEnsureSchemaCreeDeclaredTypes` : ouvre une DB vierge, `EnsureSchema`, vérifie `declared_types` existe et accepte un insert + doublon refusé (PRIMARY KEY).
- [ ] **Step 2: Run test to verify it fails** — Run: `go test ./internal/db/ -run TestEnsureSchemaCreeDeclaredTypes -v` — Expected: FAIL (table inexistante).
- [ ] **Step 3: Implement** — créer `db/migrations/005_types_gestion.sql` (CREATE TABLE IF NOT EXISTS) + appel `Migrate(.../005_types_gestion.sql)` dans `EnsureSchema` après 004.
- [ ] **Step 4: Run test to verify it passes** — Run: `go test ./internal/db/ -v` — Expected: PASS, suite verte.
- [ ] **Step 5: Commit** — `git add db/migrations/005_types_gestion.sql internal/db/db.go <test> && git commit -m "types: table declared_types (categories declarees, meme vides)"`.

### Task 2: Backend créer (`action=creer`)

**Files:**
- Modify: `internal/handlers/admin_extra.go` (AdminTypes POST)
- Test: `internal/handlers/types_gestion_test.go`

**Interfaces:**
- Consumes: `declared_types` (Task 1).
- Produces: `POST /admin/types` avec `action=creer&top=<top>&sub=<nom>` → insert déclaré (vide accepté), `reload()`, `auditAdmin(r, "types.creer", top+"|"+sub)`, redirect `msg=Categorie creee` ; erreurs → `err=` (vide/interdit/doublon).

- [ ] **Step 1: Write the failing test** — `TestAdminTypesCreeCategorie` : POST sans produits concernés → 303 `msg=`, ligne en `declared_types`, `app.Store` rechargé ; `TestAdminTypesCreeRefuseVideEtDoublon` : sub vide et doublon → 303 `err=`, rien en DB.
- [ ] **Step 2: Run test to verify it fails** — Run: `go test ./internal/handlers/ -run TestAdminTypesCree -v` — Expected: FAIL (action inconnue → err).
- [ ] **Step 3: Implement** — branche `action=creer` : trim, refuse vide et `|` dans le nom, refuse doublon (produits ou déclarés), insert, reload, audit, redirect honnête.
- [ ] **Step 4: Run test to verify it passes** — Run: `go test ./internal/handlers/ -run 'TestAdminTypes' -v` — Expected: PASS (anciens tests fusion inclus).
- [ ] **Step 5: Commit.**

### Task 3: Rendu carte + bouton `+` + ligne inline

**Files:**
- Modify: `internal/handlers/admin_extra.go` (AdminTypes GET : cartes + `+` + formulaire inline)
- Test: `internal/handlers/types_gestion_test.go` (même fichier)

**Interfaces:**
- Consumes: `declared_types` (Task 1).
- Produces: chaque carte a `data-types-top="<top>"` + `<button data-types-add>` ; chaque ligne `data-types-row` avec `data-top`/`data-sub` ; déclarés vides affichés en bas avec compteur 0 ; formulaire inline `action=creer` (champ `sub`, top en hidden server-safe).

- [ ] **Step 1: Write the failing test** — `TestAdminTypesRenduBoutonPlus` : GET avec un déclaré vide → 200, contient `data-types-add`, la ligne vide en bas avec `>0<`, pas de JS inline (`<script` sans src absent).
- [ ] **Step 2: Run test to verify it fails** — Expected: FAIL (`data-types-add` absent).
- [ ] **Step 3: Implement** — rendre `+`, lignes avec data attrs, formulaire inline minimal (progressif : marche sans JS via POST natif).
- [ ] **Step 4: Run test to verify it passes.**
- [ ] **Step 5: Commit.**

### Task 4: Backend renommer (`action=renommer`)

**Files:**
- Modify: `internal/handlers/admin_extra.go`
- Test: `internal/handlers/types_gestion_test.go`

**Interfaces:**
- Consumes: rien de nouveau.
- Produces: `action=renommer&top=&ancien=&nouveau=` → UPDATE `products` + `declared_types` en transaction, SyncFTS des slugs touchés, reload, audit, `msg=` ; cible déjà prise → `err=`.

- [ ] **Step 1: Write the failing test** — `TestAdminTypesRenomme` : 2 produits + 1 déclaré `x` → POST renommer x→z → DB et `Store.BySlug[].LiveSub=="z"`, `msg=` ; `TestAdminTypesRenommeRefuseCiblePrise` → `err=`, rien bougé.
- [ ] **Step 2: Run test to verify it fails** — Expected: FAIL.
- [ ] **Step 3: Implement** — transaction (BEGIN/COMMIT comme AdminCommande), refus cible prise/vide/interdite.
- [ ] **Step 4: Run test to verify it passes.**
- [ ] **Step 5: Commit.**

### Task 5: Menu `⋯` par ligne (HTML/CSS, sans logique serveur nouvelle)

**Files:**
- Modify: `internal/handlers/admin_extra.go` (menu dans chaque ligne), `static/admin.css`
- Test: `internal/handlers/types_gestion_test.go`

**Interfaces:**
- Consumes: lignes `data-types-row` (Task 3), futurs endpoints (Tasks 4/6/7 déjà ou à venir — le menu ne fait que pointer).
- Produces: chaque ligne a `<button data-types-menu>⋯</button>` + menu (Renommer / Ajouter / Fusionner / Voir les produits / Supprimer) ; « Voir » = lien `/admin/produits?cat=<top>&q=` ; Supprimer/Fusionner portent `data-mx-confirm`.

- [ ] **Step 1: Write the failing test** — `TestAdminTypesRenduMenu` : GET → contient `data-types-menu` + les 5 libellés + `data-mx-confirm` sur Supprimer.
- [ ] **Step 2: Run test to verify it fails** — Expected: FAIL.
- [ ] **Step 3: Implement** — HTML + CSS hover/focus du menu (pur CSS si possible, sinon micro-JS Task 8).
- [ ] **Step 4: Run test to verify it passes.**
- [ ] **Step 5: Commit.**

### Task 6: Fusion depuis la carte + suppression vide

**Files:**
- Modify: `internal/handlers/admin_extra.go`
- Test: `internal/handlers/types_gestion_test.go`

**Interfaces:**
- Consumes: `splitType` existant (fusion legacy).
- Produces: `action=fusionner&src=top|sub&dst=top|sub` (même moteur que le formulaire legacy, qui est conservé ou retiré — décision : le formulaire legacy du bas est **retiré**, remplacé par le menu) ; `action=supprimer&top=&sub=` → 0 produit exigé (vérif serveur), delete déclaré, `msg=` ; sinon `err=`.

- [ ] **Step 1: Write the failing test** — `TestAdminTypesFusionneDepuisCarte` (src/dst composite, DB + store) ; `TestAdminTypesSupprimeVideEtRefuseNonVide` (vide → `msg=` + ligne partie ; avec produits → `err=` + intact).
- [ ] **Step 2: Run test to verify it fails** — Expected: FAIL (actions inconnues).
- [ ] **Step 3: Implement** — réutiliser le moteur de fusion (split/rows-affected/FTS/reload/audit), ajouter supprimer avec garde COUNT, retirer l'ancien formulaire + son hidden top.
- [ ] **Step 4: Run test to verify it passes** — toute la suite handlers verte.
- [ ] **Step 5: Commit.**

### Task 7: `+ Nouveau store` + alertes/rollback/audit complets

**Files:**
- Modify: `internal/handlers/admin_extra.go`, `static/admin.css`
- Test: `internal/handlers/types_gestion_test.go`

**Interfaces:**
- Consumes: Tasks 2/4/6.
- Produces: bouton page `data-types-add-top` + `action=creer_top&top=` (store vide déclaré via `declared_types(top, sub='')` ? — décision : un store sans type se déclare avec `sub=''` exclu de l'affichage, juste pour lister la carte) ; `data-mx-confirm` sur fusion+supprimer ; audit sur les 4 actions ; aucune action ne laisse d'état partiel (transaction ou Exec unique).

- [ ] **Step 1: Write the failing test** — `TestAdminTypesCreeStore` (carte vide listée) ; `TestAdminTypesActionsAuditees` (creer/renommer/fusionner/supprimer → 4 lignes `admin_audit`) ; confirmations présentes dans le HTML (`data-mx-confirm`).
- [ ] **Step 2: Run test to verify it fails** — Expected: FAIL.
- [ ] **Step 3: Implement.**
- [ ] **Step 4: Run test to verify it passes.**
- [ ] **Step 5: Commit.**

### Task 8: JS (`admin.js`) + CSS finitions + vérification navigateur

**Files:**
- Modify: `static/admin.js` (bump `?v=` dans `admin_chrome.go`), `static/admin.css`
- Test: staging Playwright (pas de test Go : comportement navigateur)

**Interfaces:**
- Consumes: Tasks 3/5.
- Produces: `+` → ligne inline + focus + sélection du texte ; Entrée valide / Échap annule ; `⋯` toggle au clic (repli tactile, pas seulement hover) ; aperçu sans erreur JS ; version `admin.js` bumpée.

- [ ] **Step 1: Staging** — build binaire, DB copie, serveur port dédié.
- [ ] **Step 2: Scénario Playwright** — `+` crée ligne présélectionnée → Entrée → ligne en bas compteur 0 → renommer inline → `⋯` → fusionner → `msg=` + DB + fiche boutique ; suppression vide avec confirm ; tentative suppression non vide → refusé.
- [ ] **Step 3: Full suite** — `go clean -testcache && go test ./...` vert + parité.
- [ ] **Step 4: Commit** — JS/CSS/version.
- [ ] **Step 5: Rapport + attente GO déploiement** (commit/push/rebuild/live, comme d'habitude).
