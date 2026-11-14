/* view.js — Studio Mediatec (MVC). Généré par découpe de studio-prototype-v16-2.html. */

function renderModuleTexts(mod){
  const el = document.querySelector('#mt-' + mod); if(!el) return;
  const list = MODULE_TEXTS[mod] || [];
  el.innerHTML = list.map((t,i)=>`
    <div class="txt-item">
      ${miniIco('type', i)}
      <span class="k">${t.label}</span>
      <input type="text" value="${esc(Tx[t.key])}" data-k="${t.key}" oninput="setT('${t.key}',this.value)">
    </div>`).join('');
}

function renderAllModuleTexts(){
  Object.keys(MODULE_TEXTS).forEach(renderModuleTexts);
  renderAccueilTexts();
}

function renderAccueilTexts(){
  const el = document.querySelector('#mt-accueil'); if(!el) return;
  let i = 0;
  el.innerHTML = rows.map((r,ri)=>r.cols.map((c,ci)=>{
    const b = c.block;
    /* oninput : met à jour le bloc + l'aperçu, SANS re-créer le champ (sinon on perd le curseur) ;
       onchange (fin de saisie) : rafraîchit l'éditeur visuel */
    return `<div class="txt-item">${miniIco('type', i++)}<span class="k">${esc(r.label)} — bloc ${ci+1} · titre</span>
      <input type="text" value="${esc(b.title)}" oninput="rows[${ri}].cols[${ci}].block.title=this.value;renderPreview()" onchange="renderRows()"></div>
    <div class="txt-item">${miniIco('type', i++)}<span class="k">${esc(r.label)} — bloc ${ci+1} · texte</span>
      <input type="text" value="${esc(b.text)}" oninput="rows[${ri}].cols[${ci}].block.text=this.value;renderPreview()" onchange="renderRows()"></div>`;
  }).join('')).join('');
}

function toast(msg){
  const t = $('#toast'), ok = msg.startsWith('✅');
  t.innerHTML = ic(ok?'check-circle':'info') + '<span>' + esc(msg.replace(/^✅\s*/,'')) + '</span>'; t.classList.add('show');
  clearTimeout(t._h); t._h = setTimeout(()=>t.classList.remove('show'), 2600);
}

function swatchRow(elId, key, list){
  $(elId).innerHTML = list.map(p=>
    `<button class="swatch ${p.hex===front[key]?'active':''}" title="${esc(p.name)} ${p.hex}" aria-label="${esc(p.name)}" style="background:${p.hex}" onclick="setFront('${key}','${p.hex}')"></button>`
  ).join('');
}

function renderSwatches(){
  swatchRow('#swAccent','accent',PALETTE);
  swatchRow('#swPrice','priceColor',PALETTE);
  swatchRow('#swBg','bg',BGS);
  swatchRow('#swFoot','footerBg',PALETTE);
}

