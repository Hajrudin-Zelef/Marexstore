# RAPPORT — Phase 3.0 Studio : socle (design actif + parité)

Base : `3bad320` (phase 2 déployée). Date : 2026-11-14.

## Ce que fait cette livraison

Le socle de la phase 3 : le serveur charge le design publié et le garde en
mémoire pour le rendu. **Aucun changement visible** — c'est le filet de
sécurité des sous-phases suivantes (3.1 à 3.6).

## Fichiers

- `internal/design/active.go` (nouveau) : le design actif.
  - `LoadActive(db)` : charge le design actif depuis `design_files` (version
    active), le parse une seule fois, le stocke en global protégé par un
    RWMutex (lu à chaque requête, écrit sur publication).
  - `Refresh(db)` : recharge l'actif (après publication / activation).
  - `Current()` : le design actif (nil si rien n'est publié).
  - Accesseurs défensifs, récepteur nil accepté : `Text(key, def)` (textes
    `Tx`), `Bool(section, key, def)`, `Str(section, key, def)`,
    `Int(section, key, def)`, `List(section, key)` (listes d'objets, ex.
    `header.navLinks`), `Has(section)`, `Version()`. Une clé absente ou
    d'un mauvais type retourne le défaut — un design partiel ne plante
    jamais le rendu.
- `cmd/store/main.go` : `design.LoadActive(database)` après le chargement du
  thème. Best-effort : en cas d'erreur on journalise et on continue sans
  design (jamais fatal).
- `internal/handlers/admin_design.go` : après une publication
  (`AdminDesignAPI` POST) et après une activation (`AdminDesignActivate`),
  `design.Refresh(a.DB)` recharge l'actif pour le rendu. Un échec de
  rechargement ne bloque pas la publication (le thème est déjà en ligne).
- `internal/design/active_test.go` (nouveau, 5 tests) : sans design (actif
  nil, accesseurs = défauts), design partiel (valeurs présentes + défauts
  ailleurs), types incorrects (défauts, jamais de panique), Refresh vers une
  nouvelle version, récepteur nil.
- `internal/render/parity_design_test.go` (nouveau) : **parité sans design**.
  Accueil, boutique, fiche, catégorie rendus avec le thème par défaut et
  sans design, normalisés comme la parité v3.25, comparés par SHA256 à des
  empreintes enregistrées. Toute sous-phase qui modifierait le rendu sans
  design fera échouer ce test.

## Vérification (locale, base de test 664 produits)

- `gofmt` propre, `go build ./...`, `go vet ./...`, `go test ./...` verts
  (design : 5 nouveaux tests ; render : parité sans design, empreintes
  stables sur 2 passages).
- Serveur live (port 18099) : démarrage sans design → `/`, `/boutique.html`,
  `/produits/15-pro.html`, `/categories/materiels-reseaux.html`, `/admin`,
  `/admin/systeme` → 200.
- API : POST design v1 → design v1 / thème v1 ; POST v2 → design v2 / thème
  v2 ; activer v1 → design v1 / thème v3 ; GET design → v1. Pages toujours
  200, 0 erreur dans les logs.
- Redémarrage avec un design déjà en base (v1 actif) : le serveur démarre,
  pages 200, le bandeau v1 (issu du design) est présent sur l'accueil —
  `LoadActive` au démarrage fonctionne. 0 erreur dans les logs.

## Déploiement

Extraire l'archive à la racine du dépôt, `chmod -R a+rX internal cmd`,
`go build ./... && go vet ./... && go test ./...`, puis
`docker compose up -d --build --force-recreate`. Aucun changement de
Dockerfile, aucune migration, aucune route nouvelle. Sans design publié en
prod, le site est strictement identique à aujourd'hui.

## Prochaine étape (3.1)

Textes + visibilité : les libellés `Tx` remplacent les chaînes en dur dans
le rendu, les bascules de visibilité masquent/affichent les blocs. Attend
le feu vert de Zelef.
