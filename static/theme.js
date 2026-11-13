// Thème + flag js — chargé en synchrone dans <head> (avant le rendu, anti-flash).
// Ex-internalisé : aucun script inline => script-src 'self' strict, zéro warning CSP.
(function () {
  try {
    var t = JSON.parse(localStorage.getItem('mediatec-theme') || '{}');
    var r = document.documentElement;
    if (t.mode === 'sombre') r.setAttribute('data-theme', 'dark');
    if (t.color === 'liquid') r.setAttribute('data-color', 'liquid');
  } catch (e) {}
  document.documentElement.classList.add('js');
})();
