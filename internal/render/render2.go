package render

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// ---------- Liens ----------

// qesc reproduit urllib.parse.quote (espaces en %20, pas en + comme QueryEscape).
func qesc(s string) string {
	q := url.QueryEscape(s)
	q = strings.ReplaceAll(q, "+", "%20")
	return strings.ReplaceAll(q, "%2F", "/")
}

func WaLink(p *catalog.RawProduct) string {
	ref := p.RefFabricant
	if ref == "" {
		ref = p.Ref
	}
	return "https://wa.me/" + catalog.WA + "?text=" + qesc(fmt.Sprintf("Bonjour, je voudrais commander : %s (Réf. %s)", p.Nom, ref))
}

func DevisLink(p *catalog.RawProduct) string {
	return "mailto:" + catalog.Email + "?subject=" + qesc("Demande de devis — "+p.Nom)
}

// ---------- Prix ----------

func BadgePromo(p *catalog.RawProduct) string {
	if p.PrixAvant != nil && p.PrixAffiche() != nil && *p.PrixAvant > *p.PrixAffiche() {
		pct := (1 - float64(*p.PrixAffiche())/float64(*p.PrixAvant)) * 100
		return fmt.Sprintf(`<span class="promo">-%d%%</span>`, int(pct+0.5))
	}
	return ""
}

func PrixHTML(p *catalog.RawProduct) string {
	prix := p.PrixAffiche()
	h := esc(catalog.FmtPrix(prix))
	if p.PrixAvant != nil && prix != nil && *p.PrixAvant > *prix {
		h += fmt.Sprintf(` <span class="old">%s</span>`, esc(catalog.FmtPrix(p.PrixAvant)))
	}
	return h
}

// ---------- Chrome ----------

