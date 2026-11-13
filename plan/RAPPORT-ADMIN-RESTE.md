# Admin — reste des modules premium (sidebar + Types + Commandes + Clients + Médias + Réglages)

Livré le 2026-11-13. Base : origin/main 22ab75d (module Produits déployé).
Fichiers : `internal/handlers/admin.go`, `internal/handlers/admin_extra.go`,
`internal/render/admin_chrome.go`, `static/admin.css` (v5). Aucun changement de
données (mêmes champs, mêmes POST, mêmes CSRF, mêmes transitions).

## Sidebar (MxSidebar, admin_chrome.go)
- Libellé de section « Administration » au-dessus du menu (comme « Boutique »).
- Actif premium : fond vert #2e5136 translucide + bordure gauche citron #8dc63f
  + icône citron (au lieu du fond noir / bordure bleue).
- Pied remplacé par une carte utilisateur : avatar rond « A » (bordure citron),
  « admin » / « Administrateur », lien retour boutique (↗). L'ancien
  « Connecté : admin » disparaît (CSS `.mx-side-foot` supprimé).
- « Voir le site » gagne une icône lien externe.

## Types (AdminTypes)
- Mini-stats : Tops (8), Types (29), Produits (664), Plus gros top (281 · Matériels réseaux).
- Le tableau plat devient une **grille de cartes par top** : libellé français
  (topLabels), badge total, liste des types avec compteurs, chaque type = lien
  vers sa page boutique. Les sous-types vides s'affichent « Sans type »
  (lien vers la page du top) au lieu d'une ligne vide.
- Fusion restylée en formulaire premium (.mx-fsec, étiquettes MAJUSCULES,
  bouton Enregistrer vert) ; champs et JS inchangés.

## Commandes — liste (AdminCommandes)
- Mini-stats : Commandes (total), CA encaissé (somme des totaux paye/prépare/livre),
  En cours (ni livré ni annulé), Livrées.
- Total formaté « 620 000 F CFA » (fmtMoney), « devis » discret si null.
- Client et Statut séparés en deux colonnes ; le client affiche « Nom · tél »
  (nouveau helper `clientLabel`, parse le client_json) au lieu du JSON brut.
- Pager sorti du tableau (comme le module Produits).

## Commande — détail (AdminCommande)
- **Stepper visuel** de la machine à états (Devis → WhatsApp reçu → Attente MoMo
  → Payée → Préparée → Livrée) : étapes passées vertes, étape courante en vert
  plein, suivantes grisées. Si annulée : bandeau rouge « Commande annulée ».
- Lignes : colonne « Total ligne » ajoutée (qté × prix unit.), prix formatés,
  total des lignes en pied.
- Résumé : total en gros vert, moyen, client lisible (clientLabel), compte
  (lien), date de création. Statut + boutons de transition conservés.
- Historique (timeline) conservé.

## Clients — liste (AdminClients)
- Mini-stats : Clients, Avec commande, Sans commande, Commandes.
- Avatar rond avec initiale du nom (vert) devant l'e-mail ; compteur de
  commandes en badge (vert si > 0, gris si 0). Pager avec liens URL-échappés.

## Client — détail (AdminClient)
- Carte profil style fiche employé : bandeau vert, avatar rond (initiale),
  nom, e-mail en badge citron, infos (Téléphone, Inscription, Commandes,
  Total dépensé formaté). Tableau des commandes conservé (totaux formatés).

## Médias (AdminMedias)
- Mini-stats : Images, Utilisées, Orphelines, Taille totale (Mo).
- Recherche par nom de fichier (`?q=`, insensible à la casse, combinée à la vue).
- Grille : survol (légère élévation + ombre), bouton Supprimer compact.
  Vues Toutes/Orphelins conservées (liens avec `q` préservé).

## Réglages (AdminReglages)
- Bandeau de statut en tête : vert « Boutique en ligne » ou ambre
  « Mode maintenance actif ».
- **Interrupteur** à la place de la case à cocher (CSS .mx-switch).
- Deux sections (.mx-fsec) : Boutique (maintenance + message) et Contact
  (WhatsApp) ; étiquettes MAJUSCULES, champs fond clair, bouton Enregistrer vert.
  Champs et POST inchangés.

## Vérification locale
- `go build ./...`, `go vet ./...`, `gofmt` propres.
- Base de test 664 produits + 3 clients + 5 commandes (devis/payée/livrée/
  annulée/attente_momo) + 6 événements + 6 images (3 utilisées, 3 orphelines).
- Captures navigateur (1440 px) de chaque page : dashboard, types, commandes,
  commande #1, clients, client #1, médias (+ orphelins), réglages.
- 2 corrections après capture : « Sans type » (sous-types vides) et
  `clientLabel` (JSON brut → « Nom · tél »). Re-vérifiées par capture.
- Les images des médias apparaissent cassées en file:// (attendu : le test
  local ne sert pas /images/ ; en production elles se chargent).

## Déploiement
Extraire à la racine du dépôt, puis :
`go build ./... && go vet ./... && docker compose up -d --build --force-recreate`
