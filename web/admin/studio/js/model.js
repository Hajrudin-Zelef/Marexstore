/* model.js — Studio Mediatec (MVC). Généré par découpe de studio-prototype-v16-2.html. */

const $ = s => document.querySelector(s);

const PALETTE = [
  {name:'Vert Mediatec', hex:'#2E5136'},
  {name:'Bleu', hex:'#2563eb'},
  {name:'Orange', hex:'#ea580c'},
  {name:'Violet', hex:'#7c3aed'},
  {name:'Rouge', hex:'#dc2626'},
  {name:'Noir', hex:'#111827'},
];

const BGS = [
  {name:'Blanc', hex:'#ffffff'},
  {name:'Gris clair', hex:'#f3f4f6'},
  {name:'Crème', hex:'#fef9c3'},
  {name:'Sombre', hex:'#111827'},
];

const DEFAULT_FRONT = {
  accent:'#2E5136', priceColor:'#dc2626', bg:'#ffffff', text:'#111827',
  font:'Inter', titleSize:'M', btnShape:'arrondi', cardLayout:'grille',
  showBadge:true, showAnnonce:false, footerBg:'#111827',
  customCSS:'.pv-hero{border-radius:16px}\n.pv-card{box-shadow:0 2px 10px rgba(0,0,0,.08)}',
  customHTML:''
};

const DEFAULT_ROWS = [
  {id:'hero', label:'Bandeau défilant (hero)', cols:[
    {w:100, block:{type:'Bannière', title:'HP EliteBook 850 G3', text:'15" — 8 Go — 256 Go SSD', price:'185.000 F', cta:'ACHETER', img:'💻'}}
  ]},
  {id:'minis', label:'Mini-casiers promo', cols:[
    {w:34, block:{type:'Produit', title:'Toners HP 650A', text:'Noir • Cyan • Magenta • Jaune', price:'45.000 F', img:'🖨️'}},
    {w:33, block:{type:'Produit', title:'Câble RJ45 Cat6', text:'Boîte 305 m — cuivre', price:'85.000 F', img:'🔌'}},
    {w:33, block:{type:'Produit', title:'Switch 24 ports PoE', text:'Gigabit — 250 W', price:'Sur demande', img:'🗄️'}}
  ]},
  {id:'produits', label:'Grille « Nos produits »', cols:[
    {w:50, block:{type:'Produit', title:'Lenovo ThinkPad E14', text:'i7 — 16 Go — 512 Go', price:'650.000 F', img:'💻'}},
    {w:50, block:{type:'Produit', title:'Onduleur APC 1500 VA', text:'Line-interactive — USB', price:'210.000 F', img:'🔋'}}
  ]},
];

let front = JSON.parse(JSON.stringify(DEFAULT_FRONT));

let rows = JSON.parse(JSON.stringify(DEFAULT_ROWS));

let rowSeq = 100;

let currentTab = 'front';

let shop = {layout:'grille', perPage:24, sort:'stock', showBadge:true, quickAdd:true, showStock:false,
  facets:[
    {key:'marque', label:'Marque', sub:'124 Cisco · 113 HP · 54 Huawei…', visible:true},
    {key:'categorie', label:'Catégorie', sub:'8 familles', visible:true},
    {key:'prix', label:'Prix', sub:'4 tranches', visible:false},
    {key:'dispo', label:'Disponibilité', sub:'En stock / Rupture', visible:true},
  ],
  /* v9 — boutique complétée */
  previewPage:'liste',
  showPriceSlider:true, showRefSearch:true, showStockOnly:true, showMoreFilters:true,
  showLanding:true, showSidebar:true,
  sidebarAvant:true, sidebarVentes:true, sidebarOffres:true};

let header = {
  showTopbar:true, showPhone:true, showEmail:true,
  showBoutique:true, showCompte:true, showPanier:true, showPrefs:true,
  badgeColor:'#dc2626', navBg:'#84cc16', sticky:true,
  navLinks:[
    {key:'accueil', label:'Accueil', visible:true},
    {key:'logiciels', label:'Nos logiciels de Gestion', visible:true},
    {key:'devis', label:'Demander un devis', visible:true},
    {key:'contact', label:'Contact', visible:true},
    {key:'qui', label:'Qui sommes-nous ?', visible:true},
  ]};

let footer = {
  showNewsletter:true, showCols:true,
  cols:[
    {key:'rapide', title:'Accès Rapide', visible:true, links:['Demande de devis','Store','Suivi de commande','Mon compte','Contact']},
    {key:'infos', title:'Informations', visible:true, links:['Conditions générales de vente','Livraison & retours','Politique de confidentialité','Mentions légales']},
    {key:'groupe', title:'Le Groupe', visible:true, links:['Nos services','Infrastructure IT','Collaboration','Cloud & Sécurité','Nos références','À propos de nous','Contactez-nous']},
    {key:'liens', title:'Liens utiles', visible:true, links:['Accueil','Nos Logiciels de Gestion','À propos de Mediatec','Demande de devis']},
    {key:'contact', title:'Contact', visible:true, links:['Tél : (+225) 07 05 08 75 07','Tél : (+225) 27 20 24 33 19','E-mail : sales@mediatec-ci.com','Abidjan, Côte d\u2019Ivoire']},
  ]};

let fiche = {showSpecs:true, showSimilar:true, showAvis:false, showConso:true, showWa:true, showDevis:true,
  showDiscount:true, showOld:true, showStars:true, showExpress:true, showGarantie:true, showTva:true, showShare:true, showLike:true, showQty:true, similarNb:4, gallery:'grande', btnDevis:'Demander un devis', btnWa:'WhatsApp'};

