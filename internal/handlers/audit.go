package handlers

import (
	"net"
	"net/http"
)

// auditAdmin : trace une action d'administration (table admin_audit, T1).
// Best-effort : un échec d'écriture ne bloque jamais l'action.
// admin = compte Basic Auth, ip = adresse du client (sans le port).
func (a *App) auditAdmin(r *http.Request, action, objet string) {
	ip := r.RemoteAddr
	if h, _, err := net.SplitHostPort(ip); err == nil {
		ip = h
	}
	_, _ = a.DB.Exec(`INSERT INTO admin_audit(admin,action,objet,ip) VALUES(?,?,?,?)`,
		a.Cfg.AdminUser, action, objet, ip)
}
