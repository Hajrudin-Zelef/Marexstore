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
