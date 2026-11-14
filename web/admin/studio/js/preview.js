/* preview.js — Studio Mediatec (MVC). Généré par découpe de studio-prototype-v16-2.html. */

function pvHeaderHTML(){
  const f = front, h = header;
  const topbar = h.showTopbar
    ? `<div class="pv-top"><span>${esc(T('bienvenue'))}</span><span>${esc(T('suivi_commande'))}${h.showPhone?' · '+esc(reglages.tel):''}${h.showEmail?' · '+esc(reglages.email):''} · ${esc(T('inscription'))}</span></div>` : '';
  const ico = (show, emoji, label)=> show
    ? `<div style="text-align:center;font-size:10px;font-weight:700;color:${f.text};line-height:1.3">${emoji}<br>${esc(label)}</div>` : '';
  const icons = [
    ico(h.showBoutique, '🛍️', T('boutique_ico')),
    ico(h.showCompte, '👤', T('compte_ico')),
    ico(h.showPanier, '🛒', T('panier_ico')),
    ico(h.showPrefs, '⚙️', T('prefs_ico')),
  ].filter(Boolean).join('');
  const navLinks = h.navLinks.filter(l=>l.visible).map(l=>`<span style="margin-left:16px">${esc(l.label)}</span>`).join('');
  return `${topbar}
    <div class="pv-head">
      <div class="pv-logo" style="color:${f.accent}">${esc(reglages.nom).toUpperCase()}</div>
      <div class="pv-search"><input value="" placeholder="${esc(T('recherche_ph'))}"><button style="background:${f.accent};border-radius:0 ${btnRadius()} ${btnRadius()} 0">${esc(T('rechercher'))}</button></div>
      <div style="display:flex;gap:14px;align-items:center">${icons}</div>
    </div>
    <div class="pv-menu" style="background:${h.navBg}">☰ ${esc(T('toutes_cat'))}${navLinks}</div>`;
}

function pvFooterHTML(){
  const f = front, ft = footer;
  const news = ft.showNewsletter
    ? `<div style="background:${header.navBg};color:#fff;padding:14px 16px;text-align:center"><strong>${esc(T('newsletter_titre'))}</strong> — ${esc(T('newsletter_texte'))} ${esc(reglages.tel)} · ${esc(reglages.email)}<br><span style="display:inline-flex;gap:6px;margin-top:8px"><input value="" placeholder="${esc(T('newsletter_ph'))}" style="padding:7px 10px;border:0;border-radius:6px;font-size:12px"><button style="background:#fff;color:${header.navBg};border:0;border-radius:6px;padding:7px 14px;font-weight:700;cursor:pointer">${esc(T('newsletter_btn'))}</button></span></div>` : '';
  const cols = ft.showCols
    ? `<div style="display:grid;grid-template-columns:repeat(5,1fr);gap:10px;padding:14px 16px;background:#fff;border-top:1px solid #e5e7eb">${ft.cols.filter(c=>c.visible).map(c=>`<div><strong style="font-size:12px">${esc(c.title)}</strong>${c.links.map(l=>`<div style="font-size:11px;color:#6b7280;margin-top:4px">${esc(l)}</div>`).join('')}</div>`).join('')}</div>` : '';
  return `${news}${cols}
    <div class="pv-foot" style="background:${f.footerBg};color:#fff">${esc(T('copyright'))}</div>`;
}

function renderPreview(){
  updateStatus(); setPath();
  /* L'aperçu suit l'onglet actif : chaque module montre sa vraie page. */
  if(currentTab==='boutique') return pvShop();
  if(currentTab==='fiche') return pvFiche();
  if(currentTab==='panier') return pvPanier();
  if(currentTab==='textes') return pvTextes();
  if(currentTab==='menu') return pvMenu();
  if(currentTab==='seo') return pvSeo();
  if(currentTab==='reglages') return pvReglages();
  if(currentTab==='header') return pvHeader();
  if(currentTab==='footer') return pvFooter();
  if(currentTab==='compte') return pvCompte();
  return pvHome();
}