func Header(s *catalog.Store, prefix, searchQ string) string {
	var opts strings.Builder
	for _, m := range catalog.MainOrder {
		fmt.Fprintf(&opts, `<option value="%s">%s</option>`, catalog.MainSlug[m], esc(m))
	}
	return `<div class="topbar"><div class="in"><span>Bienvenue dans notre boutique</span><span class="sp"></span>
<a href="` + prefix + `suivi-commande.html">Suivi de commande</a><span>` + esc(catalog.Tel) + `</span><span>` + esc(catalog.Email) + `</span><a href="` + prefix + `compte.html" data-auth-open>Inscription / Connexion</a></div></div>
<header class="main"><div class="in">
<a class="logo" href="` + prefix + `index.html"><img src="` + prefix + `images/logo-mediatec.png" alt="Mediatec"><span class="store">Store</span></a>
<form class="search" action="` + prefix + `boutique.html" method="get"><select name="cat"><option value="">Toutes les catégories</option>
` + opts.String() + `</select>
<input name="q" type="search" placeholder="Rechercher un produit, une référence…" value="` + esc(searchQ) + `"><button type="submit">Rechercher</button></form>
<div class="hicons">
<a class="hicon" href="` + prefix + `boutique.html"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg><span class="hic-lab">Boutique</span></a>
<a class="hicon" href="` + prefix + `compte.html" data-auth-open><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg><span class="hic-lab" data-compte-lab>Compte</span></a>
<button class="hicon" type="button" data-cart-open><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/><path d="M16 10a4 4 0 0 1-8 0"/></svg><span class="hic-badge" data-cart-count>0</span><span class="hic-lab">Panier</span></button>
<a class="hicon" href="` + prefix + `parametres.html"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg><span class="hic-lab">Préférences</span></a>
</div></div></header>
<nav class="main"><div class="in">
` + MegaBtn() + `
<a href="` + prefix + `index.html">Accueil</a>
<a href="` + prefix + `services.html">Nos logiciels de Gestion</a>
<a href="` + prefix + `devis.html">Demander un devis</a>
<a href="` + prefix + `contact.html">Contact</a>
<a href="` + prefix + `qui-sommes-nous.html">Qui sommes-nous ?</a>
</div>
` + MegaPanel(s, prefix) + `
</nav>
<div class="overlay" data-cart-overlay></div>
<aside class="drawer" data-cart-drawer data-prefix="` + prefix + `" aria-label="Panier">
<div class="drawer-head"><h3>Votre panier</h3><button type="button" data-cart-close aria-label="Fermer">×</button></div>
<div class="drawer-items" data-cart-items></div>
<div class="drawer-foot">
<div class="drawer-total">Sous-total : <b data-cart-subtotal>0 F CFA</b></div>
<p class="drawer-note">Frais de livraison à confirmer avec notre équipe sur WhatsApp.</p>
<a class="btn wa" data-cart-order href="https://wa.me/` + catalog.WA + `">Commander sur WhatsApp</a>
<button class="btn ghost" type="button" data-cart-clear>Vider le panier</button>
<button class="linklike" type="button" data-cart-close>Continuer mes achats</button>
</div></aside>
<div class="toast" data-toast></div>
<div class="overlay" data-auth-overlay></div>
<div class="auth-modal" data-auth-modal>
<div class="auth-box">
<button type="button" class="auth-close" data-auth-close aria-label="Fermer">×</button>
<div class="auth-tabs">
<button type="button" data-auth-tab="login" class="on">Connexion</button>
<button type="button" data-auth-tab="create">Créer un compte</button>
</div>
<div data-auth-pane="login">
<h2>Déjà client ?</h2>
<form class="acct-form" data-login-form>
<label class="full">Votre email <input name="email" type="email" required placeholder="Votre email"></label>
<label class="full">Votre mot de passe <input name="password" type="password" required placeholder="Votre mot de passe"></label>
<div class="full acct-actions" style="justify-content:flex-start"><button class="btn" type="submit">Connexion</button></div>
<p class="full acct-note" data-login-error hidden style="color:#dc3545">Email ou mot de passe incorrect.</p>
</form>
</div>
<div data-auth-pane="create" hidden>
<h2>Nouveau client ?</h2>
<form class="acct-form" data-create-form>
<label>Civilité <select name="civilite"><option>M.</option><option>Mme</option></select></label>
<label>Prénom * <input name="prenom" required placeholder="Votre prénom"></label>
<label>Nom de famille * <input name="nom" required placeholder="Votre nom"></label>
<label>Email * <input name="email" type="email" required placeholder="Votre email"></label>
<label>Mot de passe * <input name="password" type="password" required minlength="6" placeholder="6 caractères minimum"></label>
<label>Confirmer le mot de passe * <input name="password2" type="password" required minlength="6" placeholder="Retapez votre mot de passe"></label>
<div class="full acct-actions" style="justify-content:flex-start"><button class="btn" type="submit">Créez un compte</button></div>
<p class="full acct-note" data-create-error hidden style="color:#dc3545"></p>
</form>
</div>
<p class="acct-note">Votre compte est synchronisé sur tous vos appareils : connectez-vous pour retrouver votre panier et vos commandes.</p>
</div>
</div>`
}

func MegaBtn() string {
	return `<button type="button" class="mega-btn" data-mega-btn aria-expanded="false">☰ Toutes les Catégories</button>`
}

// ---------- Cartes ----------

func AddCartBtn(p *catalog.RawProduct, qtyAttr string) string {
	pr := p.PrixAffiche()
	if pr == nil {
		return ""
	}
	img := p.Image
	return `<button class="round cart" type="button" data-add-cart data-slug="` + esc(p.Slug) + `" data-nom="` + esc(p.Nom) + `" data-prix="` + fmt.Sprint(*pr) + `" data-image="` + esc(img) + `" ` + qtyAttr + ` title="Ajouter au panier">` + CartSVG + `</button>`
}

func AddCartBtnFull(p *catalog.RawProduct) string {
	pr := p.PrixAffiche()
	if pr == nil {
		return ""
	}
	return `<button class="btn cart-btn" type="button" data-add-cart data-slug="` + esc(p.Slug) + `" data-nom="` + esc(p.Nom) + `" data-prix="` + fmt.Sprint(*pr) + `" data-image="` + esc(p.Image) + `" data-qty="#qty">` + CartSVG + ` Ajouter au panier</button>`
}