let acc = {
  hero:{show:true, arrows:true, dots:true, panel:true, panelT:'Imprimante Professionnelle', panelX:'HP PageWide Pro 750dw',
    slides:[{t:'Titre de la diapositive 1',x:'Texte de la diapositive 1',b:'Découvrir',img:''},{t:'Titre de la diapositive 2',x:'Texte de la diapositive 2',b:'Découvrir',img:''},{t:'Titre de la diapositive 3',x:'Texte de la diapositive 3',b:'Découvrir',img:''}]},
  av:{show:true, cards:[{i:'🚚',t:'Avantage 1',x:'Texte de la carte'},{i:'🛡',t:'Avantage 2',x:'Texte de la carte'},{i:'💬',t:'Avantage 3',x:'Texte de la carte'},{i:'💳',t:'Avantage 4',x:'Texte de la carte'}]},
  find:{show:true, title:'Trouver un produit', l1:'Marque', l2:'Catégorie', l3:'Type', reset:'Réinitialiser', search:'Rechercher', offer:true, offT:'OFFRE SPÉCIALE', offX:'Texte de l\u2019offre'},
  tabs:{labels:['En vedette','En Promotion','Les Mieux notés'], show:[true,true,true]},
  secs:[{t:'NOS PRODUITS',s:'',show:true},{t:'LES PLUS VENDUS',s:'',show:true},{t:'AJOUTÉ RÉCEMMENT',s:'',show:true},{t:'MIS EN AVANT',s:'',show:true}],
  dark:{show:true, t:'Titre du bandeau', x:'Texte du bandeau', b:'Bouton'},
  brands:{show:true, list:['HP','Dell','Lenovo','Canon','Epson','Cisco']},
  drawer:{vider:true,continuer:true,livraison:true}, cartView:'drawer', compteView:'modal', cpt:{oubli:true,sync:true}, help:{show:true}, misc:{wa:true,badge:'#dc2626'}
};

let panier = {title:'Votre panier', emptyText:'Votre panier est vide.', btnOrder:'Commander',
  btnWa:'Commander via WhatsApp', showFees:true, useSteps:true,
  steps:['Panier','Coordonnées','Paiement','Confirmation']};

const Tx = {
  // Front — chrome du site (visible partout)
  bienvenue:'Bienvenue dans notre boutique',
  recherche_ph:'Rechercher un produit…',
  rechercher:'Rechercher',
  toutes_cat:'Toutes les Catégories',
  bandeau:'Livraison offerte dès 100 000 F',
  acheter:'ACHETER',
  // Boutique
  boutique_titre:'Boutique',
  filtres:'Filtres',
  tri_label:'Tri :',
  btn_ajouter:'Ajouter',
  en_stock:'En stock',
  rupture:'Rupture de stock',
  produits_suffix:'produits',
  livraison_mention:'Livraison offerte dès 100 000 F',
  // Fiche
  specs:'Spécifications',
  similaires:'Produits similaires',
  avis:'Avis clients',
  conso:'Consommables compatibles',
  ajouter_panier:'Ajouter au panier',
  devis:'Demander un devis',
  wa:'WhatsApp',
  fiche_ref:'Réf.',
  // Panier
  sous_total:'Sous-total',
  livraison:'Livraison',
  livraison_offerte:'Offerte',
  total:'Total',
  articles:'articles',
  // Menu
  voir_tous:'Voir tous',
  // v9 — En-tête
  suivi_commande:'Suivi de commande',
  inscription:'Inscription / Connexion',
  boutique_ico:'BOUTIQUE',
  compte_ico:'COMPTE',
  panier_ico:'PANIER',
  prefs_ico:'PRÉFÉRENCES',
  // v9 — Pied de page
  newsletter_titre:"S'abonner à la newsletter",
  newsletter_texte:'Recevez nos offres et nouveautés.',
  newsletter_ph:'Votre adresse e-mail',
  newsletter_btn:"S'inscrire",
  copyright:'© Mediatec - Innovating For All - Tout droit reservé',
  // v9 — Boutique (compléments)
  ref_label:'Chercher une référence',
  ref_ph:'Nom du produit, référence…',
  prix_min:'Min',
  prix_max:'Max',
  appliquer:'Appliquer',
  stock_only:'Voir uniquement en stock',
  reset_filtres:'Réinitialiser',
  plus_filtres:'+ DE FILTRES',
  dispo_label:'Disponibilité',
  etat_label:'État',
  categories_titre:'Nos catégories',
  categories_sub:'Explorez toutes nos catégories de produits',
  voir_tous_produits:'Voir tous les produits',
  sidebar_avant:'Mis en Avant',
  sidebar_ventes:'Les Meilleures ventes',
  sidebar_offres:'Offres exclusives',
  produits_trouves:'produits trouvés',
  page_sur:'Page',
  tiroir_fermer:'Fermer',
  tiroir_retirer:'Retirer',
  tiroir_vider:'Vider le panier',
  tiroir_continuer:'Continuer mes achats',
  tiroir_livraison:'Frais de livraison à confirmer avec notre équipe sur WhatsApp.',
  toast_ajoute:'Ajouté au panier :',
  devis_nom:'Votre nom complet',
  devis_email:'Votre e-mail',
  devis_tel:'Téléphone',
  devis_ent:'Entreprise / Organisation',
  devis_client:'Êtes-vous déjà client ?',
  devis_oui:'Oui',
  devis_non:'Non',
  devis_continuer:'Continuer',
  aide_titre:'Besoin d’aide ?',
  aide_texte:'Notre équipe vous répond rapidement.',
  suite_titre:'Étapes suivantes',
  suite_texte:'Texte de la carte « Étapes suivantes ».',
  ph_texte:'Photo à venir',
  cpt_onglet_conn:'Connexion',
  cpt_onglet_crea:'Créer un compte',
  cpt_titre:'DÉJÀ CLIENT ?',
  cpt_email:'E-mail',
  cpt_mdp:'Mot de passe',
  cpt_oubli:'Mot de passe oublié ?',
  cpt_btn_conn:'Connexion',
  cpt_civ:'Civilité',
  cpt_prenom:'Prénom',
  cpt_nom:'Nom',
  cpt_mdp2:'Confirmer le mot de passe',
  cpt_btn_crea:'Créez un compte',
  cpt_sync:'Note de synchronisation.',
  pref_infos:'Mes informations',
  pref_cmd:'Mes commandes',
  pref_adr:'Mes adresses',
  pref_paniers:'Mes paniers sauvegardés',
  pref_params:'Mes paramètres',
  pref_requis:'COMPTE REQUIS',
  pref_theme:'Choisir votre thème et votre couleur',
  express:'Livraison Express', garantie:'Garantie 1 an', tva:'TVA incluse', partager:'Partager :', tab_desc:'Description', tab_marque:'Marque', tab_avis:'Avis (0)', aucun_avis:'Aucun avis pour le moment', premier_avis:'Soyez le premier à donner votre avis !', evaluer:'Évaluez ce produit', votre_nom:'Votre nom', votre_avis:'Votre avis', envoyer_avis:'Envoyer mon avis',
};

