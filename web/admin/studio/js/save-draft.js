/* Sauvegarde explicite du Studio (bouton disquette + Ctrl+S).

   Modèle : tant qu'on ne sauvegarde pas, les modifications restent « en
   attente » (compteur, pastille rouge, résumé tronqué au clic). Sauvegarder
   écrit un BROUILLON côté serveur (design_files, nom distinct) : le site
   public ne change pas — seule la publication l'engage. */
(function () {
  'use strict';

  function snapActuel() {
    return (typeof snap === 'function') ? snap() : '';
  }

  function elements() {
    return {
      btn: document.getElementById('btnSave'),
      badge: document.getElementById('pendingCount'),
      pop: document.getElementById('pendingPop')
    };
  }

  /* --- Compteur + pastille + résumé ------------------------------------ */
  function resume() {
    // Reprend le calcul déjà fait par renderCountersV16 (aucune duplication).
    var out = { total: 0, parModule: [] };
    try {
      var base = JSON.parse(savedSnap || pubSnap || '{}');
      var cur = { front: front, rows: rows, shop: shop, fiche: fiche, panier: panier, Tx: Tx,
        txtPages: txtPages, menu: menu, seo: seo, reglages: reglages, header: header,
        footer: footer, acc: acc };
      var noms = { front: 'Apparence', rows: 'Page d’accueil', shop: 'Boutique', fiche: 'Fiche',
        panier: 'Panier', Tx: 'Textes', txtPages: 'Pages', menu: 'Menu', seo: 'SEO',
        reglages: 'Réglages', header: 'En-tête', footer: 'Pied de page', acc: 'Compte' };
      Object.keys(cur).forEach(function (k) {
        if (JSON.stringify(cur[k]) !== JSON.stringify(base[k])) {
          var d = (typeof v16DiffLeaves === 'function') ? v16DiffLeaves(cur[k], base[k]).length : 1;
          if (d > 0) { out.parModule.push({ nom: noms[k] || k, n: d }); out.total += d; }
        }
      });
    } catch (e) { /* base illisible : compteur figé à 0 */ }
    return out;
  }

  function rafraichirCompteur() {
    var el = elements();
    var r = resume();

    if (el.badge) {
      el.badge.textContent = r.total > 0 ? String(r.total) : '';
      el.badge.hidden = r.total === 0;
      el.badge.title = r.total > 0 ? r.total + ' modification(s) en attente de sauvegarde' : '';
    }
    if (el.btn) {
      el.btn.disabled = r.total === 0;
      el.btn.classList.toggle('has-pending', r.total > 0);
      el.btn.title = r.total > 0
        ? 'Sauvegarder ' + r.total + ' modification(s) en attente (Ctrl+S)'
        : 'Aucune modification en attente (Ctrl+S)';
    }
    if (el.pop && !el.pop.hidden) el.pop.innerHTML = resumeHTML(r);
  }

  function resumeHTML(r) {
    if (r.total === 0) {
      return '<p class="pend-empty">Aucune modification en attente.</p>';
    }
    var lignes = r.parModule.slice(0, 5).map(function (m) {
      return '<li><span>' + m.nom + '</span><b>' + m.n + '</b></li>';
    }).join('');
    var reste = r.parModule.length > 5 ? '<li class="pend-more">+ ' + (r.parModule.length - 5) + ' autre(s) section(s)</li>' : '';
    return '<p class="pend-head">' + r.total + ' modification' + (r.total > 1 ? 's' : '') + ' en attente</p>' +
      '<ul class="pend-list">' + lignes + reste + '</ul>' +
      '<p class="pend-foot">Rien n’est en ligne tant que vous ne cliquez pas <b>Publier</b>.</p>';
  }

  /* --- Sauvegarde ------------------------------------------------------ */
  function sauvegarder() {
    var el = elements();
    var r = resume();
    if (r.total === 0) return;
    if (el.btn) el.btn.disabled = true;
    var corps = (typeof designJSON === 'function') ? JSON.stringify(designJSON()) : snapActuel();
    fetch('/admin/api/design/draft', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-CSRF': window.__STUDIO_CSRF__ || '' },
      credentials: 'same-origin',
      body: corps
    }).then(function (res) {
      if (!res.ok) {
        return res.text().then(function (t) { throw new Error(t.slice(0, 160)); });
      }
      return res.json();
    }).then(function (j) {
      savedSnap = snapActuel();
      draftMeta = { version: j.version, heure: new Date() };
      if (typeof updateStatus === 'function') updateStatus();
      rafraichirCompteur();
      var h = new Date();
      var hh = ('0' + h.getHours()).slice(-2) + ':' + ('0' + h.getMinutes()).slice(-2);
      if (typeof toast === 'function') toast('💾 Brouillon sauvegardé à ' + hh + ' (v' + j.version + '). Le site public n’a pas changé.');
    }).catch(function (e) {
      if (typeof toast === 'function') toast('❌ Sauvegarde impossible : ' + e.message);
    }).then(function () {
      rafraichirCompteur();
    });
  }

  /* --- Reprise d'un brouillon existant --------------------------------- */
  function proposerReprise() {
    fetch('/admin/api/design/draft', { credentials: 'same-origin' })
      .then(function (r) { return r.ok ? r.json() : null; })
      .then(function (j) {
        if (!j || !j.design) return;
        var bar = document.createElement('div');
        bar.className = 'draft-banner';
        bar.innerHTML = '<span>Un brouillon sauvegardé existe (' + (j.cree_le || '—') + ', v' + j.version + ').</span>' +
          '<button type="button" class="btn btn-ghost" id="draftResume">Reprendre</button>' +
          '<button type="button" class="btn btn-ghost" id="draftIgnore">Ignorer</button>';
        var host = document.querySelector('.studio-top');
        if (!host) return;
        host.insertAdjacentElement('afterend', bar);
        bar.querySelector('#draftIgnore').addEventListener('click', function () { bar.remove(); });
        bar.querySelector('#draftResume').addEventListener('click', function () {
          try {
            var o = typeof j.design === 'string' ? JSON.parse(j.design) : j.design;
            if (typeof pushHist === 'function') pushHist();
            if (typeof restore === 'function') restore(JSON.stringify(o));
            savedSnap = snapActuel();
            if (typeof updateStatus === 'function') updateStatus();
            rafraichirCompteur();
            if (typeof toast === 'function') toast('Brouillon repris (non publié).');
          } catch (e) {
            if (typeof toast === 'function') toast('❌ Brouillon illisible.');
          }
          bar.remove();
        });
      })
      .catch(function () {});
  }

  /* --- Câblage --------------------------------------------------------- */
  function init() {
    var el = elements();
    if (el.btn) el.btn.addEventListener('click', sauvegarder);
    if (el.badge) {
      el.badge.addEventListener('click', function () {
        var pop = elements().pop;
        if (!pop) return;
        pop.hidden = !pop.hidden;
        if (!pop.hidden) pop.innerHTML = resumeHTML(resume());
      });
    }
    document.addEventListener('mousedown', function (e) {
      var pop = document.getElementById('pendingPop');
      if (!pop || pop.hidden) return;
      if (pop.contains(e.target) || (el.badge && el.badge.contains(e.target))) return;
      pop.hidden = true;
    });
    document.addEventListener('keydown', function (e) {
      if ((e.ctrlKey || e.metaKey) && (e.key === 's' || e.key === 'S')) {
        e.preventDefault();
        sauvegarder();
      }
    });
    // Filet anti-perte : prévenir si on quitte avec des modifications non
    // sauvegardées (le navigateur affiche son propre message).
    window.addEventListener('beforeunload', function (e) {
      var r = resume();
      if (r.total > 0) { e.preventDefault(); e.returnValue = ''; return ''; }
    });
    // Mise à jour du compteur à chaque changement d'état du Studio.
    var ancien = window.updateStatus;
    if (typeof ancien === 'function' && !ancien.__avecCompteur) {
      window.updateStatus = function () {
        var r = ancien.apply(this, arguments);
        rafraichirCompteur();
        return r;
      };
      window.updateStatus.__avecCompteur = true;
    }
    setInterval(rafraichirCompteur, 1200);
    rafraichirCompteur();
    proposerReprise();
  }

  // Point d'entrée utilisé par le bouton (onclick) et par Ctrl+S.
  window.sauvegarderStudio = sauvegarder;

  if (document.readyState === 'complete') init();
  else window.addEventListener('load', init);
})();