function pvHome(){
  const f = front;
  const annonce = f.showAnnonce
    ? `<div style="background:${f.accent};color:#fff;text-align:center;font-size:12px;font-weight:700;padding:7px 10px">${esc(f.banner||'')}</div>` : '';
  const customHtml = f.customHTML
    ? `<div class="pv-custom-html" style="background:${f.accent}22;color:${f.accent}">${f.customHTML}</div>` : '';
  const body = rows.map(r=>{
    const cols = r.cols.map(c=>{
      const b = c.block;
      if(b.type==='Bannière'){
        return `<div class="pv-col" style="flex:0 0 ${c.w}%">
          <div class="pv-hero" style="background:linear-gradient(135deg,${f.accent},#0f2a18);border-radius:${btnRadius()}">
            <h4 style="font-size:${titleSize()}">${esc(b.title)}</h4><p>${esc(b.text)}</p>
            <div class="pv-price">${esc(b.price||'')}</div>
            ${b.cta?`<span class="pv-cta" style="color:${f.accent};border-radius:${btnRadius()}">${esc(b.cta)}</span>`:''}
          </div></div>`;
      }
      const badge = f.showBadge ? `<span class="pv-badge">PROMO</span><br>` : '';
      if(f.cardLayout==='liste'){
        return `<div class="pv-col" style="flex:0 0 ${c.w}%">
          <div class="pv-card" style="display:flex;gap:10px;text-align:left;align-items:center">
            <div class="thumb" style="width:64px;height:56px;flex:0 0 64px;margin:0">${b.img||'📦'}</div>
            <div>${badge}<h5>${esc(b.title)}</h5>
            <p style="margin:0;color:#6b7280;font-size:11px">${esc(b.text)}</p>
            <div class="pr" style="color:${f.priceColor}">${esc(b.price||'')}</div></div>
          </div></div>`;
      }
      return `<div class="pv-col" style="flex:0 0 ${c.w}%">
        <div class="pv-card">${badge}<div class="thumb">${b.img||'📦'}</div>
        <h5>${esc(b.title)}</h5><p style="margin:0;color:#6b7280;font-size:11px">${esc(b.text)}</p>
        <div class="pr" style="color:${f.priceColor}">${esc(b.price||'')}</div></div>
      </div>`;
    }).join('');
    return `<div class="pv-row">${cols}</div>`;
  }).join('');
  PV.html = `
    ${customHtml}
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">${body || '<p style="color:#6b7280;text-align:center">Aucune ligne — ajoutez-en une.</p>'}</div>
      ${pvFooterHTML()}
    </div>`;
}

