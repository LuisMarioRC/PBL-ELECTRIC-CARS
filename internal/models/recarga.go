type Recarga struct {
	ID          string  `json:"id"`
	preco	    float64 `json:"preco"` // Preço por kWh
	tempo	    float64 `json:"tempo"` // Tempo de recarga (em horas)
}