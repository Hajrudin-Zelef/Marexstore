package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// Miroir serveur : appelé par mirror.js en parallèle du flux localStorage.
// Ne change jamais l'UX : 204 même si la donnée est partielle.
func (a *App) checkCSRF(r *http.Request) bool {
	return auth.CheckCSRF(a.Cfg.SessionSecret, a.sessionID(r), r.FormValue("csrf"))
}

func (a *App) mirror(r *http.Request) bool {
	return r.Header.Get("X-Mirror") == "1" || r.Header.Get("HX-Request") == "true"
}

func (a *App) ok(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

// POST /api/panier {slug, qty} — trace d'ajout.
func (a *App) ApiPanier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !a.checkCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	slug := strings.TrimSpace(r.FormValue("slug"))
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	if qty < 1 {
		qty = 1
	}
	if qty > 10 {
		qty = 10
	}
	var id int64
	if err := a.DB.QueryRow(`SELECT id FROM products WHERE slug=?`, slug).Scan(&id); err != nil {
		http.Error(w, "introuvable", http.StatusNotFound)
		return
	}
	a.ok(w)
}

// POST /api/commande {items: "slugxqty,...", moyen, nom?, tel?}
// Enregistre la commande WhatsApp (le message wa.me reste construit côté client).
// T1 — backend dur : commande + lignes + événement de création + clé
// d'idempotence en UNE SEULE transaction (jamais de commande partielle).
// Idempotency-Key : rejouer la même requête renvoie la commande existante
// (jamais de doublon sur retry réseau). Total calculé côté serveur depuis
// la base (prix AFFICHÉ = COALESCE(prix_boutique, prix) : le client paie
// ce qu'il a vu).
func (a *App) ApiCommande(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !a.checkCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	// Idempotence : clé fournie par le client (max 128 car.), rejouée = même commande.
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(key) > 128 {
		key = key[:128]
	}
	if key != "" {
		var oid int64
		if err := a.DB.QueryRow(`SELECT order_id FROM idempotency_keys WHERE cle=?`, key).Scan(&oid); err == nil {
			a.ok(w)
			return
		}
	}
	type item struct {
		id  int64
		qty int
		pr  *int64
	}
	var items []item
	var total int64
	for _, part := range strings.Split(r.FormValue("items"), ",") {
		// Le séparateur est le DERNIER "x" : les slugs peuvent eux-mêmes
		// contenir des "x" (ex. lux-pad-ux → "lux-pad-uxx3").
		piece := strings.TrimSpace(part)
		sep := strings.LastIndex(piece, "x")
		if sep < 0 {
			continue
		}
		slug, sqty := piece[:sep], piece[sep+1:]
		q, _ := strconv.Atoi(sqty)
		if q < 1 {
			q = 1
		}
		if q > 10 {
			q = 10
		}
		var id int64
		var prix sql.NullInt64
		if err := a.DB.QueryRow(`SELECT id,COALESCE(prix_boutique,prix) FROM products WHERE slug=?`, slug).Scan(&id, &prix); err != nil {
			continue
		}
		var p *int64
		if prix.Valid {
			v := prix.Int64
			p = &v
			total += v * int64(q)
		}
		items = append(items, item{id, q, p})
	}
	if len(items) == 0 {
		http.Error(w, "panier vide", http.StatusBadRequest)
		return
	}
	moyen := strings.TrimSpace(r.FormValue("moyen"))
	// Clés acceptées : whatsapp (canal historique) + moyens cochés en admin
	// (+ livraison/virement, canaux legacy). Un moyen désactivé après
	// chargement du panier est refusé : repli explicite sur whatsapp,
	// jamais de confiance aveugle au client.
	okMoyen := moyen == "whatsapp" || moyen == "livraison" || moyen == "virement"
	if !okMoyen {
		for _, m := range a.moyensPaiementActifs() {
			if m.Cle == moyen {
				okMoyen = true
				break
			}
		}
	}
	if !okMoyen {
		moyen = "whatsapp"
	}
	// Promo : code client optionnel, validé côté serveur (actif, non
	// expiré, sous le plafond d'usage). Invalide/expiré/épuisé = ignoré
	// sans erreur ; remise plafonnée au total (jamais de total négatif).
	promoCode := ""
	var remise int64
	if pc := strings.ToUpper(strings.TrimSpace(r.FormValue("promo"))); pc != "" {
		var typ, expire string
		var valeur, usages int64
		var limite sql.NullInt64
		var actif int
		if err := a.DB.QueryRow(`SELECT type,valeur,usage_count,limite,expire,actif FROM promos WHERE code=?`, pc).Scan(&typ, &valeur, &usages, &limite, &expire, &actif); err == nil && actif == 1 &&
			(expire == "" || expire >= time.Now().Format("2006-01-02")) &&
			(!limite.Valid || usages < limite.Int64) && valeur > 0 {
			switch typ {
			case "pourcentage":
				if valeur > 100 {
					valeur = 100
				}
				remise = total * valeur / 100
			default:
				remise = valeur
			}
			if remise > total {
				remise = total
			}
			if remise > 0 {
				promoCode = pc
			}
		}
	}
	total -= remise
	client, _ := json.Marshal(map[string]string{"nom": r.FormValue("nom"), "tel": r.FormValue("tel")})
	var uid any
	if id := auth.CurrentUserID(a.DB, r); id != nil {
		uid = *id
	}
	tx, err := a.DB.Begin()
	if err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	defer tx.Rollback()
	// Re-vérifie la clé dans la transaction (course entre deux requêtes).
	if key != "" {
		var oid int64
		if err := tx.QueryRow(`SELECT order_id FROM idempotency_keys WHERE cle=?`, key).Scan(&oid); err == nil {
			a.ok(w)
			return
		}
	}
	res, err := tx.Exec(`INSERT INTO orders(user_id,statut,moyen_paiement,total,client_json,promo_code,remise) VALUES(?,?,?,?,?,?,?)`,
		uid, "whatsapp_recu", moyen, total, string(client), promoCode, remise)
	if err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	oid, _ := res.LastInsertId()
	if promoCode != "" {
		if _, err := tx.Exec(`UPDATE promos SET usage_count=usage_count+1 WHERE code=?`, promoCode); err != nil {
			http.Error(w, "indisponible", 500)
			return
		}
	}
	for _, it := range items {
		var p any
		if it.pr != nil {
			p = *it.pr
		}
		if _, err := tx.Exec(`INSERT INTO order_items(order_id,product_id,qty,prix_unitaire) VALUES(?,?,?,?)`, oid, it.id, it.qty, p); err != nil {
			http.Error(w, "indisponible", 500)
			return
		}
	}
	// Événement de création (la timeline part de la création, pas du 1er changement).
	if _, err := tx.Exec(`INSERT INTO order_events(order_id,de_statut,vers_statut,par) VALUES(?,?,?,?)`, oid, "", "whatsapp_recu", "site"); err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	if key != "" {
		if _, err := tx.Exec(`INSERT INTO idempotency_keys(cle,order_id) VALUES(?,?)`, key, oid); err != nil {
			http.Error(w, "indisponible", 500)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		// Course : une requête concurrente a gagné — rejoue sa commande.
		if key != "" {
			var oid2 int64
			if err := a.DB.QueryRow(`SELECT order_id FROM idempotency_keys WHERE cle=?`, key).Scan(&oid2); err == nil {
				a.ok(w)
				return
			}
		}
		http.Error(w, "indisponible", 500)
		return
	}
	a.ok(w)
}

// POST /api/devis|suivi|contact|infos — demandes enregistrées comme commandes "devis".
func (a *App) ApiDemande(statut string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !a.checkCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		r.ParseForm()
		kv := map[string]string{}
		for k, v := range r.Form {
			if k == "csrf" || len(v) == 0 {
				continue
			}
			kv[k] = v[0]
		}
		client, _ := json.Marshal(kv)
		var uid any
		if id := auth.CurrentUserID(a.DB, r); id != nil {
			uid = *id
		}
		_, _ = a.DB.Exec(`INSERT INTO orders(user_id,statut,moyen_paiement,total,client_json) VALUES(?,?,?,?,?)`,
			uid, statut, "whatsapp", 0, string(client))
		a.ok(w)
	}
}