function txt(k){ return Tx[k] !== undefined ? Tx[k] : ''; }

function setT(k, v){
  Tx[k] = v;
  /* synchronise tous les champs qui affichent ce texte (module + onglet Textes),
     sauf celui qu'on est en train de taper dedans */
  document.querySelectorAll('input[data-k="'+k+'"]').forEach(el=>{
    if(el !== document.activeElement) el.value = v;
  });
  renderPreview();
}

const MODULE_TEXTS = {
  front:[
    {key:'bienvenue', label:'Texte de la barre du haut'},
    {key:'recherche_ph', label:'Texte dans le champ de recherche'},
    {key:'rechercher', label:'Bouton « Rechercher »'},
    {key:'toutes_cat', label:'Barre « Toutes les Catégories »'},
    {key:'bandeau', label:'Texte du bandeau annonce'},
    {key:'acheter', label:'Bouton « Acheter » (hero)'},
  ],
  boutique:[
    {key:'boutique_titre', label:'Titre de la page'},
    {key:'filtres', label:'Titre des filtres'},
    {key:'tri_label', label:'Libellé du tri'},
    {key:'btn_ajouter', label:'Bouton « Ajouter » sur les cartes'},
    {key:'en_stock', label:'Mention « En stock »'},
    {key:'rupture', label:'Mention « Rupture de stock »'},
    {key:'produits_suffix', label:'Mot après le nombre de produits'},
    {key:'livraison_mention', label:'Mention « Livraison offerte »'},
    {key:'ref_label', label:'Titre « Chercher une référence »'},
    {key:'ref_ph', label:'Texte du champ référence'},
    {key:'appliquer', label:'Bouton « Appliquer » (prix)'},
    {key:'stock_only', label:'Case « Voir uniquement en stock »'},
    {key:'reset_filtres', label:'Bouton « Réinitialiser » (filtres)'},
    {key:'plus_filtres', label:'Bouton « + DE FILTRES »'},
    {key:'dispo_label', label:'Titre « Disponibilité »'},
    {key:'etat_label', label:'Titre « État »'},
    {key:'categories_titre', label:'Titre « Nos catégories »'},
    {key:'categories_sub', label:'Sous-titre des catégories'},
    {key:'voir_tous_produits', label:'Lien « Voir tous les produits »'},
    {key:'sidebar_avant', label:'Titre « Mis en Avant »'},
    {key:'sidebar_ventes', label:'Titre « Les Meilleures ventes »'},
    {key:'sidebar_offres', label:'Titre « Offres exclusives »'},
    {key:'produits_trouves', label:'Mot « produits trouvés »'},
  ],
  fiche:[
    {key:'specs', label:'Titre « Spécifications »'},
    {key:'similaires', label:'Titre « Produits similaires »'},
    {key:'avis', label:'Titre « Avis clients »'},
    {key:'conso', label:'Titre « Consommables compatibles »'},
    {key:'ajouter_panier', label:'Bouton « Ajouter au panier »'},
    {key:'devis', label:'Bouton « Demander un devis »'},
    {key:'wa', label:'Bouton « WhatsApp »'},
    {key:'fiche_ref', label:'Libellé « Réf. »'},
    {key:'en_stock', label:'Mention « En stock »'},
    {key:'rupture', label:'Mention « Rupture de stock »'},
    {key:'express', label:'Texte « Livraison Express »'},
    {key:'garantie', label:'Texte « Garantie »'},
    {key:'tva', label:'Texte « TVA incluse »'},
    {key:'partager', label:'Libellé « Partager : »'},
    {key:'tab_desc', label:'Onglet « Description »'},
    {key:'tab_marque', label:'Onglet « Marque »'},
    {key:'tab_avis', label:'Onglet « Avis »'},
    {key:'aucun_avis', label:'Message « Aucun avis »'},
    {key:'premier_avis', label:'Message « Soyez le premier… »'},
    {key:'evaluer', label:'Titre « Évaluez ce produit »'},
    {key:'votre_nom', label:'Champ « Votre nom »'},
    {key:'votre_avis', label:'Champ « Votre avis »'},
    {key:'envoyer_avis', label:'Bouton « Envoyer mon avis »'},
  ],
  panier:[
    {key:'sous_total', label:'Ligne « Sous-total »'},
    {key:'livraison', label:'Ligne « Livraison »'},
    {key:'livraison_offerte', label:'Valeur « Offerte » (livraison gratuite)'},
    {key:'total', label:'Ligne « Total »'},
    {key:'articles', label:'Mot après le nombre d’articles'},
    {key:'tiroir_fermer', label:'Tiroir — bouton « Fermer »'},
    {key:'tiroir_retirer', label:'Tiroir — lien « Retirer »'},
    {key:'tiroir_vider', label:'Tiroir — « Vider le panier »'},
    {key:'tiroir_continuer', label:'Tiroir — « Continuer mes achats »'},
    {key:'tiroir_livraison', label:'Tiroir — note de livraison'},
    {key:'toast_ajoute', label:'Message « Ajouté au panier : … »'},
    {key:'devis_nom', label:'Devis — champ nom'},
    {key:'devis_email', label:'Devis — champ e-mail'},
    {key:'devis_tel', label:'Devis — champ téléphone'},
    {key:'devis_ent', label:'Devis — champ entreprise'},
    {key:'devis_client', label:'Devis — question client'},
    {key:'devis_oui', label:'Devis — réponse « Oui »'},
    {key:'devis_non', label:'Devis — réponse « Non »'},
    {key:'devis_continuer', label:'Devis — bouton « Continuer »'},
    {key:'aide_titre', label:'Carte « Besoin d’aide ? » — titre'},
    {key:'aide_texte', label:'Carte « Besoin d’aide ? » — texte'},
    {key:'suite_titre', label:'Carte « Étapes suivantes » — titre'},
    {key:'suite_texte', label:'Carte « Étapes suivantes » — texte'},
    {key:'ph_texte', label:'Cadre « Photo à venir »'},
  ],
  compte:[
    {key:'cpt_onglet_conn', label:'Onglet « Connexion »'},
    {key:'cpt_onglet_crea', label:'Onglet « Créer un compte »'},
    {key:'cpt_titre', label:'Titre « DÉJÀ CLIENT ? »'},
    {key:'cpt_email', label:'Libellé e-mail'},
    {key:'cpt_mdp', label:'Libellé mot de passe'},
    {key:'cpt_oubli', label:'Lien « Mot de passe oublié ? »'},
    {key:'cpt_btn_conn', label:'Bouton « Connexion »'},
    {key:'cpt_civ', label:'Libellé civilité'},
    {key:'cpt_prenom', label:'Libellé prénom'},
    {key:'cpt_nom', label:'Libellé nom'},
    {key:'cpt_mdp2', label:'Libellé confirmation'},
    {key:'cpt_btn_crea', label:'Bouton « Créez un compte »'},
    {key:'cpt_sync', label:'Note de synchronisation'},
    {key:'pref_infos', label:'Menu — informations'},
    {key:'pref_cmd', label:'Menu — commandes'},
    {key:'pref_adr', label:'Menu — adresses'},
    {key:'pref_paniers', label:'Menu — paniers sauvegardés'},
    {key:'pref_params', label:'Menu — paramètres'},
    {key:'pref_requis', label:'Texte « COMPTE REQUIS »'},
    {key:'pref_theme', label:'Texte « thème et couleur »'},
  ],
  menu:[
    {key:'toutes_cat', label:'Barre « Toutes les Catégories »'},
    {key:'voir_tous', label:'Lien « Voir tous »'},
  ],
  header:[
    {key:'bienvenue', label:'Texte de la barre du haut'},
    {key:'suivi_commande', label:'Lien « Suivi de commande »'},
    {key:'inscription', label:'Lien « Inscription / Connexion »'},
    {key:'boutique_ico', label:'Icône « BOUTIQUE »'},
    {key:'compte_ico', label:'Icône « COMPTE »'},
    {key:'panier_ico', label:'Icône « PANIER »'},
    {key:'prefs_ico', label:'Icône « PRÉFÉRENCES »'},
  ],
  footer:[
    {key:'newsletter_titre', label:'Titre du bandeau newsletter'},
    {key:'newsletter_texte', label:'Texte du bandeau newsletter'},
    {key:'newsletter_ph', label:'Texte du champ e-mail'},
    {key:'newsletter_btn', label:'Bouton « S\'inscrire »'},
    {key:'copyright', label:'Texte du copyright'},
  ],
};

