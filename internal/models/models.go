package models

type Spec struct {
	Label  string `json:"label"`
	Valeur string `json:"valeur"`
}

type Product struct {
	ID            int64
	Slug          string
	Nom           string
	Marque        string
	Prix          *int64 // nil = sur devis
	Devise        string
	Dispo         string
	Categorie     string
	SousCategorie string
	Image         string
	Accroche      string
	Description   string
	SpecsJSON     string
	Specs         []Spec
	SearchIndex   string
	Stock         int
}

type CartItem struct {
	ProductID int64
	Slug      string
	Nom       string
	Prix      *int64
	Qty       int
	Image     string
}

type Order struct {
	ID            int64
	UserID        *int64
	Statut        string
	MoyenPaiement string
	Total         *int64
	ClientJSON    string
	CreatedAt     string
}

type User struct {
	ID        int64
	Email     string
	PassHash  string
	Nom       string
	Tel       string
	CreatedAt string
}
