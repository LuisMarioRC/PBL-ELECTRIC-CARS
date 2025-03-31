package main

import (
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
)

var mu sync.Mutex

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		mensagem := string(buffer[:n])
		// fmt.Printf("Recebido: %s\n", mensagem)

		mu.Lock()
		if strings.HasPrefix(mensagem, "Carro recarregado|") {
			parts := strings.Split(mensagem, "|")
			if len(parts) < 2 {
				conn.Write([]byte("Formato inválido da mensagem"))
				mu.Unlock()
				continue
			}
			ponto := parts[1]
			if _, exists := models.PontosDisponiveis[ponto]; exists {
				models.PontosDisponiveis[ponto] = true
				fmt.Printf("Carro terminou de recarregar no ponto %s\n", ponto)

				if models.FilaEspera[ponto] > 0 {
					fmt.Printf("Ponto %s liberado. Próximo carro na fila será atendido.\n", ponto)

					for carroID, fila := range models.FilaEspera {
						if fila > 0 {
							models.FilaEspera[carroID]--
							if models.FilaEspera[carroID] == 0 {
								delete(models.FilaEspera, carroID)
							}
							models.PontosDisponiveis[ponto] = false
							fmt.Printf("Notificando carro %s para usar o ponto %s.\n", carroID, ponto)
							conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto %s", carroID, ponto)))
							break
						}
					}
				} else {
					fmt.Printf("Ponto %s liberado. Nenhum carro na fila.\n", ponto)
				}
			} else {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não encontrado.", ponto)))
			}
		}

		parts := strings.Split(mensagem, "|")
		if len(parts) < 2 {
			conn.Write([]byte("Formato inválido da mensagem"))
			mu.Unlock()
			continue
		}
		comando := parts[0]
		carroID := parts[1]

		if comando == "Carro precisa recarga" {
			if carroID == "" {
				conn.Write([]byte("Formato inválido da mensagem: ID do carro está vazio"))
				mu.Unlock()
				continue
			}

			destino, _, distancia := models.Dijkstra("Ponto-1")

			if destino != "" {
				if models.PontosDisponiveis[destino] {
					models.PontosDisponiveis[destino] = false
					fmt.Printf("Carro %s iniciou recarga no ponto %s (distância: %.2f)\n", carroID, destino, distancia)
					conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s (distância: %.2f)", carroID, destino, distancia)))
				} else {
					models.FilaEspera[destino]++
					posicaoFila := models.FilaEspera[destino]
					conn.Write([]byte(fmt.Sprintf("Carro %s: Todos os pontos estão ocupados. Você está na fila do %s: posição %d", carroID, destino, posicaoFila)))
				}
			} else {
				var menorFilaPonto string
				menorDistancia := math.Inf(1)
				for ponto, disponivel := range models.PontosDisponiveis {
					if !disponivel {
						_, _, distancia := models.Dijkstra(ponto)
						if distancia < menorDistancia {
							menorDistancia = distancia
							menorFilaPonto = ponto
						} else if distancia == menorDistancia && models.FilaEspera[ponto] < models.FilaEspera[menorFilaPonto] {
							menorFilaPonto = ponto
						}
					}
				}
				models.FilaEspera[menorFilaPonto]++
				posicaoFila := models.FilaEspera[menorFilaPonto]
				conn.Write([]byte(fmt.Sprintf("Carro %s: Nenhum ponto disponível. Você está na fila do %s: posição %d", carroID, menorFilaPonto, posicaoFila)))
			}
		}
		mu.Unlock()

		time.Sleep(2 * time.Second)
	}
}

func main() {
	models.InicializarGrafo()

	carros := []*models.Carro{
		{ID: "Carro1", Bateria: 100},
		{ID: "Carro2", Bateria: 100},
		{ID: "Carro3", Bateria: 100},
		{ID: "Carro4", Bateria: 100},
		{ID: "Carro5", Bateria: 100},
		{ID: "Carro6", Bateria: 100},
		{ID: "Carro7", Bateria: 100},
	}

	var wg sync.WaitGroup
	for _, carro := range carros {
		wg.Add(1)
		go func(c *models.Carro) {
			defer wg.Done()
			c.AtualizarBateria()
		}(carro)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Servidor Nuvem iniciado na porta 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}

	wg.Wait()
}