let txtPages = {
  apropos:'<h1>À propos de Mediatec</h1>\n<p>Mediatec est votre partenaire informatique à Abidjan.</p>\n<p>Ordinateurs, imprimantes, réseaux et consommables.</p>',
  cgv:'<h1>Conditions Générales de Vente</h1>\n<p>Les prix sont affichés en francs CFA, toutes taxes comprises.</p>\n<p>La livraison est offerte dès 100 000 F d\u2019achat à Abidjan.</p>',
  contact:'<h1>Contact</h1>\n<p>Plateau, Abidjan — Lun–Ven 8h00–17h30, Sam 8h30–13h00.</p>\n<p>Email : infos@mediatec-ci.com</p>',
  livraison:'<h1>Livraison</h1>\n<p>Livraison en 24–72h à Abidjan, offerte dès 100 000 F.</p>\n<p>Retrait gratuit au Plateau.</p>',
};

let txtPageSel = 'apropos';

let menu = {showVoirTous:true, showCount:true,
  fams:[
    {key:'reseaux', label:'Matériels réseaux', count:281, visible:true},
    {key:'accessoires', label:'Accessoires informatiques', count:211, visible:true},
    {key:'imprimantes', label:'Imprimantes', count:60, visible:true},
    {key:'laptops', label:'Ordinateurs portables', count:37, visible:true},
    {key:'bureau', label:'Ordinateurs de bureau', count:36, visible:true},
    {key:'serveurs', label:'Serveurs', count:19, visible:true},
    {key:'cameras', label:'Caméras & Vidéosurveillance', count:0, visible:false},
  ]};

let seo = {title:'Mediatec Store — Informatique & réseaux à Abidjan',
  desc:'Ordinateurs, imprimantes, réseaux et consommables — Mediatec, Abidjan.',
  og:'', index:true};

let reglages = {nom:'Mediatec Store', tel:'(+225) 07 05 08 75 07', email:'sales@mediatec-ci.com',
  wa:'2250705087507', adresse:'Plateau, Abidjan', horaires:'Lun–Ven 8h00–17h30 · Sam 8h30–13h00',
  maintenance:false, maintMsg:'Site en maintenance, revenez bientôt.'};

