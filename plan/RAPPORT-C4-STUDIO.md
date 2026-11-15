# RAPPORT C4 — Écriture des fichiers (T3) dans le Studio

**Date :** 2026-11-15 — **Base :** origin/main 847355a (C1+C2+C3 déployés)
**Feu vert :** Zelef, « Fait C4 » puis « Go » (plan présenté le 2026-11-15 10:06, GO 10:24)

## Ce que C4 fait

Le mode Code du Studio passe en **écriture** : on peut modifier un vrai fichier
du site et l'enregistrer depuis le Studio. Avant C4, il était en lecture seule.

## Architecture (le point important)

**Je n'écris pas sur disque, je stocke en base.** Le conteneur est
`read_only: true` et `./static` est monté `ro` — écrire sur disque échouerait,
et serait perdu au prochain rebuild de l'image. Chaque enregistrement va donc
dans `design_files`, **versionné** (nouvelle version à chaque save, comme le
design) : rollback en 1 clic via l'Historique.

- **Gabarits** (`templates/static/*.html`) : nom en base = nom de fichier nu
  (`_top.html`, `apropos.html`) — c'est exactement ce que `LoadStaticDB`
  interroge déjà (T0). Un gabarit enregistré prend effet au prochain
  `reload()`, sans redémarrage.
- **Fichiers static** (`app.js`, `style.css`, `theme.js`) : nom en base =
  chemin complet (`static/app.js`). Ils sont servis par un nouveau handler
  `StaticFile` qui lit la base d'abord, puis le disque. La version base est
  servie avec `Cache-Control: no-cache` (le contenu a changé sans que `?v=`
  ne bouge) ; le disque garde le cache long habituel.

## Fichiers

| Fichier | Changement |
|---|---|
| `internal/design/fichiers.go` (nouveau) | `FichierNom` (liste fermée, pas de traversée), `SaveFichier`, `LoadFichier`, `ListFichierVersions`, `ActivateFichier`, `MaxFichierBytes` (2 Mo) |
| `internal/design/fichiers_test.go` (nouveau) | Noms, save/load/versions/rollback, refus, contenu `<script>`/`{{...}}` accepté |
| `internal/handlers/admin_fichiers.go` | POST save (kill-switch, CSRF, audit, reload gabarits), GET versions, POST activer, `StaticFile` (fallback base → disque, no-cache), `buildFichiers` lit la base pour static/* |
| `internal/handlers/admin_fichiers_test.go` | POST enregistrer, refus (401/403/400/405), versions + rollback, StaticFile (base no-cache / disque cache long) |
| `cmd/store/main.go` | Routes POST `/admin/api/fichiers`, GET/POST versions ; `/static/*` et `/assets/*` via `app.StaticFile()` ; `withCache` mort supprimé |
| `web/admin/studio/studio.html` | Textarea sans `readonly`, bouton **Enregistrer** + **Historique**, section `#codeVersions`, hints corrigés (plus de « Lecture seule », plus de fausse validation `<script>`/`{{...}}`) |
| `web/admin/studio/js/controller.js` (v8) | `saveCodeFile()`, `loadFichierVersions()`, `renderFichierVersions()`, `toggleFichierVersions()`, `activateFichierVersion(n)`, `refreshCodeFiles()` ; toast sans « lecture seule » |
| `web/admin/studio/js/view.js` (v6) | `loadFile` mémorise `f.saved` + charge l'historique ; `updateCodeMeta` affiche « ● Modifié » |
| `internal/handlers/admin_design_test.go` | Versions JS attendues (view v6, controller v8) |

## Sécurité

- **Kill-switch** `DISALLOW_FILE_EDIT=1` : POST et activer → 403, GET reste 200
  (lecture seule). Vérifié en live (serveur dédié port 18098).
- **Liste fermée** des noms (`design.FichierNom`) : `templates/static/*.html`
  (nom simple, `.html`) + les 3 fichiers static. Traversée (`../`), chemin
  absolu, autre dossier/extension → 400. Le faux `theme.css` reste impossible.
- **Auth admin + CSRF** (X-CSRF) sur POST et activer ; 401 sans auth, 403 sans CSRF.
- **Audit** de chaque enregistrement / activation (`fichier.enregistrer`,
  `fichier.activer`).
- **Limite 2 Mo** par fichier.
- **Aucune validation de contenu** : le code du site contient légitimement
  `<script>` et `{{...}}` (gabarits `_top.html`…). La sécurité, c'est
  l'allowlist + l'auth + le versionnage, pas le contenu.

## Risques (dits dans le plan, inchangés)

- Une faute de syntaxe dans `app.js` casserait le site. Mitigation :
  versionnage + rollback 1 clic (Historique), avertissement dans le Studio.
  Pas de validation JS côté serveur (le serveur Go ne fait pas tourner node).
- Les modifs vivent en base (sauvegardée chaque nuit par `backup.sh`), **pas
  dans Git**. Un export vers Git sera à prévoir si Zelef veut les versionner.

## Vérification

- `gofmt` propre, `go build` + `go vet` + `go test ./...` verts (parité
  v3.25 + design OK), `node --check` OK sur les 4 JS du Studio.
- **CDP C4 : 17/17** (`cdp_c4_9223.js`) — éditeur modifiable, bouton
  Enregistrer, indicateur « Modifié », toast « Fichier enregistré (version
  1/2) », historique 2 versions + bouton Réactiver, rollback v1 (marqueur v2
  parti, v1 présent), aperçu = vrai site, 0 erreur JS.
- **Vérifié par curl** : page apropos publique = marqueur v1 (pas v2) après
  rollback ; `/static/theme.js` = contenu base + `Cache-Control: no-cache` ;
  versions API (v2 inactive, v1 active) ; POST sans auth → 401.
- **Kill-switch en live** : POST → 403, activer → 403, GET → 200.
- **Non-régression** : CDP C1/C2/C3 **38/38** (le test C2 « lecture seule » a
  été mis à jour : l'éditeur est maintenant modifiable — c'est le but de C4),
  CDP 3.1 **12/12**.
- Captures : `shot-c4-code.png` (éditeur + vrai site), `shot-c4-code-foot.png`
  (bouton Enregistrer + Historique avec Réactiver).

## Incident de test (pas un défaut du code)

Pendant la non-régression, le serveur redémarré n'a pas pu se lier au port
(« address already in use ») et l'ANCIEN serveur continuait de servir
l'ancienne base — le check A1 échouait alors que le code était bon. Cause :
mauvais PID tué au redémarrage (leçon déjà notée : vérifier le log du nouveau
serveur et confirmer que le PID qui sert est le nouveau). Après redémarrage
propre : 38/38.

## Déploiement (agent)

Comme d'habitude : extraire l'archive à la racine du dépôt, `go build ./...`
+ `go vet ./...` + `go test ./...`, `docker compose up -d --build
--force-recreate`. Aucune migration, aucun changement de config. Vérifier :
accueil 200, boutique 663, `/admin/design` + `/admin/api/fichiers` → 401
sans auth, Studio Code : éditeur modifiable + Enregistrer + Historique.

## Reste

C4 ferme le T3. Reste les phases **3.2 → 3.6** (rendu boutique depuis le
design, une sous-phase à la fois, validées par Zelef).
