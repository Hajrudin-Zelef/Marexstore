# Dashboard admin premium — Marexstore

**Date :** 2026-11-13 — **Base :** origin/main 933ba2d (coquille WordPress de l'admin).
**Objectif :** transformer le tableau de bord (5 chiffres + 2 tableaux, plat) en dashboard digne de ce nom : donut, graphiques, cartes premium.

## Contenu

| Fichier | Changement |
|---|---|
| `internal/handlers/admin_dash.go` | **Nouveau** : helpers dashboard (donut SVG, barres, cartes KPI premium, icônes SVG) |
| `internal/handlers/admin.go` | `Admin()` réécrite : requêtes d'agrégation + mise en page premium |
| `internal/render/admin_chrome.go` | `admin.css?v=1` → `?v=2` (cache) |
| `static/admin.css` | Styles dashboard premium (KPI, donut, légende, barres, état vide) |

## Ce qui change (visuel)

- **5 cartes KPI premium** : icône (SVG trait), bordure d'accent couleur (vert CA, bleu Commandes, vert Produits, rouge Ruptures, violet Comptes), gros chiffre, sous-ligne avec lien, léger soulèvement au survol.
- **Donut « Produits par catégorie »** (SVG généré côté serveur, zéro JS, zéro CDN) : 8 segments, total au centre (663), légende avec compteur + pourcentage. Les 3 catégories à 0 produit sont exclues du donut et signalées en note.
- **Donut « Commandes par statut »** (SVG) : état vide élégant quand 0 commande (« Aucune commande pour le moment ») — il se remplira tout seul quand les commandes arriveront.
- **Barres « Top marques »** : top 8 marques (Cisco 124, HP 113, Sans marque 63, Huawei 54, Canon 24, Apple 22, Dell 19, TP-Link 18), barres horizontales colorées avec compteurs.
- **Tableaux conservés** : Dernières commandes + À réassortir, inchangés, en bas du dashboard.

## Technique

- **Graphiques en SVG inline générés en Go** (technique stroke-dasharray) : pas de librairie JS, pas de CDN, pas de dépendance, marche hors-ligne, rapide.
- **Données** : que des comptages SQL sur l'existant (`GROUP BY live_top`, `GROUP BY statut`, `GROUP BY marque`) — rien de nouveau côté base.
- **Palette** : vert marque #2E5136 en tête, puis bleu, rouge, ambre, violet, turquoise… (12 couleurs, cyclées).
- **CSS** : classes `mx-` uniquement (zéro impact boutique). `color-mix` utilisé pour les pastilles d'icônes, avec repli `background:#eef0f2` pour les vieux navigateurs.

## Vérification (locale, avant livraison)

- **Go build + vet OK.**
- **Rendu réel vérifié par capture navigateur** (Chromium, 1440 px + page complète) : KPI avec icônes et bordures couleur, donut 8 segments + légende lisible, état vide commandes, barres top 8, tableaux en bas. Aucun débordement, aucun chevauchement.
- **Données de test** : base locale 664 produits (0 commande, 0 compte, 2 ruptures) — les chiffres en prod seront 663 produits / 2 comptes (le test n'avait pas les comptes), le rendu est identique.

## Déploiement (agent sur place)

1. Extraire le tar à la racine du repo (`tar -xzf …`).
2. `go build ./... && go vet ./internal/handlers/ ./internal/render/` (doit passer).
3. `docker compose up -d --build --force-recreate` (le compose a un `build: .`, il reconstruit l'image).
4. Vérifier `/admin` : les 5 cartes, le donut, les barres, les tableaux.

## Limites connues

- Le donut « Commandes par statut » est vide tant qu'il n'y a pas de commande — normal, il se remplira.
- Les 3 catégories à 0 produit ne figurent pas dans le donut (note affichée sous le graphique).
- Le CA reste « — » tant qu'aucune commande n'est payée/préparée/livrée.
