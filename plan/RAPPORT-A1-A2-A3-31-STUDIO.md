# Rapport — Studio branché sur le vrai site (A1 + A2 + A3) + 3.1

Base : origin/main **9c1b72f** (Phase 3.0). Livraison du plan « Brancher le
Studio sur le vrai site » de Zelef (GO 2026-11-14 21:07 UTC) : A1, A2, A3,
puis B = 3.1 (textes + visibilité).

## Ce qui change

### A1 — L'aperçu est le vrai site (iframe)
- `AdminDesignPreview` (`/admin/design/apercu`) accepte `?page=accueil`
  (défaut) | `boutique` | `fiche` (+ `?slug=` pour la fiche, défaut : le
  premier produit). Il rend la page COMPLÈTE (en-tête + contenu + pied de
  page) avec le thème brouillon (POST ou query string, mécanisme existant).
- `preview.js` : `renderPreview()` ne construit plus de HTML local — il
  pilote l'URL de l'iframe (`/admin/design/apercu?...&page=...`), avec
  debounce ~400 ms et saut si l'URL est inchangée. L'ancienne maquette
  (`renderPreviewMockup`) reste en repli, NON branchée.
- `BoutiqueBodyQueryTheme(s, q, t)` (nouveau, render3.go) pour la boutique
  avec thème brouillon. `IndexBodyTheme` inclut DÉJÀ en-tête + pied de page
  — ne pas les rajouter (en-tête doublé sinon, régression couverte par test).
- Versions JS : preview.js v=4, controller.js v=4, view.js v=3.

### A2 — Snapshot du vrai catalogue
- Nouveau endpoint `GET /admin/api/snapshot` (auth admin) :
  `navLinks` (5), `megaMenu` (11 tops, groupes), `hero` (3 bannières
  réelles : OUKITEL, Kyocera, HP EliteBook 850 G3), `minis` (4),
  `marques` (9), `topVentes` (8, prix boutique), `produits` (14, voir A3),
  `footer` (5 colonnes), `tx` (Tx du design actif, null sinon).
- Source unique : `render.Banners`, `render.Minis`, `render.Marques`,
  `render.Top8Slugs`, `render.NavLinks`, `render.FooterCols` extraits de
  render2.go/render3.go (parité vérifiée, octet pour octet).
- Injecté dans la page Studio (`window.__STUDIO_SNAPSHOT__`) ; `model.js`
  n'utilise plus ses placeholders : `acc.hero.slides` = les vraies
  bannières, `acc.brands.list` = les 9 vraies marques, les mini-casiers =
  les 4 vrais.
- Décision : snapshot construit frais à chaque chargement du Studio (petites
  données) — pas de cache en mémoire ni d'invalidation à gérer.

### A3 — Éditeurs branchés catalogue
- Snapshot `produits` = union (sans doublon) des top ventes (8) + des
  mini-casiers (4) + des produits liés par les bannières du hero (2) =
  **14 produits réels** (nom, slug, prix, prix affiché, image, marque).
- Éditeur de bloc (onglet Front) : nouveau sélecteur « Produit du
  catalogue » (visible pour les blocs Produit/Bannière, masqué pour Texte).
  Choisir un produit remplit titre (nom réel), prix (prix affiché réel) et
  texte (marque). Les champs libres restent possibles.
- Lecture seule : aucun changement DB côté boutique.

### 3.1 — Textes Tx + visibilité dans le rendu
- En-tête (`HeaderTheme`) : textes depuis Tx (`bienvenue`,
  `suivi_commande`, `inscription`, `recherche_ph`, `rechercher`,
  `toutes_cat`, `boutique_ico`, `compte_ico`, `panier_ico`, `prefs_ico`) ;
  bascules `header.showTopbar/showPhone/showEmail/showBoutique/showCompte/
  showPanier/showPrefs` (défauts = visibles).
