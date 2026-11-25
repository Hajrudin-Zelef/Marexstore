/* Retour d'onglet dans le Studio.
   Le bouton « Retour à l'admin » existant reste inchangé ; celui-ci recule
   dans l'historique des onglets et bascule sur l'admin quand il est vide.
   Aucun onglet n'est codé en dur : l'historique suit les clics réels. */
(function(){
  const MAX=30, pile=[];
  let retour=false;

  /* Capture : s'exécute avant les handlers de #studioNav, donc currentTab
     contient encore l'onglet que l'on quitte. Un retour programmatique ne
     doit pas empiler l'onglet courant. */
  document.addEventListener('click',function(e){
    const b=e.target.closest&&e.target.closest('#studioNav button[data-tab]');
    if(!b) return;
    if(retour){retour=false; return;}
    const depuis=(typeof currentTab==='string')?currentTab:'';
    const vers=b.dataset.tab;
    if(depuis&&depuis!==vers&&pile[pile.length-1]!==depuis){
      pile.push(depuis);
      if(pile.length>MAX)pile.shift();
    }
  },true);

  function boutonOnglet(t){
    return document.querySelector('#studioNav button[data-tab="'+t+'"]');
  }

  window.studioTabBack=function(){
    while(pile.length){
      const t=pile.pop();
      const b=boutonOnglet(t);
      if(b){retour=true; b.click(); return;}
    }
    /* Plus d'historique : on quitte le Studio proprement. */
    location.href='/admin';
  };

  /* État visuel : atténué quand il n'y a rien à remonter. */
  function majEtat(){
    const b=document.getElementById('btnTabBack');
    if(!b)return;
    b.classList.toggle('vide',pile.length===0);
  }
  document.addEventListener('click',function(){setTimeout(majEtat,0);},true);
  addEventListener('load',majEtat);
})();
