package models

type Recarga struct {
	ID          string  `json:"id"`
	Preco	    float64 `json:"preco"` // Preço por kWh
	Tempo	    float64 `json:"tempo"` // Tempo de recarga (em horas)
}