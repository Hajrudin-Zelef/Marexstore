# RAPPORT — Corrections C1 + C2 + C3 (contenus maquette du Studio)

**Date :** 2026-11-14
**Base :** origin/main 4aac904 (A1+A2+A3+3.1 déployés)
**Contexte :** audit complet du Studio (12 problèmes, voir `AUDIT-STUDIO-CONTENU-MAQUETTE.md`).
Zelef : « Même la personnalisation en code a des fichiers qui n'ont rien à voir avec l'existant » +
« vérifie tout pour voir s'il n'y a pas d'autres qui ont le même souci, on va tout corriger ».
GO « Corrige tout » le 2026-11-14 23:01 UTC pour **C1+C2+C3**. **C4 (écriture / T3) NON inclus.**

## C1 — Contenus réels partout (défauts maquette remplacés)

### C1a — Snapshot étendu (`internal/handlers/admin_snapshot.go`, `internal/render/render3.go`)
- `bandeauSombre` : le vrai bandeau sombre — « Clés USB Imation & disques durs WD Elements »
  / « Stockage fiable pour tous vos besoins — demandez votre devis. » / « VOIR LE PRODUIT » / « CONTACT ».
  Extraction `DarkBanner` + `DarkBannerData` (render3.go, parité vérifiée).
- `offreSpeciale` : le vrai produit deal — Kyocera TASKalfa 3011i, avec son accroche réelle
  (« Multifonction laser monochrome A3, 30 pages par minute, écran tactile 9 pouces. »).
  `DealSlug = "kyocera-taskalfa-3011i"` au niveau package.
- `menuFams` : les **11** tops (il en manquait 4 : Téléphones & Tablettes 14, Reconditionné & Casi neuf 6,
  Accessoires Téléphones & Tablettes 0, Uncategorized 0), compteurs calculés sur le catalogue réel.
- `pagesInfo` : les **vraies** pages apropos / cgv / contact / livraison, lues depuis `a.Bodies`
  (déjà chargé par reload(), reflète les surcharges design_files — pas de relecture disque par requête).

### C1b — Défauts réels dans le Studio (`model.js`, `controller.js`)
- `front.customCSS` par défaut → `''` (le CSS `.pv-hero{…}` de la maquette est parti).
- Lignes Front : hero = « OUKITEL RT3 PLUS / TABLETTE PROFESSIONELLE / 200.000 CFA » ;
  minis = les 4 vrais mini-casiers ; produits = vrais produits du catalogue.
- `menu.fams` = 11 tops avec les vrais compteurs (281/211/60/37/36/19/14/6/0/0/0).
- `acc.hero` = les 3 vraies bannières (OUKITEL, DÉCOUVREZ NOS COPIEURS KYOCERA, HP ELITEBOOK 850 G3) ;
  `acc.dark` = vrai bandeau sombre ; `acc.find` = Kyocera TASKalfa 3011i + accroche ;
  `acc.brands.list` = les 9 vraies marques.
- Tx alignés sur le site : `recherche_ph` = « Rechercher un produit, une référence… »,
  `aide_texte` = « Une question avant d'envoyer votre demande? Notre équipe vous répond. »,
  `cpt_sync` = le vrai texte de synchronisation (render2.go:211).
