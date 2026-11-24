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
        setCollapsed(s, norm(t.textContent).indexOf('style fin') === 0);
      })(secs[i]);
    }
  }

  /* 5. Recherche de réglage */
  function applySearch() {
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

  function init() {
    buildTop();
    makeCollapsible();
    refresh();
    ['input', 'change', 'click', 'keyup'].forEach(function (ev) {
      document.addEventListener(ev, schedule, true);
    });
    new MutationObserver(function () {
      buildTop();
      makeCollapsible();
      schedule();
    }).observe(panel, { childList: true, subtree: true });
  }
  if (document.readyState === 'complete') init();
  else window.addEventListener('load', init);
})();
