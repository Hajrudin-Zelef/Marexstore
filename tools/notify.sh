#!/bin/sh
# Marexstore — notifications multi-canal, best-effort (T1, backend dur).
# Usage : tools/notify.sh "<evenement>" "<message>"
#   ALERT_CHANNELS env : sous-ensemble de "whatsapp,mail,telegram"
#     (défaut : tous ceux configurés ; vide = aucun envoi, exit 0).
#   WhatsApp (Meta Cloud API) : WA_ALERT_TOKEN + WA_PHONE_ID (+ WA_TO,
#     défaut 2250705087507). Sans token : repli log + lien wa.me manuel.
#   Mail : ALERT_MAIL_TO (via msmtp/sendmail, sujet "[Marexstore] <event>").
#   Telegram : TG_BOT_TOKEN + TG_CHAT_ID (api.telegram.org, timeout 10s).
# Chaque canal est isolé (échec d'un canal = warning, jamais fatal) et le
# script sort TOUJOURS 0 : une alerte ne doit jamais casser un backup.
set -eu

EVENT="${1:-info}"
MSG="${2:-}"
CHANNELS="${ALERT_CHANNELS:-whatsapp,mail,telegram}"

log() { echo "notify[$EVENT]: $*" >&2; }
have() { command -v "$1" >/dev/null 2>&1; }
want() { case ",$CHANNELS," in *",$1,"*) return 0;; *) return 1;; esac; }

send_whatsapp() {
	TO="${WA_TO:-2250705087507}"
	if [ -z "${WA_ALERT_TOKEN:-}" ] || [ -z "${WA_PHONE_ID:-}" ]; then
		log "whatsapp non configuré (WA_ALERT_TOKEN/WA_PHONE_ID absents) — message : $MSG — envoi manuel : https://wa.me/$TO"
		return 0
	fi
	if ! have curl; then log "whatsapp : curl absent"; return 0; fi
	BODY=$(printf '{"messaging_product":"whatsapp","to":"%s","type":"text","text":{"body":%s}}' "$TO" "$(printf '%s' "[Marexstore][$EVENT] $MSG" | sed 's/\\/\\\\/g; s/"/\\"/g' | tr '\n' ' ')")
	curl -s -m 10 -o /dev/null -w "%{http_code}" \
		-H "Authorization: Bearer $WA_ALERT_TOKEN" -H "Content-Type: application/json" \
		-d "$BODY" "https://graph.facebook.com/v21.0/$WA_PHONE_ID/messages" 2>/dev/null | grep -q "^2" \
		&& log "whatsapp envoyé" || log "whatsapp : envoi KO (voir token/numéro)"
	return 0
}

send_mail() {
	if [ -z "${ALERT_MAIL_TO:-}" ]; then log "mail non configuré (ALERT_MAIL_TO absent)"; return 0; fi
	if have msmtp; then SENDER="msmtp -t";
	elif have sendmail; then SENDER="sendmail -t";
	else log "mail : ni msmtp ni sendmail"; return 0; fi
	{ printf 'To: %s\nSubject: [Marexstore][%s]\n\n%s\n' "$ALERT_MAIL_TO" "$EVENT" "$MSG"; } | $SENDER 2>/dev/null \
		&& log "mail envoyé à $ALERT_MAIL_TO" || log "mail : envoi KO"
	return 0
}

send_telegram() {
	if [ -z "${TG_BOT_TOKEN:-}" ] || [ -z "${TG_CHAT_ID:-}" ]; then
		log "telegram non configuré (TG_BOT_TOKEN/TG_CHAT_ID absents)"; return 0
	fi
	if ! have curl; then log "telegram : curl absent"; return 0; fi
	curl -s -m 10 -o /dev/null --data-urlencode "chat_id=$TG_CHAT_ID" \
		--data-urlencode "text=[Marexstore][$EVENT] $MSG" \
		"https://api.telegram.org/bot$TG_BOT_TOKEN/sendMessage" 2>/dev/null | grep -q '"ok":true' \
		&& log "telegram envoyé" || log "telegram : envoi KO (voir token/chat_id)"
	return 0
}

want "whatsapp" && send_whatsapp || true
want "mail" && send_mail || true
want "telegram" && send_telegram || true
exit 0
