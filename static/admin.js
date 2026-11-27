// Aperçu live de la fiche produit : la carte se met à jour pendant la saisie.
(function () {
  var form = document.getElementById('mx-edit-form');
  if (!form) return;
  function money(n) {
    return String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ' ');
  }
  function badgeCls(d) {
    if (d === 'en stock') return 'ok';
    if (d === 'rupture de stock') return 'bad';
    return 'warn';
  }
  function val(name) {
    var el = form.querySelector('[name="' + name + '"]');
    return el ? el.value.trim() : '';
  }
  function set(id, txt) {
    var e = document.getElementById(id);
    if (e) e.textContent = txt;
  }
  function upd() {
    set('pv-nom', val('nom') || 'Nouveau produit');
    var marque = val('marque');
    var pvMarque = document.getElementById('pv-marque');
    if (pvMarque) {
      pvMarque.textContent = marque;
      pvMarque.style.display = marque !== '' ? '' : 'none';
    }
    var pb = val('prix_boutique');
    var prix = val('prix');
    var pvPrix = document.getElementById('pv-prix');
    if (pvPrix) {
      var p = (pb !== '' ? pb : prix).replace(/\D/g, '');
      if (p !== '') {
        pvPrix.textContent = money(p) + ' F CFA';
        pvPrix.classList.remove('devis');
      } else {
        pvPrix.textContent = 'devis';
        pvPrix.classList.add('devis');
      }
    }
    var dispo = val('dispo') || 'en stock';
    var pvDispo = document.getElementById('pv-dispo');
    if (pvDispo) pvDispo.innerHTML = '<span class="mx-badge ' + badgeCls(dispo) + '">' + dispo.replace(/</g, '&lt;') + '</span>';
    set('pv-cat', val('main'));
    set('pv-stock', val('stock') !== '' ? val('stock') : '—');
    var im = document.getElementById('pv-img');
    var ph = document.getElementById('pv-img-ph');
    var img = val('image');
    if (im && ph) {
      if (img !== '') {
        im.src = '/' + img;
        im.style.display = 'block';
        ph.style.display = 'none';
      } else {
        im.style.display = 'none';
        ph.style.display = 'block';
      }
    }
  }
  form.addEventListener('input', upd);
  form.addEventListener('change', upd);
  upd();
})();

// Bascule d'affichage liste/grille/miniatures (préférence mémorisée).
(function () {
  var wrap = document.querySelector('[data-mx-aff]');
  if (!wrap) return;
  var btns = document.querySelectorAll('[data-aff]');
  function apply(v) {
    wrap.setAttribute('data-mx-aff', v);
    btns.forEach(function (b) { b.classList.toggle('on', b.getAttribute('data-aff') === v); });
    try { localStorage.setItem('mx-aff', v); } catch (e) {}
  }
  btns.forEach(function (b) {
    b.addEventListener('click', function () { apply(b.getAttribute('data-aff')); });
  });
  try {
    var saved = localStorage.getItem('mx-aff');
    if (saved === 'grid' || saved === 'minis' || saved === 'list') apply(saved);
  } catch (e) {}
})();

// Aperçu live du Studio Design : recharge l'iframe en GET avec les réglages
// du formulaire (?couleur=...&sections=...). Navigation normale : pas de
// doc.write, pas d'URL blob, scripts et styles chargés normalement.
// Repli sans JS : submit natif vers l'iframe (POST).
(function () {
  var form = document.getElementById('mx-design-form');
  if (!form) return;
  var frame = document.querySelector('iframe[name="mx-preview"]');
  if (!frame) return;
  var timer = null, seq = 0;
  function refresh() {
    var my = ++seq;
    var fd = new FormData(form);
    fd.delete('csrf');
    var qs = new URLSearchParams(fd).toString();
    // Anti-course : seule la dernière demande s'applique.
    setTimeout(function () {
      if (my !== seq) return;
      frame.src = '/admin/design/apercu?' + qs;
    }, 50);
  }
  var btn = form.querySelector('[formaction="/admin/design/apercu"]');
  if (btn) btn.addEventListener('click', function (e) { e.preventDefault(); refresh(); });
  form.addEventListener('change', function () {
    if (timer) clearTimeout(timer);
    timer = setTimeout(refresh, 700);
  });
  form.addEventListener('input', function (e) {
    if (e.target && ['accent', 'annonce_texte', 'css', 'ordre'].indexOf(e.target.name) >= 0) {
      if (timer) clearTimeout(timer);
      timer = setTimeout(refresh, 900);
    }
  });
})();

