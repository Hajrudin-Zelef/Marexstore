/* Mediatec Store — interactions (onglets, diaporama, catalogue paginé, vues, recherche).
   Aucun gestionnaire en ligne : tout est branché ici, au chargement de la page. */
(function () {
  'use strict';

  function norm(s) {
    return (s || '').toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '').replace(/[^a-z0-9]+/g, ' ').trim();
  }

  function ready(fn) {
    if (document.readyState !== 'loading') fn();
    else document.addEventListener('DOMContentLoaded', fn);
  }

  ready(function () {
    /* ---------- Onglets ---------- */
    document.querySelectorAll('[data-tabs]').forEach(function (box) {
      var btns = box.querySelectorAll('[data-tab]');
      btns.forEach(function (b) {
        b.addEventListener('click', function () {
          btns.forEach(function (x) { x.classList.remove('on'); });
          b.classList.add('on');
          box.querySelectorAll('[data-pane]').forEach(function (p) { p.hidden = true; });
          var pane = box.querySelector('[data-pane="' + b.getAttribute('data-tab') + '"]');
          if (pane) pane.hidden = false;
        });
      });
    });

    /* ---------- Diaporama ---------- */
    var slide = document.getElementById('slideshow');
    if (slide) {
      var slides = slide.querySelectorAll('.s');
      var cur = 0;
      var dots = slide.querySelectorAll('[data-dot]');
      function show(i) {
        if (!slides.length) return;
        cur = (i + slides.length) % slides.length;
        slides.forEach(function (s, k) { s.classList.toggle('on', k === cur); });
        dots.forEach(function (d, k) { d.classList.toggle('on', k === cur); });
      }
      var timer = setInterval(function () { show(cur + 1); }, 5000);
      slide.querySelectorAll('[data-prev]').forEach(function (b) {
        b.addEventListener('click', function () { clearInterval(timer); show(cur - 1); });
      });
      slide.querySelectorAll('[data-next]').forEach(function (b) {
        b.addEventListener('click', function () { clearInterval(timer); show(cur + 1); });
      });
      dots.forEach(function (d, k) {
        d.addEventListener('click', function () { clearInterval(timer); show(k); });
      });
    }

    /* ---------- Révélation au défilement (fondu + montée douce) ---------- */
    var reveal = document.querySelectorAll('[data-reveal]');
    if ('IntersectionObserver' in window) {
      var io = new IntersectionObserver(function (entries) {
        entries.forEach(function (e) {
          if (e.isIntersecting) { e.target.classList.add('in'); io.unobserve(e.target); }
        });
      }, { threshold: 0.08 });
      reveal.forEach(function (el) { io.observe(el); });
    } else {
      reveal.forEach(function (el) { el.classList.add('in'); });
    }

    /* ---------- En-tête : ombre + flou renforcés au défilement ---------- */
    var mainHeader = document.querySelector('header.main');
    if (mainHeader) {
      var onScrollHeader = function () { mainHeader.classList.toggle('scrolled', window.scrollY > 8); };
      window.addEventListener('scroll', onScrollHeader, { passive: true });
      onScrollHeader();
    }

    /* ---------- Panier (volet latéral, conservé dans le navigateur) ---------- */
    var CART_KEY = 'mediatec-panier-v1';
    var WA_NUMBER = '2250705087507';
    function loadCart() { try { return JSON.parse(window.localStorage.getItem(CART_KEY)) || []; } catch (e) { return []; } }
    function saveCart(c) { try { window.localStorage.setItem(CART_KEY, JSON.stringify(c)); } catch (e) {} }
    var cart = loadCart();
    function fmt(n) { return n.toLocaleString('fr-FR').replace(/\u202F/g, ' ') + ' F CFA'; }
    function cartCount() { return cart.reduce(function (a, i) { return a + i.qty; }, 0); }
    function cartSubtotal() { return cart.reduce(function (a, i) { return a + i.qty * i.prix; }, 0); }
    var drawer = document.querySelector('[data-cart-drawer]');
    var overlay = document.querySelector('[data-cart-overlay]');
    var toastEl = document.querySelector('[data-toast]');
    var toastTimer = null;
    function showToast(msg) {
      if (!toastEl) return;
      toastEl.textContent = msg;
      toastEl.classList.add('show');
      clearTimeout(toastTimer);
      toastTimer = setTimeout(function () { toastEl.classList.remove('show'); }, 2600);
    }
    function openCart() { if (drawer) drawer.classList.add('open'); if (overlay) overlay.classList.add('open'); }
    function closeCart() { if (drawer) drawer.classList.remove('open'); if (overlay) overlay.classList.remove('open'); }
    function renderCart() {
      document.querySelectorAll('[data-cart-count]').forEach(function (el) {
        var n = String(cartCount());
        if (el.textContent !== n) {
          el.textContent = n;
          el.classList.remove('pop');
          void el.offsetWidth;
          el.classList.add('pop');
        }
      });
      document.querySelectorAll('[data-cart-total]').forEach(function (el) { el.textContent = fmt(cartSubtotal()); });
      document.querySelectorAll('[data-cart-subtotal]').forEach(function (el) { el.textContent = fmt(cartSubtotal()); });
      var box = document.querySelector('[data-cart-items]');
      if (!box) return;
      var prefix = drawer ? (drawer.getAttribute('data-prefix') || '') : '';
      if (!cart.length) {
        box.innerHTML = '<div class="cart-empty">Votre panier est vide.<br><span>Parcourez la boutique et ajoutez vos produits.</span></div>';
      } else {
        box.innerHTML = cart.map(function (it, i) {
          var im = it.image ? '<img src="' + prefix + it.image + '" alt="">' : '<div class="noimg">Photo à venir</div>';
          return '<div class="cart-item">' + im + '<div class="ci-info"><h4>' + it.nom + '</h4>'
            + '<div class="ci-pr">' + fmt(it.prix) + '</div>'
            + '<div class="ci-qty"><button type="button" data-cart-dec="' + i + '" aria-label="Moins">−</button>'
            + '<span>' + it.qty + '</span>'
            + '<button type="button" data-cart-inc="' + i + '" aria-label="Plus">+</button>'
            + '<button type="button" class="ci-del" data-cart-remove="' + i + '" aria-label="Retirer">×</button></div>'
            + '</div><div class="ci-line">' + fmt(it.qty * it.prix) + '</div></div>';
        }).join('');
      }
      var order = document.querySelector('[data-cart-order]');
      if (order) {
        if (!cart.length) { order.classList.add('disabled'); order.removeAttribute('href'); }
        else {
          order.classList.remove('disabled');
          var lines = cart.map(function (it) { return '• ' + it.qty + ' × ' + it.nom + ' — ' + fmt(it.qty * it.prix); });
          var msg = 'Bonjour, je voudrais commander :\n' + lines.join('\n') + '\nTotal : ' + fmt(cartSubtotal());
          order.setAttribute('href', 'https://wa.me/' + WA_NUMBER + '?text=' + encodeURIComponent(msg));
        }
      }
    }
    document.addEventListener('click', function (e) {
      var add = e.target.closest('[data-add-cart]');
      if (add) {
        var q = 1;
        var qSel = add.getAttribute('data-qty');
        if (qSel) { var inp = document.querySelector(qSel); if (inp) q = Math.max(1, parseInt(inp.value, 10) || 1); }
        var slug = add.getAttribute('data-slug');
        var found = cart.find(function (it) { return it.slug === slug; });
        if (found) found.qty += q;
        else cart.push({ slug: slug, nom: add.getAttribute('data-nom'), prix: parseInt(add.getAttribute('data-prix'), 10) || 0, image: add.getAttribute('data-image') || '', qty: q });
        saveCart(cart);
        renderCart();
        showToast('Ajouté au panier : ' + add.getAttribute('data-nom'));
        openCart();
        return;
      }
      if (e.target.closest('[data-cart-open]')) { renderCart(); openCart(); return; }
      if (e.target.closest('[data-cart-close]') || e.target.closest('[data-cart-overlay]')) { closeCart(); return; }
      if (e.target.closest('[data-cart-clear]')) { cart = []; saveCart(cart); renderCart(); return; }
      var inc = e.target.closest('[data-cart-inc]');
      if (inc) { cart[+inc.getAttribute('data-cart-inc')].qty += 1; saveCart(cart); renderCart(); return; }
      var dec = e.target.closest('[data-cart-dec]');
      if (dec) { var it = cart[+dec.getAttribute('data-cart-dec')]; it.qty -= 1; if (it.qty < 1) cart.splice(+dec.getAttribute('data-cart-dec'), 1); saveCart(cart); renderCart(); return; }
      var rem = e.target.closest('[data-cart-remove]');
      if (rem) { cart.splice(+rem.getAttribute('data-cart-remove'), 1); saveCart(cart); renderCart(); return; }
    });
    document.addEventListener('keydown', function (e) { if (e.key === 'Escape') { closeCart(); closeAuth(); } });
    renderCart();

    /* ---------- Modale Connexion / Créer un compte ---------- */
    var authModal = document.querySelector('[data-auth-modal]');
    var authOverlay = document.querySelector('[data-auth-overlay]');
    function openAuth(tab) {
      if (!authModal) return;
      if (tab) showAuthTab(tab);
      authModal.classList.add('open');
      if (authOverlay) authOverlay.classList.add('open');
      var pane = authModal.querySelector('[data-auth-pane="' + (tab || 'login') + '"] input');
      if (pane) pane.focus();
    }
    function closeAuth() {
      if (!authModal) return;
      authModal.classList.remove('open');
      if (authOverlay) authOverlay.classList.remove('open');
    }
    function showAuthTab(tab) {
      document.querySelectorAll('[data-auth-tab]').forEach(function (b) { b.classList.toggle('on', b.getAttribute('data-auth-tab') === tab); });
      document.querySelectorAll('[data-auth-pane]').forEach(function (p) { p.hidden = (p.getAttribute('data-auth-pane') !== tab); });
    }
    document.querySelectorAll('[data-auth-tab]').forEach(function (b) {
      b.addEventListener('click', function () { showAuthTab(b.getAttribute('data-auth-tab')); });
    });
    document.querySelectorAll('[data-auth-close]').forEach(function (b) { b.addEventListener('click', closeAuth); });
    if (authOverlay) authOverlay.addEventListener('click', closeAuth);
    document.addEventListener('click', function (e) {
      var opener = e.target.closest('[data-auth-open]');
      if (opener) {
        if (loadAcct() && sessionOn()) return; // connecté : le lien mène à compte.html
        e.preventDefault();
        openAuth(opener.getAttribute('data-auth-tab') || 'login');
        return;
      }
    });

    /* ---------- Suivi de commande : prépare le message WhatsApp ---------- */
    var suivi = document.querySelector('[data-suivi-form]');
    if (suivi) {
      suivi.addEventListener('submit', function (e) {
        e.preventDefault();
        var fd = new FormData(suivi);
        var msg = 'Bonjour, je souhaite suivre ma commande n° ' + (fd.get('commande') || '') + (fd.get('telephone') ? ' (tél : ' + fd.get('telephone') + ')' : '');
        window.open('https://wa.me/' + WA_NUMBER + '?text=' + encodeURIComponent(msg), '_blank');
      });
    }

    /* ---------- Recherche : suggestions automatiques dans l'en-tête ----------
       Quand on tape dans la recherche de l'en-tête, les produits ciblés
       s'affichent en propositions (API serveur /api/recherche, FTS).
       Un clic va directement sur la fiche ; Entrée lance la recherche normale.
       Si l'API ne répond pas, rien ne s'affiche et le formulaire marche comme avant. */
    (function () {
      var form = document.querySelector('form.search');
      if (!form) return;
      var input = form.querySelector('input[name="q"]');
      var catSel = form.querySelector('select[name="cat"]');
      if (!input) return;
      input.setAttribute('autocomplete', 'off');
      // Base des URLs ("" sur la racine, "../" sur les sous-pages) : déduite de l'action du formulaire.
      var action = form.getAttribute('action') || 'boutique.html';
      var base = action.replace(/boutique\.html$/, '');
      // Le menu est rattaché à l'en-tête (position:sticky), PAS au formulaire :
      // le formulaire a overflow:hidden (pilule arrondie) qui couperait le menu.
      // Positionné sous le formulaire via ses coordonnées dans l'en-tête.
      var header = form.closest('header.main') || form.parentNode;
      var st = document.createElement('style');
      st.textContent = '.search-suggest{position:absolute;z-index:60;background:#fff;border:1px solid #d8e2d8;border-radius:10px;box-shadow:0 10px 30px rgba(0,0,0,.18);overflow:hidden;display:none;max-height:340px;overflow-y:auto;text-align:left}'
        + '.search-suggest.open{display:block}'
        + '.search-suggest a{display:flex;align-items:center;gap:10px;padding:8px 10px;text-decoration:none;color:#1a1a1a;border-top:1px solid #eef2ee}'
        + '.search-suggest a:first-child{border-top:0}'
        + '.search-suggest a.sel,.search-suggest a:hover{background:#e9efe9}'
        + '.search-suggest img{width:44px;height:44px;object-fit:contain;flex:none;background:#f4f6f4;border-radius:6px}'
        + '.search-suggest .sg-noimg{width:44px;height:44px;flex:none;background:#f4f6f4;border-radius:6px;display:flex;align-items:center;justify-content:center;color:#888;font-size:11px}'
        + '.search-suggest .sg-nom{flex:1;min-width:0;font-size:14px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}'
        + '.search-suggest .sg-prix{flex:none;font-size:13px;font-weight:700;color:#2e5136;white-space:nowrap}'
        + '.search-suggest .sg-prix.dev{font-weight:400;color:#666}'
        + '.search-suggest .sg-empty{padding:10px 12px;color:#666;font-size:14px}';
      document.head.appendChild(st);
      var box = document.createElement('div');
      box.className = 'search-suggest';
      box.setAttribute('role', 'listbox');
      header.appendChild(box);
      var timer = null, sel = -1, current = [];
      function place() {
        box.style.left = form.offsetLeft + 'px';
        box.style.top = (form.offsetTop + form.offsetHeight + 6) + 'px';
        box.style.width = form.offsetWidth + 'px';
      }
      window.addEventListener('resize', function () { if (box.classList.contains('open')) place(); });
      function esc(s) {
        return String(s == null ? '' : s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
      }
      function fmtPrix(n) {
        return String(n).replace(/\B(?=(\d{3})+(?!\d))/g, ' ') + ' F CFA';
      }
      function hide() { box.classList.remove('open'); box.innerHTML = ''; sel = -1; current = []; }
      function render(results) {
        current = results || [];
        sel = -1;
        if (!current.length) {
          box.innerHTML = '<div class="sg-empty">Aucun produit trouvé pour « ' + esc(input.value) + ' ».</div>';
        } else {
          var html = '';
          current.forEach(function (r, i) {
            var img = r.image ? '<img src="' + base + esc(r.image) + '" alt="">' : '<span class="sg-noimg">—</span>';
            var prix = (r.prix === null || r.prix === undefined)
              ? '<span class="sg-prix dev">Prix sur demande</span>'
              : '<span class="sg-prix">' + fmtPrix(r.prix) + '</span>';
            html += '<a href="' + base + 'produits/' + encodeURIComponent(r.slug) + '.html" role="option" data-i="' + i + '">' + img + '<span class="sg-nom">' + esc(r.nom) + '</span>' + prix + '</a>';
          });
          box.innerHTML = html;
        }
        box.classList.add('open');
        place();
      }
      function fetchSuggest(q) {
        var url = base + 'api/recherche?q=' + encodeURIComponent(q);
        if (catSel && catSel.value) url += '&cat=' + encodeURIComponent(catSel.value);
        fetch(url).then(function (resp) { return resp.json(); }).then(function (d) {
          if (input.value.trim() !== q) return; // une frappe plus récente a eu lieu
          render((d && d.results) || []);
        }).catch(function () { hide(); });
      }
      input.addEventListener('input', function () {
        var q = input.value.trim();
        if (timer) clearTimeout(timer);
        if (q.length < 2) { hide(); return; }
        timer = setTimeout(function () { fetchSuggest(q); }, 200);
      });
      input.addEventListener('keydown', function (e) {
        if (!box.classList.contains('open')) return;
        var links = box.querySelectorAll('a');
        if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
          e.preventDefault();
          sel = e.key === 'ArrowDown' ? Math.min(sel + 1, links.length - 1) : Math.max(sel - 1, 0);
          links.forEach(function (a, i) { a.classList.toggle('sel', i === sel); });
        } else if (e.key === 'Enter' && sel >= 0 && links[sel]) {
          e.preventDefault();
          window.location.href = links[sel].getAttribute('href');
        } else if (e.key === 'Escape') {
          hide();
          input.blur();
        }
      });
      box.addEventListener('mousedown', function (e) {
        // mousedown (avant blur) pour que le clic parte avant la fermeture.
        var a = e.target.closest ? e.target.closest('a') : null;
        if (a) window.location.href = a.getAttribute('href');
      });
      document.addEventListener('click', function (e) {
        if (!form.contains(e.target) && !box.contains(e.target)) hide();
      });
      input.addEventListener('blur', function () { setTimeout(hide, 150); });
    })();

    /* ---------- Catalogue : pagination + vues + recherche + sous-catégories ---------- */
    document.querySelectorAll('[data-catalog-wrap]').forEach(function (wrap) {
      var grid = wrap.querySelector('[data-catalog]');
      if (!grid) return;
      var items = Array.prototype.slice.call(grid.querySelectorAll('.card'));
      var pager = wrap.querySelector('[data-pager]');
      var info = wrap.querySelector('[data-page-info]');
      var perPageSel = wrap.querySelector('[data-per-page]');
      var searchInput = wrap.querySelector('[data-filter]');
      var state = { page: 1, perPage: perPageSel ? parseInt(perPageSel.value, 10) : 12, q: '', sub: 'all', store: 'all', liveSubs: [], main: 'all', marque: '', prixMin: null, prixMax: null, dispo: '', etat: '', specs: {}, sort: 'defaut' };
      var baseOrder = items.slice();

      function visibleItems() {
        return items.filter(function (c) {
          if (state.sub !== 'all' && c.getAttribute('data-sub') !== state.sub) return false;
          if (state.store !== 'all' && (' ' + (c.getAttribute('data-store') || '') + ' ').indexOf(' ' + state.store + ' ') === -1) return false;
          if (state.liveSubs.length && state.liveSubs.indexOf(c.getAttribute('data-live-sub')) === -1) return false;
          if (state.main !== 'all' && c.getAttribute('data-main') !== state.main) return false;
          if (state.marque && c.getAttribute('data-marque') !== state.marque) return false;
          if (state.dispo && c.getAttribute('data-dispo') !== state.dispo) return false;
          if (state.etat && c.getAttribute('data-etat') !== state.etat) return false;
          for (var sk in state.specs) {
            if (!state.specs[sk]) continue;
            var sv = c.getAttribute('data-spec-' + sk) || '';
            if ((' ' + sv + ' ').indexOf(' ' + state.specs[sk] + ' ') === -1) return false;
          }
          if (state.prixMin !== null || state.prixMax !== null) {
            var p = parseInt(c.getAttribute('data-prix'), 10);
            if (isNaN(p)) return false;
            if (state.prixMin !== null && p < state.prixMin) return false;
            if (state.prixMax !== null && p > state.prixMax) return false;
          }
          if (state.q) {
            var sq = c.getAttribute('data-search') || c.getAttribute('data-name') || '';
            var sqc = sq.replace(/[^a-z0-9]/g, '');
            var qw = state.q.split(' ');
            for (var wi = 0; wi < qw.length; wi++) {
              if (sq.indexOf(qw[wi]) === -1 && sqc.indexOf(qw[wi]) === -1) return false;
            }
          }
          return true;
        });
      }

      function renderPager(totalPages) {
        if (!pager) return;
        pager.innerHTML = '';
        if (totalPages <= 1) return;
        function btn(label, page, cls, disabled) {
          var b = document.createElement('button');
          b.type = 'button';
          b.textContent = label;
          if (cls) b.className = cls;
          b.disabled = !!disabled;
          b.addEventListener('click', function () { state.page = page; apply(); if (wrap.scrollIntoView) wrap.scrollIntoView({ block: 'start' }); });
          pager.appendChild(b);
        }
        btn('‹', Math.max(1, state.page - 1), 'pg-prev', state.page === 1);
        var start = Math.max(1, Math.min(state.page - 2, totalPages - 4));
        var end = Math.min(totalPages, start + 4);
        for (var p = start; p <= end; p++) {
          btn(String(p), p, p === state.page ? 'pg-cur' : '');
        }
        btn('›', Math.min(totalPages, state.page + 1), 'pg-next', state.page === totalPages);
      }

      function apply() {
        var vis = visibleItems();
        var totalPages = state.perPage > 0 ? Math.max(1, Math.ceil(vis.length / state.perPage)) : 1;
        if (state.page > totalPages) state.page = totalPages;
        items.forEach(function (c) { c.style.display = 'none'; });
        var slice = state.perPage > 0 ? vis.slice((state.page - 1) * state.perPage, state.page * state.perPage) : vis;
        slice.forEach(function (c) { c.style.display = ''; });
        renderPager(state.perPage > 0 ? totalPages : 1);
        if (info) {
          info.textContent = vis.length === 0 ? 'Aucun produit trouvé.'
            : (state.perPage > 0 ? 'Page ' + state.page + ' sur ' + totalPages + ' — ' : '') + vis.length + ' produit' + (vis.length > 1 ? 's' : '');
        }
        syncHub();
      }

      /* Hub des catégories (boutique.html) : la liste des produits ne s'affiche
         que quand une catégorie, une sous-catégorie, une recherche ou un filtre
         est actif — sinon on montre les cartes des catégories. */
      var hub = document.querySelector('[data-hub]');
      var catHome = document.querySelector('[data-catalog-home]');
      function syncHub() {
        if (!hub || !catHome) return;
        var filtered = state.store !== 'all' || state.liveSubs.length || state.sub !== 'all' || state.main !== 'all' || !!state.q || !!state.marque || state.prixMin !== null || state.prixMax !== null || !!state.dispo || !!state.etat || Object.keys(state.specs).some(function (k) { return !!state.specs[k]; }) || new URLSearchParams(window.location.search).get('all');
        hub.hidden = !!filtered;
        catHome.hidden = !filtered;
      }

      if (perPageSel) {
        perPageSel.addEventListener('change', function () {
          state.perPage = parseInt(perPageSel.value, 10) || 0;
          state.page = 1;
          apply();
        });
      }
      if (searchInput) {
        searchInput.addEventListener('input', function () {
          state.q = norm(searchInput.value);
          var fq = document.querySelector('[data-f-q]');
          if (fq) fq.value = searchInput.value;
          state.page = 1;
          apply();
        });
        var preset = new URLSearchParams(window.location.search).get('q');
        if (preset) { searchInput.value = preset; state.q = norm(preset); }
      }
      wrap.querySelectorAll('[data-sub]').forEach(function (chip) {
        chip.addEventListener('click', function () {
          wrap.querySelectorAll('[data-sub]').forEach(function (x) { x.classList.remove('on'); });
          chip.classList.add('on');
          state.sub = chip.getAttribute('data-sub');
          state.page = 1;
          apply();
        });
      });
      wrap.querySelectorAll('[data-store]').forEach(function (chip) {
        chip.addEventListener('click', function () {
          wrap.querySelectorAll('[data-store]').forEach(function (x) { x.classList.remove('on'); });
          chip.classList.add('on');
          state.store = chip.getAttribute('data-store');
          state.page = 1;
          apply();
        });
      });
      /* Présélection depuis l'URL : ?cat= (recherche de l'en-tête), ?store= (catégorie),
         ?sub= (sous-catégorie, ou plusieurs séparées par des virgules — « Voir + » d'un groupe). */
      var params = new URLSearchParams(window.location.search);
      var presetCat = params.get('cat');
      if (presetCat) state.main = presetCat;
      var presetStore = params.get('store');
      if (presetStore) state.store = presetStore;
      var presetSub = params.get('sub');
      if (presetSub) state.liveSubs = presetSub.split(',').map(function (s) { return s.trim(); }).filter(Boolean);
      /* Présélections du filtre global de l'accueil et du hub des types :
         ?type= (type d'équipement, sous-catégorie), ?marque=, ?prixmin=,
         ?prixmax=, ?dispo= — les contrôles latéraux sont synchronisés juste après. */
      var presetType = params.get('type');
      if (presetType) state.sub = presetType;
      var presetMarque = params.get('marque');
      if (presetMarque) state.marque = presetMarque;
      var presetPrixMin = params.get('prixmin');
      if (presetPrixMin) { var vmin = parseInt(presetPrixMin, 10); if (vmin > 0) state.prixMin = vmin; }
      var presetPrixMax = params.get('prixmax');
      if (presetPrixMax) { var vmax = parseInt(presetPrixMax, 10); if (vmax > 0) state.prixMax = vmax; }
      var presetDispo = params.get('dispo');
      if (presetDispo) state.dispo = presetDispo;
      /* Synchronise les contrôles de filtre et la pastille du type avec les
         présélections URL (filtre global de l'accueil, hub des types). */
      (function () {
        var f = document.querySelector('[data-filters]');
        if (f) {
          var mq = f.querySelector('[data-f-marque]'); if (mq && state.marque) mq.value = state.marque;
          var pmn = f.querySelector('[data-f-prix-min]'); if (pmn && state.prixMin !== null) pmn.value = state.prixMin;
          var pmx = f.querySelector('[data-f-prix-max]'); if (pmx && state.prixMax !== null) pmx.value = state.prixMax;
          var rmn = f.querySelector('[data-f-prix-min-range]'); if (rmn && state.prixMin !== null) rmn.value = state.prixMin;
          var rmx = f.querySelector('[data-f-prix-max-range]'); if (rmx && state.prixMax !== null) rmx.value = state.prixMax;
          var ds = f.querySelector('[data-f-dispo]'); if (ds && state.dispo) ds.value = state.dispo;
          var sk = f.querySelector('[data-f-stock]'); if (sk) sk.checked = state.dispo === 'en-stock';
        }
        if (state.sub !== 'all') {
          wrap.querySelectorAll('[data-sub]').forEach(function (x) { x.classList.toggle('on', x.getAttribute('data-sub') === state.sub); });
        }
      })();
      wrap.querySelectorAll('[data-view]').forEach(function (b) {
        b.addEventListener('click', function () {
          wrap.querySelectorAll('[data-view]').forEach(function (x) { x.classList.remove('on'); });
          b.classList.add('on');
          grid.classList.remove('view-grid', 'view-list', 'view-thumbs');
          grid.classList.add('view-' + b.getAttribute('data-view'));
        });
      });
      /* Tri : réordonne les cartes (prix, nom) puis ré-applique la pagination. */
      var sortSel = wrap.querySelector('[data-sort]');
      if (sortSel) {
        sortSel.addEventListener('change', function () {
          state.sort = sortSel.value;
          var key = state.sort;
          if (key === 'defaut') {
            items = baseOrder.slice();
          } else {
            items.sort(function (a, b) {
              if (key === 'prix-asc' || key === 'prix-desc') {
                var pa = parseInt(a.getAttribute('data-prix'), 10);
                var pb = parseInt(b.getAttribute('data-prix'), 10);
                var aNaN = isNaN(pa);
                var bNaN = isNaN(pb);
                if (aNaN && bNaN) return 0;
                if (aNaN) return 1;
                if (bNaN) return -1;
                return key === 'prix-asc' ? pa - pb : pb - pa;
              }
              var na = a.getAttribute('data-name') || '';
              var nb = b.getAttribute('data-name') || '';
              if (key === 'nom-asc') return na < nb ? -1 : na > nb ? 1 : 0;
              return na > nb ? -1 : na < nb ? 1 : 0;
            });
          }
          items.forEach(function (c) { grid.appendChild(c); });
          state.page = 1;
          apply();
        });
      }
      /* Bouton « Filtres » discret sur mobile : replié par défaut, l'utilisateur l'ouvre au besoin. */
      document.querySelectorAll('[data-f-toggle]').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var panel = btn.nextElementSibling;
          while (panel && !panel.hasAttribute('data-filters')) panel = panel.nextElementSibling;
          if (!panel) panel = document.querySelector('[data-filters]');
          if (!panel) return;
          var open = panel.classList.toggle('open');
          btn.setAttribute('aria-expanded', open ? 'true' : 'false');
        });
      });
      /* Filtres latéraux (référence, marque, prix, disponibilité, état) — hors du bloc catalogue. */
      var filters = document.querySelector('[data-filters]');
      if (filters && !filters._wired) {
        filters._wired = true;
        var sel = function (n) { return filters.querySelector(n); };
        var fq = sel('[data-f-q]');
        if (fq) {
          if (searchInput && searchInput.value) fq.value = searchInput.value;
          fq.addEventListener('input', function () {
            state.q = norm(fq.value);
            if (searchInput) searchInput.value = fq.value;
            state.page = 1;
            apply();
          });
        }
        sel('[data-f-marque]').addEventListener('change', function (e) { state.marque = e.target.value; state.page = 1; apply(); });
        var moreBtn = sel('[data-f-more]');
        var extra = sel('[data-f-extra]');
        if (moreBtn && extra) {
          moreBtn.addEventListener('click', function () {
            extra.hidden = !extra.hidden;
            moreBtn.textContent = extra.hidden ? '+ DE FILTRES' : '− MOINS DE FILTRES';
          });
        }
        /* Facettes propres à la catégorie / sous-catégorie (facets.js) : listes
           déroulantes injectées selon le contexte — la catégorie choisie dans
           l'URL sur le Store, les sous-catégories de la page sur les pages
           catégories. Une facette n'apparaît que si elle a au moins 2 choix. */
        var facetsBox = sel('[data-f-facets]');
        var facetsMode = filters.getAttribute('data-facets-mode') || 'selection';
        var stockBox = sel('[data-f-stock]');
        var dispoSel = sel('[data-f-dispo]');
        function renderFacets() {
          if (!facetsBox || !window.MEDIATEC_FACETS) return;
          facetsBox.innerHTML = '';
          var keys = [];
          if (facetsMode === 'page') {
            var subs = {};
            items.forEach(function (c) { var s = c.getAttribute('data-live-sub'); if (s) subs[s] = true; });
            Object.keys(subs).forEach(function (s) { keys.push('sub:' + s); });
          } else if (state.liveSubs.length) {
            state.liveSubs.forEach(function (s) { keys.push('sub:' + s); });
          } else if (state.store !== 'all') {
            keys.push('top:' + state.store);
          }
          var merged = {};
          keys.forEach(function (key) {
            (window.MEDIATEC_FACETS[key] || []).forEach(function (fc) {
              var m = merged[fc.slug] || (merged[fc.slug] = { label: fc.label, options: {} });
              fc.options.forEach(function (o) { m.options[o.vs] = { v: o.v, n: (m.options[o.vs] ? m.options[o.vs].n : 0) + o.n }; });
            });
          });
          Object.keys(merged).forEach(function (slug) {
            var m = merged[slug];
            var opts = Object.keys(m.options).map(function (vs) { return { vs: vs, v: m.options[vs].v, n: m.options[vs].n }; });
            if (opts.length < 2) return;
            opts.sort(function (a, b) { return b.n - a.n || (a.v < b.v ? -1 : 1); });
            var lab = document.createElement('label');
            lab.textContent = m.label;
            var s = document.createElement('select');
            s.setAttribute('data-f-spec', slug);
            var o0 = document.createElement('option');
            o0.value = '';
            o0.textContent = 'Tous';
            s.appendChild(o0);
            opts.forEach(function (o) {
              var op = document.createElement('option');
              op.value = o.vs;
              op.textContent = o.v + ' (' + o.n + ')';
              s.appendChild(op);
            });
            s.addEventListener('change', function () { state.specs[slug] = s.value; state.page = 1; apply(); });
            lab.appendChild(s);
            facetsBox.appendChild(lab);
          });
        }
        renderFacets();
        /* Case « Voir uniquement les produits en stock » (synchronisée avec la
           liste Disponibilité, qui reste disponible derrière + DE FILTRES). */
        if (stockBox) {
          stockBox.addEventListener('change', function () {
            state.dispo = stockBox.checked ? 'en-stock' : '';
            if (dispoSel) dispoSel.value = state.dispo;
            state.page = 1;
            apply();
          });
        }
        sel('[data-f-dispo]').addEventListener('change', function (e) {
          state.dispo = e.target.value;
          if (stockBox) stockBox.checked = state.dispo === 'en-stock';
          state.page = 1;
          apply();
        });
        sel('[data-f-etat]').addEventListener('change', function (e) { state.etat = e.target.value; state.page = 1; apply(); });
        var pmin = sel('[data-f-prix-min]');
        var pmax = sel('[data-f-prix-max]');
        var rmin = sel('[data-f-prix-min-range]');
        var rmax = sel('[data-f-prix-max-range]');
        var prixMax = rmax ? parseInt(rmax.getAttribute('max'), 10) : 0;
        var prix = function () {
          var lo = pmin.value ? parseInt(pmin.value, 10) : 0;
          var hi = pmax.value ? parseInt(pmax.value, 10) : prixMax;
          if (rmin) rmin.value = lo;
          if (rmax) rmax.value = hi;
          state.prixMin = pmin.value ? lo : null;
          state.prixMax = pmax.value ? hi : null;
          state.page = 1;
          apply();
        };
        var prixRange = function (fromMin) {
          var lo = rmin ? parseInt(rmin.value, 10) : 0;
          var hi = rmax ? parseInt(rmax.value, 10) : prixMax;
          if (lo > hi) {
            if (fromMin) { rmin.value = hi; lo = hi; }
            else { rmax.value = lo; hi = lo; }
          }
          pmin.value = lo > 0 ? lo : '';
          pmax.value = hi < prixMax ? hi : '';
          state.prixMin = lo > 0 ? lo : null;
          state.prixMax = hi < prixMax ? hi : null;
          state.page = 1;
          apply();
        };
        pmin.addEventListener('change', prix);
        pmax.addEventListener('change', prix);
        if (rmin) rmin.addEventListener('input', function () { prixRange(true); });
        if (rmax) rmax.addEventListener('input', function () { prixRange(false); });
        sel('[data-f-reset]').addEventListener('click', function () {
          filters.querySelectorAll('select').forEach(function (s) { s.value = ''; });
          pmin.value = '';
          pmax.value = '';
          if (rmin) rmin.value = 0;
          if (rmax) rmax.value = prixMax;
          if (fq) fq.value = '';
          if (searchInput) searchInput.value = '';
          if (extra) extra.hidden = true;
          if (moreBtn) moreBtn.textContent = '+ DE FILTRES';
          if (stockBox) stockBox.checked = false;
          state.specs = {};
          state.marque = '';
          state.dispo = '';
          state.etat = '';
          state.prixMin = null;
          state.prixMax = null;
          state.q = '';
          state.page = 1;
          apply();
        });
      }
      apply();
    });

    /* ---------- Hub des catégories : le chevron déplie les sous-catégories ---------- */
    document.addEventListener('click', function (e) {
      var t = e.target.closest('[data-hub-toggle]');
      if (!t) return;
      var card = t.closest('[data-hub-card]');
      if (!card) return;
      var subs = card.querySelector('.hub-subs');
      card.classList.toggle('open');
      if (subs) subs.hidden = !card.classList.contains('open');
    });

    /* ---------- Filtre global de l'accueil : le choix du type dépend de la catégorie. ---------- */
    (function () {
      var form = document.querySelector('[data-home-filter] form');
      if (!form || !window.MEDIATEC_TYPES) return;
      var cat = form.querySelector('[data-hf-cat]');
      var typ = form.querySelector('[data-hf-type]');
      if (!cat || !typ) return;
      var fill = function () {
        var list = [];
        if (cat.value) {
          list = window.MEDIATEC_TYPES[cat.value] || [];
        } else {
          Object.keys(window.MEDIATEC_TYPES).forEach(function (k) { list = list.concat(window.MEDIATEC_TYPES[k]); });
        }
        typ.innerHTML = '<option value="">Tous les types</option>';
        list.forEach(function (t) {
          var o = document.createElement('option');
          o.value = t.slug;
          o.textContent = t.label;
          typ.appendChild(o);
        });
      };
      cat.addEventListener('change', fill);
      fill();
    })();

    /* ---------- Méga-menu « Toutes les Catégories » (comme LDLC : survol pour ouvrir) ---------- */
    var megaBtn = document.querySelector('[data-mega-btn]');
    var megaPanel = document.querySelector('[data-mega-panel]');
    if (megaBtn && megaPanel) {
      var closeMega = function () { megaPanel.hidden = true; megaBtn.setAttribute('aria-expanded', 'false'); megaHoverOpened = false; };
      var openMega = function () { megaPanel.hidden = false; megaBtn.setAttribute('aria-expanded', 'true'); };
      /* 1er clic : le survol (mouseenter) ouvre déjà le menu, puis le clic le
         refermait aussitôt (bascule) — le 1er clic semblait ne rien faire et
         il fallait cliquer une 2e fois. On ignore le clic qui suit juste
         l'ouverture par survol. */
      var megaHoverOpened = false;
      /* Survol des catégories : délai d'intention (120 ms). Sans délai, un
         trajet en diagonale vers « Voir tous » survole les autres catégories
         et le panneau change de catégorie en route. Un passage rapide ne
         déclenche plus le changement ; un survol posé, si. */
      var megaHoverTimer = null;
      /* Survol réservé aux appareils à souris : sur tactile, le tap émule
         mouseenter puis click, ce qui ouvrait puis refermait aussitôt le menu.
         Le test (hover: hover) ne suffit pas sur les portables à écran tactile
         (souris + tactile) : on ignore aussi tout survol juste après un toucher. */
      var canHover = window.matchMedia('(hover: hover)').matches;
      var lastTouch = 0;
      document.addEventListener('touchend', function () { lastTouch = Date.now(); }, { passive: true });
      var touchedRecently = function () { return Date.now() - lastTouch < 700; };
      megaBtn.addEventListener('mouseenter', function () { if (canHover && !touchedRecently()) { openMega(); megaHoverOpened = true; } });
      var navEl = megaBtn.closest('nav');
      if (navEl) navEl.addEventListener('mouseleave', function () { if (canHover && !touchedRecently()) closeMega(); });
      megaBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        if (megaHoverOpened) { megaHoverOpened = false; return; }
        if (megaPanel.hidden) openMega(); else closeMega();
      });
      megaPanel.querySelectorAll('[data-mega-top]').forEach(function (b) {
        var show = function () {
          megaPanel.querySelectorAll('[data-mega-top]').forEach(function (x) { x.classList.remove('on'); });
          b.classList.add('on');
          var slug = b.getAttribute('data-mega-top');
          megaPanel.querySelectorAll('[data-mega-cols]').forEach(function (c) { c.hidden = c.getAttribute('data-mega-cols') !== slug; });
        };
        b.addEventListener('click', function (e) {
          e.stopPropagation();
          /* Mobile (accordéon) : un second appui replie la catégorie. */
          if (megaMq.matches && b.classList.contains('on')) {
            b.classList.remove('on');
            var slug2 = b.getAttribute('data-mega-top');
            megaPanel.querySelectorAll('[data-mega-cols]').forEach(function (c) { if (c.getAttribute('data-mega-cols') === slug2) c.hidden = true; });
            return;
          }
          show();
        });
        b.addEventListener('mouseenter', function () {
          if (!canHover || megaMq.matches || touchedRecently()) return;
          if (megaHoverTimer) clearTimeout(megaHoverTimer);
          megaHoverTimer = setTimeout(function () { megaHoverTimer = null; show(); }, 120);
        });
        b.addEventListener('mouseleave', function () {
          if (megaHoverTimer) { clearTimeout(megaHoverTimer); megaHoverTimer = null; }
        });
      });
      /* Mobile : chaque bloc de sous-catégories déménage juste sous son bouton
         (accordéon) au lieu de s'afficher en bas après « Uncategorized ».
         Desktop : on remet tout en place (rail à gauche, contenu à droite). */
      var megaMq = window.matchMedia('(max-width:860px)');
      var megaRail = megaPanel.querySelector('.mega-rail');
      var megaColsWrap = megaRail ? megaRail.nextElementSibling : null;
      var megaCols = Array.prototype.slice.call(megaPanel.querySelectorAll('[data-mega-cols]'));
      function layoutMega() {
        if (megaMq.matches) {
          megaCols.forEach(function (c) {
            var btn = megaPanel.querySelector('[data-mega-top="' + c.getAttribute('data-mega-cols') + '"]');
            if (btn && c.previousElementSibling !== btn) btn.parentNode.insertBefore(c, btn.nextSibling);
          });
          megaPanel.querySelectorAll('[data-mega-top]').forEach(function (x) { x.classList.remove('on'); });
          megaCols.forEach(function (c) { c.hidden = true; });
        } else if (megaColsWrap) {
          megaCols.forEach(function (c) { megaColsWrap.appendChild(c); });
          var anyOn = megaPanel.querySelector('[data-mega-top].on');
          if (!anyOn) {
            var first = megaPanel.querySelector('[data-mega-top]');
            if (first) {
              first.classList.add('on');
              var s = first.getAttribute('data-mega-top');
              megaCols.forEach(function (c) { c.hidden = c.getAttribute('data-mega-cols') !== s; });
            }
          } else {
            var s2 = anyOn.getAttribute('data-mega-top');
            megaCols.forEach(function (c) { c.hidden = c.getAttribute('data-mega-cols') !== s2; });
          }
        }
      }
      if (megaMq.addEventListener) megaMq.addEventListener('change', layoutMega);
      else if (megaMq.addListener) megaMq.addListener(layoutMega);
      layoutMega();
      document.addEventListener('click', function (e) {
        if (!megaPanel.hidden && !megaPanel.contains(e.target)) closeMega();
      });
      document.addEventListener('keydown', function (e) { if (e.key === 'Escape') closeMega(); });
    }

    /* ---------- Formulaires : devis (WhatsApp) et contact (e-mail) ---------- */
    var WA_NUM = '2250705087507';
    var devis = document.querySelector('[data-devis-form]');
    if (devis) {
      /* Assistant en 4 étapes : une étape visible à la fois, validation avant de passer à la suivante. */
      var steps = devis.querySelectorAll('[data-step]');
      var indicators = document.querySelectorAll('[data-step-indicator] li');
      var current = 0;
      function showStep(i) {
        if (i < 0 || i >= steps.length) return;
        current = i;
        steps.forEach(function (s, k) { s.hidden = k !== i; });
        indicators.forEach(function (li, k) {
          li.classList.toggle('on', k === i);
          li.classList.toggle('done', k < i);
          var dot = li.querySelector('.dot');
          if (dot) dot.textContent = k < i ? '✓' : String(k + 1);
        });
        var card = devis.closest('.form-card');
        if (card && card.scrollIntoView) card.scrollIntoView({ block: 'start' });
      }
      function stepValid(i) {
        var bad = null;
        steps[i].querySelectorAll('[required]').forEach(function (el) {
          if (!el.checkValidity()) {
            el.classList.add('invalid');
            if (!bad) bad = el;
          } else {
            el.classList.remove('invalid');
          }
        });
        if (bad) {
          if (bad.reportValidity) bad.reportValidity();
          bad.focus();
          return false;
        }
        return true;
      }
      devis.querySelectorAll('[data-next-step]').forEach(function (b) {
        b.addEventListener('click', function () { if (stepValid(current)) showStep(current + 1); });
      });
      devis.querySelectorAll('[data-prev-step]').forEach(function (b) {
        b.addEventListener('click', function () { showStep(current - 1); });
      });
      showStep(0);
      devis.addEventListener('submit', function (e) {
        e.preventDefault();
        var g = function (n) { var el = devis.querySelector('[name="' + n + '"]'); return el ? el.value.trim() : ''; };
        var services = [];
        devis.querySelectorAll('[name="services"]:checked').forEach(function (c) { services.push(c.value); });
        var lines = ['Bonjour, je demande un devis :', '', 'Nom : ' + g('nom'), 'E-mail : ' + g('email'),
          'Téléphone : ' + g('telephone'), 'Entreprise : ' + g('entreprise'), '',
          'Type de client : ' + g('type_client'), 'Catégorie : ' + g('categorie'), 'État souhaité : ' + g('etat'),
          'Marque souhaitée : ' + g('marque'), 'Spécifications : ' + g('specs'), '',
          'Services associés : ' + (services.join(', ') || '—'), 'Délai : ' + g('delai'),
          'Budget : ' + g('budget'), 'Paiement : ' + g('paiement'), 'Connu via : ' + g('source'), '',
          'Message : ' + g('message')];
        window.open('https://wa.me/' + WA_NUM + '?text=' + encodeURIComponent(lines.join('\n')), '_blank');
      });
    }
    var contact = document.querySelector('[data-contact-form]');
    if (contact) {
      contact.addEventListener('submit', function (e) {
        e.preventDefault();
        var g = function (n) { var el = contact.querySelector('[name="' + n + '"]'); return el ? el.value.trim() : ''; };
        var body = 'Nom : ' + g('nom') + '\nE-mail : ' + g('email') + '\nTéléphone : ' + g('telephone') + '\n\n' + g('message');
        window.location.href = 'mailto:infos@mediatec-ci.com?subject=' + encodeURIComponent(g('objet')) + '&body=' + encodeURIComponent(body);
      });
    }

    /* ---------- Thèmes : sombre / clair + couleur Liquid / Gazonnier ---------- */
    var THEME_KEY = 'mediatec-theme';
    function loadTheme() { try { return JSON.parse(window.localStorage.getItem(THEME_KEY)) || {}; } catch (e) { return {}; } }
    function saveTheme(t) { try { window.localStorage.setItem(THEME_KEY, JSON.stringify(t)); } catch (e) {} }
    function applyTheme() {
      var t = loadTheme();
      var r = document.documentElement;
      if (t.mode === 'sombre') r.setAttribute('data-theme', 'dark'); else r.removeAttribute('data-theme');
      if (t.color === 'liquid') r.setAttribute('data-color', 'liquid'); else r.removeAttribute('data-color');
      document.querySelectorAll('[data-theme-choice]').forEach(function (el) { el.checked = (el.getAttribute('data-theme-choice') === (t.mode || 'clair')); });
      document.querySelectorAll('[data-color-choice]').forEach(function (el) { el.checked = (el.getAttribute('data-color-choice') === (t.color || 'gazonnier')); });
    }
    document.addEventListener('change', function (e) {
      var th = e.target.closest('[data-theme-choice]');
      if (th) { var t = loadTheme(); t.mode = th.getAttribute('data-theme-choice'); saveTheme(t); applyTheme(); return; }
      var co = e.target.closest('[data-color-choice]');
      if (co) { var t2 = loadTheme(); t2.color = co.getAttribute('data-color-choice'); saveTheme(t2); applyTheme(); return; }
    });
    applyTheme();

    /* ---------- Compte client (stocké dans ce navigateur, site statique) ---------- */
    var ACCT_KEY = 'mediatec-compte-v1';
    var SESSION_KEY = 'mediatec-session-v1';
    var ORDERS_KEY = 'mediatec-commandes-v1';
    var CARTS_KEY = 'mediatec-paniers-v1';
    function loadAcct() { try { return JSON.parse(window.localStorage.getItem(ACCT_KEY)); } catch (e) { return null; } }
    function saveAcct(a) { try { if (a) window.localStorage.setItem(ACCT_KEY, JSON.stringify(a)); else window.localStorage.removeItem(ACCT_KEY); } catch (e) {} }
    function sessionOn() { try { return window.localStorage.getItem(SESSION_KEY) === '1'; } catch (e) { return false; } }
    function saveSession(on) { try { if (on) window.localStorage.setItem(SESSION_KEY, '1'); else window.localStorage.removeItem(SESSION_KEY); } catch (e) {} }
    function loadList(k) { try { return JSON.parse(window.localStorage.getItem(k)) || []; } catch (e) { return []; } }
    function saveList(k, v) { try { window.localStorage.setItem(k, JSON.stringify(v)); } catch (e) {} }
    function escHtml(s) { return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); }
    function acctInitials(a) { return ((a.prenom || '?')[0] + (a.nom || '?')[0]).toUpperCase(); }
    function renderOrders() {
      var list = loadList(ORDERS_KEY);
      var box = document.querySelector('[data-acct-commandes]');
      var empty = document.querySelector('[data-acct-commandes-empty]');
      if (!box) return;
      box.innerHTML = list.map(function (o) {
        return '<li><div><b>' + escHtml(o.id) + '</b> — ' + escHtml(o.date) + '<br><span style="color:var(--muted)">' + o.items.map(function (i) { return escHtml(i.qty + ' × ' + i.nom); }).join(', ') + '</span></div><span class="pr">' + fmt(o.total) + '</span></li>';
      }).join('');
      if (empty) empty.hidden = list.length > 0;
    }
    function renderAdresses() {
      var a = loadAcct(); if (!a) return;
      var box = document.querySelector('[data-acct-adresses]');
      var empty = document.querySelector('[data-acct-adresses-empty]');
      if (!box) return;
      var ads = a.adresses || [];
      box.innerHTML = ads.map(function (ad, i) {
        return '<li><div>' + escHtml(ad.rue) + '<br>' + escHtml(ad.ville) + (ad.tel ? ' — ' + escHtml(ad.tel) : '') + '</div><button type="button" class="linklike" data-adresse-del="' + i + '">Supprimer</button></li>';
      }).join('');
      if (empty) empty.hidden = ads.length > 0;
    }
    function renderPaniers() {
      var list = loadList(CARTS_KEY);
      var box = document.querySelector('[data-acct-paniers]');
      var empty = document.querySelector('[data-acct-paniers-empty]');
      if (!box) return;
      box.innerHTML = list.map(function (p, i) {
        return '<li><div><b>' + escHtml(p.nom) + '</b> — ' + escHtml(p.date) + '<br><span style="color:var(--muted)">' + p.items.length + ' article(s)</span></div><span><span class="pr">' + fmt(p.total) + '</span> <button type="button" class="linklike" data-panier-load="' + i + '">Reprendre</button> <button type="button" class="linklike" data-panier-del="' + i + '">Supprimer</button></span></li>';
      }).join('');
      if (empty) empty.hidden = list.length > 0;
    }
    function renderParamGate() {
      var locked = document.querySelector('[data-param-locked]');
      var content = document.querySelector('[data-param-content]');
      if (!locked || !content) return;
      var ok = !!loadAcct() && sessionOn();
      locked.hidden = ok;
      content.hidden = !ok;
    }
    function renderAcct() {
      var a = loadAcct();
      var logged = !!a && sessionOn();
      renderParamGate();
      document.querySelectorAll('[data-compte-lab]').forEach(function (el) { el.textContent = logged && a.prenom ? a.prenom : 'Compte'; });
      var out = document.querySelector('[data-acct-logged-out]');
      var inn = document.querySelector('[data-acct-logged-in]');
      if (out) out.hidden = logged;
      if (inn) inn.hidden = !logged;
      document.querySelectorAll('[data-acct-id]').forEach(function (el) { el.hidden = !logged; });
      document.querySelectorAll('[data-acct-guest]').forEach(function (el) { el.hidden = logged; });
      if (!logged) return;
      document.querySelectorAll('[data-acct-initials]').forEach(function (el) { el.textContent = acctInitials(a); });
      document.querySelectorAll('[data-acct-name]').forEach(function (el) { el.textContent = (a.prenom || '') + ' ' + (a.nom || ''); });
      document.querySelectorAll('[data-acct-num]').forEach(function (el) { el.textContent = a.numero || ''; });
      var f = document.querySelector('[data-infos-form]');
      if (f) {
        f.querySelectorAll('[name="civilite"]').forEach(function (r) { r.checked = (r.value === a.civilite); });
        ['prenom', 'nom', 'email', 'naissance', 'pseudo'].forEach(function (n) { var el = f.querySelector('[name="' + n + '"]'); if (el) el.value = a[n] || ''; });
      }
      renderOrders(); renderAdresses(); renderPaniers();
    }
    function showPane(slug) {
      document.querySelectorAll('[data-acct-pane]').forEach(function (p) { p.hidden = (p.getAttribute('data-acct-pane') !== slug); });
      document.querySelectorAll('[data-acct-tab]').forEach(function (b) { b.classList.toggle('on', b.getAttribute('data-acct-tab') === slug); });
    }
    document.querySelectorAll('[data-acct-tab]').forEach(function (b) {
      b.addEventListener('click', function () { showPane(b.getAttribute('data-acct-tab')); });
    });
    if (document.querySelector('[data-acct-logged-in]')) {
      var h = (window.location.hash || '#informations').replace('#', '');
      if (['informations', 'commandes', 'adresses', 'paniers'].indexOf(h) === -1) h = 'informations';
      showPane(h);
    }
    var createForm = document.querySelector('[data-create-form]');
    if (createForm) {
      createForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var fd = new FormData(createForm);
        var err = document.querySelector('[data-create-error]');
        if ((fd.get('password') || '') !== (fd.get('password2') || '')) { if (err) { err.textContent = 'Les deux mots de passe ne correspondent pas.'; err.hidden = false; } return; }
        if (loadAcct()) { if (err) { err.textContent = 'Un compte existe déjà dans ce navigateur. Déconnectez-vous pour en créer un autre.'; err.hidden = false; } return; }
        var num = 'MED-' + String(Math.floor(100000 + Math.random() * 900000));
        saveAcct({ civilite: fd.get('civilite') || 'M.', prenom: (fd.get('prenom') || '').trim(), nom: (fd.get('nom') || '').trim().toUpperCase(), email: (fd.get('email') || '').trim(), naissance: '', pseudo: '', password: fd.get('password') || '', numero: num, adresses: [] });
        saveSession(true);
        renderAcct();
        closeAuth();
        showToast('Compte créé. Bienvenue ' + (fd.get('prenom') || '') + ' !');
      });
    }
    var loginForm = document.querySelector('[data-login-form]');
    if (loginForm) {
      loginForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var fd = new FormData(loginForm);
        var a = loadAcct();
        var err = document.querySelector('[data-login-error]');
        if (a && a.email === (fd.get('email') || '').trim() && a.password === (fd.get('password') || '')) { saveSession(true); renderAcct(); closeAuth(); showToast('Bonjour ' + (a.prenom || '') + ' !'); }
        else if (err) err.hidden = false;
      });
    }
    document.querySelectorAll('[data-logout]').forEach(function (b) {
      b.addEventListener('click', function () { saveSession(false); renderAcct(); showToast('Vous êtes déconnecté.'); });
    });
    var infosForm = document.querySelector('[data-infos-form]');
    if (infosForm) {
      infosForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var a = loadAcct(); if (!a) return;
        var fd = new FormData(infosForm);
        a.civilite = fd.get('civilite') || a.civilite;
        ['prenom', 'nom', 'email', 'naissance', 'pseudo'].forEach(function (n) { a[n] = (fd.get(n) || '').trim(); });
        a.nom = a.nom.toUpperCase();
        saveAcct(a); renderAcct(); showToast('Coordonnées enregistrées.');
      });
    }
    var passForm = document.querySelector('[data-pass-form]');
    if (passForm) {
      passForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var a = loadAcct(); if (!a) return;
        var fd = new FormData(passForm);
        var err = document.querySelector('[data-pass-error]');
        var ok = document.querySelector('[data-pass-ok]');
        if (err) err.hidden = true;
        if (ok) ok.hidden = true;
        if ((fd.get('actuel') || '') !== a.password) { if (err) err.hidden = false; return; }
        a.password = fd.get('nouveau') || '';
        saveAcct(a);
        if (ok) ok.hidden = false;
        passForm.reset();
      });
    }
    var adrForm = document.querySelector('[data-adresse-form]');
    if (adrForm) {
      adrForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var a = loadAcct(); if (!a) return;
        var fd = new FormData(adrForm);
        a.adresses = a.adresses || [];
        a.adresses.push({ rue: (fd.get('rue') || '').trim(), ville: (fd.get('ville') || '').trim(), tel: (fd.get('tel') || '').trim() });
        saveAcct(a); renderAdresses(); adrForm.reset(); showToast('Adresse ajoutée.');
      });
    }
    document.addEventListener('click', function (e) {
      var del = e.target.closest('[data-adresse-del]');
      if (del) { var a = loadAcct(); if (a) { a.adresses.splice(+del.getAttribute('data-adresse-del'), 1); saveAcct(a); renderAdresses(); } return; }
      var pl = e.target.closest('[data-panier-load]');
      if (pl) { var list = loadList(CARTS_KEY); var p = list[+pl.getAttribute('data-panier-load')]; if (p) { cart = JSON.parse(JSON.stringify(p.items)); saveCart(cart); renderCart(); showToast('Panier repris.'); } return; }
      var pd = e.target.closest('[data-panier-del]');
      if (pd) { var list2 = loadList(CARTS_KEY); list2.splice(+pd.getAttribute('data-panier-del'), 1); saveList(CARTS_KEY, list2); renderPaniers(); return; }
    });
    var savePanierBtn = document.querySelector('[data-save-panier]');
    if (savePanierBtn) {
      savePanierBtn.addEventListener('click', function () {
        if (!cart.length) { showToast('Votre panier est vide.'); return; }
        var list = loadList(CARTS_KEY);
        var d = new Date();
        list.push({ nom: 'Panier du ' + d.toLocaleDateString('fr-FR'), date: d.toLocaleDateString('fr-FR'), items: JSON.parse(JSON.stringify(cart)), total: cartSubtotal() });
        saveList(CARTS_KEY, list); renderPaniers(); showToast('Panier sauvegardé.');
      });
    }
    document.addEventListener('click', function (e) {
      var ord = e.target.closest('[data-cart-order]');
      if (ord && cart.length && ord.getAttribute('href')) {
        var list = loadList(ORDERS_KEY);
        var d = new Date();
        list.push({ id: 'CMD-' + d.getFullYear() + '-' + String(list.length + 1).padStart(4, '0'), date: d.toLocaleDateString('fr-FR'), items: JSON.parse(JSON.stringify(cart)), total: cartSubtotal() });
        saveList(ORDERS_KEY, list); renderOrders();
      }
    });
    renderAcct();
  });
})();
