# Module Produits premium — Marexstore

**Date :** 2026-11-13 — **Base :** origin/main 933ba2d + dashboard premium (déjà déployé).
**Objectif :** liste Produits premium + fiche produit refaite comme la fiche employé (aperçu live pendant le remplissage).

## Contenu

| Fichier | Changement |
|---|---|
| `internal/handlers/admin.go` | `AdminProduits` (mini-stats, filtre catégorie, tableau enrichi) + `AdminProduitEdit` (carte aperçu live + formulaire grille 2 col) |
| `internal/handlers/admin_dash.go` | **Nouveau** (dashboard) + helpers partagés : `topLabels`, `fmtMoney`, `miniStat` |
| `internal/render/admin_chrome.go` | `admin.css?v=2` → `?v=3`, `admin.js?v=1` → `?v=2` (cache) |
| `static/admin.css` | Styles liste premium + fiche (style fiche employé) + **correctif `body.mx a`** |
| `static/admin.js` | Aperçu live de la fiche (mise à jour pendant la saisie) |

## 1. Liste Produits

- **Bandeau de 4 mini-stats** cliquables (Total, En stock, Rupture, Sans prix) avec accent couleur, dans le style des KPI du dashboard.
- **Filtre par catégorie** (nouveau) : liste déroulante « Toutes catégories » avec compteurs (Matériels réseaux 281, Accessoires 211, Imprimantes 60…). Se combine avec la recherche, la vue et le tri ; conservé dans la pagination.
- **Tableau enrichi** : vignette 52 px, nom en gras + catégorie en petit en dessous, prix formaté avec espaces (« 620 000 F CFA » au lieu de 620000), « devis » en discret, badge dispo et actions Modifier/Voir/Dupliquer conservés.
- **Tri corrigé** : la liste déroulante garde maintenant la valeur choisie (avant, elle retombait toujours sur « Tri : nom »).

## 2. Fiche produit (style fiche employé)

- **Carte aperçu à gauche** (comme le badge employé) : bandeau vert foncé (#2E5136) avec la photo du produit en cadre blanc, le nom en blanc, la **marque en badge citron** (comme le poste), puis un encadré blanc avec les infos en lignes (Prix, Catégorie, Stock, Statut) et la zone photo en pointillés (chemin image + upload).
- **Formulaire à droite** : étiquettes en MAJUSCULES vertes avec espacement (comme la fiche employé), champs en fond vert très clair (#f9fbf9) avec bordure 1,5 px et coins 10 px, en grille 2 colonnes (Nom + Marque, Prix + Prix boutique, Dispo + Stock…). Accroche et Description en pleine largeur.
- **Aperçu live** : la carte se met à jour **pendant la saisie** (petit JS inline, sans librairie) — nom, marque, prix (formaté « 620 000 F CFA » ou « devis »), dispo (badge), catégorie, stock, photo. Testé : 20/20 en jsdom.
- **Boutons** : Enregistrer (vert foncé, comme le bouton principal de la fiche employé) + Voir (contour).
- **Aucun changement côté données** : mêmes champs, même enregistrement, même upload, même CSRF. Le formulaire « Nouveau produit » prend la même forme (aperçu vide qui se remplit à la saisie).

## 3. Correctif CSS (bug pré-existant, trouvé pendant la vérification)

- **La règle `body.mx a` (liens bleus) écrasait les boutons et la navigation** : le texte du bouton « + Ajouter un produit » était bleu sur fond bleu (invisible) et les liens de la sidebar étaient bleus au lieu de gris clair. Corrigée en `.mx-main a:not(.mx-btn)` — les liens du contenu restent bleus, les boutons et la sidebar retrouvent leurs couleurs. Vérifié par style calculé + capture.

## Vérification (locale, avant livraison)

- **Go build + vet + gofmt OK.**
- **Captures navigateur** (Chromium, 1440 px) : liste (mini-stats, filtre, tableau, bouton visible, sidebar claire), fiche (carte aperçu + formulaire, boutons), dashboard (intact, aucune régression).
- **Filtre catégorie testé** : `?cat=imprimantes` → 50 lignes, toutes « Imprimantes ».
- **Aperçu live testé** : 20/20 en jsdom (nom, marque, prix, dispo, catégorie, stock, image).
- **Données de test** : base locale 664 produits (0 commande, 0 compte, 2 ruptures) — en prod : 663 produits, 2 comptes.

## Déploiement (agent sur place)

1. Extraire le tar à la racine du repo (`tar -xzf …`).
2. `go build ./... && go vet ./internal/handlers/ ./internal/render/` (doit passer).
3. `docker compose up -d --build --force-recreate` (le compose a un `build: .`).
4. Vérifier `/admin/produits` (mini-stats, filtre, tableau) et `/admin/produits/15-pro` (carte aperçu + formulaire, saisir un prix pour voir la carte se mettre à jour).

## Limites connues

- Les vignettes de la liste et la photo de l'aperçu ne s'affichent que si `images/` est présent (monté en volume, comme en prod).
- 515 produits « Sans prix » (78 %) : le compteur est mis en évidence, mais les prix restent à saisir côté MEDIATEC.
- La colonne Stock reste « — » pour le lot (pas de stock) — normal.