// Galerie fiche produit : bouton "+" qui ajoute un volet d'upload par image
// (un fichier par volet, nom "photos" identique au serveur), plafond 4
// extras. Sans JS : aucun volet, les extras existants restent gérables.
(function () {
  var box = document.getElementById('mx-gal-slots');
  if (!box) return;
  var add = document.getElementById('mx-gal-add');
  if (!add) return;
  var max = parseInt(box.getAttribute('data-gal-free') || '0', 10);
  if (max < 0) max = 0;
  function count() {
    return box.querySelectorAll('[data-gal-slot]').length;
  }
  function refresh() {
    var left = max - count();
    if (left <= 0) {
      add.style.display = 'none';
      return;
    }
    add.style.display = '';
    add.textContent = '+ Ajouter une image (' + left + ' restante' + (left > 1 ? 's' : '') + ')';
  }
  add.addEventListener('click', function () {
    if (count() >= max) return;
    var d = document.createElement('div');
    d.className = 'mx-gslot';
    d.setAttribute('data-gal-slot', '');
    var n = count() + 1;
    var lab = document.createElement('span');
    lab.className = 'mx-gslot-t';
    lab.textContent = 'Image +' + n;
    var inp = document.createElement('input');
    inp.type = 'file';
    inp.name = 'photos';
    inp.accept = '.jpg,.jpeg,.png,.webp';
    var prev = document.createElement('img');
    prev.alt = '';
    prev.style.display = 'none';
    var rm = document.createElement('button');
    rm.type = 'button';
    rm.className = 'mx-btn ghost';
    rm.setAttribute('data-gal-rm', '');
    rm.textContent = 'Retirer';
    d.appendChild(lab);
    d.appendChild(inp);
    d.appendChild(prev);
    d.appendChild(rm);
    box.appendChild(d);
    refresh();
  });
  box.addEventListener('change', function (e) {
    var t = e.target;
    if (!t || t.name !== 'photos' || !t.files || !t.files[0]) return;
    var slot = t.closest('[data-gal-slot]');
    if (!slot) return;
    var prev = slot.querySelector('img');
    if (!prev) return;
    try {
      prev.src = URL.createObjectURL(t.files[0]);
      prev.style.display = 'block';
    } catch (err) {}
  });
  box.addEventListener('click', function (e) {
    var rm = e.target.closest('[data-gal-rm]');
    if (!rm) return;
    var slot = rm.closest('[data-gal-slot]');
    if (slot) slot.remove();
    refresh();
  });
  refresh();
})();

