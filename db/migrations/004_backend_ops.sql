-- Backend fiabilité (T0+T1) : jobs, audit admin, aperçus temporaires, idempotence.
-- Jamais destructif : CREATE TABLE / INDEX IF NOT EXISTS, rejouable sans risque.
CREATE TABLE IF NOT EXISTS ops_jobs(
  nom TEXT PRIMARY KEY,
  derniere_ok TEXT NOT NULL DEFAULT '',
  detail TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS admin_audit(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  admin TEXT NOT NULL,
  action TEXT NOT NULL,
  objet TEXT NOT NULL DEFAULT '',
  ip TEXT NOT NULL DEFAULT '',
  date TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_admin_audit_date ON admin_audit(date);
CREATE TABLE IF NOT EXISTS design_previews(
  token TEXT PRIMARY KEY,
  payload TEXT NOT NULL DEFAULT '',
  expire_le TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS idempotency_keys(
  cle TEXT PRIMARY KEY,
  order_id INTEGER NOT NULL,
  cree_le TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