func (a *App) ApiCSRF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"csrf": a.csrf(r)})
}

// GET /api/recherche?q=... — recherche serveur (FTS5 sur l'index complet :
// nom, marque, catégorie, accroche, description, index de recherche).
// Les mots sont nettoyés et préfixés : "ton" trouve "toner".
func (a *App) ApiRecherche(w http.ResponseWriter, r *http.Request) {
	type hit struct {
		Slug  string `json:"slug"`
		Nom   string `json:"nom"`
		Prix  *int64 `json:"prix"`
		Image string `json:"image"`
	}
	out := []hit{}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 100 {
		q = q[:100]
	}
	// Filtre optionnel par catégorie principale (sélecteur de l'en-tête) :
	// le slug vient de catalog.MainSlug, on ne fait confiance qu'aux valeurs connues.
	cat := strings.TrimSpace(r.URL.Query().Get("cat"))
	catOK := false
	for _, m := range catalog.MainOrder {
		if catalog.MainSlug[m] == cat {
			catOK = true
			break
		}
	}
	if !catOK {
		cat = ""
	}
	var terms []string
	for _, w0 := range strings.FieldsFunc(q, func(c rune) bool {
		return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c > 127)
	}) {
		if len(w0) >= 2 {
			terms = append(terms, `"`+w0+`"*`)
		}
	}
	if len(terms) > 0 {
		match := strings.Join(terms, " AND ")
		// Prix affiché = prix boutique s'il existe, sinon prix stock.
		sqlQ := `SELECT p.slug, p.nom, COALESCE(p.prix_boutique, p.prix), p.image
			FROM products_fts JOIN products p ON p.slug = products_fts.slug
			WHERE products_fts MATCH ?`
		args := []any{match}
		if cat != "" {
			sqlQ += ` AND p.main_slug = ?`
			args = append(args, cat)
		}
		sqlQ += ` ORDER BY rank LIMIT 10`
		rows, err := a.DB.Query(sqlQ, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var h hit
				var prix sql.NullInt64
				if err := rows.Scan(&h.Slug, &h.Nom, &prix, &h.Image); err == nil {
					if prix.Valid {
						v := prix.Int64
						h.Prix = &v
					}
					out = append(out, h)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"results": out})
}

