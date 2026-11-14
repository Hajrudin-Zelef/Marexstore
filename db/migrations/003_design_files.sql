-- Studio Design (T3, socle en T0) : fichiers éditables, versions append-only.
-- Jamais destructif : CREATE TABLE / INDEX IF NOT EXISTS, rejouable sans risque.
-- Une seule version active par nom (actif=1) ; l'historique est conservé
-- (jamais UPDATE/DELETE du contenu, rollback = réactiver une ancienne version).
CREATE TABLE IF NOT EXISTS design_files(
  nom TEXT NOT NULL,
  contenu TEXT NOT NULL DEFAULT '',
  version INTEGER NOT NULL DEFAULT 1,
  auteur TEXT NOT NULL DEFAULT '',
  cree_le TEXT NOT NULL DEFAULT (datetime('now','localtime')),
  actif INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(nom, version)
);
CREATE INDEX IF NOT EXISTS idx_design_files_actif ON design_files(nom, actif);