const FILES = {
  'theme.css': {lang:'css', path:'static/theme.css', content:
'/* Thème Mediatec Store — couleurs du site */\n:root{\n  --accent:#2E5136;\n  --price:#dc2626;\n  --bg:#ffffff;\n  --text:#111827;\n}\nbody{\n  font-family:Inter,system-ui,sans-serif;\n  background:var(--bg);\n  color:var(--text);\n}\n.pv-hero{border-radius:16px}\n.pv-card{box-shadow:0 2px 10px rgba(0,0,0,.08)}\n.btn{\n  background:var(--accent);\n  color:#fff;\n  border-radius:8px;\n}'},
  '_top.html': {lang:'html', path:'templates/static/_top.html', content:
'<header class="site-top">\n  <a class="logo" href="/">MEDIATEC STORE</a>\n  <form class="search" action="/recherche">\n    <input name="q" placeholder="Rechercher un produit…">\n    <button>Rechercher</button>\n  </form>\n</header>'},
  '_foot.html': {lang:'html', path:'templates/static/_foot.html', content:
'<footer class="site-foot">\n  <p>© Mediatec Store — Abidjan</p>\n  <p>Contact : sales@mediatec-ci.com · (+225) 07 05 08 75 07</p>\n  <nav>\n    <a href="/apropos.html">À propos</a> ·\n    <a href="/cgv.html">CGV</a>\n  </nav>\n</footer>\n<script src="/static/app.js?v=339"><\/script>'},
  'apropos.html': {lang:'html', path:'templates/static/apropos.html', content:
'<h1>À propos de Mediatec</h1>\n<p>Mediatec est votre partenaire informatique à Abidjan.</p>\n<p>Ordinateurs, imprimantes, réseaux et consommables.</p>'},
  'cgv.html': {lang:'html', path:'templates/static/cgv.html', content:
'<h1>Conditions Générales de Vente</h1>\n<p>Les prix sont affichés en francs CFA, toutes taxes comprises.</p>\n<p>La livraison est offerte dès 100 000 F d\u2019achat à Abidjan.</p>'},
  'contact.html': {lang:'html', path:'templates/static/contact.html', content:
'<h1>Contact</h1>\n<p>Plateau, Abidjan — Lun–Ven 8h00–17h30, Sam 8h30–13h00.</p>\n<p>Email : infos@mediatec-ci.com</p>\n<p>Tél : (+225) 07 05 08 75 07</p>'},
  'app.js': {lang:'js', path:'static/app.js', content:
'// Marexstore — app.js (v339)\nfunction applyTheme(){\n  const t = loadTheme();\n  document.documentElement.dataset.color = t.color || "default";\n}\ndocument.addEventListener("DOMContentLoaded", applyTheme);'},
};

let openTabs = ['theme.css'];

let currentFile = 'theme.css';

