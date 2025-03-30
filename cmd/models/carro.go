package models

import (
    "errors"
    "fmt"
    "math/rand"
    "time"
)

type Carro struct {
    ID          string  `json:"id"`
    Bateria     float64 `json:"bateria"` // Nível da bateria (0-100%)
    Localizacao string  `json:"localizacao"`
    Conectado   bool    `json:"conectado"`
    Debito      float64 `json:"debito"`
}

// ConsumoPorKm define o consumo de bateria por quilômetro rodado
const ConsumoPorKm = 7.5

// Rodar simula o carro rodando uma certa distância e descarrega a bateria
func (c *Carro) Rodar(distancia float64) error {
    if c.Bateria <= 0 {
        return errors.New("bateria descarregada, recarregue no ponto de recarga")
    }

    consumo := distancia * ConsumoPorKm
    if consumo > c.Bateria {
        return errors.New("bateria insuficiente para completar a distância")
    }

    c.Bateria -= consumo
    fmt.Printf("Carro %s rodou %.2f km. Bateria restante: %.2f%%\n", c.ID, distancia, c.Bateria)
    return nil
}

// AtualizarBateria simula o consumo de bateria ao longo do tempo
func (c *Carro) AtualizarBateria() {
    for {
        time.Sleep(10 * time.Second) // Atualiza o estado da bateria a cada 10 segundos

        // Verifica se a bateria está descarregada
        if c.Bateria <= 0 {
            fmt.Printf("Carro %s: Bateria descarregada. Dirija-se ao ponto de recarga.\n", c.ID)
            return // Encerra a rotina
        }

        distancia := rand.Float64() * 10 // Simula uma distância aleatória entre 0 e 10 km
        err := c.Rodar(distancia)
        if err != nil {
            fmt.Printf("Carro %s: %s\n", c.ID, err.Error())
            continue // Continua a execução para verificar novamente
        }

        // Verifica se a bateria está em estágio crítico
        if c.Bateria <= 25 {
            fmt.Printf("Carro %s: ALERTA! Bateria em estado crítico (%.2f%%). Dirija-se ao ponto de recarga mais próximo.\n", c.ID, c.Bateria)
        }

    }
}

// PrecoPorUnidade define o custo por 1% de recarga da bateria
const PrecoPorUnidade = 0.50 // Exemplo: R$ 0,50 por 1%

// CalcularCustoRecarga calcula o custo para recarregar a bateria até 100%
func (c *Carro) CalcularCustoRecarga() float64 {
    bateriaFaltante := 100 - c.Bateria
    custo := bateriaFaltante * PrecoPorUnidade
    return custo
}	