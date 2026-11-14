// Design actif (phase 3, socle 3.0) : le design publié, chargé une seule fois
// au démarrage du serveur et rechargé à chaque publication / activation
// (comme le thème). Le rendu le lit via des accesseurs défensifs : une clé
// absente ou d'un mauvais type retourne le défaut (la valeur codée en dur
// aujourd'hui) — un design partiel ne plante jamais le rendu, et sans design
// publié le site est identique à aujourd'hui.
//
// Concurrence : le design est lu par le rendu à chaque requête et écrit sur
// publication — l'accès est protégé par un RWMutex. Le document est
// immuable une fois parsé (on remplace le pointeur, on ne le modifie pas).
package design

import (
	"database/sql"
	"encoding/json"
	"sync"
)

// Document : un design publié, parsé (map JSON). Version = numéro de la
// version active dans design_files (0 si rien n'est publié).
type Document struct {
	raw     map[string]any
	version int
}

var (
	mu     sync.RWMutex
	active *Document
)

// LoadActive : (re)charge le design actif depuis la base. Sans design publié,
// l'actif est nil. Une erreur de lecture ou un JSON illisible est retournée
// (l'appelant décide : au démarrage on journalise et on continue sans design).
func LoadActive(db *sql.DB) error {
	raw, version, err := Load(db)
	if err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	if raw == nil {
		active = nil
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	active = &Document{raw: m, version: version}
	return nil
}

// Refresh : recharge le design actif (après publication / activation).
func Refresh(db *sql.DB) error { return LoadActive(db) }

// Current : le design actif (nil si rien n'est publié). Les accesseurs
// acceptent un récepteur nil (ils retournent alors le défaut).
func Current() *Document {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// Version : numéro de la version active (0 si rien n'est publié).
func (d *Document) Version() int {
	if d == nil {
		return 0
	}
	return d.version
}

// section : l'objet d'une section (front, header, shop…), ou nil.
func (d *Document) section(name string) map[string]any {
	if d == nil || d.raw == nil {
		return nil
	}
	s, _ := d.raw[name].(map[string]any)
	return s
}

// Text : un texte d'affichage (Tx[key]), ou le défaut si absent / non-string.
func (d *Document) Text(key, def string) string {
	if d == nil || d.raw == nil {
		return def
	}
	tx, _ := d.raw["Tx"].(map[string]any)
	if tx == nil {
		return def
	}
	if v, ok := tx[key].(string); ok {
		return v
	}
	return def
}

// Bool : un booléen de section (section[key]), ou le défaut.
func (d *Document) Bool(section, key string, def bool) bool {
	s := d.section(section)
	if s == nil {
		return def
	}
	if v, ok := s[key].(bool); ok {
		return v
	}
	return def
}

// Str : une chaîne de section (section[key]), ou le défaut.
func (d *Document) Str(section, key, def string) string {
	s := d.section(section)
	if s == nil {
		return def
	}
	if v, ok := s[key].(string); ok {
		return v
	}
	return def
}

// Int : un nombre de section (section[key]), ou le défaut. Le JSON décode
// les nombres en float64.
func (d *Document) Int(section, key string, def int) int {
	s := d.section(section)
	if s == nil {
		return def
	}
	if v, ok := s[key].(float64); ok {
		return int(v)
	}
	return def
}

// List : une liste d'objets de section (section[key]), ou nil. Sert par
// exemple aux liens de navigation (header.navLinks : [{key,label,visible}]).
func (d *Document) List(section, key string) []map[string]any {
	s := d.section(section)
	if s == nil {
		return nil
	}
	arr, _ := s[key].([]any)
	if arr == nil {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// Has : true si la section existe dans le design.
func (d *Document) Has(section string) bool {
	return d.section(section) != nil
}
