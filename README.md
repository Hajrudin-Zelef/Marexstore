# Marexstore — Go + html/template + htmx + SQLite → Postgres

E-commerce Mediatec : 171 produits importés de `catalogue-mediatec-v3.25.tar.gz`.

## Démarrage local

```bash
go run ./import --json /tmp/opencode/ref/sources/produits-v3.json --db store.db --sql db/migrations/001_init.sql
# images de démo (archive) :
mkdir -p images && tar -xzf catalogue-mediatec-v3.25.tar.gz -C /tmp/ref && cp -r /tmp/ref/images ./images
PORT=8080 DB_PATH=./store.db APP_ENV=dev go run ./cmd/store
# http://localhost:8080/  /boutique  /produits/{slug}  /commande  /compte  /admin
```

## Variables (.env, voir .env.example)

`PORT DB_PATH SESSION_SECRET BASE_URL WHATSAPP_NUMBER ADMIN_USER ADMIN_PASS_HASH CSP_REPORT_ONLY APP_ENV`

## Sécu / perf / responsive

- Headers : CSP Report-Only → enforce, HSTS, DENY frame, nosniff, Referrer strict, Permissions-Policy.
- CSRF sur POST, sessions DB SHA256 + bcrypt12, cookies HttpOnly/Secure/Lax.
- Rate-limit doux 429+Retry-After, validation slug/prix/qty serveur, SQL préparé.
- Cache immutable `/static`, gzip nginx, pagination 24, FTS, `srcset/lazy`, mobile-first 44px.

## Docker

```bash
docker compose build && docker compose up -d
# nginx : voir nginx/store.neva-ci.pro.conf + certbot
```

## Ne jamais commiter

`.env`, `*.db`, `images/` locales volumineuses (servies via volume en prod), tokens GitHub.
