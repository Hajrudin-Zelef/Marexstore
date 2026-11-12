/* Marexstore — miroir serveur (ne change rien au comportement local).
   Chaque action locale (panier, commande WhatsApp, compte, devis, suivi)
   est aussi postée au serveur, qui tient le registre marchand.
   app.js n'est pas modifié : ce fichier se contente d'écouter en capture. */
(function () {
  'use strict';
  var csrfCache = null;
  function csrf() {
    if (csrfCache) return csrfCache;
    var m = document.querySelector('meta[name="csrf-token"]');
    csrfCache = m ? m.getAttribute('content') : '';
    return csrfCache;
  }
  function refreshCsrf() {
    return fetch('/api/csrf', { credentials: 'same-origin' })
      .then(function (r) { return r.json(); })
      .then(function (j) { csrfCache = j.csrf || csrfCache; })
      .catch(function () {});
  }
  function post(url, data, retried) {
    try {
      var fd = new FormData();
      fd.append('csrf', csrf());
      for (var k in data) fd.append(k, data[k]);
      fetch(url, { method: 'POST', body: fd, credentials: 'same-origin', keepalive: true, headers: { 'X-Mirror': '1' } })
        .then(function (r) {
          if (r.status === 403 && !retried) refreshCsrf().then(function () { post(url, data, true); });
        })
        .catch(function () {});
    } catch (e) {}
  }
  function cart() {
    try { return JSON.parse(window.localStorage.getItem('mediatec-panier-v1') || '[]'); }
    catch (e) { return []; }
  }
  document.addEventListener('click', function (e) {
    var add = e.target.closest ? e.target.closest('[data-add-cart]') : null;
    if (add) {
      var q = 1;
      var qs = add.getAttribute('data-qty');
      if (qs) { var inp = document.querySelector(qs); if (inp) q = Math.max(1, parseInt(inp.value, 10) || 1); }
      post('/api/panier', { slug: add.getAttribute('data-slug') || '', qty: String(q) });
      return;
    }
    var ord = e.target.closest ? e.target.closest('[data-cart-order]') : null;
    if (ord && ord.getAttribute('href')) {
      var items = cart().map(function (it) { return it.slug + 'x' + it.qty; }).join(',');
      post('/api/commande', { items: items, moyen: 'whatsapp' });
      return;
    }
    var lo = e.target.closest ? e.target.closest('[data-logout]') : null;
    if (lo) post('/auth/logout', {});
  }, true);
  function hookForm(sel, url, extra) {
    var f = document.querySelector(sel);
    if (!f || f.__mirrored) return;
    f.__mirrored = true;
    f.addEventListener('submit', function () {
      try {
        var fd = new FormData(f);
        var data = {};
        fd.forEach(function (v, k) { if (k !== 'csrf' && k !== 'password2') data[k] = v; });
        if (extra) for (var k in extra) data[k] = extra[k];
        post(url, data);
      } catch (e) {}
    });
  }
  function hookAll() {
    hookForm('[data-create-form]', '/auth/inscription');
    hookForm('[data-login-form]', '/auth/connexion');
    hookForm('[data-devis-form]', '/api/devis');
    hookForm('[data-suivi-form]', '/api/suivi');
    hookForm('[data-infos-form]', '/api/infos');
    hookForm('[data-contact-form]', '/api/contact');
  }
  if (document.readyState !== 'loading') hookAll();
  else document.addEventListener('DOMContentLoaded', hookAll);
})();