// Types d'équipements : + (focus + sélection), renommer inline,
// ajouter (ouvre le + de la carte), fermeture des autres menus ⋯.
// Les formulaires postent en natif : marche aussi sans JS via <details>.
(function () {
  function csrf() {
    var el = document.querySelector('form input[name="csrf"]');
    return el ? el.value : '';
  }
  document.addEventListener('toggle', function (e) {
    var d = e.target;
    if (!(d instanceof HTMLDetailsElement)) return;
    if (d.hasAttribute('data-types-add') && d.open) {
      var inp = d.querySelector('input[name="sub"]');
      if (inp) { inp.focus(); inp.select(); }
      return;
    }
    if (d.hasAttribute('data-types-menu') && d.open) {
      document.querySelectorAll('details[data-types-menu][open]').forEach(function (o) {
        if (o !== d) o.open = false;
      });
    }
  }, true);
  document.addEventListener('click', function (e) {
    var b = e.target.closest('[data-types-act]');
    if (!b) return;
    var li = b.closest('[data-types-row]');
    if (!li) return;
    var top = li.getAttribute('data-top') || '';
    var sub = li.getAttribute('data-sub') || '';
    if (b.getAttribute('data-types-act') === 'ajouter') {
      var card = b.closest('[data-types-card]');
      var add = card ? card.querySelector('details[data-types-add]') : null;
      if (add) {
        add.open = true;
        var inp = add.querySelector('input[name="sub"]');
        if (inp) { inp.focus(); inp.select(); }
      }
      var m = b.closest('details[data-types-menu]');
      if (m) m.open = false;
      return;
    }
    if (b.getAttribute('data-types-act') !== 'renommer' || li.querySelector('[data-types-rename]')) return;
    var link = li.querySelector('a');
    var form = document.createElement('form');
    form.method = 'post';
    form.setAttribute('data-types-rename', '');
    function hidden(n, v) {
      var i = document.createElement('input');
      i.type = 'hidden'; i.name = n; i.value = v;
      form.appendChild(i);
    }
    hidden('csrf', csrf());
    hidden('action', 'renommer');
    hidden('top', top);
    hidden('ancien', sub);
    var input = document.createElement('input');
    input.name = 'nouveau';
    input.value = sub;
    input.maxLength = 80;
    input.required = true;
    var ok = document.createElement('button');
    ok.className = 'mx-btn primary';
    ok.textContent = 'OK';
    var ko = document.createElement('button');
    ko.type = 'button';
    ko.className = 'mx-btn ghost';
    ko.textContent = 'Annuler';
    ko.addEventListener('click', function () { form.remove(); if (link) link.style.display = ''; });
    form.appendChild(input);
    form.appendChild(ok);
    form.appendChild(ko);
    input.addEventListener('keydown', function (ev) {
      if (ev.key === 'Escape') { form.remove(); if (link) link.style.display = ''; }
    });
    if (link) link.style.display = 'none';
    li.insertBefore(form, li.firstChild);
    input.focus();
    input.select();
    var m2 = b.closest('details[data-types-menu]');
    if (m2) m2.open = false;
  });
})();

// Flèches de repliage des modules : injectées dans chaque titre de module
// (.mx-postbox, .mx-fsec, .mx-card, carte de type), le handler data-mx-toggle
// existant replie la carte (CSS .closed). Sans JS : tout reste déplié.
(function () {
  document.querySelectorAll('.mx-postbox > h2, .mx-fsec > h3, .mx-card > h2, .mx-type-head > h2').forEach(function (h) {
    if (h.querySelector('[data-mx-toggle]')) return;
    var b = document.createElement('button');
    b.type = 'button';
    b.className = 'mx-fold';
    b.setAttribute('data-mx-toggle', '');
    b.setAttribute('aria-label', 'Replier / déplier');
    b.textContent = '▾';
    h.appendChild(b);
  });
})();

// Postboxes repliables + confirmations destructives (dupliquer, supprimer).
(function () {
  document.addEventListener('click', function (e) {
    var t = e.target.closest('[data-mx-toggle]');
    if (t) {
      var box = t.closest('.mx-postbox,.mx-fsec,.mx-card,.mx-type');
      if (box) box.classList.toggle('closed');
      return;
    }
    var del = e.target.closest('[data-mx-confirm]');
    if (del && !window.confirm(del.getAttribute('data-mx-confirm') || 'Confirmer ?')) {
      e.preventDefault();
    }
  });
})();

