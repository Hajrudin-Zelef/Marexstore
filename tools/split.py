#!/usr/bin/env python3
"""Découpe les pages statiques de site-orig en gabarits Go verbatim.

Vérifie que le chrome (top/drawer) est identique partout (au préfixe près),
puis écrit pour chaque page son corps spécifique avec {{.Prefix}} et les
points d'insertion CSRF. Les pages dynamiques (index/boutique/categories/
produits) sont rendues par internal/render et ignorées ici.
"""
import re, sys
from pathlib import Path

REF = Path("site-orig")
OUT = Path("templates/static")
OUT.mkdir(parents=True, exist_ok=True)

DYNAMIC = {"index.html", "boutique.html"}
DYNAMIC_DIRS = {"produits", "categories"}

DRAWER_START = '<div class="overlay" data-cart-overlay></div>'
DRAWER_END_MARK = "Pensez à noter votre mot de passe.</p>\n</div>\n</div>"
WIDGETS_START = '<div class="sec"><div class="widgets">'

tops, drawers = {}, {}

def norm_prefix(h, prefix):
    if prefix == "../":
        # sous-dossier : ../X -> {{.Prefix}}X
        return h.replace('"../', '"{{.Prefix}}')
    return h.replace('"assets/', '"{{.Prefix}}assets/').replace('"images/', '"{{.Prefix}}images/') \
        .replace('"index.html', '"{{.Prefix}}index.html').replace('"boutique.html', '"{{.Prefix}}boutique.html') \
        .replace('"devis.html', '"{{.Prefix}}devis.html').replace('"contact.html', '"{{.Prefix}}contact.html') \
        .replace('"compte.html', '"{{.Prefix}}compte.html').replace('"suivi-commande.html', '"{{.Prefix}}suivi-commande.html') \
        .replace('"parametres.html', '"{{.Prefix}}parametres.html').replace('"services.html', '"{{.Prefix}}services.html') \
        .replace('"apropos.html', '"{{.Prefix}}apropos.html').replace('"qui-sommes-nous.html', '"{{.Prefix}}qui-sommes-nous.html') \
        .replace('"nos-services.html', '"{{.Prefix}}nos-services.html').replace('"nos-references.html', '"{{.Prefix}}nos-references.html') \
        .replace('"contactez-nous.html', '"{{.Prefix}}contactez-nous.html').replace('"solutions-', '"{{.Prefix}}solutions-') \
        .replace('"cgv.html', '"{{.Prefix}}cgv.html').replace('"livraison.html', '"{{.Prefix}}livraison.html') \
        .replace('"confidentialite.html', '"{{.Prefix}}confidentialite.html').replace('"mentions-legales.html', '"{{.Prefix}}mentions-legales.html') \
        .replace('"categories/', '"{{.Prefix}}categories/').replace('"produits/', '"{{.Prefix}}produits/')

pages = sorted([p for p in REF.glob("*.html") if p.name not in DYNAMIC])
assert pages, "aucune page statique"

ref_top = ref_drawer = None
for p in pages:
    h = p.read_text(encoding="utf-8")
    prefix = ""  # toutes les statiques sont à la racine
    bi = h.find("<body>")
    assert bi > 0, p
    ni = h.find("</nav>", bi)
    assert ni > 0, p
    top = h[bi + len("<body>"):ni + len("</nav>")]
    di = h.find(DRAWER_START, ni)
    assert di > 0, p
    ei = h.find(DRAWER_END_MARK, di)
    assert ei > 0, p
    drawer = h[di:ei + len(DRAWER_END_MARK)]
    wi = h.find(WIDGETS_START, ei)
    assert wi > 0, p
    pre = h[ni + len("</nav>"):di]
    assert pre.strip() == "", f"inattendu entre nav et drawer: {p.name}"
    body = h[ei + len(DRAWER_END_MARK):wi]
    foot = h[wi:]
    assert foot.rstrip().endswith("</html>"), p
    tops[p.name] = norm_prefix(top, prefix)
    drawers[p.name] = norm_prefix(drawer, prefix)
    # corps : normalise le préfixe (racine : déjà absolu-relatif, rien à faire
    # sauf les liens mailto/tel inchangés) + point CSRF pour les formulaires POST
    b = body
    # formulaires compte/devis : point d'insertion CSRF avant </form> des acct-form
    b = b.replace('class="acct-form" data-login-form>',
                  'class="acct-form" data-login-form>\n<input type="hidden" name="csrf" value="{{.CSRF}}">')
    b = b.replace('class="acct-form" data-create-form>',
                  'class="acct-form" data-create-form>\n<input type="hidden" name="csrf" value="{{.CSRF}}">')
    (OUT / (p.stem + ".html")).write_text(b, encoding="utf-8")

# le chrome doit être identique partout
t0 = next(iter(tops.values()))
for name, t in tops.items():
    assert t == t0, f"top différent: {name}"
d0 = next(iter(drawers.values()))
for name, d in drawers.items():
    assert d == d0, f"drawer différent: {name}"

# pied : widgets (rendus par Go) + reste verbatim
feet = set()
for p in pages:
    h = p.read_text(encoding="utf-8")
    wi = h.find(WIDGETS_START)
    feet.add(h[wi:])
assert len(feet) == 1, "pieds différents"
foot = next(iter(feet))
wi_end = foot.find("</div></div>\n<div class=\"news\">")
assert wi_end > 0
widgets_part = foot[:wi_end + len("</div></div>")]
rest = foot[wi_end + len("</div></div>"):]
(OUT / "_top.html").write_text(t0, encoding="utf-8")
(OUT / "_drawer.html").write_text(d0, encoding="utf-8")
(OUT / "_foot.html").write_text("{{.Widgets}}" + rest, encoding="utf-8")
print(f"chrome identique sur {len(pages)} pages ; corps écrits dans {OUT}")
