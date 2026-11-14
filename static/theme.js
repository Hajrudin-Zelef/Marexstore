// Thème + flag js — chargé en synchrone dans <head> (avant le rendu, anti-flash).
// Le CHOIX du visiteur (localStorage) prime ; sinon on garde les attributs
// posés par le serveur (thème global /admin/design), avec repli historique :
// couleur liquide quand le serveur ne dit rien.
// Ex-internalisé : aucun script inline => script-src 'self' strict, zéro warning CSP.
(function () {
  try {
    var t = JSON.parse(localStorage.getItem('mediatec-theme') || '{}');
    var r = document.documentElement;
    if (t.mode === 'sombre') r.setAttribute('data-theme', 'dark');
    else if (t.mode === 'clair') r.removeAttribute('data-theme');
    if (t.color) { r.setAttribute('data-color', t.color); }
    else if (!r.getAttribute('data-color')) { r.setAttribute('data-color', 'liquid'); }
  } catch (e) {}
  document.documentElement.classList.add('js');
})();
