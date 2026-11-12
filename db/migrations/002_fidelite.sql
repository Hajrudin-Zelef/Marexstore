-- Fidélité v3.25 : colonnes dérivées du générateur d'origine.
-- (accroche, description, specs_json, search_index existent déjà en 001.)
ALTER TABLE products ADD COLUMN prix_boutique INTEGER NULL;
ALTER TABLE products ADD COLUMN prix_avant INTEGER NULL;
ALTER TABLE products ADD COLUMN pointsforts_json TEXT NOT NULL DEFAULT '[]';
ALTER TABLE products ADD COLUMN ref TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN ref_fabricant TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN main TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN main_slug TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN sub TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN sub_slug TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN live_top TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN live_sub TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN facets_json TEXT NOT NULL DEFAULT '{}';
ALTER TABLE products ADD COLUMN consommables_json TEXT NOT NULL DEFAULT '[]';
ALTER TABLE products ADD COLUMN store_url TEXT NOT NULL DEFAULT '';
