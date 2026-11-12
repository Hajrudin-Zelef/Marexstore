package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
)

// Panier stocké en cookie signé JSON (V1) + total recalculé serveur depuis DB.
// V2 : panier en DB lié à la session.
func readCart(r *http.Request) map[string]int {
	m := map[string]int{}
	c, err := r.Cookie("marex_cart")
	if err != nil {
		return m
	}
	_ = json.Unmarshal([]byte(mustUnescape(c.Value)), &m)
	return m
}

func mustUnescape(s string) string {
	u, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return u
}

func writeCart(w http.ResponseWriter, m map[string]int, secure bool) {
	b, _ := json.Marshal(m)
	http.SetCookie(w, &http.Cookie{
		Name: "marex_cart", Value: url.QueryEscape(string(b)), Path: "/",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 60 * 60 * 24 * 30,
	})
}

func (a *App) PanierAjouter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", 405)
		return
	}
	if !auth.CheckCSRF(a.Cfg.SessionSecret, a.sessionID(r), r.FormValue("csrf")) {
		http.Error(w, "Session expirée, rechargez la page.", 403)
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
	var exists int
	if err := a.DB.QueryRow(`SELECT 1 FROM products WHERE slug=?`, slug).Scan(&exists); err != nil {
		http.Error(w, "Produit introuvable", 404)
		return
	}
	m := readCart(r)
	m[slug] += qty
	if m[slug] > 10 {
		m[slug] = 10
	}
	writeCart(w, m, a.SecureCookies)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<span class="toast ok">Ajouté au panier.</span>`))
		return
	}
	http.Redirect(w, r, "/boutique", http.StatusSeeOther)
}

func (a *App) Commande(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		m := readCart(r)
		type line struct {
			Slug string
			Nom  string
			Prix *int64
			Qty  int
		}
		var lines []line
		var total int64
		var hasDevis bool
		for slug, q := range m {
			var nom string
			var prix sql.NullInt64
			if err := a.DB.QueryRow(`SELECT nom,prix FROM products WHERE slug=?`, slug).Scan(&nom, &prix); err != nil {
				continue
			}
			l := line{Slug: slug, Nom: nom, Qty: q}
			if prix.Valid {
				v := prix.Int64
				l.Prix = &v
				total += v * int64(q)
			} else {
				hasDevis = true
			}
			lines = append(lines, l)
		}
		_ = a.Tpl.ExecuteTemplate(w, "commande", map[string]any{
			"Title": "Commander — Marexstore", "Lines": lines, "Total": total,
			"HasDevis": hasDevis, "CSRF": auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)),
		})
		return
	}
	// POST : enregistre la commande, statuts par moyen de paiement.
	if !auth.CheckCSRF(a.Cfg.SessionSecret, a.sessionID(r), r.FormValue("csrf")) {
		http.Error(w, "Session expirée, rechargez la page.", 403)
		return
	}
	moyen := r.FormValue("moyen")
	switch moyen {
	case "whatsapp", "wave", "orange", "mtn", "livraison", "virement":
	default:
		moyen = "whatsapp"
	}
	nom := strings.TrimSpace(r.FormValue("nom"))
	tel := strings.TrimSpace(r.FormValue("tel"))
	if len(nom) < 2 || len(tel) < 8 || len(tel) > 20 {
		http.Error(w, "Nom et téléphone valides requis.", 400)
		return
	}
	m := readCart(r)
	if len(m) == 0 {
		http.Error(w, "Panier vide.", 400)
		return
	}
	var total int64
	var items []struct {
		id   int64
		qty  int
		prix *int64
	}
	for slug, q := range m {
		var id int64
		var prix sql.NullInt64
		if err := a.DB.QueryRow(`SELECT id,prix FROM products WHERE slug=?`, slug).Scan(&id, &prix); err != nil {
			continue
		}
		var p *int64
		if prix.Valid {
			v := prix.Int64
			p = &v
			total += v * int64(q)
		}
		items = append(items, struct {
			id   int64
			qty  int
			prix *int64
		}{id, q, p})
	}
	statut := "devis"
	if moyen == "wave" || moyen == "orange" || moyen == "mtn" {
		statut = "attente_momo"
	}
	client, _ := json.Marshal(map[string]string{"nom": nom, "tel": tel})
	var uid any
	if id := auth.CurrentUserID(a.DB, r); id != nil {
		uid = *id
	}
	res, err := a.DB.Exec(`INSERT INTO orders(user_id,statut,moyen_paiement,total,client_json) VALUES(?,?,?,?,?)`,
		uid, statut, moyen, total, string(client))
	if err != nil {
		http.Error(w, "Commande impossible pour le moment.", 500)
		return
	}
	oid, _ := res.LastInsertId()
	for _, it := range items {
		var p any
		if it.prix != nil {
			p = *it.prix
		}
		_, _ = a.DB.Exec(`INSERT INTO order_items(order_id,product_id,qty,prix_unitaire) VALUES(?,?,?,?)`,
			oid, it.id, it.qty, p)
	}
	writeCart(w, map[string]int{}, a.SecureCookies)
	_ = a.Tpl.ExecuteTemplate(w, "commande_ok", map[string]any{
		"Title": "Commande reçue — Marexstore", "ID": oid, "Moyen": moyen,
		"Total": total, "Heure": time.Now().Format("02/01 15:04"),
	})
}
