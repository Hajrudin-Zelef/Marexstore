#!/bin/sh
# Marexstore — sauvegarde SQLite (T0, socle backend).
# Usage : tools/backup.sh [chemin/vers/store.db]
#   DB_PATH env (défaut : ./data/store.db) — la base live, même si le serveur tourne.
#   BACKUP_DIR env (défaut : /var/backups/marexstore).
#   BACKUP_TARGET env : local|s3|ftp|all (défaut : local).
# La sauvegarde est faite par `sqlite3 .backup` (cohérente même si le serveur
# tourne), vérifiée par integrity_check, puis copiée vers la cible.
# Rotation : 7 jours + 4 mensuelles (1re sauvegarde de chaque mois). chmod 600.
# Échec → notify.sh (si présent, fourni par l'agent) + ops_jobs + exit 1.
# Le cron est installé par l'agent sur la VM : 0 2 * * * backup.sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DB="${1:-${DB_PATH:-./data/store.db}}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/marexstore}"
TARGET="${BACKUP_TARGET:-local}"

if [ ! -f "$DB" ]; then
	echo "backup: base introuvable : $DB" >&2
	exit 1
fi
mkdir -p "$BACKUP_DIR" || { echo "backup: impossible de créer $BACKUP_DIR" >&2; exit 1; }

notify() {
	# $1 = événement (backup_ok / backup_ko), $2 = message. Best-effort.
	if [ -x "$SCRIPT_DIR/notify.sh" ]; then
		"$SCRIPT_DIR/notify.sh" "$1" "$2" || true
	fi
}

ops() {
	# $1 = detail. Best-effort (ne doit jamais faire échouer la sauvegarde).
	sqlite3 "$DB" "INSERT OR REPLACE INTO ops_jobs(nom,derniere_ok,detail) VALUES('backup', datetime('now','localtime'), '$1')" 2>/dev/null || true
}

TS=$(date +%Y%m%d-%H%M%S)-$$
TMP="$BACKUP_DIR/.marexstore-$TS.tmp"
FINAL="$BACKUP_DIR/marexstore-$TS.db"

# 1. Copie cohérente (.backup fonctionne même si le serveur tourne).
# Repli python3 (API sqlite3.backup) si le binaire sqlite3 est absent.
dobackup() {
	if command -v sqlite3 >/dev/null 2>&1; then
		sqlite3 "$1" ".backup '$2'"
		return $?
	fi
	python3 - "$1" "$2" <<'PYEOF' || return 1
import sqlite3, sys
src = sqlite3.connect(sys.argv[1])
dst = sqlite3.connect(sys.argv[2])
with dst:
    src.backup(dst)
dst.close()
src.close()
PYEOF
}
checkok() {
	if command -v sqlite3 >/dev/null 2>&1; then
		[ "$(sqlite3 "$1" "PRAGMA integrity_check;" 2>/dev/null)" = "ok" ]
		return $?
	fi
	python3 - "$1" <<'PYEOF'
import sqlite3, sys
c = sqlite3.connect(sys.argv[1])
sys.exit(0 if c.execute("PRAGMA integrity_check").fetchone()[0] == "ok" else 1)
PYEOF
}
if ! dobackup "$DB" "$TMP"; then
	rm -f "$TMP"
	ops "ÉCHEC .backup ($TS)"
	notify backup_ko "Sauvegarde Marexstore : .backup a échoué ($TS)."
	echo "backup: .backup a échoué" >&2
	exit 1
fi
# 2. Vérification d'intégrité de la copie.
if ! checkok "$TMP"; then
	rm -f "$TMP"
	ops "ÉCHEC integrity_check ($TS)"
	notify backup_ko "Sauvegarde Marexstore : integrity_check KO ($TS)."
	echo "backup: integrity_check KO sur la copie" >&2
	exit 1
fi
mv "$TMP" "$FINAL"
chmod 600 "$FINAL"

# 3. Cibles distantes (best-effort : un avertissement, jamais un échec).
case "$TARGET" in
	s3|all)
		if [ -n "${AWS_S3_BUCKET:-}" ] && command -v aws >/dev/null 2>&1; then
			aws s3 cp "$FINAL" "s3://$AWS_S3_BUCKET/${AWS_S3_PREFIX:-marexstore}/$(basename "$FINAL")" >/dev/null 2>&1 \
				|| echo "backup: avertissement — envoi S3 échoué (copie locale OK)." >&2
		else
			echo "backup: avertissement — S3 demandé mais AWS_S3_BUCKET ou aws absent (copie locale OK)." >&2
		fi
		;;
esac
case "$TARGET" in
	ftp|all)
		if [ -n "${FTP_URL:-}" ] && command -v curl >/dev/null 2>&1; then
			curl -sS --fail -T "$FINAL" "$FTP_URL/$(basename "$FINAL")" ${FTP_USER:+--user "$FTP_USER:${FTP_PASS:-}"} >/dev/null 2>&1 \
				|| echo "backup: avertissement — envoi FTP échoué (copie locale OK)." >&2
		else
			echo "backup: avertissement — FTP demandé mais FTP_URL ou curl absent (copie locale OK)." >&2
		fi
		;;
esac

# 4. Rotation : 7 jours + 4 mensuelles (1re sauvegarde de chaque mois).
#    On garde tout ce qui a moins de 7 jours, puis la 1re sauvegarde de
#    chacun des 4 derniers mois ; le reste est supprimé.
CUTOFF=$(date -d "7 days ago" +%Y%m%d)
# Mois à garder (AAAAMM), du plus récent au plus ancien (4 mois).
KEEP_MONTHS=$(for i in 0 1 2 3; do date -d "$i month ago" +%Y%m; done)
for f in "$BACKUP_DIR"/marexstore-*.db; do
	[ -f "$f" ] || continue
	base=$(basename "$f")
	d=${base#marexstore-}; d=${d%%-*} # AAAAMMJJ
	if [ "$d" -ge "$CUTOFF" ]; then
		continue # moins de 7 jours : gardé
	fi
	ym=${d%??} # AAAAMM
	keep=0
	for km in $KEEP_MONTHS; do
		[ "$ym" = "$km" ] && keep=1
	done
	[ "$keep" = 1 ] || { rm -f "$f"; continue; }
	# 1re sauvegarde du mois ? (la plus ancienne = 1re par ordre alphabétique,
	# le nom contient l'horodatage).
	first=$(ls "$BACKUP_DIR"/marexstore-"$ym"*.db 2>/dev/null | head -n 1)
	[ "$f" = "$first" ] || rm -f "$f"
done

ops "OK $FINAL"
notify backup_ok "Sauvegarde Marexstore OK : $(basename "$FINAL")."
echo "backup OK : $FINAL"