// ---------- Auth (compte optionnel, sessions + bcrypt) ----------

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !a.checkCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	pw := r.FormValue("password")
	nom := strings.TrimSpace(r.FormValue("nom"))
	prenom := strings.TrimSpace(r.FormValue("prenom"))
	if !strings.Contains(email, "@") || len(pw) < 6 || len(email) > 120 {
		http.Error(w, "données invalides", http.StatusBadRequest)
		return
	}
	h, _ := auth.HashPassword(pw)
	_, err := a.DB.Exec(`INSERT INTO users(email,pass_hash,nom,tel) VALUES(?,?,?,?)`, email, h, strings.TrimSpace(prenom+" "+nom), "")
	if err != nil {
		http.Error(w, "compte existant ou invalide", http.StatusBadRequest)
		return
	}
	var uid int64
	_ = a.DB.QueryRow(`SELECT id FROM users WHERE email=?`, email).Scan(&uid)
	tok, _ := auth.CreateSession(a.DB, uid, r, 30*24*time.Hour)
	auth.SetCookie(w, tok, 30*24*time.Hour, a.SecureCookies)
	if !a.mirror(r) {
		http.Redirect(w, r, "/compte.html", http.StatusSeeOther)
		return
	}
	a.ok(w)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !a.checkCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	var uid int64
	var hash, nom string
	if err := a.DB.QueryRow(`SELECT id,pass_hash,nom FROM users WHERE email=?`, email).Scan(&uid, &hash, &nom); err != nil ||
		!auth.CheckPassword(hash, r.FormValue("password")) {
		http.Error(w, "identifiants invalides", http.StatusUnauthorized)
		return
	}
	tok, _ := auth.CreateSession(a.DB, uid, r, 30*24*time.Hour)
	auth.SetCookie(w, tok, 30*24*time.Hour, a.SecureCookies)
	if !a.mirror(r) {
		http.Redirect(w, r, "/compte.html", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ok": "1", "nom": nom})
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	auth.DeleteSession(a.DB, r)
	auth.ClearCookie(w)
	a.ok(w)
}

// POST /auth/mot-de-passe — via page compte (miroir).
func (a *App) Password(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !a.checkCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	uid := auth.CurrentUserID(a.DB, r)
	if uid == nil {
		http.Error(w, "non connecté", http.StatusUnauthorized)
		return
	}
	pw := r.FormValue("password")
	if len(pw) < 6 {
		http.Error(w, "6 caractères minimum", http.StatusBadRequest)
		return
	}
	h, _ := auth.HashPassword(pw)
	_, _ = a.DB.Exec(`UPDATE users SET pass_hash=? WHERE id=?`, h, *uid)
	a.ok(w)
}
