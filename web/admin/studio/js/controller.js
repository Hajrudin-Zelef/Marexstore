/* controller.js — Studio Mediatec (MVC). Généré par découpe de studio-prototype-v16-2.html. */

function setFront(k, v){
  if(k==='accent' && !/^#[0-9a-fA-F]{6}$/.test(v)) return;
  front[k] = v;
  if(k==='accent'){
    const h=$('#hexAccent'); if(h) h.value=v; renderSwatches();
    FILES['theme.css'].content = FILES['theme.css'].content.replace(/--accent\s*:\s*#[0-9a-fA-F]{3,6}/, '--accent:'+v);
    front.customCSS = FILES['theme.css'].content;
    if(currentFile==='theme.css' && !$('#front-code').classList.contains('hidden')){ codeInput.value = FILES['theme.css'].content; syncEditor(); }
  }
  const TV={priceColor:'--price',bg:'--bg'};
  if(TV[k]){
    FILES['theme.css'].content = FILES['theme.css'].content.replace(new RegExp(TV[k]+'\\s*:\\s*#[0-9a-fA-F]{3,6}'), TV[k]+':'+v);
    front.customCSS = FILES['theme.css'].content;
    if(currentFile==='theme.css' && !$('#front-code').classList.contains('hidden')){ codeInput.value = FILES['theme.css'].content; syncEditor(); }
  }
  renderPreview();
}

function setFrontSeg(btn, key){
  btn.parentElement.querySelectorAll('button').forEach(x=>x.classList.toggle('active', x===btn));
  front[key] = btn.dataset.v; renderPreview();
}

$('#studioNav').addEventListener('click', e=>{
  const b = e.target.closest('button[data-tab]'); if(!b) return;
  document.querySelectorAll('#studioNav button').forEach(x=>x.classList.toggle('active', x===b));
  document.querySelectorAll('#editorPanel > div').forEach(d=>d.classList.add('hidden'));
  $('#tab-'+b.dataset.tab).classList.remove('hidden');
  currentTab = b.dataset.tab;
  renderPreview();
});

function setFrontMode(mode){
  document.querySelectorAll('#frontMode button').forEach(x=>x.classList.toggle('active', x.dataset.mode===mode));
  $('#front-ui').classList.toggle('hidden', mode!=='ui');
  $('#front-code').classList.toggle('hidden', mode!=='code');
  if(mode==='code'){ renderTabs(); loadFile(currentFile); toast('Mode Code : ouvrez un fichier, modifiez-le, Publiez.'); }
}

codeInput.addEventListener('input', ()=>{
  FILES[currentFile].content = codeInput.value;
  syncEditor();
  if(FILES[currentFile].lang==='css'){
    front.customCSS = codeInput.value;
    const m = codeInput.value.match(/--accent\s*:\s*(#[0-9a-fA-F]{3,6})/);
    if(m){ front.accent = m[1]; const h=$('#hexAccent'); if(h) h.value=m[1]; renderSwatches(); }
    renderPreview();
  }
});
codeInput.addEventListener('scroll', ()=>{
  codePre.scrollTop = codeInput.scrollTop; codePre.scrollLeft = codeInput.scrollLeft;
  codeGutter.scrollTop = codeInput.scrollTop;
});
codeInput.addEventListener('keydown', e=>{
  if(e.key==='Tab'){ e.preventDefault();
    const s = codeInput.selectionStart;
    codeInput.value = codeInput.value.slice(0,s) + '  ' + codeInput.value.slice(codeInput.selectionEnd);
    codeInput.selectionStart = codeInput.selectionEnd = s + 2;
    codeInput.dispatchEvent(new Event('input'));
  }
});

document.querySelectorAll('.ft-file').forEach(b=>b.addEventListener('click', ()=>openFile(b.dataset.file)));

function setShopSeg(btn, field){
  btn.parentElement.querySelectorAll('button').forEach(b=>b.classList.remove('active'));
  btn.classList.add('active');
  if(field==='perPage') shop[field] = btn.dataset.v==='all' ? 'all' : parseInt(btn.dataset.v,10);
  else shop[field] = btn.dataset.v;
  renderPreview();
}

function setShopPage(btn, page){
  btn.parentElement.querySelectorAll('button').forEach(b=>b.classList.remove('active'));
  btn.classList.add('active');
  shop.previewPage = page;
  renderPreview();
}

function setFicheSeg(btn, field){
  btn.parentElement.querySelectorAll('button').forEach(b=>b.classList.remove('active'));
  btn.classList.add('active');
  fiche[field] = field==='similarNb' ? parseInt(btn.dataset.v,10) : btn.dataset.v;
  renderPreview();
}

function setTxt(key, v){ setT(key, v); applyTxtToFront(); }

function applyTxtToFront(){
  front.banner = Tx.bandeau || '';
  const b = $('#bannerTxt'); if(b) b.value = front.banner;
}

function loadTxtPage(){
  txtPageSel = $('#selTxtPage').value;
  $('#taTxtPage').value = txtPages[txtPageSel] || '';
  renderPreview();
}

function saveTxtPage(){
  txtPages[txtPageSel] = $('#taTxtPage').value;
}

Object.assign(ICONS,{
  up:'<path d="m6 15 6-6 6 6"/>',down:'<path d="m6 9 6 6 6-6"/>',x:'<path d="M6 6l12 12M18 6 6 18"/>',plus:'<path d="M12 5v14M5 12h14"/>',minus:'<path d="M5 12h14"/>',
  back:'<path d="M19 12H5M11 6l-6 6 6 6"/>',undo:'<path d="M9 14 4 9l5-5"/><path d="M4 9h10a6 6 0 0 1 0 12h-3"/>',redo:'<path d="m15 14 5-5-5-5"/><path d="M20 9H10a6 6 0 0 0 0 12h3"/>',
  reset:'<path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/>',upload:'<path d="M12 16V4M7 9l5-5 5 5"/><path d="M4 20h16"/>',lock:'<rect x="5" y="11" width="14" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>',
  monitor:'<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M8 21h8M12 16v5"/>',phone:'<rect x="7" y="2" width="10" height="20" rx="2.5"/><path d="M11 18h2"/>',info:'<circle cx="12" cy="12" r="9"/><path d="M12 11v5M12 8h.01"/>',
  spinner:'<path d="M12 3a9 9 0 1 0 9 9"/>','check-circle':'<circle cx="12" cy="12" r="9"/><path d="m8 12 3 3 5-6"/>',ban:'<circle cx="12" cy="12" r="9"/><path d="m5.6 5.6 12.8 12.8"/>',alert:'<path d="M12 3 2 20h20L12 3z"/><path d="M12 10v4M12 17h.01"/>',menu:'<path d="M4 7h16M4 12h16M4 17h16"/>',
  laptop:'<rect x="4" y="5" width="16" height="11" rx="1.5"/><path d="M2 19h20"/>',printer:'<path d="M7 9V3h10v6"/><rect x="4" y="9" width="16" height="8" rx="2"/><path d="M7 14h10v7H7z"/>',globe:'<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>',
  plug:'<path d="M9 2v5M15 2v5"/><path d="M6 7h12v4a6 6 0 0 1-12 0V7z"/><path d="M12 17v5"/>',battery:'<rect x="2" y="7" width="17" height="10" rx="2"/><path d="M22 11v2M7 11v2M11 11v2"/>',server:'<rect x="3" y="3" width="18" height="7" rx="2"/><rect x="3" y="14" width="18" height="7" rx="2"/><path d="M7 6.5h.01M7 17.5h.01"/>',
  antenna:'<path d="M12 12v9"/><circle cx="12" cy="10" r="2"/><path d="M7.8 5.8a6 6 0 0 0 0 8.4M16.2 5.8a6 6 0 0 1 0 8.4"/>',keyboard:'<rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h.01M10 10h.01M14 10h.01M18 10h.01M7 14h10"/>',mouse:'<rect x="6" y="3" width="12" height="18" rx="6"/><path d="M12 7v4"/>',box:'<path d="M21 8 12 3 3 8v8l9 5 9-5V8z"/><path d="m3 8 9 5 9-5M12 13v8"/>'
});

let pvReady=false, pvPending=null;

pvI.addEventListener('load',()=>{pvReady=true;if(pvPending!==null)PV.html=pvPending});
pvI.srcdoc=`<!doctype html><html lang="fr"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="${FONTS}"><style>${PV_BASE}${pvCss()}</style><style id="theme"></style></head><body><div id="root"></div></body></html>`;

function openModal(){$('#modal').classList.remove('hidden');$('#mCancel').focus()}

function closeModal(){$('#modal').classList.add('hidden')}

function resetStudio(){openModal()}

function doReset(){closeModal();clearTimeout(histTimer);pushHist();restore(ORIGIN);pushHist();histBtns();toast('Studio réinitialisé (design d\u2019origine).')}

/* Thème serveur depuis l'état du Studio (front + textes + sections).
   Phase 1 : seul le thème part vers le serveur (POST /admin/design) ; le
   reste du design (header, footer, shop…) sera persisté en phase 2 (API JSON). */
function themeFromFront(){
  const accent=(front.accent||'#2E5136').toUpperCase();
  let couleur='perso';
  if(accent==='#2E5136')couleur='gazonnier';
  else if(accent==='#84CC16')couleur='liquid';
  return {
    couleur:couleur,
    accent:(front.accent||'').replace(/^#/,''),
    mode:'auto',
    police:(front.font||'Inter')==='Inter'?'inter':'systeme',
    annonce_on:front.showAnnonce?'1':'0',
    annonce_texte:Tx.bandeau||'',
    /* Sections du serveur, renvoyées telles quelles (phase 1 : le Studio ne
       les gère pas — surtout pas rows.map, l'éditeur visuel n'a que 3 lignes
       alors que le site en a 9). */
    sections:(front._serverSections&&front._serverSections.length)?front._serverSections.slice():[],
    css:front.customCSS||''
  };
}
/* Thème courant du serveur (injecté par AdminDesign dans window.__STUDIO_THEME__)
   appliqué à l'état du Studio au démarrage, avant le premier rendu. */
function applyServerTheme(){
  const t=window.__STUDIO_THEME__; if(!t)return;
  if(t.accent)front.accent=t.accent;
  if(t.font)front.font=(t.font==='systeme')?'system-ui':'Inter';
  if(typeof t.showAnnonce==='boolean')front.showAnnonce=t.showAnnonce;
  if(typeof t.annonceTexte==='string'&&t.annonceTexte)Tx.bandeau=t.annonceTexte;
  if(typeof t.css==='string'&&t.css)front.customCSS=t.css;
  if(t.couleur)front.theme=t.couleur;
  /* Sections d'accueil du serveur : le Studio ne les gère pas en phase 1,
     il les renvoie telles quelles à la publication (themeFromFront). */
  if(Array.isArray(t.sections)&&t.sections.length)front._serverSections=t.sections.slice();
}
/* Publication serveur : envoie le thème à /admin/design (POST + CSRF).
   Le serveur incrémente theme_version et le site utilise le thème publié. */
function publishStudio(){
  const b=$('#btnPublish'); if(b.classList.contains('busy'))return;
  b.classList.add('busy'); b.innerHTML=ic('spinner','i spin')+'<span class="t">Publication…</span>';
  const th=themeFromFront();
  const body=new URLSearchParams();
  body.set('csrf', window.__STUDIO_CSRF__||'');
  body.set('couleur', th.couleur);
  body.set('accent', th.accent);
  body.set('mode', th.mode);
  body.set('police', th.police);
  body.set('annonce_on', th.annonce_on);
  body.set('annonce_texte', th.annonce_texte);
  th.sections.forEach(s=>body.append('sections', s));
  if(th.sections.length)body.set('ordre', th.sections.join(','));
  body.set('css', th.css);
  fetch(window.location.origin + '/admin/design',{method:'POST',headers:{'Accept':'application/json','Content-Type':'application/x-www-form-urlencoded'},body:body.toString()})
    .then(r=>{if(!r.ok)throw new Error('HTTP '+r.status);return r.json()})
    .then(d=>{
      pubSnap=snap(); b.classList.remove('busy'); b.innerHTML=ic('check-circle')+'<span class="t">Publié</span>';
      renderPreview(); toast('✅ Studio publié ! Le site est à jour (thème v'+(d.version||'?')+').');
      setTimeout(()=>{b.innerHTML=ic('upload')+'<span class="t">Publier</span>'},1800);
    })
    .catch(e=>{
      b.classList.remove('busy'); b.innerHTML=ic('upload')+'<span class="t">Publier</span>';
      toast('❌ Publication impossible : '+e.message);
    });
}

function initPremium(){
  document.querySelectorAll('[data-ic]').forEach(e=>e.insertAdjacentHTML('afterbegin',ic(e.dataset.ic)));
  document.querySelectorAll('[oninput],[onchange]').forEach(e=>{
    const m=(e.getAttribute('oninput')||e.getAttribute('onchange')).match(/^(?:setFront\('(\w+)',this\.\w+\)|(\w+)\.(\w+)=this\.\w+)/);
    if(m)e.dataset.bind=m[1]?'front.'+m[1]:m[2]+'.'+m[3];
  });
  Object.entries(TABICO).forEach(([t,n],i)=>{const h=$('#tab-'+t+' h2');if(h)h.insertAdjacentHTML('afterbegin',miniIco(n,i))});
  updateCrumb($('#studioNav .active'));
  $('#studioNav').addEventListener('click',e=>{
    const b=e.target.closest('button[data-tab]'); if(!b)return; updateCrumb(b);
    if(!matchMedia('(prefers-reduced-motion:reduce)').matches)pvI.animate([{opacity:0,transform:'translateY(10px) scale(.985)'},{opacity:1,transform:'none'}],{duration:380,easing:'cubic-bezier(.22,.9,.24,1)'});
  });
  addEventListener('keydown',e=>{
    if(e.key==='Escape')closeModal();
    if(!(e.metaKey||e.ctrlKey)||/^(INPUT|TEXTAREA|SELECT)$/.test(e.target.tagName))return;
    const k=e.key.toLowerCase();
    if(k==='z'){e.preventDefault();e.shiftKey?redo():undo()}else if(k==='y'){e.preventDefault();redo()}
  });
  BASE=ORIGIN=pubSnap=snap(); hist=[BASE]; hi=0; updateStatus();
}

/* Thème courant du serveur appliqué avant le premier rendu. */
applyServerTheme();
renderSwatches();
renderRows();
renderFacets();
renderMenuList();
renderTxtList();
renderStepEditor();
renderSeoCounters();
renderAllModuleTexts();
renderNavLinks();
renderFooterCols();
loadFile(currentFile);
loadTxtPage();
makeSortable($('#facetList'), (k, tk)=>{
  const a = shop.facets.findIndex(x=>x.key===k), b = shop.facets.findIndex(x=>x.key===tk);
  if(a<0||b<0) return;
  const [it] = shop.facets.splice(a,1); shop.facets.splice(b,0,it);
  renderFacets(); renderPreview();
});
makeSortable($('#menuList'), (k, tk)=>{
  const a = menu.fams.findIndex(x=>x.key===k), b = menu.fams.findIndex(x=>x.key===tk);
  if(a<0||b<0) return;
  const [it] = menu.fams.splice(a,1); menu.fams.splice(b,0,it);
  renderMenuList(); renderPreview();
});
makeSortable($('#navList'), (k, tk)=>{
  const a = header.navLinks.findIndex(x=>x.key===k), b = header.navLinks.findIndex(x=>x.key===tk);
  if(a<0||b<0) return;
  const [it] = header.navLinks.splice(a,1); header.navLinks.splice(b,0,it);
  renderNavLinks(); renderPreview();
});
renderPreview();
initPremium();

addEventListener('load',()=>{
 const nav=$('#studioNav'), mb={};
 nav.querySelectorAll('button[data-tab]').forEach(b=>mb[b.dataset.tab]=b);
 const first=nav.querySelector('button.active');
 nav.innerHTML='';
 CATEGORIES.forEach(c=>{
  const w=document.createElement('div'); w.className='cat'; w.dataset.cat=c.id;
  w.innerHTML='<button type="button" class="cat-btn" aria-haspopup="true" aria-expanded="false"><span class="nav-ico">'+mb[c.mods[0]].querySelector('.nav-ico').innerHTML+'</span>'+c.label+'<svg class="chev" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg></button><div class="cat-menu"></div>';
  c.mods.forEach(m=>w.querySelector('.cat-menu').appendChild(mb[m]));
  nav.appendChild(w);
 });
 const catOf=t=>CATEGORIES.find(c=>c.mods.includes(t));
 const closeAll=()=>nav.querySelectorAll('.cat').forEach(x=>{x.classList.remove('open');x.querySelector('.cat-btn').setAttribute('aria-expanded','false')});
 const sync=t=>{
  const c=catOf(t); nav.querySelectorAll('.cat').forEach(x=>x.classList.toggle('on',x.dataset.cat===c.id));
  $('#crumbTitle').textContent=c.label+' / '+mb[t].textContent.trim();
 };
 nav.addEventListener('click',e=>{
  const cb=e.target.closest('.cat-btn');
  if(cb){const w=cb.parentElement,was=w.classList.contains('open');closeAll();if(!was){w.classList.add('open');cb.setAttribute('aria-expanded','true')}return}
  const mbtn=e.target.closest('button[data-tab]'); if(mbtn){closeAll();sync(mbtn.dataset.tab)}
 });
 document.addEventListener('click',e=>{if(!e.target.closest('#studioNav'))closeAll()});
 addEventListener('keydown',e=>{if(e.key==='Escape')closeAll()});
 sync(first?first.dataset.tab:'front');

 /* --- sous-onglets Design | Textes --- */
 const dtMode={};
 const applyDT=(tab,mode)=>{
  const root=$('#tab-'+tab), mt=$('#mt-'+tab); if(!root||!mt)return;
  const holder=mt.parentElement; dtMode[tab]=mode;
  root.querySelectorAll('.dt-off').forEach(x=>x.classList.remove('dt-off'));
  root.querySelectorAll('.dt-seg button').forEach(b=>b.classList.toggle('active',b.dataset.m===mode));
  if(mode==='design'){holder.classList.add('dt-off');return}
  if(tab==='front'&&typeof setFrontMode==='function')setFrontMode('ui');
  const keepTop=new Set([root.querySelector(':scope > h2'),root.querySelector(':scope > .dt-seg'),root.querySelector(':scope > p.hint')]);
  let n=holder;
  while(n&&n!==root){
   [...n.parentElement.children].forEach(sib=>{
    if(sib===n||(n.parentElement===root&&keepTop.has(sib)))return;
    sib.classList.add('dt-off');
   });
   n=n.parentElement;
  }
 };
 Object.keys(MODULE_TEXTS).concat(['accueil']).forEach(tab=>{
  const root=$('#tab-'+tab), mt=$('#mt-'+tab), h2=root&&root.querySelector(':scope > h2'); if(!mt||!h2)return;
  const seg=document.createElement('div'); seg.className='seg dt-seg';
  seg.innerHTML='<button type="button" data-m="design" class="active">Design</button><button type="button" data-m="textes">Textes</button>';
  h2.insertAdjacentElement('afterend',seg);
  seg.addEventListener('click',e=>{const b=e.target.closest('button');if(b)applyDT(tab,b.dataset.m)});
  /* recherche */
  const q=document.createElement('input'); q.type='search'; q.className='tx-search'; q.placeholder='Rechercher un texte (libellé ou contenu)…';
  const none=document.createElement('div'); none.className='tx-none dt-off'; none.textContent='Aucun résultat.';
  mt.before(q); mt.after(none);
  q.addEventListener('input',()=>{
   const v=q.value.trim().toLowerCase(); let vis=0;
   mt.querySelectorAll('.txt-item').forEach(r=>{
    const hay=(r.querySelector('.k').textContent+' '+(r.querySelector('input')||{}).value).toLowerCase();
    const ok=!v||hay.includes(v); r.classList.toggle('dt-off',!ok); if(ok)vis++;
   });
   none.classList.toggle('dt-off',vis>0);
  });
  applyDT(tab,'design');
 });

 /* --- palette Ctrl+K --- */
 const kb=document.createElement('div'); kb.id='kbar';
 kb.innerHTML='<div class="box"><input type="text" placeholder="Aller à un module… (boutique, pied de page, SEO)"><div class="list"></div></div>';
 document.body.appendChild(kb);
 const ki=kb.querySelector('input'), kl=kb.querySelector('.list'); let items=[],sel=0;
 const all=[]; CATEGORIES.forEach(c=>c.mods.forEach(m=>all.push({t:m,name:mb[m].textContent.trim(),cat:c.label})));
 const draw=()=>{
  const v=ki.value.trim().toLowerCase();
  items=all.filter(x=>!v||(x.name+' '+x.cat).toLowerCase().includes(v)); sel=Math.min(sel,Math.max(items.length-1,0));
  kl.innerHTML=items.map((x,i)=>'<div class="it'+(i===sel?' sel':'')+'" data-i="'+i+'"><span>'+x.name+'</span><small>'+x.cat+'</small></div>').join('')||'<div class="it"><small>Aucun résultat</small></div>';
 };
 const go=i=>{const x=items[i];if(!x)return;kb.classList.remove('open');mb[x.t].click()};
 const open=()=>{kb.classList.add('open');ki.value='';sel=0;draw();ki.focus()};
 kl.addEventListener('click',e=>{const it=e.target.closest('.it[data-i]');if(it)go(+it.dataset.i)});
 kb.addEventListener('mousedown',e=>{if(e.target===kb)kb.classList.remove('open')});
 ki.addEventListener('input',()=>{sel=0;draw()});
 ki.addEventListener('keydown',e=>{
  if(e.key==='ArrowDown'){e.preventDefault();sel=Math.min(sel+1,items.length-1);draw()}
  else if(e.key==='ArrowUp'){e.preventDefault();sel=Math.max(sel-1,0);draw()}
  else if(e.key==='Enter'){e.preventDefault();go(sel)}
  else if(e.key==='Escape')kb.classList.remove('open');
 });
 addEventListener('keydown',e=>{if((e.ctrlKey||e.metaKey)&&e.key.toLowerCase()==='k'){e.preventDefault();kb.classList.contains('open')?kb.classList.remove('open'):open()}});
});

addEventListener('load',()=>{
 const _r=renderModuleTexts; renderModuleTexts=function(m){_r(m);groupTexts(m)};
 Object.keys(TG).forEach(groupTexts);
 document.querySelectorAll('.tx-search').forEach(q=>q.addEventListener('input',()=>{
  const v=q.value.trim();
  q.parentElement.querySelectorAll('.tx-grp').forEach(d=>{
   const has=!!d.querySelector('.txt-item:not(.dt-off)');
   d.classList.toggle('dt-off',!has); if(v&&has)d.open=true;
  });
 }));
 pvHeader=function(){pvHome()};
 pvFooter=function(){pvHome();setTimeout(()=>{try{const d=pvI.contentDocument;(d.scrollingElement||d.documentElement).scrollTop=1e7}catch(_){}},200)};
 renderPreview();
});

addEventListener('load',()=>{
 const _h=pvHome,_set=Object.getOwnPropertyDescriptor(PV,'html').set;
 pvHome=function(){Object.defineProperty(PV,'html',{configurable:true,set(h){Object.defineProperty(PV,'html',{configurable:true,set:_set});PV.html=h.replace('<div class="pv-body">',accHTML()+'<div class="pv-body">')}});_h()};
 TG.fiche[0][1].push('express','garantie','tva','partager');
 TG.fiche.push(['Onglets et avis',['tab_desc','tab_marque','tab_avis','aucun_avis','premier_avis','evaluer','votre_nom','votre_avis','envoyer_avis']]);
 groupTexts('fiche');
 const _setPV3=Object.getOwnPropertyDescriptor(PV,'html').set, _ppPan=pvPanier;
 pvPanier=function(){
  if(acc.cartView==='devis')return pvDevis();
  Object.defineProperty(PV,'html',{configurable:true,set(h){Object.defineProperty(PV,'html',{configurable:true,set:_setPV3});PV.html=h+drawerHTML()}});_ppPan();
 };
 TG.panier=[['Totaux',['sous_total','livraison','livraison_offerte','total','articles']],['Tiroir panier',['tiroir_fermer','tiroir_retirer','tiroir_vider','tiroir_continuer','tiroir_livraison','toast_ajoute']],['Devis — formulaire',['devis_nom','devis_email','devis_tel','devis_ent','devis_client','devis_oui','devis_non','devis_continuer']],['Devis — cartes latérales',['aide_titre','aide_texte','suite_titre','suite_texte']],['Divers',['ph_texte']]];
 TG.compte=[['Connexion',['cpt_onglet_conn','cpt_titre','cpt_email','cpt_mdp','cpt_oubli','cpt_btn_conn']],['Création de compte',['cpt_onglet_crea','cpt_civ','cpt_prenom','cpt_nom','cpt_mdp2','cpt_btn_crea','cpt_sync']],['Préférences',['pref_infos','pref_cmd','pref_adr','pref_paniers','pref_params','pref_requis','pref_theme']]];
 renderAllModuleTexts(); renderAccEditor(); renderPreview();
});

function applyThemeV13(id){
 const t=THEMES_V13.find(x=>x.id===id);if(!t)return;
 Object.assign(front,{theme:t.id,accent:t.accent,priceColor:t.price,bg:t.bg,text:t.text,font:t.font});
 if($('#selFont'))$('#selFont').value=t.font;
 syncUI();renderPreview();
}

const _pvHeaderV13=pvHeaderHTML;

pvHeaderHTML=function(){return annHtmlV13()+_pvHeaderV13()};

const _publishV13=publishStudio;

publishStudio=function(){
 const run=()=>{const before=pubSnap?JSON.parse(pubSnap):null;_publishV13();setTimeout(()=>{
  const after=JSON.parse(snap());const keys=['front','shop','fiche','panier','Tx','txtPages','menu','seo','reglages','header','footer','acc','files'];
  const modules=keys.filter(k=>!before||JSON.stringify(before[k])!==JSON.stringify(after[k]));
  publishedHistoryV13.push({date:new Intl.DateTimeFormat('fr-FR',{dateStyle:'short',timeStyle:'short'}).format(new Date()),modules,snapshot:snap()});renderHistoryV13();
 },1050)};
 const issues=contrastIssuesV13();if(issues.length)contrastDialogV13(issues,run);else run();
};

function addCommandsV13(){
 const input=$('#kbar input'),list=$('#kbar .list');if(!input||!list)return;
 const runCommand=el=>{const key=el.dataset.command;if(!key)return;const target=document.getElementById(key),tab=target?.closest('[id^="tab-"]');if(!target||!tab)return;const mod=tab.id.slice(4),btn=document.querySelector('#studioNav button[data-tab="'+mod+'"]');if(btn)btn.click();if(target.closest('.txt-item')){const mode=tab.querySelector('.dt-seg button[data-m="textes"]');if(mode)mode.click()}const group=target.closest('details');if(group)group.open=true;target.scrollIntoView({block:'center',behavior:'smooth'});const edit=target.matches('input,textarea,select,button')?target:target.querySelector('input,textarea,select,button');if(edit)edit.focus()};
 const drawTexts=()=>{const q=input.value.trim().toLowerCase();if(!q)return;const seen=new Set();document.querySelectorAll('.txt-item').forEach((row,i)=>{const k=row.querySelector('.k')?.textContent||'',v=row.querySelector('input,textarea')?.value||'';if((k+' '+v).toLowerCase().includes(q)){if(!row.id)row.id='v13cmd-'+i;seen.add(row.id)}});[...seen].map(id=>document.getElementById(id)).filter(Boolean).slice(0,8).forEach(row=>{const div=document.createElement('div');div.className='it v13-command';div.dataset.command=row.id;div.innerHTML='<span>'+esc(row.querySelector('.k')?.textContent||'Réglage')+'</span><small>Texte</small>';list.appendChild(div)})};
 const drawDesign=()=>{const q=input.value.trim().toLowerCase();if(!q)return;document.querySelectorAll('#editorPanel .field').forEach((field,i)=>{const label=field.querySelector(':scope > label')||field.querySelector('label'),txt=label?.textContent||'',target=field.querySelector('input,textarea,select,button');if(!target||!(txt+' '+target.value+' '+target.id).toLowerCase().includes(q))return;if(!target.id)target.id='v13-design-'+i;const div=document.createElement('div');div.className='it v13-command';div.dataset.command=target.id;div.innerHTML='<span>'+esc(txt.trim().slice(0,90)||target.id)+'</span><small>Design</small>';list.appendChild(div)})};
 input.addEventListener('input',drawTexts);input.addEventListener('input',drawDesign);
 list.addEventListener('click',e=>{const b=e.target.closest('.v13-command');if(b){runCommand(b);$('#kbar').classList.remove('open')}});
 input.addEventListener('keydown',e=>{if(e.key==='Enter'){const b=list.querySelector('.v13-command');if(b){e.preventDefault();e.stopImmediatePropagation();runCommand(b);$('#kbar').classList.remove('open')}}},true);
}

addEventListener('load',()=>{
 const css=document.createElement('style');css.textContent='.v13-themes{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:8px}.v13-theme{border:1px solid var(--line);border-radius:10px;background:#fff;padding:7px;text-align:left;cursor:pointer;color:var(--ink)}.v13-theme:hover{border-color:#0ea5e9;transform:translateY(-1px)}.v13-thumb{height:42px;border-radius:6px;display:block;position:relative;margin-bottom:5px}.v13-thumb i,.v13-thumb em{position:absolute;width:20px;height:5px;border-radius:4px;left:7px;top:8px}.v13-thumb em{top:18px;left:11px}.v13-theme small{display:block;color:var(--muted)}.v13-ann{border-top:1px solid var(--line)}.v13-history-row{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:10px;border:1px solid var(--line);border-radius:9px;margin:6px 0;background:#fff}.v13-history-row small{display:block;color:var(--muted);margin-top:3px}.v13-command{border-color:#bfdbfe}.pv-announcement a{text-decoration:none}.pv-announcement a:hover{text-decoration:underline}@media(max-width:720px){.v13-themes{grid-template-columns:repeat(2,minmax(0,1fr))}.v13-history-row{align-items:flex-start;flex-direction:column}}';document.head.appendChild(css);
 themeGalleryV13();announceEditorV13();historyPanelV13();syncV13Fields();renderHistoryV13();addCommandsV13();
 const mobile=document.querySelector('.devices button[title="Mobile"]');if(mobile){mobile.title='Aperçu mobile — réglages dédiés en attente d’observation du site réel';mobile.setAttribute('aria-label','Aperçu mobile uniquement; réglages dédiés en attente d’observation du site réel')}
 const origSync=syncUI;syncUI=function(){origSync();syncV13Fields()};
 const origThemeRender=renderPreview;renderPreview=function(){origThemeRender();syncV13Fields()};
 const mobileNote=document.createElement('span');mobileNote.className='hint';mobileNote.style.cssText='font-size:10px;margin-left:8px';mobileNote.textContent='Mobile : aperçu seulement (réglages dédiés après observation du vrai site).';document.querySelector('.pvbar').appendChild(mobileNote);
 renderPreview();
});

Object.assign(Tx,{
  fam_reseaux:'Matériels réseaux',
  fam_accessoires:'Accessoires informatiques',
  fam_imprimantes:'Imprimantes',
  fam_laptops:'Ordinateurs portables',
  fam_bureau:'Ordinateurs de bureau',
  fam_serveurs:'Serveurs',
  fam_cameras:'Caméras & Vidéosurveillance',
  mega_ordinateurs:'Ordinateurs',
  mega_accessoires:'Accessoires',
  mega_consommables:'Consommables'
});
MODULE_TEXTS.menu.push(
  {key:'fam_reseaux',label:'Famille « Matériels réseaux »'},
  {key:'fam_accessoires',label:'Famille « Accessoires informatiques »'},
  {key:'fam_imprimantes',label:'Famille « Imprimantes »'},
  {key:'fam_laptops',label:'Famille « Ordinateurs portables »'},
  {key:'fam_bureau',label:'Famille « Ordinateurs de bureau »'},
  {key:'fam_serveurs',label:'Famille « Serveurs »'},
  {key:'fam_cameras',label:'Famille « Caméras & Vidéosurveillance »'},
  {key:'mega_ordinateurs',label:'Sous-catégorie « Ordinateurs »'},
  {key:'mega_accessoires',label:'Sous-catégorie « Accessoires »'},
  {key:'mega_consommables',label:'Sous-catégorie « Consommables »'}
);
renderMenuList=function(){
  const el=$('#menuList');if(!el)return;
  const ico={reseaux:'grid',accessoires:'tag',imprimantes:'file',laptops:'home',bureau:'list',serveurs:'gear',cameras:'image'};
  el.innerHTML=menu.fams.map((f,i)=>{
    const key='fam_'+f.key,label=T(key)||f.label;
    return '<div class="sort-item" draggable="true" data-k="'+esc(f.key)+'">'
      +'<span class="grip">⋮⋮</span>'+miniIco(ico[f.key]||'grid',i)
      +'<span class="lbl" style="display:flex;flex-direction:column;flex:1;min-width:0">'
      +'<input class="menu-label" data-menu-label="'+key+'" data-k="'+key+'" type="text" aria-label="Libellé de la famille" value="'+esc(label)+'" style="width:100%;min-width:0;padding:6px 8px;font-size:13px;font-weight:700">'
      +'<span class="sub">'+f.count+' produit'+(f.count>1?'s':'')+' · catalogue (lecture seule)</span></span>'
      +'<label class="switch"><input type="checkbox" '+(f.visible?'checked':'')+' onchange="toggleFam(\''+esc(f.key)+'\',this.checked)"><span class="tr"></span></label></div>';
  }).join('');
  el.querySelectorAll('[data-menu-label]').forEach(input=>input.addEventListener('input',e=>setT(e.currentTarget.dataset.menuLabel,e.currentTarget.value)));
};
pvMenu=function(){
  const f=front,fams=menu.fams.filter(x=>x.visible);
  PV.html='<div style="background:'+f.bg+';color:'+f.text+';font-family:'+f.font+',system-ui,sans-serif">'+pvHeaderHTML()
    +'<div class="pv-body"><div class="mega-prev"><div class="m-bar" style="background:'+f.accent+'">'+esc(T('toutes_cat'))+'</div><div class="m-grid">'
    +fams.map(fm=>'<div class="m-group"><h5 style="color:'+f.accent+'">'+esc(T('fam_'+fm.key)||fm.label)
      +(menu.showCount?' <span style="color:#9ca3af;font-weight:400">('+fm.count+')</span>':'')
      +'</h5><a href="#">'+esc(T('mega_ordinateurs'))+'</a><a href="#">'+esc(T('mega_accessoires'))+'</a><a href="#">'+esc(T('mega_consommables'))+'</a>'
      +(menu.showVoirTous?'<a href="#" style="color:'+f.accent+';font-weight:700">'+esc(T('voir_tous'))+' →</a>':'')
      +'</div>').join('')
    +'</div></div><p style="color:#6b7280;font-size:12px;margin:10px 0 0">'+fams.length+' famille(s) affichée(s) · '+menu.fams.filter(x=>!x.visible).length+' masquée(s).</p></div>'+pvFooterHTML()+'</div>';
};
addEventListener('load',()=>{
  renderMenuList();
  renderAllModuleTexts();
  renderTxtList();
  renderPreview();
});

menu.fams.forEach(f=>MENU_SUB_ITEMS_V14.forEach(item=>{
  const key='sub_'+f.key+'_'+item.key;
  if(Tx[key]===undefined) Tx[key]=Tx[item.base]||item.label;
  if(!MODULE_TEXTS.menu.some(x=>x.key===key)) MODULE_TEXTS.menu.push({key,label:'Sous-menu · '+f.label+' · '+item.label});
}));
renderMenuList=function(){
  const el=$('#menuList');if(!el)return;
  const ico={reseaux:'grid',accessoires:'tag',imprimantes:'file',laptops:'home',bureau:'list',serveurs:'gear',cameras:'image'};
  el.innerHTML=menu.fams.map((f,i)=>{
    const key='fam_'+f.key,label=T(key)||f.label;
    const rows=MENU_SUB_ITEMS_V14.map(item=>{
      const textKey='sub_'+f.key+'_'+item.key,value=T(textKey)||item.label;
      return '<div class="menu-sub-row"><span class="menu-sub-icon" aria-hidden="true">›</span>'
        +'<span class="menu-sub-value" data-sub-value="'+textKey+'">'+esc(value)+'</span>'
        +'<input class="menu-sub-input" data-menu-label="'+textKey+'" data-k="'+textKey+'" type="text" aria-label="Modifier '+esc(value)+'" value="'+esc(value)+'" hidden>'
        +'<button type="button" class="menu-sub-edit" data-sub-edit="'+textKey+'" aria-label="Modifier le sous-menu '+esc(value)+'" title="Modifier">✎</button></div>';
    }).join('');
    return '<div class="sort-item" draggable="true" data-k="'+esc(f.key)+'">'
      +'<span class="grip">⋮⋮</span>'+miniIco(ico[f.key]||'grid',i)
      +'<span class="lbl" style="display:flex;flex-direction:column;flex:1;min-width:0">'
      +'<input class="menu-label" data-menu-label="'+key+'" data-k="'+key+'" type="text" aria-label="Libellé de la famille" value="'+esc(label)+'" style="width:100%;min-width:0;padding:6px 8px;font-size:13px;font-weight:700">'
      +'<span class="sub">'+f.count+' produit'+(f.count>1?'s':'')+' · catalogue (lecture seule)</span>'
      +'<details class="menu-sub-collapse"><summary>▸ Sous-menus <span>Modifier avec le crayon</span></summary><div class="menu-sub-list">'+rows+'</div></details>'
      +'</span><label class="switch"><input type="checkbox" '+(f.visible?'checked':'')+' onchange="toggleFam(\''+esc(f.key)+'\',this.checked)"><span class="tr"></span></label></div>';
  }).join('');
  const css=document.createElement('style');css.textContent=`
    .menu-sub-collapse{margin-top:7px;min-width:0}
    .menu-sub-collapse>summary{cursor:pointer;list-style:none;color:#355b40;font-size:12px;font-weight:700;user-select:none}
    .menu-sub-collapse>summary::-webkit-details-marker{display:none}
    .menu-sub-collapse>summary span{font-weight:400;color:#78817b;margin-left:5px}
    .menu-sub-collapse[open]>summary{margin-bottom:6px}
    .menu-sub-list{max-height:132px;overflow-y:auto;overscroll-behavior:contain;border:1px solid #e5e7eb;border-radius:8px;background:#fff;padding:3px 7px}
    .menu-sub-row{display:flex;align-items:center;gap:7px;min-height:32px;border-bottom:1px solid #f0f1f2}
    .menu-sub-row:last-child{border-bottom:0}
    .menu-sub-icon{color:#718078;font-size:17px;flex:0 0 12px}
    .menu-sub-value{font-size:12px;flex:1;min-width:0;overflow-wrap:anywhere}
    .menu-sub-input{font-size:12px;min-width:0;flex:1;padding:5px 7px;border:1px solid #cbd5ce;border-radius:6px}
    .menu-sub-edit{border:0;background:transparent;color:#42634a;cursor:pointer;border-radius:5px;padding:4px 7px;font-size:16px;line-height:1}
    .menu-sub-edit:hover{background:#edf4ee}
  `;document.head.appendChild(css);
  el.querySelectorAll('[data-menu-label]').forEach(input=>input.addEventListener('input',e=>setT(e.currentTarget.dataset.menuLabel,e.currentTarget.value)));
  el.querySelectorAll('[data-sub-edit]').forEach(button=>button.addEventListener('click',()=>{
    const row=button.closest('.menu-sub-row'),input=row.querySelector('.menu-sub-input'),value=row.querySelector('.menu-sub-value');
    const editing=input.hidden;input.hidden=!editing;value.hidden=editing;
    button.textContent=editing?'✓':'✎';button.title=editing?'Terminer la modification':'Modifier';
    button.setAttribute('aria-label',editing?'Terminer la modification':'Modifier le sous-menu');
    if(editing){input.focus();input.select()}
  }));
};
pvMenu=function(){
  const f=front,fams=menu.fams.filter(x=>x.visible);
  PV.html='<div style="background:'+f.bg+';color:'+f.text+';font-family:'+f.font+',system-ui,sans-serif">'+pvHeaderHTML()
    +'<div class="pv-body"><div class="mega-prev"><div class="m-bar" style="background:'+f.accent+'">'+esc(T('toutes_cat'))+'</div><div class="m-grid">'
    +fams.map(fm=>'<div class="m-group"><h5 style="color:'+f.accent+'">'+esc(T('fam_'+fm.key)||fm.label)
      +(menu.showCount?' <span style="color:#9ca3af;font-weight:400">('+fm.count+')</span>':'')
      +'</h5>'+MENU_SUB_ITEMS_V14.map(item=>'<a href="#">'+esc(T('sub_'+fm.key+'_'+item.key)||item.label)+'</a>').join('')
      +(menu.showVoirTous?'<a href="#" style="color:'+f.accent+';font-weight:700">'+esc(T('voir_tous'))+' →</a>':'')
      +'</div>').join('')
    +'</div></div><p style="color:#6b7280;font-size:12px;margin:10px 0 0">'+fams.length+' famille(s) affichée(s) · '+menu.fams.filter(x=>!x.visible).length+' masquée(s).</p></div>'+pvFooterHTML()+'</div>';
};
addEventListener('load',()=>{renderMenuList();renderAllModuleTexts();renderTxtList();renderPreview()});


PAGE_INFO_V16.forEach(p=>{
  [[p.title,p.titleText,'Titre · '+p.name],[p.lead,p.leadText,'Introduction · '+p.name],[p.body,p.bodyText,'Contenu · '+p.name]].forEach(([key,val,label])=>{if(Tx[key]===undefined)Tx[key]=val;if(!MODULE_TEXTS.pagesInfo)MODULE_TEXTS.pagesInfo=[];if(!MODULE_TEXTS.pagesInfo.some(x=>x.key===key))MODULE_TEXTS.pagesInfo.push({key,label})});
  Object.entries(p.form||{}).forEach(([name,key])=>{const val=p.formTexts[name];if(Tx[key]===undefined)Tx[key]=val;if(!MODULE_TEXTS.pagesInfo.some(x=>x.key===key))MODULE_TEXTS.pagesInfo.push({key,label:'Formulaire Contact · '+name})});
  (p.buttons||[]).forEach((key,i)=>{if(Tx[key]===undefined)Tx[key]=p.buttonTexts[i];if(!MODULE_TEXTS.pagesInfo.some(x=>x.key===key))MODULE_TEXTS.pagesInfo.push({key,label:'Bouton · '+p.name+' · '+p.buttonTexts[i]})});
});
MODULE_NAMES.pagesInfo='Pages d’information';
CATEGORIES.find(x=>x.id==='pages').mods.push('pagesInfo');
TABICO.pagesInfo='file';

pagesNavBtn16.type='button';pagesNavBtn16.dataset.tab='pagesInfo';pagesNavBtn16.innerHTML='<span class="nav-ico">'+miniIco('file',4)+'</span>Pages d’information';$('#studioNav').appendChild(pagesNavBtn16);
$('#editorPanel').insertAdjacentHTML('beforeend',`<div id="tab-pagesInfo" class="hidden"><h2>Pages d’information</h2><p class="hint">Modifiez les contenus et la présentation des pages hors boutique. Les formulaires restent des maquettes dans cet aperçu.</p><div class="seg"><label for="v16PageSelect">Page à gérer</label><select id="v16PageSelect">${PAGE_INFO_V16.map(p=>'<option value="'+p.id+'">'+p.name+'</option>').join('')}</select></div><div id="v16PageEditor"></div></div>`);

stylePanelV16.className='field v16-feature-panel';stylePanelV16.innerHTML=`<label>Style fin et identité</label><div class="grid2"><div class="field"><label>Arrondis</label><select id="v16Radius"><option value="square">Droits</option><option value="soft">Doux</option><option value="round">Très arrondis</option></select></div><div class="field"><label>Ombres</label><select id="v16Shadow"><option value="none">Aucune</option><option value="light">Légères</option><option value="strong">Marquées</option></select></div><div class="field"><label>Densité</label><select id="v16Density"><option value="airy">Aérée</option><option value="normal">Normale</option><option value="compact">Compacte</option></select></div><div class="field"><label>Style des boutons</label><select id="v16Buttons"><option value="solid">Plein</option><option value="outline">Contour</option></select></div><div class="field"><label>Police des titres</label><select id="v16TitleFont"><option value="">Même police que le texte</option><option value="Georgia">Georgia</option><option value="Arial">Arial</option><option value="system-ui">Système</option></select></div></div><label>Identité visuelle — URL HTTPS</label><div class="grid2"><div class="field"><label>Logo</label><input id="v16Logo" type="url" placeholder="https://…/logo.png"></div><div class="field"><label>Favicon</label><input id="v16Favicon" type="url" placeholder="https://…/favicon.ico"></div><div class="field"><label>Image de partage</label><input id="v16Share" type="url" placeholder="https://…/partage.jpg"></div><div class="field"><label>Titre de partage</label><input id="v16ShareTitle" type="text"></div></div><div class="field"><label>Description de partage</label><input id="v16ShareDesc" type="text"></div><div id="v16ShareCard" class="v16-share-card"></div><label>Mode sombre</label><div class="grid2"><label class="checkrow"><input id="v16DarkEnabled" type="checkbox"> Proposer le mode sombre aux visiteurs</label><label class="checkrow"><input id="v16DarkToggle" type="checkbox"> Afficher le bouton clair/sombre dans l’en-tête</label><div class="field"><label>Mode par défaut</label><select id="v16DarkDefault"><option value="light">Clair</option><option value="dark">Sombre</option><option value="system">Selon l’appareil</option></select></div><div class="field"><label>Fond sombre</label><input id="v16DarkBg" type="color"></div><div class="field"><label>Texte sombre</label><input id="v16DarkText" type="color"></div><div class="field"><label>Cartes sombres</label><input id="v16DarkCard" type="color"></div><div class="field"><label>Bordures sombres</label><input id="v16DarkBorder" type="color"></div><label class="checkrow"><input id="v16DarkPreview" type="checkbox"> Prévisualiser en sombre</label></div><label>Animations</label><select id="v16Animations"><option value="none">Aucune</option><option value="soft">Douces</option></select><p class="hint">Le navigateur respecte toujours « Réduire les animations » de l’appareil.</p>`;
$('#front-ui').prepend(stylePanelV16);

manageToolsV16.className='field v16-feature-panel';manageToolsV16.innerHTML=`<label>Confort et sauvegarde</label><p class="hint">Ces actions sauvegardent ou échangent un fichier de design. Elles ne publient rien sur le site public.</p><div class="v16-tools"><button type="button" class="btn btn-ghost" id="v16Export">Exporter le design JSON</button><button type="button" class="btn btn-ghost" id="v16Import">Importer un design JSON</button><span id="v16ChangeCount" class="hint">0 module modifié</span></div><input id="v16ImportFile" type="file" accept="application/json,.json" hidden><div class="v16-page-counts" id="v16Counters"></div>`;
$('#tab-reglages').appendChild(manageToolsV16);

v16StyleSheet.textContent=`
 .v16-feature-panel{border:1px solid #dce6de;border-radius:12px;padding:14px;margin:14px 0;background:#fbfdfb}
 .v16-feature-panel>label{display:block;font-weight:800;margin:12px 0 7px}.v16-feature-panel>label:first-child{margin-top:0}
 .v16-tools{display:flex;gap:8px;align-items:center;flex-wrap:wrap}.v16-page-counts{display:flex;flex-wrap:wrap;gap:6px;margin-top:10px}.v16-count-pill{font-size:11px;border-radius:999px;padding:4px 8px;background:#e8f3e9;color:#285337;font-weight:700}.v16-badge,.v16-cat-count{display:inline-flex;align-items:center;justify-content:center;min-width:17px;height:17px;margin-left:5px;padding:0 5px;border-radius:999px;background:#df7c18;color:#fff;font-size:10px;font-weight:800}.v16-badge{min-width:8px;width:8px;height:8px;padding:0;font-size:0}
 .v16-share-card{display:grid;grid-template-columns:120px 1fr;gap:10px;align-items:center;border:1px solid #e1e5e2;border-radius:10px;padding:8px;margin:7px 0;background:#fff}.v16-share-card img{width:120px;height:72px;object-fit:cover;border-radius:6px;background:#f3f4f6}.v16-share-card strong{display:block}.v16-share-card small{color:#6b7280}.v16-visibility{display:flex;gap:14px;flex-wrap:wrap}
 #v16PageEditor{margin-top:12px}.v16-page-form{display:grid;grid-template-columns:1fr 1fr;gap:10px}.v16-page-form .wide{grid-column:1/-1}.v16-page-form textarea{width:100%;min-height:88px;padding:8px;border:1px solid #d1d5db;border-radius:8px;font:inherit}.v16-page-form input[type=text]{width:100%}
 .v16-block-row{display:flex;align-items:center;gap:5px;border-top:1px solid #e8ece9;padding:5px 0}.v16-block-row .checkrow{flex:1;min-width:0}.v16-block-row .checkrow span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.v16-block-row button{border:1px solid #d7dfd9;border-radius:5px;background:#fff;padding:3px 7px;cursor:pointer}.v16-block-row button:disabled{opacity:.35;cursor:default}
 .b-stage.tablet .b-view{width:768px;max-width:100%;margin:0 auto;border-left:1px solid var(--line);border-right:1px solid var(--line)}
 .v16-local-note{margin:8px 0;padding:8px 10px;border:1px solid #f2d28a;border-radius:8px;background:#fff9e9;color:#684c13;font-size:11px;font-weight:600}
 .v16-preview-logo{max-width:150px;max-height:42px;object-fit:contain;vertical-align:middle}.v16-preview-cookie{position:sticky;bottom:0;z-index:50;display:flex;align-items:center;gap:9px;padding:10px 12px;color:#fff;font-size:11px;box-shadow:0 -4px 14px #0002}.v16-preview-cookie span{flex:1}.v16-preview-cookie button{border:1px solid currentColor;border-radius:6px;background:transparent;color:inherit;padding:5px 8px;font-weight:700}
 @media(max-width:740px){.v16-page-form{grid-template-columns:1fr}.v16-page-form .wide{grid-column:auto}.v16-share-card{grid-template-columns:90px 1fr}.v16-share-card img{width:90px}}
 `;document.head.appendChild(v16StyleSheet);

function setV16Input(id,path){const el=$('#'+id);if(!el)return;const [group,key]=path.split('.');el.value=studioV16[group][key];el.addEventListener('input',()=>{studioV16[group][key]=el.value;renderPreview()});el.addEventListener('change',()=>{studioV16[group][key]=el.value;renderPreview()})}

function exportDesignV16(){
 const data={format:'mediatec-studio-design',version:1,exportedAt:new Date().toISOString(),design:{appearance:{accent:front.accent,priceColor:front.priceColor,bg:front.bg,text:front.text,font:front.font,theme:front.theme},header,footer,menu:{showVoirTous:menu.showVoirTous,showCount:menu.showCount,fams:menu.fams.map(f=>({key:f.key,label:T('fam_'+f.key)||f.label,visible:f.visible}))},texts:Tx,studio:studioV16}};
 const blob=new Blob([JSON.stringify(data,null,2)],{type:'application/json'}),url=URL.createObjectURL(blob),a=document.createElement('a');a.href=url;a.download='mediatec-studio-design.json';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);toast('JSON du design exporté. Le site public n’a pas été modifié.');
}

function importDesignV16(file){
 const reader=new FileReader();reader.onload=()=>{try{const d=JSON.parse(reader.result);if(d.format!=='mediatec-studio-design'||d.version!==1||!d.design)throw new Error('Format ou version non reconnus.');const obj=d.design;if(!obj.appearance||!obj.studio||!obj.texts)throw new Error('Fichier incomplet.');if(!confirm('Importer ce design en brouillon ? Les textes et réglages locaux actuels seront remplacés.'))return;Object.assign(front,{accent:obj.appearance.accent,priceColor:obj.appearance.priceColor,bg:obj.appearance.bg,text:obj.appearance.text,font:obj.appearance.font,theme:obj.appearance.theme});studioV16=JSON.parse(JSON.stringify(obj.studio));front._studioV16=studioV16;const allowed=new Set(Object.values(MODULE_TEXTS).flat().map(x=>x.key));Object.keys(Tx).forEach(k=>{if(Object.hasOwn(obj.texts,k)&&allowed.has(k)&&typeof obj.texts[k]==='string')Tx[k]=obj.texts[k]});if(obj.header&&Array.isArray(obj.header.navLinks)){['showTopbar','showPhone','showEmail','showBoutique','showCompte','showPanier','showPrefs','badgeColor','navBg','sticky'].forEach(k=>{if(k in obj.header)header[k]=obj.header[k]});header.navLinks=obj.header.navLinks.filter(x=>x&&typeof x.key==='string').map(x=>({key:x.key,label:String(x.label||''),visible:!!x.visible}))}if(obj.footer){['showNewsletter','showCols','newsletterTitle','newsletterText'].forEach(k=>{if(k in obj.footer)footer[k]=obj.footer[k]})}if(obj.menu){menu.showVoirTous=!!obj.menu.showVoirTous;menu.showCount=!!obj.menu.showCount;(Array.isArray(obj.menu.fams)?obj.menu.fams:[]).forEach(x=>{const f=menu.fams.find(y=>y.key===x.key);if(f){f.visible=!!x.visible;Tx['fam_'+f.key]=String(x.label||f.label)}})}bindV16Controls();renderAllModuleTexts();renderTxtList();renderMenuList();renderPreview();toast('Design importé en brouillon local.')}catch(e){toast('Import impossible : '+e.message)}};reader.readAsText(file);
}

function resetModuleFieldV16(mod,path,baseline){
 const targetPath=path.startsWith('state.')?path.slice(6):path.startsWith('v16.')?path.slice(4):path;
 const value=JSON.parse(JSON.stringify(v16GetPath(baseline[mod],path)??null));
 if(path.startsWith('texts.')){Tx[path.slice(6)]=value;return}
 if(mod==='front'){if(path.startsWith('v16.'))v16SetPath(studioV16,targetPath,value);else v16SetPath(front,targetPath,value)}
 else if(mod==='pagesInfo'){if(path.startsWith('pages.'))v16SetPath(studioV16.pages,path.slice(6),value);else if(path.startsWith('texts.'))Tx[path.slice(6)]=value}
 else if(mod==='accueil'&&targetPath==='rows'){rows=value}
 else {const roots={header,footer,boutique:shop,fiche,panier,menu,compte:acc,seo,reglages},root=roots[mod];if(root)v16SetPath(root,targetPath,value)}
}

const originalRenderPreviewV16=renderPreview;

renderPreview=function(){if(currentTab==='pagesInfo'){updateStatus();setPath();renderInfoPageV16()}else originalRenderPreviewV16();applyV16Preview();const s=$('#statusTxt');if(s)s.textContent=(snap()===pubSnap?'Publié':'Brouillon local · non publié');};

const oldSetPathV16=setPath;

setPath=function(){if(currentTab==='pagesInfo'){const p=PAGE_INFO_V16.find(x=>x.id===pagesInfoV16Active);const el=$('#pvPath');if(el&&p)el.textContent=p.path;return}oldSetPathV16()};

const oldRestoreV16=restore;

restore=function(serialized){oldRestoreV16(serialized);studioV16=front._studioV16||studioV16;bindV16Controls();renderPageEditorV16();renderPreview()};

const originalSetDeviceV16=setDevice;

setDevice=function(mode,button){$('#bStage').classList.toggle('mobile',mode==='mobile');$('#bStage').classList.toggle('tablet',mode==='tablet');document.querySelectorAll('.devices button').forEach(x=>x.classList.toggle('active',x===button));};

function exportV16(){exportDesignV16()}

/* Intégré à l'admin : « Publier » envoie le thème au serveur (voir publishStudio
   ci-dessus). « Exporter » reste disponible dans l'onglet Réglages (#v16Export). */

localNoteV16.className='v16-local-note';localNoteV16.textContent='Studio connecté au site : « Publier » met le thème en ligne. « Exporter » (onglet Réglages) crée un fichier JSON de design.';$('#editorPanel').prepend(localNoteV16);

if(previewNote16)previewNote16.textContent='Aperçu local du brouillon — la publication met le thème en ligne.';
document.querySelectorAll('.devices button').forEach(b=>{if(b.title==='Mobile')b.title='Aperçu mobile'});

tabletBtnV16.type='button';tabletBtnV16.title='Tablette';tabletBtnV16.setAttribute('aria-label','Aperçu tablette');tabletBtnV16.innerHTML=ic('monitor');tabletBtnV16.onclick=()=>setDevice('tablet',tabletBtnV16);$('.devices').appendChild(tabletBtnV16);
$('#v16Export').onclick=exportDesignV16;$('#v16Import').onclick=()=>$('#v16ImportFile').click();$('#v16ImportFile').onchange=e=>{const file=e.target.files&&e.target.files[0];if(file)importDesignV16(file);e.target.value=''};
addEventListener('load',()=>{
  pvI.addEventListener('load',()=>applyV16Preview());
  bindV16Controls();
  const switcher=document.querySelector('#studioNav button[data-tab="pagesInfo"]');if(switcher){switcher.addEventListener('click',()=>{document.querySelectorAll('#editorPanel > div').forEach(d=>d.classList.add('hidden'));$('#tab-pagesInfo').classList.remove('hidden');currentTab='pagesInfo';renderPageEditorV16();renderPreview()})}
  const oldHint=document.querySelector('#tab-front > .hint');if(oldHint)oldHint.textContent='Le Studio règle l’apparence et les textes du site. L’aperçu est une maquette locale ; « Publier » envoie le thème sur le site.';
  $('#tab-pagesInfo').querySelector('h2').insertAdjacentHTML('afterend','<div class="v16-local-note">Les pages d’information sont éditées ici ; leur contenu sera publié avec le design en phase 2.</div>');
  const historyBox=$('#v13History');if(historyBox){const label=historyBox.querySelector(':scope > label'),hint=historyBox.querySelector(':scope > .hint');if(label)label.textContent='Historique des publications';if(hint)hint.textContent='Chaque publication enregistre les modules modifiés (thème envoyé au serveur).'}
  renderAllModuleTexts();renderTxtList();renderMenuList();
  renderPreview();
  /* Référence publiée après le premier rendu v16 : états et textes dynamiques inclus. */
  BASE=ORIGIN=pubSnap=snap();hist=[BASE];hi=0;histBtns();updateStatus();renderCountersV16();
});