function pvShop(){
  const f = front;
  const facets = shop.facets.filter(x=>x.visible);
  const cards = [
    {img:'💻', title:'HP EliteBook 850 G3', price:'185 000 F', badge:true},
    {img:'🖨️', title:'Kyocera ECOSYS M5526cdw', price:'650 000 F', badge:false},
    {img:'🌐', title:'Cisco Catalyst 9300-24P', price:'1 250 000 F', badge:true},
    {img:'🔌', title:'APC Easy UPS 3S 15 kVA', price:'10 450 185 F', badge:false},
  ];
  const card = c=>{
    if(shop.layout==='miniatures'){
      return `<div class="pv-card" style="text-align:center;padding:8px">
        <div class="thumb" style="height:56px;font-size:28px;margin:0">${c.img}</div>
        <h5 style="font-size:11px;margin:6px 0 2px">${c.title}</h5>
        <div class="pr" style="color:${f.priceColor};font-size:12px">${c.price}</div>
      </div>`;
    }
    return `<div class="pv-card" style="${shop.layout==='liste'?'display:flex;gap:10px;text-align:left;align-items:center':''}">
      ${shop.showBadge && c.badge ? '<span class="pv-badge">PROMO</span><br>' : ''}
      <div class="thumb" style="${shop.layout==='liste'?'width:64px;height:56px;flex:0 0 64px;margin:0':''}">${c.img}</div>
      <div><h5>${c.title}</h5>
      <div class="pr" style="color:${f.priceColor}">${c.price}</div>
      ${shop.showStock?`<div style="font-size:11px;color:#059669">${esc(T('en_stock'))}</div>`:''}
      ${shop.quickAdd?`<button style="margin-top:6px;background:${f.accent};color:#fff;border:0;border-radius:${btnRadius()};padding:6px 10px;font-size:12px;font-weight:700;cursor:pointer">${esc(T('btn_ajouter'))}</button>`:''}
      </div>
    </div>`;
  };
  const grid = shop.layout==='liste'
    ? `<div style="display:grid;gap:10px">${cards.map(card).join('')}</div>`
    : shop.layout==='miniatures'
    ? `<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:8px">${cards.map(card).join('')}</div>`
    : `<div style="display:grid;grid-template-columns:repeat(2,1fr);gap:12px">${cards.map(card).join('')}</div>`;
  const sortLabel = shop.sort==='stock'?'Stock d\u2019abord':shop.sort==='prix-asc'?'Prix croissant':shop.sort==='prix-desc'?'Prix décroissant':shop.sort==='nom-az'?'Nom A-Z':shop.sort==='nom-za'?'Nom Z-A':'Nouveautés';
  const perPageLabel = shop.perPage==='all' ? 'Tout afficher' : shop.perPage + ' par page';
  /* Colonne latérale droite (v9) */
  const sidebar = shop.showSidebar ? `<div style="border:1px solid #e5e7eb;border-radius:10px;padding:10px;background:#fff;width:150px;flex:0 0 150px">
    ${shop.sidebarAvant?`<div style="margin-bottom:12px"><strong style="font-size:12px">${esc(T('sidebar_avant'))}</strong><div style="font-size:11px;color:#6b7280;margin-top:4px">HP EliteBook 850 G3 — 185 000 F</div></div>`:''}
    ${shop.sidebarVentes?`<div style="margin-bottom:12px"><strong style="font-size:12px">${esc(T('sidebar_ventes'))}</strong><div style="font-size:11px;color:#6b7280;margin-top:4px">Kyocera M5526cdw — 650 000 F</div></div>`:''}
    ${shop.sidebarOffres?`<div><strong style="font-size:12px">${esc(T('sidebar_offres'))}</strong><div style="font-size:11px;color:#6b7280;margin-top:4px">Cisco Catalyst 9300 — 1 250 000 F</div></div>`:''}
  </div>` : '';
  /* Filtres latéraux enrichis (v9) */
  const filtersBox = `
    <div style="border:1px solid #e5e7eb;border-radius:10px;padding:10px;background:#fff">
      <strong style="font-size:12px">${esc(T('filtres'))}</strong>
      ${shop.showRefSearch?`<div style="margin-top:8px"><div style="font-size:12px;font-weight:700">${esc(T('ref_label'))}</div><input value="" placeholder="${esc(T('ref_ph'))}" style="width:100%;padding:6px 8px;border:1px solid #e5e7eb;border-radius:6px;font-size:11px;margin-top:4px"></div>`:''}
      ${facets.length?facets.map(x=>`<div style="margin-top:8px;font-size:12px"><strong>${x.label}</strong><div style="color:#6b7280">${x.sub}</div></div>`).join(''):'<p style="font-size:12px;color:#6b7280">Aucune facette affichée.</p>'}
      ${shop.showPriceSlider?`<div style="margin-top:8px;font-size:12px"><strong>Prix</strong><div style="display:flex;gap:4px;margin-top:4px"><input value="" placeholder="${esc(T('prix_min'))}" style="width:50%;padding:5px 6px;border:1px solid #e5e7eb;border-radius:6px;font-size:11px"><input value="" placeholder="${esc(T('prix_max'))}" style="width:50%;padding:5px 6px;border:1px solid #e5e7eb;border-radius:6px;font-size:11px"></div><button style="margin-top:6px;background:${f.accent};color:#fff;border:0;border-radius:6px;padding:5px 10px;font-size:11px;font-weight:700;cursor:pointer">${esc(T('appliquer'))}</button></div>`:''}
      ${shop.showStockOnly?`<label style="display:block;margin-top:8px;font-size:12px"><input type="checkbox"> ${esc(T('stock_only'))}</label>`:''}
      ${shop.showMoreFilters?`<div style="margin-top:8px"><button style="background:none;border:0;color:${f.accent};font-weight:700;font-size:12px;cursor:pointer;padding:0">${esc(T('plus_filtres'))}</button><div style="font-size:12px;margin-top:4px"><strong>${esc(T('dispo_label'))}</strong><div style="color:#6b7280">En stock · Rupture de stock</div><strong>${esc(T('etat_label'))}</strong><div style="color:#6b7280">Neuf · Occasion · Reconditionné</div></div></div>`:''}
      <button style="margin-top:10px;background:none;border:1px solid #e5e7eb;border-radius:6px;padding:5px 10px;font-size:11px;cursor:pointer">${esc(T('reset_filtres'))}</button>
    </div>`;
  /* Page d'atterrissage « Nos catégories » (v9) */
  const landing = `
    <div class="pv-body">
      <h3 style="margin:0 0 4px">${esc(T('categories_titre'))}</h3>
      <p style="color:#6b7280;font-size:12px;margin:0 0 12px">${esc(T('categories_sub'))}</p>
      <p style="margin:0 0 12px"><a href="#" style="color:${f.accent};font-weight:700">${esc(T('voir_tous_produits'))} (663) →</a></p>
      <div style="display:grid;grid-template-columns:repeat(2,1fr);gap:10px">
        ${[['Ordinateurs','145'],['Imprimantes & Consommables','145'],['Moniteur & Écrans','55'],['Réseaux','50'],['Stockage','50'],['Serveurs','25']].map(([n,c])=>`<div style="border:1px solid #e5e7eb;border-radius:10px;padding:12px;background:#fff;text-align:center"><div style="font-size:24px">📦</div><strong style="font-size:13px">${n}</strong><div style="font-size:11px;color:#6b7280">${c} produits</div></div>`).join('')}
      </div>
    </div>`;
  const listPage = `
    <div class="pv-body">
      <h3 style="margin:0 0 4px">${esc(T('boutique_titre'))}</h3>
      <p style="color:#6b7280;font-size:12px;margin:0 0 12px">663 ${esc(T('produits_trouves'))} · ${esc(T('tri_label'))} ${sortLabel} · ${perPageLabel}</p>
      <div style="display:flex;gap:14px;align-items:start">
        <div style="width:150px;flex:0 0 150px">${filtersBox}</div>
        <div style="flex:1;min-width:0">${grid}</div>
        ${sidebar}
      </div>
    </div>`;
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      ${shop.previewPage==='categories' && shop.showLanding ? landing : listPage}
      ${pvFooterHTML()}
    </div>`;
}

function pvFiche(){
  const f = front;
  const sim = Array.from({length: fiche.similarNb}, (_,i)=>`
    <div class="pv-card" style="flex:1"><div class="thumb">${['💻','🖨️','🌐','🔌','🖥️','📡','⌨️','🖱️'][i]||'📦'}</div>
    <h5>Produit similaire ${i+1}</h5><div class="pr" style="color:${f.priceColor}">${(120+i*35)} 000 F</div></div>`).join('');
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <p style="color:#6b7280;font-size:12px;margin:0 0 10px">Accueil / Ordinateurs portables / HP EliteBook 850 G3</p>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;align-items:start">
          <div>
            <div style="border:1px solid #e5e7eb;border-radius:12px;background:#f3f4f6;height:180px;display:flex;align-items:center;justify-content:center;font-size:64px">💻</div>
            ${fiche.gallery==='vignettes'?'<div style="display:flex;gap:6px;margin-top:8px"><div style="width:44px;height:36px;border:2px solid '+f.accent+';border-radius:6px;background:#f3f4f6;display:flex;align-items:center;justify-content:center">💻</div><div style="width:44px;height:36px;border:1px solid #e5e7eb;border-radius:6px;background:#f3f4f6;display:flex;align-items:center;justify-content:center">⌨️</div><div style="width:44px;height:36px;border:1px solid #e5e7eb;border-radius:6px;background:#f3f4f6;display:flex;align-items:center;justify-content:center">🔌</div></div>':''}
          </div>
          <div>
            <h3 style="margin:0 0 4px;font-size:${titleSize()}">HP EliteBook 850 G3</h3>
            <p style="color:#6b7280;font-size:12px;margin:0 0 8px">${esc(T('fiche_ref'))} HP-850G3 · HP · ${esc(T('en_stock'))}</p>
            <div style="font-size:20px;font-weight:800;color:${f.priceColor};margin-bottom:12px">185 000 F${fiche.showOld?'<s style="font-size:13px;font-weight:500;color:#9ca3af;margin-left:8px">197 000 F</s>':''}${fiche.showDiscount?'<span class="pv-badge" style="margin-left:8px">-6%</span>':''}</div>${fiche.showStars?'<div style="color:#f59e0b;font-size:13px;margin:-6px 0 10px">★★★★☆</div>':''}
            <div style="display:flex;gap:8px;flex-wrap:wrap">
              <button style="background:${f.accent};color:#fff;border:0;border-radius:${btnRadius()};padding:10px 16px;font-weight:700;cursor:pointer">${esc(T('ajouter_panier'))}</button>
              ${fiche.showDevis?`<button style="background:#fff;color:${f.accent};border:1px solid ${f.accent};border-radius:${btnRadius()};padding:10px 16px;font-weight:700;cursor:pointer">${esc(T('devis'))}</button>`:''}
              ${fiche.showWa?`<button style="background:#25d366;color:#fff;border:0;border-radius:${btnRadius()};padding:10px 16px;font-weight:700;cursor:pointer">${esc(T('wa'))}</button>`:''}
            </div>
          ${ficheBuyExtras()}
          </div>
        </div>
        ${fiche.showSpecs?`<div style="margin-top:16px"><strong>${esc(T('specs'))}</strong>
          <table style="width:100%;border-collapse:collapse;margin-top:8px;font-size:12px">
            <tr><td style="border:1px solid #e5e7eb;padding:6px 8px;background:#f9fafb">Processeur</td><td style="border:1px solid #e5e7eb;padding:6px 8px">Intel Core i7-6500U</td></tr>
            <tr><td style="border:1px solid #e5e7eb;padding:6px 8px;background:#f9fafb">RAM / Stockage</td><td style="border:1px solid #e5e7eb;padding:6px 8px">16 Go / 512 Go SSD</td></tr>
            <tr><td style="border:1px solid #e5e7eb;padding:6px 8px;background:#f9fafb">Écran</td><td style="border:1px solid #e5e7eb;padding:6px 8px">15,6″ Full HD</td></tr>
          </table></div>`:''}
        ${ficheTail()}
        ${fiche.showConso?`<div style="margin-top:16px"><strong>${esc(T('conso'))}</strong><p style="font-size:13px;color:#6b7280">Chargeur HP 65 W · Sacoche 15,6″ · Câble HDMI</p></div>`:''}
        ${fiche.showSimilar?`<div style="margin-top:16px"><strong>${esc(T('similaires'))}</strong><div style="display:flex;gap:10px;margin-top:8px">${sim}</div></div>`:''}
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvPanier(){
  const f = front;
  const steps = panier.useSteps ? `
    <div style="margin:0 0 14px">
      <div class="stepper">${panier.steps.map((s,i)=>`
        <div style="display:flex;flex-direction:column;align-items:center;flex:0 0 26px">
          <div class="step-dot ${i===0?'':'off'}">${i+1}</div>
          <div class="step-lbl">${esc(s)}</div>
        </div>${i<panier.steps.length-1?`<div class="step-line ${i===0?'on':''}" style="margin-bottom:18px"></div>`:''}`).join('')}
      </div>
    </div>` : '';
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <h3 style="margin:0 0 4px">${esc(panier.title)}</h3>
        <p style="color:#6b7280;font-size:12px;margin:0 0 12px">2 ${esc(T('articles'))}</p>
        ${steps}
        <div style="border:1px solid #e5e7eb;border-radius:10px;padding:12px;background:#fff">
          <div class="cart-line"><div class="thumb">💻</div><div class="nm">HP EliteBook 850 G3<small>185 000 F × 1</small></div><span class="qty"><button>−</button><span>1</span><button>+</button></span></div>
          <div class="cart-line"><div class="thumb">🖨️</div><div class="nm">Kyocera ECOSYS M5526cdw<small>650 000 F × 1</small></div><span class="qty"><button>−</button><span>1</span><button>+</button></span></div>
          ${panier.showFees?`<div class="total-row"><span>${esc(T('sous_total'))}</span><span>835 000 F</span></div>
          <div class="total-row"><span>${esc(T('livraison'))}</span><span>${esc(T('livraison_offerte'))}</span></div>`:''}
          <div class="total-row big"><span>${esc(T('total'))}</span><span style="color:${f.priceColor}">835 000 F</span></div>
          <div style="display:flex;gap:8px;margin-top:12px;flex-wrap:wrap">
            <button style="flex:1;background:${f.accent};color:#fff;border:0;border-radius:${btnRadius()};padding:11px;font-weight:700;cursor:pointer">${esc(panier.btnOrder)}</button>
            <button style="flex:1;background:#25d366;color:#fff;border:0;border-radius:${btnRadius()};padding:11px;font-weight:700;cursor:pointer">${esc(panier.btnWa)}</button>
          </div>
        </div>
        <p style="color:#6b7280;font-size:12px;margin:10px 0 0">Panier vide : « ${esc(panier.emptyText)} »</p>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvTextes(){
  const f = front;
  const pageName = {apropos:'À propos', cgv:'CGV', contact:'Contact', livraison:'Livraison'}[txtPageSel] || txtPageSel;
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <p style="color:#6b7280;font-size:12px;margin:0 0 10px">Page : ${pageName}</p>
        <div style="border:1px solid #e5e7eb;border-radius:10px;padding:14px;background:#fff;line-height:1.6;font-size:13px">${txtPages[txtPageSel]||''}</div>
        <p style="color:#6b7280;font-size:12px;margin:10px 0 0">Bandeau : « ${esc(front.banner)} » · Bouton panier : « ${esc(T('ajouter_panier'))} »</p>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvMenu(){
  const f = front;
  const fams = menu.fams.filter(x=>x.visible);
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <div class="mega-prev">
          <div class="m-bar" style="background:${f.accent}">${esc(T('toutes_cat'))}</div>
          <div class="m-grid">
            ${fams.map(fm=>`<div class="m-group"><h5 style="color:${f.accent}">${fm.label}${menu.showCount?` <span style="color:#9ca3af;font-weight:400">(${fm.count})</span>`:''}</h5>
              <a href="#">Ordinateurs</a><a href="#">Accessoires</a><a href="#">Consommables</a>
              ${menu.showVoirTous?`<a href="#" style="color:${f.accent};font-weight:700">${esc(T('voir_tous'))} →</a>`:''}
            </div>`).join('')}
          </div>
        </div>
        <p style="color:#6b7280;font-size:12px;margin:10px 0 0">${fams.length} famille(s) affichée(s) · ${menu.fams.filter(x=>!x.visible).length} masquée(s).</p>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvSeo(){
  const f = front;
  const ogImg = seo.og
    ? `<img src="${esc(seo.og)}" style="width:100%;height:120px;object-fit:cover;display:block">`
    : `<div class="og-img" style="background:linear-gradient(135deg,${f.accent},#0f2a18)">🖼️</div>`;
  PV.html = `
    <div style="background:#fff;color:#111;font-family:system-ui,sans-serif;padding:16px">
      <h3 style="margin:0 0 4px">Résultat Google</h3>
      <p style="color:#6b7280;font-size:12px;margin:0 0 12px">${seo.index?'✅ Le site est indexé.':'⛔ Indexation désactivée — le site n\u2019apparaîtra pas dans Google.'}</p>
      <div class="g-preview">
        <div class="g-url">https://store.neva-ci.pro</div>
        <div class="g-title">${esc(seo.title)}</div>
        <div class="g-desc">${esc(seo.desc)}</div>
      </div>
      <h3 style="margin:16px 0 8px">Partage sur les réseaux (WhatsApp / Facebook)</h3>
      <div class="og-card">
        ${ogImg}
        <div class="og-body">
          <div class="og-domain">store.neva-ci.pro</div>
          <div class="og-title">${esc(seo.title)}</div>
          <div class="og-desc">${esc(seo.desc)}</div>
        </div>
      </div>
    </div>`;
}

function pvReglages(){
  const f = front;
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${reglages.maintenance?`<div class="maint-banner">⚠️ ${esc(reglages.maintMsg)}</div>`:''}
      <div class="pv-head">
        <div class="pv-logo" style="color:${f.accent}">${esc(reglages.nom).toUpperCase()}</div>
        <div class="pv-search"><input value="" placeholder="${esc(T('recherche_ph'))}"><button style="background:${f.accent}">${esc(T('rechercher'))}</button></div>
      </div>
      ${pvHeaderHTML()}
      <div class="pv-body">
        <h3 style="margin:0 0 8px">Coordonnées du site</h3>
        <div style="border:1px solid #e5e7eb;border-radius:10px;padding:14px;background:#fff;font-size:13px;line-height:1.9">
          <div><strong>Boutique :</strong> ${esc(reglages.nom)}</div>
          <div><strong>Téléphone :</strong> ${esc(reglages.tel)}</div>
          <div><strong>Email :</strong> ${esc(reglages.email)}</div>
          <div><strong>WhatsApp :</strong> ${esc(reglages.wa)}</div>
          <div><strong>Adresse :</strong> ${esc(reglages.adresse)}</div>
          <div><strong>Horaires :</strong> ${esc(reglages.horaires)}</div>
          <div><strong>Mode maintenance :</strong> ${reglages.maintenance?'Activé':'Désactivé'}</div>
        </div>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvHeader(){
  const f = front;
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <h3 style="margin:0 0 4px">En-tête du site</h3>
        <p style="color:#6b7280;font-size:12px;margin:0 0 12px">C'est l'en-tête qui apparaît en haut de toutes les pages. Modifiez la barre du haut, les icônes, la couleur de la navigation et les liens — l'aperçu suit en direct.</p>
        <div style="border:1px solid #e5e7eb;border-radius:10px;padding:12px;background:#fff;font-size:13px;line-height:1.9">
          <div><strong>Barre du haut :</strong> ${header.showTopbar?'Affichée':'Masquée'} · Tél : ${header.showPhone?'Oui':'Non'} · E-mail : ${header.showEmail?'Oui':'Non'}</div>
          <div><strong>Icônes :</strong> ${[header.showBoutique?'BOUTIQUE':null,header.showCompte?'COMPTE':null,header.showPanier?'PANIER':null,header.showPrefs?'PRÉFÉRENCES':null].filter(Boolean).join(' · ')||'Aucune'}</div>
          <div><strong>Liens de navigation :</strong> ${header.navLinks.filter(l=>l.visible).length} affichés sur ${header.navLinks.length}</div>
          <div><strong>En-tête fixe :</strong> ${header.sticky?'Oui':'Non'}</div>
        </div>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

function pvFooter(){
  const f = front;
  PV.html = `
    <div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">
      ${pvHeaderHTML()}
      <div class="pv-body">
        <h3 style="margin:0 0 4px">Pied de page du site</h3>
        <p style="color:#6b7280;font-size:12px;margin:0 0 12px">C'est le pied de page qui apparaît en bas de toutes les pages. Modifiez la newsletter, les colonnes et le copyright — l'aperçu suit en direct.</p>
        <div style="border:1px solid #e5e7eb;border-radius:10px;padding:12px;background:#fff;font-size:13px;line-height:1.9">
          <div><strong>Newsletter :</strong> ${footer.showNewsletter?'Affichée':'Masquée'}</div>
          <div><strong>Colonnes :</strong> ${footer.cols.filter(c=>c.visible).length} affichées sur ${footer.cols.length}</div>
        </div>
      </div>
      ${pvFooterHTML()}
    </div>`;
}

const FONTS='https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Manrope:wght@500;700&family=DM+Sans:wght@500;700&family=Poppins:wght@500;700&family=Roboto:wght@400;700&family=Open+Sans:wght@400;700&display=swap';

function ic(n,c='i'){return `<svg class="${c}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${ICONS[n]||''}</svg>`}

const EMO={'💻':'laptop','🖨':'printer','🌐':'globe','🔌':'plug','🔋':'battery','🗄':'server','🖥':'monitor','📡':'antenna','⌨':'keyboard','🖱':'mouse','📦':'box','✏':'pencil','🖼':'image','✅':'check-circle','⛔':'ban','⚠':'alert','🔍':'search','☰':'menu','🛍️':'bag','👤':'user','🛒':'cart','⚙️':'gear'};

const EMO_RE=new RegExp('('+Object.keys(EMO).join('|')+')\\uFE0F?','g');

const emojiToSvg=h=>h.replace(EMO_RE,(m,g)=>ic(EMO[g],'e'));

const pvI=$('#pvIframe');

const PV_RE=/^(\.pv-|\.cart-line|\.qty|\.total-row|\.stepper|\.step-(dot|line|lbl)|\.g-|\.og-|\.mega-prev|\.maint-banner)/;

function pvCss(){let c='';for(const sh of document.styleSheets){let rs;try{rs=sh.cssRules}catch(_){continue}for(const r of rs)if(r.selectorText&&PV_RE.test(r.selectorText.trim()))c+=r.cssText+'\n'}return c}

const PV_BASE=`:root{--line:#e5e7eb;--muted:#6b7280;--ink:#111827}*{box-sizing:border-box}body{margin:0;-webkit-font-smoothing:antialiased}img{max-width:100%}
svg.e{width:1em;height:1em;vertical-align:-.14em;stroke-width:1.6}.pv-menu svg.e{margin-right:.5em}.thumb svg.e{color:var(--accent)}
@media(max-width:520px){.pv-row{flex-direction:column}.pv-col{flex:0 0 auto!important;width:100%}[style*="150px 1fr"],[style*="grid-template-columns:1fr 1fr"]{grid-template-columns:1fr!important}}`;

const PV={set html(h){
  const d=pvI.contentDocument; if(!pvReady||!d){pvPending=h;return}
  pvPending=null;
  d.getElementById('theme').textContent=':root{--accent:'+front.accent+'}\n'+((typeof acc!=='undefined'&&acc.misc&&acc.misc.badge)?'.pv-badge{background:'+acc.misc.badge+'!important}\n':'')+(front.customCSS||'');
  d.getElementById('root').innerHTML=emojiToSvg(h+(typeof waBtn==='function'?waBtn():''));
  d.querySelectorAll('img').forEach(i=>i.onerror=()=>{i.style.display='none'});
  const l=$('#bLive'); l.classList.remove('flash'); void l.offsetWidth; l.classList.add('flash');
}};

function setDevice(m,b){$('#bStage').classList.toggle('mobile',m==='mobile');document.querySelectorAll('.devices button').forEach(x=>x.classList.toggle('active',x===b))}

function pvDevis(){
 const f=front,r=btnRadius(),st=panier.steps||[];
 PV.html=`<div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">${pvHeaderHTML()}<div class="pv-body">
  <h3 style="margin:0 0 8px;font-size:${titleSize()}">${esc(T('devis'))}</h3>
  <div style="display:flex;gap:10px;font-size:11px;color:#6b7280;margin-bottom:12px">${st.map((x,i)=>`<span style="${i?'':'font-weight:800;color:'+f.accent}">${i+1}. ${esc(x)}</span>`).join('')}</div>
  <div style="display:flex;gap:14px;align-items:flex-start"><div style="flex:2;border:1px solid #e5e7eb;border-radius:12px;padding:14px;background:#fff">
   ${fldBox(T('devis_nom'))}${fldBox(T('devis_email'))}${fldBox(T('devis_tel'))}${fldBox(T('devis_ent'))}
   <div style="font-size:12px;margin:8px 0">${esc(T('devis_client'))} &nbsp; ◉ ${esc(T('devis_oui'))} &nbsp; ○ ${esc(T('devis_non'))}</div>
   <span style="display:inline-block;background:${f.accent};color:#fff;border-radius:${r};padding:8px 16px;font-weight:700;font-size:12px">${esc(T('devis_continuer'))}</span></div>
  ${acc.help.show?`<div style="flex:1;display:flex;flex-direction:column;gap:10px"><div style="border:1px solid #e5e7eb;border-radius:12px;padding:12px;background:#fff;font-size:12px"><b>${esc(T('aide_titre'))}</b><p style="margin:4px 0 0;color:#6b7280">${esc(T('aide_texte'))}</p></div><div style="border:1px solid #e5e7eb;border-radius:12px;padding:12px;background:#fff;font-size:12px"><b>${esc(T('suite_titre'))}</b><p style="margin:4px 0 0;color:#6b7280">${esc(T('suite_texte'))}</p></div></div>`:''}
  </div></div>${pvFooterHTML()}</div>`;
}

function pvCompte(){
 const f=front,r=btnRadius(),c=acc.cpt;
 const modal=`<div style="position:fixed;inset:0;background:rgba(17,24,39,.5);z-index:20;display:flex;align-items:center;justify-content:center"><div style="width:min(420px,92%);background:#fff;border-radius:14px;padding:16px;font-size:12px">
  <div style="display:flex;gap:16px;border-bottom:1px solid #e5e7eb;margin-bottom:10px"><b style="padding:6px 0;border-bottom:2px solid ${f.accent}">${esc(T('cpt_onglet_conn'))}</b><span style="padding:6px 0;color:#6b7280">${esc(T('cpt_onglet_crea'))}</span></div>
  <b>${esc(T('cpt_titre'))}</b>${fldBox(T('cpt_email'))}${fldBox(T('cpt_mdp'))}
  ${c.oubli?`<div style="color:#2563eb;margin:6px 0">${esc(T('cpt_oubli'))}</div>`:''}
  <span style="display:inline-block;background:${f.accent};color:#fff;border-radius:${r};padding:8px 16px;font-weight:700">${esc(T('cpt_btn_conn'))}</span>
  <div style="margin-top:12px;color:#6b7280;border-top:1px dashed #e5e7eb;padding-top:8px">${esc(T('cpt_civ'))} · ${esc(T('cpt_prenom'))} · ${esc(T('cpt_nom'))} · ${esc(T('cpt_mdp2'))}<br><b style="color:${f.accent}">${esc(T('cpt_btn_crea'))}</b>${c.sync?'<br>'+esc(T('cpt_sync')):''}</div></div></div>`;
 const prefs=`<div style="display:flex;gap:14px"><div style="flex:1;border:1px solid #e5e7eb;border-radius:12px;padding:10px;background:#fff;font-size:12px;line-height:2">${['pref_infos','pref_cmd','pref_adr','pref_paniers','pref_params'].map(k=>'<div>'+esc(T(k))+'</div>').join('')}</div>
  <div style="flex:2;border:1px solid #e5e7eb;border-radius:12px;padding:14px;background:#fff;font-size:12px"><b>${esc(T('pref_requis'))}</b><p style="color:#6b7280">${esc(T('pref_theme'))}</p></div></div>`;
 const home=acc.compteView==='prefs';
 PV.html=`<div style="background:${f.bg};color:${f.text};font-family:${f.font},system-ui,sans-serif">${pvHeaderHTML()}<div class="pv-body" style="min-height:320px">${home?prefs:'<p style="color:#6b7280;font-size:12px">La fenêtre de connexion s\u2019ouvre par-dessus la page.</p>'}</div>${pvFooterHTML()}</div>${home?'':modal}`;
}

function contrastRatioV13(hex1,hex2){
 const lum=h=>{if(!/^#[0-9a-f]{6}$/i.test(h||''))return null;const c=[1,3,5].map(i=>parseInt(h.slice(i,i+2),16)/255).map(v=>v<=.04045?v/12.92:((v+.055)/1.055)**2.4);return .2126*c[0]+.7152*c[1]+.0722*c[2]};
 const a=lum(hex1),b=lum(hex2);if(a===null||b===null)return null;return (Math.max(a,b)+.05)/(Math.min(a,b)+.05);
}

function contrastIssuesV13(){
 const pairs=[['Texte du site',front.text,front.bg],['Prix',front.priceColor,front.bg],['Boutons',front.accent,'#ffffff'],['Bandeau d’annonce',front.announcementFg||'#ffffff',front.announcementBg||front.accent],['Badge de remise',header.badgeColor||'#dc2626','#ffffff']];
 return pairs.map(([name,fg,bg])=>({name,fg,bg,ratio:contrastRatioV13(fg,bg)})).filter(x=>x.ratio!==null&&x.ratio<4.5);
}

function validPreviewUrlV16(v){try{return ['https:','http:'].includes(new URL(v).protocol)?v:''}catch(_){return ''}}

function applyV16Preview(){
 const doc=pvI.contentDocument,root=doc&&doc.getElementById('root');if(!root)return;
 const st=studioV16.style,dark=studioV16.dark,id=studioV16.identity;
 let css='';const rad={square:'0px',soft:'10px',round:'20px'}[st.radius]||'10px',shadow={none:'none',light:'0 2px 10px #0001',strong:'0 8px 24px #0003'}[st.shadow]||'none',density={airy:'1.3',normal:'1',compact:'.78'}[st.density]||'1';
 css+=`#root{--v16-radius:${rad};--v16-shadow:${shadow};--v16-density:${density}}.pv-card,.pv-hero,.pv-box,.m-group,.mega-prev,.pv-custom-html{border-radius:var(--v16-radius)!important;box-shadow:var(--v16-shadow)!important}.pv-hero h4,.pv-body h1,.pv-body h2,.pv-body h3,.pv-body h4,.m-group h5{font-family:${st.titleFont?JSON.stringify(st.titleFont):'inherit'},sans-serif!important}.pv-card,.pv-col,.pv-body{line-height:var(--v16-density)}${st.buttons==='outline'?'.pv-body button,.pv-cta{background:transparent!important;color:'+front.accent+'!important;border:1px solid '+front.accent+'!important}':''}`;
 const systemDark=dark.defaultMode==='system'&&window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches;
 const isDark=!!(dark.enabled&&(dark.preview===true||(dark.preview===null&&(dark.defaultMode==='dark'||systemDark))));if(isDark)css+=`body,#root,.pv-body{background:${dark.bg}!important;color:${dark.text}!important}.pv-head,.pv-card,.pv-col,.pv-box,.m-group,.mega-prev,.pv-custom-html{background:${dark.card}!important;color:${dark.text}!important;border-color:${dark.border}!important}.pv-top,.pv-menu,.pv-foot{border-color:${dark.border}!important}.pv-body p,.pv-body span,.pv-body a,.pv-body h1,.pv-body h2,.pv-body h3,.pv-body h4,.pv-body h5{color:${dark.text}!important}.pv-body input,.pv-body textarea{background:${dark.card}!important;color:${dark.text}!important;border-color:${dark.border}!important}`;
 css+=studioV16.animations==='none'?'*,*::before,*::after{animation:none!important;transition:none!important;scroll-behavior:auto!important}':'@media(prefers-reduced-motion:no-preference){.pv-card:hover,.pv-col:hover{transform:translateY(-2px);transition:transform .2s ease,box-shadow .2s ease}.pv-body button{transition:transform .18s ease,filter .18s ease}}';
 let theme=doc.getElementById('theme');if(theme)theme.textContent+=`\n${css}`;
 if(id.logo){const logo=root.querySelector('.pv-logo');if(logo)logo.innerHTML='<img class="v16-preview-logo" src="'+esc(validPreviewUrlV16(id.logo))+'" alt="Logo Mediatec">'}
  const old=root.querySelector('.v16-preview-cookie');if(old)old.remove();if(studioV16.pages.cookies.show&&pagesInfoV16Active!=='cookies'){const p=PAGE_INFO_V16.find(x=>x.id==='cookies'),c=studioV16.pages.cookies,bar=doc.createElement('div');bar.className='v16-preview-cookie';bar.style.background=c.bg;bar.innerHTML='<span><strong>'+esc(T(p.title))+'</strong> · '+esc(T(p.lead))+'</span><button>'+esc(T('page_cookies_reject'))+'</button><button>'+esc(T('page_cookies_accept'))+'</button>';if(c.position==='top')root.prepend(bar);else root.appendChild(bar)}
  const darkToggle=root.querySelector('.v16-dark-toggle');if(darkToggle)darkToggle.remove();if(studioV16.dark.showToggle&&studioV16.dark.enabled){const head=root.querySelector('.pv-head');if(head){const btn=doc.createElement('button');btn.className='v16-dark-toggle';btn.type='button';btn.textContent=isDark?'☀ Clair':'☾ Sombre';btn.style.cssText='border:1px solid #9ca3af;border-radius:7px;padding:5px 8px;background:transparent;color:inherit;cursor:pointer';btn.addEventListener('click',()=>{studioV16.dark.preview=!studioV16.dark.preview;renderPreview()});head.appendChild(btn)}}
  const favicon=validPreviewUrlV16(id.favicon);let link=doc.querySelector('link[rel="icon"]');if(favicon){if(!link){link=doc.createElement('link');link.rel='icon';doc.head.appendChild(link)}link.href=favicon}else if(link)link.remove();
 renderCountersV16();
}

