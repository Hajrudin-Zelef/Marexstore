// Thème + flag js — chargé en synchrone dans <head> (avant le rendu, anti-flash).
// Le CHOIX du visiteur (localStorage) prime ; sinon on garde les attributs
// posés par le serveur (thème global /admin/design), avec repli historique :
// couleur liquide quand le serveur ne dit rien.
// P1 — le serveur gagne : quand le thème publié change de version
// (data-theme-version sur <html>), on purge le choix couleur mémorisé
// (et le mode si le serveur en impose un) puis on applique le serveur.
// Le visiteur ne garde en local que clair/sombre, tant que le serveur
// n'impose pas de mode.
// Ex-internalisé : aucun script inline => script-src 'self' strict, zéro warning CSP.
(function () {
  try {
    var t = JSON.parse(localStorage.getItem('mediatec-theme') || '{}');
    var r = document.documentElement;
    var sv = r.getAttribute('data-theme-version');
    if (sv && String(t.v) !== sv) {
      delete t.color;
      if (r.getAttribute('data-theme')) delete t.mode;
      t.v = sv;
      try { localStorage.setItem('mediatec-theme', JSON.stringify(t)); } catch (e2) {}
    }
    if (t.mode === 'sombre') r.setAttribute('data-theme', 'dark');
    else if (t.mode === 'clair') r.removeAttribute('data-theme');
    if (t.color) { r.setAttribute('data-color', t.color); }
    else if (!r.getAttribute('data-color')) { r.setAttribute('data-color', 'liquid'); }
  } catch (e) {}
  document.documentElement.classList.add('js');
})();
