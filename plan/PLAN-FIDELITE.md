# Plan fidélité — ton design à 100 %, Go derrière (stack de l'agent)

Décisions figées : URLs `.html` gardées · htmx vendored local · bouton WhatsApp
identique + copie serveur · `/admin` dans le style du site.

## 0. Référence (lecture seule, constatée dans le tar)
- Cartes boutique : `data-search/data-store/data-sub/data-live-sub/data-main/marque/prix/dispo/etat`
  + `data-spec-*` (facettes `window.MEDIATEC_FACETS`), `data-add-cart` + slug/nom/prix/image.
- Fiche produit : `.pdetail`, `.gal`, `.pr`, `button data-add-cart data-qty="#qty"`,
  onglets `data-tabs`, `table.specs`, similaires, widgets, news.
- Chrome commun : topbar, `header.main` (search GET `boutique.html`), `nav.main` +
  méga-panel (11 `data-mega-top`, `data-mega-cols`), drawer (`data-cart-*`),
  modale auth (`data-auth-*`), footer. Préfixe `""` à la racine, `"../"` dans
  `produits/` et `categories/` (`data-prefix` sur le drawer).
- Hub boutique : 8 `data-hub-card` (photo+nom+compteur+chevron), liens `?store=/?sub=/`, `?all=1`.
- JSON : 171 fiches (`produits-v3.json`), `CAT_MAP` + normalisation recherche dans
  `build_site_v3.py`, 198 pages, 286 tests `test-v3.js` OK.

## 1. Gel du design
- `site-orig/` (gitignoré) = étalon visuel + tests d'origine.
- Commités : `assets/style.css`, `assets/app.js` (inchangé), `assets/facets.js`,
  `assets/htmx.min.js` (pinné), `templates/**/*.html` (HTML d'origine découpé).
- `images/` en volume (existant). `sources/` = doc de référence.

## 2. Découpage templates (aucune classe / data-* / effet modifié)
- `base` + `partials/chrome` (head+topbar+header+nav+méga), `partials/drawer`
  (panier+toast+auth), `partials/footer`, param `.Prefix`.
- `partials/card`, `partials/hubcard`, `partials/facettes`, `partials/pagination`.
- Pages : `index` (slideshow+sections), `boutique` (hub+liste+filtres+tri+pagination,
  `?store= ?sub= ?cat= ?q= ?all=1`), `categorie` (= boutique filtrée),
  `produit` (pdetail+onglets+similaires), `devis/suivi/compte/parametres`
  (mêmes `.acct-form` en POST), statiques + groupe (corps verbatim + chrome partagé).
- Méga-menu : markup identique, compteurs depuis la DB.

## 3. Routes chi (remplace le mux stdlib)
- GET `/.html` : `/index.html /boutique.html /categories/*.html /produits/*.html`
  + devis/suivi/compte/parametres/services/apropos/contact/groupe(7)/cgv/livraison/
  confidentialite/mentions-legales + `/` (= index.html) + `/health` + `/assets/*` + `/images/*`.
- POST `/panier/ajouter|/panier/vider` (303 sans JS, fragment htmx avec JS ;
  `localStorage` aussi écrit côté client).
- Clic WhatsApp : lien `wa.me` inchangé, le JS poste en plus une copie serveur.
- POST `/auth/inscription|/auth/connexion|/auth/logout` (sessions DB + bcrypt 12).
- `/admin` même style : CRUD produits, commandes (statuts, moyens
  whatsapp/wave/orange/mtn/livraison/virement), import JSON.

## 4. DB + import
- `002_fidelite.sql` : `sous_categorie, live_sub, etat, prix_old, store, sub`.
- Import Go étendu : `CAT_MAP`, `data-spec-{slug}` via labels de `MEDIATEC_FACETS`,
  index de recherche recette v3.25 (tous les mots, tout ordre, accents/tirets ignorés).

## 5. htmx sans rien casser — `app.js` inchangé
hx-get recherche/filtres/pagination → fragments de cartes identiques, hx-boost
liens, CSRF via `<meta>`, badge `data-cart-count` rafraîchi serveur. Sans JS :
formulaires natifs.

## 6. Preuve de fidélité
- Captures 1440px + 360px vs `site-orig/` : identiques.
- Tests Go = assertions `test-v3.js` : hub 8 cartes, `hp 650a / imprimante /
  serveur / wifi / tablette`, 0 lien interne cassé sur toutes les routes `.html`.
- `go vet`, Lighthouse, CSP Report-Only.

## 7. Ordre d'exécution (branche `design-fidele`)
1. Extraction + chrome partagé + chi + htmx. 2. boutique. 3. produit.
4. index + statiques. 5. panier-miroir + auth. 6. /admin. 7. migration + import.
8. tests + staging :8081. 9. bascule :8080 + push main.
