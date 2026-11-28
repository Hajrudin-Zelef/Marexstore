package handlers

// StaticRoutes : pages statiques (corps verbatim + chrome partagé).
// Valeur = {nom du corps, titre}. Servie par App.StaticPage, câblée dans
// cmd/store/main.go — un seul lieu de vérité, testé par TestCommanderRoute.
func StaticRoutes() map[string][2]string {
	return map[string][2]string{
		"/apropos.html":                     {"apropos", "À propos de nous"},
		"/cgv.html":                         {"cgv", "Conditions générales de vente"},
		"/livraison.html":                   {"livraison", "Livraison & retours"},
		"/confidentialite.html":             {"confidentialite", "Politique de confidentialité"},
		"/mentions-legales.html":            {"mentions-legales", "Mentions légales"},
		"/contact.html":                     {"contact", "Contact"},
		"/contactez-nous.html":              {"contactez-nous", "Contactez-nous"},
		"/faq.html":                         {"faq", "Questions fréquentes"},
		"/devis.html":                       {"devis", "Demande de devis"},
		"/suivi-commande.html":              {"suivi-commande", "Suivi de commande"},
		"/compte.html":                      {"compte", "Mon compte"},
		"/commander.html":                   {"commander", "Commander"},
		"/parametres.html":                  {"parametres", "Mes paramètres"},
		"/services.html":                    {"services", "Nos Logiciels Metiers"},
		"/nos-services.html":                {"nos-services", "Nos services"},
		"/nos-references.html":              {"nos-references", "Nos références"},
		"/qui-sommes-nous.html":             {"qui-sommes-nous", "Qui sommes-nous ?"},
		"/solutions-infrastructure-it.html": {"solutions-infrastructure-it", "Infrastructure IT"},
		"/solutions-collaboration.html":     {"solutions-collaboration", "Collaboration"},
		"/solutions-cloud-securite.html":    {"solutions-cloud-securite", "Cloud & Sécurité"},
	}
}