- Pied de page (`Footer`) : textes depuis Tx (`newsletter_titre`,
  `newsletter_texte`, `newsletter_ph`, `newsletter_btn`, `copyright`) ;
  bascules `footer.showNewsletter/showCols`.
- Bouton méga-menu (`MegaBtn`) : libellé depuis Tx `toutes_cat`.
- `escTx` (nouveau) : échappe `& < > "` mais PAS l'apostrophe — les
  libellés codés en dur en contiennent (S'abonner, S'inscrire) et étaient
  des littéraux bruts ; `esc()` les transformerait en `&#x27;` et romprait
  la parité. La validation serveur (phase 2) bloque déjà `<script>`,
  `<iframe>`, `<form>`, `{{...}}`, `javascript:`, `on*=` dans les designs.
- **Parité sans design : INCHANGÉE.** Les 4 empreintes SHA256 de la phase
  3.0 sont toujours vertes (index, boutique, fiche, catégorie).

## Vérification

- `go build ./...`, `go vet ./...`, `go test ./...` verts (catalog, design,
  handlers, render). `node --check` sur les 3 JS modifiés.
- Nouveaux tests : `TestAdminDesignPreviewPages` (A1), `TestAdminSnapshot`
  + `TestAdminDesignServeStudioInjecteSnapshot` (A2), `TestDesign31Textes-
  EtVisibilite` + `TestDesign31TopbarMasquee` (3.1), snapshot `produits`
  vérifié (14, sans doublon, T440 + OUKITEL présents).
- **Live CDP (serveur local port 18099, Chrome 9223)** :
  - A1 : **16/16** — iframe = vrai `/admin/design/apercu`, en-tête/pied de
    page réels, aucun texte maquette, accent brouillon appliqué.
  - A2 : **16/16** — hero OUKITEL, 9 marques réelles, 4 minis réels.
  - A3 : **14/14** — sélecteur = 14 produits réels ; choisir l'OUKITEL RT3
    remplit « Oukitel RT3 Plus (16 Go RAM / 128 Go) — 200 000 F CFA ».
  - 3.1 : **12/12** — texte modifié dans le Studio → Publier (design v3,
    thème v3) → le nouveau texte est sur le site public (curl) ET dans
    l'aperçu iframe ; 0 erreur JS.
- Bugs trouvés et corrigés pendant les tests : srcdoc qui gagne sur `src`
  (retirer l'attribut, ne jamais le remettre à vide) ; en-tête doublé sur
  l'accueil (IndexBodyTheme l'inclut déjà) ; `esc()` sur les apostrophes
  (escTx) ; retour à la ligne des icônes (chaque icône se termine par `\n`).

## Déploiement

- Extraire à la racine du dépôt, `chmod -R a+rX internal cmd web`.
- `go build ./... && go vet ./... && go test ./...`.
- `docker compose up -d --build --force-recreate`.
- Aucune migration, aucune nouvelle route à part `/admin/api/snapshot`,
  aucun changement Dockerfile (le Studio est déjà copié via `COPY web`).
- Vérifier : `/admin/design` (Studio, 401 sans auth), l'aperçu iframe =
  le vrai site, le snapshot (`/admin/api/snapshot`, 401 sans auth),
  publier un texte dans le Studio → il apparaît sur le site public.

## Reste (phases 3.2 à 3.6, plan inchangé)

- 3.2 en-tête + pied de page (libellés/visibilité/ordre des navLinks et
  des colonnes du footer, navBg, badgeColor, sticky) ;
- 3.3 boutique + fiche (tri par défaut, nb par page, vues, facettes,
  sidebar, sections fiche, boutons) ;
- 3.4 accueil (sections `acc` pilotées par le design) ;
- 3.5 panier + devis + compte (textes seulement) ;
- 3.6 méga-menu affichage seulement, EN DERNIER.
- Prix, stock, totaux et logique de commande ne sont JAMAIS touchés.
