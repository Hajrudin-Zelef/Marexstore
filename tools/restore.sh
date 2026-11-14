#!/bin/sh
# Marexstore — restauration JETABLE (vérification d'une sauvegarde, T0).
# Usage : tools/restore.sh /chemin/vers/marexstore-AAAAMMJJ-HHMMSS.db [cible.db]
# Restaure dans un fichier temporaire (jamais la base live), vérifie
# integrity_check + comptes (products, products_fts, settings), puis affiche
# le résultat et la marche à suivre pour promouvoir la sauvegarde.
# N'ÉCRASE JAMAIS la base live : la promotion est manuelle (arrêter le
# serveur, remplacer store.db, redémarrer) — les étapes sont affichées.
set -eu

SRC="${1:-}"
DST="${2:-}"
if [ -z "$SRC" ] || [ ! -f "$SRC" ]; then
	echo "usage : $0 /chemin/vers/marexstore-AAAAMMJJ-HHMMSS.db [cible.db]" >&2
	exit 1
fi
if [ -z "$DST" ]; then
	DST=$(mktemp /tmp/marexstore-restore-XXXXXX.db)
	TEMP=1
else
	TEMP=0
fi

cp "$SRC" "$DST"
echo "restore : copie vers $DST"

if [ "$(sqlite3 "$DST" "PRAGMA integrity_check;" 2>/dev/null)" != "ok" ]; then
	echo "restore : ÉCHEC — integrity_check KO sur $DST" >&2
	[ "$TEMP" = 1 ] && rm -f "$DST"
	exit 1
fi
echo "restore : integrity_check OK"

for t in products products_fts settings; do
	n=$(sqlite3 "$DST" "SELECT COUNT(*) FROM $t;" 2>/dev/null || echo "?")
	echo "restore : $t = $n"
done

echo "restore : sauvegarde VALIDE."
if [ "$TEMP" = 1 ]; then
	echo "restore : fichier temporaire conservé : $DST"
	echo "  Pour promouvoir (manuellement, serveur arrêté) :"
	echo "    docker compose stop store"
	echo "    cp $DST ./data/store.db"
	echo "    docker compose up -d store"
else
	echo "restore : restauré dans $DST (base live non touchée)."
fi
