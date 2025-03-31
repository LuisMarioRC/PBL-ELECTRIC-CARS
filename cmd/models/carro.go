package models

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
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

var CarrosEstado = make(map[string]*Carro) // Mapa para armazenar o estado dos carros

func (c *Carro) Rodar(distancia float64) error {
    if c.Bateria <= 0 {
        return errors.New("bateria descarregada, recarregue no ponto de recarga")
    }

    consumo := distancia * ConsumoPorKm
    if consumo > c.Bateria {
        return errors.New("bateria insuficiente para completar a distância")
    }

    c.Bateria -= consumo
    CarrosEstado[c.ID] = c // Atualiza o estado no mapa
    fmt.Printf("Carro %s rodou %.2f km. Bateria restante: %.2f%%\n", c.ID, distancia, c.Bateria)
    return nil
}

func (c *Carro) SolicitarRecarga() {
    conn, err := net.Dial("tcp", "nuvem:8080")
    if err != nil {
        fmt.Printf("Carro %s: Erro ao conectar na nuvem: %s\n", c.ID, err.Error())
        return
    }
    defer conn.Close()

    message := fmt.Sprintf("Carro precisa recarga|%s", c.ID)
    _, err = conn.Write([]byte(message))
    if err != nil {
        fmt.Printf("Carro %s: Erro ao enviar solicitação de recarga: %s\n", c.ID, err.Error())
        return
    }

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Printf("Carro %s: Erro ao receber resposta da nuvem: %s\n", c.ID, err.Error())
        return
    }

    resposta := string(buffer[:n])
    fmt.Printf("Carro %s: Resposta da nuvem: %s\n", c.ID, resposta)

    if strings.Contains(resposta, "Dirija-se ao ponto") {
        tempoRecarga := rand.Intn(10) + 10 // Tempo de recarga entre 10 e 20 segundos
        fmt.Printf("Carro %s está recarregando por %d segundos...\n", c.ID, tempoRecarga)
        time.Sleep(time.Duration(tempoRecarga) * time.Second)

        // Envia mensagem para a nuvem indicando que a recarga foi concluída
        message := fmt.Sprintf("Carro recarregado|%s", c.ID)
        conn.Write([]byte(message))
        fmt.Printf("Carro %s terminou a recarga\n", c.ID)

        // Atualiza a bateria para 100%
        c.Bateria = 100
        CarrosEstado[c.ID] = c // Atualiza o estado no mapa
    }
}

// AtualizarBateria simula o consumo de bateria ao longo do tempo
func (c *Carro) AtualizarBateria() {
    for {
        time.Sleep(10 * time.Second) // Atualiza o estado da bateria a cada 10 segundos

        // Simula o consumo de bateria
        distancia := rand.Float64() * 10 // Simula uma distância aleatória entre 0 e 10 km
        err := c.Rodar(distancia)
        if err != nil {
            fmt.Printf("Carro %s: %s\n", c.ID, err.Error())
            if c.Bateria <= 25 {
                c.SolicitarRecarga()
            }
            return // Encerra a rotina
        }

        // Verifica se a bateria está em estágio crítico
        if c.Bateria <= 25 {
            fmt.Printf("Carro %s: ALERTA! Bateria em estado crítico (%.2f%%). Dirija-se ao ponto de recarga mais próximo.\n", c.ID, c.Bateria)
            c.SolicitarRecarga()
            return // Encerra a rotina
        }

        fmt.Printf("Carro %s: Bateria atual: %.2f%%\n", c.ID, c.Bateria)
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