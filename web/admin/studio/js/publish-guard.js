/* Publication protégée (lot publication) : deux verrous avant d'écrire en
   production. 1) recopie d'un code à 6 caractères (4 chiffres + 2 majuscules)
   affiché à l'écran — freine le clic réflexe ; 2) mot de passe admin, vérifié
   par le serveur qui renvoie un jeton court (3 min) exigé pour publier.
   Aucune modification de l'état du Studio : si l'utilisateur annule, rien
   n'est publié et rien n'est perdu. */
(function () {
  'use strict';

  var LETTRES = 'ABCDEFGHJKLMNPQRSTUVWXYZ'; // sans I/O (confusion visuelle)
  var CHIFFRES = '0123456789';

  function codeAleatoire() {
    var c = [], i;
    for (i = 0; i < 4; i++) c.push(CHIFFRES[Math.floor(Math.random() * CHIFFRES.length)]);
    for (i = 0; i < 2; i++) c.push(LETTRES[Math.floor(Math.random() * LETTRES.length)]);
    // mélange (Fisher-Yates)
    for (i = c.length - 1; i > 0; i--) {
      var j = Math.floor(Math.random() * (i + 1));
      var t = c[i]; c[i] = c[j]; c[j] = t;
    }
    return c.join('');
  }

  function modale(html) {
    var back = document.createElement('div');
    back.className = 'pub-back';
    back.innerHTML = '<div class="pub-modal" role="dialog" aria-modal="true" aria-labelledby="pubTitle">' + html + '</div>';
    document.body.appendChild(back);
    return back;
  }

  function fermer(back) { if (back && back.parentNode) back.parentNode.removeChild(back); }

  function resume(texte) {
    var el = document.createElement('p');
    el.className = 'pub-resume';
    el.textContent = texte;
    return el;
  }

  /* Étape 1 : avertissement + code à recopier. Résout(true) si confirmé. */
  function etapeCode(nbSections) {
    return new Promise(function (resolve) {
      var code = codeAleatoire();
      var back = modale(
        '<h3 id="pubTitle">Publier sur le site ?</h3>' +
        '<p class="pub-warn">La publication <b>remplace immédiatement</b> les textes, couleurs et réglages du site public. ' +
        'Tout visiteur verra ces changements. Une version est enregistrée : vous pourrez revenir en arrière depuis l’historique.</p>' +
        '<p class="pub-count">' + (nbSections > 0 ? nbSections + ' section(s) modifiée(s) seront mises en ligne.' : 'Aucune modification depuis la dernière publication.') + '</p>' +
        '<label class="pub-lab">Pour confirmer, tapez ce code :</label>' +
        '<div class="pub-code" aria-label="Code à recopier">' + code.split('').join(' ') + '</div>' +
        '<input class="pub-input" id="pubCode" autocomplete="off" spellcheck="false" maxlength="6" placeholder="6 caractères">' +
        '<p class="pub-err" id="pubErr" hidden>Le code ne correspond pas.</p>' +
        '<div class="pub-act"><button type="button" class="btn btn-ghost" id="pubCancel">Annuler</button>' +
        '<button type="button" class="btn btn-publish" id="pubNext" disabled>Continuer</button></div>'
      );
      var inp = back.querySelector('#pubCode');
      var next = back.querySelector('#pubNext');
      var err = back.querySelector('#pubErr');
      inp.addEventListener('input', function () {
        inp.value = inp.value.toUpperCase().replace(/[^0-9A-Z]/g, '').slice(0, 6);
        next.disabled = inp.value !== code;
        err.hidden = true;
      });
      inp.addEventListener('keydown', function (e) {
        if (e.key === 'Enter' && !next.disabled) next.click();
      });
      back.querySelector('#pubCancel').addEventListener('click', function () { fermer(back); resolve(false); });
      next.addEventListener('click', function () {
        if (inp.value !== code) { err.hidden = false; return; }
        fermer(back); resolve(true);
      });
      setTimeout(function () { inp.focus(); }, 30);
    });
  }

  /* Étape 2 : mot de passe admin → jeton de publication. Résout(jeton) ou null. */
  function etapeMotDePasse() {
    return new Promise(function (resolve) {
      var back = modale(
        '<h3 id="pubTitle">Confirmer avec le mot de passe admin</h3>' +
        '<p class="pub-warn">Dernière étape : le serveur vérifie votre mot de passe et délivre une autorisation valable 3 minutes.</p>' +
        '<label class="pub-lab" for="pubPw">Mot de passe administrateur</label>' +
        '<input class="pub-input" id="pubPw" type="password" autocomplete="current-password" placeholder="Mot de passe admin">' +
        '<p class="pub-err" id="pubPwErr" hidden></p>' +
        '<div class="pub-act"><button type="button" class="btn btn-ghost" id="pubCancel2">Annuler</button>' +
        '<button type="button" class="btn btn-publish" id="pubGo">Publier</button></div>'
      );
      var pw = back.querySelector('#pubPw');
      var go = back.querySelector('#pubGo');
      var err = back.querySelector('#pubPwErr');
      function echec(msg) { err.textContent = msg; err.hidden = false; go.disabled = false; go.textContent = 'Publier'; }
      back.querySelector('#pubCancel2').addEventListener('click', function () { fermer(back); resolve(null); });
      pw.addEventListener('keydown', function (e) { if (e.key === 'Enter') go.click(); });
      go.addEventListener('click', function () {
        if (!pw.value) { echec('Mot de passe requis.'); return; }
        go.disabled = true; go.textContent = 'Vérification…';
        fetch('/admin/api/publish/confirm', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'X-CSRF': window.__STUDIO_CSRF__ || '' },
          credentials: 'same-origin',
          body: JSON.stringify({ password: pw.value })
        }).then(function (r) {
          if (r.status === 401) { echec('Mot de passe incorrect.'); return null; }
          if (!r.ok) { echec('Vérification impossible (' + r.status + ').'); return null; }
          return r.json();
        }).then(function (j) {
          if (!j || !j.token) return;
          fermer(back); resolve(j.token);
        }).catch(function () { echec('Réseau indisponible, réessayez.'); });
      });
      setTimeout(function () { pw.focus(); }, 30);
    });
  }

  window.publierAvecConfirmation = function (nbSections, envoyer) {
    etapeCode(nbSections).then(function (ok) {
      if (!ok) return null;
      return etapeMotDePasse().then(function (token) {
        if (!token) return null;
        return envoyer(token);
      });
    });
    return null;
  };
})();
