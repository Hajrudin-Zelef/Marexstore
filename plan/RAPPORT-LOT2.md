# Lot 2 — 85 produits des pages détail (catalogue commercial)

**Date :** 2026-11-13 — **Source :** `catalogue-commercial-mediatec.pptx` (13/10/2026), pages détail 44–145 (404 cartes).
**Référence :** `liste-manquants-detail-2026-11-13.md` (89 cartes sans fiche propre → 85 produits distincts après fusions).

## Contenu

- **85 fiches** : `nouveaux-produits-commercial-lot2.json` (même format que le lot 1).
- **85 photos** : `images/<slug>.jpg` (extraites du PPTX, une par produit).
- **642 specs** (moy. 7,6/fiche), **Description** (2–4 phrases) + **5 Points forts** par fiche.
- **Prix null partout** (« Prix sur demande ») — le catalogue n'a pas de prix, aucun chiffre inventé.

## Répartition

| Catégorie | Fiches |
|---|---|
| Réseau & câbles | 34 |
| Cisco & réseau | 23 |
| Stockage / NAS | 8 |
| Énergie & protection | 5 |
| Stockage | 4 |
| Serveurs | 3 |
| Téléphonie IP | 3 |
| Toners & consommables | 2 |
| Imprimantes | 1 |
| Cartouches d'encre | 1 |
| Tablettes (iPhone 16 Pro) | 1 |

## Vérification (locale, avant livraison)

- **Import testé** sur une copie de la base : 171 (stock) + 408 (lot 1) + 85 (lot 2) = **664 produits**, **664 slugs distincts** (aucun doublon).
- **Ré-import du lot 2 idempotent** (toujours 664, pas de doublon).
- **Nouveaux totaux par top** : Matériels réseaux 281 (+60), Accessoires Informatiques 211 (+20), Serveur 19 (+3), Téléphones & Tablettes 14 (+1, iPhone 16 Pro), Imprimantes 60 (+1).
- **DGS-1510-28X non recréé** : il existe déjà dans le lot 1. Seul le DGS-1520-28 est ajouté (c'est comme ça qu'on tombe à 85).

## Déploiement (agent sur place)

1. Copier `nouveaux-produits-commercial-lot2.json` dans `catalogdata/` (à côté du lot 1).
2. Copier les 85 images dans `images/` (dossier gitignoré, déjà en prod).
3. Importer : `go run ./import --json catalogdata/nouveaux-produits-commercial-lot2.json --db /data/store.db`
4. **Supprimer la fiche tronquée remplacée** (sinon doublon) :
   `sqlite3 /data/store.db "DELETE FROM products WHERE slug='firepower-ngfw';"`
   (la fiche « Firepower -NGFW) » du lot 1 est remplacée par la fiche correcte « Firepower 2140 » du lot 2).
5. Redémarrer le conteneur (les produits sont chargés en mémoire au démarrage).
6. Vérifier : 664 produits au total.

## Erreurs du catalogue corrigées / signalées

Ces noms viennent du catalogue lui-même ; les fiches sont enrichies correctement mais le **nom n'a pas été modifié** (règle : nom = catalogue). À renommer sur décision de Zelef :

- **Synology HAT5320** : l'accroche catalogue « NAS Haute capacité » est **fausse** — c'est un **disque dur d'entreprise 3,5" SATA** (4 à 24 To), pas un NAS. La fiche le décrit correctement comme un disque dur (catégorie corrigée en « Stockage »).
- **Huawei AP663** : **n'existe pas** sous ce nom. La série eKit AP600 a l'AP661 (Wi-Fi 6 tri-bande) et l'AP673 ; ou AP6630DN (Wi-Fi 5, 2019). Probable faute de frappe.
- **« Easy UPS 3S 120 kVA »** : **n'existe pas** — la gamme 3S s'arrête à 40 kVA ; le 120 kVA est le **3M** (E3MUPS120KHS). La fiche est enrichie avec les specs du 3M 120 kVA.
- **DGS-1210-24** : le modèle de base n'a **pas de PoE** ; l'accroche « PoE+ » pointe vers la variante **24P** — la fiche est enrichie comme 24P (budget 193 W).
- **S310-16T** : pas de modèle 16 ports dans la série S310 (que des 24/48 ports) ; enrichi au niveau série uniquement.
- **ISA3000-8C** : cette référence n'existe pas chez Cisco (seuls ISA-3000-4C et ISA-3000-2C2F). Enrichi avec les specs de la famille ISA3000.
- **« 6500E » / « USG6700E »** : noms de série génériques, pas de modèle précis. Enrichis au niveau série.
- **Catalyst 9300 24P** : l'accroche dit « multi-gig », mais les ports du 24P sont du 1G PoE+ (le multi-gig, c'est la variante 24U/48U) — pas de ports multi-gig revendiqués.
- **Catalyst 48T** : interprété comme **C9300-48T** (indice catalogue : « accès, stackables ») — à confirmer.

## À savoir (fin de vie)

- **Webex Desk Pro** : fin de commercialisation annoncée le 2026-02-16 (fin de vente 2026-03-16, support jusqu'au 2031-03-31).
- **FortiGate 100F** : un revendeur l'indique « End of Sale » — à confirmer avant de le mettre en avant.

## Limites connues (non corrigées ici)

- **« Caméras & Vidéosurveillances » reste à 0 produit.** Les caméras (lot 1 + lot 2) sont classées « Réseau & câbles » → elles apparaissent sous « Matériels réseaux » (281). Remplir le top « Caméras & Vidéosurveillances » exige une modification de `LiveMap` (catalogue Go) pour router les caméras vers ce top — à faire sur demande, car ça déplace aussi les caméras du lot 1.
- **« Accessoires Téléphones & Tablettes » reste à 0** (aucun accessoire téléphone dans le lot 2 ; l'iPhone 16 Pro va dans « Téléphones & Tablettes »).
- **« Uncategorized » reste à 0** (normal : tous les produits sont catégorisés).
- **431 + 85 = 516 produits « Prix sur demande »** sur 664 — le catalogue n'a pas de prix ; la tarification reste à faire côté MEDIATEC.
- **Fiches existantes au nom tronqué** (héritées du catalogue, hors scope sauf Firepower 2140 ci-dessus) : « ISR 1100 (C1111-4P », « CX 6000 48G) », « ES241(P » — à renommer un jour.
