# RAPPORT — Phase 2 Studio : persistance du design complet via API JSON

Base : `271f86c` (phase 1 déployée). Date : 2026-11-14.

## Ce qui change

Le Studio ne publie plus seulement le thème : il lit et écrit **tout le design**
en un seul document JSON, versionné, avec retour en arrière en 1 clic.
La boutique continue d'utiliser le thème publié — son rendu ne change pas.

### Côté serveur (Go)

- **Nouveau `internal/design/design.go`** — le modèle :
  - `Load` (version active, ou rien si jamais publié), `Save` (nouvelle version,
    transaction), `ListVersions`, `Activate` (retour en arrière).
  - Stockage : table `design_files` (T0), nom `design.json`, version + auteur +
    date, ajout seul (jamais d'UPDATE/DELETE du contenu).
  - `Validate` avant tout enregistrement : JSON valide, taille ≤ 2 Mo, clés
    top-level fermées (les 13 clés de `snap()`, sans `files`), aucun contenu
    actif (`<script>`, `<iframe>`, `<form>`, `{{…}}`, `javascript:`,
    attributs `on*=`), aucune valeur **numérique** sous les noms
    prix/stock/total (un vrai prix serait un nombre).
  - `ThemeFromDesign` : extrait le thème (`front` + bandeau) en partant du
    thème courant — **les 9 sections d'accueil du serveur sont préservées**
    (le design ne les contient pas).
- **API dans `internal/handlers/admin_design.go`** (+ routes `cmd/store/main.go`) :
  - `GET /admin/api/design` — design actif (ou `design:null` si rien publié).
  - `POST /admin/api/design` — valide, enregistre (nouvelle version), extrait
    le thème vers `settings` (`theme_version` +1), audit `design.publier`.
  - `GET /admin/api/design/versions` — liste des versions (plus récente d'abord).
  - `POST /admin/api/design/versions/{n}/activer` — réactive une ancienne
    version + republie son thème (`theme_version` +1), audit `design.activer`.
  - Auth admin partout. CSRF : `checkAdminCSRF` accepte désormais aussi
    l'en-tête `X-CSRF` (le corps JSON n'est pas un formulaire) — même jeton,
    autre transport. Helper `writeJSON` (remplace les réponses en ligne).
- **GET `/admin/design`** injecte en plus le design actif :
  `window.__STUDIO_DESIGN__` + `__STUDIO_DESIGN_VERSION__` (null + 0 sans
  design publié). Le JSON est re-marshalé par Go (`<` → `\u003c`) : aucune
  valeur ne peut fermer le bloc `<script>`.
- Le POST `/admin/design` (thème, phase 1) est **conservé** (hors scope de le
  retirer ; le Studio ne l'utilise plus).

### Côté Studio (JS)

- **`applyServerDesign()`** (controller.js) : applique le design injecté au
  démarrage, avant le premier rendu et avant la capture de `pubSnap`.
  **Fusion avec l'état courant** (qui porte les défauts) : un design partiel
  (header sans `navLinks`, footer sans `cols`…) ne fait plus planter l'aperçu
  — les valeurs du design gagnent, les défauts comblent les trous.
  `front._serverSections` est retiré à l'envoi (métadonnée serveur, pas du
  design) ; `studioV16` (dans `front._studioV16`) est réaligné après la fusion.
- **`designJSON()`** : les 13 clés de `snap()`, sans `files` (le contenu des
  fichiers de code, c'est T3).
- **`publishStudio()`** : POST JSON vers `/admin/api/design` (CSRF en en-tête
  `X-CSRF`). Toast « design vN, thème vM ». Le wrapper v13 (contraste,
  historique) est conservé au-dessus.
- **Historique** (view.js `renderHistoryV13` + `loadServerVersionsV13` /
  `activateServerVersionV13` dans controller.js) : les versions serveur sont
  listées avec un bouton **Réactiver** (POST + rechargement de la page avec
  cette version). L'historique de session (brouillons restaurables) est
  conservé en dessous.
- **Versions des fichiers JS** : `view.js` et `controller.js` passent en `v=2`
  dans `studio.html` (les fichiers modifiés ; `model.js`/`preview.js` restent
  en `v=1`). Sans ça, le navigateur garde l'ancienne version en cache
  (`max-age=3600`) et le Studio tourne avec l'ancien code.
- Notes de l'interface mises à jour : « Publier » met le **design** en ligne
  (plus seulement le thème).

## Validation serveur — point de vigilance (vérifié par les tests)

Les clés prix/stock/commande/total existent dans l'état **légitime** du Studio,
mais seulement comme **libellés** (strings : « En stock », « Sous-total »,
« Min »…) ou bascules d'affichage (`showStockOnly`). Un refus par nom de clé
aurait rejeté le design légitime. La règle retenue : interdire les valeurs
**numériques** sous ces noms exacts (un vrai prix ou un vrai stock serait un
nombre) + fermer la liste des clés top-level (aucune clé `products`/`orders`
ne peut entrer). Tests : `TestValidateOK`, `TestValidateRefusePrixNumerique`,
`TestValidateRefuseCleInconnue`, `TestValidateRefuseScript`,
`TestValidateRefuseGabari`.

## Tests

- `go build ./... && go vet ./... && go test ./...` verts (gofmt propre).
- Nouveaux tests `internal/design/design_test.go` (9 tests : validation,
  Save/Load/List/Activate, ThemeFromDesign gazonnier + liquid + sections
  préservées) et `internal/handlers/admin_design_test.go` (4 tests :
  publier+charger, refus 401/403/400, versions+activer, injection du design
  dans la page). Les 3 tests de la phase 1 passent toujours.
- 4 fichiers JS `node --check` OK.

## Test live (serveur local port 18099, base de test 664 produits)

### API (curl)
- GET `/admin/api/design` vide → `design:null, version:0`. POST sans CSRF → 403.
- POST v1 (gazonnier, bandeau « Test Phase 2 v1 ») → design v1, thème v1.
  POST v2 (liquid) → design v2, thème v2. Versions : v2 active, v1 inactive.
  Thème en base : couleur liquid, accent #84CC16, bandeau v2, **9 sections
  préservées**.
- Activer v1 → design v1, thème v3 (gazonnier, #2E5136, bandeau v1),
  sections toujours les 9. Boutique : `data-theme-version="3"` +
  `--brand:#2e5136`.
- Versions sans auth → 401. Activer sans CSRF → 403.

### Navigateur (CDP, 23/23 OK)
- Studio charge, 0 erreur JS. Design v1 injecté (bandeau, accent gazonnier),
  état initial = design serveur, statut **Publié** à l'arrivée.
- Aperçu : bandeau v1 visible. Versions serveur chargées (v1+v2), bouton
  Réactiver présent.
- Modification (bandeau + accent bleu) → **Brouillon**. Publier (bouton UI) →
  toast « design v3, thème v4 », statut **Publié**.
- Rechargement → design v3 injecté (bandeau UI v3, accent bleu), statut Publié.
- Historique : 2 versions réactivables. Réactiver v1 (bouton UI) → page
  rechargée avec le design v1 (bandeau v1, accent gazonnier), statut Publié.
- 0 erreur JS pendant tout le test.
- Capture : `~/workspace/marexstore-test/shot-phase2-studio.png` (Studio après
  réactivation v1, bandeau « Test Phase 2 v1 » visible dans l'aperçu).

### Non-régression
`/`, `/boutique.html?all=1`, `/categories/materiels-reseaux.html`,
`/produits/15-pro.html`, `/admin`, `/admin/systeme`, `/admin/design`,
`/admin/api/design/versions` → 200.

## Bug trouvé et corrigé pendant le test live

**Design partiel → aperçu planté** (`pvHeaderHTML` : `h.navLinks.filter` sur
`undefined`). Le design de test (curl) avait `"header":{}` (sans `navLinks`) :
`applyServerDesign` remplaçait `header` en bloc, l'aperçu plantait avec une
erreur JS au chargement. Corrigé par la **fusion avec les défauts** (voir
ci-dessus) : le design partiel est complété par les valeurs par défaut du
Studio. Leçon : un design incomplet ne doit jamais faire planter le Studio —
les défauts comblent les trous. (Le test de validation `TestValidateOK` couvre
le cas : un design avec `header:{}` et des libellés prix/stock passe.)

## Déploiement (agent sur place)

1. Extraire l'archive à la racine du dépôt (`/home/sam/MarexStore`).
2. `chmod -R a+rX web` (les fichiers du tar arrivent en mode 640, illisibles
   par l'utilisateur non-root du conteneur).
3. `go build ./... && go vet ./... && go test ./...` (doivent être verts).
4. `docker compose up -d --build --force-recreate`.
5. Vérifier : `/admin/design` 401 sans auth, 200 avec ; Publier depuis le
   Studio → design v1 (la prod n'a rien publié : première publication = v1) ;
   boutique 663 produits, thème par défaut intact tant que rien n'est publié.

## Fichiers livrés (10)

- `internal/design/design.go` (nouveau)
- `internal/design/design_test.go` (nouveau)
- `internal/handlers/admin.go` (checkAdminCSRF + X-CSRF, writeJSON)
- `internal/handlers/admin_design.go` (API + injection design)
- `internal/handlers/admin_design_test.go` (4 tests phase 2)
- `cmd/store/main.go` (4 routes API)
- `web/admin/studio/studio.html` (view.js + controller.js v=2)
- `web/admin/studio/js/controller.js`
- `web/admin/studio/js/view.js`
- `RAPPORT-PHASE2-STUDIO.md`
