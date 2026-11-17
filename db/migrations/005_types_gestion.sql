-- Gestion des catégories (Types d'équipements) : catégories déclarées,
-- même vides (0 produit) — sinon un type sans produit reste invisible car
-- les cartes sont calculées depuis products. Jamais destructif.
CREATE TABLE IF NOT EXISTS declared_types(
  top TEXT NOT NULL,
  sub TEXT NOT NULL,
  cree_le TEXT NOT NULL DEFAULT (datetime('now','localtime')),
  PRIMARY KEY(top, sub)
);