// Sidebar mobile : trois façons de la ranger — icône ✕ dans la sidebar,
// appui dans le vide (voile .mx-scrim) et glisser vers la gauche.
(function () {
  function close() { document.body.classList.remove('nav-open'); }
  document.addEventListener('click', function (e) {
    if (e.target.closest('[data-side-close]')) close();
  });
  var sx = 0, sy = 0, tracking = false;
  document.addEventListener('touchstart', function (e) {
    if (!document.body.classList.contains('nav-open')) return;
    var t = e.touches[0];
    sx = t.clientX; sy = t.clientY; tracking = true;
  }, { passive: true });
  document.addEventListener('touchend', function (e) {
    if (!tracking) return;
    tracking = false;
    var t = e.changedTouches[0];
    if (t.clientX - sx < -50 && Math.abs(t.clientY - sy) < 60) close();
  }, { passive: true });
})();

/* Pagination locale des tableaux de groupe : 10 lignes par page, flèches
   première / précédente / suivante / dernière. Aucune requête serveur. */
(function () {
  function build(root) {
    var per = parseInt(root.getAttribute('data-per'), 10) || 10;
    var tbody = root.querySelector('tbody');
    if (!tbody) return;
    var rows = Array.prototype.slice.call(tbody.querySelectorAll('tr'));
    if (rows.length <= per) return;
    var pages = Math.ceil(rows.length / per), page = 1;
    var pager = document.createElement('div');
    pager.className = 'mx-pager mx-pager-inline';
    function mk(label, title, fn) {
      var b = document.createElement('button');
      b.type = 'button'; b.className = 'mx-page-btn';
      b.textContent = label; b.title = title; b.setAttribute('aria-label', title);
      b.addEventListener('click', fn);
      return b;
    }
    var first = mk('«', 'Première page', function () { go(1); });
    var prev = mk('‹', 'Page précédente', function () { go(page - 1); });
    var info = document.createElement('span');
    info.className = 'mx-page-info';
    var next = mk('›', 'Page suivante', function () { go(page + 1); });
    var last = mk('»', 'Dernière page', function () { go(pages); });
    pager.appendChild(first); pager.appendChild(prev); pager.appendChild(info);
    pager.appendChild(next); pager.appendChild(last);
    root.appendChild(pager);
    function go(p) {
      page = Math.min(pages, Math.max(1, p));
      rows.forEach(function (tr, i) {
        tr.style.display = (i >= (page - 1) * per && i < page * per) ? '' : 'none';
      });
      info.textContent = 'Page ' + page + ' / ' + pages + ' · ' + rows.length + ' lignes';
      first.disabled = prev.disabled = (page === 1);
      next.disabled = last.disabled = (page === pages);
    }
    go(1);
  }
  function init() {
    Array.prototype.forEach.call(document.querySelectorAll('.mx-paged[data-per]'), build);
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
})();

// Menu avatar admin : ouvre au clic, ferme dehors / Échap. Déconnexion :
// l'admin est en HTTP Basic Auth (sans session serveur) : on recharge les
// identifiants avec un couple invalide, le navigateur oublie les bons et
// réaffiche la mire de connexion.
(function () {
  function menu() { return document.querySelector('[data-admin-menu]'); }
  function btn() { return document.querySelector('[data-admin-avatar]'); }
  function close() {
    var m = menu();
    if (m) m.hidden = true;
    var b = btn();
    if (b) b.setAttribute('aria-expanded', 'false');
  }
  var b = btn();
  if (b) b.addEventListener('click', function (e) {
    e.stopPropagation();
    var m = menu();
    if (!m) return;
    var open = m.hidden;
    close();
    if (open) { m.hidden = false; b.setAttribute('aria-expanded', 'true'); }
  });
  document.addEventListener('click', function (e) {
    var m = menu();
    if (m && !m.hidden && !e.target.closest('[data-admin-menu]') && !e.target.closest('[data-admin-avatar]')) close();
  });
  document.addEventListener('keydown', function (e) { if (e.key === 'Escape') close(); });
  document.querySelectorAll('[data-admin-logout]').forEach(function (el) {
    el.addEventListener('click', function () {
      close();
      try {
        window.location.href = window.location.protocol + '//deconnexion:deconnexion@' + window.location.host + '/admin';
      } catch (err) { window.location.href = '/admin'; }
    });
  });
})();