function esc(s){ return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

const MODULE_NAMES = {front:'Front — chrome du site', boutique:'Boutique', fiche:'Fiche produit', panier:'Panier & devis', menu:'Menu', accueil:'Page d’accueil', header:'En-tête', footer:'Pied de page'};

function T(k){return Tx[k]!==undefined?Tx[k]:''}

const PATHS={boutique:'/boutique',fiche:'/produit/hp-elitebook-850-g3',panier:'/panier',compte:'/compte',header:'/',footer:'/',menu:'/boutique',seo:'/',reglages:'/'};

let BASE=null,ORIGIN=null,pubSnap=null,hist=[],hi=0,restoring=false,histTimer=null;

const SEGS={segTitle:'front.titleSize',segBtn:'front.btnShape',segCard:'front.cardLayout',segShopLayout:'shop.layout',segShopPer:'shop.perPage',segShopPage:'shop.previewPage',segSimNb:'fiche.similarNb',segGallery:'fiche.gallery'};

const TABICO={front:'sparkle',accueil:'home',boutique:'bag',fiche:'file',panier:'cart',textes:'pencil',menu:'list',seo:'search',reglages:'gear',header:'monitor',footer:'list',compte:'lock'};

const snap=()=>JSON.stringify({front,rows,shop,fiche,panier,Tx,txtPages,menu,seo,reglages,header,footer,acc,files:Object.fromEntries(Object.entries(FILES).map(([k,v])=>[k,v.content]))});

function updateStatus(){
  if(!BASE||restoring)return;
  const dirty=snap()!==pubSnap;
  $('#status').classList.toggle('dirty',dirty); $('#statusTxt').textContent=dirty?'Brouillon — non publié':'Publié';
  clearTimeout(histTimer); histTimer=setTimeout(()=>{pushHist();histBtns()},350);
}

function pushHist(){const s=snap(); if(s===hist[hi])return; hist=hist.slice(0,hi+1); hist.push(s); if(hist.length>60)hist.shift(); hi=hist.length-1}

function histBtns(){$('#btnUndo').disabled=hi<=0;$('#btnRedo').disabled=hi>=hist.length-1}

function restore(str){
  const o=JSON.parse(str); restoring=true;
  ({front,rows,shop,fiche,panier,txtPages,menu,seo,reglages,header,footer,acc}=o);
  /* Tx est un const : on le vide puis on recopie les textes restaurés */
  Object.keys(Tx).forEach(k=>delete Tx[k]);
  Object.assign(Tx, o.Tx);
  for(const k in o.files)FILES[k].content=o.files[k];
  sel={ri:null,ci:null};$('#blockEditor').classList.add('hidden');
  syncUI(); restoring=false; renderPreview(); histBtns();
}

function undo(){clearTimeout(histTimer);pushHist();if(hi>0){hi--;restore(hist[hi])}}

function redo(){if(hi<hist.length-1){hi++;restore(hist[hi])}}

const CATEGORIES=[
 {id:'apparence',label:'Apparence',mods:['front','header','footer']},
 {id:'pages',label:'Pages',mods:['accueil','boutique','fiche','panier','compte']},
 {id:'contenu',label:'Contenu',mods:['textes','menu','seo']},
 {id:'reglages',label:'Réglages',mods:['reglages']}];

const TG={
 front:[['En-tête et recherche',['bienvenue','recherche_ph','rechercher','toutes_cat']],['Bandeau et boutons',['bandeau','acheter']]],
 boutique:[['Filtres',['filtres','tri_label','ref_label','ref_ph','appliquer','stock_only','reset_filtres','plus_filtres','dispo_label','etat_label']],
  ['Cartes produits',['boutique_titre','btn_ajouter','en_stock','rupture','produits_suffix','produits_trouves','livraison_mention']],
  ['Page « Nos catégories »',['categories_titre','categories_sub','voir_tous_produits']],
  ['Colonne latérale',['sidebar_avant','sidebar_ventes','sidebar_offres']]],
 fiche:[['Bloc d\u2019achat',['ajouter_panier','devis','wa','fiche_ref','en_stock','rupture']],['Sections de la page',['specs','similaires','avis','conso']]],
 header:[['Barre du haut',['bienvenue','suivi_commande','inscription']],['Icônes',['boutique_ico','compte_ico','panier_ico','prefs_ico']]],
 footer:[['Newsletter',['newsletter_titre','newsletter_texte','newsletter_ph','newsletter_btn']],['Copyright',['copyright']]]
};

const grpOpen=new Set();

const getP=(o,p)=>p.split('.').reduce((a,k)=>a==null?a:a[k],o);

const setP=(o,p,v)=>{const k=p.split('.'),l=k.pop();getP(o,k.join('.')||'')===undefined;const t=k.length?getP(o,k.join('.')):o;t[l]=v};

const root={acc:()=>acc,fiche:()=>fiche};

const ACC_D=[['g','Hero (carrousel)'],['b','Afficher le carrousel','acc','hero.show'],['b','Flèches','acc','hero.arrows'],['b','Points de pagination','acc','hero.dots'],['b','Panneau blanc sous le hero','acc','hero.panel'],
 ['g','Cartes avantages'],['b','Afficher la rangée','acc','av.show'],
 ['g','Trouver un produit'],['b','Afficher le bloc','acc','find.show'],['b','Carte « Offre spéciale »','acc','find.offer'],
 ['g','Onglets produits'],['b','Onglet 1','acc','tabs.show.0'],['b','Onglet 2','acc','tabs.show.1'],['b','Onglet 3','acc','tabs.show.2'],
 ['g','Sections'],['b','NOS PRODUITS','acc','secs.0.show'],['b','LES PLUS VENDUS','acc','secs.1.show'],['b','AJOUTÉ RÉCEMMENT','acc','secs.2.show'],['b','MIS EN AVANT','acc','secs.3.show'],
 ['g','Bandeau sombre et marques'],['b','Bandeau sombre','acc','dark.show'],['b','Ligne des marques','acc','brands.show']];

const ACC_T=[['g','Hero — panneau blanc'],['t','Titre du panneau','acc','hero.panelT'],['t','Texte du panneau','acc','hero.panelX']]
 .concat([0,1,2].flatMap(i=>[['g','Hero — diapositive '+(i+1)],['t','Titre','acc','hero.slides.'+i+'.t'],['t','Texte','acc','hero.slides.'+i+'.x'],['t','Bouton','acc','hero.slides.'+i+'.b'],['t','Image (URL)','acc','hero.slides.'+i+'.img']]))
 .concat([0,1,2,3].flatMap(i=>[['g','Carte avantage '+(i+1)],['t','Icône','acc','av.cards.'+i+'.i'],['t','Titre','acc','av.cards.'+i+'.t'],['t','Texte','acc','av.cards.'+i+'.x']]))
 .concat([['g','Trouver un produit'],['t','Titre','acc','find.title'],['t','Liste 1','acc','find.l1'],['t','Liste 2','acc','find.l2'],['t','Liste 3','acc','find.l3'],['t','Bouton « Réinitialiser »','acc','find.reset'],['t','Bouton « Rechercher »','acc','find.search'],['t','Offre spéciale — titre','acc','find.offT'],['t','Offre spéciale — texte','acc','find.offX'],
  ['g','Onglets produits'],['t','Onglet 1','acc','tabs.labels.0'],['t','Onglet 2','acc','tabs.labels.1'],['t','Onglet 3','acc','tabs.labels.2'],
  ['g','Titres de sections']]).concat([0,1,2,3].flatMap(i=>[['t','Titre '+(i+1),'acc','secs.'+i+'.t'],['t','Sous-titre '+(i+1),'acc','secs.'+i+'.s']]))
 .concat([['g','Bandeau sombre'],['t','Titre','acc','dark.t'],['t','Texte','acc','dark.x'],['t','Bouton','acc','dark.b'],['g','Marques'],['a','Marques (une par ligne, l\u2019ordre = l\u2019ordre affiché)','acc','brands.list']]);

const FICHE_D=[['g','Bloc d\u2019achat — afficher / masquer'],['b','Badge de remise','fiche','showDiscount'],['b','Ancien prix barré','fiche','showOld'],['b','Étoiles','fiche','showStars'],['b','Livraison Express','fiche','showExpress'],['b','Garantie','fiche','showGarantie'],['b','TVA incluse','fiche','showTva'],['b','Partage','fiche','showShare'],['b','Bouton « Aimer » (cœur)','fiche','showLike'],['b','Sélecteur de quantité','fiche','showQty']];

const PAN_D=[['g','Aperçu'],['s','Afficher dans l\u2019aperçu','acc','cartView',[['drawer','Tiroir panier ouvert'],['devis','Page devis']]],
 ['g','Tiroir panier — éléments'],['b','Bouton « Vider le panier »','acc','drawer.vider'],['b','Bouton « Continuer mes achats »','acc','drawer.continuer'],['b','Note de livraison WhatsApp','acc','drawer.livraison'],
 ['g','Devis — cartes latérales'],['b','Cartes « Besoin d\u2019aide ? » et « Étapes suivantes »','acc','help.show'],
 ['g','Divers'],['b','Bouton flottant WhatsApp','acc','misc.wa'],['c','Couleur des badges de remise','acc','misc.badge']];

const CPT_D=[['g','Aperçu'],['s','Afficher dans l\u2019aperçu','acc','compteView',[['modal','Fenêtre connexion / inscription'],['prefs','Page Préférences']]],
 ['g','Visibilité'],['b','Lien « Mot de passe oublié ? »','acc','cpt.oubli'],['b','Note de synchronisation','acc','cpt.sync']];

const THEMES_V13=[
 {id:'Liquid',title:'Liquid',accent:'#2E5136',price:'#dc2626',bg:'#ffffff',text:'#111827',font:'Inter'},
 {id:'Sobre',title:'Sobre',accent:'#334155',price:'#b45309',bg:'#f8fafc',text:'#0f172a',font:'Open Sans'},
 {id:'Nuit',title:'Nuit',accent:'#60a5fa',price:'#fca5a5',bg:'#111827',text:'#f9fafb',font:'Manrope'},
 {id:'Chaud',title:'Chaud',accent:'#c2410c',price:'#9f1239',bg:'#fff7ed',text:'#431407',font:'DM Sans'}
];

let publishedHistoryV13=[];

const MENU_SUB_ITEMS_V14=[
  {key:'ordinateurs',base:'mega_ordinateurs',label:'Ordinateurs'},
  {key:'accessoires',base:'mega_accessoires',label:'Accessoires'},
  {key:'consommables',base:'mega_consommables',label:'Consommables'}
];

/* État v16 par défaut (déplacé ici depuis controller.js : l'alias studioV16
   ci-dessous le lit au chargement de model.js, avant tout rendu). */
if(!front._studioV16)front._studioV16={style:{radius:'soft',shadow:'light',density:'normal',buttons:'solid',titleFont:''},identity:{logo:'',favicon:'',share:'',shareTitle:'Mediatec Store',shareDesc:'Informatique, réseaux et services à Abidjan'},dark:{enabled:false,defaultMode:'light',preview:null,showToggle:false,bg:'#111827',text:'#f9fafb',card:'#1f2937',border:'#374151'},animations:'soft',pages:{about:{showHero:true,showLead:true,showBody:true,bg:'#ffffff'},contact:{showForm:true,showLead:true,showBody:true,bg:'#ffffff'},cgv:{showLead:true,showBody:true,bg:'#ffffff'},notFound:{showLead:true,showBody:true,bg:'#ffffff'},maintenance:{showLead:true,showBody:true,bg:'#ffffff'},cookies:{show:false,showLead:true,showBody:true,bg:'#1f2937',position:'bottom'},livraison:{showLead:true,showBody:true,bg:'#ffffff'},confidentialite:{showLead:true,showBody:true,bg:'#ffffff'},legal:{showLead:true,showBody:true,bg:'#ffffff'},apropos:{showLead:true,showBody:true,bg:'#ffffff'}}};

let studioV16=front._studioV16;

let pagesInfoV16Active='about';

const PAGE_INFO_V16=[
 {id:'about',name:'Qui sommes-nous ?',path:'/qui-sommes-nous.html',title:'page_about_title',lead:'page_about_lead',body:'page_about_body',titleText:'Qui sommes-nous ?',leadText:'LE GROUPE MEDIATEC',bodyText:'Fondée en 2004, MEDIATEC a évolué pour devenir un acteur clé de la transformation digitale en Côte d’Ivoire et au-delà. Nous accompagnons les entreprises dans leur transformation numérique en proposant des solutions innovantes, adaptées et spécifiques, avec une mission : rendre la technologie accessible à tous.\n\n20 années d’expérience · 2004 année de fondation · 16 clients majeurs nous font confiance.\n\nDécouvrez le Groupe : Nos services · Infrastructure IT · Collaboration · Cloud & Sécurité · Nos références · À propos de nous · Contactez-nous.\n\nBesoin d’une solution IT sur mesure ? Nos experts vous accompagnent dans tous vos projets IT.'},
 {id:'apropos',name:'À propos de nous',path:'/apropos.html',title:'page_apropos_title',lead:'page_apropos_lead',body:'page_apropos_body',titleText:'À propos de nous',leadText:'La Tech au service de votre croissance',bodyText:'Adoptez des solutions technologiques pour booster votre performance et assurer une croissance durable.\n\n20 années d’expérience · 2004 année de fondation · 7 partenaires technologiques.\n\nNotre Mission — Rendre la Technologie Accessible à Tous\nNous nous engageons à rendre les technologies numériques accessibles aux PME et PMI, en leur offrant des solutions sur mesure adaptées à leurs besoins et à leurs budgets. Nous accompagnons nos clients à chaque étape de leur transformation digitale afin qu’ils tirent le meilleur parti de leurs investissements technologiques.\n\nNos Valeurs — Professionnalisme, Partage, Proactivité, Innovation, Diversité.\n\nNotre Histoire — Perfectionnement continuel pour mieux vous servir. Fondée en 2004, MEDIATEC a évolué pour devenir un acteur clé de la transformation digitale en Côte d’Ivoire et au-delà.\n\nPourquoi choisir notre expertise ? Nous accompagnons les entreprises dans leur transformation numérique en proposant des solutions innovantes, adaptées et spécifiques.\n\nNos solutions couvrent l’infrastructure IT, la sécurité cloud et la collaboration.\n\nNos partenaires : Cisco · Logitech · Dell · Ncomputing · HP · Kaspersky · Symantec.\n\nIls nous font confiance : Paul Kouassi (BNI) · Aminata Sylla (RTI).\n\nPrêt à transformer votre entreprise ? Remplissez le formulaire de devis, et notre équipe vous contactera rapidement.'},
 {id:'contact',name:'Contact',path:'/contact.html',title:'page_contact_title',lead:'page_contact_lead',body:'page_contact_body',titleText:'Contact',leadText:'Parlons de votre projet.',bodyText:'Un devis, un produit ASUS, un problème technique ou un partenariat : écrivez-nous, notre équipe vous répond.\n\nAdresse : Avenue Franchet d’Espérey, Plateau — Abidjan, Côte d’Ivoire.\nTéléphone : +225 27 20 24 33 19.\nE-mail : infos@mediatec-ci.com.\nHoraires : Lundi–vendredi 8h00–17h30 · Samedi 8h30–13h00.',form:{title:'page_contact_form_title',notice:'page_contact_form_notice',name:'page_contact_form_name',email:'page_contact_form_email',phone:'page_contact_form_phone',subject:'page_contact_form_subject',message:'page_contact_form_message',submit:'page_contact_form_submit'},formTexts:{title:'Envoyez-nous un message',notice:'Les champs marqués d’un * sont obligatoires. Réponse sous 24 h ouvrées.',name:'Nom complet *',email:'Adresse e-mail *',phone:'Téléphone',subject:'Objet de votre demande *',message:'Votre message *',submit:'Envoyer le message'}},
 {id:'cgv',name:'Conditions générales',path:'/cgv.html',title:'page_cgv_title',lead:'page_cgv_lead',body:'page_cgv_body',titleText:'Conditions générales de vente',leadText:'Conditions générales de vente',bodyText:'1. Prix et disponibilité\nLes prix sont affichés en francs CFA (F CFA), toutes taxes comprises. La disponibilité indiquée sur chaque fiche (« En stock », « Rupture de stock ») est donnée à titre indicatif : elle est confirmée par notre équipe avant toute commande.\n\n2. Commande\nLa commande se fait depuis le panier, par WhatsApp, ou par demande de devis. Elle n’est définitive qu’après confirmation de notre équipe (disponibilité, prix et délais).\n\n3. Paiement\nLe paiement s’effectue en espèces, par virement bancaire ou par Mobile Money, selon les modalités convenues avec notre équipe au moment de la confirmation de commande.\n\n4. Livraison\nLivraison à Abidjan et dans toute la Côte d’Ivoire. Les délais et les frais de livraison sont confirmés avec notre équipe avant la commande (voir Livraison & retours).\n\n5. Garantie\nLes produits sont couverts par la garantie du fabricant, selon ses conditions. Les consommables (toners, cartouches) ne sont pas couverts par la garantie une fois ouverts.\n\n6. Retours\nTout retour doit être signalé sous 48 h après réception, produit non utilisé et dans son emballage d’origine. Les frais de retour sont à la charge du client sauf erreur de notre part.\n\n7. Contact\nMEDIATEC — Abidjan, Côte d’Ivoire — Tél : (+225) 07 05 08 75 07 / (+225) 27 20 24 33 19 — E-mail : sales@mediatec-ci.com.\n\n8. Propriété intellectuelle\nLe contenu du site (photos, prix, descriptions, sélection des produits) est la propriété de MEDIATEC. Toute copie ou réutilisation sans autorisation écrite est interdite.'},
 {id:'notFound',name:'Page 404',path:'/404',title:'page_404_title',lead:'page_404_lead',body:'page_404_body',titleText:'Page introuvable',leadText:'La page recherchée n’existe pas ou a été déplacée.',bodyText:'Vérifiez l’adresse ou retournez à l’accueil.'},
 {id:'maintenance',name:'Maintenance (maquette)',path:'/maintenance',title:'page_maintenance_title',lead:'page_maintenance_lead',body:'page_maintenance_body',titleText:'Écran de maintenance (maquette)',leadText:'Aucune page de maintenance publique n’a été trouvée.',bodyText:'Le chemin /maintenance.html renvoie une erreur 404 sur le site actuel. Un écran de maintenance nécessitera une configuration d’hébergement.'},
 {id:'cookies',name:'Bandeau cookies (maquette)',path:'/',title:'page_cookies_title',lead:'page_cookies_lead',body:'page_cookies_body',titleText:'Bandeau cookies (maquette)',leadText:'Aucun bandeau cookies n’est apparu sur les pages consultées.',bodyText:'La maquette permet de préparer son apparence ; la gestion du consentement doit rester contrôlée par le système de consentement réel.',buttons:['page_cookies_accept','page_cookies_reject','page_cookies_settings'],buttonTexts:['Accepter','Refuser','Personnaliser']},
 {id:'livraison',name:'Livraison & retours',path:'/livraison.html',title:'page_livraison_title',lead:'page_livraison_lead',body:'page_livraison_body',titleText:'Livraison & retours',leadText:'Livraison & retours',bodyText:'Zones de livraison\nNous livrons à Abidjan et dans toute la Côte d’Ivoire. Le retrait en boutique est également possible (Plateau, Abidjan).\n\nDélais et frais\nLes délais et les frais de livraison dépendent de la destination et du volume de la commande : ils sont confirmés par notre équipe sur WhatsApp avant la commande.\n\nRetours\nSignalez tout retour sous 48 h après réception, produit non utilisé et dans son emballage d’origine. Contact : (+225) 07 05 08 75 07 / (+225) 27 20 24 33 19 — sales@mediatec-ci.com.'},
 {id:'confidentialite',name:'Politique de confidentialité',path:'/confidentialite.html',title:'page_privacy_title',lead:'page_privacy_lead',body:'page_privacy_body',titleText:'Politique de confidentialité',leadText:'Données collectées',bodyText:'Données collectées\nNous collectons uniquement les informations que vous nous transmettez : nom, e-mail, téléphone et contenu de vos messages (demande de devis, contact, newsletter, suivi de commande).\n\nUtilisation\nVos données servent uniquement à traiter vos demandes et à vous répondre. Elles ne sont ni vendues ni partagées avec des tiers.\n\nPanier\nVotre panier est enregistré uniquement dans votre navigateur (stockage local) : il n’est transmis à personne et vous pouvez le vider à tout moment.\n\nVos droits\nVous pouvez demander la correction ou la suppression de vos données à tout moment : sales@mediatec-ci.com — (+225) 07 05 08 75 07.'},
 {id:'legal',name:'Mentions légales',path:'/mentions-legales.html',title:'page_legal_title',lead:'page_legal_lead',body:'page_legal_body',titleText:'Mentions légales',leadText:'Éditeur du site',bodyText:'Éditeur du site\nMEDIATEC — Abidjan, Côte d’Ivoire.\nTél : (+225) 07 05 08 75 07 / (+225) 27 20 24 33 19 — E-mail : sales@mediatec-ci.com.\n\nObjet du site\nCatalogue en ligne des produits et services de MEDIATEC. Les prix et disponibilités sont donnés à titre indicatif et confirmés par notre équipe avant toute commande.\n\nPropriété intellectuelle\nLes marques citées (HP, Dell, Lenovo, Kyocera, Epson…) appartiennent à leurs propriétaires respectifs. Le contenu du site — textes, sélection et présentation des produits, prix, photos (dont celles marquées « store.neva-ci.pro »), mise en page et code — est la propriété exclusive de MEDIATEC. Toute reproduction, représentation ou extraction, même partielle, du catalogue (photos, prix, descriptions) par quelque moyen que ce soit, sans autorisation écrite préalable, est interdite et passible de poursuites.'}
];

function v16DiffLeaves(a,b,path=''){
 if(JSON.stringify(a)===JSON.stringify(b))return[];
 if(a&&b&&typeof a==='object'&&typeof b==='object'&&!Array.isArray(a)&&!Array.isArray(b)){
  const out=[];for(const k of new Set([...Object.keys(a),...Object.keys(b)]))out.push(...v16DiffLeaves(a[k],b[k],path?path+'.'+k:k));return out
 }
 return[{path,before:b,after:a}];
}

function v16GetPath(obj,path){return path.split('.').reduce((v,k)=>v==null?undefined:v[k],obj)}

function v16SetPath(obj,path,value){const parts=path.split('.'),last=parts.pop(),host=parts.reduce((v,k)=>v[k],obj);host[last]=value}

