# Hub des types + filtre global — Marexstore (2026-11-13)

## Ce qui a été livré

**1. Hub des types sur chaque page catégorie** — quand on entre dans une
catégorie, on arrive d'abord sur une page qui liste chaque type
d'équipement de cette catégorie, avec sa photo et son nombre de produits.
Un clic sur un type affiche ses produits. Même méthode dans les 7
catégories, même rendu que le hub des catégories de la boutique (cartes
photo + compteur, vert du logo, bouton « Voir tous les produits »).

**2. Filtre global « Trouver un produit » sur l'accueil** — une section
visible dès l'accueil : Marque, Catégorie, Type (la liste des types
dépend de la catégorie choisie), Prix min / Prix max, Disponibilité.
Le bouton « Rechercher » ouvre la boutique déjà filtrée.

Le but : trouver le produit précisément, sur tous les aspects, sans
passer par la liste complète.

## Les types (dérivés du catalogue réel, 579 produits, 37 types)

Aucun type inventé : ce sont les sous-catégories du catalogue, chaque
type n'apparaît que s'il a au moins un produit.

- **Réseaux (221)** : Réseau & câbles 139, Cisco & réseau 71,
  Téléphonie IP 8, Réseau Legrand 2, Adaptateurs 1
- **Ordinateurs (78)** : Portables 43, Tout-en-Un 21, Tablettes 13,
  Clients légers 1
- **Imprimantes & Consommables (145)** : Toners 62, Imprimantes 52,
  Cartouches d'encre 16, Scanners 7, Têtes d'impression 3,
  Kits de maintenance 2, Consommables 2, Rubans 1
- **Stockage (47)** : Disques & stockage 18, NAS 15, Cartouches LTO 6,
  Mémoire RAM 5, Clés USB 2, Cartes mémoire 1
- **Serveurs (16)** : Serveurs rack 16
- **Accessoires (58)** : Énergie & protection 13, Câbles HDMI 11,
  Kits clavier / souris 9, Audio-conférence 5, Casques & audio 5,
  Câbles RJ45 4, Pièces de rechange 3, Souris 3, Sacs & housses 2,
  Accessoires scanners 1, Enceintes 1, GPS 1
- **Moniteur & Écrans (14)** : Écrans 14

Note sur « routeur / switch / firewall » : ce découpage plus fin n'est
pas fiable à partir des noms de produits — 195 des 225 produits réseau
ne nomment pas leur type dans leur nom (ex. « Cisco Meraki MX105 » est
un firewall, « Cisco Aironet 1262N » un point d'accès, le nom ne le dit
pas). La précision vient donc de la combinaison : type + marque + prix +
disponibilité + recherche. Si un jour les fiches portent un champ
« type d'équipement » vérifié, le hub pourra descendre d'un niveau sans
changer la structure.

## Comment ça marche

- **Hub des types** (`internal/render/render3.go`, `TypeHubHTML`) : une
  tuile par type, photo = le premier produit du type (tri stock d'abord,
  donc un produit avec prix et en stock), repli « Photo à venir » si le
  type n'a aucune photo (ex. Adaptateurs : 1 seul produit, sans photo).
  Clic → même page avec `?type=<slug>` (lien partageable, bouton retour
  du navigateur OK). La liste des produits n'apparaît que quand un type
  (ou un filtre) est actif — comme le hub de la boutique ; « Voir tous
  les produits (N) » affiche toute la catégorie ; la pastille « Toutes »
  ramène au hub.
- **Filtre global** (`homeFilterHTML` + `window.MEDIATEC_TYPES`) :
  le formulaire envoie vers `boutique.html?cat=&type=&marque=&prixmin=
  &prixmax=&dispo=`. La liste des types se met à jour quand on change
  de catégorie (sans JavaScript, tous les types sont listés).
