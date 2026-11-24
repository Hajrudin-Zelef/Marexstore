/* premium.js — Studio Marexstore (v17).
   Ajouts isolés, chargés en dernier : sections repliables, recherche de réglage,
   valeur hex des couleurs, contrôle de lisibilité. N'altère aucune fonction existante. */
(function () {
  'use strict';
  var panel = document.getElementById('editorPanel');
  if (!panel) return;
  var timer = null;

  function norm(s) {
    s = String(s || '').toLowerCase();
    return s.normalize ? s.normalize('NFD').replace(/[\u0300-\u036f]/g, '') : s;
  }
  function setText(el, txt) { if (el.textContent !== txt) el.textContent = txt; }
  function setClass(el, cls) { if (el.className !== cls) el.className = cls; }

  /* 1. Valeur hex à côté de chaque sélecteur de couleur */
  function decorateColors() {
    var inputs = panel.querySelectorAll('input[type="color"]');
    for (var i = 0; i < inputs.length; i++) {
      var inp = inputs[i], tag = inp.nextElementSibling;
      if (!tag || !tag.classList.contains('px-hex')) {
        tag = document.createElement('code');
        tag.className = 'px-hex';
        inp.insertAdjacentElement('afterend', tag);
      }
      setText(tag, String(inp.value || '').toUpperCase());
    }
  }

  /* 2. Contraste du bandeau d'annonce (texte / fond) */
  function annContrast() {
    var fg = document.getElementById('v13AnnFg'), bg = document.getElementById('v13AnnBg');
    if (!fg || !bg || typeof contrastRatioV13 !== 'function') return;
    var el = document.getElementById('pxAnnContrast');
    if (!el) {
      el = document.createElement('div');
      el.id = 'pxAnnContrast';
      var grid = bg.closest('.grid2');
      (grid || bg.parentNode).insertAdjacentElement('afterend', el);
    }
    var r = contrastRatioV13(fg.value, bg.value);
    if (r === null) { setText(el, ''); return; }
    setClass(el, 'px-contrast ' + (r >= 4.5 ? 'ok' : 'warn'));
    setText(el, r >= 4.5
      ? 'Contraste ' + r.toFixed(1) + ':1 — texte lisible'
      : 'Contraste ' + r.toFixed(1) + ':1 — texte difficile à lire (minimum conseillé 4.5:1)');
  }

  /* 3. Encart de lisibilité global (réutilise le contrôle existant) */
  function a11yBox() {
    var box = document.getElementById('pxA11y');
    if (!box || typeof contrastIssuesV13 !== 'function') return;
    var issues;
    try { issues = contrastIssuesV13(); } catch (e) { return; }
    if (!issues.length) {
      setClass(box, 'px-a11y px-keep ok');
      setText(box, 'Lisibilité : tous les contrastes vérifiés sont conformes.');
    } else {
      setClass(box, 'px-a11y px-keep warn');
      setText(box, 'Lisibilité : contraste faible sur ' + issues.map(function (x) {
        return x.name + ' (' + x.ratio.toFixed(1) + ':1)';
      }).join(', ') + ' — minimum conseillé 4.5:1.');
    }
  }

  /* 4. Sections repliables */
  function setCollapsed(sec, on) {
    sec.classList.toggle('px-collapsed', on);
    var t = sec.firstElementChild;
    if (t) t.setAttribute('aria-expanded', on ? 'false' : 'true');
  }

  /* 5. Scission du panneau « Style fin et identité » : identité (logo,
     favicon, partage, mode sombre) séparée des formes (arrondis, ombres…).
     Déplace des nœuds existants — les liaisons par id de bindV16Controls
     survivent. Idempotent via data-px-split. */
  function splitStylePanel() {
    var panels = panel.querySelectorAll('.v16-feature-panel');
    var src = null;
    for (var i = 0; i < panels.length; i++) {
      var l = panels[i].querySelector('label');
      if (l && norm(l.textContent).indexOf('style fin') === 0) { src = panels[i]; break; }
    }
    if (!src || src.getAttribute('data-px-split')) return;
    src.setAttribute('data-px-split', '1');
    var kids = Array.prototype.slice.call(src.children);
    var formes = [], identite = [], bucket = 'formes';
    kids.forEach(function (c) {
      var isLabel = c.tagName === 'LABEL' && !c.querySelector('input,select,textarea');
      if (isLabel) {
        var t = norm(c.textContent);
        if (t.indexOf('identite') === 0 || t.indexOf('mode sombre') === 0) bucket = 'identite';
        else if (t.indexOf('animations') === 0 || t.indexOf('formes') === 0 || t.indexOf('style fin') === 0) bucket = 'formes';
      }
      (bucket === 'identite' ? identite : formes).push(c);
    });
    var titreFormes = formes.shift();
    if (titreFormes && titreFormes.tagName === 'LABEL') {
      titreFormes.textContent = 'Formes et finitions';
      formes.unshift(titreFormes);
    } else if (titreFormes) {
      formes.unshift(titreFormes);
    }

    var idPanel = document.createElement('section');
    idPanel.className = 'field v16-feature-panel';
    identite.forEach(function (c) { idPanel.appendChild(c); });

    src.innerHTML = '';
    formes.forEach(function (c) { src.appendChild(c); });
    if (identite.length) src.parentNode.insertBefore(idPanel, src.nextSibling);
  }

  /* 6. Ordre logique : Thèmes → Identité → Couleurs → Typographie → Formes
     → Bandeau → Badges → Textes. Idempotent : on ne touche au DOM que si
     l'ordre réel diffère (premium.js observe le panneau, un déplacement
     inutile relancerait l'observateur en boucle). */
  function sectionWeight(el) {
    var t = norm(el.textContent || '').trim();
    var l = el.querySelector('label'), fl = l ? norm(l.textContent).trim() : '';
    if (el.classList.contains('px-keep')) return 5;
    if (el.classList.contains('v13-gallery')) return 10;
    if (el.classList.contains('v16-feature-panel') && fl.indexOf('identite') === 0) return 20;
    if (t.indexOf('couleur principale') === 0) return 30;
    if (t.indexOf('fond du site') === 0) return 31;
    if (t.indexOf('police') === 0) return 40;
    if (t.indexOf('forme des boutons') === 0) return 50;
    if (el.classList.contains('v16-feature-panel')) return 55;
    if (el.classList.contains('v13-ann')) return 60;
    if (t.indexOf('afficher le badge') === 0) return 70;
    if (t.indexOf('textes de cette page') === 0) return 80;
    return 90;
  }

  function orderSectionsPx() {
    var host = document.getElementById('front-ui');
    if (!host) return false;
    var nodes = Array.prototype.slice.call(host.children);
    var want = nodes.map(function (el, i) { return { el: el, w: sectionWeight(el), i: i }; })
      .sort(function (a, b) { return a.w - b.w || a.i - b.i; });
    var same = true;
    for (var i = 0; i < want.length; i++) { if (nodes[i] !== want[i].el) { same = false; break; } }
    if (same) return false;
    want.forEach(function (x) { host.appendChild(x.el); });
    return true;
  }
  window.orderSectionsPx = orderSectionsPx;

  /* 7. Sections repliables (après la scission, pour couvrir les 2 panneaux) */
  function makeCollapsible() {
    var host = document.getElementById('front-ui');
    if (!host) return;
    var secs = host.querySelectorAll('section.field, .v16-feature-panel');
    for (var i = 0; i < secs.length; i++) {
      (function (s) {
        var t = s.firstElementChild;
        if (s.getAttribute('data-px') || !t || t.tagName !== 'LABEL') return;
        s.setAttribute('data-px', '1');
        s.classList.add('px-collapsible');
        t.setAttribute('role', 'button');
        t.setAttribute('tabindex', '0');
        var toggle = function (e) {
          e.preventDefault();
          setCollapsed(s, !s.classList.contains('px-collapsed'));
        };
        t.addEventListener('click', toggle);
        t.addEventListener('keydown', function (e) {
          if (e.key === 'Enter' || e.key === ' ') toggle(e);
        });
        var st = norm(t.textContent);
        setCollapsed(s, st.indexOf('style fin') === 0 || st.indexOf('formes') === 0);
      })(secs[i]);
    }
  }

  /* 5. Recherche de réglage */  function applySearch() {
    var host = document.getElementById('front-ui'), inp = document.getElementById('pxSearch');
    if (!host || !inp) return;
    var q = norm(inp.value).trim(), shown = 0, kids = host.children;
    for (var i = 0; i < kids.length; i++) {
      var k = kids[i];
      if (k.classList.contains('px-keep')) continue;
      var hit = !q || norm(k.textContent).indexOf(q) > -1;
      k.classList.toggle('px-hide', !hit);
      if (hit) shown++;
      if (q && hit && k.classList.contains('px-collapsed')) setCollapsed(k, false);
    }
    document.getElementById('pxNone').classList.toggle('px-hide', !(q && shown === 0));
  }
  function buildTop() {
    var host = document.getElementById('front-ui');
    if (!host || document.getElementById('pxSearch')) return;
    var wrap = document.createElement('div');
    wrap.className = 'px-keep px-searchwrap';
    var input = document.createElement('input');
    input.type = 'search';
    input.id = 'pxSearch';
    input.className = 'px-search';
    input.placeholder = 'Rechercher un réglage (couleur, bandeau, arrondi…)';
    input.setAttribute('aria-label', 'Rechercher un réglage');
    wrap.appendChild(input);
    var box = document.createElement('div');
    box.id = 'pxA11y';
    box.className = 'px-a11y px-keep';
    var none = document.createElement('p');
    none.id = 'pxNone';
    none.className = 'hint px-keep px-hide';
    none.textContent = 'Aucun réglage ne correspond à cette recherche.';
    host.insertBefore(none, host.firstChild);
    host.insertBefore(box, host.firstChild);
    host.insertBefore(wrap, host.firstChild);
    input.addEventListener('input', applySearch);
  }

  function refresh() { decorateColors(); annContrast(); a11yBox(); }
  function schedule() { clearTimeout(timer); timer = setTimeout(refresh, 80); }

  function organize() {
    splitStylePanel();
    orderSectionsPx();
    makeCollapsible();
  }

  function init() {
    buildTop();
    organize();
    refresh();
    ['input', 'change', 'click', 'keyup'].forEach(function (ev) {
      document.addEventListener(ev, schedule, true);
    });
    new MutationObserver(function () {
      buildTop();
      organize();
      schedule();
    }).observe(panel, { childList: true, subtree: true });
  }
  if (document.readyState === 'complete') init();
  else window.addEventListener('load', init);
})();