- SEO : titre = « Accueil — Mediatec Store » (le vrai), description vide (le site n'en a pas).
- `applyServerSnapshot()` étendu : bandeau sombre, offre spéciale, menuFams, pagesInfo,
  ligne hero (titre = titre + 1re partie du sous-titre, ex. « OUKITEL RT3 PLUS »),
  ligne produits (produits[0..1]).

### C1c — Blocs inexistants retirés (`model.js`, `controller.js`)
- `acc.av` (Avantage 1/2/3/4) supprimé — aucun bloc avantages sur le vrai site.
- `acc.hero` panneau blanc (panel/panelT/panelX) supprimé — aucun panneau de ce type dans le vrai hero.
- Tx `suite_titre`/`suite_texte` supprimés — aucune carte « Étapes suivantes » sur la page devis
  (la seule carte latérale réelle est « Besoin d'aide? »).
- Éditeur Accueil nettoyé (groupes panneau + avantages retirés de ACC_D / ACC_T).

### Correctif validation (découvert par le test 3.1, inclus dans C1)
- **Bug :** après C1, la publication échouait (HTTP 400 « contenu interdit : "<form" »).
  Cause : `txtPages` contient désormais le vrai HTML des pages d'info, et contact.html
  a un `<form>` légitime — la validation Phase 2 l'interdisait partout.
- **Fix (`internal/design/design.go`) :** la validation suit désormais la clé de premier niveau.
  Sous `txtPages` uniquement, `<form>` est autorisé (liste `forbiddenValueTxtPages`).
  `<script>`, `<iframe>`, `{{…}}`, `javascript:`, `on*=` restent interdits partout, y compris sous txtPages.
- **Test ajouté** (`design_test.go`, `TestValidateTxtPagesFormOK`) : `<form>` accepté sous txtPages,
  `<script>` toujours refusé sous txtPages.

## C2 — Mode Code : les vrais fichiers, en lecture seule

### Nouveau endpoint `GET /admin/api/fichiers` (`internal/handlers/admin_fichiers.go`)
- Auth admin, GET uniquement. Retourne `{"ok":true,"fichiers":[...]}`.
- **24 fichiers** : les pages statiques depuis `a.Top/a.Drawer/a.Foot/a.Bodies`
  (`_top.html`, `_drawer.html`, `_foot.html`, + les pages triées) et
  `static/app.js`, `static/style.css`, `static/theme.js` lus depuis le disque.
- Chaque entrée : {nom, lang, lignes, contenu}. **Le faux `static/theme.css` a disparu**
  (ce fichier n'existe pas — il n'était que dans la maquette).
- Route ajoutée dans `cmd/store/main.go` (après `/admin/api/snapshot`).

### Studio (`model.js`, `controller.js`, `view.js`, `studio.html`)
- `FILES` = `{}` au départ, rempli par `loadCodeFiles()` (fetch `/admin/api/fichiers`, une fois).
- L'arbre de fichiers est reconstruit dynamiquement (24 boutons, icônes VS Code par type).
- L'éditeur affiche le vrai contenu : app.js = **1203 lignes** (65 303 caractères), _top.html = 23 lignes,
  apropos.html = 62 lignes.
- **`snap()` / `restore()` n'incluent plus `files`** (220 Ko hors de l'historique undo et du design publié).
- **Lecture seule** : textarea `readonly`, hint « Lecture seule pour l'instant — l'écriture arrive avec le T3 ».
  L'écriture = C4 / T3, sur feu vert explicite uniquement.

### Bug corrigé en route (C2, trouvé par le test CDP)
- **Bug :** `TypeError: Cannot read properties of undefined (reading 'lang')` au chargement du Studio
  (`syncEditor` lisait `FILES[currentFile].lang` alors que FILES est vide et `currentFile=''`
  tant que les fichiers ne sont pas chargés). L'erreur **avortait le bloc d'init** — le bouton
  « Pages d'information » n'était jamais créé.
- **Fix (`view.js`) :** `syncEditor()` et `updateCodeMeta()` tolèrent l'absence de fichier courant
  (lang par défaut `'txt'`, méta vide).

## C3 — Nettoyage du code mort maquette + bug Pages d'info

- `renderPreviewMockup()` **supprimé** de `preview.js` (il n'avait AUCUN appelant — le commentaire
  qui le présentait comme repli d'erreur était faux).
- `PV.html` rendu défensif : si `#theme`/`#root` sont absents du document de l'iframe, il ne fait rien
  au lieu de planter.
- Nouveau `mockupShellHTML()` (coque srcdoc avec PV_BASE + pvCss + `#theme`/`#root`).
- Init (`controller.js`) : `pvI.srcdoc = mockupShellHTML()` au démarrage.
- **Bug Pages d'info corrigé :** l'onglet « Pages d'information » utilisait encore `PV.html`
  (le setter maquette) qui plantait depuis qu'A1 a remplacé le srcdoc par le vrai site
  (TypeError sur `#theme`/`#root` absents). La branche v16 fait désormais :
  `pvReady=false; pvI.srcdoc=mockupShellHTML(); renderInfoPageV16()`
  (`pvReady=false` est OBLIGATOIRE — sinon PV.html écrit dans l'ancien document du vrai site
  et le contenu est perdu).
- Les fonctions pv* (pvHeaderHTML, PV_BASE, etc.) ne sont **pas** supprimées : elles sont encore
  référencées par les wrappers v13/v14/v16 et `renderInfoPageV16`. Leur retrait propre
  demande un passage de démêlage dédié.

## Versions des assets (à bumper ensemble)
- `studio.html` + test : model.js **v=2**, view.js **v=5**, preview.js **v=5**, controller.js **v=7**.

## Vérification
- `go build ./...`, `go vet ./...`, `go test ./...` : **verts** (dont le nouveau test
  `TestValidateTxtPagesFormOK` et `TestAdminFichiers`).
- `node --check` : OK sur les 4 JS du Studio.
- **CDP C1/C2/C3 : 38/38 OK** (serveur local port 18099, base fraîche) :
  bandeau sombre / offre / menu 11 tops / pages info réelles, défauts réels (hero OUKITEL RT3 PLUS,
  customCSS vide, Tx/SEO réels), avantages + panneau + suite retirés, mode Code (24 fichiers,
  pas de theme.css, app.js 1203 lignes, lecture seule), Pages d'info sans erreur,
  A1/A2 non-régression, 0 erreur JS.
- **CDP 3.1 : 12/12 OK** (publication d'un texte → design v3/thème v3 → texte en ligne + dans l'aperçu).
  Note : un 1er passage a échoué sur « le nouveau texte est dans l'aperçu » — flake de timing
  (l'iframe se recharge ~400 ms après la publication ; le re-passage est vert). Artefact du test, pas du code.
- Captures vérifiées visuellement : `shot-c1-studio.png` (Front, vrai site),
  `shot-c1-pagesinfo.png` (Pages d'info, contenu réel « Fondée en 2004, MEDIATEC… »),
  `shot-c1-code.png` (Code, vrai app.js dans l'éditeur, arbre réel).

## Fichiers livrés (14)
- `cmd/store/main.go` (route /admin/api/fichiers)
- `internal/design/design.go` (validation txtPages)
- `internal/design/design_test.go` (test txtPages)
- `internal/handlers/admin_design_test.go` (versions JS bumpées)
- `internal/handlers/admin_snapshot.go` (snapshot étendu)
- `internal/handlers/admin_snapshot_test.go` (tests C1a)
- `internal/handlers/admin_fichiers.go` (**nouveau**)
- `internal/handlers/admin_fichiers_test.go` (**nouveau**)
- `internal/render/render3.go` (DarkBanner, DealSlug)
- `web/admin/studio/js/controller.js` (v7)
- `web/admin/studio/js/model.js` (v2)
- `web/admin/studio/js/preview.js` (v5)
- `web/admin/studio/js/view.js` (v5)
- `web/admin/studio/studio.html`
- `RAPPORT-C1-C2-C3-STUDIO.md` (ce fichier)

## Hors scope (rappel)
- **C4 / T3 (écriture des fichiers, `/admin/fichiers` POST) : NON inclus, en attente du feu vert explicite de Zelef.**
- Phases 3.2 → 3.6 inchangées (prochaine : 3.2 en-tête + pied de page).