- **Paramètres URL** (`static/app.js`, présélections) : `?type=`,
  `?marque=`, `?prixmin=`, `?prixmax=`, `?dispo=` sont lus au chargement
  et les contrôles latéraux affichent les valeurs appliquées (marque,
  prix, disponibilité, case « en stock », pastille du type).
- **Unification `data-sub`** : les cartes de la boutique portaient
  `data-sub` = le slug de la catégorie principale ; elles portent
  maintenant le slug de la sous-catégorie, comme les pages catégories —
  sinon `?type=` ne filtrait rien sur la boutique. Le tri stock d'abord
  et les suggestions de recherche ne sont pas touchés.
- **Versions** : `style.css` et `app.js` passent en v329 (les deux
  endroits pour `app.js` : `render2.go` et `templates/static/_foot.html`).

## Application

Extraire à la racine du dépôt, puis reconstruire et redémarrer (le code
est dans le binaire et les fichiers statiques — pas d'import de données,
le catalogue en base ne change pas) :

```sh
tar -xzf marexstore-types-filtre-2026-11-13.tar.gz
docker compose up -d --build
```

Vérifications attendues en prod : page Réseaux → 5 tuiles (139 / 71 / 8 /
2 / 1) ; clic sur « Réseau & câbles » → 139 produits, pastille active ;
accueil → section « Trouver un produit » ; `boutique.html?cat=reseaux&
type=cisco-reseau&marque=cisco&dispo=en-stock` → 69 produits.

## Vérifications faites par Leo (local, base de test 579 produits)

- **Go** : `go build` + `go vet` + `go test ./...` OK. Nouveaux tests
  (`internal/render/typehub_test.go`) : hub complet pour les 7
  catégories (37 types, compteurs, photos ou repli, liens `?type=`),
  filtre accueil (marques, 7 catégories, 37 types, carte des types),
  `data-sub` des cartes boutique = slug de sous-catégorie.
- **Parité v3.25** : les 5 tests passent contre l'archive v3.25
  (site-orig/ extrait de l'archive pour le contrôle). `normPage`
  retire les nouveautés (hub, bloc catalogue, filtre accueil, carte
  des types) — le reste de chaque page est identique à l'octet près.
- **Fonctionnel jsdom** (`catalogue-site/typehub-test.js`, 28 contrôles,
  TOUT OK) : hub visible / liste cachée à l'entrée d'une catégorie ;
  `?type=reseau-cables` → 139 produits, pastille active ;
  `?all=1` → 221 produits ; boutique `?cat=reseaux&type=cisco-reseau&
  marque=cisco&dispo=en-stock` → 69 produits, contrôles synchronisés ;
  boutique `?cat=ordinateurs&marque=hp&prixmin=100000&prixmax=1000000`
  → 6 produits, prix synchronisés ; accueil : 38 options de type au
  départ, 6 après choix de Réseaux, 5 après choix d'Ordinateurs.
- **Visuel** (captures Chrome, vraies photos) : hub Réseaux, filtre
  accueil, liste filtrée « Réseau & câbles » — rendu premium, conforme
  au site (vert #2E5136, cartes du hub, filtres latéraux).

## Fichiers modifiés

- `internal/render/render3.go` : `TypeHubHTML`, `CategoryPage`
  (hub + liste masquée), `homeFilterHTML`, `IndexBody` (filtre),
  `BoutiqueBody` (`data-sub` = sous-catégorie)
- `internal/render/render.go` : `brandOptions` extrait (partagé par le
  filtre latéral et le filtre accueil)
- `internal/render/render2.go` : versions `style.css` / `app.js` → v329
- `static/app.js` : présélections URL (`type`, `marque`, `prixmin`,
  `prixmax`, `dispo`), synchronisation des contrôles, type dépendant
  de la catégorie sur l'accueil
- `static/style.css` : styles du filtre global
- `templates/static/_foot.html` : `app.js` → v329
- `internal/render/parity_test.go` : `normPage` retire les nouveautés
- `internal/render/typehub_test.go` : nouveaux tests (hub, filtre, data-sub)