function highlight(code, lang){
  let h = esc(code);
  if(lang==='css'){
    h = h.replace(/(\/\*[\s\S]*?\*\/)/g, '\u0001$1\u0002');
    h = h.replace(/^([^\n{}]+?)(?=\s*\{)/gm, '<span class="tk-sel">$1</span>');
    h = h.replace(/([a-zA-Z-]+)(\s*:)/g, '<span class="tk-prop">$1</span>$2');
    h = h.replace(/(#[0-9a-fA-F]{3,6})/g, '<span class="tk-val">$1</span>');
    h = h.replace(/\u0001([\s\S]*?)\u0002/g, '<span class="tk-com">$1</span>');
  } else if(lang==='html'){
    h = h.replace(/(&lt;\/?[a-zA-Z][^&]*?&gt;)/g, '<span class="tk-tag">$1</span>');
  } else if(lang==='js'){
    h = h.replace(/(\/\/[^\n]*)/g, '\u0001$1\u0002');
    h = h.replace(/("[^"\n]*"|'[^'\n]*')/g, '<span class="tk-str">$1</span>');
    h = h.replace(/\b(function|const|let|var|return|if|else|document|addEventListener)\b/g, '<span class="tk-kw">$1</span>');
    h = h.replace(/\u0001([^\u0002]*)\u0002/g, '<span class="tk-com">$1</span>');
  }
  return h;
}

const codeInput = $('#codeInput'), codePre = $('#codePre'), codeGutter = $('#codeGutter');

function syncEditor(){
  const v = codeInput.value;
  codePre.innerHTML = highlight(v, FILES[currentFile].lang) + '\n';
  const n = v.split('\n').length;
  codeGutter.textContent = Array.from({length:n},(_,i)=>i+1).join('\n');
  updateCodeMeta();
}

function renderTabs(){
  $('#codeTabs').innerHTML = openTabs.map(n=>
    `<span class="code-tab ${n===currentFile?'active':''}" onclick="openFile('${n}')">${fileIcon(n)}${n}<span class="x" onclick="event.stopPropagation();closeTab('${n}')">✕</span></span>`
  ).join('');
  document.querySelectorAll('.ft-file').forEach(f=>f.classList.toggle('active', f.dataset.file===currentFile));
}

function updateCodeMeta(){
  const f = FILES[currentFile]; if(!f) return;
  const n = (codeInput.value || '').split('\n').length;
  $('#codeMeta').textContent = f.path + ' — ' + f.lang.toUpperCase() + ' — ' + n + ' lignes';
}

function openFile(name){
  if(!openTabs.includes(name)) openTabs.push(name);
  currentFile = name; renderTabs(); loadFile(name);
}

function closeTab(name){
  openTabs = openTabs.filter(n=>n!==name);
  if(!openTabs.length) openTabs = ['theme.css'];
  if(currentFile===name) currentFile = openTabs[openTabs.length-1];
  renderTabs(); loadFile(currentFile);
}

function loadFile(name){
  const f = FILES[name];
  codeInput.value = f.content;
  syncEditor();
  updateCodeMeta();
  codeInput.scrollTop = 0; codePre.scrollTop = 0; codeGutter.scrollTop = 0;
}

let sel = {ri:null, ci:null};

function blockHTML(b, ri, ci){ return emojiToSvg(blockHTML0(b, ri, ci)); }

function blockHTML0(b, ri, ci){
  const cta = b.cta ? `<span class="cta">${esc(b.cta)}</span>` : '';
  const price = b.price ? `<div class="price">${esc(b.price)}</div>` : '';
  const isSel = sel.ri===ri && sel.ci===ci;
  return `<div class="block${isSel?' selected':''}" onclick="selectBlock(${ri},${ci})" title="Cliquer pour modifier ce bloc">
    <span class="badge">${esc(b.type)}</span>
    <div class="imgph">${b.img||'📦'}</div>
    <h4>${esc(b.title)}</h4><p>${esc(b.text)}</p>${price}${cta}
  </div>`;
}

function renderRows(){
  $('#rows').innerHTML = rows.map((r, ri)=>{
    const cols = r.cols.map((c, ci)=>{
      const handle = ci < r.cols.length-1
        ? `<div class="resize-handle" data-row="${ri}" data-col="${ci}" title="Tirer pour ajuster la largeur"></div>` : '';
      return `<div class="col" style="flex:0 0 ${c.w}%">
          <span class="col-w">${c.w}%</span>${blockHTML(c.block, ri, ci)}
        </div>${handle}`;
    }).join('');
    return `<div class="row" draggable="true" data-row="${ri}">
      <div class="row-head">
        <span class="grip" title="Glisser pour déplacer">⋮⋮</span>
        <span class="row-label">${esc(r.label)}</span>
        <span class="spacer"></span>
        <button class="icon-btn" onclick="moveRow(${ri},-1)" title="Monter" aria-label="Monter">${ic('up')}</button>
        <button class="icon-btn" onclick="moveRow(${ri},1)" title="Descendre" aria-label="Descendre">${ic('down')}</button>
        <button class="icon-btn danger" onclick="delRow(${ri})" title="Supprimer la ligne" aria-label="Supprimer la ligne">${ic('x')}</button>
      </div>
      <div class="row-body">${cols}</div>
      <div class="row-foot">
        <button class="mini-btn" onclick="addCol(${ri})">${ic('plus')} Ajouter une colonne</button>
        <button class="mini-btn" onclick="delCol(${ri})">${ic('minus')} Retirer la dernière colonne</button>
      </div>
    </div>`;
  }).join('');
  bindDnD(); bindResize();
  renderAccueilTexts();
  renderPreview();
}

function moveRow(i, d){
  const j = i + d; if(j < 0 || j >= rows.length) return;
  [rows[i], rows[j]] = [rows[j], rows[i]]; renderRows();
}

function delRow(i){ rows.splice(i,1); sel={ri:null,ci:null}; $('#blockEditor').classList.add('hidden'); renderRows(); toast('Ligne supprimée'); }

function addRow(){
  rows.push({id:'r'+(rowSeq++), label:'Nouvelle ligne', cols:[
    {w:100, block:{type:'Texte', title:'Titre de la ligne', text:'Écrivez votre texte ici…', img:'✏️'}}
  ]});
  renderRows();
}

function addCol(ri){
  const r = rows[ri]; if(r.cols.length >= 4){ toast('4 colonnes maximum par ligne'); return; }
  const share = Math.floor(100 / (r.cols.length + 1));
  r.cols.forEach(c=>c.w = share);
  r.cols.push({w: 100 - share*r.cols.length, block:{type:'Produit', title:'Nouveau bloc', text:'Produit ou texte…', price:'Sur demande', img:'📦'}});
  renderRows();
}

function delCol(ri){
  const r = rows[ri]; if(r.cols.length <= 1){ toast('Une ligne a au moins une colonne'); return; }
  r.cols.pop();
  const share = Math.floor(100 / r.cols.length);
  r.cols.forEach((c,i)=>c.w = i===r.cols.length-1 ? 100 - share*(r.cols.length-1) : share);
  sel={ri:null,ci:null}; $('#blockEditor').classList.add('hidden');
  renderRows();
}

function selectBlock(ri, ci){
  sel = {ri, ci};
  const b = rows[ri].cols[ci].block;
  $('#beType').value = b.type;
  $('#beTitle').value = b.title || '';
  $('#beText').value = b.text || '';
  $('#bePrice').value = b.price || '';
  $('#beCta').value = b.cta || '';
  $('#blockEditor').classList.remove('hidden');
  renderRows();
  $('#blockEditor').scrollIntoView({behavior:'smooth', block:'nearest'});
}

function updateBlock(field, value){
  if(sel.ri===null) return;
  rows[sel.ri].cols[sel.ci].block[field] = value;
  renderRows();
}

let dragIdx = null;

function bindDnD(){
  document.querySelectorAll('.row').forEach(el=>{
    el.addEventListener('dragstart', e=>{ dragIdx = +el.dataset.row; el.classList.add('dragging'); e.dataTransfer.effectAllowed='move'; });
    el.addEventListener('dragend', ()=>{ dragIdx = null; document.querySelectorAll('.row').forEach(x=>x.classList.remove('dragging','drop-before')); });
    el.addEventListener('dragover', e=>{ e.preventDefault(); el.classList.add('drop-before'); });
    el.addEventListener('dragleave', ()=>el.classList.remove('drop-before'));
    el.addEventListener('drop', e=>{
      e.preventDefault(); const to = +el.dataset.row;
      if(dragIdx===null || dragIdx===to) return;
      const [moved] = rows.splice(dragIdx,1);
      rows.splice(to,0,moved);
      renderRows(); toast('Ligne déplacée');
    });
  });
}

function bindResize(){
  document.querySelectorAll('.resize-handle').forEach(h=>{
    h.addEventListener('pointerdown', e=>{
      e.preventDefault();
      const ri = +h.dataset.row, ci = +h.dataset.col;
      const rowEl = h.closest('.row-body');
      const startX = e.clientX;
      const left = rows[ri].cols[ci], right = rows[ri].cols[ci+1];
      const w0 = left.w, w1 = right.w;
      const rowW = rowEl.getBoundingClientRect().width;
      h.classList.add('active'); h.setPointerCapture(e.pointerId);
      const onMove = ev=>{
        const dPct = (ev.clientX - startX) / rowW * 100;
        let nl = Math.round(w0 + dPct);
        nl = Math.max(15, Math.min(85, nl));
        const nr = w0 + w1 - nl;
        if(nr < 15 || nr > 85) return;
        left.w = nl; right.w = nr;
        const cols = rowEl.querySelectorAll('.col');
        cols[ci].style.flexBasis = nl + '%'; cols[ci+1].style.flexBasis = nr + '%';
        cols[ci].querySelector('.col-w').textContent = nl + '%';
        cols[ci+1].querySelector('.col-w').textContent = nr + '%';
      };
      const onUp = ()=>{
        h.classList.remove('active');
        h.removeEventListener('pointermove', onMove);
        h.removeEventListener('pointerup', onUp);
        renderPreview();
      };
      h.addEventListener('pointermove', onMove);
      h.addEventListener('pointerup', onUp);
    });
  });
}

const ICONS = {
  sparkle:'<path d="M12 2.7s6.5 7 6.5 11.3a6.5 6.5 0 1 1-13 0C5.5 9.7 12 2.7 12 2.7z"/>',
  layout:'<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/>',
  code:'<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>',
  home:'<path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V21h14V9.5"/><path d="M9.5 21v-6h5v6"/>',
  bag:'<path d="M6 7h12l1.2 14H4.8L6 7z"/><path d="M9 10V6a3 3 0 0 1 6 0v4"/>',
  file:'<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6z"/><path d="M14 2v6h6"/><path d="M9 13h6M9 17h6"/>',
  cart:'<path d="M3 3h2l2.6 12.6a1.5 1.5 0 0 0 1.5 1.1h8.7a1.5 1.5 0 0 0 1.5-1.2L21 7H6"/><circle cx="9.5" cy="20.5" r="1.4"/><circle cx="17.5" cy="20.5" r="1.4"/>',
  pencil:'<path d="M17 3a2.83 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/>',
  list:'<path d="M9 6h12M9 12h12M9 18h12"/><path d="M4 6h.01M4 12h.01M4 18h.01"/>',
  search:'<circle cx="11" cy="11" r="7"/><path d="m20.5 20.5-4.2-4.2"/>',
  gear:'<circle cx="12" cy="12" r="3.2"/><path d="M12 2.8v2.7M12 18.5v2.7M2.8 12h2.7M18.5 12h2.7M5.2 5.2l1.9 1.9M16.9 16.9l1.9 1.9M18.8 5.2l-1.9 1.9M7.1 16.9l-1.9 1.9"/>',
  user:'<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
  tag:'<path d="M20.6 13.4 11 3.8A2 2 0 0 0 9.6 3.2H4a1 1 0 0 0-1 1v5.6c0 .5.2 1 .6 1.4l9.6 9.6a2 2 0 0 0 2.8 0l4.6-4.6a2 2 0 0 0 0-2.8z"/><circle cx="7.5" cy="7.5" r="1.2" fill="currentColor"/>',
  grid:'<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
  image:'<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-4.5-4.5L6 21"/>',
  type:'<path d="M4 7V4h16v3M9 20h6M12 4v16"/>',
  check:'<path d="M20 6 9 17l-5-5"/>',
  star:'<path d="m12 2 3.1 6.3 6.9 1-5 4.9 1.2 6.8L12 17.8 5.8 21l1.2-6.8-5-4.9 6.9-1L12 2z"/>',
};

const ICO_COLORS = [
  ['#dbeafe','#2563eb'], ['#ede9fe','#7c3aed'], ['#fee2e2','#dc2626'], ['#ffedd5','#ea580c'],
  ['#dcfce7','#16a34a'], ['#ccfbf1','#0d9488'], ['#fce7f3','#db2777'], ['#fef3c7','#d97706'],
];

function miniIco(name, i){
  const c = ICO_COLORS[i % ICO_COLORS.length];
  return `<span class="mini-ico" style="background:${c[0]};color:${c[1]}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${ICONS[name]||''}</svg></span>`;
}

function fileIcon(name){
  const ext = name.split('.').pop().toLowerCase();
  const map = {html:['#ffedd5','#ea580c'], css:['#dbeafe','#2563eb'], js:['#fef3c7','#b45309']};
  const c = map[ext] || ['#e5e7eb','#6b7280'];
  return `<span class="ft-ico" style="background:${c[0]};color:${c[1]}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6z"/><path d="M14 2v6h6"/></svg></span>`;
}

function makeSortable(listEl, onReorder){
  let dragEl = null;
  listEl.addEventListener('dragstart', e=>{
    const it = e.target.closest('.sort-item'); if(!it) return;
    dragEl = it; it.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
    try{ e.dataTransfer.setData('text/plain', it.dataset.k); }catch(_){}
  });
  listEl.addEventListener('dragend', ()=>{
    if(dragEl) dragEl.classList.remove('dragging');
    listEl.querySelectorAll('.sort-item').forEach(x=>x.classList.remove('drop-before'));
    dragEl = null;
  });
  listEl.addEventListener('dragover', e=>{
    e.preventDefault();
    const it = e.target.closest('.sort-item'); if(!it || it===dragEl) return;
    listEl.querySelectorAll('.sort-item').forEach(x=>x.classList.remove('drop-before'));
    it.classList.add('drop-before');
  });
  listEl.addEventListener('drop', e=>{
    e.preventDefault();
    const it = e.target.closest('.sort-item'); if(!it || !dragEl) return;
    onReorder(dragEl.dataset.k, it.dataset.k);
  });
}

function renderFacets(){
  const el = $('#facetList'); if(!el) return;
  const ico = {marque:'tag', categorie:'grid', prix:'star', dispo:'check'};
  el.innerHTML = shop.facets.map((f,i)=>`
    <div class="sort-item" draggable="true" data-k="${f.key}">
      <span class="grip">⋮⋮</span>
      ${miniIco(ico[f.key]||'tag', i)}
      <span class="lbl">${f.label}<span class="sub">${f.sub}</span></span>
      <label class="switch"><input type="checkbox" ${f.visible?'checked':''} onchange="toggleFacet('${f.key}',this.checked)"><span class="tr"></span></label>
    </div>`).join('');
}

function toggleFacet(key, on){
  const f = shop.facets.find(x=>x.key===key); if(f) f.visible = on;
  renderPreview();
}

function renderStepEditor(){
  const ed = $('#stepEditor'); if(!ed) return;
  ed.innerHTML = panier.steps.map((s,i)=>`
    <div><label style="font-size:11px;color:var(--muted)">Étape ${i+1}</label>
    <input type="text" value="${esc(s)}" data-i="${i}" oninput="panier.steps[this.dataset.i]=this.value;renderStepPreview();renderPreview()" style="padding:7px 9px;font-size:13px"></div>`).join('');
  renderStepPreview();
}

function renderStepPreview(){
  const pv = $('#stepPreview'); if(!pv) return;
  pv.innerHTML = panier.steps.map((s,i)=>`
    <div style="display:flex;flex-direction:column;align-items:center;flex:0 0 26px">
      <div class="step-dot ${i===0?'':'off'}">${i+1}</div>
      <div class="step-lbl">${esc(s)}</div>
    </div>${i<panier.steps.length-1?`<div class="step-line ${i===0?'on':''}" style="margin-bottom:18px"></div>`:''}`).join('');
}

function renderTxtList(){
  const q = ($('#txtSearch')?.value||'').toLowerCase();
  const el = $('#txtList'); if(!el) return;
  let html = '', n = 0, i = 0;
  Object.keys(MODULE_TEXTS).forEach(mod=>{
    const list = MODULE_TEXTS[mod].filter(t=>{
      const v = T[t.key] || '';
      return !q || t.label.toLowerCase().includes(q) || v.toLowerCase().includes(q);
    });
    if(!list.length) return;
    html += `<div style="font-size:11px;font-weight:800;text-transform:uppercase;letter-spacing:.6px;color:var(--muted);margin:14px 0 4px">${MODULE_NAMES[mod]||mod}</div>`;
    html += list.map(t=>{
      n++;
      return `<div class="txt-item">
      ${miniIco('type', i++)}
      <span class="k">${t.label}</span>
      <input type="text" value="${esc(Tx[t.key])}" data-k="${t.key}" oninput="setT('${t.key}',this.value)">
    </div>`;
    }).join('');
  });
  el.innerHTML = n ? html : '<p class="hint">Aucun texte ne correspond à « '+esc(q)+' ».</p>';
  applyTxtToFront();
}

function renderMenuList(){
  const el = $('#menuList'); if(!el) return;
  const ico = {reseaux:'grid', accessoires:'tag', imprimantes:'file', laptops:'home', bureau:'list', serveurs:'gear', cameras:'image'};
  el.innerHTML = menu.fams.map((f,i)=>`
    <div class="sort-item" draggable="true" data-k="${f.key}">
      <span class="grip">⋮⋮</span>
      ${miniIco(ico[f.key]||'grid', i)}
      <span class="lbl">${f.label}<span class="sub">${f.count} produit${f.count>1?'s':''}</span></span>
      <label class="switch"><input type="checkbox" ${f.visible?'checked':''} onchange="toggleFam('${f.key}',this.checked)"><span class="tr"></span></label>
    </div>`).join('');
}

function toggleFam(key, on){
  const f = menu.fams.find(x=>x.key===key); if(f) f.visible = on;
  renderPreview();
}

function renderNavLinks(){
  const el = $('#navList'); if(!el) return;
  el.innerHTML = header.navLinks.map((l,i)=>`
    <div class="sort-item" draggable="true" data-k="${l.key}">
      <span class="grip">⋮⋮</span>
      ${miniIco('list', i)}
      <input type="text" value="${esc(l.label)}" style="flex:1;padding:7px 9px;font-size:13px" oninput="header.navLinks[${i}].label=this.value;renderPreview()">
      <label class="switch"><input type="checkbox" ${l.visible?'checked':''} onchange="header.navLinks[${i}].visible=this.checked;renderPreview()"><span class="tr"></span></label>
    </div>`).join('');
}

function renderFooterCols(){
  const el = $('#footerCols'); if(!el) return;
  el.innerHTML = footer.cols.map((c,i)=>`
    <div style="border:1px solid var(--line);border-radius:12px;padding:12px;margin-bottom:10px;background:#fff">
      <div style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
        <input type="text" value="${esc(c.title)}" style="flex:1;padding:7px 9px;font-size:13px;font-weight:700" oninput="footer.cols[${i}].title=this.value;renderPreview()">
        <label class="switch"><input type="checkbox" ${c.visible?'checked':''} onchange="footer.cols[${i}].visible=this.checked;renderPreview()"><span class="tr"></span></label>
      </div>
      <textarea rows="${c.links.length+1}" style="width:100%;padding:8px 10px;font-size:13px;font-family:inherit;border:1px solid rgba(15,26,20,.12);border-radius:10px" oninput="footer.cols[${i}].links=this.value.split('\\n').map(s=>s.trim()).filter(Boolean);renderPreview()">${esc(c.links.join('\n'))}</textarea>
      <div style="font-size:11px;color:var(--muted);margin-top:4px">Un lien par ligne.</div>
    </div>`).join('');
}

function renderSeoCounters(){
  const t = $('#cntTitle'), d = $('#cntDesc');
  if(t){ const n = seo.title.length; t.textContent = n+' / 60 caractères'; t.classList.toggle('over', n>60); }
  if(d){ const n = seo.desc.length; d.textContent = n+' / 160 caractères'; d.classList.toggle('over', n>160); }
}

function btnRadius(){
  return front.btnShape==='carre' ? '4px' : front.btnShape==='pilule' ? '999px' : '8px';
}

function titleSize(){
  return front.titleSize==='S' ? '15px' : front.titleSize==='L' ? '23px' : '19px';
}

function setPath(){const p=$('#pvPath'); if(p)p.textContent=currentTab==='textes'?'/'+txtPageSel+'.html':(PATHS[currentTab]||'/')}

function syncUI(){
  const R={front,shop,fiche,panier,menu,seo,reglages,header,footer,acc};
  document.querySelectorAll('[data-bind]').forEach(el=>{const [o,k]=el.dataset.bind.split('.');const v=R[o][k];if(el.type==='checkbox')el.checked=!!v;else el.value=v??''});
  Object.entries(SEGS).forEach(([id,p])=>{const [o,k]=p.split('.');document.querySelectorAll('#'+id+' button').forEach(b=>b.classList.toggle('active',b.dataset.v===String(R[o][k])))});
  $('#maintField').style.display=reglages.maintenance?'':'none';
  renderSwatches();renderRows();renderFacets();renderMenuList();renderTxtList();renderStepEditor();renderSeoCounters();renderAllModuleTexts();renderNavLinks();renderFooterCols();typeof renderAccEditor==='function'&&renderAccEditor();
  $('#selTxtPage').value=txtPageSel;loadTxtPage();
  if(!$('#front-code').classList.contains('hidden')){renderTabs();loadFile(currentFile)}
}

function updateCrumb(b){const t=b.dataset.tab;$('#crumbTitle').textContent=b.textContent.trim();$('#crumbIco').innerHTML=miniIco(TABICO[t],Object.keys(TABICO).indexOf(t))}

function groupTexts(m){
 const g=TG[m], el=document.querySelector('#mt-'+m); if(!g||!el)return;
 const items=[...el.querySelectorAll('.txt-item')]; if(!items.length)return;
 const by={}; items.forEach(i=>by[i.querySelector('input').dataset.k]=i);
 const used=new Set(), list=g.map(x=>[x[0],x[1].filter(k=>by[k])]);
 const rest=Object.keys(by).filter(k=>!g.some(x=>x[1].includes(k))); if(rest.length)list.push(['Autres',rest]);
 el.innerHTML='';
 list.forEach((x,gi)=>{
  if(!x[1].length)return;
  const d=document.createElement('details'); d.className='tx-grp'; d.dataset.g=m+'|'+x[0];
  d.open=grpOpen.has(d.dataset.g)||(gi===0&&![...grpOpen].some(k=>k.startsWith(m+'|')));
  d.innerHTML='<summary>'+x[0]+'<span class="n">'+x[1].length+'</span></summary>';
  x[1].forEach(k=>d.appendChild(by[k]));
  d.addEventListener('toggle',()=>{d.open?grpOpen.add(d.dataset.g):grpOpen.delete(d.dataset.g)});
  el.appendChild(d);
 });
}

function buildEd(items,host){
 let box=null;
 items.forEach(it=>{
  if(it[0]==='g'){box=document.createElement('details');box.className='tx-grp acc-sec';box.open=false;box.innerHTML='<summary>'+it[1]+'</summary>';host.appendChild(box);return}
  const o=root[it[2]](),p=it[3],v=getP(o,p),w=document.createElement('label');
  if(it[0]==='c'){w.className='fld';w.innerHTML='<span>'+it[1]+'</span><input type="color">';const i=w.querySelector('input');i.value=v||'#dc2626';i.oninput=()=>{setP(root[it[2]](),p,i.value);renderPreview()}}
  else if(it[0]==='s'){w.className='fld';w.innerHTML='<span>'+it[1]+'</span><select>'+it[4].map(x=>'<option value="'+x[0]+'">'+x[1]+'</option>').join('')+'</select>';const i=w.querySelector('select');i.value=v;i.onchange=()=>{setP(root[it[2]](),p,i.value);renderPreview()}}
  else if(it[0]==='b'){w.className='brow';w.innerHTML='<span>'+it[1]+'</span><input type="checkbox">';const i=w.querySelector('input');i.checked=!!v;i.onchange=()=>{setP(root[it[2]](),p,i.checked);renderPreview()}}
  else{w.className='fld';const a=it[0]==='a';w.innerHTML='<span>'+it[1]+'</span>'+(a?'<textarea rows="4"></textarea>':'<input type="text">');const i=w.querySelector(a?'textarea':'input');i.value=a?(v||[]).join('\n'):(v??'');
   i.oninput=()=>{setP(root[it[2]](),p,a?i.value.split('\n').map(x=>x.trim()).filter(Boolean):i.value);renderPreview()}}
  box.appendChild(w);
 });
}

function waBtn(){return acc.misc&&acc.misc.wa?'<div style="position:fixed;right:14px;bottom:14px;width:46px;height:46px;border-radius:50%;background:#25d366;color:#fff;display:flex;align-items:center;justify-content:center;font-size:22px;box-shadow:0 6px 16px rgba(0,0,0,.25);z-index:5">💬</div>':''}

const fldBox=(t)=>`<div style="border:1px solid #e5e7eb;border-radius:8px;padding:8px 10px;margin:6px 0;color:#9ca3af;font-size:12px;background:#fff">${esc(t)}</div>`;

function drawerHTML(){
 const d=acc.drawer,f=front,r=btnRadius(),line=(n,p)=>`<div style="display:flex;justify-content:space-between;gap:8px;padding:8px 0;border-bottom:1px solid #e5e7eb;font-size:12px"><div><b>${n}</b><div style="color:#2563eb;cursor:pointer">${esc(T('tiroir_retirer'))}</div></div><b style="color:${f.priceColor}">${p}</b></div>`;
 return `<div style="position:fixed;inset:0;background:rgba(17,24,39,.45);z-index:20"></div>
 <div style="position:fixed;top:0;right:0;bottom:0;width:min(320px,86%);background:#fff;z-index:21;padding:14px;display:flex;flex-direction:column;box-shadow:-10px 0 30px rgba(0,0,0,.2)">
  <div style="display:flex;justify-content:space-between;align-items:center;font-weight:800"><span>${esc(panier.title)}</span><span style="font-size:12px;font-weight:600;color:#6b7280">${esc(T('tiroir_fermer'))} ✕</span></div>
  <div style="flex:1;margin-top:8px">${line('HP EliteBook 850 G3','185 000 F')}${line('Toners HP 650A','45 000 F')}</div>
  <div style="display:flex;justify-content:space-between;font-weight:800;margin:8px 0"><span>${esc(T('sous_total'))}</span><span style="color:${f.priceColor}">230 000 F</span></div>
  ${d.livraison?`<p style="font-size:11px;color:#6b7280;margin:0 0 8px">${esc(T('tiroir_livraison'))}</p>`:''}
  <span style="background:${f.accent};color:#fff;text-align:center;border-radius:${r};padding:10px;font-weight:700;font-size:13px">${esc(panier.btnOrder)}</span>
  ${d.continuer?`<span style="text-align:center;padding:8px;font-size:12px;color:${f.accent};font-weight:600">${esc(T('tiroir_continuer'))}</span>`:''}
  ${d.vider?`<span style="text-align:center;font-size:12px;color:#dc2626">${esc(T('tiroir_vider'))}</span>`:''}
 </div>
 <div style="position:fixed;left:14px;bottom:14px;background:#111827;color:#fff;border-radius:10px;padding:10px 14px;font-size:12px;z-index:22">✓ ${esc(T('toast_ajoute'))} HP EliteBook 850 G3</div>`;
}

function renderAccEditor(){
 [['accD','accueil',ACC_D,0],['accT','accueil',ACC_T,1],['ficheD','fiche',FICHE_D,0],['panD','panier',PAN_D,0],['cptD','compte',CPT_D,0]].forEach(([id,tab,items,afterMt])=>{
  let h=document.getElementById('x-'+id);
  if(!h){h=document.createElement('div');h.id='x-'+id;const mt=document.getElementById('mt-'+tab);if(!mt)return;
   if(afterMt){h.innerHTML='<div class="acc-title">Sections précises de l\u2019accueil</div>';mt.after(h)}else{h.innerHTML='<div class="acc-title">'+({accueil:'Sections précises de l\u2019accueil',fiche:'Détails de la fiche',panier:'Tiroir, devis et divers',compte:'Compte'})[tab]+'</div>';mt.parentElement.before(h)}}
  const openIdx=[...h.querySelectorAll('details')].map(d=>d.open);
  h.querySelectorAll('details').forEach(d=>d.remove()); buildEd(items,h);
  h.querySelectorAll('details').forEach((d,i)=>{if(openIdx[i])d.open=true});
 });
}

function accHTML(){
 const a=acc,f=front,bg=f.accent,r=btnRadius(),o=[];
 if(a.hero.show){const sl=a.hero.slides[0];o.push(`<div style="position:relative;background:${sl.img?'url('+esc(sl.img)+') center/cover,':''}linear-gradient(135deg,${bg},#0f2a18);color:#fff;padding:26px 36px;text-align:center">
  ${a.hero.arrows?'<span style="position:absolute;left:10px;top:45%;font-size:20px">‹</span><span style="position:absolute;right:10px;top:45%;font-size:20px">›</span>':''}
  <div style="font-size:${titleSize()};font-weight:800">${esc(sl.t)}</div><p style="margin:6px 0 10px;font-size:13px">${esc(sl.x)}</p>
  <span style="display:inline-block;background:#fff;color:${bg};border-radius:${r};padding:7px 14px;font-weight:700;font-size:12px">${esc(sl.b)}</span>
  ${a.hero.dots?'<div style="margin-top:10px;font-size:10px;letter-spacing:4px">● ○ ○</div>':''}</div>
  ${a.hero.panel?`<div style="background:#fff;padding:10px 16px;border-bottom:1px solid #e5e7eb;font-size:13px"><b>${esc(a.hero.panelT)}</b> — ${esc(a.hero.panelX)}</div>`:''}`)}
 if(a.av.show)o.push(`<div style="display:flex;gap:10px;padding:12px 14px">${a.av.cards.map(c=>`<div style="flex:1;border:1px solid #e5e7eb;border-radius:10px;padding:10px;background:#fff;font-size:12px"><div style="font-size:18px">${esc(c.i)}</div><b>${esc(c.t)}</b><div style="color:#6b7280">${esc(c.x)}</div></div>`).join('')}</div>`);
 if(a.find.show)o.push(`<div style="display:flex;gap:10px;padding:0 14px 12px"><div style="flex:2;border:1px solid #e5e7eb;border-radius:10px;padding:12px;background:#fff;font-size:12px"><b>${esc(a.find.title)}</b>
  <div style="display:flex;gap:6px;margin:8px 0">${[a.find.l1,a.find.l2,a.find.l3].map(x=>`<span style="flex:1;border:1px solid #e5e7eb;border-radius:6px;padding:6px;color:#6b7280">${esc(x)} ▾</span>`).join('')}</div>
  <span style="color:#6b7280;margin-right:10px">${esc(a.find.reset)}</span><span style="background:${bg};color:#fff;border-radius:${r};padding:6px 12px">${esc(a.find.search)}</span></div>
  ${a.find.offer?`<div style="flex:1;border-radius:10px;padding:12px;background:${bg};color:#fff;font-size:12px"><b>${esc(a.find.offT)}</b><div>${esc(a.find.offX)}</div></div>`:''}</div>`);
 const tb=a.tabs.labels.filter((_,i)=>a.tabs.show[i]);
 if(tb.length)o.push(`<div style="display:flex;gap:16px;padding:0 14px;font-size:13px;border-bottom:1px solid #e5e7eb">${tb.map((t,i)=>`<span style="padding:6px 0;${i?'':'border-bottom:2px solid '+bg+';font-weight:700'}">${esc(t)}</span>`).join('')}</div>`);
 a.secs.filter(x=>x.show).forEach(x=>o.push(`<div style="padding:10px 14px"><b style="font-size:14px">${esc(x.t)}</b>${x.s?`<div style="font-size:12px;color:#6b7280">${esc(x.s)}</div>`:''}</div>`));
 if(a.dark.show)o.push(`<div style="background:#111827;color:#fff;padding:18px 20px;text-align:center"><b>${esc(a.dark.t)}</b><p style="margin:4px 0 8px;font-size:12px;color:#d1d5db">${esc(a.dark.x)}</p><span style="background:${bg};border-radius:${r};padding:6px 14px;font-size:12px">${esc(a.dark.b)}</span></div>`);
 if(a.brands.show)o.push(`<div style="display:flex;gap:18px;justify-content:center;padding:12px;font-weight:700;color:#9ca3af;font-size:13px">${a.brands.list.map(esc).join('<span>·</span>')}</div>`);
 return o.join('');
}

function ficheBuyExtras(){
 const f=fiche,m=[];
 if(f.showExpress)m.push('🚚 '+esc(T('express')));if(f.showGarantie)m.push('🛡 '+esc(T('garantie')));if(f.showTva)m.push(esc(T('tva')));
 return (f.showQty?`<div style="margin-top:10px;display:inline-flex;border:1px solid #e5e7eb;border-radius:8px;font-size:13px"><span style="padding:5px 10px">−</span><span style="padding:5px 12px;border-inline:1px solid #e5e7eb">1</span><span style="padding:5px 10px">+</span></div>`:'')
  +(m.length?`<div style="margin-top:10px;font-size:12px;color:#4b5563;display:flex;gap:12px;flex-wrap:wrap">${m.map(x=>`<span>${x}</span>`).join('')}</div>`:'')
  +((f.showShare||f.showLike)?`<div style="margin-top:8px;font-size:12px;color:#6b7280">${f.showShare?esc(T('partager'))+' f · in · wa':''} ${f.showLike?' &nbsp;♡':''}</div>`:'');
}

function ficheTail(){
 if(!fiche.showAvis)return '';
 return `<div style="margin-top:16px"><div style="display:flex;gap:16px;border-bottom:1px solid #e5e7eb;font-size:13px"><span style="padding:6px 0">${esc(T('tab_desc'))}</span><span style="padding:6px 0">${esc(T('tab_marque'))}</span><span style="padding:6px 0;font-weight:700;border-bottom:2px solid ${front.accent}">${esc(T('tab_avis'))}</span></div>
  <p style="font-size:13px;margin:10px 0 2px"><b>${esc(T('aucun_avis'))}</b></p><p style="font-size:12px;color:#6b7280;margin:0 0 10px">${esc(T('premier_avis'))}</p>
  <div style="border:1px solid #e5e7eb;border-radius:10px;padding:12px;font-size:12px"><b>${esc(T('evaluer'))}</b> <span style="color:#f59e0b">☆☆☆☆☆</span>
  <div style="border:1px solid #e5e7eb;border-radius:6px;padding:6px;margin:8px 0;color:#9ca3af">${esc(T('votre_nom'))}</div><div style="border:1px solid #e5e7eb;border-radius:6px;padding:6px;height:44px;color:#9ca3af">${esc(T('votre_avis'))}</div>
  <span style="display:inline-block;margin-top:8px;background:${front.accent};color:#fff;border-radius:${btnRadius()};padding:6px 12px">${esc(T('envoyer_avis'))}</span></div></div>`;
}

function themeGalleryV13(){
 const host=document.createElement('section');host.className='field v13-gallery';
 host.innerHTML='<label>Thèmes prédéfinis</label><p class="hint" style="margin:4px 0 10px">Un thème s’applique en brouillon; vous pouvez ensuite affiner chaque réglage.</p><div class="v13-themes">'+THEMES_V13.map(t=>'<button type="button" class="v13-theme" data-theme="'+t.id+'"><span class="v13-thumb" style="background:'+t.bg+';color:'+t.text+'"><i style="background:'+t.accent+'"></i><em style="background:'+t.price+'"></em></span><b>'+t.title+'</b><small>Aperçu</small></button>').join('')+'</div>';
 host.addEventListener('click',e=>{const b=e.target.closest('[data-theme]');if(b)applyThemeV13(b.dataset.theme)});
 $('#front-ui').prepend(host);
}

function announceEditorV13(){
 const host=document.createElement('section');host.className='field v13-ann';
 host.innerHTML='<label>Bandeau d’annonce — toutes les pages</label><label class="checkrow"><input id="v13AnnOn" type="checkbox"> Afficher le bandeau</label><div class="grid2"><div class="field"><label>Message</label><input id="v13AnnText" type="text" placeholder="Message d’annonce"></div><div class="field"><label>Lien (facultatif)</label><input id="v13AnnUrl" type="text" placeholder="/boutique"></div></div><div class="grid2"><div class="field"><label>Fond</label><input id="v13AnnBg" type="color" value="#2E5136"></div><div class="field"><label>Texte</label><input id="v13AnnFg" type="color" value="#ffffff"></div></div><label class="checkrow"><input id="v13AnnClose" type="checkbox" checked> Afficher une fermeture (aperçu)</label><p class="hint">La fermeture est illustrative dans ce prototype; le bandeau d’annonce reste un élément de design.</p>';
 $('#front-ui').prepend(host);
 const bind=(id,fn)=>$('#'+id).addEventListener('input',()=>{fn();renderPreview()});
 $('#v13AnnOn').addEventListener('change',e=>{front.showAnnonce=e.target.checked;renderPreview()});
 bind('v13AnnText',()=>{front.banner=$('#v13AnnText').value;setT('bandeau',front.banner)});
 bind('v13AnnUrl',()=>front.announcementUrl=$('#v13AnnUrl').value);
 bind('v13AnnBg',()=>front.announcementBg=$('#v13AnnBg').value);
 bind('v13AnnFg',()=>front.announcementFg=$('#v13AnnFg').value);
 bind('v13AnnClose',()=>front.announcementClose=$('#v13AnnClose').checked);
}

function syncV13Fields(){
 if(!$('#v13AnnOn'))return;
 $('#v13AnnOn').checked=!!front.showAnnonce;
 $('#v13AnnText').value=front.banner||T('bandeau')||'';
 $('#v13AnnUrl').value=front.announcementUrl||'';
 $('#v13AnnBg').value=front.announcementBg||front.accent||'#2E5136';
 $('#v13AnnFg').value=front.announcementFg||'#ffffff';
 $('#v13AnnClose').checked=front.announcementClose!==false;
}

function annHtmlV13(){
 if(!front.showAnnonce)return '';
 const label=front.banner||T('bandeau')||'';
 const text=esc(label),bg=front.announcementBg||front.accent,fg=front.announcementFg||'#ffffff';
 const href=front.announcementUrl?esc(front.announcementUrl):'';
 return '<div class="pv-announcement" style="background:'+bg+';color:'+fg+';text-align:center;font-size:12px;font-weight:700;padding:7px 40px;position:relative">'+(href?'<a href="'+href+'" style="color:inherit">'+text+'</a>':text)+(front.announcementClose!==false?'<span aria-hidden="true" style="position:absolute;right:12px;top:4px;font-size:18px">×</span>':'')+'</div>';
}

function contrastDialogV13(issues,proceed){
 let d=$('#v13ContrastDialog');if(!d){d=document.createElement('div');d.id='v13ContrastDialog';d.className='modal-back';d.innerHTML='<div class="modal" role="dialog" aria-modal="true"><div class="modal-ico">!</div><h3>Contraste à vérifier</h3><p class="v13-contrast-list"></p><p>Le seuil conseillé est 4,5:1. Vous pouvez corriger les couleurs ou continuer la publication.</p><div class="modal-act"><button class="btn btn-ghost" data-v13-cancel>Corriger</button><button class="btn btn-publish" data-v13-go>Publier quand même</button></div></div>';document.body.appendChild(d)}
 d.querySelector('.v13-contrast-list').innerHTML=issues.map(x=>'<b>'+esc(x.name)+'</b> — '+x.ratio.toFixed(2)+':1').join('<br>');d.classList.remove('hidden');
 d.querySelector('[data-v13-cancel]').onclick=()=>d.classList.add('hidden');d.querySelector('[data-v13-go]').onclick=()=>{d.classList.add('hidden');proceed()};
}

function historyPanelV13(){
 let reg=$('#tab-reglages');if(!reg)return;const h=reg.querySelector(':scope > h2');if(!h)return;
 const box=document.createElement('section');box.className='field';box.id='v13History';box.innerHTML='<label>Historique des publications</label><p class="hint">Versions publiées sur le serveur (Réactiver recharge la page avec cette version) et publications de cette session (Restaurer remet en brouillon).</p><div class="v13-history-list"><p class="hint">Aucune publication enregistrée.</p></div>';
 h.after(box);
}

function renderHistoryV13(){
 const el=$('#v13History .v13-history-list');if(!el)return;
 let html='';
 /* Versions publiées côté serveur (phase 2) : la source de vérité.
    Réactiver recharge la page avec cette version (design + thème). */
 if(typeof serverVersionsV13!=='undefined'&&serverVersionsV13.length){
  html+='<p class="hint" style="margin:8px 0 4px">Versions publiées (serveur) :</p>';
  html+=serverVersionsV13.map(v=>'<div class="v13-history-row"><span><b>Version '+esc(String(v.version))+'</b>'+(v.actif?' <span class="pill" style="color:#15803d">● active</span>':'')+'<small>'+esc(v.cree_le||'')+(v.auteur?' · par '+esc(v.auteur):'')+'</small></span>'+(v.actif?'':'<button class="btn btn-ghost" data-v13-activate="'+v.version+'">Réactiver</button>')+'</div>').join('');
 }
 /* Historique de session (local) : restaurer remet en brouillon. */
 if(publishedHistoryV13.length){
  html+='<p class="hint" style="margin:12px 0 4px">Cette session (brouillons restaurables) :</p>';
  html+=publishedHistoryV13.slice().reverse().map((v,i)=>'<div class="v13-history-row"><span><b>Publié le '+esc(v.date)+'</b><small>Modules modifiés : '+esc(v.modules.join(', ')||'aucun changement détecté')+'</small></span><button class="btn btn-ghost" data-v13-restore="'+(publishedHistoryV13.length-1-i)+'">Restaurer cette version</button></div>').join('');
 }
 if(!html){el.innerHTML='<p class="hint">Aucune publication enregistrée.</p>';return}
 el.innerHTML=html;
 el.querySelectorAll('[data-v13-activate]').forEach(b=>b.onclick=()=>activateServerVersionV13(+b.dataset.v13Activate));
 el.querySelectorAll('[data-v13-restore]').forEach(b=>b.onclick=()=>{const v=publishedHistoryV13[+b.dataset.v13Restore];if(!v)return;restore(v.snapshot);toast('Version restaurée en brouillon.')});
}

const pagesNavBtn16=document.createElement('button');

const stylePanelV16=document.createElement('section');

const manageToolsV16=document.createElement('section');

const v16StyleSheet=document.createElement('style');

function renderPageEditorV16(){
 const page=PAGE_INFO_V16.find(x=>x.id===pagesInfoV16Active)||PAGE_INFO_V16[0],state=studioV16.pages[page.id];
 const field=(key,label,value,wide=false)=>'<div class="field '+(wide?'wide':'')+'"><label>'+label+'</label><input type="text" data-k="'+key+'" value="'+esc(T(key)||value)+'" aria-label="'+label+'"></div>';
 let extra='';if(page.form)extra='<label class="checkrow"><input type="checkbox" id="v16ContactForm" '+(state.showForm?'checked':'')+'> Afficher le formulaire de contact (présentation seulement)</label><div class="wide"><strong>Libellés du formulaire</strong><div class="v16-page-form">'+Object.entries(page.form).map(([name,key])=>field(key,'Formulaire · '+name,page.formTexts[name])).join('')+'</div></div>';
 if(page.id==='about')extra='<label class="checkrow"><input type="checkbox" id="v16AboutHero" '+(state.showHero?'checked':'')+'> Afficher le bandeau de page</label>';
 if(page.id==='cookies')extra='<label class="checkrow"><input type="checkbox" id="v16CookieShow" '+(state.show?'checked':'')+'> Afficher le bandeau de cookies</label><div class="field"><label>Position</label><select id="v16CookiePos"><option value="bottom">Bas de page</option><option value="top">Haut de page</option></select></div>';
 const imageField=['about','apropos'].includes(page.id)?'<div class="field wide"><label>Image de la page (URL HTTPS)</label><input type="url" id="v16PageImage" placeholder="https://…"></div>':'';
 const visibility='<div class="wide v16-visibility"><label class="checkrow"><input type="checkbox" id="v16ShowLead" '+(state.showLead!==false?'checked':'')+'> Afficher l’introduction</label><label class="checkrow"><input type="checkbox" id="v16ShowBody" '+(state.showBody!==false?'checked':'')+'> Afficher le contenu</label></div>';
 $('#v16PageEditor').innerHTML='<div class="v16-page-form">'+field(page.title,'Titre',page.titleText)+field(page.lead,'Introduction',page.leadText,true)+'<div class="field wide"><label>Contenu</label><textarea data-k="'+page.body+'" aria-label="Contenu">'+esc(T(page.body)||page.bodyText)+'</textarea></div>'+imageField+((page.buttons||[]).map((key,i)=>field(key,'Bouton · '+page.buttonTexts[i],page.buttonTexts[i])).join(''))+'<div class="field"><label>Fond de la page</label><input type="color" id="v16PageBg" value="'+state.bg+'"></div>'+visibility+extra+'</div>';
 $('#v16PageEditor').querySelectorAll('[data-k]').forEach(i=>i.addEventListener('input',e=>setT(e.currentTarget.dataset.k,e.currentTarget.value)));
 const blockText=T(page.body)||page.bodyText,blockParts=blockText.split(/\n\s*\n/),blockList=document.createElement('div');blockList.className='field wide v16-content-blocks';blockList.innerHTML='<label>Blocs de contenu — visibilité et ordre</label><p class="hint">Chaque paragraphe séparé par une ligne vide peut être masqué ou déplacé.</p>'+blockParts.map((txt,i)=>'<div class="v16-block-row" data-block="'+i+'"><label class="checkrow"><input type="checkbox" data-block-visible="'+i+'" '+(state.blocksVisible?.[i]!==false?'checked':'')+'> <span>Bloc '+(i+1)+' · '+esc(txt.slice(0,72))+(txt.length>72?'…':'')+'</span></label><button type="button" data-move-block="-1" data-index="'+i+'" '+(i===0?'disabled':'')+' aria-label="Monter le bloc '+(i+1)+'">↑</button><button type="button" data-move-block="1" data-index="'+i+'" '+(i===blockParts.length-1?'disabled':'')+' aria-label="Descendre le bloc '+(i+1)+'">↓</button></div>').join('');
 $('#v16PageEditor').querySelector('.v16-page-form').appendChild(blockList);
 blockList.querySelectorAll('[data-block-visible]').forEach(cb=>cb.onchange=()=>{state.blocksVisible=state.blocksVisible||[];state.blocksVisible[+cb.dataset.blockVisible]=cb.checked;renderPreview()});
 blockList.querySelectorAll('[data-move-block]').forEach(btn=>btn.onclick=()=>{const from=+btn.dataset.index,to=from+(+btn.dataset.moveBlock);if(to<0||to>=blockParts.length)return;[blockParts[from],blockParts[to]]=[blockParts[to],blockParts[from]];setT(page.body,blockParts.join('\n\n'));renderPageEditorV16()});
 const contentArea=$('#v16PageEditor').querySelector('textarea[data-k="'+page.body+'"]');contentArea.addEventListener('change',renderPageEditorV16);
 $('#v16PageEditor').querySelector('#v16PageBg').addEventListener('input',e=>{state.bg=e.target.value;renderPreview()});
 const showLead=$('#v16ShowLead'),showBody=$('#v16ShowBody');if(showLead)showLead.onchange=()=>{state.showLead=showLead.checked;renderPreview()};if(showBody)showBody.onchange=()=>{state.showBody=showBody.checked;renderPreview()};
 const pageImage=$('#v16PageImage');if(pageImage){pageImage.value=state.image||'';pageImage.oninput=()=>{state.image=pageImage.value;renderPreview()}}
 const cf=$('#v16ContactForm');if(cf)cf.onchange=()=>{state.showForm=cf.checked;renderPreview()};
 const ah=$('#v16AboutHero');if(ah)ah.onchange=()=>{state.showHero=ah.checked;renderPreview()};
 const ck=$('#v16CookieShow');if(ck)ck.onchange=()=>{state.show=ck.checked;renderPreview()};
 const cp=$('#v16CookiePos');if(cp){cp.value=state.position||'bottom';cp.onchange=()=>{state.position=cp.value;renderPreview()}}
}

function renderInfoPageV16(){
 const p=PAGE_INFO_V16.find(x=>x.id===pagesInfoV16Active)||PAGE_INFO_V16[0],st=studioV16.pages[p.id],body=(T(p.body)||p.bodyText).split(/\n\s*\n/).map((x,i)=>st.blocksVisible?.[i]===false?'':'<p>'+esc(x).replace(/\n/g,'<br>')+'</p>').join('');
 const buttons=(p.buttons||[]).map(k=>'<button style="background:'+front.accent+';color:#fff;border:0;border-radius:8px;padding:8px 12px;margin-right:6px">'+esc(T(k))+'</button>').join('');
 const lead=(st.showLead!==false&&(st.showHero!==false))?'<div style="background:'+front.accent+'18;padding:15px 18px;border-radius:12px;margin-bottom:16px"><strong>'+esc(T(p.lead))+'</strong></div>':'';
 const pageImg=st.image&&validPreviewUrlV16(st.image)?'<img src="'+esc(validPreviewUrlV16(st.image))+'" alt="Illustration '+esc(p.name)+'" style="max-width:100%;max-height:220px;object-fit:cover;border-radius:12px;margin:0 0 14px">':'';
 const pageBody=st.showBody!==false?body:'';
 let content='';if(p.form)content='<div style="display:grid;grid-template-columns:1fr 1.2fr;gap:14px"><div style="padding:14px;background:#f7f8f7;border-radius:10px">'+pageImg+pageBody+'</div>'+(st.showForm?'<div style="padding:14px;border:1px solid #e5e7eb;border-radius:10px"><h3>'+esc(T(p.form.title))+'</h3><p>'+esc(T(p.form.notice))+'</p><input placeholder="'+esc(T(p.form.name))+'" style="width:100%;padding:8px;margin:4px 0"><input placeholder="'+esc(T(p.form.email))+'" style="width:100%;padding:8px;margin:4px 0"><input placeholder="'+esc(T(p.form.phone))+'" style="width:100%;padding:8px;margin:4px 0"><input placeholder="'+esc(T(p.form.subject))+'" style="width:100%;padding:8px;margin:4px 0"><textarea placeholder="'+esc(T(p.form.message))+'" style="width:100%;padding:8px;margin:4px 0"></textarea><button style="background:'+front.accent+';color:#fff;padding:8px 12px;border:0;border-radius:7px">'+esc(T(p.form.submit))+'</button></div>':'')+'</div>';
 else if(p.id==='cookies'){const c=studioV16.pages.cookies;content='<div style="border-radius:10px;padding:16px;background:'+c.bg+';color:#fff"><h3>'+esc(T(p.title))+'</h3>'+(st.showLead!==false?'<p>'+esc(T(p.lead))+'</p>':'')+(st.showBody!==false?body:'')+buttons+'</div>'}
 else content='<div style="padding:16px 18px;background:#fff;border:1px solid #e5e7eb;border-radius:12px">'+pageImg+pageBody+buttons+'</div>';
 const title=esc(T(p.title));
 PV.html='<div style="background:'+st.bg+';color:'+front.text+';font-family:'+front.font+',system-ui,sans-serif">'+pvHeaderHTML()+'<div class="pv-body"><div style="font-size:11px;color:#778078;margin-bottom:8px">Accueil › '+esc(p.name)+'</div><h2 style="color:'+front.accent+'">'+title+'</h2>'+lead+content+'</div>'+pvFooterHTML()+'</div>';
}

function bindV16Controls(){
 const s=studioV16.style,d=studioV16.dark,i=studioV16.identity;
 [['v16Radius','radius'],['v16Shadow','shadow'],['v16Density','density'],['v16Buttons','buttons'],['v16TitleFont','titleFont']].forEach(([id,k])=>{const e=$('#'+id);e.value=s[k];e.onchange=()=>{s[k]=e.value;renderPreview()}});
 [['v16Logo','logo'],['v16Favicon','favicon'],['v16Share','share'],['v16ShareTitle','shareTitle'],['v16ShareDesc','shareDesc']].forEach(([id,k])=>{const e=$('#'+id);e.value=i[k];e.oninput=()=>{i[k]=e.value;renderShareCardV16();renderPreview()}});
 [['v16DarkBg','bg'],['v16DarkText','text'],['v16DarkCard','card'],['v16DarkBorder','border']].forEach(([id,k])=>{const e=$('#'+id);e.value=d[k];e.oninput=()=>{d[k]=e.value;renderPreview()}});
 $('#v16DarkEnabled').checked=d.enabled;$('#v16DarkEnabled').onchange=e=>{d.enabled=e.target.checked;renderPreview()};
 $('#v16DarkToggle').checked=d.showToggle;$('#v16DarkToggle').onchange=e=>{d.showToggle=e.target.checked;renderPreview()};
 $('#v16DarkPreview').checked=d.preview===true;$('#v16DarkPreview').onchange=e=>{d.preview=e.target.checked;renderPreview()};
 $('#v16DarkDefault').value=d.defaultMode;$('#v16DarkDefault').onchange=e=>{d.defaultMode=e.target.value;renderPreview()};
 $('#v16Animations').value=studioV16.animations;$('#v16Animations').onchange=e=>{studioV16.animations=e.target.value;renderPreview()};
 $('#v16PageSelect').value=pagesInfoV16Active;$('#v16PageSelect').onchange=e=>{pagesInfoV16Active=e.target.value;renderPageEditorV16();renderPreview()};
 renderShareCardV16();renderPageEditorV16();
}

function renderShareCardV16(){const i=studioV16.identity;$('#v16ShareCard').innerHTML='<img src="'+esc(validPreviewUrlV16(i.share))+'" alt="Aperçu de partage"><div><strong>'+esc(i.shareTitle||reglages.nom)+'</strong><small>'+esc(i.shareDesc||'Description de partage')+'</small><div class="hint">Carte de partage simulée — les balises SEO réelles ne sont pas modifiées.</div></div>';}

function openChangeListV16(mods){const settings=$('#studioNav button[data-tab="reglages"]');if(settings)settings.click();const box=$('#v16Counters');if(!box)return;box.querySelectorAll('.v16-change-module').forEach(d=>{const name=d.querySelector('summary')?.textContent||'';d.open=!mods||mods.some(m=>name.startsWith(({accueil:'Accueil',boutique:'Boutique',fiche:'Fiche',panier:'Panier',compte:'Compte',pagesInfo:'Pages d’information',reglages:'Réglages'}[m]||m)))});box.scrollIntoView({behavior:'smooth',block:'center'})}

function renderCountersV16(){
 if(!pubSnap)return;let base;try{base=JSON.parse(pubSnap)}catch(_){return}
 const textValues=(mod,source)=>Object.fromEntries((MODULE_TEXTS[mod]||[]).map(x=>[x.key,(source||Tx)[x.key]]));
 const current={front:{state:{accent:front.accent,priceColor:front.priceColor,bg:front.bg,text:front.text,font:front.font,theme:front.theme},v16:studioV16,texts:textValues('front')},header:{state:header,texts:textValues('header')},footer:{state:footer,texts:textValues('footer')},accueil:{state:{rows}},boutique:{state:shop},fiche:{state:fiche},panier:{state:panier},menu:{state:menu,texts:textValues('menu')},compte:{state:acc,texts:textValues('compte')},seo:{state:seo},reglages:{state:reglages},pagesInfo:{texts:textValues('pagesInfo'),pages:studioV16.pages}};
 const baseTx=base.Tx||{},baseline={front:{state:{accent:base.front?.accent,priceColor:base.front?.priceColor,bg:base.front?.bg,text:base.front?.text,font:base.front?.font,theme:base.front?.theme},v16:base.front?._studioV16,texts:textValues('front',baseTx)},header:{state:base.header,texts:textValues('header',baseTx)},footer:{state:base.footer,texts:textValues('footer',baseTx)},accueil:{state:{rows:base.rows}},boutique:{state:base.shop},fiche:{state:base.fiche},panier:{state:base.panier},menu:{state:base.menu,texts:textValues('menu',baseTx)},compte:{state:base.acc,texts:textValues('compte',baseTx)},seo:{state:base.seo},reglages:{state:base.reglages},pagesInfo:{texts:textValues('pagesInfo',baseTx),pages:(base.front?._studioV16||{}).pages}};
 const changes=Object.fromEntries(Object.keys(current).map(k=>[k,v16DiffLeaves(current[k],baseline[k])]).filter(([,v])=>v.length));
 const changed=Object.keys(changes);$('#v16ChangeCount').textContent=changed.length+' module'+(changed.length===1?'':'s')+' modifié'+(changed.length===1?'':'s');
 const names={accueil:'Accueil',boutique:'Boutique',fiche:'Fiche',panier:'Panier',compte:'Compte',pagesInfo:'Pages d’information',reglages:'Réglages'};
 const labels={};Object.values(MODULE_TEXTS).flat().forEach(x=>labels[x.key]=x.label);
 const host=$('#v16Counters');if(host)host.innerHTML=changed.length?changed.map(mod=>'<details class="v16-change-module"><summary>'+esc(names[mod]||mod)+' · '+changes[mod].length+' changement(s)</summary>'+changes[mod].map(x=>{const textKey=x.path.startsWith('texts.')?x.path.slice(6):'';const label=textKey?(labels[textKey]||textKey):x.path.replaceAll('.',' · ');return '<div class="v16-change-row"><span>'+esc(label)+'</span><button type="button" data-reset-module="'+mod+'" data-reset-path="'+x.path+'" title="Restaurer ce champ">↺</button></div>'}).join('')+'</details>').join(''):'<span class="hint">Aucun changement depuis la référence locale.</span>';
 host?.querySelectorAll('[data-reset-module]').forEach(btn=>btn.onclick=()=>{resetModuleFieldV16(btn.dataset.resetModule,btn.dataset.resetPath,baseline);syncUI();bindV16Controls();renderPageEditorV16();renderAllModuleTexts();renderTxtList();renderMenuList();renderPreview();toast('Champ restauré à sa valeur de référence.')});
 document.querySelectorAll('#studioNav button[data-tab]').forEach(b=>{b.querySelector('.v16-badge')?.remove();const key=b.dataset.tab;if(changed.includes(key)){const badge=document.createElement('span');badge.className='v16-badge';badge.textContent='•';badge.title=changes[key].length+' changement(s) non exporté(s)';badge.onclick=e=>{e.preventDefault();e.stopPropagation();openChangeListV16([key])};b.appendChild(badge)}});
 CATEGORIES.forEach(cat=>{const catHost=document.querySelector('.cat[data-cat="'+cat.id+'"] .cat-btn');if(!catHost)return;catHost.querySelector('.v16-cat-count')?.remove();const touched=cat.mods.filter(m=>changed.includes(m)),count=touched.length;if(count){const badge=document.createElement('span');badge.className='v16-cat-count';badge.textContent=String(count);badge.title=count+' module(s) modifié(s) — ouvrir les changements';badge.onclick=e=>{e.preventDefault();e.stopPropagation();openChangeListV16(touched)};catHost.appendChild(badge)}});
}

const localNoteV16=document.createElement('div');

const previewNote16=$('.pv-note');

const tabletBtnV16=document.createElement('button');