func Card(s *catalog.Store, p *catalog.RawProduct, prefix, filterKey string) string {
	im := `<div class="noimg">Photo à venir</div>`
	if p.Image != "" {
		im = `<img src="` + esc(prefix+p.Image) + `" alt="` + esc(p.Nom) + `" loading="lazy">`
	}
	var specAttrs strings.Builder
	facets := catalog.FacetValues(p)
	for _, r := range catalog.FacetRules {
		v, ok := facets[r.Slug]
		if !ok {
			continue
		}
		var vals []string
		switch t := v.(type) {
		case string:
			vals = []string{t}
		case []string:
			vals = t
		}
		var sv []string
		for _, x := range vals {
			sv = append(sv, catalog.Slugify(x))
		}
		fmt.Fprintf(&specAttrs, ` data-spec-%s="%s"`, r.Slug, esc(strings.Join(sv, " ")))
	}
	dispo := catalog.Slugify(p.Dispo)
	etat := "neuf"
	if catalog.IsReconditionne(p) {
		etat = "reconditionne"
	}
	prix := p.PrixAffiche()
	prixAttr := ""
	if prix != nil {
		prixAttr = fmt.Sprint(*prix)
	}
	filterVal := p.MainSlug
	if filterKey == "sub_slug" {
		filterVal = p.SubSlug
	}
	return `<div class="card" data-reveal data-name="` + esc(catalog.NormName(p.Nom+" "+p.RefFabricant+" "+p.Ref)) + `" data-search="` + esc(catalog.SearchIndex(p)) + `" data-sub="` + esc(filterVal) + `" data-store="` + esc(p.LiveTop) + `" data-live-sub="` + esc(p.LiveSub) + `" data-main="` + esc(p.MainSlug) + `" data-marque="` + esc(catalog.NormName(nonVide(p.Marque, "Sans marque"))) + `" data-prix="` + prixAttr + `" data-dispo="` + esc(dispo) + `" data-etat="` + etat + `"` + specAttrs.String() + `>
<div class="im">` + BadgePromo(p) + im + `</div><div class="cinfo"><div class="cat">` + esc(p.Main) + ` › ` + esc(p.Sub) + `</div>
<h3><a href="` + prefix + `produits/` + esc(p.Slug) + `.html">` + esc(p.Nom) + `</a></h3>
<p class="desc">` + esc(truncate(p.Accroche, 140)) + `</p></div>
<div class="pr">` + PrixHTML(p) + `</div>
` + AddCartBtn(p, "") + `<a class="round" href="` + WaLink(p) + `" title="Commander sur WhatsApp">&#9990;</a></div>`
}

