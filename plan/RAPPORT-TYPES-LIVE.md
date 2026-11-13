# « Voir tous » du méga-menu : la page des types d'abord — toutes les catégories

Date : 2026-11-13. Base : origin/main cbd413e (app.js v331, style v329).
Fichiers : 5 modifiés, 122 insertions / 5 suppressions. Aucune donnée à importer.

## Le problème (décrit par Zelef)

Dans le méga-menu « Toutes les catégories », on clique sur une catégorie
(ex. Matériels réseaux) puis sur « Voir tous les Matériels réseaux ».
Normalement les types de matériels devraient s'afficher AVANT les produits.
À la place, la page des types apparaît et disparaît d'un coup, et les
produits s'affichent directement.

## La cause

« Voir tous les X » ne menait pas à la page des types. Il menait à la
boutique filtrée (`boutique.html?store=<slug>`). Sur la boutique, le filtre
est actif dès l'arrivée, donc le hub s'affiche puis disparaît (le JS le
cache quand un filtre est actif) et les produits s'affichent directement.

Pire : les pages catégories — celles qui portent le hub des types — n'avaient
AUCUN lien nulle part sur le site (0 lien vers /categories/ dans le rendu).
Le hub des types était donc inatteignable en cliquant, pour toutes les
catégories, pas seulement pour les réseaux.

## Le correctif

« Voir tous les X » mène maintenant à la page catégorie de X
(`/categories/<slug>.html`), qui affiche d'abord le hub des types, puis les
produits quand on choisit un type (ou « Voir tous les produits »).

Le menu utilise les catégories du store en ligne (tops live, ex.
« Matériels réseaux » = materiels-reseaux) alors que les pages catégories
utilisaient les grandes familles (ex. « Réseaux » = reseaux). La page
catégorie accepte donc aussi un top live : elle affiche les types des
produits DE CE top live.

Détail des changements :
- `internal/render/render.go` (Méga-menu) : le lien « Voir tous les X »
  passe de `boutique.html?store=<slug>` à `categories/<slug>.html`.
  Les liens des sous-catégories (un type précis) ne changent pas : un type
  choisi affiche directement ses produits, c'est le comportement voulu.
- `internal/render/render3.go` : la page catégorie est factorisée
  (`categoryPageFor`) ; nouvelle `CategoryPageLive` + `CategoryBodyLive`
  (produits d'un top live, tri stock d'abord comme partout).
- `internal/handlers/pages.go` (`Categorie`) : `/categories/<fichier>.html`
  résout d'abord les familles, puis les tops live, sinon 404.
- `internal/render/parity_test.go` : la parité normalise le href
  « Voir tous » (changement voulu, vérifié par son propre test).
- `internal/render/typehub_test.go` : 2 nouveaux tests
  (`TestLiveTopPages`, `TestMegaMenuVoirTous`).

Aucun changement de app.js / style.css (le hub et syncHub fonctionnaient
déjà) : pas de bump de version, pas de vidage de cache nécessaire côté JS.
Les pages sont rendues par le serveur.

## Vérification (locale, base 579 produits)

- Go : build + vet OK, gofmt propre.
- Tests Go : 12/12, dont parité 5/5 (avec l'archive v3.25).
- jsdom (serveur local) : 11/11 sur la page live
  (`/categories/materiels-reseaux.html` : hub visible + liste cachée sans
  paramètre ; 5 tuiles ; `?type=reseau-cables` → 139 produits ;
  `?all=1` → 221 produits ; page Imprimantes OK).
  Suite jsdom existante : TOUT OK.
- Les 11 tops live répondent 200, dont Matériels réseaux (221 produits,
  5 types : Réseau & câbles 139, Cisco & réseau 71, Téléphonie IP 8,
  Réseau Legrand 2, Adaptateurs 1).

## À noter

- 3 tops live n'ont AUCUN produit dans le catalogue actuel :
  Accessoires Téléphones & Tablettes, Caméras & Vidéosurveillances,
  Uncategorized (0 produit). Leur page catégorie est vide, c'est normal
  (aucun produit dans le catalogue pour ces catégories). Le menu les
  affiche quand même, comme avant.
- Les cartes de la boutique (HubHTML) pointent toujours vers la boutique
  filtrée (`boutique.html?store=`) : ce n'était pas dans la demande
  (le méga-menu uniquement). Si tu veux le même comportement là aussi
  (la carte mène à la page des types), dis-le-moi.

## Application

Extraire à la racine du dépôt, puis reconstruire et redémarrer
(le code est dans le binaire ; aucune importation de données) :

    docker compose up -d --build

Vérifications attendues en prod :
- « Toutes les catégories » → Matériels réseaux → « Voir tous les
  Matériels réseaux » → la page affiche les 5 tuiles de types, pas les
  produits directement.
- Cliquer « Réseau & câbles » → 139 produits.
- « Voir tous les produits (221) » → les 221 produits.
- Idem pour les 10 autres catégories du menu.
