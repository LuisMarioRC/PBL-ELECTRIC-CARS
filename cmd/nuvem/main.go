package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/LuisMarioRC/PBL-ELECTRIC-CARS/cmd/models"
)

var mu sync.Mutex

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		mensagem, err := reader.ReadString('\n')
		if err != nil {
			return // Conexão encerrada
		}

		mensagem = strings.TrimSpace(mensagem)
		parts := strings.Split(mensagem, "|")
		if len(parts) < 2 {
			conn.Write([]byte("Formato inválido da mensagem\n"))
			return
		}

		comando := parts[0]
		id := parts[1]

		mu.Lock()
		fmt.Printf("🤖 Recebido comando: %s, ID: %s\n", comando, id)

		switch comando {
		case "Registrar ponto":
			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.\n", id)))
				mu.Unlock()
				return
			}

			if _, exists := models.PontosDisponiveis[id]; !exists {
				models.PontosDisponiveis[id] = true
				fmt.Printf("Ponto %s registrado com sucesso.\n", id)
				conn.Write([]byte(fmt.Sprintf("Ponto %s registrado com sucesso e conectado ao sistema.\n", id)))
			} else {
				models.PontosDisponiveis[id] = true
				conn.Write([]byte(fmt.Sprintf("Ponto %s já estava registrado. Atualizado como disponível e conectado.\n", id)))
			}

		case "Carro recarregado":
			if _, exists := models.GraphInstance.Nodes[id]; !exists {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não está registrado no grafo da nuvem.\n", id)))
				mu.Unlock()
				return
			}
			if _, exists := models.PontosDisponiveis[id]; exists {
				models.PontosDisponiveis[id] = true
				fmt.Printf("Carro terminou de recarregar no ponto %s\n", id)

				if models.FilaEspera[id] > 0 {
					models.FilaEspera[id]--
					models.PontosDisponiveis[id] = false
					fmt.Printf("Ponto %s: Próximo carro na fila será atendido.\n", id)
					conn.Write([]byte(fmt.Sprintf("Ponto %s: Próximo carro pode ser atendido\n", id)))
				} else {
					fmt.Printf("Ponto %s liberado. Nenhum carro na fila.\n", id)
					conn.Write([]byte(fmt.Sprintf("Ponto %s liberado.\n", id)))
				}
			} else {
				conn.Write([]byte(fmt.Sprintf("Erro: Ponto %s não encontrado.\n", id)))
			}

		case "Carro precisa recarga":
			pontoMaisProximo, _, distancia := models.Dijkstra("1") // isso pode ser parametrizado futuramente

			if pontoMaisProximo != "" {
				if models.PontosDisponiveis[pontoMaisProximo] {
					models.PontosDisponiveis[pontoMaisProximo] = false
					fmt.Printf("Carro %s iniciou recarga no ponto %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)
					conn.Write([]byte(fmt.Sprintf("Carro %s: Dirija-se ao ponto: %s (distância: %.2f)\n", id, pontoMaisProximo, distancia)))


					// PRECISA CRIAR ALGUMA COISA PRA SETAR A LOQ DO CARRO AQUI

				} else {
					models.FilaEspera[pontoMaisProximo]++
					posicaoFila := models.FilaEspera[pontoMaisProximo]
					fmt.Printf("Carro %s adicionado à fila do ponto %s. Posição na fila: %d\n", id, pontoMaisProximo, posicaoFila)
					conn.Write([]byte(fmt.Sprintf("Carro %s: Ponto ocupado. Você está na fila do %s: posição %d\n", id, pontoMaisProximo, posicaoFila)))


					// PRECISA CRIAR ALGUMA COISA PRA SETAR A LOQ DO CARRO AQUI

				}
			} else {
				fmt.Printf("Carro %s: Nenhum ponto disponível no momento.\n", id)
				conn.Write([]byte(fmt.Sprintf("Carro %s: Nenhum ponto disponível no momento.\n", id)))
			}

		default:
			conn.Write([]byte("Comando desconhecido\n"))
		}
		mu.Unlock()
	}
}

func main() {
	models.InicializarGrafo()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Servidor Nuvem iniciado na porta 8080")
	fmt.Println("Aguardando conexões dos pontos de recarga...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
