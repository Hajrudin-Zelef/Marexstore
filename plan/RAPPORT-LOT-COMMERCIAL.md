# Lot commercial — catalogue commercial -> site Marexstore

Date : 2026-11-13. Source : catalogue commercial PPTX du 13/10/2026
(145 diapos) + sources de construction (`families.json`, `products.json`).

## Contenu du lot

| | |
|---|---|
| Fiches nouvelles | **408** (104 familles + 304 modeles) |
| Doublons avec les 171 du site | **0** (noms exacts compares, 2 faux positifs ecartes : les familles generiques "Lenovo ThinkPad" / "Epson EcoTank" ne sont pas les produits specifiques du site) |
| Prix | **tous null** — le catalogue ne contient AUCUN prix (verifie : 0 occurrence "F CFA" / "FCFA" / "€" dans le PPTX). Les 408 fiches affichent donc "Prix sur demande" + "Demander un devis" (boutons WhatsApp presents comme pour les autres). |
| Images | **408 / 408** copiees (familles : `images/<id>.jpg` ; modeles : `pimg_norm/<id>_1.jpg`), renommees `images/<slug>.jpg`. 0 collision de nom avec les photos existantes. |
| Slug | uniques globalement (site + lot). 1 collision resolue : les 2 "HP 305XL" (encre et toner) -> `hp-305xl` et `hp-305xl-toners`. |

## Repartition par categorie (cles CAT_MAP du site, aucune nouvelle categorie)

Reseau & cables 136, Cisco & reseau 60, Imprimantes 47, Ordinateurs
portables 29, Ordinateurs de bureau 20, Serveurs 15, Stockage / NAS 14,
Stockage 13, Tablettes 12, Cartouches d'encre 12, Toners & consommables 16,
Energie & protection 10, Telephonie IP 7, Ecrans 6, Scanners 5,
Audio-conference 4, Kits clavier / souris 1, Consommables impression 1.

Regles de mapping : Cisco -> "Cisco & reseau", autre reseau -> "Reseau &
cables" ; copieurs/traceurs -> "Imprimantes" ; smartphones -> "Tablettes" ;
onduleurs -> "Energie & protection" ; cameras/controle d'acces -> "Reseau &
cables" (pas de categorie videosurveillance dans CAT_MAP — les compteurs du
mega-menu et de la boutique restent justes, ils sont calcules depuis la base).

## Marques

Derivees du nom (mot entier, jamais inventees). 31 fiches restent "Sans
marque" : ce sont les entrees generiques ("Disques & SSD", "Papier
thermique", "Points d'acces Wi-Fi 6/6E", "Baies & coffrets rack"...) ou
multi-marques ("Tout-en-un ASUS / Dell", "NAS QNAP / Synology", "Cameras
Hikvision / Dahua", "Smartphones Samsung / Apple"...) — on ne choisit pas
une marque quand le catalogue en cite deux. 2 noms tronques dans le
catalogue corriges par le contexte de leur famille : "Z Fold 6" -> Samsung
(Galaxy Z Fold 6), "15 Pro" -> Apple (iPhone 15 Pro).

## Disponibilite

Toutes les fiches sont en "en stock" : ces produits sont commandables /
devisables, disponibilite et prix a confirmer avant commande (comme
l'indique deja la boutique). Aucun produit n'est marque en rupture.

## Application (cote agent / VM)

```
# 1. Extraire a la racine du depot Marexstore
tar -xzf marexstore-lot-commercial-2026-11-13.tar.gz
# 2. Copier les photos dans le volume images
cp images/*.jpg ./images/
# 3. Importer le lot dans la base prod (les 171 ne sont pas touches :
#    l'import fait un upsert par slug, les slugs sont tous nouveaux)
go run ./import --json catalogdata/nouveaux-produits-commercial.json --db ./data/store.db
#    (la recherche FTS est reconstruite par l'import lui-meme)
# 4. Reconstruire + redemarrer (ce depot a une section build:)
docker compose up -d --build
```

## Verifications faites par Leo (base de test, 2026-11-13)

- Import 171 puis lot : **579 produits, FTS 579 lignes** ; les 171
  d'origine intacts (prix et prix_boutique conserves, ex. toner 650A jaune).
- Taxonomie derivee verifiee : `macbook-pro-m5` -> Ordinateurs / Portables /
  laptots/apple ; `catalyst-9200` -> Reseaux / Cisco & reseau /
  materiels-reseaux ; `hp-290-g9-tours-de-bureau` -> Ordinateurs / Tout-en-Un.
- Serveur de test (579 produits) : `/` 200, `/boutique.html` 200 ("Voir tous
  les produits (579)", compteurs du hub mis a jour automatiquement),
  `/categories/reseaux.html` 200 (221 produits), `/categories/ordinateurs.html`
  200 (contient le MacBook Pro M5), `/produits/macbook-pro-m5.html` 200
  ("Prix sur demande" + "Demander un devis" + "Commander sur WhatsApp"),
  `/images/macbook-pro-m5.jpg` 200, `/api/recherche?q=macbook` et
  `?q=catalyst` renvoient les nouvelles fiches, `?q=toner` toujours 10.
- Script de construction (reproductible) :
  `~/workspace/catalogue-site/lot-commercial/build_lot.py`.

## Apres deploiement, verifier en prod

- `https://store.neva-ci.pro/boutique.html` affiche 579 produits.
- `https://store.neva-ci.pro/api/recherche?q=macbook` renvoie les MacBook.
- `/admin` toujours 401 (pas de hash configure, normal).
