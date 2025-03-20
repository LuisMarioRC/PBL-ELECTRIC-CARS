type Ponto struct {
	ID          string  `json:"id"`
	Localizacao string  `json:"localizacao"`
	Disponivel  bool    `json:"disponivel"`
	Carregando  *Carro  `json:"carregando,omitempty"`
	Fila		[]*Carro `json:"fila,omitempty"`
}