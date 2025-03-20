package models

type Carro struct {
    ID          string  `json:"id"`
    Bateria     float64 `json:"bateria"`  // Nível da bateria (0-100%)
    Localizacao string  `json:"localizacao"`
    Conectado   bool    `json:"conectado"`
	Debito      float64 `json:"debito"`
		
}

