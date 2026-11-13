// Marexstore Admin — interactions (postboxes repliables, confirmations).
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
