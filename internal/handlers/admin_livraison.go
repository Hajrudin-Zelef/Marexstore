package handlers

// Page admin Livraison : zones (nom, type, frais, seuil franco, délai,
// ordre, actif) + réglages globaux (seuil Abidjan, adresse de retrait).
// Types fermés : retrait|livraison|expedition (enforcés ici, sans CHECK en
// base). Actions via champ caché `action` (+ `id`) : ajouter|maj|toggle|
// supprimer ; `supprimer` = DELETE réel (aucune FK). La suppression n'est
// jamais destructrice pour les commandes : les commandes archivées portent
// leur propre copie (livraison_zone_id, livraison_frais).

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

type livraisonZone struct {
	id          int64
	nom, typ    string
	frais       int64
	seuilFranco int64
	delai       string
	ordre       int64
	actif       bool
}

func parseMontantLivraison(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}

func (a *App) livraisonGetZones() []livraisonZone {
	rows, err := a.DB.Query(`SELECT id,nom,type,frais,seuil_franco,delai,ordre,actif FROM livraison_zones ORDER BY ordre ASC, id ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var zones []livraisonZone
	for rows.Next() {
		var z livraisonZone
		var actif int
		if err := rows.Scan(&z.id, &z.nom, &z.typ, &z.frais, &z.seuilFranco, &z.delai, &z.ordre, &actif); err != nil {
			continue
		}
		z.actif = actif != 0
		zones = append(zones, z)
	}
	return zones
}

func (a *App) livraisonGet(k, def string) string {
	var v string
	if err := a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, k).Scan(&v); err != nil {
		return def
	}
	return v
}

func (a *App) livraisonSet(k, v string) {
	_, _ = a.DB.Exec(`INSERT INTO settings(nom,valeur) VALUES(?,?) ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`, k, v)
}

// AdminLivraison : GET liste + formulaires ; POST actions zone/globals.
func (a *App) AdminLivraison(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		switch r.FormValue("action") {
		case "globals":
			seuil := strings.TrimSpace(r.FormValue("seuil_abidjan"))
			if seuil == "" {
				seuil = "0"
			}
			if v, err := strconv.ParseInt(seuil, 10, 64); err != nil || v < 0 {
				http.Error(w, "seuil invalide", http.StatusBadRequest)
				return
			}
			a.livraisonSet("livraison_seuil_abidjan", seuil)
			a.livraisonSet("livraison_retrait_adresse", strings.TrimSpace(r.FormValue("retrait_adresse")))
			a.auditAdmin(r, "livraison.globals", "seuil_abidjan="+seuil)
			http.Redirect(w, r, "/admin/livraison?msg=R%C3%A9glages+livraison+enregistr%C3%A9s", http.StatusSeeOther)
			return
		case "ajouter":
			nom := strings.TrimSpace(r.FormValue("nom"))
			typ := r.FormValue("type")
			switch typ {
			case "retrait", "livraison", "expedition":
			default:
				http.Error(w, "type invalide", http.StatusBadRequest)
				return
			}
			if nom == "" {
				http.Error(w, "nom requis", http.StatusBadRequest)
				return
			}
			frais, ok1 := parseMontantLivraison(r.FormValue("frais"))
			seuil, ok2 := parseMontantLivraison(r.FormValue("seuil_franco"))
			ordre, ok3 := parseMontantLivraison(r.FormValue("ordre"))
			if !ok1 || !ok2 || !ok3 {
				http.Error(w, "montant invalide", http.StatusBadRequest)
				return
			}
			if _, err := a.DB.Exec(`INSERT INTO livraison_zones(nom,type,frais,seuil_franco,delai,ordre,actif) VALUES(?,?,?,?,?,?,1)`,
				nom, typ, frais, seuil, strings.TrimSpace(r.FormValue("delai")), ordre); err != nil {
				http.Error(w, "enregistrement impossible", http.StatusInternalServerError)
				return
			}
			a.auditAdmin(r, "livraison.ajouter", nom)
			http.Redirect(w, r, "/admin/livraison?msg=Zone+ajout%C3%A9e", http.StatusSeeOther)
			return
		case "maj":
			id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "zone inconnue", http.StatusBadRequest)
				return
			}
			nom := strings.TrimSpace(r.FormValue("nom"))
			typ := r.FormValue("type")
			switch typ {
			case "retrait", "livraison", "expedition":
			default:
				http.Error(w, "type invalide", http.StatusBadRequest)
				return
			}
			if nom == "" {
				http.Error(w, "nom requis", http.StatusBadRequest)
				return
			}
			frais, ok1 := parseMontantLivraison(r.FormValue("frais"))
			seuil, ok2 := parseMontantLivraison(r.FormValue("seuil_franco"))
			ordre, ok3 := parseMontantLivraison(r.FormValue("ordre"))
			if !ok1 || !ok2 || !ok3 {
				http.Error(w, "montant invalide", http.StatusBadRequest)
				return
			}
			if _, err := a.DB.Exec(`UPDATE livraison_zones SET nom=?,type=?,frais=?,seuil_franco=?,delai=?,ordre=? WHERE id=?`,
				nom, typ, frais, seuil, strings.TrimSpace(r.FormValue("delai")), ordre, id); err != nil {
				http.Error(w, "enregistrement impossible", http.StatusInternalServerError)
				return
			}
			a.auditAdmin(r, "livraison.maj", nom)
			http.Redirect(w, r, "/admin/livraison?msg=Zone+mise+%C3%A0+jour", http.StatusSeeOther)
			return
		case "toggle":
			id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "zone inconnue", http.StatusBadRequest)
				return
			}
			if _, err := a.DB.Exec(`UPDATE livraison_zones SET actif=1-actif WHERE id=?`, id); err != nil {
				http.Error(w, "enregistrement impossible", http.StatusInternalServerError)
				return
			}
			a.auditAdmin(r, "livraison.toggle", r.FormValue("id"))
			http.Redirect(w, r, "/admin/livraison?msg=Zone+mise+%C3%A0+jour", http.StatusSeeOther)
			return
		case "supprimer":
			id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "zone inconnue", http.StatusBadRequest)
				return
			}
			if _, err := a.DB.Exec(`DELETE FROM livraison_zones WHERE id=?`, id); err != nil {
				http.Error(w, "suppression impossible", http.StatusInternalServerError)
				return
			}
			a.auditAdmin(r, "livraison.supprimer", r.FormValue("id"))
			http.Redirect(w, r, "/admin/livraison?msg=Zone+supprim%C3%A9e", http.StatusSeeOther)
			return
		default:
			http.Error(w, "action inconnue", http.StatusBadRequest)
			return
		}
	}

	zones := a.livraisonGetZones()
	seuil := a.livraisonGet("livraison_seuil_abidjan", "100000")
	adresse := a.livraisonGet("livraison_retrait_adresse", "Plateau, Abidjan")
	csrf := render.Esc(a.adminCSRF())

	var b strings.Builder
	b.WriteString(mxPageHead("Livraison", "Zones, frais et seuils franco — montants en F CFA, modifiables ici."))

	// Liste + édition inline.
	var rows strings.Builder
	for _, z := range zones {
		etat := "Active"
		if !z.actif {
			etat = "Masquée"
		}
		etatClass := ""
		if !z.actif {
			etatClass = " off"
		}
		sel := func(v string) string {
			if z.typ == v {
				return " selected"
			}
			return ""
		}
		rows.WriteString(`<form method="post" class="mx-liv-row"><input type="hidden" name="csrf" value="` + csrf + `"><input type="hidden" name="action" value="maj"><input type="hidden" name="id" value="` + strconv.FormatInt(z.id, 10) + `">` +
			`<strong>` + render.Esc(z.nom) + `</strong><span class="mx-liv-etat` + etatClass + `">` + etat + `</span>` +
			`<label>Nom <input name="nom" value="` + render.Esc(z.nom) + `" required></label>` +
			`<label>Type <select name="type"><option value="retrait"` + sel("retrait") + `>Retrait</option><option value="livraison"` + sel("livraison") + `>Livraison</option><option value="expedition"` + sel("expedition") + `>Expédition</option></select></label>` +
			`<label>Frais (F CFA) <input name="frais" type="number" min="0" value="` + strconv.FormatInt(z.frais, 10) + `"></label>` +
			`<label>Seuil franco (F CFA) <input name="seuil_franco" type="number" min="0" value="` + strconv.FormatInt(z.seuilFranco, 10) + `"></label>` +
			`<label>Délai <input name="delai" value="` + render.Esc(z.delai) + `"></label>` +
			`<label>Ordre <input name="ordre" type="number" min="0" value="` + strconv.FormatInt(z.ordre, 10) + `"></label>` +
			`<button class="mx-btn" type="submit">Mettre à jour</button></form>` +
			`<form method="post" class="mx-liv-mini"><input type="hidden" name="csrf" value="` + csrf + `"><input type="hidden" name="id" value="` + strconv.FormatInt(z.id, 10) + `">` +
			`<button class="mx-btn ghost" type="submit" name="action" value="toggle">Activer / masquer</button> ` +
			`<button class="mx-btn ghost" type="submit" name="action" value="supprimer" onclick="return confirm('Supprimer cette zone ?')">Supprimer</button></form>`)
	}
	liste := rows.String()
	if liste == "" {
		liste = mxEmptyState("Aucune zone", "Ajoutez votre première zone avec le formulaire ci-dessous.")
	}
	b.WriteString(mxSection("Zones de livraison", "Ordre d'affichage croissant ; masquée = invisible au tunnel d'achat",
		`<div class="mx-fsec mx-fsec-premium">` + reglagesHead(mxIcoShop, "Zones", "Frais et seuils franco par zone.", "") + liste + `</div>`))

	// Formulaire d'ajout.
	b.WriteString(mxSection("Nouvelle zone", "Le type est fermé : retrait, livraison ou expédition",
		mxCard(`<form method="post" class="mx-form"><input type="hidden" name="csrf" value="`+csrf+`"><input type="hidden" name="action" value="ajouter">`+
			`<label>Nom <input name="nom" required></label>`+
			`<label>Type <select name="type"><option value="retrait">Retrait</option><option value="livraison" selected>Livraison</option><option value="expedition">Expédition</option></select></label>`+
			`<label>Frais (F CFA) <input name="frais" type="number" min="0" value="0"></label>`+
			`<label>Seuil franco (F CFA) <input name="seuil_franco" type="number" min="0" value="0"></label>`+
			`<label>Délai <input name="delai" placeholder="Ex. 24–72h"></label>`+
			`<label>Ordre <input name="ordre" type="number" min="0" value="0"></label>`+
			`<button class="mx-btn" type="submit">Ajouter la zone</button></form>`)))

	// Réglages globaux.
	b.WriteString(mxSection("Réglages globaux", "Seuil franco Abidjan et adresse de retrait",
		`<div class="mx-fsec mx-fsec-premium" style="max-width:640px">`+reglagesHead(mxIcoCash, "Seuils", "Appliqués quand une zone ne définit pas son propre seuil.", "")+
			`<form method="post"><input type="hidden" name="csrf" value="`+csrf+`"><input type="hidden" name="action" value="globals">`+
			`<div class="mx-fgrid"><label>Seuil franco Abidjan (F CFA) <input name="seuil_abidjan" type="number" min="0" value="`+render.Esc(seuil)+`"></label>`+
			`<label class="span2">Adresse de retrait <input name="retrait_adresse" value="`+render.Esc(adresse)+`"></label></div>`+
			`<div class="mx-form-actions"><button class="mx-btn primary">Enregistrer</button></div></form></div>`))

	a.writeAdmin(w, r, "Livraison", "livraison", b.String())
}
