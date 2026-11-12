-- Marexstore V1 : SQLite (WAL), migrable Postgres tel quel.
PRAGMA journal_mode=WAL;

CREATE TABLE IF NOT EXISTS products(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT UNIQUE NOT NULL,
  nom TEXT NOT NULL,
  marque TEXT NOT NULL DEFAULT '',
  prix INTEGER NULL,
  devise TEXT NOT NULL DEFAULT 'F CFA',
  dispo TEXT NOT NULL DEFAULT 'en stock',
  categorie TEXT NOT NULL DEFAULT '',
  sous_categorie TEXT NOT NULL DEFAULT '',
  image TEXT NOT NULL DEFAULT '',
  accroche TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  specs_json TEXT NOT NULL DEFAULT '[]',
  search_index TEXT NOT NULL DEFAULT '',
  stock INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);
CREATE INDEX IF NOT EXISTS idx_products_categorie ON products(categorie);

CREATE VIRTUAL TABLE IF NOT EXISTS products_fts USING fts5(
  slug, nom, marque, categorie, sous_categorie, accroche, description, search_index
);

CREATE TABLE IF NOT EXISTS users(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  pass_hash TEXT NOT NULL,
  nom TEXT NOT NULL DEFAULT '',
  tel TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions(
  token TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL,
  ip TEXT NOT NULL DEFAULT '',
  ua TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS orders(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NULL REFERENCES users(id) ON DELETE SET NULL,
  statut TEXT NOT NULL DEFAULT 'devis',
  moyen_paiement TEXT NOT NULL DEFAULT 'whatsapp',
  total INTEGER NULL,
  client_json TEXT NOT NULL DEFAULT '{}',
  preuve TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at);

CREATE TABLE IF NOT EXISTS order_items(
  order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id INTEGER NOT NULL REFERENCES products(id),
  qty INTEGER NOT NULL,
  prix_unitaire INTEGER NULL,
  PRIMARY KEY(order_id, product_id)
);
