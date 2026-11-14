# ADDENDUM — Plan Studio 100% : extensions pour couvrir tous les volets

Date : 2026-11-14. Statut : PROPOSITION (en attente du OK de Zelef, puis à transmettre à l'agent).
Base : PLAN-STUDIO-100-2.md (révisé). Objectif : le studio couvre 100% des volets du front, pour junior (interactif) et senior (code), sans jamais toucher au backend Go, aux prix, au stock ni à la logique de commande.

## Constat (vérifié dans le code à 46859a7)

Le plan révisé couvre bien : thème, identité, bandeau, liens libres, sections d'accueil, hero, widgets, grille par défaut, onglets fiche, CSS libre, et en T3 l'éditeur de code (chrome + pages statiques). Il ne couvre pas :
- la boutique au-delà de la grille (tri, nombre par page, facettes) ;
- la fiche produit au-delà des onglets (specs, produits similaires, boutons) ;
- le panier et le tunnel devis/commande ;
- l'édition des textes côté junior (titres de sections, libellés, contenu des pages statiques) ;
- l'affichage du méga-menu (ordre, visibilité) ;
- le SEO.

## A1 — Boutique (T2, interactif)

Réglages (settings, upsert existant) :
- `boutique_tri_defaut` : pertinence | prix_croissant | prix_decroissant | nouveautes | nom
- `boutique_nb_par_page` : 12 | 24 | 48 | 96
- `boutique_facets_visible` (JSON, liste de slugs) + `boutique_facets_ordre` (JSON)
- `grille_defaut` déjà au plan (conservé).

Portée : les valeurs des facettes restent calculées depuis les données (règles existantes, `catalog.FacetValues`, `facets.js`) ; le studio ne contrôle que leur visibilité et leur ordre. Le tri et le nombre par page sont des réglages de lecture du rendu.

Tests : parité (défaut = comportement actuel), `TestBoutiqueTriDefaut`, `TestBoutiqueFacetsOrdre`.

## A2 — Fiche produit (T2, interactif)

Réglages (settings) :
- `fiche_specs_visible` (bool) — afficher / masquer le tableau « Caractéristiques »
- `fiche_similaires_visible` (bool) + `fiche_similaires_nb` (4 | 8) — « Produits similaires » (existe déjà, `render2.go:519`, même catégorie)
- `fiche_onglets` déjà au plan (conservé : libellés et ordre des onglets Description / Marque / Avis)
- `fiche_bouton_devis_texte`, `fiche_bouton_wa_texte` — libellés des boutons « Demander un devis » / « Commander sur WhatsApp »

Portée : cosmétique + visibilité seulement. Le prix, le stock, le bouton ajouter et la quantité ne sont jamais personnalisables.

Tests : parité, `TestFicheSimilairesNb`, `TestFicheSpecsVisible`.

## A3 — Panier + devis / commande (T2, textes et visibilité seulement)

Réglages (settings) :
- `panier_titre`, `panier_texte_vide`, `panier_bouton_commander`
- `devis_titres_etapes` (JSON, 4 titres) + `devis_bouton_wa_texte`
- `commande_confirmation_texte` (message affiché après une commande)

Portée : textes et visibilité seulement. Les étapes du devis, le calcul des totaux, le stock et la logique de commande ne sont JAMAIS personnalisables (logique métier).

Tests : parité, `TestPanierTextes`.

## A4 — Éditeur de textes (T2, pour juniors)

- Nouvelle table `005_design_textes.sql` (via `EnsureSchema`, jamais destructif) : `design_textes(cle TEXT PK, valeur TEXT)` — clés stables pour les libellés de l'interface (titres des sections d'accueil, boutons, messages).
- Contenu des pages statiques : table `pages_contenu(slug TEXT PK, titre TEXT, contenu_html TEXT)` — éditable dans le studio (textarea), repli sur le fichier statique si vide. Pages concernées : apropos, cgv, confidentialite, livraison, mentions-legales, contact, nos-services, etc. (celles de `templates/static/`).

Portée : textes et contenu informatif seulement. Aucun prix, aucun stock, aucune logique.

Tests : parité (défaut = textes et fichiers actuels), `TestTextesSurcharge`, `TestPageContenuFallback`.

## A5 — Méga-menu, affichage seulement (T2, le plus délicat)

Réglages (settings) :
- `mega_ordre` (JSON, liste ordonnée des slugs LiveTop) — ordre d'affichage des familles
- `mega_visible` (JSON, slugs visibles) — masquer une famille sans supprimer la catégorie
- `mega_groupes_ordre` (JSON, par famille) — ordre des groupes thématiques dans une famille

Portée : AFFICHAGE seulement. La taxonomie (catégories, produits, `LiveCats`, `MegaGroups`) n'est pas modifiée ; seuls l'ordre et la visibilité dans le méga-menu sont contrôlés. Le méga-menu reste la seule navigation catégories.

Le plus délicat : touche la navigation principale. À faire après A1–A4, avec tests renforcés (captures desktop et mobile).

Tests : parité (défaut = ordre actuel), `TestMegaOrdre`, `TestMegaVisible`.

## A6 — SEO (optionnel, petit)

Réglages (settings) :
- `seo_titre_accueil`, `seo_description_accueil`, `seo_titre_boutique`, `seo_description_boutique`, `seo_og_image` (media ID)

Appliqué dans le `<head>`. Petit, faible risque. Optionnel (pas du design à proprement parler, mais un volet du front).

## Correction au plan (T3, allowlist)

L'allowlist du plan dit `bodies/*.html`, mais les pages statiques sont directement dans `templates/static/` (apropos.html, cgv.html, contact.html, devis.html, livraison.html, etc. — pas de dossier `bodies/`). Vérifié dans le repo à 46859a7. Corriger l'allowlist en `templates/static/*.html` (pages statiques) + `_top.html, _foot.html, _drawer.html` (déjà là).

## Phasage

- Ces ajouts sont du **T2 étendu** (interactif), après P1, avec T0+T1. Ils n'ajoutent rien à T0+T1.
- Ordre recommandé : A1 (boutique) + A2 (fiche) + A4 (textes) d'abord, puis A3 (panier/devis), puis A5 (méga-menu, le plus délicat) en dernier, puis A6 (SEO, optionnel).
- T3 reste en décision après T2 (inchangé). P1, P2, P3 inchangés.
- Impact : T2 devient plus gros (environ +40 à 60% de travail sur T2). Le reste du plan est inchangé.

## Limites qui restent (saines)

- Prix, stock, totaux, logique de commande, auth : jamais personnalisables via le design.
- Backend Go : jamais édité par le studio. Les pages dynamiques (accueil, boutique, fiche) restent rendues par Go ; le studio les pilote par réglages, pas par gabarits. C'est pour ça que le « 100% » porte sur le pilotage par réglages + CSS/JS, pas sur la réécriture des pages.
- Le méga-menu reste la seule navigation catégories.
