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

// Postboxes repliables + confirmations destructives (dupliquer, supprimer).
(function () {
  document.addEventListener('click', function (e) {
    var t = e.target.closest('[data-mx-toggle]');
    if (t) {
      var box = t.closest('.mx-postbox');
      if (box) box.classList.toggle('closed');
      return;
    }
    var del = e.target.closest('[data-mx-confirm]');
    if (del && !window.confirm(del.getAttribute('data-mx-confirm') || 'Confirmer ?')) {
      e.preventDefault();
    }
  });
})();