func nonVide(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func CatalogWrap(s *catalog.Store, fs []*catalog.RawProduct, prefix, filterKey, chipsAttr, defaultView string, chips [][2]string, title string) string {
	var chipsHTML strings.Builder
	if chips != nil {
		chipsHTML.WriteString(`<div class="chips"><button type="button" class="on" data-` + chipsAttr + `="all">Toutes</button>`)
		for _, c := range chips {
			fmt.Fprintf(&chipsHTML, `<button type="button" data-%s="%s">%s</button>`, chipsAttr, esc(c[1]), esc(c[0]))
		}
		chipsHTML.WriteString(`</div>`)
	}
	var cards strings.Builder
	for _, f := range fs {
		cards.WriteString(Card(s, f, prefix, filterKey))
	}
	on := func(v string) string {
		if v == defaultView {
			return "on"
		}
		return ""
	}
	return `<div data-catalog-wrap>
<div class="toolbar">
<input class="searchbar" data-filter type="search" placeholder="Rechercher un produit, une référence…">
<div class="viewbtns">
<button type="button" class="` + on("grid") + `" data-view="grid">Grille</button>
<button type="button" class="` + on("list") + `" data-view="list">Liste</button>
<button type="button" class="` + on("thumbs") + `" data-view="thumbs">Miniatures</button>
</div>
<label style="font-size:14px;color:#666">Par page
<select data-per-page><option value="12" selected>12</option><option value="24">24</option><option value="48">48</option><option value="0">Tout afficher</option></select></label>
<label style="font-size:14px;color:#666">Trier par
<select data-sort><option value="defaut" selected>Défaut</option><option value="prix-asc">Prix croissant</option><option value="prix-desc">Prix décroissant</option><option value="nom-asc">Nom (A-Z)</option><option value="nom-desc">Nom (Z-A)</option></select></label>
</div>
` + chipsHTML.String() + `
<div class="catalog view-` + defaultView + `" data-catalog>` + cards.String() + `</div>
<div class="page-info" data-page-info></div>
<div class="pager" data-pager></div>
</div>`
}

// ---------- Pied ----------

func mini(fs []*catalog.RawProduct, prefix string) string {
	var out strings.Builder
	for _, f := range fs {
		im := ""
		if f.Image != "" {
			im = `<img src="` + esc(prefix+f.Image) + `" alt="">`
		}
		fmt.Fprintf(&out, `<div class="hcard">%s<div><h4><a href="%sproduits/%s.html">%s</a></h4><div class="pr">%s</div></div></div>`,
			im, prefix, esc(f.Slug), esc(f.Nom), PrixHTML(f))
	}
	return out.String()
}

func WidgetsHTML(s *catalog.Store, prefix string) string {
	bySlug := s.BySlug
	slugs := func(ss ...string) []*catalog.RawProduct {
		var out []*catalog.RawProduct
		for _, x := range ss {
			if f, ok := bySlug[x]; ok {
				out = append(out, f)
			}
		}
		return out
	}
	vedette := slugs("lenovo-thinkpad-e14g5-21jk00dagq", "hp-probook-460-g11-9b9c9ea", "lg-monitor-27-27up85np-w")
	top := slugs("toner-650a-yellow", "toner-205a-yellow", "toner-711a-black")
	exclu := slugs("toner-711a-cyan", "toner-205a-cyan", "ecran-hp-elite-m243")
	return `<div class="sec"><div class="widgets">
<div><h3>Mis en Avant</h3>` + mini(vedette, prefix) + `</div>
<div><h3>Les Meilleures ventes</h3>` + mini(top, prefix) + `</div>
<div><h3>Offres exclusives</h3>` + mini(exclu, prefix) + `</div>
</div></div>`
}

func Footer(s *catalog.Store, prefix string) string {
	return WidgetsHTML(s, prefix) + `
<div class="news"><div class="in"><div><h3>S'abonner à la newsletter</h3><div>Recevez nos offres et nouveautés. ` + esc(catalog.Tel) + ` / ` + esc(catalog.Tel2) + ` — ` + esc(catalog.Email) + `</div></div>
<form action="mailto:` + catalog.Email + `" method="get"><input type="email" name="body" placeholder="Votre adresse e-mail" required><button type="submit">S'inscrire</button></form></div></div>
<footer><div class="cols">
<div><h4>Accès Rapide</h4><ul><li><a href="` + prefix + `devis.html">Demande de devis</a></li><li><a href="` + prefix + `boutique.html">Store</a></li><li><a href="` + prefix + `suivi-commande.html">Suivi de commande</a></li><li><a href="` + prefix + `compte.html">Mon compte</a></li><li><a href="` + prefix + `contact.html">Contact</a></li></ul></div>
<div><h4>Informations</h4><ul><li><a href="` + prefix + `cgv.html">Conditions générales de vente</a></li><li><a href="` + prefix + `livraison.html">Livraison &amp; retours</a></li><li><a href="` + prefix + `confidentialite.html">Politique de confidentialité</a></li><li><a href="` + prefix + `mentions-legales.html">Mentions légales</a></li></ul></div>
<div><h4>Le Groupe</h4><ul><li><a href="` + prefix + `nos-services.html">Nos services</a></li><li><a href="` + prefix + `solutions-infrastructure-it.html">Infrastructure IT</a></li><li><a href="` + prefix + `solutions-collaboration.html">Collaboration</a></li><li><a href="` + prefix + `solutions-cloud-securite.html">Cloud &amp; Sécurité</a></li><li><a href="` + prefix + `nos-references.html">Nos références</a></li><li><a href="` + prefix + `apropos.html">À propos de nous</a></li><li><a href="` + prefix + `contactez-nous.html">Contactez-nous</a></li></ul></div>
<div><h4>Liens utiles</h4><ul><li><a href="` + prefix + `index.html">Accueil</a></li><li><a href="` + prefix + `services.html">Nos Logiciels de Gestion</a></li><li><a href="` + prefix + `apropos.html">À propos de Mediatec</a></li><li><a href="` + prefix + `devis.html">Demande de devis</a></li></ul></div>
<div><h4>Contact</h4><ul><li>Tél : ` + esc(catalog.Tel) + `</li><li>Tél : ` + esc(catalog.Tel2) + `</li><li>E-mail : ` + esc(catalog.Email) + `</li><li>Abidjan, Côte d'Ivoire</li></ul></div>
</div><div class="copy">© Mediatec - Innovating For All - Tout droit reservé</div></footer>
<a class="wa-float" href="https://wa.me/` + catalog.WA + `" title="WhatsApp">&#9990;</a>
<script src="` + prefix + `assets/facets.js?v=319"></script>
<script src="` + prefix + `assets/app.js?v=326"></script><script src="` + prefix + `assets/htmx.min.js?v=1912" defer></script><script src="` + prefix + `assets/mirror.js?v=1" defer></script>`
}

// ---------- Document ----------

func Head(title, prefix, extraHead string) string {
	return `<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="csrf-token" content="{{CSRF}}">
<title>` + esc(title) + ` — Mediatec Store</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
<link rel="stylesheet" href="` + prefix + `assets/style.css?v=327">
` + ThemeInit + `
<script>document.documentElement.classList.add('js')</script>
` + extraHead
}

func Doc(title, body, prefix, extraHead string) string {
	return `<!DOCTYPE html><html lang="fr"><head>` + Head(title, prefix, extraHead) + `</head><body>` + body + `</body></html>`
}

// ---------- Hub boutique ----------

func HubHTML(s *catalog.Store) string {
	var cards strings.Builder
	for _, t := range catalog.LiveCats {
		n := s.TopCounts[t.Slug]
		if n == 0 {
			continue
		}
		rep := s.HubPhoto(t.Slug)
		img := `<div class="noimg">Photo à venir</div>`
		if rep != nil {
			img = ImgTag(rep, "")
		}
		var subs strings.Builder
		for _, sl := range t.Subs {
			if c := s.SubCounts[[2]string{t.Slug, sl.Slug}]; c > 0 {
				fmt.Fprintf(&subs, `<li><a href="boutique.html?sub=%s">%s <span>(%d)</span></a></li>`, esc(sl.Slug), esc(sl.Label), c)
			}
		}
		chev, subsHTML := "", ""
		if subs.Len() > 0 {
			chev = `<button type="button" class="hub-chev" data-hub-toggle aria-label="Voir les sous-catégories">⌄</button>`
			subsHTML = `<ul class="hub-subs" hidden>` + subs.String() + `</ul>`
		}
		prod := "produits"
		if n == 1 {
			prod = "produit"
		}
		fmt.Fprintf(&cards, `<div class="hub-card" data-hub-card>
<a class="hub-main" href="boutique.html?store=%s"><span class="hub-img">%s</span><span class="hub-txt"><span class="hub-name">%s</span><span class="hub-count">%d %s</span></span></a>
%s%s</div>`, esc(t.Slug), img, esc(t.Label), n, prod, chev, subsHTML)
	}
	return `<div data-hub>
<h1 style="font-size:24px">Nos catégories</h1>
<p style="color:#666">Choisissez une catégorie pour voir ses produits.</p>
<div class="hub-grid">` + cards.String() + `</div>
<p style="margin-top:18px"><a class="btn" href="boutique.html?all=1">Voir tous les produits (` + fmt.Sprint(len(s.Products)) + `) →</a></p>
</div>`
}

// ---------- Fiche produit ----------

func SpecVal(v string) string {
	i := strings.Index(v, " (selon")
	if i > 0 {
		return esc(v[:i]) + ` <span class="spec-note">` + esc(v[i+1:]) + `</span>`
	}
	return esc(v)
}

func FicheBody(s *catalog.Store, p *catalog.RawProduct) string {
	var pf strings.Builder
	for _, x := range p.PointsForts {
		fmt.Fprintf(&pf, "<li>%s</li>", esc(x))
	}
	pfHTML := ""
	if pf.Len() > 0 {
		pfHTML = `<div class="pf-wrap"><h2>Points forts</h2><ul class='pf'>` + pf.String() + `</ul></div>`
	}
	var specs strings.Builder
	for _, sp := range p.SpecList() {
		fmt.Fprintf(&specs, "<tr><td>%s</td><td>%s</td></tr>", esc(sp.Label), SpecVal(sp.Valeur))
	}
	var consHTML strings.Builder
	if len(p.Consommables) > 0 {
		var cards strings.Builder
		for _, c := range p.Consommables {
			fmt.Fprintf(&cards, `<div class="card"><div class="cat">Consommable</div><h3>%s</h3>
<div class="pr">%s</div></div>`, esc(c.Nom), esc(c.Prix))
		}
		consHTML.WriteString("<h2>Consommables</h2><div class='catalog view-grid'>" + cards.String() + "</div>")
	}
	var sim strings.Builder
	n := 0
	for _, g := range s.Products {
		if g.Main == p.Main && g.Slug != p.Slug {
			sim.WriteString(Card(s, g, "../", "main_slug"))
			n++
			if n >= 3 {
				break
			}
		}
	}
	gal := `<div class="noimg">Photo à venir</div>`
	if p.Image != "" {
		gal = ImgTag(p, "../")
	}
	stock := "Disponible"
	if p.Stock != nil {
		stock = fmt.Sprintf("Stock : %d", *p.Stock)
	}
	addBtn := ""
	if p.PrixAffiche() != nil {
		addBtn = AddCartBtnFull(p)
	}
	marque := p.Marque
	if marque == "" {
		marque = "—"
	}
	refFab := p.RefFabricant
	if refFab == "" {
		refFab = "—"
	}
	body := `
<div class="wrap"><div class="crumb"><a href="../index.html">Accueil</a> › <a href="../boutique.html">Store</a> › <a href="../categories/` + esc(catalog.MainSlug[p.Main]) + `.html">` + esc(p.Main) + `</a> › ` + esc(p.Sub) + ` › ` + esc(p.Nom) + `</div>
<div><div class="pdetail" data-reveal>
<div class="gal">` + gal + `</div>
<div><div class="cat" style="color:#999;font-size:13px">` + esc(p.Main) + ` › ` + esc(p.Sub) + `</div>
<h1>` + esc(p.Nom) + `</h1>
<div>` + BadgeDispo(p) + `</div>
<p>` + esc(p.Accroche) + `</p>
<div class="pr">` + PrixHTML(p) + `</div>
<div class="qty">Quantité <input id="qty" type="number" value="1" min="1"> <span style="color:#666">` + esc(stock) + `</span></div>
` + addBtn + `
<a class="btn" href="` + DevisLink(p) + `">Demander un devis</a>
<a class="btn wa" href="` + WaLink(p) + `">Commander sur WhatsApp</a>
<div style="margin-top:12px;color:#666;font-size:13px">Réf. ` + esc(nonVide(p.RefFabricant, p.Ref)) + ` · Marque : ` + esc(marque) + `</div>
</div></div>
<div class="ptabs" data-reveal data-tabs>
<div class="tabs"><button type="button" class="on" data-tab="1">Description</button>
<button type="button" data-tab="2">Marque</button><button type="button" data-tab="3">Avis (0)</button></div>
<div data-pane="1"><h2>Description</h2><p class="desc-text">` + esc(p.Description) + `</p>` + pfHTML + `
<h2>Caractéristiques</h2><table class="specs">` + specs.String() + `</table>` + consHTML.String() + `</div>
<div data-pane="2" hidden><h2>Marque</h2><p>` + esc(nonVide(p.Marque, "Non précisée")) + ` — Référence fabricant : ` + esc(refFab) + `.</p></div>
<div data-pane="3" hidden><h2>Avis</h2><p>Aucun avis pour le moment.</p></div>
</div>
<div class="sec"><h2>Produits similaires</h2><div class="catalog view-grid">` + sim.String() + `</div></div>
</div></div>`
	return body
}
