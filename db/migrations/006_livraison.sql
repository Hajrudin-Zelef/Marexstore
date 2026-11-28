-- Module livraison : zones + frais (montants saisis par le patron en admin).
-- Types fermés (côté applicatif) : retrait|livraison|expedition.
-- Jamais destructif (IF NOT EXISTS + seed gardé), rejouable sans risque.
CREATE TABLE IF NOT EXISTS livraison_zones(id INTEGER PRIMARY KEY AUTOINCREMENT, nom TEXT NOT NULL, type TEXT NOT NULL DEFAULT 'livraison', frais INTEGER NOT NULL DEFAULT 0, seuil_franco INTEGER NOT NULL DEFAULT 0, delai TEXT NOT NULL DEFAULT '', ordre INTEGER NOT NULL DEFAULT 0, actif INTEGER NOT NULL DEFAULT 1);
INSERT INTO livraison_zones(nom, type, frais, seuil_franco, delai, ordre, actif)
SELECT 'Retrait boutique — Plateau', 'retrait', 0, 0, 'Immédiat', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM livraison_zones);
INSERT INTO livraison_zones(nom, type, frais, seuil_franco, delai, ordre, actif)
SELECT 'Livraison Abidjan', 'livraison', 0, 100000, '24–72h', 2, 1
WHERE NOT EXISTS (SELECT 1 FROM livraison_zones WHERE ordre = 2);
INSERT INTO livraison_zones(nom, type, frais, seuil_franco, delai, ordre, actif)
SELECT 'Expédition intérieur', 'expedition', 0, 0, 'À confirmer', 3, 1
WHERE NOT EXISTS (SELECT 1 FROM livraison_zones WHERE ordre = 3);
